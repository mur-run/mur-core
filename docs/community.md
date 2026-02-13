# Community Patterns

MUR has a built-in community for sharing and discovering patterns. Browse patterns shared by developers worldwide, or share your own.

## Browsing Patterns

```bash
# See popular patterns
mur community

# Search community
mur community search "API error handling"

# View featured patterns
mur community featured

# View recent patterns
mur community recent
```

## Copying Patterns

Found a pattern you like? Copy it to your collection:

```bash
mur community copy "API retry with backoff"
mur sync  # Sync to your CLIs
```

## Sharing Your Patterns

Share your patterns with the community:

```bash
# Basic share
mur community share "my-awesome-pattern"

# With category and tags
mur community share "my-pattern" \
  --category "Error Handling" \
  --tags "api,retry,resilience"
```

### Categories

Choose from these categories when sharing:
- Error Handling
- Testing
- API Design
- Performance
- Security
- Code Quality
- Debugging
- Documentation
- DevOps
- AI/ML

## User Profiles

View other developers' profiles and their shared patterns:

```bash
mur community user karajanchang
```

## Collections

Collections are curated groups of patterns:

```bash
# Browse public collections
mur collection list

# View a collection
mur collection show <collection-id>

# Create your own collection
mur collection create "Swift Best Practices" --visibility public
```

## Import from GitHub Gist

Import patterns directly from GitHub Gists:

```bash
# From URL
mur import gist https://gist.github.com/user/abc123

# Just the ID
mur import gist abc123

# Import and share immediately
mur import gist abc123 --share
```

The gist should contain either:
- A `pattern.yaml` file (mur format)
- A `README.md` with description + code files

## Search Options

Fine-tune your community searches:

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

## Dashboard

Visit [app.mur.run](https://app.mur.run) to:
- Browse and search patterns with a visual interface
- Manage your collections
- View user profiles
- Star and copy patterns
- Track your contribution stats

## Privacy

- Your local patterns are **never** shared unless you explicitly use `mur community share`
- Shared patterns are reviewed before being published
- You can make collections private or public
