package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/embed"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Semantic search for patterns",
	Long: `Search patterns using semantic similarity.

Requires indexed patterns. Run 'mur index rebuild' first.

Examples:
  mur search "Swift async testing"
  mur search --top 5 "Docker compose best practices"
  mur search --json "database optimization"
  mur search --inject "$PROMPT"   # For hooks (outputs to stderr)`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var (
	searchTopK   int
	searchJSON   bool
	searchInject bool
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVar(&searchTopK, "top", 0, "Number of results (default: from config)")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output as JSON")
	searchCmd.Flags().BoolVar(&searchInject, "inject", false, "Inject mode for hooks (output to stderr)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.Search.IsEnabled() {
		if searchInject {
			return nil // Silent fail for hooks
		}
		return fmt.Errorf("semantic search is disabled")
	}

	// Use config default if not specified
	topK := searchTopK
	if topK == 0 {
		topK = cfg.Search.TopK
	}
	if topK == 0 {
		topK = 3
	}

	indexer, err := embed.NewPatternIndexer(cfg)
	if err != nil {
		if searchInject {
			return nil // Silent fail for hooks
		}
		return fmt.Errorf("cannot create indexer: %w", err)
	}

	// Check if we have indexed patterns
	status := indexer.Status()
	if status.IndexedCount == 0 {
		if searchInject {
			return nil // Silent fail
		}
		return fmt.Errorf("no indexed patterns, run: mur index rebuild")
	}

	matches, err := indexer.Search(query, topK)
	if err != nil {
		if searchInject {
			return nil // Silent fail for hooks
		}
		return fmt.Errorf("search failed: %w", err)
	}

	// Record analytics for matches
	if len(matches) > 0 {
		home, _ := os.UserHomeDir()
		tracker := analytics.NewTracker(filepath.Join(home, ".mur"))
		for _, m := range matches {
			// Only record if above min_score
			if m.Score >= cfg.Search.MinScore {
				_ = tracker.RecordSearch(m.Pattern.ID, m.Pattern.Name, m.Score, query)
			}
		}
	}

	// Inject mode - output to stderr for hooks
	if searchInject {
		if len(matches) == 0 {
			return nil
		}

		var names []string
		for _, m := range matches {
			names = append(names, m.Pattern.Name)
		}

		hint := fmt.Sprintf("[mur] 🎯 Relevant patterns: %s\n", strings.Join(names, ", "))
		hint += fmt.Sprintf("[mur] 💡 Consider loading /%s for this task\n", getSkillPath(matches[0]))

		fmt.Fprint(os.Stderr, hint)
		return nil
	}

	// JSON output
	if searchJSON {
		output := make([]map[string]interface{}, len(matches))
		for i, m := range matches {
			output[i] = map[string]interface{}{
				"name":        m.Pattern.Name,
				"description": m.Pattern.Description,
				"score":       m.Score,
				"skill_path":  getSkillPath(m),
			}
		}
		return json.NewEncoder(os.Stdout).Encode(output)
	}

	// Pretty print
	fmt.Println("🔍 Searching patterns...")
	fmt.Println()

	for i, m := range matches {
		fmt.Printf("  %d. %s (%.2f)\n", i+1, m.Pattern.Name, m.Score)
		if m.Pattern.Description != "" {
			fmt.Printf("     %s\n", m.Pattern.Description)
		}
		fmt.Println()
	}

	fmt.Printf("Found %d patterns matching %q\n", len(matches), query)

	return nil
}

// getSkillPath returns the skill directory path for a pattern.
func getSkillPath(m embed.PatternMatch) string {
	domain := m.Pattern.GetPrimaryDomain()
	name := strings.ToLower(m.Pattern.Name)
	name = strings.ReplaceAll(name, " ", "-")

	if domain == "general" || domain == "" {
		return name
	}

	// Check if name already starts with domain
	domainPrefix := domain + "-"
	if strings.HasPrefix(name, domainPrefix) {
		return domain + "--" + strings.TrimPrefix(name, domainPrefix)
	}

	return domain + "--" + name
}
