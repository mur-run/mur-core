---
name: menubarextra-sheet-zstack-overlay
confidence: HIGH
score: 0.9
category: pattern
domain: mobile
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 2
---

# MenuBarExtra Popovers Cannot Use .sheet() Modifier

## Problem / Trigger
SwiftUI `.sheet()` modifier does not work inside MenuBarExtra popovers - sheets simply don't appear or cause layout issues

## Solution
Use a ZStack overlay pattern instead of `.sheet()`. Add an `onDismiss` callback and a `dismissSheet()` helper to manage the overlay state manually. Wrap content in ZStack and conditionally show the sheet view as an overlay.

## Verification
Present a form/dialog from within a MenuBarExtra popover and confirm it appears correctly as an overlay rather than failing silently.

## Source
Session: 2026-02-05
