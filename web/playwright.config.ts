import { defineConfig, devices } from '@playwright/test';

// The e2e run drives the REAL stack: the Go server serving the built UI and the
// live /api/report for the bundled example project, on a single origin.
const PORT = 8138;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: 'list',
  use: {
    baseURL: `http://localhost:${PORT}`,
    trace: 'retain-on-failure',
  },
  webServer: {
    command: 'bash e2e/serve.sh',
    url: `http://localhost:${PORT}/api/healthz`,
    timeout: 120_000,
    reuseExistingServer: !process.env.CI,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
