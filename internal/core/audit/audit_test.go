package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempLogger(t *testing.T) *Logger {
	t.Helper()
	dir := t.TempDir()
	return NewLogger(dir)
}

func TestLogAndRead(t *testing.T) {
	logger := tempLogger(t)

	entry := Entry{
		PatternID:   "abc-123",
		PatternName: "go-errors",
		Action:      ActionInject,
		Source:      "cli",
		ToolTarget:  "claude",
		PromptHash:  "deadbeef",
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log() error: %v", err)
	}

	entries, err := logger.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.PatternID != "abc-123" {
		t.Errorf("PatternID = %q, want %q", got.PatternID, "abc-123")
	}
	if got.PatternName != "go-errors" {
		t.Errorf("PatternName = %q, want %q", got.PatternName, "go-errors")
	}
	if got.Action != ActionInject {
		t.Errorf("Action = %q, want %q", got.Action, ActionInject)
	}
	if got.ToolTarget != "claude" {
		t.Errorf("ToolTarget = %q, want %q", got.ToolTarget, "claude")
	}
	if got.Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}
}

func TestLogMultipleEntries(t *testing.T) {
	logger := tempLogger(t)

	for i := 0; i < 5; i++ {
		err := logger.Log(Entry{
			PatternID:   "id-" + string(rune('a'+i)),
			PatternName: "pattern-" + string(rune('a'+i)),
			Action:      ActionInject,
			Source:      "cli",
		})
		if err != nil {
			t.Fatalf("Log() error on entry %d: %v", i, err)
		}
	}

	entries, err := logger.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	// Most recent first
	if entries[0].PatternName != "pattern-e" {
		t.Errorf("first entry should be most recent, got %q", entries[0].PatternName)
	}
}

func TestReadFilteredByPattern(t *testing.T) {
	logger := tempLogger(t)

	_ = logger.Log(Entry{PatternName: "go-errors", Action: ActionInject})
	_ = logger.Log(Entry{PatternName: "swift-async", Action: ActionInject})
	_ = logger.Log(Entry{PatternName: "go-errors", Action: ActionLoad})
	_ = logger.Log(Entry{PatternName: "python-typing", Action: ActionInject})

	entries, err := logger.ReadFiltered("go-errors")
	if err != nil {
		t.Fatalf("ReadFiltered() error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(entries))
	}

	for _, e := range entries {
		if e.PatternName != "go-errors" {
			t.Errorf("unexpected pattern name: %q", e.PatternName)
		}
	}
}

func TestReadEmptyLog(t *testing.T) {
	logger := tempLogger(t)

	entries, err := logger.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if entries != nil {
		t.Errorf("expected nil entries for empty log, got %d", len(entries))
	}
}

func TestLogSetsTimestamp(t *testing.T) {
	logger := tempLogger(t)

	before := time.Now()
	_ = logger.Log(Entry{PatternName: "test", Action: ActionLoad})
	after := time.Now()

	entries, _ := logger.Read()
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}

	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestLogPreservesExplicitTimestamp(t *testing.T) {
	logger := tempLogger(t)

	explicit := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	_ = logger.Log(Entry{Timestamp: explicit, PatternName: "test", Action: ActionLoad})

	entries, _ := logger.Read()
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}

	if !entries[0].Timestamp.Equal(explicit) {
		t.Errorf("timestamp = %v, want %v", entries[0].Timestamp, explicit)
	}
}

func TestRotate(t *testing.T) {
	logger := tempLogger(t)

	_ = logger.Log(Entry{PatternName: "test", Action: ActionInject})

	// Verify file exists
	if _, err := os.Stat(logger.logFile()); err != nil {
		t.Fatal("audit.jsonl should exist before rotation")
	}

	if err := logger.Rotate(); err != nil {
		t.Fatalf("Rotate() error: %v", err)
	}

	// Original file should be gone
	if _, err := os.Stat(logger.logFile()); !os.IsNotExist(err) {
		t.Error("audit.jsonl should not exist after rotation")
	}

	// Archived file should exist
	archivePattern := filepath.Join(logger.dir, "audit-*.jsonl")
	matches, _ := filepath.Glob(archivePattern)
	if len(matches) != 1 {
		t.Errorf("expected 1 archive file, got %d", len(matches))
	}
}

func TestRotateNoFile(t *testing.T) {
	logger := tempLogger(t)

	// Rotating with no file should not error
	if err := logger.Rotate(); err != nil {
		t.Fatalf("Rotate() on empty dir should not error: %v", err)
	}
}

func TestDefaultLogger(t *testing.T) {
	logger, err := DefaultLogger()
	if err != nil {
		t.Fatalf("DefaultLogger() error: %v", err)
	}
	if logger == nil {
		t.Fatal("DefaultLogger() returned nil")
	}
	if logger.dir == "" {
		t.Error("DefaultLogger() dir should not be empty")
	}
}

func TestAllActions(t *testing.T) {
	logger := tempLogger(t)

	actions := []Action{ActionInject, ActionLoad, ActionShare, ActionModify, ActionVerify}
	for _, a := range actions {
		_ = logger.Log(Entry{PatternName: "test", Action: a})
	}

	entries, err := logger.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(entries) != len(actions) {
		t.Errorf("expected %d entries, got %d", len(actions), len(entries))
	}
}
