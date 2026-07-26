package lint

import "testing"

func TestGlob(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"**/*_test.go", "internal/skill/x_test.go", true},
		{"**/*_test.go", "x_test.go", true},
		{"**/*_test.go", "internal/skill/x.go", false},
		{"web/e2e/**/*.spec.ts", "web/e2e/skills.spec.ts", true},
		{"web/e2e/*.spec.ts", "web/e2e/sub/skills.spec.ts", false},
		{"internal/*/x.go", "internal/skill/x.go", true},
		{"internal/*/x.go", "internal/a/b/x.go", false},
	}
	for _, c := range cases {
		if got := compileGlob(c.pattern).match(c.path); got != c.want {
			t.Errorf("glob %q vs %q = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func TestCoversMatch(t *testing.T) {
	cases := []struct {
		entry, file string
		want        bool
	}{
		{"internal/skill", "internal/skill/catalog.go", true},
		{"internal/skill", "internal/skillother/x.go", false},
		{"internal/skill", "internal/skill", true},
		{"web/e2e/skills.spec.ts", "web/e2e/skills.spec.ts", true},
		{"web/src/**/*.tsx", "web/src/skills/SkillList.tsx", true},
	}
	for _, c := range cases {
		if got := coversMatch(c.entry, c.file); got != c.want {
			t.Errorf("coversMatch(%q, %q) = %v, want %v", c.entry, c.file, got, c.want)
		}
	}
}
