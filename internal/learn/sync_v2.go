// Package learn provides pattern sync using Schema v2.
package learn

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/team"
)

// SyncPatternsV2 syncs all patterns (Schema v2) to CLI tools and team repo.
func SyncPatternsV2() ([]SyncResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	patterns, err := store.GetActive()
	if err != nil {
		return nil, fmt.Errorf("failed to list patterns: %w", err)
	}

	if len(patterns) == 0 {
		return []SyncResult{
			{Target: "All CLIs", Success: true, Message: "no patterns to sync"},
		}, nil
	}

	var results []SyncResult

	// Sync to all CLI tools
	cliTargets := []struct {
		name   string
		syncFn func(string, []pattern.Pattern) SyncResult
	}{
		{"Claude Code", syncToClaudeCodeV2},
		{"Gemini CLI", syncToGeminiCLIV2},
		{"Auggie", syncToAuggieV2},
		{"Codex", syncToCodexV2},
		{"OpenCode", syncToOpenCodeV2},
		{"Aider", syncToAiderV2},
		{"Continue", syncToContinueV2},
		{"Cursor", syncToCursorV2},
	}

	for _, target := range cliTargets {
		result := target.syncFn(home, patterns)
		results = append(results, result)
	}

	// Sync team-shared patterns to team repo
	if team.IsInitialized() {
		teamResult := syncToTeamRepoV2(patterns)
		results = append(results, teamResult)
	}

	return results, nil
}

// syncToClaudeCodeV2 syncs patterns to ~/.claude/skills/learned-{name}/SKILL.md
func syncToClaudeCodeV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".claude", "skills")

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
		content := patternToSkillV2(p)

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

// syncToGeminiCLIV2 syncs patterns to ~/.gemini/skills/learned-{name}.md
func syncToGeminiCLIV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".gemini", "skills")

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
		content := patternToSkillV2(p)

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

// syncToAuggieV2 syncs patterns to ~/.augment/skills/learned-{name}.md
func syncToAuggieV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".augment", "skills")

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
		content := patternToSkillV2(p)

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

// syncToCodexV2 syncs patterns to ~/.codex/instructions.md
func syncToCodexV2(home string, patterns []pattern.Pattern) SyncResult {
	codexDir := filepath.Join(home, ".codex")

	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return SyncResult{
			Target:  "Codex",
			Success: false,
			Message: fmt.Sprintf("cannot create codex directory: %v", err),
		}
	}

	// Build patterns section
	var sb strings.Builder
	sb.WriteString("\n\n## Learned Patterns (mur)\n\n")
	for _, p := range patterns {
		sb.WriteString(fmt.Sprintf("### %s\n\n", p.Name))
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
		}
		sb.WriteString(p.Content)
		sb.WriteString("\n\n")
	}

	instructionsPath := filepath.Join(codexDir, "instructions.md")

	existing, _ := os.ReadFile(instructionsPath)
	existingStr := string(existing)

	// Remove old patterns section
	if idx := strings.Index(existingStr, "\n\n## Learned Patterns"); idx != -1 {
		existingStr = existingStr[:idx]
	}

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

// syncToOpenCodeV2 syncs patterns to ~/.opencode/skills/
func syncToOpenCodeV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".opencode", "skills")

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
		content := patternToSkillV2(p)

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

// syncToAiderV2 syncs patterns to ~/.aider/conventions/
func syncToAiderV2(home string, patterns []pattern.Pattern) SyncResult {
	conventionsDir := filepath.Join(home, ".aider", "conventions")

	if err := os.MkdirAll(conventionsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Aider",
			Success: false,
			Message: fmt.Sprintf("cannot create conventions directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		conventionPath := filepath.Join(conventionsDir, fileName)
		content := patternToSkillV2(p)

		if err := os.WriteFile(conventionPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Aider",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.aider/conventions/", synced),
	}
}

// syncToContinueV2 syncs patterns to ~/.continue/skills/
func syncToContinueV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".continue", "skills")

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Continue",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		skillPath := filepath.Join(skillsDir, fileName)
		content := patternToSkillV2(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Continue",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.continue/skills/", synced),
	}
}

// syncToCursorV2 syncs patterns to ~/.cursor/skills/
func syncToCursorV2(home string, patterns []pattern.Pattern) SyncResult {
	skillsDir := filepath.Join(home, ".cursor", "skills")

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Cursor",
			Success: false,
			Message: fmt.Sprintf("cannot create skills directory: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		fileName := fmt.Sprintf("learned-%s.md", p.Name)
		skillPath := filepath.Join(skillsDir, fileName)
		content := patternToSkillV2(p)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			continue
		}
		synced++
	}

	return SyncResult{
		Target:  "Cursor",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to ~/.cursor/skills/", synced),
	}
}

// patternToSkillV2 converts a Pattern v2 to SKILL.md format.
func patternToSkillV2(p pattern.Pattern) string {
	var sb strings.Builder

	// Combine keywords from Tags.Confirmed and Applies.Keywords (deduplicated)
	seen := make(map[string]bool)
	var combinedKeywords []string
	for _, t := range p.Tags.Confirmed {
		lower := strings.ToLower(t)
		if !seen[lower] {
			seen[lower] = true
			combinedKeywords = append(combinedKeywords, t)
		}
	}
	for _, k := range p.Applies.Keywords {
		lower := strings.ToLower(k)
		if !seen[lower] {
			seen[lower] = true
			combinedKeywords = append(combinedKeywords, k)
		}
	}

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: learned-%s\n", p.Name))
	sb.WriteString("description: |\n")
	if p.Description != "" {
		sb.WriteString(fmt.Sprintf("  %s\n", p.Description))
	}
	if len(combinedKeywords) > 0 {
		sb.WriteString(fmt.Sprintf("  Trigger keywords: %s\n", strings.Join(combinedKeywords, ", ")))
	}
	sb.WriteString("---\n\n")

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", p.Name))

	// Description
	if p.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
	}

	// Tags
	sb.WriteString("## Context\n\n")

	// Confirmed tags
	if len(p.Tags.Confirmed) > 0 {
		sb.WriteString(fmt.Sprintf("- **Tags:** %s\n", strings.Join(p.Tags.Confirmed, ", ")))
	}

	// Top inferred tags
	if len(p.Tags.Inferred) > 0 {
		topTags := p.GetTopTags(3)
		tagStrs := make([]string, len(topTags))
		for i, t := range topTags {
			tagStrs[i] = fmt.Sprintf("%s (%.0f%%)", t.Tag, t.Confidence*100)
		}
		sb.WriteString(fmt.Sprintf("- **Inferred:** %s\n", strings.Join(tagStrs, ", ")))
	}

	// Effectiveness
	sb.WriteString(fmt.Sprintf("- **Effectiveness:** %.0f%%\n", p.Learning.Effectiveness*100))
	sb.WriteString(fmt.Sprintf("- **Uses:** %d\n\n", p.Learning.UsageCount))

	// Apply conditions
	if len(p.Applies.Languages) > 0 || len(p.Applies.Frameworks) > 0 {
		sb.WriteString("## Applies To\n\n")
		if len(p.Applies.Languages) > 0 {
			sb.WriteString(fmt.Sprintf("- **Languages:** %s\n", strings.Join(p.Applies.Languages, ", ")))
		}
		if len(p.Applies.Frameworks) > 0 {
			sb.WriteString(fmt.Sprintf("- **Frameworks:** %s\n", strings.Join(p.Applies.Frameworks, ", ")))
		}
		if len(p.Applies.Keywords) > 0 {
			sb.WriteString(fmt.Sprintf("- **Keywords:** %s\n", strings.Join(p.Applies.Keywords, ", ")))
		}
		sb.WriteString("\n")
	}

	// Content
	sb.WriteString("## Content\n\n")
	sb.WriteString(p.Content)
	sb.WriteString("\n\n")

	// Footer
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("*Synced from mur v2 | Trust: %s | Status: %s*\n",
		p.Security.TrustLevel, p.Lifecycle.Status))

	return sb.String()
}

// syncToTeamRepoV2 syncs team-trusted patterns to the team repo.
func syncToTeamRepoV2(patterns []pattern.Pattern) SyncResult {
	teamPatternsDir, err := team.PatternsDir()
	if err != nil {
		return SyncResult{
			Target:  "Team Repo",
			Success: false,
			Message: fmt.Sprintf("cannot get team patterns dir: %v", err),
		}
	}

	if err := os.MkdirAll(teamPatternsDir, 0755); err != nil {
		return SyncResult{
			Target:  "Team Repo",
			Success: false,
			Message: fmt.Sprintf("cannot create team patterns dir: %v", err),
		}
	}

	synced := 0
	for _, p := range patterns {
		// Only sync team or owner trusted patterns
		if p.Security.TrustLevel != pattern.TrustTeam && p.Security.TrustLevel != pattern.TrustOwner {
			continue
		}

		// Serialize pattern
		data, err := p.ToYAML()
		if err != nil {
			continue
		}

		dstPath := filepath.Join(teamPatternsDir, p.Name+".yaml")
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			continue
		}
		synced++
	}

	if synced == 0 {
		return SyncResult{
			Target:  "Team Repo",
			Success: true,
			Message: "no team patterns to sync",
		}
	}

	return SyncResult{
		Target:  "Team Repo",
		Success: true,
		Message: fmt.Sprintf("synced %d patterns to team repo", synced),
	}
}
