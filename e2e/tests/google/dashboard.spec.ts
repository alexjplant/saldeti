import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Dashboard', () => {
  test('displays dashboard heading', async ({ page }) => {
    await page.goto('/google-ui');
    await expect(page.locator('h2').first()).toHaveText('Dashboard');
    await expect(page.locator('hgroup p')).toContainText('Google Workspace');
  });

  test('displays stat cards', async ({ page }) => {
    await page.goto('/google-ui');

    // 6 cards in main grid + 1 mobile devices card = 7 total
    const statCards = page.locator('div.grid article');
    await expect(statCards).toHaveCount(7);

    // Verify each stat card header
    await expect(statCards.nth(0).locator('header')).toHaveText('Total Users');
    await expect(statCards.nth(1).locator('header')).toHaveText('Total Groups');
    await expect(statCards.nth(2).locator('header')).toHaveText('Org Units');
    await expect(statCards.nth(3).locator('header')).toHaveText('ChromeOS Devices');
    await expect(statCards.nth(4).locator('header')).toHaveText('Roles');
    await expect(statCards.nth(5).locator('header')).toHaveText('Domains');
    await expect(statCards.nth(6).locator('header')).toHaveText('Mobile Devices');
  });

  test('displays navigation items', async ({ page }) => {
    await page.goto('/google-ui');

    const nav = page.locator('header nav');
    await expect(nav.locator('a', { hasText: 'Dashboard' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Users' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Groups' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Org Units' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Devices' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Roles' })).toBeVisible();
    await expect(nav.locator('a', { hasText: 'Domains' })).toBeVisible();
  });

  test('stat card links navigate correctly', async ({ page }) => {
    await page.goto('/google-ui');

    // Click the stat card footer link for Users
    await page.locator('footer a[href="/google-ui/users"]').click();
    await expect(page).toHaveURL(/\/google-ui\/users/);
    await expect(page.locator('h2')).toHaveText('Users');

    // Go back to dashboard
    await page.goto('/google-ui');

    // Click the stat card footer link for Groups
    await page.locator('footer a[href="/google-ui/groups"]').click();
    await expect(page).toHaveURL(/\/google-ui\/groups/);
    await expect(page.locator('h2')).toHaveText('Groups');
  });
});
