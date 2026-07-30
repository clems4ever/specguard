import { describe, it, expect } from 'vitest';
import {
  specState,
  specArea,
  summarize,
  filterSpecs,
  groupByArea,
  findingsForSpec,
  areaRollup,
  buildTree,
  subtreeRollup,
} from './selectors';
import { spec, mixedReport, resultsReport } from './test/fixtures';

describe('specState', () => {
  it('covered when covered and covers ok', () => {
    expect(specState(spec({ id: 'a', covered: true, coversOk: true }))).toBe('covered');
  });
  it('warning when covered but covers unmatched', () => {
    expect(specState(spec({ id: 'a', covered: true, coversOk: false }))).toBe('warning');
  });
  it('draft when uncovered and draft', () => {
    expect(specState(spec({ id: 'a', covered: false, draft: true }))).toBe('draft');
  });
  it('uncovered when uncovered and not draft', () => {
    expect(specState(spec({ id: 'a', covered: false, draft: false }))).toBe('uncovered');
  });
});

describe('specArea', () => {
  it('reads the area from specs/<area>/<file>.md', () => {
    expect(specArea(spec({ id: 'x', path: 'specs/auth/login.md' }))).toBe('auth');
  });
  it('falls back to general for a flat path', () => {
    expect(specArea(spec({ id: 'x', path: 'specs/login.md' }))).toBe('general');
  });
});

describe('summarize', () => {
  const s = summarize(mixedReport);
  it('counts states', () => {
    expect(s.total).toBe(4);
    expect(s.covered).toBe(2); // auth-login + auth-logout (warning still counts as covered)
    expect(s.uncovered).toBe(1); // tasks-create
    expect(s.drafts).toBe(1); // sharing-permissions
  });
  it('counts findings by severity', () => {
    expect(s.errors).toBe(1);
    expect(s.warnings).toBe(2);
  });
  it('coverage excludes drafts from the denominator', () => {
    // covered 2 of (4 - 1 draft) = 2/3 = 67%
    expect(s.coveragePct).toBe(67);
  });
  it('is 100% when there are no enforceable specs', () => {
    expect(summarize({ ok: true, testFiles: 0, specs: [], findings: [] }).coveragePct).toBe(100);
  });
});

describe('filterSpecs', () => {
  const specs = mixedReport.specs;
  it('returns all for an empty query', () => {
    expect(filterSpecs(specs, '  ')).toHaveLength(4);
  });
  it('matches on id, title and area', () => {
    expect(filterSpecs(specs, 'login').map((s) => s.id)).toEqual(['auth-login']);
    expect(filterSpecs(specs, 'create a task').map((s) => s.id)).toEqual(['tasks-create']);
    expect(filterSpecs(specs, 'sharing').map((s) => s.id)).toEqual(['sharing-permissions']);
  });
  it('returns empty when nothing matches', () => {
    expect(filterSpecs(specs, 'zzz')).toHaveLength(0);
  });
});

describe('groupByArea', () => {
  it('groups and sorts areas and specs', () => {
    const groups = groupByArea(mixedReport.specs);
    expect(groups.map((g) => g.area)).toEqual(['auth', 'sharing', 'tasks']);
    expect(groups[0].specs.map((s) => s.id)).toEqual(['auth-login', 'auth-logout']);
  });
});

describe('findingsForSpec', () => {
  it('returns findings for a given spec id', () => {
    expect(findingsForSpec(mixedReport, 'tasks-create').map((f) => f.rule)).toEqual([
      'uncovered-spec',
    ]);
  });
});

describe('specState with results', () => {
  it('reflects the test outcome for covered specs when a run was ingested', () => {
    expect(specState(spec({ id: 'a', covered: true, result: 'passed' }), true)).toBe('passing');
    expect(specState(spec({ id: 'b', covered: true, result: 'failed' }), true)).toBe('failing');
    expect(specState(spec({ id: 'c', covered: true, result: 'skipped' }), true)).toBe('skipped');
    // covered but no result present in the run
    expect(specState(spec({ id: 'd', covered: true }), true)).toBe('not-run');
    // uncovered stays uncovered regardless of results
    expect(specState(spec({ id: 'e', covered: false }), true)).toBe('uncovered');
  });

  it('ignores results when none were ingested (coverage view)', () => {
    expect(specState(spec({ id: 'a', covered: true, result: 'failed' }), false)).toBe('covered');
  });
});

describe('areaRollup', () => {
  it('is all-good when every spec is covered', () => {
    const r = areaRollup([spec({ id: 'a', covered: true }), spec({ id: 'b', covered: true })]);
    expect(r.total).toBe(2);
    expect(r.allGood).toBe(true);
    expect(r.state).toBe('covered');
  });

  it('reports the worst state in the area', () => {
    const r = areaRollup([spec({ id: 'a', covered: true }), spec({ id: 'b', covered: false })]);
    expect(r.allGood).toBe(false);
    expect(r.state).toBe('uncovered');
  });

  it('a single failing spec colours the whole area (with results)', () => {
    const r = areaRollup(
      [spec({ id: 'a', covered: true, result: 'passed' }), spec({ id: 'b', covered: true, result: 'failed' })],
      true,
    );
    expect(r.state).toBe('failing');
    expect(r.allGood).toBe(false);
  });
});

describe('buildTree', () => {
  it('nests children under their parent and sorts by id', () => {
    const forest = buildTree([
      spec({ id: 'cap-b', parent: 'cap', covered: true }),
      spec({ id: 'cap', hasChild: true }),
      spec({ id: 'cap-a', parent: 'cap', covered: true }),
      spec({ id: 'lonely', covered: true }),
    ]);
    // Two roots: the capability and the unparented spec.
    expect(forest.map((n) => n.spec.id)).toEqual(['cap', 'lonely']);
    const cap = forest[0];
    expect(cap.children.map((n) => n.spec.id)).toEqual(['cap-a', 'cap-b']);
    expect(cap.children[0].depth).toBe(1);
  });

  it('treats a spec whose parent is absent as a root', () => {
    const forest = buildTree([spec({ id: 'orphan', parent: 'gone', covered: true })]);
    expect(forest).toHaveLength(1);
    expect(forest[0].spec.id).toBe('orphan');
  });

  it('does not loop on a cyclic parent link', () => {
    const forest = buildTree([
      spec({ id: 'a', parent: 'b' }),
      spec({ id: 'b', parent: 'a' }),
    ]);
    // Both point at each other; buildTree still terminates and yields nodes.
    expect(forest.length).toBeGreaterThan(0);
  });
});

describe('subtreeRollup', () => {
  it('a parent is green only when every descendant is proven', () => {
    const forest = buildTree([
      spec({ id: 'cap', hasChild: true }),
      spec({ id: 'cap-ok', parent: 'cap', covered: true }),
      spec({ id: 'cap-bad', parent: 'cap', covered: false }), // uncovered
    ]);
    const roll = subtreeRollup(forest[0]);
    expect(roll.count).toBe(2);
    expect(roll.state).toBe('uncovered'); // worst wins
  });

  it('a parent whose children all pass reads passing (with results)', () => {
    const forest = buildTree([
      spec({ id: 'cap', hasChild: true }),
      spec({ id: 'cap-a', parent: 'cap', covered: true, result: 'passed' }),
      spec({ id: 'cap-b', parent: 'cap', covered: true, result: 'passed' }),
    ]);
    expect(subtreeRollup(forest[0], true).state).toBe('passing');
  });

  it("folds in the parent's OWN failing test even when its children pass", () => {
    const forest = buildTree([
      // A parent that ALSO carries its own (failing) test.
      spec({ id: 'cap', hasChild: true, covered: true, tests: ['cap_test.go'], result: 'failed' }),
      spec({ id: 'cap-a', parent: 'cap', covered: true, result: 'passed' }),
    ]);
    expect(subtreeRollup(forest[0], true).state).toBe('failing');
  });

  it("ignores a test-less parent's own state (neutral baseline)", () => {
    const forest = buildTree([
      spec({ id: 'cap', hasChild: true, covered: false }), // no direct test
      spec({ id: 'cap-a', parent: 'cap', covered: true, result: 'passed' }),
    ]);
    expect(subtreeRollup(forest[0], true).state).toBe('passing');
  });
});

describe('summarize with results', () => {
  it('counts passing/failing/skipped and flags hasResults', () => {
    const s = summarize(resultsReport);
    expect(s.hasResults).toBe(true);
    expect(s.passing).toBe(1);
    expect(s.failing).toBe(1);
    expect(s.skipped).toBe(1);
    expect(s.uncovered).toBe(1); // tasks-create has no test
  });
});
