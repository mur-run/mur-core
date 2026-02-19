// Package pattern provides the Pattern Schema v2 for mur.core.
package pattern

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the current pattern schema version.
const SchemaVersion = 2

// Pattern represents a learned pattern with Schema v2.
type Pattern struct {
	// Core fields
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Content     string `yaml:"content"`

	// Multi-dimensional tags (replaces fixed domain/category)
	Tags TagSet `yaml:"tags"`

	// Application conditions
	Applies ApplyConditions `yaml:"applies,omitempty"`

	// Security metadata
	Security SecurityMeta `yaml:"security"`

	// Learning metadata
	Learning LearningMeta `yaml:"learning"`

	// Lifecycle metadata
	Lifecycle LifecycleMeta `yaml:"lifecycle"`

	// Schema version for migration
	SchemaVersion int `yaml:"schema_version"`

	// Pattern versioning (semantic version)
	Version string `yaml:"version,omitempty"`

	// L3 resources metadata
	Resources Resources `yaml:"resources,omitempty"`

	// Relations to other patterns
	Relations Relations `yaml:"relations,omitempty"`

	// Consolidation health metadata
	Health HealthMeta `yaml:"health,omitempty"`

	// Embedding hash for semantic search cache (SHA256 of content, first 16 chars)
	EmbeddingHash string `yaml:"embedding_hash,omitempty"`
}

// Relations tracks relationships between patterns.
type Relations struct {
	Supersedes    string   `yaml:"supersedes,omitempty"`
	Related       []string `yaml:"related,omitempty"`
	ConflictsWith []string `yaml:"conflicts_with,omitempty"`
}

// HealthMeta holds consolidation health metadata.
type HealthMeta struct {
	Score            float64    `yaml:"score,omitempty"`
	LastConsolidated *time.Time `yaml:"last_consolidated,omitempty"`
}

// Resources tracks L3 resource availability for a pattern.
type Resources struct {
	HasExamples  bool     `yaml:"has_examples,omitempty"`
	HasReference bool     `yaml:"has_reference,omitempty"`
	Scripts      []string `yaml:"scripts,omitempty"`
}

// TagSet holds multi-dimensional tags for a pattern.
type TagSet struct {
	// Inferred tags from AI classification
	Inferred []TagScore `yaml:"inferred,omitempty"`
	// Confirmed tags by user/admin
	Confirmed []string `yaml:"confirmed,omitempty"`
	// Negative tags (explicitly not applicable)
	Negative []string `yaml:"negative,omitempty"`
}

// TagScore represents a tag with confidence score.
type TagScore struct {
	Tag        string  `yaml:"tag"`
	Confidence float64 `yaml:"confidence"`
}

// ApplyConditions defines when a pattern should be applied.
type ApplyConditions struct {
	// File patterns (glob)
	FilePatterns []string `yaml:"file_patterns,omitempty"`
	// Trigger keywords
	Keywords []string `yaml:"keywords,omitempty"`
	// Sentiment conditions
	Sentiment []string `yaml:"sentiment,omitempty"`
	// Custom context conditions
	Context map[string]interface{} `yaml:"context,omitempty"`
	// Language/framework constraints
	Languages  []string `yaml:"languages,omitempty"`
	Frameworks []string `yaml:"frameworks,omitempty"`
	// Project constraints (glob patterns)
	Projects []string `yaml:"projects,omitempty"`
}

// TrustLevel represents the trust level of a pattern source.
type TrustLevel string

const (
	TrustUntrusted TrustLevel = "untrusted"
	TrustCommunity TrustLevel = "community"
	TrustVerified  TrustLevel = "verified"
	TrustTeam      TrustLevel = "team"
	TrustOwner     TrustLevel = "owner"
)

// TrustScore returns the numeric score for a trust level.
func (t TrustLevel) Score() float64 {
	switch t {
	case TrustOwner:
		return 1.0
	case TrustTeam:
		return 0.75
	case TrustVerified:
		return 0.5
	case TrustCommunity:
		return 0.25
	default:
		return 0.0
	}
}

// RiskLevel represents the risk level of a pattern.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// SecurityMeta holds security-related metadata.
type SecurityMeta struct {
	// Content hash for integrity verification
	Hash string `yaml:"hash"`
	// Source identifier
	Source string `yaml:"source,omitempty"`
	// Trust level
	TrustLevel TrustLevel `yaml:"trust_level"`
	// Review status
	Reviewed bool `yaml:"reviewed"`
	// Reviewer identifier
	Reviewer string `yaml:"reviewer,omitempty"`
	// Review timestamp
	ReviewedAt *time.Time `yaml:"reviewed_at,omitempty"`
	// Risk assessment
	Risk RiskLevel `yaml:"risk"`
	// Security warnings from scanning
	Warnings []string `yaml:"warnings,omitempty"`
}

// LearningMeta holds learning-related metadata.
type LearningMeta struct {
	// Effectiveness score (0.0 - 1.0)
	Effectiveness float64 `yaml:"effectiveness"`
	// Usage count
	UsageCount int `yaml:"usage_count"`
	// Last used timestamp
	LastUsed *time.Time `yaml:"last_used,omitempty"`
	// Source session ID (for extracted patterns)
	ExtractedFrom string `yaml:"extracted_from,omitempty"`
	// Original confidence from extraction
	OriginalConfidence float64 `yaml:"original_confidence,omitempty"`
}

// LifecycleStatus represents the lifecycle status of a pattern.
type LifecycleStatus string

const (
	StatusActive     LifecycleStatus = "active"
	StatusDeprecated LifecycleStatus = "deprecated"
	StatusArchived   LifecycleStatus = "archived"
)

// LifecycleMeta holds lifecycle-related metadata.
type LifecycleMeta struct {
	// Current status
	Status LifecycleStatus `yaml:"status"`
	// Creation timestamp
	Created time.Time `yaml:"created"`
	// Last update timestamp
	Updated time.Time `yaml:"updated"`
	// Deprecation reason (if deprecated)
	DeprecationReason string `yaml:"deprecation_reason,omitempty"`
}

// CalculateHash computes the SHA256 hash of the pattern content.
func (p *Pattern) CalculateHash() string {
	h := sha256.New()
	h.Write([]byte(p.Content))
	return hex.EncodeToString(h.Sum(nil))
}

// CalculateEmbeddingHash computes a short hash for embedding cache invalidation.
// Returns first 16 chars of SHA256 hash.
func (p *Pattern) CalculateEmbeddingHash() string {
	h := sha256.Sum256([]byte(p.Content))
	return hex.EncodeToString(h[:8]) // 8 bytes = 16 hex chars
}

// UpdateEmbeddingHash updates the pattern's embedding hash.
func (p *Pattern) UpdateEmbeddingHash() {
	p.EmbeddingHash = p.CalculateEmbeddingHash()
}

// InferResources determines L3 resource needs based on content.
func (p *Pattern) InferResources() {
	// Has examples if content is long or has code blocks
	p.Resources.HasExamples = len(p.Content) > 500 ||
		strings.Count(p.Content, "```") >= 2
}

// GetPrimaryDomain returns the primary domain from tags.
func (p *Pattern) GetPrimaryDomain() string {
	// Check confirmed tags first
	for _, t := range p.Tags.Confirmed {
		if isDomainTag(t) {
			return strings.ToLower(t)
		}
	}

	// Check high-confidence inferred tags
	for _, ts := range p.Tags.Inferred {
		if ts.Confidence >= 0.7 && isDomainTag(ts.Tag) {
			return strings.ToLower(ts.Tag)
		}
	}

	// Infer from name prefix
	prefixes := []string{"swift-", "go-", "php-", "laravel-", "docker-", "k8s-", "git-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(p.Name), prefix) {
			return strings.TrimSuffix(prefix, "-")
		}
	}

	return "general"
}

// isDomainTag returns true if the tag represents a domain.
func isDomainTag(tag string) bool {
	domains := map[string]bool{
		"swift": true, "go": true, "php": true, "python": true,
		"javascript": true, "typescript": true, "rust": true,
		"devops": true, "docker": true, "kubernetes": true,
		"database": true, "testing": true, "security": true,
	}
	return domains[strings.ToLower(tag)]
}

// UpdateHash updates the pattern's hash.
func (p *Pattern) UpdateHash() {
	p.Security.Hash = p.CalculateHash()
}

// VerifyHash checks if the current hash matches the content.
func (p *Pattern) VerifyHash() bool {
	return p.Security.Hash == p.CalculateHash()
}

// IsActive returns true if the pattern is active.
func (p *Pattern) IsActive() bool {
	// If no status set (old format), treat as active
	if p.Lifecycle.Status == "" {
		return true
	}
	return p.Lifecycle.Status == StatusActive
}

// IsTrusted returns true if the pattern has a trust level >= team.
func (p *Pattern) IsTrusted() bool {
	return p.Security.TrustLevel == TrustOwner || p.Security.TrustLevel == TrustTeam
}

// GetTopTags returns the top N inferred tags by confidence.
func (p *Pattern) GetTopTags(n int) []TagScore {
	if n >= len(p.Tags.Inferred) {
		return p.Tags.Inferred
	}

	// Copy and sort
	sorted := make([]TagScore, len(p.Tags.Inferred))
	copy(sorted, p.Tags.Inferred)

	// Simple bubble sort for small N (typically < 10)
	for i := 0; i < n; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Confidence > sorted[i].Confidence {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:n]
}

// ToYAML serializes the pattern to YAML.
func (p *Pattern) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

// FromYAML deserializes a pattern from YAML.
func FromYAML(data []byte) (*Pattern, error) {
	var p Pattern
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
