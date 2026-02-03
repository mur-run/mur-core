# Murmur 🔊

**Multi-AI CLI 統一管理層 + 跨工具學習系統**

每個 AI CLI tool 都是獨立的孤島。Murmur 統一它們。

## Features

- **Multi-tool runner** — 一個指令跑任何 AI，不用記每個工具的 flag
- **MCP 統一管理** — 設定一次同步全部
- **跨工具學習** — Claude 學到的，Gemini 也會
- **Team 知識庫** — 團隊共享 patterns，新人自動繼承經驗
- **成本路由** — 簡單任務自動走免費工具，複雜的走 Claude

## Installation

```bash
# Go install
go install github.com/karajanchang/murmur-ai/cmd/mur@latest

# Or download binary from releases
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-darwin-arm64.tar.gz | tar xz
sudo mv mur /usr/local/bin/
```

## Quick Start

```bash
# Initialize
mur init

# Check available AI tools
mur health

# Run a prompt
mur run -p "explain this code"

# Run with specific tool
mur run -t gemini -p "write a haiku"

# List learned patterns
mur learn list

# Sync to all AI tools
mur sync all
```

## Commands

| Command | Description |
|---------|-------------|
| `mur init` | Initialize configuration |
| `mur run -p "prompt"` | Run prompt with default AI |
| `mur run -t claude -p "prompt"` | Run with specific tool |
| `mur config show` | Show configuration |
| `mur config default claude` | Set default tool |
| `mur health` | Check AI tool availability |
| `mur learn list` | List learned patterns |
| `mur learn sync` | Sync patterns to AI tools |
| `mur sync all` | Sync everything |
| `mur sync mcp` | Sync MCP configuration |

## Supported AI Tools

| Tool | Status |
|------|--------|
| Claude Code | ✅ Supported |
| Gemini CLI | ✅ Supported |
| Auggie | 🔜 Coming |
| Codex | 🔜 Coming |
| OpenCode | 🔜 Coming |

## Configuration

Config file: `~/.murmur/config.yaml`

```yaml
default_tool: claude

tools:
  claude:
    enabled: true
    binary: claude
  gemini:
    enabled: true
    binary: gemini

learning:
  auto_extract: true
  sync_to_tools: true
```

## License

MIT
