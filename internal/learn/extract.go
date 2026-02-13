package learn

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ExtractedPattern represents a potential pattern found in a session.
type ExtractedPattern struct {
	Pattern    Pattern  // The pattern to potentially save
	Source     string   // Session ID
	Evidence   []string // Relevant snippets that support this pattern
	Confidence float64  // Extraction confidence
}

// PatternMatcher defines how to detect a pattern type.
type PatternMatcher struct {
	Keywords    []string
	Category    string
	Domain      string
	Description string
}

// PatternMatchers contains keyword patterns to detect.
var PatternMatchers = []PatternMatcher{
	// Best practices
	{
		Keywords:    []string{"best practice", "recommended", "should always", "prefer", "convention"},
		Category:    "pattern",
		Domain:      "dev",
		Description: "Best practice or recommendation",
	},
	// Error handling
	{
		Keywords:    []string{"error handling", "handle error", "catch", "recover", "panic"},
		Category:    "pattern",
		Domain:      "dev",
		Description: "Error handling pattern",
	},
	// Decisions
	{
		Keywords:    []string{"decided to", "chose", "trade-off", "instead of", "because"},
		Category:    "decision",
		Domain:      "dev",
		Description: "Architecture or design decision",
	},
	// Lessons learned
	{
		Keywords:    []string{"learned", "realized", "mistake", "gotcha", "pitfall", "careful", "watch out"},
		Category:    "lesson",
		Domain:      "dev",
		Description: "Lesson learned or gotcha",
	},
	// Templates
	{
		Keywords:    []string{"template", "boilerplate", "scaffold", "starter", "snippet"},
		Category:    "template",
		Domain:      "dev",
		Description: "Reusable template or snippet",
	},
	// DevOps
	{
		Keywords:    []string{"deploy", "ci/cd", "docker", "kubernetes", "infrastructure"},
		Category:    "pattern",
		Domain:      "devops",
		Description: "DevOps or infrastructure pattern",
	},
	// Testing
	{
		Keywords:    []string{"test", "testing", "mock", "fixture", "assert"},
		Category:    "pattern",
		Domain:      "dev",
		Description: "Testing pattern",
	},
}

// ExtractFromSession analyzes a session and extracts patterns.
func ExtractFromSession(sessionPath string) ([]ExtractedPattern, error) {
	session, err := LoadSession(sessionPath)
	if err != nil {
		return nil, err
	}

	return ExtractFromMessages(session.AssistantMessages(), session.ShortID())
}

// JSONPattern represents a pattern in JSON format from Claude's response.
type JSONPattern struct {
	Name         string  `json:"name"`
	Title        string  `json:"title"`
	Confidence   string  `json:"confidence"`
	Score        float64 `json:"score"`
	Category     string  `json:"category"`
	Domain       string  `json:"domain"`
	Project      string  `json:"project"`
	Problem      string  `json:"problem"`
	Solution     string  `json:"solution"`
	Verification string  `json:"verification"`
	WhyNonObvious string `json:"why_non_obvious"`
	Description  string  `json:"description"` // Alternative field
	Content      string  `json:"content"`     // Alternative field
}

// extractJSONPatterns attempts to parse JSON pattern arrays from text.
func extractJSONPatterns(text string, sourceID string) []ExtractedPattern {
	var extracted []ExtractedPattern

	// Find JSON arrays in code blocks or raw text
	jsonArrayRe := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\[.*?\\])\\s*```|^(\\[\\s*\\{.*?\\}\\s*\\])$")
	matches := jsonArrayRe.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		jsonStr := match[1]
		if jsonStr == "" {
			jsonStr = match[2]
		}
		if jsonStr == "" {
			continue
		}

		var jsonPatterns []JSONPattern
		if err := json.Unmarshal([]byte(jsonStr), &jsonPatterns); err != nil {
			continue
		}

		for _, jp := range jsonPatterns {
			// Skip patterns with invalid or empty names
			if jp.Name == "" || !isValidPatternName(jp.Name) {
				continue
			}

			// Build content from JSON fields
			var contentParts []string
			if jp.Title != "" {
				contentParts = append(contentParts, "# "+jp.Title)
			}
			if jp.Problem != "" {
				contentParts = append(contentParts, "\n## Problem\n"+jp.Problem)
			}
			if jp.Solution != "" {
				contentParts = append(contentParts, "\n## Solution\n"+jp.Solution)
			}
			if jp.Verification != "" {
				contentParts = append(contentParts, "\n## Verification\n"+jp.Verification)
			}
			if jp.WhyNonObvious != "" {
				contentParts = append(contentParts, "\n## Why This Is Non-Obvious\n"+jp.WhyNonObvious)
			}

			content := strings.Join(contentParts, "\n")
			if content == "" {
				// Fallback to description or content field
				if jp.Description != "" {
					content = jp.Description
				} else if jp.Content != "" {
					content = jp.Content
				}
			}

			// Skip if no meaningful content
			if content == "" {
				continue
			}

			// Map confidence string to float
			confidence := jp.Score
			if confidence == 0 {
				switch strings.ToUpper(jp.Confidence) {
				case "HIGH":
					confidence = 0.85
				case "MEDIUM":
					confidence = 0.65
				case "LOW":
					confidence = 0.45
				default:
					confidence = 0.5
				}
			}

			// Map domain
			domain := jp.Domain
			if domain == "" || domain == "mobile" {
				domain = "dev"
			}

			// Map category
			category := jp.Category
			if category == "" {
				category = "pattern"
			}

			// Build description from title
			description := jp.Title
			if description == "" && jp.Problem != "" {
				description = truncateText(jp.Problem, 100)
			}

			pattern := Pattern{
				Name:        jp.Name,
				Description: description,
				Content:     content,
				Domain:      domain,
				Category:    category,
				Confidence:  confidence,
				CreatedAt:   time.Now().Format(time.RFC3339),
				UpdatedAt:   time.Now().Format(time.RFC3339),
			}

			extracted = append(extracted, ExtractedPattern{
				Pattern:    pattern,
				Source:     sourceID,
				Evidence:   []string{truncateText(content, 200)},
				Confidence: confidence,
			})
		}
	}

	return extracted
}

// isValidPatternName checks if a pattern name is meaningful.
func isValidPatternName(name string) bool {
	// Skip very short names
	if len(name) < 5 {
		return false
	}

	// Skip names that look like code comments or thinking
	invalidPrefixes := []string{
		"now-", "wait-", "need-", "also-", "then-",
		"first-", "next-", "just-", "see-", "let-",
	}
	lower := strings.ToLower(name)
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}

	// Skip names that are mostly stop words
	invalidPatterns := []string{
		"now-need-update", "wait-changed-line", "now-also-need",
		"see-unloadplist-calls", "install-already-complete",
	}
	for _, invalid := range invalidPatterns {
		if lower == invalid {
			return false
		}
	}

	return true
}

// truncateText shortens text to max length.
func truncateText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// ExtractFromMessages performs extraction from a list of messages.
func ExtractFromMessages(messages []SessionMessage, sourceID string) ([]ExtractedPattern, error) {
	var extracted []ExtractedPattern
	seen := make(map[string]bool) // Dedupe by content hash

	for _, msg := range messages {
		if msg.Content == "" {
			continue
		}

		// First, try to extract JSON patterns
		jsonPatterns := extractJSONPatterns(msg.Content, sourceID)
		for _, ep := range jsonPatterns {
			hash := hashContent(ep.Pattern.Content)
			if !seen[hash] {
				seen[hash] = true
				extracted = append(extracted, ep)
			}
		}

		// If we found JSON patterns, skip keyword-based extraction for this message
		if len(jsonPatterns) > 0 {
			continue
		}

		// Split into paragraphs for analysis
		paragraphs := splitIntoParagraphs(msg.Content)

		for _, para := range paragraphs {
			// Skip very short paragraphs
			if len(para) < 50 {
				continue
			}

			// Try each matcher
			for _, matcher := range PatternMatchers {
				matches, confidence := matchPattern(para, matcher)
				if !matches || confidence < 0.3 {
					continue
				}

				// Generate a unique name
				name := generatePatternName(para, matcher)

				// Skip invalid pattern names
				if !isValidPatternName(name) {
					continue
				}

				// Check for duplicates
				hash := hashContent(para)
				if seen[hash] {
					continue
				}
				seen[hash] = true

				// Extract code blocks if any
				codeBlocks := extractCodeBlocks(para)

				// Build pattern
				pattern := Pattern{
					Name:        name,
					Description: matcher.Description,
					Content:     formatContent(para, codeBlocks),
					Domain:      matcher.Domain,
					Category:    matcher.Category,
					Confidence:  confidence,
					CreatedAt:   time.Now().Format(time.RFC3339),
					UpdatedAt:   time.Now().Format(time.RFC3339),
				}

				extracted = append(extracted, ExtractedPattern{
					Pattern:    pattern,
					Source:     sourceID,
					Evidence:   []string{truncateEvidence(para, 200)},
					Confidence: confidence,
				})
			}
		}
	}

	// Sort by confidence (highest first)
	sortByConfidence(extracted)

	// Limit results
	if len(extracted) > 10 {
		extracted = extracted[:10]
	}

	return extracted, nil
}

// matchPattern checks if text matches a pattern matcher and returns confidence.
func matchPattern(text string, matcher PatternMatcher) (bool, float64) {
	lower := strings.ToLower(text)
	score := 0.0
	matchCount := 0

	// Check keywords
	for _, kw := range matcher.Keywords {
		if strings.Contains(lower, kw) {
			score += 0.2
			matchCount++
		}
	}

	if matchCount == 0 {
		return false, 0
	}

	// Bonus for code blocks
	if strings.Contains(text, "```") {
		score += 0.15
	}

	// Bonus for structured content (bullet points, numbered lists)
	if hasStructuredContent(text) {
		score += 0.1
	}

	// Bonus for longer, detailed content
	if len(text) > 300 {
		score += 0.1
	}

	// Bonus for multiple keyword matches
	if matchCount > 1 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return true, score
}

// generatePatternName creates a pattern name from content.
func generatePatternName(text string, matcher PatternMatcher) string {
	// Extract key terms from the text
	lower := strings.ToLower(text)

	// Find the first matched keyword
	var matchedKeyword string
	for _, kw := range matcher.Keywords {
		if strings.Contains(lower, kw) {
			matchedKeyword = kw
			break
		}
	}

	// Extract significant words near the keyword
	words := extractSignificantWords(text)

	// Build name
	var nameParts []string
	if len(words) > 0 {
		nameParts = append(nameParts, words[:min(3, len(words))]...)
	}
	if matchedKeyword != "" && len(nameParts) == 0 {
		nameParts = append(nameParts, strings.ReplaceAll(matchedKeyword, " ", "-"))
	}

	if len(nameParts) == 0 {
		nameParts = []string{matcher.Category}
	}

	// Clean and join
	name := strings.Join(nameParts, "-")
	name = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(name, "")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	if name == "" {
		name = fmt.Sprintf("%s-%s", matcher.Category, hashContent(text)[:6])
	}

	// Limit length
	if len(name) > 40 {
		name = name[:40]
	}

	return strings.ToLower(name)
}

// extractSignificantWords extracts meaningful words from text.
func extractSignificantWords(text string) []string {
	// Common stop words to skip
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true, "can": true,
		"to": true, "of": true, "in": true, "for": true, "on": true,
		"with": true, "at": true, "by": true, "from": true, "as": true,
		"into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true,
		"under": true, "again": true, "further": true, "then": true,
		"once": true, "here": true, "there": true, "when": true, "where": true,
		"why": true, "how": true, "all": true, "each": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "nor": true, "not": true, "only": true, "own": true,
		"same": true, "so": true, "than": true, "too": true, "very": true,
		"just": true, "and": true, "but": true, "or": true, "because": true,
		"until": true, "while": true, "although": true, "though": true,
		"if": true, "else": true, "this": true, "that": true, "these": true,
		"those": true, "it": true, "its": true, "you": true, "your": true,
		"we": true, "our": true, "they": true, "their": true, "i": true,
		"me": true, "my": true, "he": true, "she": true, "him": true, "her": true,
	}

	// Extract words
	wordRe := regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9]*`)
	matches := wordRe.FindAllString(text, -1)

	var significant []string
	seen := make(map[string]bool)

	for _, word := range matches {
		lower := strings.ToLower(word)
		if len(lower) < 3 || stopWords[lower] || seen[lower] {
			continue
		}
		seen[lower] = true
		significant = append(significant, lower)
		if len(significant) >= 5 {
			break
		}
	}

	return significant
}

// splitIntoParagraphs splits text into meaningful paragraphs.
func splitIntoParagraphs(text string) []string {
	// Split by double newlines or markdown headers
	parts := regexp.MustCompile(`\n\n+|(?m)^#{1,3}\s+`).Split(text, -1)

	var paragraphs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			paragraphs = append(paragraphs, p)
		}
	}

	return paragraphs
}

// extractCodeBlocks extracts code blocks from text.
func extractCodeBlocks(text string) []string {
	re := regexp.MustCompile("```[a-z]*\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(text, -1)

	var blocks []string
	for _, m := range matches {
		if len(m) > 1 {
			blocks = append(blocks, strings.TrimSpace(m[1]))
		}
	}
	return blocks
}

// formatContent formats the pattern content.
func formatContent(text string, codeBlocks []string) string {
	// If we have code blocks, prioritize them
	if len(codeBlocks) > 0 {
		var sb strings.Builder

		// Add explanatory text (first 200 chars without code)
		clean := regexp.MustCompile("```[\\s\\S]*?```").ReplaceAllString(text, "")
		clean = strings.TrimSpace(clean)
		if len(clean) > 200 {
			clean = clean[:200] + "..."
		}
		if clean != "" {
			sb.WriteString(clean)
			sb.WriteString("\n\n")
		}

		// Add code blocks
		for _, block := range codeBlocks {
			sb.WriteString("```\n")
			sb.WriteString(block)
			sb.WriteString("\n```\n")
		}

		return strings.TrimSpace(sb.String())
	}

	// No code blocks, just clean up the text
	if len(text) > 500 {
		text = text[:500] + "..."
	}
	return text
}

// hasStructuredContent checks if text has bullet points or numbered lists.
func hasStructuredContent(text string) bool {
	return regexp.MustCompile(`(?m)^[\s]*[-*•]\s|^[\s]*\d+\.\s`).MatchString(text)
}

// hashContent generates a short hash of content for deduplication.
func hashContent(text string) string {
	h := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", h[:8])
}

// truncateEvidence truncates evidence text for display.
func truncateEvidence(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// sortByConfidence sorts extracted patterns by confidence (descending).
func sortByConfidence(patterns []ExtractedPattern) {
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Confidence > patterns[j].Confidence
	})
}
