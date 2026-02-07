package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import <file-or-url>",
	Short: "Import patterns from file or URL",
	Long: `Import patterns from a YAML file or URL.

Supports:
  - Local YAML files
  - Remote URLs (https://)
  - GitHub raw URLs
  - Multiple patterns in one file

Examples:
  mur import patterns.yaml
  mur import https://example.com/patterns.yaml
  mur import ./my-patterns/*.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runImport,
}

var (
	importForce bool
	importDryRun bool
)

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Overwrite existing patterns")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Show what would be imported")
}

func runImport(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		return fmt.Errorf("failed to create patterns directory: %w", err)
	}

	imported := 0
	skipped := 0
	errors := 0

	for _, arg := range args {
		// Handle glob patterns
		if strings.Contains(arg, "*") {
			matches, err := filepath.Glob(arg)
			if err != nil {
				fmt.Printf("❌ Invalid glob pattern: %s\n", arg)
				errors++
				continue
			}
			for _, match := range matches {
				i, s, e := importFile(match, patternsDir)
				imported += i
				skipped += s
				errors += e
			}
			continue
		}

		// Handle URLs
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			i, s, e := importURL(arg, patternsDir)
			imported += i
			skipped += s
			errors += e
			continue
		}

		// Handle local files
		i, s, e := importFile(arg, patternsDir)
		imported += i
		skipped += s
		errors += e
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if importDryRun {
		fmt.Printf("Would import: %d patterns\n", imported)
	} else {
		fmt.Printf("✅ Imported: %d patterns\n", imported)
	}
	if skipped > 0 {
		fmt.Printf("⏭️  Skipped: %d (already exist, use --force to overwrite)\n", skipped)
	}
	if errors > 0 {
		fmt.Printf("❌ Errors: %d\n", errors)
	}

	if imported > 0 && !importDryRun {
		fmt.Println()
		fmt.Println("Next: mur sync  # to sync to AI tools")
	}

	return nil
}

func importFile(path string, patternsDir string) (imported, skipped, errors int) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Cannot read %s: %v\n", path, err)
		return 0, 0, 1
	}

	return importContent(content, path, patternsDir)
}

func importURL(url string, patternsDir string) (imported, skipped, errors int) {
	fmt.Printf("📥 Fetching %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Cannot fetch %s: %v\n", url, err)
		return 0, 0, 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ HTTP %d for %s\n", resp.StatusCode, url)
		return 0, 0, 1
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Cannot read response: %v\n", err)
		return 0, 0, 1
	}

	return importContent(content, url, patternsDir)
}

func importContent(content []byte, source string, patternsDir string) (imported, skipped, errors int) {
	// Try to parse as single pattern first
	var single map[string]interface{}
	if err := yaml.Unmarshal(content, &single); err == nil {
		if name, ok := single["name"].(string); ok && name != "" {
			return importSinglePattern(content, name, source, patternsDir)
		}
	}

	// Try to parse as multiple patterns (YAML documents)
	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if name, ok := doc["name"].(string); ok && name != "" {
			docContent, _ := yaml.Marshal(doc)
			i, s, e := importSinglePattern(docContent, name, source, patternsDir)
			imported += i
			skipped += s
			errors += e
		}
	}

	return
}

func importSinglePattern(content []byte, name, source, patternsDir string) (imported, skipped, errors int) {
	destPath := filepath.Join(patternsDir, name+".yaml")

	// Check if exists
	if _, err := os.Stat(destPath); err == nil && !importForce {
		fmt.Printf("⏭️  %s (already exists)\n", name)
		return 0, 1, 0
	}

	if importDryRun {
		fmt.Printf("📄 Would import: %s\n", name)
		return 1, 0, 0
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		fmt.Printf("❌ Cannot write %s: %v\n", name, err)
		return 0, 0, 1
	}

	fmt.Printf("✅ Imported: %s\n", name)
	return 1, 0, 0
}
