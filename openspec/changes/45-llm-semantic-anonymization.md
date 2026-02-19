# 45: LLM Semantic Anonymization

**Status:** implemented
**Priority:** P2 — Phase 3
**Estimated effort:** 2.5 days
**Depends on:** #42 (PII protection baseline)

## Proposal

Regex-based PII scanning (#42) catches structured data (emails, IPs, paths) but misses semantic identifiers: company names in context, person names without @ signs, internal project codenames, business metrics that reveal company identity. Add an LLM-powered anonymization layer for community sharing.

### Goals
- LLM-based content analysis to detect semantic PII
- Replace identifying info while preserving technical teaching value
- Optional layer (off by default, enabled via config)
- Works with Ollama (local) or cloud LLM providers
- Caches anonymization results to avoid re-processing

### Non-goals
- Real-time anonymization during hook execution
- Guaranteed 100% PII removal (best-effort with human review)

## Design

### Package: `internal/security/anonymize.go`

```go
type SemanticAnonymizer struct {
    llm     LLMClient
    cache   map[string]string  // content hash → cleaned content
}

func (a *SemanticAnonymizer) Anonymize(content string) (string, []AnonymizationChange, error)

type AnonymizationChange struct {
    Original string
    Replaced string
    Category string  // "company", "person", "project", "metric", "location"
    Line     int
}
```

### LLM Prompt

```
Analyze this technical pattern and replace ALL identifying information:
- Company/org names → <COMPANY>
- Person names → <PERSON>
- Project codenames → <PROJECT>
- Specific metrics identifying a company → <METRIC>
- Internal jargon/codenames → generic terms
- Location-specific references → <LOCATION>

Keep ALL technical teaching value intact. Only replace identifying parts.
Output the cleaned version only.
```

### Integration into share flow

```
Pattern → Regex PII (#42) → Blocklist (#42) → LLM Semantic [THIS] → Secret scan → Preview → Upload
```

### Config

```yaml
privacy:
  semantic_anonymization:
    enabled: false            # opt-in
    provider: "ollama"        # ollama | openai | anthropic
    model: "llama3.2"         # model for anonymization
    cache_results: true
```

## Tasks

- [x] Implement `SemanticAnonymizer` with LLM prompt
- [x] Add Ollama integration for local anonymization
- [ ] Add cloud LLM fallback (OpenAI/Anthropic)
- [x] Implement result caching (content hash → cleaned)
- [x] Integrate into share flow after regex PII scan
- [x] Add config options for semantic anonymization
- [x] Show LLM-detected changes in --dry-run preview
- [x] Unit tests with mock LLM responses
- [ ] Integration test: end-to-end share with anonymization
