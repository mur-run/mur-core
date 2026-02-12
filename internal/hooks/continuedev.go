// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ContinueDevInstalled checks if Continue.dev is configured.
func ContinueDevInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check if .continue directory exists
	continueDir := filepath.Join(home, ".continue")
	_, err = os.Stat(continueDir)
	return err == nil
}

// InstallContinueDevHooks installs mur integration for Continue.dev.
func InstallContinueDevHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	murBin, err := findMurBinary()
	if err != nil {
		murBin = "mur"
	}

	continueDir := filepath.Join(home, ".continue")
	configPath := filepath.Join(continueDir, "config.json")

	// Read existing config
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			config = make(map[string]interface{})
		}
	} else {
		config = make(map[string]interface{})
	}

	// Add custom commands for mur
	customCommands := []map[string]interface{}{
		{
			"name":        "mur-patterns",
			"description": "Inject relevant mur patterns into context",
			"prompt":      fmt.Sprintf("{{#if (shell \"%s inject --format md 2>/dev/null\")}}The following patterns may be relevant:\n\n{{shell \"%s inject --format md 2>/dev/null\"}}{{/if}}", murBin, murBin),
		},
		{
			"name":        "mur-search",
			"description": "Search mur patterns",
			"prompt":      fmt.Sprintf("{{#if input}}{{shell \"%s search --format md '{{input}}'\"}}{{/if}}", murBin),
		},
	}

	// Merge with existing custom commands
	if existing, ok := config["customCommands"].([]interface{}); ok {
		// Filter out existing mur commands
		filtered := make([]interface{}, 0)
		for _, cmd := range existing {
			if cmdMap, ok := cmd.(map[string]interface{}); ok {
				if name, ok := cmdMap["name"].(string); ok {
					if name != "mur-patterns" && name != "mur-search" {
						filtered = append(filtered, cmd)
					}
				}
			}
		}
		// Add new mur commands
		for _, cmd := range customCommands {
			filtered = append(filtered, cmd)
		}
		config["customCommands"] = filtered
	} else {
		cmds := make([]interface{}, len(customCommands))
		for i, cmd := range customCommands {
			cmds[i] = cmd
		}
		config["customCommands"] = cmds
	}

	// Add context provider for patterns
	contextProviders := []map[string]interface{}{
		{
			"name": "mur",
			"params": map[string]interface{}{
				"command": fmt.Sprintf("%s inject --format md", murBin),
			},
		},
	}

	if existing, ok := config["contextProviders"].([]interface{}); ok {
		// Filter out existing mur provider
		filtered := make([]interface{}, 0)
		for _, provider := range existing {
			if providerMap, ok := provider.(map[string]interface{}); ok {
				if name, ok := providerMap["name"].(string); ok {
					if name != "mur" {
						filtered = append(filtered, provider)
					}
				}
			}
		}
		for _, provider := range contextProviders {
			filtered = append(filtered, provider)
		}
		config["contextProviders"] = filtered
	} else {
		providers := make([]interface{}, len(contextProviders))
		for i, p := range contextProviders {
			providers[i] = p
		}
		config["contextProviders"] = providers
	}

	// Ensure .continue directory exists
	if err := os.MkdirAll(continueDir, 0755); err != nil {
		return fmt.Errorf("cannot create .continue directory: %w", err)
	}

	// Write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	fmt.Printf("✓ Installed Continue.dev integration at %s\n", configPath)
	fmt.Println("  + Custom command: /mur-patterns")
	fmt.Println("  + Custom command: /mur-search")
	fmt.Println("  + Context provider: @mur")

	return nil
}
