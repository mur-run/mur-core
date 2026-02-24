package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mur-run/mur-core/internal/session"
)

// setWorkflowsDir overrides the workflows directory for testing and returns a cleanup function.
func setWorkflowsDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "workflows")
	orig := workflowsDirFunc
	workflowsDirFunc = func() (string, error) { return dir, nil }
	t.Cleanup(func() { workflowsDirFunc = orig })
	return dir
}

func sampleWorkflow(id string) *Workflow {
	return &Workflow{
		ID:          id,
		Name:        "test-workflow",
		Description: "A test workflow",
		Trigger:     "manual trigger",
		Variables: []session.Variable{
			{Name: "host", Type: "string", Required: true, Description: "target host"},
		},
		Steps: []session.Step{
			{Order: 1, Description: "step one", Tool: "shell", OnFailure: "abort"},
			{Order: 2, Description: "step two", Tool: "file-read", OnFailure: "abort"},
		},
		Tools: []string{"shell", "file-read"},
		Tags:  []string{"test", "example"},
	}
}

func TestCreate_And_List(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-create-list")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	entries, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].ID != "wf-create-list" {
		t.Errorf("entry ID = %q, want %q", entries[0].ID, "wf-create-list")
	}
	if entries[0].Name != "test-workflow" {
		t.Errorf("entry Name = %q, want %q", entries[0].Name, "test-workflow")
	}
	if entries[0].Description != "A test workflow" {
		t.Errorf("entry Description = %q, want %q", entries[0].Description, "A test workflow")
	}
}

func TestCreate_DuplicateID(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-dup")
	if err := Create(wf); err != nil {
		t.Fatalf("first Create() error: %v", err)
	}
	if err := Create(wf); err == nil {
		t.Error("expected error creating duplicate workflow")
	}
}

func TestCreate_EmptyID(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("")
	if err := Create(wf); err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestGet(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-get")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, meta, err := Get("wf-get")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.ID != "wf-get" {
		t.Errorf("ID = %q, want %q", got.ID, "wf-get")
	}
	if got.Name != "test-workflow" {
		t.Errorf("Name = %q, want %q", got.Name, "test-workflow")
	}
	if got.Trigger != "manual trigger" {
		t.Errorf("Trigger = %q, want %q", got.Trigger, "manual trigger")
	}
	if len(got.Steps) != 2 {
		t.Errorf("len(Steps) = %d, want 2", len(got.Steps))
	}
	if len(got.Variables) != 1 {
		t.Errorf("len(Variables) = %d, want 1", len(got.Variables))
	}

	// Check metadata
	if meta.ID != "wf-get" {
		t.Errorf("meta.ID = %q, want %q", meta.ID, "wf-get")
	}
	if meta.PublishedVersion != 0 {
		t.Errorf("meta.PublishedVersion = %d, want 0", meta.PublishedVersion)
	}
	if meta.RevisionCount != 1 {
		t.Errorf("meta.RevisionCount = %d, want 1", meta.RevisionCount)
	}
}

func TestGet_NotFound(t *testing.T) {
	setWorkflowsDir(t)

	_, _, err := Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}

func TestUpdate(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-update")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Modify and update
	wf.Name = "updated-workflow"
	wf.Description = "Updated description"
	wf.Steps = append(wf.Steps, session.Step{
		Order: 3, Description: "step three", Tool: "shell", OnFailure: "retry",
	})

	if err := Update(wf); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	// Verify the updated workflow
	got, meta, err := Get("wf-update")
	if err != nil {
		t.Fatalf("Get() after update error: %v", err)
	}
	if got.Name != "updated-workflow" {
		t.Errorf("Name = %q, want %q", got.Name, "updated-workflow")
	}
	if got.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", got.Description, "Updated description")
	}
	if len(got.Steps) != 3 {
		t.Errorf("len(Steps) = %d, want 3", len(got.Steps))
	}

	// Metadata should reflect the update
	if meta.RevisionCount != 2 {
		t.Errorf("RevisionCount = %d, want 2", meta.RevisionCount)
	}
	if meta.Name != "updated-workflow" {
		t.Errorf("meta.Name = %q, want %q", meta.Name, "updated-workflow")
	}

	// Verify revision file was created
	dir, _ := workflowDir("wf-update")
	revPath := filepath.Join(dir, "revisions", "rev-002.yaml")
	if _, err := os.Stat(revPath); os.IsNotExist(err) {
		t.Error("revision rev-002.yaml was not created")
	}

	// Verify index was updated
	entries, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "updated-workflow" {
		t.Errorf("index Name = %q, want %q", entries[0].Name, "updated-workflow")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("nonexistent")
	if err := Update(wf); err == nil {
		t.Error("expected error updating nonexistent workflow")
	}
}

func TestDelete(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-delete")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Confirm it exists
	entries, _ := List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry before delete, got %d", len(entries))
	}

	if err := Delete("wf-delete"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	// Verify directory is removed
	dir, _ := workflowDir("wf-delete")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("workflow directory should not exist after delete")
	}

	// Verify index is updated
	entries, err := List()
	if err != nil {
		t.Fatalf("List() after delete error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("List() returned %d entries after delete, want 0", len(entries))
	}
}

func TestDelete_NotFound(t *testing.T) {
	setWorkflowsDir(t)

	if err := Delete("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent workflow")
	}
}

func TestPublish(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-publish")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// First publish
	ver, err := Publish("wf-publish")
	if err != nil {
		t.Fatalf("Publish() error: %v", err)
	}
	if ver != 1 {
		t.Errorf("first publish version = %d, want 1", ver)
	}

	// Second publish
	ver, err = Publish("wf-publish")
	if err != nil {
		t.Fatalf("second Publish() error: %v", err)
	}
	if ver != 2 {
		t.Errorf("second publish version = %d, want 2", ver)
	}

	// Verify metadata
	_, meta, err := Get("wf-publish")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if meta.PublishedVersion != 2 {
		t.Errorf("meta.PublishedVersion = %d, want 2", meta.PublishedVersion)
	}

	// Verify index reflects published version
	entries, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].PublishedVersion != 2 {
		t.Errorf("index PublishedVersion = %d, want 2", entries[0].PublishedVersion)
	}
}

func TestMultipleWorkflows_ListOrder(t *testing.T) {
	setWorkflowsDir(t)

	// Create two workflows
	wf1 := sampleWorkflow("wf-first")
	wf1.Name = "first-workflow"
	if err := Create(wf1); err != nil {
		t.Fatalf("Create wf1 error: %v", err)
	}

	wf2 := sampleWorkflow("wf-second")
	wf2.Name = "second-workflow"
	if err := Create(wf2); err != nil {
		t.Fatalf("Create wf2 error: %v", err)
	}

	entries, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("List() returned %d entries, want 2", len(entries))
	}

	// Most recently updated should be first (wf-second was created last)
	if entries[0].ID != "wf-second" {
		t.Errorf("first entry ID = %q, want %q (most recent)", entries[0].ID, "wf-second")
	}
}
