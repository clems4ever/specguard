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
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/clems4ever/specguard/internal/diff"
	"github.com/clems4ever/specguard/internal/lint"
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
		}
	}
	lintCmd(os.Args[1:])
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

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
