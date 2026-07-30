package report

import (
	"os/exec"
	"strings"
)

// DetectGit returns the current branch and full commit SHA of the git repo at
// root, best-effort: any value that can't be determined comes back empty. This
// lets `specguard report` stamp provenance automatically, while explicit flags
// (and CI-provided values) still take precedence.
func DetectGit(root string) (branch, commit string) {
	return gitLine(root, "rev-parse", "--abbrev-ref", "HEAD"),
		gitLine(root, "rev-parse", "HEAD")
}

func gitLine(root string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if s == "HEAD" { // detached: not a useful "branch"
		return ""
	}
	return s
}
