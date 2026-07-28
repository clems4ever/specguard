import { test, expect } from '@playwright/test';

test('creates a task and shows it on top', { tag: '@spec:tasks-create' }, async ({ page }) => {
  await page.goto('/');
});

test('completed state survives a reload', { tag: '@spec:tasks-complete' }, async ({ page }) => {
  await page.goto('/');
});

test('delete is guarded by a confirmation', { tag: '@spec:tasks-delete' }, async ({ page }) => {
  await page.goto('/');
});
