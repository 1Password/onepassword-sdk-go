package internal

// OP_WSL_DESKTOP_BRIDGE opts Linux clients into a Windows helper at an absolute
// path. Private stdin/stdout carry SDK Core calls; desktop authorization remains
// in the existing Windows backend. Without the variable, behavior is unchanged.
import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const bridgeFrameLimit = 16 * 1024 * 1024

type bridgeRequest struct {
	Method  string `json:"method"`
	Payload []byte `json:"payload"`
}
type bridgeResponse struct {
	Payload []byte `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ServeDesktopBridge exposes SDK Core calls only on private stdin/stdout pipes.
// The Windows helper supplies its ordinary shared-library Core implementation.
func ServeDesktopBridge(ctx context.Context, core Core, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), bridgeFrameLimit)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		var request bridgeRequest
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			return errors.New("invalid desktop bridge request")
		}
		var response bridgeResponse
		var err error
		switch request.Method {
		case "init":
			response.Payload, err = core.InitClient(ctx, request.Payload)
		case "invoke":
			response.Payload, err = core.Invoke(ctx, request.Payload)
		case "release":
			core.ReleaseClient(request.Payload)
		default:
			err = errors.New("unknown desktop bridge method")
		}
		if err != nil {
			response.Error = err.Error()
		}
		if encoder.Encode(response) != nil {
			return errors.New("desktop bridge response failed")
		}
	}
	return scanner.Err()
}

type desktopBridgeCore struct {
	mu      sync.Mutex
	command *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Scanner
	closed  bool
}

func getDesktopBridgeCore(accountName, executable string) (*CoreWrapper, error) {
	if !filepath.IsAbs(executable) {
		return nil, errors.New("OP_WSL_DESKTOP_BRIDGE must be an absolute path to the reviewed Windows helper")
	}
	command := exec.Command(executable, accountName)
	command.Stderr = os.Stderr
	core, err := startDesktopBridge(command)
	if err != nil {
		return nil, err
	}
	return &CoreWrapper{InnerCore: core}, nil
}

func startDesktopBridge(command *exec.Cmd) (*desktopBridgeCore, error) {
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, err
	}
	if err := command.Start(); err != nil {
		input.Close()
		output.Close()
		return nil, fmt.Errorf("start Windows desktop helper: %w", err)
	}
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), bridgeFrameLimit)
	return &desktopBridgeCore{command: command, input: input, output: scanner}, nil
}

func (c *desktopBridgeCore) close() {
	if !c.closed {
		c.closed = true
		c.input.Close()
		c.command.Process.Kill()
		c.command.Wait()
	}
}

func (c *desktopBridgeCore) exchange(ctx context.Context, method string, payload []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, errors.New("desktop bridge is closed")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	frame, err := json.Marshal(bridgeRequest{Method: method, Payload: payload})
	if err != nil || len(frame)+1 >= bridgeFrameLimit {
		return nil, errors.New("desktop bridge request exceeds frame limit")
	}
	type result struct {
		response bridgeResponse
		err      error
	}
	done := make(chan result, 1)
	go func() {
		var r result
		if _, r.err = c.input.Write(append(frame, '\n')); r.err == nil {
			if !c.output.Scan() {
				r.err = errors.New("Windows desktop helper disconnected")
			} else if json.Unmarshal(c.output.Bytes(), &r.response) != nil {
				r.err = errors.New("invalid Windows desktop helper response")
			}
		}
		done <- r
	}()
	select {
	case <-ctx.Done():
		c.close()
		<-done
		return nil, ctx.Err()
	case r := <-done:
		if r.err != nil {
			c.close()
			return nil, r.err
		}
		if r.response.Error != "" {
			return nil, errors.New(r.response.Error)
		}
		return r.response.Payload, nil
	}
}

func (c *desktopBridgeCore) InitClient(ctx context.Context, config []byte) ([]byte, error) {
	payload, err := c.exchange(ctx, "init", config)
	if err != nil {
		c.mu.Lock()
		c.close()
		c.mu.Unlock()
	}
	return payload, err
}
func (c *desktopBridgeCore) Invoke(ctx context.Context, input []byte) ([]byte, error) {
	return c.exchange(ctx, "invoke", input)
}
func (c *desktopBridgeCore) ReleaseClient(id []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = c.exchange(ctx, "release", id)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.close()
}
