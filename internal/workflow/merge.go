package workflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mur-run/mur-core/internal/session"
)

// MergeWorkflows combines multiple workflows into a new one.
// Steps are concatenated in order, variables are deduplicated by name,
// tools and tags are unioned.
func MergeWorkflows(ids []string, name string) (*Workflow, error) {
	if len(ids) < 2 {
		return nil, fmt.Errorf("need at least 2 workflows to merge, got %d", len(ids))
	}

	merged := &Workflow{
		ID:   uuid.New().String(),
		Name: name,
	}

	var allSteps []session.Step
	varMap := make(map[string]session.Variable)
	toolSet := make(map[string]bool)
	tagSet := make(map[string]bool)
	var descriptions []string
	var sources []SourceRef

	stepOrder := 1
	for _, id := range ids {
		wf, _, err := Get(id)
		if err != nil {
			return nil, fmt.Errorf("get workflow %s: %w", id, err)
		}

		// Collect descriptions
		if wf.Description != "" {
			descriptions = append(descriptions, wf.Description)
		}

		// Concatenate steps with renumbered order
		for _, step := range wf.Steps {
			step.Order = stepOrder
			stepOrder++
			allSteps = append(allSteps, step)
		}

		// Deduplicate variables by name (first wins)
		for _, v := range wf.Variables {
			if _, exists := varMap[v.Name]; !exists {
				varMap[v.Name] = v
			}
		}

		// Union tools
		for _, t := range wf.Tools {
			toolSet[t] = true
		}

		// Union tags
		for _, t := range wf.Tags {
			tagSet[t] = true
		}

		// Track source sessions
		sources = append(sources, wf.SourceSessions...)

		// Also add this workflow as a source reference
		sources = append(sources, SourceRef{SessionID: id})
	}

	merged.Steps = allSteps
	merged.SourceSessions = sources

	// Convert maps to slices
	for _, v := range varMap {
		merged.Variables = append(merged.Variables, v)
	}
	for t := range toolSet {
		merged.Tools = append(merged.Tools, t)
	}
	for t := range tagSet {
		merged.Tags = append(merged.Tags, t)
	}

	// Build description
	if len(descriptions) > 0 {
		merged.Description = fmt.Sprintf("Merged from %d workflows", len(ids))
	}

	// Use first workflow's trigger if available
	if first, _, err := Get(ids[0]); err == nil && first.Trigger != "" {
		merged.Trigger = first.Trigger
	}

	// Set name
	if merged.Name == "" {
		merged.Name = fmt.Sprintf("merged-%s", time.Now().Format("20060102-150405"))
	}

	// Save the merged workflow
	if err := Create(merged); err != nil {
		return nil, fmt.Errorf("save merged workflow: %w", err)
	}

	return merged, nil
}
