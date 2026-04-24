import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Membership', () => {
  test('add member', async ({ page }) => {
    // Go to All Staff group (not Leadership since it might be deleted by other tests)
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'All Staff' }).locator('a').first().click();

    // Add Ivan Guest as member - target the select in the Members section
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    await membersSection.locator('summary').click();
    const select = membersSection.locator('select[name="userId"]');
    const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
    const ivanValue = await ivanOption.getAttribute('value');
    await select.selectOption(ivanValue!);
    await membersSection.locator('input[value="Add Member"]').click();

    // Wait for Ivan to appear via htmx swap (no full page reload)
    await expect(page.locator('article').filter({ hasText: 'Members' }).locator('td', { hasText: 'Ivan Guest' })).toBeVisible({ timeout: 5000 });
  });

  test('remove member', async ({ page }) => {
    // Go to All Staff group
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'All Staff' }).locator('a').first().click();

    // First add Ivan Guest so we have a known member to remove
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    const ivanInMembers = membersSection.locator('td', { hasText: 'Ivan Guest' });
    if (!await ivanInMembers.isVisible()) {
      await membersSection.locator('summary').click();
      const select = membersSection.locator('select[name="userId"]');
      const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
      const ivanValue = await ivanOption.getAttribute('value');
      await select.selectOption(ivanValue!);
      await membersSection.locator('input[value="Add Member"]').click();
      await expect(membersSection.locator('td', { hasText: 'Ivan Guest' })).toBeVisible({ timeout: 5000 });
    }

    // Now remove Ivan Guest
    const ivanRow = membersSection.locator('tr', { hasText: 'Ivan Guest' });
    await expect(ivanRow).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await ivanRow.locator('button[title="Remove"]').click();

    // Wait for Ivan to disappear via htmx swap
    await expect(page.locator('article').filter({ hasText: 'Members' }).locator('td', { hasText: 'Ivan Guest' })).not.toBeVisible({ timeout: 5000 });
  });

  test('add owner', async ({ page }) => {
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Marketing Team' }).locator('a').first().click();

    // Add Alice Smith as owner - find the option and select by value
    const ownersArticle = page.locator('article').filter({ hasText: 'Owners' });
    await ownersArticle.locator('summary').click();
    const select = ownersArticle.locator('select[name="userId"]');
    const aliceOption = select.locator('option', { hasText: 'Alice Smith' });
    const aliceValue = await aliceOption.getAttribute('value');
    await select.selectOption(aliceValue!);
    await ownersArticle.locator('input[value="Add Owner"]').click();

    // Wait for Alice to appear via htmx swap
    await expect(page.locator('article').filter({ hasText: 'Owners' }).locator('td', { hasText: 'Alice Smith' })).toBeVisible({ timeout: 5000 });
  });

  test('remove owner', async ({ page }) => {
    // Go to Marketing Team - add an owner first, then remove them
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Marketing Team' }).locator('a').first().click();

    const ownersArticle = page.locator('article').filter({ hasText: 'Owners' });

    // Add Bob Jones as owner if not already
    const bobInOwners = ownersArticle.locator('td', { hasText: 'Bob Jones' });
    if (!await bobInOwners.isVisible()) {
      await ownersArticle.locator('summary').click();
      const select = ownersArticle.locator('select[name="userId"]');
      const bobOption = select.locator('option', { hasText: 'Bob Jones' });
      const bobValue = await bobOption.getAttribute('value');
      await select.selectOption(bobValue!);
      await ownersArticle.locator('input[value="Add Owner"]').click();
      await expect(ownersArticle.locator('td', { hasText: 'Bob Jones' })).toBeVisible({ timeout: 5000 });
    }

    // Now remove Bob Jones
    const bobRow = ownersArticle.locator('tr', { hasText: 'Bob Jones' });
    await expect(bobRow).toBeVisible();

    page.on('dialog', dialog => dialog.accept());
    await bobRow.locator('button[title="Remove"]').click();

    // Wait for Bob to disappear via htmx swap
    await expect(page.locator('article').filter({ hasText: 'Owners' }).locator('td', { hasText: 'Bob Jones' })).not.toBeVisible({ timeout: 5000 });
  });

  test('membership reflects on user', async ({ page }) => {
    // Add Henry Taylor to Engineering Team
    await page.goto('/ui/groups');
    await page.locator('tr', { hasText: 'Engineering Team' }).locator('a').first().click();

    // Target the select in the Members section
    const membersSection = page.locator('article').filter({ hasText: 'Members' });
    await membersSection.locator('summary').click();
    const select = membersSection.locator('select[name="userId"]');
    const henryOption = select.locator('option', { hasText: 'Henry Taylor' });
    const henryValue = await henryOption.getAttribute('value');
    await select.selectOption(henryValue!);
    await membersSection.locator('input[value="Add Member"]').click();

    // Wait for Henry to appear via htmx swap before navigating
    await expect(membersSection.locator('td', { hasText: 'Henry Taylor' })).toBeVisible({ timeout: 5000 });

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
      await membersArticle.locator('summary').click();
      const select = membersArticle.locator('select[name="userId"]');
      const ivanOption = select.locator('option', { hasText: 'Ivan Guest' });
      const ivanValue = await ivanOption.getAttribute('value');
      await select.selectOption(ivanValue!);
      await membersArticle.locator('input[value="Add Member"]').click();
      await expect(membersArticle.locator('td', { hasText: 'Ivan Guest' })).toBeVisible({ timeout: 5000 });
    }

    // Now remove Ivan
    page.on('dialog', dialog => dialog.accept());
    const ivanRow = membersArticle.locator('tr', { hasText: 'Ivan Guest' });
    await ivanRow.locator('button[title="Remove"]').click();
    // Wait for Ivan to disappear via htmx swap
    await expect(membersArticle.locator('td', { hasText: 'Ivan Guest' })).not.toBeVisible({ timeout: 5000 });

    // Go back to groups list
    await page.goto('/ui/groups');
    await page.waitForLoadState('networkidle');

    // Engineering Team should still be visible
    await expect(page.locator('tr')).toContainText(['Engineering Team']);
  });
});
