import { describe, it, expect } from 'vitest';
import {
  specState,
  specArea,
  summarize,
  filterSpecs,
  groupByArea,
  findingsForSpec,
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
