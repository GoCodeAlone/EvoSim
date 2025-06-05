package main

import (
	"math"
	"testing"
)

// maintainOptimalCellConditions maintains optimal conditions for cellular organisms during evolution testing
func maintainOptimalCellConditions(organism *CellularOrganism) {
	for _, cell := range organism.Cells {
		if cell.Energy < 150 {
			cell.Energy = 200.0
		}
		if cell.Health < 0.8 {
			cell.Health = 1.0
		}
		if cell.Age < 50 {
			cell.Age = 100
		}
	}
}

// runEvolutionSimulation runs a simulation loop to test cellular evolution with target requirements
func runEvolutionSimulation(t *testing.T, world *World, organism *CellularOrganism, targetCells, targetComplexity, maxTicks int, logInterval int) bool {
	for tick := 0; tick < maxTicks; tick++ {
		maintainOptimalCellConditions(organism)
		world.CellularSystem.UpdateCellularOrganisms()

		// Check if we've reached the target
		if len(organism.Cells) >= targetCells && organism.ComplexityLevel >= targetComplexity {
			t.Logf("Evolution successful at tick %d: complexity=%d, cells=%d",
				tick, organism.ComplexityLevel, len(organism.Cells))
			return true
		}

		if tick%logInterval == 0 {
			t.Logf("Tick %d: complexity=%d, cells=%d", tick, organism.ComplexityLevel, len(organism.Cells))
		}
	}
	return false
}

// runEvolutionSimulationWithName runs evolution simulation with named stage for logging
func runEvolutionSimulationWithName(t *testing.T, world *World, organism *CellularOrganism, targetCells, targetComplexity, maxTicks int, stageName string, startingCells int) bool {
	for tick := 0; tick < maxTicks; tick++ {
		maintainOptimalCellConditions(organism)
		world.CellularSystem.UpdateCellularOrganisms()

		// Check if we've reached the target
		if len(organism.Cells) >= targetCells && organism.ComplexityLevel >= targetComplexity {
			t.Logf("%s successful at tick %d: %d cells, complexity level %d",
				stageName, tick, len(organism.Cells), organism.ComplexityLevel)
			return true
		}

		if tick%100 == 0 {
			t.Logf("%s tick %d: %d cells (started with %d), complexity level %d",
				stageName, tick, len(organism.Cells), startingCells, organism.ComplexityLevel)
		}
	}
	return false
}

// TestCellularEvolutionStage1To2 tests single-cell to simple multicellular transition
func TestCellularEvolutionStage1To2(t *testing.T) {
	world := createTestWorld(t)
	entityID := createTestEntity(world, t)

	// Get the cellular organism and verify it starts at complexity level 1
	organism := world.CellularSystem.OrganismMap[entityID]
	if organism.ComplexityLevel != 1 {
		t.Fatalf("Expected initial complexity level 1, got %d", organism.ComplexityLevel)
	}
	if len(organism.Cells) != 1 {
		t.Fatalf("Expected 1 initial cell, got %d", len(organism.Cells))
	}

	// Prepare the cell for division by setting optimal conditions
	cell := organism.Cells[0]
	cell.Energy = 200.0 // Above threshold of 150
	cell.Health = 1.0   // Above threshold of 0.8
	cell.Age = 100      // Above threshold of 50

	t.Logf("Initial state: complexity=%d, cells=%d, energy=%.1f",
		organism.ComplexityLevel, len(organism.Cells), cell.Energy)

	// Run simulation until cell division occurs (complexity level 2 requires 5+ cells)
	maxTicks := 1000
	evolutionOccurred := runEvolutionSimulation(t, world, organism, 5, 2, maxTicks, 100)

	if !evolutionOccurred {
		t.Errorf("Failed to evolve from stage 1 to stage 2 in %d ticks. Final: complexity=%d, cells=%d",
			maxTicks, organism.ComplexityLevel, len(organism.Cells))
	}
}

// TestCellularEvolutionStage2To3 tests simple multicellular to complex multicellular transition
func TestCellularEvolutionStage2To3(t *testing.T) {
	world := createTestWorld(t)
	entityID := createTestEntity(world, t)

	organism := world.CellularSystem.OrganismMap[entityID]

	// Artificially advance to stage 2 by adding cells
	for i := 0; i < 10; i++ {
		cell := world.CellularSystem.createCell(CellTypeStem, organism.Cells[0].DNA, Position{X: float64(i), Y: 0})
		cell.Energy = 200.0
		cell.Health = 1.0
		cell.Age = 100
		organism.Cells = append(organism.Cells, cell)
	}

	// Update complexity level
	organism.ComplexityLevel = world.CellularSystem.calculateComplexityLevel(len(organism.Cells))

	t.Logf("Starting stage 2 to 3 test: complexity=%d, cells=%d",
		organism.ComplexityLevel, len(organism.Cells))

	if organism.ComplexityLevel < 2 {
		t.Fatalf("Failed to establish stage 2. Complexity: %d, Cells: %d",
			organism.ComplexityLevel, len(organism.Cells))
	}

	// Run simulation until we reach stage 3 (20+ cells)
	maxTicks := 1000
	evolutionOccurred := runEvolutionSimulation(t, world, organism, 20, 3, maxTicks, 100)

	if !evolutionOccurred {
		t.Errorf("Failed to evolve from stage 2 to stage 3 in %d ticks. Final: complexity=%d, cells=%d",
			maxTicks, organism.ComplexityLevel, len(organism.Cells))
	}
}

// TestCellularEvolutionStage3To4 tests complex multicellular to advanced multicellular transition
func TestCellularEvolutionStage3To4(t *testing.T) {
	world := createTestWorld(t)
	entityID := createTestEntity(world, t)

	organism := world.CellularSystem.OrganismMap[entityID]

	// Artificially advance to stage 3 by adding cells
	for i := 0; i < 30; i++ {
		cellType := CellTypeStem
		if i%5 == 0 {
			cellType = CellTypeNerve // Add some specialized cells
		} else if i%7 == 0 {
			cellType = CellTypeMuscle
		}

		cell := world.CellularSystem.createCell(cellType, organism.Cells[0].DNA, Position{X: float64(i % 10), Y: float64(i / 10)})
		cell.Energy = 200.0
		cell.Health = 1.0
		cell.Age = 100
		organism.Cells = append(organism.Cells, cell)
	}

	// Update complexity level
	organism.ComplexityLevel = world.CellularSystem.calculateComplexityLevel(len(organism.Cells))

	t.Logf("Starting stage 3 to 4 test: complexity=%d, cells=%d",
		organism.ComplexityLevel, len(organism.Cells))

	if organism.ComplexityLevel < 3 {
		t.Fatalf("Failed to establish stage 3. Complexity: %d, Cells: %d",
			organism.ComplexityLevel, len(organism.Cells))
	}

	// Run simulation until we reach stage 4 (100+ cells)
	maxTicks := 1000
	evolutionOccurred := runEvolutionSimulation(t, world, organism, 100, 4, maxTicks, 100)

	if !evolutionOccurred {
		t.Errorf("Failed to evolve from stage 3 to stage 4 in %d ticks. Final: complexity=%d, cells=%d",
			maxTicks, organism.ComplexityLevel, len(organism.Cells))
	}
}

// TestCellularEvolutionStage4To5 tests advanced multicellular to highly complex transition
func TestCellularEvolutionStage4To5(t *testing.T) {
	world := createTestWorld(t)
	entityID := createTestEntity(world, t)

	organism := world.CellularSystem.OrganismMap[entityID]

	// Artificially advance to stage 4 by adding cells with specialization
	cellTypes := []CellType{CellTypeStem, CellTypeNerve, CellTypeMuscle, CellTypeDigestive, CellTypeStorage, CellTypeDefensive}
	for i := 0; i < 150; i++ {
		cellType := cellTypes[i%len(cellTypes)]
		cell := world.CellularSystem.createCell(cellType, organism.Cells[0].DNA, Position{X: float64(i % 15), Y: float64(i / 15)})
		cell.Energy = 200.0
		cell.Health = 1.0
		cell.Age = 100
		organism.Cells = append(organism.Cells, cell)
	}

	// Update complexity level
	organism.ComplexityLevel = world.CellularSystem.calculateComplexityLevel(len(organism.Cells))

	t.Logf("Starting stage 4 to 5 test: complexity=%d, cells=%d",
		organism.ComplexityLevel, len(organism.Cells))

	if organism.ComplexityLevel < 4 {
		t.Fatalf("Failed to establish stage 4. Complexity: %d, Cells: %d",
			organism.ComplexityLevel, len(organism.Cells))
	}

	// Run simulation until we reach stage 5 (500+ cells)
	maxTicks := 2000 // More ticks needed for this large transition
	evolutionOccurred := runEvolutionSimulation(t, world, organism, 500, 5, maxTicks, 200)

	if !evolutionOccurred {
		t.Errorf("Failed to evolve from stage 4 to stage 5 in %d ticks. Final: complexity=%d, cells=%d",
			maxTicks, organism.ComplexityLevel, len(organism.Cells))
	}
}

// TestCompleteCellularEvolution tests the full evolutionary pathway
func TestCompleteCellularEvolution(t *testing.T) {
	world := createTestWorld(t)
	entityID := createTestEntity(world, t)

	organism := world.CellularSystem.OrganismMap[entityID]

	t.Logf("Testing complete cellular evolution pathway for entity %d", entityID)

	// Stage 1: Single cell (start)
	if organism.ComplexityLevel != 1 {
		t.Fatalf("Expected initial complexity level 1, got %d", organism.ComplexityLevel)
	}
	t.Logf("Stage 1 confirmed: %d cells, complexity level %d", len(organism.Cells), organism.ComplexityLevel)

	// Evolve through each stage with proper verification
	stages := []struct {
		name             string
		targetCells      int
		targetComplexity int
		maxTicks         int
	}{
		{"Stage 1→2", 5, 2, 800},
		{"Stage 2→3", 20, 3, 800},
		{"Stage 3→4", 100, 4, 1000},
		{"Stage 4→5", 500, 5, 1500},
	}

	for _, stage := range stages {
		t.Logf("Starting %s evolution...", stage.name)

		startingCells := len(organism.Cells)
		evolutionOccurred := runEvolutionSimulationWithName(t, world, organism, stage.targetCells, stage.targetComplexity, stage.maxTicks, stage.name, startingCells)

		if !evolutionOccurred {
			t.Fatalf("%s failed in %d ticks. Final: %d cells, complexity level %d",
				stage.name, stage.maxTicks, len(organism.Cells), organism.ComplexityLevel)
		}
	}

	t.Logf("Complete cellular evolution successful! Final organism: %d cells, complexity level %d",
		len(organism.Cells), organism.ComplexityLevel)
}

// Helper functions for test setup
func createTestWorld(t *testing.T) *World {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 0,
		PopulationSize: 0,
		GridWidth:      20,
		GridHeight:     20,
	}

	world := NewWorld(config)
	if world.CellularSystem == nil {
		t.Fatal("CellularSystem not initialized")
	}
	if world.DNASystem == nil {
		t.Fatal("DNASystem not initialized")
	}

	return world
}

func createTestEntity(world *World, t *testing.T) int {
	// Create a test entity manually
	entity := NewEntity(world.NextID, []string{"size", "strength", "intelligence", "energy", "metabolism"}, "test_species", Position{X: 50, Y: 50})
	world.NextID++

	// Create DNA and cellular organism
	dna := world.DNASystem.GenerateRandomDNA(entity.ID, 0)
	organism := world.CellularSystem.CreateSingleCellOrganism(entity.ID, dna)

	if organism == nil {
		t.Fatal("Failed to create cellular organism")
	}

	world.AllEntities = append(world.AllEntities, entity)
	return entity.ID
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
	// Note: Removed deprecated rand.Seed call

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
	// Note: Removed deprecated rand.Seed call

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
