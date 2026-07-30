package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture expands the %SPEC% placeholder used in this file's test data to the
// real reference token. The placeholder keeps the literal token out of the Go
// source, so specguard doesn't read its own fixtures as real references when it
// lints itself (dogfooding).
func fixture(s string) string { return strings.ReplaceAll(s, "%SPEC%", "spec:") }

// writeTree materializes a map of relative path -> contents under a temp dir.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(fixture(content)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func run(t *testing.T, files map[string]string) *Report {
	t.Helper()
	root := writeTree(t, files)
	cfg := DefaultConfig(root)
	rep, err := Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

func findings(rep *Report, rule string) []Finding {
	var out []Finding
	for _, f := range rep.Findings {
		if f.Rule == rule {
			out = append(out, f)
		}
	}
	return out
}

// spec:lint-covered-passes
func TestCoveredSpecPasses(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":            "---\nid: skills-edit\ntitle: Editing persists\ncovers:\n  - internal/skill\n---\nbody\n",
		"internal/skill/x.go":      "package skill\n",
		"internal/skill/x_test.go": "package skill\n// %SPEC%skills-edit\nfunc TestX(t *testing.T){}\n",
	})
	if !rep.OK {
		t.Fatalf("expected PASS, got findings: %+v", rep.Findings)
	}
	if len(rep.Specs) != 1 || !rep.Specs[0].Covered {
		t.Fatalf("spec not marked covered: %+v", rep.Specs)
	}
}

// spec:lint-uncovered-is-error
func TestUncoveredSpecFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":            "---\nid: skills-edit\ntitle: Editing persists\n---\nbody\n",
		"internal/skill/x_test.go": "package skill\nfunc TestX(t *testing.T){}\n",
	})
	if rep.OK {
		t.Fatal("expected FAIL for uncovered spec")
	}
	if len(findings(rep, "uncovered-spec")) != 1 {
		t.Fatalf("want one uncovered-spec finding, got %+v", rep.Findings)
	}
}

// spec:lint-draft-uncovered-warns
func TestUncoveredDraftWarnsNotFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/planned.md":   "---\nid: planned\ntitle: Planned\nstatus: draft\n---\nbody\n",
		"internal/x_test.go": "package x\n",
	})
	if !rep.OK {
		t.Fatalf("an uncovered draft should still PASS, got %+v", rep.Findings)
	}
	fs := findings(rep, "uncovered-draft")
	if len(fs) != 1 || fs[0].Severity != Warning {
		t.Fatalf("want one uncovered-draft warning, got %+v", rep.Findings)
	}
	if len(findings(rep, "uncovered-spec")) != 0 {
		t.Fatalf("draft must not raise uncovered-spec, got %+v", rep.Findings)
	}
	if !rep.Specs[0].Draft {
		t.Fatal("spec should be marked Draft")
	}
}

func TestCoveredDraftIsClean(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/planned.md":   "---\nid: planned\ntitle: Planned\nstatus: draft\n---\nbody\n",
		"internal/x_test.go": "// %SPEC%planned\n",
	})
	if !rep.OK || len(rep.Findings) != 0 {
		t.Fatalf("a covered draft should be clean, got %+v", rep.Findings)
	}
}

func TestSpecPathIsCleanRelative(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/sharing/perms.md": "---\nid: p\ntitle: t\n---\nbody\n",
		"internal/x_test.go":     "// %SPEC%p\n",
	})
	if got := rep.Specs[0].Path; got != "specs/sharing/perms.md" {
		t.Fatalf("spec path = %q, want clean relative path (no '..')", got)
	}
}

func TestBodyIsCarried(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/e.md":         "---\nid: e\ntitle: t\n---\n## Why\nbecause\n",
		"internal/x_test.go": "// %SPEC%e\n",
	})
	if rep.Specs[0].Body != "## Why\nbecause\n" {
		t.Fatalf("body not carried: %q", rep.Specs[0].Body)
	}
}

// spec:lint-undefined-reference
func TestUndefinedReferenceFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":          "---\nid: skills-edit\ntitle: t\n---\nbody\n",
		"web/e2e/skills.spec.ts": "test('x', { tag: '@%SPEC%skills-edit' }, ()=>{})\n",
		"web/e2e/other.spec.ts":  "test('y', { tag: '@%SPEC%does-not-exist' }, ()=>{})\n",
	})
	if rep.OK {
		t.Fatal("expected FAIL for undefined reference")
	}
	fs := findings(rep, "undefined-reference")
	if len(fs) != 1 || fs[0].Spec != "does-not-exist" {
		t.Fatalf("want one undefined-reference for does-not-exist, got %+v", rep.Findings)
	}
}

// spec:lint-duplicate-id
func TestDuplicateIDFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/a.md":         "---\nid: dup\ntitle: A\n---\n",
		"specs/b.md":         "---\nid: dup\ntitle: B\n---\n",
		"internal/x_test.go": "// %SPEC%dup\n",
	})
	if len(findings(rep, "duplicate-id")) != 1 {
		t.Fatalf("want one duplicate-id finding, got %+v", rep.Findings)
	}
}

// spec:lint-covers-unmatched
func TestCoversUnmatchedWarns(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":      "---\nid: e\ntitle: t\ncovers:\n  - internal/ghost\n---\n",
		"internal/x_test.go": "// %SPEC%e\n",
	})
	fs := findings(rep, "covers-unmatched")
	if len(fs) != 1 || fs[0].Severity != Warning {
		t.Fatalf("want one covers-unmatched warning, got %+v", rep.Findings)
	}
	if !rep.OK {
		t.Fatal("a warning alone should still PASS")
	}
}

// spec:lint-strict-promotes
func TestStrictPromotesWarning(t *testing.T) {
	root := writeTree(t, map[string]string{
		"specs/edit.md":      "---\nid: e\ntitle: t\ncovers:\n  - internal/ghost\n---\n",
		"internal/x_test.go": "// %SPEC%e\n",
	})
	cfg := DefaultConfig(root)
	cfg.Strict = true
	rep, err := Run(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if rep.OK {
		t.Fatal("strict mode should FAIL on covers-unmatched")
	}
}

// spec:lint-spec-body-not-coverage
func TestSpecFilesNotScannedAsTests(t *testing.T) {
	// A stray token in the spec body must not count as coverage.
	rep := run(t, map[string]string{
		"specs/edit.test.ts.md": "---\nid: e\ntitle: t\n---\nsee %SPEC%e in prose\n",
	})
	if rep.OK {
		t.Fatal("spec body reference must not satisfy coverage")
	}
}

// spec:lint-ref-locations
func TestRefsCaptureFileAndLine(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md": "---\nid: skills-edit\ntitle: Editing persists\n---\nbody\n",
		// the first reference sits on line 3, a second on line 5.
		"internal/skill/x_test.go": "package skill\n\n// %SPEC%skills-edit\nfunc TestA(t *testing.T){}\n// %SPEC%skills-edit\nfunc TestB(t *testing.T){}\n",
		"web/e2e/x.spec.ts":        "import {test} from '@playwright/test';\ntest('edit', { tag: '@%SPEC%skills-edit' }, async () => {});\n",
	})
	if len(rep.Specs) != 1 {
		t.Fatalf("want 1 spec, got %d", len(rep.Specs))
	}
	got := rep.Specs[0].Refs
	want := []Ref{
		// Go refs also carry the test function the comment sits above.
		{File: "internal/skill/x_test.go", Line: 3, Test: "TestA"},
		{File: "internal/skill/x_test.go", Line: 5, Test: "TestB"},
		{File: "web/e2e/x.spec.ts", Line: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("refs = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ref[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	// Tests stays the distinct-file list (for the count).
	if len(rep.Specs[0].Tests) != 2 {
		t.Errorf("distinct test files = %v, want 2", rep.Specs[0].Tests)
	}
}

// spec:ui-area-overview
func TestAreaOverviewLoadedAndNotASpec(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/ui/dashboard.md": "---\nid: ui-dashboard\ntitle: Dashboard\ncovers:\n  - web\n---\nbody\n",
		"specs/ui/_area.md":     "---\ntitle: UI\n---\nThe report's own interface.\n",
		"web/e2e/x.spec.ts":     "// %SPEC%ui-dashboard\ntest('x', () => {})\n",
	})
	if !rep.OK {
		t.Fatalf("expected PASS, got findings: %+v", rep.Findings)
	}
	// The _area.md is not counted as a spec.
	if len(rep.Specs) != 1 {
		t.Fatalf("want 1 spec (the _area.md excluded), got %d: %+v", len(rep.Specs), rep.Specs)
	}
	// Its overview is loaded, keyed by the directory name.
	if len(rep.Areas) != 1 || rep.Areas[0].Name != "ui" || rep.Areas[0].Title != "UI" {
		t.Fatalf("area overview not loaded: %+v", rep.Areas)
	}
	if rep.Areas[0].Description != "The report's own interface." {
		t.Fatalf("description = %q", rep.Areas[0].Description)
	}
}
