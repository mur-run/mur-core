package cmd

import (
	"fmt"

	"github.com/mur-run/mur-cli/internal/core/pattern"
	"github.com/spf13/cobra"
)

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Recalculate content hashes for all patterns",
	Long:  `Rehash recalculates the SHA256 content hash for all patterns. Use this after migration or manual edits.`,
	RunE:  runRehash,
}

func init() {
	rootCmd.AddCommand(rehashCmd)
}

func runRehash(cmd *cobra.Command, args []string) error {
	store, err := pattern.DefaultStore()
	if err != nil {
		return err
	}

	patterns, err := store.List()
	if err != nil {
		return err
	}

	if len(patterns) == 0 {
		fmt.Println("No patterns found")
		return nil
	}

	updated := 0
	for _, p := range patterns {
		oldHash := p.Security.Hash
		p.UpdateHash()

		if oldHash != p.Security.Hash {
			if err := store.Update(&p); err != nil {
				fmt.Printf("❌ Failed to update %s: %v\n", p.Name, err)
				continue
			}
			updated++
			fmt.Printf("✅ %s: hash updated\n", p.Name)
		} else {
			fmt.Printf("⏭️  %s: hash unchanged\n", p.Name)
		}
	}

	fmt.Printf("\n📊 Updated %d of %d patterns\n", updated, len(patterns))
	return nil
}
