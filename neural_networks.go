package main

import (
	"math"
	"math/rand"
)

// NeuralNetworkType represents different types of neural network architectures
type NeuralNetworkType int

const (
	FeedForward NeuralNetworkType = iota
	Recurrent
	Convolutional
	Reinforcement
)

// Neuron represents a single neuron in the network
type Neuron struct {
	ID          int       `json:"id"`
	Value       float64   `json:"value"`        // Current activation value
	Bias        float64   `json:"bias"`         // Neuron bias
	Layer       int       `json:"layer"`        // Which layer this neuron belongs to
	Activation  string    `json:"activation"`   // Activation function type
	Connections []*Synapse `json:"connections"` // Outgoing connections
}

// Synapse represents a connection between neurons
type Synapse struct {
	FromNeuronID int     `json:"from_neuron_id"`
	ToNeuronID   int     `json:"to_neuron_id"`
	Weight       float64 `json:"weight"`
	Strength     float64 `json:"strength"`     // Connection strength (can change over time)
	LastActive   int     `json:"last_active"`  // Last tick this synapse was used
}

// EntityNeuralNetwork represents a complete neural network for an entity
type EntityNeuralNetwork struct {
	ID               int                     `json:"id"`
	EntityID         int                     `json:"entity_id"`
	Type             NeuralNetworkType       `json:"type"`
	Neurons          map[int]*Neuron         `json:"neurons"`
	InputNeurons     []int                   `json:"input_neurons"`   // IDs of input neurons
	OutputNeurons    []int                   `json:"output_neurons"`  // IDs of output neurons
	HiddenLayers     [][]int                 `json:"hidden_layers"`   // Hidden layer neuron IDs
	LearningRate     float64                 `json:"learning_rate"`
	Architecture     string                  `json:"architecture"`    // Description of network structure
	Experience       float64                 `json:"experience"`      // Total learning experience
	Adaptability     float64                 `json:"adaptability"`    // How quickly the network adapts
	CreatedTick      int                     `json:"created_tick"`
	LastUpdateTick   int                     `json:"last_update_tick"`
	
	// Learning and memory
	RecentInputs     [][]float64             `json:"recent_inputs"`   // Recent input patterns
	RecentOutputs    [][]float64             `json:"recent_outputs"`  // Recent output patterns
	SuccessfulActions map[string]float64     `json:"successful_actions"` // Action -> success rate
	BehaviorPatterns  map[string]int         `json:"behavior_patterns"` // Pattern -> frequency
	
	// Performance metrics
	CorrectDecisions  int                    `json:"correct_decisions"`
	TotalDecisions    int                    `json:"total_decisions"`
	AvgResponseTime   float64               `json:"avg_response_time"`
	ComplexityScore   float64               `json:"complexity_score"`
}

// NeuralBehavior represents a learned behavior pattern
type NeuralBehavior struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	InputPattern    []float64 `json:"input_pattern"`    // Required input conditions
	OutputPattern   []float64 `json:"output_pattern"`   // Expected output actions
	SuccessRate     float64   `json:"success_rate"`     // How often this behavior succeeds
	UsageCount      int       `json:"usage_count"`      // How many times this has been used
	LastUsed        int       `json:"last_used"`        // Last tick this behavior was used
	Complexity      float64   `json:"complexity"`       // How complex this behavior is
	EnergyEfficiency float64  `json:"energy_efficiency"` // Energy cost vs benefit
	Context         string    `json:"context"`          // When this behavior is most effective
}

// NeuralAISystem manages neural networks and AI behaviors for all entities
type NeuralAISystem struct {
	EntityNetworks      map[int]*EntityNeuralNetwork    `json:"entity_networks"`      // Entity ID -> Neural Network
	LearnedBehaviors    map[int]*NeuralBehavior   `json:"learned_behaviors"`    // Behavior ID -> Behavior
	NextNeuronID        int                       `json:"next_neuron_id"`
	NextNetworkID       int                       `json:"next_network_id"`
	NextBehaviorID      int                       `json:"next_behavior_id"`
	
	// System-wide learning
	CollectiveBehaviors map[string]*NeuralBehavior `json:"collective_behaviors"` // Shared learned behaviors
	SuccessfulStrategies []string                   `json:"successful_strategies"` // Most effective behavior patterns
	
	// System parameters
	BaseLearningRate    float64                   `json:"base_learning_rate"`
	NetworkComplexity   int                       `json:"network_complexity"`    // Default network size
	AdaptationRate      float64                   `json:"adaptation_rate"`       // How quickly networks adapt
	ExperienceDecay     float64                   `json:"experience_decay"`      // How experience fades over time
	
	// Statistics
	TotalNetworks       int                       `json:"total_networks"`
	TotalBehaviors      int                       `json:"total_behaviors"`
	TotalLearningEvents int                       `json:"total_learning_events"`
	AvgNetworkComplexity float64                  `json:"avg_network_complexity"`
	EmergentBehaviors   int                       `json:"emergent_behaviors"`    // Unprogrammed behaviors discovered
}

// NewNeuralAISystem creates a new neural AI system
func NewNeuralAISystem() *NeuralAISystem {
	return &NeuralAISystem{
		EntityNetworks:       make(map[int]*EntityNeuralNetwork),
		LearnedBehaviors:     make(map[int]*NeuralBehavior),
		CollectiveBehaviors:  make(map[string]*NeuralBehavior),
		SuccessfulStrategies: make([]string, 0),
		NextNeuronID:         1,
		NextNetworkID:        1,
		NextBehaviorID:       1,
		BaseLearningRate:     0.01,
		NetworkComplexity:    10, // Default 10 neurons per network
		AdaptationRate:       0.05,
		ExperienceDecay:      0.001,
	}
}

// CreateNeuralNetwork creates a neural network for an entity
func (nai *NeuralAISystem) CreateNeuralNetwork(entity *Entity, tick int) *EntityNeuralNetwork {
	if nai.EntityNetworks[entity.ID] != nil {
		return nai.EntityNetworks[entity.ID] // Already has a network
	}
	
	// Determine network complexity based on entity intelligence
	intelligence := entity.GetTrait("intelligence")
	complexity := int(float64(nai.NetworkComplexity) * (0.5 + intelligence*0.5)) // 5-15 neurons
	
	network := &EntityNeuralNetwork{
		ID:                nai.NextNetworkID,
		EntityID:          entity.ID,
		Type:              FeedForward, // Start with simple feedforward
		Neurons:           make(map[int]*Neuron),
		InputNeurons:      make([]int, 0),
		OutputNeurons:     make([]int, 0),
		HiddenLayers:      make([][]int, 0),
		LearningRate:      nai.BaseLearningRate * (0.5 + intelligence*0.5),
		Architecture:      "Feedforward",
		Experience:        0.0,
		Adaptability:      entity.GetTrait("curiosity"),
		CreatedTick:       tick,
		LastUpdateTick:    tick,
		RecentInputs:      make([][]float64, 0),
		RecentOutputs:     make([][]float64, 0),
		SuccessfulActions: make(map[string]float64),
		BehaviorPatterns:  make(map[string]int),
		CorrectDecisions:  0,
		TotalDecisions:    0,
		AvgResponseTime:   0.0,
		ComplexityScore:   float64(complexity),
	}
	
	// Create network architecture
	nai.buildNetworkArchitecture(network, complexity)
	
	nai.EntityNetworks[entity.ID] = network
	nai.NextNetworkID++
	nai.TotalNetworks++
	
	return network
}

// buildNetworkArchitecture creates the neural network structure
func (nai *NeuralAISystem) buildNetworkArchitecture(network *EntityNeuralNetwork, complexity int) {
	// Input layer (5 neurons for basic sensory input)
	inputSize := 5
	for i := 0; i < inputSize; i++ {
		neuron := &Neuron{
			ID:          nai.NextNeuronID,
			Value:       0.0,
			Bias:        rand.Float64()*0.2 - 0.1, // Small random bias
			Layer:       0,
			Activation:  "relu",
			Connections: make([]*Synapse, 0),
		}
		network.Neurons[neuron.ID] = neuron
		network.InputNeurons = append(network.InputNeurons, neuron.ID)
		nai.NextNeuronID++
	}
	
	// Hidden layer(s)
	hiddenSize := complexity - inputSize - 3 // Leave room for output layer
	if hiddenSize > 0 {
		hiddenLayer := make([]int, 0)
		for i := 0; i < hiddenSize; i++ {
			neuron := &Neuron{
				ID:          nai.NextNeuronID,
				Value:       0.0,
				Bias:        rand.Float64()*0.2 - 0.1,
				Layer:       1,
				Activation:  "sigmoid",
				Connections: make([]*Synapse, 0),
			}
			network.Neurons[neuron.ID] = neuron
			hiddenLayer = append(hiddenLayer, neuron.ID)
			nai.NextNeuronID++
		}
		network.HiddenLayers = append(network.HiddenLayers, hiddenLayer)
	}
	
	// Output layer (3 neurons for movement decisions: move_x, move_y, action)
	outputSize := 3
	for i := 0; i < outputSize; i++ {
		neuron := &Neuron{
			ID:          nai.NextNeuronID,
			Value:       0.0,
			Bias:        rand.Float64()*0.2 - 0.1,
			Layer:       2,
			Activation:  "tanh",
			Connections: make([]*Synapse, 0),
		}
		network.Neurons[neuron.ID] = neuron
		network.OutputNeurons = append(network.OutputNeurons, neuron.ID)
		nai.NextNeuronID++
	}
	
	// Create connections between layers
	nai.connectLayers(network)
}

// connectLayers creates synaptic connections between neural network layers
func (nai *NeuralAISystem) connectLayers(network *EntityNeuralNetwork) {
	// Connect input to hidden layer (or output if no hidden)
	if len(network.HiddenLayers) > 0 {
		// Input to first hidden layer
		for _, inputID := range network.InputNeurons {
			for _, hiddenID := range network.HiddenLayers[0] {
				synapse := &Synapse{
					FromNeuronID: inputID,
					ToNeuronID:   hiddenID,
					Weight:       rand.Float64()*2 - 1, // Random weight between -1 and 1
					Strength:     1.0,
					LastActive:   0,
				}
				network.Neurons[inputID].Connections = append(network.Neurons[inputID].Connections, synapse)
			}
		}
		
		// Hidden layer to output
		for _, hiddenID := range network.HiddenLayers[0] {
			for _, outputID := range network.OutputNeurons {
				synapse := &Synapse{
					FromNeuronID: hiddenID,
					ToNeuronID:   outputID,
					Weight:       rand.Float64()*2 - 1,
					Strength:     1.0,
					LastActive:   0,
				}
				network.Neurons[hiddenID].Connections = append(network.Neurons[hiddenID].Connections, synapse)
			}
		}
	} else {
		// Direct input to output connections
		for _, inputID := range network.InputNeurons {
			for _, outputID := range network.OutputNeurons {
				synapse := &Synapse{
					FromNeuronID: inputID,
					ToNeuronID:   outputID,
					Weight:       rand.Float64()*2 - 1,
					Strength:     1.0,
					LastActive:   0,
				}
				network.Neurons[inputID].Connections = append(network.Neurons[inputID].Connections, synapse)
			}
		}
	}
}

// ProcessNeuralDecision uses the neural network to make decisions for an entity
func (nai *NeuralAISystem) ProcessNeuralDecision(entity *Entity, environmentInputs []float64, tick int) []float64 {
	network := nai.EntityNetworks[entity.ID]
	if network == nil {
		// Create network if it doesn't exist
		network = nai.CreateNeuralNetwork(entity, tick)
	}
	
	// Feed inputs through the network
	outputs := nai.forwardPass(network, environmentInputs, tick)
	
	// Record this decision for learning
	network.TotalDecisions++
	network.LastUpdateTick = tick
	
	// Store recent inputs/outputs for learning
	network.RecentInputs = append(network.RecentInputs, environmentInputs)
	network.RecentOutputs = append(network.RecentOutputs, outputs)
	
	// Keep only recent history (last 10 decisions)
	if len(network.RecentInputs) > 10 {
		network.RecentInputs = network.RecentInputs[1:]
		network.RecentOutputs = network.RecentOutputs[1:]
	}
	
	return outputs
}

// forwardPass performs a forward pass through the neural network
func (nai *NeuralAISystem) forwardPass(network *EntityNeuralNetwork, inputs []float64, tick int) []float64 {
	// Reset all neuron values
	for _, neuron := range network.Neurons {
		neuron.Value = 0.0
	}
	
	// Set input values
	for i, inputID := range network.InputNeurons {
		if i < len(inputs) {
			network.Neurons[inputID].Value = inputs[i]
		}
	}
	
	// Process hidden layers
	for _, layer := range network.HiddenLayers {
		for _, neuronID := range layer {
			neuron := network.Neurons[neuronID]
			// Sum weighted inputs from connected neurons
			sum := neuron.Bias
			for _, otherNeuron := range network.Neurons {
				for _, synapse := range otherNeuron.Connections {
					if synapse.ToNeuronID == neuronID {
						sum += otherNeuron.Value * synapse.Weight * synapse.Strength
						synapse.LastActive = tick
					}
				}
			}
			neuron.Value = nai.activationFunction(sum, neuron.Activation)
		}
	}
	
	// Process output layer
	outputs := make([]float64, len(network.OutputNeurons))
	for i, outputID := range network.OutputNeurons {
		neuron := network.Neurons[outputID]
		sum := neuron.Bias
		for _, otherNeuron := range network.Neurons {
			for _, synapse := range otherNeuron.Connections {
				if synapse.ToNeuronID == outputID {
					sum += otherNeuron.Value * synapse.Weight * synapse.Strength
					synapse.LastActive = tick
				}
			}
		}
		neuron.Value = nai.activationFunction(sum, neuron.Activation)
		outputs[i] = neuron.Value
	}
	
	return outputs
}

// activationFunction applies the specified activation function
func (nai *NeuralAISystem) activationFunction(x float64, funcType string) float64 {
	switch funcType {
	case "sigmoid":
		return 1.0 / (1.0 + math.Exp(-x))
	case "tanh":
		return math.Tanh(x)
	case "relu":
		return math.Max(0, x)
	case "linear":
		return x
	default:
		return math.Tanh(x) // Default to tanh
	}
}

// LearnFromOutcome updates the neural network based on the success/failure of a decision
func (nai *NeuralAISystem) LearnFromOutcome(entityID int, success bool, reward float64, tick int) {
	network := nai.EntityNetworks[entityID]
	if network == nil {
		return
	}
	
	if success {
		network.CorrectDecisions++
	}
	
	// Simple reinforcement learning: adjust weights based on reward
	learningFactor := network.LearningRate * reward
	if !success {
		learningFactor *= -0.5 // Negative learning for failures
	}
	
	// Update connection weights for recently active synapses
	for _, neuron := range network.Neurons {
		for _, synapse := range neuron.Connections {
			if tick-synapse.LastActive <= 5 { // Only recent connections
				synapse.Weight += learningFactor * rand.Float64() * 0.1
				// Keep weights in reasonable range
				synapse.Weight = math.Max(-2.0, math.Min(2.0, synapse.Weight))
				
				// Update connection strength based on usage
				if success {
					synapse.Strength = math.Min(2.0, synapse.Strength+0.01)
				} else {
					synapse.Strength = math.Max(0.1, synapse.Strength-0.005)
				}
			}
		}
	}
	
	// Increase experience
	network.Experience += reward
	nai.TotalLearningEvents++
}

// Update processes neural AI system updates
func (nai *NeuralAISystem) Update(entities []*Entity, tick int) {
	// Create networks for entities that don't have them (if they have sufficient intelligence)
	for _, entity := range entities {
		if entity.IsAlive && entity.GetTrait("intelligence") > 0.3 {
			if nai.EntityNetworks[entity.ID] == nil {
				nai.CreateNeuralNetwork(entity, tick)
			}
		}
	}
	
	// Decay experience and clean up old data
	if tick%100 == 0 {
		nai.decayExperience()
	}
	
	// Update statistics
	nai.updateStatistics()
	
	// Clean up networks for dead entities
	nai.cleanupDeadEntities(entities)
}

// decayExperience reduces the experience of all networks over time
func (nai *NeuralAISystem) decayExperience() {
	for _, network := range nai.EntityNetworks {
		network.Experience *= (1.0 - nai.ExperienceDecay)
		
		// Decay connection strengths slightly
		for _, neuron := range network.Neurons {
			for _, synapse := range neuron.Connections {
				synapse.Strength *= 0.999 // Very slight decay
			}
		}
	}
}

// updateStatistics updates system-wide statistics
func (nai *NeuralAISystem) updateStatistics() {
	nai.TotalNetworks = len(nai.EntityNetworks)
	nai.TotalBehaviors = len(nai.LearnedBehaviors)
	
	// Calculate average network complexity
	totalComplexity := 0.0
	for _, network := range nai.EntityNetworks {
		totalComplexity += network.ComplexityScore
	}
	if nai.TotalNetworks > 0 {
		nai.AvgNetworkComplexity = totalComplexity / float64(nai.TotalNetworks)
	}
}

// cleanupDeadEntities removes neural networks for dead entities
func (nai *NeuralAISystem) cleanupDeadEntities(entities []*Entity) {
	aliveEntities := make(map[int]bool)
	for _, entity := range entities {
		if entity.IsAlive {
			aliveEntities[entity.ID] = true
		}
	}
	
	for entityID := range nai.EntityNetworks {
		if !aliveEntities[entityID] {
			delete(nai.EntityNetworks, entityID)
		}
	}
}

// GetNeuralStats returns statistics about the neural AI system
func (nai *NeuralAISystem) GetNeuralStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["total_networks"] = nai.TotalNetworks
	stats["total_behaviors"] = nai.TotalBehaviors
	stats["total_learning_events"] = nai.TotalLearningEvents
	stats["avg_network_complexity"] = nai.AvgNetworkComplexity
	stats["emergent_behaviors"] = nai.EmergentBehaviors
	stats["base_learning_rate"] = nai.BaseLearningRate
	stats["adaptation_rate"] = nai.AdaptationRate
	
	// Calculate collective intelligence metrics
	totalDecisions := 0
	correctDecisions := 0
	totalExperience := 0.0
	
	for _, network := range nai.EntityNetworks {
		totalDecisions += network.TotalDecisions
		correctDecisions += network.CorrectDecisions
		totalExperience += network.Experience
	}
	
	if totalDecisions > 0 {
		stats["success_rate"] = float64(correctDecisions) / float64(totalDecisions)
	} else {
		stats["success_rate"] = 0.0
	}
	
	stats["total_experience"] = totalExperience
	stats["avg_experience_per_network"] = 0.0
	if nai.TotalNetworks > 0 {
		stats["avg_experience_per_network"] = totalExperience / float64(nai.TotalNetworks)
	}
	
	return stats
}

// GetEntityNeuralData returns neural network data for a specific entity
func (nai *NeuralAISystem) GetEntityNeuralData(entityID int) map[string]interface{} {
	network := nai.EntityNetworks[entityID]
	if network == nil {
		return nil
	}
	
	data := make(map[string]interface{})
	data["network_id"] = network.ID
	data["type"] = network.Type
	data["architecture"] = network.Architecture
	data["experience"] = network.Experience
	data["adaptability"] = network.Adaptability
	data["learning_rate"] = network.LearningRate
	data["total_decisions"] = network.TotalDecisions
	data["correct_decisions"] = network.CorrectDecisions
	data["success_rate"] = 0.0
	
	if network.TotalDecisions > 0 {
		data["success_rate"] = float64(network.CorrectDecisions) / float64(network.TotalDecisions)
	}
	
	data["complexity_score"] = network.ComplexityScore
	data["neuron_count"] = len(network.Neurons)
	data["input_count"] = len(network.InputNeurons)
	data["output_count"] = len(network.OutputNeurons)
	data["hidden_layers"] = len(network.HiddenLayers)
	
	return data
}