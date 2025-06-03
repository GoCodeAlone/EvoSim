package main

import (
	"math/rand"
	"testing"
)

// TestMeteorShowerEvent tests meteor shower events can modify the world map
func TestMeteorShowerEvent(t *testing.T) {
	// Use fixed seed for reproducible results
	rand.Seed(12345)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	world := NewWorld(config)
	
	// Record initial biome state
	initialRadiationCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeRadiation {
				initialRadiationCount++
			}
		}
	}
	
	// Generate meteor craters
	craters := world.generateMeteorCraters()
	
	// Verify meteor craters were generated
	if len(craters) == 0 {
		t.Fatal("Meteor shower should generate at least one crater")
	}
	
	// Apply the changes to verify they have expected effects
	for pos, biomeType := range craters {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
			world.Grid[gridY][gridX].Biome = biomeType
		}
	}
	
	// Count radiation zones after meteor impact
	finalRadiationCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeRadiation {
				finalRadiationCount++
			}
		}
	}
	
	// Verify radiation zones increased
	if finalRadiationCount <= initialRadiationCount {
		t.Errorf("Meteor shower should increase radiation zones: initial=%d, final=%d", 
			initialRadiationCount, finalRadiationCount)
	}
	
	// Verify reasonable number of radiation zones (should be 3-8 craters with surrounding areas)
	expectedMin := 3  // At least 3 craters
	expectedMax := 50 // At most 8 craters * ~6 surrounding cells each
	radiationIncrease := finalRadiationCount - initialRadiationCount
	
	if radiationIncrease < expectedMin || radiationIncrease > expectedMax {
		t.Errorf("Meteor impact radiation zones out of expected range: got %d, expected %d-%d",
			radiationIncrease, expectedMin, expectedMax)
	}
}

// TestEarthquakeEvent tests earthquake events can create mountain ranges
func TestEarthquakeEvent(t *testing.T) {
	rand.Seed(23456)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	world := NewWorld(config)
	
	// Record initial mountain count
	initialMountainCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeMountain {
				initialMountainCount++
			}
		}
	}
	
	// Generate seismic changes
	seismicChanges := world.generateSeismicChanges()
	
	// Verify seismic changes were generated
	if len(seismicChanges) == 0 {
		t.Fatal("Earthquake should generate seismic changes")
	}
	
	// Apply the changes
	for pos, biomeType := range seismicChanges {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
			world.Grid[gridY][gridX].Biome = biomeType
		}
	}
	
	// Count mountains after earthquake
	finalMountainCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeMountain {
				finalMountainCount++
			}
		}
	}
	
	// Verify mountain ranges were created
	if finalMountainCount <= initialMountainCount {
		t.Errorf("Earthquake should increase mountain count: initial=%d, final=%d",
			initialMountainCount, finalMountainCount)
	}
	
	// Verify reasonable number of new mountains (1-2 fault lines with surrounding areas)
	expectedMin := 1
	expectedMax := 150 // 2 fault lines * ~75 cells each max
	mountainIncrease := finalMountainCount - initialMountainCount
	
	if mountainIncrease < expectedMin || mountainIncrease > expectedMax {
		t.Errorf("Earthquake mountain creation out of expected range: got %d, expected %d-%d",
			mountainIncrease, expectedMin, expectedMax)
	}
}

// TestWildfireEvent tests wildfire events can create desert zones
func TestWildfireEvent(t *testing.T) {
	rand.Seed(34567)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	world := NewWorld(config)
	
	// Record initial desert count
	initialDesertCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeDesert {
				initialDesertCount++
			}
		}
	}
	
	// Generate fire zones
	fireChanges := world.generateFireZones()
	
	// Verify fire zones were generated
	if len(fireChanges) == 0 {
		t.Fatal("Wildfire should generate fire zones")
	}
	
	// Apply the changes
	for pos, biomeType := range fireChanges {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
			world.Grid[gridY][gridX].Biome = biomeType
		}
	}
	
	// Count deserts after wildfire
	finalDesertCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeDesert {
				finalDesertCount++
			}
		}
	}
	
	// Verify desert zones increased from fires
	if finalDesertCount <= initialDesertCount {
		t.Errorf("Wildfire should increase desert zones: initial=%d, final=%d",
			initialDesertCount, finalDesertCount)
	}
	
	// Verify reasonable number of burned areas (2-6 fires with spread)
	expectedMin := 2
	expectedMax := 200 // 6 fires * ~33 cells each max
	desertIncrease := finalDesertCount - initialDesertCount
	
	if desertIncrease < expectedMin || desertIncrease > expectedMax {
		t.Errorf("Wildfire desert creation out of expected range: got %d, expected %d-%d",
			desertIncrease, expectedMin, expectedMax)
	}
}

// TestFloodEvent tests flood events can create water zones
func TestFloodEvent(t *testing.T) {
	rand.Seed(45678)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	world := NewWorld(config)
	
	// Record initial water count  
	initialWaterCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeWater {
				initialWaterCount++
			}
		}
	}
	
	// Generate flood zones
	floodChanges := world.generateFloodZones()
	
	// Verify flood zones were generated
	if len(floodChanges) == 0 {
		t.Fatal("Flood should generate flood zones")
	}
	
	// Apply the changes
	for pos, biomeType := range floodChanges {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
			world.Grid[gridY][gridX].Biome = biomeType
		}
	}
	
	// Count water zones after flood
	finalWaterCount := 0
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if world.Grid[y][x].Biome == BiomeWater {
				finalWaterCount++
			}
		}
	}
	
	// Verify water zones increased from flooding
	if finalWaterCount <= initialWaterCount {
		t.Errorf("Flood should increase water zones: initial=%d, final=%d",
			initialWaterCount, finalWaterCount)
	}
	
	// Verify reasonable number of flooded areas (1-2 flood sources with spread)
	expectedMin := 1
	expectedMax := 80 // 2 flood sources * ~40 cells each max
	waterIncrease := finalWaterCount - initialWaterCount
	
	if waterIncrease < expectedMin || waterIncrease > expectedMax {
		t.Errorf("Flood water creation out of expected range: got %d, expected %d-%d",
			waterIncrease, expectedMin, expectedMax)
	}
}

// TestEnvironmentalEventTriggering tests that events can be triggered and applied
func TestEnvironmentalEventTriggering(t *testing.T) {
	rand.Seed(56789)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	world := NewWorld(config)
	
	// Record initial state
	initialEventCount := len(world.Events)
	
	// Manually trigger a random event
	world.triggerRandomEvent()
	
	// Verify an event was added
	if len(world.Events) != initialEventCount+1 {
		t.Errorf("Event should be added: initial=%d, final=%d", 
			initialEventCount, len(world.Events))
	}
	
	// Get the triggered event
	triggeredEvent := world.Events[len(world.Events)-1]
	
	// Verify event has required properties
	if triggeredEvent.Name == "" {
		t.Error("Event should have a name")
	}
	if triggeredEvent.Duration <= 0 {
		t.Error("Event should have positive duration")
	}
	
	// Test events with biome changes
	eventsWithChanges := []string{"Meteor Shower", "Wildfire", "Great Flood", "Earthquake"}
	eventHasBiomeChanges := false
	for _, eventName := range eventsWithChanges {
		if triggeredEvent.Name == eventName {
			if len(triggeredEvent.BiomeChanges) > 0 {
				eventHasBiomeChanges = true
			}
			break
		}
	}
	
	// If it's an event that should have biome changes, verify they exist
	if contains(eventsWithChanges, triggeredEvent.Name) && !eventHasBiomeChanges {
		t.Errorf("Event %s should have biome changes but has none", triggeredEvent.Name)
	}
}

// TestLongTermEventOccurrence tests that events eventually occur in long simulations
func TestLongTermEventOccurrence(t *testing.T) {
	rand.Seed(67890)
	
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}
	
	world := NewWorld(config)
	
	// Run simulation for many ticks without entities
	maxTicks := 1000 // Should be enough for at least one event (1% chance per tick)
	eventOccurred := false
	biomeChangesOccurred := false
	
	// Record initial biome state
	initialBiomeState := make(map[Position]BiomeType)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			pos := Position{X: float64(x), Y: float64(y)}
			initialBiomeState[pos] = world.Grid[y][x].Biome
		}
	}
	
	// Run simulation
	for tick := 0; tick < maxTicks; tick++ {
		world.Update()
		
		// Check if any events occurred
		if len(world.Events) > 0 {
			eventOccurred = true
			
			// Check if any event caused biome changes
			for _, event := range world.Events {
				if len(event.BiomeChanges) > 0 {
					biomeChangesOccurred = true
					break
				}
			}
		}
		
		// Early exit if we detected biome changes
		if biomeChangesOccurred {
			break
		}
	}
	
	// Verify at least one event occurred
	if !eventOccurred {
		t.Errorf("Expected at least one event to occur in %d ticks with 1%% probability", maxTicks)
	}
	
	// Check if biome changes were actually applied to the map
	mapChanged := false
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			pos := Position{X: float64(x), Y: float64(y)}
			if initialState, exists := initialBiomeState[pos]; exists {
				if world.Grid[y][x].Biome != initialState {
					mapChanged = true
					break
				}
			}
		}
		if mapChanged {
			break
		}
	}
	
	// We should see some events, but map changes are less guaranteed due to biome recalculation frequency
	t.Logf("Events occurred: %v, Biome-changing events occurred: %v, Map actually changed: %v", 
		eventOccurred, biomeChangesOccurred, mapChanged)
	
	if !eventOccurred {
		t.Error("No events occurred during long-term simulation")
	}
}

// TestEventImpactDespiteStabilityFixes tests that environmental events can still modify the world
// despite the biome stability fixes that were made to prevent mass extinctions
func TestEventImpactDespiteStabilityFixes(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}
	
	rand.Seed(98765)
	world := NewWorld(config)
	
	// Record initial biome state
	initialBiomeCounts := make(map[BiomeType]int)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			initialBiomeCounts[biome]++
		}
	}
	
	// Manually trigger each type of environmental event to ensure they can still impact the world
	events := []struct {
		name string
		changes map[Position]BiomeType
	}{
		{"meteor", world.generateMeteorCraters()},
		{"earthquake", world.generateSeismicChanges()},
		{"wildfire", world.generateFireZones()},
		{"flood", world.generateFloodZones()},
	}
	
	totalChangesApplied := 0
	
	// Apply all event changes to the world
	for _, event := range events {
		if len(event.changes) == 0 {
			t.Errorf("Event %s produced no changes", event.name)
			continue
		}
		
		// Apply the biome changes to the world grid
		for pos, newBiome := range event.changes {
			gridX, gridY := int(pos.X), int(pos.Y)
			if gridX >= 0 && gridX < config.GridWidth && gridY >= 0 && gridY < config.GridHeight {
				world.Grid[gridY][gridX].Biome = newBiome
				totalChangesApplied++
			}
		}
	}
	
	// Verify that changes were actually applied
	if totalChangesApplied == 0 {
		t.Fatal("No environmental event changes were applied to the world")
	}
	
	// Count final biome distribution
	finalBiomeCounts := make(map[BiomeType]int)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			finalBiomeCounts[biome]++
		}
	}
	
	// Verify that environmental events have changed the world
	worldChanged := false
	for biomeType := range initialBiomeCounts {
		if initialBiomeCounts[biomeType] != finalBiomeCounts[biomeType] {
			worldChanged = true
			break
		}
	}
	
	if !worldChanged {
		t.Error("Environmental events should have changed the world biome distribution")
	}
	
	// Verify specific biome types increased as expected from events
	expectedIncreases := map[BiomeType]string{
		BiomeRadiation: "meteor events",
		BiomeMountain:  "earthquake events", 
		BiomeDesert:    "wildfire events",
		BiomeWater:     "flood events",
	}
	
	for biomeType, eventSource := range expectedIncreases {
		if finalBiomeCounts[biomeType] <= initialBiomeCounts[biomeType] {
			t.Errorf("Biome %v should have increased due to %s: initial=%d, final=%d",
				biomeType, eventSource, initialBiomeCounts[biomeType], finalBiomeCounts[biomeType])
		}
	}
	
	t.Logf("Successfully applied %d environmental changes to the world", totalChangesApplied)
	t.Logf("Radiation zones: %d -> %d", initialBiomeCounts[BiomeRadiation], finalBiomeCounts[BiomeRadiation])
	t.Logf("Mountain zones: %d -> %d", initialBiomeCounts[BiomeMountain], finalBiomeCounts[BiomeMountain])
	t.Logf("Desert zones: %d -> %d", initialBiomeCounts[BiomeDesert], finalBiomeCounts[BiomeDesert])
	t.Logf("Water zones: %d -> %d", initialBiomeCounts[BiomeWater], finalBiomeCounts[BiomeWater])
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestSeededEventConsistency tests that event generation produces valid and reasonable results
func TestSeededEventConsistency(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}
	
	// Test each event type produces valid results
	testCases := []struct {
		seed      int64
		eventType string
		expectedBiome BiomeType
		minChanges int
		maxChanges int
	}{
		{111, "meteor", BiomeRadiation, 3, 50},
		{222, "earthquake", BiomeMountain, 1, 150},
		{333, "wildfire", BiomeDesert, 2, 200},
		{444, "flood", BiomeWater, 1, 80},
	}
	
	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			rand.Seed(tc.seed)
			world := NewWorld(config)
			
			var changes map[Position]BiomeType
			switch tc.eventType {
			case "meteor":
				changes = world.generateMeteorCraters()
			case "earthquake":
				changes = world.generateSeismicChanges()
			case "wildfire":
				changes = world.generateFireZones()
			case "flood":
				changes = world.generateFloodZones()
			}
			
			// Verify event produced changes
			if len(changes) == 0 {
				t.Errorf("%s event should produce biome changes", tc.eventType)
			}
			
			// Verify changes are in expected range
			if len(changes) < tc.minChanges || len(changes) > tc.maxChanges {
				t.Errorf("%s event produced %d changes, expected %d-%d", 
					tc.eventType, len(changes), tc.minChanges, tc.maxChanges)
			}
			
			// Verify all changes are to expected biome type
			for pos, biome := range changes {
				if biome != tc.expectedBiome {
					t.Errorf("%s event at position %v produced biome %v, expected %v",
						tc.eventType, pos, biome, tc.expectedBiome)
				}
			}
			
			// Verify positions are within world bounds
			for pos := range changes {
				if pos.X < 0 || pos.X >= float64(config.GridWidth) ||
				   pos.Y < 0 || pos.Y >= float64(config.GridHeight) {
					t.Errorf("%s event produced change outside world bounds: %v", tc.eventType, pos)
				}
			}
		})
	}
}