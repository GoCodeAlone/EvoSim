package main

import (
	"fmt"
	"math"
	"strings"
)

// ViewManager handles rendering simulation state for different interfaces
type ViewManager struct {
	world *World
	// Historical data tracking
	populationHistory    []PopulationHistorySnapshot
	communicationHistory []CommunicationHistorySnapshot
	physicsHistory       []PhysicsHistorySnapshot
	maxHistoryLength     int
}

// NewViewManager creates a new view manager
func NewViewManager(world *World) *ViewManager {
	return &ViewManager{
		world:            world,
		maxHistoryLength: 50, // Keep last 50 snapshots
	}
}

// Historical data structures
type PopulationHistorySnapshot struct {
	Tick        int                `json:"tick"`
	Timestamp   string             `json:"timestamp"`
	Populations []PopulationData   `json:"populations"`
}

type CommunicationHistorySnapshot struct {
	Tick          int               `json:"tick"`
	Timestamp     string            `json:"timestamp"`
	ActiveSignals int               `json:"active_signals"`
	SignalTypes   map[string]int    `json:"signal_types"`
}

type PhysicsHistorySnapshot struct {
	Tick            int     `json:"tick"`
	Timestamp       string  `json:"timestamp"`
	Collisions      int     `json:"collisions"`
	AverageVelocity float64 `json:"average_velocity"`
	TotalMomentum   float64 `json:"total_momentum"`
}

// ViewData represents the current state of the simulation for rendering
type ViewData struct {
	Tick           int                    `json:"tick"`
	TimeString     string                 `json:"time_string"`
	EntityCount    int                    `json:"entity_count"`
	PlantCount     int                    `json:"plant_count"`
	PopulationCount int                   `json:"population_count"`
	EventCount     int                    `json:"event_count"`
	Grid           [][]CellData           `json:"grid"`
	Stats          map[string]interface{} `json:"stats"`
	Events         []EventData            `json:"events"`
	Populations    []PopulationData       `json:"populations"`
	Communication  CommunicationData      `json:"communication"`
	Civilization   CivilizationData       `json:"civilization"`
	Physics        PhysicsData            `json:"physics"`
	Wind           WindData               `json:"wind"`
	Species        SpeciesData            `json:"species"`
	Network        NetworkData            `json:"network"`
	DNA            DNAData                `json:"dna"`
	Cellular       CellularData           `json:"cellular"`
	Evolution      EvolutionData          `json:"evolution"`
	Topology       TopologyData           `json:"topology"`
	Tools          ToolData               `json:"tools"`
	EnvironmentalMod EnvironmentalModData `json:"environmental_mod"`
	EmergentBehavior EmergentBehaviorData `json:"emergent_behavior"`
	FeedbackLoops    FeedbackLoopData     `json:"feedback_loops"`
	Reproduction     ReproductionData     `json:"reproduction"`
	Statistical      StatisticalData      `json:"statistical"`
	Anomalies        AnomaliesData        `json:"anomalies"`
	// Historical data
	PopulationHistory    []PopulationHistorySnapshot    `json:"population_history"`
	CommunicationHistory []CommunicationHistorySnapshot `json:"communication_history"`
	PhysicsHistory       []PhysicsHistorySnapshot       `json:"physics_history"`
}

// CellData represents a single grid cell for rendering
type CellData struct {
	X            int      `json:"x"`
	Y            int      `json:"y"`
	Biome        string   `json:"biome"`
	BiomeSymbol  string   `json:"biome_symbol"`
	BiomeColor   string   `json:"biome_color"`
	EntityCount  int      `json:"entity_count"`
	EntitySymbol string   `json:"entity_symbol"`
	EntityColor  string   `json:"entity_color"`
	PlantCount   int      `json:"plant_count"`
	PlantSymbol  string   `json:"plant_symbol"`
	PlantColor   string   `json:"plant_color"`
	HasEvent     bool     `json:"has_event"`
	EventSymbol  string   `json:"event_symbol"`
}

// EventData represents an event for rendering
type EventData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Tick        int    `json:"tick"`
	Type        string `json:"type"`        // "active" or "historical"
	EventType   string `json:"event_type"` // Type of historical event
	Timestamp   string `json:"timestamp"`  // When the event occurred
}

// PopulationData represents population statistics
type PopulationData struct {
	Name         string             `json:"name"`
	Species      string             `json:"species"`
	Count        int                `json:"count"`
	AvgFitness   float64            `json:"avg_fitness"`
	AvgEnergy    float64            `json:"avg_energy"`
	AvgAge       float64            `json:"avg_age"`
	Generation   int                `json:"generation"`
	TraitAverages map[string]float64 `json:"trait_averages"`
	// Feedback loop adaptation data
	DietaryAdaptationCount  int     `json:"dietary_adaptation_count"`
	EnvAdaptationCount      int     `json:"env_adaptation_count"`
	AvgDietaryFitness      float64 `json:"avg_dietary_fitness"`
	AvgEnvFitness          float64 `json:"avg_env_fitness"`
	PlantPreferences       int     `json:"plant_preferences"`
	PreyPreferences        int     `json:"prey_preferences"`
}

// CommunicationData represents communication system state
type CommunicationData struct {
	ActiveSignals int `json:"active_signals"`
	SignalTypes   map[string]int `json:"signal_types"`
}

// CivilizationData represents civilization system state
type CivilizationData struct {
	TribesCount    int `json:"tribes_count"`
	StructureCount int `json:"structure_count"`
	TotalResources int `json:"total_resources"`
}

// PhysicsData represents physics system state
type PhysicsData struct {
	CollisionsLastTick int     `json:"collisions_last_tick"`
	AverageVelocity    float64 `json:"average_velocity"`
	TotalMomentum      float64 `json:"total_momentum"`
}

// WindData represents wind system state
type WindData struct {
	Direction       float64                `json:"direction"`
	Strength        float64                `json:"strength"`
	TurbulenceLevel float64                `json:"turbulence_level"`
	WeatherPattern  string                 `json:"weather_pattern"`
	PollenCount     int                    `json:"pollen_count"`
	SeedCount       int                    `json:"seed_count"`
	SeedBanks       int                    `json:"seed_banks"`
	GerminationEvents int                  `json:"germination_events"`
	DormancyActivations int                `json:"dormancy_activations"`
	DispersalStats  map[string]interface{} `json:"dispersal_stats"`
}

// SpeciesData represents species tracking state
type SpeciesData struct {
	ActiveSpecies     int                    `json:"active_species"`
	ExtinctSpecies    int                    `json:"extinct_species"`
	SpeciesDetails    []SpeciesDetailData    `json:"species_details"`
	TotalSpeciesEver  int                    `json:"total_species_ever"`
	SpeciesWithMembers int                   `json:"species_with_members"`
	SpeciesAwaitingExtinction int           `json:"species_awaiting_extinction"`
	HasSpeciationSystem bool                 `json:"has_speciation_system"`
}

// SpeciesDetailData represents individual species information
type SpeciesDetailData struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Population        int    `json:"population"`
	IsExtinct         bool   `json:"is_extinct"`
	FormationTick     int    `json:"formation_tick"`
	ExtinctionTick    int    `json:"extinction_tick"`    // 0 if not extinct/awaiting extinction
	PeakPopulation    int    `json:"peak_population"`
	AwaitingExtinction bool  `json:"awaiting_extinction"` // true if has 0 members but not extinct yet
}

// NetworkData represents plant network state
type NetworkData struct {
	ConnectionCount int `json:"connection_count"`
	SignalCount     int `json:"signal_count"`
	ClusterCount    int `json:"cluster_count"`
}

// DNAData represents DNA system state
type DNAData struct {
	OrganismCount      int     `json:"organism_count"`
	AverageMutations   float64 `json:"average_mutations"`
	AverageComplexity  float64 `json:"average_complexity"`
}

// CellularData represents cellular system state
type CellularData struct {
	TotalCells          int     `json:"total_cells"`
	AverageComplexity   float64 `json:"average_complexity"`
	CellDivisions       int     `json:"cell_divisions"`
}

// EvolutionData represents evolution tracking state
type EvolutionData struct {
	SpeciationEvents    int     `json:"speciation_events"`
	ExtinctionEvents    int     `json:"extinction_events"`
	GeneticDiversity    float64 `json:"genetic_diversity"`
	HasSpeciationSystem bool    `json:"has_speciation_system"`
	TotalPlantsTracked  int     `json:"total_plants_tracked"`
	ActivePlantCount    int     `json:"active_plant_count"`
	SpeciationDetected  bool    `json:"speciation_detected"`
}

// ToolData represents tool system state
type ToolData struct {
	TotalTools    int                    `json:"total_tools"`
	OwnedTools    int                    `json:"owned_tools"`
	DroppedTools  int                    `json:"dropped_tools"`
	AvgDurability float64                `json:"avg_durability"`
	AvgEfficiency float64                `json:"avg_efficiency"`
	ToolTypes     map[string]int         `json:"tool_types"`
}

// EnvironmentalModData represents environmental modification system state
type EnvironmentalModData struct {
	TotalModifications    int                    `json:"total_modifications"`
	ActiveModifications   int                    `json:"active_modifications"`
	InactiveModifications int                    `json:"inactive_modifications"`
	AvgDurability         float64                `json:"avg_durability"`
	TunnelNetworks        int                    `json:"tunnel_networks"`
	ModificationTypes     map[string]int         `json:"modification_types"`
}

// EmergentBehaviorData represents emergent behavior system state
type EmergentBehaviorData struct {
	TotalEntities       int                    `json:"total_entities"`
	BehaviorSpread      map[string]int         `json:"behavior_spread"`
	AvgProficiency      map[string]float64     `json:"avg_proficiency"`
	DiscoveredBehaviors int                    `json:"discovered_behaviors"`
}

// FeedbackLoopData represents feedback loop system state
type FeedbackLoopData struct {
	DietaryMemoryCount      int     `json:"dietary_memory_count"`
	EnvMemoryCount          int     `json:"env_memory_count"`
	AvgDietaryFitness       float64 `json:"avg_dietary_fitness"`
	AvgEnvFitness           float64 `json:"avg_env_fitness"`
	TotalPlantPreferences   int     `json:"total_plant_preferences"`
	TotalPreyPreferences    int     `json:"total_prey_preferences"`
	HighPressureEntities    int     `json:"high_pressure_entities"`
	EvolutionaryPressure    float64 `json:"evolutionary_pressure"`
}

// ReproductionData represents reproduction system state
type ReproductionData struct {
	ActiveEggs      int                    `json:"active_eggs"`
	DecayingItems   int                    `json:"decaying_items"`
	PregnantEntities int                   `json:"pregnant_entities"`
	ReadyToMate     int                    `json:"ready_to_mate"`
	MatingSeasonEntities int               `json:"mating_season_entities"`
	MigratingEntities int                  `json:"migrating_entities"`
	ReproductionModes map[string]int       `json:"reproduction_modes"`
	MatingStrategies map[string]int        `json:"mating_strategies"`
	SeasonalMatingRate float64             `json:"seasonal_mating_rate"`
	TerritoriesWithMating int              `json:"territories_with_mating"`
	CrossSpeciesMating int                 `json:"cross_species_mating"`
}

// TopologyData represents world topology state
type TopologyData struct {
	ElevationRange string  `json:"elevation_range"`
	FluidRegions   int     `json:"fluid_regions"`
	GeologicalAge  int     `json:"geological_age"`
}

// StatisticalData represents statistical analysis state
type StatisticalData struct {
	TotalEvents      int                     `json:"total_events"`
	TotalSnapshots   int                     `json:"total_snapshots"`
	TotalAnomalies   int                     `json:"total_anomalies"`
	TotalEnergy      float64                 `json:"total_energy"`
	EnergyChange     float64                 `json:"energy_change"`
	EnergyTrend      string                  `json:"energy_trend"`
	PopulationTrend  string                  `json:"population_trend"`
	RecentEvents     []StatisticalEventData  `json:"recent_events"`
	LatestSnapshot   *StatisticalSnapshotData `json:"latest_snapshot"`
}

// AnomaliesData represents anomaly detection state
type AnomaliesData struct {
	TotalAnomalies    int                    `json:"total_anomalies"`
	RecentAnomalies   []AnomalyData          `json:"recent_anomalies"`
	AnomalyTypes      map[string]int         `json:"anomaly_types"`
	Recommendations   []string               `json:"recommendations"`
}

// StatisticalEventData represents a statistical event for web interface
type StatisticalEventData struct {
	Tick        int     `json:"tick"`
	Type        string  `json:"type"`
	Target      string  `json:"target"`
	Change      float64 `json:"change"`
	Description string  `json:"description"`
}

// StatisticalSnapshotData represents a statistical snapshot for web interface
type StatisticalSnapshotData struct {
	Tick            int                    `json:"tick"`
	TotalEnergy     float64               `json:"total_energy"`
	PopulationCount int                   `json:"population_count"`
	TraitAverages   map[string]float64    `json:"trait_averages"`
	PhysicsMetrics  map[string]float64    `json:"physics_metrics"`
}

// AnomalyData represents an anomaly for web interface
type AnomalyData struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    float64 `json:"severity"`
	Confidence  float64 `json:"confidence"`
	Tick        int     `json:"tick"`
}

// GetCurrentViewData returns the current simulation state for rendering
func (vm *ViewManager) GetCurrentViewData() *ViewData {
	// Capture historical data every 5 ticks
	if vm.world.Tick%5 == 0 {
		vm.captureHistoricalData()
	}
	
	data := &ViewData{
		Tick:            vm.world.Tick,
		TimeString:      vm.getTimeString(),
		EntityCount:     len(vm.world.AllEntities),
		PlantCount:      len(vm.world.AllPlants),
		PopulationCount: len(vm.world.Populations),
		EventCount:      len(vm.world.Events),
		Grid:            vm.buildGridData(),
		Stats:           vm.getStatsData(),
		Events:          vm.getEventsData(),
		Populations:     vm.getPopulationsData(),
		Communication:   vm.getCommunicationData(),
		Civilization:    vm.getCivilizationData(),
		Physics:         vm.getPhysicsData(),
		Wind:            vm.getWindData(),
		Species:         vm.getSpeciesData(),
		Network:         vm.getNetworkData(),
		DNA:             vm.getDNAData(),
		Cellular:        vm.getCellularData(),
		Evolution:       vm.getEvolutionData(),
		Topology:        vm.getTopologyData(),
		Tools:           vm.getToolData(),
		EnvironmentalMod: vm.getEnvironmentalModData(),
		EmergentBehavior: vm.getEmergentBehaviorData(),
		FeedbackLoops:    vm.getFeedbackLoopData(),
		Reproduction:     vm.getReproductionData(),
		Statistical:      vm.getStatisticalData(),
		Anomalies:        vm.getAnomaliesData(),
		// Include historical data
		PopulationHistory:    vm.populationHistory,
		CommunicationHistory: vm.communicationHistory,
		PhysicsHistory:       vm.physicsHistory,
	}
	
	return data
}

// captureHistoricalData captures current state for historical tracking
func (vm *ViewManager) captureHistoricalData() {
	timestamp := vm.world.Clock.Format("15:04:05")
	
	// Capture population history
	popSnapshot := PopulationHistorySnapshot{
		Tick:        vm.world.Tick,
		Timestamp:   timestamp,
		Populations: vm.getPopulationsData(),
	}
	vm.populationHistory = append(vm.populationHistory, popSnapshot)
	
	// Capture communication history
	commData := vm.getCommunicationData()
	commSnapshot := CommunicationHistorySnapshot{
		Tick:          vm.world.Tick,
		Timestamp:     timestamp,
		ActiveSignals: commData.ActiveSignals,
		SignalTypes:   commData.SignalTypes,
	}
	vm.communicationHistory = append(vm.communicationHistory, commSnapshot)
	
	// Capture physics history
	physicsData := vm.getPhysicsData()
	physicsSnapshot := PhysicsHistorySnapshot{
		Tick:            vm.world.Tick,
		Timestamp:       timestamp,
		Collisions:      physicsData.CollisionsLastTick,
		AverageVelocity: physicsData.AverageVelocity,
		TotalMomentum:   physicsData.TotalMomentum,
	}
	vm.physicsHistory = append(vm.physicsHistory, physicsSnapshot)
	
	// Trim history to max length
	if len(vm.populationHistory) > vm.maxHistoryLength {
		vm.populationHistory = vm.populationHistory[1:]
	}
	if len(vm.communicationHistory) > vm.maxHistoryLength {
		vm.communicationHistory = vm.communicationHistory[1:]
	}
	if len(vm.physicsHistory) > vm.maxHistoryLength {
		vm.physicsHistory = vm.physicsHistory[1:]
	}
}

// buildGridData builds the grid representation
func (vm *ViewManager) buildGridData() [][]CellData {
	grid := make([][]CellData, vm.world.Config.GridHeight)
	totalEntities := 0
	totalPlants := 0
	
	for y := 0; y < vm.world.Config.GridHeight; y++ {
		grid[y] = make([]CellData, vm.world.Config.GridWidth)
		for x := 0; x < vm.world.Config.GridWidth; x++ {
			cell := vm.world.Grid[y][x]
			cellData := CellData{
				X:           x,
				Y:           y,
				EntityCount: len(cell.Entities),
				PlantCount:  len(cell.Plants),
				HasEvent:    cell.Event != nil,
			}
			
			totalEntities += len(cell.Entities)
			totalPlants += len(cell.Plants)
			
			// Set biome info
			cellData.Biome, cellData.BiomeSymbol, cellData.BiomeColor = vm.getBiomeInfo(cell.Biome)
			
			// Set entity info
			if len(cell.Entities) > 0 {
				cellData.EntitySymbol, cellData.EntityColor = vm.getEntityInfo(cell.Entities)
			}
			
			// Set plant info
			if len(cell.Plants) > 0 {
				cellData.PlantSymbol, cellData.PlantColor = vm.getPlantInfo(cell.Plants)
			}
			
			// Set event info
			if cell.Event != nil {
				cellData.EventSymbol = "⚡"
			}
			
			grid[y][x] = cellData
		}
	}
	
	// Debug: Log entity and plant counts
	if vm.world.Tick%20 == 0 { // Log every 20 ticks to avoid spam
		fmt.Printf("Grid Debug - Tick %d: Total entities in world: %d, entities in grid: %d, plants in grid: %d\n", 
			vm.world.Tick, len(vm.world.AllEntities), totalEntities, totalPlants)
	}
	
	return grid
}

// getBiomeInfo returns biome display information
func (vm *ViewManager) getBiomeInfo(biome BiomeType) (string, string, string) {
	biomes := map[BiomeType][]string{
		BiomePlains:       {"Plains", "•", "green"},
		BiomeForest:       {"Forest", "♠", "darkgreen"},
		BiomeDesert:       {"Desert", "~", "yellow"},
		BiomeMountain:     {"Mountain", "^", "gray"},
		BiomeWater:        {"Water", "≈", "blue"},
		BiomeRadiation:    {"Radiation", "☢", "red"},
		BiomeSoil:         {"Soil", "■", "brown"},
		BiomeAir:          {"Air", "○", "lightblue"},
		BiomeIce:          {"Ice", "❄", "white"},
		BiomeRainforest:   {"Rainforest", "🌳", "darkgreen"},
		BiomeDeepWater:    {"Deep Water", "≈", "darkblue"},
		BiomeHighAltitude: {"High Altitude", "▲", "lightgray"},
		BiomeHotSpring:    {"Hot Spring", "◉", "orange"},
		BiomeTundra:       {"Tundra", "○", "lightgray"},
		BiomeSwamp:        {"Swamp", "≋", "olive"},
		BiomeCanyon:       {"Canyon", "◢", "darkgray"},
	}
	
	if info, exists := biomes[biome]; exists {
		return info[0], info[1], info[2]
	}
	return "Unknown", "?", "white"
}

// getEntityInfo returns entity display information
func (vm *ViewManager) getEntityInfo(entities []*Entity) (string, string) {
	if len(entities) == 0 {
		return "", ""
	}
	
	// Use count-based symbols
	count := len(entities)
	if count == 1 {
		// Use species-based symbol for single entities
		return vm.getSpeciesSymbol(entities[0].Species), vm.getSpeciesColor(entities[0].Species)
	} else if count < 10 {
		return fmt.Sprintf("%d", count), "white"
	} else {
		return "+", "white"
	}
}

// getPlantInfo returns plant display information
func (vm *ViewManager) getPlantInfo(plants []*Plant) (string, string) {
	if len(plants) == 0 {
		return "", ""
	}
	
	// Get the most common plant type
	plantCounts := make(map[PlantType]int)
	for _, plant := range plants {
		plantCounts[plant.Type]++
	}
	
	var mostCommon PlantType
	maxCount := 0
	for plantType, count := range plantCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = plantType
		}
	}
	
	return vm.getPlantTypeSymbol(mostCommon), vm.getPlantTypeColor(mostCommon)
}

// getSpeciesSymbol returns symbol for species
func (vm *ViewManager) getSpeciesSymbol(species string) string {
	// Simple mapping for now
	symbols := map[string]string{
		"herbivore": "H",
		"predator":  "P",
		"omnivore":  "O",
	}
	
	if symbol, exists := symbols[species]; exists {
		return symbol
	}
	return "E" // Generic entity
}

// getSpeciesColor returns color for species
func (vm *ViewManager) getSpeciesColor(species string) string {
	colors := map[string]string{
		"herbivore": "green",
		"predator":  "red",
		"omnivore":  "blue",
	}
	
	if color, exists := colors[species]; exists {
		return color
	}
	return "white"
}

// getPlantTypeSymbol returns symbol for plant type
func (vm *ViewManager) getPlantTypeSymbol(plantType PlantType) string {
	symbols := map[PlantType]string{
		PlantGrass:    ".",
		PlantBush:     "♦",
		PlantTree:     "♠",
		PlantMushroom: "♪",
		PlantAlgae:    "≈",
		PlantCactus:   "†",
	}
	
	if symbol, exists := symbols[plantType]; exists {
		return symbol
	}
	return "?"
}

// getPlantTypeColor returns color for plant type
func (vm *ViewManager) getPlantTypeColor(plantType PlantType) string {
	colors := map[PlantType]string{
		PlantGrass:    "lightgreen",
		PlantBush:     "green",
		PlantTree:     "darkgreen",
		PlantMushroom: "purple",
		PlantAlgae:    "cyan",
		PlantCactus:   "olive",
	}
	
	if color, exists := colors[plantType]; exists {
		return color
	}
	return "green"
}

// getTimeString returns a formatted time string
func (vm *ViewManager) getTimeString() string {
	if vm.world.AdvancedTimeSystem != nil {
		timeOfDay := "☀️"
		if vm.world.AdvancedTimeSystem.TimeOfDay == Night {
			timeOfDay = "🌙"
		}
		
		return fmt.Sprintf("%s Day %d, Season %s", 
			timeOfDay,
			vm.world.AdvancedTimeSystem.DayNumber,
			vm.getSeasonName(vm.world.AdvancedTimeSystem.Season))
	}
	return "Time unknown"
}

// getSeasonName returns season name
func (vm *ViewManager) getSeasonName(season Season) string {
	seasons := map[Season]string{
		Spring: "Spring",
		Summer: "Summer",
		Autumn: "Autumn",
		Winter: "Winter",
	}
	
	if name, exists := seasons[season]; exists {
		return name
	}
	return "Unknown"
}

// Helper methods for getting various data sections
func (vm *ViewManager) getStatsData() map[string]interface{} {
	stats := make(map[string]interface{})
	
	if len(vm.world.AllEntities) > 0 {
		totalFitness := 0.0
		totalEnergy := 0.0
		totalAge := 0.0
		totalLifespanPercent := 0.0
		
		for _, entity := range vm.world.AllEntities {
			totalFitness += entity.Fitness
			totalEnergy += entity.Energy
			totalAge += float64(entity.Age)
			
			// Calculate age as percentage of max lifespan for better representation
			if entity.MaxLifespan > 0 {
				lifespanPercent := float64(entity.Age) / float64(entity.MaxLifespan) * 100.0
				totalLifespanPercent += lifespanPercent
			}
		}
		
		count := float64(len(vm.world.AllEntities))
		stats["avg_fitness"] = totalFitness / count
		stats["avg_energy"] = totalEnergy / count
		stats["avg_age"] = totalAge / count // Keep raw age for backward compatibility
		stats["avg_age_percent"] = totalLifespanPercent / count // Age as percentage of lifespan
	} else {
		// Provide default values when no entities exist
		stats["avg_fitness"] = 0.0
		stats["avg_energy"] = 0.0
		stats["avg_age"] = 0.0
		stats["avg_age_percent"] = 0.0
	}
	
	return stats
}

func (vm *ViewManager) getEventsData() []EventData {
	events := make([]EventData, 0)
	
	// Add current active events
	for _, event := range vm.world.Events {
		events = append(events, EventData{
			Name:        event.Name,
			Description: event.Description,
			Duration:    event.Duration,
			Tick:        vm.world.Tick,
			Type:        "active",
			EventType:   "world_event",
			Timestamp:   vm.world.Clock.Format("15:04:05"),
		})
	}
	
	// Add recent events from central event bus (prioritized)
	if vm.world.CentralEventBus != nil {
		centralEvents := vm.world.CentralEventBus.GetRecentEvents(15) // Show last 15 central events
		for _, centralEvent := range centralEvents {
			events = append(events, EventData{
				Name:        vm.formatEventName(centralEvent.Type),
				Description: centralEvent.Description,
				Duration:    0, // Central events have no duration
				Tick:        centralEvent.Tick,
				Type:        "central",
				EventType:   centralEvent.Type,
				Timestamp:   centralEvent.Timestamp.Format("15:04:05"),
			})
		}
	}
	
	// Add recent historical events from event logger (legacy)
	if vm.world.EventLogger != nil {
		historyCount := 5 // Reduced to 5 since we have central events
		logEvents := vm.world.EventLogger.Events
		startIdx := 0
		if len(logEvents) > historyCount {
			startIdx = len(logEvents) - historyCount
		}
		
		for i := startIdx; i < len(logEvents); i++ {
			logEvent := logEvents[i]
			events = append(events, EventData{
				Name:        vm.formatEventName(logEvent.Type),
				Description: logEvent.Description,
				Duration:    0, // Historical events have no duration
				Tick:        logEvent.Tick,
				Type:        "historical",
				EventType:   logEvent.Type,
				Timestamp:   logEvent.Timestamp.Format("15:04:05"),
			})
		}
	}
	
	return events
}

func (vm *ViewManager) getPopulationsData() []PopulationData {
	populations := make([]PopulationData, 0, len(vm.world.Populations))
	
	for name, pop := range vm.world.Populations {
		data := PopulationData{
			Name:          name,
			Species:       pop.Species,
			Count:         len(pop.Entities),
			TraitAverages: make(map[string]float64),
		}
		
		if len(pop.Entities) > 0 {
			// Calculate averages
			totalFitness := 0.0
			totalEnergy := 0.0
			totalAge := 0.0
			traitSums := make(map[string]float64)
			
			// Feedback loop adaptation data
			dietaryMemoryCount := 0
			envMemoryCount := 0
			totalDietaryFitness := 0.0
			totalEnvFitness := 0.0
			plantPrefs := 0
			preyPrefs := 0
			
			for _, entity := range pop.Entities {
				if entity != nil && entity.IsAlive {
					totalFitness += entity.Fitness
					totalEnergy += entity.Energy
					totalAge += float64(entity.Age)
					
					for traitName, trait := range entity.Traits {
						traitSums[traitName] += trait.Value
					}
					
					// Feedback loop data - add safety checks
					if entity.DietaryMemory != nil {
						dietaryMemoryCount++
						totalDietaryFitness += entity.DietaryMemory.DietaryFitness
						if entity.DietaryMemory.PlantTypePreferences != nil {
							plantPrefs += len(entity.DietaryMemory.PlantTypePreferences)
						}
						if entity.DietaryMemory.PreySpeciesPreferences != nil {
							preyPrefs += len(entity.DietaryMemory.PreySpeciesPreferences)
						}
					}
					
					if entity.EnvironmentalMemory != nil {
						envMemoryCount++
						totalEnvFitness += entity.EnvironmentalMemory.AdaptationFitness
					}
				}
			}
			
			count := float64(len(pop.Entities))
			data.AvgFitness = totalFitness / count
			data.AvgEnergy = totalEnergy / count
			data.AvgAge = totalAge / count
			
			for traitName, sum := range traitSums {
				data.TraitAverages[traitName] = sum / count
			}
			
			// Add feedback loop data
			data.DietaryAdaptationCount = dietaryMemoryCount
			data.EnvAdaptationCount = envMemoryCount
			data.PlantPreferences = plantPrefs
			data.PreyPreferences = preyPrefs
			
			if dietaryMemoryCount > 0 {
				data.AvgDietaryFitness = totalDietaryFitness / float64(dietaryMemoryCount)
			}
			
			if envMemoryCount > 0 {
				data.AvgEnvFitness = totalEnvFitness / float64(envMemoryCount)
			}
		}
		
		populations = append(populations, data)
	}
	
	return populations
}

func (vm *ViewManager) getCommunicationData() CommunicationData {
	data := CommunicationData{
		SignalTypes: make(map[string]int),
	}
	
	if vm.world.CommunicationSystem != nil {
		data.ActiveSignals = len(vm.world.CommunicationSystem.Signals)
		
		// Count signal types
		for _, signal := range vm.world.CommunicationSystem.Signals {
			typeName := vm.getSignalTypeName(signal.Type)
			data.SignalTypes[typeName]++
		}
	}
	
	return data
}

func (vm *ViewManager) getSignalTypeName(signalType SignalType) string {
	names := map[SignalType]string{
		SignalDanger:    "Danger",
		SignalFood:      "Food",
		SignalMating:    "Mating",
		SignalTerritory: "Territory",
		SignalHelp:      "Help",
		SignalMigration: "Migration",
	}
	
	if name, exists := names[signalType]; exists {
		return name
	}
	return "Unknown"
}

func (vm *ViewManager) getCivilizationData() CivilizationData {
	data := CivilizationData{}
	
	if vm.world.CivilizationSystem != nil {
		data.TribesCount = len(vm.world.CivilizationSystem.Tribes)
		
		for _, tribe := range vm.world.CivilizationSystem.Tribes {
			data.StructureCount += len(tribe.Structures)
			data.TotalResources += int(tribe.Resources["food"]) + int(tribe.Resources["materials"])
		}
	}
	
	return data
}

func (vm *ViewManager) getPhysicsData() PhysicsData {
	data := PhysicsData{}
	
	if vm.world.PhysicsSystem != nil {
		data.CollisionsLastTick = vm.world.PhysicsSystem.CollisionsThisTick
		
		// Calculate average velocity
		if len(vm.world.PhysicsComponents) > 0 {
			totalVelocity := 0.0
			totalMomentum := 0.0
			
			for _, component := range vm.world.PhysicsComponents {
				velocity := math.Sqrt(component.Velocity.X*component.Velocity.X + component.Velocity.Y*component.Velocity.Y)
				totalVelocity += velocity
				totalMomentum += component.Mass * velocity
			}
			
			count := float64(len(vm.world.PhysicsComponents))
			data.AverageVelocity = totalVelocity / count
			data.TotalMomentum = totalMomentum
		}
	}
	
	return data
}

func (vm *ViewManager) getWindData() WindData {
	data := WindData{}
	
	if vm.world.WindSystem != nil {
		data.Direction = vm.world.WindSystem.BaseWindDirection
		data.Strength = vm.world.WindSystem.BaseWindStrength
		data.TurbulenceLevel = vm.world.WindSystem.TurbulenceLevel
		data.WeatherPattern = vm.getWeatherPatternName(vm.world.WindSystem.WeatherPattern)
		data.PollenCount = len(vm.world.WindSystem.AllPollenGrains)
	}
	
	// Add seed dispersal system data
	if vm.world.SeedDispersalSystem != nil {
		data.SeedCount = len(vm.world.SeedDispersalSystem.AllSeeds)
		data.SeedBanks = len(vm.world.SeedDispersalSystem.SeedBanks)
		data.GerminationEvents = vm.world.SeedDispersalSystem.GerminationEvents
		data.DormancyActivations = vm.world.SeedDispersalSystem.DormancyActivations
		data.DispersalStats = vm.world.SeedDispersalSystem.GetStats()
	}
	
	return data
}

func (vm *ViewManager) getWeatherPatternName(pattern int) string {
	patterns := map[int]string{
		0: "Calm",
		1: "Windy",
		2: "Storm",
	}
	
	if name, exists := patterns[pattern]; exists {
		return name
	}
	return "Unknown"
}

func (vm *ViewManager) getSpeciesData() SpeciesData {
	data := SpeciesData{
		SpeciesDetails:            make([]SpeciesDetailData, 0),
		HasSpeciationSystem:       vm.world.SpeciationSystem != nil,
		SpeciesWithMembers:        0,
		SpeciesAwaitingExtinction: 0,
	}
	
	// If we have a speciation system, use its data
	if vm.world.SpeciationSystem != nil {
		data.ActiveSpecies = len(vm.world.SpeciationSystem.ActiveSpecies)
		data.ExtinctSpecies = len(vm.world.SpeciationSystem.AllSpecies) - len(vm.world.SpeciationSystem.ActiveSpecies)
		data.TotalSpeciesEver = len(vm.world.SpeciationSystem.AllSpecies)
		
		// Include all species from AllSpecies (both active and extinct)
		for _, species := range vm.world.SpeciationSystem.AllSpecies {
			population := len(species.Members)
			awaitingExtinction := population == 0 && species.ExtinctionTick > 0
			
			if population > 0 {
				data.SpeciesWithMembers++
			}
			if awaitingExtinction {
				data.SpeciesAwaitingExtinction++
			}
			
			detail := SpeciesDetailData{
				ID:                 species.ID,
				Name:               species.Name,
				Population:         population,
				IsExtinct:          species.IsExtinct,
				FormationTick:      species.FormationTick,
				ExtinctionTick:     species.ExtinctionTick,
				PeakPopulation:     species.PeakPopulation,
				AwaitingExtinction: awaitingExtinction,
			}
			data.SpeciesDetails = append(data.SpeciesDetails, detail)
		}
	} else {
		// Fall back to basic population data if no speciation system
		data.ActiveSpecies = len(vm.world.Populations)
		data.TotalSpeciesEver = len(vm.world.Populations)
		
		// Create species details from populations
		for name, population := range vm.world.Populations {
			livingCount := 0
			for _, entity := range population.Entities {
				if entity.IsAlive {
					livingCount++
				}
			}
			
			if livingCount > 0 {
				data.SpeciesWithMembers++
			}
			
			detail := SpeciesDetailData{
				ID:                 0, // No ID in basic populations
				Name:               name,
				Population:         livingCount,
				IsExtinct:          livingCount == 0,
				FormationTick:      0, // Unknown formation tick
				ExtinctionTick:     0,
				PeakPopulation:     livingCount, // Use current as peak for simplicity
				AwaitingExtinction: livingCount == 0,
			}
			data.SpeciesDetails = append(data.SpeciesDetails, detail)
		}
	}
	
	return data
}

func (vm *ViewManager) getNetworkData() NetworkData {
	data := NetworkData{}
	
	if vm.world.PlantNetworkSystem != nil {
		data.ConnectionCount = len(vm.world.PlantNetworkSystem.Connections)
		data.SignalCount = len(vm.world.PlantNetworkSystem.ChemicalSignals)
		data.ClusterCount = len(vm.world.PlantNetworkSystem.NetworkClusters)
	}
	
	return data
}

func (vm *ViewManager) getDNAData() DNAData {
	data := DNAData{}
	
	if vm.world.DNASystem != nil && vm.world.CellularSystem != nil {
		data.OrganismCount = len(vm.world.CellularSystem.OrganismMap)
		
		if data.OrganismCount > 0 {
			totalMutations := 0.0
			totalComplexity := 0.0
			
			for _, organism := range vm.world.CellularSystem.OrganismMap {
				if len(organism.Cells) > 0 && organism.Cells[0].DNA != nil {
					totalMutations += float64(organism.Cells[0].DNA.Mutations)
				}
				totalComplexity += float64(organism.ComplexityLevel)
			}
			
			count := float64(data.OrganismCount)
			data.AverageMutations = totalMutations / count
			data.AverageComplexity = totalComplexity / count
		}
	}
	
	return data
}

func (vm *ViewManager) getCellularData() CellularData {
	data := CellularData{}
	
	if vm.world.CellularSystem != nil {
		totalCells := 0
		totalComplexity := 0.0
		totalDivisions := 0
		
		for _, organism := range vm.world.CellularSystem.OrganismMap {
			totalCells += len(organism.Cells)
			totalComplexity += float64(organism.ComplexityLevel)
			totalDivisions += organism.CellDivisions
		}
		
		data.TotalCells = totalCells
		data.CellDivisions = totalDivisions
		
		if len(vm.world.CellularSystem.OrganismMap) > 0 {
			data.AverageComplexity = totalComplexity / float64(len(vm.world.CellularSystem.OrganismMap))
		}
	}
	
	return data
}

func (vm *ViewManager) getEvolutionData() EvolutionData {
	data := EvolutionData{
		HasSpeciationSystem: vm.world.SpeciationSystem != nil,
		ActivePlantCount:    len(vm.world.AllPlants),
	}
	
	if vm.world.SpeciationSystem != nil {
		data.SpeciationEvents = len(vm.world.SpeciationSystem.SpeciationEvents)
		data.ExtinctionEvents = len(vm.world.SpeciationSystem.ExtinctionEvents)
		data.TotalPlantsTracked = len(vm.world.AllPlants)
		
		// Consider speciation detected if we have any species or events
		data.SpeciationDetected = len(vm.world.SpeciationSystem.AllSpecies) > 0 || 
								  len(vm.world.SpeciationSystem.SpeciationEvents) > 0
		
		// Calculate genetic diversity as average distance between species
		activeSpeciesCount := len(vm.world.SpeciationSystem.ActiveSpecies)
		if activeSpeciesCount > 1 {
			// Simplified diversity calculation
			data.GeneticDiversity = float64(activeSpeciesCount) / 10.0
		} else if activeSpeciesCount == 1 {
			// Single species = low diversity but not zero
			data.GeneticDiversity = 0.1
		}
	}
	
	return data
}

func (vm *ViewManager) getTopologyData() TopologyData {
	data := TopologyData{}
	
	if vm.world.TopologySystem != nil {
		data.FluidRegions = len(vm.world.FluidRegions)
		data.GeologicalAge = vm.world.Tick / 1000 // Simplified age calculation
		
		// Find elevation range
		minElev, maxElev := 0.0, 0.0
		if len(vm.world.TopologySystem.TopologyGrid) > 0 {
			first := true
			for _, row := range vm.world.TopologySystem.TopologyGrid {
				for _, cell := range row {
					elev := cell.Elevation
					if first {
						minElev, maxElev = elev, elev
						first = false
					} else {
						if elev < minElev {
							minElev = elev
						}
						if elev > maxElev {
							maxElev = elev
						}
					}
				}
			}
		}
		
		data.ElevationRange = fmt.Sprintf("%.1f to %.1f", minElev, maxElev)
	}
	
	return data
}

// RenderGridAsText renders the grid as text for CLI or text-based interfaces
func (vm *ViewManager) RenderGridAsText(viewData *ViewData, width, height int) string {
	var result strings.Builder
	
	maxX := min(width, len(viewData.Grid[0]))
	maxY := min(height, len(viewData.Grid))
	
	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			cell := viewData.Grid[y][x]
			
			// Determine what symbol to display (priority: entities > plants > biome)
			if cell.EntityCount > 0 {
				result.WriteString(cell.EntitySymbol)
			} else if cell.PlantCount > 0 {
				result.WriteString(cell.PlantSymbol)
			} else {
				result.WriteString(cell.BiomeSymbol)
			}
		}
		if y < maxY-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// GetViewModes returns available view modes
func (vm *ViewManager) GetViewModes() []string {
	return []string{
		"GRID", "STATS", "EVENTS", "POPULATIONS", "COMMUNICATION",
		"CIVILIZATION", "PHYSICS", "WIND", "SPECIES", "NETWORK",
		"DNA", "CELLULAR", "EVOLUTION", "TOPOLOGY", "TOOLS", "ENVIRONMENT", "BEHAVIOR",
		"STATISTICAL", "ANOMALIES",
	}
}

func (vm *ViewManager) getToolData() ToolData {
	data := ToolData{}
	
	if vm.world.ToolSystem != nil {
		stats := vm.world.ToolSystem.GetToolStats()
		
		if totalTools, ok := stats["total_tools"].(int); ok {
			data.TotalTools = totalTools
		}
		if ownedTools, ok := stats["owned_tools"].(int); ok {
			data.OwnedTools = ownedTools
		}
		if droppedTools, ok := stats["dropped_tools"].(int); ok {
			data.DroppedTools = droppedTools
		}
		if avgDurability, ok := stats["avg_durability"].(float64); ok {
			data.AvgDurability = avgDurability
		}
		if avgEfficiency, ok := stats["avg_efficiency"].(float64); ok {
			data.AvgEfficiency = avgEfficiency
		}
		
		data.ToolTypes = make(map[string]int)
		if toolTypes, ok := stats["tool_types"].(map[ToolType]int); ok {
			for toolType, count := range toolTypes {
				data.ToolTypes[GetToolTypeName(toolType)] = count
			}
		}
	}
	
	return data
}

func (vm *ViewManager) getEnvironmentalModData() EnvironmentalModData {
	data := EnvironmentalModData{}
	
	if vm.world.EnvironmentalModSystem != nil {
		stats := vm.world.EnvironmentalModSystem.GetModificationStats()
		
		if totalMods, ok := stats["total_modifications"].(int); ok {
			data.TotalModifications = totalMods
		}
		if activeMods, ok := stats["active_modifications"].(int); ok {
			data.ActiveModifications = activeMods
		}
		if inactiveMods, ok := stats["inactive_modifications"].(int); ok {
			data.InactiveModifications = inactiveMods
		}
		if avgDurability, ok := stats["avg_durability"].(float64); ok {
			data.AvgDurability = avgDurability
		}
		if tunnelNetworks, ok := stats["tunnel_networks"].(int); ok {
			data.TunnelNetworks = tunnelNetworks
		}
		
		data.ModificationTypes = make(map[string]int)
		if modTypes, ok := stats["modification_types"].(map[EnvironmentalModType]int); ok {
			for modType, count := range modTypes {
				data.ModificationTypes[GetEnvironmentalModTypeName(modType)] = count
			}
		}
	}
	
	return data
}

func (vm *ViewManager) getEmergentBehaviorData() EmergentBehaviorData {
	data := EmergentBehaviorData{}
	
	if vm.world.EmergentBehaviorSystem != nil {
		stats := vm.world.EmergentBehaviorSystem.GetBehaviorStats()
		
		if totalEntities, ok := stats["total_entities"].(int); ok {
			data.TotalEntities = totalEntities
		}
		if discoveredBehaviors, ok := stats["discovered_behaviors"].(int); ok {
			data.DiscoveredBehaviors = discoveredBehaviors
		}
		
		data.BehaviorSpread = make(map[string]int)
		if behaviorSpread, ok := stats["behavior_spread"].(map[string]int); ok {
			data.BehaviorSpread = behaviorSpread
		}
		
		data.AvgProficiency = make(map[string]float64)
		if avgProficiency, ok := stats["avg_proficiency"].(map[string]float64); ok {
			data.AvgProficiency = avgProficiency
		}
	}
	
	return data
}

// getFeedbackLoopData returns feedback loop adaptation system data
func (vm *ViewManager) getFeedbackLoopData() FeedbackLoopData {
	data := FeedbackLoopData{}
	
	dietaryMemoryCount := 0
	envMemoryCount := 0
	totalDietaryFitness := 0.0
	totalEnvFitness := 0.0
	plantPreferences := 0
	preyPreferences := 0
	highPressureCount := 0
	totalPressure := 0.0
	entityCount := 0
	
	// Collect data from all entities
	for _, population := range vm.world.Populations {
		for _, entity := range population.Entities {
			if !entity.IsAlive {
				continue
			}
			entityCount++
			
			// Check dietary memory
			if entity.DietaryMemory != nil {
				dietaryMemoryCount++
				totalDietaryFitness += entity.DietaryMemory.DietaryFitness
				if entity.DietaryMemory.PlantTypePreferences != nil {
					plantPreferences += len(entity.DietaryMemory.PlantTypePreferences)
				}
				if entity.DietaryMemory.PreySpeciesPreferences != nil {
					preyPreferences += len(entity.DietaryMemory.PreySpeciesPreferences)
				}
			}
			
			// Check environmental memory
			if entity.EnvironmentalMemory != nil {
				envMemoryCount++
				totalEnvFitness += entity.EnvironmentalMemory.AdaptationFitness
			}
			
			// Calculate evolutionary pressure on this entity
			entityPressure := 0.0
			if entity.EnvironmentalMemory != nil {
				entityPressure += entity.EnvironmentalMemory.RadiationPressure * 0.1
				entityPressure += entity.EnvironmentalMemory.TemperaturePressure * 0.05
				if entity.EnvironmentalMemory.AdaptationFitness < 0.8 {
					entityPressure += (0.8 - entity.EnvironmentalMemory.AdaptationFitness) * 0.3
				}
			}
			if entity.DietaryMemory != nil && entity.DietaryMemory.DietaryFitness < 0.6 {
				entityPressure += (0.6 - entity.DietaryMemory.DietaryFitness) * 0.2
			}
			
			totalPressure += entityPressure
			if entityPressure > 0.1 { // Threshold for "high pressure"
				highPressureCount++
			}
		}
	}
	
	data.DietaryMemoryCount = dietaryMemoryCount
	data.EnvMemoryCount = envMemoryCount
	data.TotalPlantPreferences = plantPreferences
	data.TotalPreyPreferences = preyPreferences
	data.HighPressureEntities = highPressureCount
	
	if dietaryMemoryCount > 0 {
		data.AvgDietaryFitness = totalDietaryFitness / float64(dietaryMemoryCount)
	}
	
	if envMemoryCount > 0 {
		data.AvgEnvFitness = totalEnvFitness / float64(envMemoryCount)
	}
	
	if entityCount > 0 {
		data.EvolutionaryPressure = totalPressure / float64(entityCount)
	}
	
	return data
}

// getReproductionData returns reproduction system state data
func (vm *ViewManager) getReproductionData() ReproductionData {
	data := ReproductionData{
		ReproductionModes: make(map[string]int),
		MatingStrategies:  make(map[string]int),
	}
	
	// Get data from reproduction system
	if vm.world.ReproductionSystem != nil {
		data.ActiveEggs = len(vm.world.ReproductionSystem.Eggs)
		data.DecayingItems = len(vm.world.ReproductionSystem.DecayingItems)
	}
	
	// Count entities by reproductive status
	pregnantCount := 0
	readyToMateCount := 0
	matingSeasonCount := 0
	migratingCount := 0
	crossSpeciesMating := 0
	territoriesWithMating := 0
	
	for _, entity := range vm.world.AllEntities {
		if !entity.IsAlive || entity.ReproductionStatus == nil {
			continue
		}
		
		rs := entity.ReproductionStatus
		
		// Count by reproduction mode and strategy
		data.ReproductionModes[rs.Mode.String()]++
		data.MatingStrategies[rs.Strategy.String()]++
		
		// Count by status
		if rs.IsPregnant {
			pregnantCount++
		}
		if rs.ReadyToMate {
			readyToMateCount++
		}
		if rs.MatingSeason {
			matingSeasonCount++
		}
		if rs.RequiresMigration {
			migratingCount++
		}
		
		// Check for cross-species mating potential
		if rs.Mate != nil && rs.Mate.Species != entity.Species {
			crossSpeciesMating++
		}
	}
	
	data.PregnantEntities = pregnantCount
	data.ReadyToMate = readyToMateCount
	data.MatingSeasonEntities = matingSeasonCount
	data.MigratingEntities = migratingCount
	data.CrossSpeciesMating = crossSpeciesMating
	
	// Calculate seasonal mating rate
	if vm.world.AdvancedTimeSystem != nil {
		switch vm.world.AdvancedTimeSystem.Season {
		case Spring:
			data.SeasonalMatingRate = 1.5 // 50% increase in spring
		case Summer:
			data.SeasonalMatingRate = 1.2 // 20% increase in summer
		case Autumn:
			data.SeasonalMatingRate = 0.8 // 20% decrease in autumn
		case Winter:
			data.SeasonalMatingRate = 0.5 // 50% decrease in winter
		}
	} else {
		data.SeasonalMatingRate = 1.0
	}
	
	// Count territories with active mating (simplified)
	if vm.world.CivilizationSystem != nil {
		for _, tribe := range vm.world.CivilizationSystem.Tribes {
			hasActiveMating := false
			for _, entity := range tribe.Members {
				if entity.ReproductionStatus != nil && entity.ReproductionStatus.ReadyToMate {
					hasActiveMating = true
					break
				}
			}
			if hasActiveMating {
				territoriesWithMating++
			}
		}
	}
	data.TerritoriesWithMating = territoriesWithMating
	
	return data
}

// formatEventName converts event type to display name
func (vm *ViewManager) formatEventName(eventType string) string {
	names := map[string]string{
		"species_extinction": "Species Extinction",
		"species_evolution":  "Species Evolution",
		"world_event":        "World Event",
		"population_boom":    "Population Boom",
		"population_crash":   "Population Crash",
		"new_species":        "New Species",
		"major_mutation":     "Major Mutation",
		"plant_evolution":    "Plant Evolution",
		"ecosystem_shift":    "Ecosystem Shift",
	}
	
	if name, exists := names[eventType]; exists {
		return name
	}
	return eventType
}

// getStatisticalData returns statistical analysis data for web interface
func (vm *ViewManager) getStatisticalData() StatisticalData {
	if vm.world.StatisticalReporter == nil {
		return StatisticalData{}
	}

	reporter := vm.world.StatisticalReporter
	
	// Get recent events (last 20)
	recentEvents := make([]StatisticalEventData, 0)
	if len(reporter.Events) > 0 {
		startIndex := 0
		if len(reporter.Events) > 20 {
			startIndex = len(reporter.Events) - 20
		}
		for i := startIndex; i < len(reporter.Events); i++ {
			event := reporter.Events[i]
			targetID := ""
			if event.EntityID != 0 {
				targetID = fmt.Sprintf("entity-%d", event.EntityID)
			} else if event.PlantID != 0 {
				targetID = fmt.Sprintf("plant-%d", event.PlantID)
			} else {
				targetID = event.Category
			}
			
			description := event.EventType
			if len(event.Metadata) > 0 {
				if desc, ok := event.Metadata["description"].(string); ok {
					description = desc
				}
			}
			
			recentEvents = append(recentEvents, StatisticalEventData{
				Tick:        event.Tick,
				Type:        event.EventType,
				Target:      targetID,
				Change:      event.Change,
				Description: description,
			})
		}
	}

	// Get latest snapshot
	var latestSnapshot *StatisticalSnapshotData
	if len(reporter.Snapshots) > 0 {
		snapshot := reporter.Snapshots[len(reporter.Snapshots)-1]
		
		// Calculate total population count
		totalPop := 0
		for _, count := range snapshot.PopulationsBySpecies {
			totalPop += count
		}
		
		// Calculate trait averages from distributions
		traitAverages := make(map[string]float64)
		for trait, distribution := range snapshot.TraitDistributions {
			if len(distribution) > 0 {
				sum := 0.0
				for _, value := range distribution {
					sum += value
				}
				traitAverages[trait] = sum / float64(len(distribution))
			}
		}
		
		latestSnapshot = &StatisticalSnapshotData{
			Tick:            snapshot.Tick,
			TotalEnergy:     snapshot.TotalEnergy,
			PopulationCount: totalPop,
			TraitAverages:   traitAverages,
			PhysicsMetrics: map[string]float64{
				"total_momentum":    snapshot.PhysicsMetrics.TotalMomentum,
				"kinetic_energy":    snapshot.PhysicsMetrics.TotalKineticEnergy,
				"collisions":        float64(snapshot.PhysicsMetrics.CollisionCount),
				"average_velocity":  snapshot.PhysicsMetrics.AverageVelocity,
			},
		}
	}

	// Calculate energy trend
	energyTrend := "stable"
	if len(reporter.Snapshots) >= 2 {
		latest := reporter.Snapshots[len(reporter.Snapshots)-1]
		previous := reporter.Snapshots[len(reporter.Snapshots)-2]
		if latest.TotalEnergy > previous.TotalEnergy*1.1 {
			energyTrend = "increasing"
		} else if latest.TotalEnergy < previous.TotalEnergy*0.9 {
			energyTrend = "decreasing"
		}
	}

	// Calculate population trend
	popTrend := "stable"
	if len(reporter.Snapshots) >= 2 {
		latest := reporter.Snapshots[len(reporter.Snapshots)-1]
		previous := reporter.Snapshots[len(reporter.Snapshots)-2]
		
		// Calculate population counts
		latestPop := 0
		for _, count := range latest.PopulationsBySpecies {
			latestPop += count
		}
		previousPop := 0
		for _, count := range previous.PopulationsBySpecies {
			previousPop += count
		}
		
		if latestPop > int(float64(previousPop)*1.1) {
			popTrend = "growing"
		} else if latestPop < int(float64(previousPop)*0.9) {
			popTrend = "declining"
		}
	}

	// Calculate total energy
	totalEnergy := 0.0
	if latestSnapshot != nil {
		totalEnergy = latestSnapshot.TotalEnergy
	}

	return StatisticalData{
		TotalEvents:     len(reporter.Events),
		TotalSnapshots:  len(reporter.Snapshots),
		TotalAnomalies:  len(reporter.Anomalies),
		TotalEnergy:     totalEnergy,
		EnergyTrend:     energyTrend,
		PopulationTrend: popTrend,
		RecentEvents:    recentEvents,
		LatestSnapshot:  latestSnapshot,
	}
}

// getAnomaliesData returns anomaly detection data for web interface
func (vm *ViewManager) getAnomaliesData() AnomaliesData {
	if vm.world.StatisticalReporter == nil {
		return AnomaliesData{}
	}

	reporter := vm.world.StatisticalReporter
	
	// Get recent anomalies (last 50)
	allAnomalies := reporter.GetRecentAnomalies(50, vm.world.Tick)
	recentAnomalies := make([]AnomalyData, 0, len(allAnomalies))
	for _, anomaly := range allAnomalies {
		recentAnomalies = append(recentAnomalies, AnomalyData{
			Type:        string(anomaly.Type),
			Description: anomaly.Description,
			Severity:    anomaly.Severity,
			Confidence:  anomaly.Confidence,
			Tick:        anomaly.Tick,
		})
	}

	// Count anomaly types
	anomalyTypes := make(map[string]int)
	for _, anomaly := range allAnomalies {
		anomalyTypes[string(anomaly.Type)]++
	}

	// Generate recommendations
	recommendations := []string{}
	if anomalyTypes["energy_conservation"] > 0 {
		recommendations = append(recommendations, "Check entity/plant death and birth rates")
		recommendations = append(recommendations, "Verify energy gain/loss calculations are balanced")
	}
	if anomalyTypes["unrealistic_distribution"] > 0 {
		recommendations = append(recommendations, "Monitor trait mutation rates and selection pressure")
		recommendations = append(recommendations, "Check if genetic diversity is adequate")
	}
	if anomalyTypes["population_anomaly"] > 0 {
		recommendations = append(recommendations, "Review carrying capacity and resource availability")
		recommendations = append(recommendations, "Check reproduction and mortality rates")
	}

	return AnomaliesData{
		TotalAnomalies:  len(allAnomalies),
		RecentAnomalies: recentAnomalies,
		AnomalyTypes:    anomalyTypes,
		Recommendations: recommendations,
	}
}