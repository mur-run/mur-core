package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/karajanchang/murmur-ai/internal/learn"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Manage learned patterns",
	Long:  `List, add, sync, and manage patterns in your knowledge base.`,
}

var learnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		patterns, err := learn.List()
		if err != nil {
			return fmt.Errorf("failed to list patterns: %w", err)
		}

		domain, _ := cmd.Flags().GetString("domain")
		category, _ := cmd.Flags().GetString("category")

		fmt.Println("Learned Patterns")
		fmt.Println("================")
		fmt.Println("")

		count := 0
		for _, p := range patterns {
			// Filter by domain
			if domain != "" && p.Domain != domain {
				continue
			}
			// Filter by category
			if category != "" && p.Category != category {
				continue
			}

			fmt.Printf("  %-20s  [%s/%s]  %.0f%%\n", p.Name, p.Domain, p.Category, p.Confidence*100)
			if p.Description != "" {
				fmt.Printf("    %s\n", truncate(p.Description, 60))
			}
			count++
		}

		fmt.Println("")
		fmt.Printf("Total: %d patterns\n", count)

		return nil
	},
}

var learnAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new pattern",
	Long: `Add a new pattern interactively or from stdin.

Examples:
  mur learn add my-pattern              # Interactive mode
  cat pattern.yaml | mur learn add my-pattern --stdin  # From stdin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		fromStdin, _ := cmd.Flags().GetBool("stdin")

		var p learn.Pattern
		p.Name = name

		if fromStdin {
			// Read from stdin (expect YAML or simple text)
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			p.Content = strings.Join(lines, "\n")
			p.Domain = "general"
			p.Category = "pattern"
			p.Confidence = 0.5
		} else {
			// Interactive mode
			reader := bufio.NewReader(os.Stdin)

			fmt.Printf("Adding pattern: %s\n\n", name)

			fmt.Print("Description: ")
			desc, _ := reader.ReadString('\n')
			p.Description = strings.TrimSpace(desc)

			fmt.Printf("Domain (%s): ", strings.Join(learn.ValidDomains(), ", "))
			domain, _ := reader.ReadString('\n')
			p.Domain = strings.TrimSpace(domain)
			if p.Domain == "" {
				p.Domain = "general"
			}

			fmt.Printf("Category (%s): ", strings.Join(learn.ValidCategories(), ", "))
			category, _ := reader.ReadString('\n')
			p.Category = strings.TrimSpace(category)
			if p.Category == "" {
				p.Category = "pattern"
			}

			fmt.Print("Confidence (0.0-1.0, default 0.5): ")
			confStr, _ := reader.ReadString('\n')
			confStr = strings.TrimSpace(confStr)
			if confStr != "" {
				if conf, err := strconv.ParseFloat(confStr, 64); err == nil {
					p.Confidence = conf
				}
			}
			if p.Confidence == 0 {
				p.Confidence = 0.5
			}

			fmt.Println("Content (end with Ctrl+D or empty line):")
			var contentLines []string
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				line = strings.TrimRight(line, "\n")
				if line == "" && len(contentLines) > 0 {
					break
				}
				contentLines = append(contentLines, line)
			}
			p.Content = strings.Join(contentLines, "\n")
		}

		if err := learn.Add(p); err != nil {
			return fmt.Errorf("failed to add pattern: %w", err)
		}

		fmt.Printf("\n✓ Pattern '%s' added successfully\n", name)
		fmt.Println("  Run 'mur learn sync' to sync to AI tools")

		return nil
	},
}

var learnGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Show a pattern",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		p, err := learn.Get(name)
		if err != nil {
			return err
		}

		fmt.Printf("Name:        %s\n", p.Name)
		fmt.Printf("Description: %s\n", p.Description)
		fmt.Printf("Domain:      %s\n", p.Domain)
		fmt.Printf("Category:    %s\n", p.Category)
		fmt.Printf("Confidence:  %.0f%%\n", p.Confidence*100)
		fmt.Printf("Created:     %s\n", p.CreatedAt)
		fmt.Printf("Updated:     %s\n", p.UpdatedAt)
		fmt.Println("")
		fmt.Println("Content:")
		fmt.Println("--------")
		fmt.Println(p.Content)

		return nil
	},
}

var learnDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a pattern",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete pattern '%s'? [y/N] ", name)
			reader := bufio.NewReader(os.Stdin)
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm != "y" && confirm != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		if err := learn.Delete(name); err != nil {
			return err
		}

		fmt.Printf("✓ Pattern '%s' deleted\n", name)
		fmt.Println("  Run 'mur learn sync' to update AI tools")

		return nil
	},
}

var learnSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync patterns to AI tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing patterns to AI tools...")
		fmt.Println("")

		results, err := learn.SyncPatterns()
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

		// Cleanup orphaned patterns
		cleanup, _ := cmd.Flags().GetBool("cleanup")
		if cleanup {
			fmt.Println("")
			fmt.Println("Cleaning up orphaned patterns...")
			if err := learn.CleanupSyncedPatterns(); err != nil {
				fmt.Printf("  ⚠ cleanup failed: %v\n", err)
			} else {
				fmt.Println("  ✓ cleanup complete")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(learnCmd)
	learnCmd.AddCommand(learnListCmd)
	learnCmd.AddCommand(learnAddCmd)
	learnCmd.AddCommand(learnGetCmd)
	learnCmd.AddCommand(learnDeleteCmd)
	learnCmd.AddCommand(learnSyncCmd)

	learnListCmd.Flags().StringP("domain", "d", "", "Filter by domain")
	learnListCmd.Flags().StringP("category", "c", "", "Filter by category")

	learnAddCmd.Flags().Bool("stdin", false, "Read content from stdin")

	learnDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	learnSyncCmd.Flags().Bool("cleanup", false, "Remove orphaned synced patterns")
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
