# Change Spec 22: JetBrains IDE Plugin

## Overview
Create a JetBrains IDE plugin to integrate murmur-ai functionality into IntelliJ IDEA, PyCharm, WebStorm, and other JetBrains IDEs.

## Motivation
- Developers using JetBrains IDEs should be able to use murmur-ai directly from their IDE
- Consistent experience across different editors (VS Code, Vim, JetBrains)
- Easy access to sync, extract, stats, and patterns commands

## Implementation

### Plugin Structure
```
integrations/jetbrains/
├── build.gradle.kts      # Gradle build configuration (Kotlin DSL)
├── settings.gradle.kts   # Project settings
├── gradle.properties     # Gradle properties
├── src/main/
│   ├── kotlin/com/murmurai/plugin/
│   │   ├── MurmurAction.kt      # Action classes for menu items
│   │   └── MurmurToolWindow.kt  # Optional tool window for output
│   └── resources/
│       └── META-INF/
│           └── plugin.xml       # Plugin descriptor
└── README.md
```

### Actions
All actions use `ProcessBuilder` to invoke the `mur` CLI:

1. **Murmur Sync** — Executes `mur sync all`
2. **Murmur Extract** — Executes `mur learn extract --auto`
3. **Murmur Stats** — Displays statistics from `mur stats`
4. **Murmur Patterns** — Displays patterns from `mur patterns`

### Menu Location
- Tools > Murmur AI > Sync
- Tools > Murmur AI > Extract
- Tools > Murmur AI > Stats
- Tools > Murmur AI > Patterns

### Plugin Metadata
- **ID:** com.murmurai.plugin
- **Name:** Murmur AI
- **Vendor:** karajanchang
- **Compatible IDEs:** IntelliJ IDEA 2024.1+, PyCharm, WebStorm, etc.

## Build & Install
```bash
# Build the plugin
./gradlew buildPlugin

# Output: build/distributions/murmur-ai-*.zip

# Install: Settings > Plugins > ⚙️ > Install Plugin from Disk...
```

## Testing
- Manual testing in IDE sandbox via `./gradlew runIde`
- Verify each action executes the correct CLI command
- Verify output is displayed in notification or tool window

## Future Enhancements
- Real-time output streaming in tool window
- Project-specific configuration
- Status bar widget showing sync status
- Keyboard shortcuts
