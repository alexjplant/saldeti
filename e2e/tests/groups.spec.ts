import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Groups', () => {
  test('list groups', async ({ page }) => {
    await page.goto('/ui/groups');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('td', { hasText: 'Engineering Team' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Marketing Team' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'All Staff' })).toBeVisible();
    await expect(page.locator('tr', { hasText: 'Leadership' }).locator('td').first()).toBeVisible();
    await expect(page.locator('tr', { hasText: 'Project Alpha' }).locator('td').first()).toBeVisible();
  });

  test('search groups', async ({ page }) => {
    await page.goto('/ui/groups');

    await page.fill('input[name="search"]', 'Engineering');
    await page.press('input[name="search"]', 'Enter');

    await expect(page.locator('td', { hasText: 'Engineering Team' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Marketing Team' })).not.toBeVisible();
  });

  test('create group', async ({ page }) => {
    await page.goto('/ui/groups/new');

    await expect(page.locator('h2')).toHaveText('New Group');

    await page.fill('input[name="displayName"]', 'E2E Test Group');
    await page.fill('textarea[name="description"]', 'Group created by E2E test');
    await page.fill('input[name="mailNickname"]', 'e2etestgroup');
    await page.check('input[name="securityEnabled"]');
    await page.selectOption('select[name="visibility"]', 'Public');

    await page.click('button[type="submit"]');

    // Should be redirected to detail
    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toHaveText('E2E Test Group');
    await expect(page.locator('dd', { hasText: 'Group created by E2E test' })).toBeVisible();
  });

  test('create group validation', async ({ page }) => {
    await page.goto('/ui/groups/new');

    // Wait for page to be fully loaded
    await expect(page.locator('h2')).toHaveText('New Group');

    // Remove required attributes to bypass HTML5 validation
    await page.evaluate(() => {
      document.querySelectorAll('[required]').forEach(el => el.removeAttribute('required'));
    });

    // Submit empty form
    await page.click('button[type="submit"]');

    // Wait for form to re-render with error
    await page.waitForLoadState('networkidle');

    await expect(page.locator('.flash-danger')).toBeVisible();
  });

  test('view group detail', async ({ page }) => {
    await page.goto('/ui/groups');

    // Click on Engineering Team
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();

    await expect(page.locator('h2')).toHaveText('Engineering Team');

    // Should show members section (Alice, Bob, Eve, Grace are members)
    await expect(page.locator('#members')).toBeVisible();
    await expect(page.locator('#members').locator('td', { hasText: 'Alice Smith' })).toBeVisible();
  });

  test('edit group', async ({ page }) => {
    await page.goto('/ui/groups');

    await page.locator('tr', { hasText: 'Marketing Team' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);

    // Click Edit - wait for page to load and use href selector
    await page.waitForLoadState('networkidle');
    await page.locator('a[href*="/edit"]').click();
    await expect(page.locator('h2')).toHaveText('Edit Group');

    await page.fill('textarea[name="description"]', 'Updated by E2E test');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);
    await expect(page.locator('dd').filter({ hasText: 'Updated by E2E test' }).first()).toBeVisible();
  });

  test.describe('delete group', () => {
    let groupId: string;

    test.beforeEach(async ({ page }) => {
      // Create a throwaway group for the delete test
      await page.goto('/ui/groups/new');
      await page.fill('input[name="displayName"]', 'Delete Test Group');
      await page.fill('textarea[name="description"]', 'Group to be deleted');
      await page.fill('input[name="mailNickname"]', 'deletetestgroup');
      await page.click('button[type="submit"]');
      // Capture the groupId from the redirected URL
      await expect(page).toHaveURL(/\/ui\/groups\/([a-f0-9-]+)$/);
      const url = page.url();
      groupId = url.match(/\/ui\/groups\/([a-f0-9-]+)$/)?.[1] || '';
    });

    test('can delete group', async ({ page }) => {
      page.on('dialog', dialog => dialog.accept());
      await page.locator('button', { hasText: /Delete/i }).click();

      // Should be redirected to group list
      await expect(page).toHaveURL(/\/ui\/groups$/);

      // The throwaway group should be gone
      await expect(page.locator('tr', { hasText: 'Delete Test Group' })).not.toBeVisible();
    });

    test.afterEach(async ({ page }) => {
      // Cleanup: if the test didn't delete, try to clean up
      if (groupId) {
        await page.goto('/ui/groups');
        const row = page.locator('tr', { hasText: 'Delete Test Group' });
        if (await row.isVisible()) {
          await row.locator('a').first().click();
          page.on('dialog', dialog => dialog.accept());
          await page.locator('button', { hasText: /Delete/i }).click();
          await expect(page).toHaveURL(/\/ui\/groups$/);
        }
      }
    });
  });

  test('shows transitive members', async ({ page }) => {
    await page.goto('/ui/groups');

    // Click on Engineering Team
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);

    // Verify transitive members article
    const transitiveArticle = page.locator('#group-transitivemembers');
    await expect(transitiveArticle).toBeVisible();

    // Verify header contains "Transitive Members"
    await expect(transitiveArticle.locator('header')).toContainText('Transitive Members');

    // Verify either a table is present or empty message is shown
    const hasTable = await transitiveArticle.locator('table').count();
    const hasEmptyText = await transitiveArticle.locator('p', { hasText: 'No transitive members.' }).count();
    expect(hasTable + hasEmptyText).toBeGreaterThan(0);
  });

  test('shows app role assignments', async ({ page }) => {
    await page.goto('/ui/groups');

    // Click on Engineering Team
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/groups\/[a-f0-9-]+$/);

    // Verify app role assignments article
    const approleArticle = page.locator('#group-approles');
    await expect(approleArticle).toBeVisible();

    // Verify header contains "App Role Assignments"
    await expect(approleArticle.locator('header')).toContainText('App Role Assignments');

    // Verify either a table is present or empty message is shown
    const hasTable = await approleArticle.locator('table').count();
    const hasEmptyText = await approleArticle.locator('p', { hasText: 'No app role assignments.' }).count();
    expect(hasTable + hasEmptyText).toBeGreaterThan(0);
  });
});
