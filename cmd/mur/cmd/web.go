package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web [page]",
	Short: "Open mur web pages",
	Long: `Open mur documentation and resources in your browser.

Pages:
  docs      Documentation (default)
  github    GitHub repository
  issues    GitHub issues
  releases  GitHub releases

Examples:
  mur web           # Open docs
  mur web github    # Open GitHub repo
  mur web issues    # Open issues page`,
	RunE: runWeb,
}

func init() {
	rootCmd.AddCommand(webCmd)
}

var webPages = map[string]string{
	"docs":     "https://github.com/mur-run/mur-core#readme",
	"github":   "https://github.com/mur-run/mur-core",
	"issues":   "https://github.com/mur-run/mur-core/issues",
	"releases": "https://github.com/mur-run/mur-core/releases",
	"new":      "https://github.com/mur-run/mur-core/issues/new",
}

func runWeb(cmd *cobra.Command, args []string) error {
	page := "docs"
	if len(args) > 0 {
		page = args[0]
	}

	url, ok := webPages[page]
	if !ok {
		fmt.Println("Available pages:")
		for name := range webPages {
			fmt.Printf("  %s\n", name)
		}
		return fmt.Errorf("unknown page: %s", page)
	}

	fmt.Printf("Opening %s...\n", url)
	openBrowser(url)
	return nil
}
