package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

var lintCmd = &cobra.Command{
	Use:   "lint [pattern-name]",
	Short: "Validate patterns for issues",
	Long: `Lint validates patterns for common issues including:

  - Required fields (name, content)
  - Schema version compatibility
  - Content length limits
  - Tags configuration
  - Security issues (prompt injection detection)
  - Hash integrity verification

Examples:
  # Lint all patterns
  mur lint

  # Lint a specific pattern
  mur lint my-pattern

  # Show only errors (no warnings/info)
  mur lint --errors-only`,
	RunE: runLint,
}

var (
	lintErrorsOnly bool
	lintJSON       bool
)

func init() {
	lintCmd.Hidden = true
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().BoolVar(&lintErrorsOnly, "errors-only", false, "Show only errors, hide warnings and info")
	lintCmd.Flags().BoolVar(&lintJSON, "json", false, "Output in JSON format")
}

func runLint(cmd *cobra.Command, args []string) error {
	store, err := pattern.DefaultStore()
	if err != nil {
		return err
	}

	linter := pattern.NewLinter()

	if len(args) > 0 {
		// Lint specific pattern
		return lintPattern(store, linter, args[0])
	}

	// Lint all patterns
	return lintAll(store, linter)
}

func lintPattern(store *pattern.Store, linter *pattern.Linter, name string) error {
	p, err := store.Get(name)
	if err != nil {
		return err
	}

	issues := linter.Lint(p)
	if len(issues) == 0 {
		fmt.Printf("✅ %s: no issues found\n", name)
		return nil
	}

	printIssues(name, issues)

	// Return error if there are errors
	for _, issue := range issues {
		if issue.Severity == pattern.SeverityError {
			os.Exit(1)
		}
	}

	return nil
}

func lintAll(store *pattern.Store, linter *pattern.Linter) error {
	result, err := linter.LintAll(store)
	if err != nil {
		return err
	}

	if result.TotalPatterns == 0 {
		fmt.Println("No patterns found")
		return nil
	}

	// Print summary header
	fmt.Printf("\n🔍 Linting %d patterns...\n\n", result.TotalPatterns)

	// Group issues by pattern
	issuesByPattern := make(map[string][]pattern.LintIssue)
	for _, issue := range result.Issues {
		issuesByPattern[issue.Pattern] = append(issuesByPattern[issue.Pattern], issue)
	}

	// Print issues for each pattern
	for name, issues := range issuesByPattern {
		printIssues(name, issues)
	}

	// Print summary
	fmt.Println("\n" + "─────────────────────────────────────")
	fmt.Printf("📊 Summary: %d patterns checked\n", result.TotalPatterns)
	fmt.Printf("   ✅ Clean: %d\n", result.CleanPatterns)
	if result.ErrorCount > 0 {
		fmt.Printf("   ❌ Errors: %d\n", result.ErrorCount)
	}
	if result.WarningCount > 0 {
		fmt.Printf("   ⚠️  Warnings: %d\n", result.WarningCount)
	}
	if result.InfoCount > 0 && !lintErrorsOnly {
		fmt.Printf("   ℹ️  Info: %d\n", result.InfoCount)
	}

	if result.IsClean() {
		fmt.Println("\n✅ All patterns are clean!")
	} else if result.HasErrors() {
		fmt.Println("\n❌ Found errors that should be fixed")
		os.Exit(1)
	} else {
		fmt.Println("\n⚠️  Found warnings - consider reviewing")
	}

	return nil
}

func printIssues(name string, issues []pattern.LintIssue) {
	fmt.Printf("📄 %s\n", name)

	for _, issue := range issues {
		if lintErrorsOnly && issue.Severity != pattern.SeverityError {
			continue
		}

		var icon string
		switch issue.Severity {
		case pattern.SeverityError:
			icon = "❌"
		case pattern.SeverityWarning:
			icon = "⚠️"
		case pattern.SeverityInfo:
			icon = "ℹ️"
		}

		fmt.Printf("   %s [%s] %s\n", icon, issue.Field, issue.Message)
	}
	fmt.Println()
}
