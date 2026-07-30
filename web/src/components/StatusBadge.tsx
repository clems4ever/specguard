import type { SpecState } from '../selectors';
import { STATE_LABEL } from '../selectors';

const GLYPH: Record<SpecState, string> = {
  covered: '✓',
  warning: '!',
  draft: '◦',
  uncovered: '✗',
  passing: '✓',
  failing: '✗',
  skipped: '–',
  'not-run': '?',
};

export function StatusBadge({ state }: { state: SpecState }) {
  return (
    <span
      className={`badge badge-${state}`}
      data-testid={`badge-${state}`}
      title={STATE_LABEL[state]}
    >
      <span className="badge-glyph" aria-hidden>
        {GLYPH[state]}
      </span>
      {STATE_LABEL[state]}
    </span>
  );
}
