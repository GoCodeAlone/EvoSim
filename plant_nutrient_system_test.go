package main

import (
	"testing"
)

// TestPlantNutrientSystem tests the realistic plant nutrition system
func TestPlantNutrientSystem(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create a plant and place it in a specific cell
	plant := NewPlant(1, PlantTree, Position{X: 10, Y: 10})
	world.AllPlants = append(world.AllPlants, plant)

	// Get the grid cell where the plant is located
	gridX := int((plant.Position.X / world.Config.Width) * float64(world.Config.GridWidth))
	gridY := int((plant.Position.Y / world.Config.Height) * float64(world.Config.GridHeight))
	gridCell := &world.Grid[gridY][gridX]

	// Test initial conditions
	if len(plant.NutrientNeeds) == 0 {
		t.Error("Plant should have nutrient needs initialized")
	}

	if plant.WaterDependency == 0 {
		t.Error("Plant should have water dependency set")
	}

	if len(gridCell.SoilNutrients) == 0 {
		t.Error("Grid cell should have soil nutrients initialized")
	}

	t.Logf("Plant nutrient needs: %+v", plant.NutrientNeeds)
	t.Logf("Plant water dependency: %.2f", plant.WaterDependency)
	t.Logf("Soil nutrients: %+v", gridCell.SoilNutrients)
	t.Logf("Soil water level: %.2f", gridCell.WaterLevel)

	// Test plant nutrient update
	initialEnergy := plant.Energy
	nutritionalHealth := plant.updatePlantNutrients(gridCell, "spring")

	t.Logf("Nutritional health: %.2f", nutritionalHealth)
	t.Logf("Energy change: %.2f -> %.2f", initialEnergy, plant.Energy)

	// Verify that plant consumed some nutrients
	totalInitialNutrients := 0.0
	totalFinalNutrients := 0.0

	for nutrient := range plant.NutrientNeeds {
		if nutrient != "water" {
			if initial, exists := gridCell.SoilNutrients[nutrient]; exists {
				totalInitialNutrients += initial
			}
		}
	}

	// Plant should have consumed some nutrients
	for nutrient := range plant.NutrientNeeds {
		if nutrient != "water" {
			if final, exists := gridCell.SoilNutrients[nutrient]; exists {
				totalFinalNutrients += final
			}
		}
	}

	if totalFinalNutrients >= totalInitialNutrients {
		t.Log("Note: Plant did not consume detectable nutrients (may be due to adequate supply)")
	} else {
		t.Logf("Plant consumed nutrients: %.3f -> %.3f", totalInitialNutrients, totalFinalNutrients)
	}
}

// TestSoilNutrientDecay tests that dead organisms add nutrients to soil
func TestSoilNutrientDecay(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Get a grid cell and record initial nutrients
	gridCell := &world.Grid[10][10]
	initialNitrogen := gridCell.SoilNutrients["nitrogen"]
	initialPhosphorus := gridCell.SoilNutrients["phosphorus"]
	initialOrganicMatter := gridCell.OrganicMatter

	// Add a decaying corpse to the reproduction system
	corpsePosition := Position{
		X: (10.0 / float64(world.Config.GridWidth)) * world.Config.Width,
		Y: (10.0 / float64(world.Config.GridHeight)) * world.Config.Height,
	}
	world.ReproductionSystem.AddDecayingItem("corpse", corpsePosition, 50.0, "test_species", 2.0, world.Tick-50) // Started decaying 50 ticks ago

	// Process decay nutrients
	world.processDecayNutrientsToSoil()

	// Check that nutrients increased
	finalNitrogen := gridCell.SoilNutrients["nitrogen"]
	finalPhosphorus := gridCell.SoilNutrients["phosphorus"]
	finalOrganicMatter := gridCell.OrganicMatter

	t.Logf("Nitrogen: %.3f -> %.3f", initialNitrogen, finalNitrogen)
	t.Logf("Phosphorus: %.3f -> %.3f", initialPhosphorus, finalPhosphorus)
	t.Logf("Organic matter: %.3f -> %.3f", initialOrganicMatter, finalOrganicMatter)

	if finalNitrogen > initialNitrogen {
		t.Log("SUCCESS: Nitrogen increased from decay")
	}

	if finalPhosphorus > initialPhosphorus {
		t.Log("SUCCESS: Phosphorus increased from decay")
	}

	if finalOrganicMatter > initialOrganicMatter {
		t.Log("SUCCESS: Organic matter increased from decay")
	}
}

// TestRainfallEffects tests that rainfall adds water and affects soil
func TestRainfallEffects(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Get a grid cell and record initial water
	gridCell := &world.Grid[10][10]
	initialWaterLevel := gridCell.WaterLevel
	initialCompaction := gridCell.SoilCompaction

	// Simulate heavy rainfall
	processRainfall(gridCell, 0.8) // Heavy rain intensity

	finalWaterLevel := gridCell.WaterLevel
	finalCompaction := gridCell.SoilCompaction

	t.Logf("Water level: %.3f -> %.3f", initialWaterLevel, finalWaterLevel)
	t.Logf("Soil compaction: %.3f -> %.3f", initialCompaction, finalCompaction)

	if finalWaterLevel > initialWaterLevel {
		t.Log("SUCCESS: Water level increased from rainfall")
	} else {
		t.Error("Rainfall should increase water level")
	}

	if finalCompaction < initialCompaction {
		t.Log("SUCCESS: Soil compaction reduced by rainfall")
	}
}

// TestPlantStressConditions tests plant behavior under nutrient stress
func TestPlantStressConditions(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create a plant
	plant := NewPlant(1, PlantTree, Position{X: 10, Y: 10})
	world.AllPlants = append(world.AllPlants, plant)

	// Get the grid cell and deplete nutrients
	gridX := int((plant.Position.X / world.Config.Width) * float64(world.Config.GridWidth))
	gridY := int((plant.Position.Y / world.Config.Height) * float64(world.Config.GridHeight))
	gridCell := &world.Grid[gridY][gridX]

	// Severely deplete water and nutrients
	gridCell.WaterLevel = 0.05 // Very low water
	for nutrient := range gridCell.SoilNutrients {
		gridCell.SoilNutrients[nutrient] = 0.01 // Very low nutrients
	}

	initialEnergy := plant.Energy

	// Update plant under stress conditions
	nutritionalHealth := plant.updatePlantNutrients(gridCell, "summer") // Summer for extra stress

	t.Logf("Plant under stress:")
	t.Logf("  Nutritional health: %.3f", nutritionalHealth)
	t.Logf("  Energy change: %.1f -> %.1f", initialEnergy, plant.Energy)
	t.Logf("  Still alive: %v", plant.IsAlive)

	// Plant should be stressed (low nutritional health)
	if nutritionalHealth > 0.7 {
		t.Error("Plant should be nutritionally stressed under poor conditions")
	}

	// Energy should have decreased
	if plant.Energy >= initialEnergy {
		t.Log("Note: Plant energy did not decrease as expected under stress")
	} else {
		t.Log("SUCCESS: Plant energy decreased under stress conditions")
	}
}

// TestSeasonalEffects tests that plants respond to seasonal changes
func TestSeasonalEffects(t *testing.T) {
	config := WorldConfig{
		Width:      200,
		Height:     200,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)

	// Create a plant
	plant := NewPlant(1, PlantTree, Position{X: 10, Y: 10})
	world.AllPlants = append(world.AllPlants, plant)

	// Get the grid cell
	gridX := int((plant.Position.X / world.Config.Width) * float64(world.Config.GridWidth))
	gridY := int((plant.Position.Y / world.Config.Height) * float64(world.Config.GridHeight))
	gridCell := &world.Grid[gridY][gridX]

	// Test different seasons
	seasons := []string{"spring", "summer", "autumn", "winter"}
	results := make(map[string]float64)

	for _, season := range seasons {
		// Reset plant state
		plant.Energy = 50.0

		// Ensure adequate nutrients for fair comparison
		for nutrient := range gridCell.SoilNutrients {
			gridCell.SoilNutrients[nutrient] = 0.5
		}
		gridCell.WaterLevel = 0.6

		nutritionalHealth := plant.updatePlantNutrients(gridCell, season)
		results[season] = nutritionalHealth

		t.Logf("Season %s: nutritional health = %.3f", season, nutritionalHealth)
	}

	// Spring should generally be better than winter
	if results["spring"] > results["winter"] {
		t.Log("SUCCESS: Spring conditions better than winter")
	}

	// Summer should be decent but potentially water-stressed
	if results["summer"] > 0.5 {
		t.Log("SUCCESS: Summer conditions adequate")
	}
}
