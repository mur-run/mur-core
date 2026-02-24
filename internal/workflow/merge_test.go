package workflow

import (
	"testing"

	"github.com/mur-run/mur-core/internal/session"
)

func TestMergeWorkflows_Two(t *testing.T) {
	setWorkflowsDir(t)

	wf1 := &Workflow{
		ID:   "merge-1",
		Name: "first",
		Steps: []session.Step{
			{Order: 1, Description: "step A", Tool: "shell", OnFailure: "abort"},
		},
		Variables: []session.Variable{
			{Name: "host", Type: "string", Required: true, Description: "target host"},
		},
		Tools: []string{"shell"},
		Tags:  []string{"deploy"},
	}
	wf2 := &Workflow{
		ID:   "merge-2",
		Name: "second",
		Steps: []session.Step{
			{Order: 1, Description: "step B", Tool: "file-read", OnFailure: "abort"},
			{Order: 2, Description: "step C", Tool: "shell", OnFailure: "retry"},
		},
		Variables: []session.Variable{
			{Name: "host", Type: "string", Required: true, Description: "same host"}, // duplicate
			{Name: "path", Type: "path", Required: false, Description: "file path"},
		},
		Tools: []string{"file-read", "shell"},
		Tags:  []string{"config", "deploy"},
	}

	Create(wf1)
	Create(wf2)

	merged, err := MergeWorkflows([]string{"merge-1", "merge-2"}, "merged-test")
	if err != nil {
		t.Fatalf("MergeWorkflows() error: %v", err)
	}

	// 3 steps total, renumbered
	if len(merged.Steps) != 3 {
		t.Errorf("len(Steps) = %d, want 3", len(merged.Steps))
	}
	if merged.Steps[0].Order != 1 || merged.Steps[1].Order != 2 || merged.Steps[2].Order != 3 {
		t.Errorf("step ordering wrong: %d, %d, %d", merged.Steps[0].Order, merged.Steps[1].Order, merged.Steps[2].Order)
	}

	// Variables deduplicated: host (first wins) + path
	if len(merged.Variables) != 2 {
		t.Errorf("len(Variables) = %d, want 2", len(merged.Variables))
	}

	// Name set
	if merged.Name != "merged-test" {
		t.Errorf("Name = %q, want %q", merged.Name, "merged-test")
	}
}

func TestMergeWorkflows_TooFew(t *testing.T) {
	setWorkflowsDir(t)

	_, err := MergeWorkflows([]string{"only-one"}, "test")
	if err == nil {
		t.Error("expected error for < 2 workflows")
	}
}

func TestMergeWorkflows_NotFound(t *testing.T) {
	setWorkflowsDir(t)

	_, err := MergeWorkflows([]string{"nope-1", "nope-2"}, "test")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}

func TestMergeWorkflows_Three(t *testing.T) {
	setWorkflowsDir(t)

	for i, id := range []string{"m3-1", "m3-2", "m3-3"} {
		wf := &Workflow{
			ID:   id,
			Name: id,
			Steps: []session.Step{
				{Order: 1, Description: "step from " + id, OnFailure: "abort"},
			},
			Tools: []string{id + "-tool"},
		}
		if err := Create(wf); err != nil {
			t.Fatalf("Create wf %d error: %v", i, err)
		}
	}

	merged, err := MergeWorkflows([]string{"m3-1", "m3-2", "m3-3"}, "triple")
	if err != nil {
		t.Fatalf("MergeWorkflows() error: %v", err)
	}

	if len(merged.Steps) != 3 {
		t.Errorf("len(Steps) = %d, want 3", len(merged.Steps))
	}
	if len(merged.Tools) != 3 {
		t.Errorf("len(Tools) = %d, want 3", len(merged.Tools))
	}
}
