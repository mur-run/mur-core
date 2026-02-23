package session

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

// LLMProvider sends a prompt to an LLM and returns the completion text.
type LLMProvider interface {
	Complete(prompt string) (string, error)
}

// NewLLMProviderFromEnv creates an LLMProvider based on environment variables:
//   - MUR_LLM_PROVIDER: "anthropic" (default) or "openai"
//   - MUR_API_KEY: API key for the chosen provider
//   - MUR_MODEL: model name (defaults per provider)
func NewLLMProviderFromEnv() (LLMProvider, error) {
	provider := os.Getenv("MUR_LLM_PROVIDER")
	if provider == "" {
		provider = "anthropic"
	}

	apiKey := os.Getenv("MUR_API_KEY")
	if apiKey == "" {
		// Fall back to provider-specific env vars
		switch provider {
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key: set MUR_API_KEY or %s_API_KEY", strings.ToUpper(provider))
	}

	model := os.Getenv("MUR_MODEL")

	switch provider {
	case "anthropic":
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		return &anthropicProvider{apiKey: apiKey, model: model}, nil
	case "openai":
		if model == "" {
			model = "gpt-4o"
		}
		return &openaiProvider{apiKey: apiKey, model: model}, nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (use 'anthropic' or 'openai')", provider)
	}
}

// anthropicProvider calls the Anthropic Messages API.
type anthropicProvider struct {
	apiKey string
	model  string
}

func (p *anthropicProvider) Complete(prompt string) (string, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic API")
	}

	return result.Content[0].Text, nil
}

// openaiProvider calls the OpenAI Chat Completions API.
type openaiProvider struct {
	apiKey string
	model  string
}

func (p *openaiProvider) Complete(prompt string) (string, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from openai API")
	}

	return result.Choices[0].Message.Content, nil
}
