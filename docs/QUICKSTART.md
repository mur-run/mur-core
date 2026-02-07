# Quick Start Guide

Get started with mur in 5 minutes.

## Installation

```bash
# Option 1: Homebrew (macOS)
brew install mur-run/tap/mur

# Option 2: Go install
go install github.com/mur-run/mur-core/cmd/mur@latest

# Option 3: Download binary
curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv mur /usr/local/bin/
```

## Initialize

```bash
mur init
```

This creates `~/.murmur/config.yaml` with default settings.

## Run Your First Prompt

```bash
# Simple question → routes to free tool
mur run -p "what is dependency injection?"

# Complex task → routes to paid tool
mur run -p "refactor this authentication module for better security"

# See the routing decision
mur run -p "fix this bug" --explain
```

## Create Your First Pattern

```bash
# Interactive creation
mur learn add swift-error-handling

# Or create manually
cat > ~/.murmur/patterns/swift-error-handling.yaml << 'EOF'
name: swift-error-handling
description: Handle errors with Result types
content: |
  When handling errors in Swift:
  1. Use Result<Success, Error> for functions that can fail
  2. Prefer throwing functions for synchronous operations
  3. Use async/await with try for async operations

tags:
  confirmed: [swift, error-handling]
  
applies:
  languages: [swift]
  keywords: [error, Result, throw, catch]

security:
  trust_level: owner

learning:
  effectiveness: 0.5

lifecycle:
  status: active
  
schema_version: 2
EOF
```

## See Patterns in Action

```bash
# Navigate to a Swift project
cd ~/Projects/MySwiftApp

# Run with verbose to see injected patterns
mur run -v -p "fix the network error handling"

📚 Injected 2 patterns:
   • swift-error-handling
   • network-retry-logic
🔍 Context: swift project (MySwiftApp)

→ claude (auto: complexity 0.72) [2 patterns]
```

## Give Feedback

After using a pattern, rate its helpfulness:

```bash
# Pattern was helpful
mur feedback helpful swift-error-handling

# Pattern wasn't useful
mur feedback unhelpful debugging-tips -c "too generic"

# Check effectiveness scores
mur pattern-stats
```

## Enable Semantic Search

For better pattern matching using meaning (not just keywords):

```bash
# Option 1: Local with Ollama (free)
ollama pull nomic-embed-text
mur embed index

# Option 2: OpenAI (paid)
export OPENAI_API_KEY="sk-..."
mur embed index

# Test search
mur embed search "handle network timeout errors"
```

## Sync to Other AI CLIs

Share patterns across all your AI tools:

```bash
mur sync patterns

✓ Claude Code: synced 5 patterns to ~/.claude/skills/
✓ Gemini CLI: synced 5 patterns to ~/.gemini/skills/
✓ Codex: synced 5 patterns to ~/.codex/instructions.md
```

## Extract Patterns from History

Learn from your existing AI sessions:

```bash
# See what sources are available
mur cross-learn status

# Extract patterns from all CLI histories
mur cross-learn scan --interactive

# Or scan markdown/session files
mur suggest scan ~/notes/ai-sessions/
```

## Auto-Cleanup

Let mur manage pattern lifecycle:

```bash
# Check what needs attention
mur lifecycle evaluate

# Apply recommendations
mur lifecycle apply

# Manual deprecation
mur lifecycle deprecate old-pattern -r "outdated approach"
```

## Team Sync (Optional)

Share patterns with your team via Git:

```bash
# Initialize team repo
mur team init git@github.com:myorg/patterns.git

# Push your patterns
mur team push

# Pull team patterns
mur team pull
```

## What's Next?

- **View stats**: `mur stats`
- **Web dashboard**: `mur serve`
- **Full docs**: [docs/](./docs/)
- **Pattern Schema v2**: [docs/SCHEMA.md](./docs/SCHEMA.md)

## Common Workflows

### Daily Development

```bash
# Morning: pull team patterns
mur team pull

# During work: patterns auto-inject
mur run -p "fix the login timeout"

# After task: give feedback
mur feedback helpful auth-timeout-pattern

# Evening: push learnings
mur team push
```

### New Team Member

```bash
# Install and init
mur init
mur team init git@github.com:myorg/patterns.git

# Get all team patterns instantly
mur team pull
mur sync patterns

# Start working with team knowledge
mur run -p "how do we handle errors in this codebase?"
```

### Pattern Extraction Session

```bash
# Review recent sessions
mur suggest scan ~/.claude/projects/ --interactive

# Accept good patterns
mur suggest accept auth-flow-pattern

# Index for search
mur embed index

# Verify
mur embed search "authentication"
```
