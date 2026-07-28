import { useCallback, useEffect, useState } from 'react';
import type { Report } from './types';
import { fetchReport } from './api';
import { Dashboard } from './components/Dashboard';
import { SpecDetail } from './components/SpecDetail';

/** Read the selected spec id from the path `/spec/<id>`. */
function specIdFromPath(path: string): string | null {
  const m = path.match(/^\/spec\/([^/]+)$/);
  return m ? decodeURIComponent(m[1]) : null;
}

export function App() {
  const [report, setReport] = useState<Report | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
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

  useEffect(() => {
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

  const select = useCallback((id: string) => {
    window.history.pushState({}, '', `/spec/${encodeURIComponent(id)}`);
    setSelectedId(id);
  }, []);

  const back = useCallback(() => {
    window.history.pushState({}, '', '/');
    setSelectedId(null);
  }, []);

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
    return <SpecDetail spec={selected} report={report} onBack={back} />;
  }
  return (
    <Dashboard report={report} onSelect={select} onRefresh={() => load()} refreshing={refreshing} />
  );
}
