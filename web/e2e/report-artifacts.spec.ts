import { test, expect } from '@playwright/test';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

// Runs against a report generated with a screenshot attachment (see serve.sh):
// specguard copied the image into an assets dir and referenced it by relative
// URL. This proves the per-spec gallery renders a real image end to end.
const reportURL = pathToFileURL(
  path.join(process.cwd(), '.e2e-report-artifacts.html'),
).href;

test('a spec with a screenshot shows a gallery that loads the image', async ({ page }) => {
  await page.goto(reportURL);
  await page.getByTestId('spec-row-auth-login').click();
  await expect(page.getByTestId('spec-detail')).toBeVisible();

  const thumb = page.getByTestId('gallery-thumb-0');
  await expect(thumb).toBeVisible();

  // The referenced file was actually copied next to the report and decodes.
  const img = thumb.locator('img');
  await expect(img).toHaveAttribute('src', /\.e2e-assets\/specs\/auth-login\/0\.png$/);
  await expect(async () => {
    const w = await img.evaluate((el) => (el as HTMLImageElement).naturalWidth);
    expect(w).toBeGreaterThan(0);
  }).toPass();
});

test('clicking a thumbnail opens a lightbox', async ({ page }) => {
  await page.goto(reportURL);
  await page.getByTestId('spec-row-auth-login').click();
  await page.getByTestId('gallery-thumb-0').click();
  await expect(page.getByTestId('lightbox')).toBeVisible();
  await page.getByTestId('lightbox').click();
  await expect(page.getByTestId('lightbox')).toHaveCount(0);
});
