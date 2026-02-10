package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage connected devices",
	Long:  `List, view, and manage devices connected to your mur account.`,
	RunE:  runDevices,
}

var devicesLogoutCmd = &cobra.Command{
	Use:   "logout [device-name]",
	Short: "Force logout a device",
	Long:  `Force logout a device by its name or device ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDevicesLogout,
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.AddCommand(devicesLogoutCmd)
}

func runDevices(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	resp, err := client.ListDevices()
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	fmt.Println("📱 Registered Devices")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(resp.Devices) == 0 {
		fmt.Println("  No devices registered yet.")
		return nil
	}

	// Get current device ID
	currentDevice := cloud.GetDeviceInfo()

	fmt.Printf("  %-25s %-10s %s\n", "DEVICE", "OS", "LAST ACTIVE")
	fmt.Println()

	for _, d := range resp.Devices {
		status := ""
		if d.IsActive {
			status = " ⚡"
		}
		if d.DeviceID == currentDevice.DeviceID {
			status += " (this device)"
		}

		lastActive := formatLastActive(d.LastActiveAt)
		fmt.Printf("  %-25s %-10s %s%s\n", d.DeviceName, d.OS, lastActive, status)
	}

	fmt.Println()
	fmt.Printf("Limit: %d/%d concurrent devices (%s plan)\n", resp.ActiveCount, resp.DeviceLimit, resp.Plan)

	return nil
}

func runDevicesLogout(cmd *cobra.Command, args []string) error {
	deviceName := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	// First list devices to find the one to logout
	resp, err := client.ListDevices()
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	var targetDevice *cloud.Device
	for _, d := range resp.Devices {
		if d.DeviceName == deviceName || d.DeviceID == deviceName {
			targetDevice = &d
			break
		}
	}

	if targetDevice == nil {
		return fmt.Errorf("device not found: %s", deviceName)
	}

	if err := client.LogoutDevice(targetDevice.DeviceID); err != nil {
		return fmt.Errorf("failed to logout device: %w", err)
	}

	fmt.Printf("✓ Logged out \"%s\"\n", targetDevice.DeviceName)
	return nil
}

func formatLastActive(lastActiveAt string) string {
	if lastActiveAt == "" {
		return "Unknown"
	}

	t, err := time.Parse(time.RFC3339, lastActiveAt)
	if err != nil {
		return lastActiveAt
	}

	since := time.Since(t)

	if since < time.Minute {
		return "Active now"
	} else if since < time.Hour {
		return fmt.Sprintf("%d min ago", int(since.Minutes()))
	} else if since < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(since.Hours()))
	} else {
		return fmt.Sprintf("%d days ago", int(since.Hours()/24))
	}
}
