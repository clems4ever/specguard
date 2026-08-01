import { describe, it, expect } from 'vitest';
import { codeLink, refLabel, previewUrl } from './links';
import type { ReportMeta } from './types';

const meta: ReportMeta = { repo: 'clems4ever/specguard', commit: 'abc123' };

describe('codeLink', () => {
  it('builds a commit-pinned blob URL with a line anchor', () => {
    expect(codeLink(meta, 'server/auth_test.go', 42)).toBe(
      'https://github.com/clems4ever/specguard/blob/abc123/server/auth_test.go#L42',
    );
  });

  it('omits the anchor when there is no (or zero) line', () => {
    expect(codeLink(meta, 'specs/auth/login.md')).toBe(
      'https://github.com/clems4ever/specguard/blob/abc123/specs/auth/login.md',
    );
    expect(codeLink(meta, 'specs/auth/login.md', 0)).toBe(
      'https://github.com/clems4ever/specguard/blob/abc123/specs/auth/login.md',
    );
  });

  it('returns null when repo or commit is missing (live serve has no commit)', () => {
    expect(codeLink(null, 'x.go', 1)).toBeNull();
    expect(codeLink({ repo: 'a/b' }, 'x.go', 1)).toBeNull();
    expect(codeLink({ commit: 'abc' }, 'x.go', 1)).toBeNull();
  });
});

describe('refLabel', () => {
  it('formats path:line, or path alone without a line', () => {
    expect(refLabel('a/b_test.go', 9)).toBe('a/b_test.go:9');
    expect(refLabel('a/b.md')).toBe('a/b.md');
    expect(refLabel('a/b.md', 0)).toBe('a/b.md');
  });
});

describe('previewUrl', () => {
  it('joins a preview base and a spec path, normalising slashes', () => {
    expect(previewUrl('https://pr-9.example', '/login')).toBe('https://pr-9.example/login');
    expect(previewUrl('https://pr-9.example/', 'login')).toBe('https://pr-9.example/login');
  });
  it('is null unless both are present', () => {
    expect(previewUrl(undefined, '/login')).toBeNull();
    expect(previewUrl('https://x', undefined)).toBeNull();
  });
});
