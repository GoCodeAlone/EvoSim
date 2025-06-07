package main

import (
	"testing"
)

func TestBiomeSymbolsFix(t *testing.T) {
	// Create a simple world and view manager to test biome symbols
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 0, // No populations for this test
		PopulationSize: 0,
		GridWidth:      10,
		GridHeight:     10,
	}

	world := NewWorld(config)
	viewManager := NewViewManager(world)

	// Test all biome types and their symbols
	biomeTypes := []BiomeType{
		BiomePlains, BiomeForest, BiomeDesert, BiomeMountain, BiomeWater, BiomeRadiation,
		BiomeSoil, BiomeAir, BiomeIce, BiomeRainforest, BiomeDeepWater, BiomeHighAltitude,
		BiomeHotSpring, BiomeTundra, BiomeSwamp, BiomeCanyon,
	}

	t.Logf("Testing all biome symbols:")
	for i, biomeType := range biomeTypes {
		name, symbol, color := viewManager.getBiomeInfo(biomeType)
		t.Logf("Biome %d (%s): symbol='%s', color='%s'", i, name, symbol, color)

		// Verify no question marks
		if symbol == "?" {
			t.Errorf("ERROR: Biome type %d (%s) still shows '?' symbol!", int(biomeType), name)
		}

		// Verify name is not "Unknown"
		if name == "Unknown" {
			t.Errorf("ERROR: Biome type %d shows 'Unknown' name!", int(biomeType))
		}
	}

	t.Log("SUCCESS: All biome symbols are properly defined!")
}
