package learn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Session represents a Claude Code session.
type Session struct {
	ID           string
	Project      string
	Path         string
	Messages     []SessionMessage
	ToolUseCount int // Number of tool_use blocks in the session
	CreatedAt    time.Time
}

// SessionMessage represents a message in a session.
type SessionMessage struct {
	Type      string // "user", "assistant", "progress", etc.
	Role      string // "user", "assistant"
	Content   string // Text content
	Timestamp time.Time
}

// jsonlMessage represents the raw JSONL message structure from Claude Code.
type jsonlMessage struct {
	Type      string          `json:"type"`
	Message   json.RawMessage `json:"message,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
}

// messageContent represents the message field structure.
type messageContent struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock represents a content block (text or tool_use).
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ClaudeProjectsDir returns the path to ~/.claude/projects/
func ClaudeProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// ListSessions returns available sessions from Claude Code and OpenClaw.
func ListSessions() ([]Session, error) {
	var sessions []Session

	// Get Claude Code sessions
	claudeSessions, err := listClaudeSessions()
	if err == nil {
		sessions = append(sessions, claudeSessions...)
	}

	// Get OpenClaw sessions
	openclawSessions, err := ListOpenClawSessions()
	if err == nil {
		sessions = append(sessions, openclawSessions...)
	}

	// Sort by creation time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	return sessions, nil
}

// listClaudeSessions returns sessions from ~/.claude/projects/
func listClaudeSessions() ([]Session, error) {
	projectsDir, err := ClaudeProjectsDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return []Session{}, nil
	}

	var sessions []Session

	// Walk through project directories
	projectDirs, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read projects directory: %w", err)
	}

	for _, projectDir := range projectDirs {
		if !projectDir.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, projectDir.Name())
		entries, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
				continue
			}

			sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
			sessionPath := filepath.Join(projectPath, entry.Name())

			info, err := entry.Info()
			if err != nil {
				continue
			}

			sessions = append(sessions, Session{
				ID:        sessionID,
				Project:   projectDir.Name(),
				Path:      sessionPath,
				CreatedAt: info.ModTime(),
			})
		}
	}

	return sessions, nil
}

// LoadSession loads a session by ID or path.
func LoadSession(idOrPath string) (*Session, error) {
	var sessionPath string
	var sessionID string
	var project string

	// Check if it's a direct path
	if strings.HasSuffix(idOrPath, ".jsonl") {
		sessionPath = idOrPath
		sessionID = strings.TrimSuffix(filepath.Base(idOrPath), ".jsonl")
		project = filepath.Base(filepath.Dir(idOrPath))
	} else {
		// Search by ID
		sessions, err := ListSessions()
		if err != nil {
			return nil, err
		}

		found := false
		for _, s := range sessions {
			if s.ID == idOrPath || strings.HasPrefix(s.ID, idOrPath) {
				sessionPath = s.Path
				sessionID = s.ID
				project = s.Project
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("session not found: %s", idOrPath)
		}
	}

	// Parse the JSONL file
	messages, toolUseCount, err := parseJSONL(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	info, err := os.Stat(sessionPath)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:           sessionID,
		Project:      project,
		Path:         sessionPath,
		Messages:     messages,
		ToolUseCount: toolUseCount,
		CreatedAt:    info.ModTime(),
	}, nil
}

// RecentSessions returns sessions from the last N days.
func RecentSessions(days int) ([]Session, error) {
	sessions, err := ListSessions()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var recent []Session

	for _, s := range sessions {
		if s.CreatedAt.After(cutoff) {
			recent = append(recent, s)
		}
	}

	return recent, nil
}

// parseJSONL parses a Claude Code session JSONL file.
// Returns messages and tool use count.
func parseJSONL(path string) ([]SessionMessage, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = file.Close() }()

	var messages []SessionMessage
	toolUseCount := 0
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines (OpenClaw can have huge messages)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max per line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Count tool_use entries at the JSONL level
		if strings.Contains(string(line), `"type":"tool_use"`) {
			toolUseCount++
		}

		var msg jsonlMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue // Skip malformed lines
		}

		var role string
		var text string
		var timestamp time.Time

		// Handle OpenClaw format: type="message" with nested message.role
		if msg.Type == "message" && msg.Message != nil {
			var content messageContent
			if err := json.Unmarshal(msg.Message, &content); err != nil {
				continue
			}

			// Only process user and assistant messages
			if content.Role != "user" && content.Role != "assistant" {
				continue
			}

			role = content.Role
			text = extractText(content.Content)
			timestamp, _ = time.Parse(time.RFC3339, msg.Timestamp)

		} else if msg.Type == "user" || msg.Type == "assistant" {
			// Handle Claude Code format: type="user" or type="assistant"
			role = msg.Type
			timestamp, _ = time.Parse(time.RFC3339, msg.Timestamp)

			if msg.Message == nil {
				continue
			}

			var content messageContent
			if err := json.Unmarshal(msg.Message, &content); err != nil {
				continue
			}

			text = extractText(content.Content)
		} else {
			continue
		}

		if text == "" {
			continue
		}

		// Skip meta messages and command outputs
		if strings.Contains(text, "<local-command-") ||
			strings.Contains(text, "<command-name>") ||
			strings.Contains(text, "<local-command-stdout>") {
			continue
		}

		messages = append(messages, SessionMessage{
			Type:      msg.Type,
			Role:      role,
			Content:   text,
			Timestamp: timestamp,
		})
	}

	return messages, toolUseCount, scanner.Err()
}

// contentBlockExt extends contentBlock with thinking support.
type contentBlockExt struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// extractText extracts text from message content.
func extractText(raw json.RawMessage) string {
	// Try as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try as array of content blocks (with thinking support)
	var blocks []contentBlockExt
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var texts []string
		for _, block := range blocks {
			switch block.Type {
			case "text":
				if block.Text != "" {
					texts = append(texts, block.Text)
				}
			case "thinking":
				// Include thinking for pattern extraction
				if block.Thinking != "" {
					texts = append(texts, "[Thinking] "+block.Thinking)
				}
			}
		}
		return strings.Join(texts, "\n")
	}

	return ""
}

// ShortID returns a shortened session ID for display.
func (s *Session) ShortID() string {
	if len(s.ID) > 8 {
		return s.ID[:8]
	}
	return s.ID
}

// AssistantMessages returns only assistant messages.
func (s *Session) AssistantMessages() []SessionMessage {
	var msgs []SessionMessage
	for _, m := range s.Messages {
		if m.Role == "assistant" {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

// UserMessages returns only user messages.
func (s *Session) UserMessages() []SessionMessage {
	var msgs []SessionMessage
	for _, m := range s.Messages {
		if m.Role == "user" {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

// FullTranscript returns the full conversation as a single string.
func (s *Session) FullTranscript() string {
	var sb strings.Builder
	for _, m := range s.Messages {
		sb.WriteString(m.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}
