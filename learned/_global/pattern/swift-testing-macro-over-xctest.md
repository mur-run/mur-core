---
name: swift-testing-macro-over-xctest
confidence: HIGH
score: 0.9
category: pattern
domain: _global
first_seen: 2026-02-01
last_seen: 2026-02-02
times_seen: 5
---

# Swift Testing framework (@Test/@Suite) 取代 XCTest

## Problem / Trigger
Swift 6+ 專案需要寫測試，直覺用 XCTestCase 但跟 Sendable/async 衝突多。

## Solution
用 Swift Testing framework（`import Testing`）取代 XCTest：
- `@Suite("Name")` 取代 `class XXXTests: XCTestCase`
- `@Test("description")` 取代 `func testXxx()`
- `#expect(condition)` 取代 `XCTAssert*()`
- `#expect(throws: ErrorType.self) { code }` 取代 `XCTAssertThrowsError`
- struct 而非 class — 天生 Sendable，不需 `@MainActor`

好處：更簡潔、更好的錯誤訊息、原生支援 async、跟 SwiftUI previews 更合。

## Verification
BitL 專案 1144 個測試全部用 Swift Testing，零 XCTest 依賴。
