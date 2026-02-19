# Security

How mur protects your AI tools from malicious patterns.

## Threat Model

Patterns are injected into AI prompts. A malicious pattern could:

1. **Hijack the AI's behavior** — "Ignore previous instructions and..."
2. **Exfiltrate data** — Trick the AI into leaking sensitive information
3. **Degrade output quality** — Conflicting or misleading advice

This is especially relevant for **community patterns** and **team-shared patterns** from untrusted contributors.

## Injection Scanner

Every pattern is scanned before injection with 11 detection rules:

| Rule | Risk | Example |
|------|------|---------|
| Instruction override | 🔴 High | "ignore previous instructions" |
| New instruction injection | 🔴 High | "new instructions: ..." |
| System role marker | 🔴 High | "system: you are now..." |
| Role hijacking | 🔴 High | "act as", "pretend to be" |
| Assistant/User role markers | 🟡 Medium | "assistant: ", "user: " |
| Base64 encoded blocks | 🟡 Medium | Hidden instructions in base64 |
| HTML/Markdown comment injection | 🟡 Medium | `<!-- hidden prompt -->` |
| Unicode homoglyphs | 🟡 Medium | Lookalike characters hiding text |
| Delimiter injection | 🟡 Medium | "---END---" breaking prompt structure |
| Jailbreak keywords | 🟡 Medium | "DAN", "jailbreak" |

### Blocking Policy

- **High-risk + untrusted source** → Pattern **blocked** (not injected)
- **High-risk + trusted source (owner/team)** → Pattern injected with warning
- **Medium/Low risk** → Pattern injected normally

Trust levels: `owner` > `team` > `verified` > `community` > `untrusted`

## Integrity Verification

Every pattern has a SHA256 hash stored in its metadata:

```yaml
security:
  hash: "a1b2c3d4..."
  trust_level: owner
  injection_risk: low
```

On load, the hash is verified. A mismatch means the file was modified outside of mur — it gets flagged as `untrusted` regardless of its original trust level.

```bash
# Check all patterns
mur verify

# Recalculate after intentional edits
mur verify --fix
```

## Audit Trail

Every pattern injection is logged to `~/.mur/audit/audit.jsonl`:

```json
{
  "timestamp": "2026-02-19T19:30:12Z",
  "pattern_id": "abc123",
  "pattern_name": "api-retry-pattern",
  "action": "inject",
  "source": "hook",
  "prompt_hash": "sha256..."
}
```

- Append-only (tamper-evident)
- Auto-rotates at 10MB
- Prompt text is **never stored** — only its SHA256 hash

```bash
# View recent events
mur audit

# Filter by pattern
mur audit --pattern "api-retry-pattern"
```

## Secret Scanner

Separate from injection scanning, the secret scanner prevents accidental credential leakage when sharing patterns:

- 30+ secret types: AWS, GitHub, OpenAI, Anthropic, Stripe, Slack, Discord, etc.
- Patterns with detected secrets are **skipped entirely** (not shared)
- Server-side validation as defense in depth

## Recommendations

1. **Always run `mur preview`** before enabling community patterns
2. **Set up a blocklist** in `privacy.redact_terms` for your org
3. **Review `mur audit`** periodically for unexpected injection patterns
4. **Use `mur verify`** after pulling from team/community sources
5. **Keep `mur consolidate`** running weekly to catch conflicting patterns
