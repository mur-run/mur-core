package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/karajanchang/murmur-ai/internal/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a prompt with an AI tool",
	Long: `Run a prompt using the configured AI tool.
	
By default, uses the tool specified in config. Use -t to override.

Examples:
  mur run -p "explain this code"
  mur run -t gemini -p "write a haiku"
  mur run -t claude -p "refactor this function"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		tool, _ := cmd.Flags().GetString("tool")

		if prompt == "" {
			return fmt.Errorf("prompt is required. Use -p \"your prompt\"")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use default tool from config if not specified
		if tool == "" {
			tool = cfg.GetDefaultTool()
		}

		// Get tool config
		toolCfg, ok := cfg.GetTool(tool)
		if !ok {
			return fmt.Errorf("unknown tool: %s. Check ~/.murmur/config.yaml", tool)
		}

		if !toolCfg.Enabled {
			return fmt.Errorf("tool %s is disabled. Enable it in ~/.murmur/config.yaml", tool)
		}

		// Build command args
		cmdArgs := append(toolCfg.Flags, prompt)

		// Check if binary exists
		binPath, err := exec.LookPath(toolCfg.Binary)
		if err != nil {
			return fmt.Errorf("%s not found in PATH. Install it first", toolCfg.Binary)
		}

		fmt.Printf("Running with %s...\n\n", tool)

		// Execute the tool
		execCmd := exec.Command(binPath, cmdArgs...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		return execCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("prompt", "p", "", "The prompt to run")
	runCmd.Flags().StringP("tool", "t", "", "AI tool to use (claude, gemini, auggie)")
}
