# mur-core AI Tools Test Plan

測試所有 AI 工具與 mur-core 的整合。

## 測試目標

1. 確認 patterns 正確同步到各工具
2. 確認各工具能讀取並使用 patterns
3. 確認 hooks 正確觸發（Claude/Gemini）

---

## 工具清單

| Tool | Type | Sync Path | Status |
|------|------|-----------|--------|
| Claude Code | CLI + Hooks | `~/.claude/skills/mur-patterns.md` | ⬜ |
| Gemini CLI | CLI + Hooks | `~/.gemini/skills/mur-patterns.md` | ⬜ |
| Codex | CLI | `~/.codex/instructions.md` | ⬜ |
| Auggie | CLI | `~/.augment/skills/mur-patterns.md` | ⬜ |
| Aider | CLI | `~/.aider/conventions.md` | ⬜ |
| Continue | IDE | `~/.continue/rules/mur-patterns.md` | ⬜ |
| Cursor | IDE | `~/.cursor/rules/mur-patterns.md` | ⬜ |
| Windsurf | IDE | `~/.windsurf/rules/mur-patterns.md` | ⬜ |

---

## Pre-Test Setup

```bash
# 1. 確認 mur 安裝
mur version

# 2. 確認 patterns 存在
mur status

# 3. Sync patterns
mur sync

# 4. 安裝 hooks（如果還沒）
mur init --hooks
```

---

## Test Cases

### T1: Claude Code

**Prerequisites:**
- Claude Code installed (`claude --version`)
- `~/.claude/skills/mur-patterns.md` exists

**Test Steps:**
```bash
# 1. 檢查 skills 目錄
ls ~/.claude/skills/

# 2. 測試 Claude 能讀取 patterns
cd ~/Projects/BitL  # Swift project
claude "Based on your loaded skills, what Swift patterns do you know about?"

# 3. 測試 pattern injection（如果有 hooks）
claude "fix a SwiftUI navigation bug"
# → 應該會注入相關 Swift/SwiftUI patterns
```

**Expected:**
- [ ] Claude 能列出 mur patterns
- [ ] Swift 相關 patterns 被提及（如 `swiftui-main-actor-task-detached`）

---

### T2: Gemini CLI

**Prerequisites:**
- Gemini CLI installed (`gemini --version`)
- `~/.gemini/skills/mur-patterns.md` exists

**Test Steps:**
```bash
# 1. 檢查 skills
ls ~/.gemini/skills/

# 2. 測試讀取
gemini "What skills or patterns do you have access to?"

# 3. 測試 context injection
cd ~/Projects/mur-core  # Go project
gemini "How should I handle errors in this Go project?"
```

**Expected:**
- [ ] Gemini 能識別 patterns
- [ ] Go 相關 patterns 被應用

---

### T3: Codex CLI

**Prerequisites:**
- Codex installed (`codex --version`) or install: `npm i -g @openai/codex`
- `~/.codex/instructions.md` exists

**Test Steps:**
```bash
# 1. 安裝（如果需要）
npm i -g @openai/codex

# 2. 檢查 instructions
cat ~/.codex/instructions.md | head -20

# 3. 測試
codex "Show me what instructions you're following"
```

**Expected:**
- [ ] Codex 能讀取 instructions.md
- [ ] Patterns 影響回應

---

### T4: Auggie

**Prerequisites:**
- Auggie installed (`auggie --version`)
- `~/.augment/skills/mur-patterns.md` exists

**Test Steps:**
```bash
# 1. 檢查 skills
ls ~/.augment/skills/

# 2. 測試
auggie "What skills do you have loaded?"

# 3. Context test
auggie "Help me with a Swift error handling pattern"
```

**Expected:**
- [ ] Auggie 能識別 skills
- [ ] 回應包含相關 patterns

---

### T5: Aider

**Prerequisites:**
- Aider installed (`aider --version`) or install: `pip install aider-chat`
- `~/.aider/conventions.md` exists

**Test Steps:**
```bash
# 1. 安裝
pip install aider-chat

# 2. 檢查 conventions
cat ~/.aider/conventions.md | head -20

# 3. 測試
cd ~/Projects/BitL
aider --message "What conventions should I follow?"
```

**Expected:**
- [ ] Aider 讀取 conventions.md
- [ ] 回應反映 patterns

---

### T6: Continue (VS Code Extension)

**Prerequisites:**
- Continue extension installed in VS Code
- `~/.continue/rules/mur-patterns.md` exists

**Test Steps:**
1. Open VS Code
2. Open Continue sidebar
3. Ask: "What rules do you have loaded?"
4. Test in a Swift file: "Help me fix this error"

**Expected:**
- [ ] Continue 能讀取 rules
- [ ] Patterns 影響 suggestions

---

### T7: Cursor

**Prerequisites:**
- Cursor IDE installed
- `~/.cursor/rules/mur-patterns.md` exists

**Test Steps:**
1. Open Cursor
2. Open a Swift project
3. Use Cmd+K or chat: "What patterns should I follow?"
4. Test: "Fix this SwiftUI view"

**Expected:**
- [ ] Cursor 讀取 rules
- [ ] Patterns 影響回應

---

### T8: Windsurf

**Prerequisites:**
- Windsurf IDE installed
- `~/.windsurf/rules/mur-patterns.md` exists

**Test Steps:**
1. Open Windsurf
2. Open a project
3. Ask Cascade: "What rules are you following?"

**Expected:**
- [ ] Windsurf 讀取 rules
- [ ] Patterns 被應用

---

## Hooks Test

### H1: Claude UserPromptSubmit Hook

**Test:**
```bash
# 確認 hook 設定
cat ~/.claude/settings.json | jq '.hooks'

# 測試 context injection
cd ~/Projects/BitL
claude "fix a bug"
# → mur context 應該被注入
```

**Expected:**
- [ ] Hook 在 settings.json 中
- [ ] `mur context` 被呼叫
- [ ] 相關 patterns 被注入到 prompt

---

### H2: Claude SessionEnd Hook

**Test:**
```bash
# 確認 hook
cat ~/.claude/settings.json | jq '.hooks'

# 完成一個 session 後
# → mur learn extract 應該自動執行
```

**Expected:**
- [ ] Session 結束時自動提取 patterns
- [ ] 新 patterns 儲存到 ~/.mur/patterns/

---

### H3: Gemini Hooks

**Test:**
```bash
# 確認 hook 設定（如果 Gemini 支援）
cat ~/.gemini/settings.json 2>/dev/null

# 測試
gemini "help me debug"
```

**Expected:**
- [ ] Hook 正確配置
- [ ] Patterns 被注入

---

## Summary Checklist

**Tested: 2026-02-08**

| Tool | Sync ✓ | Read ✓ | Use ✓ | Hooks ✓ | Notes |
|------|--------|--------|-------|---------|-------|
| Claude Code | ✅ | ⬜ | ⬜ | ⬜ | 手動測 (CLI 超時) |
| Gemini CLI | ✅ | ✅ | ⬜ | ✅ | Hooks confirmed in logs |
| Codex | ✅ | ❌ | ❌ | N/A | 需要 OPENAI_API_KEY |
| **Auggie** | ✅ | ✅ | ✅ | N/A | **確認成功** - 列出 patterns |
| Aider | ✅ | ❌ | ❌ | N/A | 安裝失敗 (numpy) |
| Continue | ✅ | ⬜ | ⬜ | N/A | IDE 手動測 |
| Cursor | ✅ | ⬜ | ⬜ | N/A | IDE 手動測 |
| Windsurf | ✅ | ⬜ | ⬜ | N/A | IDE 手動測 |

---

## Installation Commands

```bash
# Codex
npm i -g @openai/codex

# Aider
pip install aider-chat

# Auggie (if not installed)
npm i -g @anthropics/auggie

# Claude Code
npm i -g @anthropic-ai/claude-code

# Gemini CLI
npm i -g @anthropic-ai/gemini-cli
```

---

## Troubleshooting

### Patterns not showing

```bash
# Re-sync
mur sync

# Check file exists
ls -la ~/.claude/skills/mur-patterns.md

# Check content
head -20 ~/.claude/skills/mur-patterns.md
```

### Hooks not working

```bash
# Re-install hooks
mur init --hooks

# Check settings
cat ~/.claude/settings.json | jq '.hooks'
```

### Permission issues

```bash
# Fix permissions
chmod 644 ~/.claude/skills/mur-patterns.md
```
