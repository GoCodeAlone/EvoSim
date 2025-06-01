package main

import (
	"math"
	"math/rand"
)

// WindVector represents wind direction and strength
type WindVector struct {
	X, Y       float64 // Direction and magnitude
	Strength   float64 // Wind strength (0.0 to 1.0)
	Turbulence float64 // Chaotic variations
}

// PollenGrain represents individual pollen particles
type PollenGrain struct {
	ID        int
	Position  Position
	Velocity  Vector2D
	PlantID   int // Source plant ID
	PlantType PlantType
	Genetics  map[string]Trait // Genetic material from parent
	Viability float64          // How long pollen remains viable (0.0 to 1.0)
	Age       int              // Ticks since release
	MaxAge    int              // Maximum viability age
	Size      float64          // Affects wind resistance
}

// PollenCloud represents a collection of pollen in an area
type PollenCloud struct {
	ID         int
	Position   Position
	Radius     float64
	Density    float64 // Number of pollen grains per unit area
	Grains     []*PollenGrain
	PlantTypes map[PlantType]int // Count of each plant type's pollen
	Age        int
	MaxAge     int
}

// RegionalStorm represents localized weather events
type RegionalStorm struct {
	ID          int
	Type        StormType  // Type of storm
	Center      Position   // Storm center
	Radius      float64    // Affected area
	Intensity   float64    // Storm intensity (0.0 to 1.0)
	Duration    int        // Remaining duration
	MaxDuration int        // Total duration
	MovementDir float64    // Direction storm is moving
	Speed       float64    // How fast storm moves
}

// StormType represents different types of regional storms
type StormType int

const (
	StormThunderstorm StormType = iota
	StormTornado
	StormHurricane
	StormBlizzard
	StormDustStorm
)

// WindSystem manages atmospheric conditions and pollen dispersal
type WindSystem struct {
	// Wind pattern generation
	BaseWindDirection float64        // Global wind direction (radians)
	BaseWindStrength  float64        // Global wind strength
	TurbulenceLevel   float64        // Chaos in wind patterns
	WindMap           [][]WindVector // 2D grid of wind vectors
	MapWidth          int
	MapHeight         int
	CellSize          float64 // Size of each wind cell

	// Pollen management
	AllPollenGrains []*PollenGrain
	PollenClouds    []*PollenCloud
	NextPollenID    int
	NextCloudID     int

	// Seasonal effects
	SeasonalMultiplier float64 // Wind strength varies by season
	WeatherPattern     int     // Current weather (0=calm, 1=windy, 2=storm, 3=tornado, 4=hurricane)
	WeatherDuration    int     // How long current weather lasts
	
	// Regional weather systems
	RegionalStorms []RegionalStorm // Active regional weather events

	// Statistics
	TotalPollenReleased            int
	SuccessfulPollinationsThisTick int
	TotalCrossPollinations         int
}

// NewWindSystem creates a new wind and pollen dispersal system
func NewWindSystem(worldWidth, worldHeight int) *WindSystem {
	cellSize := 10.0 // Each wind cell covers 10x10 world units
	mapWidth := int(math.Ceil(float64(worldWidth) / cellSize))
	mapHeight := int(math.Ceil(float64(worldHeight) / cellSize))

	ws := &WindSystem{
		BaseWindDirection:  rand.Float64() * 2 * math.Pi, // Random initial direction
		BaseWindStrength:   0.3 + rand.Float64()*0.4,     // 0.3-0.7 strength
		TurbulenceLevel:    0.2,
		MapWidth:           mapWidth,
		MapHeight:          mapHeight,
		CellSize:           cellSize,
		AllPollenGrains:    make([]*PollenGrain, 0),
		PollenClouds:       make([]*PollenCloud, 0),
		NextPollenID:       1,
		NextCloudID:        1,
		SeasonalMultiplier: 1.0,
		WeatherPattern:     0, // Start calm
		WeatherDuration:    100 + rand.Intn(200),
		RegionalStorms:     make([]RegionalStorm, 0),
	}

	// Initialize wind map
	ws.WindMap = make([][]WindVector, mapHeight)
	for y := 0; y < mapHeight; y++ {
		ws.WindMap[y] = make([]WindVector, mapWidth)
	}

	ws.generateWindPattern()
	return ws
}

// Update advances the wind system by one tick
func (ws *WindSystem) Update(season Season, tick int) {
	// Update seasonal effects
	ws.updateSeasonalEffects(season)

	// Update weather patterns
	ws.updateWeatherPattern()
	
	// Update regional storms
	ws.updateRegionalStorms()
	
	// Potentially spawn new regional storms
	if rand.Float64() < 0.01 { // 1% chance per tick
		ws.spawnRegionalStorm()
	}

	// Regenerate wind pattern occasionally
	if tick%50 == 0 {
		ws.generateWindPattern()
	}

	// Update pollen grains
	ws.updatePollenGrains()

	// Update pollen clouds
	ws.updatePollenClouds()

	// Reset tick counters
	ws.SuccessfulPollinationsThisTick = 0
}

// generateWindPattern creates a new wind field
func (ws *WindSystem) generateWindPattern() {
	for y := 0; y < ws.MapHeight; y++ {
		for x := 0; x < ws.MapWidth; x++ {
			// Base wind with some spatial variation
			direction := ws.BaseWindDirection +
				(rand.Float64()-0.5)*ws.TurbulenceLevel*math.Pi

			strength := ws.BaseWindStrength * ws.SeasonalMultiplier

			// Add terrain effects (simplified)
			terrainEffect := 1.0 + (rand.Float64()-0.5)*0.3
			strength *= terrainEffect

			// Weather pattern effects
			switch ws.WeatherPattern {
			case 1: // Windy
				strength *= 1.5
			case 2: // Storm
				strength *= 2.5
				direction += (rand.Float64() - 0.5) * 0.5 // More turbulent
			case 3: // Tornado (regional, handled separately)
				strength *= 1.2
			case 4: // Hurricane (regional, handled separately)
				strength *= 1.8
			}
			
			// Apply regional storm effects
			cellPos := Position{X: float64(x) * ws.CellSize, Y: float64(y) * ws.CellSize}
			stormEffect := ws.getRegionalStormEffect(cellPos)
			strength *= stormEffect.Strength
			direction += stormEffect.DirectionChange
			turbulence := ws.TurbulenceLevel * float64(ws.WeatherPattern+1) * stormEffect.TurbulenceMultiplier

			// Convert to vector components
			windX := math.Cos(direction) * strength
			windY := math.Sin(direction) * strength

			ws.WindMap[y][x] = WindVector{
				X:          windX,
				Y:          windY,
				Strength:   strength,
				Turbulence: turbulence,
			}
		}
	}
}

// GetWindAt returns the wind vector at a specific world position
func (ws *WindSystem) GetWindAt(pos Position) WindVector {
	cellX := int(pos.X / ws.CellSize)
	cellY := int(pos.Y / ws.CellSize)

	// Bounds checking
	if cellX < 0 || cellX >= ws.MapWidth || cellY < 0 || cellY >= ws.MapHeight {
		return WindVector{X: 0, Y: 0, Strength: 0, Turbulence: 0}
	}

	return ws.WindMap[cellY][cellX]
}

// ReleasePollen creates pollen from a flowering plant
func (ws *WindSystem) ReleasePollen(plant *Plant, amount int) {
	if !plant.IsAlive || amount <= 0 {
		return
	}

	for i := 0; i < amount; i++ {
		// Create pollen grain with slight random offset from plant
		offset := 0.5 + rand.Float64()*1.0
		angle := rand.Float64() * 2 * math.Pi

		pos := Position{
			X: plant.Position.X + math.Cos(angle)*offset,
			Y: plant.Position.Y + math.Sin(angle)*offset,
		}

		pollen := &PollenGrain{
			ID:        ws.NextPollenID,
			Position:  pos,
			Velocity:  Vector2D{X: 0, Y: 0}, // Initial velocity is zero
			PlantID:   plant.ID,
			PlantType: plant.Type,
			Genetics:  ws.copyGenetics(plant.Traits),
			Viability: 1.0,
			Age:       0,
			MaxAge:    50 + rand.Intn(100), // 50-150 ticks lifespan
			Size:      0.1 + rand.Float64()*0.1,
		}

		ws.AllPollenGrains = append(ws.AllPollenGrains, pollen)
		ws.NextPollenID++
		ws.TotalPollenReleased++
	}
}

// updatePollenGrains moves pollen grains and handles their lifecycle
func (ws *WindSystem) updatePollenGrains() {
	aliveGrains := make([]*PollenGrain, 0, len(ws.AllPollenGrains))

	for _, grain := range ws.AllPollenGrains {
		// Age the grain
		grain.Age++
		grain.Viability = math.Max(0, 1.0-float64(grain.Age)/float64(grain.MaxAge))

		// Remove dead pollen
		if grain.Viability <= 0 {
			continue
		}

		// Apply wind forces
		wind := ws.GetWindAt(grain.Position)
		windForce := Vector2D{X: wind.X, Y: wind.Y}

		// Smaller pollen is more affected by wind
		windEffect := 1.0 / (grain.Size + 0.1)
		windForce = windForce.Multiply(windEffect)

		// Add some random turbulence
		turbulence := Vector2D{
			X: (rand.Float64() - 0.5) * wind.Turbulence,
			Y: (rand.Float64() - 0.5) * wind.Turbulence,
		}
		windForce = windForce.Add(turbulence)

		// Update velocity (pollen has very low mass, so responds quickly to wind)
		grain.Velocity = grain.Velocity.Multiply(0.9) // Air resistance
		grain.Velocity = grain.Velocity.Add(windForce.Multiply(0.1))

		// Update position
		grain.Position.X += grain.Velocity.X
		grain.Position.Y += grain.Velocity.Y

		aliveGrains = append(aliveGrains, grain)
	}

	ws.AllPollenGrains = aliveGrains
}

// updatePollenClouds manages pollen cloud formation and dispersal
func (ws *WindSystem) updatePollenClouds() {
	// TODO: Implement pollen cloud clustering and dispersal
	// For now, pollen grains operate individually
	activeClouds := make([]*PollenCloud, 0, len(ws.PollenClouds))

	for _, cloud := range ws.PollenClouds {
		cloud.Age++
		if cloud.Age < cloud.MaxAge {
			// Update cloud position based on wind
			wind := ws.GetWindAt(cloud.Position)
			cloud.Position.X += wind.X * 0.5
			cloud.Position.Y += wind.Y * 0.5

			activeClouds = append(activeClouds, cloud)
		}
	}

	ws.PollenClouds = activeClouds
}

// TryPollination attempts cross-pollination between pollen and plants
func (ws *WindSystem) TryPollination(plants []*Plant, speciationSystem *SpeciationSystem) []*Plant {
	newOffspring := make([]*Plant, 0)

	for _, grain := range ws.AllPollenGrains {
		if grain.Viability < 0.1 {
			continue // Too old to be effective
		}

		// Find nearby plants that could be pollinated
		for _, plant := range plants {
			if !plant.IsAlive || plant.ID == grain.PlantID {
				continue // Skip source plant and dead plants
			}

			distance := math.Sqrt(
				math.Pow(plant.Position.X-grain.Position.X, 2) +
					math.Pow(plant.Position.Y-grain.Position.Y, 2),
			)

			// Pollination range depends on plant size and pollen viability
			pollinationRange := (plant.Size + 1.0) * grain.Viability

			if distance <= pollinationRange {
				// Attempt cross-pollination
				if offspring := ws.crossPollinate(plant, grain, speciationSystem); offspring != nil {
					newOffspring = append(newOffspring, offspring)
					ws.SuccessfulPollinationsThisTick++
					ws.TotalCrossPollinations++

					// Remove pollen grain after successful pollination
					grain.Viability = 0
					break
				}
			}
		}
	}

	return newOffspring
}

// crossPollinate creates a hybrid offspring from two plant genetic sources
func (ws *WindSystem) crossPollinate(motherPlant *Plant, pollenGrain *PollenGrain, speciationSystem *SpeciationSystem) *Plant {
	// Check genetic compatibility first (if speciation system is available)
	if speciationSystem != nil {
		// Calculate genetic distance directly from the pollen grain genetics and mother plant
		distance := speciationSystem.calculateGeneticDistanceFromTraits(motherPlant.Traits, pollenGrain.Genetics)
		if distance > speciationSystem.GeneticDistanceThreshold {
			return nil // Genetic distance too large for reproduction
		}
	}

	// Only pollinate compatible plant types (same type or similar)
	if motherPlant.Type != pollenGrain.PlantType {
		// Allow some cross-type pollination based on compatibility
		if !ws.areCompatibleTypes(motherPlant.Type, pollenGrain.PlantType) {
			return nil
		}
	}

	// Check if mother plant is ready for reproduction
	if !motherPlant.CanReproduce() {
		return nil
	}

	// Create offspring location near mother plant
	offset := 1.0 + rand.Float64()*2.0
	angle := rand.Float64() * 2 * math.Pi

	newPos := Position{
		X: motherPlant.Position.X + math.Cos(angle)*offset,
		Y: motherPlant.Position.Y + math.Sin(angle)*offset,
	}

	// Determine offspring type (usually mother's type, sometimes hybrid)
	var offspringType PlantType
	if rand.Float64() < 0.9 {
		offspringType = motherPlant.Type
	} else {
		offspringType = pollenGrain.PlantType
	}

	offspring := NewPlant(ws.NextPollenID, offspringType, newPos)
	ws.NextPollenID++
	offspring.Generation = motherPlant.Generation + 1

	// Genetic mixing - combine traits from both parents
	motherTraits := motherPlant.Traits
	fatherTraits := pollenGrain.Genetics

	for traitName := range motherTraits {
		var newValue float64

		motherValue := motherTraits[traitName].Value
		fatherValue := 0.0

		if fatherTrait, exists := fatherTraits[traitName]; exists {
			fatherValue = fatherTrait.Value
		}

		// Genetic mixing with some randomization
		if rand.Float64() < 0.5 {
			newValue = motherValue // Mother's trait
		} else {
			newValue = fatherValue // Father's trait
		}

		// Add some genetic variation
		newValue += rand.NormFloat64() * 0.05
		newValue = math.Max(-1.0, math.Min(1.0, newValue))

		offspring.SetTrait(traitName, newValue)
	}

	// Mother plant pays energy cost
	config := GetPlantConfigs()[motherPlant.Type]
	motherPlant.Energy -= config.BaseEnergy * 0.3

	return offspring
}

// areCompatibleTypes determines if two plant types can cross-pollinate
func (ws *WindSystem) areCompatibleTypes(type1, type2 PlantType) bool { // Define compatibility groups
	compatible := map[PlantType][]PlantType{
		PlantGrass:    {PlantGrass, PlantBush},
		PlantBush:     {PlantBush, PlantGrass, PlantTree},
		PlantTree:     {PlantTree, PlantBush},
		PlantMushroom: {PlantMushroom}, // Mushrooms only breed with mushrooms
		PlantAlgae:    {PlantAlgae},    // Algae only breeds with algae
		PlantCactus:   {PlantCactus},   // Cactus only breeds with cactus
	}

	if compatibleTypes, exists := compatible[type1]; exists {
		for _, compatibleType := range compatibleTypes {
			if compatibleType == type2 {
				return true
			}
		}
	}

	return false
}

// updateSeasonalEffects adjusts wind patterns based on current season
func (ws *WindSystem) updateSeasonalEffects(season Season) {
	switch season {
	case Spring:
		ws.SeasonalMultiplier = 1.2 // Stronger winds for pollen dispersal
	case Summer:
		ws.SeasonalMultiplier = 0.8 // Calmer summer winds
	case Autumn:
		ws.SeasonalMultiplier = 1.4 // Strong autumn winds for seed dispersal
	case Winter:
		ws.SeasonalMultiplier = 1.6 // Strong winter storms
	}
}

// updateWeatherPattern changes weather conditions over time
func (ws *WindSystem) updateWeatherPattern() {
	ws.WeatherDuration--

	if ws.WeatherDuration <= 0 {
		// Change weather pattern
		oldPattern := ws.WeatherPattern

		// Weather transition probabilities
		switch oldPattern {
		case 0: // Calm -> Windy (60%) or Storm (10%) or Tornado (1%) or Hurricane (0.5%)
			if rand.Float64() < 0.6 {
				ws.WeatherPattern = 1
			} else if rand.Float64() < 0.1 {
				ws.WeatherPattern = 2
			} else if rand.Float64() < 0.01 {
				ws.WeatherPattern = 3 // Tornado
			} else if rand.Float64() < 0.005 {
				ws.WeatherPattern = 4 // Hurricane
			}
		case 1: // Windy -> Calm (40%) or Storm (20%) or Tornado (2%)
			if rand.Float64() < 0.4 {
				ws.WeatherPattern = 0
			} else if rand.Float64() < 0.2 {
				ws.WeatherPattern = 2
			} else if rand.Float64() < 0.02 {
				ws.WeatherPattern = 3 // Tornado
			}
		case 2: // Storm -> Windy (50%) or Calm (30%) or Tornado (5%) or Hurricane (2%)
			if rand.Float64() < 0.5 {
				ws.WeatherPattern = 1
			} else if rand.Float64() < 0.3 {
				ws.WeatherPattern = 0
			} else if rand.Float64() < 0.05 {
				ws.WeatherPattern = 3 // Tornado
			} else if rand.Float64() < 0.02 {
				ws.WeatherPattern = 4 // Hurricane
			}
		case 3: // Tornado -> Windy (70%) or Storm (20%) or Calm (10%)
			if rand.Float64() < 0.7 {
				ws.WeatherPattern = 1
			} else if rand.Float64() < 0.2 {
				ws.WeatherPattern = 2
			} else {
				ws.WeatherPattern = 0
			}
		case 4: // Hurricane -> Storm (60%) or Windy (30%) or Calm (10%)
			if rand.Float64() < 0.6 {
				ws.WeatherPattern = 2
			} else if rand.Float64() < 0.3 {
				ws.WeatherPattern = 1
			} else {
				ws.WeatherPattern = 0
			}
		}

		// Set new weather duration
		switch ws.WeatherPattern {
		case 0: // Calm
			ws.WeatherDuration = 150 + rand.Intn(200)
		case 1: // Windy
			ws.WeatherDuration = 100 + rand.Intn(150)
		case 2: // Storm
			ws.WeatherDuration = 30 + rand.Intn(70)
		case 3: // Tornado
			ws.WeatherDuration = 10 + rand.Intn(20) // Short duration
		case 4: // Hurricane
			ws.WeatherDuration = 50 + rand.Intn(100) // Longer duration
		}
	}
}

// copyGenetics creates a copy of genetic traits
func (ws *WindSystem) copyGenetics(traits map[string]Trait) map[string]Trait {
	genetics := make(map[string]Trait)
	for name, trait := range traits {
		genetics[name] = trait
	}
	return genetics
}

// GetWindStats returns statistics about the wind system
func (ws *WindSystem) GetWindStats() map[string]interface{} {
	return map[string]interface{}{
		"base_wind_direction":      ws.BaseWindDirection,
		"base_wind_strength":       ws.BaseWindStrength,
		"seasonal_multiplier":      ws.SeasonalMultiplier,
		"weather_pattern":          ws.WeatherPattern,
		"weather_duration":         ws.WeatherDuration,
		"active_pollen_grains":     len(ws.AllPollenGrains),
		"total_pollen_released":    ws.TotalPollenReleased,
		"pollinations_this_tick":   ws.SuccessfulPollinationsThisTick,
		"total_cross_pollinations": ws.TotalCrossPollinations,
	}
}

// GetWeatherDescription returns human-readable weather description
func (ws *WindSystem) GetWeatherDescription() string {
	switch ws.WeatherPattern {
	case 0:
		return "Calm"
	case 1:
		return "Windy"
	case 2:
		return "Storm"
	case 3:
		return "Tornado"
	case 4:
		return "Hurricane"
	default:
		return "Unknown"
	}
}

// RegionalStormEffect represents the effect of regional storms on wind
type RegionalStormEffect struct {
	Strength             float64 // Wind strength multiplier
	DirectionChange      float64 // Change in wind direction (radians)
	TurbulenceMultiplier float64 // Turbulence multiplier
}

// getRegionalStormEffect calculates the combined effect of all regional storms at a position
func (ws *WindSystem) getRegionalStormEffect(pos Position) RegionalStormEffect {
	effect := RegionalStormEffect{
		Strength:             1.0,
		DirectionChange:      0.0,
		TurbulenceMultiplier: 1.0,
	}

	for _, storm := range ws.RegionalStorms {
		distance := math.Sqrt(math.Pow(pos.X-storm.Center.X, 2) + math.Pow(pos.Y-storm.Center.Y, 2))
		
		if distance <= storm.Radius {
			// Within storm effect radius
			stormInfluence := (storm.Radius - distance) / storm.Radius // 0 to 1
			stormInfluence *= storm.Intensity
			
			switch storm.Type {
			case StormThunderstorm:
				effect.Strength *= 1.0 + stormInfluence*0.8
				effect.TurbulenceMultiplier *= 1.0 + stormInfluence*2.0
				effect.DirectionChange += (rand.Float64() - 0.5) * stormInfluence * 0.3
				
			case StormTornado:
				// Tornado creates circular wind pattern
				angle := math.Atan2(pos.Y-storm.Center.Y, pos.X-storm.Center.X)
				effect.DirectionChange += angle + math.Pi/2 // Perpendicular to radius for rotation
				effect.Strength *= 1.0 + stormInfluence*3.0
				effect.TurbulenceMultiplier *= 1.0 + stormInfluence*5.0
				
			case StormHurricane:
				// Hurricane creates spiral pattern
				angle := math.Atan2(pos.Y-storm.Center.Y, pos.X-storm.Center.X)
				spiralAngle := angle + math.Pi/4 // 45-degree spiral
				effect.DirectionChange += spiralAngle * stormInfluence
				effect.Strength *= 1.0 + stormInfluence*2.5
				effect.TurbulenceMultiplier *= 1.0 + stormInfluence*3.0
				
			case StormBlizzard:
				effect.Strength *= 1.0 + stormInfluence*1.5
				effect.TurbulenceMultiplier *= 1.0 + stormInfluence*2.5
				effect.DirectionChange += (rand.Float64() - 0.5) * stormInfluence * 0.4
				
			case StormDustStorm:
				effect.Strength *= 1.0 + stormInfluence*1.2
				effect.TurbulenceMultiplier *= 1.0 + stormInfluence*2.0
				effect.DirectionChange += (rand.Float64() - 0.5) * stormInfluence * 0.2
			}
		}
	}

	return effect
}

// updateRegionalStorms updates all active regional storms
func (ws *WindSystem) updateRegionalStorms() {
	activeStorms := make([]RegionalStorm, 0)

	for _, storm := range ws.RegionalStorms {
		storm.Duration--
		
		// Move storm
		storm.Center.X += math.Cos(storm.MovementDir) * storm.Speed
		storm.Center.Y += math.Sin(storm.MovementDir) * storm.Speed
		
		// Intensity may change over time
		ageRatio := float64(storm.MaxDuration-storm.Duration) / float64(storm.MaxDuration)
		if ageRatio < 0.3 {
			// Growing phase
			storm.Intensity = math.Min(1.0, storm.Intensity+0.05)
		} else if ageRatio > 0.7 {
			// Weakening phase
			storm.Intensity = math.Max(0.1, storm.Intensity-0.03)
		}
		
		// Keep active storms
		if storm.Duration > 0 {
			activeStorms = append(activeStorms, storm)
		}
	}

	ws.RegionalStorms = activeStorms
}

// spawnRegionalStorm creates a new regional storm
func (ws *WindSystem) spawnRegionalStorm() {
	if len(ws.RegionalStorms) >= 3 { // Limit concurrent storms
		return
	}

	worldWidth := float64(ws.MapWidth) * ws.CellSize
	worldHeight := float64(ws.MapHeight) * ws.CellSize
	
	storm := RegionalStorm{
		ID:          len(ws.RegionalStorms) + 1,
		Center:      Position{X: rand.Float64() * worldWidth, Y: rand.Float64() * worldHeight},
		MovementDir: rand.Float64() * 2 * math.Pi,
		Speed:       0.5 + rand.Float64()*1.5,
		Intensity:   0.2 + rand.Float64()*0.3, // Start with moderate intensity
	}

	// Choose storm type based on season and probability
	switch rand.Intn(5) {
	case 0:
		storm.Type = StormThunderstorm
		storm.Radius = 15 + rand.Float64()*25
		storm.Duration = 30 + rand.Intn(50)
		storm.MaxDuration = storm.Duration
	case 1:
		storm.Type = StormTornado
		storm.Radius = 5 + rand.Float64()*15
		storm.Duration = 10 + rand.Intn(20)
		storm.MaxDuration = storm.Duration
	case 2:
		storm.Type = StormHurricane
		storm.Radius = 40 + rand.Float64()*60
		storm.Duration = 80 + rand.Intn(120)
		storm.MaxDuration = storm.Duration
	case 3:
		storm.Type = StormBlizzard
		storm.Radius = 20 + rand.Float64()*40
		storm.Duration = 60 + rand.Intn(80)
		storm.MaxDuration = storm.Duration
	case 4:
		storm.Type = StormDustStorm
		storm.Radius = 25 + rand.Float64()*35
		storm.Duration = 40 + rand.Intn(60)
		storm.MaxDuration = storm.Duration
	}

	ws.RegionalStorms = append(ws.RegionalStorms, storm)
}

// GetRegionalStormStats returns statistics about regional storms
func (ws *WindSystem) GetRegionalStormStats() map[string]interface{} {
	stormCounts := make(map[string]int)
	totalIntensity := 0.0
	
	for _, storm := range ws.RegionalStorms {
		switch storm.Type {
		case StormThunderstorm:
			stormCounts["thunderstorms"]++
		case StormTornado:
			stormCounts["tornadoes"]++
		case StormHurricane:
			stormCounts["hurricanes"]++
		case StormBlizzard:
			stormCounts["blizzards"]++
		case StormDustStorm:
			stormCounts["dust_storms"]++
		}
		totalIntensity += storm.Intensity
	}
	
	return map[string]interface{}{
		"active_storms":   len(ws.RegionalStorms),
		"storm_types":     stormCounts,
		"total_intensity": totalIntensity,
	}
}
