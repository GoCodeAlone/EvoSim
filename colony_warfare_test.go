package main

import (
	"testing"
)

func TestColonyWarfareSystemCreation(t *testing.T) {
	system := NewColonyWarfareSystem()

	if system == nil {
		t.Fatal("Failed to create ColonyWarfareSystem")
	}

	// Check initial state
	if len(system.ColonyDiplomacies) != 0 {
		t.Errorf("Expected 0 initial colony diplomacies, got %d", len(system.ColonyDiplomacies))
	}

	if len(system.ActiveConflicts) != 0 {
		t.Errorf("Expected 0 initial active conflicts, got %d", len(system.ActiveConflicts))
	}

	if system.BorderConflictChance != 0.05 {
		t.Errorf("Expected border conflict chance of 0.05, got %f", system.BorderConflictChance)
	}
}

func TestColonyRegistration(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create a test colony
	colony := &CasteColony{
		ID:           1,
		ColonySize:   10,
		NestLocation: Position{X: 0, Y: 0},
		Territory:    []Position{{X: 0, Y: 0}},
	}

	// Register colony
	system.RegisterColony(colony)

	// Check that diplomacy was created
	if _, exists := system.ColonyDiplomacies[colony.ID]; !exists {
		t.Errorf("Colony diplomacy not created for colony %d", colony.ID)
	}

	// Check initial diplomatic state
	diplomacy := system.ColonyDiplomacies[colony.ID]
	if diplomacy.Reputation != 0.0 {
		t.Errorf("Expected initial reputation of 0.0, got %f", diplomacy.Reputation)
	}
}

func TestTwoColonyRegistration(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create two test colonies
	colony1 := &CasteColony{
		ID:           1,
		ColonySize:   10,
		NestLocation: Position{X: 0, Y: 0},
		Territory:    []Position{{X: 0, Y: 0}},
	}

	colony2 := &CasteColony{
		ID:           2,
		ColonySize:   15,
		NestLocation: Position{X: 10, Y: 10},
		Territory:    []Position{{X: 10, Y: 10}},
	}

	// Register both colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Check that both have diplomacy entries
	diplomacy1 := system.ColonyDiplomacies[colony1.ID]
	diplomacy2 := system.ColonyDiplomacies[colony2.ID]

	// Check that they have neutral relations with each other
	if diplomacy1.Relations[colony2.ID] != Neutral {
		t.Errorf("Expected neutral relation from colony 1 to 2, got %v", diplomacy1.Relations[colony2.ID])
	}

	if diplomacy2.Relations[colony1.ID] != Neutral {
		t.Errorf("Expected neutral relation from colony 2 to 1, got %v", diplomacy2.Relations[colony1.ID])
	}

	// Check trust levels
	if diplomacy1.TrustLevels[colony2.ID] != 0.5 {
		t.Errorf("Expected initial trust level of 0.5, got %f", diplomacy1.TrustLevels[colony2.ID])
	}
}

func TestBorderCalculation(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create colonies with adjacent territories
	colony1 := &CasteColony{
		ID:           1,
		ColonySize:   10,
		NestLocation: Position{X: 0, Y: 0},
		Territory:    []Position{{X: 0, Y: 0}, {X: 1, Y: 0}},
	}

	colony2 := &CasteColony{
		ID:           2,
		ColonySize:   15,
		NestLocation: Position{X: 3, Y: 0},
		Territory:    []Position{{X: 3, Y: 0}, {X: 2, Y: 0}},
	}

	colonies := []*CasteColony{colony1, colony2}

	// Register colonies and update borders
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)
	system.UpdateTerritoryBorders(colonies)

	// Check that borders were created
	if len(system.TerritoryBorders) == 0 {
		t.Errorf("Expected territory borders to be created between adjacent colonies")
	}

	// Find the border between the two colonies
	var border *TerritoryBorder
	for _, b := range system.TerritoryBorders {
		if (b.Colony1ID == colony1.ID && b.Colony2ID == colony2.ID) ||
			(b.Colony1ID == colony2.ID && b.Colony2ID == colony1.ID) {
			border = b
			break
		}
	}

	if border == nil {
		t.Errorf("Expected to find border between colonies 1 and 2")
		return
	}

	if len(border.BorderPoints) == 0 {
		t.Errorf("Expected border to have border points")
	}
}

func TestConflictCreation(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create two test colonies
	colony1 := &CasteColony{
		ID:            1,
		ColonySize:    20,
		NestLocation:  Position{X: 0, Y: 0},
		Territory:     []Position{{X: 0, Y: 0}},
		ColonyFitness: 0.8,
	}

	colony2 := &CasteColony{
		ID:            2,
		ColonySize:    15,
		NestLocation:  Position{X: 5, Y: 5},
		Territory:     []Position{{X: 5, Y: 5}},
		ColonyFitness: 0.7,
	}

	// Register colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Start a conflict
	conflict := system.StartConflict(colony1, colony2, BorderSkirmish, 100)

	if conflict == nil {
		t.Fatal("Failed to create conflict")
	}

	// Check conflict properties
	if conflict.Attacker != colony1.ID {
		t.Errorf("Expected attacker to be colony %d, got %d", colony1.ID, conflict.Attacker)
	}

	if conflict.Defender != colony2.ID {
		t.Errorf("Expected defender to be colony %d, got %d", colony2.ID, conflict.Defender)
	}

	if conflict.ConflictType != BorderSkirmish {
		t.Errorf("Expected conflict type BorderSkirmish, got %v", conflict.ConflictType)
	}

	// Check that diplomatic relations changed to Enemy
	diplomacy1 := system.ColonyDiplomacies[colony1.ID]
	if diplomacy1.Relations[colony2.ID] != Enemy {
		t.Errorf("Expected enemy relation after conflict start, got %v", diplomacy1.Relations[colony2.ID])
	}

	// Check that conflict was added to active conflicts
	if len(system.ActiveConflicts) != 1 {
		t.Errorf("Expected 1 active conflict, got %d", len(system.ActiveConflicts))
	}
}

func TestWarfareStats(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create test colonies and conflicts
	colony1 := &CasteColony{ID: 1, ColonySize: 10, NestLocation: Position{X: 0, Y: 0}}
	colony2 := &CasteColony{ID: 2, ColonySize: 15, NestLocation: Position{X: 5, Y: 5}}

	system.RegisterColony(colony1)
	system.RegisterColony(colony2)
	system.StartConflict(colony1, colony2, ResourceWar, 50)

	stats := system.GetWarfareStats()

	// Check expected stats are present
	expectedKeys := []string{
		"total_colonies",
		"active_conflicts",
		"total_alliances",
		"active_trade_agreements",
		"neutral_relations",
		"allied_relations",
		"enemy_relations",
		"truce_relations",
		"total_relations",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stat '%s' not found in warfare stats", key)
		}
	}

	// Check specific values
	if stats["total_colonies"].(int) != 2 {
		t.Errorf("Expected 2 total colonies, got %d", stats["total_colonies"].(int))
	}

	if stats["active_conflicts"].(int) != 1 {
		t.Errorf("Expected 1 active conflict, got %d", stats["active_conflicts"].(int))
	}

	if stats["enemy_relations"].(int) != 2 { // Each colony considers the other an enemy
		t.Errorf("Expected 2 enemy relations, got %d", stats["enemy_relations"].(int))
	}
}

func TestDiplomaticRelationImprovement(t *testing.T) {
	system := NewColonyWarfareSystem()

	// Create test colonies
	colony1 := &CasteColony{ID: 1, ColonySize: 10, NestLocation: Position{X: 0, Y: 0}}
	colony2 := &CasteColony{ID: 2, ColonySize: 15, NestLocation: Position{X: 5, Y: 5}}

	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Start with neutral relations, improve them
	diplomacy1 := system.ColonyDiplomacies[colony1.ID]
	diplomacy2 := system.ColonyDiplomacies[colony2.ID]

	// Check initial neutral state
	if diplomacy1.Relations[colony2.ID] != Neutral {
		t.Errorf("Expected initial neutral relation, got %v", diplomacy1.Relations[colony2.ID])
	}

	// Improve relations by increasing trust
	diplomacy1.TrustLevels[colony2.ID] = 0.9
	diplomacy2.TrustLevels[colony1.ID] = 0.9

	// Call improvement function
	system.improveRelations(colony1, colony2, 100)

	// Check that relations improved to Allied
	if diplomacy1.Relations[colony2.ID] != Allied {
		t.Errorf("Expected allied relation after improvement, got %v", diplomacy1.Relations[colony2.ID])
	}
}
