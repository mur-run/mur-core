package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of AI tools",
	Long:  `Check which AI CLI tools are installed and working.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Murmur Health Check")
		fmt.Println("===================")
		fmt.Println("")

		tools := []struct {
			name    string
			binary  string
			verFlag string
		}{
			{"Claude Code", "claude", "--version"},
			{"Gemini CLI", "gemini", "--version"},
			{"Auggie", "auggie", "--version"},
			{"Codex", "codex", "--version"},
			{"OpenCode", "opencode", "--version"},
			{"Aider", "aider", "--version"},
			{"Continue", "continue", "--version"},
			{"Cursor", "cursor", "--version"},
		}

		available := 0
		for _, t := range tools {
			path, err := exec.LookPath(t.binary)
			if err != nil {
				fmt.Printf("❌ %s: not found\n", t.name)
				continue
			}

			// Try to get version
			out, err := exec.Command(path, t.verFlag).Output()
			if err != nil {
				fmt.Printf("✓ %s: found at %s\n", t.name, path)
			} else {
				ver := string(out)
				if len(ver) > 50 {
					ver = ver[:50] + "..."
				}
				fmt.Printf("✓ %s: %s\n", t.name, ver)
			}
			available++
		}

		fmt.Println("")
		fmt.Printf("Available tools: %d/%d\n", available, len(tools))

		if available == 0 {
			fmt.Println("\n⚠️  No AI tools found. Install at least one:")
			fmt.Println("   Claude Code: npm install -g @anthropic-ai/claude-code")
			fmt.Println("   Gemini CLI:  npm install -g @anthropic-ai/gemini-cli")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
