// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClaudeCodeHook represents a hook entry for Claude Code.
type ClaudeCodeHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// ClaudeCodeHooks represents the hooks configuration.
type ClaudeCodeHooks struct {
	PreToolUse map[string][]ClaudeCodeHook `json:"PreToolUse,omitempty"`
	PostToolUse map[string][]ClaudeCodeHook `json:"PostToolUse,omitempty"`
	UserPromptSubmit []ClaudeCodeHook `json:"UserPromptSubmit,omitempty"`
	Stop []ClaudeCodeHook `json:"Stop,omitempty"`
}

// ClaudeCodeSettings represents the Claude Code settings file.
type ClaudeCodeSettings struct {
	Hooks ClaudeCodeHooks `json:"hooks"`
}

// ClaudeCodeInstalled checks if Claude Code is installed.
func ClaudeCodeInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check if .claude directory exists
	claudeDir := filepath.Join(home, ".claude")
	_, err = os.Stat(claudeDir)
	return err == nil
}

// InstallClaudeCodeHooks installs mur hooks for Claude Code.
func InstallClaudeCodeHooks(enableSearch bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	murBin, err := findMurBinary()
	if err != nil {
		murBin = "mur"
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings
	var settings ClaudeCodeSettings
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			// If parsing fails, start fresh
			settings = ClaudeCodeSettings{}
		}
	}

	// Build hooks
	hooks := ClaudeCodeHooks{}

	// Learn hook (on Stop - captures learnings from session)
	learnHook := ClaudeCodeHook{
		Type:    "command",
		Command: fmt.Sprintf("%s learn --from-transcript 2>/dev/null || true", murBin),
	}
	hooks.Stop = []ClaudeCodeHook{learnHook}

	// Search hook (on UserPromptSubmit - suggests relevant patterns)
	if enableSearch {
		searchHook := ClaudeCodeHook{
			Type:    "command",
			Command: fmt.Sprintf(`%s search --inject "$PROMPT" 2>/dev/null || true`, murBin),
		}
		hooks.UserPromptSubmit = []ClaudeCodeHook{searchHook}
	}

	settings.Hooks = hooks

	// Ensure .claude directory exists
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		return fmt.Errorf("cannot create .claude directory: %w", err)
	}

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	fmt.Printf("✓ Installed Claude Code hooks at %s\n", settingsPath)
	if enableSearch {
		fmt.Println("  + Learn hook (captures patterns on stop)")
		fmt.Println("  + Search hook (suggests patterns on prompt)")
	} else {
		fmt.Println("  + Learn hook (captures patterns on stop)")
	}

	return nil
}

// UninstallClaudeCodeHooks removes mur hooks from Claude Code.
func UninstallClaudeCodeHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to uninstall
		}
		return fmt.Errorf("cannot read settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("cannot parse settings: %w", err)
	}

	// Remove hooks
	delete(settings, "hooks")

	// Write back
	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	fmt.Println("✓ Removed Claude Code hooks")
	return nil
}
