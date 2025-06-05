package main

import (
	"math"
	"testing"
)

// TestNeuralAISystemCreation tests basic neural AI system creation
func TestNeuralAISystemCreation(t *testing.T) {
	system := NewNeuralAISystem()

	if system == nil {
		t.Error("Failed to create neural AI system")
	}

	if system.BaseLearningRate != 0.01 {
		t.Errorf("Expected base learning rate 0.01, got %f", system.BaseLearningRate)
	}

	if system.NetworkComplexity != 10 {
		t.Errorf("Expected network complexity 10, got %d", system.NetworkComplexity)
	}

	if len(system.EntityNetworks) != 0 {
		t.Errorf("Expected 0 entity networks, got %d", len(system.EntityNetworks))
	}
}

// TestNeuralNetworkCreation tests creating neural networks for entities
func TestNeuralNetworkCreation(t *testing.T) {
	system := NewNeuralAISystem()

	// Create test entity with high intelligence
	entity := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("curiosity", 0.6)
	entity.IsAlive = true

	// Create neural network
	network := system.CreateNeuralNetwork(entity, 0)

	if network == nil {
		t.Error("Failed to create neural network")
	}

	if network.EntityID != entity.ID {
		t.Errorf("Expected entity ID %d, got %d", entity.ID, network.EntityID)
	}

	if network.Type != FeedForward {
		t.Errorf("Expected FeedForward network type, got %d", network.Type)
	}

	// Check network structure
	if len(network.InputNeurons) != 5 {
		t.Errorf("Expected 5 input neurons, got %d", len(network.InputNeurons))
	}

	if len(network.OutputNeurons) != 3 {
		t.Errorf("Expected 3 output neurons, got %d", len(network.OutputNeurons))
	}

	if len(network.Neurons) == 0 {
		t.Error("Network should have neurons")
	}

	// Check that network is stored in system
	if system.EntityNetworks[entity.ID] != network {
		t.Error("Network should be stored in system")
	}
}

// TestNeuralDecisionProcessing tests neural decision making
func TestNeuralDecisionProcessing(t *testing.T) {
	system := NewNeuralAISystem()

	// Create test entity
	entity := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.7)
	entity.SetTrait("curiosity", 0.5)
	entity.IsAlive = true

	// Create environmental inputs (vision, energy, threat, food, social)
	environmentInputs := []float64{0.5, 0.8, 0.1, 0.6, 0.3}

	// Process decision
	outputs := system.ProcessNeuralDecision(entity, environmentInputs, 0)

	if len(outputs) != 3 {
		t.Errorf("Expected 3 outputs, got %d", len(outputs))
	}

	// Check that network was created
	network := system.EntityNetworks[entity.ID]
	if network == nil {
		t.Error("Network should be created during decision processing")
	}

	// Check that decision was recorded
	if network.TotalDecisions != 1 {
		t.Errorf("Expected 1 total decision, got %d", network.TotalDecisions)
	}

	if len(network.RecentInputs) != 1 {
		t.Errorf("Expected 1 recent input, got %d", len(network.RecentInputs))
	}

	if len(network.RecentOutputs) != 1 {
		t.Errorf("Expected 1 recent output, got %d", len(network.RecentOutputs))
	}
}

// TestNeuralLearning tests the learning mechanism
func TestNeuralLearning(t *testing.T) {
	system := NewNeuralAISystem()

	// Create test entity
	entity := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("curiosity", 0.6)
	entity.IsAlive = true

	// Create network and make a decision
	environmentInputs := []float64{0.5, 0.8, 0.1, 0.6, 0.3}
	system.ProcessNeuralDecision(entity, environmentInputs, 0)

	network := system.EntityNetworks[entity.ID]
	initialExperience := network.Experience
	initialCorrectDecisions := network.CorrectDecisions

	// Learn from successful outcome
	system.LearnFromOutcome(entity.ID, true, 1.0, 1)

	// Check that learning occurred
	if network.Experience <= initialExperience {
		t.Error("Experience should increase after positive reward")
	}

	if network.CorrectDecisions != initialCorrectDecisions+1 {
		t.Errorf("Expected %d correct decisions, got %d", initialCorrectDecisions+1, network.CorrectDecisions)
	}

	if system.TotalLearningEvents == 0 {
		t.Error("Total learning events should be tracked")
	}

	// Learn from failure
	_ = network.Experience // Note: variable not used in later logic
	system.LearnFromOutcome(entity.ID, false, -0.5, 2)

	// Experience might decrease or stay same with negative reward
	if network.CorrectDecisions != initialCorrectDecisions+1 {
		t.Error("Correct decisions should not increase for failure")
	}
}

// TestActivationFunctions tests different activation functions
func TestActivationFunctions(t *testing.T) {
	system := NewNeuralAISystem()

	// Test sigmoid
	result := system.activationFunction(0.0, "sigmoid")
	if result != 0.5 {
		t.Errorf("Sigmoid(0) should be 0.5, got %f", result)
	}

	// Test tanh
	result = system.activationFunction(0.0, "tanh")
	if result != 0.0 {
		t.Errorf("Tanh(0) should be 0.0, got %f", result)
	}

	// Test relu
	result = system.activationFunction(-1.0, "relu")
	if result != 0.0 {
		t.Errorf("ReLU(-1) should be 0.0, got %f", result)
	}

	result = system.activationFunction(1.0, "relu")
	if result != 1.0 {
		t.Errorf("ReLU(1) should be 1.0, got %f", result)
	}

	// Test linear
	result = system.activationFunction(2.5, "linear")
	if result != 2.5 {
		t.Errorf("Linear(2.5) should be 2.5, got %f", result)
	}

	// Test default (should fallback to tanh)
	result = system.activationFunction(0.0, "unknown")
	if result != 0.0 {
		t.Errorf("Unknown activation should default to tanh, got %f", result)
	}
}

// TestNeuralSystemUpdate tests the system update mechanism
func TestNeuralSystemUpdate(t *testing.T) {
	system := NewNeuralAISystem()

	// Create entities with varying intelligence
	entity1 := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity1.SetTrait("intelligence", 0.8) // High intelligence - should get network
	entity1.SetTrait("curiosity", 0.6)
	entity1.IsAlive = true

	entity2 := NewEntity(2, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 1, Y: 1})
	entity2.SetTrait("intelligence", 0.2) // Low intelligence - should not get network
	entity2.SetTrait("curiosity", 0.3)
	entity2.IsAlive = true

	entity3 := NewEntity(3, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 2, Y: 2})
	entity3.SetTrait("intelligence", 0.5) // Medium intelligence - should get network
	entity3.SetTrait("curiosity", 0.4)
	entity3.IsAlive = true

	entities := []*Entity{entity1, entity2, entity3}

	// Update system
	system.Update(entities, 0)

	// Check that networks were created for intelligent entities
	if system.EntityNetworks[entity1.ID] == nil {
		t.Error("High intelligence entity should have neural network")
	}

	if system.EntityNetworks[entity2.ID] != nil {
		t.Error("Low intelligence entity should not have neural network")
	}

	if system.EntityNetworks[entity3.ID] == nil {
		t.Error("Medium intelligence entity should have neural network")
	}

	if system.TotalNetworks != 2 {
		t.Errorf("Expected 2 total networks, got %d", system.TotalNetworks)
	}
}

// TestNeuralStatsGeneration tests statistics generation
func TestNeuralStatsGeneration(t *testing.T) {
	system := NewNeuralAISystem()

	// Create test entity and network
	entity := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("curiosity", 0.6)
	entity.IsAlive = true

	system.CreateNeuralNetwork(entity, 0)

	// Make some decisions and learn
	environmentInputs := []float64{0.5, 0.8, 0.1, 0.6, 0.3}
	system.ProcessNeuralDecision(entity, environmentInputs, 0)
	system.LearnFromOutcome(entity.ID, true, 1.0, 1)

	// Get system statistics
	stats := system.GetNeuralStats()

	// Check that all expected stats are present
	expectedStats := []string{
		"total_networks", "total_behaviors", "total_learning_events",
		"avg_network_complexity", "emergent_behaviors", "base_learning_rate",
		"adaptation_rate", "success_rate", "total_experience", "avg_experience_per_network",
	}

	for _, stat := range expectedStats {
		if stats[stat] == nil {
			t.Errorf("Missing stat: %s", stat)
		}
	}

	// Check specific values
	if stats["total_networks"].(int) != 1 {
		t.Errorf("Expected 1 total network, got %d", stats["total_networks"].(int))
	}

	if stats["total_learning_events"].(int) != 1 {
		t.Errorf("Expected 1 learning event, got %d", stats["total_learning_events"].(int))
	}

	if stats["success_rate"].(float64) != 1.0 {
		t.Errorf("Expected 100%% success rate, got %f", stats["success_rate"].(float64))
	}
}

// TestEntityNeuralData tests entity-specific neural data retrieval
func TestEntityNeuralData(t *testing.T) {
	system := NewNeuralAISystem()

	// Create test entity
	entity := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("curiosity", 0.6)
	entity.IsAlive = true

	// Test data for non-existent network
	data := system.GetEntityNeuralData(entity.ID)
	if data != nil {
		t.Error("Should return nil for non-existent network")
	}

	// Create network and test data
	system.CreateNeuralNetwork(entity, 0)
	data = system.GetEntityNeuralData(entity.ID)

	if data == nil {
		t.Error("Should return data for existing network")
	}

	// Check expected fields
	expectedFields := []string{
		"network_id", "type", "architecture", "experience", "adaptability",
		"learning_rate", "total_decisions", "correct_decisions", "success_rate",
		"complexity_score", "neuron_count", "input_count", "output_count", "hidden_layers",
	}

	for _, field := range expectedFields {
		if data[field] == nil {
			t.Errorf("Missing field: %s", field)
		}
	}

	// Check specific values
	if data["type"].(NeuralNetworkType) != FeedForward {
		t.Error("Expected FeedForward network type")
	}

	if data["input_count"].(int) != 5 {
		t.Errorf("Expected 5 input neurons, got %d", data["input_count"].(int))
	}

	if data["output_count"].(int) != 3 {
		t.Errorf("Expected 3 output neurons, got %d", data["output_count"].(int))
	}
}

// TestNeuralNetworkIntegration tests that neural networks actually affect entity behavior
func TestNeuralNetworkIntegration(t *testing.T) {
	// Create a minimal world for testing neural integration
	config := WorldConfig{
		Width:          100,
		Height:         100,
		NumPopulations: 1,
		PopulationSize: 5,
		GridWidth:      10,
		GridHeight:     10,
	}

	world := NewWorld(config)

	// Add a population with high intelligence entities
	popConfig := PopulationConfig{
		Name:    "test",
		Species: "testspecies",
		BaseTraits: map[string]float64{
			"intelligence": 0.8,
			"vision":       0.7,
			"speed":        0.6,
			"cooperation":  0.5,
			"aggression":   0.3,
		},
		StartPos:         Position{X: 50, Y: 50},
		Spread:           10,
		BaseMutationRate: 0.01,
	}

	world.AddPopulation(popConfig)

	// Ensure entities are alive and have high intelligence
	intelligentCount := 0
	for _, entity := range world.AllEntities {
		if entity.IsAlive && entity.GetTrait("intelligence") > 0.3 {
			intelligentCount++
		}
	}

	if intelligentCount == 0 {
		t.Error("Should have intelligent entities for neural network testing")
	}

	// Run world update to create neural networks
	world.Update()

	// Check that neural networks were created for intelligent entities
	networkCount := len(world.NeuralAISystem.EntityNetworks)
	if networkCount == 0 {
		t.Error("Neural networks should be created for intelligent entities")
	}

	// Record initial positions
	initialPositions := make(map[int]Position)
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			initialPositions[entity.ID] = entity.Position
		}
	}

	// Run several updates to let neural networks make decisions
	for i := 0; i < 10; i++ {
		world.Update()
	}

	// Check that entities with neural networks moved (neural decision making worked)
	entitiesMoved := 0
	neurallControlledEntities := 0

	for _, entity := range world.AllEntities {
		if entity.IsAlive && world.NeuralAISystem.EntityNetworks[entity.ID] != nil {
			neurallControlledEntities++
			initialPos := initialPositions[entity.ID]
			currentPos := entity.Position

			// Check if entity moved significantly (more than 1 unit)
			distance := math.Sqrt(math.Pow(currentPos.X-initialPos.X, 2) + math.Pow(currentPos.Y-initialPos.Y, 2))
			if distance > 1.0 {
				entitiesMoved++
			}
		}
	}

	if neurallControlledEntities == 0 {
		t.Error("Should have entities with neural networks")
	}

	// At least some entities should have moved due to neural decisions
	// (Note: might not be 100% due to energy constraints or neural decisions to stay put)
	if entitiesMoved == 0 {
		t.Error("Entities with neural networks should have moved based on neural decisions")
	}

	// Check that neural networks are learning
	totalLearningEvents := world.NeuralAISystem.TotalLearningEvents
	if totalLearningEvents == 0 {
		t.Error("Neural networks should be learning from their decisions")
	}

	// Check that neural networks have made decisions
	totalDecisions := 0
	for _, network := range world.NeuralAISystem.EntityNetworks {
		totalDecisions += network.TotalDecisions
	}

	if totalDecisions == 0 {
		t.Error("Neural networks should be making decisions")
	}

	t.Logf("Neural Integration Test Results:")
	t.Logf("  Entities with neural networks: %d", neurallControlledEntities)
	t.Logf("  Entities that moved: %d", entitiesMoved)
	t.Logf("  Total learning events: %d", totalLearningEvents)
	t.Logf("  Total neural decisions: %d", totalDecisions)
}

// TestNeuralNetworkCleanup tests cleanup of networks for dead entities
func TestNeuralNetworkCleanup(t *testing.T) {
	system := NewNeuralAISystem()

	// Create entities
	entity1 := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	entity1.SetTrait("intelligence", 0.8)
	entity1.SetTrait("curiosity", 0.6)
	entity1.IsAlive = true

	entity2 := NewEntity(2, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 1, Y: 1})
	entity2.SetTrait("intelligence", 0.7)
	entity2.SetTrait("curiosity", 0.5)
	entity2.IsAlive = true

	// Create networks
	system.CreateNeuralNetwork(entity1, 0)
	system.CreateNeuralNetwork(entity2, 0)

	if len(system.EntityNetworks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(system.EntityNetworks))
	}

	// Kill entity2
	entity2.IsAlive = false

	// Update system with only living entities
	entities := []*Entity{entity1} // entity2 not included (dead)
	system.Update(entities, 1)

	// Check that dead entity's network was cleaned up
	if len(system.EntityNetworks) != 1 {
		t.Errorf("Expected 1 network after cleanup, got %d", len(system.EntityNetworks))
	}

	if system.EntityNetworks[entity1.ID] == nil {
		t.Error("Living entity's network should still exist")
	}

	if system.EntityNetworks[entity2.ID] != nil {
		t.Error("Dead entity's network should be cleaned up")
	}
}
