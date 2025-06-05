package main

import (
	"testing"
)

func TestReproductiveMaturityWithClassification(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	reproductionSystem := NewReproductionSystem(NewCentralEventBus(1000))
	
	// Create entities with different classifications
	young := NewEntity(1, []string{"intelligence", "endurance"}, "test", Position{})
	young.SetTrait("intelligence", -0.8) // Prokaryotic-like
	young.Classification = ClassificationProkaryotic
	young.MaxLifespan = classifier.CalculateLifespan(young, young.Classification)
	young.Age = 1 // Young age (maturation age is ~7 for prokaryotic)
	
	mature := NewEntity(2, []string{"intelligence", "endurance"}, "test", Position{})
	mature.SetTrait("intelligence", -0.3) // Eukaryotic (more reasonable for testing)
	mature.Classification = ClassificationEukaryotic
	mature.MaxLifespan = classifier.CalculateLifespan(mature, mature.Classification)
	
	// Get actual maturation age for this classification
	maturationAge := classifier.LifespanData[ClassificationEukaryotic].MaturationAge
	mature.Age = maturationAge + 5 // Mature age (well beyond maturation)
	
	// Test maturity checks
	if classifier.IsReproductivelyMature(young, young.Classification) {
		t.Error("Young prokaryotic entity should not be mature yet")
	}
	
	if !classifier.IsReproductivelyMature(mature, mature.Classification) {
		t.Errorf("Mature eukaryotic entity should be mature (age: %d, maturation: %d)", 
			mature.Age, maturationAge)
	}
	
	// Test mating compatibility
	young.Energy = 100.0
	mature.Energy = 100.0
	
	// Young entity should not be able to mate due to immaturity
	canMate := reproductionSystem.CanMateWithClassification(young, mature, classifier, 100)
	if canMate {
		t.Error("Young immature entity should not be able to mate")
	}
	
	// Make young entity mature
	young.Age = classifier.LifespanData[young.Classification].MaturationAge + 1
	
	canMate = reproductionSystem.CanMateWithClassification(young, mature, classifier, 100)
	if !canMate {
		t.Error("Both mature entities should be able to mate")
	}
}

func TestReproductiveVigor(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	entity := NewEntity(1, []string{"endurance"}, "test", Position{})
	entity.Classification = ClassificationEukaryotic
	entity.MaxLifespan = classifier.CalculateLifespan(entity, entity.Classification)
	
	data := classifier.LifespanData[entity.Classification]
	
	// Test young entity (below maturation)
	entity.Age = data.MaturationAge - 10
	vigor := classifier.CalculateReproductiveVigor(entity, entity.Classification)
	if vigor != 0.0 {
		t.Errorf("Young entity should have 0 reproductive vigor, got %f", vigor)
	}
	
	// Test mature entity at peak age
	entity.Age = data.PeakAge
	vigor = classifier.CalculateReproductiveVigor(entity, entity.Classification)
	if vigor != 1.0 {
		t.Errorf("Entity at peak age should have 1.0 reproductive vigor, got %f", vigor)
	}
	
	// Test entity in senescence (but before max lifespan)
	entity.Age = data.SenescenceAge + 100
	vigor = classifier.CalculateReproductiveVigor(entity, entity.Classification)
	if vigor >= 1.0 || vigor < 0.0 {
		t.Errorf("Senescent entity should have reduced reproductive vigor between 0 and 1, got %f", vigor)
	}
	
	// Test entity very old but not at max lifespan yet
	entity.Age = entity.MaxLifespan - 100
	vigor = classifier.CalculateReproductiveVigor(entity, entity.Classification)
	if vigor >= 1.0 || vigor < 0.0 {
		t.Errorf("Very old entity should have very reduced reproductive vigor between 0 and 1, got %f", vigor)
	}
}

func TestReproductionWithClassificationIntegration(t *testing.T) {
	// Create a minimal world to test reproduction integration
	config := WorldConfig{
		Width:          100,
		Height:         100,
		GridWidth:      10,
		GridHeight:     10,
		NumPopulations: 1,
	}
	
	world := NewWorld(config)
	
	// Create two entities
	entity1 := NewEntity(1, []string{"intelligence", "endurance"}, "test", Position{X: 50, Y: 50})
	entity1.SetTrait("intelligence", 0.5)
	entity1.SetTrait("endurance", 0.5)
	entity1.Energy = 100.0
	
	entity2 := NewEntity(2, []string{"intelligence", "endurance"}, "test", Position{X: 52, Y: 52})
	entity2.SetTrait("intelligence", 0.5)
	entity2.SetTrait("endurance", 0.5)
	entity2.Energy = 100.0
	
	// Classify entities
	entity1.Classification = world.OrganismClassifier.ClassifyEntity(entity1, world.CellularSystem)
	entity1.MaxLifespan = world.OrganismClassifier.CalculateLifespan(entity1, entity1.Classification)
	
	entity2.Classification = world.OrganismClassifier.ClassifyEntity(entity2, world.CellularSystem)
	entity2.MaxLifespan = world.OrganismClassifier.CalculateLifespan(entity2, entity2.Classification)
	
	// Make entities mature
	maturationAge := world.OrganismClassifier.LifespanData[entity1.Classification].MaturationAge
	entity1.Age = maturationAge + 10
	entity2.Age = maturationAge + 10
	
	// Test reproduction readiness
	energyThreshold := 30.0
	maintenanceCost1 := world.OrganismClassifier.CalculateEnergyMaintenance(entity1, entity1.Classification)
	
	if entity1.Energy < energyThreshold+maintenanceCost1*5 {
		t.Error("Entity1 should have sufficient energy for reproduction")
	}
	
	// Test mating compatibility
	canMate := world.ReproductionSystem.CanMateWithClassification(entity1, entity2, world.OrganismClassifier, 100)
	if !canMate {
		t.Error("Both mature entities with sufficient energy should be able to mate")
	}
}