package main

import (
	"math"
	"math/rand"
)

// InsectTraits represents insect-specific characteristics
type InsectTraits struct {
	SwarmCapability    float64 `json:"swarm_capability"`    // Ability to coordinate in large groups
	PheromoneSensitivity float64 `json:"pheromone_sensitivity"` // Sensitivity to chemical signals
	PheromoneProduction float64 `json:"pheromone_production"`  // Ability to produce pheromones
	ColonyLoyalty      float64 `json:"colony_loyalty"`       // Dedication to colony
	Metamorphosis      float64 `json:"metamorphosis"`        // Ability to change form/role
	ExoskeletonStrength float64 `json:"exoskeleton_strength"` // Protective exoskeleton
	FlightCapability   float64 `json:"flight_capability"`    // Flying ability for winged insects
	Eusociality       float64 `json:"eusociality"`          // Advanced social organization
}

// PheromoneType represents different types of chemical signals
type PheromoneType int

const (
	TrailPheromone PheromoneType = iota // For navigation and following paths
	AlarmPheromone                      // For danger warnings
	SexPheromone                        // For mating attraction
	QueenPheromone                      // Queen's presence and dominance
	FoodPheromone                       // Food source marking
	TerritoryPheromone                  // Territory marking
	BroodPheromone                      // Related to caring for young
	AggregationPheromone               // For gathering groups
)

// PheromoneTrail represents a chemical trail in the environment
type PheromoneTrail struct {
	ID           int           `json:"id"`
	Type         PheromoneType `json:"type"`
	Positions    []Position    `json:"positions"`     // Trail path
	Strength     []float64     `json:"strength"`      // Strength at each position
	ProducerID   int           `json:"producer_id"`   // Entity that created the trail
	CreationTick int           `json:"creation_tick"`
	DecayRate    float64       `json:"decay_rate"`    // How fast the trail fades
	MaxStrength  float64       `json:"max_strength"`  // Maximum pheromone strength
	// Enhanced persistence features
	ReinforcementCount int           `json:"reinforcement_count"` // How many times trail was reinforced
	LastReinforced     int           `json:"last_reinforced"`     // Last tick when reinforced
	UsageCount         int           `json:"usage_count"`         // How many entities have used this trail
	EnvironmentalFactor float64     `json:"environmental_factor"` // Environmental effects on persistence
	PersistenceBonus   float64       `json:"persistence_bonus"`   // Bonus persistence from usage
	WeatherResistance  float64       `json:"weather_resistance"`  // Resistance to weather effects
}

// SwarmUnit represents a coordinated group of entities acting as one
type SwarmUnit struct {
	ID              int       `json:"id"`
	Members         []*Entity `json:"-"` // Exclude from JSON to avoid cycles
	MemberIDs       []int     `json:"member_ids"`
	CenterPosition  Position  `json:"center_position"`  // Center of the swarm
	TargetPosition  Position  `json:"target_position"`  // Where the swarm is heading
	SwarmRadius     float64   `json:"swarm_radius"`     // Size of the swarm
	SwarmDensity    float64   `json:"swarm_density"`    // How tightly packed
	Coordination    float64   `json:"coordination"`     // How well coordinated
	SwarmSpeed      float64   `json:"swarm_speed"`      // Movement speed
	SwarmPurpose    string    `json:"swarm_purpose"`    // "foraging", "migration", "defense", etc.
	CreationTick    int       `json:"creation_tick"`
	LeaderEntity    *Entity   `json:"-"` // Swarm leader
	LeaderID        int       `json:"leader_id"`
}

// InsectSystem manages insect-specific behaviors and capabilities
type InsectSystem struct {
	PheromoneTrails []*PheromoneTrail `json:"pheromone_trails"`
	SwarmUnits      []*SwarmUnit      `json:"swarm_units"`
	NextTrailID     int               `json:"next_trail_id"`
	NextSwarmID     int               `json:"next_swarm_id"`
}

// NewInsectSystem creates a new insect management system
func NewInsectSystem() *InsectSystem {
	return &InsectSystem{
		PheromoneTrails: make([]*PheromoneTrail, 0),
		SwarmUnits:      make([]*SwarmUnit, 0),
		NextTrailID:     1,
		NextSwarmID:     1,
	}
}

// AddInsectTraitsToEntity adds insect traits to an entity if appropriate
func AddInsectTraitsToEntity(entity *Entity) {
	// Check if entity has insect-like characteristics
	size := entity.GetTrait("size")
	cooperation := entity.GetTrait("cooperation")
	intelligence := entity.GetTrait("intelligence")
	
	// Small, cooperative, intelligent entities are good insect candidates
	if size < -0.3 && cooperation > 0.4 && intelligence > 0.3 {
		// Add insect-specific traits
		entity.SetTrait("swarm_capability", 0.4 + rand.Float64()*0.6)
		entity.SetTrait("pheromone_sensitivity", 0.5 + rand.Float64()*0.5)
		entity.SetTrait("pheromone_production", 0.3 + rand.Float64()*0.7)
		entity.SetTrait("colony_loyalty", cooperation + 0.2)
		entity.SetTrait("metamorphosis", rand.Float64()*0.5)
		entity.SetTrait("exoskeleton_strength", 0.2 + rand.Float64()*0.4)
		entity.SetTrait("eusociality", cooperation * 0.8)
		
		// Flying capability for some insects
		if rand.Float64() < 0.4 { // 40% chance of flight
			entity.SetTrait("flight_capability", 0.3 + rand.Float64()*0.7)
			entity.SetTrait("flying_ability", entity.GetTrait("flight_capability"))
		}
		
		// Adjust other traits for insect nature
		entity.SetTrait("endurance", entity.GetTrait("endurance") + 0.3)
		entity.SetTrait("speed", entity.GetTrait("speed") + 0.2)
	}
}

// CreatePheromoneTrail creates a new pheromone trail
func (is *InsectSystem) CreatePheromoneTrail(producer *Entity, trailType PheromoneType, startPos, endPos Position) *PheromoneTrail {
	production := producer.GetTrait("pheromone_production")
	if production < 0.2 {
		return nil // Cannot produce strong enough pheromones
	}

	// Create path between start and end positions
	positions, strengths := is.calculateTrailPath(startPos, endPos, production)
	
	trail := &PheromoneTrail{
		ID:           is.NextTrailID,
		Type:         trailType,
		Positions:    positions,
		Strength:     strengths,
		ProducerID:   producer.ID,
		CreationTick: 0, // Will be set by world
		DecayRate:    0.02 + rand.Float64()*0.03, // 2-5% decay per tick
		MaxStrength:  production,
		// Enhanced persistence features
		ReinforcementCount:  0,
		LastReinforced:     0,
		UsageCount:         0,
		EnvironmentalFactor: 1.0, // Start with neutral factor
		PersistenceBonus:   0.0,
		WeatherResistance:  0.3 + rand.Float64()*0.4, // 30-70% weather resistance
	}
	
	is.NextTrailID++
	is.PheromoneTrails = append(is.PheromoneTrails, trail)
	
	return trail
}

// calculateTrailPath creates a path between two positions with pheromone strengths
func (is *InsectSystem) calculateTrailPath(start, end Position, production float64) ([]Position, []float64) {
	// Calculate number of segments based on distance
	distance := math.Sqrt(math.Pow(end.X-start.X, 2) + math.Pow(end.Y-start.Y, 2))
	segments := int(math.Max(3, distance/2.0)) // At least 3 segments, one every 2 units
	
	positions := make([]Position, segments)
	strengths := make([]float64, segments)
	
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		
		// Linear interpolation between start and end
		positions[i] = Position{
			X: start.X + t*(end.X-start.X),
			Y: start.Y + t*(end.Y-start.Y),
		}
		
		// Strength varies along trail (stronger in middle)
		strengthFactor := 1.0 - math.Abs(t-0.5)*2.0 // Peak at middle
		strengths[i] = production * (0.5 + strengthFactor*0.5)
	}
	
	return positions, strengths
}

// FollowPheromoneTrail makes an entity follow pheromone trails
func (is *InsectSystem) FollowPheromoneTrail(entity *Entity, trailType PheromoneType) (float64, float64, bool) {
	sensitivity := entity.GetTrait("pheromone_sensitivity")
	if sensitivity < 0.3 {
		return 0, 0, false // Not sensitive enough
	}

	bestStrength := 0.0
	bestX, bestY := 0.0, 0.0
	found := false

	for _, trail := range is.PheromoneTrails {
		if trail.Type != trailType {
			continue
		}

		// Find closest point on trail
		closestIndex, closestDistance := is.findClosestTrailPoint(entity.Position, trail)
		
		if closestDistance > 10.0 { // Too far from trail
			continue
		}

		// Check if this trail is strong enough to follow
		strength := trail.Strength[closestIndex] * sensitivity
		if strength > bestStrength {
			bestStrength = strength
			
			// Follow toward next point on trail
			nextIndex := closestIndex + 1
			if nextIndex >= len(trail.Positions) {
				nextIndex = len(trail.Positions) - 1
			}
			
			bestX = trail.Positions[nextIndex].X
			bestY = trail.Positions[nextIndex].Y
			found = true
		}
	}

	return bestX, bestY, found
}

// ReinforcePheromoneTrail strengthens a trail when an entity uses it
func (is *InsectSystem) ReinforcePheromoneTrail(entity *Entity, trailType PheromoneType, currentTick int) {
	production := entity.GetTrait("pheromone_production")
	if production < 0.1 {
		return // Cannot reinforce trails effectively
	}
	
	// Find trail that entity is currently on or near
	entityPos := entity.Position
	reinforcementRadius := 3.0
	
	for _, trail := range is.PheromoneTrails {
		if trail.Type != trailType {
			continue
		}
		
		// Check if entity is near this trail
		for _, trailPos := range trail.Positions {
			distance := math.Sqrt(math.Pow(entityPos.X-trailPos.X, 2) + math.Pow(entityPos.Y-trailPos.Y, 2))
			
			if distance <= reinforcementRadius {
				// Calculate reinforcement strength based on entity traits
				reinforcementStrength := production * 0.5
				
				// Apply reinforcement to nearby trail points
				for j := range trail.Positions {
					trailDistance := math.Sqrt(math.Pow(entityPos.X-trail.Positions[j].X, 2) + 
						math.Pow(entityPos.Y-trail.Positions[j].Y, 2))
					
					if trailDistance <= reinforcementRadius {
						// Strengthen trail proportional to distance
						strengthMultiplier := 1.0 - (trailDistance / reinforcementRadius)
						additionalStrength := reinforcementStrength * strengthMultiplier
						
						// Add strength but cap at maximum
						trail.Strength[j] = math.Min(trail.MaxStrength, trail.Strength[j] + additionalStrength)
					}
				}
				
				// Update trail statistics
				trail.ReinforcementCount++
				trail.LastReinforced = currentTick
				trail.UsageCount++
				
				// Increase persistence bonus based on usage
				trail.PersistenceBonus = math.Min(0.5, trail.PersistenceBonus + 0.02) // Max 50% bonus
				
				return // Only reinforce one trail per call
			}
		}
	}
}

// reinforceTrailAtPosition reinforces a specific trail near the entity's position
func (is *InsectSystem) reinforceTrailAtPosition(entity *Entity, trail *PheromoneTrail, currentTick int) {
	production := entity.GetTrait("pheromone_production")
	if production < 0.1 {
		return // Cannot reinforce trails effectively
	}
	
	entityPos := entity.Position
	reinforcementRadius := 3.0
	
	// Apply reinforcement to nearby trail points
	reinforced := false
	for i := range trail.Positions {
		distance := math.Sqrt(math.Pow(entityPos.X-trail.Positions[i].X, 2) + 
			math.Pow(entityPos.Y-trail.Positions[i].Y, 2))
		
		if distance <= reinforcementRadius {
			// Strengthen trail proportional to distance
			strengthMultiplier := 1.0 - (distance / reinforcementRadius)
			additionalStrength := production * 0.3 * strengthMultiplier
			
			// Add strength but cap at maximum
			trail.Strength[i] = math.Min(trail.MaxStrength, trail.Strength[i] + additionalStrength)
			reinforced = true
		}
	}
	
	// Update trail statistics if reinforcement occurred
	if reinforced {
		trail.ReinforcementCount++
		trail.LastReinforced = currentTick
		trail.UsageCount++
		
		// Increase persistence bonus based on usage
		trail.PersistenceBonus = math.Min(0.5, trail.PersistenceBonus + 0.01) // Max 50% bonus
	}
}

// findClosestTrailPoint finds the closest point on a pheromone trail
func (is *InsectSystem) findClosestTrailPoint(pos Position, trail *PheromoneTrail) (int, float64) {
	if len(trail.Positions) == 0 {
		return 0, math.Inf(1)
	}

	closestIndex := 0
	minDistance := math.Sqrt(math.Pow(pos.X-trail.Positions[0].X, 2) + math.Pow(pos.Y-trail.Positions[0].Y, 2))

	for i, trailPos := range trail.Positions {
		distance := math.Sqrt(math.Pow(pos.X-trailPos.X, 2) + math.Pow(pos.Y-trailPos.Y, 2))
		if distance < minDistance {
			minDistance = distance
			closestIndex = i
		}
	}

	return closestIndex, minDistance
}

// CreateSwarmUnit forms a swarm from compatible entities
func (is *InsectSystem) CreateSwarmUnit(entities []*Entity, purpose string) *SwarmUnit {
	if len(entities) < 5 { // Minimum swarm size
		return nil
	}

	// Check if entities are suitable for swarming
	avgSwarmCapability := 0.0
	avgCooperation := 0.0
	compatibleEntities := make([]*Entity, 0)

	for _, entity := range entities {
		if !entity.IsAlive {
			continue
		}
		
		swarmCapability := entity.GetTrait("swarm_capability")
		cooperation := entity.GetTrait("cooperation")
		
		if swarmCapability > 0.3 && cooperation > 0.4 {
			compatibleEntities = append(compatibleEntities, entity)
			avgSwarmCapability += swarmCapability
			avgCooperation += cooperation
		}
	}

	if len(compatibleEntities) < 5 {
		return nil
	}

	avgSwarmCapability /= float64(len(compatibleEntities))
	avgCooperation /= float64(len(compatibleEntities))

	// Find leader (highest intelligence + swarm capability)
	var leader *Entity
	maxLeaderScore := -1.0
	
	for _, entity := range compatibleEntities {
		score := entity.GetTrait("intelligence") + entity.GetTrait("swarm_capability")
		if score > maxLeaderScore {
			maxLeaderScore = score
			leader = entity
		}
	}

	if leader == nil {
		return nil
	}

	// Calculate swarm center
	centerX, centerY := is.calculateSwarmCenter(compatibleEntities)
	
	swarm := &SwarmUnit{
		ID:              is.NextSwarmID,
		Members:         compatibleEntities,
		MemberIDs:       make([]int, len(compatibleEntities)),
		CenterPosition:  Position{X: centerX, Y: centerY},
		TargetPosition:  Position{X: centerX, Y: centerY}, // Start with no target
		SwarmRadius:     is.calculateSwarmRadius(compatibleEntities),
		SwarmDensity:    0.7 + avgCooperation*0.3,
		Coordination:    avgSwarmCapability * avgCooperation,
		SwarmSpeed:      is.calculateSwarmSpeed(compatibleEntities),
		SwarmPurpose:    purpose,
		LeaderEntity:    leader,
		LeaderID:        leader.ID,
	}

	// Set member IDs
	for i, entity := range compatibleEntities {
		swarm.MemberIDs[i] = entity.ID
		// Mark entity as part of swarm
		entity.SetTrait("swarm_member", 1.0)
		entity.SetTrait("swarm_id", float64(swarm.ID))
	}

	is.NextSwarmID++
	is.SwarmUnits = append(is.SwarmUnits, swarm)

	return swarm
}

// calculateSwarmCenter finds the geometric center of swarm members
func (is *InsectSystem) calculateSwarmCenter(entities []*Entity) (float64, float64) {
	if len(entities) == 0 {
		return 0, 0
	}

	totalX, totalY := 0.0, 0.0
	for _, entity := range entities {
		totalX += entity.Position.X
		totalY += entity.Position.Y
	}

	return totalX / float64(len(entities)), totalY / float64(len(entities))
}

// calculateSwarmRadius determines the size of the swarm
func (is *InsectSystem) calculateSwarmRadius(entities []*Entity) float64 {
	if len(entities) <= 1 {
		return 5.0
	}

	centerX, centerY := is.calculateSwarmCenter(entities)
	maxDistance := 0.0

	for _, entity := range entities {
		distance := math.Sqrt(math.Pow(entity.Position.X-centerX, 2) + math.Pow(entity.Position.Y-centerY, 2))
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	return math.Max(5.0, maxDistance*1.2) // 20% larger than furthest member
}

// calculateSwarmSpeed determines the movement speed of the swarm
func (is *InsectSystem) calculateSwarmSpeed(entities []*Entity) float64 {
	if len(entities) == 0 {
		return 0
	}

	totalSpeed := 0.0
	for _, entity := range entities {
		totalSpeed += entity.GetTrait("speed")
	}

	avgSpeed := totalSpeed / float64(len(entities))
	return avgSpeed * 0.8 // Swarms move slightly slower than individuals
}

// UpdateSwarmMovement coordinates swarm member movements
func (is *InsectSystem) UpdateSwarmMovement(swarm *SwarmUnit) {
	if len(swarm.Members) == 0 || swarm.LeaderEntity == nil {
		return
	}

	// Update swarm center
	swarm.CenterPosition.X, swarm.CenterPosition.Y = is.calculateSwarmCenter(swarm.Members)

	// Determine swarm target based on purpose
	switch swarm.SwarmPurpose {
	case "foraging":
		is.updateForagingSwarmTarget(swarm)
	case "migration":
		is.updateMigrationSwarmTarget(swarm)
	case "defense":
		is.updateDefenseSwarmTarget(swarm)
	case "exploration":
		is.updateExplorationSwarmTarget(swarm)
	default:
		// Random movement
		angle := rand.Float64() * 2 * math.Pi
		distance := 20.0 + rand.Float64()*30.0
		swarm.TargetPosition.X = swarm.CenterPosition.X + math.Cos(angle)*distance
		swarm.TargetPosition.Y = swarm.CenterPosition.Y + math.Sin(angle)*distance
	}

	// Move swarm members toward formation positions
	formations := is.calculateSwarmFormation(swarm)
	
	for i, member := range swarm.Members {
		if !member.IsAlive || i >= len(formations) {
			continue
		}

		formation := formations[i]
		
		// Calculate movement toward formation position
		dx := formation.X - member.Position.X
		dy := formation.Y - member.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > 0.5 { // Only move if not already in position
			// Normalize and apply swarm speed
			speed := swarm.SwarmSpeed * swarm.Coordination
			dx = (dx / distance) * speed
			dy = (dy / distance) * speed

			member.Position.X += dx
			member.Position.Y += dy
		}
	}
}

// calculateSwarmFormation determines positions for swarm members
func (is *InsectSystem) calculateSwarmFormation(swarm *SwarmUnit) []Position {
	formations := make([]Position, len(swarm.Members))
	
	// Direction toward target
	targetDx := swarm.TargetPosition.X - swarm.CenterPosition.X
	targetDy := swarm.TargetPosition.Y - swarm.CenterPosition.Y
	targetDistance := math.Sqrt(targetDx*targetDx + targetDy*targetDy)
	
	if targetDistance == 0 {
		targetDx, targetDy = 1, 0 // Default direction
	} else {
		targetDx /= targetDistance
		targetDy /= targetDistance
	}

	// Create formation based on swarm density and size
	switch swarm.SwarmPurpose {
	case "foraging":
		// Spread out formation for better coverage
		is.createSpreadFormation(formations, swarm, targetDx, targetDy)
	case "defense":
		// Tight defensive formation
		is.createDefensiveFormation(formations, swarm, targetDx, targetDy)
	case "migration":
		// V-shaped formation for efficiency
		is.createMigrationFormation(formations, swarm, targetDx, targetDy)
	default:
		// Circular swarm formation
		is.createCircularFormation(formations, swarm, targetDx, targetDy)
	}

	return formations
}

// createSpreadFormation creates a spread out formation for foraging
func (is *InsectSystem) createSpreadFormation(formations []Position, swarm *SwarmUnit, dirX, dirY float64) {
	spacing := swarm.SwarmRadius / swarm.SwarmDensity
	gridSize := int(math.Ceil(math.Sqrt(float64(len(formations)))))
	
	for i := range formations {
		row := i / gridSize
		col := i % gridSize
		
		offsetX := float64(col-gridSize/2) * spacing
		offsetY := float64(row-gridSize/2) * spacing
		
		// Rotate offset based on movement direction
		rotatedX := offsetX*dirX - offsetY*dirY
		rotatedY := offsetX*dirY + offsetY*dirX
		
		formations[i] = Position{
			X: swarm.CenterPosition.X + rotatedX,
			Y: swarm.CenterPosition.Y + rotatedY,
		}
	}
}

// createDefensiveFormation creates a tight defensive formation
func (is *InsectSystem) createDefensiveFormation(formations []Position, swarm *SwarmUnit, dirX, dirY float64) {
	// Concentric circles with leader in center
	for i := range formations {
		if i == 0 && swarm.LeaderEntity != nil {
			// Leader in center
			formations[i] = swarm.CenterPosition
		} else {
			// Other members in rings around center
			ring := (i-1)/6 + 1 // 6 members per ring
			angleIndex := (i - 1) % 6
			
			angle := float64(angleIndex) * math.Pi / 3.0
			radius := float64(ring) * 3.0 * swarm.SwarmDensity
			
			formations[i] = Position{
				X: swarm.CenterPosition.X + math.Cos(angle)*radius,
				Y: swarm.CenterPosition.Y + math.Sin(angle)*radius,
			}
		}
	}
}

// createMigrationFormation creates a V-shaped formation for efficient travel
func (is *InsectSystem) createMigrationFormation(formations []Position, swarm *SwarmUnit, dirX, dirY float64) {
	leaderIndex := 0
	// Find leader index
	for i, member := range swarm.Members {
		if member == swarm.LeaderEntity {
			leaderIndex = i
			break
		}
	}

	// Leader at front
	formations[leaderIndex] = Position{
		X: swarm.CenterPosition.X + dirX*5.0,
		Y: swarm.CenterPosition.Y + dirY*5.0,
	}

	// Other members in V formation behind leader
	memberIndex := 0
	for i := range formations {
		if i == leaderIndex {
			continue
		}

		side := float64(1 - 2*(memberIndex%2)) // Alternate sides (-1, 1, -1, 1...)
		position := float64(memberIndex/2 + 1)
		
		// Calculate perpendicular direction for V shape
		perpX := -dirY * side
		perpY := dirX * side
		
		formations[i] = Position{
			X: swarm.CenterPosition.X + dirX*(-position*2.0) + perpX*position*1.5,
			Y: swarm.CenterPosition.Y + dirY*(-position*2.0) + perpY*position*1.5,
		}
		
		memberIndex++
	}
}

// createCircularFormation creates a basic circular formation
func (is *InsectSystem) createCircularFormation(formations []Position, swarm *SwarmUnit, dirX, dirY float64) {
	for i := range formations {
		angle := float64(i) * 2.0 * math.Pi / float64(len(formations))
		radius := swarm.SwarmRadius * swarm.SwarmDensity
		
		formations[i] = Position{
			X: swarm.CenterPosition.X + math.Cos(angle)*radius,
			Y: swarm.CenterPosition.Y + math.Sin(angle)*radius,
		}
	}
}

// Update target for foraging swarms
func (is *InsectSystem) updateForagingSwarmTarget(swarm *SwarmUnit) {
	// Look for pheromone trails leading to food
	if targetX, targetY, found := is.FollowPheromoneTrail(swarm.LeaderEntity, FoodPheromone); found {
		swarm.TargetPosition.X = targetX
		swarm.TargetPosition.Y = targetY
	} else {
		// Random foraging movement
		angle := rand.Float64() * 2 * math.Pi
		distance := 15.0 + rand.Float64()*25.0
		swarm.TargetPosition.X = swarm.CenterPosition.X + math.Cos(angle)*distance
		swarm.TargetPosition.Y = swarm.CenterPosition.Y + math.Sin(angle)*distance
	}
}

// Update target for migration swarms
func (is *InsectSystem) updateMigrationSwarmTarget(swarm *SwarmUnit) {
	// Follow migration trails or head toward better biomes
	if targetX, targetY, found := is.FollowPheromoneTrail(swarm.LeaderEntity, TrailPheromone); found {
		swarm.TargetPosition.X = targetX
		swarm.TargetPosition.Y = targetY
	} else {
		// Head toward better environment (simplified)
		angle := float64(swarm.ID) * 0.5 // Consistent direction based on swarm ID
		distance := 40.0 + rand.Float64()*60.0
		swarm.TargetPosition.X = swarm.CenterPosition.X + math.Cos(angle)*distance
		swarm.TargetPosition.Y = swarm.CenterPosition.Y + math.Sin(angle)*distance
	}
}

// Update target for defense swarms
func (is *InsectSystem) updateDefenseSwarmTarget(swarm *SwarmUnit) {
	// Follow alarm pheromones or stay near nest
	if targetX, targetY, found := is.FollowPheromoneTrail(swarm.LeaderEntity, AlarmPheromone); found {
		swarm.TargetPosition.X = targetX
		swarm.TargetPosition.Y = targetY
	} else {
		// Stay near current position for defense
		swarm.TargetPosition = swarm.CenterPosition
	}
}

// Update target for exploration swarms
func (is *InsectSystem) updateExplorationSwarmTarget(swarm *SwarmUnit) {
	// Explore new areas
	angle := rand.Float64() * 2 * math.Pi
	distance := 30.0 + rand.Float64()*50.0
	swarm.TargetPosition.X = swarm.CenterPosition.X + math.Cos(angle)*distance
	swarm.TargetPosition.Y = swarm.CenterPosition.Y + math.Sin(angle)*distance
}

// Update maintains the insect system
func (is *InsectSystem) Update(tick int) {
	// Update pheromone trails
	is.updatePheromoneTrails()
	
	// Update swarm units
	is.updateSwarmUnits()
}

// updatePheromoneTrails handles enhanced trail decay, environmental factors, and persistence
func (is *InsectSystem) updatePheromoneTrails() {
	activeTrails := make([]*PheromoneTrail, 0)

	for _, trail := range is.PheromoneTrails {
		stillActive := false
		
		// Calculate environmental decay modifier
		trail.EnvironmentalFactor = is.calculateEnvironmentalFactor(trail)
		
		// Calculate effective decay rate considering all factors
		effectiveDecayRate := is.calculateEffectiveDecayRate(trail)
		
		// Decay all points on the trail
		for i := range trail.Strength {
			trail.Strength[i] *= (1.0 - effectiveDecayRate)
			
			// Lower threshold for well-used trails
			persistenceThreshold := 0.1
			if trail.UsageCount > 5 {
				persistenceThreshold = 0.05 // Trails with high usage persist longer
			}
			
			if trail.Strength[i] > persistenceThreshold {
				stillActive = true
			}
		}

		// Keep trails that still have some strength
		if stillActive {
			activeTrails = append(activeTrails, trail)
		}
	}

	is.PheromoneTrails = activeTrails
}

// calculateEnvironmentalFactor determines how environmental conditions affect trail persistence
func (is *InsectSystem) calculateEnvironmentalFactor(trail *PheromoneTrail) float64 {
	factor := 1.0
	
	// Check weather conditions (simplified)
	// In a full implementation, this would check actual weather patterns
	weatherSeverity := rand.Float64() // 0-1, where 1 is severe weather
	
	if weatherSeverity > 0.7 {
		// Severe weather reduces persistence
		weatherEffect := (1.0 - weatherSeverity) * trail.WeatherResistance
		factor *= math.Max(0.3, weatherEffect) // Minimum 30% factor
	}
	
	// Temperature effects (simplified)
	temperatureFactor := 0.8 + rand.Float64()*0.4 // 0.8-1.2
	factor *= temperatureFactor
	
	// Humidity helps preserve pheromones
	humidityFactor := 0.9 + rand.Float64()*0.2 // 0.9-1.1
	factor *= humidityFactor
	
	return math.Max(0.2, math.Min(2.0, factor)) // Clamp between 0.2 and 2.0
}

// calculateEffectiveDecayRate calculates the actual decay rate considering all factors
func (is *InsectSystem) calculateEffectiveDecayRate(trail *PheromoneTrail) float64 {
	baseDecay := trail.DecayRate
	
	// Apply environmental factor
	environmentalDecay := baseDecay * trail.EnvironmentalFactor
	
	// Apply persistence bonus (reduces decay)
	persistenceReduction := trail.PersistenceBonus
	effectiveDecay := environmentalDecay * (1.0 - persistenceReduction)
	
	// Recently reinforced trails decay slower
	recentReinforcementBonus := 0.0
	if trail.ReinforcementCount > 0 {
		// Decay reduction based on reinforcement count and recency
		reinforcementFactor := math.Min(0.3, float64(trail.ReinforcementCount) * 0.05) // Max 30% reduction
		recentReinforcementBonus = reinforcementFactor
	}
	
	effectiveDecay *= (1.0 - recentReinforcementBonus)
	
	// Ensure minimum and maximum decay rates
	return math.Max(0.005, math.Min(0.15, effectiveDecay)) // Between 0.5% and 15% per tick
}

// updateSwarmUnits maintains swarm integrity and behavior
func (is *InsectSystem) updateSwarmUnits() {
	activeSwarms := make([]*SwarmUnit, 0)

	for _, swarm := range is.SwarmUnits {
		// Remove dead members
		aliveMembers := make([]*Entity, 0)
		aliveMemberIDs := make([]int, 0)
		
		for i, member := range swarm.Members {
			if member.IsAlive {
				aliveMembers = append(aliveMembers, member)
				aliveMemberIDs = append(aliveMemberIDs, swarm.MemberIDs[i])
			} else {
				// Remove swarm markings from dead entities
				member.SetTrait("swarm_member", 0.0)
				member.SetTrait("swarm_id", 0.0)
			}
		}

		swarm.Members = aliveMembers
		swarm.MemberIDs = aliveMemberIDs

		// Update leader if current leader is dead
		if swarm.LeaderEntity == nil || !swarm.LeaderEntity.IsAlive {
			swarm.LeaderEntity = is.findNewSwarmLeader(swarm.Members)
			if swarm.LeaderEntity != nil {
				swarm.LeaderID = swarm.LeaderEntity.ID
			}
		}

		// Keep swarms with enough members and a leader
		if len(swarm.Members) >= 3 && swarm.LeaderEntity != nil {
			// Update swarm movement
			is.UpdateSwarmMovement(swarm)
			activeSwarms = append(activeSwarms, swarm)
		} else {
			// Disband swarm - remove markings from remaining members
			for _, member := range swarm.Members {
				member.SetTrait("swarm_member", 0.0)
				member.SetTrait("swarm_id", 0.0)
			}
		}
	}

	is.SwarmUnits = activeSwarms
}

// findNewSwarmLeader finds a replacement leader for a swarm
func (is *InsectSystem) findNewSwarmLeader(members []*Entity) *Entity {
	if len(members) == 0 {
		return nil
	}

	var bestLeader *Entity
	maxScore := -1.0

	for _, member := range members {
		if !member.IsAlive {
			continue
		}

		score := member.GetTrait("intelligence") + member.GetTrait("swarm_capability")
		if score > maxScore {
			maxScore = score
			bestLeader = member
		}
	}

	return bestLeader
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GetSwarmByMember finds the swarm that contains a specific entity
func (is *InsectSystem) GetSwarmByMember(entity *Entity) *SwarmUnit {
	for _, swarm := range is.SwarmUnits {
		for _, member := range swarm.Members {
			if member.ID == entity.ID {
				return swarm
			}
		}
	}
	return nil
}

// IsEntityInsectLike determines if an entity has insect characteristics
func IsEntityInsectLike(entity *Entity) bool {
	size := entity.GetTrait("size")
	cooperation := entity.GetTrait("cooperation")
	swarmCapability := entity.GetTrait("swarm_capability")
	
	// Small, cooperative entities with swarm capability are insect-like
	return size < -0.2 && cooperation > 0.4 && swarmCapability > 0.3
}

// GetPheromoneStrengthAtPosition calculates total pheromone strength at a position
func (is *InsectSystem) GetPheromoneStrengthAtPosition(pos Position, pheromoneType PheromoneType) float64 {
	totalStrength := 0.0

	for _, trail := range is.PheromoneTrails {
		if trail.Type != pheromoneType {
			continue
		}

		// Find closest point on trail and get strength there
		closestIndex, distance := is.findClosestTrailPoint(pos, trail)
		
		if distance < 5.0 { // Within pheromone detection range
			// Strength decreases with distance
			strengthAtPosition := trail.Strength[closestIndex] * (1.0 - distance/5.0)
			totalStrength += strengthAtPosition
		}
	}

	return totalStrength
}