// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

// OpenClaw hook files content
const openclawHookMD = `---
name: mur-patterns
description: "Injects relevant MUR patterns into agent context at session start"
metadata: { "openclaw": { "emoji": "🧠", "events": ["agent:bootstrap"] } }
---

# MUR Patterns Hook

Automatically injects relevant learned patterns from mur-core into the agent context.

## What It Does

1. Listens for agent bootstrap events
2. Searches local patterns for relevant context
3. Injects matching patterns into the session

## Requirements

- mur-core CLI installed (` + "`go install github.com/mur-run/mur-core/cmd/mur@latest`" + `)

## Configuration

No configuration needed. Patterns are loaded from ~/.mur/patterns/
`

const openclawHandlerTS = `import type { HookHandler } from "openclaw/hooks";

const handler: HookHandler = async (event) => {
  // Only handle agent:bootstrap events
  if (event.type !== "agent" || event.action !== "bootstrap") {
    return;
  }

  try {
    // Get the bootstrap files from context
    const bootstrapFiles = event.context?.bootstrapFiles;
    if (!bootstrapFiles) return;

    // Run mur search to get relevant patterns
    const { execSync } = await import("child_process");
    
    // Extract some context from existing bootstrap files to search for patterns
    let searchQuery = "";
    for (const file of bootstrapFiles) {
      if (file.name === "AGENTS.md" || file.name === "SOUL.md") {
        // Use first 200 chars as search context
        searchQuery += file.content?.slice(0, 200) || "";
      }
    }

    if (!searchQuery) {
      searchQuery = "general development patterns";
    }

    // Search for relevant patterns (limit to 3)
    const result = execSync(
      ` + "`mur search --limit 3 --format md \"${searchQuery.slice(0, 100)}\"`" + `,
      { encoding: "utf-8", timeout: 5000 }
    ).trim();

    if (result && result.length > 10) {
      // Inject patterns into bootstrap
      bootstrapFiles.push({
        name: "PATTERNS.md",
        content: ` + "`# Learned Patterns\\n\\nRelevant patterns from your learning history:\\n\\n${result}`" + `,
        priority: 50,
      });

      console.log("[mur-patterns] Injected patterns into session");
    }
  } catch (err) {
    // Silently fail - don't break the session if mur isn't installed
    console.error("[mur-patterns] Failed to load patterns:", err instanceof Error ? err.message : String(err));
  }
};

export default handler;
`

// InstallOpenClawHooks installs mur hooks for OpenClaw.
func InstallOpenClawHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// OpenClaw hooks directory
	hookDir := filepath.Join(home, ".openclaw", "hooks", "mur-patterns")

	// Create hook directory
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("cannot create hook directory: %w", err)
	}

	// Write HOOK.md
	hookMDPath := filepath.Join(hookDir, "HOOK.md")
	if err := os.WriteFile(hookMDPath, []byte(openclawHookMD), 0644); err != nil {
		return fmt.Errorf("cannot write HOOK.md: %w", err)
	}

	// Write handler.ts
	handlerPath := filepath.Join(hookDir, "handler.ts")
	if err := os.WriteFile(handlerPath, []byte(openclawHandlerTS), 0644); err != nil {
		return fmt.Errorf("cannot write handler.ts: %w", err)
	}

	fmt.Printf("✓ Installed OpenClaw hook at %s\n", hookDir)
	fmt.Println("  Enable with: openclaw hooks enable mur-patterns")

	return nil
}
