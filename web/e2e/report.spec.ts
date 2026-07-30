import { test, expect } from '@playwright/test';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// These run against the REAL static report: `specguard report` filled the
// freshly-built single-file UI with the example project's data (see serve.sh).
// It is loaded from file:// with no server, proving the "browsable offline"
// contract end to end.
const reportURL = pathToFileURL(
  path.join(process.cwd(), '.e2e-report.html'),
).href;

test('static report boots offline from the embedded data', async ({ page }) => {
  const consoleErrors: string[] = [];
  page.on('console', (m) => m.type() === 'error' && consoleErrors.push(m.text()));
  page.on('pageerror', (e) => consoleErrors.push(e.message));

  await page.goto(reportURL);

  const banner = page.getByTestId('status-banner');
  await expect(banner).toContainText('PASS');
  await expect(banner).toContainText('8 specs');
  await expect(page.getByText('100%')).toBeVisible();
  await expect(page.getByTestId('area-auth')).toBeVisible();

  expect(consoleErrors).toEqual([]);
});

test('static report shows the provenance stamp', async ({ page }) => {
  await page.goto(reportURL);
  const stamp = page.getByTestId('stamp');
  await expect(stamp).toContainText('e2e-branch');
  await expect(stamp).toContainText('0123456'); // short commit
});

test('static report search filters the catalog', async ({ page }) => {
  await page.goto(reportURL);
  await page.getByTestId('search').fill('login');
  await expect(page.getByTestId('spec-row-auth-login')).toBeVisible();
  await expect(page.getByTestId('spec-row-tasks-create')).toBeHidden();
});

test('static report withholds server-only controls (refresh, changed-only)', async ({
  page,
}) => {
  await page.goto(reportURL);
  await expect(page.getByTestId('refresh')).toHaveCount(0);
  await expect(page.getByTestId('toggle-changed')).toHaveCount(0);
});

test('spec detail deep-links to the source and covering tests on GitHub', async ({ page }) => {
  await page.goto(reportURL);
  await page.getByTestId('spec-row-auth-login').click();

  // Spec source links to the .md at the pinned commit (from serve.sh: repo
  // clems4ever/specguard, commit 0123456789abcdef).
  const source = page.getByTestId('spec-source-link');
  await expect(source).toHaveAttribute(
    'href',
    'https://github.com/clems4ever/specguard/blob/0123456789abcdef/specs/auth/login.md',
  );

  // Each covering test links to its exact line (path:line label, #Lline anchor).
  const firstTest = page.getByTestId('test-list').locator('a').first();
  await expect(firstTest).toHaveAttribute(
    'href',
    /https:\/\/github\.com\/clems4ever\/specguard\/blob\/0123456789abcdef\/.+#L\d+/,
  );
});
