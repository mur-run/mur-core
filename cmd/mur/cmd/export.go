package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export patterns to various formats",
	Long: `Export patterns to JSON, YAML, or Markdown format.

Examples:
  mur export                           # Export all active patterns as YAML
  mur export --format json             # Export as JSON
  mur export --format md               # Export as Markdown
  mur export --tag backend             # Export patterns with 'backend' tag
  mur export --min-effectiveness 0.7   # Export high-effectiveness patterns
  mur export -o patterns.json          # Export to file`,
	RunE: runExport,
}

var (
	exportFormat         string
	exportOutput         string
	exportTag            string
	exportMinEffectiveness float64
	exportIncludeArchived bool
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "yaml", "Output format: yaml, json, md")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringVarP(&exportTag, "tag", "t", "", "Filter by tag")
	exportCmd.Flags().Float64Var(&exportMinEffectiveness, "min-effectiveness", 0.0, "Minimum effectiveness score (0.0-1.0)")
	exportCmd.Flags().BoolVar(&exportIncludeArchived, "include-archived", false, "Include archived patterns")
}

func runExport(cmd *cobra.Command, args []string) error {
	store, err := pattern.DefaultStore()
	if err != nil {
		return fmt.Errorf("cannot access pattern store: %w", err)
	}

	// Get patterns
	var patterns []pattern.Pattern
	if exportTag != "" {
		patterns, err = store.GetByTag(exportTag)
	} else if exportIncludeArchived {
		patterns, err = store.List()
	} else {
		patterns, err = store.GetActive()
	}
	if err != nil {
		return fmt.Errorf("cannot load patterns: %w", err)
	}

	// Filter by effectiveness
	if exportMinEffectiveness > 0 {
		var filtered []pattern.Pattern
		for _, p := range patterns {
			if p.Learning.Effectiveness >= exportMinEffectiveness {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	if len(patterns) == 0 {
		fmt.Println("No patterns found matching criteria.")
		return nil
	}

	// Format output
	var output string
	switch strings.ToLower(exportFormat) {
	case "json":
		data, err := json.MarshalIndent(patterns, "", "  ")
		if err != nil {
			return fmt.Errorf("cannot marshal to JSON: %w", err)
		}
		output = string(data)

	case "yaml", "yml":
		data, err := yaml.Marshal(patterns)
		if err != nil {
			return fmt.Errorf("cannot marshal to YAML: %w", err)
		}
		output = string(data)

	case "md", "markdown":
		output = formatMarkdown(patterns)

	default:
		return fmt.Errorf("unknown format: %s (use yaml, json, or md)", exportFormat)
	}

	// Write output
	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}
		fmt.Printf("Exported %d patterns to %s\n", len(patterns), exportOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}

func formatMarkdown(patterns []pattern.Pattern) string {
	var sb strings.Builder

	sb.WriteString("# Learned Patterns\n\n")
	sb.WriteString(fmt.Sprintf("Total: %d patterns\n\n", len(patterns)))

	for _, p := range patterns {
		sb.WriteString(fmt.Sprintf("## %s\n\n", p.Name))

		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
		}

		// Tags
		var tags []string
		for _, t := range p.Tags.Confirmed {
			tags = append(tags, t)
		}
		for _, ts := range p.Tags.Inferred {
			if ts.Confidence >= 0.5 {
				tags = append(tags, fmt.Sprintf("%s (%.0f%%)", ts.Tag, ts.Confidence*100))
			}
		}
		if len(tags) > 0 {
			sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(tags, ", ")))
		}

		// Metadata
		sb.WriteString(fmt.Sprintf("**Effectiveness:** %.0f%% | **Used:** %d times\n\n",
			p.Learning.Effectiveness*100, p.Learning.UsageCount))

		// Content
		sb.WriteString("```\n")
		sb.WriteString(p.Content)
		if !strings.HasSuffix(p.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
