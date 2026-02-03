package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "murmur-stats-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override home directory
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)

	// Create .murmur directory
	murmurDir := filepath.Join(tmpDir, ".murmur")
	if err := os.MkdirAll(murmurDir, 0755); err != nil {
		t.Fatalf("failed to create .murmur dir: %v", err)
	}

	cleanup := func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestRecord(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	record := UsageRecord{
		Tool:         "claude",
		Timestamp:    time.Now(),
		PromptLength: 100,
		DurationMs:   1500,
		CostEstimate: 0.0003,
		Tier:         "paid",
		RoutingMode:  "auto",
		AutoRouted:   true,
		Complexity:   0.7,
		Success:      true,
	}

	err := Record(record)
	if err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	// Verify file exists
	path, _ := StatsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("stats file was not created")
	}
}

func TestQuery(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Record multiple entries
	now := time.Now()
	records := []UsageRecord{
		{Tool: "claude", Timestamp: now.Add(-2 * time.Hour), PromptLength: 100, Tier: "paid", Success: true},
		{Tool: "gemini", Timestamp: now.Add(-1 * time.Hour), PromptLength: 50, Tier: "free", Success: true},
		{Tool: "claude", Timestamp: now, PromptLength: 200, Tier: "paid", Success: false},
	}

	for _, r := range records {
		if err := Record(r); err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	// Query all
	all, err := Query(QueryFilter{})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 records, got %d", len(all))
	}

	// Query by tool
	claudeOnly, err := Query(QueryFilter{Tool: "claude"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(claudeOnly) != 2 {
		t.Errorf("expected 2 claude records, got %d", len(claudeOnly))
	}

	// Query by tier
	freeOnly, err := Query(QueryFilter{Tier: "free"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(freeOnly) != 1 {
		t.Errorf("expected 1 free record, got %d", len(freeOnly))
	}
}

func TestSummarize(t *testing.T) {
	records := []UsageRecord{
		{Tool: "claude", Timestamp: time.Now(), PromptLength: 1000, DurationMs: 2000, CostEstimate: 0.003, Tier: "paid", AutoRouted: true, Success: true},
		{Tool: "gemini", Timestamp: time.Now(), PromptLength: 500, DurationMs: 1000, CostEstimate: 0, Tier: "free", AutoRouted: true, Success: true},
		{Tool: "gemini", Timestamp: time.Now(), PromptLength: 300, DurationMs: 800, CostEstimate: 0, Tier: "free", AutoRouted: true, Success: true},
	}

	summary := Summarize(records)

	if summary.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", summary.TotalRuns)
	}

	if summary.EstimatedCost != 0.003 {
		t.Errorf("expected cost 0.003, got %f", summary.EstimatedCost)
	}

	// Check tool breakdown
	if summary.ByTool["claude"].Count != 1 {
		t.Errorf("expected 1 claude run, got %d", summary.ByTool["claude"].Count)
	}
	if summary.ByTool["gemini"].Count != 2 {
		t.Errorf("expected 2 gemini runs, got %d", summary.ByTool["gemini"].Count)
	}

	// Check auto-route stats
	if summary.AutoRouteStats.Total != 3 {
		t.Errorf("expected 3 auto-routed, got %d", summary.AutoRouteStats.Total)
	}
	if summary.AutoRouteStats.ToFree != 2 {
		t.Errorf("expected 2 routed to free, got %d", summary.AutoRouteStats.ToFree)
	}
}

func TestReset(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a record
	record := UsageRecord{Tool: "test", Timestamp: time.Now()}
	if err := Record(record); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	// Verify file exists
	path, _ := StatsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("stats file should exist before reset")
	}

	// Reset
	if err := Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("stats file should not exist after reset")
	}
}

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		tool     string
		length   int
		expected float64
	}{
		{"claude", 1000, 0.003},
		{"claude", 500, 0.0015},
		{"gemini", 1000, 0.0},
		{"auggie", 1000, 0.0},
		{"unknown", 1000, 0.0},
	}

	for _, tt := range tests {
		cost := EstimateCost(tt.tool, tt.length)
		if cost != tt.expected {
			t.Errorf("EstimateCost(%s, %d) = %f, want %f", tt.tool, tt.length, cost, tt.expected)
		}
	}
}

func TestQueryEmptyFile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Query without any records
	records, err := Query(QueryFilter{})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestSummarizeEmpty(t *testing.T) {
	summary := Summarize([]UsageRecord{})

	if summary.TotalRuns != 0 {
		t.Errorf("expected 0 total runs, got %d", summary.TotalRuns)
	}
	if len(summary.ByTool) != 0 {
		t.Errorf("expected empty ByTool map, got %d entries", len(summary.ByTool))
	}
}
