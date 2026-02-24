// Package sync provides configuration synchronization to AI CLI tools.
// This file implements skills synchronization.
package sync

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a skill/methodology that can be synced.
type Skill struct {
	Name         string
	Description  string
	Instructions string
	SourcePath   string
}

// SkillsTarget represents where skills are synced to.
type SkillsTarget struct {
	Name      string
	SkillsDir string // relative to home
}

// DefaultSkillsTargets returns supported CLI skill directories.
func DefaultSkillsTargets() []SkillsTarget {
	return []SkillsTarget{
		{Name: "Claude Code", SkillsDir: ".claude/skills"},
		{Name: "Gemini CLI", SkillsDir: ".gemini/skills"},
		{Name: "Auggie", SkillsDir: ".augment/skills"},
		{Name: "OpenCode", SkillsDir: ".opencode/skills"},
		{Name: "OpenClaw", SkillsDir: ".agents/skills"},
		// Note: Codex uses ~/.codex/instructions.md instead of a skills directory
	}
}

// SkillsSourceDir returns the path to murmur skills directory.
func SkillsSourceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "skills"), nil
}

// SuperpowersSkillsDir returns the path to Superpowers plugin skills.
func SuperpowersSkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "plugins", "using-superpowers", "skills"), nil
}

// ListSkills returns all available skills from ~/.mur/skills/.
// This includes both standalone .md files and workflow skill directories
// (directories containing a SKILL.md file).
func ListSkills() ([]Skill, error) {
	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return nil, err
	}

	// Check if skills directory exists
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, nil // No skills yet
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read skills directory: %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if entry.IsDir() {
			// Check for workflow skill directories (contain SKILL.md)
			skillMDPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillMDPath); err == nil {
				skill, err := parseSkillFile(skillMDPath)
				if err == nil {
					skill.SourcePath = skillMDPath
					skills = append(skills, skill)
				}
			}
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		skill, err := parseSkillFile(skillPath)
		if err != nil {
			// Skip invalid files
			continue
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// parseSkillFile parses a SKILL.md file and extracts metadata.
func parseSkillFile(path string) (Skill, error) {
	file, err := os.Open(path)
	if err != nil {
		return Skill{}, err
	}
	defer func() { _ = file.Close() }()

	skill := Skill{
		SourcePath: path,
		Name:       strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var descLines []string
	var instrLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for skill name (# Title)
		if strings.HasPrefix(line, "# ") && skill.Name == strings.TrimSuffix(filepath.Base(path), ".md") {
			skill.Name = strings.TrimPrefix(line, "# ")
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "## ") {
			section := strings.ToLower(strings.TrimPrefix(line, "## "))
			if strings.Contains(section, "description") {
				currentSection = "description"
			} else if strings.Contains(section, "instruction") {
				currentSection = "instructions"
			} else {
				currentSection = ""
			}
			continue
		}

		// Collect content based on current section
		switch currentSection {
		case "description":
			descLines = append(descLines, line)
		case "instructions":
			instrLines = append(instrLines, line)
		}
	}

	skill.Description = strings.TrimSpace(strings.Join(descLines, "\n"))
	skill.Instructions = strings.TrimSpace(strings.Join(instrLines, "\n"))

	// If no description found, use first non-empty line after title
	if skill.Description == "" {
		_, _ = file.Seek(0, 0)
		scanner = bufio.NewScanner(file)
		foundTitle := false
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "# ") {
				foundTitle = true
				continue
			}
			if foundTitle && line != "" && !strings.HasPrefix(line, "#") {
				skill.Description = line
				break
			}
		}
	}

	return skill, nil
}

// SyncSkills syncs skills to all CLI tools.
// This syncs both standalone .md skill files and workflow skill directories
// (directories containing SKILL.md, exported via mur session export).
func SyncSkills() ([]SyncResult, error) {
	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return nil, err
	}

	// Check if skills directory exists
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no skills directory found at %s", skillsDir)
	}

	// List skill files
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read skills directory: %w", err)
	}

	// Collect standalone .md files
	var skillFiles []string
	// Collect workflow skill directories (contain SKILL.md)
	var workflowDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillMDPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillMDPath); err == nil {
				workflowDirs = append(workflowDirs, entry.Name())
			}
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			skillFiles = append(skillFiles, entry.Name())
		}
	}

	if len(skillFiles) == 0 && len(workflowDirs) == 0 {
		return nil, fmt.Errorf("no skill files found in %s", skillsDir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	var results []SyncResult
	for _, target := range DefaultSkillsTargets() {
		result := syncSkillsToTarget(home, skillsDir, target, skillFiles)
		if !result.Success {
			results = append(results, result)
			continue
		}

		// Also sync workflow skill directories
		wfResult := syncWorkflowDirsToTarget(home, skillsDir, target, workflowDirs)
		totalCount := len(skillFiles) + len(workflowDirs)
		if wfResult.Success {
			result.Message = fmt.Sprintf("synced %d skills (%d files, %d workflows)",
				totalCount, len(skillFiles), len(workflowDirs))
		}
		results = append(results, result)
	}

	return results, nil
}

// syncSkillsToTarget syncs skills to a single CLI target.
func syncSkillsToTarget(home, skillsDir string, target SkillsTarget, skillFiles []string) SyncResult {
	targetDir := filepath.Join(home, target.SkillsDir)

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("cannot create directory: %v", err),
		}
	}

	// Symlink each skill file
	copied := 0
	for _, filename := range skillFiles {
		srcPath := filepath.Join(skillsDir, filename)
		dstPath := filepath.Join(targetDir, filename)

		// Remove existing (old copy or stale symlink)
		if info, err := os.Lstat(dstPath); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				if linkTarget, err := os.Readlink(dstPath); err == nil && linkTarget == srcPath {
					copied++
					continue
				}
			}
			os.Remove(dstPath)
		}

		if err := os.Symlink(srcPath, dstPath); err != nil {
			// Fallback to copy if symlink fails (e.g. cross-device)
			if err := copyFile(srcPath, dstPath); err != nil {
				return SyncResult{
					Target:  target.Name,
					Success: false,
					Message: fmt.Sprintf("cannot sync %s: %v", filename, err),
				}
			}
		}
		copied++
	}

	return SyncResult{
		Target:  target.Name,
		Success: true,
		Message: fmt.Sprintf("synced %d skills", copied),
	}
}

// syncWorkflowDirsToTarget symlinks workflow skill directories to a CLI target.
// Each workflow directory is symlinked so edits in ~/.mur/skills/ are instantly
// reflected without re-running sync.
func syncWorkflowDirsToTarget(home, skillsDir string, target SkillsTarget, workflowDirs []string) SyncResult {
	targetDir := filepath.Join(home, target.SkillsDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("cannot create directory: %v", err),
		}
	}

	linked := 0
	for _, dirName := range workflowDirs {
		srcDir := filepath.Join(skillsDir, dirName)
		dstDir := filepath.Join(targetDir, dirName)

		// Remove existing (old copy or stale symlink)
		if info, err := os.Lstat(dstDir); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				// Already a symlink — check if it points to the right place
				target, err := os.Readlink(dstDir)
				if err == nil && target == srcDir {
					linked++
					continue
				}
			}
			os.RemoveAll(dstDir)
		}

		if err := os.Symlink(srcDir, dstDir); err != nil {
			return SyncResult{
				Target:  target.Name,
				Success: false,
				Message: fmt.Sprintf("cannot symlink workflow skill %s: %v", dirName, err),
			}
		}
		linked++
	}

	return SyncResult{
		Target:  target.Name,
		Success: true,
		Message: fmt.Sprintf("synced %d workflow skills", linked),
	}
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// ImportSkill imports a skill from a file path to ~/.mur/skills/
func ImportSkill(path string) error {
	// Validate source exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return err
	}

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("cannot create skills directory: %w", err)
	}

	// Copy to skills directory
	filename := filepath.Base(path)
	if !strings.HasSuffix(filename, ".md") {
		filename = filename + ".md"
	}
	dstPath := filepath.Join(skillsDir, filename)

	return copyFile(path, dstPath)
}

// ImportFromSuperpowers imports skills from Superpowers plugin.
// Returns number of skills imported.
func ImportFromSuperpowers() (int, error) {
	spDir, err := SuperpowersSkillsDir()
	if err != nil {
		return 0, err
	}

	// Check if Superpowers skills directory exists
	if _, err := os.Stat(spDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("superpowers skills not found at %s", spDir)
	}

	entries, err := os.ReadDir(spDir)
	if err != nil {
		return 0, fmt.Errorf("cannot read Superpowers skills: %w", err)
	}

	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return 0, err
	}

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return 0, fmt.Errorf("cannot create skills directory: %w", err)
	}

	imported := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		srcPath := filepath.Join(spDir, entry.Name())
		dstPath := filepath.Join(skillsDir, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			continue // Skip files that fail
		}
		imported++
	}

	if imported == 0 {
		return 0, fmt.Errorf("no skills found in Superpowers plugin")
	}

	return imported, nil
}

// EnsureSkillsDir creates the skills directory if it doesn't exist.
func EnsureSkillsDir() error {
	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(skillsDir, 0755)
}
