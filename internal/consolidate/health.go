// Package consolidate provides pattern lifecycle management through health scoring,
// duplicate detection, conflict resolution, and automated consolidation.
package consolidate

import (
	"math"
	"time"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

// Action represents the recommended action for a pattern.
type Action string

const (
	ActionKeep    Action = "keep"
	ActionArchive Action = "archive"
	ActionMerge   Action = "merge"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
)

// Weights for health score dimensions.
const (
	WeightFreshness  = 0.25
	WeightEngagement = 0.30
	WeightQuality    = 0.30
	WeightUniqueness = 0.15
)

// HealthScore represents the computed health of a single pattern.
type HealthScore struct {
	PatternID  string  `json:"pattern_id"`
	Freshness  float64 `json:"freshness"`
	Engagement float64 `json:"engagement"`
	Quality    float64 `json:"quality"`
	Uniqueness float64 `json:"uniqueness"`
	Overall    float64 `json:"overall"`
	Action     Action  `json:"action"`
	Reason     string  `json:"reason"`
}

// HealthScorer computes health scores for patterns.
type HealthScorer struct {
	cfg       config.ConsolidationConfig
	matrix    *cache.EmbeddingMatrix
	stats     map[string]*inject.EffectivenessStats
	analytics map[string]*analytics.PatternStats
	now       time.Time
}

// NewHealthScorer creates a HealthScorer with the given configuration and data sources.
func NewHealthScorer(
	cfg config.ConsolidationConfig,
	matrix *cache.EmbeddingMatrix,
	effectivenessStats []inject.EffectivenessStats,
	analyticsStats []analytics.PatternStats,
) *HealthScorer {
	statsMap := make(map[string]*inject.EffectivenessStats, len(effectivenessStats))
	for i := range effectivenessStats {
		statsMap[effectivenessStats[i].PatternID] = &effectivenessStats[i]
	}

	analyticsMap := make(map[string]*analytics.PatternStats, len(analyticsStats))
	for i := range analyticsStats {
		analyticsMap[analyticsStats[i].PatternID] = &analyticsStats[i]
	}

	return &HealthScorer{
		cfg:       cfg,
		matrix:    matrix,
		stats:     statsMap,
		analytics: analyticsMap,
		now:       time.Now(),
	}
}

// Score computes the HealthScore for a single pattern.
func (s *HealthScorer) Score(p *pattern.Pattern) HealthScore {
	hs := HealthScore{PatternID: p.ID}

	hs.Freshness = s.freshness(p)
	hs.Engagement = s.engagement(p)
	hs.Quality = s.quality(p)
	hs.Uniqueness = s.uniqueness(p)

	hs.Overall = WeightFreshness*hs.Freshness +
		WeightEngagement*hs.Engagement +
		WeightQuality*hs.Quality +
		WeightUniqueness*hs.Uniqueness

	s.decide(&hs)
	return hs
}

// ScoreAll computes health scores for all given patterns.
func (s *HealthScorer) ScoreAll(patterns []*pattern.Pattern) []HealthScore {
	scores := make([]HealthScore, len(patterns))
	for i, p := range patterns {
		scores[i] = s.Score(p)
	}
	return scores
}

// freshness computes time decay with configurable half-life.
// Grace period: patterns created < GracePeriodDays ago get 1.0 if unused.
func (s *HealthScorer) freshness(p *pattern.Pattern) float64 {
	halfLife := float64(s.cfg.DecayHalfLifeDays) * 24 * float64(time.Hour)
	gracePeriod := time.Duration(s.cfg.GracePeriodDays) * 24 * time.Hour

	// Use the most recent activity timestamp
	lastActivity := p.Lifecycle.Updated
	if p.Learning.LastUsed != nil && p.Learning.LastUsed.After(lastActivity) {
		lastActivity = *p.Learning.LastUsed
	}

	age := s.now.Sub(p.Lifecycle.Created)
	sinceActivity := s.now.Sub(lastActivity)

	// Grace period for new patterns
	if age < gracePeriod && p.Learning.UsageCount == 0 {
		return 1.0
	}

	// Exponential decay: 0.5^(t/half_life)
	decay := math.Pow(0.5, float64(sinceActivity)/halfLife)
	return clamp(decay, 0, 1)
}

// engagement computes log-scaled usage score: log2(usage+1)/7.
func (s *HealthScorer) engagement(p *pattern.Pattern) float64 {
	usage := p.Learning.UsageCount

	// Also incorporate analytics data if available
	if as, ok := s.analytics[p.ID]; ok {
		if as.TotalHits > usage {
			usage = as.TotalHits
		}
	}

	score := math.Log2(float64(usage)+1) / 7.0
	return clamp(score, 0, 1)
}

// quality computes feedback-based quality: helpful/(helpful+unhelpful), default 0.5.
func (s *HealthScorer) quality(p *pattern.Pattern) float64 {
	es, ok := s.stats[p.ID]
	if !ok {
		return 0.5 // default when no feedback
	}

	total := es.HelpfulCount + es.UnhelpfulCount
	if total == 0 {
		return 0.5
	}

	return float64(es.HelpfulCount) / float64(total)
}

// uniqueness computes 1 - max_cosine_similarity with any other pattern.
func (s *HealthScorer) uniqueness(p *pattern.Pattern) float64 {
	if s.matrix == nil || !s.matrix.IsLoaded() || s.matrix.Len() < 2 {
		return 1.0 // assume unique if no embeddings
	}

	maxSim := s.matrix.MaxSimilarity(p.ID)
	if maxSim < 0 {
		return 1.0 // pattern not in matrix
	}

	return clamp(1.0-maxSim, 0, 1)
}

// decide applies decision rules to set Action and Reason.
func (s *HealthScorer) decide(hs *HealthScore) {
	// Priority order per spec
	switch {
	case hs.Uniqueness < 0.15:
		hs.Action = ActionMerge
		hs.Reason = "high similarity with another pattern"
	case hs.Quality < 0.2 && hs.Engagement > 0.3:
		hs.Action = ActionUpdate
		hs.Reason = "low quality but actively used"
	case hs.Freshness < 0.1 && hs.Engagement < 0.1:
		hs.Action = ActionArchive
		hs.Reason = "stale and unused"
	case hs.Overall < 0.25:
		hs.Action = ActionArchive
		hs.Reason = "low overall health score"
	default:
		hs.Action = ActionKeep
		hs.Reason = "healthy"
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
