# 10: Stats Package

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

沒有追蹤工具使用量、成本和效率的機制：
- 不知道每個工具被使用多少次
- 無法估算成本（Claude 是付費的，Gemini/Auggie 是免費的）
- 無法分析 auto-routing 的效果
- 無法查看使用趨勢

需要一個 `internal/stats` package 來追蹤和分析使用數據。

## Solution

建立 `internal/stats/` package 提供：
1. **UsageRecord struct** — 記錄單次使用
2. **Append-only log** — 儲存到 `~/.murmur/stats.jsonl`
3. **Record()** — 記錄一次使用
4. **Query()** — 查詢和過濾統計

新增 `mur stats` 命令來顯示統計資料。

## Implementation

### Files to Create

```
internal/
  stats/
    stats.go        # UsageRecord struct + Record/Query
    stats_test.go   # Unit tests

cmd/mur/cmd/
  stats.go          # stats command
```

### UsageRecord Struct

```go
// internal/stats/stats.go
package stats

type UsageRecord struct {
    Tool         string    `json:"tool"`
    Timestamp    time.Time `json:"timestamp"`
    PromptLength int       `json:"prompt_length"`
    DurationMs   int64     `json:"duration_ms"`
    CostEstimate float64   `json:"cost_estimate"` // 0 for free tools
    Tier         string    `json:"tier"`          // free | paid
    RoutingMode  string    `json:"routing_mode"`  // auto | manual | cost-first | quality-first
    AutoRouted   bool      `json:"auto_routed"`   // true if router selected this tool
    Complexity   float64   `json:"complexity"`    // from analyzer
    Success      bool      `json:"success"`       // did the command succeed
}
```

### Core Functions

```go
// StatsPath returns ~/.murmur/stats.jsonl
func StatsPath() (string, error)

// Record appends a usage record to the stats file
func Record(record UsageRecord) error

// QueryFilter for filtering records
type QueryFilter struct {
    Tool      string
    StartTime time.Time
    EndTime   time.Time
    Tier      string
}

// Query reads and filters usage records
func Query(filter QueryFilter) ([]UsageRecord, error)

// Summary aggregates usage data
type Summary struct {
    TotalRuns      int
    ByTool         map[string]ToolStats
    EstimatedCost  float64
    EstimatedSaved float64 // money saved by using free tools
    AutoRouteStats AutoRouteStats
}

type ToolStats struct {
    Count        int
    TotalTimeMs  int64
    AvgTimeMs    int64
    TotalCost    float64
}

type AutoRouteStats struct {
    Total     int
    ToFree    int
    ToPaid    int
    FreeRatio float64
}

// Summarize computes summary from records
func Summarize(records []UsageRecord) Summary

// Reset clears all stats
func Reset() error
```

### Cost Estimation

```go
// Rough estimates per 1K prompt chars
var costPerKChars = map[string]float64{
    "claude": 0.003,  // ~$3/M input tokens, rough estimate
    "gemini": 0.0,    // free
    "auggie": 0.0,    // free
}
```

### Commands

#### `mur stats`
顯示總覽統計：
- 總使用次數
- 各工具使用次數和時間
- 估算成本
- 路由決策統計
- ASCII 趨勢圖

#### `mur stats --tool claude`
特定工具統計

#### `mur stats --period week`
特定時間範圍（day/week/month/all）

#### `mur stats --json`
JSON 輸出

#### `mur stats reset`
清除所有統計資料

### Integration with run.go

```go
// In runExecute, wrap execution and record stats
startTime := time.Now()
err := execCmd.Run()
duration := time.Since(startTime)

stats.Record(stats.UsageRecord{
    Tool:         tool,
    Timestamp:    startTime,
    PromptLength: len(prompt),
    DurationMs:   duration.Milliseconds(),
    CostEstimate: stats.EstimateCost(tool, len(prompt)),
    Tier:         toolCfg.Tier,
    RoutingMode:  cfg.Routing.Mode,
    AutoRouted:   forceTool == "",
    Complexity:   selection.Analysis.Complexity,
    Success:      err == nil,
})
```

### ASCII Chart

```
Usage by Day (last 7 days)
Mon  ████████████████████ 20
Tue  ██████████████ 14
Wed  ████████████████████████ 24
Thu  ████████ 8
Fri  ████████████ 12
Sat  ██ 2
Sun  ████ 4

Tool Distribution
claude  ████████████████ 40%
gemini  ████████████████████████ 55%
auggie  ██ 5%
```

## Testing

```go
func TestRecord(t *testing.T)
func TestQuery(t *testing.T)
func TestSummarize(t *testing.T)
func TestReset(t *testing.T)
```

## Success Criteria

- [x] `mur run` 後自動記錄使用資料
- [x] `mur stats` 顯示清晰的統計總覽
- [x] `mur stats --tool X` 過濾特定工具
- [x] `mur stats --period X` 過濾時間範圍
- [x] `mur stats --json` 輸出 JSON
- [x] `mur stats reset` 清除資料
- [x] 顯示估算節省的成本
- [x] `go build ./...` 成功
- [x] Unit tests 通過
