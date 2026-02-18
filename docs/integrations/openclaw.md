# OpenClaw Integration

OpenClaw is a powerful AI assistant platform with multi-modal capabilities.

## Overview

| Property | Value |
|----------|-------|
| **Tool Name** | `openclaw` |
| **Binary** | `openclaw` |
| **Tier** | Paid |
| **Capabilities** | coding, analysis, complex, multi-modal |
| **Config Path** | `~/.openclaw/config.yaml` |

## Status

🔜 **Coming Soon** - Integration is in development.

Planned features:

- MCP server sync
- Pattern sync
- Browser automation capabilities
- Multi-modal support (images, screenshots)
- Node control (mobile devices, remote machines)

## Why OpenClaw?

OpenClaw offers unique capabilities:

### Multi-Modal

- Analyze screenshots and images
- Browser automation
- Visual debugging

### Node Network

- Control mobile devices
- Remote machine access
- Cross-device workflows

### Advanced Integrations

- Slack/Discord messaging
- Email automation
- Calendar integration

## Planned Configuration

```yaml
# ~/.mur/config.yaml
tools:
  openclaw:
    enabled: false  # Enable when ready
    binary: openclaw
    tier: paid
    capabilities: [coding, analysis, complex, multi-modal, browser, nodes]
    flags: []
```

## When to Use OpenClaw

OpenClaw will be selected for:

- Tasks requiring browser interaction
- Multi-modal analysis (images, screenshots)
- Cross-device automation
- Complex workflows with external integrations

## Routing Strategy

Once integrated, murmur will route to OpenClaw for:

```
Complexity ≥ 0.7 + multi-modal keywords → OpenClaw
Browser/automation keywords → OpenClaw
Node/device keywords → OpenClaw
```

## Roadmap

1. 🔜 Basic tool definition
2. 🔜 MCP sync support
3. 🔜 Pattern sync
4. 🔜 Multi-modal routing keywords
5. 🔜 Browser automation detection
6. 🔜 Node capabilities integration

## Direct Usage (Coming)

```bash
openclaw "analyze this screenshot and fix the UI bug"
```

Through murmur:

```bash
mur run -t openclaw -p "take a screenshot and explain the layout"
```

## See Also

- [Smart Routing](../concepts/routing.md)
- [OpenClaw Documentation](https://docs.openclaw.ai)
