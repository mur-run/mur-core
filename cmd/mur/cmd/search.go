package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/embed"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search patterns (local + community)",
	Long: `Search patterns using semantic similarity.

By default, searches local patterns. Use --community to also search
community patterns from mur.run.

Examples:
  mur search "Swift async testing"           # Local only
  mur search --community "API retry"         # Local + community
  mur search --community-only "error handling"  # Community only
  mur search --top 5 "Docker best practices"
  mur search --json "database optimization"
  mur search --inject "$PROMPT"              # For hooks`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var (
	searchTopK         int
	searchJSON         bool
	searchInject       bool
	searchCommunity    bool
	searchCommunityOnly bool
	searchLocalOnly    bool
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVar(&searchTopK, "top", 0, "Number of results (default: from config)")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output as JSON")
	searchCmd.Flags().BoolVar(&searchInject, "inject", false, "Inject mode for hooks (output to stderr)")
	searchCmd.Flags().BoolVar(&searchCommunity, "community", false, "Also search community patterns")
	searchCmd.Flags().BoolVar(&searchCommunityOnly, "community-only", false, "Only search community patterns")
	searchCmd.Flags().BoolVar(&searchLocalOnly, "local", false, "Only search local patterns (default)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Use config default if not specified
	topK := searchTopK
	if topK == 0 {
		topK = cfg.Search.TopK
	}
	if topK == 0 {
		topK = 5
	}

	var localMatches []embed.PatternMatch
	var communityResults []cloud.CommunityPattern

	// Search local patterns (unless community-only)
	if !searchCommunityOnly {
		if cfg.Search.IsEnabled() {
			indexer, err := embed.NewPatternIndexer(cfg)
			if err == nil {
				status := indexer.Status()
				if status.IndexedCount > 0 {
					localMatches, _ = indexer.Search(query, topK)
				}
			}
		}
	}

	// Search community (if requested)
	if searchCommunity || searchCommunityOnly {
		client, err := cloud.NewClient(cfg.Server.URL)
		if err == nil {
			resp, err := client.SearchCommunity(query, topK)
			if err == nil {
				communityResults = resp.Patterns
			}
		}
	}

	// Record analytics for local matches
	if len(localMatches) > 0 {
		home, _ := os.UserHomeDir()
		tracker := analytics.NewTracker(filepath.Join(home, ".mur"))
		for _, m := range localMatches {
			if m.Score >= cfg.Search.MinScore {
				_ = tracker.RecordSearch(m.Pattern.ID, m.Pattern.Name, m.Score, query)
			}
		}
	}

	// Inject mode - output to stderr for hooks
	if searchInject {
		if len(localMatches) == 0 && len(communityResults) == 0 {
			return nil
		}

		var names []string
		for _, m := range localMatches {
			names = append(names, m.Pattern.Name)
		}
		for _, c := range communityResults {
			names = append(names, c.Name+" 🌐")
		}

		hint := fmt.Sprintf("[mur] 🎯 Relevant patterns: %s\n", strings.Join(names, ", "))
		if len(localMatches) > 0 {
			hint += fmt.Sprintf("[mur] 💡 Consider loading /%s for this task\n", getSkillPath(localMatches[0]))
		}
		if len(communityResults) > 0 && len(localMatches) == 0 {
			hint += fmt.Sprintf("[mur] 💡 Community pattern available: mur community copy \"%s\"\n", communityResults[0].Name)
		}

		// Cache community patterns for future use
		if len(communityResults) > 0 {
			go cacheCommunityPatterns(cfg, communityResults)
		}

		fmt.Fprint(os.Stderr, hint)
		return nil
	}

	// JSON output
	if searchJSON {
		output := map[string]interface{}{
			"local":     make([]map[string]interface{}, len(localMatches)),
			"community": make([]map[string]interface{}, len(communityResults)),
		}
		localOut := output["local"].([]map[string]interface{})
		for i, m := range localMatches {
			localOut[i] = map[string]interface{}{
				"name":        m.Pattern.Name,
				"description": m.Pattern.Description,
				"score":       m.Score,
				"source":      "local",
			}
		}
		communityOut := output["community"].([]map[string]interface{})
		for i, c := range communityResults {
			communityOut[i] = map[string]interface{}{
				"id":          c.ID,
				"name":        c.Name,
				"description": c.Description,
				"author":      c.AuthorName,
				"copies":      c.CopyCount,
				"source":      "community",
			}
		}
		return json.NewEncoder(os.Stdout).Encode(output)
	}

	// Pretty print
	fmt.Println("🔍 Searching patterns...")
	fmt.Println()

	if len(localMatches) > 0 {
		fmt.Println("📍 Local patterns:")
		for i, m := range localMatches {
			fmt.Printf("  %d. %s (%.2f)\n", i+1, m.Pattern.Name, m.Score)
			if m.Pattern.Description != "" {
				desc := m.Pattern.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Printf("     %s\n", desc)
			}
		}
		fmt.Println()
	}

	if len(communityResults) > 0 {
		fmt.Println("🌐 Community patterns:")
		for i, c := range communityResults {
			author := c.AuthorName
			if c.AuthorLogin != "" {
				author = "@" + c.AuthorLogin
			}
			fmt.Printf("  %d. %s (⭐ %d) by %s\n", i+1, c.Name, c.CopyCount, author)
			if c.Description != "" {
				desc := c.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Printf("     %s\n", desc)
			}
		}
		fmt.Println()
		fmt.Println("  💡 Use 'mur community copy <name>' to add to your patterns")
		fmt.Println()
	}

	total := len(localMatches) + len(communityResults)
	if total == 0 {
		fmt.Printf("No patterns found for %q\n", query)
		if !searchCommunity && !searchCommunityOnly {
			fmt.Println("💡 Try: mur search --community \"" + query + "\"")
		}
	} else {
		fmt.Printf("Found %d patterns matching %q\n", total, query)
	}

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

// cacheCommunityPatterns fetches and caches community pattern content.
func cacheCommunityPatterns(cfg *config.Config, patterns []cloud.CommunityPattern) {
	cacheConfig := cfg.GetCacheConfig()
	if !cacheConfig.Enabled {
		return
	}

	communityCache, err := cache.DefaultCommunityCache()
	if err != nil {
		return
	}

	client, err := cloud.NewClient(cfg.Server.URL)
	if err != nil {
		return
	}

	// Cache top 3 patterns (to not overwhelm the server)
	limit := 3
	if len(patterns) < limit {
		limit = len(patterns)
	}

	for _, p := range patterns[:limit] {
		// Check if already cached
		cached, _ := communityCache.Get(p.ID)
		if cached != nil {
			continue // Already in cache
		}

		// Fetch full pattern
		detail, err := client.GetCommunityPattern(p.ID)
		if err != nil {
			continue
		}

		// Cache it
		communityCache.Save(&cache.CachedPattern{
			ID:          detail.ID,
			Name:        detail.Name,
			Description: detail.Description,
			Content:     detail.Content,
			Author:      detail.AuthorName,
		})
	}
}
