// Package sysinfo provides system detection utilities (RAM, services).
package sysinfo

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemRAMGB returns total system RAM in GB, or 0 if detection fails.
func SystemRAMGB() int {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err != nil {
			return 0
		}
		bytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return 0
		}
		return int(bytes / (1024 * 1024 * 1024))
	case "linux":
		out, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
		if err != nil {
			return 0
		}
		var kb uint64
		if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "MemTotal: %d kB", &kb); err != nil {
			// Try space-separated fallback
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				kb, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
		if kb > 0 {
			return int(kb / (1024 * 1024))
		}
		return 0
	default:
		return 0
	}
}

// OllamaRunning checks if Ollama is reachable at the given URL.
// If url is empty, defaults to http://localhost:11434.
func OllamaRunning(url string) bool {
	if url == "" {
		url = "http://localhost:11434"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
