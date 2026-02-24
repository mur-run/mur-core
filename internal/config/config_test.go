package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	// Create temp config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".mur")
	_ = os.MkdirAll(configDir, 0755)

	configContent := `default_tool: gemini
tools:
  claude:
    enabled: true
    binary: claude
    flags: ["-p"]
  gemini:
    enabled: true
    binary: gemini
    flags: ["-p"]
`
	configPath := filepath.Join(configDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte(configContent), 0644)

	// Override home for test
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DefaultTool != "gemini" {
		t.Errorf("DefaultTool = %q, want %q", cfg.DefaultTool, "gemini")
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("len(Tools) = %d, want 2", len(cfg.Tools))
	}
}

func TestLoadMissing(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, should return default config", err)
	}

	// Should return default config
	if cfg.DefaultTool != "claude" {
		t.Errorf("DefaultTool = %q, want default %q", cfg.DefaultTool, "claude")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg := &Config{
		DefaultTool: "auggie",
		Tools: map[string]Tool{
			"auggie": {Enabled: true, Binary: "auggie", Flags: []string{}},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	path, _ := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file not created")
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.DefaultTool != "auggie" {
		t.Errorf("DefaultTool = %q, want %q", loaded.DefaultTool, "auggie")
	}
}

func TestGetDefaultTool(t *testing.T) {
	cfg := &Config{DefaultTool: ""}
	if got := cfg.GetDefaultTool(); got != "claude" {
		t.Errorf("GetDefaultTool() = %q, want default %q", got, "claude")
	}

	cfg.DefaultTool = "gemini"
	if got := cfg.GetDefaultTool(); got != "gemini" {
		t.Errorf("GetDefaultTool() = %q, want %q", got, "gemini")
	}
}

func TestSetDefaultTool(t *testing.T) {
	cfg := &Config{
		DefaultTool: "claude",
		Tools: map[string]Tool{
			"claude": {Enabled: true},
			"gemini": {Enabled: true},
		},
	}

	// Valid tool
	if err := cfg.SetDefaultTool("gemini"); err != nil {
		t.Errorf("SetDefaultTool(gemini) error = %v", err)
	}
	if cfg.DefaultTool != "gemini" {
		t.Errorf("DefaultTool = %q, want %q", cfg.DefaultTool, "gemini")
	}

	// Invalid tool
	if err := cfg.SetDefaultTool("unknown"); err == nil {
		t.Error("SetDefaultTool(unknown) should error")
	}
}

func TestGetTool(t *testing.T) {
	cfg := &Config{
		Tools: map[string]Tool{
			"claude": {Enabled: true, Binary: "claude"},
		},
	}

	tool, ok := cfg.GetTool("claude")
	if !ok {
		t.Error("GetTool(claude) should return true")
	}
	if tool.Binary != "claude" {
		t.Errorf("Binary = %q, want %q", tool.Binary, "claude")
	}

	_, ok = cfg.GetTool("unknown")
	if ok {
		t.Error("GetTool(unknown) should return false")
	}
}

func TestEnsureTool(t *testing.T) {
	cfg := &Config{
		Tools: map[string]Tool{
			"claude": {Enabled: true},
			"auggie": {Enabled: false},
		},
	}

	if err := cfg.EnsureTool("claude"); err != nil {
		t.Errorf("EnsureTool(claude) error = %v", err)
	}

	if err := cfg.EnsureTool("auggie"); err == nil {
		t.Error("EnsureTool(auggie) should error (disabled)")
	}

	if err := cfg.EnsureTool("unknown"); err == nil {
		t.Error("EnsureTool(unknown) should error")
	}
}

func TestSavePreservesComments(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	// Write a config file with comments (simulating mur init template)
	configDir := filepath.Join(tmpDir, ".mur")
	_ = os.MkdirAll(configDir, 0755)

	commentedConfig := `# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔧 Murmur Configuration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

schema_version: 2
default_tool: claude # 💡 Change with: mur config default <tool>

# ━━━ AI Tools ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
tools:
  claude:
    enabled: true
    binary: claude
    flags:
      - "-p"
    tier: paid # Requires API key or subscription
    capabilities:
      - coding
      - analysis
  gemini:
    enabled: true
    binary: gemini
    flags:
      - "-p"
    tier: free # Free tier available
    capabilities:
      - coding
      - simple-qa

# ━━━ Learning ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
learning:
  auto_extract: true
  sync_to_tools: true
  pattern_limit: 5
  llm:
    provider: ollama # 💡 Options: ollama | claude | openai | gemini
    model: llama3.2:3b
`
	configPath := filepath.Join(configDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte(commentedConfig), 0644)

	// Load + Save cycle (this is what destroys comments without the fix)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read back and check comments survived
	result, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	content := string(result)

	// Section dividers must survive
	wantComments := []string{
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"🔧 Murmur Configuration",
		"━━━ AI Tools",
		"━━━ Learning",
		"💡 Change with: mur config default <tool>",
		"💡 Options: ollama | claude | openai | gemini",
		"Requires API key or subscription",
		"Free tier available",
	}
	for _, want := range wantComments {
		if !strings.Contains(content, want) {
			t.Errorf("comment not preserved: %q\n\nFull output:\n%s", want, content)
		}
	}

	// Values must still be correct
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}
	if cfg2.DefaultTool != "claude" {
		t.Errorf("DefaultTool = %q, want %q", cfg2.DefaultTool, "claude")
	}
	if !cfg2.Learning.AutoExtract {
		t.Error("Learning.AutoExtract should be true")
	}
}

func TestSaveCreatesNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg := &Config{
		DefaultTool: "gemini",
		Tools: map[string]Tool{
			"gemini": {Enabled: true, Binary: "gemini"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, _ := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	content := string(data)

	// Should have the header comment for new files
	if !strings.Contains(content, "# Murmur Configuration") {
		t.Error("new file missing header comment")
	}

	// Values should be correct
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.DefaultTool != "gemini" {
		t.Errorf("DefaultTool = %q, want %q", loaded.DefaultTool, "gemini")
	}
}

func TestSaveUpdatesValues(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	// Write initial config with comments
	configDir := filepath.Join(tmpDir, ".mur")
	_ = os.MkdirAll(configDir, 0755)

	initial := `# Header comment
default_tool: claude # inline comment
tools:
  claude:
    enabled: true
    binary: claude
`
	configPath := filepath.Join(configDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte(initial), 0644)

	// Load, change a value, save
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.DefaultTool = "gemini"
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check value changed and comments survived
	result, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	content := string(result)

	if !strings.Contains(content, "default_tool: gemini") {
		t.Errorf("value not updated, got:\n%s", content)
	}
	if !strings.Contains(content, "# Header comment") {
		t.Errorf("header comment lost, got:\n%s", content)
	}
	if !strings.Contains(content, "inline comment") {
		t.Errorf("inline comment lost, got:\n%s", content)
	}
}

func TestMarshalDefaultConfigClean(t *testing.T) {
	cfg := defaultConfig()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal(defaultConfig()) error = %v", err)
	}

	output := string(data)

	// These zero-value fields should NOT appear in marshaled output
	noiseFields := []string{
		"tier: \"\"",
		"capabilities: []",
		"prefix_domain: null",
		"webhook_url: \"\"",
		"channel: \"\"",
		"repo: \"\"",
		"branch: \"\"",
		"interval_minutes: 0",
	}

	for _, noise := range noiseFields {
		if strings.Contains(output, noise) {
			t.Errorf("marshaled config contains zero-value noise: %q", noise)
		}
	}

	// These fields SHOULD be present (non-zero defaults)
	requiredFields := []string{
		"schema_version:",
		"default_tool:",
		"tools:",
		"enabled: true",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("marshaled config missing required field: %q", field)
		}
	}
}
