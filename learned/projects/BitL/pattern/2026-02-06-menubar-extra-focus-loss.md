---
name: menubar-extra-focus-loss
confidence: HIGH
score: 0.8
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# SwiftUI MenuBarExtra Loses Focus on File Pickers and TextFields

## Problem / Trigger
Settings panel in MenuBarExtra dismisses or loses keyboard focus when using .fileImporter or typing in TextField

## Solution
Open settings as standalone NSWindow via NSWindowController instead of inline MenuBarExtra content. Set NSApp.setActivationPolicy(.regular) for proper focus handling in menu bar apps.

## Verification
File picker stays open, TextField accepts keyboard input without dismissing

## Source
Session: 2026-02-01
