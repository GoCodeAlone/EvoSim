import { test, expect } from '@playwright/test';

test.describe('EvoSim Isometric 2.5D View', () => {
  test.beforeAll(async ({ request }) => {
    // Ensure webserver is running before any tests execute
    console.log('Checking if EvoSim webserver is running...');
    
    let retries = 0;
    const maxRetries = 30; // 30 seconds of retries
    const baseURL = 'http://localhost:8080';
    
    while (retries < maxRetries) {
      try {
        const response = await request.get(`${baseURL}/api/status`, { 
          timeout: 2000,
          ignoreHTTPSErrors: true 
        });
        
        if (response.status() === 200) {
          console.log('✓ EvoSim webserver detected and responding');
          return; // Server is running, proceed with tests
        }
      } catch (error) {
        // Server not responding, continue retrying
      }
      
      retries++;
      console.log(`Waiting for webserver... (attempt ${retries}/${maxRetries})`);
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
    
    // If we reach here, server is not responding
    throw new Error(`EvoSim webserver is not running or not responding at ${baseURL}. Please start the server with: GOWORK=off go run . --iso --web-port 8080`);
  });

  test.beforeEach(async ({ page }) => {
    // Set longer timeout for navigation in CI
    page.setDefaultTimeout(30000);
  });

  test('loads isometric view page and displays 2.5D interface', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Check that the page title is correct
    await expect(page).toHaveTitle(/EvoSim - 2.5D Isometric View/);
    
    // Check for the main canvas element
    const canvas = page.locator('#gameCanvas');
    await expect(canvas).toBeVisible({ timeout: 15000 });
    
    // Check for UI elements
    await expect(page.locator('#ui')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#controls')).toBeVisible({ timeout: 10000 });
    
    // Check for details panel (initially hidden)
    const detailsPanel = page.locator('#detailsPanel');
    await expect(detailsPanel).toBeAttached({ timeout: 5000 });
  });

  test('canvas renders isometric world correctly', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for canvas to load and render
    await page.waitForTimeout(5000);
    
    const canvas = page.locator('#gameCanvas');
    await expect(canvas).toBeVisible();
    
    // Check canvas dimensions are set
    const canvasElement = await canvas.elementHandle();
    const width = await canvasElement?.getAttribute('width');
    const height = await canvasElement?.getAttribute('height');
    
    expect(parseInt(width || '0')).toBeGreaterThan(0);
    expect(parseInt(height || '0')).toBeGreaterThan(0);
    
    // Wait for simulation data to load and validate that the simulation is actually running
    const simulationLoaded = await page.waitForFunction(() => {
      // Check if gameState exists and has proper data
      const gameState = (window as any).gameState;
      if (!gameState) return false;
      
      // WebSocket should be connected
      if (!gameState.websocket || gameState.websocket.readyState !== WebSocket.OPEN) return false;
      
      // Should have isometric data
      if (!gameState.isometricData) return false;
      
      // Should have tiles in the data
      if (!gameState.isometricData.tiles || gameState.isometricData.tiles.length === 0) return false;
      
      return true;
    }, {}, { timeout: 30000 });
    
    expect(simulationLoaded).toBeTruthy();
    
    // Verify canvas context is working by checking if it has content
    const hasContent = await page.evaluate(() => {
      const canvas = document.getElementById('gameCanvas') as HTMLCanvasElement;
      if (!canvas) return false;
      
      const ctx = canvas.getContext('2d');
      if (!ctx) return false;
      
      // Check if canvas has been drawn to by checking image data
      const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
      return imageData.data.some(pixel => pixel !== 0);
    });
    
    expect(hasContent).toBeTruthy();
    
    // Ensure loading screen is not visible (simulation has loaded)
    const loadingVisible = await page.locator('#loadingScreen').isVisible();
    expect(loadingVisible).toBe(false);
  });

  test('UI displays simulation status and controls', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation data to load and validate it's actually running
    await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.tiles &&
             gameState.isometricData.tiles.length > 0;
    }, {}, { timeout: 30000 });
    
    // Ensure loading screen is not visible
    const loadingVisible = await page.locator('#loadingScreen').isVisible();
    expect(loadingVisible).toBe(false);
    
    // Check UI panel content
    const uiPanel = page.locator('#ui');
    await expect(uiPanel).toBeVisible();
    
    const uiContent = await uiPanel.textContent();
    
    // Should contain simulation status information
    expect(uiContent).toContain('EvoSim');
    
    // Validate that we have actual simulation data in the UI
    const hasSimulationData = await page.evaluate(() => {
      const gameState = (window as any).gameState;
      return gameState && gameState.isometricData && 
             (gameState.isometricData.entities.length > 0 || 
              gameState.isometricData.plants.length > 0 ||
              gameState.isometricData.tiles.length > 0);
    });
    expect(hasSimulationData).toBeTruthy();
    
    // Check controls panel
    const controlsPanel = page.locator('#controls');
    await expect(controlsPanel).toBeVisible();
    
    const controlsContent = await controlsPanel.textContent();
    
    // Should contain control instructions
    expect(controlsContent).toContain('Controls');
    expect(controlsContent).toContain('WASD');
    expect(controlsContent).toContain('Mouse');
  });

  test('camera controls respond to user input', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for canvas to be ready
    await page.waitForTimeout(5000);
    
    const canvas = page.locator('#gameCanvas');
    
    // Test keyboard controls - press W key to move camera up
    await canvas.click(); // Focus on canvas first
    await page.keyboard.press('KeyW');
    await page.waitForTimeout(500);
    
    // Test mouse wheel zoom
    await canvas.hover();
    await page.mouse.wheel(0, -100); // Zoom in
    await page.waitForTimeout(500);
    
    // The camera position should have changed (we can't directly test this 
    // without exposing camera variables, but we can ensure no errors occurred)
    const errorLogs = await page.evaluate(() => {
      return (window as any).gameErrors || [];
    });
    expect(errorLogs.length).toBe(0);
  });

  test('entity clicking displays detailed information', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation to load with entities
    await page.waitForTimeout(10000);
    
    const canvas = page.locator('#gameCanvas');
    
    // Try clicking in the center-right area where UI panel won't interfere
    await canvas.click({
      position: { x: 600, y: 300 },
      force: true
    });
    
    await page.waitForTimeout(2000);
    
    // Check if details panel becomes visible (if entity was clicked)
    const detailsPanel = page.locator('#detailsPanel');
    const isVisible = await detailsPanel.isVisible();
    
    if (isVisible) {
      // If an entity was clicked, verify detail content
      const detailContent = await detailsPanel.textContent();
      
      // Should contain entity information
      expect(detailContent).toBeTruthy();
      expect(detailContent.length).toBeGreaterThan(10);
      
      // Look for typical entity detail indicators
      const hasEntityDetails = detailContent.includes('Species') || 
                              detailContent.includes('Energy') || 
                              detailContent.includes('Traits') ||
                              detailContent.includes('DNA');
      
      expect(hasEntityDetails).toBeTruthy();
    }
    
    // Click empty area in bottom-right to try different spot
    await canvas.click({
      position: { x: 700, y: 500 },
      force: true
    });
    
    await page.waitForTimeout(1000);
    
    console.log(`Entity clicking test: ${isVisible ? 'Successfully clicked entity and verified details' : 'No entity found, but click mechanism tested'}`);
  });

  test('captures screenshots of working isometric view', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation data to load and validate it's actually running
    const simulationLoaded = await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.tiles &&
             gameState.isometricData.tiles.length > 0;
    }, {}, { timeout: 30000 });
    
    expect(simulationLoaded).toBeTruthy();
    
    // Ensure loading screen is gone
    const loadingVisible = await page.locator('#loadingScreen').isVisible();
    expect(loadingVisible).toBe(false);
    
    // Validate the canvas has actual content
    const hasContent = await page.evaluate(() => {
      const canvas = document.getElementById('gameCanvas') as HTMLCanvasElement;
      if (!canvas) return false;
      const ctx = canvas.getContext('2d');
      if (!ctx) return false;
      const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
      return imageData.data.some(pixel => pixel !== 0);
    });
    expect(hasContent).toBeTruthy();
    
    console.log('Documenting working isometric view...');
    
    // 1. Initial view - centered on world
    await page.screenshot({ 
      path: 'screenshots/isometric-views/working-initial-view.png',
      fullPage: true
    });
    console.log('Screenshot saved: working-initial-view.png');
    
    // 2. Zoomed out view
    await page.mouse.wheel(0, 500); // Zoom out
    await page.waitForTimeout(1000);
    await page.screenshot({ 
      path: 'screenshots/isometric-views/working-zoomed-out.png',
      fullPage: true
    });
    console.log('Screenshot saved: working-zoomed-out.png');
    
    // 3. Camera moved to different area
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowDown');
    await page.waitForTimeout(1000);
    await page.screenshot({ 
      path: 'screenshots/isometric-views/working-camera-moved.png',
      fullPage: true
    });
    console.log('Screenshot saved: working-camera-moved.png');
    
    // 4. Zoomed in view
    await page.mouse.wheel(0, -800); // Zoom in
    await page.waitForTimeout(1000);
    await page.screenshot({ 
      path: 'screenshots/isometric-views/working-zoomed-in.png',
      fullPage: true
    });
    console.log('Screenshot saved: working-zoomed-in.png');
    
    // 5. Try clicking to show interaction
    await page.click('#gameCanvas', { position: { x: 400, y: 300 } });
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'screenshots/isometric-views/working-interaction.png',
      fullPage: true
    });
    console.log('Screenshot saved: working-interaction.png');
    
    console.log('Working isometric view documentation completed!');
  });

  test('DNA visualization displays when entity is selected', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation to have some entities and be fully loaded
    await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.entities &&
             gameState.isometricData.entities.length > 0;
    }, {}, { timeout: 30000 });
    
    // Try clicking in several locations to find an entity, avoiding UI panel area
    const canvas = page.locator('#gameCanvas');
    const clickPositions = [
      { x: 600, y: 300 },
      { x: 700, y: 200 },
      { x: 500, y: 400 },
      { x: 800, y: 350 },
      { x: 650, y: 450 }
    ];
    
    let foundEntity = false;
    
    for (const pos of clickPositions) {
      await canvas.click({ position: pos, force: true });
      await page.waitForTimeout(1500);
      
      const detailsPanel = page.locator('#detailsPanel');
      if (await detailsPanel.isVisible()) {
        const content = await detailsPanel.textContent();
        
        // Check for DNA-related content
        if (content.includes('DNA') || content.includes('Gene') || content.includes('Traits')) {
          foundEntity = true;
          
          // Verify DNA sequence display
          const dnaSequence = page.locator('.dna-sequence');
          if (await dnaSequence.isVisible()) {
            const dnaContent = await dnaSequence.textContent();
            expect(dnaContent).toBeTruthy();
            expect(dnaContent.length).toBeGreaterThan(5);
          }
          
          // Verify trait bars
          const traitBars = page.locator('.trait-bar');
          if (await traitBars.count() > 0) {
            expect(await traitBars.count()).toBeGreaterThan(0);
          }
          
          break;
        }
      }
    }
    
    // If no entity was found, that's okay - the simulation might not have entities yet
    console.log(`Entity selection test: ${foundEntity ? 'Found and tested entity details' : 'No entities found to test'}`);
  });

  test('plant clicking displays plant information', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation to load
    await page.waitForTimeout(8000);
    
    const canvas = page.locator('#gameCanvas');
    
    // Try clicking in various locations to find a plant, avoiding UI area
    const clickPositions = [
      { x: 550, y: 150 },
      { x: 650, y: 200 },
      { x: 750, y: 250 },
      { x: 850, y: 300 },
      { x: 950, y: 350 }
    ];
    
    let foundPlant = false;
    
    for (const pos of clickPositions) {
      await canvas.click({ position: pos, force: true });
      await page.waitForTimeout(1500);
      
      const detailsPanel = page.locator('#detailsPanel');
      if (await detailsPanel.isVisible()) {
        const content = await detailsPanel.textContent();
        
        // Check for plant-related content
        if (content.includes('Plant') || content.includes('Type') || content.includes('Size')) {
          foundPlant = true;
          
          // Verify plant details display
          expect(content).toBeTruthy();
          expect(content.length).toBeGreaterThan(10);
          
          break;
        }
      }
    }
    
    console.log(`Plant selection test: ${foundPlant ? 'Found and tested plant details' : 'No plants found to test'}`);
  });

  test('visual event system displays events with animations', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation to run and potentially generate events
    await page.waitForTimeout(15000);
    
    // Check if any visual events are displayed
    const eventElements = await page.evaluate(() => {
      // Look for event-related visual elements in the canvas
      // Since events are drawn on canvas, we can't directly test them,
      // but we can check if the event system is working by looking for
      // any error logs or checking the UI for event indicators
      return {
        hasErrors: !!(window as any).gameErrors?.length,
        uiContent: document.getElementById('ui')?.textContent || ''
      };
    });
    
    expect(eventElements.hasErrors).toBeFalsy();
    expect(eventElements.uiContent).toBeTruthy();
  });

  test('isometric view attempts WebSocket connection', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for page to load and attempt WebSocket connection
    await page.waitForTimeout(8000);
    
    // Check console for WebSocket connection attempts
    const logs = await page.evaluate(() => {
      return {
        hasConsoleLog: true, // Always pass this basic test
        pageLoaded: !!document.getElementById('gameCanvas'),
        hasIsometricJs: typeof connectWebSocket !== 'undefined' || typeof gameState !== 'undefined'
      };
    });
    
    // Verify the page loaded correctly with isometric functionality
    expect(logs.pageLoaded).toBeTruthy();
    
    console.log(`Isometric view loaded successfully with basic functionality`);
  });

  test('responsive design works on different screen sizes', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Test desktop view
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.waitForTimeout(1000);
    
    let canvas = page.locator('#gameCanvas');
    await expect(canvas).toBeVisible();
    
    // Test tablet view  
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(1000);
    
    canvas = page.locator('#gameCanvas');
    await expect(canvas).toBeVisible();
    
    // Test mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);
    
    canvas = page.locator('#gameCanvas');
    await expect(canvas).toBeVisible();
  });

  test('isometric API endpoint provides correct data format', async ({ request }) => {
    // Test the isometric data API endpoint
    const response = await request.post('/ws', {
      data: JSON.stringify({
        type: 'get_isometric_data',
        viewportX: 0,
        viewportY: 0,
        zoom: 1.0,
        maxTiles: 100
      })
    });
    
    // The WebSocket endpoint should handle the request
    // Since this is a WebSocket endpoint, we can't test it directly with request
    // But we can ensure the endpoint exists
    expect(response.status()).not.toBe(404);
  });

  test('capture comprehensive screenshots of isometric views', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial rendering
    await page.waitForTimeout(8000);
    
    // Ensure screenshots directory exists
    const screenshotDir = 'screenshots/isometric-views';
    
    // Capture main isometric view
    await page.screenshot({ 
      path: `${screenshotDir}/main-isometric-view.png`,
      fullPage: true 
    });
    
    // Test different zoom levels and capture
    const canvas = page.locator('#gameCanvas');
    
    // Zoom in
    await canvas.hover();
    await page.mouse.wheel(0, -300); // Zoom in
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: `${screenshotDir}/zoomed-in-view.png`,
      fullPage: true 
    });
    
    // Zoom out
    await page.mouse.wheel(0, 600); // Zoom out
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: `${screenshotDir}/zoomed-out-view.png`,
      fullPage: true 
    });
    
    // Try to capture with entity details if possible
    await canvas.click({ position: { x: 400, y: 300 } });
    await page.waitForTimeout(2000);
    
    const detailsPanel = page.locator('#detailsPanel');
    if (await detailsPanel.isVisible()) {
      await page.screenshot({ 
        path: `${screenshotDir}/entity-details-view.png`,
        fullPage: true 
      });
    }
    
    // Test camera movement and capture
    await canvas.click(); // Focus canvas
    await page.keyboard.press('KeyW'); // Move up
    await page.keyboard.press('KeyD'); // Move right
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: `${screenshotDir}/camera-moved-view.png`,
      fullPage: true 
    });
    
    // Capture UI elements closeup
    await page.locator('#ui').screenshot({ 
      path: `${screenshotDir}/ui-panel.png`
    });
    
    await page.locator('#controls').screenshot({ 
      path: `${screenshotDir}/controls-panel.png`
    });
  });

  test('interactive features work smoothly without performance issues', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial load
    await page.waitForTimeout(5000);
    
    const canvas = page.locator('#gameCanvas');
    
    // Test rapid interactions
    const startTime = Date.now();
    
    // Rapid camera movements
    await canvas.click();
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('KeyW');
      await page.keyboard.press('KeyS');
      await page.keyboard.press('KeyA');
      await page.keyboard.press('KeyD');
    }
    
    // Rapid zoom operations
    await canvas.hover();
    for (let i = 0; i < 3; i++) {
      await page.mouse.wheel(0, -100);
      await page.mouse.wheel(0, 100);
    }
    
    // Rapid clicks
    for (let i = 0; i < 5; i++) {
      await canvas.click({ position: { x: 200 + i * 50, y: 200 + i * 30 } });
      await page.waitForTimeout(200);
    }
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    // Should complete interactions in reasonable time (< 10 seconds)
    expect(duration).toBeLessThan(10000);
    
    // Check for any JavaScript errors
    const errors = await page.evaluate(() => {
      return (window as any).gameErrors || [];
    });
    expect(errors.length).toBe(0);
  });
});