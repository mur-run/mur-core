package cloud

import (
	"os"
	"runtime"
)

// CanOpenBrowser returns true if the current environment likely supports opening a browser.
func CanOpenBrowser() bool {
	// SSH sessions can't open a local browser
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		return false
	}

	switch runtime.GOOS {
	case "darwin", "windows":
		return true
	case "linux":
		// Need a display server
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	default:
		return false
	}
}
