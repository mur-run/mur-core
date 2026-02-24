package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/mur-run/mur-core/internal/workflow"
	"gopkg.in/yaml.v3"
)

// handleWorkflows dispatches /api/workflows requests.
func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleWorkflowsList(w, r)
	case http.MethodPost:
		s.handleWorkflowsCreate(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
	}
}

// handleWorkflowByID dispatches /api/workflows/{id} requests.
func (s *Server) handleWorkflowByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/workflows/")

	// Check for sub-routes: /api/workflows/{id}/publish, /api/workflows/{id}/export
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		id = parts[0]
		sub := parts[1]

		switch sub {
		case "publish":
			s.handleWorkflowPublish(w, r, id)
		case "export":
			s.handleWorkflowExport(w, r, id)
		default:
			writeJSON(w, http.StatusNotFound, APIResponse{Error: "not found"})
		}
		return
	}

	if id == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "workflow ID required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleWorkflowGet(w, r, id)
	case http.MethodPut:
		s.handleWorkflowUpdate(w, r, id)
	case http.MethodDelete:
		s.handleWorkflowDelete(w, r, id)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
	}
}

// GET /api/workflows
func (s *Server) handleWorkflowsList(w http.ResponseWriter, r *http.Request) {
	entries, err := workflow.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}
	if entries == nil {
		entries = []workflow.IndexEntry{}
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: entries})
}

// GET /api/workflows/{id}
func (s *Server) handleWorkflowGet(w http.ResponseWriter, r *http.Request, id string) {
	wf, meta, err := workflow.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: err.Error()})
		return
	}

	result := map[string]interface{}{
		"workflow": wf,
		"metadata": meta,
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: result})
}

// POST /api/workflows
func (s *Server) handleWorkflowsCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Start     int    `json:"start"`
		End       int    `json:"end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "invalid JSON: " + err.Error()})
		return
	}
	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "session_id is required"})
		return
	}

	wf, err := workflow.CreateFromSession(req.SessionID, req.Start, req.End)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{Success: true, Data: wf})
}

// PUT /api/workflows/{id}
func (s *Server) handleWorkflowUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var wf workflow.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "invalid JSON: " + err.Error()})
		return
	}

	// Ensure the ID matches the URL
	wf.ID = id

	if err := workflow.Update(&wf); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: wf})
}

// DELETE /api/workflows/{id}
func (s *Server) handleWorkflowDelete(w http.ResponseWriter, r *http.Request, id string) {
	if err := workflow.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: map[string]string{"deleted": id}})
}

// POST /api/workflows/{id}/publish
func (s *Server) handleWorkflowPublish(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	version, err := workflow.Publish(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: map[string]interface{}{
		"id":      id,
		"version": version,
	}})
}

// GET /api/workflows/{id}/export?format=yaml|md
func (s *Server) handleWorkflowExport(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	wf, _, err := workflow.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: err.Error()})
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "yaml"
	}

	// Sanitize filename
	safeName := sanitizeFilename(wf.Name)
	if safeName == "" {
		safeName = "workflow"
	}

	switch format {
	case "yaml":
		data, err := yaml.Marshal(wf)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", `attachment; filename="`+safeName+`.yaml"`)
		w.Write(data)

	case "json":
		data, err := json.MarshalIndent(wf, "", "  ")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="`+safeName+`.json"`)
		w.Write(data)

	default:
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "unsupported format: " + format + " (use yaml or json)"})
	}
}

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitizeFilename(name string) string {
	return unsafeFilenameChars.ReplaceAllString(name, "_")
}
