package main

import (
	"testing"
)

func TestSpeedMultiplierMethods(t *testing.T) {
	// Create a simple world config
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	
	// Test default speed
	if world.GetSpeedMultiplier() != 1.0 {
		t.Errorf("Expected default speed 1.0, got %f", world.GetSpeedMultiplier())
	}
	
	// Test setting speed
	world.SetSpeedMultiplier(2.0)
	if world.GetSpeedMultiplier() != 2.0 {
		t.Errorf("Expected speed 2.0, got %f", world.GetSpeedMultiplier())
	}
	
	// Test increase speed
	world.SetSpeedMultiplier(1.0)
	world.IncreaseSpeed()
	if world.GetSpeedMultiplier() != 2.0 {
		t.Errorf("Expected speed 2.0 after increase, got %f", world.GetSpeedMultiplier())
	}
	
	// Test decrease speed
	world.DecreaseSpeed()
	if world.GetSpeedMultiplier() != 1.0 {
		t.Errorf("Expected speed 1.0 after decrease, got %f", world.GetSpeedMultiplier())
	}
	
	// Test bounds - minimum
	world.SetSpeedMultiplier(0.01)
	if world.GetSpeedMultiplier() != 0.1 {
		t.Errorf("Expected minimum speed 0.1, got %f", world.GetSpeedMultiplier())
	}
	
	// Test bounds - maximum
	world.SetSpeedMultiplier(100.0)
	if world.GetSpeedMultiplier() != 16.0 {
		t.Errorf("Expected maximum speed 16.0, got %f", world.GetSpeedMultiplier())
	}
}

func TestSpeedSequence(t *testing.T) {
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	
	// Test the speed sequence 
	expectedSpeeds := []float64{1.0, 2.0, 4.0, 8.0, 16.0, 16.0} // 16.0 should stay at max
	
	for i, expected := range expectedSpeeds {
		if i > 0 {
			world.IncreaseSpeed()
		}
		if world.GetSpeedMultiplier() != expected {
			t.Errorf("Step %d: Expected speed %f, got %f", i, expected, world.GetSpeedMultiplier())
		}
	}
	
	// Test decrease sequence
	expectedSpeeds = []float64{16.0, 8.0, 4.0, 2.0, 1.0, 0.5, 0.25, 0.25} // 0.25 should stay at min
	world.SetSpeedMultiplier(16.0) // Reset to max
	
	for i, expected := range expectedSpeeds {
		if i > 0 {
			world.DecreaseSpeed()
		}
		if world.GetSpeedMultiplier() != expected {
			t.Errorf("Decrease step %d: Expected speed %f, got %f", i, expected, world.GetSpeedMultiplier())
		}
	}
}