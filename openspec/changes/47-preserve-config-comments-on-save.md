# CS-47: Preserve Config Comments on Save

## Problem

`Config.Save()` uses `yaml.Marshal()` which strips all YAML comments. When `mur init` writes a template with rich annotations (section dividers, cost estimates, provider examples, premium config), any subsequent `Load()+Save()` cycle destroys them.

Introduced by `bc52261` which changed init's community sharing logic to use typed `config.Load()` + `Save()` instead of string operations.

## Solution

Use `gopkg.in/yaml.v3` Node-based marshaling in `Save()` to preserve existing comments, plus add a regression test.

## Changes

### 1. `internal/config/config.go` — Comment-preserving Save

- `Save()` reads existing file as `yaml.Node` tree
- Marshals current `Config` struct into a fresh `yaml.Node` tree  
- Merges values from fresh tree into existing tree (preserving comment nodes)
- Falls back to plain `yaml.Marshal` if no existing file
- Add helper `mergeNodes(dst, src *yaml.Node)` for recursive node merge

### 2. `internal/config/config_test.go` — Regression test

- `TestSavePreservesComments`: write a config with comments, Load+Save, assert comments survive
- `TestSaveCreatesNewFile`: Save without existing file still works

### 3. `cmd/mur/cmd/init.go` — Remove workarounds

- Remove any string-based hacks that existed to work around Save() behavior
- Verify init end-to-end: template comments survive the full init flow

## Acceptance Criteria

- [ ] `mur init` produces config with all section dividers (━━━), cost notes (💡), provider lists
- [ ] After `config.Load()` + `cfg.Save()`, comments are preserved
- [ ] `TestSavePreservesComments` passes
- [ ] `go test ./...` all pass
- [ ] No change in config values/behavior — only comment preservation

## Risk

Low — yaml.v3 Node API is stable; fallback to plain marshal if no existing file.
