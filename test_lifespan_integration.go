package main

import (
	"fmt"
	"os"
	"strings"
)

// testLifespanIntegration creates a simple test to verify the new lifespan system works
func testLifespanIntegration() {
	fmt.Println("Testing new organism classification and lifespan system...")
	
	// Create a minimal world
	config := WorldConfig{
		Width:        100,
		Height:       100,
		GridWidth:    10,
		GridHeight:   10,
		NumPopulations: 1,
	}
	
	world := NewWorld(config)
	
	// Create entities with different trait profiles to test classification
	testEntities := []*Entity{
		// Prokaryotic-like (low complexity)
		createTestEntityForLifespan(world, "prokaryotic", map[string]float64{
			"intelligence": -0.8,
			"size": -0.6,
			"cooperation": -0.7,
			"endurance": 0.0,
		}),
		// Eukaryotic-like (simple single-cell)
		createTestEntityForLifespan(world, "eukaryotic", map[string]float64{
			"intelligence": -0.3,
			"size": -0.2,
			"cooperation": -0.4,
			"endurance": 0.2,
		}),
		// Simple multicellular
		createTestEntityForLifespan(world, "simple_multi", map[string]float64{
			"intelligence": 0.0,
			"size": 0.1,
			"cooperation": 0.2,
			"endurance": 0.3,
		}),
		// Complex multicellular
		createTestEntityForLifespan(world, "complex_multi", map[string]float64{
			"intelligence": 0.4,
			"size": 0.5,
			"cooperation": 0.6,
			"endurance": 0.5,
		}),
		// Advanced multicellular
		createTestEntityForLifespan(world, "advanced_multi", map[string]float64{
			"intelligence": 0.8,
			"size": 0.7,
			"cooperation": 0.9,
			"endurance": 0.8,
		}),
	}
	
	// Add entities to world
	world.AllEntities = testEntities
	
	// Classify entities and show their lifespans
	fmt.Printf("\n%-20s %-25s %-10s %-12s %-10s %-12s\n", 
		"Entity Type", "Classification", "Lifespan", "Days", "Aging Rate", "Max Age")
	fmt.Println(strings.Repeat("-", 95))
	
	for _, entity := range testEntities {
		classification := world.OrganismClassifier.ClassifyEntity(entity, world.CellularSystem)
		lifespan := world.OrganismClassifier.CalculateLifespan(entity, classification)
		agingRate := world.OrganismClassifier.CalculateAgingRate(entity, classification)
		days := float64(lifespan) / float64(world.AdvancedTimeSystem.DayLength)
		
		entity.Classification = classification
		entity.MaxLifespan = lifespan
		
		classificationName := world.OrganismClassifier.GetClassificationName(classification)
		
		fmt.Printf("%-20s %-25s %-10d %-12.1f %-10.2f %-12d\n",
			entity.Species, classificationName, lifespan, days, agingRate, entity.MaxLifespan)
	}
	
	// Run simulation for several days to test aging
	fmt.Printf("\nRunning simulation for 5 simulated days (%d ticks)...\n", 
		5 * world.AdvancedTimeSystem.DayLength)
	
	initialCounts := make(map[string]int)
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			initialCounts[entity.Species]++
		}
	}
	
	// Track aging progress
	ageCheckpoints := []int{480, 960, 1440, 1920, 2400} // 1-5 days
	checkpointIndex := 0
	
	for tick := 0; tick < 5 * world.AdvancedTimeSystem.DayLength; tick++ {
		world.Tick = tick
		
		// Update entities
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				entity.UpdateWithClassification(world.OrganismClassifier, world.CellularSystem)
			}
		}
		
		// Check at day boundaries
		if checkpointIndex < len(ageCheckpoints) && tick >= ageCheckpoints[checkpointIndex] {
			day := (checkpointIndex + 1)
			fmt.Printf("\nDay %d status:\n", day)
			
			for _, entity := range world.AllEntities {
				if entity.IsAlive || entity.Age > 0 {
					status := "ALIVE"
					if !entity.IsAlive {
						status = "DEAD"
					}
					
					fmt.Printf("  %-20s: Age %4d/%4d (%.1f%%) - %s\n", 
						entity.Species, entity.Age, entity.MaxLifespan, 
						float64(entity.Age)/float64(entity.MaxLifespan)*100, status)
				}
			}
			checkpointIndex++
		}
	}
	
	// Final summary
	fmt.Printf("\nFinal Results (after 5 days):\n")
	finalCounts := make(map[string]int)
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			finalCounts[entity.Species]++
		}
	}
	
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "Entity Type", "Initial", "Final", "Survival")
	fmt.Println(strings.Repeat("-", 50))
	
	for entityType, initial := range initialCounts {
		final := finalCounts[entityType]
		survivalRate := float64(final) / float64(initial) * 100
		fmt.Printf("%-20s %-10d %-10d %-9.1f%%\n", entityType, initial, final, survivalRate)
	}
	
	fmt.Println("\nTest completed successfully!")
}

func createTestEntityForLifespan(world *World, species string, traits map[string]float64) *Entity {
	traitNames := make([]string, 0, len(traits))
	for name := range traits {
		traitNames = append(traitNames, name)
	}
	
	entity := NewEntity(world.NextID, traitNames, species, Position{X: 50, Y: 50})
	world.NextID++
	
	// Set specific trait values
	for name, value := range traits {
		entity.SetTrait(name, value)
	}
	
	// Give entity sufficient energy to survive the test
	entity.Energy = 1000.0
	
	return entity
}

// main function for manual testing (rename to main to run)
func mainTestLifespan() {
	// Add missing import
	fmt.Println("Starting lifespan integration test...")
	
	testLifespanIntegration()
	
	os.Exit(0)
}