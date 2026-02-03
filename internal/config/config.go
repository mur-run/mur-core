// Package config provides configuration management for murmur-ai.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the murmur configuration structure.
type Config struct {
	DefaultTool string          `yaml:"default_tool"`
	Tools       map[string]Tool `yaml:"tools"`
	Learning    LearningConfig  `yaml:"learning"`
	MCP         MCPConfig       `yaml:"mcp"`
	Hooks       HooksConfig     `yaml:"hooks"`
}

// HooksConfig represents hooks configuration for sync to AI CLIs.
type HooksConfig struct {
	UserPromptSubmit []HookGroup `yaml:"UserPromptSubmit"`
	Stop             []HookGroup `yaml:"Stop"`
	BeforeTool       []HookGroup `yaml:"BeforeTool"`
	AfterTool        []HookGroup `yaml:"AfterTool"`
}

// HookGroup represents a group of hooks with a matcher pattern.
type HookGroup struct {
	Matcher string `yaml:"matcher"`
	Hooks   []Hook `yaml:"hooks"`
}

// Hook represents a single hook command.
type Hook struct {
	Type    string `yaml:"type"`
	Command string `yaml:"command"`
}

// Tool represents configuration for an AI tool.
type Tool struct {
	Enabled bool     `yaml:"enabled"`
	Binary  string   `yaml:"binary"`
	Flags   []string `yaml:"flags"`
}

// LearningConfig represents learning-related settings.
type LearningConfig struct {
	AutoExtract  bool `yaml:"auto_extract"`
	SyncToTools  bool `yaml:"sync_to_tools"`
	PatternLimit int  `yaml:"pattern_limit"`
}

// MCPConfig represents MCP-related settings.
type MCPConfig struct {
	SyncEnabled bool                   `yaml:"sync_enabled"`
	Servers     map[string]interface{} `yaml:"servers"`
}

// ConfigPath returns the path to the config file (~/.murmur/config.yaml).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".murmur", "config.yaml"), nil
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes config back to file.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot serialize config: %w", err)
	}

	// Add header comment
	header := "# Murmur Configuration\n# https://github.com/karajanchang/murmur-ai\n\n"
	content := header + string(data)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}

// GetDefaultTool returns the default tool name.
func (c *Config) GetDefaultTool() string {
	if c.DefaultTool == "" {
		return "claude"
	}
	return c.DefaultTool
}

// SetDefaultTool updates the default tool.
func (c *Config) SetDefaultTool(tool string) error {
	// Validate tool exists in config
	if _, ok := c.Tools[tool]; !ok {
		return fmt.Errorf("unknown tool: %s", tool)
	}
	c.DefaultTool = tool
	return nil
}

// GetTool returns tool config by name.
func (c *Config) GetTool(name string) (*Tool, bool) {
	tool, ok := c.Tools[name]
	if !ok {
		return nil, false
	}
	return &tool, true
}

// EnsureTool validates tool exists and is enabled.
func (c *Config) EnsureTool(name string) error {
	tool, ok := c.GetTool(name)
	if !ok {
		return fmt.Errorf("tool not found: %s", name)
	}
	if !tool.Enabled {
		return fmt.Errorf("tool is disabled: %s", name)
	}
	return nil
}

// defaultConfig returns a default configuration.
func defaultConfig() *Config {
	return &Config{
		DefaultTool: "claude",
		Tools: map[string]Tool{
			"claude": {Enabled: true, Binary: "claude", Flags: []string{"-p"}},
			"gemini": {Enabled: true, Binary: "gemini", Flags: []string{"-p"}},
			"auggie": {Enabled: false, Binary: "auggie", Flags: []string{}},
		},
		Learning: LearningConfig{
			AutoExtract:  true,
			SyncToTools:  true,
			PatternLimit: 5,
		},
		MCP: MCPConfig{
			SyncEnabled: true,
			Servers:     make(map[string]interface{}),
		},
	}
}
