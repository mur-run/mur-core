# murmur-ai Vision

## What is murmur-ai?

**Multi-AI CLI 統一管理層 + 跨工具學習系統**

每個 AI CLI tool 都是獨立的孤島。murmur-ai 統一它們。

## Core Features

1. **Multi-tool runner** — 一個指令跑任何 AI，不用記每個工具的 flag
2. **MCP 統一管理** — 設定一次同步全部
3. **跨工具學習** — Claude 學到的，Gemini 也會
4. **Team 知識庫** — 團隊共享 patterns，新人自動繼承經驗
5. **成本路由** — 簡單任務自動走免費工具，複雜的走 Claude

## Architecture

```
murmur-ai (Go CLI)
    ├── 學習 patterns（從 coding sessions）
    ├── 輸出到 Claude Code（~/.claude/ learnings, settings）
    ├── 輸出到 OpenClaw skills（如果需要）
    └── 輸出到其他 CLI（Gemini, Auggie 等）
```

Go 版本是 **source of truth**，它可以 generate 各種格式的 output。

## Supported AI Tools

| Tool | Status |
|------|--------|
| Claude Code | ✅ Supported |
| Gemini CLI | ✅ Supported |
| Auggie | 🔜 Coming |
| Codex | 🔜 Coming |
| OpenCode | 🔜 Coming |

## Target Completeness

|                  | Claude Code | Gemini CLI | Auggie | Codex | OpenCode |
|------------------|-------------|------------|--------|-------|----------|
| Hooks            | ✅          | ✅         | ✅     | ❌    | ❌       |
| Superpowers      | ✅ native   | ✅ sync    | ✅ sync | ⚠️    | ⚠️       |
| MCP              | ✅          | ✅         | ✅     | ❌    | ✅       |
| Learned patterns | ✅          | ✅         | ✅     | ⚠️    | ⚠️       |
| **自動化程度**   | **100%**    | **95%**    | **95%** | **30%** | **50%**  |

## Cross-CLI Sync Plan

### Phase 1: Unified Hooks
- Source of truth: `~/.murmur/hooks.json`
- Auto-sync to Claude/Gemini/Auggie settings

### Phase 2: Skills Sync  
- Learned patterns → each CLI's skills directory

### Phase 3: MCP Sync
- `~/.murmur/mcp.json` → each CLI's MCP config

### Phase 4: Superpowers Alternatives
- Inject methodology to non-Claude CLIs via skills

### Phase 5: One-click Default Switch
- `mur config default gemini` → everything syncs automatically
