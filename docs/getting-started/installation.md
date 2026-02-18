# Installation

## Requirements

- Go 1.21+ (for building from source)
- At least one AI CLI tool installed:
    - [Claude Code](https://docs.anthropic.com/claude-code)
    - [Gemini CLI](https://github.com/google-gemini/gemini-cli)
    - [Auggie](https://github.com/augmentcode/auggie)

## Install Methods

### Go Install (Recommended)

The simplest way if you have Go installed:

```bash
go install github.com/mur-run/mur-core/cmd/mur@latest
```

### Download from Releases

=== "macOS (Apple Silicon)"

    ```bash
    curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur-darwin-arm64 -o mur
    chmod +x mur && sudo mv mur /usr/local/bin/
    ```

=== "macOS (Intel)"

    ```bash
    curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur-darwin-amd64 -o mur
    chmod +x mur && sudo mv mur /usr/local/bin/
    ```

=== "Linux (amd64)"

    ```bash
    curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur-linux-amd64 -o mur
    chmod +x mur && sudo mv mur /usr/local/bin/
    ```

=== "Linux (arm64)"

    ```bash
    curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur-linux-arm64 -o mur
    chmod +x mur && sudo mv mur /usr/local/bin/
    ```

=== "Windows"

    Download `mur-windows-amd64.exe` from the [releases page](https://github.com/mur-run/mur-core/releases/latest) and add to your PATH.

### Build from Source

```bash
git clone https://github.com/mur-run/mur-core.git
cd murmur-ai
go build -o mur ./cmd/mur
sudo mv mur /usr/local/bin/
```

## Verify Installation

```bash
mur --version
```

You should see:

```
mur version 0.1.0
```

## Next Steps

1. [Initialize murmur](quick-start.md) with `mur init`
2. Check which AI tools are available with `mur health`
3. Start using `mur run` for smart routing!
