package learn

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/karajanchang/murmur-ai/internal/team"
)

// SyncResult holds the result of a pattern sync operation.
type SyncResult struct {
	Target  string
	Success bool
	Message string
}

// SyncPatterns syncs all patterns to CLI tools and team repo.
func SyncPatterns() ([]SyncResult, error) {
	patterns, err := List()
	if err != nil {
		return nil, fmt.Errorf("failed to list patterns: %w", err)
	}

	if len(patterns) == 0 {
		return []SyncResult{
			{Target: "Claude Code", Success: true, Message: "no patterns to sync"},
			{Target: "Gemini CLI", Success: true, Message: "no patterns to sync"},
		}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	var results []SyncResult

	// Sync to Claude Code
	claudeResult := syncToClaudeCode(home, patterns)
	results = append(results, claudeResult)

	// Sync to Gemini CLI
	geminiResult := syncToGeminiCLI(home, patterns)
	results = append(results, geminiResult)

	// Sync to Auggie
	auggieResult := syncToAuggie(home, patterns)
	results = append(results, auggieResult)

	// Sync to Codex (uses instructions.md)
	codexResult := syncToCodex(home, patterns)
	results = append(results, codexResult)

	// Sync to OpenCode
	opencodeResult := syncToOpenCode(home, patterns)
	results = append(results, opencodeResult)

	// Sync team-shared patterns to team repo
	if team.IsInitialized() {
		teamResult := syncToTeamRepo(patterns)
		results = append(results, teamResult)
	}

	return results, nil
}

// syncToClaudeCode syncs patterns to ~/.claude/skills/learned-{name}/SKILL.md
func syncToClaudeCode(home string, patterns []Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".claude", "skills")

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Claude Code",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		dirName := fmt.Sprintf("learned-%s", p.Name)
		patternDir := filepath.Join(skillsDir, dirName)

		if err := os.MkdirAll(patternDir, 0755); err != nil {
			continue
		}

		skillPath := filepath.Join(patternDir, "SKILL.md")
		content := patternToSkill(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Claude Code",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.claude/skills/", synced),
	}
}

// syncToGeminiCLI syncs patterns to ~/.gemini/skills/learned-{name}.md
func syncToGeminiCLI(home string, patterns []Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".gemini", "skills")

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Gemini CLI",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		skillPath := filepath.Join(skillsDir, fileName)
		content := patternToSkill(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Gemini CLI",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.gemini/skills/", synced),
	}
}

// syncToAuggie syncs patterns to ~/.augment/skills/learned-{name}.md
func syncToAuggie(home string, patterns []Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".augment", "skills")

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Auggie",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		skillPath := filepath.Join(skillsDir, fileName)
		content := patternToSkill(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Auggie",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.augment/skills/", synced),
	}
}

// syncToCodex syncs patterns to ~/.codex/instructions.md (appends patterns section)
func syncToCodex(home string, patterns []Pattern) SyncResult {
	codexDir := filepath.Join(home, ".codex")

	// Ensure directory exists
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return SyncResult{
			Target:  "Codex",
			Success: false,
			Message: fmt.Sprintf("cannot create codex directory: %v", err),
		}
	}

	// Build patterns section
	var sb strings.Builder
	sb.WriteString("\n\n## Learned Patterns (murmur-ai)\n\n")
	for _, p := range patterns {
		sb.WriteString(fmt.Sprintf("### %s\n\n", p.Name))
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
		}
		sb.WriteString(p.Content)
		sb.WriteString("\n\n")
	}

	instructionsPath := filepath.Join(codexDir, "instructions.md")

	// Read existing content
	existing, _ := os.ReadFile(instructionsPath)
	existingStr := string(existing)

	// Remove old patterns section if present
	if idx := strings.Index(existingStr, "\n\n## Learned Patterns (murmur-ai)"); idx != -1 {
		existingStr = existingStr[:idx]
	}

	// Append new patterns section
	newContent := existingStr + sb.String()

	if err := os.WriteFile(instructionsPath, []byte(newContent), 0644); err != nil {
		return SyncResult{
			Target:  "Codex",
			Success: false,
			Message: fmt.Sprintf("cannot write instructions.md: %v", err),
		}
	}

	return SyncResult{
		Target:  "Codex",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.codex/instructions.md", len(patterns)),
	}
}

// syncToOpenCode syncs patterns to ~/.opencode/skills/learned-{name}.md
func syncToOpenCode(home string, patterns []Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".opencode", "skills")

	// Ensure skills directory exists
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "OpenCode",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		skillPath := filepath.Join(skillsDir, fileName)
		content := patternToSkill(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "OpenCode",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.opencode/skills/", synced),
	}
}

// patternToSkill converts a Pattern to SKILL.md format.
func patternToSkill(p Pattern) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", p.Name))

	// Description
	if p.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
	}

	// Domain and Category
	sb.WriteString("## Context\n\n")
	sb.WriteString(fmt.Sprintf("- **Domain:** %s\n", p.Domain))
	sb.WriteString(fmt.Sprintf("- **Category:** %s\n", p.Category))
	sb.WriteString(fmt.Sprintf("- **Confidence:** %.0f%%\n\n", p.Confidence*100))

	// Content
	sb.WriteString("## Content\n\n")
	sb.WriteString(p.Content)
	sb.WriteString("\n\n")

	// Footer
	sb.WriteString("---\n")
	sb.WriteString("*Synced from murmur-ai*\n")

	return sb.String()
}

// CleanupSyncedPatterns removes synced patterns that no longer exist in the source.
func CleanupSyncedPatterns() error {
	patterns, err := List()
	if err != nil {
		return err
	}

	// Build set of valid pattern names
	validNames := make(map[string]bool)
	for _, p := range patterns {
		validNames[p.Name] = true
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Clean up Claude Code
	claudeSkills := filepath.Join(home, ".claude", "skills")
	if entries, err := os.ReadDir(claudeSkills); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "learned-") {
				name := strings.TrimPrefix(entry.Name(), "learned-")
				if !validNames[name] {
					_ = os.RemoveAll(filepath.Join(claudeSkills, entry.Name()))
				}
			}
		}
	}

	// Clean up Gemini CLI
	geminiSkills := filepath.Join(home, ".gemini", "skills")
	if entries, err := os.ReadDir(geminiSkills); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "learned-") && strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimPrefix(entry.Name(), "learned-")
				name = strings.TrimSuffix(name, ".md")
				if !validNames[name] {
					_ = os.Remove(filepath.Join(geminiSkills, entry.Name()))
				}
			}
		}
	}

	// Clean up Auggie
	auggieSkills := filepath.Join(home, ".augment", "skills")
	if entries, err := os.ReadDir(auggieSkills); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "learned-") && strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimPrefix(entry.Name(), "learned-")
				name = strings.TrimSuffix(name, ".md")
				if !validNames[name] {
					_ = os.Remove(filepath.Join(auggieSkills, entry.Name()))
				}
			}
		}
	}

	// Clean up OpenCode
	opencodeSkills := filepath.Join(home, ".opencode", "skills")
	if entries, err := os.ReadDir(opencodeSkills); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "learned-") && strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimPrefix(entry.Name(), "learned-")
				name = strings.TrimSuffix(name, ".md")
				if !validNames[name] {
					_ = os.Remove(filepath.Join(opencodeSkills, entry.Name()))
				}
			}
		}
	}

	// Note: Codex uses a single instructions.md file, so we regenerate the whole file
	// instead of cleaning up individual patterns. This is handled in syncToCodex.

	return nil
}

// syncToTeamRepo syncs team-shared patterns to the team repo.
func syncToTeamRepo(patterns []Pattern) SyncResult {
	teamPatternsDir, err := team.PatternsDir()
	if err != nil {
		return SyncResult{
			Target:  "Team Repo",
			Success: false,
			Message: fmt.Sprintf("cannot get team patterns dir: %v", err),
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(teamPatternsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Team Repo",
			Success: false,
			Message: fmt.Sprintf("cannot create team patterns dir: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		if !p.TeamShared {
			continue
		}

		// Copy pattern file to team repo
		srcPath, err := patternPath(p.Name)
		if err != nil {
			continue
		}

		dstPath := filepath.Join(teamPatternsDir, p.Name+".yaml")
		if err := copyFile(srcPath, dstPath); err != nil {
			continue
		}
		synced++
	}

	if synced == 0 {
		return SyncResult{
			Target:  "Team Repo",
			Success: true,
			Message: "no team-shared patterns",
		}
	}

	return SyncResult{
		Target:  "Team Repo",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to team repo", synced),
	}
}

// SyncFromTeamRepo imports patterns from team repo to local.
func SyncFromTeamRepo() ([]SyncResult, error) {
	if !team.IsInitialized() {
		return nil, fmt.Errorf("team repo not initialized")
	}

	teamPatternsDir, err := team.PatternsDir()
	if err != nil {
		return nil, err
	}

	// Check if team patterns directory exists
	if _, err := os.Stat(teamPatternsDir); os.IsNotExist(err) {
		return []SyncResult{
			{Target: "Local", Success: true, Message: "no team patterns found"},
		}, nil
	}

	entries, err := os.ReadDir(teamPatternsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read team patterns: %w", err)
	}

	imported := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		srcPath := filepath.Join(teamPatternsDir, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".yaml")

		dstPath, err := patternPath(name)
		if err != nil {
			continue
		}

		// Copy team pattern to local
		if err := copyFile(srcPath, dstPath); err != nil {
			continue
		}
		imported++
	}

	return []SyncResult{
		{
			Target:  "Local",
			Success: true,
			Message: fmt.Sprintf("imported %d patterns from team repo", imported),
		},
	}, nil
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
