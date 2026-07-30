// Mirrors the JSON emitted by `specguard serve` (internal/lint.Report).

export interface Finding {
  severity: 'error' | 'warning';
  rule: string;
  spec?: string;
  file?: string;
  message: string;
}

// The outcome of a test run for one spec/reference (empty = not ingested).
export type TestStatus = 'passed' | 'failed' | 'skipped';

// A `spec:<id>` reference: the test file and 1-based line it sits on.
export interface Ref {
  file: string;
  line: number;
  test?: string; // Go test function this reference sits above
  status?: TestStatus; // outcome once results are ingested
}

// A visual proof (screenshot) captured by a test, attached to a spec.
export interface Artifact {
  name?: string;
  path: string; // URL the report loads it from
}

export interface SpecStatus {
  id: string;
  title: string;
  status?: string;
  path: string;
  body?: string;
  tests: string[] | null;
  refs?: Ref[] | null;
  covers?: string[];
  covered: boolean;
  coversOk: boolean;
  draft: boolean;
  result?: TestStatus; // aggregate outcome of covering tests
  artifacts?: Artifact[] | null; // screenshots captured by covering tests
}

// AreaInfo is an optional human overview of a spec area, sourced from a
// `specs/<area>/_area.md` file, so the report can say what a group of specs is
// about instead of showing a bare directory name.
export interface AreaInfo {
  name: string;
  title?: string;
  description?: string;
}

export interface Report {
  specs: SpecStatus[];
  findings: Finding[] | null;
  testFiles: number;
  ok: boolean;
  areas?: AreaInfo[] | null; // optional per-area overviews
  hasResults?: boolean; // a test run was ingested → show pass/fail
}

// ReportMeta stamps a generated static report so a viewer knows exactly what
// they are looking at (which branch, which commit, how fresh). `specguard
// report` injects it as window.__SPECGUARD_META__.
export interface ReportMeta {
  repo?: string;
  branch?: string;
  commit?: string;
  commitShort?: string;
  generatedAt?: string; // RFC3339
}

// Mirrors internal/diff (the "what changed" delta vs a base git ref).
export type ChangeKind =
  | 'added'
  | 'removed'
  | 'coverage-lost'
  | 'coverage-gained'
  | 'edited'
  | 'impl-changed';

export interface DiffChange {
  id: string;
  title: string;
  kind: ChangeKind;
  detail?: string;
  draft: boolean;
  regression: boolean;
  files?: string[] | null;
}

export interface Delta {
  base: string;
  changes: DiffChange[] | null;
  added: number;
  removed: number;
  coverageLost: number;
  coverageGained: number;
  edited: number;
  implChanged: number;
  regressions: number;
  baseOk: boolean;
  headOk: boolean;
}

export interface DiffResponse {
  enabled: boolean;
  delta?: Delta;
  error?: string;
  base?: string;
}
