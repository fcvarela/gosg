package core

import "runtime"

// Platform is an OS
type Platform uint8

// Supported platforms
const (
	PlatformDarwin  Platform = iota
	PlatformLinux   Platform = iota
	PlatformWindows Platform = iota
)

var (
	platform Platform
)

func init() {
	switch runtime.GOOS {
	case "darwin":
		platform = PlatformDarwin
	case "linux":
		platform = PlatformLinux
	case "windows":
		platform = PlatformWindows
	}
}

// GetPlatform returns the current platform.
func GetPlatform() Platform {
	return platform
}
