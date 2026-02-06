---
name: xpc-helper-osascript-alternative
confidence: MEDIUM
score: 0.65
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# osascript 'with administrator privileges' Replaces XPC Helper for Privileged Ops

## Problem / Trigger
SMAppService XPC helper requires code signing and proper entitlements — doesn't work with swift build ad-hoc signed apps

## Solution
Use osascript -e 'do shell script "..." with administrator privileges' for privileged operations like binding port 53. User gets standard macOS password prompt.

## Verification
Privileged command executes after user authenticates, no codesigning errors

## Source
Session: 2026-02-01
