import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SpecDetail } from './SpecDetail';
import { mixedReport } from '../test/fixtures';

const login = mixedReport.specs.find((s) => s.id === 'auth-login')!;
const uncovered = mixedReport.specs.find((s) => s.id === 'tasks-create')!;

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
});
