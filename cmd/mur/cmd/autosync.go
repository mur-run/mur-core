package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mur-run/mur-core/internal/config"
)

var autoSyncCmd = &cobra.Command{
	Use:   "auto",
	Short: "Manage automatic sync",
	Long: `Configure automatic pattern sync.

When enabled, mur will automatically sync patterns at regular intervals.

Examples:
  mur sync auto enable     # Enable auto-sync (interactive setup)
  mur sync auto disable    # Disable auto-sync
  mur sync auto status     # Check auto-sync status`,
}

var autoSyncEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable automatic sync",
	RunE:  runAutoSyncEnable,
}

var autoSyncDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable automatic sync",
	RunE:  runAutoSyncDisable,
}

var autoSyncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check auto-sync status",
	RunE:  runAutoSyncStatus,
}

func init() {
	syncCmd.AddCommand(autoSyncCmd)
	autoSyncCmd.AddCommand(autoSyncEnableCmd)
	autoSyncCmd.AddCommand(autoSyncDisableCmd)
	autoSyncCmd.AddCommand(autoSyncStatusCmd)
}

func runAutoSyncEnable(cmd *cobra.Command, args []string) error {
	// Ask for sync interval
	intervalOptions := []string{
		"Every 15 minutes",
		"Every 30 minutes",
		"Every hour",
		"Every 6 hours",
		"Every 24 hours",
	}

	var intervalChoice string
	prompt := &survey.Select{
		Message: "How often should mur sync?",
		Options: intervalOptions,
		Default: "Every 30 minutes",
	}
	if err := survey.AskOne(prompt, &intervalChoice); err != nil {
		return err
	}

	// Parse interval
	intervalMinutes := 30
	switch intervalChoice {
	case "Every 15 minutes":
		intervalMinutes = 15
	case "Every 30 minutes":
		intervalMinutes = 30
	case "Every hour":
		intervalMinutes = 60
	case "Every 6 hours":
		intervalMinutes = 360
	case "Every 24 hours":
		intervalMinutes = 1440
	}

	// Save to config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	cfg.Sync.Auto = true
	cfg.Sync.IntervalMinutes = intervalMinutes

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".mur", "config.yaml")
	data, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0644)

	// Install platform-specific scheduler
	switch runtime.GOOS {
	case "darwin":
		return installMacOSLaunchAgent(intervalMinutes)
	case "linux":
		return installLinuxSystemdTimer(intervalMinutes)
	case "windows":
		return installWindowsTaskScheduler(intervalMinutes)
	default:
		fmt.Printf("⚠️  Auto-sync not supported on %s\n", runtime.GOOS)
		fmt.Println("Add 'mur sync --quiet' to your crontab manually")
		return nil
	}
}

func runAutoSyncDisable(cmd *cobra.Command, args []string) error {
	// Update config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	cfg.Sync.Auto = false

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".mur", "config.yaml")
	data, _ := yaml.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0644)

	// Remove platform-specific scheduler
	switch runtime.GOOS {
	case "darwin":
		return uninstallMacOSLaunchAgent()
	case "linux":
		return uninstallLinuxSystemdTimer()
	case "windows":
		return uninstallWindowsTaskScheduler()
	default:
		fmt.Println("✓ Auto-sync disabled in config")
		return nil
	}
}

func runAutoSyncStatus(cmd *cobra.Command, args []string) error {
	cfg, _ := config.Load()

	fmt.Println("🔄 Auto-sync Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━")

	if cfg != nil && cfg.Sync.Auto {
		fmt.Println("Status: ✅ Enabled")
		fmt.Printf("Interval: Every %d minutes\n", cfg.Sync.IntervalMinutes)
	} else {
		fmt.Println("Status: ❌ Disabled")
	}
	fmt.Println()

	// Check platform-specific status
	switch runtime.GOOS {
	case "darwin":
		checkMacOSLaunchAgent()
	case "linux":
		checkLinuxSystemdTimer()
	case "windows":
		checkWindowsTaskScheduler()
	}

	return nil
}

// ============ macOS LaunchAgent ============

const macOSPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>run.mur.sync</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.MurPath}}</string>
        <string>sync</string>
        <string>--quiet</string>
    </array>
    <key>StartInterval</key>
    <integer>{{.IntervalSeconds}}</integer>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`

func installMacOSLaunchAgent(intervalMinutes int) error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "run.mur.sync.plist")
	logPath := filepath.Join(home, ".mur", "sync.log")

	// Find mur binary path
	murPath, err := exec.LookPath("mur")
	if err != nil {
		// Try common locations
		possiblePaths := []string{
			"/usr/local/bin/mur",
			"/opt/homebrew/bin/mur",
			filepath.Join(home, "go", "bin", "mur"),
			filepath.Join(home, ".local", "bin", "mur"),
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				murPath = p
				break
			}
		}
		if murPath == "" {
			return fmt.Errorf("mur binary not found in PATH")
		}
	}

	// Create plist from template
	tmpl, err := template.New("plist").Parse(macOSPlistTemplate)
	if err != nil {
		return err
	}

	data := struct {
		MurPath         string
		IntervalSeconds int
		LogPath         string
	}{
		MurPath:         murPath,
		IntervalSeconds: intervalMinutes * 60,
		LogPath:         logPath,
	}

	// Ensure LaunchAgents directory exists
	_ = os.MkdirAll(filepath.Dir(plistPath), 0755)

	f, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	// Unload if already loaded, then load
	_ = exec.Command("launchctl", "unload", plistPath).Run()
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load launch agent: %w", err)
	}

	fmt.Println("✅ Auto-sync enabled (macOS LaunchAgent)")
	fmt.Printf("   Interval: Every %d minutes\n", intervalMinutes)
	fmt.Printf("   Plist: %s\n", plistPath)
	fmt.Printf("   Log: %s\n", logPath)

	return nil
}

func uninstallMacOSLaunchAgent() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "run.mur.sync.plist")

	_ = exec.Command("launchctl", "unload", plistPath).Run()
	_ = os.Remove(plistPath)

	fmt.Println("✅ Auto-sync disabled (macOS LaunchAgent removed)")
	return nil
}

func checkMacOSLaunchAgent() {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "run.mur.sync.plist")

	if _, err := os.Stat(plistPath); err == nil {
		fmt.Println("LaunchAgent: ✅ Installed")
		fmt.Printf("  Path: %s\n", plistPath)

		// Check if loaded
		output, _ := exec.Command("launchctl", "list", "run.mur.sync").Output()
		if len(output) > 0 {
			fmt.Println("  Status: Running")
		}
	} else {
		fmt.Println("LaunchAgent: ❌ Not installed")
	}
}

// ============ Linux systemd ============

const linuxTimerTemplate = `[Unit]
Description=MUR Pattern Sync Timer

[Timer]
OnBootSec=5min
OnUnitActiveSec={{.IntervalMinutes}}min
Persistent=true

[Install]
WantedBy=timers.target
`

const linuxServiceTemplate = `[Unit]
Description=MUR Pattern Sync

[Service]
Type=oneshot
ExecStart={{.MurPath}} sync --quiet
`

func installLinuxSystemdTimer(intervalMinutes int) error {
	home, _ := os.UserHomeDir()
	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	timerPath := filepath.Join(systemdDir, "mur-sync.timer")
	servicePath := filepath.Join(systemdDir, "mur-sync.service")

	// Find mur binary
	murPath, err := exec.LookPath("mur")
	if err != nil {
		murPath = filepath.Join(home, "go", "bin", "mur")
	}

	// Create systemd user directory
	_ = os.MkdirAll(systemdDir, 0755)

	// Write timer
	timerTmpl, _ := template.New("timer").Parse(linuxTimerTemplate)
	timerFile, err := os.Create(timerPath)
	if err != nil {
		return err
	}
	_ = timerTmpl.Execute(timerFile, struct{ IntervalMinutes int }{intervalMinutes})
	timerFile.Close()

	// Write service
	serviceTmpl, _ := template.New("service").Parse(linuxServiceTemplate)
	serviceFile, err := os.Create(servicePath)
	if err != nil {
		return err
	}
	_ = serviceTmpl.Execute(serviceFile, struct{ MurPath string }{murPath})
	serviceFile.Close()

	// Enable and start timer
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	_ = exec.Command("systemctl", "--user", "enable", "mur-sync.timer").Run()
	_ = exec.Command("systemctl", "--user", "start", "mur-sync.timer").Run()

	fmt.Println("✅ Auto-sync enabled (systemd user timer)")
	fmt.Printf("   Interval: Every %d minutes\n", intervalMinutes)
	fmt.Printf("   Timer: %s\n", timerPath)

	return nil
}

func uninstallLinuxSystemdTimer() error {
	home, _ := os.UserHomeDir()
	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	timerPath := filepath.Join(systemdDir, "mur-sync.timer")
	servicePath := filepath.Join(systemdDir, "mur-sync.service")

	_ = exec.Command("systemctl", "--user", "stop", "mur-sync.timer").Run()
	_ = exec.Command("systemctl", "--user", "disable", "mur-sync.timer").Run()
	os.Remove(timerPath)
	os.Remove(servicePath)
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("✅ Auto-sync disabled (systemd timer removed)")
	return nil
}

func checkLinuxSystemdTimer() {
	output, err := exec.Command("systemctl", "--user", "is-active", "mur-sync.timer").Output()
	status := strings.TrimSpace(string(output))

	if err == nil && status == "active" {
		fmt.Println("systemd timer: ✅ Active")
	} else {
		fmt.Println("systemd timer: ❌ Not active")
	}
}

// ============ Windows Task Scheduler ============

func installWindowsTaskScheduler(intervalMinutes int) error {
	// Find mur binary
	murPath, err := exec.LookPath("mur.exe")
	if err != nil {
		home, _ := os.UserHomeDir()
		murPath = filepath.Join(home, "go", "bin", "mur.exe")
	}

	// Create scheduled task using schtasks
	taskName := "MUR_Sync"

	// Delete existing task if any
	_ = exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()

	// Create new task
	cmd := exec.Command("schtasks", "/create",
		"/tn", taskName,
		"/tr", fmt.Sprintf(`"%s" sync --quiet`, murPath),
		"/sc", "minute",
		"/mo", fmt.Sprintf("%d", intervalMinutes),
		"/ru", os.Getenv("USERNAME"),
		"/f",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create task: %s", output)
	}

	fmt.Println("✅ Auto-sync enabled (Windows Task Scheduler)")
	fmt.Printf("   Interval: Every %d minutes\n", intervalMinutes)
	fmt.Printf("   Task: %s\n", taskName)

	return nil
}

func uninstallWindowsTaskScheduler() error {
	taskName := "MUR_Sync"
	_ = exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()

	fmt.Println("✅ Auto-sync disabled (Windows task removed)")
	return nil
}

func checkWindowsTaskScheduler() {
	taskName := "MUR_Sync"
	output, err := exec.Command("schtasks", "/query", "/tn", taskName).CombinedOutput()

	if err == nil && strings.Contains(string(output), taskName) {
		fmt.Println("Task Scheduler: ✅ Task exists")
	} else {
		fmt.Println("Task Scheduler: ❌ Task not found")
	}
}
