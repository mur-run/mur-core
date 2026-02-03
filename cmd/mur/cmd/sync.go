package cmd

import (
	"fmt"

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
		fmt.Println("")

		steps := []struct {
			name   string
			status string
		}{
			{"MCP servers", "⚠️ not implemented"},
			{"Hooks", "⚠️ not implemented"},
			{"Skills/Patterns", "⚠️ not implemented"},
		}

		for _, s := range steps {
			fmt.Printf("  → %s: %s\n", s.name, s.status)
		}

		fmt.Println("")
		fmt.Println("Run the bash scripts for now:")
		fmt.Println("  ~/clawd/skills/murmur-ai/scripts/mcp_sync.sh")
		fmt.Println("  ~/clawd/skills/murmur-ai/scripts/sync_to_claude_code.sh")

		return nil
	},
}

var syncMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Sync MCP server configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing MCP configuration...")
		fmt.Println("")
		fmt.Println("⚠️ Not implemented yet.")
		fmt.Println("")
		fmt.Println("Run the bash script for now:")
		fmt.Println("  ~/clawd/skills/murmur-ai/scripts/mcp_sync.sh")
		return nil
	},
}

var syncHooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Sync hooks configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing hooks configuration...")
		fmt.Println("")
		fmt.Println("⚠️ Not implemented yet. This will be hooks_sync.sh in Go.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncAllCmd)
	syncCmd.AddCommand(syncMcpCmd)
	syncCmd.AddCommand(syncHooksCmd)
}
