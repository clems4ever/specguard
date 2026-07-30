package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func TestWriteArtifactsCopiesAndRewritesPaths(t *testing.T) {
	tmp := t.TempDir()
	// A source screenshot on disk (as a runner would leave it).
	src := filepath.Join(tmp, "shot.png")
	if err := os.WriteFile(src, []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep := &lint.Report{Specs: []lint.SpecStatus{{ID: "auth-login"}, {ID: "no-shots"}}}
	srcBySpec := map[string][]lint.Artifact{
		"auth-login": {{Name: "login screen", Path: src}},
	}

	assets := filepath.Join(tmp, "public", "assets")
	n, err := WriteArtifacts(rep, srcBySpec, assets, "assets")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("wrote %d artifacts, want 1", n)
	}

	// The file was copied under specs/<id>/ and its bytes preserved.
	dst := filepath.Join(assets, "specs", "auth-login", "0.png")
	if b, err := os.ReadFile(dst); err != nil || string(b) != "PNGDATA" {
		t.Fatalf("copied file wrong: b=%q err=%v", b, err)
	}
	// The report now references it by URL, rooted at the assets base.
	got := rep.Specs[0].Artifacts
	if len(got) != 1 || got[0].Path != "assets/specs/auth-login/0.png" || got[0].Name != "login screen" {
		t.Fatalf("artifact rewrite = %+v", got)
	}
	// A spec with no artifacts is untouched.
	if rep.Specs[1].Artifacts != nil {
		t.Errorf("no-shots spec got artifacts: %+v", rep.Specs[1].Artifacts)
	}
}

func TestWriteArtifactsSkipsMissingSource(t *testing.T) {
	tmp := t.TempDir()
	rep := &lint.Report{Specs: []lint.SpecStatus{{ID: "x"}}}
	srcBySpec := map[string][]lint.Artifact{
		"x": {{Name: "gone", Path: filepath.Join(tmp, "does-not-exist.png")}},
	}
	// A missing source must not fail the whole report — it's skipped.
	n, err := WriteArtifacts(rep, srcBySpec, filepath.Join(tmp, "assets"), "assets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("wrote %d, want 0", n)
	}
	if len(rep.Specs[0].Artifacts) != 0 {
		t.Errorf("expected no artifacts recorded, got %+v", rep.Specs[0].Artifacts)
	}
}
