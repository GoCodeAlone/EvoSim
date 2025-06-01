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
	WeatherPattern     int     // Current weather (0=calm, 1=windy, 2=storm)
	WeatherDuration    int     // How long current weather lasts

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
			}

			// Convert to vector components
			windX := math.Cos(direction) * strength
			windY := math.Sin(direction) * strength

			ws.WindMap[y][x] = WindVector{
				X:          windX,
				Y:          windY,
				Strength:   strength,
				Turbulence: ws.TurbulenceLevel * float64(ws.WeatherPattern+1),
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
		case 0: // Calm -> Windy (60%) or Storm (10%)
			if rand.Float64() < 0.6 {
				ws.WeatherPattern = 1
			} else if rand.Float64() < 0.1 {
				ws.WeatherPattern = 2
			}
		case 1: // Windy -> Calm (40%) or Storm (20%)
			if rand.Float64() < 0.4 {
				ws.WeatherPattern = 0
			} else if rand.Float64() < 0.2 {
				ws.WeatherPattern = 2
			}
		case 2: // Storm -> Windy (50%) or Calm (30%)
			if rand.Float64() < 0.5 {
				ws.WeatherPattern = 1
			} else if rand.Float64() < 0.3 {
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
	default:
		return "Unknown"
	}
}
