package main

import (
	"testing"
)

func TestNewEntity(t *testing.T) {
	traitNames := []string{"strength", "agility", "intelligence"}
	pos := Position{X: 10, Y: 20}
	entity := NewEntity(1, traitNames, "test", pos)

	if entity.ID != 1 {
		t.Errorf("Expected entity ID to be 1, got %d", entity.ID)
	}

	if len(entity.Traits) != len(traitNames) {
		t.Errorf("Expected %d traits, got %d", len(traitNames), len(entity.Traits))
	}

	for _, name := range traitNames {
		if _, exists := entity.Traits[name]; !exists {
			t.Errorf("Expected trait %s to exist", name)
		}
	}

	if entity.Fitness != 0.0 {
		t.Errorf("Expected initial fitness to be 0.0, got %f", entity.Fitness)
	}

	if entity.Species != "test" {
		t.Errorf("Expected species to be 'test', got %s", entity.Species)
	}

	if entity.Position.X != 10 || entity.Position.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%.1f, %.1f)", entity.Position.X, entity.Position.Y)
	}
}

func TestEntityGetTrait(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)
	entity.SetTrait("strength", 0.5)

	value := entity.GetTrait("strength")
	if value != 0.5 {
		t.Errorf("Expected trait value 0.5, got %f", value)
	}

	// Test non-existent trait
	value = entity.GetTrait("nonexistent")
	if value != 0.0 {
		t.Errorf("Expected 0.0 for non-existent trait, got %f", value)
	}
}

func TestEntitySetTrait(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{}, "test", pos)
	entity.SetTrait("newTrait", 0.75)

	if len(entity.Traits) != 1 {
		t.Errorf("Expected 1 trait after setting, got %d", len(entity.Traits))
	}

	value := entity.GetTrait("newTrait")
	if value != 0.75 {
		t.Errorf("Expected trait value 0.75, got %f", value)
	}
}

func TestEntityMutate(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)
	originalValue := entity.GetTrait("strength")

	// Test with 100% mutation rate
	entity.Mutate(1.0, 0.1)

	newValue := entity.GetTrait("strength")
	// Value should have changed (with very high probability)
	// We can't guarantee it will always change due to randomness, but it's extremely likely
	if newValue == originalValue {
		t.Logf("Warning: Mutation may not have occurred (original: %f, new: %f)", originalValue, newValue)
	}

	// Test with 0% mutation rate
	entity.SetTrait("strength", 0.5)
	entity.Mutate(0.0, 0.1)

	finalValue := entity.GetTrait("strength")
	if finalValue != 0.5 {
		t.Errorf("Expected no mutation with 0%% rate, value changed from 0.5 to %f", finalValue)
	}
}

func TestEntityClone(t *testing.T) {
	pos := Position{X: 5, Y: 10}
	original := NewEntity(1, []string{"strength", "agility"}, "test", pos)
	original.SetTrait("strength", 0.5)
	original.SetTrait("agility", 0.3)
	original.Fitness = 1.5

	clone := original.Clone()

	if clone.ID != original.ID {
		t.Errorf("Expected clone ID %d, got %d", original.ID, clone.ID)
	}

	if clone.Fitness != original.Fitness {
		t.Errorf("Expected clone fitness %f, got %f", original.Fitness, clone.Fitness)
	}

	if len(clone.Traits) != len(original.Traits) {
		t.Errorf("Expected clone to have %d traits, got %d", len(original.Traits), len(clone.Traits))
	}

	// Test that it's a deep copy
	clone.SetTrait("strength", 0.9)
	if original.GetTrait("strength") == 0.9 {
		t.Error("Clone modification affected original entity")
	}
}

func TestCrossover(t *testing.T) {
	pos1 := Position{X: 0, Y: 0}
	pos2 := Position{X: 10, Y: 10}

	parent1 := NewEntity(1, []string{"strength", "agility"}, "test", pos1)
	parent1.SetTrait("strength", 0.8)
	parent1.SetTrait("agility", 0.2)

	parent2 := NewEntity(2, []string{"strength", "intelligence"}, "test", pos2)
	parent2.SetTrait("strength", 0.3)
	parent2.SetTrait("intelligence", 0.7)

	child := Crossover(parent1, parent2, 3, "test")

	if child.ID != 3 {
		t.Errorf("Expected child ID 3, got %d", child.ID)
	}

	// Child should have all traits from both parents
	expectedTraits := []string{"strength", "agility", "intelligence"}
	for _, trait := range expectedTraits {
		if _, exists := child.Traits[trait]; !exists {
			t.Errorf("Expected child to have trait %s", trait)
		}
	}

	// Child fitness should be 0 initially
	if child.Fitness != 0.0 {
		t.Errorf("Expected child fitness to be 0.0, got %f", child.Fitness)
	}
}
