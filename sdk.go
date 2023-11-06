package sdk

/*
#cgo LDFLAGS: -ldl
#include <stdlib.h>
#include <dlfcn.h>
typedef char* (*init_client_func)(char* saToken, int* returnsErr);
typedef char* (*invoke_func)(unsigned long long clientID, char* method, char* parameters, int* returnsError);
typedef char* (*release_client_func)(unsigned long long clientID, int* returnsError);

char* call_init_client(void* func_ptr, char* saToken, int* returnsErr) {
    return ((init_client_func)func_ptr)(saToken, returnsErr);
}

char* call_invoke(void* func_ptr, unsigned long long clientID, char* method, char* parameters, int* returnsError) {
    return ((invoke_func)func_ptr)(clientID, method, parameters, returnsError);
}

char* call_release_client(void* func_ptr, unsigned long long clientID, int* returnsError) {
    return ((release_client_func)func_ptr)(clientID, returnsError);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"unsafe"

	core "github.com/1Password/1password-sdk-core"
)

var initClient unsafe.Pointer
var invoke unsafe.Pointer
var releaseClient unsafe.Pointer

// init loads the core shared library object for the current platform as well as its exported functions.
func init() {
	var err error
	libPath := core.SharedLibraryPath()
	cPath := C.CString(libPath)
	defer C.free(unsafe.Pointer(cPath))
	lib := C.dlopen(cPath, C.RTLD_LAZY)

	initClient, err = loadFunction(lib, "initClient")
	if err != nil {
		panic(err)
	}
	invoke, err = loadFunction(lib, "invoke")
	if err != nil {
		panic(err)
	}
	releaseClient, err = loadFunction(lib, "releaseClient")
	if err != nil {
		panic(err)
	}
}

// opClient The client instance.
type opClient struct {
	id uint64
}

// loadFunction loads function handle from the shared library
func loadFunction(dllHandle unsafe.Pointer, funcName string) (unsafe.Pointer, error) {
	cFuncName := C.CString(funcName)
	defer C.free(unsafe.Pointer(cFuncName))
	funcPtr := C.dlsym(dllHandle, cFuncName)
	if funcPtr == nil {
		err := C.GoString(C.dlerror())
		return nil, fmt.Errorf("failed to find the %s function: %s", funcName, err)
	}
	return funcPtr, nil
}

// NewServiceAccountClient constructor for `opClient`.
func NewServiceAccountClient(saToken string) (*opClient, error) {
	returnsErr := C.int(0)
	//TODO: find a way to defer C.free(unsafe.Pointer(&returnsErr))

	cSAToken := C.CString(saToken)
	defer C.free(unsafe.Pointer(cSAToken))

	response := C.call_init_client(initClient, cSAToken, &returnsErr)
	defer C.free(unsafe.Pointer(response))
	if int(returnsErr) == 1 {
		return nil, errors.New(C.GoString(response))
	}

	clientID, err := strconv.ParseUint(C.GoString(response), 10, 64)
	if err != nil {
		return nil, err
	}
	client := &opClient{id: clientID}

	runtime.SetFinalizer(client, func(f *opClient) {
		returnsError := C.int(0)
		//TODO: find a way to defer C.free(unsafe.Pointer(&returnsErr))

		releaseResponse := C.call_release_client(releaseClient, C.ulonglong(clientID), &returnsError)
		if int(returnsErr) == 1 {
			panic(errors.New(C.GoString(releaseResponse)))
		}
	})

	return client, nil
}

// NewServiceAccountClientFromEnv constructor for `opClient` from the environment.
func NewServiceAccountClientFromEnv() (*opClient, error) {
	const tokenEnvVar = "OP_SERVICE_ACCOUNT_TOKEN"
	token, ok := os.LookupEnv(tokenEnvVar)
	if !ok {
		return nil, fmt.Errorf("no variable %s was found in the enviroment", tokenEnvVar)
	}

	return NewServiceAccountClient(token)
}

// callFFI calls the appropriate function into the core shared library with the given parameters
func (c *opClient) callFFI(method string, parameters string) (*string, error) {
	returnsErr := C.int(0)
	//TODO: find a way to defer C.free(unsafe.Pointer(&returnsErr))

	cID := C.ulonglong(c.id)
	//TODO: find a way to defer C.free(unsafe.Pointer(&cID))

	cMethod := C.CString(method)
	defer C.free(unsafe.Pointer(cMethod))

	cParameters := C.CString(parameters)
	defer C.free(unsafe.Pointer(cParameters))

	response := C.call_invoke(invoke, cID, cMethod, cParameters, &returnsErr)
	defer C.free(unsafe.Pointer(response))
	output := C.GoString(response)
	if int(returnsErr) == 1 {
		return nil, errors.New(output)
	}
	return &output, nil
}

// Resolve returns a secret pointed to by the given secret reference
func (c *opClient) Resolve(secretReference string) ([]byte, error) {
	response, err := c.callFFI("Resolve", secretReference)
	if err != nil {
		return nil, err
	}
	secret := []byte(*response)
	return secret, nil
}
