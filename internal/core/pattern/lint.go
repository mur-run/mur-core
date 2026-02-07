package pattern

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/core/security"
)

// LintSeverity represents the severity of a lint issue.
type LintSeverity string

const (
	SeverityError   LintSeverity = "error"
	SeverityWarning LintSeverity = "warning"
	SeverityInfo    LintSeverity = "info"
)

// LintIssue represents a single lint issue.
type LintIssue struct {
	Pattern  string
	Field    string
	Severity LintSeverity
	Message  string
	Line     int // Optional line number
}

// LintResult holds the result of linting one or more patterns.
type LintResult struct {
	TotalPatterns int
	CleanPatterns int
	Issues        []LintIssue
	ErrorCount    int
	WarningCount  int
	InfoCount     int
}

// IsClean returns true if there are no errors or warnings.
func (r *LintResult) IsClean() bool {
	return r.ErrorCount == 0 && r.WarningCount == 0
}

// HasErrors returns true if there are errors.
func (r *LintResult) HasErrors() bool {
	return r.ErrorCount > 0
}

// Linter provides pattern validation and linting.
type Linter struct {
	sanitizer *security.Sanitizer
	rules     []LintRule
}

// LintRule defines a linting rule.
type LintRule interface {
	Name() string
	Check(p *Pattern) []LintIssue
}

// NewLinter creates a new Linter with default rules.
func NewLinter() *Linter {
	return &Linter{
		sanitizer: security.NewSanitizer(),
		rules:     DefaultLintRules(),
	}
}

// Lint checks a single pattern.
func (l *Linter) Lint(p *Pattern) []LintIssue {
	var issues []LintIssue

	// Run all rules
	for _, rule := range l.rules {
		ruleIssues := rule.Check(p)
		issues = append(issues, ruleIssues...)
	}

	// Security check
	securityIssues := l.checkSecurity(p)
	issues = append(issues, securityIssues...)

	return issues
}

// LintAll checks all patterns in a store.
func (l *Linter) LintAll(store *Store) (*LintResult, error) {
	patterns, err := store.List()
	if err != nil {
		return nil, err
	}

	result := &LintResult{
		TotalPatterns: len(patterns),
	}

	for _, p := range patterns {
		issues := l.Lint(&p)
		if len(issues) == 0 {
			result.CleanPatterns++
			continue
		}

		result.Issues = append(result.Issues, issues...)
		for _, issue := range issues {
			switch issue.Severity {
			case SeverityError:
				result.ErrorCount++
			case SeverityWarning:
				result.WarningCount++
			case SeverityInfo:
				result.InfoCount++
			}
		}
	}

	return result, nil
}

// checkSecurity runs security checks on a pattern.
func (l *Linter) checkSecurity(p *Pattern) []LintIssue {
	var issues []LintIssue

	// Check content for prompt injection
	result := l.sanitizer.Sanitize(p.Content)
	if result.Rejected {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "content",
			Severity: SeverityError,
			Message:  fmt.Sprintf("Security: %s", result.RejectReason),
		})
	}

	for _, warning := range result.Warnings {
		severity := SeverityWarning
		if warning.Risk == "high" {
			severity = SeverityError
		}
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "content",
			Severity: severity,
			Message:  fmt.Sprintf("Security: %s (match: %q)", warning.Description, warning.Match),
		})
	}

	// Verify hash
	if p.Security.Hash != "" && !p.VerifyHash() {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "security.hash",
			Severity: SeverityError,
			Message:  "Content hash mismatch - pattern may have been tampered with",
		})
	}

	return issues
}

// DefaultLintRules returns the default set of lint rules.
func DefaultLintRules() []LintRule {
	return []LintRule{
		&RequiredFieldsRule{},
		&SchemaVersionRule{},
		&ContentLengthRule{MinLength: 10, MaxLength: 50000},
		&TagsRule{},
		&LifecycleRule{},
		&TrustLevelRule{},
	}
}

// RequiredFieldsRule checks for required fields.
type RequiredFieldsRule struct{}

func (r *RequiredFieldsRule) Name() string { return "required-fields" }

func (r *RequiredFieldsRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	if p.Name == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "name",
			Severity: SeverityError,
			Message:  "Pattern name is required",
		})
	}

	if p.Content == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "content",
			Severity: SeverityError,
			Message:  "Pattern content is required",
		})
	}

	if p.ID == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "id",
			Severity: SeverityWarning,
			Message:  "Pattern should have an ID",
		})
	}

	return issues
}

// SchemaVersionRule checks schema version.
type SchemaVersionRule struct{}

func (r *SchemaVersionRule) Name() string { return "schema-version" }

func (r *SchemaVersionRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	if p.SchemaVersion == 0 {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "schema_version",
			Severity: SeverityWarning,
			Message:  "Pattern is missing schema version (run migration)",
		})
	} else if p.SchemaVersion < SchemaVersion {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "schema_version",
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("Pattern uses old schema version %d (current: %d)", p.SchemaVersion, SchemaVersion),
		})
	}

	return issues
}

// ContentLengthRule checks content length.
type ContentLengthRule struct {
	MinLength int
	MaxLength int
}

func (r *ContentLengthRule) Name() string { return "content-length" }

func (r *ContentLengthRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	if len(p.Content) < r.MinLength {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "content",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Content is very short (%d chars, minimum recommended: %d)", len(p.Content), r.MinLength),
		})
	}

	if len(p.Content) > r.MaxLength {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "content",
			Severity: SeverityError,
			Message:  fmt.Sprintf("Content exceeds maximum length (%d chars, max: %d)", len(p.Content), r.MaxLength),
		})
	}

	return issues
}

// TagsRule checks tags configuration.
type TagsRule struct{}

func (r *TagsRule) Name() string { return "tags" }

func (r *TagsRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	// Check if pattern has any tags
	hasTags := len(p.Tags.Confirmed) > 0 || len(p.Tags.Inferred) > 0
	if !hasTags {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "tags",
			Severity: SeverityInfo,
			Message:  "Pattern has no tags (consider adding confirmed tags for better classification)",
		})
	}

	// Check inferred tag confidence
	for _, ts := range p.Tags.Inferred {
		if ts.Confidence < 0 || ts.Confidence > 1 {
			issues = append(issues, LintIssue{
				Pattern:  p.Name,
				Field:    "tags.inferred",
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Tag %q has invalid confidence %f (must be 0.0-1.0)", ts.Tag, ts.Confidence),
			})
		}
	}

	return issues
}

// LifecycleRule checks lifecycle configuration.
type LifecycleRule struct{}

func (r *LifecycleRule) Name() string { return "lifecycle" }

func (r *LifecycleRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	if p.Lifecycle.Status == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "lifecycle.status",
			Severity: SeverityWarning,
			Message:  "Pattern has no lifecycle status",
		})
	}

	if p.Lifecycle.Created.IsZero() {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "lifecycle.created",
			Severity: SeverityWarning,
			Message:  "Pattern has no creation timestamp",
		})
	}

	// Check for deprecated pattern without reason
	if p.Lifecycle.Status == StatusDeprecated && p.Lifecycle.DeprecationReason == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "lifecycle.deprecation_reason",
			Severity: SeverityInfo,
			Message:  "Deprecated pattern should have a deprecation reason",
		})
	}

	return issues
}

// TrustLevelRule checks trust level configuration.
type TrustLevelRule struct{}

func (r *TrustLevelRule) Name() string { return "trust-level" }

func (r *TrustLevelRule) Check(p *Pattern) []LintIssue {
	var issues []LintIssue

	if p.Security.TrustLevel == "" {
		issues = append(issues, LintIssue{
			Pattern:  p.Name,
			Field:    "security.trust_level",
			Severity: SeverityWarning,
			Message:  "Pattern has no trust level (defaulting to untrusted)",
		})
	}

	// Check if low trust pattern has no hash
	if p.Security.TrustLevel == TrustCommunity || p.Security.TrustLevel == TrustUntrusted {
		if p.Security.Hash == "" {
			issues = append(issues, LintIssue{
				Pattern:  p.Name,
				Field:    "security.hash",
				Severity: SeverityWarning,
				Message:  "Low-trust pattern should have a content hash",
			})
		}
		if !p.Security.Reviewed {
			issues = append(issues, LintIssue{
				Pattern:  p.Name,
				Field:    "security.reviewed",
				Severity: SeverityInfo,
				Message:  "Low-trust pattern has not been reviewed",
			})
		}
	}

	return issues
}
