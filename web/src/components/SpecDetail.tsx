import type { Report, SpecStatus } from '../types';
import { specArea, specState, findingsForSpec } from '../selectors';
import { StatusBadge } from './StatusBadge';
import { Markdown } from './Markdown';
import { FindingsPanel } from './FindingsPanel';

export function SpecDetail({
  spec,
  report,
  onBack,
}: {
  spec: SpecStatus;
  report: Report;
  onBack: () => void;
}) {
  const tests = spec.tests ?? [];
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
      <div className="detail-path">{spec.path}</div>

      <section className="detail-meta">
        <div>
          <h3>Covering tests ({tests.length})</h3>
          {tests.length ? (
            <ul className="test-list" data-testid="test-list">
              {tests.map((t) => (
                <li key={t}>
                  <code>{t}</code>
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
