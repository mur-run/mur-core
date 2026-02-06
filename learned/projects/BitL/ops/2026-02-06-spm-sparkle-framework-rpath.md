---
name: spm-sparkle-framework-rpath
confidence: HIGH
score: 0.75
category: ops
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 2
---

# SPM-Built Apps Need Manual Sparkle.framework Copying and rpath

## Problem / Trigger
SPM builds that depend on Sparkle produce the framework in .build/ but the app bundle won't find it at runtime — dyld fails to load Sparkle.framework

## Solution
Copy Sparkle.framework from .build/arm64-apple-macosx/release/ to Contents/Frameworks/ in the app bundle, then add rpath with install_name_tool -add_rpath @executable_path/../Frameworks

## Verification
Launch the app bundle and confirm no dyld errors, Sparkle update check works

## Source
Session: 2026-02-01
