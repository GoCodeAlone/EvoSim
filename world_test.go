package main

import (
	"testing"
)

func TestNewWorld(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 2,
		PopulationSize: 10,
		GridWidth:      20,
		GridHeight:     15,
	}

	world := NewWorld(config)

	if world.Config.Width != 100.0 {
		t.Errorf("Expected width 100.0, got %f", world.Config.Width)
	}

	if world.Config.Height != 100.0 {
		t.Errorf("Expected height 100.0, got %f", world.Config.Height)
	}

	if len(world.Populations) != 0 {
		t.Errorf("Expected 0 populations initially, got %d", len(world.Populations))
	}

	if len(world.AllEntities) != 0 {
		t.Errorf("Expected 0 entities initially, got %d", len(world.AllEntities))
	}

	// Test grid initialization
	if len(world.Grid) != config.GridHeight {
		t.Errorf("Expected grid height %d, got %d", config.GridHeight, len(world.Grid))
	}

	if len(world.Grid[0]) != config.GridWidth {
		t.Errorf("Expected grid width %d, got %d", config.GridWidth, len(world.Grid[0]))
	}

	// Test biomes are initialized
	if len(world.Biomes) == 0 {
		t.Errorf("Expected biomes to be initialized, got 0")
	}

	if world.NextID != 0 {
		t.Errorf("Expected NextID to be 0, got %d", world.NextID)
	}
}

func TestWorldAddPopulation(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      20,
		GridHeight:     15,
	}

	world := NewWorld(config)

	popConfig := PopulationConfig{
		Name:    "TestPop",
		Species: "herbivore",
		BaseTraits: map[string]float64{
			"strength": 0.5,
			"speed":    0.3,
		},
		StartPos: Position{X: 50, Y: 50},
		Spread:   10.0,
		Color:    "blue",
	}

	world.AddPopulation(popConfig)

	if len(world.Populations) != 1 {
		t.Errorf("Expected 1 population, got %d", len(world.Populations))
	}

	if len(world.AllEntities) != 5 {
		t.Errorf("Expected 5 entities, got %d", len(world.AllEntities))
	}

	// Check that entities have generated species names (not the generic type)
	var generatedSpeciesName string
	for _, entity := range world.AllEntities {
		if generatedSpeciesName == "" {
			generatedSpeciesName = entity.Species
		}
		if entity.Species != generatedSpeciesName {
			t.Errorf("Expected all entities to have same species name '%s', got '%s'", generatedSpeciesName, entity.Species)
		}
		if entity.Species == "herbivore" {
			t.Errorf("Expected generated species name, got generic type '%s'", entity.Species)
		}

		// Check that entities have the base traits
		if entity.GetTrait("strength") == 0.0 {
			t.Errorf("Expected entity to have strength trait")
		}

		if entity.GetTrait("speed") == 0.0 {
			t.Errorf("Expected entity to have speed trait")
		}
	}
}

func TestWorldUpdate(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 1,
		PopulationSize: 3,
		GridWidth:      20,
		GridHeight:     15,
	}

	// Use custom simulation config that ensures energy decreases
	simConfig := DefaultSimulationConfig()
	simConfig.Energy.EnergyRegenerationRate = 0.0 // No regeneration for predictable testing
	
	world := NewWorldWithConfig(config, simConfig)

	popConfig := PopulationConfig{
		Name:    "TestPop",
		Species: "herbivore",
		BaseTraits: map[string]float64{
			"strength": 0.5,
		},
		StartPos: Position{X: 50, Y: 50},
		Spread:   5.0,
		Color:    "blue",
	}

	world.AddPopulation(popConfig)

	initialTick := world.Tick
	initialEntityCount := len(world.AllEntities)

	// Update the world
	world.Update()

	if world.Tick != initialTick+1 {
		t.Errorf("Expected tick to increment from %d to %d, got %d",
			initialTick, initialTick+1, world.Tick)
	}

	// Entities should age and lose energy
	for _, entity := range world.AllEntities {
		if entity.Age != 1 {
			t.Errorf("Expected entity age to be 1, got %d", entity.Age)
		}

		if entity.Energy >= 100.0 {
			t.Errorf("Expected entity energy to decrease from 100.0, got %f", entity.Energy)
		}
	}

	// Entity count might change due to deaths, but should be <= initial
	currentEntityCount := len(world.AllEntities)
	if currentEntityCount > initialEntityCount {
		t.Errorf("Entity count increased unexpectedly from %d to %d",
			initialEntityCount, currentEntityCount)
	}
}

func TestWorldGetStats(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 2,
		PopulationSize: 5,
		GridWidth:      20,
		GridHeight:     15,
	}

	world := NewWorld(config)

	// Add two different populations
	popConfig1 := PopulationConfig{
		Name:    "Pop1",
		Species: "herbivore",
		BaseTraits: map[string]float64{
			"strength": 0.5,
		},
		StartPos: Position{X: 25, Y: 25},
		Spread:   5.0,
		Color:    "red",
	}

	popConfig2 := PopulationConfig{
		Name:    "Pop2",
		Species: "predator",
		BaseTraits: map[string]float64{
			"strength": 0.8,
		},
		StartPos: Position{X: 75, Y: 75},
		Spread:   5.0,
		Color:    "blue",
	}

	world.AddPopulation(popConfig1)
	world.AddPopulation(popConfig2)

	stats := world.GetStats()

	if stats["tick"] != 0 {
		t.Errorf("Expected tick to be 0, got %v", stats["tick"])
	}

	if stats["total_entities"] != 10 {
		t.Errorf("Expected total_entities to be 10, got %v", stats["total_entities"])
	}

	populations, ok := stats["populations"].(map[string]map[string]interface{})
	if !ok {
		t.Errorf("Expected populations to be a map")
	}

	if len(populations) != 2 {
		t.Errorf("Expected 2 populations in stats, got %d", len(populations))
	}

	// Check that both species are present (should be generated names, not original types)
	speciesNames := make([]string, 0, len(populations))
	for name := range populations {
		speciesNames = append(speciesNames, name)
	}
	
	if len(speciesNames) != 2 {
		t.Errorf("Expected exactly 2 species in stats, got %d: %v", len(speciesNames), speciesNames)
	}
}

func TestEntityInteractions(t *testing.T) {
	// Test entity distance calculation
	pos1 := Position{X: 0, Y: 0}
	pos2 := Position{X: 3, Y: 4}

	entity1 := NewEntity(1, []string{"strength"}, "herbivore", pos1)
	entity2 := NewEntity(2, []string{"strength"}, "herbivore", pos2)

	distance := entity1.DistanceTo(entity2)
	expectedDistance := 5.0 // 3-4-5 triangle

	if distance != expectedDistance {
		t.Errorf("Expected distance %.1f, got %.1f", expectedDistance, distance)
	}
}

func TestEntityMovement(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "herbivore", pos)

	initialEnergy := entity.Energy

	// Move to position (3, 4)
	entity.MoveTo(3, 4, 1.0)

	// Check that entity moved
	if entity.Position.X == 0 && entity.Position.Y == 0 {
		t.Errorf("Entity did not move from origin")
	}

	// Check that energy decreased
	if entity.Energy >= initialEnergy {
		t.Errorf("Expected energy to decrease after movement, was %f, now %f",
			initialEnergy, entity.Energy)
	}
}

func TestEntityCombat(t *testing.T) {
	pos1 := Position{X: 0, Y: 0}
	pos2 := Position{X: 1, Y: 1}

	// Create a strong predator
	predator := NewEntity(1, []string{"aggression", "strength", "size"}, "predator", pos1)
	predator.SetTrait("aggression", 0.9)
	predator.SetTrait("strength", 0.8)
	predator.SetTrait("size", 0.7)

	// Create a weak prey
	prey := NewEntity(2, []string{"defense", "strength", "size"}, "herbivore", pos2)
	prey.SetTrait("defense", 0.1)
	prey.SetTrait("strength", 0.2)
	prey.SetTrait("size", -0.5)

	// Predator should be able to kill prey
	canKill := predator.CanKill(prey)
	if !canKill {
		t.Errorf("Expected predator to be able to kill prey")
	}

	// Perform the kill
	success := predator.Kill(prey)
	if !success {
		t.Errorf("Expected kill to succeed")
	}

	// Prey should be dead
	if prey.IsAlive {
		t.Errorf("Expected prey to be dead after being killed")
	}

	// Predator should be able to eat the dead prey
	canEat := predator.CanEat(prey)
	if !canEat {
		t.Errorf("Expected predator to be able to eat dead prey")
	}
}

func TestEntityMerging(t *testing.T) {
	pos1 := Position{X: 0, Y: 0}
	pos2 := Position{X: 1, Y: 1}

	// Create two compatible entities of same species
	entity1 := NewEntity(1, []string{"intelligence", "cooperation"}, "herbivore", pos1)
	entity1.SetTrait("intelligence", 0.6)
	entity1.SetTrait("cooperation", 0.5)
	entity1.Energy = 60

	entity2 := NewEntity(2, []string{"intelligence", "cooperation"}, "herbivore", pos2)
	entity2.SetTrait("intelligence", 0.7)
	entity2.SetTrait("cooperation", 0.4)
	entity2.Energy = 60

	// They should be able to merge
	canMerge := entity1.CanMerge(entity2)
	if !canMerge {
		t.Errorf("Expected entities to be able to merge")
	}

	// Perform the merge
	merged := entity1.Merge(entity2, 999)
	if merged == nil {
		t.Errorf("Expected merge to produce a new entity")
	}

	// Check merged entity properties
	if merged.ID != 999 {
		t.Errorf("Expected merged entity ID to be 999, got %d", merged.ID)
	}

	if merged.Species != "herbivore" {
		t.Errorf("Expected merged entity species to be 'herbivore', got '%s'", merged.Species)
	}

	// Original entities should be dead
	if entity1.IsAlive {
		t.Errorf("Expected entity1 to be dead after merge")
	}

	if entity2.IsAlive {
		t.Errorf("Expected entity2 to be dead after merge")
	}

	// Merged entity should have averaged traits
	mergedIntelligence := merged.GetTrait("intelligence")
	expectedIntelligence := (0.6 + 0.7) / 2.0
	tolerance := 0.2 // Allow some variation due to random mutation in merge

	if mergedIntelligence < expectedIntelligence-tolerance ||
		mergedIntelligence > expectedIntelligence+tolerance {
		t.Errorf("Expected merged intelligence around %.3f, got %.3f",
			expectedIntelligence, mergedIntelligence)
	}
}
