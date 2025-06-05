package main

import (
	"os"
	"testing"
)

func TestStateManagerSaveLoad(t *testing.T) {
	// Create a test world
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      10,
		GridHeight:     10,
	}

	// Create original world
	world1 := NewWorld(config)

	// Add a small population for testing
	popConfig := PopulationConfig{
		Name:    "TestSpecies",
		Species: "test",
		BaseTraits: map[string]float64{
			"size":     0.5,
			"speed":    0.3,
			"strength": 0.7,
		},
		StartPos:         Position{X: 25, Y: 25},
		Spread:           5.0,
		Color:            "green",
		BaseMutationRate: 0.1,
	}
	world1.AddPopulation(popConfig)

	// Advance the world a few ticks to create some state
	for i := 0; i < 5; i++ {
		world1.Update()
	}

	originalTick := world1.Tick
	originalEntityCount := len(world1.AllEntities)
	originalPlantCount := len(world1.AllPlants)

	// Save state
	stateManager1 := NewStateManager(world1)
	filename := "test_state_temp.json"
	defer os.Remove(filename) // Clean up

	err := stateManager1.SaveToFile(filename)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create a new world
	world2 := NewWorld(config)
	stateManager2 := NewStateManager(world2)

	// Load state
	err = stateManager2.LoadFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify state was restored correctly
	if world2.Tick != originalTick {
		t.Errorf("Tick mismatch: expected %d, got %d", originalTick, world2.Tick)
	}

	if len(world2.AllEntities) != originalEntityCount {
		t.Errorf("Entity count mismatch: expected %d, got %d", originalEntityCount, len(world2.AllEntities))
	}

	if len(world2.AllPlants) != originalPlantCount {
		t.Errorf("Plant count mismatch: expected %d, got %d", originalPlantCount, len(world2.AllPlants))
	}

	// Verify that we can continue running the simulation
	world2.Update()

	if world2.Tick != originalTick+1 {
		t.Errorf("World didn't advance correctly after load: expected %d, got %d", originalTick+1, world2.Tick)
	}
}

func TestStateManagerEmptyWorld(t *testing.T) {
	// Create an empty world
	config := WorldConfig{
		Width:          20.0,
		Height:         20.0,
		NumPopulations: 0,
		PopulationSize: 0,
		GridWidth:      5,
		GridHeight:     5,
	}

	world := NewWorld(config)
	stateManager := NewStateManager(world)

	filename := "test_empty_state.json"
	defer os.Remove(filename)

	// Test saving empty world
	err := stateManager.SaveToFile(filename)
	if err != nil {
		t.Fatalf("Failed to save empty world state: %v", err)
	}

	// Create new world and load
	world2 := NewWorld(config)
	stateManager2 := NewStateManager(world2)

	err = stateManager2.LoadFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to load empty world state: %v", err)
	}

	// Verify empty state (entities should be empty, but plants may be auto-generated)
	if len(world2.AllEntities) != 0 {
		t.Errorf("Expected no entities, got %d", len(world2.AllEntities))
	}

	// Note: Plants may be auto-generated in the world initialization, so we don't test for zero plants
}
