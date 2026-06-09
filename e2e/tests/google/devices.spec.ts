import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Devices', () => {
  test('displays devices page', async ({ page }) => {
    await page.goto('/google-ui/devices');

    await expect(page.locator('h2')).toHaveText('Devices');
    await expect(page.locator('hgroup p')).toContainText('Managed devices');
  });

  test('displays ChromeOS devices section', async ({ page }) => {
    await page.goto('/google-ui/devices');

    await expect(page.locator('h3', { hasText: 'ChromeOS Devices' })).toBeVisible();
  });

  test('displays Mobile devices section', async ({ page }) => {
    await page.goto('/google-ui/devices');

    await expect(page.locator('h3', { hasText: 'Mobile Devices' })).toBeVisible();
  });
});
