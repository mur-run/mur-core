---
name: openspec-driven-development
confidence: HIGH
score: 0.95
category: pattern
domain: _global
first_seen: 2026-02-01
last_seen: 2026-02-02
times_seen: 69
---

# OpenSpec 驅動開發：Spec → Implement → Archive

## Problem / Trigger
大型專案需要系統性地規劃和追蹤功能開發進度。

## Solution
用 OpenSpec change specs 驅動開發流程：
1. 在 `openspec/changes/` 寫 numbered spec（如 `51-notification-system.md`）
2. Spec 包含：Problem、Requirements、Files to Create/Modify、Acceptance Criteria
3. 用 AI CLI tool 實作：`ai_run.sh -p "Implement spec 51" --workdir ~/Projects/myapp`
4. 驗證：`swift build` zero warnings + `swift test` all pass
5. Commit → `mv spec to openspec/changes/archive/` → next spec

**批次作業效率極高** — BitL 專案兩天內完成 69 個 specs，從 758 到 1144 tests。

關鍵：spec 要寫清楚「要改哪些檔案」和「最少幾個新 tests」，AI 才能一次到位。

## Verification
BitL 69 specs 全部完成，zero build warnings，1144 tests pass。
