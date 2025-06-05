package main

import (
	"fmt"
	"math"
	"math/rand"
)

// EnvironmentalPressure represents a long-term environmental stress
type EnvironmentalPressure struct {
	ID           int                    `json:"id"`
	Type         string                 `json:"type"` // "climate_change", "pollution", "habitat_fragmentation", etc.
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Severity     float64                `json:"severity"` // 0.0 to 1.0, how intense the pressure is
	StartTick    int                    `json:"start_tick"`
	Duration     int                    `json:"duration"`      // -1 for permanent
	AffectedArea Position               `json:"affected_area"` // Center of affected region
	Radius       float64                `json:"radius"`        // Radius of effect
	Effects      map[string]interface{} `json:"effects"`       // Various effects this pressure causes
	IsActive     bool                   `json:"is_active"`
}

// EnvironmentalPressureSystem manages long-term environmental pressures
type EnvironmentalPressureSystem struct {
	ActivePressures []*EnvironmentalPressure `json:"active_pressures"`
	PressureHistory []*EnvironmentalPressure `json:"pressure_history"`
	NextPressureID  int                      `json:"next_pressure_id"`

	// Configuration
	MaxActivePressures     int     `json:"max_active_pressures"`
	ClimateChangeRate      float64 `json:"climate_change_rate"`     // How quickly climate changes
	PollutionAccumulation  float64 `json:"pollution_accumulation"`  // Rate of pollution buildup
	FragmentationThreshold float64 `json:"fragmentation_threshold"` // Population density threshold for fragmentation
}

// PressureType constants
const (
	PressureClimateChange        = "climate_change"
	PressurePollution            = "pollution"
	PressureHabitatFragmentation = "habitat_fragmentation"
	PressureInvasiveSpecies      = "invasive_species"
	PressureResourceDepletion    = "resource_depletion"
	PressureExtremeWeather       = "extreme_weather"
)

// NewEnvironmentalPressureSystem creates a new environmental pressure system
func NewEnvironmentalPressureSystem() *EnvironmentalPressureSystem {
	return &EnvironmentalPressureSystem{
		ActivePressures:        make([]*EnvironmentalPressure, 0),
		PressureHistory:        make([]*EnvironmentalPressure, 0),
		NextPressureID:         1,
		MaxActivePressures:     3,      // Maximum 3 active pressures at once
		ClimateChangeRate:      0.001,  // 0.1% change per tick
		PollutionAccumulation:  0.0005, // 0.05% pollution increase per tick
		FragmentationThreshold: 0.1,    // 10% population density threshold
	}
}

// Update processes all active environmental pressures
func (eps *EnvironmentalPressureSystem) Update(world *World, tick int) {
	// Update existing pressures
	eps.updateActivePressures(world, tick)

	// Potentially trigger new pressures
	eps.evaluateNewPressures(world, tick)

	// Apply pressure effects to world
	eps.applyPressureEffects(world, tick)
}

// updateActivePressures updates duration and effects of active pressures
func (eps *EnvironmentalPressureSystem) updateActivePressures(world *World, tick int) {
	for i := len(eps.ActivePressures) - 1; i >= 0; i-- {
		pressure := eps.ActivePressures[i]

		// Check if pressure should end
		if pressure.Duration > 0 {
			pressure.Duration--
			if pressure.Duration <= 0 {
				pressure.IsActive = false
				eps.PressureHistory = append(eps.PressureHistory, pressure)
				eps.ActivePressures = append(eps.ActivePressures[:i], eps.ActivePressures[i+1:]...)
				continue
			}
		}

		// Update pressure intensity based on type
		switch pressure.Type {
		case PressureClimateChange:
			eps.updateClimateChangePressure(pressure, world, tick)
		case PressurePollution:
			eps.updatePollutionPressure(pressure, world, tick)
		case PressureHabitatFragmentation:
			eps.updateFragmentationPressure(pressure, world, tick)
		}
	}
}

// evaluateNewPressures checks if new pressures should be triggered
func (eps *EnvironmentalPressureSystem) evaluateNewPressures(world *World, tick int) {
	if len(eps.ActivePressures) >= eps.MaxActivePressures {
		return // Already at maximum
	}

	// Climate change trigger - based on population density and energy usage
	if rand.Float64() < 0.001 && !eps.hasPressureType(PressureClimateChange) {
		totalPopulation := len(world.AllEntities) + len(world.AllPlants)
		if totalPopulation > 100 { // High population threshold
			eps.triggerClimateChange(world, tick)
		}
	}

	// Pollution trigger - based on civilization development
	if rand.Float64() < 0.002 && !eps.hasPressureType(PressurePollution) {
		if world.CivilizationSystem != nil && len(world.CivilizationSystem.Structures) > 20 {
			eps.triggerPollution(world, tick)
		}
	}

	// Habitat fragmentation - based on entity distribution
	if rand.Float64() < 0.003 && !eps.hasPressureType(PressureHabitatFragmentation) {
		if eps.isPopulationFragmented(world) {
			eps.triggerHabitatFragmentation(world, tick)
		}
	}

	// Resource depletion - based on ecosystem health
	if rand.Float64() < 0.001 && !eps.hasPressureType(PressureResourceDepletion) {
		if world.EcosystemMonitor != nil {
			healthScore := world.EcosystemMonitor.GetHealthScore()
			if healthScore < 30 { // Low ecosystem health
				eps.triggerResourceDepletion(world, tick)
			}
		}
	}
}

// hasPressureType checks if a pressure type is already active
func (eps *EnvironmentalPressureSystem) hasPressureType(pressureType string) bool {
	for _, pressure := range eps.ActivePressures {
		if pressure.Type == pressureType {
			return true
		}
	}
	return false
}

// triggerClimateChange creates a climate change pressure
func (eps *EnvironmentalPressureSystem) triggerClimateChange(world *World, tick int) {
	pressure := &EnvironmentalPressure{
		ID:           eps.NextPressureID,
		Type:         PressureClimateChange,
		Name:         "Global Climate Change",
		Description:  "Rising temperatures and changing weather patterns affecting all biomes",
		Severity:     0.3 + rand.Float64()*0.4, // 30-70% severity
		StartTick:    tick,
		Duration:     -1,                                                              // Permanent
		AffectedArea: Position{X: world.Config.Width / 2, Y: world.Config.Height / 2}, // Global center
		Radius:       math.Max(world.Config.Width, world.Config.Height),               // Global effect
		Effects: map[string]interface{}{
			"temperature_change":   (rand.Float64() - 0.5) * 4.0, // ±2°C change
			"precipitation_change": (rand.Float64() - 0.5) * 0.6, // ±30% precipitation change
			"biome_shift_rate":     0.02,                         // 2% chance of biome changes per tick
			"extreme_weather_rate": 0.005,                        // Increased extreme weather
		},
		IsActive: true,
	}

	eps.NextPressureID++
	eps.ActivePressures = append(eps.ActivePressures, pressure)
}

// triggerPollution creates a pollution pressure
func (eps *EnvironmentalPressureSystem) triggerPollution(world *World, tick int) {
	// Find area with highest civilization density
	centerX, centerY := eps.findCivilizationCenter(world)

	pressure := &EnvironmentalPressure{
		ID:           eps.NextPressureID,
		Type:         PressurePollution,
		Name:         "Environmental Pollution",
		Description:  "Toxic contamination spreading from industrial activities",
		Severity:     0.2 + rand.Float64()*0.5, // 20-70% severity
		StartTick:    tick,
		Duration:     500 + rand.Intn(1000), // 500-1500 ticks
		AffectedArea: Position{X: centerX, Y: centerY},
		Radius:       10.0 + rand.Float64()*15.0, // 10-25 unit radius
		Effects: map[string]interface{}{
			"toxicity_increase":    0.1,  // 10% toxicity increase in affected area
			"reproduction_penalty": 0.3,  // 30% reproduction penalty
			"health_decay":         0.02, // 2% health decay per tick
			"mutation_rate_bonus":  0.5,  // 50% increased mutation rate (adaptation pressure)
		},
		IsActive: true,
	}

	eps.NextPressureID++
	eps.ActivePressures = append(eps.ActivePressures, pressure)
}

// triggerHabitatFragmentation creates habitat fragmentation pressure
func (eps *EnvironmentalPressureSystem) triggerHabitatFragmentation(world *World, tick int) {
	pressure := &EnvironmentalPressure{
		ID:           eps.NextPressureID,
		Type:         PressureHabitatFragmentation,
		Name:         "Habitat Fragmentation",
		Description:  "Landscape divided into isolated patches, disrupting natural movement",
		Severity:     0.4 + rand.Float64()*0.4, // 40-80% severity
		StartTick:    tick,
		Duration:     -1, // Permanent structural change
		AffectedArea: Position{X: world.Config.Width / 2, Y: world.Config.Height / 2},
		Radius:       math.Max(world.Config.Width, world.Config.Height), // Global effect
		Effects: map[string]interface{}{
			"movement_penalty":     0.5,  // 50% movement penalty between patches
			"migration_difficulty": 0.7,  // 70% more difficult migration
			"genetic_isolation":    true, // Isolated populations evolve separately
			"edge_effect":          0.3,  // 30% reduced resources at edges
		},
		IsActive: true,
	}

	eps.NextPressureID++
	eps.ActivePressures = append(eps.ActivePressures, pressure)
}

// triggerResourceDepletion creates resource depletion pressure
func (eps *EnvironmentalPressureSystem) triggerResourceDepletion(world *World, tick int) {
	pressure := &EnvironmentalPressure{
		ID:           eps.NextPressureID,
		Type:         PressureResourceDepletion,
		Name:         "Resource Depletion Crisis",
		Description:  "Overharvesting has depleted critical ecosystem resources",
		Severity:     0.5 + rand.Float64()*0.3, // 50-80% severity
		StartTick:    tick,
		Duration:     200 + rand.Intn(500), // 200-700 ticks recovery time
		AffectedArea: Position{X: world.Config.Width / 2, Y: world.Config.Height / 2},
		Radius:       math.Max(world.Config.Width, world.Config.Height) * 0.7, // Most of world
		Effects: map[string]interface{}{
			"food_availability": 0.4, // 60% reduction in food availability
			"water_scarcity":    0.3, // 70% reduction in water
			"soil_depletion":    0.5, // 50% reduction in soil nutrients
			"carrying_capacity": 0.6, // 40% reduction in carrying capacity
		},
		IsActive: true,
	}

	eps.NextPressureID++
	eps.ActivePressures = append(eps.ActivePressures, pressure)
}

// Helper functions for pressure management

// findCivilizationCenter finds the center of civilization activity
func (eps *EnvironmentalPressureSystem) findCivilizationCenter(world *World) (float64, float64) {
	if world.CivilizationSystem == nil || len(world.CivilizationSystem.Structures) == 0 {
		return world.Config.Width / 2, world.Config.Height / 2
	}

	totalX, totalY := 0.0, 0.0
	count := 0

	for _, structure := range world.CivilizationSystem.Structures {
		totalX += structure.Position.X
		totalY += structure.Position.Y
		count++
	}

	if count == 0 {
		return world.Config.Width / 2, world.Config.Height / 2
	}

	return totalX / float64(count), totalY / float64(count)
}

// isPopulationFragmented checks if the population is fragmented into isolated groups
func (eps *EnvironmentalPressureSystem) isPopulationFragmented(world *World) bool {
	if len(world.AllEntities) < 20 {
		return false // Too small to fragment
	}

	// Simple fragmentation check: count isolated groups
	gridSize := 5.0 // Grid cell size for grouping
	groups := make(map[string]int)

	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			gridX := int(entity.Position.X / gridSize)
			gridY := int(entity.Position.Y / gridSize)
			key := fmt.Sprintf("%d,%d", gridX, gridY)
			groups[key]++
		}
	}

	// If population is spread across many small groups, it's fragmented
	avgGroupSize := float64(len(world.AllEntities)) / float64(len(groups))
	return len(groups) > 5 && avgGroupSize < 3.0
}

// applyPressureEffects applies the effects of all active pressures to the world
func (eps *EnvironmentalPressureSystem) applyPressureEffects(world *World, tick int) {
	for _, pressure := range eps.ActivePressures {
		switch pressure.Type {
		case PressureClimateChange:
			eps.applyClimateChangeEffects(pressure, world, tick)
		case PressurePollution:
			eps.applyPollutionEffects(pressure, world, tick)
		case PressureHabitatFragmentation:
			eps.applyFragmentationEffects(pressure, world, tick)
		case PressureResourceDepletion:
			eps.applyResourceDepletionEffects(pressure, world, tick)
		}
	}
}

// Pressure-specific effect application methods

// updateClimateChangePressure updates climate change pressure over time
func (eps *EnvironmentalPressureSystem) updateClimateChangePressure(pressure *EnvironmentalPressure, world *World, tick int) {
	// Climate change gradually intensifies
	if rand.Float64() < eps.ClimateChangeRate {
		pressure.Severity = math.Min(pressure.Severity+0.01, 1.0)
	}
}

// updatePollutionPressure updates pollution pressure over time
func (eps *EnvironmentalPressureSystem) updatePollutionPressure(pressure *EnvironmentalPressure, world *World, tick int) {
	// Pollution can spread slowly
	if rand.Float64() < eps.PollutionAccumulation {
		pressure.Radius = math.Min(pressure.Radius+0.5, 50.0)
	}
}

// updateFragmentationPressure updates fragmentation pressure over time
func (eps *EnvironmentalPressureSystem) updateFragmentationPressure(pressure *EnvironmentalPressure, world *World, tick int) {
	// Fragmentation effects compound over time
	if tick%100 == 0 { // Every 100 ticks
		pressure.Severity = math.Min(pressure.Severity+0.05, 1.0)
	}
}

// applyClimateChangeEffects applies climate change effects to the world
func (eps *EnvironmentalPressureSystem) applyClimateChangeEffects(pressure *EnvironmentalPressure, world *World, tick int) {
	biomeShiftRate := pressure.Effects["biome_shift_rate"].(float64) * pressure.Severity

	// Random biome changes due to climate shift
	if rand.Float64() < biomeShiftRate {
		x := rand.Intn(world.Config.GridWidth)
		y := rand.Intn(world.Config.GridHeight)

		// Shift towards warmer/dryer biomes
		currentBiome := world.Grid[y][x].Biome
		newBiome := eps.getClimateShiftedBiome(currentBiome)
		if newBiome != currentBiome {
			world.Grid[y][x].Biome = newBiome
		}
	}

	// Increase extreme weather events
	extremeWeatherRate := pressure.Effects["extreme_weather_rate"].(float64) * pressure.Severity
	if rand.Float64() < extremeWeatherRate {
		world.triggerEnhancedEnvironmentalEvent()
	}
}

// applyPollutionEffects applies pollution effects to entities and plants
func (eps *EnvironmentalPressureSystem) applyPollutionEffects(pressure *EnvironmentalPressure, world *World, tick int) {
	healthDecay := pressure.Effects["health_decay"].(float64) * pressure.Severity
	reproductionPenalty := pressure.Effects["reproduction_penalty"].(float64) * pressure.Severity
	mutationBonus := pressure.Effects["mutation_rate_bonus"].(float64) * pressure.Severity

	// Apply effects to entities in affected area
	for _, entity := range world.AllEntities {
		if entity.IsAlive && eps.isInAffectedArea(entity.Position, pressure) {
			// Health decay - check if trait exists first
			if healthTrait, exists := entity.Traits["health"]; exists {
				currentHealth := healthTrait.Value
				entity.Traits["health"] = Trait{Value: math.Max(0, currentHealth-healthDecay)}
			}

			// Reproduction penalty (temporary trait modification)
			if fertilityTrait, exists := entity.Traits["fertility"]; exists {
				currentFertility := fertilityTrait.Value
				entity.Traits["fertility"] = Trait{Value: math.Max(0, currentFertility*(1.0-reproductionPenalty))}
			}

			// Increased mutation rate for adaptation - use energy as proxy for mutation stress
			if entity.Energy > 0 {
				entity.Energy *= (1.0 - mutationBonus*0.1) // Stress reduces energy
			}
		}
	}

	// Apply effects to plants in affected area
	for _, plant := range world.AllPlants {
		if plant.IsAlive && eps.isInAffectedArea(plant.Position, pressure) {
			// Reduce plant health and growth
			plant.Energy *= (1.0 - healthDecay)
			plant.GrowthRate *= (1.0 - healthDecay/2)
		}
	}
}

// applyFragmentationEffects applies habitat fragmentation effects
func (eps *EnvironmentalPressureSystem) applyFragmentationEffects(pressure *EnvironmentalPressure, world *World, tick int) {
	movementPenalty := pressure.Effects["movement_penalty"].(float64) * pressure.Severity

	// Apply movement penalties to entities crossing "boundaries"
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			// Simulate movement difficulty in fragmented landscape
			if speedTrait, exists := entity.Traits["speed"]; exists {
				currentSpeed := speedTrait.Value
				entity.Traits["speed"] = Trait{Value: currentSpeed * (1.0 - movementPenalty*0.5)}
			}
		}
	}
}

// applyResourceDepletionEffects applies resource depletion effects
func (eps *EnvironmentalPressureSystem) applyResourceDepletionEffects(pressure *EnvironmentalPressure, world *World, tick int) {
	foodReduction := 1.0 - pressure.Effects["food_availability"].(float64)
	waterReduction := 1.0 - pressure.Effects["water_scarcity"].(float64)

	// Reduce available resources in grid cells
	for y := 0; y < world.Config.GridHeight; y++ {
		for x := 0; x < world.Config.GridWidth; x++ {
			cellPos := Position{X: float64(x), Y: float64(y)}
			if eps.isInAffectedArea(cellPos, pressure) {
				cell := &world.Grid[y][x]
				cell.WaterLevel *= (1.0 - waterReduction*pressure.Severity)

				// Reduce soil nutrients
				for nutrient, level := range cell.SoilNutrients {
					cell.SoilNutrients[nutrient] = level * (1.0 - foodReduction*pressure.Severity*0.5)
				}
			}
		}
	}
}

// Helper functions

// isInAffectedArea checks if a position is within the pressure's affected area
func (eps *EnvironmentalPressureSystem) isInAffectedArea(pos Position, pressure *EnvironmentalPressure) bool {
	distance := math.Sqrt(math.Pow(pos.X-pressure.AffectedArea.X, 2) + math.Pow(pos.Y-pressure.AffectedArea.Y, 2))
	return distance <= pressure.Radius
}

// getClimateShiftedBiome returns a biome that represents climate change effects
func (eps *EnvironmentalPressureSystem) getClimateShiftedBiome(currentBiome BiomeType) BiomeType {
	// Climate change generally shifts towards warmer, drier conditions
	switch currentBiome {
	case BiomeIce:
		return BiomeTundra
	case BiomeTundra:
		return BiomePlains
	case BiomeRainforest:
		return BiomeForest
	case BiomeForest:
		if rand.Float64() < 0.5 {
			return BiomePlains
		}
		return BiomeDesert
	case BiomePlains:
		return BiomeDesert
	case BiomeSwamp:
		return BiomePlains
	default:
		return currentBiome // No change for other biomes
	}
}

// GetPressureStats returns statistics about active pressures
func (eps *EnvironmentalPressureSystem) GetPressureStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["active_pressures"] = len(eps.ActivePressures)
	stats["total_pressure_history"] = len(eps.PressureHistory)

	// Count by type
	typeCount := make(map[string]int)
	totalSeverity := 0.0
	for _, pressure := range eps.ActivePressures {
		typeCount[pressure.Type]++
		totalSeverity += pressure.Severity
	}

	stats["pressure_types"] = typeCount
	if len(eps.ActivePressures) > 0 {
		stats["average_severity"] = totalSeverity / float64(len(eps.ActivePressures))
	} else {
		stats["average_severity"] = 0.0
	}

	return stats
}
