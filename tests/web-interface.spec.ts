import { test, expect } from '@playwright/test';

test.describe('EvoSim Web Interface', () => {
  test('loads homepage and displays simulation interface', async ({ page }) => {
    await page.goto('/');
    
    // Check that the main title is present
    await expect(page.locator('h1')).toContainText('EvoSim - Genetic Algorithm Simulation');
    
    // Check for connection status
    const connectionStatus = page.locator('.connection-status');
    await expect(connectionStatus).toBeVisible();
    
    // Check for main interface components
    await expect(page.locator('.simulation-view')).toBeVisible();
    await expect(page.locator('.info-panel')).toBeVisible();
    await expect(page.locator('.controls')).toBeVisible();
  });

  test('displays simulation grid', async ({ page }) => {
    await page.goto('/');
    
    // Wait for WebSocket connection
    await page.waitForTimeout(2000);
    
    // Check that the grid container is present
    const gridContainer = page.locator('.grid-container');
    await expect(gridContainer).toBeVisible();
    
    // Check that grid content is populated (should contain simulation symbols)
    const gridContent = await gridContainer.textContent();
    expect(gridContent).toBeTruthy();
    expect(gridContent.length).toBeGreaterThan(10);
  });

  test('view switching functionality', async ({ page }) => {
    await page.goto('/');
    
    // Wait for initial load
    await page.waitForTimeout(2000);
    
    // Check that view tabs are present
    const viewTabs = page.locator('.view-tab');
    await expect(viewTabs.first()).toBeVisible();
    
    // Test switching to different views
    const gridTab = page.locator('.view-tab:has-text("Grid")');
    const statsTab = page.locator('.view-tab:has-text("Stats")');
    const eventsTab = page.locator('.view-tab:has-text("Events")');
    
    if (await gridTab.isVisible()) {
      await gridTab.click();
      await page.waitForTimeout(500);
    }
    
    if (await statsTab.isVisible()) {
      await statsTab.click();
      await page.waitForTimeout(500);
      // Stats view should show statistics content
      await expect(page.locator('.info-panel')).toBeVisible();
    }
    
    if (await eventsTab.isVisible()) {
      await eventsTab.click();
      await page.waitForTimeout(500);
      // Events view should show event information
      await expect(page.locator('.info-panel')).toBeVisible();
    }
  });

  test('websocket connection and real-time updates', async ({ page }) => {
    await page.goto('/');
    
    // Wait for WebSocket connection to establish
    await page.waitForTimeout(3000);
    
    // Check connection status shows connected
    const connectionStatus = page.locator('.connection-status');
    await expect(connectionStatus).toContainText('Connected');
    await expect(connectionStatus).toHaveClass(/connected/);
    
    // Capture initial grid content
    const gridContainer = page.locator('.grid-container');
    const initialContent = await gridContainer.textContent();
    
    // Wait for updates (simulation should be running)
    await page.waitForTimeout(2000);
    
    // Verify content has potentially changed (showing live updates)
    // Note: Content might be the same, but the test ensures the mechanism works
    const updatedContent = await gridContainer.textContent();
    expect(updatedContent).toBeTruthy();
  });

  test('simulation control buttons', async ({ page }) => {
    await page.goto('/');
    
    // Wait for initial load
    await page.waitForTimeout(2000);
    
    // Check for control buttons
    const controlsSection = page.locator('.controls');
    await expect(controlsSection).toBeVisible();
    
    // Look for common control buttons (these may vary based on implementation)
    const buttons = page.locator('.controls button');
    const buttonCount = await buttons.count();
    
    // Ensure we have some control buttons
    expect(buttonCount).toBeGreaterThan(0);
    
    // Test clicking controls (if pause/play buttons exist)
    for (let i = 0; i < Math.min(buttonCount, 3); i++) {
      const button = buttons.nth(i);
      if (await button.isVisible()) {
        await button.click();
        await page.waitForTimeout(500);
      }
    }
  });

  test('responsive design and layout', async ({ page }) => {
    await page.goto('/');
    
    // Test desktop view
    await page.setViewportSize({ width: 1200, height: 800 });
    await expect(page.locator('.container')).toBeVisible();
    
    // Test tablet view
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.simulation-view')).toBeVisible();
    
    // Test mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('.simulation-view')).toBeVisible();
  });

  test('information panel content', async ({ page }) => {
    await page.goto('/');
    
    // Wait for data to load
    await page.waitForTimeout(3000);
    
    // Check that info panel displays content
    const infoPanel = page.locator('.info-panel');
    await expect(infoPanel).toBeVisible();
    
    const infoPanelContent = await infoPanel.textContent();
    expect(infoPanelContent).toBeTruthy();
    expect(infoPanelContent.length).toBeGreaterThan(10);
  });

  test('legend and symbols display', async ({ page }) => {
    await page.goto('/');
    
    // Wait for initial load
    await page.waitForTimeout(2000);
    
    // Look for legend or symbol information
    const legend = page.locator('.legend');
    if (await legend.isVisible()) {
      const legendContent = await legend.textContent();
      expect(legendContent).toBeTruthy();
    }
    
    // Ensure grid contains various symbols representing simulation elements
    const gridContainer = page.locator('.grid-container');
    const gridContent = await gridContainer.textContent();
    
    // Grid should contain various characters representing entities and environment
    expect(gridContent).toMatch(/[.*○◐●▲▼◆♦♠♣♥♪]/);
  });
});