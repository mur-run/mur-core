# mur.commander — AI Orchestrator Specification

**Version:** 1.0  
**Status:** Draft  
**Date:** 2026-02-07

---

## Overview

mur.commander 是一個小型本地模型，作為團隊 AI 的「指揮官」，負責：
1. 理解使用者意圖
2. 注入團隊 patterns 製作最佳 prompt
3. 選擇最適合的 AI 執行
4. 驗證輸出符合團隊標準
5. 持續學習改進

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     mur.commander                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Understanding Layer                     │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │  Intent   │ │  Context  │ │  Domain   │         │   │
│  │  │ Detection │ │ Extraction│ │ Classify  │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│                            ▼                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Prompt Crafting Layer                   │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │  Pattern  │ │  Context  │ │   Style   │         │   │
│  │  │ Retrieval │ │ Enrichment│ │Enforcement│         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│                            ▼                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Routing Layer                           │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │   Cost    │ │  Quality  │ │   Load    │         │   │
│  │  │  Optimize │ │  Optimize │ │  Balance  │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│                            ▼                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Validation Layer                        │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │  Output   │ │Compliance │ │  Pattern  │         │   │
│  │  │  Quality  │ │   Check   │ │  Feedback │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐      ┌────────┐      ┌────────┐
         │ Claude │      │  GPT   │      │ Gemini │
         └────────┘      └────────┘      └────────┘
```

## Core Components

### 1. Understanding Layer

```go
// 意圖檢測
type IntentDetector struct {
    model      LocalLLM  // 小型本地模型
    classifier Classifier // 規則 + embedding
}

type Intent struct {
    Primary    string            // 主要意圖
    Secondary  []string          // 次要意圖
    Confidence float64
    Entities   map[string]string // 提取的實體
}

func (d *IntentDetector) Detect(input string, ctx Context) (Intent, error) {
    // 1. 快速規則匹配
    if intent := d.classifier.MatchRules(input); intent != nil {
        return *intent, nil
    }
    
    // 2. Embedding 相似度
    candidates := d.classifier.FindSimilar(input, 5)
    
    // 3. 如果信心不足，用本地 LLM
    if len(candidates) == 0 || candidates[0].Confidence < 0.7 {
        return d.model.DetectIntent(input, ctx)
    }
    
    return candidates[0].ToIntent(), nil
}

// 上下文萃取
type ContextExtractor struct {
    fileAnalyzer   FileAnalyzer
    historyParser  HistoryParser
    envDetector    EnvironmentDetector
}

type ExtractedContext struct {
    Files       []FileContext
    History     []Message
    Environment Environment
    Metadata    map[string]any
}

// 領域分類（使用 mur.core）
type DomainClassifier struct {
    core *core.ClassificationService
}
```

### 2. Prompt Crafting Layer

```go
// Prompt 製作器
type PromptCrafter struct {
    patternRetriever PatternRetriever
    contextEnricher  ContextEnricher
    styleEnforcer    StyleEnforcer
    templateEngine   TemplateEngine
}

func (c *PromptCrafter) Craft(
    input string,
    intent Intent,
    ctx ExtractedContext,
    patterns []Pattern,
) (string, error) {
    // 1. 選擇相關 patterns
    relevant := c.patternRetriever.Filter(patterns, intent, ctx)
    
    // 2. 豐富上下文
    enriched := c.contextEnricher.Enrich(ctx, intent)
    
    // 3. 應用團隊風格
    styled := c.styleEnforcer.Apply(enriched)
    
    // 4. 組合 prompt
    return c.templateEngine.Render(PromptTemplate{
        UserInput:  input,
        Intent:     intent,
        Patterns:   relevant,
        Context:    styled,
    })
}

// Pattern 檢索
type PatternRetriever struct {
    store          PatternStore
    embeddingIndex EmbeddingIndex
}

func (r *PatternRetriever) Filter(
    patterns []Pattern,
    intent Intent,
    ctx ExtractedContext,
) []Pattern {
    var result []Pattern
    
    for _, p := range patterns {
        score := r.calculateRelevance(p, intent, ctx)
        if score > 0.5 {
            result = append(result, p)
        }
    }
    
    // 排序並限制數量（token budget）
    sort.Slice(result, func(i, j int) bool {
        return result[i].Relevance > result[j].Relevance
    })
    
    return r.trimToTokenBudget(result, 2000)
}

// 風格強制器
type StyleEnforcer struct {
    teamStyle  StyleGuide
    domainRules map[string][]StyleRule
}

type StyleGuide struct {
    Language       string   // 回應語言
    Tone           string   // formal, casual, technical
    CodeStyle      string   // 程式碼風格
    NamingConvention string
    Preferences    []string
}
```

### 3. Routing Layer

```go
// AI 路由器
type AIRouter struct {
    providers   map[string]AIProvider
    costModel   CostModel
    qualityModel QualityModel
    loadBalancer LoadBalancer
}

type RoutingDecision struct {
    Provider    string
    Model       string
    Reason      string
    CostEstimate float64
    QualityScore float64
    Fallback    []string // 備選
}

func (r *AIRouter) Route(
    intent Intent,
    prompt string,
    constraints RoutingConstraints,
) RoutingDecision {
    scores := make(map[string]float64)
    
    for name, provider := range r.providers {
        // 成本評估
        costScore := r.costModel.Evaluate(provider, prompt)
        
        // 品質評估
        qualityScore := r.qualityModel.Evaluate(provider, intent)
        
        // 負載評估
        loadScore := r.loadBalancer.GetScore(name)
        
        // 加權組合
        scores[name] = constraints.Apply(costScore, qualityScore, loadScore)
    }
    
    best := maxScore(scores)
    return RoutingDecision{
        Provider:     best,
        Model:        r.providers[best].RecommendModel(intent),
        Reason:       r.explainChoice(best, scores),
        CostEstimate: r.costModel.Estimate(r.providers[best], prompt),
        QualityScore: r.qualityModel.Predict(r.providers[best], intent),
        Fallback:     r.getFallbacks(scores, best),
    }
}

// 成本模型
type CostModel struct {
    pricing map[string]PricingInfo
    history UsageHistory
}

type PricingInfo struct {
    InputTokenCost  float64 // per 1K tokens
    OutputTokenCost float64
    FreeQuota       int
}

// 品質模型（基於歷史表現）
type QualityModel struct {
    intentPerformance map[string]map[string]float64 // intent -> provider -> score
}
```

### 4. Validation Layer

```go
// 輸出驗證器
type OutputValidator struct {
    qualityChecker  QualityChecker
    complianceCheck ComplianceChecker
    feedbackRecorder FeedbackRecorder
}

type ValidationResult struct {
    Valid       bool
    Score       float64
    Issues      []Issue
    Suggestions []string
}

func (v *OutputValidator) Validate(
    output string,
    intent Intent,
    patterns []Pattern,
) ValidationResult {
    result := ValidationResult{Valid: true, Score: 1.0}
    
    // 1. 品質檢查
    qualityIssues := v.qualityChecker.Check(output, intent)
    result.Issues = append(result.Issues, qualityIssues...)
    
    // 2. 合規檢查
    complianceIssues := v.complianceCheck.Check(output, patterns)
    result.Issues = append(result.Issues, complianceIssues...)
    
    // 3. 計算分數
    result.Score = v.calculateScore(result.Issues)
    result.Valid = result.Score >= 0.7
    
    // 4. 生成建議
    result.Suggestions = v.generateSuggestions(result.Issues)
    
    return result
}

// 品質檢查
type QualityChecker struct {
    rules []QualityRule
}

type QualityRule interface {
    Check(output string, intent Intent) []Issue
}

// 合規檢查
type ComplianceChecker struct {
    patternRules  []PatternRule
    domainRules   map[string][]DomainRule
}
```

## Model Options

### Hybrid Approach (Recommended)

```
┌─────────────────────────────────────────────────────────────┐
│                    Hybrid Commander                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Fast Path (10ms)          Medium Path (50ms)               │
│  ┌─────────────────┐       ┌─────────────────┐             │
│  │   Rule Engine   │       │   Embedding     │             │
│  │                 │       │   Retrieval     │             │
│  │  • Keyword match│       │                 │             │
│  │  • Pattern match│       │  • Semantic     │             │
│  │  • Template     │       │    similarity   │             │
│  └─────────────────┘       └─────────────────┘             │
│           │                        │                        │
│           ▼                        ▼                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Slow Path (200ms)                       │   │
│  │              ┌─────────────────┐                     │   │
│  │              │   Local LLM     │                     │   │
│  │              │  (Llama 3.2 3B) │                     │   │
│  │              │                 │                     │   │
│  │              │ • Complex intent│                     │   │
│  │              │ • Disambiguation│                     │   │
│  │              │ • Generation    │                     │   │
│  │              └─────────────────┘                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Model Selection

| Model | Size | Speed | Quality | Use Case |
|-------|------|-------|---------|----------|
| Llama 3.2 1B | 1GB | 50ms | Medium | Simple intent |
| Llama 3.2 3B | 2GB | 150ms | Good | General |
| Phi-3 Mini | 2GB | 100ms | Good | Reasoning |
| Qwen 2.5 3B | 2GB | 150ms | Good | Multilingual |

### Fine-tuning Pipeline

```
┌─────────────────────────────────────────────────────────────┐
│                    Training Pipeline                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Data Collection                                         │
│     ├── Successful prompts from history                     │
│     ├── Pattern → enhanced prompt pairs                     │
│     └── Team feedback annotations                           │
│                                                             │
│  2. Data Processing                                         │
│     ├── Clean and deduplicate                               │
│     ├── Format as instruction pairs                         │
│     └── Split train/val/test                                │
│                                                             │
│  3. Fine-tuning                                             │
│     ├── Base model: Llama 3.2 3B                            │
│     ├── Method: LoRA / QLoRA                                │
│     ├── Epochs: 3-5                                         │
│     └── Validation: effectiveness improvement               │
│                                                             │
│  4. Evaluation                                              │
│     ├── A/B test vs base model                              │
│     ├── Measure prompt quality improvement                  │
│     └── User satisfaction survey                            │
│                                                             │
│  5. Deployment                                              │
│     ├── Export to GGUF                                      │
│     ├── Integrate with Ollama                               │
│     └── Monitor performance                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Training Data Format

```json
{
  "instruction": "Transform this user request into an optimal prompt",
  "input": {
    "user_request": "重構這個 API endpoint",
    "context": {
      "language": "go",
      "framework": "gin",
      "project": "backend-service"
    },
    "patterns": [
      {"name": "go-api-style", "content": "..."},
      {"name": "error-handling", "content": "..."}
    ]
  },
  "output": {
    "enhanced_prompt": "重構以下 API endpoint，遵循團隊標準：\n\n## 風格要求\n- 使用 gin.Context...",
    "routing_decision": {
      "provider": "claude",
      "model": "claude-4-sonnet",
      "reason": "Code refactoring task, Claude excels at..."
    }
  }
}
```

## API

```go
// Commander 主要介面
type Commander interface {
    // 處理使用者輸入
    Process(ctx context.Context, input ProcessInput) (*ProcessResult, error)
    
    // 取得處理解釋
    Explain(ctx context.Context, input ProcessInput) (*Explanation, error)
}

type ProcessInput struct {
    UserInput   string
    Context     Context
    Constraints Constraints
}

type ProcessResult struct {
    EnhancedPrompt string
    RoutingDecision RoutingDecision
    AppliedPatterns []Pattern
    ExecutionResult *ExecutionResult // 如果執行了
}

type Explanation struct {
    IntentAnalysis    IntentAnalysis
    PatternMatches    []PatternMatch
    PromptTransform   TransformSteps
    RoutingRationale  string
}
```

## Configuration

```yaml
# ~/.mur/commander.yaml

model:
  type: hybrid  # hybrid | llm-only | rules-only
  llm:
    provider: ollama
    model: llama3.2:3b
    temperature: 0.3
    
routing:
  strategy: balanced  # cost | quality | balanced
  providers:
    claude:
      enabled: true
      models: [claude-4-sonnet, claude-4-opus]
      budget_limit: 100  # USD per month
    openai:
      enabled: true
      models: [gpt-4, gpt-4-turbo]
    gemini:
      enabled: true
      models: [gemini-2.0-pro]
      
validation:
  min_quality_score: 0.7
  auto_retry: true
  max_retries: 2
  
training:
  data_dir: ~/.mur/training
  auto_collect: true
  min_samples: 100
  schedule: weekly
```

## Metrics

```go
// 追蹤的指標
type CommanderMetrics struct {
    // 處理指標
    ProcessingLatency    Histogram
    IntentAccuracy       Gauge
    PatternRelevance     Gauge
    
    // 路由指標
    ProviderUsage        Counter
    CostPerRequest       Histogram
    QualityScore         Histogram
    
    // 學習指標
    PromptImprovement    Gauge  // 改善率
    FeedbackPositive     Counter
    FeedbackNegative     Counter
}
```

---

*This specification defines the mur.commander orchestration system.*
