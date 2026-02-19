package security

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AnonymizationChange represents a single change made by LLM anonymization.
type AnonymizationChange struct {
	Original string `json:"original"`
	Replaced string `json:"replaced"`
	Category string `json:"category"` // "company", "person", "project", "metric", "location"
	Line     int    `json:"line"`
}

// LLMClient is the interface for LLM providers used by semantic anonymization.
type LLMClient interface {
	Generate(prompt string) (string, error)
}

// SemanticAnonymizer uses an LLM to detect and replace identifying information
// that regex-based scanning cannot catch (company names, person names, etc.).
type SemanticAnonymizer struct {
	llm      LLMClient
	cache    map[string]string // content hash → cleaned content
	cacheDir string            // persistent cache directory
}

// NewSemanticAnonymizer creates a new SemanticAnonymizer with the given LLM client.
// If cacheDir is non-empty, results are persisted to disk.
func NewSemanticAnonymizer(llm LLMClient, cacheDir string) *SemanticAnonymizer {
	return &SemanticAnonymizer{
		llm:      llm,
		cache:    make(map[string]string),
		cacheDir: cacheDir,
	}
}

const anonymizationPrompt = `Analyze this technical pattern and replace ALL identifying information:
- Company/org names → <COMPANY>
- Person names → <PERSON>
- Project codenames → <PROJECT>
- Specific metrics identifying a company → <METRIC>
- Internal jargon/codenames → generic terms
- Location-specific references → <LOCATION>

Keep ALL technical teaching value intact. Only replace identifying parts.
Return ONLY the cleaned text with no additional commentary.

Additionally, after the cleaned text, output a line "---CHANGES---" followed by a JSON array of changes you made. Each change should have these fields:
- "original": the original text
- "replaced": the replacement text
- "category": one of "company", "person", "project", "metric", "location"

If no changes were needed, output "---CHANGES---" followed by "[]".

Content to anonymize:
`

// Anonymize analyzes content with an LLM and replaces identifying information.
func (a *SemanticAnonymizer) Anonymize(content string) (string, []AnonymizationChange, error) {
	hash := contentHash(content)

	// Check in-memory cache
	if cached, ok := a.cache[hash]; ok {
		cleaned, changes := parseLLMResponse(cached, content)
		return cleaned, changes, nil
	}

	// Check disk cache
	if a.cacheDir != "" {
		if cached, err := a.loadFromDisk(hash); err == nil {
			a.cache[hash] = cached
			cleaned, changes := parseLLMResponse(cached, content)
			return cleaned, changes, nil
		}
	}

	// Call LLM
	prompt := anonymizationPrompt + content
	response, err := a.llm.Generate(prompt)
	if err != nil {
		return content, nil, fmt.Errorf("LLM anonymization failed: %w", err)
	}

	// Cache the result
	a.cache[hash] = response
	if a.cacheDir != "" {
		_ = a.saveToDisk(hash, response)
	}

	cleaned, changes := parseLLMResponse(response, content)
	return cleaned, changes, nil
}

// maxDivergenceRatio is the maximum allowed Levenshtein distance ratio between
// cleaned and original text. If the LLM response diverges by more than this
// fraction, it is considered adversarial/malformed and the original is returned.
const maxDivergenceRatio = 0.50

// parseLLMResponse splits the LLM response into cleaned text and change list.
// If the cleaned text differs from the original by more than 50% (Levenshtein
// ratio), it logs a warning and returns the original content unchanged.
func parseLLMResponse(response, originalContent string) (string, []AnonymizationChange) {
	parts := strings.SplitN(response, "---CHANGES---", 2)
	cleaned := strings.TrimSpace(parts[0])

	// Validate that the LLM didn't rewrite the content too aggressively
	if originalContent != "" && cleaned != "" {
		ratio := levenshteinRatio(originalContent, cleaned)
		if ratio > maxDivergenceRatio {
			log.Printf("[anonymize] WARNING: LLM response diverges %.0f%% from original (threshold %.0f%%), returning original",
				ratio*100, maxDivergenceRatio*100)
			return originalContent, nil
		}
	}

	var changes []AnonymizationChange
	if len(parts) == 2 {
		changesJSON := strings.TrimSpace(parts[1])
		_ = json.Unmarshal([]byte(changesJSON), &changes)
	}

	// Populate line numbers from the original content
	for i := range changes {
		if changes[i].Original != "" {
			changes[i].Line = findLineNumber(originalContent, changes[i].Original)
		}
	}

	return cleaned, changes
}

// levenshteinRatio returns the Levenshtein distance between a and b divided by
// the length of the longer string, giving a value between 0.0 (identical) and
// 1.0 (completely different).
func levenshteinRatio(a, b string) float64 {
	dist := levenshteinDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 0
	}
	return float64(dist) / float64(maxLen)
}

// levenshteinDistance computes the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows to save memory
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// findLineNumber returns the 1-indexed line number where needle first appears in content.
func findLineNumber(content, needle string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	return 0
}

// contentHash returns a SHA256 hex digest of content.
func contentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func (a *SemanticAnonymizer) loadFromDisk(hash string) (string, error) {
	path := filepath.Join(a.cacheDir, hash+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *SemanticAnonymizer) saveToDisk(hash, content string) error {
	if err := os.MkdirAll(a.cacheDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(a.cacheDir, hash+".txt")
	return os.WriteFile(path, []byte(content), 0644)
}

// FormatAnonymizationChanges returns a human-readable summary of anonymization changes.
func FormatAnonymizationChanges(changes []AnonymizationChange) string {
	if len(changes) == 0 {
		return "  No semantic PII detected."
	}

	var b strings.Builder
	for _, c := range changes {
		if c.Line > 0 {
			fmt.Fprintf(&b, "  Line %d: [%s] %q → %s\n", c.Line, c.Category, c.Original, c.Replaced)
		} else {
			fmt.Fprintf(&b, "  [%s] %q → %s\n", c.Category, c.Original, c.Replaced)
		}
	}
	return b.String()
}

// OllamaClient implements LLMClient using the Ollama HTTP API.
type OllamaClient struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// ollamaRequest is the request body for Ollama's /api/generate endpoint.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is the response body from Ollama's /api/generate endpoint.
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaClient creates an OllamaClient with sensible defaults.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Generate sends a prompt to Ollama and returns the generated text.
func (c *OllamaClient) Generate(prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/api/generate"
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

// NewLLMClient creates an LLMClient based on provider configuration.
// Supported providers: "ollama". Returns an error for unknown providers.
func NewLLMClient(provider, model, ollamaURL string) (LLMClient, error) {
	switch provider {
	case "ollama":
		return NewOllamaClient(ollamaURL, model), nil
	default:
		return nil, fmt.Errorf("unsupported anonymization provider: %s (supported: ollama)", provider)
	}
}
