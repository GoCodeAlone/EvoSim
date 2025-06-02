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
		loadState  = flag.String("load", "", "Load simulation state from file")
		saveState  = flag.String("save", "", "Save simulation state to file and exit")
		webMode    = flag.Bool("web", false, "Enable web interface mode")
		webPort    = flag.Int("web-port", 8080, "Port for web interface")
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
		fmt.Println("• Tool creation and environmental modification")
		fmt.Println("• Emergent behavior discovery and social learning")
		fmt.Println("• Web interface with real-time visualization")
		fmt.Println("• State persistence for save/load functionality")
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
		fmt.Println("Web Interface:")
		fmt.Println("  Use --web flag to enable web interface mode")
		fmt.Println("  Access via browser at http://localhost:<port> (default: 8080)")
		fmt.Println("  All 14 view modes available with real-time updates")
		fmt.Println("  WebSocket-based live simulation streaming")
		fmt.Println()
		fmt.Println("State Management:")
		fmt.Println("  --save <file>   Save simulation state to JSON file")
		fmt.Println("  --load <file>   Load simulation state from JSON file")
		fmt.Println("  State includes all entities, tools, behaviors, and environment")
		fmt.Println()
		fmt.Println("The simulation will display a real-time grid showing entities, plants,")
		fmt.Println("biomes, tools, and environmental modifications. Different symbols represent")
		fmt.Println("different species and plant types. Check the in-simulation help (?) for")
		fmt.Println("detailed symbol meanings.")
		return
	}

	// Show version
	if *version {
		fmt.Println("Genetic Ecosystem Simulation v2.0")
		fmt.Println("Enhanced with:")
		fmt.Println("• Plant life, event logging, and evolutionary pressure")
		fmt.Println("• Tool creation and environmental modification systems")
		fmt.Println("• Emergent behavior discovery and social learning")
		fmt.Println("• Web interface with real-time visualization")
		fmt.Println("• Complete state persistence functionality")
		fmt.Println("• Underground plant networks and wind dispersal")
		fmt.Println("• DNA/RNA genetic systems and cellular evolution")
		fmt.Println("• Species formation and macro evolution tracking")
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
	
	// Create state manager
	stateManager := NewStateManager(world)
	
	// Load state if specified
	if *loadState != "" {
		err := stateManager.LoadFromFile(*loadState)
		if err != nil {
			log.Fatalf("Error loading state: %v", err)
		}
	} else {
		// Define predator-prey ecosystem populations only if not loading state
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
	}
	
	// Save state if specified and exit
	if *saveState != "" {
		err := stateManager.SaveToFile(*saveState)
		if err != nil {
			log.Fatalf("Error saving state: %v", err)
		}
		return
	}
	// Run the interface
	if *webMode {
		// Create and run the web interface
		if err := RunWebInterface(world, *webPort); err != nil {
			log.Fatalf("Error running web interface: %v", err)
		}
	} else {
		// Create and run the CLI
		if err := RunCLI(world); err != nil {
			log.Fatalf("Error running CLI: %v", err)
		}
	}
}
