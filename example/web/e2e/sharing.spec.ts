import { test, expect } from '@playwright/test';

test('owner invites a collaborator by email', { tag: '@spec:sharing-invite' }, async ({ page }) => {
  await page.goto('/');
});
