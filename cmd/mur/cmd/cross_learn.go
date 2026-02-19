package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/core/suggest"
	"github.com/mur-run/mur-core/internal/learn"
)

var crossLearnCmd = &cobra.Command{
	Use:   "cross-learn",
	Short: "Learn patterns from AI CLI session histories",
	Long: `Extract patterns from session histories of various AI CLI tools.

Supported CLI tools:
  - Claude Code (~/.claude/projects/*)
  - Gemini CLI (~/.gemini/history/*)
  - Codex (~/.codex/history/*)
  - Aider (~/.aider/history/*)
  - Continue (~/.continue/sessions/*)

The learner extracts:
  - Problem-solution patterns
  - Code patterns
  - Workflow patterns

Examples:
  mur cross-learn scan              # Scan all CLI sources
  mur cross-learn scan --source claude  # Scan only Claude
  mur cross-learn status            # Show available sources`,
}

var crossLearnScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan CLI session histories for patterns",
	RunE:  crossLearnScanExecute,
}

var crossLearnStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of CLI sources",
	RunE:  crossLearnStatusExecute,
}

func crossLearnScanExecute(cmd *cobra.Command, args []string) error {
	source, _ := cmd.Flags().GetString("source")
	interactive, _ := cmd.Flags().GetBool("interactive")

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	learner := learn.NewCrossCLILearner(store)

	var results []learn.LearnResult
	var err error

	if source != "" {
		// Learn from specific source
		fmt.Printf("🔍 Scanning %s sessions...\n\n", source)
		result, err := learner.LearnFromSource(source)
		if err != nil {
			return err
		}
		results = []learn.LearnResult{*result}
	} else {
		// Learn from all sources
		fmt.Println("🔍 Scanning all CLI session histories...")
		fmt.Println()
		results, err = learner.LearnFromAll()
		if err != nil {
			return err
		}
	}

	// Display results
	totalSuggestions := 0
	var allSuggestions []suggest.Suggestion

	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("⚠️  %s: %v\n", r.Source, r.Error)
			continue
		}

		fmt.Printf("📚 %s\n", r.Source)
		fmt.Printf("   Files: %d | Entries: %d | Suggestions: %d\n",
			r.FilesRead, r.Entries, len(r.Suggestions))

		totalSuggestions += len(r.Suggestions)
		allSuggestions = append(allSuggestions, r.Suggestions...)
	}

	fmt.Println()

	if totalSuggestions == 0 {
		fmt.Println("No new patterns found.")
		return nil
	}

	fmt.Printf("Found %d pattern suggestions\n\n", totalSuggestions)

	// Show suggestions
	for i, s := range allSuggestions {
		if i >= 10 {
			fmt.Printf("... and %d more\n", len(allSuggestions)-10)
			break
		}

		confBar := makeBar(s.Confidence, 10)
		fmt.Printf("%d. %s  %s %.0f%%\n", i+1, s.Name, confBar, s.Confidence*100)
		fmt.Printf("   %s\n", truncateStr(s.Description, 60))
		fmt.Printf("   Tags: %s | %s\n", strings.Join(s.Tags, ", "), s.Reason)
		fmt.Println()
	}

	// Interactive mode
	if interactive && len(allSuggestions) > 0 {
		return interactiveAcceptCrossLearn(store, allSuggestions)
	}

	fmt.Println("Run with --interactive to review and accept patterns")
	return nil
}

func interactiveAcceptCrossLearn(store *pattern.Store, suggestions []suggest.Suggestion) error {
	home, _ := os.UserHomeDir()
	suggestDir := filepath.Join(home, ".mur", "suggestions")
	extractor := suggest.NewExtractor(store, suggestDir, suggest.DefaultExtractorConfig())

	return interactiveAccept(extractor, suggestions)
}

func crossLearnStatusExecute(cmd *cobra.Command, args []string) error {
	sources := learn.DefaultCLISources()

	fmt.Println("Cross-CLI Learning Sources")
	fmt.Println("==========================")
	fmt.Println()

	for _, s := range sources {
		status := "✗ not found"
		if info, err := os.Stat(s.SessionDir); err == nil && info.IsDir() {
			// Count files
			pattern := filepath.Join(s.SessionDir, s.FilePattern)
			files, _ := filepath.Glob(pattern)
			status = fmt.Sprintf("✓ %d session files", len(files))
		}

		fmt.Printf("%-15s %s\n", s.Name+":", status)
		fmt.Printf("                %s\n", s.SessionDir)
		fmt.Println()
	}

	return nil
}

func init() {
	crossLearnCmd.Hidden = true
	rootCmd.AddCommand(crossLearnCmd)
	crossLearnCmd.AddCommand(crossLearnScanCmd)
	crossLearnCmd.AddCommand(crossLearnStatusCmd)

	crossLearnScanCmd.Flags().StringP("source", "s", "", "Specific CLI source to scan")
	crossLearnScanCmd.Flags().BoolP("interactive", "i", false, "Interactive mode to review suggestions")
}
