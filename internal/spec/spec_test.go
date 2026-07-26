package spec

import "testing"

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
