package main

import (
	"testing"
	"time"
)

func TestStatisticalReporter(t *testing.T) {
	// Create a statistical reporter
	reporter := NewStatisticalReporter(100, 10, 5, 10)

	if reporter == nil {
		t.Fatal("Failed to create statistical reporter")
	}

	if len(reporter.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(reporter.Events))
	}

	if len(reporter.Snapshots) != 0 {
		t.Errorf("Expected 0 snapshots, got %d", len(reporter.Snapshots))
	}

	if len(reporter.Anomalies) != 0 {
		t.Errorf("Expected 0 anomalies, got %d", len(reporter.Anomalies))
	}
}

func TestStatisticalEventLogging(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Create a test entity
	entity := NewEntity(1, []string{"speed", "size", "energy"}, "test_species", Position{X: 10, Y: 10})
	
	// Log an entity event
	reporter.LogEntityEvent(1, "test_event", entity, 50.0, 75.0, nil)
	
	if len(reporter.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(reporter.Events))
	}
	
	event := reporter.Events[0]
	if event.EventType != "test_event" {
		t.Errorf("Expected event type 'test_event', got '%s'", event.EventType)
	}
	
	if event.Category != "entity" {
		t.Errorf("Expected category 'entity', got '%s'", event.Category)
	}
	
	if event.EntityID != 1 {
		t.Errorf("Expected entity ID 1, got %d", event.EntityID)
	}
	
	if event.Change != 25.0 {
		t.Errorf("Expected change 25.0, got %f", event.Change)
	}
}

func TestStatisticalSnapshot(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Create a test world
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      20,
		GridHeight:     20,
	}
	world := NewWorld(config)
	
	// Clear existing entities and plants to have a clean test
	world.AllEntities = []*Entity{}
	world.AllPlants = []*Plant{}
	
	// Add some test entities
	for i := 0; i < 3; i++ {
		entity := NewEntity(i, []string{"speed", "size", "energy"}, "test_species", Position{X: float64(i * 10), Y: float64(i * 10)})
		entity.Energy = 50.0 + float64(i*10)
		world.AllEntities = append(world.AllEntities, entity)
	}
	
	// Take a snapshot
	reporter.TakeSnapshot(world)
	
	if len(reporter.Snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(reporter.Snapshots))
	}
	
	snapshot := reporter.Snapshots[0]
	if snapshot.TotalEntities != 3 {
		t.Errorf("Expected 3 entities, got %d", snapshot.TotalEntities)
	}
	
	expectedEnergy := 50.0 + 60.0 + 70.0 // Sum of entity energies
	if snapshot.TotalEnergy < expectedEnergy-1 || snapshot.TotalEnergy > expectedEnergy+1 {
		t.Errorf("Expected total energy around %f, got %f", expectedEnergy, snapshot.TotalEnergy)
	}
}

func TestAnomalyDetection(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Create a test world
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      20,
		GridHeight:     20,
	}
	world := NewWorld(config)
	
	// Set up baseline energy
	reporter.totalEnergyBaseline = 1000.0
	
	// Create snapshots with energy conservation violation
	snapshot1 := StatisticalSnapshot{
		Tick:        1,
		Timestamp:   time.Now(),
		TotalEnergy: 1000.0,
	}
	reporter.addSnapshot(snapshot1)
	
	snapshot2 := StatisticalSnapshot{
		Tick:        2,
		Timestamp:   time.Now(),
		TotalEnergy: 1500.0, // Large energy increase (50%)
	}
	reporter.addSnapshot(snapshot2)
	
	// Run analysis
	anomalies := reporter.PerformAnalysis(world)
	
	if len(anomalies) == 0 {
		t.Error("Expected to detect energy conservation anomaly")
	}
	
	// Check if energy conservation anomaly was detected
	foundEnergyAnomaly := false
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyEnergyConservation {
			foundEnergyAnomaly = true
			if anomaly.Severity <= 0 {
				t.Errorf("Expected positive severity, got %f", anomaly.Severity)
			}
		}
	}
	
	if !foundEnergyAnomaly {
		t.Error("Expected to find energy conservation anomaly")
	}
}

func TestTraitDistributionAnalysis(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Create a snapshot with unrealistic trait distribution (all same value)
	traitDistributions := make(map[string][]float64)
	uniformValues := make([]float64, 20)
	for i := range uniformValues {
		uniformValues[i] = 0.5 // All exactly the same
	}
	traitDistributions["speed"] = uniformValues
	
	snapshot := StatisticalSnapshot{
		Tick:               1,
		Timestamp:          time.Now(),
		TraitDistributions: traitDistributions,
	}
	reporter.addSnapshot(snapshot)
	
	// Run trait distribution analysis
	anomalies := reporter.analyzeTraitDistributions()
	
	if len(anomalies) == 0 {
		t.Error("Expected to detect unrealistic distribution anomaly")
	}
	
	// Check if unrealistic distribution anomaly was detected
	foundDistributionAnomaly := false
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyUnrealisticDistribution {
			foundDistributionAnomaly = true
		}
	}
	
	if !foundDistributionAnomaly {
		t.Error("Expected to find unrealistic distribution anomaly")
	}
}

func TestExportFunctionality(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Add some test data
	reporter.LogSystemEvent(1, "test_event", "test description", map[string]interface{}{"test": "data"})
	
	// Test JSON export
	err := reporter.ExportToJSON("/tmp/test_export.json")
	if err != nil {
		t.Errorf("Failed to export JSON: %v", err)
	}
	
	// Test CSV export
	err = reporter.ExportToCSV("/tmp/test_export.csv")
	if err != nil {
		t.Errorf("Failed to export CSV: %v", err)
	}
}

func TestSummaryStatistics(t *testing.T) {
	reporter := NewStatisticalReporter(100, 10, 5, 10)
	
	// Add some test data
	reporter.totalEnergyBaseline = 1000.0
	
	snapshot := StatisticalSnapshot{
		Tick:          5,
		Timestamp:     time.Now(),
		TotalEntities: 10,
		TotalPlants:   20,
		TotalEnergy:   1100.0,
		SpeciesCount:  3,
	}
	reporter.addSnapshot(snapshot)
	
	// Get summary statistics
	summary := reporter.GetSummaryStatistics()
	
	if summary["latest_tick"] != 5 {
		t.Errorf("Expected latest tick 5, got %v", summary["latest_tick"])
	}
	
	if summary["total_entities"] != 10 {
		t.Errorf("Expected 10 entities, got %v", summary["total_entities"])
	}
	
	if summary["energy_baseline"] != 1000.0 {
		t.Errorf("Expected energy baseline 1000.0, got %v", summary["energy_baseline"])
	}
}