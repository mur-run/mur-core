// Package sync provides configuration synchronization to AI CLI tools.
package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karajanchang/murmur-ai/internal/config"
)

// CLITarget represents an AI CLI tool that can receive synced config.
type CLITarget struct {
	Name       string
	ConfigPath string // relative to home, e.g. ".claude/settings.json"
}

// DefaultTargets returns the list of supported CLI tools.
func DefaultTargets() []CLITarget {
	return []CLITarget{
		{Name: "Claude Code", ConfigPath: ".claude/settings.json"},
		{Name: "Gemini CLI", ConfigPath: ".gemini/settings.json"},
	}
}

// SyncResult holds the result of a sync operation for one target.
type SyncResult struct {
	Target  string
	Success bool
	Message string
}

// SyncMCP syncs MCP server configuration to all CLI tools.
func SyncMCP() ([]SyncResult, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load murmur config: %w", err)
	}

	if len(cfg.MCP.Servers) == 0 {
		return nil, fmt.Errorf("no MCP servers configured in ~/.murmur/config.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	var results []SyncResult
	for _, target := range DefaultTargets() {
		result := syncMCPToTarget(home, target, cfg.MCP.Servers)
		results = append(results, result)
	}

	return results, nil
}

// syncMCPToTarget syncs MCP config to a single CLI target.
func syncMCPToTarget(home string, target CLITarget, servers map[string]interface{}) SyncResult {
	configPath := filepath.Join(home, target.ConfigPath)

	// Read existing config or create empty object
	var settings map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			settings = make(map[string]interface{})
		} else {
			return SyncResult{
				Target:  target.Name,
				Success: false,
				Message: fmt.Sprintf("cannot read config: %v", err),
			}
		}
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return SyncResult{
				Target:  target.Name,
				Success: false,
				Message: fmt.Sprintf("cannot parse config: %v", err),
			}
		}
	}

	// Merge MCP servers
	settings["mcpServers"] = servers

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("cannot create directory: %v", err),
		}
	}

	// Write back
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("cannot serialize config: %v", err),
		}
	}

	if err := os.WriteFile(configPath, output, 0644); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("cannot write config: %v", err),
		}
	}

	return SyncResult{
		Target:  target.Name,
		Success: true,
		Message: fmt.Sprintf("synced %d MCP servers", len(servers)),
	}
}

// SyncAll syncs all configuration (MCP, hooks, etc.) to CLI tools.
func SyncAll() (map[string][]SyncResult, error) {
	results := make(map[string][]SyncResult)

	// Sync MCP
	mcpResults, err := SyncMCP()
	if err != nil {
		return nil, fmt.Errorf("MCP sync failed: %w", err)
	}
	results["mcp"] = mcpResults

	// Future: add hooks sync here

	return results, nil
}
