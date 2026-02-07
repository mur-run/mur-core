package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <pattern-name>",
	Short: "Copy pattern content to clipboard",
	Long: `Copy a pattern's content to the system clipboard.

Useful for quickly pasting patterns into chats or documents.

Examples:
  mur copy go-error-handling    # Copy to clipboard
  mur copy --yaml my-pattern    # Copy full YAML`,
	Args: cobra.ExactArgs(1),
	RunE: runCopy,
}

var copyYAML bool

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolVar(&copyYAML, "yaml", false, "Copy full YAML instead of just content")
}

func runCopy(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternPath := filepath.Join(home, ".mur", "patterns", patternName+".yaml")
	content, err := os.ReadFile(patternPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pattern not found: %s", patternName)
		}
		return err
	}

	var toCopy string
	if copyYAML {
		toCopy = string(content)
	} else {
		// Extract just the content field
		toCopy = extractContent(string(content))
	}

	if err := copyToClipboard(toCopy); err != nil {
		// If clipboard fails, just print
		fmt.Println(toCopy)
		fmt.Fprintln(os.Stderr, "\n(Clipboard not available, printed to stdout)")
		return nil
	}

	lines := countLines(toCopy)
	fmt.Printf("✅ Copied %s to clipboard (%d lines)\n", patternName, lines)
	return nil
}

func extractContent(yaml string) string {
	// Simple extraction - find content: | block
	inContent := false
	var content string
	indent := 0

	for _, line := range splitLines(yaml) {
		if !inContent {
			if len(line) >= 8 && line[:8] == "content:" {
				inContent = true
				continue
			}
		} else {
			// Check if we've left the content block
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				break
			}
			if indent == 0 && len(line) > 0 {
				// Determine indent level
				for i, c := range line {
					if c != ' ' {
						indent = i
						break
					}
				}
			}
			// Remove indent and add to content
			if len(line) >= indent {
				content += line[indent:] + "\n"
			} else {
				content += line + "\n"
			}
		}
	}

	return content
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := pipe.Write([]byte(text)); err != nil {
		return err
	}

	if err := pipe.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}
