// Package lint implements specguard's traceability check: every spec must be
// referenced by at least one test, and every `spec:<id>` reference in a test
// must point at a spec that exists.
package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/clems4ever/specguard/internal/spec"
)

// Config controls a lint run. It is normally loaded from `.specguard.yml`.
type Config struct {
	// Root is the directory the run is anchored at. All globs and reported
	// paths are relative to it.
	Root string
	// SpecsDir is the directory (relative to Root) holding spec markdown files.
	SpecsDir string
	// Tests are globs (relative to Root) selecting the test files scanned for
	// `spec:<id>` references.
	Tests []string
	// Strict promotes warnings (e.g. a `covers` entry that matches no file) to
	// errors.
	Strict bool
}

// DefaultConfig is used when no config file is present.
func DefaultConfig(root string) Config {
	return Config{
		Root:     root,
		SpecsDir: "specs",
		Tests:    []string{"**/*_test.go", "**/*.spec.ts", "**/*.test.ts", "**/*.test.tsx"},
	}
}

// LoadConfig reads a `.specguard.yml`-style config, falling back to defaults
// for any unset field. A missing file is not an error.
func LoadConfig(root, path string) (Config, error) {
	cfg := DefaultConfig(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	fields, err := spec.ParseFields(string(data))
	if err != nil {
		return cfg, err
	}
	if v := fields["specsDir"]; len(v) > 0 {
		cfg.SpecsDir = v[0]
	}
	if v := fields["tests"]; len(v) > 0 {
		cfg.Tests = v
	}
	return cfg, nil
}

// Severity of a finding.
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

// Finding is a single problem discovered during a run.
type Finding struct {
	Severity Severity `json:"severity"`
	Rule     string   `json:"rule"`
	Spec     string   `json:"spec,omitempty"`
	File     string   `json:"file,omitempty"`
	Message  string   `json:"message"`
}

// TestStatus is the outcome of a test run, set by `specguard report --results`.
// The empty value means "no result ingested".
type TestStatus string

const (
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
)

// Ref is one `spec:<id>` reference: the test file and the 1-based line it sits
// on, so the report can link straight to the covering test on GitHub.
type Ref struct {
	File string `json:"file"`
	Line int    `json:"line"`
	// Test is the Go test function this reference sits above (e.g. "TestLogin"),
	// used to correlate `go test -json` results. Empty for non-Go references.
	Test string `json:"test,omitempty"`
	// Status is this test's outcome, filled in when results are ingested.
	Status TestStatus `json:"status,omitempty"`
}

// SpecStatus is the resolved traceability state of one spec.
type SpecStatus struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status,omitempty"`
	Path   string `json:"path"`
	Body   string `json:"body,omitempty"`
	// Tests are the distinct test files that reference this spec (kept for the
	// count); Refs are the precise (file, line) locations, for deep links.
	Tests    []string `json:"tests"`
	Refs     []Ref    `json:"refs,omitempty"`
	Covers   []string `json:"covers,omitempty"`
	Covered  bool     `json:"covered"`
	CoversOK bool     `json:"coversOk"`
	Draft    bool     `json:"draft"`
	// Result is the aggregate outcome of this spec's covering tests, set when
	// results are ingested (failed if any covering test failed).
	Result TestStatus `json:"result,omitempty"`
}

// Report is the full result of a run.
type Report struct {
	Specs     []SpecStatus `json:"specs"`
	Findings  []Finding    `json:"findings"`
	TestFiles int          `json:"testFiles"`
	OK        bool         `json:"ok"`
	// HasResults is true when a test run was ingested, so the UI knows to show
	// pass/fail state rather than coverage alone.
	HasResults bool `json:"hasResults,omitempty"`
}

// refPattern matches a `spec:<id>` reference embedded in a test file — a
// Playwright tag (`@spec:skills-edit`) or a Go comment (`// spec:skills-edit`).
var refPattern = regexp.MustCompile(`spec:([A-Za-z0-9._-]+)`)

// funcPattern matches a Go test/benchmark/example function declaration, so a
// `// spec:x` comment can be associated with the function it sits above.
var funcPattern = regexp.MustCompile(`^func ((?:Test|Benchmark|Example)[A-Za-z0-9_]*)\s*\(`)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	"dist": true, "build": true, ".next": true,
}

// Run performs the traceability check.
func Run(cfg Config) (*Report, error) {
	rep := &Report{OK: true}
	testGlobs := make([]glob, len(cfg.Tests))
	for i, g := range cfg.Tests {
		testGlobs[i] = compileGlob(g)
	}

	// 1. Parse every spec file.
	specs, findings := loadSpecs(cfg)
	rep.Findings = append(rep.Findings, findings...)
	byID := map[string]*spec.Spec{}
	for _, s := range specs {
		byID[s.ID] = s
	}

	// 2. Walk the tree once: scan test files for references, and collect all
	// file paths so `covers` entries can be validated.
	refs := map[string][]Ref{} // spec id -> (file, line) references
	var allFiles []string
	specsPrefix := filepath.ToSlash(filepath.Clean(cfg.SpecsDir)) + "/"
	walkRoot := cfg.Root
	if walkRoot == "" {
		walkRoot = "."
	}
	err := filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != walkRoot && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(walkRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		allFiles = append(allFiles, rel)
		if !matchesAny(testGlobs, rel) {
			return nil
		}
		// Don't treat spec files as tests even if a glob is broad.
		if strings.HasPrefix(rel, specsPrefix) {
			return nil
		}
		rep.TestFiles++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Scan line by line so each reference carries its 1-based line number.
		lines := strings.Split(string(data), "\n")
		isGo := strings.HasSuffix(rel, "_test.go")
		for i, line := range lines {
			for _, m := range refPattern.FindAllStringSubmatch(line, -1) {
				id := m[1]
				if _, ok := byID[id]; !ok {
					rep.Findings = append(rep.Findings, Finding{
						Severity: Error, Rule: "undefined-reference", Spec: id, File: rel,
						Message: "references spec:" + id + " but no such spec is defined",
					})
					continue
				}
				ref := Ref{File: rel, Line: i + 1}
				if isGo {
					ref.Test = nextTestFunc(lines, i)
				}
				refs[id] = append(refs[id], ref)
			}
		}
		return nil
	})
	if err != nil {
		return rep, err
	}

	// 3. Resolve each spec's status.
	for _, s := range specs {
		draft := s.Status == spec.StatusDraft
		specRefs := sortedRefs(refs[s.ID])
		st := SpecStatus{
			// s.Path is already relative to cfg.Root (set in loadSpecs).
			ID: s.ID, Title: s.Title, Status: s.Status, Path: s.Path,
			Body: s.Body, Covers: s.Covers,
			Tests: distinctFiles(specRefs), Refs: specRefs,
			Covered: len(specRefs) > 0, CoversOK: true, Draft: draft,
		}
		if !st.Covered {
			// A draft spec is allowed to have no test yet — it is a planned
			// behaviour, reported as a warning rather than a build failure.
			if draft {
				rep.Findings = append(rep.Findings, Finding{
					Severity: Warning, Rule: "uncovered-draft", Spec: s.ID, File: st.Path,
					Message: "draft spec has no covering test yet",
				})
			} else {
				rep.Findings = append(rep.Findings, Finding{
					Severity: Error, Rule: "uncovered-spec", Spec: s.ID, File: st.Path,
					Message: "no test references spec:" + s.ID,
				})
			}
		}
		for _, entry := range s.Covers {
			if !anyFileMatches(entry, allFiles) {
				st.CoversOK = false
				sev := Warning
				if cfg.Strict {
					sev = Error
				}
				rep.Findings = append(rep.Findings, Finding{
					Severity: sev, Rule: "covers-unmatched", Spec: s.ID, File: st.Path,
					Message: "covers entry " + entry + " matches no file",
				})
			}
		}
		rep.Specs = append(rep.Specs, st)
	}
	sort.Slice(rep.Specs, func(i, j int) bool { return rep.Specs[i].ID < rep.Specs[j].ID })

	for _, f := range rep.Findings {
		if f.Severity == Error {
			rep.OK = false
		}
	}
	return rep, nil
}

// loadSpecs parses every *.md under the specs dir and flags duplicate ids.
func loadSpecs(cfg Config) ([]*spec.Spec, []Finding) {
	var specs []*spec.Spec
	var findings []Finding
	dir := filepath.Join(cfg.Root, cfg.SpecsDir)
	seen := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel := relPath(cfg.Root, path)
		s, perr := spec.ParseFile(path)
		if perr != nil {
			findings = append(findings, Finding{
				Severity: Error, Rule: "parse-error", File: rel, Message: perr.Error(),
			})
			return nil
		}
		s.Path = rel
		if prev, ok := seen[s.ID]; ok {
			findings = append(findings, Finding{
				Severity: Error, Rule: "duplicate-id", Spec: s.ID, File: rel,
				Message: "duplicate spec id (already defined in " + prev + ")",
			})
			return nil
		}
		seen[s.ID] = rel
		specs = append(specs, s)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		findings = append(findings, Finding{
			Severity: Error, Rule: "specs-dir", File: cfg.SpecsDir, Message: err.Error(),
		})
	}
	return specs, findings
}

func anyFileMatches(entry string, files []string) bool {
	for _, f := range files {
		if coversMatch(entry, f) {
			return true
		}
	}
	return false
}

// nextTestFunc returns the name of the first Go test function at or below line
// index `from` (0-based), i.e. the function a `// spec:x` comment sits above.
// Empty if none follows within the file.
func nextTestFunc(lines []string, from int) string {
	for i := from; i < len(lines); i++ {
		if m := funcPattern.FindStringSubmatch(lines[i]); m != nil {
			return m[1]
		}
	}
	return ""
}

// sortedRefs returns refs ordered by file then line, so link lists are stable.
func sortedRefs(refs []Ref) []Ref {
	if len(refs) == 0 {
		return nil
	}
	out := append([]Ref(nil), refs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out
}

// distinctFiles returns the unique test files among refs, in order.
func distinctFiles(refs []Ref) []string {
	var files []string
	seen := map[string]bool{}
	for _, r := range refs {
		if !seen[r.File] {
			seen[r.File] = true
			files = append(files, r.File)
		}
	}
	return files
}

func relPath(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}
