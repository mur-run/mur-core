package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AnalysisResult is the structured workflow extracted from a session transcript.
type AnalysisResult struct {
	Name        string     `json:"name" yaml:"name"`
	Trigger     string     `json:"trigger" yaml:"trigger"`
	Description string     `json:"description" yaml:"description"`
	Variables   []Variable `json:"variables" yaml:"variables,omitempty"`
	Steps       []Step     `json:"steps" yaml:"steps"`
	Tools       []string   `json:"tools" yaml:"tools,omitempty"`
	Tags        []string   `json:"tags" yaml:"tags,omitempty"`
}

// Variable represents a parameterizable value in a workflow.
type Variable struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"` // string, path, url, number, bool
	Required    bool   `json:"required" yaml:"required"`
	Default     string `json:"default" yaml:"default,omitempty"`
	Description string `json:"description" yaml:"description"`
}

// Step represents a single action in a workflow.
type Step struct {
	Order         int    `json:"order" yaml:"order"`
	Description   string `json:"description" yaml:"description"`
	Command       string `json:"command,omitempty" yaml:"command,omitempty"`
	Tool          string `json:"tool,omitempty" yaml:"tool,omitempty"`
	NeedsApproval bool   `json:"needs_approval" yaml:"needs_approval"`
	OnFailure     string `json:"on_failure" yaml:"on_failure"` // skip, abort, retry
}

// qaCoTPrompt is the Question-Answer Chain of Thought prompt for analysis.
const qaCoTPrompt = `Analyze this AI conversation transcript and extract a reusable workflow.

IMPORTANT: Focus ONLY on the events in the transcript below. Do NOT infer, hallucinate, or reference any context not explicitly present in the transcript. If the transcript is about price comparison, the workflow must be about price comparison — not about something else entirely.

Answer each question step by step:

Q1: What was the user's FIRST message? What specific problem or goal does it describe?
Q2: What was the root cause (if debugging)? If not debugging, write "N/A".
Q3: What steps were attempted? Which succeeded, which failed? List them chronologically.
Q4: What is the minimal correct sequence of steps to solve this?
Q5: What tools or commands were used at each step?
Q6: Which values are environment-specific and should be variables?
Q7: Are there conditional branches (if X then Y)?
Q8: Which steps need human approval before proceeding?
Q9: Based ONLY on Q1's answer, what's a good kebab-case name and trigger description?
Q10: What tags would help find this workflow later?

After your analysis, output ONLY a JSON object (no markdown fences) with this structure:
{
  "name": "kebab-case-name",
  "trigger": "when to use this workflow",
  "description": "what this workflow does",
  "variables": [
    {"name": "var_name", "type": "string", "required": true, "default": "", "description": "what it is"}
  ],
  "steps": [
    {"order": 1, "description": "what to do", "command": "optional command", "tool": "optional tool", "needs_approval": false, "on_failure": "abort"}
  ],
  "tools": ["tool1", "tool2"],
  "tags": ["tag1", "tag2"]
}

TRANSCRIPT:
%s`

// Analyze reads a session's JSONL transcript, sends it through the LLM
// with the QA-CoT prompt, and returns a structured AnalysisResult.
func Analyze(sessionID string, provider LLMProvider) (*AnalysisResult, error) {
	events, err := ReadEvents(sessionID)
	if err != nil {
		return nil, fmt.Errorf("read transcript: %w", err)
	}

	// Filter events: only keep events after session start time
	// and exclude mur's own management commands.
	events = filterSessionEvents(sessionID, events)

	if len(events) == 0 {
		return nil, fmt.Errorf("session %s has no events", sessionID)
	}

	transcript := formatTranscript(events)
	prompt := fmt.Sprintf(qaCoTPrompt, transcript)

	raw, err := provider.Complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis: %w", err)
	}

	result, err := parseAnalysisResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	return result, nil
}

// filterSessionEvents removes events that don't belong to the session:
//  1. Events with timestamps before the session's startedAt time
//  2. Events that are mur's own management commands (session start/stop/analyze, etc.)
func filterSessionEvents(sessionID string, events []EventRecord) []EventRecord {
	// Get session start time from state or first event
	var startTS int64
	resolved, err := ResolveSessionID(sessionID)
	if err == nil {
		sessionID = resolved
	}
	// Try reading the recording's start time from active.json metadata
	// If unavailable, use the first event's timestamp as a fallback
	if state, err := loadSessionMeta(sessionID); err == nil && state.StartedAt > 0 {
		startTS = state.StartedAt
	} else if len(events) > 0 {
		startTS = events[0].Timestamp
	}

	filtered := make([]EventRecord, 0, len(events))
	for _, e := range events {
		// Skip events before session start
		if startTS > 0 && e.Timestamp > 0 && e.Timestamp < startTS {
			continue
		}
		// Skip mur's own management commands
		if isMurMetaEvent(e) {
			continue
		}
		// Skip system prompt / config injections
		if isSystemPromptEvent(e) {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

// isSystemPromptEvent returns true if the event looks like a system prompt
// or configuration injection (AGENTS.md, SOUL.md, etc.) that shouldn't
// be part of workflow transcripts.
func isSystemPromptEvent(e EventRecord) bool {
	if e.Type != "user" && e.Type != "assistant" {
		return false
	}
	// Detect common system prompt markers
	markers := []string{
		"# AGENTS.md",
		"# SOUL.md",
		"# CLAUDE.md",
		"<system>",
		"<!-- system prompt",
		"## Runtime\nRuntime: agent=",
	}
	for _, m := range markers {
		if strings.Contains(e.Content, m) {
			return true
		}
	}
	return false
}

// isMurMetaEvent returns true if the event is a mur self-management command
// (e.g. "mur session start", "mur session stop", "mur context", "mur search").
func isMurMetaEvent(e EventRecord) bool {
	if e.Type != "tool_call" && e.Type != "tool_result" {
		return false
	}
	content := strings.TrimSpace(e.Content)
	// Check for common mur CLI invocations
	murPrefixes := []string{
		"mur session",
		"mur context",
		"mur search",
		"mur learn",
		"mur sync",
		"mur feedback",
		"mur stats",
		"mur doctor",
		"mur consolidate",
	}
	lower := strings.ToLower(content)
	for _, prefix := range murPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	// Also check tool name metadata
	if strings.HasPrefix(strings.ToLower(e.Tool), "mur") {
		return true
	}
	return false
}

// loadSessionMeta tries to read the session metadata from a sidecar JSON file.
// Sessions store their start time when recording begins.
func loadSessionMeta(sessionID string) (*RecordingState, error) {
	recDir, err := recordingsDir()
	if err != nil {
		return nil, err
	}
	metaPath := filepath.Join(recDir, sessionID+".meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var state RecordingState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// formatTranscript converts EventRecords into a readable text transcript.
// Deduplicates consecutive events with the same content (retry protection).
func formatTranscript(events []EventRecord) string {
	var b strings.Builder
	var prevContent string

	for _, e := range events {
		// Skip exact consecutive duplicates (retry dedup)
		key := e.Type + "|" + e.Tool + "|" + e.Content
		if key == prevContent {
			continue
		}
		prevContent = key
		switch e.Type {
		case "user":
			fmt.Fprintf(&b, "[USER] %s\n", e.Content)
		case "assistant":
			fmt.Fprintf(&b, "[ASSISTANT] %s\n", e.Content)
		case "tool_call":
			if e.Tool != "" {
				fmt.Fprintf(&b, "[TOOL_CALL: %s] %s\n", e.Tool, e.Content)
			} else {
				fmt.Fprintf(&b, "[TOOL_CALL] %s\n", e.Content)
			}
		case "tool_result":
			if e.Tool != "" {
				fmt.Fprintf(&b, "[TOOL_RESULT: %s] %s\n", e.Tool, e.Content)
			} else {
				fmt.Fprintf(&b, "[TOOL_RESULT] %s\n", e.Content)
			}
		default:
			fmt.Fprintf(&b, "[%s] %s\n", strings.ToUpper(e.Type), e.Content)
		}
	}

	return b.String()
}

// parseAnalysisResponse extracts the JSON object from the LLM's response.
// The LLM may include reasoning text before the JSON; we find the JSON block.
func parseAnalysisResponse(raw string) (*AnalysisResult, error) {
	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON object found in LLM response")
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// LLMs sometimes return numbers/bools for string fields (e.g. "default": 8080).
		// Try lenient parsing: decode into raw map, coerce types, re-marshal.
		var raw map[string]any
		if jsonErr := json.Unmarshal([]byte(jsonStr), &raw); jsonErr == nil {
			if vars, ok := raw["variables"].([]any); ok {
				for _, v := range vars {
					if vm, ok := v.(map[string]any); ok {
						// Coerce "default" to string
						if d, exists := vm["default"]; exists {
							vm["default"] = fmt.Sprintf("%v", d)
						}
					}
				}
			}
			if fixed, marshalErr := json.Marshal(raw); marshalErr == nil {
				if retryErr := json.Unmarshal(fixed, &result); retryErr == nil {
					err = nil
				}
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid JSON in response: %w", err)
		}
	}

	// Normalize step ordering
	for i := range result.Steps {
		if result.Steps[i].Order == 0 {
			result.Steps[i].Order = i + 1
		}
		if result.Steps[i].OnFailure == "" {
			result.Steps[i].OnFailure = "abort"
		}
	}

	return &result, nil
}

// analysisDir returns the path to ~/.mur/session/analysis/.
func analysisDir() (string, error) {
	dir, err := sessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "analysis"), nil
}

// SaveAnalysis persists an AnalysisResult to disk for later use by the web UI.
func SaveAnalysis(sessionID string, result *AnalysisResult) error {
	dir, err := analysisDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create analysis dir: %w", err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal analysis: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, sessionID+".json"), data, 0644)
}

// LoadAnalysis reads a previously saved AnalysisResult from disk.
func LoadAnalysis(sessionID string) (*AnalysisResult, error) {
	dir, err := analysisDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, sessionID+".json"))
	if err != nil {
		return nil, fmt.Errorf("load analysis: %w", err)
	}
	var result AnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse analysis: %w", err)
	}
	return &result, nil
}

// extractJSON finds the first top-level JSON object in a string.
// It handles cases where the LLM wraps JSON in markdown code fences.
func extractJSON(s string) string {
	// Strip markdown code fences if present
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "```json"); idx != -1 {
		s = s[idx+7:]
		if end := strings.Index(s, "```"); end != -1 {
			s = s[:end]
		}
		s = strings.TrimSpace(s)
	} else if idx := strings.Index(s, "```"); idx != -1 {
		s = s[idx+3:]
		if end := strings.Index(s, "```"); end != -1 {
			s = s[:end]
		}
		s = strings.TrimSpace(s)
	}

	// Find the first { and match braces to find the complete object
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		ch := s[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	return ""
}
