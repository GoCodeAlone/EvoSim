package main

import (
	"testing"
)

func TestBioRhythmBasics(t *testing.T) {
	// Create a test entity
	entity := NewEntity(1, []string{"circadian_preference", "sleep_need", "hunger_need", "thirst_need"}, "test", Position{X: 50, Y: 50})

	// Set some biorhythm traits
	entity.SetTrait("circadian_preference", -0.6) // Nocturnal
	entity.SetTrait("sleep_need", 0.4)
	entity.SetTrait("hunger_need", 0.6)
	entity.SetTrait("thirst_need", 0.3)

	// Verify biorhythm was initialized
	if entity.BioRhythm == nil {
		t.Fatal("BioRhythm should be initialized")
	}

	// Check that activities are initialized
	if len(entity.BioRhythm.Activities) == 0 {
		t.Fatal("Activities should be initialized")
	}

	// Check specific activities exist
	activities := []ActivityType{ActivitySleep, ActivityEat, ActivityDrink, ActivityPlay, ActivityExplore, ActivityScavenge}
	for _, activity := range activities {
		if _, exists := entity.BioRhythm.Activities[activity]; !exists {
			t.Errorf("Activity %d should exist", activity)
		}
	}
}

func TestBioRhythmTimeEffects(t *testing.T) {
	// Create entities with different circadian preferences
	nocturnalEntity := NewEntity(1, []string{"circadian_preference", "sleep_need", "hunger_need", "thirst_need"}, "predator", Position{X: 50, Y: 50})
	nocturnalEntity.SetTrait("circadian_preference", -0.8) // Strongly nocturnal
	// Reinitialize biorhythm with correct traits
	nocturnalEntity.BioRhythm = NewBioRhythm(nocturnalEntity.ID, nocturnalEntity)

	diurnalEntity := NewEntity(2, []string{"circadian_preference", "sleep_need", "hunger_need", "thirst_need"}, "herbivore", Position{X: 50, Y: 50})
	diurnalEntity.SetTrait("circadian_preference", 0.7) // Strongly diurnal
	// Reinitialize biorhythm with correct traits
	diurnalEntity.BioRhythm = NewBioRhythm(diurnalEntity.ID, diurnalEntity)

	// Test night time effects
	nightTimeState := TimeState{
		TimeOfDay:    Night,
		Season:       Summer,
		Temperature:  0.5,
		Illumination: 0.1,
		SeasonalMod:  1.0,
	}

	// Force entities to choose activities appropriate for night
	// For nocturnal: should be active (exploring, scavenging, eating)
	nocturnalEntity.BioRhythm.Activities[ActivityExplore].NeedLevel = 0.8
	nocturnalEntity.BioRhythm.Activities[ActivitySleep].NeedLevel = 0.1

	// For diurnal: should be sleeping/resting
	diurnalEntity.BioRhythm.Activities[ActivitySleep].NeedLevel = 0.9
	diurnalEntity.BioRhythm.Activities[ActivityExplore].NeedLevel = 0.1

	// Update biorhythms
	nocturnalEntity.BioRhythm.Update(100, nocturnalEntity, nightTimeState)
	diurnalEntity.BioRhythm.Update(100, diurnalEntity, nightTimeState)

	// Check activity modifiers
	nocturnalModifier := nocturnalEntity.BioRhythm.GetActivityModifier(nocturnalEntity, nightTimeState)
	diurnalModifier := diurnalEntity.BioRhythm.GetActivityModifier(diurnalEntity, nightTimeState)

	// Debug output
	nocturnalActivity := nocturnalEntity.BioRhythm.GetCurrentActivity()
	diurnalActivity := diurnalEntity.BioRhythm.GetCurrentActivity()

	// Nocturnal entities should be more active at night
	if nocturnalModifier <= diurnalModifier {
		t.Errorf("Nocturnal entity should be more active at night: nocturnal=%.2f (activity=%d), diurnal=%.2f (activity=%d)",
			nocturnalModifier, nocturnalActivity, diurnalModifier, diurnalActivity)
	}

	// Test day time effects
	dayTimeState := TimeState{
		TimeOfDay:    Midday,
		Season:       Summer,
		Temperature:  0.7,
		Illumination: 1.0,
		SeasonalMod:  1.0,
	}

	// Force entities to choose activities appropriate for day
	// For diurnal: should be active (eating, exploring, playing)
	diurnalEntity.BioRhythm.Activities[ActivityEat].NeedLevel = 0.8
	diurnalEntity.BioRhythm.Activities[ActivitySleep].NeedLevel = 0.1

	// For nocturnal: should be sleeping/resting
	nocturnalEntity.BioRhythm.Activities[ActivitySleep].NeedLevel = 0.9
	nocturnalEntity.BioRhythm.Activities[ActivityEat].NeedLevel = 0.1

	// Update biorhythms so entities choose appropriate activities for midday
	nocturnalEntity.BioRhythm.Update(200, nocturnalEntity, dayTimeState)
	diurnalEntity.BioRhythm.Update(200, diurnalEntity, dayTimeState)

	// Check activity modifiers
	nocturnalModifierDay := nocturnalEntity.BioRhythm.GetActivityModifier(nocturnalEntity, dayTimeState)
	diurnalModifierDay := diurnalEntity.BioRhythm.GetActivityModifier(diurnalEntity, dayTimeState)

	// Diurnal entities should be more active during the day
	if diurnalModifierDay <= nocturnalModifierDay {
		nocturnalActivityDay := nocturnalEntity.BioRhythm.GetCurrentActivity()
		diurnalActivityDay := diurnalEntity.BioRhythm.GetCurrentActivity()
		t.Errorf("Diurnal entity should be more active during day: diurnal=%.2f (activity=%d), nocturnal=%.2f (activity=%d)",
			diurnalModifierDay, diurnalActivityDay, nocturnalModifierDay, nocturnalActivityDay)
	}
}

func TestBioRhythmNeeds(t *testing.T) {
	entity := NewEntity(1, []string{"circadian_preference", "sleep_need", "hunger_need", "thirst_need"}, "test", Position{X: 50, Y: 50})
	entity.SetTrait("sleep_need", 0.5)
	entity.SetTrait("hunger_need", 0.7)
	entity.SetTrait("thirst_need", 0.4)

	timeState := TimeState{
		TimeOfDay:    Morning,
		Season:       Spring,
		Temperature:  0.6,
		Illumination: 0.8,
		SeasonalMod:  1.2,
	}

	// Set initial need levels to be low so we can see them increase
	entity.BioRhythm.Activities[ActivitySleep].NeedLevel = 0.1
	entity.BioRhythm.Activities[ActivityEat].NeedLevel = 0.1
	entity.BioRhythm.Activities[ActivityDrink].NeedLevel = 0.1

	// Record initial need levels
	initialSleep := entity.BioRhythm.GetActivityNeed(ActivitySleep)
	initialHunger := entity.BioRhythm.GetActivityNeed(ActivityEat)
	initialThirst := entity.BioRhythm.GetActivityNeed(ActivityDrink)

	// Update biorhythm many times to simulate time passing
	for i := 0; i < 200; i++ {
		entity.BioRhythm.Update(i, entity, timeState)
	}

	// Need levels should have increased over time
	finalSleep := entity.BioRhythm.GetActivityNeed(ActivitySleep)
	finalHunger := entity.BioRhythm.GetActivityNeed(ActivityEat)
	finalThirst := entity.BioRhythm.GetActivityNeed(ActivityDrink)

	if finalSleep <= initialSleep {
		t.Errorf("Sleep need should increase over time: initial=%.3f, final=%.3f", initialSleep, finalSleep)
	}
	if finalHunger <= initialHunger {
		t.Errorf("Hunger need should increase over time: initial=%.3f, final=%.3f", initialHunger, finalHunger)
	}
	if finalThirst <= initialThirst {
		t.Errorf("Thirst need should increase over time: initial=%.3f, final=%.3f", initialThirst, finalThirst)
	}

	// Needs should not exceed maximum
	if finalSleep > 1.0 {
		t.Error("Sleep need should not exceed 1.0")
	}
	if finalHunger > 1.0 {
		t.Error("Hunger need should not exceed 1.0")
	}
	if finalThirst > 1.0 {
		t.Error("Thirst need should not exceed 1.0")
	}
}

func TestBioRhythmActivitySchedule(t *testing.T) {
	// Test nocturnal entity schedule
	nocturnalEntity := NewEntity(1, []string{"circadian_preference"}, "predator", Position{X: 50, Y: 50})
	nocturnalEntity.SetTrait("circadian_preference", -0.7) // Nocturnal
	// Reinitialize biorhythm after setting traits
	nocturnalEntity.BioRhythm = NewBioRhythm(nocturnalEntity.ID, nocturnalEntity)

	// Check that night activities include active behaviors
	nightActivities := nocturnalEntity.BioRhythm.ActivitySchedule[Night]
	hasActiveActivity := false
	for _, activity := range nightActivities {
		if activity == ActivityExplore || activity == ActivityScavenge || activity == ActivityEat {
			hasActiveActivity = true
			break
		}
	}
	if !hasActiveActivity {
		t.Error("Nocturnal entity should have active behaviors scheduled for night")
	}

	// Check that day activities include sleep
	dayActivities := nocturnalEntity.BioRhythm.ActivitySchedule[Midday]
	hasSleep := false
	for _, activity := range dayActivities {
		if activity == ActivitySleep {
			hasSleep = true
			break
		}
	}
	if !hasSleep {
		t.Error("Nocturnal entity should have sleep scheduled for day")
	}

	// Test diurnal entity schedule
	diurnalEntity := NewEntity(2, []string{"circadian_preference", "cooperation", "intelligence"}, "herbivore", Position{X: 50, Y: 50})
	diurnalEntity.SetTrait("circadian_preference", 0.8) // Diurnal
	diurnalEntity.SetTrait("cooperation", 0.7)
	diurnalEntity.SetTrait("intelligence", 0.6)
	// Reinitialize biorhythm after setting traits
	diurnalEntity.BioRhythm = NewBioRhythm(diurnalEntity.ID, diurnalEntity)

	// Check that day activities include active behaviors
	dayActivitiesDiurnal := diurnalEntity.BioRhythm.ActivitySchedule[Morning]
	hasActiveActivityDiurnal := false
	for _, activity := range dayActivitiesDiurnal {
		if activity == ActivityExplore || activity == ActivityEat || activity == ActivityScavenge {
			hasActiveActivityDiurnal = true
			break
		}
	}
	if !hasActiveActivityDiurnal {
		t.Error("Diurnal entity should have active behaviors scheduled for day")
	}
}

func TestBioRhythmEatingBehavior(t *testing.T) {
	entity := NewEntity(1, []string{"circadian_preference", "hunger_need", "species"}, "herbivore", Position{X: 50, Y: 50})
	entity.SetTrait("hunger_need", 0.8) // High hunger need

	// Create a plant to eat
	plant := NewPlant(0, PlantGrass, Position{X: 50, Y: 50})
	plant.Energy = 50.0

	// Set high hunger need
	entity.BioRhythm.Activities[ActivityEat].NeedLevel = 0.9

	// Entity should be able to eat when very hungry
	if !entity.CanEatPlant(plant) {
		t.Error("Herbivore entity should be able to eat plant")
	}

	if !entity.EatPlant(plant, 100) {
		t.Error("Entity should eat plant when hungry")
	}

	// Hunger need should be reduced after eating
	hungerAfter := entity.BioRhythm.GetActivityNeed(ActivityEat)
	if hungerAfter >= 0.9 {
		t.Error("Hunger need should be reduced after eating")
	}

	// Set low hunger need
	entity.BioRhythm.Activities[ActivityEat].NeedLevel = 0.1
	entity.Energy = 80 // High energy

	// Entity should not eat when not hungry and energy is high
	plant2 := NewPlant(1, PlantGrass, Position{X: 50, Y: 50})
	plant2.Energy = 50.0

	// With low hunger and high energy, entity should not eat
	if entity.EatPlant(plant2, 200) {
		t.Error("Entity should not eat when not hungry and energy is high")
	}
}

func TestBioRhythmDrinkingBehavior(t *testing.T) {
	// Create world and entity
	config := WorldConfig{
		Width:          100,
		Height:         100,
		GridWidth:      10,
		GridHeight:     10,
		PopulationSize: 1,
	}
	world := NewWorld(config)

	// Set water biome at entity location
	gridX := 5
	gridY := 5
	world.Grid[gridY][gridX].Biome = BiomeWater

	entity := NewEntity(1, []string{"thirst_need"}, "test", Position{X: 50, Y: 50})
	entity.SetTrait("thirst_need", 0.6)
	entity.Energy = 80.0 // Start with less than full energy to see the effect

	// Set high thirst need
	entity.BioRhythm.Activities[ActivityDrink].NeedLevel = 0.8
	initialEnergy := entity.Energy

	// Entity should be able to drink when thirsty
	drinkResult := entity.DrinkWater(world, 100)
	if !drinkResult {
		t.Error("Entity should be able to drink when thirsty and in water biome")
	}

	// Energy should increase after drinking (only check if drink was successful)
	if drinkResult && entity.Energy <= initialEnergy {
		t.Errorf("Energy should increase after drinking: before=%.1f, after=%.1f", initialEnergy, entity.Energy)
	}

	// Thirst need should be reduced
	thirstAfter := entity.BioRhythm.GetActivityNeed(ActivityDrink)
	if thirstAfter >= 0.8 {
		t.Error("Thirst need should be reduced after drinking")
	}

	// Set low thirst need
	entity.BioRhythm.Activities[ActivityDrink].NeedLevel = 0.3

	// Entity should not drink when not thirsty
	if entity.DrinkWater(world, 200) {
		t.Error("Entity should not drink when not thirsty")
	}
}
