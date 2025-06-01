package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Define command-line flags
	var (
		help       = flag.Bool("help", false, "Show help message")
		h          = flag.Bool("h", false, "Show help message (short)")
		width      = flag.Float64("width", 100.0, "World width")
		height     = flag.Float64("height", 100.0, "World height")
		gridWidth  = flag.Int("grid-width", 40, "Grid cells width for visualization")
		gridHeight = flag.Int("grid-height", 25, "Grid cells height for visualization")
		popSize    = flag.Int("pop-size", 20, "Population size per species")
		seed       = flag.Int64("seed", 0, "Random seed (0 for current time)")
		version    = flag.Bool("version", false, "Show version information")
	)

	flag.Parse()

	// Show help
	if *help || *h {
		fmt.Println("Genetic Ecosystem Simulation")
		fmt.Println("============================")
		fmt.Println()
		fmt.Println("A genetic algorithm simulation featuring a complete ecosystem with:")
		fmt.Println("• Multiple species (herbivores, predators, omnivores)")
		fmt.Println("• Plant life system with 6 plant types")
		fmt.Println("• Dynamic biomes and world events")
		fmt.Println("• Evolutionary pressure and species adaptation")
		fmt.Println("• Event logging system")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Controls (in simulation):")
		fmt.Println("  space      Pause/Resume simulation")
		fmt.Println("  enter      Manual step (when paused)")
		fmt.Println("  v          Cycle through views (grid/stats/events/populations)")
		fmt.Println("  ?          Toggle help screen")
		fmt.Println("  q          Quit")
		fmt.Println()
		fmt.Println("The simulation will display a real-time grid showing entities, plants,")
		fmt.Println("and biomes. Different symbols represent different species and plant types.")
		fmt.Println("Check the in-simulation help (?) for detailed symbol meanings.")
		return
	}

	// Show version
	if *version {
		fmt.Println("Genetic Ecosystem Simulation v1.0")
		fmt.Println("Enhanced with plant life, event logging, and evolutionary pressure")
		return
	}
	// Initialize random seed
	if *seed == 0 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(*seed)
		fmt.Printf("Using random seed: %d\n", *seed)
	}

	// Create world configuration
	worldConfig := WorldConfig{
		Width:          *width,
		Height:         *height,
		NumPopulations: 3,
		PopulationSize: *popSize,
		GridWidth:      *gridWidth,
		GridHeight:     *gridHeight,
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
