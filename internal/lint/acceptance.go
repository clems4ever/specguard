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
	"regexp"
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
// the verbatim source of each covering test — the specific test *function or
// block*, not the whole file, and keyed by source rather than location. So
// adding an unrelated test to a shared file, reordering tests, or renaming the
// file leaves the fingerprint unchanged; only editing a covering test's own
// source (or the intent) moves it. Pass/fail and screenshots are excluded —
// those are guarded continuously by CI, not by the one-time human review.
func fingerprint(id, title, body string, refs []Ref, content map[string]string) string {
	h := sha256.New()
	io := func(s string) { _, _ = h.Write([]byte(s)); _, _ = h.Write([]byte{0}) }
	io("id")
	io(id)
	io("title")
	io(title)
	io("intent")
	io(body)

	// Collect the source of each covering test block, de-duplicated by content
	// and sorted, so neither location nor order affects the hash.
	seen := map[string]bool{}
	var blocks []string
	for _, r := range refs {
		src, ok := content[r.File]
		if !ok {
			continue
		}
		block := coveringBlock(src, r, strings.HasSuffix(r.File, "_test.go"))
		if !seen[block] {
			seen[block] = true
			blocks = append(blocks, block)
		}
	}
	sort.Strings(blocks)
	for _, b := range blocks {
		io("test")
		io(b)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// tsTestStart matches the opener of a JS/TS test construct, so a reference's
// enclosing test(...) call can be located.
var tsTestStart = regexp.MustCompile(`\b(test|it|describe)(\.\w+)?\s*\(`)

// coveringBlock returns the source of the test a reference sits in: the Go
// function it names, or the enclosing test(...) call for other languages. It
// falls back to the whole file when the block can't be isolated, so acceptance
// is never silently kept across a behaviour change (over- rather than
// under-triggering).
func coveringBlock(content string, r Ref, isGo bool) string {
	lines := strings.Split(content, "\n")
	start := -1
	if isGo && r.Test != "" {
		for i, ln := range lines {
			if m := funcPattern.FindStringSubmatch(ln); m != nil && m[1] == r.Test {
				start = i
				break
			}
		}
	} else {
		// r.Line is 1-based; the tag usually sits on the test(...) opener line.
		from := r.Line - 1
		if from >= len(lines) {
			from = len(lines) - 1
		}
		for i := from; i >= 0 && i >= from-8; i-- {
			if tsTestStart.MatchString(lines[i]) {
				start = i
				break
			}
		}
	}
	if start < 0 {
		return content
	}
	if block, ok := braceBlock(lines, start); ok {
		return block
	}
	return content
}

// braceBlock returns lines[start:] up to and including the line where brace
// depth (opened on or after start) first returns to zero — i.e. the construct's
// full source. Not string/comment aware; good enough for well-formed tests.
func braceBlock(lines []string, start int) (string, bool) {
	depth, started := 0, false
	var b strings.Builder
	for i := start; i < len(lines); i++ {
		b.WriteString(lines[i])
		b.WriteByte('\n')
		for _, c := range lines[i] {
			switch c {
			case '{':
				depth++
				started = true
			case '}':
				depth--
			}
		}
		if started && depth <= 0 {
			return b.String(), true
		}
	}
	return "", false
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
