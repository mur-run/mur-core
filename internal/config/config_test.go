package config

import (
	"os"
	"path/filepath"
	"testing"
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
