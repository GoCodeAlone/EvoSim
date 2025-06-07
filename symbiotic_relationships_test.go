package main

import (
	"testing"
)

func TestSymbioticRelationshipSystem(t *testing.T) {
	system := NewSymbioticRelationshipSystem()

	if system == nil {
		t.Fatal("Failed to create symbiotic relationship system")
	}

	if len(system.Relationships) != 0 {
		t.Errorf("Expected 0 initial relationships, got %d", len(system.Relationships))
	}

	if system.NextRelationshipID != 1 {
		t.Errorf("Expected NextRelationshipID to be 1, got %d", system.NextRelationshipID)
	}
}

func TestSymbioticRelationshipCreation(t *testing.T) {
	system := NewSymbioticRelationshipSystem()

	// Create test entities with higher aggression for parasitic relationship
	entity1 := &Entity{
		ID:       1,
		IsAlive:  true,
		Position: Position{X: 0, Y: 0},
		Energy:   50,
		Traits: map[string]Trait{
			"size":         {Value: 0.9}, // Larger host
			"aggression":   {Value: 0.8}, // High aggression
			"cooperation":  {Value: 0.2},
			"intelligence": {Value: 0.5},
			"defense":      {Value: 0.3},
		},
	}

	entity2 := &Entity{
		ID:       2,
		IsAlive:  true,
		Position: Position{X: 1, Y: 1},
		Energy:   30,
		Traits: map[string]Trait{
			"size":         {Value: 0.3}, // Smaller symbiont
			"aggression":   {Value: 0.7}, // High aggression
			"cooperation":  {Value: 0.1},
			"intelligence": {Value: 0.3},
			"defense":      {Value: 0.5},
		},
	}

	// Test compatibility check
	compatibility, relType := system.checkCompatibility(entity1, entity2)

	if compatibility <= 0 {
		t.Errorf("Expected positive compatibility, got %f", compatibility)
	}

	if relType != RelationshipParasitic {
		t.Errorf("Expected parasitic relationship type, got %d", relType)
	}

	// Test relationship creation
	system.createRelationship(entity1, entity2, relType, compatibility, 0)

	if len(system.Relationships) != 1 {
		t.Errorf("Expected 1 relationship after creation, got %d", len(system.Relationships))
	}

	relationship := system.Relationships[0]
	if relationship.Type != RelationshipParasitic {
		t.Errorf("Expected parasitic relationship, got %d", relationship.Type)
	}

	if relationship.HostBenefit >= 0 {
		t.Errorf("Expected negative host benefit for parasitic relationship, got %f", relationship.HostBenefit)
	}

	if relationship.SymbiontBenefit <= 0 {
		t.Errorf("Expected positive symbiont benefit for parasitic relationship, got %f", relationship.SymbiontBenefit)
	}
}

func TestMutualisticRelationship(t *testing.T) {
	system := NewSymbioticRelationshipSystem()

	// Create entities with high cooperation and intelligence for mutualistic relationship
	entity1 := &Entity{
		ID:       1,
		IsAlive:  true,
		Position: Position{X: 0, Y: 0},
		Energy:   50,
		Traits: map[string]Trait{
			"size":         {Value: 0.5},
			"aggression":   {Value: 0.2},
			"cooperation":  {Value: 0.8},
			"intelligence": {Value: 0.7},
			"defense":      {Value: 0.3},
		},
	}

	entity2 := &Entity{
		ID:       2,
		IsAlive:  true,
		Position: Position{X: 1, Y: 1},
		Energy:   40,
		Traits: map[string]Trait{
			"size":         {Value: 0.6},
			"aggression":   {Value: 0.1},
			"cooperation":  {Value: 0.9},
			"intelligence": {Value: 0.8},
			"defense":      {Value: 0.4},
		},
	}

	compatibility, relType := system.checkCompatibility(entity1, entity2)

	if relType != RelationshipMutualistic {
		t.Errorf("Expected mutualistic relationship type, got %d", relType)
	}

	system.createRelationship(entity1, entity2, relType, compatibility, 0)
	relationship := system.Relationships[0]

	if relationship.HostBenefit <= 0 {
		t.Errorf("Expected positive host benefit for mutualistic relationship, got %f", relationship.HostBenefit)
	}

	if relationship.SymbiontBenefit <= 0 {
		t.Errorf("Expected positive symbiont benefit for mutualistic relationship, got %f", relationship.SymbiontBenefit)
	}

	if relationship.Virulence != 0 {
		t.Errorf("Expected zero virulence for mutualistic relationship, got %f", relationship.Virulence)
	}
}

func TestSymbioticRelationshipStats(t *testing.T) {
	system := NewSymbioticRelationshipSystem()

	// Create a few test relationships
	system.ActiveParasitic = 2
	system.ActiveMutualistic = 1
	system.ActiveCommensal = 1
	system.TotalRelationships = 4
	system.AverageRelationshipAge = 15.5
	system.DiseaseTransmissionRate = 0.5

	stats := system.GetSymbioticStats()

	if totalRel, ok := stats["total_relationships"].(int); !ok || totalRel != 4 {
		t.Errorf("Expected total_relationships 4, got %v", stats["total_relationships"])
	}

	if parasitic, ok := stats["active_parasitic"].(int); !ok || parasitic != 2 {
		t.Errorf("Expected active_parasitic 2, got %v", stats["active_parasitic"])
	}

	if mutualistic, ok := stats["active_mutualistic"].(int); !ok || mutualistic != 1 {
		t.Errorf("Expected active_mutualistic 1, got %v", stats["active_mutualistic"])
	}

	if relationshipTypes, ok := stats["relationship_types"].(map[string]int); ok {
		if relationshipTypes["parasitic"] != 2 {
			t.Errorf("Expected parasitic type count 2, got %d", relationshipTypes["parasitic"])
		}
	} else {
		t.Errorf("Expected relationship_types to be map[string]int")
	}
}

func TestHasRelationship(t *testing.T) {
	system := NewSymbioticRelationshipSystem()

	// Create a test relationship
	relationship := &SymbioticRelationship{
		ID:         1,
		HostID:     10,
		SymbiontID: 20,
		Type:       RelationshipParasitic,
		IsActive:   true,
	}
	system.Relationships = append(system.Relationships, relationship)

	// Test has relationship
	if !system.hasRelationship(10, 20) {
		t.Error("Expected hasRelationship to return true for existing relationship")
	}

	if !system.hasRelationship(20, 10) {
		t.Error("Expected hasRelationship to return true for reverse lookup")
	}

	if system.hasRelationship(10, 30) {
		t.Error("Expected hasRelationship to return false for non-existing relationship")
	}

	// Test parasitic relationship check
	if !system.hasParasiticRelationship(10) {
		t.Error("Expected hasParasiticRelationship to return true for host with parasitic relationship")
	}

	if system.hasParasiticRelationship(20) {
		t.Error("Expected hasParasiticRelationship to return false for symbiont (not host)")
	}
}
