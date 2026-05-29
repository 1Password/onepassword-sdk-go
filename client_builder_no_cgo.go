//go:build !cgo && (darwin || linux)

package onepassword

import "errors"

// WithDesktopAppIntegration specifies a client should use the desktop app to authenticate. Set to your 1Password account name as shown at the top left sidebar of the app, or your account UUID.
func WithDesktopAppIntegration(accountName string) ClientOption {
	return func(c *Client) error {
		return errors.New("WithDesktopAppIntegration requires CGO to cross-compile. See README CGO section")
	}
}
