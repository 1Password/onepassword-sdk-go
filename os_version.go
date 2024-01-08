package onepassword

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const defaultVersionString = "0.0.0"

// OSVersion returns the version of the OS the client is running on
func OSVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return osVersionDarwin()
	case "windows":
		return osVersionWindows()
	case "freebsd",
		"netbsd",
		"openbsd",
		"solaris":
		return osVersionUnix()
	case "linux":
		return osVersionLinux()
	default:
		return defaultVersionString
	}
}

func osVersionDarwin() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return defaultVersionString
	}
	version := out.String()
	if len(version) > 0 {
		return version
	}
	return defaultVersionString
}

func osVersionWindows() string {
	cmd := exec.Command("cmd", "/c", "ver")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return defaultVersionString
	}
	return windowsVersionFromVerOutput(out.String())
}

func osVersionUnix() string {
	cmd := exec.Command("sysctl", "-n", "kern.osrelease")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return defaultVersionString
	}
	version := out.String()
	if len(version) > 0 {
		return version
	}
	return defaultVersionString
}

func osVersionLinux() string {
	filepath := "/etc/os-release"
	file, err := os.ReadFile(filepath)
	if err != nil {
		return defaultVersionString
	}

	version := linuxVersionFromOSReleaseOutput(string(file))
	if len(version) > 0 {
		return version
	}

	return defaultVersionString
}

// windowsVersionFromVerOutput takes something like "Microsoft Windows XP [Version 5.1.2600]" and returns "5.1.2600"
func windowsVersionFromVerOutput(verOutput string) string {
	marker := "[Version "
	markerLocation := strings.Index(verOutput, marker)
	if markerLocation == -1 {
		return "0.0.0"
	}
	postMarkerString := verOutput[markerLocation+len(marker):]
	// remove trailing space padding
	postMarkerString = strings.TrimSpace(postMarkerString)
	// remove closing square bracket
	versionString := strings.TrimRight(postMarkerString, "]")
	return versionString
}

func linuxVersionFromOSReleaseOutput(osRelease string) string {
	lines := strings.Split(osRelease, "\n")

	for _, l := range lines {
		if strings.HasPrefix(l, "VERSION_ID") {
			versionWithQuotes := strings.Split(l, "=")[1]
			version := strings.Trim(versionWithQuotes, "\"")

			return version
		}
	}

	return ""
}
