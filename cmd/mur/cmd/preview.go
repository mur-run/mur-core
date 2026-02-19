package cmd

import (
	"fmt"
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/security"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview <pattern-name>",
	Short: "Preview a pattern with security scan results",
	Long: `Show pattern content, trust level, hash status, and injection scan results.

Useful for inspecting community patterns before use.

Examples:
  mur preview go-error-handling
  mur preview community-pattern-name`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := pattern.DefaultStore()
		if err != nil {
			return fmt.Errorf("cannot open pattern store: %w", err)
		}

		p, err := store.LoadVerified(name)
		if err != nil {
			return fmt.Errorf("pattern not found: %w", err)
		}

		// Header
		fmt.Printf("Pattern: %s\n", p.Name)
		fmt.Println(strings.Repeat("=", len(p.Name)+9))
		fmt.Println("")

		// Metadata
		fmt.Printf("  ID:          %s\n", p.ID)
		if p.Description != "" {
			fmt.Printf("  Description: %s\n", p.Description)
		}
		fmt.Printf("  Trust Level: %s\n", p.Security.TrustLevel)
		fmt.Printf("  Source:      %s\n", valueOrDefault(p.Security.Source, "local"))
		fmt.Printf("  Risk:        %s\n", p.Security.Risk)
		fmt.Printf("  Status:      %s\n", p.Lifecycle.Status)
		fmt.Println("")

		// Hash verification
		fmt.Println("Integrity")
		fmt.Println("---------")
		if p.Security.Hash == "" {
			fmt.Println("  ⚠ No hash set")
		} else if p.VerifyHash() {
			fmt.Println("  ✓ Hash verified")
		} else {
			fmt.Println("  ✗ Hash mismatch — content may have been tampered with")
		}
		fmt.Println("")

		// Injection scan
		fmt.Println("Injection Scan")
		fmt.Println("--------------")
		injScanner := security.NewInjectionScanner()
		risk, findings := injScanner.Scan(p.Content)
		fmt.Printf("  Risk: %s\n", risk)
		fmt.Print(security.FormatInjectionFindings(findings))
		fmt.Println("")

		// Warnings
		if len(p.Security.Warnings) > 0 {
			fmt.Println("Warnings")
			fmt.Println("--------")
			for _, w := range p.Security.Warnings {
				fmt.Printf("  ⚠ %s\n", w)
			}
			fmt.Println("")
		}

		// Tags
		if len(p.Tags.Confirmed) > 0 || len(p.Tags.Inferred) > 0 {
			fmt.Println("Tags")
			fmt.Println("----")
			if len(p.Tags.Confirmed) > 0 {
				fmt.Printf("  Confirmed: %s\n", strings.Join(p.Tags.Confirmed, ", "))
			}
			for _, ts := range p.Tags.Inferred {
				fmt.Printf("  Inferred:  %s (%.0f%%)\n", ts.Tag, ts.Confidence*100)
			}
			fmt.Println("")
		}

		// Content
		fmt.Println("Content")
		fmt.Println("-------")
		contentLines := strings.Split(p.Content, "\n")
		for i, line := range contentLines {
			fmt.Printf("  %3d │ %s\n", i+1, line)
		}
		fmt.Println("")

		return nil
	},
}

func valueOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func init() {
	rootCmd.AddCommand(previewCmd)
}
