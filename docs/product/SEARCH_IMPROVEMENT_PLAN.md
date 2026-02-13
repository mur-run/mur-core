# MUR Core Search 技術改進計劃

**Version:** 1.0  
**Date:** 2026-02-13  
**Status:** Planning  
**Author:** David + AI Collaboration

---

## 📋 Executive Summary

本計劃解決兩個核心問題：

1. **Sync 架構不可擴展** — 所有 patterns 同步給所有用戶，伺服器負擔大、浪費頻寬
2. **搜尋精準度不足** — 純 vector search 只有 ~62% 精準度，漏掉關鍵字精確匹配

**目標：**
- Sync 架構改為 **Selective Sync + On-Demand Community**
- 搜尋升級為 **Hybrid Search (BM25 + Vector + RRF)**
- 精準度從 ~62% 提升到 ~84%
- 支援百萬級 community patterns

---

## 🔴 Problem Statement

### 問題 1：Sync 架構不可擴展

**現況：**
```
mur sync (Pro/Team)
    │
    ▼
下載所有 patterns    ← 伺服器負擔
    │
    ▼
存到 ~/.mur/patterns/  ← Swift 開發者有 PHP patterns
    │
    ▼
全部送到 AI          ← 浪費 tokens
```

**數據：**
- 10,000 patterns × 10,000 users = 100M 次下載
- 每個 pattern ~2KB，總計 ~200GB 頻寬/月
- 伺服器成本隨用戶數線性增長

### 問題 2：搜尋精準度不足

**現況：**
- 本地搜尋：Embedding vector search only
- 純語意搜尋，可能漏掉關鍵字精確匹配
- 無 metadata filtering，不相關結果混入

**研究數據（2025-2026 業界基準）：**

| 方法 | 精準度 | 缺點 |
|------|--------|------|
| 純關鍵字 (BM25) | ~50% | 不懂語意 |
| 純向量 (Vector) | ~62% | 可能漏關鍵字 |
| **Hybrid + RRF** | **~84%** | 複雜度較高 |

---

## ✅ Solution Architecture

### Part 1: Selective Sync + On-Demand Community

#### 核心原則

> **你自己的 patterns：同步**  
> **Community patterns：搜尋，不下載**

#### 三種 Pattern 來源

| 來源 | 同步方式 | 儲存位置 | 頻寬消耗 |
|------|---------|----------|----------|
| **你的 patterns** | ✅ 雲端同步 | `~/.mur/patterns/` | O(your patterns) |
| **Team patterns** | ✅ 雲端同步 | `~/.mur/patterns/` | O(team patterns) |
| **Community** | ❌ 搜尋不下載 | Server only | O(search queries) |

#### 新的 Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    mur search "API retry"                    │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
    ┌─────────────────┐            ┌─────────────────┐
    │ Local Search    │            │ Community API   │
    │ ~/.mur/patterns │            │ api.mur.run     │
    │ (instant)       │            │ (100-200ms)     │
    └────────┬────────┘            └────────┬────────┘
              │                               │
              └───────────────┬───────────────┘
                              ▼
    ┌─────────────────────────────────────────────────────────┐
    │ Results:                                                 │
    │ 📍 Local: 2 patterns                                    │
    │ 🌐 Community: 15 patterns (showing top 5)               │
    └─────────────────────────────────────────────────────────┘
```

#### 使用 Community Pattern

```bash
# 方式 1：臨時注入（不下載）
mur search "API retry" --inject
# → 直接把 community 結果注入 AI，用完即丟

# 方式 2：複製到本地（永久）
mur community copy abc123
# → 加入你的 patterns，之後會同步

# 方式 3：搜尋時自動混合
mur search "API retry"
# → 本地 + community 結果一起顯示
```

#### 頻寬對比

| 場景 | 現在 | 改後 |
|------|------|------|
| 10K users, 100K community patterns | ~2TB/月 | ~20GB/月 |
| 100K users, 1M community patterns | 不可行 | ~200GB/月 |

---

### Part 2: Hybrid Search Architecture

#### 搜尋流程

```
Query: "Swift async testing"
              │
              ▼
┌─────────────────────────────────────────┐
│ Step 1: Metadata Pre-Filter             │
│ WHERE tech_stack @> '["swift"]'         │
│ → 從 100萬 patterns 篩到 5萬            │
└──────────────────┬──────────────────────┘
                   │
       ┌───────────┴───────────┐
       ▼                       ▼
┌─────────────┐         ┌─────────────┐
│ BM25        │         │ Vector      │
│ 關鍵字搜尋  │         │ 語意搜尋    │
│ PostgreSQL  │         │ pgvector    │
│ tsvector    │         │ cosine      │
└──────┬──────┘         └──────┬──────┘
       │                       │
       └───────────┬───────────┘
                   ▼
┌─────────────────────────────────────────┐
│ Step 2: Reciprocal Rank Fusion (RRF)    │
│                                         │
│ score = Σ 1/(k + rank), k=60            │
│                                         │
│ → 合併兩種排名，取最佳                  │
└──────────────────┬──────────────────────┘
                   ▼
┌─────────────────────────────────────────┐
│ Step 3: (Optional) Cross-Encoder Rerank │
│ → 對 top 10-20 做精細排序               │
└──────────────────┬──────────────────────┘
                   ▼
            Top 5 Results
```

#### 為什麼 Hybrid Search？

| 查詢類型 | BM25 | Vector | Hybrid |
|----------|------|--------|--------|
| "Swift XCTest" (精確) | ✅ | ⚠️ | ✅ |
| "如何測試非同步程式" (語意) | ❌ | ✅ | ✅ |
| "Swift async testing" (混合) | ⚠️ | ⚠️ | ✅ |

#### RRF 算法說明

```
Reciprocal Rank Fusion (RRF)
─────────────────────────────

對於每個搜尋結果：
  RRF_score = Σ 1/(k + rank_i)

其中：
  k = 60 (常數，平衡權重)
  rank_i = 該結果在第 i 個排名列表中的位置

範例：
  Pattern A: BM25 rank=1, Vector rank=5
  RRF_A = 1/(60+1) + 1/(60+5) = 0.0164 + 0.0154 = 0.0318

  Pattern B: BM25 rank=10, Vector rank=2
  RRF_B = 1/(60+10) + 1/(60+2) = 0.0143 + 0.0161 = 0.0304

  → Pattern A 排名更高 (在兩邊都相對靠前)
```

---

## 🛠️ Technical Implementation

### Database Schema Changes

```sql
-- 1. 加入 tsvector 欄位（BM25 搜尋用）
ALTER TABLE patterns ADD COLUMN content_tsv tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(content, '')), 'C')
  ) STORED;

-- 2. 加入 tech_stack 欄位（metadata filter 用）
ALTER TABLE patterns ADD COLUMN tech_stack jsonb DEFAULT '[]';

-- 3. 建立索引
CREATE INDEX idx_patterns_tsv ON patterns USING GIN(content_tsv);
CREATE INDEX idx_patterns_tech ON patterns USING GIN(tech_stack);
CREATE INDEX idx_patterns_embedding ON patterns USING ivfflat(embedding vector_cosine_ops);
```

### Hybrid Search SQL

```sql
-- Community Hybrid Search API
CREATE OR REPLACE FUNCTION search_patterns_hybrid(
  query_text TEXT,
  query_embedding vector(1536),
  tech_filter jsonb DEFAULT NULL,
  limit_count INT DEFAULT 10
) RETURNS TABLE (
  id UUID,
  name TEXT,
  description TEXT,
  score FLOAT,
  source TEXT
) AS $$
WITH 
-- Metadata pre-filter
filtered AS (
  SELECT p.* FROM patterns p
  WHERE p.visibility = 'public'
    AND (tech_filter IS NULL OR p.tech_stack @> tech_filter)
),
-- BM25 full-text search
fulltext AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY ts_rank_cd(content_tsv, query) DESC) AS r
  FROM filtered, plainto_tsquery('english', query_text) query
  WHERE content_tsv @@ query
  LIMIT 50
),
-- Vector semantic search  
semantic AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY embedding <=> query_embedding) AS r
  FROM filtered
  LIMIT 50
),
-- RRF fusion
rrf AS (
  SELECT id, 1.0 / (60 + r) AS score FROM fulltext
  UNION ALL
  SELECT id, 1.0 / (60 + r) AS score FROM semantic
)
SELECT 
  p.id,
  p.name,
  p.description,
  SUM(rrf.score) AS score,
  'community' AS source
FROM rrf 
JOIN patterns p USING (id)
GROUP BY p.id, p.name, p.description
ORDER BY score DESC
LIMIT limit_count;
$$ LANGUAGE SQL;
```

### API Endpoints

```yaml
# Community Search API
GET /api/v1/community/search
  Query Parameters:
    q: string           # 搜尋關鍵字
    tech: string[]      # 技術棧過濾 (e.g., swift,go)
    limit: int          # 結果數量 (default: 10)
    include_content: bool  # 是否包含完整內容

  Response:
    patterns:
      - id: uuid
        name: string
        description: string
        score: float
        author: string
        stars: int
        copies: int
        content: string?  # 只在 include_content=true 時

# Inject endpoint (臨時注入)
POST /api/v1/community/inject
  Body:
    pattern_ids: uuid[]
    
  Response:
    content: string  # 合併的 pattern 內容，直接注入 AI
```

### CLI Changes

```go
// cmd/mur/cmd/search.go

var searchCmd = &cobra.Command{
    Use:   "search <query>",
    Short: "Search patterns (local + community)",
    RunE: func(cmd *cobra.Command, args []string) error {
        query := strings.Join(args, " ")
        
        // 1. Search local
        localResults := searchLocal(query)
        
        // 2. Search community (if enabled)
        var communityResults []Pattern
        if !localOnly {
            techStack := cfg.GetTechStack()
            communityResults = searchCommunity(query, techStack)
        }
        
        // 3. Display results
        displayResults(localResults, communityResults)
        
        // 4. Optional: inject to AI
        if inject {
            injectToAI(selectedResults)
        }
        
        return nil
    },
}

// Flags
var (
    localOnly bool    // --local: 只搜本地
    communityOnly bool // --community: 只搜社群
    inject bool       // --inject: 注入 AI
)
```

### Config: Tech Stack

```yaml
# ~/.mur/config.yaml
tech_stack:
  - swift
  - go
  - docker

search:
  include_community: true  # 預設搜尋包含 community
  auto_inject: false       # 是否自動注入結果
```

```bash
# CLI commands
mur config set tech-stack swift,go,docker
mur config get tech-stack
```

---

## 📈 Embedding Model Strategy

### 現況 vs 目標

| 項目 | 現在 | Phase 1 | Phase 2 |
|------|------|---------|---------|
| Model | text-embedding-ada-002 | text-embedding-3-large | Voyage Code 3 |
| Dimensions | 1536 | 3072 | 2048 |
| 適合場景 | 通用 | 通用高品質 | Code 專用 |
| 成本 | $ | $$ | $$$ |

### 遷移策略

```
Phase 1: 保持 ada-002
  └─ 先完成 Hybrid Search 架構
  
Phase 2: 升級 text-embedding-3-large
  └─ 背景 re-embed 所有 patterns
  └─ 支援漸進式遷移（新舊共存）
  
Phase 3: 評估 Voyage Code 3
  └─ A/B 測試比較效果
  └─ Code 相關查詢使用 Voyage
  └─ 通用查詢使用 OpenAI
```

---

## 🗓️ Implementation Roadmap

### Phase 1: Selective Sync (Week 1-2)

| Task | Priority | Effort |
|------|----------|--------|
| 移除 community 批次 sync | P0 | 2d |
| `mur sync` 只同步 user/team patterns | P0 | 1d |
| 更新 sync 相關文檔 | P1 | 0.5d |

**Deliverable:** Sync 不再下載 community patterns

### Phase 2: Community Search API (Week 2-3)

| Task | Priority | Effort |
|------|----------|--------|
| DB schema: content_tsv, tech_stack | P0 | 1d |
| Hybrid search SQL function | P0 | 2d |
| `GET /api/v1/community/search` | P0 | 1d |
| `POST /api/v1/community/inject` | P1 | 1d |
| Backfill tsvector for existing patterns | P0 | 0.5d |

**Deliverable:** Community search API live

### Phase 3: CLI Integration (Week 3-4)

| Task | Priority | Effort |
|------|----------|--------|
| `mur search` 加 `--community` flag | P0 | 1d |
| `mur search` 預設搜 local + community | P0 | 1d |
| `mur config set tech-stack` | P1 | 0.5d |
| `--inject` 臨時注入功能 | P1 | 1d |
| 更新 mur-index/SKILL.md 模板 | P0 | 0.5d |

**Deliverable:** CLI 支援 community search

### Phase 4: Optimization (Week 4-5)

| Task | Priority | Effort |
|------|----------|--------|
| Search result caching | P2 | 1d |
| Rate limiting for search API | P1 | 0.5d |
| Analytics: search queries tracking | P2 | 1d |
| A/B test: hybrid vs vector-only | P2 | 2d |

**Deliverable:** Production-ready search

### Phase 5: Embedding Upgrade (Future)

| Task | Priority | Effort |
|------|----------|--------|
| 評估 text-embedding-3-large | P3 | 1d |
| 設計 dual-embedding 架構 | P3 | 1d |
| 背景 re-embed pipeline | P3 | 2d |

**Deliverable:** 升級 embedding model

---

## 📊 Success Metrics

### Technical Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Search precision | ~62% | >80% | Manual eval on 100 queries |
| Search latency (P95) | - | <300ms | API monitoring |
| Sync bandwidth | O(users×patterns) | O(users×their_patterns) | Server metrics |

### Business Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Community search usage | 0 | >1000/day | API logs |
| Pattern copy rate | - | >5% | Conversion tracking |
| Server cost | $X | <$X | Fly.io billing |

---

## 🔒 Security Considerations

### Community Pattern Trust

```yaml
# Pattern trust levels
trust_levels:
  owner: 1.0       # 自己建立
  team: 0.8        # 團隊成員
  verified: 0.6    # 經過審核（社群高星）
  community: 0.3   # 一般社群
  
# Inject 時的處理
inject_policy:
  verified: inject_directly
  community: show_warning_first
  low_star: require_confirmation
```

### Rate Limiting

```yaml
# Search API rate limits
rate_limits:
  free: 100/hour
  pro: 1000/hour
  team: 10000/hour
  enterprise: unlimited
```

---

## 📎 Appendix

### Research References

1. **Hybrid Search**
   - ParadeDB: "Hybrid Search in PostgreSQL: The Missing Manual" (2025-10)
   - Elastic: "Hybrid Search and Semantic Reranking" (2025-09)
   - Supabase: "Hybrid Search Documentation" (2026-02)

2. **RRF Algorithm**
   - Original paper: Cormack, Clarke, Buettcher (2009)
   - DEV.to: "Building Hybrid Search for RAG" (2026-02) — 62% → 84% improvement

3. **Embedding Models**
   - Modal: "6 Best Code Embedding Models Compared" (2025)
   - Elephas: "13 Best Embedding Models in 2026" (2025-12)

### Related Documents

- [MUR_MASTER_PLAN.md](./MUR_MASTER_PLAN.md) — 產品總體規劃
- [SAAS_PLAN.md](./SAAS_PLAN.md) — SaaS 商業計畫
- [PRODUCTHUNT.md](../PRODUCTHUNT.md) — ProductHunt 發布計畫

---

*This document is the source of truth for MUR search improvements.*  
*Last updated: 2026-02-13*
