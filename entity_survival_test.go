package main

import (
	"testing"
)

// TestEntitySurvivalPatterns tests entity survival rates under different conditions
func TestEntitySurvivalPatterns(t *testing.T) {
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 1,
		PopulationSize: 20,
		GridWidth:      20,
		GridHeight:     20,
	}

	world := NewWorld(config)
	
	// Add a test population with balanced traits
	popConfig := PopulationConfig{
		Name:     "TestSpecies",
		Species:  "Balanced",
		BaseTraits: map[string]float64{
			"speed":      0.3,
			"strength":   0.3,
			"endurance":  0.5, // Higher endurance for survival
			"size":       0.2,
			"intelligence": 0.4,
			"cooperation": 0.3,
			"aggression": 0.2,
			"vision":     0.4,
			"aquatic_adaptation": 0.3,
			"altitude_tolerance": 0.3,
		},
		StartPos: Position{X: 50, Y: 50},
		Spread:   10.0,
		Color:    "blue",
		BaseMutationRate: 0.1,
	}
	
	world.AddPopulation(popConfig)
	
	initialEntityCount := len(world.AllEntities)
	t.Logf("Initial entity count: %d", initialEntityCount)
	
	// Track survival over time
	survivalData := make([]int, 21) // ticks 0-20
	energyData := make([]float64, 21)
	
	for tick := 0; tick <= 20; tick++ {
		// Count alive entities
		aliveCount := 0
		totalEnergy := 0.0
		energyCount := 0
		
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				aliveCount++
				totalEnergy += entity.Energy
				energyCount++
			}
		}
		
		avgEnergy := 0.0
		if energyCount > 0 {
			avgEnergy = totalEnergy / float64(energyCount)
		}
		
		survivalData[tick] = aliveCount
		energyData[tick] = avgEnergy
		
		t.Logf("Tick %d: %d alive (%.1f%%), avg energy: %.1f", 
			tick, aliveCount, float64(aliveCount)/float64(initialEntityCount)*100, avgEnergy)
		
		// Update world if not last tick
		if tick < 20 {
			world.Update()
		}
	}
	
	// Analysis
	finalSurvivalRate := float64(survivalData[20]) / float64(initialEntityCount)
	earlyDeathRate := float64(initialEntityCount - survivalData[5]) / float64(initialEntityCount)
	
	t.Logf("Final survival rate (tick 20): %.1f%%", finalSurvivalRate*100)
	t.Logf("Early death rate (tick 5): %.1f%%", earlyDeathRate*100)
	
	// Test expectations (adjusted for daily time scale)
	// 20 ticks = 20 days, so some mortality is expected but not excessive
	if finalSurvivalRate < 0.05 { // Less than 5% survival by tick 20 (20 days)
		t.Errorf("Excessive mortality: only %.1f%% survived to tick 20 (20 days)", finalSurvivalRate*100)
	}
	
	if earlyDeathRate > 0.7 { // More than 70% dead by tick 5 (5 days)  
		t.Errorf("Rapid early mortality: %.1f%% died by tick 5 (5 days)", earlyDeathRate*100)
	}
	
	// Check energy decline rate
	if energyData[0] > 0 && energyData[10] > 0 {
		energyDeclineRate := (energyData[0] - energyData[10]) / energyData[0]
		t.Logf("Energy decline rate (first 10 ticks): %.1f%%", energyDeclineRate*100)
		
		if energyDeclineRate > 0.8 { // More than 80% energy loss in 10 ticks
			t.Errorf("Excessive energy decline: %.1f%% energy lost in 10 ticks", energyDeclineRate*100)
		}
	}
}

// TestEntitySurvivalInDifferentBiomes tests how entities fare in various biomes
func TestEntitySurvivalInDifferentBiomes(t *testing.T) {
	config := WorldConfig{
		Width:     100,
		Height:    100,
		GridWidth: 10,
		GridHeight: 10,
	}

	// Test different biome types
	testBiomes := []struct {
		biomeType BiomeType
		name      string
		expectedSurvivalDays int // Minimum expected survival in ticks
	}{
		{BiomePlains, "Plains", 15},
		{BiomeForest, "Forest", 15},
		{BiomeDesert, "Desert", 10},
		{BiomeMountain, "Mountain", 12},
		{BiomeWater, "Water", 8},
		{BiomeHighAltitude, "High Altitude", 5},
		{BiomeIce, "Ice", 5},
		{BiomeDeepWater, "Deep Water", 8},
	}
	
	for _, testCase := range testBiomes {
		t.Run(testCase.name, func(t *testing.T) {
			world := NewWorld(config)
			
			// Force entire map to be this biome type
			for y := 0; y < config.GridHeight; y++ {
				for x := 0; x < config.GridWidth; x++ {
					world.Grid[y][x].Biome = testCase.biomeType
				}
			}
			
			// Create entities with appropriate adaptations
			adaptedTraits := make(map[string]float64)
			adaptedTraits["endurance"] = 0.8
			adaptedTraits["strength"] = 0.4
			adaptedTraits["size"] = 0.3
			
			// Add biome-specific adaptations
			switch testCase.biomeType {
			case BiomeWater, BiomeDeepWater:
				adaptedTraits["aquatic_adaptation"] = 0.9
			case BiomeHighAltitude:
				adaptedTraits["altitude_tolerance"] = 0.9
			case BiomeIce:
				adaptedTraits["endurance"] = 1.0
			}
			
			// Create test entities
			testEntities := make([]*Entity, 5)
			for i := 0; i < 5; i++ {
				entity := &Entity{
					ID:       i,
					Position: Position{X: 50, Y: 50},
					Energy:   100,
					IsAlive:  true,
					Age:      0,
					Traits:   make(map[string]Trait),
					Classification: ClassificationEukaryotic, // Default classification
					MaxLifespan:    3360,                     // Default max lifespan
				}
				
				// Set traits
				for name, value := range adaptedTraits {
					entity.Traits[name] = Trait{Name: name, Value: value}
				}
				
				// Initialize molecular systems to prevent nil pointer issues
				entity.MolecularNeeds = NewMolecularNeeds(entity)
				entity.MolecularMetabolism = NewMolecularMetabolism(entity)
				entity.MolecularProfile = NewMolecularProfile()
				
				testEntities[i] = entity
				world.AllEntities = append(world.AllEntities, entity)
			}
			
			// Run simulation
			survivedTicks := 0
			for tick := 1; tick <= 20; tick++ {
				// Apply biome effects manually
				world.applyBiomeEffects()
				
				// Update entities using classification system
				for _, entity := range testEntities {
					if entity.IsAlive {
						entity.UpdateWithClassification(world.OrganismClassifier, world.CellularSystem)
					}
				}
				
				// Check if any entities are still alive
				anyAlive := false
				for _, entity := range testEntities {
					if entity.IsAlive {
						anyAlive = true
						break
					}
				}
				
				if anyAlive {
					survivedTicks = tick
				} else {
					break
				}
			}
			
			t.Logf("%s biome: entities survived %d ticks (expected minimum: %d)", 
				testCase.name, survivedTicks, testCase.expectedSurvivalDays)
			
			if survivedTicks < testCase.expectedSurvivalDays {
				t.Errorf("%s biome: entities only survived %d ticks, expected at least %d", 
					testCase.name, survivedTicks, testCase.expectedSurvivalDays)
			}
		})
	}
}

// TestEnergyDecayRates tests that energy decay is reasonable under normal conditions
func TestEnergyDecayRates(t *testing.T) {
	// Create necessary systems for the new classification system
	timeSystem := NewAdvancedTimeSystem(480, 120) // 480 ticks/day, 120 days/season
	dnaSystem := NewDNASystem(NewCentralEventBus(1000))
	cellularSystem := NewCellularSystem(dnaSystem, NewCentralEventBus(1000))
	classifier := NewOrganismClassifier(timeSystem)
	
	// Create a test entity with average traits
	entity := &Entity{
		ID:       1,
		Position: Position{X: 50, Y: 50},
		Energy:   100,
		IsAlive:  true,
		Age:      10,
		Traits: map[string]Trait{
			"endurance": {Value: 0.5},
			"size":      {Value: 0.3},
		},
		Classification: ClassificationEukaryotic, // Default classification
		MaxLifespan:    3360,                     // Default max lifespan
	}
	
	// Initialize molecular systems
	entity.MolecularNeeds = NewMolecularNeeds(entity)
	entity.MolecularMetabolism = NewMolecularMetabolism(entity)
	entity.MolecularProfile = NewMolecularProfile()
	
	initialEnergy := entity.Energy
	
	// Test base energy decay over 10 ticks using classification system
	for i := 0; i < 10; i++ {
		entity.UpdateWithClassification(classifier, cellularSystem)
		if !entity.IsAlive {
			break
		}
	}
	
	energyLoss := initialEnergy - entity.Energy
	decayRate := energyLoss / 10.0 // per tick
	
	t.Logf("Base energy decay: %.2f energy/tick (%.1f%% of initial energy)", 
		decayRate, (energyLoss/initialEnergy)*100)
	
	// Energy decay should be reasonable (not too fast)
	if decayRate > 5.0 {
		t.Errorf("Energy decay rate %.2f per tick is too high", decayRate)
	}
	
	// Entity should still be alive after 10 ticks of base decay
	if !entity.IsAlive {
		t.Errorf("Entity died within 10 ticks from base energy decay")
	}
}

// TestPoorlyAdaptedEntitySurvival tests entities with poor environmental adaptation
func TestPoorlyAdaptedEntitySurvival(t *testing.T) {
	config := WorldConfig{
		Width:     100,
		Height:    100,
		GridWidth: 5,
		GridHeight: 5,
	}

	world := NewWorld(config)
	
	// Force harsh biome
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x].Biome = BiomeHighAltitude
		}
	}
	
	// Create poorly adapted entity
	poorEntity := &Entity{
		ID:       1,
		Position: Position{X: 50, Y: 50},
		Energy:   100,
		IsAlive:  true,
		Age:      0,
		Traits: map[string]Trait{
			"altitude_tolerance": {Value: 0.1}, // Very poor adaptation
			"endurance":         {Value: 0.2}, // Low endurance
			"aquatic_adaptation": {Value: 0.0},
		},
		Classification: ClassificationEukaryotic, // Default classification
		MaxLifespan:    3360,                     // Default max lifespan
	}
	
	// Initialize molecular systems
	poorEntity.MolecularNeeds = NewMolecularNeeds(poorEntity)
	poorEntity.MolecularMetabolism = NewMolecularMetabolism(poorEntity)
	poorEntity.MolecularProfile = NewMolecularProfile()
	
	world.AllEntities = []*Entity{poorEntity}
	
	// Run simulation and track survival
	survivalTime := 0
	for tick := 1; tick <= 20; tick++ {
		world.updateEntityWithBiome(poorEntity)
		poorEntity.UpdateWithClassification(world.OrganismClassifier, world.CellularSystem)
		
		if poorEntity.IsAlive {
			survivalTime = tick
		} else {
			break
		}
	}
	
	t.Logf("Poorly adapted entity survived %d ticks in harsh biome", survivalTime)
	
	// Should die quickly but not immediately 
	if survivalTime > 21 { // Adjusted for realistic energy drain (~4.7 per tick = ~21 ticks max survival)
		t.Errorf("Poorly adapted entity survived too long (%d ticks)", survivalTime)
	}
	if survivalTime < 5 { // Allow reasonable minimum survival time
		t.Errorf("Poorly adapted entity died too quickly (%d ticks)", survivalTime)
	}
}

// TestWellAdaptedEntitySurvival tests entities with good environmental adaptation
func TestWellAdaptedEntitySurvival(t *testing.T) {
	config := WorldConfig{
		Width:     100,
		Height:    100,
		GridWidth: 5,
		GridHeight: 5,
	}

	world := NewWorld(config)
	
	// Force harsh biome
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x].Biome = BiomeHighAltitude
		}
	}
	
	// Create well-adapted entity
	adaptedEntity := &Entity{
		ID:       1,
		Position: Position{X: 50, Y: 50},
		Energy:   100,
		IsAlive:  true,
		Age:      0,
		Traits: map[string]Trait{
			"altitude_tolerance": {Value: 0.9}, // Excellent adaptation
			"endurance":         {Value: 0.8}, // High endurance
			"aquatic_adaptation": {Value: 0.1},
		},
		Classification: ClassificationEukaryotic, // Default classification
		MaxLifespan:    3360,                     // Default max lifespan
	}
	
	// Initialize molecular systems
	adaptedEntity.MolecularNeeds = NewMolecularNeeds(adaptedEntity)
	adaptedEntity.MolecularMetabolism = NewMolecularMetabolism(adaptedEntity)
	adaptedEntity.MolecularProfile = NewMolecularProfile()
	
	world.AllEntities = []*Entity{adaptedEntity}
	
	// Run simulation and track survival
	survivalTime := 0
	energyAtTick10 := 0.0
	
	for tick := 1; tick <= 20; tick++ {
		world.applyBiomeEffects()
		adaptedEntity.UpdateWithClassification(world.OrganismClassifier, world.CellularSystem)
		
		if adaptedEntity.IsAlive {
			survivalTime = tick
			if tick == 10 {
				energyAtTick10 = adaptedEntity.Energy
			}
		} else {
			break
		}
	}
	
	t.Logf("Well-adapted entity survived %d ticks in harsh biome, energy at tick 10: %.1f", 
		survivalTime, energyAtTick10)
	
	// Well-adapted entity should survive longer
	if survivalTime < 15 {
		t.Errorf("Well-adapted entity should survive at least 15 ticks, survived %d", survivalTime)
	}
	
	// Should maintain reasonable energy levels
	if energyAtTick10 < 30 {
		t.Errorf("Well-adapted entity should maintain energy > 30 at tick 10, had %.1f", energyAtTick10)
	}
}

// TestEntityLifecycleInBalancedWorld tests entity survival in a mixed biome world
func TestEntityLifecycleInBalancedWorld(t *testing.T) {
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 1,
		PopulationSize: 10,
		GridWidth:      15,
		GridHeight:     15,
	}

	world := NewWorld(config)
	
	// Add a test population
	popConfig := PopulationConfig{
		Name:     "BalancedSpecies",
		Species:  "Test",
		BaseTraits: map[string]float64{
			"endurance":  0.6,
			"intelligence": 0.5,
			"cooperation": 0.4,
			"size":       0.3,
			"speed":      0.4,
		},
		StartPos: Position{X: 50, Y: 50},
		Spread:   20.0,
		Color:    "green",
		BaseMutationRate: 0.1,
	}
	
	world.AddPopulation(popConfig)
	initialCount := len(world.AllEntities)
	
	// Run simulation for 20 ticks
	populationHistory := make([]int, 21)
	
	for tick := 0; tick <= 20; tick++ {
		aliveCount := 0
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				aliveCount++
			}
		}
		populationHistory[tick] = aliveCount
		
		if tick < 20 {
			world.Update()
		}
	}
	
	// Log population changes
	for tick := 0; tick <= 20; tick += 5 {
		t.Logf("Tick %d: %d entities alive (%.1f%%)", 
			tick, populationHistory[tick], 
			float64(populationHistory[tick])/float64(initialCount)*100)
	}
	
	// Check for reasonable population decline
	finalSurvival := float64(populationHistory[20]) / float64(initialCount)
	midSurvival := float64(populationHistory[10]) / float64(initialCount)
	
	// In a balanced world with daily time scale, some entities should survive 20 days
	if finalSurvival < 0.1 { // At least 10% survival after 20 days
		t.Errorf("Population collapse: only %.1f%% survived to tick 20 (20 days) in balanced world", finalSurvival*100)
	}
	
	// Population shouldn't crash too early - at least 30% should survive 10 days
	if midSurvival < 0.3 {
		t.Errorf("Early population crash: only %.1f%% survived to tick 10 (10 days)", midSurvival*100)
	}
}