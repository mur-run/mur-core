// Package notify provides notification functionality.
package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mur-run/mur-core/internal/config"
)

// Level represents notification severity.
type Level string

const (
	LevelInfo     Level = "info"
	LevelWarning  Level = "warning"
	LevelError    Level = "error"
	LevelCritical Level = "critical"
)

// SystemNotify sends a system notification (macOS Notification Center).
func SystemNotify(title, message string, level Level) error {
	cfg, err := config.Load()
	if err != nil {
		// If config fails, still try to notify for errors
		if level != LevelError && level != LevelCritical {
			return nil
		}
	} else if !cfg.Notifications.System {
		return nil
	}

	if runtime.GOOS != "darwin" {
		return nil // Only macOS supported for now
	}

	// Choose sound based on level
	sound := "default"
	switch level {
	case LevelError, LevelCritical:
		sound = "Basso"
	case LevelWarning:
		sound = "Purr"
	case LevelInfo:
		sound = "Pop"
	}

	// Try terminal-notifier first (better UX, supports click actions)
	if _, err := exec.LookPath("terminal-notifier"); err == nil {
		args := []string{
			"-title", title,
			"-message", message,
			"-sound", sound,
			"-group", "mur",
		}

		// Add sender for icon
		args = append(args, "-sender", "com.apple.Terminal")

		cmd := exec.Command("terminal-notifier", args...)
		if err := cmd.Run(); err == nil {
			return nil
		}
		// Fall through to osascript if terminal-notifier fails
	}

	// Fallback to osascript
	script := fmt.Sprintf(
		`display notification %q with title %q sound name %q`,
		escapeAppleScript(message),
		escapeAppleScript(title),
		sound,
	)

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// NotifyError sends an error notification.
func NotifyError(message string) error {
	return SystemNotify("mur: Error", message, LevelError)
}

// NotifyCritical sends a critical error notification.
func NotifyCritical(title, message string) error {
	return SystemNotify(title, message, LevelCritical)
}

// NotifySuccess sends a success notification.
func NotifySuccess(message string) error {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	if !cfg.Notifications.OnPatterns {
		return nil
	}
	return SystemNotify("mur: Success", message, LevelInfo)
}

// escapeAppleScript escapes a string for use in AppleScript.
func escapeAppleScript(s string) string {
	// Escape backslashes and quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	// Truncate long messages
	if len(s) > 200 {
		s = s[:197] + "..."
	}
	return s
}

// IsSystemNotifyAvailable returns true if system notifications are available.
func IsSystemNotifyAvailable() bool {
	return runtime.GOOS == "darwin"
}
