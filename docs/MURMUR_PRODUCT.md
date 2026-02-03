# Murmur — 產品規劃

**Created:** 2026-02-03  
**Updated:** 2026-02-03  
**Status:** 規劃中  
**Tagline:** Multi-AI CLI 統一管理層 + 跨工具學習系統

---

## 品牌

- **產品名：** Murmur
- **CLI 指令：** `mur`（3 個字母，打字快）
- **安裝：** `go install github.com/karajanchang/murmur@latest` 或 brew

---

## 市場定位

目前**沒有人**做這件事 — 每個 AI CLI tool 都是獨立的孤島。

> **統一管理層 + 跨工具學習系統**  
> 就像 Docker Desktop 統一了 container runtime 一樣。

---

## 核心賣點

| 功能 | 為什麼值錢 |
|------|-----------|
| **Multi-tool runner** | 一個指令跑任何 AI，不用記每個工具的 flag |
| **MCP 統一管理** | 設定一次同步全部，目前沒人做 |
| **跨工具學習** | 用 Claude 學到的，Gemini 也會 — 獨家 |
| **Team 知識庫** | 團隊共享 patterns，新人自動繼承所有經驗 |
| **成本路由** | 簡單任務自動走免費 Gemini，複雜的走 Claude |

---

## 商業模式：Open Core

**Free** (`go install` / `brew install mur`)
```
├── Multi-tool runner
├── Health check
├── MCP sync
├── Local learning (個人)
└── 5 patterns 上限
```

**Pro** ($15/mo per user)
```
├── 無限 patterns
├── Git-based team sync
├── Pattern analytics dashboard
├── Smart routing（成本最佳化）
└── Priority support
```

**Team** ($49/mo per 5 users)
```
├── 集中管理 patterns
├── Team merge + review workflow
├── Privacy controls
├── Onboarding automation
└── Admin dashboard
```

---

## 要做的事

1. **獨立 repo** — `github.com/karajanchang/murmur`
2. **Go CLI** — 用 Cobra 框架
3. **CLI 入口** — `mur run`、`mur config`、`mur learn`、`mur sync`
4. **Landing page** — 一頁式官網說明價值
5. **README + demo GIF** — GitHub 星星是 dev tools 的命脈

---

## CLI 命令設計

```bash
mur init           # 初始化
mur run -p "prompt"  # 跑任務（自動選工具）
mur run -t claude -p "prompt"  # 指定工具
mur config         # 互動設定
mur learn          # 管理 patterns
mur learn list     # 列出所有 patterns
mur learn add      # 手動新增 pattern
mur sync           # 同步 MCP + skills 到各工具
mur health         # 健康檢查
mur team           # 團隊管理 (Pro/Team)
```

---

## 技術選型：Go

| | Go | Rust | Node.js |
|---|---|---|---|
| FreeBSD | ✅ 原生支援 | ❌ 官方不支援 | ✅ |
| 編譯速度 | 快 | 慢 | N/A |
| Binary 大小 | ~10MB | ~5MB | ~50MB+ |
| Cross-compile | 一行指令 | 需 toolchain | 不適合 |
| CLI 框架 | Cobra（成熟） | Clap | Commander |

**選 Go 的原因：**
- FreeBSD 支援（你的需求）
- 跨平台編譯簡單：`GOOS=freebsd GOARCH=amd64 go build`
- 生態成熟：docker, kubectl, gh, terraform 都用 Go

**支援平台：**
```
├── darwin-arm64   (macOS Apple Silicon)
├── darwin-amd64   (macOS Intel)
├── linux-amd64
├── linux-arm64
├── freebsd-amd64  ✅
└── windows-amd64
```

---

## License 保護策略

### 推薦組合：Go CLI + Server validation

```
用戶 CLI (Go binary)                Murmur API
        │                                 │
        ├── mur run -p "prompt"           │
        │   → 本地跑 AI tool（不需 server）│
        │                                 │
        ├── mur learn sync                │
        │   → POST /api/patterns          │
        │   → server 驗證 license + 計數  │
        │                                 │
        └── mur health                    │
            → GET /api/license/status     │
            → 回傳：plan, pattern_count,  │
               pattern_limit, expiry      │
```

---

## 各層級限制方式

### Free

```yaml
limits:
  patterns: 5           # 本地 + server 雙重檢查
  tools: unlimited      # 不限（吸引用戶）
  mcp_sync: true        # 不限（基本功能）
  team_sync: false      # 禁用
  smart_routing: false  # 禁用

enforcement:
  method: local_count + server_reject
```

### Pro ($15/mo)

```yaml
limits:
  patterns: unlimited
  tools: unlimited
  mcp_sync: true
  team_sync: false      # 僅個人
  smart_routing: true   # 自動選最便宜的工具
  analytics: basic      # CLI 內看統計

enforcement:
  method: license_key + monthly_heartbeat
  # 本地存 ~/.murmur/license.json
  # 每月驗證一次（離線也能用 30 天）
```

**License file 格式：**
```json
{
  "key": "mur_pro_abc123...",
  "plan": "pro",
  "email": "david@example.com",
  "verified_at": "2026-02-01T00:00:00Z",
  "expires_at": "2026-03-01T00:00:00Z",
  "offline_grace_days": 30
}
```

### Team ($49/mo)

```yaml
limits:
  patterns: unlimited
  users: 5              # server 端管理
  team_sync: true
  merge_workflow: true
  privacy_controls: true
  admin_dashboard: true # Web UI

enforcement:
  method: org_license + server_managed
```

---

## 關鍵原則

```
能在本地跑的           需要 server 的
─────────             ─────────────
Free  mur run, mur sync  pattern sync (限 5)
Pro   + smart routing    license 驗證 (月)
Team  (同 Pro)           team sync, merge, admin UI
```

**核心思路：讓免費功能跑在本地（不設防），付費功能綁 server（無法繞過）。**

---

## Desktop App（待討論）

**問題：是否需要 GUI？**

| 選項 | 優點 | 缺點 |
|------|------|------|
| **純 CLI** | 開發快、開發者愛、維護簡單 | 非開發者難上手 |
| **CLI + TUI** | 保持終端內、互動體驗好 | 還是需要打開終端 |
| **CLI + Desktop App** | 視覺化、通知整合、系統匣 | 開發成本高、要維護多平台 |
| **CLI + Web Dashboard** | 一次開發、跨平台 | 需要開 server |

**如果要做 Desktop：**
- **macOS：** SwiftUI（你已經會）
- **跨平台：** Wails（Go + Web UI）或 Tauri（Rust + Web UI）
- **功能：** pattern 瀏覽、sync 狀態、license 管理、通知

**建議：** 先做好 CLI，Desktop 是 Phase 2。CLI 做好了，Desktop 只是包一層 UI。
