package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

/*
#cgo LDFLAGS: -ldl

#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>

// Function pointer types matching Rust exports
typedef int32_t (*send_message_t)(
    const uint8_t* msg_ptr,
    size_t msg_len,
    uint8_t** out_buf,
    size_t* out_len,
    size_t* out_cap
);

typedef void (*free_message_t)(
    uint8_t* buf,
    size_t len,
    size_t cap
);

// Trampoline wrappers so Go can call the function pointers
static inline int32_t call_send_message(
    send_message_t fn,
    const uint8_t* msg_ptr,
    size_t msg_len,
    uint8_t** out_buf,
    size_t* out_len,
    size_t* out_cap
) {
    return fn(msg_ptr, msg_len, out_buf, out_len, out_cap);
}

static inline void call_free_message(
    free_message_t fn,
    uint8_t* buf,
    size_t len,
    size_t cap
) {
    fn(buf, len, cap);
}

// dlopen wrapper
static void* open_library(const char* path) {
    return dlopen(path, RTLD_NOW);
}

// dlsym wrapper
static void* load_symbol(void* handle, const char* name) {
    return dlsym(handle, name);
}

// dlclose wrapper
static int close_library(void* handle) {
    return dlclose(handle);
}
*/
import "C"

type SharedLibCore struct {
	handle      unsafe.Pointer
	sendMessage C.send_message_t
	freeMessage C.free_message_t
}

var coreLib *SharedLibCore

// find1PasswordLibPath returns the path to the 1Password shared library
// (libop_sdk_ipc_client.dylib/.so/.dll) depending on OS.
func find1PasswordLibPath() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		// Typical locations for the 1Password bundle
		executables := []string{
			"/Applications/1Password/Contents/MacOS/1Password",
		}

		for _, exe := range executables {
			if _, err := os.Stat(exe); err == nil {
				// Replace the executable name with the dylib name
				libPath := filepath.Join(filepath.Dir(exe), "libop_sdk_ipc_client.dylib")
				if _, err := os.Stat(libPath); err == nil {
					return libPath, nil
				}
			}
		}

	case "linux":
		// On Linux, it might live under /opt/1Password or similar
		candidates := []string{
			"/opt/1Password/libop_sdk_ipc_client.so",
			"/usr/lib/1password/libop_sdk_ipc_client.so",
		}
		for _, lib := range candidates {
			if _, err := os.Stat(lib); err == nil {
				return lib, nil
			}
		}

	case "windows":
		// On Windows, shared libs are DLLs in the install directory
		executables := []string{
			`C:\Program Files\1Password\1Password.exe`,
		}
		for _, exe := range executables {
			if _, err := os.Stat(exe); err == nil {
				libPath := filepath.Join(filepath.Dir(exe), "op_sdk_ipc_client.dll")
				if _, err := os.Stat(libPath); err == nil {
					return libPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("1Password desktop application not found")
}

func GetSharedLibCore() (*SharedLibCore, error) {
	if coreLib == nil {
		path, err := find1PasswordLibPath()
		if err != nil {
			return nil, err
		}
		coreLib, err = loadCore(path)
		if err != nil {
			return nil, err
		}
	}

	return coreLib, nil
}

func loadCore(path string) (*SharedLibCore, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.open_library(cPath)
	if handle == nil {
		return nil, errors.New("failed to open library")
	}

	symbol := C.CString("op_sdk_ipc_send_message")
	defer C.free(unsafe.Pointer(symbol))

	fnSend := C.load_symbol(handle, symbol)
	if fnSend == nil {
		C.close_library(handle)
		return nil, errors.New("failed to load send_message")
	}

	symbolFree := C.CString("op_sdk_ipc_free_response")
	defer C.free(unsafe.Pointer(symbolFree))

	fnFree := C.load_symbol(handle, symbolFree)
	if fnFree == nil {
		C.close_library(handle)
		return nil, errors.New("failed to load free_message")
	}

	return &SharedLibCore{
		handle:      handle,
		sendMessage: (C.send_message_t)(fnSend),
		freeMessage: (C.free_message_t)(fnFree),
	}, nil
}

func (c *SharedLibCore) Close() {
	if c.handle != nil {
		C.close_library(c.handle)
		c.handle = nil
	}
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c *SharedLibCore) InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	res, err := c.callSharedLibrary(marshaledConfig)
	if err != nil {
		return nil, err
	}
	var id uint64
	err = json.Unmarshal(res, &id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (c *SharedLibCore) Invoke(ctx context.Context, invokeConfig InvokeConfig) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}

	res, err := c.callSharedLibrary(input)
	if err != nil {
		return nil, err
	}

	response := string(res)
	return &response, nil
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (c *SharedLibCore) ReleaseClient(clientID uint64) {
	marshaledClientID, err := json.Marshal(clientID)
	if err != nil {
		log.Println("failed to marshal clientID")
	}
	_, err = c.callSharedLibrary(marshaledClientID)
	if err != nil {
		log.Println("failed to release client")
	}
}

func (c *SharedLibCore) callSharedLibrary(input []byte) ([]byte, error) {
	var outBuf *C.uint8_t
	var outLen C.size_t
	var outCap C.size_t

	retCode := C.call_send_message(
		c.sendMessage,
		(*C.uint8_t)(unsafe.Pointer(&input[0])),
		C.size_t(len(input)),
		&outBuf,
		&outLen,
		&outCap,
	)

	if retCode != 0 {
		return nil, fmt.Errorf("send_message failed: %d", int(retCode))
	}

	resp := C.GoBytes(unsafe.Pointer(outBuf), C.int(outLen))
	// Call trampoline with the function pointer
	C.call_free_message(c.freeMessage, outBuf, outLen, outCap)

	return resp, nil
}
