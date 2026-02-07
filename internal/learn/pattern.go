// Package learn provides pattern management for murmur-ai.
package learn

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Pattern represents a learned pattern.
type Pattern struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Content     string  `yaml:"content"`
	Domain      string  `yaml:"domain"`      // dev, devops, business
	Category    string  `yaml:"category"`    // pattern, decision, lesson
	Confidence  float64 `yaml:"confidence"`  // 0.0 - 1.0
	TeamShared  bool    `yaml:"team_shared"` // share to team repo
	CreatedAt   string  `yaml:"created_at"`
	UpdatedAt   string  `yaml:"updated_at"`
}

// ValidDomains returns the list of valid domains.
func ValidDomains() []string {
	return []string{"dev", "devops", "business", "personal", "general"}
}

// ValidCategories returns the list of valid categories.
func ValidCategories() []string {
	return []string{"pattern", "decision", "lesson", "reference", "template"}
}

// PatternsDir returns the path to ~/.mur/patterns/
func PatternsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "patterns"), nil
}

// ensureDir creates the patterns directory if it doesn't exist.
func ensureDir() error {
	dir, err := PatternsDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// patternPath returns the file path for a pattern.
func patternPath(name string) (string, error) {
	dir, err := PatternsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".yaml"), nil
}

// validateName checks if a pattern name is valid.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("pattern name cannot be empty")
	}
	// Allow alphanumeric, dashes, and underscores
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
func List() ([]Pattern, error) {
	var patterns []Pattern

	// Check ~/.mur/patterns/
	dir, err := PatternsDir()
	if err == nil {
		patterns = append(patterns, listFromDir(dir)...)
	}

	// Also check ~/.mur/repo/patterns/
	home, _ := os.UserHomeDir()
	repoDir := filepath.Join(home, ".mur", "repo", "patterns")
	patterns = append(patterns, listFromDir(repoDir)...)

	return patterns, nil
}

// listFromDir reads patterns from a specific directory.
func listFromDir(dir string) []Pattern {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

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

// Get returns a pattern by name.
func Get(name string) (*Pattern, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	path, err := patternPath(name)
	if err != nil {
		return nil, err
	}

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

// Add creates or updates a pattern.
func Add(p Pattern) error {
	if err := validateName(p.Name); err != nil {
		return err
	}

	if err := ensureDir(); err != nil {
		return fmt.Errorf("cannot create patterns directory: %w", err)
	}

	// Set timestamps
	now := time.Now().Format(time.RFC3339)
	if p.CreatedAt == "" {
		// Check if updating existing pattern
		existing, err := Get(p.Name)
		if err == nil {
			p.CreatedAt = existing.CreatedAt
		} else {
			p.CreatedAt = now
		}
	}
	p.UpdatedAt = now

	// Default confidence
	if p.Confidence == 0 {
		p.Confidence = 0.5
	}

	// Default domain and category
	if p.Domain == "" {
		p.Domain = "general"
	}
	if p.Category == "" {
		p.Category = "pattern"
	}

	path, err := patternPath(p.Name)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("cannot serialize pattern: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write pattern: %w", err)
	}

	return nil
}

// Delete removes a pattern.
func Delete(name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	path, err := patternPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pattern not found: %s", name)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("cannot delete pattern: %w", err)
	}

	return nil
}
