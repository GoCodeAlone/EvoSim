package main

import (
	"testing"
)

func TestCasteSystem(t *testing.T) {
	// Create test entities suitable for caste system
	entities := make([]*Entity, 0)
	traitNames := []string{"intelligence", "cooperation", "strength", "aggression", "leadership", "reproductive_capability"}
	
	// Create a queen candidate
	queen := NewEntity(1, traitNames, "test_species", Position{X: 0, Y: 0})
	queen.SetTrait("intelligence", 0.8)
	queen.SetTrait("leadership", 0.7)
	queen.SetTrait("reproductive_capability", 0.9)
	queen.SetTrait("cooperation", 0.8)
	entities = append(entities, queen)

	// Create worker candidates
	for i := 2; i <= 4; i++ {
		worker := NewEntity(i, traitNames, "test_species", Position{X: float64(i * 2), Y: 0})
		worker.SetTrait("intelligence", 0.4)
		worker.SetTrait("cooperation", 0.7)
		worker.SetTrait("endurance", 0.6)
		entities = append(entities, worker)
	}

	// Create soldier candidates
	for i := 5; i <= 6; i++ {
		soldier := NewEntity(i, traitNames, "test_species", Position{X: float64(i * 2), Y: 0})
		soldier.SetTrait("aggression", 0.8)
		soldier.SetTrait("strength", 0.7)
		soldier.SetTrait("cooperation", 0.5)
		soldier.SetTrait("intelligence", 0.3) // Add minimum intelligence
		entities = append(entities, soldier)
	}

	// Create caste system
	cs := NewCasteSystem()
	nestLocation := Position{X: 0, Y: 0}
	
	// Try to form colony
	colony := cs.TryFormCasteColony(entities, nestLocation)
	
	if colony == nil {
		t.Fatal("Failed to create caste colony from suitable entities")
	}

	if colony.ColonySize != len(entities) {
		t.Errorf("Expected colony size %d, got %d", len(entities), colony.ColonySize)
	}

	if len(colony.Queens) != 1 {
		t.Errorf("Expected 1 queen, got %d", len(colony.Queens))
	}

	if colony.Queens[0].ID != queen.ID {
		t.Error("Wrong entity assigned as queen")
	}

	// Check caste distribution
	if colony.CasteDistribution[Queen] != 1 {
		t.Errorf("Expected 1 queen in distribution, got %d", colony.CasteDistribution[Queen])
	}

	if colony.CasteDistribution[Worker] == 0 {
		t.Error("Expected workers in colony")
	}

	if colony.CasteDistribution[Soldier] == 0 {
		t.Error("Expected soldiers in colony")
	}

	// Test colony update
	cs.Update(nil, 100) // Mock world parameter
	
	if len(cs.Colonies) != 1 {
		t.Errorf("Expected 1 colony after update, got %d", len(cs.Colonies))
	}
}

func TestCasteRoleAssignment(t *testing.T) {
	// Create colony
	queen := NewEntity(1, []string{"intelligence", "leadership", "reproductive_capability"}, "test", Position{})
	queen.SetTrait("intelligence", 0.8)
	queen.SetTrait("leadership", 0.7)
	queen.SetTrait("reproductive_capability", 0.9)
	
	colony := NewCasteColony(1, queen, Position{X: 0, Y: 0})

	// Test role assignment for different entity types
	testCases := []struct {
		traits           map[string]float64
		expectedRole     CasteRole
		description      string
	}{
		{
			traits:       map[string]float64{"aggression": 0.8, "strength": 0.7},
			expectedRole: Soldier,
			description:  "high aggression and strength should become soldier",
		},
		{
			traits:       map[string]float64{"speed": 0.8, "intelligence": 0.6},
			expectedRole: Scout,
			description:  "high speed and intelligence should become scout",
		},
		{
			traits:       map[string]float64{"construction_skill": 0.7, "intelligence": 0.5},
			expectedRole: Builder,
			description:  "high construction skill should become builder",
		},
		{
			traits:       map[string]float64{"cooperation": 0.8, "nurturing": 0.7},
			expectedRole: Nurse,
			description:  "high cooperation and nurturing should become nurse",
		},
		{
			traits:       map[string]float64{"intelligence": 0.9},
			expectedRole: Specialist,
			description:  "very high intelligence should become specialist",
		},
		{
			traits:       map[string]float64{"cooperation": 0.5, "intelligence": 0.3},
			expectedRole: Worker,
			description:  "average traits should become worker",
		},
	}

	for i, tc := range testCases {
		entity := NewEntity(i+10, []string{}, "test", Position{})
		for trait, value := range tc.traits {
			entity.SetTrait(trait, value)
		}

		role := colony.DetermineOptimalRole(entity)
		if role != tc.expectedRole {
			t.Errorf("Test case %d (%s): expected role %s, got %s", 
				i, tc.description, tc.expectedRole.String(), role.String())
		}
	}
}

func TestCasteStatusCreation(t *testing.T) {
	// Test caste status creation for different roles
	roles := []CasteRole{Queen, Worker, Soldier, Drone, Scout, Nurse, Builder, Specialist}
	
	for _, role := range roles {
		status := NewCasteStatus(role)
		
		if status.Role != role {
			t.Errorf("Expected role %s, got %s", role.String(), status.Role.String())
		}

		// Check role-specific properties
		switch role {
		case Queen:
			if status.ReproductiveCapability != 3.0 {
				t.Errorf("Queen should have reproductive capability 3.0, got %.1f", status.ReproductiveCapability)
			}
			if status.CanChangeRole {
				t.Error("Queen should not be able to change role")
			}
		case Worker:
			if status.ReproductiveCapability != 0.1 {
				t.Errorf("Worker should have reproductive capability 0.1, got %.1f", status.ReproductiveCapability)
			}
			if !status.CanChangeRole {
				t.Error("Worker should be able to change role")
			}
		case Soldier:
			if status.ReproductiveCapability != 0.2 {
				t.Errorf("Soldier should have reproductive capability 0.2, got %.1f", status.ReproductiveCapability)
			}
		}

		// All castes should have some specialization and loyalty
		if status.RoleSpecialization <= 0 {
			t.Errorf("Role specialization should be positive, got %.2f", status.RoleSpecialization)
		}
		if status.CasteLoyalty <= 0 {
			t.Errorf("Caste loyalty should be positive, got %.2f", status.CasteLoyalty)
		}
	}
}

func TestCasteTraitModification(t *testing.T) {
	// Create test colony
	queen := NewEntity(1, []string{"intelligence", "leadership"}, "test", Position{})
	queen.SetTrait("intelligence", 0.5)
	queen.SetTrait("leadership", 0.5)
	
	colony := NewCasteColony(1, queen, Position{})

	// Test trait modification for worker
	worker := NewEntity(2, []string{"foraging_efficiency", "endurance", "cooperation"}, "test", Position{})
	worker.SetTrait("foraging_efficiency", 0.3)
	worker.SetTrait("endurance", 0.3)
	worker.SetTrait("cooperation", 0.3)
	
	originalForaging := worker.GetTrait("foraging_efficiency")
	originalEndurance := worker.GetTrait("endurance")
	originalCooperation := worker.GetTrait("cooperation")

	colony.modifyTraitsForRole(worker, Worker)

	// Check that traits were enhanced
	if worker.GetTrait("foraging_efficiency") <= originalForaging {
		t.Error("Worker foraging efficiency should be enhanced")
	}
	if worker.GetTrait("endurance") <= originalEndurance {
		t.Error("Worker endurance should be enhanced")
	}
	if worker.GetTrait("cooperation") <= originalCooperation {
		t.Error("Worker cooperation should be enhanced")
	}

	// Test trait modification for soldier
	soldier := NewEntity(3, []string{"aggression", "strength", "defense"}, "test", Position{})
	soldier.SetTrait("aggression", 0.3)
	soldier.SetTrait("strength", 0.3)
	soldier.SetTrait("defense", 0.3)
	
	originalAggression := soldier.GetTrait("aggression")
	originalStrength := soldier.GetTrait("strength")
	originalDefense := soldier.GetTrait("defense")

	colony.modifyTraitsForRole(soldier, Soldier)

	// Check that traits were enhanced
	if soldier.GetTrait("aggression") <= originalAggression {
		t.Error("Soldier aggression should be enhanced")
	}
	if soldier.GetTrait("strength") <= originalStrength {
		t.Error("Soldier strength should be enhanced")
	}
	if soldier.GetTrait("defense") <= originalDefense {
		t.Error("Soldier defense should be enhanced")
	}
}

func TestColonyMemberManagement(t *testing.T) {
	// Create test colony
	queen := NewEntity(1, []string{"intelligence"}, "test", Position{})
	colony := NewCasteColony(1, queen, Position{})

	// Test adding members
	worker1 := NewEntity(2, []string{"cooperation", "intelligence"}, "test", Position{})
	worker1.SetTrait("cooperation", 0.6)
	worker1.SetTrait("intelligence", 0.4)

	worker2 := NewEntity(3, []string{"cooperation", "intelligence"}, "test", Position{})
	worker2.SetTrait("cooperation", 0.7)
	worker2.SetTrait("intelligence", 0.5)

	if !colony.AddMember(worker1, Worker) {
		t.Error("Should be able to add suitable worker")
	}
	if !colony.AddMember(worker2, Worker) {
		t.Error("Should be able to add another suitable worker")
	}

	if colony.ColonySize != 3 { // Queen + 2 workers
		t.Errorf("Expected colony size 3, got %d", colony.ColonySize)
	}

	// Test removing member
	colony.RemoveMember(worker1)
	
	if colony.ColonySize != 2 {
		t.Errorf("Expected colony size 2 after removal, got %d", colony.ColonySize)
	}

	// Check that caste distribution was updated
	if colony.CasteDistribution[Worker] != 1 {
		t.Errorf("Expected 1 worker after removal, got %d", colony.CasteDistribution[Worker])
	}

	// Test that removed entity no longer has tribe markings
	if worker1.TribeID != 0 {
		t.Error("Removed entity should not have tribe ID")
	}
}

func TestColonyCanJoinChecks(t *testing.T) {
	// Create test colony
	queen := NewEntity(1, []string{"intelligence", "cooperation"}, "test_species", Position{})
	queen.SetTrait("intelligence", 0.8)
	queen.SetTrait("cooperation", 0.7)
	colony := NewCasteColony(1, queen, Position{})

	// Test suitable candidate
	suitable := NewEntity(2, []string{"cooperation", "intelligence"}, "test_species", Position{})
	suitable.SetTrait("cooperation", 0.6)
	suitable.SetTrait("intelligence", 0.4)

	if !colony.CanJoinColony(suitable) {
		t.Error("Suitable entity should be able to join colony")
	}

	// Test entity with low cooperation
	lowCooperation := NewEntity(3, []string{"cooperation", "intelligence"}, "test_species", Position{})
	lowCooperation.SetTrait("cooperation", 0.2) // Too low
	lowCooperation.SetTrait("intelligence", 0.5)

	if colony.CanJoinColony(lowCooperation) {
		t.Error("Entity with low cooperation should not be able to join")
	}

	// Test entity with low intelligence
	lowIntelligence := NewEntity(4, []string{"cooperation", "intelligence"}, "test_species", Position{})
	lowIntelligence.SetTrait("cooperation", 0.6)
	lowIntelligence.SetTrait("intelligence", 0.1) // Too low

	if colony.CanJoinColony(lowIntelligence) {
		t.Error("Entity with low intelligence should not be able to join")
	}

	// Test different species (should require higher traits)
	differentSpecies := NewEntity(5, []string{"cooperation", "intelligence"}, "other_species", Position{})
	differentSpecies.SetTrait("cooperation", 0.6) // Not high enough for cross-species
	differentSpecies.SetTrait("intelligence", 0.6)

	if colony.CanJoinColony(differentSpecies) {
		t.Error("Different species with moderate traits should not be able to join")
	}

	// Test different species with high traits
	differentSpeciesHigh := NewEntity(6, []string{"cooperation", "intelligence"}, "other_species", Position{})
	differentSpeciesHigh.SetTrait("cooperation", 0.9) // High enough for cross-species
	differentSpeciesHigh.SetTrait("intelligence", 0.8)

	if !colony.CanJoinColony(differentSpeciesHigh) {
		t.Error("Different species with high traits should be able to join")
	}

	// Test dead entity
	deadEntity := NewEntity(7, []string{"cooperation", "intelligence"}, "test_species", Position{})
	deadEntity.SetTrait("cooperation", 0.8)
	deadEntity.SetTrait("intelligence", 0.7)
	deadEntity.IsAlive = false

	if colony.CanJoinColony(deadEntity) {
		t.Error("Dead entity should not be able to join colony")
	}
}

func TestColonyFitnessCalculation(t *testing.T) {
	// Create test colony with known fitness values
	queen := NewEntity(1, []string{}, "test", Position{})
	queen.Fitness = 0.8
	colony := NewCasteColony(1, queen, Position{})

	worker1 := NewEntity(2, []string{}, "test", Position{})
	worker1.Fitness = 0.6
	colony.AddMember(worker1, Worker)

	worker2 := NewEntity(3, []string{}, "test", Position{})
	worker2.Fitness = 0.7
	colony.AddMember(worker2, Worker)

	// Update colony fitness
	colony.updateColonyFitness()

	expectedAvgFitness := (0.8 + 0.6 + 0.7) / 3.0 // 0.7
	
	// Colony fitness should be at least the average (may be higher due to distribution bonus)
	if colony.ColonyFitness < expectedAvgFitness {
		t.Errorf("Expected colony fitness at least %.2f, got %.2f", 
			expectedAvgFitness, colony.ColonyFitness)
	}
}

func TestCasteRoleReassignment(t *testing.T) {
	// Create colony with suboptimal role assignments
	queen := NewEntity(1, []string{}, "test", Position{})
	colony := NewCasteColony(1, queen, Position{})

	// Add entity as worker but with soldier traits
	entity := NewEntity(2, []string{"aggression", "strength", "cooperation", "intelligence"}, "test", Position{})
	entity.SetTrait("aggression", 0.9) // High aggression
	entity.SetTrait("strength", 0.8)   // High strength  
	entity.SetTrait("cooperation", 0.5)
	entity.SetTrait("intelligence", 0.4)
	
	// Force assignment as worker (suboptimal)
	colony.AddMember(entity, Worker)
	
	if entity.CasteStatus.Role != Worker {
		t.Error("Entity should initially be assigned as worker")
	}

	// Set low role efficiency to trigger reassignment
	entity.CasteStatus.RoleEfficiency = 0.3

	// Trigger role reassignment
	colony.reassignRoles()

	// Entity should now be reassigned to soldier (optimal role)
	if entity.CasteStatus.Role != Soldier {
		t.Errorf("Entity should be reassigned to soldier, got %s", entity.CasteStatus.Role.String())
	}
}

func TestAddCasteStatusToEntity(t *testing.T) {
	// Test automatic caste assignment based on traits
	
	// Queen candidate
	queenCandidate := NewEntity(1, []string{"intelligence", "leadership"}, "test", Position{})
	queenCandidate.SetTrait("intelligence", 0.8)
	queenCandidate.SetTrait("leadership", 0.6)
	
	AddCasteStatusToEntity(queenCandidate)
	
	if queenCandidate.CasteStatus == nil {
		t.Error("Caste status should be added to entity")
	}
	if queenCandidate.CasteStatus.Role != Queen {
		t.Errorf("Entity with high intelligence and leadership should be assigned Queen role, got %s", 
			queenCandidate.CasteStatus.Role.String())
	}

	// Soldier candidate
	soldierCandidate := NewEntity(2, []string{"aggression", "strength"}, "test", Position{})
	soldierCandidate.SetTrait("aggression", 0.7)
	soldierCandidate.SetTrait("strength", 0.6)
	
	AddCasteStatusToEntity(soldierCandidate)
	
	if soldierCandidate.CasteStatus.Role != Soldier {
		t.Errorf("Entity with high aggression and strength should be assigned Soldier role, got %s", 
			soldierCandidate.CasteStatus.Role.String())
	}

	// Scout candidate
	scoutCandidate := NewEntity(3, []string{"speed", "intelligence"}, "test", Position{})
	scoutCandidate.SetTrait("speed", 0.7)
	scoutCandidate.SetTrait("intelligence", 0.5)
	
	AddCasteStatusToEntity(scoutCandidate)
	
	if scoutCandidate.CasteStatus.Role != Scout {
		t.Errorf("Entity with high speed and intelligence should be assigned Scout role, got %s", 
			scoutCandidate.CasteStatus.Role.String())
	}

	// Worker (default)
	workerCandidate := NewEntity(4, []string{"cooperation"}, "test", Position{})
	workerCandidate.SetTrait("cooperation", 0.5)
	
	AddCasteStatusToEntity(workerCandidate)
	
	if workerCandidate.CasteStatus.Role != Worker {
		t.Errorf("Entity with average traits should be assigned Worker role, got %s", 
			workerCandidate.CasteStatus.Role.String())
	}
}