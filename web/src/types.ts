// Mirrors the JSON emitted by `specguard serve` (internal/lint.Report).

export interface Finding {
  severity: 'error' | 'warning';
  rule: string;
  spec?: string;
  file?: string;
  message: string;
}

export interface SpecStatus {
  id: string;
  title: string;
  status?: string;
  path: string;
  body?: string;
  tests: string[] | null;
  covers?: string[];
  covered: boolean;
  coversOk: boolean;
  draft: boolean;
}

export interface Report {
  specs: SpecStatus[];
  findings: Finding[] | null;
  testFiles: number;
  ok: boolean;
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
