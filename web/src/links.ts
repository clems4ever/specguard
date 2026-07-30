import type { ReportMeta } from './types';

// Building a GitHub link needs the repo slug and the exact commit the report was
// generated from — both carried in ReportMeta. When either is missing (e.g. the
// live `serve` UI, which has no commit), we return null and callers fall back to
// plain text. Pinning to the commit means a link always points at the code as it
// was when the report was built, not a moving branch head.
export function codeLink(meta: ReportMeta | null | undefined, path: string, line?: number): string | null {
  if (!meta || !meta.repo || !meta.commit || !path) return null;
  const base = `https://github.com/${meta.repo}/blob/${meta.commit}/${path}`;
  return line && line > 0 ? `${base}#L${line}` : base;
}

/** "path:line" or just "path" when no line — the human-readable label. */
export function refLabel(path: string, line?: number): string {
  return line && line > 0 ? `${path}:${line}` : path;
}
