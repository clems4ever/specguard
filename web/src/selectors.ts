// Pure derivations over a Report. Kept free of React/DOM so they can be unit
// tested in isolation and reused by every component.

import type { Report, SpecStatus, Finding } from './types';

export type SpecState = 'covered' | 'warning' | 'draft' | 'uncovered';

/** The single visual state a spec row should render. */
export function specState(s: SpecStatus): SpecState {
  if (!s.covered) return s.draft ? 'draft' : 'uncovered';
  if (!s.coversOk) return 'warning';
  return 'covered';
}

export const STATE_LABEL: Record<SpecState, string> = {
  covered: 'Covered',
  warning: 'Covers unmatched',
  draft: 'Draft',
  uncovered: 'Uncovered',
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
}

export function summarize(report: Report): Summary {
  const specs = report.specs ?? [];
  const findings = report.findings ?? [];
  let covered = 0;
  let uncovered = 0;
  let drafts = 0;
  let warnings = 0;
  for (const s of specs) {
    const st = specState(s);
    if (st === 'draft') drafts++;
    else if (st === 'uncovered') uncovered++;
    else covered++;
    if (st === 'warning') warnings++;
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
