package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <pattern-name>",
	Short: "Edit a pattern in your editor",
	Long: `Open a pattern file in your default editor.

Uses $EDITOR environment variable, falls back to vim/nano.

Examples:
  mur edit go-error-handling     # Edit pattern
  EDITOR=code mur edit my-pattern  # Use VS Code`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternPath := filepath.Join(home, ".mur", "patterns", patternName+".yaml")

	// Check if pattern exists
	if _, err := os.Stat(patternPath); os.IsNotExist(err) {
		return fmt.Errorf("pattern not found: %s\nUse 'mur learn list' to see available patterns", patternName)
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable")
	}

	// Open editor
	editorCmd := exec.Command(editor, patternPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Pattern saved:", patternName)
	fmt.Println("   Run 'mur lint", patternName+"' to validate")
	fmt.Println("   Run 'mur sync' to sync to AI tools")

	return nil
}
