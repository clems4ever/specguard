import { useEffect, useState } from 'react';
import type { Artifact } from '../types';
import { assetUrl } from '../base';

const isVideo = (path: string) => /\.(webm|mp4|mov|m4v)$/i.test(path);

// Gallery shows what a test captured for a spec — screenshots and video clips,
// each with its caption — laid out as an ordered walkthrough. Captions are the
// concrete steps ("Given …", "When …", "Then …"), so a reader sees the exact
// behaviour, not just prose. Images open in a lightbox; videos play inline. It
// renders nothing when there are no artifacts.
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

  // With captions the gallery reads as a step-by-step walkthrough; without, it's
  // just captured screens.
  const captioned = shots.some((s) => s.name);

  return (
    <div>
      <h3>{captioned ? 'Walkthrough' : 'Screenshots'} ({shots.length})</h3>
      <ol className="gallery" data-testid="gallery">
        {shots.map((a, i) => (
          <li key={a.path} className="gallery-step" data-testid={`gallery-step-${i}`}>
            {isVideo(a.path) ? (
              <video
                className="gallery-video"
                data-testid={`gallery-video-${i}`}
                src={assetUrl(a.path)}
                controls
                preload="metadata"
              />
            ) : (
              <button
                className="gallery-thumb"
                data-testid={`gallery-thumb-${i}`}
                onClick={() => setOpen(i)}
                title={a.name || `screenshot ${i + 1}`}
              >
                <img src={assetUrl(a.path)} alt={a.name || `screenshot ${i + 1}`} loading="lazy" />
              </button>
            )}
            {a.name && (
              <div className="gallery-caption" data-testid={`gallery-caption-${i}`}>
                {a.name}
              </div>
            )}
          </li>
        ))}
      </ol>

      {open !== null && shots[open] && (
        <div
          className="lightbox"
          data-testid="lightbox"
          role="dialog"
          aria-modal="true"
          onClick={() => setOpen(null)}
        >
          <img src={assetUrl(shots[open].path)} alt={shots[open].name || 'screenshot'} />
          <div className="lightbox-caption">{shots[open].name}</div>
        </div>
      )}
    </div>
  );
}
