package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/clems4ever/specguard/internal/lint"
)

// WriteArtifacts copies each spec's source artifacts (screenshots produced by a
// test run) into assetsDir under specs/<slug>/, and rewrites the report's
// SpecStatus.Artifacts to URL paths rooted at assetsBase (e.g. "assets"). This
// is what turns a single-file report into a Pages bundle (index.html + assets/).
// A missing or unreadable source is skipped — a report should still render — and
// the number of artifacts actually written is returned.
func WriteArtifacts(rep *lint.Report, srcBySpec map[string][]lint.Artifact, assetsDir, assetsBase string) (int, error) {
	if assetsDir == "" || len(srcBySpec) == 0 {
		return 0, nil
	}
	written := 0
	for i := range rep.Specs {
		s := &rep.Specs[i]
		srcs := srcBySpec[s.ID]
		if len(srcs) == 0 {
			continue
		}
		slug := SlugBranch(s.ID)
		destDir := filepath.Join(assetsDir, "specs", slug)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return written, err
		}
		var arts []lint.Artifact
		for j, src := range srcs {
			ext := filepath.Ext(src.Path)
			if ext == "" {
				ext = ".png"
			}
			name := fmt.Sprintf("%d%s", j, ext)
			if err := copyFile(src.Path, filepath.Join(destDir, name)); err != nil {
				fmt.Fprintf(os.Stderr, "specguard: skipping artifact %q: %v\n", src.Path, err)
				continue
			}
			arts = append(arts, lint.Artifact{
				Name: src.Name,
				Path: assetsBase + "/specs/" + slug + "/" + name,
			})
			written++
		}
		s.Artifacts = arts
	}
	return written, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
