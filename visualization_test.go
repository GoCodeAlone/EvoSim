package main

import (
	"strings"
	"testing"
)

func TestVisualizationSystem(t *testing.T) {
	// Create test world with visualization components
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create CLI model for testing visualization
	cli := CLIModel{
		world:          world,
		width:          80,
		height:         25,
		selectedView:   "cellular",
		viewModes:      []string{"grid", "stats", "events", "populations", "communication", "civilization", "physics", "wind", "species", "network", "dna", "cellular", "evolution", "topology"},
		speciesColors:  make(map[string]string),
		speciesSymbols: make(map[string]rune),
	}

	// Test cellular view rendering
	t.Run("CellularVisualization", func(t *testing.T) {
		cellularView := cli.cellularView()

		if !strings.Contains(cellularView, "Cellular Analysis & Visualization") {
			t.Error("Cellular view should contain enhanced title")
		}

		if !strings.Contains(cellularView, "INDIVIDUAL ORGANISM VISUALIZATION") {
			t.Error("Cellular view should include individual organism visualization section")
		}

		if len(cellularView) < 100 {
			t.Error("Enhanced cellular view should be substantially longer than basic view")
		}
	})

	// Test species view rendering
	t.Run("SpeciesVisualization", func(t *testing.T) {
		speciesView := cli.speciesView()

		if !strings.Contains(speciesView, "INDIVIDUAL SPECIES DETAILS") {
			t.Error("Species view should include individual species details section")
		}

		if !strings.Contains(speciesView, "Species Evolution & Tracking") {
			t.Error("Species view should contain enhanced title")
		}

		if len(speciesView) < 200 {
			t.Error("Enhanced species view should be substantially longer")
		}
	})

	// Test topology view rendering
	t.Run("TopologyVisualization", func(t *testing.T) {
		topologyView := cli.topologyView()

		if !strings.Contains(topologyView, "Underground Visualization") {
			t.Error("Topology view should include underground visualization")
		}

		if !strings.Contains(topologyView, "VIEWING ANGLES") {
			t.Error("Topology view should include viewing angles section")
		}

		if !strings.Contains(topologyView, "ENHANCED TOPOGRAPHIC MAP") {
			t.Error("Topology view should include enhanced topographic map")
		}

		if !strings.Contains(topologyView, "UNDERGROUND FEATURES") {
			t.Error("Topology view should include underground features section")
		}

		if !strings.Contains(topologyView, "CROSS-SECTION VIEW") {
			t.Error("Topology view should include cross-section view")
		}
	})

	// Test helper functions
	t.Run("VisualizationHelpers", func(t *testing.T) {
		// Test elevation symbol generation
		symbol := cli.getElevationSymbol(0.9)
		if symbol != "▲" {
			t.Errorf("High elevation should return mountain symbol, got %s", symbol)
		}

		symbol = cli.getElevationSymbol(0.1)
		if symbol != "." {
			t.Errorf("Low elevation should return plains symbol, got %s", symbol)
		}

		symbol = cli.getElevationSymbol(-0.5)
		if symbol != "≈" {
			t.Errorf("Negative elevation should return water symbol, got %s", symbol)
		}

		// Test terrain feature symbols
		terrainSymbol := cli.getTerrainFeatureSymbol(2) // Mountain
		if terrainSymbol != "▲" {
			t.Errorf("Mountain terrain should return triangle symbol, got %s", terrainSymbol)
		}

		// Test water body symbols
		waterSymbol := cli.getWaterBodySymbol("river")
		if waterSymbol != "≈" {
			t.Errorf("River should return wave symbol, got %s", waterSymbol)
		}
	})

	// Test cell symbol mapping
	t.Run("CellSymbolMapping", func(t *testing.T) {
		symbol := cli.getCellSymbol(0) // Stem
		if symbol != "S" {
			t.Errorf("Stem cell should return 'S', got %s", symbol)
		}

		symbol = cli.getCellSymbol(1) // Nerve
		if symbol != "N" {
			t.Errorf("Nerve cell should return 'N', got %s", symbol)
		}

		symbol = cli.getCellSymbol(999) // Invalid
		if symbol != "?" {
			t.Errorf("Invalid cell type should return '?', got %s", symbol)
		}
	})
}

func TestTopographicMapRendering(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	cli := CLIModel{
		world: world,
	}

	// Test topographic map rendering
	topoMap := cli.renderTopographicMap()

	if !strings.Contains(topoMap, "Surface Elevation Map") {
		t.Error("Topographic map should include title")
	}

	if !strings.Contains(topoMap, "Elevation Legend") {
		t.Error("Topographic map should include legend")
	}

	if len(topoMap) < 50 {
		t.Error("Topographic map should contain substantial content")
	}
}

func TestUndergroundMapRendering(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	cli := CLIModel{
		world: world,
	}

	// Test underground map rendering
	undergroundMap := cli.renderUndergroundMap()

	if !strings.Contains(undergroundMap, "Underground Structure Map") {
		t.Error("Underground map should include title")
	}

	// Should either show structures or indicate system unavailable
	if !strings.Contains(undergroundMap, "Underground Legend") && !strings.Contains(undergroundMap, "not available") {
		t.Error("Underground map should show legend or unavailable message")
	}
}

func TestCrossSectionRendering(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	cli := CLIModel{
		world: world,
	}

	// Test cross-section rendering
	crossSection := cli.renderCrossSectionView()

	if !strings.Contains(crossSection, "World Cross-Section") {
		t.Error("Cross-section should include title")
	}

	if !strings.Contains(crossSection, "Surface:") {
		t.Error("Cross-section should show surface layer")
	}

	if !strings.Contains(crossSection, "Layer") {
		t.Error("Cross-section should show underground layers")
	}

	if !strings.Contains(crossSection, "Cross-Section Legend") {
		t.Error("Cross-section should include legend")
	}
}

func TestOrganismVisualization(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	cli := CLIModel{
		world: world,
	}

	// Test organism selection
	organism := cli.getSelectedOrganism()

	// Should return nil when no organisms exist
	if organism != nil {
		t.Error("Should return nil when no organisms exist")
	}
}

func TestSpeciesDetailRendering(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	cli := CLIModel{
		world: world,
	}

	// Create mock species data
	mockSpecies := map[string]interface{}{
		"id":                 1,
		"name":               "TestGrass-S1",
		"origin_type":        PlantGrass,
		"current_population": 5,
	}

	// Test species detail rendering
	detail := cli.renderSpeciesDetail(mockSpecies)

	if !strings.Contains(detail, "TestGrass-S1") {
		t.Error("Species detail should contain species name")
	}

	if !strings.Contains(detail, "Species Visual Representation") {
		t.Error("Species detail should include visual representation")
	}

	if !strings.Contains(detail, "Genetic Trait Analysis") {
		t.Error("Species detail should include trait analysis")
	}

	if !strings.Contains(detail, "Environmental Adaptation") {
		t.Error("Species detail should include habitat information")
	}
}
