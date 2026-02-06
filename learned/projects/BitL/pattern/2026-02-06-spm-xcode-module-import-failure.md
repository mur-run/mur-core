---
name: spm-xcode-module-import-failure
confidence: MEDIUM
score: 0.7
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# SPM Module Imports Don't Work in xcodegen-Merged Targets

## Problem / Trigger
Trying to use xcodegen to create Xcode project for code signing, but import BitLCore fails when all sources are merged into single target

## Solution
Cannot use xcodegen with merged sources for multi-module SPM projects. Would need proper Xcode project with separate targets matching SPM module structure, or stick with manual ad-hoc signing.

## Verification
N/A — this is a limitation requiring architectural change

## Source
Session: 2026-02-01
