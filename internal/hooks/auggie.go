// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AuggieHook defines a single hook command for Auggie (Augment CLI).
type AuggieHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// AuggieHookMatcher defines a hook matcher with hooks array.
type AuggieHookMatcher struct {
	Matcher string       `json:"matcher,omitempty"`
	Hooks   []AuggieHook `json:"hooks"`
}

// InstallAuggieHooks installs mur hooks for Auggie (Augment CLI).
// Auggie supports: SessionStart, SessionEnd, PreToolUse, PostToolUse, Stop
func InstallAuggieHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Check if Auggie is installed (settings directory exists)
	auggieDir := filepath.Join(home, ".augment")
	if _, err := os.Stat(auggieDir); os.IsNotExist(err) {
		return fmt.Errorf("auggie not configured (~/.augment not found)")
	}

	murDir := filepath.Join(home, ".mur")
	promptScriptPath := filepath.Join(murDir, "hooks", "on-prompt.sh")
	stopScriptPath := filepath.Join(murDir, "hooks", "on-stop.sh")

	// Verify hook scripts exist
	if _, err := os.Stat(promptScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hook script not found: %s", promptScriptPath)
	}
	if _, err := os.Stat(stopScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("hook script not found: %s", stopScriptPath)
	}

	settingsPath := filepath.Join(auggieDir, "settings.json")

	hooks := map[string][]AuggieHookMatcher{
		"SessionStart": {
			{
				Hooks: []AuggieHook{
					{Type: "command", Command: fmt.Sprintf("bash %s", promptScriptPath)},
				},
			},
		},
		"Stop": {
			{
				Hooks: []AuggieHook{
					{Type: "command", Command: fmt.Sprintf("bash %s", stopScriptPath)},
				},
			},
		},
	}

	// Load existing settings
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			settings = make(map[string]interface{})
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Backup existing settings
	if _, err := os.Stat(settingsPath); err == nil {
		backupPath := settingsPath + ".backup"
		if data, err := os.ReadFile(settingsPath); err == nil {
			_ = os.WriteFile(backupPath, data, 0644)
		}
	}

	// Merge mur hooks into existing hooks (preserve user-added hooks)
	existingHooks, _ := settings["hooks"].(map[string]interface{})
	if existingHooks == nil {
		existingHooks = make(map[string]interface{})
	}
	for event, matchers := range hooks {
		existingHooks[event] = matchers
	}
	settings["hooks"] = existingHooks

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	return nil
}
