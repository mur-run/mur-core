package pattern

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if store == nil {
		t.Fatal("NewStore returned nil")
	}
	if store.Dir() != dir {
		t.Errorf("store.Dir() = %q, want %q", store.Dir(), dir)
	}
}

func TestStore_Create_Get(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Pattern{
		ID:          "test-pattern",
		Name:        "test-pattern",
		Description: "A test pattern",
		Content:     "Test content here",
		Tags: TagSet{
			Confirmed: []string{"test", "go"},
		},
		SchemaVersion: 2,
	}

	// Create
	if err := store.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, "test-pattern.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Pattern file not created")
	}

	// Get
	got, err := store.Get("test-pattern")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if got.Description != p.Description {
		t.Errorf("Description = %q, want %q", got.Description, p.Description)
	}
	if got.Content != p.Content {
		t.Errorf("Content = %q, want %q", got.Content, p.Content)
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create a few patterns
	patterns := []*Pattern{
		{ID: "p1", Name: "p1", Content: "content1", SchemaVersion: 2},
		{ID: "p2", Name: "p2", Content: "content2", SchemaVersion: 2},
		{ID: "p3", Name: "p3", Content: "content3", SchemaVersion: 2},
	}

	for _, p := range patterns {
		if err := store.Create(p); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List
	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("List returned %d patterns, want 3", len(list))
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Pattern{
		ID:            "delete-me",
		Name:          "delete-me",
		Content:       "content",
		SchemaVersion: 2,
	}

	if err := store.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete
	if err := store.Delete("delete-me"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify gone
	_, err := store.Get("delete-me")
	if err == nil {
		t.Error("Get should fail after Delete")
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Get should fail for nonexistent pattern")
	}
}

func TestStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Pattern{
		ID:            "exists-test",
		Name:          "exists-test",
		Content:       "content",
		SchemaVersion: 2,
	}

	if store.Exists("exists-test") {
		t.Error("Exists should return false before create")
	}

	if err := store.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !store.Exists("exists-test") {
		t.Error("Exists should return true after create")
	}
}

func TestPattern_Lifecycle(t *testing.T) {
	p := &Pattern{
		Lifecycle: LifecycleMeta{
			Status:  StatusActive,
			Created: time.Now(),
		},
	}

	if p.Lifecycle.Status != StatusActive {
		t.Errorf("Status = %q, want %q", p.Lifecycle.Status, StatusActive)
	}
}

func TestTagSet(t *testing.T) {
	tags := TagSet{
		Confirmed: []string{"go", "testing"},
		Inferred: []TagScore{
			{Tag: "best-practices", Confidence: 0.8},
			{Tag: "error-handling", Confidence: 0.5},
		},
		Negative: []string{"python"},
	}

	if len(tags.Confirmed) != 2 {
		t.Errorf("Confirmed count = %d, want 2", len(tags.Confirmed))
	}
	if len(tags.Inferred) != 2 {
		t.Errorf("Inferred count = %d, want 2", len(tags.Inferred))
	}
	if tags.Inferred[0].Tag != "best-practices" {
		t.Errorf("First inferred tag = %q, want best-practices", tags.Inferred[0].Tag)
	}
}

func TestStore_Count(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}

	// Add a pattern
	p := &Pattern{ID: "test", Name: "test", Content: "c", SchemaVersion: 2}
	store.Create(p)

	count, err = store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}
}
