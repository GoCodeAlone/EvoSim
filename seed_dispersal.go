package main

import (
	"fmt"
	"math"
	"math/rand"
)

// DispersalMechanism represents different ways seeds can be dispersed
type DispersalMechanism int

const (
	DispersalWind DispersalMechanism = iota // Wind-based dispersal (existing)
	DispersalAnimal                         // Animal-mediated dispersal
	DispersalExplosive                      // Explosive/ballistic dispersal
	DispersalGravity                        // Gravity-based dispersal
	DispersalWater                          // Water-based dispersal (basic)
)

// SeedType represents different seed characteristics
type SeedType int

const (
	SeedSmallLight SeedType = iota // Small, light seeds (wind dispersal)
	SeedLargeHeavy                 // Large, heavy seeds (gravity/animal)
	SeedHooked                     // Seeds with hooks (animal attachment)
	SeedFleshy                     // Fleshy seeds (animal consumption)
	SeedExplosive                  // Seeds in explosive pods
	SeedFloating                   // Seeds that float on water
)

// Seed represents individual seed particles
type Seed struct {
	ID               int                    `json:"id"`
	Position         Position               `json:"position"`
	Velocity         Vector2D               `json:"velocity"`
	PlantID          int                    `json:"plant_id"`     // Source plant ID
	PlantType        PlantType              `json:"plant_type"`   // Parent plant type
	SeedType         SeedType               `json:"seed_type"`    // Seed characteristics
	Genetics         map[string]Trait       `json:"genetics"`     // Genetic material from parent
	Viability        float64                `json:"viability"`    // How viable the seed is (0.0 to 1.0)
	Age              int                    `json:"age"`          // Ticks since dispersal
	MaxAge           int                    `json:"max_age"`      // Maximum viability age
	Size             float64                `json:"size"`         // Affects dispersal mechanics
	Mass             float64                `json:"mass"`         // Affects gravity/wind resistance
	DispersalMethod  DispersalMechanism     `json:"dispersal_method"`
	IsDormant        bool                   `json:"is_dormant"`   // Waiting for germination conditions
	CarriedByEntity  int                    `json:"carried_by_entity"` // ID of entity carrying seed (0 if none)
	DistanceFromHome float64                `json:"distance_from_home"` // Distance from parent plant
	
	// Dormancy and germination factors
	RequiredTemperature float64 `json:"required_temperature"`
	RequiredMoisture    float64 `json:"required_moisture"`
	RequiredSunlight    float64 `json:"required_sunlight"`
	DormancyTrigger     bool    `json:"dormancy_trigger"` // Whether seed entered dormancy
}

// SeedBank represents accumulated seeds in soil waiting to germinate
type SeedBank struct {
	Position Position           `json:"position"`
	Seeds    []*Seed           `json:"seeds"`
	Capacity int               `json:"capacity"`        // Maximum seeds that can be stored
	Depth    float64           `json:"depth"`           // How deep seeds are buried
	Moisture float64           `json:"moisture"`        // Soil moisture level
	Nutrients map[string]float64 `json:"nutrients"`     // Available soil nutrients
}

// SeedDispersalSystem manages all seed dispersal mechanics
type SeedDispersalSystem struct {
	AllSeeds       []*Seed                     `json:"all_seeds"`
	SeedBanks      map[Position]*SeedBank      `json:"seed_banks"`
	NextSeedID     int                         `json:"next_seed_id"`
	DispersalStats map[DispersalMechanism]int  `json:"dispersal_stats"`
	GerminationEvents int                      `json:"germination_events"`
	DormancyActivations int                    `json:"dormancy_activations"`
}

// NewSeedDispersalSystem creates a new seed dispersal system
func NewSeedDispersalSystem() *SeedDispersalSystem {
	return &SeedDispersalSystem{
		AllSeeds:        make([]*Seed, 0),
		SeedBanks:       make(map[Position]*SeedBank),
		NextSeedID:      1,
		DispersalStats:  make(map[DispersalMechanism]int),
		GerminationEvents: 0,
		DormancyActivations: 0,
	}
}

// CreateSeed creates a new seed from a parent plant
func (sds *SeedDispersalSystem) CreateSeed(parent *Plant, world *World) *Seed {
	config := GetPlantConfigs()[parent.Type]
	
	// Determine seed type based on plant type and traits
	seedType := sds.determineSeedType(parent)
	dispersalMethod := sds.determineDispersalMethod(seedType, parent, world)
	
	seed := &Seed{
		ID:               sds.NextSeedID,
		Position:         parent.Position,
		Velocity:         Vector2D{X: 0, Y: 0},
		PlantID:          parent.ID,
		PlantType:        parent.Type,
		SeedType:         seedType,
		Genetics:         parent.Traits,
		Viability:        1.0,
		Age:              0,
		MaxAge:           config.MaxAge / 10, // Seeds viable for 1/10th of plant lifespan
		Size:             parent.Size * 0.1,  // Seeds are much smaller than parent
		Mass:             sds.calculateSeedMass(seedType, parent.Size),
		DispersalMethod:  dispersalMethod,
		IsDormant:        false,
		CarriedByEntity:  0,
		DistanceFromHome: 0,
		RequiredTemperature: 15.0 + rand.Float64()*10.0, // Temperature requirements
		RequiredMoisture:    0.3 + rand.Float64()*0.4,    // Moisture requirements  
		RequiredSunlight:    0.2 + rand.Float64()*0.6,    // Sunlight requirements
		DormancyTrigger:     false,
	}
	
	sds.NextSeedID++
	sds.AllSeeds = append(sds.AllSeeds, seed)
	sds.DispersalStats[dispersalMethod]++
	
	return seed
}

// determineSeedType determines seed characteristics based on plant traits
func (sds *SeedDispersalSystem) determineSeedType(parent *Plant) SeedType {
	switch parent.Type {
	case PlantGrass:
		return SeedSmallLight // Grass seeds are typically small and light
	case PlantBush:
		if parent.GetTrait("defense") > 0.3 {
			return SeedHooked // Defensive bushes might have hooked seeds
		}
		return SeedFleshy // Many bush fruits are animal-dispersed
	case PlantTree:
		if parent.Size > 2.5 {
			return SeedLargeHeavy // Large trees have heavy seeds/nuts
		}
		return SeedFleshy // Many tree fruits are animal-dispersed
	case PlantMushroom:
		return SeedSmallLight // Spores are tiny and wind-dispersed
	case PlantAlgae:
		return SeedFloating // Aquatic dispersal
	case PlantCactus:
		if rand.Float64() < 0.6 {
			return SeedFleshy // Cactus fruits are often fleshy
		}
		return SeedExplosive // Some cacti have explosive seed pods
	case PlantLily:
		return SeedFloating // Water lily seeds float on water surface
	case PlantReed:
		return SeedFloating // Reed seeds disperse via water currents
	case PlantKelp:
		return SeedFloating // Kelp spores drift with water currents
	default:
		return SeedSmallLight
	}
}

// determineDispersalMethod determines how seeds will be dispersed
func (sds *SeedDispersalSystem) determineDispersalMethod(seedType SeedType, parent *Plant, world *World) DispersalMechanism {
	// Check environmental factors
	windStrength := 0.0
	if world.WindSystem != nil && len(world.WindSystem.WindMap) > 0 && len(world.WindSystem.WindMap[0]) > 0 {
		// Get wind strength at plant location with bounds checking
		gridX := int(parent.Position.X)
		gridY := int(parent.Position.Y)
		
		// Handle negative coordinates and ensure positive modulo
		if gridX < 0 {
			gridX = ((gridX % len(world.WindSystem.WindMap)) + len(world.WindSystem.WindMap)) % len(world.WindSystem.WindMap)
		} else {
			gridX = gridX % len(world.WindSystem.WindMap)
		}
		
		if gridY < 0 {
			gridY = ((gridY % len(world.WindSystem.WindMap[0])) + len(world.WindSystem.WindMap[0])) % len(world.WindSystem.WindMap[0])
		} else {
			gridY = gridY % len(world.WindSystem.WindMap[0])
		}
		
		windStrength = world.WindSystem.WindMap[gridX][gridY].Strength
	}
	
	switch seedType {
	case SeedSmallLight:
		return DispersalWind // Always wind-dispersed
	case SeedFloating:
		return DispersalWater // Always water-dispersed
	case SeedExplosive:
		return DispersalExplosive // Always explosive
	case SeedLargeHeavy:
		if sds.hasNearbyAnimals(parent.Position, world) {
			return DispersalAnimal // Animal-mediated if animals nearby
		}
		return DispersalGravity // Otherwise gravity
	case SeedHooked, SeedFleshy:
		if sds.hasNearbyAnimals(parent.Position, world) {
			return DispersalAnimal // Animal-mediated if animals nearby
		}
		if windStrength > 0.3 {
			return DispersalWind // Wind if strong enough
		}
		return DispersalGravity // Default to gravity
	default:
		return DispersalGravity
	}
}

// hasNearbyAnimals checks if there are entities nearby that could disperse seeds
func (sds *SeedDispersalSystem) hasNearbyAnimals(position Position, world *World) bool {
	searchRadius := 5.0
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			distance := math.Sqrt(math.Pow(entity.Position.X-position.X, 2) + 
				math.Pow(entity.Position.Y-position.Y, 2))
			if distance <= searchRadius {
				return true
			}
		}
	}
	return false
}

// calculateSeedMass calculates seed mass based on type and parent size
func (sds *SeedDispersalSystem) calculateSeedMass(seedType SeedType, parentSize float64) float64 {
	baseMass := parentSize * 0.05 // Base mass relative to parent
	
	switch seedType {
	case SeedSmallLight:
		return baseMass * 0.1 // Very light
	case SeedLargeHeavy:
		return baseMass * 3.0 // Heavy
	case SeedHooked:
		return baseMass * 0.5 // Medium-light
	case SeedFleshy:
		return baseMass * 1.5 // Medium-heavy  
	case SeedExplosive:
		return baseMass * 0.8 // Medium
	case SeedFloating:
		return baseMass * 0.3 // Light but buoyant
	default:
		return baseMass
	}
}

// Update updates all seeds and handles dispersal mechanics
func (sds *SeedDispersalSystem) Update(world *World) {
	// Update existing seeds
	for i := len(sds.AllSeeds) - 1; i >= 0; i-- {
		seed := sds.AllSeeds[i]
		sds.updateSeed(seed, world)
		
		// Remove dead/germinated seeds
		if seed.Age > seed.MaxAge || seed.Viability <= 0 {
			sds.AllSeeds = append(sds.AllSeeds[:i], sds.AllSeeds[i+1:]...)
		}
	}
	
	// Update seed banks
	sds.updateSeedBanks(world)
}

// updateSeed updates individual seed behavior
func (sds *SeedDispersalSystem) updateSeed(seed *Seed, world *World) {
	seed.Age++
	
	// Check if seed should enter dormancy based on conditions
	if !seed.IsDormant && sds.shouldEnterDormancy(seed, world) {
		seed.IsDormant = true
		seed.DormancyTrigger = true
		sds.DormancyActivations++
		sds.addToSeedBank(seed, world)
		return
	}
	
	// Check if dormant seed should germinate
	if seed.IsDormant && sds.canGerminate(seed, world) {
		sds.germinate(seed, world)
		return
	}
	
	// Handle dispersal based on method
	switch seed.DispersalMethod {
	case DispersalWind:
		sds.disperseByWind(seed, world)
	case DispersalAnimal:
		sds.disperseByAnimal(seed, world)
	case DispersalExplosive:
		sds.disperseExplosively(seed, world)
	case DispersalGravity:
		sds.disperseByGravity(seed, world)
	case DispersalWater:
		sds.disperseByWater(seed, world)
	}
	
	// Update distance from home
	parentPos := seed.Position // We'd need to track original position
	seed.DistanceFromHome = math.Sqrt(math.Pow(seed.Position.X-parentPos.X, 2) + 
		math.Pow(seed.Position.Y-parentPos.Y, 2))
	
	// Reduce viability over time
	seed.Viability -= 0.01 // 1% per tick
}

// shouldEnterDormancy checks if environmental conditions trigger dormancy
func (sds *SeedDispersalSystem) shouldEnterDormancy(seed *Seed, world *World) bool {
	// Get environmental conditions at seed location
	temperature := world.getTemperatureAt(seed.Position)
	moisture := world.getMoistureAt(seed.Position)
	
	// Seeds enter dormancy if conditions aren't right for germination
	if temperature < seed.RequiredTemperature-5 || temperature > seed.RequiredTemperature+15 {
		return true
	}
	if moisture < seed.RequiredMoisture-0.2 {
		return true
	}
	
	// Random chance of dormancy to prevent immediate germination
	return rand.Float64() < 0.3
}

// canGerminate checks if conditions are suitable for germination
func (sds *SeedDispersalSystem) canGerminate(seed *Seed, world *World) bool {
	if !seed.IsDormant {
		return false
	}
	
	// Get environmental conditions
	temperature := world.getTemperatureAt(seed.Position)
	moisture := world.getMoistureAt(seed.Position)
	sunlight := world.getSunlightAt(seed.Position)
	
	// Check if all conditions are met
	tempOK := temperature >= seed.RequiredTemperature && temperature <= seed.RequiredTemperature+10
	moistureOK := moisture >= seed.RequiredMoisture
	sunlightOK := sunlight >= seed.RequiredSunlight
	
	return tempOK && moistureOK && sunlightOK && seed.Viability > 0.1
}

// germinate converts a seed into a new plant
func (sds *SeedDispersalSystem) germinate(seed *Seed, world *World) {
	// Create new plant at seed location
	newPlant := NewPlant(world.NextPlantID, seed.PlantType, seed.Position)
	world.NextPlantID++
	
	// Inherit genetics from seed
	newPlant.Traits = seed.Genetics
	
	// Add some genetic variation
	for traitName, trait := range newPlant.Traits {
		trait.Value += (rand.Float64()*2 - 1) * 0.1 // Small mutations
		newPlant.Traits[traitName] = trait
	}
	
	// Add to world
	world.AllPlants = append(world.AllPlants, newPlant)
	
	// Log germination event
	if world.CentralEventBus != nil {
		world.CentralEventBus.EmitPlantEvent(world.Tick, "germination", "seed_germination", 
			"plant_lifecycle", fmt.Sprintf("Seed %d germinated into plant %d at distance %.1f from parent", 
			seed.ID, newPlant.ID, seed.DistanceFromHome), newPlant, false, true)
	}
	
	sds.GerminationEvents++
	
	// Mark seed as used
	seed.Viability = 0
}

// addToSeedBank adds a seed to the local seed bank
func (sds *SeedDispersalSystem) addToSeedBank(seed *Seed, world *World) {
	// Round position to grid coordinates for seed bank
	bankPos := Position{
		X: math.Round(seed.Position.X),
		Y: math.Round(seed.Position.Y),
	}
	
	// Get or create seed bank
	if sds.SeedBanks[bankPos] == nil {
		sds.SeedBanks[bankPos] = &SeedBank{
			Position:  bankPos,
			Seeds:     make([]*Seed, 0),
			Capacity:  20, // Maximum seeds per location
			Depth:     1.0,
			Moisture:  world.getMoistureAt(bankPos),
			Nutrients: make(map[string]float64),
		}
	}
	
	bank := sds.SeedBanks[bankPos]
	if len(bank.Seeds) < bank.Capacity {
		bank.Seeds = append(bank.Seeds, seed)
	}
}

// updateSeedBanks updates all seed banks
func (sds *SeedDispersalSystem) updateSeedBanks(world *World) {
	for pos, bank := range sds.SeedBanks {
		// Update environmental conditions
		bank.Moisture = world.getMoistureAt(pos)
		
		// Check seeds for germination
		for i := len(bank.Seeds) - 1; i >= 0; i-- {
			seed := bank.Seeds[i]
			if sds.canGerminate(seed, world) {
				sds.germinate(seed, world)
				// Remove from seed bank
				bank.Seeds = append(bank.Seeds[:i], bank.Seeds[i+1:]...)
			}
		}
		
		// Remove empty seed banks
		if len(bank.Seeds) == 0 {
			delete(sds.SeedBanks, pos)
		}
	}
}

// Dispersal method implementations

// disperseByWind moves seeds based on wind patterns
func (sds *SeedDispersalSystem) disperseByWind(seed *Seed, world *World) {
	if world.WindSystem == nil || len(world.WindSystem.WindMap) == 0 {
		return
	}
	
	// Get wind at seed location with proper bounds checking
	mapWidth := len(world.WindSystem.WindMap)
	mapHeight := len(world.WindSystem.WindMap[0])
	
	// Ensure coordinates are within bounds
	gridX := int(seed.Position.X)
	if gridX < 0 {
		gridX = 0
	} else if gridX >= mapWidth {
		gridX = mapWidth - 1
	}
	
	gridY := int(seed.Position.Y)
	if gridY < 0 {
		gridY = 0
	} else if gridY >= mapHeight {
		gridY = mapHeight - 1
	}
	
	wind := world.WindSystem.WindMap[gridX][gridY]
	
	// Apply wind force (lighter seeds travel further)
	windEffect := wind.Strength / seed.Mass
	seed.Velocity.X += wind.X * windEffect * 0.1
	seed.Velocity.Y += wind.Y * windEffect * 0.1
	
	// Add some turbulence
	seed.Velocity.X += (rand.Float64()*2 - 1) * wind.Turbulence * 0.05
	seed.Velocity.Y += (rand.Float64()*2 - 1) * wind.Turbulence * 0.05
	
	// Apply drag
	seed.Velocity.X *= 0.9
	seed.Velocity.Y *= 0.9
	
	// Update position
	seed.Position.X += seed.Velocity.X
	seed.Position.Y += seed.Velocity.Y
}

// disperseByAnimal handles animal-mediated dispersal
func (sds *SeedDispersalSystem) disperseByAnimal(seed *Seed, world *World) {
	// If not currently carried, try to attach to nearby entity
	if seed.CarriedByEntity == 0 {
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				distance := math.Sqrt(math.Pow(entity.Position.X-seed.Position.X, 2) + 
					math.Pow(entity.Position.Y-seed.Position.Y, 2))
				if distance <= 1.0 { // Close enough to pick up
					// Probability of pickup based on seed type
					pickupChance := 0.1 // Base chance
					if seed.SeedType == SeedFleshy {
						pickupChance = 0.3 // Animals more likely to eat fleshy seeds
					} else if seed.SeedType == SeedHooked {
						pickupChance = 0.2 // Hooked seeds attach more easily
					}
					
					if rand.Float64() < pickupChance {
						seed.CarriedByEntity = entity.ID
						seed.Position = entity.Position
						break
					}
				}
			}
		}
	} else {
		// If carried, follow the entity
		for _, entity := range world.AllEntities {
			if entity.ID == seed.CarriedByEntity && entity.IsAlive {
				seed.Position = entity.Position
				
				// Random chance of being dropped
				dropChance := 0.05 // 5% chance per tick
				if seed.SeedType == SeedFleshy {
					dropChance = 0.1 // Eaten seeds are "dropped" more often
				}
				
				if rand.Float64() < dropChance {
					seed.CarriedByEntity = 0
					// Add some randomness to drop location
					seed.Position.X += (rand.Float64()*2 - 1) * 2.0
					seed.Position.Y += (rand.Float64()*2 - 1) * 2.0
				}
				break
			}
		}
		
		// If entity is dead, drop the seed
		entityFound := false
		for _, entity := range world.AllEntities {
			if entity.ID == seed.CarriedByEntity {
				entityFound = true
				break
			}
		}
		if !entityFound {
			seed.CarriedByEntity = 0
		}
	}
}

// disperseExplosively handles explosive dispersal
func (sds *SeedDispersalSystem) disperseExplosively(seed *Seed, world *World) {
	// Explosive dispersal happens quickly
	if seed.Age == 1 { // First tick after creation
		// Launch seed in random direction
		angle := rand.Float64() * 2 * math.Pi
		force := 2.0 + rand.Float64()*3.0 // Random force 2-5
		
		seed.Velocity.X = math.Cos(angle) * force
		seed.Velocity.Y = math.Sin(angle) * force
	}
	
	// Apply gravity and air resistance
	seed.Velocity.Y -= 0.1 // Gravity
	seed.Velocity.X *= 0.95 // Air resistance
	seed.Velocity.Y *= 0.95
	
	// Update position
	seed.Position.X += seed.Velocity.X
	seed.Position.Y += seed.Velocity.Y
	
	// Stop when hitting ground (velocity near zero)
	if math.Abs(seed.Velocity.X) < 0.1 && math.Abs(seed.Velocity.Y) < 0.1 {
		seed.Velocity.X = 0
		seed.Velocity.Y = 0
	}
}

// disperseByGravity handles simple gravity-based dispersal
func (sds *SeedDispersalSystem) disperseByGravity(seed *Seed, world *World) {
	// Heavy seeds just fall down with some randomness
	if seed.Age == 1 { // First tick after creation
		// Small random horizontal movement
		seed.Velocity.X = (rand.Float64()*2 - 1) * 0.5
		seed.Velocity.Y = -(0.5 + rand.Float64()*1.0) // Downward movement
	}
	
	// Apply gravity
	seed.Velocity.Y -= 0.05
	
	// Update position
	seed.Position.X += seed.Velocity.X
	seed.Position.Y += seed.Velocity.Y
	
	// Stop when hitting ground
	if seed.Position.Y <= 0 || math.Abs(seed.Velocity.Y) < 0.05 {
		seed.Velocity.X = 0
		seed.Velocity.Y = 0
	}
}

// disperseByWater handles water-based dispersal with realistic flow mechanics
func (sds *SeedDispersalSystem) disperseByWater(seed *Seed, world *World) {
	biome := world.getBiomeAt(seed.Position)
	
	// Enhanced water flow patterns based on biome type
	switch biome {
	case BiomeWater:
		// Shallow water - moderate flow with surface currents
		flowDirection := sds.calculateWaterFlow(seed.Position, world, false)
		flowStrength := 0.15
		
		// Surface currents affected by wind
		windInfluence := sds.getWindInfluenceOnWater(seed.Position, world)
		
		seed.Velocity.X += flowDirection.X * flowStrength + windInfluence.X * 0.05
		seed.Velocity.Y += flowDirection.Y * flowStrength + windInfluence.Y * 0.05
		
	case BiomeDeepWater:
		// Deep water - stronger, more consistent currents
		flowDirection := sds.calculateWaterFlow(seed.Position, world, true)
		flowStrength := 0.25
		
		seed.Velocity.X += flowDirection.X * flowStrength
		seed.Velocity.Y += flowDirection.Y * flowStrength
		
	case BiomeSwamp:
		// Swamp water - slow, stagnant with occasional flow
		flowDirection := sds.calculateWaterFlow(seed.Position, world, false)
		flowStrength := 0.08 // Much slower flow
		
		// Random vegetation obstacles
		if rand.Float64() < 0.3 {
			flowStrength *= 0.5 // Vegetation slows flow
		}
		
		seed.Velocity.X += flowDirection.X * flowStrength
		seed.Velocity.Y += flowDirection.Y * flowStrength
		
	default:
		// Not in water - check for seasonal flooding
		if sds.isFloodedArea(seed.Position, world) {
			// Temporary water flow during floods
			flowDirection := sds.calculateFloodFlow(seed.Position, world)
			flowStrength := 0.3 // Strong flood currents
			
			seed.Velocity.X += flowDirection.X * flowStrength
			seed.Velocity.Y += flowDirection.Y * flowStrength
		} else {
			// Out of water, seeds settle
			seed.Velocity.X *= 0.7 // Friction with ground
			seed.Velocity.Y *= 0.7
			
			// Stop if velocity is very low
			if math.Abs(seed.Velocity.X) < 0.01 && math.Abs(seed.Velocity.Y) < 0.01 {
				seed.Velocity.X = 0
				seed.Velocity.Y = 0
			}
		}
	}
	
	// Apply water resistance (seeds slow down in water)
	seed.Velocity.X *= 0.95
	seed.Velocity.Y *= 0.95
	
	// Update position
	seed.Position.X += seed.Velocity.X
	seed.Position.Y += seed.Velocity.Y
	
	// Floating seeds have enhanced survival in water
	if seed.SeedType == SeedFloating && (biome == BiomeWater || biome == BiomeDeepWater || biome == BiomeSwamp) {
		seed.Viability = math.Min(1.0, seed.Viability + 0.001) // Slight recovery in water
	}
}

// calculateWaterFlow determines water flow direction based on terrain and currents
func (sds *SeedDispersalSystem) calculateWaterFlow(position Position, world *World, isDeepWater bool) Vector2D {
	// Calculate flow based on local topography and nearby water bodies
	flowVector := Vector2D{}
	
	// Check elevation differences in surrounding area
	currentElevation := world.getElevationAt(position)
	checkRadius := 3.0
	
	for dx := -checkRadius; dx <= checkRadius; dx += 1.0 {
		for dy := -checkRadius; dy <= checkRadius; dy += 1.0 {
			if dx == 0 && dy == 0 {
				continue
			}
			
			checkPos := Position{
				X: position.X + dx,
				Y: position.Y + dy,
			}
			
			if world.isValidPosition(checkPos) {
				elevation := world.getElevationAt(checkPos)
				
				// Water flows from high to low elevation
				elevationDiff := currentElevation - elevation
				if elevationDiff > 0 {
					distance := math.Sqrt(dx*dx + dy*dy)
					strength := elevationDiff / distance
					
					flowVector.X += (dx / distance) * strength
					flowVector.Y += (dy / distance) * strength
				}
			}
		}
	}
	
	// Add some randomness for natural water turbulence
	turbulence := 0.2
	if isDeepWater {
		turbulence = 0.1 // Deep water has less surface turbulence
	}
	
	flowVector.X += (rand.Float64()*2 - 1) * turbulence
	flowVector.Y += (rand.Float64()*2 - 1) * turbulence
	
	// Normalize the flow vector
	magnitude := math.Sqrt(flowVector.X*flowVector.X + flowVector.Y*flowVector.Y)
	if magnitude > 0 {
		flowVector.X /= magnitude
		flowVector.Y /= magnitude
	}
	
	return flowVector
}

// getWindInfluenceOnWater calculates how wind affects surface water currents
func (sds *SeedDispersalSystem) getWindInfluenceOnWater(position Position, world *World) Vector2D {
	windInfluence := Vector2D{}
	
	if world.WindSystem != nil {
		// Get wind vector at position
		gridX := int(position.X)
		gridY := int(position.Y)
		
		if gridX >= 0 && gridX < len(world.WindSystem.WindMap[0]) && 
		   gridY >= 0 && gridY < len(world.WindSystem.WindMap) {
			windVector := world.WindSystem.WindMap[gridY][gridX]
			
			// Wind creates surface currents (reduced effect underwater)
			// Calculate normalized direction from wind vector
			magnitude := math.Sqrt(windVector.X*windVector.X + windVector.Y*windVector.Y)
			if magnitude > 0 {
				windInfluence.X = (windVector.X / magnitude) * windVector.Strength * 0.3
				windInfluence.Y = (windVector.Y / magnitude) * windVector.Strength * 0.3
			}
		}
	}
	
	return windInfluence
}

// isFloodedArea checks if the position is temporarily flooded (seasonal floods)
func (sds *SeedDispersalSystem) isFloodedArea(position Position, world *World) bool {
	// Check if it's flood season (spring/early summer)
	season := world.AdvancedTimeSystem.Season
	if season != Spring && season != Summer {
		return false
	}
	
	// Check if position is near water bodies
	waterCheckRadius := 8.0
	for dx := -waterCheckRadius; dx <= waterCheckRadius; dx += 1.0 {
		for dy := -waterCheckRadius; dy <= waterCheckRadius; dy += 1.0 {
			checkPos := Position{
				X: position.X + dx,
				Y: position.Y + dy,
			}
			
			if world.isValidPosition(checkPos) {
				biome := world.getBiomeAt(checkPos)
				if biome == BiomeWater || biome == BiomeDeepWater || biome == BiomeSwamp {
					distance := math.Sqrt(dx*dx + dy*dy)
					
					// Flood probability decreases with distance from water
					floodProbability := math.Max(0, 1.0 - (distance/waterCheckRadius))
					
					// Random chance based on distance and season intensity
					if rand.Float64() < floodProbability * 0.3 {
						return true
					}
				}
			}
		}
	}
	
	return false
}

// calculateFloodFlow determines water flow direction during floods
func (sds *SeedDispersalSystem) calculateFloodFlow(position Position, world *World) Vector2D {
	flowVector := Vector2D{}
	
	// Find direction to nearest major water body
	waterCheckRadius := 15.0
	nearestWaterDistance := math.Inf(1)
	
	for dx := -waterCheckRadius; dx <= waterCheckRadius; dx += 2.0 {
		for dy := -waterCheckRadius; dy <= waterCheckRadius; dy += 2.0 {
			checkPos := Position{
				X: position.X + dx,
				Y: position.Y + dy,
			}
			
			if world.isValidPosition(checkPos) {
				biome := world.getBiomeAt(checkPos)
				if biome == BiomeWater || biome == BiomeDeepWater {
					distance := math.Sqrt(dx*dx + dy*dy)
					
					if distance < nearestWaterDistance {
						nearestWaterDistance = distance
						
						// Flow towards the water body
						flowVector.X = dx / distance
						flowVector.Y = dy / distance
					}
				}
			}
		}
	}
	
	// Add strong turbulence for flood conditions
	flowVector.X += (rand.Float64()*2 - 1) * 0.4
	flowVector.Y += (rand.Float64()*2 - 1) * 0.4
	
	return flowVector
}

// GetStats returns dispersal statistics
func (sds *SeedDispersalSystem) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_seeds"] = len(sds.AllSeeds)
	stats["total_seed_banks"] = len(sds.SeedBanks)
	stats["germination_events"] = sds.GerminationEvents
	stats["dormancy_activations"] = sds.DormancyActivations
	
	// Seeds by dispersal method
	for method, count := range sds.DispersalStats {
		methodName := ""
		switch method {
		case DispersalWind:
			methodName = "wind"
		case DispersalAnimal:
			methodName = "animal"
		case DispersalExplosive:
			methodName = "explosive"
		case DispersalGravity:
			methodName = "gravity"
		case DispersalWater:
			methodName = "water"
		}
		stats["dispersal_"+methodName] = count
	}
	
	// Active seeds by type
	seedCounts := make(map[string]int)
	for _, seed := range sds.AllSeeds {
		typeName := ""
		switch seed.SeedType {
		case SeedSmallLight:
			typeName = "small_light"
		case SeedLargeHeavy:
			typeName = "large_heavy"
		case SeedHooked:
			typeName = "hooked"
		case SeedFleshy:
			typeName = "fleshy"
		case SeedExplosive:
			typeName = "explosive"
		case SeedFloating:
			typeName = "floating"
		}
		seedCounts[typeName]++
	}
	stats["active_seeds_by_type"] = seedCounts
	
	return stats
}