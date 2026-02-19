package consolidate

import (
	"sort"

	"github.com/mur-run/mur-core/internal/cache"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

// MergeStrategy represents how duplicate patterns should be merged.
type MergeStrategy string

const (
	StrategyKeepBest MergeStrategy = "keep-best"
	StrategyLLMMerge MergeStrategy = "llm-merge"
	StrategyManual   MergeStrategy = "manual"
)

// MergeProposal represents a proposed merge of duplicate patterns.
type MergeProposal struct {
	Patterns   []*pattern.Pattern `json:"patterns"`
	Similarity float64            `json:"similarity"`
	Strategy   MergeStrategy      `json:"strategy"`
	KeepID     string             `json:"keep_id,omitempty"`    // ID of pattern to keep (for keep-best)
	RemoveIDs  []string           `json:"remove_ids,omitempty"` // IDs of patterns to remove
}

// DuplicateDetector finds duplicate patterns using embedding similarity.
type DuplicateDetector struct {
	matrix    *cache.EmbeddingMatrix
	threshold float64
	strategy  MergeStrategy
	scores    map[string]*HealthScore // pattern ID → health score
}

// NewDuplicateDetector creates a detector with the given threshold and strategy.
func NewDuplicateDetector(matrix *cache.EmbeddingMatrix, threshold float64, strategy MergeStrategy) *DuplicateDetector {
	return &DuplicateDetector{
		matrix:    matrix,
		threshold: threshold,
		strategy:  strategy,
		scores:    make(map[string]*HealthScore),
	}
}

// WithHealthScores attaches health scores for keep-best strategy decisions.
func (d *DuplicateDetector) WithHealthScores(scores []HealthScore) {
	for i := range scores {
		d.scores[scores[i].PatternID] = &scores[i]
	}
}

// Detect finds all duplicate pairs above the similarity threshold.
func (d *DuplicateDetector) Detect(patterns []*pattern.Pattern) []MergeProposal {
	if d.matrix == nil || !d.matrix.IsLoaded() || d.matrix.Len() < 2 {
		return nil
	}

	// Get all pairwise similarities above threshold
	pairs := d.matrix.AllPairs(d.threshold)

	// Build a map of pattern IDs to patterns
	patternMap := make(map[string]*pattern.Pattern, len(patterns))
	for _, p := range patterns {
		patternMap[p.ID] = p
	}

	// Group overlapping pairs into clusters using union-find
	uf := newUnionFind(len(patterns))
	idxMap := make(map[string]int, len(patterns))
	for i, p := range patterns {
		idxMap[p.ID] = i
	}

	for _, pair := range pairs {
		idxA, okA := idxMap[pair.IDA]
		idxB, okB := idxMap[pair.IDB]
		if okA && okB {
			uf.union(idxA, idxB)
		}
	}

	// Collect clusters
	clusters := make(map[int][]int)
	for i := range patterns {
		root := uf.find(i)
		clusters[root] = append(clusters[root], i)
	}

	// Build merge proposals for clusters with 2+ members
	var proposals []MergeProposal
	for _, members := range clusters {
		if len(members) < 2 {
			continue
		}

		clusterPatterns := make([]*pattern.Pattern, len(members))
		for i, idx := range members {
			clusterPatterns[i] = patterns[idx]
		}

		// Find max pairwise similarity within cluster
		maxSim := 0.0
		for _, pair := range pairs {
			_, inA := idxMap[pair.IDA]
			_, inB := idxMap[pair.IDB]
			if inA && inB && pair.Similarity > maxSim {
				maxSim = pair.Similarity
			}
		}

		proposal := MergeProposal{
			Patterns:   clusterPatterns,
			Similarity: maxSim,
			Strategy:   d.strategy,
		}

		if d.strategy == StrategyKeepBest {
			d.selectBest(&proposal)
		}

		proposals = append(proposals, proposal)
	}

	// Sort proposals by similarity (highest first)
	sort.Slice(proposals, func(i, j int) bool {
		return proposals[i].Similarity > proposals[j].Similarity
	})

	return proposals
}

// selectBest picks the best pattern to keep based on health scores.
func (d *DuplicateDetector) selectBest(proposal *MergeProposal) {
	if len(proposal.Patterns) == 0 {
		return
	}

	bestIdx := 0
	bestScore := -1.0

	for i, p := range proposal.Patterns {
		score := 0.5 // default
		if hs, ok := d.scores[p.ID]; ok {
			score = hs.Overall
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	proposal.KeepID = proposal.Patterns[bestIdx].ID
	proposal.RemoveIDs = make([]string, 0, len(proposal.Patterns)-1)
	for i, p := range proposal.Patterns {
		if i != bestIdx {
			proposal.RemoveIDs = append(proposal.RemoveIDs, p.ID)
		}
	}
}

// unionFind is a simple disjoint-set / union-find structure.
type unionFind struct {
	parent []int
	rank   []int
}

func newUnionFind(n int) *unionFind {
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	return &unionFind{parent: parent, rank: rank}
}

func (uf *unionFind) find(x int) int {
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]] // path compression
		x = uf.parent[x]
	}
	return x
}

func (uf *unionFind) union(a, b int) {
	ra, rb := uf.find(a), uf.find(b)
	if ra == rb {
		return
	}
	if uf.rank[ra] < uf.rank[rb] {
		ra, rb = rb, ra
	}
	uf.parent[rb] = ra
	if uf.rank[ra] == uf.rank[rb] {
		uf.rank[ra]++
	}
}
