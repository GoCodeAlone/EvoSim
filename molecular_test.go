package main

import (
	"testing"
)

func TestMolecularProfile(t *testing.T) {
	profile := NewMolecularProfile()

	// Test adding components
	profile.AddComponent(ProteinStructural, 0.5, 0.8, 1.0)
	profile.AddComponent(CarboSimple, 0.3, 0.7, 0.9)

	if len(profile.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(profile.Components))
	}

	// Test derived metrics calculation
	if profile.TotalBiomass <= 0 {
		t.Error("Expected positive total biomass")
	}

	if profile.Diversity <= 0 {
		t.Error("Expected positive diversity")
	}
}

func TestMolecularNeeds(t *testing.T) {
	// Create a test entity
	entity := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity.SetTrait("strength", 0.5)
	entity.SetTrait("intelligence", 0.7)
	entity.SetTrait("speed", 0.3)
	entity.SetTrait("size", 0.4)
	entity.SetTrait("aggression", 0.2)

	needs := NewMolecularNeeds(entity)

	// Test that requirements are set
	if len(needs.Requirements) == 0 {
		t.Error("Expected some molecular requirements")
	}

	// Test that high intelligence increases ATP needs
	if needs.Requirements[NucleicATP] <= 0.5 {
		t.Error("Expected high intelligence to increase ATP requirements")
	}

	// Test nutritional status calculation
	status := needs.GetOverallNutritionalStatus()
	if status < 0 || status > 1 {
		t.Errorf("Nutritional status should be between 0 and 1, got %f", status)
	}
}

func TestMolecularMetabolism(t *testing.T) {
	entity := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity.SetTrait("intelligence", 0.8)
	entity.SetTrait("size", 0.6)
	entity.SetTrait("cooperation", 0.4)

	metabolism := NewMolecularMetabolism(entity)

	// Test that efficiencies are set
	if len(metabolism.Efficiency) == 0 {
		t.Error("Expected molecular processing efficiencies")
	}

	// Test that intelligence affects processing rate
	expectedMinRate := 0.5 + 0.8*0.3 // Base + intelligence boost
	if metabolism.ProcessingRate < expectedMinRate {
		t.Errorf("Expected processing rate >= %f, got %f", expectedMinRate, metabolism.ProcessingRate)
	}
}

func TestPlantMolecularProfile(t *testing.T) {
	// Test different plant types
	plantTypes := []PlantType{PlantGrass, PlantBush, PlantTree, PlantMushroom, PlantAlgae, PlantCactus}

	for _, plantType := range plantTypes {
		plant := &Plant{
			ID:      1,
			Type:    plantType,
			Traits:  make(map[string]Trait),
			Age:     10,
			IsAlive: true,
		}
		plant.Traits["nutrition_value"] = Trait{Name: "nutrition_value", Value: 0.5}
		plant.Traits["toxicity"] = Trait{Name: "toxicity", Value: 0.3}

		profile := CreatePlantMolecularProfile(plant)

		if profile == nil {
			t.Errorf("Expected molecular profile for plant type %d", plantType)
			continue
		}

		if len(profile.Components) == 0 {
			t.Errorf("Expected molecular components for plant type %d", plantType)
		}

		// Test that different plant types have different profiles
		switch plantType {
		case PlantMushroom:
			// Mushrooms should have high protein
			if component, exists := profile.Components[ProteinEnzymatic]; exists {
				if component.Concentration < 0.5 {
					t.Error("Expected mushrooms to have high enzymatic protein")
				}
			}
		case PlantAlgae:
			// Algae should have high unsaturated lipids
			if component, exists := profile.Components[LipidUnsaturated]; exists {
				if component.Concentration < 0.5 {
					t.Error("Expected algae to have high unsaturated lipids")
				}
			}
		case PlantCactus:
			// Cactus should have toxins
			foundToxin := false
			for molType := range profile.Components {
				if molType >= ToxinAlkaloid {
					foundToxin = true
					break
				}
			}
			if !foundToxin {
				t.Error("Expected cactus to have toxins")
			}
		}
	}
}

func TestEntityMolecularProfile(t *testing.T) {
	entity := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		Species: "carnivore",
		Age:     20,
		IsAlive: true,
	}
	entity.SetTrait("size", 0.8)
	entity.SetTrait("strength", 0.9)
	entity.SetTrait("intelligence", 0.6)
	entity.SetTrait("aquatic_adaptation", 0.7)

	profile := CreateEntityMolecularProfile(entity)

	if profile == nil {
		t.Fatal("Expected entity molecular profile")
	}

	// Test that large, strong entities have more structural protein
	if component, exists := profile.Components[ProteinStructural]; exists {
		expectedMin := 0.4 + 0.8*0.2 + 0.9*0.2 // Base + size + strength
		if component.Concentration < expectedMin {
			t.Errorf("Expected structural protein >= %f, got %f", expectedMin, component.Concentration)
		}
	}

	// Test aquatic adaptation effect
	if component, exists := profile.Components[LipidUnsaturated]; exists {
		if component.Concentration <= 0.3 { // Should be boosted by aquatic adaptation
			t.Error("Expected aquatic adaptation to boost unsaturated lipids")
		}
	}
}

func TestMolecularDesirability(t *testing.T) {
	// Create a simple food profile
	foodProfile := NewMolecularProfile()
	foodProfile.AddComponent(ProteinStructural, 0.6, 0.8, 1.0)
	foodProfile.AddComponent(CarboSimple, 0.4, 0.7, 1.0)
	foodProfile.AddComponent(AminoEssential, 0.5, 0.9, 1.0)

	// Create entity needs
	entity := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity.SetTrait("strength", 0.8) // High protein needs
	needs := NewMolecularNeeds(entity)

	desirability := GetMolecularDesirability(foodProfile, needs)

	if desirability < 0 || desirability > 1 {
		t.Errorf("Desirability should be between 0 and 1, got %f", desirability)
	}

	// Test with toxic food
	toxicProfile := NewMolecularProfile()
	toxicProfile.AddComponent(ProteinStructural, 0.3, 0.8, 1.0)
	toxicProfile.AddComponent(ToxinAlkaloid, 0.8, 0.9, 1.0)

	toxicDesirability := GetMolecularDesirability(toxicProfile, needs)

	if toxicDesirability >= desirability {
		t.Error("Expected toxic food to be less desirable")
	}
}

func TestNutrientConsumption(t *testing.T) {
	// Create food profile
	foodProfile := NewMolecularProfile()
	foodProfile.AddComponent(CarboSimple, 0.8, 0.9, 1.0)
	foodProfile.AddComponent(ProteinStructural, 0.6, 0.8, 1.0)
	foodProfile.AddComponent(AminoEssential, 0.7, 0.9, 1.0)

	// Create entity
	entity := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity.SetTrait("intelligence", 0.7)
	entity.SetTrait("size", 0.5)

	needs := NewMolecularNeeds(entity)
	metabolism := NewMolecularMetabolism(entity)

	// Test consumption
	energyGained, toxinDamage := foodProfile.ConsumeNutrients(needs, metabolism, 0.5)

	if energyGained <= 0 {
		t.Error("Expected positive energy gain from nutrient consumption")
	}

	if toxinDamage < 0 {
		t.Error("Expected non-negative toxin damage")
	}

	// Test that deficiencies are reduced
	// Ensure there's an initial deficiency by setting one manually
	needs.Deficiencies[CarboSimple] = 0.5
	initialDeficiency := needs.Deficiencies[CarboSimple]
	t.Logf("Initial deficiency: %f", initialDeficiency)

	// Create fresh food profile for consumption test
	freshFoodProfile := NewMolecularProfile()
	freshFoodProfile.AddComponent(CarboSimple, 0.8, 0.9, 1.0)
	freshFoodProfile.AddComponent(ProteinStructural, 0.6, 0.8, 1.0)
	freshFoodProfile.AddComponent(AminoEssential, 0.7, 0.9, 1.0)

	freshFoodProfile.ConsumeNutrients(needs, metabolism, 0.3)
	finalDeficiency := needs.Deficiencies[CarboSimple]
	t.Logf("Final deficiency: %f", finalDeficiency)

	if finalDeficiency >= initialDeficiency {
		t.Errorf("Expected deficiency to decrease after consuming nutrients. Initial: %f, Final: %f", initialDeficiency, finalDeficiency)
	}
}

func TestMolecularEvolution(t *testing.T) {
	// Test that molecular needs evolve with traits
	entity1 := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity1.SetTrait("strength", 0.2)
	entity1.SetTrait("intelligence", 0.3)

	entity2 := &Entity{
		ID:      2,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity2.SetTrait("strength", 0.9)
	entity2.SetTrait("intelligence", 0.8)

	needs1 := NewMolecularNeeds(entity1)
	needs2 := NewMolecularNeeds(entity2)

	// High-strength entity should need more structural protein
	if needs2.Requirements[ProteinStructural] <= needs1.Requirements[ProteinStructural] {
		t.Error("Expected high-strength entity to need more structural protein")
	}

	// High-intelligence entity should need more ATP
	if needs2.Requirements[NucleicATP] <= needs1.Requirements[NucleicATP] {
		t.Error("Expected high-intelligence entity to need more ATP")
	}
}

func TestMolecularFitnessIntegration(t *testing.T) {
	// Test that molecular system affects fitness
	evaluationEngine := NewEvaluationEngine()
	evaluationEngine.AddRule("basic", "strength + intelligence", 1.0, 2.0, false)

	// Create two similar entities with different nutritional status
	entity1 := &Entity{
		ID:      1,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity1.SetTrait("strength", 0.5)
	entity1.SetTrait("intelligence", 0.5)
	entity1.MolecularNeeds = NewMolecularNeeds(entity1)
	entity1.MolecularMetabolism = NewMolecularMetabolism(entity1)

	entity2 := &Entity{
		ID:      2,
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	entity2.SetTrait("strength", 0.5)
	entity2.SetTrait("intelligence", 0.5)
	entity2.MolecularNeeds = NewMolecularNeeds(entity2)
	entity2.MolecularMetabolism = NewMolecularMetabolism(entity2)

	// Give entity2 better nutritional status
	for molType := range entity2.MolecularNeeds.Deficiencies {
		entity2.MolecularNeeds.Deficiencies[molType] *= 0.3 // Reduce deficiencies
	}

	fitness1 := evaluationEngine.Evaluate(entity1)
	fitness2 := evaluationEngine.Evaluate(entity2)

	if fitness2 <= fitness1 {
		t.Error("Expected entity with better nutrition to have higher fitness")
	}
}

func TestSpeciesSpecificMolecularProfiles(t *testing.T) {
	// Test that different species have different molecular profiles
	herbivore := &Entity{
		ID:      1,
		Species: "herbivore",
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	herbivore.SetTrait("size", 0.5)
	herbivore.SetTrait("strength", 0.4)

	carnivore := &Entity{
		ID:      2,
		Species: "carnivore",
		Traits:  make(map[string]Trait),
		IsAlive: true,
	}
	carnivore.SetTrait("size", 0.5)
	carnivore.SetTrait("strength", 0.4)

	herbProfile := CreateEntityMolecularProfile(herbivore)
	carnProfile := CreateEntityMolecularProfile(carnivore)

	// Carnivores should have more protein overall
	herbProtein := 0.0
	carnProtein := 0.0

	for molType, component := range herbProfile.Components {
		if molType >= ProteinStructural && molType <= ProteinDefensive {
			herbProtein += component.Concentration
		}
	}

	for molType, component := range carnProfile.Components {
		if molType >= ProteinStructural && molType <= ProteinDefensive {
			carnProtein += component.Concentration
		}
	}

	if carnProtein <= herbProtein {
		t.Error("Expected carnivores to have more protein than herbivores")
	}
}
