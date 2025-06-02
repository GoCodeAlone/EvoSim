package main

import (
	"fmt"
	"math"
	"strings"
)

// ViewManager handles rendering simulation state for different interfaces
type ViewManager struct {
	world *World
}

// NewViewManager creates a new view manager
func NewViewManager(world *World) *ViewManager {
	return &ViewManager{
		world: world,
	}
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
	Direction       float64 `json:"direction"`
	Strength        float64 `json:"strength"`
	TurbulenceLevel float64 `json:"turbulence_level"`
	WeatherPattern  string  `json:"weather_pattern"`
	PollenCount     int     `json:"pollen_count"`
}

// SpeciesData represents species tracking state
type SpeciesData struct {
	ActiveSpecies   int                    `json:"active_species"`
	ExtinctSpecies  int                    `json:"extinct_species"`
	SpeciesDetails  []SpeciesDetailData    `json:"species_details"`
}

// SpeciesDetailData represents individual species information
type SpeciesDetailData struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Population int    `json:"population"`
	IsExtinct  bool   `json:"is_extinct"`
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
	SpeciationEvents int     `json:"speciation_events"`
	ExtinctionEvents int     `json:"extinction_events"`
	GeneticDiversity float64 `json:"genetic_diversity"`
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

// TopologyData represents world topology state
type TopologyData struct {
	ElevationRange string  `json:"elevation_range"`
	FluidRegions   int     `json:"fluid_regions"`
	GeologicalAge  int     `json:"geological_age"`
}

// GetCurrentViewData returns the current simulation state for rendering
func (vm *ViewManager) GetCurrentViewData() *ViewData {
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
	}
	
	return data
}

// buildGridData builds the grid representation
func (vm *ViewManager) buildGridData() [][]CellData {
	grid := make([][]CellData, vm.world.Config.GridHeight)
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
	return grid
}

// getBiomeInfo returns biome display information
func (vm *ViewManager) getBiomeInfo(biome BiomeType) (string, string, string) {
	biomes := map[BiomeType][]string{
		BiomePlains:    {"Plains", ".", "green"},
		BiomeForest:    {"Forest", "♠", "darkgreen"},
		BiomeDesert:    {"Desert", "~", "yellow"},
		BiomeMountain:  {"Mountain", "^", "gray"},
		BiomeWater:     {"Water", "≈", "blue"},
		BiomeRadiation: {"Radiation", "☢", "red"},
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
		
		for _, entity := range vm.world.AllEntities {
			totalFitness += entity.Fitness
			totalEnergy += entity.Energy
			totalAge += float64(entity.Age)
		}
		
		count := float64(len(vm.world.AllEntities))
		stats["avg_fitness"] = totalFitness / count
		stats["avg_energy"] = totalEnergy / count
		stats["avg_age"] = totalAge / count
	}
	
	return stats
}

func (vm *ViewManager) getEventsData() []EventData {
	events := make([]EventData, len(vm.world.Events))
	for i, event := range vm.world.Events {
		events[i] = EventData{
			Name:        event.Name,
			Description: event.Description,
			Duration:    event.Duration,
			Tick:        vm.world.Tick,
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
			
			for _, entity := range pop.Entities {
				if entity != nil {
					totalFitness += entity.Fitness
					totalEnergy += entity.Energy
					totalAge += float64(entity.Age)
					
					for traitName, trait := range entity.Traits {
						traitSums[traitName] += trait.Value
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
		SpeciesDetails: make([]SpeciesDetailData, 0),
	}
	
	if vm.world.SpeciationSystem != nil {
		data.ActiveSpecies = len(vm.world.SpeciationSystem.ActiveSpecies)
		data.ExtinctSpecies = len(vm.world.SpeciationSystem.AllSpecies) - len(vm.world.SpeciationSystem.ActiveSpecies)
		
		for _, species := range vm.world.SpeciationSystem.ActiveSpecies {
			detail := SpeciesDetailData{
				ID:         species.ID,
				Name:       species.Name,
				Population: len(species.Members),
				IsExtinct:  species.IsExtinct,
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
	data := EvolutionData{}
	
	if vm.world.SpeciationSystem != nil {
		data.SpeciationEvents = len(vm.world.SpeciationSystem.SpeciationEvents)
		data.ExtinctionEvents = len(vm.world.SpeciationSystem.ExtinctionEvents)
		
		// Calculate genetic diversity as average distance between species
		if len(vm.world.SpeciationSystem.ActiveSpecies) > 1 {
			// Simplified diversity calculation
			data.GeneticDiversity = float64(len(vm.world.SpeciationSystem.ActiveSpecies)) / 10.0
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