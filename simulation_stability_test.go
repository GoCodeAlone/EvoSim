package main

import (
	"testing"
)

// TestSimulationStability tests that the simulation runs for 20 ticks without mass extinctions
func TestSimulationStability(t *testing.T) {
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 3,  // Multiple populations
		PopulationSize: 15, // Reasonable population size
		GridWidth:      20,
		GridHeight:     20,
	}

	world := NewWorld(config)
	
	// Add diverse populations with different adaptations
	popConfigs := []PopulationConfig{
		{
			Name:     "Generalists",
			Species:  "Adaptable",
			BaseTraits: map[string]float64{
				"endurance":  0.7,
				"intelligence": 0.6,
				"cooperation": 0.5,
				"size":       0.4,
				"speed":      0.5,
				"vision":     0.6,
				"aquatic_adaptation": 0.3,
				"altitude_tolerance": 0.3,
			},
			StartPos: Position{X: 30, Y: 30},
			Spread:   15.0,
			Color:    "blue",
			BaseMutationRate: 0.1,
		},
		{
			Name:     "AquaticSpecialists",
			Species:  "Aquatic",
			BaseTraits: map[string]float64{
				"endurance":  0.6,
				"intelligence": 0.4,
				"cooperation": 0.6,
				"size":       0.3,
				"speed":      0.4,
				"vision":     0.5,
				"aquatic_adaptation": 0.9,
				"altitude_tolerance": 0.1,
			},
			StartPos: Position{X: 70, Y: 30},
			Spread:   15.0,
			Color:    "cyan",
			BaseMutationRate: 0.1,
		},
		{
			Name:     "MountainAdapted",
			Species:  "Highland",
			BaseTraits: map[string]float64{
				"endurance":  0.8,
				"intelligence": 0.5,
				"cooperation": 0.4,
				"size":       0.5,
				"speed":      0.3,
				"vision":     0.7,
				"aquatic_adaptation": 0.1,
				"altitude_tolerance": 0.9,
			},
			StartPos: Position{X: 50, Y: 70},
			Spread:   15.0,
			Color:    "brown",
			BaseMutationRate: 0.1,
		},
	}
	
	for _, popConfig := range popConfigs {
		world.AddPopulation(popConfig)
	}
	
	initialEntityCount := len(world.AllEntities)
	t.Logf("Starting simulation with %d entities across %d populations", initialEntityCount, len(popConfigs))
	
	// Check biome distribution
	biomeCount := make(map[BiomeType]int)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			biomeCount[biome]++
		}
	}
	t.Logf("Biome distribution: %v", biomeCount)
	
	// Track population over 20 ticks
	survivalData := make([]int, 21)
	populationData := make(map[string][]int)
	
	for _, popConfig := range popConfigs {
		populationData[popConfig.Species] = make([]int, 21)
	}
	
	for tick := 0; tick <= 20; tick++ {
		// Count total alive entities
		aliveCount := 0
		speciesCounts := make(map[string]int)
		
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				aliveCount++
				if _, exists := speciesCounts[entity.Species]; !exists {
					speciesCounts[entity.Species] = 0
				}
				speciesCounts[entity.Species]++
			}
		}
		
		survivalData[tick] = aliveCount
		// Store counts for all detected species
		for species, count := range speciesCounts {
			if _, exists := populationData[species]; !exists {
				populationData[species] = make([]int, 21)
			}
			populationData[species][tick] = count
		}
		
		if tick%5 == 0 {
			t.Logf("Tick %d: %d alive (%.1f%%) | Species: %v", 
				tick, aliveCount, float64(aliveCount)/float64(initialEntityCount)*100, speciesCounts)
		}
		
		// Update world if not last tick
		if tick < 20 {
			world.Update()
		}
	}
	
	// Analysis
	finalSurvival := float64(survivalData[20]) / float64(initialEntityCount)
	midSurvival := float64(survivalData[10]) / float64(initialEntityCount)
	
	t.Logf("Final survival rate (tick 20): %.1f%%", finalSurvival*100)
	t.Logf("Mid survival rate (tick 10): %.1f%%", midSurvival*100)
	
	// Check that no mass extinction occurred
	if finalSurvival < 0.3 {
		t.Errorf("Mass extinction detected: only %.1f%% survived to tick 20", finalSurvival*100)
	}
	
	// Check that we don't have rapid early die-off
	if midSurvival < 0.6 {
		t.Errorf("Rapid early mortality: only %.1f%% survived to tick 10", midSurvival*100)
	}
	
	// Check that at least 2 species survive
	survivingSpecies := 0
	for species, counts := range populationData {
		if counts[20] > 0 {
			survivingSpecies++
			t.Logf("Species %s: survived with %d entities", species, counts[20])
		}
	}
	
	if survivingSpecies < 2 {
		t.Errorf("Species diversity lost: only %d species survived", survivingSpecies)
	}
	
	t.Logf("SUCCESS: Simulation completed 20 ticks with %.1f%% survival and %d species", 
		finalSurvival*100, survivingSpecies)
}