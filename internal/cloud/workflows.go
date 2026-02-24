package cloud

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/workflow"
)

// WorkflowSyncStatus returns workflow sync status for a team.
func (c *Client) WorkflowSyncStatus(teamID string, version int64) (*workflow.WorkflowSyncStatus, error) {
	var status workflow.WorkflowSyncStatus
	path := fmt.Sprintf("/api/v1/core/teams/%s/workflows/sync/status?version=%d", teamID, version)
	if err := c.get(path, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// WorkflowPull pulls workflow changes from the server.
func (c *Client) WorkflowPull(teamID string, sinceVersion int64) (*workflow.WorkflowPullResponse, error) {
	var resp workflow.WorkflowPullResponse
	path := fmt.Sprintf("/api/v1/core/teams/%s/workflows/sync/pull?since=%d", teamID, sinceVersion)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WorkflowPush pushes workflow changes to the server.
func (c *Client) WorkflowPush(teamID string, req workflow.WorkflowPushRequest) (*workflow.WorkflowPushResponse, error) {
	var resp workflow.WorkflowPushResponse
	path := fmt.Sprintf("/api/v1/core/teams/%s/workflows/sync/push", teamID)
	if err := c.post(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
