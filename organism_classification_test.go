package main

import (
	"testing"
)

func TestOrganismClassifierInitialization(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	if classifier == nil {
		t.Fatal("OrganismClassifier should not be nil")
	}
	
	if len(classifier.LifespanData) != 5 {
		t.Errorf("Expected 5 organism classifications, got %d", len(classifier.LifespanData))
	}
	
	// Check that all classification types have data
	expectedClassifications := []OrganismClassification{
		ClassificationProkaryotic,
		ClassificationEukaryotic,
		ClassificationSimpleMulticellular,
		ClassificationComplexMulticellular,
		ClassificationAdvancedMulticellular,
	}
	
	for _, classification := range expectedClassifications {
		if data, exists := classifier.LifespanData[classification]; !exists {
			t.Errorf("Missing lifespan data for classification %d", classification)
		} else {
			if data.BaseLifespanTicks <= 0 {
				t.Errorf("Invalid base lifespan for classification %d: %d", classification, data.BaseLifespanTicks)
			}
			if data.AgingRate <= 0 {
				t.Errorf("Invalid aging rate for classification %d: %f", classification, data.AgingRate)
			}
		}
	}
}

func TestEntityClassificationByTraits(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	// Test prokaryotic classification (low intelligence, size, cooperation)
	prokaryoticEntity := NewEntity(1, []string{"intelligence", "size", "cooperation"}, "test", Position{})
	prokaryoticEntity.SetTrait("intelligence", -0.8)
	prokaryoticEntity.SetTrait("size", -0.6)
	prokaryoticEntity.SetTrait("cooperation", -0.7)
	
	classification := classifier.ClassifyEntity(prokaryoticEntity, nil)
	if classification != ClassificationProkaryotic {
		t.Errorf("Expected ClassificationProkaryotic, got %d", classification)
	}
	
	// Test advanced multicellular classification (high intelligence, size, cooperation)
	advancedEntity := NewEntity(2, []string{"intelligence", "size", "cooperation"}, "test", Position{})
	advancedEntity.SetTrait("intelligence", 0.8)
	advancedEntity.SetTrait("size", 0.7)
	advancedEntity.SetTrait("cooperation", 0.9)
	
	classification = classifier.ClassifyEntity(advancedEntity, nil)
	if classification != ClassificationAdvancedMulticellular {
		t.Errorf("Expected ClassificationAdvancedMulticellular, got %d", classification)
	}
}

func TestLifespanCalculation(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	entity := NewEntity(1, []string{"endurance", "size"}, "test", Position{})
	entity.SetTrait("endurance", 0.5)
	entity.SetTrait("size", 0.3)
	
	// Test different classifications
	classifications := []OrganismClassification{
		ClassificationProkaryotic,
		ClassificationEukaryotic,
		ClassificationSimpleMulticellular,
		ClassificationComplexMulticellular,
		ClassificationAdvancedMulticellular,
	}
	
	for _, classification := range classifications {
		lifespan := classifier.CalculateLifespan(entity, classification)
		baseLifespan := classifier.LifespanData[classification].BaseLifespanTicks
		
		if lifespan <= 0 {
			t.Errorf("Invalid lifespan for classification %d: %d", classification, lifespan)
		}
		
		// Check that lifespan is within reasonable bounds (30% to 200% of base)
		minLifespan := int(float64(baseLifespan) * 0.3)
		maxLifespan := int(float64(baseLifespan) * 2.0)
		
		if lifespan < minLifespan || lifespan > maxLifespan {
			t.Errorf("Lifespan %d for classification %d is outside reasonable bounds [%d, %d]", 
				lifespan, classification, minLifespan, maxLifespan)
		}
	}
}

func TestAgingRateCalculation(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	entity := NewEntity(1, []string{"metabolism", "size"}, "test", Position{})
	entity.SetTrait("metabolism", 0.5)
	entity.SetTrait("size", 0.2)
	
	// Test different classifications
	for classification := ClassificationProkaryotic; classification <= ClassificationAdvancedMulticellular; classification++ {
		agingRate := classifier.CalculateAgingRate(entity, classification)
		
		if agingRate <= 0 {
			t.Errorf("Invalid aging rate for classification %d: %f", classification, agingRate)
		}
		
		// Check that aging rate is reasonable (0.1 to 5.0)
		if agingRate < 0.1 || agingRate > 5.0 {
			t.Errorf("Aging rate %f for classification %d is outside reasonable bounds [0.1, 5.0]", 
				agingRate, classification)
		}
	}
}

func TestLifespanProgression(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	// Test that more complex organisms have longer lifespans
	entity := NewEntity(1, []string{"endurance", "size"}, "test", Position{})
	entity.SetTrait("endurance", 0.0)
	entity.SetTrait("size", 0.0)
	
	var prevLifespan int
	for classification := ClassificationProkaryotic; classification <= ClassificationAdvancedMulticellular; classification++ {
		lifespan := classifier.CalculateLifespan(entity, classification)
		
		if classification > ClassificationProkaryotic && lifespan <= prevLifespan {
			t.Errorf("Expected longer lifespan for more complex organisms: classification %d has lifespan %d, previous was %d", 
				classification, lifespan, prevLifespan)
		}
		
		prevLifespan = lifespan
	}
}

func TestDeathByOldAge(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	entity := NewEntity(1, []string{"endurance"}, "test", Position{})
	entity.SetTrait("endurance", 0.0)
	
	classification := ClassificationEukaryotic
	maxLifespan := classifier.CalculateLifespan(entity, classification)
	
	// Young entity should not die of old age
	entity.Age = maxLifespan / 4
	if classifier.IsDeathByOldAge(entity, classification, maxLifespan) {
		t.Error("Young entity should not die of old age")
	}
	
	// Entity at max lifespan should die
	entity.Age = maxLifespan
	if !classifier.IsDeathByOldAge(entity, classification, maxLifespan) {
		t.Error("Entity at max lifespan should die of old age")
	}
	
	// Entity beyond max lifespan should definitely die
	entity.Age = maxLifespan + 100
	if !classifier.IsDeathByOldAge(entity, classification, maxLifespan) {
		t.Error("Entity beyond max lifespan should die of old age")
	}
}

func TestEntityUpdateWithClassification(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	dnaSystem := NewDNASystem(NewCentralEventBus(1000))
	cellularSystem := NewCellularSystem(dnaSystem, NewCentralEventBus(1000))
	
	entity := NewEntity(1, []string{"intelligence", "size", "cooperation", "endurance"}, "test", Position{})
	entity.SetTrait("intelligence", 0.0)
	entity.SetTrait("size", 0.0)
	entity.SetTrait("cooperation", 0.0)
	entity.SetTrait("endurance", 0.5)
	
	initialAge := entity.Age
	
	// Update entity multiple times
	for i := 0; i < 10; i++ {
		entity.UpdateWithClassification(classifier, cellularSystem)
	}
	
	// Entity should still be alive after a few updates
	if !entity.IsAlive {
		t.Error("Entity should still be alive after a few updates")
	}
	
	// Age should have been processed (might not increase every tick due to fractional aging)
	if entity.Age < initialAge {
		t.Error("Entity age should not decrease")
	}
	
	// Entity should have been classified
	if entity.Classification == ClassificationEukaryotic && entity.MaxLifespan == 3360 {
		t.Error("Entity should have been classified and assigned a proper lifespan")
	}
}

func TestRealisticLifespanRanges(t *testing.T) {
	timeSystem := NewAdvancedTimeSystemLegacy(480, 120)
	classifier := NewOrganismClassifier(timeSystem)
	
	ticksPerDay := float64(timeSystem.DayLength)
	
	// Test that lifespans are biologically realistic
	// Updated for new time scale with variance and trait modifiers
	testCases := []struct {
		classification OrganismClassification
		minDays       float64
		maxDays       float64
	}{
		{ClassificationProkaryotic, 2, 30},           // Hours to days (scaled with variance)
		{ClassificationEukaryotic, 15, 80},           // Days to weeks (scaled with variance)
		{ClassificationSimpleMulticellular, 40, 200}, // Weeks to months (scaled with variance)
		{ClassificationComplexMulticellular, 150, 800}, // Months to year (scaled with variance)
		{ClassificationAdvancedMulticellular, 500, 2500}, // Year+ (scaled with variance)
	}
	
	entity := NewEntity(1, []string{"endurance", "size"}, "test", Position{})
	entity.SetTrait("endurance", 0.0)
	entity.SetTrait("size", 0.0)
	
	for _, testCase := range testCases {
		lifespan := classifier.CalculateLifespan(entity, testCase.classification)
		lifespanDays := float64(lifespan) / ticksPerDay
		
		if lifespanDays < testCase.minDays || lifespanDays > testCase.maxDays {
			t.Errorf("Lifespan for %s is %.1f days, expected between %.1f and %.1f days", 
				classifier.GetClassificationName(testCase.classification),
				lifespanDays, testCase.minDays, testCase.maxDays)
		}
	}
}