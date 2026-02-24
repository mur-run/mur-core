package session

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

// EventRecord represents a single recorded event in a session.
type EventRecord struct {
	Timestamp int64          `json:"ts"`
	Type      string         `json:"type"` // "user", "assistant", "tool_call", "tool_result"
	Content   string         `json:"content"`
	Tool      string         `json:"tool,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// RecordingInfo summarizes a past recording for listing.
type RecordingInfo struct {
	SessionID  string
	StartedAt  time.Time
	EventCount int
	FileSize   int64
	Source     string
}

// RecordEvent appends an EventRecord to the session's JSONL file.
func RecordEvent(sessionID string, event EventRecord) error {
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	recDir, err := recordingsDir()
	if err != nil {
		return err
	}

	jsonlPath := filepath.Join(recDir, sessionID+".jsonl")

	f, err := os.OpenFile(jsonlPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open recording file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cannot marshal event: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("cannot write event: %w", err)
	}

	return nil
}

// RecordEventForActive records an event to the currently active session.
// Returns nil if no session is active (no-op).
func RecordEventForActive(event EventRecord) error {
	active, sessionID := IsRecording()
	if !active {
		return nil
	}
	return RecordEvent(sessionID, event)
}

// ReadEvents reads all events from a session JSONL file.
func ReadEvents(sessionID string) ([]EventRecord, error) {
	// Resolve short prefix to full UUID
	resolved, err := ResolveSessionID(sessionID)
	if err == nil {
		sessionID = resolved
	}

	recDir, err := recordingsDir()
	if err != nil {
		return nil, err
	}

	jsonlPath := filepath.Join(recDir, sessionID+".jsonl")
	f, err := os.Open(jsonlPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open recording: %w", err)
	}
	defer f.Close()

	var events []EventRecord
	scanner := bufio.NewScanner(f)
	// Allow large lines (1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event EventRecord
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // skip malformed lines
		}
		events = append(events, event)
	}

	return events, scanner.Err()
}

// ListRecordings returns info about all past recording sessions.
func ListRecordings() ([]RecordingInfo, error) {
	recDir, err := recordingsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(recDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot read recordings directory: %w", err)
	}

	var recordings []RecordingInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Count events and get first timestamp
		events, _ := ReadEvents(sessionID)
		var startedAt time.Time
		var source string
		if len(events) > 0 {
			startedAt = time.Unix(events[0].Timestamp, 0)
		} else {
			startedAt = info.ModTime()
		}

		// Try to read source from the active.json metadata stored in the first event
		if len(events) > 0 && events[0].Meta != nil {
			if s, ok := events[0].Meta["source"].(string); ok {
				source = s
			}
		}

		recordings = append(recordings, RecordingInfo{
			SessionID:  sessionID,
			StartedAt:  startedAt,
			EventCount: len(events),
			FileSize:   info.Size(),
			Source:     source,
		})
	}

	// Sort by start time, newest first
	sort.Slice(recordings, func(i, j int) bool {
		return recordings[i].StartedAt.After(recordings[j].StartedAt)
	})

	return recordings, nil
}

// RecordingPath returns the full path to a session's JSONL file.
func RecordingPath(sessionID string) (string, error) {
	recDir, err := recordingsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(recDir, sessionID+".jsonl"), nil
}

// ResolveSessionID resolves a session ID prefix to a full session ID.
// Accepts both full UUIDs and short prefixes (e.g. first 8 chars).
func ResolveSessionID(prefix string) (string, error) {
	recDir, err := recordingsDir()
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(recDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no recordings found")
		}
		return "", fmt.Errorf("read recordings: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		sid := strings.TrimSuffix(entry.Name(), ".jsonl")
		if strings.HasPrefix(sid, prefix) {
			matches = append(matches, sid)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session found matching %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous prefix %q matches %d sessions", prefix, len(matches))
	}
}
