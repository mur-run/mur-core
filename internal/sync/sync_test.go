package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTargets(t *testing.T) {
	targets := DefaultTargets()
	if len(targets) == 0 {
		t.Error("DefaultTargets() returned empty list")
	}

	// Should have at least Claude and Gemini
	names := make(map[string]bool)
	for _, target := range targets {
		names[target.Name] = true
	}

	if !names["Claude Code"] {
		t.Error("missing Claude Code target")
	}
	if !names["Gemini CLI"] {
		t.Error("missing Gemini CLI target")
	}
}

func TestDefaultSkillsTargets(t *testing.T) {
	targets := DefaultSkillsTargets()
	if len(targets) == 0 {
		t.Error("DefaultSkillsTargets() returned empty list")
	}
}

func TestSkillsSourceDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := SkillsSourceDir()
	if err != nil {
		t.Fatalf("SkillsSourceDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "skills")
	if dir != expected {
		t.Errorf("SkillsSourceDir() = %q, want %q", dir, expected)
	}
}

func TestListSkillsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if skills != nil && len(skills) > 0 {
		t.Errorf("ListSkills() on empty = %v, want nil/empty", skills)
	}
}

func TestListSkillsWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create skills dir with a skill file
	skillsDir := filepath.Join(tmpDir, ".murmur", "skills")
	os.MkdirAll(skillsDir, 0755)

	skillContent := `# Test Skill

## Description
A test skill for testing.

## Instructions
Do test things.
`
	os.WriteFile(filepath.Join(skillsDir, "test-skill.md"), []byte(skillContent), 0644)

	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if len(skills) > 0 && skills[0].Name != "Test Skill" {
		t.Errorf("Skill name = %q, want 'Test Skill'", skills[0].Name)
	}
}

func TestEnsureSkillsDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	if err := EnsureSkillsDir(); err != nil {
		t.Fatalf("EnsureSkillsDir() error = %v", err)
	}

	skillsDir := filepath.Join(tmpDir, ".murmur", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("EnsureSkillsDir() did not create directory")
	}
}

func TestImportSkillNotFound(t *testing.T) {
	err := ImportSkill("/nonexistent/path/skill.md")
	if err == nil {
		t.Error("ImportSkill() with nonexistent file should error")
	}
}

func TestImportSkill(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create a source skill file
	srcPath := filepath.Join(tmpDir, "source-skill.md")
	os.WriteFile(srcPath, []byte("# My Skill\n\nContent here."), 0644)

	if err := ImportSkill(srcPath); err != nil {
		t.Fatalf("ImportSkill() error = %v", err)
	}

	// Check skill was imported
	skillsDir := filepath.Join(tmpDir, ".murmur", "skills")
	dstPath := filepath.Join(skillsDir, "source-skill.md")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("ImportSkill() did not copy file to skills directory")
	}
}

func TestSyncResult(t *testing.T) {
	result := SyncResult{
		Target:  "Test Target",
		Success: true,
		Message: "ok",
	}

	if !result.Success {
		t.Error("SyncResult.Success should be true")
	}
	if result.Target != "Test Target" {
		t.Errorf("Target = %q, want 'Test Target'", result.Target)
	}
}

func TestSyncSkillsNoDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := SyncSkills()
	if err == nil {
		t.Error("SyncSkills() with no skills dir should error")
	}
}

func TestSyncSkillsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create empty skills dir
	skillsDir := filepath.Join(tmpDir, ".murmur", "skills")
	os.MkdirAll(skillsDir, 0755)

	_, err := SyncSkills()
	if err == nil {
		t.Error("SyncSkills() with empty skills dir should error")
	}
}
