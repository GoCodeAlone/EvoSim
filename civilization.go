package main

import (
	"fmt"
	"math"
	"math/rand"
)

// StructureType represents different types of structures entities can build
type StructureType int

const (
	StructureNest    StructureType = iota // Basic shelter
	StructureCache                        // Food storage
	StructureBarrier                      // Defensive wall
	StructureTrap                         // Hunting trap
	StructureFarm                         // Cultivated plant area
	StructureWell                         // Water source
	StructureTower                        // Observation post
	StructureMarket                       // Trading post
)

// Structure represents a built structure in the world
type Structure struct {
	ID               int
	Type             StructureType
	Position         Position
	Builder          *Entity // Entity that built it
	Tribe            *Tribe  // Tribe that owns it
	Health           float64 // Structural integrity
	MaxHealth        float64
	Resources        map[string]float64 // Stored resources
	Capacity         float64            // Storage/functionality capacity
	IsActive         bool
	MaintenanceeCost float64 // Energy cost per tick to maintain
	CreationTick     int
}

// NewStructure creates a new structure
func NewStructure(id int, structureType StructureType, position Position, builder *Entity) *Structure {
	maxHealth := 100.0
	capacity := 50.0
	maintenanceCost := 1.0

	// Adjust properties based on structure type
	switch structureType {
	case StructureNest:
		maxHealth = 150.0
		capacity = 20.0
		maintenanceCost = 0.5
	case StructureCache:
		maxHealth = 80.0
		capacity = 100.0
		maintenanceCost = 0.3
	case StructureBarrier:
		maxHealth = 200.0
		capacity = 0.0
		maintenanceCost = 0.2
	case StructureFarm:
		maxHealth = 60.0
		capacity = 30.0
		maintenanceCost = 2.0
	case StructureMarket:
		maxHealth = 120.0
		capacity = 200.0
		maintenanceCost = 1.5
	}

	return &Structure{
		ID:               id,
		Type:             structureType,
		Position:         position,
		Builder:          builder,
		Health:           maxHealth,
		MaxHealth:        maxHealth,
		Resources:        make(map[string]float64),
		Capacity:         capacity,
		IsActive:         true,
		MaintenanceeCost: maintenanceCost,
	}
}

// Update maintains the structure
func (s *Structure) Update() {
	if !s.IsActive {
		return
	}

	// Natural decay
	s.Health -= 0.1

	// Structure-specific updates
	switch s.Type {
	case StructureFarm:
		s.updateFarm()
	case StructureCache:
		s.updateCache()
	case StructureTrap:
		s.updateTrap()
	}

	// Deactivate if health is too low
	if s.Health <= 0 {
		s.IsActive = false
	}
}

// updateFarm handles farm production
func (s *Structure) updateFarm() {
	// Farms produce food over time if maintained
	if s.Health > s.MaxHealth*0.5 {
		production := 2.0 + rand.Float64()*3.0
		s.Resources["food"] += production

		// Limit storage to capacity
		if s.Resources["food"] > s.Capacity {
			s.Resources["food"] = s.Capacity
		}
	}
}

// updateCache handles resource storage decay
func (s *Structure) updateCache() {
	// Small spoilage rate for organic materials
	if food, exists := s.Resources["food"]; exists {
		s.Resources["food"] = food * 0.99 // 1% spoilage
	}
}

// updateTrap handles trap functionality
func (s *Structure) updateTrap() {
	// Traps have a chance to catch prey
	if rand.Float64() < 0.05 { // 5% chance per tick
		s.Resources["food"] += 10.0 + rand.Float64()*15.0
	}
}

// Repair attempts to repair the structure
func (s *Structure) Repair(entity *Entity, amount float64) bool {
	intelligence := entity.GetTrait("intelligence")
	if intelligence < 0.3 {
		return false // Not smart enough to repair
	}

	energyCost := amount * 2.0
	if entity.Energy < energyCost {
		return false // Not enough energy
	}

	entity.Energy -= energyCost
	s.Health = math.Min(s.MaxHealth, s.Health+amount)
	return true
}

// Tribe represents a collection of entities working together
type Tribe struct {
	ID         int
	Name       string
	Members    []*Entity
	Leader     *Entity
	Structures []*Structure
	Territory  []Position // Claimed territory
	Resources  map[string]float64
	TechLevel  int                // Technological advancement
	Culture    map[string]float64 // Cultural traits
	Alliances  []*Tribe           // Allied tribes
	Enemies    []*Tribe           // Enemy tribes
}

// NewTribe creates a new tribe
func NewTribe(id int, name string, founder *Entity) *Tribe {
	return &Tribe{
		ID:         id,
		Name:       name,
		Members:    []*Entity{founder},
		Leader:     founder,
		Structures: make([]*Structure, 0),
		Territory:  make([]Position, 0),
		Resources: map[string]float64{
			"food":  50.0,
			"wood":  20.0,
			"stone": 10.0,
		},
		TechLevel: 1,
		Culture: map[string]float64{
			"cooperation": 0.5,
			"aggression":  0.3,
			"innovation":  0.4,
		},
		Alliances: make([]*Tribe, 0),
		Enemies:   make([]*Tribe, 0),
	}
}

// AddMember adds an entity to the tribe
func (t *Tribe) AddMember(entity *Entity) {
	t.Members = append(t.Members, entity)

	// Update cultural traits based on new member
	cooperation := entity.GetTrait("cooperation")
	aggression := entity.GetTrait("aggression")
	intelligence := entity.GetTrait("intelligence")

	memberCount := float64(len(t.Members))
	t.Culture["cooperation"] = (t.Culture["cooperation"]*(memberCount-1) + cooperation) / memberCount
	t.Culture["aggression"] = (t.Culture["aggression"]*(memberCount-1) + aggression) / memberCount
	t.Culture["innovation"] = (t.Culture["innovation"]*(memberCount-1) + intelligence) / memberCount
}

// CanBuild checks if the tribe can build a structure
func (t *Tribe) CanBuild(structureType StructureType) bool {
	cost := t.getBuildingCost(structureType)

	for resource, required := range cost {
		if t.Resources[resource] < required {
			return false
		}
	}

	return t.TechLevel >= t.getRequiredTechLevel(structureType)
}

// getBuildingCost returns the resource cost for a structure
func (t *Tribe) getBuildingCost(structureType StructureType) map[string]float64 {
	switch structureType {
	case StructureNest:
		return map[string]float64{"wood": 20.0}
	case StructureCache:
		return map[string]float64{"wood": 15.0, "stone": 5.0}
	case StructureBarrier:
		return map[string]float64{"stone": 30.0, "wood": 10.0}
	case StructureFarm:
		return map[string]float64{"wood": 10.0, "food": 20.0}
	case StructureTrap:
		return map[string]float64{"wood": 25.0}
	case StructureWell:
		return map[string]float64{"stone": 40.0}
	case StructureTower:
		return map[string]float64{"stone": 50.0, "wood": 30.0}
	case StructureMarket:
		return map[string]float64{"wood": 60.0, "stone": 40.0}
	default:
		return map[string]float64{}
	}
}

// getRequiredTechLevel returns minimum tech level for a structure
func (t *Tribe) getRequiredTechLevel(structureType StructureType) int {
	switch structureType {
	case StructureNest, StructureCache:
		return 1
	case StructureBarrier, StructureTrap:
		return 2
	case StructureFarm, StructureWell:
		return 3
	case StructureTower, StructureMarket:
		return 4
	default:
		return 1
	}
}

// BuildStructure attempts to build a structure
func (t *Tribe) BuildStructure(structureType StructureType, position Position, builder *Entity, nextID int, eventBus *CentralEventBus, tick int) *Structure {
	if !t.CanBuild(structureType) {
		return nil
	}

	// Deduct resources
	cost := t.getBuildingCost(structureType)
	for resource, required := range cost {
		t.Resources[resource] -= required
	}

	// Create structure
	structure := NewStructure(nextID, structureType, position, builder)
	structure.Tribe = t
	structure.CreationTick = tick
	t.Structures = append(t.Structures, structure)

	// Emit event for structure building
	if eventBus != nil {
		structureTypeNames := []string{"nest", "cache", "barrier", "trap", "farm", "well", "tower", "market"}
		structureTypeName := "unknown"
		if int(structureType) < len(structureTypeNames) {
			structureTypeName = structureTypeNames[structureType]
		}

		metadata := map[string]interface{}{
			"structure_id":   structure.ID,
			"structure_type": structureTypeName,
			"tribe_id":       t.ID,
			"tribe_name":     t.Name,
			"builder_id":     builder.ID,
			"tech_level":     t.TechLevel,
			"cost":           cost,
			"max_health":     structure.MaxHealth,
			"capacity":       structure.Capacity,
		}

		eventBus.EmitSystemEvent(tick, "structure_built", "civilization", "civilization_system",
			fmt.Sprintf("Tribe %s built %s at (%.1f, %.1f)", t.Name, structureTypeName, position.X, position.Y),
			&position, metadata)
	}

	return structure
}

// Update maintains the tribe
func (t *Tribe) Update(eventBus *CentralEventBus, tick int) {
	originalMemberCount := len(t.Members)

	// Remove dead members
	aliveMembers := make([]*Entity, 0)
	for _, member := range t.Members {
		if member.IsAlive {
			aliveMembers = append(aliveMembers, member)
		}
	}
	t.Members = aliveMembers

	// Disband if no members
	if len(t.Members) == 0 {
		if eventBus != nil {
			metadata := map[string]interface{}{
				"tribe_id":         t.ID,
				"tribe_name":       t.Name,
				"original_members": originalMemberCount,
				"tech_level":       t.TechLevel,
				"structure_count":  len(t.Structures),
			}

			eventBus.EmitSystemEvent(tick, "tribe_disbanded", "civilization", "civilization_system",
				fmt.Sprintf("Tribe %s disbanded (no surviving members)", t.Name),
				nil, metadata)
		}
		return
	}

	// Update leader if current leader is dead
	oldLeaderID := 0
	if t.Leader != nil {
		oldLeaderID = t.Leader.ID
	}

	if t.Leader == nil || !t.Leader.IsAlive {
		oldLeader := t.Leader
		t.electNewLeader()

		if eventBus != nil && t.Leader != nil && (oldLeader == nil || t.Leader.ID != oldLeaderID) {
			metadata := map[string]interface{}{
				"tribe_id":      t.ID,
				"tribe_name":    t.Name,
				"old_leader_id": oldLeaderID,
				"new_leader_id": t.Leader.ID,
				"member_count":  len(t.Members),
			}

			eventBus.EmitSystemEvent(tick, "tribe_leader_changed", "civilization", "civilization_system",
				fmt.Sprintf("Tribe %s elected new leader (entity %d)", t.Name, t.Leader.ID),
				&t.Leader.Position, metadata)
		}
	}

	// Collect resources from structures
	t.collectResources()

	// Research and development
	oldTechLevel := t.TechLevel
	t.advanceTechnology()

	if eventBus != nil && t.TechLevel > oldTechLevel {
		metadata := map[string]interface{}{
			"tribe_id":       t.ID,
			"tribe_name":     t.Name,
			"old_tech_level": oldTechLevel,
			"new_tech_level": t.TechLevel,
			"member_count":   len(t.Members),
		}

		eventBus.EmitSystemEvent(tick, "tech_advancement", "civilization", "civilization_system",
			fmt.Sprintf("Tribe %s advanced to tech level %d", t.Name, t.TechLevel),
			&t.Leader.Position, metadata)
	}

	// Maintain structures
	t.maintainStructures()
}

// electNewLeader chooses a new leader
func (t *Tribe) electNewLeader() {
	if len(t.Members) == 0 {
		t.Leader = nil
		return
	}

	// Choose member with highest intelligence and cooperation
	bestScore := -1.0
	var newLeader *Entity

	for _, member := range t.Members {
		score := member.GetTrait("intelligence")*0.6 + member.GetTrait("cooperation")*0.4
		if score > bestScore {
			bestScore = score
			newLeader = member
		}
	}

	t.Leader = newLeader
}

// collectResources gathers resources from tribe structures
func (t *Tribe) collectResources() {
	for _, structure := range t.Structures {
		if !structure.IsActive {
			continue
		}

		// Transfer resources from structures to tribe
		for resource, amount := range structure.Resources {
			t.Resources[resource] += amount
			structure.Resources[resource] = 0
		}
	}
}

// advanceTechnology improves tribe's technological level
func (t *Tribe) advanceTechnology() {
	if len(t.Members) == 0 {
		return
	}

	// Research rate based on average intelligence and innovation culture
	avgIntelligence := 0.0
	for _, member := range t.Members {
		avgIntelligence += member.GetTrait("intelligence")
	}
	avgIntelligence /= float64(len(t.Members))

	researchRate := avgIntelligence * t.Culture["innovation"] * 0.001

	if rand.Float64() < researchRate {
		t.TechLevel++
	}
}

// maintainStructures handles structure maintenance
func (t *Tribe) maintainStructures() {
	activeStructures := make([]*Structure, 0)

	for _, structure := range t.Structures {
		if structure.IsActive {
			// Pay maintenance cost
			if t.Resources["food"] >= structure.MaintenanceeCost {
				t.Resources["food"] -= structure.MaintenanceeCost
				activeStructures = append(activeStructures, structure)
			} else {
				// Cannot maintain, structure degrades faster
				structure.Health -= 2.0
			}
		} else if structure.Health > 0 {
			// Keep inactive structures for potential repair
			activeStructures = append(activeStructures, structure)
		}
	}

	t.Structures = activeStructures
}

// TradeSystem manages inter-tribe trading
type TradeSystem struct {
	TradeRoutes  map[int]map[int]float64 // tribe1 -> tribe2 -> trade relationship
	ActiveTrades []Trade
}

// Trade represents an active trade between tribes
type Trade struct {
	ID         int
	FromTribe  *Tribe
	ToTribe    *Tribe
	Offering   map[string]float64
	Requesting map[string]float64
	Duration   int    // Ticks remaining
	Status     string // "proposed", "active", "completed", "cancelled"
}

// NewTradeSystem creates a new trade system
func NewTradeSystem() *TradeSystem {
	return &TradeSystem{
		TradeRoutes:  make(map[int]map[int]float64),
		ActiveTrades: make([]Trade, 0),
	}
}

// ProposeTrace creates a new trade proposal
func (ts *TradeSystem) ProposeTrade(fromTribe, toTribe *Tribe, offering, requesting map[string]float64) *Trade {
	// Check if tribes have established trade relations
	relationship := ts.getTradeRelationship(fromTribe.ID, toTribe.ID)
	if relationship < 0.2 {
		return nil // Not enough trust to trade
	}

	trade := &Trade{
		ID:         len(ts.ActiveTrades) + 1,
		FromTribe:  fromTribe,
		ToTribe:    toTribe,
		Offering:   offering,
		Requesting: requesting,
		Duration:   100, // 100 ticks to complete
		Status:     "proposed",
	}

	ts.ActiveTrades = append(ts.ActiveTrades, *trade)
	return trade
}

// getTradeRelationship returns trade relationship strength between tribes
func (ts *TradeSystem) getTradeRelationship(tribe1ID, tribe2ID int) float64 {
	if routes, exists := ts.TradeRoutes[tribe1ID]; exists {
		if relationship, exists := routes[tribe2ID]; exists {
			return relationship
		}
	}
	return 0.1 // Default minimal relationship
}

// ProcessTrades handles active trades
func (ts *TradeSystem) ProcessTrades() {
	activeTrades := make([]Trade, 0)

	for i := range ts.ActiveTrades {
		trade := &ts.ActiveTrades[i]

		switch trade.Status {
		case "proposed":
			// Check if receiving tribe accepts
			if ts.shouldAcceptTrade(trade) {
				trade.Status = "active"
				ts.executeTrade(trade)
			} else {
				trade.Status = "cancelled"
			}
		case "active":
			trade.Duration--
			if trade.Duration <= 0 {
				trade.Status = "completed"
				ts.improveTradRelations(trade.FromTribe.ID, trade.ToTribe.ID)
			}
		}

		if trade.Status == "active" {
			activeTrades = append(activeTrades, *trade)
		}
	}

	ts.ActiveTrades = activeTrades
}

// shouldAcceptTrade determines if a tribe should accept a trade proposal
func (ts *TradeSystem) shouldAcceptTrade(trade *Trade) bool {
	// Calculate trade value for receiving tribe
	offerValue := 0.0
	requestValue := 0.0

	for resource, amount := range trade.Offering {
		offerValue += amount * ts.getResourceValue(trade.ToTribe, resource)
	}

	for resource, amount := range trade.Requesting {
		requestValue += amount * ts.getResourceValue(trade.ToTribe, resource)
	}

	// Accept if offer is worth more than what's requested
	return offerValue > requestValue*1.1 // 10% profit margin
}

// getResourceValue calculates how valuable a resource is to a tribe
func (ts *TradeSystem) getResourceValue(tribe *Tribe, resource string) float64 {
	currentAmount := tribe.Resources[resource]

	// Value is inversely related to current stock
	switch resource {
	case "food":
		if currentAmount < 20 {
			return 2.0 // High value when scarce
		} else if currentAmount > 100 {
			return 0.5 // Low value when abundant
		}
		return 1.0
	case "wood", "stone":
		if currentAmount < 10 {
			return 1.5
		} else if currentAmount > 50 {
			return 0.3
		}
		return 0.8
	default:
		return 1.0
	}
}

// executeTrade performs the actual resource exchange
func (ts *TradeSystem) executeTrade(trade *Trade) {
	// Check if both tribes can fulfill their parts
	canFromTribeTrade := true
	canToTribeTrade := true

	for resource, amount := range trade.Offering {
		if trade.FromTribe.Resources[resource] < amount {
			canFromTribeTrade = false
			break
		}
	}

	for resource, amount := range trade.Requesting {
		if trade.ToTribe.Resources[resource] < amount {
			canToTribeTrade = false
			break
		}
	}

	if !canFromTribeTrade || !canToTribeTrade {
		trade.Status = "cancelled"
		return
	}

	// Execute the trade
	for resource, amount := range trade.Offering {
		trade.FromTribe.Resources[resource] -= amount
		trade.ToTribe.Resources[resource] += amount
	}

	for resource, amount := range trade.Requesting {
		trade.ToTribe.Resources[resource] -= amount
		trade.FromTribe.Resources[resource] += amount
	}
}

// improveTradRelations increases trade relationship between tribes
func (ts *TradeSystem) improveTradRelations(tribe1ID, tribe2ID int) {
	if _, exists := ts.TradeRoutes[tribe1ID]; !exists {
		ts.TradeRoutes[tribe1ID] = make(map[int]float64)
	}
	if _, exists := ts.TradeRoutes[tribe2ID]; !exists {
		ts.TradeRoutes[tribe2ID] = make(map[int]float64)
	}

	// Improve mutual relationship
	current1 := ts.TradeRoutes[tribe1ID][tribe2ID]
	current2 := ts.TradeRoutes[tribe2ID][tribe1ID]

	ts.TradeRoutes[tribe1ID][tribe2ID] = math.Min(1.0, current1+0.1)
	ts.TradeRoutes[tribe2ID][tribe1ID] = math.Min(1.0, current2+0.1)
}

// CivilizationSystem manages all civilization features
type CivilizationSystem struct {
	Tribes          []*Tribe
	Structures      []*Structure
	TradeSystem     *TradeSystem
	NextTribeID     int
	NextStructureID int
	EventBus        *CentralEventBus // For event tracking
}

// NewCivilizationSystem creates a new civilization system
func NewCivilizationSystem(eventBus *CentralEventBus) *CivilizationSystem {
	return &CivilizationSystem{
		Tribes:          make([]*Tribe, 0),
		Structures:      make([]*Structure, 0),
		TradeSystem:     NewTradeSystem(),
		NextTribeID:     1,
		NextStructureID: 1,
		EventBus:        eventBus,
	}
}

// Update maintains all civilization features
func (cs *CivilizationSystem) Update(tick int) {
	// Update all tribes
	activeTrbs := make([]*Tribe, 0)
	for _, tribe := range cs.Tribes {
		tribe.Update(cs.EventBus, tick)
		if len(tribe.Members) > 0 {
			activeTrbs = append(activeTrbs, tribe)
		}
	}
	cs.Tribes = activeTrbs

	// Update all structures
	activeStructures := make([]*Structure, 0)
	for _, structure := range cs.Structures {
		wasActive := structure.IsActive
		structure.Update()

		// Emit event for structure destruction
		if wasActive && !structure.IsActive && cs.EventBus != nil {
			structureTypeNames := []string{"nest", "cache", "barrier", "trap", "farm", "well", "tower", "market"}
			structureTypeName := "unknown"
			if int(structure.Type) < len(structureTypeNames) {
				structureTypeName = structureTypeNames[structure.Type]
			}

			metadata := map[string]interface{}{
				"structure_id":   structure.ID,
				"structure_type": structureTypeName,
				"tribe_id":       0,
				"tribe_name":     "",
			}
			if structure.Tribe != nil {
				metadata["tribe_id"] = structure.Tribe.ID
				metadata["tribe_name"] = structure.Tribe.Name
			}

			cs.EventBus.EmitSystemEvent(tick, "structure_destroyed", "civilization", "civilization_system",
				fmt.Sprintf("Structure %s destroyed at (%.1f, %.1f)", structureTypeName, structure.Position.X, structure.Position.Y),
				&structure.Position, metadata)
		}

		if structure.IsActive || structure.Health > 0 {
			activeStructures = append(activeStructures, structure)
		}
	}
	cs.Structures = activeStructures

	// Process trades
	cs.TradeSystem.ProcessTrades()

	// Generate random trade proposals
	cs.generateRandomTrades()
}

// generateRandomTrades creates occasional trade proposals between tribes
func (cs *CivilizationSystem) generateRandomTrades() {
	if len(cs.Tribes) < 2 || rand.Float64() > 0.05 {
		return // 5% chance per tick
	}

	// Pick two random tribes
	tribe1 := cs.Tribes[rand.Intn(len(cs.Tribes))]
	tribe2 := cs.Tribes[rand.Intn(len(cs.Tribes))]

	if tribe1 == tribe2 {
		return
	}

	// Generate trade based on resource needs
	offering := make(map[string]float64)
	requesting := make(map[string]float64)

	// Tribe1 offers what they have excess of
	for resource, amount := range tribe1.Resources {
		if amount > 50 {
			offering[resource] = amount * 0.2 // Offer 20% of excess
		}
	}

	// Tribe1 requests what they need
	for resource, amount := range tribe1.Resources {
		if amount < 20 {
			requesting[resource] = 30 - amount // Request to get to 30
		}
	}

	if len(offering) > 0 && len(requesting) > 0 {
		cs.TradeSystem.ProposeTrade(tribe1, tribe2, offering, requesting)
	}
}

// FormTribe creates a new tribe from compatible entities
func (cs *CivilizationSystem) FormTribe(entities []*Entity, name string, tick int) *Tribe {
	if len(entities) == 0 {
		return nil
	}

	// Find the most suitable leader (highest intelligence + cooperation)
	var leader *Entity
	bestScore := -1.0

	for _, entity := range entities {
		score := entity.GetTrait("intelligence")*0.6 + entity.GetTrait("cooperation")*0.4
		if score > bestScore {
			bestScore = score
			leader = entity
		}
	}

	if bestScore < 0.2 {
		return nil // Not civilized enough to form a tribe (reduced from 0.4)
	}

	tribe := NewTribe(cs.NextTribeID, name, leader)
	cs.NextTribeID++

	// Add all entities to the tribe
	memberIDs := make([]int, 0, len(entities))
	for _, entity := range entities {
		memberIDs = append(memberIDs, entity.ID)
		if entity != leader {
			tribe.AddMember(entity)
		}
	}

	cs.Tribes = append(cs.Tribes, tribe)

	// Emit event for tribe formation
	if cs.EventBus != nil {
		metadata := map[string]interface{}{
			"tribe_id":         tribe.ID,
			"tribe_name":       name,
			"leader_id":        leader.ID,
			"member_count":     len(entities),
			"member_ids":       memberIDs,
			"tech_level":       tribe.TechLevel,
			"avg_cooperation":  tribe.Culture["cooperation"],
			"avg_intelligence": tribe.Culture["innovation"],
		}

		cs.EventBus.EmitSystemEvent(tick, "tribe_formed", "civilization", "civilization_system",
			fmt.Sprintf("Tribe %s formed with %d members (leader: entity %d)", name, len(entities), leader.ID),
			&leader.Position, metadata)
	}

	return tribe
}
