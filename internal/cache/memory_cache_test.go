package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"gopkg.in/yaml.v3"
)

func TestMemoryCacheDisabled(t *testing.T) {
	mc, err := NewMemoryCache(MemoryCacheOptions{Disabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if mc != nil {
		t.Error("Disabled cache should return nil")
	}
}

func TestMemoryCacheEagerLoad(t *testing.T) {
	pDir := t.TempDir()
	eDir := t.TempDir()

	// Write a pattern
	p := &pattern.Pattern{ID: "id-1", Name: "test-p", Content: "hello"}
	data, _ := yaml.Marshal(p)
	_ = os.WriteFile(filepath.Join(pDir, "test-p.yaml"), data, 0644)

	// Write embedding cache
	entries := []embeddingCacheEntry{
		{ID: "id-1", Vector: []float64{1, 0, 0}},
	}
	eData, _ := json.Marshal(entries)
	_ = os.WriteFile(filepath.Join(eDir, "embeddings.json"), eData, 0644)

	mc, err := NewMemoryCache(MemoryCacheOptions{
		PatternsDirs:   []string{pDir},
		EmbeddingsDir:  eDir,
		EmbeddingDim:   3,
		LazyEmbeddings: false,
	})
	if err != nil {
		t.Fatalf("NewMemoryCache: %v", err)
	}

	if mc.Patterns.Len() != 1 {
		t.Errorf("Patterns.Len = %d, want 1", mc.Patterns.Len())
	}
	if mc.Embeddings.Len() != 1 {
		t.Errorf("Embeddings.Len = %d, want 1", mc.Embeddings.Len())
	}
}

func TestMemoryCacheLazyLoad(t *testing.T) {
	pDir := t.TempDir()
	eDir := t.TempDir()

	// Write a pattern
	p := &pattern.Pattern{ID: "id-1", Name: "test-p"}
	data, _ := yaml.Marshal(p)
	_ = os.WriteFile(filepath.Join(pDir, "test-p.yaml"), data, 0644)

	// Write embedding cache
	entries := []embeddingCacheEntry{
		{ID: "id-1", Vector: []float64{1, 0, 0}},
	}
	eData, _ := json.Marshal(entries)
	_ = os.WriteFile(filepath.Join(eDir, "embeddings.json"), eData, 0644)

	mc, err := NewMemoryCache(MemoryCacheOptions{
		PatternsDirs:   []string{pDir},
		EmbeddingsDir:  eDir,
		EmbeddingDim:   3,
		LazyEmbeddings: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Patterns should be loaded
	if mc.Patterns.Len() != 1 {
		t.Errorf("Patterns loaded eagerly: Len = %d, want 1", mc.Patterns.Len())
	}

	// Embeddings should NOT be loaded yet
	if mc.Embeddings.IsLoaded() {
		t.Error("Embeddings should not be loaded yet (lazy)")
	}

	// Trigger lazy load
	if err := mc.EnsureEmbeddings(); err != nil {
		t.Fatalf("EnsureEmbeddings: %v", err)
	}

	if !mc.Embeddings.IsLoaded() {
		t.Error("Embeddings should be loaded after EnsureEmbeddings")
	}
	if mc.Embeddings.Len() != 1 {
		t.Errorf("Embeddings.Len = %d, want 1", mc.Embeddings.Len())
	}
}
