# MUR run

Execute prompts with smart routing to the best AI tool.

## Usage

```bash
mur run -p "your prompt here" [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--prompt` | `-p` | The prompt to run (required) |
| `--tool` | `-t` | Force specific tool (overrides routing) |
| `--explain` | | Show routing decision without executing |

## Smart Routing

When you run a prompt without `-t`, MUR Core analyzes it and routes to the best tool:

```bash
mur run -p "what is a mutex?"
# → gemini (auto: complexity 0.10 < 0.50 threshold, using free tool)
```

### Complexity Analysis

MUR Core scores prompts based on:

1. **Keywords** - Words like "refactor", "architecture", "debug" increase complexity
2. **Length** - Longer prompts tend to be more complex
3. **Tool Use** - Mentions of files, code, or project context suggest tool use

### Routing Decision

| Complexity | Tool Use Needed | Result |
|------------|-----------------|--------|
| < 0.5 | No | Free tool (Gemini) |
| ≥ 0.5 | No | Paid tool (Claude) |
| Any | Yes | Paid tool (Claude) |

## Examples

### Auto-Routing

```bash
# Simple question → free tool
mur run -p "explain dependency injection"

# Complex task → paid tool
mur run -p "refactor this authentication module to use JWT tokens"

# Tool use detected → paid tool
mur run -p "read the config file and fix the bug"
```

### Explain Mode

See why a tool was selected without actually running:

```bash
mur run -p "debug this race condition" --explain
```

Output:

```
Routing Decision
================
Prompt:     debug this race condition
Complexity: 0.65
Category:   debugging
Tool Use:   false
Keywords:   [debug]

Selected:   claude
Reason:     auto: complexity 0.65 >= 0.50 threshold, using paid tool
Fallback:   gemini
```

### Force Tool

Override routing for specific use cases:

```bash
# Use Claude even for simple questions
mur run -t claude -p "what is a pointer?"

# Use Gemini even for complex tasks
mur run -t gemini -p "refactor everything"
```

## Statistics

Every run is tracked for analytics:

```bash
mur stats
```

See [stats command](stats.md) for details.

## See Also

- [Smart Routing](../concepts/routing.md) - How routing works in detail
- [Configuration](../getting-started/configuration.md) - Routing modes and thresholds
