//go:build !cgo && (darwin || linux)

package onepassword

import "fmt"

// WithDesktopAppIntegration specifies a client should use the desktop app to authenticate. Set to your 1Password account name as shown at the top left sidebar of the app, or your account UUID.
func WithDesktopAppIntegration(accountName string) ClientOption {
	return func(c *Client) error {
		return fmt.Errorf("the desktop app integration feature requires CGO (CGO_ENABLED=1 and a working C toolchain - see README.md for how to build this)")
	}
}
