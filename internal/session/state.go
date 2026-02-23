// Package session manages conversation recording sessions for workflow extraction.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// RecordingState represents the current recording state persisted as active.json.
type RecordingState struct {
	Active    bool   `json:"active"`
	SessionID string `json:"session_id"`
	StartedAt int64  `json:"started_at"`
	Source    string `json:"source"` // "claude-code", "codex", etc.
	Marker    string `json:"marker"` // original /mur:in message context
}

// sessionDir returns the path to ~/.mur/session/.
func sessionDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "session"), nil
}

// recordingsDir returns the path to ~/.mur/session/recordings/.
func recordingsDir() (string, error) {
	dir, err := sessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "recordings"), nil
}

// activeStatePath returns the path to active.json.
func activeStatePath() (string, error) {
	dir, err := sessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "active.json"), nil
}

// StartRecording creates a new recording session.
// If a session is already active, it stops the old one first.
func StartRecording(source, marker string) (*RecordingState, error) {
	// If already recording, stop the current session
	if state, _ := GetState(); state != nil && state.Active {
		if _, err := StopRecording(); err != nil {
			return nil, fmt.Errorf("failed to stop existing session: %w", err)
		}
	}

	// Ensure directories exist
	recDir, err := recordingsDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(recDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create recordings directory: %w", err)
	}

	state := &RecordingState{
		Active:    true,
		SessionID: uuid.New().String(),
		StartedAt: time.Now().Unix(),
		Source:    source,
		Marker:    marker,
	}

	// Create the JSONL file (empty, ready for appends)
	jsonlPath := filepath.Join(recDir, state.SessionID+".jsonl")
	f, err := os.Create(jsonlPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create recording file: %w", err)
	}
	f.Close()

	// Write active.json
	if err := writeState(state); err != nil {
		// Cleanup the JSONL file on failure
		os.Remove(jsonlPath)
		return nil, err
	}

	return state, nil
}

// StopRecording stops the active recording session and removes active.json.
// Returns the final state (with Active=false) or an error.
func StopRecording() (*RecordingState, error) {
	state, err := GetState()
	if err != nil {
		return nil, fmt.Errorf("cannot read recording state: %w", err)
	}
	if state == nil || !state.Active {
		return nil, fmt.Errorf("no active recording session")
	}

	state.Active = false

	// Remove active.json
	path, err := activeStatePath()
	if err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("cannot remove active state: %w", err)
	}

	return state, nil
}

// GetState reads the current recording state from active.json.
// Returns nil, nil if no active.json exists.
func GetState() (*RecordingState, error) {
	path, err := activeStatePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot read active state: %w", err)
	}

	var state RecordingState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("cannot parse active state: %w", err)
	}

	return &state, nil
}

// IsRecording returns true if there's an active recording, plus the session ID.
func IsRecording() (bool, string) {
	state, err := GetState()
	if err != nil || state == nil {
		return false, ""
	}
	return state.Active, state.SessionID
}

// writeState writes a RecordingState to active.json.
func writeState(state *RecordingState) error {
	path, err := activeStatePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal state: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
