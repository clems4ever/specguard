// Package results ingests test-runner output and correlates it back to specs,
// so the report can show whether each spec's covering tests actually passed —
// not just that they exist. It understands two formats, auto-detected per file:
//
//   - Playwright JSON (`--reporter=json`): the `@spec:<id>` tag on each test
//     maps a result straight to a spec.
//   - Go `go test -json`: a stream of events keyed by test-function name, which
//     the linter has already associated with a spec via the `// spec:x` comment.
package results

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/clems4ever/specguard/internal/lint"
)

// Set holds outcomes from one or more result files, keyed both ways so either
// correlation path can find them.
type Set struct {
	BySpec map[string]lint.TestStatus // spec id -> status (Playwright tags)
	ByTest map[string]lint.TestStatus // Go test function name -> status
	// ArtifactsBySpec holds image attachments (screenshots) keyed by spec id.
	// Path is the source filesystem path from the runner; the report command
	// copies these next to the published report and rewrites Path to a URL.
	ArtifactsBySpec map[string][]lint.Artifact
}

// New returns an empty Set.
func New() *Set {
	return &Set{
		BySpec:          map[string]lint.TestStatus{},
		ByTest:          map[string]lint.TestStatus{},
		ArtifactsBySpec: map[string][]lint.Artifact{},
	}
}

var specToken = regexp.MustCompile(`spec:([A-Za-z0-9._-]+)`)

// Load reads and merges every result file, auto-detecting its format.
func Load(paths []string) (*Set, error) {
	set := New()
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read results %q: %w", p, err)
		}
		if looksLikePlaywright(data) {
			if err := set.addPlaywright(data); err != nil {
				return nil, fmt.Errorf("parse playwright %q: %w", p, err)
			}
		} else if err := set.addGoTest(data); err != nil {
			return nil, fmt.Errorf("parse go test json %q: %w", p, err)
		}
	}
	return set, nil
}

// Apply overlays the outcomes onto a report in place: each covering test's
// Status, each spec's aggregate Result, and Report.HasResults. Aggregation is
// worst-wins — a spec is failing if any covering test failed.
func (set *Set) Apply(rep *lint.Report) {
	rep.HasResults = true
	for i := range rep.Specs {
		s := &rep.Specs[i]
		agg, seen := lint.TestStatus(""), false
		// A Playwright test tags the spec directly.
		if st, ok := set.BySpec[s.ID]; ok {
			agg, seen = worst(agg, st), true
		}
		for j := range s.Refs {
			r := &s.Refs[j]
			var st lint.TestStatus
			ok := false
			switch {
			case r.Test != "": // Go: correlate by function name
				st, ok = set.ByTest[r.Test]
			default: // Playwright reference: attribute the spec's status
				st, ok = set.BySpec[s.ID]
			}
			if ok {
				r.Status = st
				agg, seen = worst(agg, st), true
			}
		}
		if seen {
			s.Result = agg
		}
	}
}

// worst returns the more severe of two statuses: failed > passed > skipped.
func worst(a, b lint.TestStatus) lint.TestStatus {
	if a == lint.StatusFailed || b == lint.StatusFailed {
		return lint.StatusFailed
	}
	if a == lint.StatusPassed || b == lint.StatusPassed {
		return lint.StatusPassed
	}
	if a == lint.StatusSkipped || b == lint.StatusSkipped {
		return lint.StatusSkipped
	}
	if a != "" {
		return a
	}
	return b
}

func (set *Set) putSpec(id string, st lint.TestStatus) {
	set.BySpec[id] = worst(set.BySpec[id], st)
}
func (set *Set) putTest(name string, st lint.TestStatus) {
	set.ByTest[name] = worst(set.ByTest[name], st)
}

// ---- Playwright ----

type pwReport struct {
	Suites []pwSuite `json:"suites"`
}
type pwSuite struct {
	Specs  []pwSpec  `json:"specs"`
	Suites []pwSuite `json:"suites"`
}
type pwSpec struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
	Tests []pwTest `json:"tests"`
}
type pwTest struct {
	Results []pwResult `json:"results"`
}
type pwResult struct {
	Status      string         `json:"status"` // passed | failed | timedOut | skipped | interrupted
	Attachments []pwAttachment `json:"attachments"`
}
type pwAttachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Path        string `json:"path"`
	Body        string `json:"body"` // base64, when Playwright inlines the attachment
}

// source returns a filesystem path to the attachment's bytes. A `path`
// attachment is used directly; a base64 `body` (Playwright's default for a
// buffer attachment) is decoded to a temp file so the report can copy it.
func (a pwAttachment) source() (string, bool) {
	if a.Path != "" {
		return a.Path, true
	}
	if a.Body == "" {
		return "", false
	}
	raw, err := base64.StdEncoding.DecodeString(a.Body)
	if err != nil {
		return "", false
	}
	f, err := os.CreateTemp("", "specguard-shot-*"+extFor(a.ContentType))
	if err != nil {
		return "", false
	}
	defer f.Close()
	if _, err := f.Write(raw); err != nil {
		return "", false
	}
	return f.Name(), true
}

func extFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func looksLikePlaywright(data []byte) bool {
	t := bytes.TrimSpace(data)
	if len(t) == 0 || t[0] != '{' {
		return false // go test -json is newline-delimited events, not one object
	}
	var probe struct {
		Suites json.RawMessage `json:"suites"`
		Config json.RawMessage `json:"config"`
	}
	// A multi-line NDJSON stream fails to unmarshal as a single object.
	return json.Unmarshal(t, &probe) == nil && (probe.Suites != nil || probe.Config != nil)
}

func (set *Set) addPlaywright(data []byte) error {
	var rep pwReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return err
	}
	var walk func(s pwSuite)
	walk = func(s pwSuite) {
		for _, sp := range s.Specs {
			st := pwSpecStatus(sp)
			imgs := pwImages(sp)
			for _, id := range specIDs(sp.Tags, sp.Title) {
				set.putSpec(id, st)
				set.ArtifactsBySpec[id] = append(set.ArtifactsBySpec[id], imgs...)
			}
		}
		for _, child := range s.Suites {
			walk(child)
		}
	}
	for _, s := range rep.Suites {
		walk(s)
	}
	return nil
}

// pwSpecStatus aggregates a Playwright spec's test results: any hard failure →
// failed; else any pass → passed; else skipped.
func pwSpecStatus(sp pwSpec) lint.TestStatus {
	st := lint.TestStatus("")
	for _, tt := range sp.Tests {
		for _, r := range tt.Results {
			switch r.Status {
			case "failed", "timedOut", "interrupted":
				st = worst(st, lint.StatusFailed)
			case "passed", "expected":
				st = worst(st, lint.StatusPassed)
			case "skipped":
				st = worst(st, lint.StatusSkipped)
			}
		}
	}
	return st
}

// pwImages collects the image attachments (screenshots) across a spec's test
// results, keeping their source filesystem path for the report to copy.
func pwImages(sp pwSpec) []lint.Artifact {
	var out []lint.Artifact
	for _, tt := range sp.Tests {
		for _, r := range tt.Results {
			for _, a := range r.Attachments {
				if !strings.HasPrefix(a.ContentType, "image/") {
					continue
				}
				if src, ok := a.source(); ok {
					out = append(out, lint.Artifact{Name: a.Name, Path: src})
				}
			}
		}
	}
	return out
}

// specIDs extracts spec ids from Playwright tags (e.g. "@spec:auth-login") and,
// as a fallback, from the test title.
func specIDs(tags []string, title string) []string {
	var ids []string
	seen := map[string]bool{}
	add := func(s string) {
		for _, m := range specToken.FindAllStringSubmatch(s, -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				ids = append(ids, m[1])
			}
		}
	}
	for _, t := range tags {
		add(t)
	}
	add(title)
	return ids
}

// ---- Go test -json ----

type goEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
}

func (set *Set) addGoTest(data []byte) error {
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var e goEvent
		if err := json.Unmarshal(line, &e); err != nil {
			continue // tolerate interleaved non-JSON build output
		}
		if e.Test == "" {
			continue // package-level event
		}
		switch e.Action {
		case "pass":
			set.putTest(e.Test, lint.StatusPassed)
		case "fail":
			set.putTest(e.Test, lint.StatusFailed)
		case "skip":
			set.putTest(e.Test, lint.StatusSkipped)
		}
	}
	return nil
}
