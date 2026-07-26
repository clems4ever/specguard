package lint

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Render writes a human-readable summary of the report to w. When color is
// true, status glyphs are ANSI-colored.
func Render(w io.Writer, rep *Report, color bool) {
	green := colorizer("\033[32m", color)
	red := colorizer("\033[31m", color)
	yellow := colorizer("\033[33m", color)
	dim := colorizer("\033[2m", color)

	fmt.Fprintln(w, "specguard — spec traceability")
	fmt.Fprintln(w)

	// Per-spec coverage table.
	for _, s := range rep.Specs {
		glyph := green("✓")
		if !s.Covered {
			glyph = red("✗")
		} else if !s.CoversOK {
			glyph = yellow("!")
		}
		fmt.Fprintf(w, "  %s  %-28s %s\n", glyph, s.ID, s.Title)
		if s.Covered {
			fmt.Fprintf(w, "        %s\n", dim(fmt.Sprintf("%d test(s): %s", len(s.Tests), strings.Join(s.Tests, ", "))))
		} else {
			fmt.Fprintf(w, "        %s\n", dim("no covering test"))
		}
	}
	if len(rep.Specs) == 0 {
		fmt.Fprintln(w, dim("  (no specs found)"))
	}
	fmt.Fprintln(w)

	// Findings, errors first.
	findings := append([]Finding(nil), rep.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Severity == Error && findings[j].Severity != Error
	})
	errs, warns := 0, 0
	for _, f := range findings {
		tag := red("error")
		if f.Severity == Warning {
			tag = yellow("warning")
			warns++
		} else {
			errs++
		}
		loc := f.File
		if f.Spec != "" {
			loc = f.Spec
		}
		fmt.Fprintf(w, "  %s  [%s] %s\n", tag, f.Rule, f.Message)
		if loc != "" {
			fmt.Fprintf(w, "        %s\n", dim(f.File))
		}
	}
	if len(findings) > 0 {
		fmt.Fprintln(w)
	}

	summary := fmt.Sprintf("%d spec(s), %d test file(s) — %d error(s), %d warning(s)",
		len(rep.Specs), rep.TestFiles, errs, warns)
	if rep.OK {
		fmt.Fprintln(w, green("PASS")+" "+summary)
	} else {
		fmt.Fprintln(w, red("FAIL")+" "+summary)
	}
}

func colorizer(code string, on bool) func(string) string {
	if !on {
		return func(s string) string { return s }
	}
	return func(s string) string { return code + s + "\033[0m" }
}
