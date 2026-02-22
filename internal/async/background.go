// Package async provides cross-platform background process execution.
package async

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunBackground re-executes the given mur subcommand as a detached background
// process. The --async flag is stripped so the child runs normally. The parent
// returns immediately after spawning.
//
// This works cross-platform: on Unix it sets Setsid, on Windows the Go runtime
// handles process detachment via CREATE_NEW_PROCESS_GROUP automatically when
// we don't call cmd.Wait().
func RunBackground(args []string) error {
	// Find our own binary
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find mur binary: %w", err)
	}

	// Strip --async from args
	var cleanArgs []string
	for _, a := range args {
		if a == "--async" {
			continue
		}
		// Handle --async=true / --async=false
		if strings.HasPrefix(a, "--async=") {
			continue
		}
		cleanArgs = append(cleanArgs, a)
	}

	cmd := exec.Command(self, cleanArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Platform-specific detach is in background_unix.go / background_windows.go
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	// Release the process so parent can exit without waiting
	_ = cmd.Process.Release()
	return nil
}
