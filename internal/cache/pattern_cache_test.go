package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"gopkg.in/yaml.v3"
)

// writeTestPattern writes a pattern YAML file to dir.
func writeTestPattern(t *testing.T, dir string, p *pattern.Pattern) {
	t.Helper()
	data, err := yaml.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, p.Name+".yaml"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPatternCacheLoadAndGet(t *testing.T) {
	dir := t.TempDir()

	p1 := &pattern.Pattern{
		ID:   "id-1",
		Name: "go-error-handling",
		Tags: pattern.TagSet{
			Confirmed: []string{"go", "error-handling"},
			Inferred:  []pattern.TagScore{{Tag: "backend", Confidence: 0.8}},
		},
		Content:       "Handle errors in Go",
		SchemaVersion: 2,
	}
	p2 := &pattern.Pattern{
		ID:   "id-2",
		Name: "swift-optionals",
		Tags: pattern.TagSet{
			Confirmed: []string{"swift", "optionals"},
		},
		Content:       "Use optionals in Swift",
		SchemaVersion: 2,
	}
	writeTestPattern(t, dir, p1)
	writeTestPattern(t, dir, p2)

	c := NewPatternCache()
	if err := c.Load(dir); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Len
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}

	// Get by ID
	got := c.Get("id-1")
	if got == nil {
		t.Fatal("Get(id-1) returned nil")
	}
	if got.Name != "go-error-handling" {
		t.Errorf("Get(id-1).Name = %q, want %q", got.Name, "go-error-handling")
	}

	// Get by name
	got = c.GetByName("Swift-Optionals") // case insensitive
	if got == nil {
		t.Fatal("GetByName(Swift-Optionals) returned nil")
	}
	if got.ID != "id-2" {
		t.Errorf("GetByName.ID = %q, want %q", got.ID, "id-2")
	}

	// Get non-existent
	if c.Get("no-such-id") != nil {
		t.Error("Get(no-such-id) should return nil")
	}
}

func TestPatternCacheGetByTag(t *testing.T) {
	dir := t.TempDir()

	p1 := &pattern.Pattern{
		ID:   "id-1",
		Name: "p1",
		Tags: pattern.TagSet{Confirmed: []string{"go", "testing"}},
	}
	p2 := &pattern.Pattern{
		ID:   "id-2",
		Name: "p2",
		Tags: pattern.TagSet{Confirmed: []string{"go", "error-handling"}},
	}
	p3 := &pattern.Pattern{
		ID:   "id-3",
		Name: "p3",
		Tags: pattern.TagSet{
			Inferred: []pattern.TagScore{{Tag: "Go", Confidence: 0.9}},
		},
	}
	writeTestPattern(t, dir, p1)
	writeTestPattern(t, dir, p2)
	writeTestPattern(t, dir, p3)

	c := NewPatternCache()
	if err := c.Load(dir); err != nil {
		t.Fatal(err)
	}

	// "go" should match all three
	goPatterns := c.GetByTag("go")
	if len(goPatterns) != 3 {
		t.Errorf("GetByTag(go) = %d patterns, want 3", len(goPatterns))
	}

	// "testing" should match only p1
	testPatterns := c.GetByTag("testing")
	if len(testPatterns) != 1 {
		t.Errorf("GetByTag(testing) = %d patterns, want 1", len(testPatterns))
	}

	// non-existent tag
	nonePatterns := c.GetByTag("rust")
	if len(nonePatterns) != 0 {
		t.Errorf("GetByTag(rust) = %d patterns, want 0", len(nonePatterns))
	}
}

func TestPatternCacheAll(t *testing.T) {
	dir := t.TempDir()

	for i, name := range []string{"a", "b", "c"} {
		writeTestPattern(t, dir, &pattern.Pattern{
			ID:   name,
			Name: name,
			Lifecycle: pattern.LifecycleMeta{
				Status: pattern.LifecycleStatus([]string{"active", "active", "deprecated"}[i]),
			},
		})
	}

	c := NewPatternCache()
	_ = c.Load(dir)

	all := c.All()
	if len(all) != 3 {
		t.Errorf("All() = %d, want 3", len(all))
	}

	active := c.Active()
	if len(active) != 2 {
		t.Errorf("Active() = %d, want 2", len(active))
	}
}

func TestPatternCacheEmptyDir(t *testing.T) {
	dir := t.TempDir()
	c := NewPatternCache()
	if err := c.Load(dir); err != nil {
		t.Fatalf("Load on empty dir failed: %v", err)
	}
	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
}

func TestPatternCacheNonExistentDir(t *testing.T) {
	c := NewPatternCache()
	// Loading from a non-existent directory should not error (os.IsNotExist is ok)
	err := c.Load("/tmp/this-dir-does-not-exist-" + t.Name())
	if err != nil {
		t.Fatalf("Load on non-existent dir should not error, got: %v", err)
	}
}
