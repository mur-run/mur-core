// Package learn provides cross-CLI learning capabilities.
package learn

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// OpenClawParser parses OpenClaw session transcript files.
type OpenClawParser struct{}

// openclawMessage represents a message in OpenClaw transcript format.
type openclawMessage struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	ParentID  string          `json:"parentId"`
	Timestamp string          `json:"timestamp"`
	Message   json.RawMessage `json:"message,omitempty"`
}

// openclawMessageContent represents the inner message content.
type openclawMessageContent struct {
	Role    string `json:"role"`
	Content []struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"`
	} `json:"content"`
}

// Parse reads an OpenClaw .jsonl transcript and extracts session entries.
func (p *OpenClawParser) Parse(path string) ([]SessionEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []SessionEntry
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines (OpenClaw messages can be huge)
	buf := make([]byte, 0, 1024*1024) // 1MB initial
	scanner.Buffer(buf, 10*1024*1024) // 10MB max

	for scanner.Scan() {
		var msg openclawMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		// Only process "message" type entries
		if msg.Type != "message" {
			continue
		}

		// Parse the inner message content
		var content openclawMessageContent
		if err := json.Unmarshal(msg.Message, &content); err != nil {
			continue
		}

		// Only process user and assistant messages
		if content.Role != "user" && content.Role != "assistant" {
			continue
		}

		// Extract text content
		var textContent string
		for _, block := range content.Content {
			switch block.Type {
			case "text":
				if textContent != "" {
					textContent += "\n"
				}
				textContent += block.Text
			case "thinking":
				// Include thinking for pattern extraction (valuable insights)
				if textContent != "" {
					textContent += "\n"
				}
				textContent += "[Thinking] " + block.Thinking
			}
		}

		if textContent == "" {
			continue
		}

		// Parse timestamp
		ts, _ := time.Parse(time.RFC3339, msg.Timestamp)

		entries = append(entries, SessionEntry{
			Role:      content.Role,
			Content:   textContent,
			Timestamp: ts,
		})
	}

	return entries, scanner.Err()
}

// OpenClawProjectsDir returns the path to ~/.openclaw/agents/main/sessions/
func OpenClawProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openclaw", "agents", "main", "sessions"), nil
}

// ListOpenClawSessions returns available sessions from OpenClaw.
func ListOpenClawSessions() ([]Session, error) {
	sessionsDir, err := OpenClawProjectsDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []Session{}, nil
	}

	var sessions []Session

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		// Skip lock files
		if filepath.Ext(filepath.Base(entry.Name())) == ".lock" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-6] // Remove .jsonl
		sessionPath := filepath.Join(sessionsDir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		sessions = append(sessions, Session{
			ID:        sessionID,
			Project:   "OpenClaw",
			Path:      sessionPath,
			CreatedAt: info.ModTime(),
		})
	}

	return sessions, nil
}
