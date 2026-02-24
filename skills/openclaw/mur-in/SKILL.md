---
name: mur-in
description: Start a mur session recording to capture your workflow for pattern extraction
---

# mur-in

Start recording a mur session. This captures the current timestamp and optional context so that when you later run `/mur-out`, mur can extract learned patterns from your activity.

## Instructions

When the user invokes this skill:

1. Generate a unique session ID (use a UUID v4 or timestamp-based ID)
2. Record the current timestamp in ISO 8601 format
3. Ask the user (optionally) for:
   - **project**: The project name or directory they're working in (default: current directory name)
   - **goal**: A brief description of what they plan to accomplish
4. Write the session state to `~/.mur/openclaw-session.json` with this format:

```json
{
  "start_time": "2024-01-01T12:00:00Z",
  "project": "my-project",
  "goal": "implement feature X",
  "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

5. Confirm to the user that recording has started:
   - Show the session ID (abbreviated)
   - Show the start time
   - Remind them to run `/mur-out` when they're done to extract patterns

If a session is already active (the state file exists), warn the user and ask if they want to start a new session (overwriting the old one).

## Auto-trigger (recommended)

This skill can be triggered automatically by the OpenClaw agent. When the agent detects the user is starting a new development task (e.g., "lets build X", "fix the bug in Y", "implement feature Z"), the agent should proactively start a mur session without being asked.

**Trigger signals:**
- User describes a new task, feature, or bug fix
- User switches to a different project
- User says "lets start working on..." or similar
- A significant time gap since last activity (>2 hours)

**Do NOT auto-trigger when:**
- A session is already active (check ~/.mur/openclaw-session.json)
- User is just asking questions or chatting
- User is doing quick one-off tasks

When auto-triggering, briefly inform the user: "Started mur session recording for this task."
