package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/core/embed"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

var contextCmd = &cobra.Command{
	Use:    "context",
	Short:  "Output relevant patterns for current context",
	Hidden: true, // Used by hooks, not meant for direct use
	Long: `Output patterns relevant to the current project context.

This command is designed to be used by Claude Code hooks to inject
context-aware patterns into prompts.

Examples:
  mur context                    # Detect context from cwd
  mur context --prompt "fix bug" # Also consider prompt
  mur context --max 3            # Limit to 3 patterns`,
	RunE: runContext,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.Flags().StringP("prompt", "p", "", "Prompt to consider for matching")
	contextCmd.Flags().Int("max", 5, "Maximum patterns to output")
	contextCmd.Flags().Bool("compact", false, "Compact output (names only)")
}

func runContext(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	maxPatterns, _ := cmd.Flags().GetInt("max")
	compact, _ := cmd.Flags().GetBool("compact")

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Initialize pattern store
	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	// Check if we have any patterns
	patterns, err := store.List()
	if err != nil || len(patterns) == 0 {
		// No patterns, output nothing
		return nil
	}

	// Create injector
	injector := inject.NewInjector(store)

	// Try to enable semantic search
	embedCfg := embed.DefaultConfig()
	_ = injector.WithSemanticSearch(embedCfg) // Non-fatal if fails

	// Get context-aware patterns
	// Use empty prompt if not provided - we'll match based on project context
	queryPrompt := prompt
	if queryPrompt == "" {
		queryPrompt = "general development task"
	}

	result, err := injector.Inject(queryPrompt, workDir)
	if err != nil {
		return nil // Silent fail, don't break the hook
	}

	if len(result.Patterns) == 0 {
		return nil
	}

	// Limit patterns
	if len(result.Patterns) > maxPatterns {
		result.Patterns = result.Patterns[:maxPatterns]
	}

	if compact {
		// Just output pattern names
		var names []string
		for _, p := range result.Patterns {
			names = append(names, p.Name)
		}
		fmt.Println("[mur] Relevant patterns:", strings.Join(names, ", "))
		return nil
	}

	// Output patterns in a format suitable for prompt injection
	fmt.Println()
	fmt.Println("─── Relevant Patterns (mur) ───")
	if result.Context != nil && result.Context.ProjectType != "" {
		fmt.Printf("Project: %s (%s)\n", result.Context.ProjectName, result.Context.ProjectType)
	}
	fmt.Println()

	for _, p := range result.Patterns {
		fmt.Printf("## %s\n", p.Name)
		if p.Description != "" {
			fmt.Printf("*%s*\n", p.Description)
		}

		// Truncate content for prompt injection
		content := p.Content
		if len(content) > 500 {
			content = content[:500] + "\n...(truncated)"
		}
		fmt.Println(content)
		fmt.Println()
	}
	fmt.Println("────────────────────────────────")
	fmt.Println()

	return nil
}
