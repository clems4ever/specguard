import type { DiffResponse, DiffChange } from '../types';
import { KIND_LABEL, KIND_GLYPH } from '../selectors';

function ChangeRow({
  change,
  onSelect,
}: {
  change: DiffChange;
  onSelect: (id: string) => void;
}) {
  const removed = change.kind === 'removed';
  const cls = `change-row change-${change.kind}${change.regression ? ' change-regression' : ''}`;
  const body = (
    <>
      <span className={`change-glyph kind-${change.kind}`} aria-hidden>
        {KIND_GLYPH[change.kind]}
      </span>
      <span className="change-main">
        <span className="change-title">{change.title}</span>
        <code className="change-id">{change.id}</code>
      </span>
      <span className="change-meta">
        <span className="change-kind">{KIND_LABEL[change.kind]}</span>
        {change.detail && <span className="change-detail muted">{change.detail}</span>}
        {change.draft && <span className="tag tag-draft">draft</span>}
        {change.regression && <span className="tag tag-regression">regression</span>}
      </span>
    </>
  );
  if (removed) {
    return (
      <div className={cls} data-testid={`change-row-${change.id}`} data-kind={change.kind}>
        {body}
      </div>
    );
  }
  return (
    <button
      className={cls}
      data-testid={`change-row-${change.id}`}
      data-kind={change.kind}
      onClick={() => onSelect(change.id)}
    >
      {body}
    </button>
  );
}

/** The "Changed only" view: the delta of the working tree vs a base ref. */
export function ChangesView({
  diff,
  loading,
  onSelect,
}: {
  diff: DiffResponse | null;
  loading: boolean;
  onSelect: (id: string) => void;
}) {
  if (loading || !diff) {
    return (
      <div className="state" data-testid="changes-loading">
        Computing changes…
      </div>
    );
  }
  if (!diff.enabled) {
    return (
      <p className="empty" data-testid="changes-disabled">
        The changed-only view needs a git repository. Start the server with a{' '}
        <code>--diff-base</code> ref.
      </p>
    );
  }
  if (diff.error) {
    return (
      <p className="empty" data-testid="changes-error">
        Could not compute the diff: {diff.error}
      </p>
    );
  }
  const delta = diff.delta;
  const changes = delta?.changes ?? [];
  const base = delta?.base ?? diff.base ?? 'base';

  if (changes.length === 0) {
    return (
      <p className="empty" data-testid="no-changes">
        No spec-relevant changes vs <code>{base}</code>. ✅
      </p>
    );
  }
  return (
    <div className="changes">
      {delta && delta.regressions > 0 && (
        <div className="banner banner-fail" data-testid="regression-banner">
          <span className="banner-verdict">⚠ {delta.regressions} regression{delta.regressions === 1 ? '' : 's'}</span>
          <span className="banner-detail">a spec lost coverage or arrived uncovered</span>
        </div>
      )}
      <p className="changes-caption muted" data-testid="changes-caption">
        {changes.length} spec{changes.length === 1 ? '' : 's'} changed vs <code>{base}</code>
      </p>
      <div className="change-list">
        {changes.map((c) => (
          <ChangeRow key={c.id} change={c} onSelect={onSelect} />
        ))}
      </div>
    </div>
  );
}
