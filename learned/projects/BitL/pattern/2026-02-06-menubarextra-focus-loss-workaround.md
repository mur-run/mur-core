---
name: menubarextra-focus-loss-workaround
confidence: HIGH
score: 0.78
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# MenuBarExtra Panel Loses Focus for File Pickers and TextFields

## Problem / Trigger
SwiftUI MenuBarExtra panel loses keyboard focus when opening file pickers (.fileImporter) or typing in TextFields, making settings UI unusable

## Solution
Open settings as a standalone NSWindow via a custom SettingsWindowController instead of rendering inside the MenuBarExtra panel. Requires NSApp.setActivationPolicy(.regular) for proper focus handling in menu bar-only apps

## Verification
Open settings window, confirm TextFields accept keyboard input and file picker dialogs appear correctly

## Source
Session: 2026-02-01
