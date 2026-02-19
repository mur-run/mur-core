package cache

import (
	"os"
	"path/filepath"
	"sync"
)

// MemoryCache is the top-level in-process cache that holds both patterns
// and their embedding matrix. It supports lazy loading of embeddings:
// metadata is loaded immediately, embeddings only on first search call.
type MemoryCache struct {
	Patterns   *PatternCache
	Embeddings *EmbeddingMatrix

	embeddingsFile string

	lazyEmbeddings bool
	embedOnce      sync.Once
	embedErr       error
}

// MemoryCacheOptions configures the memory cache.
type MemoryCacheOptions struct {
	// PatternsDirs lists directories to load patterns from.
	// The first directory takes precedence for duplicate IDs.
	PatternsDirs []string
	// EmbeddingsDir overrides ~/.mur/embeddings.
	EmbeddingsDir string
	// EmbeddingDim is the vector dimensionality (default 768).
	EmbeddingDim int
	// LazyEmbeddings defers loading embeddings until the first search.
	LazyEmbeddings bool
	// Disabled skips caching entirely (--no-cache).
	Disabled bool
}

// DefaultMemoryCacheOptions returns sensible defaults, including both
// the primary patterns dir and the repo patterns dir.
func DefaultMemoryCacheOptions() MemoryCacheOptions {
	home, _ := os.UserHomeDir()
	dirs := []string{filepath.Join(home, ".mur", "patterns")}
	repoDir := filepath.Join(home, ".mur", "repo", "patterns")
	if info, err := os.Stat(repoDir); err == nil && info.IsDir() {
		dirs = append(dirs, repoDir)
	}
	return MemoryCacheOptions{
		PatternsDirs:   dirs,
		EmbeddingsDir:  filepath.Join(home, ".mur", "embeddings"),
		EmbeddingDim:   768,
		LazyEmbeddings: true,
	}
}

// NewMemoryCache creates and loads a MemoryCache.
// If opts.Disabled is true, returns nil (callers should fall back to disk).
func NewMemoryCache(opts MemoryCacheOptions) (*MemoryCache, error) {
	if opts.Disabled {
		return nil, nil
	}

	mc := &MemoryCache{
		Patterns:       NewPatternCache(),
		Embeddings:     NewEmbeddingMatrix(opts.EmbeddingDim),
		embeddingsFile: filepath.Join(opts.EmbeddingsDir, "embeddings.json"),
		lazyEmbeddings: opts.LazyEmbeddings,
	}

	// Always load patterns eagerly
	if err := mc.Patterns.Load(opts.PatternsDirs...); err != nil {
		return nil, err
	}

	// Load embeddings eagerly or defer
	if !opts.LazyEmbeddings {
		if err := mc.Embeddings.Load(mc.embeddingsFile); err != nil {
			return nil, err
		}
	}

	return mc, nil
}

// EnsureEmbeddings triggers lazy loading of embeddings if not yet loaded.
func (mc *MemoryCache) EnsureEmbeddings() error {
	if mc.Embeddings.IsLoaded() {
		return nil
	}
	mc.embedOnce.Do(func() {
		mc.embedErr = mc.Embeddings.Load(mc.embeddingsFile)
	})
	return mc.embedErr
}
