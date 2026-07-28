// Package server exposes a spec-traceability report over HTTP so a browser UI
// can render it. The report is recomputed on every request, so the data is
// always live — edit a spec or a test and refresh.
package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/clems4ever/specguard/internal/lint"
)

// Server serves the JSON API and, if a built UI is present, the static assets.
type Server struct {
	cfg    lint.Config
	webDir string // optional: a directory of built UI assets to serve
}

// New returns a Server that lints cfg.Root. webDir may be empty (API only).
func New(cfg lint.Config, webDir string) *Server {
	return &Server{cfg: cfg, webDir: webDir}
}

// Handler builds the HTTP routing.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/report", s.handleReport)
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if s.webDir != "" {
		mux.Handle("/", s.staticHandler())
	}
	return withCORS(mux)
}

// handleReport runs the lint and returns the report as JSON.
func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	rep, err := lint.Run(s.cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

// staticHandler serves the built UI, falling back to index.html so client-side
// routes (e.g. /spec/skills-edit) resolve on a hard refresh.
func (s *Server) staticHandler() http.Handler {
	fs := http.FileServer(http.Dir(s.webDir))
	index := filepath.Join(s.webDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := filepath.Clean(r.URL.Path)
		candidate := filepath.Join(s.webDir, clean)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/" {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// withCORS allows the Vite dev server (a different origin during development) to
// call the API.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ResolveWebDir returns dir if it contains an index.html, else "". This lets
// `specguard serve` degrade to an API-only server when no UI has been built.
func ResolveWebDir(dir string) string {
	if dir == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
		return strings.TrimRight(dir, "/")
	}
	return ""
}

// ListenAndServe is a thin wrapper so the command layer needn't import net/http.
func ListenAndServe(addr string, h http.Handler) error {
	return http.ListenAndServe(addr, h)
}
