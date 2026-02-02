---
name: swift-spm-multi-target-architecture
confidence: HIGH
score: 0.85
category: pattern
domain: _global
first_seen: 2026-02-01
last_seen: 2026-02-02
times_seen: 10
---

# Swift Package Manager 多 target 分層架構

## Problem / Trigger
Swift 專案越來越大，全部放同一個 target 編譯慢、職責不清。

## Solution
按職責分層成多個 SPM targets：
```
Sources/
├── BitLCore/      → Models, protocols, paths（零依賴，最底層）
├── BitLInfra/     → Managers, parsers, config generators（依賴 Core）
├── BitLServices/  → Service orchestration（依賴 Core + Infra）
├── BitLCLI/       → CLI commands via ArgumentParser（依賴全部）
├── BitLApp/       → SwiftUI views（依賴全部）
└── BitLHelper/    → Privileged helper（最小依賴）
```

Tests 也對應分層：`BitLCoreTests`, `BitLInfraTests`, `BitLServicesTests` etc.

好處：
- 改 Core 的 model 只需重編 Core tests
- 強制依賴方向（Core 不能 import Services）
- 新人一看 target 名就知道東西放哪裡
- Test target 也可以單獨跑

## Verification
BitL 6 個 source targets + 6 個 test targets，1144 tests，build < 3 秒（incremental）。
