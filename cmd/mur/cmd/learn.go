package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Manage learned patterns",
	Long:  `List, add, and manage patterns in your knowledge base.`,
}

var learnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		learnedDir := filepath.Join(home, ".murmur", "learned")

		domain, _ := cmd.Flags().GetString("domain")

		// Also check the murmur-ai skill directory
		skillDir := filepath.Join(home, "clawd", "skills", "murmur-ai", "learned")

		dirs := []string{learnedDir, skillDir}
		count := 0

		fmt.Println("Learned Patterns")
		fmt.Println("================")
		fmt.Println("")

		for _, dir := range dirs {
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(path, ".md") {
					rel, _ := filepath.Rel(dir, path)
					
					// Filter by domain if specified
					if domain != "" && !strings.Contains(rel, domain) {
						return nil
					}

					fmt.Printf("  %s\n", rel)
					count++
				}
				return nil
			})
			if err != nil {
				continue
			}
		}

		fmt.Println("")
		fmt.Printf("Total: %d patterns\n", count)

		return nil
	},
}

var learnAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new pattern (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚠️  Interactive pattern creation not implemented yet.")
		fmt.Println("")
		fmt.Println("Create a pattern manually:")
		fmt.Println("  1. Create a file: ~/.murmur/learned/[domain]/pattern/[name].md")
		fmt.Println("  2. Use this template:")
		fmt.Println("")
		fmt.Println("---")
		fmt.Println("name: my-pattern")
		fmt.Println("confidence: MEDIUM")
		fmt.Println("category: pattern")
		fmt.Println("domain: dev")
		fmt.Println("---")
		fmt.Println("")
		fmt.Println("# Pattern Title")
		fmt.Println("")
		fmt.Println("## Problem / Trigger")
		fmt.Println("When you encounter...")
		fmt.Println("")
		fmt.Println("## Solution")
		fmt.Println("Do this...")
		return nil
	},
}

var learnSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync patterns to AI tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing patterns to AI tools...")
		fmt.Println("")

		// TODO: implement actual sync logic
		// This would call the sync_to_claude_code.sh equivalent

		tools := []string{"Claude Code", "Gemini CLI", "Auggie"}
		for _, t := range tools {
			fmt.Printf("  → %s: ⚠️ sync not implemented\n", t)
		}

		fmt.Println("")
		fmt.Println("Run the bash script for now:")
		fmt.Println("  ~/clawd/skills/murmur-ai/scripts/sync_to_claude_code.sh")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(learnCmd)
	learnCmd.AddCommand(learnListCmd)
	learnCmd.AddCommand(learnAddCmd)
	learnCmd.AddCommand(learnSyncCmd)

	learnListCmd.Flags().StringP("domain", "d", "", "Filter by domain (dev, devops, business, etc.)")
}
