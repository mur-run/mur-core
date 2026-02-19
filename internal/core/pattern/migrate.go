package pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// V1Pattern represents the old pattern schema (v1).
type V1Pattern struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Content     string  `yaml:"content"`
	Domain      string  `yaml:"domain"`
	Category    string  `yaml:"category"`
	Confidence  float64 `yaml:"confidence"`
	TeamShared  bool    `yaml:"team_shared"`
	CreatedAt   string  `yaml:"created_at"`
	UpdatedAt   string  `yaml:"updated_at"`
}

// MigrationResult holds the result of migrating patterns.
type MigrationResult struct {
	TotalPatterns int
	MigratedCount int
	SkippedCount  int
	ErrorCount    int
	Errors        []MigrationError
	MigratedFiles []string
	BackupDir     string
}

// MigrationError represents an error during migration.
type MigrationError struct {
	File    string
	Pattern string
	Error   string
}

// Migrate converts v1 patterns to v2 in the given directory.
func Migrate(patternsDir string, options MigrateOptions) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Check if directory exists
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		return result, nil // Nothing to migrate
	}

	// Create backup if requested
	if options.CreateBackup {
		backupDir := filepath.Join(patternsDir, ".backup-v1", time.Now().Format("20060102-150405"))
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}
		result.BackupDir = backupDir
	}

	// Read all pattern files
	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(patternsDir, entry.Name())
		result.TotalPatterns++

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				File:  entry.Name(),
				Error: fmt.Sprintf("failed to read: %v", err),
			})
			continue
		}

		// Try to parse as v2 first (already migrated)
		var v2Check Pattern
		if err := yaml.Unmarshal(data, &v2Check); err == nil && v2Check.SchemaVersion >= 2 {
			result.SkippedCount++
			continue
		}

		// Parse as v1
		var v1 V1Pattern
		if err := yaml.Unmarshal(data, &v1); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				File:  entry.Name(),
				Error: fmt.Sprintf("failed to parse: %v", err),
			})
			continue
		}

		// Backup if requested
		if options.CreateBackup && result.BackupDir != "" {
			backupPath := filepath.Join(result.BackupDir, entry.Name())
			if err := os.WriteFile(backupPath, data, 0644); err != nil {
				result.Errors = append(result.Errors, MigrationError{
					File:  entry.Name(),
					Error: fmt.Sprintf("failed to backup: %v", err),
				})
			}
		}

		// Convert to v2
		v2, err := migrateV1ToV2(v1, options)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				File:    entry.Name(),
				Pattern: v1.Name,
				Error:   fmt.Sprintf("migration failed: %v", err),
			})
			continue
		}

		// Write v2 pattern
		v2Data, err := yaml.Marshal(v2)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				File:    entry.Name(),
				Pattern: v1.Name,
				Error:   fmt.Sprintf("failed to serialize: %v", err),
			})
			continue
		}

		if !options.DryRun {
			if err := os.WriteFile(filePath, v2Data, 0644); err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, MigrationError{
					File:    entry.Name(),
					Pattern: v1.Name,
					Error:   fmt.Sprintf("failed to write: %v", err),
				})
				continue
			}
		}

		result.MigratedCount++
		result.MigratedFiles = append(result.MigratedFiles, entry.Name())
	}

	return result, nil
}

// MigrateOptions holds options for migration.
type MigrateOptions struct {
	CreateBackup bool
	DryRun       bool
	DefaultOwner string
}

// migrateV1ToV2 converts a v1 pattern to v2.
func migrateV1ToV2(v1 V1Pattern, options MigrateOptions) (*Pattern, error) {
	now := time.Now()

	// Parse timestamps
	var created, updated time.Time
	if t, err := time.Parse(time.RFC3339, v1.CreatedAt); err == nil {
		created = t
	} else {
		created = now
	}
	if t, err := time.Parse(time.RFC3339, v1.UpdatedAt); err == nil {
		updated = t
	} else {
		updated = now
	}

	// Map domain to inferred tag
	var inferredTags []TagScore
	if v1.Domain != "" && v1.Domain != "general" {
		inferredTags = append(inferredTags, TagScore{
			Tag:        v1.Domain,
			Confidence: v1.Confidence,
		})
	}
	if v1.Category != "" && v1.Category != "pattern" {
		inferredTags = append(inferredTags, TagScore{
			Tag:        v1.Category,
			Confidence: v1.Confidence,
		})
	}

	// Determine trust level
	trustLevel := TrustOwner
	if options.DefaultOwner == "" {
		trustLevel = TrustOwner
	}

	p := &Pattern{
		ID:          uuid.New().String(),
		Name:        v1.Name,
		Description: v1.Description,
		Content:     v1.Content,
		Tags: TagSet{
			Inferred: inferredTags,
		},
		Applies: ApplyConditions{},
		Security: SecurityMeta{
			TrustLevel: trustLevel,
			Reviewed:   true, // Existing patterns are considered reviewed
			Risk:       RiskLow,
		},
		Learning: LearningMeta{
			Effectiveness:      v1.Confidence,
			UsageCount:         0, // Reset usage count
			OriginalConfidence: v1.Confidence,
		},
		Lifecycle: LifecycleMeta{
			Status:  StatusActive,
			Created: created,
			Updated: updated,
		},
		SchemaVersion: SchemaVersion,
		Version:       "1.0.0", // Default version for migrated patterns
	}

	// Calculate hashes
	p.UpdateHash()
	p.UpdateEmbeddingHash()

	// Infer resources
	p.InferResources()

	return p, nil
}

// DetectVersion returns the schema version of a pattern file.
func DetectVersion(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	// Try v2
	var v2 Pattern
	if err := yaml.Unmarshal(data, &v2); err == nil && v2.SchemaVersion >= 2 {
		return v2.SchemaVersion, nil
	}

	// Try v1
	var v1 V1Pattern
	if err := yaml.Unmarshal(data, &v1); err == nil && v1.Name != "" {
		return 1, nil
	}

	return 0, fmt.Errorf("unknown schema version")
}

// NeedsMigration checks if the patterns directory needs migration.
func NeedsMigration(patternsDir string) (bool, int, error) {
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		return false, 0, nil
	}

	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		return false, 0, err
	}

	v1Count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		version, err := DetectVersion(filepath.Join(patternsDir, entry.Name()))
		if err != nil {
			continue
		}

		if version < SchemaVersion {
			v1Count++
		}
	}

	return v1Count > 0, v1Count, nil
}
