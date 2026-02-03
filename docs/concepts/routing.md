# Smart Routing

Murmur's smart routing analyzes your prompts and automatically selects the best AI tool.

## The Problem

You have multiple AI CLI tools:

- **Claude Code** - Powerful but costs money
- **Gemini CLI** - Free but less capable for complex tasks
- **Auggie** - Free, good for basic coding

Without murmur, you're either:

1. Always using the paid tool (expensive)
2. Always using the free tool (suboptimal for complex tasks)
3. Manually deciding each time (tedious)

## How Routing Works

```
┌──────────────────────────────────────────────────────────────┐
│                        Your Prompt                           │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                    Prompt Analyzer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Keywords   │  │   Length    │  │    Tool Use?        │  │
│  │  Analysis   │  │   Factor    │  │    Detection        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Complexity: 0.65│
                    │ Category: debug │
                    │ Tool Use: true  │
                    └─────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                    Route Selection                           │
│                                                              │
│   Mode: auto   │   Threshold: 0.5   │   Available: 2 tools  │
│                                                              │
│   Decision: claude (complexity 0.65 >= 0.5, needs tool use)  │
└──────────────────────────────────────────────────────────────┘
```

## Complexity Analysis

### Keyword Scoring

Certain keywords indicate complexity:

**High Complexity (adds 0.25-0.30)**
```
refactor, architecture, redesign, optimize, debug, 
fix bug, performance, security
```

**Medium Complexity (adds 0.10-0.20)**
```
implement, create, build, write, code, function, 
class, module, test
```

**Low Complexity (subtracts 0.05-0.15)**
```
what is, explain, tell me, help, define, meaning, 
simple, basic, quick, how to
```

### Length Factor

| Prompt Length | Factor Added |
|--------------|--------------|
| < 50 chars | 0.0 |
| 50-150 | 0.1 |
| 150-300 | 0.2 |
| 300-500 | 0.3 |
| 500-1000 | 0.5 |
| > 1000 | 0.7 |

### Tool Use Detection

Keywords suggesting file/code operations:

```
file, read file, write file, edit file, create file,
run, execute, test, build, compile,
in this project, in this repo, this code, this file,
current directory, codebase, repository,
modify, change, update the
```

## Routing Modes

### `auto` (Default)

Balances cost and quality:

```yaml
routing:
  mode: auto
  complexity_threshold: 0.5
```

| Complexity | Tool Use | Selection |
|------------|----------|-----------|
| < 0.5 | No | Free tool |
| ≥ 0.5 | No | Paid tool |
| Any | Yes | Paid tool |

### `cost-first`

Aggressively minimizes cost:

```yaml
routing:
  mode: cost-first
```

| Complexity | Tool Use | Selection |
|------------|----------|-----------|
| ≤ 0.8 | No | Free tool |
| > 0.8 | No | Paid tool |
| Any | Yes | Paid tool |

### `quality-first`

Prioritizes best results:

```yaml
routing:
  mode: quality-first
```

| Complexity | Tool Use | Selection |
|------------|----------|-----------|
| < 0.2 | No | Free tool |
| ≥ 0.2 | No | Paid tool |
| Any | Yes | Paid tool |

### `manual`

No automatic routing:

```yaml
routing:
  mode: manual
```

Always uses `default_tool`.

## Category Detection

Murmur categorizes prompts for statistics:

| Category | Trigger Keywords |
|----------|-----------------|
| `simple-qa` | what is, explain, tell me, define |
| `architecture` | refactor, architecture, design, restructure |
| `debugging` | debug, fix, error, bug, issue, crash |
| `analysis` | analyze, review, evaluate, compare |
| `coding` | write, create, implement, build, code |
| `general` | (default) |

## See Routing Decisions

Preview without executing:

```bash
mur run -p "refactor the auth module" --explain
```

```
Routing Decision
================
Prompt:     refactor the auth module
Complexity: 0.72
Category:   architecture
Tool Use:   false
Keywords:   [refactor, module]

Selected:   claude
Reason:     auto: complexity 0.72 >= 0.50 threshold, using paid tool
Fallback:   gemini
```

## Customizing Thresholds

Adjust the complexity threshold to your needs:

```yaml
routing:
  mode: auto
  complexity_threshold: 0.3  # More aggressive (use paid more often)
  # or
  complexity_threshold: 0.7  # Conservative (use free more often)
```

## Override Routing

When you know better:

```bash
# Force Claude
mur run -t claude -p "simple question but I want the best answer"

# Force Gemini  
mur run -t gemini -p "complex task but I'm just experimenting"
```

## See Also

- [run Command](../commands/run.md) - Running prompts
- [Configuration](../getting-started/configuration.md) - Routing configuration
- [stats Command](../commands/stats.md) - View routing effectiveness
