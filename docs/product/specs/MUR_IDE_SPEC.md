# mur.ide — Parallel Development Specification

**Version:** 1.0  
**Status:** Draft  
**Date:** 2026-02-07

---

## Overview

mur.ide 是一個 AI 驅動的平行開發環境，能夠自動分派任務給多個 AI agents，實現真正的平行開發。

## Inspiration

- **Google Antigravity**: Manager view 同時管理多個 AI agents
- **Cursor**: Background agents 背景執行任務
- **Conductor/Verdent**: 隔離 branch 平行執行
- **Simon Willison**: "Parallel coding agent lifestyle"

## Core Concepts

### 1. Task Decomposition

```
User: "這個 PR 需要 code review，同時寫 unit tests，還有更新 API 文件"

mur.ide 分析:
┌─────────────────────────────────────────────────────────────┐
│  Original Task                                               │
│  "PR review + unit tests + API docs"                         │
└─────────────────────────────────────────────────────────────┘
                              │
                    Task Decomposer
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   Task #1     │     │   Task #2     │     │   Task #3     │
│  Code Review  │     │  Unit Tests   │     │   API Docs    │
│               │     │               │     │               │
│ Deps: none    │     │ Deps: none    │     │ Deps: none    │
│ Priority: P1  │     │ Priority: P2  │     │ Priority: P2  │
│ Agent: Claude │     │ Agent: GPT    │     │ Agent: Gemini │
└───────────────┘     └───────────────┘     └───────────────┘
        │                     │                     │
        ▼                     ▼                     ▼
  review/pr-123        test/pr-123          docs/pr-123
     (branch)             (branch)             (branch)
```

### 2. Branch Isolation

```
main
 │
 ├── task/1-code-review      ◄── Agent A working
 │
 ├── task/2-unit-tests       ◄── Agent B working  
 │
 └── task/3-api-docs         ◄── Agent C working

每個任務在獨立 branch，使用 git worktree 實現真正隔離
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       mur.ide                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   Task Manager                       │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │Decomposer │ │Dependency │ │ Scheduler │         │   │
│  │  │           │ │  Analyzer │ │           │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   Agent Pool                         │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │   │
│  │  │ Agent A │ │ Agent B │ │ Agent C │ │ Agent D │   │   │
│  │  │ (busy)  │ │ (busy)  │ │ (busy)  │ │ (idle)  │   │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                 Workspace Manager                    │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │ Worktree  │ │  Branch   │ │  Merge    │         │   │
│  │  │  Manager  │ │  Manager  │ │  Manager  │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    Monitor                           │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐         │   │
│  │  │ Progress  │ │  Conflict │ │  Resource │         │   │
│  │  │  Tracker  │ │  Detector │ │  Monitor  │         │   │
│  │  └───────────┘ └───────────┘ └───────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Task Manager

```go
// 任務定義
type Task struct {
    ID          string
    Title       string
    Description string
    Type        TaskType
    Status      TaskStatus
    Priority    int
    Dependencies []string
    AssignedAgent string
    Branch      string
    Worktree    string
    CreatedAt   time.Time
    StartedAt   *time.Time
    CompletedAt *time.Time
    Result      *TaskResult
}

type TaskType string

const (
    TaskCodeReview    TaskType = "code_review"
    TaskUnitTest      TaskType = "unit_test"
    TaskRefactor      TaskType = "refactor"
    TaskDocumentation TaskType = "documentation"
    TaskBugFix        TaskType = "bug_fix"
    TaskFeature       TaskType = "feature"
)

type TaskStatus string

const (
    StatusPending   TaskStatus = "pending"
    StatusQueued    TaskStatus = "queued"
    StatusRunning   TaskStatus = "running"
    StatusCompleted TaskStatus = "completed"
    StatusFailed    TaskStatus = "failed"
    StatusBlocked   TaskStatus = "blocked"
)

// 任務分解器
type TaskDecomposer struct {
    classifier *TaskClassifier
    patterns   []DecompositionPattern
}

func (d *TaskDecomposer) Decompose(input string, context Context) ([]Task, error) {
    // 1. 分析輸入意圖
    intents := d.classifier.Classify(input)
    
    // 2. 應用分解模式
    var tasks []Task
    for _, intent := range intents {
        pattern := d.findPattern(intent.Type)
        subtasks := pattern.Apply(intent, context)
        tasks = append(tasks, subtasks...)
    }
    
    // 3. 分析依賴關係
    d.analyzeDependencies(tasks)
    
    // 4. 設定優先級
    d.setPriorities(tasks)
    
    return tasks, nil
}

// 依賴分析器
type DependencyAnalyzer struct {
    rules []DependencyRule
}

type DependencyRule interface {
    Check(task1, task2 Task) bool
}

// 例如：測試依賴於實作
type TestDependsOnImplementation struct{}

func (r *TestDependsOnImplementation) Check(task1, task2 Task) bool {
    return task1.Type == TaskUnitTest && 
           task2.Type == TaskFeature &&
           affectsSameFiles(task1, task2)
}
```

### 2. Agent Pool

```go
// Agent 定義
type Agent struct {
    ID       string
    Name     string
    Provider string  // claude, openai, gemini
    Model    string
    Status   AgentStatus
    CurrentTask *string
    Capabilities []string
    Stats    AgentStats
}

type AgentStatus string

const (
    AgentIdle    AgentStatus = "idle"
    AgentBusy    AgentStatus = "busy"
    AgentError   AgentStatus = "error"
    AgentOffline AgentStatus = "offline"
)

// Agent Pool 管理
type AgentPool struct {
    agents   map[string]*Agent
    queue    *TaskQueue
    assigner *TaskAssigner
    mu       sync.RWMutex
}

func (p *AgentPool) AssignTask(task Task) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // 1. 找到最適合的空閒 agent
    agent := p.assigner.FindBest(task, p.getIdleAgents())
    if agent == nil {
        // 沒有空閒 agent，加入隊列
        p.queue.Push(task)
        return nil
    }
    
    // 2. 分配任務
    agent.Status = AgentBusy
    agent.CurrentTask = &task.ID
    
    // 3. 啟動執行
    go p.executeTask(agent, task)
    
    return nil
}

// 任務分配器
type TaskAssigner struct {
    capabilities map[string][]string  // provider -> capabilities
    performance  map[string]map[TaskType]float64  // agent -> task_type -> score
}

func (a *TaskAssigner) FindBest(task Task, agents []*Agent) *Agent {
    var best *Agent
    var bestScore float64
    
    for _, agent := range agents {
        score := a.calculateScore(agent, task)
        if score > bestScore {
            best = agent
            bestScore = score
        }
    }
    
    return best
}

func (a *TaskAssigner) calculateScore(agent *Agent, task Task) float64 {
    // 基於能力匹配
    capabilityScore := a.matchCapabilities(agent, task)
    
    // 基於歷史表現
    performanceScore := a.performance[agent.ID][task.Type]
    
    // 基於成本
    costScore := a.getCostScore(agent)
    
    return capabilityScore*0.4 + performanceScore*0.4 + costScore*0.2
}
```

### 3. Workspace Manager

```go
// Workspace 使用 git worktree 實現隔離
type WorkspaceManager struct {
    repo      *git.Repository
    worktrees map[string]*Worktree
    basePath  string
}

type Worktree struct {
    Path   string
    Branch string
    TaskID string
    Status WorktreeStatus
}

func (m *WorkspaceManager) CreateWorktree(task Task) (*Worktree, error) {
    branchName := fmt.Sprintf("task/%s-%s", task.ID, slugify(task.Title))
    worktreePath := filepath.Join(m.basePath, ".mur-worktrees", task.ID)
    
    // 1. 創建 branch
    err := m.repo.CreateBranch(branchName, "main")
    if err != nil {
        return nil, err
    }
    
    // 2. 創建 worktree
    cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
    if err := cmd.Run(); err != nil {
        return nil, err
    }
    
    wt := &Worktree{
        Path:   worktreePath,
        Branch: branchName,
        TaskID: task.ID,
        Status: WorktreeActive,
    }
    
    m.worktrees[task.ID] = wt
    return wt, nil
}

func (m *WorkspaceManager) CleanupWorktree(taskID string) error {
    wt, ok := m.worktrees[taskID]
    if !ok {
        return nil
    }
    
    // 1. 移除 worktree
    cmd := exec.Command("git", "worktree", "remove", wt.Path)
    if err := cmd.Run(); err != nil {
        return err
    }
    
    delete(m.worktrees, taskID)
    return nil
}

// Merge 管理
type MergeManager struct {
    conflictResolver ConflictResolver
}

func (m *MergeManager) MergeTask(task Task) (*MergeResult, error) {
    // 1. 嘗試自動合併
    result, err := m.autoMerge(task.Branch, "main")
    if err == nil {
        return result, nil
    }
    
    // 2. 如果有衝突，嘗試自動解決
    if result.HasConflicts {
        resolved, err := m.conflictResolver.Resolve(result.Conflicts)
        if err != nil {
            // 無法自動解決，需要人工介入
            return &MergeResult{
                Status:      MergeNeedsReview,
                Conflicts:   result.Conflicts,
                Suggestions: m.getSuggestions(result.Conflicts),
            }, nil
        }
        return resolved, nil
    }
    
    return result, nil
}
```

### 4. Monitor

```go
// 進度追蹤
type ProgressTracker struct {
    tasks   map[string]*TaskProgress
    updates chan ProgressUpdate
}

type TaskProgress struct {
    TaskID      string
    Status      TaskStatus
    Progress    float64  // 0.0 - 1.0
    CurrentStep string
    StartedAt   time.Time
    EstimatedCompletion time.Time
    Logs        []LogEntry
}

type ProgressUpdate struct {
    TaskID   string
    Type     UpdateType
    Progress float64
    Message  string
}

// 衝突偵測
type ConflictDetector struct {
    fileWatcher *fsnotify.Watcher
    activeFiles map[string][]string  // file -> taskIDs
}

func (d *ConflictDetector) CheckPotentialConflicts(task Task) []PotentialConflict {
    var conflicts []PotentialConflict
    
    for file, taskIDs := range d.activeFiles {
        if task.AffectsFile(file) && len(taskIDs) > 0 {
            for _, otherTaskID := range taskIDs {
                if otherTaskID != task.ID {
                    conflicts = append(conflicts, PotentialConflict{
                        File:     file,
                        Task1:    task.ID,
                        Task2:    otherTaskID,
                        Severity: d.calculateSeverity(file),
                    })
                }
            }
        }
    }
    
    return conflicts
}

// 資源監控
type ResourceMonitor struct {
    tokenUsage   map[string]int64  // provider -> tokens
    costTracker  *CostTracker
    limits       ResourceLimits
}

type ResourceLimits struct {
    MaxConcurrentAgents int
    MaxTokensPerHour    int64
    MaxCostPerDay       float64
}

func (m *ResourceMonitor) CanStartTask(task Task) (bool, string) {
    // 檢查併發限制
    if m.getActiveAgentCount() >= m.limits.MaxConcurrentAgents {
        return false, "max concurrent agents reached"
    }
    
    // 檢查 token 限制
    if m.getHourlyTokens() >= m.limits.MaxTokensPerHour {
        return false, "hourly token limit reached"
    }
    
    // 檢查成本限制
    if m.costTracker.GetDailyCost() >= m.limits.MaxCostPerDay {
        return false, "daily cost limit reached"
    }
    
    return true, ""
}
```

## UI/UX

### CLI Interface

```bash
# 啟動平行開發 session
mur ide start

# 提交任務
mur ide task "Review this PR, add tests, update docs"

# 查看狀態
mur ide status

# Output:
# ┌─────────────────────────────────────────────────────────────┐
# │                    mur.ide Status                           │
# ├─────────────────────────────────────────────────────────────┤
# │                                                             │
# │  📋 Tasks                          🤖 Agents                │
# │  ├── #1 Code Review     ████████░░ 80%   Agent-A (Claude)  │
# │  ├── #2 Unit Tests      ███░░░░░░░ 30%   Agent-B (GPT)     │
# │  └── #3 API Docs        █████████░ 90%   Agent-C (Gemini)  │
# │                                                             │
# │  🌲 Branches                       📊 Resources            │
# │  ├── task/1-code-review  +42 -15   Tokens: 45k/100k       │
# │  ├── task/2-unit-tests   +128 -0   Cost: $2.35/$10       │
# │  └── task/3-api-docs     +89 -12   Agents: 3/5           │
# │                                                             │
# │  ⚠️  Potential Conflict: auth.go (Task #1 & #2)            │
# │                                                             │
# └─────────────────────────────────────────────────────────────┘

# 監控特定任務
mur ide watch 1

# 暫停/恢復
mur ide pause 2
mur ide resume 2

# 合併完成的任務
mur ide merge 3
mur ide merge --all

# 取消任務
mur ide cancel 1
```

### Web Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  mur.ide Dashboard                                    🔄 Live       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  📋 Active Tasks                                             │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │  #1 Code Review PR-123                    80% ████░  │    │   │
│  │  │  Agent: Claude  │  Branch: task/1-review │  +42 -15  │    │   │
│  │  │  ──────────────────────────────────────────────────  │    │   │
│  │  │  ✓ File analysis complete                            │    │   │
│  │  │  ✓ Security review done                              │    │   │
│  │  │  ⏳ Performance suggestions in progress              │    │   │
│  │  └─────────────────────────────────────────────────────┘    │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │  #2 Unit Tests for auth module            30% ██░░░  │    │   │
│  │  │  Agent: GPT-4  │  Branch: task/2-tests  │  +128 -0   │    │   │
│  │  │  ──────────────────────────────────────────────────  │    │   │
│  │  │  ✓ Test structure created                            │    │   │
│  │  │  ⏳ Writing test cases                               │    │   │
│  │  │  ○ Coverage check pending                            │    │   │
│  │  └─────────────────────────────────────────────────────┘    │   │
│  │                                                              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌───────────────────────┐  ┌───────────────────────────────────┐ │
│  │  🤖 Agent Pool        │  │  📊 Resource Usage                │ │
│  ├───────────────────────┤  ├───────────────────────────────────┤ │
│  │  Claude-A  ● busy     │  │  Tokens: ████████░░ 45k/100k      │ │
│  │  GPT-B     ● busy     │  │  Cost:   ██░░░░░░░░ $2.35/$10     │ │
│  │  Gemini-C  ● busy     │  │  Time:   ███████░░░ 35min/1hr     │ │
│  │  Claude-D  ○ idle     │  │                                   │ │
│  │  GPT-E     ○ idle     │  │  [Add Budget] [View History]      │ │
│  └───────────────────────┘  └───────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Configuration

```yaml
# ~/.murmur/ide.yaml

agents:
  pool_size: 5
  providers:
    claude:
      enabled: true
      max_concurrent: 2
      models: [claude-4-sonnet, claude-4-opus]
    openai:
      enabled: true
      max_concurrent: 2
      models: [gpt-4, gpt-4-turbo]
    gemini:
      enabled: true
      max_concurrent: 2
      models: [gemini-2.0-pro]

workspace:
  worktree_dir: .mur-worktrees
  auto_cleanup: true
  cleanup_after: 24h
  
tasks:
  max_concurrent: 5
  default_timeout: 30m
  auto_merge: false
  require_review: true
  
resources:
  max_tokens_per_hour: 100000
  max_cost_per_day: 10.00
  warn_at_percentage: 80

ui:
  refresh_interval: 2s
  show_logs: true
  notify_on_complete: true
  notify_on_conflict: true
```

## Integration with mur.core

```go
// IDE 使用 mur.core 來載入相關 patterns
type IDECoreIntegration struct {
    core       *core.Core
    commander  *commander.Commander
}

func (i *IDECoreIntegration) PrepareTask(task Task) (*PreparedTask, error) {
    // 1. 分類任務
    domains := i.core.Classify(ClassifyInput{
        Content: task.Description,
        Context: task.Context,
    })
    
    // 2. 檢索相關 patterns
    patterns := i.core.GetPatterns(domains)
    
    // 3. 使用 Commander 製作 prompt
    prompt, routing := i.commander.Craft(task.Description, patterns)
    
    return &PreparedTask{
        Task:     task,
        Prompt:   prompt,
        Patterns: patterns,
        Routing:  routing,
    }, nil
}
```

---

*This specification defines the parallel development environment for mur.*
