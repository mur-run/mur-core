package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHookVersion(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"v1", "#!/bin/bash\n# mur-managed-hook v1\nmur sync\n", 1},
		{"v3", "#!/bin/bash\n# mur-managed-hook v3\n", 3},
		{"no version", "#!/bin/bash\nmur sync\n", 0},
		{"version on line 5", "#!/bin/bash\n#\n#\n#\n# mur-managed-hook v2\n", 2},
		{"version too deep", "#!/bin/bash\n#\n#\n#\n#\n# mur-managed-hook v2\n", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := filepath.Join(dir, tt.name+".sh")
			os.WriteFile(p, []byte(tt.content), 0644)
			got := parseHookVersion(p)
			if got != tt.want {
				t.Errorf("parseHookVersion() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestShouldUpgradeHook(t *testing.T) {
	dir := t.TempDir()

	// Non-existent file
	if !shouldUpgradeHook(filepath.Join(dir, "nope.sh")) {
		t.Error("should upgrade non-existent file")
	}

	// Old version
	old := filepath.Join(dir, "old.sh")
	os.WriteFile(old, []byte("#!/bin/bash\n# mur-managed-hook v0\n"), 0644)
	if !shouldUpgradeHook(old) {
		t.Error("should upgrade v0")
	}

	// Current version
	cur := filepath.Join(dir, "current.sh")
	os.WriteFile(cur, []byte("#!/bin/bash\n# mur-managed-hook v1\n"), 0644)
	if shouldUpgradeHook(cur) {
		t.Error("should NOT upgrade current version")
	}
}

func TestShouldUpgradeHookForce(t *testing.T) {
	dir := t.TempDir()

	// Current version, no force — should not upgrade
	cur := filepath.Join(dir, "current.sh")
	os.WriteFile(cur, []byte("#!/bin/bash\n# mur-managed-hook v1\n"), 0644)
	if ShouldUpgradeHook(cur, false) {
		t.Error("should NOT upgrade current version without force")
	}

	// Current version, with force — should upgrade
	if !ShouldUpgradeHook(cur, true) {
		t.Error("should upgrade current version with force")
	}

	// Non-existent file, no force — should upgrade
	if !ShouldUpgradeHook(filepath.Join(dir, "nope.sh"), false) {
		t.Error("should upgrade non-existent file")
	}
}

func TestParseHookVersionExported(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.sh")
	os.WriteFile(p, []byte("#!/bin/bash\n# mur-managed-hook v3\n"), 0644)
	if got := ParseHookVersion(p); got != 3 {
		t.Errorf("ParseHookVersion() = %d, want 3", got)
	}
}
