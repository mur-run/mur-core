// Package analytics provides pattern usage tracking for mur.
package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// EventType represents the type of pattern usage event.
type EventType string

const (
	EventSearch    EventType = "search"    // Pattern found via semantic search
	EventInject    EventType = "inject"    // Pattern injected via sync/hooks
	EventView      EventType = "view"      // Pattern viewed in dashboard/CLI
	EventFeedback  EventType = "feedback"  // User feedback on pattern
)

// Event represents a single pattern usage event.
type Event struct {
	PatternID   string    `json:"pattern_id"`
	PatternName string    `json:"pattern_name"`
	EventType   EventType `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	Score       float64   `json:"score,omitempty"`      // For search events
	Source      string    `json:"source,omitempty"`     // hook, cli, dashboard
	Helpful     *bool     `json:"helpful,omitempty"`    // For feedback events
	Context     string    `json:"context,omitempty"`    // Query or prompt snippet
}

// PatternStats aggregates usage stats for a pattern.
type PatternStats struct {
	PatternID     string    `json:"pattern_id"`
	PatternName   string    `json:"pattern_name"`
	SearchCount   int       `json:"search_count"`
	InjectCount   int       `json:"inject_count"`
	ViewCount     int       `json:"view_count"`
	TotalHits     int       `json:"total_hits"`
	AvgScore      float64   `json:"avg_score"`
	LastUsed      time.Time `json:"last_used"`
	Effectiveness float64   `json:"effectiveness"` // 0-1 based on feedback
}

// Tracker manages pattern usage analytics.
type Tracker struct {
	dir    string
	mu     sync.Mutex
	events []Event
}

// NewTracker creates a new analytics tracker.
func NewTracker(murDir string) *Tracker {
	return &Tracker{
		dir:    filepath.Join(murDir, "analytics"),
		events: make([]Event, 0),
	}
}

// eventsFile returns the path to the events file.
func (t *Tracker) eventsFile() string {
	return filepath.Join(t.dir, "events.jsonl")
}

// Record records a pattern usage event.
func (t *Tracker) Record(event Event) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Ensure directory exists
	if err := os.MkdirAll(t.dir, 0755); err != nil {
		return fmt.Errorf("create analytics dir: %w", err)
	}

	// Append to file
	f, err := os.OpenFile(t.eventsFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open events file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	return nil
}

// RecordSearch records a search hit event.
func (t *Tracker) RecordSearch(patternID, patternName string, score float64, query string) error {
	return t.Record(Event{
		PatternID:   patternID,
		PatternName: patternName,
		EventType:   EventSearch,
		Score:       score,
		Source:      "search",
		Context:     truncate(query, 100),
	})
}

// RecordInject records a pattern injection event.
func (t *Tracker) RecordInject(patternID, patternName, source string) error {
	return t.Record(Event{
		PatternID:   patternID,
		PatternName: patternName,
		EventType:   EventInject,
		Source:      source,
	})
}

// RecordFeedback records user feedback on a pattern.
func (t *Tracker) RecordFeedback(patternID, patternName string, helpful bool) error {
	return t.Record(Event{
		PatternID:   patternID,
		PatternName: patternName,
		EventType:   EventFeedback,
		Helpful:     &helpful,
	})
}

// LoadEvents loads all events from disk.
func (t *Tracker) LoadEvents() ([]Event, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := os.ReadFile(t.eventsFile())
	if os.IsNotExist(err) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read events file: %w", err)
	}

	var events []Event
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
	}

	return events, nil
}

// GetPatternStats returns aggregated stats for all patterns.
func (t *Tracker) GetPatternStats() ([]PatternStats, error) {
	events, err := t.LoadEvents()
	if err != nil {
		return nil, err
	}

	// Aggregate by pattern
	statsMap := make(map[string]*PatternStats)

	for _, e := range events {
		key := e.PatternID
		if key == "" {
			key = e.PatternName
		}

		stats, ok := statsMap[key]
		if !ok {
			stats = &PatternStats{
				PatternID:   e.PatternID,
				PatternName: e.PatternName,
			}
			statsMap[key] = stats
		}

		switch e.EventType {
		case EventSearch:
			stats.SearchCount++
			stats.AvgScore = (stats.AvgScore*float64(stats.SearchCount-1) + e.Score) / float64(stats.SearchCount)
		case EventInject:
			stats.InjectCount++
		case EventView:
			stats.ViewCount++
		case EventFeedback:
			if e.Helpful != nil {
				// Update effectiveness based on feedback
				// Simple moving average
				if *e.Helpful {
					stats.Effectiveness = stats.Effectiveness*0.9 + 0.1
				} else {
					stats.Effectiveness = stats.Effectiveness * 0.9
				}
			}
		}

		stats.TotalHits = stats.SearchCount + stats.InjectCount
		if e.Timestamp.After(stats.LastUsed) {
			stats.LastUsed = e.Timestamp
		}
	}

	// Convert to slice and sort by total hits
	result := make([]PatternStats, 0, len(statsMap))
	for _, s := range statsMap {
		result = append(result, *s)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalHits > result[j].TotalHits
	})

	return result, nil
}

// GetTopPatterns returns the top N most used patterns.
func (t *Tracker) GetTopPatterns(n int) ([]PatternStats, error) {
	stats, err := t.GetPatternStats()
	if err != nil {
		return nil, err
	}

	if n > len(stats) {
		n = len(stats)
	}

	return stats[:n], nil
}

// GetColdPatterns returns patterns not used in the given duration.
func (t *Tracker) GetColdPatterns(since time.Duration) ([]PatternStats, error) {
	stats, err := t.GetPatternStats()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-since)
	var cold []PatternStats

	for _, s := range stats {
		if s.LastUsed.Before(cutoff) || s.LastUsed.IsZero() {
			cold = append(cold, s)
		}
	}

	return cold, nil
}

// Summary returns an overall analytics summary.
type Summary struct {
	TotalEvents      int            `json:"total_events"`
	TotalPatterns    int            `json:"total_patterns"`
	SearchEvents     int            `json:"search_events"`
	InjectEvents     int            `json:"inject_events"`
	TopPatterns      []PatternStats `json:"top_patterns"`
	ColdPatterns     int            `json:"cold_patterns"`
	AvgEffectiveness float64        `json:"avg_effectiveness"`
	Period           string         `json:"period"`
}

// GetSummary returns an analytics summary.
func (t *Tracker) GetSummary() (*Summary, error) {
	events, err := t.LoadEvents()
	if err != nil {
		return nil, err
	}

	stats, err := t.GetPatternStats()
	if err != nil {
		return nil, err
	}

	summary := &Summary{
		TotalEvents:   len(events),
		TotalPatterns: len(stats),
		Period:        "all-time",
	}

	// Count event types
	for _, e := range events {
		switch e.EventType {
		case EventSearch:
			summary.SearchEvents++
		case EventInject:
			summary.InjectEvents++
		}
	}

	// Top 5 patterns
	if len(stats) > 5 {
		summary.TopPatterns = stats[:5]
	} else {
		summary.TopPatterns = stats
	}

	// Cold patterns (not used in 30 days)
	cold, _ := t.GetColdPatterns(30 * 24 * time.Hour)
	summary.ColdPatterns = len(cold)

	// Average effectiveness
	var totalEff float64
	var effCount int
	for _, s := range stats {
		if s.Effectiveness > 0 {
			totalEff += s.Effectiveness
			effCount++
		}
	}
	if effCount > 0 {
		summary.AvgEffectiveness = totalEff / float64(effCount)
	}

	return summary, nil
}

// Helper functions

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
