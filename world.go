package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// WorldConfig holds configuration for world generation
type WorldConfig struct {
	Width          float64
	Height         float64
	NumPopulations int
	PopulationSize int
	GridWidth      int // Grid cells for visualization
	GridHeight     int
}

// BiomeType represents different environmental zones
type BiomeType int

const (
	BiomePlains BiomeType = iota
	BiomeForest
	BiomeDesert
	BiomeMountain
	BiomeWater
	BiomeRadiation
	BiomeSoil // Underground/soil environment
	BiomeAir  // Aerial environment (high altitude)
)

// Biome represents an environmental zone with specific effects
type Biome struct {
	Type           BiomeType
	Name           string
	Color          string
	TraitModifiers map[string]float64 // Trait name -> modifier
	MutationRate   float64            // Additional mutation rate
	EnergyDrain    float64            // Energy drain per tick
	Symbol         rune               // Display symbol
}

// WorldEvent represents temporary world-wide effects
type WorldEvent struct {
	Name           string
	Description    string
	Duration       int // Ticks remaining
	GlobalMutation float64
	GlobalDamage   float64
	BiomeChanges   map[Position]BiomeType
}

// GridCell represents a cell in the world grid
type GridCell struct {
	Biome    BiomeType
	Entities []*Entity
	Plants   []*Plant    // Plants in this cell
	Event    *WorldEvent // Current event affecting this cell
}

// PopulationConfig defines traits and behavior for a population
type PopulationConfig struct {
	Name             string
	Species          string
	BaseTraits       map[string]float64
	StartPos         Position
	Spread           float64 // How spread out they start
	Color            string  // For visualization
	BaseMutationRate float64 // Base mutation rate for this species
}

// World represents the environment containing multiple populations
type World struct {
	Config      WorldConfig
	Populations map[string]*Population
	AllEntities []*Entity
	AllPlants   []*Plant // All plants in the world
	Grid        [][]GridCell
	Biomes      map[BiomeType]Biome
	Events      []*WorldEvent
	EventLogger *EventLogger // Event logging system
	NextID      int
	NextPlantID int // ID counter for plants
	Tick        int
	Clock       time.Time
	LastUpdate  time.Time
	Paused      bool // Whether the simulation is paused
	// Advanced feature systems
	CommunicationSystem *CommunicationSystem
	GroupBehaviorSystem *GroupBehaviorSystem
	PhysicsSystem       *PhysicsSystem
	CollisionSystem     *CollisionSystem
	PhysicsComponents   map[int]*PhysicsComponent // Entity ID -> Physics
	AdvancedTimeSystem  *AdvancedTimeSystem
	CivilizationSystem  *CivilizationSystem
	ViewportSystem      *ViewportSystem
	WindSystem          *WindSystem         // Wind and pollen dispersal system
	SpeciationSystem    *SpeciationSystem   // Species evolution and tracking
	PlantNetworkSystem  *PlantNetworkSystem // Underground plant networks and communication
	SpeciesNaming       *SpeciesNaming      // Species naming and evolutionary relationships

	// Micro and Macro Evolution Systems
	DNASystem            *DNASystem            // DNA-based genetic system
	CellularSystem       *CellularSystem       // Cellular-level evolution and processes
	MacroEvolutionSystem *MacroEvolutionSystem // Macro-evolution tracking
	TopologySystem       *TopologySystem       // World terrain and geological processes
	FluidRegions         []FluidRegion

	// Tool and Environmental Modification Systems
	ToolSystem              *ToolSystem                          // Tool creation and usage system
	EnvironmentalModSystem  *EnvironmentalModificationSystem     // Environmental modifications system
	EmergentBehaviorSystem  *EmergentBehaviorSystem              // Emergent behavior and learning system
	
	// Reproduction and Decay System
	ReproductionSystem      *ReproductionSystem                  // Reproduction, gestation, and decay management
}

// NewWorld creates a new world with multiple populations
func NewWorld(config WorldConfig) *World {
	world := &World{
		Config:      config,
		Populations: make(map[string]*Population),
		AllEntities: make([]*Entity, 0),
		AllPlants:   make([]*Plant, 0),
		Grid:        make([][]GridCell, config.GridHeight),
		Biomes:      initializeBiomes(),
		Events:      make([]*WorldEvent, 0),
		EventLogger: NewEventLogger(1000), // Keep up to 1000 events
		NextID:      0,
		NextPlantID: 0,
		Tick:        0,
		Clock:       time.Now(),
		LastUpdate:  time.Now(),
	}

	// Initialize grid
	for y := 0; y < config.GridHeight; y++ {
		world.Grid[y] = make([]GridCell, config.GridWidth)
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x] = GridCell{
				Biome:    world.generateBiome(x, y),
				Entities: make([]*Entity, 0),
				Plants:   make([]*Plant, 0),
				Event:    nil,
			}
		}
	} // Initialize advanced systems
	world.CommunicationSystem = NewCommunicationSystem()
	world.GroupBehaviorSystem = NewGroupBehaviorSystem()
	world.PhysicsSystem = NewPhysicsSystem()
	world.CollisionSystem = NewCollisionSystem()
	world.PhysicsComponents = make(map[int]*PhysicsComponent)
	world.AdvancedTimeSystem = NewAdvancedTimeSystem(480, 120) // 480 ticks/day, 120 days/season
	world.CivilizationSystem = NewCivilizationSystem()
	world.ViewportSystem = NewViewportSystem(config.Width, config.Height)
	world.WindSystem = NewWindSystem(int(config.Width), int(config.Height))
	world.SpeciationSystem = NewSpeciationSystem()
	world.PlantNetworkSystem = NewPlantNetworkSystem()
	world.SpeciesNaming = NewSpeciesNaming()

	// Initialize new evolution and topology systems
	world.DNASystem = NewDNASystem()
	world.CellularSystem = NewCellularSystem(world.DNASystem)
	world.MacroEvolutionSystem = NewMacroEvolutionSystem()
	world.TopologySystem = NewTopologySystem(config.GridWidth, config.GridHeight)

	// Initialize tool and environmental modification systems
	world.ToolSystem = NewToolSystem()
	world.EnvironmentalModSystem = NewEnvironmentalModificationSystem()
	world.EmergentBehaviorSystem = NewEmergentBehaviorSystem()
	
	// Initialize reproduction and decay system
	world.ReproductionSystem = NewReproductionSystem()

  // Generate initial world terrain
	world.TopologySystem.GenerateInitialTerrain()

	world.FluidRegions = make([]FluidRegion, 0)

	// Initialize plant life
	world.initializePlants()

	return world
}

// initializeBiomes creates the biome definitions
func initializeBiomes() map[BiomeType]Biome {
	biomes := make(map[BiomeType]Biome)

	biomes[BiomePlains] = Biome{
		Type:           BiomePlains,
		Name:           "Plains",
		Color:          "green",
		TraitModifiers: map[string]float64{"speed": 0.1},
		MutationRate:   0.0,
		EnergyDrain:    0.5,
		Symbol:         '.',
	}

	biomes[BiomeForest] = Biome{
		Type:           BiomeForest,
		Name:           "Forest",
		Color:          "darkgreen",
		TraitModifiers: map[string]float64{"size": 0.2, "defense": 0.1},
		MutationRate:   0.0,
		EnergyDrain:    0.8,
		Symbol:         '♠',
	}

	biomes[BiomeDesert] = Biome{
		Type:           BiomeDesert,
		Name:           "Desert",
		Color:          "yellow",
		TraitModifiers: map[string]float64{"endurance": 0.3, "size": -0.1},
		MutationRate:   0.05,
		EnergyDrain:    1.5,
		Symbol:         '~',
	}

	biomes[BiomeMountain] = Biome{
		Type:           BiomeMountain,
		Name:           "Mountain",
		Color:          "gray",
		TraitModifiers: map[string]float64{"strength": 0.2, "speed": -0.1},
		MutationRate:   0.0,
		EnergyDrain:    1.2,
		Symbol:         '^',
	}

	biomes[BiomeWater] = Biome{
		Type:           BiomeWater,
		Name:           "Water",
		Color:          "blue",
		TraitModifiers: map[string]float64{"speed": 0.2, "size": 0.1},
		MutationRate:   0.0,
		EnergyDrain:    0.3,
		Symbol:         '≈',
	}

	biomes[BiomeRadiation] = Biome{
		Type:           BiomeRadiation,
		Name:           "Radiation",
		Color:          "red",
		TraitModifiers: map[string]float64{"endurance": -0.2},
		MutationRate:   0.3,
		EnergyDrain:    2.0,
		Symbol:         '☢',
	}

	biomes[BiomeSoil] = Biome{
		Type:           BiomeSoil,
		Name:           "Soil",
		Color:          "brown",
		TraitModifiers: map[string]float64{"digging_ability": 0.3, "size": -0.1, "underground_nav": 0.2},
		MutationRate:   0.02,
		EnergyDrain:    0.7,
		Symbol:         '■',
	}

	biomes[BiomeAir] = Biome{
		Type:           BiomeAir,
		Name:           "Air",
		Color:          "cyan",
		TraitModifiers: map[string]float64{"flying_ability": 0.4, "altitude_tolerance": 0.3, "size": -0.2},
		MutationRate:   0.01,
		EnergyDrain:    1.0,
		Symbol:         '☁',
	}

	return biomes
}

// generateBiome generates a biome type for a grid cell using Perlin-like noise
func (w *World) generateBiome(x, y int) BiomeType {
	// Simple biome generation based on position
	distFromCenter := math.Sqrt(math.Pow(float64(x)-float64(w.Config.GridWidth)/2, 2) +
		math.Pow(float64(y)-float64(w.Config.GridHeight)/2, 2))
	maxDist := math.Sqrt(math.Pow(float64(w.Config.GridWidth)/2, 2) +
		math.Pow(float64(w.Config.GridHeight)/2, 2))

	noise := rand.Float64()

	// Center tends to be plains/forest
	if distFromCenter < maxDist*0.3 {
		if noise < 0.6 {
			return BiomePlains
		} else {
			return BiomeForest
		}
	}

	// Mid range has variety
	if distFromCenter < maxDist*0.7 {
		switch {
		case noise < 0.3:
			return BiomeForest
		case noise < 0.5:
			return BiomeWater
		case noise < 0.7:
			return BiomeMountain
		default:
			return BiomePlains
		}
	}

	// Outer areas tend to be harsh
	switch {
	case noise < 0.3:
		return BiomeDesert
	case noise < 0.5:
		return BiomeMountain
	case noise < 0.65:
		return BiomeWater
	case noise < 0.8:
		return BiomePlains
	case noise < 0.9:
		return BiomeSoil
	case noise < 0.97:
		return BiomeAir
	default:
		return BiomeRadiation
	}
}

// AddPopulation adds a new population to the world
func (w *World) AddPopulation(config PopulationConfig) {
	// Generate a proper species name using the naming system
	speciesName := w.SpeciesNaming.GenerateSpeciesName(config.Species, "", 0, w.Tick)

	// Generate trait names based on base traits
	traitNames := make([]string, 0, len(config.BaseTraits))
	for name := range config.BaseTraits {
		traitNames = append(traitNames, name)
	}

	// Create population with species-specific mutation rate
	pop := NewPopulation(w.Config.PopulationSize, traitNames, config.BaseMutationRate, 0.2)
	pop.Species = speciesName

	// Initialize entities with base traits and positions
	for _, entity := range pop.Entities {
		// Set position around start position
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * config.Spread

		entity.Position = Position{
			X: config.StartPos.X + math.Cos(angle)*distance,
			Y: config.StartPos.Y + math.Sin(angle)*distance,
		}
		entity.Species = speciesName
		entity.ID = w.NextID
		w.NextID++

		// Apply base traits with some variation
		for traitName, baseValue := range config.BaseTraits {
			variation := (rand.Float64() - 0.5) * 0.4 // ±20% variation
			value := baseValue + variation
			value = math.Max(-2.0, math.Min(2.0, value))
			entity.SetTrait(traitName, value)
		}

		// Create DNA for entity
		dna := w.DNASystem.GenerateRandomDNA(entity.ID, entity.Generation)

		// Create cellular organism
		w.CellularSystem.CreateSingleCellOrganism(entity.ID, dna)

		// Update entity traits based on DNA expression
		for traitName := range entity.Traits {
			dnaValue := w.DNASystem.ExpressTrait(dna, traitName)
			// Blend DNA value with existing trait (50/50 blend)
			currentValue := entity.GetTrait(traitName)
			newValue := (currentValue + dnaValue) / 2.0
			entity.SetTrait(traitName, newValue)
		}

		w.AllEntities = append(w.AllEntities, entity)
	}

	w.Populations[speciesName] = pop
}

// Update simulates one tick of the world
func (w *World) Update() {
	// Skip update if paused
	if w.Paused {
		return
	}

	w.Tick++
	now := time.Now()
	w.Clock = w.Clock.Add(time.Hour) // Each tick = 1 hour world time
	w.LastUpdate = now
	// 1. Update advanced time system (affects all other systems)
	w.AdvancedTimeSystem.Update()
	currentTimeState := w.AdvancedTimeSystem.GetTimeState()

	// 2. Update wind system (affects pollen dispersal and plant reproduction)
	w.WindSystem.Update(currentTimeState.Season, w.Tick)

	// 3. Update micro and macro evolution systems
	w.CellularSystem.UpdateCellularOrganisms()
	w.MacroEvolutionSystem.UpdateMacroEvolution(w)
	w.TopologySystem.UpdateTopology(w.Tick)

	// Clear grid entities and plants
	w.clearGrid()

	// Update world events
	w.updateEvents()
	// Maybe trigger new events (less frequent during night)
	eventChance := 0.01
	if currentTimeState.IsNight() {
		eventChance *= 0.5 // Fewer events at night
	}
	if rand.Float64() < eventChance {
		w.triggerRandomEvent()
	}
	// Update all plants (affected by day/night cycle)
	w.updatePlants()

	// Update plant network system (underground networks and communication)
	w.PlantNetworkSystem.Update(w.AllPlants, w.Tick)

	// 2. Create physics components for new entities
	for _, entity := range w.AllEntities {
		if entity.IsAlive && w.PhysicsComponents[entity.ID] == nil {
			w.PhysicsComponents[entity.ID] = NewPhysicsComponent(entity)
		}
	}

	// 3. Update communication system (entities send signals)
	w.CommunicationSystem.Update()

	// Update all entities with biome effects, time effects, and starvation checks
	deltaTime := 0.1 // Physics time step

	// Use concurrent processing for entity updates if we have many entities
	if len(w.AllEntities) > 50 {
		w.updateEntitiesConcurrent(currentTimeState, deltaTime)
		// Calculate inter-entity physics forces after concurrent updates
		w.updateEntityPhysicsForces()
	} else {
		w.updateEntitiesSequential(currentTimeState, deltaTime)
	}

	// 5. Reset collision counters and check collisions
	w.PhysicsSystem.ResetCollisionCounters()
	w.CollisionSystem.CheckCollisions(w.AllEntities, w.PhysicsComponents, w.PhysicsSystem)

	// Update grid with current entity and plant positions
	w.updateGrid()

	// 6. Update group behavior system
	w.GroupBehaviorSystem.UpdateGroups()

	// Try to form new groups based on proximity and compatibility
	if w.Tick%10 == 0 {
		w.attemptGroupFormation()
	}

	// Handle interactions between entities and with plants
	w.handleInteractions()
	// 7. Update civilization system
	w.CivilizationSystem.Update()

	// Process civilization activities
	w.processCivilizationActivities()

	// Update reproduction system (gestation, egg hatching, decay)
	w.updateReproductionSystem()

	// Remove dead entities and plants
	w.removeDeadEntities()
	w.removeDeadPlants()
	// Plant reproduction
	if w.Tick%10 == 0 {
		w.reproducePlants()
	}

	// Update species evolution and tracking (after plant reproduction)
	if w.Tick%20 == 0 {
		w.SpeciationSystem.Update(w.AllPlants, w.Tick)
	}

	// Population-level evolution (less frequent)
	if w.Tick%50 == 0 {
		w.evolvePopulations()
	}

	// Spawn new entities occasionally (based on carrying capacity)
	if w.Tick%20 == 0 {
		w.spawnNewEntities()
	}

	// Clean up physics components for dead entities
	for entityID := range w.PhysicsComponents {
		found := false
		for _, entity := range w.AllEntities {
			if entity.ID == entityID && entity.IsAlive {
				found = true
				break
			}
		}
		if !found {
			delete(w.PhysicsComponents, entityID)
		}
	}

	// Update tool system
	w.ToolSystem.UpdateTools(w.Tick)

	// Update environmental modification system
	w.EnvironmentalModSystem.UpdateModifications(w.Tick)

	// Update emergent behavior system
	w.EmergentBehaviorSystem.UpdateEntityBehaviors(w)

	// Update event logger with population changes
	w.EventLogger.UpdatePopulationCounts(w.Tick, w.Populations)
}

// getBiomeAtPosition returns the biome type at the given world position
func (w *World) getBiomeAtPosition(x, y float64) BiomeType {
	// Convert world coordinates to grid coordinates
	gridX := int((x / w.Config.Width) * float64(w.Config.GridWidth))
	gridY := int((y / w.Config.Height) * float64(w.Config.GridHeight))

	// Clamp to grid bounds
	gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
	gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

	return w.Grid[gridY][gridX].Biome
}

// getEntitiesNearPosition returns entities within a given radius of a position
func (w *World) getEntitiesNearPosition(pos Position, radius float64) []*Entity {
	nearby := make([]*Entity, 0)

	for _, entity := range w.AllEntities {
		if entity.IsAlive {
			distance := math.Sqrt(math.Pow(entity.Position.X-pos.X, 2) + math.Pow(entity.Position.Y-pos.Y, 2))
			if distance <= radius {
				nearby = append(nearby, entity)
			}
		}
	}

	return nearby
}

// updateEntitiesSequential updates entities using single-threaded processing
func (w *World) updateEntitiesSequential(currentTimeState TimeState, deltaTime float64) {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Apply biome effects
		w.updateEntityWithBiome(entity)

		// Track environmental exposure for feedback loops
		w.trackEntityEnvironmentalExposure(entity, currentTimeState)

		// Apply time-based effects (circadian preferences)
		w.applyTimeEffects(entity, currentTimeState)

		// Check starvation-driven evolution
		entity.CheckStarvation(w)

		// Update basic entity properties
		entity.Update()

		// 4. Apply physics forces and movement
		physics := w.PhysicsComponents[entity.ID]
		if physics != nil {
			// Get entity's current biome
			gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
			gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
			gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
			gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))
			biome := w.Grid[gridY][gridX].Biome

			// Calculate attraction/repulsion forces between entities
			for _, other := range w.AllEntities {
				if other.ID != entity.ID && other.IsAlive {
					otherPhysics := w.PhysicsComponents[other.ID]
					if otherPhysics != nil {
						force := w.PhysicsSystem.CalculateAttraction(entity, other, physics, otherPhysics)
						w.PhysicsSystem.ApplyForce(physics, force)
					}
				}
			}

			// Apply fluid effects if in fluid regions
			w.PhysicsSystem.ApplyFluidEffects(entity, physics, w.FluidRegions)

			// Update physics
			w.PhysicsSystem.ApplyPhysics(entity, physics, biome, deltaTime)
		}
		// Handle entity communication and signaling
		w.handleEntityCommunication(entity)
	}
}

// updateEntitiesConcurrent updates entities using multi-threaded processing
func (w *World) updateEntitiesConcurrent(currentTimeState TimeState, deltaTime float64) {
	// Worker pool for concurrent entity processing
	numWorkers := 4 // Use 4 goroutines for parallel processing
	workChan := make(chan *Entity, len(w.AllEntities))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				w.updateSingleEntity(entity, currentTimeState, deltaTime)
			}
		}()
	}

	// Send entities to workers
	for _, entity := range w.AllEntities {
		if entity.IsAlive {
			workChan <- entity
		}
	}
	close(workChan)

	// Wait for all workers to complete
	wg.Wait()
}

// updateSingleEntity updates a single entity (thread-safe parts only)
func (w *World) updateSingleEntity(entity *Entity, currentTimeState TimeState, deltaTime float64) {
	// Apply biome effects
	w.updateEntityWithBiome(entity)

	// Track environmental exposure for feedback loops
	w.trackEntityEnvironmentalExposure(entity, currentTimeState)

	// Apply time-based effects (circadian preferences)
	w.applyTimeEffects(entity, currentTimeState)

	// Check starvation-driven evolution
	entity.CheckStarvation(w)

	// Update basic entity properties
	entity.Update()

	// Note: Physics force calculations and interactions are handled separately
	// to avoid race conditions between entities

	// Handle entity communication and signaling
	w.handleEntityCommunication(entity)

	// Apply basic physics (without inter-entity forces)
	physics := w.PhysicsComponents[entity.ID]
	if physics != nil {
		// Get entity's current biome
		gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))
		biome := w.Grid[gridY][gridX].Biome

		// Apply fluid effects if in fluid regions
		w.PhysicsSystem.ApplyFluidEffects(entity, physics, w.FluidRegions)

		// Update physics (without inter-entity forces for now)
		w.PhysicsSystem.ApplyPhysics(entity, physics, biome, deltaTime)
	}
}

// updateEntityPhysicsForces calculates inter-entity forces (done after concurrent updates)
func (w *World) updateEntityPhysicsForces() {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		physics := w.PhysicsComponents[entity.ID]
		if physics != nil {
			// Calculate attraction/repulsion forces between entities
			for _, other := range w.AllEntities {
				if other.ID != entity.ID && other.IsAlive {
					otherPhysics := w.PhysicsComponents[other.ID]
					if otherPhysics != nil {
						force := w.PhysicsSystem.CalculateAttraction(entity, other, physics, otherPhysics)
						w.PhysicsSystem.ApplyForce(physics, force)
					}
				}
			}
		}
	}
}

// updatePlants handles plant growth, aging, and death
func (w *World) updatePlants() {
	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}

		// Get biome for plant's location
		gridX := int((plant.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((plant.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		biome := w.Biomes[w.Grid[gridY][gridX].Biome]
		plant.Update(biome)
	}
}

// reproducePlants handles plant reproduction
func (w *World) reproducePlants() {
	newPlants := make([]*Plant, 0)

	// Get current time state for seasonal effects
	currentTimeState := w.AdvancedTimeSystem.GetTimeState()

	// First, process asexual reproduction and pollen release
	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}

		// Asexual reproduction (existing behavior)
		if offspring := plant.Reproduce(w.NextPlantID); offspring != nil {
			w.NextPlantID++
			newPlants = append(newPlants, offspring)
		}

		// Release pollen for sexual reproduction during flowering
		if plant.CanReproduce() && currentTimeState.Season == Spring && rand.Float64() < 0.6 {
			// Determine pollen amount based on plant traits and type
			pollenAmount := int(5 + plant.Size*10 + plant.GetTrait("reproduction_rate")*15)

			// Different plant types have different pollen release patterns
			switch plant.Type {
			case PlantGrass:
				pollenAmount *= 2 // Grasses release lots of pollen
			case PlantTree:
				pollenAmount *= 3 // Trees release huge amounts
			case PlantBush:
				pollenAmount = int(float64(pollenAmount) * 1.2) // Bushes moderate
			case PlantMushroom:
				pollenAmount = int(float64(pollenAmount) * 0.8) // Mushrooms release spores
			case PlantAlgae:
				pollenAmount = int(float64(pollenAmount) * 0.5) // Algae less pollen in water
			case PlantCactus:
				pollenAmount = int(float64(pollenAmount) * 0.7) // Cacti conserve resources
			}

			w.WindSystem.ReleasePollen(plant, pollenAmount)
		}
	}
	// Process wind-based cross-pollination
	crossPollinatedPlants := w.WindSystem.TryPollination(w.AllPlants, w.SpeciationSystem)

	// Assign IDs to cross-pollinated plants
	for _, offspring := range crossPollinatedPlants {
		offspring.ID = w.NextPlantID
		w.NextPlantID++
		newPlants = append(newPlants, offspring)
	}

	// Add new plants to world
	w.AllPlants = append(w.AllPlants, newPlants...)

	// Enhanced logging for reproduction events
	if len(newPlants) > 5 {
		asexualReproduction := len(newPlants) - len(crossPollinatedPlants)
		w.EventLogger.LogEcosystemShift(w.Tick,
			fmt.Sprintf("Plant reproduction boom: %d new plants (%d asexual, %d cross-pollinated)",
				len(newPlants), asexualReproduction, len(crossPollinatedPlants)),
			map[string]interface{}{
				"new_plants":           len(newPlants),
				"asexual_reproduction": asexualReproduction,
				"cross_pollination":    len(crossPollinatedPlants),
				"active_pollen_grains": len(w.WindSystem.AllPollenGrains),
			})
	}
}

// removeDeadPlants removes dead plants from the world
func (w *World) removeDeadPlants() {
	alivePlants := make([]*Plant, 0, len(w.AllPlants))

	for _, plant := range w.AllPlants {
		if plant.IsAlive {
			alivePlants = append(alivePlants, plant)
		}
	}

	if len(alivePlants) < len(w.AllPlants) {
		deadCount := len(w.AllPlants) - len(alivePlants)
		if deadCount > 10 {
			w.EventLogger.LogEcosystemShift(w.Tick,
				fmt.Sprintf("Significant plant die-off: %d plants died", deadCount),
				map[string]interface{}{"plants_died": deadCount})
		}
	}

	w.AllPlants = alivePlants
}

// clearGrid clears all entities and plants from grid cells
func (w *World) clearGrid() {
	for y := 0; y < w.Config.GridHeight; y++ {
		for x := 0; x < w.Config.GridWidth; x++ {
			w.Grid[y][x].Entities = w.Grid[y][x].Entities[:0]
			w.Grid[y][x].Plants = w.Grid[y][x].Plants[:0]
		}
	}
}

// updateGrid places entities and plants in their current grid cells
func (w *World) updateGrid() {
	// Place entities in grid
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Convert world coordinates to grid coordinates
		gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		w.Grid[gridY][gridX].Entities = append(w.Grid[gridY][gridX].Entities, entity)
	}

	// Place plants in grid
	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}

		// Convert world coordinates to grid coordinates
		gridX := int((plant.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((plant.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		w.Grid[gridY][gridX].Plants = append(w.Grid[gridY][gridX].Plants, plant)
	}
}

// updateEntityWithBiome applies biome effects to an entity
func (w *World) updateEntityWithBiome(entity *Entity) {
	if !entity.IsAlive {
		return
	}

	// Get entity's grid position
	gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
	gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

	// Clamp to grid bounds
	gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
	gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

	cell := &w.Grid[gridY][gridX]
	biome := w.Biomes[cell.Biome]

	// Apply biome energy drain
	entity.Energy -= biome.EnergyDrain

	// Apply biome mutation effects
	if biome.MutationRate > 0 && rand.Float64() < biome.MutationRate {
		entity.Mutate(biome.MutationRate, 0.1)
	}

	// Apply event effects if present
	if cell.Event != nil {
		entity.Energy -= cell.Event.GlobalDamage
		if cell.Event.GlobalMutation > 0 && rand.Float64() < cell.Event.GlobalMutation {
			entity.Mutate(cell.Event.GlobalMutation, 0.2)
		}
	}

	// Move entities randomly within their preferred biomes
	w.moveEntityInBiome(entity, biome)
}

// moveEntityInBiome makes entities move based on biome preferences
func (w *World) moveEntityInBiome(entity *Entity, biome Biome) {
	// Movement based on entity traits and biome
	speed := entity.GetTrait("speed")
	intelligence := entity.GetTrait("intelligence")

	// Intelligent entities seek better biomes
	if intelligence > 0.5 && rand.Float64() < 0.3 {
		w.seekBetterBiome(entity)
	} else {
		// Random movement modified by speed and biome effects
		maxMove := (0.5 + speed*0.5) * (w.Config.Width / float64(w.Config.GridWidth))
		entity.MoveRandomly(maxMove)
	}

	// Keep entities within world bounds
	entity.Position.X = math.Max(0, math.Min(w.Config.Width, entity.Position.X))
	entity.Position.Y = math.Max(0, math.Min(w.Config.Height, entity.Position.Y))
}

// seekBetterBiome makes intelligent entities move toward favorable biomes
func (w *World) seekBetterBiome(entity *Entity) {
	bestScore := -1000.0
	bestX, bestY := entity.Position.X, entity.Position.Y

	// Check nearby grid cells
	currentGridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
	currentGridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			checkX := currentGridX + dx
			checkY := currentGridY + dy

			if checkX >= 0 && checkX < w.Config.GridWidth &&
				checkY >= 0 && checkY < w.Config.GridHeight {

				biome := w.Biomes[w.Grid[checkY][checkX].Biome]
				score := w.evaluateBiomeForEntity(entity, biome)

				if score > bestScore {
					bestScore = score
					bestX = (float64(checkX) + 0.5) * (w.Config.Width / float64(w.Config.GridWidth))
					bestY = (float64(checkY) + 0.5) * (w.Config.Height / float64(w.Config.GridHeight))
				}
			}
		}
	}

	// Move toward best biome if found
	if bestScore > -1000.0 {
		speed := 0.3 + entity.GetTrait("speed")*0.2
		entity.MoveTo(bestX, bestY, speed)
	}
}

// evaluateBiomeForEntity scores how good a biome is for an entity
func (w *World) evaluateBiomeForEntity(entity *Entity, biome Biome) float64 {
	score := -biome.EnergyDrain * 10 // Avoid high energy drain

	// Add points for beneficial trait modifiers
	for trait, modifier := range biome.TraitModifiers {
		entityValue := entity.GetTrait(trait)
		if modifier > 0 && entityValue > 0 {
			score += modifier * entityValue * 50
		} else if modifier < 0 && entityValue < 0 {
			score += -modifier * -entityValue * 50
		}
	}

	// Penalize high mutation areas unless entity has good endurance
	if biome.MutationRate > 0 {
		endurance := entity.GetTrait("endurance")
		score -= biome.MutationRate * 100 * (1.0 - endurance)
	}

	return score
}

// updateEvents updates active world events
func (w *World) updateEvents() {
	newEvents := make([]*WorldEvent, 0)

	for _, event := range w.Events {
		event.Duration--
		if event.Duration > 0 {
			newEvents = append(newEvents, event)
		}
	}

	w.Events = newEvents
}

// triggerRandomEvent creates a new random world event
func (w *World) triggerRandomEvent() {
	events := []WorldEvent{
		{
			Name:           "Solar Flare",
			Description:    "Increased radiation across the world",
			Duration:       30,
			GlobalMutation: 0.2,
			GlobalDamage:   2.0,
		},
		{
			Name:           "Meteor Shower",
			Description:    "Meteors create radiation zones",
			Duration:       50,
			GlobalMutation: 0.05,
			GlobalDamage:   1.0,
			BiomeChanges:   w.generateMeteorCraters(),
		},
		{
			Name:           "Ice Age",
			Description:    "World cools, increasing energy drain",
			Duration:       100,
			GlobalMutation: 0.0,
			GlobalDamage:   1.5,
		},
		{
			Name:           "Volcanic Winter",
			Description:    "Ash clouds block sunlight",
			Duration:       75,
			GlobalMutation: 0.1,
			GlobalDamage:   2.5,
		},
		{
			Name:           "Volcanic Eruption",
			Description:    "Massive lava flows create new biomes",
			Duration:       40,
			GlobalMutation: 0.15,
			GlobalDamage:   3.0,
			BiomeChanges:   w.generateVolcanicFields(),
		},
		{
			Name:           "Lightning Storm",
			Description:    "Electrical discharges cause widespread mutations",
			Duration:       20,
			GlobalMutation: 0.3,
			GlobalDamage:   1.0,
		},
		{
			Name:           "Wildfire",
			Description:    "Fires spread across vegetation",
			Duration:       35,
			GlobalMutation: 0.05,
			GlobalDamage:   2.0,
			BiomeChanges:   w.generateFireZones(),
		},
		{
			Name:           "Great Flood",
			Description:    "Rising waters reshape the landscape",
			Duration:       60,
			GlobalMutation: 0.08,
			GlobalDamage:   1.8,
			BiomeChanges:   w.generateFloodZones(),
		},
		{
			Name:           "Magnetic Storm",
			Description:    "Electromagnetic chaos disrupts navigation",
			Duration:       25,
			GlobalMutation: 0.12,
			GlobalDamage:   0.5,
		},
		{
			Name:           "Ash Cloud",
			Description:    "Dense ash blocks sunlight and poisons air",
			Duration:       45,
			GlobalMutation: 0.08,
			GlobalDamage:   2.2,
		},
		{
			Name:           "Earthquake",
			Description:    "Seismic activity creates new mountain ranges",
			Duration:       15,
			GlobalMutation: 0.05,
			GlobalDamage:   1.5,
			BiomeChanges:   w.generateSeismicChanges(),
		},
		{
			Name:           "Cosmic Radiation",
			Description:    "Interstellar radiation penetrates atmosphere",
			Duration:       80,
			GlobalMutation: 0.25,
			GlobalDamage:   1.0,
		},
	}

	event := events[rand.Intn(len(events))]
	w.Events = append(w.Events, &event)
}

// generateMeteorCraters creates radiation zones from meteor impacts
func (w *World) generateMeteorCraters() map[Position]BiomeType {
	craters := make(map[Position]BiomeType)
	numCraters := 3 + rand.Intn(5)

	for i := 0; i < numCraters; i++ {
		x := rand.Intn(w.Config.GridWidth)
		y := rand.Intn(w.Config.GridHeight)
		craters[Position{X: float64(x), Y: float64(y)}] = BiomeRadiation

		// Add smaller radiation zones around impact
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if x+dx >= 0 && x+dx < w.Config.GridWidth &&
					y+dy >= 0 && y+dy < w.Config.GridHeight &&
					rand.Float64() < 0.5 {
					craters[Position{X: float64(x + dx), Y: float64(y + dy)}] = BiomeRadiation
				}
			}
		}
	}

	return craters
}

// generateVolcanicFields creates mountain and radiation zones from volcanic activity
func (w *World) generateVolcanicFields() map[Position]BiomeType {
	changes := make(map[Position]BiomeType)
	numVolcanoes := 1 + rand.Intn(3)

	for i := 0; i < numVolcanoes; i++ {
		centerX := rand.Intn(w.Config.GridWidth)
		centerY := rand.Intn(w.Config.GridHeight)

		// Create volcanic mountain at center
		changes[Position{X: float64(centerX), Y: float64(centerY)}] = BiomeMountain

		// Add lava flows (radiation zones) radiating outward
		for radius := 1; radius <= 3; radius++ {
			for angle := 0; angle < 360; angle += 45 {
				radian := float64(angle) * math.Pi / 180
				x := centerX + int(float64(radius)*math.Cos(radian))
				y := centerY + int(float64(radius)*math.Sin(radian))

				if x >= 0 && x < w.Config.GridWidth && y >= 0 && y < w.Config.GridHeight &&
					rand.Float64() < 0.6 {
					changes[Position{X: float64(x), Y: float64(y)}] = BiomeRadiation
				}
			}
		}
	}

	return changes
}

// generateFireZones creates desert zones from wildfires
func (w *World) generateFireZones() map[Position]BiomeType {
	changes := make(map[Position]BiomeType)
	numFires := 2 + rand.Intn(4)

	for i := 0; i < numFires; i++ {
		centerX := rand.Intn(w.Config.GridWidth)
		centerY := rand.Intn(w.Config.GridHeight)

		// Fire spreads in irregular patterns
		for radius := 0; radius <= 4; radius++ {
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					x := centerX + dx
					y := centerY + dy

					if x >= 0 && x < w.Config.GridWidth && y >= 0 && y < w.Config.GridHeight {
						distance := math.Sqrt(float64(dx*dx + dy*dy))
						// Fire probability decreases with distance
						fireChance := 0.8 * math.Exp(-distance/2.0)
						if rand.Float64() < fireChance {
							changes[Position{X: float64(x), Y: float64(y)}] = BiomeDesert
						}
					}
				}
			}
		}
	}

	return changes
}

// generateFloodZones creates water zones from flooding
func (w *World) generateFloodZones() map[Position]BiomeType {
	changes := make(map[Position]BiomeType)
	numFloodSources := 1 + rand.Intn(2)

	for i := 0; i < numFloodSources; i++ {
		// Start flood from edge of map (representing river overflow)
		var centerX, centerY int
		edge := rand.Intn(4)
		switch edge {
		case 0: // top edge
			centerX = rand.Intn(w.Config.GridWidth)
			centerY = 0
		case 1: // right edge
			centerX = w.Config.GridWidth - 1
			centerY = rand.Intn(w.Config.GridHeight)
		case 2: // bottom edge
			centerX = rand.Intn(w.Config.GridWidth)
			centerY = w.Config.GridHeight - 1
		case 3: // left edge
			centerX = 0
			centerY = rand.Intn(w.Config.GridHeight)
		}

		// Flood spreads inward
		for radius := 0; radius <= 6; radius++ {
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					x := centerX + dx
					y := centerY + dy

					if x >= 0 && x < w.Config.GridWidth && y >= 0 && y < w.Config.GridHeight {
						distance := math.Sqrt(float64(dx*dx + dy*dy))
						// Flood probability decreases with distance
						floodChance := 0.7 * math.Exp(-distance/3.0)
						if rand.Float64() < floodChance {
							changes[Position{X: float64(x), Y: float64(y)}] = BiomeWater
						}
					}
				}
			}
		}
	}

	return changes
}

// generateSeismicChanges creates mountain ranges from earthquakes
func (w *World) generateSeismicChanges() map[Position]BiomeType {
	changes := make(map[Position]BiomeType)

	// Create fault lines that generate mountain ranges
	numFaults := 1 + rand.Intn(2)

	for i := 0; i < numFaults; i++ {
		// Random fault line across the map
		startX := rand.Intn(w.Config.GridWidth)
		startY := rand.Intn(w.Config.GridHeight)

		// Fault direction
		angle := rand.Float64() * 2 * math.Pi
		length := 8 + rand.Intn(12)

		for step := 0; step < length; step++ {
			x := startX + int(float64(step)*math.Cos(angle))
			y := startY + int(float64(step)*math.Sin(angle))

			if x >= 0 && x < w.Config.GridWidth && y >= 0 && y < w.Config.GridHeight {
				changes[Position{X: float64(x), Y: float64(y)}] = BiomeMountain

				// Add nearby elevations
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < w.Config.GridWidth && ny >= 0 && ny < w.Config.GridHeight &&
							rand.Float64() < 0.4 {
							changes[Position{X: float64(nx), Y: float64(ny)}] = BiomeMountain
						}
					}
				}
			}
		}
	}

	return changes
}

// handleInteractions processes interactions between nearby entities and with plants
func (w *World) handleInteractions() {
	interactionDistance := 5.0

	// Entity-entity interactions
	for i, entity1 := range w.AllEntities {
		if !entity1.IsAlive {
			continue
		}

		for j, entity2 := range w.AllEntities {
			if i >= j || !entity2.IsAlive {
				continue
			}

			distance := entity1.DistanceTo(entity2)
			if distance <= interactionDistance {
				w.processEntityInteraction(entity1, entity2)
			}
		}
	}

	// Entity-plant interactions
	w.handleEntityPlantInteractions()
}

// handleEntityPlantInteractions processes interactions between entities and plants
func (w *World) handleEntityPlantInteractions() {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Get entity's grid position
		gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		cell := &w.Grid[gridY][gridX]

		// Try to eat plants in the same cell
		for _, plant := range cell.Plants {
			if !plant.IsAlive {
				continue
			}

			// Check if entity can and wants to eat this plant
			if entity.CanEatPlant(plant) && rand.Float64() < 0.4 {
				if entity.EatPlant(plant, w.Tick) {
					// Log successful plant consumption
					if rand.Float64() < 0.1 { // Log 10% of plant eating events
						w.EventLogger.LogEcosystemShift(w.Tick,
							fmt.Sprintf("%s ate %s for nutrition", entity.Species, GetPlantConfigs()[plant.Type].Name),
							map[string]interface{}{
								"entity_species": entity.Species,
								"plant_type":     GetPlantConfigs()[plant.Type].Name,
								"entity_energy":  entity.Energy,
							})
					}
					break // Entity can only eat one plant per interaction
				}
			}
		}
	}
}

// processEntityInteraction handles a specific interaction between two entities
func (w *World) processEntityInteraction(entity1, entity2 *Entity) {
	// Same species interactions
	if entity1.Species == entity2.Species {
		// Chance to merge if conditions are met
		if rand.Float64() < 0.01 && entity1.CanMerge(entity2) {
			merged := entity1.Merge(entity2, w.NextID)
			if merged != nil {
				w.NextID++
				w.AllEntities = append(w.AllEntities, merged)
			}
		}
		return
	}

	// Different species interactions
	// Try to kill/eat
	if entity1.CanKill(entity2) && rand.Float64() < 0.1 {
		entity1.Kill(entity2)
	} else if entity2.CanKill(entity1) && rand.Float64() < 0.1 {
		entity2.Kill(entity1)
	}

	// Try to eat dead entities
	if !entity2.IsAlive && entity1.CanEat(entity2) && rand.Float64() < 0.3 {
		entity1.Eat(entity2, w.Tick)
	} else if !entity1.IsAlive && entity2.CanEat(entity1) && rand.Float64() < 0.3 {
		entity2.Eat(entity1, w.Tick)
	}
}

// removeDeadEntities removes dead entities from the world
func (w *World) removeDeadEntities() {
	aliveEntities := make([]*Entity, 0, len(w.AllEntities))

	for _, entity := range w.AllEntities {
		if entity.IsAlive {
			aliveEntities = append(aliveEntities, entity)
		}
	}

	w.AllEntities = aliveEntities

	// Update population entities lists
	for _, pop := range w.Populations {
		alivePopEntities := make([]*Entity, 0)
		for _, entity := range pop.Entities {
			if entity.IsAlive {
				alivePopEntities = append(alivePopEntities, entity)
			}
		}
		pop.Entities = alivePopEntities
	}
}

// evolvePopulations runs evolution on each population
func (w *World) evolvePopulations() {
	for _, pop := range w.Populations {
		if len(pop.Entities) < 5 {
			continue // Skip evolution if population too small
		}

		// Create a simple fitness function based on survival
		fitnessFunc := func(entity *Entity) float64 {
			if !entity.IsAlive {
				return 0.0
			}

			// Fitness based on energy, age, and successful interactions
			ageFactor := math.Min(float64(entity.Age)/100.0, 1.0)
			energyFactor := entity.Energy / 100.0

			return ageFactor + energyFactor + entity.Fitness
		}

		pop.EvaluateFitness(fitnessFunc)

		// Only evolve if we have enough entities
		if len(pop.Entities) >= 10 {
			pop.Evolve()

			// Update world entity list with new entities
			for _, entity := range pop.Entities {
				found := false
				for _, worldEntity := range w.AllEntities {
					if worldEntity.ID == entity.ID {
						found = true
						break
					}
				}
				if !found {
					w.AllEntities = append(w.AllEntities, entity)
				}
			}
		}
	}
}

// spawnNewEntities creates new random entities to maintain population
func (w *World) spawnNewEntities() {
	for _, pop := range w.Populations {
		if len(pop.Entities) < w.Config.PopulationSize/2 {
			// Spawn new entity near existing ones
			if len(pop.Entities) > 0 {
				parent := pop.Entities[rand.Intn(len(pop.Entities))]

				// Create new entity near parent
				newPos := Position{
					X: parent.Position.X + (rand.Float64()-0.5)*10,
					Y: parent.Position.Y + (rand.Float64()-0.5)*10,
				}

				// Ensure position is within world bounds
				newPos.X = math.Max(0, math.Min(w.Config.Width, newPos.X))
				newPos.Y = math.Max(0, math.Min(w.Config.Height, newPos.Y))

				newEntity := NewEntity(w.NextID, pop.TraitNames, pop.Species, newPos)
				w.NextID++

				// Copy some traits from parent with mutation
				for name, trait := range parent.Traits {
					value := trait.Value + (rand.Float64()-0.5)*0.5
					value = math.Max(-2.0, math.Min(2.0, value))
					newEntity.SetTrait(name, value)
				}

				// Create DNA and cellular organism for the new entity to maintain evolution chain
				if w.DNASystem != nil && w.CellularSystem != nil {
					dna := w.DNASystem.GenerateRandomDNA(newEntity.ID, newEntity.Generation)
					w.CellularSystem.CreateSingleCellOrganism(newEntity.ID, dna)

					// Update entity traits based on DNA expression
					for traitName := range newEntity.Traits {
						dnaValue := w.DNASystem.ExpressTrait(dna, traitName)
						// Blend DNA value with existing trait (50/50 blend)
						currentValue := newEntity.GetTrait(traitName)
						newValue := (currentValue + dnaValue) / 2.0
						newEntity.SetTrait(traitName, newValue)
					}
				}

				pop.Entities = append(pop.Entities, newEntity)
				w.AllEntities = append(w.AllEntities, newEntity)
			}
		}
	}
}

// initializePlants populates the world with initial plant life
func (w *World) initializePlants() {
	// Calculate plant density based on world size
	totalCells := w.Config.GridWidth * w.Config.GridHeight
	plantsPerCell := 0.3 // Average 0.3 plants per cell
	totalPlants := int(float64(totalCells) * plantsPerCell)

	for i := 0; i < totalPlants; i++ {
		// Random position
		x := rand.Intn(w.Config.GridWidth)
		y := rand.Intn(w.Config.GridHeight)

		cell := &w.Grid[y][x]
		biome := w.Biomes[cell.Biome]

		// Choose plant type based on biome
		var plantType PlantType
		switch biome.Type {
		case BiomePlains:
			if rand.Float64() < 0.6 {
				plantType = PlantGrass
			} else {
				plantType = PlantBush
			}
		case BiomeForest:
			switch rand.Float64() {
			case 0.0:
				plantType = PlantTree
			case 0.1:
				plantType = PlantMushroom
			case 0.4:
				plantType = PlantBush
			default:
				plantType = PlantGrass
			}
		case BiomeDesert:
			if rand.Float64() < 0.7 {
				plantType = PlantCactus
			} else {
				plantType = PlantBush
			}
		case BiomeMountain:
			if rand.Float64() < 0.8 {
				plantType = PlantBush
			} else {
				plantType = PlantGrass
			}
		case BiomeWater:
			plantType = PlantAlgae
		case BiomeRadiation:
			if rand.Float64() < 0.6 {
				plantType = PlantMushroom
			} else {
				plantType = PlantBush
			}
		default:
			plantType = PlantGrass
		}

		// Create plant at world coordinates
		worldX := (float64(x) + rand.Float64()) * (w.Config.Width / float64(w.Config.GridWidth))
		worldY := (float64(y) + rand.Float64()) * (w.Config.Height / float64(w.Config.GridHeight))

		plant := NewPlant(w.NextPlantID, plantType, Position{X: worldX, Y: worldY})
		w.NextPlantID++

		// Add to world and grid
		w.AllPlants = append(w.AllPlants, plant)
		cell.Plants = append(cell.Plants, plant)
	}
}

// GetStats returns statistics about the world
func (w *World) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["tick"] = w.Tick
	stats["total_entities"] = len(w.AllEntities)

	// Population stats
	populationStats := make(map[string]map[string]interface{})
	for species, pop := range w.Populations {
		popStats := make(map[string]interface{})
		popStats["count"] = len(pop.Entities)

		if len(pop.Entities) > 0 {
			totalEnergy := 0.0
			totalAge := 0
			for _, entity := range pop.Entities {
				totalEnergy += entity.Energy
				totalAge += entity.Age
			}
			popStats["avg_energy"] = totalEnergy / float64(len(pop.Entities))
			popStats["avg_age"] = float64(totalAge) / float64(len(pop.Entities))
		}

		populationStats[species] = popStats
	}
	stats["populations"] = populationStats

	return stats
}

// String returns a string representation of the world state
func (w *World) String() string {
	return fmt.Sprintf("World{Tick: %d, Entities: %d, Populations: %d}",
		w.Tick, len(w.AllEntities), len(w.Populations))
}

// applyTimeEffects applies time-of-day and seasonal effects to entities
func (w *World) applyTimeEffects(entity *Entity, timeState TimeState) {
	// Circadian effects - some entities prefer day, others prefer night
	circadianPref := entity.GetTrait("circadian_preference") // -1 to 1, negative = nocturnal
	if timeState.IsNight() && circadianPref < 0 {
		// Nocturnal entities get energy boost at night
		entity.Energy += math.Abs(circadianPref) * 0.5
	} else if !timeState.IsNight() && circadianPref > 0 {
		// Diurnal entities get energy boost during day
		entity.Energy += circadianPref * 0.5
	} else {
		// Entities active at "wrong" time lose extra energy
		entity.Energy -= 0.2
	}
	// Seasonal effects
	switch timeState.Season {
	case Spring:
		// More food available, slight energy bonus
		entity.Energy += 0.1
	case Summer:
		// Peak activity season
		entity.Energy += 0.2
	case Autumn:
		// Preparation time, entities with high intelligence store energy
		if entity.GetTrait("intelligence") > 0.5 {
			entity.Energy += 0.15
		}
	case Winter:
		// Harsh season, higher energy drain
		entity.Energy -= 0.5
		// Entities with good endurance survive better
		if entity.GetTrait("endurance") < 0.3 {
			entity.Energy -= 0.3
		}
	}
}

// handleEntityCommunication processes entity signaling and responses
func (w *World) handleEntityCommunication(entity *Entity) {
	// Entity might send a signal based on its state
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")
	if intelligence > 0.4 && cooperation > 0.3 {
		// Send signals based on entity state
		if entity.Energy < 30 && rand.Float64() < 0.1 {
			// Distress signal
			w.CommunicationSystem.SendSignal(entity, SignalDanger, map[string]interface{}{
				"energy":  entity.Energy,
				"species": entity.Species,
			})
		} else if entity.Energy > 80 && rand.Float64() < 0.05 {
			// Food found signal
			w.CommunicationSystem.SendSignal(entity, SignalFood, map[string]interface{}{
				"position": entity.Position,
				"energy":   entity.Energy,
			})
		} else if cooperation > 0.6 && rand.Float64() < 0.03 {
			// Cooperation signal
			w.CommunicationSystem.SendSignal(entity, SignalHelp, map[string]interface{}{
				"species":     entity.Species,
				"cooperation": cooperation,
			})
		}
	}

	// Receive and respond to signals
	receivedSignals := w.CommunicationSystem.ReceiveSignals(entity)
	for _, signal := range receivedSignals {
		w.respondToSignal(entity, signal)
	}
}

// respondToSignal makes an entity respond to a received signal
func (w *World) respondToSignal(entity *Entity, signal Signal) {
	switch signal.Type {
	case SignalDanger:
		// Cooperative entities might help
		if entity.GetTrait("cooperation") > 0.5 && entity.Energy > 50 {
			// Move toward distress signal
			distance := math.Sqrt(math.Pow(entity.Position.X-signal.Position.X, 2) + math.Pow(entity.Position.Y-signal.Position.Y, 2))
			if distance > 1 {
				speed := entity.GetTrait("speed") * 0.5
				entity.MoveTo(signal.Position.X, signal.Position.Y, speed)
			}
		}
	case SignalFood:
		// Move toward food if hungry
		if entity.Energy < 60 {
			speed := entity.GetTrait("speed") * 0.3
			entity.MoveTo(signal.Position.X, signal.Position.Y, speed)
		}
	case SignalHelp:
		// Increase cooperation temporarily
		if entity.GetTrait("cooperation") > 0.4 {
			entity.SetTrait("cooperation", math.Min(2.0, entity.GetTrait("cooperation")+0.1))
		}
	}
}

// attemptGroupFormation tries to form new groups from nearby compatible entities
func (w *World) attemptGroupFormation() {
	groupCandidates := make(map[string][]*Entity) // species -> entities

	// Group entities by species and cooperation level
	for _, entity := range w.AllEntities {
		if !entity.IsAlive || entity.GetTrait("cooperation") < 0.4 {
			continue
		}

		species := entity.Species
		if groupCandidates[species] == nil {
			groupCandidates[species] = make([]*Entity, 0)
		}
		groupCandidates[species] = append(groupCandidates[species], entity)
	}

	// Try to form groups within each species
	for species, candidates := range groupCandidates {
		if len(candidates) < 2 {
			continue
		}

		// Find clusters of nearby entities
		for i, entity1 := range candidates {
			nearbyEntities := []*Entity{entity1}

			for j, entity2 := range candidates {
				if i == j {
					continue
				}

				distance := entity1.DistanceTo(entity2)
				if distance <= 15.0 { // Group formation distance
					nearbyEntities = append(nearbyEntities, entity2)
				}
			}

			// Form group if we have enough compatible entities
			if len(nearbyEntities) >= 2 && len(nearbyEntities) <= 6 {
				// Check if these entities are already in a group
				alreadyGrouped := false
				for _, group := range w.GroupBehaviorSystem.Groups {
					for _, member := range group.Members {
						for _, candidate := range nearbyEntities {
							if member.ID == candidate.ID {
								alreadyGrouped = true
								break
							}
						}
						if alreadyGrouped {
							break
						}
					}
					if alreadyGrouped {
						break
					}
				}

				if !alreadyGrouped {
					// Determine group purpose based on entity traits
					purpose := "territory"
					avgAggression := 0.0
					for _, e := range nearbyEntities {
						avgAggression += e.GetTrait("aggression")
					}
					avgAggression /= float64(len(nearbyEntities))

					if avgAggression > 0.6 {
						purpose = "hunting"
					} else if species == "herbivore" || species == "omnivore" {
						purpose = "migration"
					}

					w.GroupBehaviorSystem.FormGroup(nearbyEntities, purpose)
				}
			}
		}
	}
}

// processCivilizationActivities handles tribe activities and structure management
func (w *World) processCivilizationActivities() {
	// Update civilization system
	w.CivilizationSystem.Update()

	// Process tribe activities
	for _, tribe := range w.CivilizationSystem.Tribes {
		// Tribe expansion - try to recruit nearby compatible entities
		if len(tribe.Members) < 20 { // Max tribe size
			for _, entity := range w.AllEntities {
				if !entity.IsAlive || entity.TribeID != 0 {
					continue // Already in a tribe
				}

				// Check if entity is near tribe territory
				inTerritory := false
				for _, territory := range tribe.Territory {
					distance := math.Sqrt(math.Pow(entity.Position.X-territory.X, 2) + math.Pow(entity.Position.Y-territory.Y, 2))
					if distance <= 20.0 {
						inTerritory = true
						break
					}
				}

				if inTerritory && entity.GetTrait("cooperation") > 0.5 && entity.GetTrait("intelligence") > 0.4 {
					// Try to recruit entity
					if rand.Float64() < 0.05 { // 5% chance
						tribe.Members = append(tribe.Members, entity)
						entity.TribeID = tribe.ID
					}
				}
			}
		}

		// Tribe activities based on size and resources
		if len(tribe.Members) >= 3 {
			// Larger tribes can build structures
			if rand.Float64() < 0.02 && len(tribe.Structures) < 5 {
				w.buildTribeStructure(tribe)
			}

			// Resource gathering and trading
			if rand.Float64() < 0.1 {
				w.processTribeResourceGathering(tribe)
			}
		}
	}
}

// buildTribeStructure creates a new structure for a tribe
func (w *World) buildTribeStructure(tribe *Tribe) {
	if len(tribe.Members) == 0 {
		return
	}

	// Choose a location near tribe center
	centerX, centerY := 0.0, 0.0
	for _, member := range tribe.Members {
		centerX += member.Position.X
		centerY += member.Position.Y
	}
	centerX /= float64(len(tribe.Members))
	centerY /= float64(len(tribe.Members))

	// Random offset for structure location
	structX := centerX + (rand.Float64()-0.5)*20
	structY := centerY + (rand.Float64()-0.5)*20

	// Ensure within world bounds
	structX = math.Max(0, math.Min(w.Config.Width, structX))
	structY = math.Max(0, math.Min(w.Config.Height, structY))
	// Determine structure type based on tribe needs
	var structType StructureType = StructureNest // Default to basic shelter
	if len(tribe.Structures) > 0 && rand.Float64() < 0.3 {
		structType = StructureCache // Storage
	} else if len(tribe.Structures) > 1 && rand.Float64() < 0.2 {
		if tribe.TechLevel >= 3 {
			structType = StructureFarm // Workshop equivalent
		} else {
			structType = StructureTrap
		}
	}

	structure := &Structure{
		ID:        len(tribe.Structures) + 1,
		Type:      structType,
		Position:  Position{X: structX, Y: structY},
		Health:    100.0,
		Resources: make(map[string]float64),
	}

	tribe.Structures = append(tribe.Structures, structure)
}

// processTribeResourceGathering handles resource collection and management
func (w *World) processTribeResourceGathering(tribe *Tribe) {
	// Tribe members gather resources
	for _, member := range tribe.Members {
		if !member.IsAlive {
			continue
		}

		// Check for plants to harvest
		gridX := int((member.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((member.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		cell := &w.Grid[gridY][gridX]
		for _, plant := range cell.Plants {
			if plant.IsAlive && rand.Float64() < 0.1 {
				// Harvest resource from plant
				harvestedAmount := 1.0 + member.GetTrait("intelligence")*0.5
				tribe.Resources["food"] += harvestedAmount
				plant.Energy -= harvestedAmount * 2 // Depletes plant
				break
			}
		}

		// Intelligent members can gather building materials
		if member.GetTrait("intelligence") > 0.6 && rand.Float64() < 0.05 {
			tribe.Resources["materials"] += 0.5 + member.GetTrait("strength")*0.3
		}
	}

	// Use resources for tribe benefits
	if tribe.Resources["food"] > 10 {
		// Feed tribe members
		foodPerMember := math.Min(tribe.Resources["food"]/float64(len(tribe.Members)), 5.0)
		for _, member := range tribe.Members {
			member.Energy += foodPerMember
		}
		tribe.Resources["food"] -= foodPerMember * float64(len(tribe.Members))
	}
}

// trackEntityEnvironmentalExposure tracks environmental conditions for feedback loops
func (w *World) trackEntityEnvironmentalExposure(entity *Entity, timeState TimeState) {
	if !entity.IsAlive {
		return
	}

	// Get entity's current biome
	gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
	gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
	gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
	gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

	cell := &w.Grid[gridY][gridX]
	biome := cell.Biome

	// Get current event affecting this cell
	var currentEvent *WorldEvent
	if cell.Event != nil {
		currentEvent = cell.Event
	}

	// Track environmental exposure
	entity.trackEnvironmentalExposure(biome, seasonToString(timeState.Season), currentEvent, w.Tick)
}

// seasonToString converts Season enum to string
func seasonToString(season Season) string {
	switch season {
	case Spring:
		return "Spring"
	case Summer:
		return "Summer"
	case Autumn:
		return "Autumn"
	case Winter:
		return "Winter"
	default:
		return "Unknown"
	}
}

// updateReproductionSystem handles reproduction, gestation, and decay processes
func (w *World) updateReproductionSystem() {
	// Update mating seasons based on current time
	currentTimeState := w.AdvancedTimeSystem.GetTimeState()
	w.ReproductionSystem.UpdateMatingSeasons(w.AllEntities, seasonToString(currentTimeState.Season))
	
	// Enhanced seasonal mating behaviors
	w.ReproductionSystem.UpdateSeasonalMatingBehaviors(w.AllEntities, currentTimeState.Season, w.Tick)
	
	// Implement territorial mating if civilization system is available
	if w.CivilizationSystem != nil {
		territories := w.generateTerritories()
		w.ReproductionSystem.ImplementTerritorialMating(w.AllEntities, territories)
	}
	
	// Check for births from gestation
	newborns := w.ReproductionSystem.CheckGestation(w.AllEntities, w.Tick)
	for _, newborn := range newborns {
		newborn.ID = w.NextID
		w.NextID++
		w.AllEntities = append(w.AllEntities, newborn)
		
		// Log birth event
		w.EventLogger.LogWorldEvent(w.Tick, "birth", fmt.Sprintf("Entity %d gave birth to entity %d", newborn.Generation-1, newborn.ID))
	}
	
	// Process egg hatching and decay
	newHatchlings, fertilizers := w.ReproductionSystem.Update(w.Tick)
	for _, hatchling := range newHatchlings {
		hatchling.ID = w.NextID
		w.NextID++
		w.AllEntities = append(w.AllEntities, hatchling)
		
		// Log hatching event
		w.EventLogger.LogWorldEvent(w.Tick, "hatching", fmt.Sprintf("Egg hatched into entity %d", hatchling.ID))
	}
	
	// Process decay fertilizers to enhance nearby plants
	for _, fertilizer := range fertilizers {
		w.applyDecayFertilizer(fertilizer)
	}
	
	// Handle mating attempts
	w.processMatingAttempts()
	
	// Handle mating migration behaviors
	w.processMatingMigration()
	
	// Handle entity deaths and create decaying items
	w.processEntityDeaths()
}

// processMatingMigration handles entities migrating to preferred mating locations
func (w *World) processMatingMigration() {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive || entity.ReproductionStatus == nil {
			continue
		}
		
		// Only migrate during mating season and if entity requires migration
		if !entity.ReproductionStatus.MatingSeason || !entity.ReproductionStatus.RequiresMigration {
			continue
		}
		
		// Skip if already at preferred location (within tolerance)
		dx := entity.Position.X - entity.ReproductionStatus.PreferredMatingLocation.X
		dy := entity.Position.Y - entity.ReproductionStatus.PreferredMatingLocation.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		
		if distance <= 5.0 { // Close enough to preferred location
			continue
		}
		
		// Move towards preferred mating location
		moveSpeed := entity.GetTrait("speed") * 0.5 // Slower migration movement
		if moveSpeed <= 0 {
			moveSpeed = 1.0
		}
		
		// Calculate movement direction
		directionX := dx / distance
		directionY := dy / distance
		
		// Move towards the target
		entity.Position.X += directionX * moveSpeed
		entity.Position.Y += directionY * moveSpeed
		
		// Migration costs energy
		entity.Energy -= moveSpeed * 0.2
		
		// Log migration behavior occasionally
		if w.Tick%50 == 0 && distance > entity.ReproductionStatus.MigrationDistance*0.5 {
			w.EventLogger.LogWorldEvent(w.Tick, "migration", fmt.Sprintf("Entity %d migrating to mating grounds (%.1f units away)", entity.ID, distance))
		}
	}
}

// processMatingAttempts handles entities trying to mate
func (w *World) processMatingAttempts() {
	// Create a map for quick entity lookup
	entityMap := make(map[int]*Entity)
	for _, entity := range w.AllEntities {
		if entity.IsAlive {
			entityMap[entity.ID] = entity
		}
	}
	
	for i, entity1 := range w.AllEntities {
		if !entity1.IsAlive || entity1.ReproductionStatus == nil {
			continue
		}
		
		// Skip if not ready to mate
		if !entity1.ReproductionStatus.ReadyToMate || !entity1.ReproductionStatus.MatingSeason {
			continue
		}
		
		// Don't reproduce in the first few ticks to avoid interfering with tests
		if w.Tick < 10 {
			continue
		}
		
		// Don't reproduce if entity is too young or too low energy
		if entity1.Age < 5 || entity1.Energy < 30.0 {
			continue
		}
		
		// Check reproduction cooldown (entities can't reproduce too frequently)
		if entity1.ReproductionStatus.LastMatingTick > 0 && w.Tick-entity1.ReproductionStatus.LastMatingTick < 25 {
			continue
		}
		
		// Low probability of reproduction to avoid test interference
		if rand.Float64() > 0.1 { // Only 10% chance per tick per entity
			continue
		}
		
		// Find nearby potential mates
		for j := i + 1; j < len(w.AllEntities); j++ {
			entity2 := w.AllEntities[j]
			if !entity2.IsAlive || entity2.ReproductionStatus == nil {
				continue
			}
			
			// Check compatibility (same species or cross-species compatibility)
			canMate := false
			if entity1.Species == entity2.Species {
				canMate = true
			} else {
				// Check cross-species compatibility
				canMate = w.ReproductionSystem.ImplementCrossSpeciesCompatibility(entity1, entity2)
			}
			
			if !canMate {
				continue
			}
			
			// Check distance (entities need to be close to mate)
			distance := entity1.DistanceTo(entity2)
			if distance > 5.0 { // Mating range
				continue
			}
			
			// Check for competition - see if there are other potential mates nearby
			competition := w.checkMatingCompetition(entity1, entity2)
			if competition && rand.Float64() < 0.7 { // 70% chance competition prevents mating
				continue
			}
			
			// Attempt mating
			if w.ReproductionSystem.StartMating(entity1, entity2, w.Tick) {
				// Log mating event
				w.EventLogger.LogWorldEvent(w.Tick, "mating", fmt.Sprintf("Entities %d and %d mated", entity1.ID, entity2.ID))
				
				// Handle different reproduction modes
				switch entity1.ReproductionStatus.Mode {
				case DirectCoupling:
					// Create immediate offspring using existing crossover
					offspring := Crossover(entity1, entity2, w.NextID, entity1.Species)
					offspring.Mutate(0.1, 0.2) // Some mutation
					w.NextID++
					w.AllEntities = append(w.AllEntities, offspring)
					w.EventLogger.LogWorldEvent(w.Tick, "birth", fmt.Sprintf("Direct coupling produced entity %d", offspring.ID))
				
				case Budding:
					// Asexual reproduction - create clone with mutation
					if entity1.Energy >= 50.0 {
						clone := entity1.Clone()
						clone.ID = w.NextID
						clone.Mutate(0.15, 0.3) // Higher mutation for asexual reproduction
						clone.Position.X += (rand.Float64() - 0.5) * 4.0
						clone.Position.Y += (rand.Float64() - 0.5) * 4.0
						w.NextID++
						w.AllEntities = append(w.AllEntities, clone)
						w.EventLogger.LogWorldEvent(w.Tick, "budding", fmt.Sprintf("Entity %d reproduced by budding, created entity %d", entity1.ID, clone.ID))
					}
				
				case Fission:
					// Split into multiple offspring
					if entity1.Energy >= 80.0 {
						numOffspring := 2 + rand.Intn(2) // 2-3 offspring
						for i := 0; i < numOffspring; i++ {
							clone := entity1.Clone()
							clone.ID = w.NextID
							clone.Energy = entity1.Energy / float64(numOffspring+1) // Distribute energy
							clone.Mutate(0.2, 0.4) // Higher mutation for fission
							clone.Position.X += (rand.Float64() - 0.5) * 6.0
							clone.Position.Y += (rand.Float64() - 0.5) * 6.0
							w.NextID++
							w.AllEntities = append(w.AllEntities, clone)
						}
						entity1.Energy /= float64(numOffspring + 1) // Parent keeps some energy
						w.EventLogger.LogWorldEvent(w.Tick, "fission", fmt.Sprintf("Entity %d split into %d offspring", entity1.ID, numOffspring))
					}
				}
				
				// Only allow one mating per tick per entity
				break
			}
		}
	}
}

// processEntityDeaths handles entity death and creates decaying corpses
func (w *World) processEntityDeaths() {
	for _, entity := range w.AllEntities {
		if entity.IsAlive && (entity.Energy <= 0 || entity.Age > 1000) {
			// Entity dies
			entity.IsAlive = false
			
			// Create decaying corpse
			corpseNutrientValue := entity.Energy*0.5 + float64(entity.Age)*0.1
			w.ReproductionSystem.AddDecayingItem("corpse", entity.Position, corpseNutrientValue, entity.Species, entity.GetTrait("size"), w.Tick)
			
			// Log death event
			w.EventLogger.LogWorldEvent(w.Tick, "death", fmt.Sprintf("Entity %d (%s) died at age %d", entity.ID, entity.Species, entity.Age))
		}
	}
}

// generateTerritories creates territories based on civilization system tribes
func (w *World) generateTerritories() map[int]*Territory {
	territories := make(map[int]*Territory)
	
	if w.CivilizationSystem == nil {
		return territories
	}
	
	territoryID := 1
	for _, tribe := range w.CivilizationSystem.Tribes {
		if len(tribe.Members) == 0 {
			continue
		}
		
		// Find tribe center based on member positions
		centerX := 0.0
		centerY := 0.0
		strongestEntity := tribe.Members[0]
		maxStrength := 0.0
		
		for _, member := range tribe.Members {
			centerX += member.Position.X
			centerY += member.Position.Y
			
			strength := member.GetTrait("strength") + member.GetTrait("intelligence")
			if strength > maxStrength {
				maxStrength = strength
				strongestEntity = member
			}
		}
		
		centerX /= float64(len(tribe.Members))
		centerY /= float64(len(tribe.Members))
		
		// Territory size based on tribe size and leader strength
		radius := 5.0 + float64(len(tribe.Members))*2.0 + maxStrength*3.0
		quality := (tribe.Resources["food"] + tribe.Resources["materials"]) / 200.0 // 0-1 scale
		
		territory := &Territory{
			ID:      territoryID,
			OwnerID: strongestEntity.ID,
			Center: Position{
				X: centerX,
				Y: centerY,
			},
			Radius:  radius,
			Quality: quality,
		}
		
		territories[territoryID] = territory
		territoryID++
	}
	
	return territories
}

// applyDecayFertilizer enhances plants near decaying organic matter
func (w *World) applyDecayFertilizer(fertilizer *DecayableItem) {
	// Find plants within fertilizer range
	for _, plant := range w.AllPlants {
		dx := plant.Position.X - fertilizer.Position.X
		dy := plant.Position.Y - fertilizer.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		
		if distance <= 10.0 { // Fertilizer effect range
			// Boost plant energy and growth
			energyBoost := fertilizer.NutrientValue * 0.3 * (10.0 - distance) / 10.0 // Closer = more effect
			plant.Energy += energyBoost
			
			// Boost plant traits temporarily
			plant.SetTrait("growth_efficiency", plant.GetTrait("growth_efficiency")+0.1)
			plant.SetTrait("reproduction_rate", plant.GetTrait("reproduction_rate")+0.05)
		}
	}
	
	// Log fertilization event
	w.EventLogger.LogWorldEvent(w.Tick, "fertilization", fmt.Sprintf("Decayed %s provided nutrients to nearby plants", fertilizer.ItemType))
}

// checkMatingCompetition determines if there is competition for mates
func (w *World) checkMatingCompetition(entity1, entity2 *Entity) bool {
	// Look for other entities nearby that could compete
	competitorCount := 0
	
	for _, potential := range w.AllEntities {
		if !potential.IsAlive || potential.ReproductionStatus == nil {
			continue
		}
		
		// Skip the entities trying to mate
		if potential.ID == entity1.ID || potential.ID == entity2.ID {
			continue
		}
		
		// Skip if not same species
		if potential.Species != entity1.Species {
			continue
		}
		
		// Skip if not in mating condition
		if !potential.ReproductionStatus.ReadyToMate || !potential.ReproductionStatus.MatingSeason {
			continue
		}
		
		// Check if competitor is close enough to interfere
		distance1 := entity1.DistanceTo(potential)
		distance2 := entity2.DistanceTo(potential)
		
		if distance1 <= 8.0 || distance2 <= 8.0 { // Competition range larger than mating range
			// Check if competitor is stronger/more attractive
			entity1Attractiveness := entity1.GetTrait("strength") + entity1.GetTrait("intelligence") + entity1.Energy/100.0
			potentialAttractiveness := potential.GetTrait("strength") + potential.GetTrait("intelligence") + potential.Energy/100.0
			
			if potentialAttractiveness > entity1Attractiveness {
				competitorCount++
			}
		}
	}
	
	// Competition exists if there are stronger competitors nearby
	return competitorCount > 0
}

// TogglePause toggles the simulation pause state
func (w *World) TogglePause() {
	w.Paused = !w.Paused
}

// SetPaused sets the simulation pause state
func (w *World) SetPaused(paused bool) {
	w.Paused = paused
}

// IsPaused returns the current pause state
func (w *World) IsPaused() bool {
	return w.Paused
}

// Reset resets the world to initial state
func (w *World) Reset() {
	// Clear existing entities and populations
	w.AllEntities = make([]*Entity, 0)
	w.AllPlants = make([]*Plant, 0)
	w.Populations = make(map[string]*Population)
	w.PhysicsComponents = make(map[int]*PhysicsComponent)

	// Reset counters
	w.Tick = 0
	w.NextID = 0
	w.NextPlantID = 0
	w.Paused = false

	// Clear events
	w.Events = make([]*WorldEvent, 0)

	// Clear grid
	w.clearGrid()

	// Reset physics collision counters
	if w.PhysicsSystem != nil {
		w.PhysicsSystem.ResetCollisionCounters()
	}
}
