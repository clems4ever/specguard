import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Dashboard } from './Dashboard';
import { mixedReport, resultsReport } from '../test/fixtures';

function renderDash(overrides = {}) {
  const onSelect = vi.fn();
  const onRefresh = vi.fn();
  render(
    <Dashboard
      report={mixedReport}
      onSelect={onSelect}
      onRefresh={onRefresh}
      refreshing={false}
      {...overrides}
    />,
  );
  return { onSelect, onRefresh };
}

describe('Dashboard', () => {
  it('shows the FAIL banner with error and warning counts', () => {
    renderDash();
    const banner = screen.getByTestId('status-banner');
    expect(banner).toHaveTextContent('FAIL');
    expect(banner).toHaveTextContent('1 error');
    expect(banner).toHaveTextContent('2 warnings');
  });

  it('renders coverage and count stats', () => {
    renderDash();
    expect(screen.getByText('67%')).toBeInTheDocument(); // coverage
  });

  it('groups specs by area with sorted headers', () => {
    renderDash();
    expect(screen.getByTestId('area-auth')).toBeInTheDocument();
    expect(screen.getByTestId('area-tasks')).toBeInTheDocument();
    expect(screen.getByTestId('area-sharing')).toBeInTheDocument();
  });

  it('shows the right badge state per spec', () => {
    renderDash();
    const login = screen.getByTestId('spec-row-auth-login');
    expect(within(login).getByTestId('badge-covered')).toBeInTheDocument();
    const logout = screen.getByTestId('spec-row-auth-logout');
    expect(within(logout).getByTestId('badge-warning')).toBeInTheDocument();
    const create = screen.getByTestId('spec-row-tasks-create');
    expect(within(create).getByTestId('badge-uncovered')).toBeInTheDocument();
    const draft = screen.getByTestId('spec-row-sharing-permissions');
    expect(within(draft).getByTestId('badge-draft')).toBeInTheDocument();
  });

  it('filters specs by the search box', async () => {
    renderDash();
    await userEvent.type(screen.getByTestId('search'), 'login');
    expect(screen.getByTestId('spec-row-auth-login')).toBeInTheDocument();
    expect(screen.queryByTestId('spec-row-tasks-create')).not.toBeInTheDocument();
  });

  it('shows an empty state when nothing matches', async () => {
    renderDash();
    await userEvent.type(screen.getByTestId('search'), 'zzzzz');
    expect(screen.getByTestId('no-match')).toBeInTheDocument();
  });

  it('calls onSelect when a spec row is clicked', async () => {
    const { onSelect } = renderDash();
    await userEvent.click(screen.getByTestId('spec-row-auth-login'));
    expect(onSelect).toHaveBeenCalledWith('auth-login');
  });

  it('calls onRefresh when Refresh is clicked', async () => {
    const { onRefresh } = renderDash();
    await userEvent.click(screen.getByTestId('refresh'));
    expect(onRefresh).toHaveBeenCalled();
  });

  it('lists all findings, errors first', () => {
    renderDash();
    const findings = screen.getByTestId('findings');
    const items = within(findings).getAllByRole('listitem');
    expect(items[0]).toHaveTextContent('uncovered-spec'); // error before warnings
  });

  it('offers a Changed-only toggle and fires it', async () => {
    const onToggleChanged = vi.fn();
    renderDash({ onToggleChanged });
    const toggle = screen.getByTestId('toggle-changed');
    expect(toggle).toHaveTextContent('Changed only');
    await userEvent.click(toggle);
    expect(onToggleChanged).toHaveBeenCalled();
  });

  it('shows pass/fail state when a test run was ingested', () => {
    render(<Dashboard report={resultsReport} onSelect={vi.fn()} onRefresh={vi.fn()} />);
    // Verdict flips to FAIL because a covered spec is failing (traceability ok).
    const banner = screen.getByTestId('status-banner');
    expect(banner).toHaveTextContent('FAIL');
    expect(banner).toHaveTextContent('1 failing');
    // Passing/Failing stat tiles replace Covered/Uncovered.
    expect(screen.getByText('Passing', { selector: '.stat-label' })).toBeInTheDocument();
    expect(screen.getByText('Failing', { selector: '.stat-label' })).toBeInTheDocument();
    // The failing spec's row shows the failing badge, the passing one passing.
    expect(
      within(screen.getByTestId('spec-row-auth-logout')).getByTestId('badge-failing'),
    ).toBeInTheDocument();
    expect(
      within(screen.getByTestId('spec-row-auth-login')).getByTestId('badge-passing'),
    ).toBeInTheDocument();
  });

  it('swaps the area list for the changed view when changedMode is on', () => {
    renderDash({
      changedMode: true,
      onToggleChanged: vi.fn(),
      diff: {
        enabled: true,
        delta: {
          base: 'HEAD',
          changes: [
            { id: 'x-lost', title: 'X', kind: 'coverage-lost', detail: 'covered → uncovered', draft: false, regression: true },
          ],
          added: 0, removed: 0, coverageLost: 1, coverageGained: 0, edited: 0, implChanged: 0,
          regressions: 1, baseOk: true, headOk: false,
        },
      },
    });
    // The changed view is shown; the full area grouping is not.
    expect(screen.getByTestId('change-row-x-lost')).toBeInTheDocument();
    expect(screen.queryByTestId('area-auth')).not.toBeInTheDocument();
  });
});
