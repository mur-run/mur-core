package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Give feedback on pattern effectiveness",
	Long: `Give feedback on pattern effectiveness to improve future injections.

Use after running mur run to rate how helpful the injected patterns were.

Ratings:
  helpful   (+1) - The pattern helped complete the task
  neutral   (0)  - The pattern didn't make a difference  
  unhelpful (-1) - The pattern was distracting or wrong

Examples:
  mur feedback helpful swift-error-handling
  mur feedback unhelpful debugging-tips --comment "too generic"
  mur feedback neutral testing-patterns`,
	Args: cobra.ExactArgs(2),
	RunE: feedbackExecute,
}

func feedbackExecute(cmd *cobra.Command, args []string) error {
	ratingStr := args[0]
	patternName := args[1]
	comment, _ := cmd.Flags().GetString("comment")

	// Parse rating
	var rating int
	switch ratingStr {
	case "helpful", "+1", "1":
		rating = 1
	case "neutral", "0":
		rating = 0
	case "unhelpful", "-1":
		rating = -1
	default:
		return fmt.Errorf("invalid rating: %s (use helpful, neutral, or unhelpful)", ratingStr)
	}

	// Create tracker
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".murmur", "patterns")
	trackingDir := filepath.Join(home, ".murmur", "tracking")
	tracker := inject.NewTracker(pattern.NewStore(patternsDir), trackingDir)

	// Record feedback
	if err := tracker.RecordFeedback(patternName, rating, comment); err != nil {
		return fmt.Errorf("failed to record feedback: %w", err)
	}

	// Update pattern effectiveness
	if err := tracker.UpdatePatternEffectiveness(patternName); err != nil {
		// Non-fatal warning
		fmt.Fprintf(os.Stderr, "⚠ Could not update effectiveness: %v\n", err)
	}

	// Show confirmation
	ratingEmoji := map[int]string{1: "👍", 0: "😐", -1: "👎"}[rating]
	fmt.Printf("%s Recorded feedback for '%s'\n", ratingEmoji, patternName)

	// Show updated stats
	stats, err := tracker.GetPatternStats(patternName)
	if err == nil && stats.TotalUses > 0 {
		fmt.Printf("   Uses: %d | Effectiveness: %.0f%%\n", stats.TotalUses, stats.Effectiveness*100)
	}

	return nil
}

var patternStatsCmd = &cobra.Command{
	Use:   "pattern-stats",
	Short: "Show pattern effectiveness statistics",
	Long: `Show effectiveness statistics for all patterns based on usage tracking.

Statistics include:
  - Total uses
  - Success rate
  - User feedback scores
  - Computed effectiveness

Examples:
  mur pattern-stats
  mur pattern-stats --update   # Update all pattern effectiveness scores`,
	RunE: patternStatsExecute,
}

func patternStatsExecute(cmd *cobra.Command, args []string) error {
	update, _ := cmd.Flags().GetBool("update")

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".murmur", "patterns")
	trackingDir := filepath.Join(home, ".murmur", "tracking")
	tracker := inject.NewTracker(pattern.NewStore(patternsDir), trackingDir)

	if update {
		fmt.Println("Updating pattern effectiveness scores...")
		if err := tracker.UpdateAllEffectiveness(); err != nil {
			return fmt.Errorf("failed to update effectiveness: %w", err)
		}
		fmt.Println("✓ Updated")
		fmt.Println()
	}

	allStats, err := tracker.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	if len(allStats) == 0 {
		fmt.Println("No usage data yet. Run `mur run` to start tracking.")
		return nil
	}

	fmt.Println("Pattern Effectiveness")
	fmt.Println("=====================")
	fmt.Println()

	for _, s := range allStats {
		effectBar := makeBar(s.Effectiveness, 10)
		fmt.Printf("%-30s %s %.0f%%\n", s.PatternName, effectBar, s.Effectiveness*100)
		fmt.Printf("  Uses: %-4d Success: %.0f%% | 👍 %d 😐 %d 👎 %d\n",
			s.TotalUses, s.SuccessRate*100,
			s.HelpfulCount, s.NeutralCount, s.UnhelpfulCount)
		fmt.Println()
	}

	return nil
}

func makeBar(value float64, width int) string {
	filled := int(value * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func init() {
	rootCmd.AddCommand(feedbackCmd)
	feedbackCmd.Flags().StringP("comment", "c", "", "Add a comment to the feedback")

	rootCmd.AddCommand(patternStatsCmd)
	patternStatsCmd.Flags().Bool("update", false, "Update pattern effectiveness scores from tracking data")
}
