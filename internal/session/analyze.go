package session

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnalysisResult is the structured workflow extracted from a session transcript.
type AnalysisResult struct {
	Name        string     `json:"name"`
	Trigger     string     `json:"trigger"`
	Description string     `json:"description"`
	Variables   []Variable `json:"variables"`
	Steps       []Step     `json:"steps"`
	Tools       []string   `json:"tools"`
	Tags        []string   `json:"tags"`
}

// Variable represents a parameterizable value in a workflow.
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, path, url, number, bool
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

// Step represents a single action in a workflow.
type Step struct {
	Order         int    `json:"order"`
	Description   string `json:"description"`
	Command       string `json:"command,omitempty"`
	Tool          string `json:"tool,omitempty"`
	NeedsApproval bool   `json:"needs_approval"`
	OnFailure     string `json:"on_failure"` // skip, abort, retry
}

// qaCoTPrompt is the Question-Answer Chain of Thought prompt for analysis.
const qaCoTPrompt = `Analyze this AI conversation transcript and extract a reusable workflow.

Answer each question step by step:

Q1: What was the user's initial problem or goal?
Q2: What was the root cause (if debugging)?
Q3: What steps were attempted? Which succeeded, which failed?
Q4: What is the minimal correct sequence of steps to solve this?
Q5: What tools or commands were used at each step?
Q6: Which values are environment-specific and should be variables?
Q7: Are there conditional branches (if X then Y)?
Q8: Which steps need human approval before proceeding?
Q9: What's a good name and trigger description for this workflow?
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

// formatTranscript converts EventRecords into a readable text transcript.
func formatTranscript(events []EventRecord) string {
	var b strings.Builder

	for _, e := range events {
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
		return nil, fmt.Errorf("invalid JSON in response: %w", err)
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
