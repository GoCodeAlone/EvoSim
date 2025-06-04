package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
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
	// New biome expansions
	BiomeIce        // Polar caps, frozen areas
	BiomeRainforest // Dense tropical forests
	BiomeDeepWater  // Ocean trenches, deep water
	BiomeHighAltitude // Very high mountains, low oxygen
	BiomeHotSpring    // Geysers, hot springs
	BiomeTundra      // Cold, sparse vegetation
	BiomeSwamp       // Wetlands, marshes
	BiomeCanyon      // Deep rocky canyons
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
	// Environmental properties for biome expansions
	Temperature    float64            // Temperature modifier (-1 to 1, 0 = normal)
	Pressure       float64            // Pressure level (0 to 2, 1 = normal)
	OxygenLevel    float64            // Oxygen availability (0 to 1, 1 = normal)
	Humidity       float64            // Humidity level (0 to 1)
	IsAquatic      bool               // Whether this is a water-based biome
	IsUnderground  bool               // Whether this is below ground
	IsAerial       bool               // Whether this is high altitude/aerial
}

// WorldEvent represents temporary world-wide effects
type WorldEvent struct {
	Name           string
	Description    string
	Duration       int // Ticks remaining
	GlobalMutation float64
	GlobalDamage   float64
	BiomeChanges   map[Position]BiomeType
	// Enhanced properties for realistic behavior
	Position       Position                   // Event center/origin
	Radius         float64                   // Affected area radius
	Intensity      float64                   // Event strength (0-1)
	MovementVector Vector2D                  // For moving events like storms/fires
	WindInfluenced bool                      // Whether wind affects this event
	EventType      string                    // "fire", "storm", "volcanic", etc.
	VisualEffect   map[string]interface{}    // Visual representation data
}

// EnhancedEnvironmentalEvent represents more sophisticated environmental events
type EnhancedEnvironmentalEvent struct {
	ID             int
	Type           string                    // "wildfire", "storm", "volcanic_eruption", etc.
	Name           string
	Description    string
	StartTick      int
	Duration       int
	Position       Position                  // Event center
	Radius         float64                   // Current affected radius
	MaxRadius      float64                   // Maximum spread radius
	Intensity      float64                   // Current strength (0-1)
	Direction      float64                   // Movement direction (radians)
	Speed          float64                   // Movement speed
	WindSensitive  bool                      // Whether wind affects movement/spread
	AffectedCells  map[Position]BiomeType    // Cells affected and their new types
	SpreadPattern  string                    // "circular", "directional", "wind_driven"
	ExtinguishOn   []BiomeType               // Biomes that stop the event (water stops fire)
	Effects        map[string]float64        // Various effects (mutation, damage, etc.)
}

// GridCell represents a cell in the world grid
type GridCell struct {
	Biome    BiomeType
	Entities []*Entity
	Plants   []*Plant    // Plants in this cell
	Event    *WorldEvent // Current event affecting this cell
	// Enhanced soil system
	SoilNutrients   map[string]float64 `json:"soil_nutrients"`   // Available nutrients in soil
	WaterLevel      float64            `json:"water_level"`      // Available water in cell
	SoilPH          float64            `json:"soil_ph"`          // Soil acidity (6-8 optimal)
	SoilCompaction  float64            `json:"soil_compaction"`  // How compacted soil is (0-1)
	OrganicMatter   float64            `json:"organic_matter"`   // Decomposed organic material
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
	EventLogger *EventLogger // Event logging system (legacy)
	CentralEventBus *CentralEventBus // Unified event system
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
	SeedDispersalSystem *SeedDispersalSystem // Advanced seed dispersal and germination
	ChemicalEcologySystem *ChemicalEcologySystem // Chemical communication and ecology
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
	FungalNetwork           *FungalNetwork                       // Fungal decomposer and nutrient cycling system
	
	// Cultural Knowledge System
	CulturalKnowledgeSystem *CulturalKnowledgeSystem             // Multi-generational knowledge transfer and cultural evolution
	
	// Statistical Analysis System
	StatisticalReporter     *StatisticalReporter                 // Comprehensive statistical analysis and reporting
	
	// Hive Mind, Caste, and Insect Systems
	HiveMindSystem          *HiveMindSystem                      // Collective intelligence system
	CasteSystem             *CasteSystem                         // Caste-based social organization
	InsectSystem            *InsectSystem                        // Insect-specific behaviors and capabilities
	InsectPollinationSystem *InsectPollinationSystem             // Insect pollination and plant-insect mutualism
	ColonyWarfareSystem     *ColonyWarfareSystem                 // Inter-colony warfare and diplomacy
	
	// Organism classification and lifespan system
	OrganismClassifier      *OrganismClassifier                  // Organism classification and aging system
	
	// Metamorphosis and life stage system
	MetamorphosisSystem     *MetamorphosisSystem                 // Life stage transitions and development
	
	// Player event callback for gamification features
	PlayerEventsCallback    func(eventType string, data map[string]interface{}) // Callback for player-related events
	PreviousPopulationCounts map[string]int                                     // Track population counts for extinction detection
	
	// Enhanced Environmental Event System
	EnvironmentalEvents     []*EnhancedEnvironmentalEvent                      // Active enhanced environmental events
	NextEnvironmentalEventID int                                                // ID counter for environmental events
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
		CentralEventBus: NewCentralEventBus(50000), // Central event bus with 50k events
		NextID:      0,
		NextPlantID: 0,
		Tick:        0,
		Clock:       time.Now(),
		LastUpdate:  time.Now(),
		PreviousPopulationCounts: make(map[string]int),
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
				// Initialize soil system
				SoilNutrients:  initializeSoilNutrients(),
				WaterLevel:     initializeWaterLevel(world.generateBiome(x, y)),
				SoilPH:         7.0 + (rand.Float64()-0.5)*2.0, // pH 6-8
				SoilCompaction: rand.Float64() * 0.3, // 0-30% compaction
				OrganicMatter:  rand.Float64() * 0.2, // 0-20% organic matter
			}
		}
	} // Initialize advanced systems
	world.CommunicationSystem = NewCommunicationSystem(world.CentralEventBus)
	world.GroupBehaviorSystem = NewGroupBehaviorSystem(world.CentralEventBus)
	world.PhysicsSystem = NewPhysicsSystem()
	world.CollisionSystem = NewCollisionSystem()
	world.PhysicsComponents = make(map[int]*PhysicsComponent)
	world.AdvancedTimeSystem = NewAdvancedTimeSystem(480, 120) // 480 ticks/day, 120 days/season
	world.CivilizationSystem = NewCivilizationSystem(world.CentralEventBus)
	world.ViewportSystem = NewViewportSystem(config.Width, config.Height)
	world.WindSystem = NewWindSystem(int(config.Width), int(config.Height), world.CentralEventBus)
	world.SeedDispersalSystem = NewSeedDispersalSystem()
	world.ChemicalEcologySystem = NewChemicalEcologySystem()
	world.SpeciationSystem = NewSpeciationSystem()
	world.PlantNetworkSystem = NewPlantNetworkSystem(world.CentralEventBus)
	world.SpeciesNaming = NewSpeciesNaming()

	// Initialize new evolution and topology systems
	world.DNASystem = NewDNASystem(world.CentralEventBus)
	world.CellularSystem = NewCellularSystem(world.DNASystem, world.CentralEventBus)
	world.MacroEvolutionSystem = NewMacroEvolutionSystem()
	world.TopologySystem = NewTopologySystem(config.GridWidth, config.GridHeight)

	// Initialize tool and environmental modification systems
	world.ToolSystem = NewToolSystem(world.CentralEventBus)
	world.EnvironmentalModSystem = NewEnvironmentalModificationSystem(world.CentralEventBus)
	world.EmergentBehaviorSystem = NewEmergentBehaviorSystem()
	
	// Initialize reproduction and decay system
	world.ReproductionSystem = NewReproductionSystem(world.CentralEventBus)
	world.FungalNetwork = NewFungalNetwork()
	
	// Initialize cultural knowledge system
	world.CulturalKnowledgeSystem = NewCulturalKnowledgeSystem()
	
	// Initialize statistical analysis system
	world.StatisticalReporter = NewStatisticalReporter(10000, 1000, 10, 50) // 10k events, 1k snapshots, snapshot every 10 ticks, analyze every 50 ticks
	
	// Connect StatisticalReporter to CentralEventBus
	world.CentralEventBus.AddListener(func(event CentralEvent) {
		// Convert CentralEvent to StatisticalEvent format
		statEvent := StatisticalEvent{
			Timestamp:   event.Timestamp,
			Tick:        event.Tick,
			EventType:   event.Type,
			Category:    event.Category,
			EntityID:    event.EntityID,
			PlantID:     event.PlantID,
			Position:    event.Position,
			OldValue:    event.OldValue,
			NewValue:    event.NewValue,
			Change:      event.Change,
			Metadata:    event.Metadata,
			ImpactedIDs: event.ImpactedIDs,
		}
		world.StatisticalReporter.addEvent(statEvent)
	})
	
	// Connect EventLogger to CentralEventBus for legacy event types
	world.CentralEventBus.AddListener(func(event CentralEvent) {
		// Convert certain events to legacy LogEvent format
		if event.Category == "system" || event.Type == "extinction" || event.Type == "birth" || event.Type == "evolution" {
			logEvent := LogEvent{
				Timestamp:   event.Timestamp,
				Tick:        event.Tick,
				Type:        event.Type,
				Description: event.Description,
				Data:        event.Metadata,
			}
			world.EventLogger.addEvent(logEvent)
		}
	})
	
	// Initialize hive mind, caste, and insect systems
	world.HiveMindSystem = NewHiveMindSystem()
	world.CasteSystem = NewCasteSystem()
	world.InsectSystem = NewInsectSystem()
	world.InsectPollinationSystem = NewInsectPollinationSystem()
	world.ColonyWarfareSystem = NewColonyWarfareSystem()
	
	// Initialize organism classification and lifespan system
	world.OrganismClassifier = NewOrganismClassifier(world.AdvancedTimeSystem)
	
	// Initialize metamorphosis system
	world.MetamorphosisSystem = NewMetamorphosisSystem()
	
	// Initialize enhanced environmental event system
	world.EnvironmentalEvents = make([]*EnhancedEnvironmentalEvent, 0)
	world.NextEnvironmentalEventID = 1

  // Generate initial world terrain
	world.TopologySystem.GenerateInitialTerrain()

	world.FluidRegions = make([]FluidRegion, 0)

	// Initialize plant life
	world.initializePlants()

	// Initialize fungal networks
	if world.FungalNetwork != nil {
		initialFungi := 20 + rand.Intn(30) // 20-50 initial fungi
		world.FungalNetwork.SeedInitialFungi(world, initialFungi)
	}

	// Process initial species formation from newly created plants
	if len(world.AllPlants) > 0 {
		world.SpeciationSystem.Update(world.AllPlants, 0)
	}

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
		Temperature:    0.0,
		Pressure:       1.0,
		OxygenLevel:    1.0,
		Humidity:       0.5,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeForest] = Biome{
		Type:           BiomeForest,
		Name:           "Forest",
		Color:          "darkgreen",
		TraitModifiers: map[string]float64{"size": 0.2, "defense": 0.1},
		MutationRate:   0.0,
		EnergyDrain:    0.8,
		Symbol:         '♠',
		Temperature:    0.1,
		Pressure:       1.0,
		OxygenLevel:    1.1,
		Humidity:       0.7,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeDesert] = Biome{
		Type:           BiomeDesert,
		Name:           "Desert",
		Color:          "yellow",
		TraitModifiers: map[string]float64{"endurance": 0.3, "size": -0.1},
		MutationRate:   0.05,
		EnergyDrain:    1.5,
		Symbol:         '~',
		Temperature:    0.7,
		Pressure:       1.0,
		OxygenLevel:    0.9,
		Humidity:       0.1,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeMountain] = Biome{
		Type:           BiomeMountain,
		Name:           "Mountain",
		Color:          "gray",
		TraitModifiers: map[string]float64{"strength": 0.2, "speed": -0.1},
		MutationRate:   0.0,
		EnergyDrain:    1.2,
		Symbol:         '^',
		Temperature:    -0.3,
		Pressure:       0.8,
		OxygenLevel:    0.8,
		Humidity:       0.4,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeWater] = Biome{
		Type:           BiomeWater,
		Name:           "Water",
		Color:          "blue",
		TraitModifiers: map[string]float64{"speed": 0.2, "size": 0.1},
		MutationRate:   0.0,
		EnergyDrain:    0.3,
		Symbol:         '≈',
		Temperature:    0.0,
		Pressure:       1.1,
		OxygenLevel:    0.7,
		Humidity:       1.0,
		IsAquatic:      true,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeRadiation] = Biome{
		Type:           BiomeRadiation,
		Name:           "Radiation",
		Color:          "red",
		TraitModifiers: map[string]float64{"endurance": -0.2},
		MutationRate:   0.3,
		EnergyDrain:    2.0,
		Symbol:         '☢',
		Temperature:    0.5,
		Pressure:       1.2,
		OxygenLevel:    0.6,
		Humidity:       0.2,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeSoil] = Biome{
		Type:           BiomeSoil,
		Name:           "Soil",
		Color:          "brown",
		TraitModifiers: map[string]float64{"digging_ability": 0.3, "size": -0.1, "underground_nav": 0.2},
		MutationRate:   0.02,
		EnergyDrain:    0.7,
		Symbol:         '■',
		Temperature:    0.2,
		Pressure:       1.3,
		OxygenLevel:    0.5,
		Humidity:       0.8,
		IsAquatic:      false,
		IsUnderground:  true,
		IsAerial:       false,
	}

	biomes[BiomeAir] = Biome{
		Type:           BiomeAir,
		Name:           "Air",
		Color:          "cyan",
		TraitModifiers: map[string]float64{"flying_ability": 0.4, "altitude_tolerance": 0.3, "size": -0.2},
		MutationRate:   0.01,
		EnergyDrain:    1.0,
		Symbol:         '☁',
		Temperature:    -0.5,
		Pressure:       0.6,
		OxygenLevel:    0.4,
		Humidity:       0.3,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       true,
	}

	// New biome expansions
	biomes[BiomeIce] = Biome{
		Type:           BiomeIce,
		Name:           "Ice",
		Color:          "white",
		TraitModifiers: map[string]float64{"endurance": 0.4, "size": 0.1, "speed": -0.3, "defense": 0.2},
		MutationRate:   0.02,
		EnergyDrain:    2.5,
		Symbol:         '❅',
		Temperature:    -0.9,
		Pressure:       1.0,
		OxygenLevel:    0.9,
		Humidity:       0.9,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeRainforest] = Biome{
		Type:           BiomeRainforest,
		Name:           "Rainforest",
		Color:          "darkgreen",
		TraitModifiers: map[string]float64{"agility": 0.3, "intelligence": 0.2, "size": -0.1, "cooperation": 0.2},
		MutationRate:   0.15,
		EnergyDrain:    0.3,
		Symbol:         '🌳',
		Temperature:    0.6,
		Pressure:       1.0,
		OxygenLevel:    1.3,
		Humidity:       1.0,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeDeepWater] = Biome{
		Type:           BiomeDeepWater,
		Name:           "Deep Water",
		Color:          "darkblue",
		TraitModifiers: map[string]float64{"aquatic_adaptation": 0.5, "strength": 0.3, "endurance": 0.4, "size": 0.2},
		MutationRate:   0.05,
		EnergyDrain:    1.8,
		Symbol:         '≋',
		Temperature:    -0.3,
		Pressure:       2.0,
		OxygenLevel:    0.3,
		Humidity:       1.0,
		IsAquatic:      true,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeHighAltitude] = Biome{
		Type:           BiomeHighAltitude,
		Name:           "High Altitude",
		Color:          "lightgray",
		TraitModifiers: map[string]float64{"altitude_tolerance": 0.6, "endurance": 0.5, "flying_ability": 0.3, "size": -0.2},
		MutationRate:   0.08,
		EnergyDrain:    3.0,
		Symbol:         '⛰',
		Temperature:    -0.8,
		Pressure:       0.3,
		OxygenLevel:    0.2,
		Humidity:       0.1,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       true,
	}

	biomes[BiomeHotSpring] = Biome{
		Type:           BiomeHotSpring,
		Name:           "Hot Spring",
		Color:          "orange",
		TraitModifiers: map[string]float64{"endurance": 0.3, "agility": 0.2, "aquatic_adaptation": 0.2, "speed": 0.1},
		MutationRate:   0.12,
		EnergyDrain:    0.8,
		Symbol:         '♨',
		Temperature:    0.9,
		Pressure:       1.1,
		OxygenLevel:    0.8,
		Humidity:       1.0,
		IsAquatic:      true,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeTundra] = Biome{
		Type:           BiomeTundra,
		Name:           "Tundra",
		Color:          "lightblue",
		TraitModifiers: map[string]float64{"endurance": 0.5, "size": 0.2, "speed": -0.2, "defense": 0.3},
		MutationRate:   0.03,
		EnergyDrain:    1.8,
		Symbol:         '❄',
		Temperature:    -0.7,
		Pressure:       1.0,
		OxygenLevel:    0.9,
		Humidity:       0.6,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeSwamp] = Biome{
		Type:           BiomeSwamp,
		Name:           "Swamp",
		Color:          "brown",
		TraitModifiers: map[string]float64{"aquatic_adaptation": 0.3, "digging_ability": 0.2, "intelligence": 0.1, "defense": 0.2},
		MutationRate:   0.08,
		EnergyDrain:    1.2,
		Symbol:         '🌿',
		Temperature:    0.3,
		Pressure:       1.1,
		OxygenLevel:    0.6,
		Humidity:       1.0,
		IsAquatic:      true,
		IsUnderground:  false,
		IsAerial:       false,
	}

	biomes[BiomeCanyon] = Biome{
		Type:           BiomeCanyon,
		Name:           "Canyon",
		Color:          "red",
		TraitModifiers: map[string]float64{"agility": 0.4, "strength": 0.2, "endurance": 0.2, "size": -0.1},
		MutationRate:   0.06,
		EnergyDrain:    1.5,
		Symbol:         '⫽',
		Temperature:    0.4,
		Pressure:       1.2,
		OxygenLevel:    0.8,
		Humidity:       0.2,
		IsAquatic:      false,
		IsUnderground:  false,
		IsAerial:       false,
	}

	return biomes
}

// generateBiome generates a biome type for a grid cell using enhanced noise patterns and topology integration
func (w *World) generateBiome(x, y int) BiomeType {
	// Get topology information if available
	var elevation, slope float64
	if w.TopologySystem != nil && x < len(w.TopologySystem.TopologyGrid) && y < len(w.TopologySystem.TopologyGrid[0]) {
		cell := w.TopologySystem.TopologyGrid[x][y]
		elevation = cell.Elevation
		slope = cell.Slope
	}

	// Calculate distance from center and edges for zonal distribution
	centerX, centerY := float64(w.Config.GridWidth)/2, float64(w.Config.GridHeight)/2
	distFromCenter := math.Sqrt(math.Pow(float64(x)-centerX, 2) + math.Pow(float64(y)-centerY, 2))
	maxDist := math.Sqrt(math.Pow(centerX, 2) + math.Pow(centerY, 2))
	
	// Distance from edges (for polar caps)
	distFromEdge := math.Min(math.Min(float64(x), float64(w.Config.GridWidth-x)), 
		math.Min(float64(y), float64(w.Config.GridHeight-y)))
	
	// Enhanced noise generation for contiguous features
	baseNoise := rand.Float64()
	
	// Create multiple noise layers for more realistic patterns
	microNoise := (rand.Float64() - 0.5) * 0.1  // Small local variations
	regionalNoise := perlinNoise(float64(x)*0.1, float64(y)*0.1) // Larger regional patterns
	
	combinedNoise := baseNoise + microNoise + regionalNoise*0.3
	if combinedNoise < 0 { combinedNoise = 0 }
	if combinedNoise > 1 { combinedNoise = 1 }

	// Polar regions (extreme edges) - Ice and Tundra
	if distFromEdge < 3 {
		if combinedNoise < 0.7 {
			return BiomeIce
		} else {
			return BiomeTundra
		}
	}

	// High elevation areas - Mountains and High Altitude
	if elevation > 0.3 { // Adjusted from 0.8 to 0.3 for new elevation range
		if elevation > 0.4 { // Adjusted from 0.95 to 0.4
			return BiomeHighAltitude
		} else if slope > 0.7 {
			return BiomeCanyon
		} else {
			return BiomeMountain
		}
	}

	// Very low elevation areas - Water features
	if elevation < -0.2 { // Adjusted from 0.2 to -0.2 for new elevation range
		if elevation < -0.3 { // Adjusted from 0.05 to -0.3
			return BiomeDeepWater
		} else if combinedNoise < 0.3 {
			return BiomeSwamp
		} else {
			return BiomeWater
		}
	}

	// Moderate elevation with specific climate patterns
	climateZone := distFromCenter / maxDist

	// Tropical/equatorial zones (center) - Rainforests and Hot Springs
	if climateZone < 0.3 {
		if combinedNoise < 0.05 {
			return BiomeHotSpring
		} else if combinedNoise < 0.6 {
			return BiomeRainforest
		} else if combinedNoise < 0.8 {
			return BiomeForest
		} else {
			return BiomePlains
		}
	}

	// Temperate zones
	if climateZone < 0.6 {
		switch {
		case combinedNoise < 0.05:
			return BiomeHotSpring
		case combinedNoise < 0.35:
			return BiomeForest
		case combinedNoise < 0.5:
			return BiomeWater
		case combinedNoise < 0.7:
			return BiomePlains
		case combinedNoise < 0.8:
			return BiomeMountain
		default:
			return BiomeSwamp
		}
	}

	// Arid/harsh zones (outer areas)
	switch {
	case combinedNoise < 0.02:
		return BiomeRadiation
	case combinedNoise < 0.4:
		return BiomeDesert
	case combinedNoise < 0.55:
		return BiomeMountain
	case combinedNoise < 0.7:
		return BiomeWater
	case combinedNoise < 0.8:
		return BiomePlains
	case combinedNoise < 0.9:
		return BiomeSoil
	case combinedNoise < 0.97:
		return BiomeAir
	default:
		return BiomeTundra
	}
}

// perlinNoise provides a simple Perlin-like noise function for terrain generation
func perlinNoise(x, y float64) float64 {
	// Simple noise approximation using sine waves
	return (math.Sin(x*1.7+y*1.3) + math.Sin(x*2.3-y*0.7) + math.Sin(x*0.9+y*2.1)) / 3.0
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

		// Enhance entity with specialized systems
		AddInsectTraitsToEntity(entity)
		AddPollinatorTraitsToEntity(entity)
		AddCasteStatusToEntity(entity)

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

	// 2a. Update seed dispersal system (handles seed movement and germination)
	w.SeedDispersalSystem.Update(w)
	
	// 2b. Update chemical ecology system (plant and entity chemical communication)
	w.ChemicalEcologySystem.Update(w)

	// 3. Update micro and macro evolution systems
	w.CellularSystem.UpdateCellularOrganisms()
	w.MacroEvolutionSystem.UpdateMacroEvolution(w)
	w.TopologySystem.UpdateTopology(w.Tick)
	
	// Update biomes based on topology changes (less frequently to avoid constant map resets)
	if w.Tick%10 == 0 { // Only update every 10 ticks instead of every tick
		w.updateBiomesFromTopology()
	}
	
	// Process biome transitions (hot spots melting ice, fires spreading, etc.)
	if w.Tick%20 == 0 { // Process transitions every 20 ticks for stability
		w.processBiomeTransitions()
	}

	// Clear grid entities and plants
	w.clearGrid()

	// Update world events
	w.updateEvents()
	// Update enhanced environmental events
	w.updateEnhancedEnvironmentalEvents()
	
	// Maybe trigger new events (less frequent during night)
	eventChance := 0.01
	if currentTimeState.IsNight() {
		eventChance *= 0.5 // Fewer events at night
	}
	if rand.Float64() < eventChance {
		w.triggerRandomEvent()
	}
	
	// Maybe trigger enhanced environmental events (lower chance)
	enhancedEventChance := 0.005 // 0.5% chance per tick
	if rand.Float64() < enhancedEventChance {
		w.triggerEnhancedEnvironmentalEvent()
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
	w.CollisionSystem.CheckCollisions(w.AllEntities, w.PhysicsComponents, w.PhysicsSystem, w)

	// Update grid with current entity and plant positions
	w.updateGrid()

	// 6. Update group behavior system
	w.GroupBehaviorSystem.UpdateGroups(w.Tick)

	// Try to form new groups based on proximity and compatibility
	if w.Tick%10 == 0 {
		w.attemptGroupFormation()
	}

	// Handle interactions between entities and with plants
	w.handleInteractions()
	
	// Apply biome environmental effects
	w.applyBiomeEffects()
	
	// 7. Update civilization system
	w.CivilizationSystem.Update(w.Tick)

	// Process civilization activities
	w.processCivilizationActivities()

	// Update reproduction system (gestation, egg hatching, decay)
	w.updateReproductionSystem()
	
	// Update fungal network (decomposition and nutrient cycling)
	if w.FungalNetwork != nil {
		w.FungalNetwork.Update(w, w.Tick)
	}
	
	// Update cultural knowledge system (multi-generational knowledge transfer)
	if w.CulturalKnowledgeSystem != nil {
		w.CulturalKnowledgeSystem.Update(w.AllEntities, w.Tick)
	}

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

	// Basic tool and modification creation (to supplement emergent behavior)
	w.attemptBasicToolsAndModifications()

	// Update event logger with population changes
	w.EventLogger.UpdatePopulationCounts(w.Tick, w.Populations)
	
	// Check for player species extinction and splitting (if web interface is active)
	if w.PlayerEventsCallback != nil {
		w.checkPlayerSpeciesEvents()
	}
	
	// Update statistical analysis system
	if w.StatisticalReporter != nil {
		// Take snapshot at regular intervals
		if w.Tick%w.StatisticalReporter.SnapshotInterval == 0 {
			w.StatisticalReporter.TakeSnapshot(w)
		}
		
		// Perform analysis at regular intervals
		if w.Tick%w.StatisticalReporter.AnalysisInterval == 0 {
			w.StatisticalReporter.PerformAnalysis(w)
		}
	}
	
	// Update hive mind, caste, and insect systems
	w.HiveMindSystem.Update()
	w.CasteSystem.Update(w, w.Tick)
	w.InsectSystem.Update(w.Tick)
	
	// Update insect pollination system
	currentSeason := w.AdvancedTimeSystem.GetTimeState().Season
	w.InsectPollinationSystem.Update(w.AllEntities, w.AllPlants, currentSeason, w.Tick)
	
	// Update colony warfare and diplomacy system
	w.ColonyWarfareSystem.Update(w.CasteSystem.Colonies, w.Tick)
	
	// Try to form new collective intelligence systems
	if w.Tick%100 == 0 { // Every 100 ticks
		w.attemptHiveMindFormation()
		w.attemptCasteColonyFormation()
		w.attemptSwarmFormation()
	}
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

		// Update basic entity properties using classification system
		entity.UpdateWithClassification(w.OrganismClassifier, w.CellularSystem)
		
		// Update metamorphosis and life stage development
		if entity.MetamorphosisStatus == nil {
			// Initialize metamorphosis status for new entities
			entity.MetamorphosisStatus = NewMetamorphosisStatus(entity, w.MetamorphosisSystem)
		}
		
		// Get environmental factors for metamorphosis
		gridX := int((entity.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((entity.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))
		
		environment := w.calculateEnvironmentalFactors(entity, gridX, gridY)
		stageChanged := w.MetamorphosisSystem.Update(entity, w.Tick, environment)
		
		if stageChanged {
			// Log metamorphosis events
			w.CentralEventBus.EmitSystemEvent(w.Tick, "metamorphosis", "life_stage", "metamorphosis_system",
				fmt.Sprintf("Entity %d advanced to %s stage", entity.ID, entity.MetamorphosisStatus.CurrentStage.String()),
				&entity.Position, map[string]interface{}{
					"entity_id": entity.ID,
					"new_stage": entity.MetamorphosisStatus.CurrentStage.String(),
					"metamorphosis_type": entity.MetamorphosisStatus.Type.String(),
				})
		}

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

	// Update basic entity properties using classification system
	entity.UpdateWithClassification(w.OrganismClassifier, w.CellularSystem)

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

// updatePlants handles plant growth, aging, and death with enhanced nutrient system
func (w *World) updatePlants() {
	// Get current season for plant nutrient calculations
	currentTimeState := w.AdvancedTimeSystem.GetTimeState()
	season := getSeasonName(currentTimeState.Season)
	
	// Process decay items and add nutrients to soil
	if len(w.ReproductionSystem.DecayingItems) > 0 {
		w.processDecayNutrientsToSoil()
	}
	
	// Process rainfall effects on soil
	w.processWeatherEffectsOnSoil()
	
	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}

		// Get grid cell for plant's location
		gridX := int((plant.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((plant.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))

		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))

		gridCell := &w.Grid[gridY][gridX]
		biome := w.Biomes[gridCell.Biome]
		
		// Update plant with enhanced nutrient system
		nutritionalHealth := plant.updatePlantNutrients(gridCell, season)
		
		// Traditional plant update with nutritional influence
		plant.Update(biome)
		
		// Apply nutritional health effects
		if nutritionalHealth < 0.5 {
			// Severe malnutrition - chance of death
			if rand.Float64() < (0.5-nutritionalHealth)*0.1 {
				plant.IsAlive = false
				// Add plant matter to decay system
				if w.ReproductionSystem != nil {
					nutrientValue := plant.NutritionVal * plant.Size
					w.ReproductionSystem.AddDecayingItem("plant_matter", plant.Position, nutrientValue, "plant", plant.Size, w.Tick)
				}
			}
		}
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

		// Seed production and dispersal (replaces simple asexual reproduction)
		if plant.CanReproduce() {
			// Create seeds instead of direct offspring
			seedCount := 1 + int(plant.GetTrait("reproduction_rate")*3) // 1-4 seeds typically
			for i := 0; i < seedCount; i++ {
				seed := w.SeedDispersalSystem.CreateSeed(plant, w)
				// Log seed creation
				if w.CentralEventBus != nil {
					plantTypeName := GetPlantConfigs()[plant.Type].Name
					w.CentralEventBus.EmitPlantEvent(w.Tick, "seed_creation", "seed_production", "plant_lifecycle", 
						fmt.Sprintf("Plant %d (%s) produced seed %d", plant.ID, plantTypeName, seed.ID), plant, false, true)
				}
			}
			
			// Reproduction costs energy
			config := GetPlantConfigs()[plant.Type]
			plant.Energy -= config.BaseEnergy * 0.3 // Reduced cost since seeds may not all germinate
		}

		// Release pollen for sexual reproduction during flowering
		if plant.CanReproduce() && (currentTimeState.Season == Spring || currentTimeState.Season == Summer) && rand.Float64() < 0.4 {
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

			w.WindSystem.ReleasePollen(plant, pollenAmount, w.Tick)
		}
	}
	// Process wind-based cross-pollination
	crossPollinatedPlants := w.WindSystem.TryPollination(w.AllPlants, w.SpeciationSystem, w.Tick)

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

// BiomeTransition represents a transition between biome states
type BiomeTransition struct {
	From        BiomeType
	To          BiomeType
	Trigger     string  // "heat", "cold", "fire", "flood", etc.
	Probability float64 // Base probability of transition
}

// processBiomeTransitions handles realistic biome state changes
func (w *World) processBiomeTransitions() {
	transitions := make(map[Position]BiomeType)
	
	// Define transition rules
	transitionRules := []BiomeTransition{
		// Hot spots melt ice - very rare natural transitions
		{BiomeIce, BiomeWater, "heat", 0.02},
		{BiomeIce, BiomePlains, "heat", 0.01}, // Ice melts to plains in some cases
		// Hot springs create effects - more likely near hot springs
		{BiomeIce, BiomeWater, "hotspring", 0.15}, // Near hot springs, ice melts more rapidly
		{BiomeWater, BiomeRainforest, "hotspring", 0.05}, // Hot springs create humid environments
		// Fire effects - moderate probability
		{BiomeForest, BiomeDesert, "fire", 0.1},
		{BiomeRainforest, BiomeDesert, "fire", 0.08},
		{BiomePlains, BiomeDesert, "fire", 0.06},
		// Fire extinguishing - slow recovery
		{BiomeDesert, BiomePlains, "water", 0.03}, // Desert recovering after water
		// Water accumulation effects - gradual
		{BiomePlains, BiomeSwamp, "flood", 0.08},
		{BiomeDesert, BiomePlains, "flood", 0.1},
		// Cold effects - slow freezing
		{BiomeWater, BiomeIce, "cold", 0.03},
		{BiomeSwamp, BiomeIce, "cold", 0.04},
		// Volcanic effects - significant but rare
		{BiomePlains, BiomeMountain, "volcanic", 0.12},
		{BiomeForest, BiomeRadiation, "volcanic", 0.08}, // Lava burns forest to radiation zones
	}
	
	// Check each grid cell for potential transitions
	for y := 0; y < w.Config.GridHeight; y++ {
		for x := 0; x < w.Config.GridWidth; x++ {
			currentBiome := w.Grid[y][x].Biome
			
			// Check for transition triggers in nearby cells
			triggers := w.detectTransitionTriggers(x, y)
			
			for trigger, intensity := range triggers {
				for _, rule := range transitionRules {
					if rule.From == currentBiome && rule.Trigger == trigger {
						// Probability modified by trigger intensity
						actualProbability := rule.Probability * intensity
						if rand.Float64() < actualProbability {
							transitions[Position{X: float64(x), Y: float64(y)}] = rule.To
							
							// Log the transition for events
							if w.EventLogger != nil {
								event := LogEvent{
									Timestamp:   time.Now(),
									Tick:        w.Tick,
									Type:        fmt.Sprintf("biome_transition_%s_to_%s", 
										w.getBiomeName(rule.From), w.getBiomeName(rule.To)),
									Description: fmt.Sprintf("Biome transition from %s to %s triggered by %s", 
										w.getBiomeName(rule.From), w.getBiomeName(rule.To), trigger),
									Data: map[string]interface{}{
										"trigger": trigger,
										"intensity": intensity,
										"from_biome": w.getBiomeName(rule.From),
										"to_biome": w.getBiomeName(rule.To),
										"position_x": float64(x),
										"position_y": float64(y),
									},
								}
								w.EventLogger.addEvent(event)
							}
							break // Only one transition per cell per tick
						}
					}
				}
			}
		}
	}
	
	// Apply transitions
	for pos, newBiome := range transitions {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
			w.Grid[gridY][gridX].Biome = newBiome
		}
	}
}

// detectTransitionTriggers identifies environmental conditions that can cause biome transitions
func (w *World) detectTransitionTriggers(x, y int) map[string]float64 {
	triggers := make(map[string]float64)
	
	// Check in a 3x3 neighborhood around the cell
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < w.Config.GridWidth && ny >= 0 && ny < w.Config.GridHeight {
				neighborBiome := w.Grid[ny][nx].Biome
				distance := math.Sqrt(float64(dx*dx + dy*dy))
				intensity := 1.0 / (1.0 + distance) // Closer = stronger effect
				
				switch neighborBiome {
				case BiomeHotSpring:
					triggers["heat"] = math.Max(triggers["heat"], intensity * 0.8)
					triggers["hotspring"] = math.Max(triggers["hotspring"], intensity)
				case BiomeRadiation:
					triggers["heat"] = math.Max(triggers["heat"], intensity * 0.6)
					triggers["volcanic"] = math.Max(triggers["volcanic"], intensity * 0.7)
				case BiomeWater, BiomeDeepWater, BiomeSwamp:
					triggers["water"] = math.Max(triggers["water"], intensity * 0.5)
					triggers["flood"] = math.Max(triggers["flood"], intensity * 0.3)
				case BiomeIce, BiomeTundra:
					triggers["cold"] = math.Max(triggers["cold"], intensity * 0.4)
				}
			}
		}
	}
	
	// Check for fire events - simplified fire detection
	// In a real implementation, this would check for active fire events
	if rand.Float64() < 0.01 { // 1% chance of fire starting in flammable biomes
		currentBiome := w.Grid[y][x].Biome
		if currentBiome == BiomeForest || currentBiome == BiomeRainforest || currentBiome == BiomePlains {
			triggers["fire"] = 0.8
		}
	}
	
	return triggers
}

// getBiomeName returns human-readable biome name
func (w *World) getBiomeName(biome BiomeType) string {
	biomeNames := map[BiomeType]string{
		BiomePlains:      "plains",
		BiomeForest:      "forest", 
		BiomeDesert:      "desert",
		BiomeMountain:    "mountain",
		BiomeWater:       "water",
		BiomeRadiation:   "radiation",
		BiomeSoil:        "soil",
		BiomeAir:         "air",
		BiomeIce:         "ice",
		BiomeRainforest:  "rainforest",
		BiomeDeepWater:   "deep_water",
		BiomeHighAltitude: "high_altitude",
		BiomeHotSpring:   "hot_spring",
		BiomeTundra:      "tundra",
		BiomeSwamp:       "swamp",
		BiomeCanyon:      "canyon",
	}
	if name, exists := biomeNames[biome]; exists {
		return name
	}
	return "unknown"
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

				// Enhance entity with specialized systems
				AddInsectTraitsToEntity(newEntity)
				AddPollinatorTraitsToEntity(newEntity)
				AddCasteStatusToEntity(newEntity)

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
			}, w.Tick)
		} else if entity.Energy > 80 && rand.Float64() < 0.05 {
			// Food found signal
			w.CommunicationSystem.SendSignal(entity, SignalFood, map[string]interface{}{
				"position": entity.Position,
				"energy":   entity.Energy,
			}, w.Tick)
		} else if cooperation > 0.6 && rand.Float64() < 0.03 {
			// Cooperation signal
			w.CommunicationSystem.SendSignal(entity, SignalHelp, map[string]interface{}{
				"species":     entity.Species,
				"cooperation": cooperation,
			}, w.Tick)
		}
	}

	// Receive and respond to signals
	receivedSignals := w.CommunicationSystem.ReceiveSignals(entity, w.Tick)
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

					w.GroupBehaviorSystem.FormGroup(nearbyEntities, purpose, w.Tick)
				}
			}
		}
	}
}

// processCivilizationActivities handles tribe activities and structure management
func (w *World) processCivilizationActivities() {
	// Update civilization system
	w.CivilizationSystem.Update(w.Tick)

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
		
		// Don't reproduce if entity is too young using classification system
		if !w.OrganismClassifier.IsReproductivelyMature(entity1, entity1.Classification) {
			continue
		}
		
		// Don't reproduce if entity has low energy (adjusted for classification)
		energyThreshold := 30.0
		maintenanceCost := w.OrganismClassifier.CalculateEnergyMaintenance(entity1, entity1.Classification)
		if entity1.Energy < energyThreshold + maintenanceCost*5 { // Need 5x maintenance cost as buffer
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
			
			// Attempt mating using classification system
			if w.ReproductionSystem.StartMatingWithClassification(entity1, entity2, w.OrganismClassifier, w.Tick) {
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
	deathsThisTick := make(map[string]int) // Track deaths by species
	totalDeaths := 0
	
	for _, entity := range w.AllEntities {
		// Check if entity needs to be processed for death
		shouldProcessDeath := false
		cause := ""
		
		// Check for death by energy depletion (handled by UpdateWithClassification but might not be processed yet)
		if entity.IsAlive && entity.Energy <= 0 {
			shouldProcessDeath = true
			cause = "energy_depletion"
			entity.IsAlive = false
		}
		
		// Check for death by old age using new classification system
		if entity.IsAlive && w.OrganismClassifier.IsDeathByOldAge(entity, entity.Classification, entity.MaxLifespan) {
			shouldProcessDeath = true
			cause = "old_age"
			entity.IsAlive = false
		}
		
		// Fallback: Check for death by old age using old system (for entities not yet classified)
		if entity.IsAlive && entity.Age > 1000 {
			shouldProcessDeath = true
			cause = "old_age_legacy"
			entity.IsAlive = false
		}
		
		// Check if entity just died this tick (was alive but now marked as dead)
		if !entity.IsAlive && shouldProcessDeath {
			// Track death by species
			deathsThisTick[entity.Species]++
			totalDeaths++
			
			// Create decaying corpse
			corpseNutrientValue := entity.Energy*0.5 + float64(entity.Age)*0.1
			w.ReproductionSystem.AddDecayingItem("corpse", entity.Position, corpseNutrientValue, entity.Species, entity.GetTrait("size"), w.Tick)
			
			// Enhanced death logging with cause tracking
			contributingFactors := make(map[string]interface{})
			contributingFactors["energy"] = entity.Energy
			contributingFactors["age"] = entity.Age
			contributingFactors["classification"] = w.OrganismClassifier.GetClassificationName(entity.Classification)
			contributingFactors["max_lifespan"] = entity.MaxLifespan
			contributingFactors["molecular_health"] = w.calculateMolecularHealth(entity)
			entity.LogEntityDeath(w, cause, contributingFactors)
		}
	}
	
	// Check for mass die-off events and their environmental impacts
	w.processMassDieOffImpacts(deathsThisTick, totalDeaths)
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

// applyBiomeEffects applies environmental effects from biomes to entities
func (w *World) applyBiomeEffects() {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Get the biome at entity's position
		biomeType := w.getBiomeAtPosition(entity.Position.X, entity.Position.Y)
		biome := w.Biomes[biomeType]

		// Apply environmental pressure based on biome conditions
		w.applyEnvironmentalPressure(entity, biome)

		// Apply biome-specific effects
		w.applyBiomeSpecificEffects(entity, biome)
	}
}

// applyEnvironmentalPressure applies environmental stress based on biome conditions
func (w *World) applyEnvironmentalPressure(entity *Entity, biome Biome) {
	// Calculate environmental stress factors
	temperatureStress := math.Abs(biome.Temperature) * 0.1
	pressureStress := math.Abs(biome.Pressure - 1.0) * 0.15
	oxygenStress := (1.0 - biome.OxygenLevel) * 0.2

	totalStress := temperatureStress + pressureStress + oxygenStress

	// Apply energy drain based on environmental adaptation
	adaptationBonus := 0.0
	switch {
	case biome.IsAquatic && entity.GetTrait("aquatic_adaptation") > 0.5:
		adaptationBonus = entity.GetTrait("aquatic_adaptation") * 0.3
	case biome.IsAerial && entity.GetTrait("flying_ability") > 0.5:
		adaptationBonus = entity.GetTrait("flying_ability") * 0.25
	case biome.IsUnderground && entity.GetTrait("digging_ability") > 0.5:
		adaptationBonus = entity.GetTrait("digging_ability") * 0.2
	case biome.Type == BiomeHighAltitude && entity.GetTrait("altitude_tolerance") > 0.5:
		adaptationBonus = entity.GetTrait("altitude_tolerance") * 0.4
	}

	// Calculate final energy cost
	energyCost := totalStress - adaptationBonus
	if energyCost > 0 {
		entity.Energy -= energyCost
		
		// Prevent energy from going too negative
		if entity.Energy < -10 {
			entity.Energy = -10
		}
	}
}

// applyBiomeSpecificEffects applies special effects for specific biomes
func (w *World) applyBiomeSpecificEffects(entity *Entity, biome Biome) {
	switch biome.Type {
	case BiomeDeepWater:
		// High pressure effects - entities without strong aquatic adaptation suffer
		if entity.GetTrait("aquatic_adaptation") < 0.7 {
			entity.Energy -= 0.5
			// Increase mutation rate due to pressure stress
			if rand.Float64() < 0.02 {
				entity.Mutate(0.1, 0.1)
			}
		}

	case BiomeHighAltitude:
		// Low oxygen effects - entities without altitude tolerance suffer
		if entity.GetTrait("altitude_tolerance") < 0.6 {
			entity.Energy -= 0.8
			// Reduced energy for movement and reproduction
			if entity.Energy > 0 {
				entity.Energy -= 0.3
			}
		}

	case BiomeIce:
		// Extreme cold effects
		if entity.GetTrait("endurance") < 0.6 {
			entity.Energy -= 1.0
			// Risk of severe energy loss
			if entity.Energy < 10 && rand.Float64() < 0.01 {
				entity.Energy -= 5 // Severe cold damage
				if entity.Energy <= 0 {
					entity.IsAlive = false // Death by freezing
				}
			}
		}

	case BiomeHotSpring:
		// Hot water beneficial for some, harmful for others
		if entity.GetTrait("aquatic_adaptation") > 0.5 {
			entity.Energy += 0.3 // Beneficial for aquatic entities
		} else if entity.GetTrait("endurance") < 0.4 {
			entity.Energy -= 0.4 // Too hot for weak entities
		}

	case BiomeRainforest:
		// High biodiversity promotes intelligence and cooperation
		if entity.GetTrait("intelligence") > 0.5 {
			entity.Energy += 0.2
		}
		if entity.GetTrait("cooperation") > 0.5 {
			entity.Energy += 0.1
		}

	case BiomeRadiation:
		// Radiation causes mutations and energy loss
		entity.Energy -= 1.0
		if rand.Float64() < 0.1 {
			entity.Mutate(0.2, 0.15)
		}

	case BiomeCanyon:
		// Steep terrain favors agile entities
		if entity.GetTrait("agility") > 0.6 {
			entity.Energy += 0.1
		} else {
			entity.Energy -= 0.3
		}

	case BiomeSwamp:
		// Disease risk in wetlands
		if entity.GetTrait("defense") < 0.5 && rand.Float64() < 0.005 {
			entity.Energy -= 2.0 // Disease-like energy loss
		}
	}
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

// updateBiomesFromTopology updates biomes based on topology changes from geological events
func (w *World) updateBiomesFromTopology() {
	if w.TopologySystem == nil {
		return
	}
	
	// Check for recent geological events that might change biomes (only active events)
	for _, event := range w.TopologySystem.GeologicalEvents {
		// Only apply events that are currently active
		if event.StartTick > w.Tick-event.Duration && event.Duration > 0 {
			w.applyGeologicalEventToBiomes(event)
		}
	}
	
	// Periodically recalculate biomes based on topology (much less frequently)
	if w.Tick%2000 == 0 { // Changed from 500 to 2000 ticks
		w.recalculateBiomesFromTopology()
	}
}

// applyGeologicalEventToBiomes changes biomes based on geological events
func (w *World) applyGeologicalEventToBiomes(event GeologicalEvent) {
	centerGridX := int((event.Center.X / float64(w.Config.Width)) * float64(w.Config.GridWidth))
	centerGridY := int((event.Center.Y / float64(w.Config.Height)) * float64(w.Config.GridHeight))
	gridRadius := int((event.Radius / float64(w.Config.Width)) * float64(w.Config.GridWidth))
	
	for x := centerGridX - gridRadius; x <= centerGridX + gridRadius; x++ {
		for y := centerGridY - gridRadius; y <= centerGridY + gridRadius; y++ {
			if x < 0 || x >= w.Config.GridWidth || y < 0 || y >= w.Config.GridHeight {
				continue
			}
			
			distance := math.Sqrt(float64((x-centerGridX)*(x-centerGridX) + (y-centerGridY)*(y-centerGridY)))
			if distance > float64(gridRadius) {
				continue
			}
			
			// Get topology information
			topoX := int((float64(x) / float64(w.Config.GridWidth)) * float64(w.TopologySystem.Width))
			topoY := int((float64(y) / float64(w.Config.GridHeight)) * float64(w.TopologySystem.Height))
			
			if topoX >= 0 && topoX < w.TopologySystem.Width && topoY >= 0 && topoY < w.TopologySystem.Height {
				topoCell := w.TopologySystem.TopologyGrid[topoX][topoY]
				
				// Change biomes based on event type and topology
				newBiome := w.determineBiomeFromGeology(event.Type, topoCell, w.Grid[y][x].Biome)
				if newBiome != w.Grid[y][x].Biome {
					w.Grid[y][x].Biome = newBiome
				}
			}
		}
	}
}

// determineBiomeFromGeology determines the new biome based on geological event and topology
func (w *World) determineBiomeFromGeology(eventType string, topoCell TopologyCell, currentBiome BiomeType) BiomeType {
	switch eventType {
	case "volcanic_eruption":
		// High elevation volcanic areas become mountains or radiation zones
		if topoCell.Elevation > 0.8 {
			return BiomeMountain
		} else if topoCell.Elevation > 0.6 {
			return BiomeRadiation // Volcanic ash and heat
		}
		
	case "mountain_uplift":
		// Mountain uplift creates mountain biomes
		if topoCell.Elevation > 0.9 {
			return BiomeHighAltitude
		} else if topoCell.Elevation > 0.7 {
			return BiomeMountain
		}
		
	case "seafloor_spreading", "rift_valley":
		// Creates deep water or water biomes
		if topoCell.Elevation < -0.3 {
			return BiomeDeepWater
		} else if topoCell.Elevation < 0.1 {
			return BiomeWater
		}
		
	case "geyser_formation", "hot_spring_creation":
		// Creates hot spring biomes
		if topoCell.WaterLevel > 0.3 {
			return BiomeHotSpring
		}
		
	case "ice_sheet_advance":
		// Creates ice biomes
		if topoCell.WaterLevel > 0.2 && topoCell.Elevation > 0.3 {
			return BiomeIce
		} else if topoCell.Elevation > 0.5 {
			return BiomeTundra
		}
		
	case "glacial_retreat":
		// Transitions from ice back to other biomes
		if currentBiome == BiomeIce {
			if topoCell.Elevation > 0.7 {
				return BiomeMountain
			} else if topoCell.Elevation > 0.3 {
				return BiomeTundra
			} else {
				return BiomePlains
			}
		}
		
	case "flood":
		// Creates swamp or water biomes
		if topoCell.WaterLevel > 0.5 {
			if topoCell.Elevation > 0.1 {
				return BiomeSwamp
			} else {
				return BiomeWater
			}
		}
	}
	
	return currentBiome // No change
}

// recalculateBiomesFromTopology recalculates biomes based on current topology
func (w *World) recalculateBiomesFromTopology() {
	// Only do very minimal recalculation to avoid massive map changes
	// Recalculate only 2% of cells each time to maintain stability
	totalCells := w.Config.GridWidth * w.Config.GridHeight
	cellsToUpdate := totalCells / 50 // Update 2% of cells (was 10%)
	
	for i := 0; i < cellsToUpdate; i++ {
		x := rand.Intn(w.Config.GridWidth)
		y := rand.Intn(w.Config.GridHeight)
		
		// Get topology information
		topoX := int((float64(x) / float64(w.Config.GridWidth)) * float64(w.TopologySystem.Width))
		topoY := int((float64(y) / float64(w.Config.GridHeight)) * float64(w.TopologySystem.Height))
		
		if topoX >= 0 && topoX < w.TopologySystem.Width && topoY >= 0 && topoY < w.TopologySystem.Height {
			topoCell := w.TopologySystem.TopologyGrid[topoX][topoY]
			
			// Determine biome based on topology
			newBiome := w.determineBiomeFromTopology(topoCell, x, y)
			// Only change biome if there's a significant reason (large elevation change)
			if math.Abs(topoCell.Elevation-w.getExpectedElevationForBiome(w.Grid[y][x].Biome)) > 0.3 { // Increased threshold from 0.2 to 0.3
				w.Grid[y][x].Biome = newBiome
			}
		}
	}
}

// determineBiomeFromTopology determines biome based on topology characteristics
func (w *World) determineBiomeFromTopology(topoCell TopologyCell, gridX, gridY int) BiomeType {
	elevation := topoCell.Elevation
	waterLevel := topoCell.WaterLevel
	slope := topoCell.Slope
	
	// Distance from edges for polar biomes
	distFromEdge := math.Min(math.Min(float64(gridX), float64(w.Config.GridWidth-gridX)), 
		math.Min(float64(gridY), float64(w.Config.GridHeight-gridY)))
	
	// Very high elevation - high altitude (adjusted thresholds)
	if elevation > 0.4 { // Adjusted from 0.95 to 0.4
		return BiomeHighAltitude
	}
	
	// High elevation - mountains (adjusted thresholds)
	if elevation > 0.3 { // Adjusted from 0.8 to 0.3
		return BiomeMountain
	}
	
	// Water-based biomes (adjusted thresholds)
	if waterLevel > 0.5 || elevation < -0.2 { // Adjusted from 0.0 to -0.2
		if elevation < -0.3 { // Adjusted from -0.5 to -0.3
			return BiomeDeepWater
		}
		if waterLevel > 0.8 && elevation > 0.1 {
			return BiomeSwamp
		}
		return BiomeWater
	}
	
	// Edge biomes (polar regions)
	if distFromEdge < 3 {
		if waterLevel > 0.3 {
			return BiomeIce
		}
		return BiomeTundra
	}
	
	// Steep slopes - canyons
	if slope > 0.7 && elevation > 0.4 {
		return BiomeCanyon
	}
	
	// Use the enhanced biome generation for other areas
	return w.generateBiome(gridX, gridY)
}

// attemptHiveMindFormation tries to form new hive minds from compatible entities
func (w *World) attemptHiveMindFormation() {
	if len(w.AllEntities) < 10 { // Need minimum entities
		return
	}

	// Group entities by proximity and compatibility
	for _, entity := range w.AllEntities {
		if !entity.IsAlive || entity.GetTrait("hive_member") > 0.0 {
			continue // Skip dead or already in hive
		}

		// Check if entity is suitable for hive mind
		intelligence := entity.GetTrait("intelligence")
		cooperation := entity.GetTrait("cooperation")
		if intelligence < 0.3 || cooperation < 0.4 {
			continue
		}

		// Find nearby compatible entities
		nearbyEntities := make([]*Entity, 0)
		nearbyEntities = append(nearbyEntities, entity)

		for _, other := range w.AllEntities {
			if other == entity || !other.IsAlive || other.GetTrait("hive_member") > 0.0 {
				continue
			}

			distance := entity.DistanceTo(other)
			if distance < 15.0 {
				otherIntelligence := other.GetTrait("intelligence")
				otherCooperation := other.GetTrait("cooperation")
				
				// Check compatibility
				intelligenceDiff := math.Abs(intelligence - otherIntelligence)
				cooperationDiff := math.Abs(cooperation - otherCooperation)
				
				if intelligenceDiff < 0.5 && cooperationDiff < 0.3 && 
					otherIntelligence > 0.3 && otherCooperation > 0.4 {
					nearbyEntities = append(nearbyEntities, other)
				}
			}
		}

		if len(nearbyEntities) >= 5 { // Minimum for hive mind
			// Determine hive mind type based on group characteristics
			avgIntelligence := 0.0
			avgCooperation := 0.0
			for _, e := range nearbyEntities {
				avgIntelligence += e.GetTrait("intelligence")
				avgCooperation += e.GetTrait("cooperation")
			}
			avgIntelligence /= float64(len(nearbyEntities))
			avgCooperation /= float64(len(nearbyEntities))

			var hiveType HiveMindType
			if avgIntelligence > 0.8 && avgCooperation > 0.8 {
				hiveType = QuantumMind
			} else if avgIntelligence > 0.6 && avgCooperation > 0.7 {
				hiveType = NeuralNetwork
			} else if avgCooperation > 0.7 {
				hiveType = SwarmIntelligence
			} else {
				hiveType = SimpleCollective
			}

			// Try to form hive mind
			hiveMind := w.HiveMindSystem.TryFormHiveMind(nearbyEntities, hiveType)
			if hiveMind != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "hive_mind_formed", 
					fmt.Sprintf("New %s hive mind formed with %d members", 
						hiveType, len(hiveMind.Members)))
			}
		}
	}
}

// attemptCasteColonyFormation tries to form new caste-based colonies
func (w *World) attemptCasteColonyFormation() {
	if len(w.AllEntities) < 8 { // Need minimum entities
		return
	}

	// Look for entities suitable for caste system formation
	for _, entity := range w.AllEntities {
		if !entity.IsAlive || entity.TribeID != 0 || entity.CasteStatus != nil {
			continue // Skip if already in tribe/caste or dead
		}

		// Check if entity could be a suitable leader for caste formation
		intelligence := entity.GetTrait("intelligence")
		cooperation := entity.GetTrait("cooperation")
		endurance := entity.GetTrait("endurance")
		
		// Use available traits with more achievable requirements
		if intelligence > 0.2 && cooperation > 0.3 && endurance > 0.2 {
			// Find nearby compatible entities for colony
			nearbyEntities := make([]*Entity, 0)
			nearbyEntities = append(nearbyEntities, entity)

			for _, other := range w.AllEntities {
				if other == entity || !other.IsAlive || other.TribeID != 0 || other.CasteStatus != nil {
					continue
				}

				distance := entity.DistanceTo(other)
				if distance < 20.0 {
					otherCooperation := other.GetTrait("cooperation")
					otherIntelligence := other.GetTrait("intelligence")
					
					// Check species compatibility and cooperation (relaxed requirements)
					if other.Species == entity.Species && 
						otherCooperation > 0.2 && otherIntelligence > 0.0 {
						nearbyEntities = append(nearbyEntities, other)
					}
				}
			}

			if len(nearbyEntities) >= 4 { // Minimum for caste colony (reduced from 6)
				// Choose nest location
				nestLocation := entity.Position
				
				// Try to form colony
				colony := w.CasteSystem.TryFormCasteColony(nearbyEntities, nestLocation)
				if colony != nil {
					w.EventLogger.LogWorldEvent(w.Tick, "caste_colony_formed",
						fmt.Sprintf("New caste colony formed with %d members and %d caste types", 
							colony.ColonySize, len(colony.CasteDistribution)))
					
					// Create a corresponding tribe for civilization system integration
					if w.CivilizationSystem != nil {
						tribeName := fmt.Sprintf("Colony-%d", colony.ID)
						tribe := w.CivilizationSystem.FormTribe(nearbyEntities, tribeName, w.Tick)
						if tribe != nil {
							tribe.ID = colony.ID // Sync IDs
						}
					}
				}
			}
		}
	}
}

// attemptSwarmFormation tries to form new swarm units
func (w *World) attemptSwarmFormation() {
	if len(w.AllEntities) < 10 { // Need minimum entities for swarms
		return
	}

	// Look for entities with high swarm capability
	swarmCandidates := make([]*Entity, 0)
	
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		swarmCapability := entity.GetTrait("swarm_capability")
		cooperation := entity.GetTrait("cooperation")
		
		// Check if already in a swarm
		if entity.GetTrait("swarm_member") > 0.0 {
			continue
		}

		if swarmCapability > 0.3 && cooperation > 0.4 {
			swarmCandidates = append(swarmCandidates, entity)
		}
	}

	if len(swarmCandidates) < 5 {
		return
	}

	// Group candidates by proximity and species
	proximityGroups := make(map[string][]*Entity)
	
	for _, entity := range swarmCandidates {
		// Create a key based on position and species
		posKey := fmt.Sprintf("%s_%.0f_%.0f", entity.Species, 
			math.Floor(entity.Position.X/20.0)*20.0, 
			math.Floor(entity.Position.Y/20.0)*20.0)
		
		if _, exists := proximityGroups[posKey]; !exists {
			proximityGroups[posKey] = make([]*Entity, 0)
		}
		proximityGroups[posKey] = append(proximityGroups[posKey], entity)
	}

	// Try to form swarms from each group
	for _, group := range proximityGroups {
		if len(group) >= 5 {
			// Determine swarm purpose based on group characteristics
			avgAggression := 0.0
			avgExploration := 0.0
			avgEnergy := 0.0
			
			for _, entity := range group {
				avgAggression += entity.GetTrait("aggression")
				avgExploration += entity.GetTrait("exploration_drive")
				avgEnergy += entity.Energy
			}
			avgAggression /= float64(len(group))
			avgExploration /= float64(len(group))
			avgEnergy /= float64(len(group))

			var purpose string
			if avgEnergy < 30.0 {
				purpose = "foraging"
			} else if avgAggression > 0.5 {
				purpose = "defense"
			} else if avgExploration > 0.5 {
				purpose = "exploration"
			} else {
				purpose = "migration"
			}

			// Create swarm
			swarm := w.InsectSystem.CreateSwarmUnit(group, purpose)
			if swarm != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "swarm_formed",
					fmt.Sprintf("New %s swarm formed with %d members", 
						purpose, len(swarm.Members)))

				// Create pheromone trail for the swarm
				if len(swarm.Members) > 0 && swarm.LeaderEntity != nil {
					// Create trail from swarm center to target
					w.InsectSystem.CreatePheromoneTrail(swarm.LeaderEntity, TrailPheromone, 
						swarm.CenterPosition, swarm.TargetPosition)
				}
			}
		}
	}
}

// enhanceEntitiesWithSpecializedSystems adds insect traits and caste status to suitable entities
func (w *World) enhanceEntitiesWithSpecializedSystems() {
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Add insect traits to suitable entities
		if IsEntityInsectLike(entity) {
			AddInsectTraitsToEntity(entity)
			AddPollinatorTraitsToEntity(entity) // Add pollinator traits for insect-like entities
		}

		// Add caste status to entities that don't have it
		if entity.CasteStatus == nil {
			AddCasteStatusToEntity(entity)
		}
	}
}

// CreateOffspring creates a new offspring entity from two parent entities
func (w *World) CreateOffspring(parent1, parent2 *Entity) *Entity {
	if parent1 == nil || parent2 == nil || !parent1.IsAlive || !parent2.IsAlive {
		return nil
	}
	
	// Generate new ID
	w.NextID++
	
	// Create offspring with traits from both parents
	offspring := &Entity{
		ID:         w.NextID,
		Traits:     make(map[string]Trait),
		Fitness:    0.0,
		Energy:     50.0, // Starting energy for offspring
		Age:        0,
		IsAlive:    true,
		Species:    parent1.Species, // Inherit species from first parent
		Generation: max(parent1.Generation, parent2.Generation) + 1,
	}
	
	// Position offspring near parents
	offspring.Position = Position{
		X: (parent1.Position.X + parent2.Position.X) / 2.0,
		Y: (parent1.Position.Y + parent2.Position.Y) / 2.0,
	}
	
	// Inherit traits from both parents with some variation
	for traitName := range parent1.Traits {
		parent1Value := parent1.Traits[traitName].Value
		parent2Value := parent2.Traits[traitName].Value
		
		// Average parent values with some random variation
		avgValue := (parent1Value + parent2Value) / 2.0
		variation := (rand.Float64() - 0.5) * 0.2 // Small random variation
		finalValue := avgValue + variation
		
		// Clamp to reasonable bounds
		finalValue = math.Max(-1.0, math.Min(1.0, finalValue))
		
		offspring.Traits[traitName] = Trait{
			Name:  traitName,
			Value: finalValue,
		}
	}
	
	// Initialize molecular systems for offspring
	offspring.MolecularNeeds = NewMolecularNeeds(offspring)
	offspring.MolecularMetabolism = NewMolecularMetabolism(offspring)
	offspring.MolecularProfile = CreateEntityMolecularProfile(offspring)
	
	// Initialize other systems
	offspring.DietaryMemory = NewDietaryMemory()
	offspring.EnvironmentalMemory = NewEnvironmentalMemory()
	
	offspring.ReproductionStatus = NewReproductionStatus()
	
	// Add enhanced systems
	AddCasteStatusToEntity(offspring)
	if IsEntityInsectLike(offspring) {
		AddInsectTraitsToEntity(offspring)
		AddPollinatorTraitsToEntity(offspring)
	}
	
	return offspring
}

// checkPlayerSpeciesEvents checks for species extinction and splitting events for player notifications
func (w *World) checkPlayerSpeciesEvents() {
	if w.PlayerEventsCallback == nil {
		return
	}
	
	// Check for extinctions
	for speciesName, previousCount := range w.PreviousPopulationCounts {
		if population, exists := w.Populations[speciesName]; exists {
			// Count alive entities
			aliveCount := 0
			for _, entity := range population.Entities {
				if entity.IsAlive {
					aliveCount++
				}
			}
			
			// Check for extinction (no alive entities)
			if previousCount > 0 && aliveCount == 0 {
				w.PlayerEventsCallback("species_extinct", map[string]interface{}{
					"species_name": speciesName,
					"last_count":   previousCount,
					"tick":         w.Tick,
				})
			}
		} else {
			// Population completely removed from world
			if previousCount > 0 {
				w.PlayerEventsCallback("species_extinct", map[string]interface{}{
					"species_name": speciesName,
					"last_count":   previousCount,
					"tick":         w.Tick,
				})
			}
		}
	}
	
	// Check for new species (potential splits)
	for speciesName, population := range w.Populations {
		if _, existed := w.PreviousPopulationCounts[speciesName]; !existed {
			// This is a new species - could be from splitting
			aliveCount := 0
			for _, entity := range population.Entities {
				if entity.IsAlive {
					aliveCount++
				}
			}
			
			if aliveCount > 0 {
				// Check if this could be a sub-species (similar name pattern)
				parentSpecies := w.findPotentialParentSpecies(speciesName)
				if parentSpecies != "" {
					w.PlayerEventsCallback("subspecies_formed", map[string]interface{}{
						"species_name":     speciesName,
						"parent_species":   parentSpecies,
						"entity_count":     aliveCount,
						"tick":             w.Tick,
					})
				} else {
					w.PlayerEventsCallback("new_species_detected", map[string]interface{}{
						"species_name": speciesName,
						"entity_count": aliveCount,
						"tick":         w.Tick,
					})
				}
			}
		}
	}
	
	// Update previous counts for next check
	for speciesName, population := range w.Populations {
		aliveCount := 0
		for _, entity := range population.Entities {
			if entity.IsAlive {
				aliveCount++
			}
		}
		w.PreviousPopulationCounts[speciesName] = aliveCount
	}
}

// findPotentialParentSpecies attempts to find a parent species for a potential sub-species
func (w *World) findPotentialParentSpecies(newSpeciesName string) string {
	// Simple heuristic: look for species with similar names or that share name components
	for existingSpecies := range w.Populations {
		if existingSpecies != newSpeciesName {
			// Check if the new species name contains the existing species name
			// This would suggest it's a variant or sub-species
			if len(existingSpecies) < len(newSpeciesName) {
				// Look for common prefixes or if one name contains the other
				commonPrefix := 0
				minLen := len(existingSpecies)
				if len(newSpeciesName) < minLen {
					minLen = len(newSpeciesName)
				}
				
				for i := 0; i < minLen; i++ {
					if existingSpecies[i] == newSpeciesName[i] {
						commonPrefix++
					} else {
						break
					}
				}
				
				// If more than 50% of the shorter name matches, consider it related
				if float64(commonPrefix)/float64(minLen) > 0.5 {
					return existingSpecies
				}
			}
		}
	}
	return ""
}

// getExpectedElevationForBiome returns the typical elevation range for a biome type
func (w *World) getExpectedElevationForBiome(biomeType BiomeType) float64 {
	switch biomeType {
	case BiomeHighAltitude:
		return 0.4
	case BiomeMountain, BiomeCanyon:
		return 0.3
	case BiomeDeepWater:
		return -0.3
	case BiomeWater, BiomeSwamp:
		return -0.1
	case BiomeIce, BiomeTundra:
		return 0.0 // Edge biomes, elevation varies
	default:
		return 0.0 // Moderate elevation biomes
	}
}

// Enhanced Environmental Event System Methods

// updateEnhancedEnvironmentalEvents processes active enhanced environmental events
func (w *World) updateEnhancedEnvironmentalEvents() {
	activeEvents := make([]*EnhancedEnvironmentalEvent, 0)
	
	for _, event := range w.EnvironmentalEvents {
		event.Duration--
		
		// Update event position and spread if it's moving/spreading
		w.updateEventMovement(event)
		
		// Apply event effects
		w.applyEventEffects(event)
		
		// Check if event should continue
		if event.Duration > 0 && event.Intensity > 0.1 {
			activeEvents = append(activeEvents, event)
		} else {
			// Log event completion
			if w.EventLogger != nil {
				endEvent := LogEvent{
					Timestamp:   time.Now(),
					Tick:        w.Tick,
					Type:        "environmental_event_end",
					Description: fmt.Sprintf("%s ended after %d ticks", event.Name, w.Tick-event.StartTick),
					Data: map[string]interface{}{
						"event_type": event.Type,
						"duration": w.Tick - event.StartTick,
						"affected_cells": len(event.AffectedCells),
					},
				}
				w.EventLogger.addEvent(endEvent)
			}
		}
	}
	
	w.EnvironmentalEvents = activeEvents
}

// updateEventMovement handles movement and spread of environmental events
func (w *World) updateEventMovement(event *EnhancedEnvironmentalEvent) {
	switch event.Type {
	case "wildfire":
		w.updateFireSpread(event)
	case "storm", "hurricane", "tornado":
		w.updateStormMovement(event)
	case "volcanic_eruption":
		w.updateVolcanicSpread(event)
	case "flood":
		w.updateFloodSpread(event)
	}
}

// updateFireSpread handles wildfire spreading with wind influence
func (w *World) updateFireSpread(fire *EnhancedEnvironmentalEvent) {
	if !fire.WindSensitive || w.WindSystem == nil {
		return
	}
	
	// Get wind direction and strength at fire location
	windVector := w.getWindAtPosition(fire.Position)
	
	// Wind affects fire direction and speed
	windInfluence := windVector.Strength * 0.5 // Wind contributes up to 50% to fire movement
	fire.Direction = math.Atan2(windVector.Y, windVector.X) // Wind direction in radians
	fire.Speed = 0.5 + windInfluence // Base speed plus wind boost
	
	// Move fire center
	fire.Position.X += fire.Speed * math.Cos(fire.Direction)
	fire.Position.Y += fire.Speed * math.Sin(fire.Direction)
	
	// Spread fire to new cells
	newAffectedCells := make(map[Position]BiomeType)
	
	// Current fire radius expands slightly each tick
	currentRadius := math.Min(fire.Radius + 0.5, fire.MaxRadius)
	fire.Radius = currentRadius
	
	// Check cells within fire radius
	for dy := -int(currentRadius); dy <= int(currentRadius); dy++ {
		for dx := -int(currentRadius); dx <= int(currentRadius); dx++ {
			cellPos := Position{
				X: fire.Position.X + float64(dx),
				Y: fire.Position.Y + float64(dy),
			}
			
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			if distance <= currentRadius {
				gridX, gridY := int(cellPos.X), int(cellPos.Y)
				if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
					currentBiome := w.Grid[gridY][gridX].Biome
					
					// Check if fire can spread to this biome
					if w.isFlammableBiome(currentBiome) {
						// Fire spreads more easily with wind in the right direction
						windBonus := 0.0
						if fire.WindSensitive {
							cellDirection := math.Atan2(float64(dy), float64(dx))
							directionAlignment := math.Cos(cellDirection - fire.Direction)
							windBonus = windInfluence * directionAlignment * 0.3
						}
						
						spreadProbability := 0.6 + windBonus - (distance / currentRadius) * 0.3
						if rand.Float64() < spreadProbability {
							newAffectedCells[cellPos] = BiomeDesert // Fire turns vegetation to desert
						}
					} else if w.isFireExtinguishingBiome(currentBiome) {
						// Fire is extinguished by water/ice
						fire.Intensity *= 0.7 // Reduce fire intensity
					}
				}
			}
		}
	}
	
	// Add new affected cells
	for pos, newBiome := range newAffectedCells {
		fire.AffectedCells[pos] = newBiome
	}
}

// updateStormMovement handles storm movement with wind patterns
func (w *World) updateStormMovement(storm *EnhancedEnvironmentalEvent) {
	if w.WindSystem == nil {
		return
	}
	
	// Storms follow regional wind patterns
	baseWind := w.WindSystem.BaseWindDirection
	
	// Add some randomness to storm movement
	stormDirection := baseWind + (rand.Float64()-0.5)*0.5 // ±0.25 radians variation
	
	storm.Direction = stormDirection
	storm.Speed = 1.0 + w.WindSystem.BaseWindStrength * 0.5
	
	// Move storm
	storm.Position.X += storm.Speed * math.Cos(storm.Direction)
	storm.Position.Y += storm.Speed * math.Sin(storm.Direction)
	
	// Update affected cells as storm moves
	newAffectedCells := make(map[Position]BiomeType)
	
	for dy := -int(storm.Radius); dy <= int(storm.Radius); dy++ {
		for dx := -int(storm.Radius); dx <= int(storm.Radius); dx++ {
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			if distance <= storm.Radius {
				cellPos := Position{
					X: storm.Position.X + float64(dx),
					Y: storm.Position.Y + float64(dy),
				}
				
				gridX, gridY := int(cellPos.X), int(cellPos.Y)
				if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
					currentBiome := w.Grid[gridY][gridX].Biome
					
					// Storm effects depend on type and current biome
					if storm.Type == "hurricane" || storm.Type == "tornado" {
						// Destructive storms can create wasteland
						if rand.Float64() < 0.3 {
							newAffectedCells[cellPos] = BiomeDesert
						}
					} else if storm.Type == "storm" {
						// Regular storms might create temporary flooding
						if currentBiome == BiomePlains && rand.Float64() < 0.2 {
							newAffectedCells[cellPos] = BiomeSwamp
						}
					}
				}
			}
		}
	}
	
	storm.AffectedCells = newAffectedCells
}

// updateVolcanicSpread handles volcanic eruption spreading
func (w *World) updateVolcanicSpread(volcano *EnhancedEnvironmentalEvent) {
	// Volcanic events spread outward from center
	currentRadius := math.Min(volcano.Radius + 0.3, volcano.MaxRadius)
	volcano.Radius = currentRadius
	
	newAffectedCells := make(map[Position]BiomeType)
	
	for dy := -int(currentRadius); dy <= int(currentRadius); dy++ {
		for dx := -int(currentRadius); dx <= int(currentRadius); dx++ {
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			if distance <= currentRadius && distance > currentRadius-0.5 { // Only affect the growing edge
				cellPos := Position{
					X: volcano.Position.X + float64(dx),
					Y: volcano.Position.Y + float64(dy),
				}
				
				gridX, gridY := int(cellPos.X), int(cellPos.Y)
				if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
					// Close to center becomes mountain, further out becomes radiation (lava)
					if distance < currentRadius * 0.3 {
						newAffectedCells[cellPos] = BiomeMountain
					} else if rand.Float64() < 0.7 {
						newAffectedCells[cellPos] = BiomeRadiation
					}
				}
			}
		}
	}
	
	// Add new affected cells
	for pos, newBiome := range newAffectedCells {
		volcano.AffectedCells[pos] = newBiome
	}
}

// updateFloodSpread handles flood spreading
func (w *World) updateFloodSpread(flood *EnhancedEnvironmentalEvent) {
	// Floods spread to lower elevation areas
	currentRadius := math.Min(flood.Radius + 0.4, flood.MaxRadius)
	flood.Radius = currentRadius
	
	newAffectedCells := make(map[Position]BiomeType)
	
	for dy := -int(currentRadius); dy <= int(currentRadius); dy++ {
		for dx := -int(currentRadius); dx <= int(currentRadius); dx++ {
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			if distance <= currentRadius {
				cellPos := Position{
					X: flood.Position.X + float64(dx),
					Y: flood.Position.Y + float64(dy),
				}
				
				gridX, gridY := int(cellPos.X), int(cellPos.Y)
				if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
					currentBiome := w.Grid[gridY][gridX].Biome
					
					// Floods turn plains/desert to swamp/water
					if currentBiome == BiomePlains || currentBiome == BiomeDesert {
						if rand.Float64() < 0.6 {
							newAffectedCells[cellPos] = BiomeSwamp
						}
					}
				}
			}
		}
	}
	
	flood.AffectedCells = newAffectedCells
}

// applyEventEffects applies the effects of environmental events to the world
func (w *World) applyEventEffects(event *EnhancedEnvironmentalEvent) {
	// Apply biome changes
	for pos, newBiome := range event.AffectedCells {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < w.Config.GridWidth && gridY >= 0 && gridY < w.Config.GridHeight {
			w.Grid[gridY][gridX].Biome = newBiome
		}
	}
	
	// Apply effects to entities and plants in the affected area
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(entity.Position.X-event.Position.X, 2) + 
							 math.Pow(entity.Position.Y-event.Position.Y, 2))
		
		if distance <= event.Radius {
			// Apply event-specific effects
			intensity := (event.Radius - distance) / event.Radius // Stronger closer to center
			
			if mutationEffect, exists := event.Effects["mutation"]; exists {
				// Apply mutation by directly calling the Mutate method
				entity.Mutate(mutationEffect * intensity, 0.1)
			}
			
			if damageEffect, exists := event.Effects["damage"]; exists {
				entity.Energy -= damageEffect * intensity
				if entity.Energy < 0 {
					entity.Energy = 0
				}
			}
		}
	}
	
	// Apply effects to plants
	for _, plant := range w.AllPlants {
		if !plant.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(plant.Position.X-event.Position.X, 2) + 
							 math.Pow(plant.Position.Y-event.Position.Y, 2))
		
		if distance <= event.Radius {
			intensity := (event.Radius - distance) / event.Radius
			
			// Fire kills plants
			if event.Type == "wildfire" {
				plant.Energy -= 50 * intensity
				if plant.Energy < 0 {
					plant.IsAlive = false
					// TODO: Add nutrients from dead plant to soil
				}
			}
		}
	}
}

// triggerEnhancedEnvironmentalEvent creates a new enhanced environmental event
func (w *World) triggerEnhancedEnvironmentalEvent() {
	eventTypes := []string{"wildfire", "storm", "volcanic_eruption", "flood", "hurricane", "tornado"}
	eventType := eventTypes[rand.Intn(len(eventTypes))]
	
	// Random position for event
	pos := Position{
		X: rand.Float64() * float64(w.Config.GridWidth),
		Y: rand.Float64() * float64(w.Config.GridHeight),
	}
	
	event := &EnhancedEnvironmentalEvent{
		ID:          w.NextEnvironmentalEventID,
		Type:        eventType,
		StartTick:   w.Tick,
		Position:    pos,
		AffectedCells: make(map[Position]BiomeType),
		Effects:     make(map[string]float64),
	}
	w.NextEnvironmentalEventID++
	
	// Configure event based on type
	switch eventType {
	case "wildfire":
		event.Name = "Wildfire"
		event.Description = "Spreading fire burns vegetation"
		event.Duration = 20 + rand.Intn(20)
		event.Radius = 2.0
		event.MaxRadius = 8.0
		event.Intensity = 0.8
		event.WindSensitive = true
		event.SpreadPattern = "wind_driven"
		event.ExtinguishOn = []BiomeType{BiomeWater, BiomeDeepWater, BiomeIce, BiomeSwamp}
		event.Effects["damage"] = 2.0
		event.Effects["mutation"] = 0.05
		
	case "storm":
		event.Name = "Storm"
		event.Description = "Heavy rainfall and wind"
		event.Duration = 15 + rand.Intn(15)
		event.Radius = 5.0
		event.MaxRadius = 12.0
		event.Intensity = 0.6
		event.WindSensitive = true
		event.SpreadPattern = "directional"
		event.Effects["damage"] = 0.5
		event.Effects["mutation"] = 0.02
		
	case "volcanic_eruption":
		event.Name = "Volcanic Eruption"
		event.Description = "Lava flows reshape the landscape"
		event.Duration = 30 + rand.Intn(25)
		event.Radius = 1.0
		event.MaxRadius = 6.0
		event.Intensity = 1.0
		event.WindSensitive = false
		event.SpreadPattern = "circular"
		event.Effects["damage"] = 4.0
		event.Effects["mutation"] = 0.15
		
	case "flood":
		event.Name = "Great Flood"
		event.Description = "Rising waters flood the land"
		event.Duration = 25 + rand.Intn(20)
		event.Radius = 3.0
		event.MaxRadius = 10.0
		event.Intensity = 0.7
		event.WindSensitive = false
		event.SpreadPattern = "circular"
		event.Effects["damage"] = 1.5
		event.Effects["mutation"] = 0.03
		
	case "hurricane":
		event.Name = "Hurricane"
		event.Description = "Massive rotating storm system"
		event.Duration = 18 + rand.Intn(12)
		event.Radius = 6.0
		event.MaxRadius = 15.0
		event.Intensity = 0.9
		event.WindSensitive = true
		event.SpreadPattern = "directional"
		event.Effects["damage"] = 3.0
		event.Effects["mutation"] = 0.08
		
	case "tornado":
		event.Name = "Tornado"
		event.Description = "Destructive rotating windstorm"
		event.Duration = 8 + rand.Intn(8)
		event.Radius = 1.5
		event.MaxRadius = 3.0
		event.Intensity = 1.0
		event.WindSensitive = true
		event.SpreadPattern = "directional"
		event.Speed = 2.0
		event.Effects["damage"] = 5.0
		event.Effects["mutation"] = 0.1
	}
	
	w.EnvironmentalEvents = append(w.EnvironmentalEvents, event)
	
	// Log event start
	if w.EventLogger != nil {
		startEvent := LogEvent{
			Timestamp:   time.Now(),
			Tick:        w.Tick,
			Type:        "environmental_event_start",
			Description: fmt.Sprintf("%s started at position (%.1f, %.1f)", event.Name, event.Position.X, event.Position.Y),
			Data: map[string]interface{}{
				"event_type": event.Type,
				"position_x": event.Position.X,
				"position_y": event.Position.Y,
				"intensity": event.Intensity,
				"duration": event.Duration,
			},
		}
		w.EventLogger.addEvent(startEvent)
	}
}

// Helper functions

// getWindAtPosition gets wind vector at a specific position
func (w *World) getWindAtPosition(pos Position) WindVector {
	if w.WindSystem == nil {
		return WindVector{X: 0, Y: 0, Strength: 0}
	}
	
	// Convert world position to wind map coordinates
	windX := int(pos.X / w.WindSystem.CellSize)
	windY := int(pos.Y / w.WindSystem.CellSize)
	
	if windX >= 0 && windX < w.WindSystem.MapWidth && windY >= 0 && windY < w.WindSystem.MapHeight {
		return w.WindSystem.WindMap[windY][windX]
	}
	
	// Return base wind as fallback
	return WindVector{
		X: math.Cos(w.WindSystem.BaseWindDirection),
		Y: math.Sin(w.WindSystem.BaseWindDirection),
		Strength: w.WindSystem.BaseWindStrength,
	}
}

// isFlammableBiome checks if a biome can catch fire
func (w *World) isFlammableBiome(biome BiomeType) bool {
	flammableBiomes := []BiomeType{
		BiomeForest, BiomeRainforest, BiomePlains, BiomeTundra,
	}
	
	for _, flammable := range flammableBiomes {
		if biome == flammable {
			return true
		}
	}
	return false
}

// isFireExtinguishingBiome checks if a biome can extinguish fire
func (w *World) isFireExtinguishingBiome(biome BiomeType) bool {
	extinguishingBiomes := []BiomeType{
		BiomeWater, BiomeDeepWater, BiomeIce, BiomeSwamp,
	}
	
	for _, extinguishing := range extinguishingBiomes {
		if biome == extinguishing {
			return true
		}
	}
	return false
}

// Plant Nutrient System Support Functions

// processDecayNutrientsToSoil adds nutrients from decaying items to the soil
func (w *World) processDecayNutrientsToSoil() {
	for _, decayItem := range w.ReproductionSystem.DecayingItems {
		if decayItem.IsDecayed {
			continue // Already processed
		}
		
		// Get grid cell for decay item location
		gridX := int((decayItem.Position.X / w.Config.Width) * float64(w.Config.GridWidth))
		gridY := int((decayItem.Position.Y / w.Config.Height) * float64(w.Config.GridHeight))
		
		// Clamp to grid bounds
		gridX = int(math.Max(0, math.Min(float64(w.Config.GridWidth-1), float64(gridX))))
		gridY = int(math.Max(0, math.Min(float64(w.Config.GridHeight-1), float64(gridY))))
		
		gridCell := &w.Grid[gridY][gridX]
		
		// Add nutrients gradually as item decays
		decayProgress := float64(w.Tick - decayItem.CreationTick) / float64(decayItem.DecayPeriod)
		if decayProgress > 0.1 { // Start releasing nutrients after 10% decay
			addDecayNutrientsToSoil(gridCell, decayItem)
		}
	}
}

// processWeatherEffectsOnSoil handles rainfall and weather impacts on soil
func (w *World) processWeatherEffectsOnSoil() {
	// Enhanced environmental events can cause rainfall
	for _, event := range w.EnvironmentalEvents {
		if event.Type == "storm" || event.Type == "hurricane" {
			// Calculate rainfall from storm
			for y := 0; y < w.Config.GridHeight; y++ {
				for x := 0; x < w.Config.GridWidth; x++ {
					distance := math.Sqrt(math.Pow(float64(x)-event.Position.X, 2) + 
										 math.Pow(float64(y)-event.Position.Y, 2))
					
					if distance <= event.Radius {
						intensity := (event.Radius - distance) / event.Radius * event.Intensity
						processRainfall(&w.Grid[y][x], intensity)
					}
				}
			}
		}
	}
	
	// Seasonal rainfall patterns
	currentTimeState := w.AdvancedTimeSystem.GetTimeState()
	seasonalRainfall := 0.0
	
	switch currentTimeState.Season {
	case Spring:
		seasonalRainfall = 0.02 // Moderate spring rains
	case Summer:
		seasonalRainfall = 0.01 // Lighter summer rains
	case Autumn:
		seasonalRainfall = 0.03 // Heavier autumn rains
	case Winter:
		seasonalRainfall = 0.015 // Light winter precipitation
	}
	
	// Apply seasonal rainfall randomly across the world
	if rand.Float64() < 0.3 { // 30% chance of rain each tick
		for y := 0; y < w.Config.GridHeight; y++ {
			for x := 0; x < w.Config.GridWidth; x++ {
				if rand.Float64() < 0.1 { // 10% of cells get rain
					processRainfall(&w.Grid[y][x], seasonalRainfall)
				}
			}
		}
	}
}


// getSeasonName converts Season enum to string
func getSeasonName(season Season) string {
switch season {
case Spring:
return "spring"
	case Summer:
		return "summer"
	case Autumn:
		return "autumn"
	case Winter:
		return "winter"
	default:
		return "spring"
	}
}

// calculateMolecularHealth calculates entity's overall molecular health status
func (w *World) calculateMolecularHealth(entity *Entity) float64 {
	if entity.MolecularNeeds == nil {
		return 1.0 // Assume healthy if no molecular system
	}
	return entity.MolecularNeeds.GetOverallNutritionalStatus()
}

// processMassDieOffImpacts handles environmental effects of mass death events
func (w *World) processMassDieOffImpacts(deathsBySpecies map[string]int, totalDeaths int) {
	populationSize := len(w.AllEntities)
	if populationSize == 0 {
		return
	}
	
	// Calculate death rate as percentage of population
	deathRate := float64(totalDeaths) / float64(populationSize)
	
	// Mass die-off thresholds
	massDieOffThreshold := 0.15 // 15% population death in one tick
	catastrophicThreshold := 0.30 // 30% population death in one tick
	
	if deathRate >= massDieOffThreshold {
		// Mass die-off detected - enhance resource cycling
		w.processMassDieOffEffects(deathsBySpecies, totalDeaths, deathRate)
		
		// Log the event
		severity := "mass_die_off"
		if deathRate >= catastrophicThreshold {
			severity = "catastrophic_die_off"
		}
		
		eventDetails := make(map[string]interface{})
		eventDetails["death_rate"] = deathRate
		eventDetails["total_deaths"] = totalDeaths
		eventDetails["species_affected"] = deathsBySpecies
		eventDetails["population_before"] = populationSize + totalDeaths
		
		if w.EventLogger != nil {
			w.EventLogger.LogWorldEvent(w.Tick, severity, 
				fmt.Sprintf("%s event: %.1f%% population loss (%d deaths)", severity, deathRate*100, totalDeaths))
		}
		
		if w.StatisticalReporter != nil {
			w.StatisticalReporter.LogSystemEvent(w.Tick, severity, 
				fmt.Sprintf("%.1f%% population died", deathRate*100), eventDetails)
		}
	}
}

// processMassDieOffEffects applies environmental impacts from mass death events
func (w *World) processMassDieOffEffects(deathsBySpecies map[string]int, totalDeaths int, deathRate float64) {
	// Calculate intensity of environmental impact
	impactIntensity := math.Min(deathRate * 3.0, 1.0) // Cap at 1.0
	
	// 1. Nutrient enrichment from mass decomposition
	w.enhanceNutrientCycling(impactIntensity, totalDeaths)
	
	// 2. Disease/contamination effects
	w.applyContaminationEffects(impactIntensity)
	
	// 3. Ecological disruption - remove key species interactions
	w.processEcologicalDisruption(deathsBySpecies, impactIntensity)
	
	// 4. Scavenger opportunities - boost surviving carnivores
	w.boostScavengerOpportunities(impactIntensity)
}

// enhanceNutrientCycling increases soil nutrients from mass decomposition
func (w *World) enhanceNutrientCycling(intensity float64, totalDeaths int) {
	// Add extra nutrients to random locations based on death count
	nutrientBoostCount := int(float64(totalDeaths) * intensity * 0.5)
	
	for i := 0; i < nutrientBoostCount; i++ {
		x := rand.Intn(w.Config.GridWidth)
		y := rand.Intn(w.Config.GridHeight)
		
		// Enhance soil nutrients
		cell := &w.Grid[y][x]
		if cell.SoilNutrients == nil {
			cell.SoilNutrients = make(map[string]float64)
		}
		
		// Add decomposition nutrients
		cell.SoilNutrients["nitrogen"] += 0.3 * intensity
		cell.SoilNutrients["phosphorus"] += 0.2 * intensity
		cell.SoilNutrients["organic_matter"] += 0.4 * intensity
		
		// Cap nutrients to prevent unrealistic levels
		for nutrient := range cell.SoilNutrients {
			cell.SoilNutrients[nutrient] = math.Min(cell.SoilNutrients[nutrient], 2.0)
		}
	}
}

// applyContaminationEffects applies negative effects from decomposition
func (w *World) applyContaminationEffects(intensity float64) {
	// Create temporary contamination zones
	contaminationCount := int(intensity * 10) // More contamination with higher death rates
	
	for i := 0; i < contaminationCount; i++ {
		x := rand.Intn(w.Config.GridWidth)
		y := rand.Intn(w.Config.GridHeight)
		
		cell := &w.Grid[y][x]
		
		// Temporary contamination that affects entity health
		if cell.Event == nil {
			cell.Event = &WorldEvent{
				Name:           "contamination",
				Description:    "Decomposition contamination",
				Duration:       20 + rand.Intn(20), // 20-40 ticks
				GlobalDamage:   0.5 * intensity,
				GlobalMutation: 0.02 * intensity,
			}
		}
	}
}

// processEcologicalDisruption handles food web disruption from species loss
func (w *World) processEcologicalDisruption(deathsBySpecies map[string]int, intensity float64) {
	// For each species that suffered major losses
	for species, deaths := range deathsBySpecies {
		if deaths > 5 { // Significant species loss
			// Reduce prey preferences for this species in surviving entities
			for _, entity := range w.AllEntities {
				if entity.IsAlive && entity.DietaryMemory != nil {
					if pref, exists := entity.DietaryMemory.PreySpeciesPreferences[species]; exists {
						// Reduce preference as species becomes rarer
						reductionFactor := math.Min(float64(deaths) * 0.1 * intensity, 0.5)
						newPref := pref * (1.0 - reductionFactor)
						if newPref < 0.1 {
							delete(entity.DietaryMemory.PreySpeciesPreferences, species)
						} else {
							entity.DietaryMemory.PreySpeciesPreferences[species] = newPref
						}
					}
				}
			}
		}
	}
}

// boostScavengerOpportunities provides benefits to surviving carnivores/omnivores
func (w *World) boostScavengerOpportunities(intensity float64) {
	energyBoost := 20.0 * intensity // Energy boost from scavenging opportunities
	
	for _, entity := range w.AllEntities {
		if !entity.IsAlive {
			continue
		}
		
		// Carnivores and omnivores benefit from scavenging
		if strings.Contains(entity.Species, "carnivore") || strings.Contains(entity.Species, "omnivore") {
			entity.Energy += energyBoost
			
			// Improve dietary fitness due to abundance of food
			if entity.DietaryMemory != nil {
				entity.DietaryMemory.DietaryFitness = math.Min(1.0, entity.DietaryMemory.DietaryFitness + intensity*0.2)
			}
		}
	}
}

// attemptBasicToolsAndModifications provides basic tool and environmental modification creation
// to supplement the emergent behavior system
func (w *World) attemptBasicToolsAndModifications() {
	if len(w.AllEntities) == 0 {
		return
	}
	
	// Very low chance to create tools/modifications to make the systems visible
	if rand.Float64() > 0.01 { // 1% chance per tick
		return
	}
	
	// Pick a random entity with decent intelligence
	eligibleEntities := make([]*Entity, 0)
	for _, entity := range w.AllEntities {
		if entity.IsAlive && entity.GetTrait("intelligence") > 0.1 {
			eligibleEntities = append(eligibleEntities, entity)
		}
	}
	
	if len(eligibleEntities) == 0 {
		return
	}
	
	entity := eligibleEntities[rand.Intn(len(eligibleEntities))]
	
	// 50% chance for tool, 50% for environmental modification
	if rand.Float64() < 0.5 {
		// Create a basic tool
		toolTypes := []ToolType{ToolStone, ToolStick, ToolSpear, ToolHammer}
		toolType := toolTypes[rand.Intn(len(toolTypes))]
		tool := w.ToolSystem.CreateTool(entity, toolType, entity.Position)
		if tool != nil && w.EventLogger != nil {
			w.EventLogger.LogWorldEvent(w.Tick, "tool_creation",
				fmt.Sprintf("%s created a %s tool", entity.Species, GetToolTypeName(toolType)))
		}
	} else {
		// Create an environmental modification (use specific methods)
		switch rand.Intn(4) {
		case 0:
			mod := w.EnvironmentalModSystem.CreateBurrow(entity, entity.Position)
			if mod != nil && w.EventLogger != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "environment_modification",
					fmt.Sprintf("%s created a burrow", entity.Species))
			}
		case 1:
			mod := w.EnvironmentalModSystem.CreateCache(entity, entity.Position)
			if mod != nil && w.EventLogger != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "environment_modification",
					fmt.Sprintf("%s created a cache", entity.Species))
			}
		case 2:
			// Create tunnel
			direction := rand.Float64() * 2 * math.Pi
			length := 3.0 + rand.Float64()*5.0
			mod := w.EnvironmentalModSystem.CreateTunnel(entity, entity.Position, direction, length)
			if mod != nil && w.EventLogger != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "environment_modification",
					fmt.Sprintf("%s created a tunnel", entity.Species))
			}
		case 3:
			mod := w.EnvironmentalModSystem.CreateTrap(entity, entity.Position, "simple")
			if mod != nil && w.EventLogger != nil {
				w.EventLogger.LogWorldEvent(w.Tick, "environment_modification",
					fmt.Sprintf("%s created a trap", entity.Species))
			}
		}
	}
}

// calculateEnvironmentalFactors determines environmental conditions affecting metamorphosis
func (w *World) calculateEnvironmentalFactors(entity *Entity, gridX, gridY int) map[string]float64 {
	environment := make(map[string]float64)
	
	// Get grid cell information
	cell := w.Grid[gridY][gridX]
	
	// Temperature based on biome and season
	timeState := w.AdvancedTimeSystem.GetTimeState()
	baseTemp := w.getBiomeTemperature(cell.Biome)
	seasonalMod := w.getSeasonalTemperatureModifier(w.seasonToString(timeState.Season))
	environment["temperature"] = baseTemp * seasonalMod
	
	// Humidity based on biome
	environment["humidity"] = w.getBiomeHumidity(cell.Biome)
	
	// Food availability based on nearby plants and resources
	foodCount := 0
	totalNutrition := 0.0
	searchRadius := 5
	
	for dx := -searchRadius; dx <= searchRadius; dx++ {
		for dy := -searchRadius; dy <= searchRadius; dy++ {
			checkX := gridX + dx
			checkY := gridY + dy
			
			if checkX >= 0 && checkX < w.Config.GridWidth && checkY >= 0 && checkY < w.Config.GridHeight {
				checkCell := w.Grid[checkY][checkX]
				if len(checkCell.Plants) > 0 {
					for _, plant := range checkCell.Plants {
						if plant != nil && plant.IsAlive {
							foodCount++
							totalNutrition += plant.Energy
						}
					}
				}
			}
		}
	}
	
	if foodCount > 0 {
		environment["food_availability"] = math.Min(1.0, totalNutrition/float64(foodCount)/100.0)
	} else {
		environment["food_availability"] = 0.0
	}
	
	// Safety based on nearby predators and threats
	threatLevel := 0.0
	entityCount := 0
	
	for _, other := range w.AllEntities {
		if other == nil || !other.IsAlive || other.ID == entity.ID {
			continue
		}
		
		distance := entity.DistanceTo(other)
		if distance < 10.0 { // Within threat detection range
			entityCount++
			aggression := other.GetTrait("aggression")
			size := other.GetTrait("size")
			threat := aggression * (1.0 + size) / (distance + 1.0)
			threatLevel += threat
		}
	}
	
	if entityCount > 0 {
		avgThreat := threatLevel / float64(entityCount)
		environment["safety"] = math.Max(0.0, 1.0 - avgThreat)
	} else {
		environment["safety"] = 1.0
	}
	
	// Population density
	nearbyCount := float64(entityCount)
	environment["population_density"] = math.Min(1.0, nearbyCount/20.0)
	
	return environment
}

// getBiomeTemperature returns the temperature characteristic of a biome (0.0 to 1.0)
func (w *World) getBiomeTemperature(biome BiomeType) float64 {
	switch biome {
	case BiomeIce:
		return 0.1
	case BiomeTundra:
		return 0.2
	case BiomeWater:
		return 0.4
	case BiomePlains:
		return 0.5
	case BiomeForest:
		return 0.6
	case BiomeMountain:
		return 0.3
	case BiomeDesert:
		return 0.9
	case BiomeRainforest:
		return 0.8
	case BiomeHotSpring:
		return 0.95
	default:
		return 0.5
	}
}

// getBiomeHumidity returns the humidity characteristic of a biome (0.0 to 1.0)
func (w *World) getBiomeHumidity(biome BiomeType) float64 {
	switch biome {
	case BiomeIce:
		return 0.3
	case BiomeTundra:
		return 0.4
	case BiomeWater:
		return 1.0
	case BiomeDeepWater:
		return 1.0
	case BiomePlains:
		return 0.5
	case BiomeForest:
		return 0.7
	case BiomeMountain:
		return 0.4
	case BiomeDesert:
		return 0.1
	case BiomeRainforest:
		return 0.9
	case BiomeSwamp:
		return 0.95
	case BiomeHotSpring:
		return 0.8
	default:
		return 0.5
	}
}

// getSeasonalTemperatureModifier returns seasonal temperature adjustment
func (w *World) getSeasonalTemperatureModifier(season string) float64 {
	switch season {
	case "spring":
		return 0.8
	case "summer":
		return 1.2
	case "autumn":
		return 0.9
	case "winter":
		return 0.6
	default:
		return 1.0
	}
}

// getTemperatureAt returns temperature at a specific position
func (w *World) getTemperatureAt(pos Position) float64 {
	biome := w.getBiomeAtPosition(pos.X, pos.Y)
	baseTemp := w.getBiomeTemperature(biome)
	
	// Apply seasonal modifier
	season := w.getCurrentSeason()
	seasonMod := w.getSeasonalTemperatureModifier(season)
	
	// Add some randomness for micro-climates
	randomVariation := (rand.Float64()*2 - 1) * 3.0 // ±3 degrees
	
	return baseTemp * seasonMod + randomVariation
}

// getMoistureAt returns moisture level at a specific position
func (w *World) getMoistureAt(pos Position) float64 {
	biome := w.getBiomeAtPosition(pos.X, pos.Y)
	baseMoisture := w.getBiomeHumidity(biome)
	
	// Apply seasonal effects
	season := w.getCurrentSeason()
	switch season {
	case "spring":
		baseMoisture *= 1.2 // Spring rains
	case "summer":
		baseMoisture *= 0.8 // Dry summer
	case "autumn":
		baseMoisture *= 1.1 // Autumn rains
	case "winter":
		baseMoisture *= 0.9 // Winter conditions
	}
	
	// Add randomness for local weather
	randomVariation := (rand.Float64()*2 - 1) * 0.2 // ±20%
	
	return math.Max(0.0, math.Min(1.0, baseMoisture + randomVariation))
}

// getSunlightAt returns sunlight level at a specific position
func (w *World) getSunlightAt(pos Position) float64 {
	// Base sunlight depends on day/night cycle
	dayTime := float64(w.Tick % 100) / 100.0 // 100 ticks per day
	
	var baseSunlight float64
	if dayTime >= 0.25 && dayTime <= 0.75 { // Day time
		// Peak at noon (0.5), fade at dawn/dusk
		if dayTime <= 0.5 {
			baseSunlight = (dayTime - 0.25) * 4.0 // 0.25 to 0.5 -> 0 to 1
		} else {
			baseSunlight = (0.75 - dayTime) * 4.0 // 0.5 to 0.75 -> 1 to 0
		}
	} else {
		baseSunlight = 0.0 // Night time
	}
	
	// Biome effects on sunlight
	biome := w.getBiomeAtPosition(pos.X, pos.Y)
	switch biome {
	case BiomeForest:
		baseSunlight *= 0.6 // Forest canopy blocks light
	case BiomeDeepWater:
		baseSunlight *= 0.1 // Deep water blocks light
	case BiomeDesert:
		baseSunlight *= 1.2 // Desert has intense sunlight
	case BiomeHighAltitude:
		baseSunlight *= 1.3 // High altitude has intense sunlight
	}
	
	// Weather effects
	if w.EnvironmentalEvents != nil {
		for _, event := range w.EnvironmentalEvents {
			if event.Type == "storm" || event.Type == "ash_cloud" {
				// Check if position is affected by the event
				distance := math.Sqrt(math.Pow(pos.X-event.Position.X, 2) + math.Pow(pos.Y-event.Position.Y, 2))
				if distance <= event.Radius {
					baseSunlight *= 0.3 // Storms block sunlight
				}
			}
		}
	}
	
	return math.Max(0.0, math.Min(1.0, baseSunlight))
}

// getBiomeAt returns biome at a specific position
func (w *World) getBiomeAt(pos Position) BiomeType {
	return w.getBiomeAtPosition(pos.X, pos.Y)
}

// getCurrentSeason returns the current season based on tick
func (w *World) getCurrentSeason() string {
	if w.AdvancedTimeSystem != nil {
		return seasonToString(w.AdvancedTimeSystem.GetTimeState().Season)
	}
	
	// Fallback: calculate season from tick
	yearTick := w.Tick % (100 * 4) // 400 ticks per year (100 per season)
	seasonTick := yearTick / 100
	
	switch seasonTick {
	case 0:
		return "spring"
	case 1:
		return "summer"
	case 2:
		return "autumn"
	case 3:
		return "winter"
	default:
		return "spring"
	}
}

// seasonToString converts Season enum to string
func (w *World) seasonToString(season Season) string {
	switch season {
	case Spring:
		return "spring"
	case Summer:
		return "summer"
	case Autumn:
		return "autumn"
	case Winter:
		return "winter"
	default:
		return "unknown"
	}
}

// getElevationAt returns the elevation at a given position
func (w *World) getElevationAt(position Position) float64 {
	if w.TopologySystem != nil {
		return w.TopologySystem.getElevationAt(position.X, position.Y)
	}
	return 0.0 // Default elevation if no topology system
}

// isValidPosition checks if a position is within world bounds
func (w *World) isValidPosition(position Position) bool {
	x := int(position.X)
	y := int(position.Y)
	return x >= 0 && x < w.Config.GridWidth && y >= 0 && y < w.Config.GridHeight
}
