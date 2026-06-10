import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Roles', () => {
  test('displays roles page', async ({ page }) => {
    await page.goto('/google-ui/roles');

    await expect(page.locator('h2')).toHaveText('Roles');
    await expect(page.locator('hgroup p')).toContainText('Admin roles');
  });

  test('displays roles table', async ({ page }) => {
    await page.goto('/google-ui/roles');

    await expect(page.locator('table')).toBeVisible();

    // Verify table headers
    await expect(page.locator('th', { hasText: 'Role Name' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Role Description' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Is System Role' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Is Super Admin Role' })).toBeVisible();
  });
});
