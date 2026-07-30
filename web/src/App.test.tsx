import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from './App';
import { mixedReport } from './test/fixtures';
import * as api from './api';

vi.mock('./api');

const fetchReport = vi.mocked(api.fetchReport);

beforeEach(() => {
  vi.clearAllMocks();
  window.history.pushState({}, '', '/');
});

describe('App', () => {
  it('shows a loading state then the dashboard', async () => {
    fetchReport.mockResolvedValue(mixedReport);
    render(<App />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
    expect(await screen.findByTestId('status-banner')).toBeInTheDocument();
  });

  it('navigates to a spec detail and back, updating the URL', async () => {
    fetchReport.mockResolvedValue(mixedReport);
    render(<App />);
    await screen.findByTestId('status-banner');

    await userEvent.click(screen.getByTestId('spec-row-auth-login'));
    expect(await screen.findByTestId('spec-detail')).toBeInTheDocument();
    expect(window.location.pathname).toBe('/spec/auth-login');

    await userEvent.click(screen.getByTestId('back'));
    expect(await screen.findByTestId('status-banner')).toBeInTheDocument();
    expect(window.location.pathname).toBe('/');
  });

  it('deep-links straight to a spec when the path is /spec/<id>', async () => {
    fetchReport.mockResolvedValue(mixedReport);
    window.history.pushState({}, '', '/spec/tasks-create');
    render(<App />);
    expect(await screen.findByTestId('spec-detail')).toBeInTheDocument();
    expect(screen.getByTestId('detail-title')).toHaveTextContent('create a task');
  });

  it('shows an error state and can retry', async () => {
    fetchReport.mockRejectedValueOnce(new Error('connection refused'));
    render(<App />);
    expect(await screen.findByTestId('error')).toHaveTextContent('connection refused');

    fetchReport.mockResolvedValueOnce(mixedReport);
    await userEvent.click(screen.getByRole('button', { name: 'Retry' }));
    await waitFor(() => expect(screen.getByTestId('status-banner')).toBeInTheDocument());
  });
});
