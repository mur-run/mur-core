# Plan 005: `/mur:in` `/mur:out` Session Recording

> Date: 2026-02-23
> Status: Draft
> Related: [MUR Commander Architecture](https://github.com/mur-run/mur-commander/blob/main/ARCHITECTURE.md)

## Summary

Add `/mur:in` and `/mur:out` custom commands to mur-core, enabling users to mark conversation segments for workflow extraction. When `/mur:out` is triggered, MUR analyzes the recorded segment and generates a structured workflow, presented via an interactive web UI for editing and export as a reusable skill.

## Background

Current mur-core extracts **atomic patterns** from transcripts — useful but fragmented. Real problem-solving knowledge spans multiple conversation turns and involves sequential steps, tool calls, and decision points. Users need a way to capture **complete workflows**.

Discussion thread: Slack #mur 2026-02-23

## User Flow

```
1. User starts AI conversation, begins solving a problem

2. User types: /mur:in
   → MUR hook intercepts, starts recording
   → Visual indicator: "🔴 MUR recording..."

3. User continues conversation normally...
   (multiple turns, tool calls, file edits, commands)

4. User types: /mur:out
   → MUR hook intercepts, stops recording
   → Collects transcript between in~out markers
   → Sends to LLM for QA-CoT analysis
   → Opens interactive web page in browser

5. User edits workflow in web UI:
   - Name, trigger description
   - Variables (name, type, required, default)
   - Steps (reorder, edit, add, remove)
   - Tools used
   - Approval gates

6. User clicks [Save as Skill]
   → MUR exports as skill format
   → Available for future use: mur run <workflow-name>
```

## Implementation

### Phase 1: Command Registration & Hook Integration

**Goal:** `/mur:in` and `/mur:out` work in Claude Code (primary target).

#### 1.1 Session State File

```
~/.mur/session/
├── active.json          # current recording state
└── recordings/
    └── <session-id>.jsonl   # recorded events
```

```go
// internal/session/state.go

type RecordingState struct {
    Active    bool   `json:"active"`
    SessionID string `json:"session_id"`
    StartedAt int64  `json:"started_at"`
    Source    string `json:"source"`    // "claude-code", "codex", etc.
    Marker   string `json:"marker"`    // original /mur:in message context
}
```

#### 1.2 Hook Scripts

Add two new hook scripts installed alongside existing mur hooks:

```bash
# ~/.mur/hooks/mur-session-in.sh
#!/bin/bash
# mur-managed-hook v1
# Triggered by UserPromptSubmit when user types /mur:in

mur session start --source "$MUR_SOURCE"
```

```bash
# ~/.mur/hooks/mur-session-out.sh
#!/bin/bash
# mur-managed-hook v1
# Triggered by UserPromptSubmit when user types /mur:out

mur session stop --analyze
```

#### 1.3 Claude Code Hook Registration

Update `internal/hooks/claudecode.go` to register `/mur:in` and `/mur:out` matchers:

```go
// Add to UserPromptSubmit hooks:
{
    Matcher: "/mur:in",
    Hooks: []ClaudeCodeHook{
        {Type: "command", Command: murBin + " session start --source claude-code"},
    },
},
{
    Matcher: "/mur:out",
    Hooks: []ClaudeCodeHook{
        {Type: "command", Command: murBin + " session stop --analyze --open"},
    },
},
```

#### 1.4 CLI Commands

```go
// cmd/mur/cmd/session.go

var sessionCmd = &cobra.Command{
    Use:   "session",
    Short: "Manage conversation recording sessions",
}

var sessionStartCmd = &cobra.Command{
    Use:   "start",
    Short: "Start recording a conversation segment",
    // Creates active.json, starts recording to JSONL
}

var sessionStopCmd = &cobra.Command{
    Use:   "stop",
    Short: "Stop recording and optionally analyze",
    // Flags: --analyze, --open (open web UI)
}

var sessionStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show current recording status",
}

var sessionListCmd = &cobra.Command{
    Use:   "list",
    Short: "List past recordings",
}
```

### Phase 2: Transcript Capture

**Goal:** Record all conversation events between `/mur:in` and `/mur:out`.

#### 2.1 Event Recording

Leverage existing mur-core hook infrastructure. When recording is active, all hook events (UserPromptSubmit, PostToolUse, Stop) append to the session JSONL:

```go
// internal/session/recorder.go

type EventRecord struct {
    Timestamp int64       `json:"ts"`
    Type      string      `json:"type"`     // "user", "assistant", "tool_call", "tool_result"
    Content   string      `json:"content"`
    Tool      string      `json:"tool,omitempty"`
    Meta      map[string]any `json:"meta,omitempty"`
}

func RecordEvent(sessionID string, event EventRecord) error {
    // Append to ~/.mur/session/recordings/<sessionID>.jsonl
}

func IsRecording() (bool, string) {
    // Check active.json
}
```

#### 2.2 Existing Hook Modification

Update the existing `UserPromptSubmit` and `Stop` hook scripts to check recording state:

```bash
# In existing post-tool-use hook, add:
if mur session status --quiet; then
    mur session record --type tool_result --tool "$TOOL_NAME" --content "$RESULT"
fi
```

### Phase 3: LLM Analysis (QA-CoT)

**Goal:** Analyze recorded transcript and extract structured workflow.

#### 3.1 Analyzer

```go
// internal/session/analyze.go

type AnalysisResult struct {
    Name        string     `json:"name"`
    Trigger     string     `json:"trigger"`
    Description string     `json:"description"`
    Variables   []Variable `json:"variables"`
    Steps       []Step     `json:"steps"`
    Tools       []string   `json:"tools"`
    Tags        []string   `json:"tags"`
}

type Variable struct {
    Name        string `json:"name"`
    Type        string `json:"type"`      // string, path, url, number, bool
    Required    bool   `json:"required"`
    Default     string `json:"default"`
    Description string `json:"description"`
}

type Step struct {
    Order        int    `json:"order"`
    Description  string `json:"description"`
    Command      string `json:"command,omitempty"`
    Tool         string `json:"tool,omitempty"`
    NeedsApproval bool  `json:"needs_approval"`
    OnFailure    string `json:"on_failure"`  // skip, abort, retry
}

func Analyze(sessionPath string, provider LLMProvider) (*AnalysisResult, error) {
    // 1. Read JSONL transcript
    // 2. Build QA-CoT prompt
    // 3. Call LLM
    // 4. Parse structured output
}
```

#### 3.2 QA-CoT Prompt

```go
const qaCoTPrompt = `Analyze this AI conversation transcript and extract a reusable workflow.

Answer each question step by step:

Q1: What was the user's initial problem or goal?
Q2: What was the root cause (if debugging)?
Q3: What steps were attempted? Which succeeded, which failed?
Q4: What is the minimal correct sequence of steps to solve this?
Q5: What tools or commands were used at each step?
Q6: Which values are environment-specific and should be variables?
Q7: Are there conditional branches (if X then Y)?
Q8: Which steps need human approval before proceeding?
Q9: What's a good name and trigger description for this workflow?
Q10: What tags would help find this workflow later?

Then output a JSON object with this structure:
{
  "name": "kebab-case-name",
  "trigger": "when to use this workflow",
  "description": "what this workflow does",
  "variables": [...],
  "steps": [...],
  "tools": [...],
  "tags": [...]
}

TRANSCRIPT:
%s`
```

#### 3.3 LLM Provider

```go
// internal/session/llm.go

type LLMProvider interface {
    Complete(prompt string) (string, error)
}

// Reuse existing mur-core LLM config if available,
// or use environment variables: MUR_LLM_PROVIDER, MUR_API_KEY, MUR_MODEL
```

### Phase 4: Interactive Web UI

**Goal:** Present workflow in editable web page, bidirectional with MUR backend.

#### 4.1 Web Server

Extend existing `internal/server/` or create `internal/session/ui/`:

```go
// internal/session/ui/server.go

func Serve(result *AnalysisResult, port int) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleIndex)           // main editor page
    mux.HandleFunc("/ws", handleWebSocket)     // bidirectional updates
    mux.HandleFunc("/api/save", handleSave)    // save as skill
    mux.HandleFunc("/api/discard", handleDiscard)
    
    // Auto-open browser
    openBrowser(fmt.Sprintf("http://localhost:%d", port))
    
    return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
```

#### 4.2 Page Templates

**MVP: Procedure Editor** (single HTML template with embedded JS)

```
┌─────────────────────────────────────────────┐
│ 📋 MUR Workflow Editor                      │
│                                             │
│ Name:    [fix-nginx-502_______________]     │
│ Trigger: [Nginx returns 502 after deploy__] │
│ Tags:    [nginx] [deploy] [docker] [+]      │
│                                             │
│ ─── Variables ───────────────────────────── │
│ $server_host  [string] [required] [✏️] [🗑️] │
│ $service_name [string] [optional] [✏️] [🗑️] │
│ [+ Add Variable]                            │
│                                             │
│ ─── Steps ───────────────────────────────── │
│ ☰ 1. SSH to $server_host          [✏️] [🗑️] │
│ ☰ 2. Check Docker logs            [✏️] [🗑️] │
│ ☰ 3. Fix nginx.conf               [✏️] [🗑️] │
│ ☰ 4. 🔒 Restart containers (approval) [✏️]  │
│ ☰ 5. Verify health check          [✏️] [🗑️] │
│ [+ Add Step]                                │
│                                             │
│ ─── Tools Used ──────────────────────────── │
│ [shell] [file-edit] [docker]                │
│                                             │
│ [💾 Save as Skill] [📄 Export YAML] [🗑️ Discard] │
│                                             │
│ ─── AI Assistant ────────────────────────── │
│ 💡 Suggestion: Add a config backup step     │
│    before modifying nginx.conf              │
│    [Accept] [Dismiss]                       │
└─────────────────────────────────────────────┘
```

#### 4.3 WebSocket Protocol

```json
// Client → Server
{ "type": "step.reorder", "from": 2, "to": 0 }
{ "type": "step.update", "index": 1, "field": "description", "value": "..." }
{ "type": "step.delete", "index": 3 }
{ "type": "step.add", "after": 2 }
{ "type": "variable.add", "name": "port", "var_type": "number" }
{ "type": "workflow.update", "field": "name", "value": "fix-nginx-502" }
{ "type": "ai.review", "step_index": 2 }

// Server → Client
{ "type": "state.full", "workflow": { ... } }
{ "type": "ai.suggestion", "message": "...", "action": { ... } }
{ "type": "save.success", "path": "~/.mur/skills/fix-nginx-502/" }
```

### Phase 5: Skill Export

**Goal:** Save confirmed workflow as a reusable mur skill.

#### 5.1 Export Formats

```go
// internal/session/export.go

func ExportAsSkill(result *AnalysisResult, outputDir string) error {
    // Creates:
    // <outputDir>/<name>/
    //   ├── SKILL.md        # OpenClaw-compatible skill
    //   ├── workflow.yaml    # Structured workflow definition
    //   ├── run.sh           # Entry point script
    //   └── steps/
    //       ├── 01-<step>.sh
    //       └── 02-<step>.sh
}

func ExportAsYAML(result *AnalysisResult, path string) error { ... }
func ExportAsMarkdown(result *AnalysisResult, path string) error { ... }
```

#### 5.2 Workflow YAML Format

```yaml
# ~/.mur/skills/fix-nginx-502/workflow.yaml
kind: workflow
version: "1"
name: fix-nginx-502
trigger: "Nginx returns 502 error after deployment"
description: "Diagnose and fix Nginx 502 errors in Docker environments"

variables:
  - name: server_host
    type: string
    required: true
    description: "SSH host of the target server"
  - name: service_name
    type: string
    required: false
    default: "web-app"
    description: "Docker service name"

steps:
  - description: "SSH to server"
    command: "ssh $server_host"
  - description: "Check Docker container logs"
    command: "docker compose logs $service_name --tail 100"
  - description: "Identify config issue in nginx.conf"
    tool: file-edit
  - description: "Restart containers"
    command: "docker compose restart"
    needs_approval: true
  - description: "Verify health check"
    command: "curl -f http://localhost/health"
    on_failure: retry

tools: [shell, file-edit, docker]
tags: [nginx, deploy, docker, devops]

source:
  recorded_at: "2026-02-23T15:30:00+08:00"
  session_id: "abc123"
```

## Implementation Roadmap

### Milestone 1: Recording (1 week)
- [ ] `internal/session/state.go` — recording state management
- [ ] `internal/session/recorder.go` — event JSONL recording
- [ ] `cmd/mur/cmd/session.go` — `mur session start/stop/status/list`
- [ ] Update `internal/hooks/claudecode.go` — register `/mur:in` `/mur:out` matchers
- [ ] Create hook scripts `mur-session-in.sh`, `mur-session-out.sh`
- [ ] Update existing hooks to check recording state and append events
- [ ] Tests for session lifecycle

### Milestone 2: Analysis (1 week)
- [ ] `internal/session/llm.go` — LLM provider interface
- [ ] `internal/session/analyze.go` — QA-CoT prompt + response parsing
- [ ] `internal/session/analyze_test.go` — tests with fixture transcripts
- [ ] Config: `mur config set llm.provider anthropic`
- [ ] `mur session analyze <session-id>` CLI command

### Milestone 3: Web UI (2 weeks)
- [ ] `internal/session/ui/server.go` — HTTP + WebSocket server
- [ ] `internal/session/ui/templates/` — HTML template (embedded via `embed`)
- [ ] Procedure Editor — step list with drag-and-drop
- [ ] Variable editor — add/edit/delete
- [ ] WebSocket bidirectional protocol
- [ ] AI review button (send step to LLM for suggestions)
- [ ] `mur session ui <session-id>` CLI command
- [ ] Auto-open browser on `/mur:out --analyze --open`

### Milestone 4: Export (1 week)
- [ ] `internal/session/export.go` — skill/YAML/markdown exporters
- [ ] Save button in web UI → export to `~/.mur/skills/`
- [ ] `mur session export <session-id> --format skill`
- [ ] Integration: exported skills work with `mur run <name>`

### Milestone 5: Polish (1 week)
- [ ] Error handling and edge cases
- [ ] Recording indicator feedback to user
- [ ] Support Codex/Gemini hooks (beyond Claude Code)
- [ ] Documentation
- [ ] `mur sync` includes exported workflow skills

## File Changes Summary

```
mur-core/
├── cmd/mur/cmd/
│   └── session.go              # NEW — session subcommands
├── internal/
│   ├── session/                # NEW — entire package
│   │   ├── state.go            # recording state
│   │   ├── recorder.go         # event recording
│   │   ├── analyze.go          # LLM analysis
│   │   ├── llm.go              # LLM provider
│   │   ├── export.go           # skill export
│   │   └── ui/
│   │       ├── server.go       # web server
│   │       ├── ws.go           # websocket handler
│   │       ├── templates/      # HTML templates (embed)
│   │       └── static/         # CSS/JS assets (embed)
│   └── hooks/
│       └── claudecode.go       # MODIFIED — add /mur:in /mur:out matchers
└── plans/
    └── 005-mur-in-out-session-recording.md  # this file
```

## Open Questions

1. **Recording across sessions** — if user closes Claude Code and reopens, should recording persist? (Suggest: yes, active.json survives)
2. **Max recording length** — cap at N tokens? (Suggest: configurable, default 100k tokens)
3. **Multiple recordings** — allow overlapping `/mur:in` without `/mur:out`? (Suggest: no, `/mur:in` while recording = restart)
4. **Web UI port conflict** — what if port 3939 is taken? (Suggest: auto-find available port)
5. **LLM cost** — analysis uses ~10k-50k input tokens. Show estimated cost before analyzing? (Suggest: yes for paid APIs)

## Relationship to MUR Commander

This plan implements the **manual mode** (`/mur:in` + `/mur:out`) in mur-core. MUR Commander (Rust) will later add **automatic mode** (completion detection without manual markers). The data formats (session JSONL, workflow YAML) are designed to be compatible — Commander can read mur-core recordings and vice versa.
