package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/session"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage conversation recording sessions",
	Long: `Record conversation segments for workflow extraction.

Use /mur:in and /mur:out in Claude Code to mark segments, or use
these commands directly:

  mur session start          # Start recording
  mur session stop           # Stop recording
  mur session status         # Check recording state
  mur session list           # List past recordings
  mur session record         # Append an event to the active session`,
}

var sessionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start recording a conversation segment",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		marker, _ := cmd.Flags().GetString("marker")

		if source == "" {
			source = "cli"
		}

		state, err := session.StartRecording(source, marker)
		if err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Recording started (session: %s)\n", state.SessionID[:8])
		return nil
	},
}

var sessionStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop recording and optionally analyze",
	RunE: func(cmd *cobra.Command, args []string) error {
		analyze, _ := cmd.Flags().GetBool("analyze")

		state, err := session.StopRecording()
		if err != nil {
			return err
		}

		events, _ := session.ReadEvents(state.SessionID)
		duration := time.Since(time.Unix(state.StartedAt, 0)).Truncate(time.Second)

		fmt.Fprintf(os.Stderr, "Recording stopped (session: %s)\n", state.SessionID[:8])
		fmt.Fprintf(os.Stderr, "  Duration: %s\n", duration)
		fmt.Fprintf(os.Stderr, "  Events:   %d\n", len(events))

		if analyze {
			fmt.Fprintf(os.Stderr, "\nAnalysis not yet implemented (Phase 3).\n")
			fmt.Fprintf(os.Stderr, "Recording saved: ~/.mur/session/recordings/%s.jsonl\n", state.SessionID)
		}

		return nil
	},
}

var sessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current recording status",
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")

		state, err := session.GetState()
		if err != nil {
			if quiet {
				os.Exit(1)
			}
			return fmt.Errorf("cannot read state: %w", err)
		}

		if state == nil || !state.Active {
			if quiet {
				os.Exit(1)
			}
			fmt.Println("No active recording session.")
			return nil
		}

		if quiet {
			// Exit 0 = recording active (for use in hook scripts)
			os.Exit(0)
		}

		duration := time.Since(time.Unix(state.StartedAt, 0)).Truncate(time.Second)
		events, _ := session.ReadEvents(state.SessionID)

		fmt.Printf("Recording active\n")
		fmt.Printf("  Session: %s\n", state.SessionID[:8])
		fmt.Printf("  Source:  %s\n", state.Source)
		fmt.Printf("  Started: %s ago\n", duration)
		fmt.Printf("  Events:  %d\n", len(events))

		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List past recordings",
	RunE: func(cmd *cobra.Command, args []string) error {
		recordings, err := session.ListRecordings()
		if err != nil {
			return fmt.Errorf("failed to list recordings: %w", err)
		}

		if len(recordings) == 0 {
			fmt.Println("No recordings found.")
			return nil
		}

		fmt.Println("Session Recordings")
		fmt.Println("==================")
		fmt.Println()

		for _, r := range recordings {
			source := r.Source
			if source == "" {
				source = "unknown"
			}
			fmt.Printf("  %s  %s  %d events  (%s)\n",
				r.SessionID[:8],
				r.StartedAt.Format("2006-01-02 15:04"),
				r.EventCount,
				source,
			)
		}

		fmt.Printf("\nTotal: %d recordings\n", len(recordings))
		return nil
	},
}

var sessionRecordCmd = &cobra.Command{
	Use:    "record",
	Short:  "Append an event to the active session",
	Hidden: true, // Used by hook scripts, not end users
	RunE: func(cmd *cobra.Command, args []string) error {
		eventType, _ := cmd.Flags().GetString("type")
		content, _ := cmd.Flags().GetString("content")
		tool, _ := cmd.Flags().GetString("tool")

		if eventType == "" {
			return fmt.Errorf("--type is required")
		}

		event := session.EventRecord{
			Type:    eventType,
			Content: content,
			Tool:    tool,
		}

		return session.RecordEventForActive(event)
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionStartCmd)
	sessionCmd.AddCommand(sessionStopCmd)
	sessionCmd.AddCommand(sessionStatusCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionRecordCmd)

	sessionStartCmd.Flags().String("source", "", "Recording source (e.g. claude-code, codex)")
	sessionStartCmd.Flags().String("marker", "", "Context marker from /mur:in message")

	sessionStopCmd.Flags().Bool("analyze", false, "Analyze the recording after stopping (Phase 3)")
	sessionStopCmd.Flags().Bool("open", false, "Open web UI after analysis (Phase 4)")

	sessionStatusCmd.Flags().BoolP("quiet", "q", false, "Exit 0 if recording, 1 if not (for scripts)")

	sessionRecordCmd.Flags().String("type", "", "Event type: user, assistant, tool_call, tool_result")
	sessionRecordCmd.Flags().String("content", "", "Event content")
	sessionRecordCmd.Flags().String("tool", "", "Tool name (for tool_call/tool_result events)")
}
