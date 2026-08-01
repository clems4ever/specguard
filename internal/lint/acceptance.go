package lint

// Acceptance turns a PM's one-time behavioural review into a durable, auditable
// gate. The PM reviews a spec's behaviour once (via the demo / its acceptance
// tests) and ratifies it; that sign-off is recorded against a *fingerprint* of
// the expectation — the spec's intent plus the source of its covering tests.
//
// From then on the covering tests guard non-regression automatically: as long
// as the fingerprint is unchanged, the acceptance holds and the PM is never
// asked again. If the intent or a covering test's source changes, the
// fingerprint changes, the acceptance goes stale, and exactly that spec routes
// back for re-review. A pure code refactor (tests unchanged) never disturbs it.

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LedgerPath is the append-only record of PM acceptances, relative to the root.
const LedgerPath = ".specguard/acceptance.jsonl"

// Lifecycle values for a spec.
const (
	LifeProposed    = "proposed"    // wanted by a PM, not yet built (no covering test)
	LifeImplemented = "implemented" // built (covered by a test) but never accepted
	LifeAccepted    = "accepted"    // PM-accepted at the current fingerprint
	LifeStale       = "stale"       // accepted before, but the expectation has since changed
)

// Acceptance is one entry in the ledger: a PM's verdict on a spec at a specific
// fingerprint. The ledger is append-only, so it doubles as a review history.
type Acceptance struct {
	Spec        string `json:"spec"`
	Fingerprint string `json:"fingerprint"`
	Verdict     string `json:"verdict"` // "accepted" (others reserved, e.g. "changes-requested")
	By          string `json:"by,omitempty"`
	At          string `json:"at,omitempty"`
	Evidence    string `json:"evidence,omitempty"` // link to the demo / preview the PM reviewed
}

// fingerprint hashes the *expectation*: the spec's intent (id, title, body) and
// the verbatim source of every covering test, sorted for stability. It
// deliberately excludes pass/fail and screenshots — those are guarded
// continuously by CI, not by the one-time human review — so a refactor that
// keeps the tests' source identical does not invalidate an acceptance.
func fingerprint(id, title, body string, coveringFiles []string, content map[string]string) string {
	h := sha256.New()
	io := func(s string) { _, _ = h.Write([]byte(s)); _, _ = h.Write([]byte{0}) }
	io("id")
	io(id)
	io("title")
	io(title)
	io("intent")
	io(body)
	files := append([]string(nil), coveringFiles...)
	sort.Strings(files)
	for _, f := range files {
		io("test")
		io(f)
		io(content[f])
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// LoadLedger reads the acceptance ledger under root. A missing ledger is not an
// error (no acceptances yet); malformed lines are skipped so one bad edit can't
// wedge the whole gate.
func LoadLedger(root string) ([]Acceptance, error) {
	f, err := os.Open(filepath.Join(root, LedgerPath))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Acceptance
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var a Acceptance
		if json.Unmarshal([]byte(line), &a) == nil && a.Spec != "" {
			out = append(out, a)
		}
	}
	return out, sc.Err()
}

// acceptedFingerprints maps each spec to the set of fingerprints it has ever
// been accepted at (reverting to a previously-accepted state restores it).
func acceptedFingerprints(ledger []Acceptance) map[string]map[string]bool {
	m := map[string]map[string]bool{}
	for _, a := range ledger {
		if a.Verdict != "accepted" {
			continue
		}
		if m[a.Spec] == nil {
			m[a.Spec] = map[string]bool{}
		}
		m[a.Spec][a.Fingerprint] = true
	}
	return m
}

// latestAcceptance returns the most recent accepted entry per (spec,fingerprint),
// so a spec's current acceptance can show who signed off and when.
func latestAcceptance(ledger []Acceptance) map[string]Acceptance {
	m := map[string]Acceptance{}
	for _, a := range ledger {
		if a.Verdict == "accepted" {
			m[a.Spec+"@"+a.Fingerprint] = a
		}
	}
	return m
}

// lifecycleOf resolves a spec's state from its coverage and acceptance history.
func lifecycleOf(covered bool, fp string, accepted map[string]bool) string {
	switch {
	case accepted[fp]:
		return LifeAccepted
	case len(accepted) > 0:
		return LifeStale
	case covered:
		return LifeImplemented
	default:
		return LifeProposed
	}
}

// AppendAcceptance appends one entry to the ledger under root, creating the
// .specguard directory if needed.
func AppendAcceptance(root string, a Acceptance) error {
	dir := filepath.Join(root, ".specguard")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "acceptance.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}
