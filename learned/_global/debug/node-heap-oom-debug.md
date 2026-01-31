---
name: node-heap-oom-debug
confidence: HIGH
score: 0.9
category: debug
domain: _global
first_seen: 2025-01-15
last_seen: 2025-01-28
times_seen: 4
---

# Node.js Heap OOM 除錯技巧

## Problem / Trigger
Node.js process 突然被 OOM kill，但 heap usage 看起來不高。

## Solution
用 `--max-old-space-size` 設定不夠，真正原因常是 ArrayBuffer 或 external memory 不計入 V8 heap。
1. 用 `process.memoryUsage()` 看 `external` 和 `arrayBuffers`
2. 用 `--trace-gc` 觀察 GC 行為
3. 用 `node --heap-prof` 產生 heap profile，用 Chrome DevTools 分析

## Verification
在問題重現環境下 `process.memoryUsage().external` 確認外部記憶體持續增長。

## Source
Production debugging session, 2025-01-15
