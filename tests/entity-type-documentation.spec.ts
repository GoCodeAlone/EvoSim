import { test, expect } from '@playwright/test';

test.describe('Entity Type Visual Documentation', () => {
  test.beforeAll(async ({ request }) => {
    console.log('Checking if EvoSim webserver is running...');
    
    let retries = 0;
    const maxRetries = 10;
    const baseURL = 'http://localhost:8080';
    
    while (retries < maxRetries) {
      try {
        const response = await request.get(`${baseURL}/api/status`, { 
          timeout: 2000,
          ignoreHTTPSErrors: true 
        });
        
        if (response.status() === 200) {
          console.log('✓ EvoSim webserver detected and responding');
          return;
        }
      } catch (error) {
        // Server not responding, continue retrying
      }
      
      retries++;
      console.log(`Waiting for webserver... (attempt ${retries}/${maxRetries})`);
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
    
    throw new Error(`EvoSim webserver is not running or not responding at ${baseURL}`);
  });

  test('document all entity types with detailed screenshots', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 30000 });
    
    // Wait for simulation data to load
    await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.entities &&
             gameState.isometricData.entities.length > 5;
    }, { timeout: 15000 });

    // Inject enhanced label styling
    await page.addStyleTag({
      content: `
        .entity-type-label {
          position: absolute;
          top: 20px;
          left: 50%;
          transform: translateX(-50%);
          background: linear-gradient(135deg, rgba(0, 0, 0, 0.9), rgba(20, 20, 40, 0.9));
          color: white;
          padding: 15px 25px;
          border-radius: 12px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 20px;
          font-weight: bold;
          text-align: center;
          z-index: 1000;
          border: 3px solid #4CAF50;
          box-shadow: 0 6px 20px rgba(0, 0, 0, 0.6);
          backdrop-filter: blur(5px);
        }
        
        .entity-info {
          background: rgba(30, 30, 60, 0.8);
          color: #E0E0E0;
          padding: 8px 15px;
          border-radius: 8px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 14px;
          margin-top: 8px;
          border-left: 4px solid #4CAF50;
        }

        .trait-highlight {
          color: #FFD700;
          font-weight: bold;
        }
      `
    });

    // Collect entity data
    const entityData = await page.evaluate(() => {
      const gameState = (window as any).gameState;
      if (!gameState?.isometricData?.entities) return [];

      const entities = gameState.isometricData.entities;
      const entityTypes = new Map();

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

      function getKeyTraits(entity) {
        return Object.entries(entity.traits)
          .filter(([_, value]) => Math.abs(value) > 0.1)
          .sort((a, b) => Math.abs(b[1]) - Math.abs(a[1]))
          .slice(0, 4)
          .map(([trait, value]) => `${trait}: ${value.toFixed(2)}`);
      }

      entities.forEach(entity => {
        const type = determineEntityType(entity);
        const entityInfo = {
          ...entity,
          keyTraits: getKeyTraits(entity),
          displayName: type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
        };

        if (!entityTypes.has(type)) {
          entityTypes.set(type, []);
        }
        entityTypes.get(type).push(entityInfo);
      });

      return Array.from(entityTypes.entries()).map(([type, entities]) => ({
        type,
        displayName: type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase()),
        count: entities.length,
        examples: entities.slice(0, 2)
      }));
    });

    console.log(`Found ${entityData.length} entity types:`, entityData.map(e => `${e.type} (${e.count})`).join(', '));

    // Clear existing screenshots
    const screenshotDir = 'screenshots/entity-types';

    // Create individual entity type screenshots
    for (const typeData of entityData) {
      if (typeData.examples.length === 0) continue;

      console.log(`Creating screenshot for: ${typeData.displayName}`);
      
      const example = typeData.examples[0];
      
      // Navigate to entity
      await page.evaluate((entity) => {
        const gameState = (window as any).gameState;
        if (gameState) {
          gameState.camera.x = entity.x;
          gameState.camera.y = entity.y;
          gameState.zoom = 3.0; // Close-up view
        }
      }, example);

      await page.waitForTimeout(800);

      // Add informative label
      await page.evaluate((data) => {
        document.querySelectorAll('.entity-type-label').forEach(el => el.remove());

        const label = document.createElement('div');
        label.className = 'entity-type-label';
        label.innerHTML = `
          <div style="font-size: 24px; margin-bottom: 10px;">${data.displayName}</div>
          <div class="entity-info">
            Population: <span class="trait-highlight">${data.count}</span> entities
          </div>
          <div class="entity-info">
            Species: <span class="trait-highlight">${data.examples[0].species}</span>
          </div>
          <div class="entity-info">
            Entity ID: <span class="trait-highlight">${data.examples[0].id}</span>
          </div>
          <div class="entity-info">
            Key Traits: <span class="trait-highlight">${data.examples[0].keyTraits.join(', ')}</span>
          </div>
        `;
        document.body.appendChild(label);
      }, typeData);

      // Take screenshot
      await page.screenshot({ 
        path: `${screenshotDir}/${typeData.type}.png`,
        fullPage: true
      });

      console.log(`✓ Screenshot saved: ${typeData.type}.png`);
    }

    // Create overview
    await page.evaluate((worldInfo) => {
      const gameState = (window as any).gameState;
      if (gameState?.isometricData?.worldInfo) {
        gameState.camera.x = worldInfo.width / 2;
        gameState.camera.y = worldInfo.height / 2;
        gameState.zoom = 0.7;
      }
    }, { width: 40, height: 25 });

    await page.waitForTimeout(1000);

    await page.evaluate((data) => {
      document.querySelectorAll('.entity-type-label').forEach(el => el.remove());

      const label = document.createElement('div');
      label.className = 'entity-type-label';
      label.innerHTML = `
        <div style="font-size: 26px; margin-bottom: 12px;">🧬 EvoSim Entity Types Overview</div>
        <div class="entity-info">
          Total Types Discovered: <span class="trait-highlight">${data.length}</span>
        </div>
        <div class="entity-info">
          ${data.map(e => `${e.displayName}: ${e.count}`).join(' • ')}
        </div>
      `;
      document.body.appendChild(label);
    }, entityData);

    await page.screenshot({ 
      path: `${screenshotDir}/entity-types-overview.png`,
      fullPage: true
    });

    console.log('✓ Overview screenshot saved');

    // Clean up
    await page.evaluate(() => {
      document.querySelectorAll('.entity-type-label').forEach(el => el.remove());
    });

    console.log(`\n🎯 Entity Type Documentation Complete!`);
    console.log(`📸 Generated ${entityData.length + 1} screenshots in ${screenshotDir}/`);
    console.log(`🧬 Entity types found: ${entityData.map(e => e.displayName).join(', ')}`);
  });
});