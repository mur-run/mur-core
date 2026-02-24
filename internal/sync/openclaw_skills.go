package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

const openclawMurInSkill = `---
name: mur-in
description: Start a mur session recording to capture your workflow for pattern extraction
---

# mur-in

Start recording a mur session. This captures the current timestamp and optional context so that when you later run ` + "`/mur-out`" + `, mur can extract learned patterns from your activity.

## Instructions

When the user invokes this skill:

1. Generate a unique session ID (use a UUID v4 or timestamp-based ID)
2. Record the current timestamp in ISO 8601 format
3. Ask the user (optionally) for:
   - **project**: The project name or directory they're working in (default: current directory name)
   - **goal**: A brief description of what they plan to accomplish
4. Write the session state to ` + "`~/.mur/openclaw-session.json`" + ` with this format:

` + "```json" + `
{
  "start_time": "2024-01-01T12:00:00Z",
  "project": "my-project",
  "goal": "implement feature X",
  "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
` + "```" + `

5. Confirm to the user that recording has started:
   - Show the session ID (abbreviated)
   - Show the start time
   - Remind them to run ` + "`/mur-out`" + ` when they're done to extract patterns

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
`

const openclawMurOutSkill = `---
name: mur-out
description: Stop recording and extract learned patterns from the captured mur session
---

# mur-out

Stop the active mur session recording and trigger pattern extraction and sync.

## Instructions

When the user invokes this skill:

1. Read ` + "`~/.mur/openclaw-session.json`" + ` to get the session state
   - If the file doesn't exist, inform the user that no active session was found and suggest running ` + "`/mur-in`" + ` first
2. Calculate the session duration from ` + "`start_time`" + ` to now
3. Run pattern extraction using the session's start_time to scope to only this session's activity:
   ` + "```bash" + `
   mur learn extract --llm --auto --accept-all --since "<start_time from session state>"
   ` + "```" + `
   Replace ` + "`<start_time from session state>`" + ` with the actual ` + "`start_time`" + ` value read from the session JSON in step 1.
   This analyzes only the activity within the recording window and extracts reusable patterns using the configured LLM.
4. Run sync to push patterns to all configured CLI tools:
   ` + "```bash" + `
   mur sync --quiet
   ` + "```" + `
5. Upload session data to get a shareable workflow URL:
   ` + "```bash" + `
   mur workflows upload --file ~/.mur/last-session.json
   ` + "```" + `
   - If the upload succeeds, capture the URL from stdout and save it as ` + "`workflow_url`" + `
   - If the upload fails (network error, timeout, API down), **do not treat this as a fatal error**. Instead:
     - Show: "Upload skipped (offline/unavailable). Results saved locally."
     - Set ` + "`workflow_url`" + ` to ` + "`\"(local only)\"`" + `
     - Continue with the remaining steps
6. Save the session to history for future reference:
   ` + "```bash" + `
   echo '{"id":"<session_id>","start_time":"<start_time>","end_time":"<now in ISO 8601>","project":"<project>","goal":"<goal>","patterns_extracted":<count>,"workflow_url":"<url or (local only)>","tool":"openclaw"}' | mur sessions save
   ` + "```" + `
   Replace placeholders with actual values from the session state and extraction results.
7. Clean up by deleting ` + "`~/.mur/openclaw-session.json`" + `
8. Report results to the user:
   - Session duration
   - Number of patterns extracted (parse from mur learn extract output)
   - Confirmation that patterns were synced to CLI tools
   - The project and goal from the session (if provided)
   - If upload succeeded, show the workflow URL: "Open Workflow: <url>"
   - If upload was skipped, show: "Upload skipped (offline/unavailable). Results saved locally."

If any command fails, report the error but continue with the remaining steps. Always clean up the state file.
`

// EnsureOpenClawSkills writes the built-in mur-in and mur-out skills to
// ~/.mur/skills/ so they can be synced to OpenClaw's ~/.agents/skills/ directory.
func EnsureOpenClawSkills() error {
	skillsDir, err := SkillsSourceDir()
	if err != nil {
		return err
	}

	skills := []struct {
		dir     string
		content string
	}{
		{dir: "mur-in", content: openclawMurInSkill},
		{dir: "mur-out", content: openclawMurOutSkill},
	}

	for _, s := range skills {
		skillDir := filepath.Join(skillsDir, s.dir)
		skillPath := filepath.Join(skillDir, "SKILL.md")

		// Check if already up to date
		if existing, err := os.ReadFile(skillPath); err == nil {
			if string(existing) == s.content {
				continue // Already up to date
			}
		}

		// Create directory and write skill file
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("cannot create skill directory %s: %w", skillDir, err)
		}

		if err := os.WriteFile(skillPath, []byte(s.content), 0644); err != nil {
			return fmt.Errorf("cannot write skill %s: %w", skillPath, err)
		}
	}

	return nil
}
