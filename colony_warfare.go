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
}