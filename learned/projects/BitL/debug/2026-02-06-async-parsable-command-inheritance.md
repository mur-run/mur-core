---
name: async-parsable-command-inheritance
confidence: HIGH
score: 0.85
category: debug
domain: backend
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Swift ArgumentParser Root Command Must Be Async for Async Subcommands

## Problem / Trigger
All async subcommands just print help text instead of executing when the root CLI command inherits from sync ParsableCommand

## Solution
Change root command from ParsableCommand to AsyncParsableCommand. The root command's protocol determines whether async subcommand run() methods are called.

## Verification
Run any async subcommand and verify it executes instead of printing help

## Source
Session: 2026-02-01
