// Package diff computes what changed between two specguard reports (a base
// revision and the working tree), so a reviewer reads only the delta of a change
// instead of the whole spec set. It answers: which specs were added or removed,
// which flipped coverage, which had their definition edited, and which govern
// code the change touched (and so may have silently drifted).
package diff

import (
	"sort"

	"github.com/clems4ever/specguard/internal/lint"
)

// Kind classifies one spec's change between base and head.
type Kind string

const (
	// Added: the spec exists at head but not at base.
	Added Kind = "added"
	// Removed: the spec existed at base but not at head.
	Removed Kind = "removed"
	// CoverageLost: a spec that was covered at base is uncovered at head — the
	// most important signal, a regression (unless the spec is a draft).
	CoverageLost Kind = "coverage-lost"
	// CoverageGained: an uncovered spec became covered.
	CoverageGained Kind = "coverage-gained"
	// Edited: the spec's definition (title/status/body/covers) changed while its
	// coverage did not.
	Edited Kind = "edited"
	// ImplChanged: the spec is unchanged and still covered, but the change touched
	// a file under its `covers:` — the behaviour's implementation moved, so the
	// spec is worth a human/agent glance even though its test still passes.
	ImplChanged Kind = "impl-changed"
)

// Change is one spec's delta.
type Change struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Kind   Kind   `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Draft  bool   `json:"draft"`
	// Files lists the changed files that fall under the spec's covers (ImplChanged).
	Files []string `json:"files,omitempty"`
}

// Delta is the full result: the per-spec changes plus rollup counts.
type Delta struct {
	Base    string   `json:"base"`
	Changes []Change `json:"changes"`

	Added          int `json:"added"`
	Removed        int `json:"removed"`
	CoverageLost   int `json:"coverageLost"`
	CoverageGained int `json:"coverageGained"`
	Edited         int `json:"edited"`
	ImplChanged    int `json:"implChanged"`
	// Regressions counts changes a reviewer must not miss: a non-draft spec that
	// lost coverage, or a newly-added spec that arrives uncovered.
	Regressions int  `json:"regressions"`
	BaseOK      bool `json:"baseOk"`
	HeadOK      bool `json:"headOk"`
}

// Empty reports whether nothing relevant changed.
func (d *Delta) Empty() bool { return len(d.Changes) == 0 }

// Compute diffs two reports. changedFiles are the repo-relative paths the change
// touched (base → working tree); they drive the ImplChanged signal. It is a pure
// function so it can be unit-tested without git.
func Compute(base, head *lint.Report, changedFiles []string) *Delta {
	d := &Delta{}
	if base != nil {
		d.BaseOK = base.OK
	}
	baseByID := map[string]lint.SpecStatus{}
	if base != nil {
		for _, s := range base.Specs {
			baseByID[s.ID] = s
		}
	}
	headByID := map[string]lint.SpecStatus{}
	if head != nil {
		d.HeadOK = head.OK
		for _, s := range head.Specs {
			headByID[s.ID] = s
		}
	}

	// Walk head specs: added / coverage flips / edited / impl-changed.
	for _, h := range headOrEmpty(head) {
		b, existed := baseByID[h.ID]
		switch {
		case !existed:
			d.add(Change{ID: h.ID, Title: h.Title, Kind: Added, Draft: h.Draft,
				Detail: coverageWord(h.Covered)})
			if !h.Covered && !h.Draft {
				d.Regressions++
			}
		case b.Covered && !h.Covered:
			d.add(Change{ID: h.ID, Title: h.Title, Kind: CoverageLost, Draft: h.Draft,
				Detail: "covered → uncovered"})
			if !h.Draft {
				d.Regressions++
			}
		case !b.Covered && h.Covered:
			d.add(Change{ID: h.ID, Title: h.Title, Kind: CoverageGained, Draft: h.Draft,
				Detail: "uncovered → covered"})
		case defChanged(b, h):
			d.add(Change{ID: h.ID, Title: h.Title, Kind: Edited, Draft: h.Draft,
				Detail: defChangeDetail(b, h)})
		default:
			// Unchanged definition and coverage — but did the change touch code
			// this spec governs? If so, flag it for a glance.
			if files := coveredChangedFiles(h, changedFiles); len(files) > 0 {
				d.add(Change{ID: h.ID, Title: h.Title, Kind: ImplChanged, Draft: h.Draft,
					Detail: "implementation changed", Files: files})
			}
		}
	}

	// Removed specs (in base, gone at head).
	for _, b := range baseOrEmpty(base) {
		if _, ok := headByID[b.ID]; !ok {
			d.add(Change{ID: b.ID, Title: b.Title, Kind: Removed, Draft: b.Draft})
		}
	}

	sort.SliceStable(d.Changes, func(i, j int) bool {
		pi, pj := rank(d.Changes[i].Kind), rank(d.Changes[j].Kind)
		if pi != pj {
			return pi < pj
		}
		return d.Changes[i].ID < d.Changes[j].ID
	})
	return d
}

func (d *Delta) add(c Change) {
	d.Changes = append(d.Changes, c)
	switch c.Kind {
	case Added:
		d.Added++
	case Removed:
		d.Removed++
	case CoverageLost:
		d.CoverageLost++
	case CoverageGained:
		d.CoverageGained++
	case Edited:
		d.Edited++
	case ImplChanged:
		d.ImplChanged++
	}
}

// rank orders change kinds by how much they demand attention (regressions first).
func rank(k Kind) int {
	switch k {
	case CoverageLost:
		return 0
	case Removed:
		return 1
	case Added:
		return 2
	case Edited:
		return 3
	case CoverageGained:
		return 4
	case ImplChanged:
		return 5
	}
	return 6
}

func coveredChangedFiles(s lint.SpecStatus, changed []string) []string {
	var out []string
	for _, f := range changed {
		for _, entry := range s.Covers {
			if lint.CoversMatch(entry, f) {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

func defChanged(a, b lint.SpecStatus) bool { return defChangeDetail(a, b) != "" }

func defChangeDetail(a, b lint.SpecStatus) string {
	switch {
	case a.Title != b.Title:
		return "title changed"
	case a.Status != b.Status:
		return "status " + orNone(a.Status) + " → " + orNone(b.Status)
	case a.Body != b.Body:
		return "body changed"
	case !sameStrings(a.Covers, b.Covers):
		return "covers changed"
	}
	return ""
}

func coverageWord(covered bool) string {
	if covered {
		return "new, covered"
	}
	return "new, uncovered"
}

func orNone(s string) string {
	if s == "" {
		return "active"
	}
	return s
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func headOrEmpty(r *lint.Report) []lint.SpecStatus {
	if r == nil {
		return nil
	}
	return r.Specs
}
func baseOrEmpty(r *lint.Report) []lint.SpecStatus {
	if r == nil {
		return nil
	}
	return r.Specs
}
