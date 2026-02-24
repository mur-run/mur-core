---
description: "Start recording this conversation for workflow extraction. Events (prompts, tool calls, stops) will be captured until you run /mur:out."
allowed-tools: Bash(mur session start:*)
---

Start a mur session recording by running:

```bash
mur session start --source claude-code
```

After starting, confirm to the user:
- The session ID
- That recording is active
- Remind them to use `/mur:out` when done
