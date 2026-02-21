package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/stats"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mur status overview",
	Long: `Display a quick overview of mur status including:
  - Pattern count and health
  - Sync status for all targets
  - Recent usage statistics
  - Configuration status

Examples:
  mur status           # Quick overview
  mur status --verbose # Detailed status`,
	RunE: runStatus,
}

var statusVerbose bool

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "V", false, "Show detailed status")
}

func runStatus(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🔮 mur status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Patterns
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)
	patterns, _ := store.List()

	activeCount := 0
	deprecatedCount := 0
	var totalEffectiveness float64
	effectiveCount := 0
	var totalUsage int

	for _, p := range patterns {
		if p.Lifecycle.Status == "deprecated" {
			deprecatedCount++
		} else {
			activeCount++
		}
		if p.Learning.Effectiveness > 0 {
			totalEffectiveness += p.Learning.Effectiveness
			effectiveCount++
		}
		totalUsage += p.Learning.UsageCount
	}

	avgEffectiveness := 0.0
	if effectiveCount > 0 {
		avgEffectiveness = totalEffectiveness / float64(effectiveCount) * 100
	}

	fmt.Println()
	fmt.Println("📚 Patterns")
	fmt.Printf("   Total: %d (%d active, %d deprecated)\n", len(patterns), activeCount, deprecatedCount)
	fmt.Printf("   Usage: %d injections\n", totalUsage)
	if effectiveCount > 0 {
		fmt.Printf("   Avg Effectiveness: %.0f%%\n", avgEffectiveness)
	}

	// Cloud status
	fmt.Println()
	fmt.Println("☁️  Cloud")
	authStore, authErr := cloud.NewAuthStore()
	authData, _ := authStore.Load()

	if authErr == nil && authStore.IsLoggedIn() {
		if authData != nil && authData.User != nil {
			fmt.Printf("   Logged in as: %s\n", authData.User.Email)
		} else {
			fmt.Println("   Logged in (API key)")
		}

		// Show trial status by calling /me API
		cfg, _ := config.Load()
		if client, err := cloud.NewClient(func() string {
			if cfg != nil {
				return cfg.Server.URL
			}
			return ""
		}()); err == nil {
			if me, err := client.Me(); err == nil && me.Plan == "trial" {
				if me.TrialDaysRemaining > 14 {
					fmt.Printf("   Trial: %d days remaining\n", me.TrialDaysRemaining)
				} else if me.TrialDaysRemaining > 0 {
					fmt.Printf("   ⚠️  Trial: %d days remaining! Upgrade: mur billing | Extend: mur referral\n", me.TrialDaysRemaining)
				} else {
					fmt.Println("   ⚠️  Trial expired — Free plan (cloud sync disabled)")
					fmt.Println("   Upgrade: app.mur.run/billing | Extend: mur referral")
				}
			}
		}

		// Show active team from config
		if cfg != nil && cfg.Server.Team != "" {
			fmt.Printf("   Active team: %s\n", cfg.Server.Team)
		}

		// Show last sync time
		syncStatePath := filepath.Join(home, ".mur", "sync-state.yaml")
		if info, err := os.Stat(syncStatePath); err == nil {
			syncAge := time.Since(info.ModTime())
			var syncAgeStr string
			if syncAge < time.Minute {
				syncAgeStr = "just now"
			} else if syncAge < time.Hour {
				syncAgeStr = fmt.Sprintf("%d min ago", int(syncAge.Minutes()))
			} else if syncAge < 24*time.Hour {
				syncAgeStr = fmt.Sprintf("%d hours ago", int(syncAge.Hours()))
			} else {
				syncAgeStr = fmt.Sprintf("%d days ago", int(syncAge.Hours()/24))
			}
			fmt.Printf("   Last sync: %s\n", syncAgeStr)
		}

		if statusVerbose {
			fmt.Println("   Commands: mur cloud teams, mur cloud sync")
		}
	} else if authData != nil && authData.AccessToken != "" {
		// Has token but expired
		fmt.Println("   ⚠️  Session expired")
		fmt.Println("   Run: mur login")
	} else {
		fmt.Println("   Not logged in")
		fmt.Println("   Run: mur login")
	}

	// Sync targets
	fmt.Println()
	fmt.Println("🔄 Sync Targets")

	type syncTarget struct {
		name string
		path string
		icon string
	}

	targets := []syncTarget{
		{"Claude Code", filepath.Join(home, ".claude", "skills", "mur"), "⌨️"},
		{"Gemini CLI", filepath.Join(home, ".gemini", "skills", "mur"), "⌨️"},
		{"Codex CLI", filepath.Join(home, ".codex", "instructions.md"), "⌨️"},
		{"Auggie", filepath.Join(home, ".augment", "skills", "mur"), "⌨️"},
		{"Aider", filepath.Join(home, ".aider", "mur-patterns.md"), "⌨️"},
		{"Continue", filepath.Join(home, ".continue", "rules", "mur"), "🖥️"},
		{"Cursor", filepath.Join(home, ".cursor", "rules", "mur"), "🖥️"},
		{"Windsurf", filepath.Join(home, ".windsurf", "rules", "mur"), "🖥️"},
	}

	syncedCount := 0
	for _, t := range targets {
		info, err := os.Stat(t.path)
		if err == nil {
			syncedCount++
			if statusVerbose {
				fileCount := 1
				if info.IsDir() {
					files, _ := os.ReadDir(t.path)
					fileCount = len(files)
				}
				lastMod := info.ModTime().Format("Jan 2 15:04")
				fmt.Printf("   %s %-12s ✓ %d files, %s\n", t.icon, t.name, fileCount, lastMod)
			}
		} else if statusVerbose {
			fmt.Printf("   %s %-12s ✗ not synced\n", t.icon, t.name)
		}
	}

	if !statusVerbose {
		fmt.Printf("   %d/%d targets synced\n", syncedCount, len(targets))
		fmt.Println("   Run with --verbose for details")
	}

	// Usage stats
	records, _ := stats.Query(stats.QueryFilter{
		StartTime: time.Now().AddDate(0, 0, -7),
	})

	if len(records) > 0 {
		summary := stats.Summarize(records)
		fmt.Println()
		fmt.Println("📊 Last 7 Days")
		fmt.Printf("   Runs: %d\n", summary.TotalRuns)
		if summary.EstimatedCost > 0 {
			fmt.Printf("   Cost: $%.4f\n", summary.EstimatedCost)
		}
		if summary.EstimatedSaved > 0 {
			fmt.Printf("   Saved: $%.4f (via free tools)\n", summary.EstimatedSaved)
		}
		if summary.AutoRouteStats.Total > 0 {
			fmt.Printf("   Auto-routed: %d (%.0f%% to free)\n",
				summary.AutoRouteStats.Total, summary.AutoRouteStats.FreeRatio)
		}

		// Tool breakdown in verbose mode
		if statusVerbose && len(summary.ByTool) > 0 {
			fmt.Println()
			fmt.Println("   By Tool:")
			for tool, ts := range summary.ByTool {
				fmt.Printf("   • %s: %d runs, $%.4f\n", tool, ts.Count, ts.TotalCost)
			}
		}
	}

	// Config status
	fmt.Println()
	fmt.Println("⚙️  Config")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("   ⚠️  No config found (using defaults)")
	} else {
		// Count enabled tools
		enabledTools := 0
		for _, tool := range cfg.Tools {
			if tool.Enabled {
				enabledTools++
			}
		}
		fmt.Printf("   %d tools configured\n", enabledTools)

		if statusVerbose {
			for name, tool := range cfg.Tools {
				status := "✗"
				if tool.Enabled {
					status = "✓"
				}
				fmt.Printf("   %s %s (%s)\n", status, name, tool.Tier)
			}
		}
	}

	// Repo status
	repoPath := filepath.Join(home, ".mur", "repo")
	if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
		fmt.Println()
		fmt.Println("📦 Learning Repo")
		// Try to get remote URL
		remoteFile := filepath.Join(repoPath, ".git", "config")
		if content, err := os.ReadFile(remoteFile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.Contains(line, "url = ") {
					url := strings.TrimSpace(strings.TrimPrefix(line, "\turl = "))
					fmt.Printf("   %s\n", url)
					break
				}
			}
		}
	}

	// Hooks status
	fmt.Println()
	fmt.Println("🪝 Hooks")
	hookPaths := []struct {
		name string
		path string
	}{
		{"Claude Code", filepath.Join(home, ".claude", "hooks.json")},
		{"Gemini CLI", filepath.Join(home, ".gemini", "hooks.json")},
	}

	hooksInstalled := 0
	for _, h := range hookPaths {
		if _, err := os.Stat(h.path); err == nil {
			hooksInstalled++
			if statusVerbose {
				fmt.Printf("   ✓ %s\n", h.name)
			}
		} else if statusVerbose {
			fmt.Printf("   ✗ %s (not installed)\n", h.name)
		}
	}

	if !statusVerbose {
		if hooksInstalled > 0 {
			fmt.Printf("   %d/%d CLI hooks installed\n", hooksInstalled, len(hookPaths))
		} else {
			fmt.Println("   No hooks installed")
			fmt.Println("   Run: mur init --hooks")
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Dashboard: mur serve")
	fmt.Println("Help: mur --help")
	fmt.Println()

	return nil
}
