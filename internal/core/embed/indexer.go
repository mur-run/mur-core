// Package embed provides embedding-based semantic search for patterns.
package embed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

// PatternIndexer manages pattern embeddings.
type PatternIndexer struct {
	cfg      *config.Config
	embedder Embedder
	cache    *Cache
	store    *pattern.Store
}

// IndexStatus represents the current index state.
type IndexStatus struct {
	TotalPatterns  int
	IndexedCount   int
	LastUpdated    time.Time
	CacheSize      int64
	EmbeddingModel string
	OllamaRunning  bool
	ModelAvailable bool
}

// NewPatternIndexer creates a new pattern indexer.
func NewPatternIndexer(cfg *config.Config) (*PatternIndexer, error) {
	store, err := pattern.DefaultStore()
	if err != nil {
		return nil, fmt.Errorf("cannot access pattern store: %w", err)
	}

	// Expand cache dir
	cacheDir := cfg.Embeddings.CacheDir
	if strings.HasPrefix(cacheDir, "~") {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, cacheDir[2:])
	}

	// Create embedder based on config
	apiKey := ""
	if cfg.Search.APIKeyEnv != "" {
		apiKey = os.Getenv(cfg.Search.APIKeyEnv)
	}
	embedCfg := Config{
		Provider: cfg.Search.Provider,
		Model:    cfg.Search.Model,
		Endpoint: cfg.Search.OllamaURL,
		APIKey:   apiKey,
	}
	embedder, err := NewEmbedder(embedCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create embedder: %w", err)
	}

	cache := NewCache(cacheDir, embedder)
	_ = cache.Load() // Ignore load errors, start with empty cache

	return &PatternIndexer{
		cfg:      cfg,
		embedder: embedder,
		cache:    cache,
		store:    store,
	}, nil
}

// Status returns the current index status.
func (idx *PatternIndexer) Status() IndexStatus {
	status := IndexStatus{
		EmbeddingModel: idx.cfg.Search.Model,
	}

	// Count patterns
	patterns, err := idx.store.List()
	if err == nil {
		status.TotalPatterns = len(patterns)
	}

	// Count indexed (in cache)
	for _, p := range patterns {
		cacheKey := idx.cacheKey(p)
		if _, ok := idx.cache.Get(cacheKey); ok {
			status.IndexedCount++
		}
	}

	// Check cache file
	cacheFile := idx.cache.cacheFile()
	if info, err := os.Stat(cacheFile); err == nil {
		status.CacheSize = info.Size()
		status.LastUpdated = info.ModTime()
	}

	// Check Ollama availability
	if idx.cfg.Search.Provider == "ollama" {
		status.OllamaRunning = IsOllamaRunning(idx.cfg.Search.OllamaURL)
		if status.OllamaRunning {
			status.ModelAvailable = HasOllamaModel(idx.cfg.Search.OllamaURL, idx.cfg.Search.Model)
		}
	}

	return status
}

// cacheKey returns the cache key for a pattern.
func (idx *PatternIndexer) cacheKey(p pattern.Pattern) string {
	// Use embedding hash if available, otherwise use name
	if p.EmbeddingHash != "" {
		return p.Name + ":" + p.EmbeddingHash
	}
	return p.Name + ":" + p.CalculateEmbeddingHash()
}

// IndexPattern indexes a single pattern.
func (idx *PatternIndexer) IndexPattern(p pattern.Pattern) error {
	return idx.indexPatternWithExpansion(p, nil)
}

// IndexPatternExpanded indexes a pattern with optional expanded queries.
func (idx *PatternIndexer) indexPatternWithExpansion(p pattern.Pattern, eq *ExpandedQueries) error {
	cacheKey := idx.cacheKey(p)

	// Skip if already cached with same hash
	if _, ok := idx.cache.Get(cacheKey); ok {
		return nil
	}

	// Generate rich embedding text: name, tags, keywords, description, content
	text := strings.ToLower(buildIndexText(p))

	// Append expanded queries if available
	if eq != nil {
		if queries := eq.Get(p.Name); len(queries) > 0 {
			text += " | search queries: " + strings.Join(queries, " | ")
		}
	}

	// Embed
	vec, err := idx.embedder.Embed(text)
	if err != nil {
		return fmt.Errorf("failed to embed %s: %w", p.Name, err)
	}

	// Cache
	idx.cache.Set(cacheKey, vec)
	return nil
}

// IndexAll indexes all patterns.
func (idx *PatternIndexer) IndexAll(progress func(current, total int)) error {
	patterns, err := idx.store.List()
	if err != nil {
		return fmt.Errorf("cannot list patterns: %w", err)
	}

	for i, p := range patterns {
		if progress != nil {
			progress(i+1, len(patterns))
		}

		if err := idx.IndexPattern(p); err != nil {
			return err
		}
	}

	// Save cache
	return idx.cache.Save()
}

// Rebuild forces a complete rebuild of the index.
func (idx *PatternIndexer) Rebuild(progress func(current, total int)) error {
	// Clear cache
	idx.cache = NewCache(idx.cache.dir, idx.embedder)

	// Re-index all
	return idx.IndexAll(progress)
}

// RebuildWithExpansion rebuilds the index with LLM-generated query expansion.
func (idx *PatternIndexer) RebuildWithExpansion(ollamaURL, llmModel string, progress func(current, total int, phase string)) error {
	patterns, err := idx.store.List()
	if err != nil {
		return fmt.Errorf("cannot list patterns: %w", err)
	}

	// Phase 1: Generate expanded queries
	eq := LoadExpandedQueries(idx.cache.dir)
	generated := 0
	for i, p := range patterns {
		if progress != nil {
			progress(i+1, len(patterns), "expanding")
		}
		if eq.Get(p.Name) != nil {
			continue // Already expanded
		}
		if err := eq.GenerateForPattern(p, ollamaURL, llmModel); err != nil {
			// Non-fatal: skip this pattern
			continue
		}
		generated++
		// Save periodically
		if generated%10 == 0 {
			_ = eq.Save()
		}
	}
	_ = eq.Save()

	// Phase 2: Rebuild embeddings with expanded text
	idx.cache = NewCache(idx.cache.dir, idx.embedder)
	for i, p := range patterns {
		if progress != nil {
			progress(i+1, len(patterns), "embedding")
		}
		if err := idx.indexPatternWithExpansion(p, eq); err != nil {
			return err
		}
	}

	return idx.cache.Save()
}

// Search searches for patterns similar to the query.
func (idx *PatternIndexer) Search(query string, topK int) ([]PatternMatch, error) {
	// Normalize query to lowercase for case-insensitive matching
	query = strings.ToLower(query)

	// Expand compound words (e.g. "codesigning" → "codesigning code signing")
	query = expandCompoundQuery(query)

	// Apply query prefix if the model requires it (e.g. mxbai-embed-large)
	embQuery := query
	if qp, ok := idx.embedder.(interface{ QueryPrefix() string }); ok {
		if prefix := qp.QueryPrefix(); prefix != "" {
			embQuery = prefix + query
		}
	}

	// Embed query
	queryVec, err := idx.embedder.Embed(embQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search cache - get more results to filter
	results := idx.cache.Search(queryVec, topK*3)

	// Load patterns
	matches := make([]PatternMatch, 0, len(results))
	for _, r := range results {
		// Extract pattern name from cache key
		name := r.ID
		if i := strings.Index(name, ":"); i > 0 {
			name = name[:i]
		}

		p, err := idx.store.Get(name)
		if err != nil {
			continue
		}

		if r.Score >= idx.cfg.Search.MinScore {
			matches = append(matches, PatternMatch{
				Pattern:    p,
				Score:      r.Score,
				Confidence: r.Score, // Use score as confidence for now
			})
		}

		// Stop once we have enough matches
		if len(matches) >= topK {
			break
		}
	}

	return matches, nil
}

// IsOllamaRunning checks if Ollama is running.
func IsOllamaRunning(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// HasOllamaModel checks if a model is available in Ollama.
func HasOllamaModel(baseURL, model string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	for _, m := range result.Models {
		// Model names can be "nomic-embed-text:latest" or just "nomic-embed-text"
		if m.Name == model || strings.HasPrefix(m.Name, model+":") {
			return true
		}
	}

	return false
}

// buildIndexText creates a rich text representation of a pattern for embedding.
// Includes name, tags, keywords, description, and content with structured formatting
// to give the embedding model better semantic signals.
func buildIndexText(p pattern.Pattern) string {
	var parts []string

	// Name (most important, repeated for emphasis)
	parts = append(parts, p.Name)

	// Tags
	var tags []string
	for _, t := range p.Tags.Confirmed {
		tags = append(tags, t)
	}
	for _, ts := range p.Tags.Inferred {
		if ts.Confidence >= 0.7 {
			tags = append(tags, ts.Tag)
		}
	}
	if len(tags) > 0 {
		parts = append(parts, "tags: "+strings.Join(tags, " "))
	}

	// Apply conditions keywords
	if len(p.Applies.Keywords) > 0 {
		parts = append(parts, "keywords: "+strings.Join(p.Applies.Keywords, " "))
	}

	// Languages and frameworks
	if len(p.Applies.Languages) > 0 {
		parts = append(parts, "languages: "+strings.Join(p.Applies.Languages, " "))
	}
	if len(p.Applies.Frameworks) > 0 {
		parts = append(parts, "frameworks: "+strings.Join(p.Applies.Frameworks, " "))
	}

	// Description
	if p.Description != "" {
		parts = append(parts, p.Description)
	}

	// Content (truncate to avoid diluting the embedding with too much text)
	content := p.Content
	if len(content) > 1000 {
		content = content[:1000]
	}
	if content != "" {
		parts = append(parts, content)
	}

	return strings.Join(parts, " | ")
}

// expandCompoundQuery splits camelCase and concatenated words to help embedding models.
// e.g. "codesigning" → "codesigning code signing", "swiftui" → "swiftui swift ui"
func expandCompoundQuery(query string) string {
	words := strings.Fields(query)
	var expanded []string
	for _, word := range words {
		expanded = append(expanded, word)
		// Try common split points for compound words
		if parts := trySplitCompound(word); len(parts) > 1 {
			expanded = append(expanded, strings.Join(parts, " "))
		}
	}
	return strings.Join(expanded, " ")
}

// trySplitCompound tries to split a compound word at common boundaries.
func trySplitCompound(word string) []string {
	if len(word) < 5 {
		return nil
	}

	// Known compound word prefixes in dev context
	prefixes := []string{
		"code", "auto", "web", "dev", "pre", "post", "re",
		"un", "multi", "cross", "over", "under", "sub",
		"super", "meta", "type", "live", "hot",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(word, prefix) && len(word) > len(prefix)+2 {
			rest := word[len(prefix):]
			return []string{prefix, rest}
		}
	}

	return nil
}

// SaveCache saves the embedding cache to disk.
func (idx *PatternIndexer) SaveCache() error {
	return idx.cache.Save()
}

// NewPatternStore is a helper to create a pattern store.
func NewPatternStore() (*pattern.Store, error) {
	return pattern.DefaultStore()
}
