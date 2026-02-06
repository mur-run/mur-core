---
name: macos-menubar-nswindow-focus
confidence: HIGH
score: 0.78
category: debug
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# MenuBarExtra Panels Lose Focus for File Pickers and TextFields

## Problem / Trigger
SwiftUI MenuBarExtra panels lose keyboard focus when opening file pickers (.fileImporter) or typing in TextFields, making settings UI unusable

## Solution
Open settings as a standalone NSWindow via a custom SettingsWindowController instead of within the MenuBarExtra panel. Must call `NSApp.setActivationPolicy(.regular)` for the window to receive proper focus in a menu bar-only app.

## Verification
Open settings, click a TextField and type — focus should stay. Open a file picker — it should appear and be interactive without the window losing focus.

## Source
Session: 2026-02-01
