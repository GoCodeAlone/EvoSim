package main

import (
	"testing"
)

func TestInsectPollinationSystemCreation(t *testing.T) {
	system := NewInsectPollinationSystem()

	if system == nil {
		t.Fatal("Failed to create InsectPollinationSystem")
	}

	// Check initial state
	if len(system.FlowerPatches) != 0 {
		t.Errorf("Expected 0 initial flower patches, got %d", len(system.FlowerPatches))
	}

	if len(system.PollinationEvents) != 0 {
		t.Errorf("Expected 0 initial pollination events, got %d", len(system.PollinationEvents))
	}

	if system.SeasonalModifier != 1.0 {
		t.Errorf("Expected initial seasonal modifier of 1.0, got %f", system.SeasonalModifier)
	}
}

func TestPollinatorTraitsAddition(t *testing.T) {
	// Create a small flying insect-like entity
	entity := &Entity{
		ID:       1,
		IsAlive:  true,
		Position: Position{X: 0, Y: 0},
		Traits:   make(map[string]Trait),
	}

	// Set up insect-like characteristics
	entity.SetTrait("size", -0.3)
	entity.SetTrait("flying_ability", 0.5)
	entity.SetTrait("swarm_capability", 0.4)
	entity.SetTrait("intelligence", 0.6)

	// Add pollinator traits
	AddPollinatorTraitsToEntity(entity)

	// Check that pollinator traits were added
	if entity.GetTrait("pollination_efficiency") < 0.3 {
		t.Errorf("Expected pollination efficiency >= 0.3, got %f", entity.GetTrait("pollination_efficiency"))
	}

	if entity.GetTrait("nectar_detection") < 0.4 {
		t.Errorf("Expected nectar detection >= 0.4, got %f", entity.GetTrait("nectar_detection"))
	}

	if entity.GetTrait("flower_memory") <= 0 {
		t.Errorf("Expected flower memory > 0, got %f", entity.GetTrait("flower_memory"))
	}

	if entity.GetTrait("flight_range") < 10 {
		t.Errorf("Expected flight range >= 10, got %f", entity.GetTrait("flight_range"))
	}
}

func TestFlowerPatchCreation(t *testing.T) {
	system := NewInsectPollinationSystem()

	// Create a healthy mature plant
	plant := NewPlant(1, PlantTree, Position{X: 10, Y: 10})
	plant.Age = 20 // Mature enough to flower
	plant.Energy = 100

	// Create flower patch
	patch := system.CreateFlowerPatch(plant, Spring)

	if patch == nil {
		t.Fatal("Failed to create flower patch from healthy plant")
	}

	if patch.PlantID != plant.ID {
		t.Errorf("Expected flower patch plant ID %d, got %d", plant.ID, patch.PlantID)
	}

	if patch.PlantType != plant.Type {
		t.Errorf("Expected flower patch type %d, got %d", plant.Type, patch.PlantType)
	}

	if patch.NectarAmount <= 0 {
		t.Errorf("Expected positive nectar amount, got %f", patch.NectarAmount)
	}

	if patch.BloomingLevel <= 0 || patch.BloomingLevel > 1 {
		t.Errorf("Expected blooming level between 0-1, got %f", patch.BloomingLevel)
	}
}

func TestPollinatorBehaviorSystem(t *testing.T) {
	system := NewInsectPollinationSystem()

	// Create a pollinator entity
	pollinator := &Entity{
		ID:       1,
		IsAlive:  true,
		Position: Position{X: 0, Y: 0},
		Traits:   make(map[string]Trait),
		Energy:   50,
	}

	// Set up as pollinator
	pollinator.SetTrait("size", -0.3)
	pollinator.SetTrait("flying_ability", 0.6)
	pollinator.SetTrait("swarm_capability", 0.4)
	pollinator.SetTrait("intelligence", 0.5)
	AddPollinatorTraitsToEntity(pollinator)

	// Create a flowering plant
	plant := NewPlant(1, PlantBush, Position{X: 5, Y: 5})
	plant.Age = 15
	plant.Energy = 80

	entities := []*Entity{pollinator}
	plants := []*Plant{plant}

	// Update system for a few ticks
	for i := 0; i < 5; i++ {
		system.Update(entities, plants, Summer, i)
	}

	// Check that flower patches were created
	if len(system.FlowerPatches) == 0 {
		t.Errorf("Expected flower patches to be created, got %d", len(system.FlowerPatches))
	}

	// Check that pollinator memories might be created
	if len(system.PollinatorMemories) == 0 {
		// This is okay - memories are only created after successful pollination
		t.Logf("No pollinator memories created yet (expected for short test)")
	}
}

func TestSeasonalModifiers(t *testing.T) {
	system := NewInsectPollinationSystem()

	// Test different seasonal modifiers
	testCases := []struct {
		season   Season
		expected float64
	}{
		{Spring, 1.3},
		{Summer, 1.0},
		{Autumn, 0.6},
		{Winter, 0.2},
	}

	for _, tc := range testCases {
		system.updateSeasonalModifier(tc.season)
		if system.SeasonalModifier != tc.expected {
			t.Errorf("Expected seasonal modifier %.2f for season %d, got %.2f",
				tc.expected, tc.season, system.SeasonalModifier)
		}
	}
}

func TestPollinationStats(t *testing.T) {
	system := NewInsectPollinationSystem()

	stats := system.GetPollinationStats()

	// Check that all expected stats are present
	expectedKeys := []string{
		"active_flower_patches",
		"active_pollinators",
		"total_pollinations",
		"cross_species_pollinations",
		"cross_species_rate",
		"nectar_produced",
		"nectar_consumed",
		"recent_pollination_events",
		"seasonal_modifier",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stat '%s' not found in pollination stats", key)
		}
	}

	// Check initial values
	if stats["active_flower_patches"].(int) != 0 {
		t.Errorf("Expected 0 active flower patches initially, got %d", stats["active_flower_patches"].(int))
	}

	if stats["total_pollinations"].(int) != 0 {
		t.Errorf("Expected 0 total pollinations initially, got %d", stats["total_pollinations"].(int))
	}
}
