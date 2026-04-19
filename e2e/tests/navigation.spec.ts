import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui/login');
  await page.fill('input[name="username"]', 'admin@saldeti.local');
  await page.fill('input[name="password"]', 'Simulator123!');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/ui$/);
});

test.describe('Navigation', () => {
  test('user to groups via membership', async ({ page }) => {
    // Go to Alice's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();

    // Should show group memberships
    await expect(page.locator('article').filter({ hasText: 'Group Memberships' })).toBeVisible();

    // Click a group link
    const groupLink = page.locator('article').filter({ hasText: 'Group Memberships' }).locator('a').first();
    await groupLink.click();

    // Should land on group detail
    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);
  });

  test('group to users via members', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();

    // Should show members
    const membersArticle = page.locator('article').filter({ hasText: 'Members' });
    await expect(membersArticle).toBeVisible();

    // Click a member name
    const memberLink = membersArticle.locator('a').first();
    await memberLink.click();

    // Should land on user detail
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);
  });

  test('manager link navigation', async ({ page }) => {
    // Go to Alice Smith's page (manager is Eve Wilson)
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();

    // Manager section should be visible
    await expect(page.locator('article').filter({ hasText: 'Manager' })).toBeVisible();

    // Click manager link
    await page.locator('article').filter({ hasText: 'Manager' }).locator('a').click();

    // Should land on Eve Wilson's page
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);
    await expect(page.locator('h1')).toHaveText('Eve Wilson');
  });

  test('direct reports link', async ({ page }) => {
    // Go to Eve Wilson (has Alice and Bob as direct reports)
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Eve Wilson' }).locator('a').first().click();

    // Direct reports section
    await expect(page.locator('article').filter({ hasText: 'Direct Reports' })).toBeVisible();

    // Click a direct report
    const drLink = page.locator('article').filter({ hasText: 'Direct Reports' }).locator('a').first();
    await drLink.click();

    // Should land on user detail
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);
  });

  test('nav bar navigation', async ({ page }) => {
    // Use nav bar to navigate between sections
    await page.click('a[href="/ui/users"]');
    await expect(page).toHaveURL(/\/ui\/users/);

    await page.click('a[href="/ui/groups"]');
    await expect(page).toHaveURL(/\/ui\/groups/);

    await page.click('a[href="/ui"]');
    await expect(page).toHaveURL(/\/ui/);
  });
});
