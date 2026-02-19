# 44: Pattern Integrity & Security

**Status:** implemented
**Priority:** P1 — Phase 2
**Estimated effort:** 3.5 days

## Proposal

Pattern schema v2 has `Security.Hash` and `TrustLevel` fields but they're not actively enforced. Community patterns could contain prompt injection payloads. No audit trail exists for pattern injection history. This spec adds active integrity verification, injection scanning, and audit logging.

### Goals
- Verify SHA256 hash on pattern load; warn on mismatch
- Scan for known prompt injection patterns before injection
- `mur verify` command for manual integrity check
- `mur preview <pattern>` for community pattern inspection
- Audit log: which pattern → when → injected into what → source

### Non-goals
- Runtime sandboxing of pattern content
- ML-based injection detection (future)

## Design

### Integrity Verification

In pattern loading path (`internal/core/pattern/store.go`):

```go
func (s *Store) LoadVerified(name string) (*Pattern, error) {
    p, err := s.Get(name)
    if err != nil { return nil, err }
    
    if p.Security.Hash != "" && !p.VerifyHash() {
        p.Security.Warnings = append(p.Security.Warnings,
            "hash mismatch: content may have been tampered with")
        p.Security.TrustLevel = TrustUntrusted
    }
    return p, nil
}
```

### Injection Scanner (`internal/security/injection.go`)

```go
type InjectionScanner struct {
    rules []injectionRule
}

type InjectionRisk string // low | medium | high

func (s *InjectionScanner) Scan(content string) (InjectionRisk, []InjectionFinding)
```

Detection patterns:
- "ignore previous instructions"
- "system:", "assistant:", "user:" role markers
- "you are now", "act as", "pretend to be"
- Base64-encoded instruction blocks
- Markdown/HTML comment injection
- Unicode homoglyph obfuscation

### Audit Log (`internal/core/audit/`)

```go
type AuditEntry struct {
    Timestamp   time.Time
    PatternID   string
    PatternName string
    Action      string   // "inject", "load", "share", "modify"
    Source      string   // "hook", "cli", "sync"
    ToolTarget  string   // "claude", "cursor", etc.
    PromptHash  string   // SHA256 of prompt (not full text)
}
```

Storage: `~/.mur/audit/audit.jsonl` (append-only, rotated monthly)

### CLI

```bash
mur verify                  # check all patterns
mur verify --fix            # recalculate hashes
mur preview "pattern-name"  # show content + scan results + trust level
mur audit                   # show recent audit log
mur audit --pattern "name"  # filter by pattern
```

## Tasks

- [x] Implement hash verification on pattern load
- [x] Add `mur verify` command (check all, --fix to recalculate)
- [x] Implement `InjectionScanner` with rule set
- [x] Integrate injection scan into hook injection pipeline
- [x] Add `injection_risk` field to pattern Security metadata
- [x] Implement audit logger (append-only JSONL)
- [x] Integrate audit logging into inject and sync paths
- [x] Add `mur audit` command with filters
- [x] Add `mur preview` command for community patterns
- [x] Block high-risk community patterns by default (configurable)
- [x] Unit tests for injection patterns
- [x] Unit tests for audit log rotation
