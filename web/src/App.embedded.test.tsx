import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { Report, ReportMeta } from './types';

// A static export bakes the report onto window before the app boots. These tests
// assert the app renders that data offline — with no fetch, no refresh, and a
// provenance stamp — which is the whole contract of `specguard report`.

const report: Report = {
  ok: true,
  testFiles: 3,
  findings: [],
  specs: [
    {
      id: 'auth-login',
      title: 'Log in',
      path: 'specs/auth/login.md',
      tests: ['server/auth_test.go'],
      covered: true,
      coversOk: true,
      draft: false,
    },
  ],
};

const meta: ReportMeta = {
  branch: 'main',
  commit: 'deadbeefcafe',
  commitShort: 'deadbee',
  generatedAt: new Date().toISOString(),
};

describe('App (embedded / static export)', () => {
  beforeEach(() => {
    vi.resetModules();
    window.__SPECGUARD_REPORT__ = report;
    window.__SPECGUARD_META__ = meta;
    window.history.pushState({}, '', '/');
  });

  afterEach(() => {
    delete window.__SPECGUARD_REPORT__;
    delete window.__SPECGUARD_META__;
    vi.restoreAllMocks();
  });

  it('renders the embedded report without fetching', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch');
    const { App } = await import('./App');
    render(<App />);

    // Data is on screen immediately from the embed — no loading state, no fetch.
    expect(screen.getByTestId('status-banner')).toHaveTextContent('PASS');
    expect(screen.getByTestId('spec-row-auth-login')).toBeInTheDocument();
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it('shows the provenance stamp (branch + commit)', async () => {
    const { App } = await import('./App');
    render(<App />);
    const stamp = screen.getByTestId('stamp');
    expect(stamp).toHaveTextContent('main');
    expect(stamp).toHaveTextContent('deadbee');
  });

  it('withholds refresh and changed-only in static mode', async () => {
    const { App } = await import('./App');
    render(<App />);
    expect(screen.queryByTestId('refresh')).not.toBeInTheDocument();
    expect(screen.queryByTestId('toggle-changed')).not.toBeInTheDocument();
  });
});
