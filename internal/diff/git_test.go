package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDiffAgainstGitBase drives the real git helpers: a committed base with one
// covered spec, then an uncommitted working-tree change that adds a second,
// uncovered spec. The diff must report exactly that — an added regression —
// computed from a base checkout via worktree.
func TestDiffAgainstGitBase(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "-b", "main")
	run(t, dir, "git", "config", "user.email", "t@t")
	run(t, dir, "git", "config", "user.name", "t")

	write(t, dir, ".specguard.yml", "specsDir: specs\ntests:\n  - \"**/*_test.go\"\n")
	write(t, dir, "specs/a.md", "---\nid: a\ntitle: A\ncovers:\n  - pkg\n---\nbody\n")
	write(t, dir, "pkg/a_test.go", "package pkg\n// spec:a\n")
	run(t, dir, "git", "add", "-A")
	run(t, dir, "git", "commit", "-qm", "base")

	if !HasGit(dir) {
		t.Fatal("HasGit should be true for an initialized repo")
	}

	// Working-tree change: a new spec with no covering test, and touch pkg code.
	write(t, dir, "specs/b.md", "---\nid: b\ntitle: B\n---\nbody\n")
	write(t, dir, "pkg/impl.go", "package pkg\n")

	head, err := lint.Run(lint.DefaultConfig(dir))
	if err != nil {
		t.Fatal(err)
	}
	base, err := BaseReport(dir, "HEAD", ".specguard.yml")
	if err != nil {
		t.Fatal(err)
	}
	changed, err := ChangedFiles(dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	// Base saw only spec a; head sees a + b.
	if len(base.Specs) != 1 || base.Specs[0].ID != "a" {
		t.Fatalf("base should have only spec a, got %+v", base.Specs)
	}
	d := Compute(base, head, changed)

	b := find(d, "b")
	if b == nil || b.Kind != Added {
		t.Fatalf("spec b should be added, got %+v", d.Changes)
	}
	if d.Regressions != 1 {
		t.Fatalf("new uncovered spec b is a regression: %d", d.Regressions)
	}
	// spec a: unchanged definition + coverage, but pkg/impl.go (under covers: pkg)
	// changed → impl-changed.
	a := find(d, "a")
	if a == nil || a.Kind != ImplChanged {
		t.Fatalf("spec a should be impl-changed (its covered code moved), got %+v", d.Changes)
	}
}
