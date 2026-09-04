package internal

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type extismCoreResult struct {
	wrapper *CoreWrapper
	err     error
}

type observedCoreLock struct {
	mutex   *sync.Mutex
	blocked chan<- struct{}
}

func (l *observedCoreLock) Lock() {
	if l.mutex.TryLock() {
		return
	}
	l.blocked <- struct{}{}
	l.mutex.Lock()
}

func (l *observedCoreLock) Unlock() {
	l.mutex.Unlock()
}

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	value, err := loadWASM(ctx)
	require.NoError(t, err)

	// check ExportedFunctionsDefinitions names contain init_client, invoke and release_client
	functions := [3]string{"init_client", "invoke", "release_client"}
	count := 0

	for _, function := range functions {
		if value.FunctionExists(function) {
			count++
		}
	}

	assert.Equal(t, len(functions), count)

	// check AllowedHosts field matches allowed1PHosts
	pluginHosts := sort.StringSlice(value.AllowedHosts)
	opHosts := sort.StringSlice(allowed1PHosts())

	assert.Equal(t, len(pluginHosts), len(opHosts))

	for x := range pluginHosts {
		assert.Equal(t, pluginHosts[x], opHosts[x])
	}
}

func TestGetExtismCoreConcurrentColdStartUsesOneCore(t *testing.T) {
	ReleaseCore()
	t.Cleanup(ReleaseCore)

	initializerEntered := make(chan struct{})
	finishInitialization := make(chan struct{})
	var initializerCalls atomic.Int32
	initializer := func() (*ExtismCore, error) {
		initializedCore := &ExtismCore{}
		if initializerCalls.Add(1) == 1 {
			close(initializerEntered)
			<-finishInitialization
		}
		return initializedCore, nil
	}

	lockBlocked := make(chan struct{}, 1)
	observedLock := &observedCoreLock{
		mutex:   &coreMu,
		blocked: lockBlocked,
	}
	ready := make(chan struct{}, 2)
	startFirst := make(chan struct{})
	startSecond := make(chan struct{})
	firstResult := make(chan extismCoreResult, 1)
	secondResult := make(chan extismCoreResult, 1)
	secondDone := make(chan struct{})

	var callers sync.WaitGroup
	callers.Add(2)
	go func() {
		defer callers.Done()
		ready <- struct{}{}
		<-startFirst
		wrapper, err := getExtismCoreWithLock(observedLock, initializer)
		firstResult <- extismCoreResult{wrapper: wrapper, err: err}
	}()
	go func() {
		defer callers.Done()
		defer close(secondDone)
		ready <- struct{}{}
		<-startSecond
		wrapper, err := getExtismCoreWithLock(observedLock, initializer)
		secondResult <- extismCoreResult{wrapper: wrapper, err: err}
	}()

	for range 2 {
		<-ready
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	close(startFirst)
	select {
	case <-initializerEntered:
	case <-deadline.C:
		close(startSecond)
		close(finishInitialization)
		t.Fatal("timed out waiting for the first initializer to start")
	}

	close(startSecond)
	secondCompletedBeforeInitialization := false
	select {
	case <-lockBlocked:
	case <-secondDone:
		secondCompletedBeforeInitialization = true
	case <-deadline.C:
		close(finishInitialization)
		t.Fatal("timed out waiting for the second caller to reach the lifecycle lock")
	}
	close(finishInitialization)

	callersDone := make(chan struct{})
	go func() {
		callers.Wait()
		close(callersDone)
	}()
	select {
	case <-callersDone:
	case <-deadline.C:
		t.Fatal("timed out waiting for core callers to finish")
	}

	first := <-firstResult
	second := <-secondResult
	assert.False(t, secondCompletedBeforeInitialization, "second caller completed while the first initialization was blocked")
	require.NoError(t, first.err)
	require.NoError(t, second.err)
	require.NotNil(t, first.wrapper)
	require.NotNil(t, second.wrapper)
	assert.Equal(t, int32(1), initializerCalls.Load())
	assert.Same(t, first.wrapper.InnerCore, second.wrapper.InnerCore)
}

func TestGetExtismCoreRetriesAfterInitializationError(t *testing.T) {
	ReleaseCore()
	t.Cleanup(ReleaseCore)

	initializationError := errors.New("initialization failed")
	sentinelCore := &ExtismCore{}
	var initializerCalls atomic.Int32
	initializer := func() (*ExtismCore, error) {
		if initializerCalls.Add(1) == 1 {
			return nil, initializationError
		}
		return sentinelCore, nil
	}

	firstWrapper, firstErr := getExtismCore(initializer)
	assert.Nil(t, firstWrapper)
	assert.ErrorIs(t, firstErr, initializationError)

	secondWrapper, secondErr := getExtismCore(initializer)
	require.NoError(t, secondErr)
	require.NotNil(t, secondWrapper)
	assert.Same(t, sentinelCore, secondWrapper.InnerCore)

	thirdWrapper, thirdErr := getExtismCore(initializer)
	require.NoError(t, thirdErr)
	require.NotNil(t, thirdWrapper)
	assert.Same(t, sentinelCore, thirdWrapper.InnerCore)
	assert.Equal(t, int32(2), initializerCalls.Load())
}

func TestReleaseCoreSynchronizesWithColdInitialization(t *testing.T) {
	ReleaseCore()
	t.Cleanup(ReleaseCore)

	sentinelCore := &ExtismCore{}
	initializerEntered := make(chan struct{})
	finishInitialization := make(chan struct{})
	initializerResult := make(chan extismCoreResult, 1)
	releaseDone := make(chan struct{})
	operationsDone := make(chan struct{})
	lockBlocked := make(chan struct{}, 1)
	observedLock := &observedCoreLock{
		mutex:   &coreMu,
		blocked: lockBlocked,
	}

	var operations sync.WaitGroup
	operations.Add(1)
	go func() {
		defer operations.Done()
		wrapper, err := getExtismCoreWithLock(observedLock, func() (*ExtismCore, error) {
			close(initializerEntered)
			<-finishInitialization
			return sentinelCore, nil
		})
		initializerResult <- extismCoreResult{wrapper: wrapper, err: err}
	}()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	select {
	case <-initializerEntered:
	case <-deadline.C:
		close(finishInitialization)
		t.Fatal("timed out waiting for initializer to start")
	}

	operations.Add(1)
	go func() {
		defer operations.Done()
		defer close(releaseDone)
		releaseCoreWithLock(observedLock)
	}()

	releaseCompletedDuringInitialization := false
	select {
	case <-lockBlocked:
		select {
		case <-releaseDone:
			releaseCompletedDuringInitialization = true
		default:
		}
	case <-releaseDone:
		releaseCompletedDuringInitialization = true
	case <-deadline.C:
		close(finishInitialization)
		t.Fatal("timed out waiting for release to reach the lifecycle lock")
	}
	close(finishInitialization)

	go func() {
		operations.Wait()
		close(operationsDone)
	}()

	select {
	case <-operationsDone:
	case <-deadline.C:
		t.Fatal("timed out waiting for initialization and release to finish")
	}

	result := <-initializerResult
	assert.False(t, releaseCompletedDuringInitialization, "release completed while initialization held the lifecycle lock")
	require.NoError(t, result.err)
	require.NotNil(t, result.wrapper)
	assert.Same(t, sentinelCore, result.wrapper.InnerCore)

	coreMu.Lock()
	finalCore := core
	coreMu.Unlock()
	assert.Nil(t, finalCore)
}
