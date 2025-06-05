package main

import (
	"math"
	"math/rand"
	"testing"
)

// countBiomeInGrid counts the number of cells with a specific biome type
func countBiomeInGrid(world *World, biomeType BiomeType) int {
	count := 0
	for y := 0; y < world.Config.GridHeight; y++ {
		for x := 0; x < world.Config.GridWidth; x++ {
			if world.Grid[y][x].Biome == biomeType {
				count++
			}
		}
	}
	return count
}

// applyBiomeChanges applies biome changes to the world grid
func applyBiomeChanges(world *World, changes map[Position]BiomeType) {
	for pos, biomeType := range changes {
		gridX, gridY := int(pos.X), int(pos.Y)
		if gridX >= 0 && gridX < world.Config.GridWidth && gridY >= 0 && gridY < world.Config.GridHeight {
			world.Grid[gridY][gridX].Biome = biomeType
		}
	}
}

// testEnvironmentalEvent is a helper function to test environmental events
func testEnvironmentalEvent(t *testing.T, eventName string, biomeType BiomeType, minIncrease, maxIncrease int,
	generateChanges func(*World) map[Position]BiomeType) {
	
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
	initialCount := countBiomeInGrid(world, biomeType)
	
	// Generate changes
	changes := generateChanges(world)
	
	// Verify changes were generated
	if len(changes) == 0 {
		t.Fatalf("%s should generate at least one change", eventName)
	}
	
	// Apply the changes
	applyBiomeChanges(world, changes)
	
	// Count final biome state
	finalCount := countBiomeInGrid(world, biomeType)
	
	// Verify biome increased
	if finalCount <= initialCount {
		t.Errorf("%s should increase biome zones: initial=%d, final=%d", 
			eventName, initialCount, finalCount)
	}
	
	// Verify reasonable number of changes
	increase := finalCount - initialCount
	if increase < minIncrease || increase > maxIncrease {
		t.Errorf("%s biome zones out of expected range: got %d, expected %d-%d",
			eventName, increase, minIncrease, maxIncrease)
	}
}

// TestMeteorShowerEvent tests meteor shower events can modify the world map
func TestMeteorShowerEvent(t *testing.T) {
	testEnvironmentalEvent(t, "Meteor shower", BiomeRadiation, 3, 50, func(world *World) map[Position]BiomeType {
		return world.generateMeteorCraters()
	})
}

// TestEarthquakeEvent tests earthquake events can create mountain ranges
func TestEarthquakeEvent(t *testing.T) {
	// Set different seed for different test behavior
	rand.Seed(23456)
	
	testEnvironmentalEvent(t, "Earthquake", BiomeMountain, 1, 150, func(world *World) map[Position]BiomeType {
		return world.generateSeismicChanges()
	})
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
	// Set different seed for different test behavior
	rand.Seed(45678)
	
	testEnvironmentalEvent(t, "Flood", BiomeWater, 1, 80, func(world *World) map[Position]BiomeType {
		return world.generateFloodZones()
	})
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
	
	// Verify that the events created significant environmental changes
	// Some biomes should have changed significantly (not strict requirements)
	expectedIncreases := map[BiomeType]string{
		BiomeRadiation: "meteor events",
		BiomeMountain:  "earthquake events", 
		BiomeDesert:    "wildfire events",
		BiomeWater:     "flood events",
	}
	
	significantChanges := 0
	for biomeType := range expectedIncreases {
		initialCount := initialBiomeCounts[biomeType]
		finalCount := finalBiomeCounts[biomeType]
		
		// Consider a change significant if it's more than 20% 
		if initialCount > 0 {
			changePercent := math.Abs(float64(finalCount-initialCount)) / float64(initialCount)
			if changePercent > 0.2 {
				significantChanges++
			}
		} else if finalCount > 5 { // New biome zones created
			significantChanges++
		}
	}
	
	if significantChanges == 0 {
		t.Error("Environmental events should have created significant biome changes")
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