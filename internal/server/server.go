// Package server exposes a spec-traceability report over HTTP so a browser UI
// can render it. The report is recomputed on every request, so the data is
// always live — edit a spec or a test and refresh.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/clems4ever/specguard/internal/diff"
	"github.com/clems4ever/specguard/internal/lint"
)

// Server serves the JSON API and, if a built UI is present, the static assets.
type Server struct {
	cfg    lint.Config
	webDir string // optional: a directory of built UI assets to serve
	// diffBase is the default git ref the "Changed only" view diffs against
	// (e.g. "HEAD" for uncommitted work, "origin/main" for a branch). Empty
	// leaves /api/diff reporting disabled.
	diffBase string
	// configName is the config file's basename, needed to re-run the linter in a
	// base worktree. Defaults to ".specguard.yml".
	configName string
}

// New returns a Server that lints cfg.Root. webDir may be empty (API only).
func New(cfg lint.Config, webDir string) *Server {
	return &Server{cfg: cfg, webDir: webDir}
}

// EnableDiff turns on the /api/diff endpoint, diffing against base (a git ref)
// and re-running the linter in a base worktree loaded from configName.
func (s *Server) EnableDiff(base, configName string) {
	s.diffBase = base
	s.configName = configName
}

// Handler builds the HTTP routing.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/report", s.handleReport)
	mux.HandleFunc("/api/diff", s.handleDiff)
	mux.HandleFunc("/api/badge", s.handleBadge)
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

// handleDiff computes the delta of the working tree against a base git ref, so
// the UI can show only the specs a change touched. Query param `base` overrides
// the server default.
func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	root := s.cfg.Root
	if root == "" {
		root = "."
	}
	if s.diffBase == "" || !diff.HasGit(root) {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
		return
	}
	base := r.URL.Query().Get("base")
	if base == "" {
		base = s.diffBase
	}
	configName := s.configName
	if configName == "" {
		configName = ".specguard.yml"
	}

	head, err := lint.Run(s.cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	baseRep, err := diff.BaseReport(root, base, configName)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "error": err.Error(), "base": base})
		return
	}
	changed, err := diff.ChangedFiles(root, base)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "error": err.Error(), "base": base})
		return
	}
	d := diff.Compute(baseRep, head, changed)
	d.Base = base
	writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "delta": d})
}

// handleBadge serves a shields-style SVG summarising the current check, so a
// README can embed a live pass/coverage badge.
func (s *Server) handleBadge(w http.ResponseWriter, r *http.Request) {
	rep, err := lint.Run(s.cfg)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	if err != nil {
		_, _ = w.Write(badgeSVG("specguard", "error", "#9f9f9f"))
		return
	}
	covered := 0
	for _, sp := range rep.Specs {
		if sp.Covered {
			covered++
		}
	}
	pct := 100
	if len(rep.Specs) > 0 {
		pct = covered * 100 / len(rep.Specs)
	}
	msg := fmt.Sprintf("%d specs · %d%%", len(rep.Specs), pct)
	color := "#4c1" // green
	if !rep.OK {
		msg = "failing"
		color = "#e05d44" // red
	}
	_, _ = w.Write(badgeSVG("specguard", msg, color))
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
