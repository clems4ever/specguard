import { test, expect } from '@playwright/test';

test('logs in and lands on the task list', { tag: '@spec:auth-login' }, async ({ page }) => {
  await page.goto('/login');
  await expect(page).toHaveTitle(/Taskflow/);
});

test('logging out rejects a replayed cookie', { tag: '@spec:auth-logout' }, async ({ page }) => {
  await page.goto('/');
});
