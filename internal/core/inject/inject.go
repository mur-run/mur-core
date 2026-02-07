// Package inject provides pattern injection for mur run.
package inject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mur-run/mur-cli/internal/core/classifier"
	"github.com/mur-run/mur-cli/internal/core/embed"
	"github.com/mur-run/mur-cli/internal/core/pattern"
)

// InjectionResult holds the result of pattern injection.
type InjectionResult struct {
	// Patterns to inject
	Patterns []*pattern.Pattern
	// Formatted prompt with patterns
	FormattedPrompt string
	// Context info
	Context *ProjectContext
	// Classification scores
	Classifications []classifier.DomainScore
}

// ProjectContext holds detected project context.
type ProjectContext struct {
	// Root directory of the project
	RootDir string
	// Detected project type (go, swift, python, etc.)
	ProjectType string
	// Project name (from go.mod, package.json, etc.)
	ProjectName string
	// Current file being worked on
	CurrentFile string
	// Languages detected
	Languages []string
	// Frameworks detected
	Frameworks []string
}

// Injector handles pattern injection based on context.
type Injector struct {
	store      *pattern.Store
	classifier *classifier.HybridClassifier
	searcher   *embed.PatternSearcher // Optional semantic search
}

// NewInjector creates a new pattern injector.
func NewInjector(store *pattern.Store) *Injector {
	return &Injector{
		store:      store,
		classifier: classifier.NewHybridClassifier(),
	}
}

// WithSemanticSearch enables semantic search for pattern matching.
func (inj *Injector) WithSemanticSearch(cfg embed.Config) error {
	searcher, err := embed.NewPatternSearcher(inj.store, cfg)
	if err != nil {
		return err
	}
	inj.searcher = searcher
	return nil
}

// Inject finds and formats relevant patterns for a prompt.
func (inj *Injector) Inject(prompt string, workDir string) (*InjectionResult, error) {
	// 1. Detect project context
	ctx := inj.detectContext(workDir)

	// 2. Classify the prompt + context
	classInput := classifier.ClassifyInput{
		Content:     prompt,
		CurrentFile: ctx.CurrentFile,
		Context: map[string]interface{}{
			"project_type": ctx.ProjectType,
			"project_name": ctx.ProjectName,
			"languages":    ctx.Languages,
			"frameworks":   ctx.Frameworks,
		},
	}
	classifications := inj.classifier.Classify(classInput)

	// 3. Find matching patterns
	patterns, err := inj.findMatchingPatterns(ctx, classifications, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to find patterns: %w", err)
	}

	// 4. Format prompt with patterns
	formatted := inj.formatPrompt(prompt, patterns)

	return &InjectionResult{
		Patterns:        patterns,
		FormattedPrompt: formatted,
		Context:         ctx,
		Classifications: classifications,
	}, nil
}

// detectContext analyzes the working directory to detect project context.
func (inj *Injector) detectContext(workDir string) *ProjectContext {
	ctx := &ProjectContext{
		RootDir:   workDir,
		Languages: []string{},
		Frameworks: []string{},
	}

	// Find project root (look for common markers)
	root := findProjectRoot(workDir)
	if root != "" {
		ctx.RootDir = root
	}

	// Detect project type and name
	if info := detectGoProject(ctx.RootDir); info != nil {
		ctx.ProjectType = "go"
		ctx.ProjectName = info.moduleName
		ctx.Languages = append(ctx.Languages, "go")
	} else if info := detectSwiftProject(ctx.RootDir); info != nil {
		ctx.ProjectType = "swift"
		ctx.ProjectName = info.packageName
		ctx.Languages = append(ctx.Languages, "swift")
		if info.hasSwiftUI {
			ctx.Frameworks = append(ctx.Frameworks, "swiftui")
		}
	} else if info := detectNodeProject(ctx.RootDir); info != nil {
		ctx.ProjectType = "node"
		ctx.ProjectName = info.name
		if info.hasTypeScript {
			ctx.Languages = append(ctx.Languages, "typescript")
		} else {
			ctx.Languages = append(ctx.Languages, "javascript")
		}
		ctx.Frameworks = append(ctx.Frameworks, info.frameworks...)
	} else if info := detectPythonProject(ctx.RootDir); info != nil {
		ctx.ProjectType = "python"
		ctx.ProjectName = info.name
		ctx.Languages = append(ctx.Languages, "python")
		ctx.Frameworks = append(ctx.Frameworks, info.frameworks...)
	}

	return ctx
}

// findMatchingPatterns finds patterns that match the context and classifications.
func (inj *Injector) findMatchingPatterns(ctx *ProjectContext, classes []classifier.DomainScore, prompt string) ([]*pattern.Pattern, error) {
	maxPatterns := 5

	// Try semantic search first if available
	if inj.searcher != nil {
		searchCtx := &embed.SearchContext{
			ProjectType: ctx.ProjectType,
			ProjectName: ctx.ProjectName,
			Languages:   ctx.Languages,
			Frameworks:  ctx.Frameworks,
			CurrentFile: ctx.CurrentFile,
		}

		matches, err := inj.searcher.SearchWithContext(prompt, searchCtx, maxPatterns)
		if err == nil && len(matches) > 0 {
			// Use semantic results
			result := make([]*pattern.Pattern, 0, len(matches))
			for _, m := range matches {
				if m.Confidence > 0.3 { // Minimum semantic threshold
					result = append(result, m.Pattern)
				}
			}
			if len(result) > 0 {
				return result, nil
			}
		}
		// Fall through to keyword matching if semantic fails
	}

	// Fallback: keyword-based matching
	allPatterns, err := inj.store.List()
	if err != nil {
		return nil, err
	}

	// Score each pattern
	type scoredPattern struct {
		pattern pattern.Pattern
		score   float64
	}

	var scored []scoredPattern
	promptLower := strings.ToLower(prompt)

	for _, p := range allPatterns {
		if !p.IsActive() {
			continue
		}

		score := inj.scorePattern(&p, ctx, classes, promptLower)
		if score > 0.1 { // Minimum threshold
			scored = append(scored, scoredPattern{p, score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Take top N patterns
	if len(scored) < maxPatterns {
		maxPatterns = len(scored)
	}

	result := make([]*pattern.Pattern, maxPatterns)
	for i := 0; i < maxPatterns; i++ {
		pCopy := scored[i].pattern
		result[i] = &pCopy
	}

	return result, nil
}

// scorePattern calculates a relevance score for a pattern.
func (inj *Injector) scorePattern(p *pattern.Pattern, ctx *ProjectContext, classes []classifier.DomainScore, promptLower string) float64 {
	var score float64

	// 1. Tag matching (inferred + confirmed)
	for _, tag := range p.Tags.Confirmed {
		tagLower := strings.ToLower(tag)

		// Match against project context
		if ctx.ProjectType != "" && strings.Contains(tagLower, ctx.ProjectType) {
			score += 0.3
		}
		for _, lang := range ctx.Languages {
			if strings.Contains(tagLower, strings.ToLower(lang)) {
				score += 0.25
			}
		}
		for _, fw := range ctx.Frameworks {
			if strings.Contains(tagLower, strings.ToLower(fw)) {
				score += 0.25
			}
		}

		// Match against classifications
		for _, c := range classes {
			if strings.Contains(tagLower, strings.ToLower(c.Domain)) {
				score += c.Confidence * 0.3
			}
		}
	}

	// 2. Inferred tag matching
	for _, ts := range p.Tags.Inferred {
		tagLower := strings.ToLower(ts.Tag)

		for _, c := range classes {
			if strings.Contains(tagLower, strings.ToLower(c.Domain)) {
				score += ts.Confidence * c.Confidence * 0.2
			}
		}
	}

	// 3. Keyword matching from ApplyConditions
	for _, kw := range p.Applies.Keywords {
		if strings.Contains(promptLower, strings.ToLower(kw)) {
			score += 0.2
		}
	}

	// 4. Language/framework matching from ApplyConditions
	for _, lang := range p.Applies.Languages {
		for _, ctxLang := range ctx.Languages {
			if strings.EqualFold(lang, ctxLang) {
				score += 0.25
			}
		}
	}
	for _, fw := range p.Applies.Frameworks {
		for _, ctxFw := range ctx.Frameworks {
			if strings.EqualFold(fw, ctxFw) {
				score += 0.25
			}
		}
	}

	// 5. Project matching
	for _, proj := range p.Applies.Projects {
		if matched, _ := filepath.Match(proj, ctx.ProjectName); matched {
			score += 0.4
		}
	}

	// 6. Trust level bonus
	score *= (1.0 + p.Security.TrustLevel.Score()*0.2)

	// 7. Effectiveness bonus
	score *= (1.0 + p.Learning.Effectiveness*0.3)

	return score
}

// formatPrompt formats the prompt with injected patterns.
func (inj *Injector) formatPrompt(prompt string, patterns []*pattern.Pattern) string {
	if len(patterns) == 0 {
		return prompt
	}

	var sb strings.Builder

	// Add patterns as context
	sb.WriteString("<context>\n")
	sb.WriteString("The following patterns are relevant to this task:\n\n")

	for idx, p := range patterns {
		sb.WriteString(fmt.Sprintf("## Pattern %d: %s\n", idx+1, p.Name))
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("*%s*\n\n", p.Description))
		}
		sb.WriteString(p.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("</context>\n\n")
	sb.WriteString(prompt)

	return sb.String()
}

// ============================================================
// Project Detection Helpers
// ============================================================

type goProjectInfo struct {
	moduleName string
}

type swiftProjectInfo struct {
	packageName string
	hasSwiftUI  bool
}

type nodeProjectInfo struct {
	name          string
	hasTypeScript bool
	frameworks    []string
}

type pythonProjectInfo struct {
	name       string
	frameworks []string
}

func findProjectRoot(dir string) string {
	markers := []string{".git", "go.mod", "Package.swift", "package.json", "pyproject.toml", "Cargo.toml"}

	current := dir
	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return dir
}

func detectGoProject(root string) *goProjectInfo {
	goMod := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goMod)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return &goProjectInfo{
				moduleName: strings.TrimSpace(strings.TrimPrefix(line, "module ")),
			}
		}
	}

	return &goProjectInfo{}
}

func detectSwiftProject(root string) *swiftProjectInfo {
	packageSwift := filepath.Join(root, "Package.swift")
	data, err := os.ReadFile(packageSwift)
	if err != nil {
		// Try Xcode project
		entries, _ := os.ReadDir(root)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".xcodeproj") {
				name := strings.TrimSuffix(e.Name(), ".xcodeproj")
				return &swiftProjectInfo{packageName: name}
			}
		}
		return nil
	}

	content := string(data)
	info := &swiftProjectInfo{
		hasSwiftUI: strings.Contains(content, "SwiftUI"),
	}

	// Parse package name from Package.swift
	if idx := strings.Index(content, `name: "`); idx != -1 {
		start := idx + 7
		end := strings.Index(content[start:], `"`)
		if end != -1 {
			info.packageName = content[start : start+end]
		}
	}

	return info
}

func detectNodeProject(root string) *nodeProjectInfo {
	pkgJSON := filepath.Join(root, "package.json")
	data, err := os.ReadFile(pkgJSON)
	if err != nil {
		return nil
	}

	content := string(data)
	info := &nodeProjectInfo{
		hasTypeScript: fileExists(filepath.Join(root, "tsconfig.json")),
		frameworks:    []string{},
	}

	// Parse name
	if idx := strings.Index(content, `"name":`); idx != -1 {
		// Simple parse - find the value
		rest := content[idx+7:]
		start := strings.Index(rest, `"`)
		if start != -1 {
			end := strings.Index(rest[start+1:], `"`)
			if end != -1 {
				info.name = rest[start+1 : start+1+end]
			}
		}
	}

	// Detect frameworks
	if strings.Contains(content, `"react"`) {
		info.frameworks = append(info.frameworks, "react")
	}
	if strings.Contains(content, `"vue"`) {
		info.frameworks = append(info.frameworks, "vue")
	}
	if strings.Contains(content, `"next"`) {
		info.frameworks = append(info.frameworks, "nextjs")
	}
	if strings.Contains(content, `"express"`) {
		info.frameworks = append(info.frameworks, "express")
	}

	return info
}

func detectPythonProject(root string) *pythonProjectInfo {
	// Try pyproject.toml first
	pyproject := filepath.Join(root, "pyproject.toml")
	if data, err := os.ReadFile(pyproject); err == nil {
		content := string(data)
		info := &pythonProjectInfo{frameworks: []string{}}

		if idx := strings.Index(content, `name = "`); idx != -1 {
			start := idx + 8
			end := strings.Index(content[start:], `"`)
			if end != -1 {
				info.name = content[start : start+end]
			}
		}

		// Detect frameworks
		if strings.Contains(content, "django") {
			info.frameworks = append(info.frameworks, "django")
		}
		if strings.Contains(content, "flask") {
			info.frameworks = append(info.frameworks, "flask")
		}
		if strings.Contains(content, "fastapi") {
			info.frameworks = append(info.frameworks, "fastapi")
		}

		return info
	}

	// Try requirements.txt
	reqTxt := filepath.Join(root, "requirements.txt")
	if data, err := os.ReadFile(reqTxt); err == nil {
		content := strings.ToLower(string(data))
		info := &pythonProjectInfo{
			name:       filepath.Base(root),
			frameworks: []string{},
		}

		if strings.Contains(content, "django") {
			info.frameworks = append(info.frameworks, "django")
		}
		if strings.Contains(content, "flask") {
			info.frameworks = append(info.frameworks, "flask")
		}
		if strings.Contains(content, "fastapi") {
			info.frameworks = append(info.frameworks, "fastapi")
		}

		return info
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
