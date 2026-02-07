package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var (
	initNonInteractive bool
	initHooks          bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mur configuration",
	Long: `Initialize mur with an interactive setup wizard.

Use --non-interactive for scripted setup (uses defaults).`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Skip interactive prompts, use defaults")
	initCmd.Flags().BoolVar(&initHooks, "hooks", false, "Install Claude Code hooks (non-interactive mode)")
}

// CLI tool configuration
type cliTool struct {
	Name       string
	Binary     string
	Installed  bool
	HooksSupport bool
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	murDir := filepath.Join(home, ".mur")

	// Non-interactive mode
	if initNonInteractive {
		return runNonInteractiveInit(home, murDir)
	}

	// Interactive mode
	return runInteractiveInit(home, murDir)
}

func runInteractiveInit(home, murDir string) error {
	fmt.Println()
	fmt.Println("🚀 Welcome to mur!")
	fmt.Println()

	// Detect installed CLIs
	tools := detectCLIs()

	// Show detected tools
	var installedNames []string
	for _, t := range tools {
		if t.Installed {
			installedNames = append(installedNames, t.Name)
		}
	}

	if len(installedNames) > 0 {
		fmt.Printf("Detected AI CLIs: %s\n", strings.Join(installedNames, ", "))
		fmt.Println()
	}

	// Select which CLIs to use
	var selectedCLIs []string
	cliOptions := []string{}
	for _, t := range tools {
		status := ""
		if t.Installed {
			status = " (installed)"
		}
		cliOptions = append(cliOptions, t.Name+status)
	}

	cliPrompt := &survey.MultiSelect{
		Message: "Which AI CLIs do you want to use?",
		Options: cliOptions,
		Default: installedNames,
	}
	if err := survey.AskOne(cliPrompt, &selectedCLIs); err != nil {
		return err
	}

	// Clean up selection (remove " (installed)" suffix)
	for i, s := range selectedCLIs {
		selectedCLIs[i] = strings.TrimSuffix(s, " (installed)")
	}

	// Check if Claude is selected and ask about hooks
	installHooks := false
	claudeSelected := contains(selectedCLIs, "Claude Code")
	if claudeSelected {
		hookPrompt := &survey.Confirm{
			Message: "Install Claude Code hooks for real-time learning?",
			Default: true,
		}
		if err := survey.AskOne(hookPrompt, &installHooks); err != nil {
			return err
		}
	}

	// Ask for default CLI
	defaultCLI := ""
	if len(selectedCLIs) > 0 {
		defaultPrompt := &survey.Select{
			Message: "Which CLI should be the default?",
			Options: selectedCLIs,
			Default: selectedCLIs[0],
		}
		if err := survey.AskOne(defaultPrompt, &defaultCLI); err != nil {
			return err
		}
	}

	// Create directories
	fmt.Println()
	dirs := []string{
		murDir,
		filepath.Join(murDir, "patterns"),
		filepath.Join(murDir, "hooks"),
		filepath.Join(murDir, "transcripts"),
		filepath.Join(murDir, "tracking"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	fmt.Println("✓ Created ~/.mur/ directory")

	// Create config
	if err := createConfig(murDir, selectedCLIs, defaultCLI); err != nil {
		return err
	}
	fmt.Println("✓ Created config.yaml")

	// Install hooks if requested
	if installHooks {
		if err := installClaudeHooks(home, murDir); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	}

	// Ask about learning repo
	fmt.Println()
	if err := SetupLearningRepo(home); err != nil {
		fmt.Printf("  ⚠ Warning: %v\n", err)
	}

	// Sync patterns to all selected CLIs
	fmt.Println()
	fmt.Println("Syncing patterns to CLIs...")
	results, err := sync.SyncPatternsToAllCLIs()
	if err != nil {
		fmt.Printf("  ⚠ Warning: %v\n", err)
	} else {
		for _, r := range results {
			if r.Success {
				fmt.Printf("  ✓ %s: %s\n", r.Target, r.Message)
			}
		}
	}

	// Final message
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ mur is ready!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  mur run -p \"your first task\"")
	if defaultCLI != "" {
		fmt.Printf("  # uses %s\n", defaultCLI)
	} else {
		fmt.Println()
	}
	fmt.Println("  mur stats                          # see your progress")
	fmt.Println()

	return nil
}

func runNonInteractiveInit(home, murDir string) error {
	// Create directories
	dirs := []string{
		murDir,
		filepath.Join(murDir, "patterns"),
		filepath.Join(murDir, "hooks"),
		filepath.Join(murDir, "transcripts"),
		filepath.Join(murDir, "tracking"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default config
	if err := createConfig(murDir, []string{"Claude Code"}, "Claude Code"); err != nil {
		return err
	}

	fmt.Println("✓ mur initialized at ~/.mur")

	// Install hooks if flag set
	if initHooks {
		if err := installClaudeHooks(home, murDir); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("Next: mur run -p \"your task\"")

	return nil
}

func detectCLIs() []cliTool {
	tools := []cliTool{
		{Name: "Claude Code", Binary: "claude", HooksSupport: true},
		{Name: "Gemini CLI", Binary: "gemini", HooksSupport: false},
		{Name: "Codex", Binary: "codex", HooksSupport: false},
		{Name: "Auggie", Binary: "auggie", HooksSupport: false},
		{Name: "Aider", Binary: "aider", HooksSupport: false},
	}

	for i := range tools {
		_, err := exec.LookPath(tools[i].Binary)
		tools[i].Installed = err == nil
	}

	return tools
}

func createConfig(murDir string, selectedCLIs []string, defaultCLI string) error {
	configPath := filepath.Join(murDir, "config.yaml")

	// Map CLI names to config keys
	cliMap := map[string]string{
		"Claude Code": "claude",
		"Gemini CLI":  "gemini",
		"Codex":       "codex",
		"Auggie":      "auggie",
		"Aider":       "aider",
	}

	defaultKey := "claude"
	if key, ok := cliMap[defaultCLI]; ok {
		defaultKey = key
	}

	// Build tools section
	toolsYaml := ""
	for name, key := range cliMap {
		enabled := contains(selectedCLIs, name)
		toolsYaml += fmt.Sprintf("  %s:\n    enabled: %t\n    binary: %s\n", key, enabled, key)
	}

	config := fmt.Sprintf(`# mur Configuration
# https://github.com/mur-run/mur-core

# Default AI CLI
default_tool: %s

# Available tools
tools:
%s
# Learning settings
learning:
  auto_extract: true
  sync_to_tools: true

# Routing
routing:
  mode: auto  # auto | manual | cost-first
`, defaultKey, toolsYaml)

	return os.WriteFile(configPath, []byte(config), 0644)
}

func installClaudeHooks(home, murDir string) error {
	// Create on-prompt-reminder.md
	reminderPath := filepath.Join(murDir, "hooks", "on-prompt-reminder.md")
	reminderContent := `[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a pattern), save it:

  mur learn add --name "pattern-name" --content "description"

Or create a file in ~/.mur/patterns/

Only save if: it required discovery, it helps future tasks, and it's verified.
`
	if err := os.WriteFile(reminderPath, []byte(reminderContent), 0644); err != nil {
		return err
	}

	// Create on-stop.sh
	stopScriptPath := filepath.Join(murDir, "hooks", "on-stop.sh")
	stopScript := `#!/bin/bash
# Sync patterns after session
mur sync patterns --quiet 2>/dev/null || true
`
	if err := os.WriteFile(stopScriptPath, []byte(stopScript), 0755); err != nil {
		return err
	}

	// Update Claude settings
	claudeSettingsPath := filepath.Join(home, ".claude", "settings.json")

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
					{"type": "command", "command": fmt.Sprintf("bash %s", stopScriptPath)},
				},
			},
		},
	}

	var settings map[string]interface{}
	if data, err := os.ReadFile(claudeSettingsPath); err == nil {
		json.Unmarshal(data, &settings)
	} else {
		os.MkdirAll(filepath.Join(home, ".claude"), 0755)
		settings = make(map[string]interface{})
	}

	// Backup
	if _, err := os.Stat(claudeSettingsPath); err == nil {
		backupPath := claudeSettingsPath + ".backup"
		if data, err := os.ReadFile(claudeSettingsPath); err == nil {
			os.WriteFile(backupPath, data, 0644)
		}
	}

	// Merge hooks
	existingHooks, _ := settings["hooks"].(map[string]interface{})
	if existingHooks == nil {
		existingHooks = make(map[string]interface{})
	}

	for event, eventHooks := range hooks {
		if _, exists := existingHooks[event]; !exists {
			existingHooks[event] = eventHooks
		}
	}
	settings["hooks"] = existingHooks

	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(claudeSettingsPath, data, 0644); err != nil {
		return err
	}

	fmt.Println("✓ Installed Claude Code hooks")
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
