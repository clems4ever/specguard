import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { createServer, type Server } from 'node:http';
import { mkdtempSync, copyFileSync, readFileSync, existsSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';

// Reproduces GitHub *project* Pages: the report is hosted under a repo subpath
// (`/specguard/`), not at the origin root, and unknown paths fall back to
// 404.html (how Pages serves client-side routes on a hard refresh). A relative
// screenshot URL on a deep route like `/specguard/spec/<id>` must still resolve
// to `/specguard/assets/…` — the bug this guards against resolved it to
// `/specguard/spec/assets/…` and every screenshot 404'd. See src/base.ts.
//
// The screenshot fixture lives in e2e/fixtures/ (not a *.spec.ts), so its
// example-project tag isn't scanned as one of specguard's own spec references.

const BASE = '/specguard'; // stand-in for a project-Pages subpath
let server: Server;
let origin: string;

test.beforeAll(async () => {
  const web = process.cwd(); // playwright runs from web/
  const bin = path.join(web, '.e2e-specguard'); // built by e2e/serve.sh
  const site = mkdtempSync(path.join(tmpdir(), 'specguard-pages-'));
  const specguardDir = path.join(site, 'specguard');

  // Generate the real report through `specguard report`, laid out as a Pages
  // site would be: index.html + a sibling assets/ dir, referenced relatively.
  // The screenshot fixture's `path` is relative to web/, so run from there.
  execFileSync(
    bin,
    [
      'report',
      '-C', path.join(web, '..', 'example'),
      '-web', path.join(web, 'dist-single', 'index.html'),
      '-branch', 'e2e', '-commit', '0123456789abcdef', '-repo', 'clems4ever/specguard',
      '-results', path.join('e2e', 'fixtures', 'subpath-shots.json'),
      '-assets', path.join(specguardDir, 'assets'), '-assets-base', 'assets',
      '-o', path.join(specguardDir, 'index.html'),
    ],
    { cwd: web },
  );
  copyFileSync(path.join(specguardDir, 'index.html'), path.join(specguardDir, '404.html'));

  // A minimal static host with GitHub-Pages semantics: serve the file if it
  // exists, otherwise fall back to the subpath's 404.html.
  server = createServer((req, res) => {
    const url = decodeURIComponent((req.url ?? '/').split('?')[0]);
    let file = path.join(site, url);
    if (url.endsWith('/')) file = path.join(file, 'index.html');
    if (!existsSync(file) || !path.resolve(file).startsWith(path.resolve(site))) {
      file = path.join(specguardDir, '404.html'); // Pages' per-repo 404
    }
    res.setHeader('Content-Type', path.extname(file) === '.png' ? 'image/png' : 'text/html');
    res.end(readFileSync(file));
  });
  await new Promise<void>((r) => server.listen(0, '127.0.0.1', r));
  const addr = server.address();
  origin = `http://127.0.0.1:${typeof addr === 'object' && addr ? addr.port : 0}`;
});

test.afterAll(async () => {
  await new Promise<void>((r) => server.close(() => r()));
});

test('a screenshot loads on a deep spec route under a project subpath', {
  tag: '@spec:ui-subpath-assets',
}, async ({ page }) => {
  // Hard-refresh straight to the deep route (served via 404.html, like Pages).
  await page.goto(`${origin}${BASE}/spec/auth-login`);
  await expect(page.getByTestId('spec-detail')).toBeVisible();

  const img = page.getByTestId('gallery-thumb-0').locator('img');
  // Resolved against the base, not the current route.
  await expect(img).toHaveAttribute('src', `${BASE}/assets/specs/auth-login/0.png`);
  // And it actually decodes — the regression was a 404'd, broken image.
  await expect(async () => {
    const w = await img.evaluate((el) => (el as HTMLImageElement).naturalWidth);
    expect(w).toBeGreaterThan(0);
  }).toPass();
});
