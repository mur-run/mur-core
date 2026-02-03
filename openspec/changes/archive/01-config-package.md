# 01: Config Package

**Status:** Done  
**Priority:** High  
**Effort:** Medium (2-3 hours)

## Problem

目前所有設定操作都是 stub 或 hardcoded：
- `mur run` hardcoded default tool 為 "claude"
- `mur config set` 印出 "not implemented"
- `mur config default` 印出 "not implemented"
- 沒有統一的 config 讀寫邏輯

需要一個 `internal/config` package 作為其他功能的基礎。

## Solution

建立 `internal/config/` package 提供：
1. **Config struct** — 對應 `~/.murmur/config.yaml` 結構
2. **Load()** — 讀取並解析 config
3. **Save()** — 寫入 config（保留 comments）
4. **GetDefault/SetDefault** — 讀寫 default_tool
5. **GetTool/SetTool** — 工具設定操作

## Implementation

### Files to Create

```
internal/
  config/
    config.go       # Config struct + Load/Save
    config_test.go  # Unit tests
```

### Config Struct

```go
// internal/config/config.go
package config

type Config struct {
    DefaultTool string           `yaml:"default_tool"`
    Tools       map[string]Tool  `yaml:"tools"`
    Learning    LearningConfig   `yaml:"learning"`
    MCP         MCPConfig        `yaml:"mcp"`
}

type Tool struct {
    Enabled bool     `yaml:"enabled"`
    Binary  string   `yaml:"binary"`
    Flags   []string `yaml:"flags"`
}

type LearningConfig struct {
    AutoExtract  bool `yaml:"auto_extract"`
    SyncToTools  bool `yaml:"sync_to_tools"`
    PatternLimit int  `yaml:"pattern_limit"`
}

type MCPConfig struct {
    SyncEnabled bool                   `yaml:"sync_enabled"`
    Servers     map[string]interface{} `yaml:"servers"`
}
```

### Core Functions

```go
// ConfigPath returns ~/.murmur/config.yaml
func ConfigPath() (string, error)

// Load reads and parses the config file
func Load() (*Config, error)

// Save writes config back to file
func (c *Config) Save() error

// GetDefaultTool returns the default tool name
func (c *Config) GetDefaultTool() string

// SetDefaultTool updates the default tool
func (c *Config) SetDefaultTool(tool string) error

// GetTool returns tool config by name
func (c *Config) GetTool(name string) (*Tool, bool)

// EnsureTool validates tool exists and is enabled
func (c *Config) EnsureTool(name string) error
```

### Update Commands

修改以下 commands 使用新的 config package：

1. **cmd/run.go**
   ```go
   // Before
   if tool == "" {
       tool = "claude" // TODO: read from config
   }
   
   // After
   if tool == "" {
       cfg, _ := config.Load()
       tool = cfg.GetDefaultTool()
   }
   ```

2. **cmd/config.go**
   - `config set` 使用 `config.Load()` + modify + `config.Save()`
   - `config default` 使用 `config.SetDefaultTool()`

## Tests

```go
// internal/config/config_test.go

func TestLoad(t *testing.T)           // 測試正常讀取
func TestLoadMissing(t *testing.T)    // 測試檔案不存在
func TestSave(t *testing.T)           // 測試寫入
func TestGetDefaultTool(t *testing.T) // 測試預設工具
func TestSetDefaultTool(t *testing.T) // 測試設定預設工具
func TestGetTool(t *testing.T)        // 測試取得工具設定
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `go test ./internal/config/...` 全部通過
- [x] `mur run -p "test"` 使用 config 中的 default_tool
- [x] `mur config default gemini` 實際修改 config.yaml
- [x] `mur config set default_tool auggie` 實際修改 config.yaml

## Dependencies

- `gopkg.in/yaml.v3` (已在 go.mod)

## Related

- Phase 1 of VISION.md (Unified Hooks) 也需要讀寫設定
- `mur sync` 需要知道哪些工具 enabled
