// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// OpenCodeHook defines a single hook command for OpenCode.
type OpenCodeHook struct {
	Run    string `yaml:"run"`
	Inject string `yaml:"inject,omitempty"`
}

// OpenCodeHooks defines before/after hooks for OpenCode.
type OpenCodeHooks struct {
	Before []OpenCodeHook `yaml:"before,omitempty"`
	After  []OpenCodeHook `yaml:"after,omitempty"`
}

// OpenCodePlugin defines the plugin configuration for OpenCode.
type OpenCodePlugin struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Hooks       OpenCodeHooks `yaml:"hooks"`
}

// InstallOpenCodeHooks installs mur hooks for OpenCode.
func InstallOpenCodeHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Find mur binary
	murPath, err := findMurBinary()
	if err != nil {
		return err
	}

	// Create plugin directory
	pluginDir := filepath.Join(home, ".config", "opencode", "plugins", "mur")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("cannot create plugin directory: %w", err)
	}

	// Create plugin configuration
	plugin := OpenCodePlugin{
		Name:        "mur",
		Description: "Continuous learning for AI assistants - injects patterns and extracts learnings",
		Hooks: OpenCodeHooks{
			Before: []OpenCodeHook{{
				Run:    fmt.Sprintf("%s context 2>/dev/null", murPath),
				Inject: "# Learned Patterns\nApply these patterns when relevant:\n\n{stdout}",
			}},
			After: []OpenCodeHook{{
				Run: fmt.Sprintf("%s learn extract --llm --auto --accept-all --quiet 2>/dev/null || true", murPath),
			}},
		},
	}

	// Write plugin.yaml
	pluginPath := filepath.Join(pluginDir, "plugin.yaml")
	data, err := yaml.Marshal(plugin)
	if err != nil {
		return fmt.Errorf("cannot marshal plugin config: %w", err)
	}

	if err := os.WriteFile(pluginPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write plugin file: %w", err)
	}

	return nil
}

// CheckOpenCodeHooks checks if mur hooks are installed for OpenCode.
func CheckOpenCodeHooks() (installed bool, path string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "mur", "plugin.yaml")
	if _, err := os.Stat(pluginPath); err == nil {
		return true, pluginPath
	}

	return false, pluginPath
}

// RemoveOpenCodeHooks removes mur hooks for OpenCode.
func RemoveOpenCodeHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(home, ".config", "opencode", "plugins", "mur")
	return os.RemoveAll(pluginDir)
}
