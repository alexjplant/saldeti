import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Applications', () => {
  test('list applications', async ({ page }) => {
    await page.goto('/ui/applications');

    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('h2')).toContainText('App Registrations');
  });

  test('search applications', async ({ page }) => {
    await page.goto('/ui/applications');

    await page.fill('input[name="search"]', 'Simulator');
    await page.press('input[name="search"]', 'Enter');

    // Should filter results
    await expect(page.locator('table')).toBeVisible();
  });

  test('create application', async ({ page }) => {
    await page.goto('/ui/applications/new');

    await expect(page.locator('h2')).toHaveText('New Application');

    await page.fill('input[name="displayName"]', 'E2E Test App');
    await page.fill('textarea[name="description"]', 'App created by E2E test');
    await page.selectOption('select[name="signInAudience"]', 'AzureADMyOrg');

    await page.click('button[type="submit"]');

    // Should be redirected to detail
    await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);
    await expect(page.locator('h2')).toHaveText('E2E Test App');
    await expect(page.locator('dd', { hasText: 'App created by E2E test' })).toBeVisible();
  });

  test('create application validation', async ({ page }) => {
    await page.goto('/ui/applications/new');

    await expect(page.locator('h2')).toHaveText('New Application');

    // Remove required attributes to bypass HTML5 validation
    await page.evaluate(() => {
      document.querySelectorAll('[required]').forEach(el => el.removeAttribute('required'));
    });

    await page.click('button[type="submit"]');

    await page.waitForLoadState('networkidle');

    await expect(page.locator('.flash-danger')).toBeVisible();
  });

  test('view application detail', async ({ page }) => {
    await page.goto('/ui/applications');

    // Click on the first application in the list
    const firstAppLink = page.locator('table tbody tr').first().locator('a').first();
    await firstAppLink.click();

    // Should show detail page
    await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);
    await expect(page.locator('article').filter({ hasText: 'Application Info' })).toBeVisible();
  });

  test('edit application', async ({ page }) => {
    // Create an app first
    await page.goto('/ui/applications/new');
    await page.fill('input[name="displayName"]', 'Edit Test App');
    await page.fill('textarea[name="description"]', 'Original description');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);

    // Click Edit
    await page.waitForLoadState('networkidle');
    await page.locator('a[href*="/edit"]').click();
    await expect(page.locator('h2')).toHaveText('Edit Application');

    await page.fill('textarea[name="description"]', 'Updated by E2E test');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);
    await expect(page.locator('dd').filter({ hasText: 'Updated by E2E test' }).first()).toBeVisible();
  });

  test.describe('delete application', () => {
    let appId: string;

    test.beforeEach(async ({ page }) => {
      await page.goto('/ui/applications/new');
      await page.fill('input[name="displayName"]', 'Delete Test App');
      await page.fill('textarea[name="description"]', 'App to be deleted');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/\/ui\/applications\/([a-f0-9-]+)$/);
      const url = page.url();
      appId = url.match(/\/ui\/applications\/([a-f0-9-]+)$/)?.[1] || '';
    });

    test('can delete application', async ({ page }) => {
      page.on('dialog', dialog => dialog.accept());
      await page.locator('button', { hasText: /Delete/i }).click();

      await expect(page).toHaveURL(/\/ui\/applications$/);
      await expect(page.locator('tr', { hasText: 'Delete Test App' })).not.toBeVisible();
    });

    test.afterEach(async ({ page }) => {
      if (appId) {
        await page.goto('/ui/applications');
        const row = page.locator('tr', { hasText: 'Delete Test App' });
        if (await row.isVisible()) {
          await row.locator('a').first().click();
          page.on('dialog', dialog => dialog.accept());
          await page.locator('button', { hasText: /Delete/i }).click();
          await expect(page).toHaveURL(/\/ui\/applications$/);
        }
      }
    });
  });

  test.describe('Credential Management', () => {
    async function createAppAndNavigate(page: import('@playwright/test').Page, name: string) {
      await page.goto('/ui/applications/new');
      await page.fill('input[name="displayName"]', name);
      await page.fill('textarea[name="description"]', 'App for credential tests');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);
    }

    test('add password credential', async ({ page }) => {
      await createAppAndNavigate(page, 'Cred Test App');

      // Expand "New Client Secret" details element
      await page.locator('details').filter({ hasText: 'New Client Secret' }).locator('summary').click();

      // Fill the credential display name
      await page.fill('input[name="credentialDisplayName"]', 'E2E Test Secret');

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
        page.locator('#credentials table tbody tr').first().locator('td').first()
      ).toContainText('E2E Test Secret');
    });

    test('remove password credential', async ({ page }) => {
      await createAppAndNavigate(page, 'Cred Remove Test App');

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
      await page.locator('#credentials button.outline.danger').first().click();
      await removeResponse;

      // Assert success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Assert the password is gone — should show empty message
      await expect(page.locator('#credentials')).toContainText('No client secrets.');
    });

    test('add key credential', async ({ page }) => {
      await createAppAndNavigate(page, 'Key Test App');

      // Expand "Upload Certificate" details element
      await page.locator('details').filter({ hasText: 'Upload Certificate' }).locator('summary').click();

      // Fill the key display name
      await page.fill('input[name="keyDisplayName"]', 'E2E Test Cert');

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
        page.locator('#credentials table')
          .filter({ has: page.locator('th', { hasText: 'Type' }) })
          .locator('tbody tr').first().locator('td').first()
      ).toContainText('E2E Test Cert');
    });

    test('remove key credential', async ({ page }) => {
      await createAppAndNavigate(page, 'Key Remove Test App');

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
      await page.locator('#credentials table')
        .filter({ has: page.locator('th', { hasText: 'Type' }) })
        .locator('button.outline.danger').first().click();
      await removeResponse;

      // Assert success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Assert the key is gone — should show empty message
      await expect(page.locator('#credentials')).toContainText('No certificates.');
    });
  });

  test.describe('Extension Properties', () => {
    async function createAppForExtensions(page: import('@playwright/test').Page, name: string) {
      await page.goto('/ui/applications/new');
      await page.fill('input[name="displayName"]', name);
      await page.fill('textarea[name="description"]', 'App for extension tests');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/\/ui\/applications\/[a-f0-9-]+$/);
    }

    test('create extension property', async ({ page }) => {
      await createAppForExtensions(page, 'Ext Test App');

      // Expand "New Extension Property" details element
      await page.locator('details').filter({ hasText: 'New Extension Property' }).locator('summary').click();

      // Fill the extension name and data type
      await page.fill('input[name="name"]', 'extension_e2e_test');
      await page.selectOption('select[name="dataType"]', 'String');

      // Submit and wait for HTMX response
      const responsePromise = page.waitForResponse(
        resp => resp.url().includes('/extensions/create') && resp.status() === 200
      );
      await page.click('input[type="submit"][value="Create Extension Property"]');
      await responsePromise;

      // Assert success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Assert the extensions table contains the new extension
      await expect(page.locator('#extensions table')).toContainText('extension_e2e_test');
    });

    test('delete extension property', async ({ page }) => {
      await createAppForExtensions(page, 'Ext Delete Test App');

      // Create an extension first
      await page.locator('details').filter({ hasText: 'New Extension Property' }).locator('summary').click();
      await page.fill('input[name="name"]', 'extension_e2e_del_test');
      await page.selectOption('select[name="dataType"]', 'String');
      const addResponse = page.waitForResponse(
        resp => resp.url().includes('/extensions/create') && resp.status() === 200
      );
      await page.click('input[type="submit"][value="Create Extension Property"]');
      await addResponse;
      await expect(page.locator('.flash-success')).toBeVisible();

      // Delete it — use page.once to avoid handler accumulation
      page.once('dialog', dialog => dialog.accept());
      const deleteResponse = page.waitForResponse(
        resp => resp.url().includes('/extensions/') && resp.url().includes('/delete') && resp.status() === 200
      );
      await page.locator('#extensions button.outline.danger').first().click();
      await deleteResponse;

      // Assert success flash
      await expect(page.locator('.flash-success')).toBeVisible();

      // Assert the extension is gone — should show empty message
      await expect(page.locator('#extensions')).toContainText('No extension properties.');
    });
  });
});
