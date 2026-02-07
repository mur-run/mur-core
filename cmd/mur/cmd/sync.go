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
	syncPush  bool
	syncQuiet bool
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
  mur sync --push   # Also push local changes to remote
  mur sync --quiet  # Silent mode (for hooks)`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncPush, "push", false, "Push local changes to remote repo")
	syncCmd.Flags().BoolVar(&syncQuiet, "quiet", false, "Silent mode (minimal output)")
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
		if !syncQuiet {
			fmt.Println("Pulling from remote...")
		}
		pullCmd := exec.Command("git", "-C", patternsDir, "pull", "--rebase")
		if !syncQuiet {
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
		}
		if err := pullCmd.Run(); err != nil {
			if !syncQuiet {
				fmt.Printf("  ⚠ Pull failed (continuing with local patterns): %v\n", err)
			}
		} else if !syncQuiet {
			fmt.Println("  ✓ Pulled latest patterns")
		}
		if !syncQuiet {
			fmt.Println()
		}
	}

	// Sync patterns to all CLIs
	if !syncQuiet {
		fmt.Println("Syncing patterns to CLIs...")
	}
	results, err := sync.SyncPatternsToAllCLIs()
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if !syncQuiet {
		for _, r := range results {
			status := "✓"
			if !r.Success {
				status = "✗"
			}
			fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
		}
	}

	// Push if requested
	if syncPush && hasRepo {
		if !syncQuiet {
			fmt.Println()
			fmt.Println("Pushing to remote...")
		}

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
		if !syncQuiet {
			pushCmd.Stdout = os.Stdout
			pushCmd.Stderr = os.Stderr
		}
		if err := pushCmd.Run(); err != nil {
			if !syncQuiet {
				fmt.Printf("  ⚠ Push failed: %v\n", err)
			}
		} else if !syncQuiet {
			fmt.Println("  ✓ Pushed to remote")
		}
	}

	if !syncQuiet {
		fmt.Println()
		fmt.Println("Done.")
	}

	return nil
}
