package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/core/audit"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View pattern audit log",
	Long: `Show recent audit log entries for pattern operations.

Examples:
  mur audit                        # Show recent entries
  mur audit --pattern my-pattern   # Filter by pattern name
  mur audit --limit 50             # Show last 50 entries`,
	RunE: func(cmd *cobra.Command, args []string) error {
		patternFilter, _ := cmd.Flags().GetString("pattern")
		limit, _ := cmd.Flags().GetInt("limit")

		logger, err := audit.DefaultLogger()
		if err != nil {
			return fmt.Errorf("cannot open audit log: %w", err)
		}

		entries, err := logger.ReadFiltered(patternFilter)
		if err != nil {
			return fmt.Errorf("cannot read audit log: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No audit entries found.")
			return nil
		}

		fmt.Println("Audit Log")
		fmt.Println("=========")
		fmt.Println("")

		shown := 0
		for _, e := range entries {
			if limit > 0 && shown >= limit {
				break
			}

			fmt.Printf("  %s  %-8s  %-25s", e.Timestamp.Format("2006-01-02 15:04:05"), e.Action, e.PatternName)
			if e.ToolTarget != "" {
				fmt.Printf("  → %s", e.ToolTarget)
			}
			if e.Source != "" {
				fmt.Printf("  [%s]", e.Source)
			}
			if e.Details != "" {
				fmt.Printf("  %s", e.Details)
			}
			fmt.Println("")
			shown++
		}

		fmt.Println("")
		fmt.Printf("Showing %d of %d entries\n", shown, len(entries))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().StringP("pattern", "p", "", "Filter by pattern name")
	auditCmd.Flags().IntP("limit", "l", 25, "Maximum entries to show")
}
