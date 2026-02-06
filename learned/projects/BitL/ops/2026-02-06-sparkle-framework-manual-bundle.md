---
name: sparkle-framework-manual-bundle
confidence: HIGH
score: 0.75
category: ops
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# SPM Sparkle Requires Manual Framework Copy and rpath

## Problem / Trigger
App built with swift build crashes on launch when using Sparkle auto-update framework — framework not found at runtime

## Solution
After swift build: (1) Copy .build/arm64-apple-macosx/release/Sparkle.framework to Contents/Frameworks/, (2) Add rpath with install_name_tool -add_rpath @executable_path/../Frameworks, (3) codesign --force --deep --sign -

## Verification
otool -L shows framework resolved, app launches without dylib errors

## Source
Session: 2026-02-01
