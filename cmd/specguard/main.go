// Command specguard checks spec traceability: every spec markdown file must be
// referenced by at least one test, and every `spec:<id>` reference in a test
// must resolve to a defined spec. It exits non-zero when the check fails, so it
// can gate a CI build.
//
// Usage:
//
//	specguard [flags]         run the check and print a report (exit 1 on failure)
//	specguard serve [flags]   serve the report over HTTP for the web UI
//	specguard diff [flags]    show only the specs a change touched (vs a base ref)
//	specguard report [flags]  write a self-contained, browsable HTML report
//	specguard accept [flags]  record a PM's behavioural sign-off, or gate on it
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clems4ever/specguard/internal/diff"
	"github.com/clems4ever/specguard/internal/lint"
	"github.com/clems4ever/specguard/internal/report"
	"github.com/clems4ever/specguard/internal/results"
	"github.com/clems4ever/specguard/internal/server"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			serveCmd(os.Args[2:])
			return
		case "diff":
			diffCmd(os.Args[2:])
			return
		case "report":
			reportCmd(os.Args[2:])
			return
		case "accept":
			acceptCmd(os.Args[2:])
			return
		}
	}
	lintCmd(os.Args[1:])
}

// needsReview is true for a spec that carries a direct test (a behaviour a PM
// should sign off) but is not accepted at its current fingerprint. Test-less
// parent specs are verified through their children and are not gated directly.
func needsReview(s lint.SpecStatus) bool {
	return s.Covered && (s.Lifecycle == lint.LifeImplemented || s.Lifecycle == lint.LifeStale)
}

// acceptCmd records a PM's behavioural acceptance, or gates a build on it.
//
//	specguard accept <id> [--by ..] [--evidence URL]   accept one spec at its current fingerprint
//	specguard accept --all [--by ..]                    accept every spec awaiting review
//	specguard accept --check                            exit non-zero if any spec awaits review
func acceptCmd(args []string) {
	fs := flag.NewFlagSet("specguard accept", flag.ExitOnError)
	var (
		root       = fs.String("C", ".", "directory to run in (repo root)")
		configPath = fs.String("config", "", "config file (default: <root>/.specguard.yml)")
		all        = fs.Bool("all", false, "accept every spec currently awaiting review")
		check      = fs.Bool("check", false, "exit non-zero if any spec awaits PM review (a CI gate)")
		by         = fs.String("by", "", "who is accepting (reviewer identity)")
		evidence   = fs.String("evidence", "", "link to the demo / preview the behaviour was reviewed in")
	)
	_ = fs.Parse(args)

	cfg := loadConfig(*root, *configPath, false)
	rep, err := lint.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard accept:", err)
		os.Exit(2)
	}

	var pending []lint.SpecStatus
	for _, s := range rep.Specs {
		if needsReview(s) {
			pending = append(pending, s)
		}
	}

	// Gate mode: fail the build while any behaviour lacks a current sign-off.
	if *check {
		if len(pending) == 0 {
			fmt.Printf("specguard: all implemented specs are accepted at their current fingerprint.\n")
			return
		}
		fmt.Printf("specguard: %d spec(s) await PM behavioural review:\n", len(pending))
		for _, s := range pending {
			note := "needs review"
			if s.Lifecycle == lint.LifeStale {
				note = "behaviour changed since acceptance — re-review"
			}
			fmt.Printf("  • %-28s %s (%s)\n", s.ID, s.Title, note)
		}
		fmt.Printf("\nA reviewer records acceptance with: specguard accept <id> --by <you>\n")
		os.Exit(1)
	}

	at := time.Now().UTC().Format(time.RFC3339)
	record := func(s lint.SpecStatus) {
		if err := lint.AppendAcceptance(*root, lint.Acceptance{
			Spec: s.ID, Fingerprint: s.Fingerprint, Verdict: "accepted",
			By: *by, At: at, Evidence: *evidence,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "specguard accept:", err)
			os.Exit(2)
		}
		fmt.Printf("accepted %s @ %s\n", s.ID, s.Fingerprint)
	}

	if *all {
		if len(pending) == 0 {
			fmt.Println("specguard: nothing to accept — all implemented specs already accepted.")
			return
		}
		for _, s := range pending {
			record(s)
		}
		return
	}

	id := fs.Arg(0)
	if id == "" {
		fmt.Fprintln(os.Stderr, "specguard accept: give a spec id, or --all, or --check")
		os.Exit(2)
	}
	// Go's flag parser stops at the first positional, so trailing flags would be
	// silently dropped — guard against `accept <id> --by X` (flags must lead).
	if fs.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "specguard accept: put flags before the spec id, e.g. accept --by you <id>")
		os.Exit(2)
	}
	for _, s := range rep.Specs {
		if s.ID == id {
			if s.Lifecycle == lint.LifeAccepted {
				fmt.Printf("specguard: %s already accepted at %s\n", id, s.Fingerprint)
				return
			}
			record(s)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "specguard accept: no such spec: %s\n", id)
	os.Exit(2)
}

// loadConfig resolves the config file and builds the lint config.
func loadConfig(root, configPath string, strict bool) lint.Config {
	cfgFile := configPath
	if cfgFile == "" {
		cfgFile = filepath.Join(root, ".specguard.yml")
	}
	cfg, err := lint.LoadConfig(root, cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard: config:", err)
		os.Exit(2)
	}
	cfg.Strict = strict
	return cfg
}

func lintCmd(args []string) {
	fs := flag.NewFlagSet("specguard", flag.ExitOnError)
	var (
		root       = fs.String("C", ".", "directory to run in (repo root)")
		configPath = fs.String("config", "", "config file (default: <root>/.specguard.yml)")
		asJSON     = fs.Bool("json", false, "emit the report as JSON (for tooling / UI)")
		strict     = fs.Bool("strict", false, "treat warnings as errors")
		noColor    = fs.Bool("no-color", false, "disable ANSI color")
	)
	_ = fs.Parse(args)

	cfg := loadConfig(*root, *configPath, *strict)
	rep, err := lint.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard:", err)
		os.Exit(2)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	} else {
		color := !*noColor && os.Getenv("NO_COLOR") == "" && isTTY(os.Stdout)
		lint.Render(os.Stdout, rep, color)
	}

	if !rep.OK {
		os.Exit(1)
	}
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("specguard serve", flag.ExitOnError)
	var (
		root       = fs.String("C", ".", "directory to lint (repo root)")
		configPath = fs.String("config", "", "config file (default: <root>/.specguard.yml)")
		strict     = fs.Bool("strict", false, "treat warnings as errors")
		addr       = fs.String("addr", ":8137", "listen address")
		webDir     = fs.String("web", "", "directory of built UI assets to serve (optional)")
		diffBase   = fs.String("diff-base", "HEAD", "git ref the 'Changed only' view diffs against (empty disables)")
	)
	_ = fs.Parse(args)

	cfg := loadConfig(*root, *configPath, *strict)
	srv := server.New(cfg, server.ResolveWebDir(*webDir))
	if *diffBase != "" {
		configName := ".specguard.yml"
		if *configPath != "" {
			configName = filepath.Base(*configPath)
		}
		srv.EnableDiff(*diffBase, configName)
	}
	fmt.Fprintf(os.Stderr, "specguard: serving report for %s on http://localhost%s/api/report\n", cfg.Root, *addr)
	if err := server.ListenAndServe(*addr, srv.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, "specguard: serve:", err)
		os.Exit(2)
	}
}

func diffCmd(args []string) {
	fs := flag.NewFlagSet("specguard diff", flag.ExitOnError)
	var (
		root             = fs.String("C", ".", "directory to run in (repo root)")
		configPath       = fs.String("config", "", "config file (default: <root>/.specguard.yml)")
		strict           = fs.Bool("strict", false, "treat warnings as errors")
		base             = fs.String("base", "origin/main", "git ref to diff against")
		format           = fs.String("format", "text", "output format: text | markdown | json")
		failOnRegression = fs.Bool("fail-on-regression", false, "exit non-zero if a spec lost coverage or arrived uncovered")
	)
	_ = fs.Parse(args)

	if !diff.HasGit(*root) {
		fmt.Fprintln(os.Stderr, "specguard diff: not a git repository (diff needs git):", *root)
		os.Exit(2)
	}

	cfg := loadConfig(*root, *configPath, *strict)
	head, err := lint.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard diff: head:", err)
		os.Exit(2)
	}
	configName := ".specguard.yml"
	if *configPath != "" {
		configName = filepath.Base(*configPath)
	}
	baseRep, err := diff.BaseReport(*root, *base, configName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard diff: base:", err)
		os.Exit(2)
	}
	changed, err := diff.ChangedFiles(*root, *base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard diff: changed files:", err)
		os.Exit(2)
	}

	d := diff.Compute(baseRep, head, changed)
	d.Base = *base

	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(d)
	case "markdown", "md":
		diff.RenderMarkdown(os.Stdout, d, *base)
	default:
		diff.RenderText(os.Stdout, d, *base)
	}

	if *failOnRegression && d.Regressions > 0 {
		os.Exit(1)
	}
}

// reportCmd writes a self-contained, browsable HTML report (the whole spec
// catalog with client-side search) — the artifact you publish per branch so
// anyone can explore the specs without running a server.
func reportCmd(args []string) {
	fs := flag.NewFlagSet("specguard report", flag.ExitOnError)
	var (
		root       = fs.String("C", ".", "directory to run in (repo root)")
		configPath = fs.String("config", "", "config file (default: <root>/.specguard.yml)")
		strict     = fs.Bool("strict", false, "treat warnings as errors")
		out        = fs.String("o", "", "output file (default: stdout)")
		format     = fs.String("format", "html", "output format: html | json")
		webFile    = fs.String("web", "", "override the embedded UI template with this built single-file HTML")
		branch     = fs.String("branch", "", "branch name to stamp (default: detected from git)")
		commit     = fs.String("commit", "", "commit SHA to stamp (default: detected from git)")
		repo       = fs.String("repo", "", "repository slug to stamp (e.g. owner/name)")
		resultsArg = fs.String("results", "", "comma-separated test result files (Playwright JSON / `go test -json`) to show pass/fail")
		assetsDir  = fs.String("assets", "", "directory to copy test screenshots into (enables per-spec galleries; makes the report a bundle, not one file)")
		assetsBase = fs.String("assets-base", "assets", "URL prefix the report loads copied screenshots from")
	)
	_ = fs.Parse(args)

	cfg := loadConfig(*root, *configPath, *strict)
	rep, err := lint.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard report:", err)
		os.Exit(2)
	}

	// Overlay a test run, if provided, so the report shows pass/fail per spec.
	if *resultsArg != "" {
		set, err := results.Load(strings.Split(*resultsArg, ","))
		if err != nil {
			fmt.Fprintln(os.Stderr, "specguard report:", err)
			os.Exit(2)
		}
		set.Apply(rep)
		// Publish any screenshots the run captured, as per-spec galleries.
		if *assetsDir != "" {
			if _, err := report.WriteArtifacts(rep, set.ArtifactsBySpec, *assetsDir, *assetsBase); err != nil {
				fmt.Fprintln(os.Stderr, "specguard report: artifacts:", err)
				os.Exit(2)
			}
		}
	}

	// Open the output sink up front, so we don't run a build only to fail on a
	// bad path.
	w := io.Writer(os.Stdout)
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, "specguard report:", err)
			os.Exit(2)
		}
		defer f.Close()
		w = f
	}

	if *format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			fmt.Fprintln(os.Stderr, "specguard report:", err)
			os.Exit(2)
		}
		return
	}

	// Resolve the UI template: an explicit freshly-built file wins, else the
	// version embedded in the binary.
	tmpl, err := resolveTemplate(*webFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard report:", err)
		os.Exit(2)
	}

	meta := buildMeta(*root, *branch, *commit, *repo)
	if err := report.Render(w, tmpl, rep, meta); err != nil {
		fmt.Fprintln(os.Stderr, "specguard report:", err)
		os.Exit(2)
	}
}

// resolveTemplate returns the single-file UI template: the file at webFile if
// given, otherwise the embedded default.
func resolveTemplate(webFile string) (string, error) {
	if webFile != "" {
		b, err := os.ReadFile(webFile)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return report.Template()
}

// buildMeta fills provenance, preferring explicit flags and falling back to git.
func buildMeta(root, branch, commit, repo string) report.Meta {
	gitBranch, gitCommit := report.DetectGit(root)
	if branch == "" {
		branch = gitBranch
	}
	if commit == "" {
		commit = gitCommit
	}
	short := commit
	if len(short) > 7 {
		short = short[:7]
	}
	return report.Meta{
		Repo:        repo,
		Branch:      branch,
		Commit:      commit,
		CommitShort: short,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
