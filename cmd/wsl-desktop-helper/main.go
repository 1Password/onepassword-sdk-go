//go:build windows

// Stdout is exclusively protocol data. This helper uses the ordinary Windows
// desktop SDK backend; it does not implement or bypass desktop IPC itself.
// Build: GOOS=windows GOARCH=amd64 go build ./cmd/wsl-desktop-helper
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/1password/onepassword-sdk-go/internal"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "" {
		fmt.Fprintln(os.Stderr, "usage: wsl-desktop-helper.exe ACCOUNT")
		os.Exit(2)
	}
	core, err := internal.GetSharedLibCore(os.Args[1])
	if err == nil {
		err = internal.ServeDesktopBridge(context.Background(), core.InnerCore, os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Windows desktop helper:", err)
		os.Exit(1)
	}
}
