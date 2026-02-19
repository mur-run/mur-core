package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/core/suggest"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Extract pattern suggestions from session files",
	Long: `Analyze session transcripts and extract recurring patterns.

Scans markdown files, logs, and transcripts to find:
- Recurring code patterns
- Common workflows
- Repeated solutions

Suggestions can be reviewed and accepted as new patterns.

Examples:
  mur suggest scan ~/sessions/     # Scan directory for patterns
  mur suggest scan .               # Scan current directory
  mur suggest accept <name>        # Accept a suggestion as pattern
  mur suggest list                 # List pending suggestions`,
}

var suggestScanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan files for pattern suggestions",
	Args:  cobra.MaximumNArgs(1),
	RunE:  suggestScanExecute,
}

var suggestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending pattern suggestions",
	RunE:  suggestListExecute,
}

var suggestAcceptCmd = &cobra.Command{
	Use:   "accept <name>",
	Short: "Accept a suggestion as a new pattern",
	Args:  cobra.ExactArgs(1),
	RunE:  suggestAcceptExecute,
}

var suggestRejectCmd = &cobra.Command{
	Use:   "reject <name>",
	Short: "Reject a suggestion",
	Args:  cobra.ExactArgs(1),
	RunE:  suggestRejectExecute,
}

// Pending suggestions stored in memory for the session
var pendingSuggestions []suggest.Suggestion

func suggestScanExecute(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Expand path
	if strings.HasPrefix(dir, "~") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[1:])
	}

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	suggestDir := filepath.Join(home, ".mur", "suggestions")
	store := pattern.NewStore(patternsDir)

	cfg := suggest.DefaultExtractorConfig()
	extractor := suggest.NewExtractor(store, suggestDir, cfg)

	fmt.Printf("🔍 Scanning %s for patterns...\n\n", dir)

	result, err := extractor.Extract(dir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("Scanned %d files\n", result.FilesRead)
	fmt.Printf("Found %d suggestions\n\n", len(result.Suggestions))

	if len(result.Suggestions) == 0 {
		fmt.Println("No new patterns found.")
		return nil
	}

	// Store for later acceptance
	pendingSuggestions = result.Suggestions

	// Display suggestions
	for i, s := range result.Suggestions {
		confBar := makeBar(s.Confidence, 10)
		fmt.Printf("%d. %s  %s %.0f%%\n", i+1, s.Name, confBar, s.Confidence*100)
		fmt.Printf("   %s\n", truncateStr(s.Description, 60))
		fmt.Printf("   Tags: %s | Sources: %d | %s\n", strings.Join(s.Tags, ", "), len(s.Sources), s.Reason)
		fmt.Println()
	}

	// Interactive mode if terminal
	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		return interactiveAccept(extractor, result.Suggestions)
	}

	fmt.Println("Run 'mur suggest accept <name>' to add a pattern")
	fmt.Println("Run 'mur suggest scan --interactive' for interactive mode")

	return nil
}

func interactiveAccept(extractor *suggest.Extractor, suggestions []suggest.Suggestion) error {
	reader := bufio.NewReader(os.Stdin)

	for _, s := range suggestions {
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("📋 %s (%.0f%% confidence)\n", s.Name, s.Confidence*100)
		fmt.Printf("   %s\n\n", s.Description)

		// Show preview
		preview := s.Content
		if len(preview) > 300 {
			preview = preview[:300] + "..."
		}
		fmt.Println("Content preview:")
		fmt.Println("```")
		fmt.Println(preview)
		fmt.Println("```")
		fmt.Println()

		fmt.Print("Accept this pattern? [y/n/q]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			p, err := extractor.Accept(s)
			if err != nil {
				fmt.Printf("❌ Failed to create pattern: %v\n", err)
			} else {
				fmt.Printf("✅ Created pattern: %s\n", p.Name)
			}
		case "q", "quit":
			fmt.Println("Stopped.")
			return nil
		default:
			fmt.Println("Skipped.")
		}
		fmt.Println()
	}

	return nil
}

func suggestListExecute(cmd *cobra.Command, args []string) error {
	if len(pendingSuggestions) == 0 {
		fmt.Println("No pending suggestions.")
		fmt.Println("Run 'mur suggest scan <directory>' to find patterns.")
		return nil
	}

	fmt.Printf("Pending Suggestions (%d)\n", len(pendingSuggestions))
	fmt.Println("========================")

	for i, s := range pendingSuggestions {
		fmt.Printf("%d. %s (%.0f%%)\n", i+1, s.Name, s.Confidence*100)
		fmt.Printf("   %s\n", truncateStr(s.Description, 60))
	}

	return nil
}

func suggestAcceptExecute(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Find in pending
	var found *suggest.Suggestion
	for i := range pendingSuggestions {
		if pendingSuggestions[i].Name == name || fmt.Sprintf("%d", i+1) == name {
			found = &pendingSuggestions[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("suggestion not found: %s", name)
	}

	home, _ := os.UserHomeDir()
	patternsDir := filepath.Join(home, ".mur", "patterns")
	suggestDir := filepath.Join(home, ".mur", "suggestions")
	store := pattern.NewStore(patternsDir)

	cfg := suggest.DefaultExtractorConfig()
	extractor := suggest.NewExtractor(store, suggestDir, cfg)

	p, err := extractor.Accept(*found)
	if err != nil {
		return fmt.Errorf("failed to create pattern: %w", err)
	}

	fmt.Printf("✅ Created pattern: %s\n", p.Name)
	fmt.Printf("   Tags: %v\n", p.Tags.Inferred)

	// Remove from pending
	for i := range pendingSuggestions {
		if pendingSuggestions[i].Hash == found.Hash {
			pendingSuggestions = append(pendingSuggestions[:i], pendingSuggestions[i+1:]...)
			break
		}
	}

	return nil
}

func suggestRejectExecute(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Find and remove from pending
	for i := range pendingSuggestions {
		if pendingSuggestions[i].Name == name || fmt.Sprintf("%d", i+1) == name {
			pendingSuggestions = append(pendingSuggestions[:i], pendingSuggestions[i+1:]...)
			fmt.Printf("🗑️ Rejected: %s\n", name)
			return nil
		}
	}

	return fmt.Errorf("suggestion not found: %s", name)
}

func init() {
	suggestCmd.Hidden = true
	rootCmd.AddCommand(suggestCmd)
	suggestCmd.AddCommand(suggestScanCmd)
	suggestCmd.AddCommand(suggestListCmd)
	suggestCmd.AddCommand(suggestAcceptCmd)
	suggestCmd.AddCommand(suggestRejectCmd)

	suggestScanCmd.Flags().BoolP("interactive", "i", false, "Interactive mode - review each suggestion")
}
