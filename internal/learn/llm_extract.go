package learn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LLMProvider represents an LLM backend for extraction.
type LLMProvider string

const (
	LLMOllama   LLMProvider = "ollama"
	LLMClaude   LLMProvider = "claude"
	LLMOpenAI   LLMProvider = "openai"   // OpenAI-compatible APIs
	LLMGemini   LLMProvider = "gemini"
)

// LLMExtractOptions configures LLM-based extraction.
type LLMExtractOptions struct {
	Provider    LLMProvider
	Model       string // e.g., "llama3.2" for Ollama, "claude-sonnet-4-20250514" for Claude
	OllamaURL   string // default: http://localhost:11434
	ClaudeKey   string // from env ANTHROPIC_API_KEY
	OpenAIKey   string // from env OPENAI_API_KEY
	OpenAIURL   string // default: https://api.openai.com/v1 (or any compatible endpoint)
	GeminiKey   string // from env GEMINI_API_KEY
	MaxPatterns int    // max patterns to extract per session
}

// DefaultLLMOptions returns sensible defaults.
func DefaultLLMOptions() LLMExtractOptions {
	return LLMExtractOptions{
		Provider:    LLMOllama,
		Model:       "llama3.2",
		OllamaURL:   "http://localhost:11434",
		ClaudeKey:   os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIURL:   "https://api.openai.com/v1",
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		MaxPatterns: 10,
	}
}

// extractionPrompt is the system prompt for pattern extraction.
const extractionPrompt = `You are a pattern extraction assistant. Analyze the following coding session transcript and extract reusable patterns, lessons learned, and best practices.

For each pattern you find, output a JSON object with these fields:
- name: kebab-case identifier (e.g., "swift-async-error-handling")
- title: human-readable title
- confidence: "HIGH", "MEDIUM", or "LOW"
- score: 0.0-1.0 confidence score
- category: "pattern", "lesson", "decision", "template", or "debug"
- domain: "dev", "devops", "mobile", "web", "backend", or "general"
- problem: what problem this pattern solves
- solution: the solution or pattern
- why_non_obvious: why this is worth remembering (not obvious)

Only extract patterns that are:
1. Reusable across projects
2. Non-obvious (not basic/common knowledge)
3. Specific enough to be actionable

Output a JSON array of patterns. If no good patterns found, output an empty array [].

Example output:
[
  {
    "name": "swiftui-sheet-menubar-workaround",
    "title": "SwiftUI Sheet in MenuBarExtra Workaround",
    "confidence": "HIGH",
    "score": 0.85,
    "category": "pattern",
    "domain": "mobile",
    "problem": "SwiftUI .sheet() doesn't work in MenuBarExtra popovers",
    "solution": "Use ZStack overlay pattern with manual state management",
    "why_non_obvious": "Apple docs don't mention this limitation; failure is silent"
  }
]`

// ExtractWithLLM uses an LLM to extract patterns from a session.
func ExtractWithLLM(session *Session, opts LLMExtractOptions) ([]ExtractedPattern, error) {
	// Build transcript text
	var transcript strings.Builder
	transcript.WriteString(fmt.Sprintf("Project: %s\n\n", session.Project))

	for _, msg := range session.Messages {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		transcript.WriteString(fmt.Sprintf("### %s:\n%s\n\n", role, msg.Content))
	}

	// Truncate if too long (keep last 20k chars for context)
	text := transcript.String()
	if len(text) > 20000 {
		text = text[len(text)-20000:]
	}

	// Call LLM
	var response string
	var err error

	switch opts.Provider {
	case LLMOllama:
		response, err = callOllama(text, opts)
	case LLMClaude:
		response, err = callClaude(text, opts)
	case LLMOpenAI:
		response, err = callOpenAI(text, opts)
	case LLMGemini:
		response, err = callGemini(text, opts)
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", opts.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse JSON patterns from response
	patterns := extractJSONPatterns(response, session.ShortID())

	// Also try to parse if the response itself is a JSON array
	if len(patterns) == 0 {
		patterns = parseJSONArray(response, session.ShortID())
	}

	return patterns, nil
}

// parseJSONArray tries to parse the response as a direct JSON array.
func parseJSONArray(text string, sourceID string) []ExtractedPattern {
	var extracted []ExtractedPattern

	// Clean up the response - find JSON array
	text = strings.TrimSpace(text)

	// Try to find JSON array bounds
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end <= start {
		return nil
	}

	jsonStr := text[start : end+1]

	var jsonPatterns []JSONPattern
	if err := json.Unmarshal([]byte(jsonStr), &jsonPatterns); err != nil {
		return nil
	}

	for _, jp := range jsonPatterns {
		if jp.Name == "" || !isValidPatternName(jp.Name) {
			continue
		}

		// Build content
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
		if jp.WhyNonObvious != "" {
			contentParts = append(contentParts, "\n## Why Non-Obvious\n"+jp.WhyNonObvious)
		}

		content := strings.Join(contentParts, "\n")
		if content == "" {
			continue
		}

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

		domain := jp.Domain
		if domain == "" || domain == "mobile" {
			domain = "dev"
		}

		category := jp.Category
		if category == "" {
			category = "pattern"
		}

		pattern := Pattern{
			Name:        jp.Name,
			Description: jp.Title,
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

	return extracted
}

// callOllama calls the Ollama API.
func callOllama(transcript string, opts LLMExtractOptions) (string, error) {
	url := opts.OllamaURL + "/api/generate"

	prompt := extractionPrompt + "\n\n---\n\nTranscript:\n" + transcript

	payload := map[string]interface{}{
		"model":  opts.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.3,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	return result.Response, nil
}

// callClaude calls the Anthropic Claude API.
func callClaude(transcript string, opts LLMExtractOptions) (string, error) {
	if opts.ClaudeKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	url := "https://api.anthropic.com/v1/messages"

	model := opts.Model
	if model == "" || model == "llama3.2" {
		model = "claude-sonnet-4-20250514"
	}

	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"system":     extractionPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": "Extract patterns from this coding session:\n\n" + transcript},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", opts.ClaudeKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse claude response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}

	return result.Content[0].Text, nil
}

// CheckOllamaAvailable checks if Ollama is running.
func CheckOllamaAvailable(url string) bool {
	if url == "" {
		url = "http://localhost:11434"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// callOpenAI calls any OpenAI-compatible API (OpenAI, Groq, Together, etc.).
func callOpenAI(transcript string, opts LLMExtractOptions) (string, error) {
	if opts.OpenAIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	url := opts.OpenAIURL
	if url == "" {
		url = "https://api.openai.com/v1"
	}
	url = strings.TrimSuffix(url, "/") + "/chat/completions"

	model := opts.Model
	if model == "" || model == "llama3.2" {
		model = "gpt-4o"
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": extractionPrompt},
			{"role": "user", "content": "Extract patterns from this coding session:\n\n" + transcript},
		},
		"temperature": 0.3,
		"max_tokens":  4096,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opts.OpenAIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from openai")
	}

	return result.Choices[0].Message.Content, nil
}

// callGemini calls the Google Gemini API.
func callGemini(transcript string, opts LLMExtractOptions) (string, error) {
	if opts.GeminiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	model := opts.Model
	if model == "" || model == "llama3.2" {
		model = "gemini-2.0-flash"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, opts.GeminiKey)

	payload := map[string]interface{}{
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": extractionPrompt},
			},
		},
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": "Extract patterns from this coding session:\n\n" + transcript},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.3,
			"maxOutputTokens": 4096,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse gemini response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from gemini")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
