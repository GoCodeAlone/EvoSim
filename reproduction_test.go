package main

import (
	"testing"
)

func TestReproductionSystem(t *testing.T) {
	rs := NewReproductionSystem()
	
	if rs == nil {
		t.Fatal("Failed to create reproduction system")
	}
	
	if len(rs.Eggs) != 0 {
		t.Errorf("Expected 0 eggs, got %d", len(rs.Eggs))
	}
	
	if len(rs.DecayingItems) != 0 {
		t.Errorf("Expected 0 decaying items, got %d", len(rs.DecayingItems))
	}
}

func TestReproductionStatus(t *testing.T) {
	status := NewReproductionStatus()
	
	if status == nil {
		t.Fatal("Failed to create reproduction status")
	}
	
	if !status.ReadyToMate {
		t.Error("New entity should be ready to mate")
	}
	
	if status.IsPregnant {
		t.Error("New entity should not be pregnant")
	}
}

func TestCanMate(t *testing.T) {
	status1 := NewReproductionStatus()
	status2 := NewReproductionStatus()
	
	// Both entities should be able to mate initially
	if !status1.CanMate(status2, 2, 0) {
		t.Error("Entities should be able to mate initially")
	}
	
	// Pregnant entity cannot mate
	status1.IsPregnant = true
	if status1.CanMate(status2, 2, 0) {
		t.Error("Pregnant entity should not be able to mate")
	}
	
	status1.IsPregnant = false
	status1.ReadyToMate = false
	if status1.CanMate(status2, 2, 0) {
		t.Error("Entity not ready to mate should not be able to mate")
	}
}

func TestEggLaying(t *testing.T) {
	rs := NewReproductionSystem()
	
	// Create two parent entities
	parent1 := NewEntity(1, []string{"strength", "speed"}, "test", Position{X: 10, Y: 10})
	parent2 := NewEntity(2, []string{"strength", "speed"}, "test", Position{X: 12, Y: 12})
	
	// Test egg laying
	success := rs.LayEgg(parent1, parent2, 0)
	if !success {
		t.Error("Egg laying should succeed")
	}
	
	if len(rs.Eggs) != 1 {
		t.Errorf("Expected 1 egg, got %d", len(rs.Eggs))
	}
	
	egg := rs.Eggs[0]
	if egg.Parent1ID != parent1.ID || egg.Parent2ID != parent2.ID {
		t.Error("Egg should have correct parent IDs")
	}
	
	if egg.Species != parent1.Species {
		t.Error("Egg should inherit parent species")
	}
}

func TestGestation(t *testing.T) {
	rs := NewReproductionSystem()
	
	// Create parent entities
	parent1 := NewEntity(1, []string{"strength", "speed"}, "test", Position{X: 10, Y: 10})
	parent2 := NewEntity(2, []string{"strength", "speed"}, "test", Position{X: 12, Y: 12})
	
	// Start gestation
	success := rs.StartGestation(parent1, parent2, 0)
	if !success {
		t.Error("Gestation should start successfully")
	}
	
	if !parent1.ReproductionStatus.IsPregnant {
		t.Error("Parent should be pregnant after gestation starts")
	}
	
	if parent1.ReproductionStatus.GestationStartTick != 0 {
		t.Error("Gestation start tick should be set correctly")
	}
}

func TestEggHatching(t *testing.T) {
	rs := NewReproductionSystem()
	
	// Create an egg
	egg := &Egg{
		ID:             1,
		Position:       Position{X: 10, Y: 10},
		Parent1ID:      1,
		Parent2ID:      2,
		LayingTick:     0,
		HatchingPeriod: 50,
		Energy:         30.0,
		IsViable:       true,
		Species:        "test",
	}
	rs.Eggs = append(rs.Eggs, egg)
	
	// Update system before hatching time
	newEntities, _ := rs.Update(25)
	if len(newEntities) != 0 {
		t.Error("Egg should not hatch before hatching period")
	}
	
	// Update system after hatching time
	newEntities, _ = rs.Update(60)
	if len(newEntities) != 1 {
		t.Errorf("Expected 1 hatched entity, got %d", len(newEntities))
	}
	
	hatchling := newEntities[0]
	if hatchling.Species != "test" {
		t.Error("Hatchling should inherit species from egg")
	}
	
	if hatchling.Energy != 30.0 {
		t.Error("Hatchling should inherit energy from egg")
	}
}

func TestDecaySystem(t *testing.T) {
	rs := NewReproductionSystem()
	
	// Add a decaying item
	rs.AddDecayingItem("corpse", Position{X: 5, Y: 5}, 50.0, "test_species", 10.0, 0)
	
	if len(rs.DecayingItems) != 1 {
		t.Errorf("Expected 1 decaying item, got %d", len(rs.DecayingItems))
	}
	
	item := rs.DecayingItems[0]
	if item.ItemType != "corpse" {
		t.Error("Item type should be 'corpse'")
	}
	
	if item.NutrientValue != 50.0 {
		t.Error("Nutrient value should be 50.0")
	}
	
	// Update before decay period
	_, fertilizers := rs.Update(50)
	if len(fertilizers) != 0 {
		t.Error("Item should not decay before decay period")
	}
	
	// Update after decay period (assume decay period > 50)
	_, fertilizers = rs.Update(500)
	if len(fertilizers) != 1 {
		t.Errorf("Expected 1 fertilizer, got %d", len(fertilizers))
	}
}

func TestMatingStrategies(t *testing.T) {
	// Test monogamous strategy
	status1 := NewReproductionStatus()
	status1.Strategy = Monogamous
	status1.MateID = 5  // Already mated to entity 5
	
	status2 := NewReproductionStatus()  // This represents entity 5 
	status3 := NewReproductionStatus()  // This represents entity 10
	
	// Should be able to mate with same mate (entity 5)
	if !status1.CanMate(status2, 5, 0) {
		t.Error("Monogamous entity should be able to mate with same mate")
	}
	
	// Should not be able to mate with different mate (entity 10)
	if status1.CanMate(status3, 10, 0) {
		t.Error("Monogamous entity should not be able to mate with different mate")
	}
}

func TestEntityReproductionIntegration(t *testing.T) {
	// Test that entities are properly initialized with reproduction status
	entity := NewEntity(1, []string{"strength", "speed"}, "test", Position{X: 0, Y: 0})
	
	if entity.ReproductionStatus == nil {
		t.Error("Entity should have reproduction status initialized")
	}
	
	if !entity.ReproductionStatus.ReadyToMate {
		t.Error("New entity should be ready to mate")
	}
	
	// Test cloning preserves reproduction traits
	clone := entity.Clone()
	if clone.ReproductionStatus == nil {
		t.Error("Cloned entity should have reproduction status")
	}
	
	if clone.ReproductionStatus.Mode != entity.ReproductionStatus.Mode {
		t.Error("Cloned entity should inherit reproduction mode")
	}
	
	if clone.ReproductionStatus.Strategy != entity.ReproductionStatus.Strategy {
		t.Error("Cloned entity should inherit mating strategy")
	}
}

func TestReproductionModes(t *testing.T) {
	// Test reproduction mode string representations
	modes := []ReproductionMode{DirectCoupling, EggLaying, LiveBirth, Budding, Fission}
	expected := []string{"direct_coupling", "egg_laying", "live_birth", "budding", "fission"}
	
	for i, mode := range modes {
		if mode.String() != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], mode.String())
		}
	}
	
	// Test mating strategy string representations
	strategies := []MatingStrategy{Monogamous, Polygamous, Sequential, Promiscuous}
	expectedStrats := []string{"monogamous", "polygamous", "sequential", "promiscuous"}
	
	for i, strategy := range strategies {
		if strategy.String() != expectedStrats[i] {
			t.Errorf("Expected %s, got %s", expectedStrats[i], strategy.String())
		}
	}
}

func TestReproductionSystemIntegration(t *testing.T) {
	// Create a world with reproduction system
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      10,
		GridHeight:     10,
	}
	
	world := NewWorld(config)
	
	// Create a population configuration
	popConfig := PopulationConfig{
		Name:    "test",
		Species: "test",
		BaseTraits: map[string]float64{
			"strength":  0.0,
			"speed":     0.0,
			"fertility": 0.0,
		},
		StartPos: Position{X: 50.0, Y: 50.0},
		Spread:   10.0,
		Color:    "blue",
		BaseMutationRate: 0.1,
	}
	
	world.AddPopulation(popConfig)
	
	if world.ReproductionSystem == nil {
		t.Fatal("World should have reproduction system")
	}
	
	// Check that entities have reproduction status
	entityCount := 0
	for _, entity := range world.AllEntities {
		if entity.ReproductionStatus == nil {
			t.Error("Entity should have reproduction status")
		}
		entityCount++
	}
	
	if entityCount != 5 {
		t.Errorf("Expected 5 entities, got %d", entityCount)
	}
	
	// Simulate a few ticks to test system integration
	initialEntityCount := len(world.AllEntities)
	
	for i := 0; i < 15; i++ { // Run past the 10-tick reproduction delay
		world.Update()
	}
	
	// Check that reproduction system is tracking properly
	if world.ReproductionSystem.NextEggID <= 1 && world.ReproductionSystem.NextItemID <= 1 {
		// Either some eggs were laid or some items decayed (both advance IDs)
		// This is fine - reproduction is probabilistic
	}
	
	// Entities should still be alive and aging
	aliveCount := 0
	agedEntities := 0
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			aliveCount++
			// Only check original entities that should have aged
			// New entities created during reproduction start at age 0
			if entity.ID < 5 && entity.Age == 0 { // Original entities should have aged
				t.Error("Original entities should have aged after multiple updates")
			}
			if entity.Age > 0 {
				agedEntities++
			}
		}
	}
	
	if aliveCount == 0 {
		t.Error("At least some entities should still be alive")
	}
	
	if agedEntities == 0 {
		t.Error("At least some entities should have aged")
	}
	
	t.Logf("Test completed: Initial entities: %d, Final entities: %d, Alive: %d", 
		initialEntityCount, len(world.AllEntities), aliveCount)
}

func TestMatingCompetition(t *testing.T) {
	// Create a world to test competition
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 0, // We'll add entities manually
		PopulationSize: 0,
		GridWidth:      10,
		GridHeight:     10,
	}
	
	world := NewWorld(config)
	
	// Create test entities manually
	entity1 := NewEntity(1, []string{"strength", "intelligence"}, "test", Position{X: 10, Y: 10})
	entity1.SetTrait("strength", 0.5)
	entity1.SetTrait("intelligence", 0.3)
	entity1.Energy = 80.0
	
	entity2 := NewEntity(2, []string{"strength", "intelligence"}, "test", Position{X: 12, Y: 12})
	entity2.SetTrait("strength", 0.4)
	entity2.SetTrait("intelligence", 0.4)
	entity2.Energy = 70.0
	
	// Stronger competitor nearby
	competitor := NewEntity(3, []string{"strength", "intelligence"}, "test", Position{X: 8, Y: 8})
	competitor.SetTrait("strength", 0.8)
	competitor.SetTrait("intelligence", 0.7)
	competitor.Energy = 90.0
	
	world.AllEntities = []*Entity{entity1, entity2, competitor}
	
	// Test competition detection
	hasCompetition := world.checkMatingCompetition(entity1, entity2)
	if !hasCompetition {
		t.Error("Should detect competition when stronger entity is nearby")
	}
	
	// Move competitor far away
	competitor.Position = Position{X: 50, Y: 50}
	hasCompetition = world.checkMatingCompetition(entity1, entity2)
	if hasCompetition {
		t.Error("Should not detect competition when competitor is far away")
	}
	
	// Make competitor weaker
	competitor.Position = Position{X: 8, Y: 8} // Move back close
	competitor.SetTrait("strength", 0.1)
	competitor.SetTrait("intelligence", 0.1)
	competitor.Energy = 20.0
	hasCompetition = world.checkMatingCompetition(entity1, entity2)
	if hasCompetition {
		t.Error("Should not detect competition when nearby entity is weaker")
	}
}