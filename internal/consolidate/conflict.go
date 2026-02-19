package consolidate

import (
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// ConflictType represents the type of conflict between patterns.
type ConflictType string

const (
	ConflictContradiction ConflictType = "contradiction" // opposite advice
	ConflictOutdated      ConflictType = "outdated"      // one supersedes the other
	ConflictScope         ConflictType = "scope"          // different contexts that may confuse
)

// Conflict represents a detected conflict between two patterns.
type Conflict struct {
	PatternA *pattern.Pattern `json:"pattern_a"`
	PatternB *pattern.Pattern `json:"pattern_b"`
	Type     ConflictType     `json:"type"`
	Reason   string           `json:"reason"`
}

// ConflictDetector detects conflicts between patterns.
// This is the interface for pluggable detectors (keyword-based, LLM-based, etc.)
type ConflictDetector interface {
	Detect(patterns []*pattern.Pattern) []Conflict
}

// KeywordConflictDetector is a basic conflict detector that uses keyword overlap
// and negation patterns to find potential contradictions.
type KeywordConflictDetector struct{}

// NewKeywordConflictDetector creates a new keyword-based conflict detector.
func NewKeywordConflictDetector() *KeywordConflictDetector {
	return &KeywordConflictDetector{}
}

// negationPairs are keyword pairs that suggest contradictory advice.
var negationPairs = [][2]string{
	{"always", "never"},
	{"enable", "disable"},
	{"use", "avoid"},
	{"prefer", "avoid"},
	{"require", "optional"},
	{"must", "must not"},
	{"should", "should not"},
	{"do", "don't"},
	{"do", "do not"},
	{"allow", "disallow"},
	{"allow", "forbid"},
}

// Detect finds potential conflicts between patterns using keyword analysis.
func (d *KeywordConflictDetector) Detect(patterns []*pattern.Pattern) []Conflict {
	var conflicts []Conflict

	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			a, b := patterns[i], patterns[j]

			// Only compare patterns in the same domain
			if a.GetPrimaryDomain() != b.GetPrimaryDomain() {
				continue
			}

			if c := d.checkContradiction(a, b); c != nil {
				conflicts = append(conflicts, *c)
			}

			if c := d.checkSupersedes(a, b); c != nil {
				conflicts = append(conflicts, *c)
			}
		}
	}

	return conflicts
}

// checkContradiction looks for negation keyword pairs between two patterns.
func (d *KeywordConflictDetector) checkContradiction(a, b *pattern.Pattern) *Conflict {
	aLower := strings.ToLower(a.Content)
	bLower := strings.ToLower(b.Content)

	for _, pair := range negationPairs {
		// Check if A contains one keyword and B contains the negation (or vice versa)
		aHasFirst := strings.Contains(aLower, pair[0])
		aHasSecond := strings.Contains(aLower, pair[1])
		bHasFirst := strings.Contains(bLower, pair[0])
		bHasSecond := strings.Contains(bLower, pair[1])

		if (aHasFirst && bHasSecond && !aHasSecond && !bHasFirst) ||
			(aHasSecond && bHasFirst && !aHasFirst && !bHasSecond) {
			// Check they share at least one keyword from applies conditions
			if shareKeywords(a, b) {
				return &Conflict{
					PatternA: a,
					PatternB: b,
					Type:     ConflictContradiction,
					Reason:   "patterns may give contradictory advice (detected keywords: " + pair[0] + "/" + pair[1] + ")",
				}
			}
		}
	}

	return nil
}

// checkSupersedes looks for patterns where one is likely a newer version of another.
func (d *KeywordConflictDetector) checkSupersedes(a, b *pattern.Pattern) *Conflict {
	// If patterns have the same name prefix and different versions
	aName := strings.ToLower(a.Name)
	bName := strings.ToLower(b.Name)

	// Check if one is explicitly marked as superseding the other
	if a.Relations.Supersedes == b.ID || b.Relations.Supersedes == a.ID {
		older, newer := a, b
		if b.Relations.Supersedes == a.ID {
			older, newer = b, a
		}
		return &Conflict{
			PatternA: older,
			PatternB: newer,
			Type:     ConflictOutdated,
			Reason:   "pattern is superseded by a newer version",
		}
	}

	// Check if names are very similar (one might be a revision)
	if aName == bName && a.ID != b.ID {
		return &Conflict{
			PatternA: a,
			PatternB: b,
			Type:     ConflictOutdated,
			Reason:   "patterns share the same name — one may be outdated",
		}
	}

	return nil
}

// shareKeywords checks if two patterns share any trigger keywords.
func shareKeywords(a, b *pattern.Pattern) bool {
	if len(a.Applies.Keywords) == 0 || len(b.Applies.Keywords) == 0 {
		// If neither has keywords, consider them potentially related by domain alone
		return a.GetPrimaryDomain() == b.GetPrimaryDomain()
	}

	bSet := make(map[string]bool, len(b.Applies.Keywords))
	for _, k := range b.Applies.Keywords {
		bSet[strings.ToLower(k)] = true
	}

	for _, k := range a.Applies.Keywords {
		if bSet[strings.ToLower(k)] {
			return true
		}
	}
	return false
}
