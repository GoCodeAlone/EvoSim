package main

import (
	"fmt"
	"testing"
)

func TestEcosystemMonitor(t *testing.T) {
	// Create test world
	config := WorldConfig{
		Width:          20,
		Height:         20,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 5,
	}
	world := NewWorld(config)

	// Add some entities and plants
	popConfig := PopulationConfig{
		Name:       "TestPop",
		Species:    "TestSpecies",
		BaseTraits: map[string]float64{"energy": 100, "speed": 1.0},
		StartPos:   Position{X: 5, Y: 5},
		Spread:     2.0,
	}
	world.AddPopulation(popConfig)

	// Add some plants
	for i := 0; i < 5; i++ {
		plant := NewPlant(i+1, PlantGrass, Position{X: float64(i * 2), Y: float64(i * 2)})
		world.AllPlants = append(world.AllPlants, plant)
	}

	// Update ecosystem metrics
	world.EcosystemMonitor.UpdateMetrics(world)

	// Check that metrics were calculated
	metrics := world.EcosystemMonitor.CurrentMetrics

	if metrics.TotalPopulation <= 0 {
		t.Errorf("Expected positive total population, got %d", metrics.TotalPopulation)
	}

	if metrics.SpeciesRichness <= 0 {
		t.Errorf("Expected positive species richness, got %d", metrics.SpeciesRichness)
	}

	// Shannon diversity should be >= 0
	if metrics.ShannonDiversity < 0 {
		t.Errorf("Expected non-negative Shannon diversity, got %f", metrics.ShannonDiversity)
	}

	// Simpson diversity should be between 0 and 1
	if metrics.SimpsonDiversity < 0 || metrics.SimpsonDiversity > 1 {
		t.Errorf("Expected Simpson diversity between 0 and 1, got %f", metrics.SimpsonDiversity)
	}
}

func TestDiversityIndices(t *testing.T) {
	monitor := NewEcosystemMonitor(100)

	// Create test world with diverse populations
	config := WorldConfig{
		Width:          20,
		Height:         20,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 10,
	}
	world := NewWorld(config)

	// Add multiple species with different populations
	species := []string{"Species1", "Species2", "Species3"}

	for i, speciesName := range species {
		popConfig := PopulationConfig{
			Name:       speciesName,
			Species:    speciesName,
			BaseTraits: map[string]float64{"energy": 100, "speed": 1.0},
			StartPos:   Position{X: float64(i * 3), Y: float64(i * 3)},
			Spread:     1.0,
		}
		world.AddPopulation(popConfig)
	}

	// Update metrics
	monitor.UpdateMetrics(world)
	metrics := monitor.CurrentMetrics

	// Check diversity calculations
	if metrics.SpeciesRichness < 1 {
		t.Errorf("Expected at least 1 species, got %d", metrics.SpeciesRichness)
	}

	// Shannon diversity should be non-negative
	if metrics.ShannonDiversity < 0 {
		t.Errorf("Expected non-negative Shannon diversity, got %f", metrics.ShannonDiversity)
	}

	// Simpson diversity should be between 0 and 1
	if metrics.SimpsonDiversity < 0 || metrics.SimpsonDiversity > 1 {
		t.Errorf("Expected Simpson diversity between 0 and 1, got %f", metrics.SimpsonDiversity)
	}
}

func TestEcosystemHealthScore(t *testing.T) {
	monitor := NewEcosystemMonitor(100)

	// Create healthy ecosystem
	config := WorldConfig{
		Width:          20,
		Height:         20,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 8,
	}
	world := NewWorld(config)

	// Add diverse populations
	for i := 0; i < 3; i++ {
		popConfig := PopulationConfig{
			Name:       fmt.Sprintf("Species%d", i),
			Species:    fmt.Sprintf("Species%d", i),
			BaseTraits: map[string]float64{"energy": 100, "speed": 1.0},
			StartPos:   Position{X: float64(i * 5), Y: float64(i * 5)},
			Spread:     2.0,
		}
		world.AddPopulation(popConfig)
	}

	// Update metrics multiple times to establish trends
	for tick := 0; tick < 5; tick++ {
		monitor.UpdateMetrics(world)
	}

	// Check health score
	healthScore := monitor.GetHealthScore()
	if healthScore < 0 || healthScore > 100 {
		t.Errorf("Expected health score between 0 and 100, got %f", healthScore)
	}

	// With good diversity, should have reasonable health score
	if healthScore < 10 {
		t.Errorf("Expected higher health score for diverse ecosystem, got %f", healthScore)
	}
}

func TestEcosystemTrends(t *testing.T) {
	monitor := NewEcosystemMonitor(100)

	config := WorldConfig{
		Width:          20,
		Height:         20,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 5,
	}
	world := NewWorld(config)

	// Add initial population
	popConfig := PopulationConfig{
		Name:       "TestSpecies",
		Species:    "TestSpecies",
		BaseTraits: map[string]float64{"energy": 100, "speed": 1.0},
		StartPos:   Position{X: 5, Y: 5},
		Spread:     2.0,
	}
	world.AddPopulation(popConfig)

	// Take initial measurements
	monitor.UpdateMetrics(world)
	monitor.UpdateMetrics(world)

	// Add more entities (simulate population growth) - change population size config
	world.Config.PopulationSize = 10
	popConfig2 := PopulationConfig{
		Name:       "TestSpecies2",
		Species:    "TestSpecies2",
		BaseTraits: map[string]float64{"energy": 100, "speed": 1.0},
		StartPos:   Position{X: 8, Y: 8},
		Spread:     2.0,
	}
	world.AddPopulation(popConfig2)
	monitor.UpdateMetrics(world)

	trends := monitor.GetTrends()

	// Should detect some trends (might be insufficient_data initially)
	validTrends := []string{"increasing", "decreasing", "stable", "growing", "declining", "improving", "degrading", "insufficient_data"}

	for metric, trend := range trends {
		found := false
		for _, validTrend := range validTrends {
			if trend == validTrend {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Invalid trend for %s: %s", metric, trend)
		}
	}
}

func TestNetworkConnectivity(t *testing.T) {
	monitor := NewEcosystemMonitor(100)

	config := WorldConfig{
		Width:          20,
		Height:         20,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 5,
	}
	world := NewWorld(config)

	// Add plants close together to encourage network formation
	for i := 0; i < 5; i++ {
		plant := NewPlant(i+1, PlantGrass, Position{X: float64(i), Y: float64(i)})
		world.AllPlants = append(world.AllPlants, plant)
	}

	// Update metrics
	monitor.UpdateMetrics(world)
	metrics := monitor.CurrentMetrics

	// Network connectivity should be between 0 and 1
	if metrics.NetworkConnectivity < 0 || metrics.NetworkConnectivity > 1 {
		t.Errorf("Expected network connectivity between 0 and 1, got %f", metrics.NetworkConnectivity)
	}

	// With no connections established, should be 0
	if metrics.NetworkConnectivity != 0 {
		t.Logf("Network connectivity: %f (expected 0 for new network)", metrics.NetworkConnectivity)
	}
}
