import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/ui');
});

test.describe('Licenses', () => {
  test('licensed user shows license table', async ({ page }) => {
    // Navigate to Alice's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Alice Smith' }).locator('a').first().click();

    // Assert the Licenses section is visible
    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });
    await expect(licensesSection).toBeVisible();

    // Assert SPE_E3 appears in the license table
    await expect(licensesSection.locator('td', { hasText: 'SPE_E3' })).toBeVisible();

    // Assert MCOSTANDARD appears as a disabled plan
    await expect(licensesSection.locator('td', { hasText: 'MCOSTANDARD' })).toBeVisible();
  });

  test('licensed user shows multiple licenses', async ({ page }) => {
    // Navigate to Eve's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Eve Wilson' }).locator('a').first().click();

    // Assert the Licenses section is visible
    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });
    await expect(licensesSection).toBeVisible();

    // Assert both SPE_E5 and EMS appear in the license table
    await expect(licensesSection.locator('td', { hasText: 'SPE_E5' })).toBeVisible();
    await expect(licensesSection.locator('td', { hasText: 'EMS' })).toBeVisible();

    // Should have 2 rows in the table (count rows in tbody)
    const licenseRows = licensesSection.locator('tbody tr');
    await expect(licenseRows).toHaveCount(2);
  });

  test('unlicensed user shows no licenses message', async ({ page }) => {
    // Navigate to Grace's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Grace Lee' }).locator('a').first().click();

    // Assert "No licenses assigned" text appears
    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });
    await expect(licensesSection.locator('p', { hasText: 'No licenses assigned' })).toBeVisible();
  });

  test('add license to user', async ({ page }) => {
    // Navigate to Grace's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Grace Lee' }).locator('a').first().click();

    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });

    // Expand the "Add License" details element
    const addLicenseDetails = licensesSection.locator('details summary', { hasText: 'Add License' });
    await addLicenseDetails.click();

    // Select INTUNE_A from dropdown
    const select = licensesSection.locator('select[name="skuId"]');
    await select.selectOption('061f9ace-7d42-4136-88ac-31dc755f143f');

    // Click Assign
    await licensesSection.locator('button', { hasText: 'Assign' }).click();

    // Wait for page reload
    await page.waitForLoadState('networkidle');

    // Assert INTUNE_A now appears in the license table
    await expect(licensesSection.locator('td', { hasText: 'INTUNE_A' })).toBeVisible();
  });

  test('remove license from user', async ({ page }) => {
    // Navigate to Bob's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Bob Jones' }).locator('a').first().click();

    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });

    // Find the SPE_E3 row
    const speE3Row = licensesSection.locator('tr', { hasText: 'SPE_E3' });
    await expect(speE3Row).toBeVisible();

    // Click the remove button
    await speE3Row.locator('button[title="Remove license"]').click();

    // Wait for page reload
    await page.waitForLoadState('networkidle');

    // Assert SPE_E3 is gone
    await expect(licensesSection.locator('td', { hasText: 'SPE_E3' })).not.toBeVisible();

    // Assert "No licenses assigned" appears
    await expect(licensesSection.locator('p', { hasText: 'No licenses assigned' })).toBeVisible();
  });

  test('available SKUs exclude already assigned', async ({ page }) => {
    // Navigate to Admin's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Admin User' }).locator('a').first().click();

    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });

    // Expand "Add License"
    const addLicenseDetails = licensesSection.locator('details summary', { hasText: 'Add License' });
    await addLicenseDetails.click();

    // The select dropdown should NOT have SPE_E5 as an option
    const select = licensesSection.locator('select[name="skuId"]');
    const speE5Option = select.locator('option[value="06ebc4ee-1bb5-47dd-8120-11324bc54e06"]');
    await expect(speE5Option).toHaveCount(0);
  });

  test('license persists after page reload', async ({ page }) => {
    // Navigate to Grace's detail page
    await page.goto('/ui/users');
    await page.locator('tr', { hasText: 'Grace Lee' }).locator('a').first().click();

    const licensesSection = page.locator('article').filter({ hasText: 'Licenses' });

    // Expand the "Add License" details element
    const addLicenseDetails = licensesSection.locator('details summary', { hasText: 'Add License' });
    await addLicenseDetails.click();

    // Select O365_BUSINESS_ESSENTIALS from dropdown
    const select = licensesSection.locator('select[name="skuId"]');
    await select.selectOption('3b555118-da6a-4418-894f-7df1e2096870');

    // Click Assign
    await licensesSection.locator('button', { hasText: 'Assign' }).click();

    // Wait for page reload
    await page.waitForLoadState('networkidle');

    // Navigate away to user list
    await page.goto('/ui/users');
    await page.waitForLoadState('networkidle');

    // Come back to Grace's detail
    await page.locator('tr', { hasText: 'Grace Lee' }).locator('a').first().click();
    await page.waitForLoadState('networkidle');

    // Assert O365_BUSINESS_ESSENTIALS is still there
    await expect(page.locator('article').filter({ hasText: 'Licenses' }).locator('td', { hasText: 'O365_BUSINESS_ESSENTIALS' })).toBeVisible();
  });
});
