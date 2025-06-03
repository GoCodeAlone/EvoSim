package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Species represents a group of genetically similar plants
type Species struct {
	ID              int                `json:"id"`
	Name            string             `json:"name"`
	OriginPlantType PlantType          `json:"origin_plant_type"`
	Members         []*Plant           `json:"members"`
	FoundingTraits  map[string]float64 `json:"founding_traits"` // Average traits when species formed
	CurrentTraits   map[string]float64 `json:"current_traits"`  // Current average traits
	FormationTick   int                `json:"formation_tick"`
	FormationTime   time.Time          `json:"formation_time"`
	ExtinctionTick  int                `json:"extinction_tick"` // 0 if still alive
	ExtinctionTime  time.Time          `json:"extinction_time"`
	IsExtinct       bool               `json:"is_extinct"`
	ParentSpeciesID int                `json:"parent_species_id"` // Species this split from
	ChildSpeciesIDs []int              `json:"child_species_ids"` // Species that split from this
	TotalMembers    int                `json:"total_members"`     // Historical count
	PeakPopulation  int                `json:"peak_population"`
	PeakTick        int                `json:"peak_tick"`
}

// SpeciationEvent represents when a new species forms
type SpeciationEvent struct {
	Tick            int                `json:"tick"`
	Time            time.Time          `json:"time"`
	NewSpeciesID    int                `json:"new_species_id"`
	ParentSpeciesID int                `json:"parent_species_id"`
	SplitReason     string             `json:"split_reason"`
	GeneticDistance float64            `json:"genetic_distance"`
	MemberCount     int                `json:"member_count"`
	DivergentTraits map[string]float64 `json:"divergent_traits"`
}

// ExtinctionEvent represents when a species goes extinct
type ExtinctionEvent struct {
	Tick            int       `json:"tick"`
	Time            time.Time `json:"time"`
	SpeciesID       int       `json:"species_id"`
	SpeciesName     string    `json:"species_name"`
	ExtinctionCause string    `json:"extinction_cause"`
	FinalPopulation int       `json:"final_population"`
	Lifespan        int       `json:"lifespan"` // Ticks from formation to extinction
}

// SpeciationSystem manages plant species evolution and tracking
type SpeciationSystem struct {
	AllSpecies    map[int]*Species `json:"all_species"`
	ActiveSpecies map[int]*Species `json:"active_species"`
	NextSpeciesID int              `json:"next_species_id"`

	// Events
	SpeciationEvents []SpeciationEvent `json:"speciation_events"`
	ExtinctionEvents []ExtinctionEvent `json:"extinction_events"`

	// Configuration
	GeneticDistanceThreshold float64 `json:"genetic_distance_threshold"`
	MinPopulationForSpecies  int     `json:"min_population_for_species"`
	ExtinctionThreshold      int     `json:"extinction_threshold"` // Ticks with 0 population

	// Statistics
	TotalSpeciesFormed   int `json:"total_species_formed"`
	TotalSpeciesExtinct  int `json:"total_species_extinct"`
	MaxActiveSpecies     int `json:"max_active_species"`
	MaxActiveSpeciesTick int `json:"max_active_species_tick"`
}

// NewSpeciationSystem creates a new speciation tracking system
func NewSpeciationSystem() *SpeciationSystem {
	return &SpeciationSystem{
		AllSpecies:               make(map[int]*Species),
		ActiveSpecies:            make(map[int]*Species),
		NextSpeciesID:            1,
		SpeciationEvents:         make([]SpeciationEvent, 0),
		ExtinctionEvents:         make([]ExtinctionEvent, 0),
		GeneticDistanceThreshold: 0.4, // Fairly high threshold for clear species distinction
		MinPopulationForSpecies:  3,   // Need at least 3 individuals to form species
		ExtinctionThreshold:      100, // 100 ticks with no members = extinction
		TotalSpeciesFormed:       0,
		TotalSpeciesExtinct:      0,
		MaxActiveSpecies:         0,
		MaxActiveSpeciesTick:     0,
	}
}

// Update processes plant populations and manages species formation/extinction
func (ss *SpeciationSystem) Update(allPlants []*Plant, tick int) {
	// Clear previous assignments
	for _, species := range ss.ActiveSpecies {
		species.Members = make([]*Plant, 0)
	}

	// Group living plants by species
	ss.assignPlantsToSpecies(allPlants)

	// Look for new species formation (speciation events)
	ss.checkForSpeciation(tick)

	// Update species statistics
	ss.updateSpeciesStats(tick)

	// Check for extinctions
	ss.checkForExtinctions(tick)

	// Update system statistics
	activeCount := len(ss.ActiveSpecies)
	if activeCount > ss.MaxActiveSpecies {
		ss.MaxActiveSpecies = activeCount
		ss.MaxActiveSpeciesTick = tick
	}
}

// assignPlantsToSpecies groups plants into species based on genetic similarity
func (ss *SpeciationSystem) assignPlantsToSpecies(allPlants []*Plant) {
	unassignedPlants := make([]*Plant, 0)

	// First pass: assign plants to existing species
	for _, plant := range allPlants {
		if !plant.IsAlive {
			continue
		}

		assigned := false
		bestSpecies := (*Species)(nil)
		bestDistance := math.Inf(1)

		// Find closest species
		for _, species := range ss.ActiveSpecies {
			distance := ss.calculateGeneticDistance(plant, species.CurrentTraits)
			if distance < ss.GeneticDistanceThreshold && distance < bestDistance {
				bestDistance = distance
				bestSpecies = species
			}
		}

		if bestSpecies != nil {
			bestSpecies.Members = append(bestSpecies.Members, plant)
			assigned = true
		}

		if !assigned {
			unassignedPlants = append(unassignedPlants, plant)
		}
	}

	// Second pass: group unassigned plants into new species
	if len(unassignedPlants) >= ss.MinPopulationForSpecies {
		ss.formNewSpeciesFromUnassigned(unassignedPlants)
	}
}

// formNewSpeciesFromUnassigned creates new species from genetically similar unassigned plants
func (ss *SpeciationSystem) formNewSpeciesFromUnassigned(unassignedPlants []*Plant) {
	// Group unassigned plants by plant type first
	typeGroups := make(map[PlantType][]*Plant)
	for _, plant := range unassignedPlants {
		typeGroups[plant.Type] = append(typeGroups[plant.Type], plant)
	}

	// For each type group, form species based on genetic similarity
	for plantType, plants := range typeGroups {
		if len(plants) < ss.MinPopulationForSpecies {
			continue
		}

		// Use clustering to group genetically similar plants
		clusters := ss.clusterPlantsByGenetics(plants)

		for _, cluster := range clusters {
			if len(cluster) >= ss.MinPopulationForSpecies {
				ss.createNewSpecies(cluster, plantType, "initial_population")
			}
		}
	}
}

// clusterPlantsByGenetics groups plants by genetic similarity using simple clustering
func (ss *SpeciationSystem) clusterPlantsByGenetics(plants []*Plant) [][]*Plant {
	if len(plants) == 0 {
		return [][]*Plant{}
	}

	clusters := make([][]*Plant, 0)
	used := make(map[int]bool)

	for i, plant1 := range plants {
		if used[i] {
			continue
		}

		cluster := []*Plant{plant1}
		used[i] = true

		for j, plant2 := range plants {
			if used[j] || i == j {
				continue
			}

			distance := ss.calculateGeneticDistanceBetweenPlants(plant1, plant2)
			if distance < ss.GeneticDistanceThreshold {
				cluster = append(cluster, plant2)
				used[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// calculateGeneticDistance computes genetic distance between a plant and species average traits
func (ss *SpeciationSystem) calculateGeneticDistance(plant *Plant, speciesTraits map[string]float64) float64 {
	if len(speciesTraits) == 0 {
		return math.Inf(1)
	}

	totalDistance := 0.0
	traitCount := 0

	for traitName, speciesValue := range speciesTraits {
		plantValue := plant.GetTrait(traitName)
		distance := math.Abs(plantValue - speciesValue)
		totalDistance += distance * distance // Squared distance for Euclidean
		traitCount++
	}

	if traitCount == 0 {
		return math.Inf(1)
	}

	return math.Sqrt(totalDistance / float64(traitCount))
}

// calculateGeneticDistanceBetweenPlants computes genetic distance between two plants
func (ss *SpeciationSystem) calculateGeneticDistanceBetweenPlants(plant1, plant2 *Plant) float64 {
	totalDistance := 0.0
	traitCount := 0

	// Get all trait names from both plants
	allTraits := make(map[string]bool)
	for traitName := range plant1.Traits {
		allTraits[traitName] = true
	}
	for traitName := range plant2.Traits {
		allTraits[traitName] = true
	}

	for traitName := range allTraits {
		value1 := plant1.GetTrait(traitName)
		value2 := plant2.GetTrait(traitName)
		distance := math.Abs(value1 - value2)
		totalDistance += distance * distance
		traitCount++
	}

	if traitCount == 0 {
		return 0.0
	}

	return math.Sqrt(totalDistance / float64(traitCount))
}

// createNewSpecies forms a new species from a group of plants
func (ss *SpeciationSystem) createNewSpecies(plants []*Plant, plantType PlantType, reason string) *Species {
	species := &Species{
		ID:              ss.NextSpeciesID,
		Name:            ss.generateSpeciesName(plantType, ss.NextSpeciesID),
		OriginPlantType: plantType,
		Members:         plants,
		FoundingTraits:  ss.calculateAverageTraits(plants),
		CurrentTraits:   ss.calculateAverageTraits(plants),
		FormationTick:   0, // Will be set by caller
		FormationTime:   time.Now(),
		ExtinctionTick:  0,
		IsExtinct:       false,
		ParentSpeciesID: 0, // Will be set if this is a split
		ChildSpeciesIDs: make([]int, 0),
		TotalMembers:    len(plants),
		PeakPopulation:  len(plants),
		PeakTick:        0,
	}

	ss.AllSpecies[species.ID] = species
	ss.ActiveSpecies[species.ID] = species
	ss.NextSpeciesID++
	ss.TotalSpeciesFormed++

	return species
}

// generateSpeciesName creates a name for a new species
func (ss *SpeciationSystem) generateSpeciesName(plantType PlantType, speciesID int) string {
	config := GetPlantConfigs()[plantType]
	return fmt.Sprintf("%s-S%d", config.Name, speciesID)
}

// calculateAverageTraits computes average trait values for a group of plants
func (ss *SpeciationSystem) calculateAverageTraits(plants []*Plant) map[string]float64 {
	if len(plants) == 0 {
		return make(map[string]float64)
	}

	traitSums := make(map[string]float64)
	traitCounts := make(map[string]int)

	for _, plant := range plants {
		for traitName, trait := range plant.Traits {
			traitSums[traitName] += trait.Value
			traitCounts[traitName]++
		}
	}

	averageTraits := make(map[string]float64)
	for traitName, sum := range traitSums {
		count := traitCounts[traitName]
		if count > 0 {
			averageTraits[traitName] = sum / float64(count)
		}
	}

	return averageTraits
}

// checkForSpeciation looks for opportunities to split existing species
func (ss *SpeciationSystem) checkForSpeciation(tick int) {
	speciesToCheck := make([]*Species, 0)
	for _, species := range ss.ActiveSpecies {
		if len(species.Members) >= ss.MinPopulationForSpecies*2 { // Need enough members to split
			speciesToCheck = append(speciesToCheck, species)
		}
	}

	for _, species := range speciesToCheck {
		ss.checkSpeciesForSplit(species, tick)
	}
}

// checkSpeciesForSplit examines a species for potential splitting based on genetic divergence
func (ss *SpeciationSystem) checkSpeciesForSplit(species *Species, tick int) {
	if len(species.Members) < ss.MinPopulationForSpecies*2 {
		return
	}

	// Look for genetic clusters within the species
	clusters := ss.clusterPlantsByGenetics(species.Members)

	if len(clusters) <= 1 {
		return // No genetic divergence
	}

	// Check if any cluster is genetically distant enough to warrant speciation
	speciesAverage := ss.calculateAverageTraits(species.Members)

	for _, cluster := range clusters {
		if len(cluster) < ss.MinPopulationForSpecies {
			continue
		}

		clusterAverage := ss.calculateAverageTraits(cluster)
		distance := ss.calculateTraitDistance(speciesAverage, clusterAverage)

		if distance > ss.GeneticDistanceThreshold {
			// Create new species from this cluster
			newSpecies := ss.createNewSpecies(cluster, species.OriginPlantType, "genetic_divergence")
			newSpecies.FormationTick = tick
			newSpecies.ParentSpeciesID = species.ID

			// Update parent species
			species.ChildSpeciesIDs = append(species.ChildSpeciesIDs, newSpecies.ID)

			// Remove cluster members from original species
			ss.removeClusterFromSpecies(species, cluster)

			// Log speciation event
			event := SpeciationEvent{
				Tick:            tick,
				Time:            time.Now(),
				NewSpeciesID:    newSpecies.ID,
				ParentSpeciesID: species.ID,
				SplitReason:     "genetic_divergence",
				GeneticDistance: distance,
				MemberCount:     len(cluster),
				DivergentTraits: clusterAverage,
			}
			ss.SpeciationEvents = append(ss.SpeciationEvents, event)

			break // Only split once per update cycle
		}
	}
}

// calculateTraitDistance computes distance between two sets of average traits
func (ss *SpeciationSystem) calculateTraitDistance(traits1, traits2 map[string]float64) float64 {
	totalDistance := 0.0
	traitCount := 0

	// Compare common traits
	for traitName, value1 := range traits1 {
		if value2, exists := traits2[traitName]; exists {
			distance := math.Abs(value1 - value2)
			totalDistance += distance * distance
			traitCount++
		}
	}

	if traitCount == 0 {
		return 0.0
	}

	return math.Sqrt(totalDistance / float64(traitCount))
}

// removeClusterFromSpecies removes plants in a cluster from a species
func (ss *SpeciationSystem) removeClusterFromSpecies(species *Species, cluster []*Plant) {
	clusterIDs := make(map[int]bool)
	for _, plant := range cluster {
		clusterIDs[plant.ID] = true
	}

	newMembers := make([]*Plant, 0)
	for _, member := range species.Members {
		if !clusterIDs[member.ID] {
			newMembers = append(newMembers, member)
		}
	}

	species.Members = newMembers
}

// updateSpeciesStats updates statistics for all active species
func (ss *SpeciationSystem) updateSpeciesStats(tick int) {
	for _, species := range ss.ActiveSpecies {
		currentPop := len(species.Members)
		species.TotalMembers += currentPop

		if currentPop > species.PeakPopulation {
			species.PeakPopulation = currentPop
			species.PeakTick = tick
		}

		// Update current average traits
		if currentPop > 0 {
			species.CurrentTraits = ss.calculateAverageTraits(species.Members)
		}
	}
}

// checkForExtinctions identifies and processes species extinctions
func (ss *SpeciationSystem) checkForExtinctions(tick int) {
	extinctSpecies := make([]*Species, 0)

	for _, species := range ss.ActiveSpecies {
		if len(species.Members) == 0 {
			if species.ExtinctionTick == 0 {
				species.ExtinctionTick = tick // Start extinction timer
			} else if tick-species.ExtinctionTick >= ss.ExtinctionThreshold {
				extinctSpecies = append(extinctSpecies, species)
			}
		} else {
			species.ExtinctionTick = 0 // Reset extinction timer
		}
	}

	// Process extinctions
	for _, species := range extinctSpecies {
		ss.processExtinction(species, tick)
	}
}

// processExtinction handles the extinction of a species
func (ss *SpeciationSystem) processExtinction(species *Species, tick int) {
	species.IsExtinct = true
	species.ExtinctionTime = time.Now()

	// Log extinction event
	event := ExtinctionEvent{
		Tick:            tick,
		Time:            time.Now(),
		SpeciesID:       species.ID,
		SpeciesName:     species.Name,
		ExtinctionCause: "population_collapse",
		FinalPopulation: 0,
		Lifespan:        tick - species.FormationTick,
	}
	ss.ExtinctionEvents = append(ss.ExtinctionEvents, event)

	// Remove from active species
	delete(ss.ActiveSpecies, species.ID)
	ss.TotalSpeciesExtinct++
}

// IsReproductivelyCompatible checks if two plants can cross-pollinate based on genetic distance
func (ss *SpeciationSystem) IsReproductivelyCompatible(plant1, plant2 *Plant) bool {
	// Same plant type always compatible
	if plant1.Type == plant2.Type {
		distance := ss.calculateGeneticDistanceBetweenPlants(plant1, plant2)
		return distance < ss.GeneticDistanceThreshold*1.5 // Slightly more permissive for reproduction
	}

	// Different plant types have reduced compatibility
	distance := ss.calculateGeneticDistanceBetweenPlants(plant1, plant2)
	return distance < ss.GeneticDistanceThreshold*0.5 // Much stricter for inter-type breeding
}

// GetSpeciesStats returns statistics about the speciation system
func (ss *SpeciationSystem) GetSpeciesStats() map[string]interface{} {
	activeSpeciesCount := len(ss.ActiveSpecies)

	// Count plants by species
	plantCounts := make(map[int]int)
	for _, species := range ss.ActiveSpecies {
		plantCounts[species.ID] = len(species.Members)
	}

	// Get largest species
	largestSpeciesID := 0
	largestSpeciesSize := 0
	for speciesID, count := range plantCounts {
		if count > largestSpeciesSize {
			largestSpeciesSize = count
			largestSpeciesID = speciesID
		}
	}

	largestSpeciesName := ""
	if largestSpeciesID > 0 {
		largestSpeciesName = ss.ActiveSpecies[largestSpeciesID].Name
	}

	return map[string]interface{}{
		"active_species":             activeSpeciesCount,
		"total_species_formed":       ss.TotalSpeciesFormed,
		"total_species_extinct":      ss.TotalSpeciesExtinct,
		"max_active_species":         ss.MaxActiveSpecies,
		"max_active_species_tick":    ss.MaxActiveSpeciesTick,
		"recent_speciation_events":   len(ss.SpeciationEvents),
		"recent_extinction_events":   len(ss.ExtinctionEvents),
		"largest_species_name":       largestSpeciesName,
		"largest_species_size":       largestSpeciesSize,
		"genetic_distance_threshold": ss.GeneticDistanceThreshold,
	}
}

// GetActiveSpeciesList returns a list of active species with their populations
func (ss *SpeciationSystem) GetActiveSpeciesList() []map[string]interface{} {
	species := make([]map[string]interface{}, 0, len(ss.ActiveSpecies))

	for _, s := range ss.ActiveSpecies {
		speciesInfo := map[string]interface{}{
			"id":                 s.ID,
			"name":               s.Name,
			"origin_type":        s.OriginPlantType,
			"current_population": len(s.Members),
			"peak_population":    s.PeakPopulation,
			"formation_tick":     s.FormationTick,
			"parent_species_id":  s.ParentSpeciesID,
			"child_count":        len(s.ChildSpeciesIDs),
		}
		species = append(species, speciesInfo)
	}

	// Sort by name for consistent ordering
	sort.Slice(species, func(i, j int) bool {
		return species[i]["name"].(string) < species[j]["name"].(string)
	})

	return species
}

// GetRecentEvents returns recent speciation and extinction events
func (ss *SpeciationSystem) GetRecentEvents(maxEvents int) map[string]interface{} {
	recentSpeciations := ss.SpeciationEvents
	if len(recentSpeciations) > maxEvents {
		recentSpeciations = recentSpeciations[len(recentSpeciations)-maxEvents:]
	}

	recentExtinctions := ss.ExtinctionEvents
	if len(recentExtinctions) > maxEvents {
		recentExtinctions = recentExtinctions[len(recentExtinctions)-maxEvents:]
	}

	return map[string]interface{}{
		"speciation_events": recentSpeciations,
		"extinction_events": recentExtinctions,
	}
}

// calculateGeneticDistanceFromTraits computes distance between plant traits and pollen genetics
func (ss *SpeciationSystem) calculateGeneticDistanceFromTraits(plantTraits map[string]Trait, pollenGenetics map[string]Trait) float64 {
	totalDistance := 0.0
	traitCount := 0

	// Get all trait names from both sources
	allTraits := make(map[string]bool)
	for traitName := range plantTraits {
		allTraits[traitName] = true
	}
	for traitName := range pollenGenetics {
		allTraits[traitName] = true
	}

	for traitName := range allTraits {
		var value1, value2 float64

		if trait, exists := plantTraits[traitName]; exists {
			value1 = trait.Value
		}

		if trait, exists := pollenGenetics[traitName]; exists {
			value2 = trait.Value
		}

		distance := math.Abs(value1 - value2)
		totalDistance += distance * distance
		traitCount++
	}

	if traitCount == 0 {
		return 0.0
	}

	return math.Sqrt(totalDistance / float64(traitCount))
}
