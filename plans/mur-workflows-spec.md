# MUR Workflows Spec

> Date: 2026-02-24
> Status: Draft
> Authors: David + Clawd

## Overview

Workflows 是從 mur session 錄製的可重播操作流程，與現有 Patterns（知識片段/偏好）分開管理。

## Concepts

| | Patterns | Workflows |
|---|---|---|
| 來源 | transcript 自動提取 | session 錄製 (in/out) |
| 本質 | 知識片段、偏好 | 可重播的操作流程 (SOP) |
| 預設分享 | 社群 ✅ | 私有 🔒 |
| 可編輯 | 有限 | 完整編輯（切割/合併） |
| 未來 | 免費為主 | 可販售 💰 |

### Data Model

- **Session** — 原始錄製，不可變（immutable recording）
- **Workflow** — 從 session 切出或組合的產物（editable, versionable）
  - 一個 session 可切出多個 workflows
  - 多個 workflows 可合併為一個（Phase 2）
  - Workflow 引用 source session(s) 但獨立存在

## Permission Model

三層權限，先不做獨立 admin role（owner = admin）：

| 權限 | 說明 |
|---|---|
| **read** | 可以查看完整內容、可以使用 |
| **write** | 可以編輯 workflow |
| **execute-only** | 只能透過 Commander 執行，看不到實作細節 |

Execute-only 是核心差異化功能：
- Teams 場景：員工能跑 SOP 但看不到 know-how
- Marketplace 場景：買家能用但不能複製轉賣

## Pricing

### 統一 Cloud Sync 計價

```
Free:
  - Patterns: 本地無限 + 社群同步
  - Workflows: 本地錄製/分析/播放（無限）
  - Cloud Sync: ❌

Pro ($12/mo, Launch Sale $9/mo):
  - Cloud Sync: Patterns + Workflows（統一，無上限）
  - 個人使用，無 Teams 功能
  - Device 限 3 個

Teams ($49/mo flat, 5 members included, +$10/extra member):
  - Cloud Sync: 無限
  - 完整權限控制 (read/write/execute-only per workflow)
  - Team Workflow Library（共享 workflow 集合，CLI + Web UI）
  - Admin dashboard

Commander Add-on ($15-20/mo, 需 Pro 以上):
  - 閉源產品（不 open source）
  - 代理執行 workflows
  - 搭配 server sync 使用

Marketplace (未來):
  - 賣家需 Pro 以上
  - 平台抽成: 20-30%
  - 支援 execute-only 販售模式（買家能跑不能看 source）
```

### Team Workflow Library
Team 共享的 workflow 集合，成員可瀏覽、搜尋、使用：
```bash
mur workflows list --team        # 看團隊的 workflows
mur workflows pull <id>          # 拉到本地用
mur workflows push --team        # 分享給團隊
```
API 後續再開放（供 CI/CD、外部系統整合）。

### 為什麼 Commander 是 add-on 不是獨立產品
- 降低購買摩擦（用戶已在生態系內）
- 避免維護兩套 billing/onboarding
- mur-core 開源建立社群 → Commander 閉源作為商業護城河

## Marketplace IP 保護

### Execute-only 模式
- 買家拿到封裝過的 workflow（只有 metadata：名稱、描述、input/output schema）
- 執行時透過 Commander 從 server 拉 workflow 到 runtime，不落地到買家本地
- 買家看得到：名稱、描述、inputs、outputs、評價
- 買家看不到：具體步驟、prompts、邏輯

### 保護策略（務實路線）
1. **Execute-only** 擋住 casual copying（90% 的人）
2. **License agreement** 法律層面保護
3. **持續更新** — 買家買的是 subscription，不是一次性檔案；raw copy 很快過時

不做複雜 DRM，等 marketplace 有量再評估。

## Implementation Priority

### Phase 1 🔴 — Core
1. Workflow 資料結構設計（區別於 Session）
2. Session → Workflow 切割功能（標記起點/終點）
3. 本地錄製/播放
4. mur-server Workflows 後台（獨立於 Patterns）

### Phase 2 🟡 — Cloud & Teams
5. Cloud Sync — 統一 Patterns + Workflows（$12/mo Pro）
6. Teams 權限控制（read/write/execute-only）
7. Workflow 合併功能

### Phase 3 🟢 — Commander & Marketplace
8. MUR Commander（閉源 add-on，$15-20/mo）
9. Marketplace 基礎架構
10. 販售/抽成系統

## Workflow 版本控制

簡單版本 + 發布概念（不做 git-like）：
- 每次編輯自動存 snapshot（內部遞增 revision number）
- 用戶看到的是 **published version**：v1, v2, v3
- 只有明確「發布」才會產生新版本號
- 保留最近 N 個 revision（Free 5 個，Pro 無限）

模式類似 Figma — auto-save + 手動 version milestone。

## Commander 架構

Server/Client 分離，一個 Server 對應多個 Client：

```
Commander Server (Daemon):
  - 本地 or 雲端部署
  - 排程/觸發 workflow 執行
  - 管理多個 Client 連線
  - API endpoint
  - Teams 場景：公司架一個 Server，員工各裝 Client

Commander Client (GUI):
  - 用戶端安裝
  - 監控執行狀態
  - 觸發/暫停/取消 workflow
  - 多個 Client → 一個 Server
```

### Adapter 層（核心價值，閉源）

Workflow 使用 tool-agnostic 中間格式，執行時由 adapter 翻譯：

```
Workflow Step (abstract):
  intent: "refactor function"
  input: { file: "auth.go", function: "Login" }
  expected: { type: "code_change" }
       ↓
Adapter (Claude):  → Claude Code API call
Adapter (GPT):     → OpenAI API call
Adapter (Gemini):  → Gemini API call
```

跨 AI tool portability 是 MUR 的核心定位，adapter 層是 Commander 閉源的技術理由。

## Marketplace Review/Quality

### Phase 1 — 社群自治
- ⭐ 評分 + 文字評價
- 下載/使用次數顯示
- 賣家 profile（驗證帳號）
- 舉報機制（spam/broken/malicious）

### Phase 2 — 輕度審核
- 自動化檢查（workflow 格式正確、能執行）
- Staff picks / featured（人工推薦）
- 「Verified」badge 給高品質賣家

不做 App Store 級審核，靠社群 + 自動化。

## Execution Limits

不限制 workflow 執行次數。原因：執行資源是用戶自己的（機器 + API key），不花我們的錢。

```
資料流：
Commander Server (用戶機器)
  ↕ mur-server API (我們的雲端)
  │  - Workflow sync（拉/推 workflow 內容）
  │  - License 驗證
  │  - Marketplace workflow 下載
  │  - 使用量回報（analytics）
  ↕ AI Provider APIs (用戶自己的 key)
  │  - Claude / GPT / Gemini
  │  - 執行成本 = 用戶負擔
```

用戶付我們的是**軟體授權 + 雲端同步**，不是算力。

所有 tier 都提供：
- 執行歷史記錄
- 用量統計（`mur stats --workflows`）
- 成本追蹤（各 AI tool 花了多少錢）

### 離線支援
Commander 連不上 mur-server API 時，已 sync 到本地的 workflow 仍可繼續執行（local cache）。

### 未來：雲端託管版
若之後推出 MUR 代管的 Commander Server（我們出算力），再用 execution 計價。目前不需要。

## CLI Commands

### Session（錄製階段，已實作）
```bash
mur session start [--source claude-code]
mur session stop [--analyze] [--open]
mur session status
mur session list
mur session analyze <session-id>
mur session ui <session-id>          # 互動式 Web UI
mur session export <session-id>
```

### Workflows（Phase 1 新增）
```bash
mur workflows list                   # 列出本地 workflows
mur workflows show <id>              # 查看詳情
mur workflows create --from-session <session-id> [--start <event> --end <event>]
mur workflows edit <id>              # 開 Web UI 編輯
mur workflows run <id>               # 本地播放/執行
mur workflows export <id> [--format skill|yaml|md]
mur workflows delete <id>
mur workflows publish <id>           # 產生新版本號

# Phase 2: Cloud & Teams
mur workflows sync                   # 同步到 mur-server
mur workflows list --team            # 團隊 workflows
mur workflows push --team            # 分享給團隊
mur workflows pull <id>              # 從團隊/marketplace 拉到本地
mur workflows share <id> --user <email> --permission read|write|execute-only

# Phase 3: Marketplace
mur workflows marketplace list       # 瀏覽 marketplace
mur workflows marketplace publish <id> [--price <amount>] [--execute-only]
```

### Storage 結構
```
~/.mur/
├── session/
│   ├── active.json                  # 目前錄製狀態
│   └── recordings/
│       └── <session-id>.jsonl       # 原始錄製（immutable）
├── workflows/
│   ├── index.json                   # 本地 workflow 索引
│   └── <workflow-id>/
│       ├── workflow.yaml            # workflow 定義
│       ├── metadata.json            # 名稱、描述、版本、權限
│       └── revisions/              # auto-save snapshots
│           ├── rev-001.yaml
│           └── rev-002.yaml
└── ...
```

## Implementation Status

### ✅ Phase 1 — Core (Done 2026-02-24)
- Workflow 資料結構 (types.go)
- Session → Workflow 切割 (extract.go)
- 本地 CRUD + revisions (store.go)
- CLI: `mur workflows list/show/create/run/export/delete/publish`
- Server API endpoints (7 個 REST endpoints)
- 19 unit tests

### ✅ Phase 2 — Cloud & Teams (Done 2026-02-24)
- Cloud Sync types + client methods (push/pull/status)
- Permission model: read/write/execute-only
- Workflow merge (合併多個 workflows)
- CLI: `mur wf sync/share/merge`
- 12 new tests (31 total)

### ⏳ Phase 3 — Commander & Marketplace (TODO, 等有用戶再做)
> 收費功能等有用戶基礎再做
- [ ] Commander Server/Client 架構設計（需要更詳細 spec）
- [ ] Adapter 層（Claude/GPT/Gemini 跨 AI tool portability）
- [ ] Marketplace 基礎架構（社群評分、瀏覽、搜尋）
- [ ] 販售/抽成系統（20%，早期賣家 10%）
- [ ] Execute-only 封裝（strip 實作細節）
- [ ] `mur wf marketplace list/publish/search` CLI

### ⏳ Phase 4 — Developer Experience & Community (TODO)
- [ ] Workflows Web UI（瀏覽/編輯/視覺化 steps，像 `mur session ui`）
- [ ] `mur workflows import <url|file>` — 從 GitHub/URL 匯入別人的 workflow
- [ ] Community 免費分享（開源社群版，非 marketplace）
- [ ] Workflow templates（常見 SOP 模板：deploy, code review, debug...）
- [ ] `mur workflows watch` — 監聽檔案變更自動更新 workflow
- [ ] Analytics dashboard（哪些 workflows 最常跑、成功率、平均耗時）

## Decisions

- **Commander Server 最低配備：** Mac Mini M4 16G
- **Adapter 初期支援：** Claude + Gemini + GPT
- **Marketplace 抽成：** 20%（早期賣家前 100 名第一年 10%，高營收賣家月銷 >$1000 可談 15%）
