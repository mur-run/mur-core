# mur consolidate

Pattern health analysis, duplicate detection, and automated cleanup.

## Overview

As your pattern library grows, quality degrades: duplicates accumulate, outdated patterns linger, and contradictory advice creeps in. `mur consolidate` fixes this by analyzing every pattern across four health dimensions and recommending actions.

Think of it like your brain consolidating memories during sleep — keeping what matters, merging what overlaps, and forgetting what's stale.

## Usage

```bash
# Preview what would happen (default, safe)
mur consolidate

# Auto-execute safe actions (archive stale, keep-best merges)
mur consolidate --auto

# Review each action one by one
mur consolidate --interactive

# Skip minimum pattern count check
mur consolidate --force
```

## Health Score

Every pattern gets a score from 0.0 to 1.0 based on four dimensions:

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| **Freshness** | 25% | Time since last use. Half-life of 90 days — unused for 90 days = 0.5, 180 days = 0.25. New patterns (< 14 days) get a grace period. |
| **Engagement** | 30% | How often the pattern is used. Log-scaled so high-frequency patterns don't dominate. 128 uses = max score. |
| **Quality** | 30% | User feedback ratio. `helpful / (helpful + unhelpful)`. No feedback = 0.5 (neutral). |
| **Uniqueness** | 15% | How different from other patterns. `1 - max_cosine_similarity`. Near-duplicate = low score. |

## Decision Rules

Applied in priority order:

1. **Merge** — Uniqueness < 0.15 (cosine similarity > 0.85 with another pattern)
2. **Update** — Quality < 0.2 but Engagement > 0.3 (frequently used, poorly rated)
3. **Archive** — Freshness < 0.1 and Engagement < 0.1 (stale and unused)
4. **Archive** — Overall score < 0.25
5. **Keep** — Everything else

## Example Output

```
🧠 MUR Pattern Consolidation Report
════════════════════════════════════

📊 Scanned: 147 patterns | 1,234 usage events | 89 feedback records

🟢 Healthy (112 patterns)
   Top performers:
   ├── go-error-handling    │ Health: 0.92 │ Used: 47x │ 👍 4.5/5
   ├── api-retry-pattern    │ Health: 0.88 │ Used: 31x │ 👍 4.2/5
   └── swift-async-await    │ Health: 0.85 │ Used: 28x │ 👍 4.0/5

🔀 Merge Candidates (3 groups, 8 patterns)
   Group 1: 89% similarity
   ├── "react-useeffect-cleanup" (Health: 0.71)
   └── "react-effect-teardown"   (Health: 0.54)
   → Keep: react-useeffect-cleanup | Archive: react-effect-teardown

📦 Archive Candidates (12 patterns)
   ├── "laravel-mix-config"     │ Last used: 203 days ago │ Health: 0.08
   └── "webpack-4-treeshaking"  │ Last used: 178 days ago │ Health: 0.12

Summary: merge 8 → 3 | archive 12 | update 3 | keep 112
```

## Configuration

```yaml
consolidation:
  enabled: true
  schedule: "weekly"           # daily | weekly | monthly
  auto_archive: true           # auto-archive patterns below threshold
  auto_merge: "keep-best"      # off | keep-best | llm-merge
  merge_threshold: 0.85        # cosine similarity threshold for duplicates
  decay_half_life_days: 90     # freshness half-life in days
  grace_period_days: 14        # new patterns get this long before decay applies
  min_patterns_before_run: 50  # don't run consolidation below this count
```

## Tips

- Run `mur consolidate` (dry-run) weekly to keep your library healthy
- Use `mur feedback helpful <pattern>` and `mur feedback unhelpful <pattern>` to improve Quality scores
- Patterns with `team_shared: true` are never auto-archived
- The `--force` flag skips the minimum pattern count check — useful for testing
