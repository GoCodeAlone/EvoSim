package main

import (
	"math"
	"math/rand"
)

// FungalOrganism represents a fungal decomposer entity
type FungalOrganism struct {
	ID                 int      `json:"id"`
	Position           Position `json:"position"`
	Species            string   `json:"species"`             // "decomposer", "mycorrhizal", "pathogenic"
	Biomass            float64  `json:"biomass"`             // Size/strength of the fungal organism
	DecompositionRate  float64  `json:"decomposition_rate"`  // How fast it breaks down organic matter
	NutrientStorage    float64  `json:"nutrient_storage"`    // Nutrients stored from decomposition
	SporeProduction    float64  `json:"spore_production"`    // Rate of spore creation for reproduction
	NetworkConnections []int    `json:"network_connections"` // Connected fungal organisms
	LastReproduced     int      `json:"last_reproduced"`     // Last tick when it reproduced
	IsAlive            bool     `json:"is_alive"`
	Age                int      `json:"age"` // Age in ticks
}

// FungalSpore represents fungal reproduction units
type FungalSpore struct {
	ID              int      `json:"id"`
	Position        Position `json:"position"`
	ParentID        int      `json:"parent_id"`
	Species         string   `json:"species"`
	GerminationTime int      `json:"germination_time"` // Ticks until germination
	Viability       float64  `json:"viability"`        // 0.0-1.0, chance of successful germination
	NutrientReq     float64  `json:"nutrient_req"`     // Nutrient requirement for germination
}

// FungalNetwork represents the network of interconnected fungi
type FungalNetwork struct {
	Organisms            []*FungalOrganism `json:"organisms"`
	Spores               []*FungalSpore    `json:"spores"`
	NextID               int               `json:"next_id"`
	TotalBiomass         float64           `json:"total_biomass"`
	TotalNutrientCycling float64           `json:"total_nutrient_cycling"` // Nutrients processed per tick
	DecompositionEvents  int               `json:"decomposition_events"`   // Total decomposition events
}

// NewFungalNetwork creates a new fungal network system
func NewFungalNetwork() *FungalNetwork {
	return &FungalNetwork{
		Organisms:            make([]*FungalOrganism, 0),
		Spores:               make([]*FungalSpore, 0),
		NextID:               1,
		TotalBiomass:         0.0,
		TotalNutrientCycling: 0.0,
		DecompositionEvents:  0,
	}
}

// Update processes the fungal network for one simulation tick
func (fn *FungalNetwork) Update(world *World, tick int) {
	// Process existing organisms
	fn.updateOrganisms(world, tick)

	// Process spore germination
	fn.updateSpores(world, tick)

	// Process decomposition of dead organic matter
	fn.processDecomposition(world, tick)

	// Calculate network statistics
	fn.updateStatistics()
}

// updateOrganisms processes all living fungal organisms
func (fn *FungalNetwork) updateOrganisms(world *World, tick int) {
	aliveOrganisms := make([]*FungalOrganism, 0)

	for _, organism := range fn.Organisms {
		if !organism.IsAlive {
			continue
		}

		organism.Age++

		// Fungi have natural lifespan
		maxAge := 500 + rand.Intn(1000) // 500-1500 tick lifespan
		if organism.Age > maxAge {
			organism.IsAlive = false
			continue
		}

		// Consume stored nutrients for survival
		maintenanceCost := organism.Biomass * 0.01 // 1% biomass maintenance cost
		if organism.NutrientStorage >= maintenanceCost {
			organism.NutrientStorage -= maintenanceCost
		} else {
			// Insufficient nutrients - reduce biomass
			organism.Biomass *= 0.95
			if organism.Biomass < 0.1 {
				organism.IsAlive = false
				continue
			}
		}

		// Grow biomass if sufficient nutrients
		if organism.NutrientStorage > organism.Biomass*2 {
			growthAmount := organism.NutrientStorage * 0.1
			organism.Biomass += growthAmount
			organism.NutrientStorage -= growthAmount * 2 // Growth requires energy
		}

		// Reproduce via spores if conditions are good
		if organism.Biomass > 5.0 && tick-organism.LastReproduced > 200 {
			fn.createSpores(organism, tick)
			organism.LastReproduced = tick
		}

		// Form network connections with nearby fungi
		fn.formNetworkConnections(organism, world)

		aliveOrganisms = append(aliveOrganisms, organism)
	}

	fn.Organisms = aliveOrganisms
}

// updateSpores processes spore germination
func (fn *FungalNetwork) updateSpores(world *World, tick int) {
	activeSpores := make([]*FungalSpore, 0)

	for _, spore := range fn.Spores {
		spore.GerminationTime--

		if spore.GerminationTime <= 0 {
			// Try to germinate
			if fn.attemptGermination(spore, world, tick) {
				// Successful germination - create new organism
				fn.createOrganismFromSpore(spore, tick)
			}
			// Spore is consumed regardless of germination success
		} else {
			activeSpores = append(activeSpores, spore)
		}
	}

	fn.Spores = activeSpores
}

// processDecomposition handles breakdown of dead organic matter
func (fn *FungalNetwork) processDecomposition(world *World, tick int) {
	if world.ReproductionSystem == nil {
		return
	}

	// Process decaying items through fungal decomposition
	for _, item := range world.ReproductionSystem.DecayingItems {
		if item.IsDecayed {
			continue
		}

		// Find nearby fungal decomposers
		nearbyFungi := fn.findNearbyDecomposers(item.Position, 5.0)

		if len(nearbyFungi) > 0 {
			// Accelerate decomposition with fungi present
			decompositionRate := 0.1 // Base rate
			for _, fungus := range nearbyFungi {
				decompositionRate += fungus.DecompositionRate * 0.1
			}

			// Process decomposition
			nutrientsReleased := item.NutrientValue * decompositionRate

			// Distribute nutrients to fungi and soil
			fungalShare := nutrientsReleased * 0.6 // Fungi get 60%
			// Soil gets remaining 40% - recorded for future integration

			// Feed nearby fungi
			for _, fungus := range nearbyFungi {
				share := fungalShare / float64(len(nearbyFungi))
				fungus.NutrientStorage += share
			}

			// Add nutrients to soil at item location
			// For now, just record that nutrients were added
			// TODO: Integrate with soil nutrient system when available

			// Reduce item nutrients
			item.NutrientValue -= nutrientsReleased
			if item.NutrientValue <= 0 {
				item.IsDecayed = true
			}

			fn.DecompositionEvents++
			fn.TotalNutrientCycling += nutrientsReleased
		}
	}
}

// findNearbyDecomposers finds decomposer fungi within range
func (fn *FungalNetwork) findNearbyDecomposers(position Position, maxDistance float64) []*FungalOrganism {
	nearby := make([]*FungalOrganism, 0)

	for _, organism := range fn.Organisms {
		if !organism.IsAlive || organism.Species != "decomposer" {
			continue
		}

		distance := math.Sqrt(math.Pow(organism.Position.X-position.X, 2) +
			math.Pow(organism.Position.Y-position.Y, 2))

		if distance <= maxDistance {
			nearby = append(nearby, organism)
		}
	}

	return nearby
}

// createSpores generates fungal spores for reproduction
func (fn *FungalNetwork) createSpores(parent *FungalOrganism, tick int) {
	sporeCount := 1 + rand.Intn(3) // 1-3 spores

	for i := 0; i < sporeCount; i++ {
		// Spores disperse randomly around parent
		angle := rand.Float64() * 2 * math.Pi
		distance := 1.0 + rand.Float64()*10.0 // 1-11 unit dispersal

		sporePos := Position{
			X: parent.Position.X + math.Cos(angle)*distance,
			Y: parent.Position.Y + math.Sin(angle)*distance,
		}

		spore := &FungalSpore{
			ID:              fn.NextID,
			Position:        sporePos,
			ParentID:        parent.ID,
			Species:         parent.Species,
			GerminationTime: 50 + rand.Intn(100),      // 50-150 ticks to germinate
			Viability:       0.3 + rand.Float64()*0.5, // 30-80% viability
			NutrientReq:     1.0 + rand.Float64()*2.0, // 1-3 nutrient requirement
		}

		fn.Spores = append(fn.Spores, spore)
		fn.NextID++
	}

	// Spore production costs nutrients
	parent.NutrientStorage -= float64(sporeCount) * 0.5
}

// attemptGermination tries to germinate a spore
func (fn *FungalNetwork) attemptGermination(spore *FungalSpore, world *World, tick int) bool {
	// Check viability
	if rand.Float64() > spore.Viability {
		return false
	}

	// Check nutrient availability at germination site
	availableNutrients := fn.getNutrientsAtPosition(spore.Position, world)
	if availableNutrients < spore.NutrientReq {
		return false
	}

	// Check for competition (too many fungi nearby)
	nearbyFungi := fn.findNearbyFungi(spore.Position, 3.0)
	if len(nearbyFungi) > 2 {
		return false // Too crowded
	}

	return true
}

// createOrganismFromSpore creates a new fungal organism from a successful spore
func (fn *FungalNetwork) createOrganismFromSpore(spore *FungalSpore, tick int) {
	organism := &FungalOrganism{
		ID:                 fn.NextID,
		Position:           spore.Position,
		Species:            spore.Species,
		Biomass:            0.5 + rand.Float64()*0.5, // Small initial biomass
		DecompositionRate:  0.1 + rand.Float64()*0.1,
		NutrientStorage:    spore.NutrientReq, // Start with required nutrients
		SporeProduction:    0.05 + rand.Float64()*0.05,
		NetworkConnections: make([]int, 0),
		LastReproduced:     tick,
		IsAlive:            true,
		Age:                0,
	}

	fn.Organisms = append(fn.Organisms, organism)
	fn.NextID++
}

// formNetworkConnections establishes fungal network links
func (fn *FungalNetwork) formNetworkConnections(organism *FungalOrganism, world *World) {
	// Limit to 5 connections per organism
	if len(organism.NetworkConnections) >= 5 {
		return
	}

	// Find nearby fungi to connect with
	nearbyFungi := fn.findNearbyFungi(organism.Position, 8.0)

	for _, nearby := range nearbyFungi {
		if nearby.ID == organism.ID {
			continue
		}

		// Check if already connected
		alreadyConnected := false
		for _, connID := range organism.NetworkConnections {
			if connID == nearby.ID {
				alreadyConnected = true
				break
			}
		}

		if !alreadyConnected && len(organism.NetworkConnections) < 5 {
			organism.NetworkConnections = append(organism.NetworkConnections, nearby.ID)
			// Make bidirectional connection
			if len(nearby.NetworkConnections) < 5 {
				nearby.NetworkConnections = append(nearby.NetworkConnections, organism.ID)
			}
		}
	}
}

// findNearbyFungi finds all fungi within range
func (fn *FungalNetwork) findNearbyFungi(position Position, maxDistance float64) []*FungalOrganism {
	nearby := make([]*FungalOrganism, 0)

	for _, organism := range fn.Organisms {
		if !organism.IsAlive {
			continue
		}

		distance := math.Sqrt(math.Pow(organism.Position.X-position.X, 2) +
			math.Pow(organism.Position.Y-position.Y, 2))

		if distance <= maxDistance {
			nearby = append(nearby, organism)
		}
	}

	return nearby
}

// getNutrientsAtPosition gets available nutrients at a position
func (fn *FungalNetwork) getNutrientsAtPosition(position Position, world *World) float64 {
	// For now, estimate nutrients from nearby decaying matter
	// TODO: Integrate with soil nutrient system when available
	nutrients := 2.0 // Base soil nutrients
	if world.ReproductionSystem != nil {
		for _, item := range world.ReproductionSystem.DecayingItems {
			if item.IsDecayed {
				continue
			}

			distance := math.Sqrt(math.Pow(item.Position.X-position.X, 2) +
				math.Pow(item.Position.Y-position.Y, 2))

			if distance <= 2.0 {
				nutrients += item.NutrientValue * 0.1 // 10% of nearby organic matter
			}
		}
	}

	return nutrients
}

// updateStatistics calculates network-wide statistics
func (fn *FungalNetwork) updateStatistics() {
	fn.TotalBiomass = 0.0

	for _, organism := range fn.Organisms {
		if organism.IsAlive {
			fn.TotalBiomass += organism.Biomass
		}
	}
}

// SeedInitialFungi creates initial fungal populations
func (fn *FungalNetwork) SeedInitialFungi(world *World, count int) {
	for i := 0; i < count; i++ {
		position := Position{
			X: rand.Float64() * float64(world.Config.GridWidth),
			Y: rand.Float64() * float64(world.Config.GridHeight),
		}

		// Create mostly decomposer fungi initially
		species := "decomposer"
		if rand.Float64() < 0.2 {
			species = "mycorrhizal" // 20% chance of beneficial fungi
		}

		organism := &FungalOrganism{
			ID:                 fn.NextID,
			Position:           position,
			Species:            species,
			Biomass:            1.0 + rand.Float64()*2.0,
			DecompositionRate:  0.1 + rand.Float64()*0.1,
			NutrientStorage:    2.0 + rand.Float64()*3.0,
			SporeProduction:    0.05 + rand.Float64()*0.05,
			NetworkConnections: make([]int, 0),
			LastReproduced:     0,
			IsAlive:            true,
			Age:                rand.Intn(100), // Varied starting ages
		}

		fn.Organisms = append(fn.Organisms, organism)
		fn.NextID++
	}
}

// GetStats returns fungal network statistics
func (fn *FungalNetwork) GetStats() map[string]interface{} {
	aliveCount := 0
	decomposerCount := 0
	mycorrhizalCount := 0
	totalConnections := 0

	for _, organism := range fn.Organisms {
		if organism.IsAlive {
			aliveCount++
			if organism.Species == "decomposer" {
				decomposerCount++
			} else if organism.Species == "mycorrhizal" {
				mycorrhizalCount++
			}
			totalConnections += len(organism.NetworkConnections)
		}
	}

	avgConnections := 0.0
	if aliveCount > 0 {
		avgConnections = float64(totalConnections) / float64(aliveCount)
	}

	return map[string]interface{}{
		"total_organisms":      aliveCount,
		"decomposer_count":     decomposerCount,
		"mycorrhizal_count":    mycorrhizalCount,
		"active_spores":        len(fn.Spores),
		"total_biomass":        fn.TotalBiomass,
		"nutrient_cycling":     fn.TotalNutrientCycling,
		"decomposition_events": fn.DecompositionEvents,
		"network_connections":  totalConnections,
		"avg_connections":      avgConnections,
	}
}
