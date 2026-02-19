package cache

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"sync"
)

// EmbeddingMatrix stores embedding vectors in contiguous float32 memory
// with pre-normalized copies for fast cosine similarity via dot product.
type EmbeddingMatrix struct {
	mu     sync.RWMutex
	data   []float32 // raw vectors: [p0_d0, p0_d1, ..., p1_d0, ...]
	normed []float32 // pre-normalized copy (unit vectors)
	ids    []string  // pattern ID at each row
	dim    int       // dimensionality of each vector
	n      int       // number of vectors (rows)
	loaded bool      // whether embeddings have been loaded
}

// NewEmbeddingMatrix creates an empty matrix with the given dimensionality.
func NewEmbeddingMatrix(dim int) *EmbeddingMatrix {
	return &EmbeddingMatrix{
		dim: dim,
	}
}

// embeddingCacheEntry mirrors the JSON structure used by embed.Cache.
type embeddingCacheEntry struct {
	ID     string    `json:"id"`
	Vector []float64 `json:"vector"`
}

// Load reads the embeddings.json cache file and packs all vectors into
// contiguous float32 storage. It also pre-normalizes each vector.
func (m *EmbeddingMatrix) Load(cacheFile string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(cacheFile)
	if os.IsNotExist(err) {
		m.loaded = true
		return nil // no cache yet
	}
	if err != nil {
		return err
	}

	var entries []embeddingCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	if len(entries) == 0 {
		m.loaded = true
		return nil
	}

	// Determine dimension from first entry
	dim := len(entries[0].Vector)
	if dim == 0 {
		m.loaded = true
		return nil
	}
	m.dim = dim

	n := len(entries)
	m.n = n
	m.ids = make([]string, n)
	m.data = make([]float32, n*dim)
	m.normed = make([]float32, n*dim)

	for i, e := range entries {
		m.ids[i] = e.ID
		off := i * dim
		for j := 0; j < dim && j < len(e.Vector); j++ {
			m.data[off+j] = float32(e.Vector[j])
		}
		// Normalize
		m.normalizeRow(off, dim)
	}

	m.loaded = true
	return nil
}

// normalizeRow normalizes data[off:off+dim] into normed[off:off+dim].
func (m *EmbeddingMatrix) normalizeRow(off, dim int) {
	var sumSq float32
	for j := 0; j < dim; j++ {
		v := m.data[off+j]
		sumSq += v * v
	}
	norm := float32(math.Sqrt(float64(sumSq)))
	if norm == 0 {
		// Copy zeros
		for j := 0; j < dim; j++ {
			m.normed[off+j] = 0
		}
		return
	}
	invNorm := 1.0 / norm
	for j := 0; j < dim; j++ {
		m.normed[off+j] = m.data[off+j] * invNorm
	}
}

// SearchResult holds a pattern ID and its similarity score.
type MatrixSearchResult struct {
	ID    string
	Score float64
}

// Search finds the topK most similar vectors to queryVec using dot product
// on pre-normalized vectors (equivalent to cosine similarity).
// The queryVec is float64 to match the embed.Vector type.
func (m *EmbeddingMatrix) Search(queryVec []float64, topK int) []MatrixSearchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.n == 0 || len(queryVec) == 0 {
		return nil
	}

	// Normalize query to float32
	dim := m.dim
	qNormed := make([]float32, dim)
	var qSumSq float32
	for j := 0; j < dim && j < len(queryVec); j++ {
		v := float32(queryVec[j])
		qNormed[j] = v
		qSumSq += v * v
	}
	if qSumSq == 0 {
		return nil
	}
	qInvNorm := float32(1.0 / math.Sqrt(float64(qSumSq)))
	for j := 0; j < dim; j++ {
		qNormed[j] *= qInvNorm
	}

	// Compute dot products (cosine similarity on unit vectors)
	scores := make([]MatrixSearchResult, m.n)
	for i := 0; i < m.n; i++ {
		off := i * dim
		var dot float32
		for j := 0; j < dim; j++ {
			dot += qNormed[j] * m.normed[off+j]
		}
		scores[i] = MatrixSearchResult{
			ID:    m.ids[i],
			Score: float64(dot),
		}
	}

	// Partial sort for topK
	sort.Slice(scores, func(a, b int) bool {
		return scores[a].Score > scores[b].Score
	})

	if topK > 0 && topK < len(scores) {
		scores = scores[:topK]
	}

	return scores
}

// SimilarityPair holds a pair of pattern IDs and their cosine similarity.
type SimilarityPair struct {
	IDA        string
	IDB        string
	Similarity float64
}

// AllPairs returns all pairs of patterns with cosine similarity >= threshold.
func (m *EmbeddingMatrix) AllPairs(threshold float64) []SimilarityPair {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.n < 2 {
		return nil
	}

	var pairs []SimilarityPair
	dim := m.dim

	for i := 0; i < m.n; i++ {
		offI := i * dim
		for j := i + 1; j < m.n; j++ {
			offJ := j * dim
			var dot float32
			for k := 0; k < dim; k++ {
				dot += m.normed[offI+k] * m.normed[offJ+k]
			}
			sim := float64(dot)
			if sim >= threshold {
				pairs = append(pairs, SimilarityPair{
					IDA:        m.ids[i],
					IDB:        m.ids[j],
					Similarity: sim,
				})
			}
		}
	}

	return pairs
}

// MaxSimilarity returns the maximum cosine similarity between the given pattern
// and any other pattern in the matrix. Returns -1 if the pattern is not found.
func (m *EmbeddingMatrix) MaxSimilarity(patternID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx := -1
	for i, id := range m.ids {
		if id == patternID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return -1
	}

	dim := m.dim
	offI := idx * dim
	maxSim := -1.0

	for j := 0; j < m.n; j++ {
		if j == idx {
			continue
		}
		offJ := j * dim
		var dot float32
		for k := 0; k < dim; k++ {
			dot += m.normed[offI+k] * m.normed[offJ+k]
		}
		if float64(dot) > maxSim {
			maxSim = float64(dot)
		}
	}

	return maxSim
}

// Len returns the number of vectors in the matrix.
func (m *EmbeddingMatrix) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.n
}

// Dim returns the vector dimensionality.
func (m *EmbeddingMatrix) Dim() int {
	return m.dim
}

// IsLoaded returns whether the matrix has been loaded.
func (m *EmbeddingMatrix) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loaded
}
