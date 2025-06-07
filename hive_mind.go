package main

import (
	"math"
	"math/rand"
)

// HiveMindType represents different types of collective intelligence
type HiveMindType int

const (
	SimpleCollective  HiveMindType = iota // Basic shared awareness
	SwarmIntelligence                     // Coordinated group actions
	NeuralNetwork                         // Distributed decision making
	QuantumMind                           // Advanced collective consciousness
)

// String returns the string representation of HiveMindType
func (hmt HiveMindType) String() string {
	switch hmt {
	case SimpleCollective:
		return "SimpleCollective"
	case SwarmIntelligence:
		return "SwarmIntelligence"
	case NeuralNetwork:
		return "NeuralNetwork"
	case QuantumMind:
		return "QuantumMind"
	default:
		return "Unknown"
	}
}

// CollectiveMemory represents shared knowledge among hive members
type CollectiveMemory struct {
	FoodSources     map[Position]float64 `json:"food_sources"`     // Known food locations and quality
	ThreatAreas     map[Position]float64 `json:"threat_areas"`     // Dangerous areas to avoid
	SafeZones       map[Position]float64 `json:"safe_zones"`       // Safe locations for shelter
	TradeRoutes     []Position           `json:"trade_routes"`     // Known trading paths
	SuccessPatterns map[string]float64   `json:"success_patterns"` // Learned successful behaviors
	SharedDecisions map[string]float64   `json:"shared_decisions"` // Group consensus on decisions
}

// HiveMind represents a collective intelligence system
type HiveMind struct {
	ID                int               `json:"id"`
	Members           []*Entity         `json:"-"` // Exclude from JSON to avoid cycles
	MemberIDs         []int             `json:"member_ids"`
	Type              HiveMindType      `json:"type"`
	CollectiveMemory  *CollectiveMemory `json:"collective_memory"`
	Intelligence      float64           `json:"intelligence"`       // Combined intelligence of members
	Cohesion          float64           `json:"cohesion"`           // How well the hive works together
	DecisionThreshold float64           `json:"decision_threshold"` // Threshold for group decisions
	MemoryRetention   float64           `json:"memory_retention"`   // How long memories are retained
	MaxMembers        int               `json:"max_members"`        // Maximum hive size
	CreationTick      int               `json:"creation_tick"`
}

// NewHiveMind creates a new hive mind from compatible entities
func NewHiveMind(id int, founder *Entity, hiveType HiveMindType) *HiveMind {
	return &HiveMind{
		ID:                id,
		Members:           []*Entity{founder},
		MemberIDs:         []int{founder.ID},
		Type:              hiveType,
		CollectiveMemory:  NewCollectiveMemory(),
		Intelligence:      founder.GetTrait("intelligence"),
		Cohesion:          founder.GetTrait("cooperation"),
		DecisionThreshold: 0.6,  // 60% consensus needed for decisions
		MemoryRetention:   0.95, // 95% memory retention per tick
		MaxMembers:        getMaxMembersForType(hiveType),
	}
}

// NewCollectiveMemory creates a new collective memory system
func NewCollectiveMemory() *CollectiveMemory {
	return &CollectiveMemory{
		FoodSources:     make(map[Position]float64),
		ThreatAreas:     make(map[Position]float64),
		SafeZones:       make(map[Position]float64),
		TradeRoutes:     make([]Position, 0),
		SuccessPatterns: make(map[string]float64),
		SharedDecisions: make(map[string]float64),
	}
}

// getMaxMembersForType returns the maximum members for each hive type
func getMaxMembersForType(hiveType HiveMindType) int {
	switch hiveType {
	case SimpleCollective:
		return 10
	case SwarmIntelligence:
		return 50
	case NeuralNetwork:
		return 25
	case QuantumMind:
		return 100
	default:
		return 10
	}
}

// CanJoinHive determines if an entity can join this hive mind
func (hm *HiveMind) CanJoinHive(entity *Entity) bool {
	if !entity.IsAlive || len(hm.Members) >= hm.MaxMembers {
		return false
	}

	// Check if entity is already in the hive
	for _, member := range hm.Members {
		if member.ID == entity.ID {
			return false
		}
	}

	// Compatibility requirements
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")

	// Must have minimum intelligence and cooperation
	if intelligence < 0.3 || cooperation < 0.4 {
		return false
	}

	// Check compatibility with existing hive members
	avgIntelligence := hm.Intelligence / float64(len(hm.Members))
	avgCooperation := hm.Cohesion / float64(len(hm.Members))

	// New member should be within reasonable range of existing members
	intelligenceDiff := math.Abs(intelligence - avgIntelligence)
	cooperationDiff := math.Abs(cooperation - avgCooperation)

	return intelligenceDiff < 0.5 && cooperationDiff < 0.3
}

// AddMember adds an entity to the hive mind
func (hm *HiveMind) AddMember(entity *Entity) bool {
	if !hm.CanJoinHive(entity) {
		return false
	}

	hm.Members = append(hm.Members, entity)
	hm.MemberIDs = append(hm.MemberIDs, entity.ID)

	// Update collective intelligence and cohesion
	hm.Intelligence += entity.GetTrait("intelligence")
	hm.Cohesion += entity.GetTrait("cooperation")

	// Mark entity as part of hive
	entity.SetTrait("hive_member", 1.0)
	entity.SetTrait("hive_id", float64(hm.ID))

	return true
}

// RemoveMember removes an entity from the hive mind
func (hm *HiveMind) RemoveMember(entity *Entity) {
	for i, member := range hm.Members {
		if member.ID == entity.ID {
			// Remove from members slice
			hm.Members = append(hm.Members[:i], hm.Members[i+1:]...)
			// Remove from member IDs
			for j, id := range hm.MemberIDs {
				if id == entity.ID {
					hm.MemberIDs = append(hm.MemberIDs[:j], hm.MemberIDs[j+1:]...)
					break
				}
			}

			// Update collective intelligence and cohesion
			hm.Intelligence -= entity.GetTrait("intelligence")
			hm.Cohesion -= entity.GetTrait("cooperation")

			// Remove hive markings from entity
			entity.SetTrait("hive_member", 0.0)
			entity.SetTrait("hive_id", 0.0)
			break
		}
	}
}

// ShareKnowledge adds knowledge to the collective memory
func (hm *HiveMind) ShareKnowledge(knowledgeType string, position Position, value float64) {
	switch knowledgeType {
	case "food":
		hm.CollectiveMemory.FoodSources[position] = value
	case "threat":
		hm.CollectiveMemory.ThreatAreas[position] = value
	case "safe":
		hm.CollectiveMemory.SafeZones[position] = value
	}
}

// GetCollectiveDecision makes a group decision based on member input
func (hm *HiveMind) GetCollectiveDecision(decisionType string, options []string) string {
	if len(hm.Members) == 0 {
		return ""
	}

	// Collect votes from members based on their traits and intelligence
	votes := make(map[string]float64)

	for _, member := range hm.Members {
		if !member.IsAlive {
			continue
		}

		// Member's influence is based on intelligence and cooperation
		influence := (member.GetTrait("intelligence") + member.GetTrait("cooperation")) / 2.0

		// Simple decision making based on member traits
		var preferredOption string
		switch decisionType {
		case "food_search":
			if member.Energy < 30 {
				preferredOption = "aggressive_search"
			} else {
				preferredOption = "conservative_search"
			}
		case "threat_response":
			if member.GetTrait("aggression") > 0.5 {
				preferredOption = "fight"
			} else {
				preferredOption = "flee"
			}
		case "territory":
			if member.GetTrait("territorial") > 0.3 {
				preferredOption = "defend"
			} else {
				preferredOption = "retreat"
			}
		default:
			// Random choice for unknown decision types
			if len(options) > 0 {
				preferredOption = options[rand.Intn(len(options))]
			}
		}

		votes[preferredOption] += influence
	}

	// Find the option with highest vote weight
	bestOption := ""
	bestScore := 0.0
	totalVotes := 0.0

	for option, score := range votes {
		totalVotes += score
		if score > bestScore {
			bestScore = score
			bestOption = option
		}
	}

	// Only make decision if enough consensus (above threshold)
	if totalVotes > 0 && (bestScore/totalVotes) >= hm.DecisionThreshold {
		hm.CollectiveMemory.SharedDecisions[decisionType] = bestScore / totalVotes
		return bestOption
	}

	return "" // No consensus reached
}

// CoordinateMovement coordinates movement of hive members
func (hm *HiveMind) CoordinateMovement(targetX, targetY float64) {
	if len(hm.Members) == 0 {
		return
	}

	// Calculate formation based on hive type
	formations := hm.calculateFormation(targetX, targetY)

	for i, member := range hm.Members {
		if !member.IsAlive || i >= len(formations) {
			continue
		}

		formation := formations[i]
		speed := member.GetTrait("speed") * 0.8 // Slightly slower when coordinated

		// Move toward formation position
		member.MoveTo(formation.X, formation.Y, speed)
	}
}

// calculateFormation determines optimal positions for coordinated movement
func (hm *HiveMind) calculateFormation(targetX, targetY float64) []Position {
	formations := make([]Position, len(hm.Members))

	switch hm.Type {
	case SimpleCollective:
		// Simple cluster formation
		for i := range hm.Members {
			angle := float64(i) * 2.0 * math.Pi / float64(len(hm.Members))
			radius := 5.0
			formations[i] = Position{
				X: targetX + math.Cos(angle)*radius,
				Y: targetY + math.Sin(angle)*radius,
			}
		}

	case SwarmIntelligence:
		// Dense swarm formation
		gridSize := int(math.Ceil(math.Sqrt(float64(len(hm.Members)))))
		spacing := 2.0

		for i := range hm.Members {
			row := i / gridSize
			col := i % gridSize
			offsetX := float64(col-gridSize/2) * spacing
			offsetY := float64(row-gridSize/2) * spacing

			formations[i] = Position{
				X: targetX + offsetX,
				Y: targetY + offsetY,
			}
		}

	case NeuralNetwork:
		// Hierarchical formation based on intelligence
		// Sort by intelligence (simplified)
		for i := range hm.Members {
			layer := float64(i%3) - 1.0 // -1, 0, 1
			position := float64(i/3) * 3.0

			formations[i] = Position{
				X: targetX + position,
				Y: targetY + layer*4.0,
			}
		}

	case QuantumMind:
		// Dynamic formation that adapts to environment
		for i := range hm.Members {
			angle := float64(i) * 2.0 * math.Pi / float64(len(hm.Members))
			// Variable radius based on member intelligence
			intelligence := hm.Members[i].GetTrait("intelligence")
			radius := 3.0 + intelligence*7.0

			formations[i] = Position{
				X: targetX + math.Cos(angle)*radius,
				Y: targetY + math.Sin(angle)*radius,
			}
		}

	default:
		// Default to random positions around target
		for i := range hm.Members {
			formations[i] = Position{
				X: targetX + (rand.Float64()-0.5)*10.0,
				Y: targetY + (rand.Float64()-0.5)*10.0,
			}
		}
	}

	return formations
}

// Update maintains the hive mind system
func (hm *HiveMind) Update() {
	if len(hm.Members) == 0 {
		return
	}

	// Remove dead members
	aliveMembers := make([]*Entity, 0)
	aliveMemberIDs := make([]int, 0)

	for i, member := range hm.Members {
		if member.IsAlive {
			aliveMembers = append(aliveMembers, member)
			aliveMemberIDs = append(aliveMemberIDs, hm.MemberIDs[i])
		} else {
			// Update collective intelligence when member dies
			hm.Intelligence -= member.GetTrait("intelligence")
			hm.Cohesion -= member.GetTrait("cooperation")
		}
	}

	hm.Members = aliveMembers
	hm.MemberIDs = aliveMemberIDs

	// Decay old memories
	hm.decayMemories()

	// Update collective intelligence
	hm.updateCollectiveIntelligence()
}

// decayMemories reduces the strength of old memories
func (hm *HiveMind) decayMemories() {
	// Decay food source memories
	for pos, strength := range hm.CollectiveMemory.FoodSources {
		newStrength := strength * hm.MemoryRetention
		if newStrength < 0.1 {
			delete(hm.CollectiveMemory.FoodSources, pos)
		} else {
			hm.CollectiveMemory.FoodSources[pos] = newStrength
		}
	}

	// Decay threat area memories
	for pos, strength := range hm.CollectiveMemory.ThreatAreas {
		newStrength := strength * hm.MemoryRetention
		if newStrength < 0.1 {
			delete(hm.CollectiveMemory.ThreatAreas, pos)
		} else {
			hm.CollectiveMemory.ThreatAreas[pos] = newStrength
		}
	}

	// Decay safe zone memories
	for pos, strength := range hm.CollectiveMemory.SafeZones {
		newStrength := strength * hm.MemoryRetention
		if newStrength < 0.1 {
			delete(hm.CollectiveMemory.SafeZones, pos)
		} else {
			hm.CollectiveMemory.SafeZones[pos] = newStrength
		}
	}
}

// updateCollectiveIntelligence recalculates the collective intelligence
func (hm *HiveMind) updateCollectiveIntelligence() {
	if len(hm.Members) == 0 {
		hm.Intelligence = 0
		hm.Cohesion = 0
		return
	}

	totalIntelligence := 0.0
	totalCooperation := 0.0

	for _, member := range hm.Members {
		totalIntelligence += member.GetTrait("intelligence")
		totalCooperation += member.GetTrait("cooperation")
	}

	// Collective intelligence is greater than sum of parts for well-coordinated hives
	synergy := 1.0 + (totalCooperation/float64(len(hm.Members)))*0.5
	hm.Intelligence = totalIntelligence * synergy
	hm.Cohesion = totalCooperation / float64(len(hm.Members))
}

// GetBestFoodSource returns the best known food source
func (hm *HiveMind) GetBestFoodSource() (Position, float64, bool) {
	bestPos := Position{}
	bestValue := 0.0
	found := false

	for pos, value := range hm.CollectiveMemory.FoodSources {
		if value > bestValue {
			bestValue = value
			bestPos = pos
			found = true
		}
	}

	return bestPos, bestValue, found
}

// IsPositionSafe checks if a position is considered safe by the hive
func (hm *HiveMind) IsPositionSafe(pos Position) bool {
	// Check threat areas
	for threatPos, strength := range hm.CollectiveMemory.ThreatAreas {
		distance := math.Sqrt(math.Pow(pos.X-threatPos.X, 2) + math.Pow(pos.Y-threatPos.Y, 2))
		if distance < 10.0 && strength > 0.5 {
			return false // Too close to known threat
		}
	}

	// Check safe zones
	for safePos, strength := range hm.CollectiveMemory.SafeZones {
		distance := math.Sqrt(math.Pow(pos.X-safePos.X, 2) + math.Pow(pos.Y-safePos.Y, 2))
		if distance < 5.0 && strength > 0.7 {
			return true // Within known safe zone
		}
	}

	return true // Neutral - no known threats or safe zones
}

// HiveMindSystem manages all hive minds in the simulation
type HiveMindSystem struct {
	HiveMinds      []*HiveMind `json:"hive_minds"`
	NextHiveMindID int         `json:"next_hive_mind_id"`
}

// NewHiveMindSystem creates a new hive mind management system
func NewHiveMindSystem() *HiveMindSystem {
	return &HiveMindSystem{
		HiveMinds:      make([]*HiveMind, 0),
		NextHiveMindID: 1,
	}
}

// TryFormHiveMind attempts to form a new hive mind from compatible entities
func (hms *HiveMindSystem) TryFormHiveMind(entities []*Entity, hiveType HiveMindType) *HiveMind {
	if len(entities) < 2 {
		return nil
	}

	// Find the most intelligent entity as founder
	var founder *Entity
	maxIntelligence := -1.0

	for _, entity := range entities {
		if entity.IsAlive && entity.GetTrait("intelligence") > maxIntelligence {
			// Check if entity is not already in a hive
			if entity.GetTrait("hive_member") == 0.0 {
				maxIntelligence = entity.GetTrait("intelligence")
				founder = entity
			}
		}
	}

	if founder == nil || maxIntelligence < 0.4 {
		return nil // No suitable founder found
	}

	// Create new hive mind
	hiveMind := NewHiveMind(hms.NextHiveMindID, founder, hiveType)
	hms.NextHiveMindID++

	// Try to add other compatible entities
	for _, entity := range entities {
		if entity != founder && hiveMind.CanJoinHive(entity) {
			hiveMind.AddMember(entity)
		}
	}

	// Only create hive if we have at least 2 members
	if len(hiveMind.Members) >= 2 {
		hms.HiveMinds = append(hms.HiveMinds, hiveMind)
		return hiveMind
	}

	return nil
}

// Update maintains all hive minds
func (hms *HiveMindSystem) Update() {
	activeHiveMinds := make([]*HiveMind, 0)

	for _, hiveMind := range hms.HiveMinds {
		hiveMind.Update()

		// Keep hive minds with at least 2 members
		if len(hiveMind.Members) >= 2 {
			activeHiveMinds = append(activeHiveMinds, hiveMind)
		}
	}

	hms.HiveMinds = activeHiveMinds
}

// GetHiveMindByMember finds the hive mind that contains a specific entity
func (hms *HiveMindSystem) GetHiveMindByMember(entity *Entity) *HiveMind {
	for _, hiveMind := range hms.HiveMinds {
		for _, member := range hiveMind.Members {
			if member.ID == entity.ID {
				return hiveMind
			}
		}
	}
	return nil
}
