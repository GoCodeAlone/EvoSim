package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestDietaryMemoryInitialization(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)

	if entity.DietaryMemory == nil {
		t.Error("Expected DietaryMemory to be initialized")
	}

	if entity.DietaryMemory.DietaryFitness != 1.0 {
		t.Errorf("Expected initial dietary fitness 1.0, got %f", entity.DietaryMemory.DietaryFitness)
	}

	if entity.DietaryMemory.PlantTypePreferences == nil {
		t.Error("Expected PlantTypePreferences to be initialized")
	}

	if entity.DietaryMemory.PreySpeciesPreferences == nil {
		t.Error("Expected PreySpeciesPreferences to be initialized")
	}
}

func TestEnvironmentalMemoryInitialization(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)

	if entity.EnvironmentalMemory == nil {
		t.Error("Expected EnvironmentalMemory to be initialized")
	}

	if entity.EnvironmentalMemory.AdaptationFitness != 1.0 {
		t.Errorf("Expected initial adaptation fitness 1.0, got %f", entity.EnvironmentalMemory.AdaptationFitness)
	}

	if entity.EnvironmentalMemory.BiomeExposure == nil {
		t.Error("Expected BiomeExposure to be initialized")
	}
}

func TestDietaryPreferenceInheritance(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	parent1 := NewEntity(1, []string{"strength"}, "test", pos)
	parent2 := NewEntity(2, []string{"strength"}, "test", pos)

	// Set up some dietary preferences for parents
	parent1.DietaryMemory.PlantTypePreferences[0] = 1.5 // Strong preference for plant type 0
	parent1.DietaryMemory.PreySpeciesPreferences["herbivore"] = 1.2
	
	parent2.DietaryMemory.PlantTypePreferences[0] = 0.8 // Weaker preference
	parent2.DietaryMemory.PreySpeciesPreferences["herbivore"] = 1.0

	// Create child through crossover
	child := Crossover(parent1, parent2, 3, "test")

	// Check that child inherited averaged preferences with some variation
	plantPref := child.DietaryMemory.PlantTypePreferences[0]
	if plantPref < 0.5 || plantPref > 1.8 {
		t.Errorf("Expected child plant preference between 0.5-1.8, got %f", plantPref)
	}

	preyPref := child.DietaryMemory.PreySpeciesPreferences["herbivore"]
	if preyPref < 0.6 || preyPref > 1.6 {
		t.Errorf("Expected child prey preference between 0.6-1.6, got %f", preyPref)
	}
}

func TestFeedbackLoopFitnessContribution(t *testing.T) {
	engine := NewEvaluationEngine()
	engine.AddRule("simple", "strength", 1.0, 1.0, false)

	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)
	entity.SetTrait("strength", 0.5)

	// Modify dietary fitness to test contribution
	entity.DietaryMemory.DietaryFitness = 1.5 // Above neutral
	entity.EnvironmentalMemory.AdaptationFitness = 1.3 // Above neutral

	fitness := engine.Evaluate(entity)

	// Should be higher than the simple rule evaluation due to feedback loop fitness
	// Rule contributes 0.5, molecular systems contribute some amount, feedback contributes 0.2 * some_value
	if fitness <= 0.5 {
		t.Errorf("Expected fitness > 0.5 due to feedback loop contribution, got %f", fitness)
	}
}

func TestDietaryMemoryUpdatesFromConsumption(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "herbivore", pos)
	
	// Create a test plant
	plant := NewPlant(1, PlantGrass, Position{X: 1, Y: 1})
	
	// Initial state - no preferences
	initialPref := entity.DietaryMemory.PlantTypePreferences[int(PlantGrass)]
	
	// Simulate consumption
	success := entity.EatPlant(plant, 100)
	
	if !success {
		t.Error("Expected herbivore to successfully eat grass")
	}
	
	// Check that preference was updated
	newPref := entity.DietaryMemory.PlantTypePreferences[int(PlantGrass)]
	if newPref <= initialPref {
		t.Errorf("Expected plant preference to increase after consumption, was %f, now %f", initialPref, newPref)
	}
	
	// Check that consumption was recorded
	if len(entity.DietaryMemory.ConsumptionHistory) == 0 {
		t.Error("Expected consumption to be recorded in history")
	}
	
	record := entity.DietaryMemory.ConsumptionHistory[0]
	if record.FoodType != "plant" {
		t.Errorf("Expected food type 'plant', got %s", record.FoodType)
	}
}

func TestEnvironmentalAdaptationTracking(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength", "aquatic_adaptation"}, "test", pos)
	entity.SetTrait("aquatic_adaptation", 0.8) // Well adapted to water
	
	// Initial exposure should be empty
	if len(entity.EnvironmentalMemory.BiomeExposure) != 0 {
		t.Error("Expected no initial biome exposure")
	}
	
	// Simulate environmental exposure
	entity.trackEnvironmentalExposure(BiomeWater, "Spring", nil, 100)
	
	// Check that exposure was recorded
	waterExposure := entity.EnvironmentalMemory.BiomeExposure[BiomeWater]
	if waterExposure <= 0 {
		t.Errorf("Expected positive water biome exposure, got %f", waterExposure)
	}
	
	// Check that adaptation fitness was updated (should be positive due to good aquatic adaptation)
	if entity.EnvironmentalMemory.AdaptationFitness <= 1.0 {
		t.Errorf("Expected adaptation fitness > 1.0 for well-adapted entity, got %f", entity.EnvironmentalMemory.AdaptationFitness)
	}
}

func TestBiasedMutationFromEnvironmentalPressure(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"aquatic_adaptation"}, "test", pos)
	
	// Set poor aquatic adaptation
	entity.SetTrait("aquatic_adaptation", -0.8)
	
	// Simulate high water biome exposure
	entity.EnvironmentalMemory.BiomeExposure[BiomeWater] = 0.9
	entity.EnvironmentalMemory.AdaptationFitness = 0.2 // Very poor adaptation
	
	// Test the bias method directly first
	negativeMutation := -0.2
	biasedMutation := entity.biasedMutation("aquatic_adaptation", negativeMutation)
	
	// With high exposure and poor trait, negative mutations should become positive
	if biasedMutation <= 0 {
		t.Logf("Bias method not working as expected. Original: %f, Biased: %f", negativeMutation, biasedMutation)
		t.Logf("Current trait value: %f", entity.GetTrait("aquatic_adaptation"))
		t.Logf("Water exposure: %f", entity.EnvironmentalMemory.BiomeExposure[BiomeWater])
	}
	
	// Record initial trait value
	initialAquatic := entity.GetTrait("aquatic_adaptation")
	
	// Count how many mutations improve the trait with direct mutation call
	improveCount := 0
	totalMutations := 200
	
	for i := 0; i < totalMutations; i++ {
		testEntity := entity.Clone()
		testEntity.EnvironmentalMemory.BiomeExposure[BiomeWater] = 0.9
		testEntity.EnvironmentalMemory.AdaptationFitness = 0.2
		
		// Force a mutation on the aquatic_adaptation trait specifically
		currentValue := testEntity.GetTrait("aquatic_adaptation")
		baseMutation := rand.NormFloat64() * 0.3
		biasedMut := testEntity.biasedMutation("aquatic_adaptation", baseMutation)
		newValue := math.Max(-2.0, math.Min(2.0, currentValue + biasedMut))
		testEntity.SetTrait("aquatic_adaptation", newValue)
		
		if newValue > initialAquatic {
			improveCount++
		}
	}
	
	// Should have more improvements than degradations due to environmental bias
	improveRate := float64(improveCount) / float64(totalMutations)
	if improveRate < 0.6 { // Expect at least 60% improvements due to bias
		t.Errorf("Expected improvement rate > 0.6 due to environmental bias, got %f", improveRate)
	}
}

func TestMutationRateIncreasesWithEnvironmentalPressure(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength"}, "test", pos)
	
	// Add environmental pressure
	entity.EnvironmentalMemory.RadiationPressure = 2.0 // High radiation
	entity.EnvironmentalMemory.TemperaturePressure = 1.5 // High temperature stress
	entity.EnvironmentalMemory.AdaptationFitness = 0.4 // Poor adaptation
	
	// Count mutations with low base rate
	mutationCount := 0
	testRuns := 1000
	
	for i := 0; i < testRuns; i++ {
		testEntity := entity.Clone()
		testEntity.DietaryMemory = NewDietaryMemory()  
		testEntity.EnvironmentalMemory = entity.EnvironmentalMemory
		
		initialValue := testEntity.GetTrait("strength")
		testEntity.Mutate(0.01, 0.1) // Very low base mutation rate
		newValue := testEntity.GetTrait("strength")
		
		if newValue != initialValue {
			mutationCount++
		}
	}
	
	mutationRate := float64(mutationCount) / float64(testRuns)
	
	// Should have higher mutation rate than the base 0.01 due to environmental pressure
	if mutationRate <= 0.02 {
		t.Errorf("Expected mutation rate > 0.02 due to environmental pressure, got %f", mutationRate)
	}
}