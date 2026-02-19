package pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Store provides pattern storage operations.
type Store struct {
	baseDir string
}

// NewStore creates a new Store with the given base directory.
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// DefaultStore returns a Store using the default ~/.mur/patterns directory.
func DefaultStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	return NewStore(filepath.Join(home, ".mur", "patterns")), nil
}

// Dir returns the patterns directory path.
func (s *Store) Dir() string {
	return s.baseDir
}

// EnsureDir creates the patterns directory if it doesn't exist.
func (s *Store) EnsureDir() error {
	return os.MkdirAll(s.baseDir, 0755)
}

// patternPath returns the file path for a pattern.
// Checks baseDir and repo/patterns/.
func (s *Store) patternPath(name string) string {
	// First check baseDir (~/.mur/patterns/)
	path := filepath.Join(s.baseDir, name+".yaml")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Check repo patterns (~/.mur/repo/patterns/)
	home, _ := os.UserHomeDir()
	repoPath := filepath.Join(home, ".mur", "repo", "patterns", name+".yaml")
	if _, err := os.Stat(repoPath); err == nil {
		return repoPath
	}

	// Default to baseDir
	return path
}

// validateName checks if a pattern name is valid.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("pattern name cannot be empty")
	}
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("pattern name must contain only letters, numbers, dashes, and underscores")
	}
	if len(name) > 64 {
		return fmt.Errorf("pattern name must be 64 characters or less")
	}
	return nil
}

// List returns all patterns.
func (s *Store) List() ([]Pattern, error) {
	var patterns []Pattern

	// Check for patterns in baseDir (~/.mur/patterns/)
	if _, err := os.Stat(s.baseDir); err == nil {
		patterns = append(patterns, s.listFromDir(s.baseDir)...)
	}

	// Also check repo patterns (~/.mur/repo/patterns/)
	home, _ := os.UserHomeDir()
	repoDir := filepath.Join(home, ".mur", "repo", "patterns")
	if info, err := os.Stat(repoDir); err == nil && info.IsDir() {
		patterns = append(patterns, s.listFromDir(repoDir)...)
	}

	return patterns, nil
}

// listFromDir reads patterns from a specific directory.
func (s *Store) listFromDir(dir string) []Pattern {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var patterns []Pattern
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var p Pattern
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		patterns = append(patterns, p)
	}

	return patterns
}

// LoadVerified returns a pattern with hash integrity verification.
// If the hash doesn't match, the pattern is returned with warnings and TrustLevel set to untrusted.
func (s *Store) LoadVerified(name string) (*Pattern, error) {
	p, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	if p.Security.Hash != "" && !p.VerifyHash() {
		p.Security.Warnings = append(p.Security.Warnings,
			"hash mismatch: content may have been tampered with")
		p.Security.TrustLevel = TrustUntrusted
	}
	return p, nil
}

// Get returns a pattern by name.
func (s *Store) Get(name string) (*Pattern, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	path := s.patternPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pattern not found: %s", name)
		}
		return nil, fmt.Errorf("cannot read pattern: %w", err)
	}

	var p Pattern
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("cannot parse pattern: %w", err)
	}

	return &p, nil
}

// Create creates a new pattern.
func (s *Store) Create(p *Pattern) error {
	if err := validateName(p.Name); err != nil {
		return err
	}

	// Check if already exists
	if _, err := s.Get(p.Name); err == nil {
		return fmt.Errorf("pattern already exists: %s", p.Name)
	}

	if err := s.EnsureDir(); err != nil {
		return fmt.Errorf("cannot create patterns directory: %w", err)
	}

	// Set defaults
	now := time.Now()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Lifecycle.Created.IsZero() {
		p.Lifecycle.Created = now
	}
	p.Lifecycle.Updated = now
	if p.Lifecycle.Status == "" {
		p.Lifecycle.Status = StatusActive
	}
	if p.Security.TrustLevel == "" {
		p.Security.TrustLevel = TrustOwner
	}
	if p.Security.Risk == "" {
		p.Security.Risk = RiskLow
	}
	if p.Learning.Effectiveness == 0 {
		p.Learning.Effectiveness = 0.5
	}
	p.SchemaVersion = SchemaVersion

	// Calculate hash
	p.UpdateHash()

	return s.save(p)
}

// Update updates an existing pattern.
func (s *Store) Update(p *Pattern) error {
	if err := validateName(p.Name); err != nil {
		return err
	}

	// Check if exists
	existing, err := s.Get(p.Name)
	if err != nil {
		return fmt.Errorf("pattern not found: %s", p.Name)
	}

	// Preserve creation time
	p.Lifecycle.Created = existing.Lifecycle.Created
	p.Lifecycle.Updated = time.Now()

	// Recalculate hash if content changed
	if p.Content != existing.Content {
		p.UpdateHash()
	}

	return s.save(p)
}

// Delete removes a pattern.
func (s *Store) Delete(name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	path := s.patternPath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pattern not found: %s", name)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("cannot delete pattern: %w", err)
	}

	return nil
}

// Search returns patterns matching the query.
func (s *Store) Search(query string) ([]Pattern, error) {
	patterns, err := s.List()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []Pattern

	for _, p := range patterns {
		// Search in name, description, content, and tags
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) ||
			strings.Contains(strings.ToLower(p.Content), query) {
			results = append(results, p)
			continue
		}

		// Search in tags
		for _, tag := range p.Tags.Confirmed {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, p)
				break
			}
		}
		for _, ts := range p.Tags.Inferred {
			if strings.Contains(strings.ToLower(ts.Tag), query) {
				results = append(results, p)
				break
			}
		}
	}

	return results, nil
}

// GetByTag returns patterns with the given tag.
func (s *Store) GetByTag(tag string) ([]Pattern, error) {
	patterns, err := s.List()
	if err != nil {
		return nil, err
	}

	tag = strings.ToLower(tag)
	var results []Pattern

	for _, p := range patterns {
		for _, t := range p.Tags.Confirmed {
			if strings.ToLower(t) == tag {
				results = append(results, p)
				break
			}
		}
		for _, ts := range p.Tags.Inferred {
			if strings.ToLower(ts.Tag) == tag {
				results = append(results, p)
				break
			}
		}
	}

	return results, nil
}

// GetActive returns only active patterns.
func (s *Store) GetActive() ([]Pattern, error) {
	patterns, err := s.List()
	if err != nil {
		return nil, err
	}

	var results []Pattern
	for _, p := range patterns {
		if p.IsActive() {
			results = append(results, p)
		}
	}

	return results, nil
}

// RecordUsage records that a pattern was used.
func (s *Store) RecordUsage(name string) error {
	p, err := s.Get(name)
	if err != nil {
		return err
	}

	now := time.Now()
	p.Learning.UsageCount++
	p.Learning.LastUsed = &now

	return s.save(p)
}

// save writes a pattern to disk.
func (s *Store) save(p *Pattern) error {
	path := s.patternPath(p.Name)
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("cannot serialize pattern: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write pattern: %w", err)
	}

	return nil
}

// Exists checks if a pattern exists.
func (s *Store) Exists(name string) bool {
	_, err := s.Get(name)
	return err == nil
}

// Count returns the total number of patterns.
func (s *Store) Count() (int, error) {
	patterns, err := s.List()
	if err != nil {
		return 0, err
	}
	return len(patterns), nil
}
