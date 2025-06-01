package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	// Create world configuration
	worldConfig := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		NumPopulations: 3,
		PopulationSize: 20,
		GridWidth:      40, // Grid cells for visualization
		GridHeight:     25,
	}

	// Create the world
	world := NewWorld(worldConfig)
	// Define predator-prey ecosystem populations
	populations := []PopulationConfig{
		{
			Name:    "Herbivores",
			Species: "herbivore",
			BaseTraits: map[string]float64{
				"size":         -0.5, // Smaller
				"speed":        0.3,  // Moderate speed
				"aggression":   -0.8, // Very peaceful
				"defense":      0.2,  // Some defense
				"cooperation":  0.6,  // Cooperative
				"intelligence": 0.1,  // Basic intelligence
				"endurance":    0.4,  // Good endurance
				"strength":     -0.2, // Weaker
			},
			StartPos:         Position{X: 20, Y: 20},
			Spread:           15.0,
			Color:            "green",
			BaseMutationRate: 0.08, // Low mutation rate - stable species
		},
		{
			Name:    "Predators",
			Species: "predator",
			BaseTraits: map[string]float64{
				"size":         0.8,  // Larger
				"speed":        0.6,  // Fast
				"aggression":   0.9,  // Very aggressive
				"defense":      0.4,  // Good defense
				"cooperation":  -0.2, // Less cooperative
				"intelligence": 0.7,  // Smart hunters
				"endurance":    0.3,  // Lower endurance
				"strength":     0.8,  // Strong
			},
			StartPos:         Position{X: 80, Y: 80},
			Spread:           10.0,
			Color:            "red",
			BaseMutationRate: 0.12, // Higher mutation rate - adaptive hunters
		},
		{
			Name:    "Omnivores",
			Species: "omnivore",
			BaseTraits: map[string]float64{
				"size":         0.0, // Medium size
				"speed":        0.4, // Decent speed
				"aggression":   0.2, // Moderately aggressive
				"defense":      0.5, // Good defense
				"cooperation":  0.3, // Somewhat cooperative
				"intelligence": 0.5, // Intelligent
				"endurance":    0.6, // Good endurance
				"strength":     0.3, // Moderate strength
			},
			StartPos:         Position{X: 50, Y: 20},
			Spread:           12.0,
			Color:            "blue",
			BaseMutationRate: 0.10, // Moderate mutation rate - adaptable
		},
	}

	// Add populations to the world
	for _, popConfig := range populations {
		world.AddPopulation(popConfig)
	}
	// Create and run the CLI
	if err := RunCLI(world); err != nil {
		log.Fatalf("Error running CLI: %v", err)
	}
}
