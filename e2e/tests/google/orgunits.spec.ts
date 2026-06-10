import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Org Units', () => {
  test('list org units shows empty state', async ({ page }) => {
    await page.goto('/google-ui/orgunits');

    await expect(page.locator('h2')).toHaveText('Org Units');
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('thead th').first()).toBeVisible();
  });

  test('create org unit', async ({ page }) => {
    await page.goto('/google-ui/orgunits/new');

    await expect(page.locator('h2')).toHaveText('New Org Unit');

    await page.fill('input[name="name"]', 'E2E Test OU');
    await page.fill('input[name="orgUnitPath"]', '/e2e-test-ou');

    await page.click('button[type="submit"]');

    // Should be redirected to org units list page
    await expect(page).toHaveURL(/\/google-ui\/orgunits$/);

    // Verify success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Verify the org unit appears in the list
    await expect(page.locator('td', { hasText: 'E2E Test OU' })).toBeVisible();
  });

  test('list org units shows created org unit', async ({ page }) => {
    await page.goto('/google-ui/orgunits');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: 'E2E Test OU' })).toBeVisible();
  });
});
