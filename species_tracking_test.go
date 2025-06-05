package main

import (
	"testing"
)

func TestSpeciesTrackingIssue(t *testing.T) {
	// Create a test world with some plants
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      20,
		GridHeight:     20,
	}

	world := NewWorld(config)

	// Add some test plants manually
	for i := 0; i < 30; i++ {
		x := float64(i % 10)
		y := float64(i / 10)
		plantType := PlantType(i % int(PlantCactus+1)) // Cycle through plant types
		plant := NewPlant(i, plantType, Position{X: x, Y: y})
		world.AllPlants = append(world.AllPlants, plant)
		world.NextPlantID = i + 1
	}

	// Initialize view manager
	vm := NewViewManager(world)

	// Let simulation run for several ticks to trigger speciation updates
	for tick := 0; tick < 100; tick++ {
		world.Update()
	}

	// Get species and evolution data
	viewData := vm.GetCurrentViewData()
	speciesData := viewData.Species
	evolutionData := viewData.Evolution

	t.Logf("Species Data: Active=%d, Extinct=%d, Details=%d",
		speciesData.ActiveSpecies, speciesData.ExtinctSpecies, len(speciesData.SpeciesDetails))
	t.Logf("Evolution Data: Speciation=%d, Extinction=%d, Diversity=%.3f",
		evolutionData.SpeciationEvents, evolutionData.ExtinctionEvents, evolutionData.GeneticDiversity)

	// Check plant count
	t.Logf("Plant Count: %d", len(world.AllPlants))

	// Check speciation system state
	if world.SpeciationSystem != nil {
		t.Logf("SpeciationSystem Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))
		t.Logf("SpeciationSystem All Species: %d", len(world.SpeciationSystem.AllSpecies))
		t.Logf("SpeciationSystem Speciation Events: %d", len(world.SpeciationSystem.SpeciationEvents))
		t.Logf("SpeciationSystem Extinction Events: %d", len(world.SpeciationSystem.ExtinctionEvents))

		// Log details of each active species
		for id, species := range world.SpeciationSystem.ActiveSpecies {
			t.Logf("Species %d: Name=%s, Members=%d, Type=%d",
				id, species.Name, len(species.Members), species.OriginPlantType)
		}
	} else {
		t.Errorf("SpeciationSystem is nil")
	}

	// The issue seems to be that SpeciationSystem shows 0 active species but there are plants
	// This suggests the initial species assignment isn't happening
	if len(world.AllPlants) > 0 && speciesData.ActiveSpecies == 0 {
		t.Errorf("Expected active species when plants exist. Plants: %d, Active Species: %d",
			len(world.AllPlants), speciesData.ActiveSpecies)
	}
}

func TestSpeciationSystemInitialization(t *testing.T) {
	// Test if the SpeciationSystem properly handles initial plants
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		NumPopulations: 1,
		PopulationSize: 3,
		GridWidth:      10,
		GridHeight:     10,
	}

	world := NewWorld(config)

	// Add some test plants manually
	for i := 0; i < 10; i++ {
		x := float64(i % 5)
		y := float64(i / 5)
		plant := NewPlant(i, PlantGrass, Position{X: x, Y: y})
		world.AllPlants = append(world.AllPlants, plant)
		world.NextPlantID = i + 1
	}

	t.Logf("Initial Plants: %d", len(world.AllPlants))

	// Check that SpeciationSystem exists
	if world.SpeciationSystem == nil {
		t.Fatalf("SpeciationSystem is nil after world creation")
	}

	// Initially should have no species
	t.Logf("Initial Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))

	// Update once to trigger species assignment
	world.Update() // This won't trigger SpeciationSystem.Update due to tick % 20 check

	// Manually call SpeciationSystem update to see what happens
	world.SpeciationSystem.Update(world.AllPlants, world.Tick)

	t.Logf("After manual update - Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))
	t.Logf("After manual update - All Species: %d", len(world.SpeciationSystem.AllSpecies))

	// There should be species now if plants exist
	if len(world.AllPlants) > 0 {
		if len(world.SpeciationSystem.ActiveSpecies) == 0 {
			t.Errorf("Expected active species after manual update. Plants: %d, Active Species: %d",
				len(world.AllPlants), len(world.SpeciationSystem.ActiveSpecies))
		}
	}

	// Log plant traits to understand genetic diversity
	if len(world.AllPlants) > 0 {
		firstPlant := world.AllPlants[0]
		t.Logf("First plant type: %d, traits: %v", firstPlant.Type, firstPlant.Traits)
	}
}

func TestSpeciationSystemPlantAssignment(t *testing.T) {
	// Test the plant assignment logic directly
	ss := NewSpeciationSystem()

	// Create some test plants with different traits
	plants := make([]*Plant, 0)
	for i := 0; i < 5; i++ {
		plant := &Plant{
			ID:      i,
			Type:    PlantGrass,
			IsAlive: true,
			Traits:  make(map[string]Trait),
		}
		plant.Traits["growth_rate"] = Trait{Value: float64(i) * 0.1}
		plant.Traits["height"] = Trait{Value: float64(i) * 0.2}
		plants = append(plants, plant)
	}

	t.Logf("Created %d test plants", len(plants))

	// Call Update to assign plants to species
	ss.Update(plants, 1)

	t.Logf("After update - Active Species: %d", len(ss.ActiveSpecies))
	t.Logf("After update - All Species: %d", len(ss.AllSpecies))

	// Should have at least one species
	if len(ss.ActiveSpecies) == 0 {
		t.Errorf("Expected at least one active species after update")
	}

	// Check species details
	for id, species := range ss.ActiveSpecies {
		t.Logf("Species %d: Name=%s, Members=%d", id, species.Name, len(species.Members))
	}
}
