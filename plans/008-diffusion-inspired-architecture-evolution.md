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

### B1. Pattern Content Types

**問題：** Patterns 目前只支援文字。現代 AI CLI（Claude Code、Gemini）支援圖片理解，UI patterns、架構圖、錯誤截圖都無法被 pattern 捕捉。

**方案：** 擴展 Pattern schema 支援多模態內容。

```yaml
# Pattern Schema v2 擴展
name: ios-layout-pattern
content: |
  When building iOS layouts, use this constraint pattern...
  
attachments:
  - type: image
    path: assets/layout-example.png    # 相對於 pattern 目錄
    description: "正確的 Auto Layout constraint 結構"
    role: example                       # example | reference | error | before-after
  - type: image  
    path: assets/layout-antipattern.png
    description: "常見的錯誤 layout 方式"
    role: error
  - type: snippet
    path: assets/layout.swift
    language: swift
    description: "完整可用的 layout code"

content_type: text         # text | rich (with attachments) | visual-only
```

**Storage：**
```
~/.mur/patterns/
├── ios-layout-pattern.yaml
└── ios-layout-pattern/           # 同名目錄放 assets
    ├── layout-example.png
    └── layout-antipattern.png
```

**Injection 策略：**
- 文字 CLI（不支援圖片的）→ 只注入 text content + attachment descriptions
- 多模態 CLI（Claude Code、Gemini）→ 注入 text + 圖片 reference
- 圖片本身不 inline inject（太大），而是生成參考路徑讓 AI 自行讀取

**使用場景：**
1. **UI/UX patterns** — 正確的 layout 截圖 + 反面教材
2. **Architecture diagrams** — 系統架構圖作為 pattern 附件
3. **Error screenshots** — "遇到這個錯誤畫面時，解法是..."
4. **Before/After** — 重構前後的對比截圖

### B2. Visual Pattern Extraction

**問題：** Session 中的截圖、UI 操作目前無法被提取為 patterns。

**方案：** 從 session recordings 中提取視覺 patterns。

```
Session recording (mur:in/mur:out):
  event: tool_call(screenshot) → 檢測到截圖
  event: user("這個 UI 應該長這樣") → 語義標記
  → 提取: visual pattern with screenshot + description
```

**實作（Phase 2，需要 LLM vision）：**
- `internal/learn/visual_extract.go` — 掃描 session events 中的圖片 tool calls
- 用 vision LLM 生成圖片描述
- 與文字 context 合併，生成 rich pattern

### B3. Multimodal Search

**現有：** 文字 embedding search（OpenAI/Ollama）

**擴展：** 支援 CLIP-like 跨模態搜索

```bash
mur search "rounded corner card layout"
# → 找到文字 pattern + 帶截圖的 visual pattern

mur search --image screenshot.png
# → 以圖搜圖，找到類似 UI 的 patterns
```

**分階段：**
- Phase 1: 圖片的 description 參與文字 embedding search（零成本）
- Phase 2: 真正的 vision embedding（需要 CLIP model，Ollama 支援）

---

## Part C: Pattern ↔ Workflow Unified Architecture

### C1. 統一知識圖譜

**問題：** Patterns 和 Workflows 是獨立的兩套系統，但實際上高度相關。

**現狀：**
```
~/.mur/patterns/     → Pattern Store (YAML)
~/.mur/workflows/    → Workflow Store (YAML + index.json)
```
沒有 cross-reference。一個 workflow 可能用到某些 patterns，但系統不知道。

**方案：** 建立雙向連結。

```yaml
# Pattern: swift-testing.yaml
relations:
  used_in_workflows: ["wf-ios-test-setup", "wf-swift-migration"]
  
# Workflow: ios-test-setup/workflow.yaml  
pattern_refs:
  - pattern: swift-testing
    step: 3                    # 在第 3 步用到
    role: prerequisite         # prerequisite | reference | output
  - pattern: xcode-config
    step: 1
    role: reference
```

**自動發現：**
- Workflow 的 steps 與 pattern content 做語義匹配
- 高相似度 → 自動建立 `pattern_refs`
- 利用 Rust v2 的 `evolve/linker.rs` Zettelkasten 機制

### C2. Workflow-to-Pattern Extraction

**問題：** Workflows 太重（多步驟 SOP），有時候用戶只需要其中一個 step 的知識。

**方案：** 支援 Workflow → Pattern 反向提取。

```bash
mur workflows decompose <workflow-id>
# → 分析每個 step
# → 有通用價值的 step → 提取為獨立 pattern
# → 例：workflow "deploy-to-production" 的 step 3 "check nginx config" 
#      → 提取為 pattern "nginx-config-checklist"
```

**自動觸發：**
- 當同一個 workflow step 被多個不同 workflow 引用時 → 建議提取為 pattern
- 這就是 DRY 原則在知識層面的應用

### C3. Pattern-to-Workflow Composition

**反向操作：** 多個相關 patterns → 自動建議組成 workflow。

```
Patterns:
  - "swift-testing-setup"
  - "xcode-scheme-config" 
  - "ci-fastlane-config"
  
系統檢測到這三個 patterns 經常在同一個 session 中被一起使用
→ 建議: "要不要把這些組成一個 'iOS CI Setup' workflow?"
```

**實作：**
- `internal/core/suggest/composition.go`
- 基於 co-occurrence matrix（哪些 patterns 常一起被注入/使用）
- Threshold: 3+ sessions 中共同出現

### C4. Unified Lifecycle

Pattern 和 Workflow 共享生命週期語義：

| 階段 | Pattern | Workflow |
|------|---------|----------|
| draft | 剛提取，低 confidence | 剛錄製，未編輯 |
| emerging | 被使用幾次，效果待驗證 | 被跑過幾次，step 可能需要調整 |
| stable | 持續有效 | 穩定可靠，可分享 |
| canonical | 團隊標準 | 團隊 SOP |
| archived | 過時 | 不再使用 |

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

### Phase 3: Multimodal Patterns (Week 3)
**目標：** Patterns 支援圖片附件

| Task | Where | Est |
|------|-------|-----|
| Pattern schema 加 `attachments` field | Rust v2 `pattern.rs` | 2h |
| Asset directory convention + resolver | Rust v2 `store/` | 3h |
| Inject formatter: multimodal vs text-only output | Rust v2 `retrieve/` | 3h |
| `mur new --image <path>` CLI | CLI | 2h |
| `mur search` 包含 attachment descriptions | search | 2h |
| Tests | | 2h |

### Phase 4: Pattern ↔ Workflow Links (Week 4)
**目標：** Patterns 和 Workflows 雙向連結

| Task | Where | Est |
|------|-------|-----|
| Pattern `relations.used_in_workflows` field | schema | 1h |
| Workflow `pattern_refs` field | workflow types | 1h |
| Auto-discovery: workflow steps ↔ pattern matching | `suggest/composition.rs` | 4h |
| `mur workflows decompose` command | CLI | 3h |
| Co-occurrence matrix tracking | analytics | 3h |
| Composition suggestion ("these patterns → workflow?") | suggest | 3h |
| Tests | | 3h |

### Phase 5: Cross-Session Emergence (Week 5)
**目標：** 跨 session 行為自動浮現為 patterns

| Task | Where | Est |
|------|-------|-----|
| Behavior fingerprint extraction | `capture/emergence.rs` | 4h |
| Fingerprint storage + indexing | store | 3h |
| Clustering + emergence detection | `evolve/emergence.rs` | 4h |
| LLM-based evidence synthesis | evolve | 3h |
| `mur learn --emerge` CLI flag | CLI | 1h |
| Tests | | 3h |

### Phase 6: Advanced (Week 6+, 依需求)
- Speculative pre-loading
- Pattern decomposition
- Proactive hallucination
- Visual pattern extraction (需要 vision LLM)
- CLIP-based multimodal search

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
| Maturity 幾個階段 | 4 (draft/emerging/stable/canonical) | 太多太複雜，4 個剛好 |
| Decay 觸發時機 | 每次 sync/inject | 不需要 daemon，利用現有命令 |
| Multimodal 圖片存放 | 同名子目錄 | 符合 pattern 一個 YAML 一個目錄的慣例 |
| Feedback 初版 | Keyword-based | 不需要 LLM，快速可靠 |
| Pattern↔Workflow links | 雙向 + auto-discovery | 手動維護不現實 |
| 先做哪裡 | Rust v2 為主 | Go v1 已有完整功能，新架構直接做在 v2 |

---

## Open Questions

1. **Decay half-life 預設值？** 30 天 vs 14 天 — 需要 real data 驗證
2. **Multimodal inject 格式？** Claude Code 支援 `![](path)` 嗎？還是需要 base64？
3. **Emergence threshold？** 3 次 vs 5 次 — 太低會有噪音，太高會漏掉
4. **Proactive hallucination** 要不要做？— 風險是注入錯誤 pattern，但有 maturity 機制保護
5. **Rust v2 timeline** — Phase 1-3 done，但 real-data validation 還沒做，要先驗證再加新功能？
