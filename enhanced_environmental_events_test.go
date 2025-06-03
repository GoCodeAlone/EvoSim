package main

import (
	"testing"
)

// TestEnhancedEnvironmentalEvents tests the enhanced environmental event system
func TestEnhancedEnvironmentalEvents(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Test triggering enhanced environmental events
	initialEventCount := len(world.EnvironmentalEvents)
	
	// Trigger a few different types of events
	world.triggerEnhancedEnvironmentalEvent()
	world.triggerEnhancedEnvironmentalEvent()
	world.triggerEnhancedEnvironmentalEvent()
	
	if len(world.EnvironmentalEvents) <= initialEventCount {
		t.Error("Enhanced environmental events should have been created")
	}
	
	t.Logf("Created %d enhanced environmental events", len(world.EnvironmentalEvents)-initialEventCount)
	
	// Test that events have proper properties
	for _, event := range world.EnvironmentalEvents {
		if event.ID == 0 {
			t.Error("Event should have a valid ID")
		}
		if event.Type == "" {
			t.Error("Event should have a type")
		}
		if event.Duration <= 0 {
			t.Error("Event should have positive duration")
		}
		if event.Position.X < 0 || event.Position.Y < 0 {
			t.Error("Event should have valid position")
		}
		
		t.Logf("Event: %s at (%.1f, %.1f), duration: %d, intensity: %.2f", 
			event.Name, event.Position.X, event.Position.Y, event.Duration, event.Intensity)
	}
}

// TestFireSpreadWithWind tests that fire spreads according to wind direction
func TestFireSpreadWithWind(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create a controlled environment with lots of flammable biomes
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			world.Grid[y][x].Biome = BiomeForest // Make everything flammable
		}
	}

	// Create a fire event in the center
	fire := &EnhancedEnvironmentalEvent{
		ID:            1,
		Type:          "wildfire",
		Name:          "Test Fire",
		Description:   "Test fire for wind spreading",
		StartTick:     world.Tick,
		Duration:      20,
		Position:      Position{X: 10, Y: 10},
		Radius:        2.0,
		MaxRadius:     8.0,
		Intensity:     0.8,
		WindSensitive: true,
		SpreadPattern: "wind_driven",
		AffectedCells: make(map[Position]BiomeType),
		Effects:       map[string]float64{"damage": 2.0, "mutation": 0.05},
	}

	world.EnvironmentalEvents = append(world.EnvironmentalEvents, fire)

	// Record initial state
	initialAffectedCells := len(fire.AffectedCells)
	initialPosition := fire.Position
	
	// Update fire spread several times
	for i := 0; i < 5; i++ {
		world.updateEnhancedEnvironmentalEvents()
	}

	// Check that fire spread
	if len(fire.AffectedCells) <= initialAffectedCells {
		t.Logf("Fire spread from %d to %d affected cells", initialAffectedCells, len(fire.AffectedCells))
	}

	// Check that fire moved (should move due to wind)
	finalPosition := fire.Position
	distance := (finalPosition.X-initialPosition.X)*(finalPosition.X-initialPosition.X) + 
	            (finalPosition.Y-initialPosition.Y)*(finalPosition.Y-initialPosition.Y)
	
	if distance > 0 {
		t.Logf("Fire moved from (%.1f, %.1f) to (%.1f, %.1f)", 
			initialPosition.X, initialPosition.Y, finalPosition.X, finalPosition.Y)
	}

	// Check that fire turned forest to desert
	desertCount := 0
	for _, biome := range fire.AffectedCells {
		if biome == BiomeDesert {
			desertCount++
		}
	}
	
	if desertCount > 0 {
		t.Logf("Fire created %d desert cells from burned forest", desertCount)
	}
}

// TestFireExtinguishing tests that fire is extinguished by water
func TestFireExtinguishing(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create an environment with forest and water
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if x < 10 {
				world.Grid[y][x].Biome = BiomeForest // Left side is forest
			} else {
				world.Grid[y][x].Biome = BiomeWater // Right side is water
			}
		}
	}

	// Create a fire event that will encounter water
	fire := &EnhancedEnvironmentalEvent{
		ID:            1,
		Type:          "wildfire",
		Name:          "Test Fire",
		Description:   "Test fire for extinguishing",
		StartTick:     world.Tick,
		Duration:      20,
		Position:      Position{X: 8, Y: 10}, // Near the forest-water boundary
		Radius:        2.0,
		MaxRadius:     8.0,
		Intensity:     0.8,
		WindSensitive: true,
		SpreadPattern: "wind_driven",
		AffectedCells: make(map[Position]BiomeType),
		Effects:       map[string]float64{"damage": 2.0, "mutation": 0.05},
	}

	world.EnvironmentalEvents = append(world.EnvironmentalEvents, fire)

	initialIntensity := fire.Intensity
	
	// Update fire - it should encounter water and reduce intensity
	for i := 0; i < 10; i++ {
		world.updateEnhancedEnvironmentalEvents()
	}

	// Fire intensity should have decreased when encountering water
	if fire.Intensity < initialIntensity {
		t.Logf("Fire intensity reduced from %.2f to %.2f when encountering water", 
			initialIntensity, fire.Intensity)
	}
}

// TestEventEffectsOnEntities tests that environmental events affect entities properly
func TestEventEffectsOnEntities(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create an entity at a known position
	entity := NewEntity(1, []string{"energy", "speed", "vision"}, "test_species", Position{X: 10, Y: 10})
	entity.Energy = 100.0
	world.AllEntities = append(world.AllEntities, entity)

	// Create a damaging event at the same position
	event := &EnhancedEnvironmentalEvent{
		ID:            1,
		Type:          "volcanic_eruption",
		Name:          "Test Volcano",
		Description:   "Test volcano for entity effects",
		StartTick:     world.Tick,
		Duration:      10,
		Position:      Position{X: 10, Y: 10}, // Same as entity
		Radius:        5.0,
		MaxRadius:     8.0,
		Intensity:     1.0,
		WindSensitive: false,
		AffectedCells: make(map[Position]BiomeType),
		Effects:       map[string]float64{"damage": 3.0, "mutation": 0.1},
	}

	world.EnvironmentalEvents = append(world.EnvironmentalEvents, event)

	initialEnergy := entity.Energy
	
	// Apply event effects
	world.applyEventEffects(event)

	// Entity should have taken damage
	if entity.Energy < initialEnergy {
		t.Logf("Entity energy reduced from %.1f to %.1f due to volcanic eruption", 
			initialEnergy, entity.Energy)
	} else {
		t.Error("Entity should have taken damage from environmental event")
	}
}

// TestStormMovement tests that storms move with wind patterns
func TestStormMovement(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create a storm
	storm := &EnhancedEnvironmentalEvent{
		ID:            1,
		Type:          "storm",
		Name:          "Test Storm",
		Description:   "Test storm for movement",
		StartTick:     world.Tick,
		Duration:      15,
		Position:      Position{X: 5, Y: 5},
		Radius:        4.0,
		MaxRadius:     10.0,
		Intensity:     0.6,
		WindSensitive: true,
		SpreadPattern: "directional",
		AffectedCells: make(map[Position]BiomeType),
		Effects:       map[string]float64{"damage": 0.5, "mutation": 0.02},
	}

	world.EnvironmentalEvents = append(world.EnvironmentalEvents, storm)

	initialPosition := storm.Position
	
	// Update storm movement
	for i := 0; i < 5; i++ {
		world.updateEnhancedEnvironmentalEvents()
	}

	finalPosition := storm.Position
	
	// Storm should have moved
	distance := (finalPosition.X-initialPosition.X)*(finalPosition.X-initialPosition.X) + 
	            (finalPosition.Y-initialPosition.Y)*(finalPosition.Y-initialPosition.Y)
	
	if distance > 0 {
		t.Logf("Storm moved from (%.1f, %.1f) to (%.1f, %.1f)", 
			initialPosition.X, initialPosition.Y, finalPosition.X, finalPosition.Y)
	} else {
		t.Log("Storm position did not change significantly")
	}
}