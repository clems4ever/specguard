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
  if (!s.covered) return s.draft ? 'draft' : 'uncovered';
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
    // Coverage counts are independent of any test run.
    if (!s.covered) {
      if (s.draft) drafts++;
      else uncovered++;
    } else {
      covered++;
      if (hasResults) {
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
