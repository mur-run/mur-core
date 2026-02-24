package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/config"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update mur components",
	Long: `Update mur binary, skills, and hooks.

Examples:
  mur update          # Update everything
  mur update binary   # Only update mur binary
  mur update skills   # Only update skill definitions
  mur update hooks    # Only update hooks`,
}

var updateAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Update everything (binary + config + skills + hooks)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating mur...")
		fmt.Println()

		// 1. Update binary
		fmt.Println("1. Binary")
		if err := updateBinary(); err != nil {
			fmt.Printf("   ⚠ Binary update failed: %v\n", err)
		}

		// 2. Update config
		fmt.Println("\n2. Config")
		if err := updateConfig(); err != nil {
			fmt.Printf("   ⚠ Config update failed: %v\n", err)
		}

		// 3. Update skills
		fmt.Println("\n3. Skills")
		if err := updateSkillDefinitions(); err != nil {
			fmt.Printf("   ⚠ Skills update failed: %v\n", err)
		}

		// 4. Update hooks
		fmt.Println("\n4. Hooks")
		if err := updateHookTemplates(); err != nil {
			fmt.Printf("   ⚠ Hooks update failed: %v\n", err)
		}

		fmt.Println()
		fmt.Println("✅ Update complete!")
		return nil
	},
}

var updateBinaryCmd = &cobra.Command{
	Use:   "binary",
	Short: "Update mur binary to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating mur binary...")
		if err := updateBinary(); err != nil {
			return err
		}
		fmt.Println("✅ Binary updated!")
		return nil
	},
}

var updateSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Update skill definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating skill definitions...")
		if err := updateSkillDefinitions(); err != nil {
			return err
		}
		fmt.Println("✅ Skills updated!")
		return nil
	},
}

var updateHooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Update hook templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating hook templates...")
		if err := updateHookTemplates(); err != nil {
			return err
		}
		fmt.Println("✅ Hooks updated!")
		return nil
	},
}

var updateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Migrate config to latest schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating config...")
		if err := updateConfig(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateAllCmd)
	updateCmd.AddCommand(updateBinaryCmd)
	updateCmd.AddCommand(updateSkillsCmd)
	updateCmd.AddCommand(updateHooksCmd)
	updateCmd.AddCommand(updateConfigCmd)

	// Make "mur update" without subcommand run "all"
	updateCmd.RunE = updateAllCmd.RunE
}

func updateBinary() error {
	// Detect installation method by checking binary path
	installMethod := detectInstallMethod()

	var cmd *exec.Cmd
	switch installMethod {
	case "homebrew":
		fmt.Println("  📦 Detected Homebrew installation")
		// Force-refresh the tap to ensure we see the latest release
		fmt.Println("  ↻ Refreshing tap...")
		refreshTap := exec.Command("brew", "tap", "--force", "mur-run/tap")
		refreshTap.Stdout = os.Stdout
		refreshTap.Stderr = os.Stderr
		if err := refreshTap.Run(); err != nil {
			// Fallback: full brew update (slower but reliable)
			fmt.Println("  ↻ Full brew update...")
			fullUpdate := exec.Command("brew", "update")
			fullUpdate.Stdout = os.Stdout
			fullUpdate.Stderr = os.Stderr
			_ = fullUpdate.Run()
		}
		cmd = exec.Command("brew", "upgrade", "mur")
	case "go":
		fmt.Println("  🐹 Detected Go installation")
		cmd = exec.Command("go", "install", "github.com/mur-run/mur-core/cmd/mur@latest")
	default:
		fmt.Println("  🐹 Using Go install (default)")
		cmd = exec.Command("go", "install", "github.com/mur-run/mur-core/cmd/mur@latest")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if installMethod == "homebrew" {
			// brew upgrade returns error when already up to date
			fmt.Println("  ✓ mur is already up to date")
			return nil
		}
		return fmt.Errorf("failed to update binary: %w", err)
	}
	fmt.Println("  ✓ mur binary updated to latest")
	return nil
}

// detectInstallMethod checks how mur was installed
func detectInstallMethod() string {
	// Get the path of the current executable
	exePath, err := exec.LookPath("mur")
	if err != nil {
		return "unknown"
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Check for Homebrew paths
	homebrewPaths := []string{
		"/opt/homebrew/",     // Apple Silicon
		"/usr/local/Cellar/", // Intel Mac
		"/home/linuxbrew/",   // Linux Homebrew
	}

	for _, prefix := range homebrewPaths {
		if len(realPath) >= len(prefix) && realPath[:len(prefix)] == prefix {
			return "homebrew"
		}
	}

	// Check for Go bin paths
	home, _ := os.UserHomeDir()
	goPaths := []string{
		filepath.Join(home, "go", "bin"),
		"/usr/local/go/bin",
	}

	for _, goPath := range goPaths {
		if len(realPath) >= len(goPath) && realPath[:len(goPath)] == goPath {
			return "go"
		}
	}

	// Check GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		gopathBin := filepath.Join(gopath, "bin")
		if len(realPath) >= len(gopathBin) && realPath[:len(gopathBin)] == gopathBin {
			return "go"
		}
	}

	return "unknown"
}

func updateConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	oldVersion := cfg.SchemaVersion
	defaults := config.Default()
	merged := config.MergeConfig(cfg, defaults)

	changed, changes := config.MigrateConfig(merged)
	if changed {
		if err := merged.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("   ✓ Migrated: v%d → v%d\n", oldVersion, merged.SchemaVersion)
		for _, c := range changes {
			fmt.Printf("   + Added: %s (%s)\n", c.Field, c.Description)
		}
	} else {
		fmt.Println("   ✓ Config already up to date")
	}

	return nil
}

func updateSkillDefinitions() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	murDir := filepath.Join(home, ".mur")
	skillsDir := filepath.Join(murDir, "skills")

	// Create skills directory
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}

	// Write default skill template
	defaultSkill := `# mur Learning Skill

This skill helps you learn from development sessions.

## Instructions

When you discover something non-obvious during development:
1. Note the pattern, fix, or technique
2. Run: mur learn add --name "pattern-name" --content "description"

Patterns are automatically synced to all your AI CLIs.
`
	skillPath := filepath.Join(skillsDir, "learning.md")
	if err := os.WriteFile(skillPath, []byte(defaultSkill), 0644); err != nil {
		return err
	}

	fmt.Println("  ✓ Skill definitions updated")
	return nil
}

func updateHookTemplates() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	murDir := filepath.Join(home, ".mur")
	hooksDir := filepath.Join(murDir, "hooks")

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Update on-prompt-reminder.md
	reminderContent := `[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a pattern), save it:

  mur learn add --name "pattern-name" --content "description"

Or create a file in ~/.mur/patterns/

Only save if: it required discovery, it helps future tasks, and it's verified.
`
	reminderPath := filepath.Join(hooksDir, "on-prompt-reminder.md")
	if err := os.WriteFile(reminderPath, []byte(reminderContent), 0644); err != nil {
		return err
	}

	// Update on-stop.sh
	stopScript := `#!/bin/bash
# Sync patterns after session
mur sync patterns --quiet 2>/dev/null || true
`
	stopPath := filepath.Join(hooksDir, "on-stop.sh")
	if err := os.WriteFile(stopPath, []byte(stopScript), 0755); err != nil {
		return err
	}

	// Re-install hooks to Claude settings
	claudeSettingsPath := filepath.Join(home, ".claude", "settings.json")
	if _, err := os.Stat(claudeSettingsPath); err == nil {
		// Read existing settings
		data, err := os.ReadFile(claudeSettingsPath)
		if err != nil {
			return err
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			return err
		}

		// Update hook paths
		hooks := map[string]interface{}{
			"UserPromptSubmit": []map[string]interface{}{
				{
					"matcher": "",
					"hooks": []map[string]interface{}{
						{"type": "command", "command": fmt.Sprintf("cat %s >&2", reminderPath)},
					},
				},
			},
			"Stop": []map[string]interface{}{
				{
					"matcher": "",
					"hooks": []map[string]interface{}{
						{"type": "command", "command": fmt.Sprintf("bash %s", stopPath)},
					},
				},
			},
		}

		settings["hooks"] = hooks

		// Write back
		newData, _ := json.MarshalIndent(settings, "", "  ")
		if err := os.WriteFile(claudeSettingsPath, newData, 0644); err != nil {
			return err
		}
	}

	fmt.Println("  ✓ Hook templates updated")
	return nil
}
