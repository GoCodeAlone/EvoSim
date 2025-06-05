package main

import (
	"testing"
)

// TestResourceManagementInCasteColony tests the enhanced resource management in colonies
func TestResourceManagementInCasteColony(t *testing.T) {
	// Create test entities
	queen := NewEntity(1, []string{"intelligence", "cooperation", "leadership"}, "testspecies", Position{X: 0, Y: 0})
	queen.IsAlive = true

	// Create colony
	colony := NewCasteColony(1, queen, Position{X: 0, Y: 0})

	// Test initial resources
	if len(colony.Resources) == 0 {
		t.Error("Colony should have initial resources")
	}

	// Test resource consumption
	initialFood := colony.Resources["food"]
	canAfford := colony.CanAffordResource("food", 10.0)
	if !canAfford && initialFood >= 10.0 {
		t.Error("Colony should be able to afford resource it has")
	}

	// Test resource consumption
	if colony.ConsumeResource("food", 10.0) {
		if colony.Resources["food"] >= initialFood {
			t.Error("Resource consumption should reduce stockpile")
		}
	}

	// Test resource addition
	colony.AddResource("energy", 50.0)
	if colony.Resources["energy"] < 50.0 {
		t.Error("Adding resources should increase stockpile")
	}

	// Test resource updates
	colony.UpdateResources(100)
	// Resources should have been generated and consumed
	if colony.Resources["food"] < 0 {
		t.Error("Resources should not go negative")
	}
}

// TestTradeAgreementExecution tests trade agreement processing
func TestTradeAgreementExecution(t *testing.T) {
	// Create warfare system
	system := NewColonyWarfareSystem()

	// Create test colonies
	queen1 := NewEntity(1, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 0, Y: 0})
	queen2 := NewEntity(2, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 10, Y: 10})
	colony1 := NewCasteColony(1, queen1, Position{X: 0, Y: 0})
	colony2 := NewCasteColony(2, queen2, Position{X: 10, Y: 10})

	// Register colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Set up resources for trading
	colony1.Resources["food"] = 100.0     // Colony1 has surplus food
	colony1.Resources["materials"] = 10.0 // Colony1 needs materials
	colony2.Resources["food"] = 20.0      // Colony2 needs food
	colony2.Resources["materials"] = 80.0 // Colony2 has surplus materials

	// Create trade agreement
	offeredResources := map[string]float64{"food": 10.0}
	wantedResources := map[string]float64{"materials": 5.0}

	agreement := system.CreateTradeAgreement(colony1, colony2, offeredResources, wantedResources, 1000, 0)

	if agreement == nil {
		t.Error("Trade agreement should be created")
	}

	if !agreement.IsActive {
		t.Error("Trade agreement should be active")
	}

	// Test trade execution
	initialFood1 := colony1.Resources["food"]
	initialMaterials2 := colony2.Resources["materials"]

	colonies := []*CasteColony{colony1, colony2}
	system.ProcessTradeAgreements(colonies, 10) // Process after 10 ticks

	// Check if trade occurred
	if colony1.Resources["food"] >= initialFood1 {
		t.Error("Colony1 should have less food after trading")
	}

	if colony2.Resources["materials"] >= initialMaterials2 {
		t.Error("Colony2 should have less materials after trading")
	}

	// Check if both colonies received what they wanted
	if colony1.Resources["materials"] <= 10.0 {
		t.Error("Colony1 should have received materials")
	}

	if colony2.Resources["food"] <= 20.0 {
		t.Error("Colony2 should have received food")
	}
}

// TestAllianceFormationAndBenefits tests alliance creation and functionality
func TestAllianceFormationAndBenefits(t *testing.T) {
	// Create warfare system
	system := NewColonyWarfareSystem()

	// Create test colonies
	queen1 := NewEntity(1, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 0, Y: 0})
	queen2 := NewEntity(2, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 10, Y: 10})
	colony1 := NewCasteColony(1, queen1, Position{X: 0, Y: 0})
	colony2 := NewCasteColony(2, queen2, Position{X: 10, Y: 10})

	// Register colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Set up good relations for alliance formation
	system.ColonyDiplomacies[colony1.ID].Relations[colony2.ID] = Neutral
	system.ColonyDiplomacies[colony2.ID].Relations[colony1.ID] = Neutral
	system.ColonyDiplomacies[colony1.ID].TrustLevels[colony2.ID] = 0.8
	system.ColonyDiplomacies[colony2.ID].TrustLevels[colony1.ID] = 0.8

	// Create alliance
	members := []int{colony1.ID, colony2.ID}
	alliance := system.CreateAlliance(members, "defensive", 0.2, 0)

	if alliance == nil {
		t.Error("Alliance should be created")
	}

	if !alliance.IsActive {
		t.Error("Alliance should be active")
	}

	if len(alliance.Members) != 2 {
		t.Error("Alliance should have 2 members")
	}

	// Check diplomatic relations were updated
	if system.ColonyDiplomacies[colony1.ID].Relations[colony2.ID] != Allied {
		t.Error("Colonies should be allies after alliance formation")
	}

	// Test resource sharing
	// Set up proper consumption rates first
	colony1.ResourceConsumption["food"] = 1.0
	colony2.ResourceConsumption["food"] = 1.0
	colony1.ColonySize = 10 // Size affects calculation
	colony2.ColonySize = 10

	colony1.Resources["food"] = 500.0 // Large surplus (should be > 20*1.0*10*0.1 = 20 reserve)
	colony2.Resources["food"] = 5.0   // Large need (should need 30*1.0*10*0.1 = 30 ideal)

	colonies := []*CasteColony{colony1, colony2}
	system.ProcessAlliances(colonies, 0)

	// Resource sharing should have occurred
	if colony2.Resources["food"] <= 5.0 {
		t.Errorf("Resource sharing should have increased colony2's food from 5.0 to more, got %.1f", colony2.Resources["food"])
	}

	if colony1.Resources["food"] >= 500.0 {
		t.Errorf("Resource sharing should have decreased colony1's food from 500.0, got %.1f", colony1.Resources["food"])
	}
}

// TestAutomaticTradeAndAllianceFormation tests the automatic formation systems
func TestAutomaticTradeAndAllianceFormation(t *testing.T) {
	// Create warfare system
	system := NewColonyWarfareSystem()

	// Create test colonies with complementary resources
	queen1 := NewEntity(1, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 0, Y: 0})
	queen2 := NewEntity(2, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 20, Y: 20})
	colony1 := NewCasteColony(1, queen1, Position{X: 0, Y: 0})
	colony2 := NewCasteColony(2, queen2, Position{X: 20, Y: 20})

	// Set up resources to create trade incentives
	colony1.Resources["food"] = 200.0      // Surplus food
	colony1.Resources["materials"] = 5.0   // Needs materials
	colony2.Resources["food"] = 0.0        // No food, definitely needs food
	colony2.Resources["materials"] = 150.0 // Surplus materials

	// Increase colony sizes to increase consumption needs
	colony1.ColonySize = 20
	colony2.ColonySize = 20

	// Register colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)

	// Set up neutral relations with sufficient trust
	system.ColonyDiplomacies[colony1.ID].Relations[colony2.ID] = Neutral
	system.ColonyDiplomacies[colony2.ID].Relations[colony1.ID] = Neutral
	system.ColonyDiplomacies[colony1.ID].TrustLevels[colony2.ID] = 0.6
	system.ColonyDiplomacies[colony2.ID].TrustLevels[colony1.ID] = 0.6

	colonies := []*CasteColony{colony1, colony2}

	// Test automatic trade formation
	system.AttemptAutomaticTrading(colonies, 100) // Tick 100 is divisible by 100

	// Check if trade agreement was created
	if len(system.TradeAgreements) == 0 {
		t.Error("Automatic trading should have created trade agreements")
	}

	// Increase trust for alliance formation
	system.ColonyDiplomacies[colony1.ID].TrustLevels[colony2.ID] = 0.8
	system.ColonyDiplomacies[colony2.ID].TrustLevels[colony1.ID] = 0.8

	// Add a common enemy to encourage alliance
	queen3 := NewEntity(3, []string{"aggression", "strength"}, "predator", Position{X: 50, Y: 50})
	colony3 := NewCasteColony(3, queen3, Position{X: 50, Y: 50})
	system.RegisterColony(colony3)

	// Set up enemy relations in both directions
	system.ColonyDiplomacies[colony1.ID].Relations[colony3.ID] = Enemy
	system.ColonyDiplomacies[colony2.ID].Relations[colony3.ID] = Enemy
	system.ColonyDiplomacies[colony3.ID].Relations[colony1.ID] = Enemy
	system.ColonyDiplomacies[colony3.ID].Relations[colony2.ID] = Enemy

	// Test automatic alliance formation (try multiple times due to randomness)
	for i := 0; i < 10; i++ { // Try up to 10 times
		system.AttemptAllianceFormation(colonies, 300+(i*300)) // Different tick values
		if len(system.Alliances) > 0 {
			break // Alliance formed, stop trying
		}
	}

	// Check if alliance was created
	if len(system.Alliances) == 0 {
		t.Error("Automatic alliance formation should have created alliances")
	}
}

// TestSharedDefenseMechanism tests alliance shared defense
func TestSharedDefenseMechanism(t *testing.T) {
	// Create warfare system
	system := NewColonyWarfareSystem()

	// Create three colonies: two allies and one aggressor
	queen1 := NewEntity(1, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 0, Y: 0})
	queen2 := NewEntity(2, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 20, Y: 20})
	queen3 := NewEntity(3, []string{"aggression", "strength"}, "predator", Position{X: 50, Y: 50})

	colony1 := NewCasteColony(1, queen1, Position{X: 0, Y: 0})
	colony2 := NewCasteColony(2, queen2, Position{X: 20, Y: 20})
	colony3 := NewCasteColony(3, queen3, Position{X: 50, Y: 50})

	// Set reasonable colony sizes and initial fitness
	colony1.ColonySize = 50
	colony2.ColonySize = 40
	colony3.ColonySize = 60
	colony1.ColonyFitness = 1.0 // Set non-zero fitness for military calculations
	colony2.ColonyFitness = 1.0
	colony3.ColonyFitness = 1.0

	// Add soldiers and workers to colonies for military strength
	colony1.CasteDistribution[Soldier] = 10
	colony1.CasteDistribution[Worker] = 20
	colony2.CasteDistribution[Soldier] = 8
	colony2.CasteDistribution[Worker] = 15
	colony3.CasteDistribution[Soldier] = 15
	colony3.CasteDistribution[Worker] = 25

	// Register colonies
	system.RegisterColony(colony1)
	system.RegisterColony(colony2)
	system.RegisterColony(colony3)

	// Create alliance between colony1 and colony2
	members := []int{colony1.ID, colony2.ID}
	alliance := system.CreateAlliance(members, "defensive", 0.1, 0)

	if alliance == nil {
		t.Error("Alliance should be created")
	}

	// Create conflict where colony3 attacks colony1
	conflict := system.StartConflict(colony3, colony1, TotalWar, 0)

	if conflict == nil {
		t.Error("Conflict should be created")
	}

	// Record initial colony fitness and size
	initialColony1Fitness := colony1.ColonyFitness
	initialColony2Size := colony2.ColonySize

	colonies := []*CasteColony{colony1, colony2, colony3}

	// Process alliance shared defense
	system.ProcessAlliances(colonies, 1)

	// Colony2 should have helped defend colony1 (taking some losses)
	if colony2.ColonySize >= initialColony2Size {
		t.Error("Ally should take losses when helping in defense")
	}

	// Colony1 should have received defensive bonus (fitness increase)
	if colony1.ColonyFitness <= initialColony1Fitness {
		t.Errorf("Defended colony should have received support bonus: initial=%.3f, final=%.3f",
			initialColony1Fitness, colony1.ColonyFitness)
	}
}
