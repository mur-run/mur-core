package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/karajanchang/murmur-ai/internal/stats"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Long: `Display usage statistics for murmur-ai.

Shows tool usage counts, estimated costs, routing decisions,
and usage trends over time.

Examples:
  mur stats                    # Show overall statistics
  mur stats --tool claude      # Stats for specific tool
  mur stats --period week      # Stats for last week
  mur stats --json             # Output as JSON
  mur stats reset              # Clear all statistics`,
	RunE: statsExecute,
}

var statsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all statistics",
	Long:  `Remove all recorded usage data.`,
	RunE:  statsResetExecute,
}

func statsExecute(cmd *cobra.Command, args []string) error {
	tool, _ := cmd.Flags().GetString("tool")
	period, _ := cmd.Flags().GetString("period")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Build filter
	filter := stats.QueryFilter{}

	if tool != "" {
		filter.Tool = tool
	}

	// Parse period
	now := time.Now()
	switch period {
	case "day":
		filter.StartTime = now.AddDate(0, 0, -1)
	case "week":
		filter.StartTime = now.AddDate(0, 0, -7)
	case "month":
		filter.StartTime = now.AddDate(0, -1, 0)
	case "year":
		filter.StartTime = now.AddDate(-1, 0, 0)
	case "", "all":
		// No time filter
	default:
		return fmt.Errorf("invalid period: %s (use day/week/month/year/all)", period)
	}

	// Query records
	records, err := stats.Query(filter)
	if err != nil {
		return fmt.Errorf("failed to query stats: %w", err)
	}

	// If filtering by tool, show tool-specific stats
	if tool != "" && !jsonOutput {
		fmt.Print(stats.FormatToolStats(tool, records))
		return nil
	}

	// Compute summary
	summary := stats.Summarize(records)
	summary.Period = period
	if summary.Period == "" {
		summary.Period = "all"
	}

	// Output
	if jsonOutput {
		data, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize stats: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Print(stats.FormatSummary(summary))
	}

	return nil
}

func statsResetExecute(cmd *cobra.Command, args []string) error {
	if err := stats.Reset(); err != nil {
		return fmt.Errorf("failed to reset stats: %w", err)
	}
	fmt.Println("✓ Statistics cleared")
	return nil
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.AddCommand(statsResetCmd)

	statsCmd.Flags().StringP("tool", "t", "", "Filter by tool (claude/gemini/auggie)")
	statsCmd.Flags().StringP("period", "p", "", "Time period (day/week/month/year/all)")
	statsCmd.Flags().Bool("json", false, "Output as JSON")
}
