package cloud

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DeviceInfo holds device identification
type DeviceInfo struct {
	DeviceID   string
	DeviceName string
	OS         string
}

// GetDeviceInfo returns the current device info
func GetDeviceInfo() *DeviceInfo {
	return &DeviceInfo{
		DeviceID:   getOrCreateDeviceID(),
		DeviceName: getDeviceName(),
		OS:         runtime.GOOS,
	}
}

// getOrCreateDeviceID returns a persistent device ID
func getOrCreateDeviceID() string {
	// Try to load from file first
	configDir := getMurConfigDir()
	deviceFile := filepath.Join(configDir, "device_id")

	if data, err := os.ReadFile(deviceFile); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id
		}
	}

	// Generate new device ID
	id := generateDeviceID()

	// Save for future use
	_ = os.MkdirAll(configDir, 0700)
	_ = os.WriteFile(deviceFile, []byte(id), 0600)

	return id
}

// generateDeviceID creates a unique device identifier
func generateDeviceID() string {
	hostname, _ := os.Hostname()
	u, _ := user.Current()
	username := ""
	if u != nil {
		username = u.Username
	}

	// Create a hash from hostname + username + timestamp
	data := fmt.Sprintf("%s:%s:%d", hostname, username, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // 32 char hex string
}

// getDeviceName returns a human-readable device name
func getDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown Device"
	}

	// TODO: Try to get a nicer name on macOS via scutil --get ComputerName

	return hostname
}

// getMurConfigDir returns the mur config directory
func getMurConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mur")
}

// Device represents a device from the server
type Device struct {
	ID           string `json:"id"`
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	OS           string `json:"os"`
	LastActiveAt string `json:"last_active_at"`
	IsActive     bool   `json:"is_active"`
}

// DeviceListResponse is the response from /devices
type DeviceListResponse struct {
	Devices     []Device `json:"devices"`
	ActiveCount int      `json:"active_count"`
	DeviceLimit int      `json:"device_limit"`
	Plan        string   `json:"plan"`
}

// DeviceLimitError is returned when device limit is exceeded
type DeviceLimitError struct {
	Limit   int      `json:"limit"`
	Active  []Device `json:"active"`
	Message string   `json:"message"`
}

func (e *DeviceLimitError) Error() string {
	return e.Message
}
