package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/security"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify pattern integrity",
	Long: `Check all patterns for hash integrity and injection risks.

Examples:
  mur verify          # Check all patterns
  mur verify --fix    # Recalculate hashes for mismatched patterns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")

		store, err := pattern.DefaultStore()
		if err != nil {
			return fmt.Errorf("cannot open pattern store: %w", err)
		}

		patterns, err := store.List()
		if err != nil {
			return fmt.Errorf("cannot list patterns: %w", err)
		}

		if len(patterns) == 0 {
			fmt.Println("No patterns found.")
			return nil
		}

		injScanner := security.NewInjectionScanner()

		fmt.Println("Pattern Integrity Check")
		fmt.Println("=======================")
		fmt.Println("")

		hashOK := 0
		hashMismatch := 0
		hashMissing := 0
		injectionWarnings := 0
		fixed := 0

		for _, p := range patterns {
			issues := []string{}

			// Check hash
			if p.Security.Hash == "" {
				hashMissing++
				issues = append(issues, "no hash")
				if fix {
					p.UpdateHash()
					if err := store.Update(&p); err == nil {
						fixed++
						issues[len(issues)-1] = "no hash (fixed)"
					}
				}
			} else if !p.VerifyHash() {
				hashMismatch++
				issues = append(issues, "hash mismatch")
				if fix {
					p.UpdateHash()
					if err := store.Update(&p); err == nil {
						fixed++
						issues[len(issues)-1] = "hash mismatch (fixed)"
					}
				}
			} else {
				hashOK++
			}

			// Check for injection
			risk, findings := injScanner.Scan(p.Content)
			if risk != security.InjectionRiskLow || len(findings) > 0 {
				injectionWarnings++
				issues = append(issues, fmt.Sprintf("injection risk: %s (%d findings)", risk, len(findings)))
			}

			// Print status
			if len(issues) > 0 {
				fmt.Printf("  ⚠ %-30s  %s  [%s]\n", p.Name, p.Security.TrustLevel, joinIssues(issues))
			} else {
				fmt.Printf("  ✓ %-30s  %s\n", p.Name, p.Security.TrustLevel)
			}
		}

		fmt.Println("")
		fmt.Printf("Results: %d OK, %d mismatch, %d missing hash, %d injection warnings\n",
			hashOK, hashMismatch, hashMissing, injectionWarnings)
		if fix && fixed > 0 {
			fmt.Printf("Fixed: %d patterns\n", fixed)
		}

		return nil
	},
}

func joinIssues(issues []string) string {
	result := ""
	for i, issue := range issues {
		if i > 0 {
			result += ", "
		}
		result += issue
	}
	return result
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().Bool("fix", false, "Recalculate hashes for mismatched patterns")
}
