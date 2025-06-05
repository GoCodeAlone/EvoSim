package main

import (
	"testing"
)

func TestTimeSystemConfigurationIntegration(t *testing.T) {
	// Create a custom configuration
	config := DefaultSimulationConfig()
	config.Time.TicksPerDay = 10
	config.Time.DaysPerSeason = 50
	config.Time.SeasonalVariation = 0.5

	// Create time system with configuration
	timeSystem := NewAdvancedTimeSystem(&config.Time)

	// Verify configuration was applied
	if timeSystem.DayLength != 10 {
		t.Errorf("Expected DayLength 10, got %d", timeSystem.DayLength)
	}

	if timeSystem.SeasonLength != 50 {
		t.Errorf("Expected SeasonLength 50, got %d", timeSystem.SeasonLength)
	}

	if timeSystem.Config.SeasonalVariation != 0.5 {
		t.Errorf("Expected SeasonalVariation 0.5, got %f", timeSystem.Config.SeasonalVariation)
	}

	// Test seasonal modifier calculation with custom configuration
	timeSystem.Season = Winter
	winterMod := timeSystem.getSeasonalModifier()
	expectedWinterMod := 1.0 - (0.4 * 0.5) // base - (variation * config)
	if abs(winterMod-expectedWinterMod) > 0.001 {
		t.Errorf("Expected winter modifier %f, got %f", expectedWinterMod, winterMod)
	}

	timeSystem.Season = Spring
	springMod := timeSystem.getSeasonalModifier()
	expectedSpringMod := 1.0 + (0.2 * 0.5) // base + (variation * config)
	if abs(springMod-expectedSpringMod) > 0.001 {
		t.Errorf("Expected spring modifier %f, got %f", expectedSpringMod, springMod)
	}
}

func TestEnergySystemConfigurationIntegration(t *testing.T) {
	// Create a custom configuration
	config := DefaultSimulationConfig()
	config.Energy.BaseEnergyDrain = 0.05
	config.Energy.MovementEnergyCost = 0.02
	config.Energy.EnergyRegenerationRate = 0.15

	// Create world with custom configuration
	worldConfig := WorldConfig{
		Width: 100, Height: 100,
		GridWidth: 20, GridHeight: 20,
		NumPopulations: 1, PopulationSize: 5,
	}

	world := NewWorldWithConfig(worldConfig, config)

	// Verify configuration was applied to world
	if world.SimConfig.Energy.BaseEnergyDrain != 0.05 {
		t.Errorf("Expected BaseEnergyDrain 0.05, got %f", world.SimConfig.Energy.BaseEnergyDrain)
	}

	// Test biome energy drain calculation
	plainsEnergyDrain := world.SimConfig.GetBiomeEnergyDrain(BiomePlains)
	expectedPlainsEnergyDrain := 0.05 * 0.5 // base * plains multiplier
	if abs(plainsEnergyDrain-expectedPlainsEnergyDrain) > 0.001 {
		t.Errorf("Expected plains energy drain %f, got %f", expectedPlainsEnergyDrain, plainsEnergyDrain)
	}

	desertEnergyDrain := world.SimConfig.GetBiomeEnergyDrain(BiomeDesert)
	expectedDesertEnergyDrain := 0.05 * 1.5 // base * desert multiplier
	if abs(desertEnergyDrain-expectedDesertEnergyDrain) > 0.001 {
		t.Errorf("Expected desert energy drain %f, got %f", expectedDesertEnergyDrain, desertEnergyDrain)
	}
}

func TestSpeedMultiplierConfigurationIntegration(t *testing.T) {
	// Create a world with default configuration
	worldConfig := WorldConfig{
		Width: 100, Height: 100,
		GridWidth: 20, GridHeight: 20,
		NumPopulations: 1, PopulationSize: 5,
	}

	baseConfig := DefaultSimulationConfig()
	world := NewWorldWithConfig(worldConfig, baseConfig)

	// Capture original energy values
	originalBaseEnergyDrain := baseConfig.Energy.BaseEnergyDrain
	originalMovementCost := baseConfig.Energy.MovementEnergyCost
	originalRegenRate := baseConfig.Energy.EnergyRegenerationRate

	// Apply speed multiplier
	speedMultiplier := 2.0
	world.SetSpeedMultiplier(speedMultiplier)

	// Verify world speed multiplier was set
	if world.GetSpeedMultiplier() != speedMultiplier {
		t.Errorf("Expected speed multiplier %f, got %f", speedMultiplier, world.GetSpeedMultiplier())
	}

	// Verify configuration was updated
	if world.SimConfig.Energy.BaseEnergyDrain != originalBaseEnergyDrain*speedMultiplier {
		t.Errorf("Expected updated BaseEnergyDrain %f, got %f",
			originalBaseEnergyDrain*speedMultiplier, world.SimConfig.Energy.BaseEnergyDrain)
	}

	if world.SimConfig.Energy.MovementEnergyCost != originalMovementCost*speedMultiplier {
		t.Errorf("Expected updated MovementEnergyCost %f, got %f",
			originalMovementCost*speedMultiplier, world.SimConfig.Energy.MovementEnergyCost)
	}

	if world.SimConfig.Energy.EnergyRegenerationRate != originalRegenRate*speedMultiplier {
		t.Errorf("Expected updated EnergyRegenerationRate %f, got %f",
			originalRegenRate*speedMultiplier, world.SimConfig.Energy.EnergyRegenerationRate)
	}
}

func TestConfigurationBiomeEnergyConsistency(t *testing.T) {
	// Create world with configuration
	worldConfig := WorldConfig{
		Width: 100, Height: 100,
		GridWidth: 20, GridHeight: 20,
		NumPopulations: 1, PopulationSize: 5,
	}

	config := DefaultSimulationConfig()
	world := NewWorldWithConfig(worldConfig, config)

	// Test that biome energy drain matches configuration values
	allBiomes := []BiomeType{
		BiomePlains, BiomeForest, BiomeDesert, BiomeMountain, BiomeWater,
		BiomeRadiation, BiomeSoil, BiomeAir, BiomeIce, BiomeRainforest,
		BiomeDeepWater, BiomeHighAltitude, BiomeHotSpring, BiomeTundra,
		BiomeSwamp, BiomeCanyon,
	}

	for _, biomeType := range allBiomes {
		biome := world.Biomes[biomeType]
		configEnergyDrain := config.GetBiomeEnergyDrain(biomeType)

		if abs(biome.EnergyDrain-configEnergyDrain) > 0.001 {
			t.Errorf("Biome %v energy drain mismatch: biome has %f, config has %f",
				biomeType, biome.EnergyDrain, configEnergyDrain)
		}
	}
}

func TestWorldConfigurationSpeedMultiplierEffect(t *testing.T) {
	// Test that speed multiplier changes affect world behavior consistently

	// Create two worlds with different speed multipliers
	worldConfig := WorldConfig{
		Width: 100, Height: 100,
		GridWidth: 20, GridHeight: 20,
		NumPopulations: 1, PopulationSize: 5,
	}

	baseConfig := DefaultSimulationConfig()

	world1x := NewWorldWithConfig(worldConfig, baseConfig)
	world1x.SetSpeedMultiplier(1.0)

	world2x := NewWorldWithConfig(worldConfig, baseConfig)
	world2x.SetSpeedMultiplier(2.0)

	// Compare energy configuration values
	energy1x := world1x.SimConfig.Energy.BaseEnergyDrain
	energy2x := world2x.SimConfig.Energy.BaseEnergyDrain

	expectedRatio := 2.0
	actualRatio := energy2x / energy1x

	if abs(actualRatio-expectedRatio) > 0.001 {
		t.Errorf("Expected energy ratio %f between 2x and 1x speed, got %f", expectedRatio, actualRatio)
	}

	// Test biome energy drains scale properly
	plains1x := world1x.SimConfig.GetBiomeEnergyDrain(BiomePlains)
	plains2x := world2x.SimConfig.GetBiomeEnergyDrain(BiomePlains)

	plainsRatio := plains2x / plains1x
	if abs(plainsRatio-expectedRatio) > 0.001 {
		t.Errorf("Expected plains energy ratio %f between 2x and 1x speed, got %f", expectedRatio, plainsRatio)
	}
}
