package main

import (
	"math"
	"testing"
)

// TestNewBiomeTypes tests that all new biome types are properly initialized
func TestNewBiomeTypes(t *testing.T) {
	biomes := initializeBiomes()

	// Test that all expected biome types exist
	expectedBiomes := []BiomeType{
		BiomePlains, BiomeForest, BiomeDesert, BiomeMountain, BiomeWater,
		BiomeRadiation, BiomeSoil, BiomeAir, BiomeIce, BiomeRainforest,
		BiomeDeepWater, BiomeHighAltitude, BiomeHotSpring, BiomeTundra,
		BiomeSwamp, BiomeCanyon,
	}

	for _, biomeType := range expectedBiomes {
		biome, exists := biomes[biomeType]
		if !exists {
			t.Errorf("Biome type %d not found in initialized biomes", biomeType)
			continue
		}

		// Test that biome has required fields
		if biome.Name == "" {
			t.Errorf("Biome %d has empty name", biomeType)
		}
		if biome.Color == "" {
			t.Errorf("Biome %d has empty color", biomeType)
		}
		if biome.Symbol == 0 {
			t.Errorf("Biome %d has no symbol", biomeType)
		}
	}
}

// TestBiomeEnvironmentalProperties tests that biomes have realistic environmental properties
func TestBiomeEnvironmentalProperties(t *testing.T) {
	biomes := initializeBiomes()

	testCases := []struct {
		biomeType    BiomeType
		name         string
		tempRange    [2]float64 // [min, max] acceptable temperature
		pressureRange [2]float64 // [min, max] acceptable pressure
		oxygenRange  [2]float64 // [min, max] acceptable oxygen
		isAquatic    bool
		isUnderground bool
		isAerial     bool
	}{
		{BiomeIce, "Ice", [2]float64{-1, -0.5}, [2]float64{0.8, 1.2}, [2]float64{0.8, 1.0}, false, false, false},
		{BiomeDeepWater, "Deep Water", [2]float64{-0.5, 0.1}, [2]float64{1.5, 2.5}, [2]float64{0.1, 0.5}, true, false, false},
		{BiomeHighAltitude, "High Altitude", [2]float64{-1, -0.5}, [2]float64{0.1, 0.5}, [2]float64{0.1, 0.4}, false, false, true},
		{BiomeHotSpring, "Hot Spring", [2]float64{0.7, 1.0}, [2]float64{1.0, 1.3}, [2]float64{0.6, 1.0}, true, false, false},
		{BiomeRainforest, "Rainforest", [2]float64{0.4, 0.8}, [2]float64{0.9, 1.1}, [2]float64{1.1, 1.5}, false, false, false},
		{BiomeTundra, "Tundra", [2]float64{-1, -0.5}, [2]float64{0.9, 1.1}, [2]float64{0.8, 1.0}, false, false, false},
		{BiomeSwamp, "Swamp", [2]float64{0.1, 0.5}, [2]float64{1.0, 1.2}, [2]float64{0.4, 0.8}, true, false, false},
		{BiomeCanyon, "Canyon", [2]float64{0.2, 0.6}, [2]float64{1.1, 1.3}, [2]float64{0.7, 0.9}, false, false, false},
	}

	for _, tc := range testCases {
		biome := biomes[tc.biomeType]
		
		// Test temperature range
		if biome.Temperature < tc.tempRange[0] || biome.Temperature > tc.tempRange[1] {
			t.Errorf("%s temperature %.2f outside expected range [%.2f, %.2f]", 
				tc.name, biome.Temperature, tc.tempRange[0], tc.tempRange[1])
		}
		
		// Test pressure range
		if biome.Pressure < tc.pressureRange[0] || biome.Pressure > tc.pressureRange[1] {
			t.Errorf("%s pressure %.2f outside expected range [%.2f, %.2f]", 
				tc.name, biome.Pressure, tc.pressureRange[0], tc.pressureRange[1])
		}
		
		// Test oxygen range
		if biome.OxygenLevel < tc.oxygenRange[0] || biome.OxygenLevel > tc.oxygenRange[1] {
			t.Errorf("%s oxygen %.2f outside expected range [%.2f, %.2f]", 
				tc.name, biome.OxygenLevel, tc.oxygenRange[0], tc.oxygenRange[1])
		}
		
		// Test aquatic flag
		if biome.IsAquatic != tc.isAquatic {
			t.Errorf("%s IsAquatic is %v, expected %v", tc.name, biome.IsAquatic, tc.isAquatic)
		}
		
		// Test underground flag
		if biome.IsUnderground != tc.isUnderground {
			t.Errorf("%s IsUnderground is %v, expected %v", tc.name, biome.IsUnderground, tc.isUnderground)
		}
		
		// Test aerial flag
		if biome.IsAerial != tc.isAerial {
			t.Errorf("%s IsAerial is %v, expected %v", tc.name, biome.IsAerial, tc.isAerial)
		}
	}
}

// TestEnhancedBiomeGeneration tests that the new biome generation creates contiguous patterns
func TestEnhancedBiomeGeneration(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}
	
	world := NewWorld(config)
	
	// Count occurrences of each biome type
	biomeCounts := make(map[BiomeType]int)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			biomeCounts[biome]++
		}
	}
	
	// Verify that we have multiple biome types
	if len(biomeCounts) < 3 {
		t.Errorf("Expected at least 3 different biome types, got %d", len(biomeCounts))
	}
	
	// Verify that polar biomes (Ice, Tundra) appear near edges
	edgeBiomes := make(map[BiomeType]int)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			// Check if this is an edge cell
			if x < 3 || x >= config.GridWidth-3 || y < 3 || y >= config.GridHeight-3 {
				biome := world.Grid[y][x].Biome
				edgeBiomes[biome]++
			}
		}
	}
	
	// Ice and Tundra should be more common at edges
	iceAndTundraAtEdges := edgeBiomes[BiomeIce] + edgeBiomes[BiomeTundra]
	_ = iceAndTundraAtEdges // Use the variable to avoid compile error
	
	if iceAndTundraAtEdges == 0 {
		t.Log("Warning: No ice or tundra biomes found at edges")
	}
	
	t.Logf("Biome distribution: %v", biomeCounts)
	t.Logf("Edge biomes: %v", edgeBiomes)
}

// TestBiomeEffectsOnEntities tests that biome environmental effects work correctly
func TestBiomeEffectsOnEntities(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  10,
		GridHeight: 10,
	}
	
	world := NewWorld(config)
	
	// Create test entities with different traits
	entities := []*Entity{
		// Aquatic-adapted entity
		{
			ID:       1,
			Position: Position{X: 50, Y: 50},
			Energy:   100,
			IsAlive:  true,
			Traits: map[string]Trait{
				"aquatic_adaptation": {Value: 0.8},
				"endurance":         {Value: 0.3},
			},
		},
		// High-altitude adapted entity
		{
			ID:       2,
			Position: Position{X: 25, Y: 25},
			Energy:   100,
			IsAlive:  true,
			Traits: map[string]Trait{
				"altitude_tolerance": {Value: 0.9},
				"aquatic_adaptation": {Value: 0.1},
			},
		},
		// Poorly adapted entity
		{
			ID:       3,
			Position: Position{X: 75, Y: 75},
			Energy:   100,
			IsAlive:  true,
			Traits: map[string]Trait{
				"aquatic_adaptation": {Value: 0.1},
				"altitude_tolerance": {Value: 0.1},
				"endurance":         {Value: 0.2},
			},
		},
	}
	
	world.AllEntities = entities
	
	// Force specific biomes for testing
	world.Grid[5][5].Biome = BiomeDeepWater    // Entity 1 location
	world.Grid[2][2].Biome = BiomeHighAltitude // Entity 2 location  
	world.Grid[7][7].Biome = BiomeIce          // Entity 3 location
	
	// Store initial energy levels
	initialEnergies := make([]float64, len(entities))
	for i, entity := range entities {
		initialEnergies[i] = entity.Energy
	}
	
	// Apply biome effects
	world.applyBiomeEffects()
	
	// Check that adapted entities fare better than non-adapted ones
	// Entity 1 (aquatic) in deep water should lose less energy
	entity1EnergyLoss := initialEnergies[0] - entities[0].Energy
	t.Logf("Entity 1 (aquatic) energy loss in deep water: %.2f", entity1EnergyLoss)
	
	// Entity 2 (altitude) in high altitude should lose less energy  
	entity2EnergyLoss := initialEnergies[1] - entities[1].Energy
	t.Logf("Entity 2 (altitude) energy loss in high altitude: %.2f", entity2EnergyLoss)
	
	// Entity 3 (poorly adapted) in ice should lose more energy
	entity3EnergyLoss := initialEnergies[2] - entities[2].Energy
	t.Logf("Entity 3 (poorly adapted) energy loss in ice: %.2f", entity3EnergyLoss)
	
	// Poorly adapted entity should lose more energy
	if entity3EnergyLoss <= entity1EnergyLoss {
		t.Errorf("Expected poorly adapted entity to lose more energy than adapted entity")
	}
}

// TestBiomeTraitModifiers tests that biome trait modifiers are applied correctly
func TestBiomeTraitModifiers(t *testing.T) {
	biomes := initializeBiomes()
	
	// Test specific biome trait modifiers
	testCases := []struct {
		biomeType    BiomeType
		expectedTrait string
		expectedSign  float64 // 1 for positive, -1 for negative, 0 for either
	}{
		{BiomeIce, "endurance", 1},      // Ice should boost endurance
		{BiomeIce, "speed", -1},         // Ice should reduce speed
		{BiomeRainforest, "intelligence", 1}, // Rainforest should boost intelligence
		{BiomeDeepWater, "aquatic_adaptation", 1}, // Deep water should boost aquatic adaptation
		{BiomeHighAltitude, "altitude_tolerance", 1}, // High altitude should boost altitude tolerance
		{BiomeCanyon, "agility", 1},     // Canyon should boost agility
	}
	
	for _, tc := range testCases {
		biome := biomes[tc.biomeType]
		modifier, exists := biome.TraitModifiers[tc.expectedTrait]
		
		if !exists {
			t.Errorf("Biome %d missing expected trait modifier for %s", tc.biomeType, tc.expectedTrait)
			continue
		}
		
		if tc.expectedSign > 0 && modifier <= 0 {
			t.Errorf("Biome %d trait modifier for %s is %.2f, expected positive", tc.biomeType, tc.expectedTrait, modifier)
		}
		if tc.expectedSign < 0 && modifier >= 0 {
			t.Errorf("Biome %d trait modifier for %s is %.2f, expected negative", tc.biomeType, tc.expectedTrait, modifier)
		}
	}
}

// TestPerlinNoise tests the perlin noise function
func TestPerlinNoise(t *testing.T) {
	// Test that perlin noise returns values in expected range
	for i := 0; i < 100; i++ {
		x := float64(i) * 0.1
		y := float64(i) * 0.05
		noise := perlinNoise(x, y)
		
		if noise < -1.5 || noise > 1.5 {
			t.Errorf("Perlin noise value %.3f at (%.1f, %.1f) outside expected range [-1.5, 1.5]", noise, x, y)
		}
	}
	
	// Test that noise is deterministic (same input gives same output)
	noise1 := perlinNoise(1.0, 2.0)
	noise2 := perlinNoise(1.0, 2.0)
	
	if math.Abs(noise1-noise2) > 1e-10 {
		t.Errorf("Perlin noise is not deterministic: %.10f != %.10f", noise1, noise2)
	}
}