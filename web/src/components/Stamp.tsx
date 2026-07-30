import type { ReportMeta } from '../types';

/** Human "3h ago" from an RFC3339 timestamp; empty string if unparseable. */
function relativeTime(iso?: string): string {
  if (!iso) return '';
  const then = Date.parse(iso);
  if (Number.isNaN(then)) return '';
  const secs = Math.max(0, Math.round((Date.now() - then) / 1000));
  if (secs < 60) return 'just now';
  const mins = Math.round(secs / 60);
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.round(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  return `${days}d ago`;
}

// Stamp shows the provenance of a static export — branch, short commit, and how
// fresh it is — so a reader never mistakes a stale snapshot for live state. It
// renders nothing when there is no meta (i.e. the live `serve` UI).
export function Stamp({ meta }: { meta?: ReportMeta | null }) {
  if (!meta || (!meta.branch && !meta.commitShort && !meta.commit && !meta.generatedAt)) {
    return null;
  }
  const commit = meta.commitShort || (meta.commit ? meta.commit.slice(0, 7) : '');
  const rel = relativeTime(meta.generatedAt);
  return (
    <span className="stamp" data-testid="stamp">
      {meta.branch && <span className="stamp-branch">{meta.branch}</span>}
      {commit && <span className="stamp-commit">{commit}</span>}
      {rel && (
        <span className="stamp-time" title={meta.generatedAt}>
          {rel}
        </span>
      )}
    </span>
  );
}
