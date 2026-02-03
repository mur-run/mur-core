package cmd

import (
	"fmt"
	"os"
	"os/exec"

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

		if tool == "" {
			tool = "claude" // TODO: read from config
		}

		// Map tool to binary
		binaries := map[string]struct {
			bin  string
			args []string
		}{
			"claude": {bin: "claude", args: []string{"-p", prompt}},
			"gemini": {bin: "gemini", args: []string{"-p", prompt}},
			"auggie": {bin: "auggie", args: []string{prompt}},
		}

		toolConfig, ok := binaries[tool]
		if !ok {
			return fmt.Errorf("unknown tool: %s. Available: claude, gemini, auggie", tool)
		}

		// Check if binary exists
		binPath, err := exec.LookPath(toolConfig.bin)
		if err != nil {
			return fmt.Errorf("%s not found in PATH. Install it first", toolConfig.bin)
		}

		fmt.Printf("Running with %s...\n\n", tool)

		// Execute the tool
		execCmd := exec.Command(binPath, toolConfig.args...)
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
