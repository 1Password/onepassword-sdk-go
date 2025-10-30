//go:build windows

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"unsafe"

	"golang.org/x/sys/windows"
)

type SharedLibCore struct {
	accountName string
	dll         *windows.DLL
	procSend    *windows.Proc
	procFree    *windows.Proc
}

var coreLib *SharedLibCore

// Request/Response mirror your Unix file (kept identical)
type Request struct {
	Kind        string `json:"kind"`
	AccountName string `json:"account_name"`
	Payload     []byte `json:"payload"`
}

type Response struct {
	Success bool   `json:"success"`
	Payload []byte `json:"payload"`
}

func (r Response) Error() string { return string(r.Payload) }

// find1PasswordLibPath returns the path to the DLL on Windows
func find1PasswordLibPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	locations := []string{
		path.Join(home, `AppData\Local\1Password\app\8\op_sdk_ipc_client.dll`),
		`C:\Program Files\1Password\op_sdk_ipc_client.dll`,
		`C:\Program Files (x86)\1Password\op_sdk_ipc_client.dll`,
	}
	for _, p := range locations {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("1Password desktop application not found")
}

// API identical to Unix version
func GetSharedLibCore(accountName string) (*CoreWrapper, error) {
	if coreLib == nil {
		path, err := find1PasswordLibPath()
		if err != nil {
			return nil, err
		}
		coreLib, err = loadCore(path)
		if err != nil {
			return nil, err
		}
		coreLib.accountName = accountName
	}
	coreWrapper := CoreWrapper{InnerCore: coreLib}
	return &coreWrapper, nil
}

func loadCore(path string) (*SharedLibCore, error) {
	dll, err := windows.LoadDLL(path) // absolute path avoids search path surprises
	if err != nil {
		return nil, err
	}
	send, err := dll.FindProc("op_sdk_ipc_send_message")
	if err != nil {
		dll.Release()
		return nil, errors.New("failed to load send_message")
	}
	free, err := dll.FindProc("op_sdk_ipc_free_response")
	if err != nil {
		dll.Release()
		return nil, errors.New("failed to load free_message")
	}
	return &SharedLibCore{
		dll:      dll,
		procSend: send,
		procFree: free,
	}, nil
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (slc *SharedLibCore) InitClient(ctx context.Context, config []byte) ([]byte, error) {
	const kind = "init_client"
	req := Request{
		Kind:        kind,
		AccountName: slc.accountName,
		Payload:     config,
	}
	input, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return slc.callSharedLibrary(input)
}

func (slc *SharedLibCore) Invoke(ctx context.Context, invokeConfig []byte) ([]byte, error) {
	const kind = "invoke"
	req := Request{
		Kind:        kind,
		AccountName: slc.accountName,
		Payload:     invokeConfig,
	}
	input, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return slc.callSharedLibrary(input)
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (slc *SharedLibCore) ReleaseClient(clientID []byte) {
	_, err := slc.callSharedLibrary(clientID)
	if err != nil {
		log.Println("failed to release client")
	}
}

func (slc *SharedLibCore) callSharedLibrary(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("internal: empty input")
	}

	// Signature we’re calling (from your Rust exports):
	// int32_t op_sdk_ipc_send_message(const uint8_t* msg_ptr, size_t msg_len,
	//                                 uint8_t** out_buf, size_t* out_len, size_t* out_cap);
	var outBuf *byte
	var outLen uintptr
	var outCap uintptr

	r1, _, callErr := slc.procSend.Call(
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(len(input)),
		uintptr(unsafe.Pointer(&outBuf)),
		uintptr(unsafe.Pointer(&outLen)),
		uintptr(unsafe.Pointer(&outCap)),
	)
	// syscall layer error?
	if callErr != nil && callErr != windows.ERROR_SUCCESS {
		return nil, callErr
	}
	// library-level return code
	if int32(r1) != 0 {
		return nil, fmt.Errorf("failed to send message to Desktop App. Return code: %d", int32(r1))
	}

	// Copy response out of the DLL buffer, then free via exported function
	resp := unsafe.Slice(outBuf, outLen)
	out := make([]byte, outLen)
	copy(out, resp)

	// void op_sdk_ipc_free_response(uint8_t* buf, size_t len, size_t cap);
	_, _, _ = slc.procFree.Call(
		uintptr(unsafe.Pointer(outBuf)),
		outLen,
		outCap,
	)

	// Match Unix: decode envelope and return payload or error
	var response Response
	if err := json.Unmarshal(out, &response); err != nil {
		return nil, err
	}
	if response.Success {
		return response.Payload, nil
	}
	return nil, response
}
