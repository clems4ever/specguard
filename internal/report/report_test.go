package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func sampleReport() *lint.Report {
	return &lint.Report{
		OK:        true,
		TestFiles: 3,
		Specs: []lint.SpecStatus{
			{ID: "auth-login", Title: "Log in", Covered: true},
		},
	}
}

const tinyTemplate = `<!doctype html><html><head><title>t</title>` +
	marker + `</head><body><script type="module">boot()</script></body></html>`

func render(t *testing.T, tmpl string, rep *lint.Report, meta Meta) string {
	t.Helper()
	var b strings.Builder
	if err := Render(&b, tmpl, rep, meta); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return b.String()
}

// extractPayload pulls the two JSON literals out of the injected data script.
func extractPayload(t *testing.T, html string) (repJSON, metaJSON string) {
	t.Helper()
	const a = "window.__SPECGUARD_REPORT__="
	const b = ";window.__SPECGUARD_META__="
	const c = ";</script>"
	i := strings.Index(html, a)
	j := strings.Index(html, b)
	k := strings.Index(html, c)
	if i < 0 || j < 0 || k < 0 || !(i < j && j < k) {
		t.Fatalf("data script not found/ordered in output:\n%s", html)
	}
	return html[i+len(a) : j], html[j+len(b) : k]
}

func TestRenderEmbedsRoundTrippableReport(t *testing.T) {
	rep := sampleReport()
	out := render(t, tinyTemplate, rep, Meta{Branch: "main", Commit: "deadbeef"})

	repJSON, metaJSON := extractPayload(t, out)

	var gotRep lint.Report
	if err := json.Unmarshal([]byte(repJSON), &gotRep); err != nil {
		t.Fatalf("report JSON does not parse: %v", err)
	}
	if len(gotRep.Specs) != 1 || gotRep.Specs[0].ID != "auth-login" || !gotRep.OK {
		t.Fatalf("round-trip lost data: %+v", gotRep)
	}
	var gotMeta Meta
	if err := json.Unmarshal([]byte(metaJSON), &gotMeta); err != nil {
		t.Fatalf("meta JSON does not parse: %v", err)
	}
	if gotMeta.Branch != "main" || gotMeta.Commit != "deadbeef" {
		t.Fatalf("meta round-trip: %+v", gotMeta)
	}
}

func TestRenderReplacesMarkerAndKeepsAppScript(t *testing.T) {
	out := render(t, tinyTemplate, sampleReport(), Meta{})
	if strings.Contains(out, marker) {
		t.Error("marker should have been replaced")
	}
	if !strings.Contains(out, "boot()") {
		t.Error("app script must be preserved")
	}
	// The data script must precede the app's module script so the report is on
	// window before the app boots.
	if strings.Index(out, "__SPECGUARD_REPORT__") > strings.Index(out, `<script type="module">`) {
		t.Error("data script must come before the app module script")
	}
}

func TestRenderFallsBackToHeadWhenNoMarker(t *testing.T) {
	tmpl := `<html><head><title>x</title></head><body></body></html>`
	out := render(t, tmpl, sampleReport(), Meta{})
	if !strings.Contains(out, "__SPECGUARD_REPORT__") {
		t.Fatal("expected data script injected via </head> fallback")
	}
	if strings.Index(out, "__SPECGUARD_REPORT__") > strings.Index(out, "</head>") {
		t.Error("fallback data script must be inside <head>")
	}
}

func TestRenderErrorsWithNoInjectionPoint(t *testing.T) {
	var b strings.Builder
	if err := Render(&b, "<html>nothing</html>", sampleReport(), Meta{}); err == nil {
		t.Fatal("expected an error when there is no marker or </head>")
	}
}

// The payload must be inert inside a <script>: a spec body containing </script>
// or an HTML comment must not break out of the tag.
func TestRenderEscapesHTMLBreakout(t *testing.T) {
	rep := &lint.Report{
		OK: true,
		Specs: []lint.SpecStatus{
			{ID: "x", Title: "danger", Body: "</script><!-- <img>  "},
		},
	}
	repJSON, metaJSON := extractPayload(t, render(t, tinyTemplate, rep, Meta{}))
	payload := repJSON + metaJSON
	for _, bad := range []string{"</script", "<!--", "<", " ", " "} {
		if strings.Contains(payload, bad) {
			t.Errorf("payload must not contain raw %q (HTML/JS breakout risk)", bad)
		}
	}
}

func TestTemplateEmbeddedIsUsable(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("embedded template: %v", err)
	}
	if !strings.Contains(tmpl, marker) && !strings.Contains(tmpl, "</head>") {
		t.Fatal("embedded template has no injection point")
	}
	// It must actually render without error and boot from the embedded report.
	out := render(t, tmpl, sampleReport(), Meta{Branch: "main"})
	if !strings.Contains(out, "__SPECGUARD_REPORT__") {
		t.Fatal("embedded template did not receive the data script")
	}
}
