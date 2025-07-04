package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// createPrimitiveTraits creates a base trait map for primitive organisms with variations
func createPrimitiveTraits(sizeModifier, intelligenceModifier, cooperationModifier float64) map[string]float64 {
	traits := map[string]float64{
		"size":               -1.5 + sizeModifier,
		"speed":              -0.5,
		"aggression":         -0.9,
		"defense":            -0.5,
		"cooperation":        cooperationModifier,
		"intelligence":       -1.0 + intelligenceModifier,
		"endurance":          0.8,
		"strength":           -1.0,
		"aquatic_adaptation": 0.5,
		"digging_ability":    -0.8,
		"underground_nav":    -0.9,
		"flying_ability":     -1.0,
		"altitude_tolerance": -1.0,
		// Biorhythm traits
		"circadian_preference": 0.2,
		"sleep_need":           0.3,
		"hunger_need":          0.5,
		"thirst_need":          0.4,
		"play_drive":           -0.8,
		"exploration_drive":    0.1,
		"scavenging_behavior":  0.6,
	}

	// Apply some variation for "Simple Organisms"
	if intelligenceModifier > 0 {
		traits["speed"] = -0.2
		traits["aggression"] = -0.6
		traits["defense"] = -0.3
		traits["strength"] = -0.8
		traits["aquatic_adaptation"] = 0.2
		traits["digging_ability"] = -0.5
		traits["underground_nav"] = -0.7
		traits["flying_ability"] = -0.9
		traits["altitude_tolerance"] = -0.8
		traits["circadian_preference"] = 0.0
		traits["sleep_need"] = 0.4
		traits["hunger_need"] = 0.4
		traits["thirst_need"] = 0.3
		traits["play_drive"] = -0.5
		traits["exploration_drive"] = 0.3
		traits["scavenging_behavior"] = 0.4
	}

	return traits
}

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
		isoMode    = flag.Bool("iso", false, "Enable 2.5D isometric game view")
		primitive  = flag.Bool("primitive", false, "Start with primitive life forms that can evolve into complex species")
	)

	flag.Parse()

	// Show help
	if *help || *h {
		fmt.Println("Genetic Ecosystem Simulation")
		fmt.Println("============================")
		fmt.Println()
		fmt.Println("A genetic algorithm simulation featuring a complete ecosystem with:")
		fmt.Println("• Multiple species (herbivores, predators, omnivores)")
		fmt.Println("• Primitive life form evolution from simple organisms")
		fmt.Println("• Environment-specific adaptations (aquatic, soil, aerial)")
		fmt.Println("• Plant life system with 6 plant types")
		fmt.Println("• Dynamic biomes and world events")
		fmt.Println("• Evolutionary pressure and species adaptation")
		fmt.Println("• Event logging system")
		fmt.Println("• Tool creation and environmental modification")
		fmt.Println("• Emergent behavior discovery and social learning")
		fmt.Println("• Web interface with real-time visualization")
		fmt.Println("• 2.5D isometric game view with interactive features")
		fmt.Println("• State persistence for save/load functionality")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Primitive Evolution Mode:")
		fmt.Println("  Use --primitive flag to start with basic microbes and simple organisms")
		fmt.Println("  that can evolve into complex species through environmental pressure.")
		fmt.Println("  This mode demonstrates evolution from the ground up.")
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
		fmt.Println("2.5D Isometric View:")
		fmt.Println("  Use --iso flag to enable 2.5D isometric game interface")
		fmt.Println("  Launches web server with isometric view at http://localhost:<port>/iso")
		fmt.Println("  Interactive world with clickable entities and plants")
		fmt.Println("  Real-time DNA and species details on selection")
		fmt.Println("  WASD/Arrow Keys - Camera movement, Mouse Wheel - Zoom")
		fmt.Println("  Left Click - Select entity/plant, Space - Toggle details")
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
	// Initialize random seed (deprecated but functional)
	if *seed == 0 {
		// Use current time for randomness
	} else {
		// For deterministic behavior, we'd need to modify NewWorld to accept a rand source
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
		var populations []PopulationConfig

		if *primitive {
			// Start with primitive life forms that can evolve into complex species
			populations = []PopulationConfig{
				{
					Name:             "Primitive Microbes",
					Species:          "microbe",
					BaseTraits:       createPrimitiveTraits(0.0, 0.0, 0.0),
					StartPos:         Position{X: 30, Y: 30},
					Spread:           25.0, // Widely spread
					Color:            "gray",
					BaseMutationRate: 0.25, // Very high mutation rate for rapid evolution
				},
				{
					Name:             "Simple Organisms",
					Species:          "simple",
					BaseTraits:       createPrimitiveTraits(0.5, 0.3, 0.1), // Slightly larger, smarter, more cooperative
					StartPos:         Position{X: 70, Y: 40},
					Spread:           20.0,
					Color:            "yellow",
					BaseMutationRate: 0.20, // High mutation rate for evolution
				},
			}
		} else {
			// Define predator-prey ecosystem populations only if not loading state
			populations = []PopulationConfig{
				{
					Name:    "Herbivores",
					Species: "herbivore",
					BaseTraits: map[string]float64{
						"size":               -0.5, // Smaller
						"speed":              0.3,  // Moderate speed
						"aggression":         -0.8, // Very peaceful
						"defense":            0.2,  // Some defense
						"cooperation":        0.6,  // Cooperative
						"intelligence":       0.1,  // Basic intelligence
						"endurance":          0.4,  // Good endurance
						"strength":           -0.2, // Weaker
						"aquatic_adaptation": -0.5, // Poor in water initially
						"digging_ability":    0.1,  // Basic digging
						"underground_nav":    -0.3, // Poor underground navigation
						"flying_ability":     -0.8, // Cannot fly
						"altitude_tolerance": -0.6, // Poor at altitude
						// Biorhythm traits
						"circadian_preference": 0.7, // Strongly diurnal (active during day)
						"sleep_need":           0.2, // Lower sleep requirement (grazing animals)
						"hunger_need":          0.8, // High hunger needs (constant grazing)
						"thirst_need":          0.6, // High water needs
						"play_drive":           0.3, // Some play behavior (social animals)
						"exploration_drive":    0.5, // Moderate exploration for food
						"scavenging_behavior":  0.1, // Minimal scavenging (prefer fresh plants)
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
						"size":               0.8,  // Larger
						"speed":              0.6,  // Fast
						"aggression":         0.9,  // Very aggressive
						"defense":            0.4,  // Good defense
						"cooperation":        -0.2, // Less cooperative
						"intelligence":       0.7,  // Smart hunters
						"endurance":          0.3,  // Lower endurance
						"strength":           0.8,  // Strong
						"aquatic_adaptation": -0.2, // Somewhat poor in water
						"digging_ability":    0.0,  // Average digging
						"underground_nav":    0.2,  // Decent underground navigation
						"flying_ability":     -0.5, // Poor flying ability
						"altitude_tolerance": 0.1,  // Slightly better at altitude
						// Biorhythm traits
						"circadian_preference": -0.6, // Nocturnal (hunt at night)
						"sleep_need":           0.4,  // Moderate sleep needs (conserve energy)
						"hunger_need":          0.3,  // Lower hunger frequency (large meals)
						"thirst_need":          0.2,  // Lower water needs
						"play_drive":           -0.3, // Limited play (focus on survival)
						"exploration_drive":    0.8,  // High exploration (hunting territory)
						"scavenging_behavior":  0.7,  // High scavenging behavior
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
						"size":               0.0,  // Medium size
						"speed":              0.4,  // Decent speed
						"aggression":         0.2,  // Moderately aggressive
						"defense":            0.5,  // Good defense
						"cooperation":        0.3,  // Somewhat cooperative
						"intelligence":       0.5,  // Intelligent
						"endurance":          0.6,  // Good endurance
						"strength":           0.3,  // Moderate strength
						"aquatic_adaptation": 0.1,  // Slightly adapted to water
						"digging_ability":    0.2,  // Good digging ability
						"underground_nav":    0.1,  // Basic underground navigation
						"flying_ability":     -0.3, // Limited flying ability
						"altitude_tolerance": 0.0,  // Average altitude tolerance
						// Biorhythm traits
						"circadian_preference": 0.3, // Slightly diurnal but adaptable
						"sleep_need":           0.3, // Moderate sleep needs
						"hunger_need":          0.6, // High hunger (active foragers)
						"thirst_need":          0.5, // Moderate water needs
						"play_drive":           0.6, // High play behavior (intelligent species)
						"exploration_drive":    0.7, // High exploration (opportunistic)
						"scavenging_behavior":  0.8, // Very high scavenging (opportunistic feeders)
					},
					StartPos:         Position{X: 50, Y: 20},
					Spread:           12.0,
					Color:            "blue",
					BaseMutationRate: 0.10, // Moderate mutation rate - adaptable
				},
			}
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
	} else if *isoMode {
		// Create and run the web interface with isometric view
		fmt.Printf("Starting isometric 2.5D interface on http://localhost:%d/iso\n", *webPort)
		if err := RunWebInterface(world, *webPort); err != nil {
			log.Fatalf("Error running isometric interface: %v", err)
		}
	} else {
		// Create and run the CLI
		if err := RunCLI(world); err != nil {
			log.Fatalf("Error running CLI: %v", err)
		}
	}
}
