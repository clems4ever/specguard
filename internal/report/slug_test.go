package report

import "testing"

func TestSlugBranchExamples(t *testing.T) {
	cases := map[string]string{
		"main":        "main",
		"feat/x":      "feat%2Fx",
		"feat-x":      "feat-x",
		"release/1.2": "release%2F1.2",
		"a_b.c-d":     "a_b.c-d",
		"user/fix#42": "user%2Ffix%2342",
		"caps/UPPER":  "caps%2FUPPER",
	}
	for in, want := range cases {
		if got := SlugBranch(in); got != want {
			t.Errorf("SlugBranch(%q) = %q, want %q", in, got, want)
		}
	}
}

// The whole point of a reversible encoding: two branch names that a lossy slug
// would collapse together must stay distinct, so one report can never overwrite
// another.
// spec:report-branch-slug
func TestSlugBranchNoCollision(t *testing.T) {
	pairs := [][2]string{
		{"feat/x", "feat-x"},
		{"a/b", "a%2Fb"}, // literal '%' in a branch name still can't alias
		{"x/y", "x-y"},
	}
	for _, p := range pairs {
		if SlugBranch(p[0]) == SlugBranch(p[1]) {
			t.Errorf("collision: %q and %q both slug to %q", p[0], p[1], SlugBranch(p[0]))
		}
	}
}
