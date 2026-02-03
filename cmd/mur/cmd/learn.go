package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/karajanchang/murmur-ai/internal/config"
	"github.com/karajanchang/murmur-ai/internal/learn"
	"github.com/karajanchang/murmur-ai/internal/learning"
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

var learnExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract patterns from coding sessions",
	Long: `Extract patterns from Claude Code session transcripts.

Examples:
  mur learn extract                      # Interactive: choose session
  mur learn extract --session abc123     # From specific session
  mur learn extract --auto               # Scan recent sessions
  mur learn extract --auto --dry-run     # Preview without saving`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID, _ := cmd.Flags().GetString("session")
		auto, _ := cmd.Flags().GetBool("auto")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if auto {
			return runExtractAuto(dryRun)
		}

		if sessionID != "" {
			return runExtractSession(sessionID, dryRun)
		}

		// Interactive mode: list sessions and let user choose
		return runExtractInteractive(dryRun)
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

The threshold is configured in ~/.murmur/config.yaml under learning.merge_threshold
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

func runExtractAuto(dryRun bool) error {
	fmt.Println("Scanning recent sessions...")
	fmt.Println("")

	sessions, err := learn.RecentSessions(7)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No recent sessions found.")
		return nil
	}

	totalExtracted := 0
	savedCount := 0
	for _, session := range sessions {
		patterns, err := learn.ExtractFromSession(session.Path)
		if err != nil {
			continue
		}

		if len(patterns) == 0 {
			continue
		}

		fmt.Printf("Session: %s (%s)\n", session.ShortID(), session.Project)
		fmt.Println(strings.Repeat("-", 40))

		for _, ep := range patterns {
			displayExtractedPattern(ep)
			totalExtracted++

			if !dryRun {
				if confirmSave(ep.Pattern.Name) {
					if err := learn.Add(ep.Pattern); err != nil {
						fmt.Printf("  ✗ Failed to save: %v\n", err)
					} else {
						fmt.Printf("  ✓ Saved as '%s'\n", ep.Pattern.Name)
						savedCount++
					}
				}
			}
			fmt.Println("")
		}
	}

	if totalExtracted == 0 {
		fmt.Println("No patterns found in recent sessions.")
	} else if dryRun {
		fmt.Printf("\nFound %d potential patterns (dry-run, not saved)\n", totalExtracted)
	}

	// Auto-push if enabled and patterns were saved
	if !dryRun && savedCount > 0 {
		cfg, err := config.Load()
		if err == nil && cfg.Learning.AutoPush && learning.IsInitialized() {
			fmt.Println("")
			fmt.Println("Auto-pushing to learning repo...")
			if err := learning.Push(); err != nil {
				fmt.Printf("  ⚠ auto-push failed: %v\n", err)
			} else {
				fmt.Println("  ✓ Patterns pushed to learning repo")
			}
		}
	}

	return nil
}

func runExtractSession(sessionID string, dryRun bool) error {
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

	for i, ep := range patterns {
		fmt.Printf("%d. ", i+1)
		displayExtractedPattern(ep)

		if !dryRun {
			if confirmSave(ep.Pattern.Name) {
				if err := learn.Add(ep.Pattern); err != nil {
					fmt.Printf("  ✗ Failed to save: %v\n", err)
				} else {
					fmt.Printf("  ✓ Saved as '%s'\n", ep.Pattern.Name)
				}
			}
		}
		fmt.Println("")
	}

	if dryRun {
		fmt.Println("(dry-run mode, patterns not saved)")
	}

	return nil
}

func runExtractInteractive(dryRun bool) error {
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

	return runExtractSession(selected.ID, dryRun)
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
	learnExtractCmd.Flags().Bool("auto", false, "Automatically scan recent sessions")
	learnExtractCmd.Flags().Bool("dry-run", false, "Show what would be extracted without saving")

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
