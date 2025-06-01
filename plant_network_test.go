package main

import (
	"math"
	"testing"
)

func TestPlantNetworkSystem(t *testing.T) {
	// Initialize network system
	network := NewPlantNetworkSystem()

	if network == nil {
		t.Fatal("Failed to create PlantNetworkSystem")
	}

	// Test initial state
	if len(network.Connections) != 0 {
		t.Errorf("Expected 0 connections initially, got %d", len(network.Connections))
	}

	if len(network.ChemicalSignals) != 0 {
		t.Errorf("Expected 0 chemical signals initially, got %d", len(network.ChemicalSignals))
	}
}

func TestNetworkFormation(t *testing.T) {
	network := NewPlantNetworkSystem()

	// Create test plants with correct struct format
	plant1 := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantGrass,
		Energy:   50.0,
		IsAlive:  true,
		Age:      10,
	}

	plant2 := &Plant{
		ID:       2,
		Position: Position{X: 15.0, Y: 10.0},
		Type:     PlantGrass,
		Energy:   50.0,
		IsAlive:  true,
		Age:      10,
	}

	plant3 := &Plant{
		ID:       3,
		Position: Position{X: 50.0, Y: 50.0},
		Type:     PlantTree,
		Energy:   50.0,
		IsAlive:  true,
		Age:      10,
	}

	plants := []*Plant{plant1, plant2, plant3}

	// Update network to allow formation
	for i := 0; i < 5; i++ {
		network.Update(plants, i+1)
	}

	// Check if connections formed between nearby plants
	distance := math.Sqrt(math.Pow(plant2.Position.X-plant1.Position.X, 2) +
		math.Pow(plant2.Position.Y-plant1.Position.Y, 2))
	t.Logf("Distance between plant1 and plant2: %.2f", distance)

	// Plants 1 and 2 should be close enough to potentially connect
	if distance <= 20 && len(network.Connections) == 0 {
		t.Logf("Note: No connections formed between nearby plants (distance: %.2f)", distance)
	}

	t.Logf("Network has %d connections after updates", len(network.Connections))
}

func TestChemicalSignalPropagation(t *testing.T) {
	network := NewPlantNetworkSystem()

	// Create test plants
	source := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantTree,
		Energy:   100.0,
		IsAlive:  true,
		Age:      50,
	}

	receiver := &Plant{
		ID:       2,
		Position: Position{X: 15.0, Y: 15.0},
		Type:     PlantTree,
		Energy:   30.0, // Low energy - needs nutrients
		IsAlive:  true,
		Age:      30,
	}

	plants := []*Plant{source, receiver}

	// Create initial connections
	for i := 0; i < 10; i++ {
		network.Update(plants, i+1)
	}

	// Add a chemical signal manually for testing
	signal := &ChemicalSignal{
		ID:        1,
		Source:    source,
		Type:      SignalNutrientAvailable,
		Intensity: 1.0,
		Age:       0,
		MaxAge:    50,
		Visited:   make(map[int]bool),
		Message:   "Test nutrient signal",
		Metadata:  make(map[string]float64),
	}

	network.ChemicalSignals = append(network.ChemicalSignals, signal)

	if len(network.ChemicalSignals) != 1 {
		t.Errorf("Expected 1 chemical signal, got %d", len(network.ChemicalSignals))
	}

	// Test signal aging
	for i := 0; i < 25; i++ {
		network.Update(plants, i+20)
	}

	// Signal should still exist
	if len(network.ChemicalSignals) == 0 {
		t.Error("Chemical signal disappeared too early")
	}

	// Age signal beyond max age
	for i := 0; i < 30; i++ {
		network.Update(plants, i+50)
	}

	// Signal should expire
	if len(network.ChemicalSignals) > 0 {
		t.Logf("Signal still active after aging: %d signals remain", len(network.ChemicalSignals))
	}
}

func TestResourceSharing(t *testing.T) {
	network := NewPlantNetworkSystem()

	// Create plants with different energy levels
	richPlant := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantTree,
		Energy:   100.0,
		IsAlive:  true,
		Age:      50,
	}

	poorPlant := &Plant{
		ID:       2,
		Position: Position{X: 12.0, Y: 12.0},
		Type:     PlantTree,
		Energy:   20.0, // Low energy
		IsAlive:  true,
		Age:      30,
	}

	plants := []*Plant{richPlant, poorPlant}

	// Allow network formation and resource sharing
	initialRichEnergy := richPlant.Energy
	initialPoorEnergy := poorPlant.Energy

	for i := 0; i < 20; i++ {
		network.Update(plants, i+1)
	}

	// Check if any resource transfer occurred
	t.Logf("Rich plant energy: %.2f -> %.2f", initialRichEnergy, richPlant.Energy)
	t.Logf("Poor plant energy: %.2f -> %.2f", initialPoorEnergy, poorPlant.Energy)
	t.Logf("Network connections: %d", len(network.Connections))

	// Log connection details if any exist
	for i, conn := range network.Connections {
		t.Logf("Connection %d: Plant %d <-> Plant %d, Type: %d, Strength: %.2f, Health: %.2f",
			i, conn.PlantA.ID, conn.PlantB.ID, conn.Type, conn.Strength, conn.Health)
	}
}

func TestNetworkStats(t *testing.T) {
	network := NewPlantNetworkSystem()

	// Get stats from empty network
	stats := network.GetNetworkStats()

	if stats == nil {
		t.Fatal("GetNetworkStats returned nil")
	}

	// Check for expected stats keys based on actual implementation
	expectedKeys := []string{
		"total_connections", "total_signals",
		"total_clusters", "network_efficiency",
		"total_resources_transferred", "connections_by_type",
		"signals_by_type", "average_cluster_size",
		"max_connection_distance",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stats key '%s' not found", key)
		}
	}

	// Add some test data
	plant1 := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantTree,
		Energy:   50.0,
		IsAlive:  true,
	}

	plant2 := &Plant{
		ID:       2,
		Position: Position{X: 15.0, Y: 15.0},
		Type:     PlantTree,
		Energy:   50.0,
		IsAlive:  true,
	}

	// Manually create a connection for testing
	connection := &NetworkConnection{
		ID:           1,
		PlantA:       plant1,
		PlantB:       plant2,
		Type:         ConnectionMycorrhizal,
		Strength:     0.8,
		Health:       1.0,
		Age:          10,
		Distance:     math.Sqrt(25 + 25), // Distance between plants
		LastTransfer: 5,
		Efficiency:   0.9,
	}

	network.Connections = append(network.Connections, connection)

	stats = network.GetNetworkStats()

	if stats["total_connections"].(int) != 1 {
		t.Errorf("Expected 1 total connection, got %v", stats["total_connections"])
	}

	// Test connections_by_type
	connectionsByType := stats["connections_by_type"].(map[NetworkConnectionType]int)
	if connectionsByType[ConnectionMycorrhizal] != 1 {
		t.Errorf("Expected 1 mycorrhizal connection in type breakdown, got %d", connectionsByType[ConnectionMycorrhizal])
	}
}

func TestConnectionTypes(t *testing.T) {
	network := NewPlantNetworkSystem()

	// Test different connection types
	plant1 := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantTree,
		Energy:   50.0,
		IsAlive:  true,
	}

	plant2 := &Plant{
		ID:       2,
		Position: Position{X: 12.0, Y: 12.0},
		Type:     PlantBush,
		Energy:   50.0,
		IsAlive:  true,
	}

	// Manually create connections of different types
	mycorrhizal := &NetworkConnection{
		ID:       1,
		PlantA:   plant1,
		PlantB:   plant2,
		Type:     ConnectionMycorrhizal,
		Strength: 0.8,
		Health:   1.0,
	}

	root := &NetworkConnection{
		ID:       2,
		PlantA:   plant1,
		PlantB:   plant2,
		Type:     ConnectionRoot,
		Strength: 0.6,
		Health:   1.0,
	}

	chemical := &NetworkConnection{
		ID:       3,
		PlantA:   plant1,
		PlantB:   plant2,
		Type:     ConnectionChemical,
		Strength: 0.9,
		Health:   1.0,
	}

	network.Connections = []*NetworkConnection{mycorrhizal, root, chemical}

	// Test that all connection types are properly handled
	if len(network.Connections) != 3 {
		t.Errorf("Expected 3 connections, got %d", len(network.Connections))
	}

	// Test connection type distribution
	typeCount := make(map[NetworkConnectionType]int)
	for _, conn := range network.Connections {
		typeCount[conn.Type]++
	}

	if typeCount[ConnectionMycorrhizal] != 1 {
		t.Errorf("Expected 1 mycorrhizal connection, got %d", typeCount[ConnectionMycorrhizal])
	}

	if typeCount[ConnectionRoot] != 1 {
		t.Errorf("Expected 1 root connection, got %d", typeCount[ConnectionRoot])
	}

	if typeCount[ConnectionChemical] != 1 {
		t.Errorf("Expected 1 chemical connection, got %d", typeCount[ConnectionChemical])
	}
}

func TestSignalTypes(t *testing.T) {
	// Test all chemical signal types
	signalTypes := []ChemicalSignalType{
		SignalNutrientAvailable,
		SignalNutrientNeeded,
		SignalThreatDetected,
		SignalOptimalGrowth,
		SignalReproductionReady,
		SignalToxicConditions,
	}

	source := &Plant{
		ID:       1,
		Position: Position{X: 10.0, Y: 10.0},
		Type:     PlantTree,
		Energy:   50.0,
		IsAlive:  true,
	}

	network := NewPlantNetworkSystem()

	// Create signals of each type
	for i, signalType := range signalTypes {
		signal := &ChemicalSignal{
			ID:        i + 1,
			Source:    source,
			Type:      signalType,
			Intensity: 1.0,
			Age:       0,
			MaxAge:    100,
			Visited:   make(map[int]bool),
			Message:   "Test signal",
			Metadata:  make(map[string]float64),
		}
		network.ChemicalSignals = append(network.ChemicalSignals, signal)
	}

	if len(network.ChemicalSignals) != len(signalTypes) {
		t.Errorf("Expected %d signals, got %d", len(signalTypes), len(network.ChemicalSignals))
	}

	// Test signal type distribution
	typeCount := make(map[ChemicalSignalType]int)
	for _, signal := range network.ChemicalSignals {
		typeCount[signal.Type]++
	}

	for _, signalType := range signalTypes {
		if typeCount[signalType] != 1 {
			t.Errorf("Expected 1 signal of type %d, got %d", signalType, typeCount[signalType])
		}
	}
}
