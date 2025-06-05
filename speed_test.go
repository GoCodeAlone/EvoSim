package main

import (
	"math"
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
	world.SetSpeedMultiplier(16.0)                                        // Reset to max

	for i, expected := range expectedSpeeds {
		if i > 0 {
			world.DecreaseSpeed()
		}
		if world.GetSpeedMultiplier() != expected {
			t.Errorf("Decrease step %d: Expected speed %f, got %f", i, expected, world.GetSpeedMultiplier())
		}
	}
}

// TestSpeedMultiplierImpactOnEnergySystem tests that speed multiplier affects energy calculations
func TestSpeedMultiplierImpactOnEnergySystem(t *testing.T) {
	// Create world with default config
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
		PopulationSize: 1,
	}

	world := NewWorld(config)

	// Create an entity for testing
	entity := &Entity{
		ID:             1,
		Position:       Position{X: 25.0, Y: 25.0},
		Energy:         100.0,
		Age:            1,
		IsAlive:        true,
		Species:        "test",
		Traits:         make(map[string]Trait),
		Classification: ClassificationEukaryotic,
		MaxLifespan:    210, // 30 days * 7 scale factor
	}

	// Initialize traits
	entity.Traits["size"] = Trait{Name: "size", Value: 0.0}
	entity.Traits["endurance"] = Trait{Name: "endurance", Value: 0.0}

	// Test at normal speed (1x)
	world.SetSpeedMultiplier(1.0)
	initialEnergy := entity.Energy
	entity.UpdateWithConfig(world.SimConfig)
	energyChangeNormal := initialEnergy - entity.Energy

	// Reset entity
	entity.Energy = 100.0
	entity.Age = 1

	// Test at double speed (2x)
	world.SetSpeedMultiplier(2.0)
	entity.UpdateWithConfig(world.SimConfig)
	energyChangeDouble := initialEnergy - entity.Energy

	// Energy change should be approximately doubled (within small tolerance due to regeneration)
	expectedChange := energyChangeNormal * 2.0
	tolerance := 0.1

	if energyChangeDouble < expectedChange-tolerance || energyChangeDouble > expectedChange+tolerance {
		t.Errorf("Expected energy change at 2x speed to be ~%f, got %f (normal speed: %f)",
			expectedChange, energyChangeDouble, energyChangeNormal)
	}

	t.Logf("Energy change at 1x speed: %f", energyChangeNormal)
	t.Logf("Energy change at 2x speed: %f", energyChangeDouble)
	t.Logf("Speed multiplier effect verified within tolerance: %f", tolerance)
}

// TestSpeedMultiplierImpactOnMovement tests that speed multiplier affects movement energy costs
func TestSpeedMultiplierImpactOnMovement(t *testing.T) {
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
		PopulationSize: 1,
	}

	world := NewWorld(config)

	// Create an entity for testing
	entity := &Entity{
		ID:       1,
		Position: Position{X: 25.0, Y: 25.0},
		Energy:   100.0,
		Age:      1,
		IsAlive:  true,
		Species:  "test",
		Traits:   make(map[string]Trait),
	}

	// Test movement at normal speed
	world.SetSpeedMultiplier(1.0)
	initialEnergy := entity.Energy
	entity.MoveToWithConfig(30.0, 30.0, 5.0, world.SimConfig)
	energyCostNormal := initialEnergy - entity.Energy

	// Reset entity
	entity.Energy = 100.0
	entity.Position = Position{X: 25.0, Y: 25.0}

	// Test movement at double speed
	world.SetSpeedMultiplier(2.0)
	entity.MoveToWithConfig(30.0, 30.0, 5.0, world.SimConfig)
	energyCostDouble := initialEnergy - entity.Energy

	// Energy cost should be approximately doubled
	expectedCost := energyCostNormal * 2.0
	tolerance := 0.01

	if energyCostDouble < expectedCost-tolerance || energyCostDouble > expectedCost+tolerance {
		t.Errorf("Expected movement energy cost at 2x speed to be ~%f, got %f (normal speed: %f)",
			expectedCost, energyCostDouble, energyCostNormal)
	}

	t.Logf("Movement energy cost at 1x speed: %f", energyCostNormal)
	t.Logf("Movement energy cost at 2x speed: %f", energyCostDouble)
	t.Logf("Speed multiplier effect on movement verified")
}

// TestSpeedMultiplierConsistencyAcrossModules tests that speed multiplier is consistently applied
func TestSpeedMultiplierConsistencyAcrossModules(t *testing.T) {
	config := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
		PopulationSize: 3,
	}

	world := NewWorld(config)

	// Test different speed multipliers
	speedMultipliers := []float64{0.5, 1.0, 2.0, 4.0, 8.0}

	for _, speed := range speedMultipliers {
		world.SetSpeedMultiplier(speed)

		// Verify world's simulation config is updated
		if world.SimConfig == nil {
			t.Errorf("SimConfig should not be nil at speed %f", speed)
			continue
		}

		// Check that key values scale appropriately
		baseConfig := DefaultSimulationConfig()

		expectedEnergyDrain := baseConfig.Energy.BaseEnergyDrain * speed
		expectedMovementCost := baseConfig.Energy.MovementEnergyCost * speed
		expectedEnergyRegen := baseConfig.Energy.EnergyRegenerationRate * speed
		expectedDailyBase := baseConfig.Time.DailyEnergyBase * speed

		actualEnergyDrain := world.SimConfig.Energy.BaseEnergyDrain
		actualMovementCost := world.SimConfig.Energy.MovementEnergyCost
		actualEnergyRegen := world.SimConfig.Energy.EnergyRegenerationRate
		actualDailyBase := world.SimConfig.Time.DailyEnergyBase

		tolerance := 0.001

		if math.Abs(actualEnergyDrain-expectedEnergyDrain) > tolerance {
			t.Errorf("Speed %f: Expected energy drain %f, got %f", speed, expectedEnergyDrain, actualEnergyDrain)
		}

		if math.Abs(actualMovementCost-expectedMovementCost) > tolerance {
			t.Errorf("Speed %f: Expected movement cost %f, got %f", speed, expectedMovementCost, actualMovementCost)
		}

		if math.Abs(actualEnergyRegen-expectedEnergyRegen) > tolerance {
			t.Errorf("Speed %f: Expected energy regen %f, got %f", speed, expectedEnergyRegen, actualEnergyRegen)
		}

		if math.Abs(actualDailyBase-expectedDailyBase) > tolerance {
			t.Errorf("Speed %f: Expected daily base %f, got %f", speed, expectedDailyBase, actualDailyBase)
		}
	}
}

// TestSpeedMultiplierSimulationStability tests that simulation remains stable at different speeds
func TestSpeedMultiplierSimulationStability(t *testing.T) {
	speedTests := []struct {
		speed    float64
		name     string
		maxTicks int
	}{
		{0.25, "quarter speed", 40},
		{0.5, "half speed", 30},
		{1.0, "normal speed", 20},
		{2.0, "double speed", 15},
		{4.0, "quad speed", 10},
	}

	for _, test := range speedTests {
		t.Run(test.name, func(t *testing.T) {
			config := WorldConfig{
				Width:          50.0,
				Height:         50.0,
				GridWidth:      10,
				GridHeight:     10,
				NumPopulations: 1,
				PopulationSize: 5,
			}

			world := NewWorld(config)
			world.SetSpeedMultiplier(test.speed)

			// Add entities to the world
			for i := 0; i < 5; i++ {
				entity := &Entity{
					ID:             i + 1,
					Position:       Position{X: 25.0 + float64(i), Y: 25.0 + float64(i)},
					Energy:         100.0,
					Age:            1,
					IsAlive:        true,
					Species:        "test",
					Traits:         make(map[string]Trait),
					Classification: ClassificationEukaryotic,
					MaxLifespan:    210, // 30 days * 7 scale factor
				}

				// Initialize basic traits
				entity.Traits["size"] = Trait{Name: "size", Value: 0.0}
				entity.Traits["endurance"] = Trait{Name: "endurance", Value: 0.0}

				world.AllEntities = append(world.AllEntities, entity)
			}

			// Run simulation for specified ticks
			for tick := 0; tick < test.maxTicks; tick++ {
				world.Update()
			}

			// Check that at least some entities survived
			aliveCount := 0
			for _, entity := range world.AllEntities {
				if entity.IsAlive {
					aliveCount++
				}
			}

			survivalRate := float64(aliveCount) / float64(len(world.AllEntities))

			// Expect at least 20% survival regardless of speed
			if survivalRate < 0.2 {
				t.Errorf("Speed %f: Poor survival rate %f%% after %d ticks",
					test.speed, survivalRate*100, test.maxTicks)
			}

			t.Logf("Speed %f: %d/%d entities survived (%f%%) after %d ticks",
				test.speed, aliveCount, len(world.AllEntities), survivalRate*100, test.maxTicks)
		})
	}
}
