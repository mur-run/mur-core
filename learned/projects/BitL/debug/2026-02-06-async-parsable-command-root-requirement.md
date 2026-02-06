---
name: async-parsable-command-root-requirement
confidence: HIGH
score: 0.85
category: debug
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Root Command Must Be AsyncParsableCommand for Async Subcommands

## Problem / Trigger
Swift ArgumentParser CLI where the root command is ParsableCommand (sync) but subcommands are AsyncParsableCommand — subcommands silently fail and just print help text instead of executing

## Solution
Change the root command from ParsableCommand to AsyncParsableCommand. The root command's async/sync type determines the execution context for all subcommands.

## Verification
Run any async subcommand and confirm it executes instead of printing help

## Source
Session: 2026-02-01
