package main

import (
	"testing"
)

func TestToolSystem(t *testing.T) {
	toolSystem := NewToolSystem(NewCentralEventBus(1000))
	
	// Create a test entity
	entity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Energy:   100.0,
		Position: Position{X: 0, Y: 0},
	}
	entity.Traits["intelligence"] = Trait{Name: "intelligence", Value: 0.5}
	entity.Traits["strength"] = Trait{Name: "strength", Value: 0.6}
	
	// Test tool creation
	tool := toolSystem.CreateTool(entity, ToolStone, entity.Position)
	if tool == nil {
		t.Error("Failed to create stone tool")
	}
	
	if tool.Type != ToolStone {
		t.Errorf("Expected tool type %d, got %d", ToolStone, tool.Type)
	}
	
	if tool.Owner != entity {
		t.Error("Tool owner not set correctly")
	}
	
	// Test tool usage
	effectiveness := toolSystem.UseTool(tool, entity, 1.0)
	if effectiveness <= 0 {
		t.Error("Tool usage should return positive effectiveness")
	}
	
	if tool.Durability >= tool.MaxDurability {
		t.Error("Tool durability should decrease after use")
	}
}

func TestEnvironmentalModificationSystem(t *testing.T) {
	modSystem := NewEnvironmentalModificationSystem(NewCentralEventBus(1000))
	
	// Create a test entity
	entity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Energy:   100.0,
		Position: Position{X: 5, Y: 5},
	}
	entity.Traits["intelligence"] = Trait{Name: "intelligence", Value: 0.6}
	entity.Traits["strength"] = Trait{Name: "strength", Value: 0.7}
	
	// Test burrow creation
	burrow := modSystem.CreateBurrow(entity, entity.Position)
	if burrow == nil {
		t.Error("Failed to create burrow")
	}
	
	if burrow.Type != EnvModBurrow {
		t.Errorf("Expected modification type %d, got %d", EnvModBurrow, burrow.Type)
	}
	
	if burrow.Creator != entity {
		t.Error("Burrow creator not set correctly")
	}
	
	// Test using the burrow
	benefit := modSystem.UseModification(burrow, entity, 1)
	if benefit <= 0 {
		t.Error("Using burrow should provide positive benefit")
	}
	
	// Test cache creation
	cache := modSystem.CreateCache(entity, Position{X: 6, Y: 6})
	if cache == nil {
		t.Error("Failed to create cache")
	}
	
	if cache.Type != EnvModCache {
		t.Errorf("Expected modification type %d, got %d", EnvModCache, cache.Type)
	}
}

func TestEmergentBehaviorSystem(t *testing.T) {
	behaviorSystem := NewEmergentBehaviorSystem()
	
	// Create a test entity
	entity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Energy:   100.0,
		Position: Position{X: 10, Y: 10},
		IsAlive:  true,
	}
	entity.Traits["intelligence"] = Trait{Name: "intelligence", Value: 0.8}
	entity.Traits["curiosity"] = Trait{Name: "curiosity", Value: 0.7}
	entity.Traits["cooperation"] = Trait{Name: "cooperation", Value: 0.6}
	entity.Traits["strength"] = Trait{Name: "strength", Value: 0.5}
	entity.Traits["aggression"] = Trait{Name: "aggression", Value: 0.3}
	entity.Traits["speed"] = Trait{Name: "speed", Value: 0.4}
	
	// Initialize behavior pattern
	behaviorSystem.InitializeEntityBehavior(entity)
	
	pattern, exists := behaviorSystem.BehaviorPatterns[entity.ID]
	if !exists {
		t.Error("Behavior pattern not created")
	}
	
	if pattern.EntityID != entity.ID {
		t.Error("Behavior pattern entity ID mismatch")
	}
	
	if pattern.LearningRate <= 0 {
		t.Error("Learning rate should be positive")
	}
	
	if pattern.Curiosity <= 0 {
		t.Error("Curiosity should be positive")
	}
	
	// Check that tool preferences were initialized
	if len(pattern.ToolPreferences) == 0 {
		t.Error("Tool preferences should be initialized")
	}
	
	// Check basic behaviors are available for discovery
	if len(behaviorSystem.LearnedBehaviors) == 0 {
		t.Error("No learned behaviors available")
	}
	
	// Verify specific behaviors exist
	expectedBehaviors := []string{"tool_making", "tunnel_digging", "cache_hiding", "trap_setting"}
	for _, behaviorName := range expectedBehaviors {
		if _, exists := behaviorSystem.LearnedBehaviors[behaviorName]; !exists {
			t.Errorf("Expected behavior %s not found", behaviorName)
		}
	}
}

func TestIntegratedToolAndBehaviorSystem(t *testing.T) {
	// Create a small world for integration testing
	config := WorldConfig{
		Width:          20.0,
		Height:         20.0,
		NumPopulations: 1,
		PopulationSize: 3,
		GridWidth:      10,
		GridHeight:     10,
	}
	
	world := NewWorld(config)
	
	// Add a test population
	popConfig := PopulationConfig{
		Name:    "TestSpecies",
		Species: "test",
		BaseTraits: map[string]float64{
			"intelligence": 0.7,
			"strength":     0.6,
			"curiosity":    0.8,
			"cooperation":  0.5,
		},
		StartPos:         Position{X: 10, Y: 10},
		Spread:           2.0,
		Color:            "green",
		BaseMutationRate: 0.1,
	}
	world.AddPopulation(popConfig)
	
	// Verify systems are initialized
	if world.ToolSystem == nil {
		t.Error("ToolSystem not initialized")
	}
	
	if world.EnvironmentalModSystem == nil {
		t.Error("EnvironmentalModSystem not initialized")
	}
	
	if world.EmergentBehaviorSystem == nil {
		t.Error("EmergentBehaviorSystem not initialized")
	}
	
	// Run a few simulation steps
	for i := 0; i < 5; i++ {
		world.Update()
	}
	
	// Check that entities have behavior patterns
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			_, exists := world.EmergentBehaviorSystem.BehaviorPatterns[entity.ID]
			if !exists {
				t.Errorf("Entity %d should have behavior pattern", entity.ID)
			}
		}
	}
	
	// Verify world stats include new systems
	toolStats := world.ToolSystem.GetToolStats()
	if toolStats == nil {
		t.Error("Tool stats should be available")
	}
	
	modStats := world.EnvironmentalModSystem.GetModificationStats()
	if modStats == nil {
		t.Error("Environmental modification stats should be available")
	}
	
	behaviorStats := world.EmergentBehaviorSystem.GetBehaviorStats()
	if behaviorStats == nil {
		t.Error("Behavior stats should be available")
	}
}