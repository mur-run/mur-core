# Murmur AI - JetBrains Plugin

Integrate Murmur AI into IntelliJ IDEA, PyCharm, WebStorm, and other JetBrains IDEs.

## Features

| Action | Command | Description |
|--------|---------|-------------|
| **Sync** | `mur sync all` | Sync all learning data |
| **Extract** | `mur learn extract --auto` | Extract patterns from code |
| **Stats** | `mur stats` | Show statistics |
| **Patterns** | `mur patterns` | List learned patterns |

Access via: **Tools > Murmur AI > ...**

## Prerequisites

The `mur` CLI must be installed and available in your PATH:

```bash
# Using Homebrew
brew install karajanchang/tap/mur

# Or using Go
go install github.com/karajanchang/murmur-ai/cmd/mur@latest
```

Verify installation:
```bash
mur --version
```

## Building the Plugin

### Requirements
- JDK 17+
- Gradle 8.x (wrapper included)

### Build

```bash
# Build the plugin
./gradlew buildPlugin

# Output: build/distributions/murmur-ai-jetbrains-0.1.0.zip
```

### Development

```bash
# Run IDE with plugin for testing
./gradlew runIde

# Clean build artifacts
./gradlew clean
```

## Installation

### From Disk (Development)

1. Build the plugin: `./gradlew buildPlugin`
2. Open your JetBrains IDE
3. Go to **Settings** (⌘, on macOS)
4. Navigate to **Plugins**
5. Click the ⚙️ gear icon
6. Select **Install Plugin from Disk...**
7. Choose `build/distributions/murmur-ai-jetbrains-0.1.0.zip`
8. Restart the IDE

### From JetBrains Marketplace (Future)

*Coming soon - the plugin will be published to the JetBrains Marketplace.*

## Usage

1. Open any project in your JetBrains IDE
2. Go to **Tools > Murmur AI**
3. Select an action:
   - **Sync** - Synchronize learning data
   - **Extract** - Extract patterns from your codebase
   - **Stats** - View learning statistics
   - **Patterns** - Browse learned patterns

Results appear as notifications and in the **Murmur AI** tool window (bottom panel).

## Project Structure

```
integrations/jetbrains/
├── build.gradle.kts              # Gradle build configuration
├── settings.gradle.kts           # Project settings
├── gradle.properties             # Build properties
├── src/main/
│   ├── kotlin/com/murmurai/plugin/
│   │   ├── MurmurAction.kt       # Action implementations
│   │   └── MurmurToolWindow.kt   # Tool window UI
│   └── resources/META-INF/
│       └── plugin.xml            # Plugin descriptor
└── README.md
```

## Compatibility

- **IDE Version:** 2024.1+
- **Supported IDEs:**
  - IntelliJ IDEA (Community & Ultimate)
  - PyCharm (Community & Professional)
  - WebStorm
  - GoLand
  - CLion
  - Rider
  - RubyMine
  - PhpStorm
  - DataGrip

## Troubleshooting

### "mur: command not found"

The `mur` CLI is not in your PATH. Either:
- Install via Homebrew: `brew install karajanchang/tap/mur`
- Add the installation directory to your PATH
- Restart your IDE after installation

### Plugin not showing in Tools menu

1. Ensure the plugin is enabled in Settings > Plugins
2. Restart the IDE
3. Check the IDE logs for errors

### Commands timeout

Commands have a 60-second timeout. For large codebases, consider:
- Running commands manually in terminal
- Breaking extraction into smaller chunks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with `./gradlew runIde`
5. Submit a pull request

## License

MIT License - see the main [murmur-ai repository](https://github.com/karajanchang/murmur-ai) for details.
