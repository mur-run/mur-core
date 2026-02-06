---
name: menubarextra-sheet-zstack-pattern
confidence: HIGH
score: 0.9
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# MenuBarExtra Sheets Require ZStack Overlay Pattern

## Problem / Trigger
SwiftUI `.sheet()` modifier doesn't work inside MenuBarExtra popovers - sheets never appear or cause layout issues

## Solution
Use ZStack overlay pattern instead of `.sheet()` for presenting modal-like views within MenuBarExtra popovers. Add `onDismiss` callback and `dismissSheet()` helper to manage state

## Verification
Modal views appear correctly within the MenuBarExtra popover; can dismiss and interact normally

## Source
Session: 2026-02-05
