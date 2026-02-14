// Package learn provides cross-CLI learning capabilities.
package learn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/core/suggest"
)

// CLISource represents an AI CLI tool as a learning source.
type CLISource struct {
	Name        string
	SessionDir  string   // Path to session/history directory
	FilePattern string   // Glob pattern for session files
	Parser      SessionParser
}

// SessionParser parses session files from a specific CLI.
type SessionParser interface {
	Parse(path string) ([]SessionEntry, error)
}

// SessionEntry represents a single exchange in a session.
type SessionEntry struct {
	Role      string    // "user" or "assistant"
	Content   string
	Timestamp time.Time
	Tool      string    // Which tool was used (if any)
	Success   bool      // Whether the action succeeded
}

// DefaultCLISources returns the known CLI sources.
func DefaultCLISources() []CLISource {
	home, _ := os.UserHomeDir()

	return []CLISource{
		{
			Name:        "Claude Code",
			SessionDir:  filepath.Join(home, ".claude", "projects"),
			FilePattern: "*/conversation.jsonl",
			Parser:      &ClaudeParser{},
		},
		{
			Name:        "Gemini CLI",
			SessionDir:  filepath.Join(home, ".gemini", "history"),
			FilePattern: "*.json",
			Parser:      &GeminiParser{},
		},
		{
			Name:        "Auggie",
			SessionDir:  filepath.Join(home, ".augment", "sessions"),
			FilePattern: "*.json",
			Parser:      &AuggieParser{},
		},
		{
			Name:        "Codex",
			SessionDir:  filepath.Join(home, ".codex", "history"),
			FilePattern: "*.jsonl",
			Parser:      &CodexParser{},
		},
		{
			Name:        "Aider",
			SessionDir:  filepath.Join(home, ".aider", "history"),
			FilePattern: "*.md",
			Parser:      &AiderParser{},
		},
		{
			Name:        "Continue",
			SessionDir:  filepath.Join(home, ".continue", "sessions"),
			FilePattern: "*.json",
			Parser:      &ContinueParser{},
		},
		{
			Name:        "OpenClaw",
			SessionDir:  filepath.Join(home, ".openclaw", "agents", "main", "sessions"),
			FilePattern: "*.jsonl",
			Parser:      &OpenClawParser{},
		},
	}
}

// CrossCLILearner extracts patterns from multiple CLI sources.
type CrossCLILearner struct {
	sources   []CLISource
	extractor *suggest.Extractor
	store     *pattern.Store
}

// NewCrossCLILearner creates a new cross-CLI learner.
func NewCrossCLILearner(store *pattern.Store) *CrossCLILearner {
	home, _ := os.UserHomeDir()
	suggestDir := filepath.Join(home, ".mur", "suggestions")

	return &CrossCLILearner{
		sources:   DefaultCLISources(),
		extractor: suggest.NewExtractor(store, suggestDir, suggest.DefaultExtractorConfig()),
		store:     store,
	}
}

// LearnResult holds the result of learning from CLI sources.
type LearnResult struct {
	Source      string
	FilesRead   int
	Entries     int
	Suggestions []suggest.Suggestion
	Error       error
}

// LearnFromAll extracts patterns from all configured CLI sources.
func (l *CrossCLILearner) LearnFromAll() ([]LearnResult, error) {
	var results []LearnResult

	for _, source := range l.sources {
		result := l.learnFromSource(source)
		results = append(results, result)
	}

	return results, nil
}

// LearnFromSource extracts patterns from a specific CLI source.
func (l *CrossCLILearner) LearnFromSource(name string) (*LearnResult, error) {
	for _, source := range l.sources {
		if strings.EqualFold(source.Name, name) {
			result := l.learnFromSource(source)
			return &result, result.Error
		}
	}
	return nil, fmt.Errorf("unknown CLI source: %s", name)
}

// learnFromSource extracts patterns from a single source.
func (l *CrossCLILearner) learnFromSource(source CLISource) LearnResult {
	result := LearnResult{
		Source: source.Name,
	}

	// Check if directory exists
	if _, err := os.Stat(source.SessionDir); os.IsNotExist(err) {
		result.Error = fmt.Errorf("session directory not found: %s", source.SessionDir)
		return result
	}

	// Find session files
	pattern := filepath.Join(source.SessionDir, source.FilePattern)
	files, err := filepath.Glob(pattern)
	if err != nil {
		result.Error = fmt.Errorf("failed to find session files: %w", err)
		return result
	}

	result.FilesRead = len(files)

	// Parse all sessions
	var allEntries []SessionEntry
	for _, f := range files {
		entries, err := source.Parser.Parse(f)
		if err != nil {
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	result.Entries = len(allEntries)

	// Extract patterns from entries
	suggestions := l.extractFromEntries(allEntries, source.Name)
	result.Suggestions = suggestions

	return result
}

// extractFromEntries extracts pattern suggestions from session entries.
func (l *CrossCLILearner) extractFromEntries(entries []SessionEntry, source string) []suggest.Suggestion {
	var suggestions []suggest.Suggestion

	// Group consecutive entries into conversations
	conversations := groupConversations(entries)

	for _, conv := range conversations {
		// Look for successful problem-solution patterns
		if pattern := extractProblemSolution(conv); pattern != nil {
			pattern.Sources = []string{source}
			suggestions = append(suggestions, *pattern)
		}

		// Look for code patterns
		if patterns := extractCodePatterns(conv); len(patterns) > 0 {
			for i := range patterns {
				patterns[i].Sources = []string{source}
			}
			suggestions = append(suggestions, patterns...)
		}

		// Look for workflow patterns
		if pattern := extractWorkflowPattern(conv); pattern != nil {
			pattern.Sources = []string{source}
			suggestions = append(suggestions, *pattern)
		}
	}

	// Deduplicate
	return deduplicateSuggestions(suggestions)
}

// Conversation represents a series of related entries.
type Conversation struct {
	Entries []SessionEntry
	Topic   string
}

// groupConversations groups entries into logical conversations.
func groupConversations(entries []SessionEntry) []Conversation {
	var conversations []Conversation
	var current []SessionEntry

	for i, entry := range entries {
		current = append(current, entry)

		// Start new conversation on significant gaps or topic changes
		isLast := i == len(entries)-1
		hasGap := !isLast && entries[i+1].Timestamp.Sub(entry.Timestamp) > 30*time.Minute

		if isLast || hasGap || len(current) >= 20 {
			if len(current) >= 2 {
				conversations = append(conversations, Conversation{
					Entries: current,
					Topic:   detectTopic(current),
				})
			}
			current = nil
		}
	}

	return conversations
}

// detectTopic tries to detect the main topic of a conversation.
func detectTopic(entries []SessionEntry) string {
	// Look for keywords in first user message
	for _, e := range entries {
		if e.Role == "user" {
			content := strings.ToLower(e.Content)

			topics := map[string][]string{
				"debugging":    {"bug", "error", "fix", "issue", "problem"},
				"refactoring":  {"refactor", "clean", "improve", "optimize"},
				"testing":      {"test", "spec", "coverage"},
				"feature":      {"add", "implement", "create", "build"},
				"documentation": {"doc", "readme", "comment", "explain"},
			}

			for topic, keywords := range topics {
				for _, kw := range keywords {
					if strings.Contains(content, kw) {
						return topic
					}
				}
			}
			break
		}
	}

	return "general"
}

// extractProblemSolution extracts problem-solution patterns.
func extractProblemSolution(conv Conversation) *suggest.Suggestion {
	// Look for: user describes problem -> assistant provides solution -> success
	var problem, solution string

	for i, entry := range conv.Entries {
		if entry.Role == "user" && problem == "" {
			// First user message as problem
			if len(entry.Content) > 30 {
				problem = entry.Content
			}
		} else if entry.Role == "assistant" && problem != "" && solution == "" {
			// Assistant response as potential solution
			if len(entry.Content) > 100 {
				solution = entry.Content

				// Check if followed by success indicators
				if i+1 < len(conv.Entries) {
					next := conv.Entries[i+1]
					if next.Role == "user" && isPositiveFeedback(next.Content) {
						return &suggest.Suggestion{
							Name:        fmt.Sprintf("%s-solution", conv.Topic),
							Description: truncate(problem, 100),
							Content:     extractKeyContent(solution),
							Confidence:  0.7,
							Tags:        []string{conv.Topic, "solution"},
							Reason:      "Extracted from successful problem-solution exchange",
						}
					}
				}
			}
		}
	}

	return nil
}

// extractCodePatterns extracts code patterns from conversations.
func extractCodePatterns(conv Conversation) []suggest.Suggestion {
	var suggestions []suggest.Suggestion

	codeBlockRe := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")

	for _, entry := range conv.Entries {
		if entry.Role != "assistant" {
			continue
		}

		matches := codeBlockRe.FindAllStringSubmatch(entry.Content, -1)
		for _, m := range matches {
			if len(m) > 2 && len(m[2]) > 50 {
				lang := m[1]
				code := m[2]

				// Skip if too short or looks like output
				if len(code) < 50 || looksLikeOutput(code) {
					continue
				}

				suggestions = append(suggestions, suggest.Suggestion{
					Name:        fmt.Sprintf("%s-%s-pattern", conv.Topic, lang),
					Description: fmt.Sprintf("Code pattern for %s", conv.Topic),
					Content:     code,
					Confidence:  0.6,
					Tags:        []string{lang, conv.Topic, "code"},
					Reason:      "Extracted code block from AI response",
				})
			}
		}
	}

	return suggestions
}

// extractWorkflowPattern extracts multi-step workflow patterns.
func extractWorkflowPattern(conv Conversation) *suggest.Suggestion {
	// Look for numbered steps or bullet lists
	stepRe := regexp.MustCompile(`(?m)^(\d+\.|[-*])\s+.+`)

	for _, entry := range conv.Entries {
		if entry.Role != "assistant" {
			continue
		}

		matches := stepRe.FindAllString(entry.Content, -1)
		if len(matches) >= 3 {
			workflow := strings.Join(matches, "\n")
			return &suggest.Suggestion{
				Name:        fmt.Sprintf("%s-workflow", conv.Topic),
				Description: fmt.Sprintf("Workflow for %s", conv.Topic),
				Content:     workflow,
				Confidence:  0.65,
				Tags:        []string{conv.Topic, "workflow"},
				Reason:      "Extracted multi-step workflow",
			}
		}
	}

	return nil
}

// Helper functions

func isPositiveFeedback(content string) bool {
	positive := []string{
		"thanks", "thank you", "perfect", "great", "works",
		"awesome", "excellent", "nice", "good", "solved",
		"fixed", "done", "correct", "yes",
	}
	lower := strings.ToLower(content)
	for _, p := range positive {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func looksLikeOutput(code string) bool {
	// Check for common output patterns
	outputIndicators := []string{
		"error:", "warning:", "npm ERR", "go:", "===",
		"PASS", "FAIL", "✓", "✗", ">>>",
	}
	for _, ind := range outputIndicators {
		if strings.Contains(code, ind) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func extractKeyContent(s string) string {
	// Try to extract the most important part
	lines := strings.Split(s, "\n")
	var keyLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip fluff
		if strings.HasPrefix(line, "Sure") || strings.HasPrefix(line, "I'll") ||
			strings.HasPrefix(line, "Let me") || strings.HasPrefix(line, "Here") {
			continue
		}
		keyLines = append(keyLines, line)
		if len(keyLines) >= 10 {
			break
		}
	}

	return strings.Join(keyLines, "\n")
}

func deduplicateSuggestions(suggestions []suggest.Suggestion) []suggest.Suggestion {
	seen := make(map[string]bool)
	var unique []suggest.Suggestion

	for _, s := range suggestions {
		key := strings.ToLower(s.Name + s.Content[:min(50, len(s.Content))])
		if !seen[key] {
			seen[key] = true
			unique = append(unique, s)
		}
	}

	return unique
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================
// CLI-Specific Parsers
// ============================================================

// ClaudeParser parses Claude Code session files.
type ClaudeParser struct{}

func (p *ClaudeParser) Parse(path string) ([]SessionEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []SessionEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var msg struct {
			Type      string `json:"type"`
			Role      string `json:"role"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.Type == "message" && (msg.Role == "user" || msg.Role == "assistant") {
			ts, _ := time.Parse(time.RFC3339, msg.Timestamp)
			entries = append(entries, SessionEntry{
				Role:      msg.Role,
				Content:   msg.Content,
				Timestamp: ts,
			})
		}
	}

	return entries, nil
}

// GeminiParser parses Gemini CLI session files.
type GeminiParser struct{}

func (p *GeminiParser) Parse(path string) ([]SessionEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	var entries []SessionEntry
	for _, msg := range session.Messages {
		entries = append(entries, SessionEntry{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return entries, nil
}

// AuggieParser parses Auggie session files.
type AuggieParser struct{}

func (p *AuggieParser) Parse(path string) ([]SessionEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session struct {
		SessionID   string `json:"sessionId"`
		Created     string `json:"created"`
		ChatHistory []struct {
			Exchange struct {
				RequestMessage string `json:"request_message"`
				ResponseText   string `json:"response_text"`
			} `json:"exchange"`
		} `json:"chatHistory"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	var entries []SessionEntry
	ts, _ := time.Parse(time.RFC3339, session.Created)

	for _, chat := range session.ChatHistory {
		if chat.Exchange.RequestMessage != "" {
			entries = append(entries, SessionEntry{
				Role:      "user",
				Content:   chat.Exchange.RequestMessage,
				Timestamp: ts,
			})
		}
		if chat.Exchange.ResponseText != "" {
			entries = append(entries, SessionEntry{
				Role:      "assistant",
				Content:   chat.Exchange.ResponseText,
				Timestamp: ts,
			})
		}
	}

	return entries, nil
}

// CodexParser parses Codex session files.
type CodexParser struct{}

func (p *CodexParser) Parse(path string) ([]SessionEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []SessionEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var msg struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		entries = append(entries, SessionEntry{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return entries, nil
}

// AiderParser parses Aider session files (markdown).
type AiderParser struct{}

func (p *AiderParser) Parse(path string) ([]SessionEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	var entries []SessionEntry

	// Parse markdown format: > user message, assistant response
	userRe := regexp.MustCompile(`(?m)^>\s*(.+)$`)
	userMatches := userRe.FindAllStringSubmatch(content, -1)

	for _, m := range userMatches {
		if len(m) > 1 {
			entries = append(entries, SessionEntry{
				Role:    "user",
				Content: m[1],
			})
		}
	}

	// Everything else is assistant
	cleaned := userRe.ReplaceAllString(content, "")
	if len(cleaned) > 50 {
		entries = append(entries, SessionEntry{
			Role:    "assistant",
			Content: cleaned,
		})
	}

	return entries, nil
}

// ContinueParser parses Continue session files.
type ContinueParser struct{}

func (p *ContinueParser) Parse(path string) ([]SessionEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session struct {
		History []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"history"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	var entries []SessionEntry
	for _, msg := range session.History {
		entries = append(entries, SessionEntry{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return entries, nil
}
