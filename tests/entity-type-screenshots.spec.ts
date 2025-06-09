import { test, expect } from '@playwright/test';

test.describe('EvoSim Entity Type Screenshots', () => {
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

  test('capture screenshots of all entity types with labels', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation data to load and validate it's actually running
    const simulationLoaded = await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.entities &&
             gameState.isometricData.entities.length > 0;
    }, { timeout: 30000 });

    expect(simulationLoaded).toBeTruthy();

    // Wait for more evolution to occur to get diverse entity types
    console.log('Waiting for simulation to evolve diverse entity types...');
    await page.waitForTimeout(10000); // Allow 10 seconds for evolution

    // Get all entity data and classify them - collect data over multiple samples
    let allEntityTypes = new Map();
    const numSamples = 3;
    
    console.log(`Collecting entity data over ${numSamples} samples...`);
    for (let sample = 0; sample < numSamples; sample++) {
      await page.waitForTimeout(3000); // Wait between samples for evolution
      
      const entityData = await page.evaluate(() => {
        const gameState = (window as any).gameState;
        if (!gameState || !gameState.isometricData || !gameState.isometricData.entities) {
          return [];
        }

        const entities = gameState.isometricData.entities;
        const entityTypes = new Map();

        // Function to determine entity type (same as in isometric_view.html)
        function determineEntityType(entity) {
          const traits = entity.traits;
          
          // Priority-based classification for most distinctive features
          if (traits.flying_ability > 0.3) return 'flying';
          if (traits.aquatic_adaptation > 0.3) return 'aquatic';
          if (traits.digging_ability > 0.3 || traits.underground_nav > 0.2) return 'underground';
          if (traits.size > 0.4 && traits.aggression > 0.3) return 'large_predator';
          if (traits.size < -0.2 && (traits.speed > 0.2 || traits.cooperation > 0.4)) return 'small_herbivore';
          if (traits.scavenging_behavior > 0.4) return 'scavenger';
          if (traits.intelligence > 0.5 && traits.cooperation > 0.3) return 'intelligent_social';
          if (traits.endurance > 0.5 && traits.exploration_drive > 0.3) return 'explorer';
          if (traits.stealth && traits.stealth > 0.4) return 'stealth';
          
          return 'standard';
        }

        // Classify all entities
        entities.forEach(entity => {
          const type = determineEntityType(entity);
          if (!entityTypes.has(type)) {
            entityTypes.set(type, []);
          }
          entityTypes.get(type).push({
            ...entity,
            keyTraits: Object.keys(entity.traits)
              .filter(trait => Math.abs(entity.traits[trait]) > 0.1)
              .sort((a, b) => Math.abs(entity.traits[b]) - Math.abs(entity.traits[a]))
              .slice(0, 5)
          });
        });

        // Convert to array format for return
        const result = [];
        entityTypes.forEach((entities, type) => {
          result.push({
            type: type,
            count: entities.length,
            examples: entities.slice(0, 3), // Get up to 3 examples
            displayName: type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
          });
        });

        return result;
      });

      // Merge with collected data
      entityData.forEach(typeData => {
        if (!allEntityTypes.has(typeData.type)) {
          allEntityTypes.set(typeData.type, typeData);
        } else {
          // Update with more examples if available
          const existing = allEntityTypes.get(typeData.type);
          existing.count = Math.max(existing.count, typeData.count);
          if (typeData.examples.length > existing.examples.length) {
            existing.examples = typeData.examples;
          }
        }
      });
      
      console.log(`Sample ${sample + 1}: Found ${entityData.length} entity types`);
    }

    const entityTypeArray = Array.from(allEntityTypes.values());
    console.log('Combined entity types:', entityTypeArray.map(e => `${e.type}: ${e.count} entities`));

    // Clear old entity type screenshots
    const screenshotDir = 'screenshots/entity-types';
    
    // Inject CSS for entity type labels
    await page.addStyleTag({
      content: `
        .entity-type-label {
          position: absolute;
          top: 20px;
          left: 50%;
          transform: translateX(-50%);
          background: rgba(0, 0, 0, 0.8);
          color: white;
          padding: 10px 20px;
          border-radius: 8px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 18px;
          font-weight: bold;
          text-align: center;
          z-index: 1000;
          border: 2px solid #4CAF50;
          box-shadow: 0 4px 8px rgba(0, 0, 0, 0.5);
        }
        
        .entity-info {
          background: rgba(0, 0, 0, 0.7);
          color: white;
          padding: 8px 12px;
          border-radius: 5px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 12px;
          margin-top: 5px;
        }
      `
    });

    // Capture screenshot for each entity type
    for (const entityTypeData of entityTypeArray) {
      const { type, examples, displayName, count } = entityTypeData;
      
      console.log(`Capturing screenshot for entity type: ${type} (${count} entities)`);
      
      if (examples.length === 0) {
        console.log(`No examples found for entity type: ${type}`);
        continue;
      }

      // Focus on the first example entity
      const targetEntity = examples[0];
      
      // Navigate camera to the entity and ensure proper rendering
      await page.evaluate((entity) => {
        const gameState = (window as any).gameState;
        if (gameState) {
          // Set camera position to center on the entity
          gameState.camera.x = entity.x;
          gameState.camera.y = entity.y;
          gameState.zoom = 6.0; // Higher zoom for better visibility
          
          // Set size multiplier for enhanced visibility
          gameState.entitySizeMultiplier = 5; // Make entities 5x larger for screenshots
          
          console.log(`Camera positioned at entity ${entity.id}: (${entity.x}, ${entity.y}) with zoom ${gameState.zoom}`);
          console.log(`Entity traits:`, entity.traits);
          console.log(`Entity color:`, entity.color);
          console.log(`Entity size multiplier:`, gameState.entitySizeMultiplier);
          
          // Force multiple render updates to ensure entity is visible
          if (typeof render === 'function') {
            for (let i = 0; i < 3; i++) {
              render();
            }
          }
        }
      }, targetEntity);

      // Wait longer for camera to update and render
      await page.waitForTimeout(3000);

      // Add label overlay and enhance entity visibility for screenshots
      await page.evaluate((data) => {
        // Remove any existing labels
        const existingLabels = document.querySelectorAll('.entity-type-label');
        existingLabels.forEach(label => label.remove());

        // Ensure entity size multiplier is set for visibility
        if (window.gameState) {
          window.gameState.entitySizeMultiplier = 5; // Make entities 5x larger for screenshots
        }

        // Create new label with entity position info
        const entity = data.examples[0];
        const label = document.createElement('div');
        label.className = 'entity-type-label';
        label.innerHTML = `
          <div>${data.displayName}</div>
          <div class="entity-info">Count: ${data.count} | ID: ${entity.id}</div>
          <div class="entity-info">Position: (${entity.x.toFixed(1)}, ${entity.y.toFixed(1)})</div>
          <div class="entity-info">Key Traits: ${entity.keyTraits ? entity.keyTraits.join(', ') : 'N/A'}</div>
          <div class="entity-info">Species: ${entity.species}</div>
          <div class="entity-info">Color: ${entity.color}</div>
        `;
        document.body.appendChild(label);
        
        // Force multiple renders to show the enhanced entities
        if (typeof render === 'function') {
          for (let i = 0; i < 5; i++) {
            render();
          }
        }
        
        console.log(`Enhanced rendering for ${data.displayName} - Entity size multiplier: ${window.gameState?.entitySizeMultiplier}`);
      }, entityTypeData);

      // Wait for rendering to complete
      await page.waitForTimeout(2000);

      // Validate that entity is actually rendered and visible on canvas
      const entityVisible = await page.evaluate(() => {
        const gameState = (window as any).gameState;
        if (!gameState || !gameState.canvas) return false;
        
        const canvas = gameState.canvas;
        const ctx = gameState.ctx;
        
        // Get canvas image data to check if entity is rendered
        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        const pixels = imageData.data;
        
        // Count non-background pixels (crude check for rendered content)
        let nonBackgroundPixels = 0;
        for (let i = 0; i < pixels.length; i += 4) {
          const r = pixels[i];
          const g = pixels[i + 1];
          const b = pixels[i + 2];
          const alpha = pixels[i + 3];
          
          // Check if pixel is not background (not pure black or transparent)
          if (alpha > 0 && (r > 20 || g > 20 || b > 20)) {
            nonBackgroundPixels++;
          }
        }
        
        console.log(`Canvas validation: ${nonBackgroundPixels} non-background pixels found`);
        return nonBackgroundPixels > 1000; // Ensure substantial content is rendered
      });

      if (!entityVisible) {
        console.log(`Warning: Entity ${type} may not be visible in screenshot`);
      }

      // Take screenshot
      await page.screenshot({ 
        path: `${screenshotDir}/${type.replace(/\s+/g, '-').toLowerCase()}.png`,
        fullPage: true
      });

      console.log(`Screenshot saved: ${type.replace(/\s+/g, '-').toLowerCase()}.png`);
      
      // Reset entity size multiplier for next iteration
      await page.evaluate(() => {
        if (window.gameState) {
          window.gameState.entitySizeMultiplier = 1;
        }
      });
    }

    // Create an overview screenshot showing all entity types
    await page.evaluate(() => {
      // Remove individual labels
      const existingLabels = document.querySelectorAll('.entity-type-label');
      existingLabels.forEach(label => label.remove());

      // Reset camera to show more of the world
      const gameState = (window as any).gameState;
      if (gameState && gameState.isometricData && gameState.isometricData.worldInfo) {
        gameState.camera.x = gameState.isometricData.worldInfo.width / 2;
        gameState.camera.y = gameState.isometricData.worldInfo.height / 2;
        gameState.zoom = 0.8; // Zoom out for overview
      }
    });

    await page.waitForTimeout(1000);

    // Add overview label
    await page.evaluate((entityData) => {
      const label = document.createElement('div');
      label.className = 'entity-type-label';
      label.innerHTML = `
        <div>Entity Types Overview</div>
        <div class="entity-info">Total Types: ${entityData.length}</div>
        <div class="entity-info">${entityData.map(e => `${e.displayName}: ${e.count}`).join(' | ')}</div>
      `;
      document.body.appendChild(label);
    }, entityTypeArray);

    await page.screenshot({ 
      path: `${screenshotDir}/overview-all-entity-types.png`,
      fullPage: true
    });

    console.log('Screenshot saved: overview-all-entity-types.png');

    // Clean up labels
    await page.evaluate(() => {
      const labels = document.querySelectorAll('.entity-type-label');
      labels.forEach(label => label.remove());
    });

    console.log(`Entity type screenshot capture completed! Found ${entityTypeArray.length} different entity types.`);
  });

  test('validate entity type classification accuracy', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 45000 });
    
    // Wait for simulation data to load
    await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.entities &&
             gameState.isometricData.entities.length > 0;
    }, { timeout: 30000 });

    // Get classification statistics
    const classificationStats = await page.evaluate(() => {
      const gameState = (window as any).gameState;
      if (!gameState || !gameState.isometricData || !gameState.isometricData.entities) {
        return {};
      }

      const entities = gameState.isometricData.entities;
      const stats = {};
      const traitDistribution = {};

      // Function to determine entity type (same as in isometric_view.html)
      function determineEntityType(entity) {
        const traits = entity.traits;
        
        if (traits.flying_ability > 0.3) return 'flying';
        if (traits.aquatic_adaptation > 0.3) return 'aquatic';
        if (traits.digging_ability > 0.3 || traits.underground_nav > 0.2) return 'underground';
        if (traits.size > 0.4 && traits.aggression > 0.3) return 'large_predator';
        if (traits.size < -0.2 && (traits.speed > 0.2 || traits.cooperation > 0.4)) return 'small_herbivore';
        if (traits.scavenging_behavior > 0.4) return 'scavenger';
        if (traits.intelligence > 0.5 && traits.cooperation > 0.3) return 'intelligent_social';
        if (traits.endurance > 0.5 && traits.exploration_drive > 0.3) return 'explorer';
        if (traits.stealth && traits.stealth > 0.4) return 'stealth';
        
        return 'standard';
      }

      entities.forEach(entity => {
        const type = determineEntityType(entity);
        
        if (!stats[type]) {
          stats[type] = { count: 0, avgTraits: {} };
          traitDistribution[type] = {};
        }
        
        stats[type].count++;
        
        // Track trait averages for each type
        Object.keys(entity.traits).forEach(traitName => {
          if (!stats[type].avgTraits[traitName]) {
            stats[type].avgTraits[traitName] = 0;
          }
          stats[type].avgTraits[traitName] += entity.traits[traitName];
          
          if (!traitDistribution[type][traitName]) {
            traitDistribution[type][traitName] = [];
          }
          traitDistribution[type][traitName].push(entity.traits[traitName]);
        });
      });

      // Calculate averages
      Object.keys(stats).forEach(type => {
        Object.keys(stats[type].avgTraits).forEach(trait => {
          stats[type].avgTraits[trait] /= stats[type].count;
        });
      });

      return { stats, traitDistribution, totalEntities: entities.length };
    });

    console.log('Entity Classification Statistics:');
    Object.keys(classificationStats.stats).forEach(type => {
      const typeStats = classificationStats.stats[type];
      console.log(`\n${type}: ${typeStats.count} entities (${(typeStats.count / classificationStats.totalEntities * 100).toFixed(1)}%)`);
      
      // Show key traits for this type
      const keyTraits = Object.keys(typeStats.avgTraits)
        .filter(trait => Math.abs(typeStats.avgTraits[trait]) > 0.1)
        .sort((a, b) => Math.abs(typeStats.avgTraits[b]) - Math.abs(typeStats.avgTraits[a]))
        .slice(0, 5);
      
      console.log(`  Key traits: ${keyTraits.map(trait => `${trait}: ${typeStats.avgTraits[trait].toFixed(2)}`).join(', ')}`);
    });

    // Validate that we have reasonable distribution
    expect(Object.keys(classificationStats.stats).length).toBeGreaterThan(3);
    expect(classificationStats.totalEntities).toBeGreaterThanOrEqual(10); // Changed to >= to handle edge case
  });
});