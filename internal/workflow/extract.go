package workflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mur-run/mur-core/internal/session"
)

// ExtractOptions controls how a workflow is extracted from a session analysis.
type ExtractOptions struct {
	SessionID string
	Start     int // start event index (0 = beginning)
	End       int // end event index (0 = all)
}

// ExtractFromAnalysis creates a new Workflow from a session AnalysisResult.
// If opts.Start/End are set, only the steps within that range are included.
func ExtractFromAnalysis(result *session.AnalysisResult, opts ExtractOptions) (*Workflow, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is nil")
	}

	steps := result.Steps

	// Apply start/end range filtering on steps
	if opts.Start > 0 || opts.End > 0 {
		start := opts.Start
		end := opts.End
		if start < 0 {
			start = 0
		}
		if end <= 0 || end > len(steps) {
			end = len(steps)
		}
		if start >= len(steps) {
			return nil, fmt.Errorf("start index %d exceeds step count %d", start, len(steps))
		}
		if start >= end {
			return nil, fmt.Errorf("start index %d >= end index %d", start, end)
		}
		steps = steps[start:end]

		// Re-number steps sequentially
		for i := range steps {
			steps[i].Order = i + 1
		}
	}

	wf := &Workflow{
		ID:          uuid.New().String(),
		Name:        result.Name,
		Description: result.Description,
		Trigger:     result.Trigger,
		Variables:   result.Variables,
		Steps:       steps,
		Tools:       result.Tools,
		Tags:        result.Tags,
	}

	if opts.SessionID != "" {
		wf.SourceSessions = []SourceRef{
			{
				SessionID:  opts.SessionID,
				StartEvent: opts.Start,
				EndEvent:   opts.End,
			},
		}
	}

	return wf, nil
}

// CreateFromSession is a convenience that loads a session analysis and creates a persisted workflow.
func CreateFromSession(sessionID string, start, end int) (*Workflow, error) {
	result, err := session.LoadAnalysis(sessionID)
	if err != nil {
		return nil, fmt.Errorf("load analysis for session %s: %w", sessionID, err)
	}

	wf, err := ExtractFromAnalysis(result, ExtractOptions{
		SessionID: sessionID,
		Start:     start,
		End:       end,
	})
	if err != nil {
		return nil, fmt.Errorf("extract workflow: %w", err)
	}

	// Add creation timestamp to name for uniqueness
	if wf.Name == "" {
		wf.Name = fmt.Sprintf("workflow-%s", time.Now().Format("20060102-150405"))
	}

	if err := Create(wf); err != nil {
		return nil, fmt.Errorf("save workflow: %w", err)
	}

	return wf, nil
}
