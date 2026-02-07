// Package pattern provides pattern lifecycle management.
package pattern

import (
	"fmt"
	"time"
)

// LifecycleManager handles automatic pattern lifecycle transitions.
type LifecycleManager struct {
	store *Store
	cfg   LifecycleConfig
}

// LifecycleConfig holds lifecycle management configuration.
type LifecycleConfig struct {
	// MinUsesForEvaluation: minimum uses before evaluating effectiveness
	MinUsesForEvaluation int `yaml:"min_uses_for_evaluation"`

	// DeprecateThreshold: effectiveness below this triggers deprecation warning
	DeprecateThreshold float64 `yaml:"deprecate_threshold"`

	// ArchiveThreshold: effectiveness below this triggers archival
	ArchiveThreshold float64 `yaml:"archive_threshold"`

	// StaleAfterDays: days without use before considering stale
	StaleAfterDays int `yaml:"stale_after_days"`

	// AutoDeprecateStale: automatically deprecate stale patterns
	AutoDeprecateStale bool `yaml:"auto_deprecate_stale"`

	// DryRun: only report, don't make changes
	DryRun bool `yaml:"dry_run"`
}

// DefaultLifecycleConfig returns sensible defaults.
func DefaultLifecycleConfig() LifecycleConfig {
	return LifecycleConfig{
		MinUsesForEvaluation: 5,
		DeprecateThreshold:   0.3,
		ArchiveThreshold:     0.1,
		StaleAfterDays:       90,
		AutoDeprecateStale:   false,
		DryRun:               false,
	}
}

// NewLifecycleManager creates a new lifecycle manager.
func NewLifecycleManager(store *Store, cfg LifecycleConfig) *LifecycleManager {
	return &LifecycleManager{
		store: store,
		cfg:   cfg,
	}
}

// LifecycleAction represents an action to take on a pattern.
type LifecycleAction struct {
	PatternName string
	PatternID   string
	Action      ActionType
	Reason      string
	OldStatus   LifecycleStatus
	NewStatus   LifecycleStatus
}

// ActionType represents the type of lifecycle action.
type ActionType string

const (
	ActionDeprecate ActionType = "deprecate"
	ActionArchive   ActionType = "archive"
	ActionReactivate ActionType = "reactivate"
	ActionKeep      ActionType = "keep"
)

// LifecycleReport summarizes lifecycle evaluation.
type LifecycleReport struct {
	Evaluated   int
	Deprecated  int
	Archived    int
	Reactivated int
	Skipped     int
	Actions     []LifecycleAction
	Timestamp   time.Time
}

// Evaluate evaluates all patterns and returns recommended actions.
func (m *LifecycleManager) Evaluate() (*LifecycleReport, error) {
	patterns, err := m.store.List()
	if err != nil {
		return nil, err
	}

	report := &LifecycleReport{
		Actions:   make([]LifecycleAction, 0),
		Timestamp: time.Now(),
	}

	now := time.Now()

	for _, p := range patterns {
		report.Evaluated++

		action := m.evaluatePattern(&p, now)
		if action.Action == ActionKeep {
			report.Skipped++
			continue
		}

		report.Actions = append(report.Actions, action)

		switch action.Action {
		case ActionDeprecate:
			report.Deprecated++
		case ActionArchive:
			report.Archived++
		case ActionReactivate:
			report.Reactivated++
		}
	}

	return report, nil
}

// evaluatePattern evaluates a single pattern.
func (m *LifecycleManager) evaluatePattern(p *Pattern, now time.Time) LifecycleAction {
	action := LifecycleAction{
		PatternName: p.Name,
		PatternID:   p.ID,
		OldStatus:   p.Lifecycle.Status,
		Action:      ActionKeep,
	}

	// Check for stale patterns
	if p.Learning.LastUsed != nil {
		daysSinceUse := int(now.Sub(*p.Learning.LastUsed).Hours() / 24)
		if daysSinceUse > m.cfg.StaleAfterDays && m.cfg.AutoDeprecateStale {
			if p.Lifecycle.Status == StatusActive {
				action.Action = ActionDeprecate
				action.NewStatus = StatusDeprecated
				action.Reason = fmt.Sprintf("stale: unused for %d days", daysSinceUse)
				return action
			}
		}
	}

	// Skip if not enough usage data
	if p.Learning.UsageCount < m.cfg.MinUsesForEvaluation {
		return action // Keep, not enough data
	}

	effectiveness := p.Learning.Effectiveness

	// Check for archive threshold (very low effectiveness)
	if effectiveness < m.cfg.ArchiveThreshold {
		if p.Lifecycle.Status != StatusArchived {
			action.Action = ActionArchive
			action.NewStatus = StatusArchived
			action.Reason = fmt.Sprintf("very low effectiveness: %.0f%% (threshold: %.0f%%)",
				effectiveness*100, m.cfg.ArchiveThreshold*100)
			return action
		}
	}

	// Check for deprecation threshold
	if effectiveness < m.cfg.DeprecateThreshold {
		if p.Lifecycle.Status == StatusActive {
			action.Action = ActionDeprecate
			action.NewStatus = StatusDeprecated
			action.Reason = fmt.Sprintf("low effectiveness: %.0f%% (threshold: %.0f%%)",
				effectiveness*100, m.cfg.DeprecateThreshold*100)
			return action
		}
	}

	// Check for reactivation (deprecated but now effective)
	if p.Lifecycle.Status == StatusDeprecated && effectiveness >= m.cfg.DeprecateThreshold*1.5 {
		action.Action = ActionReactivate
		action.NewStatus = StatusActive
		action.Reason = fmt.Sprintf("effectiveness improved: %.0f%%", effectiveness*100)
		return action
	}

	return action
}

// Apply applies the recommended actions.
func (m *LifecycleManager) Apply(report *LifecycleReport) error {
	if m.cfg.DryRun {
		return nil
	}

	for _, action := range report.Actions {
		if action.Action == ActionKeep {
			continue
		}

		p, err := m.store.Get(action.PatternName)
		if err != nil {
			continue
		}

		p.Lifecycle.Status = action.NewStatus
		if action.Action == ActionDeprecate || action.Action == ActionArchive {
			p.Lifecycle.DeprecationReason = action.Reason
		} else if action.Action == ActionReactivate {
			p.Lifecycle.DeprecationReason = ""
		}

		if err := m.store.Update(p); err != nil {
			return fmt.Errorf("failed to update pattern %s: %w", action.PatternName, err)
		}
	}

	return nil
}

// EvaluateAndApply is a convenience method that evaluates and applies.
func (m *LifecycleManager) EvaluateAndApply() (*LifecycleReport, error) {
	report, err := m.Evaluate()
	if err != nil {
		return nil, err
	}

	if err := m.Apply(report); err != nil {
		return report, err
	}

	return report, nil
}

// Deprecate manually deprecates a pattern.
func (m *LifecycleManager) Deprecate(name, reason string) error {
	p, err := m.store.Get(name)
	if err != nil {
		return err
	}

	p.Lifecycle.Status = StatusDeprecated
	p.Lifecycle.DeprecationReason = reason

	return m.store.Update(p)
}

// Archive manually archives a pattern.
func (m *LifecycleManager) Archive(name, reason string) error {
	p, err := m.store.Get(name)
	if err != nil {
		return err
	}

	p.Lifecycle.Status = StatusArchived
	p.Lifecycle.DeprecationReason = reason

	return m.store.Update(p)
}

// Reactivate reactivates a deprecated or archived pattern.
func (m *LifecycleManager) Reactivate(name string) error {
	p, err := m.store.Get(name)
	if err != nil {
		return err
	}

	p.Lifecycle.Status = StatusActive
	p.Lifecycle.DeprecationReason = ""

	return m.store.Update(p)
}

// GetDeprecated returns all deprecated patterns.
func (m *LifecycleManager) GetDeprecated() ([]Pattern, error) {
	patterns, err := m.store.List()
	if err != nil {
		return nil, err
	}

	var deprecated []Pattern
	for _, p := range patterns {
		if p.Lifecycle.Status == StatusDeprecated {
			deprecated = append(deprecated, p)
		}
	}

	return deprecated, nil
}

// GetArchived returns all archived patterns.
func (m *LifecycleManager) GetArchived() ([]Pattern, error) {
	patterns, err := m.store.List()
	if err != nil {
		return nil, err
	}

	var archived []Pattern
	for _, p := range patterns {
		if p.Lifecycle.Status == StatusArchived {
			archived = append(archived, p)
		}
	}

	return archived, nil
}

// Cleanup permanently deletes archived patterns older than the given duration.
func (m *LifecycleManager) Cleanup(olderThan time.Duration) (int, error) {
	if m.cfg.DryRun {
		return 0, nil
	}

	patterns, err := m.store.List()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	deleted := 0

	for _, p := range patterns {
		if p.Lifecycle.Status == StatusArchived && p.Lifecycle.Updated.Before(cutoff) {
			if err := m.store.Delete(p.Name); err != nil {
				continue
			}
			deleted++
		}
	}

	return deleted, nil
}
