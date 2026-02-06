---
name: swift-async-parsable-command-root
confidence: HIGH
score: 0.85
category: debug
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Root ParsableCommand Silently Breaks Async Subcommands

## Problem / Trigger
When root CLI command is `ParsableCommand` (sync), all async subcommands silently fail and just print help text instead of executing

## Solution
Change the root command from `ParsableCommand` to `AsyncParsableCommand`. Even if the root itself has no async work, it must be async for async subcommands to dispatch correctly.

## Verification
Run any async subcommand — it should execute instead of printing help

## Source
Session: 2026-02-01
