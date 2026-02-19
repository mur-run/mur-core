package consolidate

import (
	"fmt"
	"time"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/analytics"
	"github.com/mur-run/mur-core/internal/core/inject"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

// Mode controls how consolidation actions are applied.
type Mode string

const (
	ModeDryRun      Mode = "dry-run"
	ModeAuto        Mode = "auto"
	ModeInteractive Mode = "interactive"
)

// ConsolidationReport holds the results of a consolidation run.
type ConsolidationReport struct {
	Timestamp        time.Time       `json:"timestamp"`
	Mode             Mode            `json:"mode"`
	TotalPatterns    int             `json:"total_patterns"`
	HealthScores     []HealthScore   `json:"health_scores"`
	MergeProposals   []MergeProposal `json:"merge_proposals"`
	Conflicts        []Conflict      `json:"conflicts"`
	ActionsApplied   int             `json:"actions_applied"`
	PatternsKept     int             `json:"patterns_kept"`
	PatternsArchived int             `json:"patterns_archived"`
	PatternsMerged   int             `json:"patterns_merged"`
	PatternsUpdated  int             `json:"patterns_updated"`
	Duration         time.Duration   `json:"duration"`
}

// Consolidator orchestrates the pattern consolidation process.
type Consolidator struct {
	cfg              config.ConsolidationConfig
	store            *pattern.Store
	patternCache     *cache.PatternCache
	embeddingMatrix  *cache.EmbeddingMatrix
	injTracker       *inject.Tracker
	analyticsTracker *analytics.Tracker
	conflictDetector ConflictDetector
}

// NewConsolidator creates a new Consolidator.
func NewConsolidator(
	cfg config.ConsolidationConfig,
	store *pattern.Store,
	patternCache *cache.PatternCache,
	embeddingMatrix *cache.EmbeddingMatrix,
	injTracker *inject.Tracker,
	analyticsTracker *analytics.Tracker,
) *Consolidator {
	return &Consolidator{
		cfg:              cfg,
		store:            store,
		patternCache:     patternCache,
		embeddingMatrix:  embeddingMatrix,
		injTracker:       injTracker,
		analyticsTracker: analyticsTracker,
		conflictDetector: NewKeywordConflictDetector(),
	}
}

// WithConflictDetector sets a custom conflict detector (e.g., LLM-based).
func (c *Consolidator) WithConflictDetector(d ConflictDetector) {
	c.conflictDetector = d
}

// Run executes the full consolidation pipeline.
func (c *Consolidator) Run(mode Mode, force bool) (*ConsolidationReport, error) {
	start := time.Now()

	// Load patterns
	patterns := c.patternCache.Active()
	if len(patterns) == 0 {
		return nil, fmt.Errorf("no active patterns found")
	}

	// Check minimum patterns threshold (unless forced)
	if !force && len(patterns) < c.cfg.MinPatternsBeforeRun {
		return nil, fmt.Errorf("only %d patterns found (minimum: %d); use --force to override",
			len(patterns), c.cfg.MinPatternsBeforeRun)
	}

	// Phase 1: Gather analytics data
	effectivenessStats, err := c.getEffectivenessStats()
	if err != nil {
		effectivenessStats = nil // non-fatal
	}

	analyticsStats, err := c.getAnalyticsStats()
	if err != nil {
		analyticsStats = nil // non-fatal
	}

	// Phase 2: Score health
	scorer := NewHealthScorer(c.cfg, c.embeddingMatrix, effectivenessStats, analyticsStats)
	healthScores := scorer.ScoreAll(patterns)

	// Phase 3: Detect duplicates
	var mergeProposals []MergeProposal
	strategy := parseMergeStrategy(c.cfg.AutoMerge)
	if strategy != "" {
		detector := NewDuplicateDetector(c.embeddingMatrix, c.cfg.MergeThreshold, strategy)
		detector.WithHealthScores(healthScores)
		mergeProposals = detector.Detect(patterns)
	}

	// Phase 4: Detect conflicts
	var conflicts []Conflict
	if c.conflictDetector != nil {
		conflicts = c.conflictDetector.Detect(patterns)
	}

	// Build report
	report := &ConsolidationReport{
		Timestamp:      time.Now(),
		Mode:           mode,
		TotalPatterns:  len(patterns),
		HealthScores:   healthScores,
		MergeProposals: mergeProposals,
		Conflicts:      conflicts,
	}

	// Phase 5: Apply actions (only in auto mode)
	if mode == ModeAuto {
		c.applyActions(report, patterns, healthScores, mergeProposals)
	}

	// Count action summary
	for _, hs := range healthScores {
		switch hs.Action {
		case ActionKeep:
			report.PatternsKept++
		case ActionArchive:
			report.PatternsArchived++
		case ActionMerge:
			report.PatternsMerged++
		case ActionUpdate:
			report.PatternsUpdated++
		}
	}

	report.Duration = time.Since(start)
	return report, nil
}

// applyActions executes safe automatic actions (archive, keep-best merge).
func (c *Consolidator) applyActions(report *ConsolidationReport, patterns []*pattern.Pattern, scores []HealthScore, proposals []MergeProposal) {
	patternMap := make(map[string]*pattern.Pattern, len(patterns))
	for _, p := range patterns {
		patternMap[p.ID] = p
	}

	now := time.Now()

	// Apply archive actions
	if c.cfg.AutoArchive {
		for _, hs := range scores {
			if hs.Action == ActionArchive {
				p, ok := patternMap[hs.PatternID]
				if !ok {
					continue
				}
				p.Lifecycle.Status = pattern.StatusArchived
				p.Lifecycle.DeprecationReason = "auto-archived: " + hs.Reason
				p.Health.Score = hs.Overall
				p.Health.LastConsolidated = &now
				if err := c.store.Update(p); err == nil {
					report.ActionsApplied++
				}
			}
		}
	}

	// Apply keep-best merges
	if c.cfg.AutoMerge == "keep-best" {
		for _, proposal := range proposals {
			if proposal.Strategy != StrategyKeepBest || proposal.KeepID == "" {
				continue
			}

			// Archive the losers
			for _, removeID := range proposal.RemoveIDs {
				p, ok := patternMap[removeID]
				if !ok {
					continue
				}
				p.Lifecycle.Status = pattern.StatusArchived
				p.Lifecycle.DeprecationReason = fmt.Sprintf("merged: duplicate of %s", proposal.KeepID)
				p.Relations.Supersedes = "" // the kept pattern supersedes this one
				p.Health.Score = 0
				p.Health.LastConsolidated = &now
				if err := c.store.Update(p); err == nil {
					report.ActionsApplied++
				}
			}

			// Update the keeper with relations
			keeper, ok := patternMap[proposal.KeepID]
			if ok {
				keeper.Relations.Related = append(keeper.Relations.Related, proposal.RemoveIDs...)
				keeper.Health.LastConsolidated = &now
				_ = c.store.Update(keeper)
			}
		}
	}

	// Update health metadata on all patterns
	for _, hs := range scores {
		if hs.Action == ActionKeep || hs.Action == ActionUpdate {
			p, ok := patternMap[hs.PatternID]
			if !ok {
				continue
			}
			p.Health.Score = hs.Overall
			p.Health.LastConsolidated = &now
			_ = c.store.Update(p)
		}
	}
}

func (c *Consolidator) getEffectivenessStats() ([]inject.EffectivenessStats, error) {
	if c.injTracker == nil {
		return nil, nil
	}
	return c.injTracker.GetStats()
}

func (c *Consolidator) getAnalyticsStats() ([]analytics.PatternStats, error) {
	if c.analyticsTracker == nil {
		return nil, nil
	}
	return c.analyticsTracker.GetPatternStats()
}

func parseMergeStrategy(s string) MergeStrategy {
	switch s {
	case "keep-best":
		return StrategyKeepBest
	case "llm-merge":
		return StrategyLLMMerge
	case "manual":
		return StrategyManual
	case "off":
		return ""
	default:
		return StrategyKeepBest
	}
}
