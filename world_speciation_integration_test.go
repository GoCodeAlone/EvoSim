package main

import (
	"testing"
)

func TestWorldInitialSpeciesCreation(t *testing.T) {
	// Test that species are created immediately when world is created with plants
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		NumPopulations: 1,
		PopulationSize: 3,
		GridWidth:      20,
		GridHeight:     20,
	}

	world := NewWorld(config)

	t.Logf("Initial Plants: %d", len(world.AllPlants))
	t.Logf("Initial Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))

	// Now we should have species created immediately
	if len(world.AllPlants) > 0 {
		if len(world.SpeciationSystem.ActiveSpecies) == 0 {
			t.Errorf("Expected active species immediately after world creation. Plants: %d, Active Species: %d", 
				len(world.AllPlants), len(world.SpeciationSystem.ActiveSpecies))
		}
	}

	// Test ViewManager with immediate data
	vm := NewViewManager(world)
	viewData := vm.GetCurrentViewData()

	t.Logf("Initial ViewManager Data:")
	t.Logf("  Active Species: %d", viewData.Species.ActiveSpecies)
	t.Logf("  Species With Members: %d", viewData.Species.SpeciesWithMembers)
	t.Logf("  Speciation Detected: %t", viewData.Evolution.SpeciationDetected)
	t.Logf("  Active Plant Count: %d", viewData.Evolution.ActivePlantCount)

	// Verify that we have species with members
	if len(world.AllPlants) > 0 {
		if viewData.Species.SpeciesWithMembers == 0 {
			t.Errorf("Expected species with members immediately after world creation")
		}
		if !viewData.Evolution.SpeciationDetected {
			t.Errorf("Expected speciation to be detected immediately")
		}
	}
}

func TestWorldSpeciesTrackingAfterUpdate(t *testing.T) {
	// Test that species tracking works correctly during normal updates
	config := WorldConfig{
		Width:          30.0,
		Height:         30.0,
		NumPopulations: 1,
		PopulationSize: 3,
		GridWidth:      15,
		GridHeight:     15,
	}

	world := NewWorld(config)

	// Should have immediate species
	initialActiveSpecies := len(world.SpeciationSystem.ActiveSpecies)
	initialPlants := len(world.AllPlants)
	
	t.Logf("After world creation:")
	t.Logf("  Plants: %d, Active Species: %d", initialPlants, initialActiveSpecies)

	// Run several updates to see how the system behaves
	for tick := 1; tick <= 50; tick++ {
		world.Update()
		
		if tick%10 == 0 {
			vm := NewViewManager(world)
			viewData := vm.GetCurrentViewData()
			
			t.Logf("Tick %d: Plants=%d, Active Species=%d, With Members=%d, Awaiting Extinction=%d",
				tick, len(world.AllPlants), viewData.Species.ActiveSpecies, 
				viewData.Species.SpeciesWithMembers, viewData.Species.SpeciesAwaitingExtinction)
		}
	}
	
	// Final check
	vm := NewViewManager(world)
	viewData := vm.GetCurrentViewData()
	
	t.Logf("Final state:")
	t.Logf("  Plants: %d", len(world.AllPlants))
	t.Logf("  Active Species: %d", viewData.Species.ActiveSpecies)
	t.Logf("  Species With Members: %d", viewData.Species.SpeciesWithMembers)
	t.Logf("  Species Awaiting Extinction: %d", viewData.Species.SpeciesAwaitingExtinction)
	t.Logf("  Extinction Events: %d", viewData.Evolution.ExtinctionEvents)
	t.Logf("  Speciation Detected: %t", viewData.Evolution.SpeciationDetected)
}