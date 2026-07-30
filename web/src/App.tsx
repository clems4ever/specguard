import { useCallback, useEffect, useState } from 'react';
import type { Report, DiffResponse } from './types';
import { fetchReport, fetchDiff } from './api';
import { embeddedReport, embeddedMeta } from './embedded';
import { Dashboard } from './components/Dashboard';
import { SpecDetail } from './components/SpecDetail';
import { basePath } from './base';

/** Read the selected spec id from a `<base>spec/<id>` path, base-prefix aware. */
function specIdFromPath(path: string): string | null {
  const m = path.match(/\/spec\/([^/]+)\/?$/);
  return m ? decodeURIComponent(m[1]) : null;
}

// A static export (from `specguard report`) bakes the report onto `window`. When
// present we render it offline: no server, no fetch, no live diff/refresh.
const EMBEDDED = embeddedReport();
const META = embeddedMeta();
const STATIC = EMBEDDED !== null;

export function App() {
  const [report, setReport] = useState<Report | null>(EMBEDDED);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [changedMode, setChangedMode] = useState(false);
  const [diff, setDiff] = useState<DiffResponse | null>(null);
  const [diffLoading, setDiffLoading] = useState(false);
  const [selectedId, setSelectedId] = useState<string | null>(
    specIdFromPath(window.location.pathname),
  );

  const load = useCallback(async (signal?: AbortSignal) => {
    setRefreshing(true);
    try {
      const rep = await fetchReport(signal);
      setReport(rep);
      setError(null);
    } catch (e) {
      if ((e as Error).name !== 'AbortError') setError((e as Error).message);
    } finally {
      setRefreshing(false);
    }
  }, []);

  const loadDiff = useCallback(async (signal?: AbortSignal) => {
    setDiffLoading(true);
    try {
      setDiff(await fetchDiff(undefined, signal));
    } catch (e) {
      if ((e as Error).name !== 'AbortError') {
        setDiff({ enabled: true, error: (e as Error).message });
      }
    } finally {
      setDiffLoading(false);
    }
  }, []);

  // Toggle the "Changed only" view, lazily fetching the delta the first time.
  const toggleChanged = useCallback(() => {
    setChangedMode((m) => {
      const next = !m;
      if (next) loadDiff();
      return next;
    });
  }, [loadDiff]);

  // Refresh re-reads the report and, in changed mode, recomputes the delta — so
  // edits on disk show up live.
  const refresh = useCallback(() => {
    load();
    if (changedMode) loadDiff();
  }, [load, loadDiff, changedMode]);

  useEffect(() => {
    if (STATIC) return; // embedded report: nothing to fetch
    const ctrl = new AbortController();
    load(ctrl.signal);
    return () => ctrl.abort();
  }, [load]);

  // Keep selection in sync with browser back/forward.
  useEffect(() => {
    const onPop = () => setSelectedId(specIdFromPath(window.location.pathname));
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  // pushState throws on the opaque file:// origin (a report opened as a local
  // file), so we guard it: navigation still works, the URL just doesn't update.
  // On http(s)/Pages the URL updates as before.
  const navigate = useCallback((path: string) => {
    try {
      window.history.pushState({}, '', path);
    } catch {
      /* file:// — keep in-app navigation working without a URL change */
    }
  }, []);

  const select = useCallback(
    (id: string) => {
      navigate(`${basePath()}spec/${encodeURIComponent(id)}`);
      setSelectedId(id);
    },
    [navigate],
  );

  const back = useCallback(() => {
    navigate(basePath());
    setSelectedId(null);
  }, [navigate]);

  if (error) {
    return (
      <div className="state state-error" data-testid="error">
        <p>Could not load the report: {error}</p>
        <p className="muted">Is `specguard serve` running?</p>
        <button onClick={() => load()}>Retry</button>
      </div>
    );
  }
  if (!report) {
    return (
      <div className="state" data-testid="loading">
        Loading report…
      </div>
    );
  }

  const selected = selectedId
    ? (report.specs ?? []).find((s) => s.id === selectedId)
    : undefined;

  if (selectedId && selected) {
    return (
      <SpecDetail spec={selected} report={report} onBack={back} onSelect={select} meta={META} />
    );
  }
  // In a static export there is no server to refresh from and no live git to
  // diff against, so both affordances are withheld (the props are omitted).
  return (
    <Dashboard
      report={report}
      onSelect={select}
      onRefresh={STATIC ? undefined : refresh}
      refreshing={refreshing}
      changedMode={changedMode}
      onToggleChanged={STATIC ? undefined : toggleChanged}
      diff={diff}
      diffLoading={diffLoading}
      meta={META}
    />
  );
}
