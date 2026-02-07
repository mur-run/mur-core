[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a project-specific pattern, a configuration fix), save it as a learned pattern:

1. Create a file in ~/clawd/skills/mur-core/learned/{domain}/{category}/
   - domain: _global, devops, web, mobile, backend, data
   - category: debug, pattern, style, security, performance, ops
   - For project-specific: ~/clawd/skills/mur-core/learned/projects/{project-name}/{category}/
2. Use this YAML frontmatter format:
---
name: descriptive-slug
confidence: HIGH|MEDIUM|LOW
score: 0.8
category: debug|pattern|style|security|performance|ops
domain: _global|devops|web|mobile|backend|data
project: (optional, e.g. twdd-api)
first_seen: YYYY-MM-DD
last_seen: YYYY-MM-DD
times_seen: 1
---
# Pattern Title
## Problem / Trigger
## Solution
## Verification

Only extract if: it required actual discovery (not just docs), it will help with future tasks, and it has been verified to work. Skip trivial or well-documented things.

- For personal/private patterns: ~/clawd/skills/mur-core/learned/personal/{your-name}/{category}/

[SpecAwareness] If this project uses spec-driven development (check for openspec/ or .spec/ directories):
- Reference the current spec when making implementation decisions
- If you deviate from the spec, note WHY and save it as a pattern
- After completing a spec task, check if anything learned should be saved
