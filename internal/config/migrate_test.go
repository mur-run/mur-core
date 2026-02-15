package config

import "testing"

func TestMigrateConfig_V0ToV2(t *testing.T) {
	// Config without schema_version is treated as v1
	cfg := &Config{}

	changed, changes := MigrateConfig(cfg)

	if !changed {
		t.Error("expected migration to report changes")
	}

	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected schema version %d, got %d", CurrentSchemaVersion, cfg.SchemaVersion)
	}

	if len(changes) == 0 {
		t.Error("expected at least one change to be reported")
	}

	// Check community settings were added
	if cfg.Community == (CommunityConfig{}) {
		t.Error("expected community config to be populated")
	}
}

func TestMigrateConfig_V1ToV2(t *testing.T) {
	cfg := &Config{
		SchemaVersion: 1,
	}

	changed, changes := MigrateConfig(cfg)

	if !changed {
		t.Error("expected migration to report changes")
	}

	if cfg.SchemaVersion != 2 {
		t.Errorf("expected schema version 2, got %d", cfg.SchemaVersion)
	}

	// Should have added community settings
	foundCommunity := false
	for _, c := range changes {
		if c.Field == "community.share_enabled" {
			foundCommunity = true
			break
		}
	}
	if !foundCommunity {
		t.Error("expected community.share_enabled to be in changes")
	}
}

func TestMigrateConfig_AlreadyLatest(t *testing.T) {
	cfg := &Config{
		SchemaVersion: CurrentSchemaVersion,
		Community:     DefaultCommunityConfig(),
	}

	changed, changes := MigrateConfig(cfg)

	if changed {
		t.Error("expected no changes for already up-to-date config")
	}

	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestMigrateConfig_PreservesExistingCommunity(t *testing.T) {
	cfg := &Config{
		SchemaVersion: 1,
		Community: CommunityConfig{
			ShareEnabled:    true,
			AutoShareOnPush: false,
		},
	}

	changed, _ := MigrateConfig(cfg)

	if !changed {
		t.Error("expected migration to report version change")
	}

	// Should NOT overwrite existing community settings
	if !cfg.Community.ShareEnabled {
		t.Error("existing community.share_enabled should be preserved")
	}
	if cfg.Community.AutoShareOnPush {
		t.Error("existing community.auto_share_on_push should be preserved")
	}
}

func TestNeedsMigration(t *testing.T) {
	tests := []struct {
		name     string
		version  int
		expected bool
	}{
		{"v0 needs migration", 0, true},
		{"v1 needs migration", 1, true},
		{"current version no migration", CurrentSchemaVersion, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{SchemaVersion: tt.version}
			if NeedsMigration(cfg) != tt.expected {
				t.Errorf("NeedsMigration() = %v, expected %v", NeedsMigration(cfg), tt.expected)
			}
		})
	}
}
