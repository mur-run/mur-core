# Cloud Sync

Sync patterns across devices and share with your team via [mur.run](https://mur.run).

## Quick Start

```bash
# Login (opens browser for OAuth)
mur login

# Smart sync - auto-detects your plan
mur sync                    # Trial/Pro/Team → cloud, Free → git

# Manual control
mur sync --cloud            # Force cloud sync
mur sync --git              # Force git sync
mur sync --cli              # Only sync to local AI tools
```

## Plans

| Plan | Price | Features |
|------|-------|----------|
| **Free** | $0/mo | Local patterns, git sync, all AI tools |
| **Trial** | $0 (3 mo) | Cloud sync trial for new users |
| **Pro** | $9/mo | Cloud sync, 3 devices |
| **Team** | $49/mo | 5 members, shared patterns, analytics |

Learn more at [mur.run](https://mur.run)

## Push & Pull

```bash
# Push local patterns to server
mur cloud push

# Pull patterns from server
mur cloud pull
mur cloud pull --force      # Overwrite local

# Full bidirectional sync
mur cloud sync
```

## Team Sync

```bash
# List your teams
mur cloud teams

# Select active team
mur cloud select <team-slug>

# Sync with team
mur sync
```

## Conflict Resolution

When conflicts occur, MUR helps you resolve them interactively:

```
⚠️  3 conflict(s) detected

? How do you want to resolve conflicts?
  › [i] Interactive - choose for each pattern
    [s] Accept all from server
    [l] Accept all from local
    [x] Skip all (no changes)
```

For each pattern:
- **[s]** Keep server version
- **[l]** Keep local version
- **[d]** View diff
- **[x]** Skip

## Auto-Sync

Set up automatic background sync:

```bash
# Enable (interactive)
mur sync auto enable

? How often should mur sync?
  › Every 15 minutes
    Every 30 minutes
    Every hour

# Check status
mur sync auto status

# Disable
mur sync auto disable
```

**Platform support:**
| Platform | Implementation |
|----------|---------------|
| macOS | LaunchAgent |
| Linux | systemd user timer |
| Windows | Task Scheduler |

## API Key Authentication

For CI/automation:

```bash
# Create API key at: https://app.mur.run/core/settings
mur login --api-key mur_xxx_...

# Check authentication
mur whoami

# Logout
mur logout
```

## Git Sync (Free Alternative)

For free users without cloud:

```yaml
# ~/.mur/config.yaml
learning:
  repo: git@github.com:you/mur-patterns.git
  auto_push: true
```

```bash
mur repo init              # Clone/setup repo
mur sync --git             # Sync with git
```
