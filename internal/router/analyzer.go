// Package router provides intelligent tool selection based on prompt analysis.
package router

import (
	"strings"
)

// ComplexityKeywords maps keywords to their complexity weight.
var ComplexityKeywords = map[string]float64{
	// High complexity (0.25-0.3 each)
	"refactor":     0.30,
	"architecture": 0.30,
	"redesign":     0.30,
	"optimize":     0.25,
	"debug":        0.25,
	"fix bug":      0.25,
	"performance":  0.25,
	"security":     0.25,

	// Medium complexity (0.1-0.2 each)
	"implement": 0.20,
	"create":    0.15,
	"build":     0.15,
	"write":     0.10,
	"code":      0.10,
	"function":  0.10,
	"class":     0.10,
	"module":    0.15,
	"test":      0.10,

	// Low complexity indicators (reduce score)
	"what is":   -0.15,
	"what's":    -0.15,
	"explain":   -0.10,
	"tell me":   -0.10,
	"help":      -0.05,
	"define":    -0.10,
	"meaning":   -0.10,
	"simple":    -0.10,
	"basic":     -0.10,
	"quick":     -0.10,
	"how do i":  -0.05,
	"how to":    -0.05,
	"example":   -0.05,
	"summarize": -0.05,
}

// ToolUseKeywords indicate file/code operations that may need tool use.
var ToolUseKeywords = []string{
	"file", "read file", "write file", "edit file", "create file",
	"run", "execute", "test", "build", "compile",
	"in this project", "in this repo", "this code", "this file",
	"current directory", "codebase", "repository",
	"modify", "change", "update the",
}

// PromptAnalysis contains the result of analyzing a prompt.
type PromptAnalysis struct {
	Complexity   float64  // 0-1 score
	Length       int      // Character count
	Keywords     []string // Detected complexity keywords
	NeedsToolUse bool     // Appears to need file/code operations
	Category     string   // simple-qa, coding, analysis, architecture, debugging, general
}

// AnalyzePrompt returns a complexity analysis of the prompt.
func AnalyzePrompt(prompt string) PromptAnalysis {
	lower := strings.ToLower(prompt)

	analysis := PromptAnalysis{
		Length:   len(prompt),
		Category: detectCategory(lower),
	}

	// Calculate keyword-based complexity
	keywordScore := 0.0
	for keyword, weight := range ComplexityKeywords {
		if strings.Contains(lower, keyword) {
			keywordScore += weight
			if weight > 0 {
				analysis.Keywords = append(analysis.Keywords, keyword)
			}
		}
	}

	// Add length factor
	lengthScore := lengthFactor(len(prompt))

	// Check for tool use indicators
	analysis.NeedsToolUse = detectToolUse(lower)
	toolUseScore := 0.0
	if analysis.NeedsToolUse {
		toolUseScore = 0.2
	}

	// Combine scores (weighted)
	analysis.Complexity = clamp(keywordScore*0.6+lengthScore*0.2+toolUseScore*0.2, 0, 1)

	return analysis
}

// lengthFactor returns a complexity factor based on prompt length.
func lengthFactor(length int) float64 {
	switch {
	case length < 50:
		return 0.0 // Very short = simple
	case length < 150:
		return 0.1
	case length < 300:
		return 0.2
	case length < 500:
		return 0.3
	case length < 1000:
		return 0.5
	default:
		return 0.7 // Long prompts tend to be complex
	}
}

// detectCategory determines the type of task from the prompt.
func detectCategory(lower string) string {
	if containsAny(lower, []string{"what is", "what's", "explain", "tell me about", "how does", "define", "meaning of"}) {
		return "simple-qa"
	}
	if containsAny(lower, []string{"refactor", "architecture", "design", "redesign", "restructure"}) {
		return "architecture"
	}
	if containsAny(lower, []string{"debug", "fix", "error", "bug", "issue", "crash", "failing"}) {
		return "debugging"
	}
	// Check analysis before coding since "analyze" contains common words
	if containsAny(lower, []string{"analyze", "review", "evaluate", "compare", "assess"}) {
		return "analysis"
	}
	if containsAny(lower, []string{"write", "create", "implement", "build", "code", "function", "class"}) {
		return "coding"
	}

	return "general"
}

// detectToolUse checks if the prompt likely needs file/code operations.
func detectToolUse(lower string) bool {
	return containsAny(lower, ToolUseKeywords)
}

// containsAny checks if s contains any of the substrings.
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// clamp restricts v to the range [min, max].
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
