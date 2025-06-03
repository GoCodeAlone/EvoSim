package main

import (
	"math"
	"testing"
)

func TestHiveMindSystem(t *testing.T) {
	// Create test entities
	entities := make([]*Entity, 0)
	traitNames := []string{"intelligence", "cooperation", "strength", "speed"}
	
	for i := 0; i < 5; i++ {
		entity := NewEntity(i, traitNames, "test_species", Position{X: float64(i * 2), Y: 0})
		entity.SetTrait("intelligence", 0.5 + float64(i)*0.1)
		entity.SetTrait("cooperation", 0.6 + float64(i)*0.05)
		entities = append(entities, entity)
	}

	// Create hive mind system
	hms := NewHiveMindSystem()
	
	// Try to form hive mind
	hiveMind := hms.TryFormHiveMind(entities, SimpleCollective)
	
	if hiveMind == nil {
		t.Fatal("Failed to create hive mind from compatible entities")
	}

	if len(hiveMind.Members) != 5 {
		t.Errorf("Expected 5 members, got %d", len(hiveMind.Members))
	}

	if hiveMind.Type != SimpleCollective {
		t.Errorf("Expected SimpleCollective type, got %v", hiveMind.Type)
	}

	// Test collective decision making
	decision := hiveMind.GetCollectiveDecision("food_search", []string{"aggressive_search", "conservative_search"})
	if decision == "" {
		t.Error("Expected decision but got empty string")
	}

	// Test knowledge sharing
	testPos := Position{X: 10, Y: 5}
	hiveMind.ShareKnowledge("food", testPos, 0.8)
	
	if len(hiveMind.CollectiveMemory.FoodSources) != 1 {
		t.Error("Expected food source to be shared in collective memory")
	}

	// Test coordinated movement
	hiveMind.CoordinateMovement(20, 10)
	
	// Verify members moved toward target formation
	moved := false
	for _, member := range hiveMind.Members {
		// Check if any member moved (simple check)
		if member.Position.X != float64(member.ID*2) || member.Position.Y != 0 {
			moved = true
			break
		}
	}

	if !moved {
		t.Error("Expected hive members to move in coordinated fashion")
	}

	// Test hive mind update
	hms.Update()
	
	if len(hms.HiveMinds) != 1 {
		t.Errorf("Expected 1 hive mind after update, got %d", len(hms.HiveMinds))
	}
}

func TestHiveMindMemoryDecay(t *testing.T) {
	entity := NewEntity(1, []string{"intelligence", "cooperation"}, "test", Position{})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("cooperation", 0.7)
	
	hiveMind := NewHiveMind(1, entity, SimpleCollective)
	
	// Add some memories
	testPos1 := Position{X: 5, Y: 5}
	testPos2 := Position{X: 10, Y: 10}
	
	hiveMind.ShareKnowledge("food", testPos1, 1.0)
	hiveMind.ShareKnowledge("threat", testPos2, 0.8)
	
	if len(hiveMind.CollectiveMemory.FoodSources) != 1 {
		t.Error("Expected 1 food source")
	}
	if len(hiveMind.CollectiveMemory.ThreatAreas) != 1 {
		t.Error("Expected 1 threat area")
	}

	// Simulate memory decay over many ticks
	for i := 0; i < 100; i++ {
		hiveMind.decayMemories()
	}

	// Memories should be significantly decayed or removed
	foundFood := false
	for _, strength := range hiveMind.CollectiveMemory.FoodSources {
		if strength > 0.1 {
			foundFood = true
			break
		}
	}

	if foundFood {
		t.Error("Expected food memories to decay significantly after many ticks")
	}
}

func TestHiveMindCompatibility(t *testing.T) {
	// Create hive mind with one member
	founder := NewEntity(1, []string{"intelligence", "cooperation"}, "test", Position{})
	founder.SetTrait("intelligence", 0.8)
	founder.SetTrait("cooperation", 0.7)
	
	hiveMind := NewHiveMind(1, founder, SimpleCollective)

	// Test compatible entity
	compatible := NewEntity(2, []string{"intelligence", "cooperation"}, "test", Position{})
	compatible.SetTrait("intelligence", 0.7) // Close to founder
	compatible.SetTrait("cooperation", 0.6)  // Close to founder

	if !hiveMind.CanJoinHive(compatible) {
		t.Error("Compatible entity should be able to join hive")
	}

	// Test incompatible entity (low intelligence)
	incompatible1 := NewEntity(3, []string{"intelligence", "cooperation"}, "test", Position{})
	incompatible1.SetTrait("intelligence", 0.1) // Too low
	incompatible1.SetTrait("cooperation", 0.7)

	if hiveMind.CanJoinHive(incompatible1) {
		t.Error("Entity with low intelligence should not be able to join hive")
	}

	// Test incompatible entity (low cooperation)
	incompatible2 := NewEntity(4, []string{"intelligence", "cooperation"}, "test", Position{})
	incompatible2.SetTrait("intelligence", 0.8)
	incompatible2.SetTrait("cooperation", 0.2) // Too low

	if hiveMind.CanJoinHive(incompatible2) {
		t.Error("Entity with low cooperation should not be able to join hive")
	}

	// Test incompatible entity (too different traits)
	incompatible3 := NewEntity(5, []string{"intelligence", "cooperation"}, "test", Position{})
	incompatible3.SetTrait("intelligence", 0.2) // Very different from founder
	incompatible3.SetTrait("cooperation", 0.9)

	if hiveMind.CanJoinHive(incompatible3) {
		t.Error("Entity with very different traits should not be able to join hive")
	}
}

func TestHiveMindTypes(t *testing.T) {
	// Test different hive mind types have appropriate max members
	types := []HiveMindType{SimpleCollective, SwarmIntelligence, NeuralNetwork, QuantumMind}
	expectedMaxMembers := []int{10, 50, 25, 100}

	for i, hiveType := range types {
		entity := NewEntity(1, []string{"intelligence", "cooperation"}, "test", Position{})
		entity.SetTrait("intelligence", 0.8)
		entity.SetTrait("cooperation", 0.7)
		
		hiveMind := NewHiveMind(1, entity, hiveType)
		
		if hiveMind.MaxMembers != expectedMaxMembers[i] {
			t.Errorf("Expected max members %d for type %v, got %d", 
				expectedMaxMembers[i], hiveType, hiveMind.MaxMembers)
		}
	}
}

func TestCollectiveIntelligence(t *testing.T) {
	// Create entities with varying intelligence
	entities := make([]*Entity, 3)
	traitNames := []string{"intelligence", "cooperation"}
	
	entities[0] = NewEntity(1, traitNames, "test", Position{})
	entities[0].SetTrait("intelligence", 0.5)
	entities[0].SetTrait("cooperation", 0.8)
	
	entities[1] = NewEntity(2, traitNames, "test", Position{})
	entities[1].SetTrait("intelligence", 0.7)
	entities[1].SetTrait("cooperation", 0.9)
	
	entities[2] = NewEntity(3, traitNames, "test", Position{})
	entities[2].SetTrait("intelligence", 0.6)
	entities[2].SetTrait("cooperation", 0.7)

	hiveMind := NewHiveMind(1, entities[0], SimpleCollective)
	hiveMind.AddMember(entities[1])
	hiveMind.AddMember(entities[2])

	// Test that collective intelligence is sum of member intelligence
	expectedIntelligence := 0.5 + 0.7 + 0.6 // 1.8 (simple addition in current implementation)

	if math.Abs(hiveMind.Intelligence - expectedIntelligence) > 0.01 {
		t.Errorf("Expected collective intelligence %.2f, got %.2f", 
			expectedIntelligence, hiveMind.Intelligence)
	}
}

func TestHiveMindSafetyCheck(t *testing.T) {
	entity := NewEntity(1, []string{"intelligence", "cooperation"}, "test", Position{})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("cooperation", 0.7)
	
	hiveMind := NewHiveMind(1, entity, SimpleCollective)
	
	// Add threat and safe zones
	threatPos := Position{X: 5, Y: 5}
	safePos := Position{X: 15, Y: 15}
	
	hiveMind.ShareKnowledge("threat", threatPos, 0.8)
	hiveMind.ShareKnowledge("safe", safePos, 0.9)

	// Test position near threat
	nearThreat := Position{X: 7, Y: 7}
	if hiveMind.IsPositionSafe(nearThreat) {
		t.Error("Position near threat should not be considered safe")
	}

	// Test position near safe zone
	nearSafe := Position{X: 16, Y: 16}
	if !hiveMind.IsPositionSafe(nearSafe) {
		t.Error("Position near safe zone should be considered safe")
	}

	// Test neutral position
	neutral := Position{X: 50, Y: 50}
	if !hiveMind.IsPositionSafe(neutral) {
		t.Error("Neutral position should be considered safe")
	}
}