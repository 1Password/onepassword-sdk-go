package onepassword

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	// "fmt"

	"runtime"

	// "runtime"

	"github.com/1password/onepassword-sdk-go/internal"
)

const (
	DefaultIntegrationName    = "Unknown"
	DefaultIntegrationVersion = "Unknown"
	socketPath                = "/Users/omarmiraj/echo.sock"
)

// type SDKClient interface{
// 	NewClient(ctx context.Context, opts ...ClientOption) (*Client, error)
// }

// type ServiceAccountClient struct{}

// type UserAuthClient struct{}

// func (sa ServiceAccountClient) NewClient(name string) string {
// 	core, err := internal.GetSharedCore()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return createClient(ctx, opts...)
// }


// func (sa ServiceAccountClient) NewClient(name string) string {
// 	return createClient(ctx, opts...)
// }

// NewClient returns a 1Password Go SDK client using the provided ClientOption list.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	// core, err := internal.GetSharedCore()
	// if err != nil {
	// 	return nil, err
	// }
	return createClient(ctx, opts...)
}

func createClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	client := Client{
		config: internal.NewDefaultAuthPromptConfig(),
	}

	for _, opt := range opts {
		err := opt(&client)
		if err != nil {
			return nil, err
		}
	}
	// cmd := exec.Command("/Applications/op", "sdksession")
	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	panic(string(out))
	// }
	// key := string(out);
	// fmt.Println("this is the key", key)
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	// fmt.Println("Server is listening on", socketPath)
	jsonConfig, err := json.Marshal(client.config)
	if err != nil {
		panic(err)
	}
	message := &Message{
		FFIMethod: "initClient",
		Payload:   jsonConfig,
	}
	msgConfig, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	// Frame requires length of payload then the payload
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(msgConfig)))
	buf.Write(msgConfig)
	conn.Write(buf.Bytes())
	// fmt.Println("Sent:", string(msgConfig))
	b := make([]byte, 12)
	conn.Read(b)
	fmt.Println("Received:", string(b))
	fmt.Println("Bytes Received:",b)
	// b contains 8 bytes, the first 4 is the length of the payload and the last 4 is the actual payload
	// we are interested in just the payload so we ignore the length
	// payload:= b[4:]
	// clientID := binary.BigEndian.Uint32(payload)

	inner := internal.InnerClient{
		// ID: uint64(clientID),
		Connection: &conn,
	}
	initAPIs(&client, inner)

	runtime.SetFinalizer(&client, func(f *Client) {
		conn.Close()
	})
	return &client, nil
}

type ClientOption func(client *Client) error

// WithServiceAccountToken specifies the [1Password Service Account](https://developer.1password.com/docs/service-accounts) token to use to authenticate the SDK client. Read more about how to get started with service accounts: https://developer.1password.com/docs/service-accounts/get-started/#create-a-service-account
func WithServiceAccountToken(token string) ClientOption {
	return func(c *Client) error {
		c.config.SAToken = token
		return nil
	}
}

// WithIntegrationInfo specifies the name and version of the integration built using the 1Password Go SDK. If you don't know which name and version to use, use `DefaultIntegrationName` and `DefaultIntegrationVersion`, respectively.
func WithIntegrationInfo(name string, version string) ClientOption {
	return func(c *Client) error {
		c.config.IntegrationName = name
		c.config.IntegrationVersion = version
		return nil
	}
}

func clientInvoke(ctx context.Context, innerClient internal.InnerClient, invocation string, params map[string]interface{}) (*string, error) {
	invocationConfig := internal.Parameters{
				MethodName:       invocation,
				SerializedParams: params,
			}
	conn := *innerClient.Connection
	if conn == nil {
		return nil, errors.New("connection is nil")
	}

	marsheldInvoke, err := json.Marshal(invocationConfig)
	if err != nil {
		panic(err)
	}
	message := &Message{
		FFIMethod: "invoke",
		Payload:   marsheldInvoke,
	}
	marsheldMsg, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(len(marsheldMsg))) // Length prefix
	buf.Write(marsheldMsg)
	conn.Write(buf.Bytes())
	// fmt.Println("Sent:", string(marsheldMsg))
	b := make([]byte, 30)
	conn.Read(b)
	fmt.Println("Bytes Received",b)
	fmt.Println("Received:", string(b))
	invocationResponse := string(b)

	return &invocationResponse, nil
}

type Message struct {
	FFIMethod string          `json:"ffiMethod"`
	Payload   json.RawMessage `json:"payload"`
}
