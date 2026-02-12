package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/analytics"
	"github.com/spf13/cobra"
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback [pattern-name] [rating]",
	Short: "Provide feedback on a pattern",
	Long: `Record feedback on how helpful a pattern was.

Ratings:
  helpful      - The pattern was useful
  not_helpful  - The pattern wasn't useful
  skip         - Skip rating this pattern

Without arguments, shows an interactive selection of recently used patterns.`,
	Example: `  mur feedback                              # Interactive mode
  mur feedback swift-testing helpful        # Quick feedback
  mur feedback go-error-handling not_helpful`,
	Args: cobra.MaximumNArgs(2),
	RunE: runFeedback,
}

func init() {
	rootCmd.AddCommand(feedbackCmd)
}

func runFeedback(cmd *cobra.Command, args []string) error {
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

	var patternID, patternName, rating string

	if len(args) >= 1 {
		// Pattern provided as argument
		patternName = args[0]
		patternID, err = findPatternID(store, patternName)
		if err != nil {
			return err
		}

		if len(args) >= 2 {
			// Rating also provided
			rating = normalizeRating(args[1])
			if rating == "" {
				return fmt.Errorf("invalid rating '%s'. Use: helpful, not_helpful, or skip", args[1])
			}
		} else {
			// Prompt for rating
			rating, err = promptRating()
			if err != nil {
				return err
			}
		}
	} else {
		// Interactive mode - select from recent patterns
		patternID, patternName, err = selectRecentPattern(store)
		if err != nil {
			return err
		}

		rating, err = promptRating()
		if err != nil {
			return err
		}
	}

	// Record the feedback
	event := analytics.FeedbackEvent{
		PatternID: patternID,
		Rating:    rating,
	}

	if err := store.RecordFeedback(event); err != nil {
		return fmt.Errorf("failed to record feedback: %w", err)
	}

	emoji := map[string]string{
		"helpful":     "👍",
		"not_helpful": "👎",
		"skip":        "⏭️",
	}

	fmt.Printf("✓ %s Feedback recorded for '%s'\n", emoji[rating], patternName)
	return nil
}

func findPatternID(store *analytics.Store, nameOrID string) (string, error) {
	allStats, err := store.GetAllStats(1000)
	if err != nil {
		return "", err
	}

	// Try exact match first
	for _, s := range allStats {
		if s.PatternID == nameOrID || strings.EqualFold(s.PatternName, nameOrID) {
			return s.PatternID, nil
		}
	}

	// Try partial match
	for _, s := range allStats {
		if strings.Contains(strings.ToLower(s.PatternName), strings.ToLower(nameOrID)) {
			return s.PatternID, nil
		}
	}

	return "", fmt.Errorf("pattern '%s' not found in usage history", nameOrID)
}

func selectRecentPattern(store *analytics.Store) (string, string, error) {
	recent, err := store.GetRecentUsage(10)
	if err != nil {
		return "", "", err
	}

	if len(recent) == 0 {
		return "", "", fmt.Errorf("no patterns in usage history yet")
	}

	fmt.Println("\n📋 Recently Used Patterns:")
	fmt.Println()

	// Deduplicate by pattern ID
	seen := make(map[string]bool)
	var unique []*analytics.UsageEvent
	for _, e := range recent {
		if !seen[e.PatternID] {
			seen[e.PatternID] = true
			unique = append(unique, e)
		}
	}

	for i, e := range unique {
		timeAgo := formatTimeAgo(e.InjectedAt)
		fmt.Printf("  %d. %s (%s)\n", i+1, e.PatternName, timeAgo)
	}

	fmt.Println()
	fmt.Print("Select pattern (number): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	input = strings.TrimSpace(input)
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return "", "", fmt.Errorf("invalid selection")
	}

	if selection < 1 || selection > len(unique) {
		return "", "", fmt.Errorf("selection out of range")
	}

	selected := unique[selection-1]
	return selected.PatternID, selected.PatternName, nil
}

func promptRating() (string, error) {
	fmt.Println()
	fmt.Println("Was this pattern helpful?")
	fmt.Println("  1. 👍 Helpful")
	fmt.Println("  2. 👎 Not helpful")
	fmt.Println("  3. ⏭️  Skip")
	fmt.Println()
	fmt.Print("Select (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	switch input {
	case "1", "helpful", "h":
		return "helpful", nil
	case "2", "not_helpful", "n", "not helpful":
		return "not_helpful", nil
	case "3", "skip", "s":
		return "skip", nil
	default:
		return "", fmt.Errorf("invalid selection '%s'", input)
	}
}

func normalizeRating(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	switch input {
	case "helpful", "h", "good", "yes", "y", "1", "👍":
		return "helpful"
	case "not_helpful", "not-helpful", "nothelpful", "bad", "no", "n", "2", "👎":
		return "not_helpful"
	case "skip", "s", "3", "⏭️":
		return "skip"
	default:
		return ""
	}
}
