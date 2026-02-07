# mur Master Plan
## Unified AI Learning & Orchestration Platform

**Version:** 1.0  
**Date:** 2026-02-07  
**Author:** David + AI Collaboration

---

## 🎯 Executive Summary

mur 是一個 **統一 AI 學習與協作平台**，透過持續學習的 patterns 系統，讓團隊知識成為可複製、可累積的資產。核心是 **mur.core** 學習引擎，上層衍生多個領域應用，最終由 **mur.commander** 小型模型統籌，形成企業專屬的 AI 指揮系統。

```
                              ┌─────────────────┐
                              │  User / Team    │
                              └────────┬────────┘
                                       │
                              ┌────────▼────────┐
                              │  mur.commander  │  ◄── 小型模型，團隊知識蒸餾
                              │  (Orchestrator) │
                              └────────┬────────┘
                                       │
       ┌───────────────────────────────┼───────────────────────────────┐
       │                               │                               │
┌──────▼──────┐               ┌────────▼────────┐              ┌───────▼───────┐
│  mur.code   │               │    mur.help     │              │mur.marketing  │
│  (開發者)   │               │    (客服)       │              │   (行銷)      │
└──────┬──────┘               └────────┬────────┘              └───────┬───────┘
       │                               │                               │
       └───────────────────────────────┼───────────────────────────────┘
                                       │
                              ┌────────▼────────┐
                              │    mur.core     │  ◄── 學習引擎核心
                              │   (Learning)    │
                              └────────┬────────┘
                                       │
                    ┌──────────────────┼──────────────────┐
                    │                  │                  │
             ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐
             │   Claude    │    │    GPT      │    │   Gemini    │
             └─────────────┘    └─────────────┘    └─────────────┘
```

---

## 📦 Product Portfolio

### 1. mur.core — 學習引擎核心

**定位**：所有 mur 產品的共用基礎設施

| 元件 | 功能 | 技術 |
|------|------|------|
| **Pattern Engine** | CRUD、搜尋、版本控制 | Go + SQLite + Git |
| **Auto Classifier** | 自動分類、標籤推斷 | Embedding + Rules |
| **Security Layer** | Prompt injection 防護 | Dual LLM pattern |
| **Sync Engine** | 跨機器/團隊同步 | Git-based |
| **Analytics** | 使用追蹤、效果評估 | ClickHouse / SQLite |

**核心資料結構**：
```yaml
# Pattern Schema v2
id: uuid
name: string
content: string

tags:
  inferred:    # AI 推斷
    - domain: coding
      confidence: 0.92
  confirmed:   # 人工確認
    - error-handling
  negative:    # 明確排除
    - legacy

applies:
  file_patterns: ["*.swift", "*.go"]
  keywords: ["error", "exception"]
  context: {}

security:
  hash: sha256
  source: owner | team | community
  reviewed: boolean
  risk: low | medium | high

learning:
  effectiveness: 0.85
  usage_count: 42
  last_used: timestamp
  
lifecycle:
  status: active | deprecated | archived
  created: timestamp
  updated: timestamp
```

---

### 2. mur.code (mur.run) — 開發者 CLI

**定位**：統一 Multi-AI CLI 管理 + 程式碼學習系統

| 功能 | 說明 | 狀態 |
|------|------|------|
| Multi-CLI Runner | 統一執行 Claude/Gemini/Codex | ✅ Done |
| Pattern Learning | 從 session 萃取 patterns | 🔄 In Progress |
| Smart Routing | 依任務複雜度選 AI | ✅ Done |
| Team Sync | Git-based 知識共享 | ✅ Done |
| Web Dashboard | 視覺化管理 | ✅ Done |
| IDE Plugins | VS Code, Sublime, JetBrains | 🔄 Partial |

**差異化**：
- 不只是 wrapper，是 **知識累積系統**
- Patterns 越用越準，形成團隊 AI 記憶

---

### 3. mur.ide — 平行開發協作

**定位**：AI 驅動的任務分派 + 平行開發環境

**靈感來源**：
- Google Antigravity 的 Manager view
- Cursor 的 background agents
- Simon Willison 的 "parallel coding agent lifestyle"

**核心功能**：
```
┌─────────────────────────────────────────────────────────────┐
│                      mur.ide Manager                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  📋 Task Queue                    🔄 Active Agents          │
│  ├── #1 Refactor auth module      ├── Agent A → #1 (45%)   │
│  ├── #2 Add unit tests            ├── Agent B → #2 (80%)   │
│  ├── #3 Update API docs           ├── Agent C → #3 (20%)   │
│  └── #4 Fix login bug             └── Agent D → idle       │
│                                                             │
│  🌲 Branch Status                 📊 Resource Usage         │
│  ├── feature/auth-refactor        ├── Claude: 45k tokens   │
│  ├── feature/unit-tests           ├── GPT-4: 12k tokens    │
│  └── feature/api-docs             └── Gemini: 8k tokens    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**任務分派流程**：
```
User: "這個 PR 需要 review，同時幫我寫 unit tests，還有更新文件"

mur.ide 自動：
1. 拆解任務 → 3 個獨立工作項
2. 分析依賴 → tests 和 docs 可平行
3. 分配 branch → 每任務一個分支
4. 啟動 agents → 平行執行
5. 監控進度 → 即時狀態更新
6. 合併結果 → 衝突檢測 + 解決
```

**隔離機制**：
- 每個任務在獨立 Git branch
- 使用 worktree 實現真正隔離
- 自動偵測衝突，人工決策合併

---

### 4. mur.help — 智能客服

**定位**：AI 客服 + 知識庫管理

| 功能 | 說明 |
|------|------|
| Multi-channel | Zendesk, Intercom, Slack, Email |
| Pattern-based Response | 回覆模板 + 情境適配 |
| Sentiment Analysis | 情緒偵測 + 自動升級 |
| Knowledge Base | FAQ 自動維護 |
| Effectiveness Tracking | 解決率、滿意度追蹤 |

**Pattern 類型**：
```yaml
name: angry-customer-refund
applies:
  sentiment: [negative, angry]
  keywords: ["refund", "money back", "退款"]
  channel: [email, chat]

escalation:
  conditions:
    - repeated_contact: 3+
    - mention: ["lawyer", "BBB", "消保"]
  target: supervisor
  
response_template: |
  1. 同理情緒：我完全理解您的困擾...
  2. 確認問題：請讓我確認一下...
  3. 提供方案：我們可以...
```

---

### 5. mur.marketing — AI 行銷助手

**定位**：內容生成 + Campaign 管理

| 功能 | 說明 |
|------|------|
| Content Generation | 文案、社群貼文、Email |
| A/B Testing | 變體生成 + 效果追蹤 |
| Brand Voice | 品牌風格一致性 |
| Campaign Scheduling | 排程 + 自動發布 |
| Analytics | 轉換追蹤 + ROI |

---

### 6. mur.commander — 智能指揮官

**定位**：小型本地模型，蒸餾團隊知識，統籌多 AI 執行

**為什麼需要 Commander**：

| 問題 | Commander 解法 |
|------|----------------|
| 知識外流 | Patterns 留本地，只送 prompt |
| 團隊記憶 | 蒸餾成模型，永久保存 |
| 供應商依賴 | 核心邏輯自有，執行用外部 |
| 成本優化 | 本地分類，選最適合的 AI |
| 品質一致 | 自動套用公司標準 |

**架構**：
```
┌─────────────────────────────────────────────────────────────┐
│                      mur.commander                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │  Understanding  │    │  Prompt Craft   │                │
│  │  Layer          │    │  Layer          │                │
│  │                 │    │                 │                │
│  │  • Intent       │───▶│  • Pattern      │                │
│  │    detection    │    │    injection    │                │
│  │  • Context      │    │  • Context      │                │
│  │    extraction   │    │    enrichment   │                │
│  │  • Domain       │    │  • Style        │                │
│  │    classification│   │    enforcement  │                │
│  └─────────────────┘    └────────┬────────┘                │
│                                  │                          │
│  ┌─────────────────┐    ┌────────▼────────┐                │
│  │  Validation     │    │  Routing        │                │
│  │  Layer          │    │  Layer          │                │
│  │                 │    │                 │                │
│  │  • Output       │◀───│  • Cost-based   │                │
│  │    verification │    │  • Quality-based│                │
│  │  • Compliance   │    │  • Load balance │                │
│  │    check        │    │                 │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐      ┌────────┐      ┌────────┐
         │ Claude │      │  GPT   │      │ Gemini │
         └────────┘      └────────┘      └────────┘
```

**模型選擇**：
| Option | 模型 | Size | 優點 | 缺點 |
|--------|------|------|------|------|
| A | Llama 3.2 3B | ~2GB | 平衡 | 需 fine-tune |
| B | Phi-3 Mini | ~2GB | 快速 | 能力有限 |
| C | Qwen 2.5 3B | ~2GB | 多語言 | 社群較小 |
| D | **Hybrid** | - | 最佳 | 複雜 |

**推薦：Hybrid 方案**
- Rule-based → 快速路徑（10ms）
- Embedding retrieval → 語意搜尋（50ms）
- Small LLM → 複雜情況（200ms）

---

### 7. BitL Integration — 開發環境管理

**定位**：macOS 開發環境管理 + mur.core 整合

**整合點**：
| BitL 功能 | mur 整合 |
|-----------|----------|
| 專案管理 | 自動載入專案 patterns |
| 環境切換 | Context 自動切換 |
| 服務管理 | 學習 debug patterns |
| CLI 工具 | mur CLI 內建 |

```swift
// BitL + mur.core integration
class ProjectManager {
    let murCore: MurCore
    
    func openProject(_ project: Project) {
        // 1. 啟動環境
        startServices(project.services)
        
        // 2. 載入專案 patterns
        murCore.loadPatterns(
            project: project.name,
            tags: project.techStack
        )
        
        // 3. 設定 context
        murCore.setContext([
            "project": project.name,
            "language": project.primaryLanguage,
            "framework": project.frameworks
        ])
    }
}
```

---

## 🔒 Security Architecture

### Prompt Injection 防護

**參考設計模式**（來自 IBM/Google/Microsoft 論文）：

| Pattern | 說明 | 適用場景 |
|---------|------|----------|
| **Dual LLM** | 特權/隔離雙模型 | 處理不信任輸入 |
| **Plan-Then-Execute** | 先規劃再執行 | 多步驟任務 |
| **Context Minimization** | 移除不必要內容 | 資料查詢 |
| **Action-Selector** | 無反饋行動 | 簡單操作 |

**mur 實作**：

```go
// Security layers
type SecurityPipeline struct {
    // Layer 1: Input validation
    validator       InputValidator
    
    // Layer 2: Content sanitization  
    sanitizer       PatternSanitizer
    
    // Layer 3: Trust verification
    trustVerifier   TrustVerifier
    
    // Layer 4: Dual LLM isolation
    privilegedLLM   LLM  // 不接觸不信任內容
    quarantinedLLM  LLM  // 處理不信任內容
    
    // Layer 5: Output validation
    outputValidator OutputValidator
}

func (p *SecurityPipeline) Process(input Input) (Output, error) {
    // 1. Validate input
    if err := p.validator.Validate(input); err != nil {
        return nil, err
    }
    
    // 2. Sanitize patterns
    sanitized, warnings := p.sanitizer.Sanitize(input.Patterns)
    if len(warnings) > 0 {
        log.Warn("Sanitization warnings", warnings)
    }
    
    // 3. Check trust level
    trust := p.trustVerifier.Verify(input.Source)
    
    // 4. Route based on trust
    var result Output
    if trust >= TrustLevel.Verified {
        result = p.privilegedLLM.Process(sanitized)
    } else {
        // Quarantined processing
        symbolic := p.quarantinedLLM.Process(sanitized)
        result = p.privilegedLLM.Interpret(symbolic)
    }
    
    // 5. Validate output
    return p.outputValidator.Validate(result)
}
```

**Pattern 安全層級**：

```yaml
security:
  trust_levels:
    owner: 1.0       # 自己建立
    team: 0.8        # 團隊成員
    verified: 0.6    # 經過審核
    community: 0.3   # 社群貢獻
    untrusted: 0.0   # 未知來源
    
  deny_patterns:
    - "ignore previous"
    - "disregard instructions"
    - "you are now"
    - "system:"
    - "<|.*|>"
    
  actions:
    on_suspicious:
      - log
      - quarantine
      - notify_admin
    on_malicious:
      - reject
      - block_source
      - alert
```

---

## 💰 Business Model

### Pricing Tiers

| Tier | mur.code | mur.help | mur.marketing | mur.ide | mur.commander |
|------|----------|----------|---------------|---------|---------------|
| **Free** | ✅ 5 patterns | ✅ 100 對話/月 | ✅ 10 內容/月 | ❌ | ❌ |
| **Pro** $19/mo | ✅ Unlimited | - | - | - | - |
| **Team** $49/user/mo | ✅ + Sync | ✅ 多通道 | ✅ Campaign | ✅ 3 agents | ✅ Rule-based |
| **Enterprise** Custom | ✅ All | ✅ All | ✅ All | ✅ Unlimited | ✅ Fine-tuned |

### Value Differentiators

| Tier | Free | Team | Enterprise |
|------|------|------|------------|
| Pattern storage | Local | Git sync | Central server |
| Search | Keyword | Embedding | Semantic + Vector |
| Analytics | Basic | Advanced | Custom dashboards |
| Security | Hash only | Review workflow | Full audit |
| Commander | ❌ | Rule-based | Fine-tuned model |
| Support | Community | Email | Dedicated |

---

## 🗓️ Roadmap

### Phase 1: Foundation (Q1 2026) ← **現在**

| 週 | 目標 | 交付物 |
|----|------|--------|
| W1-2 | mur.core Pattern Schema v2 | 新格式 + 遷移工具 |
| W3-4 | Auto Classifier MVP | Keyword + File-based |
| W5-6 | Security Layer v1 | Hash + Source + Deny list |
| W7-8 | mur.code v0.5.0 Release | 整合新 core |

### Phase 2: Intelligence (Q2 2026)

| 月 | 目標 | 交付物 |
|----|------|--------|
| Apr | Embedding-based classification | 語意搜尋 |
| May | Effectiveness tracking | 學習迴路 |
| Jun | mur.ide MVP | 平行執行 2 agents |

### Phase 3: Commander (Q3 2026)

| 月 | 目標 | 交付物 |
|----|------|--------|
| Jul | Commander architecture | Hybrid routing |
| Aug | Fine-tuning pipeline | 訓練工具鏈 |
| Sep | mur.commander v1.0 | Enterprise ready |

### Phase 4: Ecosystem (Q4 2026)

| 月 | 目標 | 交付物 |
|----|------|--------|
| Oct | mur.help v1.0 | Multi-channel support |
| Nov | mur.marketing v1.0 | Campaign management |
| Dec | BitL integration | 完整整合 |

---

## 🏗️ Technical Architecture

### Monorepo Structure

```
mur/
├── core/                       # mur.core - 共用核心
│   ├── pattern/                # Pattern engine
│   │   ├── schema.go
│   │   ├── store.go
│   │   ├── retrieval.go
│   │   └── sync.go
│   ├── classifier/             # Auto classification
│   │   ├── keyword.go
│   │   ├── embedding.go
│   │   ├── rules.go
│   │   └── hybrid.go
│   ├── security/               # Security layer
│   │   ├── sanitizer.go
│   │   ├── trust.go
│   │   ├── dual_llm.go
│   │   └── validator.go
│   ├── learning/               # Learning engine
│   │   ├── feedback.go
│   │   ├── effectiveness.go
│   │   └── auto_deprecate.go
│   └── analytics/              # Usage tracking
│       ├── tracker.go
│       └── reporter.go
│
├── cmd/                        # CLI applications
│   ├── mur/                    # Main CLI (mur.code)
│   ├── mur-ide/                # IDE manager
│   └── mur-server/             # Central server
│
├── apps/                       # Domain applications
│   ├── code/                   # mur.code specific
│   ├── help/                   # mur.help specific
│   ├── marketing/              # mur.marketing specific
│   └── commander/              # mur.commander
│       ├── model/              # Model management
│       ├── prompt_craft/       # Prompt engineering
│       ├── router/             # AI routing
│       └── validator/          # Output validation
│
├── integrations/               # External integrations
│   ├── vscode/
│   ├── sublime/
│   ├── jetbrains/
│   ├── bitl/                   # BitL integration
│   └── openclaw/               # OpenClaw skill
│
├── server/                     # Web services
│   ├── api/                    # REST API
│   ├── dashboard/              # Web UI
│   └── webhook/                # External webhooks
│
└── training/                   # Commander training
    ├── data/                   # Training data generation
    ├── pipeline/               # Fine-tuning pipeline
    └── eval/                   # Evaluation tools
```

### Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Input                               │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Security Pipeline                             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │Validate │→│Sanitize │→│ Trust   │→│Quarantine│→│Validate │   │
│  │ Input   │ │ Content │ │ Verify  │ │(if needed)│ │ Output  │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Classification Engine                         │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐                           │
│  │ Signal  │→│ Domain  │→│ Pattern │                           │
│  │Extract  │ │ Score   │ │Retrieve │                           │
│  └─────────┘ └─────────┘ └─────────┘                           │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Commander (Orchestrator)                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐               │
│  │ Prompt  │→│ Router  │→│ Execute │→│Validate │               │
│  │ Craft   │ │ Select  │ │ AI Call │ │ Result  │               │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘               │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Learning Engine                               │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐                           │
│  │ Track   │→│ Update  │→│ Refine  │                           │
│  │Outcome  │ │Effective│ │ Tags    │                           │
│  └─────────┘ └─────────┘ └─────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔑 Key Success Metrics

### Product Metrics

| 產品 | 核心指標 | 目標 |
|------|----------|------|
| mur.code | Pattern 使用率 | >60% prompts 套用 pattern |
| mur.ide | 任務完成率 | >85% 自動完成 |
| mur.help | 首次解決率 | >70% |
| mur.marketing | 內容採用率 | >50% |
| mur.commander | Prompt 改善率 | >30% 效果提升 |

### Business Metrics

| 指標 | Y1 目標 | Y2 目標 |
|------|---------|---------|
| 活躍用戶 | 5,000 | 25,000 |
| 付費用戶 | 500 | 3,000 |
| ARR | $150K | $1M |
| Team 客戶 | 20 | 100 |
| Enterprise 客戶 | 2 | 10 |

---

## 🎯 Competitive Advantages

1. **知識資產化**
   - Patterns 是團隊資產，非供應商
   - 離職不流失，新人即繼承

2. **自有技術護城河**
   - Commander 模型越用越準
   - 難以複製的累積優勢

3. **Multi-AI 不鎖定**
   - 不依賴單一供應商
   - 隨時切換最佳選擇

4. **安全設計**
   - 敏感資料不外流
   - 企業等級安全架構

5. **統一生態系**
   - 開發、客服、行銷一套系統
   - 跨領域知識共享

---

## 📋 Immediate Next Steps (This Week)

### Priority 1: Pattern Schema v2
```bash
# 1. 設計新 schema
# 2. 寫遷移工具
# 3. 更新 CLI 讀寫邏輯
```

### Priority 2: Security Foundation
```bash
# 1. 實作 hash + source tracking
# 2. 加入 deny list scanning
# 3. 基本 trust level
```

### Priority 3: mur.code v0.5.0
```bash
# 1. 整合新 core
# 2. --show-classification flag
# 3. mur lint command
```

---

## 附錄：Research References

### Multi-Agent Orchestration
- Microsoft Azure AI Agent Design Patterns
- CrewAI role-driven orchestration
- n8n LangGraph integration

### Security Patterns
- "Design Patterns for Securing LLM Agents" (IBM/Google/Microsoft, 2025)
- OWASP LLM Security Top 10
- Dual LLM Pattern (Simon Willison)

### Fine-tuning
- H2O.ai Enterprise LLM Studio
- Snorkel AI distillation
- PEFT/LoRA for enterprise

### Parallel Development
- Google Antigravity Manager view
- Cursor background agents
- Conductor parallel runner

---

*This document is the living source of truth for mur product strategy.*
*Last updated: 2026-02-07*
