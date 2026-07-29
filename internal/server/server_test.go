package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func newTestServer(t *testing.T, files map[string]string, webDir string) *httptest.Server {
	t.Helper()
	root := writeTree(t, files)
	cfg := lint.DefaultConfig(root)
	ts := httptest.NewServer(New(cfg, ResolveWebDir(webDir)).Handler())
	t.Cleanup(ts.Close)
	return ts
}

func TestReportEndpointReturnsLiveJSON(t *testing.T) {
	ts := newTestServer(t, map[string]string{
		"specs/edit.md":            "---\nid: skills-edit\ntitle: Editing persists\n---\n## Why\nbecause\n",
		"internal/skill/x_test.go": "// spec:skills-edit\n",
	}, "")

	resp, err := http.Get(ts.URL + "/api/report")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}
	var rep lint.Report
	if err := json.NewDecoder(resp.Body).Decode(&rep); err != nil {
		t.Fatal(err)
	}
	if !rep.OK || len(rep.Specs) != 1 {
		t.Fatalf("unexpected report: %+v", rep)
	}
	s := rep.Specs[0]
	if s.ID != "skills-edit" || !s.Covered {
		t.Fatalf("spec status wrong: %+v", s)
	}
	// The markdown body must reach the client so the UI can render it.
	if s.Body != "## Why\nbecause\n" {
		t.Fatalf("body not served: %q", s.Body)
	}
}

func TestReportReflectsChangesBetweenRequests(t *testing.T) {
	root := writeTree(t, map[string]string{
		"specs/edit.md":            "---\nid: skills-edit\ntitle: t\n---\nbody\n",
		"internal/skill/x_test.go": "// spec:skills-edit\n",
	})
	ts := httptest.NewServer(New(lint.DefaultConfig(root), "").Handler())
	t.Cleanup(ts.Close)

	get := func() lint.Report {
		resp, err := http.Get(ts.URL + "/api/report")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var rep lint.Report
		if err := json.NewDecoder(resp.Body).Decode(&rep); err != nil {
			t.Fatal(err)
		}
		return rep
	}

	if !get().OK {
		t.Fatal("expected initial PASS")
	}
	// Remove the coverage on disk; the very next request must reflect it.
	if err := os.WriteFile(filepath.Join(root, "internal/skill/x_test.go"), []byte("// no tag\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if get().OK {
		t.Fatal("expected FAIL after coverage removed — report was stale")
	}
}

func TestHealthz(t *testing.T) {
	ts := newTestServer(t, map[string]string{}, "")
	resp, err := http.Get(ts.URL + "/api/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCORSHeaderPresent(t *testing.T) {
	ts := newTestServer(t, map[string]string{}, "")
	resp, err := http.Get(ts.URL + "/api/report")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header for the Vite dev origin")
	}
}

func TestStaticSPAFallback(t *testing.T) {
	web := writeTree(t, map[string]string{
		"index.html":    "<!doctype html><title>specguard</title>",
		"assets/app.js": "console.log('app')",
	})
	ts := newTestServer(t, map[string]string{}, web)

	// A real asset is served as-is.
	resp, err := http.Get(ts.URL + "/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("asset status = %d", resp.StatusCode)
	}

	// An unknown client-side route falls back to index.html (SPA routing).
	resp, err = http.Get(ts.URL + "/spec/skills-edit")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body := make([]byte, 64)
	n, _ := resp.Body.Read(body)
	if resp.StatusCode != http.StatusOK || string(body[:n]) == "" {
		t.Fatalf("SPA fallback failed: status %d", resp.StatusCode)
	}
}

func TestDiffDisabledWithoutGit(t *testing.T) {
	// A tempdir tree is not a git repo → /api/diff reports disabled, never errors.
	ts := newTestServer(t, map[string]string{
		"specs/a.md":         "---\nid: a\ntitle: A\n---\n",
		"internal/x_test.go": "// spec:a\n",
	}, "")
	resp, err := http.Get(ts.URL + "/api/diff")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["enabled"] != false {
		t.Fatalf("diff should be disabled when not enabled/no git; got %+v", body)
	}
}

func TestBadgeSVG(t *testing.T) {
	ts := newTestServer(t, map[string]string{
		"specs/a.md":         "---\nid: a\ntitle: A\n---\n",
		"internal/x_test.go": "// spec:a\n",
	}, "")
	resp, err := http.Get(ts.URL + "/api/badge")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "image/svg+xml" {
		t.Fatalf("badge content-type = %q", ct)
	}
	buf := make([]byte, 512)
	n, _ := resp.Body.Read(buf)
	svg := string(buf[:n])
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "specguard") {
		t.Fatalf("badge is not a specguard SVG: %q", svg)
	}
	// One covered spec → green pass badge, not "failing".
	if strings.Contains(svg, "failing") {
		t.Fatal("a passing project must not render a failing badge")
	}
}

func TestResolveWebDir(t *testing.T) {
	if ResolveWebDir("") != "" {
		t.Fatal("empty should stay empty")
	}
	if ResolveWebDir("/nonexistent-xyz") != "" {
		t.Fatal("missing index.html should resolve to empty (API-only)")
	}
	web := writeTree(t, map[string]string{"index.html": "x"})
	if ResolveWebDir(web) != web {
		t.Fatal("a dir with index.html should resolve to itself")
	}
}
