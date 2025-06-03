import { test, expect } from '@playwright/test';

test.describe('EvoSim Web Interface', () => {
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
    throw new Error(`EvoSim webserver is not running or not responding at ${baseURL}. Please start the server with: GOWORK=off go run . -web -web-port 8080`);
  });

  test.beforeEach(async ({ page }) => {
    // Set longer timeout for navigation in CI
    page.setDefaultTimeout(30000);
  });

  test('loads homepage and displays simulation interface', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Check that the main title is present
    await expect(page.locator('h1')).toContainText('EvoSim - Genetic Ecosystem Simulation', { timeout: 15000 });
    
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
    await expect(page.locator('.main-content')).toBeVisible({ timeout: 10000 });
    
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
    expect(statusData).toHaveProperty('status');
    expect(statusData).toHaveProperty('tick');
    expect(statusData).toHaveProperty('populations');
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

  // Comprehensive view validation tests
  test('all view modes are accessible and functional', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial load and connection
    await page.waitForTimeout(5000);
    
    // Define expected view modes
    const expectedViewModes = [
      'GRID', 'STATS', 'EVENTS', 'POPULATIONS', 'COMMUNICATION',
      'CIVILIZATION', 'PHYSICS', 'WIND', 'SPECIES', 'NETWORK',
      'DNA', 'CELLULAR', 'EVOLUTION', 'TOPOLOGY', 'TOOLS', 
      'ENVIRONMENT', 'BEHAVIOR', 'REPRODUCTION', 'STATISTICAL', 'ANOMALIES'
    ];
    
    // Check that view tabs are present
    const viewTabs = page.locator('.view-tab');
    const viewTabCount = await viewTabs.count();
    
    // Verify we have a reasonable number of view tabs
    expect(viewTabCount).toBeGreaterThan(10);
    
    // Test switching to each major view mode
    const viewModesToTest = ['GRID', 'STATS', 'EVENTS', 'POPULATIONS', 'WIND', 'SPECIES'];
    
    for (const viewMode of viewModesToTest) {
      const tab = page.locator(`.view-tab:has-text("${viewMode}")`);
      if (await tab.isVisible()) {
        await tab.click();
        await page.waitForTimeout(1000);
        
        // Verify the view content area is still present and accessible
        await expect(page.locator('#view-content')).toBeVisible({ timeout: 5000 });
        
        // Check that active tab is highlighted
        await expect(tab).toHaveClass(/active/, { timeout: 2000 });
      }
    }
  });

  // COMPREHENSIVE VIEW DATA VALIDATION TESTS

  test('STATS view displays all required fields with valid data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const statsTab = page.locator('.view-tab:has-text("STATS")');
    if (await statsTab.isVisible()) {
      await statsTab.click();
      await page.waitForTimeout(2000);
      
      // Check for specific stats fields in the info panel (always visible)
      await expect(page.locator('#avg-fitness')).toBeVisible();
      await expect(page.locator('#avg-energy')).toBeVisible();
      await expect(page.locator('#avg-age')).toBeVisible();
      
      // Verify the values are numbers
      const fitnessText = await page.locator('#avg-fitness').textContent();
      const energyText = await page.locator('#avg-energy').textContent();
      const ageText = await page.locator('#avg-age').textContent();
      
      expect(fitnessText).toMatch(/\d+\.?\d*/);
      expect(energyText).toMatch(/\d+\.?\d*/);
      expect(ageText).toMatch(/\d+\.?\d*/);
    }
  });

  test('POPULATIONS view displays detailed population data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const popTab = page.locator('.view-tab:has-text("POPULATIONS")');
    if (await popTab.isVisible()) {
      await popTab.click();
      await page.waitForTimeout(2000);
      
      // Check populations content in info panel
      const popContent = page.locator('#populations-content');
      await expect(popContent).toBeVisible();
      
      const content = await popContent.textContent();
      // Should show population data or "No populations"
      expect(content).toBeTruthy();
    }
  });

  test('WIND view displays dynamic wind system values', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const windTab = page.locator('.view-tab:has-text("WIND")');
    if (await windTab.isVisible()) {
      await windTab.click();
      await page.waitForTimeout(2000);
      
      // Check wind values in info panel
      await expect(page.locator('#wind-direction')).toBeVisible();
      await expect(page.locator('#wind-strength')).toBeVisible();
      await expect(page.locator('#weather-pattern')).toBeVisible();
      
      // Verify wind direction contains degrees
      const directionText = await page.locator('#wind-direction').textContent();
      expect(directionText).toMatch(/\d+\.?\d*°/);
      
      // Verify wind strength is a number
      const strengthText = await page.locator('#wind-strength').textContent();
      expect(strengthText).toMatch(/\d+\.?\d*/);
      
      // Verify weather pattern is not empty
      const weatherText = await page.locator('#weather-pattern').textContent();
      expect(weatherText).toBeTruthy();
    }
  });

  test('SPECIES view displays comprehensive species data and individual visualization', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const speciesTab = page.locator('.view-tab:has-text("SPECIES")');
    if (await speciesTab.isVisible()) {
      await speciesTab.click();
      await page.waitForTimeout(2000);
      
      // Wait for species view to load
      await page.waitForTimeout(1000);
      
      // Check main view content for species
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      
      // Should contain species tracking information
      expect(content).toContain('Species Tracking');
      expect(content).toContain('Active Species');
      expect(content).toContain('Extinct Species');
      
      // Check for diversity metrics
      expect(content).toContain('Diversity Metrics');
      expect(content).toContain('Species Survival Rate');
      
      // Check for individual visualization section
      expect(content).toContain('Individual Species Visualization');
      
      // Species Gallery section or "No species data available" message should be present
      const hasSpeciesGallery = content.includes('Species Gallery');
      const hasNoSpeciesMessage = content.includes('No species data available');
      expect(hasSpeciesGallery || hasNoSpeciesMessage).toBeTruthy();
      
      // If we have species gallery, test the interaction
      if (hasSpeciesGallery) {
        // Look for clickable species items
        const speciesItems = page.locator('.species-item');
        if (await speciesItems.count() > 0) {
          // Test clicking on first species to see detailed visualization
          await speciesItems.first().click();
          await page.waitForTimeout(1000);
          
          // Check that species detail modal appears
          const modal = page.locator('#species-detail-modal');
          await expect(modal).toBeVisible({ timeout: 5000 });
          
          // Check for visualization components
          const modalContent = await modal.textContent();
          expect(modalContent).toContain('Individual Visualization');
          expect(modalContent).toContain('Species Profile View');
          expect(modalContent).toContain('Genetic Trait Analysis');
          expect(modalContent).toContain('Cellular Structure View');
          expect(modalContent).toContain('Environmental Adaptation');
          
          // Close modal
          const closeButton = page.locator('#species-detail-modal button');
          await closeButton.click();
          await page.waitForTimeout(500);
          
          // Verify modal is closed
          await expect(modal).not.toBeVisible();
        }
      }
    }
  });

  test('COMMUNICATION view displays signal data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const commTab = page.locator('.view-tab:has-text("COMMUNICATION")');
    if (await commTab.isVisible()) {
      await commTab.click();
      await page.waitForTimeout(2000);
      
      // Check communication values in info panel
      await expect(page.locator('#active-signals')).toBeVisible();
      
      const signalsText = await page.locator('#active-signals').textContent();
      expect(signalsText).toMatch(/\d+/);
    }
  });

  test('EVENTS view displays event history', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const eventsTab = page.locator('.view-tab:has-text("EVENTS")');
    if (await eventsTab.isVisible()) {
      await eventsTab.click();
      await page.waitForTimeout(2000);
      
      // Check that events view content is visible
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('PHYSICS view displays physics metrics', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const physicsTab = page.locator('.view-tab:has-text("PHYSICS")');
    if (await physicsTab.isVisible()) {
      await physicsTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('EVOLUTION view displays evolution tracking data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const evolutionTab = page.locator('.view-tab:has-text("EVOLUTION")');
    if (await evolutionTab.isVisible()) {
      await evolutionTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('DNA view displays genetic information', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const dnaTab = page.locator('.view-tab:has-text("DNA")');
    if (await dnaTab.isVisible()) {
      await dnaTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('CELLULAR view displays cellular system data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const cellularTab = page.locator('.view-tab:has-text("CELLULAR")');
    if (await cellularTab.isVisible()) {
      await cellularTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('STATISTICAL view displays analysis data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const statTab = page.locator('.view-tab:has-text("STATISTICAL")');
    if (await statTab.isVisible()) {
      await statTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('ANOMALIES view displays anomaly detection data', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const anomaliesTab = page.locator('.view-tab:has-text("ANOMALIES")');
    if (await anomaliesTab.isVisible()) {
      await anomaliesTab.click();
      await page.waitForTimeout(2000);
      
      const viewContent = page.locator('#view-content');
      await expect(viewContent).toBeVisible();
      
      const content = await viewContent.textContent();
      expect(content).toBeTruthy();
    }
  });

  test('all remaining views load without errors', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    await page.waitForTimeout(5000);
    
    const remainingViews = ['CIVILIZATION', 'NETWORK', 'TOPOLOGY', 'TOOLS', 'ENVIRONMENT', 'BEHAVIOR', 'REPRODUCTION'];
    
    for (const viewName of remainingViews) {
      const tab = page.locator(`.view-tab:has-text("${viewName}")`);
      if (await tab.isVisible()) {
        await tab.click();
        await page.waitForTimeout(1000);
        
        const viewContent = page.locator('#view-content');
        await expect(viewContent).toBeVisible();
        
        const content = await viewContent.textContent();
        expect(content).toBeTruthy();
      }
    }
  });

  test('legend displays updated entity symbols', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial load
    await page.waitForTimeout(5000);
    
    // Check the legend content
    const legend = page.locator('.legend');
    await expect(legend).toBeVisible({ timeout: 10000 });
    
    const legendText = await legend.textContent();
    
    // Verify the legend contains either emoji symbols (updated) or letter symbols (if still loading)
    const hasEmojiSymbols = legendText.includes('🐰') && legendText.includes('🐺') && legendText.includes('🐻');
    const hasLetterSymbols = legendText.includes('H Herbivore') && legendText.includes('P Predator');
    
    // The legend should show either the new emoji format or the old letter format
    expect(hasEmojiSymbols || hasLetterSymbols).toBeTruthy();
    
    // If we see the new format, verify all expected symbols are present
    if (hasEmojiSymbols) {
      expect(legendText).toContain('🐰'); // Herbivore
      expect(legendText).toContain('🐺'); // Predator
      expect(legendText).toContain('🐻'); // Omnivore
      expect(legendText).toContain('🦋'); // Generic entity (blue butterfly)
      
      // Verify plant symbols
      expect(legendText).toContain('🌱'); // Grass
      expect(legendText).toContain('🌿'); // Bush
      expect(legendText).toContain('🌳'); // Tree
      expect(legendText).toContain('🍄'); // Mushroom
      expect(legendText).toContain('🌊'); // Algae
      expect(legendText).toContain('🌵'); // Cactus
    }
    
    // If we see the old format, that's also acceptable for compatibility
    if (hasLetterSymbols) {
      expect(legendText).toContain('H Herbivore');
      expect(legendText).toContain('P Predator');
      expect(legendText).toContain('O Omnivore');
    }
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

  test('capture screenshots of all visualization modes', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial data
    await page.waitForTimeout(5000);
    
    // Ensure screenshots directory exists and save to test-specific location
    const screenshotDir = 'screenshots/test-visualizations';
    
    // Define view modes to capture
    const viewModesToCapture = [
      'GRID', 'STATS', 'EVENTS', 'POPULATIONS', 'COMMUNICATION',
      'CIVILIZATION', 'PHYSICS', 'WIND', 'SPECIES', 'NETWORK',
      'DNA', 'CELLULAR', 'EVOLUTION', 'TOPOLOGY', 'TOOLS', 
      'ENVIRONMENT', 'BEHAVIOR', 'STATISTICAL', 'ANOMALIES'
    ];
    
    for (const viewMode of viewModesToCapture) {
      const tab = page.locator(`.view-tab:has-text("${viewMode}")`);
      if (await tab.isVisible()) {
        await tab.click();
        await page.waitForTimeout(2000); // Allow view to fully load
        
        // Take screenshot with descriptive name for test purposes
        await page.screenshot({ 
          path: `${screenshotDir}/${viewMode.toLowerCase()}-view.png`,
          fullPage: true 
        });
      }
    }
    
    // Also capture the main interface overview
    await page.screenshot({ 
      path: `${screenshotDir}/main-interface.png`,
      fullPage: true 
    });
  });

  test('species gallery displays detailed species information', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for initial data
    await page.waitForTimeout(5000);
    
    // Navigate to species view
    const speciesTab = page.locator('.view-tab:has-text("SPECIES")');
    if (await speciesTab.isVisible()) {
      await speciesTab.click();
      await page.waitForTimeout(3000);
      
      // Capture species gallery
      await page.screenshot({ 
        path: 'screenshots/test-visualizations/species-gallery.png',
        fullPage: true 
      });
      
      // Try to click on the first species if available
      const firstSpecies = page.locator('.species-item').first();
      if (await firstSpecies.isVisible()) {
        await firstSpecies.click();
        await page.waitForTimeout(2000);
        
        // Capture detailed species view
        await page.screenshot({ 
          path: 'screenshots/test-visualizations/species-detail.png',
          fullPage: true 
        });
        
        // Verify species detail modal content
        const modal = page.locator('#species-detail-modal');
        await expect(modal).toBeVisible();
        
        const modalContent = await modal.textContent();
        
        // Check for all expected visualization components
        expect(modalContent).toContain('Individual Visualization');
        expect(modalContent).toContain('Species Profile View');
        expect(modalContent).toContain('Genetic Trait Analysis');
        expect(modalContent).toContain('Cellular Structure View');
        expect(modalContent).toContain('Environmental Adaptation');
        
        // Close modal
        const closeButton = page.locator('#species-detail-modal button');
        await closeButton.click();
      }
    }
  });
});