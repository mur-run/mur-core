// Package analytics provides pattern usage tracking and effectiveness metrics.
package analytics

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store handles analytics data persistence.
type Store struct {
	db *sql.DB
}

// UsageEvent represents a pattern injection event.
type UsageEvent struct {
	ID          int64
	PatternID   string
	PatternName string
	Tool        string
	ContextType string // e.g., file extension
	InjectedAt  time.Time
	SessionID   string
}

// FeedbackEvent represents user feedback on a pattern.
type FeedbackEvent struct {
	ID         int64
	PatternID  string
	Rating     string // helpful, not_helpful, skip
	FeedbackAt time.Time
	Notes      string
}

// PatternStats holds aggregated stats for a pattern.
type PatternStats struct {
	PatternID       string
	PatternName     string
	UsageCount      int
	HelpfulCount    int
	NotHelpfulCount int
	SkipCount       int
	Effectiveness   float64 // helpful / (helpful + not_helpful)
	LastUsed        *time.Time
}

// DailyStats holds daily aggregates.
type DailyStats struct {
	PatternID       string
	Date            time.Time
	InjectionCount  int
	HelpfulCount    int
	NotHelpfulCount int
}

// NewStore creates a new analytics store.
func NewStore(dataDir string) (*Store, error) {
	dbPath := filepath.Join(dataDir, "analytics.db")

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open analytics database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate analytics database: %w", err)
	}

	return store, nil
}

// migrate creates the database schema.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pattern_usage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern_id TEXT NOT NULL,
		pattern_name TEXT NOT NULL,
		tool TEXT NOT NULL,
		context_type TEXT,
		injected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_usage_pattern ON pattern_usage(pattern_id);
	CREATE INDEX IF NOT EXISTS idx_usage_time ON pattern_usage(injected_at);

	CREATE TABLE IF NOT EXISTS pattern_feedback (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern_id TEXT NOT NULL,
		rating TEXT NOT NULL,
		feedback_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		notes TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_feedback_pattern ON pattern_feedback(pattern_id);

	CREATE TABLE IF NOT EXISTS pattern_daily_stats (
		pattern_id TEXT NOT NULL,
		date DATE NOT NULL,
		injection_count INT DEFAULT 0,
		helpful_count INT DEFAULT 0,
		not_helpful_count INT DEFAULT 0,
		PRIMARY KEY (pattern_id, date)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// RecordUsage records a pattern injection event.
func (s *Store) RecordUsage(event UsageEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO pattern_usage (pattern_id, pattern_name, tool, context_type, session_id)
		VALUES (?, ?, ?, ?, ?)
	`, event.PatternID, event.PatternName, event.Tool, event.ContextType, event.SessionID)
	if err != nil {
		return err
	}

	// Update daily stats
	today := time.Now().Format("2006-01-02")
	_, err = s.db.Exec(`
		INSERT INTO pattern_daily_stats (pattern_id, date, injection_count)
		VALUES (?, ?, 1)
		ON CONFLICT(pattern_id, date) DO UPDATE SET
			injection_count = injection_count + 1
	`, event.PatternID, today)

	return err
}

// RecordFeedback records user feedback on a pattern.
func (s *Store) RecordFeedback(event FeedbackEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO pattern_feedback (pattern_id, rating, notes)
		VALUES (?, ?, ?)
	`, event.PatternID, event.Rating, event.Notes)
	if err != nil {
		return err
	}

	// Update daily stats
	today := time.Now().Format("2006-01-02")
	switch event.Rating {
	case "helpful":
		_, err = s.db.Exec(`
			INSERT INTO pattern_daily_stats (pattern_id, date, helpful_count)
			VALUES (?, ?, 1)
			ON CONFLICT(pattern_id, date) DO UPDATE SET
				helpful_count = helpful_count + 1
		`, event.PatternID, today)
	case "not_helpful":
		_, err = s.db.Exec(`
			INSERT INTO pattern_daily_stats (pattern_id, date, not_helpful_count)
			VALUES (?, ?, 1)
			ON CONFLICT(pattern_id, date) DO UPDATE SET
				not_helpful_count = not_helpful_count + 1
		`, event.PatternID, today)
	}

	return err
}

// GetPatternStats returns aggregated stats for a specific pattern.
func (s *Store) GetPatternStats(patternID string) (*PatternStats, error) {
	stats := &PatternStats{PatternID: patternID}

	// Get usage count and last used
	err := s.db.QueryRow(`
		SELECT COUNT(*), MAX(injected_at), 
			COALESCE((SELECT pattern_name FROM pattern_usage WHERE pattern_id = ? LIMIT 1), '')
		FROM pattern_usage WHERE pattern_id = ?
	`, patternID, patternID).Scan(&stats.UsageCount, &stats.LastUsed, &stats.PatternName)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get feedback counts
	rows, err := s.db.Query(`
		SELECT rating, COUNT(*) FROM pattern_feedback 
		WHERE pattern_id = ? GROUP BY rating
	`, patternID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rating string
		var count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, err
		}
		switch rating {
		case "helpful":
			stats.HelpfulCount = count
		case "not_helpful":
			stats.NotHelpfulCount = count
		case "skip":
			stats.SkipCount = count
		}
	}

	// Calculate effectiveness
	total := stats.HelpfulCount + stats.NotHelpfulCount
	if total > 0 {
		stats.Effectiveness = float64(stats.HelpfulCount) / float64(total)
	}

	return stats, nil
}

// GetAllStats returns stats for all patterns, sorted by usage.
func (s *Store) GetAllStats(limit int) ([]*PatternStats, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT 
			u.pattern_id,
			u.pattern_name,
			COUNT(DISTINCT u.id) as usage_count,
			MAX(u.injected_at) as last_used
		FROM pattern_usage u
		GROUP BY u.pattern_id
		ORDER BY usage_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allStats []*PatternStats
	for rows.Next() {
		stats := &PatternStats{}
		if err := rows.Scan(&stats.PatternID, &stats.PatternName, &stats.UsageCount, &stats.LastUsed); err != nil {
			return nil, err
		}

		// Get feedback for this pattern
		feedbackRows, err := s.db.Query(`
			SELECT rating, COUNT(*) FROM pattern_feedback 
			WHERE pattern_id = ? GROUP BY rating
		`, stats.PatternID)
		if err != nil {
			return nil, err
		}

		for feedbackRows.Next() {
			var rating string
			var count int
			if err := feedbackRows.Scan(&rating, &count); err != nil {
				feedbackRows.Close()
				return nil, err
			}
			switch rating {
			case "helpful":
				stats.HelpfulCount = count
			case "not_helpful":
				stats.NotHelpfulCount = count
			case "skip":
				stats.SkipCount = count
			}
		}
		feedbackRows.Close()

		// Calculate effectiveness
		total := stats.HelpfulCount + stats.NotHelpfulCount
		if total > 0 {
			stats.Effectiveness = float64(stats.HelpfulCount) / float64(total)
		}

		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// GetRecentUsage returns recently used patterns.
func (s *Store) GetRecentUsage(limit int) ([]*UsageEvent, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
		SELECT id, pattern_id, pattern_name, tool, context_type, injected_at, session_id
		FROM pattern_usage
		ORDER BY injected_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*UsageEvent
	for rows.Next() {
		e := &UsageEvent{}
		var contextType, sessionID sql.NullString
		if err := rows.Scan(&e.ID, &e.PatternID, &e.PatternName, &e.Tool, &contextType, &e.InjectedAt, &sessionID); err != nil {
			return nil, err
		}
		if contextType.Valid {
			e.ContextType = contextType.String
		}
		if sessionID.Valid {
			e.SessionID = sessionID.String
		}
		events = append(events, e)
	}

	return events, nil
}

// GetOverallStats returns summary statistics.
func (s *Store) GetOverallStats(days int) (*OverallStats, error) {
	if days <= 0 {
		days = 30
	}

	stats := &OverallStats{}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// Total patterns with usage
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT pattern_id) FROM pattern_usage WHERE injected_at >= ?
	`, since).Scan(&stats.ActivePatterns)
	if err != nil {
		return nil, err
	}

	// Total injections
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pattern_usage WHERE injected_at >= ?
	`, since).Scan(&stats.TotalInjections)
	if err != nil {
		return nil, err
	}

	// Total unique patterns
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT pattern_id) FROM pattern_usage
	`).Scan(&stats.TotalPatterns)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// OverallStats holds summary metrics.
type OverallStats struct {
	TotalPatterns   int
	ActivePatterns  int // Used in last N days
	TotalInjections int
}

// GetUsageByTool returns usage breakdown by tool.
func (s *Store) GetUsageByTool(patternID string) (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT tool, COUNT(*) FROM pattern_usage 
		WHERE pattern_id = ? GROUP BY tool ORDER BY COUNT(*) DESC
	`, patternID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var tool string
		var count int
		if err := rows.Scan(&tool, &count); err != nil {
			return nil, err
		}
		result[tool] = count
	}
	return result, nil
}

// GetUsageByContext returns usage breakdown by context/file type.
func (s *Store) GetUsageByContext(patternID string) (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT COALESCE(context_type, 'unknown'), COUNT(*) FROM pattern_usage 
		WHERE pattern_id = ? GROUP BY context_type ORDER BY COUNT(*) DESC
	`, patternID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var context string
		var count int
		if err := rows.Scan(&context, &count); err != nil {
			return nil, err
		}
		result[context] = count
	}
	return result, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
