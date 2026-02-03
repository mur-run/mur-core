# murmur-ai 跨 CLI 切換計劃

**Created:** 2026-02-03  
**Status:** 規劃中

## 各 CLI Tool 功能完整度矩陣

|                  | Claude Code | Gemini CLI | Auggie | Codex | OpenCode |
|------------------|-------------|------------|--------|-------|----------|
| Hooks            | ✅ native   | ✅ native  | ✅     | ❌    | ❌       |
| Superpowers      | ✅ native   | ❌         | ❌     | ⚠️手動 | ⚠️手動   |
| MCP              | ✅          | ✅         | ✅     | ❌    | ✅       |
| Native Skills    | ✅          | ✅         | ✅     | ❌    | ❌       |
| murmur-ai hooks  | ✅ 已完成   | 🔜 可做    | 🔜 可做 | ❌    | ❌       |
| murmur-ai skills | ✅ 已完成   | 🔜 可做    | 🔜 可做 | ⚠️注入 | ⚠️注入   |
| 價格             | Anthropic   | 免費       | 免費   | OpenAI | 任意LLM  |

---

## 切換計劃：讓任何 CLI 都有完整體驗

### Phase 1：統一 hooks 設定（murmur-ai 跨 CLI）

**新增:** `scripts/hooks_sync.sh`

**功能：** 讀 `hooks/claude-code-hooks.json` 作為 source of truth，自動寫入各 CLI 的 settings.json

```
hooks/claude-code-hooks.json (source of truth)
            │
            │ hooks_sync.sh 自動轉換 + 寫入：
            │
    ┌───────┼───────────┬──────────────┐
    ▼       ▼           ▼              ▼
~/.claude/  ~/.gemini/  ~/.augment/
settings    settings    settings
.json       .json       .json
```

**Event mapping:**

| murmur-ai         | Gemini       | Auggie       |
|-------------------|--------------|--------------|
| UserPromptSubmit  | BeforeAgent  | SessionStart |
| Stop              | AfterAgent   | Stop         |

---

### Phase 2：統一 skills 同步

**更新:** `scripts/sync_skills.sh`

learned patterns → 各 CLI 的 skills 目錄：

```
~/clawd/skills/murmur-ai/learned/**/*.md
    │
    ├──→ ~/.claude/skills/learned-*/SKILL.md    (已完成)
    ├──→ ~/.gemini/skills/learned-*.md          (新增)
    ├──→ ~/.augment/commands/learned-*.md       (新增)
    ├──→ .codex/instructions.md                 (append)
    └──→ .opencode/skills/learned-*.md          (新增)
```

---

### Phase 3：統一 MCP 同步

**已有:** `scripts/mcp_sync.sh`

```
config/mcp.json → 各 CLI 的 MCP 設定：
    │
    ├──→ ~/.claude/settings.json mcpServers     (已完成)
    ├──→ ~/.gemini/settings.json mcpServers     (新增)
    └──→ ~/.augment/settings.json mcpServers    (新增)
```

---

### Phase 4：Superpowers 替代方案（給非 Claude Code）

Superpowers 的核心是 skills (方法論)。對於 Gemini/Auggie，可以：

| 方案 | 做法 | 說明 |
|------|------|------|
| **A** | 把 Superpowers 的 SKILL.md 複製到各 CLI 的 skills 目錄 | 不依賴 plugin 系統，直接當 skills 讀 |
| **B** | 用 SessionStart hook 注入 using-superpowers 內容 | 跟 Superpowers 在 Claude Code 的做法一樣 |

**推薦方案 B**

**新增:** `scripts/superpowers_sync.sh`
- 讀 Superpowers plugin 的 skills/
- 寫入 Gemini/Auggie 的 skills 目錄

---

### Phase 5：一鍵切換 default tool

**更新:** `scripts/ai_config.sh default <tool>`

切換時自動：
1. 更新 `config/tools.yaml`
2. 執行 `hooks_sync.sh` → hooks 就位
3. 執行 `sync_skills.sh` → patterns 就位
4. 執行 `mcp_sync.sh` → MCP 就位
5. 執行 `superpowers_sync.sh` → 方法論就位

**結果：** 切換到任何 CLI 都有完整體驗

---

## 切換後的完整度預估

|                  | Claude Code | Gemini CLI | Auggie | Codex | OpenCode |
|------------------|-------------|------------|--------|-------|----------|
| Hooks            | ✅          | ✅         | ✅     | ❌    | ❌       |
| Superpowers      | ✅ native   | ✅ sync    | ✅ sync | ⚠️    | ⚠️       |
| MCP              | ✅          | ✅         | ✅     | ❌    | ✅       |
| Learned patterns | ✅          | ✅         | ✅     | ⚠️    | ⚠️       |
| **自動化程度**   | **100%**    | **95%**    | **95%** | **30%** | **50%**  |

---

## 下一步

建議先做 **Phase 1（hooks_sync.sh）**，這是最核心的，做完就能切 Gemini/Auggie 用了。
