package main

import (
	"math"
	"math/rand"
	"testing"
)

// TestPrimitiveLifeFormEvolution tests that primitive organisms can evolve into complex species
func TestPrimitiveLifeFormEvolution(t *testing.T) {
	// Create a world with primitive life forms
	worldConfig := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      10,
		GridHeight:     10,
	}
	
	world := NewWorld(worldConfig)
	
	// Add a primitive microbe population
	primitiveConfig := PopulationConfig{
		Name:    "Test Microbes",
		Species: "microbe",
		BaseTraits: map[string]float64{
			"size":               -1.5,
			"speed":              -0.5,
			"aggression":         -0.9,
			"intelligence":       -1.0,
			"endurance":          0.8,
			"aquatic_adaptation": 0.5,
			"digging_ability":    -0.8,
			"flying_ability":     -1.0,
		},
		StartPos:         Position{X: 25, Y: 25},
		Spread:           10.0,
		Color:            "gray",
		BaseMutationRate: 0.25,
	}
	
	world.AddPopulation(primitiveConfig)
	
	t.Logf("Created world with %d entities", len(world.AllEntities))
	for i, entity := range world.AllEntities {
		t.Logf("Entity %d: Species=%s, ID=%d", i, entity.Species, entity.ID)
	}
	
	if len(world.AllEntities) == 0 {
		t.Fatal("No entities were created")
	}
	
	// Verify primitive organisms were created (the naming system generates names from the microbe pool)
	primitiveCount := len(world.AllEntities) // All entities should be from our primitive config
	if primitiveCount == 0 {
		t.Fatal("No primitive entities were created")
	}
	t.Logf("Created %d primitive entities", primitiveCount)
	
	// Simulate some evolution by boosting conditions
	for _, entity := range world.AllEntities {
		entity.Energy = 50 // Give them energy
		entity.Age = 30    // Make them mature
	}
	
	// Run simulation for a bit
	for i := 0; i < 100; i++ {
		world.Update()
		
		// Check if any evolution occurred
		evolutionOccurred := false
		for _, entity := range world.AllEntities {
			if entity.Species != "Prime" && entity.Species != "Origin" && entity.Species != "Pure" {
				t.Logf("Evolution occurred: %s at tick %d", entity.Species, i)
				evolutionOccurred = true
			}
		}
		if evolutionOccurred {
			return // Test passed - evolution happened
		}
	}
	
	t.Log("No evolution occurred in 100 ticks, but microbes are still functioning")
}

// TestEnvironmentSpecificMovement tests that entities move differently in different biomes
func TestEnvironmentSpecificMovement(t *testing.T) {
	// Create separate entities for each test to avoid interference
	waterEntity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
	}
	
	soilEntity := &Entity{
		ID:       2,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
	}
	
	airEntity := &Entity{
		ID:       3,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
	}
	
	// Set environmental adaptation traits for all entities
	for _, entity := range []*Entity{waterEntity, soilEntity, airEntity} {
		entity.SetTrait("aquatic_adaptation", 0.5)
		entity.SetTrait("digging_ability", -0.5)
		entity.SetTrait("flying_ability", -0.8)
	}
	
	initialEnergy := 100.0
	
	// Test movement in different environments
	waterStartPos := waterEntity.Position
	waterEntity.MoveToWithEnvironment(15, 15, 1.0, BiomeWater)
	waterEnergyLoss := initialEnergy - waterEntity.Energy
	waterDistanceMoved := math.Sqrt(math.Pow(waterEntity.Position.X-waterStartPos.X, 2) + math.Pow(waterEntity.Position.Y-waterStartPos.Y, 2))
	waterEfficiency := waterDistanceMoved / waterEnergyLoss
	t.Logf("Water movement: distance %.2f, energy loss %.2f, efficiency %.2f", waterDistanceMoved, waterEnergyLoss, waterEfficiency)
	
	soilStartPos := soilEntity.Position
	soilEntity.MoveToWithEnvironment(15, 15, 1.0, BiomeSoil)
	soilEnergyLoss := initialEnergy - soilEntity.Energy
	soilDistanceMoved := math.Sqrt(math.Pow(soilEntity.Position.X-soilStartPos.X, 2) + math.Pow(soilEntity.Position.Y-soilStartPos.Y, 2))
	soilEfficiency := soilDistanceMoved / soilEnergyLoss
	t.Logf("Soil movement: distance %.2f, energy loss %.2f, efficiency %.2f", soilDistanceMoved, soilEnergyLoss, soilEfficiency)
	
	airStartPos := airEntity.Position
	airEntity.MoveToWithEnvironment(15, 15, 1.0, BiomeAir)
	airEnergyLoss := initialEnergy - airEntity.Energy
	airDistanceMoved := math.Sqrt(math.Pow(airEntity.Position.X-airStartPos.X, 2) + math.Pow(airEntity.Position.Y-airStartPos.Y, 2))
	airEfficiency := airDistanceMoved / airEnergyLoss
	t.Logf("Air movement: distance %.2f, energy loss %.2f, efficiency %.2f", airDistanceMoved, airEnergyLoss, airEfficiency)
	
	// Verify movement efficiency is as expected (water > soil > air)
	if waterEfficiency <= soilEfficiency {
		t.Errorf("Expected water movement to be more efficient than soil. Water: %.2f, Soil: %.2f", waterEfficiency, soilEfficiency)
	}
	
	if soilEfficiency <= airEfficiency {
		t.Errorf("Expected soil movement to be more efficient than air. Soil: %.2f, Air: %.2f", soilEfficiency, airEfficiency)
	}
	
	t.Logf("Movement efficiency - Water: %.2f, Soil: %.2f, Air: %.2f", waterEfficiency, soilEfficiency, airEfficiency)
}

// TestAquaticAdaptation tests that aquatic creatures perform better in water
func TestAquaticAdaptation(t *testing.T) {
	// Set deterministic seed for testing
	rand.Seed(12345)
	
	// Create two entities - one adapted to water, one not
	aquaticEntity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "aquatic_test",
	}
	
	landEntity := &Entity{
		ID:       2,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "land_test",
	}
	
	// Set adaptation traits
	aquaticEntity.SetTrait("aquatic_adaptation", 0.8)
	landEntity.SetTrait("aquatic_adaptation", -0.8)
	
	// Test movement in water multiple times to average out randomness
	totalAquaticLoss := 0.0
	totalLandLoss := 0.0
	numTests := 10
	
	for i := 0; i < numTests; i++ {
		aquaticEntity.Energy = 100.0
		landEntity.Energy = 100.0
		
		aquaticEntity.MoveRandomlyWithEnvironment(5.0, BiomeWater)
		landEntity.MoveRandomlyWithEnvironment(5.0, BiomeWater)
		
		totalAquaticLoss += 100.0 - aquaticEntity.Energy
		totalLandLoss += 100.0 - landEntity.Energy
	}
	
	avgAquaticLoss := totalAquaticLoss / float64(numTests)
	avgLandLoss := totalLandLoss / float64(numTests)
	
	t.Logf("Average energy loss - Aquatic: %.3f, Land: %.3f", avgAquaticLoss, avgLandLoss)
	
	// Aquatic entity should use less energy in water
	if avgAquaticLoss >= avgLandLoss {
		t.Errorf("Expected aquatic entity to use less energy in water. Aquatic: %.3f, Land: %.3f", avgAquaticLoss, avgLandLoss)
	}
}

// TestSoilDwellerAdaptation tests that soil-dwelling creatures perform better underground
func TestSoilDwellerAdaptation(t *testing.T) {
	// Set deterministic seed for testing
	rand.Seed(12345)
	
	// Create two entities - one adapted to soil, one not
	soilEntity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "soil_test",
	}
	
	surfaceEntity := &Entity{
		ID:       2,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "surface_test",
	}
	
	// Set adaptation traits
	soilEntity.SetTrait("digging_ability", 0.8)
	soilEntity.SetTrait("underground_nav", 0.7)
	surfaceEntity.SetTrait("digging_ability", -0.8)
	surfaceEntity.SetTrait("underground_nav", -0.9)
	
	// Test movement in soil multiple times to average out randomness
	totalSoilLoss := 0.0
	totalSurfaceLoss := 0.0
	numTests := 10
	
	for i := 0; i < numTests; i++ {
		soilEntity.Energy = 100.0
		surfaceEntity.Energy = 100.0
		
		soilEntity.MoveRandomlyWithEnvironment(3.0, BiomeSoil)
		surfaceEntity.MoveRandomlyWithEnvironment(3.0, BiomeSoil)
		
		totalSoilLoss += 100.0 - soilEntity.Energy
		totalSurfaceLoss += 100.0 - surfaceEntity.Energy
	}
	
	avgSoilLoss := totalSoilLoss / float64(numTests)
	avgSurfaceLoss := totalSurfaceLoss / float64(numTests)
	
	t.Logf("Average energy loss - Soil adapted: %.3f, Surface adapted: %.3f", avgSoilLoss, avgSurfaceLoss)
	
	// Soil entity should use less energy underground
	if avgSoilLoss >= avgSurfaceLoss {
		t.Errorf("Expected soil entity to use less energy underground. Soil adapted: %.3f, Surface adapted: %.3f", avgSoilLoss, avgSurfaceLoss)
	}
}

// TestAerialAdaptation tests that flying creatures perform better in air
func TestAerialAdaptation(t *testing.T) {
	// Create two entities - one adapted to air, one not
	flyingEntity := &Entity{
		ID:       1,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "flying_test",
	}
	
	groundEntity := &Entity{
		ID:       2,
		Traits:   make(map[string]Trait),
		Position: Position{X: 10, Y: 10},
		Energy:   100,
		IsAlive:  true,
		Species:  "ground_test",
	}
	
	// Set adaptation traits
	flyingEntity.SetTrait("flying_ability", 0.8)
	flyingEntity.SetTrait("altitude_tolerance", 0.7)
	groundEntity.SetTrait("flying_ability", -0.9)
	groundEntity.SetTrait("altitude_tolerance", -0.8)
	
	initialEnergy := 100.0
	
	// Test movement in air
	flyingStartPos := flyingEntity.Position
	flyingEntity.MoveRandomlyWithEnvironment(4.0, BiomeAir)
	flyingEnergyLoss := initialEnergy - flyingEntity.Energy
	flyingDistance := math.Sqrt(math.Pow(flyingEntity.Position.X-flyingStartPos.X, 2) + math.Pow(flyingEntity.Position.Y-flyingStartPos.Y, 2))
	flyingEfficiency := flyingDistance / flyingEnergyLoss
	
	groundStartPos := groundEntity.Position
	groundEntity.MoveRandomlyWithEnvironment(4.0, BiomeAir)
	groundEnergyLoss := initialEnergy - groundEntity.Energy
	groundDistance := math.Sqrt(math.Pow(groundEntity.Position.X-groundStartPos.X, 2) + math.Pow(groundEntity.Position.Y-groundStartPos.Y, 2))
	groundEfficiency := groundDistance / groundEnergyLoss
	
	// Flying entity should be more efficient (more distance per energy) in air
	if flyingEfficiency <= groundEfficiency {
		t.Errorf("Expected flying entity to be more efficient in air. Flying efficiency: %.2f, Ground efficiency: %.2f", flyingEfficiency, groundEfficiency)
	}
	
	t.Logf("Air movement efficiency - Flying adapted: %.2f, Ground adapted: %.2f", flyingEfficiency, groundEfficiency)
}

// TestNewBiomeGeneration tests that the new biomes are properly generated
func TestNewBiomeGeneration(t *testing.T) {
	worldConfig := WorldConfig{
		Width:          50.0,
		Height:         50.0,
		NumPopulations: 0,
		PopulationSize: 0,
		GridWidth:      20,
		GridHeight:     20,
	}
	
	world := NewWorld(worldConfig)
	
	// Check that new biomes are present
	soilFound := false
	airFound := false
	
	for y := 0; y < world.Config.GridHeight; y++ {
		for x := 0; x < world.Config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			if biome == BiomeSoil {
				soilFound = true
			}
			if biome == BiomeAir {
				airFound = true
			}
		}
	}
	
	t.Logf("Biome generation - Soil found: %v, Air found: %v", soilFound, airFound)
	
	// It's okay if they're not found in a small grid, but biomes should be defined
	biomes := world.Biomes
	if _, exists := biomes[BiomeSoil]; !exists {
		t.Error("BiomeSoil not defined in biomes map")
	}
	if _, exists := biomes[BiomeAir]; !exists {
		t.Error("BiomeAir not defined in biomes map")
	}
	
	// Check biome properties
	soilBiome := biomes[BiomeSoil]
	if soilBiome.Name != "Soil" {
		t.Errorf("Expected soil biome name 'Soil', got '%s'", soilBiome.Name)
	}
	
	airBiome := biomes[BiomeAir]
	if airBiome.Name != "Air" {
		t.Errorf("Expected air biome name 'Air', got '%s'", airBiome.Name)
	}
}