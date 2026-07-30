import { useMemo, useState } from 'react';
import type { AreaInfo, Report, ReportMeta, SpecStatus, DiffResponse } from '../types';
import { areaRollup, filterSpecs, groupByArea, specState, summarize } from '../selectors';
import { StatusBadge } from './StatusBadge';
import { FindingsPanel } from './FindingsPanel';
import { ChangesView } from './ChangesView';
import { Intro } from './Intro';
import { Stamp } from './Stamp';

function SpecRow({
  spec,
  onSelect,
  hasResults,
}: {
  spec: SpecStatus;
  onSelect: (id: string) => void;
  hasResults: boolean;
}) {
  const tests = spec.tests ?? [];
  return (
    <button
      className="spec-row"
      data-testid={`spec-row-${spec.id}`}
      onClick={() => onSelect(spec.id)}
    >
      <StatusBadge state={specState(spec, hasResults)} />
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

// One area, collapsible, with a roll-up so a reader can judge it without
// expanding: the spec count and its worst state ("all passing" vs "1 failing").
function AreaSection({
  area,
  specs,
  onSelect,
  hasResults,
  info,
}: {
  area: string;
  specs: SpecStatus[];
  onSelect: (id: string) => void;
  hasResults: boolean;
  info?: AreaInfo;
}) {
  const [open, setOpen] = useState(true);
  const roll = areaRollup(specs, hasResults);
  return (
    <section id={`area-${area}`} className="area" data-testid={`area-${area}`}>
      <button
        className="area-header"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        data-testid={`area-toggle-${area}`}
      >
        <span className="area-chevron" aria-hidden>
          {open ? '▾' : '▸'}
        </span>
        <h2 className="area-title">{info?.title || area}</h2>
        <span className="area-count muted">
          {roll.total} spec{roll.total === 1 ? '' : 's'}
        </span>
        <StatusBadge state={roll.state} />
      </button>
      {info?.description && (
        <p className="area-desc muted" data-testid={`area-desc-${area}`}>
          {info.description}
        </p>
      )}
      {open && (
        <div className="spec-list">
          {specs.map((s) => (
            <SpecRow key={s.id} spec={s} onSelect={onSelect} hasResults={hasResults} />
          ))}
        </div>
      )}
    </section>
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
  refreshing = false,
  changedMode = false,
  onToggleChanged,
  diff = null,
  diffLoading = false,
  meta = null,
}: {
  report: Report;
  onSelect: (id: string) => void;
  onRefresh?: () => void;
  refreshing?: boolean;
  changedMode?: boolean;
  onToggleChanged?: () => void;
  diff?: DiffResponse | null;
  diffLoading?: boolean;
  meta?: ReportMeta | null;
}) {
  const [query, setQuery] = useState('');
  const summary = useMemo(() => summarize(report), [report]);
  // The headline reflects traceability AND, when a run was ingested, its
  // outcome: a covered-but-failing spec flips the verdict to FAIL.
  const pass = report.ok && !(summary.hasResults && summary.failing > 0);
  const groups = useMemo(
    () => groupByArea(filterSpecs(report.specs ?? [], query)),
    [report.specs, query],
  );
  const areaInfo = useMemo(
    () => new Map((report.areas ?? []).map((a) => [a.name, a])),
    [report.areas],
  );
  const totalShown = groups.reduce((n, g) => n + g.specs.length, 0);

  return (
    <div className="dashboard">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark">◉</span> specguard
          <Stamp meta={meta} />
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
          {onRefresh && (
            <button className="refresh" onClick={onRefresh} disabled={refreshing} data-testid="refresh">
              {refreshing ? 'Refreshing…' : 'Refresh'}
            </button>
          )}
        </div>
      </header>

      <div
        className={`banner ${pass ? 'banner-pass' : 'banner-fail'}`}
        data-testid="status-banner"
      >
        <span className="banner-verdict">{pass ? 'PASS' : 'FAIL'}</span>
        <span className="banner-detail">
          {summary.errors} error{summary.errors === 1 ? '' : 's'}, {summary.warnings} warning
          {summary.warnings === 1 ? '' : 's'}
          {summary.hasResults && (
            <>
              {', '}
              <strong>{summary.failing}</strong> failing
            </>
          )}
          {' across '}
          {summary.total} spec{summary.total === 1 ? '' : 's'}
        </span>
      </div>

      <div className="stats">
        <Stat label="Coverage" value={`${summary.coveragePct}%`} tone="accent" />
        {summary.hasResults ? (
          <>
            <Stat label="Passing" value={summary.passing} tone="covered" />
            <Stat label="Failing" value={summary.failing} tone="uncovered" />
            <Stat label="Not run" value={summary.skipped + summary.notRun} tone="draft" />
          </>
        ) : (
          <>
            <Stat label="Covered" value={summary.covered} tone="covered" />
            <Stat label="Uncovered" value={summary.uncovered} tone="uncovered" />
            <Stat label="Drafts" value={summary.drafts} tone="draft" />
          </>
        )}
        <Stat label="Test files" value={summary.testFiles} />
      </div>

      {changedMode ? (
        <ChangesView diff={diff} loading={diffLoading} onSelect={onSelect} />
      ) : (
        <>
          <Intro hasResults={summary.hasResults} />

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
            <>
              {groups.length > 1 && (
                <nav className="toc" data-testid="toc" aria-label="Areas">
                  {groups.map((g) => (
                    <a key={g.area} className="toc-chip" href={`#area-${g.area}`}>
                      {g.area}
                      <span className="toc-count">{g.specs.length}</span>
                    </a>
                  ))}
                </nav>
              )}

              {groups.map((g) => (
                <AreaSection
                  key={g.area}
                  area={g.area}
                  specs={g.specs}
                  onSelect={onSelect}
                  hasResults={summary.hasResults}
                  info={areaInfo.get(g.area)}
                />
              ))}
            </>
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
