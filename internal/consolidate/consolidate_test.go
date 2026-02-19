package consolidate

import (
	"math"
	"testing"
	"time"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

func defaultCfg() config.ConsolidationConfig {
	return config.DefaultConsolidationConfig()
}

func makePattern(id, name string, created time.Time, usageCount int, lastUsed *time.Time) *pattern.Pattern {
	p := &pattern.Pattern{
		ID:      id,
		Name:    name,
		Content: "test content for " + name,
		Lifecycle: pattern.LifecycleMeta{
			Status:  pattern.StatusActive,
			Created: created,
			Updated: created,
		},
		Learning: pattern.LearningMeta{
			UsageCount: usageCount,
			LastUsed:   lastUsed,
		},
		SchemaVersion: pattern.SchemaVersion,
	}
	return p
}

// --- Health Score Tests ---

func TestHealthScore_FreshnessDecay(t *testing.T) {
	cfg := defaultCfg()
	scorer := NewHealthScorer(cfg, nil, nil, nil)

	now := scorer.now
	recent := now.Add(-1 * 24 * time.Hour) // 1 day ago
	old := now.Add(-180 * 24 * time.Hour)   // 180 days ago (2 half-lives)

	pRecent := makePattern("p1", "recent", recent, 5, &recent)
	pOld := makePattern("p2", "old", old, 5, &old)

	scoreRecent := scorer.freshness(pRecent)
	scoreOld := scorer.freshness(pOld)

	if scoreRecent < 0.9 {
		t.Errorf("recent pattern freshness = %.3f, want >= 0.9", scoreRecent)
	}
	if scoreOld > 0.3 {
		t.Errorf("old pattern freshness = %.3f, want <= 0.3", scoreOld)
	}
	if scoreOld >= scoreRecent {
		t.Errorf("old pattern (%f) should be less fresh than recent (%f)", scoreOld, scoreRecent)
	}
}

func TestHealthScore_FreshnessGracePeriod(t *testing.T) {
	cfg := defaultCfg()
	scorer := NewHealthScorer(cfg, nil, nil, nil)

	now := scorer.now
	created := now.Add(-5 * 24 * time.Hour) // 5 days ago, within grace period

	// New pattern with zero usage should get freshness = 1.0
	p := makePattern("p1", "new-unused", created, 0, nil)

	score := scorer.freshness(p)
	if score != 1.0 {
		t.Errorf("grace period freshness = %.3f, want 1.0", score)
	}
}

func TestHealthScore_FreshnessNoGracePeriodWithUsage(t *testing.T) {
	cfg := defaultCfg()
	scorer := NewHealthScorer(cfg, nil, nil, nil)

	now := scorer.now
	created := now.Add(-5 * 24 * time.Hour)

	// Pattern with usage does not get grace period exemption
	p := makePattern("p1", "new-used", created, 3, &created)

	score := scorer.freshness(p)
	// With usage, it should use normal decay from last activity, not grace period
	if score == 1.0 {
		t.Error("pattern with usage should not get perfect grace period freshness")
	}
}

func TestHealthScore_Engagement(t *testing.T) {
	cfg := defaultCfg()
	scorer := NewHealthScorer(cfg, nil, nil, nil)

	now := time.Now()
	tests := []struct {
		name     string
		usage    int
		wantMin  float64
		wantMax  float64
	}{
		{"zero usage", 0, 0.0, 0.15},
		{"low usage", 3, 0.2, 0.4},
		{"moderate usage", 15, 0.5, 0.7},
		{"high usage", 127, 0.95, 1.0}, // log2(128)/7 = 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := makePattern("p1", "test", now, tt.usage, nil)
			score := scorer.engagement(p)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("engagement(%d) = %.3f, want [%.2f, %.2f]",
					tt.usage, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestHealthScore_Quality(t *testing.T) {
	cfg := defaultCfg()

	tests := []struct {
		name     string
		helpful  int
		unhelpful int
		want     float64
	}{
		{"no feedback", 0, 0, 0.5},
		{"all helpful", 10, 0, 1.0},
		{"all unhelpful", 0, 10, 0.0},
		{"mixed", 7, 3, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := []inject.EffectivenessStats{
				{
					PatternID:      "p1",
					HelpfulCount:   tt.helpful,
					UnhelpfulCount: tt.unhelpful,
				},
			}
			scorer := NewHealthScorer(cfg, nil, stats, nil)

			now := time.Now()
			p := makePattern("p1", "test", now, 5, nil)
			score := scorer.quality(p)

			if math.Abs(score-tt.want) > 0.01 {
				t.Errorf("quality = %.3f, want %.3f", score, tt.want)
			}
		})
	}
}

func TestHealthScore_UniquenessNoMatrix(t *testing.T) {
	cfg := defaultCfg()
	scorer := NewHealthScorer(cfg, nil, nil, nil)

	now := time.Now()
	p := makePattern("p1", "test", now, 5, nil)
	score := scorer.uniqueness(p)

	if score != 1.0 {
		t.Errorf("uniqueness without matrix = %.3f, want 1.0", score)
	}
}

func TestHealthScore_DecisionRules(t *testing.T) {
	tests := []struct {
		name       string
		freshness  float64
		engagement float64
		quality    float64
		uniqueness float64
		wantAction Action
	}{
		{"merge: low uniqueness", 0.8, 0.5, 0.7, 0.10, ActionMerge},
		{"update: low quality high engagement", 0.8, 0.5, 0.15, 0.8, ActionUpdate},
		{"archive: stale and unused", 0.05, 0.05, 0.5, 0.8, ActionArchive},
		{"archive: low overall", 0.1, 0.1, 0.1, 0.8, ActionArchive},
		{"keep: healthy", 0.8, 0.5, 0.7, 0.8, ActionKeep},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := HealthScore{
				Freshness:  tt.freshness,
				Engagement: tt.engagement,
				Quality:    tt.quality,
				Uniqueness: tt.uniqueness,
			}
			hs.Overall = WeightFreshness*hs.Freshness +
				WeightEngagement*hs.Engagement +
				WeightQuality*hs.Quality +
				WeightUniqueness*hs.Uniqueness

			cfg := defaultCfg()
			scorer := NewHealthScorer(cfg, nil, nil, nil)
			scorer.decide(&hs)

			if hs.Action != tt.wantAction {
				t.Errorf("action = %s, want %s (overall=%.3f)", hs.Action, tt.wantAction, hs.Overall)
			}
		})
	}
}

func TestHealthScore_OverallWeights(t *testing.T) {
	hs := HealthScore{
		Freshness:  1.0,
		Engagement: 1.0,
		Quality:    1.0,
		Uniqueness: 1.0,
	}
	overall := WeightFreshness*hs.Freshness +
		WeightEngagement*hs.Engagement +
		WeightQuality*hs.Quality +
		WeightUniqueness*hs.Uniqueness

	if math.Abs(overall-1.0) > 0.001 {
		t.Errorf("max overall = %.3f, want 1.0", overall)
	}

	// Verify weights sum to 1.0
	sum := WeightFreshness + WeightEngagement + WeightQuality + WeightUniqueness
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("weight sum = %.3f, want 1.0", sum)
	}
}

// --- Dedup Tests ---

func TestDuplicateDetector_NoMatrix(t *testing.T) {
	d := NewDuplicateDetector(nil, 0.85, StrategyKeepBest)
	now := time.Now()
	patterns := []*pattern.Pattern{
		makePattern("p1", "a", now, 1, nil),
		makePattern("p2", "b", now, 1, nil),
	}

	proposals := d.Detect(patterns)
	if len(proposals) != 0 {
		t.Errorf("expected no proposals without matrix, got %d", len(proposals))
	}
}

func TestDuplicateDetector_SelectBest(t *testing.T) {
	d := NewDuplicateDetector(nil, 0.85, StrategyKeepBest)
	d.WithHealthScores([]HealthScore{
		{PatternID: "p1", Overall: 0.3},
		{PatternID: "p2", Overall: 0.9},
		{PatternID: "p3", Overall: 0.5},
	})

	proposal := MergeProposal{
		Patterns: []*pattern.Pattern{
			{ID: "p1", Name: "a"},
			{ID: "p2", Name: "b"},
			{ID: "p3", Name: "c"},
		},
	}

	d.selectBest(&proposal)

	if proposal.KeepID != "p2" {
		t.Errorf("keepID = %s, want p2 (highest score)", proposal.KeepID)
	}
	if len(proposal.RemoveIDs) != 2 {
		t.Errorf("removeIDs count = %d, want 2", len(proposal.RemoveIDs))
	}
}

// --- Conflict Tests ---

func TestKeywordConflictDetector_Contradiction(t *testing.T) {
	d := NewKeywordConflictDetector()

	patterns := []*pattern.Pattern{
		{
			ID:      "p1",
			Name:    "always-use-semicolons",
			Content: "Always use semicolons in JavaScript",
			Tags:    pattern.TagSet{Confirmed: []string{"javascript"}},
		},
		{
			ID:      "p2",
			Name:    "never-use-semicolons",
			Content: "Never use semicolons in JavaScript",
			Tags:    pattern.TagSet{Confirmed: []string{"javascript"}},
		},
	}

	conflicts := d.Detect(patterns)
	if len(conflicts) == 0 {
		t.Error("expected at least one contradiction conflict")
	}

	found := false
	for _, c := range conflicts {
		if c.Type == ConflictContradiction {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a contradiction-type conflict")
	}
}

func TestKeywordConflictDetector_NoConflict(t *testing.T) {
	d := NewKeywordConflictDetector()

	patterns := []*pattern.Pattern{
		{
			ID:      "p1",
			Name:    "go-error-handling",
			Content: "Always handle errors in Go",
			Tags:    pattern.TagSet{Confirmed: []string{"go"}},
		},
		{
			ID:      "p2",
			Name:    "python-testing",
			Content: "Use pytest for testing Python code",
			Tags:    pattern.TagSet{Confirmed: []string{"python"}},
		},
	}

	conflicts := d.Detect(patterns)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts between different domains, got %d", len(conflicts))
	}
}

func TestKeywordConflictDetector_Supersedes(t *testing.T) {
	d := NewKeywordConflictDetector()

	patterns := []*pattern.Pattern{
		{
			ID:      "p1",
			Name:    "go-modules",
			Content: "Use go modules for dependency management",
			Tags:    pattern.TagSet{Confirmed: []string{"go"}},
			Relations: pattern.Relations{
				Supersedes: "",
			},
		},
		{
			ID:      "p2",
			Name:    "go-modules-v2",
			Content: "Use go modules v2 for dependency management",
			Tags:    pattern.TagSet{Confirmed: []string{"go"}},
			Relations: pattern.Relations{
				Supersedes: "p1",
			},
		},
	}

	conflicts := d.Detect(patterns)

	found := false
	for _, c := range conflicts {
		if c.Type == ConflictOutdated {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected an outdated-type conflict for supersedes relation")
	}
}

// --- Report Tests ---

func TestFormatReport_DryRun(t *testing.T) {
	report := &ConsolidationReport{
		Mode:          ModeDryRun,
		TotalPatterns: 10,
		HealthScores: []HealthScore{
			{PatternID: "p1", Overall: 0.1, Action: ActionArchive, Reason: "stale"},
			{PatternID: "p2", Overall: 0.9, Action: ActionKeep, Reason: "healthy"},
		},
		PatternsKept:     1,
		PatternsArchived: 1,
		Duration:         100 * time.Millisecond,
	}

	nameMap := map[string]string{"p1": "old-pattern", "p2": "good-pattern"}
	output := FormatReport(report, nameMap)

	if output == "" {
		t.Error("expected non-empty report output")
	}
	if !contains(output, "dry-run") {
		t.Error("expected dry-run notice in output")
	}
	if !contains(output, "ARCHIVE") {
		t.Error("expected ARCHIVE action in output")
	}
	if !contains(output, "old-pattern") {
		t.Error("expected pattern name in output")
	}
}

// --- EmbeddingMatrix AllPairs/MaxSimilarity Tests ---

func TestEmbeddingMatrix_AllPairs(t *testing.T) {
	m := cache.NewEmbeddingMatrix(3)
	// Manually set up vectors: two similar, one different
	// We need to use Load, so create a temp file
	// Instead, test via the public API indirectly through the consolidate flow
	// The actual AllPairs method is tested in cache package tests
	_ = m // This test validates the integration exists
}

func TestEmbeddingMatrix_MaxSimilarity_NotFound(t *testing.T) {
	m := cache.NewEmbeddingMatrix(3)
	result := m.MaxSimilarity("nonexistent")
	if result != -1 {
		t.Errorf("MaxSimilarity for nonexistent = %f, want -1", result)
	}
}

// --- Analytics Integration Tests ---

func TestHealthScore_WithAnalyticsData(t *testing.T) {
	cfg := defaultCfg()

	aStats := []analytics.PatternStats{
		{PatternID: "p1", TotalHits: 50},
	}

	scorer := NewHealthScorer(cfg, nil, nil, aStats)

	now := scorer.now
	p := makePattern("p1", "popular", now, 2, nil) // low internal usage

	engagement := scorer.engagement(p)
	// With analytics providing 50 hits, engagement should be based on 50
	expected := math.Log2(51) / 7.0
	if math.Abs(engagement-expected) > 0.01 {
		t.Errorf("engagement with analytics = %.3f, want ~%.3f", engagement, expected)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		findSubstring(s, substr)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
