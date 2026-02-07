package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initHooks bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize murmur configuration",
	Long: `Initialize murmur by creating the config directory and default settings.

Use --hooks to also install Claude Code hooks for real-time learning.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initHooks, "hooks", false, "Install Claude Code hooks for real-time learning")
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	murDir := filepath.Join(home, ".mur")

	// Create directories
	dirs := []string{
		murDir,
		filepath.Join(murDir, "patterns"),
		filepath.Join(murDir, "hooks"),
		filepath.Join(murDir, "transcripts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default config
	configPath := filepath.Join(murDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `# Murmur Configuration
# https://github.com/mur-run/mur-core

# Default AI tool to use
default_tool: claude

# Available tools
tools:
  claude:
    enabled: true
    binary: claude
    flags: ["-p"]
  gemini:
    enabled: true
    binary: gemini
    flags: ["-p"]
  auggie:
    enabled: false
    binary: auggie
    flags: []

# Learning settings
learning:
  auto_extract: true
  sync_to_tools: true
  pattern_limit: 5  # Free tier limit

# MCP settings
mcp:
  sync_enabled: true
  servers: {}
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	fmt.Println("✓ Murmur initialized at ~/.mur")

	// Install hooks if requested
	if initHooks {
		fmt.Println()
		if err := installClaudeHooks(home, murDir); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  mur health     # Check tool availability")
	fmt.Println("  mur run -p \"your prompt\"  # Run a task")
	if !initHooks {
		fmt.Println("  mur init --hooks  # Install Claude Code hooks")
	}

	return nil
}

func installClaudeHooks(home, murDir string) error {
	fmt.Println("Installing Claude Code hooks...")

	// Create on-prompt-reminder.md
	reminderPath := filepath.Join(murDir, "hooks", "on-prompt-reminder.md")
	reminderContent := `[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a project-specific pattern, a configuration fix), save it using:

  mur learn add --name "pattern-name" --content "description of what you learned"

Or create a pattern file directly in ~/.mur/patterns/

Only extract if: it required actual discovery (not just docs), it will help with future tasks, and it has been verified to work. Skip trivial or well-documented things.

[SpecAwareness] If this project uses spec-driven development (check for openspec/ or .spec/ directories):
- Reference the current spec when making implementation decisions
- If you deviate from the spec, note WHY and save it as a pattern
`
	if err := os.WriteFile(reminderPath, []byte(reminderContent), 0644); err != nil {
		return fmt.Errorf("failed to create reminder file: %w", err)
	}

	// Create on-stop.sh
	stopScriptPath := filepath.Join(murDir, "hooks", "on-stop.sh")
	stopScript := `#!/bin/bash
# mur-core: on-stop hook
# Runs after each Claude Code session

# Check for new patterns and sync if needed
mur sync patterns --quiet 2>/dev/null || true
`
	if err := os.WriteFile(stopScriptPath, []byte(stopScript), 0755); err != nil {
		return fmt.Errorf("failed to create stop script: %w", err)
	}

	// Update Claude Code settings
	claudeSettingsPath := filepath.Join(home, ".claude", "settings.json")

	// Prepare hooks config
	hooks := map[string]interface{}{
		"UserPromptSubmit": []map[string]interface{}{
			{
				"matcher": "",
				"hooks": []map[string]interface{}{
					{
						"type":    "command",
						"command": fmt.Sprintf("cat %s >&2", reminderPath),
					},
				},
			},
		},
		"Stop": []map[string]interface{}{
			{
				"matcher": "",
				"hooks": []map[string]interface{}{
					{
						"type":    "command",
						"command": fmt.Sprintf("bash %s", stopScriptPath),
					},
				},
			},
		},
	}

	// Read existing settings or create new
	var settings map[string]interface{}
	if data, err := os.ReadFile(claudeSettingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	} else if os.IsNotExist(err) {
		// Create .claude directory
		if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
			return fmt.Errorf("failed to create .claude directory: %w", err)
		}
		settings = make(map[string]interface{})
	} else {
		return fmt.Errorf("failed to read settings: %w", err)
	}

	// Backup existing settings
	if _, err := os.Stat(claudeSettingsPath); err == nil {
		backupPath := claudeSettingsPath + ".backup"
		if data, err := os.ReadFile(claudeSettingsPath); err == nil {
			os.WriteFile(backupPath, data, 0644)
			fmt.Printf("  Backed up existing settings to %s\n", backupPath)
		}
	}

	// Merge hooks
	existingHooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		existingHooks = make(map[string]interface{})
	}

	// Add our hooks (check for duplicates)
	for event, eventHooks := range hooks {
		if existing, ok := existingHooks[event]; ok {
			// Check if mur hook already exists
			existingList, _ := existing.([]interface{})
			hasMur := false
			for _, h := range existingList {
				hMap, _ := h.(map[string]interface{})
				hooksList, _ := hMap["hooks"].([]interface{})
				for _, hook := range hooksList {
					hookMap, _ := hook.(map[string]interface{})
					if cmd, ok := hookMap["command"].(string); ok {
						if contains(cmd, ".mur") {
							hasMur = true
							break
						}
					}
				}
			}
			if !hasMur {
				// Append our hooks
				newHooks := eventHooks.([]map[string]interface{})
				for _, h := range newHooks {
					existingList = append(existingList, h)
				}
				existingHooks[event] = existingList
			}
		} else {
			existingHooks[event] = eventHooks
		}
	}

	settings["hooks"] = existingHooks

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(claudeSettingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	fmt.Println("✓ Claude Code hooks installed")
	fmt.Println()
	fmt.Println("Hooks installed:")
	fmt.Println("  • UserPromptSubmit — reminds Claude to save patterns")
	fmt.Println("  • Stop — syncs patterns after each session")

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
