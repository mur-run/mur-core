// Package embed provides embedding-based semantic search for patterns.
package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Vector represents an embedding vector.
type Vector []float64

// Embedder is the interface for embedding providers.
type Embedder interface {
	Embed(text string) (Vector, error)
	EmbedBatch(texts []string) ([]Vector, error)
	Dimension() int
	Name() string
}

// Config holds embedding configuration.
type Config struct {
	// Provider: "openai", "ollama", "local"
	Provider string `yaml:"provider"`
	// Model name (e.g., "text-embedding-3-small", "nomic-embed-text")
	Model string `yaml:"model"`
	// API endpoint (for ollama/local)
	Endpoint string `yaml:"endpoint,omitempty"`
	// API key (for openai)
	APIKey string `yaml:"api_key,omitempty"`
}

// DefaultConfig returns the default embedding config.
func DefaultConfig() Config {
	return Config{
		Provider: "ollama",
		Model:    "nomic-embed-text",
		Endpoint: "http://localhost:11434",
	}
}

// NewEmbedder creates an embedder from config.
func NewEmbedder(cfg Config) (Embedder, error) {
	switch cfg.Provider {
	case "openai":
		apiKey := cfg.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key required: set OPENAI_API_KEY env var")
		}
		return NewOpenAIEmbedder(apiKey, cfg.Model), nil

	case "voyage":
		apiKey := cfg.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("VOYAGE_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("Voyage API key required: set VOYAGE_API_KEY env var")
		}
		e := NewOpenAIEmbedder(apiKey, cfg.Model)
		e.baseURL = "https://api.voyageai.com/v1"
		return e, nil

	case "google":
		apiKey := cfg.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("Google API key required: set GEMINI_API_KEY env var")
		}
		e := NewOpenAIEmbedder(apiKey, cfg.Model)
		e.baseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
		return e, nil

	case "ollama":
		endpoint := cfg.Endpoint
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		return NewOllamaEmbedder(endpoint, cfg.Model), nil

	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", cfg.Provider)
	}
}

// ============================================================
// OpenAI Embedder
// ============================================================

// OpenAIEmbedder uses OpenAI's embedding API.
type OpenAIEmbedder struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAIEmbedder creates an OpenAI embedder.
func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &OpenAIEmbedder{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *OpenAIEmbedder) Name() string { return "openai" }

func (e *OpenAIEmbedder) Dimension() int {
	switch e.model {
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-ada-002":
		return 1536
	default:
		return 1536
	}
}

type openAIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (e *OpenAIEmbedder) Embed(text string) (Vector, error) {
	vectors, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (e *OpenAIEmbedder) EmbedBatch(texts []string) ([]Vector, error) {
	reqBody, _ := json.Marshal(openAIRequest{
		Model: e.model,
		Input: texts,
	})

	req, _ := http.NewRequest("POST", e.baseURL+"/embeddings", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", result.Error.Message)
	}

	vectors := make([]Vector, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}

	return vectors, nil
}

// ============================================================
// Ollama Embedder
// ============================================================

// OllamaEmbedder uses Ollama's local embedding API.
type OllamaEmbedder struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewOllamaEmbedder creates an Ollama embedder.
func NewOllamaEmbedder(endpoint, model string) *OllamaEmbedder {
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaEmbedder{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *OllamaEmbedder) Name() string { return "ollama" }

func (e *OllamaEmbedder) Dimension() int {
	switch {
	case strings.Contains(e.model, "mxbai-embed-large"):
		return 1024
	case strings.Contains(e.model, "nomic-embed"):
		return 768
	case strings.Contains(e.model, "all-minilm"):
		return 384
	default:
		return 768
	}
}

// QueryPrefix returns the prefix needed for retrieval queries.
// Some models (like mxbai-embed-large) require a specific prefix for queries
// but NOT for documents being indexed.
func (e *OllamaEmbedder) QueryPrefix() string {
	if strings.Contains(e.model, "mxbai-embed-large") {
		return "Represent this sentence for searching relevant passages: "
	}
	return ""
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

func (e *OllamaEmbedder) Embed(text string) (Vector, error) {
	reqBody, _ := json.Marshal(ollamaRequest{
		Model:  e.model,
		Prompt: text,
	})

	resp, err := e.client.Post(e.endpoint+"/api/embeddings", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	var result ollamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", result.Error)
	}

	return result.Embedding, nil
}

func (e *OllamaEmbedder) EmbedBatch(texts []string) ([]Vector, error) {
	// Ollama doesn't support batch, do sequentially
	vectors := make([]Vector, len(texts))
	for i, text := range texts {
		v, err := e.Embed(text)
		if err != nil {
			return nil, err
		}
		vectors[i] = v
	}
	return vectors, nil
}

// ============================================================
// Vector Operations
// ============================================================

// CosineSimilarity calculates cosine similarity between two vectors.
func CosineSimilarity(a, b Vector) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ============================================================
// Embedding Cache
// ============================================================

// Cache stores embeddings for patterns.
type Cache struct {
	dir      string
	embedder Embedder
	mu       sync.RWMutex
	cache    map[string]Vector
}

// CacheEntry represents a cached embedding.
type CacheEntry struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Vector    Vector    `json:"vector"`
	Model     string    `json:"model"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewCache creates a new embedding cache.
func NewCache(dir string, embedder Embedder) *Cache {
	return &Cache{
		dir:      dir,
		embedder: embedder,
		cache:    make(map[string]Vector),
	}
}

// cacheFile returns the path to the cache file.
func (c *Cache) cacheFile() string {
	return filepath.Join(c.dir, "embeddings.json")
}

// Load loads the cache from disk.
func (c *Cache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.cacheFile()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var entries []CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	for _, e := range entries {
		c.cache[e.ID] = e.Vector
	}

	return nil
}

// Save saves the cache to disk.
func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}

	entries := make([]CacheEntry, 0, len(c.cache))
	for id, vec := range c.cache {
		entries = append(entries, CacheEntry{
			ID:        id,
			Vector:    vec,
			Model:     c.embedder.Name(),
			UpdatedAt: time.Now(),
		})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return os.WriteFile(c.cacheFile(), data, 0644)
}

// Get returns a cached embedding.
func (c *Cache) Get(id string) (Vector, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.cache[id]
	return v, ok
}

// Set stores an embedding in the cache.
func (c *Cache) Set(id string, vec Vector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = vec
}

// GetOrEmbed gets from cache or embeds the text.
func (c *Cache) GetOrEmbed(id, text string) (Vector, error) {
	if v, ok := c.Get(id); ok {
		return v, nil
	}

	v, err := c.embedder.Embed(text)
	if err != nil {
		return nil, err
	}

	c.Set(id, v)
	return v, nil
}

// Search finds the most similar entries to the query.
func (c *Cache) Search(query Vector, topK int) []SearchResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]SearchResult, 0, len(c.cache))
	for id, vec := range c.cache {
		score := CosineSimilarity(query, vec)
		results = append(results, SearchResult{
			ID:    id,
			Score: score,
		})
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if topK > 0 && topK < len(results) {
		results = results[:topK]
	}

	return results
}

// SearchResult represents a search result.
type SearchResult struct {
	ID    string
	Score float64
}
