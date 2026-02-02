---
name: l10n-dictionary-based-no-bundle
confidence: HIGH
score: 0.8
category: pattern
domain: _global
first_seen: 2026-02-02
last_seen: 2026-02-02
times_seen: 3
---

# Dictionary-based i18n（不用 .strings/.xcstrings）

## Problem / Trigger
Swift CLI + GUI 專案需要 i18n，但 SPM CLI target 沒有 Bundle.module 的 .strings 支援。

## Solution
用純 Swift dictionary 做 i18n，不依賴 Bundle/NSLocalizedString：
```swift
public enum Strings {
    public static let en: [String: String] = [
        "services.running": "Running",
        "php.current": "Current PHP version: %@",
    ]
    public static let zhHant: [String: String] = [
        "services.running": "執行中",
        "php.current": "目前 PHP 版本：%@",
    ]
}

public struct L10n {
    public static var current: Locale = .en
    public static func tr(_ key: String, _ args: CVarArg...) -> String {
        let template = strings(for: current)[key] ?? strings(for: .en)[key] ?? key
        return String(format: template, arguments: args)
    }
}
```

好處：CLI 和 GUI 共用、compile-time 驗證 key 存在、新增語言只要加一個 `Strings+xx.swift`。

## Verification
BitL 支援 en/zh-Hant/zh-Hans，95 個 key，全部 compile-time safe。
