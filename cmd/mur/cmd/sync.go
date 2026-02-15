package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/learn"
	"github.com/mur-run/mur-core/internal/security"
	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var (
	syncPush     bool
	syncQuiet    bool
	syncFormat   string
	syncCleanOld bool
	syncCloud    bool
	syncGit      bool
	syncCLI      bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync patterns (cloud, git, or to CLIs)",
	Long: `Smart sync command that detects your plan and syncs accordingly.

For Trial/Pro/Team/Enterprise users:
  - Syncs patterns with mur.run cloud (bidirectional)
  - Then syncs to local AI CLIs

For Free users:
  - Syncs with git repo (if configured)
  - Then syncs to local AI CLIs

You can override the default behavior with flags.

Examples:
  mur sync                    # Smart sync based on your plan
  mur sync --cloud            # Force cloud sync
  mur sync --git              # Force git sync
  mur sync --cli              # Only sync to local CLIs (no remote)
  mur sync --quiet            # Silent mode`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncCloud, "cloud", false, "Force cloud sync (requires Trial/Pro/Team/Enterprise)")
	syncCmd.Flags().BoolVar(&syncGit, "git", false, "Force git sync")
	syncCmd.Flags().BoolVar(&syncCLI, "cli", false, "Only sync to local CLIs (no remote sync)")
	syncCmd.Flags().BoolVar(&syncPush, "push", false, "Push local changes to remote (git mode)")
	syncCmd.Flags().BoolVar(&syncQuiet, "quiet", false, "Silent mode (minimal output)")
	syncCmd.Flags().StringVar(&syncFormat, "format", "", "CLI sync format: directory (default) or single")
	syncCmd.Flags().BoolVar(&syncCleanOld, "clean-old", false, "Remove old single-file format files")
}

func runSync(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Override config with flags
	if syncFormat != "" {
		cfg.Sync.Format = syncFormat
	}
	if syncCleanOld {
		cfg.Sync.CleanOld = true
	}

	// Determine sync mode
	useCloud := syncCloud
	useGit := syncGit
	cliOnly := syncCLI

	// If no explicit flag, detect based on plan
	if !useCloud && !useGit && !cliOnly {
		authStore, err := cloud.NewAuthStore()
		if err == nil && authStore.IsLoggedIn() {
			// User is logged in, check plan via /me endpoint
			client, err := cloud.NewClient(cfg.Server.URL)
			if err == nil {
				user, err := client.Me()
				if err == nil {
					plan := user.Plan
					// Trial/Pro/Team/Enterprise get cloud sync
					if plan == "trial" || plan == "pro" || plan == "team" || plan == "enterprise" {
						useCloud = true
						if !syncQuiet {
							fmt.Printf("☁️  Cloud sync (%s plan)\n", plan)
							fmt.Println()
						}
					}
				}
			}
		}

		// If not using cloud, check for git repo
		if !useCloud {
			patternsDir := filepath.Join(home, ".mur", "repo")
			gitDir := filepath.Join(patternsDir, ".git")
			if _, err := os.Stat(gitDir); err == nil {
				useGit = true
				if !syncQuiet {
					fmt.Println("📦 Git sync (local repo)")
					fmt.Println()
				}
			}
		}

		// If neither cloud nor git, just sync to CLIs
		if !useCloud && !useGit {
			cliOnly = true
			if !syncQuiet {
				fmt.Println("💻 Syncing to local CLIs only")
				fmt.Println()
			}
		}
	}

	// Execute cloud sync
	if useCloud {
		if err := runCloudSync(cmd, cfg); err != nil {
			if !syncQuiet {
				fmt.Printf("⚠️  Cloud sync failed: %v\n", err)
			}
			// Continue to CLI sync even if cloud fails
		}
		if !syncQuiet {
			fmt.Println()
		}
	}

	// Execute git sync
	if useGit {
		if err := runGitSync(home, cfg); err != nil {
			if !syncQuiet {
				fmt.Printf("⚠️  Git sync failed: %v\n", err)
			}
		}
		if !syncQuiet {
			fmt.Println()
		}
	}

	// Community auto-share (if enabled and pushing)
	if syncPush && cfg.Community.ShareEnabled && cfg.Community.AutoShareOnPush {
		if err := runCommunityAutoShare(cfg); err != nil {
			if !syncQuiet {
				fmt.Printf("⚠️  Community share: %v\n", err)
			}
		}
	}

	// Sync patterns to all CLIs
	if !syncQuiet {
		format := cfg.Sync.Format
		if format == "" {
			format = "directory"
		}
		fmt.Printf("Syncing patterns to CLIs (format: %s)...\n", format)
	}

	results, err := sync.SyncPatternsWithFormat(cfg)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if !syncQuiet {
		for _, r := range results {
			status := "✓"
			if !r.Success {
				status = "✗"
			}
			fmt.Printf("  %s %s: %s\n", status, r.Target, r.Message)
		}
	}

	// Cleanup community cache (if configured)
	cacheConfig := cfg.GetCacheConfig()
	if cacheConfig.Cleanup == "on_sync" {
		if communityCache, err := cache.DefaultCommunityCache(); err == nil {
			removed, _ := communityCache.Cleanup()
			if removed > 0 && !syncQuiet {
				fmt.Printf("  🧹 Cleaned %d expired community cache entries\n", removed)
			}
		}
	}

	if !syncQuiet {
		fmt.Println()
		fmt.Println("✅ Sync complete")
	}

	return nil
}

// runCloudSync executes cloud sync with mur.run
func runCloudSync(cmd *cobra.Command, cfg *config.Config) error {
	client, err := cloud.NewClient(cfg.Server.URL)
	if err != nil {
		return err
	}

	if !client.AuthStore().IsLoggedIn() {
		return fmt.Errorf("not logged in. Run 'mur login' first")
	}

	// Get team from config
	teamSlug := cfg.Server.Team
	if teamSlug == "" {
		return fmt.Errorf("no team configured. Run 'mur cloud select <team>'")
	}

	// This calls the same logic as 'mur cloud sync'
	// For now, we'll print a message and suggest using mur cloud sync
	if !syncQuiet {
		fmt.Printf("Syncing with cloud team: %s\n", teamSlug)
	}

	// Execute cloud sync by calling the cloud sync function directly
	// Reuse cloudSyncCmd logic
	cmd.Flags().Set("team", teamSlug)
	return cloudSyncCmd.RunE(cmd, []string{})
}

// runGitSync executes git-based sync
func runGitSync(home string, cfg *config.Config) error {
	patternsDir := filepath.Join(home, ".mur", "repo")
	gitDir := filepath.Join(patternsDir, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("no git repo configured at %s", patternsDir)
	}

	if !syncQuiet {
		fmt.Println("Pulling from remote...")
	}
	pullCmd := exec.Command("git", "-C", patternsDir, "pull", "--rebase")
	if !syncQuiet {
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
	}
	if err := pullCmd.Run(); err != nil {
		if !syncQuiet {
			fmt.Printf("  ⚠ Pull failed: %v\n", err)
		}
		return err
	}
	if !syncQuiet {
		fmt.Println("  ✓ Pulled latest patterns")
	}

	// Push if requested
	if syncPush {
		if !syncQuiet {
			fmt.Println("Pushing to remote...")
		}

		addCmd := exec.Command("git", "-C", patternsDir, "add", "-A")
		addCmd.Run()

		diffCmd := exec.Command("git", "-C", patternsDir, "diff", "--cached", "--quiet")
		if diffCmd.Run() != nil {
			commitCmd := exec.Command("git", "-C", patternsDir, "commit", "-m", "mur: sync patterns")
			commitCmd.Run()
		}

		pushCmd := exec.Command("git", "-C", patternsDir, "push")
		if !syncQuiet {
			pushCmd.Stdout = os.Stdout
			pushCmd.Stderr = os.Stderr
		}
		if err := pushCmd.Run(); err != nil {
			if !syncQuiet {
				fmt.Printf("  ⚠ Push failed: %v\n", err)
			}
		} else if !syncQuiet {
			fmt.Println("  ✓ Pushed to remote")
		}
	}

	return nil
}

// runCommunityAutoShare shares patterns to community with secret scanning
func runCommunityAutoShare(cfg *config.Config) error {
	if !syncQuiet {
		fmt.Println()
		fmt.Println("🌍 Community sharing...")
	}

	// Check if logged in
	client, err := cloud.NewClient(cfg.Server.URL)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if !client.AuthStore().IsLoggedIn() {
		if !syncQuiet {
			fmt.Println("  ⏭ Skipped (not logged in)")
		}
		return nil
	}

	// Load all patterns
	patterns, err := learn.List()
	if err != nil {
		return fmt.Errorf("failed to load patterns: %w", err)
	}

	if len(patterns) == 0 {
		if !syncQuiet {
			fmt.Println("  ⏭ No patterns to share")
		}
		return nil
	}

	// Initialize scanner
	scanner := security.NewScanner()

	var shared, skipped int
	for _, p := range patterns {
		// Skip patterns without content
		if p.Content == "" {
			continue
		}

		// Build content to scan (name + description + content)
		contentToScan := p.Name + "\n" + p.Description + "\n" + p.Content

		// Scan for secrets
		result := scanner.ScanContent(contentToScan)
		if !result.Safe {
			if !syncQuiet {
				fmt.Printf("  ⚠️ %s → skipped (secrets detected)\n", p.Name)
				for _, f := range result.Findings {
					fmt.Printf("     └─ %s at line %d: %s\n", f.Type, f.Line, f.Match)
				}
			}
			skipped++
			continue
		}

		// Share to community
		req := &cloud.ShareLocalPatternRequest{
			Name:        p.Name,
			Description: p.Description,
			Content:     p.Content,
			Domain:      p.Domain,
			Category:    p.Category,
			Tags:        p.Tags,
		}

		resp, err := client.ShareLocalPattern(req)
		if err != nil {
			if !syncQuiet {
				fmt.Printf("  ✗ %s → failed: %v\n", p.Name, err)
			}
			continue
		}

		if !syncQuiet {
			status := "shared"
			if resp.Status == "pending" {
				status = "pending review"
			}
			fmt.Printf("  ✓ %s → %s\n", p.Name, status)
		}
		shared++
	}

	if !syncQuiet {
		if shared > 0 {
			fmt.Printf("\n✨ %d patterns shared! You're helping developers worldwide.\n", shared)
		}
		if skipped > 0 {
			fmt.Printf("   %d patterns skipped due to detected secrets.\n", skipped)
		}
	}

	return nil
}
