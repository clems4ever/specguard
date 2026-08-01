import { test, expect } from '@playwright/test';

// Record a video of each run and attach it — published as the spec's clip
// alongside its screenshots, so a reviewer can watch the behaviour, not just
// read it. Must be file-level (Playwright forbids test.use({video}) in describe).
test.use({ video: 'on' });

// These run against the real Go server serving the real built UI and the live
// report for the bundled example project (9 specs, one draft → PASS with one
// warning). No mocking — this is the full stack end to end.

// The tags below trace these tests to specguard's own UI specs, and the
// attached screenshots are published as those specs' galleries — so the
// self-report shows the very screens these tests exercise.
test('specs derive from one another', { tag: '@spec:spec-hierarchy' }, async ({ page }, testInfo) => {
  await page.goto('/');

  // The auth area is a tree: an "account access" capability over the concrete
  // login / logout / session specs.
  const parentRow = page.getByTestId('spec-row-auth-access');
  await expect(parentRow).toBeVisible();
  // The parent reads as covered though it has no direct test of its own — it is
  // verified by its children.
  await expect(parentRow.getByTestId('badge-covered')).toBeVisible();

  // Children are nested under it and collapse away with the parent's toggle.
  const child = page.getByTestId('spec-row-auth-login');
  await expect(child).toBeVisible();
  await page.getByTestId('tree-toggle-auth-access').click();
  await expect(child).toBeHidden();
  await page.getByTestId('tree-toggle-auth-access').click();

  await testInfo.attach('hierarchy', { body: await page.screenshot(), contentType: 'image/png' });

  // Drilling into a child shows a breadcrumb back up to the parent intent.
  await child.click();
  await expect(page.getByTestId('refines')).toContainText('Refines');
  await page.getByTestId('refines').click();
  await expect(page.getByTestId('detail-title')).toHaveText('Users can access their account');
  // …and the parent lists its refinements.
  await expect(page.getByTestId('child-link-auth-login')).toBeVisible();
});

test('the report orients a first-time reader', { tag: '@spec:ui-orientation' }, async ({ page }, testInfo) => {
  await page.goto('/');

  // A plain-language explanation of the product and a badge legend.
  const intro = page.getByTestId('intro');
  await expect(intro).toContainText('specguard');
  await expect(intro).toContainText('behaviour');
  await expect(page.getByTestId('intro-legend')).toBeVisible();

  // A table-of-contents to jump between areas.
  const toc = page.getByTestId('toc');
  await expect(toc.getByRole('link', { name: /auth/i })).toHaveAttribute('href', '#area-auth');

  // Areas summarise and collapse, so the catalog reads by section not row.
  const authRow = page.getByTestId('spec-row-auth-login');
  await expect(authRow).toBeVisible();
  await page.getByTestId('area-toggle-auth').click();
  await expect(authRow).toBeHidden();

  await testInfo.attach('orientation', { body: await page.screenshot(), contentType: 'image/png' });

  // The intro can be dismissed once understood.
  await page.getByTestId('intro-dismiss').click();
  await expect(intro).toBeHidden();
});

test('each area can describe itself', { tag: '@spec:ui-area-overview' }, async ({ page }) => {
  await page.goto('/');
  // The example's auth area ships a specs/auth/_area.md overview.
  const desc = page.getByTestId('area-desc-auth');
  await expect(desc).toBeVisible();
  await expect(desc).toContainText('Signing in and out');
});

test('dashboard shows the example project report', { tag: '@spec:ui-dashboard' }, async ({ page }, testInfo) => {
  await page.goto('/');
  const banner = page.getByTestId('status-banner');
  await expect(banner).toContainText('PASS');
  await expect(banner).toContainText('0 errors');
  await expect(banner).toContainText('1 warning');
  await expect(banner).toContainText('9 specs');

  // 100% coverage (drafts excluded from the denominator).
  await expect(page.getByText('100%')).toBeVisible();

  // Areas are grouped.
  await expect(page.getByTestId('area-auth')).toBeVisible();
  await expect(page.getByTestId('area-tasks')).toBeVisible();
  await expect(page.getByTestId('area-sharing')).toBeVisible();

  // The draft spec renders a draft badge and appears in the findings panel.
  const draftRow = page.getByTestId('spec-row-sharing-permissions');
  await expect(draftRow.getByTestId('badge-draft')).toBeVisible();
  await expect(page.getByTestId('finding-uncovered-draft')).toBeVisible();

  await testInfo.attach('dashboard', { body: await page.screenshot(), contentType: 'image/png' });
});

// A captioned Given/When/Then walkthrough, published as this spec's gallery —
// the concrete example a reviewer reads instead of prose alone. `video: 'on'`
// (file-level, below the imports) makes Playwright also attach a recording.
test('search filters the spec list', { tag: '@spec:ui-search' }, async ({ page }, testInfo) => {
  await page.goto('/');
  await testInfo.attach('Given the full catalog of specs', {
    body: await page.screenshot(),
    contentType: 'image/png',
  });

  await page.getByTestId('search').fill('login');
  await expect(page.getByTestId('spec-row-auth-login')).toBeVisible();
  await expect(page.getByTestId('spec-row-tasks-create')).toBeHidden();
  await testInfo.attach('When filtering for “login”, only matching specs remain', {
    body: await page.screenshot(),
    contentType: 'image/png',
  });

  await page.getByTestId('search').fill('nothing-xyz');
  await expect(page.getByTestId('no-match')).toBeVisible();
  await testInfo.attach('Then a query that matches nothing shows an empty state', {
    body: await page.screenshot(),
    contentType: 'image/png',
  });
});

test('opens a spec detail with rendered body and covering tests', { tag: '@spec:ui-spec-detail' }, async ({ page }, testInfo) => {
  await page.goto('/');
  await page.getByTestId('spec-row-auth-login').click();

  await expect(page).toHaveURL(/\/spec\/auth-login$/);
  await expect(page.getByTestId('detail-title')).toHaveText(
    'A user can log in with email and password',
  );
  // Body markdown rendered as a real heading.
  await expect(page.getByTestId('markdown').locator('h2').first()).toBeVisible();
  // Covering tests listed.
  await expect(page.getByTestId('test-list')).toContainText('server/auth_test.go');
  await expect(page.getByTestId('test-list')).toContainText('web/e2e/auth.spec.ts');

  await testInfo.attach('spec detail', { body: await page.screenshot(), contentType: 'image/png' });

  await page.getByTestId('back').click();
  await expect(page.getByTestId('status-banner')).toBeVisible();
});

test('deep-links directly to a spec detail', async ({ page }) => {
  await page.goto('/spec/tasks-delete');
  await expect(page.getByTestId('spec-detail')).toBeVisible();
  await expect(page.getByTestId('detail-title')).toContainText('Deleting a task');
});
