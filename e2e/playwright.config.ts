import { defineConfig } from '@playwright/test';
import { resolve } from 'path';

const projectRoot = resolve(__dirname, '..');

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: 0,
  fullyParallel: false, // Run tests serially to avoid conflicts
  workers: 1,
  use: {
    baseURL: 'http://localhost:9443',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: `${resolve(projectRoot, 'bin', 'saldeti')} -port 9443 -seed ${resolve(projectRoot, 'seed.json')}`,
    port: 9443,
    reuseExistingServer: !process.env.CI,
    timeout: 10_000,
    cwd: projectRoot,
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
});
