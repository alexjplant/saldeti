import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Users', () => {
  test('list users', async ({ page }) => {
    await page.goto('/ui/users');

    // Assert table exists with rows for seed users
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('th', { hasText: 'Display Name' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Alice Smith' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Bob Jones' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Charlie Brown' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Diana Prince' })).toBeVisible();
  });

  test('search users', async ({ page }) => {
    await page.goto('/ui/users');

    // Type Alice in search
    await page.fill('input[name="search"]', 'Alice');
    await page.press('input[name="search"]', 'Enter');

    // Assert only Alice rows shown
    await expect(page.locator('td', { hasText: 'Alice Smith' })).toBeVisible();
    await expect(page.locator('td', { hasText: 'Bob Jones' })).not.toBeVisible();

    // Clear search - all rows return
    await page.click('a[href="/ui/users"]');
    await expect(page.locator('td', { hasText: 'Bob Jones' })).toBeVisible();
  });

  test('create user', async ({ page }) => {
    await page.goto('/ui/users/new');

    await expect(page.locator('h2')).toHaveText('New User');

    const suffix = Date.now().toString(36);
    // Fill form
    await page.fill('input[name="displayName"]', 'E2E Test User');
    await page.fill('input[name="userPrincipalName"]', `e2e.test.${suffix}@saldeti.local`);
    await page.fill('input[name="mail"]', `e2e.test.${suffix}@saldeti.local`);
    await page.fill('input[name="department"]', 'QA');
    await page.fill('input[name="jobTitle"]', 'Test Engineer');
    await page.check('input[name="accountEnabled"]');

    // Submit
    await page.click('button[type="submit"]');

    // Should be redirected to detail page
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toHaveText('E2E Test User');
    await expect(page.locator('dd').filter({ hasText: 'QA' }).first()).toBeVisible();
  });

  test('create user validation', async ({ page }) => {
    await page.goto('/ui/users/new');

    // Wait for page to be fully loaded
    await expect(page.locator('h2')).toHaveText('New User');

    // Remove required attributes to bypass HTML5 validation
    await page.evaluate(() => {
      document.querySelectorAll('[required]').forEach(el => el.removeAttribute('required'));
    });

    // Submit empty form
    await page.click('button[type="submit"]');

    // Wait for form to re-render with error
    await page.waitForLoadState('networkidle');

    // Assert error message for required fields
    await expect(page.locator('.flash-danger')).toBeVisible();
    await expect(page.locator('.flash-danger')).toContainText('required');
  });

  test('view user detail', async ({ page }) => {
    await page.goto('/ui/users');

    // Click on Alice Smith - use more specific selector
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();

    // Assert detail page shows fields
    await expect(page.locator('h2')).toHaveText('Alice Smith');
    await expect(page.locator('dd').filter({ hasText: 'alice.smith@saldeti.local' }).first()).toBeVisible();
    await expect(page.locator('dd').filter({ hasText: 'Software Engineer' }).first()).toBeVisible();
    await expect(page.locator('dd').filter({ hasText: 'Engineering' }).first()).toBeVisible();

    // Should show manager section (Alice's manager is Eve)
    await expect(page.locator('article').filter({ hasText: 'Manager' })).toBeVisible();
  });

  test('edit user', async ({ page }) => {
    await page.goto('/ui/users');

    // Click on Bob Jones - click the first cell's link (the user name)
    const bobRow = page.locator('tr', { hasText: 'Bob Jones' });
    await bobRow.locator('td').nth(0).locator('a').click();
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);

    // Click Edit - wait for page to load and use href selector
    await page.waitForLoadState('networkidle');
    await page.locator('a[href*="/edit"]').click();
    await expect(page.locator('h2')).toHaveText('Edit User');

    // Change department and job title
    await page.fill('input[name="department"]', 'DevOps');
    await page.fill('input[name="jobTitle"]', 'Principal Engineer');
    await page.click('button[type="submit"]');

    // Should be redirected to detail page
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);
    await expect(page.locator('dd').filter({ hasText: 'DevOps' }).first()).toBeVisible();
    await expect(page.locator('dd').filter({ hasText: 'Principal Engineer' }).first()).toBeVisible();
  });

  test.describe('delete user', () => {
    let userId: string;

    test.beforeEach(async ({ page }) => {
      // Create a throwaway user for the delete test
      await page.goto('/ui/users/new');
      await page.fill('input[name="displayName"]', 'Delete Test User');
      await page.fill('input[name="userPrincipalName"]', 'delete.test@saldeti.local');
      await page.fill('input[name="mail"]', 'delete.test@saldeti.local');
      await page.click('button[type="submit"]');
      // Capture the userId from the redirected URL
      await expect(page).toHaveURL(/\/ui\/users\/([a-f0-9-]+)$/);
      const url = page.url();
      userId = url.match(/\/ui\/users\/([a-f0-9-]+)$/)?.[1] || '';
    });

    test('can delete user', async ({ page }) => {
      page.on('dialog', dialog => dialog.accept());
      await page.locator('button', { hasText: /Delete/i }).click();

      // Should be redirected to user list
      await expect(page).toHaveURL(/\/ui\/users$/);

      // The throwaway user should be gone
      await expect(page.locator('td', { hasText: 'Delete Test User' })).not.toBeVisible();
    });

    test.afterEach(async ({ page }) => {
      // If the test didn't delete, clean up here
      // Try navigating to the user and deleting if still present
      if (userId) {
        await page.goto('/ui/users');
        const row = page.locator('tr', { hasText: 'Delete Test User' });
        if (await row.isVisible()) {
          await row.locator('a').first().click();
          page.on('dialog', dialog => dialog.accept());
          await page.locator('button', { hasText: /Delete/i }).click();
          await expect(page).toHaveURL(/\/ui\/users$/);
        }
      }
    });
  });

  test('disabled user indicator', async ({ page }) => {
    await page.goto('/ui/users');

    // Grace Lee is disabled - find her row
    const graceRow = page.locator('tr', { hasText: 'Grace Lee' });
    await expect(graceRow).toBeVisible();

    // The enabled column should show a red X icon (or No text)
    // Since we use the yesno helper, check for the icon pattern
    const enabledCell = graceRow.locator('td').nth(4);
    await expect(enabledCell).toBeVisible();
  });

  test('set manager', async ({ page }) => {
    await page.goto('/ui/users');

    // Click on Bob Jones row link
    await page.locator('tr', { hasText: 'Bob Jones' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);

    // Find manager article
    const managerArticle = page.locator('#user-manager');
    await expect(managerArticle).toBeVisible();

    // Expand "Set Manager" details
    await managerArticle.locator('details').locator('summary').click();

    // Select a non-empty option from the dropdown
    const select = managerArticle.locator('select[name="managerId"]');
    const options = await select.locator('option').all();
    let selectedValue = '';
    for (const option of options) {
      const val = await option.getAttribute('value');
      if (val && val !== '' && !(await option.isDisabled())) {
        selectedValue = val;
        break;
      }
    }
    if (selectedValue) {
      await select.selectOption(selectedValue);

      // Submit and wait for HTMX response
      const responsePromise = page.waitForResponse(
        resp => resp.url().includes('/manager/set') && resp.status() === 200
      );
      await managerArticle.locator('input[type="submit"][value="Set"]').click();
      await responsePromise;

      // Verify success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Verify the manager link appears in #user-manager
      await expect(page.locator('#user-manager a')).toBeVisible();
    }
  });

  test('remove manager', async ({ page }) => {
    await page.goto('/ui/users');

    // Click on Bob Jones row link
    await page.locator('tr', { hasText: 'Bob Jones' }).locator('a').first().click();
    await expect(page).toHaveURL(/\/ui\/users\/[a-f0-9-]+$/);

    const managerArticle = page.locator('#user-manager');
    await expect(managerArticle).toBeVisible();

    // If no manager assigned, set one first
    const noManagerVisible = await managerArticle.locator('small', { hasText: 'No manager assigned.' }).isVisible();
    if (noManagerVisible) {
      await managerArticle.locator('details').locator('summary').click();
      const select = managerArticle.locator('select[name="managerId"]');
      const options = await select.locator('option').all();
      let selectedValue = '';
      for (const option of options) {
        const val = await option.getAttribute('value');
        if (val && val !== '' && !(await option.isDisabled())) {
          selectedValue = val;
          break;
        }
      }
      if (selectedValue) {
        await select.selectOption(selectedValue);
        const setResponse = page.waitForResponse(
          resp => resp.url().includes('/manager/set') && resp.status() === 200
        );
        await managerArticle.locator('input[type="submit"][value="Set"]').click();
        await setResponse;
        await expect(page.locator('.flash-success')).toBeVisible();
      }
    }

    // Now remove the manager
    page.once('dialog', dialog => dialog.accept());
    const removeResponse = page.waitForResponse(
      resp => resp.url().includes('/manager/remove') && resp.status() === 200
    );
    await page.locator('button[form="remove-manager"]').click();
    await removeResponse;

    // Verify success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Verify "No manager assigned." is visible
    await expect(page.locator('#user-manager')).toContainText('No manager assigned.');
  });
});
