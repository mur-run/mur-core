---
name: swift-enum-visibility-across-files
confidence: HIGH
score: 0.75
category: debug
domain: mobile
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Swift Nested Enum Not Visible When Parent Type Decomposed to Separate Files

## Problem / Trigger
When decomposing a large Swift file (e.g., AppState.swift) into smaller files, nested enums defined inside the original type are not visible to the new extracted files, causing build failures

## Solution
Extract nested enums (like DebugTab) to their own standalone files before or alongside extracting the state that depends on them. The enum must be a top-level type or in a file that's compiled before its dependents

## Verification
Build succeeds after moving the enum to its own file and all references in decomposed files resolve correctly

## Source
Session: 2026-02-06
