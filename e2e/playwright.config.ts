import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: 0,
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: 'https://localhost:9443',
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
});
