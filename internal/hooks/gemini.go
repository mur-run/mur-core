// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GeminiHook represents a hook entry for Gemini CLI.
type GeminiHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// GeminiSettings represents the Gemini CLI settings file.
type GeminiSettings struct {
	Hooks map[string][]GeminiHook `json:"hooks,omitempty"`
}

// GeminiCLIInstalled checks if Gemini CLI is installed.
func GeminiCLIInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check if .gemini directory exists
	geminiDir := filepath.Join(home, ".gemini")
	_, err = os.Stat(geminiDir)
	return err == nil
}

// InstallGeminiHooks installs mur hooks for Gemini CLI.
func InstallGeminiHooks(enableSearch bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	murBin, err := findMurBinary()
	if err != nil {
		murBin = "mur"
	}

	settingsPath := filepath.Join(home, ".gemini", "settings.json")

	// Read existing settings
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			settings = make(map[string]interface{})
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Build hooks
	hooks := make(map[string][]GeminiHook)

	// Learn hook (on exit)
	hooks["exit"] = []GeminiHook{
		{
			Type:    "command",
			Command: fmt.Sprintf("%s learn extract --auto --quiet 2>/dev/null || true", murBin),
		},
	}

	// Search hook (on prompt)
	if enableSearch {
		hooks["prompt"] = []GeminiHook{
			{
				Type:    "command",
				Command: fmt.Sprintf(`%s search --inject "$PROMPT" 2>/dev/null || true`, murBin),
			},
		}
	}

	settings["hooks"] = hooks

	// Ensure .gemini directory exists
	if err := os.MkdirAll(filepath.Join(home, ".gemini"), 0755); err != nil {
		return fmt.Errorf("cannot create .gemini directory: %w", err)
	}

	// Write settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	fmt.Printf("✓ Installed Gemini CLI hooks at %s\n", settingsPath)
	if enableSearch {
		fmt.Println("  + Learn hook (captures patterns on exit)")
		fmt.Println("  + Search hook (suggests patterns on prompt)")
	} else {
		fmt.Println("  + Learn hook (captures patterns on exit)")
	}

	return nil
}
