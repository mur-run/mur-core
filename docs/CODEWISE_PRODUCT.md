# Codewise — 產品規劃

**Created:** 2026-02-03  
**Status:** 規劃中  
**Tagline:** Multi-AI CLI 統一管理層 + 跨工具學習系統

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

**Free** (`npm install -g codewise`)
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

1. **改名 + 獨立 repo** — `codewise` 從 clawdbot skill 獨立出來
2. **npm package** — `npm install -g codewise`
3. **CLI 入口** — `codewise run`、`codewise config`、`codewise learn`、`codewise sync`
4. **Landing page** — 一頁式官網說明價值
5. **README + demo GIF** — GitHub 星星是 dev tools 的命脈

---

## CLI 命令設計

```bash
codewise init      # 初始化
codewise run -p "prompt"  # 跑任務
codewise config    # 互動設定
codewise learn     # 管理 patterns
codewise sync      # 同步 MCP + skills
codewise health    # 健康檢查
codewise team      # 團隊管理
```

---

## License 保護策略

### 防護方案比較

| 方案 | 做法 | 防護力 |
|------|------|--------|
| **Server-side validation** | 關鍵邏輯在雲端 API，CLI 只是 client | ⭐⭐⭐⭐⭐ |
| **License key + heartbeat** | 本地存 key，定期跟 server 驗證 | ⭐⭐⭐⭐ |
| **Rust/Go binary** | 編譯語言取代 Node.js，反編譯難度高 | ⭐⭐⭐ |
| **JS 混淆 (obfuscate)** | `javascript-obfuscator` | ⭐⭐ |

### 推薦組合：Rust CLI + Server validation

```
用戶 CLI (Rust binary)              Codewise API
        │                                 │
        ├── codewise run -p "prompt"      │
        │   → 本地跑 AI tool（不需 server）│
        │                                 │
        ├── codewise learn (sync patterns)│
        │   → POST /api/patterns          │
        │   → server 驗證 license + 計數  │
        │                                 │
        └── codewise health               │
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
  tools: unlimited      # ai_run.sh 不限（吸引用戶）
  mcp_sync: true        # 不限（基本功能）
  team_sync: false      # 禁用
  smart_routing: false  # 禁用

enforcement:
  method: local_count + server_reject
  # 純本地檢查（可被 bypass，但無所謂）
  # sync 時 server 會拒絕超過 5 個 patterns
```

**怎麼限：**
- `sync_skills.sh` 同步前 → count patterns → 超過 5 個提示升級
- Server 端也檢查 → 即使改了本地 code，server 拒絕同步
- **不限 AI tool 使用** — 讓免費用戶養成習慣

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
  # 本地存 ~/.codewise/license.json
  # 每月驗證一次（離線也能用 30 天）
```

**License file 格式：**
```json
{
  "key": "cw_pro_abc123...",
  "plan": "pro",
  "email": "david@example.com",
  "verified_at": "2026-02-01T00:00:00Z",
  "expires_at": "2026-03-01T00:00:00Z",
  "offline_grace_days": 30
}
```

**怎麼限：**
- 啟動時檢查 license file
- 過期 → 降級回 Free（不刪資料，只限功能）
- 離線寬限 30 天（對開發者友善）

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
  # Team features 全部走 server
  # merge_team.sh → POST /api/team/merge（server 驗證 seat count）
```

**怎麼限：**
- Team 功能全部需要 API call → server 端控制 seat 數
- `merge_team.sh` → 呼叫 API → server 檢查這個 org 有幾個 seat
- Admin dashboard 是 web app → 天生 server-side

---

## 關鍵原則

```
能在本地跑的           需要 server 的
─────────             ─────────────
Free  AI run, MCP sync  pattern sync (限 5)
Pro   + smart routing   license 驗證 (月)
Team  (同 Pro)          team sync, merge, admin UI
```

**核心思路：讓免費功能跑在本地（不設防），付費功能綁 server（無法繞過）。**
