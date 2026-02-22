package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/embed"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/router"
	"github.com/mur-run/mur-core/internal/stats"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a prompt with an AI tool",
	Long: `Run a prompt using the configured AI tool.

With routing.mode=auto (default), murmur selects the best tool based on
prompt complexity. Simple questions use free tools; complex tasks use paid.

Patterns are automatically injected based on project context and prompt analysis.
Use --no-inject to disable pattern injection.

Use -t to override automatic selection.

Examples:
  mur run -p "what is git?"              # Auto-routes to free tool
  mur run -p "refactor this module"      # Auto-routes to paid tool
  mur run -p "explain x" -t claude       # Force specific tool
  mur run -p "test" --explain            # Show routing decision only
  mur run -p "fix bug" --no-inject       # Skip pattern injection`,
	RunE: runExecute,
}

func runExecute(cmd *cobra.Command, args []string) error {
	prompt, _ := cmd.Flags().GetString("prompt")
	forceTool, _ := cmd.Flags().GetString("tool")
	explain, _ := cmd.Flags().GetBool("explain")
	noInject, _ := cmd.Flags().GetBool("no-inject")
	verbose, _ := cmd.Flags().GetBool("verbose")
	timeoutStr, _ := cmd.Flags().GetString("timeout")

	// --timeout: create context with deadline (no default = unlimited)
	var ctx context.Context
	var cancel context.CancelFunc
	if timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid --timeout value %q: %w", timeoutStr, err)
		}
		ctx, cancel = context.WithTimeout(context.Background(), d)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	if prompt == "" {
		return fmt.Errorf("prompt is required. Use -p \"your prompt\"")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Pattern injection
	finalPrompt := prompt
	var injectionResult *inject.InjectionResult

	if !noInject {
		// Get working directory
		workDir, _ := os.Getwd()

		// Initialize pattern store
		patternsDir := filepath.Join(os.Getenv("HOME"), ".mur", "patterns")
		store := pattern.NewStore(patternsDir)

		// Create injector and inject patterns
		injector := inject.NewInjector(store)

		// Try to enable semantic search (non-fatal if it fails)
		embedCfg := embed.DefaultConfig()
		if err := injector.WithSemanticSearch(embedCfg); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "⚠ Semantic search unavailable: %v\n", err)
			}
			// Fall back to keyword matching (built-in)
		}

		injectionResult, err = injector.Inject(prompt, workDir)
		if err != nil {
			// Non-fatal: warn but continue
			if verbose {
				fmt.Fprintf(os.Stderr, "⚠ Pattern injection failed: %v\n", err)
			}
		} else if len(injectionResult.Patterns) > 0 {
			finalPrompt = injectionResult.FormattedPrompt
			if verbose {
				fmt.Printf("📚 Injected %d patterns:\n", len(injectionResult.Patterns))
				for _, p := range injectionResult.Patterns {
					fmt.Printf("   • %s\n", p.Name)
				}
				if injectionResult.Context != nil && injectionResult.Context.ProjectType != "" {
					fmt.Printf("🔍 Context: %s project (%s)\n",
						injectionResult.Context.ProjectType,
						injectionResult.Context.ProjectName)
				}
				fmt.Println()
			}
		}
	}

	var tool string
	var reason string
	var complexity float64
	autoRouted := forceTool == ""

	if forceTool != "" {
		// User explicitly chose a tool
		tool = forceTool
		reason = "user specified with -t flag"
		// Still analyze for stats
		analysis := router.AnalyzePrompt(prompt)
		complexity = analysis.Complexity
	} else {
		// Use router
		selection, err := router.SelectTool(prompt, cfg)
		if err != nil {
			return fmt.Errorf("routing failed: %w", err)
		}
		tool = selection.Tool
		reason = selection.Reason
		complexity = selection.Analysis.Complexity

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

			// Show pattern info
			if injectionResult != nil && len(injectionResult.Patterns) > 0 {
				fmt.Println("Pattern Injection")
				fmt.Println("-----------------")
				fmt.Printf("Patterns:   %d matched\n", len(injectionResult.Patterns))
				for _, p := range injectionResult.Patterns {
					fmt.Printf("  • %s\n", p.Name)
				}
				if injectionResult.Context != nil && injectionResult.Context.ProjectType != "" {
					fmt.Printf("Project:    %s (%s)\n", injectionResult.Context.ProjectName, injectionResult.Context.ProjectType)
				}
				fmt.Println()
			}

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

	// Build command args (use finalPrompt with injected patterns)
	cmdArgs := append(toolCfg.Flags, finalPrompt)

	// Check if binary exists
	binPath, err := exec.LookPath(toolCfg.Binary)
	if err != nil {
		return fmt.Errorf("%s not found in PATH. Install it first", toolCfg.Binary)
	}

	// Show execution info
	if injectionResult != nil && len(injectionResult.Patterns) > 0 {
		fmt.Printf("→ %s (%s) [%d patterns]\n\n", tool, reason, len(injectionResult.Patterns))
	} else {
		fmt.Printf("→ %s (%s)\n\n", tool, reason)
	}

	// Execute the tool and track stats
	startTime := time.Now()
	execCmd := exec.CommandContext(ctx, binPath, cmdArgs...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	runErr := execCmd.Run()
	duration := time.Since(startTime)

	// Record stats (ignore errors - stats are non-critical)
	_ = stats.Record(stats.UsageRecord{
		Tool:         tool,
		Timestamp:    startTime,
		PromptLength: len(prompt),
		DurationMs:   duration.Milliseconds(),
		CostEstimate: stats.EstimateCost(tool, len(prompt)),
		Tier:         toolCfg.Tier,
		RoutingMode:  cfg.Routing.Mode,
		AutoRouted:   autoRouted,
		Complexity:   complexity,
		Success:      runErr == nil,
	})

	// Track pattern usage for effectiveness learning
	if injectionResult != nil && len(injectionResult.Patterns) > 0 {
		trackingDir := filepath.Join(os.Getenv("HOME"), ".mur", "tracking")
		patternsDir := filepath.Join(os.Getenv("HOME"), ".mur", "patterns")
		tracker := inject.NewTracker(pattern.NewStore(patternsDir), trackingDir)
		_ = tracker.RecordUsage(injectionResult.Patterns, injectionResult.Context, prompt, runErr == nil)
	}

	return runErr
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
	runCmd.Flags().Bool("no-inject", false, "Disable automatic pattern injection")
	runCmd.Flags().BoolP("verbose", "V", false, "Show pattern injection details")
	runCmd.Flags().String("timeout", "", "Timeout duration (e.g. '30s', '5m'). Default: unlimited")
}
