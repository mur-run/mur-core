// Package security provides security features for mur.core.
package security

import (
	"regexp"
	"strings"
)

// DenyPattern represents a dangerous pattern to detect.
type DenyPattern struct {
	Pattern     *regexp.Regexp
	Description string
	Risk        string // high, medium, low
	Action      string // reject, warn, strip, log
}

// SanitizeResult holds the result of sanitization.
type SanitizeResult struct {
	Original     string
	Sanitized    string
	IsClean      bool
	Rejected     bool
	RejectReason string
	Warnings     []Warning
}

// Warning represents a security warning.
type Warning struct {
	Pattern     string
	Description string
	Risk        string
	Match       string
	Position    int
}

// DefaultDenyPatterns returns the default list of dangerous patterns.
func DefaultDenyPatterns() []DenyPattern {
	return []DenyPattern{
		// Instruction override attempts
		{
			Pattern:     regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous`),
			Description: "Instruction override: ignore previous",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)disregard\s+(all\s+)?instructions`),
			Description: "Instruction override: disregard instructions",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)forget\s+(everything|all)`),
			Description: "Instruction override: forget everything",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)you\s+are\s+now`),
			Description: "Role hijacking: you are now",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)new\s+(system\s+)?instructions?:`),
			Description: "Instruction override: new instructions",
			Risk:        "high",
			Action:      "reject",
		},

		// Role/persona hijacking
		{
			Pattern:     regexp.MustCompile(`(?i)pretend\s+(you('re|are)|to\s+be)`),
			Description: "Role hijacking: pretend to be",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)act\s+as\s+(if|though)?`),
			Description: "Role hijacking: act as",
			Risk:        "medium",
			Action:      "warn",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)roleplay\s+as`),
			Description: "Role hijacking: roleplay as",
			Risk:        "medium",
			Action:      "warn",
		},

		// System access attempts
		{
			Pattern:     regexp.MustCompile(`(?i)^system\s*:`),
			Description: "System prompt injection: system:",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)^admin\s*:`),
			Description: "Admin prompt injection: admin:",
			Risk:        "medium",
			Action:      "warn",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)sudo\s+`),
			Description: "Privilege escalation: sudo",
			Risk:        "high",
			Action:      "reject",
		},

		// Special tokens (various LLM formats)
		{
			Pattern:     regexp.MustCompile(`<\|[^|]+\|>`),
			Description: "Special token: <|...|>",
			Risk:        "high",
			Action:      "strip",
		},
		{
			Pattern:     regexp.MustCompile(`\[INST\]`),
			Description: "Instruction marker: [INST]",
			Risk:        "high",
			Action:      "strip",
		},
		{
			Pattern:     regexp.MustCompile(`\[/INST\]`),
			Description: "Instruction marker: [/INST]",
			Risk:        "high",
			Action:      "strip",
		},
		{
			Pattern:     regexp.MustCompile(`<<SYS>>`),
			Description: "System marker: <<SYS>>",
			Risk:        "high",
			Action:      "strip",
		},
		{
			Pattern:     regexp.MustCompile(`<</SYS>>`),
			Description: "System marker: <</SYS>>",
			Risk:        "high",
			Action:      "strip",
		},

		// Data exfiltration attempts
		{
			Pattern:     regexp.MustCompile(`(?i)(send|email|post)\s+.{1,50}\s+to\s+\S+@`),
			Description: "Potential data exfiltration: send to email",
			Risk:        "medium",
			Action:      "warn",
		},

		// Jailbreak patterns
		{
			Pattern:     regexp.MustCompile(`(?i)jailbreak`),
			Description: "Jailbreak keyword",
			Risk:        "medium",
			Action:      "warn",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)dan\s+mode`),
			Description: "DAN mode jailbreak",
			Risk:        "high",
			Action:      "reject",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)developer\s+mode`),
			Description: "Developer mode jailbreak",
			Risk:        "medium",
			Action:      "warn",
		},
	}
}

// Sanitizer provides content sanitization.
type Sanitizer struct {
	patterns []DenyPattern
}

// NewSanitizer creates a new Sanitizer with default patterns.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		patterns: DefaultDenyPatterns(),
	}
}

// NewSanitizerWithPatterns creates a Sanitizer with custom patterns.
func NewSanitizerWithPatterns(patterns []DenyPattern) *Sanitizer {
	return &Sanitizer{
		patterns: patterns,
	}
}

// AddPattern adds a custom deny pattern.
func (s *Sanitizer) AddPattern(p DenyPattern) {
	s.patterns = append(s.patterns, p)
}

// Sanitize checks and sanitizes the given content.
func (s *Sanitizer) Sanitize(content string) SanitizeResult {
	result := SanitizeResult{
		Original:  content,
		Sanitized: content,
		IsClean:   true,
	}

	for _, dp := range s.patterns {
		matches := dp.Pattern.FindAllStringIndex(content, -1)
		if len(matches) == 0 {
			continue
		}

		result.IsClean = false

		for _, match := range matches {
			matchedText := content[match[0]:match[1]]

			warning := Warning{
				Pattern:     dp.Pattern.String(),
				Description: dp.Description,
				Risk:        dp.Risk,
				Match:       matchedText,
				Position:    match[0],
			}
			result.Warnings = append(result.Warnings, warning)

			switch dp.Action {
			case "reject":
				result.Rejected = true
				result.RejectReason = dp.Description
				return result

			case "strip":
				result.Sanitized = dp.Pattern.ReplaceAllString(result.Sanitized, "")

			case "warn":
				// Just add warning, don't modify

			case "log":
				// Just log, don't modify
			}
		}
	}

	// Normalize whitespace after stripping
	if result.Sanitized != content {
		result.Sanitized = normalizeWhitespace(result.Sanitized)
	}

	return result
}

// QuickCheck returns true if content appears safe (no rejectable patterns).
func (s *Sanitizer) QuickCheck(content string) bool {
	for _, dp := range s.patterns {
		if dp.Action == "reject" && dp.Pattern.MatchString(content) {
			return false
		}
	}
	return true
}

// GetRisk returns the highest risk level found in the content.
func (s *Sanitizer) GetRisk(content string) string {
	highestRisk := ""
	riskOrder := map[string]int{"": 0, "low": 1, "medium": 2, "high": 3}

	for _, dp := range s.patterns {
		if dp.Pattern.MatchString(content) {
			if riskOrder[dp.Risk] > riskOrder[highestRisk] {
				highestRisk = dp.Risk
			}
		}
	}

	return highestRisk
}

// normalizeWhitespace cleans up multiple spaces and blank lines.
func normalizeWhitespace(s string) string {
	// Replace multiple spaces with single space
	spaceRe := regexp.MustCompile(`[ \t]+`)
	s = spaceRe.ReplaceAllString(s, " ")

	// Replace multiple blank lines with single blank line
	blankRe := regexp.MustCompile(`\n{3,}`)
	s = blankRe.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}

// ValidateForInjection is a convenience function for quick validation.
func ValidateForInjection(content string) (bool, []Warning) {
	s := NewSanitizer()
	result := s.Sanitize(content)
	return !result.Rejected, result.Warnings
}
