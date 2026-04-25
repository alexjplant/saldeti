import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Service Principals', () => {
  test('list service principals', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('h2')).toContainText('Enterprise Apps');
  });

  test('search service principals', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    await page.fill('input[name="search"]', 'Simulator');
    await page.press('input[name="search"]', 'Enter');

    await expect(page.locator('table')).toBeVisible();
  });

  test('view service principal detail', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    // Click on the first SP in the list
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();

    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);
    await expect(page.locator('article').filter({ hasText: 'General Info' })).toBeVisible();
  });

  test('seed SP exists', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    // The simulator should have at least one SP from seed data
    const rows = page.locator('table tbody tr');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('SP detail shows memberOf section', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    const memberOfArticle = page.locator('article').filter({ hasText: 'Member Of' });
    await expect(memberOfArticle).toBeVisible();

    // Verify section renders with either table or empty message
    const hasTable = await memberOfArticle.locator('table').count();
    const hasEmptyText = await memberOfArticle.locator('p', { hasText: 'Not a member of any groups.' }).count();
    expect(hasTable + hasEmptyText).toBeGreaterThan(0);
  });

  test('SP detail shows appRoleAssignments section', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    const appRoleArticle = page.locator('article').filter({ hasText: 'App Role Assignments' });
    await expect(appRoleArticle).toBeVisible();

    const hasTable = await appRoleArticle.locator('table').count();
    const hasEmptyText = await appRoleArticle.locator('p', { hasText: 'No app role assignments.' }).count();
    expect(hasTable + hasEmptyText).toBeGreaterThan(0);
  });

  test('SP detail shows oauth2PermissionGrants section', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    const oauth2Article = page.locator('article').filter({ hasText: 'OAuth2 Permission Grants' });
    await expect(oauth2Article).toBeVisible();

    const hasTable = await oauth2Article.locator('table').count();
    const hasEmptyText = await oauth2Article.locator('p', { hasText: 'No OAuth2 permission grants.' }).count();
    expect(hasTable + hasEmptyText).toBeGreaterThan(0);
  });

  test('SP owner management — add owner', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    // Verify owners section is visible
    const ownersArticle = page.locator('#sp-owners');
    await expect(ownersArticle).toBeVisible();

    // Expand the "Add Owner" details element
    await ownersArticle.locator('details').locator('summary').click();

    // Select a non-empty option from the dropdown
    const select = ownersArticle.locator('select[name="userId"]');
    const options = await select.locator('option').all();
    let selectedValue = '';
    for (const option of options) {
      const val = await option.getAttribute('value');
      if (val && val !== '') {
        selectedValue = val;
        break;
      }
    }
    if (selectedValue) {
      await select.selectOption(selectedValue);

      // Submit and wait for HTMX response
      const responsePromise = page.waitForResponse(
        resp => resp.url().includes('/owners/add') && resp.status() === 200
      );
      await ownersArticle.locator('input[type="submit"][value="Add Owner"]').click();
      await responsePromise;

      // Verify success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Verify "No owners." is gone (owners list now has entries)
      await expect(page.locator('#sp-owners p', { hasText: 'No owners.' })).not.toBeVisible();
    }
  });

  test('SP pagination controls', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    await expect(page.locator('table')).toBeVisible();

    // Pagination nav may or may not be present depending on item count
    const paginationNav = page.locator('nav[aria-label="Pagination"]');
    const isPaginationVisible = await paginationNav.isVisible().catch(() => false);

    if (isPaginationVisible) {
      // Verify page info text is present
      await expect(paginationNav).toContainText(/Page \d+ of \d+/);
    }
    // If no pagination, the list simply loaded — that's acceptable
  });

  test('Non-existent SP redirects to list', async ({ page }) => {
    await page.goto('/ui/servicePrincipals/non-existent-id-12345');

    // Handler does 303 redirect back to list
    await expect(page).toHaveURL(/\/ui\/servicePrincipals$/);

    // Verify flash danger message appears
    await expect(page.locator('.flash-danger')).toBeVisible();
  });
});

test.describe('SP CRUD', () => {
  test('create service principal', async ({ page }) => {
    await page.goto('/ui/servicePrincipals/new');

    await expect(page.locator('h2')).toHaveText('New Service Principal');

    await page.fill('input[name="displayName"]', 'E2E Test SP');
    await page.fill('textarea[name="notes"]', 'SP created by E2E test');

    await page.click('button[type="submit"]');

    // Should be redirected to detail page
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toContainText('E2E Test SP');
  });

  test('edit service principal', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');

    // Click on the first SP in the list
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();

    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    // Click Edit
    await page.waitForLoadState('networkidle');
    await page.locator('a[href*="/edit"]').click();
    await expect(page.locator('h2')).toHaveText('Edit Service Principal');

    // Change display name
    await page.fill('input[name="displayName"]', 'E2E Edited SP');
    await page.click('button[type="submit"]');

    // Should be redirected to detail page
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toContainText('E2E Edited SP');
  });

  test.describe('delete service principal', () => {
    let spId: string;

    test.beforeEach(async ({ page }) => {
      // Create a throwaway SP for the delete test
      await page.goto('/ui/servicePrincipals/new');
      await page.fill('input[name="displayName"]', 'Delete Test SP');
      await page.fill('textarea[name="notes"]', 'SP to be deleted');
      await page.click('button[type="submit"]');
      // Capture the spId from the redirected URL
      await expect(page).toHaveURL(/\/ui\/servicePrincipals\/([a-f0-9-]+)$/);
      const url = page.url();
      spId = url.match(/\/ui\/servicePrincipals\/([a-f0-9-]+)$/)?.[1] || '';
    });

    test('can delete service principal', async ({ page }) => {
      page.on('dialog', dialog => dialog.accept());
      await page.locator('button[form="delete-sp"]').click();

      // Should be redirected to SP list
      await expect(page).toHaveURL(/\/ui\/servicePrincipals$/);

      // The throwaway SP should be gone
      await expect(page.locator('tr', { hasText: 'Delete Test SP' })).not.toBeVisible();
    });

    test.afterEach(async ({ page }) => {
      // Cleanup: if the test didn't delete, try to clean up
      if (spId) {
        await page.goto('/ui/servicePrincipals');
        const row = page.locator('tr', { hasText: 'Delete Test SP' });
        if (await row.isVisible()) {
          await row.locator('a').first().click();
          page.on('dialog', dialog => dialog.accept());
          await page.locator('button[form="delete-sp"]').click();
          await expect(page).toHaveURL(/\/ui\/servicePrincipals$/);
        }
      }
    });
  });
});

test.describe('Credential Management', () => {
  async function createSPAndNavigate(page: import('@playwright/test').Page, name: string) {
    await page.goto('/ui/servicePrincipals/new');
    await page.fill('input[name="displayName"]', name);
    await page.fill('textarea[name="notes"]', 'SP for credential tests');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);
  }

  test('add password credential', async ({ page }) => {
    await createSPAndNavigate(page, 'Cred Test SP');

    // Expand "New Client Secret" details element
    await page.locator('details').filter({ hasText: 'New Client Secret' }).locator('summary').click();

    // Fill the credential display name
    await page.fill('input[name="credentialDisplayName"]', 'E2E SP Test Secret');

    // Submit and wait for HTMX response
    const responsePromise = page.waitForResponse(
      resp => resp.url().includes('/credentials/password/add') && resp.status() === 200
    );
    await page.click('input[type="submit"][value="Add Secret"]');
    await responsePromise;

    // Assert success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Assert the password table has a row with the display name
    await expect(
      page.locator('#sp-credentials table tbody tr').first().locator('td').first()
    ).toContainText('E2E SP Test Secret');
  });

  test('remove password credential', async ({ page }) => {
    await createSPAndNavigate(page, 'Cred Remove Test SP');

    // Add a password first
    await page.locator('details').filter({ hasText: 'New Client Secret' }).locator('summary').click();
    await page.fill('input[name="credentialDisplayName"]', 'Secret To Remove');
    const addResponse = page.waitForResponse(
      resp => resp.url().includes('/credentials/password/add') && resp.status() === 200
    );
    await page.click('input[type="submit"][value="Add Secret"]');
    await addResponse;
    await expect(page.locator('.flash-success')).toBeVisible();

    // Remove it — use page.once to avoid handler accumulation
    page.once('dialog', dialog => dialog.accept());
    const removeResponse = page.waitForResponse(
      resp => resp.url().includes('/credentials/password/remove') && resp.status() === 200
    );
    await page.locator('#sp-credentials button.outline.danger').first().click();
    await removeResponse;

    // Assert success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Assert the password is gone — should show empty message
    await expect(page.locator('#sp-credentials')).toContainText('No client secrets.');
  });

  test('add key credential', async ({ page }) => {
    await createSPAndNavigate(page, 'Key Test SP');

    // Expand "Upload Certificate" details element
    await page.locator('details').filter({ hasText: 'Upload Certificate' }).locator('summary').click();

    // Fill the key display name
    await page.fill('input[name="keyDisplayName"]', 'E2E SP Test Cert');

    // Submit and wait for HTMX response
    const responsePromise = page.waitForResponse(
      resp => resp.url().includes('/credentials/key/add') && resp.status() === 200
    );
    await page.click('input[type="submit"][value="Add Certificate"]');
    await responsePromise;

    // Assert success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Assert the key table has a row with the display name
    // The key table is distinguishable by having a "Type" column header
    await expect(
      page.locator('#sp-credentials table')
        .filter({ has: page.locator('th', { hasText: 'Type' }) })
        .locator('tbody tr').first().locator('td').first()
    ).toContainText('E2E SP Test Cert');
  });

  test('remove key credential', async ({ page }) => {
    await createSPAndNavigate(page, 'Key Remove Test SP');

    // Add a key first
    await page.locator('details').filter({ hasText: 'Upload Certificate' }).locator('summary').click();
    await page.fill('input[name="keyDisplayName"]', 'Cert To Remove');
    const addResponse = page.waitForResponse(
      resp => resp.url().includes('/credentials/key/add') && resp.status() === 200
    );
    await page.click('input[type="submit"][value="Add Certificate"]');
    await addResponse;
    await expect(page.locator('.flash-success')).toBeVisible();

    // Remove it — use page.once to avoid handler accumulation
    page.once('dialog', dialog => dialog.accept());
    const removeResponse = page.waitForResponse(
      resp => resp.url().includes('/credentials/key/remove') && resp.status() === 200
    );
    // Find the danger button within the certificates table (has "Type" column header)
    await page.locator('#sp-credentials table')
      .filter({ has: page.locator('th', { hasText: 'Type' }) })
      .locator('button.outline.danger').first().click();
    await removeResponse;

    // Assert success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Assert the key is gone — should show empty message
    await expect(page.locator('#sp-credentials')).toContainText('No certificates.');
  });
});

test.describe('SP Owner Management', () => {
  test('add owner', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    // Verify owners section is visible
    const ownersArticle = page.locator('#sp-owners');
    await expect(ownersArticle).toBeVisible();

    // Expand the "Add Owner" details element
    await ownersArticle.locator('details').locator('summary').click();

    // Select a non-empty option from the dropdown
    const select = ownersArticle.locator('select[name="userId"]');
    const options = await select.locator('option').all();
    let selectedValue = '';
    for (const option of options) {
      const val = await option.getAttribute('value');
      if (val && val !== '') {
        selectedValue = val;
        break;
      }
    }
    if (selectedValue) {
      await select.selectOption(selectedValue);

      // Submit and wait for HTMX response
      const responsePromise = page.waitForResponse(
        resp => resp.url().includes('/owners/add') && resp.status() === 200
      );
      await ownersArticle.locator('input[type="submit"][value="Add Owner"]').click();
      await responsePromise;

      // Verify success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Verify "No owners." is gone
      await expect(page.locator('#sp-owners p', { hasText: 'No owners.' })).not.toBeVisible();
    }
  });

  test('remove owner', async ({ page }) => {
    await page.goto('/ui/servicePrincipals');
    const firstSPLink = page.locator('table tbody tr').first().locator('a').first();
    await firstSPLink.click();
    await expect(page).toHaveURL(/\/ui\/servicePrincipals\/[a-f0-9-]+$/);

    const ownersArticle = page.locator('#sp-owners');
    await expect(ownersArticle).toBeVisible();

    // Add an owner first
    await ownersArticle.locator('details').locator('summary').click();
    const select = ownersArticle.locator('select[name="userId"]');
    const options = await select.locator('option').all();
    let selectedValue = '';
    for (const option of options) {
      const val = await option.getAttribute('value');
      if (val && val !== '') {
        selectedValue = val;
        break;
      }
    }
    if (selectedValue) {
      await select.selectOption(selectedValue);
      const addResponse = page.waitForResponse(
        resp => resp.url().includes('/owners/add') && resp.status() === 200
      );
      await ownersArticle.locator('input[type="submit"][value="Add Owner"]').click();
      await addResponse;
      await expect(page.locator('.flash-success')).toBeVisible();
    }

    // Now remove the owner
    page.once('dialog', dialog => dialog.accept());
    const removeResponse = page.waitForResponse(
      resp => resp.url().includes('/owners/') && resp.url().includes('/remove') && resp.status() === 200
    );
    await page.locator('#sp-owners button.outline.danger').first().click();
    await removeResponse;

    // Verify success flash
    await expect(page.locator('.flash-success')).toBeVisible();

    // Verify "No owners." is visible
    await expect(page.locator('#sp-owners p', { hasText: 'No owners.' })).toBeVisible();
  });
});
