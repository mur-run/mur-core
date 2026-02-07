// Package sync provides configuration synchronization to AI CLI tools.
package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/config"
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
		{Name: "Auggie", ConfigPath: ".augment/settings.json"},
		{Name: "OpenCode", ConfigPath: ".opencode/settings.json"},
		{Name: "Continue", ConfigPath: ".continue/config.json"},
	}
}

// CodexInstructionsPath returns the path to Codex instructions file.
func CodexInstructionsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".codex", "instructions.md"), nil
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

// SyncAll syncs all configuration (MCP, hooks, skills, etc.) to CLI tools.
func SyncAll() (map[string][]SyncResult, error) {
	results := make(map[string][]SyncResult)

	// Sync MCP
	mcpResults, err := SyncMCP()
	if err != nil {
		return nil, fmt.Errorf("MCP sync failed: %w", err)
	}
	results["mcp"] = mcpResults

	// Sync Hooks
	hooksResults, err := SyncHooks()
	if err != nil {
		return nil, fmt.Errorf("hooks sync failed: %w", err)
	}
	results["hooks"] = hooksResults

	// Sync Skills (don't fail if no skills exist)
	skillsResults, err := SyncSkills()
	if err == nil {
		results["skills"] = skillsResults
	}
	// Note: We don't return error for skills - they're optional

	return results, nil
}

// eventMapping defines how murmur events map to CLI-specific events.
var eventMapping = map[string]map[string]string{
	"Claude Code": {
		"UserPromptSubmit": "UserPromptSubmit",
		"Stop":             "Stop",
		"BeforeTool":       "PreToolUse",
		"AfterTool":        "PostToolUse",
	},
	"Gemini CLI": {
		"UserPromptSubmit": "BeforeAgent",
		"Stop":             "AfterAgent",
		"BeforeTool":       "BeforeTool",
		"AfterTool":        "AfterTool",
	},
	// Auggie doesn't support hooks natively, but we include it for MCP sync.
	// Empty mapping means no hooks will be synced.
	"Auggie": {},
	// OpenCode doesn't support hooks natively, but we include it for MCP sync.
	"OpenCode": {},
	// Codex uses ~/.codex/instructions.md instead of settings.json.
	// No hooks support, patterns synced separately via SyncPatternsToCodex.
	// Continue supports MCP but uses different format.
	"Continue": {},
}

// SyncHooks syncs hooks configuration to all CLI tools.
func SyncHooks() ([]SyncResult, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load murmur config: %w", err)
	}

	// Check if any hooks are configured
	if !hasHooks(cfg.Hooks) {
		return nil, fmt.Errorf("no hooks configured in ~/.murmur/config.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	var results []SyncResult
	for _, target := range DefaultTargets() {
		result := syncHooksToTarget(home, target, cfg.Hooks)
		results = append(results, result)
	}

	return results, nil
}

// hasHooks checks if any hooks are configured.
func hasHooks(h config.HooksConfig) bool {
	return len(h.UserPromptSubmit) > 0 || len(h.Stop) > 0 ||
		len(h.BeforeTool) > 0 || len(h.AfterTool) > 0
}

// syncHooksToTarget syncs hooks config to a single CLI target.
func syncHooksToTarget(home string, target CLITarget, hooks config.HooksConfig) SyncResult {
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

	// Get event mapping for this CLI
	mapping, ok := eventMapping[target.Name]
	if !ok {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: "no event mapping defined for this target",
		}
	}

	// Build hooks object for this CLI
	cliHooks := make(map[string]interface{})
	hookCount := 0

	// Map each murmur event to the CLI-specific event
	murmurEvents := map[string][]config.HookGroup{
		"UserPromptSubmit": hooks.UserPromptSubmit,
		"Stop":             hooks.Stop,
		"BeforeTool":       hooks.BeforeTool,
		"AfterTool":        hooks.AfterTool,
	}

	for murmurEvent, hookGroups := range murmurEvents {
		if len(hookGroups) == 0 {
			continue
		}

		cliEvent, ok := mapping[murmurEvent]
		if !ok {
			continue
		}

		// Convert hook groups to JSON-compatible format
		var jsonGroups []map[string]interface{}
		for _, group := range hookGroups {
			jsonHooks := make([]map[string]interface{}, len(group.Hooks))
			for i, h := range group.Hooks {
				jsonHooks[i] = map[string]interface{}{
					"type":    h.Type,
					"command": h.Command,
				}
				hookCount++
			}
			jsonGroups = append(jsonGroups, map[string]interface{}{
				"matcher": group.Matcher,
				"hooks":   jsonHooks,
			})
		}

		cliHooks[cliEvent] = jsonGroups
	}

	// Merge hooks into settings
	settings["hooks"] = cliHooks

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
		Message: fmt.Sprintf("synced %d hooks", hookCount),
	}
}
