const { test, expect } = require('@playwright/test');

test('Enhanced 2.5D Isometric View - Visual Documentation', async ({ page }) => {
  console.log('Documenting enhanced isometric view...');
  
  // Navigate to the enhanced isometric view
  await page.goto('http://localhost:8082/iso');
  
  // Wait for the canvas to load
  await page.waitForSelector('#gameCanvas', { timeout: 10000 });
  
  // Take initial screenshot
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-initial-load.png' });
  console.log('Screenshot saved: enhanced-initial-load.png');
  
  // Wait more time for data to potentially load
  await page.waitForTimeout(8000);
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-after-waiting.png' });
  console.log('Screenshot saved: enhanced-after-waiting.png');
  
  // Test zoom controls
  await page.mouse.wheel(0, -300); // Zoom out
  await page.waitForTimeout(1000);
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-zoomed-out.png' });
  console.log('Screenshot saved: enhanced-zoomed-out.png');
  
  // Test camera movement
  for (let i = 0; i < 5; i++) {
    await page.keyboard.press('KeyW');
    await page.waitForTimeout(50);
  }
  await page.waitForTimeout(500);
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-camera-moved.png' });
  console.log('Screenshot saved: enhanced-camera-moved.png');
  
  // Zoom back in
  await page.mouse.wheel(0, 400); // Zoom in
  await page.waitForTimeout(1000);
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-zoomed-in.png' });
  console.log('Screenshot saved: enhanced-zoomed-in.png');
  
  // Test UI elements exist
  const canvas = await page.locator('#gameCanvas');
  expect(await canvas.isVisible()).toBeTruthy();
  
  const terrainInfo = await page.locator('#terrainInfo');
  expect(await terrainInfo.isVisible()).toBeTruthy();
  
  const geoEvents = await page.locator('#geoEvents');
  expect(await geoEvents.isVisible()).toBeTruthy();
  
  // Final screenshot with UI
  await page.screenshot({ path: 'screenshots/isometric-views/enhanced-final-ui.png' });
  console.log('Screenshot saved: enhanced-final-ui.png');
  
  console.log('Enhanced isometric view documentation completed!');
});