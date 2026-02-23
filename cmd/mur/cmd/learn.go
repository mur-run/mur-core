package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/async"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/learn"
	"github.com/mur-run/mur-core/internal/learning"
	"github.com/mur-run/mur-core/internal/notify"
	"github.com/mur-run/mur-core/internal/sysinfo"
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

		// Send notification
		if notify.IsConfigured() {
			opts := notify.Options{
				PatternName: p.Name,
				Confidence:  p.Confidence,
				Preview:     p.Description,
			}
			if err := notify.Notify(notify.EventPatternAdded, opts); err != nil {
				// Don't fail on notification error, just log
				fmt.Printf("  ⚠ Notification failed: %v\n", err)
			}
		}

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

var learnExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract patterns from coding sessions",
	Long: `Extract patterns from Claude Code session transcripts.

Examples:
  mur learn extract                      # Interactive: choose session
  mur learn extract --session abc123     # From specific session
  mur learn extract --auto               # Auto mode (quiet, strict, accept-all)
  mur learn extract --auto --dry-run     # Preview without saving
  mur learn extract --auto --verbose     # Auto mode with output
  mur learn extract --auto --no-strict   # Auto mode without quality filter
  mur learn extract --llm                # Use LLM (default from config)
  mur learn extract --llm ollama         # Use local Ollama

When --auto is specified, these defaults apply:
  --quiet       (use --verbose to override)
  --strict      (use --no-strict to override)
  --accept-all  (use --interactive to override)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --async: re-exec as detached background process
		asyncMode, _ := cmd.Flags().GetBool("async")
		if asyncMode {
			return async.RunBackground(os.Args[1:])
		}

		// --timeout: wrap in context with deadline
		timeoutStr, _ := cmd.Flags().GetString("timeout")
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
			// Default timeout for extract: 2 minutes
			ctx, cancel = context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
		}

		sessionID, _ := cmd.Flags().GetString("session")
		auto, _ := cmd.Flags().GetBool("auto")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		minConfidence, _ := cmd.Flags().GetFloat64("min-confidence")
		llm, _ := cmd.Flags().GetString("llm")
		llmModel, _ := cmd.Flags().GetString("llm-model")

		// Get explicit flag values
		acceptAll, _ := cmd.Flags().GetBool("accept-all")
		quiet, _ := cmd.Flags().GetBool("quiet")
		strict, _ := cmd.Flags().GetBool("strict")
		verbose, _ := cmd.Flags().GetBool("verbose")
		noStrict, _ := cmd.Flags().GetBool("no-strict")
		interactive, _ := cmd.Flags().GetBool("interactive")

		// When --auto is specified, apply sensible defaults
		if auto {
			// Default to quiet unless --verbose is specified
			if !cmd.Flags().Changed("quiet") && !verbose {
				quiet = true
			}
			if verbose {
				quiet = false
			}

			// Default to strict unless --no-strict is specified
			if !cmd.Flags().Changed("strict") && !noStrict {
				strict = true
			}
			if noStrict {
				strict = false
			}

			// Default to accept-all unless --interactive is specified
			if !cmd.Flags().Changed("accept-all") && !interactive {
				acceptAll = true
			}
			if interactive {
				acceptAll = false
			}
		}

		// LLM mode
		if llm != "" {
			return runExtractLLM(ctx, sessionID, llm, llmModel, dryRun, acceptAll, quiet, strict, minConfidence)
		}

		if auto {
			return runExtractAuto(ctx, dryRun, acceptAll, quiet, minConfidence)
		}

		if sessionID != "" {
			return runExtractSession(ctx, sessionID, dryRun, acceptAll, minConfidence)
		}

		// Interactive mode: list sessions and let user choose
		return runExtractInteractive(ctx, dryRun)
	},
}

var learnInitRepoCmd = &cobra.Command{
	Use:   "init <repo-url>",
	Short: "Initialize learning repo for pattern sync",
	Long: `Initialize a git repository for syncing learned patterns across machines.

Each machine uses its own branch (based on hostname by default).
Shared patterns can be pulled from the main branch.

Examples:
  mur learn init git@github.com:user/learning-patterns.git
  mur learn init https://github.com/user/learning-patterns.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		fmt.Printf("Initializing learning repo: %s\n", repoURL)

		if err := learning.InitRepo(repoURL); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		branch, _ := learning.GetBranch()
		fmt.Println("")
		fmt.Println("✓ Learning repo initialized")
		fmt.Printf("  Branch: %s\n", branch)
		fmt.Println("  Run 'mur learn push' to sync patterns")

		return nil
	},
}

var learnPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push patterns to learning repo",
	Long: `Push local patterns to your branch in the learning repo.

The branch name defaults to your hostname, so each machine
has its own set of patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !learning.IsInitialized() {
			return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
		}

		fmt.Println("Pushing patterns to learning repo...")

		if err := learning.Push(); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("✓ Patterns pushed")

		// Check for auto-merge flag
		autoMerge, _ := cmd.Flags().GetBool("auto-merge")
		if autoMerge {
			fmt.Println("")
			fmt.Println("Checking for high-confidence patterns to merge...")

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			result, err := learning.AutoMerge(dryRun)
			if err != nil {
				return fmt.Errorf("auto-merge failed: %w", err)
			}

			if result.PatternsChecked == 0 {
				fmt.Println("  No patterns meet the confidence threshold")
			} else {
				for _, pr := range result.Patterns {
					if pr.Error != nil {
						fmt.Printf("  ✗ %s: %v\n", pr.Pattern.Name, pr.Error)
					} else {
						fmt.Printf("  ✓ %s: %s\n", pr.Pattern.Name, pr.PRURL)
					}
				}
				fmt.Printf("\nPRs created: %d, failed: %d\n", result.PRsCreated, result.PRsFailed)
			}
		}

		return nil
	},
}

var learnAutoMergeCmd = &cobra.Command{
	Use:   "auto-merge",
	Short: "Create PRs for high-confidence patterns",
	Long: `Check patterns with confidence >= threshold and create PRs to main branch.

The threshold is configured in ~/.mur/config.yaml under learning.merge_threshold
(default: 0.8).

Examples:
  mur learn auto-merge              # Create PRs for patterns >= 80% confidence
  mur learn auto-merge --dry-run    # Preview without creating PRs
  mur learn auto-merge --threshold 0.9  # Use custom threshold`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !learning.IsInitialized() {
			return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		threshold, _ := cmd.Flags().GetFloat64("threshold")

		// Override threshold in config if specified
		if threshold > 0 {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}
			cfg.Learning.MergeThreshold = threshold
			// Don't save, just use for this run
		}

		if dryRun {
			fmt.Println("Checking for high-confidence patterns (dry-run)...")
		} else {
			fmt.Println("Checking for high-confidence patterns...")
		}
		fmt.Println("")

		result, err := learning.AutoMerge(dryRun)
		if err != nil {
			return fmt.Errorf("auto-merge failed: %w", err)
		}

		if result.PatternsChecked == 0 {
			fmt.Println("No patterns meet the confidence threshold.")
			return nil
		}

		fmt.Printf("Found %d pattern(s) with high confidence:\n\n", result.PatternsChecked)

		for _, pr := range result.Patterns {
			status := "✓"
			msg := pr.PRURL
			if pr.Error != nil {
				status = "✗"
				msg = pr.Error.Error()
			}
			fmt.Printf("  %s %-20s (%.0f%%) - %s\n",
				status,
				pr.Pattern.Name,
				pr.Pattern.Confidence*100,
				msg)
		}

		fmt.Println("")
		if dryRun {
			fmt.Println("(dry-run mode, no PRs created)")
		} else {
			fmt.Printf("PRs created: %d, failed: %d\n", result.PRsCreated, result.PRsFailed)

			// Send notifications for created PRs
			if notify.IsConfigured() && result.PRsCreated > 0 {
				for _, pr := range result.Patterns {
					if pr.Error == nil && pr.PRURL != "" {
						opts := notify.Options{
							PatternName: pr.Pattern.Name,
							Confidence:  pr.Pattern.Confidence,
							PRURL:       pr.PRURL,
						}
						if err := notify.Notify(notify.EventPRCreated, opts); err != nil {
							fmt.Printf("  ⚠ Notification failed for %s: %v\n", pr.Pattern.Name, err)
						}
					}
				}
			}
		}

		return nil
	},
}

var learnPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull shared patterns from main branch",
	Long: `Pull shared patterns from the main branch of the learning repo.

This imports patterns that others have shared without overwriting
your local patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !learning.IsInitialized() {
			return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
		}

		fmt.Println("Pulling patterns from main branch...")

		if err := learning.Pull(); err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}

		fmt.Println("✓ Patterns pulled")
		return nil
	},
}

var learnSyncRepoCmd = &cobra.Command{
	Use:   "repo-sync",
	Short: "Sync patterns with learning repo (push + pull)",
	Long: `Sync patterns bidirectionally with the learning repo.

This pushes your local patterns to your branch, then pulls
shared patterns from main (if pull_from_main is enabled).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !learning.IsInitialized() {
			return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
		}

		fmt.Println("Syncing with learning repo...")

		if err := learning.Sync(); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		fmt.Println("✓ Sync complete")
		return nil
	},
}

func runExtractAuto(ctx context.Context, dryRun, acceptAll, quiet bool, minConfidence float64) error {
	if minConfidence == 0 {
		minConfidence = 0.6 // Default threshold for auto-accept
	}

	if !quiet {
		fmt.Println("Scanning recent sessions...")
		fmt.Println("")
	}

	sessions, err := learn.RecentSessions(7)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		if !quiet {
			fmt.Println("No recent sessions found.")
		}
		return nil
	}

	totalExtracted := 0
	savedCount := 0
	skippedCount := 0

	for _, session := range sessions {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("timeout exceeded: %w", err)
		}
		patterns, err := learn.ExtractFromSession(session.Path)
		if err != nil {
			continue
		}

		if len(patterns) == 0 {
			continue
		}

		if !quiet {
			fmt.Printf("Session: %s (%s)\n", session.ShortID(), session.Project)
			fmt.Println(strings.Repeat("-", 40))
		}

		for _, ep := range patterns {
			totalExtracted++

			if !quiet {
				displayExtractedPattern(ep)
			}

			if dryRun {
				if !quiet {
					fmt.Println("")
				}
				continue
			}

			// Accept all mode: auto-save if confidence >= threshold
			if acceptAll {
				if ep.Confidence >= minConfidence {
					if err := learn.Add(ep.Pattern); err != nil {
						if !quiet {
							fmt.Printf("  ✗ Failed to save: %v\n", err)
						}
					} else {
						if !quiet {
							fmt.Printf("  ✓ Auto-saved '%s' (%.0f%% confidence)\n", ep.Pattern.Name, ep.Confidence*100)
						}
						savedCount++
					}
				} else {
					skippedCount++
					if !quiet {
						fmt.Printf("  ⊘ Skipped (%.0f%% < %.0f%% threshold)\n", ep.Confidence*100, minConfidence*100)
					}
				}
			} else {
				// Interactive mode
				if confirmSave(ep.Pattern.Name) {
					if err := learn.Add(ep.Pattern); err != nil {
						fmt.Printf("  ✗ Failed to save: %v\n", err)
					} else {
						fmt.Printf("  ✓ Saved as '%s'\n", ep.Pattern.Name)
						savedCount++
					}
				}
			}

			if !quiet {
				fmt.Println("")
			}
		}
	}

	if !quiet {
		if totalExtracted == 0 {
			fmt.Println("No patterns found in recent sessions.")
		} else if dryRun {
			fmt.Printf("\nFound %d potential patterns (dry-run, not saved)\n", totalExtracted)
		} else if acceptAll {
			fmt.Printf("\nProcessed %d patterns: %d saved, %d skipped\n", totalExtracted, savedCount, skippedCount)
		}
	}

	// Auto-push if enabled and patterns were saved
	if !dryRun && savedCount > 0 {
		cfg, err := config.Load()
		if err == nil && cfg.Learning.AutoPush && learning.IsInitialized() {
			if !quiet {
				fmt.Println("")
				fmt.Println("Auto-pushing to learning repo...")
			}
			if err := learning.Push(); err != nil {
				if !quiet {
					fmt.Printf("  ⚠ auto-push failed: %v\n", err)
				}
			} else if !quiet {
				fmt.Println("  ✓ Patterns pushed to learning repo")
			}
		}

		// Send notification about extracted patterns
		if notify.IsConfigured() {
			opts := notify.Options{
				Count: savedCount,
			}
			if err := notify.Notify(notify.EventPatternsExtracted, opts); err != nil && !quiet {
				fmt.Printf("  ⚠ Notification failed: %v\n", err)
			}
		}
	}

	return nil
}

func runExtractLLM(ctx context.Context, sessionID, provider, model string, dryRun, acceptAll, quiet, strict bool, minConfidence float64) error {
	// Setup quality config for strict mode
	qualityCfg := learn.DefaultExtractionConfig()

	// Setup LLM options
	opts := learn.DefaultLLMOptions()
	configuredProvider := false

	// Load config for defaults
	cfg, _ := config.Load()
	if cfg != nil && cfg.Learning.LLM.Provider != "" {
		configuredProvider = true
		// Use config defaults
		switch strings.ToLower(cfg.Learning.LLM.Provider) {
		case "ollama":
			opts.Provider = learn.LLMOllama
		case "claude":
			opts.Provider = learn.LLMClaude
		case "openai":
			opts.Provider = learn.LLMOpenAI
		case "gemini":
			opts.Provider = learn.LLMGemini
		}
		if cfg.Learning.LLM.Model != "" {
			opts.Model = cfg.Learning.LLM.Model
		}
		if cfg.Learning.LLM.OllamaURL != "" {
			opts.OllamaURL = cfg.Learning.LLM.OllamaURL
		}
		if cfg.Learning.LLM.OpenAIURL != "" {
			opts.OpenAIURL = cfg.Learning.LLM.OpenAIURL
		}
		// Support custom API key env var
		if cfg.Learning.LLM.APIKeyEnv != "" {
			key := os.Getenv(cfg.Learning.LLM.APIKeyEnv)
			if key != "" {
				switch opts.Provider {
				case learn.LLMOpenAI:
					opts.OpenAIKey = key
				case learn.LLMGemini:
					opts.GeminiKey = key
				case learn.LLMClaude:
					opts.ClaudeKey = key
				}
			}
		}
	}

	// Command line flags override config
	switch strings.ToLower(provider) {
	case "ollama":
		opts.Provider = learn.LLMOllama
		configuredProvider = true
	case "claude":
		opts.Provider = learn.LLMClaude
		configuredProvider = true
	case "openai":
		opts.Provider = learn.LLMOpenAI
		configuredProvider = true
	case "gemini":
		opts.Provider = learn.LLMGemini
		configuredProvider = true
	case "", "default":
		// Use config default (already set above), or auto-detect
	default:
		return fmt.Errorf("unknown LLM provider: %s (use 'ollama', 'claude', 'openai', or 'gemini')", provider)
	}

	if model != "" {
		opts.Model = model
	}

	// Auto-detect: if no provider configured, try Ollama
	if !configuredProvider {
		if sysinfo.OllamaRunning(opts.OllamaURL) {
			opts.Provider = learn.LLMOllama
			if !quiet {
				fmt.Println("💡 No LLM configured, using local Ollama")
				fmt.Println("   Tip: Configure in ~/.mur/config.yaml for better control")
				fmt.Println()
			}
		} else {
			// No LLM available - always warn (even in quiet mode)
			fmt.Fprintln(os.Stderr, "⚠️  No LLM available (Ollama not running, no API keys)")
			fmt.Fprintln(os.Stderr, "   Falling back to keyword extraction (lower quality)")
			// Call keyword-based extraction instead
			return runExtractAuto(ctx, dryRun, acceptAll, quiet, minConfidence)
		}
	}

	// Validate provider setup
	switch opts.Provider {
	case learn.LLMOllama:
		if !sysinfo.OllamaRunning(opts.OllamaURL) {
			// Always warn (even in quiet mode)
			fmt.Fprintln(os.Stderr, "⚠️  Ollama not available, falling back to keyword extraction")
			return runExtractAuto(ctx, dryRun, acceptAll, quiet, minConfidence)
		}
	case learn.LLMClaude:
		if opts.ClaudeKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY not set")
		}
	case learn.LLMOpenAI:
		if opts.OpenAIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY not set")
		}
	case learn.LLMGemini:
		if opts.GeminiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY not set")
		}
	}

	if minConfidence == 0 {
		minConfidence = 0.6
	}

	// Get sessions to process
	var sessions []*learn.Session

	if sessionID != "" {
		// Single session
		session, err := learn.LoadSession(sessionID)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		sessions = append(sessions, session)
	} else {
		// Recent sessions
		if !quiet {
			fmt.Println("Scanning recent sessions...")
		}
		recentSessions, err := learn.RecentSessions(7)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		for _, s := range recentSessions {
			sess, err := learn.LoadSession(s.Path)
			if err != nil {
				continue
			}
			sessions = append(sessions, sess)
		}
	}

	if len(sessions) == 0 {
		if !quiet {
			fmt.Println("No sessions found.")
		}
		return nil
	}

	// Setup premium options if configured
	var premiumOpts *learn.LLMExtractOptions
	if cfg != nil && cfg.Learning.LLM.Premium != nil {
		p := cfg.Learning.LLM.Premium
		po := learn.DefaultLLMOptions()
		switch strings.ToLower(p.Provider) {
		case "ollama":
			po.Provider = learn.LLMOllama
		case "claude":
			po.Provider = learn.LLMClaude
		case "openai":
			po.Provider = learn.LLMOpenAI
		case "gemini":
			po.Provider = learn.LLMGemini
		}
		if p.Model != "" {
			po.Model = p.Model
		}
		if p.OllamaURL != "" {
			po.OllamaURL = p.OllamaURL
		}
		if p.OpenAIURL != "" {
			po.OpenAIURL = p.OpenAIURL
		}
		if p.APIKeyEnv != "" {
			key := os.Getenv(p.APIKeyEnv)
			if key != "" {
				switch po.Provider {
				case learn.LLMOpenAI:
					po.OpenAIKey = key
				case learn.LLMGemini:
					po.GeminiKey = key
				case learn.LLMClaude:
					po.ClaudeKey = key
				}
			}
		}
		premiumOpts = &po
	}

	if !quiet {
		fmt.Printf("Using %s for extraction...\n", opts.Provider)
		if premiumOpts != nil {
			fmt.Printf("Premium model: %s (for important sessions)\n", premiumOpts.Provider)
		}
		fmt.Println()
	}

	totalExtracted := 0
	savedCount := 0
	skippedSessions := 0
	consecutiveErrors := 0
	var lastError string

	for _, session := range sessions {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("timeout exceeded: %w", err)
		}
		// Stop if we get too many consecutive errors (likely config issue)
		if consecutiveErrors >= 3 {
			errMsg := fmt.Sprintf("LLM Error: %s", lastError)
			fmt.Fprintln(os.Stderr, "⛔ Stopping: 3 consecutive extraction failures")
			fmt.Fprintf(os.Stderr, "   Last error: %s\n", lastError)
			fmt.Fprintln(os.Stderr, "   Check your LLM configuration in ~/.mur/config.yaml")
			// Send system notification
			_ = notify.NotifyCritical("mur: Extraction Failed", errMsg)
			break
		}

		// Strict mode: pre-filter sessions by quality
		if strict {
			quality := learn.AnalyzeSessionQuality(session)
			shouldExtract, reason := learn.ShouldExtract(quality, qualityCfg)
			if !shouldExtract {
				skippedSessions++
				if !quiet {
					fmt.Printf("⊘ Skipping %s: %s\n", session.ShortID(), reason)
				}
				continue
			}
		}

		// Check if this session should use premium model
		useOpts := opts
		usePremium := false
		if premiumOpts != nil && cfg.Learning.LLM.Routing != nil {
			routing := cfg.Learning.LLM.Routing
			// Check message count
			if routing.MinMessages > 0 && len(session.Messages) >= routing.MinMessages {
				usePremium = true
			}
			// Check project patterns
			for _, proj := range routing.Projects {
				if strings.Contains(strings.ToLower(session.Project), strings.ToLower(proj)) {
					usePremium = true
					break
				}
			}
			if usePremium {
				useOpts = *premiumOpts
			}
		}

		if !quiet {
			if usePremium {
				fmt.Printf("📝 Session: %s (%s) ⭐ premium\n", session.ShortID(), session.Project)
			} else {
				fmt.Printf("📝 Session: %s (%s)\n", session.ShortID(), session.Project)
			}
		}

		patterns, err := learn.ExtractWithLLM(session, useOpts)
		if err != nil {
			// If premium failed, fallback to default model
			if usePremium {
				fmt.Fprintf(os.Stderr, "⚠️  Premium model failed for %s: %v\n", session.ShortID(), err)
				if !quiet {
					fmt.Printf("   ↪ Falling back to %s...\n", opts.Provider)
				}
				patterns, err = learn.ExtractWithLLM(session, opts)
			}
			if err != nil {
				// Track consecutive errors
				consecutiveErrors++
				lastError = err.Error()
				// Only print first error of each type
				if consecutiveErrors == 1 {
					fmt.Fprintf(os.Stderr, "⚠️  Extraction failed: %v\n", err)
				}
				continue
			}
		}

		// Reset consecutive error counter on success
		consecutiveErrors = 0

		// Strict mode: filter patterns by quality
		if strict {
			patterns = learn.FilterPatterns(patterns, qualityCfg)
		}

		if len(patterns) == 0 {
			if !quiet {
				fmt.Println("   No patterns found")
			}
			continue
		}

		if !quiet {
			fmt.Printf("   Found %d patterns:\n", len(patterns))
		}

		for _, ep := range patterns {
			totalExtracted++

			if !quiet {
				fmt.Printf("   • [%s] %s (%.0f%%)\n", ep.Pattern.Category, ep.Pattern.Name, ep.Confidence*100)
			}

			if dryRun {
				continue
			}

			if acceptAll {
				if ep.Confidence >= minConfidence {
					if err := learn.Add(ep.Pattern); err != nil {
						if !quiet {
							fmt.Printf("     ✗ Failed to save: %v\n", err)
						}
					} else {
						if !quiet {
							fmt.Printf("     ✓ Saved\n")
						}
						savedCount++
					}
				}
			} else {
				// Interactive mode
				if confirmSave(ep.Pattern.Name) {
					if err := learn.Add(ep.Pattern); err != nil {
						fmt.Printf("     ✗ Failed to save: %v\n", err)
					} else {
						fmt.Printf("     ✓ Saved\n")
						savedCount++
					}
				}
			}
		}

		if !quiet {
			fmt.Println()
		}
	}

	if !quiet {
		if dryRun {
			fmt.Printf("Found %d patterns (dry-run, not saved)\n", totalExtracted)
		} else {
			fmt.Printf("Extracted %d patterns, saved %d\n", totalExtracted, savedCount)
		}
		if strict && skippedSessions > 0 {
			fmt.Printf("Skipped %d low-quality sessions (strict mode)\n", skippedSessions)
		}
	}

	// Send notification for successful extraction
	if !dryRun && savedCount > 0 {
		_ = notify.NotifySuccess(fmt.Sprintf("%d new patterns extracted", savedCount))
	}

	return nil
}

func runExtractSession(_ context.Context, sessionID string, dryRun, acceptAll bool, minConfidence float64) error {
	session, err := learn.LoadSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	fmt.Printf("Extracting from session: %s\n", session.ShortID())
	fmt.Printf("Project: %s\n", session.Project)
	fmt.Printf("Messages: %d\n", len(session.Messages))
	fmt.Println("")

	patterns, err := learn.ExtractFromSession(session.Path)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	if len(patterns) == 0 {
		fmt.Println("No patterns found in this session.")
		return nil
	}

	fmt.Printf("Found %d potential patterns:\n\n", len(patterns))

	saved := 0
	skipped := 0

	for i, ep := range patterns {
		fmt.Printf("%d. ", i+1)
		displayExtractedPattern(ep)

		if !dryRun {
			shouldSave := false

			if acceptAll {
				// Auto-accept if confidence meets threshold
				if ep.Confidence >= minConfidence {
					shouldSave = true
				} else {
					fmt.Printf("   Skipped (confidence %.0f%% < %.0f%%)\n", ep.Confidence*100, minConfidence*100)
					skipped++
				}
			} else {
				// Interactive mode
				shouldSave = confirmSave(ep.Pattern.Name)
			}

			if shouldSave {
				if err := learn.Add(ep.Pattern); err != nil {
					fmt.Printf("  ✗ Failed to save: %v\n", err)
				} else {
					fmt.Printf("  ✓ Saved as '%s'\n", ep.Pattern.Name)
					saved++
				}
			}
		}
		fmt.Println("")
	}

	if dryRun {
		fmt.Println("(dry-run mode, patterns not saved)")
	} else if acceptAll {
		fmt.Printf("Saved %d patterns, skipped %d (below %.0f%% confidence)\n", saved, skipped, minConfidence*100)
	}

	return nil
}

func runExtractInteractive(ctx context.Context, dryRun bool) error {
	sessions, err := learn.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found in ~/.claude/projects/")
		return nil
	}

	// Show recent sessions
	fmt.Println("Recent Sessions")
	fmt.Println("===============")
	fmt.Println("")

	limit := 10
	if len(sessions) < limit {
		limit = len(sessions)
	}

	for i, s := range sessions[:limit] {
		fmt.Printf("  %d. %s  (%s)  %s\n",
			i+1,
			s.ShortID(),
			s.Project,
			s.CreatedAt.Format("2006-01-02 15:04"))
	}

	fmt.Println("")
	fmt.Print("Select session (1-10) or 'q' to quit: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "" {
		return nil
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > limit {
		return fmt.Errorf("invalid selection")
	}

	selected := sessions[idx-1]
	fmt.Println("")

	// In interactive mode, don't auto-accept (user chose to interact)
	return runExtractSession(ctx, selected.ID, dryRun, false, 0.6)
}

func displayExtractedPattern(ep learn.ExtractedPattern) {
	fmt.Printf("[%s] %s (confidence: %.0f%%)\n",
		ep.Pattern.Category,
		ep.Pattern.Name,
		ep.Confidence*100)
	fmt.Printf("   Source: session %s\n", ep.Source)
	fmt.Printf("   Domain: %s\n", ep.Pattern.Domain)
	if len(ep.Evidence) > 0 {
		fmt.Printf("   Preview: %s\n", truncate(ep.Evidence[0], 80))
	}
}

func confirmSave(name string) bool {
	fmt.Printf("   Save pattern '%s'? [y/N/e(dit)] ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func init() {
	rootCmd.AddCommand(learnCmd)
	learnCmd.AddCommand(learnListCmd)
	learnCmd.AddCommand(learnAddCmd)
	learnCmd.AddCommand(learnGetCmd)
	learnCmd.AddCommand(learnDeleteCmd)
	learnCmd.AddCommand(learnSyncCmd)
	learnCmd.AddCommand(learnExtractCmd)
	learnCmd.AddCommand(learnInitRepoCmd)
	learnCmd.AddCommand(learnPushCmd)
	learnCmd.AddCommand(learnPullCmd)
	learnCmd.AddCommand(learnSyncRepoCmd)
	learnCmd.AddCommand(learnAutoMergeCmd)

	learnListCmd.Flags().StringP("domain", "d", "", "Filter by domain")
	learnListCmd.Flags().StringP("category", "c", "", "Filter by category")

	learnAddCmd.Flags().Bool("stdin", false, "Read content from stdin")

	learnDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	learnSyncCmd.Flags().Bool("cleanup", false, "Remove orphaned synced patterns")

	learnExtractCmd.Flags().StringP("session", "s", "", "Session ID to extract from")
	learnExtractCmd.Flags().Bool("auto", false, "Automatically scan recent sessions (implies --quiet --strict --accept-all)")
	learnExtractCmd.Flags().Bool("dry-run", false, "Show what would be extracted without saving")
	learnExtractCmd.Flags().Bool("accept-all", false, "Auto-save patterns above confidence threshold")
	learnExtractCmd.Flags().Bool("quiet", false, "Silent mode (for hooks, minimal output)")
	learnExtractCmd.Flags().Bool("strict", false, "Enable strict quality filtering (skip Q&A sessions, validate patterns)")
	learnExtractCmd.Flags().BoolP("verbose", "V", false, "Show detailed output (overrides --quiet in auto mode)")
	learnExtractCmd.Flags().Bool("no-strict", false, "Disable strict quality filtering in auto mode")
	learnExtractCmd.Flags().BoolP("interactive", "i", false, "Prompt for each pattern in auto mode (overrides --accept-all)")
	learnExtractCmd.Flags().Float64("min-confidence", 0.6, "Minimum confidence for auto-accept (default: 0.6)")
	learnExtractCmd.Flags().StringP("llm", "l", "", "LLM provider: ollama, claude, openai, gemini (default from config)")
	learnExtractCmd.Flags().Lookup("llm").NoOptDefVal = "default" // --llm without value uses config default
	learnExtractCmd.Flags().String("llm-model", "", "LLM model (default from config)")
	learnExtractCmd.Flags().Bool("async", false, "Run in background (detached process, parent exits immediately)")
	learnExtractCmd.Flags().String("timeout", "", "Timeout duration (e.g. '30s', '2m'). Default: 2m")

	learnPushCmd.Flags().Bool("auto-merge", false, "Check and create PRs for high-confidence patterns after push")
	learnPushCmd.Flags().Bool("dry-run", false, "Preview auto-merge without creating PRs")

	learnAutoMergeCmd.Flags().Bool("dry-run", false, "Preview without creating PRs")
	learnAutoMergeCmd.Flags().Float64("threshold", 0, "Override confidence threshold (default: from config or 0.8)")
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
