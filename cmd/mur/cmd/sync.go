package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var (
	syncPush bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync patterns to all AI CLIs",
	Long: `Sync learned patterns to all AI CLI tools.

This command:
  1. Pulls latest patterns from remote repo (if configured)
  2. Syncs patterns to all CLI skill directories

Examples:
  mur sync          # Pull + sync to CLIs
  mur sync --push   # Also push local changes to remote`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncPush, "push", false, "Push local changes to remote repo")
}

func runSync(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	gitDir := filepath.Join(patternsDir, ".git")

	// Check if we have a git repo configured
	hasRepo := false
	if _, err := os.Stat(gitDir); err == nil {
		hasRepo = true
	}

	// Pull from remote if repo is configured
	if hasRepo {
		fmt.Println("Pulling from remote...")
		pullCmd := exec.Command("git", "-C", patternsDir, "pull", "--rebase")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			fmt.Printf("  ⚠ Pull failed (continuing with local patterns): %v\n", err)
		} else {
			fmt.Println("  ✓ Pulled latest patterns")
		}
		fmt.Println()
	}

	// Sync patterns to all CLIs
	fmt.Println("Syncing patterns to CLIs...")
	results, err := sync.SyncPatternsToAllCLIs()
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	for _, r := range results {
		status := "✓"
		if !r.Success {
			status = "✗"
		}
		fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
	}

	// Push if requested
	if syncPush && hasRepo {
		fmt.Println()
		fmt.Println("Pushing to remote...")

		// Add all changes
		addCmd := exec.Command("git", "-C", patternsDir, "add", "-A")
		addCmd.Run()

		// Check if there are changes to commit
		diffCmd := exec.Command("git", "-C", patternsDir, "diff", "--cached", "--quiet")
		if diffCmd.Run() != nil {
			// There are changes, commit them
			commitCmd := exec.Command("git", "-C", patternsDir, "commit", "-m", "mur: sync patterns")
			commitCmd.Run()
		}

		// Push
		pushCmd := exec.Command("git", "-C", patternsDir, "push")
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			fmt.Printf("  ⚠ Push failed: %v\n", err)
		} else {
			fmt.Println("  ✓ Pushed to remote")
		}
	}

	fmt.Println()
	fmt.Println("Done.")

	return nil
}
