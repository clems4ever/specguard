package diff

import (
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func rep(ok bool, specs ...lint.SpecStatus) *lint.Report {
	return &lint.Report{OK: ok, Specs: specs}
}
func spec(id string, covered bool, opts ...func(*lint.SpecStatus)) lint.SpecStatus {
	s := lint.SpecStatus{ID: id, Title: id, Covered: covered}
	for _, o := range opts {
		o(&s)
	}
	return s
}
func covers(c ...string) func(*lint.SpecStatus) { return func(s *lint.SpecStatus) { s.Covers = c } }
func draft() func(*lint.SpecStatus)             { return func(s *lint.SpecStatus) { s.Draft = true } }
func title(t string) func(*lint.SpecStatus)     { return func(s *lint.SpecStatus) { s.Title = t } }

func find(d *Delta, id string) *Change {
	for i := range d.Changes {
		if d.Changes[i].ID == id {
			return &d.Changes[i]
		}
	}
	return nil
}

func TestAddedAndRemoved(t *testing.T) {
	base := rep(true, spec("a", true))
	head := rep(true, spec("a", true), spec("b", true))
	d := Compute(base, head, nil)
	if c := find(d, "b"); c == nil || c.Kind != Added {
		t.Fatalf("b should be added: %+v", d.Changes)
	}
	// Remove a, keep b.
	d = Compute(rep(true, spec("a", true), spec("b", true)), rep(true, spec("b", true)), nil)
	if c := find(d, "a"); c == nil || c.Kind != Removed {
		t.Fatalf("a should be removed: %+v", d.Changes)
	}
}

func TestCoverageLostIsRegression(t *testing.T) {
	base := rep(true, spec("a", true))
	head := rep(false, spec("a", false))
	d := Compute(base, head, nil)
	c := find(d, "a")
	if c == nil || c.Kind != CoverageLost {
		t.Fatalf("expected coverage-lost, got %+v", d.Changes)
	}
	if d.Regressions != 1 {
		t.Fatalf("coverage-lost on a non-draft must count as a regression: %d", d.Regressions)
	}
}

func TestDraftCoverageLossIsNotRegression(t *testing.T) {
	base := rep(true, spec("a", true, draft()))
	head := rep(true, spec("a", false, draft()))
	d := Compute(base, head, nil)
	if d.Regressions != 0 {
		t.Fatalf("a draft losing coverage is not a regression: %d", d.Regressions)
	}
	if c := find(d, "a"); c == nil || c.Kind != CoverageLost {
		t.Fatalf("still reported as coverage-lost: %+v", d.Changes)
	}
}

func TestNewUncoveredSpecIsRegression(t *testing.T) {
	d := Compute(rep(true), rep(false, spec("a", false)), nil)
	c := find(d, "a")
	if c == nil || c.Kind != Added || d.Regressions != 1 {
		t.Fatalf("a new uncovered spec is an added regression: %+v regs=%d", d.Changes, d.Regressions)
	}
}

func TestEditedDefinition(t *testing.T) {
	base := rep(true, spec("a", true, title("Old")))
	head := rep(true, spec("a", true, title("New")))
	d := Compute(base, head, nil)
	c := find(d, "a")
	if c == nil || c.Kind != Edited || c.Detail != "title changed" {
		t.Fatalf("expected edited/title changed, got %+v", d.Changes)
	}
}

// spec:diff-impl-changed
func TestImplChangedWhenCoveredCodeTouched(t *testing.T) {
	s := spec("a", true, covers("internal/hub"))
	base := rep(true, s)
	head := rep(true, s)
	// A file under the spec's covers changed, but the spec + coverage did not.
	d := Compute(base, head, []string{"internal/hub/archive.go", "docs/readme.md"})
	c := find(d, "a")
	if c == nil || c.Kind != ImplChanged {
		t.Fatalf("expected impl-changed, got %+v", d.Changes)
	}
	if len(c.Files) != 1 || c.Files[0] != "internal/hub/archive.go" {
		t.Fatalf("impl-changed should list only the covered file: %+v", c.Files)
	}
	if d.Regressions != 0 {
		t.Fatal("impl-changed is not a regression")
	}
}

func TestImplChangeYieldsToRealChange(t *testing.T) {
	// If a spec both lost coverage AND its code changed, the regression wins.
	base := rep(true, spec("a", true, covers("internal/hub")))
	head := rep(false, spec("a", false, covers("internal/hub")))
	d := Compute(base, head, []string{"internal/hub/x.go"})
	c := find(d, "a")
	if c == nil || c.Kind != CoverageLost {
		t.Fatalf("coverage-lost must take precedence over impl-changed: %+v", d.Changes)
	}
}

func TestUnchangedProducesEmptyDelta(t *testing.T) {
	s := spec("a", true, covers("internal/hub"))
	d := Compute(rep(true, s), rep(true, s), []string{"web/app.ts"})
	if !d.Empty() {
		t.Fatalf("no relevant change expected, got %+v", d.Changes)
	}
}

func TestRegressionsSortFirst(t *testing.T) {
	base := rep(true, spec("keep", true), spec("lose", true))
	head := rep(false, spec("keep", true, title("edited")), spec("lose", false), spec("new", true))
	d := Compute(base, head, nil)
	if d.Changes[0].Kind != CoverageLost {
		t.Fatalf("the regression should sort first, got %s", d.Changes[0].Kind)
	}
}
