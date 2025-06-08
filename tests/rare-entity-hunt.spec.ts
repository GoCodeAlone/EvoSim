import { test, expect } from '@playwright/test';

test.describe('Comprehensive Entity Type Documentation', () => {
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

  test('hunt for rare entity types and document them', async ({ page }) => {
    await page.goto('/iso', { waitUntil: 'networkidle', timeout: 30000 });
    
    // Wait for simulation data to load
    await page.waitForFunction(() => {
      const gameState = (window as any).gameState;
      return gameState && 
             gameState.websocket && 
             gameState.websocket.readyState === WebSocket.OPEN &&
             gameState.isometricData &&
             gameState.isometricData.entities &&
             gameState.isometricData.entities.length > 10;
    }, { timeout: 20000 });

    console.log('🧬 Starting rare entity type hunt...');

    // Track all types found over time
    const allFoundTypes = new Map();
    const targetRareTypes = ['intelligent_social', 'explorer', 'stealth'];
    let foundNewTypes = false;

    // Enhanced styling for rare types
    await page.addStyleTag({
      content: `
        .entity-type-label {
          position: absolute;
          top: 20px;
          left: 50%;
          transform: translateX(-50%);
          background: linear-gradient(135deg, rgba(30, 0, 50, 0.95), rgba(70, 20, 120, 0.95));
          color: white;
          padding: 18px 30px;
          border-radius: 15px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 22px;
          font-weight: bold;
          text-align: center;
          z-index: 1000;
          border: 3px solid #FFD700;
          box-shadow: 0 8px 25px rgba(255, 215, 0, 0.3);
          backdrop-filter: blur(8px);
        }
        
        .entity-info {
          background: rgba(50, 20, 80, 0.9);
          color: #F0F0F0;
          padding: 10px 18px;
          border-radius: 10px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 14px;
          margin-top: 10px;
          border-left: 5px solid #FFD700;
        }

        .trait-highlight {
          color: #FFD700;
          font-weight: bold;
          text-shadow: 0 0 3px rgba(255, 215, 0, 0.5);
        }

        .rare-type {
          background: linear-gradient(135deg, rgba(255, 0, 100, 0.3), rgba(255, 100, 0, 0.3));
          border: 3px solid #FF1493 !important;
        }
      `
    });

    // Check for new types multiple times with evolution time
    const maxChecks = 8;
    for (let check = 0; check < maxChecks; check++) {
      console.log(`🔍 Evolution check ${check + 1}/${maxChecks}...`);
      
      // Wait for evolution
      if (check > 0) {
        await page.waitForTimeout(3000);
      }

      const currentTypes = await page.evaluate(() => {
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
          examples: entities.slice(0, 1),
          totalEntities: gameState.isometricData.entities.length
        }));
      });

      // Update our collection
      currentTypes.forEach(typeData => {
        if (!allFoundTypes.has(typeData.type)) {
          allFoundTypes.set(typeData.type, typeData);
          console.log(`✨ New entity type discovered: ${typeData.displayName} (${typeData.count} entities)`);
          foundNewTypes = true;
        }
      });

      console.log(`Current types: ${currentTypes.map(t => `${t.type}(${t.count})`).join(', ')}`);
      console.log(`Total entities in simulation: ${currentTypes[0]?.totalEntities || 0}`);

      // Check if we found any of the rare types
      const foundRare = targetRareTypes.filter(rareType => allFoundTypes.has(rareType));
      if (foundRare.length > 0) {
        console.log(`🎯 Found rare types: ${foundRare.join(', ')}`);
      }

      // Stop early if population is getting too low
      const totalEntities = currentTypes.reduce((sum, t) => sum + t.count, 0);
      if (totalEntities < 5) {
        console.log(`⚠️ Population too low (${totalEntities}), stopping search`);
        break;
      }
    }

    const allTypes = Array.from(allFoundTypes.values());
    console.log(`\n🧬 Final entity type census: ${allTypes.length} types found`);
    allTypes.forEach(type => {
      console.log(`  - ${type.displayName}: ${type.count} entities`);
    });

    // Document any new types found
    if (foundNewTypes) {
      const screenshotDir = 'screenshots/entity-types';

      for (const typeData of allTypes) {
        if (typeData.examples.length === 0) continue;

        const isRareType = targetRareTypes.includes(typeData.type);
        console.log(`📸 Documenting: ${typeData.displayName}${isRareType ? ' ⭐ RARE' : ''}`);
        
        const example = typeData.examples[0];
        
        // Navigate to entity
        await page.evaluate((entity) => {
          const gameState = (window as any).gameState;
          if (gameState) {
            gameState.camera.x = entity.x;
            gameState.camera.y = entity.y;
            gameState.zoom = 3.5;
          }
        }, example);

        await page.waitForTimeout(500);

        // Add enhanced label for rare types
        await page.evaluate((data) => {
          document.querySelectorAll('.entity-type-label').forEach(el => el.remove());

          const isRare = ['intelligent_social', 'explorer', 'stealth'].includes(data.type);
          const label = document.createElement('div');
          label.className = 'entity-type-label' + (isRare ? ' rare-type' : '');
          label.innerHTML = `
            <div style="font-size: 26px; margin-bottom: 12px;">
              ${isRare ? '⭐ ' : ''}${data.displayName}${isRare ? ' ⭐' : ''}
            </div>
            <div class="entity-info">
              Population: <span class="trait-highlight">${data.count}</span> entities
              ${isRare ? ' | <span style="color: #FF69B4;">🎉 RARE TYPE</span>' : ''}
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

        console.log(`✓ Screenshot updated: ${typeData.type}.png`);
      }

      // Create comprehensive overview
      await page.evaluate((worldInfo) => {
        const gameState = (window as any).gameState;
        if (gameState?.isometricData?.worldInfo) {
          gameState.camera.x = worldInfo.width / 2;
          gameState.camera.y = worldInfo.height / 2;
          gameState.zoom = 0.6;
        }
      }, { width: 40, height: 25 });

      await page.waitForTimeout(1000);

      await page.evaluate((data) => {
        document.querySelectorAll('.entity-type-label').forEach(el => el.remove());

        const rareTypes = data.filter(t => ['intelligent_social', 'explorer', 'stealth'].includes(t.type));
        const commonTypes = data.filter(t => !['intelligent_social', 'explorer', 'stealth'].includes(t.type));

        const label = document.createElement('div');
        label.className = 'entity-type-label';
        label.innerHTML = `
          <div style="font-size: 28px; margin-bottom: 15px;">🧬 Complete Entity Type Catalog</div>
          <div class="entity-info">
            Total Types Discovered: <span class="trait-highlight">${data.length}</span>
          </div>
          <div class="entity-info">
            Common Types (${commonTypes.length}): ${commonTypes.map(e => `${e.displayName}: ${e.count}`).join(' • ')}
          </div>
          ${rareTypes.length > 0 ? `
          <div class="entity-info" style="border-left-color: #FF1493;">
            ⭐ Rare Types Found (${rareTypes.length}): ${rareTypes.map(e => `${e.displayName}: ${e.count}`).join(' • ')}
          </div>
          ` : ''}
        `;
        document.body.appendChild(label);
      }, allTypes);

      await page.screenshot({ 
        path: `screenshots/entity-types/complete-entity-catalog.png`,
        fullPage: true
      });

      console.log('✓ Complete catalog screenshot saved');
    }

    // Clean up
    await page.evaluate(() => {
      document.querySelectorAll('.entity-type-label').forEach(el => el.remove());
    });

    const rareFound = targetRareTypes.filter(rare => allFoundTypes.has(rare));
    console.log(`\n🎯 Hunt Complete!`);
    console.log(`📸 Total entity types documented: ${allTypes.length}`);
    console.log(`⭐ Rare types found: ${rareFound.length > 0 ? rareFound.join(', ') : 'None this time'}`);
    console.log(`🧬 All types: ${allTypes.map(e => e.displayName).join(', ')}`);
  });
});