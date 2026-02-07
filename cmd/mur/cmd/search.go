package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search patterns by keyword",
	Long: `Search patterns by name, description, tags, or content.

Examples:
  mur search error           # Find patterns about errors
  mur search "go testing"    # Multiple keywords
  mur search --tag backend   # Search by tag only
  mur search --domain go     # Search by domain`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

var (
	searchTag    string
	searchDomain string
	searchLimit  int
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "Filter by tag")
	searchCmd.Flags().StringVar(&searchDomain, "domain", "", "Filter by domain")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 20, "Max results")
}

type searchResult struct {
	pattern pattern.Pattern
	score   int
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(strings.Join(args, " "))

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	patterns, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to load patterns: %w", err)
	}

	if len(patterns) == 0 {
		fmt.Println("No patterns found.")
		fmt.Println("Create one with: mur new <name>")
		return nil
	}

	// Search and score
	var results []searchResult
	for _, p := range patterns {
		score := scorePattern(&p, query, searchTag, searchDomain)
		if score > 0 {
			results = append(results, searchResult{pattern: p, score: score})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Limit results
	if len(results) > searchLimit {
		results = results[:searchLimit]
	}

	if len(results) == 0 {
		fmt.Printf("No patterns found matching '%s'\n", query)
		return nil
	}

	fmt.Println()
	fmt.Printf("🔍 Found %d patterns matching '%s'\n", len(results), query)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, r := range results {
		p := r.pattern

		// Name and effectiveness
		eff := ""
		if p.Learning.Effectiveness > 0 {
			eff = fmt.Sprintf(" (%.0f%%)", p.Learning.Effectiveness*100)
		}
		fmt.Printf("📄 %s%s\n", p.Name, eff)

		// Description
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   %s\n", desc)
		}

		// Tags
		var tags []string
		tags = append(tags, p.Tags.Confirmed...)
		for _, t := range p.Tags.Inferred {
			if t.Confidence >= 0.7 {
				tags = append(tags, t.Tag)
			}
		}
		if len(tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(tags, ", "))
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("View:   mur learn get <name>")
	fmt.Println("Edit:   mur edit <name>")
	fmt.Println()

	return nil
}

func scorePattern(p *pattern.Pattern, query, tagFilter, domainFilter string) int {
	score := 0

	// Tag filter
	if tagFilter != "" {
		hasTag := false
		for _, t := range p.Tags.Confirmed {
			if strings.EqualFold(t, tagFilter) {
				hasTag = true
				break
			}
		}
		for _, t := range p.Tags.Inferred {
			if strings.EqualFold(t.Tag, tagFilter) && t.Confidence >= 0.7 {
				hasTag = true
				break
			}
		}
		if !hasTag {
			return 0
		}
		score += 5
	}

	// Domain filter (from tags)
	if domainFilter != "" {
		hasDomain := false
		domains := []string{"go", "swift", "python", "node", "rust", "javascript", "typescript"}
		for _, t := range p.Tags.Confirmed {
			if strings.EqualFold(t, domainFilter) {
				for _, d := range domains {
					if strings.EqualFold(t, d) {
						hasDomain = true
						break
					}
				}
			}
		}
		if !hasDomain {
			return 0
		}
		score += 5
	}

	// Empty query with filters only
	if query == "" {
		if tagFilter != "" || domainFilter != "" {
			return score + 1
		}
		return 0
	}

	// Name match (highest weight)
	nameLower := strings.ToLower(p.Name)
	if strings.Contains(nameLower, query) {
		score += 10
		if strings.HasPrefix(nameLower, query) {
			score += 5
		}
	}

	// Description match
	descLower := strings.ToLower(p.Description)
	if strings.Contains(descLower, query) {
		score += 5
	}

	// Tag match
	for _, t := range p.Tags.Confirmed {
		if strings.Contains(strings.ToLower(t), query) {
			score += 3
		}
	}
	for _, t := range p.Tags.Inferred {
		if strings.Contains(strings.ToLower(t.Tag), query) {
			score += 2
		}
	}

	// Content match (lowest weight)
	contentLower := strings.ToLower(p.Content)
	if strings.Contains(contentLower, query) {
		score += 1
	}

	return score
}
