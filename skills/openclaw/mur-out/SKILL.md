---
name: mur-out
description: Stop recording and extract learned patterns from the captured mur session
---

# mur-out

Stop the active mur session recording and trigger pattern extraction and sync.

## Instructions

When the user invokes this skill:

1. Read `~/.mur/openclaw-session.json` to get the session state
   - If the file doesn't exist, inform the user that no active session was found and suggest running `/mur-in` first
2. Calculate the session duration from `start_time` to now
3. Run pattern extraction:
   ```bash
   mur learn extract --llm --auto --accept-all
   ```
   This analyzes recent activity and extracts reusable patterns using the configured LLM.
4. Run sync to push patterns to all configured CLI tools:
   ```bash
   mur sync --quiet
   ```
5. Clean up by deleting `~/.mur/openclaw-session.json`
6. Report results to the user:
   - Session duration
   - Number of patterns extracted (parse from mur learn extract output)
   - Confirmation that patterns were synced to CLI tools
   - The project and goal from the session (if provided)

If either command fails, report the error but continue with the remaining steps. Always clean up the state file.

**Note:** In the future, this will also upload session data to mur.run and provide a workflow design URL. For now, only local analysis is performed.
