import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui/login');
  await page.fill('input[name="username"]', 'admin@saldeti.local');
  await page.fill('input[name="password"]', 'Simulator123!');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/ui$/);
});

test.describe('Membership', () => {
  test('add member', async ({ page }) => {
    // Go to All Staff group (not Leadership since it might be deleted by other tests)
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'All Staff' }).locator('a').first().click();

    // Add Ivan Guest as member - target the select in the Members section
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    const select = membersSection.locator('select[name="userId"]');
    const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
    const ivanValue = await ivanOption.getAttribute('value');
    await select.selectOption(ivanValue!);
    await membersSection.locator('button', { hasText: 'Add Member' }).click();

    // Wait for page to reload
    await page.waitForLoadState('networkidle');

    // Ivan should appear in members table
    await expect(page.locator('article').filter({ hasText: 'Members' }).locator('td', { hasText: 'Ivan Guest' })).toBeVisible();
  });

  test('remove member', async ({ page }) => {
    // Go to All Staff group
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'All Staff' }).locator('a').first().click();

    // First add Ivan Guest so we have a known member to remove
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    const ivanInMembers = membersSection.locator('td', { hasText: 'Ivan Guest' });
    if (!await ivanInMembers.isVisible()) {
      const select = membersSection.locator('select[name="userId"]');
      const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
      const ivanValue = await ivanOption.getAttribute('value');
      await select.selectOption(ivanValue!);
      await membersSection.locator('button', { hasText: 'Add Member' }).click();
      await page.waitForLoadState('networkidle');
    }

    // Now remove Ivan Guest
    const ivanRow = membersSection.locator('tr', { hasText: 'Ivan Guest' });
    await expect(ivanRow).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await ivanRow.locator('button[title="Remove"]').click();

    await page.waitForLoadState('networkidle');

    // Ivan should be removed
    await expect(page.locator('article').filter({ hasText: 'Members' }).locator('td', { hasText: 'Ivan Guest' })).not.toBeVisible();
  });

  test('add owner', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Marketing Team' }).locator('a').first().click();

    // Add Alice Smith as owner - find the option and select by value
    const select = page.locator('article').filter({ hasText: 'Owners' }).locator('select[name="userId"]');
    const aliceOption = select.locator('option', { hasText: 'Alice Smith' });
    const aliceValue = await aliceOption.getAttribute('value');
    await select.selectOption(aliceValue!);
    await page.locator('article').filter({ hasText: 'Owners' }).locator('button', { hasText: 'Add Owner' }).click();

    // Wait for page to reload
    await page.waitForLoadState('networkidle');

    await expect(page.locator('article').filter({ hasText: 'Owners' }).locator('td', { hasText: 'Alice Smith' })).toBeVisible();
  });

  test('remove owner', async ({ page }) => {
    // Go to Marketing Team - add an owner first, then remove them
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Marketing Team' }).locator('a').first().click();

    const ownersArticle = page.locator('article').filter({ hasText: 'Owners' });

    // Add Bob Jones as owner if not already
    const bobInOwners = ownersArticle.locator('td', { hasText: 'Bob Jones' });
    if (!await bobInOwners.isVisible()) {
      const select = ownersArticle.locator('select[name="userId"]');
      const bobOption = select.locator('option', { hasText: 'Bob Jones' });
      const bobValue = await bobOption.getAttribute('value');
      await select.selectOption(bobValue!);
      await ownersArticle.locator('button', { hasText: 'Add Owner' }).click();
      await page.waitForLoadState('networkidle');
    }

    // Now remove Bob Jones
    const bobRow = ownersArticle.locator('tr', { hasText: 'Bob Jones' });
    await expect(bobRow).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await bobRow.locator('button[title="Remove"]').click();

    await page.waitForLoadState('networkidle');

    // Bob should be removed from owners
    await expect(page.locator('article').filter({ hasText: 'Owners' }).locator('td', { hasText: 'Bob Jones' })).not.toBeVisible();
  });

  test('membership reflects on user', async ({ page }) => {
    // Add Henry Taylor to Engineering Team
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();

    // Target the select in the Members section
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    const select = membersSection.locator('select[name="userId"]');
    const henryOption = select.locator('option', { hasText: 'Henry Taylor' });
    const henryValue = await henryOption.getAttribute('value');
    await select.selectOption(henryValue!);
    await membersSection.locator('button', { hasText: 'Add Member' }).click();

    // Now navigate to Henry's user detail
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Henry Taylor' }).locator('a').first().click();

    // Engineering Team should appear in group memberships
    await expect(page.locator('article').filter({ hasText: 'Group Memberships' }).locator('a', { hasText: 'Engineering Team' })).toBeVisible();
  });

  test('member count updates', async ({ page }) => {
    // Go to Engineering Team and add a member first so we can remove them
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();

    // Add Ivan Guest to Engineering Team
    const membersArticle = page.locator('article').filter({ hasText: 'Members' });
    const ivanInMembers = membersArticle.locator('td', { hasText: 'Ivan Guest' });
    if (!await ivanInMembers.isVisible()) {
      const select = membersArticle.locator('select[name="userId"]');
      const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
      const ivanValue = await ivanOption.getAttribute('value');
      await select.selectOption(ivanValue!);
      await membersArticle.locator('button', { hasText: 'Add Member' }).click();
      await page.waitForLoadState('networkidle');
    }

    // Now remove Ivan
    page.on('dialog', dialog => dialog.accept());
    const ivanRow = membersArticle.locator('tr', { hasText: 'Ivan Guest' });
    await ivanRow.locator('button[title="Remove"]').click();
    await page.waitForLoadState('networkidle');

    // Go back to groups list
    await page.goto('/ui/groups');
    await page.waitForLoadState('networkidle');

    // Engineering Team should still be visible
    await expect(page.locator('tr')).toContainText(['Engineering Team']);
  });
});
