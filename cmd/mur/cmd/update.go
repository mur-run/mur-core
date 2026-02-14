package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
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
	Short: "Update everything (binary + skills + hooks)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔄 Updating mur...")
		fmt.Println()

		// Update binary
		if err := updateBinary(); err != nil {
			fmt.Printf("⚠ Binary update failed: %v\n", err)
		}

		// Update skills
		if err := updateSkillDefinitions(); err != nil {
			fmt.Printf("⚠ Skills update failed: %v\n", err)
		}

		// Update hooks
		if err := updateHookTemplates(); err != nil {
			fmt.Printf("⚠ Hooks update failed: %v\n", err)
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

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateAllCmd)
	updateCmd.AddCommand(updateBinaryCmd)
	updateCmd.AddCommand(updateSkillsCmd)
	updateCmd.AddCommand(updateHooksCmd)

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
		// If homebrew upgrade fails (e.g., already latest), try update first
		if installMethod == "homebrew" {
			fmt.Println("  ℹ️  Running brew update first...")
			updateCmd := exec.Command("brew", "update")
			updateCmd.Run()
			
			// Retry upgrade
			retryCmd := exec.Command("brew", "upgrade", "mur")
			retryCmd.Stdout = os.Stdout
			retryCmd.Stderr = os.Stderr
			if retryErr := retryCmd.Run(); retryErr != nil {
				// Check if already up to date
				fmt.Println("  ✓ mur is already up to date")
				return nil
			}
		} else {
			return fmt.Errorf("failed to update binary: %w", err)
		}
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
		"/opt/homebrew/",      // Apple Silicon
		"/usr/local/Cellar/",  // Intel Mac
		"/home/linuxbrew/",    // Linux Homebrew
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
