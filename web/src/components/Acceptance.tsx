import type { SpecStatus } from '../types';

// Acceptance surfaces a spec's PM behavioural-review state — the human gate on
// top of the automated tests. A parent spec with no direct test isn't gated
// itself (its children are), so nothing is shown for it.
export function Acceptance({ spec }: { spec: SpecStatus }) {
  if (!spec.covered || !spec.lifecycle) return null;

  if (spec.lifecycle === 'accepted') {
    const who = spec.acceptedBy ? ` by ${spec.acceptedBy}` : '';
    const when = spec.acceptedAt ? ` on ${spec.acceptedAt.slice(0, 10)}` : '';
    return (
      <div className="accept accept-ok" data-testid="acceptance">
        <span className="accept-mark">✓</span> Behaviour accepted{who}
        {when}
        <code className="accept-fp">{spec.fingerprint}</code>
      </div>
    );
  }
  if (spec.lifecycle === 'stale') {
    return (
      <div className="accept accept-stale" data-testid="acceptance">
        <span className="accept-mark">⟳</span> Behaviour changed since it was accepted — awaiting
        PM re-review.
      </div>
    );
  }
  // implemented (built, never accepted)
  return (
    <div className="accept accept-pending" data-testid="acceptance">
      <span className="accept-mark">◷</span> Awaiting PM behavioural review.
    </div>
  );
}
