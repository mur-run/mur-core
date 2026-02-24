// Package server provides the HTTP server for the web dashboard.
package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Config holds server configuration.
type Config struct {
	Port int
}

// Server represents the dashboard HTTP server.
type Server struct {
	config Config
	mux    *http.ServeMux
}

// New creates a new server with the given config.
func New(cfg Config) *Server {
	s := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes registers all HTTP routes.
func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/stats/daily", s.handleStatsDaily)
	s.mux.HandleFunc("/api/patterns", s.handlePatterns)
	s.mux.HandleFunc("/api/patterns/", s.handlePatternByName)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/sync", s.handleSync)

	// Workflow API routes
	s.mux.HandleFunc("/api/workflows", s.handleWorkflows)
	s.mux.HandleFunc("/api/workflows/", s.handleWorkflowByID)

	// Static files (embedded HTML)
	s.mux.HandleFunc("/", s.handleIndex)
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🌐 Dashboard running at http://localhost%s", addr)
	log.Printf("   Press Ctrl+C to stop")

	return srv.ListenAndServe()
}
