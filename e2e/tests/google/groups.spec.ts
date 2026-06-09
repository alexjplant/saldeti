import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('Google Workspace Groups', () => {
  const groupEmail = `e2e-group-${Date.now().toString(36)}@example.com`;

  test('list groups shows empty state', async ({ page }) => {
    await page.goto('/google-ui/groups');

    await expect(page.locator('h2')).toHaveText('Groups');
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: 'No groups found.' })).toBeVisible();
  });

  test('create group', async ({ page }) => {
    await page.goto('/google-ui/groups/new');

    await expect(page.locator('h2')).toHaveText('New Group');

    await page.fill('input[name="email"]', groupEmail);
    await page.fill('input[name="name"]', 'E2E Test Group');
    await page.fill('textarea[name="description"]', 'Group created by E2E test');

    await page.click('button[type="submit"]');

    // Should be redirected to group detail page
    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toHaveText('E2E Test Group');

    // Verify group email shown
    await expect(page.locator('p')).toContainText(groupEmail);
  });

  test('list groups shows created group', async ({ page }) => {
    await page.goto('/google-ui/groups');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: groupEmail })).toBeVisible();
  });

  test('view group detail', async ({ page }) => {
    await page.goto('/google-ui/groups');

    // Click on the created group's email link
    await page.locator('tr', { hasText: groupEmail }).locator('a').first().click();

    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toHaveText('E2E Test Group');
    await expect(page.locator('p')).toContainText(groupEmail);

    // Verify group information article
    await expect(page.locator('article').filter({ hasText: 'Group Information' })).toBeVisible();

    // Verify breadcrumb
    await expect(page.locator('nav[aria-label="breadcrumb"] a', { hasText: 'Groups' })).toBeVisible();

    // Verify members section
    await expect(page.locator('h3', { hasText: 'Members' })).toBeVisible();
  });

  test('add member to group', async ({ page }) => {
    // First, create a user to add as a member
    await page.goto('/google-ui/users/new');
    const memberEmail = `member.${Date.now().toString(36)}@example.com`;
    await page.fill('input[name="primaryEmail"]', memberEmail);
    await page.fill('input[name="givenName"]', 'Member');
    await page.fill('input[name="familyName"]', 'TestUser');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/google-ui\/users\//);

    // Navigate to group detail
    await page.goto('/google-ui/groups');
    await page.locator('tr', { hasText: groupEmail }).locator('a').first().click();
    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);

    // Select a user from the dropdown (first non-empty option)
    const emailSelect = page.locator('select[name="email"]');
    await emailSelect.selectOption({ index: 1 });

    // Submit add member
    await page.locator('button', { hasText: 'Add Member' }).click();

    // Should remain on detail page
    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);

    // Verify success flash message
    await expect(page.locator('.flash-success')).toBeVisible();

    // Verify member appears in members table
    await expect(page.locator('h3 + figure table')).toBeVisible();
    await expect(page.locator('h3 + figure table td', { hasText: memberEmail })).toBeVisible();
  });

  test('remove member from group', async ({ page }) => {
    // Navigate to group detail
    await page.goto('/google-ui/groups');
    await page.locator('tr', { hasText: groupEmail }).locator('a').first().click();
    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);

    // Verify there is at least one member row
    const memberRows = page.locator('h3 + figure table tbody tr');
    const count = await memberRows.count();
    expect(count).toBeGreaterThan(0);

    // Accept confirm dialog and click remove button
    page.on('dialog', dialog => dialog.accept());
    await page.locator('button[aria-label="Remove"]').first().click();

    // Should redirect back to the group detail page
    await expect(page).toHaveURL(/\/google-ui\/groups\/[a-f0-9-]+$/);

    // Verify success flash or member removed
    await expect(page.locator('.flash-success')).toBeVisible();
  });
});
