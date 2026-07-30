package diff

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/clems4ever/specguard/internal/lint"
)

// ChangedFiles returns the repo-relative paths that differ between baseRef and
// the current working tree of the git repo at root. It includes tracked
// modifications (staged or not) AND brand-new untracked files, so it reflects a
// developer's in-progress change locally as well as a committed PR branch in CI
// (where base is e.g. origin/main).
func ChangedFiles(root, baseRef string) ([]string, error) {
	tracked, err := git(root, "diff", "--name-only", baseRef)
	if err != nil {
		return nil, err
	}
	// Untracked, non-ignored files are new to the tree and count as changes.
	untracked, err := git(root, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var files []string
	for _, block := range []string{tracked, untracked} {
		for _, line := range strings.Split(block, "\n") {
			s := filepath.ToSlash(strings.TrimSpace(line))
			if s != "" && !seen[s] {
				seen[s] = true
				files = append(files, s)
			}
		}
	}
	return files, nil
}

// BaseReport materializes baseRef in a throwaway detached worktree and runs the
// linter there, so the base report is computed by exactly the same code path as
// head — no reimplementation of the walk. The worktree is always cleaned up.
func BaseReport(root, baseRef, configName string) (*lint.Report, error) {
	tmp, err := os.MkdirTemp("", "specguard-base-")
	if err != nil {
		return nil, err
	}
	// git refuses to add a worktree at an existing non-empty dir; remove the
	// freshly-created temp dir and let git create it.
	_ = os.RemoveAll(tmp)
	defer func() {
		_, _ = git(root, "worktree", "remove", "--force", tmp)
		_ = os.RemoveAll(tmp)
	}()

	if _, err := git(root, "worktree", "add", "--detach", "--quiet", tmp, baseRef); err != nil {
		return nil, fmt.Errorf("checkout base %q: %w", baseRef, err)
	}
	cfgPath := filepath.Join(tmp, configName)
	cfg, err := lint.LoadConfig(tmp, cfgPath)
	if err != nil {
		return nil, err
	}
	return lint.Run(cfg)
}

// HasGit reports whether root is inside a git work tree (diff mode needs it).
func HasGit(root string) bool {
	out, err := git(root, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
