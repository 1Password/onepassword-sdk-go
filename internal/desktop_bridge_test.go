package internal

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type bridgeFakeCore struct{}

func (bridgeFakeCore) InitClient(_ context.Context, payload []byte) ([]byte, error) {
	return payload, nil
}
func (bridgeFakeCore) Invoke(_ context.Context, payload []byte) ([]byte, error) {
	if string(payload) == "fail" {
		return nil, errors.New("dummy SDK failure")
	}
	if string(payload) == "block" {
		time.Sleep(time.Minute)
	}
	return payload, nil
}
func (bridgeFakeCore) ReleaseClient([]byte) {}

func TestDesktopBridgeChild(t *testing.T) {
	if os.Args[len(os.Args)-1] != "dummy-bridge-child" {
		return
	}
	if err := ServeDesktopBridge(context.Background(), bridgeFakeCore{}, os.Stdin, os.Stdout); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func testDesktopBridge(t *testing.T) *desktopBridgeCore {
	t.Helper()
	executable := os.Args[0]
	if windowsTest := os.Getenv("OP_TEST_WINDOWS_BRIDGE"); windowsTest != "" {
		executable = windowsTest
	}
	command := exec.Command(executable, "-test.run=^TestDesktopBridgeChild$", "--", "dummy-bridge-child")
	core, err := startDesktopBridge(command)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { core.mu.Lock(); defer core.mu.Unlock(); core.close() })
	return core
}

func TestDesktopBridgeRoundTripAndRelease(t *testing.T) {
	core := testDesktopBridge(t)
	for _, method := range []string{"init", "invoke"} {
		payload := []byte("dummy value with newline\nand quotes\" and $dollar")
		got, err := core.exchange(context.Background(), method, payload)
		if err != nil || !bytes.Equal(got, payload) {
			t.Fatalf("%s: round trip failed: %v", method, err)
		}
	}
	_, err := core.Invoke(context.Background(), []byte("fail"))
	if err == nil || err.Error() != "dummy SDK failure" {
		t.Fatalf("SDK error not propagated: %v", err)
	}
	core.ReleaseClient([]byte("1"))
	if _, err := core.Invoke(context.Background(), nil); err == nil {
		t.Fatal("released helper still usable")
	}
}

func TestDesktopBridgeCancellationKillsHelper(t *testing.T) {
	core := testDesktopBridge(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := core.Invoke(ctx, []byte("block"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if !core.closed || core.command.ProcessState == nil {
		t.Fatal("helper was not reaped")
	}
}

func TestDesktopBridgeMalformedInputDoesNotEchoPayload(t *testing.T) {
	var output bytes.Buffer
	err := ServeDesktopBridge(context.Background(), bridgeFakeCore{}, strings.NewReader("not-json-DUMMY\n"), &output)
	if err == nil || strings.Contains(err.Error(), "DUMMY") || output.Len() != 0 {
		t.Fatal("malformed request leaked or succeeded")
	}
}

func TestDesktopBridgeRequiresExplicitAbsoluteHelper(t *testing.T) {
	_, err := getDesktopBridgeCore("dummy-account", "helper.exe")
	if err == nil {
		t.Fatal("relative helper accepted")
	}
}
