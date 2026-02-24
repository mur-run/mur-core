// Package workflow manages editable, versionable workflows extracted from sessions.
package workflow

import (
	"time"

	"github.com/mur-run/mur-core/internal/session"
)

// Workflow is an editable, versionable workflow extracted from one or more sessions.
type Workflow struct {
	ID          string             `json:"id" yaml:"id"`
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Trigger     string             `json:"trigger" yaml:"trigger"`
	Variables   []session.Variable `json:"variables" yaml:"variables,omitempty"`
	Steps       []session.Step     `json:"steps" yaml:"steps"`
	Tools       []string           `json:"tools" yaml:"tools,omitempty"`
	Tags        []string           `json:"tags" yaml:"tags,omitempty"`

	// SourceSessions references the sessions this workflow was extracted from.
	SourceSessions []SourceRef `json:"source_sessions" yaml:"source_sessions,omitempty"`
}

// SourceRef references a session that contributed to this workflow.
type SourceRef struct {
	SessionID string `json:"session_id" yaml:"session_id"`
	StartEvent int   `json:"start_event,omitempty" yaml:"start_event,omitempty"`
	EndEvent   int   `json:"end_event,omitempty" yaml:"end_event,omitempty"`
}

// Metadata stores workflow metadata separate from the workflow definition.
type Metadata struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PublishedVersion int       `json:"published_version"`
	RevisionCount    int       `json:"revision_count"`
}

// IndexEntry is a summary of a workflow stored in the index file.
type IndexEntry struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Tags             []string  `json:"tags,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PublishedVersion int       `json:"published_version"`
}

// Index is the top-level structure of ~/.mur/workflows/index.json.
type Index struct {
	Workflows []IndexEntry `json:"workflows"`
}
