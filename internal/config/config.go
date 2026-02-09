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
	DefaultTool   string              `yaml:"default_tool"`
	Tools         map[string]Tool     `yaml:"tools"`
	Routing       RoutingConfig       `yaml:"routing"`
	Learning      LearningConfig      `yaml:"learning"`
	Sync          SyncConfig          `yaml:"sync"`
	Search        SearchConfig        `yaml:"search"`
	Embeddings    EmbeddingsConfig    `yaml:"embeddings"`
	MCP           MCPConfig           `yaml:"mcp"`
	Hooks         HooksConfig         `yaml:"hooks"`
	Team          TeamConfig          `yaml:"team"`
	Server        ServerConfig        `yaml:"server"`
	Notifications NotificationsConfig `yaml:"notifications"`
}

// ServerConfig represents mur-server cloud sync settings.
type ServerConfig struct {
	URL  string `yaml:"url"`  // Server URL (default: https://api.mur.run)
	Team string `yaml:"team"` // Active team slug
}

// NotificationsConfig represents notification settings.
type NotificationsConfig struct {
	Enabled bool          `yaml:"enabled"`
	Slack   SlackConfig   `yaml:"slack"`
	Discord DiscordConfig `yaml:"discord"`
}

// SlackConfig represents Slack webhook settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// DiscordConfig represents Discord webhook settings.
type DiscordConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

// TeamConfig represents team sharing settings.
type TeamConfig struct {
	Repo     string `yaml:"repo"`      // Git repo URL
	Branch   string `yaml:"branch"`    // Branch name (default: main)
	AutoSync bool   `yaml:"auto_sync"` // Auto sync on pull
}

// RoutingConfig controls automatic tool selection.
type RoutingConfig struct {
	Mode                string  `yaml:"mode"`                 // auto | manual | cost-first | quality-first
	ComplexityThreshold float64 `yaml:"complexity_threshold"` // 0-1, default 0.5
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
	Enabled      bool     `yaml:"enabled"`
	Binary       string   `yaml:"binary"`
	Flags        []string `yaml:"flags"`
	Tier         string   `yaml:"tier"`         // free | paid
	Capabilities []string `yaml:"capabilities"` // coding, analysis, simple-qa, tool-use, architecture
}

// SyncConfig represents sync-related settings.
type SyncConfig struct {
	Format        string `yaml:"format"`          // "directory" or "single"
	PrefixDomain  *bool  `yaml:"prefix_domain"`   // use domain--name format (default: true)
	L3Threshold   int    `yaml:"l3_threshold"`    // chars above which content goes to examples.md
	CleanOld      bool   `yaml:"clean_old"`       // remove old single-file format on sync
}

// SearchConfig represents semantic search settings.
type SearchConfig struct {
	Enabled    *bool  `yaml:"enabled"`     // nil = use default (true)
	Provider   string `yaml:"provider"`    // ollama | openai | none
	Model      string `yaml:"model"`       // embedding model name
	OllamaURL  string `yaml:"ollama_url"`  // Ollama API URL
	TopK       int    `yaml:"top_k"`       // default number of results
	MinScore   float64 `yaml:"min_score"`  // minimum similarity score
	AutoInject *bool  `yaml:"auto_inject"` // auto-inject to prompt via hooks (default: true)
}

// IsEnabled returns whether search is enabled (default: true).
func (s SearchConfig) IsEnabled() bool {
	if s.Enabled == nil {
		return true // default to enabled
	}
	return *s.Enabled
}

// IsAutoInject returns whether auto-inject is enabled (default: true).
func (s SearchConfig) IsAutoInject() bool {
	if s.AutoInject == nil {
		return true // default to enabled
	}
	return *s.AutoInject
}

// EmbeddingsConfig represents embedding cache settings.
type EmbeddingsConfig struct {
	CacheEnabled bool   `yaml:"cache_enabled"`
	CacheDir     string `yaml:"cache_dir"`
	BatchSize    int    `yaml:"batch_size"`
}

// GetPrefixDomain returns whether to use domain prefixes (default: true).
func (s SyncConfig) GetPrefixDomain() bool {
	if s.PrefixDomain == nil {
		return true // default
	}
	return *s.PrefixDomain
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// LearningConfig represents learning-related settings.
type LearningConfig struct {
	AutoExtract  bool `yaml:"auto_extract"`
	SyncToTools  bool `yaml:"sync_to_tools"`
	PatternLimit int  `yaml:"pattern_limit"`
	// Learning repo sync settings
	Repo         string `yaml:"repo"`           // git repo URL for syncing patterns
	Branch       string `yaml:"branch"`         // branch name (default: hostname)
	AutoPush     bool   `yaml:"auto_push"`      // auto push after extract
	PullFromMain bool   `yaml:"pull_from_main"` // also pull shared patterns from main
	// Auto-merge settings
	AutoMerge      bool    `yaml:"auto_merge"`      // enable auto-merge to main
	MergeThreshold float64 `yaml:"merge_threshold"` // confidence threshold for auto-merge (default: 0.8)
	// LLM extraction settings
	LLM LLMConfig `yaml:"llm"`
}

// LLMConfig represents LLM settings for pattern extraction.
type LLMConfig struct {
	Provider  string `yaml:"provider"`    // ollama | claude | openai | gemini
	Model     string `yaml:"model"`       // model name
	OllamaURL string `yaml:"ollama_url"`  // Ollama API URL (default: http://localhost:11434)
	OpenAIURL string `yaml:"openai_url"`  // OpenAI-compatible API URL
	APIKeyEnv string `yaml:"api_key_env"` // Env var name for API key

	// Premium model for important sessions
	Premium *LLMProviderConfig `yaml:"premium,omitempty"`

	// Routing rules for when to use premium
	Routing *LLMRoutingConfig `yaml:"routing,omitempty"`
}

// LLMProviderConfig represents a single LLM provider configuration.
type LLMProviderConfig struct {
	Provider  string `yaml:"provider"`
	Model     string `yaml:"model"`
	OllamaURL string `yaml:"ollama_url,omitempty"`
	OpenAIURL string `yaml:"openai_url,omitempty"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"`
}

// LLMRoutingConfig defines when to use premium model.
type LLMRoutingConfig struct {
	MinMessages int      `yaml:"min_messages"` // Use premium if session has >= N messages
	Projects    []string `yaml:"projects"`     // Use premium for these projects
}

// MCPConfig represents MCP-related settings.
type MCPConfig struct {
	SyncEnabled bool                   `yaml:"sync_enabled"`
	Servers     map[string]interface{} `yaml:"servers"`
}

// ConfigPath returns the path to the config file (~/.mur/config.yaml).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "config.yaml"), nil
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

	// Parse the config file first to see what's specified
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	// Apply defaults for missing sections
	cfg.applyDefaults()

	return &cfg, nil
}

// applyDefaults fills in zero values with sensible defaults.
func (c *Config) applyDefaults() {
	// Sync defaults
	if c.Sync.Format == "" {
		c.Sync.Format = "directory"
	}
	if c.Sync.L3Threshold == 0 {
		c.Sync.L3Threshold = 500
	}

	// Search defaults
	if c.Search.Provider == "" {
		c.Search.Provider = "ollama"
	}
	if c.Search.Model == "" {
		c.Search.Model = "nomic-embed-text"
	}
	if c.Search.OllamaURL == "" {
		c.Search.OllamaURL = "http://localhost:11434"
	}
	if c.Search.TopK == 0 {
		c.Search.TopK = 3
	}
	if c.Search.MinScore == 0 {
		c.Search.MinScore = 0.5
	}

	// Embeddings defaults
	if c.Embeddings.CacheDir == "" {
		c.Embeddings.CacheDir = "~/.mur/embeddings"
	}
	if c.Embeddings.BatchSize == 0 {
		c.Embeddings.BatchSize = 10
	}

	// Default tool
	if c.DefaultTool == "" {
		c.DefaultTool = "claude"
	}
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
	header := "# Murmur Configuration\n# https://github.com/mur-run/mur-core\n\n"
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

// Default returns a default configuration.
func Default() *Config {
	return defaultConfig()
}

// defaultConfig returns a default configuration.
func defaultConfig() *Config {
	return &Config{
		DefaultTool: "claude",
		Tools: map[string]Tool{
			"claude": {
				Enabled:      true,
				Binary:       "claude",
				Flags:        []string{"-p"},
				Tier:         "paid",
				Capabilities: []string{"coding", "analysis", "tool-use", "architecture"},
			},
			"gemini": {
				Enabled:      true,
				Binary:       "gemini",
				Flags:        []string{"-p"},
				Tier:         "free",
				Capabilities: []string{"coding", "simple-qa", "analysis"},
			},
			"auggie": {
				Enabled:      true,
				Binary:       "auggie",
				Flags:        []string{"-p", "-i"},
				Tier:         "free",
				Capabilities: []string{"coding", "simple-qa"},
			},
			"codex": {
				Enabled:      true,
				Binary:       "codex",
				Flags:        []string{},
				Tier:         "paid",
				Capabilities: []string{"coding", "analysis", "tool-use", "architecture"},
			},
			"opencode": {
				Enabled:      true,
				Binary:       "opencode",
				Flags:        []string{},
				Tier:         "free",
				Capabilities: []string{"coding", "simple-qa", "analysis"},
			},
			"aider": {
				Enabled:      true,
				Binary:       "aider",
				Flags:        []string{},
				Tier:         "free", // Can use free models, but also supports paid APIs
				Capabilities: []string{"coding", "analysis"},
			},
			"continue": {
				Enabled:      true,
				Binary:       "continue",
				Flags:        []string{},
				Tier:         "free", // Open source, bring your own API keys
				Capabilities: []string{"coding", "analysis", "tool-use"},
			},
			"cursor": {
				Enabled:      true,
				Binary:       "cursor",
				Flags:        []string{},
				Tier:         "paid", // Subscription model
				Capabilities: []string{"coding", "analysis", "tool-use", "architecture"},
			},
		},
		Routing: RoutingConfig{
			Mode:                "auto",
			ComplexityThreshold: 0.5,
		},
		Learning: LearningConfig{
			AutoExtract:  true,
			SyncToTools:  true,
			PatternLimit: 5,
		},
		Sync: SyncConfig{
			Format:       "directory",
			PrefixDomain: boolPtr(true),
			L3Threshold:  500,
			CleanOld:     false,
		},
		Search: SearchConfig{
			Enabled:    boolPtr(true),
			Provider:   "ollama",
			Model:      "nomic-embed-text",
			OllamaURL:  "http://localhost:11434",
			TopK:       3,
			MinScore:   0.6,
			AutoInject: boolPtr(true),
		},
		Embeddings: EmbeddingsConfig{
			CacheEnabled: true,
			CacheDir:     "~/.mur/embeddings",
			BatchSize:    10,
		},
		MCP: MCPConfig{
			SyncEnabled: true,
			Servers:     make(map[string]interface{}),
		},
	}
}
