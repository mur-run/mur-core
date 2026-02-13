# MUR Core Demo Script

影片長度：2-3 分鐘
目標：展示 mur-core 是真實可用的產品

---

## 場景 1: 安裝 (30 秒)

**畫面：** Terminal，乾淨背景

**指令：**
```bash
# 使用 Homebrew 安裝
brew tap mur-run/tap && brew install mur

# 確認安裝成功
mur version
```

**預期輸出：**
```
mur 1.6.0
  commit:  ...
  built:   ...
```

**旁白：**
> "mur-core 可以透過 Homebrew 一行指令安裝，也支援 go install。"

---

## 場景 2: 初始化 (30 秒)

**指令：**
```bash
# 初始化並安裝 hooks
mur init --hooks
```

**預期輸出：**
```
✓ mur initialized at ~/.mur
✓ Installed Claude Code hooks
✓ Installed Gemini CLI hooks
✓ Installed Cursor hooks
✓ Installed Windsurf hooks
... (more tools)
```

**旁白：**
> "一行指令就能整合 10+ AI 工具，包括 Claude Code、Cursor、Windsurf 等。"

---

## 場景 3: 健康檢查 (20 秒)

**指令：**
```bash
mur doctor
```

**預期輸出：** (全綠)
```
✅ ~/.mur directory
✅ ~/.mur/patterns      132 patterns
✅ Claude Code          Found
✅ Gemini CLI           Found
✅ Claude hooks
...
```

**旁白：**
> "mur doctor 可以快速檢查系統狀態，確保所有工具都正常連接。"

---

## 場景 4: 搜尋 Patterns (30 秒)

**指令：**
```bash
# 語意搜尋
mur search "Swift testing best practices"
```

**預期輸出：**
```
🔍 Found 3 patterns:

1. swift-testing-macro (0.92)
   Use Swift Testing macros over XCTest...

2. swift-async-testing (0.85)
   Async testing patterns...
```

**旁白：**
> "mur 使用語意搜尋，不只是關鍵字比對。輸入自然語言就能找到相關的 patterns。"

---

## 場景 5: 同步到所有工具 (30 秒)

**指令：**
```bash
mur sync
```

**預期輸出：**
```
Syncing patterns to CLIs...
  ✓ Claude Code: Synced 132 patterns
  ✓ Cursor: Synced 132 patterns
  ✓ Windsurf: Synced 132 patterns
  ✓ Gemini CLI: Synced 132 patterns
  ✓ GitHub Copilot: Synced 132 patterns
  ... (more)

✅ Sync complete
```

**旁白：**
> "一鍵同步到所有 AI 工具。你的 patterns 會自動注入到 Claude、Cursor、Copilot 等。"

---

## 場景 6: 本地 Dashboard (30 秒)

**指令：**
```bash
mur serve
# 打開 http://localhost:3377
```

**畫面：** 切換到瀏覽器，展示 Dashboard

- Pattern 列表
- 搜尋功能
- Pattern 詳情

**旁白：**
> "內建的 Web Dashboard 讓你管理 patterns、查看統計數據。"

---

## 場景 7: 雲端功能預覽 (20 秒)

**畫面：** 打開 app.mur.run

- 展示 Login 頁面
- 展示 Pricing 頁面 (Free 可用, Pro/Team 即將推出)

**旁白：**
> "雲端功能包括跨裝置同步、團隊共享。免費版已完整可用，付費方案即將推出。"

---

## 結尾 (10 秒)

**畫面：** 回到 Terminal

**指令：**
```bash
mur status
```

**旁白：**
> "mur-core，讓你的 AI 工具持續學習、越用越聰明。"

---

## 錄製注意事項

1. **Terminal 設定：**
   - 字體大小：18-20pt
   - 深色背景，高對比度
   - 隱藏多餘的 prompt 資訊

2. **錄製工具：**
   - macOS: QuickTime 或 ScreenFlow
   - 解析度：1920x1080 或 1280x720

3. **節奏：**
   - 每個指令後停頓 2 秒
   - 讓輸出完整顯示

4. **清理：**
   - 錄製前執行 `clear`
   - 關閉不需要的通知
