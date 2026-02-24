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
3. Run pattern extraction using the session's start_time to scope to only this session's activity:
   ```bash
   mur learn extract --llm --auto --accept-all --since "<start_time from session state>"
   ```
   Replace `<start_time from session state>` with the actual `start_time` value read from the session JSON in step 1.
   This analyzes only the activity within the recording window and extracts reusable patterns using the configured LLM.
4. Run sync to push patterns to all configured CLI tools:
   ```bash
   mur sync --quiet
   ```
5. Upload session data to get a shareable workflow URL:
   ```bash
   mur workflows upload --file ~/.mur/last-session.json
   ```
   - If the upload succeeds, capture the URL from stdout and save it as `workflow_url`
   - If the upload fails (network error, timeout, API down), **do not treat this as a fatal error**. Instead:
     - Show: "Upload skipped (offline/unavailable). Results saved locally."
     - Set `workflow_url` to `"(local only)"`
     - Continue with the remaining steps
6. Save the session to history for future reference:
   ```bash
   echo '{"id":"<session_id>","start_time":"<start_time>","end_time":"<now in ISO 8601>","project":"<project>","goal":"<goal>","patterns_extracted":<count>,"workflow_url":"<url or (local only)>","tool":"openclaw"}' | mur sessions save
   ```
   Replace placeholders with actual values from the session state and extraction results.
7. Clean up by deleting `~/.mur/openclaw-session.json`
8. Report results to the user:
   - Session duration
   - Number of patterns extracted (parse from mur learn extract output)
   - Confirmation that patterns were synced to CLI tools
   - The project and goal from the session (if provided)
   - If upload succeeded, show the workflow URL: "Open Workflow: <url>"
   - If upload was skipped, show: "Upload skipped (offline/unavailable). Results saved locally."

If any command fails, report the error but continue with the remaining steps. Always clean up the state file.
