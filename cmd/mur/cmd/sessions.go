package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/sessions"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "View session history",
	Long: `View past mur session records.

Commands:
  mur sessions list           List past sessions
  mur sessions save           Save a session record (from stdin or --json)`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List past sessions",
	Long: `Show past sessions with time, project, patterns count, and workflow URL.

Examples:
  mur sessions list
  mur sessions list --limit 10
  mur sessions list --project myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		project, _ := cmd.Flags().GetString("project")

		records, err := sessions.ListSessions()
		if err != nil {
			return fmt.Errorf("load sessions: %w", err)
		}

		if project != "" {
			var filtered []sessions.SessionRecord
			for _, r := range records {
				if strings.EqualFold(r.Project, project) {
					filtered = append(filtered, r)
				}
			}
			records = filtered
		}

		if limit > 0 && len(records) > limit {
			records = records[:limit]
		}

		if len(records) == 0 {
			fmt.Println("No sessions found.")
			return nil
		}

		// Print table header
		fmt.Printf("%-10s | %-18s | %-12s | %-8s | %s\n",
			"ID", "Time", "Project", "Patterns", "URL")
		fmt.Printf("%-10s-+-%-18s-+-%-12s-+-%-8s-+-%s\n",
			strings.Repeat("-", 10),
			strings.Repeat("-", 18),
			strings.Repeat("-", 12),
			strings.Repeat("-", 8),
			strings.Repeat("-", 20))

		for _, r := range records {
			id := r.ID
			if len(id) > 8 {
				id = id[:8]
			}

			proj := r.Project
			if len(proj) > 12 {
				proj = proj[:12]
			}

			timeStr := r.StartTime.Format("2006-01-02 15:04")

			url := r.URL
			if url == "" {
				url = "(local only)"
			}

			fmt.Printf("%-10s | %-18s | %-12s | %-8d | %s\n",
				id, timeStr, proj, r.Patterns, url)
		}

		return nil
	},
}

var sessionsSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save a session record",
	Long: `Save a session record to history.

Input via flags (--id required) or JSON (stdin / --json).

Examples:
  mur sessions save --id abc --project myapp --goal "fix bug" --patterns 5 --tool claude
  echo '{"id":"abc",...}' | mur sessions save
  mur sessions save --json '{"id":"abc",...}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		idFlag, _ := cmd.Flags().GetString("id")

		var record sessions.SessionRecord

		if idFlag != "" {
			// Build record from flags
			record.ID = idFlag
			record.Project, _ = cmd.Flags().GetString("project")
			record.Goal, _ = cmd.Flags().GetString("goal")
			record.Patterns, _ = cmd.Flags().GetInt("patterns")
			record.Tool, _ = cmd.Flags().GetString("tool")
			record.URL, _ = cmd.Flags().GetString("url")

			startStr, _ := cmd.Flags().GetString("start-time")
			endStr, _ := cmd.Flags().GetString("end-time")

			if startStr != "" {
				t, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return fmt.Errorf("parse start-time: %w", err)
				}
				record.StartTime = t
			}
			if endStr != "" {
				t, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return fmt.Errorf("parse end-time: %w", err)
				}
				record.EndTime = t
			}
		} else {
			// Existing JSON input path
			jsonStr, _ := cmd.Flags().GetString("json")

			var data []byte
			var err error

			if jsonStr != "" {
				data = []byte(jsonStr)
			} else {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
			}

			if len(data) == 0 {
				return fmt.Errorf("no data provided; use --id with flags, pipe JSON to stdin, or use --json")
			}

			if err := json.Unmarshal(data, &record); err != nil {
				return fmt.Errorf("parse session record: %w", err)
			}
		}

		if record.ID == "" {
			return fmt.Errorf("session record must have an id")
		}

		if err := sessions.SaveSession(record); err != nil {
			return fmt.Errorf("save session: %w", err)
		}

		fmt.Printf("Session %s saved to history.\n", record.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sessionsCmd)

	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsListCmd.Flags().Int("limit", 0, "Maximum number of sessions to show")
	sessionsListCmd.Flags().String("project", "", "Filter by project name")

	sessionsCmd.AddCommand(sessionsSaveCmd)
	sessionsSaveCmd.Flags().String("json", "", "JSON session record (reads from stdin if omitted)")
	sessionsSaveCmd.Flags().String("id", "", "Session ID (triggers flag-based input)")
	sessionsSaveCmd.Flags().String("project", "", "Project name")
	sessionsSaveCmd.Flags().String("goal", "", "Session goal")
	sessionsSaveCmd.Flags().Int("patterns", 0, "Number of patterns extracted")
	sessionsSaveCmd.Flags().String("tool", "", "Tool used (e.g. claude, gemini)")
	sessionsSaveCmd.Flags().String("start-time", "", "Start time (RFC3339)")
	sessionsSaveCmd.Flags().String("end-time", "", "End time (RFC3339)")
	sessionsSaveCmd.Flags().String("url", "", "Workflow URL")
}
