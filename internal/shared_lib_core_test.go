//go:build cgo || windows

package internal

import (
	"runtime"
	"testing"
)

func TestDesktopLibLocations(t *testing.T) {
	home := t.TempDir()

	t.Run("env override has priority", func(t *testing.T) {
		t.Setenv("OP_DESKTOP_APP_LIB_PATH", "/custom/location/libop_sdk_ipc_client.so")
		locations, err := desktopLibLocations(home)
		if err != nil {
			t.Fatal(err)
		}
		if len(locations) == 0 || locations[0] != "/custom/location/libop_sdk_ipc_client.so" {
			t.Fatalf("expected override as first location, got %v", locations)
		}
	})

	t.Run("empty env override falls back to defaults", func(t *testing.T) {
		t.Setenv("OP_DESKTOP_APP_LIB_PATH", "")
		locations, err := desktopLibLocations(home)
		if err != nil {
			t.Fatal(err)
		}
		if len(locations) == 0 {
			t.Fatal("expected default locations")
		}
		switch runtime.GOOS {
		case "linux":
			want := "/usr/bin/1password/libop_sdk_ipc_client.so"
			if locations[0] != want {
				t.Fatalf("expected %s as first location, got %s", want, locations[0])
			}
			var foundNixOS bool
			for _, l := range locations {
				if l == "/run/current-system/sw/share/1password/libop_sdk_ipc_client.so" {
					foundNixOS = true
				}
			}
			if !foundNixOS {
				t.Fatal("missing NixOS system profile location")
			}
		}
	})

	t.Run("unsupported OS returns error", func(t *testing.T) {
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			t.Skip("test only applies to unsupported GOOS values")
		}
		if _, err := desktopLibLocations(home); err == nil {
			t.Fatal("expected an error for unsupported OS")
		}
	})
}
