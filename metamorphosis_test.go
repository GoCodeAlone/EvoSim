package main

import (
	"testing"
)

func TestMetamorphosisSystemCreation(t *testing.T) {
	system := NewMetamorphosisSystem()

	if system == nil {
		t.Fatal("Expected metamorphosis system to be created")
	}

	// Check that stage requirements are initialized
	if len(system.StageMinDurations) == 0 {
		t.Error("Expected stage durations to be initialized")
	}

	if len(system.StageEnergyRequired) == 0 {
		t.Error("Expected stage energy requirements to be initialized")
	}

	if len(system.StageTraitModifiers) == 0 {
		t.Error("Expected stage trait modifiers to be initialized")
	}
}

func TestMetamorphosisStatusCreation(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create a basic entity
	entity := NewEntity(1, []string{"size", "intelligence", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})

	// Set traits that would trigger complete metamorphosis
	entity.SetTrait("size", -0.2)
	entity.SetTrait("swarm_capability", 0.5)
	entity.SetTrait("pollination_efficiency", 0.6)
	entity.SetTrait("intelligence", 0.4)

	status := NewMetamorphosisStatus(entity, system)

	if status == nil {
		t.Fatal("Expected metamorphosis status to be created")
	}

	if status.Type != CompleteMetamorphosis {
		t.Errorf("Expected CompleteMetamorphosis, got %v", status.Type)
	}

	if status.CurrentStage != StageEgg {
		t.Errorf("Expected to start in egg stage, got %v", status.CurrentStage)
	}

	if status.CanMove {
		t.Error("Expected eggs to be unable to move")
	}
}

func TestNoMetamorphosisEntities(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create an entity that doesn't undergo metamorphosis
	entity := NewEntity(1, []string{"size", "intelligence", "swarm_capability"}, "test_species", Position{X: 50, Y: 50})

	// Set traits that would not trigger metamorphosis
	entity.SetTrait("size", 0.5)             // Large size
	entity.SetTrait("swarm_capability", 0.1) // Low swarm capability
	entity.SetTrait("intelligence", 0.3)

	status := NewMetamorphosisStatus(entity, system)

	if status.Type != NoMetamorphosis {
		t.Errorf("Expected NoMetamorphosis, got %v", status.Type)
	}

	if status.CurrentStage != StageAdult {
		t.Errorf("Expected to start as adult, got %v", status.CurrentStage)
	}

	if !status.CanMove {
		t.Error("Expected adults to be able to move")
	}
}

func TestStageAdvancement(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create entity with complete metamorphosis
	entity := NewEntity(1, []string{"size", "intelligence", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})
	entity.SetTrait("size", -0.2)
	entity.SetTrait("swarm_capability", 0.5)
	entity.SetTrait("pollination_efficiency", 0.6)
	entity.SetTrait("intelligence", 0.4)

	status := NewMetamorphosisStatus(entity, system)
	entity.MetamorphosisStatus = status

	// Create favorable environment
	environment := map[string]float64{
		"temperature":        0.6,
		"humidity":           0.7,
		"food_availability":  0.8,
		"safety":             0.9,
		"population_density": 0.3,
	}

	// Give entity sufficient energy
	entity.Energy = 100.0

	// Skip minimum time requirement by advancing to future tick
	currentTick := system.StageMinDurations[StageEgg] + 10

	// Should advance from egg to larva
	stageChanged := system.Update(entity, currentTick, environment)

	if !stageChanged {
		t.Error("Expected stage to change from egg to larva")
	}

	if entity.MetamorphosisStatus.CurrentStage != StageLarva {
		t.Errorf("Expected larva stage, got %v", entity.MetamorphosisStatus.CurrentStage)
	}

	if !entity.MetamorphosisStatus.CanMove {
		t.Error("Expected larvae to be able to move")
	}
}

func TestPupaStageImmobility(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create entity and advance to pupa stage
	entity := NewEntity(1, []string{"size", "intelligence", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})
	entity.SetTrait("size", -0.2)
	entity.SetTrait("swarm_capability", 0.5)
	entity.SetTrait("pollination_efficiency", 0.6)
	entity.SetTrait("intelligence", 0.4)

	status := NewMetamorphosisStatus(entity, system)
	entity.MetamorphosisStatus = status

	// Manually set to pupa stage
	status.CurrentStage = StagePupa
	status.IsMetamorphosing = true // Manually set since we're not going through normal advancement
	system.updateStageCapabilities(entity, StagePupa)

	if status.CanMove {
		t.Error("Expected pupae to be unable to move")
	}

	if status.VulnerabilityModifier <= 1.0 {
		t.Error("Expected pupae to be more vulnerable")
	}

	if !status.IsMetamorphosing {
		t.Error("Expected pupae to be marked as metamorphosing")
	}
}

func TestTraitModification(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create entity
	entity := NewEntity(1, []string{"size", "speed", "reproduction_rate", "intelligence"}, "test_species", Position{X: 50, Y: 50})
	entity.SetTrait("size", 0.5)
	entity.SetTrait("speed", 0.8)
	entity.SetTrait("reproduction_rate", 0.6)
	entity.SetTrait("intelligence", 0.7)

	// Store original values
	originalSize := entity.GetTrait("size")
	originalSpeed := entity.GetTrait("speed")
	_ = entity.GetTrait("reproduction_rate") // Store but don't use to avoid compiler warning

	// Create metamorphosis status with larva stage
	status := &MetamorphosisStatus{
		Type:         CompleteMetamorphosis,
		CurrentStage: StageLarva,
		CanMove:      true,
	}
	entity.MetamorphosisStatus = status

	// Apply stage modifiers
	system.applyStageModifiers(entity)

	// Check that traits were modified according to larva stage
	newSize := entity.GetTrait("size")
	newSpeed := entity.GetTrait("speed")
	newReproduction := entity.GetTrait("reproduction_rate")

	if newSize >= originalSize {
		t.Error("Expected size to be reduced in larva stage")
	}

	if newSpeed >= originalSpeed {
		t.Error("Expected speed to be reduced in larva stage")
	}

	if newReproduction != 0.0 {
		t.Error("Expected reproduction rate to be 0 in larva stage")
	}

	// Check that original traits are stored
	if entity.OriginalTraits == nil {
		t.Error("Expected original traits to be stored")
	}

	if entity.OriginalTraits["size"] != originalSize {
		t.Error("Expected original size to be preserved")
	}
}

func TestEnvironmentalRequirements(t *testing.T) {
	system := NewMetamorphosisSystem()

	entity := NewEntity(1, []string{"size", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})
	entity.SetTrait("size", -0.2)
	entity.SetTrait("swarm_capability", 0.5)
	entity.SetTrait("pollination_efficiency", 0.6)

	status := NewMetamorphosisStatus(entity, system)
	entity.MetamorphosisStatus = status

	// Test egg stage environmental requirements
	goodEnvironment := map[string]float64{
		"temperature": 0.6,
		"humidity":    0.7,
	}

	badEnvironment := map[string]float64{
		"temperature": 0.1, // Too cold
		"humidity":    0.2, // Too dry
	}

	if !system.checkEnvironmentalTriggers(entity, goodEnvironment) {
		t.Error("Expected good environment to allow metamorphosis")
	}

	if system.checkEnvironmentalTriggers(entity, badEnvironment) {
		t.Error("Expected bad environment to prevent metamorphosis")
	}
}

func TestMetamorphosisStats(t *testing.T) {
	system := NewMetamorphosisSystem()

	// Create a mix of entities with different metamorphosis types and stages
	entities := make([]*Entity, 5)

	for i := 0; i < 5; i++ {
		entity := NewEntity(i+1, []string{"size", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})

		if i < 3 {
			// Complete metamorphosis entities
			entity.SetTrait("size", -0.2)
			entity.SetTrait("swarm_capability", 0.5)
			entity.SetTrait("pollination_efficiency", 0.6)
		} else {
			// No metamorphosis entities
			entity.SetTrait("size", 0.5)
			entity.SetTrait("swarm_capability", 0.1)
		}

		entity.MetamorphosisStatus = NewMetamorphosisStatus(entity, system)
		entities[i] = entity
	}

	// Set different stages for complete metamorphosis entities
	entities[0].MetamorphosisStatus.CurrentStage = StageLarva
	entities[1].MetamorphosisStatus.CurrentStage = StagePupa
	entities[1].MetamorphosisStatus.IsMetamorphosing = true
	entities[1].MetamorphosisStatus.PupalShelter = true
	entities[2].MetamorphosisStatus.CurrentStage = StageAdult

	stats := system.GetMetamorphosisStats(entities)

	if stats == nil {
		t.Fatal("Expected stats to be returned")
	}

	stageCounts := stats["stage_counts"].(map[LifeStage]int)

	if stageCounts[StageLarva] != 1 {
		t.Errorf("Expected 1 larva, got %d", stageCounts[StageLarva])
	}

	if stageCounts[StagePupa] != 1 {
		t.Errorf("Expected 1 pupa, got %d", stageCounts[StagePupa])
	}

	if stageCounts[StageAdult] != 3 { // 1 complete metamorphosis + 2 no metamorphosis
		t.Errorf("Expected 3 adults, got %d", stageCounts[StageAdult])
	}

	if stats["currently_metamorphosing"].(int) != 1 {
		t.Errorf("Expected 1 metamorphosing entity, got %d", stats["currently_metamorphosing"])
	}

	if stats["pupal_shelters"].(int) != 1 {
		t.Errorf("Expected 1 pupal shelter, got %d", stats["pupal_shelters"])
	}
}

func TestLifeStageString(t *testing.T) {
	stages := []LifeStage{StageEgg, StageLarva, StagePupa, StageAdult, StageElder}
	expected := []string{"egg", "larva", "pupa", "adult", "elder"}

	for i, stage := range stages {
		if stage.String() != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], stage.String())
		}
	}
}

func TestMetamorphosisTypeString(t *testing.T) {
	types := []MetamorphosisType{NoMetamorphosis, SimpleMetamorphosis, CompleteMetamorphosis, HolometabolousMetamorphosis}
	expected := []string{"none", "simple", "complete", "holometabolous"}

	for i, metamorphosisType := range types {
		if metamorphosisType.String() != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], metamorphosisType.String())
		}
	}
}

func TestStageDescription(t *testing.T) {
	system := NewMetamorphosisSystem()

	entity := NewEntity(1, []string{"size", "swarm_capability", "pollination_efficiency"}, "test_species", Position{X: 50, Y: 50})
	entity.SetTrait("size", -0.2)
	entity.SetTrait("swarm_capability", 0.5)
	entity.SetTrait("pollination_efficiency", 0.6)

	status := NewMetamorphosisStatus(entity, system)
	entity.MetamorphosisStatus = status

	description := system.GetStageDescription(entity)
	if description != "egg" {
		t.Errorf("Expected 'egg', got '%s'", description)
	}

	// Test metamorphosing stage
	status.CurrentStage = StagePupa
	status.IsMetamorphosing = true
	status.PupalShelter = true

	description = system.GetStageDescription(entity)
	expected := "pupa (metamorphosing) (sheltered)"
	if description != expected {
		t.Errorf("Expected '%s', got '%s'", expected, description)
	}
}
