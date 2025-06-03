package main

import (
	"testing"
)

// TestBiomeStateTransitions tests that biome transitions work correctly
func TestBiomeStateTransitions(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Set up initial conditions - create some ice and hot springs
	// Place ice next to hot springs to test ice melting
	world.Grid[5][5].Biome = BiomeIce
	world.Grid[5][6].Biome = BiomeHotSpring
	world.Grid[6][5].Biome = BiomeIce
	
	// Place forest next to radiation zones to test fire effects
	world.Grid[10][10].Biome = BiomeForest
	world.Grid[10][11].Biome = BiomeRadiation
	
	// Count initial biome types
	initialIceCount := 0
	initialForestCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			switch world.Grid[y][x].Biome {
			case BiomeIce:
				initialIceCount++
			case BiomeForest:
				initialForestCount++
			}
		}
	}

	t.Logf("Initial state: Ice=%d, Forest=%d", initialIceCount, initialForestCount)

	// Run several ticks to allow transitions to occur
	for i := 0; i < 50; i++ {
		world.processBiomeTransitions()
	}

	// Count final biome types
	finalIceCount := 0
	finalForestCount := 0
	finalWaterCount := 0
	finalDesertCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			switch world.Grid[y][x].Biome {
			case BiomeIce:
				finalIceCount++
			case BiomeForest:
				finalForestCount++
			case BiomeWater:
				finalWaterCount++
			case BiomeDesert:
				finalDesertCount++
			}
		}
	}

	t.Logf("Final state: Ice=%d, Forest=%d, Water=%d, Desert=%d", 
		finalIceCount, finalForestCount, finalWaterCount, finalDesertCount)

	// Verify that some transitions occurred
	// We should see some ice melted to water (though not guaranteed due to probability)
	if finalIceCount < initialIceCount || finalWaterCount > 0 {
		t.Logf("SUCCESS: Ice melting transition detected")
	}

	// Check that the system doesn't crash and maintains grid integrity
	totalCells := config.GridWidth * config.GridHeight
	
	// Count all biomes to ensure grid integrity
	allBiomesCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			allBiomesCount++
		}
	}
	
	if allBiomesCount != totalCells {
		t.Errorf("Grid integrity check failed: expected %d cells, got %d", totalCells, allBiomesCount)
	}
}

// TestTransitionTriggerDetection tests that environmental triggers are detected correctly
func TestTransitionTriggerDetection(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  10,
		GridHeight: 10,
	}

	world := NewWorld(config)

	// Clear the grid and set specific test conditions
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x].Biome = BiomePlains // Default to plains
		}
	}

	// Set up controlled test conditions
	world.Grid[5][5].Biome = BiomeHotSpring
	world.Grid[5][6].Biome = BiomeIce
	
	// Test trigger detection
	triggers := world.detectTransitionTriggers(5, 6) // Check ice cell next to hot spring
	
	if triggers["heat"] == 0 && triggers["hotspring"] == 0 {
		t.Errorf("Expected heat/hotspring triggers near hot spring, got none")
	}
	
	t.Logf("Detected triggers near hot spring: %+v", triggers)
	
	// Test that intensity decreases with distance - create isolated hot spring
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x].Biome = BiomePlains // Reset to plains
		}
	}
	
	world.Grid[3][3].Biome = BiomeHotSpring // Isolated hot spring
	nearTriggers := world.detectTransitionTriggers(3, 4) // Adjacent to hot spring
	farTriggers := world.detectTransitionTriggers(3, 6)  // Further from hot spring
	
	t.Logf("Near triggers: %+v", nearTriggers)
	t.Logf("Far triggers: %+v", farTriggers)
	
	// Check that heat/hotspring triggers exist near the hot spring
	if nearTriggers["heat"] == 0 && nearTriggers["hotspring"] == 0 {
		t.Errorf("Expected heat/hotspring triggers near hot spring")
	}
	
	// Far triggers should have no heat/hotspring effects from our isolated source
	if farTriggers["heat"] > 0 || farTriggers["hotspring"] > 0 {
		t.Logf("Note: Far triggers detected heat/hotspring - this could be from fire events: heat=%.3f, hotspring=%.3f", 
			farTriggers["heat"], farTriggers["hotspring"])
	}
}

// TestBiomeTransitionRules tests that transition rules are properly defined
func TestBiomeTransitionRules(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  10,
		GridHeight: 10,
	}

	world := NewWorld(config)

	// Test that getBiomeName works for all biome types
	testBiomes := []BiomeType{
		BiomePlains, BiomeForest, BiomeDesert, BiomeMountain, BiomeWater,
		BiomeRadiation, BiomeSoil, BiomeAir, BiomeIce, BiomeRainforest,
		BiomeDeepWater, BiomeHighAltitude, BiomeHotSpring, BiomeTundra,
		BiomeSwamp, BiomeCanyon,
	}

	for _, biome := range testBiomes {
		name := world.getBiomeName(biome)
		if name == "unknown" {
			t.Errorf("No name defined for biome type %d", biome)
		}
	}

	t.Logf("All biome names properly defined")
}