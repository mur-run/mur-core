# MUR Demo GIF
# Run: vhs demo.tape

Output assets/demo.gif

Set FontSize 16
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

# Clear screen before recording
Hide
Type "clear"
Enter
Sleep 500ms
Show

# Title
Type "# MUR - Continuous learning for AI assistants"
Enter
Sleep 1s

# Step 1: Show status
Type "mur status"
Enter
Sleep 2s

# Step 2: Create a pattern
Type "# Create a pattern"
Enter
Sleep 500ms
Type "mur new swift-testing"
Enter
Sleep 1s

# Step 3: Show pattern created
Type "# Pattern created! Now sync to AI tools"
Enter
Sleep 500ms
Type "mur sync"
Enter
Sleep 2s

# Step 4: Search
Type "# Search patterns by meaning"
Enter
Sleep 500ms
Type 'mur search "async testing"'
Enter
Sleep 2s

# Step 5: Done
Type "# Done! Your AI CLIs now know your patterns 🎉"
Enter
Sleep 2s
