package workflow

import (
	"time"
)

// WorkflowSyncPayload is the wire format for syncing workflows to/from the cloud.
type WorkflowSyncPayload struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Trigger        string   `json:"trigger"`
	Steps          int      `json:"steps"`
	Tools          []string `json:"tools,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Version        int64    `json:"version"`
	PublishedVer   int      `json:"published_version"`
	Deleted        bool     `json:"deleted"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	WorkflowData   []byte   `json:"workflow_data"` // full YAML, encrypted at rest on server
}

// WorkflowSyncChange represents a single change in a sync push.
type WorkflowSyncChange struct {
	Action   string               `json:"action"` // "create", "update", "delete"
	ID       string               `json:"id,omitempty"`
	Workflow *WorkflowSyncPayload `json:"workflow,omitempty"`
}

// WorkflowPushRequest is the request body for pushing workflow changes.
type WorkflowPushRequest struct {
	BaseVersion int64                 `json:"base_version"`
	Changes     []WorkflowSyncChange  `json:"changes"`
}

// WorkflowPushResponse is the response from a push operation.
type WorkflowPushResponse struct {
	OK       bool                  `json:"ok"`
	Version  int64                 `json:"version"`
	Conflicts []WorkflowConflict   `json:"conflicts,omitempty"`
}

// WorkflowConflict represents a sync conflict.
type WorkflowConflict struct {
	WorkflowID    string               `json:"workflow_id"`
	WorkflowName  string               `json:"workflow_name"`
	ServerVersion *WorkflowSyncPayload `json:"server_version"`
	ClientVersion *WorkflowSyncPayload `json:"client_version"`
}

// WorkflowPullResponse is the response from a pull operation.
type WorkflowPullResponse struct {
	Workflows []WorkflowSyncPayload `json:"workflows"`
	Version   int64                 `json:"version"`
}

// WorkflowSyncStatus represents the sync status for workflows.
type WorkflowSyncStatus struct {
	ServerVersion int64 `json:"server_version"`
	HasUpdates    bool  `json:"has_updates"`
}

// BuildSyncPayload converts a local Workflow + Metadata into a sync payload.
func BuildSyncPayload(wf *Workflow, meta *Metadata) *WorkflowSyncPayload {
	return &WorkflowSyncPayload{
		ID:           wf.ID,
		Name:         wf.Name,
		Description:  wf.Description,
		Trigger:      wf.Trigger,
		Steps:        len(wf.Steps),
		Tools:        wf.Tools,
		Tags:         wf.Tags,
		PublishedVer: meta.PublishedVersion,
		CreatedAt:    meta.CreatedAt,
		UpdatedAt:    meta.UpdatedAt,
	}
}

// BuildChangesFromLocal scans all local workflows and builds sync changes.
func BuildChangesFromLocal() ([]WorkflowSyncChange, error) {
	entries, err := List()
	if err != nil {
		return nil, err
	}

	var changes []WorkflowSyncChange
	for _, entry := range entries {
		wf, meta, err := Get(entry.ID)
		if err != nil {
			continue
		}
		payload := BuildSyncPayload(wf, meta)
		changes = append(changes, WorkflowSyncChange{
			Action:   "create",
			ID:       wf.ID,
			Workflow: payload,
		})
	}

	return changes, nil
}
