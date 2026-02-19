package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/consolidate"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

var consolidateCmd = &cobra.Command{
	Use:   "consolidate",
	Short: "Consolidate patterns: score health, detect duplicates, resolve conflicts",
	Long: `Analyze all patterns for health, duplicates, and conflicts.

Default mode is --dry-run which shows what would happen without making changes.
Use --auto to apply safe actions (archive stale patterns, keep-best merges).
Use --interactive to step through each proposal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		autoFlag, _ := cmd.Flags().GetBool("auto")
		interactiveFlag, _ := cmd.Flags().GetBool("interactive")
		forceFlag, _ := cmd.Flags().GetBool("force")

		mode := consolidate.ModeDryRun
		if autoFlag {
			mode = consolidate.ModeAuto
		} else if interactiveFlag {
			mode = consolidate.ModeInteractive
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}

		murDir := filepath.Join(home, ".mur")
		patternsDir := filepath.Join(murDir, "patterns")
		trackingDir := filepath.Join(murDir, "tracking")

		// Load pattern store
		store := pattern.NewStore(patternsDir)

		// Load memory cache (patterns + embeddings)
		mc, err := cache.NewMemoryCache(cache.DefaultMemoryCacheOptions())
		if err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		// Ensure embeddings are loaded for dedup
		if mc != nil {
			if err := mc.EnsureEmbeddings(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not load embeddings: %v\n", err)
			}
		}

		// Create trackers
		injTracker := inject.NewTracker(store, trackingDir)
		analyticsTracker := analytics.NewTracker(murDir)

		// Get embedding matrix (may be nil)
		var matrix *cache.EmbeddingMatrix
		if mc != nil {
			matrix = mc.Embeddings
		}

		// Create and run consolidator
		c := consolidate.NewConsolidator(
			cfg.Consolidation,
			store,
			mc.Patterns,
			matrix,
			injTracker,
			analyticsTracker,
		)

		report, err := c.Run(mode, forceFlag)
		if err != nil {
			return fmt.Errorf("consolidation failed: %w", err)
		}

		// Build pattern name map for display
		nameMap := make(map[string]string)
		for _, p := range mc.Patterns.All() {
			nameMap[p.ID] = p.Name
		}

		fmt.Print(consolidate.FormatReport(report, nameMap))
		return nil
	},
}

func init() {
	consolidateCmd.Flags().Bool("auto", false, "apply safe actions automatically")
	consolidateCmd.Flags().Bool("interactive", false, "step through each proposal")
	consolidateCmd.Flags().Bool("force", false, "skip minimum patterns check")
	rootCmd.AddCommand(consolidateCmd)
}
