package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Permission represents an access level for a workflow.
type Permission string

const (
	PermissionRead        Permission = "read"
	PermissionWrite       Permission = "write"
	PermissionExecuteOnly Permission = "execute-only"
)

// WorkflowPermission represents a permission grant for a specific user on a workflow.
type WorkflowPermission struct {
	WorkflowID string     `json:"workflow_id"`
	UserEmail  string     `json:"user_email"`
	Permission Permission `json:"permission"`
	GrantedBy  string     `json:"granted_by"`
	GrantedAt  time.Time  `json:"granted_at"`
}

// PermissionsFile holds all permissions for a workflow.
type PermissionsFile struct {
	Permissions []WorkflowPermission `json:"permissions"`
}

// ValidPermission checks if a permission value is valid.
func ValidPermission(p string) bool {
	switch Permission(p) {
	case PermissionRead, PermissionWrite, PermissionExecuteOnly:
		return true
	}
	return false
}

// permissionsPath returns the path to a workflow's permissions file.
func permissionsPath(workflowID string) (string, error) {
	dir, err := workflowDir(workflowID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "permissions.json"), nil
}

// SetPermission grants or updates a permission for a user on a workflow.
func SetPermission(workflowID, userEmail string, perm Permission, grantedBy string) error {
	if !ValidPermission(string(perm)) {
		return fmt.Errorf("invalid permission: %s", perm)
	}

	// Verify workflow exists
	dir, err := workflowDir(workflowID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	perms, err := loadPermissions(workflowID)
	if err != nil {
		perms = &PermissionsFile{}
	}

	// Update or add
	found := false
	for i, p := range perms.Permissions {
		if p.UserEmail == userEmail {
			perms.Permissions[i].Permission = perm
			perms.Permissions[i].GrantedBy = grantedBy
			perms.Permissions[i].GrantedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		perms.Permissions = append(perms.Permissions, WorkflowPermission{
			WorkflowID: workflowID,
			UserEmail:  userEmail,
			Permission: perm,
			GrantedBy:  grantedBy,
			GrantedAt:  time.Now(),
		})
	}

	return savePermissions(workflowID, perms)
}

// GetPermission returns the permission for a user on a workflow.
// Returns empty string if no permission is set.
func GetPermission(workflowID, userEmail string) (Permission, error) {
	perms, err := loadPermissions(workflowID)
	if err != nil {
		return "", nil // No permissions file = no permissions
	}

	for _, p := range perms.Permissions {
		if p.UserEmail == userEmail {
			return p.Permission, nil
		}
	}
	return "", nil
}

// ListPermissions returns all permissions for a workflow.
func ListPermissions(workflowID string) ([]WorkflowPermission, error) {
	perms, err := loadPermissions(workflowID)
	if err != nil {
		return nil, nil
	}
	return perms.Permissions, nil
}

// RemovePermission removes a user's permission from a workflow.
func RemovePermission(workflowID, userEmail string) error {
	perms, err := loadPermissions(workflowID)
	if err != nil {
		return nil
	}

	var filtered []WorkflowPermission
	for _, p := range perms.Permissions {
		if p.UserEmail != userEmail {
			filtered = append(filtered, p)
		}
	}
	perms.Permissions = filtered

	return savePermissions(workflowID, perms)
}

func loadPermissions(workflowID string) (*PermissionsFile, error) {
	path, err := permissionsPath(workflowID)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pf PermissionsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	return &pf, nil
}

func savePermissions(workflowID string, pf *PermissionsFile) error {
	path, err := permissionsPath(workflowID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
