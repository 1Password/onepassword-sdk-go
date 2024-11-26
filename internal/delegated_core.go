package internal

import (
	"context"
	"encoding/json"
	"io"
	"net"
)

type DelegatedCore struct {
	connection net.Conn
}

type DelegatedCoreMessage struct {
	FFIMethod string          `json:"ffi_method"`
	Payload   json.RawMessage `json:"payload"`
}

func NewDelegatedCore() *DelegatedCore {
	c, err := net.Dial("unix", "/Users/andititu/echo.sock")
	if err != nil {
		panic(err)
	}
	return &DelegatedCore{
		connection: c,
	}
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c *DelegatedCore) InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	message := DelegatedCoreMessage{
		FFIMethod: "init_client",
		Payload:   json.RawMessage(marshaledConfig),
	}

	serializedMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	_, err = c.connection.Write(serializedMessage)
	if err != nil {
		panic(err)
	}

	res, err := io.ReadAll(c.connection)
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

// Invoke calls specified business logic from core
func (c *DelegatedCore) Invoke(ctx context.Context, invokeConfig InvokeConfig) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}

	message := DelegatedCoreMessage{
		FFIMethod: "invoke",
		Payload:   json.RawMessage(input),
	}

	serializedMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	_, err = c.connection.Write(serializedMessage)
	if err != nil {
		panic(err)
	}

	res, err := io.ReadAll(c.connection)
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (c *DelegatedCore) ReleaseClient(clientID uint64) {
	marshaledClientID, err := json.Marshal(clientID)
	if err != nil {
		panic(err)
	}

	message := DelegatedCoreMessage{
		FFIMethod: "release_client",
		Payload:   json.RawMessage(marshaledClientID),
	}

	serializedMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	_, err = c.connection.Write(serializedMessage)
	if err != nil {
		panic(err)
	}

	_, err = io.ReadAll(c.connection)
	if err != nil {
		panic(err)
	}
}
