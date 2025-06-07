package main

import (
	"testing"
)

func TestSpeciesExtinctionBehavior(t *testing.T) {
	// Test the extinction timer behavior
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
	for i := 0; i < 5; i++ {
		x := float64(i % 5)
		y := float64(i / 5)
		plant := NewPlant(i, PlantGrass, Position{X: x, Y: y})
		world.AllPlants = append(world.AllPlants, plant)
		world.NextPlantID = i + 1
	}

	t.Logf("Initial Plants: %d", len(world.AllPlants))

	// Trigger species creation by calling SpeciationSystem.Update manually
	world.SpeciationSystem.Update(world.AllPlants, 1)
	t.Logf("After speciation - Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))

	// Now remove all plants to simulate death
	world.AllPlants = make([]*Plant, 0)

	// Update species system with empty plant list
	world.SpeciationSystem.Update(world.AllPlants, 2)
	t.Logf("After plants removed - Active Species: %d", len(world.SpeciationSystem.ActiveSpecies))

	// Check species members
	for id, species := range world.SpeciationSystem.ActiveSpecies {
		t.Logf("Species %d: Members=%d, ExtinctionTick=%d", id, len(species.Members), species.ExtinctionTick)
	}

	// Update again after extinction threshold (100 ticks)
	world.SpeciationSystem.Update(world.AllPlants, 103) // 103 > 2 + 100
	t.Logf("After extinction threshold - Active Species: %d, All Species: %d",
		len(world.SpeciationSystem.ActiveSpecies), len(world.SpeciationSystem.AllSpecies))
	t.Logf("Extinction Events: %d", len(world.SpeciationSystem.ExtinctionEvents))

	// All species should now be extinct
	for id, species := range world.SpeciationSystem.AllSpecies {
		t.Logf("All Species %d: IsExtinct=%t", id, species.IsExtinct)
	}
}

func TestViewManagerWithEmptySpecies(t *testing.T) {
	// Test what ViewManager shows when species exist but have no members
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
	for i := 0; i < 5; i++ {
		plant := NewPlant(i, PlantGrass, Position{X: float64(i), Y: 0})
		world.AllPlants = append(world.AllPlants, plant)
		world.NextPlantID = i + 1
	}

	// Create species
	world.SpeciationSystem.Update(world.AllPlants, 1)

	// Remove plants but keep species (simulating recent death before extinction)
	world.AllPlants = make([]*Plant, 0)
	world.SpeciationSystem.Update(world.AllPlants, 2)

	// Test ViewManager output
	vm := NewViewManager(world)
	viewData := vm.GetCurrentViewData()

	t.Logf("ViewManager Species Data:")
	t.Logf("  Active Species: %d", viewData.Species.ActiveSpecies)
	t.Logf("  Extinct Species: %d", viewData.Species.ExtinctSpecies)
	t.Logf("  Species Details: %d", len(viewData.Species.SpeciesDetails))
	t.Logf("  Total Species Ever: %d", viewData.Species.TotalSpeciesEver)
	t.Logf("  Species With Members: %d", viewData.Species.SpeciesWithMembers)
	t.Logf("  Species Awaiting Extinction: %d", viewData.Species.SpeciesAwaitingExtinction)
	t.Logf("  Has Speciation System: %t", viewData.Species.HasSpeciationSystem)

	for i, detail := range viewData.Species.SpeciesDetails {
		t.Logf("  Species %d: ID=%d, Name=%s, Population=%d, IsExtinct=%t, AwaitingExtinction=%t",
			i, detail.ID, detail.Name, detail.Population, detail.IsExtinct, detail.AwaitingExtinction)
	}

	t.Logf("ViewManager Evolution Data:")
	t.Logf("  Speciation Events: %d", viewData.Evolution.SpeciationEvents)
	t.Logf("  Extinction Events: %d", viewData.Evolution.ExtinctionEvents)
	t.Logf("  Genetic Diversity: %.3f", viewData.Evolution.GeneticDiversity)
	t.Logf("  Has Speciation System: %t", viewData.Evolution.HasSpeciationSystem)
	t.Logf("  Total Plants Tracked: %d", viewData.Evolution.TotalPlantsTracked)
	t.Logf("  Active Plant Count: %d", viewData.Evolution.ActivePlantCount)
	t.Logf("  Speciation Detected: %t", viewData.Evolution.SpeciationDetected)
}
