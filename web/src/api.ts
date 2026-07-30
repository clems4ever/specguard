import type { Report, DiffResponse } from './types';

/** Fetch the live report from the Go server (proxied by Vite in dev). */
export async function fetchReport(signal?: AbortSignal): Promise<Report> {
  const res = await fetch('/api/report', { signal });
  if (!res.ok) {
    throw new Error(`report request failed: ${res.status}`);
  }
  return (await res.json()) as Report;
}

/** Fetch the "what changed" delta against a base ref (default: server's). */
export async function fetchDiff(base?: string, signal?: AbortSignal): Promise<DiffResponse> {
  const q = base ? `?base=${encodeURIComponent(base)}` : '';
  const res = await fetch(`/api/diff${q}`, { signal });
  if (!res.ok) {
    throw new Error(`diff request failed: ${res.status}`);
  }
  return (await res.json()) as DiffResponse;
}
