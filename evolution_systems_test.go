package main

import (
	"testing"
)

func TestDNASystem(t *testing.T) {
	dnaSystem := NewDNASystem(NewCentralEventBus(1000))
	
	// Test DNA generation
	dna := dnaSystem.GenerateRandomDNA(1, 0)
	
	if dna.EntityID != 1 {
		t.Errorf("Expected EntityID 1, got %d", dna.EntityID)
	}
	
	if dna.Generation != 0 {
		t.Errorf("Expected Generation 0, got %d", dna.Generation)
	}
	
	if len(dna.Chromosomes) != 2 {
		t.Errorf("Expected 2 chromosomes, got %d", len(dna.Chromosomes))
	}
	
	// Test trait expression
	traitValue := dnaSystem.ExpressTrait(dna, "size")
	if traitValue < -1.0 || traitValue > 1.0 {
		t.Errorf("Trait value out of range: %f", traitValue)
	}
}

func TestCellularSystem(t *testing.T) {
	dnaSystem := NewDNASystem(NewCentralEventBus(1000))
	cellularSystem := NewCellularSystem(dnaSystem, NewCentralEventBus(1000))
	
	// Create DNA and cellular organism
	dna := dnaSystem.GenerateRandomDNA(1, 0)
	organism := cellularSystem.CreateSingleCellOrganism(1, dna)
	
	if organism.EntityID != 1 {
		t.Errorf("Expected EntityID 1, got %d", organism.EntityID)
	}
	
	if organism.ComplexityLevel != 1 {
		t.Errorf("Expected complexity level 1, got %d", organism.ComplexityLevel)
	}
	
	if len(organism.Cells) != 1 {
		t.Errorf("Expected 1 cell, got %d", len(organism.Cells))
	}
	
	// Test system update
	cellularSystem.UpdateCellularOrganisms()
	
	// Check organism still exists
	if _, exists := cellularSystem.OrganismMap[1]; !exists {
		t.Error("Organism disappeared after update")
	}
}

func TestMacroEvolutionSystem(t *testing.T) {
	macroEvolution := NewMacroEvolutionSystem()
	
	if len(macroEvolution.Events) != 0 {
		t.Errorf("Expected no initial events, got %d", len(macroEvolution.Events))
	}
	
	if len(macroEvolution.SpeciesLineages) != 0 {
		t.Errorf("Expected no initial lineages, got %d", len(macroEvolution.SpeciesLineages))
	}
	
	// Test stats
	stats := macroEvolution.GetEvolutionStats()
	if stats == nil {
		t.Error("Expected stats, got nil")
	}
	
	if stats["total_events"] != 0 {
		t.Errorf("Expected 0 events, got %v", stats["total_events"])
	}
}

func TestTopologySystem(t *testing.T) {
	topology := NewTopologySystem(10, 10)
	
	if topology.Width != 10 || topology.Height != 10 {
		t.Errorf("Expected 10x10 grid, got %dx%d", topology.Width, topology.Height)
	}
	
	if len(topology.TopologyGrid) != 10 {
		t.Errorf("Expected 10 rows, got %d", len(topology.TopologyGrid))
	}
	
	if len(topology.TopologyGrid[0]) != 10 {
		t.Errorf("Expected 10 columns, got %d", len(topology.TopologyGrid[0]))
	}
	
	// Test terrain generation
	topology.GenerateInitialTerrain()
	
	// Check that some features were created
	if len(topology.TerrainFeatures) == 0 {
		t.Error("Expected terrain features to be generated")
	}
	
	if len(topology.WaterBodies) == 0 {
		t.Error("Expected water bodies to be generated")
	}
	
	// Test update
	topology.UpdateTopology(1)
	
	// Test stats
	stats := topology.GetTopologyStats()
	if stats == nil {
		t.Error("Expected stats, got nil")
	}
}

func TestNewSystemsIntegration(t *testing.T) {
	// Test that new systems integrate with world
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      10,
		GridHeight:     10,
	}
	
	world := NewWorld(config)
	
	// Check that all systems are initialized
	if world.DNASystem == nil {
		t.Error("DNA system not initialized")
	}
	
	if world.CellularSystem == nil {
		t.Error("Cellular system not initialized")
	}
	
	if world.MacroEvolutionSystem == nil {
		t.Error("Macro evolution system not initialized")
	}
	
	if world.TopologySystem == nil {
		t.Error("Topology system not initialized")
	}
	
	// Add a population and check DNA/cellular integration
	populationConfig := PopulationConfig{
		Name:             "test",
		Species:          "testSpecies",
		BaseTraits:       map[string]float64{"size": 0.5, "strength": 0.3},
		StartPos:         Position{X: 50, Y: 50},
		Spread:           10.0,
		Color:            "blue",
		BaseMutationRate: 0.01,
	}
	
	world.AddPopulation(populationConfig)
	
	// Check that entities have cellular organisms
	if len(world.AllEntities) == 0 {
		t.Error("No entities created")
	}
	
	entity := world.AllEntities[0]
	if organism := world.CellularSystem.OrganismMap[entity.ID]; organism == nil {
		t.Error("Entity does not have corresponding cellular organism")
	}
	
	// Test world update
	world.Update()
	
	// Check that macro evolution is tracking
	if world.MacroEvolutionSystem.CurrentTick != world.Tick {
		t.Error("Macro evolution system not tracking world ticks")
	}
}