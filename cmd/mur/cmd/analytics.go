package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/spf13/cobra"
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View pattern usage analytics",
	Long: `View pattern usage analytics and statistics.

Analytics are collected automatically when patterns are searched or injected.

Examples:
  mur analytics              # Show summary
  mur analytics top          # Show top patterns
  mur analytics cold         # Show unused patterns
  mur analytics feedback     # Record feedback on patterns`,
	RunE: runAnalyticsSummary,
}

var analyticsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top used patterns",
	RunE:  runAnalyticsTop,
}

var analyticsColdCmd = &cobra.Command{
	Use:   "cold",
	Short: "Show patterns not used recently",
	RunE:  runAnalyticsCold,
}

var analyticsFeedbackCmd = &cobra.Command{
	Use:   "feedback <pattern> <helpful|not-helpful>",
	Short: "Record feedback on a pattern",
	Args:  cobra.ExactArgs(2),
	RunE:  runAnalyticsFeedback,
}

var (
	analyticsJSON bool
	analyticsTop  int
	analyticsDays int
)

func init() {
	rootCmd.AddCommand(analyticsCmd)
	analyticsCmd.AddCommand(analyticsTopCmd)
	analyticsCmd.AddCommand(analyticsColdCmd)
	analyticsCmd.AddCommand(analyticsFeedbackCmd)

	analyticsCmd.PersistentFlags().BoolVar(&analyticsJSON, "json", false, "Output as JSON")
	analyticsTopCmd.Flags().IntVar(&analyticsTop, "limit", 10, "Number of patterns to show")
	analyticsColdCmd.Flags().IntVar(&analyticsDays, "days", 30, "Days threshold for cold patterns")
}

func getTracker() *analytics.Tracker {
	home, _ := os.UserHomeDir()
	return analytics.NewTracker(filepath.Join(home, ".mur"))
}

func runAnalyticsSummary(_ *cobra.Command, _ []string) error {
	tracker := getTracker()
	summary, err := tracker.GetSummary()
	if err != nil {
		return fmt.Errorf("get summary: %w", err)
	}

	if analyticsJSON {
		return json.NewEncoder(os.Stdout).Encode(summary)
	}

	fmt.Println("📊 Pattern Analytics")
	fmt.Println("====================")
	fmt.Println()

	if summary.TotalEvents == 0 {
		fmt.Println("No analytics data yet.")
		fmt.Println()
		fmt.Println("Analytics are collected automatically when you:")
		fmt.Println("  • Search patterns: mur search <query>")
		fmt.Println("  • Use hooks that trigger pattern search")
		fmt.Println()
		return nil
	}

	fmt.Printf("Total Events:    %d\n", summary.TotalEvents)
	fmt.Printf("Total Patterns:  %d\n", summary.TotalPatterns)
	fmt.Printf("Search Events:   %d\n", summary.SearchEvents)
	fmt.Printf("Inject Events:   %d\n", summary.InjectEvents)
	fmt.Println()

	if len(summary.TopPatterns) > 0 {
		fmt.Println("🏆 Top Patterns")
		fmt.Println("---------------")
		for i, p := range summary.TopPatterns {
			fmt.Printf("  %d. %-30s  %d hits (%.0f%% search)\n",
				i+1,
				truncateName(p.PatternName, 30),
				p.TotalHits,
				float64(p.SearchCount)/float64(p.TotalHits)*100,
			)
		}
		fmt.Println()
	}

	if summary.ColdPatterns > 0 {
		fmt.Printf("❄️  Cold Patterns: %d (not used in 30 days)\n", summary.ColdPatterns)
		fmt.Println("   Run: mur analytics cold")
		fmt.Println()
	}

	if summary.AvgEffectiveness > 0 {
		fmt.Printf("📈 Avg Effectiveness: %.0f%%\n", summary.AvgEffectiveness*100)
	}

	return nil
}

func runAnalyticsTop(_ *cobra.Command, _ []string) error {
	tracker := getTracker()
	stats, err := tracker.GetTopPatterns(analyticsTop)
	if err != nil {
		return fmt.Errorf("get top patterns: %w", err)
	}

	if analyticsJSON {
		return json.NewEncoder(os.Stdout).Encode(stats)
	}

	fmt.Println("🏆 Top Patterns")
	fmt.Println("===============")
	fmt.Println()

	if len(stats) == 0 {
		fmt.Println("No pattern usage recorded yet.")
		return nil
	}

	// Find max hits for bar chart
	maxHits := 0
	for _, s := range stats {
		if s.TotalHits > maxHits {
			maxHits = s.TotalHits
		}
	}

	for i, s := range stats {
		barLen := 0
		if maxHits > 0 {
			barLen = s.TotalHits * 20 / maxHits
		}
		if barLen == 0 && s.TotalHits > 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)

		lastUsed := "never"
		if !s.LastUsed.IsZero() {
			lastUsed = humanizeTime(s.LastUsed)
		}

		fmt.Printf("%2d. %-28s %s %d\n", i+1, truncateName(s.PatternName, 28), bar, s.TotalHits)
		fmt.Printf("    search: %d | inject: %d | avg score: %.2f | last: %s\n",
			s.SearchCount, s.InjectCount, s.AvgScore, lastUsed)
		fmt.Println()
	}

	return nil
}

func runAnalyticsCold(_ *cobra.Command, _ []string) error {
	tracker := getTracker()
	duration := time.Duration(analyticsDays) * 24 * time.Hour
	stats, err := tracker.GetColdPatterns(duration)
	if err != nil {
		return fmt.Errorf("get cold patterns: %w", err)
	}

	if analyticsJSON {
		return json.NewEncoder(os.Stdout).Encode(stats)
	}

	fmt.Printf("❄️  Patterns Not Used in %d Days\n", analyticsDays)
	fmt.Println("=================================")
	fmt.Println()

	if len(stats) == 0 {
		fmt.Println("All patterns have been used recently! 🎉")
		return nil
	}

	for _, s := range stats {
		lastUsed := "never"
		if !s.LastUsed.IsZero() {
			lastUsed = s.LastUsed.Format("2006-01-02")
		}
		fmt.Printf("  • %-35s last: %s\n", truncateName(s.PatternName, 35), lastUsed)
	}

	fmt.Println()
	fmt.Printf("Consider reviewing these %d patterns:\n", len(stats))
	fmt.Println("  • Archive if no longer relevant")
	fmt.Println("  • Update if outdated")
	fmt.Println("  • Delete if redundant")

	return nil
}

func runAnalyticsFeedback(_ *cobra.Command, args []string) error {
	patternName := args[0]
	feedbackStr := args[1]

	var helpful bool
	switch strings.ToLower(feedbackStr) {
	case "helpful", "yes", "good", "1", "true":
		helpful = true
	case "not-helpful", "no", "bad", "0", "false":
		helpful = false
	default:
		return fmt.Errorf("invalid feedback: %s (use 'helpful' or 'not-helpful')", feedbackStr)
	}

	tracker := getTracker()
	if err := tracker.RecordFeedback("", patternName, helpful); err != nil {
		return fmt.Errorf("record feedback: %w", err)
	}

	if helpful {
		fmt.Printf("✅ Recorded positive feedback for %q\n", patternName)
	} else {
		fmt.Printf("📝 Recorded feedback for %q (not helpful)\n", patternName)
	}

	return nil
}

func truncateName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func humanizeTime(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
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
		return t.Format("Jan 2")
	}
}
