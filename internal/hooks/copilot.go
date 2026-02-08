// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CopilotHookDef defines a single hook for GitHub Copilot.
type CopilotHookDef struct {
	Type    string `json:"type"`
	Bash    string `json:"bash,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// CopilotHooksConfig defines the hooks configuration for GitHub Copilot.
type CopilotHooksConfig struct {
	Version int                        `json:"version"`
	Hooks   map[string][]CopilotHookDef `json:"hooks"`
}

// InstallCopilotHooks installs mur hooks for GitHub Copilot.
func InstallCopilotHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Find mur binary
	murPath, err := findMurBinary()
	if err != nil {
		return err
	}

	// Create hooks directory (global location)
	hooksDir := filepath.Join(home, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	// Create hooks configuration
	config := CopilotHooksConfig{
		Version: 1,
		Hooks: map[string][]CopilotHookDef{
			"sessionStart": {{
				Type:    "command",
				Bash:    fmt.Sprintf("echo '# Learned Patterns\\nApply these patterns when relevant:\\n'; %s context 2>/dev/null || true", murPath),
				Timeout: 5,
			}},
			"sessionEnd": {{
				Type:    "command",
				Bash:    fmt.Sprintf("%s learn extract --llm --auto --accept-all --quiet 2>/dev/null || true", murPath),
				Timeout: 300,
			}},
		},
	}

	// Write mur.json
	hooksPath := filepath.Join(hooksDir, "mur.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal hooks config: %w", err)
	}

	if err := os.WriteFile(hooksPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write hooks file: %w", err)
	}

	return nil
}

// CheckCopilotHooks checks if mur hooks are installed for GitHub Copilot.
func CheckCopilotHooks() (installed bool, path string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	hooksPath := filepath.Join(home, ".github", "hooks", "mur.json")
	if _, err := os.Stat(hooksPath); err == nil {
		return true, hooksPath
	}

	return false, hooksPath
}

// RemoveCopilotHooks removes mur hooks for GitHub Copilot.
func RemoveCopilotHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	hooksPath := filepath.Join(home, ".github", "hooks", "mur.json")
	return os.Remove(hooksPath)
}
