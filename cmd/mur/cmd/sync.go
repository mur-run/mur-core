package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/learn"
	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync configuration to AI tools",
	Long: `Sync MCP servers, hooks, and skills configuration to all AI CLI tools.

This ensures that Claude Code, Gemini CLI, and Auggie all have the same:
  • MCP server configurations
  • Hook configurations
  • Learned patterns`,
}

var syncAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Sync everything (MCP, hooks, skills)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing all configurations...")
		fmt.Println()

		results, err := sync.SyncAll()
		if err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		// Print MCP results
		if mcpResults, ok := results["mcp"]; ok {
			fmt.Println("MCP servers:")
			for _, r := range mcpResults {
				status := "✓"
				if !r.Success {
					status = "✗"
				}
				fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
			}
		}

		// Print hooks results
		if hooksResults, ok := results["hooks"]; ok {
			fmt.Println()
			fmt.Println("Hooks:")
			for _, r := range hooksResults {
				status := "✓"
				if !r.Success {
					status = "✗"
				}
				fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
			}
		}

		// Sync patterns (using Schema v2)
		fmt.Println()
		fmt.Println("Patterns (v2):")
		patternResults, err := learn.SyncPatternsV2()
		if err != nil {
			fmt.Printf("  ⚠ %v\n", err)
		} else {
			for _, r := range patternResults {
				status := "✓"
				if !r.Success {
					status = "✗"
				}
				fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
			}
		}

		// Print skills results
		if skillsResults, ok := results["skills"]; ok {
			fmt.Println()
			fmt.Println("Skills:")
			for _, r := range skillsResults {
				status := "✓"
				if !r.Success {
					status = "✗"
				}
				fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
			}
		}

		fmt.Println()
		fmt.Println("Done.")
		return nil
	},
}

var syncMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Sync MCP server configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing MCP configuration...")
		fmt.Println()

		results, err := sync.SyncMCP()
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

		fmt.Println()
		fmt.Println("Done.")
		return nil
	},
}

var syncHooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Sync hooks configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing hooks configuration...")
		fmt.Println()

		results, err := sync.SyncHooks()
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

		fmt.Println()
		fmt.Println("Done.")
		return nil
	},
}

var syncSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Sync skills to all CLI tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing skills...")
		fmt.Println()

		results, err := sync.SyncSkills()
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

		fmt.Println()
		fmt.Println("Done.")
		return nil
	},
}

var syncPatternsCmd = &cobra.Command{
	Use:   "patterns",
	Short: "Sync patterns to all CLI tools",
	Long: `Sync learned patterns to all AI CLI tools.

This generates a skill file from your patterns and copies it to:
  • ~/.claude/skills/mur-patterns.md
  • ~/.gemini/skills/mur-patterns.md
  • ~/.codex/instructions.md
  • ~/.augment/skills/mur-patterns.md
  • ~/.aider/conventions.md

After syncing, your patterns are available in any CLI you use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing patterns to all CLIs...")
		fmt.Println()

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

		fmt.Println()
		fmt.Println("Done. Your patterns are now available in all CLIs.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncAllCmd)
	syncCmd.AddCommand(syncMcpCmd)
	syncCmd.AddCommand(syncHooksCmd)
	syncCmd.AddCommand(syncSkillsCmd)
	syncCmd.AddCommand(syncPatternsCmd)
}
