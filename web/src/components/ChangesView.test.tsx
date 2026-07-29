import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChangesView } from './ChangesView';
import type { DiffResponse } from '../types';

const withChanges: DiffResponse = {
  enabled: true,
  delta: {
    base: 'HEAD',
    changes: [
      { id: 'a-lost', title: 'A', kind: 'coverage-lost', detail: 'covered → uncovered', draft: false, regression: true },
      { id: 'b-added', title: 'B', kind: 'added', detail: 'new, covered', draft: false, regression: false },
      { id: 'c-gone', title: 'C', kind: 'removed', draft: false, regression: false },
      { id: 'd-impl', title: 'D', kind: 'impl-changed', detail: 'implementation changed', draft: false, regression: false, files: ['internal/hub/x.go'] },
    ],
    added: 1, removed: 1, coverageLost: 1, coverageGained: 0, edited: 0, implChanged: 1,
    regressions: 1, baseOk: true, headOk: false,
  },
};

describe('ChangesView', () => {
  it('shows a loading state until the diff arrives', () => {
    render(<ChangesView diff={null} loading onSelect={vi.fn()} />);
    expect(screen.getByTestId('changes-loading')).toBeInTheDocument();
  });

  it('explains when diff is disabled (no git / not enabled)', () => {
    render(<ChangesView diff={{ enabled: false }} loading={false} onSelect={vi.fn()} />);
    expect(screen.getByTestId('changes-disabled')).toBeInTheDocument();
  });

  it('reports an empty delta as no changes', () => {
    render(
      <ChangesView
        diff={{ enabled: true, delta: { base: 'HEAD', changes: [], added: 0, removed: 0, coverageLost: 0, coverageGained: 0, edited: 0, implChanged: 0, regressions: 0, baseOk: true, headOk: true } }}
        loading={false}
        onSelect={vi.fn()}
      />,
    );
    expect(screen.getByTestId('no-changes')).toBeInTheDocument();
  });

  it('surfaces a regression banner and one row per change', () => {
    render(<ChangesView diff={withChanges} loading={false} onSelect={vi.fn()} />);
    expect(screen.getByTestId('regression-banner')).toHaveTextContent('1 regression');
    expect(screen.getByTestId('change-row-a-lost')).toHaveAttribute('data-kind', 'coverage-lost');
    expect(screen.getByTestId('change-row-b-added')).toBeInTheDocument();
    expect(screen.getByTestId('change-row-d-impl')).toBeInTheDocument();
  });

  it('navigates to a still-present spec but not a removed one', async () => {
    const onSelect = vi.fn();
    render(<ChangesView diff={withChanges} loading={false} onSelect={onSelect} />);

    await userEvent.click(screen.getByTestId('change-row-b-added'));
    expect(onSelect).toHaveBeenCalledWith('b-added');

    // A removed spec no longer exists to open — it's not a button.
    const removed = screen.getByTestId('change-row-c-gone');
    expect(removed.tagName).not.toBe('BUTTON');
    onSelect.mockClear();
    await userEvent.click(removed);
    expect(onSelect).not.toHaveBeenCalled();
  });
});
