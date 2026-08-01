package spec

import "testing"

// spec:spec-frontmatter
func TestParseValid(t *testing.T) {
	s, err := Parse([]byte("---\nid: skills-edit\ntitle: Editing persists\nstatus: active\ncovers:\n  - internal/skill\n  - web/e2e/skills.spec.ts\n---\n\n# body\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "skills-edit" || s.Title != "Editing persists" || s.Status != "active" {
		t.Fatalf("bad scalars: %+v", s)
	}
	if len(s.Covers) != 2 || s.Covers[0] != "internal/skill" || s.Covers[1] != "web/e2e/skills.spec.ts" {
		t.Fatalf("bad covers: %+v", s.Covers)
	}
}

func TestParseCapturesBody(t *testing.T) {
	s, err := Parse([]byte("---\nid: x\ntitle: t\n---\n\n## Behaviour\n\nsome prose\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Body != "## Behaviour\n\nsome prose\n" {
		t.Fatalf("body = %q", s.Body)
	}
}

func TestParseDraftStatus(t *testing.T) {
	s, err := Parse([]byte("---\nid: x\ntitle: t\nstatus: draft\n---\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != StatusDraft {
		t.Fatalf("status = %q", s.Status)
	}
}

func TestParseBadStatus(t *testing.T) {
	if _, err := Parse([]byte("---\nid: x\ntitle: t\nstatus: wip\n---\n")); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// spec:spec-frontmatter-invalid
func TestParseMissingID(t *testing.T) {
	if _, err := Parse([]byte("---\ntitle: No id\n---\n")); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestParseBadID(t *testing.T) {
	if _, err := Parse([]byte("---\nid: has spaces\ntitle: x\n---\n")); err == nil {
		t.Fatal("expected error for invalid id")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	if _, err := Parse([]byte("# just markdown\n")); err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParseUnterminated(t *testing.T) {
	if _, err := Parse([]byte("---\nid: x\ntitle: y\n")); err == nil {
		t.Fatal("expected error for unterminated frontmatter")
	}
}

func TestParseFieldsQuotesAndComments(t *testing.T) {
	f, err := ParseFields("# comment\nid: \"quoted\"\ntests:\n  - 'a'\n  - b\n")
	if err != nil {
		t.Fatal(err)
	}
	if f["id"][0] != "quoted" {
		t.Fatalf("unquote failed: %+v", f["id"])
	}
	if len(f["tests"]) != 2 || f["tests"][0] != "a" || f["tests"][1] != "b" {
		t.Fatalf("bad list: %+v", f["tests"])
	}
}

// spec:ui-area-overview
func TestParseAreaFrontmatterAndBody(t *testing.T) {
	a, err := ParseArea([]byte("---\ntitle: Lint\n---\nThe core rules.\nSecond line.\n\nIgnored paragraph.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != "Lint" {
		t.Fatalf("title = %q", a.Title)
	}
	// The description is the first paragraph, joined to one line.
	if a.Description != "The core rules. Second line." {
		t.Fatalf("description = %q", a.Description)
	}
}

func TestParseAreaExplicitDescriptionWins(t *testing.T) {
	a, err := ParseArea([]byte("---\ntitle: UI\ndescription: A short one.\n---\nBody paragraph.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Description != "A short one." {
		t.Fatalf("description = %q", a.Description)
	}
}

func TestParseAreaNoFrontmatter(t *testing.T) {
	a, err := ParseArea([]byte("Just prose describing the area.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != "" || a.Description != "Just prose describing the area." {
		t.Fatalf("got %+v", a)
	}
}

func TestParsePreview(t *testing.T) {
	s, err := Parse([]byte("---\nid: x\ntitle: y\npreview: /login\n---\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Preview != "/login" {
		t.Fatalf("preview = %q", s.Preview)
	}
}
