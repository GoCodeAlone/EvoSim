package main

import (
	"math"
	"math/rand"
)

// RelationshipType represents the type of symbiotic relationship
type RelationshipType int

const (
	RelationshipParasitic RelationshipType = iota // Parasite benefits, host harmed
	RelationshipMutualistic                       // Both organisms benefit
	RelationshipCommensal                         // One benefits, other neutral
)

// SymbioticRelationship represents a symbiotic relationship between two entities
type SymbioticRelationship struct {
	ID           int              `json:"id"`
	HostID       int              `json:"host_id"`
	SymbiontID   int              `json:"symbiont_id"`
	Type         RelationshipType `json:"type"`
	Strength     float64          `json:"strength"`     // 0.0 to 1.0, intensity of relationship
	Duration     int              `json:"duration"`     // How long relationship has lasted
	HostBenefit  float64          `json:"host_benefit"` // Positive or negative effect on host
	SymbiontBenefit float64       `json:"symbiont_benefit"` // Effect on symbiont
	IsActive     bool             `json:"is_active"`
	
	// Disease/pathogen specific properties
	Virulence       float64 `json:"virulence"`        // How harmful the symbiont is
	Transmission    float64 `json:"transmission"`     // How easily it spreads
	Resistance      float64 `json:"resistance"`       // Host's resistance level
	MutationRate    float64 `json:"mutation_rate"`    // Rate of evolutionary change
}

// SymbioticRelationshipSystem manages all symbiotic relationships
type SymbioticRelationshipSystem struct {
	Relationships     []*SymbioticRelationship `json:"relationships"`
	RelationshipHistory []*SymbioticRelationship `json:"relationship_history"`
	NextRelationshipID int                     `json:"next_relationship_id"`
	
	// Configuration
	FormationRate      float64 `json:"formation_rate"`      // Rate of new relationship formation
	DissolutionRate    float64 `json:"dissolution_rate"`    // Rate of relationship breakdown
	TransmissionRadius float64 `json:"transmission_radius"` // Distance for disease transmission
	CoevolutionRate    float64 `json:"coevolution_rate"`   // Rate of host-parasite evolution
	
	// Statistics
	TotalRelationships      int     `json:"total_relationships"`
	ActiveParasitic         int     `json:"active_parasitic"`
	ActiveMutualistic       int     `json:"active_mutualistic"`
	ActiveCommensal         int     `json:"active_commensal"`
	AverageRelationshipAge  float64 `json:"average_relationship_age"`
	DiseaseTransmissionRate float64 `json:"disease_transmission_rate"`
}

// NewSymbioticRelationshipSystem creates a new symbiotic relationship system
func NewSymbioticRelationshipSystem() *SymbioticRelationshipSystem {
	return &SymbioticRelationshipSystem{
		Relationships:       make([]*SymbioticRelationship, 0),
		RelationshipHistory: make([]*SymbioticRelationship, 0),
		NextRelationshipID:  1,
		FormationRate:       0.01,  // 1% chance per tick for compatible entities
		DissolutionRate:     0.005, // 0.5% chance per tick
		TransmissionRadius:  5.0,   // 5 units radius for transmission
		CoevolutionRate:     0.002, // 0.2% evolutionary change per tick
	}
}

// Update processes all symbiotic relationships
func (srs *SymbioticRelationshipSystem) Update(world *World, tick int) {
	// Update existing relationships
	srs.updateExistingRelationships(world, tick)
	
	// Form new relationships
	srs.formNewRelationships(world, tick)
	
	// Handle disease transmission
	srs.handleDiseaseTransmission(world, tick)
	
	// Apply co-evolutionary pressure
	if tick%50 == 0 { // Every 50 ticks
		srs.applyCoevolutionaryPressure(world, tick)
	}
	
	// Update statistics
	srs.updateStatistics()
}

// updateExistingRelationships processes all active relationships
func (srs *SymbioticRelationshipSystem) updateExistingRelationships(world *World, tick int) {
	for i := len(srs.Relationships) - 1; i >= 0; i-- {
		relationship := srs.Relationships[i]
		
		// Find host and symbiont entities
		host := world.findEntityByID(relationship.HostID)
		symbiont := world.findEntityByID(relationship.SymbiontID)
		
		// Remove relationship if either entity is dead or missing
		if host == nil || symbiont == nil || !host.IsAlive || !symbiont.IsAlive {
			relationship.IsActive = false
			srs.RelationshipHistory = append(srs.RelationshipHistory, relationship)
			srs.Relationships = append(srs.Relationships[:i], srs.Relationships[i+1:]...)
			continue
		}
		
		// Update relationship duration
		relationship.Duration++
		
		// Apply relationship effects
		srs.applyRelationshipEffects(relationship, host, symbiont, world, tick)
		
		// Check for relationship dissolution
		if rand.Float64() < srs.DissolutionRate {
			relationship.IsActive = false
			srs.RelationshipHistory = append(srs.RelationshipHistory, relationship)
			srs.Relationships = append(srs.Relationships[:i], srs.Relationships[i+1:]...)
		}
	}
}

// formNewRelationships creates new symbiotic relationships
func (srs *SymbioticRelationshipSystem) formNewRelationships(world *World, tick int) {
	if rand.Float64() > srs.FormationRate {
		return
	}
	
	// Find compatible entity pairs
	for i, entity1 := range world.AllEntities {
		if !entity1.IsAlive {
			continue
		}
		
		for j, entity2 := range world.AllEntities {
			if i >= j || !entity2.IsAlive {
				continue
			}
			
			// Check if they're already in a relationship
			if srs.hasRelationship(entity1.ID, entity2.ID) {
				continue
			}
			
			// Check proximity
			distance := math.Sqrt(math.Pow(entity1.Position.X-entity2.Position.X, 2) + 
								 math.Pow(entity1.Position.Y-entity2.Position.Y, 2))
			if distance > srs.TransmissionRadius {
				continue
			}
			
			// Determine relationship compatibility and type
			if compatibility, relType := srs.checkCompatibility(entity1, entity2); compatibility > 0.5 {
				srs.createRelationship(entity1, entity2, relType, compatibility, tick)
				return // Only create one relationship per update
			}
		}
	}
}

// checkCompatibility determines if two entities can form a symbiotic relationship
func (srs *SymbioticRelationshipSystem) checkCompatibility(entity1, entity2 *Entity) (float64, RelationshipType) {
	// Size difference factor (larger entities can host smaller ones)
	sizeDiff := math.Abs(entity1.Traits["size"].Value - entity2.Traits["size"].Value)
	
	// Intelligence and cooperation factors
	intel1 := entity1.Traits["intelligence"].Value
	intel2 := entity2.Traits["intelligence"].Value
	coop1 := entity1.Traits["cooperation"].Value
	coop2 := entity2.Traits["cooperation"].Value
	
	// Aggression factor
	aggr1 := entity1.Traits["aggression"].Value
	aggr2 := entity2.Traits["aggression"].Value
	
	// Determine relationship type based on traits
	avgCooperation := (coop1 + coop2) / 2
	avgAggression := (aggr1 + aggr2) / 2
	avgIntelligence := (intel1 + intel2) / 2
	
	var compatibility float64
	var relType RelationshipType
	
	if avgAggression > 0.7 && sizeDiff > 0.3 {
		// High aggression and size difference -> Parasitic
		relType = RelationshipParasitic
		compatibility = avgAggression * (sizeDiff + 0.2) * 0.8
	} else if avgCooperation > 0.6 && avgIntelligence > 0.5 {
		// High cooperation and intelligence -> Mutualistic
		relType = RelationshipMutualistic
		compatibility = avgCooperation * avgIntelligence * 0.9
	} else if avgCooperation > 0.4 {
		// Moderate cooperation -> Commensal
		relType = RelationshipCommensal
		compatibility = avgCooperation * 0.7
	} else {
		// No compatible relationship
		compatibility = 0.0
	}
	
	return compatibility, relType
}

// createRelationship creates a new symbiotic relationship
func (srs *SymbioticRelationshipSystem) createRelationship(entity1, entity2 *Entity, relType RelationshipType, strength float64, tick int) {
	// Determine host and symbiont based on size (larger is typically host)
	var host, symbiont *Entity
	if entity1.Traits["size"].Value >= entity2.Traits["size"].Value {
		host = entity1
		symbiont = entity2
	} else {
		host = entity2
		symbiont = entity1
	}
	
	relationship := &SymbioticRelationship{
		ID:         srs.NextRelationshipID,
		HostID:     host.ID,
		SymbiontID: symbiont.ID,
		Type:       relType,
		Strength:   strength,
		Duration:   0,
		IsActive:   true,
	}
	
	// Set relationship effects based on type
	switch relType {
	case RelationshipParasitic:
		relationship.HostBenefit = -strength * 0.5    // Host is harmed
		relationship.SymbiontBenefit = strength * 0.8 // Symbiont benefits significantly
		relationship.Virulence = strength * 0.6
		relationship.Transmission = strength * 0.4
		relationship.Resistance = host.Traits["defense"].Value * 0.5
		relationship.MutationRate = 0.02
		
	case RelationshipMutualistic:
		relationship.HostBenefit = strength * 0.6     // Both benefit
		relationship.SymbiontBenefit = strength * 0.6
		relationship.Virulence = 0.0
		relationship.Transmission = 0.0
		relationship.Resistance = 1.0
		relationship.MutationRate = 0.001
		
	case RelationshipCommensal:
		relationship.HostBenefit = 0.0                // Host neutral
		relationship.SymbiontBenefit = strength * 0.4 // Symbiont benefits mildly
		relationship.Virulence = 0.0
		relationship.Transmission = 0.0
		relationship.Resistance = 1.0
		relationship.MutationRate = 0.005
	}
	
	srs.NextRelationshipID++
	srs.Relationships = append(srs.Relationships, relationship)
}

// applyRelationshipEffects applies the effects of a symbiotic relationship
func (srs *SymbioticRelationshipSystem) applyRelationshipEffects(relationship *SymbioticRelationship, host, symbiont *Entity, world *World, tick int) {
	// Apply energy/health effects
	if relationship.HostBenefit != 0 {
		host.Energy = math.Max(0, host.Energy+relationship.HostBenefit*0.1)
		if relationship.HostBenefit < 0 && rand.Float64() < math.Abs(relationship.HostBenefit)*0.1 {
			// Parasitic relationship can reduce host health
			if healthTrait, exists := host.Traits["health"]; exists {
				host.Traits["health"] = Trait{Value: math.Max(0, healthTrait.Value-0.05)}
			}
		}
	}
	
	if relationship.SymbiontBenefit != 0 {
		symbiont.Energy = math.Max(0, symbiont.Energy+relationship.SymbiontBenefit*0.1)
	}
	
	// Handle disease effects for parasitic relationships
	if relationship.Type == RelationshipParasitic {
		// Reduce host fitness and reproduction
		if fertilityTrait, exists := host.Traits["fertility"]; exists {
			reducedFertility := fertilityTrait.Value * (1.0 - relationship.Virulence*0.3)
			host.Traits["fertility"] = Trait{Value: math.Max(0, reducedFertility)}
		}
		
		// Increase mutation rate in host (evolutionary pressure)
		if relationship.Duration%10 == 0 && rand.Float64() < srs.CoevolutionRate {
			// Small chance to mutate defense or health traits
			if rand.Float64() < 0.5 {
				if defenseTrait, exists := host.Traits["defense"]; exists {
					newDefense := defenseTrait.Value + (rand.Float64()-0.5)*0.1
					host.Traits["defense"] = Trait{Value: math.Max(0, math.Min(1, newDefense))}
				}
			} else {
				if healthTrait, exists := host.Traits["health"]; exists {
					newHealth := healthTrait.Value + (rand.Float64()-0.5)*0.1
					host.Traits["health"] = Trait{Value: math.Max(0, math.Min(1, newHealth))}
				}
			}
		}
	}
	
	// Handle mutualistic benefits
	if relationship.Type == RelationshipMutualistic {
		// Boost both entities' traits slightly
		if speedTrait, exists := host.Traits["speed"]; exists {
			boostedSpeed := speedTrait.Value * (1.0 + relationship.HostBenefit*0.1)
			host.Traits["speed"] = Trait{Value: math.Min(1, boostedSpeed)}
		}
		if efficiencyTrait, exists := symbiont.Traits["efficiency"]; exists {
			boostedEfficiency := efficiencyTrait.Value * (1.0 + relationship.SymbiontBenefit*0.1)
			symbiont.Traits["efficiency"] = Trait{Value: math.Min(1, boostedEfficiency)}
		}
	}
}

// handleDiseaseTransmission handles pathogen spread between entities
func (srs *SymbioticRelationshipSystem) handleDiseaseTransmission(world *World, tick int) {
	for _, relationship := range srs.Relationships {
		if relationship.Type != RelationshipParasitic || relationship.Transmission <= 0 {
			continue
		}
		
		// Find infected host
		host := world.findEntityByID(relationship.HostID)
		if host == nil || !host.IsAlive {
			continue
		}
		
		// Look for nearby susceptible entities
		for _, nearbyEntity := range world.AllEntities {
			if nearbyEntity.ID == host.ID || !nearbyEntity.IsAlive {
				continue
			}
			
			distance := math.Sqrt(math.Pow(host.Position.X-nearbyEntity.Position.X, 2) + 
								 math.Pow(host.Position.Y-nearbyEntity.Position.Y, 2))
			
			if distance <= srs.TransmissionRadius {
				// Check if entity is already infected
				if srs.hasParasiticRelationship(nearbyEntity.ID) {
					continue
				}
				
				// Calculate transmission probability
				transmissionProb := relationship.Transmission * 0.1
				nearbyResistance := nearbyEntity.Traits["defense"].Value
				finalProb := transmissionProb * (1.0 - nearbyResistance*0.5)
				
				if rand.Float64() < finalProb {
					// Create new parasitic relationship (disease transmission)
					symbiont := world.findEntityByID(relationship.SymbiontID)
					if symbiont != nil && symbiont.IsAlive {
						srs.createTransmittedRelationship(nearbyEntity, relationship, tick)
					}
				}
			}
		}
	}
}

// createTransmittedRelationship creates a new relationship via disease transmission
func (srs *SymbioticRelationshipSystem) createTransmittedRelationship(newHost *Entity, originalRelationship *SymbioticRelationship, tick int) {
	newRelationship := &SymbioticRelationship{
		ID:              srs.NextRelationshipID,
		HostID:          newHost.ID,
		SymbiontID:      originalRelationship.SymbiontID, // Same pathogen
		Type:            RelationshipParasitic,
		Strength:        originalRelationship.Strength * (0.8 + rand.Float64()*0.4), // Some variation
		Duration:        0,
		HostBenefit:     originalRelationship.HostBenefit,
		SymbiontBenefit: originalRelationship.SymbiontBenefit,
		Virulence:       originalRelationship.Virulence * (0.9 + rand.Float64()*0.2), // Evolution
		Transmission:    originalRelationship.Transmission * (0.9 + rand.Float64()*0.2),
		Resistance:      newHost.Traits["defense"].Value * 0.5,
		MutationRate:    originalRelationship.MutationRate,
		IsActive:        true,
	}
	
	srs.NextRelationshipID++
	srs.Relationships = append(srs.Relationships, newRelationship)
}

// applyCoevolutionaryPressure applies evolutionary pressure based on symbiotic relationships
func (srs *SymbioticRelationshipSystem) applyCoevolutionaryPressure(world *World, tick int) {
	for _, relationship := range srs.Relationships {
		if relationship.Type != RelationshipParasitic {
			continue
		}
		
		host := world.findEntityByID(relationship.HostID)
		symbiont := world.findEntityByID(relationship.SymbiontID)
		
		if host == nil || symbiont == nil || !host.IsAlive || !symbiont.IsAlive {
			continue
		}
		
		// Host evolution: increase resistance
		if rand.Float64() < srs.CoevolutionRate {
			if defenseTrait, exists := host.Traits["defense"]; exists {
				mutationAmount := (rand.Float64() - 0.5) * 0.2
				newDefense := defenseTrait.Value + mutationAmount
				host.Traits["defense"] = Trait{Value: math.Max(0, math.Min(1, newDefense))}
			}
		}
		
		// Symbiont evolution: adjust virulence and transmission
		if rand.Float64() < relationship.MutationRate {
			virulenceMutation := (rand.Float64() - 0.5) * 0.1
			transmissionMutation := (rand.Float64() - 0.5) * 0.1
			
			relationship.Virulence = math.Max(0, math.Min(1, relationship.Virulence+virulenceMutation))
			relationship.Transmission = math.Max(0, math.Min(1, relationship.Transmission+transmissionMutation))
		}
	}
}

// Helper functions

// hasRelationship checks if two entities already have a relationship
func (srs *SymbioticRelationshipSystem) hasRelationship(entityID1, entityID2 int) bool {
	for _, relationship := range srs.Relationships {
		if (relationship.HostID == entityID1 && relationship.SymbiontID == entityID2) ||
		   (relationship.HostID == entityID2 && relationship.SymbiontID == entityID1) {
			return true
		}
	}
	return false
}

// hasParasiticRelationship checks if an entity has a parasitic relationship (as host)
func (srs *SymbioticRelationshipSystem) hasParasiticRelationship(entityID int) bool {
	for _, relationship := range srs.Relationships {
		if relationship.HostID == entityID && relationship.Type == RelationshipParasitic {
			return true
		}
	}
	return false
}

// updateStatistics updates system statistics
func (srs *SymbioticRelationshipSystem) updateStatistics() {
	srs.TotalRelationships = len(srs.Relationships) + len(srs.RelationshipHistory)
	srs.ActiveParasitic = 0
	srs.ActiveMutualistic = 0
	srs.ActiveCommensal = 0
	
	totalAge := 0
	diseaseCount := 0
	
	for _, relationship := range srs.Relationships {
		switch relationship.Type {
		case RelationshipParasitic:
			srs.ActiveParasitic++
			diseaseCount++
		case RelationshipMutualistic:
			srs.ActiveMutualistic++
		case RelationshipCommensal:
			srs.ActiveCommensal++
		}
		totalAge += relationship.Duration
	}
	
	if len(srs.Relationships) > 0 {
		srs.AverageRelationshipAge = float64(totalAge) / float64(len(srs.Relationships))
	} else {
		srs.AverageRelationshipAge = 0
	}
	
	// Calculate disease transmission rate (diseases per total relationships)
	if srs.TotalRelationships > 0 {
		srs.DiseaseTransmissionRate = float64(diseaseCount) / float64(srs.TotalRelationships)
	} else {
		srs.DiseaseTransmissionRate = 0
	}
}

// GetSymbioticStats returns statistics about symbiotic relationships
func (srs *SymbioticRelationshipSystem) GetSymbioticStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["total_relationships"] = srs.TotalRelationships
	stats["active_relationships"] = len(srs.Relationships)
	stats["active_parasitic"] = srs.ActiveParasitic
	stats["active_mutualistic"] = srs.ActiveMutualistic
	stats["active_commensal"] = srs.ActiveCommensal
	stats["average_relationship_age"] = srs.AverageRelationshipAge
	stats["disease_transmission_rate"] = srs.DiseaseTransmissionRate
	
	// Relationship type breakdown
	relationshipTypes := make(map[string]int)
	relationshipTypes["parasitic"] = srs.ActiveParasitic
	relationshipTypes["mutualistic"] = srs.ActiveMutualistic
	relationshipTypes["commensal"] = srs.ActiveCommensal
	stats["relationship_types"] = relationshipTypes
	
	// Average virulence and transmission for parasitic relationships
	if srs.ActiveParasitic > 0 {
		totalVirulence := 0.0
		totalTransmission := 0.0
		count := 0
		
		for _, relationship := range srs.Relationships {
			if relationship.Type == RelationshipParasitic {
				totalVirulence += relationship.Virulence
				totalTransmission += relationship.Transmission
				count++
			}
		}
		
		if count > 0 {
			stats["average_virulence"] = totalVirulence / float64(count)
			stats["average_transmission"] = totalTransmission / float64(count)
		}
	} else {
		stats["average_virulence"] = 0.0
		stats["average_transmission"] = 0.0
	}
	
	return stats
}

// findEntityByID helper function to find entity by ID (should be added to world.go or used from existing method)
func (w *World) findEntityByID(id int) *Entity {
	for _, entity := range w.AllEntities {
		if entity.ID == id {
			return entity
		}
	}
	return nil
}