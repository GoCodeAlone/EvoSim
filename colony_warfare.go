package main

import (
	"math"
	"math/rand"
)

// DiplomaticRelation represents the relationship between two colonies
type DiplomaticRelation int

const (
	Neutral DiplomaticRelation = iota // No special relationship
	Allied                            // Friendly cooperation
	Enemy                             // Active hostility
	Truce                             // Temporary peace
	Trading                           // Commercial relationship
	Vassal                            // Subordinate relationship
)

// String returns the string representation of DiplomaticRelation
func (dr DiplomaticRelation) String() string {
	switch dr {
	case Neutral:
		return "neutral"
	case Allied:
		return "allied"
	case Enemy:
		return "enemy"
	case Truce:
		return "truce"
	case Trading:
		return "trading"
	case Vassal:
		return "vassal"
	default:
		return "unknown"
	}
}

// ConflictType represents different types of inter-colony conflicts
type ConflictType int

const (
	BorderSkirmish ConflictType = iota // Small territorial disputes
	ResourceWar                        // Fighting over resources
	TotalWar                           // All-out conflict
	Raid                               // Quick resource grab
)

// ColonyDiplomacy tracks diplomatic relationships and history
type ColonyDiplomacy struct {
	ColonyID         int                            `json:"colony_id"`
	Relations        map[int]DiplomaticRelation     `json:"relations"`        // Other colony ID -> relation
	RelationHistory  map[int][]DiplomaticEvent     `json:"relation_history"` // History of diplomatic events
	TrustLevels      map[int]float64               `json:"trust_levels"`     // 0.0-1.0 trust with other colonies
	TradeAgreements  map[int]*TradeAgreement       `json:"trade_agreements"` // Active trade agreements
	Alliances        map[int]*Alliance             `json:"alliances"`        // Military alliances
	Conflicts        map[int]*Conflict             `json:"conflicts"`        // Active conflicts
	TerritoryBorders map[int][]TerritoryBorder     `json:"territory_borders"` // Shared borders with other colonies
	Reputation       float64                       `json:"reputation"`       // Overall reputation (-1.0 to 1.0)
}

// DiplomaticEvent records significant diplomatic interactions
type DiplomaticEvent struct {
	Tick             int                `json:"tick"`
	EventType        string             `json:"event_type"` // "alliance", "war_declaration", "trade", etc.
	OtherColonyID    int                `json:"other_colony_id"`
	Description      string             `json:"description"`
	ImpactOnTrust    float64            `json:"impact_on_trust"`
	ImpactOnReputation float64          `json:"impact_on_reputation"`
}

// TradeAgreement represents a trade relationship between colonies
type TradeAgreement struct {
	ID               int                `json:"id"`
	Colony1ID        int                `json:"colony1_id"`
	Colony2ID        int                `json:"colony2_id"`
	StartTick        int                `json:"start_tick"`
	Duration         int                `json:"duration"`
	ResourcesOffered map[string]float64 `json:"resources_offered"` // What colony1 offers
	ResourcesWanted  map[string]float64 `json:"resources_wanted"`  // What colony1 wants
	TradeRoute       []Position         `json:"trade_route"`       // Path between colonies
	TradeVolume      float64            `json:"trade_volume"`      // Amount traded per tick
	IsActive         bool               `json:"is_active"`
	LastTradeTime    int                `json:"last_trade_time"`   // Last tick when trade occurred
}

// Alliance represents a military alliance between colonies
type Alliance struct {
	ID            int     `json:"id"`
	Members       []int   `json:"members"`        // Colony IDs in the alliance
	StartTick     int     `json:"start_tick"`
	Duration      int     `json:"duration"`       // -1 for permanent
	AllianceType  string  `json:"alliance_type"`  // "defensive", "offensive", "mutual_aid"
	SharedDefense bool    `json:"shared_defense"` // Automatic defense assistance
	ResourceShare float64 `json:"resource_share"` // Percentage of resources shared
	IsActive      bool    `json:"is_active"`
}

// Conflict represents an active conflict between colonies
type Conflict struct {
	ID              int          `json:"id"`
	Attacker        int          `json:"attacker"`        // Attacking colony ID
	Defender        int          `json:"defender"`        // Defending colony ID
	ConflictType    ConflictType `json:"conflict_type"`
	StartTick       int          `json:"start_tick"`
	TurnsActive     int          `json:"turns_active"`
	CasualtyCount   int          `json:"casualty_count"`
	ResourcesLost   float64      `json:"resources_lost"`
	TerritoryClaimed []Position  `json:"territory_claimed"` // Territory taken during conflict
	Intensity       float64      `json:"intensity"`         // 0.0-1.0 conflict intensity
	WarGoal         string       `json:"war_goal"`          // "territory", "resources", "dominance"
	IsActive        bool         `json:"is_active"`
}

// TerritoryBorder represents a shared border between two colonies
type TerritoryBorder struct {
	Colony1ID     int        `json:"colony1_id"`
	Colony2ID     int        `json:"colony2_id"`
	BorderPoints  []Position `json:"border_points"`  // Points along the border
	BorderLength  float64    `json:"border_length"`  // Length of shared border
	Disputed      bool       `json:"disputed"`       // Whether border is disputed
	Fortifications int       `json:"fortifications"` // Number of defensive structures
	LastConflict  int        `json:"last_conflict"`  // Tick of last border conflict
}

// ColonyWarfareSystem manages inter-colony conflicts and diplomacy
type ColonyWarfareSystem struct {
	ColonyDiplomacies map[int]*ColonyDiplomacy `json:"colony_diplomacies"` // Colony ID -> diplomacy
	ActiveConflicts   []*Conflict              `json:"active_conflicts"`
	TradeAgreements   []*TradeAgreement        `json:"trade_agreements"`
	Alliances         []*Alliance              `json:"alliances"`
	TerritoryBorders  []*TerritoryBorder       `json:"territory_borders"`
	NextConflictID    int                      `json:"next_conflict_id"`
	NextTradeID       int                      `json:"next_trade_id"`
	NextAllianceID    int                      `json:"next_alliance_id"`
	
	// System configuration
	BorderConflictChance  float64              `json:"border_conflict_chance"`  // Chance of border conflicts
	DiplomacyUpdateRate   int                  `json:"diplomacy_update_rate"`   // Ticks between diplomacy updates
	ResourceCompetition   float64              `json:"resource_competition"`    // How much colonies compete for resources
	MaxActiveConflicts    int                  `json:"max_active_conflicts"`
}

// NewColonyWarfareSystem creates a new inter-colony warfare and diplomacy system
func NewColonyWarfareSystem() *ColonyWarfareSystem {
	return &ColonyWarfareSystem{
		ColonyDiplomacies:    make(map[int]*ColonyDiplomacy),
		ActiveConflicts:      make([]*Conflict, 0),
		TradeAgreements:      make([]*TradeAgreement, 0),
		Alliances:            make([]*Alliance, 0),
		TerritoryBorders:     make([]*TerritoryBorder, 0),
		NextConflictID:       1,
		NextTradeID:          1,
		NextAllianceID:       1,
		BorderConflictChance: 0.05, // 5% chance per tick
		DiplomacyUpdateRate:  50,   // Update every 50 ticks
		ResourceCompetition:  0.7,  // Moderate competition
		MaxActiveConflicts:   10,   // Maximum simultaneous conflicts
	}
}

// RegisterColony adds a new colony to the diplomatic system
func (cws *ColonyWarfareSystem) RegisterColony(colony *CasteColony) {
	if _, exists := cws.ColonyDiplomacies[colony.ID]; exists {
		return // Already registered
	}

	diplomacy := &ColonyDiplomacy{
		ColonyID:         colony.ID,
		Relations:        make(map[int]DiplomaticRelation),
		RelationHistory:  make(map[int][]DiplomaticEvent),
		TrustLevels:      make(map[int]float64),
		TradeAgreements:  make(map[int]*TradeAgreement),
		Alliances:        make(map[int]*Alliance),
		Conflicts:        make(map[int]*Conflict),
		TerritoryBorders: make(map[int][]TerritoryBorder),
		Reputation:       0.0, // Start neutral
	}

	// Initialize relations with existing colonies as neutral
	for existingColonyID := range cws.ColonyDiplomacies {
		diplomacy.Relations[existingColonyID] = Neutral
		diplomacy.TrustLevels[existingColonyID] = 0.5 // Neutral trust
		
		// Add reverse relation
		cws.ColonyDiplomacies[existingColonyID].Relations[colony.ID] = Neutral
		cws.ColonyDiplomacies[existingColonyID].TrustLevels[colony.ID] = 0.5
	}

	cws.ColonyDiplomacies[colony.ID] = diplomacy
}

// UpdateTerritoryBorders calculates shared borders between colonies
func (cws *ColonyWarfareSystem) UpdateTerritoryBorders(colonies []*CasteColony) {
	cws.TerritoryBorders = make([]*TerritoryBorder, 0)

	for i, colony1 := range colonies {
		for j, colony2 := range colonies {
			if i >= j { // Only check each pair once
				continue
			}

			borderPoints := cws.calculateSharedBorder(colony1.Territory, colony2.Territory)
			if len(borderPoints) > 0 {
				border := &TerritoryBorder{
					Colony1ID:     colony1.ID,
					Colony2ID:     colony2.ID,
					BorderPoints:  borderPoints,
					BorderLength:  cws.calculateBorderLength(borderPoints),
					Disputed:      false,
					Fortifications: 0,
					LastConflict:  0,
				}
				cws.TerritoryBorders = append(cws.TerritoryBorders, border)

				// Update colony diplomacies with border information
				if diplomacy1, exists := cws.ColonyDiplomacies[colony1.ID]; exists {
					diplomacy1.TerritoryBorders[colony2.ID] = []TerritoryBorder{*border}
				}
				if diplomacy2, exists := cws.ColonyDiplomacies[colony2.ID]; exists {
					diplomacy2.TerritoryBorders[colony1.ID] = []TerritoryBorder{*border}
				}
			}
		}
	}
}

// calculateSharedBorder finds border points between two territories
func (cws *ColonyWarfareSystem) calculateSharedBorder(territory1, territory2 []Position) []Position {
	borderPoints := make([]Position, 0)
	borderDistance := 3.0 // Maximum distance to consider adjacent

	for _, pos1 := range territory1 {
		for _, pos2 := range territory2 {
			distance := math.Sqrt(math.Pow(pos1.X-pos2.X, 2) + math.Pow(pos1.Y-pos2.Y, 2))
			if distance <= borderDistance {
				// These points are close enough to form a border
				borderPoint := Position{
					X: (pos1.X + pos2.X) / 2,
					Y: (pos1.Y + pos2.Y) / 2,
				}
				borderPoints = append(borderPoints, borderPoint)
			}
		}
	}

	return borderPoints
}

// calculateBorderLength calculates the total length of a border
func (cws *ColonyWarfareSystem) calculateBorderLength(borderPoints []Position) float64 {
	if len(borderPoints) < 2 {
		return 0.0
	}

	totalLength := 0.0
	for i := 1; i < len(borderPoints); i++ {
		distance := math.Sqrt(
			math.Pow(borderPoints[i].X-borderPoints[i-1].X, 2) +
			math.Pow(borderPoints[i].Y-borderPoints[i-1].Y, 2),
		)
		totalLength += distance
	}

	return totalLength
}

// CheckForConflicts evaluates potential conflicts between colonies
func (cws *ColonyWarfareSystem) CheckForConflicts(colonies []*CasteColony, tick int) {
	if len(cws.ActiveConflicts) >= cws.MaxActiveConflicts {
		return // Too many active conflicts
	}

	for _, border := range cws.TerritoryBorders {
		if rand.Float64() < cws.BorderConflictChance {
			colony1 := cws.findColonyByID(colonies, border.Colony1ID)
			colony2 := cws.findColonyByID(colonies, border.Colony2ID)

			if colony1 != nil && colony2 != nil {
				// Check if conditions are right for conflict
				if cws.shouldStartConflict(colony1, colony2, border, tick) {
					cws.StartConflict(colony1, colony2, BorderSkirmish, tick)
				}
			}
		}
	}
}

// shouldStartConflict determines if two colonies should enter conflict
func (cws *ColonyWarfareSystem) shouldStartConflict(colony1, colony2 *CasteColony, border *TerritoryBorder, tick int) bool {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]

	// Don't fight allies
	if diplomacy1.Relations[colony2.ID] == Allied || diplomacy2.Relations[colony1.ID] == Allied {
		return false
	}

	// More likely if already enemies
	if diplomacy1.Relations[colony2.ID] == Enemy {
		return rand.Float64() < 0.3 // 30% chance
	}

	// Resource competition increases conflict chance
	resourcePressure := cws.calculateResourcePressure(colony1, colony2)
	conflictChance := resourcePressure * cws.ResourceCompetition * 0.1

	// Border disputes increase chance
	if border.Disputed {
		conflictChance *= 2.0
	}

	// Recent conflicts reduce chance (war weariness)
	if tick-border.LastConflict < 100 {
		conflictChance *= 0.5
	}

	return rand.Float64() < conflictChance
}

// calculateResourcePressure determines how much colonies compete for resources
func (cws *ColonyWarfareSystem) calculateResourcePressure(colony1, colony2 *CasteColony) float64 {
	// Calculate based on colony sizes and proximity
	distance := math.Sqrt(
		math.Pow(colony1.NestLocation.X-colony2.NestLocation.X, 2) +
		math.Pow(colony1.NestLocation.Y-colony2.NestLocation.Y, 2),
	)

	// Closer colonies have more resource pressure
	proximityPressure := math.Max(0, 1.0-distance/50.0)

	// Larger colonies create more pressure
	sizePressure := float64(colony1.ColonySize+colony2.ColonySize) / 200.0

	return math.Min(1.0, proximityPressure+sizePressure)
}

// StartConflict initiates a conflict between two colonies
func (cws *ColonyWarfareSystem) StartConflict(attacker, defender *CasteColony, conflictType ConflictType, tick int) *Conflict {
	conflict := &Conflict{
		ID:              cws.NextConflictID,
		Attacker:        attacker.ID,
		Defender:        defender.ID,
		ConflictType:    conflictType,
		StartTick:       tick,
		TurnsActive:     0,
		CasualtyCount:   0,
		ResourcesLost:   0,
		TerritoryClaimed: make([]Position, 0),
		Intensity:       cws.getInitialConflictIntensity(conflictType),
		WarGoal:         cws.determineWarGoal(attacker, defender),
		IsActive:        true,
	}

	cws.NextConflictID++
	cws.ActiveConflicts = append(cws.ActiveConflicts, conflict)

	// Update diplomatic relations
	attackerDiplomacy := cws.ColonyDiplomacies[attacker.ID]
	defenderDiplomacy := cws.ColonyDiplomacies[defender.ID]

	attackerDiplomacy.Relations[defender.ID] = Enemy
	defenderDiplomacy.Relations[attacker.ID] = Enemy
	attackerDiplomacy.Conflicts[defender.ID] = conflict
	defenderDiplomacy.Conflicts[attacker.ID] = conflict

	// Reduce trust
	attackerDiplomacy.TrustLevels[defender.ID] = math.Max(0, attackerDiplomacy.TrustLevels[defender.ID]-0.3)
	defenderDiplomacy.TrustLevels[attacker.ID] = math.Max(0, defenderDiplomacy.TrustLevels[attacker.ID]-0.3)

	// Record diplomatic event
	event := DiplomaticEvent{
		Tick:               tick,
		EventType:          "war_declaration",
		OtherColonyID:      defender.ID,
		Description:        "War declared",
		ImpactOnTrust:      -0.3,
		ImpactOnReputation: -0.1,
	}
	attackerDiplomacy.RelationHistory[defender.ID] = append(attackerDiplomacy.RelationHistory[defender.ID], event)

	return conflict
}

// getInitialConflictIntensity determines starting intensity based on conflict type
func (cws *ColonyWarfareSystem) getInitialConflictIntensity(conflictType ConflictType) float64 {
	switch conflictType {
	case BorderSkirmish:
		return 0.2 + rand.Float64()*0.3 // 0.2-0.5
	case ResourceWar:
		return 0.4 + rand.Float64()*0.4 // 0.4-0.8
	case TotalWar:
		return 0.7 + rand.Float64()*0.3 // 0.7-1.0
	case Raid:
		return 0.1 + rand.Float64()*0.2 // 0.1-0.3
	default:
		return 0.3
	}
}

// determineWarGoal determines what the attacker wants from the conflict
func (cws *ColonyWarfareSystem) determineWarGoal(attacker, defender *CasteColony) string {
	// Based on colony characteristics and needs
	if attacker.ColonySize > int(float64(defender.ColonySize)*1.5) {
		return "dominance" // Larger colony seeks dominance
	}

	if len(attacker.Territory) < 5 {
		return "territory" // Small territory seeks expansion
	}

	return "resources" // Default to resource competition
}

// ProcessConflicts updates all active conflicts
func (cws *ColonyWarfareSystem) ProcessConflicts(colonies []*CasteColony, tick int) {
	activeConflicts := make([]*Conflict, 0)

	for _, conflict := range cws.ActiveConflicts {
		if cws.updateConflict(conflict, colonies, tick) {
			activeConflicts = append(activeConflicts, conflict)
		}
	}

	cws.ActiveConflicts = activeConflicts
}

// updateConflict processes one turn of a conflict
func (cws *ColonyWarfareSystem) updateConflict(conflict *Conflict, colonies []*CasteColony, tick int) bool {
	attacker := cws.findColonyByID(colonies, conflict.Attacker)
	defender := cws.findColonyByID(colonies, conflict.Defender)

	if attacker == nil || defender == nil {
		return false // One colony no longer exists
	}

	conflict.TurnsActive++

	// Calculate battle results
	attackerStrength := cws.calculateMilitaryStrength(attacker)
	defenderStrength := cws.calculateMilitaryStrength(defender)

	// Apply conflict
	casualties := cws.resolveBattle(attacker, defender, attackerStrength, defenderStrength, conflict)
	conflict.CasualtyCount += casualties

	// Update conflict intensity based on progress
	if conflict.TurnsActive > 20 {
		conflict.Intensity *= 0.95 // Gradual decline over time
	}

	// Check for conflict resolution
	if cws.shouldEndConflict(conflict, attacker, defender) {
		cws.resolveConflict(conflict, attacker, defender, tick)
		return false
	}

	return true
}

// calculateMilitaryStrength determines a colony's fighting capability
func (cws *ColonyWarfareSystem) calculateMilitaryStrength(colony *CasteColony) float64 {
	soldiers := float64(colony.CasteDistribution[Soldier])
	workers := float64(colony.CasteDistribution[Worker]) * 0.3 // Workers can fight but less effectively
	
	// Base military strength
	strength := soldiers*2.0 + workers
	
	// Multiply by average fitness and size
	strength *= colony.ColonyFitness
	
	// Territory size provides defensive bonus
	territoryBonus := math.Min(2.0, float64(len(colony.Territory))/10.0)
	strength *= (1.0 + territoryBonus)
	
	return strength
}

// resolveBattle processes combat between two colonies
func (cws *ColonyWarfareSystem) resolveBattle(attacker, defender *CasteColony, attackerStrength, defenderStrength float64, conflict *Conflict) int {
	// Calculate battle outcome
	strengthRatio := attackerStrength / (attackerStrength + defenderStrength)
	
	// Add some randomness
	battleRoll := rand.Float64()
	
	casualties := 0
	resourcesLost := 0.0
	
	if strengthRatio > 0.6 && battleRoll > 0.3 { // Attacker victory
		// Defender loses
		defenderLosses := int(float64(defender.ColonySize) * 0.1 * conflict.Intensity)
		casualties = defenderLosses
		resourcesLost = defenderStrength * 0.2
		
		// Attacker gains territory
		if len(defender.Territory) > 1 && conflict.WarGoal == "territory" {
			claimedTerritory := defender.Territory[rand.Intn(len(defender.Territory))]
			conflict.TerritoryClaimed = append(conflict.TerritoryClaimed, claimedTerritory)
			
			// Remove from defender territory
			newTerritory := make([]Position, 0)
			for _, pos := range defender.Territory {
				if pos.X != claimedTerritory.X || pos.Y != claimedTerritory.Y {
					newTerritory = append(newTerritory, pos)
				}
			}
			defender.Territory = newTerritory
		}
		
	} else if strengthRatio < 0.4 && battleRoll < 0.7 { // Defender victory
		// Attacker loses
		attackerLosses := int(float64(attacker.ColonySize) * 0.1 * conflict.Intensity)
		casualties = attackerLosses
		resourcesLost = attackerStrength * 0.2
		
	} else { // Stalemate
		// Both sides lose a little
		mutualLosses := int(float64(attacker.ColonySize+defender.ColonySize) * 0.05 * conflict.Intensity)
		casualties = mutualLosses
		resourcesLost = (attackerStrength + defenderStrength) * 0.1
	}
	
	// Apply losses (simplified - just reduce colony size)
	attacker.ColonySize = int(math.Max(1, float64(attacker.ColonySize) - float64(casualties)*0.3))
	defender.ColonySize = int(math.Max(1, float64(defender.ColonySize) - float64(casualties)*0.7))
	
	conflict.ResourcesLost += resourcesLost
	
	return casualties
}

// shouldEndConflict determines if a conflict should end
func (cws *ColonyWarfareSystem) shouldEndConflict(conflict *Conflict, attacker, defender *CasteColony) bool {
	// End if one colony is nearly destroyed
	if attacker.ColonySize < 5 || defender.ColonySize < 5 {
		return true
	}
	
	// End if conflict has gone on too long
	maxDuration := 100
	if conflict.ConflictType == Raid {
		maxDuration = 20
	} else if conflict.ConflictType == TotalWar {
		maxDuration = 200
	}
	
	if conflict.TurnsActive > maxDuration {
		return true
	}
	
	// End if intensity is very low
	if conflict.Intensity < 0.1 {
		return true
	}
	
	// Random chance to end (war weariness)
	if conflict.TurnsActive > 30 && rand.Float64() < 0.05 {
		return true
	}
	
	return false
}

// resolveConflict finalizes a conflict and sets post-war relations
func (cws *ColonyWarfareSystem) resolveConflict(conflict *Conflict, attacker, defender *CasteColony, tick int) {
	attackerDiplomacy := cws.ColonyDiplomacies[attacker.ID]
	defenderDiplomacy := cws.ColonyDiplomacies[defender.ID]
	
	// Determine victor based on goals achieved
	attackerVictory := cws.evaluateWarOutcome(conflict, attacker, defender)
	
	var winner, loser *CasteColony
	var winnerDiplomacy, loserDiplomacy *ColonyDiplomacy
	
	if attackerVictory {
		winner = attacker
		loser = defender
		winnerDiplomacy = attackerDiplomacy
		loserDiplomacy = defenderDiplomacy
	} else {
		winner = defender
		loser = attacker
		winnerDiplomacy = defenderDiplomacy
		loserDiplomacy = attackerDiplomacy
	}
	
	// Apply post-war effects
	winnerDiplomacy.Reputation += 0.1
	loserDiplomacy.Reputation -= 0.1
	
	// Set post-war relations
	if conflict.CasualtyCount > 20 {
		// High casualties lead to lasting enmity
		winnerDiplomacy.Relations[loser.ID] = Enemy
		loserDiplomacy.Relations[winner.ID] = Enemy
		winnerDiplomacy.TrustLevels[loser.ID] = 0.1
		loserDiplomacy.TrustLevels[winner.ID] = 0.1
	} else {
		// Low casualties may lead to truce
		winnerDiplomacy.Relations[loser.ID] = Truce
		loserDiplomacy.Relations[winner.ID] = Truce
		winnerDiplomacy.TrustLevels[loser.ID] = 0.3
		loserDiplomacy.TrustLevels[winner.ID] = 0.3
	}
	
	// Transfer claimed territory
	for _, territory := range conflict.TerritoryClaimed {
		winner.Territory = append(winner.Territory, territory)
	}
	
	// Record peace event
	peaceEvent := DiplomaticEvent{
		Tick:               tick,
		EventType:          "peace_treaty",
		OtherColonyID:      loser.ID,
		Description:        "Conflict resolved",
		ImpactOnTrust:      0.1,
		ImpactOnReputation: 0.05,
	}
	winnerDiplomacy.RelationHistory[loser.ID] = append(winnerDiplomacy.RelationHistory[loser.ID], peaceEvent)
	
	// Remove from active conflicts
	delete(attackerDiplomacy.Conflicts, defender.ID)
	delete(defenderDiplomacy.Conflicts, attacker.ID)
	
	conflict.IsActive = false
}

// evaluateWarOutcome determines if attacker achieved their war goals
func (cws *ColonyWarfareSystem) evaluateWarOutcome(conflict *Conflict, attacker, defender *CasteColony) bool {
	switch conflict.WarGoal {
	case "territory":
		return len(conflict.TerritoryClaimed) > 0
	case "resources":
		return conflict.ResourcesLost > 0 && attacker.ColonySize >= defender.ColonySize
	case "dominance":
		return attacker.ColonySize > int(float64(defender.ColonySize)*1.2)
	default:
		return rand.Float64() < 0.5 // Random if unclear
	}
}

// findColonyByID finds a colony by its ID
func (cws *ColonyWarfareSystem) findColonyByID(colonies []*CasteColony, id int) *CasteColony {
	for _, colony := range colonies {
		if colony.ID == id {
			return colony
		}
	}
	return nil
}

// AttemptDiplomacy tries to establish or improve relations between colonies
func (cws *ColonyWarfareSystem) AttemptDiplomacy(colonies []*CasteColony, tick int) {
	if tick%cws.DiplomacyUpdateRate != 0 {
		return // Not time for diplomacy update
	}

	for _, colony1 := range colonies {
		for _, colony2 := range colonies {
			if colony1.ID >= colony2.ID {
				continue // Only process each pair once
			}

			if cws.shouldAttemptDiplomacy(colony1, colony2) {
				cws.processDiplomaticInteraction(colony1, colony2, tick)
			}
		}
	}
}

// shouldAttemptDiplomacy determines if two colonies should engage in diplomacy
func (cws *ColonyWarfareSystem) shouldAttemptDiplomacy(colony1, colony2 *CasteColony) bool {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	
	// Don't negotiate during active conflicts
	if diplomacy1.Relations[colony2.ID] == Enemy {
		if _, hasConflict := diplomacy1.Conflicts[colony2.ID]; hasConflict {
			return false
		}
	}
	
	// More likely if neutral or truce
	relation := diplomacy1.Relations[colony2.ID]
	return relation == Neutral || relation == Truce
}

// processDiplomaticInteraction handles diplomatic negotiations
func (cws *ColonyWarfareSystem) processDiplomaticInteraction(colony1, colony2 *CasteColony, tick int) {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
	
	// Calculate diplomatic success chance
	trust := (diplomacy1.TrustLevels[colony2.ID] + diplomacy2.TrustLevels[colony1.ID]) / 2.0
	reputation := (diplomacy1.Reputation + diplomacy2.Reputation) / 2.0
	proximity := cws.calculateProximity(colony1, colony2)
	
	successChance := (trust + reputation + proximity) / 3.0
	
	if rand.Float64() < successChance {
		// Successful diplomacy - improve relations
		cws.improveRelations(colony1, colony2, tick)
	}
}

// calculateProximity determines how close two colonies are (affects diplomacy)
func (cws *ColonyWarfareSystem) calculateProximity(colony1, colony2 *CasteColony) float64 {
	distance := math.Sqrt(
		math.Pow(colony1.NestLocation.X-colony2.NestLocation.X, 2) +
		math.Pow(colony1.NestLocation.Y-colony2.NestLocation.Y, 2),
	)
	
	// Closer colonies have higher proximity score
	return math.Max(0, 1.0-distance/100.0)
}

// improveRelations enhances the relationship between two colonies
func (cws *ColonyWarfareSystem) improveRelations(colony1, colony2 *CasteColony, tick int) {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
	
	// Increase trust
	diplomacy1.TrustLevels[colony2.ID] = math.Min(1.0, diplomacy1.TrustLevels[colony2.ID]+0.1)
	diplomacy2.TrustLevels[colony1.ID] = math.Min(1.0, diplomacy2.TrustLevels[colony1.ID]+0.1)
	
	// Improve relations
	currentRelation := diplomacy1.Relations[colony2.ID]
	newRelation := currentRelation
	
	switch currentRelation {
	case Enemy:
		if diplomacy1.TrustLevels[colony2.ID] > 0.4 {
			newRelation = Truce
		}
	case Truce:
		if diplomacy1.TrustLevels[colony2.ID] > 0.6 {
			newRelation = Neutral
		}
	case Neutral:
		if diplomacy1.TrustLevels[colony2.ID] > 0.8 {
			newRelation = Allied
		}
	}
	
	if newRelation != currentRelation {
		diplomacy1.Relations[colony2.ID] = newRelation
		diplomacy2.Relations[colony1.ID] = newRelation
		
		// Record diplomatic event
		event := DiplomaticEvent{
			Tick:               tick,
			EventType:          "relation_improvement",
			OtherColonyID:      colony2.ID,
			Description:        "Relations improved to " + newRelation.String(),
			ImpactOnTrust:      0.1,
			ImpactOnReputation: 0.05,
		}
		diplomacy1.RelationHistory[colony2.ID] = append(diplomacy1.RelationHistory[colony2.ID], event)
	}
}

// GetWarfareStats returns statistics about the warfare system
func (cws *ColonyWarfareSystem) GetWarfareStats() map[string]interface{} {
	totalColonies := len(cws.ColonyDiplomacies)
	activeConflicts := len(cws.ActiveConflicts)
	totalAlliances := len(cws.Alliances)
	activeTradeAgreements := 0
	
	for _, agreement := range cws.TradeAgreements {
		if agreement.IsActive {
			activeTradeAgreements++
		}
	}
	
	// Calculate average relations
	relationCounts := make(map[DiplomaticRelation]int)
	totalRelations := 0
	
	for _, diplomacy := range cws.ColonyDiplomacies {
		for _, relation := range diplomacy.Relations {
			relationCounts[relation]++
			totalRelations++
		}
	}
	
	return map[string]interface{}{
		"total_colonies":          totalColonies,
		"active_conflicts":        activeConflicts,
		"total_alliances":         totalAlliances,
		"active_trade_agreements": activeTradeAgreements,
		"neutral_relations":       relationCounts[Neutral],
		"allied_relations":        relationCounts[Allied],
		"enemy_relations":         relationCounts[Enemy],
		"truce_relations":         relationCounts[Truce],
		"total_relations":         totalRelations,
		"border_conflicts":        cws.BorderConflictChance,
		"resource_competition":    cws.ResourceCompetition,
	}
}

// Update processes the warfare system for one tick
func (cws *ColonyWarfareSystem) Update(colonies []*CasteColony, tick int) {
	// Register new colonies
	for _, colony := range colonies {
		cws.RegisterColony(colony)
	}
	
	// Update territory borders
	cws.UpdateTerritoryBorders(colonies)
	
	// Check for new conflicts
	cws.CheckForConflicts(colonies, tick)
	
	// Process active conflicts
	cws.ProcessConflicts(colonies, tick)
	
	// Attempt diplomatic interactions
	cws.AttemptDiplomacy(colonies, tick)
	
	// Process active trade agreements
	cws.ProcessTradeAgreements(colonies, tick)
	
	// Attempt to establish new beneficial trade agreements
	cws.AttemptAutomaticTrading(colonies, tick)
	
	// Process alliance benefits and military cooperation
	cws.ProcessAlliances(colonies, tick)
	
	// Attempt to form new beneficial alliances
	cws.AttemptAllianceFormation(colonies, tick)
}

// ProcessTradeAgreements executes resource trades between colonies
func (cws *ColonyWarfareSystem) ProcessTradeAgreements(colonies []*CasteColony, tick int) {
	activeAgreements := make([]*TradeAgreement, 0)
	
	for _, agreement := range cws.TradeAgreements {
		if !agreement.IsActive {
			continue
		}
		
		// Check if agreement has expired
		if agreement.Duration > 0 && tick-agreement.StartTick > agreement.Duration {
			agreement.IsActive = false
			continue
		}
		
		colony1 := cws.findColonyByID(colonies, agreement.Colony1ID)
		colony2 := cws.findColonyByID(colonies, agreement.Colony2ID)
		
		if colony1 == nil || colony2 == nil {
			agreement.IsActive = false
			continue
		}
		
		// Execute trade if enough time has passed
		if tick-agreement.LastTradeTime >= 10 { // Trade every 10 ticks
			if cws.executeTrade(agreement, colony1, colony2, tick) {
				agreement.LastTradeTime = tick
				activeAgreements = append(activeAgreements, agreement)
			} else {
				// Trade failed, possibly suspend agreement
				agreement.IsActive = false
			}
		} else {
			activeAgreements = append(activeAgreements, agreement)
		}
	}
	
	cws.TradeAgreements = activeAgreements
}

// executeTrade performs the actual resource exchange
func (cws *ColonyWarfareSystem) executeTrade(agreement *TradeAgreement, colony1, colony2 *CasteColony, tick int) bool {
	// Check if both colonies can fulfill their trade obligations
	canTrade := true
	
	// Check if colony1 can provide what it offers (considering strategic reserves)
	for resourceType, amount := range agreement.ResourcesOffered {
		if !colony1.CanAffordResourceForTrade(resourceType, amount*agreement.TradeVolume) {
			canTrade = false
			break
		}
	}
	
	// Check if colony2 can provide what colony1 wants (considering strategic reserves)
	for resourceType, amount := range agreement.ResourcesWanted {
		if !colony2.CanAffordResourceForTrade(resourceType, amount*agreement.TradeVolume) {
			canTrade = false
			break
		}
	}
	
	if !canTrade {
		return false
	}
	
	// Execute the trade with route efficiency and relationship bonuses
	efficiency := cws.calculateTradeRouteEfficiency(colony1, colony2, agreement)
	trustBonus := cws.calculateTrustBonus(colony1, colony2)
	relationshipMultiplier := cws.calculateRelationshipMultiplier(colony1, colony2)
	
	actualVolume := agreement.TradeVolume * efficiency * trustBonus * relationshipMultiplier
	
	// Colony1 gives what it offers, gets what it wants
	for resourceType, amount := range agreement.ResourcesOffered {
		if colony1.ConsumeResource(resourceType, amount*actualVolume) {
			colony2.AddResource(resourceType, amount*actualVolume)
		}
	}
	
	for resourceType, amount := range agreement.ResourcesWanted {
		if colony2.ConsumeResource(resourceType, amount*actualVolume) {
			colony1.AddResource(resourceType, amount*actualVolume)
		}
	}
	
	// Update diplomatic relations (trading improves trust)
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
	
	if diplomacy1 != nil && diplomacy2 != nil {
		trustIncrease := 0.02 * efficiency // Small trust increase from successful trade
		diplomacy1.TrustLevels[colony2.ID] = math.Min(1.0, diplomacy1.TrustLevels[colony2.ID]+trustIncrease)
		diplomacy2.TrustLevels[colony1.ID] = math.Min(1.0, diplomacy2.TrustLevels[colony1.ID]+trustIncrease)
		
		// Trading relationships tend toward "Trading" diplomatic status
		if diplomacy1.Relations[colony2.ID] == Neutral && diplomacy1.TrustLevels[colony2.ID] > 0.6 {
			diplomacy1.Relations[colony2.ID] = Trading
			diplomacy2.Relations[colony1.ID] = Trading
		}
	}
	
	return true
}

// calculateTradeRouteEfficiency determines how efficient a trade route is
func (cws *ColonyWarfareSystem) calculateTradeRouteEfficiency(colony1, colony2 *CasteColony, agreement *TradeAgreement) float64 {
	// Base efficiency
	efficiency := 0.8
	
	// Distance penalty
	distance := math.Sqrt(
		math.Pow(colony1.NestLocation.X-colony2.NestLocation.X, 2) +
		math.Pow(colony1.NestLocation.Y-colony2.NestLocation.Y, 2),
	)
	distancePenalty := math.Min(0.3, distance/200.0) // Max 30% penalty for distance
	efficiency -= distancePenalty
	
	// Diplomatic relation bonus
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 != nil {
		relation := diplomacy1.Relations[colony2.ID]
		switch relation {
		case Allied:
			efficiency += 0.2
		case Trading:
			efficiency += 0.1
		case Enemy:
			efficiency -= 0.5 // Very inefficient to trade with enemies
		case Truce:
			efficiency -= 0.1
		}
		
		// Trust level affects efficiency
		trust := diplomacy1.TrustLevels[colony2.ID]
		efficiency += (trust - 0.5) * 0.2 // ±0.1 based on trust
	}
	
	// Security affects efficiency (conflicts reduce efficiency)
	for _, conflict := range cws.ActiveConflicts {
		if (conflict.Attacker == colony1.ID || conflict.Defender == colony1.ID) ||
		   (conflict.Attacker == colony2.ID || conflict.Defender == colony2.ID) {
			efficiency -= 0.3 * conflict.Intensity // Conflicts disrupt trade
		}
	}
	
	return math.Max(0.1, math.Min(1.0, efficiency))
}

// CreateTradeAgreement establishes a new trade agreement between colonies
func (cws *ColonyWarfareSystem) CreateTradeAgreement(colony1, colony2 *CasteColony, 
	offeredResources, wantedResources map[string]float64, duration int, tick int) *TradeAgreement {
	
	// Check diplomatic compatibility
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 == nil {
		return nil
	}
	
	relation := diplomacy1.Relations[colony2.ID]
	if relation == Enemy {
		return nil // Cannot trade with enemies
	}
	
	agreement := &TradeAgreement{
		ID:               cws.NextTradeID,
		Colony1ID:        colony1.ID,
		Colony2ID:        colony2.ID,
		StartTick:        tick,
		Duration:         duration,
		ResourcesOffered: make(map[string]float64),
		ResourcesWanted:  make(map[string]float64),
		TradeVolume:      1.0, // Default trade volume
		IsActive:         true,
		LastTradeTime:    tick,
	}
	
	// Copy resource maps
	for resource, amount := range offeredResources {
		agreement.ResourcesOffered[resource] = amount
	}
	for resource, amount := range wantedResources {
		agreement.ResourcesWanted[resource] = amount
	}
	
	cws.NextTradeID++
	cws.TradeAgreements = append(cws.TradeAgreements, agreement)
	
	// Update diplomatic tracking
	diplomacy1.TradeAgreements[colony2.ID] = agreement
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
	if diplomacy2 != nil {
		diplomacy2.TradeAgreements[colony1.ID] = agreement
	}
	
	// Record diplomatic event
	event := DiplomaticEvent{
		Tick:               tick,
		EventType:          "trade_agreement",
		OtherColonyID:      colony2.ID,
		Description:        "Trade agreement established",
		ImpactOnTrust:      0.1,
		ImpactOnReputation: 0.05,
	}
	diplomacy1.RelationHistory[colony2.ID] = append(diplomacy1.RelationHistory[colony2.ID], event)
	
	return agreement
}

// AttemptAutomaticTrading tries to establish beneficial trade agreements
func (cws *ColonyWarfareSystem) AttemptAutomaticTrading(colonies []*CasteColony, tick int) {
	if tick%100 != 0 { // Check for new trades every 100 ticks
		return
	}
	
	for i, colony1 := range colonies {
		for j, colony2 := range colonies {
			if i >= j {
				continue
			}
			
			// Check if they should consider trading
			if cws.shouldAttemptTrading(colony1, colony2) {
				cws.evaluateAndCreateTrade(colony1, colony2, tick)
			}
		}
	}
}

// shouldAttemptTrading determines if two colonies should consider trading
func (cws *ColonyWarfareSystem) shouldAttemptTrading(colony1, colony2 *CasteColony) bool {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 == nil {
		return false
	}
	
	relation := diplomacy1.Relations[colony2.ID]
	
	// Don't trade with enemies or if already have trade agreement
	if relation == Enemy {
		return false
	}
	
	if _, hasAgreement := diplomacy1.TradeAgreements[colony2.ID]; hasAgreement {
		return false
	}
	
	// Must have sufficient trust
	trust := diplomacy1.TrustLevels[colony2.ID]
	return trust > 0.4
}

// evaluateAndCreateTrade analyzes mutual benefits and creates trade if beneficial
func (cws *ColonyWarfareSystem) evaluateAndCreateTrade(colony1, colony2 *CasteColony, tick int) {
	// Find resources where colony1 has surplus and colony2 has need
	offeredResources := make(map[string]float64)
	wantedResources := make(map[string]float64)
	
	resourceTypes := []string{"food", "biomass", "energy", "materials"}
	
	for _, resourceType := range resourceTypes {
		surplus1 := colony1.GetResourceSurplus(resourceType)
		need2 := colony2.GetResourceNeed(resourceType)
		
		surplus2 := colony2.GetResourceSurplus(resourceType)
		need1 := colony1.GetResourceNeed(resourceType)
		
		// Colony1 can offer what colony2 needs
		if surplus1 > 10 && need2 > 10 {
			tradeAmount := math.Min(surplus1*0.3, need2*0.5) // Conservative trade amount
			offeredResources[resourceType] = tradeAmount
		}
		
		// Colony1 wants what colony2 can offer
		if surplus2 > 10 && need1 > 10 {
			tradeAmount := math.Min(surplus2*0.3, need1*0.5) // Conservative trade amount
			wantedResources[resourceType] = tradeAmount
		}
	}
	
	// Only create trade if both sides have something to offer and want
	if len(offeredResources) > 0 && len(wantedResources) > 0 {
		duration := 500 + rand.Intn(1000) // 500-1500 tick duration
		agreement := cws.CreateTradeAgreement(colony1, colony2, offeredResources, wantedResources, duration, tick)
		
		if agreement != nil {
			// Successful trade agreement
			diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
			diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
			
			// Improve relations
			if diplomacy1.Relations[colony2.ID] == Neutral {
				diplomacy1.Relations[colony2.ID] = Trading
			}
			if diplomacy2.Relations[colony1.ID] == Neutral {
				diplomacy2.Relations[colony1.ID] = Trading
			}
		}
	}
}

// Alliance Enhancement - Military Cooperation Methods

// ProcessAlliances handles alliance benefits and coordination
func (cws *ColonyWarfareSystem) ProcessAlliances(colonies []*CasteColony, tick int) {
	for _, alliance := range cws.Alliances {
		if !alliance.IsActive {
			continue
		}
		
		// Check if alliance has expired
		if alliance.Duration > 0 && tick-alliance.StartTick > alliance.Duration {
			alliance.IsActive = false
			continue
		}
		
		// Process resource sharing if enabled
		if alliance.ResourceShare > 0 {
			cws.processAllianceResourceSharing(alliance, colonies)
		}
		
		// Check for shared defense opportunities
		if alliance.SharedDefense {
			cws.processSharedDefense(alliance, colonies, tick)
		}
		
		// Coordinate joint military actions
		cws.coordinateJointOperations(alliance, colonies, tick)
	}
}

// processAllianceResourceSharing redistributes resources among alliance members
func (cws *ColonyWarfareSystem) processAllianceResourceSharing(alliance *Alliance, colonies []*CasteColony) {
	memberColonies := make([]*CasteColony, 0)
	
	// Gather alliance member colonies
	for _, memberID := range alliance.Members {
		colony := cws.findColonyByID(colonies, memberID)
		if colony != nil {
			memberColonies = append(memberColonies, colony)
		}
	}
	
	if len(memberColonies) < 2 {
		return
	}
	
	resourceTypes := []string{"food", "biomass", "energy", "materials"}
	
	for _, resourceType := range resourceTypes {
		// Calculate total resources and needs
		totalSurplus := 0.0
		totalNeed := 0.0
		surplusColonies := make([]*CasteColony, 0)
		needyColonies := make([]*CasteColony, 0)
		
		for _, colony := range memberColonies {
			surplus := colony.GetResourceSurplus(resourceType)
			need := colony.GetResourceNeed(resourceType)
			
			if surplus > 20 {
				totalSurplus += surplus
				surplusColonies = append(surplusColonies, colony)
			}
			if need > 20 {
				totalNeed += need
				needyColonies = append(needyColonies, colony)
			}
		}
		
		// Redistribute if there's surplus and need
		if totalSurplus > 0 && totalNeed > 0 && len(surplusColonies) > 0 && len(needyColonies) > 0 {
			redistributionAmount := math.Min(totalSurplus, totalNeed) * alliance.ResourceShare
			
			// Take from surplus colonies proportionally
			for _, colony := range surplusColonies {
				surplus := colony.GetResourceSurplus(resourceType)
				proportion := surplus / totalSurplus
				contribution := redistributionAmount * proportion
				colony.ConsumeResource(resourceType, contribution)
			}
			
			// Give to needy colonies proportionally
			for _, colony := range needyColonies {
				need := colony.GetResourceNeed(resourceType)
				proportion := need / totalNeed
				allocation := redistributionAmount * proportion
				colony.AddResource(resourceType, allocation)
			}
		}
	}
}

// processSharedDefense coordinates defensive assistance among allies
func (cws *ColonyWarfareSystem) processSharedDefense(alliance *Alliance, colonies []*CasteColony, tick int) {
	// Find alliance members under attack
	for _, conflict := range cws.ActiveConflicts {
		if !conflict.IsActive {
			continue
		}
		
		// Check if defender is alliance member
		defenderIsAlly := false
		for _, memberID := range alliance.Members {
			if conflict.Defender == memberID {
				defenderIsAlly = true
				break
			}
		}
		
		if defenderIsAlly {
			// Rally other alliance members to help
			defender := cws.findColonyByID(colonies, conflict.Defender)
			if defender != nil {
				cws.rallyAlliesForDefense(alliance, defender, conflict, colonies, tick)
			}
		}
	}
}

// rallyAlliesForDefense brings in alliance members to help defend
func (cws *ColonyWarfareSystem) rallyAlliesForDefense(alliance *Alliance, defender *CasteColony, 
	conflict *Conflict, colonies []*CasteColony, tick int) {
	
	for _, memberID := range alliance.Members {
		if memberID == defender.ID {
			continue // Skip the defender itself
		}
		
		ally := cws.findColonyByID(colonies, memberID)
		if ally == nil {
			continue
		}
		
		// Check if ally is not already in conflict
		allyInConflict := false
		for _, activeConflict := range cws.ActiveConflicts {
			if activeConflict.Attacker == ally.ID || activeConflict.Defender == ally.ID {
				allyInConflict = true
				break
			}
		}
		
		if !allyInConflict {
			// Calculate distance to determine if ally can help
			distance := math.Sqrt(
				math.Pow(ally.NestLocation.X-defender.NestLocation.X, 2) +
				math.Pow(ally.NestLocation.Y-defender.NestLocation.Y, 2),
			)
			
			// Only help if reasonably close (within 100 units)
			if distance <= 100 {
				// Add defensive bonus to the conflict
				cws.addAlliedDefensiveSupport(conflict, ally, defender)
			}
		}
	}
}

// addAlliedDefensiveSupport provides military support to an ally
func (cws *ColonyWarfareSystem) addAlliedDefensiveSupport(conflict *Conflict, ally, defender *CasteColony) {
	// Calculate support strength (based on ally's military capacity)
	supportStrength := cws.calculateMilitaryStrength(ally) * 0.3 // 30% of ally's strength
	
	// Apply support by temporarily boosting defender's strength
	// This is simplified - in a more complex system, you'd track supporting units
	defender.ColonyFitness += supportStrength * 0.1
	
	// Ally takes some losses for helping
	supportCost := int(float64(ally.ColonySize) * 0.05) // 5% casualty risk for helping
	if supportCost > 0 {
		ally.ColonySize = int(math.Max(1, float64(ally.ColonySize) - float64(supportCost)))
	}
}

// coordinateJointOperations plans cooperative military actions
func (cws *ColonyWarfareSystem) coordinateJointOperations(alliance *Alliance, colonies []*CasteColony, tick int) {
	if len(alliance.Members) < 2 {
		return
	}
	
	// Only attempt joint operations occasionally
	if tick%200 != 0 { // Every 200 ticks
		return
	}
	
	// Find potential common enemies (colonies that are enemies to multiple alliance members)
	enemyCounts := make(map[int]int)
	
	for _, memberID := range alliance.Members {
		diplomacy := cws.ColonyDiplomacies[memberID]
		if diplomacy == nil {
			continue
		}
		
		for otherColonyID, relation := range diplomacy.Relations {
			if relation == Enemy {
				enemyCounts[otherColonyID]++
			}
		}
	}
	
	// Find enemies that are hostile to multiple alliance members
	for enemyID, hostileCount := range enemyCounts {
		if hostileCount >= 2 { // At least 2 alliance members consider this an enemy
			enemy := cws.findColonyByID(colonies, enemyID)
			if enemy != nil {
				// Consider joint operation against this enemy
				if cws.shouldLaunchJointOperation(alliance, enemy, colonies) {
					cws.launchJointOperation(alliance, enemy, colonies, tick)
					break // Only one joint operation at a time
				}
			}
		}
	}
}

// shouldLaunchJointOperation determines if a joint attack is worthwhile
func (cws *ColonyWarfareSystem) shouldLaunchJointOperation(alliance *Alliance, target *CasteColony, colonies []*CasteColony) bool {
	// Calculate combined alliance strength
	combinedStrength := 0.0
	memberCount := 0
	
	for _, memberID := range alliance.Members {
		member := cws.findColonyByID(colonies, memberID)
		if member != nil {
			combinedStrength += cws.calculateMilitaryStrength(member)
			memberCount++
		}
	}
	
	if memberCount < 2 {
		return false
	}
	
	targetStrength := cws.calculateMilitaryStrength(target)
	
	// Only attack if alliance has significant advantage
	return combinedStrength > targetStrength * 1.5
}

// launchJointOperation executes a coordinated attack by alliance members
func (cws *ColonyWarfareSystem) launchJointOperation(alliance *Alliance, target *CasteColony, colonies []*CasteColony, tick int) {
	// Find the alliance member closest to the target to lead the attack
	var primaryAttacker *CasteColony
	minDistance := math.Inf(1)
	
	for _, memberID := range alliance.Members {
		member := cws.findColonyByID(colonies, memberID)
		if member != nil {
			distance := math.Sqrt(
				math.Pow(member.NestLocation.X-target.NestLocation.X, 2) +
				math.Pow(member.NestLocation.Y-target.NestLocation.Y, 2),
			)
			if distance < minDistance {
				minDistance = distance
				primaryAttacker = member
			}
		}
	}
	
	if primaryAttacker == nil {
		return
	}
	
	// Start the conflict with the primary attacker
	conflict := cws.StartConflict(primaryAttacker, target, TotalWar, tick)
	if conflict != nil {
		// Add support from other alliance members
		for _, memberID := range alliance.Members {
			if memberID == primaryAttacker.ID {
				continue
			}
			
			supporter := cws.findColonyByID(colonies, memberID)
			if supporter != nil {
				// Add 20% of supporter's strength to the conflict
				supportStrength := cws.calculateMilitaryStrength(supporter) * 0.2
				primaryAttacker.ColonyFitness += supportStrength * 0.1
				
				// Supporter takes minor losses for participating
				supportCost := int(float64(supporter.ColonySize) * 0.03)
				supporter.ColonySize = int(math.Max(1, float64(supporter.ColonySize) - float64(supportCost)))
			}
		}
		
		// Mark this as a joint operation
		conflict.WarGoal = "alliance_dominance"
		conflict.Intensity = math.Min(1.0, conflict.Intensity + 0.2) // Higher intensity for joint ops
	}
}

// CreateAlliance establishes a new military alliance
func (cws *ColonyWarfareSystem) CreateAlliance(members []int, allianceType string, resourceShare float64, tick int) *Alliance {
	if len(members) < 2 {
		return nil
	}
	
	// Check that all members can form alliances (not enemies with each other)
	for i, member1 := range members {
		for j, member2 := range members {
			if i >= j {
				continue
			}
			
			diplomacy1 := cws.ColonyDiplomacies[member1]
			if diplomacy1 == nil {
				return nil
			}
			
			if diplomacy1.Relations[member2] == Enemy {
				return nil // Cannot ally with enemies
			}
		}
	}
	
	alliance := &Alliance{
		ID:            cws.NextAllianceID,
		Members:       make([]int, len(members)),
		StartTick:     tick,
		Duration:      -1, // Permanent by default
		AllianceType:  allianceType,
		SharedDefense: true,
		ResourceShare: math.Min(0.3, math.Max(0.0, resourceShare)), // Cap at 30%
		IsActive:      true,
	}
	
	copy(alliance.Members, members)
	cws.NextAllianceID++
	cws.Alliances = append(cws.Alliances, alliance)
	
	// Update diplomatic relations for all members
	for i, member1 := range members {
		diplomacy1 := cws.ColonyDiplomacies[member1]
		if diplomacy1 == nil {
			continue
		}
		
		diplomacy1.Alliances[alliance.ID] = alliance
		
		for j, member2 := range members {
			if i == j {
				continue
			}
			
			// Set all members as allies
			diplomacy1.Relations[member2] = Allied
			diplomacy1.TrustLevels[member2] = math.Max(diplomacy1.TrustLevels[member2], 0.8)
			
			// Record alliance event
			event := DiplomaticEvent{
				Tick:               tick,
				EventType:          "alliance_formed",
				OtherColonyID:      member2,
				Description:        "Military alliance established",
				ImpactOnTrust:      0.3,
				ImpactOnReputation: 0.1,
			}
			diplomacy1.RelationHistory[member2] = append(diplomacy1.RelationHistory[member2], event)
		}
	}
	
	return alliance
}

// AttemptAllianceFormation tries to form beneficial alliances
func (cws *ColonyWarfareSystem) AttemptAllianceFormation(colonies []*CasteColony, tick int) {
	if tick%300 != 0 { // Check for new alliances every 300 ticks
		return
	}
	
	// Look for potential alliance pairs
	for i, colony1 := range colonies {
		for j, colony2 := range colonies {
			if i >= j {
				continue
			}
			
			if cws.shouldFormAlliance(colony1, colony2) {
				// Check if they're not already in an alliance together
				alreadyAllied := false
				for _, alliance := range cws.Alliances {
					if !alliance.IsActive {
						continue
					}
					
					hasColony1 := false
					hasColony2 := false
					for _, memberID := range alliance.Members {
						if memberID == colony1.ID {
							hasColony1 = true
						}
						if memberID == colony2.ID {
							hasColony2 = true
						}
					}
					
					if hasColony1 && hasColony2 {
						alreadyAllied = true
						break
					}
				}
				
				if !alreadyAllied {
					members := []int{colony1.ID, colony2.ID}
					resourceShare := 0.1 + rand.Float64()*0.1 // 10-20% resource sharing
					alliance := cws.CreateAlliance(members, "defensive", resourceShare, tick)
					
					if alliance != nil {
						// Successfully formed alliance
						break
					}
				}
			}
		}
	}
}

// shouldFormAlliance determines if two colonies should form an alliance
func (cws *ColonyWarfareSystem) shouldFormAlliance(colony1, colony2 *CasteColony) bool {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 == nil {
		return false
	}
	
	// Must have good relations and high trust
	relation := diplomacy1.Relations[colony2.ID]
	trust := diplomacy1.TrustLevels[colony2.ID]
	
	if relation == Enemy || trust < 0.7 {
		return false
	}
	
	// Look for common threats (shared enemies)
	commonThreats := 0
	diplomacy2 := cws.ColonyDiplomacies[colony2.ID]
	if diplomacy2 != nil {
		for enemyID := range diplomacy1.Relations {
			if diplomacy1.Relations[enemyID] == Enemy && diplomacy2.Relations[enemyID] == Enemy {
				commonThreats++
			}
		}
	}
	
	// More likely to ally if they have common enemies
	return commonThreats > 0 && rand.Float64() < 0.3
}

// calculateTrustBonus returns a multiplier based on trust level between colonies
func (cws *ColonyWarfareSystem) calculateTrustBonus(colony1, colony2 *CasteColony) float64 {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 == nil {
		return 1.0 // No diplomacy data, neutral bonus
	}
	
	trust := diplomacy1.TrustLevels[colony2.ID]
	
	// Trust bonus ranges from 0.5x (low trust) to 2.0x (very high trust)
	// Base trust of 0.5 gives 1.0x multiplier
	if trust <= 0.0 {
		return 0.5 // Very low trust penalty
	} else if trust <= 0.3 {
		return 0.5 + (trust * 1.67) // 0.5 to 1.0
	} else if trust <= 0.7 {
		return 1.0 + ((trust - 0.3) * 1.25) // 1.0 to 1.5
	} else {
		return 1.5 + ((trust - 0.7) * 1.67) // 1.5 to 2.0
	}
}

// calculateRelationshipMultiplier returns a multiplier based on diplomatic relationship
func (cws *ColonyWarfareSystem) calculateRelationshipMultiplier(colony1, colony2 *CasteColony) float64 {
	diplomacy1 := cws.ColonyDiplomacies[colony1.ID]
	if diplomacy1 == nil {
		return 1.0 // No diplomacy data, neutral multiplier
	}
	
	relation := diplomacy1.Relations[colony2.ID]
	
	switch relation {
	case Allied:
		return 1.5 // 50% bonus for allies
	case Trading:
		return 1.3 // 30% bonus for trading partners
	case Neutral:
		return 1.0 // No bonus/penalty for neutral
	case Truce:
		return 0.8 // 20% penalty for truce (cautious trading)
	case Vassal:
		return 1.2 // 20% bonus for vassal relationships
	case Enemy:
		return 0.1 // 90% penalty for enemies (barely any trade)
	default:
		return 1.0
	}
}