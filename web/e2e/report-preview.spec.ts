import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// A spec with a `preview` path, in a report generated with a preview base
// (a per-PR deploy), deep-links straight to the live feature — so a PM can jump
// from the spec to the running behaviour to review it. See preview.yml.
const web = process.cwd();
const bin = path.join(web, '.e2e-specguard'); // built by e2e/serve.sh
const out = path.join(web, '.e2e-report-preview.html');
const PREVIEW_BASE = 'https://preview.example';

test.beforeAll(() => {
  execFileSync(
    bin,
    [
      'report',
      '-C', path.join(web, '..', 'example'),
      '-web', path.join(web, 'dist-single', 'index.html'),
      '-branch', 'e2e', '-commit', '0123456789abcdef', '-repo', 'clems4ever/specguard',
      '-preview-base', PREVIEW_BASE,
      '-o', out,
    ],
    { cwd: web },
  );
});

test('a spec deep-links to the live preview', { tag: '@spec:ui-live-preview' }, async ({ page }) => {
  await page.goto(pathToFileURL(out).href);
  await page.getByTestId('spec-row-auth-login').click();
  await expect(page.getByTestId('spec-detail')).toBeVisible();

  // The example's auth-login declares `preview: /login`; joined to the base it
  // becomes a direct link to the running feature.
  const live = page.getByTestId('try-live');
  await expect(live).toBeVisible();
  await expect(live).toHaveAttribute('href', `${PREVIEW_BASE}/login`);
});
