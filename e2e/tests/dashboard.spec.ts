import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Dashboard', () => {
  test('stats display', async ({ page }) => {
    await page.goto('/ui');

    await expect(page.locator('h1')).toHaveText('Dashboard');

    // 11 users total (admin + 10 sample)
    const statCards = page.locator('div[role="group"] article');
    await expect(statCards).toHaveCount(4);

    // Total Users should be > 0
    const totalUsers = await statCards.nth(0).locator('h2').textContent();
    expect(parseInt(totalUsers!)).toBeGreaterThanOrEqual(11);

    // Total Groups should be 5
    const totalGroups = await statCards.nth(3).locator('h2').textContent();
    expect(parseInt(totalGroups!)).toBeGreaterThanOrEqual(5);
  });

  test('stat links navigate correctly', async ({ page }) => {
    await page.goto('/ui');

    // Click "Total Users" link
    await page.click('a[href="/ui/users"]');
    await expect(page).toHaveURL(/\/ui\/users/);
    await expect(page.locator('h1')).toHaveText('Users');

    // Go back to dashboard
    await page.click('a[href="/ui"]');
    await expect(page).toHaveURL(/\/ui/);

    // Click "Total Groups" link
    await page.click('a[href="/ui/groups"]');
    await expect(page).toHaveURL(/\/ui\/groups/);
    await expect(page.locator('h1')).toHaveText('Groups');
  });
});
