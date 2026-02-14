# Collections

Collections are curated groups of related patterns. Use them to organize patterns by topic, project, or workflow.

## Listing Collections

```bash
# List public collections
mur collection list

# Alias: mur collections
mur collections
```

### Example Output

```
📚 Public Collections
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  📁 Swift Best Practices (⬇️ 234)
     Essential patterns for Swift development
     ID: col_abc123

  📁 API Design Patterns (⬇️ 189)
     RESTful API design guidelines
     ID: col_def456
```

## Viewing a Collection

```bash
mur collection show <collection-id>
```

### Example

```bash
mur collection show col_abc123
```

Output:

```
📁 Swift Best Practices
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Essential patterns for Swift development

Patterns (8):
  • Swift Error Handling (⬇️ 142)
  • Async/Await Best Practices (⬇️ 98)
  • SwiftUI State Management (⬇️ 76)
  • Codable Strategies (⬇️ 54)
```

## Creating a Collection

```bash
# Basic creation (private by default)
mur collection create "My Collection"

# With description
mur collection create "Swift Best Practices" \
  --description "Essential patterns for Swift development"

# Public collection (visible to everyone)
mur collection create "API Patterns" --visibility public

# Private collection (only you can see)
mur collection create "Work Patterns" --visibility private
```

### Visibility Options

| Visibility | Description |
|------------|-------------|
| `private` | Only you can see (default) |
| `public` | Visible to everyone on mur community |

### Requirements

- Must be logged in: `mur login`

## Example Workflow

```bash
# 1. Create a collection
mur collection create "My API Patterns" --visibility public

# Output:
# ✓ Created collection "My API Patterns"
#   ID: col_xyz789
#   Visibility: public
#
# Add patterns with: mur collection add <collection-id> <pattern-id>

# 2. Add patterns (via web dashboard at app.mur.run)

# 3. Share the collection ID with your team
# They can view it with:
mur collection show col_xyz789
```

## Managing Collections

Collections can be fully managed via the web dashboard at [app.mur.run](https://app.mur.run):

- Add/remove patterns
- Edit name and description
- Change visibility
- Delete collections

## Use Cases

### Team Onboarding

Create a collection of patterns new team members should learn:

```bash
mur collection create "Team Onboarding" --visibility private
```

### Conference Talks

Share patterns from your talk:

```bash
mur collection create "SwiftConf 2024 Patterns" --visibility public
```

### Project-Specific

Group patterns for a specific project:

```bash
mur collection create "Project X Patterns" --visibility private
```

## Related Commands

- [Community](community.md) - Browse and share community patterns
- [Import](import.md) - Import patterns from GitHub Gists
