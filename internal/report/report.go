// Package report turns a lint report into a single self-contained HTML file:
// the same web UI, but with the report and its provenance baked in so the spec
// catalog is browsable offline, with no server. This is what `specguard report`
// publishes (e.g. to GitHub Pages, one page per branch).
package report

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/clems4ever/specguard/internal/lint"
)

// marker is the placeholder in web/index.html (preserved through the single-file
// build) that we replace with the data script. Injecting a classic <script>
// here — before the app's deferred module script — guarantees the report is on
// window before the app boots.
const marker = "<!-- specguard:data -->"

// template is the self-contained UI built by `npm run build:single` and copied
// into assets/. It changes only when the frontend changes; a CI check keeps it
// in sync (see .github/workflows/ci.yml). Callers may override it with a freshly
// built file via `specguard report -web`.
//
//go:embed assets/index.html
var template string

// Meta stamps a report with its provenance so a viewer knows exactly what they
// are looking at. Field names mirror the web ReportMeta type.
type Meta struct {
	Repo        string `json:"repo,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Commit      string `json:"commit,omitempty"`
	CommitShort string `json:"commitShort,omitempty"`
	GeneratedAt string `json:"generatedAt,omitempty"` // RFC3339
	// PreviewBase is the root URL of a running preview of this build (e.g. a
	// per-PR deploy). The report joins it to each spec's `preview` path so a PM
	// can open the live feature to review it.
	PreviewBase string `json:"previewBase,omitempty"`
}

// Template returns the embedded single-file UI template. It errors if the embed
// is empty (the binary was built without the built asset).
func Template() (string, error) {
	if strings.TrimSpace(template) == "" {
		return "", fmt.Errorf("no embedded UI template (build the web app: cd web && npm run build:single)")
	}
	return template, nil
}

// Render writes a self-contained HTML report: tmpl with the data script spliced
// in at the marker. rep and meta are serialised as JSON onto window. If tmpl has
// no marker (an unexpected template), the script is injected before </head> as a
// fallback so the output is still usable.
func Render(w io.Writer, tmpl string, rep *lint.Report, meta Meta) error {
	// encoding/json escapes <, >, & (and U+2028/2029) by default, so the JSON is
	// safe to inline in a <script> — no </script> or comment breakout is possible.
	repJSON, err := json.Marshal(rep)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}

	var b strings.Builder
	b.WriteString("<script>window.__SPECGUARD_REPORT__=")
	b.Write(repJSON)
	b.WriteString(";window.__SPECGUARD_META__=")
	b.Write(metaJSON)
	b.WriteString(";</script>")
	dataScript := b.String()

	var out string
	switch {
	case strings.Contains(tmpl, marker):
		out = strings.Replace(tmpl, marker, dataScript, 1)
	case strings.Contains(tmpl, "</head>"):
		out = strings.Replace(tmpl, "</head>", dataScript+"</head>", 1)
	default:
		return fmt.Errorf("template has neither %q nor </head> to inject into", marker)
	}

	_, err = io.WriteString(w, out)
	return err
}
