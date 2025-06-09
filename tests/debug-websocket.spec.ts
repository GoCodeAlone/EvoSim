import { test, expect } from '@playwright/test';

test('Test isometric view WebSocket connectivity', async ({ page }) => {
  // Enable console logging
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));
  
  // Navigate to isometric view
  await page.goto('http://localhost:8080/iso', { 
    waitUntil: 'networkidle', 
    timeout: 30000 
  });
  
  // Wait for the page to load and attempt WebSocket connection
  await page.waitForTimeout(5000);
  
  // Check if loading screen is still visible
  const loadingElement = await page.locator('#loading');
  const isLoadingVisible = await loadingElement.isVisible();
  
  console.log('Loading screen visible:', isLoadingVisible);
  
  // Take screenshot for debugging
  await page.screenshot({ 
    path: 'screenshots/isometric-debug-websocket.png',
    fullPage: true 
  });
  
  // Get some debugging info from the page
  const debugInfo = await page.evaluate(() => {
    return {
      websocketState: window.gameState?.websocket?.readyState,
      hasIsometricData: !!window.gameState?.isometricData,
      cameraPosition: window.gameState?.camera
    };
  });
  
  console.log('Debug info:', debugInfo);
});