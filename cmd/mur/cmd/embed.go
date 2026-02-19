package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/core/embed"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

var embedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Manage pattern embeddings for semantic search",
	Long: `Manage embeddings for semantic pattern search.

Embeddings enable semantic (meaning-based) pattern matching,
finding relevant patterns even when keywords don't match exactly.

Requires an embedding provider:
  - ollama (default, local): ollama pull nomic-embed-text
  - openai: Set OPENAI_API_KEY

Examples:
  mur embed index             # Index all patterns
  mur embed status            # Show embedding status
  mur embed search "query"    # Test semantic search
  mur embed rehash            # Rebuild all embeddings`,
}

var embedIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index patterns for semantic search",
	RunE:  embedIndexExecute,
}

var embedStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show embedding status",
	RunE:  embedStatusExecute,
}

var embedSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Test semantic search",
	Args:  cobra.ExactArgs(1),
	RunE:  embedSearchExecute,
}

var embedRehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rebuild all embeddings",
	RunE:  embedRehashExecute,
}

func getEmbedConfig() embed.Config {
	cfg := embed.DefaultConfig()

	// Check for OpenAI key
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.Provider = "openai"
		cfg.Model = "text-embedding-3-small"
		cfg.APIKey = key
	}

	return cfg
}

func embedIndexExecute(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := getEmbedConfig()
	fmt.Printf("Using %s embeddings...\n", cfg.Provider)

	searcher, err := embed.NewPatternSearcher(store, cfg)
	if err != nil {
		return fmt.Errorf("failed to create searcher: %w", err)
	}

	fmt.Println("Indexing patterns...")
	if err := searcher.IndexPatterns(); err != nil {
		return fmt.Errorf("failed to index: %w", err)
	}

	status, _ := searcher.Status()
	fmt.Printf("✓ Indexed %d/%d patterns\n", status.Indexed, status.TotalPatterns)

	return nil
}

func embedStatusExecute(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := getEmbedConfig()
	searcher, err := embed.NewPatternSearcher(store, cfg)
	if err != nil {
		return fmt.Errorf("failed to create searcher: %w", err)
	}

	status, err := searcher.Status()
	if err != nil {
		return err
	}

	fmt.Println("Embedding Status")
	fmt.Println("================")
	fmt.Printf("Provider:   %s\n", status.Provider)
	fmt.Printf("Dimension:  %d\n", status.Dimension)
	fmt.Printf("Patterns:   %d total\n", status.TotalPatterns)
	fmt.Printf("Indexed:    %d (%.0f%%)\n", status.Indexed,
		float64(status.Indexed)/float64(max(status.TotalPatterns, 1))*100)

	if status.Indexed < status.TotalPatterns {
		fmt.Println("\nRun 'mur embed index' to index all patterns")
	}

	return nil
}

func embedSearchExecute(cmd *cobra.Command, args []string) error {
	query := args[0]
	topK, _ := cmd.Flags().GetInt("top")

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := getEmbedConfig()
	searcher, err := embed.NewPatternSearcher(store, cfg)
	if err != nil {
		return fmt.Errorf("failed to create searcher: %w", err)
	}

	matches, err := searcher.Search(query, topK)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matching patterns found.")
		fmt.Println("Run 'mur embed index' to index patterns first.")
		return nil
	}

	fmt.Printf("Semantic search: \"%s\"\n", query)
	fmt.Println()

	for i, m := range matches {
		scoreBar := makeBar(m.Score, 10)
		fmt.Printf("%d. %s  %s %.0f%%\n", i+1, m.Pattern.Name, scoreBar, m.Score*100)
		if m.Pattern.Description != "" {
			fmt.Printf("   %s\n", truncateStr(m.Pattern.Description, 60))
		}
	}

	return nil
}

func embedRehashExecute(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := getEmbedConfig()
	fmt.Printf("Using %s embeddings...\n", cfg.Provider)

	searcher, err := embed.NewPatternSearcher(store, cfg)
	if err != nil {
		return fmt.Errorf("failed to create searcher: %w", err)
	}

	fmt.Println("Rebuilding embeddings...")
	if err := searcher.Rehash(); err != nil {
		return fmt.Errorf("failed to rehash: %w", err)
	}

	status, _ := searcher.Status()
	fmt.Printf("✓ Rebuilt %d embeddings\n", status.Indexed)

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func init() {
	embedCmd.Hidden = true
	rootCmd.AddCommand(embedCmd)
	embedCmd.AddCommand(embedIndexCmd)
	embedCmd.AddCommand(embedStatusCmd)
	embedCmd.AddCommand(embedSearchCmd)
	embedCmd.AddCommand(embedRehashCmd)

	embedSearchCmd.Flags().Int("top", 5, "Number of results to return")
}
