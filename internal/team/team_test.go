package team

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTeamDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := TeamDir()
	if err != nil {
		t.Fatalf("TeamDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "team")
	if dir != expected {
		t.Errorf("TeamDir() = %q, want %q", dir, expected)
	}
}

func TestIsInitializedFalse(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	if IsInitialized() {
		t.Error("IsInitialized() should return false for empty dir")
	}
}

func TestIsInitializedTrue(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create team dir with .git
	teamDir := filepath.Join(tmpDir, ".murmur", "team", ".git")
	_ = os.MkdirAll(teamDir, 0755)

	if !IsInitialized() {
		t.Error("IsInitialized() should return true when .git exists")
	}
}

func TestStatusNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	status, err := Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if status.Initialized {
		t.Error("Status.Initialized should be false")
	}
	if status.LocalPath == "" {
		t.Error("Status.LocalPath should be set")
	}
}

func TestPatternsDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := PatternsDir()
	if err != nil {
		t.Fatalf("PatternsDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "team", "patterns")
	if dir != expected {
		t.Errorf("PatternsDir() = %q, want %q", dir, expected)
	}
}

func TestHooksDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := HooksDir()
	if err != nil {
		t.Fatalf("HooksDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "team", "hooks")
	if dir != expected {
		t.Errorf("HooksDir() = %q, want %q", dir, expected)
	}
}

func TestSkillsDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := SkillsDir()
	if err != nil {
		t.Fatalf("SkillsDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "team", "skills")
	if dir != expected {
		t.Errorf("SkillsDir() = %q, want %q", dir, expected)
	}
}

func TestMCPDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := MCPDir()
	if err != nil {
		t.Fatalf("MCPDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "team", "mcp")
	if dir != expected {
		t.Errorf("MCPDir() = %q, want %q", dir, expected)
	}
}

func TestEnsureStructure(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	if err := EnsureStructure(); err != nil {
		t.Fatalf("EnsureStructure() error = %v", err)
	}

	// Check all directories created
	dirs := []string{"patterns", "hooks", "skills", "mcp"}
	teamDir := filepath.Join(tmpDir, ".murmur", "team")
	for _, d := range dirs {
		path := filepath.Join(teamDir, d)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("EnsureStructure() did not create %s", d)
		}
	}
}

func TestPullNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	err := Pull()
	if err == nil {
		t.Error("Pull() should error when not initialized")
	}
}

func TestPushNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	err := Push("test message")
	if err == nil {
		t.Error("Push() should error when not initialized")
	}
}

func TestSyncNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	err := Sync()
	if err == nil {
		t.Error("Sync() should error when not initialized")
	}
}

func TestCloneAlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create team dir with .git
	teamDir := filepath.Join(tmpDir, ".murmur", "team", ".git")
	_ = os.MkdirAll(teamDir, 0755)

	err := Clone("https://github.com/example/repo.git")
	if err == nil {
		t.Error("Clone() should error when already initialized")
	}
}
