package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/learn"
	"github.com/mur-run/mur-core/internal/stats"
	"github.com/mur-run/mur-core/internal/sync"
)

// OutputStatus represents the full status output
type OutputStatus struct {
	Version   string          `json:"version"`
	Timestamp string          `json:"timestamp"`
	Stats     *stats.Summary  `json:"stats,omitempty"`
	Patterns  []learn.Pattern `json:"patterns,omitempty"`
}

// OutputSyncResult represents sync operation results
type OutputSyncResult struct {
	Success bool                `json:"success"`
	Results map[string][]Result `json:"results"`
	Error   string              `json:"error,omitempty"`
}

// Result is a simplified sync result for JSON output
type Result struct {
	Target  string `json:"target"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var outputCmd = &cobra.Command{
	Use:   "output",
	Short: "Output data in JSON format for programmatic use",
	Long: `Output murmur-ai data in JSON format for integration with other tools.

Examples:
  mur output --json           # Full status (stats + patterns)
  mur output sync --json      # Run sync and output results
  mur output stats --json     # Statistics only
  mur output learn --json     # Patterns only`,
	RunE: outputExecute,
}

var outputSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run sync and output results as JSON",
	RunE:  outputSyncExecute,
}

var outputStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Output statistics as JSON",
	RunE:  outputStatsExecute,
}

var outputLearnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Output patterns as JSON",
	RunE:  outputLearnExecute,
}

func outputExecute(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if !jsonOutput {
		fmt.Println("Use --json flag for JSON output")
		fmt.Println()
		fmt.Println("Available subcommands:")
		fmt.Println("  mur output --json         Full status")
		fmt.Println("  mur output sync --json    Sync results")
		fmt.Println("  mur output stats --json   Statistics")
		fmt.Println("  mur output learn --json   Patterns")
		return nil
	}

	// Build full status
	status := OutputStatus{
		Version:   Version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Get stats
	records, err := stats.Query(stats.QueryFilter{})
	if err == nil {
		summary := stats.Summarize(records)
		status.Stats = &summary
	}

	// Get patterns
	patterns, err := learn.List()
	if err == nil {
		status.Patterns = patterns
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize output: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputSyncExecute(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if !jsonOutput {
		// Fall back to regular sync
		return runSync(cmd, args)
	}

	result := OutputSyncResult{
		Success: true,
		Results: make(map[string][]Result),
	}

	// Run sync
	syncResults, err := sync.SyncAll()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		for category, results := range syncResults {
			for _, r := range results {
				result.Results[category] = append(result.Results[category], Result{
					Target:  r.Target,
					Success: r.Success,
					Message: r.Message,
				})
				if !r.Success {
					result.Success = false
				}
			}
		}

		// Also sync patterns
		patternResults, err := learn.SyncPatterns()
		if err != nil {
			result.Results["patterns"] = []Result{{
				Target:  "patterns",
				Success: false,
				Message: err.Error(),
			}}
		} else {
			for _, r := range patternResults {
				result.Results["patterns"] = append(result.Results["patterns"], Result{
					Target:  r.Target,
					Success: r.Success,
					Message: r.Message,
				})
			}
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize output: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputStatsExecute(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if !jsonOutput {
		fmt.Println("Use: mur output stats --json")
		return nil
	}

	records, err := stats.Query(stats.QueryFilter{})
	if err != nil {
		return fmt.Errorf("failed to query stats: %w", err)
	}

	summary := stats.Summarize(records)

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize stats: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputLearnExecute(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if !jsonOutput {
		fmt.Println("Use: mur output learn --json")
		return nil
	}

	patterns, err := learn.List()
	if err != nil {
		return fmt.Errorf("failed to list patterns: %w", err)
	}

	output := struct {
		Count    int             `json:"count"`
		Patterns []learn.Pattern `json:"patterns"`
	}{
		Count:    len(patterns),
		Patterns: patterns,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize patterns: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func init() {
	outputCmd.Hidden = true
	rootCmd.AddCommand(outputCmd)
	outputCmd.AddCommand(outputSyncCmd)
	outputCmd.AddCommand(outputStatsCmd)
	outputCmd.AddCommand(outputLearnCmd)

	outputCmd.Flags().Bool("json", false, "Output as JSON")
	outputSyncCmd.Flags().Bool("json", false, "Output as JSON")
	outputStatsCmd.Flags().Bool("json", false, "Output as JSON")
	outputLearnCmd.Flags().Bool("json", false, "Output as JSON")
}
