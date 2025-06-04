package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// NetworkConnectionType represents different types of underground connections
type NetworkConnectionType int

const (
	ConnectionMycorrhizal NetworkConnectionType = iota // Fungal networks
	ConnectionRoot                                     // Direct root connections
	ConnectionChemical                                 // Chemical signaling paths
)

// NetworkConnection represents a connection between two plants
type NetworkConnection struct {
	ID           int                   `json:"id"`
	PlantA       *Plant                `json:"plant_a"`
	PlantB       *Plant                `json:"plant_b"`
	Type         NetworkConnectionType `json:"type"`
	Strength     float64               `json:"strength"`      // 0.0 to 1.0
	Health       float64               `json:"health"`        // 0.0 to 1.0
	Age          int                   `json:"age"`           // Ticks since formation
	Distance     float64               `json:"distance"`      // Physical distance
	LastTransfer int                   `json:"last_transfer"` // Last tick when resources were transferred
	Efficiency   float64               `json:"efficiency"`    // How well the connection transfers resources
}

// ChemicalSignal represents chemical communication through networks
type ChemicalSignal struct {
	ID        int                `json:"id"`
	Source    *Plant             `json:"source"`
	Type      ChemicalSignalType `json:"type"`
	Intensity float64            `json:"intensity"`
	Age       int                `json:"age"`
	MaxAge    int                `json:"max_age"`
	Visited   map[int]bool       `json:"visited"` // Plants that have received this signal
	Message   string             `json:"message"`
	Metadata  map[string]float64 `json:"metadata"`
}

// ChemicalSignalType represents different types of chemical signals
type ChemicalSignalType int

const (
	SignalNutrientAvailable ChemicalSignalType = iota // Excess nutrients available
	SignalNutrientNeeded                              // Need nutrients
	SignalThreatDetected                              // Herbivore or other threat
	SignalOptimalGrowth                               // Good growing conditions
	SignalReproductionReady                           // Ready to reproduce
	SignalToxicConditions                             // Environmental toxins detected
)

// getSignalTypeName returns the string name for a chemical signal type
func getSignalTypeName(signalType ChemicalSignalType) string {
	switch signalType {
	case SignalNutrientAvailable:
		return "nutrient_available"
	case SignalNutrientNeeded:
		return "nutrient_needed"
	case SignalThreatDetected:
		return "threat_detected"
	case SignalOptimalGrowth:
		return "optimal_growth"
	case SignalReproductionReady:
		return "reproduction_ready"
	case SignalToxicConditions:
		return "toxic_conditions"
	default:
		return "unknown"
	}
}

// PlantNetworkSystem manages all underground plant networks
type PlantNetworkSystem struct {
	Connections      []*NetworkConnection `json:"connections"`
	ChemicalSignals  []*ChemicalSignal    `json:"chemical_signals"`
	NetworkClusters  []*NetworkCluster    `json:"network_clusters"`
	NextConnectionID int                  `json:"next_connection_id"`
	NextSignalID     int                  `json:"next_signal_id"`
	NextClusterID    int                  `json:"next_cluster_id"`

	// Configuration
	MaxConnectionDistance   float64 `json:"max_connection_distance"`
	ConnectionFormationRate float64 `json:"connection_formation_rate"`
	ResourceTransferRate    float64 `json:"resource_transfer_rate"`
	SignalDecayRate         float64 `json:"signal_decay_rate"`
	NetworkMaintenanceCost  float64 `json:"network_maintenance_cost"`

	// Statistics
	TotalResourceTransferred float64 `json:"total_resource_transferred"`
	ActiveConnections        int     `json:"active_connections"`
	ActiveSignals            int     `json:"active_signals"`
	NetworkEfficiency        float64 `json:"network_efficiency"`

	// Event tracking
	eventBus *CentralEventBus `json:"-"`
}

// NetworkCluster represents a group of connected plants
type NetworkCluster struct {
	ID                int                  `json:"id"`
	Plants            []*Plant             `json:"plants"`
	Connections       []*NetworkConnection `json:"connections"`
	CentralHub        *Plant               `json:"central_hub"` // Most connected plant
	TotalResources    float64              `json:"total_resources"`
	AverageHealth     float64              `json:"average_health"`
	Size              int                  `json:"size"`
	Efficiency        float64              `json:"efficiency"`
	FormationTick     int                  `json:"formation_tick"`
	LastResourceShare int                  `json:"last_resource_share"`
}

// NewPlantNetworkSystem creates a new plant network system
func NewPlantNetworkSystem(eventBus *CentralEventBus) *PlantNetworkSystem {
	return &PlantNetworkSystem{
		Connections:              make([]*NetworkConnection, 0),
		ChemicalSignals:          make([]*ChemicalSignal, 0),
		NetworkClusters:          make([]*NetworkCluster, 0),
		NextConnectionID:         1,
		NextSignalID:             1,
		NextClusterID:            1,
		MaxConnectionDistance:    15.0, // Maximum distance for network connections
		ConnectionFormationRate:  0.02, // Chance per tick to form new connections
		ResourceTransferRate:     0.1,  // Fraction of resources transferred per tick
		SignalDecayRate:          0.05, // How quickly chemical signals decay
		NetworkMaintenanceCost:   0.5,  // Energy cost per connection per tick
		TotalResourceTransferred: 0.0,
		ActiveConnections:        0,
		ActiveSignals:            0,
		NetworkEfficiency:        0.0,
		eventBus:                 eventBus,
	}
}

// Update processes all network activities for one tick
func (pns *PlantNetworkSystem) Update(allPlants []*Plant, currentTick int) {
	// 1. Form new connections between nearby compatible plants
	pns.formNewConnections(allPlants, currentTick)

	// 2. Update existing connections (aging, health changes)
	pns.updateConnections(currentTick)

	// 3. Transfer resources through network
	pns.transferResources(currentTick)

	// 4. Propagate chemical signals
	pns.propagateChemicalSignals(currentTick)

	// 5. Update network clusters
	pns.updateNetworkClusters(currentTick)

	// 6. Maintain network health (some connections may weaken or break)
	pns.maintainNetworkHealth()

	// 7. Clean up dead connections and expired signals
	pns.cleanup()

	// 8. Update statistics
	pns.updateStatistics()
}

// formNewConnections attempts to create new connections between compatible plants
func (pns *PlantNetworkSystem) formNewConnections(allPlants []*Plant, currentTick int) {
	for i, plantA := range allPlants {
		if !plantA.IsAlive || plantA.Energy < 20 { // Need energy to form connections
			continue
		}

		// Limit number of attempts per plant per tick
		attempts := 0
		maxAttempts := 3

		for j := i + 1; j < len(allPlants) && attempts < maxAttempts; j++ {
			plantB := allPlants[j]
			if !plantB.IsAlive || plantB.Energy < 20 {
				continue
			}

			attempts++

			// Check if connection already exists
			if pns.hasConnection(plantA, plantB) {
				continue
			}

			// Calculate distance
			distance := math.Sqrt(math.Pow(plantA.Position.X-plantB.Position.X, 2) +
				math.Pow(plantA.Position.Y-plantB.Position.Y, 2))

			// Too far apart
			if distance > pns.MaxConnectionDistance {
				continue
			}

			// Check compatibility and formation probability
			compatibility := pns.calculateCompatibility(plantA, plantB)
			formationChance := pns.ConnectionFormationRate * compatibility * (1.0 - distance/pns.MaxConnectionDistance)

			if rand.Float64() < formationChance {
				connection := pns.createConnection(plantA, plantB, distance, currentTick)
				pns.Connections = append(pns.Connections, connection)

				// Both plants pay energy cost for forming connection
				energyCost := 2.0 + distance*0.1
				plantA.Energy -= energyCost
				plantB.Energy -= energyCost
			}
		}
	}
}

// calculateCompatibility determines how likely two plants are to form a network connection
func (pns *PlantNetworkSystem) calculateCompatibility(plantA, plantB *Plant) float64 {
	compatibility := 0.5 // Base compatibility

	// Same species have higher compatibility
	if plantA.Type == plantB.Type {
		compatibility += 0.3
	}

	// Similar size plants connect better
	sizeDiff := math.Abs(plantA.Size - plantB.Size)
	if sizeDiff < 1.0 {
		compatibility += 0.2
	}

	// Healthy plants form better connections
	healthFactorA := math.Min(1.0, plantA.Energy/50.0)
	healthFactorB := math.Min(1.0, plantB.Energy/50.0)
	compatibility += (healthFactorA + healthFactorB) * 0.1

	// Age factor - mature plants form more stable connections
	ageFactorA := math.Min(1.0, float64(plantA.Age)/100.0)
	ageFactorB := math.Min(1.0, float64(plantB.Age)/100.0)
	compatibility += (ageFactorA + ageFactorB) * 0.1

	return math.Min(1.0, compatibility)
}

// createConnection creates a new network connection between two plants
func (pns *PlantNetworkSystem) createConnection(plantA, plantB *Plant, distance float64, currentTick int) *NetworkConnection {
	// Determine connection type based on plant types and characteristics
	var connectionType NetworkConnectionType
	if plantA.Type == PlantMushroom || plantB.Type == PlantMushroom {
		connectionType = ConnectionMycorrhizal // Mushrooms create fungal networks
	} else if plantA.Type == plantB.Type {
		connectionType = ConnectionRoot // Same species can connect roots
	} else {
		connectionType = ConnectionChemical // Different species use chemical signaling
	}

	// Calculate initial connection properties
	initialStrength := 0.1 + rand.Float64()*0.3 // Start weak, can grow stronger
	initialHealth := 0.7 + rand.Float64()*0.3   // Start reasonably healthy
	efficiency := 0.3 + rand.Float64()*0.4      // Initial efficiency

	connection := &NetworkConnection{
		ID:           pns.NextConnectionID,
		PlantA:       plantA,
		PlantB:       plantB,
		Type:         connectionType,
		Strength:     initialStrength,
		Health:       initialHealth,
		Age:          0,
		Distance:     distance,
		LastTransfer: currentTick,
		Efficiency:   efficiency,
	}

	// Emit connection formation event
	if pns.eventBus != nil {
		metadata := map[string]interface{}{
			"connection_id":   connection.ID,
			"connection_type": connectionType,
			"distance":        distance,
			"initial_strength": initialStrength,
			"initial_health":  initialHealth,
			"efficiency":      efficiency,
			"plant_a_id":      plantA.ID,
			"plant_b_id":      plantB.ID,
			"plant_a_type":    plantA.Type,
			"plant_b_type":    plantB.Type,
		}
		
		pos := Position{
			X: (plantA.Position.X + plantB.Position.X) / 2,
			Y: (plantA.Position.Y + plantB.Position.Y) / 2,
		}
		
		pns.eventBus.EmitSystemEvent(
			currentTick,
			"connection_formed",
			"network",
			"plant_network_system",
			fmt.Sprintf("Network connection %d formed between plants %d and %d (distance: %.2f)", 
				connection.ID, plantA.ID, plantB.ID, distance),
			&pos,
			metadata,
		)
	}

	pns.NextConnectionID++
	return connection
}

// hasConnection checks if a connection already exists between two plants
func (pns *PlantNetworkSystem) hasConnection(plantA, plantB *Plant) bool {
	for _, conn := range pns.Connections {
		if (conn.PlantA.ID == plantA.ID && conn.PlantB.ID == plantB.ID) ||
			(conn.PlantA.ID == plantB.ID && conn.PlantB.ID == plantA.ID) {
			return true
		}
	}
	return false
}

// updateConnections updates all existing connections
func (pns *PlantNetworkSystem) updateConnections(currentTick int) {
	for _, conn := range pns.Connections {
		if !conn.PlantA.IsAlive || !conn.PlantB.IsAlive {
			conn.Health = 0 // Connection dies if either plant dies
			continue
		}

		conn.Age++

		// Connections strengthen with age and use
		if currentTick-conn.LastTransfer < 50 { // Recently used
			conn.Strength = math.Min(1.0, conn.Strength+0.01)
			conn.Efficiency = math.Min(1.0, conn.Efficiency+0.005)
		} else {
			// Unused connections weaken slowly
			conn.Strength = math.Max(0.0, conn.Strength-0.001)
			conn.Efficiency = math.Max(0.1, conn.Efficiency-0.002)
		}

		// Health changes based on plant health and environmental factors
		avgPlantHealth := (math.Min(1.0, conn.PlantA.Energy/50.0) + math.Min(1.0, conn.PlantB.Energy/50.0)) / 2.0
		healthChange := (avgPlantHealth - 0.5) * 0.02
		conn.Health = math.Max(0.0, math.Min(1.0, conn.Health+healthChange))

		// Maintenance cost
		maintenanceCost := pns.NetworkMaintenanceCost * conn.Strength
		conn.PlantA.Energy -= maintenanceCost / 2
		conn.PlantB.Energy -= maintenanceCost / 2
	}
}

// transferResources handles resource sharing through network connections
func (pns *PlantNetworkSystem) transferResources(currentTick int) {
	for _, conn := range pns.Connections {
		if conn.Health < 0.1 || conn.Strength < 0.1 {
			continue // Connection too weak for resource transfer
		}

		// Check if plants have significantly different energy levels
		energyDiff := conn.PlantA.Energy - conn.PlantB.Energy
		if math.Abs(energyDiff) < 5.0 {
			continue // Not enough difference to warrant transfer
		}

		// Determine donor and recipient
		var donor, recipient *Plant
		if energyDiff > 0 {
			donor = conn.PlantA
			recipient = conn.PlantB
		} else {
			donor = conn.PlantB
			recipient = conn.PlantA
			energyDiff = -energyDiff
		}

		// Calculate transfer amount
		maxTransfer := donor.Energy * pns.ResourceTransferRate * conn.Efficiency * conn.Health
		transferAmount := math.Min(maxTransfer, energyDiff*0.3) // Don't completely equalize

		// Only transfer if donor has enough energy to spare
		if donor.Energy > 25 && transferAmount > 0.5 {
			oldDonorEnergy := donor.Energy
			oldRecipientEnergy := recipient.Energy
			
			donor.Energy -= transferAmount
			recipient.Energy += transferAmount * 0.9 // 10% transfer loss

			conn.LastTransfer = currentTick
			pns.TotalResourceTransferred += transferAmount

			// Emit resource transfer event
			if pns.eventBus != nil {
				metadata := map[string]interface{}{
					"connection_id":         conn.ID,
					"donor_id":             donor.ID,
					"recipient_id":         recipient.ID,
					"transfer_amount":      transferAmount,
					"transfer_efficiency":  transferAmount * 0.9,
					"connection_efficiency": conn.Efficiency,
					"connection_health":    conn.Health,
					"donor_energy_before":  oldDonorEnergy,
					"donor_energy_after":   donor.Energy,
					"recipient_energy_before": oldRecipientEnergy,
					"recipient_energy_after":  recipient.Energy,
				}
				
				pos := Position{
					X: (donor.Position.X + recipient.Position.X) / 2,
					Y: (donor.Position.Y + recipient.Position.Y) / 2,
				}
				
				pns.eventBus.EmitSystemEvent(
					currentTick,
					"resource_transfer",
					"network",
					"plant_network_system",
					fmt.Sprintf("Resource transfer: %.2f energy from plant %d to plant %d via connection %d", 
						transferAmount, donor.ID, recipient.ID, conn.ID),
					&pos,
					metadata,
				)
			}

			// Send chemical signal about successful resource sharing
			if rand.Float64() < 0.3 { // 30% chance to signal
				pns.sendChemicalSignal(donor, SignalNutrientAvailable, 0.6, currentTick,
					"Shared resources through network")
			}
		}
	}
}

// sendChemicalSignal creates a new chemical signal in the network
func (pns *PlantNetworkSystem) sendChemicalSignal(source *Plant, signalType ChemicalSignalType,
	intensity float64, currentTick int, message string) {

	signal := &ChemicalSignal{
		ID:        pns.NextSignalID,
		Source:    source,
		Type:      signalType,
		Intensity: intensity,
		Age:       0,
		MaxAge:    30 + rand.Intn(20), // 30-50 ticks lifespan
		Visited:   make(map[int]bool),
		Message:   message,
		Metadata:  make(map[string]float64),
	}

	// Add metadata based on signal type
	switch signalType {
	case SignalNutrientAvailable:
		signal.Metadata["energy_level"] = source.Energy
		signal.Metadata["size"] = source.Size
	case SignalThreatDetected:
		signal.Metadata["threat_level"] = intensity
	case SignalReproductionReady:
		signal.Metadata["reproduction_rate"] = source.GetTrait("reproduction_rate")
	}

	pns.ChemicalSignals = append(pns.ChemicalSignals, signal)
	pns.NextSignalID++

	// Emit signal creation event
	if pns.eventBus != nil {
		metadata := map[string]interface{}{
			"signal_id":     signal.ID,
			"signal_type":   signalType,
			"intensity":     intensity,
			"source_id":     source.ID,
			"source_type":   source.Type,
			"source_energy": source.Energy,
			"message":       message,
			"max_age":       signal.MaxAge,
		}
		
		// Add signal-specific metadata
		for key, value := range signal.Metadata {
			metadata[key] = value
		}
		
		pns.eventBus.EmitSystemEvent(
			currentTick,
			"chemical_signal_created",
			"network",
			"plant_network_system",
			fmt.Sprintf("Chemical signal %d (%s) created by plant %d: %s", 
				signal.ID, getSignalTypeName(signalType), source.ID, message),
			&source.Position,
			metadata,
		)
	}
}

// propagateChemicalSignals spreads chemical signals through the network
func (pns *PlantNetworkSystem) propagateChemicalSignals(currentTick int) {
	for _, signal := range pns.ChemicalSignals {
		signal.Age++

		// Decay intensity over time
		signal.Intensity *= (1.0 - pns.SignalDecayRate)

		// Find plants connected to already visited plants
		plantsToVisit := make([]*Plant, 0)

		for _, conn := range pns.Connections {
			if conn.Health < 0.3 { // Connection too weak for signal propagation
				continue
			}

			// Check if signal can propagate through this connection
			var fromPlant, toPlant *Plant
			if signal.Visited[conn.PlantA.ID] && !signal.Visited[conn.PlantB.ID] {
				fromPlant = conn.PlantA
				toPlant = conn.PlantB
			} else if signal.Visited[conn.PlantB.ID] && !signal.Visited[conn.PlantA.ID] {
				fromPlant = conn.PlantB
				toPlant = conn.PlantA
			}

			if fromPlant != nil && toPlant != nil && toPlant.IsAlive {
				// Signal propagation probability based on connection quality
				propagationChance := conn.Health * conn.Strength * signal.Intensity
				if rand.Float64() < propagationChance {
					plantsToVisit = append(plantsToVisit, toPlant)
				}
			}
		}

		// Process signal effects on newly reached plants
		for _, plant := range plantsToVisit {
			if signal.Visited[plant.ID] {
				continue
			}

			signal.Visited[plant.ID] = true
			pns.processSignalEffect(plant, signal)

			// Emit signal propagation event
			if pns.eventBus != nil {
				metadata := map[string]interface{}{
					"signal_id":        signal.ID,
					"signal_type":      signal.Type,
					"source_id":        signal.Source.ID,
					"target_id":        plant.ID,
					"signal_intensity": signal.Intensity,
					"signal_age":       signal.Age,
				}
				
				pns.eventBus.EmitSystemEvent(
					currentTick,
					"chemical_signal_propagated",
					"network",
					"plant_network_system",
					fmt.Sprintf("Chemical signal %d propagated to plant %d (intensity: %.2f)", 
						signal.ID, plant.ID, signal.Intensity),
					&plant.Position,
					metadata,
				)
			}
		}

		// Mark source plant as visited on first propagation
		if signal.Age == 1 {
			signal.Visited[signal.Source.ID] = true
		}
	}
}

// processSignalEffect applies the effect of a chemical signal to a plant
func (pns *PlantNetworkSystem) processSignalEffect(plant *Plant, signal *ChemicalSignal) {
	effectiveness := signal.Intensity * math.Min(1.0, plant.Energy/30.0) // More effective on healthy plants

	switch signal.Type {
	case SignalNutrientAvailable:
		// Plant might move towards nutrient source or prepare for receiving resources
		if energyLevel, exists := signal.Metadata["energy_level"]; exists && energyLevel > plant.Energy*1.5 {
			// Boost growth rate temporarily
			plant.GrowthRate += effectiveness * 0.1
		}

	case SignalNutrientNeeded:
		// If plant has excess energy, it might be more willing to share
		if plant.Energy > 40 {
			// Increase resource sharing tendency (implemented through higher transfer rates)
			// This is handled in the transfer logic
		}

	case SignalThreatDetected:
		// Prepare defensive responses
		if rand.Float64() < effectiveness {
			// Temporarily increase toxicity as defense
			plant.Toxicity += effectiveness * 0.1
			// Reduce energy as plant prepares defenses
			plant.Energy -= 1.0
		}

	case SignalOptimalGrowth:
		// Boost growth if conditions are favorable
		if plant.Energy > 15 {
			plant.GrowthRate += effectiveness * 0.05
		}

	case SignalReproductionReady:
		// Might trigger reproductive behavior if plant is also ready
		reproductionThreshold := 30.0 + plant.GetTrait("reproduction_rate")*20
		if plant.Energy > reproductionThreshold && rand.Float64() < effectiveness*0.5 {
			// Increase likelihood of reproduction (this would be handled by the main reproduction system)
			// For now, just boost energy slightly as preparation
			plant.Energy += 2.0
		}

	case SignalToxicConditions:
		// Prepare for environmental stress
		if rand.Float64() < effectiveness {
			// Increase hardiness temporarily
			hardiness := plant.GetTrait("hardiness")
			plant.Traits["hardiness"] = Trait{Value: math.Min(1.0, hardiness+0.1)}
		}
	}
}

// updateNetworkClusters identifies and updates clusters of connected plants
func (pns *PlantNetworkSystem) updateNetworkClusters(currentTick int) {
	// Clear existing clusters
	pns.NetworkClusters = make([]*NetworkCluster, 0)

	// Find connected components using DFS
	visited := make(map[int]bool)
	plantConnections := pns.buildPlantConnectionMap()

	for plantID := range plantConnections {
		if visited[plantID] {
			continue
		}

		// Find all plants in this connected component
		clusterPlants := pns.findConnectedPlants(plantID, plantConnections, visited)
		if len(clusterPlants) >= 2 { // Only consider clusters with 2+ plants
			cluster := pns.createNetworkCluster(clusterPlants, currentTick)
			pns.NetworkClusters = append(pns.NetworkClusters, cluster)
		}
	}
}

// buildPlantConnectionMap creates a map of plant connections for cluster analysis
func (pns *PlantNetworkSystem) buildPlantConnectionMap() map[int][]int {
	plantConnections := make(map[int][]int)

	for _, conn := range pns.Connections {
		if conn.Health < 0.2 || !conn.PlantA.IsAlive || !conn.PlantB.IsAlive {
			continue // Skip weak or dead connections
		}

		idA := conn.PlantA.ID
		idB := conn.PlantB.ID

		if plantConnections[idA] == nil {
			plantConnections[idA] = make([]int, 0)
		}
		if plantConnections[idB] == nil {
			plantConnections[idB] = make([]int, 0)
		}

		plantConnections[idA] = append(plantConnections[idA], idB)
		plantConnections[idB] = append(plantConnections[idB], idA)
	}

	return plantConnections
}

// findConnectedPlants uses DFS to find all plants in a connected component
func (pns *PlantNetworkSystem) findConnectedPlants(startPlantID int, plantConnections map[int][]int, visited map[int]bool) []*Plant {
	var clusterPlants []*Plant
	stack := []int{startPlantID}

	for len(stack) > 0 {
		currentID := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[currentID] {
			continue
		}

		visited[currentID] = true

		// Find the plant with this ID
		for _, conn := range pns.Connections {
			var plant *Plant
			if conn.PlantA.ID == currentID && conn.PlantA.IsAlive {
				plant = conn.PlantA
			} else if conn.PlantB.ID == currentID && conn.PlantB.IsAlive {
				plant = conn.PlantB
			}

			if plant != nil {
				clusterPlants = append(clusterPlants, plant)
				break
			}
		}

		// Add connected plants to stack
		for _, connectedID := range plantConnections[currentID] {
			if !visited[connectedID] {
				stack = append(stack, connectedID)
			}
		}
	}

	return clusterPlants
}

// createNetworkCluster creates a new network cluster from a group of connected plants
func (pns *PlantNetworkSystem) createNetworkCluster(plants []*Plant, currentTick int) *NetworkCluster {
	cluster := &NetworkCluster{
		ID:                pns.NextClusterID,
		Plants:            plants,
		Connections:       make([]*NetworkConnection, 0),
		Size:              len(plants),
		FormationTick:     currentTick,
		LastResourceShare: currentTick,
	}

	pns.NextClusterID++

	// Find connections within this cluster
	plantIDs := make(map[int]bool)
	for _, plant := range plants {
		plantIDs[plant.ID] = true
	}

	totalResources := 0.0
	totalHealth := 0.0
	connectionCount := 0

	for _, conn := range pns.Connections {
		if plantIDs[conn.PlantA.ID] && plantIDs[conn.PlantB.ID] {
			cluster.Connections = append(cluster.Connections, conn)
			totalHealth += conn.Health
			connectionCount++
		}
	}

	// Calculate cluster properties
	for _, plant := range plants {
		totalResources += plant.Energy
	}

	cluster.TotalResources = totalResources
	if connectionCount > 0 {
		cluster.AverageHealth = totalHealth / float64(connectionCount)
		cluster.Efficiency = cluster.AverageHealth * math.Min(1.0, float64(len(cluster.Connections))/float64(len(plants)))
	}

	// Find central hub (most connected plant)
	connectionCounts := make(map[int]int)
	for _, conn := range cluster.Connections {
		connectionCounts[conn.PlantA.ID]++
		connectionCounts[conn.PlantB.ID]++
	}

	maxConnections := 0
	for _, plant := range plants {
		if connectionCounts[plant.ID] > maxConnections {
			maxConnections = connectionCounts[plant.ID]
			cluster.CentralHub = plant
		}
	}

	return cluster
}

// maintainNetworkHealth handles network maintenance and degradation
func (pns *PlantNetworkSystem) maintainNetworkHealth() {
	for _, conn := range pns.Connections {
		// Environmental stress can damage connections
		if rand.Float64() < 0.01 { // 1% chance of environmental damage per tick
			conn.Health = math.Max(0.0, conn.Health-0.05)
		}

		// Very weak connections may break entirely
		if conn.Health < 0.05 && rand.Float64() < 0.1 {
			conn.Health = 0.0
		}

		// Distance stress - longer connections are harder to maintain
		distanceStress := math.Max(0.0, (conn.Distance-5.0)/10.0) // Stress increases beyond 5 units
		if distanceStress > 0 && rand.Float64() < distanceStress*0.01 {
			conn.Health = math.Max(0.0, conn.Health-0.02)
		}
	}
}

// cleanup removes dead connections and expired signals
func (pns *PlantNetworkSystem) cleanup() {
	// Remove dead connections
	aliveConnections := make([]*NetworkConnection, 0)
	for _, conn := range pns.Connections {
		if conn.Health > 0 && conn.PlantA.IsAlive && conn.PlantB.IsAlive {
			aliveConnections = append(aliveConnections, conn)
		}
	}
	pns.Connections = aliveConnections

	// Remove expired signals
	activeSignals := make([]*ChemicalSignal, 0)
	for _, signal := range pns.ChemicalSignals {
		if signal.Age < signal.MaxAge && signal.Intensity > 0.01 {
			activeSignals = append(activeSignals, signal)
		}
	}
	pns.ChemicalSignals = activeSignals
}

// updateStatistics calculates current network statistics
func (pns *PlantNetworkSystem) updateStatistics() {
	pns.ActiveConnections = len(pns.Connections)
	pns.ActiveSignals = len(pns.ChemicalSignals)

	// Calculate overall network efficiency
	if len(pns.Connections) > 0 {
		totalEfficiency := 0.0
		for _, conn := range pns.Connections {
			totalEfficiency += conn.Efficiency * conn.Health
		}
		pns.NetworkEfficiency = totalEfficiency / float64(len(pns.Connections))
	} else {
		pns.NetworkEfficiency = 0.0
	}
}

// GetNetworkStats returns comprehensive statistics about the network system
func (pns *PlantNetworkSystem) GetNetworkStats() map[string]interface{} {
	connectionsByType := make(map[NetworkConnectionType]int)
	signalsByType := make(map[ChemicalSignalType]int)

	// Count active connections (healthy ones)
	activeConnections := 0
	totalConnectionStrength := 0.0
	healthyConnections := 0
	degradingConnections := 0

	for _, conn := range pns.Connections {
		connectionsByType[conn.Type]++
		if conn.Health > 0.5 {
			activeConnections++
			healthyConnections++
		} else if conn.Health > 0.1 {
			degradingConnections++
		}
		totalConnectionStrength += conn.Strength
	}

	avgConnectionStrength := 0.0
	if len(pns.Connections) > 0 {
		avgConnectionStrength = totalConnectionStrength / float64(len(pns.Connections))
	}

	for _, signal := range pns.ChemicalSignals {
		signalsByType[signal.Type]++
	}

	// Calculate average cluster size
	avgClusterSize := 0.0
	if len(pns.NetworkClusters) > 0 {
		totalSize := 0
		for _, cluster := range pns.NetworkClusters {
			totalSize += cluster.Size
		}
		avgClusterSize = float64(totalSize) / float64(len(pns.NetworkClusters))
	}

	// Build cluster information
	clusters := make([]map[string]interface{}, 0)
	for _, cluster := range pns.NetworkClusters {
		plantTypes := make([]string, 0)
		typeMap := make(map[PlantType]bool)

		for _, plant := range cluster.Plants {
			if !typeMap[plant.Type] {
				typeMap[plant.Type] = true
				configs := GetPlantConfigs()
				if config, exists := configs[plant.Type]; exists {
					plantTypes = append(plantTypes, config.Name)
				}
			}
		}

		clusterInfo := map[string]interface{}{
			"id":          cluster.ID,
			"size":        cluster.Size,
			"avg_health":  cluster.AverageHealth,
			"efficiency":  cluster.Efficiency,
			"plant_types": plantTypes,
		}
		clusters = append(clusters, clusterInfo)
	}

	// Resource sharing statistics
	resourceSharing := map[string]interface{}{
		"transfers_this_tick":         pns.ActiveConnections, // Approximation
		"total_resources_transferred": pns.TotalResourceTransferred,
		"avg_transfer_efficiency":     pns.NetworkEfficiency,
		"recent_beneficiaries":        len(pns.NetworkClusters),
	}

	// Network health metrics
	healthPercentage := 0.0
	if len(pns.Connections) > 0 {
		healthPercentage = float64(healthyConnections) / float64(len(pns.Connections))
	}

	networkHealth := map[string]interface{}{
		"healthy_percentage":    healthPercentage,
		"degrading_connections": degradingConnections,
		"connections_lost":      0, // Could track this in future
		"new_connections":       0, // Could track this in future
	}

	// Recent events (placeholder - could be enhanced with actual event tracking)
	recentEvents := []string{}
	if len(pns.Connections) > 0 {
		recentEvents = append(recentEvents, fmt.Sprintf("%d active network connections", activeConnections))
	}
	if len(pns.ChemicalSignals) > 0 {
		recentEvents = append(recentEvents, fmt.Sprintf("%d chemical signals propagating", len(pns.ChemicalSignals)))
	}
	if len(pns.NetworkClusters) > 0 {
		recentEvents = append(recentEvents, fmt.Sprintf("%d plant clusters formed", len(pns.NetworkClusters)))
	}

	return map[string]interface{}{
		// Basic stats (original keys)
		"total_connections":           len(pns.Connections),
		"total_signals":               len(pns.ChemicalSignals),
		"total_clusters":              len(pns.NetworkClusters),
		"network_efficiency":          pns.NetworkEfficiency,
		"total_resources_transferred": pns.TotalResourceTransferred,
		"connections_by_type":         connectionsByType,
		"signals_by_type":             signalsByType,
		"average_cluster_size":        avgClusterSize,
		"max_connection_distance":     pns.MaxConnectionDistance,

		// Enhanced stats (expected by CLI)
		"active_connections":      activeConnections,
		"cluster_count":           len(pns.NetworkClusters),
		"active_signals":          len(pns.ChemicalSignals),
		"avg_connection_strength": avgConnectionStrength,
		"connection_types":        connectionsByType,
		"signal_activity":         signalsByType,
		"clusters":                clusters,
		"resource_sharing":        resourceSharing,
		"recent_events":           recentEvents,
		"network_health":          networkHealth,
	}
}

// GetConnectionsForPlant returns all connections involving a specific plant
func (pns *PlantNetworkSystem) GetConnectionsForPlant(plant *Plant) []*NetworkConnection {
	connections := make([]*NetworkConnection, 0)
	for _, conn := range pns.Connections {
		if conn.PlantA.ID == plant.ID || conn.PlantB.ID == plant.ID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// GetClusterForPlant returns the network cluster containing a specific plant
func (pns *PlantNetworkSystem) GetClusterForPlant(plant *Plant) *NetworkCluster {
	for _, cluster := range pns.NetworkClusters {
		for _, clusterPlant := range cluster.Plants {
			if clusterPlant.ID == plant.ID {
				return cluster
			}
		}
	}
	return nil
}

// GetLargestCluster returns the largest network cluster
func (pns *PlantNetworkSystem) GetLargestCluster() *NetworkCluster {
	if len(pns.NetworkClusters) == 0 {
		return nil
	}

	largestCluster := pns.NetworkClusters[0]
	for _, cluster := range pns.NetworkClusters[1:] {
		if cluster.Size > largestCluster.Size {
			largestCluster = cluster
		}
	}
	return largestCluster
}

// GetTopClusters returns the N largest clusters, sorted by size
func (pns *PlantNetworkSystem) GetTopClusters(n int) []*NetworkCluster {
	clusters := make([]*NetworkCluster, len(pns.NetworkClusters))
	copy(clusters, pns.NetworkClusters)

	// Sort by size (descending)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	if n > len(clusters) {
		n = len(clusters)
	}

	return clusters[:n]
}

// DetectNetworkThreats checks for threats to network plants and sends warnings
func (pns *PlantNetworkSystem) DetectNetworkThreats(allEntities []*Entity, currentTick int) {
	threatRadius := 8.0 // Distance to consider entities as threats

	for _, cluster := range pns.NetworkClusters {
		if cluster.Size < 3 { // Only protect larger clusters
			continue
		}

		threatsDetected := 0
		for _, plant := range cluster.Plants {
			for _, entity := range allEntities {
				if !entity.IsAlive {
					continue
				}

				// Calculate distance to plant
				distance := math.Sqrt(math.Pow(entity.Position.X-plant.Position.X, 2) +
					math.Pow(entity.Position.Y-plant.Position.Y, 2))

				// Check if entity is a threat (predator or omnivore near plants)
				if distance < threatRadius && (entity.Species == "Predator" || entity.Species == "Omnivore") {
					threatsDetected++
					break // One threat per plant is enough to trigger response
				}
			}
		}

		// If significant threats detected, send network-wide warning
		if threatsDetected > 0 && rand.Float64() < 0.4 { // 40% chance to send warning
			centralHub := cluster.CentralHub
			if centralHub != nil {
				threatLevel := math.Min(1.0, float64(threatsDetected)/float64(cluster.Size))
				pns.sendChemicalSignal(centralHub, SignalThreatDetected, threatLevel, currentTick,
					"Herbivores detected near network plants")
			}
		}
	}
}
