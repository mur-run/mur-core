# Community Patterns

MUR has a built-in community for sharing and discovering patterns. Browse patterns shared by developers worldwide, or share your own.

## Browsing Patterns

```bash
# See popular patterns
mur community

# Search community
mur community search "API error handling"

# View featured patterns (curated by maintainers)
mur community featured

# View recently shared patterns
mur community recent
```

### Example Output

```
🌍 Community Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Popular:
  1. Swift Error Handling (⭐ 142) by @karajanchang
     Use Result types and async/await for clean error handling
  2. API Retry with Backoff (⭐ 98) by @johndoe
     Exponential backoff strategy for resilient API calls
```

## Searching Patterns

```bash
# Basic search
mur community search "error handling"

# Limit results
mur community search "testing" --limit 5

# Include community in local semantic search
mur search --community "best practices"
```

## Copying Patterns

Found a pattern you like? Copy it to your collection:

```bash
mur community copy "API retry with backoff"

# Copy to a specific team
mur community copy "my-pattern" --team my-team-id

# Then sync to your CLIs
mur sync
```

## Sharing Your Patterns

Share your patterns with the community:

```bash
# Basic share (submits for review)
mur community share "my-awesome-pattern"

# With category and tags
mur community share "my-pattern" \
  --category "Error Handling" \
  --tags "api,retry,resilience"

# With custom description
mur community share "my-pattern" \
  --description "A better way to handle API errors"
```

### Requirements

- Must be logged in: `mur login`
- Pattern must exist in your team
- Patterns are reviewed before being published

### Categories

Choose from these categories when sharing:

| Category | Description |
|----------|-------------|
| Error Handling | Error handling, recovery, resilience |
| Testing | Unit tests, integration tests, mocking |
| API Design | REST, GraphQL, API patterns |
| Performance | Optimization, caching, profiling |
| Security | Auth, encryption, secure coding |
| Code Quality | Linting, formatting, best practices |
| Debugging | Logging, tracing, debugging techniques |
| Documentation | Comments, docs, README patterns |
| DevOps | CI/CD, deployment, infrastructure |
| AI/ML | Machine learning, AI integrations |

### Tags

Add comma-separated tags for discoverability:

```bash
--tags "swift,ios,concurrency,async-await"
```

## User Profiles

View other developers' profiles and their shared patterns:

```bash
mur community user karajanchang
```

### Example Output

```
👤 Karajan Chang (@karajanchang)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Swift developer, AI enthusiast

   📊 12 patterns | ⬇️ 1,234 copies | ⭐ 89 stars

   🌐 example.com
   🐙 github.com/karajanchang
   🐦 @karajanchang

Patterns:
   • Swift Error Handling (⬇️ 142)
   • Async/Await Best Practices (⬇️ 98)
   • SwiftUI State Management (⬇️ 76)
```

## Search Options

Fine-tune your searches:

```bash
# Include community in local search
mur search --community "error handling"

# Search only community (skip local)
mur search --community-only "API patterns"

# Search only local patterns
mur search --local "my patterns"
```

### Tech Stack Filtering

Set your tech stack to filter community results:

```bash
mur config set tech_stack "swift,go,docker"

# Now community searches prioritize these technologies
mur search --community "best practices"
```

## Web Dashboard

Visit [app.mur.run](https://app.mur.run) to:

- Browse and search patterns with a visual interface
- Manage your collections
- View user profiles
- Star and copy patterns
- Track your contribution stats

## Privacy & Moderation

- Your local patterns are **never** shared unless you explicitly use `mur community share`
- Shared patterns are reviewed before being published
- You can report inappropriate patterns via the web dashboard
- Authors can update or unpublish their patterns anytime

## Related Commands

- [Collections](collections.md) - Curate groups of patterns
- [Import](import.md) - Import patterns from GitHub Gists
