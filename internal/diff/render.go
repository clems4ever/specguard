package diff

import (
	"fmt"
	"io"
	"strings"
)

var glyphs = map[Kind]string{
	CoverageLost:   "✗",
	Removed:        "−",
	Added:          "+",
	Edited:         "~",
	CoverageGained: "✓",
	ImplChanged:    "•",
}

// RenderText writes a compact, human-readable summary — the terminal view.
func RenderText(w io.Writer, d *Delta, base string) {
	fmt.Fprintf(w, "specguard diff — vs %s\n\n", base)
	if d.Empty() {
		fmt.Fprintln(w, "No spec-relevant changes.")
		return
	}
	for _, c := range d.Changes {
		draft := ""
		if c.Draft {
			draft = " (draft)"
		}
		fmt.Fprintf(w, "  %s %-14s %s%s\n", glyphs[c.Kind], c.Kind, c.ID, draft)
		detail := c.Detail
		if len(c.Files) > 0 {
			detail = detail + ": " + strings.Join(c.Files, ", ")
		}
		if detail != "" {
			fmt.Fprintf(w, "       %s\n", detail)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, summaryLine(d))
}

// RenderMarkdown writes a PR-comment-friendly summary.
func RenderMarkdown(w io.Writer, d *Delta, base string) {
	fmt.Fprintln(w, "### specguard — spec changes")
	fmt.Fprintf(w, "_vs `%s`_\n\n", base)
	if d.Empty() {
		fmt.Fprintln(w, "No spec-relevant changes in this diff. ✅")
		return
	}
	if d.Regressions > 0 {
		fmt.Fprintf(w, "> ⚠️ **%d regression(s)** — a spec lost coverage or arrived uncovered.\n\n", d.Regressions)
	}
	fmt.Fprintln(w, "| | Change | Spec | Detail |")
	fmt.Fprintln(w, "|---|---|---|---|")
	for _, c := range d.Changes {
		detail := c.Detail
		if len(c.Files) > 0 {
			detail = detail + ": " + "`" + strings.Join(c.Files, "`, `") + "`"
		}
		id := "`" + c.ID + "`"
		if c.Draft {
			id += " _(draft)_"
		}
		fmt.Fprintf(w, "| %s | %s | %s | %s |\n", glyphs[c.Kind], c.Kind, id, detail)
	}
	fmt.Fprintf(w, "\n%s\n", summaryLine(d))
}

func summaryLine(d *Delta) string {
	parts := []string{}
	add := func(n int, label string) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, label))
		}
	}
	add(d.Added, "added")
	add(d.Removed, "removed")
	add(d.CoverageLost, "coverage-lost")
	add(d.CoverageGained, "coverage-gained")
	add(d.Edited, "edited")
	add(d.ImplChanged, "impl-changed")
	if len(parts) == 0 {
		parts = append(parts, "no changes")
	}
	verdict := "OK"
	if d.Regressions > 0 {
		verdict = fmt.Sprintf("%d REGRESSION(S)", d.Regressions)
	}
	return fmt.Sprintf("%s — %s", verdict, strings.Join(parts, ", "))
}
