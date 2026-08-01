import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Gallery } from './Gallery';

const shots = [
  { name: 'login screen', path: 'assets/specs/auth-login/0.png' },
  { name: 'error state', path: 'assets/specs/auth-login/1.png' },
];

describe('Gallery', () => {
  it('renders nothing when there are no artifacts', () => {
    const { container } = render(<Gallery artifacts={[]} />);
    expect(container).toBeEmptyDOMElement();
    const { container: c2 } = render(<Gallery artifacts={null} />);
    expect(c2).toBeEmptyDOMElement();
  });

  it('renders a thumbnail per screenshot', () => {
    render(<Gallery artifacts={shots} />);
    const gallery = screen.getByTestId('gallery');
    expect(gallery.querySelectorAll('img')).toHaveLength(2);
    const first = screen.getByTestId('gallery-thumb-0').querySelector('img')!;
    // Resolved against the app base (`/` under jsdom) so it loads on any route.
    expect(first).toHaveAttribute('src', '/assets/specs/auth-login/0.png');
  });

  it('opens a lightbox on click and closes it on click', async () => {
    render(<Gallery artifacts={shots} />);
    expect(screen.queryByTestId('lightbox')).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('gallery-thumb-1'));
    const lightbox = screen.getByTestId('lightbox');
    expect(lightbox.querySelector('img')).toHaveAttribute('src', '/assets/specs/auth-login/1.png');

    await userEvent.click(lightbox);
    expect(screen.queryByTestId('lightbox')).not.toBeInTheDocument();
  });

  it('shows each artifact caption as a walkthrough step', () => {
    render(<Gallery artifacts={shots} />);
    expect(screen.getByText('Walkthrough (2)')).toBeInTheDocument();
    expect(screen.getByTestId('gallery-caption-0')).toHaveTextContent('login screen');
    expect(screen.getByTestId('gallery-caption-1')).toHaveTextContent('error state');
  });

  it('renders a video artifact as an inline player, not a thumbnail', () => {
    render(
      <Gallery
        artifacts={[{ name: 'the whole flow', path: 'assets/specs/x/0.webm' }]}
      />,
    );
    const vid = screen.getByTestId('gallery-video-0');
    expect(vid.tagName).toBe('VIDEO');
    expect(vid).toHaveAttribute('src', '/assets/specs/x/0.webm');
    expect(screen.queryByTestId('gallery-thumb-0')).not.toBeInTheDocument();
  });
});
