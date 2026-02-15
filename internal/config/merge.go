package config

// MergeConfig merges defaults into existing config.
// Existing values take precedence; only missing fields are added from defaults.
func MergeConfig(existing, defaults *Config) *Config {
	result := *existing

	// NOTE: Don't update SchemaVersion here - let MigrateConfig handle it
	// This preserves the original version for migration detection

	// Merge DefaultTool (if not set)
	if result.DefaultTool == "" {
		result.DefaultTool = defaults.DefaultTool
	}

	// Merge Tools (add missing tools, don't overwrite existing)
	if result.Tools == nil {
		result.Tools = defaults.Tools
	} else {
		for name, tool := range defaults.Tools {
			if _, exists := result.Tools[name]; !exists {
				result.Tools[name] = tool
			}
		}
	}

	// Merge Routing (if zero value)
	if result.Routing.Mode == "" {
		result.Routing.Mode = defaults.Routing.Mode
	}
	if result.Routing.ComplexityThreshold == 0 {
		result.Routing.ComplexityThreshold = defaults.Routing.ComplexityThreshold
	}

	// Merge Learning (preserve existing, fill missing)
	if result.Learning.PatternLimit == 0 {
		result.Learning.PatternLimit = defaults.Learning.PatternLimit
	}

	// Merge Sync (preserve existing, fill missing)
	if result.Sync.Format == "" {
		result.Sync.Format = defaults.Sync.Format
	}
	if result.Sync.L3Threshold == 0 {
		result.Sync.L3Threshold = defaults.Sync.L3Threshold
	}

	// Merge Search (preserve existing, fill missing)
	if result.Search.Provider == "" {
		result.Search.Provider = defaults.Search.Provider
	}
	if result.Search.Model == "" {
		result.Search.Model = defaults.Search.Model
	}
	if result.Search.OllamaURL == "" {
		result.Search.OllamaURL = defaults.Search.OllamaURL
	}
	if result.Search.TopK == 0 {
		result.Search.TopK = defaults.Search.TopK
	}
	if result.Search.MinScore == 0 {
		result.Search.MinScore = defaults.Search.MinScore
	}

	// Merge Embeddings (preserve existing, fill missing)
	if result.Embeddings.CacheDir == "" {
		result.Embeddings.CacheDir = defaults.Embeddings.CacheDir
	}
	if result.Embeddings.BatchSize == 0 {
		result.Embeddings.BatchSize = defaults.Embeddings.BatchSize
	}

	// Merge MCP (preserve existing)
	if result.MCP.Servers == nil {
		result.MCP.Servers = defaults.MCP.Servers
	}

	// Merge Community (if zero value - new in v2)
	if result.Community == (CommunityConfig{}) {
		result.Community = defaults.Community
	}

	// Merge Notifications (preserve existing, fill defaults if all zero)
	if !result.Notifications.Enabled && !result.Notifications.System &&
		!result.Notifications.OnError && !result.Notifications.OnPatterns {
		result.Notifications = defaults.Notifications
	}

	return &result
}
