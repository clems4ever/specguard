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

  it('links to the parent and lists refinements, navigating on click', async () => {
    const onSelect = vi.fn();
    const report = {
      ok: true,
      testFiles: 1,
      findings: [],
      specs: [
        spec({ id: 'cap', title: 'A capability', path: 'specs/x/cap.md', hasChild: true }),
        spec({
          id: 'cap-leaf',
          title: 'A refinement',
          path: 'specs/x/leaf.md',
          parent: 'cap',
          covered: true,
          tests: ['x_test.go'],
        }),
      ],
    };
    const parent = report.specs[0];
    render(<SpecDetail spec={parent} report={report} onBack={() => {}} onSelect={onSelect} />);
    // A parent with a covered child reads as satisfied, not uncovered.
    expect(within(screen.getByTestId('detail-head')).getByTestId('badge-covered')).toBeInTheDocument();
    // Its refinement is listed and navigates on click.
    await userEvent.click(screen.getByTestId('child-link-cap-leaf'));
    expect(onSelect).toHaveBeenCalledWith('cap-leaf');

    // The child shows a breadcrumb back up to its parent.
    const child = report.specs[1];
    render(<SpecDetail spec={child} report={report} onBack={() => {}} onSelect={onSelect} />);
    await userEvent.click(screen.getByTestId('refines'));
    expect(onSelect).toHaveBeenCalledWith('cap');
  });

  it('shows PM acceptance state (accepted / awaiting / stale)', () => {
    const mk = (lifecycle: 'accepted' | 'implemented' | 'stale') =>
      spec({
        id: 'auth-login',
        title: 'Log in',
        path: 'specs/auth/login.md',
        covered: true,
        tests: ['a_test.go'],
        fingerprint: 'abc123',
        lifecycle,
        acceptedBy: lifecycle === 'accepted' ? 'pm@acme' : undefined,
        acceptedAt: lifecycle === 'accepted' ? '2026-08-01T00:00:00Z' : undefined,
      });
    const rep = { ok: true, testFiles: 1, findings: [], specs: [] };

    const { unmount } = render(<SpecDetail spec={mk('accepted')} report={rep} onBack={() => {}} />);
    expect(screen.getByTestId('acceptance')).toHaveTextContent('accepted by pm@acme');
    unmount();

    render(<SpecDetail spec={mk('implemented')} report={rep} onBack={() => {}} />);
    expect(screen.getByTestId('acceptance')).toHaveTextContent('Awaiting PM');
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

  it('renders a screenshot gallery when a spec has artifacts', () => {
    const withShots = spec({
      id: 'auth-login',
      title: 'A user can log in',
      path: 'specs/auth/login.md',
      covered: true,
      artifacts: [{ name: 'login', path: 'assets/specs/auth-login/0.png' }],
    });
    render(<SpecDetail spec={withShots} report={mixedReport} onBack={() => {}} />);
    expect(screen.getByTestId('gallery')).toBeInTheDocument();
    expect(screen.getByTestId('gallery-thumb-0')).toBeInTheDocument();
  });
});
