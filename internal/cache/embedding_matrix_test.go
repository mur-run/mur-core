package cache

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// writeEmbeddingsJSON writes an embeddings.json file for testing.
func writeEmbeddingsJSON(t *testing.T, dir string, entries []embeddingCacheEntry) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "embeddings.json")
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEmbeddingMatrixLoadAndSearch(t *testing.T) {
	dir := t.TempDir()

	// Create 3 vectors of dimension 4
	// v0 = [1, 0, 0, 0] (unit x)
	// v1 = [0, 1, 0, 0] (unit y)
	// v2 = [0.7, 0.7, 0, 0] (mix of x and y, not normalized)
	entries := []embeddingCacheEntry{
		{ID: "p0", Vector: []float64{1, 0, 0, 0}},
		{ID: "p1", Vector: []float64{0, 1, 0, 0}},
		{ID: "p2", Vector: []float64{0.7, 0.7, 0, 0}},
	}
	cacheFile := writeEmbeddingsJSON(t, dir, entries)

	m := NewEmbeddingMatrix(4)
	if err := m.Load(cacheFile); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if m.Len() != 3 {
		t.Errorf("Len = %d, want 3", m.Len())
	}
	if m.Dim() != 4 {
		t.Errorf("Dim = %d, want 4", m.Dim())
	}

	// Search for [1, 0, 0, 0] - should rank p0 first, then p2, then p1
	results := m.Search([]float64{1, 0, 0, 0}, 3)
	if len(results) != 3 {
		t.Fatalf("Search returned %d results, want 3", len(results))
	}
	if results[0].ID != "p0" {
		t.Errorf("top result = %q, want p0", results[0].ID)
	}
	if results[1].ID != "p2" {
		t.Errorf("second result = %q, want p2", results[1].ID)
	}
	if results[2].ID != "p1" {
		t.Errorf("third result = %q, want p1", results[2].ID)
	}

	// p0 should have score ~1.0 (identical direction)
	if math.Abs(results[0].Score-1.0) > 0.001 {
		t.Errorf("p0 score = %f, want ~1.0", results[0].Score)
	}
	// p1 should have score ~0.0 (orthogonal)
	if math.Abs(results[2].Score) > 0.001 {
		t.Errorf("p1 score = %f, want ~0.0", results[2].Score)
	}
	// p2 should have score = cos(45°) ≈ 0.707
	expected := 1.0 / math.Sqrt(2.0)
	if math.Abs(results[1].Score-expected) > 0.01 {
		t.Errorf("p2 score = %f, want ~%f", results[1].Score, expected)
	}
}

func TestEmbeddingMatrixTopK(t *testing.T) {
	dir := t.TempDir()

	entries := make([]embeddingCacheEntry, 10)
	for i := range entries {
		vec := make([]float64, 4)
		vec[0] = float64(i) // spread them out
		entries[i] = embeddingCacheEntry{
			ID:     "p" + string(rune('0'+i)),
			Vector: vec,
		}
	}
	cacheFile := writeEmbeddingsJSON(t, dir, entries)

	m := NewEmbeddingMatrix(4)
	_ = m.Load(cacheFile)

	results := m.Search([]float64{9, 0, 0, 0}, 3)
	if len(results) != 3 {
		t.Errorf("topK=3 returned %d results", len(results))
	}
}

func TestEmbeddingMatrixPreNormalized(t *testing.T) {
	dir := t.TempDir()

	// Two vectors pointing in the same direction but different magnitudes
	// should get the same score against a query
	entries := []embeddingCacheEntry{
		{ID: "small", Vector: []float64{1, 1, 0}},
		{ID: "large", Vector: []float64{100, 100, 0}},
	}
	cacheFile := writeEmbeddingsJSON(t, dir, entries)

	m := NewEmbeddingMatrix(3)
	_ = m.Load(cacheFile)

	results := m.Search([]float64{1, 1, 0}, 2)
	if len(results) != 2 {
		t.Fatal("expected 2 results")
	}

	// Both should have the same score since they point in the same direction
	if math.Abs(results[0].Score-results[1].Score) > 0.001 {
		t.Errorf("scores differ: %f vs %f (should be equal after normalization)",
			results[0].Score, results[1].Score)
	}
}

func TestEmbeddingMatrixEmptyCache(t *testing.T) {
	m := NewEmbeddingMatrix(768)
	// Load non-existent file
	err := m.Load(filepath.Join(t.TempDir(), "embeddings.json"))
	if err != nil {
		t.Fatalf("Load non-existent should not error: %v", err)
	}
	if m.Len() != 0 {
		t.Errorf("Len = %d, want 0", m.Len())
	}
	if !m.IsLoaded() {
		t.Error("should be marked as loaded")
	}

	results := m.Search([]float64{1, 0}, 5)
	if len(results) != 0 {
		t.Errorf("Search on empty should return 0 results, got %d", len(results))
	}
}

func TestEmbeddingMatrixZeroVector(t *testing.T) {
	dir := t.TempDir()

	entries := []embeddingCacheEntry{
		{ID: "zero", Vector: []float64{0, 0, 0}},
		{ID: "ok", Vector: []float64{1, 0, 0}},
	}
	cacheFile := writeEmbeddingsJSON(t, dir, entries)

	m := NewEmbeddingMatrix(3)
	_ = m.Load(cacheFile)

	// Should not panic on zero vector
	results := m.Search([]float64{1, 0, 0}, 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Zero query vector should return nil
	results = m.Search([]float64{0, 0, 0}, 2)
	if results != nil {
		t.Errorf("zero query should return nil, got %d results", len(results))
	}
}
