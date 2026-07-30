import type { Report, ReportMeta, SpecStatus } from '../types';
import { specArea, specState, findingsForSpec } from '../selectors';
import { codeLink, refLabel } from '../links';
import { StatusBadge } from './StatusBadge';
import { Markdown } from './Markdown';
import { FindingsPanel } from './FindingsPanel';

// CodeRef renders a path (optionally with a line) as a GitHub link when the
// report carries repo+commit, and as plain text otherwise (e.g. live `serve`).
function CodeRef({
  meta,
  path,
  line,
  testid,
}: {
  meta: ReportMeta | null | undefined;
  path: string;
  line?: number;
  testid?: string;
}) {
  const href = codeLink(meta, path, line);
  const label = refLabel(path, line);
  if (!href) {
    return (
      <code data-testid={testid}>{label}</code>
    );
  }
  return (
    <a className="code-link" href={href} target="_blank" rel="noreferrer" data-testid={testid}>
      <code>{label}</code>
      <span className="code-link-mark" aria-hidden="true"> ↗</span>
    </a>
  );
}

export function SpecDetail({
  spec,
  report,
  onBack,
  meta = null,
}: {
  spec: SpecStatus;
  report: Report;
  onBack: () => void;
  meta?: ReportMeta | null;
}) {
  // Prefer precise (file, line) refs; fall back to the distinct-file list for
  // reports generated before refs existed.
  const refs = spec.refs ?? (spec.tests ?? []).map((file) => ({ file, line: 0 }));
  const findings = findingsForSpec(report, spec.id);
  return (
    <article className="detail" data-testid="spec-detail">
      <button className="back" onClick={onBack} data-testid="back">
        ← All specs
      </button>

      <div className="detail-head">
        <span className="area-tag">{specArea(spec)}</span>
        <StatusBadge state={specState(spec)} />
      </div>
      <h1 data-testid="detail-title">{spec.title}</h1>
      <code className="spec-id">{spec.id}</code>
      <div className="detail-path" data-testid="detail-path">
        <CodeRef meta={meta} path={spec.path} testid="spec-source-link" />
      </div>

      <section className="detail-meta">
        <div>
          <h3>Covering tests ({refs.length})</h3>
          {refs.length ? (
            <ul className="test-list" data-testid="test-list">
              {refs.map((r) => (
                <li key={`${r.file}:${r.line}`}>
                  <CodeRef meta={meta} path={r.file} line={r.line} />
                </li>
              ))}
            </ul>
          ) : (
            <p className="muted">None yet.</p>
          )}
        </div>
        {spec.covers && spec.covers.length > 0 && (
          <div>
            <h3>Covers</h3>
            <ul className="covers-list" data-testid="covers-list">
              {spec.covers.map((c) => (
                <li key={c}>
                  <code>{c}</code>
                </li>
              ))}
            </ul>
          </div>
        )}
      </section>

      {findings.length > 0 && (
        <section>
          <h3>Findings</h3>
          <FindingsPanel findings={findings} />
        </section>
      )}

      <section className="detail-body">
        <Markdown source={spec.body ?? '_No description provided._'} />
      </section>
    </article>
  );
}
