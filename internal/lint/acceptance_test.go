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
	refs := []Ref{{File: "a_test.go", Line: 2, Test: "TestLogin"}}
	// Two test functions share one file; the spec is covered only by TestLogin.
	src := func(login, other string) map[string]string {
		return map[string]string{"a_test.go": "package x\nfunc TestLogin(t *testing.T){ " + login + " }\nfunc TestOther(t *testing.T){ " + other + " }\n"}
	}
	base := fingerprint("login", "Log in", "intent", refs, src("want(1)", "misc(9)"))

	if base != fingerprint("login", "Log in", "intent", refs, src("want(1)", "misc(9)")) {
		t.Fatal("fingerprint must be stable for identical input")
	}
	// Editing the COVERING test's body changes the fingerprint.
	if base == fingerprint("login", "Log in", "intent", refs, src("want(2)", "misc(9)")) {
		t.Fatal("fingerprint must change when the covering test changes")
	}
	// Editing an UNRELATED test in the same file does NOT (function-granular).
	if base != fingerprint("login", "Log in", "intent", refs, src("want(1)", "misc(42)")) {
		t.Fatal("an unrelated test in the same file must not change the fingerprint")
	}
	// Changing the intent changes the fingerprint.
	if base == fingerprint("login", "Different", "intent", refs, src("want(1)", "misc(9)")) {
		t.Fatal("fingerprint must change when the intent changes")
	}
}

// spec:spec-acceptance
func TestFingerprintIgnoresLocationShifts(t *testing.T) {
	refs := []Ref{{File: "a_test.go", Line: 2, Test: "TestLogin"}}
	before := fingerprint("login", "Log in", "i", refs,
		map[string]string{"a_test.go": "package x\nfunc TestLogin(t *testing.T){ want(1) }\n"})
	// A new test added ABOVE shifts TestLogin down; Go locates it by name, so the
	// covering block's source is unchanged and the fingerprint holds.
	after := fingerprint("login", "Log in", "i", refs,
		map[string]string{"a_test.go": "package x\nfunc TestNew(t *testing.T){ x() }\nfunc TestLogin(t *testing.T){ want(1) }\n"})
	if before != after {
		t.Fatal("adding a test above must not change the fingerprint (located by name, keyed by source)")
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

// spec:spec-acceptance
func TestUnrelatedTestEditKeepsAcceptance(t *testing.T) {
	files := map[string]string{
		"specs/login.md":        "---\nid: login\ntitle: Log in\ncovers:\n  - internal/auth\n---\nintent\n",
		"internal/auth/auth.go": "package auth\n",
		// Two tests in one file; only TestLogin covers the spec.
		"internal/auth/auth_test.go": "package auth\n// %SPEC%login\nfunc TestLogin(t *testing.T){ want(1) }\nfunc TestUnrelated(t *testing.T){ misc(1) }\n",
	}
	root := writeTree(t, files)
	cfg := DefaultConfig(root)

	rep, _ := Run(cfg)
	s := specByID(rep, "login")
	if err := AppendAcceptance(root, Acceptance{Spec: "login", Fingerprint: s.Fingerprint, Verdict: "accepted", By: "pm"}); err != nil {
		t.Fatal(err)
	}
	// Edit the UNRELATED test only.
	if err := os.WriteFile(filepath.Join(root, "internal/auth/auth_test.go"),
		[]byte(fixture("package auth\n// %SPEC%login\nfunc TestLogin(t *testing.T){ want(1) }\nfunc TestUnrelated(t *testing.T){ misc(999) }\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, _ = Run(cfg)
	if specByID(rep, "login").Lifecycle != LifeAccepted {
		t.Fatal("editing an unrelated test in the same file must not disturb acceptance")
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
