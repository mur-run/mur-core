package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var transcriptsCmd = &cobra.Command{
	Use:   "transcripts",
	Short: "Browse Claude Code session transcripts",
	Long: `List and view Claude Code session transcripts.

Transcripts are stored in ~/.claude/projects/ and contain
conversation history that can be used for pattern extraction.

Examples:
  mur transcripts                    # List recent sessions
  mur transcripts --all              # List all sessions
  mur transcripts show <session>     # View a session
  mur transcripts --project BitL     # Filter by project name`,
	RunE: runTranscripts,
}

var transcriptsShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show transcript content",
	Args:  cobra.ExactArgs(1),
	RunE:  runTranscriptsShow,
}

var (
	transcriptsAll     bool
	transcriptsProject string
	transcriptsLimit   int
)

func init() {
	rootCmd.AddCommand(transcriptsCmd)
	transcriptsCmd.AddCommand(transcriptsShowCmd)

	transcriptsCmd.Flags().BoolVar(&transcriptsAll, "all", false, "Show all sessions")
	transcriptsCmd.Flags().StringVar(&transcriptsProject, "project", "", "Filter by project name")
	transcriptsCmd.Flags().IntVarP(&transcriptsLimit, "limit", "n", 20, "Max sessions to show")
}

// Session represents a Claude Code session
type Session struct {
	ID           string
	Project      string
	Path         string
	LastModified time.Time
	MessageCount int
	Size         int64
}

// TranscriptMessage represents a message in a transcript
type TranscriptMessage struct {
	Type      string `json:"type"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func runTranscripts(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	projectsDir := filepath.Join(home, ".claude", "projects")
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		fmt.Println("No Claude Code sessions found.")
		fmt.Println("Sessions are stored in ~/.claude/projects/")
		return nil
	}

	sessions, err := findSessions(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to scan sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	// Filter by project
	if transcriptsProject != "" {
		var filtered []Session
		for _, s := range sessions {
			if strings.Contains(strings.ToLower(s.Project), strings.ToLower(transcriptsProject)) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Sort by last modified (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastModified.After(sessions[j].LastModified)
	})

	// Limit results
	if !transcriptsAll && len(sessions) > transcriptsLimit {
		sessions = sessions[:transcriptsLimit]
	}

	// Display
	fmt.Println()
	fmt.Println("📜 Claude Code Sessions")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-20s %-30s %6s %s\n", "SESSION", "PROJECT", "MSGS", "LAST MODIFIED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range sessions {
		project := s.Project
		if len(project) > 28 {
			project = project[:25] + "..."
		}

		age := formatAge(s.LastModified)
		fmt.Printf("%-20s %-30s %6d %s\n", s.ID[:min(20, len(s.ID))], project, s.MessageCount, age)
	}

	fmt.Println()
	fmt.Printf("Showing %d sessions", len(sessions))
	if !transcriptsAll {
		fmt.Printf(" (use --all for more)")
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("View session:   mur transcripts show <session-id>")
	fmt.Println("Extract:        mur learn extract --session <session-id>")
	fmt.Println()

	return nil
}

func runTranscriptsShow(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	projectsDir := filepath.Join(home, ".claude", "projects")
	sessions, err := findSessions(projectsDir)
	if err != nil {
		return err
	}

	// Find matching session
	var found *Session
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, sessionID) || s.ID == sessionID {
			found = &s
			break
		}
	}

	if found == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Read and display transcript
	content, err := os.ReadFile(found.Path)
	if err != nil {
		return fmt.Errorf("failed to read transcript: %w", err)
	}

	var messages []TranscriptMessage
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg TranscriptMessage
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			messages = append(messages, msg)
		}
	}

	fmt.Println()
	fmt.Printf("📜 Session: %s\n", found.ID)
	fmt.Printf("📁 Project: %s\n", found.Project)
	fmt.Printf("📊 Messages: %d\n", len(messages))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for i, msg := range messages {
		if msg.Type == "user" || msg.Role == "user" {
			fmt.Printf("👤 [%d] User:\n", i+1)
			printWrapped(msg.Content, "   ")
			fmt.Println()
		} else if msg.Type == "assistant" || msg.Role == "assistant" {
			// Truncate long assistant messages
			content := msg.Content
			if len(content) > 500 {
				content = content[:500] + "... [truncated]"
			}
			fmt.Printf("🤖 [%d] Assistant:\n", i+1)
			printWrapped(content, "   ")
			fmt.Println()
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Extract patterns: mur learn extract --session", found.ID)
	fmt.Println()

	return nil
}

func findSessions(projectsDir string) ([]Session, error) {
	var sessions []Session

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())
		transcriptFiles, _ := filepath.Glob(filepath.Join(projectPath, "*.jsonl"))

		for _, tf := range transcriptFiles {
			info, err := os.Stat(tf)
			if err != nil {
				continue
			}

			// Count messages
			content, _ := os.ReadFile(tf)
			lines := strings.Split(string(content), "\n")
			msgCount := 0
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					msgCount++
				}
			}

			// Parse project name from directory
			projectName := entry.Name()
			projectName = strings.TrimPrefix(projectName, "-")
			projectName = strings.ReplaceAll(projectName, "-", "/")

			// Session ID is the filename without extension
			sessionID := strings.TrimSuffix(filepath.Base(tf), ".jsonl")

			sessions = append(sessions, Session{
				ID:           sessionID,
				Project:      projectName,
				Path:         tf,
				LastModified: info.ModTime(),
				MessageCount: msgCount,
				Size:         info.Size(),
			})
		}
	}

	return sessions, nil
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

func printWrapped(s string, prefix string) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		// Simple wrap at 80 chars
		for len(line) > 76 {
			fmt.Println(prefix + line[:76])
			line = line[76:]
		}
		fmt.Println(prefix + line)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
