package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/analytics"
)

var statsCmd = &cobra.Command{
	Use:   "stats [pattern-name]",
	Short: "Show pattern usage analytics",
	Long: `Display analytics and effectiveness metrics for patterns.

Without arguments, shows overall statistics.
With a pattern name, shows detailed stats for that pattern.`,
	Example: `  mur stats                    # Show overall analytics
  mur stats swift-testing      # Show detailed stats for a pattern`,
	RunE: runStats,
}

var (
	statsDays int
)

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().IntVarP(&statsDays, "days", "d", 30, "Number of days to analyze")
}

func runStats(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(home, ".mur")
	store, err := analytics.NewStore(dataDir)
	if err != nil {
		return fmt.Errorf("failed to open analytics store: %w", err)
	}
	defer store.Close()

	if len(args) > 0 {
		return showPatternStats(store, args[0])
	}

	return showOverallStats(store, statsDays)
}

func showOverallStats(store *analytics.Store, days int) error {
	overall, err := store.GetOverallStats(days)
	if err != nil {
		return fmt.Errorf("failed to get overall stats: %w", err)
	}

	allStats, err := store.GetAllStats(100)
	if err != nil {
		return fmt.Errorf("failed to get pattern stats: %w", err)
	}

	// Calculate active patterns (used in last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	activeCount := 0
	for _, s := range allStats {
		if s.LastUsed != nil && s.LastUsed.After(sevenDaysAgo) {
			activeCount++
		}
	}

	fmt.Printf("\n📊 Pattern Analytics (last %d days)\n", days)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Total Patterns: %d\n", overall.TotalPatterns)
	fmt.Printf("Active Patterns: %d (used in last 7 days)\n", activeCount)
	fmt.Printf("Total Injections: %d\n", overall.TotalInjections)
	fmt.Println()

	if len(allStats) == 0 {
		fmt.Println("No usage data yet. Patterns will be tracked when injected.")
		fmt.Println()
		fmt.Println("💡 Tip: Run 'mur init --hooks' to set up automatic tracking.")
		return nil
	}

	// Top 5 most used
	fmt.Println("Top 5 Most Used:")
	topCount := 5
	if len(allStats) < topCount {
		topCount = len(allStats)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i := 0; i < topCount; i++ {
		s := allStats[i]
		effectiveness := "N/A"
		if s.HelpfulCount+s.NotHelpfulCount > 0 {
			effectiveness = fmt.Sprintf("%.0f%%", s.Effectiveness*100)
		}
		fmt.Fprintf(w, "  %d. %s\t│ %d uses\t│ %s effective\n",
			i+1, truncateStr(s.PatternName, 25), s.UsageCount, effectiveness)
	}
	w.Flush()
	fmt.Println()

	// Patterns needing review (low effectiveness)
	var needsReview []*analytics.PatternStats
	for _, s := range allStats {
		total := s.HelpfulCount + s.NotHelpfulCount
		if total >= 5 && s.Effectiveness < 0.6 {
			needsReview = append(needsReview, s)
		}
	}

	if len(needsReview) > 0 {
		fmt.Println("Needs Review (low effectiveness):")
		for _, s := range needsReview {
			fmt.Printf("  ⚠️  %s\t│ %d uses\t│ %.0f%% effective\n",
				truncateStr(s.PatternName, 25), s.UsageCount, s.Effectiveness*100)
		}
		fmt.Println()
	}

	fmt.Println("💡 Tip: Run 'mur feedback' to rate patterns after use")
	fmt.Println()

	return nil
}

func showPatternStats(store *analytics.Store, patternName string) error {
	// Try to find pattern by name or ID
	allStats, err := store.GetAllStats(1000)
	if err != nil {
		return err
	}

	var found *analytics.PatternStats
	for _, s := range allStats {
		if strings.EqualFold(s.PatternName, patternName) || s.PatternID == patternName {
			found = s
			break
		}
		if strings.Contains(strings.ToLower(s.PatternName), strings.ToLower(patternName)) {
			found = s
			// Continue looking for exact match
		}
	}

	if found == nil {
		return fmt.Errorf("pattern '%s' not found in analytics data", patternName)
	}

	// Get detailed stats
	stats, err := store.GetPatternStats(found.PatternID)
	if err != nil {
		return err
	}

	byTool, err := store.GetUsageByTool(found.PatternID)
	if err != nil {
		return err
	}

	byContext, err := store.GetUsageByContext(found.PatternID)
	if err != nil {
		return err
	}

	fmt.Printf("\n📊 %s\n", stats.PatternName)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// Effectiveness
	total := stats.HelpfulCount + stats.NotHelpfulCount
	if total > 0 {
		fmt.Printf("Effectiveness: %.0f%% (%d helpful / %d rated)\n",
			stats.Effectiveness*100, stats.HelpfulCount, total)
	} else {
		fmt.Println("Effectiveness: N/A (no feedback yet)")
	}

	fmt.Printf("Total Uses: %d\n", stats.UsageCount)
	if stats.LastUsed != nil {
		fmt.Printf("Last Used: %s\n", formatTimeAgo(*stats.LastUsed))
	}
	fmt.Println()

	// Usage by tool
	if len(byTool) > 0 {
		fmt.Println("Usage by Tool:")
		maxCount := 0
		for _, count := range byTool {
			if count > maxCount {
				maxCount = count
			}
		}
		for tool, count := range byTool {
			bar := makeBarInt(count, maxCount, 20)
			pct := float64(count) / float64(stats.UsageCount) * 100
			fmt.Printf("  %-10s %s %d (%.0f%%)\n", tool, bar, count, pct)
		}
		fmt.Println()
	}

	// Usage by context
	if len(byContext) > 0 {
		fmt.Println("Usage by Context:")
		maxCount := 0
		for _, count := range byContext {
			if count > maxCount {
				maxCount = count
			}
		}
		for ctx, count := range byContext {
			bar := makeBarInt(count, maxCount, 20)
			pct := float64(count) / float64(stats.UsageCount) * 100
			fmt.Printf("  %-10s %s %d (%.0f%%)\n", ctx, bar, count, pct)
		}
		fmt.Println()
	}

	return nil
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

// makeBar creates a progress bar (also used by cross_learn.go, embed.go, suggest.go)
func makeBar(value float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	filled := int(value * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func makeBarInt(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := value * width / max
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
