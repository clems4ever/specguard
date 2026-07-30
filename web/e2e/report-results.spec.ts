import { test, expect } from '@playwright/test';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// Runs against a report generated with a real `go test -json` (all pass) plus a
// Playwright fixture that fails `auth-logout` (see serve.sh + e2e/fixtures). It
// proves the covered-but-failing state end to end: traceability passes, yet the
// report reads FAIL because a covering test failed.
const reportURL = pathToFileURL(
  path.join(process.cwd(), '.e2e-report-results.html'),
).href;

test('a failing covering test flips the verdict and marks the spec red', async ({ page }) => {
  await page.goto(reportURL);

  const banner = page.getByTestId('status-banner');
  await expect(banner).toContainText('FAIL');
  await expect(banner).toContainText('1 failing');

  // Result-aware badges: the failing spec is red, its passing siblings green.
  await expect(
    page.getByTestId('spec-row-auth-logout').getByTestId('badge-failing'),
  ).toBeVisible();
  await expect(
    page.getByTestId('spec-row-auth-login').getByTestId('badge-passing'),
  ).toBeVisible();
});

test('spec detail shows which covering test failed', async ({ page }) => {
  await page.goto(reportURL);
  await page.getByTestId('spec-row-auth-logout').click();
  await expect(page.getByTestId('spec-detail')).toBeVisible();
  await expect(page.getByTestId('badge-failing')).toBeVisible();

  const tests = page.getByTestId('test-list');
  // The unit test passed, the e2e failed — both shown per covering test.
  await expect(tests.getByTestId('ref-status-passed')).toBeVisible();
  await expect(tests.getByTestId('ref-status-failed')).toBeVisible();
});
