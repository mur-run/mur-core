#!/bin/bash
# Batch add tags to existing patterns using AI

PATTERNS_DIR="$HOME/.mur/patterns"

for file in "$PATTERNS_DIR"/*.yaml; do
    name=$(basename "$file" .yaml)
    
    # Check if already has tags
    if grep -q "^tags:" "$file"; then
        echo "⏭️  Skip (has tags): $name"
        continue
    fi
    
    # Get content
    content=$(cat "$file" | head -50)
    
    # Use Claude to generate tags
    echo "🏷️  Generating tags for: $name"
    
    tags=$(cat <<EOF | claude --print -m sonnet
Based on this pattern, generate 3-5 relevant tags as a YAML array.
Only output the YAML array, nothing else.
Example output: ["swift", "swiftui", "macos"]

Pattern:
$content
EOF
)
    
    if [[ "$tags" == "["* ]]; then
        # Insert tags after description line
        sed -i '' "/^description:/a\\
tags: $tags
" "$file"
        echo "✅ Added tags: $tags"
    else
        echo "❌ Failed to generate tags"
    fi
    
    sleep 1  # Rate limiting
done

echo "Done!"
