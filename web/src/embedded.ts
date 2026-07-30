import type { Report, ReportMeta } from './types';

// When `specguard report` generates a static, self-contained HTML file, it
// injects the pre-computed report and its provenance onto `window` before the
// app boots. Their presence is what flips the app into offline "static" mode:
// it renders the embedded data instead of fetching /api/report from a server.
declare global {
  interface Window {
    __SPECGUARD_REPORT__?: Report;
    __SPECGUARD_META__?: ReportMeta;
  }
}

/** The report baked into a static export, or null when running against a server. */
export function embeddedReport(): Report | null {
  return typeof window !== 'undefined' && window.__SPECGUARD_REPORT__
    ? window.__SPECGUARD_REPORT__
    : null;
}

/** Provenance of a static export (branch/commit/time), or null when live. */
export function embeddedMeta(): ReportMeta | null {
  return typeof window !== 'undefined' && window.__SPECGUARD_META__
    ? window.__SPECGUARD_META__
    : null;
}
