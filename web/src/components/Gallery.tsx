import { useEffect, useState } from 'react';
import type { Artifact } from '../types';

// Gallery shows the screenshots a test captured for a spec — the visual proof a
// PM can look at. Thumbnails open a lightbox; Escape or a click closes it. It
// renders nothing when there are no artifacts, so specs without screenshots are
// unaffected.
export function Gallery({ artifacts }: { artifacts?: Artifact[] | null }) {
  const shots = artifacts ?? [];
  const [open, setOpen] = useState<number | null>(null);

  useEffect(() => {
    if (open === null) return;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(null);
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open]);

  if (shots.length === 0) return null;

  return (
    <div>
      <h3>Screenshots ({shots.length})</h3>
      <div className="gallery" data-testid="gallery">
        {shots.map((a, i) => (
          <button
            key={a.path}
            className="gallery-thumb"
            data-testid={`gallery-thumb-${i}`}
            onClick={() => setOpen(i)}
            title={a.name || `screenshot ${i + 1}`}
          >
            <img src={a.path} alt={a.name || `screenshot ${i + 1}`} loading="lazy" />
          </button>
        ))}
      </div>

      {open !== null && shots[open] && (
        <div
          className="lightbox"
          data-testid="lightbox"
          role="dialog"
          aria-modal="true"
          onClick={() => setOpen(null)}
        >
          <img src={shots[open].path} alt={shots[open].name || 'screenshot'} />
          <div className="lightbox-caption">{shots[open].name}</div>
        </div>
      )}
    </div>
  );
}
