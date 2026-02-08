package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/config"
	murhooks "github.com/mur-run/mur-core/internal/hooks"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose mur setup and fix common issues",
	Long: `Check mur configuration and diagnose common issues.

Checks:
  - Directory structure
  - AI CLI installations
  - Hook configurations
  - Pattern validity
  - Sync status

Examples:
  mur doctor         # Run all checks
  mur doctor --fix   # Auto-fix issues where possible`,
	RunE: runDoctor,
}

var doctorFix bool

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Auto-fix issues where possible")
}

type checkResult struct {
	name    string
	status  string // "ok", "warn", "error"
	message string
	fix     func() error
}

func runDoctor(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("🩺 mur doctor")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	var checks []checkResult
	var fixable []checkResult

	// Check 1: .mur directory
	murDir := filepath.Join(home, ".mur")
	if info, err := os.Stat(murDir); err != nil || !info.IsDir() {
		checks = append(checks, checkResult{
			name:    "~/.mur directory",
			status:  "error",
			message: "Directory not found",
			fix: func() error {
				return os.MkdirAll(murDir, 0755)
			},
		})
	} else {
		checks = append(checks, checkResult{
			name:   "~/.mur directory",
			status: "ok",
		})
	}

	// Check 2: patterns directory
	patternsDir := filepath.Join(murDir, "patterns")
	if info, err := os.Stat(patternsDir); err != nil || !info.IsDir() {
		checks = append(checks, checkResult{
			name:    "~/.mur/patterns",
			status:  "warn",
			message: "No patterns directory",
			fix: func() error {
				return os.MkdirAll(patternsDir, 0755)
			},
		})
	} else {
		files, _ := os.ReadDir(patternsDir)
		yamlCount := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".yaml") {
				yamlCount++
			}
		}
		checks = append(checks, checkResult{
			name:    "~/.mur/patterns",
			status:  "ok",
			message: fmt.Sprintf("%d patterns", yamlCount),
		})
	}

	// Check 3: AI CLIs
	clis := []struct {
		name   string
		binary string
	}{
		{"Claude Code", "claude"},
		{"Gemini CLI", "gemini"},
		{"Codex CLI", "codex"},
		{"Auggie", "auggie"},
		{"Aider", "aider"},
		{"OpenCode", "opencode"},
		{"GitHub Copilot", "gh"},
	}

	// Common installation paths to check (beyond PATH)
	extraPaths := []string{
		filepath.Join(home, ".npm-global", "bin"),
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".cargo", "bin"),
		"/usr/local/bin",
		"/opt/homebrew/bin",
	}

	installedCLIs := 0
	for _, cli := range clis {
		found := false
		foundPath := ""

		// First check PATH
		if path, err := exec.LookPath(cli.binary); err == nil {
			found = true
			foundPath = path
		} else {
			// Check extra paths
			for _, dir := range extraPaths {
				path := filepath.Join(dir, cli.binary)
				if _, err := os.Stat(path); err == nil {
					found = true
					foundPath = path
					break
				}
			}
		}

		if found {
			installedCLIs++
			// Check if in PATH
			if _, err := exec.LookPath(cli.binary); err != nil {
				checks = append(checks, checkResult{
					name:    cli.name,
					status:  "ok",
					message: fmt.Sprintf("Found at %s (not in PATH)", foundPath),
				})
			} else {
				checks = append(checks, checkResult{
					name:   cli.name,
					status: "ok",
				})
			}
		} else {
			checks = append(checks, checkResult{
				name:    cli.name,
				status:  "warn",
				message: "Not installed",
			})
		}
	}

	// Check 4: Claude hooks (in settings.json)
	claudeSettingsPath := filepath.Join(home, ".claude", "settings.json")
	if _, err := os.Stat(claudeSettingsPath); err == nil {
		content, _ := os.ReadFile(claudeSettingsPath)
		if strings.Contains(string(content), "mur") || strings.Contains(string(content), "on-prompt.sh") {
			checks = append(checks, checkResult{
				name:   "Claude hooks",
				status: "ok",
			})
		} else {
			checks = append(checks, checkResult{
				name:    "Claude hooks",
				status:  "warn",
				message: "Settings exist but mur not configured",
			})
		}
	} else {
		checks = append(checks, checkResult{
			name:    "Claude hooks",
			status:  "warn",
			message: "Not installed (run: mur init --hooks)",
		})
	}

	// Check 5: Gemini hooks (in settings.json)
	geminiSettingsPath := filepath.Join(home, ".gemini", "settings.json")
	if _, err := os.Stat(geminiSettingsPath); err == nil {
		content, _ := os.ReadFile(geminiSettingsPath)
		if strings.Contains(string(content), "mur") || strings.Contains(string(content), "on-prompt.sh") {
			checks = append(checks, checkResult{
				name:   "Gemini hooks",
				status: "ok",
			})
		} else {
			checks = append(checks, checkResult{
				name:    "Gemini hooks",
				status:  "warn",
				message: "Settings exist but mur not configured",
			})
		}
	} else {
		checks = append(checks, checkResult{
			name:    "Gemini hooks",
			status:  "warn",
			message: "Not installed (run: mur init --hooks)",
		})
	}

	// Check 5b: OpenCode hooks
	if installed, _ := murhooks.CheckOpenCodeHooks(); installed {
		checks = append(checks, checkResult{
			name:   "OpenCode hooks",
			status: "ok",
		})
	} else {
		checks = append(checks, checkResult{
			name:    "OpenCode hooks",
			status:  "warn",
			message: "Not installed (run: mur init --hooks)",
		})
	}

	// Check 5c: GitHub Copilot hooks
	if installed, _ := murhooks.CheckCopilotHooks(); installed {
		checks = append(checks, checkResult{
			name:   "Copilot hooks",
			status: "ok",
		})
	} else {
		checks = append(checks, checkResult{
			name:    "Copilot hooks",
			status:  "warn",
			message: "Not installed (run: mur init --hooks)",
		})
	}

	// Check 6: Sync targets
	syncTargets := []struct {
		name string
		path string
	}{
		{"Claude skills", filepath.Join(home, ".claude", "skills", "mur")},
		{"Gemini skills", filepath.Join(home, ".gemini", "skills", "mur")},
		{"Codex instructions", filepath.Join(home, ".codex", "instructions.md")},
		{"Continue rules", filepath.Join(home, ".continue", "rules", "mur")},
		{"Cursor rules", filepath.Join(home, ".cursor", "rules", "mur")},
	}

	syncedCount := 0
	for _, t := range syncTargets {
		if _, err := os.Stat(t.path); err == nil {
			syncedCount++
		}
	}

	if syncedCount == 0 {
		checks = append(checks, checkResult{
			name:    "Sync targets",
			status:  "warn",
			message: "No targets synced (run: mur sync)",
		})
	} else {
		checks = append(checks, checkResult{
			name:    "Sync targets",
			status:  "ok",
			message: fmt.Sprintf("%d/%d synced", syncedCount, len(syncTargets)),
		})
	}

	// Check 7: Ollama (for embeddings and LLM extraction)
	ollamaRunning := false
	if _, err := exec.LookPath("ollama"); err == nil {
		// Check if running
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://localhost:11434/api/tags")
		if err == nil {
			resp.Body.Close()
			ollamaRunning = true
			checks = append(checks, checkResult{
				name:    "Ollama",
				status:  "ok",
				message: "Running (available for LLM extraction)",
			})
		} else {
			checks = append(checks, checkResult{
				name:    "Ollama",
				status:  "warn",
				message: "Installed but not running (start with: ollama serve)",
			})
		}
	} else {
		checks = append(checks, checkResult{
			name:    "Ollama",
			status:  "info",
			message: "Not installed (optional, for local LLM)",
		})
	}

	// Check 8: LLM configuration for extraction
	cfg, _ := config.Load()
	llmConfigured := cfg != nil && cfg.Learning.LLM.Provider != ""
	if llmConfigured {
		msg := fmt.Sprintf("Configured: %s", cfg.Learning.LLM.Provider)
		if cfg.Learning.LLM.Model != "" {
			msg += "/" + cfg.Learning.LLM.Model
		}
		checks = append(checks, checkResult{
			name:    "LLM extraction",
			status:  "ok",
			message: msg,
		})
	} else if ollamaRunning {
		checks = append(checks, checkResult{
			name:    "LLM extraction",
			status:  "ok",
			message: "Will use Ollama (auto-detected)",
		})
	} else {
		checks = append(checks, checkResult{
			name:    "LLM extraction",
			status:  "warn",
			message: "Not configured (extraction quality limited)",
		})
	}

	// Check 9: Premium LLM configuration
	if cfg != nil && cfg.Learning.LLM.Premium != nil {
		p := cfg.Learning.LLM.Premium
		msg := fmt.Sprintf("%s", p.Provider)
		if p.Model != "" {
			msg += "/" + p.Model
		}
		// Check if API key is available
		if p.APIKeyEnv != "" {
			if os.Getenv(p.APIKeyEnv) != "" {
				msg += " ✓"
			} else {
				msg += fmt.Sprintf(" (missing %s)", p.APIKeyEnv)
			}
		}
		checks = append(checks, checkResult{
			name:    "Premium LLM",
			status:  "ok",
			message: msg,
		})

		// Show routing rules
		if cfg.Learning.LLM.Routing != nil {
			r := cfg.Learning.LLM.Routing
			var rules []string
			if r.MinMessages > 0 {
				rules = append(rules, fmt.Sprintf(">=%d msgs", r.MinMessages))
			}
			if len(r.Projects) > 0 {
				rules = append(rules, fmt.Sprintf("projects: %v", r.Projects))
			}
			if len(rules) > 0 {
				checks = append(checks, checkResult{
					name:    "Premium routing",
					status:  "ok",
					message: strings.Join(rules, ", "),
				})
			}
		}
	}

	// Print results
	errorCount := 0
	warnCount := 0

	for _, c := range checks {
		icon := "✓"
		switch c.status {
		case "ok":
			icon = "✅"
		case "warn":
			icon = "⚠️"
			warnCount++
			if c.fix != nil {
				fixable = append(fixable, c)
			}
		case "error":
			icon = "❌"
			errorCount++
			if c.fix != nil {
				fixable = append(fixable, c)
			}
		case "info":
			icon = "ℹ️"
		}

		if c.message != "" {
			fmt.Printf("%s %-20s %s\n", icon, c.name, c.message)
		} else {
			fmt.Printf("%s %s\n", icon, c.name)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if errorCount == 0 && warnCount == 0 {
		fmt.Println("✅ All checks passed!")
	} else {
		fmt.Printf("📊 %d errors, %d warnings\n", errorCount, warnCount)
	}

	// Auto-fix
	if doctorFix && len(fixable) > 0 {
		fmt.Println()
		fmt.Println("🔧 Attempting fixes...")
		for _, c := range fixable {
			if err := c.fix(); err != nil {
				fmt.Printf("   ❌ %s: %v\n", c.name, err)
			} else {
				fmt.Printf("   ✅ Fixed: %s\n", c.name)
			}
		}
	} else if len(fixable) > 0 {
		fmt.Println()
		fmt.Printf("💡 %d issues can be auto-fixed with: mur doctor --fix\n", len(fixable))
	}

	// Suggestions
	if warnCount > 0 || errorCount > 0 {
		fmt.Println()
		fmt.Println("💡 Suggestions:")
		if installedCLIs == 0 {
			fmt.Println("   • Install an AI CLI (claude, gemini, codex, auggie, aider)")
		}
		if syncedCount == 0 {
			fmt.Println("   • Run 'mur sync' to sync patterns to AI tools")
		}
		fmt.Println("   • Run 'mur init --hooks' to set up automatic learning")
	}

	fmt.Println()

	return nil
}
