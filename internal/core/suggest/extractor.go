// Package suggest provides automatic pattern extraction from sessions.
package suggest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// Extractor extracts potential patterns from session transcripts.
type Extractor struct {
	store     *pattern.Store
	outputDir string
	cfg       ExtractorConfig
}

// ExtractorConfig holds configuration for pattern extraction.
type ExtractorConfig struct {
	// Minimum content length for a pattern
	MinContentLength int
	// Minimum occurrences to suggest
	MinOccurrences int
	// Categories to look for
	Categories []string
	// File patterns to analyze
	FilePatterns []string
}

// DefaultExtractorConfig returns sensible defaults.
func DefaultExtractorConfig() ExtractorConfig {
	return ExtractorConfig{
		MinContentLength: 50,
		MinOccurrences:   2,
		Categories: []string{
			"error-handling",
			"testing",
			"refactoring",
			"debugging",
			"architecture",
			"performance",
			"security",
		},
		FilePatterns: []string{
			"*.md",
			"*.txt",
			"session-*.log",
		},
	}
}

// NewExtractor creates a new pattern extractor.
func NewExtractor(store *pattern.Store, outputDir string, cfg ExtractorConfig) *Extractor {
	return &Extractor{
		store:     store,
		outputDir: outputDir,
		cfg:       cfg,
	}
}

// Suggestion represents a suggested pattern.
type Suggestion struct {
	// Suggested name for the pattern
	Name string `json:"name"`
	// Suggested description
	Description string `json:"description"`
	// Extracted content
	Content string `json:"content"`
	// Source file(s)
	Sources []string `json:"sources"`
	// Number of occurrences
	Occurrences int `json:"occurrences"`
	// Confidence score (0-1)
	Confidence float64 `json:"confidence"`
	// Suggested tags
	Tags []string `json:"tags"`
	// Why this was suggested
	Reason string `json:"reason"`
	// Content hash for deduplication
	Hash string `json:"hash"`
}

// ExtractionResult holds the result of pattern extraction.
type ExtractionResult struct {
	Suggestions []Suggestion
	FilesRead   int
	Timestamp   time.Time
}

// Extract analyzes files and extracts potential patterns.
func (e *Extractor) Extract(inputDir string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Suggestions: make([]Suggestion, 0),
		Timestamp:   time.Now(),
	}

	// Find files to analyze
	files, err := e.findFiles(inputDir)
	if err != nil {
		return nil, err
	}
	result.FilesRead = len(files)

	// Extract content blocks from all files
	var allBlocks []contentBlock
	for _, f := range files {
		blocks, err := e.extractBlocks(f)
		if err != nil {
			continue
		}
		allBlocks = append(allBlocks, blocks...)
	}

	// Find recurring patterns
	patterns := e.findRecurringPatterns(allBlocks)

	// Filter and deduplicate
	seen := make(map[string]bool)
	for _, p := range patterns {
		if seen[p.Hash] {
			continue
		}
		seen[p.Hash] = true

		// Skip if similar pattern already exists
		if e.similarPatternExists(p.Content) {
			continue
		}

		result.Suggestions = append(result.Suggestions, p)
	}

	return result, nil
}

// contentBlock represents a block of content from a file.
type contentBlock struct {
	Content  string
	Source   string
	Category string
	Context  string
}

// findFiles finds files matching the configured patterns.
func (e *Extractor) findFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		for _, pattern := range e.cfg.FilePatterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				files = append(files, path)
				break
			}
		}
		return nil
	})

	return files, err
}

// extractBlocks extracts content blocks from a file.
func (e *Extractor) extractBlocks(path string) ([]contentBlock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	var blocks []contentBlock

	// Extract code blocks
	codeBlockRe := regexp.MustCompile("```[a-z]*\\n([\\s\\S]*?)```")
	matches := codeBlockRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 && len(m[1]) >= e.cfg.MinContentLength {
			blocks = append(blocks, contentBlock{
				Content:  strings.TrimSpace(m[1]),
				Source:   path,
				Category: detectCategory(m[1]),
			})
		}
	}

	// Extract list patterns (e.g., numbered steps, bullet points)
	listRe := regexp.MustCompile(`(?m)^(\d+\.|[-*])\s+.+(\n(\d+\.|[-*])\s+.+){2,}`)
	listMatches := listRe.FindAllString(content, -1)
	for _, m := range listMatches {
		if len(m) >= e.cfg.MinContentLength {
			blocks = append(blocks, contentBlock{
				Content:  strings.TrimSpace(m),
				Source:   path,
				Category: "workflow",
			})
		}
	}

	// Extract sections with headers (simplified regex without lookahead)
	sectionRe := regexp.MustCompile(`(?m)^(#{1,3})\s+(.+)\n`)
	sectionMatches := sectionRe.FindAllStringSubmatchIndex(content, -1)

	for i := 0; i < len(sectionMatches); i++ {
		headerStart := sectionMatches[i][0]
		headerEnd := sectionMatches[i][1]
		titleStart := sectionMatches[i][4]
		titleEnd := sectionMatches[i][5]

		// Find content end (next header or EOF)
		contentEnd := len(content)
		if i+1 < len(sectionMatches) {
			contentEnd = sectionMatches[i+1][0]
		}

		title := content[titleStart:titleEnd]
		sectionContent := content[headerEnd:contentEnd]

		if len(sectionContent) >= e.cfg.MinContentLength {
			blocks = append(blocks, contentBlock{
				Content:  strings.TrimSpace(sectionContent),
				Source:   path,
				Category: detectCategory(title + " " + sectionContent),
				Context:  title,
			})
		}
		_ = headerStart // silence unused warning
	}

	return blocks, nil
}

// findRecurringPatterns finds patterns that appear multiple times.
func (e *Extractor) findRecurringPatterns(blocks []contentBlock) []Suggestion {
	// Group similar blocks
	groups := make(map[string][]contentBlock)

	for _, b := range blocks {
		// Create a simplified key for grouping
		key := simplifyContent(b.Content)
		groups[key] = append(groups[key], b)
	}

	var suggestions []Suggestion

	for _, group := range groups {
		if len(group) < e.cfg.MinOccurrences {
			continue
		}

		// Use the first block as representative
		rep := group[0]

		// Collect sources
		sources := make([]string, 0, len(group))
		for _, b := range group {
			sources = append(sources, b.Source)
		}

		// Generate suggestion
		suggestion := Suggestion{
			Name:        generateName(rep.Category, rep.Context),
			Description: generateDescription(rep.Content),
			Content:     rep.Content,
			Sources:     unique(sources),
			Occurrences: len(group),
			Confidence:  calculateConfidence(group),
			Tags:        detectTags(rep.Content),
			Reason:      fmt.Sprintf("Found %d occurrences across %d files", len(group), len(unique(sources))),
			Hash:        hashContent(rep.Content),
		}

		suggestions = append(suggestions, suggestion)
	}

	// Sort by confidence
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Confidence > suggestions[i].Confidence {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions
}

// similarPatternExists checks if a similar pattern already exists.
func (e *Extractor) similarPatternExists(content string) bool {
	patterns, err := e.store.List()
	if err != nil {
		return false
	}

	contentLower := strings.ToLower(content)
	contentWords := extractWords(contentLower)

	for _, p := range patterns {
		patternWords := extractWords(strings.ToLower(p.Content))

		// Check word overlap
		overlap := wordOverlap(contentWords, patternWords)
		if overlap > 0.7 {
			return true
		}
	}

	return false
}

// Accept converts a suggestion into a pattern.
func (e *Extractor) Accept(s Suggestion) (*pattern.Pattern, error) {
	p := &pattern.Pattern{
		Name:        s.Name,
		Description: s.Description,
		Content:     s.Content,
		Tags: pattern.TagSet{
			Inferred: make([]pattern.TagScore, 0, len(s.Tags)),
		},
		Security: pattern.SecurityMeta{
			TrustLevel: pattern.TrustOwner,
			Risk:       pattern.RiskLow,
			Source:     "extracted",
		},
		Learning: pattern.LearningMeta{
			Effectiveness:      0.5, // Start at neutral
			OriginalConfidence: s.Confidence,
			ExtractedFrom:      strings.Join(s.Sources, ", "),
		},
		Lifecycle: pattern.LifecycleMeta{
			Status: pattern.StatusActive,
		},
	}

	for _, tag := range s.Tags {
		p.Tags.Inferred = append(p.Tags.Inferred, pattern.TagScore{
			Tag:        tag,
			Confidence: s.Confidence,
		})
	}

	if err := e.store.Create(p); err != nil {
		return nil, err
	}

	return p, nil
}

// ============================================================
// Helper Functions
// ============================================================

func detectCategory(content string) string {
	contentLower := strings.ToLower(content)

	categories := map[string][]string{
		"error-handling": {"error", "exception", "catch", "try", "throw", "handle"},
		"testing":        {"test", "assert", "expect", "mock", "stub", "verify"},
		"debugging":      {"debug", "log", "print", "trace", "breakpoint"},
		"refactoring":    {"refactor", "clean", "extract", "rename", "move"},
		"performance":    {"performance", "optimize", "cache", "fast", "slow"},
		"security":       {"security", "auth", "encrypt", "validate", "sanitize"},
		"architecture":   {"architecture", "design", "pattern", "structure", "layer"},
	}

	for cat, keywords := range categories {
		for _, kw := range keywords {
			if strings.Contains(contentLower, kw) {
				return cat
			}
		}
	}

	return "general"
}

func detectTags(content string) []string {
	var tags []string

	contentLower := strings.ToLower(content)

	// Detect languages
	langPatterns := map[string][]string{
		"swift":      {"swift", "swiftui", "uikit", "@State", "func "},
		"go":         {"golang", "go ", "func (", "package "},
		"typescript": {"typescript", "interface ", ": string", ": number"},
		"python":     {"python", "def ", "import ", "__init__"},
		"rust":       {"rust", "fn ", "impl ", "mut "},
	}

	for lang, patterns := range langPatterns {
		for _, p := range patterns {
			if strings.Contains(contentLower, p) {
				tags = append(tags, lang)
				break
			}
		}
	}

	// Detect categories
	tags = append(tags, detectCategory(content))

	return unique(tags)
}

func generateName(category, context string) string {
	if context != "" {
		// Clean context for use as name
		name := strings.ToLower(context)
		name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "-")
		name = strings.Trim(name, "-")
		if len(name) > 30 {
			name = name[:30]
		}
		return name
	}

	return category + "-pattern-" + time.Now().Format("20060102")
}

func generateDescription(content string) string {
	// Take first line or first 100 chars
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		first := strings.TrimSpace(lines[0])
		if len(first) > 100 {
			return first[:97] + "..."
		}
		if first != "" {
			return first
		}
	}

	if len(content) > 100 {
		return content[:97] + "..."
	}
	return content
}

func simplifyContent(content string) string {
	// Remove whitespace variations
	simplified := strings.TrimSpace(content)
	simplified = regexp.MustCompile(`\s+`).ReplaceAllString(simplified, " ")

	// Take first 200 chars for grouping
	if len(simplified) > 200 {
		simplified = simplified[:200]
	}

	return strings.ToLower(simplified)
}

func calculateConfidence(blocks []contentBlock) float64 {
	// More occurrences = higher confidence
	occurrences := float64(len(blocks))
	conf := 0.5 + (occurrences-1)*0.1

	// Different sources = higher confidence
	sources := make(map[string]bool)
	for _, b := range blocks {
		sources[b.Source] = true
	}
	sourceCount := float64(len(sources))
	conf += (sourceCount - 1) * 0.1

	if conf > 1.0 {
		conf = 1.0
	}

	return conf
}

func hashContent(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func extractWords(s string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(s) {
		w = strings.Trim(w, ".,!?;:'\"()[]{}/<>")
		if len(w) > 2 {
			words[w] = true
		}
	}
	return words
}

func wordOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	overlap := 0
	for w := range a {
		if b[w] {
			overlap++
		}
	}

	smaller := len(a)
	if len(b) < smaller {
		smaller = len(b)
	}

	return float64(overlap) / float64(smaller)
}

func unique(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
