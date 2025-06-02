import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false, // Run sequentially for more stability
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 1, // Single worker for stability
  reporter: process.env.CI ? 'github' : 'line',
  timeout: 60000, // 60 second test timeout
  expect: {
    timeout: 10000, // 10 second assertion timeout
  },
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000, // 10 second action timeout
  },

  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Add headless mode for CI
        launchOptions: {
          args: ['--no-sandbox', '--disable-setuid-sandbox'],
        },
      },
    },
    // Only run chromium in CI for speed
    ...(process.env.CI ? [] : [
      {
        name: 'firefox',
        use: { ...devices['Desktop Firefox'] },
      },
    ]),
  ],

  // Start local web server before running tests
  webServer: {
    command: 'GOWORK=off go run . -web -web-port 8080 -pop-size 3',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 60000, // Increased timeout for server startup
    stdout: 'pipe',
    stderr: 'pipe',
  },
});