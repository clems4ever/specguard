import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FindingsPanel } from './FindingsPanel';

describe('FindingsPanel', () => {
  it('shows an all-clear message when empty', () => {
    render(<FindingsPanel findings={[]} />);
    expect(screen.getByTestId('findings')).toHaveTextContent('every spec is traced');
  });

  it('orders errors before warnings regardless of input order', () => {
    render(
      <FindingsPanel
        findings={[
          { severity: 'warning', rule: 'w', message: 'a warning' },
          { severity: 'error', rule: 'e', message: 'an error' },
        ]}
      />,
    );
    const items = screen.getAllByRole('listitem');
    expect(items[0]).toHaveTextContent('an error');
    expect(items[1]).toHaveTextContent('a warning');
  });
});
