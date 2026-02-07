package server

import (
	_ "embed"
	"net/http"
)

//go:embed static/index.html
var indexHTML string

// handleIndex serves the main dashboard page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}
