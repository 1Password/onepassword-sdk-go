package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	core "github.com/1password/1password-sdk-core/wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const (
	allocateFuncName      = "allocate"
	deallocateFuncName    = "deallocate"
	invokeFuncName        = "invoke"
	initClientFuncName    = "init_client"
	releaseClientFuncName = "release_client"
)

var coreWASMModule api.Module

// connect creates a wazero runtime and instantiates the core WebAssembly module that also imports Go defined callbacks.
func connect() error {
	ctx := context.Background()

	wasmRuntimeConfig := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntimeWithConfig(ctx, wasmRuntimeConfig)

	// Instantiate a Go-defined module named "env" that exports a function to
	// make http requests. This will be imported by the core WASM.
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(httpGet).Export("http_get").
		Instantiate(ctx)
	if err != nil {
		return err
	}

	mod, err := r.Instantiate(ctx, core.GetWASMCore())
	if err != nil {
		return err
	}

	coreWASMModule = mod
	return nil
}

// isConnected checks whether the wasm core module is loaded.
func isConnected() bool {
	return coreWASMModule != nil && !coreWASMModule.IsClosed()
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	if !isConnected() {
		err := connect()
		if err != nil {
			return nil, err
		}
	}

	// TODO: fix json error: when deserializing in Rust unsupported control characters keep being found
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	initClient := coreWASMModule.ExportedFunction(initClientFuncName)
	allocate := coreWASMModule.ExportedFunction(allocateFuncName)
	deallocate := coreWASMModule.ExportedFunction(deallocateFuncName)

	configPtr, err := initializeWASMByteArray(ctx, jsonConfig, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, configPtr.offset, configPtr.len)

	pointerToErrPtr, err := initializeWASMUInt32(ctx, 0, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, pointerToErrPtr.offset, pointerToErrPtr.len)

	pointerToIDResponse, err := initializeWASMUInt32(ctx, 0, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, pointerToIDResponse.offset, pointerToIDResponse.len)
	pointerToErrSize, err := initializeWASMUInt32(ctx, 0, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, pointerToErrSize.offset, pointerToErrSize.len)

	hasErr, err := initClient.Call(ctx, configPtr.offset, configPtr.len, pointerToIDResponse.offset, pointerToErrPtr.offset, pointerToErrSize.offset)
	if err != nil {
		return nil, err
	}
	if hasErr[0] == 1 {
		errorOffset, err := dereferencePointer(pointerToErrPtr)
		if err != nil {
			return nil, err
		}
		errorSize, err := dereferencePointer(pointerToErrSize)
		if err != nil {
			return nil, err
		}
		errorBytes, err := dereferencePointer(&pointer{
			offset: uint64(binary.LittleEndian.Uint32(errorOffset)),
			len:    uint64(binary.LittleEndian.Uint32(errorSize)),
		})
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(string(errorBytes))
	} else {
		idBytes, err := dereferencePointer(pointerToIDResponse)
		if err != nil {
			return nil, err
		}
		id := uint64(binary.LittleEndian.Uint32(idBytes))
		return &id, nil
	}
}

// Invoke calls specified business logic from core
func Invoke(ctx context.Context, method string, serializedParams string) (*string, error) {
	invoke := coreWASMModule.ExportedFunction(invokeFuncName)
	allocate := coreWASMModule.ExportedFunction(allocateFuncName)
	deallocate := coreWASMModule.ExportedFunction(deallocateFuncName)

	methodPtr, err := initializeWASMByteArray(ctx, []byte(method), allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, methodPtr.offset, methodPtr.len)

	argsPtr, err := initializeWASMByteArray(ctx, []byte(serializedParams), allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, argsPtr.offset, argsPtr.len)

	pointerToResponseOffset, err := initializeWASMUInt32(ctx, 0, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, pointerToResponseOffset.offset, pointerToResponseOffset.len)

	pointerToResponseSize, err := initializeWASMUInt32(ctx, 0, allocate)
	if err != nil {
		return nil, err
	}
	defer deallocate.Call(ctx, pointerToResponseSize.offset, pointerToResponseSize.len)

	hasError, err := invoke.Call(ctx, methodPtr.offset, methodPtr.len, argsPtr.offset, argsPtr.len, pointerToResponseOffset.offset, pointerToResponseSize.offset)
	if err != nil {
		return nil, err
	}
	responseOffset, err := dereferencePointer(pointerToResponseOffset)
	if err != nil {
		return nil, err
	}
	responseSize, err := dereferencePointer(pointerToResponseSize)
	if err != nil {
		return nil, err
	}
	response, err := dereferencePointer(&pointer{
		offset: uint64(binary.LittleEndian.Uint32(responseOffset)),
		len:    uint64(binary.LittleEndian.Uint32(responseSize)),
	})
	if hasError[0] == 1 {
		return nil, fmt.Errorf(string(response))
	}
	str := string(response)
	return &str, nil
}

// ReleaseClient releases memory in core associated to the given client ID.
func ReleaseClient(ctx context.Context, clientId uint64) {
	releaseClient := coreWASMModule.ExportedFunction(releaseClientFuncName)
	allocate := coreWASMModule.ExportedFunction(allocateFuncName)
	deallocate := coreWASMModule.ExportedFunction(deallocateFuncName)

	idPtr, _ := initializeWASMUInt64(ctx, clientId, allocate)

	defer deallocate.Call(ctx, idPtr.offset, idPtr.len)

	releaseClient.Call(ctx, idPtr.offset)
}

// httpGet is a callback which defines how http request are sent by the core.
func httpGet(ctx context.Context, m api.Module, requestOffset, requestLen, responseLen uint32) uint64 {
	allocate := m.ExportedFunction(allocateFuncName)
	deallocate := m.ExportedFunction(deallocateFuncName)

	req, err := dereferencePointer(&pointer{
		offset: uint64(requestOffset),
		len:    uint64(requestLen),
	})
	if err != nil {
		log.Panicf(err.Error())
	}

	var httpRequest http.Request
	err = json.Unmarshal(req, &httpRequest)
	if err != nil {
		errPtr, err1 := initializeWASMByteArray(ctx, []byte(err.Error()), allocate)
		defer deallocate.Call(ctx, errPtr.offset, errPtr.len)
		if err1 != nil {
			log.Panicf(err1.Error())
		}
		if !m.Memory().WriteUint32Le(responseLen, uint32(errPtr.len)) {
			log.Panicf("Could not write error length")
		}
		return errPtr.offset
	}
	resp, err := http.DefaultClient.Do(&httpRequest)
	if err != nil {
		errPtr, err1 := initializeWASMByteArray(ctx, []byte(err.Error()), allocate)
		defer deallocate.Call(ctx, errPtr.offset, errPtr.len)
		if err1 != nil {
			log.Panicf(err1.Error())
		}
		if !m.Memory().WriteUint32Le(responseLen, uint32(errPtr.len)) {
			log.Panicf("Could not write error length")
		}
		return errPtr.offset
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf(err.Error())
	}

	resPtr, err := initializeWASMByteArray(ctx, resBody, allocate)
	if err != nil {
		log.Panicf(err.Error())
	}

	if !m.Memory().WriteUint32Le(responseLen, uint32(len(resBody))) {
		log.Panicf("Could not write response body length")
	}
	return resPtr.offset
}

type pointer struct {
	offset uint64
	len    uint64
}

// initializeWASMByteArray writes bytes into core's memory and returns a pointer to its address.
func initializeWASMByteArray(ctx context.Context, value []byte, allocateFunc api.Function) (*pointer, error) {
	size := uint64(len(value))

	ptr, err := allocateFunc.Call(ctx, size)
	if err != nil {
		return nil, err
	}

	returnedOffset := ptr[0]

	// The pointer is a linear memory offset, which is where we write the name.
	if !coreWASMModule.Memory().Write(uint32(returnedOffset), value) {
		return nil, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d",
			returnedOffset, size, coreWASMModule.Memory().Size())
	}

	return &pointer{
		offset: ptr[0],
		len:    size,
	}, nil
}

// initializeWASMUInt64 writes a long integer into core's memory and returns a pointer to its address.
func initializeWASMUInt64(ctx context.Context, number uint64, allocateFunc api.Function) (*pointer, error) {
	const uint64Len = 8
	pointerBytes := make([]byte, uint64Len)
	binary.LittleEndian.AppendUint64(pointerBytes, number)
	return initializeWASMByteArray(ctx, pointerBytes, allocateFunc)
}

// initializeWASMUInt32 writes an integer into core's memory and returns a pointer to its address.
func initializeWASMUInt32(ctx context.Context, number uint32, allocateFunc api.Function) (*pointer, error) {
	const uint32Len = 4
	pointerBytes := make([]byte, uint32Len)
	binary.LittleEndian.AppendUint32(pointerBytes, number)
	return initializeWASMByteArray(ctx, pointerBytes, allocateFunc)
}

// dereferencePointer returns the value the given pointer references
func dereferencePointer(ptr *pointer) ([]byte, error) {
	value, ok := coreWASMModule.Memory().Read(uint32(ptr.offset), uint32(ptr.len))
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			ptr.offset, ptr.len, coreWASMModule.Memory().Size())
	}
	return value, nil
}
