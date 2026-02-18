# Pricing & Cost Optimization

## MUR Core Subscription Plans

| Plan | Price | Features |
|------|-------|----------|
> ⚠️ **See canonical pricing:** `~/Projects/mur-server/docs/product/PRICING.md`

| **Free** | $0/forever | 50 patterns, 2 CLI integrations, local storage, community support |
| **Pro** | $9/month | Unlimited patterns, all CLI integrations, semantic search, cloud sync, priority support |
| **Team** | $49/month flat + $10/extra | Everything in Pro + 5 team members, shared library, admin dashboard |

## How MUR Core Saves Money

MUR Core helps you save money by intelligently routing prompts to free tools when appropriate.

### The Core is Open Source

**MUR Core CLI is 100% free and open source.**

Your costs come from the underlying AI tools:

| Tool | Cost |
|------|------|
| Claude Code | Paid (Anthropic API) |
| Gemini CLI | Free |
| Auggie | Free |
| OpenClaw | Paid (subscription) |

## How Murmur Saves Money

### Smart Routing

Simple questions go to free tools:

```bash
mur run -p "what is a mutex?"
# → gemini (free)
# Saved: ~$0.01
```

Complex tasks go to paid tools only when needed:

```bash
mur run -p "refactor this authentication module"
# → claude (paid, but worth it)
```

### Cost Tracking

See your savings:

```bash
mur stats
```

```
Cost Analysis:
  Estimated spend: $12.45
  Saved by routing: $8.20 (66 prompts → free tool)
```

### Routing Modes

Adjust how aggressively murmur saves:

```yaml
routing:
  mode: cost-first  # Maximum savings
```

| Mode | Strategy |
|------|----------|
| `auto` | Balanced (default) |
| `cost-first` | Maximize free tool usage |
| `quality-first` | Maximize paid tool usage |
| `manual` | You decide |

## Cost Estimates

Approximate costs per request (varies by length):

| Tool | Short Prompt | Medium | Long |
|------|--------------|--------|------|
| Claude Code | $0.005 | $0.01 | $0.03 |
| Gemini CLI | Free | Free | Free |
| Auggie | Free | Free | Free |

!!! note
    Actual costs depend on your API plan, prompt length, and response length.

## Optimization Tips

### 1. Use `--explain` First

Preview routing before running:

```bash
mur run -p "complex task" --explain
# See if it routes to paid tool
# Adjust prompt if needed
```

### 2. Batch Simple Questions

Instead of asking Claude multiple simple questions, batch them or use Gemini:

```bash
# Don't do this:
mur run -t claude -p "what is X?"
mur run -t claude -p "what is Y?"
mur run -t claude -p "what is Z?"

# Do this:
mur run -p "what is X?"  # Auto-routes to Gemini
mur run -p "what is Y?"
mur run -p "what is Z?"
```

### 3. Adjust Threshold

Lower threshold = more free tool usage:

```yaml
routing:
  complexity_threshold: 0.3  # More conservative
```

Higher threshold = more paid tool usage:

```yaml
routing:
  complexity_threshold: 0.7  # Only complex tasks to Claude
```

### 4. Use `cost-first` Mode

When budget is tight:

```bash
mur config routing cost-first
```

Only very complex tasks (>0.8 complexity) go to paid tools.

### 5. Review Stats Weekly

```bash
mur stats --days 7
```

Check if you're using paid tools for simple tasks.

## Free Tier Maximization

To maximize free usage:

1. Enable Gemini CLI as primary free tool
2. Set `mode: cost-first`
3. Use `--explain` to understand routing
4. Review weekly stats

```yaml
# Maximum savings config
routing:
  mode: cost-first
  complexity_threshold: 0.8

tools:
  gemini:
    enabled: true
    tier: free
  claude:
    enabled: true
    tier: paid
```

## Enterprise Pricing

For teams, consider:

1. **Pattern Sharing** - Avoid repeated explanations
2. **Team Knowledge Base** - Collective learning reduces queries
3. **Usage Analytics** - Track team-wide costs

```bash
mur team sync
mur stats --team
```

## See Also

- [Smart Routing](concepts/routing.md) - How routing works
- [stats Command](commands/stats.md) - Usage analytics
- [Configuration](getting-started/configuration.md) - Routing settings
