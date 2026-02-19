package classifier

import (
	"sort"
	"strings"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// PatternMatch represents a pattern with its relevance score.
type PatternMatch struct {
	Pattern pattern.Pattern
	Score   float64
	Reasons []string
	Domains []string
}

// Retriever finds relevant patterns based on classification.
type Retriever struct {
	store      *pattern.Store
	classifier *HybridClassifier
}

// NewRetriever creates a new Retriever.
func NewRetriever(store *pattern.Store) *Retriever {
	return &Retriever{
		store:      store,
		classifier: NewHybridClassifier(),
	}
}

// Retrieve finds patterns relevant to the input.
func (r *Retriever) Retrieve(input ClassifyInput, limit int) ([]PatternMatch, error) {
	// 1. Classify the input
	domains := r.classifier.Classify(input)

	// 2. Get all active patterns
	patterns, err := r.store.GetActive()
	if err != nil {
		return nil, err
	}

	// 3. Score each pattern
	matches := make([]PatternMatch, 0)
	for _, p := range patterns {
		score, reasons := r.scorePattern(p, input, domains)
		if score > 0.1 { // Minimum threshold
			domainNames := make([]string, 0)
			for _, d := range domains {
				domainNames = append(domainNames, d.Domain)
			}
			matches = append(matches, PatternMatch{
				Pattern: p,
				Score:   score,
				Reasons: reasons,
				Domains: domainNames,
			})
		}
	}

	// 4. Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// 5. Limit results
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}

// scorePattern calculates the relevance score for a pattern.
func (r *Retriever) scorePattern(p pattern.Pattern, input ClassifyInput, domains []DomainScore) (float64, []string) {
	var score float64
	var reasons []string

	// 1. Domain matching
	for _, d := range domains {
		for _, tag := range p.Tags.Inferred {
			if matchesDomain(tag.Tag, d.Domain) {
				domainScore := d.Confidence * tag.Confidence * 0.4
				score += domainScore
				reasons = append(reasons, "domain:"+d.Domain)
			}
		}
		for _, tag := range p.Tags.Confirmed {
			if matchesDomain(tag, d.Domain) {
				score += d.Confidence * 0.5
				reasons = append(reasons, "confirmed-tag:"+tag)
			}
		}
	}

	// 2. Keyword matching in pattern content/description
	contentLower := strings.ToLower(input.Content)
	patternContent := strings.ToLower(p.Content + " " + p.Description + " " + p.Name)

	// Extract significant words from input
	words := extractSignificantWords(contentLower)
	matchedWords := 0
	for _, word := range words {
		if strings.Contains(patternContent, word) {
			matchedWords++
		}
	}
	if len(words) > 0 {
		wordScore := float64(matchedWords) / float64(len(words)) * 0.3
		score += wordScore
		if matchedWords > 0 {
			reasons = append(reasons, "keyword-match")
		}
	}

	// 3. File pattern matching
	if input.CurrentFile != "" {
		for _, fp := range p.Applies.FilePatterns {
			if matchesFilePattern(input.CurrentFile, fp) {
				score += 0.3
				reasons = append(reasons, "file-pattern:"+fp)
			}
		}
		for _, lang := range p.Applies.Languages {
			if matchesLanguage(input.CurrentFile, lang) {
				score += 0.25
				reasons = append(reasons, "language:"+lang)
			}
		}
	}

	// 4. Apply pattern effectiveness as a multiplier
	if p.Learning.Effectiveness > 0 {
		score *= (0.5 + p.Learning.Effectiveness*0.5)
	}

	// 5. Boost recently used patterns slightly
	if p.Learning.LastUsed != nil && p.Learning.UsageCount > 0 {
		score *= 1.1
		reasons = append(reasons, "recently-used")
	}

	// Cap score at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score, reasons
}

// matchesDomain checks if a tag matches a domain.
func matchesDomain(tag, domain string) bool {
	tag = strings.ToLower(tag)
	domain = strings.ToLower(domain)

	if tag == domain {
		return true
	}

	// Partial matches
	if strings.Contains(tag, domain) || strings.Contains(domain, tag) {
		return true
	}

	// Common aliases
	aliases := map[string][]string{
		"swift":      {"ios", "macos", "apple"},
		"go":         {"golang"},
		"javascript": {"js", "node", "nodejs"},
		"typescript": {"ts"},
		"python":     {"py"},
		"testing":    {"test", "tests", "unit-test"},
		"debugging":  {"debug", "error", "fix"},
		"backend":    {"api", "server"},
	}

	if alts, ok := aliases[tag]; ok {
		for _, alt := range alts {
			if alt == domain {
				return true
			}
		}
	}
	if alts, ok := aliases[domain]; ok {
		for _, alt := range alts {
			if alt == tag {
				return true
			}
		}
	}

	return false
}

// extractSignificantWords extracts significant words from content.
func extractSignificantWords(content string) []string {
	// Common stop words to ignore
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true,
		"this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
		"we": true, "they": true, "my": true, "your": true, "his": true,
		"her": true, "its": true, "our": true, "their": true,
		"and": true, "or": true, "but": true, "if": true, "then": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "with": true, "by": true, "from": true, "as": true,
		"what": true, "how": true, "when": true, "where": true, "why": true,
		"all": true, "each": true, "some": true, "any": true, "no": true,
		"not": true, "just": true, "only": true, "very": true,
		"me": true, "help": true, "please": true, "want": true, "need": true,
	}

	// Split and filter
	words := strings.Fields(content)
	significant := make([]string, 0)

	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,!?;:'\"()[]{}/<>")
		word = strings.ToLower(word)

		// Skip short words and stop words
		if len(word) < 3 || stopWords[word] {
			continue
		}

		significant = append(significant, word)
	}

	return significant
}

// matchesFilePattern checks if a file matches a pattern.
func matchesFilePattern(file, pattern string) bool {
	file = strings.ToLower(file)
	pattern = strings.ToLower(pattern)

	if strings.HasPrefix(pattern, "*.") {
		ext := pattern[1:]
		return strings.HasSuffix(file, ext)
	}

	return strings.Contains(file, pattern)
}

// matchesLanguage checks if a file is of a certain language.
func matchesLanguage(file, lang string) bool {
	file = strings.ToLower(file)
	lang = strings.ToLower(lang)

	langExtensions := map[string][]string{
		"swift":      {".swift"},
		"go":         {".go"},
		"rust":       {".rs"},
		"python":     {".py"},
		"javascript": {".js", ".jsx", ".mjs"},
		"typescript": {".ts", ".tsx"},
		"java":       {".java"},
		"kotlin":     {".kt", ".kts"},
		"c":          {".c", ".h"},
		"cpp":        {".cpp", ".hpp", ".cc", ".hh"},
		"ruby":       {".rb"},
		"php":        {".php"},
		"sql":        {".sql"},
	}

	exts, ok := langExtensions[lang]
	if !ok {
		return false
	}

	for _, ext := range exts {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}

	return false
}

// RetrieveByDomain finds patterns for a specific domain.
func (r *Retriever) RetrieveByDomain(domain string, limit int) ([]PatternMatch, error) {
	patterns, err := r.store.GetActive()
	if err != nil {
		return nil, err
	}

	matches := make([]PatternMatch, 0)
	for _, p := range patterns {
		score := 0.0
		var reasons []string

		for _, tag := range p.Tags.Inferred {
			if matchesDomain(tag.Tag, domain) {
				score += tag.Confidence
				reasons = append(reasons, "inferred:"+tag.Tag)
			}
		}
		for _, tag := range p.Tags.Confirmed {
			if matchesDomain(tag, domain) {
				score += 1.0
				reasons = append(reasons, "confirmed:"+tag)
			}
		}

		if score > 0 {
			matches = append(matches, PatternMatch{
				Pattern: p,
				Score:   score,
				Reasons: reasons,
				Domains: []string{domain},
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}
