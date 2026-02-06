---
name: task-detached-for-blocking-system-calls
confidence: MEDIUM
score: 0.65
category: performance
domain: mobile
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Use Task.detached for Blocking System Calls in SwiftUI

## Problem / Trigger
ResourceMonitorView calling `getProcessStats()` with regular async/await blocks the UI, causing the app to appear frozen while system stats are being collected

## Solution
Use `Task.detached(priority: .background)` instead of regular `Task` or `Task { }` for operations that involve blocking system calls like process stat collection. Regular tasks inherit the actor context and can block the main thread.

## Verification
Open ResourceMonitorView and confirm the UI remains responsive while process statistics are being loaded/refreshed.

## Source
Session: 2026-02-05
