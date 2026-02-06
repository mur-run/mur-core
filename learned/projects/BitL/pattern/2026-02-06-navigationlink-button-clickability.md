---
name: navigationlink-button-clickability
confidence: MEDIUM
score: 0.7
category: pattern
domain: mobile
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 2
---

# NavigationLink Background Pattern for Button Clickability

## Problem / Trigger
Buttons inside NavigationLink rows are not clickable - the NavigationLink captures all taps

## Solution
Use NavigationLink as background pattern (NavigationLink in .background modifier) rather than wrapping content, and increase spacing between interactive elements (e.g., 4→8 points)

## Verification
Individual buttons in list rows respond to taps independently from navigation action

## Source
Session: 2026-02-05
