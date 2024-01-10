package main

import (
	"context"
	"os"
	"strconv"
	"time"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example for retrieving a secret from 1Password and setting it as SECRET_ENV_VAR using the SDK client.

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	println("Initializing client")
	beforeInit := time.Now().UnixNano()
	client, err := onepassword.Client(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
		onepassword.WithContext(context.Background()),
	)
	afterInit := time.Now().UnixNano()

	println("Total time to initialize client: " + strconv.FormatInt(afterInit-beforeInit, 10))
	println("Done initializing client")

	if err != nil {
		panic("A " + err.Error())
	}
	for i := 0; i < 5; i++ {
		println("Making invocation #" + strconv.Itoa(i))
		beforeInvocation := time.Now().UnixNano()
		secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/password")
		if err != nil {
			panic("B " + err.Error())
		}
		afterInvocation := time.Now().UnixNano()
		println("Secret: " + *secret)
		println("Finished invocation #" + strconv.Itoa(i))
		println("Total time for invocation: " + strconv.FormatInt(afterInvocation-beforeInvocation, 10))
	}

}

// Before running test, update the httpRequest function from `go/pkg/mod/github.com/extism/go-sdk@v1.0.0-rc3/host.go`` to:
/*
func httpRequest(ctx context.Context, m api.Module, requestOffset uint64, bodyOffset uint64) uint64 {
	if plugin, ok := ctx.Value("plugin").(*Plugin); ok {
		cp := plugin.currentPlugin()

		requestJson, err := cp.ReadBytes(requestOffset)
		var request HttpRequest
		err = json.Unmarshal(requestJson, &request)
		if err != nil {
			panic(fmt.Errorf("Invalid HTTP Request: %v", err))
		}

		url, err := url.Parse(request.Url)
		if err != nil {
			panic(fmt.Errorf("Invalid Url: %v", err))
		}

		// deny all requests by default
		hostMatches := false
		for _, allowedHost := range plugin.AllowedHosts {
			if allowedHost == url.Hostname() {
				hostMatches = true
				break
			}

			pattern := glob.MustCompile(allowedHost)
			if pattern.Match(url.Hostname()) {
				hostMatches = true
				break
			}
		}

		if !hostMatches {
			panic(fmt.Errorf("HTTP request to '%v' is not allowed", request.Url))
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
		beforeReq := time.Now().UnixNano()

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		afterReq := time.Now().UnixNano()

		fmt.Printf("Total time to make request: %d\n", afterReq-beforeReq)

		plugin.LastStatusCode = resp.StatusCode

		// TODO: make this limit configurable
		// TODO: the rust implementation silently truncates the response body, should we keep the behavior here?
		limiter := http.MaxBytesReader(nil, resp.Body, 1024*1024*50)
		body, err := io.ReadAll(limiter)
		if err != nil {
			panic(err)
		}

		if len(body) == 0 {
			return 0
		} else {
			offset, err := cp.WriteBytes(body)
			if err != nil {
				panic("Failed to write resposne body to memory")
			}

			return offset
		}
	}

	panic("Invalid context, `plugin` key not found")
}

*/
