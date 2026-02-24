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
	varOrder := []string{} // track insertion order for deterministic output
	toolSet := make(map[string]bool)
	toolOrder := []string{}
	tagSet := make(map[string]bool)
	tagOrder := []string{}
	var descriptions []string
	var sources []SourceRef
	var firstTrigger string

	stepOrder := 1
	for _, id := range ids {
		wf, _, err := Get(id)
		if err != nil {
			return nil, fmt.Errorf("get workflow %s: %w", id, err)
		}

		// Capture first workflow's trigger
		if firstTrigger == "" && wf.Trigger != "" {
			firstTrigger = wf.Trigger
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

		// Deduplicate variables by name (first wins), deterministic order
		for _, v := range wf.Variables {
			if _, exists := varMap[v.Name]; !exists {
				varMap[v.Name] = v
				varOrder = append(varOrder, v.Name)
			}
		}

		// Union tools, deterministic order
		for _, t := range wf.Tools {
			if !toolSet[t] {
				toolSet[t] = true
				toolOrder = append(toolOrder, t)
			}
		}

		// Union tags, deterministic order
		for _, t := range wf.Tags {
			if !tagSet[t] {
				tagSet[t] = true
				tagOrder = append(tagOrder, t)
			}
		}

		// Track source sessions
		sources = append(sources, wf.SourceSessions...)

		// Also add this workflow as a source reference
		sources = append(sources, SourceRef{SessionID: id})
	}

	merged.Steps = allSteps
	merged.SourceSessions = sources
	merged.Trigger = firstTrigger

	// Convert maps to slices in deterministic order
	for _, name := range varOrder {
		merged.Variables = append(merged.Variables, varMap[name])
	}
	merged.Tools = toolOrder
	merged.Tags = tagOrder

	// Build description
	if len(descriptions) > 0 {
		merged.Description = fmt.Sprintf("Merged from %d workflows", len(ids))
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
