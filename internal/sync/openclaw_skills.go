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
3. Run pattern extraction:
   ` + "```bash" + `
   mur learn extract --llm --auto --accept-all
   ` + "```" + `
   This analyzes recent activity and extracts reusable patterns using the configured LLM.
4. Run sync to push patterns to all configured CLI tools:
   ` + "```bash" + `
   mur sync --quiet
   ` + "```" + `
5. Upload session data to get a shareable workflow URL:
   ` + "```bash" + `
   mur workflows upload --file ~/.mur/last-session.json
   ` + "```" + `
   - If the upload succeeds, capture the URL from stdout
   - If the upload fails (no internet, API down), skip gracefully and continue
6. Clean up by deleting ` + "`~/.mur/openclaw-session.json`" + `
7. Report results to the user:
   - Session duration
   - Number of patterns extracted (parse from mur learn extract output)
   - Confirmation that patterns were synced to CLI tools
   - The project and goal from the session (if provided)
   - If upload succeeded, show the workflow URL: "Open Workflow: <url>"

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
