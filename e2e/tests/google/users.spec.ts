import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Users', () => {
  const testEmail = `e2e.test.${Date.now().toString(36)}@example.com`;

  test('list users shows empty state', async ({ page }) => {
    await page.goto('/google-ui/users');

    await expect(page.locator('h2')).toHaveText('Users');
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: 'No users found.' })).toBeVisible();
  });

  test('create user', async ({ page }) => {
    await page.goto('/google-ui/users/new');

    await expect(page.locator('h2')).toHaveText('New User');

    await page.fill('input[name="primaryEmail"]', testEmail);
    await page.fill('input[name="givenName"]', 'E2E');
    await page.fill('input[name="familyName"]', 'TestUser');

    await page.click('button[type="submit"]');

    // Should be redirected to detail page
    await expect(page).toHaveURL(/\/google-ui\/users\/[a-f0-9-]+$/);

    // Verify the user email is shown on the detail page
    await expect(page.locator('p')).toContainText(testEmail);
  });

  test('list users shows created user', async ({ page }) => {
    await page.goto('/google-ui/users');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: testEmail })).toBeVisible();
  });

  test('view user detail', async ({ page }) => {
    await page.goto('/google-ui/users');

    // Click on the created user's email link
    await page.locator('tr', { hasText: testEmail }).locator('a').first().click();

    await expect(page).toHaveURL(/\/google-ui\/users\/[a-f0-9-]+$/);

    // Verify detail page shows Primary Email
    await expect(page.locator('p')).toContainText(testEmail);

    // Verify detail sections are visible
    await expect(page.locator('article').filter({ hasText: 'General' })).toBeVisible();
    await expect(page.locator('article').filter({ hasText: 'Status' })).toBeVisible();

    // Verify breadcrumb
    await expect(page.locator('nav[aria-label="breadcrumb"] a', { hasText: 'Users' })).toBeVisible();
  });

  test('edit user', async ({ page }) => {
    await page.goto('/google-ui/users');

    // Click on the test user's email link
    await page.locator('tr', { hasText: testEmail }).locator('a').first().click();
    await expect(page).toHaveURL(/\/google-ui\/users\/[a-f0-9-]+$/);

    // Click Edit link
    await page.locator('a[href*="/edit"]').click();
    await expect(page.locator('h2')).toHaveText('Edit User');

    // Modify givenName
    await page.fill('input[name="givenName"]', 'E2EUpdated');

    await page.click('button[type="submit"]');

    // Should be redirected back to detail page
    await expect(page).toHaveURL(/\/google-ui\/users\/[a-f0-9-]+$/);
    await expect(page.locator('dd').filter({ hasText: 'E2EUpdated' }).first()).toBeVisible();
  });

  test('delete user', async ({ page }) => {
    // Navigate to user detail
    await page.goto('/google-ui/users');
    await page.locator('tr', { hasText: testEmail }).locator('a').first().click();
    await expect(page).toHaveURL(/\/google-ui\/users\/[a-f0-9-]+$/);

    // Accept the confirm dialog
    page.on('dialog', dialog => dialog.accept());

    // Click Delete button
    await page.locator('button[form="delete-user"]').click();

    // Should be redirected to user list
    await expect(page).toHaveURL(/\/google-ui\/users$/);

    // The deleted user should be gone
    await expect(page.locator('td', { hasText: testEmail })).not.toBeVisible();
  });
});
