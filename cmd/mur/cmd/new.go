package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <pattern-name>",
	Short: "Create a new pattern from template",
	Long: `Create a new pattern file with a template and open in editor.

Examples:
  mur new go-testing          # Create and edit new pattern
  mur new swift-concurrency   # Create Swift-related pattern`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		return fmt.Errorf("failed to create patterns directory: %w", err)
	}

	patternPath := filepath.Join(patternsDir, patternName+".yaml")

	// Check if pattern already exists
	if _, err := os.Stat(patternPath); err == nil {
		return fmt.Errorf("pattern already exists: %s\nUse 'mur edit %s' to modify it", patternName, patternName)
	}

	// Infer domain from name
	domain := inferDomain(patternName)
	tags := inferTags(patternName)

	// Create template
	template := fmt.Sprintf(`# Pattern: %s
# Created: %s

id: %s
name: %s
description: |
  TODO: Describe what this pattern teaches.

content: |
  TODO: Write the pattern content here.
  
  This content will be injected into AI assistant context
  when relevant to the current task.

tags:
  confirmed:
%s
  inferred: []
  negative: []

applies:
  contexts: []
  file_patterns: []
  keywords: []

security:
  trust_level: user
  content_hash: ""

learning:
  usage_count: 0
  effectiveness: 0.0
  last_used: null
  feedback_scores: []

lifecycle:
  status: active
  created: %s

schema_version: 2
`, patternName, time.Now().Format("2006-01-02"),
		patternName, patternName,
		formatTags(tags),
		time.Now().Format(time.RFC3339))

	if err := os.WriteFile(patternPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create pattern file: %w", err)
	}

	fmt.Printf("✨ Created pattern: %s\n", patternName)
	fmt.Printf("   Domain: %s\n", domain)
	fmt.Printf("   Tags: %s\n", strings.Join(tags, ", "))
	fmt.Println()

	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}

	if editor != "" {
		fmt.Println("Opening in editor...")
		editorCmd := exec.Command(editor, patternPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		_ = editorCmd.Run()
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit the pattern content")
	fmt.Println("  2. Run 'mur lint", patternName+"' to validate")
	fmt.Println("  3. Run 'mur sync' to sync to AI tools")

	return nil
}

func inferDomain(name string) string {
	name = strings.ToLower(name)
	domains := map[string][]string{
		"go":     {"go-", "golang", "goroutine"},
		"swift":  {"swift-", "swiftui", "ios-", "macos-"},
		"python": {"python-", "py-", "django", "flask"},
		"node":   {"node-", "npm-", "javascript", "typescript", "ts-", "js-"},
		"rust":   {"rust-", "cargo-"},
	}

	for domain, patterns := range domains {
		for _, p := range patterns {
			if strings.Contains(name, p) {
				return domain
			}
		}
	}
	return "general"
}

func inferTags(name string) []string {
	tags := []string{}
	name = strings.ToLower(name)

	// Domain tags
	if domain := inferDomain(name); domain != "general" {
		tags = append(tags, domain)
	}

	// Common pattern tags
	tagPatterns := map[string][]string{
		"error":       {"error-", "exception", "handling"},
		"testing":     {"test", "spec", "mock"},
		"performance": {"perf", "optim", "cache"},
		"security":    {"secur", "auth", "crypto"},
		"api":         {"api-", "rest-", "graphql"},
		"database":    {"db-", "sql", "database", "postgres", "mysql"},
		"concurrency": {"concur", "async", "parallel", "thread"},
	}

	for tag, patterns := range tagPatterns {
		for _, p := range patterns {
			if strings.Contains(name, p) {
				tags = append(tags, tag)
				break
			}
		}
	}

	if len(tags) == 0 {
		tags = append(tags, "general")
	}

	return tags
}

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "    - general"
	}
	var lines []string
	for _, t := range tags {
		lines = append(lines, "    - "+t)
	}
	return strings.Join(lines, "\n")
}
