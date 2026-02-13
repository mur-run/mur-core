package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or edit mur configuration",
	Long: `View or edit mur configuration.

Examples:
  mur config              # Show current config
  mur config edit         # Edit in $EDITOR
  mur config path         # Show config file path
  mur config get <key>    # Get a specific value
  mur config set <k> <v>  # Set a value`,
	RunE: runConfigShow,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit config in $EDITOR",
	RunE:  runConfigEdit,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE:  runConfigPath,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset config to defaults",
	RunE:  runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mur", "config.yaml"), nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No config file found.")
			fmt.Println("Run 'mur init' to create one.")
			return nil
		}
		return err
	}

	fmt.Printf("# %s\n", path)
	fmt.Println()
	fmt.Print(string(content))

	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := createDefaultConfig(path); err != nil {
			return err
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR")
	}

	editorCmd := exec.Command(editor, path)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	path, err := configPath()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no config file found")
		}
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Simple dot notation support
	value := getNestedValue(config, key)
	if value == nil {
		return fmt.Errorf("key not found: %s", key)
	}

	// Pretty print
	out, err := yaml.Marshal(value)
	if err != nil {
		fmt.Println(value)
	} else {
		fmt.Print(string(out))
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	path, err := configPath()
	if err != nil {
		return err
	}

	// Read existing config
	var config map[string]interface{}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			config = make(map[string]interface{})
		} else {
			return err
		}
	} else {
		if err := yaml.Unmarshal(content, &config); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	// Set value
	setNestedValue(config, key, value)

	// Write back
	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return err
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Backup existing
	if _, err := os.Stat(path); err == nil {
		backup := path + ".backup"
		if err := os.Rename(path, backup); err != nil {
			return fmt.Errorf("failed to backup: %w", err)
		}
		fmt.Printf("Backed up to: %s\n", backup)
	}

	if err := createDefaultConfig(path); err != nil {
		return err
	}

	fmt.Println("Config reset to defaults.")
	return nil
}

func createDefaultConfig(path string) error {
	defaultConfig := `# mur configuration
# See: https://github.com/mur-run/mur-core

default_tool: claude

tools:
  claude:
    enabled: true
    binary: claude
    tier: paid
  gemini:
    enabled: true
    binary: gemini
    tier: free
  codex:
    enabled: false
    binary: codex
    tier: paid
  auggie:
    enabled: false
    binary: auggie
    tier: free
  aider:
    enabled: false
    binary: aider
    tier: free

routing:
  mode: auto  # auto | cost-first | quality-first | manual
  complexity_threshold: 0.5

learning:
  repo: ""
  auto_extract: true
  min_confidence: 0.6
`

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

func getNestedValue(m map[string]interface{}, key string) interface{} {
	// Simple implementation - supports "tools.claude.enabled"
	parts := splitKey(key)
	current := interface{}(m)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

func setNestedValue(m map[string]interface{}, key string, value string) {
	parts := splitKey(key)
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set value
			// Special handling for tech_stack (comma-separated list)
			if part == "tech_stack" || part == "tech-stack" {
				// Parse comma-separated values into array
				items := strings.Split(value, ",")
				var arr []string
				for _, item := range items {
					item = strings.TrimSpace(item)
					if item != "" {
						arr = append(arr, item)
					}
				}
				current["tech_stack"] = arr
			} else if value == "true" {
				current[part] = true
			} else if value == "false" {
				current[part] = false
			} else {
				current[part] = value
			}
		} else {
			// Create nested map if needed
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]interface{})
			}
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				// Can't go deeper
				return
			}
		}
	}
}

func splitKey(key string) []string {
	var parts []string
	current := ""
	for _, c := range key {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
