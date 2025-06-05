package main

import (
	"fmt"
	"time"
)

// SimulationConfig holds all configuration for the simulation
type SimulationConfig struct {
	Time       TimeConfig               `json:"time"`
	Energy     EnergyConfig             `json:"energy"`
	Population PopulationConfigSettings `json:"population"`
	Physics    PhysicsConfig            `json:"physics"`
	World      WorldConfigSettings      `json:"world"`
	Evolution  EvolutionConfig          `json:"evolution"`
	Biomes     BiomesConfig             `json:"biomes"`
	Plants     PlantsConfig             `json:"plants"`
	Web        WebConfig                `json:"web"`
}

// TimeConfig holds all time-related configuration
type TimeConfig struct {
	TicksPerDay       int     `json:"ticks_per_day"`      // How many simulation ticks equal one day
	DaysPerSeason     int     `json:"days_per_season"`    // How many days in each season
	DailyEnergyBase   float64 `json:"daily_energy_base"`  // Base daily energy requirement
	NightPenalty      float64 `json:"night_penalty"`      // Energy penalty during night
	SeasonalVariation float64 `json:"seasonal_variation"` // How much seasons affect environment
}

// EnergyConfig holds all energy-related configuration
type EnergyConfig struct {
	BaseEnergyDrain        float64            `json:"base_energy_drain"`        // Base energy cost per tick
	MovementEnergyCost     float64            `json:"movement_energy_cost"`     // Energy cost for movement
	ReproductionCost       float64            `json:"reproduction_cost"`        // Energy cost for reproduction
	SurvivalThreshold      float64            `json:"survival_threshold"`       // Energy level below which entities die
	MaxEnergyLevel         float64            `json:"max_energy_level"`         // Maximum energy an entity can have
	EnergyRegenerationRate float64            `json:"energy_regeneration_rate"` // How fast energy naturally regenerates
	BiomeEnergyModifiers   map[string]float64 `json:"biome_energy_modifiers"`   // Energy drain per biome type
}

// PopulationConfigSettings holds population-related configuration
type PopulationConfigSettings struct {
	DefaultPopSize    int     `json:"default_pop_size"`    // Default population size per species
	MaxPopulation     int     `json:"max_population"`      // Maximum total population
	MutationRateBase  float64 `json:"mutation_rate_base"`  // Base mutation rate
	MutationRateRange float64 `json:"mutation_rate_range"` // Range of mutation rate variation
	SelectionPressure float64 `json:"selection_pressure"`  // How strong natural selection is
	CarryingCapacity  int     `json:"carrying_capacity"`   // Maximum entities per area unit
}

// PhysicsConfig holds physics-related configuration
type PhysicsConfig struct {
	CollisionDetection bool    `json:"collision_detection"` // Whether to enable collision detection
	MaxVelocity        float64 `json:"max_velocity"`        // Maximum entity velocity
	Friction           float64 `json:"friction"`            // Friction coefficient
	WindStrength       float64 `json:"wind_strength"`       // Base wind strength
	GravityStrength    float64 `json:"gravity_strength"`    // Gravity effect strength
}

// WorldConfigSettings holds world generation configuration
type WorldConfigSettings struct {
	Width          float64 `json:"width"`           // World width
	Height         float64 `json:"height"`          // World height
	GridWidth      int     `json:"grid_width"`      // Grid cells width for visualization
	GridHeight     int     `json:"grid_height"`     // Grid cells height for visualization
	BiomeVariety   float64 `json:"biome_variety"`   // How diverse biomes should be
	EventFrequency float64 `json:"event_frequency"` // How often world events occur
}

// EvolutionConfig holds evolution-related configuration
type EvolutionConfig struct {
	TraitMutationStrength float64               `json:"trait_mutation_strength"` // How much traits can change
	TraitBounds           map[string][2]float64 `json:"trait_bounds"`            // Min/max values for each trait
	FitnessWeights        map[string]float64    `json:"fitness_weights"`         // Weight of each factor in fitness
	SpeciationThreshold   float64               `json:"speciation_threshold"`    // Genetic distance for new species
}

// BiomesConfig holds biome-related configuration
type BiomesConfig struct {
	EnergyDrainMultipliers map[string]float64    `json:"energy_drain_multipliers"` // Energy drain by biome
	MutationRateModifiers  map[string]float64    `json:"mutation_rate_modifiers"`  // Mutation rate by biome
	TemperatureRanges      map[string][2]float64 `json:"temperature_ranges"`       // Temperature ranges by biome
	CarryingCapacities     map[string]int        `json:"carrying_capacities"`      // Max entities per biome type
}

// PlantsConfig holds plant-related configuration
type PlantsConfig struct {
	GrowthRate          float64            `json:"growth_rate"`          // How fast plants grow
	MaxAge              int                `json:"max_age"`              // Maximum plant age
	SeedProductionRate  float64            `json:"seed_production_rate"` // Rate of seed production
	PollinationRadius   float64            `json:"pollination_radius"`   // Range for pollination
	NutrientRequirement map[string]float64 `json:"nutrient_requirement"` // Nutrient needs by plant type
}

// WebConfig holds web interface configuration
type WebConfig struct {
	UpdateInterval time.Duration `json:"update_interval"` // How often to update web interface
	Port           int           `json:"port"`            // Web server port
	MaxClients     int           `json:"max_clients"`     // Maximum concurrent clients
}

// DefaultSimulationConfig returns a default configuration
func DefaultSimulationConfig() *SimulationConfig {
	return &SimulationConfig{
		Time: TimeConfig{
			TicksPerDay:       1,    // 1 tick = 1 day (realistic evolutionary time scale)
			DaysPerSeason:     91,   // ~3 months per season
			DailyEnergyBase:   0.02, // Base daily energy requirement
			NightPenalty:      0.01, // Additional energy cost at night
			SeasonalVariation: 0.3,  // 30% variation between seasons
		},
		Energy: EnergyConfig{
			BaseEnergyDrain:        0.01,  // Base energy cost per tick
			MovementEnergyCost:     0.005, // Energy cost for movement
			ReproductionCost:       20.0,  // Energy cost for reproduction
			SurvivalThreshold:      10.0,  // Die if energy < 10
			MaxEnergyLevel:         100.0, // Maximum energy
			EnergyRegenerationRate: 0.1,   // Natural energy gain per tick
			BiomeEnergyModifiers: map[string]float64{
				"plains":        0.5, // Low energy drain
				"forest":        0.8, // Moderate energy drain
				"desert":        1.5, // High energy drain
				"mountain":      1.2, // High energy drain
				"water":         0.3, // Low energy drain for aquatic
				"radiation":     2.0, // Very high energy drain
				"soil":          0.7, // Moderate energy drain
				"air":           1.0, // Standard energy drain
				"ice":           2.5, // Very high energy drain
				"rainforest":    0.3, // Low energy drain (abundant resources)
				"deep_water":    1.8, // High energy drain
				"high_altitude": 3.0, // Very high energy drain
				"hot_spring":    0.8, // Moderate energy drain
				"tundra":        1.8, // High energy drain
				"swamp":         1.2, // Moderate energy drain
				"canyon":        1.5, // High energy drain
			},
		},
		Population: PopulationConfigSettings{
			DefaultPopSize:    20,   // Default 20 entities per species
			MaxPopulation:     1000, // Maximum 1000 entities total
			MutationRateBase:  0.1,  // 10% base mutation rate
			MutationRateRange: 0.15, // ±15% variation in mutation rate
			SelectionPressure: 0.3,  // Moderate selection pressure
			CarryingCapacity:  50,   // 50 entities per area unit
		},
		Physics: PhysicsConfig{
			CollisionDetection: true,
			MaxVelocity:        10.0,
			Friction:           0.1,
			WindStrength:       5.0,
			GravityStrength:    1.0,
		},
		World: WorldConfigSettings{
			Width:          100.0,
			Height:         100.0,
			GridWidth:      40,
			GridHeight:     25,
			BiomeVariety:   0.7, // 70% biome diversity
			EventFrequency: 0.1, // 10% chance of events
		},
		Evolution: EvolutionConfig{
			TraitMutationStrength: 0.1, // 10% trait change per mutation
			TraitBounds: map[string][2]float64{
				"size":                 {-2.0, 2.0},
				"speed":                {-1.0, 2.0},
				"aggression":           {-1.0, 1.0},
				"defense":              {-1.0, 2.0},
				"cooperation":          {-1.0, 1.0},
				"intelligence":         {-1.0, 2.0},
				"endurance":            {-1.0, 2.0},
				"strength":             {-1.0, 2.0},
				"aquatic_adaptation":   {-1.0, 1.0},
				"digging_ability":      {-1.0, 1.0},
				"underground_nav":      {-1.0, 1.0},
				"flying_ability":       {-1.0, 1.0},
				"altitude_tolerance":   {-1.0, 1.0},
				"circadian_preference": {-1.0, 1.0},
				"sleep_need":           {0.0, 1.0},
				"hunger_need":          {0.0, 1.0},
				"thirst_need":          {0.0, 1.0},
				"play_drive":           {-1.0, 1.0},
				"exploration_drive":    {0.0, 1.0},
				"scavenging_behavior":  {0.0, 1.0},
			},
			FitnessWeights: map[string]float64{
				"survival":     1.0,
				"reproduction": 0.8,
				"cooperation":  0.3,
				"exploration":  0.2,
			},
			SpeciationThreshold: 0.5, // 50% genetic difference for new species
		},
		Biomes: BiomesConfig{
			EnergyDrainMultipliers: map[string]float64{
				"plains":        1.0,
				"forest":        1.6,
				"desert":        3.0,
				"mountain":      2.4,
				"water":         0.6,
				"radiation":     4.0,
				"soil":          1.4,
				"air":           2.0,
				"ice":           5.0,
				"rainforest":    0.6,
				"deep_water":    3.6,
				"high_altitude": 6.0,
				"hot_spring":    1.6,
				"tundra":        3.6,
				"swamp":         2.4,
				"canyon":        3.0,
			},
			MutationRateModifiers: map[string]float64{
				"radiation":     2.0, // 2x mutation rate in radiation
				"hot_spring":    1.5, // 1.5x mutation rate
				"desert":        1.2, // 1.2x mutation rate
				"high_altitude": 1.3, // 1.3x mutation rate
			},
			TemperatureRanges: map[string][2]float64{
				"ice":        {-1.0, -0.5},
				"tundra":     {-0.5, 0.0},
				"mountain":   {-0.2, 0.3},
				"plains":     {0.0, 0.7},
				"forest":     {0.2, 0.8},
				"rainforest": {0.6, 0.9},
				"desert":     {0.7, 1.0},
				"hot_spring": {0.8, 1.0},
			},
			CarryingCapacities: map[string]int{
				"plains":        60,
				"forest":        50,
				"rainforest":    40,
				"water":         30,
				"desert":        15,
				"mountain":      25,
				"radiation":     5,
				"soil":          35,
				"air":           20,
				"ice":           10,
				"deep_water":    20,
				"high_altitude": 10,
				"hot_spring":    25,
				"tundra":        15,
				"swamp":         35,
				"canyon":        20,
			},
		},
		Plants: PlantsConfig{
			GrowthRate:         0.1,
			MaxAge:             200,
			SeedProductionRate: 0.05,
			PollinationRadius:  5.0,
			NutrientRequirement: map[string]float64{
				"grass":     1.0,
				"flowers":   1.5,
				"shrubs":    2.0,
				"trees":     3.0,
				"mushrooms": 0.8,
				"aquatic":   1.2,
			},
		},
		Web: WebConfig{
			UpdateInterval: 100 * time.Millisecond,
			Port:           8080,
			MaxClients:     100,
		},
	}
}

// ApplySpeedMultiplier adjusts time-based configuration values based on speed multiplier
func (config *SimulationConfig) ApplySpeedMultiplier(speedMultiplier float64) *SimulationConfig {
	// Create a copy of the config to avoid modifying the original
	speedConfig := *config

	// Adjust time-based values - with speed multiplier, things happen faster
	// But we keep the relative proportions the same
	speedConfig.Energy.BaseEnergyDrain *= speedMultiplier
	speedConfig.Energy.MovementEnergyCost *= speedMultiplier
	speedConfig.Energy.EnergyRegenerationRate *= speedMultiplier
	speedConfig.Time.DailyEnergyBase *= speedMultiplier
	speedConfig.Time.NightPenalty *= speedMultiplier

	// Plant growth should scale with speed
	speedConfig.Plants.GrowthRate *= speedMultiplier
	speedConfig.Plants.SeedProductionRate *= speedMultiplier

	// Web update interval should decrease with higher speed (more frequent updates)
	if speedMultiplier > 1.0 {
		speedConfig.Web.UpdateInterval = time.Duration(float64(config.Web.UpdateInterval) / speedMultiplier)
	}

	return &speedConfig
}

// ValidateConfig ensures all configuration values are within reasonable bounds
func (config *SimulationConfig) Validate() error {
	if config.Time.TicksPerDay <= 0 {
		return fmt.Errorf("ticks per day must be positive")
	}
	if config.Time.DaysPerSeason <= 0 {
		return fmt.Errorf("days per season must be positive")
	}
	if config.Energy.SurvivalThreshold >= config.Energy.MaxEnergyLevel {
		return fmt.Errorf("survival threshold must be less than max energy level")
	}
	if config.Population.DefaultPopSize <= 0 {
		return fmt.Errorf("default population size must be positive")
	}
	if config.World.Width <= 0 || config.World.Height <= 0 {
		return fmt.Errorf("world dimensions must be positive")
	}
	if config.World.GridWidth <= 0 || config.World.GridHeight <= 0 {
		return fmt.Errorf("grid dimensions must be positive")
	}
	return nil
}

// GetBiomeEnergyDrain returns the energy drain for a specific biome
func (config *SimulationConfig) GetBiomeEnergyDrain(biomeType BiomeType) float64 {
	biomeNames := map[BiomeType]string{
		BiomePlains:       "plains",
		BiomeForest:       "forest",
		BiomeDesert:       "desert",
		BiomeMountain:     "mountain",
		BiomeWater:        "water",
		BiomeRadiation:    "radiation",
		BiomeSoil:         "soil",
		BiomeAir:          "air",
		BiomeIce:          "ice",
		BiomeRainforest:   "rainforest",
		BiomeDeepWater:    "deep_water",
		BiomeHighAltitude: "high_altitude",
		BiomeHotSpring:    "hot_spring",
		BiomeTundra:       "tundra",
		BiomeSwamp:        "swamp",
		BiomeCanyon:       "canyon",
	}

	biomeName, exists := biomeNames[biomeType]
	if !exists {
		return config.Energy.BaseEnergyDrain // Default to base energy drain
	}

	multiplier, exists := config.Energy.BiomeEnergyModifiers[biomeName]
	if !exists {
		return config.Energy.BaseEnergyDrain // Default to base energy drain
	}

	return config.Energy.BaseEnergyDrain * multiplier
}

// GetBiomeMutationModifier returns the mutation rate modifier for a specific biome
func (config *SimulationConfig) GetBiomeMutationModifier(biomeType BiomeType) float64 {
	biomeNames := map[BiomeType]string{
		BiomePlains:       "plains",
		BiomeForest:       "forest",
		BiomeDesert:       "desert",
		BiomeMountain:     "mountain",
		BiomeWater:        "water",
		BiomeRadiation:    "radiation",
		BiomeSoil:         "soil",
		BiomeAir:          "air",
		BiomeIce:          "ice",
		BiomeRainforest:   "rainforest",
		BiomeDeepWater:    "deep_water",
		BiomeHighAltitude: "high_altitude",
		BiomeHotSpring:    "hot_spring",
		BiomeTundra:       "tundra",
		BiomeSwamp:        "swamp",
		BiomeCanyon:       "canyon",
	}

	biomeName, exists := biomeNames[biomeType]
	if !exists {
		return 1.0 // Default multiplier
	}

	modifier, exists := config.Biomes.MutationRateModifiers[biomeName]
	if !exists {
		return 1.0 // Default multiplier
	}

	return modifier
}
