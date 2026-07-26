// Command specguard checks spec traceability: every spec markdown file must be
// referenced by at least one test, and every `spec:<id>` reference in a test
// must resolve to a defined spec. It exits non-zero when the check fails, so it
// can gate a CI build.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/clems4ever/specguard/internal/lint"
)

func main() {
	var (
		root       = flag.String("C", ".", "directory to run in (repo root)")
		configPath = flag.String("config", "", "config file (default: <root>/.specguard.yml)")
		asJSON     = flag.Bool("json", false, "emit the report as JSON (for tooling / UI)")
		strict     = flag.Bool("strict", false, "treat warnings as errors")
		noColor    = flag.Bool("no-color", false, "disable ANSI color")
	)
	flag.Parse()

	cfgFile := *configPath
	if cfgFile == "" {
		cfgFile = filepath.Join(*root, ".specguard.yml")
	}
	cfg, err := lint.LoadConfig(*root, cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "specguard: config:", err)
		os.Exit(2)
	}
	cfg.Strict = *strict

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

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
