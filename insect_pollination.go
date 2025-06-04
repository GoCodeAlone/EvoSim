package main

import (
	"math"
	"math/rand"
)

// PollinatorType represents different types of pollinating insects
type PollinatorType int

const (
	GeneralistPollinator PollinatorType = iota // Can pollinate many plant types
	SpecialistPollinator                       // Focused on specific plant types
	HybridPollinator                           // Mix of generalist and specialist traits
)

// PollinatorTraits represents pollination-specific characteristics
type PollinatorTraits struct {
	PollinationEfficiency float64          `json:"pollination_efficiency"` // How effective at transferring pollen
	NectarDetection      float64          `json:"nectar_detection"`       // Ability to find nectar sources
	FlowerMemory         float64          `json:"flower_memory"`          // Memory of productive flower locations
	PollenCapacity       float64          `json:"pollen_capacity"`        // How much pollen can carry
	FlightRange          float64          `json:"flight_range"`           // Distance can travel for foraging
	PlantPreferences     map[PlantType]float64 `json:"plant_preferences"` // Preference for each plant type
	NectarNeeds          float64          `json:"nectar_needs"`          // Energy requirements from nectar
	SeasonalActivity     float64          `json:"seasonal_activity"`     // Activity level by season
}

// FlowerPatch represents a flowering plant with nectar rewards
type FlowerPatch struct {
	PlantID          int                `json:"plant_id"`
	Position         Position           `json:"position"`
	PlantType        PlantType          `json:"plant_type"`
	NectarAmount     float64            `json:"nectar_amount"`     // Available nectar
	NectarQuality    float64            `json:"nectar_quality"`    // Nutritional value
	BloomingLevel    float64            `json:"blooming_level"`    // How attractive to pollinators (0-1)
	PollenLoad       float64            `json:"pollen_load"`       // Available pollen for collection
	LastVisitTick    int                `json:"last_visit_tick"`   // When last visited by pollinator
	VisitorCount     int                `json:"visitor_count"`     // Number of recent visitors
	Specialization   []PollinatorType   `json:"specialization"`    // Which pollinator types this attracts
	BloomSeason      Season             `json:"bloom_season"`      // When this plant blooms
	NectarRegenRate  float64            `json:"nectar_regen_rate"` // How fast nectar regenerates
}

// PollinatorMemory represents a pollinator's memory of productive flower locations
type PollinatorMemory struct {
	Location        Position   `json:"location"`
	PlantType       PlantType  `json:"plant_type"`
	LastVisitTick   int        `json:"last_visit_tick"`
	NectarQuality   float64    `json:"nectar_quality"`
	SuccessRate     float64    `json:"success_rate"`     // Historical success at this location
	Distance        float64    `json:"distance"`         // Distance from current position
	MemoryStrength  float64    `json:"memory_strength"`  // How well remembered (0-1)
	VisitCount      int        `json:"visit_count"`      // Number of visits to this location
}

// PollinationEvent represents a successful pollination interaction
type PollinationEvent struct {
	PollinatorID    int       `json:"pollinator_id"`
	SourcePlantID   int       `json:"source_plant_id"`
	TargetPlantID   int       `json:"target_plant_id"`
	SourceType      PlantType `json:"source_type"`
	TargetType      PlantType `json:"target_type"`
	Efficiency      float64   `json:"efficiency"`      // Success rate of pollination
	NectarGained    float64   `json:"nectar_gained"`   // Energy gained by pollinator
	PollenTransfer  float64   `json:"pollen_transfer"` // Amount of pollen transferred
	IsCrossSpecies  bool      `json:"is_cross_species"`// Whether this is cross-species pollination
	Tick            int       `json:"tick"`
}

// InsectPollinationSystem manages pollinator-plant interactions
type InsectPollinationSystem struct {
	FlowerPatches        []*FlowerPatch     `json:"flower_patches"`
	PollinationEvents    []*PollinationEvent `json:"pollination_events"`
	PollinatorMemories   map[int][]*PollinatorMemory `json:"pollinator_memories"` // Entity ID -> memories
	TotalPollinations    int                `json:"total_pollinations"`
	CrossSpeciesPollinations int            `json:"cross_species_pollinations"`
	NectarProduced       float64            `json:"nectar_produced"`
	NectarConsumed       float64            `json:"nectar_consumed"`
	SeasonalModifier     float64            `json:"seasonal_modifier"`
	NextEventID          int                `json:"next_event_id"`
}

// NewInsectPollinationSystem creates a new pollination management system
func NewInsectPollinationSystem() *InsectPollinationSystem {
	return &InsectPollinationSystem{
		FlowerPatches:        make([]*FlowerPatch, 0),
		PollinationEvents:    make([]*PollinationEvent, 0),
		PollinatorMemories:   make(map[int][]*PollinatorMemory),
		SeasonalModifier:     1.0,
		NextEventID:          1,
	}
}

// AddPollinatorTraitsToEntity adds pollination traits to insect-like entities
func AddPollinatorTraitsToEntity(entity *Entity) {
	// Check if entity is suitable for pollination (insect-like + flying capability)
	size := entity.GetTrait("size")
	flyingAbility := entity.GetTrait("flying_ability")
	swarmCapability := entity.GetTrait("swarm_capability")
	intelligence := entity.GetTrait("intelligence")
	
	// Small flying insects with some social capability make good pollinators
	if size < -0.2 && flyingAbility > 0.3 && swarmCapability > 0.2 {
		// Determine pollinator type based on traits
		pollinatorType := GeneralistPollinator
		if intelligence > 0.7 {
			pollinatorType = SpecialistPollinator
		} else if intelligence > 0.5 && swarmCapability > 0.6 {
			pollinatorType = HybridPollinator
		}
		
		// Add pollination-specific traits
		entity.SetTrait("pollination_efficiency", 0.3 + rand.Float64()*0.6)
		entity.SetTrait("nectar_detection", 0.4 + rand.Float64()*0.5)
		entity.SetTrait("flower_memory", intelligence*0.8 + 0.2)
		entity.SetTrait("pollen_capacity", 0.2 + rand.Float64()*0.4)
		entity.SetTrait("flight_range", flyingAbility*30 + 10) // 10-40 unit range
		entity.SetTrait("nectar_needs", 0.1 + rand.Float64()*0.2)
		entity.SetTrait("seasonal_activity", 0.6 + rand.Float64()*0.4)
		entity.SetTrait("pollinator_type", float64(pollinatorType))
		
		// Set plant preferences based on pollinator type
		switch pollinatorType {
		case GeneralistPollinator:
			// Equal preference for all flowering plants
			entity.SetTrait("grass_preference", 0.6 + rand.Float64()*0.3)
			entity.SetTrait("bush_preference", 0.7 + rand.Float64()*0.3)
			entity.SetTrait("tree_preference", 0.5 + rand.Float64()*0.4)
			entity.SetTrait("cactus_preference", 0.4 + rand.Float64()*0.3)
		case SpecialistPollinator:
			// Strong preference for 1-2 plant types
			preferredType := rand.Intn(4) // Choose primary preference
			for i := 0; i < 4; i++ {
				preference := 0.1 + rand.Float64()*0.2 // Low base preference
				if i == preferredType {
					preference = 0.8 + rand.Float64()*0.2 // High preference for chosen type
				} else if rand.Float64() < 0.3 { // 30% chance of secondary preference
					preference = 0.4 + rand.Float64()*0.3
				}
				
				switch i {
				case 0:
					entity.SetTrait("grass_preference", preference)
				case 1:
					entity.SetTrait("bush_preference", preference)
				case 2:
					entity.SetTrait("tree_preference", preference)
				case 3:
					entity.SetTrait("cactus_preference", preference)
				}
			}
		case HybridPollinator:
			// Moderate preferences with some specialization
			entity.SetTrait("grass_preference", 0.3 + rand.Float64()*0.5)
			entity.SetTrait("bush_preference", 0.4 + rand.Float64()*0.5)
			entity.SetTrait("tree_preference", 0.3 + rand.Float64()*0.6)
			entity.SetTrait("cactus_preference", 0.2 + rand.Float64()*0.4)
		}
		
		// Enhance energy and endurance for pollinators
		entity.SetTrait("endurance", entity.GetTrait("endurance") + 0.2)
		entity.SetTrait("speed", entity.GetTrait("speed") + 0.1)
	}
}

// CreateFlowerPatch creates a flowering patch from a plant
func (ips *InsectPollinationSystem) CreateFlowerPatch(plant *Plant, season Season) *FlowerPatch {
	if !plant.IsAlive || plant.Age < 10 {
		return nil // Too young to flower
	}
	
	config := GetPlantConfigs()[plant.Type]
	
	// Determine if plant can flower this season
	bloomSeason := ips.getPlantBloomSeason(plant.Type)
	if season != bloomSeason && season != Spring { // Spring is universal bloom season
		return nil
	}
	
	// Calculate nectar production based on plant traits and energy
	nectarAmount := (plant.Energy * 0.1) * (1.0 + plant.GetTrait("nutrition_density")*0.5)
	nectarQuality := config.BaseNutrition * (1.0 + plant.GetTrait("nutrition_density")*0.3)
	
	// Blooming level based on plant health and age
	bloomingLevel := math.Min(1.0, (plant.Energy/config.BaseEnergy) * (math.Min(plant.Size, 3.0)/3.0))
	bloomingLevel *= (1.0 + plant.GetTrait("growth_efficiency")*0.2)
	
	// Pollen load based on plant size and type
	pollenLoad := plant.Size * config.BaseSize * (1.0 + plant.GetTrait("reproduction_rate")*0.4)
	
	// Determine which pollinators this plant attracts
	specialization := ips.determinePollinatorAttraction(plant.Type, plant.GetTrait("toxin_production"))
	
	patch := &FlowerPatch{
		PlantID:         plant.ID,
		Position:        plant.Position,
		PlantType:       plant.Type,
		NectarAmount:    nectarAmount,
		NectarQuality:   nectarQuality,
		BloomingLevel:   bloomingLevel,
		PollenLoad:      pollenLoad,
		LastVisitTick:   0,
		VisitorCount:    0,
		Specialization:  specialization,
		BloomSeason:     bloomSeason,
		NectarRegenRate: 0.1 + plant.GetTrait("growth_efficiency")*0.05,
	}
	
	return patch
}

// getPlantBloomSeason determines when different plants bloom
func (ips *InsectPollinationSystem) getPlantBloomSeason(plantType PlantType) Season {
	switch plantType {
	case PlantGrass:
		return Spring
	case PlantBush:
		return Summer
	case PlantTree:
		return Spring
	case PlantCactus:
		return Summer
	default:
		return Spring // Default to spring blooming
	}
}

// determinePollinatorAttraction determines which pollinator types are attracted to a plant
func (ips *InsectPollinationSystem) determinePollinatorAttraction(plantType PlantType, toxicity float64) []PollinatorType {
	attraction := make([]PollinatorType, 0)
	
	// All plants attract generalists to some degree
	attraction = append(attraction, GeneralistPollinator)
	
	// Specialists are attracted to specific plants
	switch plantType {
	case PlantTree, PlantBush:
		// Large flowering plants attract specialists
		attraction = append(attraction, SpecialistPollinator)
	case PlantCactus:
		// Unique plants attract hybrid pollinators
		if toxicity < 0.3 { // Not too toxic
			attraction = append(attraction, HybridPollinator)
		}
	case PlantGrass:
		// Simple plants mainly attract generalists
		if rand.Float64() < 0.3 {
			attraction = append(attraction, HybridPollinator)
		}
	}
	
	return attraction
}

// FindNearbyFlowers finds flower patches within a pollinator's range
func (ips *InsectPollinationSystem) FindNearbyFlowers(pollinator *Entity, maxDistance float64) []*FlowerPatch {
	nearbyFlowers := make([]*FlowerPatch, 0)
	
	for _, patch := range ips.FlowerPatches {
		distance := math.Sqrt(
			math.Pow(patch.Position.X-pollinator.Position.X, 2) +
			math.Pow(patch.Position.Y-pollinator.Position.Y, 2),
		)
		
		if distance <= maxDistance && patch.BloomingLevel > 0.1 {
			nearbyFlowers = append(nearbyFlowers, patch)
		}
	}
	
	return nearbyFlowers
}

// SelectBestFlower chooses the most attractive flower for a pollinator
func (ips *InsectPollinationSystem) SelectBestFlower(pollinator *Entity, flowers []*FlowerPatch) *FlowerPatch {
	if len(flowers) == 0 {
		return nil
	}
	
	pollinatorType := PollinatorType(pollinator.GetTrait("pollinator_type"))
	nectarDetection := pollinator.GetTrait("nectar_detection")
	memories := ips.PollinatorMemories[pollinator.ID]
	
	bestFlower := flowers[0]
	bestScore := 0.0
	
	for _, flower := range flowers {
		score := ips.calculateFlowerAttractiveness(pollinator, flower, pollinatorType, nectarDetection, memories)
		if score > bestScore {
			bestScore = score
			bestFlower = flower
		}
	}
	
	return bestFlower
}

// calculateFlowerAttractiveness calculates how attractive a flower is to a pollinator
func (ips *InsectPollinationSystem) calculateFlowerAttractiveness(pollinator *Entity, flower *FlowerPatch, 
	pollinatorType PollinatorType, nectarDetection float64, memories []*PollinatorMemory) float64 {
	
	score := 0.0
	
	// Base attractiveness from nectar and blooming level
	score += flower.NectarAmount * nectarDetection * flower.BloomingLevel
	
	// Plant type preference
	preference := ips.getPlantPreference(pollinator, flower.PlantType)
	score *= preference
	
	// Distance penalty
	distance := math.Sqrt(
		math.Pow(flower.Position.X-pollinator.Position.X, 2) +
		math.Pow(flower.Position.Y-pollinator.Position.Y, 2),
	)
	score *= (1.0 / (1.0 + distance*0.1))
	
	// Memory bonus
	for _, memory := range memories {
		if memory.PlantType == flower.PlantType {
			memoryDistance := math.Sqrt(
				math.Pow(memory.Location.X-flower.Position.X, 2) +
				math.Pow(memory.Location.Y-flower.Position.Y, 2),
			)
			if memoryDistance < 5.0 { // Close to remembered location
				score *= (1.0 + memory.SuccessRate*memory.MemoryStrength)
			}
		}
	}
	
	// Penalty for recent visits
	if flower.VisitorCount > 3 {
		score *= 0.5 // Reduce attractiveness of crowded flowers
	}
	
	// Pollinator type specialization bonus
	for _, specialization := range flower.Specialization {
		if specialization == pollinatorType {
			score *= 1.5
			break
		}
	}
	
	return score
}

// getPlantPreference gets a pollinator's preference for a specific plant type
func (ips *InsectPollinationSystem) getPlantPreference(pollinator *Entity, plantType PlantType) float64 {
	switch plantType {
	case PlantGrass:
		return pollinator.GetTrait("grass_preference")
	case PlantBush:
		return pollinator.GetTrait("bush_preference")
	case PlantTree:
		return pollinator.GetTrait("tree_preference")
	case PlantCactus:
		return pollinator.GetTrait("cactus_preference")
	default:
		return 0.5 // Default moderate preference
	}
}

// AttemptPollination processes a pollination attempt between pollinator and flower
func (ips *InsectPollinationSystem) AttemptPollination(pollinator *Entity, flower *FlowerPatch, tick int) bool {
	efficiency := pollinator.GetTrait("pollination_efficiency")
	pollenCapacity := pollinator.GetTrait("pollen_capacity")
	
	// Check if pollination is successful
	successChance := efficiency * flower.BloomingLevel * ips.SeasonalModifier
	
	if rand.Float64() > successChance {
		return false // Pollination failed
	}
	
	// Calculate nectar gained
	nectarGained := flower.NectarAmount * (0.1 + efficiency*0.2)
	nectarGained = math.Min(nectarGained, flower.NectarAmount) // Can't take more than available
	
	// Calculate pollen transfer
	pollenTransfer := flower.PollenLoad * pollenCapacity * efficiency
	pollenTransfer = math.Min(pollenTransfer, flower.PollenLoad)
	
	// Update flower resources
	flower.NectarAmount -= nectarGained
	flower.PollenLoad -= pollenTransfer
	flower.LastVisitTick = tick
	flower.VisitorCount++
	
	// Give energy to pollinator
	pollinator.Energy += nectarGained * flower.NectarQuality
	ips.NectarConsumed += nectarGained
	
	// Update pollinator memory
	ips.updatePollinatorMemory(pollinator, flower, nectarGained, tick)
	
	// Check for cross-pollination with carried pollen
	ips.tryTransferPollen(pollinator, flower, pollenTransfer, tick)
	
	return true
}

// updatePollinatorMemory updates a pollinator's memory of flower locations
func (ips *InsectPollinationSystem) updatePollinatorMemory(pollinator *Entity, flower *FlowerPatch, nectarGained float64, tick int) {
	memories := ips.PollinatorMemories[pollinator.ID]
	if memories == nil {
		memories = make([]*PollinatorMemory, 0)
	}
	
	// Look for existing memory of this location
	var existingMemory *PollinatorMemory
	for _, memory := range memories {
		distance := math.Sqrt(
			math.Pow(memory.Location.X-flower.Position.X, 2) +
			math.Pow(memory.Location.Y-flower.Position.Y, 2),
		)
		if distance < 3.0 && memory.PlantType == flower.PlantType {
			existingMemory = memory
			break
		}
	}
	
	if existingMemory != nil {
		// Update existing memory
		existingMemory.LastVisitTick = tick
		existingMemory.VisitCount++
		
		// Update success rate based on nectar gained
		successValue := nectarGained / 10.0 // Normalize
		existingMemory.SuccessRate = (existingMemory.SuccessRate*0.7) + (successValue*0.3)
		existingMemory.MemoryStrength = math.Min(1.0, existingMemory.MemoryStrength + 0.1)
		existingMemory.NectarQuality = (existingMemory.NectarQuality*0.8) + (flower.NectarQuality*0.2)
	} else {
		// Create new memory
		newMemory := &PollinatorMemory{
			Location:       flower.Position,
			PlantType:      flower.PlantType,
			LastVisitTick:  tick,
			NectarQuality:  flower.NectarQuality,
			SuccessRate:    nectarGained / 10.0,
			MemoryStrength: 0.5,
			VisitCount:     1,
		}
		
		memories = append(memories, newMemory)
		
		// Limit memory size based on flower memory trait
		maxMemories := int(pollinator.GetTrait("flower_memory") * 20) + 5
		if len(memories) > maxMemories {
			// Remove oldest/weakest memories
			memories = ips.pruneMemories(memories, maxMemories)
		}
	}
	
	ips.PollinatorMemories[pollinator.ID] = memories
}

// pruneMemories removes old or weak memories to maintain memory limit
func (ips *InsectPollinationSystem) pruneMemories(memories []*PollinatorMemory, maxSize int) []*PollinatorMemory {
	if len(memories) <= maxSize {
		return memories
	}
	
	// Sort by memory strength (keep strongest)
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[i].MemoryStrength < memories[j].MemoryStrength {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}
	
	return memories[:maxSize]
}

// tryTransferPollen attempts to transfer pollen between plants
func (ips *InsectPollinationSystem) tryTransferPollen(pollinator *Entity, flower *FlowerPatch, pollenTransfer float64, tick int) {
	// Pollinators carry pollen from previous visits
	carriedPollen := pollinator.GetTrait("carried_pollen")
	carriedPollenType := PlantType(pollinator.GetTrait("carried_pollen_type"))
	
	if carriedPollen > 0.1 && carriedPollenType != flower.PlantType {
		// Attempt cross-pollination
		crossPollinationSuccess := pollinator.GetTrait("pollination_efficiency") * 0.7
		
		if rand.Float64() < crossPollinationSuccess {
			// Record pollination event
			event := &PollinationEvent{
				PollinatorID:   pollinator.ID,
				SourcePlantID:  int(pollinator.GetTrait("carried_pollen_source")),
				TargetPlantID:  flower.PlantID,
				SourceType:     carriedPollenType,
				TargetType:     flower.PlantType,
				Efficiency:     crossPollinationSuccess,
				NectarGained:   0, // Already gained from flower visit
				PollenTransfer: carriedPollen,
				IsCrossSpecies: carriedPollenType != flower.PlantType,
				Tick:           tick,
			}
			
			ips.PollinationEvents = append(ips.PollinationEvents, event)
			ips.TotalPollinations++
			
			if carriedPollenType != flower.PlantType {
				ips.CrossSpeciesPollinations++
			}
			
			// Clear carried pollen
			pollinator.SetTrait("carried_pollen", 0)
			pollinator.SetTrait("carried_pollen_type", 0)
			pollinator.SetTrait("carried_pollen_source", 0)
		}
	}
	
	// Pollinator picks up new pollen
	if pollenTransfer > 0 {
		pollinator.SetTrait("carried_pollen", pollenTransfer)
		pollinator.SetTrait("carried_pollen_type", float64(flower.PlantType))
		pollinator.SetTrait("carried_pollen_source", float64(flower.PlantID))
	}
}

// UpdateFlowerPatches updates flower patch status and regeneration
func (ips *InsectPollinationSystem) UpdateFlowerPatches(plants []*Plant, season Season, tick int) {
	// Remove old patches and create new ones
	activePatches := make([]*FlowerPatch, 0)
	
	// Check existing patches
	for _, patch := range ips.FlowerPatches {
		// Find corresponding plant
		var plant *Plant
		for _, p := range plants {
			if p.ID == patch.PlantID {
				plant = p
				break
			}
		}
		
		if plant != nil && plant.IsAlive && plant.Age >= 10 {
			// Regenerate nectar
			patch.NectarAmount += patch.NectarRegenRate
			patch.NectarAmount = math.Min(patch.NectarAmount, plant.Energy*0.15) // Cap based on plant energy
			
			// Reduce visitor count over time
			if tick-patch.LastVisitTick > 10 {
				patch.VisitorCount = int(float64(patch.VisitorCount) * 0.8)
			}
			
			// Update bloom season compatibility
			if season == patch.BloomSeason || season == Spring {
				activePatches = append(activePatches, patch)
			}
		}
	}
	
	// Create new patches for plants that don't have them
	plantPatchMap := make(map[int]bool)
	for _, patch := range activePatches {
		plantPatchMap[patch.PlantID] = true
	}
	
	for _, plant := range plants {
		if !plantPatchMap[plant.ID] {
			if newPatch := ips.CreateFlowerPatch(plant, season); newPatch != nil {
				activePatches = append(activePatches, newPatch)
			}
		}
	}
	
	ips.FlowerPatches = activePatches
	
	// Update nectar production tracking
	nectarProduced := 0.0
	for _, patch := range ips.FlowerPatches {
		nectarProduced += patch.NectarAmount
	}
	ips.NectarProduced = nectarProduced
}

// UpdatePollinatorMemories decays old memories
func (ips *InsectPollinationSystem) UpdatePollinatorMemories(tick int) {
	for entityID, memories := range ips.PollinatorMemories {
		activeMemories := make([]*PollinatorMemory, 0)
		
		for _, memory := range memories {
			// Decay memory strength over time
			timeSinceVisit := tick - memory.LastVisitTick
			decayRate := 0.01 // 1% decay per tick
			
			memory.MemoryStrength -= float64(timeSinceVisit) * decayRate
			memory.MemoryStrength = math.Max(0, memory.MemoryStrength)
			
			// Keep memories that are still strong enough
			if memory.MemoryStrength > 0.1 {
				activeMemories = append(activeMemories, memory)
			}
		}
		
		if len(activeMemories) > 0 {
			ips.PollinatorMemories[entityID] = activeMemories
		} else {
			delete(ips.PollinatorMemories, entityID)
		}
	}
}

// GetPollinationStats returns statistics about the pollination system
func (ips *InsectPollinationSystem) GetPollinationStats() map[string]interface{} {
	activePatches := len(ips.FlowerPatches)
	activePollinators := len(ips.PollinatorMemories)
	recentEvents := 0
	
	// Count recent pollination events (last 100 ticks)
	for _, event := range ips.PollinationEvents {
		if len(ips.PollinationEvents) > 0 && event.Tick >= ips.PollinationEvents[len(ips.PollinationEvents)-1].Tick-100 {
			recentEvents++
		}
	}
	
	crossSpeciesRate := 0.0
	if ips.TotalPollinations > 0 {
		crossSpeciesRate = float64(ips.CrossSpeciesPollinations) / float64(ips.TotalPollinations)
	}
	
	return map[string]interface{}{
		"active_flower_patches":     activePatches,
		"active_pollinators":        activePollinators,
		"total_pollinations":        ips.TotalPollinations,
		"cross_species_pollinations": ips.CrossSpeciesPollinations,
		"cross_species_rate":        crossSpeciesRate,
		"nectar_produced":           ips.NectarProduced,
		"nectar_consumed":           ips.NectarConsumed,
		"recent_pollination_events": recentEvents,
		"seasonal_modifier":         ips.SeasonalModifier,
	}
}

// Update processes the pollination system for one tick
func (ips *InsectPollinationSystem) Update(entities []*Entity, plants []*Plant, season Season, tick int) {
	// Update seasonal modifier
	ips.updateSeasonalModifier(season)
	
	// Update flower patches
	ips.UpdateFlowerPatches(plants, season, tick)
	
	// Update pollinator memories
	ips.UpdatePollinatorMemories(tick)
	
	// Process pollinator behaviors
	for _, entity := range entities {
		if entity.IsAlive && ips.isEntityPollinator(entity) {
			ips.processPollinatorBehavior(entity, tick)
		}
	}
	
	// Limit event history to prevent memory issues
	if len(ips.PollinationEvents) > 1000 {
		ips.PollinationEvents = ips.PollinationEvents[len(ips.PollinationEvents)-800:]
	}
}

// updateSeasonalModifier adjusts pollination activity based on season
func (ips *InsectPollinationSystem) updateSeasonalModifier(season Season) {
	switch season {
	case Spring:
		ips.SeasonalModifier = 1.3 // High activity in spring
	case Summer:
		ips.SeasonalModifier = 1.0 // Normal activity in summer
	case Autumn:
		ips.SeasonalModifier = 0.6 // Reduced activity in autumn
	case Winter:
		ips.SeasonalModifier = 0.2 // Very low activity in winter
	default:
		ips.SeasonalModifier = 1.0
	}
}

// isEntityPollinator checks if an entity has pollinator capabilities
func (ips *InsectPollinationSystem) isEntityPollinator(entity *Entity) bool {
	return entity.GetTrait("pollination_efficiency") > 0.1 && 
		   entity.GetTrait("flying_ability") > 0.2 &&
		   entity.GetTrait("nectar_detection") > 0.1
}

// processPollinatorBehavior handles pollinator decision-making and actions
func (ips *InsectPollinationSystem) processPollinatorBehavior(pollinator *Entity, tick int) {
	flightRange := pollinator.GetTrait("flight_range")
	nectarNeeds := pollinator.GetTrait("nectar_needs")
	seasonalActivity := pollinator.GetTrait("seasonal_activity")
	
	// Adjust activity based on seasonal modifier
	if rand.Float64() > seasonalActivity*ips.SeasonalModifier {
		return // Not active this tick
	}
	
	// Check if pollinator needs nectar
	energyRatio := pollinator.Energy / pollinator.GetTrait("endurance")
	needsNectar := energyRatio < nectarNeeds*2.0
	
	if needsNectar || rand.Float64() < 0.3 { // 30% chance of foraging even when not needy
		// Look for nearby flowers
		nearbyFlowers := ips.FindNearbyFlowers(pollinator, flightRange)
		
		if len(nearbyFlowers) > 0 {
			// Select best flower based on preferences and memory
			targetFlower := ips.SelectBestFlower(pollinator, nearbyFlowers)
			
			if targetFlower != nil {
				// Move toward flower
				ips.moveTowardFlower(pollinator, targetFlower)
				
				// Check if close enough to pollinate
				distance := math.Sqrt(
					math.Pow(targetFlower.Position.X-pollinator.Position.X, 2) +
					math.Pow(targetFlower.Position.Y-pollinator.Position.Y, 2),
				)
				
				if distance < 2.0 { // Close enough to interact
					ips.AttemptPollination(pollinator, targetFlower, tick)
				}
			}
		} else {
			// No flowers nearby, search for new areas
			ips.searchForFlowers(pollinator)
		}
	}
}

// moveTowardFlower moves a pollinator toward a target flower
func (ips *InsectPollinationSystem) moveTowardFlower(pollinator *Entity, flower *FlowerPatch) {
	dx := flower.Position.X - pollinator.Position.X
	dy := flower.Position.Y - pollinator.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	
	if distance > 0.5 {
		speed := pollinator.GetTrait("speed") * pollinator.GetTrait("flying_ability")
		
		// Normalize direction and apply speed
		pollinator.Position.X += (dx / distance) * speed
		pollinator.Position.Y += (dy / distance) * speed
	}
}

// searchForFlowers makes a pollinator explore for new flower patches
func (ips *InsectPollinationSystem) searchForFlowers(pollinator *Entity) {
	// Use memory to guide search
	memories := ips.PollinatorMemories[pollinator.ID]
	
	if len(memories) > 0 && rand.Float64() < 0.7 { // 70% chance to use memory
		// Move toward remembered location
		bestMemory := memories[0]
		for _, memory := range memories {
			if memory.MemoryStrength > bestMemory.MemoryStrength {
				bestMemory = memory
			}
		}
		
		dx := bestMemory.Location.X - pollinator.Position.X
		dy := bestMemory.Location.Y - pollinator.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		
		if distance > 1.0 {
			speed := pollinator.GetTrait("speed") * 0.5 // Slower exploratory movement
			pollinator.Position.X += (dx / distance) * speed
			pollinator.Position.Y += (dy / distance) * speed
		}
	} else {
		// Random exploration
		angle := rand.Float64() * 2 * math.Pi
		speed := pollinator.GetTrait("speed") * 0.3
		
		pollinator.Position.X += math.Cos(angle) * speed
		pollinator.Position.Y += math.Sin(angle) * speed
	}
}