package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/config"
	murhooks "github.com/mur-run/mur-core/internal/hooks"
	"github.com/mur-run/mur-core/internal/sync"
)

var (
	initNonInteractive bool
	initHooks          bool
	initSearchHooks    bool
	initForce          bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mur configuration",
	Long: `Initialize mur with an interactive setup wizard.

Examples:
  mur init          # Interactive: choose CLIs, configure hooks, set up repo
  mur init --hooks  # Quick: install hooks with defaults (non-interactive)

The --hooks flag is a shortcut for quick setup. It installs Claude Code
and Gemini CLI hooks using default settings. Use plain 'mur init' for
full control over configuration.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Skip interactive prompts, use defaults")
	initCmd.Flags().BoolVar(&initHooks, "hooks", false, "Quick setup: install hooks with defaults (implies --non-interactive)")
	initCmd.Flags().BoolVar(&initSearchHooks, "search", true, "Enable search hooks (suggest patterns on prompt)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Force overwrite existing config (ignore existing settings)")
}

// CLI tool configuration
type cliTool struct {
	Name         string
	Binary       string
	Installed    bool
	HooksSupport bool
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	murDir := filepath.Join(home, ".mur")

	// --hooks implies --non-interactive
	if initHooks {
		initNonInteractive = true
	}

	// Non-interactive mode
	if initNonInteractive {
		return runNonInteractiveInit(home, murDir)
	}

	// Interactive mode
	return runInteractiveInit(home, murDir)
}

func runInteractiveInit(home, murDir string) error {
	fmt.Println()
	fmt.Println("🚀 Welcome to mur!")
	fmt.Println()

	// Detect installed CLIs
	tools := detectCLIs()

	// Show detected tools
	var installedNames []string
	for _, t := range tools {
		if t.Installed {
			installedNames = append(installedNames, t.Name)
		}
	}

	if len(installedNames) > 0 {
		fmt.Printf("Detected AI CLIs: %s\n", strings.Join(installedNames, ", "))
		fmt.Println()
	}

	// Select which CLIs to use
	var selectedCLIs []string
	cliOptions := []string{}
	for _, t := range tools {
		status := ""
		if t.Installed {
			status = " (installed)"
		}
		cliOptions = append(cliOptions, t.Name+status)
	}

	cliPrompt := &survey.MultiSelect{
		Message: "Which AI CLIs do you want to use?",
		Options: cliOptions,
		Default: installedNames,
	}
	if err := survey.AskOne(cliPrompt, &selectedCLIs); err != nil {
		return err
	}

	// Clean up selection (remove " (installed)" suffix)
	for i, s := range selectedCLIs {
		selectedCLIs[i] = strings.TrimSuffix(s, " (installed)")
	}

	// Check if Claude is selected and ask about hooks
	installHooks := false
	claudeSelected := contains(selectedCLIs, "Claude Code")
	if claudeSelected {
		hookPrompt := &survey.Confirm{
			Message: "Install Claude Code hooks for real-time learning?",
			Default: true,
		}
		if err := survey.AskOne(hookPrompt, &installHooks); err != nil {
			return err
		}
	}

	// Ask for default CLI
	defaultCLI := ""
	if len(selectedCLIs) > 0 {
		defaultPrompt := &survey.Select{
			Message: "Which CLI should be the default?",
			Options: selectedCLIs,
			Default: selectedCLIs[0],
		}
		if err := survey.AskOne(defaultPrompt, &defaultCLI); err != nil {
			return err
		}
	}

	// Model setup
	fmt.Println()
	models, err := askModelSetup()
	if err != nil {
		return err
	}

	// Create directories
	fmt.Println()
	dirs := []string{
		murDir,
		filepath.Join(murDir, "patterns"),
		filepath.Join(murDir, "hooks"),
		filepath.Join(murDir, "transcripts"),
		filepath.Join(murDir, "tracking"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	fmt.Println("✓ Created ~/.mur/ directory")

	// Create config
	if err := createConfigWithModels(murDir, selectedCLIs, defaultCLI, models); err != nil {
		return err
	}
	fmt.Println("✓ Created config.yaml")

	// Install hooks if requested
	if installHooks {
		if err := installClaudeHooks(home, murDir); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	}

	// Ask about learning repo
	fmt.Println()
	if err := SetupLearningRepo(home); err != nil {
		fmt.Printf("  ⚠ Warning: %v\n", err)
	}

	// Sync patterns to all selected CLIs
	fmt.Println()
	fmt.Println("Syncing patterns to CLIs...")
	results, err := sync.SyncPatternsToAllCLIs()
	if err != nil {
		fmt.Printf("  ⚠ Warning: %v\n", err)
	} else {
		for _, r := range results {
			if r.Success {
				fmt.Printf("  ✓ %s: %s\n", r.Target, r.Message)
			}
		}
	}

	// Final message
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ mur is ready!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  mur run -p \"your first task\"")
	if defaultCLI != "" {
		fmt.Printf("  # uses %s\n", defaultCLI)
	} else {
		fmt.Println()
	}
	fmt.Println("  mur stats                          # see your progress")
	fmt.Println()
	fmt.Println("Optional: Enable semantic search for smarter pattern matching:")
	fmt.Println("  ollama pull nomic-embed-text       # install local embeddings")
	fmt.Println("  mur embed index                    # index your patterns")
	fmt.Println()

	return nil
}

func runNonInteractiveInit(home, murDir string) error {
	// Create directories
	dirs := []string{
		murDir,
		filepath.Join(murDir, "patterns"),
		filepath.Join(murDir, "hooks"),
		filepath.Join(murDir, "transcripts"),
		filepath.Join(murDir, "tracking"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Check if config exists
	configPath := filepath.Join(murDir, "config.yaml")
	configExists := fileExists(configPath)

	if configExists && !initForce {
		// Existing config - merge new fields and migrate
		existing, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}

		oldVersion := existing.SchemaVersion
		defaults := config.Default()
		merged := config.MergeConfig(existing, defaults)

		changed, changes := config.MigrateConfig(merged)
		if changed {
			fmt.Printf("✓ Config migrated: v%d → v%d\n", oldVersion, merged.SchemaVersion)
			for _, c := range changes {
				fmt.Printf("  + Added: %s (%s)\n", c.Field, c.Description)
			}
		}

		if err := merged.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("✓ Config updated (preserved your settings)")
	} else {
		// First time or force - create new config
		if err := createConfig(murDir, []string{"Claude Code"}, "Claude Code"); err != nil {
			return err
		}
		if initForce && configExists {
			fmt.Println("✓ Config overwritten (--force)")
		} else {
			fmt.Println("✓ mur initialized at ~/.mur (using defaults)")
		}
	}

	// Install hooks if flag set
	if initHooks {
		if err := installClaudeHooks(home, murDir); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	}

	fmt.Println()
	if initHooks {
		fmt.Println("You're all set! Use claude or gemini directly — patterns auto-inject.")
		fmt.Println()
		fmt.Println("Tip: Run 'mur init' (without --hooks) for interactive setup with more options.")
	} else {
		fmt.Println("Next: mur run -p \"your task\"")
	}

	return nil
}

func detectCLIs() []cliTool {
	tools := []cliTool{
		{Name: "Claude Code", Binary: "claude", HooksSupport: true},
		{Name: "Gemini CLI", Binary: "gemini", HooksSupport: true},
		{Name: "Codex", Binary: "codex", HooksSupport: false},
		{Name: "Auggie", Binary: "auggie", HooksSupport: true},
		{Name: "Aider", Binary: "aider", HooksSupport: false},
	}

	for i := range tools {
		_, err := exec.LookPath(tools[i].Binary)
		tools[i].Installed = err == nil
	}

	return tools
}

// modelSetup holds the user's model provider choices.
type modelSetup struct {
	Mode           string // "cloud", "local", "custom"
	LLMProvider    string // "openai", "gemini", "claude", "ollama"
	LLMModel       string
	LLMAPIKeyEnv   string
	EmbedProvider  string // "openai", "ollama", "google", "voyage"
	EmbedModel     string
	EmbedAPIKeyEnv string
	EmbedMinScore  string
	OllamaURL      string
}

func defaultCloudSetup() modelSetup {
	return modelSetup{
		Mode:           "cloud",
		LLMProvider:    "openai",
		LLMModel:       "gpt-4o-mini",
		LLMAPIKeyEnv:   "OPENAI_API_KEY",
		EmbedProvider:  "openai",
		EmbedModel:     "text-embedding-3-small",
		EmbedAPIKeyEnv: "OPENAI_API_KEY",
		EmbedMinScore:  "0.3",
		OllamaURL:      "http://localhost:11434",
	}
}

func defaultLocalSetup() modelSetup {
	return modelSetup{
		Mode:          "local",
		LLMProvider:   "ollama",
		LLMModel:      "llama3.2:3b",
		EmbedProvider: "ollama",
		EmbedModel:    "qwen3-embedding",
		EmbedMinScore: "0.5",
		OllamaURL:     "http://localhost:11434",
	}
}

func askModelSetup() (modelSetup, error) {
	fmt.Println("📦 Model Setup")
	fmt.Println("  mur uses AI models for pattern search and extraction.")
	fmt.Println()

	var mode string
	modePrompt := &survey.Select{
		Message: "Choose setup mode:",
		Options: []string{
			"☁️  Cloud (recommended) - API keys, best quality, ~$0.02/month",
			"🏠 Local - Ollama, free, needs ~2.7GB disk",
			"🔧 Custom - pick providers individually",
		},
		Default: "☁️  Cloud (recommended) - API keys, best quality, ~$0.02/month",
	}
	if err := survey.AskOne(modePrompt, &mode); err != nil {
		return modelSetup{}, fmt.Errorf("setup cancelled")
	}

	if strings.HasPrefix(mode, "🏠") {
		fmt.Println()
		fmt.Println("  Models needed: qwen3-embedding (639MB) + llama3.2:3b (2GB)")
		fmt.Println("  Install with: ollama pull qwen3-embedding && ollama pull llama3.2:3b")
		return defaultLocalSetup(), nil
	}

	if strings.HasPrefix(mode, "☁️") {
		return askCloudSetup()
	}

	return askCustomSetup()
}

func askCloudSetup() (modelSetup, error) {
	m := defaultCloudSetup()

	fmt.Println()
	fmt.Println("  Using OpenAI for both embedding and extraction.")
	fmt.Println("  Estimated cost: ~$0.02/month for typical usage.")
	fmt.Println()

	// Check if OPENAI_API_KEY is set
	if os.Getenv("OPENAI_API_KEY") != "" {
		fmt.Println("  ✅ OPENAI_API_KEY detected in environment")
	} else {
		fmt.Println("  ⚠️  Set OPENAI_API_KEY in your shell profile:")
		fmt.Println("     export OPENAI_API_KEY=sk-...")
	}

	// Ask if they want a different provider
	var wantDifferent bool
	diffPrompt := &survey.Confirm{
		Message: "Use a different provider? (Gemini, Claude, etc.)",
		Default: false,
	}
	if err := survey.AskOne(diffPrompt, &wantDifferent); err != nil {
		return modelSetup{}, fmt.Errorf("setup cancelled")
	}
	if !wantDifferent {
		return m, nil
	}

	return askCustomSetup()
}

func askCustomSetup() (modelSetup, error) {
	m := defaultCloudSetup()

	fmt.Println()

	// Embedding provider
	var embedChoice string
	embedPrompt := &survey.Select{
		Message: "Embedding provider (for search):",
		Options: []string{
			"OpenAI text-embedding-3-small ($0.02/1M tokens, recommended)",
			"Google text-embedding-004 (free tier available)",
			"Voyage voyage-3-large ($0.18/1M, best for code)",
			"Ollama mxbai-embed-large (free, local, 669MB)",
		},
	}
	if err := survey.AskOne(embedPrompt, &embedChoice); err != nil {
		return modelSetup{}, fmt.Errorf("setup cancelled")
	}

	switch {
	case strings.HasPrefix(embedChoice, "OpenAI"):
		m.EmbedProvider = "openai"
		m.EmbedModel = "text-embedding-3-small"
		m.EmbedAPIKeyEnv = "OPENAI_API_KEY"
		m.EmbedMinScore = "0.3"
	case strings.HasPrefix(embedChoice, "Google"):
		m.EmbedProvider = "google"
		m.EmbedModel = "text-embedding-004"
		m.EmbedAPIKeyEnv = "GEMINI_API_KEY"
		m.EmbedMinScore = "0.3"
	case strings.HasPrefix(embedChoice, "Voyage"):
		m.EmbedProvider = "voyage"
		m.EmbedModel = "voyage-3-large"
		m.EmbedAPIKeyEnv = "VOYAGE_API_KEY"
		m.EmbedMinScore = "0.3"
	case strings.HasPrefix(embedChoice, "Ollama"):
		m.EmbedProvider = "ollama"
		m.EmbedModel = "qwen3-embedding"
		m.EmbedMinScore = "0.5"
	}

	// LLM provider
	var llmChoice string
	llmPrompt := &survey.Select{
		Message: "LLM provider (for extraction & expansion):",
		Options: []string{
			"OpenAI gpt-4o-mini ($0.15/$0.60 per 1M, recommended)",
			"Gemini gemini-2.0-flash ($0.10/$0.40 per 1M, cheapest)",
			"Claude claude-haiku ($0.25/$1.25 per 1M, best quality)",
			"Ollama llama3.2:3b (free, local, 2GB)",
		},
	}
	if err := survey.AskOne(llmPrompt, &llmChoice); err != nil {
		return modelSetup{}, fmt.Errorf("setup cancelled")
	}

	switch {
	case strings.HasPrefix(llmChoice, "OpenAI"):
		m.LLMProvider = "openai"
		m.LLMModel = "gpt-4o-mini"
		m.LLMAPIKeyEnv = "OPENAI_API_KEY"
	case strings.HasPrefix(llmChoice, "Gemini"):
		m.LLMProvider = "gemini"
		m.LLMModel = "gemini-2.0-flash"
		m.LLMAPIKeyEnv = "GEMINI_API_KEY"
	case strings.HasPrefix(llmChoice, "Claude"):
		m.LLMProvider = "claude"
		m.LLMModel = "claude-haiku"
		m.LLMAPIKeyEnv = "ANTHROPIC_API_KEY"
	case strings.HasPrefix(llmChoice, "Ollama"):
		m.LLMProvider = "ollama"
		m.LLMModel = "llama3.2:3b"
	}

	return m, nil
}

func (m modelSetup) llmYaml() string {
	switch m.LLMProvider {
	case "ollama":
		return fmt.Sprintf("    provider: ollama\n    model: %s\n    ollama_url: %s", m.LLMModel, m.OllamaURL)
	default:
		return fmt.Sprintf("    provider: %s\n    model: %s\n    api_key_env: %s", m.LLMProvider, m.LLMModel, m.LLMAPIKeyEnv)
	}
}

func (m modelSetup) searchYaml() string {
	switch m.EmbedProvider {
	case "ollama":
		return fmt.Sprintf("  provider: ollama\n  model: %s\n  ollama_url: %s\n  min_score: %s", m.EmbedModel, m.OllamaURL, m.EmbedMinScore)
	default:
		return fmt.Sprintf("  provider: %s\n  model: %s\n  api_key_env: %s\n  min_score: %s", m.EmbedProvider, m.EmbedModel, m.EmbedAPIKeyEnv, m.EmbedMinScore)
	}
}

func createConfig(murDir string, selectedCLIs []string, defaultCLI string) error {
	return createConfigWithModels(murDir, selectedCLIs, defaultCLI, defaultLocalSetup())
}

func createConfigWithModels(murDir string, selectedCLIs []string, defaultCLI string, models modelSetup) error {
	configPath := filepath.Join(murDir, "config.yaml")

	// Map CLI names to config keys
	cliMap := map[string]string{
		"Claude Code": "claude",
		"Gemini CLI":  "gemini",
		"Codex":       "codex",
		"Auggie":      "auggie",
		"Aider":       "aider",
	}

	defaultKey := "claude"
	if key, ok := cliMap[defaultCLI]; ok {
		defaultKey = key
	}

	// Build tools section
	toolsYaml := ""
	for name, key := range cliMap {
		enabled := contains(selectedCLIs, name)
		toolsYaml += fmt.Sprintf("  %s:\n    enabled: %t\n    binary: %s\n", key, enabled, key)
	}

	config := fmt.Sprintf(`# mur Configuration
# https://github.com/mur-run/mur-core

schema_version: 2

# Default AI CLI
default_tool: %s

# Available tools
tools:
%s
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧠 Learning & Pattern Extraction
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# LLM used for: pattern extraction from transcripts,
# search query expansion (mur index rebuild --expand).
#
# 💡 Cost: ~$0.02 for 200 patterns. Paid models extract
#    2-3x more patterns with better quality.
#
# Providers: ollama (free/local) | openai | gemini | claude
learning:
  auto_extract: true
  sync_to_tools: true
  llm:
%s
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔍 Semantic Search
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Embedding model for pattern search and auto-injection.
#
# 💡 Cost: ~$0.001 for 200 patterns. OpenAI gives ~15%%
#    better search quality vs local models.
#
# Providers: ollama (free/local) | openai | voyage | google
search:
  enabled: true
%s
  top_k: 3
  auto_inject: true

# Pattern Consolidation
consolidation:
  enabled: true
  schedule: weekly
  auto_archive: false           # Enable after accumulating patterns
  auto_merge: keep-best
  merge_threshold: 0.85
  decay_half_life_days: 90
  grace_period_days: 14
  min_patterns_before_run: 50
  notify_on_run: false

# Community Sharing
community:
  share_enabled: true
  auto_share_on_push: true

# Routing
routing:
  mode: auto  # auto | manual | cost-first

# Notifications (optional)
# notifications:
#   enabled: false
#   slack:
#     webhook_url: ""
#     channel: ""
#   discord:
#     webhook_url: ""

# Custom Hooks (optional)
# hooks:
#   UserPromptSubmit: []
#   Stop: []
#   BeforeTool: []
#   AfterTool: []
`, defaultKey, toolsYaml, models.llmYaml(), models.searchYaml())

	return os.WriteFile(configPath, []byte(config), 0644)
}

func installClaudeHooks(home, murDir string) error {
	// Load config to check search settings
	cfg, _ := config.Load()
	searchEnabled := cfg != nil && cfg.Search.IsEnabled() && cfg.Search.IsAutoInject()

	// Create on-prompt.sh - injects context-aware patterns
	promptScriptPath := filepath.Join(murDir, "hooks", "on-prompt.sh")
	promptScript := `#!/bin/bash
# Inject context-aware patterns based on current project
mur context --compact 2>/dev/null || true
`
	if err := os.WriteFile(promptScriptPath, []byte(promptScript), 0755); err != nil {
		return err
	}

	// Create on-prompt-reminder.md
	reminderPath := filepath.Join(murDir, "hooks", "on-prompt-reminder.md")
	reminderContent := `[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a pattern), save it:

  mur learn add --name "pattern-name" --content "description"

Or create a file in ~/.mur/patterns/

Only save if: it required discovery, it helps future tasks, and it's verified.
`
	if err := os.WriteFile(reminderPath, []byte(reminderContent), 0644); err != nil {
		return err
	}

	// Create on-stop.sh
	stopScriptPath := filepath.Join(murDir, "hooks", "on-stop.sh")
	stopScript := fmt.Sprintf(`#!/bin/bash
# mur-managed-hook v%d
# Lightweight sync (blocking, fast)
mur sync --quiet 2>/dev/null || true

# LLM extract in background (non-blocking)
(mur learn extract --llm --auto --accept-all --quiet 2>/dev/null &) || true

# Load user customizations if they exist
[ -f ~/.mur/hooks/on-stop.local.sh ] && source ~/.mur/hooks/on-stop.local.sh
`, murhooks.CurrentHookVersion)
	if err := os.WriteFile(stopScriptPath, []byte(stopScript), 0755); err != nil {
		return err
	}

	// Update Claude settings
	claudeSettingsPath := filepath.Join(home, ".claude", "settings.json")

	// Build UserPromptSubmit hooks
	promptHooks := []map[string]interface{}{
		// Inject context-aware patterns
		{"type": "command", "command": fmt.Sprintf("bash %s >&2", promptScriptPath)},
		// Learning reminder
		{"type": "command", "command": fmt.Sprintf("cat %s >&2", reminderPath)},
	}

	// Add semantic search hook if enabled
	if searchEnabled {
		promptHooks = append(promptHooks, map[string]interface{}{
			"type":    "command",
			"command": `mur search --inject "$PROMPT" 2>/dev/null || true`,
		})
		fmt.Println("  + Added semantic search hook (auto-inject enabled)")
	}

	hooks := map[string]interface{}{
		"UserPromptSubmit": []map[string]interface{}{
			{
				"matcher": "",
				"hooks":   promptHooks,
			},
		},
		"Stop": []map[string]interface{}{
			{
				"matcher": "",
				"hooks": []map[string]interface{}{
					{"type": "command", "command": fmt.Sprintf("bash %s", stopScriptPath)},
				},
			},
		},
	}

	var settings map[string]interface{}
	if data, err := os.ReadFile(claudeSettingsPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	} else {
		_ = os.MkdirAll(filepath.Join(home, ".claude"), 0755)
		settings = make(map[string]interface{})
	}

	// Backup
	if _, err := os.Stat(claudeSettingsPath); err == nil {
		backupPath := claudeSettingsPath + ".backup"
		if data, err := os.ReadFile(claudeSettingsPath); err == nil {
			_ = os.WriteFile(backupPath, data, 0644)
		}
	}

	// Set mur hooks (overwrites any existing hooks to ensure correct paths)
	settings["hooks"] = hooks

	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(claudeSettingsPath, data, 0644); err != nil {
		return err
	}

	fmt.Println("✓ Installed Claude Code hooks")

	// Install Gemini CLI hooks (v0.26.0+)
	if err := installGeminiHooks(home, promptScriptPath, stopScriptPath); err != nil {
		// Non-fatal, Gemini might not be installed
		fmt.Printf("  ⚠ Gemini hooks: %v\n", err)
	}

	return nil
}

func installGeminiHooks(home, promptScriptPath, stopScriptPath string) error {
	geminiSettingsPath := filepath.Join(home, ".gemini", "settings.json")

	// Gemini uses different event names
	hooks := map[string]interface{}{
		"BeforeAgent": []map[string]interface{}{
			{
				"matcher": "",
				"hooks": []map[string]interface{}{
					// Inject context-aware patterns
					{"type": "command", "command": fmt.Sprintf("bash %s", promptScriptPath)},
				},
			},
		},
		"SessionEnd": []map[string]interface{}{
			{
				"matcher": "",
				"hooks": []map[string]interface{}{
					{"type": "command", "command": fmt.Sprintf("bash %s", stopScriptPath)},
				},
			},
		},
	}

	var settings map[string]interface{}
	if data, err := os.ReadFile(geminiSettingsPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	} else {
		_ = os.MkdirAll(filepath.Join(home, ".gemini"), 0755)
		settings = make(map[string]interface{})
	}

	// Backup
	if _, err := os.Stat(geminiSettingsPath); err == nil {
		backupPath := geminiSettingsPath + ".backup"
		if data, err := os.ReadFile(geminiSettingsPath); err == nil {
			_ = os.WriteFile(backupPath, data, 0644)
		}
	}

	// Set mur hooks (overwrites any existing hooks)
	settings["hooks"] = hooks

	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(geminiSettingsPath, data, 0644); err != nil {
		return err
	}

	fmt.Println("✓ Installed Gemini CLI hooks")

	// Ask about community sharing after all hooks are installed
	askCommunitySharing()

	// Install Claude Code hooks
	if murhooks.ClaudeCodeInstalled() {
		if err := murhooks.InstallClaudeCodeHooks(initSearchHooks); err != nil {
			fmt.Printf("  ⚠ Claude Code hooks: %v\n", err)
		} else {
			if initSearchHooks {
				fmt.Println("✓ Installed Claude Code hooks (learn + search)")
			} else {
				fmt.Println("✓ Installed Claude Code hooks (learn)")
			}
		}
	}

	// Install OpenCode hooks
	if err := murhooks.InstallOpenCodeHooks(); err != nil {
		fmt.Printf("  ⚠ OpenCode hooks: %v\n", err)
	} else {
		fmt.Println("✓ Installed OpenCode hooks")
	}

	// Install GitHub Copilot hooks
	if err := murhooks.InstallCopilotHooks(); err != nil {
		fmt.Printf("  ⚠ GitHub Copilot hooks: %v\n", err)
	} else {
		fmt.Println("✓ Installed GitHub Copilot hooks")
	}

	// Install Auggie (Augment CLI) hooks
	if err := murhooks.InstallAuggieHooks(); err != nil {
		fmt.Printf("  ⚠ Auggie hooks: %v\n", err)
	} else {
		fmt.Println("✓ Installed Auggie hooks")
	}

	// Install OpenClaw hooks
	if err := murhooks.InstallOpenClawHooks(); err != nil {
		fmt.Printf("  ⚠ OpenClaw hooks: %v\n", err)
	} else {
		fmt.Println("✓ Installed OpenClaw hooks")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// askCommunitySharing prompts user about community pattern sharing.
func askCommunitySharing() {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("🌍 Community Sharing")
	fmt.Println()
	fmt.Println("Share your patterns with 10,000+ developers worldwide!")
	fmt.Println()
	fmt.Println("  • Your patterns help others solve problems faster")
	fmt.Println("  • Get ⭐ recognition and a public profile")
	fmt.Println("  • AI tools improve for everyone")
	fmt.Println()
	fmt.Println("🔒 All patterns are scanned for secrets before sharing.")
	fmt.Println("   API keys, passwords, tokens are automatically blocked.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enable community sharing? [Y/n]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	enabled := input == "" || input == "y" || input == "yes"

	// Update config file without destroying template comments
	// Use sed-style replacement on the YAML key
	configPath, err := config.ConfigPath()
	if err == nil {
		content, readErr := os.ReadFile(configPath)
		if readErr == nil {
			s := string(content)
			// The template already has a community section placeholder
			// Just update the values in-place
			s = strings.ReplaceAll(s, "share_enabled: true", fmt.Sprintf("share_enabled: %t", enabled))
			s = strings.ReplaceAll(s, "share_enabled: false", fmt.Sprintf("share_enabled: %t", enabled))
			s = strings.ReplaceAll(s, "auto_share_on_push: true", fmt.Sprintf("auto_share_on_push: %t", enabled))
			s = strings.ReplaceAll(s, "auto_share_on_push: false", fmt.Sprintf("auto_share_on_push: %t", enabled))
			_ = os.WriteFile(configPath, []byte(s), 0644)
		}
	}

	fmt.Println()
	if enabled {
		fmt.Println("✓ Community sharing enabled!")
		fmt.Println("  Change anytime: mur config set community.share_enabled false")
	} else {
		fmt.Println("✓ Community sharing disabled.")
		fmt.Println("  Enable anytime: mur config set community.share_enabled true")
	}
	fmt.Println()
}
