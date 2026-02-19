# mur verify & mur preview

Pattern integrity verification and security preview.

## mur verify

Checks SHA256 hashes of all patterns to detect tampering or corruption.

```bash
# Check all patterns
mur verify

# Recalculate hashes (after intentional edits)
mur verify --fix
```

### Output

```
🔍 Verifying 147 patterns...

  ✓ go-error-handling         — hash OK
  ✓ api-retry-pattern         — hash OK
  ⚠ docker-compose-setup     — HASH MISMATCH (content modified)
  ✓ swift-async-await         — hash OK
  ...

Results: 145 OK | 2 mismatched | 0 missing hash
```

### When to use

- After pulling community patterns
- After team sync (verify nothing was corrupted in transit)
- After any suspicious system behavior
- `--fix` after you intentionally edit a pattern file

## mur preview

Inspect a pattern's content, trust level, and security scan results before enabling it.

```bash
mur preview "api-retry-pattern"
```

### Output

```
📋 Pattern: api-retry-pattern
─────────────────────────────

Trust Level:    owner
Hash Status:    ✓ valid
Injection Risk: low
Schema Version: 2

Content:
  ────────────────────────
  # API Call Retry Pattern
  
  Implement a retry mechanism with exponential
  backoff for failed API calls...
  ────────────────────────

Security Scan:
  ✓ No injection patterns detected
  ✓ No secrets detected
```

### When to use

- Before enabling a community pattern you downloaded
- Reviewing team-shared patterns from a new contributor
- Debugging why a pattern was blocked during injection
