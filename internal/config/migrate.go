package config

import "fmt"

// MigrationChange describes a single config change during migration.
type MigrationChange struct {
	Field       string
	Description string
}

// MigrateConfig migrates config to the latest schema version.
// Returns true if any changes were made, along with a list of changes.
func MigrateConfig(cfg *Config) (changed bool, changes []MigrationChange) {
	startVersion := cfg.SchemaVersion

	// Configs without schema_version are v1 (pre-versioning)
	if cfg.SchemaVersion == 0 {
		cfg.SchemaVersion = 1
	}

	// v1 → v2: Add community settings
	if cfg.SchemaVersion < 2 {
		// Only set defaults if community config is empty (zero value)
		if cfg.Community == (CommunityConfig{}) {
			cfg.Community = DefaultCommunityConfig()
			changes = append(changes, MigrationChange{
				Field:       "community.share_enabled",
				Description: fmt.Sprintf("default: %v", cfg.Community.ShareEnabled),
			})
			changes = append(changes, MigrationChange{
				Field:       "community.auto_share_on_push",
				Description: fmt.Sprintf("default: %v", cfg.Community.AutoShareOnPush),
			})
		}
		cfg.SchemaVersion = 2
	}

	// Future migrations go here:
	// if cfg.SchemaVersion < 3 { ... }

	changed = cfg.SchemaVersion != startVersion
	return changed, changes
}

// NeedsMigration returns true if the config needs migration.
func NeedsMigration(cfg *Config) bool {
	return cfg.SchemaVersion < CurrentSchemaVersion
}
