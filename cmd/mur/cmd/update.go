package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var (
	updateSkipSync bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update mur and sync patterns",
	Long: `Update mur to the latest version and sync patterns to all CLIs.

This command:
  1. Updates mur binary to the latest version
  2. Syncs patterns to all CLI tools

Use --skip-sync to only update without syncing.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateSkipSync, "skip-sync", false, "Skip pattern sync after update")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Updating mur...")
	fmt.Println()

	// Update mur binary
	updateCmd := exec.Command("go", "install", "github.com/mur-run/mur-core/cmd/mur@latest")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update mur: %w", err)
	}
	fmt.Println("✓ mur updated to latest version")

	// Sync patterns unless skipped
	if !updateSkipSync {
		fmt.Println()
		fmt.Println("Syncing patterns to all CLIs...")

		results, err := sync.SyncPatternsToAllCLIs()
		if err != nil {
			fmt.Printf("⚠ Warning: sync failed: %v\n", err)
		} else {
			for _, r := range results {
				if r.Success {
					fmt.Printf("  ✓ %s: %s\n", r.Target, r.Message)
				} else {
					fmt.Printf("  ✗ %s: %s\n", r.Target, r.Message)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ Update complete!")

	return nil
}
