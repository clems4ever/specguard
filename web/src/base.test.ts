import { describe, it, expect } from 'vitest';
import { basePath, assetUrl } from './base';

describe('basePath', () => {
  it('is the directory at a site root', () => {
    expect(basePath('/')).toBe('/');
    expect(basePath('/index.html')).toBe('/');
  });

  it('recovers the base under a project subpath', () => {
    expect(basePath('/specguard/')).toBe('/specguard/');
    expect(basePath('/specguard/index.html')).toBe('/specguard/');
  });

  it('strips a spec-detail route to recover the base', () => {
    expect(basePath('/spec/auth-login')).toBe('/');
    expect(basePath('/specguard/spec/ui-code-links')).toBe('/specguard/');
    expect(basePath('/specguard/spec/ui-code-links/')).toBe('/specguard/');
  });

  it('handles a local file path', () => {
    expect(basePath('/home/u/report.html')).toBe('/home/u/');
  });
});

describe('assetUrl', () => {
  it('resolves a relative asset against the base — the deep-route bug', () => {
    // On a spec page under a project subpath, a bare relative path would 404;
    // it must resolve to `/specguard/assets/…`, not `/specguard/spec/assets/…`.
    expect(assetUrl('assets/specs/x/0.png', '/specguard/spec/x')).toBe(
      '/specguard/assets/specs/x/0.png',
    );
    expect(assetUrl('assets/specs/x/0.png', '/spec/x')).toBe('/assets/specs/x/0.png');
  });

  it('leaves absolute paths and full URLs unchanged', () => {
    expect(assetUrl('/assets/x.png', '/specguard/spec/x')).toBe('/assets/x.png');
    expect(assetUrl('https://cdn/x.png', '/specguard/spec/x')).toBe('https://cdn/x.png');
    expect(assetUrl('data:image/png;base64,AAAA', '/spec/x')).toBe('data:image/png;base64,AAAA');
  });
});
