// Package ui provides an interactive web UI for editing session workflows.
package ui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mur-run/mur-core/internal/session"
	"gopkg.in/yaml.v3"
)

//go:embed templates/editor.html
var templateFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server hosts the workflow editor web UI.
type Server struct {
	workflow  *session.AnalysisResult
	mu        sync.RWMutex
	clients   map[*websocket.Conn]bool
	clientMu  sync.Mutex
	sessionID string
	tmpl      *template.Template
	shutdown  chan struct{}
	llm       session.LLMProvider
}

type templateData struct {
	SessionID    string
	WorkflowJSON template.JS
}

// NewServer creates a new web UI server for the given analysis result.
func NewServer(result *session.AnalysisResult, sessionID string) (*Server, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/editor.html")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	// Optional LLM provider for AI review
	llm, _ := session.NewLLMProviderFromEnv()

	return &Server{
		workflow:  result,
		sessionID: sessionID,
		clients:   make(map[*websocket.Conn]bool),
		tmpl:      tmpl,
		shutdown:  make(chan struct{}),
		llm:       llm,
	}, nil
}

// Serve starts the HTTP server on the given port.
// If the port is taken, it auto-finds an available one.
func (s *Server) Serve(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/save", s.handleSave)
	mux.HandleFunc("/api/discard", s.handleDiscard)
	mux.HandleFunc("/api/export", s.handleExport)

	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("no available port: %w", err)
		}
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)
	fmt.Printf("MUR Workflow Editor running at %s\n", url)
	fmt.Printf("Press Ctrl+C to stop\n")

	go openBrowser(url)

	srv := &http.Server{Handler: mux}
	go func() {
		<-s.shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	err = srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	wfJSON, err := json.Marshal(s.workflow)
	s.mu.RUnlock()

	if err != nil {
		http.Error(w, "failed to marshal workflow", http.StatusInternalServerError)
		return
	}

	shortID := s.sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	data := templateData{
		SessionID:    shortID,
		WorkflowJSON: template.JS(wfJSON),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	result := s.workflow
	s.mu.RUnlock()

	if err := session.SaveAnalysis(s.sessionID, result); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	shortID := s.sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	s.broadcastMessage(ServerMessage{
		Type: "save.success",
		Path: fmt.Sprintf("~/.mur/session/analysis/%s.json", shortID),
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "name": result.Name})

	go func() {
		time.Sleep(500 * time.Millisecond)
		close(s.shutdown)
	}()
}

func (s *Server) handleDiscard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "discarded"})

	go func() {
		time.Sleep(500 * time.Millisecond)
		close(s.shutdown)
	}()
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	result := s.workflow
	s.mu.RUnlock()

	type workflowExport struct {
		Kind        string             `yaml:"kind"`
		Version     string             `yaml:"version"`
		Name        string             `yaml:"name"`
		Trigger     string             `yaml:"trigger"`
		Description string             `yaml:"description"`
		Variables   []session.Variable `yaml:"variables,omitempty"`
		Steps       []session.Step     `yaml:"steps"`
		Tools       []string           `yaml:"tools,omitempty"`
		Tags        []string           `yaml:"tags,omitempty"`
	}

	export := workflowExport{
		Kind:        "workflow",
		Version:     "1",
		Name:        result.Name,
		Trigger:     result.Trigger,
		Description: result.Description,
		Variables:   result.Variables,
		Steps:       result.Steps,
		Tools:       result.Tools,
		Tags:        result.Tags,
	}

	yamlData, err := yaml.Marshal(export)
	if err != nil {
		http.Error(w, "failed to marshal YAML", http.StatusInternalServerError)
		return
	}

	filename := result.Name
	if filename == "" {
		filename = "workflow"
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.yaml", filename))
	_, _ = w.Write(yamlData)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
