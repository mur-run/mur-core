package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import patterns from external sources",
	Long:  `Import patterns from GitHub Gists, URLs, or files.`,
}

var importGistCmd = &cobra.Command{
	Use:   "gist <gist-url-or-id>",
	Short: "Import a pattern from a GitHub Gist",
	Long: `Import a pattern from a GitHub Gist.

The gist should contain either:
- A pattern.yaml file (mur pattern format)
- A README.md with description + code files

Examples:
  mur import gist https://gist.github.com/user/abc123
  mur import gist abc123
  mur import gist abc123 --share  # Import and share to community`,
	Args: cobra.ExactArgs(1),
	RunE: runImportGist,
}

var (
	importShare bool
	importName  string
)

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.AddCommand(importGistCmd)

	importGistCmd.Flags().BoolVar(&importShare, "share", false, "Share to community after import")
	importGistCmd.Flags().StringVarP(&importName, "name", "n", "", "Override pattern name")
}

// GistResponse represents GitHub Gist API response
type GistResponse struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Files       map[string]GistFile `json:"files"`
	Owner       *GistOwner          `json:"owner"`
	HTMLURL     string              `json:"html_url"`
}

type GistFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
}

type GistOwner struct {
	Login string `json:"login"`
}

func runImportGist(cmd *cobra.Command, args []string) error {
	gistInput := args[0]

	// Extract gist ID from URL or use as-is
	gistID := extractGistID(gistInput)
	if gistID == "" {
		return fmt.Errorf("invalid gist URL or ID: %s", gistInput)
	}

	fmt.Printf("📥 Fetching gist %s...\n", gistID)

	// Fetch gist from GitHub API
	gist, err := fetchGist(gistID)
	if err != nil {
		return fmt.Errorf("failed to fetch gist: %w", err)
	}

	fmt.Printf("   Found: %s\n", gist.Description)
	if gist.Owner != nil {
		fmt.Printf("   Author: @%s\n", gist.Owner.Login)
	}

	// Convert gist to pattern
	p, err := gistToPattern(gist)
	if err != nil {
		return fmt.Errorf("failed to convert gist to pattern: %w", err)
	}

	// Override name if specified
	if importName != "" {
		p.Name = importName
	}

	fmt.Printf("   Pattern: %s\n", p.Name)

	// Save pattern locally
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	if err := store.Create(p); err != nil {
		return fmt.Errorf("failed to save pattern: %w", err)
	}

	fmt.Printf("\n✓ Imported \"%s\" to ~/.mur/patterns/\n", p.Name)

	// Share if requested
	if importShare {
		fmt.Println("\nSharing to community...")
		// TODO: Call share API
		_ = cfg // Will use for share
		fmt.Println("  (Share functionality coming soon)")
	}

	fmt.Println("\nRun 'mur sync' to sync to your CLIs")

	return nil
}

func extractGistID(input string) string {
	// Full URL: https://gist.github.com/user/abc123
	// Or just: abc123

	re := regexp.MustCompile(`gist\.github\.com/[^/]+/([a-f0-9]+)`)
	if matches := re.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// Check if it's a valid gist ID (hex string)
	if matched, _ := regexp.MatchString(`^[a-f0-9]+$`, input); matched {
		return input
	}

	return ""
}

func fetchGist(gistID string) (*GistResponse, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("gist not found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var gist GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}

	return &gist, nil
}

func gistToPattern(gist *GistResponse) (*pattern.Pattern, error) {
	// Check for pattern.yaml first
	if file, ok := gist.Files["pattern.yaml"]; ok {
		return parsePatternYAML(file.Content)
	}
	if file, ok := gist.Files["pattern.yml"]; ok {
		return parsePatternYAML(file.Content)
	}

	// Otherwise, construct from files
	return constructPatternFromFiles(gist)
}

func parsePatternYAML(content string) (*pattern.Pattern, error) {
	var p pattern.Pattern
	if err := yaml.Unmarshal([]byte(content), &p); err != nil {
		return nil, fmt.Errorf("invalid pattern.yaml: %w", err)
	}
	return &p, nil
}

func constructPatternFromFiles(gist *GistResponse) (*pattern.Pattern, error) {
	p := &pattern.Pattern{
		SchemaVersion: 2,
	}

	// Use gist description as pattern description
	p.Description = gist.Description

	// Try to find a name from README or first file
	var contentParts []string
	var readmeContent string

	for filename, file := range gist.Files {
		lower := strings.ToLower(filename)

		if strings.HasPrefix(lower, "readme") {
			readmeContent = file.Content
			continue
		}

		// Add code files to content
		if file.Language != "" {
			lang := strings.ToLower(file.Language)
			contentParts = append(contentParts, fmt.Sprintf("```%s\n%s\n```", lang, file.Content))
		}
	}

	// Extract name from README or use gist description
	if readmeContent != "" {
		// Try to extract title from README
		lines := strings.Split(readmeContent, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				p.Name = strings.TrimPrefix(line, "# ")
				break
			}
		}
		// Use README as description if no separate description
		if p.Description == "" {
			p.Description = readmeContent
		}
	}

	// Fallback name
	if p.Name == "" {
		if gist.Description != "" {
			// Use first 50 chars of description as name
			name := gist.Description
			if len(name) > 50 {
				name = name[:47] + "..."
			}
			p.Name = name
		} else {
			p.Name = "Imported from Gist"
		}
	}

	// Combine content
	if len(contentParts) > 0 {
		p.Content = strings.Join(contentParts, "\n\n")
	} else if readmeContent != "" {
		p.Content = readmeContent
	} else {
		// Use any file content
		for _, file := range gist.Files {
			p.Content = file.Content
			break
		}
	}

	if p.Content == "" {
		return nil, fmt.Errorf("gist has no usable content")
	}

	// Add source metadata
	p.Learning = pattern.LearningMeta{
		ExtractedFrom: gist.HTMLURL,
	}

	return p, nil
}
