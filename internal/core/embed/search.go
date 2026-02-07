// Package embed provides embedding-based semantic search for patterns.
package embed

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// PatternSearcher provides semantic search over patterns.
type PatternSearcher struct {
	store    *pattern.Store
	cache    *Cache
	embedder Embedder
}

// PatternMatch represents a semantically matched pattern.
type PatternMatch struct {
	Pattern    *pattern.Pattern
	Score      float64 // Cosine similarity (0-1)
	Confidence float64 // Combined confidence
}

// NewPatternSearcher creates a new semantic pattern searcher.
func NewPatternSearcher(store *pattern.Store, cfg Config) (*PatternSearcher, error) {
	embedder, err := NewEmbedder(cfg)
	if err != nil {
		return nil, err
	}

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".mur", "embeddings")
	cache := NewCache(cacheDir, embedder)

	// Load existing cache
	_ = cache.Load()

	searcher := &PatternSearcher{
		store:    store,
		cache:    cache,
		embedder: embedder,
	}

	// Auto-index if cache is empty or stale
	patterns, _ := store.List()
	indexed := 0
	for _, p := range patterns {
		if _, ok := cache.Get(p.ID); ok {
			indexed++
		}
	}

	// If less than half patterns are indexed, do a background index
	if len(patterns) > 0 && indexed < len(patterns)/2 {
		go func() {
			_ = searcher.IndexPatterns()
		}()
	}

	return searcher, nil
}

// IndexPatterns indexes all patterns for semantic search.
func (s *PatternSearcher) IndexPatterns() error {
	patterns, err := s.store.List()
	if err != nil {
		return err
	}

	for _, p := range patterns {
		// Create searchable text from pattern
		text := s.patternToText(&p)

		// Embed and cache
		_, err := s.cache.GetOrEmbed(p.ID, text)
		if err != nil {
			// Log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to embed pattern %s: %v\n", p.Name, err)
			continue
		}
	}

	// Save cache
	return s.cache.Save()
}

// Search finds patterns semantically similar to the query.
func (s *PatternSearcher) Search(query string, topK int) ([]PatternMatch, error) {
	// Embed query
	queryVec, err := s.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search cache
	results := s.cache.Search(queryVec, topK*2) // Get extra to filter

	// Convert to PatternMatch
	matches := make([]PatternMatch, 0, topK)
	for _, r := range results {
		// Find pattern by ID
		patterns, _ := s.store.List()
		for _, p := range patterns {
			if p.ID == r.ID {
				pCopy := p
				matches = append(matches, PatternMatch{
					Pattern:    &pCopy,
					Score:      r.Score,
					Confidence: s.computeConfidence(r.Score, &pCopy),
				})
				break
			}
		}
		if len(matches) >= topK {
			break
		}
	}

	return matches, nil
}

// SearchWithContext combines semantic search with context.
func (s *PatternSearcher) SearchWithContext(query string, ctx *SearchContext, topK int) ([]PatternMatch, error) {
	// Build enhanced query
	enhancedQuery := query
	if ctx != nil {
		if ctx.ProjectType != "" {
			enhancedQuery += " " + ctx.ProjectType
		}
		for _, lang := range ctx.Languages {
			enhancedQuery += " " + lang
		}
		for _, fw := range ctx.Frameworks {
			enhancedQuery += " " + fw
		}
	}

	matches, err := s.Search(enhancedQuery, topK)
	if err != nil {
		return nil, err
	}

	// Boost scores based on context match
	if ctx != nil {
		for i := range matches {
			boost := s.contextBoost(matches[i].Pattern, ctx)
			matches[i].Confidence *= boost
		}

		// Re-sort by confidence
		for i := 0; i < len(matches); i++ {
			for j := i + 1; j < len(matches); j++ {
				if matches[j].Confidence > matches[i].Confidence {
					matches[i], matches[j] = matches[j], matches[i]
				}
			}
		}
	}

	return matches, nil
}

// SearchContext provides context for semantic search.
type SearchContext struct {
	ProjectType string
	ProjectName string
	Languages   []string
	Frameworks  []string
	CurrentFile string
}

// patternToText creates searchable text from a pattern.
func (s *PatternSearcher) patternToText(p *pattern.Pattern) string {
	text := p.Name + "\n"
	if p.Description != "" {
		text += p.Description + "\n"
	}
	text += p.Content + "\n"

	// Add tags
	for _, tag := range p.Tags.Confirmed {
		text += tag + " "
	}
	for _, ts := range p.Tags.Inferred {
		text += ts.Tag + " "
	}

	// Add keywords
	for _, kw := range p.Applies.Keywords {
		text += kw + " "
	}

	return text
}

// computeConfidence computes combined confidence score.
func (s *PatternSearcher) computeConfidence(score float64, p *pattern.Pattern) float64 {
	// Base confidence from similarity
	conf := score

	// Boost by effectiveness
	conf *= (0.5 + p.Learning.Effectiveness*0.5)

	// Boost by trust level
	conf *= (0.8 + p.Security.TrustLevel.Score()*0.2)

	// Cap at 1.0
	if conf > 1.0 {
		conf = 1.0
	}

	return conf
}

// contextBoost calculates boost based on context match.
func (s *PatternSearcher) contextBoost(p *pattern.Pattern, ctx *SearchContext) float64 {
	boost := 1.0

	// Language match
	for _, lang := range p.Applies.Languages {
		for _, ctxLang := range ctx.Languages {
			if lang == ctxLang {
				boost += 0.2
			}
		}
	}

	// Framework match
	for _, fw := range p.Applies.Frameworks {
		for _, ctxFw := range ctx.Frameworks {
			if fw == ctxFw {
				boost += 0.2
			}
		}
	}

	// Project match
	for _, proj := range p.Applies.Projects {
		if proj == ctx.ProjectName {
			boost += 0.3
		}
	}

	return boost
}

// Rehash rebuilds the embedding cache for all patterns.
func (s *PatternSearcher) Rehash() error {
	// Clear cache
	s.cache = NewCache(s.cache.dir, s.embedder)

	// Re-index
	return s.IndexPatterns()
}

// Status returns the status of the embedding system.
func (s *PatternSearcher) Status() (EmbedStatus, error) {
	patterns, err := s.store.List()
	if err != nil {
		return EmbedStatus{}, err
	}

	indexed := 0
	for _, p := range patterns {
		if _, ok := s.cache.Get(p.ID); ok {
			indexed++
		}
	}

	return EmbedStatus{
		Provider:      s.embedder.Name(),
		TotalPatterns: len(patterns),
		Indexed:       indexed,
		Dimension:     s.embedder.Dimension(),
	}, nil
}

// EmbedStatus represents embedding system status.
type EmbedStatus struct {
	Provider      string
	TotalPatterns int
	Indexed       int
	Dimension     int
}
