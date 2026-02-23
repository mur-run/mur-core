---
description: "Stop recording and analyze the captured conversation to extract a reusable workflow."
allowed-tools: Bash(mur session stop:*), Bash(mur session list:*), Bash(mur session export:*)
---

Stop the active mur session recording by running:

```bash
mur session stop
```

After stopping, show the user:
- Session duration
- Number of events captured
- Ask if they want to analyze the session: `mur session stop --analyze`
- Or export it: `mur session export <session-id> --format skill`

If `--analyze` fails due to missing API key, tell the user to set `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`.
