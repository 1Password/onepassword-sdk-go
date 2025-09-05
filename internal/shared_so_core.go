package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"
)

/*
#cgo LDFLAGS: -ldl

#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>

// Define a function pointer type matching Rust's send_message
typedef int32_t (*send_message_t)(
    const uint8_t* msg_ptr,
    size_t msg_len,
    uint8_t** out_buf,
    size_t* out_len
);

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
}

// Find1PasswordPath tries to return the default installation path of 1Password
func find1PasswordPath() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		// 1Password is typically installed as an app bundle
		paths := []string{
			"/Applications/1Password.app",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}

	case "linux":
		// Linux installations vary
		paths := []string{
			"/var/lib/1password",
			"/opt/1Password",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}

	case "windows":
		// On Windows, 1Password installs under Program Files
		paths := []string{
			`C:\Program Files\1Password\1Password.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("1Password not found")
}

func NewSharedLibCore() (*SharedLibCore, error) {
	path, err := find1PasswordPath()
	if err != nil {
		return nil, err
	}
	coreLib, err := LoadCore(path)
	if err != nil {
		return nil, err
	}

	return coreLib, nil
}

func LoadCore(path string) (*SharedLibCore, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.open_library(cPath)
	if handle == nil {
		return nil, errors.New("failed to open library")
	}

	symbol := C.CString("send_message")
	defer C.free(unsafe.Pointer(symbol))

	fn := C.load_symbol(handle, symbol)
	if fn == nil {
		C.close_library(handle)
		return nil, errors.New("failed to load send_message")
	}

	return &SharedLibCore{
		handle:      handle,
		sendMessage: (C.send_message_t)(fn),
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

	retCode := c.sendMessage(
		(*C.uint8_t)(unsafe.Pointer(&input[0])),
		C.size_t(len(input)),
		&outBuf,
		&outLen,
	)

	if retCode != 0 {
		return nil, fmt.Errorf("send_message failed: %d", int(retCode))
	}

	resp := C.GoBytes(unsafe.Pointer(outBuf), C.int(outLen))
	C.free(unsafe.Pointer(outBuf))

	return resp, nil
}
