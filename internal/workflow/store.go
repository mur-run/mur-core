package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// workflowsDirFunc is the function used to resolve the workflows directory.
// Tests can override this to point at a temp directory.
var workflowsDirFunc = defaultWorkflowsDir

func defaultWorkflowsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "workflows"), nil
}

func workflowsDir() (string, error) {
	return workflowsDirFunc()
}

// workflowDir returns the path for a specific workflow: ~/.mur/workflows/<id>/
func workflowDir(id string) (string, error) {
	base, err := workflowsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, id), nil
}

// Create persists a new workflow to disk and updates the index.
func Create(wf *Workflow) error {
	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	dir, err := workflowDir(wf.ID)
	if err != nil {
		return err
	}

	// Check if already exists
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("workflow %s already exists", wf.ID)
	}

	// Create directory structure
	revisionsDir := filepath.Join(dir, "revisions")
	if err := os.MkdirAll(revisionsDir, 0755); err != nil {
		return fmt.Errorf("create workflow directory: %w", err)
	}

	// Write workflow.yaml
	if err := writeWorkflowFile(dir, wf); err != nil {
		return err
	}

	// Write metadata.json
	now := time.Now()
	meta := &Metadata{
		ID:               wf.ID,
		Name:             wf.Name,
		CreatedAt:        now,
		UpdatedAt:        now,
		PublishedVersion: 0,
		RevisionCount:    1,
	}
	if err := writeMetadata(dir, meta); err != nil {
		return err
	}

	// Save initial revision
	if err := saveRevision(dir, wf, 1); err != nil {
		return err
	}

	// Update index
	return updateIndex(wf, meta)
}

// Get loads a workflow by ID.
func Get(id string) (*Workflow, *Metadata, error) {
	dir, err := workflowDir(id)
	if err != nil {
		return nil, nil, err
	}

	wf, err := readWorkflowFile(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("load workflow %s: %w", id, err)
	}

	meta, err := readMetadata(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("load metadata %s: %w", id, err)
	}

	return wf, meta, nil
}

// List returns all workflows from the index.
func List() ([]IndexEntry, error) {
	idx, err := readIndex()
	if err != nil {
		return nil, err
	}
	return idx.Workflows, nil
}

// Update saves changes to an existing workflow and creates a new revision.
func Update(wf *Workflow) error {
	dir, err := workflowDir(wf.ID)
	if err != nil {
		return err
	}

	// Verify it exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("workflow %s not found", wf.ID)
	}

	// Read current metadata
	meta, err := readMetadata(dir)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	// Write updated workflow.yaml
	if err := writeWorkflowFile(dir, wf); err != nil {
		return err
	}

	// Bump revision
	meta.RevisionCount++
	meta.UpdatedAt = time.Now()
	meta.Name = wf.Name
	if err := writeMetadata(dir, meta); err != nil {
		return err
	}

	// Save revision snapshot
	if err := saveRevision(dir, wf, meta.RevisionCount); err != nil {
		return err
	}

	return updateIndex(wf, meta)
}

// Delete removes a workflow from disk and the index.
func Delete(id string) error {
	dir, err := workflowDir(id)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("workflow %s not found", id)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}

	return removeFromIndex(id)
}

// Publish bumps the published version number on a workflow.
func Publish(id string) (int, error) {
	dir, err := workflowDir(id)
	if err != nil {
		return 0, err
	}

	meta, err := readMetadata(dir)
	if err != nil {
		return 0, fmt.Errorf("read metadata: %w", err)
	}

	meta.PublishedVersion++
	meta.UpdatedAt = time.Now()
	if err := writeMetadata(dir, meta); err != nil {
		return 0, err
	}

	// Update index with new version
	wf, err := readWorkflowFile(dir)
	if err != nil {
		return 0, err
	}
	if err := updateIndex(wf, meta); err != nil {
		return 0, err
	}

	return meta.PublishedVersion, nil
}

// --- file I/O helpers ---

func writeWorkflowFile(dir string, wf *Workflow) error {
	data, err := yaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "workflow.yaml"), data, 0644)
}

func readWorkflowFile(dir string) (*Workflow, error) {
	data, err := os.ReadFile(filepath.Join(dir, "workflow.yaml"))
	if err != nil {
		return nil, err
	}
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}
	return &wf, nil
}

func writeMetadata(dir string, meta *Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0644)
}

func readMetadata(dir string) (*Metadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

func saveRevision(dir string, wf *Workflow, revNum int) error {
	revisionsDir := filepath.Join(dir, "revisions")
	if err := os.MkdirAll(revisionsDir, 0755); err != nil {
		return fmt.Errorf("create revisions dir: %w", err)
	}

	data, err := yaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshal revision: %w", err)
	}

	filename := fmt.Sprintf("rev-%03d.yaml", revNum)
	return os.WriteFile(filepath.Join(revisionsDir, filename), data, 0644)
}

// --- index management ---

func indexPath() (string, error) {
	base, err := workflowsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "index.json"), nil
}

func readIndex() (*Index, error) {
	path, err := indexPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Index{}, nil
		}
		return nil, fmt.Errorf("read index: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	return &idx, nil
}

func writeIndex(idx *Index) error {
	path, err := indexPath()
	if err != nil {
		return err
	}

	base := filepath.Dir(path)
	if err := os.MkdirAll(base, 0755); err != nil {
		return fmt.Errorf("create workflows dir: %w", err)
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func updateIndex(wf *Workflow, meta *Metadata) error {
	idx, err := readIndex()
	if err != nil {
		return err
	}

	entry := IndexEntry{
		ID:               wf.ID,
		Name:             wf.Name,
		Description:      wf.Description,
		Tags:             wf.Tags,
		CreatedAt:        meta.CreatedAt,
		UpdatedAt:        meta.UpdatedAt,
		PublishedVersion: meta.PublishedVersion,
	}

	// Update existing or append
	found := false
	for i, e := range idx.Workflows {
		if e.ID == wf.ID {
			idx.Workflows[i] = entry
			found = true
			break
		}
	}
	if !found {
		idx.Workflows = append(idx.Workflows, entry)
	}

	// Sort by updated time, newest first
	sort.Slice(idx.Workflows, func(i, j int) bool {
		return idx.Workflows[i].UpdatedAt.After(idx.Workflows[j].UpdatedAt)
	})

	return writeIndex(idx)
}

func removeFromIndex(id string) error {
	idx, err := readIndex()
	if err != nil {
		return err
	}

	filtered := idx.Workflows[:0]
	for _, e := range idx.Workflows {
		if e.ID != id {
			filtered = append(filtered, e)
		}
	}
	idx.Workflows = filtered

	return writeIndex(idx)
}
