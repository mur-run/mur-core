package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/embed"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage pattern embedding index",
	Long: `Manage the embedding index for semantic pattern search.

The index stores vector embeddings of patterns for fast similarity search.
Requires Ollama running with the nomic-embed-text model.

Setup:
  brew install ollama
  ollama serve &
  ollama pull nomic-embed-text

Examples:
  mur index status           # Show index status
  mur index rebuild          # Rebuild all embeddings
  mur index pattern <name>   # Index a single pattern`,
}

var indexStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show index status",
	RunE:  runIndexStatus,
}

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild all embeddings",
	RunE:  runIndexRebuild,
}

var indexPatternCmd = &cobra.Command{
	Use:   "pattern <name>",
	Short: "Index a single pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  runIndexPattern,
}

var indexExpand bool

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.AddCommand(indexStatusCmd)
	indexCmd.AddCommand(indexRebuildCmd)
	indexCmd.AddCommand(indexPatternCmd)
	indexRebuildCmd.Flags().BoolVar(&indexExpand, "expand", false, "Generate search queries per pattern using LLM (slower but better search)")
}

func runIndexStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Search.IsEnabled() {
		fmt.Println("⚠️  Semantic search is disabled")
		fmt.Println("Enable with: mur config set search.enabled true")
		return nil
	}

	indexer, err := embed.NewPatternIndexer(cfg)
	if err != nil {
		return fmt.Errorf("cannot create indexer: %w", err)
	}

	status := indexer.Status()

	fmt.Println("📊 Pattern Index Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Provider status
	fmt.Printf("  Provider: %s\n", cfg.Search.Provider)
	if cfg.Search.Provider == "ollama" {
		if status.OllamaRunning {
			fmt.Printf("  Ollama: ✅ Running at %s\n", cfg.Search.OllamaURL)
			if status.ModelAvailable {
				fmt.Printf("  Model: ✅ %s\n", status.EmbeddingModel)
			} else {
				fmt.Printf("  Model: ❌ %s not found\n", status.EmbeddingModel)
				fmt.Printf("         Run: ollama pull %s\n", status.EmbeddingModel)
			}
		} else {
			fmt.Printf("  Ollama: ❌ Not running at %s\n", cfg.Search.OllamaURL)
			fmt.Println("         Run: ollama serve")
		}
	}

	fmt.Println()

	// Index stats
	if status.TotalPatterns > 0 {
		pct := status.IndexedCount * 100 / status.TotalPatterns
		fmt.Printf("  Patterns: %d total\n", status.TotalPatterns)
		fmt.Printf("  Indexed:  %d (%d%%)\n", status.IndexedCount, pct)
	} else {
		fmt.Println("  Patterns: 0")
	}

	if !status.LastUpdated.IsZero() {
		fmt.Printf("  Updated:  %s\n", status.LastUpdated.Format("2006-01-02 15:04"))
	}

	if status.CacheSize > 0 {
		fmt.Printf("  Cache:    %.1f KB\n", float64(status.CacheSize)/1024)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

func runIndexRebuild(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Search.IsEnabled() {
		return fmt.Errorf("semantic search is disabled, enable with: mur config set search.enabled true")
	}

	fmt.Println("🔄 Rebuilding pattern index...")
	fmt.Println()

	// Check prerequisites
	if cfg.Search.Provider == "ollama" {
		fmt.Print("  Checking Ollama... ")
		if !embed.IsOllamaRunning(cfg.Search.OllamaURL) {
			fmt.Println("❌")
			return fmt.Errorf("ollama is not running at %s\nStart with: ollama serve", cfg.Search.OllamaURL)
		}
		fmt.Println("✅")

		fmt.Printf("  Checking model %s... ", cfg.Search.Model)
		if !embed.HasOllamaModel(cfg.Search.OllamaURL, cfg.Search.Model) {
			fmt.Println("❌")
			return fmt.Errorf("model %s not found\nInstall with: ollama pull %s", cfg.Search.Model, cfg.Search.Model)
		}
		fmt.Println("✅")
	}

	fmt.Println()

	indexer, err := embed.NewPatternIndexer(cfg)
	if err != nil {
		return fmt.Errorf("cannot create indexer: %w", err)
	}

	start := time.Now()
	var lastProgress int

	if indexExpand {
		// Determine LLM model for expansion (use learn.model or fallback)
		llmModel := cfg.Learning.LLM.Model
		if llmModel == "" {
			llmModel = "qwen2.5:3b" // Fast small model for query generation
		}
		fmt.Printf("  📝 Expanding with LLM (%s)...\n\n", llmModel)

		err = indexer.RebuildWithExpansion(cfg.Search.OllamaURL, llmModel, func(current, total int, phase string) {
			pct := current * 100 / total
			if pct != lastProgress || current == total {
				lastProgress = pct
				bar := progressBar(current, total, 30)
				label := "expanding"
				if phase == "embedding" {
					label = "embedding"
				}
				fmt.Printf("\r  %s %d/%d [%s]", bar, current, total, label)
			}
		})
	} else {
		err = indexer.Rebuild(func(current, total int) {
			pct := current * 100 / total
			if pct != lastProgress && pct%5 == 0 {
				lastProgress = pct
				bar := progressBar(current, total, 30)
				fmt.Printf("\r  %s %d/%d", bar, current, total)
			}
		})
	}

	fmt.Println() // New line after progress

	if err != nil {
		return fmt.Errorf("rebuild failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Index rebuilt in %.1fs\n", elapsed.Seconds())

	return nil
}

func runIndexPattern(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Search.IsEnabled() {
		return fmt.Errorf("semantic search is disabled")
	}

	indexer, err := embed.NewPatternIndexer(cfg)
	if err != nil {
		return fmt.Errorf("cannot create indexer: %w", err)
	}

	// Get pattern
	store, err := embed.NewPatternStore()
	if err != nil {
		return err
	}

	p, err := store.Get(name)
	if err != nil {
		return fmt.Errorf("pattern not found: %s", name)
	}

	fmt.Printf("Indexing %s...\n", name)

	if err := indexer.IndexPattern(*p); err != nil {
		return err
	}

	// Save cache
	if err := indexer.SaveCache(); err != nil {
		return err
	}

	fmt.Printf("✅ Indexed %s\n", name)

	return nil
}

// progressBar creates an ASCII progress bar.
func progressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}

	filled := current * width / total
	empty := width - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}
