import type { Finding } from '../types';

export function FindingsPanel({ findings }: { findings: Finding[] }) {
  if (findings.length === 0) {
    return (
      <div className="findings findings-empty" data-testid="findings">
        No findings — every spec is traced.
      </div>
    );
  }
  // Errors first, then warnings.
  const sorted = [...findings].sort((a, b) =>
    a.severity === b.severity ? 0 : a.severity === 'error' ? -1 : 1,
  );
  return (
    <ul className="findings" data-testid="findings">
      {sorted.map((f, i) => (
        <li key={i} className={`finding finding-${f.severity}`} data-testid={`finding-${f.rule}`}>
          <span className={`finding-sev finding-sev-${f.severity}`}>{f.severity}</span>
          <span className="finding-rule">[{f.rule}]</span>
          <span className="finding-msg">{f.message}</span>
          {f.file && <span className="finding-file">{f.file}</span>}
        </li>
      ))}
    </ul>
  );
}
