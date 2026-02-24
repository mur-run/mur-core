// Package sync provides pattern synchronization to AI CLI tools.
package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

// SyncFormat determines how patterns are synced.
type SyncFormat string

const (
	FormatDirectory SyncFormat = "directory" // Individual skill directories (L1/L2/L3)
	FormatSingle    SyncFormat = "single"    // Single merged file (legacy)
)

// L3Threshold is the default character count above which content is split to examples.md.
const L3Threshold = 500

// CleanOldPatternDirs removes old pattern directories from all skills folders.
// This cleans up the legacy format where each pattern had its own directory.
func CleanOldPatternDirs() (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("cannot determine home directory: %w", err)
	}

	cleaned := 0
	for _, target := range DefaultPatternTargets() {
		if !supportsDirectoryFormat(target) {
			continue
		}

		targetDir := filepath.Join(home, target.SkillsDir)
		count := cleanOldPatternDirsInTarget(targetDir)
		cleaned += count
	}

	return cleaned, nil
}

// cleanOldPatternDirsInTarget removes old pattern directories from a specific target.
func cleanOldPatternDirsInTarget(targetDir string) int {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return 0
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return 0
	}

	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip mur-index (the new format)
		if entry.Name() == "mur-index" {
			continue
		}
		// Check if this looks like a mur pattern directory (has SKILL.md inside)
		skillPath := filepath.Join(targetDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			// Read SKILL.md to confirm it's mur-generated
			content, err := os.ReadFile(skillPath)
			if err == nil && looksLikeMurPattern(string(content)) {
				// This is an old mur pattern directory, remove it
				if err := os.RemoveAll(filepath.Join(targetDir, entry.Name())); err == nil {
					cleaned++
				}
			}
		}
	}

	return cleaned
}

// looksLikeMurPattern checks if SKILL.md content looks like a mur-generated pattern.
func looksLikeMurPattern(content string) bool {
	// Check for common mur pattern markers
	markers := []string{
		"mur",
		"pattern",
		"Tags:",
		"## Instructions",
		"## Problem",
		"## Solution",
		"## Why This Is Non-Obvious",
		"## Verification",
	}
	for _, marker := range markers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
}

// PatternTarget defines where patterns are synced to for each CLI.
type PatternTarget struct {
	Name      string
	SkillsDir string // relative to home
	FileName  string // the skill file name
	Format    string // "markdown" or "yaml"
}

// DefaultPatternTargets returns all supported CLI targets.
func DefaultPatternTargets() []PatternTarget {
	return []PatternTarget{
		// Terminal CLIs
		{Name: "Claude Code", SkillsDir: ".claude/skills", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "Gemini CLI", SkillsDir: ".gemini/skills", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "Codex", SkillsDir: ".codex", FileName: "instructions.md", Format: "markdown"},
		{Name: "Auggie", SkillsDir: ".augment/skills", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "Aider", SkillsDir: ".aider", FileName: "conventions.md", Format: "markdown"},
		{Name: "OpenCode", SkillsDir: ".opencode", FileName: "instructions.md", Format: "markdown"},
		// IDE integrations
		{Name: "Continue", SkillsDir: ".continue/rules", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "Cursor", SkillsDir: ".cursor/rules", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "Windsurf", SkillsDir: ".windsurf/rules", FileName: "mur-patterns.md", Format: "markdown"},
		{Name: "GitHub Copilot", SkillsDir: ".github", FileName: "copilot-instructions.md", Format: "markdown"},
	}
}

// SyncPatternsToAllCLIs syncs patterns from ~/.mur/patterns/ to all CLI skill directories.
func SyncPatternsToAllCLIs() ([]SyncResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Load patterns
	store, err := pattern.DefaultStore()
	if err != nil {
		return nil, fmt.Errorf("cannot access pattern store: %w", err)
	}

	patterns, err := store.GetActive()
	if err != nil {
		return nil, fmt.Errorf("cannot load patterns: %w", err)
	}

	if len(patterns) == 0 {
		return []SyncResult{{
			Target:  "patterns",
			Success: true,
			Message: "No patterns to sync",
		}}, nil
	}

	// Sort patterns by effectiveness
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Learning.Effectiveness > patterns[j].Learning.Effectiveness
	})

	// Generate skill content
	content := generatePatternSkill(patterns)

	// Sync to each target
	var results []SyncResult
	for _, target := range DefaultPatternTargets() {
		targetDir := filepath.Join(home, target.SkillsDir)
		targetPath := filepath.Join(targetDir, target.FileName)

		// Create directory if needed
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			results = append(results, SyncResult{
				Target:  target.Name,
				Success: false,
				Message: fmt.Sprintf("Cannot create directory: %v", err),
			})
			continue
		}

		// For Codex, append to existing instructions.md
		if target.Name == "Codex" {
			content = generateCodexInstructions(patterns, targetPath)
		}

		// Write skill file
		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			results = append(results, SyncResult{
				Target:  target.Name,
				Success: false,
				Message: fmt.Sprintf("Cannot write file: %v", err),
			})
			continue
		}

		results = append(results, SyncResult{
			Target:  target.Name,
			Success: true,
			Message: fmt.Sprintf("Synced %d patterns", len(patterns)),
		})
	}

	return results, nil
}

// generatePatternSkill generates a markdown skill file from patterns.
func generatePatternSkill(patterns []pattern.Pattern) string {
	var sb strings.Builder

	sb.WriteString("# Learned Patterns\n\n")
	sb.WriteString("*Auto-generated by [mur](https://github.com/mur-run/mur-core). ")
	sb.WriteString(fmt.Sprintf("Updated: %s*\n\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString("These patterns were learned from previous development sessions. ")
	sb.WriteString("Apply them when relevant.\n\n")
	sb.WriteString("---\n\n")

	for _, p := range patterns {
		sb.WriteString(fmt.Sprintf("## %s\n\n", p.Name))

		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
		}

		// Tags
		var tags []string
		for _, t := range p.Tags.Confirmed {
			tags = append(tags, "`"+t+"`")
		}
		for _, ts := range p.Tags.Inferred {
			if ts.Confidence >= 0.7 {
				tags = append(tags, "`"+ts.Tag+"`")
			}
		}
		if len(tags) > 0 {
			sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(tags, " ")))
		}

		// Content
		content := p.Content
		if len(content) > 1000 {
			content = content[:1000] + "\n\n*(truncated)*"
		}
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString("\n*Run `mur sync` to update these patterns.*\n")

	return sb.String()
}

// generateCodexInstructions generates Codex-specific instructions format.
func generateCodexInstructions(patterns []pattern.Pattern, existingPath string) string {
	var sb strings.Builder

	// Read existing content if file exists
	if data, err := os.ReadFile(existingPath); err == nil {
		content := string(data)
		// Remove old mur section if present
		if idx := strings.Index(content, "<!-- mur:start -->"); idx != -1 {
			if endIdx := strings.Index(content, "<!-- mur:end -->"); endIdx != -1 {
				content = content[:idx] + content[endIdx+len("<!-- mur:end -->"):]
			}
		}
		sb.WriteString(strings.TrimSpace(content))
		sb.WriteString("\n\n")
	}

	// Add mur section
	sb.WriteString("<!-- mur:start -->\n")
	sb.WriteString("## Learned Patterns (mur)\n\n")

	for _, p := range patterns {
		sb.WriteString(fmt.Sprintf("### %s\n", p.Name))
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n", p.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("<!-- mur:end -->\n")

	return sb.String()
}

// SyncPatternsToTarget syncs patterns to a specific CLI target.
func SyncPatternsToTarget(targetName string) (*SyncResult, error) {
	for _, target := range DefaultPatternTargets() {
		if strings.EqualFold(target.Name, targetName) {
			results, err := SyncPatternsToAllCLIs()
			if err != nil {
				return nil, err
			}
			for _, r := range results {
				if strings.EqualFold(r.Target, targetName) {
					return &r, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unknown target: %s", targetName)
}

// SyncPatternsWithFormat syncs patterns using the specified format.
func SyncPatternsWithFormat(cfg *config.Config) ([]SyncResult, error) {
	format := SyncFormat(cfg.Sync.Format)
	if format == "" {
		format = FormatDirectory // default
	}

	switch format {
	case FormatDirectory:
		return SyncPatternsDirectory(cfg)
	case FormatSingle:
		return SyncPatternsToAllCLIs()
	default:
		return nil, fmt.Errorf("unknown sync format: %s", format)
	}
}

// SyncPatternsDirectory syncs a lightweight mur-index skill that instructs AI to use `mur search`.
// Patterns stay in ~/.mur/patterns/ and are loaded on-demand via semantic search.
func SyncPatternsDirectory(cfg *config.Config) ([]SyncResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Load patterns to get count
	store, err := pattern.DefaultStore()
	if err != nil {
		return nil, fmt.Errorf("cannot access pattern store: %w", err)
	}

	patterns, err := store.GetActive()
	if err != nil {
		return nil, fmt.Errorf("cannot load patterns: %w", err)
	}

	patternCount := len(patterns)

	// Sync to each target
	var results []SyncResult
	for _, target := range DefaultPatternTargets() {
		// For single-file targets, use legacy format
		if !supportsDirectoryFormat(target) {
			if patternCount > 0 {
				result := syncSingleFile(home, target, patterns)
				results = append(results, result)
			}
			continue
		}

		// For directory-supporting targets, create lightweight mur-index
		result := syncMurIndex(home, target, patternCount, cfg)
		results = append(results, result)
	}

	return results, nil
}

// supportsDirectoryFormat returns true if the target supports directory-based skills.
func supportsDirectoryFormat(target PatternTarget) bool {
	// These targets don't support directory format
	noDirectory := map[string]bool{
		"Codex":          true, // Uses single instructions.md
		"Aider":          true, // Uses single conventions.md
		"GitHub Copilot": true, // Uses single copilot-instructions.md
	}
	return !noDirectory[target.Name]
}

// syncMurIndex creates a lightweight mur-index skill that instructs AI to use `mur search`.
func syncMurIndex(home string, target PatternTarget, patternCount int, cfg *config.Config) SyncResult {
	targetDir := filepath.Join(home, target.SkillsDir)
	indexDir := filepath.Join(targetDir, "mur-index")

	// Clean old single-file format
	oldFile := filepath.Join(targetDir, target.FileName)
	_ = os.Remove(oldFile)

	// Clean old pattern directories (legacy format)
	cleanOldPatternDirsInTarget(targetDir)

	// Create mur-index directory
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("Cannot create mur-index directory: %v", err),
		}
	}

	// Generate lightweight SKILL.md
	skillContent := generateLightweightIndex(patternCount)
	skillPath := filepath.Join(indexDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("Cannot write SKILL.md: %v", err),
		}
	}

	return SyncResult{
		Target:  target.Name,
		Success: true,
		Message: fmt.Sprintf("Synced mur-index (%d patterns available)", patternCount),
	}
}

// generateLightweightIndex creates a minimal SKILL.md that tells AI to use mur search.
func generateLightweightIndex(patternCount int) string {
	return fmt.Sprintf(`---
name: mur-index
description: |
  Pattern library for searching learned development patterns and community knowledge.
  Search patterns before solving problems using mur CLI.
  Trigger keywords: mur, patterns, search, learned patterns, community patterns,
  best practices, error handling, debugging, solutions, 模式, 搜尋, 最佳實踐
---

# mur-index

Pattern library with %d local patterns + community patterns. **Do NOT load all at once.**

## Usage

Search for relevant patterns before solving problems:

%s

For patterns you don't have locally, search the community:

%s

## Examples

- %s - Search local patterns
- %s - Include community patterns
- %s - Only community patterns

Patterns are loaded on-demand to save tokens and stay relevant.

## Contributing

When you discover a non-obvious pattern, workaround, or fix during development:

%s

*Managed by [mur](https://github.com/mur-run/mur-core). Updated: %s*
`, patternCount,
		"`mur search \"<your query>\"`",
		"`mur search --community \"<your query>\"`",
		"`mur search \"API retry\"`",
		"`mur search --community \"Swift async\"`",
		"`mur search --community-only \"error handling\"`",
		"`mur learn add --name \"pattern-name\" --content \"what you discovered\"`",
		time.Now().Format("2006-01-02 15:04"))
}

// syncSingleFile syncs patterns as a single file (legacy format).
func syncSingleFile(home string, target PatternTarget, patterns []pattern.Pattern) SyncResult {
	targetDir := filepath.Join(home, target.SkillsDir)
	targetPath := filepath.Join(targetDir, target.FileName)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("Cannot create directory: %v", err),
		}
	}

	var content string
	if target.Name == "Codex" {
		content = generateCodexInstructions(patterns, targetPath)
	} else {
		content = generatePatternSkill(patterns)
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return SyncResult{
			Target:  target.Name,
			Success: false,
			Message: fmt.Sprintf("Cannot write file: %v", err),
		}
	}

	return SyncResult{
		Target:  target.Name,
		Success: true,
		Message: fmt.Sprintf("Synced %d patterns (single file)", len(patterns)),
	}
}

// NOTE: Legacy functions generateIndexSkill, generatePatternSkillDir, generateL2Skill,
// getSkillDirName, and extractSummary were removed in favor of the lightweight
// mur-index approach. See git history if needed.
