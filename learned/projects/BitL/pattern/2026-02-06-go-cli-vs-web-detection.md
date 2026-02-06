---
name: go-cli-vs-web-detection
confidence: MEDIUM
score: 0.6
category: pattern
domain: backend
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Go Project Type Detection via go.mod Framework Analysis

## Problem / Trigger
Need to distinguish Go CLI projects from Go web projects for appropriate tooling/domain configuration, but both use the same go.mod file

## Solution
Parse go.mod and check for web framework imports (gin, echo, fiber, chi). If any are found, classify as web project; otherwise classify as CLI. Use `ProjectType.needsDomain` property to control whether domain-related features are shown.

## Verification
Create Go projects with and without web framework dependencies and confirm correct classification. Verify domain features are hidden for CLI projects.

## Source
Session: 2026-02-05
