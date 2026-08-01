package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func specByID(rep *Report, id string) SpecStatus {
	for _, s := range rep.Specs {
		if s.ID == id {
			return s
		}
	}
	return SpecStatus{}
}

// spec:spec-acceptance
func TestFingerprintStableAndBehaviourSensitive(t *testing.T) {
	base := fingerprint("login", "Log in", "intent body", []string{"a_test.go"}, map[string]string{"a_test.go": "SRC"})
	same := fingerprint("login", "Log in", "intent body", []string{"a_test.go"}, map[string]string{"a_test.go": "SRC"})
	if base != same {
		t.Fatal("fingerprint must be stable for identical input")
	}
	// Changing a covering test's source changes the fingerprint (behaviour moved).
	if base == fingerprint("login", "Log in", "intent body", []string{"a_test.go"}, map[string]string{"a_test.go": "SRC2"}) {
		t.Fatal("fingerprint must change when a covering test changes")
	}
	// Changing the intent changes the fingerprint.
	if base == fingerprint("login", "Log in", "different intent", []string{"a_test.go"}, map[string]string{"a_test.go": "SRC"}) {
		t.Fatal("fingerprint must change when the intent changes")
	}
	// Content of a NON-covering file is irrelevant.
	if base != fingerprint("login", "Log in", "intent body", []string{"a_test.go"}, map[string]string{"a_test.go": "SRC", "other.go": "X"}) {
		t.Fatal("only covering-test source should feed the fingerprint")
	}
}

// spec:spec-acceptance
func TestLifecycleAcceptThenStaleOnBehaviourChange(t *testing.T) {
	files := map[string]string{
		"specs/login.md":             "---\nid: login\ntitle: Log in\ncovers:\n  - internal/auth\n---\nintent\n",
		"internal/auth/auth.go":      "package auth\n",
		"internal/auth/auth_test.go": "package auth\n// %SPEC%login\nfunc TestLogin(t *testing.T){}\n",
	}
	root := writeTree(t, files)
	cfg := DefaultConfig(root)

	// 1. Built but never accepted → implemented.
	rep, _ := Run(cfg)
	s := specByID(rep, "login")
	if s.Lifecycle != LifeImplemented {
		t.Fatalf("want implemented, got %q", s.Lifecycle)
	}

	// 2. PM accepts at the current fingerprint → accepted.
	if err := AppendAcceptance(root, Acceptance{Spec: "login", Fingerprint: s.Fingerprint, Verdict: "accepted", By: "pm"}); err != nil {
		t.Fatal(err)
	}
	rep, _ = Run(cfg)
	s = specByID(rep, "login")
	if s.Lifecycle != LifeAccepted || s.AcceptedBy != "pm" {
		t.Fatalf("want accepted by pm, got %q by %q", s.Lifecycle, s.AcceptedBy)
	}

	// 3. A pure refactor of NON-test code must NOT disturb acceptance.
	if err := os.WriteFile(filepath.Join(root, "internal/auth/auth.go"), []byte("package auth\n// refactored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, _ = Run(cfg)
	if specByID(rep, "login").Lifecycle != LifeAccepted {
		t.Fatal("a code-only refactor must not invalidate acceptance")
	}

	// 4. Changing the covering TEST (the expectation) → stale.
	if err := os.WriteFile(filepath.Join(root, "internal/auth/auth_test.go"),
		[]byte(fixture("package auth\n// %SPEC%login\nfunc TestLogin(t *testing.T){ _ = 1 }\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, _ = Run(cfg)
	if specByID(rep, "login").Lifecycle != LifeStale {
		t.Fatalf("changing the covering test must make it stale, got %q", specByID(rep, "login").Lifecycle)
	}

	// 5. Re-accepting at the new fingerprint → accepted again.
	s = specByID(rep, "login")
	if err := AppendAcceptance(root, Acceptance{Spec: "login", Fingerprint: s.Fingerprint, Verdict: "accepted", By: "pm"}); err != nil {
		t.Fatal(err)
	}
	rep, _ = Run(cfg)
	if specByID(rep, "login").Lifecycle != LifeAccepted {
		t.Fatal("re-acceptance at the new fingerprint should restore accepted")
	}
}

func TestProposedWhenNoTestYet(t *testing.T) {
	rep := run(t, map[string]string{
		"specs/planned.md": "---\nid: planned\ntitle: Planned\nstatus: draft\n---\nintent\n",
	})
	if specByID(rep, "planned").Lifecycle != LifeProposed {
		t.Fatalf("an untested spec should be proposed, got %q", specByID(rep, "planned").Lifecycle)
	}
}
