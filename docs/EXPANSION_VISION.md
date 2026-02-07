# mur-core 擴展願景：從 CLI Hooks 到通用學習系統

**Created:** 2026-02-03  
**Status:** 願景規劃

---

## 核心概念

mur-core 的核心其實不是「開發工具」，是**「持續學習系統」** — 從工作過程中提取 patterns，存起來，同步到各種 AI 工具。

**現在的架構：**
```
Hook 觸發 → AI 工作時提醒 → 學到東西存 pattern → 同步到 AI tools
```

這個循環不限於寫 code。

---

## Domain 結構擴充

**現在的 domain 結構已經有預留：**
```
learned/
  _global/        ← 通用
  devops/         ← DevOps（空的）
  web/            ← Web（空的）
  mobile/         ← Mobile（空的）
  backend/        ← Backend（有 2 個 patterns）
  data/           ← Data（空的）
  projects/       ← 專案特定
  personal/       ← 個人
```

**可以擴充成：**
```
learned/
  _global/
  dev/            ← 軟體開發
  devops/         ← CI/CD、部署、監控
  infra/          ← 基礎設施、雲端
  security/       ← 安全
  data/           ← 資料工程、分析
  product/        ← 產品設計、PM
  business/       ← 商業分析、策略
  ops/            ← 日常營運
  projects/
  personal/
```

---

## 具體場景範例

### DevOps patterns

```yaml
name: docker-layer-cache-optimization
domain: devops
category: performance
---
# Docker Build 快取層優化

## Problem / Trigger
Docker build 每次都要重跑 npm install，CI 跑很慢

## Solution
把 package.json COPY 在 source code 之前，利用 layer cache
```

### Infra patterns

```yaml
name: aws-rds-failover-gotcha
domain: infra
category: debug
---
# AWS RDS Failover 後 connection pool 不會自動重連

## Problem / Trigger
RDS failover 後 app 持續報錯，但 RDS 已經 healthy

## Solution
connection pool 要設 maxLifetime，強制定期重建連線
```

### Business patterns

```yaml
name: saas-pricing-anchor-effect
domain: business
category: pattern
---
# SaaS 三級定價的錨定效應

## Problem / Trigger
只有一個方案，客戶覺得貴不買單

## Solution
三級定價，中間方案是真正要賣的，最貴的是錨點
```

---

## 需要的改動

架構**幾乎不用改**，只需要：

| 改動 | 說明 |
|------|------|
| 擴充 domain 目錄 | `mkdir -p learned/{devops,infra,security,product,business,ops}` |
| 更新 on-prompt-reminder.md | 提醒不只是 code patterns，也包含 DevOps/business 等 |
| 更新 category 列表 | 加入 `deploy, monitoring, incident, strategy` 等 |
| Hook 掛到更多工具 | 不只 Claude Code，也掛到 Gemini/Auggie（Phase 1） |

---

## 更大的願景：mur-core 不只是 CLI hooks

**現在：**
```
Claude Code session → 學到 pattern → 存 learned/
```

**擴大成：**
```
任何 AI 互動 → 學到 pattern → 存 learned/
```

### 來源可以是

```
├── Claude Code sessions（現有）
├── Clawdbot/OpenClaw 對話
├── Slack/Teams 裡的討論
├── Incident postmortem
├── Meeting notes
└── 手動輸入
```

### 輸出同步到

```
├── Claude Code skills
├── Gemini CLI skills
├── Auggie commands
├── Clawdbot/OpenClaw memory
├── Notion/Obsidian knowledge base
└── Team wiki
```

---

## 最有價值的方向

如果要選一個擴展方向，建議是 **Clawdbot 對話 → patterns**：

```
你跟我今天聊了 TWDD 平台策略，裡面有很多洞察：

- 牽送費佔比決定服務可行性
- SaaS 賣給 B 端不要碰定價權
- 台灣市場 LINE > App
- 季節性數據要等淡季才能看基本盤

這些都是 patterns，但現在只存在 memory/ 裡，
如果也能自動提取成 learned/ patterns，
下次討論類似問題時，AI 就會記得這些經驗。
```

---

## 下一步選項

1. **先把 domain 目錄擴充好** — 基礎設施就位
2. **實作 Clawdbot 對話 → patterns 提取** — 最有價值的擴展
3. **先完成 Phase 1 hooks_sync.sh** — 讓跨 CLI 先能用
