// Package classifier provides automatic pattern classification for mur.core.
package classifier

import (
	"path/filepath"
	"regexp"
	"strings"
)

// DomainScore represents a domain with its confidence score.
type DomainScore struct {
	Domain     string   `json:"domain"`
	Confidence float64  `json:"confidence"`
	Signals    []string `json:"signals,omitempty"`
}

// ClassifyInput holds the input for classification.
type ClassifyInput struct {
	// User's prompt or query
	Content string
	// Current file being worked on
	CurrentFile string
	// All files in context
	Files []string
	// Additional context
	Context map[string]interface{}
}

// Classifier is the interface for all classifiers.
type Classifier interface {
	Classify(input ClassifyInput) []DomainScore
	Name() string
}

// HybridClassifier combines multiple classification strategies.
type HybridClassifier struct {
	keyword *KeywordClassifier
	file    *FilePatternClassifier
	rule    *RuleClassifier
}

// NewHybridClassifier creates a new HybridClassifier with default settings.
func NewHybridClassifier() *HybridClassifier {
	return &HybridClassifier{
		keyword: NewKeywordClassifier(),
		file:    NewFilePatternClassifier(),
		rule:    NewRuleClassifier(),
	}
}

// Name returns the classifier name.
func (h *HybridClassifier) Name() string {
	return "hybrid"
}

// Classify performs classification using multiple strategies.
func (h *HybridClassifier) Classify(input ClassifyInput) []DomainScore {
	scores := make(map[string]*DomainScore)

	// 1. File pattern classification (highest confidence for file context)
	if input.CurrentFile != "" {
		fileScores := h.file.Classify(input)
		for _, s := range fileScores {
			if existing, ok := scores[s.Domain]; ok {
				existing.Confidence = maxFloat(existing.Confidence, s.Confidence)
				existing.Signals = append(existing.Signals, s.Signals...)
			} else {
				scores[s.Domain] = &DomainScore{
					Domain:     s.Domain,
					Confidence: s.Confidence,
					Signals:    s.Signals,
				}
			}
		}
	}

	// 2. Keyword classification
	if input.Content != "" {
		keywordScores := h.keyword.Classify(input)
		for _, s := range keywordScores {
			if existing, ok := scores[s.Domain]; ok {
				// Weighted merge
				existing.Confidence = (existing.Confidence*0.6 + s.Confidence*0.4)
				existing.Signals = append(existing.Signals, s.Signals...)
			} else {
				scores[s.Domain] = &DomainScore{
					Domain:     s.Domain,
					Confidence: s.Confidence * 0.8, // Lower base confidence for keyword-only
					Signals:    s.Signals,
				}
			}
		}
	}

	// 3. Rule-based classification
	ruleScores := h.rule.Classify(input)
	for _, s := range ruleScores {
		if existing, ok := scores[s.Domain]; ok {
			existing.Confidence = maxFloat(existing.Confidence, s.Confidence)
			existing.Signals = append(existing.Signals, s.Signals...)
		} else {
			scores[s.Domain] = &DomainScore{
				Domain:     s.Domain,
				Confidence: s.Confidence,
				Signals:    s.Signals,
			}
		}
	}

	// Convert to slice and sort by confidence
	result := make([]DomainScore, 0, len(scores))
	for _, s := range scores {
		result = append(result, *s)
	}

	// Sort descending by confidence
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Confidence > result[i].Confidence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// GetTopDomains returns the top N domains.
func (h *HybridClassifier) GetTopDomains(input ClassifyInput, n int) []DomainScore {
	all := h.Classify(input)
	if n >= len(all) {
		return all
	}
	return all[:n]
}

// GetDomainConfidence returns the confidence for a specific domain.
func (h *HybridClassifier) GetDomainConfidence(input ClassifyInput, domain string) float64 {
	all := h.Classify(input)
	for _, s := range all {
		if s.Domain == domain {
			return s.Confidence
		}
	}
	return 0.0
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ============================================================
// KeywordClassifier - keyword-based classification
// ============================================================

// KeywordClassifier classifies based on keywords in content.
type KeywordClassifier struct {
	domainKeywords map[string][]string
}

// NewKeywordClassifier creates a KeywordClassifier with default keywords.
func NewKeywordClassifier() *KeywordClassifier {
	return &KeywordClassifier{
		domainKeywords: defaultDomainKeywords(),
	}
}

func (k *KeywordClassifier) Name() string {
	return "keyword"
}

func (k *KeywordClassifier) Classify(input ClassifyInput) []DomainScore {
	content := strings.ToLower(input.Content)
	scores := make(map[string]float64)
	signals := make(map[string][]string)

	for domain, keywords := range k.domainKeywords {
		for _, kw := range keywords {
			if strings.Contains(content, strings.ToLower(kw)) {
				scores[domain] += 0.1
				signals[domain] = append(signals[domain], "keyword:"+kw)
			}
		}
	}

	var result []DomainScore
	for domain, score := range scores {
		// Cap at 0.95
		if score > 0.95 {
			score = 0.95
		}
		result = append(result, DomainScore{
			Domain:     domain,
			Confidence: score,
			Signals:    signals[domain],
		})
	}

	return result
}

func defaultDomainKeywords() map[string][]string {
	return map[string][]string{
		// Programming languages
		"swift": {"swift", "swiftui", "uikit", "appkit", "xctest", "xcode", "@State", "@Published", "Sendable"},
		"go":    {"golang", "go ", "goroutine", "chan ", "defer", "go.mod", "go.sum"},
		"rust":  {"rust", "cargo", "rustc", "impl ", "fn ", "mut ", "unwrap"},
		"python": {"python", "pip", "pytest", "django", "flask", "numpy", "pandas", "__init__"},
		"typescript": {"typescript", "tsx", "interface ", "type ", "as const"},
		"javascript": {"javascript", "nodejs", "npm", "yarn", "react", "vue", "angular"},

		// Frameworks/platforms
		"ios":     {"ios", "iphone", "ipad", "watchos", "tvos", "cocoapods", "spm"},
		"android": {"android", "kotlin", "gradle", "jetpack", "compose"},
		"web":     {"html", "css", "dom", "browser", "webpack", "vite"},
		"backend": {"api", "rest", "graphql", "grpc", "microservice", "endpoint"},

		// DevOps/infrastructure
		"devops":     {"docker", "kubernetes", "k8s", "helm", "terraform", "ansible", "ci/cd", "pipeline"},
		"database":   {"sql", "postgres", "mysql", "mongodb", "redis", "database", "query", "migration"},
		"cloud":      {"aws", "azure", "gcp", "cloud", "lambda", "s3", "ec2"},

		// Development practices
		"testing":      {"test", "unit test", "integration", "mock", "stub", "coverage", "tdd"},
		"debugging":    {"debug", "error", "exception", "crash", "stack trace", "breakpoint", "log"},
		"refactoring":  {"refactor", "clean code", "code smell", "pattern", "solid", "dry"},
		"architecture": {"architecture", "design pattern", "mvc", "mvvm", "clean architecture"},
		"security":     {"security", "auth", "oauth", "jwt", "encryption", "vulnerability", "xss", "csrf"},
		"performance":  {"performance", "optimize", "cache", "latency", "memory", "cpu", "profil"},

		// AI/ML
		"ai":  {"ai", "llm", "gpt", "claude", "gemini", "prompt", "embedding", "vector"},
		"ml":  {"machine learning", "model", "training", "inference", "neural", "tensorflow", "pytorch"},

		// Business/other
		"documentation": {"document", "readme", "changelog", "api doc", "swagger", "openapi"},
	}
}

// ============================================================
// FilePatternClassifier - file extension based classification
// ============================================================

// FilePatternClassifier classifies based on file extensions and paths.
type FilePatternClassifier struct {
	patterns map[string][]string // domain -> file patterns
}

// NewFilePatternClassifier creates a new FilePatternClassifier.
func NewFilePatternClassifier() *FilePatternClassifier {
	return &FilePatternClassifier{
		patterns: defaultFilePatterns(),
	}
}

func (f *FilePatternClassifier) Name() string {
	return "file-pattern"
}

func (f *FilePatternClassifier) Classify(input ClassifyInput) []DomainScore {
	var files []string
	if input.CurrentFile != "" {
		files = append(files, input.CurrentFile)
	}
	files = append(files, input.Files...)

	scores := make(map[string]float64)
	signals := make(map[string][]string)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		base := strings.ToLower(filepath.Base(file))

		for domain, patterns := range f.patterns {
			for _, pattern := range patterns {
				matched := false

				if strings.HasPrefix(pattern, "*.") {
					// Extension match
					if ext == pattern[1:] {
						matched = true
					}
				} else if strings.Contains(pattern, "*") {
					// Glob pattern
					if m, _ := filepath.Match(pattern, base); m {
						matched = true
					}
				} else {
					// Exact match
					if base == pattern || ext == pattern {
						matched = true
					}
				}

				if matched {
					scores[domain] += 0.3
					signals[domain] = append(signals[domain], "file:"+file)
				}
			}
		}
	}

	var result []DomainScore
	for domain, score := range scores {
		if score > 0.95 {
			score = 0.95
		}
		result = append(result, DomainScore{
			Domain:     domain,
			Confidence: score,
			Signals:    signals[domain],
		})
	}

	return result
}

func defaultFilePatterns() map[string][]string {
	return map[string][]string{
		// Languages
		"swift":      {"*.swift", "Package.swift", "*.xcodeproj", "*.xcworkspace"},
		"go":         {"*.go", "go.mod", "go.sum"},
		"rust":       {"*.rs", "Cargo.toml", "Cargo.lock"},
		"python":     {"*.py", "requirements.txt", "setup.py", "pyproject.toml"},
		"typescript": {"*.ts", "*.tsx", "tsconfig.json"},
		"javascript": {"*.js", "*.jsx", "package.json"},

		// Web
		"web":  {"*.html", "*.css", "*.scss", "*.sass", "*.less"},
		"vue":  {"*.vue"},
		"react": {"*.jsx", "*.tsx"},

		// Config/DevOps
		"devops":   {"Dockerfile", "docker-compose.yml", "*.yaml", "*.yml", "Makefile", ".github"},
		"database": {"*.sql", "migrations", "schema.prisma"},

		// Docs
		"documentation": {"*.md", "README*", "CHANGELOG*", "docs"},

		// iOS/macOS
		"ios": {"*.swift", "*.storyboard", "*.xib", "Info.plist", "*.entitlements"},

		// Testing
		"testing": {"*_test.go", "*_test.swift", "*.spec.ts", "*.test.js", "__tests__"},
	}
}

// ============================================================
// RuleClassifier - rule-based classification
// ============================================================

// RuleClassifier uses predefined rules for classification.
type RuleClassifier struct {
	rules []ClassificationRule
}

// ClassificationRule defines a classification rule.
type ClassificationRule struct {
	Name       string
	Domain     string
	Confidence float64
	Match      func(input ClassifyInput) bool
}

// NewRuleClassifier creates a RuleClassifier with default rules.
func NewRuleClassifier() *RuleClassifier {
	return &RuleClassifier{
		rules: defaultRules(),
	}
}

func (r *RuleClassifier) Name() string {
	return "rule"
}

func (r *RuleClassifier) Classify(input ClassifyInput) []DomainScore {
	var result []DomainScore

	for _, rule := range r.rules {
		if rule.Match(input) {
			result = append(result, DomainScore{
				Domain:     rule.Domain,
				Confidence: rule.Confidence,
				Signals:    []string{"rule:" + rule.Name},
			})
		}
	}

	return result
}

func defaultRules() []ClassificationRule {
	return []ClassificationRule{
		// Error handling rules
		{
			Name:       "error-pattern",
			Domain:     "debugging",
			Confidence: 0.8,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				errorPatterns := []string{"error:", "exception", "failed", "crash", "panic", "fatal"}
				for _, p := range errorPatterns {
					if strings.Contains(content, p) {
						return true
					}
				}
				return false
			},
		},
		// Question patterns
		{
			Name:       "how-to-question",
			Domain:     "learning",
			Confidence: 0.6,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				patterns := []string{"how to", "how do i", "what is", "why does", "explain"}
				for _, p := range patterns {
					if strings.Contains(content, p) {
						return true
					}
				}
				return false
			},
		},
		// Refactoring patterns
		{
			Name:       "refactor-request",
			Domain:     "refactoring",
			Confidence: 0.85,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				patterns := []string{"refactor", "clean up", "improve", "optimize", "simplify"}
				for _, p := range patterns {
					if strings.Contains(content, p) {
						return true
					}
				}
				return false
			},
		},
		// Test patterns
		{
			Name:       "test-request",
			Domain:     "testing",
			Confidence: 0.85,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				patterns := []string{"write test", "add test", "unit test", "test case", "test coverage"}
				for _, p := range patterns {
					if strings.Contains(content, p) {
						return true
					}
				}
				return false
			},
		},
		// Documentation patterns
		{
			Name:       "doc-request",
			Domain:     "documentation",
			Confidence: 0.8,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				patterns := []string{"document", "readme", "comment", "explain this", "add docs"}
				for _, p := range patterns {
					if strings.Contains(content, p) {
						return true
					}
				}
				return false
			},
		},
		// API patterns
		{
			Name:       "api-work",
			Domain:     "backend",
			Confidence: 0.75,
			Match: func(input ClassifyInput) bool {
				content := strings.ToLower(input.Content)
				apiPattern := regexp.MustCompile(`(get|post|put|delete|patch)\s+(api|endpoint|route)`)
				return apiPattern.MatchString(content)
			},
		},
	}
}
