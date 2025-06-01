package main

import (
	"fmt"
	"math/rand"
	"strings"
)

// SpeciesNaming manages the naming system for species in the simulation
type SpeciesNaming struct {
	// Base names for different species types
	herbivoreNames []string
	predatorNames  []string
	omnivoreNames  []string

	// Name registry to track used names and evolutionary relationships
	nameRegistry map[string]*SpeciesNameInfo
	nextID       int
}

// SpeciesNameInfo contains information about a species name
type SpeciesNameInfo struct {
	ID             int
	ScientificName string
	CommonName     string
	Species        string // herbivore, predator, omnivore
	ParentName     string // Name of the parent species (if evolved)
	Generation     int    // Evolutionary generation from original species
	FormationTick  int    // When this species was named/formed
	IsExtinct      bool   // Whether this species is extinct
}

// NewSpeciesNaming creates a new species naming system
func NewSpeciesNaming() *SpeciesNaming {
	return &SpeciesNaming{
		herbivoreNames: []string{
			"Grazer", "Browser", "Muncher", "Leafy", "Greeny", "Nibbler",
			"Tender", "Peaceful", "Gentle", "Quiet", "Swift", "Cautious",
			"Meadow", "Forest", "Garden", "Plain", "Field", "Pasture",
		},
		predatorNames: []string{
			"Hunter", "Stalker", "Prowler", "Shadow", "Razor", "Fang",
			"Claw", "Strike", "Fierce", "Savage", "Swift", "Deadly",
			"Night", "Blood", "Sharp", "Wild", "Apex", "Terror",
		},
		omnivoreNames: []string{
			"Adapt", "Flex", "Balance", "Wise", "Clever", "Versatile",
			"Mixed", "Dual", "Smart", "Curious", "Explorer", "Survivor",
			"Hybrid", "Multi", "Complex", "Varied", "Diverse", "Broad",
		},
		nameRegistry: make(map[string]*SpeciesNameInfo),
		nextID:       1,
	}
}

// GenerateSpeciesName creates a new species name based on type and evolutionary history
func (sn *SpeciesNaming) GenerateSpeciesName(species string, parentName string, generation int, tick int) string {
	var baseName string
	var namePool []string

	// Select appropriate name pool
	switch species {
	case "herbivore":
		namePool = sn.herbivoreNames
	case "predator":
		namePool = sn.predatorNames
	case "omnivore":
		namePool = sn.omnivoreNames
	default:
		namePool = []string{"Unknown", "Mystery", "Strange", "Odd"}
	}

	// If this is an evolved species (has parent), create derivative name
	if parentName != "" && generation > 0 {
		parentInfo, exists := sn.nameRegistry[parentName]
		if exists {
			// Create evolved variant of parent name
			baseCommon := strings.Split(parentInfo.CommonName, " ")[0]

			// Add evolutionary suffixes
			suffixes := []string{"II", "Neo", "Advanced", "Evolved", "Prime", "Alpha", "Beta", "Gamma"}
			suffix := suffixes[generation%len(suffixes)]

			baseName = fmt.Sprintf("%s %s", baseCommon, suffix)
		} else {
			// Fallback if parent not found
			baseName = namePool[rand.Intn(len(namePool))]
			if generation > 0 {
				baseName += fmt.Sprintf(" Gen%d", generation)
			}
		}
	} else {
		// Original species - use base names
		baseName = namePool[rand.Intn(len(namePool))]
	}

	// Create scientific name (Genus species format)
	genus := strings.Title(species[:4]) + "us" // e.g., "Herbus", "Predus", "Omnius"
	speciesName := strings.ToLower(strings.ReplaceAll(baseName, " ", ""))
	scientificName := fmt.Sprintf("%s %s", genus, speciesName)

	// Ensure uniqueness
	finalName := baseName
	counter := 1
	for sn.nameExists(finalName) {
		finalName = fmt.Sprintf("%s-%d", baseName, counter)
		counter++
	}

	// Register the name
	nameInfo := &SpeciesNameInfo{
		ID:             sn.nextID,
		ScientificName: scientificName,
		CommonName:     finalName,
		Species:        species,
		ParentName:     parentName,
		Generation:     generation,
		FormationTick:  tick,
		IsExtinct:      false,
	}

	sn.nameRegistry[finalName] = nameInfo
	sn.nextID++

	return finalName
}

// GetSpeciesInfo returns information about a named species
func (sn *SpeciesNaming) GetSpeciesInfo(name string) *SpeciesNameInfo {
	return sn.nameRegistry[name]
}

// GetAllSpecies returns all registered species
func (sn *SpeciesNaming) GetAllSpecies() map[string]*SpeciesNameInfo {
	return sn.nameRegistry
}

// GetEvolutionaryLineage returns the evolutionary chain for a species
func (sn *SpeciesNaming) GetEvolutionaryLineage(name string) []string {
	lineage := []string{}
	current := name

	for current != "" {
		lineage = append([]string{current}, lineage...) // prepend
		if info, exists := sn.nameRegistry[current]; exists {
			current = info.ParentName
		} else {
			break
		}
	}

	return lineage
}

// MarkExtinct marks a species as extinct
func (sn *SpeciesNaming) MarkExtinct(name string) {
	if info, exists := sn.nameRegistry[name]; exists {
		info.IsExtinct = true
	}
}

// GetActiveSpecies returns all non-extinct species
func (sn *SpeciesNaming) GetActiveSpecies() map[string]*SpeciesNameInfo {
	active := make(map[string]*SpeciesNameInfo)
	for name, info := range sn.nameRegistry {
		if !info.IsExtinct {
			active[name] = info
		}
	}
	return active
}

// GetExtinctSpecies returns all extinct species
func (sn *SpeciesNaming) GetExtinctSpecies() map[string]*SpeciesNameInfo {
	extinct := make(map[string]*SpeciesNameInfo)
	for name, info := range sn.nameRegistry {
		if info.IsExtinct {
			extinct[name] = info
		}
	}
	return extinct
}

// GetSpeciesByType returns all species of a given type (herbivore, predator, omnivore)
func (sn *SpeciesNaming) GetSpeciesByType(speciesType string) map[string]*SpeciesNameInfo {
	result := make(map[string]*SpeciesNameInfo)
	for name, info := range sn.nameRegistry {
		if info.Species == speciesType {
			result[name] = info
		}
	}
	return result
}

// nameExists checks if a name is already registered
func (sn *SpeciesNaming) nameExists(name string) bool {
	_, exists := sn.nameRegistry[name]
	return exists
}

// GetSpeciesStats returns statistics about the naming system
func (sn *SpeciesNaming) GetSpeciesStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalSpecies := len(sn.nameRegistry)
	activeSpecies := len(sn.GetActiveSpecies())
	extinctSpecies := len(sn.GetExtinctSpecies())

	stats["total_species"] = totalSpecies
	stats["active_species"] = activeSpecies
	stats["extinct_species"] = extinctSpecies

	// Count by type
	typeCount := make(map[string]int)
	for _, info := range sn.nameRegistry {
		typeCount[info.Species]++
	}
	stats["species_by_type"] = typeCount

	// Evolutionary generations
	maxGeneration := 0
	for _, info := range sn.nameRegistry {
		if info.Generation > maxGeneration {
			maxGeneration = info.Generation
		}
	}
	stats["max_generation"] = maxGeneration

	return stats
}

// GetDefaultSpeciesNames returns the original species names for the simulation
func (sn *SpeciesNaming) GetDefaultSpeciesNames() map[string]string {
	defaults := make(map[string]string)

	// Generate default names for the three starting species
	herbName := sn.GenerateSpeciesName("herbivore", "", 0, 0)
	predName := sn.GenerateSpeciesName("predator", "", 0, 0)
	omnName := sn.GenerateSpeciesName("omnivore", "", 0, 0)

	defaults["herbivore"] = herbName
	defaults["predator"] = predName
	defaults["omnivore"] = omnName

	return defaults
}
