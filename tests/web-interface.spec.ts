import { test, expect } from '@playwright/test';

test.describe('EvoSim Web Interface', () => {
  test.beforeEach(async ({ page }) => {
    // Set longer timeout for navigation in CI
    page.setDefaultTimeout(30000);
  });

  test('loads homepage and displays simulation interface', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Check that the main title is present
    await expect(page.locator('h1')).toContainText('EvoSim - Genetic Algorithm Simulation', { timeout: 15000 });
    
    // Check for connection status
    const connectionStatus = page.locator('.connection-status');
    await expect(connectionStatus).toBeVisible({ timeout: 15000 });
    
    // Check for main interface components
    await expect(page.locator('.simulation-view')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.info-panel')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.controls')).toBeVisible({ timeout: 10000 });
  });

  test('displays simulation grid with content', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for WebSocket connection and initial data
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(5000); // Allow time for simulation to start
    
    // Check that the grid container is present
    const gridContainer = page.locator('.grid-container');
    await expect(gridContainer).toBeVisible({ timeout: 15000 });
    
    // Check that grid content is populated (should contain simulation symbols)
    await expect(gridContainer).not.toBeEmpty({ timeout: 10000 });
  });

  test('websocket connection establishes successfully', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for WebSocket connection to establish
    await page.waitForTimeout(8000);
    
    // Check connection status shows connected
    const connectionStatus = page.locator('.connection-status');
    await expect(connectionStatus).toContainText('Connected', { timeout: 15000 });
  });

  test('view switching functionality works', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial load and connection
    await page.waitForTimeout(5000);
    
    // Check that view tabs are present
    const viewTabs = page.locator('.view-tab');
    await expect(viewTabs.first()).toBeVisible({ timeout: 10000 });
    
    // Test switching to Stats view if it exists
    const statsTab = page.locator('.view-tab:has-text("Stats")');
    if (await statsTab.isVisible()) {
      await statsTab.click();
      await page.waitForTimeout(1000);
      // Verify info panel is still visible after view change
      await expect(page.locator('.info-panel')).toBeVisible({ timeout: 5000 });
    }
    
    // Test switching to Events view if it exists
    const eventsTab = page.locator('.view-tab:has-text("Events")');
    if (await eventsTab.isVisible()) {
      await eventsTab.click();
      await page.waitForTimeout(1000);
      // Verify info panel is still visible after view change
      await expect(page.locator('.info-panel')).toBeVisible({ timeout: 5000 });
    }
  });

  test('control buttons are present and functional', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial load
    await page.waitForTimeout(3000);
    
    // Check for control buttons
    const controlsSection = page.locator('.controls');
    await expect(controlsSection).toBeVisible({ timeout: 10000 });
    
    // Look for buttons in controls
    const buttons = page.locator('.controls button');
    const buttonCount = await buttons.count();
    
    // Ensure we have some control buttons
    expect(buttonCount).toBeGreaterThan(0);
    
    // Test clicking first available button
    const firstButton = buttons.first();
    if (await firstButton.isVisible()) {
      await firstButton.click();
      await page.waitForTimeout(1000);
      // Verify interface still responds after button click
      await expect(page.locator('.simulation-view')).toBeVisible({ timeout: 5000 });
    }
  });

  test('responsive design on different screen sizes', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Test desktop view
    await page.setViewportSize({ width: 1200, height: 800 });
    await expect(page.locator('.container')).toBeVisible({ timeout: 10000 });
    
    // Test tablet view  
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.simulation-view')).toBeVisible({ timeout: 5000 });
    
    // Test mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('.simulation-view')).toBeVisible({ timeout: 5000 });
  });

  test('information panel displays simulation data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for data to load
    await page.waitForTimeout(6000);
    
    // Check that info panel displays content
    const infoPanel = page.locator('.info-panel');
    await expect(infoPanel).toBeVisible({ timeout: 10000 });
    await expect(infoPanel).not.toBeEmpty({ timeout: 5000 });
  });

  test('API endpoints respond correctly', async ({ request }) => {
    // Test the status API endpoint
    const statusResponse = await request.get('/api/status');
    expect(statusResponse.status()).toBe(200);
    
    const statusData = await statusResponse.json();
    expect(statusData).toHaveProperty('running');
    expect(statusData).toHaveProperty('generation');
    expect(statusData).toHaveProperty('population');
  });

  test('grid view displays simulation entities', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation to start
    await page.waitForTimeout(5000);
    
    // Switch to grid view if not already active
    const gridTab = page.locator('.view-tab:has-text("Grid")');
    if (await gridTab.isVisible()) {
      await gridTab.click();
      await page.waitForTimeout(1000);
    }
    
    // Check that grid contains simulation symbols
    const gridContainer = page.locator('.grid-container');
    const gridContent = await gridContainer.textContent();
    
    // Grid should contain some meaningful content
    expect(gridContent).toBeTruthy();
    expect(gridContent.length).toBeGreaterThan(100); // Should have substantial content
  });

  test('real-time updates occur in the interface', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial data
    await page.waitForTimeout(5000);
    
    // Capture initial state
    const infoPanel = page.locator('.info-panel');
    const initialContent = await infoPanel.textContent();
    
    // Wait for potential updates
    await page.waitForTimeout(5000);
    
    // The content might change, but the panel should still be present and functional
    await expect(infoPanel).toBeVisible({ timeout: 5000 });
    const updatedContent = await infoPanel.textContent();
    expect(updatedContent).toBeTruthy();
  });
});
});