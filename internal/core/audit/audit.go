// Package audit provides append-only audit logging for pattern operations.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Action represents the type of audit event.
type Action string

const (
	ActionInject Action = "inject"
	ActionLoad   Action = "load"
	ActionShare  Action = "share"
	ActionModify Action = "modify"
	ActionVerify Action = "verify"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp   time.Time `json:"timestamp"`
	PatternID   string    `json:"pattern_id"`
	PatternName string    `json:"pattern_name"`
	Action      Action    `json:"action"`
	Source      string    `json:"source,omitempty"`
	ToolTarget  string    `json:"tool_target,omitempty"`
	PromptHash  string    `json:"prompt_hash,omitempty"`
	Details     string    `json:"details,omitempty"`
}

// defaultMaxSizeBytes is the default max audit log size before auto-rotation (10 MB).
const defaultMaxSizeBytes int64 = 10 * 1024 * 1024

// Logger provides append-only audit logging.
type Logger struct {
	dir          string
	maxSizeBytes int64
	mu           sync.Mutex
}

// NewLogger creates a new audit logger writing to the given directory.
func NewLogger(dir string) *Logger {
	return &Logger{dir: dir, maxSizeBytes: defaultMaxSizeBytes}
}

// SetMaxSize sets the maximum audit log file size in bytes before auto-rotation.
func (l *Logger) SetMaxSize(bytes int64) {
	l.maxSizeBytes = bytes
}

// DefaultLogger returns an audit logger using ~/.mur/audit/.
func DefaultLogger() (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	return NewLogger(filepath.Join(home, ".mur", "audit")), nil
}

// logFile returns the path to the current audit log file.
func (l *Logger) logFile() string {
	return filepath.Join(l.dir, "audit.jsonl")
}

// Log appends an entry to the audit log. If the log file exceeds maxSizeBytes,
// it is automatically rotated before writing.
func (l *Logger) Log(entry Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(l.dir, 0755); err != nil {
		return fmt.Errorf("cannot create audit directory: %w", err)
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Auto-rotate if file exceeds max size
	if l.maxSizeBytes > 0 {
		if info, err := os.Stat(l.logFile()); err == nil && info.Size() >= l.maxSizeBytes {
			if err := l.rotateLocked(); err != nil {
				return fmt.Errorf("auto-rotation failed: %w", err)
			}
		}
	}

	f, err := os.OpenFile(l.logFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open audit log: %w", err)
	}
	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(entry)
}

// Read returns all audit entries, most recent first.
func (l *Logger) Read() ([]Entry, error) {
	return l.ReadFiltered("")
}

// ReadFiltered returns audit entries, optionally filtered by pattern name.
func (l *Logger) ReadFiltered(patternName string) ([]Entry, error) {
	path := l.logFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read audit log: %w", err)
	}

	var entries []Entry
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	for decoder.More() {
		var e Entry
		if err := decoder.Decode(&e); err != nil {
			continue // skip malformed entries
		}
		if patternName != "" && e.PatternName != patternName {
			continue
		}
		entries = append(entries, e)
	}

	// Reverse for most-recent-first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

// Rotate moves the current log to a dated archive file.
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotateLocked()
}

// rotateLocked performs rotation without acquiring the mutex (caller must hold it).
func (l *Logger) rotateLocked() error {
	src := l.logFile()
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // nothing to rotate
	}

	dst := filepath.Join(l.dir, fmt.Sprintf("audit-%s.jsonl", time.Now().Format("2006-01")))
	return os.Rename(src, dst)
}
