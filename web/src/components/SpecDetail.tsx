import type { Ref, Report, ReportMeta, SpecStatus, TestStatus } from '../types';
import { specArea, specState, findingsForSpec, buildTree, subtreeRollup } from '../selectors';
import type { SpecNode } from '../selectors';
import { codeLink, refLabel } from '../links';
import { StatusBadge } from './StatusBadge';
import { Markdown } from './Markdown';
import { FindingsPanel } from './FindingsPanel';
import { Gallery } from './Gallery';

const RESULT_GLYPH: Record<TestStatus, string> = { passed: '✓', failed: '✗', skipped: '–' };

// A small pass/fail dot shown next to a covering test once results are ingested.
function RefStatus({ status }: { status?: TestStatus }) {
  if (!status) return null;
  return (
    <span
      className={`ref-status ref-status-${status}`}
      data-testid={`ref-status-${status}`}
      title={status}
      aria-label={status}
    >
      {RESULT_GLYPH[status]}
    </span>
  );
}

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

// findNode locates a spec's node in the derivation forest, so the detail view
// can show its refinements and a subtree roll-up.
function findNode(forest: SpecNode[], id: string): SpecNode | undefined {
  for (const n of forest) {
    if (n.spec.id === id) return n;
    const hit = findNode(n.children, id);
    if (hit) return hit;
  }
  return undefined;
}

export function SpecDetail({
  spec,
  report,
  onBack,
  onSelect,
  meta = null,
}: {
  spec: SpecStatus;
  report: Report;
  onBack: () => void;
  onSelect?: (id: string) => void;
  meta?: ReportMeta | null;
}) {
  // Prefer precise (file, line) refs; fall back to the distinct-file list for
  // reports generated before refs existed.
  const refs: Ref[] = spec.refs ?? (spec.tests ?? []).map((file) => ({ file, line: 0 }));
  const findings = findingsForSpec(report, spec.id);
  const specs = report.specs ?? [];
  const parent = spec.parent ? specs.find((s) => s.id === spec.parent) : undefined;
  const node = findNode(buildTree(specs), spec.id);
  const children = node?.children ?? [];
  // A parent's badge reflects its subtree (green only when all refinements are
  // proven), not its own possibly test-less state.
  const headState =
    children.length > 0 ? subtreeRollup(node!, report.hasResults).state : specState(spec, report.hasResults);
  return (
    <article className="detail" data-testid="spec-detail">
      <button className="back" onClick={onBack} data-testid="back">
        ← All specs
      </button>

      {parent && (
        <button
          className="refines"
          data-testid="refines"
          onClick={() => onSelect?.(parent.id)}
          title={`Refines ${parent.id}`}
        >
          ↑ Refines <strong>{parent.title}</strong>
        </button>
      )}

      <div className="detail-head" data-testid="detail-head">
        <span className="area-tag">{specArea(spec)}</span>
        <StatusBadge state={headState} />
      </div>
      <h1 data-testid="detail-title">{spec.title}</h1>
      <code className="spec-id">{spec.id}</code>
      <div className="detail-path" data-testid="detail-path">
        <CodeRef meta={meta} path={spec.path} testid="spec-source-link" />
      </div>

      {children.length > 0 && (
        <section className="detail-children" data-testid="detail-children">
          <h3>Refined by ({children.length})</h3>
          <ul className="children-list">
            {children.map((c) => (
              <li key={c.spec.id}>
                <button
                  className="child-link"
                  data-testid={`child-link-${c.spec.id}`}
                  onClick={() => onSelect?.(c.spec.id)}
                >
                  <StatusBadge state={subtreeRollup(c, report.hasResults).state} />
                  <span>{c.spec.title}</span>
                  <code>{c.spec.id}</code>
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className="detail-meta">
        <div>
          <h3>Covering tests ({refs.length})</h3>
          {refs.length ? (
            <ul className="test-list" data-testid="test-list">
              {refs.map((r) => (
                <li key={`${r.file}:${r.line}`}>
                  <RefStatus status={r.status} />
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

      <section className="detail-gallery">
        <Gallery artifacts={spec.artifacts} />
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
