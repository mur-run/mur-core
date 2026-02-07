// Package inject provides pattern injection and effectiveness tracking.
package inject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// UsageRecord tracks a single pattern usage.
type UsageRecord struct {
	// Pattern ID
	PatternID string `json:"pattern_id"`
	// Pattern name (for readability)
	PatternName string `json:"pattern_name"`
	// When the pattern was injected
	Timestamp time.Time `json:"timestamp"`
	// Project context
	ProjectType string `json:"project_type,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	// Prompt that triggered the injection (truncated)
	PromptPreview string `json:"prompt_preview,omitempty"`
	// Whether the run succeeded
	Success bool `json:"success"`
	// User feedback (if provided)
	Feedback *Feedback `json:"feedback,omitempty"`
}

// Feedback represents user feedback on pattern usefulness.
type Feedback struct {
	// Rating: -1 (unhelpful), 0 (neutral), 1 (helpful)
	Rating int `json:"rating"`
	// Optional comment
	Comment string `json:"comment,omitempty"`
	// When feedback was given
	Timestamp time.Time `json:"timestamp"`
}

// EffectivenessStats holds aggregated stats for a pattern.
type EffectivenessStats struct {
	PatternID   string  `json:"pattern_id"`
	PatternName string  `json:"pattern_name"`
	TotalUses   int     `json:"total_uses"`
	SuccessRate float64 `json:"success_rate"`
	// Feedback stats
	HelpfulCount   int     `json:"helpful_count"`
	UnhelpfulCount int     `json:"unhelpful_count"`
	NeutralCount   int     `json:"neutral_count"`
	FeedbackScore  float64 `json:"feedback_score"` // -1.0 to 1.0
	// Computed effectiveness (combines success rate + feedback)
	Effectiveness float64 `json:"effectiveness"`
	// Last used
	LastUsed time.Time `json:"last_used"`
}

// Tracker tracks pattern usage and effectiveness.
type Tracker struct {
	store   *pattern.Store
	dataDir string
	mu      sync.Mutex
}

// NewTracker creates a new Tracker.
func NewTracker(store *pattern.Store, dataDir string) *Tracker {
	return &Tracker{
		store:   store,
		dataDir: dataDir,
	}
}

// DefaultTracker returns a Tracker using default paths.
func DefaultTracker() (*Tracker, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	dataDir := filepath.Join(home, ".mur", "tracking")

	return &Tracker{
		store:   pattern.NewStore(patternsDir),
		dataDir: dataDir,
	}, nil
}

// usageFile returns the path to the usage log file.
func (t *Tracker) usageFile() string {
	return filepath.Join(t.dataDir, "usage.jsonl")
}

// RecordUsage records that patterns were used in a run.
func (t *Tracker) RecordUsage(patterns []*pattern.Pattern, ctx *ProjectContext, prompt string, success bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(t.dataDir, 0755); err != nil {
		return fmt.Errorf("cannot create tracking directory: %w", err)
	}

	// Open file for appending
	f, err := os.OpenFile(t.usageFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open usage file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Truncate prompt for storage
	promptPreview := prompt
	if len(promptPreview) > 100 {
		promptPreview = promptPreview[:100] + "..."
	}

	// Record each pattern
	encoder := json.NewEncoder(f)
	for _, p := range patterns {
		record := UsageRecord{
			PatternID:     p.ID,
			PatternName:   p.Name,
			Timestamp:     time.Now(),
			PromptPreview: promptPreview,
			Success:       success,
		}
		if ctx != nil {
			record.ProjectType = ctx.ProjectType
			record.ProjectName = ctx.ProjectName
		}

		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("cannot write usage record: %w", err)
		}

		// Update pattern's usage count
		_ = t.store.RecordUsage(p.Name)
	}

	return nil
}

// RecordFeedback records user feedback for a pattern.
func (t *Tracker) RecordFeedback(patternName string, rating int, comment string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate rating
	if rating < -1 || rating > 1 {
		return fmt.Errorf("rating must be -1, 0, or 1")
	}

	// Get pattern
	p, err := t.store.Get(patternName)
	if err != nil {
		return err
	}

	// Read all records and find the most recent one for this pattern
	records, err := t.readUsageRecords()
	if err != nil {
		return err
	}

	// Find most recent record for this pattern without feedback
	var targetIdx = -1
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].PatternID == p.ID && records[i].Feedback == nil {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return fmt.Errorf("no recent usage found for pattern: %s", patternName)
	}

	// Add feedback
	records[targetIdx].Feedback = &Feedback{
		Rating:    rating,
		Comment:   comment,
		Timestamp: time.Now(),
	}

	// Rewrite file
	return t.writeUsageRecords(records)
}

// GetStats returns effectiveness stats for all patterns.
func (t *Tracker) GetStats() ([]EffectivenessStats, error) {
	records, err := t.readUsageRecords()
	if err != nil {
		return nil, err
	}

	// Aggregate by pattern
	statsMap := make(map[string]*EffectivenessStats)

	for _, r := range records {
		stats, ok := statsMap[r.PatternID]
		if !ok {
			stats = &EffectivenessStats{
				PatternID:   r.PatternID,
				PatternName: r.PatternName,
			}
			statsMap[r.PatternID] = stats
		}

		stats.TotalUses++
		if r.Success {
			stats.SuccessRate += 1.0
		}
		if r.Timestamp.After(stats.LastUsed) {
			stats.LastUsed = r.Timestamp
		}

		if r.Feedback != nil {
			switch r.Feedback.Rating {
			case 1:
				stats.HelpfulCount++
			case -1:
				stats.UnhelpfulCount++
			case 0:
				stats.NeutralCount++
			}
		}
	}

	// Calculate final scores
	result := make([]EffectivenessStats, 0, len(statsMap))
	for _, stats := range statsMap {
		// Success rate
		if stats.TotalUses > 0 {
			stats.SuccessRate /= float64(stats.TotalUses)
		}

		// Feedback score
		totalFeedback := stats.HelpfulCount + stats.UnhelpfulCount + stats.NeutralCount
		if totalFeedback > 0 {
			stats.FeedbackScore = float64(stats.HelpfulCount-stats.UnhelpfulCount) / float64(totalFeedback)
		}

		// Combined effectiveness
		// Weight: 30% success rate + 70% feedback score
		// If no feedback, use success rate only
		if totalFeedback > 0 {
			stats.Effectiveness = stats.SuccessRate*0.3 + (stats.FeedbackScore+1.0)/2.0*0.7
		} else {
			stats.Effectiveness = stats.SuccessRate
		}

		result = append(result, *stats)
	}

	return result, nil
}

// GetPatternStats returns stats for a specific pattern.
func (t *Tracker) GetPatternStats(patternName string) (*EffectivenessStats, error) {
	p, err := t.store.Get(patternName)
	if err != nil {
		return nil, err
	}

	allStats, err := t.GetStats()
	if err != nil {
		return nil, err
	}

	for _, stats := range allStats {
		if stats.PatternID == p.ID {
			return &stats, nil
		}
	}

	// No usage records yet
	return &EffectivenessStats{
		PatternID:     p.ID,
		PatternName:   p.Name,
		Effectiveness: 0.5, // Default
	}, nil
}

// UpdatePatternEffectiveness updates a pattern's effectiveness based on tracked data.
func (t *Tracker) UpdatePatternEffectiveness(patternName string) error {
	stats, err := t.GetPatternStats(patternName)
	if err != nil {
		return err
	}

	p, err := t.store.Get(patternName)
	if err != nil {
		return err
	}

	// Update effectiveness
	p.Learning.Effectiveness = stats.Effectiveness

	return t.store.Update(p)
}

// UpdateAllEffectiveness updates effectiveness for all patterns with usage data.
func (t *Tracker) UpdateAllEffectiveness() error {
	allStats, err := t.GetStats()
	if err != nil {
		return err
	}

	for _, stats := range allStats {
		p, err := t.store.Get(stats.PatternName)
		if err != nil {
			continue // Pattern may have been deleted
		}

		p.Learning.Effectiveness = stats.Effectiveness
		if err := t.store.Update(p); err != nil {
			continue // Non-fatal
		}
	}

	return nil
}

// readUsageRecords reads all usage records from the log.
func (t *Tracker) readUsageRecords() ([]UsageRecord, error) {
	path := t.usageFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []UsageRecord{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read usage file: %w", err)
	}

	var records []UsageRecord
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	for decoder.More() {
		var r UsageRecord
		if err := decoder.Decode(&r); err != nil {
			continue // Skip malformed records
		}
		records = append(records, r)
	}

	return records, nil
}

// writeUsageRecords rewrites all usage records.
func (t *Tracker) writeUsageRecords(records []UsageRecord) error {
	// Ensure data directory exists
	if err := os.MkdirAll(t.dataDir, 0755); err != nil {
		return fmt.Errorf("cannot create tracking directory: %w", err)
	}

	f, err := os.Create(t.usageFile())
	if err != nil {
		return fmt.Errorf("cannot create usage file: %w", err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	for _, r := range records {
		if err := encoder.Encode(r); err != nil {
			return fmt.Errorf("cannot write usage record: %w", err)
		}
	}

	return nil
}

// PruneOldRecords removes usage records older than the given duration.
func (t *Tracker) PruneOldRecords(maxAge time.Duration) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	records, err := t.readUsageRecords()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	var kept []UsageRecord
	pruned := 0

	for _, r := range records {
		if r.Timestamp.After(cutoff) {
			kept = append(kept, r)
		} else {
			pruned++
		}
	}

	if pruned > 0 {
		if err := t.writeUsageRecords(kept); err != nil {
			return 0, err
		}
	}

	return pruned, nil
}
