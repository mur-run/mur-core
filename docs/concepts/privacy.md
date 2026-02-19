# Privacy & PII Protection

mur takes privacy seriously. When you share patterns with the community, a multi-layer pipeline ensures your personal and organizational information stays private.

## The Problem

Patterns are extracted from real coding sessions. They naturally contain:

- Company names ("Fix bug in GlobalOK admin panel")
- Person names ("David's deployment script")
- Email addresses (karajan@company.com)
- Internal IPs (192.168.1.100)
- File paths (/Users/david/Projects/secret-project)
- Internal URLs (admin.company.internal)

Without filtering, sharing a pattern means sharing your identity.

## 4-Layer Protection Pipeline

Every pattern goes through this pipeline before community upload:

```
Pattern content
  │
  ├── Layer 1: Regex PII Scanner
  │     Emails, internal IPs, file paths, phone numbers
  │     → Auto-redacted: david@co.com → <REDACTED_EMAIL>
  │
  ├── Layer 2: User Blocklist
  │     Your custom terms from config
  │     → Replaced: "GlobalOK" → <COMPANY>
  │
  ├── Layer 3: LLM Semantic Anonymization (opt-in)
  │     AI-powered detection of names, projects, metrics
  │     → Catches what regex can't
  │
  ├── Layer 4: Secret Scanner (existing)
  │     API keys, tokens, passwords
  │     → Pattern SKIPPED entirely if found
  │
  └── Preview → You confirm → Upload
```

## Configuration

```yaml
privacy:
  # Custom blocklist — add your company/project names
  redact_terms:
    - "my-company"
    - "secret-project"
    - "john.doe"

  # Custom replacements
  replacements:
    "my-company.com": "<COMPANY_DOMAIN>"
    "ProjectX": "<PROJECT>"

  # Auto-detection rules (all default to true)
  auto_detect:
    emails: true
    internal_ips: true
    file_paths: true
    phone_numbers: true
    internal_urls: true

  # LLM-powered anonymization (opt-in)
  semantic_anonymization:
    enabled: false
    provider: "ollama"
    model: "llama3.2"
    cache_results: true
```

## Preview Before Sharing

Always see what will be uploaded:

```bash
mur community share "my-pattern" --dry-run
```

```
📋 Pattern: my-pattern
🔍 Privacy scan results:

  ⚠️  Email found: david@glo... → <REDACTED_EMAIL>
  ⚠️  Path found: /Users/david/... → <REDACTED_PATH>
  ⚠️  Blocklist match: "GlobalOK" → <COMPANY>

📝 Cleaned content preview:
  ────────────────────────
  # API Call Retry Pattern for <COMPANY>
  When deploying to <COMPANY> servers, use exponential
  backoff with jitter...
  ────────────────────────

Share this version? [y/n]
```

## Key Differences from Secret Scanner

| | Secret Scanner | PII Scanner |
|--|---------------|-------------|
| Detects | API keys, tokens, passwords | Names, emails, IPs, paths |
| Action | **Skip entire pattern** | **Redact and continue** |
| False positives | Low (specific patterns) | Higher (broader matching) |
| User config | None needed | Blocklist recommended |

The secret scanner is strict (a leaked API key is catastrophic). The PII scanner is conservative (redacts in-place, you review before sharing).

## LLM Semantic Anonymization

Regex catches structured data. But what about:

> "Our e-commerce platform processes 500K orders/day. CEO Zhang wants to migrate to microservices..."

- "e-commerce platform" + "500K orders" → company identifiable by scale
- "CEO Zhang" → person identifiable
- Regex catches neither

With `semantic_anonymization.enabled: true`, an LLM analyzes content and replaces identifying information while preserving technical value. Results are cached to avoid re-processing.

!!! note "Requires Ollama"
    Semantic anonymization currently requires a local Ollama instance. Cloud LLM support (OpenAI, Anthropic) is planned.
