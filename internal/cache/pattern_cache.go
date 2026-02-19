// Package cache provides in-process memory caching for patterns and embeddings.
package cache

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// PatternCache holds all patterns in memory with an inverted tag index.
type PatternCache struct {
	mu       sync.RWMutex
	patterns map[string]*pattern.Pattern // id → pattern
	byName   map[string]string           // name → id (for name-based lookups)
	index    map[string][]string         // tag → pattern IDs
	loadedAt time.Time
}

// NewPatternCache creates an empty PatternCache.
func NewPatternCache() *PatternCache {
	return &PatternCache{
		patterns: make(map[string]*pattern.Pattern),
		byName:   make(map[string]string),
		index:    make(map[string][]string),
	}
}

// Load reads all YAML pattern files from the given directories into memory,
// building the id-map and tag index. The first directory takes precedence
// for duplicate IDs.
func (c *PatternCache) Load(dirs ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset
	c.patterns = make(map[string]*pattern.Pattern)
	c.byName = make(map[string]string)
	c.index = make(map[string][]string)

	for _, dir := range dirs {
		if err := c.loadDir(dir); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	c.loadedAt = time.Now()
	return nil
}

// loadDir reads all .yaml files from a directory into the cache.
// Caller must hold c.mu.
func (c *PatternCache) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var p pattern.Pattern
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}

		// Skip duplicates (primary dir wins)
		if _, exists := c.patterns[p.ID]; exists {
			continue
		}

		c.addPatternLocked(&p)
	}
	return nil
}

// addPatternLocked inserts a single pattern and updates indices.
// Caller must hold c.mu.
func (c *PatternCache) addPatternLocked(p *pattern.Pattern) {
	pCopy := *p
	c.patterns[pCopy.ID] = &pCopy
	c.byName[strings.ToLower(pCopy.Name)] = pCopy.ID

	// Index confirmed tags
	for _, tag := range pCopy.Tags.Confirmed {
		key := strings.ToLower(tag)
		c.index[key] = append(c.index[key], pCopy.ID)
	}

	// Index inferred tags (with any confidence)
	for _, ts := range pCopy.Tags.Inferred {
		key := strings.ToLower(ts.Tag)
		c.index[key] = append(c.index[key], pCopy.ID)
	}
}

// Get returns a pattern by ID. Returns nil if not found.
func (c *PatternCache) Get(id string) *pattern.Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.patterns[id]
}

// GetByName returns a pattern by name (case-insensitive).
func (c *PatternCache) GetByName(name string) *pattern.Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()
	id, ok := c.byName[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return c.patterns[id]
}

// GetByTag returns all patterns that have the given tag (confirmed or inferred).
func (c *PatternCache) GetByTag(tag string) []*pattern.Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := c.index[strings.ToLower(tag)]
	result := make([]*pattern.Pattern, 0, len(ids))
	for _, id := range ids {
		if p, ok := c.patterns[id]; ok {
			result = append(result, p)
		}
	}
	return result
}

// All returns all cached patterns as a slice.
func (c *PatternCache) All() []*pattern.Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*pattern.Pattern, 0, len(c.patterns))
	for _, p := range c.patterns {
		result = append(result, p)
	}
	return result
}

// Active returns only active patterns.
func (c *PatternCache) Active() []*pattern.Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*pattern.Pattern, 0, len(c.patterns))
	for _, p := range c.patterns {
		if p.IsActive() {
			result = append(result, p)
		}
	}
	return result
}

// Len returns the number of cached patterns.
func (c *PatternCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.patterns)
}

// LoadedAt returns when the cache was last loaded.
func (c *PatternCache) LoadedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadedAt
}
