import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SpecDetail } from './SpecDetail';
import { mixedReport, resultsReport, spec } from '../test/fixtures';
import type { ReportMeta } from '../types';

const login = mixedReport.specs.find((s) => s.id === 'auth-login')!;
const uncovered = mixedReport.specs.find((s) => s.id === 'tasks-create')!;

const withRefs = spec({
  id: 'auth-login',
  title: 'A user can log in',
  path: 'specs/auth/login.md',
  covered: true,
  tests: ['server/auth_test.go'],
  refs: [{ file: 'server/auth_test.go', line: 42 }],
});
const meta: ReportMeta = { repo: 'clems4ever/specguard', commit: 'abc123' };

describe('SpecDetail', () => {
  it('renders title, id and covering tests', () => {
    render(<SpecDetail spec={login} report={mixedReport} onBack={() => {}} />);
    expect(screen.getByTestId('detail-title')).toHaveTextContent('A user can log in');
    const tests = screen.getByTestId('test-list');
    expect(within(tests).getByText('server/auth_test.go')).toBeInTheDocument();
    expect(within(tests).getByText('web/e2e/auth.spec.ts')).toBeInTheDocument();
  });

  it('renders the covers list', () => {
    render(<SpecDetail spec={login} report={mixedReport} onBack={() => {}} />);
    expect(within(screen.getByTestId('covers-list')).getByText('server/auth.go')).toBeInTheDocument();
  });

  it('renders the markdown body as html', () => {
    render(<SpecDetail spec={login} report={mixedReport} onBack={() => {}} />);
    const md = screen.getByTestId('markdown');
    expect(md.querySelector('h2')).not.toBeNull();
    expect(md).toHaveTextContent('Valid credentials return a session.');
  });

  it('surfaces the spec findings on an uncovered spec', () => {
    render(<SpecDetail spec={uncovered} report={mixedReport} onBack={() => {}} />);
    expect(screen.getByTestId('finding-uncovered-spec')).toBeInTheDocument();
  });

  it('calls onBack from the back button', async () => {
    const onBack = vi.fn();
    render(<SpecDetail spec={login} report={mixedReport} onBack={onBack} />);
    await userEvent.click(screen.getByTestId('back'));
    expect(onBack).toHaveBeenCalled();
  });

  it('links the spec source and each covering test to GitHub at the pinned commit', () => {
    render(<SpecDetail spec={withRefs} report={mixedReport} onBack={() => {}} meta={meta} />);
    const source = screen.getByTestId('spec-source-link') as HTMLAnchorElement;
    expect(source.tagName).toBe('A');
    expect(source).toHaveAttribute(
      'href',
      'https://github.com/clems4ever/specguard/blob/abc123/specs/auth/login.md',
    );
    const testRef = within(screen.getByTestId('test-list')).getByText('server/auth_test.go:42');
    const anchor = testRef.closest('a')!;
    expect(anchor).toHaveAttribute(
      'href',
      'https://github.com/clems4ever/specguard/blob/abc123/server/auth_test.go#L42',
    );
  });

  it('falls back to plain text when there is no commit (live serve)', () => {
    render(<SpecDetail spec={withRefs} report={mixedReport} onBack={() => {}} />);
    expect(screen.getByTestId('spec-source-link').tagName).toBe('CODE');
    // The label still carries the line, just not a link.
    expect(within(screen.getByTestId('test-list')).getByText('server/auth_test.go:42')).toBeInTheDocument();
  });

  it('shows the outcome badge and per-test status when results are present', () => {
    const failing = resultsReport.specs.find((s) => s.id === 'auth-logout')!;
    render(<SpecDetail spec={failing} report={resultsReport} onBack={() => {}} />);
    // Head badge reflects the failing outcome.
    expect(screen.getByTestId('badge-failing')).toBeInTheDocument();
    // The covering test shows a failed dot.
    expect(within(screen.getByTestId('test-list')).getByTestId('ref-status-failed')).toBeInTheDocument();
  });
});
