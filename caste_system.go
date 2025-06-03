package main

import (
	"math"
	"math/rand"
)

// CasteRole represents different roles entities can specialize into
type CasteRole int

const (
	Worker CasteRole = iota // Basic labor and foraging
	Soldier                 // Defense and combat
	Queen                   // Reproduction and leadership
	Drone                   // Male reproductive role
	Scout                   // Exploration and intelligence gathering
	Nurse                   // Care for young and weak
	Builder                 // Construction and maintenance
	Specialist              // Technical roles (varies by environment)
)

// String returns the string representation of CasteRole
func (cr CasteRole) String() string {
	switch cr {
	case Worker:
		return "worker"
	case Soldier:
		return "soldier"
	case Queen:
		return "queen"
	case Drone:
		return "drone"
	case Scout:
		return "scout"
	case Nurse:
		return "nurse"
	case Builder:
		return "builder"
	case Specialist:
		return "specialist"
	default:
		return "unknown"
	}
}

// CasteStatus tracks an entity's role within a caste system
type CasteStatus struct {
	Role                CasteRole `json:"role"`
	RoleSpecialization  float64   `json:"role_specialization"`  // How specialized the entity is (0.0-1.0)
	CasteLoyalty        float64   `json:"caste_loyalty"`        // Loyalty to the caste system
	RoleEfficiency      float64   `json:"role_efficiency"`      // How efficiently they perform their role
	CanChangeRole       bool      `json:"can_change_role"`      // Whether role can change
	RoleAssignmentTick  int       `json:"role_assignment_tick"` // When role was assigned
	ReproductiveCapability float64 `json:"reproductive_capability"` // Reproductive ability (varies by caste)
}

// NewCasteStatus creates a new caste status for an entity
func NewCasteStatus(role CasteRole) *CasteStatus {
	reproductiveCapability := 1.0
	canChangeRole := true

	// Set role-specific defaults
	switch role {
	case Queen:
		reproductiveCapability = 3.0 // Queens are highly reproductive
		canChangeRole = false        // Queens don't change roles
	case Drone:
		reproductiveCapability = 1.5 // Drones are reproductive males
		canChangeRole = false        // Drones don't change roles
	case Worker:
		reproductiveCapability = 0.1 // Workers have reduced reproduction
		canChangeRole = true
	case Soldier:
		reproductiveCapability = 0.2 // Soldiers have reduced reproduction
		canChangeRole = true
	case Scout:
		reproductiveCapability = 0.3 // Scouts have reduced reproduction
		canChangeRole = true
	case Nurse:
		reproductiveCapability = 0.2 // Nurses focus on care, not reproduction
		canChangeRole = true
	case Builder:
		reproductiveCapability = 0.2 // Builders focus on construction
		canChangeRole = true
	case Specialist:
		reproductiveCapability = 0.5 // Specialists vary
		canChangeRole = true
	}

	return &CasteStatus{
		Role:                   role,
		RoleSpecialization:     0.5 + rand.Float64()*0.3, // 0.5-0.8
		CasteLoyalty:           0.7 + rand.Float64()*0.3, // 0.7-1.0
		RoleEfficiency:         0.4 + rand.Float64()*0.4, // 0.4-0.8
		CanChangeRole:          canChangeRole,
		ReproductiveCapability: reproductiveCapability,
	}
}

// CasteColony represents a colony with a caste-based social structure
type CasteColony struct {
	ID                int               `json:"id"`
	Members           []*Entity         `json:"-"` // Exclude from JSON to avoid cycles
	MemberIDs         []int             `json:"member_ids"`
	Queens            []*Entity         `json:"-"` // Queens in the colony
	QueenIDs          []int             `json:"queen_ids"`
	CasteDistribution map[CasteRole]int `json:"caste_distribution"`
	ColonySize        int               `json:"colony_size"`
	Territory         []Position        `json:"territory"`
	NestLocation      Position          `json:"nest_location"`
	ColonyAge         int               `json:"colony_age"`
	ColonyFitness     float64           `json:"colony_fitness"`
	MaxColonySize     int               `json:"max_colony_size"`
	ReproductionRate  float64           `json:"reproduction_rate"`
	CreationTick      int               `json:"creation_tick"`
}

// NewCasteColony creates a new caste-based colony
func NewCasteColony(id int, queen *Entity, nestLocation Position) *CasteColony {
	// Ensure queen has proper caste status
	if queen.CasteStatus == nil {
		queen.CasteStatus = NewCasteStatus(Queen)
	} else {
		queen.CasteStatus.Role = Queen
		queen.CasteStatus.ReproductiveCapability = 3.0
		queen.CasteStatus.CanChangeRole = false
	}

	// Modify queen traits for her role
	queen.SetTrait("reproductive_capability", 3.0)
	queen.SetTrait("leadership", queen.GetTrait("leadership") + 0.5)
	queen.SetTrait("intelligence", queen.GetTrait("intelligence") + 0.3)

	colony := &CasteColony{
		ID:                id,
		Members:           []*Entity{queen},
		MemberIDs:         []int{queen.ID},
		Queens:            []*Entity{queen},
		QueenIDs:          []int{queen.ID},
		CasteDistribution: make(map[CasteRole]int),
		ColonySize:        1,
		Territory:         []Position{nestLocation},
		NestLocation:      nestLocation,
		ColonyAge:         0,
		ColonyFitness:     queen.Fitness,
		MaxColonySize:     100 + rand.Intn(400), // 100-500 members
		ReproductionRate:  0.1,
	}

	colony.CasteDistribution[Queen] = 1
	return colony
}

// CanJoinColony determines if an entity can join this colony
func (cc *CasteColony) CanJoinColony(entity *Entity) bool {
	if !entity.IsAlive || cc.ColonySize >= cc.MaxColonySize {
		return false
	}

	// Check if entity is already in the colony
	for _, member := range cc.Members {
		if member.ID == entity.ID {
			return false
		}
	}

	// Must have minimum cooperation and intelligence
	cooperation := entity.GetTrait("cooperation")
	intelligence := entity.GetTrait("intelligence")

	if cooperation < 0.3 || intelligence < 0.2 {
		return false
	}

	// Check species compatibility (should be same or compatible species)
	if len(cc.Queens) > 0 {
		queen := cc.Queens[0]
		if entity.Species != queen.Species {
			// Allow some cross-species compatibility for advanced colonies
			if intelligence < 0.7 || cooperation < 0.8 {
				return false
			}
		}
	}

	return true
}

// AddMember adds an entity to the colony and assigns a caste role
func (cc *CasteColony) AddMember(entity *Entity, assignedRole CasteRole) bool {
	if !cc.CanJoinColony(entity) {
		return false
	}

	// Assign caste status if not already present
	if entity.CasteStatus == nil {
		entity.CasteStatus = NewCasteStatus(assignedRole)
	} else {
		// Update role if changeable
		if entity.CasteStatus.CanChangeRole {
			entity.CasteStatus.Role = assignedRole
		}
	}

	// Modify entity traits based on assigned role
	cc.modifyTraitsForRole(entity, assignedRole)

	// Add to colony
	cc.Members = append(cc.Members, entity)
	cc.MemberIDs = append(cc.MemberIDs, entity.ID)
	cc.ColonySize++
	cc.CasteDistribution[assignedRole]++

	// Update entity tribe ID to match colony
	entity.TribeID = cc.ID

	// Add queens to special tracking
	if assignedRole == Queen {
		cc.Queens = append(cc.Queens, entity)
		cc.QueenIDs = append(cc.QueenIDs, entity.ID)
	}

	return true
}

// modifyTraitsForRole adjusts entity traits based on their caste role
func (cc *CasteColony) modifyTraitsForRole(entity *Entity, role CasteRole) {
	switch role {
	case Worker:
		// Workers are efficient foragers
		entity.SetTrait("foraging_efficiency", entity.GetTrait("foraging_efficiency")+0.3)
		entity.SetTrait("endurance", entity.GetTrait("endurance")+0.2)
		entity.SetTrait("cooperation", entity.GetTrait("cooperation")+0.2)
		entity.SetTrait("reproductive_capability", 0.1)

	case Soldier:
		// Soldiers are combat specialists
		entity.SetTrait("aggression", entity.GetTrait("aggression")+0.4)
		entity.SetTrait("strength", entity.GetTrait("strength")+0.3)
		entity.SetTrait("defense", entity.GetTrait("defense")+0.3)
		entity.SetTrait("size", entity.GetTrait("size")+0.2)
		entity.SetTrait("reproductive_capability", 0.2)

	case Queen:
		// Queens are reproductive and intelligent leaders
		entity.SetTrait("reproductive_capability", 3.0)
		entity.SetTrait("intelligence", entity.GetTrait("intelligence")+0.3)
		entity.SetTrait("leadership", entity.GetTrait("leadership")+0.5)
		entity.SetTrait("size", entity.GetTrait("size")+0.3)

	case Drone:
		// Drones are reproductive males
		entity.SetTrait("reproductive_capability", 1.5)
		entity.SetTrait("mating_drive", entity.GetTrait("mating_drive")+0.4)
		entity.SetTrait("speed", entity.GetTrait("speed")+0.2)

	case Scout:
		// Scouts are fast and perceptive
		entity.SetTrait("speed", entity.GetTrait("speed")+0.4)
		entity.SetTrait("vision", entity.GetTrait("vision")+0.3)
		entity.SetTrait("intelligence", entity.GetTrait("intelligence")+0.2)
		entity.SetTrait("exploration_drive", entity.GetTrait("exploration_drive")+0.3)
		entity.SetTrait("reproductive_capability", 0.3)

	case Nurse:
		// Nurses care for young and weak
		entity.SetTrait("cooperation", entity.GetTrait("cooperation")+0.4)
		entity.SetTrait("nurturing", entity.GetTrait("nurturing")+0.5)
		entity.SetTrait("intelligence", entity.GetTrait("intelligence")+0.2)
		entity.SetTrait("reproductive_capability", 0.2)

	case Builder:
		// Builders construct and maintain structures
		entity.SetTrait("construction_skill", entity.GetTrait("construction_skill")+0.4)
		entity.SetTrait("intelligence", entity.GetTrait("intelligence")+0.2)
		entity.SetTrait("strength", entity.GetTrait("strength")+0.2)
		entity.SetTrait("reproductive_capability", 0.2)

	case Specialist:
		// Specialists have enhanced intelligence and specific skills
		entity.SetTrait("intelligence", entity.GetTrait("intelligence")+0.3)
		entity.SetTrait("specialization", entity.GetTrait("specialization")+0.4)
		entity.SetTrait("reproductive_capability", 0.5)
	}

	// Ensure traits stay within bounds
	for name, trait := range entity.Traits {
		value := math.Max(-2.0, math.Min(2.0, trait.Value))
		entity.SetTrait(name, value)
	}
}

// DetermineOptimalRole determines the best role for an entity based on traits
func (cc *CasteColony) DetermineOptimalRole(entity *Entity) CasteRole {
	// Queens are special - only one per colony typically
	if len(cc.Queens) == 0 || (len(cc.Queens) < 3 && cc.ColonySize > 50) {
		if entity.GetTrait("intelligence") > 0.7 && entity.GetTrait("leadership") > 0.5 {
			return Queen
		}
	}

	// Check for soldiers - need combat traits
	if entity.GetTrait("aggression") > 0.5 && entity.GetTrait("strength") > 0.4 {
		soldierCount := cc.CasteDistribution[Soldier]
		if float64(soldierCount)/float64(cc.ColonySize) < 0.2 { // Max 20% soldiers
			return Soldier
		}
	}

	// Check for scouts - need speed and intelligence
	if entity.GetTrait("speed") > 0.5 && entity.GetTrait("intelligence") > 0.4 {
		scoutCount := cc.CasteDistribution[Scout]
		if float64(scoutCount)/float64(cc.ColonySize) < 0.1 { // Max 10% scouts
			return Scout
		}
	}

	// Check for builders - need construction skills
	if entity.GetTrait("construction_skill") > 0.4 && entity.GetTrait("intelligence") > 0.3 {
		builderCount := cc.CasteDistribution[Builder]
		if float64(builderCount)/float64(cc.ColonySize) < 0.15 { // Max 15% builders
			return Builder
		}
	}

	// Check for nurses - need cooperation and nurturing
	if entity.GetTrait("cooperation") > 0.6 && entity.GetTrait("nurturing") > 0.4 {
		nurseCount := cc.CasteDistribution[Nurse]
		if float64(nurseCount)/float64(cc.ColonySize) < 0.15 { // Max 15% nurses
			return Nurse
		}
	}

	// Check for specialists - need high intelligence
	if entity.GetTrait("intelligence") > 0.6 {
		specialistCount := cc.CasteDistribution[Specialist]
		if float64(specialistCount)/float64(cc.ColonySize) < 0.1 { // Max 10% specialists
			return Specialist
		}
	}

	// Default to worker
	return Worker
}

// RemoveMember removes an entity from the colony
func (cc *CasteColony) RemoveMember(entity *Entity) {
	for i, member := range cc.Members {
		if member.ID == entity.ID {
			// Remove from members
			cc.Members = append(cc.Members[:i], cc.Members[i+1:]...)
			for j, id := range cc.MemberIDs {
				if id == entity.ID {
					cc.MemberIDs = append(cc.MemberIDs[:j], cc.MemberIDs[j+1:]...)
					break
				}
			}

			// Update caste distribution
			if entity.CasteStatus != nil {
				cc.CasteDistribution[entity.CasteStatus.Role]--
			}

			// Remove from queens if applicable
			if entity.CasteStatus != nil && entity.CasteStatus.Role == Queen {
				for j, queen := range cc.Queens {
					if queen.ID == entity.ID {
						cc.Queens = append(cc.Queens[:j], cc.Queens[j+1:]...)
						for k, id := range cc.QueenIDs {
							if id == entity.ID {
								cc.QueenIDs = append(cc.QueenIDs[:k], cc.QueenIDs[k+1:]...)
								break
							}
						}
						break
					}
				}
			}

			cc.ColonySize--

			// Remove caste markings from entity
			entity.TribeID = 0
			if entity.CasteStatus != nil {
				entity.CasteStatus.CasteLoyalty = 0.0
			}

			break
		}
	}
}

// PerformCasteSpecificActions executes role-specific behaviors for colony members
func (cc *CasteColony) PerformCasteSpecificActions(world *World, tick int) {
	for _, member := range cc.Members {
		if !member.IsAlive || member.CasteStatus == nil {
			continue
		}

		switch member.CasteStatus.Role {
		case Worker:
			cc.performWorkerActions(member, world, tick)
		case Soldier:
			cc.performSoldierActions(member, world, tick)
		case Queen:
			cc.performQueenActions(member, world, tick)
		case Scout:
			cc.performScoutActions(member, world, tick)
		case Nurse:
			cc.performNurseActions(member, world, tick)
		case Builder:
			cc.performBuilderActions(member, world, tick)
		case Specialist:
			cc.performSpecialistActions(member, world, tick)
		}
	}
}

// performWorkerActions defines worker-specific behaviors
func (cc *CasteColony) performWorkerActions(worker *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Workers focus on foraging and basic labor
	if worker.Energy < 50 {
		// Seek food sources
		if targetX, targetY, found := worker.SeekNutrition(world.AllPlants, world.AllEntities, 20.0); found {
			speed := worker.GetTrait("speed") * worker.CasteStatus.RoleEfficiency
			worker.MoveTo(targetX, targetY, speed)
		}
	} else {
		// Return to nest area when well-fed
		distance := math.Sqrt(math.Pow(worker.Position.X-cc.NestLocation.X, 2) + 
			math.Pow(worker.Position.Y-cc.NestLocation.Y, 2))
		if distance > 15.0 {
			speed := worker.GetTrait("speed") * 0.8
			worker.MoveTo(cc.NestLocation.X, cc.NestLocation.Y, speed)
		}
	}

	// Increase role efficiency through practice
	if rand.Float64() < 0.01 {
		worker.CasteStatus.RoleEfficiency = math.Min(1.0, worker.CasteStatus.RoleEfficiency+0.01)
	}
}

// performSoldierActions defines soldier-specific behaviors
func (cc *CasteColony) performSoldierActions(soldier *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Soldiers defend territory and hunt threats
	nearbyThreats := cc.findNearbyThreats(soldier, world.AllEntities, 25.0)
	
	if len(nearbyThreats) > 0 {
		// Engage closest threat
		closest := nearbyThreats[0]
		minDistance := soldier.DistanceTo(closest)
		
		for _, threat := range nearbyThreats[1:] {
			distance := soldier.DistanceTo(threat)
			if distance < minDistance {
				minDistance = distance
				closest = threat
			}
		}

		// Attack if close enough
		if minDistance < 3.0 && soldier.CanKill(closest) {
			soldier.Kill(closest)
		} else {
			// Move toward threat
			speed := soldier.GetTrait("speed") * 1.2 // Soldiers move faster when hunting
			soldier.MoveTo(closest.Position.X, closest.Position.Y, speed)
		}
	} else {
		// Patrol territory
		cc.patrolTerritory(soldier)
	}
}

// performQueenActions defines queen-specific behaviors
func (cc *CasteColony) performQueenActions(queen *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Queens focus on reproduction and colony management
	if queen.Energy > 60 && cc.ColonySize < cc.MaxColonySize {
		// Look for reproduction opportunities
		if queen.ReproductionStatus != nil && queen.ReproductionStatus.ReadyToMate {
			nearbyDrones := cc.findNearbyDrones(queen, 30.0)
			if len(nearbyDrones) > 0 {
				// Attempt reproduction with best drone
				bestDrone := cc.selectBestMate(queen, nearbyDrones)
				if bestDrone != nil {
					distance := queen.DistanceTo(bestDrone)
					if distance < 5.0 {
						// Reproduce (handled by reproduction system)
						world.ReproductionSystem.StartMating(queen, bestDrone, tick)
					} else {
						// Move toward mate
						queen.MoveTo(bestDrone.Position.X, bestDrone.Position.Y, queen.GetTrait("speed")*0.5)
					}
				}
			}
		}
	}

	// Stay near nest
	distance := math.Sqrt(math.Pow(queen.Position.X-cc.NestLocation.X, 2) + 
		math.Pow(queen.Position.Y-cc.NestLocation.Y, 2))
	if distance > 10.0 {
		queen.MoveTo(cc.NestLocation.X, cc.NestLocation.Y, queen.GetTrait("speed")*0.3)
	}
}

// performScoutActions defines scout-specific behaviors  
func (cc *CasteColony) performScoutActions(scout *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Scouts explore and gather information
	explorationRadius := 50.0 + scout.GetTrait("exploration_drive")*30.0
	
	// Move to unexplored areas
	angle := rand.Float64() * 2 * math.Pi
	distance := 20.0 + rand.Float64()*explorationRadius
	
	targetX := scout.Position.X + math.Cos(angle)*distance
	targetY := scout.Position.Y + math.Sin(angle)*distance
	
	speed := scout.GetTrait("speed") * 1.3 // Scouts are fast
	scout.MoveTo(targetX, targetY, speed)

	// Look for food sources and report back
	if targetX, targetY, found := scout.SeekNutrition(world.AllPlants, world.AllEntities, 30.0); found {
		// Send food signal to colony
		world.CommunicationSystem.SendSignal(scout, SignalFood, map[string]interface{}{
			"x": targetX,
			"y": targetY,
			"quality": 0.8,
		}, world.Tick)
	}
}

// performNurseActions defines nurse-specific behaviors
func (cc *CasteColony) performNurseActions(nurse *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Nurses care for young and weak colony members
	weakMembers := cc.findWeakMembers(nurse, 15.0)
	
	if len(weakMembers) > 0 {
		// Help closest weak member
		closest := weakMembers[0]
		minDistance := nurse.DistanceTo(closest)
		
		for _, weak := range weakMembers[1:] {
			distance := nurse.DistanceTo(weak)
			if distance < minDistance {
				minDistance = distance
				closest = weak
			}
		}

		if minDistance < 5.0 {
			// Provide care (transfer some energy)
			if nurse.Energy > 40 && closest.Energy < 30 {
				energyTransfer := math.Min(10.0, nurse.Energy-30.0)
				nurse.Energy -= energyTransfer
				closest.Energy += energyTransfer * 0.8 // Some loss in transfer
			}
		} else {
			// Move toward weak member
			nurse.MoveTo(closest.Position.X, closest.Position.Y, nurse.GetTrait("speed"))
		}
	} else {
		// Stay near nest when no one needs care
		distance := math.Sqrt(math.Pow(nurse.Position.X-cc.NestLocation.X, 2) + 
			math.Pow(nurse.Position.Y-cc.NestLocation.Y, 2))
		if distance > 12.0 {
			nurse.MoveTo(cc.NestLocation.X, cc.NestLocation.Y, nurse.GetTrait("speed")*0.6)
		}
	}
}

// performBuilderActions defines builder-specific behaviors
func (cc *CasteColony) performBuilderActions(builder *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Builders construct and maintain colony structures
	if world.CivilizationSystem != nil {
		// Look for structures that need repair
		for _, structure := range world.CivilizationSystem.Structures {
			if structure.Tribe != nil && structure.Tribe.ID == cc.ID {
				if structure.Health < structure.MaxHealth*0.7 {
					// Repair structure
					distance := math.Sqrt(math.Pow(builder.Position.X-structure.Position.X, 2) + 
						math.Pow(builder.Position.Y-structure.Position.Y, 2))
					
					if distance < 5.0 {
						structure.Repair(builder, 10.0)
					} else {
						// Move toward structure
						builder.MoveTo(structure.Position.X, structure.Position.Y, builder.GetTrait("speed"))
					}
					return
				}
			}
		}

		// Build new structures if needed
		if cc.ColonySize > 20 && len(world.CivilizationSystem.Structures) < cc.ColonySize/10 {
			// Try to build a cache near nest
			buildPos := Position{
				X: cc.NestLocation.X + (rand.Float64()-0.5)*20.0,
				Y: cc.NestLocation.Y + (rand.Float64()-0.5)*20.0,
			}
			
			distance := math.Sqrt(math.Pow(builder.Position.X-buildPos.X, 2) + 
				math.Pow(builder.Position.Y-buildPos.Y, 2))
			
			if distance < 3.0 {
				// Attempt to build
				tribe := cc.findColonyTribe(world)
				if tribe != nil && tribe.CanBuild(StructureCache) {
					tribe.BuildStructure(StructureCache, buildPos, builder, 
						world.CivilizationSystem.NextStructureID, world.CentralEventBus, world.Tick)
					world.CivilizationSystem.NextStructureID++
				}
			} else {
				// Move toward build location
				builder.MoveTo(buildPos.X, buildPos.Y, builder.GetTrait("speed"))
			}
		}
	}
}

// performSpecialistActions defines specialist-specific behaviors
func (cc *CasteColony) performSpecialistActions(specialist *Entity, world *World, tick int) {
	if world == nil {
		return // Skip if world is nil (testing scenario)
	}
	
	// Specialists perform advanced tasks based on colony needs
	intelligence := specialist.GetTrait("intelligence")
	
	if intelligence > 0.8 {
		// High intelligence specialists can coordinate other castes
		cc.coordinateColonyActions(specialist, world, tick)
	} else if intelligence > 0.6 {
		// Medium intelligence specialists help with complex tasks
		cc.assistComplexTasks(specialist, world, tick)
	} else {
		// Lower intelligence specialists help with basic tasks
		cc.performWorkerActions(specialist, world, tick)
	}
}

// Helper methods for caste actions

func (cc *CasteColony) findNearbyThreats(soldier *Entity, allEntities []*Entity, radius float64) []*Entity {
	threats := make([]*Entity, 0)
	
	for _, entity := range allEntities {
		if !entity.IsAlive || entity.TribeID == cc.ID {
			continue
		}
		
		distance := soldier.DistanceTo(entity)
		if distance <= radius {
			// Consider aggressive entities as threats
			if entity.GetTrait("aggression") > 0.3 || entity.Species == "predator" {
				threats = append(threats, entity)
			}
		}
	}
	
	return threats
}

func (cc *CasteColony) findNearbyDrones(queen *Entity, radius float64) []*Entity {
	drones := make([]*Entity, 0)
	
	for _, member := range cc.Members {
		if !member.IsAlive || member.CasteStatus == nil {
			continue
		}
		
		if member.CasteStatus.Role == Drone {
			distance := queen.DistanceTo(member)
			if distance <= radius {
				drones = append(drones, member)
			}
		}
	}
	
	return drones
}

func (cc *CasteColony) selectBestMate(queen *Entity, drones []*Entity) *Entity {
	if len(drones) == 0 {
		return nil
	}
	
	bestDrone := drones[0]
	bestScore := cc.calculateMateScore(queen, bestDrone)
	
	for _, drone := range drones[1:] {
		score := cc.calculateMateScore(queen, drone)
		if score > bestScore {
			bestScore = score
			bestDrone = drone
		}
	}
	
	return bestDrone
}

func (cc *CasteColony) calculateMateScore(queen *Entity, drone *Entity) float64 {
	// Score based on genetic fitness and reproductive traits
	intelligenceScore := (queen.GetTrait("intelligence") + drone.GetTrait("intelligence")) / 2.0
	fitnessScore := (queen.Fitness + drone.Fitness) / 2.0
	reproductiveScore := drone.GetTrait("reproductive_capability")
	
	return intelligenceScore*0.4 + fitnessScore*0.4 + reproductiveScore*0.2
}

func (cc *CasteColony) findWeakMembers(nurse *Entity, radius float64) []*Entity {
	weakMembers := make([]*Entity, 0)
	
	for _, member := range cc.Members {
		if !member.IsAlive || member == nurse {
			continue
		}
		
		distance := nurse.DistanceTo(member)
		if distance <= radius && member.Energy < 25.0 {
			weakMembers = append(weakMembers, member)
		}
	}
	
	return weakMembers
}

func (cc *CasteColony) patrolTerritory(soldier *Entity) {
	// Simple patrol pattern around nest
	angle := rand.Float64() * 2 * math.Pi
	patrolRadius := 20.0 + rand.Float64()*15.0
	
	targetX := cc.NestLocation.X + math.Cos(angle)*patrolRadius
	targetY := cc.NestLocation.Y + math.Sin(angle)*patrolRadius
	
	soldier.MoveTo(targetX, targetY, soldier.GetTrait("speed")*0.8)
}

func (cc *CasteColony) findColonyTribe(world *World) *Tribe {
	if world.CivilizationSystem == nil {
		return nil
	}
	
	for _, tribe := range world.CivilizationSystem.Tribes {
		if tribe.ID == cc.ID {
			return tribe
		}
	}
	
	return nil
}

func (cc *CasteColony) coordinateColonyActions(specialist *Entity, world *World, tick int) {
	// Advanced coordination - send strategic signals
	if rand.Float64() < 0.1 { // 10% chance per tick
		// Analyze colony needs and send appropriate signals
		workerCount := cc.CasteDistribution[Worker]
		if float64(workerCount)/float64(cc.ColonySize) < 0.4 {
			// Need more workers - signal for role changes
			world.CommunicationSystem.SendSignal(specialist, SignalHelp, map[string]interface{}{
				"action": "recruit_workers",
				"urgency": 0.7,
			}, world.Tick)
		}
	}
}

func (cc *CasteColony) assistComplexTasks(specialist *Entity, world *World, tick int) {
	// Help with various colony tasks
	if rand.Float64() < 0.5 {
		// Act as advanced worker
		cc.performWorkerActions(specialist, world, tick)
	} else {
		// Act as advanced scout
		cc.performScoutActions(specialist, world, tick)
	}
}

// Update maintains the caste colony
func (cc *CasteColony) Update(world *World, tick int) {
	if len(cc.Members) == 0 {
		return
	}

	// Remove dead members
	aliveMembers := make([]*Entity, 0)
	aliveMemberIDs := make([]int, 0)
	aliveQueens := make([]*Entity, 0)
	aliveQueenIDs := make([]int, 0)

	// Reset caste distribution
	cc.CasteDistribution = make(map[CasteRole]int)

	for i, member := range cc.Members {
		if member.IsAlive {
			aliveMembers = append(aliveMembers, member)
			aliveMemberIDs = append(aliveMemberIDs, cc.MemberIDs[i])
			
			if member.CasteStatus != nil {
				cc.CasteDistribution[member.CasteStatus.Role]++
				
				if member.CasteStatus.Role == Queen {
					aliveQueens = append(aliveQueens, member)
					for _, queenID := range cc.QueenIDs {
						if queenID == member.ID {
							aliveQueenIDs = append(aliveQueenIDs, queenID)
							break
						}
					}
				}
			}
		}
	}

	cc.Members = aliveMembers
	cc.MemberIDs = aliveMemberIDs
	cc.Queens = aliveQueens
	cc.QueenIDs = aliveQueenIDs
	cc.ColonySize = len(aliveMembers)

	// Perform caste-specific actions
	cc.PerformCasteSpecificActions(world, tick)

	// Update colony age and fitness
	cc.ColonyAge++
	cc.updateColonyFitness()

	// Check for role reassignments
	if tick%100 == 0 { // Every 100 ticks
		cc.reassignRoles()
	}
}

// updateColonyFitness calculates the overall colony fitness
func (cc *CasteColony) updateColonyFitness() {
	if len(cc.Members) == 0 {
		cc.ColonyFitness = 0
		return
	}

	totalFitness := 0.0
	for _, member := range cc.Members {
		totalFitness += member.Fitness
	}

	avgFitness := totalFitness / float64(len(cc.Members))
	
	// Bonus for good caste distribution
	distributionBonus := cc.calculateDistributionBonus()
	
	cc.ColonyFitness = avgFitness * (1.0 + distributionBonus)
}

// calculateDistributionBonus gives bonus for optimal caste distribution
func (cc *CasteColony) calculateDistributionBonus() float64 {
	if cc.ColonySize == 0 {
		return 0.0
	}

	bonus := 0.0
	
	// Optimal ratios (can be adjusted)
	optimalRatios := map[CasteRole]float64{
		Worker:     0.50, // 50% workers
		Soldier:    0.15, // 15% soldiers
		Scout:      0.10, // 10% scouts
		Nurse:      0.10, // 10% nurses
		Builder:    0.10, // 10% builders
		Specialist: 0.04, // 4% specialists
		Queen:      0.01, // 1% queens
	}

	for role, optimalRatio := range optimalRatios {
		currentRatio := float64(cc.CasteDistribution[role]) / float64(cc.ColonySize)
		diff := math.Abs(currentRatio - optimalRatio)
		
		// Bonus decreases as we deviate from optimal ratio
		roleBonus := math.Max(0, 0.1 - diff)
		bonus += roleBonus
	}

	return bonus / float64(len(optimalRatios))
}

// reassignRoles checks if any members should change roles
func (cc *CasteColony) reassignRoles() {
	for _, member := range cc.Members {
		if !member.IsAlive || member.CasteStatus == nil || !member.CasteStatus.CanChangeRole {
			continue
		}

		optimalRole := cc.DetermineOptimalRole(member)
		if optimalRole != member.CasteStatus.Role {
			// Check if role change is beneficial
			currentEfficiency := member.CasteStatus.RoleEfficiency
			
			// Assume new role starts with lower efficiency
			if currentEfficiency < 0.6 || rand.Float64() < 0.1 { // 10% chance to change
				// Change role
				cc.CasteDistribution[member.CasteStatus.Role]--
				member.CasteStatus.Role = optimalRole
				member.CasteStatus.RoleEfficiency = 0.4 // Reset efficiency
				member.CasteStatus.RoleAssignmentTick = cc.ColonyAge
				cc.modifyTraitsForRole(member, optimalRole)
				cc.CasteDistribution[optimalRole]++
			}
		}
	}
}

// CasteSystem manages all caste-based colonies in the simulation
type CasteSystem struct {
	Colonies       []*CasteColony `json:"colonies"`
	NextColonyID   int            `json:"next_colony_id"`
}

// NewCasteSystem creates a new caste management system
func NewCasteSystem() *CasteSystem {
	return &CasteSystem{
		Colonies:     make([]*CasteColony, 0),
		NextColonyID: 1,
	}
}

// TryFormCasteColony attempts to form a new caste-based colony
func (cs *CasteSystem) TryFormCasteColony(entities []*Entity, nestLocation Position) *CasteColony {
	if len(entities) < 3 {
		return nil
	}

	// Find the most suitable queen candidate
	var queenCandidate *Entity
	maxScore := -1.0

	for _, entity := range entities {
		if !entity.IsAlive {
			continue
		}

		// Score based on intelligence, leadership, and reproductive traits
		score := entity.GetTrait("intelligence")*0.4 +
			entity.GetTrait("leadership")*0.3 +
			entity.GetTrait("reproductive_capability")*0.2 +
			entity.GetTrait("cooperation")*0.1

		if score > maxScore && score > 0.6 { // Minimum threshold for queen
			maxScore = score
			queenCandidate = entity
		}
	}

	if queenCandidate == nil {
		return nil
	}

	// Create new colony
	colony := NewCasteColony(cs.NextColonyID, queenCandidate, nestLocation)
	cs.NextColonyID++

	// Add other entities with appropriate roles
	for _, entity := range entities {
		if entity != queenCandidate && colony.CanJoinColony(entity) {
			role := colony.DetermineOptimalRole(entity)
			colony.AddMember(entity, role)
		}
	}

	// Only create colony if we have enough diversity
	if len(colony.Members) >= 3 && len(colony.CasteDistribution) >= 2 {
		cs.Colonies = append(cs.Colonies, colony)
		return colony
	}

	return nil
}

// Update maintains all caste colonies
func (cs *CasteSystem) Update(world *World, tick int) {
	activeColonies := make([]*CasteColony, 0)

	for _, colony := range cs.Colonies {
		colony.Update(world, tick)

		// Keep colonies with at least one queen and some members
		if len(colony.Queens) > 0 && colony.ColonySize >= 2 {
			activeColonies = append(activeColonies, colony)
		}
	}

	cs.Colonies = activeColonies
}

// GetColonyByMember finds the colony that contains a specific entity
func (cs *CasteSystem) GetColonyByMember(entity *Entity) *CasteColony {
	for _, colony := range cs.Colonies {
		for _, member := range colony.Members {
			if member.ID == entity.ID {
				return colony
			}
		}
	}
	return nil
}

// AddCasteStatusToEntity adds caste status to an entity if it doesn't have one
func AddCasteStatusToEntity(entity *Entity) {
	if entity.CasteStatus == nil {
		// Determine role based on traits
		role := Worker // Default role
		
		if entity.GetTrait("intelligence") > 0.7 && entity.GetTrait("leadership") > 0.5 {
			role = Queen
		} else if entity.GetTrait("aggression") > 0.6 && entity.GetTrait("strength") > 0.5 {
			role = Soldier
		} else if entity.GetTrait("speed") > 0.6 && entity.GetTrait("intelligence") > 0.4 {
			role = Scout
		}
		
		entity.CasteStatus = NewCasteStatus(role)
	}
}