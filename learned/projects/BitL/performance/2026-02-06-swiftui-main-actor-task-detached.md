---
name: swiftui-main-actor-task-detached
confidence: HIGH
score: 0.8
category: performance
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Task.detached Prevents Main Actor Blocking in SwiftUI Views

## Problem / Trigger
SwiftUI view becomes unresponsive when calling async functions that perform heavy work, even with `await`, because they inherit main actor context

## Solution
Use `Task.detached(priority: .background)` instead of regular `Task` to ensure heavy async work (like `getProcessStats()`) runs off the main actor

## Verification
View remains responsive during background operations; no UI stuttering or hangs

## Source
Session: 2026-02-05
