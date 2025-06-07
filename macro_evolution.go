package main

import (
	"fmt"
	"math"
	"sort"
)

// EvolutionaryEvent represents a significant evolutionary milestone
type EvolutionaryEvent struct {
	Tick         int                    `json:"tick"`
	Type         string                 `json:"type"` // "speciation", "extinction", "adaptation", "mutation_burst"
	Description  string                 `json:"description"`
	Species      string                 `json:"species"`
	AffectedIDs  []int                  `json:"affected_ids"`  // Entity IDs involved
	TraitChanges map[string]float64     `json:"trait_changes"` // Trait -> change magnitude
	Significance float64                `json:"significance"`  // 0-1, how important this event was
	Environment  map[string]interface{} `json:"environment"`   // Environmental conditions at time
}

// SpeciesLineage tracks the evolutionary history of a species
type SpeciesLineage struct {
	SpeciesName    string               `json:"species_name"`
	ParentSpecies  string               `json:"parent_species"`  // What species this evolved from
	OriginTick     int                  `json:"origin_tick"`     // When this species first appeared
	ExtinctionTick int                  `json:"extinction_tick"` // When it went extinct (0 if still alive)
	PeakPopulation int                  `json:"peak_population"` // Highest population reached
	GenerationSpan int                  `json:"generation_span"` // Total generations survived
	ChildSpecies   []string             `json:"child_species"`   // Species that evolved from this one
	Adaptations    []EvolutionaryEvent  `json:"adaptations"`     // Major adaptations
	TraitEvolution map[string][]float64 `json:"trait_evolution"` // Historical trait values
	DominantTraits map[string]float64   `json:"dominant_traits"` // Current/final dominant traits
	Niches         []string             `json:"niches"`          // Ecological niches occupied
}

// PhylogeneticTree represents evolutionary relationships
type PhylogeneticTree struct {
	Root     *PhylogeneticNode   `json:"root"`
	AllNodes []*PhylogeneticNode `json:"all_nodes"`
	Depth    int                 `json:"depth"`
}

// PhylogeneticNode represents a node in the evolutionary tree
type PhylogeneticNode struct {
	SpeciesName    string              `json:"species_name"`
	Parent         *PhylogeneticNode   `json:"-"` // Avoid circular JSON
	Children       []*PhylogeneticNode `json:"children"`
	BranchLength   float64             `json:"branch_length"` // Evolutionary distance
	Depth          int                 `json:"depth"`
	PopulationSize int                 `json:"population_size"`
	IsExtinct      bool                `json:"is_extinct"`
	DivergenceTime int                 `json:"divergence_time"` // When this species diverged
}

// MacroEvolutionSystem tracks large-scale evolutionary patterns
type MacroEvolutionSystem struct {
	Events           []EvolutionaryEvent        `json:"events"`
	SpeciesLineages  map[string]*SpeciesLineage `json:"species_lineages"`
	PhylogeneticTree *PhylogeneticTree          `json:"phylogenetic_tree"`
	EvolutionRates   map[string]float64         `json:"evolution_rates"` // Trait -> rate of change
	ExtinctionEvents []EvolutionaryEvent        `json:"extinction_events"`
	CurrentTick      int                        `json:"current_tick"`
	TraitHistory     map[string]map[int]float64 `json:"trait_history"` // Trait -> Tick -> Average value

	// Analysis parameters
	SignificanceThreshold float64 `json:"significance_threshold"`
	AdaptationThreshold   float64 `json:"adaptation_threshold"`
	SpeciationThreshold   float64 `json:"speciation_threshold"`
	ExtinctionThreshold   int     `json:"extinction_threshold"` // Ticks without population
}

// NewMacroEvolutionSystem creates a new macro evolution tracking system
func NewMacroEvolutionSystem() *MacroEvolutionSystem {
	return &MacroEvolutionSystem{
		Events:                make([]EvolutionaryEvent, 0),
		SpeciesLineages:       make(map[string]*SpeciesLineage),
		PhylogeneticTree:      &PhylogeneticTree{AllNodes: make([]*PhylogeneticNode, 0)},
		EvolutionRates:        make(map[string]float64),
		ExtinctionEvents:      make([]EvolutionaryEvent, 0),
		TraitHistory:          make(map[string]map[int]float64),
		SignificanceThreshold: 0.3,
		AdaptationThreshold:   0.2,
		SpeciationThreshold:   0.5,
		ExtinctionThreshold:   100,
	}
}

// UpdateMacroEvolution processes current world state and tracks evolutionary changes
func (mes *MacroEvolutionSystem) UpdateMacroEvolution(world *World) {
	mes.CurrentTick = world.Tick

	// Track trait changes over time
	mes.recordTraitHistory(world)

	// Detect new species formations
	mes.detectSpeciationEvents(world)

	// Track species extinctions
	mes.detectExtinctionEvents(world)

	// Identify significant adaptations
	mes.detectAdaptationEvents(world)

	// Update phylogenetic tree
	mes.updatePhylogeneticTree(world)

	// Calculate evolution rates
	mes.calculateEvolutionRates()

	// Clean up old data (keep only significant events)
	if world.Tick%1000 == 0 {
		mes.pruneEvents()
	}
}

// recordTraitHistory tracks average trait values over time
func (mes *MacroEvolutionSystem) recordTraitHistory(world *World) {
	speciesTraitSums := make(map[string]map[string]float64)
	speciesCounts := make(map[string]int)

	// Calculate averages for each species
	for _, entity := range world.AllEntities {
		if !entity.IsAlive {
			continue
		}

		if speciesTraitSums[entity.Species] == nil {
			speciesTraitSums[entity.Species] = make(map[string]float64)
		}

		for traitName, trait := range entity.Traits {
			speciesTraitSums[entity.Species][traitName] += trait.Value
		}
		speciesCounts[entity.Species]++
	}

	// Store averages in history
	for species, traitSums := range speciesTraitSums {
		count := speciesCounts[species]
		if count == 0 {
			continue
		}

		for traitName, sum := range traitSums {
			avg := sum / float64(count)

			if mes.TraitHistory[traitName] == nil {
				mes.TraitHistory[traitName] = make(map[int]float64)
			}

			// Store with species-specific key
			key := fmt.Sprintf("%s_%s", species, traitName)
			if mes.TraitHistory[key] == nil {
				mes.TraitHistory[key] = make(map[int]float64)
			}
			mes.TraitHistory[key][world.Tick] = avg
		}
	}
}

// detectSpeciationEvents identifies when new species form
func (mes *MacroEvolutionSystem) detectSpeciationEvents(world *World) {
	// Check for new species in speciation system
	if world.SpeciationSystem != nil {
		for _, species := range world.SpeciationSystem.ActiveSpecies {
			if _, exists := mes.SpeciesLineages[species.Name]; !exists {
				// New species detected
				mes.recordSpeciationEvent(species.Name, world)
			}
		}
	}

	// Also check for new entity species
	currentSpecies := make(map[string]bool)
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			currentSpecies[entity.Species] = true
		}
	}

	for species := range currentSpecies {
		if _, exists := mes.SpeciesLineages[species]; !exists {
			mes.recordSpeciationEvent(species, world)
		}
	}
}

// recordSpeciationEvent creates a record of a new species formation
func (mes *MacroEvolutionSystem) recordSpeciationEvent(speciesName string, world *World) {
	// Find entities of this species
	var speciesEntities []*Entity
	for _, entity := range world.AllEntities {
		if entity.IsAlive && entity.Species == speciesName {
			speciesEntities = append(speciesEntities, entity)
		}
	}

	if len(speciesEntities) == 0 {
		return
	}

	// Calculate dominant traits
	dominantTraits := make(map[string]float64)
	for traitName := range speciesEntities[0].Traits {
		sum := 0.0
		for _, entity := range speciesEntities {
			sum += entity.GetTrait(traitName)
		}
		dominantTraits[traitName] = sum / float64(len(speciesEntities))
	}

	// Try to determine parent species
	parentSpecies := mes.findParentSpecies(dominantTraits)

	// Create lineage record
	lineage := &SpeciesLineage{
		SpeciesName:    speciesName,
		ParentSpecies:  parentSpecies,
		OriginTick:     world.Tick,
		ExtinctionTick: 0,
		PeakPopulation: len(speciesEntities),
		GenerationSpan: 0,
		ChildSpecies:   make([]string, 0),
		Adaptations:    make([]EvolutionaryEvent, 0),
		TraitEvolution: make(map[string][]float64),
		DominantTraits: dominantTraits,
		Niches:         mes.identifyNiches(speciesEntities, world),
	}

	mes.SpeciesLineages[speciesName] = lineage

	// Update parent's children list
	if parentSpecies != "" {
		if parentLineage, exists := mes.SpeciesLineages[parentSpecies]; exists {
			parentLineage.ChildSpecies = append(parentLineage.ChildSpecies, speciesName)
		}
	}

	// Create evolutionary event
	event := EvolutionaryEvent{
		Tick:         world.Tick,
		Type:         "speciation",
		Description:  fmt.Sprintf("New species '%s' emerged from '%s'", speciesName, parentSpecies),
		Species:      speciesName,
		AffectedIDs:  mes.getEntityIDs(speciesEntities),
		TraitChanges: dominantTraits,
		Significance: mes.calculateSpeciationSignificance(dominantTraits, parentSpecies),
		Environment:  mes.captureEnvironmentalState(world),
	}

	mes.Events = append(mes.Events, event)
}

// findParentSpecies attempts to identify the most likely parent species
func (mes *MacroEvolutionSystem) findParentSpecies(traits map[string]float64) string {
	if len(mes.SpeciesLineages) == 0 {
		return ""
	}

	bestMatch := ""
	smallestDistance := math.Inf(1)

	for speciesName, lineage := range mes.SpeciesLineages {
		if lineage.ExtinctionTick != 0 {
			continue // Skip extinct species
		}

		distance := mes.calculateTraitDistance(traits, lineage.DominantTraits)
		if distance < smallestDistance {
			smallestDistance = distance
			bestMatch = speciesName
		}
	}

	// Only consider it a parent if the distance is reasonable
	if smallestDistance < mes.SpeciationThreshold {
		return bestMatch
	}

	return ""
}

// calculateTraitDistance computes genetic distance between trait sets
func (mes *MacroEvolutionSystem) calculateTraitDistance(traits1, traits2 map[string]float64) float64 {
	distance := 0.0
	count := 0

	for traitName, value1 := range traits1 {
		if value2, exists := traits2[traitName]; exists {
			diff := value1 - value2
			distance += diff * diff
			count++
		}
	}

	if count == 0 {
		return math.Inf(1)
	}

	return math.Sqrt(distance / float64(count))
}

// identifyNiches determines ecological niches occupied by species
func (mes *MacroEvolutionSystem) identifyNiches(entities []*Entity, world *World) []string {
	niches := make([]string, 0)

	if len(entities) == 0 {
		return niches
	}

	// Analyze traits to determine niches
	avgTraits := make(map[string]float64)
	for traitName := range entities[0].Traits {
		sum := 0.0
		for _, entity := range entities {
			sum += entity.GetTrait(traitName)
		}
		avgTraits[traitName] = sum / float64(len(entities))
	}

	// Classify based on trait profiles
	if avgTraits["aggression"] > 0.3 && avgTraits["strength"] > 0.3 {
		niches = append(niches, "predator")
	}
	if avgTraits["speed"] > 0.4 && avgTraits["aggression"] < 0.0 {
		niches = append(niches, "prey")
	}
	if avgTraits["cooperation"] > 0.3 {
		niches = append(niches, "social")
	}
	if avgTraits["intelligence"] > 0.5 {
		niches = append(niches, "intelligent")
	}
	if avgTraits["size"] > 0.5 {
		niches = append(niches, "large")
	} else if avgTraits["size"] < -0.3 {
		niches = append(niches, "small")
	}

	// Biome preferences
	biomePrefs := mes.analyzeBiomePreference(entities, world)
	niches = append(niches, biomePrefs...)

	if len(niches) == 0 {
		niches = append(niches, "generalist")
	}

	return niches
}

// analyzeBiomePreference determines preferred biomes
func (mes *MacroEvolutionSystem) analyzeBiomePreference(entities []*Entity, world *World) []string {
	biomeCounts := make(map[BiomeType]int)

	for _, entity := range entities {
		x := int(entity.Position.X)
		y := int(entity.Position.Y)

		if x >= 0 && x < len(world.Grid) && y >= 0 && y < len(world.Grid[0]) {
			biome := world.Grid[x][y].Biome
			biomeCounts[biome]++
		}
	}

	preferences := make([]string, 0)
	totalEntities := len(entities)

	for biome, count := range biomeCounts {
		if float64(count)/float64(totalEntities) > 0.3 {
			switch biome {
			case BiomePlains:
				preferences = append(preferences, "plains-dweller")
			case BiomeForest:
				preferences = append(preferences, "forest-dweller")
			case BiomeDesert:
				preferences = append(preferences, "desert-adapted")
			case BiomeMountain:
				preferences = append(preferences, "mountain-dweller")
			case BiomeWater:
				preferences = append(preferences, "aquatic")
			}
		}
	}

	return preferences
}

// detectExtinctionEvents identifies when species go extinct
func (mes *MacroEvolutionSystem) detectExtinctionEvents(world *World) {
	currentSpecies := make(map[string]bool)

	// Get currently living species
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			currentSpecies[entity.Species] = true
		}
	}

	// Check for extinctions
	for speciesName, lineage := range mes.SpeciesLineages {
		if lineage.ExtinctionTick == 0 && !currentSpecies[speciesName] {
			// Species has gone extinct
			lineage.ExtinctionTick = world.Tick

			event := EvolutionaryEvent{
				Tick:         world.Tick,
				Type:         "extinction",
				Description:  fmt.Sprintf("Species '%s' went extinct", speciesName),
				Species:      speciesName,
				AffectedIDs:  make([]int, 0),
				Significance: mes.calculateExtinctionSignificance(lineage),
				Environment:  mes.captureEnvironmentalState(world),
			}

			mes.Events = append(mes.Events, event)
			mes.ExtinctionEvents = append(mes.ExtinctionEvents, event)
		}
	}
}

// detectAdaptationEvents identifies significant evolutionary adaptations
func (mes *MacroEvolutionSystem) detectAdaptationEvents(world *World) {
	// This is a simplified version - could be much more sophisticated
	if world.Tick%100 != 0 { // Check every 100 ticks
		return
	}

	for speciesName, lineage := range mes.SpeciesLineages {
		if lineage.ExtinctionTick != 0 {
			continue // Skip extinct species
		}

		// Check for significant trait changes
		adaptations := mes.detectSignificantTraitChanges(speciesName, world)

		for _, adaptation := range adaptations {
			lineage.Adaptations = append(lineage.Adaptations, adaptation)
			mes.Events = append(mes.Events, adaptation)
		}
	}
}

// detectSignificantTraitChanges identifies major trait adaptations
func (mes *MacroEvolutionSystem) detectSignificantTraitChanges(speciesName string, world *World) []EvolutionaryEvent {
	adaptations := make([]EvolutionaryEvent, 0)

	// Get current trait averages for species
	currentTraits := mes.calculateCurrentSpeciesTraits(speciesName, world)
	if currentTraits == nil {
		return adaptations
	}

	lineage := mes.SpeciesLineages[speciesName]

	for traitName, currentValue := range currentTraits {
		originalValue := lineage.DominantTraits[traitName]
		change := math.Abs(currentValue - originalValue)

		if change > mes.AdaptationThreshold {
			// Significant adaptation detected
			adaptation := EvolutionaryEvent{
				Tick:         world.Tick,
				Type:         "adaptation",
				Description:  fmt.Sprintf("Species '%s' adapted trait '%s' (change: %.3f)", speciesName, traitName, currentValue-originalValue),
				Species:      speciesName,
				TraitChanges: map[string]float64{traitName: currentValue - originalValue},
				Significance: change,
				Environment:  mes.captureEnvironmentalState(world),
			}

			adaptations = append(adaptations, adaptation)

			// Update lineage dominant traits
			lineage.DominantTraits[traitName] = currentValue
		}
	}

	return adaptations
}

// calculateCurrentSpeciesTraits gets current average traits for a species
func (mes *MacroEvolutionSystem) calculateCurrentSpeciesTraits(speciesName string, world *World) map[string]float64 {
	var speciesEntities []*Entity
	for _, entity := range world.AllEntities {
		if entity.IsAlive && entity.Species == speciesName {
			speciesEntities = append(speciesEntities, entity)
		}
	}

	if len(speciesEntities) == 0 {
		return nil
	}

	traits := make(map[string]float64)
	for traitName := range speciesEntities[0].Traits {
		sum := 0.0
		for _, entity := range speciesEntities {
			sum += entity.GetTrait(traitName)
		}
		traits[traitName] = sum / float64(len(speciesEntities))
	}

	return traits
}

// updatePhylogeneticTree maintains the evolutionary tree structure
func (mes *MacroEvolutionSystem) updatePhylogeneticTree(world *World) {
	// Rebuild tree from lineages (simplified approach)
	nodeMap := make(map[string]*PhylogeneticNode)

	// Create nodes for all species
	for speciesName, lineage := range mes.SpeciesLineages {
		node := &PhylogeneticNode{
			SpeciesName:    speciesName,
			Children:       make([]*PhylogeneticNode, 0),
			BranchLength:   float64(world.Tick - lineage.OriginTick),
			IsExtinct:      lineage.ExtinctionTick != 0,
			DivergenceTime: lineage.OriginTick,
			PopulationSize: mes.getCurrentPopulationSize(speciesName, world),
		}
		nodeMap[speciesName] = node
		mes.PhylogeneticTree.AllNodes = append(mes.PhylogeneticTree.AllNodes, node)
	}

	// Build parent-child relationships
	for speciesName, lineage := range mes.SpeciesLineages {
		node := nodeMap[speciesName]

		if lineage.ParentSpecies != "" && nodeMap[lineage.ParentSpecies] != nil {
			parent := nodeMap[lineage.ParentSpecies]
			node.Parent = parent
			parent.Children = append(parent.Children, node)
			node.Depth = parent.Depth + 1
		} else if mes.PhylogeneticTree.Root == nil {
			// First root found
			mes.PhylogeneticTree.Root = node
			node.Depth = 0
		}
	}

	// Calculate tree depth
	maxDepth := 0
	for _, node := range mes.PhylogeneticTree.AllNodes {
		if node.Depth > maxDepth {
			maxDepth = node.Depth
		}
	}
	mes.PhylogeneticTree.Depth = maxDepth
}

// Helper functions

func (mes *MacroEvolutionSystem) getEntityIDs(entities []*Entity) []int {
	ids := make([]int, len(entities))
	for i, entity := range entities {
		ids[i] = entity.ID
	}
	return ids
}

func (mes *MacroEvolutionSystem) calculateSpeciationSignificance(traits map[string]float64, parentSpecies string) float64 {
	if parentSpecies == "" {
		return 0.5 // New root species
	}

	parent := mes.SpeciesLineages[parentSpecies]
	if parent == nil {
		return 0.5
	}

	distance := mes.calculateTraitDistance(traits, parent.DominantTraits)
	return math.Min(1.0, distance*2) // Scale to 0-1
}

func (mes *MacroEvolutionSystem) calculateExtinctionSignificance(lineage *SpeciesLineage) float64 {
	// Base significance on how long species survived and its influence
	longevityFactor := float64(lineage.ExtinctionTick-lineage.OriginTick) / 10000.0
	influenceFactor := float64(len(lineage.ChildSpecies)) * 0.2
	populationFactor := float64(lineage.PeakPopulation) / 100.0

	significance := longevityFactor + influenceFactor + populationFactor
	return math.Min(1.0, significance)
}

func (mes *MacroEvolutionSystem) captureEnvironmentalState(world *World) map[string]interface{} {
	env := make(map[string]interface{})
	env["tick"] = world.Tick
	env["total_entities"] = len(world.AllEntities)
	env["total_populations"] = len(world.Populations)

	if world.AdvancedTimeSystem != nil {
		timeState := world.AdvancedTimeSystem.GetTimeState()
		env["season"] = timeState.Season
		env["time_of_day"] = timeState.TimeOfDay
	}

	return env
}

func (mes *MacroEvolutionSystem) getCurrentPopulationSize(speciesName string, world *World) int {
	count := 0
	for _, entity := range world.AllEntities {
		if entity.IsAlive && entity.Species == speciesName {
			count++
		}
	}
	return count
}

func (mes *MacroEvolutionSystem) calculateEvolutionRates() {
	// Calculate rates of trait change over time
	for traitKey, history := range mes.TraitHistory {
		if len(history) < 2 {
			continue
		}

		// Sort ticks
		ticks := make([]int, 0, len(history))
		for tick := range history {
			ticks = append(ticks, tick)
		}
		sort.Ints(ticks)

		// Calculate rate as change per tick
		firstTick := ticks[0]
		lastTick := ticks[len(ticks)-1]
		firstValue := history[firstTick]
		lastValue := history[lastTick]

		if lastTick > firstTick {
			rate := math.Abs(lastValue-firstValue) / float64(lastTick-firstTick)
			mes.EvolutionRates[traitKey] = rate
		}
	}
}

func (mes *MacroEvolutionSystem) pruneEvents() {
	// Keep only significant events and recent events
	filteredEvents := make([]EvolutionaryEvent, 0)

	for _, event := range mes.Events {
		keep := false

		// Keep if significant
		if event.Significance >= mes.SignificanceThreshold {
			keep = true
		}

		// Keep if recent (last 1000 ticks)
		if mes.CurrentTick-event.Tick < 1000 {
			keep = true
		}

		// Always keep speciation and extinction events
		if event.Type == "speciation" || event.Type == "extinction" {
			keep = true
		}

		if keep {
			filteredEvents = append(filteredEvents, event)
		}
	}

	mes.Events = filteredEvents
}

// GetEvolutionStats returns comprehensive evolution statistics
func (mes *MacroEvolutionSystem) GetEvolutionStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_events"] = len(mes.Events)
	stats["total_species"] = len(mes.SpeciesLineages)
	stats["extinction_events"] = len(mes.ExtinctionEvents)
	stats["current_tick"] = mes.CurrentTick

	// Count living vs extinct species
	livingSpecies := 0
	extinctSpecies := 0
	for _, lineage := range mes.SpeciesLineages {
		if lineage.ExtinctionTick == 0 {
			livingSpecies++
		} else {
			extinctSpecies++
		}
	}

	stats["living_species"] = livingSpecies
	stats["extinct_species"] = extinctSpecies

	// Recent events (last 100 ticks)
	recentEvents := 0
	for _, event := range mes.Events {
		if mes.CurrentTick-event.Tick < 100 {
			recentEvents++
		}
	}
	stats["recent_events"] = recentEvents

	// Phylogenetic tree stats
	if mes.PhylogeneticTree != nil {
		stats["tree_depth"] = mes.PhylogeneticTree.Depth
		stats["tree_nodes"] = len(mes.PhylogeneticTree.AllNodes)
	}

	return stats
}

// GetRecentEvents returns the most recent evolutionary events
func (mes *MacroEvolutionSystem) GetRecentEvents(count int) []EvolutionaryEvent {
	if len(mes.Events) == 0 {
		return make([]EvolutionaryEvent, 0)
	}

	// Sort events by tick (most recent first)
	sorted := make([]EvolutionaryEvent, len(mes.Events))
	copy(sorted, mes.Events)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Tick > sorted[j].Tick
	})

	if count > len(sorted) {
		count = len(sorted)
	}

	return sorted[:count]
}

// GetSpeciesLineage returns detailed lineage information for a species
func (mes *MacroEvolutionSystem) GetSpeciesLineage(speciesName string) *SpeciesLineage {
	return mes.SpeciesLineages[speciesName]
}
