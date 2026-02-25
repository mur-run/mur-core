# Plan 008: Diffusion-Inspired Architecture Evolution

> Date: 2026-02-25
> Status: Draft
> Inspired by: Mercury 2 (Inception Labs) — diffusion-based reasoning LLM
> Scope: mur-core (Go v1) + mur (Rust v2)

## Executive Summary

借鏡 Mercury 2 擴散模型的核心理念（迭代精煉、平行處理、內建糾錯），結合多模態 Pattern 支援和 Pattern↔Workflow 統一架構，全面提升 MUR 的學習能力。

核心主張：**MUR 不是記憶系統，是學習系統。Pattern 會進化、Workflow 會精煉、系統會自我修正。**

---

## Part A: Diffusion-Inspired Pattern Lifecycle

### A1. Pattern Maturity Stages

**問題：** Pattern 提取後品質固定，好壞混雜。

**方案：** 引入 maturity 生命階段，像擴散模型從噪音到清晰。

```yaml
# Pattern Schema v2 新增欄位
maturity: draft          # draft → emerging → stable → canonical
confidence: 0.4          # 動態分數，非固定
decay:
  last_active: 2026-02-25
  half_life_days: 30     # 不使用自然衰減
```

**Maturity 規則：**

| Stage | Confidence | 條件 |
|-------|-----------|------|
| `draft` | 0.0 - 0.3 | 剛提取，或系統推測生成 |
| `emerging` | 0.3 - 0.6 | 被注入 3+ 次且未被否定 |
| `stable` | 0.6 - 0.85 | 被注入 10+ 次 + 正面 feedback |
| `canonical` | 0.85 - 1.0 | 用戶明確確認 or 高頻穩定使用 30+ 天 |

**Injection 優先級：** `canonical > stable > emerging > draft`（同分數時）

**實作位置：**
- Go: `internal/core/pattern/schema.go` — 加 `Maturity` + `Confidence` + `DecayMeta` 欄位
- Rust: `mur-common/src/pattern.rs` — 同步
- Inject scoring: `internal/core/inject/inject.go` — maturity bonus weight

### A2. Automatic Decay & Renewal (Half-Life System)

**問題：** 過時 pattern 堆積，consolidation 是手動的。

**方案：** 持續被動衰減 + 主動回血，模擬擴散的 forward/reverse process。

```
Forward (遺忘): confidence -= decay_rate * days_inactive
Reverse (回想): 
  - 注入且未否定 → +0.05
  - 明確正面 feedback → +0.15  
  - 被用戶在 session 中重複教 → +0.20 (emergence signal)
  - 低於 0.1 → auto-archive
  - archived pattern 被搜到並使用 → 復活到 emerging
```

**實作：**
- 新增 `internal/core/pattern/decay.go`
- 在每次 `mur sync` 或 `mur inject` 時觸發 decay 計算（O(n) scan，<1ms for 500 patterns）
- 不需要 daemon，利用現有命令的 lifecycle

### A3. Pattern Self-Correction (Feedback Loop)

**問題：** 注入的 pattern 效果好壞，系統不知道。

**方案：** Post-session 自動分析，回寫 pattern 分數。

```
Hook flow:
  pre-hook: inject patterns → record which patterns injected
  [user session]
  post-hook: analyze session → detect contradictions/confirmations → update patterns
```

**Contradiction detection（輕量版，不需 LLM）：**
1. 掃描 session 中用戶的修正語句（"不要用 X"、"改用 Y"、"錯了"）
2. 與 injected patterns 做 keyword match
3. 有矛盾 → pattern confidence -= 0.1 + 記錄 contradiction evidence

**Confirmation detection：**
1. AI 回應中引用了 pattern 的關鍵內容
2. 用戶沒有否定 → implicit confirmation → confidence += 0.03

**實作：**
- 新增 `internal/core/feedback/` package
- 擴展 `internal/hooks/` 的 post-hook 支援（Claude Code hooks 已有 PostToolUse）
- 整合 `inject.Tracker` 的 effectiveness tracking

---

## Part B: Multi-Modal Pattern Support

> **核心原則：** MUR 的使用者是 AI 模型，不是人類。模型已經知道「蘋果」是什麼。
> 有價值的多模態 pattern 是：**模型預訓練裡沒有的、你專屬的視覺知識。**
> 例如：你的 app UI mockup、你的系統架構圖、你專案特有的 error screenshot。

### B1. Phase 1 — Diagram Attachments（v2.1，成本最低價值最高）

Mermaid / PlantUML 格式的架構圖，**純文字存儲但可渲染**。

```yaml
name: bitl-architecture
content:
  technical: |
    BitL uses a 3-layer architecture...
  principle: |
    Always route through ServiceManager, never call Homebrew directly.
attachments:
  - type: diagram
    format: mermaid
    path: assets/bitl-architecture.mermaid
    description: "系統架構圖"
```

- 存儲：純文字（mermaid/plantuml），不需要 binary asset 管理
- 注入：直接 inline 到 prompt（模型能讀 mermaid）
- 搜尋：description 參與 text embedding search（零額外成本）
- **一張圖勝過 500 字的文字描述**

### B2. Phase 2 — Image Attachments（v3，等 CLIP 成熟）

等 CLIP-aligned embedding 在本地跑得順時（~6-12 個月），加圖片。

**挑戰（不急著解決）：**
1. Embedding 不統一 — text 用 qwen3，圖片要 CLIP/SigLIP，兩個向量空間不同
2. Token 預算 — 一張圖 ~1000 tokens base64，2000 budget 放不下幾張
3. 提取方式 — `mur learn extract` 怎麼判斷哪張截圖值得保存？
4. ROI — 99% AI coding 場景是純文字

### B3. 聲音 — 不做

除非 MUR 定位從 coding assistant 擴展到通用 AI 記憶系統。

---

## Part C: Pattern ↔ Workflow — 分開但共享基礎

> **決策：現在就分，不等 v2 完成。越晚分越痛。**
> **修正：分開的 struct + 目錄，但共享 KnowledgeBase + 統一 LanceDB index。**
>
> 理由：Pattern 和 Workflow 的演化方向不同。Pattern 往 maturity/decay/emergence，
> Workflow 往 execution/permissions/marketplace。硬綁在一起，兩邊都走不快。

### C1. 共享基礎 + 獨立結構

```rust
// mur-common/src/knowledge.rs — 共用欄位
pub struct KnowledgeBase {
    pub id: String,
    pub name: String,
    pub description: String,
    pub content: DualLayer,
    pub tags: TagSet,
    pub maturity: Maturity,        // draft → emerging → stable → canonical
    pub confidence: f64,
    pub decay: DecayMeta,
    pub security: SecurityMeta,
    pub learning: LearningMeta,
    pub lifecycle: LifecycleMeta,
    pub embedding_hash: Option<String>,
}

// mur-common/src/pattern.rs — 知識片段
pub struct Pattern {
    #[serde(flatten)]
    pub base: KnowledgeBase,
    pub applies: ApplyConditions,
    pub relations: Relations,
    pub attachments: Vec<Attachment>,    // diagrams (Phase 3)
    pub resources: Resources,
    pub health: HealthMeta,
}

// mur-common/src/workflow.rs — 可執行流程
pub struct Workflow {
    #[serde(flatten)]
    pub base: KnowledgeBase,
    pub steps: Vec<Step>,
    pub variables: Vec<Variable>,
    pub source_sessions: Vec<SourceRef>,
    pub permissions: Permissions,        // read/write/execute-only
    pub published_version: u32,
    pub trigger: String,                 // 何時使用
    pub tools: Vec<String>,             // 用到的工具
}

pub struct Step {
    pub order: u32,
    pub description: String,
    pub command: Option<String>,
    pub tool: Option<String>,
    pub needs_approval: bool,
    pub on_failure: FailureAction,       // skip, abort, retry
}

pub struct Variable {
    pub name: String,
    pub var_type: VarType,               // string, path, url, number, bool
    pub required: bool,
    pub default: Option<String>,
    pub description: String,
}
```

### C2. 儲存

```
~/.mur/
├── patterns/          # Knowledge patterns (YAML)
│   ├── swift-testing.yaml
│   └── swift-testing/         # 附件目錄 (diagrams)
├── workflows/         # Workflows (YAML)
│   ├── deploy-prod.yaml
│   └── deploy-prod/
│       └── revisions/         # auto-save snapshots
└── index/             # 統一 LanceDB
    └── mur.lance              # type 欄位區分 pattern vs workflow
```

### C3. 統一搜尋

LanceDB 統一 index，`item_type` 欄位 filter：

```rust
pub enum SearchResult {
    Pattern(Pattern),
    Workflow(Workflow),
}

// 搜尋 API
pub fn search(query: &str, filter: Option<ItemType>) -> Vec<SearchResult>;
```

```bash
mur search "swift testing"           # 搜全部
mur search "deploy" --type workflow  # 只搜 workflows
mur search "nginx" --type pattern    # 只搜 patterns
```

### C4. Inject 行為

- **Pattern** → 照舊注入 content
- **Workflow** → 注入時格式化為有序步驟 + 觸發條件
- Hook 選擇邏輯：prompt 分析後，同時從 patterns + workflows 取最相關的

### C5. 共享 Lifecycle

兩邊都走 maturity/decay/feedback：
- Pattern: draft → emerging → stable → canonical（基於 inject effectiveness）
- Workflow: draft → emerging → stable → canonical（基於 execution 成功率）

### C6. 進階功能（後續 Phase）

- **Workflow → Pattern 提取：** workflow 某 step 被多處引用 → 建議提取為 pattern
- **Pattern → Workflow 組合：** co-occurrence 偵測 → 建議組成 workflow
- **Cross-reference：** Pattern.relations 可引用 Workflow，反之亦然

### C7. 與 Go v1 的關係

Go v1 已有獨立 `internal/workflow/` 模組（types, store, extract, merge, permissions）。

**策略：**
- Go v1 保持現狀
- Rust v2 的 Workflow struct 對齊 Go v1 的 types（Steps, Variables, SourceRef, Permissions）
- Migration: v1 `~/.mur/workflows/` → v2 `~/.mur/workflows/`（結構接近，轉換低成本）
- v1 patterns 全部映射為 v2 Pattern（現有行為不變）

### C8. 工作量

| 改動 | 預估 |
|------|------|
| KnowledgeBase 共用 struct | 1h |
| Pattern struct 重構（embed KnowledgeBase） | 1h |
| Workflow struct（新增） | 1.5h |
| WorkflowStore (YAML read/write) | 1.5h |
| LanceDB 統一 index + item_type | 1h |
| CLI: `mur workflow list/show/create` 基礎命令 | 2h |
| inject 支援 workflow 格式化 | 30m |
| tests | 1.5h |
| **合計** | **~10h** |

---

## Part D: Advanced Diffusion Concepts

### D1. Cross-Session Emergence Detection

**問題：** 單 session 提取的 patterns 品質有限，真正有價值的 patterns 跨越多個 sessions。

**方案：**

```go
// internal/learn/emergence.go
type EmergenceDetector struct {
    behaviorIndex map[string][]BehaviorFingerprint
    threshold     int  // 出現 N 次 → emergent pattern
}

type BehaviorFingerprint struct {
    SessionID   string
    Behavior    string    // 行為摘要 (LLM 生成)
    EmbeddingID string    // 向量 ID
    Timestamp   time.Time
}
```

**流程：**
1. 每次 `mur learn extract` 後，除了提取明確 patterns，也生成 behavior fingerprints
2. 定期（`mur consolidate` 時）掃描 fingerprints
3. Cluster similar behaviors → 出現 3+ 次 → 候選 emergent pattern
4. 用 LLM 綜合所有 evidence → 生成一個精煉的 pattern（maturity: emerging）

### D2. Speculative Pre-loading

**問題：** Pattern injection 在 hook 的 critical path 上。

**方案：** 預測 + 預載。

```
背景：git status → 偵測當前 context
預載：該 project + 該 language 的 top patterns → memory cache
Hook 觸發時：直接從 cache 取，<1ms
```

**實作：**
- `internal/core/inject/speculative.go`
- 用 `fsnotify` 監聽 project 目錄變化
- 或更簡單：每次 `mur sync` 時預算 top patterns per project

### D3. Pattern Decomposition (Reverse Diffusion)

**問題：** 大 pattern 中部分內容過時，整體被降級。

**方案：** 自動分解 → 保留好的部分。

```
"swift-testing-guide" (effectiveness: 0.3, declining)
  ↓ decompose (triggered when stable → emerging regression)
  "use-test-macro"      → effectiveness: 0.7 ✓
  "use-expect-macro"    → effectiveness: 0.6 ✓  
  "use-suite-macro"     → effectiveness: 0.1 ✗ (archived)
```

**觸發條件：**
- Pattern maturity 從 stable 退回 emerging
- 且 content 長度 > 200 字（有分解空間）
- 用 LLM 分解 + 分別追蹤

### D4. Proactive Pattern Hallucination

**最實驗性的功能。** 系統觀察用戶行為，主動猜測未被明確表達的偏好。

```
觀察：用戶在 Go 專案中連續 5 個 session 都用 errgroup
系統：生成 draft pattern "prefer-errgroup-for-concurrency"
      confidence: 0.15 (very low)
      maturity: draft
注入時：混入 stable patterns 中（低權重）
觀察：AI 遵循了，用戶沒否定 → confidence += 0.05
3 個 sessions 後：confidence 0.30 → 升級 emerging
```

**安全機制：**
- Draft patterns 永遠排在最後
- 每次最多注入 1 個 draft pattern（避免噪音）
- 3 次被否定 → 自動刪除

---

## Part E: Implementation Roadmap

### 📍 考量：Go v1 vs Rust v2

Rust v2 已開始，Phase 1-3 完成。策略：
- **架構設計** → 直接寫入 Rust v2 spec
- **快速驗證的功能** → 可在 Go v1 先 prototype
- **重大新模組** → 直接在 Rust v2 做

### Phase 0: Pattern/Workflow 分離 + KnowledgeBase (Day 1-2, ~10hr) ⚡ DO FIRST
**目標：** Pattern 和 Workflow 獨立 struct，共享 KnowledgeBase，統一 LanceDB

| Task | Where | Est |
|------|-------|-----|
| KnowledgeBase 共用 struct | `mur-common/src/knowledge.rs` | 1h |
| Pattern struct 重構（embed KnowledgeBase） | `mur-common/src/pattern.rs` | 1h |
| Workflow struct + Step/Variable types | `mur-common/src/workflow.rs` | 1.5h |
| WorkflowStore (YAML read/write) | `mur-core/src/store/workflow.rs` | 1.5h |
| LanceDB 統一 index + item_type 欄位 | `mur-core/src/store/lance.rs` | 1h |
| CLI: `mur workflow list/show/create` | `mur-core/src/main.rs` | 2h |
| inject 支援 workflow 格式化 | `mur-core/src/retrieve/` | 30m |
| tests | | 1.5h |

### Phase 1: Pattern Maturity + Decay (Week 1)
**目標：** Pattern 會自動進化和衰退

| Task | Where | Est |
|------|-------|-----|
| Pattern schema 加 maturity/confidence/decay fields | Rust v2 `pattern.rs` | 2h |
| Decay calculator | Rust v2 `evolve/decay.rs` | 3h |
| Inject scoring 加 maturity weight | Rust v2 `retrieve/` | 2h |
| `mur status` 顯示 maturity 分佈 | CLI | 1h |
| Migration: v1 patterns 預設 maturity=stable | migrate | 1h |
| Tests | | 2h |

### Phase 2: Feedback Loop (Week 2)
**目標：** 系統自動從 session 結果學習

| Task | Where | Est |
|------|-------|-----|
| Feedback signal types + storage | Rust v2 `evolve/feedback.rs` | 3h |
| Post-session contradiction detector (keyword-based) | Rust v2 `capture/feedback.rs` | 4h |
| Implicit confirmation detector | same | 3h |
| Hook integration (post-hook writes feedback) | Go v1 hooks (Claude Code) | 3h |
| Feedback → confidence update pipeline | Rust v2 | 2h |
| Tests | | 3h |

### Phase 3: Diagram Attachments (Week 3)
**目標：** Patterns 支援 mermaid/plantuml 圖表附件

| Task | Where | Est |
|------|-------|-----|
| Pattern schema 加 `attachments` field (diagram only) | Rust v2 `pattern.rs` | 1h |
| Mermaid/PlantUML inline inject formatter | Rust v2 `retrieve/` | 2h |
| `mur new --diagram <path>` CLI | CLI | 1h |
| `mur search` 包含 attachment descriptions | search | 1h |
| Tests | | 1h |

### Phase 4: Cross-Session Emergence (Week 4)
**目標：** 跨 session 行為自動浮現為 patterns

| Task | Where | Est |
|------|-------|-----|
| Behavior fingerprint extraction | `capture/emergence.rs` | 4h |
| Fingerprint storage + indexing | store | 3h |
| Clustering + emergence detection | `evolve/emergence.rs` | 4h |
| LLM-based evidence synthesis | evolve | 3h |
| `mur learn --emerge` CLI flag | CLI | 1h |
| Tests | | 3h |

### Phase 5: Knowledge↔Workflow Intelligence (Week 5)
**目標：** 自動偵測 patterns 間的 workflow 關係

| Task | Where | Est |
|------|-------|-----|
| Co-occurrence matrix tracking | analytics | 3h |
| Workflow decompose → knowledge extraction | evolve | 3h |
| Knowledge composition suggestion | suggest | 3h |
| Tests | | 2h |

### Phase 6: Advanced (Week 6+, 依需求)
- Speculative pre-loading
- Pattern decomposition (大 pattern 拆小)
- Proactive hallucination
- Image attachments (等 CLIP 成熟, ~6-12 months)
- Visual pattern extraction (需要 vision LLM)

---

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                        MUR Core v2                            │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐                │
│  │ Capture  │    │ Retrieve │    │  Evolve  │                │
│  │          │    │          │    │          │                │
│  │ • Extract│    │ • Gate   │    │ • Decay  │  ← NEW        │
│  │ • Noise  │    │ • Score  │    │ • Feedback│ ← NEW        │
│  │ • Dedup  │    │ • Inject │    │ • Linker │                │
│  │ • Visual │←   │ • Spec.  │←   │ • Emerge │ ← NEW        │
│  │   NEW    │    │   NEW    │    │ • Decomp │ ← NEW        │
│  └────┬─────┘    └────┬─────┘    └────┬─────┘                │
│       │               │               │                       │
│  ┌────▼───────────────▼───────────────▼─────┐                │
│  │              Pattern Store                 │                │
│  │  • YAML (source of truth)                 │                │
│  │  • LanceDB (vector index)                 │                │
│  │  • Maturity + Confidence + Decay    ← NEW │                │
│  │  • Multimodal Attachments           ← NEW │                │
│  └────────────────┬──────────────────────────┘                │
│                   │                                           │
│  ┌────────────────▼──────────────────────────┐                │
│  │           Workflow Store                    │                │
│  │  • pattern_refs (bidirectional)     ← NEW  │                │
│  │  • Shared maturity lifecycle        ← NEW  │                │
│  │  • Decompose / Compose              ← NEW  │                │
│  └────────────────────────────────────────────┘                │
│                                                               │
│  ┌────────────────────────────────────────────┐                │
│  │           Suggest Engine              NEW   │                │
│  │  • Co-occurrence matrix                     │                │
│  │  • Pattern → Workflow composition           │                │
│  │  • Workflow → Pattern decomposition         │                │
│  │  • Proactive hallucination                  │                │
│  └────────────────────────────────────────────┘                │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

---

## Product Positioning Shift

### Before
> MUR captures patterns and injects them into AI tools.

### After  
> **MUR: Your AI tools don't just remember — they learn.**
>
> Patterns evolve with use. Bad ones fade, good ones sharpen.
> Workflows emerge from repeated behavior.
> Visual knowledge is captured alongside code.
> Everything connects — patterns reference workflows, workflows decompose into patterns.
>
> It's not a database. It's a learning loop.

### Key Metrics to Surface (Dashboard)
- **Pattern Maturity Distribution** — 多少 canonical vs draft
- **Learning Velocity** — 每週有幾個 patterns 從 draft → stable
- **Repetition Reduction** — 用了 MUR 後重複教 AI 的次數減少 %
- **Emergence Count** — 系統自動發現了幾個 cross-session patterns
- **Feedback Loop Health** — 多少 patterns 有 feedback data

---

## Decisions Log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Pattern/Workflow 分離時機 | 現在（Phase 0） | 越晚分越痛，alpha 階段改成本最低 |
| Pattern/Workflow 架構 | 分開 struct + 目錄，共享 KnowledgeBase | 演化方向不同，硬綁會成 God Object |
| Maturity 幾個階段 | 4 (draft/emerging/stable/canonical) | 太多太複雜，4 個剛好 |
| Decay 觸發時機 | 每次 sync/inject | 不需要 daemon，利用現有命令 |
| Multimodal Phase 1 | Diagram only (mermaid/plantuml) | 純文字存儲，成本最低價值最高 |
| Multimodal Phase 2 | Image (等 CLIP 成熟) | 6-12 月後再評估 |
| 聲音支援 | 不做 | Coding assistant 不需要 |
| Feedback 初版 | Keyword-based | 不需要 LLM，快速可靠 |
| 先做哪裡 | Rust v2 為主 | Go v1 已有完整功能，新架構直接做在 v2 |

---

## Open Questions

1. **Decay half-life 預設值？** 30 天 vs 14 天 — 需要 real data 驗證
2. **Emergence threshold？** 3 次 vs 5 次 — 太低會有噪音，太高會漏掉
3. **Proactive hallucination** 要不要做？— 風險是注入錯誤 pattern，但有 maturity 機制保護
4. **Rust v2 real-data validation** — 要先驗證 Phase 1-3 再加新功能？還是邊加邊驗？
5. **v1→v2 workflow migration** — Go v1 的 `~/.mur/workflows/` 結構接近 v2，直接搬還是需要轉換？
