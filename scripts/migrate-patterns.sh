#!/bin/bash
# Migration script: Convert old murmur-ai patterns to new format
# Usage: ./migrate-patterns.sh

OLD_DIR="${HOME}/clawd/skills/murmur-ai/learned"
NEW_DIR="${HOME}/.murmur/patterns"

# Create new directory if not exists
mkdir -p "$NEW_DIR"

# Count patterns
count=0

# Find all .md files in old directory
find "$OLD_DIR" -name "*.md" | while read -r file; do
    # Extract filename without path and extension
    filename=$(basename "$file" .md)
    
    # Extract frontmatter values using grep/sed
    name=$(grep "^name:" "$file" | sed 's/name: //')
    confidence=$(grep "^confidence:" "$file" | sed 's/confidence: //')
    category=$(grep "^category:" "$file" | sed 's/category: //')
    domain=$(grep "^domain:" "$file" | sed 's/domain: //')
    times_seen=$(grep "^times_seen:" "$file" | sed 's/times_seen: //')
    
    # Extract content (everything after the second ---)
    content=$(awk '/^---$/{n++; next} n>=2' "$file")
    
    # Map confidence to float
    case "$confidence" in
        HIGH) conf_val="0.9" ;;
        MEDIUM) conf_val="0.7" ;;
        LOW) conf_val="0.5" ;;
        *) conf_val="0.7" ;;
    esac
    
    # Create new YAML file
    new_file="$NEW_DIR/${filename}.yaml"
    
    cat > "$new_file" << EOF
name: ${name:-$filename}
description: "Migrated from old murmur-ai"
domain: ${domain:-_global}
category: ${category:-pattern}
confidence: ${conf_val}
times_seen: ${times_seen:-1}
content: |
$(echo "$content" | sed 's/^/  /')
EOF
    
    echo "✓ Migrated: $filename"
    ((count++))
done

echo ""
echo "Migration complete!"
echo "Patterns migrated to: $NEW_DIR"
echo ""
echo "Next steps:"
echo "1. Run 'mur learn list' to verify"
echo "2. Run 'mur learn push' to sync to learning repo"
