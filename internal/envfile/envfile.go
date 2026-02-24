package envfile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// envFilePath returns the path to ~/.mur/.env
func envFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".mur", ".env"), nil
}

// Load reads ~/.mur/.env and sets environment variables.
// It does NOT override existing environment variables.
func Load() error {
	path, err := envFilePath()
	if err != nil {
		return err
	}

	entries, err := parse(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no .env file is fine
		}
		return err
	}

	for key, value := range entries {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	return nil
}

// Get reads a key from ~/.mur/.env without setting env vars.
func Get(key string) string {
	path, err := envFilePath()
	if err != nil {
		return ""
	}

	entries, err := parse(path)
	if err != nil {
		return ""
	}

	return entries[key]
}

// Set writes or updates a key=value in ~/.mur/.env.
// Creates the file with 0600 permissions if it doesn't exist.
func Set(key, value string) error {
	path, err := envFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Read existing content
	var lines []string
	data, err := os.ReadFile(path)
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}

	// Find and replace existing key, or append
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && parts[0] == key {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	if !found {
		// Remove trailing empty lines before appending
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0600)
}

// GenerateToken creates a random token with prefix "mur_" (32 hex chars).
func GenerateToken() string {
	b := make([]byte, 16) // 16 bytes = 32 hex chars
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return "mur_" + hex.EncodeToString(b)
}

// parse reads a .env file and returns key-value pairs.
func parse(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Handle double-quoted values
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		entries[key] = value
	}

	return entries, nil
}
