package main

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaultConfigurationCreation(t *testing.T) {
	config := DefaultSimulationConfig()

	if config == nil {
		t.Fatal("DefaultSimulationConfig returned nil")
	}

	// Test time configuration
	if config.Time.TicksPerDay <= 0 {
		t.Errorf("TicksPerDay should be positive, got %d", config.Time.TicksPerDay)
	}

	if config.Time.DaysPerSeason <= 0 {
		t.Errorf("DaysPerSeason should be positive, got %d", config.Time.DaysPerSeason)
	}

	// Test energy configuration
	if config.Energy.SurvivalThreshold >= config.Energy.MaxEnergyLevel {
		t.Errorf("SurvivalThreshold (%f) should be less than MaxEnergyLevel (%f)",
			config.Energy.SurvivalThreshold, config.Energy.MaxEnergyLevel)
	}

	// Test population configuration
	if config.Population.DefaultPopSize <= 0 {
		t.Errorf("DefaultPopSize should be positive, got %d", config.Population.DefaultPopSize)
	}

	// Test world configuration
	if config.World.Width <= 0 || config.World.Height <= 0 {
		t.Errorf("World dimensions should be positive, got %fx%f", config.World.Width, config.World.Height)
	}
}

func TestConfigurationValidation(t *testing.T) {
	// Test valid configuration
	config := DefaultSimulationConfig()
	err := config.Validate()
	if err != nil {
		t.Errorf("Valid configuration failed validation: %v", err)
	}

	// Test invalid configurations
	testCases := []struct {
		name        string
		modify      func(*SimulationConfig)
		expectError bool
	}{
		{
			name:        "negative ticks per day",
			modify:      func(c *SimulationConfig) { c.Time.TicksPerDay = -1 },
			expectError: true,
		},
		{
			name:        "zero days per season",
			modify:      func(c *SimulationConfig) { c.Time.DaysPerSeason = 0 },
			expectError: true,
		},
		{
			name:        "survival threshold too high",
			modify:      func(c *SimulationConfig) { c.Energy.SurvivalThreshold = 200.0 },
			expectError: true,
		},
		{
			name:        "negative population size",
			modify:      func(c *SimulationConfig) { c.Population.DefaultPopSize = -5 },
			expectError: true,
		},
		{
			name:        "zero world width",
			modify:      func(c *SimulationConfig) { c.World.Width = 0 },
			expectError: true,
		},
		{
			name:        "negative grid height",
			modify:      func(c *SimulationConfig) { c.World.GridHeight = -1 },
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultSimulationConfig()
			tc.modify(config)
			err := config.Validate()

			if tc.expectError && err == nil {
				t.Errorf("Expected validation error for %s, but got none", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
			}
		})
	}
}

func TestSpeedMultiplierApplication(t *testing.T) {
	baseConfig := DefaultSimulationConfig()

	testCases := []struct {
		multiplier float64
		name       string
	}{
		{0.25, "quarter speed"},
		{0.5, "half speed"},
		{1.0, "normal speed"},
		{2.0, "double speed"},
		{4.0, "quadruple speed"},
		{8.0, "octuple speed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			speedConfig := baseConfig.ApplySpeedMultiplier(tc.multiplier)

			// Test that energy values scale with multiplier
			expectedBaseDrain := baseConfig.Energy.BaseEnergyDrain * tc.multiplier
			if speedConfig.Energy.BaseEnergyDrain != expectedBaseDrain {
				t.Errorf("BaseEnergyDrain: expected %f, got %f", expectedBaseDrain, speedConfig.Energy.BaseEnergyDrain)
			}

			expectedMovementCost := baseConfig.Energy.MovementEnergyCost * tc.multiplier
			if speedConfig.Energy.MovementEnergyCost != expectedMovementCost {
				t.Errorf("MovementEnergyCost: expected %f, got %f", expectedMovementCost, speedConfig.Energy.MovementEnergyCost)
			}

			expectedRegenRate := baseConfig.Energy.EnergyRegenerationRate * tc.multiplier
			if speedConfig.Energy.EnergyRegenerationRate != expectedRegenRate {
				t.Errorf("EnergyRegenerationRate: expected %f, got %f", expectedRegenRate, speedConfig.Energy.EnergyRegenerationRate)
			}

			// Test that time values scale with multiplier
			expectedDailyBase := baseConfig.Time.DailyEnergyBase * tc.multiplier
			if speedConfig.Time.DailyEnergyBase != expectedDailyBase {
				t.Errorf("DailyEnergyBase: expected %f, got %f", expectedDailyBase, speedConfig.Time.DailyEnergyBase)
			}

			// Test that plant values scale with multiplier
			expectedGrowthRate := baseConfig.Plants.GrowthRate * tc.multiplier
			if speedConfig.Plants.GrowthRate != expectedGrowthRate {
				t.Errorf("Plant GrowthRate: expected %f, got %f", expectedGrowthRate, speedConfig.Plants.GrowthRate)
			}

			// Test web update interval for speeds > 1.0
			if tc.multiplier > 1.0 {
				expectedInterval := time.Duration(float64(baseConfig.Web.UpdateInterval) / tc.multiplier)
				if speedConfig.Web.UpdateInterval != expectedInterval {
					t.Errorf("Web UpdateInterval: expected %v, got %v", expectedInterval, speedConfig.Web.UpdateInterval)
				}
			} else {
				// For speeds <= 1.0, interval should remain unchanged
				if speedConfig.Web.UpdateInterval != baseConfig.Web.UpdateInterval {
					t.Errorf("Web UpdateInterval should not change for multiplier <= 1.0, expected %v, got %v",
						baseConfig.Web.UpdateInterval, speedConfig.Web.UpdateInterval)
				}
			}

			// Test that original config is not modified
			if baseConfig.Energy.BaseEnergyDrain == speedConfig.Energy.BaseEnergyDrain && tc.multiplier != 1.0 {
				t.Errorf("Original config was modified when applying speed multiplier")
			}
		})
	}
}

func TestBiomeEnergyDrainRetrieval(t *testing.T) {
	config := DefaultSimulationConfig()

	testCases := []struct {
		biome    BiomeType
		expected float64
		name     string
	}{
		{BiomePlains, config.Energy.BaseEnergyDrain * 0.5, "plains"},
		{BiomeForest, config.Energy.BaseEnergyDrain * 0.8, "forest"},
		{BiomeDesert, config.Energy.BaseEnergyDrain * 1.5, "desert"},
		{BiomeRadiation, config.Energy.BaseEnergyDrain * 2.0, "radiation"},
		{BiomeIce, config.Energy.BaseEnergyDrain * 2.5, "ice"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := config.GetBiomeEnergyDrain(tc.biome)
			if actual != tc.expected {
				t.Errorf("Expected energy drain %f for %s, got %f", tc.expected, tc.name, actual)
			}
		})
	}

	// Test unknown biome type falls back to base energy drain
	unknownBiome := BiomeType(999)
	actual := config.GetBiomeEnergyDrain(unknownBiome)
	if actual != config.Energy.BaseEnergyDrain {
		t.Errorf("Unknown biome should return base energy drain %f, got %f", config.Energy.BaseEnergyDrain, actual)
	}
}

func TestBiomeMutationModifierRetrieval(t *testing.T) {
	config := DefaultSimulationConfig()

	testCases := []struct {
		biome    BiomeType
		expected float64
		name     string
	}{
		{BiomeRadiation, 2.0, "radiation"},
		{BiomeHotSpring, 1.5, "hot spring"},
		{BiomeDesert, 1.2, "desert"},
		{BiomeHighAltitude, 1.3, "high altitude"},
		{BiomePlains, 1.0, "plains (no modifier)"},
		{BiomeForest, 1.0, "forest (no modifier)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := config.GetBiomeMutationModifier(tc.biome)
			if actual != tc.expected {
				t.Errorf("Expected mutation modifier %f for %s, got %f", tc.expected, tc.name, actual)
			}
		})
	}

	// Test unknown biome type falls back to 1.0
	unknownBiome := BiomeType(999)
	actual := config.GetBiomeMutationModifier(unknownBiome)
	if actual != 1.0 {
		t.Errorf("Unknown biome should return default modifier 1.0, got %f", actual)
	}
}

func TestConfigurationConsistency(t *testing.T) {
	config := DefaultSimulationConfig()

	// Test that all biome types have energy modifiers
	allBiomes := []BiomeType{
		BiomePlains, BiomeForest, BiomeDesert, BiomeMountain, BiomeWater,
		BiomeRadiation, BiomeSoil, BiomeAir, BiomeIce, BiomeRainforest,
		BiomeDeepWater, BiomeHighAltitude, BiomeHotSpring, BiomeTundra,
		BiomeSwamp, BiomeCanyon,
	}

	for _, biome := range allBiomes {
		energyDrain := config.GetBiomeEnergyDrain(biome)
		if energyDrain <= 0 {
			t.Errorf("Biome %v should have positive energy drain, got %f", biome, energyDrain)
		}
	}

	// Test that trait bounds are symmetric where appropriate
	for traitName, bounds := range config.Evolution.TraitBounds {
		if bounds[0] > bounds[1] {
			t.Errorf("Trait %s has invalid bounds: min (%f) > max (%f)", traitName, bounds[0], bounds[1])
		}
	}

	// Test that fitness weights are positive
	for factor, weight := range config.Evolution.FitnessWeights {
		if weight < 0 {
			t.Errorf("Fitness weight for %s should be non-negative, got %f", factor, weight)
		}
	}
}

func TestSpeedMultiplierImpactOnSimulation(t *testing.T) {
	baseConfig := DefaultSimulationConfig()

	// Test different speed multipliers and ensure they create proportional impacts
	multipliers := []float64{0.5, 1.0, 2.0, 4.0}

	for _, multiplier := range multipliers {
		t.Run(fmt.Sprintf("speed_%0.1fx", multiplier), func(t *testing.T) {
			speedConfig := baseConfig.ApplySpeedMultiplier(multiplier)

			// Energy consumption should scale linearly with speed
			energyRatio := speedConfig.Energy.BaseEnergyDrain / baseConfig.Energy.BaseEnergyDrain
			if abs(energyRatio-multiplier) > 0.001 {
				t.Errorf("Energy ratio should equal multiplier: expected %f, got %f", multiplier, energyRatio)
			}

			// Plant growth should scale linearly with speed
			growthRatio := speedConfig.Plants.GrowthRate / baseConfig.Plants.GrowthRate
			if abs(growthRatio-multiplier) > 0.001 {
				t.Errorf("Growth ratio should equal multiplier: expected %f, got %f", multiplier, growthRatio)
			}

			// Validate the modified configuration
			err := speedConfig.Validate()
			if err != nil {
				t.Errorf("Speed config for %fx failed validation: %v", multiplier, err)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
