package cmd

import (
	"fmt"
	"strings"

	"github.com/mur-run/mur-core/internal/core/classifier"
	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
)

var classifyCmd = &cobra.Command{
	Use:   "classify [prompt]",
	Short: "Classify a prompt and find relevant patterns",
	Long: `Classify analyzes a prompt to determine relevant domains and finds matching patterns.

The classifier uses multiple strategies:
  - File patterns: Detects language/framework from file extensions
  - Keywords: Matches domain-specific keywords
  - Rules: Applies predefined classification rules

Examples:
  # Classify a prompt
  mur classify "How do I handle errors in Swift?"

  # Classify with file context
  mur classify "Refactor this function" --file main.swift

  # Show only classification (no pattern matching)
  mur classify "Debug this API" --domains-only

  # Show top N patterns
  mur classify "Write unit tests" --limit 5`,
	RunE: runClassify,
}

var (
	classifyFile        string
	classifyDomainsOnly bool
	classifyLimit       int
)

func init() {
	classifyCmd.Hidden = true
	rootCmd.AddCommand(classifyCmd)
	classifyCmd.Flags().StringVarP(&classifyFile, "file", "f", "", "Current file context")
	classifyCmd.Flags().BoolVar(&classifyDomainsOnly, "domains-only", false, "Show only domains, skip pattern matching")
	classifyCmd.Flags().IntVarP(&classifyLimit, "limit", "n", 5, "Maximum number of patterns to show")
}

func runClassify(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a prompt to classify")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	prompt := strings.Join(args, " ")

	// Build input
	input := classifier.ClassifyInput{
		Content:     prompt,
		CurrentFile: classifyFile,
	}

	// Create classifier
	hybrid := classifier.NewHybridClassifier()
	domains := hybrid.Classify(input)

	// Print domains
	fmt.Println("📊 Classification Results")
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Input: %q\n", prompt)
	if classifyFile != "" {
		fmt.Printf("File:  %s\n", classifyFile)
	}
	fmt.Println()

	if len(domains) == 0 {
		fmt.Println("No domains detected")
	} else {
		fmt.Println("🏷️  Detected Domains:")
		for _, d := range domains {
			bar := makeProgressBar(d.Confidence, 20)
			fmt.Printf("   %-15s %s %.0f%%\n", d.Domain, bar, d.Confidence*100)
			if len(d.Signals) > 0 && verbose {
				for _, s := range d.Signals {
					fmt.Printf("                    └─ %s\n", s)
				}
			}
		}
	}

	if classifyDomainsOnly {
		return nil
	}

	// Find relevant patterns
	fmt.Println()
	fmt.Println("📚 Relevant Patterns:")

	store, err := pattern.DefaultStore()
	if err != nil {
		return err
	}

	retriever := classifier.NewRetriever(store)
	matches, err := retriever.Retrieve(input, classifyLimit)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Println("   No matching patterns found")
		fmt.Println()
		fmt.Println("💡 Tip: Add patterns with 'mur learn add <name>'")
	} else {
		for i, m := range matches {
			bar := makeProgressBar(m.Score, 15)
			fmt.Printf("   %d. %-30s %s %.0f%%\n", i+1, m.Pattern.Name, bar, m.Score*100)
			if verbose && len(m.Reasons) > 0 {
				reasons := strings.Join(m.Reasons, ", ")
				fmt.Printf("      └─ %s\n", reasons)
			}
		}
	}

	return nil
}

func makeProgressBar(value float64, width int) string {
	filled := int(value * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}
