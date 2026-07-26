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

// SpecStatus is the resolved traceability state of one spec.
type SpecStatus struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status,omitempty"`
	Path     string   `json:"path"`
	Tests    []string `json:"tests"`
	Covers   []string `json:"covers,omitempty"`
	Covered  bool     `json:"covered"`
	CoversOK bool     `json:"coversOk"`
}

// Report is the full result of a run.
type Report struct {
	Specs     []SpecStatus `json:"specs"`
	Findings  []Finding    `json:"findings"`
	TestFiles int          `json:"testFiles"`
	OK        bool         `json:"ok"`
}

// refPattern matches a `spec:<id>` reference embedded in a test file — a
// Playwright tag (`@spec:skills-edit`) or a Go comment (`// spec:skills-edit`).
var refPattern = regexp.MustCompile(`spec:([A-Za-z0-9._-]+)`)

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
	refs := map[string][]string{} // spec id -> referencing test files
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
		for _, m := range refPattern.FindAllStringSubmatch(string(data), -1) {
			id := m[1]
			if _, ok := byID[id]; !ok {
				rep.Findings = append(rep.Findings, Finding{
					Severity: Error, Rule: "undefined-reference", Spec: id, File: rel,
					Message: "references spec:" + id + " but no such spec is defined",
				})
				continue
			}
			refs[id] = appendUnique(refs[id], rel)
		}
		return nil
	})
	if err != nil {
		return rep, err
	}

	// 3. Resolve each spec's status.
	for _, s := range specs {
		st := SpecStatus{
			ID: s.ID, Title: s.Title, Status: s.Status, Path: relPath(walkRoot, s.Path),
			Covers: s.Covers, Tests: refs[s.ID], Covered: len(refs[s.ID]) > 0, CoversOK: true,
		}
		sort.Strings(st.Tests)
		if !st.Covered {
			rep.Findings = append(rep.Findings, Finding{
				Severity: Error, Rule: "uncovered-spec", Spec: s.ID, File: st.Path,
				Message: "no test references spec:" + s.ID,
			})
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

func appendUnique(xs []string, x string) []string {
	for _, e := range xs {
		if e == x {
			return xs
		}
	}
	return append(xs, x)
}

func relPath(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}
