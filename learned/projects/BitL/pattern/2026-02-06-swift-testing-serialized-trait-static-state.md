---
name: swift-testing-serialized-trait-static-state
confidence: HIGH
score: 0.82
category: pattern
domain: mobile
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Swift Testing .serialized Trait for Static Mutable State

## Problem / Trigger
Swift Testing tests that modify a static variable (e.g., L10n.current for locale) cause race conditions when run in parallel, leading to flaky test failures

## Solution
Add the `.serialized` trait to test suites that mutate shared static state like `L10n.current`, forcing sequential execution within that suite while keeping other suites parallel

## Verification
Run full test suite multiple times - all 1120 tests should pass consistently without intermittent locale-related failures

## Source
Session: 2026-02-06
