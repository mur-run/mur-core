package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
)

var lifecycleCmd = &cobra.Command{
	Use:   "lifecycle",
	Short: "Manage pattern lifecycle (deprecation, archival)",
	Long: `Manage pattern lifecycle states.

Patterns go through lifecycle stages:
  active     → In use, will be injected
  deprecated → Low effectiveness, warning shown
  archived   → Removed from injection, can be deleted

Examples:
  mur lifecycle evaluate          # Check which patterns should be deprecated
  mur lifecycle apply             # Apply recommended lifecycle changes
  mur lifecycle deprecate <name>  # Manually deprecate a pattern
  mur lifecycle reactivate <name> # Reactivate a deprecated pattern
  mur lifecycle cleanup           # Delete old archived patterns`,
}

var lifecycleEvaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate patterns and recommend lifecycle changes",
	RunE:  lifecycleEvaluateExecute,
}

var lifecycleApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply recommended lifecycle changes",
	RunE:  lifecycleApplyExecute,
}

var lifecycleDeprecateCmd = &cobra.Command{
	Use:   "deprecate <pattern>",
	Short: "Manually deprecate a pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  lifecycleDeprecateExecute,
}

var lifecycleArchiveCmd = &cobra.Command{
	Use:   "archive <pattern>",
	Short: "Archive a pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  lifecycleArchiveExecute,
}

var lifecycleReactivateCmd = &cobra.Command{
	Use:   "reactivate <pattern>",
	Short: "Reactivate a deprecated/archived pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  lifecycleReactivateExecute,
}

var lifecycleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List patterns by lifecycle status",
	RunE:  lifecycleListExecute,
}

var lifecycleCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete old archived patterns",
	RunE:  lifecycleCleanupExecute,
}

func getLifecycleManager() (*pattern.LifecycleManager, error) {
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := pattern.DefaultLifecycleConfig()
	return pattern.NewLifecycleManager(store, cfg), nil
}

func lifecycleEvaluateExecute(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := pattern.DefaultLifecycleConfig()
	cfg.DryRun = dryRun

	mgr := pattern.NewLifecycleManager(store, cfg)

	report, err := mgr.Evaluate()
	if err != nil {
		return err
	}

	fmt.Println("Lifecycle Evaluation")
	fmt.Println("====================")
	fmt.Printf("Evaluated:    %d patterns\n", report.Evaluated)
	fmt.Printf("To Deprecate: %d\n", report.Deprecated)
	fmt.Printf("To Archive:   %d\n", report.Archived)
	fmt.Printf("To Reactivate: %d\n", report.Reactivated)
	fmt.Println()

	if len(report.Actions) == 0 {
		fmt.Println("✓ All patterns are healthy")
		return nil
	}

	fmt.Println("Recommended Actions:")
	for _, a := range report.Actions {
		emoji := map[pattern.ActionType]string{
			pattern.ActionDeprecate:  "⚠️",
			pattern.ActionArchive:    "📦",
			pattern.ActionReactivate: "✨",
		}[a.Action]

		fmt.Printf("%s %s: %s → %s\n", emoji, a.PatternName, a.OldStatus, a.NewStatus)
		fmt.Printf("   Reason: %s\n", a.Reason)
	}

	if dryRun {
		fmt.Println("\n(dry-run mode, no changes made)")
	} else {
		fmt.Println("\nRun 'mur lifecycle apply' to apply these changes")
	}

	return nil
}

func lifecycleApplyExecute(cmd *cobra.Command, args []string) error {
	mgr, err := getLifecycleManager()
	if err != nil {
		return err
	}

	report, err := mgr.EvaluateAndApply()
	if err != nil {
		return err
	}

	if len(report.Actions) == 0 {
		fmt.Println("✓ No changes needed")
		return nil
	}

	fmt.Println("Applied Lifecycle Changes")
	fmt.Println("=========================")

	for _, a := range report.Actions {
		emoji := map[pattern.ActionType]string{
			pattern.ActionDeprecate:  "⚠️",
			pattern.ActionArchive:    "📦",
			pattern.ActionReactivate: "✨",
		}[a.Action]

		fmt.Printf("%s %s: %s\n", emoji, a.PatternName, a.Reason)
	}

	fmt.Printf("\n✓ Applied %d changes\n", len(report.Actions))
	return nil
}

func lifecycleDeprecateExecute(cmd *cobra.Command, args []string) error {
	name := args[0]
	reason, _ := cmd.Flags().GetString("reason")

	if reason == "" {
		reason = "manually deprecated"
	}

	mgr, err := getLifecycleManager()
	if err != nil {
		return err
	}

	if err := mgr.Deprecate(name, reason); err != nil {
		return err
	}

	fmt.Printf("⚠️ Deprecated: %s\n", name)
	fmt.Printf("   Reason: %s\n", reason)
	return nil
}

func lifecycleArchiveExecute(cmd *cobra.Command, args []string) error {
	name := args[0]
	reason, _ := cmd.Flags().GetString("reason")

	if reason == "" {
		reason = "manually archived"
	}

	mgr, err := getLifecycleManager()
	if err != nil {
		return err
	}

	if err := mgr.Archive(name, reason); err != nil {
		return err
	}

	fmt.Printf("📦 Archived: %s\n", name)
	fmt.Printf("   Reason: %s\n", reason)
	return nil
}

func lifecycleReactivateExecute(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getLifecycleManager()
	if err != nil {
		return err
	}

	if err := mgr.Reactivate(name); err != nil {
		return err
	}

	fmt.Printf("✨ Reactivated: %s\n", name)
	return nil
}

func lifecycleListExecute(cmd *cobra.Command, args []string) error {
	status, _ := cmd.Flags().GetString("status")

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	patterns, err := store.List()
	if err != nil {
		return err
	}

	// Count by status
	counts := map[pattern.LifecycleStatus]int{}
	for _, p := range patterns {
		counts[p.Lifecycle.Status]++
	}

	fmt.Println("Pattern Lifecycle Status")
	fmt.Println("========================")
	fmt.Printf("Active:     %d\n", counts[pattern.StatusActive])
	fmt.Printf("Deprecated: %d\n", counts[pattern.StatusDeprecated])
	fmt.Printf("Archived:   %d\n", counts[pattern.StatusArchived])
	fmt.Println()

	// Filter by status if specified
	filterStatus := pattern.LifecycleStatus(status)
	if status != "" && status != "all" {
		fmt.Printf("Patterns with status '%s':\n", status)
		for _, p := range patterns {
			if p.Lifecycle.Status == filterStatus {
				fmt.Printf("  • %s", p.Name)
				if p.Lifecycle.DeprecationReason != "" {
					fmt.Printf(" (%s)", p.Lifecycle.DeprecationReason)
				}
				fmt.Println()
			}
		}
	}

	return nil
}

func lifecycleCleanupExecute(cmd *cobra.Command, args []string) error {
	days, _ := cmd.Flags().GetInt("days")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	cfg := pattern.DefaultLifecycleConfig()
	cfg.DryRun = dryRun

	mgr := pattern.NewLifecycleManager(store, cfg)

	olderThan := time.Duration(days) * 24 * time.Hour

	if dryRun {
		// Count what would be deleted
		archived, _ := mgr.GetArchived()
		cutoff := time.Now().Add(-olderThan)
		count := 0
		for _, p := range archived {
			if p.Lifecycle.Updated.Before(cutoff) {
				count++
				fmt.Printf("Would delete: %s (archived %s)\n", p.Name, p.Lifecycle.Updated.Format("2006-01-02"))
			}
		}
		fmt.Printf("\n(dry-run) Would delete %d patterns\n", count)
		return nil
	}

	deleted, err := mgr.Cleanup(olderThan)
	if err != nil {
		return err
	}

	if deleted == 0 {
		fmt.Println("✓ No patterns to clean up")
	} else {
		fmt.Printf("🗑️ Deleted %d archived patterns older than %d days\n", deleted, days)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(lifecycleCmd)
	lifecycleCmd.AddCommand(lifecycleEvaluateCmd)
	lifecycleCmd.AddCommand(lifecycleApplyCmd)
	lifecycleCmd.AddCommand(lifecycleDeprecateCmd)
	lifecycleCmd.AddCommand(lifecycleArchiveCmd)
	lifecycleCmd.AddCommand(lifecycleReactivateCmd)
	lifecycleCmd.AddCommand(lifecycleListCmd)
	lifecycleCmd.AddCommand(lifecycleCleanupCmd)

	lifecycleEvaluateCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	lifecycleDeprecateCmd.Flags().StringP("reason", "r", "", "Reason for deprecation")
	lifecycleArchiveCmd.Flags().StringP("reason", "r", "", "Reason for archival")
	lifecycleListCmd.Flags().StringP("status", "s", "", "Filter by status (active/deprecated/archived)")
	lifecycleCleanupCmd.Flags().Int("days", 30, "Delete archived patterns older than this many days")
	lifecycleCleanupCmd.Flags().Bool("dry-run", false, "Show what would be deleted")
}
