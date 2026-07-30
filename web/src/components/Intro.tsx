import { useState } from 'react';
import type { SpecState } from '../selectors';
import { StatusBadge } from './StatusBadge';

const KEY = 'specguard:introDismissed';

// localStorage is unavailable on the opaque file:// origin and in some privacy
// modes; treat any failure as "not dismissed" so the card still renders.
function readDismissed(): boolean {
  try {
    return window.localStorage.getItem(KEY) === '1';
  } catch {
    return false;
  }
}

// The states worth explaining up front. Which set applies depends on whether a
// test run was overlaid (pass/fail) or we're showing coverage alone.
function legendStates(hasResults: boolean): SpecState[] {
  return hasResults
    ? ['passing', 'failing', 'not-run']
    : ['covered', 'uncovered', 'draft'];
}

// A short "what am I looking at?" for a first-time reader, plus a badge legend.
// It explains the product in one breath and can be dismissed once understood.
export function Intro({ hasResults = false }: { hasResults?: boolean }) {
  const [dismissed, setDismissed] = useState(readDismissed);
  if (dismissed) return null;

  const dismiss = () => {
    try {
      window.localStorage.setItem(KEY, '1');
    } catch {
      /* best effort — hide it for this view regardless */
    }
    setDismissed(true);
  };

  return (
    <section className="intro" data-testid="intro">
      <button className="intro-dismiss" onClick={dismiss} data-testid="intro-dismiss" title="Dismiss">
        ×
      </button>
      <h2 className="intro-title">What you’re looking at</h2>
      <p className="intro-body">
        <strong>specguard</strong> pins every product behaviour to the test that proves it. Each row
        below is one behaviour — a <em>spec</em> — linked to the code that implements it and the
        test that verifies it. Click any spec to read what it does and see that proof: the source,
        the covering tests, and screenshots.
      </p>
      <div className="intro-legend" data-testid="intro-legend">
        {legendStates(hasResults).map((s) => (
          <StatusBadge key={s} state={s} />
        ))}
      </div>
    </section>
  );
}
