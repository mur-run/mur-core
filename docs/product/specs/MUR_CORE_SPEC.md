# mur.core — Learning Engine Specification

**Version:** 1.0  
**Status:** Draft  
**Date:** 2026-02-07

---

## Overview

mur.core 是所有 mur 產品的共用學習引擎，負責 pattern 管理、自動分類、安全防護和效果追蹤。

## Core Components

### 1. Pattern Engine

```go
// Pattern 核心結構
type Pattern struct {
    ID          string         `yaml:"id"`
    Name        string         `yaml:"name"`
    Content     string         `yaml:"content"`
    Tags        TagSet         `yaml:"tags"`
    Applies     ApplyCondition `yaml:"applies"`
    Security    SecurityMeta   `yaml:"security"`
    Learning    LearningMeta   `yaml:"learning"`
    Lifecycle   LifecycleMeta  `yaml:"lifecycle"`
}

type TagSet struct {
    Inferred  []TagScore `yaml:"inferred"`   // AI 推斷
    Confirmed []string   `yaml:"confirmed"`  // 人工確認
    Negative  []string   `yaml:"negative"`   // 明確排除
}

type TagScore struct {
    Tag        string  `yaml:"tag"`
    Confidence float64 `yaml:"confidence"`
}

type ApplyCondition struct {
    FilePatterns []string       `yaml:"file_patterns"`
    Keywords     []string       `yaml:"keywords"`
    Sentiment    []string       `yaml:"sentiment"`
    Context      map[string]any `yaml:"context"`
}

type SecurityMeta struct {
    Hash     string     `yaml:"hash"`
    Source   TrustLevel `yaml:"source"`
    Reviewed bool       `yaml:"reviewed"`
    Reviewer string     `yaml:"reviewer"`
    Risk     RiskLevel  `yaml:"risk"`
}

type LearningMeta struct {
    Effectiveness float64   `yaml:"effectiveness"`
    UsageCount    int       `yaml:"usage_count"`
    LastUsed      time.Time `yaml:"last_used"`
    Feedback      []Feedback `yaml:"-"` // Not persisted
}

type LifecycleMeta struct {
    Status  LifecycleStatus `yaml:"status"`
    Created time.Time       `yaml:"created"`
    Updated time.Time       `yaml:"updated"`
}
```

### 2. Auto Classifier

```go
// 多層分類器
type Classifier interface {
    Classify(input ClassifyInput) []DomainScore
}

type ClassifyInput struct {
    Content     string
    FileContext []FileInfo
    History     []Message
    Context     map[string]any
}

type DomainScore struct {
    Domain     string
    Confidence float64
    Signals    []string
}

// 混合分類器實作
type HybridClassifier struct {
    keyword   *KeywordClassifier
    rule      *RuleClassifier
    embedding *EmbeddingClassifier
    llm       *LLMClassifier // 可選，複雜情況
}

func (h *HybridClassifier) Classify(input ClassifyInput) []DomainScore {
    // Fast path: 規則匹配
    if scores := h.rule.Classify(input); len(scores) > 0 {
        return scores
    }
    
    // Medium: 關鍵字 + 向量
    keywordScores := h.keyword.Classify(input)
    embeddingScores := h.embedding.Classify(input)
    
    // 融合結果
    merged := mergeScores(keywordScores, embeddingScores)
    
    // 如果信心不足，用 LLM
    if maxConfidence(merged) < 0.6 && h.llm != nil {
        return h.llm.Classify(input)
    }
    
    return merged
}
```

### 3. Security Layer

```go
// 安全管線
type SecurityPipeline struct {
    validator    InputValidator
    sanitizer    ContentSanitizer
    trustManager TrustManager
    dualLLM      *DualLLMGuard
}

// 輸入驗證
type InputValidator struct {
    maxLength   int
    allowedMime []string
    schemas     map[string]Schema
}

// 內容清理
type ContentSanitizer struct {
    denyPatterns []regexp.Regexp
    warnPatterns []regexp.Regexp
}

var defaultDenyPatterns = []string{
    `(?i)ignore\s+(all\s+)?previous`,
    `(?i)disregard\s+(all\s+)?instructions`,
    `(?i)you\s+are\s+now`,
    `(?i)new\s+instructions?:`,
    `(?i)system\s*:`,
    `<\|[^|]+\|>`,  // Special tokens
    `\[INST\]`,     // Instruction markers
    `<<SYS>>`,      // System markers
}

// 信任管理
type TrustManager struct {
    levels map[string]TrustLevel
}

type TrustLevel int

const (
    Untrusted TrustLevel = iota
    Community
    Verified
    Team
    Owner
)

// Dual LLM 防護
type DualLLMGuard struct {
    privileged  LLM // 不接觸不信任內容
    quarantined LLM // 處理不信任內容
}

func (d *DualLLMGuard) Process(input string, trust TrustLevel) (string, error) {
    if trust >= Team {
        return d.privileged.Generate(input)
    }
    
    // 隔離處理，返回符號化結果
    symbolic := d.quarantined.GenerateSymbolic(input)
    
    // 特權模型解釋符號，不接觸原始內容
    return d.privileged.InterpretSymbolic(symbolic)
}
```

### 4. Learning Engine

```go
// 學習引擎
type LearningEngine struct {
    feedbackStore FeedbackStore
    analyzer      EffectivenessAnalyzer
    deprecator    AutoDeprecator
}

// 回饋記錄
type Feedback struct {
    PatternID  string
    SessionID  string
    Timestamp  time.Time
    Applied    bool
    Outcome    Outcome
    UserRating int // 1-5, optional
    Context    map[string]any
}

type Outcome int

const (
    Success Outcome = iota
    Partial
    Failed
    Skipped
)

// 效果分析器
type EffectivenessAnalyzer struct {
    windowSize int // 計算窗口大小
    weights    []float64
}

func (e *EffectivenessAnalyzer) Calculate(feedbacks []Feedback) float64 {
    if len(feedbacks) == 0 {
        return 0.5 // 初始值
    }
    
    // 加權移動平均
    var sum, weightSum float64
    for i, fb := range feedbacks {
        weight := e.weights[min(i, len(e.weights)-1)]
        sum += float64(fb.Outcome) * weight
        weightSum += weight
    }
    
    return sum / weightSum
}

// 自動淘汰
type AutoDeprecator struct {
    minEffectiveness float64       // 最低效果閾值
    maxIdleDuration  time.Duration // 最長閒置時間
}

func (a *AutoDeprecator) Check(p Pattern) DeprecationAction {
    if p.Learning.Effectiveness < a.minEffectiveness {
        return DeprecationAction{
            Action: Deprecate,
            Reason: "effectiveness below threshold",
        }
    }
    
    if time.Since(p.Learning.LastUsed) > a.maxIdleDuration {
        return DeprecationAction{
            Action: Warn,
            Reason: "unused for extended period",
        }
    }
    
    return DeprecationAction{Action: None}
}
```

### 5. Sync Engine

```go
// Git-based 同步
type SyncEngine struct {
    repo     *git.Repository
    branch   string
    remote   string
    conflictResolver ConflictResolver
}

func (s *SyncEngine) Push() error {
    // 1. Commit 本地變更
    // 2. Pull remote
    // 3. 解決衝突
    // 4. Push
}

func (s *SyncEngine) Pull() error {
    // 1. Fetch remote
    // 2. Merge/Rebase
    // 3. 解決衝突
}

// 衝突解決策略
type ConflictResolver interface {
    Resolve(local, remote Pattern) (Pattern, error)
}

type DefaultResolver struct{}

func (d *DefaultResolver) Resolve(local, remote Pattern) (Pattern, error) {
    // 策略：更高效果的 pattern 優先
    if local.Learning.Effectiveness >= remote.Learning.Effectiveness {
        return local, nil
    }
    return remote, nil
}
```

## API Reference

### Pattern CRUD

```go
type PatternStore interface {
    Create(p Pattern) error
    Get(id string) (Pattern, error)
    Update(p Pattern) error
    Delete(id string) error
    List(filter Filter) ([]Pattern, error)
    Search(query string) ([]PatternMatch, error)
}
```

### Classification

```go
type ClassificationService interface {
    Classify(input ClassifyInput) ([]DomainScore, error)
    GetRelevantPatterns(scores []DomainScore, limit int) ([]Pattern, error)
}
```

### Security

```go
type SecurityService interface {
    Validate(input string) error
    Sanitize(content string) (string, []Warning, error)
    GetTrustLevel(source string) TrustLevel
    ProcessSecure(input string, trust TrustLevel) (string, error)
}
```

### Learning

```go
type LearningService interface {
    RecordFeedback(fb Feedback) error
    GetEffectiveness(patternID string) (float64, error)
    RunDeprecationCheck() ([]DeprecationAction, error)
    RefreshClassifications() error
}
```

## Configuration

```yaml
# ~/.murmur/core.yaml

pattern:
  storage: ~/.murmur/patterns
  format: yaml  # yaml | json | markdown
  
classifier:
  keyword:
    enabled: true
  embedding:
    enabled: true
    model: bge-small-en-v1.5
    index: ~/.murmur/vectors
  llm:
    enabled: false  # Optional, for complex cases
    model: llama3.2:3b
    
security:
  validation:
    max_length: 50000
  sanitization:
    deny_patterns: default  # or custom list
    on_detect: warn  # warn | reject | strip
  trust:
    default_level: community
    
learning:
  effectiveness:
    window_size: 20
    weights: [1.0, 0.9, 0.8, 0.7, 0.6]
  deprecation:
    min_effectiveness: 0.3
    max_idle_days: 90
    
sync:
  enabled: true
  repo: git@github.com:org/patterns.git
  branch: auto  # auto = hostname
  auto_push: true
  auto_pull: true
```

## Events

```go
// Core 發出的事件
type Event interface {
    Type() string
    Timestamp() time.Time
}

type PatternCreated struct { Pattern Pattern }
type PatternUpdated struct { Pattern Pattern; Changes []Change }
type PatternDeleted struct { PatternID string }
type PatternApplied struct { PatternID string; SessionID string }
type FeedbackReceived struct { Feedback Feedback }
type ClassificationUpdated struct { PatternID string; NewTags []TagScore }
type DeprecationWarning struct { PatternID string; Reason string }
type SecurityAlert struct { Input string; Threat string; Action string }
```

## Versioning

- Core version: semver (1.0.0)
- Pattern schema version: integer (v2)
- 向下相容：新版本必須能讀舊格式
- 遷移工具：`mur migrate patterns`

---

*This specification is the source of truth for mur.core implementation.*
