---
name: spm-xcodegen-module-import-failure
confidence: MEDIUM
score: 0.7
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 2
---

# Xcodegen Cannot Preserve SPM Multi-Target Module Imports

## Problem / Trigger
Trying to use xcodegen to create an Xcode project from SPM package for code signing — module imports (import BitLCore) fail when sources are merged into a single target

## Solution
Don't use xcodegen for SPM projects with multiple internal targets. Either maintain a proper Xcode project with separate targets manually, or use ad-hoc codesign for development

## Verification
Build with swift build succeeds; xcodegen-generated project fails on module imports

## Source
Session: 2026-02-01
