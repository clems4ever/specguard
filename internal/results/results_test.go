package results

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/clems4ever/specguard/internal/lint"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const playwrightJSON = `{
  "config": {},
  "suites": [
    {
      "specs": [
        {"title": "logs in", "tags": ["@spec:auth-login"], "tests": [{"results": [{"status": "passed"}]}]}
      ],
      "suites": [
        {"specs": [
          {"title": "logout @spec:auth-logout", "tags": [], "tests": [{"results": [{"status": "failed"}]}]},
          {"title": "skipped one", "tags": ["@spec:auth-skip"], "tests": [{"results": [{"status": "skipped"}]}]}
        ]}
      ]
    }
  ]
}`

const goTestJSON = `{"Action":"run","Test":"TestPass"}
{"Action":"output","Test":"TestPass","Output":"ok\n"}
{"Action":"pass","Test":"TestPass"}
{"Action":"run","Test":"TestFail"}
{"Action":"fail","Test":"TestFail"}
{"Action":"skip","Test":"TestSkip"}
{"Action":"pass","Package":"example/pkg"}
not-json build noise
`

func TestPlaywrightParsing(t *testing.T) {
	set, err := Load([]string{writeFile(t, "pw.json", playwrightJSON)})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]lint.TestStatus{
		"auth-login":  lint.StatusPassed,  // from tag
		"auth-logout": lint.StatusFailed,  // from title, nested suite
		"auth-skip":   lint.StatusSkipped, // skipped
	}
	for id, w := range want {
		if set.BySpec[id] != w {
			t.Errorf("BySpec[%q] = %q, want %q", id, set.BySpec[id], w)
		}
	}
}

func TestGoTestParsing(t *testing.T) {
	set, err := Load([]string{writeFile(t, "go.json", goTestJSON)})
	if err != nil {
		t.Fatal(err)
	}
	for name, w := range map[string]lint.TestStatus{
		"TestPass": lint.StatusPassed,
		"TestFail": lint.StatusFailed,
		"TestSkip": lint.StatusSkipped,
	} {
		if set.ByTest[name] != w {
			t.Errorf("ByTest[%q] = %q, want %q", name, set.ByTest[name], w)
		}
	}
	// A package-level event (no Test) must not create a bogus entry.
	if _, ok := set.ByTest[""]; ok {
		t.Error("package-level event leaked an empty-name entry")
	}
}

func TestFormatAutoDetection(t *testing.T) {
	if !looksLikePlaywright([]byte(playwrightJSON)) {
		t.Error("playwright report should be detected")
	}
	if looksLikePlaywright([]byte(goTestJSON)) {
		t.Error("go test -json (NDJSON) must not be detected as playwright")
	}
}

func TestApplyCorrelatesAndAggregates(t *testing.T) {
	rep := &lint.Report{
		Specs: []lint.SpecStatus{
			{ID: "auth-login", Covered: true, Refs: []lint.Ref{
				{File: "server/auth_test.go", Line: 5, Test: "TestPass"}, // go pass
				{File: "web/e2e/auth.spec.ts", Line: 3},                  // playwright, by spec id → passed
			}},
			{ID: "auth-logout", Covered: true, Refs: []lint.Ref{
				{File: "web/e2e/auth.spec.ts", Line: 8}, // playwright → failed
			}},
			{ID: "mixed", Covered: true, Refs: []lint.Ref{
				{File: "x_test.go", Line: 1, Test: "TestPass"}, // pass
				{File: "y_test.go", Line: 1, Test: "TestFail"}, // fail → spec fails (worst-wins)
			}},
			{ID: "auth-skip", Covered: true, Refs: []lint.Ref{
				{File: "web/e2e/s.spec.ts", Line: 1},
			}},
			{ID: "no-result", Covered: true, Refs: []lint.Ref{
				{File: "z_test.go", Line: 1, Test: "TestUnknown"}, // no result ingested
			}},
		},
	}
	set, err := Load([]string{
		writeFile(t, "pw.json", playwrightJSON),
		writeFile(t, "go.json", goTestJSON),
	})
	if err != nil {
		t.Fatal(err)
	}
	set.Apply(rep)

	if !rep.HasResults {
		t.Fatal("HasResults should be true after Apply")
	}
	want := map[string]lint.TestStatus{
		"auth-login":  lint.StatusPassed,
		"auth-logout": lint.StatusFailed,
		"mixed":       lint.StatusFailed, // one failing covering test fails the spec
		"auth-skip":   lint.StatusSkipped,
		"no-result":   "", // nothing ingested → no result
	}
	byID := map[string]lint.SpecStatus{}
	for _, s := range rep.Specs {
		byID[s.ID] = s
	}
	for id, w := range want {
		if byID[id].Result != w {
			t.Errorf("spec %q Result = %q, want %q", id, byID[id].Result, w)
		}
	}
	// Per-covering-test status is filled in for the Go ref.
	loginGoRef := byID["auth-login"].Refs[0]
	if loginGoRef.Status != lint.StatusPassed {
		t.Errorf("go ref status = %q, want passed", loginGoRef.Status)
	}
}
