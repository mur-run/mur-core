package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
)

// ExpandedQueries stores LLM-generated search queries for patterns.
// This is a sidecar file alongside the embedding cache.
type ExpandedQueries struct {
	Queries map[string][]string `json:"queries"` // pattern name → generated queries
	path    string
}

// LoadExpandedQueries loads or creates the expanded queries file.
func LoadExpandedQueries(cacheDir string) *ExpandedQueries {
	path := filepath.Join(cacheDir, "expanded_queries.json")
	eq := &ExpandedQueries{
		Queries: make(map[string][]string),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, eq)
	}
	return eq
}

// Save persists the expanded queries to disk.
func (eq *ExpandedQueries) Save() error {
	data, err := json.MarshalIndent(eq, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(eq.path, data, 0644)
}

// Get returns expanded queries for a pattern, or nil if not yet generated.
func (eq *ExpandedQueries) Get(name string) []string {
	return eq.Queries[name]
}

// GenerateForPattern uses a local LLM to generate likely search queries.
func (eq *ExpandedQueries) GenerateForPattern(p pattern.Pattern, ollamaURL, model string) error {
	// Build a concise summary for the LLM
	summary := fmt.Sprintf("Name: %s\nDescription: %s\nTags: %s",
		p.Name, p.Description, strings.Join(p.Tags.Confirmed, ", "))

	content := p.Content
	if len(content) > 500 {
		content = content[:500]
	}
	if content != "" {
		summary += "\nContent: " + content
	}

	prompt := fmt.Sprintf(`Generate 5 search queries for this pattern. One per line, no numbering, no explanation.

%s

Queries:`, summary)

	payload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 200,
		},
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse lines into queries
	var queries []string
	for _, line := range strings.Split(result.Response, "\n") {
		line = strings.TrimSpace(line)
		// Remove numbering like "1. " or "- "
		line = strings.TrimLeft(line, "0123456789.-) ")
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 3 && len(line) < 100 {
			queries = append(queries, strings.ToLower(line))
		}
	}

	if len(queries) > 7 {
		queries = queries[:7]
	}

	eq.Queries[p.Name] = queries
	return nil
}
