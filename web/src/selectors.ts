// Pure derivations over a Report. Kept free of React/DOM so they can be unit
// tested in isolation and reused by every component.

import type { Report, SpecStatus, Finding, ChangeKind } from './types';

/** Human labels for the diff change kinds. */
export const KIND_LABEL: Record<ChangeKind, string> = {
  added: 'Added',
  removed: 'Removed',
  'coverage-lost': 'Coverage lost',
  'coverage-gained': 'Coverage gained',
  edited: 'Edited',
  'impl-changed': 'Impl changed',
};

/** A compact glyph per change kind, mirroring the CLI. */
export const KIND_GLYPH: Record<ChangeKind, string> = {
  added: '+',
  removed: '−',
  'coverage-lost': '✗',
  'coverage-gained': '✓',
  edited: '~',
  'impl-changed': '•',
};

export type SpecState =
  | 'covered'
  | 'warning'
  | 'draft'
  | 'uncovered'
  | 'passing'
  | 'failing'
  | 'skipped'
  | 'not-run';

/**
 * The single visual state a spec row should render. When a test run has been
 * ingested (hasResults), a covered spec shows its outcome — passing / failing /
 * skipped / not-run — so "covered but failing" reads as red, not green.
 * Otherwise it falls back to coverage state.
 */
export function specState(s: SpecStatus, hasResults = false): SpecState {
  // A parent spec is verified by its children, so it counts as covered even
  // without a direct test. (Its accurate subtree state is computed separately by
  // subtreeRollup; this is the fallback when a parent is shown on its own.)
  if (!s.covered && !s.hasChild) return s.draft ? 'draft' : 'uncovered';
  if (hasResults) {
    if (s.result === 'failed') return 'failing';
    if (s.result === 'passed') return 'passing';
    if (s.result === 'skipped') return 'skipped';
    return 'not-run'; // covered, but no result for it in the run
  }
  if (!s.coversOk) return 'warning';
  return 'covered';
}

export const STATE_LABEL: Record<SpecState, string> = {
  covered: 'Covered',
  warning: 'Covers unmatched',
  draft: 'Draft',
  uncovered: 'Uncovered',
  passing: 'Passing',
  failing: 'Failing',
  skipped: 'Skipped',
  'not-run': 'Not run',
};

// How alarming each state is, so a group of specs can be summarised by its
// worst member (a single failing spec should colour the whole area red).
const STATE_SEVERITY: Record<SpecState, number> = {
  failing: 5,
  uncovered: 5,
  warning: 3,
  'not-run': 2,
  skipped: 2,
  draft: 1,
  passing: 0,
  covered: 0,
};

/**
 * A one-glance summary of an area: how many specs it holds, its worst state
 * (for the roll-up badge) and whether every spec is in a good state — so the
 * header can say "all passing" instead of making a reader scan the whole list.
 */
export function areaRollup(
  specs: SpecStatus[],
  hasResults = false,
): { total: number; state: SpecState; allGood: boolean } {
  let worst: SpecState = hasResults ? 'passing' : 'covered';
  for (const s of specs) {
    const st = specState(s, hasResults);
    if (STATE_SEVERITY[st] > STATE_SEVERITY[worst]) worst = st;
  }
  return { total: specs.length, state: worst, allGood: STATE_SEVERITY[worst] === 0 };
}

// A spec and its refinements — the derivation tree built from `parent` links.
export interface SpecNode {
  spec: SpecStatus;
  children: SpecNode[];
  depth: number;
}

/**
 * Build a forest from `parent` links. A spec is a root within `specs` when it
 * has no parent, or its parent is not in this set (e.g. filtered out, or in
 * another area). Order is stable (by id) at every level, and a malformed cycle
 * can never loop — a node is only ever expanded once.
 */
export function buildTree(specs: SpecStatus[]): SpecNode[] {
  const byId = new Map(specs.map((s) => [s.id, s]));
  const kids = new Map<string, SpecStatus[]>();
  const roots: SpecStatus[] = [];
  for (const s of specs) {
    if (s.parent && byId.has(s.parent)) {
      (kids.get(s.parent) ?? kids.set(s.parent, []).get(s.parent)!).push(s);
    } else {
      roots.push(s);
    }
  }
  const byId2 = (a: SpecStatus, b: SpecStatus) => a.id.localeCompare(b.id);
  const seen = new Set<string>();
  const build = (s: SpecStatus, depth: number): SpecNode => {
    seen.add(s.id);
    const children = (kids.get(s.id) ?? [])
      .filter((c) => !seen.has(c.id))
      .sort(byId2)
      .map((c) => build(c, depth + 1));
    return { spec: s, children, depth };
  };
  const forest = roots.sort(byId2).map((r) => build(r, 0));
  // Any spec not reached from a root (only possible under a parent cycle, which
  // the linter rejects) is surfaced as its own root so it never vanishes.
  for (const s of specs.filter((s) => !seen.has(s.id)).sort(byId2)) {
    forest.push(build(s, 0));
  }
  return forest;
}

/**
 * The state a node should show: for a leaf, its own state; for a parent, the
 * worst state anywhere in its subtree — so an intent reads green only when
 * everything derived from it is proven. `count` is the number of descendants.
 */
export function subtreeRollup(
  node: SpecNode,
  hasResults = false,
): { state: SpecState; count: number } {
  // A leaf shows its own state. A parent takes the worst state in its subtree —
  // including its OWN state when it has a direct test of its own (a parent may
  // be verified both by children and by its own tests). A test-less parent
  // starts from a neutral-good baseline so it doesn't drag the rollup down.
  if (node.children.length === 0) {
    return { state: specState(node.spec, hasResults), count: 0 };
  }
  let state: SpecState = node.spec.covered
    ? specState(node.spec, hasResults)
    : hasResults
      ? 'passing'
      : 'covered';
  let count = 0;
  for (const c of node.children) {
    const r = subtreeRollup(c, hasResults);
    count += 1 + r.count;
    if (STATE_SEVERITY[r.state] > STATE_SEVERITY[state]) state = r.state;
  }
  return { state, count };
}

/** The area a spec belongs to, taken from `specs/<area>/<file>.md`. */
export function specArea(s: SpecStatus): string {
  const parts = s.path.split('/');
  // parts[0] is the specs dir; the next segment (if any) is the area.
  if (parts.length >= 3) return parts[1];
  return 'general';
}

export interface Summary {
  total: number;
  covered: number;
  uncovered: number;
  drafts: number;
  warnings: number;
  errors: number;
  testFiles: number;
  ok: boolean;
  coveragePct: number; // covered / (non-draft specs), 0..100
  hasResults: boolean;
  passing: number;
  failing: number;
  skipped: number;
  notRun: number; // covered specs with no result in the run
}

export function summarize(report: Report): Summary {
  const specs = report.specs ?? [];
  const findings = report.findings ?? [];
  const hasResults = !!report.hasResults;
  let covered = 0;
  let uncovered = 0;
  let drafts = 0;
  let passing = 0;
  let failing = 0;
  let skipped = 0;
  let notRun = 0;
  for (const s of specs) {
    // A parent spec is verified by its children, so it counts as covered. Its
    // pass/fail is a roll-up of its leaves (counted below), not a direct result.
    const isCovered = s.covered || !!s.hasChild;
    // Coverage counts are independent of any test run.
    if (!isCovered) {
      if (s.draft) drafts++;
      else uncovered++;
    } else {
      covered++;
      // Count a pass/fail result for any spec with a direct test — leaves, and
      // parents that also carry their own tests. A test-less parent has no
      // direct result of its own (its leaves are counted instead).
      if (hasResults && s.covered) {
        if (s.result === 'failed') failing++;
        else if (s.result === 'passed') passing++;
        else if (s.result === 'skipped') skipped++;
        else notRun++;
      }
    }
  }
  const enforceable = specs.length - drafts;
  return {
    total: specs.length,
    covered,
    uncovered,
    drafts,
    warnings: findings.filter((f) => f.severity === 'warning').length,
    errors: findings.filter((f) => f.severity === 'error').length,
    testFiles: report.testFiles,
    ok: report.ok,
    coveragePct: enforceable === 0 ? 100 : Math.round((covered / enforceable) * 100),
    hasResults,
    passing,
    failing,
    skipped,
    notRun,
  };
}

/** Case-insensitive filter over id, title and area. */
export function filterSpecs(specs: SpecStatus[], query: string): SpecStatus[] {
  const q = query.trim().toLowerCase();
  if (!q) return specs;
  return specs.filter(
    (s) =>
      s.id.toLowerCase().includes(q) ||
      s.title.toLowerCase().includes(q) ||
      specArea(s).toLowerCase().includes(q),
  );
}

export interface AreaGroup {
  area: string;
  specs: SpecStatus[];
}

/** Group specs by area, areas sorted alphabetically, specs sorted by id. */
export function groupByArea(specs: SpecStatus[]): AreaGroup[] {
  const byArea = new Map<string, SpecStatus[]>();
  for (const s of specs) {
    const a = specArea(s);
    const list = byArea.get(a) ?? [];
    list.push(s);
    byArea.set(a, list);
  }
  return [...byArea.entries()]
    .map(([area, list]) => ({
      area,
      specs: [...list].sort((x, y) => x.id.localeCompare(y.id)),
    }))
    .sort((x, y) => x.area.localeCompare(y.area));
}

/** Findings attached to a given spec id. */
export function findingsForSpec(report: Report, id: string): Finding[] {
  return (report.findings ?? []).filter((f) => f.spec === id);
}
