import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Intro } from './Intro';

beforeEach(() => window.localStorage.clear());

describe('Intro', () => {
  it('explains the product and shows a coverage legend by default', () => {
    render(<Intro />);
    expect(screen.getByTestId('intro')).toHaveTextContent('specguard');
    const legend = screen.getByTestId('intro-legend');
    expect(legend).toHaveTextContent('Covered');
    expect(legend).toHaveTextContent('Uncovered');
  });

  it('shows pass/fail legend when a run was ingested', () => {
    render(<Intro hasResults />);
    const legend = screen.getByTestId('intro-legend');
    expect(legend).toHaveTextContent('Passing');
    expect(legend).toHaveTextContent('Failing');
  });

  it('can be dismissed and stays dismissed', async () => {
    const { unmount } = render(<Intro />);
    await userEvent.click(screen.getByTestId('intro-dismiss'));
    expect(screen.queryByTestId('intro')).not.toBeInTheDocument();

    // A fresh mount honours the persisted dismissal.
    unmount();
    render(<Intro />);
    expect(screen.queryByTestId('intro')).not.toBeInTheDocument();
  });
});
