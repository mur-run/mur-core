// Package cache provides local caching for community patterns.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CommunityCache manages cached community patterns.
type CommunityCache struct {
	dir       string
	ttlDays   int
	maxSizeMB int
}

// CachedPattern represents a cached community pattern.
type CachedPattern struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	CachedAt    time.Time `json:"cached_at"`
	LastUsed    time.Time `json:"last_used"`
}

// CacheMeta stores metadata about all cached patterns.
type CacheMeta struct {
	Patterns map[string]*CacheEntry `json:"patterns"`
}

// CacheEntry tracks a single cached pattern.
type CacheEntry struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	CachedAt time.Time `json:"cached_at"`
	LastUsed time.Time `json:"last_used"`
	SizeKB   int64     `json:"size_kb"`
}

// NewCommunityCache creates a new community cache.
func NewCommunityCache(baseDir string, ttlDays, maxSizeMB int) *CommunityCache {
	if ttlDays == 0 {
		ttlDays = 7
	}
	if maxSizeMB == 0 {
		maxSizeMB = 50
	}

	dir := filepath.Join(baseDir, "cache", "community")
	os.MkdirAll(dir, 0755)

	return &CommunityCache{
		dir:       dir,
		ttlDays:   ttlDays,
		maxSizeMB: maxSizeMB,
	}
}

// Get retrieves a pattern from cache. Returns nil if not found or expired.
func (c *CommunityCache) Get(id string) (*CachedPattern, error) {
	path := filepath.Join(c.dir, id+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not in cache
		}
		return nil, err
	}

	var pattern CachedPattern
	if err := json.Unmarshal(data, &pattern); err != nil {
		return nil, err
	}

	// Check TTL
	if time.Since(pattern.LastUsed) > time.Duration(c.ttlDays)*24*time.Hour {
		// Expired, remove it
		os.Remove(path)
		c.updateMeta(id, nil)
		return nil, nil
	}

	// Update last used
	pattern.LastUsed = time.Now()
	c.Save(&pattern)

	return &pattern, nil
}

// Save stores a pattern in the cache.
func (c *CommunityCache) Save(pattern *CachedPattern) error {
	if pattern.CachedAt.IsZero() {
		pattern.CachedAt = time.Now()
	}
	pattern.LastUsed = time.Now()

	data, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.dir, pattern.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Update metadata
	entry := &CacheEntry{
		ID:       pattern.ID,
		Name:     pattern.Name,
		CachedAt: pattern.CachedAt,
		LastUsed: pattern.LastUsed,
		SizeKB:   int64(len(data)) / 1024,
	}
	c.updateMeta(pattern.ID, entry)

	// Check if we need to cleanup
	c.cleanupIfNeeded()

	return nil
}

// Delete removes a pattern from cache.
func (c *CommunityCache) Delete(id string) error {
	path := filepath.Join(c.dir, id+".json")
	os.Remove(path)
	c.updateMeta(id, nil)
	return nil
}

// List returns all cached patterns.
func (c *CommunityCache) List() ([]*CacheEntry, error) {
	meta, err := c.loadMeta()
	if err != nil {
		return nil, err
	}

	var entries []*CacheEntry
	for _, e := range meta.Patterns {
		entries = append(entries, e)
	}

	// Sort by last used (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUsed.After(entries[j].LastUsed)
	})

	return entries, nil
}

// Cleanup removes expired and excess patterns.
func (c *CommunityCache) Cleanup() (int, error) {
	meta, err := c.loadMeta()
	if err != nil {
		meta = &CacheMeta{Patterns: make(map[string]*CacheEntry)}
	}

	removed := 0
	cutoff := time.Now().Add(-time.Duration(c.ttlDays) * 24 * time.Hour)

	// Remove expired patterns
	for id, entry := range meta.Patterns {
		if entry.LastUsed.Before(cutoff) {
			path := filepath.Join(c.dir, id+".json")
			os.Remove(path)
			delete(meta.Patterns, id)
			removed++
		}
	}

	// Check total size
	var totalKB int64
	for _, entry := range meta.Patterns {
		totalKB += entry.SizeKB
	}

	// If over limit, remove LRU patterns
	maxKB := int64(c.maxSizeMB) * 1024
	if totalKB > maxKB {
		// Sort by last used (oldest first)
		var entries []*CacheEntry
		for _, e := range meta.Patterns {
			entries = append(entries, e)
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].LastUsed.Before(entries[j].LastUsed)
		})

		// Remove oldest until under limit
		for _, entry := range entries {
			if totalKB <= maxKB {
				break
			}
			path := filepath.Join(c.dir, entry.ID+".json")
			os.Remove(path)
			delete(meta.Patterns, entry.ID)
			totalKB -= entry.SizeKB
			removed++
		}
	}

	c.saveMeta(meta)
	return removed, nil
}

// Stats returns cache statistics.
func (c *CommunityCache) Stats() (count int, sizeKB int64) {
	meta, err := c.loadMeta()
	if err != nil {
		return 0, 0
	}

	for _, entry := range meta.Patterns {
		count++
		sizeKB += entry.SizeKB
	}
	return
}

// cleanupIfNeeded runs cleanup if cache is over limit.
func (c *CommunityCache) cleanupIfNeeded() {
	count, sizeKB := c.Stats()
	maxKB := int64(c.maxSizeMB) * 1024

	// Cleanup if over 90% of limit or > 1000 patterns
	if sizeKB > maxKB*9/10 || count > 1000 {
		c.Cleanup()
	}
}

func (c *CommunityCache) metaPath() string {
	return filepath.Join(c.dir, ".cache-meta.json")
}

func (c *CommunityCache) loadMeta() (*CacheMeta, error) {
	path := c.metaPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheMeta{Patterns: make(map[string]*CacheEntry)}, nil
		}
		return nil, err
	}

	var meta CacheMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return &CacheMeta{Patterns: make(map[string]*CacheEntry)}, nil
	}

	if meta.Patterns == nil {
		meta.Patterns = make(map[string]*CacheEntry)
	}

	return &meta, nil
}

func (c *CommunityCache) saveMeta(meta *CacheMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.metaPath(), data, 0644)
}

func (c *CommunityCache) updateMeta(id string, entry *CacheEntry) {
	meta, _ := c.loadMeta()
	if entry == nil {
		delete(meta.Patterns, id)
	} else {
		meta.Patterns[id] = entry
	}
	c.saveMeta(meta)
}

// DefaultCommunityCache creates a cache with default settings.
func DefaultCommunityCache() (*CommunityCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	return NewCommunityCache(filepath.Join(home, ".mur"), 7, 50), nil
}
