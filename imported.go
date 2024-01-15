package onepassword

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

// ImportedFunctions returns all functions 1Password SDK core must import.
func ImportedFunctions() []extism.HostFunction {
	return []extism.HostFunction{randomFillFunc(), httpRequestFunc()}
}

// randomFillFunc returns an Extism Function that writes random bytes into the WASM core's memory using crypto/rand.
func randomFillFunc() extism.HostFunction {
	randomFill := extism.NewHostFunctionWithStack("random_fill_imported", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		ptr := api.DecodeU32(stack[0])
		length := api.DecodeU32(stack[1])

		b := make([]byte, length)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		p.Memory().Write(ptr, b)
	}, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{})
	randomFill.SetNamespace("op-random")

	return randomFill
}

func httpRequestFunc() extism.HostFunction {
	const requestBodySizeLimit = 1024 * 1024 * 50
	httpRequest := extism.NewHostFunctionWithStack("http_request_imported", func(ctx context.Context, cp *extism.CurrentPlugin, stack []uint64) {
		reqPtr := stack[0]
		bodyOffset := stack[1]
		responseLen := stack[2]
		statusRaw := stack[3]

		requestJson, err := cp.ReadBytes(reqPtr)
		if err != nil {
			panic(fmt.Errorf("invalid request %v", err))
		}
		var request extism.HttpRequest
		err = json.Unmarshal(requestJson, &request)
		if err != nil {
			panic(fmt.Errorf("invalid HTTP Request: %v", err))
		}

		var bodyReader io.Reader = nil
		if bodyOffset != 0 {
			body, err := cp.ReadBytes(bodyOffset)
			if err != nil {
				panic("Failed to read response body from memory")
			}

			cp.Free(bodyOffset)

			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, request.Method, request.Url, bodyReader)
		if err != nil {
			panic(err)
		}

		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		limiter := http.MaxBytesReader(nil, resp.Body, requestBodySizeLimit)
		body, err := io.ReadAll(limiter)
		if err != nil {
			panic(err)
		}

		ok := cp.Memory().WriteUint16Le(uint32(statusRaw), uint16(resp.StatusCode))
		if !ok {
			panic("memory out of range")
		}
		ok = cp.Memory().WriteUint64Le(uint32(responseLen), uint64(len(body)))
		if !ok {
			panic("memory out of range")
		}
		if len(body) == 0 {
			stack[0] = 0
		} else {
			bodyOffset, err := cp.WriteBytes(body)
			if err != nil {
				panic("Failed to write resposne body to memory")
			}
			stack[0] = bodyOffset
		}
	}, []api.ValueType{api.ValueTypeI64, api.ValueTypeI64, api.ValueTypeI64, api.ValueTypeI64}, []api.ValueType{api.ValueTypeI32})
	httpRequest.SetNamespace("op-sdk-core")

	return httpRequest
}
