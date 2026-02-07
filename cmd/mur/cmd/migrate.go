package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate patterns to the latest schema version",
	Long: `Migrate converts patterns from older schema versions to the latest version.

Currently supports:
  - v1 → v2: Adds security metadata, multi-dimensional tags, and learning metrics

The migration:
  - Creates a backup of v1 patterns (in .backup-v1/)
  - Converts domain/category to inferred tags
  - Adds security hash and trust level
  - Sets up learning metadata
  - Updates schema version

Examples:
  # Check if migration is needed
  mur migrate --check

  # Dry run (show what would change)
  mur migrate --dry-run

  # Migrate all patterns
  mur migrate

  # Migrate without creating backup
  mur migrate --no-backup`,
	RunE: runMigrate,
}

var (
	migrateCheck    bool
	migrateDryRun   bool
	migrateNoBackup bool
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&migrateCheck, "check", false, "Check if migration is needed without migrating")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be migrated without making changes")
	migrateCmd.Flags().BoolVar(&migrateNoBackup, "no-backup", false, "Skip creating backup of v1 patterns")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	store, err := pattern.DefaultStore()
	if err != nil {
		return err
	}

	patternsDir := store.Dir()

	// Check mode
	if migrateCheck {
		return checkMigration(patternsDir)
	}

	// Check if migration is needed
	needsMigration, count, err := pattern.NeedsMigration(patternsDir)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needsMigration {
		fmt.Println("✅ All patterns are already at the latest schema version")
		return nil
	}

	fmt.Printf("📦 Found %d patterns that need migration (v1 → v2)\n\n", count)

	if migrateDryRun {
		fmt.Println("🔍 Dry run mode - no changes will be made")
		fmt.Println()
	}

	// Run migration
	options := pattern.MigrateOptions{
		CreateBackup: !migrateNoBackup,
		DryRun:       migrateDryRun,
	}

	result, err := pattern.Migrate(patternsDir, options)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Print results
	printMigrationResult(result)

	return nil
}

func checkMigration(patternsDir string) error {
	needsMigration, count, err := pattern.NeedsMigration(patternsDir)
	if err != nil {
		return err
	}

	if needsMigration {
		fmt.Printf("⚠️  Found %d patterns that need migration\n", count)
		fmt.Println("Run 'mur migrate' to upgrade to the latest schema")
	} else {
		fmt.Println("✅ All patterns are at the latest schema version")
	}

	return nil
}

func printMigrationResult(result *pattern.MigrationResult) {
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("📊 Migration Summary\n\n")

	if result.BackupDir != "" {
		fmt.Printf("📁 Backup: %s\n\n", result.BackupDir)
	}

	fmt.Printf("   Total:    %d\n", result.TotalPatterns)
	fmt.Printf("   Migrated: %d\n", result.MigratedCount)
	fmt.Printf("   Skipped:  %d (already v2)\n", result.SkippedCount)
	if result.ErrorCount > 0 {
		fmt.Printf("   Errors:   %d\n", result.ErrorCount)
	}

	// Print migrated files
	if len(result.MigratedFiles) > 0 {
		fmt.Println("\n✅ Migrated patterns:")
		for _, f := range result.MigratedFiles {
			fmt.Printf("   - %s\n", f)
		}
	}

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Println("\n❌ Errors:")
		for _, e := range result.Errors {
			fmt.Printf("   - %s: %s\n", e.File, e.Error)
		}
	}

	if result.ErrorCount == 0 && result.MigratedCount > 0 {
		fmt.Println("\n✅ Migration complete!")
		fmt.Println("Run 'mur lint' to verify the migrated patterns")
	}
}
