package report

import "strings"

// SlugBranch maps a git branch name to a filesystem/URL-safe directory segment
// for per-branch report publishing. It percent-encodes every byte outside
// [A-Za-z0-9._-] (including "/"), so the mapping is REVERSIBLE and therefore
// collision-free by construction: distinct branch names always yield distinct
// slugs. (A lossy slug that turned both "feat/x" and "feat-x" into "feat-x"
// would silently overwrite one report with the other.)
//
//	feat/x  -> feat%2Fx
//	feat-x  -> feat-x     (distinct — no collision)
//
// Only main is published today, but baking this in keeps multi-branch safe.
func SlugBranch(name string) string {
	const upper = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(upper[c>>4])
		b.WriteByte(upper[c&0x0f])
	}
	return b.String()
}
