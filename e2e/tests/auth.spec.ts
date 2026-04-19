import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('login and see dashboard', async ({ page }) => {
    // Navigate to login page
    await page.goto('/ui/login');
    
    // Verify login form is visible
    await expect(page.locator('h1')).toHaveText('Saldeti Admin');
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    
    // Fill in credentials
    await page.fill('input[name="username"]', 'admin@saldeti.local');
    await page.fill('input[name="password"]', 'Simulator123!');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Should be redirected to dashboard
    await expect(page).toHaveURL(/\/ui$/);
    await expect(page.locator('h1')).toHaveText('Dashboard');

    // Should show stat cards
    await expect(page.locator('div[role="group"] article')).toHaveCount(4);
  });

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/ui/login');

    await page.fill('input[name="username"]', 'admin@saldeti.local');
    await page.fill('input[name="password"]', 'wrong-password');
    await page.click('button[type="submit"]');

    // Should stay on login page with error
    await expect(page).toHaveURL(/\/ui\/login/);
    await expect(page.getByRole('alert')).toBeVisible();
  });

  test('logout redirects to login', async ({ page }) => {
    // Login first
    await page.goto('/ui/login');
    await page.fill('input[name="username"]', 'admin@saldeti.local');
    await page.fill('input[name="password"]', 'Simulator123!');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/ui$/);
    
    // Click logout
    await page.click('a[href="/ui/logout"]');
    
    // Should be redirected to login page
    await expect(page).toHaveURL(/\/ui\/login/);
    await expect(page.locator('h1')).toHaveText('Saldeti Admin');
  });

  test('dashboard shows correct stats', async ({ page }) => {
    // Login
    await page.goto('/ui/login');
    await page.fill('input[name="username"]', 'admin@saldeti.local');
    await page.fill('input[name="password"]', 'Simulator123!');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/ui$/);
    
    // Dashboard should show user and group counts
    // Seed data has 11 users (admin + 10 sample) and 5 groups
    const statCards = page.locator('div[role="group"] article');
    await expect(statCards).toHaveCount(4);

    // Total Users card should contain a number
    const totalUsersText = await statCards.nth(0).locator('h2').textContent();
    expect(parseInt(totalUsersText!)).toBeGreaterThan(0);

    // Total Groups card should contain a number
    const totalGroupsText = await statCards.nth(3).locator('h2').textContent();
    expect(parseInt(totalGroupsText!)).toBeGreaterThan(0);
  });
});
