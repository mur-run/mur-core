package server

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/karajanchang/murmur-ai/internal/config"
	"github.com/karajanchang/murmur-ai/internal/learn"
	"github.com/karajanchang/murmur-ai/internal/stats"
)

// APIResponse is the standard API response format.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// handleStats returns usage statistics summary.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	// Get all records
	records, err := stats.Query(stats.QueryFilter{})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	summary := stats.Summarize(records)
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: summary})
}

// handleStatsDaily returns daily usage breakdown.
func (s *Server) handleStatsDaily(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	// Get records from last 30 days
	filter := stats.QueryFilter{
		StartTime: time.Now().AddDate(0, 0, -30),
	}
	records, err := stats.Query(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	// Group by day
	daily := make(map[string]int)
	for _, rec := range records {
		day := rec.Timestamp.Format("2006-01-02")
		daily[day]++
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: daily})
}

// handlePatterns returns all patterns.
func (s *Server) handlePatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	patterns, err := learn.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	// Apply filters
	domain := r.URL.Query().Get("domain")
	category := r.URL.Query().Get("category")

	var filtered []learn.Pattern
	for _, p := range patterns {
		if domain != "" && p.Domain != domain {
			continue
		}
		if category != "" && p.Category != category {
			continue
		}
		filtered = append(filtered, p)
	}

	if filtered == nil {
		filtered = []learn.Pattern{}
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: filtered})
}

// handlePatternByName returns a single pattern.
func (s *Server) handlePatternByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	// Extract pattern name from path: /api/patterns/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/patterns/")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "pattern name required"})
		return
	}

	pattern, err := learn.Get(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: pattern})
}

// handleConfig returns current configuration.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: cfg})
}

// ToolHealth represents health status of a tool.
type ToolHealth struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Installed bool   `json:"installed"`
	Binary    string `json:"binary"`
	Tier      string `json:"tier"`
}

// handleHealth returns health status of all tools.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Error: err.Error()})
		return
	}

	var tools []ToolHealth
	for name, tool := range cfg.Tools {
		// Check if binary exists
		_, err := exec.LookPath(tool.Binary)
		installed := err == nil

		tools = append(tools, ToolHealth{
			Name:      name,
			Enabled:   tool.Enabled,
			Installed: installed,
			Binary:    tool.Binary,
			Tier:      tool.Tier,
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: tools})
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// handleSync triggers sync to AI CLI tools.
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		return
	}

	// Run mur sync command
	cmd := exec.Command("mur", "sync")
	output, err := cmd.CombinedOutput()

	result := SyncResult{
		Success: err == nil,
		Message: string(output),
	}

	if err != nil {
		result.Message = err.Error() + ": " + string(output)
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: result.Success, Data: result})
}
