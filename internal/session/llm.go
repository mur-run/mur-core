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

	"github.com/mur-run/mur-core/internal/config"
)

// LLMProvider sends a prompt to an LLM and returns the completion text.
type LLMProvider interface {
	Complete(prompt string) (string, error)
}

// fallbackProvider wraps a primary and fallback LLMProvider. If the primary
// fails, it automatically retries with the fallback provider.
type fallbackProvider struct {
	primary     LLMProvider
	fallback    LLMProvider
	primaryName string
}

func (f *fallbackProvider) Complete(prompt string) (string, error) {
	result, err := f.primary.Complete(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ %s failed (%v), falling back...\n", f.primaryName, err)
		return f.fallback.Complete(prompt)
	}
	return result, nil
}

// NewLLMProvider creates an LLMProvider from simple parameters (no config needed).
// For ollama: apiKey is ignored, baseURL is the Ollama URL.
// For openai: baseURL is the OpenAI-compatible API URL (empty = default).
// For anthropic/gemini: baseURL is ignored.
func NewLLMProvider(provider, model, apiKey, baseURL string) (LLMProvider, error) {
	switch provider {
	case "claude", "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("no Anthropic API key provided")
		}
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		return &anthropicProvider{apiKey: apiKey, model: model}, nil
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("no OpenAI API key provided")
		}
		if model == "" {
			model = "gpt-4o"
		}
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return &openaiProvider{apiKey: apiKey, model: model, baseURL: baseURL}, nil
	case "ollama":
		if model == "" {
			model = "llama3.2:3b"
		}
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return &ollamaProvider{model: model, baseURL: baseURL}, nil
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("no Gemini API key provided")
		}
		if model == "" {
			model = "gemini-2.0-flash"
		}
		return &geminiProvider{apiKey: apiKey, model: model}, nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (use: anthropic, openai, ollama, gemini)", provider)
	}
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
		return &openaiProvider{apiKey: apiKey, model: model, baseURL: "https://api.openai.com/v1"}, nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (use 'anthropic' or 'openai')", provider)
	}
}

// NewLLMProviderFromConfig creates an LLMProvider from config.yaml settings.
// Falls back to NewLLMProviderFromEnv if config has no provider set.
func NewLLMProviderFromConfig(cfg *config.Config) (LLMProvider, error) {
	return NewLLMProviderWithOverrides(cfg, "", "", "")
}

// NewLLMProviderWithOverrides creates an LLMProvider from config with CLI flag overrides.
// Empty override values are ignored and config/defaults are used instead.
// If a premium provider is configured, wraps with fallback support.
func NewLLMProviderWithOverrides(cfg *config.Config, provider, model, ollamaURL string) (LLMProvider, error) {
	llmCfg := cfg.Learning.LLM

	// Apply overrides
	if provider != "" {
		llmCfg.Provider = provider
	}
	if model != "" {
		llmCfg.Model = model
	}
	if ollamaURL != "" {
		llmCfg.OllamaURL = ollamaURL
	}

	// If config has no provider, fall back to env-based constructor
	if llmCfg.Provider == "" {
		return NewLLMProviderFromEnv()
	}

	primary, err := newProviderFromLLMConfig(llmCfg)
	if err != nil {
		return nil, err
	}

	// If premium provider is configured and no CLI overrides were given,
	// wrap with fallback so primary failures automatically retry with premium.
	if cfg.Learning.LLM.Premium != nil && provider == "" {
		premiumCfg := config.LLMConfig{
			Provider:  cfg.Learning.LLM.Premium.Provider,
			Model:     cfg.Learning.LLM.Premium.Model,
			OllamaURL: cfg.Learning.LLM.Premium.OllamaURL,
			OpenAIURL: cfg.Learning.LLM.Premium.OpenAIURL,
			APIKeyEnv: cfg.Learning.LLM.Premium.APIKeyEnv,
		}
		fb, fbErr := newProviderFromLLMConfig(premiumCfg)
		if fbErr == nil {
			return &fallbackProvider{
				primary:     primary,
				fallback:    fb,
				primaryName: llmCfg.Provider,
			}, nil
		}
		// If premium setup fails, just use primary without fallback
	}

	return primary, nil
}

// newProviderFromLLMConfig creates an LLMProvider from an LLMConfig.
func newProviderFromLLMConfig(llmCfg config.LLMConfig) (LLMProvider, error) {
	switch llmCfg.Provider {
	case "claude", "anthropic":
		return newAnthropicFromConfig(llmCfg)
	case "openai":
		return newOpenAIFromConfig(llmCfg)
	case "ollama":
		return newOllamaFromConfig(llmCfg)
	case "gemini":
		return newGeminiFromConfig(llmCfg)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (use: anthropic, openai, ollama, gemini)", llmCfg.Provider)
	}
}

// resolveAPIKey looks up the API key: first from a custom env var name (APIKeyEnv),
// then from the standard env var for the provider.
func resolveAPIKey(apiKeyEnv, standardEnv string) string {
	if apiKeyEnv != "" {
		if key := os.Getenv(apiKeyEnv); key != "" {
			return key
		}
	}
	return os.Getenv(standardEnv)
}

func newAnthropicFromConfig(cfg config.LLMConfig) (LLMProvider, error) {
	apiKey := resolveAPIKey(cfg.APIKeyEnv, "ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("no Anthropic API key: set ANTHROPIC_API_KEY or configure learning.llm.api_key_env")
	}
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &anthropicProvider{apiKey: apiKey, model: model}, nil
}

func newOpenAIFromConfig(cfg config.LLMConfig) (LLMProvider, error) {
	apiKey := resolveAPIKey(cfg.APIKeyEnv, "OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("no OpenAI API key: set OPENAI_API_KEY or configure learning.llm.api_key_env")
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	baseURL := cfg.OpenAIURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &openaiProvider{apiKey: apiKey, model: model, baseURL: baseURL}, nil
}

func newOllamaFromConfig(cfg config.LLMConfig) (LLMProvider, error) {
	model := cfg.Model
	if model == "" {
		model = "llama3.2:3b"
	}
	ollamaURL := cfg.OllamaURL
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	return &ollamaProvider{model: model, baseURL: ollamaURL}, nil
}

func newGeminiFromConfig(cfg config.LLMConfig) (LLMProvider, error) {
	apiKey := resolveAPIKey(cfg.APIKeyEnv, "GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("no Gemini API key: set GEMINI_API_KEY or configure learning.llm.api_key_env")
	}
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &geminiProvider{apiKey: apiKey, model: model}, nil
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

// openaiProvider calls the OpenAI Chat Completions API (or any compatible endpoint).
type openaiProvider struct {
	apiKey  string
	model   string
	baseURL string
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

	url := strings.TrimSuffix(p.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
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

// ollamaProvider calls the Ollama generate API.
type ollamaProvider struct {
	model   string
	baseURL string
}

func (p *ollamaProvider) Complete(prompt string) (string, error) {
	body := map[string]any{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]any{
			"temperature": 0.3,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.baseURL, "/") + "/api/generate"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.Response, nil
}

// geminiProvider calls the Google Gemini API.
type geminiProvider struct {
	apiKey string
	model  string
}

func (p *geminiProvider) Complete(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	body := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.3,
			"maxOutputTokens": 4096,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, string(respBody))
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
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from gemini API")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
