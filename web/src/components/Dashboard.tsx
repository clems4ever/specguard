import { useMemo, useState } from 'react';
import type { Report, SpecStatus, DiffResponse } from '../types';
import { filterSpecs, groupByArea, specState, summarize } from '../selectors';
import { StatusBadge } from './StatusBadge';
import { FindingsPanel } from './FindingsPanel';
import { ChangesView } from './ChangesView';

function SpecRow({ spec, onSelect }: { spec: SpecStatus; onSelect: (id: string) => void }) {
  const tests = spec.tests ?? [];
  return (
    <button
      className="spec-row"
      data-testid={`spec-row-${spec.id}`}
      onClick={() => onSelect(spec.id)}
    >
      <StatusBadge state={specState(spec)} />
      <span className="spec-row-main">
        <span className="spec-row-title">{spec.title}</span>
        <code className="spec-row-id">{spec.id}</code>
      </span>
      <span className="spec-row-tests muted">
        {tests.length ? `${tests.length} test${tests.length > 1 ? 's' : ''}` : '—'}
      </span>
    </button>
  );
}

function Stat({ label, value, tone }: { label: string; value: number | string; tone?: string }) {
  return (
    <div className={`stat ${tone ? `stat-${tone}` : ''}`}>
      <div className="stat-value">{value}</div>
      <div className="stat-label">{label}</div>
    </div>
  );
}

export function Dashboard({
  report,
  onSelect,
  onRefresh,
  refreshing,
  changedMode = false,
  onToggleChanged,
  diff = null,
  diffLoading = false,
}: {
  report: Report;
  onSelect: (id: string) => void;
  onRefresh: () => void;
  refreshing: boolean;
  changedMode?: boolean;
  onToggleChanged?: () => void;
  diff?: DiffResponse | null;
  diffLoading?: boolean;
}) {
  const [query, setQuery] = useState('');
  const summary = useMemo(() => summarize(report), [report]);
  const groups = useMemo(
    () => groupByArea(filterSpecs(report.specs ?? [], query)),
    [report.specs, query],
  );
  const totalShown = groups.reduce((n, g) => n + g.specs.length, 0);

  return (
    <div className="dashboard">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark">◉</span> specguard
        </div>
        <div className="topbar-actions">
          {onToggleChanged && (
            <button
              className={`toggle ${changedMode ? 'toggle-on' : ''}`}
              onClick={onToggleChanged}
              data-testid="toggle-changed"
              aria-pressed={changedMode}
              title="Show only the specs this change touched"
            >
              {changedMode ? 'All specs' : 'Changed only'}
            </button>
          )}
          <button className="refresh" onClick={onRefresh} disabled={refreshing} data-testid="refresh">
            {refreshing ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </header>

      <div
        className={`banner ${report.ok ? 'banner-pass' : 'banner-fail'}`}
        data-testid="status-banner"
      >
        <span className="banner-verdict">{report.ok ? 'PASS' : 'FAIL'}</span>
        <span className="banner-detail">
          {summary.errors} error{summary.errors === 1 ? '' : 's'}, {summary.warnings} warning
          {summary.warnings === 1 ? '' : 's'} across {summary.total} spec
          {summary.total === 1 ? '' : 's'}
        </span>
      </div>

      <div className="stats">
        <Stat label="Coverage" value={`${summary.coveragePct}%`} tone="accent" />
        <Stat label="Covered" value={summary.covered} tone="covered" />
        <Stat label="Uncovered" value={summary.uncovered} tone="uncovered" />
        <Stat label="Drafts" value={summary.drafts} tone="draft" />
        <Stat label="Test files" value={summary.testFiles} />
      </div>

      {changedMode ? (
        <ChangesView diff={diff} loading={diffLoading} onSelect={onSelect} />
      ) : (
        <>
          <input
            className="search"
            data-testid="search"
            placeholder="Filter specs by id, title or area…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />

          {totalShown === 0 ? (
            <p className="empty" data-testid="no-match">
              No specs match “{query}”.
            </p>
          ) : (
            groups.map((g) => (
              <section key={g.area} className="area" data-testid={`area-${g.area}`}>
                <h2 className="area-title">{g.area}</h2>
                <div className="spec-list">
                  {g.specs.map((s) => (
                    <SpecRow key={s.id} spec={s} onSelect={onSelect} />
                  ))}
                </div>
              </section>
            ))
          )}

          <section className="findings-section">
            <h2>Findings</h2>
            <FindingsPanel findings={report.findings ?? []} />
          </section>
        </>
      )}
    </div>
  );
}
