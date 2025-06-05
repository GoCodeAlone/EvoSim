package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// StateManager handles saving and loading simulation state
type StateManager struct {
	world *World
}

// NewStateManager creates a new state manager for the given world
func NewStateManager(world *World) *StateManager {
	return &StateManager{
		world: world,
	}
}

// SimulationState represents the complete state of the simulation
type SimulationState struct {
	Version     string                 `json:"version"`
	SavedAt     time.Time              `json:"saved_at"`
	Tick        int                    `json:"tick"`
	NextID      int                    `json:"next_id"`
	NextPlantID int                    `json:"next_plant_id"`
	Config      WorldConfig            `json:"config"`
	Entities    []*EntityState         `json:"entities"`
	Plants      []*PlantState          `json:"plants"`
	Biomes      [][]BiomeType          `json:"biomes"`
	Events      []*WorldEventState     `json:"events"`
	Time        AdvancedTimeState      `json:"time"`
	Wind        WindSystemState        `json:"wind"`
	Species     SpeciationSystemState  `json:"species"`
	Network     PlantNetworkState      `json:"network"`
}

// EntityState represents serializable entity data
type EntityState struct {
	ID       int                    `json:"id"`
	Species  string                 `json:"species"`
	Position Position               `json:"position"`
	Traits   map[string]float64     `json:"traits"`
	Fitness  float64                `json:"fitness"`
	Energy   float64                `json:"energy"`
	Age      int                    `json:"age"`
	DNA      *DNAState              `json:"dna,omitempty"`
	Cellular *CellularState         `json:"cellular,omitempty"`
}

// PlantState represents serializable plant data
type PlantState struct {
	ID         int                `json:"id"`
	Type       PlantType          `json:"type"`
	Position   Position           `json:"position"`
	Energy     float64            `json:"energy"`
	Age        int                `json:"age"`
	Size       float64            `json:"size"`
	Traits     map[string]float64 `json:"traits"`
	Generation int                `json:"generation"`
	IsAlive    bool               `json:"is_alive"`
	NutritionVal float64          `json:"nutrition_value"`
	Toxicity     float64          `json:"toxicity"`
	GrowthRate   float64          `json:"growth_rate"`
}

// WorldEventState represents serializable world event data
type WorldEventState struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Duration       int                    `json:"duration"`
	GlobalMutation float64                `json:"global_mutation"`
	GlobalDamage   float64                `json:"global_damage"`
	BiomeChanges   map[string]BiomeType   `json:"biome_changes"`
}

// AdvancedTimeState represents serializable time system data
type AdvancedTimeState struct {
	WorldTick    int       `json:"world_tick"`
	DayLength    int       `json:"day_length"`
	SeasonLength int       `json:"season_length"`
	TimeOfDay    TimeOfDay `json:"time_of_day"`
	Season       Season    `json:"season"`
	DayNumber    int       `json:"day_number"`
	SeasonDay    int       `json:"season_day"`
	Temperature  float64   `json:"temperature"`
	Illumination float64   `json:"illumination"`
	SeasonalMod  float64   `json:"seasonal_mod"`
}

// WindSystemState represents serializable wind system data
type WindSystemState struct {
	BaseWindDirection   float64 `json:"base_wind_direction"`
	BaseWindStrength    float64 `json:"base_wind_strength"`
	TurbulenceLevel     float64 `json:"turbulence_level"`
	SeasonalMultiplier  float64 `json:"seasonal_multiplier"`
	WeatherPattern      int     `json:"weather_pattern"`
}

// SpeciationSystemState represents serializable speciation data
type SpeciationSystemState struct {
	Species        map[string]*SpeciesState `json:"species"`
	NextSpeciesID  int                      `json:"next_species_id"`
}

// SpeciesState represents serializable species data
type SpeciesState struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	ParentID     int       `json:"parent_id"`
	CreatedAt    int       `json:"created_at"`
	ExtinctAt    int       `json:"extinct_at"`
	IsExtinct    bool      `json:"is_extinct"`
	Population   int       `json:"population"`
	BaseTraits   map[string]float64 `json:"base_traits"`
}

// PlantNetworkState represents serializable plant network data
type PlantNetworkState struct {
	Connections    []*NetworkConnectionState `json:"connections"`
	ActiveSignals  []*ChemicalSignalState    `json:"active_signals"`
}

// NetworkConnectionState represents serializable network connection data
type NetworkConnectionState struct {
	Plant1ID     int                 `json:"plant1_id"`
	Plant2ID     int                 `json:"plant2_id"`
	Type         NetworkConnectionType `json:"type"`
	Strength     float64             `json:"strength"`
	Health       float64             `json:"health"`
	Efficiency   float64             `json:"efficiency"`
	Age          int                 `json:"age"`
}

// ChemicalSignalState represents serializable chemical signal data
type ChemicalSignalState struct {
	ID        int                    `json:"id"`
	SourceID  int                    `json:"source_id"`
	Type      ChemicalSignalType     `json:"type"`
	Intensity float64                `json:"intensity"`
	Age       int                    `json:"age"`
	MaxAge    int                    `json:"max_age"`
	Visited   map[int]bool           `json:"visited"`
	Message   string                 `json:"message"`
	Metadata  map[string]float64     `json:"metadata"`
}

// DNAState represents serializable DNA data
type DNAState struct {
	EntityID    int                `json:"entity_id"`
	Chromosomes []ChromosomeState  `json:"chromosomes"`
	Mutations   int                `json:"mutations"`
	Generation  int                `json:"generation"`
}

// ChromosomeState represents serializable chromosome data
type ChromosomeState struct {
	ID    int         `json:"id"`
	Genes []GeneState `json:"genes"`
}

// GeneState represents serializable gene data
type GeneState struct {
	Name       string       `json:"name"`
	Sequence   []string     `json:"sequence"`  // Convert nucleotides to strings
	Dominant   bool         `json:"dominant"`
	Expression float64      `json:"expression"`
}

// CellularState represents serializable cellular data
type CellularState struct {
	EntityID        int                `json:"entity_id"`
	ComplexityLevel int                `json:"complexity_level"`
	TotalEnergy     float64            `json:"total_energy"`
	CellDivisions   int                `json:"cell_divisions"`
	Generation      int                `json:"generation"`
	Cells           []CellState        `json:"cells"`
	OrganSystems    map[string][]int   `json:"organ_systems"`
}

// CellState represents serializable cell data
type CellState struct {
	ID          int                    `json:"id"`
	Type        CellType               `json:"type"`
	Size        float64                `json:"size"`
	Energy      float64                `json:"energy"`
	Health      float64                `json:"health"`
	Age         int                    `json:"age"`
	DNA         *DNAState              `json:"dna"`
	Organelles  map[string]OrganelleState `json:"organelles"` // Use string keys for JSON
	Position    Position               `json:"position"`
	Connections []int                  `json:"connections"`
	Activity    float64                `json:"activity"`
	Specialized bool                   `json:"specialized"`
}

// OrganelleState represents serializable organelle data
type OrganelleState struct {
	Type       OrganelleType `json:"type"`
	Count      int           `json:"count"`
	Efficiency float64       `json:"efficiency"`
	Energy     float64       `json:"energy"`
}

// SaveToFile saves the current simulation state to a JSON file
func (sm *StateManager) SaveToFile(filename string) error {
	state, err := sm.createState()
	if err != nil {
		return fmt.Errorf("failed to create state: %v", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %v", err)
	}

	fmt.Printf("Simulation state saved to %s\n", filename)
	return nil
}

// LoadFromFile loads simulation state from a JSON file
func (sm *StateManager) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read state file: %v", err)
	}

	var state SimulationState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %v", err)
	}

	err = sm.restoreState(&state)
	if err != nil {
		return fmt.Errorf("failed to restore state: %v", err)
	}

	fmt.Printf("Simulation state loaded from %s (saved at %s)\n", filename, state.SavedAt.Format(time.RFC3339))
	return nil
}

// LoadFromData loads simulation state from a map (used for web interface)
func (sm *StateManager) LoadFromData(data map[string]interface{}) error {
	// Convert map to JSON and then to SimulationState
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	var state SimulationState
	err = json.Unmarshal(jsonData, &state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %v", err)
	}

	err = sm.restoreState(&state)
	if err != nil {
		return fmt.Errorf("failed to restore state: %v", err)
	}

	fmt.Printf("Simulation state loaded from web interface (saved at %s)\n", state.SavedAt.Format(time.RFC3339))
	return nil
}

// createState converts the current world state to a serializable format
func (sm *StateManager) createState() (*SimulationState, error) {
	state := &SimulationState{
		Version:     "1.0",
		SavedAt:     time.Now(),
		Tick:        sm.world.Tick,
		NextID:      sm.world.NextID,
		NextPlantID: sm.world.NextPlantID,
		Config:      sm.world.Config,
		Entities:    make([]*EntityState, 0),
		Plants:      make([]*PlantState, 0),
		Biomes:      make([][]BiomeType, len(sm.world.Grid)),
		Events:      make([]*WorldEventState, 0),
	}

	// Convert biomes
	for y, row := range sm.world.Grid {
		state.Biomes[y] = make([]BiomeType, len(row))
		for x, cell := range row {
			state.Biomes[y][x] = cell.Biome
		}
	}

	// Convert entities
	for _, entity := range sm.world.AllEntities {
		entityState := &EntityState{
			ID:       entity.ID,
			Species:  entity.Species,
			Position: entity.Position,
			Traits:   make(map[string]float64),
			Fitness:  entity.Fitness,
			Energy:   entity.Energy,
			Age:      entity.Age,
		}

		// Copy traits
		for traitName, trait := range entity.Traits {
			entityState.Traits[traitName] = trait.Value
		}

		// Convert DNA if present
		if sm.world.CellularSystem != nil {
			if organism, exists := sm.world.CellularSystem.OrganismMap[entity.ID]; exists && len(organism.Cells) > 0 && organism.Cells[0].DNA != nil {
				entityState.DNA = sm.convertDNAToState(organism.Cells[0].DNA)
			}
		}

		// Convert Cellular if present
		if sm.world.CellularSystem != nil {
			if organism, exists := sm.world.CellularSystem.OrganismMap[entity.ID]; exists {
				entityState.Cellular = sm.convertCellularToState(organism)
			}
		}

		state.Entities = append(state.Entities, entityState)
	}

	// Convert plants
	for _, plant := range sm.world.AllPlants {
		plantState := &PlantState{
			ID:         plant.ID,
			Type:       plant.Type,
			Position:   plant.Position,
			Energy:     plant.Energy,
			Age:        plant.Age,
			Size:       plant.Size,
			Traits:     make(map[string]float64),
			Generation: plant.Generation,
			IsAlive:    plant.IsAlive,
			NutritionVal: plant.NutritionVal,
			Toxicity:   plant.Toxicity,
			GrowthRate: plant.GrowthRate,
		}

		// Copy plant traits
		if plant.Traits != nil {
			for traitName, trait := range plant.Traits {
				plantState.Traits[traitName] = trait.Value
			}
		}

		state.Plants = append(state.Plants, plantState)
	}

	// Convert events
	for _, event := range sm.world.Events {
		eventState := &WorldEventState{
			Name:           event.Name,
			Description:    event.Description,
			Duration:       event.Duration,
			GlobalMutation: event.GlobalMutation,
			GlobalDamage:   event.GlobalDamage,
			BiomeChanges:   make(map[string]BiomeType),
		}

		// Convert biome changes positions to string keys
		for pos, biome := range event.BiomeChanges {
			key := fmt.Sprintf("%f,%f", pos.X, pos.Y)
			eventState.BiomeChanges[key] = biome
		}

		state.Events = append(state.Events, eventState)
	}

	// Convert time system
	if sm.world.AdvancedTimeSystem != nil {
		state.Time = AdvancedTimeState{
			WorldTick:    sm.world.AdvancedTimeSystem.WorldTick,
			DayLength:    sm.world.AdvancedTimeSystem.DayLength,
			SeasonLength: sm.world.AdvancedTimeSystem.SeasonLength,
			TimeOfDay:    sm.world.AdvancedTimeSystem.TimeOfDay,
			Season:       sm.world.AdvancedTimeSystem.Season,
			DayNumber:    sm.world.AdvancedTimeSystem.DayNumber,
			SeasonDay:    sm.world.AdvancedTimeSystem.SeasonDay,
			Temperature:  sm.world.AdvancedTimeSystem.Temperature,
			Illumination: sm.world.AdvancedTimeSystem.Illumination,
			SeasonalMod:  sm.world.AdvancedTimeSystem.SeasonalMod,
		}
	}

	// Convert wind system
	if sm.world.WindSystem != nil {
		state.Wind = WindSystemState{
			BaseWindDirection:  sm.world.WindSystem.BaseWindDirection,
			BaseWindStrength:   sm.world.WindSystem.BaseWindStrength,
			TurbulenceLevel:    sm.world.WindSystem.TurbulenceLevel,
			SeasonalMultiplier: sm.world.WindSystem.SeasonalMultiplier,
			WeatherPattern:     sm.world.WindSystem.WeatherPattern,
		}
	}

	// Convert speciation system
	if sm.world.SpeciationSystem != nil {
		state.Species = SpeciationSystemState{
			Species:       make(map[string]*SpeciesState),
			NextSpeciesID: sm.world.SpeciationSystem.NextSpeciesID,
		}

		for id, species := range sm.world.SpeciationSystem.ActiveSpecies {
			key := fmt.Sprintf("%d", id)
			state.Species.Species[key] = &SpeciesState{
				ID:         species.ID,
				Name:       species.Name,
				ParentID:   species.ParentSpeciesID,
				CreatedAt:  species.FormationTick,
				ExtinctAt:  species.ExtinctionTick,
				IsExtinct:  species.IsExtinct,
				Population: len(species.Members),
				BaseTraits: make(map[string]float64),
			}

			// Copy founding traits as base traits
			for trait, value := range species.FoundingTraits {
				state.Species.Species[key].BaseTraits[trait] = value
			}
		}
	}

	// Convert plant network
	if sm.world.PlantNetworkSystem != nil {
		state.Network = PlantNetworkState{
			Connections:   make([]*NetworkConnectionState, 0),
			ActiveSignals: make([]*ChemicalSignalState, 0),
		}

		// Convert connections
		for _, conn := range sm.world.PlantNetworkSystem.Connections {
			connState := &NetworkConnectionState{
				Plant1ID:   conn.PlantA.ID,
				Plant2ID:   conn.PlantB.ID,
				Type:       conn.Type,
				Strength:   conn.Strength,
				Health:     conn.Health,
				Efficiency: conn.Efficiency,
				Age:        conn.Age,
			}
			state.Network.Connections = append(state.Network.Connections, connState)
		}

		// Convert signals
		for _, signal := range sm.world.PlantNetworkSystem.ChemicalSignals {
			signalState := &ChemicalSignalState{
				ID:        signal.ID,
				SourceID:  signal.Source.ID,
				Type:      signal.Type,
				Intensity: signal.Intensity,
				Age:       signal.Age,
				MaxAge:    signal.MaxAge,
				Visited:   make(map[int]bool),
				Message:   signal.Message,
				Metadata:  make(map[string]float64),
			}

			// Copy visited plants
			for plantID, visited := range signal.Visited {
				signalState.Visited[plantID] = visited
			}

			// Copy metadata
			for k, v := range signal.Metadata {
				signalState.Metadata[k] = v
			}

			state.Network.ActiveSignals = append(state.Network.ActiveSignals, signalState)
		}
	}

	return state, nil
}



// convertDNAToState converts DNA structure to serializable state
func (sm *StateManager) convertDNAToState(dna *DNAStrand) *DNAState {
	if dna == nil {
		return nil
	}

	state := &DNAState{
		EntityID:    dna.EntityID,
		Chromosomes: make([]ChromosomeState, len(dna.Chromosomes)),
		Mutations:   dna.Mutations,
		Generation:  dna.Generation,
	}

	for i, chromosome := range dna.Chromosomes {
		state.Chromosomes[i] = ChromosomeState{
			ID:    chromosome.ID,
			Genes: make([]GeneState, len(chromosome.Genes)),
		}

		for j, gene := range chromosome.Genes {
			// Convert nucleotides to strings
			sequence := make([]string, len(gene.Sequence))
			for k, nucleotide := range gene.Sequence {
				sequence[k] = string(nucleotide)
			}

			state.Chromosomes[i].Genes[j] = GeneState{
				Name:       gene.Name,
				Sequence:   sequence,
				Dominant:   gene.Dominant,
				Expression: gene.Expression,
			}
		}
	}

	return state
}

// convertCellularToState converts cellular structure to serializable state
func (sm *StateManager) convertCellularToState(organism *CellularOrganism) *CellularState {
	if organism == nil {
		return nil
	}

	state := &CellularState{
		EntityID:        organism.EntityID,
		ComplexityLevel: organism.ComplexityLevel,
		TotalEnergy:     organism.TotalEnergy,
		CellDivisions:   organism.CellDivisions,
		Generation:      organism.Generation,
		Cells:           make([]CellState, len(organism.Cells)),
		OrganSystems:    make(map[string][]int),
	}

	// Copy organ systems
	for system, cellIDs := range organism.OrganSystems {
		state.OrganSystems[system] = make([]int, len(cellIDs))
		copy(state.OrganSystems[system], cellIDs)
	}

	// Convert cells
	for i, cell := range organism.Cells {
		state.Cells[i] = CellState{
			ID:          cell.ID,
			Type:        cell.Type,
			Size:        cell.Size,
			Energy:      cell.Energy,
			Health:      cell.Health,
			Age:         cell.Age,
			Position:    cell.Position,
			Connections: make([]int, len(cell.Connections)),
			Activity:    cell.Activity,
			Specialized: cell.Specialized,
			Organelles:  make(map[string]OrganelleState),
		}

		// Copy connections
		copy(state.Cells[i].Connections, cell.Connections)

		// Convert DNA
		if cell.DNA != nil {
			state.Cells[i].DNA = sm.convertDNAToState(cell.DNA)
		}

		// Convert organelles
		for organelleType, organelle := range cell.Organelles {
			key := fmt.Sprintf("%d", int(organelleType)) // Convert enum to string key
			state.Cells[i].Organelles[key] = OrganelleState{
				Type:       organelle.Type,
				Count:      organelle.Count,
				Efficiency: organelle.Efficiency,
				Energy:     organelle.Energy,
			}
		}
	}

	return state
}

// restoreState restores the world from a serializable state
func (sm *StateManager) restoreState(state *SimulationState) error {
	// Basic state restoration
	sm.world.Tick = state.Tick
	sm.world.NextID = state.NextID
	sm.world.NextPlantID = state.NextPlantID
	sm.world.Config = state.Config

	// Clear existing data
	sm.world.AllEntities = make([]*Entity, 0)
	sm.world.AllPlants = make([]*Plant, 0)
	sm.world.Events = make([]*WorldEvent, 0)
	sm.world.Populations = make(map[string]*Population)

	// Restore biomes
	for y := 0; y < len(sm.world.Grid) && y < len(state.Biomes); y++ {
		for x := 0; x < len(sm.world.Grid[y]) && x < len(state.Biomes[y]); x++ {
			sm.world.Grid[y][x].Biome = state.Biomes[y][x]
			sm.world.Grid[y][x].Entities = make([]*Entity, 0)
			sm.world.Grid[y][x].Plants = make([]*Plant, 0)
		}
	}

	// Restore entities (simplified approach - recreate populations)
	populationGroups := make(map[string][]*Entity)
	for _, entityState := range state.Entities {
		entity := sm.restoreEntity(entityState)
		sm.world.AllEntities = append(sm.world.AllEntities, entity)
		
		// Group by species
		if populationGroups[entity.Species] == nil {
			populationGroups[entity.Species] = make([]*Entity, 0)
		}
		populationGroups[entity.Species] = append(populationGroups[entity.Species], entity)

		// Add to grid
		gridX := int(entity.Position.X * float64(sm.world.Config.GridWidth) / sm.world.Config.Width)
		gridY := int(entity.Position.Y * float64(sm.world.Config.GridHeight) / sm.world.Config.Height)
		if gridX >= 0 && gridX < sm.world.Config.GridWidth && gridY >= 0 && gridY < sm.world.Config.GridHeight {
			sm.world.Grid[gridY][gridX].Entities = append(sm.world.Grid[gridY][gridX].Entities, entity)
		}
	}
	
	// Recreate populations from grouped entities
	for species, entities := range populationGroups {
		if len(entities) > 0 {
			// Get trait names from first entity
			traitNames := make([]string, 0, len(entities[0].Traits))
			for name := range entities[0].Traits {
				traitNames = append(traitNames, name)
			}
			
			pop := NewPopulation(len(entities), traitNames, 0.1, 0.2)
			pop.Species = species
			
			// Copy entities to population
			for i, entity := range entities {
				if i < len(pop.Entities) {
					pop.Entities[i] = entity
				}
			}
			
			sm.world.Populations[species] = pop
		}
	}

	// Restore plants
	for _, plantState := range state.Plants {
		plant := sm.restorePlant(plantState)
		sm.world.AllPlants = append(sm.world.AllPlants, plant)

		// Add to grid
		gridX := int(plant.Position.X * float64(sm.world.Config.GridWidth) / sm.world.Config.Width)
		gridY := int(plant.Position.Y * float64(sm.world.Config.GridHeight) / sm.world.Config.Height)
		if gridX >= 0 && gridX < sm.world.Config.GridWidth && gridY >= 0 && gridY < sm.world.Config.GridHeight {
			sm.world.Grid[gridY][gridX].Plants = append(sm.world.Grid[gridY][gridX].Plants, plant)
		}
	}

	// Restore events
	for _, eventState := range state.Events {
		event := &WorldEvent{
			Name:           eventState.Name,
			Description:    eventState.Description,
			Duration:       eventState.Duration,
			GlobalMutation: eventState.GlobalMutation,
			GlobalDamage:   eventState.GlobalDamage,
			BiomeChanges:   make(map[Position]BiomeType),
		}

		// Convert string keys back to positions
		for key, biome := range eventState.BiomeChanges {
			var x, y float64
			_, _ = fmt.Sscanf(key, "%f,%f", &x, &y)
			event.BiomeChanges[Position{X: x, Y: y}] = biome
		}

		sm.world.Events = append(sm.world.Events, event)
	}

	// Restore time system
	if sm.world.AdvancedTimeSystem != nil {
		sm.world.AdvancedTimeSystem.WorldTick = state.Time.WorldTick
		sm.world.AdvancedTimeSystem.DayLength = state.Time.DayLength
		sm.world.AdvancedTimeSystem.SeasonLength = state.Time.SeasonLength
		sm.world.AdvancedTimeSystem.TimeOfDay = state.Time.TimeOfDay
		sm.world.AdvancedTimeSystem.Season = state.Time.Season
		sm.world.AdvancedTimeSystem.DayNumber = state.Time.DayNumber
		sm.world.AdvancedTimeSystem.SeasonDay = state.Time.SeasonDay
		sm.world.AdvancedTimeSystem.Temperature = state.Time.Temperature
		sm.world.AdvancedTimeSystem.Illumination = state.Time.Illumination
		sm.world.AdvancedTimeSystem.SeasonalMod = state.Time.SeasonalMod
	}

	// Restore wind system
	if sm.world.WindSystem != nil {
		sm.world.WindSystem.BaseWindDirection = state.Wind.BaseWindDirection
		sm.world.WindSystem.BaseWindStrength = state.Wind.BaseWindStrength
		sm.world.WindSystem.TurbulenceLevel = state.Wind.TurbulenceLevel
		sm.world.WindSystem.SeasonalMultiplier = state.Wind.SeasonalMultiplier
		sm.world.WindSystem.WeatherPattern = state.Wind.WeatherPattern
	}

	// Restore speciation system (simplified - skip for now to avoid complexity)
	if sm.world.SpeciationSystem != nil && len(state.Species.Species) > 0 {
		sm.world.SpeciationSystem.NextSpeciesID = state.Species.NextSpeciesID
		// TODO: Restore species data - complex due to plant references
	}

	return nil
}

// restoreEntity creates an entity from its serialized state
func (sm *StateManager) restoreEntity(state *EntityState) *Entity {
	entity := &Entity{
		ID:         state.ID,
		Species:    state.Species,
		Position:   state.Position,
		Traits:     make(map[string]Trait), // Changed to map[string]Trait
		Fitness:    state.Fitness,
		Energy:     state.Energy,
		Age:        state.Age,
		IsAlive:    true,
		Generation: 0, // Will be set later if needed
	}

	// Restore traits
	for traitName, value := range state.Traits {
		entity.Traits[traitName] = Trait{
			Name:  traitName,
			Value: value,
		}
	}

	// Restore DNA and Cellular data if present
	if state.DNA != nil && sm.world.CellularSystem != nil {
		// Restore the cellular organism and DNA
		dna := sm.restoreDNA(state.DNA)
		organism := sm.restoreCellular(state.Cellular, dna)
		
		if organism != nil {
			sm.world.CellularSystem.OrganismMap[entity.ID] = organism
		}
	}

	return entity
}

// restorePlant creates a plant from its serialized state
func (sm *StateManager) restorePlant(state *PlantState) *Plant {
	plant := &Plant{
		ID:         state.ID,
		Type:       state.Type,
		Position:   state.Position,
		Energy:     state.Energy,
		Age:        state.Age,
		Size:       state.Size,
		Generation: state.Generation,
		IsAlive:    state.IsAlive,
		NutritionVal: state.NutritionVal,
		Toxicity:   state.Toxicity,
		GrowthRate: state.GrowthRate,
		Traits:     make(map[string]Trait),
	}

	// Restore traits
	for traitName, value := range state.Traits {
		plant.Traits[traitName] = Trait{
			Name:  traitName,
			Value: value,
		}
	}

	return plant
}

// restoreDNA creates DNA from its serialized state
func (sm *StateManager) restoreDNA(state *DNAState) *DNAStrand {
	if state == nil {
		return nil
	}

	dna := &DNAStrand{
		EntityID:    state.EntityID,
		Chromosomes: make([]Chromosome, len(state.Chromosomes)),
		Mutations:   state.Mutations,
		Generation:  state.Generation,
	}

	for i, chromosomeState := range state.Chromosomes {
		dna.Chromosomes[i] = Chromosome{
			ID:    chromosomeState.ID,
			Genes: make([]Gene, len(chromosomeState.Genes)),
		}

		for j, geneState := range chromosomeState.Genes {
			// Convert strings back to nucleotides
			sequence := make([]Nucleotide, len(geneState.Sequence))
			for k, nucleotideStr := range geneState.Sequence {
				if len(nucleotideStr) > 0 {
					sequence[k] = Nucleotide(nucleotideStr[0])
				}
			}

			dna.Chromosomes[i].Genes[j] = Gene{
				Name:       geneState.Name,
				Sequence:   sequence,
				Dominant:   geneState.Dominant,
				Expression: geneState.Expression,
			}
		}
	}

	return dna
}

// restoreCellular creates cellular structure from its serialized state
func (sm *StateManager) restoreCellular(state *CellularState, dna *DNAStrand) *CellularOrganism {
	if state == nil {
		return nil
	}

	organism := &CellularOrganism{
		EntityID:        state.EntityID,
		Cells:           make([]*Cell, len(state.Cells)),
		ComplexityLevel: state.ComplexityLevel,
		TotalEnergy:     state.TotalEnergy,
		CellDivisions:   state.CellDivisions,
		Generation:      state.Generation,
		OrganSystems:    make(map[string][]int),
	}

	// Restore organ systems
	for system, cellIDs := range state.OrganSystems {
		organism.OrganSystems[system] = make([]int, len(cellIDs))
		copy(organism.OrganSystems[system], cellIDs)
	}

	// Restore cells
	for i, cellState := range state.Cells {
		cell := &Cell{
			ID:          cellState.ID,
			Type:        cellState.Type,
			Size:        cellState.Size,
			Energy:      cellState.Energy,
			Health:      cellState.Health,
			Age:         cellState.Age,
			Position:    cellState.Position,
			Connections: make([]int, len(cellState.Connections)),
			Activity:    cellState.Activity,
			Specialized: cellState.Specialized,
			Organelles:  make(map[OrganelleType]*Organelle),
		}

		// Restore connections
		copy(cell.Connections, cellState.Connections)

		// Restore DNA for the cell
		if cellState.DNA != nil {
			cell.DNA = sm.restoreDNA(cellState.DNA)
		} else if dna != nil {
			cell.DNA = dna // Use the provided DNA if no cell-specific DNA
		}

		// Restore organelles
		for key, organelleState := range cellState.Organelles {
			// Convert string key back to OrganelleType
			var organelleType OrganelleType
			_, _ = fmt.Sscanf(key, "%d", (*int)(&organelleType))

			cell.Organelles[organelleType] = &Organelle{
				Type:       organelleState.Type,
				Count:      organelleState.Count,
				Efficiency: organelleState.Efficiency,
				Energy:     organelleState.Energy,
			}
		}

		organism.Cells[i] = cell
	}

	return organism
}