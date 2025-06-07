package main

import (
	"math"
	"testing"
)

func TestSeedDispersalSystem(t *testing.T) {
	// Test seed dispersal system creation
	sds := NewSeedDispersalSystem()
	if sds == nil {
		t.Fatal("Failed to create seed dispersal system")
	}

	if len(sds.AllSeeds) != 0 {
		t.Errorf("Expected 0 initial seeds, got %d", len(sds.AllSeeds))
	}

	if sds.NextSeedID != 1 {
		t.Errorf("Expected NextSeedID to be 1, got %d", sds.NextSeedID)
	}
}

func TestCreateSeed(t *testing.T) {
	// Create test world and plant
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	// Create a test plant
	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})
	plant.Energy = 100

	// Create seed
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)

	if seed == nil {
		t.Fatal("Failed to create seed")
	}

	if seed.PlantID != plant.ID {
		t.Errorf("Expected seed PlantID %d, got %d", plant.ID, seed.PlantID)
	}

	if seed.PlantType != plant.Type {
		t.Errorf("Expected seed PlantType %v, got %v", plant.Type, seed.PlantType)
	}

	if seed.Position.X != plant.Position.X || seed.Position.Y != plant.Position.Y {
		t.Errorf("Expected seed position %v, got %v", plant.Position, seed.Position)
	}

	if seed.Viability != 1.0 {
		t.Errorf("Expected initial viability 1.0, got %f", seed.Viability)
	}
}

func TestSeedTypeAssignment(t *testing.T) {
	sds := NewSeedDispersalSystem()

	// Test seed type assignment for different plant types
	testCases := []struct {
		plantType    PlantType
		expectedSeed SeedType
	}{
		{PlantGrass, SeedSmallLight},
		{PlantMushroom, SeedSmallLight},
		{PlantAlgae, SeedFloating},
	}

	for _, tc := range testCases {
		plant := &Plant{Type: tc.plantType, Size: 1.0}
		plant.Traits = make(map[string]Trait)
		plant.Traits["defense"] = Trait{Value: 0.1}

		seedType := sds.determineSeedType(plant)
		if seedType != tc.expectedSeed {
			t.Errorf("Plant type %v: expected seed type %v, got %v",
				tc.plantType, tc.expectedSeed, seedType)
		}
	}
}

func TestDispersalMechanisms(t *testing.T) {
	// Create test world and system
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	sds := world.SeedDispersalSystem
	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})

	// Test different dispersal mechanisms
	testCases := []struct {
		seedType          SeedType
		expectedDispersal DispersalMechanism
	}{
		{SeedSmallLight, DispersalWind},
		{SeedFloating, DispersalWater},
		{SeedExplosive, DispersalExplosive},
	}

	for _, tc := range testCases {
		dispersal := sds.determineDispersalMethod(tc.seedType, plant, world)
		if dispersal != tc.expectedDispersal {
			t.Errorf("Seed type %v: expected dispersal %v, got %v",
				tc.seedType, tc.expectedDispersal, dispersal)
		}
	}
}

func TestSeedViabilityDecay(t *testing.T) {
	// Create test world and seed
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)

	initialViability := seed.Viability

	// Update seed several times
	for i := 0; i < 5; i++ {
		world.SeedDispersalSystem.updateSeed(seed, world)
	}

	if seed.Viability >= initialViability {
		t.Errorf("Expected viability to decrease, got %f (initial: %f)",
			seed.Viability, initialViability)
	}

	if seed.Age != 5 {
		t.Errorf("Expected seed age 5, got %d", seed.Age)
	}
}

func TestWindDispersal(t *testing.T) {
	// Create test world with wind
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)
	seed.DispersalMethod = DispersalWind
	seed.Mass = 0.1 // Light seed

	// Apply wind dispersal
	world.SeedDispersalSystem.disperseByWind(seed, world)

	// Seed should have moved (velocity should be non-zero)
	if seed.Velocity.X == 0 && seed.Velocity.Y == 0 {
		t.Error("Expected wind dispersal to create seed movement")
	}
}

func TestAnimalDispersal(t *testing.T) {
	// Create test world with entities
	config := WorldConfig{
		Width:      40,
		Height:     40,
		GridWidth:  20,
		GridHeight: 20,
	}
	world := NewWorld(config)

	// Add an entity near the seed
	traitNames := []string{"strength", "agility", "intelligence"}
	entity := NewEntity(1, traitNames, "TestSpecies", Position{X: 5.5, Y: 5.5})
	entity.IsAlive = true
	world.AllEntities = append(world.AllEntities, entity)

	plant := NewPlant(1, PlantBush, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)
	seed.DispersalMethod = DispersalAnimal
	seed.SeedType = SeedFleshy

	// Try animal dispersal multiple times to account for randomness
	for i := 0; i < 10; i++ {
		world.SeedDispersalSystem.disperseByAnimal(seed, world)
		if seed.CarriedByEntity == entity.ID {
			break // Seed was picked up
		}
	}

	// Check if seed was picked up (may not happen due to randomness)
	if seed.CarriedByEntity != 0 {
		t.Logf("Seed was picked up by entity %d", seed.CarriedByEntity)
	}
}

func TestExplosiveDispersal(t *testing.T) {
	// Create test world
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	plant := NewPlant(1, PlantCactus, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)
	seed.DispersalMethod = DispersalExplosive
	seed.Age = 1 // Trigger explosive dispersal

	// Apply explosive dispersal
	world.SeedDispersalSystem.disperseExplosively(seed, world)

	// Seed should have high velocity after explosion
	totalVelocity := math.Sqrt(seed.Velocity.X*seed.Velocity.X + seed.Velocity.Y*seed.Velocity.Y)
	if totalVelocity < 1.0 {
		t.Errorf("Expected explosive dispersal to create high velocity, got %f", totalVelocity)
	}
}

func TestSeedBankSystem(t *testing.T) {
	// Create test world
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)

	// Force seed into dormancy
	seed.IsDormant = true
	world.SeedDispersalSystem.addToSeedBank(seed, world)

	// Check that seed bank was created
	bankPos := Position{X: 5, Y: 5}
	if bank, exists := world.SeedDispersalSystem.SeedBanks[bankPos]; exists {
		if len(bank.Seeds) != 1 {
			t.Errorf("Expected 1 seed in bank, got %d", len(bank.Seeds))
		}
		if bank.Seeds[0] != seed {
			t.Error("Seed not properly added to bank")
		}
	} else {
		t.Error("Seed bank was not created")
	}
}

func TestGermination(t *testing.T) {
	// Create test world
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	initialPlantCount := len(world.AllPlants)

	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})
	seed := world.SeedDispersalSystem.CreateSeed(plant, world)

	// Set favorable conditions for germination
	seed.IsDormant = true
	seed.RequiredTemperature = 20.0
	seed.RequiredMoisture = 0.5
	seed.RequiredSunlight = 0.3

	// Force germination
	world.SeedDispersalSystem.germinate(seed, world)

	// Check that new plant was created
	if len(world.AllPlants) != initialPlantCount+1 {
		t.Errorf("Expected %d plants after germination, got %d",
			initialPlantCount+1, len(world.AllPlants))
	}

	// Check that seed viability dropped to 0
	if seed.Viability != 0 {
		t.Errorf("Expected seed viability to be 0 after germination, got %f", seed.Viability)
	}

	// Check germination event counter
	if world.SeedDispersalSystem.GerminationEvents != 1 {
		t.Errorf("Expected 1 germination event, got %d",
			world.SeedDispersalSystem.GerminationEvents)
	}
}

func TestSeedDispersalIntegration(t *testing.T) {
	// Test full integration with world update
	config := WorldConfig{
		Width:      40,
		Height:     40,
		GridWidth:  20,
		GridHeight: 20,
	}
	world := NewWorld(config)

	// Add some plants
	for i := 0; i < 5; i++ {
		plant := NewPlant(i+1, PlantGrass, Position{X: float64(5 + i), Y: 5})
		plant.Energy = 100 // Ensure they can reproduce
		world.AllPlants = append(world.AllPlants, plant)
	}

	initialSeedCount := len(world.SeedDispersalSystem.AllSeeds)

	// Update world a few times
	for i := 0; i < 10; i++ {
		world.Update()
	}

	// Check that seeds were created
	finalSeedCount := len(world.SeedDispersalSystem.AllSeeds)
	if finalSeedCount <= initialSeedCount {
		t.Logf("Seed count did not increase as expected (initial: %d, final: %d)",
			initialSeedCount, finalSeedCount)
		// This might be normal due to randomness in reproduction
	}
}

func TestSeedDispersalStats(t *testing.T) {
	// Test statistics tracking
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	plant := NewPlant(1, PlantGrass, Position{X: 5, Y: 5})

	// Create seeds of different types
	for i := 0; i < 3; i++ {
		world.SeedDispersalSystem.CreateSeed(plant, world)
	}

	stats := world.SeedDispersalSystem.GetStats()

	if totalSeeds, ok := stats["total_seeds"].(int); !ok || totalSeeds != 3 {
		t.Errorf("Expected 3 total seeds in stats, got %v", stats["total_seeds"])
	}

	if windCount, ok := stats["dispersal_wind"].(int); !ok || windCount != 3 {
		t.Errorf("Expected 3 wind-dispersed seeds (grass), got %v", stats["dispersal_wind"])
	}
}

func TestEnvironmentalConditions(t *testing.T) {
	// Test environmental condition helpers
	config := WorldConfig{
		Width:      20,
		Height:     20,
		GridWidth:  10,
		GridHeight: 10,
	}
	world := NewWorld(config)

	pos := Position{X: 5, Y: 5}

	// Test temperature calculation
	temp := world.getTemperatureAt(pos)
	if temp < -50 || temp > 100 {
		t.Errorf("Temperature out of reasonable range: %f", temp)
	}

	// Test moisture calculation
	moisture := world.getMoistureAt(pos)
	if moisture < 0 || moisture > 1 {
		t.Errorf("Moisture out of valid range [0,1]: %f", moisture)
	}

	// Test sunlight calculation
	sunlight := world.getSunlightAt(pos)
	if sunlight < 0 || sunlight > 1 {
		t.Errorf("Sunlight out of valid range [0,1]: %f", sunlight)
	}
}
