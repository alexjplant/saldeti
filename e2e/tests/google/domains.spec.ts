import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Domains', () => {
  test('displays domains page', async ({ page }) => {
    await page.goto('/google-ui/domains');

    await expect(page.locator('h2')).toHaveText('Domains');
    await expect(page.locator('hgroup p')).toContainText('Verified domains');
  });

  test('displays domains table', async ({ page }) => {
    await page.goto('/google-ui/domains');

    await expect(page.locator('table')).toBeVisible();

    // Verify table headers
    await expect(page.locator('th', { hasText: 'Domain Name' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Is Primary' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Verified' })).toBeVisible();
    await expect(page.locator('th', { hasText: 'Creation Time' })).toBeVisible();
  });
});
