package learn

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/session"
)

// LLMProvider represents an LLM backend for extraction.
type LLMProvider string

const (
	LLMOllama LLMProvider = "ollama"
	LLMClaude LLMProvider = "claude"
	LLMOpenAI LLMProvider = "openai" // OpenAI-compatible APIs
	LLMGemini LLMProvider = "gemini"
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
const extractionPrompt = `You are a pattern extraction assistant. Your job is to find VALUABLE, NON-OBVIOUS patterns from coding sessions.

## STEP 1: Classify the Session Type

First, determine what type of session this is:
- Q&A SESSION: User asks questions, AI explains concepts → Return []
- TUTORIAL REQUEST: User wants to learn something general → Return []
- ACTIVE DEVELOPMENT: User is building, debugging, or implementing → Proceed to Step 2

If this is a Q&A or Tutorial session, STOP and return an empty array [].

## STEP 2: Extract Patterns (Only for Active Development)

For each pattern, output a JSON object with:
- name: kebab-case identifier (e.g., "swift-async-error-handling")
- title: human-readable title (NOT generic like "How to X")
- confidence: "HIGH", "MEDIUM", or "LOW"
- score: 0.0-1.0 confidence score
- category: "pattern", "lesson", "decision", "template", or "debug"
- domain: "dev", "devops", "mobile", "web", "backend", or "general"
- tags: array of 2-5 relevant tags (e.g., ["swift", "swiftui", "macos", "menubar"])
- trigger_keywords: array of 10-20 trigger keywords for AI agent activation (include: English terms, Chinese translations 繁體中文, abbreviations, common user phrasings in the transcript language)
- problem: the SPECIFIC problem encountered (with error messages if any)
- solution: the solution that WORKED (not generic advice)
- why_non_obvious: why this can't be easily Googled

## MUST EXTRACT (High Value)
✅ User encountered an ERROR → tried multiple approaches → found solution
✅ Discovered WORKAROUND not documented officially
✅ Project-specific CONFIGURATION decision with reasoning
✅ Non-obvious BUG FIX requiring debugging
✅ Integration issues between specific tools/libraries

## MUST NOT EXTRACT (Low Value)
❌ Generic explanations from AI (tutorials, how-to guides)
❌ Code examples that exist on Stack Overflow
❌ Basic language/framework features (async/await basics, etc.)
❌ AI answering "how do I do X?" without actual problem-solving
❌ Theoretical discussions without implementation

## Quality Check
Before including a pattern, ask:
1. Did the user actually encounter this problem? (not just ask about it)
2. Was there debugging or multiple attempts?
3. Is this specific to a project/context? (not generic knowledge)
4. Would a developer NOT find this easily via Google?

If ANY answer is NO, skip that pattern.

## LANGUAGE REQUIREMENT
**IMPORTANT: Output ALL patterns in English, regardless of the transcript language.**
- Translate problem descriptions, solutions, and explanations to English
- Keep code snippets, error messages, and technical terms in their original form
- Preserve technical accuracy while translating

Output a JSON array. If no valuable patterns found, output [].

Example of GOOD pattern (from actual debugging):
[
  {
    "name": "menubarextra-sheet-zstack-workaround",
    "title": "MenuBarExtra Sheet Requires ZStack Overlay",
    "confidence": "HIGH",
    "score": 0.90,
    "category": "debug",
    "domain": "mobile",
    "tags": ["swift", "swiftui", "macos", "menubar", "workaround"],
    "trigger_keywords": ["menubarextra", "sheet", "zstack", "overlay", "popover", "MenuBarExtra", "SwiftUI sheet", "選單列", "彈出視窗", "sheet not showing", "sheet fails"],
    "problem": "SwiftUI .sheet() modifier silently fails in MenuBarExtra popovers - sheets never appear",
    "solution": "Use ZStack with overlay and manual isPresented state instead of .sheet()",
    "why_non_obvious": "Apple docs don't mention this limitation. Error is silent - no console output."
  }
]

Example of BAD pattern (generic Q&A - DO NOT EXTRACT):
User: "How do I test async in Swift?"
AI: "Add async to your test method..."
→ This is just a tutorial. Return []`

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

	// Create unified LLM provider
	provider, err := llmProviderFromOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("LLM setup failed: %w", err)
	}

	// Compose full prompt with extraction instructions + transcript
	fullPrompt := extractionPrompt + "\n\n---\n\nExtract patterns from this coding session:\n\n" + text

	response, err := provider.Complete(fullPrompt)
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

		// Merge tags and trigger_keywords (deduplicated)
		mergedTags := deduplicateTags(jp.Tags, jp.TriggerKeywords)

		pattern := Pattern{
			Name:        jp.Name,
			Description: jp.Title,
			Content:     content,
			Domain:      domain,
			Category:    category,
			Tags:        mergedTags,
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

// llmProviderFromOptions converts LLMExtractOptions to a session.LLMProvider.
func llmProviderFromOptions(opts LLMExtractOptions) (session.LLMProvider, error) {
	switch opts.Provider {
	case LLMOllama:
		baseURL := opts.OllamaURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return session.NewLLMProvider("ollama", opts.Model, "", baseURL)
	case LLMClaude:
		model := opts.Model
		if model == "" || model == "llama3.2" {
			model = "claude-sonnet-4-20250514"
		}
		return session.NewLLMProvider("anthropic", model, opts.ClaudeKey, "")
	case LLMOpenAI:
		model := opts.Model
		if model == "" || model == "llama3.2" {
			model = "gpt-4o"
		}
		return session.NewLLMProvider("openai", model, opts.OpenAIKey, opts.OpenAIURL)
	case LLMGemini:
		model := opts.Model
		if model == "" || model == "llama3.2" {
			model = "gemini-2.0-flash"
		}
		return session.NewLLMProvider("gemini", model, opts.GeminiKey, "")
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", opts.Provider)
	}
}
