import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// A report generated from a run that attached a named screenshot and a video
// renders them as a captioned walkthrough: the attachment name becomes the step
// caption, and the video attachment becomes an inline player. The results
// fixture lives in e2e/fixtures/ so its @spec token isn't scanned as a real ref.
const web = process.cwd();
const bin = path.join(web, '.e2e-specguard'); // built by e2e/serve.sh
const out = path.join(web, '.e2e-report-walkthrough.html');

test.beforeAll(() => {
  execFileSync(
    bin,
    [
      'report',
      '-C', path.join(web, '..', 'example'),
      '-web', path.join(web, 'dist-single', 'index.html'),
      '-branch', 'e2e', '-commit', '0123456789abcdef', '-repo', 'clems4ever/specguard',
      '-results', path.join('e2e', 'fixtures', 'walkthrough-results.json'),
      '-assets', path.join(web, '.e2e-walkthrough-assets'), '-assets-base', '.e2e-walkthrough-assets',
      '-o', out,
    ],
    { cwd: web },
  );
});

test('captures render as a captioned walkthrough with video', { tag: '@spec:ui-walkthrough' }, async ({ page }) => {
  await page.goto(pathToFileURL(out).href);
  await page.getByTestId('spec-row-auth-login').click();
  await expect(page.getByTestId('spec-detail')).toBeVisible();

  // The attachment name is shown as the step caption.
  await expect(page.getByTestId('gallery-caption-0')).toContainText('Given the login screen');
  // The video attachment (second capture) renders as an inline <video>, not a thumbnail.
  await expect(page.getByTestId('gallery-video-1')).toHaveCount(1);
  await expect(page.getByTestId('gallery-video-1')).toHaveAttribute('src', /\.webm$/);
});
