package main

import (
	"math"
	"math/rand"
)

// BiomeBoundaryType represents the type of boundary between biomes
type BiomeBoundaryType int

const (
	SoftBoundary    BiomeBoundaryType = iota // Gradual transition
	SharpBoundary                            // Abrupt transition
	EcotoneZone                              // Distinct transition zone with mixed characteristics
	BarrierBoundary                          // Boundary that restricts movement
)

// BiomeBoundary represents the interaction zone between two biomes
type BiomeBoundary struct {
	ID           int               `json:"id"`
	Position     Position          `json:"position"` // Central position of boundary
	BiomeA       BiomeType         `json:"biome_a"`  // First biome
	BiomeB       BiomeType         `json:"biome_b"`  // Second biome
	BoundaryType BiomeBoundaryType `json:"boundary_type"`
	Width        float64           `json:"width"`        // Width of transition zone
	Permeability float64           `json:"permeability"` // How easily entities can cross (0-1)

	// Environmental mixing effects
	TemperatureGradient float64            `json:"temperature_gradient"` // Temperature change across boundary
	MoistureGradient    float64            `json:"moisture_gradient"`    // Moisture change across boundary
	TraitModifiers      map[string]float64 `json:"trait_modifiers"`      // Special trait effects in boundary zone

	// Ecological effects
	ResourceDensity  float64 `json:"resource_density"`  // Resource availability in ecotone
	SpeciesDiversity float64 `json:"species_diversity"` // Higher diversity in transition zones
	CompetitionLevel float64 `json:"competition_level"` // Competition intensity

	// Dynamic properties
	Stability  float64 `json:"stability"`   // How stable this boundary is (0-1)
	ChangeRate float64 `json:"change_rate"` // How quickly boundary shifts
	LastUpdate int     `json:"last_update"`
}

// BiomeBoundaryEffect represents an effect that occurs at biome boundaries
type BiomeBoundaryEffect struct {
	Type       string             `json:"type"`       // "migration", "evolution", "resource", "competition"
	Intensity  float64            `json:"intensity"`  // Effect strength (0-1)
	Duration   int                `json:"duration"`   // How long effect lasts
	Radius     float64            `json:"radius"`     // Area of effect
	Parameters map[string]float64 `json:"parameters"` // Effect-specific parameters
}

// BiomeBoundarySystem manages interactions between adjacent biomes
type BiomeBoundarySystem struct {
	Boundaries     []*BiomeBoundary       `json:"boundaries"`
	ActiveEffects  []*BiomeBoundaryEffect `json:"active_effects"`
	NextBoundaryID int                    `json:"next_boundary_id"`

	// System parameters
	DetectionRadius   float64 `json:"detection_radius"`   // How far to detect boundaries
	UpdateFrequency   int     `json:"update_frequency"`   // Ticks between updates
	EvolutionPressure float64 `json:"evolution_pressure"` // Extra evolution pressure at boundaries
	MigrationBonus    float64 `json:"migration_bonus"`    // Movement bonus in ecotones

	// Statistics
	TotalBoundaryLength float64 `json:"total_boundary_length"`
	EcotoneArea         float64 `json:"ecotone_area"`
	MigrationEvents     int     `json:"migration_events"`
	EvolutionEvents     int     `json:"evolution_events"`
}

// NewBiomeBoundarySystem creates a new biome boundary system
func NewBiomeBoundarySystem() *BiomeBoundarySystem {
	return &BiomeBoundarySystem{
		Boundaries:        make([]*BiomeBoundary, 0),
		ActiveEffects:     make([]*BiomeBoundaryEffect, 0),
		NextBoundaryID:    1,
		DetectionRadius:   5.0,
		UpdateFrequency:   25, // Update every 25 ticks
		EvolutionPressure: 0.15,
		MigrationBonus:    0.3,
	}
}

// Update processes biome boundary interactions
func (bbs *BiomeBoundarySystem) Update(world *World, tick int) {
	if tick%bbs.UpdateFrequency != 0 {
		return
	}

	// Update boundary detection and positions
	bbs.detectBoundaries(world)

	// Process boundary effects on entities
	bbs.processEntityBoundaryEffects(world, tick)

	// Process boundary effects on plants
	bbs.processBoundaryEffectsOnPlants(world, tick)

	// Update boundary dynamics
	bbs.updateBoundaryDynamics(world, tick)

	// Clean up expired effects
	bbs.cleanupExpiredEffects(tick)

	// Update statistics
	bbs.updateStatistics()
}

// detectBoundaries identifies and creates biome boundaries
func (bbs *BiomeBoundarySystem) detectBoundaries(world *World) {
	// Clear existing boundaries to recalculate
	bbs.Boundaries = make([]*BiomeBoundary, 0)

	// Scan grid for adjacent different biomes
	for y := 0; y < world.Config.GridHeight-1; y++ {
		for x := 0; x < world.Config.GridWidth-1; x++ {
			currentBiome := world.Grid[y][x].Biome

			// Check right neighbor
			if x < world.Config.GridWidth-1 {
				rightBiome := world.Grid[y][x+1].Biome
				if currentBiome != rightBiome {
					boundary := bbs.createBoundary(world, x, y, currentBiome, rightBiome, "horizontal")
					if boundary != nil {
						bbs.Boundaries = append(bbs.Boundaries, boundary)
					}
				}
			}

			// Check bottom neighbor
			if y < world.Config.GridHeight-1 {
				bottomBiome := world.Grid[y+1][x].Biome
				if currentBiome != bottomBiome {
					boundary := bbs.createBoundary(world, x, y, currentBiome, bottomBiome, "vertical")
					if boundary != nil {
						bbs.Boundaries = append(bbs.Boundaries, boundary)
					}
				}
			}
		}
	}
}

// createBoundary creates a boundary between two biomes
func (bbs *BiomeBoundarySystem) createBoundary(world *World, x, y int, biomeA, biomeB BiomeType, orientation string) *BiomeBoundary {
	// Calculate boundary position
	var pos Position
	if orientation == "horizontal" {
		pos = Position{
			X: float64(x) + 0.5,
			Y: float64(y),
		}
	} else {
		pos = Position{
			X: float64(x),
			Y: float64(y) + 0.5,
		}
	}

	boundary := &BiomeBoundary{
		ID:             bbs.NextBoundaryID,
		Position:       pos,
		BiomeA:         biomeA,
		BiomeB:         biomeB,
		BoundaryType:   bbs.determineBoundaryType(world, biomeA, biomeB),
		Width:          bbs.calculateBoundaryWidth(biomeA, biomeB),
		Permeability:   bbs.calculatePermeability(biomeA, biomeB),
		TraitModifiers: make(map[string]float64),
		Stability:      bbs.calculateStability(biomeA, biomeB),
		ChangeRate:     bbs.calculateChangeRate(biomeA, biomeB),
		LastUpdate:     world.Tick,
	}

	bbs.NextBoundaryID++

	// Calculate environmental gradients
	boundary.TemperatureGradient = bbs.calculateTemperatureGradient(world, biomeA, biomeB)
	boundary.MoistureGradient = bbs.calculateMoistureGradient(world, biomeA, biomeB)

	// Set ecological properties
	boundary.ResourceDensity = bbs.calculateResourceDensity(biomeA, biomeB)
	boundary.SpeciesDiversity = bbs.calculateSpeciesDiversity(biomeA, biomeB)
	boundary.CompetitionLevel = bbs.calculateCompetitionLevel(biomeA, biomeB)

	// Set trait modifiers for ecotone effects
	bbs.setBoundaryTraitModifiers(boundary, biomeA, biomeB)

	return boundary
}

// determineBoundaryType determines what type of boundary exists between biomes
func (bbs *BiomeBoundarySystem) determineBoundaryType(world *World, biomeA, biomeB BiomeType) BiomeBoundaryType {
	// Check for barrier combinations first (takes priority)
	if bbs.isBarrierCombination(biomeA, biomeB) {
		return BarrierBoundary // Special combinations that create barriers
	}

	// Get biome data
	biomeDataA := world.Biomes[biomeA]
	biomeDataB := world.Biomes[biomeB]

	// Calculate environmental differences
	tempDiff := math.Abs(biomeDataA.Temperature - biomeDataB.Temperature)
	humidityDiff := math.Abs(biomeDataA.Humidity - biomeDataB.Humidity)
	pressureDiff := math.Abs(biomeDataA.Pressure - biomeDataB.Pressure)

	totalDiff := tempDiff + humidityDiff + pressureDiff

	// Determine boundary type based on differences
	if totalDiff > 1.5 {
		return SharpBoundary // Very different biomes
	} else if totalDiff > 0.8 {
		return EcotoneZone // Moderately different, creates ecotone
	} else {
		return SoftBoundary // Similar biomes, gradual transition
	}
}

// isBarrierCombination checks if two biomes create a natural barrier
func (bbs *BiomeBoundarySystem) isBarrierCombination(biomeA, biomeB BiomeType) bool {
	barriers := map[BiomeType][]BiomeType{
		BiomeWater:     {BiomeDesert, BiomeMountain, BiomeHighAltitude},
		BiomeDeepWater: {BiomeDesert, BiomeMountain, BiomeIce, BiomeHighAltitude},
		BiomeIce:       {BiomeDesert, BiomeHotSpring, BiomeRainforest},
		BiomeMountain:  {BiomeWater, BiomeDeepWater, BiomeSwamp},
		BiomeRadiation: {BiomeRainforest, BiomeIce, BiomeTundra},
	}

	if barriersA, exists := barriers[biomeA]; exists {
		for _, barrier := range barriersA {
			if barrier == biomeB {
				return true
			}
		}
	}

	if barriersB, exists := barriers[biomeB]; exists {
		for _, barrier := range barriersB {
			if barrier == biomeA {
				return true
			}
		}
	}

	return false
}

// calculateBoundaryWidth determines the width of the transition zone
func (bbs *BiomeBoundarySystem) calculateBoundaryWidth(biomeA, biomeB BiomeType) float64 {
	baseWidth := 2.0

	// Wider boundaries for more different biomes
	if bbs.isBarrierCombination(biomeA, biomeB) {
		return baseWidth * 0.5 // Narrow barrier boundaries
	}

	// Ecotones are wider
	return baseWidth + rand.Float64()*1.5
}

// calculatePermeability determines how easily entities can cross boundaries
func (bbs *BiomeBoundarySystem) calculatePermeability(biomeA, biomeB BiomeType) float64 {
	// Base permeability
	permeability := 0.8

	// Reduce permeability for barrier combinations
	if bbs.isBarrierCombination(biomeA, biomeB) {
		permeability *= 0.3
	}

	// Aquatic boundaries are less permeable for terrestrial entities
	if (biomeA == BiomeWater || biomeA == BiomeDeepWater) &&
		(biomeB != BiomeWater && biomeB != BiomeDeepWater && biomeB != BiomeSwamp) {
		permeability *= 0.4
	}

	return math.Max(0.1, math.Min(1.0, permeability))
}

// processEntityBoundaryEffects applies boundary effects to entities
func (bbs *BiomeBoundarySystem) processEntityBoundaryEffects(world *World, tick int) {
	for _, entity := range world.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Find nearby boundaries
		nearbyBoundaries := bbs.findNearbyBoundaries(entity.Position)

		for _, boundary := range nearbyBoundaries {
			distance := math.Sqrt(math.Pow(entity.Position.X-boundary.Position.X, 2) +
				math.Pow(entity.Position.Y-boundary.Position.Y, 2))

			if distance <= boundary.Width {
				// Apply boundary effects
				bbs.applyBoundaryEffectsToEntity(entity, boundary, distance, world, tick)
			}
		}
	}
}

// applyBoundaryEffectsToEntity applies specific boundary effects to an entity
func (bbs *BiomeBoundarySystem) applyBoundaryEffectsToEntity(entity *Entity, boundary *BiomeBoundary, distance float64, world *World, tick int) {
	// Calculate effect intensity based on distance to boundary center
	intensity := 1.0 - (distance / boundary.Width)

	// Apply trait modifiers
	for trait, modifier := range boundary.TraitModifiers {
		currentValue := entity.GetTrait(trait)
		adjustment := modifier * intensity * 0.1 // Small adjustment
		entity.SetTrait(trait, currentValue+adjustment)
	}

	// Ecotone effects
	if boundary.BoundaryType == EcotoneZone {
		// Enhanced mutation rate in ecotones (evolutionary pressure)
		if rand.Float64() < bbs.EvolutionPressure*intensity*0.01 {
			entity.Mutate(0.05, 0.05)
			bbs.EvolutionEvents++
		}

		// Resource bonus in ecotones
		entity.Energy += boundary.ResourceDensity * intensity * 0.5

		// Migration bonus (increased speed)
		if entity.GetTrait("speed") > 0 {
			speedBonus := bbs.MigrationBonus * intensity
			currentSpeed := entity.GetTrait("speed")
			entity.SetTrait("speed", currentSpeed+speedBonus)
		}
	}

	// Barrier effects
	if boundary.BoundaryType == BarrierBoundary {
		// Movement penalty at barriers
		if entity.GetTrait("speed") > 0 {
			speedPenalty := (1.0 - boundary.Permeability) * intensity * 0.5
			currentSpeed := entity.GetTrait("speed")
			entity.SetTrait("speed", math.Max(0.1, currentSpeed-speedPenalty))
		}

		// Energy cost for crossing barriers
		energyCost := (1.0 - boundary.Permeability) * intensity * 2.0
		entity.Energy -= energyCost
	}
}

// Helper methods for calculations
func (bbs *BiomeBoundarySystem) calculateTemperatureGradient(world *World, biomeA, biomeB BiomeType) float64 {
	tempA := world.Biomes[biomeA].Temperature
	tempB := world.Biomes[biomeB].Temperature
	return math.Abs(tempA - tempB)
}

func (bbs *BiomeBoundarySystem) calculateMoistureGradient(world *World, biomeA, biomeB BiomeType) float64 {
	humidityA := world.Biomes[biomeA].Humidity
	humidityB := world.Biomes[biomeB].Humidity
	return math.Abs(humidityA - humidityB)
}

func (bbs *BiomeBoundarySystem) calculateStability(biomeA, biomeB BiomeType) float64 {
	// More stable boundaries between similar biomes
	if bbs.isBarrierCombination(biomeA, biomeB) {
		return 0.9 // Barriers are very stable
	}
	return 0.6 + rand.Float64()*0.3 // Base stability with variation
}

func (bbs *BiomeBoundarySystem) calculateChangeRate(biomeA, biomeB BiomeType) float64 {
	// Barriers change slowly, ecotones change more rapidly
	if bbs.isBarrierCombination(biomeA, biomeB) {
		return 0.01 // Very slow change for barriers
	}
	return 0.05 + rand.Float64()*0.05 // Moderate change rate
}

func (bbs *BiomeBoundarySystem) calculateResourceDensity(biomeA, biomeB BiomeType) float64 {
	// Ecotones typically have higher resource density
	return 0.8 + rand.Float64()*0.4 // 0.8 to 1.2
}

func (bbs *BiomeBoundarySystem) calculateSpeciesDiversity(biomeA, biomeB BiomeType) float64 {
	// Edge effect - higher diversity at boundaries
	return 1.2 + rand.Float64()*0.3 // 1.2 to 1.5
}

func (bbs *BiomeBoundarySystem) calculateCompetitionLevel(biomeA, biomeB BiomeType) float64 {
	// Higher competition in resource-rich ecotones
	return 0.7 + rand.Float64()*0.6 // 0.7 to 1.3
}

// setBoundaryTraitModifiers sets trait modifiers for boundary zones
func (bbs *BiomeBoundarySystem) setBoundaryTraitModifiers(boundary *BiomeBoundary, biomeA, biomeB BiomeType) {
	boundary.TraitModifiers = make(map[string]float64)

	// Enhanced adaptability in ecotones
	if boundary.BoundaryType == EcotoneZone {
		boundary.TraitModifiers["adaptability"] = 0.2
		boundary.TraitModifiers["intelligence"] = 0.1
		boundary.TraitModifiers["vision"] = 0.15
	}

	// Endurance benefits at harsh boundaries
	if bbs.isBarrierCombination(biomeA, biomeB) {
		boundary.TraitModifiers["endurance"] = 0.3
		boundary.TraitModifiers["strength"] = 0.1
	}

	// Movement adaptations based on biome types
	if biomeA == BiomeWater || biomeB == BiomeWater {
		boundary.TraitModifiers["aquatic_adaptation"] = 0.25
	}

	if biomeA == BiomeMountain || biomeB == BiomeMountain {
		boundary.TraitModifiers["agility"] = 0.2
	}
}

// findNearbyBoundaries finds boundaries near a given position
func (bbs *BiomeBoundarySystem) findNearbyBoundaries(pos Position) []*BiomeBoundary {
	nearby := make([]*BiomeBoundary, 0)

	for _, boundary := range bbs.Boundaries {
		distance := math.Sqrt(math.Pow(pos.X-boundary.Position.X, 2) +
			math.Pow(pos.Y-boundary.Position.Y, 2))

		if distance <= boundary.Width+bbs.DetectionRadius {
			nearby = append(nearby, boundary)
		}
	}

	return nearby
}

// processBoundaryEffectsOnPlants applies boundary effects to plants
func (bbs *BiomeBoundarySystem) processBoundaryEffectsOnPlants(world *World, tick int) {
	for _, plant := range world.AllPlants {
		if !plant.IsAlive {
			continue
		}

		nearbyBoundaries := bbs.findNearbyBoundaries(plant.Position)

		for _, boundary := range nearbyBoundaries {
			distance := math.Sqrt(math.Pow(plant.Position.X-boundary.Position.X, 2) +
				math.Pow(plant.Position.Y-boundary.Position.Y, 2))

			if distance <= boundary.Width {
				// Plants in ecotones get resource bonus
				if boundary.BoundaryType == EcotoneZone {
					intensity := 1.0 - (distance / boundary.Width)
					plant.Energy += boundary.ResourceDensity * intensity * 0.3
				}
			}
		}
	}
}

// updateBoundaryDynamics handles boundary movement and evolution
func (bbs *BiomeBoundarySystem) updateBoundaryDynamics(world *World, tick int) {
	for _, boundary := range bbs.Boundaries {
		// Boundaries can shift over time based on environmental pressures
		if rand.Float64() < boundary.ChangeRate {
			// Small boundary adjustments
			boundary.Position.X += (rand.Float64() - 0.5) * 0.1
			boundary.Position.Y += (rand.Float64() - 0.5) * 0.1

			// Keep boundaries within world bounds
			boundary.Position.X = math.Max(0, math.Min(float64(world.Config.GridWidth), boundary.Position.X))
			boundary.Position.Y = math.Max(0, math.Min(float64(world.Config.GridHeight), boundary.Position.Y))
		}
	}
}

// cleanupExpiredEffects removes expired boundary effects
func (bbs *BiomeBoundarySystem) cleanupExpiredEffects(tick int) {
	active := make([]*BiomeBoundaryEffect, 0)

	for _, effect := range bbs.ActiveEffects {
		if effect.Duration > 0 {
			effect.Duration--
			active = append(active, effect)
		}
	}

	bbs.ActiveEffects = active
}

// updateStatistics updates system statistics
func (bbs *BiomeBoundarySystem) updateStatistics() {
	totalLength := 0.0
	ecotoneArea := 0.0

	for _, boundary := range bbs.Boundaries {
		totalLength += boundary.Width
		if boundary.BoundaryType == EcotoneZone {
			ecotoneArea += boundary.Width * boundary.Width
		}
	}

	bbs.TotalBoundaryLength = totalLength
	bbs.EcotoneArea = ecotoneArea
}

// GetBoundaryData returns boundary data for interfaces
func (bbs *BiomeBoundarySystem) GetBoundaryData() map[string]interface{} {
	data := make(map[string]interface{})

	data["boundary_count"] = len(bbs.Boundaries)
	data["total_boundary_length"] = bbs.TotalBoundaryLength
	data["ecotone_area"] = bbs.EcotoneArea
	data["migration_events"] = bbs.MigrationEvents
	data["evolution_events"] = bbs.EvolutionEvents
	data["evolution_pressure"] = bbs.EvolutionPressure
	data["migration_bonus"] = bbs.MigrationBonus

	// Boundary type distribution
	boundaryTypes := make(map[string]int)
	for _, boundary := range bbs.Boundaries {
		switch boundary.BoundaryType {
		case SoftBoundary:
			boundaryTypes["soft"]++
		case SharpBoundary:
			boundaryTypes["sharp"]++
		case EcotoneZone:
			boundaryTypes["ecotone"]++
		case BarrierBoundary:
			boundaryTypes["barrier"]++
		}
	}
	data["boundary_types"] = boundaryTypes

	return data
}
