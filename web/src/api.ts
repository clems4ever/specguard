import type { Report } from './types';

/** Fetch the live report from the Go server (proxied by Vite in dev). */
export async function fetchReport(signal?: AbortSignal): Promise<Report> {
  const res = await fetch('/api/report', { signal });
  if (!res.ok) {
    throw new Error(`report request failed: ${res.status}`);
  }
  return (await res.json()) as Report;
}
