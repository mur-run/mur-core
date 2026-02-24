package workflow

import (
	"testing"
)

func TestSetAndGetPermission(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Set permission
	if err := SetPermission("wf-perm", "alice@example.com", PermissionRead, "owner"); err != nil {
		t.Fatalf("SetPermission() error: %v", err)
	}

	// Get permission
	perm, err := GetPermission("wf-perm", "alice@example.com")
	if err != nil {
		t.Fatalf("GetPermission() error: %v", err)
	}
	if perm != PermissionRead {
		t.Errorf("permission = %q, want %q", perm, PermissionRead)
	}
}

func TestSetPermission_Update(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm-update")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Set read, then update to write
	SetPermission("wf-perm-update", "bob@example.com", PermissionRead, "owner")
	SetPermission("wf-perm-update", "bob@example.com", PermissionWrite, "owner")

	perm, _ := GetPermission("wf-perm-update", "bob@example.com")
	if perm != PermissionWrite {
		t.Errorf("permission = %q, want %q", perm, PermissionWrite)
	}
}

func TestGetPermission_NotFound(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm-nf")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	perm, err := GetPermission("wf-perm-nf", "nobody@example.com")
	if err != nil {
		t.Fatalf("GetPermission() error: %v", err)
	}
	if perm != "" {
		t.Errorf("permission = %q, want empty", perm)
	}
}

func TestListPermissions(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm-list")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	SetPermission("wf-perm-list", "a@test.com", PermissionRead, "owner")
	SetPermission("wf-perm-list", "b@test.com", PermissionExecuteOnly, "owner")

	perms, err := ListPermissions("wf-perm-list")
	if err != nil {
		t.Fatalf("ListPermissions() error: %v", err)
	}
	if len(perms) != 2 {
		t.Fatalf("len(perms) = %d, want 2", len(perms))
	}
}

func TestRemovePermission(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm-rm")
	if err := Create(wf); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	SetPermission("wf-perm-rm", "remove@test.com", PermissionRead, "owner")
	RemovePermission("wf-perm-rm", "remove@test.com")

	perm, _ := GetPermission("wf-perm-rm", "remove@test.com")
	if perm != "" {
		t.Errorf("permission after remove = %q, want empty", perm)
	}
}

func TestValidPermission(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"read", true},
		{"write", true},
		{"execute-only", true},
		{"admin", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := ValidPermission(tt.input); got != tt.want {
			t.Errorf("ValidPermission(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSetPermission_InvalidPerm(t *testing.T) {
	setWorkflowsDir(t)

	wf := sampleWorkflow("wf-perm-invalid")
	Create(wf)

	err := SetPermission("wf-perm-invalid", "x@test.com", Permission("admin"), "owner")
	if err == nil {
		t.Error("expected error for invalid permission")
	}
}

func TestSetPermission_WorkflowNotFound(t *testing.T) {
	setWorkflowsDir(t)

	err := SetPermission("nonexistent", "x@test.com", PermissionRead, "owner")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}
