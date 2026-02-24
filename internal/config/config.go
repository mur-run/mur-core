// Package config provides configuration management for murmur-ai.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CurrentSchemaVersion is the latest config schema version.
// Increment this when adding new config fields that need migration.
const CurrentSchemaVersion = 2

// Config represents the murmur configuration structure.
type Config struct {
	SchemaVersion int                 `yaml:"schema_version" json:"schema_version"`
	DefaultTool   string              `yaml:"default_tool"`
	Tools         map[string]Tool     `yaml:"tools"`
	Routing       RoutingConfig       `yaml:"routing,omitempty"`
	Learning      LearningConfig      `yaml:"learning,omitempty"`
	Sync          SyncConfig          `yaml:"sync,omitempty"`
	Search        SearchConfig        `yaml:"search,omitempty"`
	Embeddings    EmbeddingsConfig    `yaml:"embeddings,omitempty"`
	MCP           MCPConfig           `yaml:"mcp,omitempty"`
	Hooks         HooksConfig         `yaml:"hooks,omitempty"`
	Team          TeamConfig          `yaml:"team,omitempty"`
	Server        ServerConfig        `yaml:"server,omitempty"`
	Notifications NotificationsConfig `yaml:"notifications,omitempty"`
	TechStack     []string            `yaml:"tech_stack,omitempty"`    // User's tech stack for filtering (e.g., ["swift", "go", "docker"])
	Cache         CacheConfig         `yaml:"cache,omitempty"`         // Local cache settings
	Community     CommunityConfig     `yaml:"community,omitempty"`     // Community sharing settings
	Privacy       PrivacyConfig       `yaml:"privacy,omitempty"`       // Privacy & PII protection settings
	Consolidation ConsolidationConfig `yaml:"consolidation,omitempty"` // Pattern consolidation settings
}

// CacheConfig represents local cache settings for community patterns.
type CacheConfig struct {
	Community CommunityCacheConfig `yaml:"community,omitempty"`
}

// CommunityConfig represents community sharing settings.
type CommunityConfig struct {
	ShareEnabled    bool `yaml:"share_enabled,omitempty"`      // Enable community sharing
	AutoShareOnPush bool `yaml:"auto_share_on_push,omitempty"` // Auto-share when pushing
	ShareExtracted  bool `yaml:"share_extracted,omitempty"`    // Share extracted patterns (may contain secrets)
}

// DefaultCommunityConfig returns default community settings.
func DefaultCommunityConfig() CommunityConfig {
	return CommunityConfig{
		ShareEnabled:    false, // Will be asked during init, default N until confirmed
		AutoShareOnPush: true,  // If sharing enabled, auto-share on push
		ShareExtracted:  false, // Extracted patterns may contain secrets
	}
}

// PrivacyConfig represents privacy and PII protection settings.
type PrivacyConfig struct {
	RedactTerms           []string                    `yaml:"redact_terms,omitempty"`           // Terms to always redact
	Replacements          map[string]string           `yaml:"replacements,omitempty"`           // Custom replacement mappings
	AutoDetect            AutoDetectConfig            `yaml:"auto_detect,omitempty"`            // Auto-detection toggles
	SemanticAnonymization SemanticAnonymizationConfig `yaml:"semantic_anonymization,omitempty"` // LLM-based anonymization
}

// SemanticAnonymizationConfig controls LLM-based semantic anonymization.
type SemanticAnonymizationConfig struct {
	Enabled      bool   `yaml:"enabled,omitempty"`       // Opt-in (default: false)
	Provider     string `yaml:"provider,omitempty"`      // ollama | openai | anthropic
	Model        string `yaml:"model,omitempty"`         // Model for anonymization
	OllamaURL    string `yaml:"ollama_url,omitempty"`    // Ollama API URL
	CacheResults bool   `yaml:"cache_results,omitempty"` // Cache anonymization results (default: true)
}

// AutoDetectConfig controls which PII types are auto-detected.
type AutoDetectConfig struct {
	Emails       *bool `yaml:"emails,omitempty"`        // Detect email addresses (default: true)
	InternalIPs  *bool `yaml:"internal_ips,omitempty"`  // Detect internal IPs (default: true)
	FilePaths    *bool `yaml:"file_paths,omitempty"`    // Detect user file paths (default: true)
	PhoneNumbers *bool `yaml:"phone_numbers,omitempty"` // Detect phone numbers (default: true)
	InternalURLs *bool `yaml:"internal_urls,omitempty"` // Detect internal URLs (default: true)
}

// IsEmailsEnabled returns whether email detection is enabled (default: true).
func (a AutoDetectConfig) IsEmailsEnabled() bool {
	if a.Emails == nil {
		return true
	}
	return *a.Emails
}

// IsInternalIPsEnabled returns whether internal IP detection is enabled (default: true).
func (a AutoDetectConfig) IsInternalIPsEnabled() bool {
	if a.InternalIPs == nil {
		return true
	}
	return *a.InternalIPs
}

// IsFilePathsEnabled returns whether file path detection is enabled (default: true).
func (a AutoDetectConfig) IsFilePathsEnabled() bool {
	if a.FilePaths == nil {
		return true
	}
	return *a.FilePaths
}

// IsPhoneNumbersEnabled returns whether phone number detection is enabled (default: true).
func (a AutoDetectConfig) IsPhoneNumbersEnabled() bool {
	if a.PhoneNumbers == nil {
		return true
	}
	return *a.PhoneNumbers
}

// IsInternalURLsEnabled returns whether internal URL detection is enabled (default: true).
func (a AutoDetectConfig) IsInternalURLsEnabled() bool {
	if a.InternalURLs == nil {
		return true
	}
	return *a.InternalURLs
}

// DefaultPrivacyConfig returns default privacy settings.
func DefaultPrivacyConfig() PrivacyConfig {
	return PrivacyConfig{
		AutoDetect: AutoDetectConfig{
			Emails:       boolPtr(true),
			InternalIPs:  boolPtr(true),
			FilePaths:    boolPtr(true),
			PhoneNumbers: boolPtr(true),
			InternalURLs: boolPtr(true),
		},
		SemanticAnonymization: SemanticAnonymizationConfig{
			Enabled:      false,
			Provider:     "ollama",
			Model:        "llama3.2",
			OllamaURL:    "http://localhost:11434",
			CacheResults: true,
		},
	}
}

// ConsolidationConfig represents pattern consolidation settings.
type ConsolidationConfig struct {
	Enabled              bool    `yaml:"enabled,omitempty"`
	Schedule             string  `yaml:"schedule,omitempty"` // daily | weekly | monthly
	AutoArchive          bool    `yaml:"auto_archive,omitempty"`
	AutoMerge            string  `yaml:"auto_merge,omitempty"`      // off | keep-best | llm-merge
	MergeThreshold       float64 `yaml:"merge_threshold,omitempty"` // cosine similarity threshold
	DecayHalfLifeDays    int     `yaml:"decay_half_life_days,omitempty"`
	GracePeriodDays      int     `yaml:"grace_period_days,omitempty"`
	MinPatternsBeforeRun int     `yaml:"min_patterns_before_run,omitempty"`
	NotifyOnRun          bool    `yaml:"notify_on_run,omitempty"`
}

// DefaultConsolidationConfig returns default consolidation settings.
func DefaultConsolidationConfig() ConsolidationConfig {
	return ConsolidationConfig{
		Enabled:              true,
		Schedule:             "weekly",
		AutoArchive:          true,
		AutoMerge:            "keep-best",
		MergeThreshold:       0.85,
		DecayHalfLifeDays:    90,
		GracePeriodDays:      14,
		MinPatternsBeforeRun: 50,
		NotifyOnRun:          true,
	}
}

// CommunityCacheConfig represents community pattern cache settings.
type CommunityCacheConfig struct {
	Enabled   bool   `yaml:"enabled,omitempty"`     // Enable caching (default: true)
	TTLDays   int    `yaml:"ttl_days,omitempty"`    // Days to keep cached patterns (default: 7)
	MaxSizeMB int    `yaml:"max_size_mb,omitempty"` // Max cache size in MB (default: 50)
	Cleanup   string `yaml:"cleanup,omitempty"`     // When to cleanup: on_sync | daily | manual (default: on_sync)
}

// GetTechStack returns the configured tech stack.
func (c *Config) GetTechStack() []string {
	return c.TechStack
}

// GetCommunityConfig returns the community config.
func (c *Config) GetCommunityConfig() CommunityConfig {
	return c.Community
}

// GetCacheConfig returns the cache config with defaults.
func (c *Config) GetCacheConfig() CommunityCacheConfig {
	cfg := c.Cache.Community

	// Default enabled to true if not explicitly set
	// We use a pointer trick: if Enabled is false and no config exists, enable by default
	if !cfg.Enabled && c.Cache.Community.TTLDays == 0 && c.Cache.Community.MaxSizeMB == 0 {
		cfg.Enabled = true // Default to enabled
	}

	if cfg.TTLDays == 0 {
		cfg.TTLDays = 7
	}
	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = 50
	}
	if cfg.Cleanup == "" {
		cfg.Cleanup = "on_sync"
	}
	return cfg
}

// ServerConfig represents mur-server cloud sync settings.
type ServerConfig struct {
	URL  string `yaml:"url,omitempty"`  // Server URL (default: https://api.mur.run)
	Team string `yaml:"team,omitempty"` // Active team slug
}

// NotificationsConfig represents notification settings.
type NotificationsConfig struct {
	Enabled    bool          `yaml:"enabled,omitempty"`
	System     bool          `yaml:"system,omitempty"`      // Enable macOS system notifications
	OnError    bool          `yaml:"on_error,omitempty"`    // Notify on errors
	OnPatterns bool          `yaml:"on_patterns,omitempty"` // Notify when patterns are extracted
	Slack      SlackConfig   `yaml:"slack,omitempty"`
	Discord    DiscordConfig `yaml:"discord,omitempty"`
}

// SlackConfig represents Slack webhook settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
}

// DiscordConfig represents Discord webhook settings.
type DiscordConfig struct {
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// TeamConfig represents team sharing settings.
type TeamConfig struct {
	Repo     string `yaml:"repo,omitempty"`      // Git repo URL
	Branch   string `yaml:"branch,omitempty"`    // Branch name (default: main)
	AutoSync bool   `yaml:"auto_sync,omitempty"` // Auto sync on pull
}

// RoutingConfig controls automatic tool selection.
type RoutingConfig struct {
	Mode                string  `yaml:"mode,omitempty"`                 // auto | manual | cost-first | quality-first
	ComplexityThreshold float64 `yaml:"complexity_threshold,omitempty"` // 0-1, default 0.5
}

// HooksConfig represents hooks configuration for sync to AI CLIs.
type HooksConfig struct {
	UserPromptSubmit []HookGroup `yaml:"UserPromptSubmit,omitempty"`
	Stop             []HookGroup `yaml:"Stop,omitempty"`
	BeforeTool       []HookGroup `yaml:"BeforeTool,omitempty"`
	AfterTool        []HookGroup `yaml:"AfterTool,omitempty"`
}

// HookGroup represents a group of hooks with a matcher pattern.
type HookGroup struct {
	Matcher string `yaml:"matcher,omitempty"`
	Hooks   []Hook `yaml:"hooks,omitempty"`
}

// Hook represents a single hook command.
type Hook struct {
	Type    string `yaml:"type,omitempty"`
	Command string `yaml:"command,omitempty"`
}

// Tool represents configuration for an AI tool.
type Tool struct {
	Enabled      bool     `yaml:"enabled"`
	Binary       string   `yaml:"binary,omitempty"`
	Flags        []string `yaml:"flags,omitempty"`
	Tier         string   `yaml:"tier,omitempty"`         // free | paid
	Capabilities []string `yaml:"capabilities,omitempty"` // coding, analysis, simple-qa, tool-use, architecture
}

// SyncConfig represents sync-related settings.
type SyncConfig struct {
	Format          string `yaml:"format,omitempty"`           // "directory" or "single"
	PrefixDomain    *bool  `yaml:"prefix_domain,omitempty"`    // use domain--name format (default: true)
	L3Threshold     int    `yaml:"l3_threshold,omitempty"`     // chars above which content goes to examples.md
	CleanOld        bool   `yaml:"clean_old,omitempty"`        // remove old single-file format on sync
	Auto            bool   `yaml:"auto,omitempty"`             // enable automatic sync
	IntervalMinutes int    `yaml:"interval_minutes,omitempty"` // sync interval in minutes (default: 30)
}

// SearchConfig represents semantic search settings.
type SearchConfig struct {
	Enabled    *bool   `yaml:"enabled,omitempty"`     // nil = use default (true)
	Provider   string  `yaml:"provider,omitempty"`    // ollama | openai | google | voyage | none
	Model      string  `yaml:"model,omitempty"`       // embedding model name
	OllamaURL  string  `yaml:"ollama_url,omitempty"`  // Ollama API URL
	APIKeyEnv  string  `yaml:"api_key_env,omitempty"` // env var name for API key (e.g. OPENAI_API_KEY)
	TopK       int     `yaml:"top_k,omitempty"`       // default number of results
	MinScore   float64 `yaml:"min_score,omitempty"`   // minimum similarity score
	AutoInject *bool   `yaml:"auto_inject,omitempty"` // auto-inject to prompt via hooks (default: true)
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
	CacheEnabled bool   `yaml:"cache_enabled,omitempty"`
	CacheDir     string `yaml:"cache_dir,omitempty"`
	BatchSize    int    `yaml:"batch_size,omitempty"`
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
	AutoExtract  bool `yaml:"auto_extract,omitempty"`
	SyncToTools  bool `yaml:"sync_to_tools,omitempty"`
	PatternLimit int  `yaml:"pattern_limit,omitempty"`
	// Learning repo sync settings
	Repo         string `yaml:"repo,omitempty"`           // git repo URL for syncing patterns
	Branch       string `yaml:"branch,omitempty"`         // branch name (default: hostname)
	AutoPush     bool   `yaml:"auto_push,omitempty"`      // auto push after extract
	PullFromMain bool   `yaml:"pull_from_main,omitempty"` // also pull shared patterns from main
	// Auto-merge settings
	AutoMerge      bool    `yaml:"auto_merge,omitempty"`      // enable auto-merge to main
	MergeThreshold float64 `yaml:"merge_threshold,omitempty"` // confidence threshold for auto-merge (default: 0.8)
	// LLM extraction settings
	LLM LLMConfig `yaml:"llm,omitempty"`
}

// LLMConfig represents LLM settings for pattern extraction.
type LLMConfig struct {
	Provider  string `yaml:"provider,omitempty"`    // ollama | claude | openai | gemini
	Model     string `yaml:"model,omitempty"`       // model name
	OllamaURL string `yaml:"ollama_url,omitempty"`  // Ollama API URL (default: http://localhost:11434)
	OpenAIURL string `yaml:"openai_url,omitempty"`  // OpenAI-compatible API URL
	APIKeyEnv string `yaml:"api_key_env,omitempty"` // Env var name for API key

	// Premium model for important sessions
	Premium *LLMProviderConfig `yaml:"premium,omitempty"`

	// Routing rules for when to use premium
	Routing *LLMRoutingConfig `yaml:"routing,omitempty"`
}

// IsZero reports whether the LLM config is empty (enables yaml omitempty on structs).
func (l LLMConfig) IsZero() bool {
	return l.Provider == "" && l.Model == "" && l.Premium == nil && l.Routing == nil
}

// LLMProviderConfig represents a single LLM provider configuration.
type LLMProviderConfig struct {
	Provider  string `yaml:"provider,omitempty"`
	Model     string `yaml:"model,omitempty"`
	OllamaURL string `yaml:"ollama_url,omitempty"`
	OpenAIURL string `yaml:"openai_url,omitempty"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"`
}

// LLMRoutingConfig defines when to use premium model.
type LLMRoutingConfig struct {
	MinMessages int      `yaml:"min_messages,omitempty"` // Use premium if session has >= N messages
	Projects    []string `yaml:"projects,omitempty"`     // Use premium for these projects
}

// MCPConfig represents MCP-related settings.
type MCPConfig struct {
	SyncEnabled bool                   `yaml:"sync_enabled,omitempty"`
	Servers     map[string]interface{} `yaml:"servers,omitempty"`
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
		c.Search.Model = "qwen3-embedding"
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

	// Consolidation defaults
	if c.Consolidation.Schedule == "" {
		c.Consolidation.Schedule = "weekly"
	}
	if c.Consolidation.AutoMerge == "" {
		c.Consolidation.AutoMerge = "keep-best"
	}
	if c.Consolidation.MergeThreshold == 0 {
		c.Consolidation.MergeThreshold = 0.85
	}
	if c.Consolidation.DecayHalfLifeDays == 0 {
		c.Consolidation.DecayHalfLifeDays = 90
	}
	if c.Consolidation.GracePeriodDays == 0 {
		c.Consolidation.GracePeriodDays = 14
	}
	if c.Consolidation.MinPatternsBeforeRun == 0 {
		c.Consolidation.MinPatternsBeforeRun = 50
	}

	// Default tool
	if c.DefaultTool == "" {
		c.DefaultTool = "claude"
	}
}

// Save writes config back to file, preserving any existing comments.
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

	// Marshal current config into a yaml.Node tree
	var freshDoc yaml.Node
	freshBytes, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot serialize config: %w", err)
	}
	if err := yaml.Unmarshal(freshBytes, &freshDoc); err != nil {
		return fmt.Errorf("cannot parse serialized config: %w", err)
	}

	// Try to read existing file to preserve comments
	existing, readErr := os.ReadFile(path)
	if readErr == nil && len(existing) > 0 {
		var existingDoc yaml.Node
		if err := yaml.Unmarshal(existing, &existingDoc); err == nil {
			// Merge fresh values into existing tree (preserving comments)
			mergeNodes(&existingDoc, &freshDoc)
			return writeNodeToFile(path, &existingDoc)
		}
	}

	// No existing file or unparseable — write fresh with header comment
	header := "# Murmur Configuration\n# https://github.com/mur-run/mur-core\n\n"
	content := header + string(freshBytes)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}

// writeNodeToFile encodes a yaml.Node document to a file.
func writeNodeToFile(path string, doc *yaml.Node) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("cannot encode config: %w", err)
	}
	return enc.Close()
}

// mergeNodes recursively merges values from src into dst, preserving
// dst's comments. Both must be document or mapping nodes.
func mergeNodes(dst, src *yaml.Node) {
	// Unwrap document nodes
	if dst.Kind == yaml.DocumentNode && src.Kind == yaml.DocumentNode {
		if len(dst.Content) > 0 && len(src.Content) > 0 {
			mergeNodes(dst.Content[0], src.Content[0])
		}
		return
	}

	// Only merge mappings
	if dst.Kind != yaml.MappingNode || src.Kind != yaml.MappingNode {
		return
	}

	srcKeys := buildKeyIndex(src)

	// Update existing keys in dst
	for i := 0; i+1 < len(dst.Content); i += 2 {
		keyNode := dst.Content[i]
		valNode := dst.Content[i+1]

		srcIdx, ok := srcKeys[keyNode.Value]
		if !ok {
			continue
		}
		srcVal := src.Content[srcIdx+1]

		if valNode.Kind == yaml.MappingNode && srcVal.Kind == yaml.MappingNode {
			// Recurse into nested mappings
			mergeNodes(valNode, srcVal)
		} else {
			// Replace value but preserve comments on the value node
			headComment := valNode.HeadComment
			lineComment := valNode.LineComment
			footComment := valNode.FootComment

			*valNode = *srcVal

			// Restore original comments if the new node has none
			if valNode.HeadComment == "" {
				valNode.HeadComment = headComment
			}
			if valNode.LineComment == "" {
				valNode.LineComment = lineComment
			}
			if valNode.FootComment == "" {
				valNode.FootComment = footComment
			}
		}
	}

	// Add keys present in src but missing from dst
	dstKeys := buildKeyIndex(dst)
	for i := 0; i+1 < len(src.Content); i += 2 {
		srcKey := src.Content[i]
		if _, exists := dstKeys[srcKey.Value]; !exists {
			dst.Content = append(dst.Content, src.Content[i], src.Content[i+1])
		}
	}
}

// buildKeyIndex returns a map from key string to its index in a mapping node's Content slice.
func buildKeyIndex(m *yaml.Node) map[string]int {
	idx := make(map[string]int, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		idx[m.Content[i].Value] = i
	}
	return idx
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
		SchemaVersion: CurrentSchemaVersion,
		DefaultTool:   "claude",
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
			AutoExtract: true,
			SyncToTools: true,
			LLM: LLMConfig{
				Provider: "ollama",
				Model:    "llama3.2:3b",
			},
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
			Model:      "qwen3-embedding",
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
		Notifications: NotificationsConfig{
			Enabled:    true,
			System:     true,
			OnError:    true,
			OnPatterns: true,
		},
		Community:     DefaultCommunityConfig(),
		Privacy:       DefaultPrivacyConfig(),
		Consolidation: DefaultConsolidationConfig(),
	}
}
