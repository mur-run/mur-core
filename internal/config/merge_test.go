package config

import "testing"

func TestMergeConfig_PreservesExisting(t *testing.T) {
	existing := &Config{
		SchemaVersion: 1,
		DefaultTool:   "gemini",
		Tools: map[string]Tool{
			"gemini": {Enabled: true, Binary: "gemini"},
		},
		Routing: RoutingConfig{
			Mode:                "manual",
			ComplexityThreshold: 0.8,
		},
	}

	defaults := Default()

	merged := MergeConfig(existing, defaults)

	// Should preserve existing values
	if merged.DefaultTool != "gemini" {
		t.Errorf("expected DefaultTool 'gemini', got '%s'", merged.DefaultTool)
	}

	if merged.Routing.Mode != "manual" {
		t.Errorf("expected Routing.Mode 'manual', got '%s'", merged.Routing.Mode)
	}

	if merged.Routing.ComplexityThreshold != 0.8 {
		t.Errorf("expected ComplexityThreshold 0.8, got %f", merged.Routing.ComplexityThreshold)
	}

	// Schema version should be preserved (MigrateConfig handles the upgrade)
	if merged.SchemaVersion != 1 {
		t.Errorf("expected SchemaVersion to be preserved as 1, got %d", merged.SchemaVersion)
	}
}

func TestMergeConfig_AddsMissingTools(t *testing.T) {
	existing := &Config{
		SchemaVersion: 1,
		DefaultTool:   "claude",
		Tools: map[string]Tool{
			"claude": {Enabled: true, Binary: "claude"},
		},
	}

	defaults := Default()

	merged := MergeConfig(existing, defaults)

	// Should have both existing and default tools
	if _, ok := merged.Tools["claude"]; !ok {
		t.Error("expected 'claude' tool to be preserved")
	}

	// Should add missing tools from defaults
	if _, ok := merged.Tools["gemini"]; !ok {
		t.Error("expected 'gemini' tool to be added from defaults")
	}
}

func TestMergeConfig_AddsMissingCommunity(t *testing.T) {
	existing := &Config{
		SchemaVersion: 1,
		DefaultTool:   "claude",
	}

	defaults := Default()

	merged := MergeConfig(existing, defaults)

	// Should add community config from defaults
	if merged.Community == (CommunityConfig{}) {
		t.Error("expected Community to be populated from defaults")
	}
}

func TestMergeConfig_PreservesExistingCommunity(t *testing.T) {
	existing := &Config{
		SchemaVersion: 1,
		DefaultTool:   "claude",
		Community: CommunityConfig{
			ShareEnabled:    true,
			AutoShareOnPush: false,
		},
	}

	defaults := Default()

	merged := MergeConfig(existing, defaults)

	// Should preserve existing community config
	if !merged.Community.ShareEnabled {
		t.Error("expected Community.ShareEnabled to be preserved as true")
	}
	if merged.Community.AutoShareOnPush {
		t.Error("expected Community.AutoShareOnPush to be preserved as false")
	}
}

func TestMergeConfig_FillsMissingDefaults(t *testing.T) {
	existing := &Config{
		SchemaVersion: 1,
		DefaultTool:   "claude",
		// Leave Search empty
	}

	defaults := Default()

	merged := MergeConfig(existing, defaults)

	// Should fill missing search defaults
	if merged.Search.Provider == "" {
		t.Error("expected Search.Provider to be filled from defaults")
	}
	if merged.Search.TopK == 0 {
		t.Error("expected Search.TopK to be filled from defaults")
	}
}
