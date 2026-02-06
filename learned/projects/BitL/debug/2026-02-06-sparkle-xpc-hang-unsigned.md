---
name: sparkle-xpc-hang-unsigned
confidence: HIGH
score: 0.78
category: debug
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Sparkle Updater XPC Bootstrap Hang on Unsigned Apps

## Problem / Trigger
App hangs at launch when Sparkle auto-updater tries to start its XPC service on unsigned/development builds. The XPC bootstrap fails silently causing the main thread to block indefinitely.

## Solution
Initialize SPUStandardUpdaterController with startingUpdater: false parameter. This prevents the XPC service from attempting to bootstrap at init time, avoiding the hang on unsigned builds.

## Verification
Launch unsigned dev build - app should start without hanging. Sparkle updates can still be triggered manually when needed.

## Source
Session: 2026-02-05
