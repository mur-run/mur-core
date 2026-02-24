// Package sessions provides session history tracking for mur.
package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SessionRecord represents a completed mur session in history.
type SessionRecord struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Project   string    `json:"project"`
	Goal      string    `json:"goal"`
	Patterns  int       `json:"patterns_extracted"`
	URL       string    `json:"workflow_url,omitempty"`
	Tool      string    `json:"tool"` // "openclaw", "claude", etc.
}

// historyPath returns the path to ~/.mur/sessions/history.json.
func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "sessions", "history.json"), nil
}

// loadHistory reads the history file and returns all records.
func loadHistory() ([]SessionRecord, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}

	var records []SessionRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return records, nil
}

// saveHistory writes the records to the history file.
func saveHistory(records []SessionRecord) error {
	path, err := historyPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create sessions directory: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// SaveSession saves a session record to ~/.mur/sessions/history.json.
func SaveSession(record SessionRecord) error {
	records, err := loadHistory()
	if err != nil {
		return err
	}

	records = append(records, record)
	return saveHistory(records)
}

// ListSessions returns all saved session records, newest first.
func ListSessions() ([]SessionRecord, error) {
	records, err := loadHistory()
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StartTime.After(records[j].StartTime)
	})

	return records, nil
}

// GetSession returns a specific session by ID.
func GetSession(id string) (*SessionRecord, error) {
	records, err := loadHistory()
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("session %s not found", id)
}
