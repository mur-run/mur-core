package learn

import (
	"strings"
)

// SessionQuality holds metrics about a session's extraction potential.
type SessionQuality struct {
	ToolUseCount     int     // Number of tool_use blocks
	UserMessageCount int     // Number of user messages
	BackAndForth     int     // Conversation turns (alternating messages)
	HasErrorPattern  bool    // Contains error/fix patterns
	AssistantRatio   float64 // Ratio of assistant content to total
	TotalMessages    int     // Total message count
}

// ExtractionConfig holds thresholds for quality filtering.
type ExtractionConfig struct {
	MinToolUses          int      // Minimum tool uses required
	MinTurns             int      // Minimum conversation turns
	MaxAssistantRatio    float64  // Max ratio of assistant content
	MinContentLength     int      // Minimum pattern content length
	RequireProblemSolve  bool     // Require problem/solution structure
	RejectKeywords       []string // Keywords that indicate generic content
}

// DefaultExtractionConfig returns sensible defaults.
func DefaultExtractionConfig() ExtractionConfig {
	return ExtractionConfig{
		MinToolUses:         3,
		MinTurns:            4,
		MaxAssistantRatio:   0.85,
		MinContentLength:    100,
		RequireProblemSolve: true,
		RejectKeywords: []string{
			"how to",
			"tutorial",
			"example",
			"introduction",
			"basic",
			"simple guide",
		},
	}
}

// AnalyzeSessionQuality analyzes a session for extraction quality.
func AnalyzeSessionQuality(session *Session) SessionQuality {
	q := SessionQuality{
		TotalMessages: len(session.Messages),
		ToolUseCount:  session.ToolUseCount, // Use pre-counted from JSONL parsing
	}

	var userContentLen, assistantContentLen int
	var lastRole string

	for _, msg := range session.Messages {
		switch msg.Role {
		case "user":
			q.UserMessageCount++
			userContentLen += len(msg.Content)
		case "assistant":
			assistantContentLen += len(msg.Content)
		}

		// Count back-and-forth
		if lastRole != "" && lastRole != msg.Role {
			q.BackAndForth++
		}
		lastRole = msg.Role
	}

	// Calculate assistant ratio
	totalContent := userContentLen + assistantContentLen
	if totalContent > 0 {
		q.AssistantRatio = float64(assistantContentLen) / float64(totalContent)
	}

	// Check for error patterns
	transcript := session.FullTranscript()
	errorIndicators := []string{
		"error:", "Error:", "ERROR:",
		"failed", "Failed", "FAILED",
		"exception", "Exception",
		"bug", "Bug", "BUG",
		"fix", "Fix", "FIX",
		"workaround", "Workaround",
		"solved", "Solved",
		"the issue was", "the problem was",
	}
	for _, indicator := range errorIndicators {
		if strings.Contains(transcript, indicator) {
			q.HasErrorPattern = true
			break
		}
	}

	return q
}

// ShouldExtract determines if a session is worth extracting from.
func ShouldExtract(q SessionQuality, cfg ExtractionConfig) (bool, string) {
	// Check tool usage
	if q.ToolUseCount < cfg.MinToolUses {
		return false, "too few tool uses (no actual coding)"
	}

	// Check conversation depth
	if q.BackAndForth < cfg.MinTurns {
		return false, "too short (likely just a question)"
	}

	// Check content balance
	if q.AssistantRatio > cfg.MaxAssistantRatio {
		return false, "mostly AI talking (likely tutorial)"
	}

	return true, ""
}

// ValidatePattern checks if an extracted pattern is high quality.
func ValidatePattern(p Pattern, cfg ExtractionConfig) (bool, string) {
	titleLower := strings.ToLower(p.Description)
	contentLower := strings.ToLower(p.Content)

	// Check for generic keywords in title
	for _, keyword := range cfg.RejectKeywords {
		if strings.Contains(titleLower, keyword) {
			return false, "title contains generic keyword: " + keyword
		}
	}

	// Check minimum content length
	if len(p.Content) < cfg.MinContentLength {
		return false, "content too short"
	}

	// Check for problem/solution structure
	if cfg.RequireProblemSolve {
		hasProblem := strings.Contains(contentLower, "problem") ||
			strings.Contains(contentLower, "error") ||
			strings.Contains(contentLower, "issue") ||
			strings.Contains(contentLower, "bug") ||
			strings.Contains(contentLower, "failed")

		hasSolution := strings.Contains(contentLower, "solution") ||
			strings.Contains(contentLower, "fix") ||
			strings.Contains(contentLower, "workaround") ||
			strings.Contains(contentLower, "resolved") ||
			strings.Contains(contentLower, "solved")

		if !hasProblem && !hasSolution {
			return false, "no problem/solution structure"
		}
	}

	// Check for overly generic pattern names
	genericNames := []string{
		"testing-pattern",
		"basic-",
		"simple-",
		"how-to-",
		"guide-",
	}
	nameLower := strings.ToLower(p.Name)
	for _, generic := range genericNames {
		if strings.HasPrefix(nameLower, generic) || nameLower == strings.TrimSuffix(generic, "-") {
			return false, "pattern name is too generic"
		}
	}

	return true, ""
}

// FilterPatterns applies validation to a list of patterns.
func FilterPatterns(patterns []ExtractedPattern, cfg ExtractionConfig) []ExtractedPattern {
	var filtered []ExtractedPattern

	for _, ep := range patterns {
		valid, _ := ValidatePattern(ep.Pattern, cfg)
		if valid {
			filtered = append(filtered, ep)
		}
	}

	return filtered
}
