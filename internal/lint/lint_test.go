package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTree materializes a map of relative path -> contents under a temp dir.
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

func TestCoveredSpecPasses(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":            "---\nid: skills-edit\ntitle: Editing persists\ncovers:\n  - internal/skill\n---\nbody\n",
		"internal/skill/x.go":      "package skill\n",
		"internal/skill/x_test.go": "package skill\n// spec:skills-edit\nfunc TestX(t *testing.T){}\n",
	})
	if !rep.OK {
		t.Fatalf("expected PASS, got findings: %+v", rep.Findings)
	}
	if len(rep.Specs) != 1 || !rep.Specs[0].Covered {
		t.Fatalf("spec not marked covered: %+v", rep.Specs)
	}
}

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

func TestUndefinedReferenceFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":          "---\nid: skills-edit\ntitle: t\n---\nbody\n",
		"web/e2e/skills.spec.ts": "test('x', { tag: '@spec:skills-edit' }, ()=>{})\n",
		"web/e2e/other.spec.ts":  "test('y', { tag: '@spec:does-not-exist' }, ()=>{})\n",
	})
	if rep.OK {
		t.Fatal("expected FAIL for undefined reference")
	}
	fs := findings(rep, "undefined-reference")
	if len(fs) != 1 || fs[0].Spec != "does-not-exist" {
		t.Fatalf("want one undefined-reference for does-not-exist, got %+v", rep.Findings)
	}
}

func TestDuplicateIDFails(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/a.md":         "---\nid: dup\ntitle: A\n---\n",
		"specs/b.md":         "---\nid: dup\ntitle: B\n---\n",
		"internal/x_test.go": "// spec:dup\n",
	})
	if len(findings(rep, "duplicate-id")) != 1 {
		t.Fatalf("want one duplicate-id finding, got %+v", rep.Findings)
	}
}

func TestCoversUnmatchedWarns(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/edit.md":      "---\nid: e\ntitle: t\ncovers:\n  - internal/ghost\n---\n",
		"internal/x_test.go": "// spec:e\n",
	})
	fs := findings(rep, "covers-unmatched")
	if len(fs) != 1 || fs[0].Severity != Warning {
		t.Fatalf("want one covers-unmatched warning, got %+v", rep.Findings)
	}
	if !rep.OK {
		t.Fatal("a warning alone should still PASS")
	}
}

func TestStrictPromotesWarning(t *testing.T) {
	root := writeTree(t, map[string]string{
		"specs/edit.md":      "---\nid: e\ntitle: t\ncovers:\n  - internal/ghost\n---\n",
		"internal/x_test.go": "// spec:e\n",
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

func TestSpecFilesNotScannedAsTests(t *testing.T) {
	// A stray token in the spec body must not count as coverage.
	rep := run(t, map[string]string{
		"specs/edit.test.ts.md": "---\nid: e\ntitle: t\n---\nsee spec:e in prose\n",
	})
	if rep.OK {
		t.Fatal("spec body reference must not satisfy coverage")
	}
}
