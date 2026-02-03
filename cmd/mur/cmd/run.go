package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/karajanchang/murmur-ai/internal/config"
	"github.com/karajanchang/murmur-ai/internal/router"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a prompt with an AI tool",
	Long: `Run a prompt using the configured AI tool.

With routing.mode=auto (default), murmur selects the best tool based on
prompt complexity. Simple questions use free tools; complex tasks use paid.

Use -t to override automatic selection.

Examples:
  mur run -p "what is git?"              # Auto-routes to free tool
  mur run -p "refactor this module"      # Auto-routes to paid tool
  mur run -p "explain x" -t claude       # Force specific tool
  mur run -p "test" --explain            # Show routing decision only`,
	RunE: runExecute,
}

func runExecute(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	forceTool, _ := cmd.Flags().GetString("tool")
	explain, _ := cmd.Flags().GetBool("explain")

	if prompt == "" {
		return fmt.Errorf("prompt is required. Use -p \"your prompt\"")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var tool string
	var reason string

	if forceTool != "" {
		// User explicitly chose a tool
		tool = forceTool
		reason = "user specified with -t flag"
	} else {
		// Use router
		selection, err := router.SelectTool(prompt, cfg)
		if err != nil {
			return fmt.Errorf("routing failed: %w", err)
		}
		tool = selection.Tool
		reason = selection.Reason

		if explain {
			// Show decision and exit
			fmt.Println("Routing Decision")
			fmt.Println("================")
			fmt.Printf("Prompt:     %s\n", truncateStr(prompt, 60))
			fmt.Printf("Complexity: %.2f\n", selection.Analysis.Complexity)
			fmt.Printf("Category:   %s\n", selection.Analysis.Category)
			fmt.Printf("Tool Use:   %v\n", selection.Analysis.NeedsToolUse)
			if len(selection.Analysis.Keywords) > 0 {
				fmt.Printf("Keywords:   %v\n", selection.Analysis.Keywords)
			}
			fmt.Println()
			fmt.Printf("Selected:   %s\n", selection.Tool)
			fmt.Printf("Reason:     %s\n", selection.Reason)
			if selection.Fallback != "" {
				fmt.Printf("Fallback:   %s\n", selection.Fallback)
			}
			return nil
		}
	}

	// Validate tool
	if err := cfg.EnsureTool(tool); err != nil {
		return err
	}

	toolCfg, _ := cfg.GetTool(tool)

	// Build command args
	cmdArgs := append(toolCfg.Flags, prompt)

	// Check if binary exists
	binPath, err := exec.LookPath(toolCfg.Binary)
	if err != nil {
		return fmt.Errorf("%s not found in PATH. Install it first", toolCfg.Binary)
	}

	fmt.Printf("→ %s (%s)\n\n", tool, reason)

	// Execute the tool
	execCmd := exec.Command(binPath, cmdArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// truncateStr truncates a string to max length, adding "..." if truncated.
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("prompt", "p", "", "The prompt to run")
	runCmd.Flags().StringP("tool", "t", "", "Force specific tool (overrides routing)")
	runCmd.Flags().Bool("explain", false, "Show routing decision without executing")
}
