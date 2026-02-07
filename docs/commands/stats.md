# mur stats

View usage statistics and cost analysis.

## Usage

```bash
mur stats [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--days N` | Show stats for last N days (default: 30) |
| `--tool <name>` | Filter by specific tool |
| `--json` | Output as JSON |

## Output

```bash
mur stats
```

```
Usage Statistics (Last 30 Days)
===============================

Tool Breakdown:
  claude    │████████████████████░░░░░░░░░░│  67%  (134 runs)
  gemini    │████████████░░░░░░░░░░░░░░░░░░│  33%  (66 runs)

Routing Effectiveness:
  Auto-routed: 85% (170 runs)
  Manual override: 15% (30 runs)

Cost Analysis:
  Estimated spend: $12.45
  Saved by routing: $8.20 (66 prompts → free tool)

Complexity Distribution:
  Simple (0.0-0.3):   ████████████░░░░░░░░  45%
  Medium (0.3-0.6):   ████████░░░░░░░░░░░░  35%
  Complex (0.6-1.0):  ████░░░░░░░░░░░░░░░░  20%

Top Categories:
  1. coding      (89 runs)
  2. simple-qa   (52 runs)
  3. debugging   (31 runs)
  4. analysis    (18 runs)
  5. architecture (10 runs)
```

## Filtering

### By Time Period

```bash
mur stats --days 7   # Last week
mur stats --days 1   # Today
mur stats --days 90  # Last quarter
```

### By Tool

```bash
mur stats --tool claude
mur stats --tool gemini
```

## JSON Output

For programmatic access:

```bash
mur stats --json
```

```json
{
  "period_days": 30,
  "total_runs": 200,
  "by_tool": {
    "claude": {"runs": 134, "percentage": 67},
    "gemini": {"runs": 66, "percentage": 33}
  },
  "auto_routed": 170,
  "manual_override": 30,
  "estimated_cost": 12.45,
  "cost_saved": 8.20,
  "by_category": {
    "coding": 89,
    "simple-qa": 52,
    "debugging": 31,
    "analysis": 18,
    "architecture": 10
  }
}
```

## Cost Estimation

Murmur estimates costs based on prompt length and tool tier:

| Tool | Cost Model |
|------|------------|
| Claude | ~$0.01 per 1K tokens (estimated) |
| Gemini | Free |
| Auggie | Free |

!!! note
    These are rough estimates. Actual costs depend on your API plan and response lengths.

## Data Storage

Statistics are stored in `~/.mur/stats.json`. Each run records:

- Tool used
- Timestamp
- Prompt length
- Duration
- Complexity score
- Category
- Whether auto-routed
- Success/failure

## Privacy

All statistics are stored locally. Nothing is sent to external servers.

## See Also

- [run Command](run.md) - Running prompts
- [Smart Routing](../concepts/routing.md) - How routing decisions are made
