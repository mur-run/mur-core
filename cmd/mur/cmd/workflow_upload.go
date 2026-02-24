package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
)

var workflowsUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload session data and get a shareable workflow URL",
	Long: `Upload session analysis data to the workflow viewer.

Reads JSON session data from stdin or a file, uploads it to the
mur workflow API, and prints the shareable URL.

Examples:
  cat session.json | mur workflows upload
  mur workflows upload --file ~/.mur/last-session.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")

		var data []byte
		var err error

		if filePath != "" {
			data, err = os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		} else {
			// Read from stdin
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
		}

		if len(data) == 0 {
			return fmt.Errorf("no data provided; pipe JSON to stdin or use --file")
		}

		// Get API URL from config (or use default)
		apiURL := cloud.DefaultWorkflowAPIURL
		if cfg, err := config.Load(); err == nil {
			if u := cfg.Server.URL; u != "" {
				// Use the configured server as base, but the workflow API
				// is separate — only override if explicitly set for workflows
				_ = u
			}
		}

		url, err := cloud.UploadSessionData(apiURL, data)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		fmt.Println(url)
		return nil
	},
}

func init() {
	workflowsCmd.AddCommand(workflowsUploadCmd)

	workflowsUploadCmd.Flags().String("file", "", "Path to session data JSON file (reads from stdin if omitted)")
}
