import { test, expect } from '@playwright/test';

test('quick isometric view screenshot', async ({ page }) => {
  // Navigate to isometric view
  await page.goto('http://localhost:8080/iso', { 
    waitUntil: 'networkidle', 
    timeout: 30000 
  });
  
  // Wait for page to load
  await page.waitForTimeout(3000);
  
  // Take screenshot
  await page.screenshot({ 
    path: 'screenshots/isometric-debug.png',
    fullPage: true 
  });
  
  console.log('Screenshot saved to screenshots/isometric-debug.png');
});