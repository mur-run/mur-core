package learn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-pattern", false},
		{"valid with underscore", "my_pattern_2", false},
		{"empty name", "", true},
		{"invalid chars", "my pattern!", true},
		{"too long", string(make([]byte, 65)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidDomains(t *testing.T) {
	domains := ValidDomains()
	if len(domains) == 0 {
		t.Error("ValidDomains() returned empty list")
	}

	// Check expected domains exist
	expected := map[string]bool{"dev": true, "devops": true, "business": true}
	for _, d := range domains {
		delete(expected, d)
	}
	if len(expected) > 0 {
		t.Errorf("ValidDomains() missing domains: %v", expected)
	}
}

func TestValidCategories(t *testing.T) {
	categories := ValidCategories()
	if len(categories) == 0 {
		t.Error("ValidCategories() returned empty list")
	}

	// Check expected categories exist
	expected := map[string]bool{"pattern": true, "decision": true, "lesson": true}
	for _, c := range categories {
		delete(expected, c)
	}
	if len(expected) > 0 {
		t.Errorf("ValidCategories() missing categories: %v", expected)
	}
}

func TestPatternCRUD(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Test Add
	p := Pattern{
		Name:        "test-pattern",
		Description: "A test pattern",
		Content:     "Test content",
		Domain:      "dev",
		Category:    "pattern",
	}

	if err := Add(p); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test Get
	got, err := Get("test-pattern")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Description != p.Description {
		t.Errorf("Description = %q, want %q", got.Description, p.Description)
	}
	if got.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}

	// Test List
	patterns, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(patterns) != 1 {
		t.Errorf("List() returned %d patterns, want 1", len(patterns))
	}

	// Test Delete
	if err := Delete("test-pattern"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	_, err = Get("test-pattern")
	if err == nil {
		t.Error("Get() after Delete() should error")
	}
}

func TestListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	patterns, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("List() on empty dir = %d, want 0", len(patterns))
	}
}

func TestPatternDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Add pattern without optional fields
	p := Pattern{
		Name:    "minimal",
		Content: "Just content",
	}

	if err := Add(p); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := Get("minimal")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Check defaults applied
	if got.Domain != "general" {
		t.Errorf("Domain = %q, want default 'general'", got.Domain)
	}
	if got.Category != "pattern" {
		t.Errorf("Category = %q, want default 'pattern'", got.Category)
	}
	if got.Confidence != 0.5 {
		t.Errorf("Confidence = %f, want default 0.5", got.Confidence)
	}
}

func TestPatternsDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	dir, err := PatternsDir()
	if err != nil {
		t.Fatalf("PatternsDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".murmur", "patterns")
	if dir != expected {
		t.Errorf("PatternsDir() = %q, want %q", dir, expected)
	}
}
