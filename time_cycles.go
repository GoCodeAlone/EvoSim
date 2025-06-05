package main

import (
	"math"
)

// TimeOfDay represents different periods of the day
type TimeOfDay int

const (
	Dawn TimeOfDay = iota
	Morning
	Midday
	Afternoon
	Evening
	Night
	Midnight
	LateNight
)

// Season represents different seasons
type Season int

const (
	Spring Season = iota
	Summer
	Autumn
	Winter
)

// TimeState represents the current time state of the world
type TimeState struct {
	TimeOfDay    TimeOfDay
	Season       Season
	Temperature  float64
	Illumination float64
	SeasonalMod  float64
}

// IsNight returns true if it's currently night time
func (ts TimeState) IsNight() bool {
	return ts.TimeOfDay == Night || ts.TimeOfDay == Midnight || ts.TimeOfDay == LateNight
}

// AdvancedTimeSystem manages complex time cycles
type AdvancedTimeSystem struct {
	Config       *TimeConfig // Configuration for time system
	WorldTick    int
	DayLength    int // Ticks per day (from config)
	SeasonLength int // Days per season (from config)
	TimeOfDay    TimeOfDay
	Season       Season
	DayNumber    int
	SeasonDay    int
	Temperature  float64 // Current temperature (affects all entities)
	Illumination float64 // Light level (0.0 to 1.0)
	SeasonalMod  float64 // Seasonal modifier for resources/difficulty
}

// NewAdvancedTimeSystem creates a new time system with configuration
func NewAdvancedTimeSystem(config *TimeConfig) *AdvancedTimeSystem {
	return &AdvancedTimeSystem{
		Config:       config,
		DayLength:    config.TicksPerDay,
		SeasonLength: config.DaysPerSeason,
		TimeOfDay:    Dawn,
		Season:       Spring,
		Temperature:  0.5, // Moderate starting temperature
		Illumination: 0.6, // Dawn lighting
		SeasonalMod:  1.0,
	}
}

// NewAdvancedTimeSystemLegacy creates a time system with legacy parameters (for compatibility)
func NewAdvancedTimeSystemLegacy(dayLength, seasonLength int) *AdvancedTimeSystem {
	config := &TimeConfig{
		TicksPerDay:       dayLength,
		DaysPerSeason:     seasonLength,
		DailyEnergyBase:   0.02,
		NightPenalty:      0.01,
		SeasonalVariation: 0.3,
	}
	return NewAdvancedTimeSystem(config)
}

// Update advances the time system and calculates environmental effects
func (ats *AdvancedTimeSystem) Update() {
	ats.WorldTick++

	// With daily time scale, cycle through time of day within each day
	// Use a simple 8-period day cycle for environmental variation
	timeOfDayIndex := ats.WorldTick % 8
	ats.updateTimeOfDayFromIndex(timeOfDayIndex)

	// Update day number (since each tick = 1 day, increment every tick)
	ats.DayNumber++
	ats.SeasonDay++

	// Check for season change
	if ats.SeasonDay >= ats.SeasonLength {
		ats.SeasonDay = 0
		ats.Season = Season((int(ats.Season) + 1) % 4)
	}

	// Update environmental factors based on season and time of day
	ats.updateEnvironmentalFactors()
}

// GetTimeState returns the current time state
func (ats *AdvancedTimeSystem) GetTimeState() TimeState {
	return TimeState{
		TimeOfDay:    ats.TimeOfDay,
		Season:       ats.Season,
		Temperature:  ats.Temperature,
		Illumination: ats.Illumination,
		SeasonalMod:  ats.SeasonalMod,
	}
}

// updateTimeOfDayFromIndex sets time of day based on daily cycle
func (ats *AdvancedTimeSystem) updateTimeOfDayFromIndex(index int) {
	timeOfDays := []TimeOfDay{
		Dawn,      // 0
		Morning,   // 1
		Midday,    // 2
		Afternoon, // 3
		Evening,   // 4
		Night,     // 5
		Midnight,  // 6
		LateNight, // 7
	}
	ats.TimeOfDay = timeOfDays[index]
}

// updateEnvironmentalFactors calculates temperature, light, and seasonal effects
func (ats *AdvancedTimeSystem) updateEnvironmentalFactors() {
	// Calculate illumination based on time of day (simplified for daily cycle)
	switch ats.TimeOfDay {
	case Dawn:
		ats.Illumination = 0.4
	case Morning:
		ats.Illumination = 0.8
	case Midday:
		ats.Illumination = 1.0
	case Afternoon:
		ats.Illumination = 0.8
	case Evening:
		ats.Illumination = 0.5
	case Night, Midnight, LateNight:
		ats.Illumination = 0.1
	}

	// Calculate temperature based on time of day and season
	baseTemp := ats.getSeasonalBaseTemperature()

	// Daily temperature variation (warmer during day, cooler at night)
	dailyVariation := 0.0
	switch ats.TimeOfDay {
	case Dawn:
		dailyVariation = -0.2
	case Morning:
		dailyVariation = 0.1
	case Midday:
		dailyVariation = 0.3
	case Afternoon:
		dailyVariation = 0.2
	case Evening:
		dailyVariation = 0.0
	case Night, Midnight, LateNight:
		dailyVariation = -0.3
	}

	ats.Temperature = baseTemp + dailyVariation

	// Calculate seasonal modifier
	ats.SeasonalMod = ats.getSeasonalModifier()
}

// getSeasonalBaseTemperature returns base temperature for current season
func (ats *AdvancedTimeSystem) getSeasonalBaseTemperature() float64 {
	switch ats.Season {
	case Spring:
		return 0.6
	case Summer:
		return 0.9
	case Autumn:
		return 0.5
	case Winter:
		return 0.2
	default:
		return 0.5
	}
}

// getSeasonalModifier returns resource/difficulty modifier for current season
func (ats *AdvancedTimeSystem) getSeasonalModifier() float64 {
	// Use config's seasonal variation to determine how much seasons affect the environment
	baseModifier := 1.0
	variation := ats.Config.SeasonalVariation

	switch ats.Season {
	case Spring:
		return baseModifier + (0.2 * variation) // Slightly abundant resources
	case Summer:
		return baseModifier // Normal resources
	case Autumn:
		return baseModifier - (0.1 * variation) // Slightly declining resources
	case Winter:
		return baseModifier - (0.4 * variation) // Scarce resources
	default:
		return baseModifier
	}
}

// GetTimeDescription returns a human-readable time description
func (ats *AdvancedTimeSystem) GetTimeDescription() string {
	timeNames := map[TimeOfDay]string{
		Dawn:      "Dawn",
		Morning:   "Morning",
		Midday:    "Midday",
		Afternoon: "Afternoon",
		Evening:   "Evening",
		Night:     "Night",
		Midnight:  "Midnight",
		LateNight: "Late Night",
	}

	seasonNames := map[Season]string{
		Spring: "Spring",
		Summer: "Summer",
		Autumn: "Autumn",
		Winter: "Winter",
	}

	return timeNames[ats.TimeOfDay] + " of " + seasonNames[ats.Season]
}

// CircadianPreferences represents entity preferences for different times
type CircadianPreferences struct {
	PreferredTime      TimeOfDay             // Most active time
	ActivityModifier   map[TimeOfDay]float64 // Multiplier for different times
	SeasonalAdaptation map[Season]float64    // How well adapted to each season
}

// NewCircadianPreferences creates preferences based on entity species and traits
func NewCircadianPreferences(entity *Entity) *CircadianPreferences {
	preferences := &CircadianPreferences{
		ActivityModifier:   make(map[TimeOfDay]float64),
		SeasonalAdaptation: make(map[Season]float64),
	}

	intelligence := entity.GetTrait("intelligence")
	aggression := entity.GetTrait("aggression")
	endurance := entity.GetTrait("endurance")

	// Determine preferred time based on species and traits
	switch entity.Species {
	case "predator":
		if aggression > 0.6 {
			preferences.PreferredTime = Night // Nocturnal hunters
		} else {
			preferences.PreferredTime = Evening // Crepuscular
		}
	case "herbivore":
		preferences.PreferredTime = Morning // Early grazers
	case "omnivore":
		if intelligence > 0.5 {
			preferences.PreferredTime = Midday // Adaptable, active when convenient
		} else {
			preferences.PreferredTime = Afternoon
		}
	}

	// Set activity modifiers
	for timeOfDay := Dawn; timeOfDay <= LateNight; timeOfDay++ {
		if timeOfDay == preferences.PreferredTime {
			preferences.ActivityModifier[timeOfDay] = 1.3 // 30% bonus during preferred time
		} else {
			// Calculate distance from preferred time
			distance := int(math.Abs(float64(timeOfDay - preferences.PreferredTime)))
			if distance > 4 {
				distance = 8 - distance // Wrap around for circular time
			}
			modifier := 1.0 - float64(distance)*0.1 // Decrease by 10% per time period away
			preferences.ActivityModifier[timeOfDay] = math.Max(0.5, modifier)
		}
	}

	// Set seasonal adaptations based on traits
	for season := Spring; season <= Winter; season++ {
		adaptation := 0.8 + endurance*0.4 // Base adaptation + endurance bonus
		switch season {
		case Winter:
			// Winter requires more endurance
			adaptation -= (1.0 - endurance) * 0.3
		case Summer:
			// Summer heat tolerance
			adaptation += intelligence * 0.1 // Smart entities adapt better
		}
		preferences.SeasonalAdaptation[season] = math.Max(0.3, math.Min(1.5, adaptation))
	}

	return preferences
}

// GetCurrentActivityLevel returns entity's current activity multiplier
func (cp *CircadianPreferences) GetCurrentActivityLevel(timeSystem *AdvancedTimeSystem) float64 {
	timeModifier := cp.ActivityModifier[timeSystem.TimeOfDay]
	seasonalModifier := cp.SeasonalAdaptation[timeSystem.Season]

	// Light dependency - some entities need light to be fully active
	lightDependency := 0.2 // How much light affects activity
	lightModifier := 1.0 - lightDependency + lightDependency*timeSystem.Illumination

	return timeModifier * seasonalModifier * lightModifier
}

// TemperatureEffect calculates how temperature affects an entity
func (cp *CircadianPreferences) TemperatureEffect(entity *Entity, temperature float64) float64 {
	endurance := entity.GetTrait("endurance")
	size := entity.GetTrait("size")

	// Larger entities handle temperature extremes better
	temperatureTolerance := 0.3 + endurance*0.4 + size*0.2

	// Optimal temperature range
	optimalTemp := 0.5 // Moderate temperature
	tempDifference := math.Abs(temperature - optimalTemp)

	if tempDifference <= temperatureTolerance {
		return 1.0 // No penalty in comfortable range
	}

	// Linear penalty outside comfort zone
	penalty := (tempDifference - temperatureTolerance) * 2.0
	return math.Max(0.2, 1.0-penalty) // Minimum 20% efficiency
}

// MigrationBehavior represents seasonal migration patterns
type MigrationBehavior struct {
	IsMigratory     bool
	PreferredBiomes map[Season]BiomeType
	MigrationRange  float64 // How far entities will travel
	GroupMigration  bool    // Whether they migrate in groups
}

// NewMigrationBehavior creates migration behavior based on entity traits
func NewMigrationBehavior(entity *Entity) *MigrationBehavior {
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")
	endurance := entity.GetTrait("endurance")

	// Only intelligent entities with good endurance migrate
	isMigratory := intelligence > 0.4 && endurance > 0.3

	behavior := &MigrationBehavior{
		IsMigratory:     isMigratory,
		PreferredBiomes: make(map[Season]BiomeType),
		MigrationRange:  20.0 + intelligence*30.0, // Smarter entities travel farther
		GroupMigration:  cooperation > 0.5,
	}

	if isMigratory {
		// Set seasonal biome preferences
		switch entity.Species {
		case "herbivore":
			behavior.PreferredBiomes[Spring] = BiomePlains
			behavior.PreferredBiomes[Summer] = BiomeForest
			behavior.PreferredBiomes[Autumn] = BiomePlains
			behavior.PreferredBiomes[Winter] = BiomeDesert // Warmer
		case "predator":
			behavior.PreferredBiomes[Spring] = BiomeForest
			behavior.PreferredBiomes[Summer] = BiomeMountain // Higher altitude, cooler
			behavior.PreferredBiomes[Autumn] = BiomeForest
			behavior.PreferredBiomes[Winter] = BiomePlains // Follow prey
		case "omnivore":
			// Omnivores are adaptable, prefer moderate biomes
			behavior.PreferredBiomes[Spring] = BiomePlains
			behavior.PreferredBiomes[Summer] = BiomePlains
			behavior.PreferredBiomes[Autumn] = BiomePlains
			behavior.PreferredBiomes[Winter] = BiomeForest // Shelter
		}
	}

	return behavior
}

// ShouldMigrate determines if an entity should start migrating
func (mb *MigrationBehavior) ShouldMigrate(entity *Entity, currentBiome BiomeType, season Season) bool {
	if !mb.IsMigratory {
		return false
	}

	preferredBiome := mb.PreferredBiomes[season]
	return currentBiome != preferredBiome
}

// GetMigrationTarget returns the target biome position for migration
func (mb *MigrationBehavior) GetMigrationTarget(entity *Entity, world *World, season Season) Position {
	preferredBiome := mb.PreferredBiomes[season]

	// Find nearest cell with preferred biome within migration range
	bestDistance := mb.MigrationRange
	bestPos := entity.Position

	for y := 0; y < world.Config.GridHeight; y++ {
		for x := 0; x < world.Config.GridWidth; x++ {
			if world.Grid[y][x].Biome == preferredBiome {
				worldX := (float64(x) + 0.5) * (world.Config.Width / float64(world.Config.GridWidth))
				worldY := (float64(y) + 0.5) * (world.Config.Height / float64(world.Config.GridHeight))

				distance := math.Sqrt(math.Pow(entity.Position.X-worldX, 2) + math.Pow(entity.Position.Y-worldY, 2))

				if distance < bestDistance {
					bestDistance = distance
					bestPos = Position{X: worldX, Y: worldY}
				}
			}
		}
	}

	return bestPos
}
