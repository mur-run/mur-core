---
name: sparkle-xpc-bootstrap-hang
confidence: HIGH
score: 0.85
category: debug
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 2
---

# Sparkle Updater XPC Bootstrap Hang on Unsigned Apps

## Problem / Trigger
macOS app hangs on startup when using Sparkle framework for updates, especially on unsigned/development builds

## Solution
Initialize SPUStandardUpdaterController with `startingUpdater: false` parameter to prevent XPC bootstrap hang on unsigned apps

## Verification
App launches without hanging; updater can be started manually later when needed

## Source
Session: 2026-02-05
