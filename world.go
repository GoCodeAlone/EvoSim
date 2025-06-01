package main

import (
	"fmt"
	"math"
	"math/rand"
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
	}

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
	case noise < 0.4:
		return BiomeDesert
	case noise < 0.6:
		return BiomeMountain
	case noise < 0.8:
		return BiomeWater
	case noise < 0.95:
		return BiomePlains
	default:
		return BiomeRadiation
	}
}

// AddPopulation adds a new population to the world
func (w *World) AddPopulation(config PopulationConfig) {
	// Generate trait names based on base traits
	traitNames := make([]string, 0, len(config.BaseTraits))
	for name := range config.BaseTraits {
		traitNames = append(traitNames, name)
	}

	// Create population with species-specific mutation rate
	pop := NewPopulation(w.Config.PopulationSize, traitNames, config.BaseMutationRate, 0.2)
	pop.Species = config.Species

	// Initialize entities with base traits and positions
	for _, entity := range pop.Entities {
		// Set position around start position
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * config.Spread

		entity.Position = Position{
			X: config.StartPos.X + math.Cos(angle)*distance,
			Y: config.StartPos.Y + math.Sin(angle)*distance,
		}
		entity.Species = config.Species
		entity.ID = w.NextID
		w.NextID++

		// Apply base traits with some variation
		for traitName, baseValue := range config.BaseTraits {
			variation := (rand.Float64() - 0.5) * 0.4 // ±20% variation
			value := baseValue + variation
			value = math.Max(-2.0, math.Min(2.0, value))
			entity.SetTrait(traitName, value)
		}

		w.AllEntities = append(w.AllEntities, entity)
	}

	w.Populations[config.Species] = pop
}

// Update simulates one tick of the world
func (w *World) Update() {
	w.Tick++
	now := time.Now()
	w.Clock = w.Clock.Add(time.Hour) // Each tick = 1 hour world time
	w.LastUpdate = now

	// Clear grid entities and plants
	w.clearGrid()

	// Update world events
	w.updateEvents()

	// Maybe trigger new events
	if rand.Float64() < 0.01 { // 1% chance per tick
		w.triggerRandomEvent()
	}

	// Update all plants
	w.updatePlants()

	// Update all entities with biome effects and starvation checks
	for _, entity := range w.AllEntities {
		w.updateEntityWithBiome(entity)
		entity.CheckStarvation(w) // Check for starvation-driven evolution
		entity.Update()
	}

	// Update grid with current entity and plant positions
	w.updateGrid()

	// Handle interactions between entities and with plants
	w.handleInteractions()

	// Remove dead entities and plants
	w.removeDeadEntities()
	w.removeDeadPlants()

	// Plant reproduction
	if w.Tick%10 == 0 {
		w.reproducePlants()
	}

	// Population-level evolution (less frequent)
	if w.Tick%50 == 0 {
		w.evolvePopulations()
	}

	// Spawn new entities occasionally (based on carrying capacity)
	if w.Tick%20 == 0 {
		w.spawnNewEntities()
	}

	// Update event logger with population changes
	w.EventLogger.UpdatePopulationCounts(w.Tick, w.Populations)
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

	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}

		if offspring := plant.Reproduce(w.NextPlantID); offspring != nil {
			w.NextPlantID++
			newPlants = append(newPlants, offspring)
		}
	}

	// Add new plants to world
	w.AllPlants = append(w.AllPlants, newPlants...)

	// Log significant plant population changes
	if len(newPlants) > 5 {
		w.EventLogger.LogEcosystemShift(w.Tick,
			fmt.Sprintf("Plant reproduction boom: %d new plants born", len(newPlants)),
			map[string]interface{}{"new_plants": len(newPlants)})
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
				if entity.EatPlant(plant) {
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
		entity1.Eat(entity2)
	} else if !entity1.IsAlive && entity2.CanEat(entity1) && rand.Float64() < 0.3 {
		entity2.Eat(entity1)
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
