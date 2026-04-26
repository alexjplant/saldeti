import { test, expect } from '@playwright/test';

test.describe('Seed Data Relationships', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/ui');
  });

  test('Engineering Team has correct members', async ({ page }) => {
    await page.goto('/ui/groups');
    // Click on Engineering Team
    await page.getByRole('link', { name: 'Engineering Team' }).click();
    await expect(page).toHaveURL(/\/ui\/groups\//);

    // Verify members
    const membersArticle = page.locator('#members');
    await expect(membersArticle.locator('td', { hasText: 'Alice Smith' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Bob Jones' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Eve Wilson' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Grace Lee' })).toBeVisible();
  });

  test('Marketing Team has correct members', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.getByRole('link', { name: 'Marketing Team' }).click();
    await expect(page).toHaveURL(/\/ui\/groups\//);

    const membersArticle = page.locator('#members');
    await expect(membersArticle.locator('td', { hasText: 'Charlie Brown' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Julia Roberts' })).toBeVisible();
  });

  test('All Staff contains nested groups', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.getByRole('link', { name: 'All Staff' }).click();
    await expect(page).toHaveURL(/\/ui\/groups\//);

    const membersArticle = page.locator('#members');
    // Should contain the nested groups
    await expect(membersArticle.getByText('Engineering Team')).toBeVisible();
    await expect(membersArticle.getByText('Marketing Team')).toBeVisible();
    // Should also contain individual users
    await expect(membersArticle.locator('td', { hasText: 'Henry Taylor' })).toBeVisible();
  });

  test('Leadership has correct members', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.getByRole('link', { name: 'Leadership' }).click();
    await expect(page).toHaveURL(/\/ui\/groups\//);

    const membersArticle = page.locator('#members');
    await expect(membersArticle.locator('td', { hasText: 'Diana Prince' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Frank Miller' })).toBeVisible();
  });

  test('Project Alpha has correct members', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.getByRole('link', { name: 'Project Alpha' }).click();
    await expect(page).toHaveURL(/\/ui\/groups\//);

    const membersArticle = page.locator('#members');
    await expect(membersArticle.locator('td', { hasText: 'Alice Smith' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Charlie Brown' })).toBeVisible();
    await expect(membersArticle.locator('td', { hasText: 'Eve Wilson' })).toBeVisible();
  });

  test('Alice has Eve as manager', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await expect(managerArticle.locator('a', { hasText: 'Eve Wilson' })).toBeVisible();
  });

  test('Bob has Eve as manager', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Bob Jones' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await expect(managerArticle.locator('a', { hasText: 'Eve Wilson' })).toBeVisible();
  });

  test('Eve has Frank as manager', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Eve Wilson' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await expect(managerArticle.locator('a', { hasText: 'Frank Miller' })).toBeVisible();
  });

  test('Frank has Admin as manager', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Frank Miller' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await expect(managerArticle.locator('a', { hasText: 'Admin User' })).toBeVisible();
  });

  test('Diana has Admin as manager', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Diana Prince' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await expect(managerArticle.locator('a', { hasText: 'Admin User' })).toBeVisible();
  });

  test('Eve has direct reports Alice and Bob', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Eve Wilson' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const reportsArticle = page.locator('article').filter({ hasText: 'Direct Reports' });
    await expect(reportsArticle.getByText('Alice Smith')).toBeVisible();
    await expect(reportsArticle.getByText('Bob Jones')).toBeVisible();
  });

  test('Admin has direct reports Frank and Diana', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Admin User' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const reportsArticle = page.locator('article').filter({ hasText: 'Direct Reports' });
    await expect(reportsArticle.getByText('Frank Miller')).toBeVisible();
    await expect(reportsArticle.getByText('Diana Prince')).toBeVisible();
  });

  test('Alice is member of Engineering Team, All Staff, and Project Alpha', async ({ page }) => {
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    const groupsArticle = page.locator('article').filter({ hasText: 'Group Memberships' });
    await expect(groupsArticle.getByText('Engineering Team')).toBeVisible();
    await expect(groupsArticle.getByText('All Staff')).toBeVisible();
    await expect(groupsArticle.getByText('Project Alpha')).toBeVisible();
  });

  test('Frank has correct manager chain', async ({ page }) => {
    // Frank -> Admin (manager)
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Frank Miller' }).locator('a').first().click();

    // Click manager link to navigate to Admin
    const managerArticle = page.locator('article').filter({ hasText: 'Manager' });
    await managerArticle.locator('a', { hasText: 'Admin User' }).click();
    await expect(page).toHaveURL(/\/ui\/users\//);

    // Admin should not have a manager (top of chain)
    await expect(page.locator('#manager').filter({ hasText: 'No manager assigned.' })).toBeVisible();
  });
});
