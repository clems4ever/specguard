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
