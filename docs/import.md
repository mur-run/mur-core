# Import Patterns

Import patterns from external sources like GitHub Gists, URLs, or files.

## Import from GitHub Gist

Import patterns directly from GitHub Gists:

```bash
# From full URL
mur import gist https://gist.github.com/user/abc123def456

# Just the gist ID
mur import gist abc123def456

# With custom name
mur import gist abc123def456 --name "My Custom Pattern"

# Import and share to community
mur import gist abc123def456 --share
```

### Example Output

```
📥 Fetching gist abc123def456...
   Found: Swift Error Handling Best Practices
   Author: @karajanchang
   Pattern: Swift Error Handling

✓ Imported "Swift Error Handling" to ~/.mur/patterns/

Run 'mur sync' to sync to your CLIs
```

### Supported Gist Formats

MUR supports two gist formats:

#### 1. Pattern YAML (Recommended)

Include a `pattern.yaml` or `pattern.yml` file in your gist:

```yaml
# pattern.yaml
name: Swift Error Handling
description: Best practices for handling errors in Swift

content: |
  When handling errors in Swift:
  1. Use Result<Success, Error> for functions that can fail
  2. Prefer throwing functions for synchronous operations
  3. Use async/await with try for async operations

tags:
  confirmed: [swift, error-handling, async]

applies:
  languages: [swift]
  keywords: [error, Result, throw, catch, try]

schema_version: 2
```

#### 2. README + Code Files

If no `pattern.yaml` exists, MUR will construct a pattern from:

- **README.md** - Used for description (first `# heading` becomes the name)
- **Code files** - Detected by language and included in pattern content

Example gist structure:
```
README.md        → Pattern description
example.swift    → Code example (auto-detected as Swift)
test.swift       → Additional code
```

MUR will combine these into a pattern with proper code blocks.

### Flags

| Flag | Description |
|------|-------------|
| `--name, -n` | Override the pattern name |
| `--share` | Share to community after import |

## Import from Files

Import patterns from local YAML files:

```bash
# Single file
mur import patterns.yaml

# Multiple files with glob
mur import ./my-patterns/*.yaml

# Preview without importing
mur import patterns.yaml --dry-run

# Overwrite existing patterns
mur import patterns.yaml --force
```

### Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be imported without importing |
| `--force, -f` | Overwrite existing patterns |

## Import from URLs

Import patterns from remote URLs:

```bash
# HTTPS URL
mur import https://example.com/patterns.yaml

# GitHub raw URL
mur import https://raw.githubusercontent.com/user/repo/main/patterns.yaml
```

## Creating Shareable Gists

To share patterns via gist:

1. Create a gist on [gist.github.com](https://gist.github.com)
2. Add your `pattern.yaml` file
3. Share the gist URL

Others can import with:

```bash
mur import gist <your-gist-id>
```

### Tips for Good Gists

- Use descriptive gist titles
- Include a clear description
- Add relevant code examples
- Use the `pattern.yaml` format for best results

## Workflow Example

```bash
# 1. Find an interesting gist
# (browse GitHub Gists or get a link from someone)

# 2. Import it
mur import gist abc123

# 3. Review the imported pattern
mur edit "Imported Pattern Name"

# 4. Sync to your CLIs
mur sync

# 5. Optionally share to community
mur community share "Imported Pattern Name"
```

## Related Commands

- [Community](community.md) - Browse and share community patterns
- [Collections](collections.md) - Curate groups of patterns
