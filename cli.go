package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CLI model for the ecosystem simulation
type CLIModel struct {
	world          *World
	width          int
	height         int
	tick           int
	paused         bool
	showHelp       bool
	selectedView   string
	viewModes      []string
	autoAdvance    bool
	lastUpdateTime time.Time
	speciesColors  map[string]string
	speciesSymbols map[string]rune
	// Viewport controls for navigation
	viewportX int
	viewportY int
	zoomLevel int
	// Interactive features
	selectedEntity *Entity
	followEntity   bool
	showSignals    bool
	showStructures bool
	showPhysics    bool
	showTime       bool
}

// tickMsg represents an auto-advance tick
type tickMsg time.Time

// Key bindings
var keys = struct {
	up         key.Binding
	down       key.Binding
	left       key.Binding
	right      key.Binding
	enter      key.Binding
	space      key.Binding
	help       key.Binding
	quit       key.Binding
	view       key.Binding
	auto       key.Binding
	zoom       key.Binding
	reset      key.Binding
	signals    key.Binding
	structures key.Binding
	physics    key.Binding
	export     key.Binding
}{
	up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up/pan up"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down/pan down"),
	),
	left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "pan left"),
	),
	right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "pan right"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "step"),
	),
	space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "pause/resume"),
	),
	help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	view: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "cycle view"),
	),
	auto: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle auto"),
	),
	zoom: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "zoom"),
	),
	reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset view"),
	),
	signals: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle signals"),
	),
	structures: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle structures"),
	),
	physics: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "toggle physics"),
	),
	export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export data"),
	),
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	gridStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	eventStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Padding(0, 1).
			Bold(true)

	biomeColors = map[BiomeType]lipgloss.Style{
		BiomePlains:    lipgloss.NewStyle().Foreground(lipgloss.Color("34")),  // Green
		BiomeForest:    lipgloss.NewStyle().Foreground(lipgloss.Color("28")),  // Dark Green
		BiomeDesert:    lipgloss.NewStyle().Foreground(lipgloss.Color("220")), // Yellow
		BiomeMountain:  lipgloss.NewStyle().Foreground(lipgloss.Color("244")), // Gray
		BiomeWater:     lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
		BiomeRadiation: lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
	}

	speciesStyles = map[string]lipgloss.Style{
		"herbivore": lipgloss.NewStyle().Foreground(lipgloss.Color("46")),  // Bright Green
		"predator":  lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		"omnivore":  lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
	}
)

// NewCLIModel creates a new CLI model
func NewCLIModel(world *World) CLIModel {
	speciesColors := map[string]string{
		"herbivore": "green",
		"predator":  "red",
		"omnivore":  "orange",
	}

	speciesSymbols := map[string]rune{
		"herbivore": '●',
		"predator":  '▲',
		"omnivore":  '◆',
	}
	return CLIModel{world: world,
		viewModes:      []string{"grid", "stats", "events", "populations", "communication", "civilization", "physics", "wind", "species", "network", "dna", "cellular", "evolution", "topology", "tools", "environment", "behavior", "reproduction", "statistical", "anomalies", "warfare"},
		selectedView:   "grid",
		autoAdvance:    true,
		lastUpdateTime: time.Now(),
		speciesColors:  speciesColors,
		speciesSymbols: speciesSymbols,
		viewportX:      0,
		viewportY:      0,
		zoomLevel:      1,
		showSignals:    true,
		showStructures: true,
		showPhysics:    false,
		showTime:       true,
	}
}

// doTick schedules the next automatic update
func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model
func (m CLIModel) Init() tea.Cmd {
	return doTick()
}

// Update handles messages
func (m CLIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.quit):
			return m, tea.Quit

		case key.Matches(msg, keys.help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, keys.space):
			m.paused = !m.paused

		case key.Matches(msg, keys.auto):
			m.autoAdvance = !m.autoAdvance

		case key.Matches(msg, keys.view):
			// Cycle through view modes
			for i, mode := range m.viewModes {
				if mode == m.selectedView {
					m.selectedView = m.viewModes[(i+1)%len(m.viewModes)]
					break
				}
			}

		case key.Matches(msg, keys.enter):
			// Manual step forward
			m.world.Update()
			m.tick++

		case key.Matches(msg, keys.left):
			if m.viewportX > 0 {
				m.viewportX--
			}

		case key.Matches(msg, keys.right):
			if m.viewportX < m.world.Config.GridWidth-20 {
				m.viewportX++
			}

		case key.Matches(msg, keys.up):
			if m.viewportY > 0 {
				m.viewportY--
			}

		case key.Matches(msg, keys.down):
			if m.viewportY < m.world.Config.GridHeight-15 {
				m.viewportY++
			}

		case key.Matches(msg, keys.zoom):
			m.zoomLevel = (m.zoomLevel % 3) + 1

		case key.Matches(msg, keys.reset):
			m.viewportX = 0
			m.viewportY = 0
			m.zoomLevel = 1
			m.selectedEntity = nil
			m.followEntity = false

		case key.Matches(msg, keys.signals):
			m.showSignals = !m.showSignals

		case key.Matches(msg, keys.structures):
			m.showStructures = !m.showStructures

		case key.Matches(msg, keys.physics):
			m.showPhysics = !m.showPhysics

		case key.Matches(msg, keys.export):
			// Export statistical data
			m.exportStatisticalData()
		}

	case tickMsg:
		if m.autoAdvance && !m.paused {
			m.world.Update()
			m.tick++
		}
		cmd = doTick()
	}

	return m, cmd
}

// View renders the interface
func (m CLIModel) View() string {
	if m.showHelp {
		return m.helpView()
	}
	var content string
	switch m.selectedView {
	case "grid":
		content = m.gridView()
	case "stats":
		content = m.statsView()
	case "events":
		content = m.eventsView()
	case "populations":
		content = m.populationsView()
	case "communication":
		content = m.communicationView()
	case "civilization":
		content = m.civilizationView()
	case "physics":
		content = m.physicsView()
	case "wind":
		content = m.windView()
	case "species":
		content = m.speciesView()
	case "network":
		content = m.networkView()
	case "dna":
		content = m.dnaView()
	case "cellular":
		content = m.cellularView()
	case "evolution":
		content = m.evolutionView()
	case "topology":
		content = m.topologyView()
	case "tools":
		content = m.toolsView()
	case "environment":
		content = m.environmentView()
	case "behavior":
		content = m.behaviorView()
	case "reproduction":
		content = m.reproductionView()
	case "statistical":
		content = m.statisticalView()
	case "anomalies":
		content = m.anomaliesView()
	case "warfare":
		content = m.warfareView()
	default:
		content = m.gridView()
	}

	// Header with world clock and status
	header := m.headerView()

	// Footer with controls
	footer := m.footerView()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// headerView renders the header with world info
func (m CLIModel) headerView() string {
	worldTime := m.world.Clock.Format("15:04 Day 2006-01-02")

	status := "▶ RUNNING"
	if m.paused {
		status = "⏸ PAUSED"
	}

	entities := len(m.world.AllEntities)
	populations := len(m.world.Populations)
	activeEvents := len(m.world.Events)

	// Advanced system indicators
	var indicators []string

	// Communication system indicator
	if m.world.CommunicationSystem != nil && len(m.world.CommunicationSystem.Signals) > 0 {
		indicators = append(indicators, fmt.Sprintf("📡%d", len(m.world.CommunicationSystem.Signals)))
	}

	// Civilization system indicator
	if m.world.CivilizationSystem != nil {
		tribesCount := len(m.world.CivilizationSystem.Tribes)
		structuresCount := len(m.world.CivilizationSystem.Structures)
		if tribesCount > 0 || structuresCount > 0 {
			indicators = append(indicators, fmt.Sprintf("🏛️T%d/S%d", tribesCount, structuresCount))
		}
	}

	// Physics system indicator
	if m.world.PhysicsSystem != nil && m.world.PhysicsSystem.CollisionsThisTick > 0 {
		indicators = append(indicators, fmt.Sprintf("⚡%d", m.world.PhysicsSystem.CollisionsThisTick))
	}

	// Time of day indicator
	hour := m.world.Clock.Hour()
	var timeIcon string
	if hour >= 6 && hour < 18 {
		timeIcon = "☀️" // Day
	} else {
		timeIcon = "🌙" // Night
	}

	title := titleStyle.Render(fmt.Sprintf("🌍 Genetic Ecosystem - Tick %d", m.world.Tick))
	infoText := fmt.Sprintf("%s | %s %s | Entities: %d | Pops: %d | Events: %d | View: %s",
		status, timeIcon, worldTime, entities, populations, activeEvents, strings.ToUpper(m.selectedView))

	if len(indicators) > 0 {
		infoText += " | " + strings.Join(indicators, " ")
	}

	info := infoStyle.Render(infoText)

	return lipgloss.JoinHorizontal(lipgloss.Left, title, " ", info)
}

// gridView renders the animated world grid
func (m CLIModel) gridView() string {
	if m.world.Config.GridWidth == 0 || m.world.Config.GridHeight == 0 {
		return "Grid not initialized"
	}

	var gridBuilder strings.Builder
	// Build grid representation with viewport support
	startX := m.viewportX
	startY := m.viewportY
	displayWidth := min(m.world.Config.GridWidth-startX, 60)
	displayHeight := min(m.world.Config.GridHeight-startY, 25)

	for y := startY; y < startY+displayHeight; y++ {
		for x := startX; x < startX+displayWidth; x++ {
			cell := m.world.Grid[y][x]
			biome := m.world.Biomes[cell.Biome]
			symbol := biome.Symbol
			style := biomeColors[cell.Biome]

			// Check for structures first (highest priority)
			if m.showStructures && m.world.CivilizationSystem != nil {
				for _, structure := range m.world.CivilizationSystem.Structures {
					if int(structure.Position.X) == x && int(structure.Position.Y) == y && structure.IsActive {
						structureSymbols := map[StructureType]rune{
							StructureNest:    '🏠',
							StructureCache:   '📦',
							StructureBarrier: '🚧',
							StructureTrap:    '🕳',
							StructureFarm:    '🌾',
							StructureWell:    '🚰',
							StructureTower:   '🗼',
							StructureMarket:  '🏪',
						}
						if structSymbol, exists := structureSymbols[structure.Type]; exists {
							symbol = structSymbol
							style = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange for structures
						}
						break
					}
				}
			}

			// Check for signals (medium priority)
			if m.showSignals && m.world.CommunicationSystem != nil {
				for _, signal := range m.world.CommunicationSystem.Signals {
					distance := math.Sqrt((signal.Position.X-float64(x))*(signal.Position.X-float64(x)) +
						(signal.Position.Y-float64(y))*(signal.Position.Y-float64(y)))
					if distance <= signal.Range {
						// Show signal effect with different symbols
						signalSymbols := map[SignalType]rune{
							SignalDanger:    '!',
							SignalFood:      '*',
							SignalMating:    '♥',
							SignalTerritory: 'T',
							SignalHelp:      '?',
							SignalMigration: '→',
						}
						if signalSymbol, exists := signalSymbols[signal.Type]; exists && distance < 2.0 {
							symbol = signalSymbol
							style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Blink(true) // Red blinking
							break
						}
					}
				}
			}

			// If entities present, show the dominant species (override signals but not structures)
			if len(cell.Entities) > 0 {
				// Don't override structure symbols
				isStructure := false
				if m.showStructures && m.world.CivilizationSystem != nil {
					for _, structure := range m.world.CivilizationSystem.Structures {
						if int(structure.Position.X) == x && int(structure.Position.Y) == y && structure.IsActive {
							isStructure = true
							break
						}
					}
				}

				if !isStructure {
					speciesCount := make(map[string]int)
					for _, entity := range cell.Entities {
						speciesCount[entity.Species]++
					}

					// Find most common species
					maxCount := 0
					dominantSpecies := ""
					for species, count := range speciesCount {
						if count > maxCount {
							maxCount = count
							dominantSpecies = species
						}
					}

					if dominantSpecies != "" {
						// Get base species type from species naming system
						baseSpecies := dominantSpecies
						if m.world.SpeciesNaming != nil {
							if info := m.world.SpeciesNaming.GetSpeciesInfo(dominantSpecies); info != nil {
								baseSpecies = info.Species
							}
						}
						
						if sym, exists := m.speciesSymbols[baseSpecies]; exists {
							symbol = sym
						}
						if entityStyle, exists := speciesStyles[baseSpecies]; exists {
							style = entityStyle
						}
					}

					// Show multiple entities with numbers
					if len(cell.Entities) > 1 {
						if len(cell.Entities) < 10 {
							symbol = rune('0' + len(cell.Entities))
						} else {
							symbol = '+'
						}
					}
				}
			} else if len(cell.Plants) > 0 {
				// Show plants if no entities are present and no structures
				isStructureOrSignal := false
				if m.showStructures && m.world.CivilizationSystem != nil {
					for _, structure := range m.world.CivilizationSystem.Structures {
						if int(structure.Position.X) == x && int(structure.Position.Y) == y && structure.IsActive {
							isStructureOrSignal = true
							break
						}
					}
				}

				if !isStructureOrSignal {
					// Find the most prominent plant (largest or most numerous)
					var dominantPlant *Plant
					maxSize := 0.0
					for _, plant := range cell.Plants {
						if plant.IsAlive && plant.Size > maxSize {
							maxSize = plant.Size
							dominantPlant = plant
						}
					}

					if dominantPlant != nil {
						config := GetPlantConfigs()[dominantPlant.Type]
						symbol = config.Symbol
						// Use a dimmer style for plants
						style = biomeColors[cell.Biome].Copy().Foreground(lipgloss.Color("240"))
					}

					// Show multiple plants with small numbers
					if len(cell.Plants) > 1 {
						if len(cell.Plants) < 5 {
							symbol = rune('0' + len(cell.Plants))
						} else {
							symbol = '■'
						}
					}
				}
			}

			gridBuilder.WriteString(style.Render(string(symbol)))
		}
		if y < startY+displayHeight-1 {
			gridBuilder.WriteString("\n")
		}
	}

	grid := gridStyle.Render(gridBuilder.String())

	// Add legend
	legend := m.legendView()

	return lipgloss.JoinHorizontal(lipgloss.Top, grid, "  ", legend)
}

// legendView renders the legend for symbols and colors
func (m CLIModel) legendView() string {
	var legend strings.Builder

	legend.WriteString(titleStyle.Render("Legend") + "\n\n")

	legend.WriteString("🌱 Biomes:\n")
	// Get sorted biome types for consistent ordering
	var biomeTypes []BiomeType
	for biomeType := range m.world.Biomes {
		biomeTypes = append(biomeTypes, biomeType)
	}
	sort.Slice(biomeTypes, func(i, j int) bool {
		return m.world.Biomes[biomeTypes[i]].Name < m.world.Biomes[biomeTypes[j]].Name
	})
	
	for _, biomeType := range biomeTypes {
		biome := m.world.Biomes[biomeType]
		style := biomeColors[biomeType]
		legend.WriteString(fmt.Sprintf("%s %s\n",
			style.Render(string(biome.Symbol)), biome.Name))
	}

	legend.WriteString("\n👥 Species:\n")
	// Show actual species names from populations
	if m.world.SpeciesNaming != nil {
		for _, pop := range m.world.Populations {
			if info := m.world.SpeciesNaming.GetSpeciesInfo(pop.Species); info != nil {
				baseSpecies := info.Species
				if symbol, exists := m.speciesSymbols[baseSpecies]; exists {
					if style, exists := speciesStyles[baseSpecies]; exists {
						legend.WriteString(fmt.Sprintf("%s %s\n",
							style.Render(string(symbol)), pop.Species))
					}
				}
			}
		}
	} else {
		// Fallback to old behavior
		for species, symbol := range m.speciesSymbols {
			style := speciesStyles[species]
			legend.WriteString(fmt.Sprintf("%s %s\n",
				style.Render(string(symbol)), strings.Title(species)))
		}
	}

	legend.WriteString("\n📊 Numbers = Multiple entities\n")
	legend.WriteString("+ = 10+ entities")

	return legend.String()
}

// statsView renders detailed statistics
func (m CLIModel) statsView() string {
	stats := m.world.GetStats()

	var content strings.Builder
	content.WriteString(titleStyle.Render("World Statistics") + "\n\n")
	content.WriteString(fmt.Sprintf("Tick: %d\n", stats["tick"]))
	content.WriteString(fmt.Sprintf("Total Entities: %d\n", stats["total_entities"]))
	content.WriteString(fmt.Sprintf("Total Plants: %d\n", len(m.world.AllPlants)))
	content.WriteString(fmt.Sprintf("World Time: %s\n", m.world.Clock.Format("15:04 Day 2006-01-02")))
	content.WriteString("\n")

	// Plant statistics
	content.WriteString("Plant Distribution:\n")
	plantTypeCount := make(map[PlantType]int)
	alivePlants := 0
	for _, plant := range m.world.AllPlants {
		if plant.IsAlive {
			alivePlants++
			plantTypeCount[plant.Type]++
		}
	}

	content.WriteString(fmt.Sprintf("  Total Alive: %d\n", alivePlants))
	for plantType, count := range plantTypeCount {
		config := GetPlantConfigs()[plantType]
		content.WriteString(fmt.Sprintf("  %s: %d\n", config.Name, count))
	}
	content.WriteString("\n")

	// Population statistics
	if populations, ok := stats["populations"].(map[string]map[string]interface{}); ok {
		content.WriteString("Population Details:\n")
		for species, popStats := range populations {
			content.WriteString(fmt.Sprintf("\n%s:\n", strings.Title(species)))
			content.WriteString(fmt.Sprintf("  Count: %v\n", popStats["count"]))
			if avgEnergy, exists := popStats["avg_energy"]; exists {
				content.WriteString(fmt.Sprintf("  Avg Energy: %.1f\n", avgEnergy))
			}
			if avgAge, exists := popStats["avg_age"]; exists {
				content.WriteString(fmt.Sprintf("  Avg Age: %.1f\n", avgAge))
			}
		}
	}

	// Evolutionary Feedback Loop Statistics
	content.WriteString("\n\nEvolutionary Feedback Loops:\n")
	adaptationStats := m.calculateAdaptationStats()
	content.WriteString(fmt.Sprintf("  Entities with Dietary Memory: %d\n", adaptationStats["dietary_memory_count"]))
	content.WriteString(fmt.Sprintf("  Entities with Environmental Memory: %d\n", adaptationStats["env_memory_count"]))
	content.WriteString(fmt.Sprintf("  Avg Dietary Fitness: %.2f\n", adaptationStats["avg_dietary_fitness"]))
	content.WriteString(fmt.Sprintf("  Avg Environmental Fitness: %.2f\n", adaptationStats["avg_env_fitness"]))
	content.WriteString(fmt.Sprintf("  Active Plant Preferences: %d\n", adaptationStats["plant_preferences"]))
	content.WriteString(fmt.Sprintf("  Active Prey Preferences: %d\n", adaptationStats["prey_preferences"]))

	// Biome distribution
	content.WriteString("\n\nBiome Distribution:\n")
	biomeCount := make(map[BiomeType]int)
	for y := 0; y < m.world.Config.GridHeight; y++ {
		for x := 0; x < m.world.Config.GridWidth; x++ {
			biomeCount[m.world.Grid[y][x].Biome]++
		}
	}

	total := m.world.Config.GridWidth * m.world.Config.GridHeight
	
	// Get sorted biome types for consistent ordering  
	var sortedBiomes []BiomeType
	for biomeType := range biomeCount {
		sortedBiomes = append(sortedBiomes, biomeType)
	}
	sort.Slice(sortedBiomes, func(i, j int) bool {
		return m.world.Biomes[sortedBiomes[i]].Name < m.world.Biomes[sortedBiomes[j]].Name
	})
	
	for _, biomeType := range sortedBiomes {
		biome := m.world.Biomes[biomeType]
		count := biomeCount[biomeType]
		percentage := float64(count) * 100.0 / float64(total)
		content.WriteString(fmt.Sprintf("  %s: %d cells (%.1f%%)\n",
			biome.Name, count, percentage))
	}

	return content.String()
}

// calculateAdaptationStats computes feedback loop adaptation statistics
func (m CLIModel) calculateAdaptationStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	dietaryMemoryCount := 0
	envMemoryCount := 0
	totalDietaryFitness := 0.0
	totalEnvFitness := 0.0
	plantPreferences := 0
	preyPreferences := 0
	
	entityCount := 0
	
	// Collect data from all entities
	for _, population := range m.world.Populations {
		for _, entity := range population.Entities {
			if !entity.IsAlive {
				continue
			}
			entityCount++
			
			// Check dietary memory
			if entity.DietaryMemory != nil {
				dietaryMemoryCount++
				totalDietaryFitness += entity.DietaryMemory.DietaryFitness
				if entity.DietaryMemory.PlantTypePreferences != nil {
					plantPreferences += len(entity.DietaryMemory.PlantTypePreferences)
				}
				if entity.DietaryMemory.PreySpeciesPreferences != nil {
					preyPreferences += len(entity.DietaryMemory.PreySpeciesPreferences)
				}
			}
			
			// Check environmental memory
			if entity.EnvironmentalMemory != nil {
				envMemoryCount++
				totalEnvFitness += entity.EnvironmentalMemory.AdaptationFitness
			}
		}
	}
	
	stats["dietary_memory_count"] = dietaryMemoryCount
	stats["env_memory_count"] = envMemoryCount
	stats["plant_preferences"] = plantPreferences
	stats["prey_preferences"] = preyPreferences
	
	if dietaryMemoryCount > 0 {
		stats["avg_dietary_fitness"] = totalDietaryFitness / float64(dietaryMemoryCount)
	} else {
		stats["avg_dietary_fitness"] = 0.0
	}
	
	if envMemoryCount > 0 {
		stats["avg_env_fitness"] = totalEnvFitness / float64(envMemoryCount)
	} else {
		stats["avg_env_fitness"] = 0.0
	}
	
	return stats
}

// eventsView renders active world events and recent event log
func (m CLIModel) eventsView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("World Events & Event Log") + "\n\n")

	// Active Events Section
	content.WriteString("=== ACTIVE EVENTS ===\n")
	if len(m.world.Events) == 0 {
		content.WriteString("No active events\n")
	} else {
		for i, event := range m.world.Events {
			content.WriteString(eventStyle.Render(fmt.Sprintf("🌪 %s", event.Name)) + "\n")
			content.WriteString(fmt.Sprintf("   %s\n", event.Description))
			content.WriteString(fmt.Sprintf("   Duration: %d ticks remaining\n", event.Duration))

			if event.GlobalMutation > 0 {
				content.WriteString(fmt.Sprintf("   Mutation Rate: +%.1f%%\n", event.GlobalMutation*100))
			}

			if event.GlobalDamage > 0 {
				content.WriteString(fmt.Sprintf("   Damage: %.1f energy/tick\n", event.GlobalDamage))
			}

			if i < len(m.world.Events)-1 {
				content.WriteString("\n")
			}
		}
	}

	// Recent Event Log Section
	content.WriteString("\n\n=== RECENT EVENT LOG ===\n")
	if m.world.EventLogger != nil {
		recentEvents := m.world.EventLogger.GetRecentEvents(10) // Get last 10 events
		if len(recentEvents) == 0 {
			content.WriteString("No events logged yet\n")
		} else {
			for i, event := range recentEvents {
				// Format timestamp relative to current tick
				ticksAgo := m.world.Tick - event.Tick
				timeStr := fmt.Sprintf("T-%d", ticksAgo)
				if ticksAgo == 0 {
					timeStr = "NOW"
				}

				content.WriteString(fmt.Sprintf("[%s] %s: %s\n",
					timeStr, event.Type, event.Description))

				if i >= 9 { // Limit display to prevent overflow
					break
				}
			}
		}
	} else {
		content.WriteString("Event logger not initialized\n")
	}

	content.WriteString("\n=== POSSIBLE EVENTS ===")
	content.WriteString("\n• Solar Flare - Increases radiation and mutations")
	content.WriteString("\n• Meteor Shower - Creates radiation zones")
	content.WriteString("\n• Ice Age - Increases energy drain worldwide")
	content.WriteString("\n• Volcanic Winter - Ash clouds cause damage and mutations")

	return content.String()
}

// populationsView renders population details
func (m CLIModel) populationsView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Population Details") + "\n\n")

	// Get sorted population names for consistent ordering
	var popNames []string
	for species := range m.world.Populations {
		popNames = append(popNames, species)
	}
	sort.Strings(popNames)

	for _, species := range popNames {
		pop := m.world.Populations[species]
		content.WriteString(fmt.Sprintf("=== %s ===\n", strings.ToUpper(species)))
		content.WriteString(fmt.Sprintf("Population Size: %d\n", len(pop.Entities)))
		content.WriteString(fmt.Sprintf("Species: %s\n", pop.Species))
		content.WriteString(fmt.Sprintf("Generation: %d\n", pop.Generation))
		content.WriteString(fmt.Sprintf("Mutation Rate: %.3f\n", pop.MutationRate))

		if len(pop.Entities) > 0 {
			// Calculate averages
			totalEnergy := 0.0
			totalAge := 0
			totalFitness := 0.0
			aliveCount := 0

			traitSums := make(map[string]float64)

			for _, entity := range pop.Entities {
				if entity.IsAlive {
					aliveCount++
					totalEnergy += entity.Energy
					totalAge += entity.Age
					totalFitness += entity.Fitness

					for name, trait := range entity.Traits {
						traitSums[name] += trait.Value
					}
				}
			}

			if aliveCount > 0 {
				content.WriteString(fmt.Sprintf("Alive: %d\n", aliveCount))
				content.WriteString(fmt.Sprintf("Avg Energy: %.1f\n", totalEnergy/float64(aliveCount)))
				content.WriteString(fmt.Sprintf("Avg Age: %.1f\n", float64(totalAge)/float64(aliveCount)))
				content.WriteString(fmt.Sprintf("Avg Fitness: %.3f\n", totalFitness/float64(aliveCount)))

				content.WriteString("\nAverage Traits:\n")
				for trait, sum := range traitSums {
					avg := sum / float64(aliveCount)
					content.WriteString(fmt.Sprintf("  %s: %.3f\n", trait, avg))
				}
				
				// Add feedback loop adaptation information
				dietaryMemoryCount := 0
				envMemoryCount := 0
				totalDietaryFitness := 0.0
				totalEnvFitness := 0.0
				plantPrefs := 0
				preyPrefs := 0
				
				for _, entity := range pop.Entities {
					if !entity.IsAlive {
						continue
					}
					
					if entity.DietaryMemory != nil {
						dietaryMemoryCount++
						totalDietaryFitness += entity.DietaryMemory.DietaryFitness
						if entity.DietaryMemory.PlantTypePreferences != nil {
							plantPrefs += len(entity.DietaryMemory.PlantTypePreferences)
						}
						if entity.DietaryMemory.PreySpeciesPreferences != nil {
							preyPrefs += len(entity.DietaryMemory.PreySpeciesPreferences)
						}
					}
					
					if entity.EnvironmentalMemory != nil {
						envMemoryCount++
						totalEnvFitness += entity.EnvironmentalMemory.AdaptationFitness
					}
				}
				
				content.WriteString("\nEvolutionary Adaptations:\n")
				content.WriteString(fmt.Sprintf("  Dietary adaptations: %d/%d (%.1f%%)\n", 
					dietaryMemoryCount, aliveCount, float64(dietaryMemoryCount)*100/float64(aliveCount)))
				content.WriteString(fmt.Sprintf("  Environmental adaptations: %d/%d (%.1f%%)\n", 
					envMemoryCount, aliveCount, float64(envMemoryCount)*100/float64(aliveCount)))
				
				if dietaryMemoryCount > 0 {
					content.WriteString(fmt.Sprintf("  Avg dietary fitness: %.3f\n", totalDietaryFitness/float64(dietaryMemoryCount)))
					content.WriteString(fmt.Sprintf("  Plant preferences: %d\n", plantPrefs))
					content.WriteString(fmt.Sprintf("  Prey preferences: %d\n", preyPrefs))
				}
				
				if envMemoryCount > 0 {
					content.WriteString(fmt.Sprintf("  Avg environmental fitness: %.3f\n", totalEnvFitness/float64(envMemoryCount)))
				}
			}
		}

		content.WriteString("\n")
	}

	return content.String()
}

// communicationView renders active signals and entity communication
func (m CLIModel) communicationView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Communication System") + "\n\n")

	if m.world.CommunicationSystem == nil {
		content.WriteString("Communication system not initialized\n")
		return content.String()
	}

	// Active Signals Section
	content.WriteString("=== ACTIVE SIGNALS ===\n")
	if len(m.world.CommunicationSystem.Signals) == 0 {
		content.WriteString("No active signals\n")
	} else {
		signalTypes := map[SignalType]string{
			SignalDanger:    "🚨 DANGER",
			SignalFood:      "🍎 FOOD",
			SignalMating:    "💕 MATING",
			SignalTerritory: "🏴 TERRITORY",
			SignalHelp:      "🆘 HELP",
			SignalMigration: "🧭 MIGRATION",
		}

		signalCounts := make(map[SignalType]int)
		for _, signal := range m.world.CommunicationSystem.Signals {
			signalCounts[signal.Type]++
		}

		for signalType, count := range signalCounts {
			typeName := signalTypes[signalType]
			content.WriteString(fmt.Sprintf("%s: %d active\n", typeName, count))
		}

		content.WriteString("\nRecent Signals:\n")
		// Show the 10 most recent signals
		recentCount := 0
		for i := len(m.world.CommunicationSystem.Signals) - 1; i >= 0 && recentCount < 10; i-- {
			signal := m.world.CommunicationSystem.Signals[i]
			typeName := signalTypes[signal.Type]
			age := m.world.Tick - signal.Timestamp
			content.WriteString(fmt.Sprintf("  %s at (%.0f,%.0f) - %d ticks ago - Strength: %.1f\n",
				typeName, signal.Position.X, signal.Position.Y, age, signal.Strength))
			recentCount++
		}
	}

	// Communication Activity Statistics
	content.WriteString("\n=== COMMUNICATION STATS ===\n")
	intelligentEntities := 0
	cooperativeEntities := 0
	totalEntities := len(m.world.AllEntities)

	for _, entity := range m.world.AllEntities {
		if entity.IsAlive {
			if entity.GetTrait("intelligence") > 0.3 {
				intelligentEntities++
			}
			if entity.GetTrait("cooperation") > 0.2 {
				cooperativeEntities++
			}
		}
	}

	content.WriteString(fmt.Sprintf("Entities capable of communication: %d/%d\n",
		intelligentEntities, totalEntities))
	content.WriteString(fmt.Sprintf("Cooperative entities: %d/%d\n",
		cooperativeEntities, totalEntities))
	content.WriteString(fmt.Sprintf("Signal efficiency: %.1f%%\n",
		float64(len(m.world.CommunicationSystem.Signals))/float64(m.world.CommunicationSystem.MaxSignals)*100))

	return content.String()
}

// civilizationView renders tribal information and structures
func (m CLIModel) civilizationView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Civilization System") + "\n\n")

	if m.world.CivilizationSystem == nil {
		content.WriteString("Civilization system not initialized\n")
		return content.String()
	}

	// Tribes Section
	content.WriteString("=== ACTIVE TRIBES ===\n")
	if len(m.world.CivilizationSystem.Tribes) == 0 {
		content.WriteString("No tribes formed yet\n")
	} else {
		for _, tribe := range m.world.CivilizationSystem.Tribes {
			// Determine dominant species in tribe
			speciesCount := make(map[string]int)
			for _, member := range tribe.Members {
				speciesCount[member.Species]++
			}
			dominantSpecies := "mixed"
			maxCount := 0
			for species, count := range speciesCount {
				if count > maxCount {
					maxCount = count
					dominantSpecies = species
				}
			}

			content.WriteString(fmt.Sprintf("🏴 Tribe %d (%s - %s)\n", tribe.ID, tribe.Name, dominantSpecies))
			content.WriteString(fmt.Sprintf("  Members: %d\n", len(tribe.Members)))

			// Calculate territory size from positions
			territorySize := float64(len(tribe.Territory))
			content.WriteString(fmt.Sprintf("  Territory: %.0f locations\n", territorySize))

			// Calculate cohesion from culture
			cohesion := tribe.Culture["cooperation"]
			content.WriteString(fmt.Sprintf("  Cohesion: %.2f\n", cohesion))

			if tribe.Leader != nil {
				content.WriteString(fmt.Sprintf("  Leader: Entity %d (Fitness: %.2f)\n",
					tribe.Leader.ID, tribe.Leader.Fitness))
			}

			// Show tribe's structures
			structureCount := 0
			for _, structure := range m.world.CivilizationSystem.Structures {
				if structure.Tribe == tribe {
					structureCount++
				}
			}
			content.WriteString(fmt.Sprintf("  Structures: %d\n", structureCount))
			content.WriteString("\n")
		}
	}

	// Structures Section
	content.WriteString("=== STRUCTURES ===\n")
	if len(m.world.CivilizationSystem.Structures) == 0 {
		content.WriteString("No structures built yet\n")
	} else {
		structureTypes := map[StructureType]string{
			StructureNest:    "🏠 Nest",
			StructureCache:   "📦 Cache",
			StructureBarrier: "🚧 Barrier",
			StructureTrap:    "🕳 Trap",
			StructureFarm:    "🌾 Farm",
			StructureWell:    "🚰 Well",
			StructureTower:   "🗼 Tower",
			StructureMarket:  "🏪 Market",
		}

		structureCounts := make(map[StructureType]int)
		activeStructures := 0
		for _, structure := range m.world.CivilizationSystem.Structures {
			structureCounts[structure.Type]++
			if structure.IsActive {
				activeStructures++
			}
		}

		content.WriteString(fmt.Sprintf("Total: %d (%d active)\n",
			len(m.world.CivilizationSystem.Structures), activeStructures))

		for structureType, count := range structureCounts {
			typeName := structureTypes[structureType]
			content.WriteString(fmt.Sprintf("  %s: %d\n", typeName, count))
		}

		// Show recent structures
		content.WriteString("\nRecent Structures:\n")
		recentCount := 0
		for i := len(m.world.CivilizationSystem.Structures) - 1; i >= 0 && recentCount < 5; i-- {
			structure := m.world.CivilizationSystem.Structures[i]
			typeName := structureTypes[structure.Type]
			age := m.world.Tick - structure.CreationTick
			status := "Active"
			if !structure.IsActive {
				status = "Inactive"
			}
			content.WriteString(fmt.Sprintf("  %s at (%.0f,%.0f) - %d ticks old - %s\n",
				typeName, structure.Position.X, structure.Position.Y, age, status))
			recentCount++
		}
	}

	// Civilization Development Index
	content.WriteString("\n=== CIVILIZATION INDEX ===\n")
	totalStructures := len(m.world.CivilizationSystem.Structures)
	totalTribes := len(m.world.CivilizationSystem.Tribes)

	developmentIndex := float64(totalStructures*2+totalTribes*5) / float64(len(m.world.AllEntities)+1)
	content.WriteString(fmt.Sprintf("Development Index: %.2f\n", developmentIndex))

	if developmentIndex < 0.1 {
		content.WriteString("Civilization Level: Primitive\n")
	} else if developmentIndex < 0.5 {
		content.WriteString("Civilization Level: Developing\n")
	} else if developmentIndex < 1.0 {
		content.WriteString("Civilization Level: Advanced\n")
	} else {
		content.WriteString("Civilization Level: Highly Advanced\n")
	}

	return content.String()
}

// physicsView renders physics and movement information
func (m CLIModel) physicsView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Physics System") + "\n\n")

	if m.world.PhysicsSystem == nil {
		content.WriteString("Physics system not initialized\n")
		return content.String()
	}

	// Movement Statistics
	content.WriteString("=== MOVEMENT STATISTICS ===\n")
	movingEntities := 0
	totalVelocity := 0.0
	maxVelocity := 0.0
	for _, entity := range m.world.AllEntities {
		if entity.IsAlive {
			physics := m.world.PhysicsComponents[entity.ID]
			var speed float64
			if physics != nil {
				speed = math.Sqrt(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y)
			}
			if speed > 0.1 {
				movingEntities++
				totalVelocity += speed
				if speed > maxVelocity {
					maxVelocity = speed
				}
			}
		}
	}

	avgVelocity := 0.0
	if movingEntities > 0 {
		avgVelocity = totalVelocity / float64(movingEntities)
	}

	content.WriteString(fmt.Sprintf("Moving entities: %d/%d\n", movingEntities, len(m.world.AllEntities)))
	content.WriteString(fmt.Sprintf("Average velocity: %.2f\n", avgVelocity))
	content.WriteString(fmt.Sprintf("Maximum velocity: %.2f\n", maxVelocity))

	// Collision Statistics
	content.WriteString("\n=== COLLISION STATISTICS ===\n")
	content.WriteString(fmt.Sprintf("Collisions this tick: %d\n", m.world.PhysicsSystem.CollisionsThisTick))
	content.WriteString(fmt.Sprintf("Total collisions: %d\n", m.world.PhysicsSystem.TotalCollisions))

	avgCollisions := 0.0
	if m.world.Tick > 0 {
		avgCollisions = float64(m.world.PhysicsSystem.TotalCollisions) / float64(m.world.Tick)
	}
	content.WriteString(fmt.Sprintf("Average collisions/tick: %.2f\n", avgCollisions))

	// Force Analysis
	content.WriteString("\n=== FORCE ANALYSIS ===\n")
	content.WriteString("Active Forces:\n")

	// Count entities affected by different forces
	gravityAffected := 0
	frictionAffected := 0
	for _, entity := range m.world.AllEntities {
		if entity.IsAlive {
			// Entities are affected by friction if they're moving
			physics := m.world.PhysicsComponents[entity.ID]
			var speed float64
			if physics != nil {
				speed = math.Sqrt(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y)
			}
			if speed > 0.1 {
				frictionAffected++
			}

			// All entities are affected by environmental forces
			gravityAffected++
		}
	}

	content.WriteString(fmt.Sprintf("  Environmental forces: %d entities\n", gravityAffected))
	content.WriteString(fmt.Sprintf("  Friction: %d entities\n", frictionAffected))

	// Entity Distribution by Speed
	content.WriteString("\n=== SPEED DISTRIBUTION ===\n")
	speedBands := []struct {
		min, max float64
		name     string
	}{
		{0.0, 0.1, "Stationary"},
		{0.1, 0.5, "Slow"},
		{0.5, 1.0, "Medium"},
		{1.0, 2.0, "Fast"},
		{2.0, 999.0, "Very Fast"},
	}
	for _, band := range speedBands {
		count := 0
		for _, entity := range m.world.AllEntities {
			if entity.IsAlive {
				physics := m.world.PhysicsComponents[entity.ID]
				var speed float64
				if physics != nil {
					speed = math.Sqrt(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y)
				}
				if speed >= band.min && speed < band.max {
					count++
				}
			}
		}
		if count > 0 {
			content.WriteString(fmt.Sprintf("  %s (%.1f-%.1f): %d entities\n",
				band.name, band.min, band.max, count))
		}
	}

	return content.String()
}

// footerView renders the footer with controls
func (m CLIModel) footerView() string {
	controls := []string{
		"space: pause/resume",
		"v: cycle view",
		"arrows: navigate",
		"enter: step",
		"s/t/p: toggles",
		"r: reset",
		"?: help",
		"q: quit",
	}

	return infoStyle.Render(strings.Join(controls, " | "))
}

// helpView renders the help screen
func (m CLIModel) helpView() string {
	help := `
🌍 Genetic Ecosystem Simulation Help

CONTROLS:
  space      Pause/Resume simulation
  enter      Manual step (when paused)
  v          Cycle through views (grid/stats/events/populations/communication/civilization/physics/wind)
  a          Toggle auto-advance
  ←→↑↓/hjkl  Navigate viewport (pan around world)
  z          Cycle zoom level
  r          Reset viewport to origin
  s          Toggle signal visualization
  t          Toggle structure visualization
  p          Toggle physics visualization
  ?          Toggle this help screen
  q          Quit

VIEWS:
  Grid         Real-time animated world map with entities and biomes
  Stats        Detailed world and population statistics
  Events       Active world events and their effects
  Populations  Detailed view of each species population
  Communication Active signals and entity communication data
  Civilization Tribal information and structure development
  Physics      Movement statistics and collision data
  Wind         Wind patterns and pollen dispersal information
  Species      Species tracking and visualization
  Network      Plant network statistics and underground connections

HEADER INDICATORS:
  📡N         Active communication signals (N = count)
  🏛️T/S       Tribes (T) and Structures (S) counts
  ⚡N         Collisions this tick (N = count)
  ☀️/🌙       Day/Night time indicator

GRID SYMBOLS:
  .          Plains biome
  ♠          Forest biome  
  ~          Desert biome
  ^          Mountain biome
  ≈          Water biome
  ☢          Radiation biome
  
  ●          Herbivore entity
  ▲          Predator entity
  ◆          Omnivore entity
  2-9        Multiple entities in cell
  +          10+ entities in cell

STRUCTURES (when structure visualization enabled):
  🏠         Nest (shelter)
  📦         Cache (food storage)
  🚧         Barrier (defensive wall)
  🕳         Trap (hunting trap)
  🌾         Farm (cultivated plants)
  🚰         Well (water source)
  🗼         Tower (observation post)
  🏪         Market (trading post)

PLANTS (shown when no entities present):
  .          Grass (nutrition: 15)
  ♦          Bush (nutrition: 25, slightly toxic)
  ♠          Tree (nutrition: 40, large)
  ♪          Mushroom (nutrition: 20, toxic)
  ~          Algae (nutrition: 10, aquatic)
  †          Cactus (nutrition: 30, toxic, desert-adapted)
  1-4        Multiple plants in cell
  ■          5+ plants in cell

ADVANCED SYSTEMS:
The simulation features multiple interconnected systems:

• Communication: Entities can send signals (danger, food, mating, etc.) 
  to coordinate behavior. Intelligence and cooperation traits affect ability.

• Civilization: Intelligent, cooperative entities can form tribes and build 
  structures that provide benefits like shelter, food storage, and defense.

• Physics: Realistic movement with velocity, collision detection, and 
  environmental forces affecting entity behavior and interaction.

• Time Cycles: Day/night cycles affect entity behavior, with some entities
  being more active during certain times.

• Plant Ecosystem: Six plant types with different nutritional values, 
  toxicity levels, and biome preferences form the food web base.

• Event System: Random world events like solar flares, meteor showers,
  and ice ages create evolutionary pressure and environmental challenges.

• Wind System: Wind patterns affect pollen dispersal, with seasonal changes
  influencing reproduction strategies of plants.

All systems interact dynamically - communication helps coordinate responses 
to events, civilization provides resilience against environmental challenges,
physics creates realistic movement and interaction patterns, and wind patterns
influence plant reproduction and migration.

Press ? again to return to the simulation.
`
	return help
}

// windView renders wind patterns and pollen dispersal information
func (m CLIModel) windView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Wind & Pollen System") + "\n\n")

	if m.world.WindSystem == nil {
		content.WriteString("Wind system not initialized\n")
		return content.String()
	}

	// Current Wind Conditions
	content.WriteString("=== CURRENT WIND CONDITIONS ===\n")
	windStats := m.world.WindSystem.GetWindStats()

	// Wind strength and direction
	baseDirection := windStats["base_wind_direction"].(float64)
	baseStrength := windStats["base_wind_strength"].(float64)
	seasonalMultiplier := windStats["seasonal_multiplier"].(float64)
	effectiveStrength := baseStrength * seasonalMultiplier

	// Convert direction to compass bearing
	directionDegrees := baseDirection * 180.0 / 3.14159
	if directionDegrees < 0 {
		directionDegrees += 360
	}

	var compassDirection string
	switch {
	case directionDegrees < 22.5 || directionDegrees >= 337.5:
		compassDirection = "N"
	case directionDegrees < 67.5:
		compassDirection = "NE"
	case directionDegrees < 112.5:
		compassDirection = "E"
	case directionDegrees < 157.5:
		compassDirection = "SE"
	case directionDegrees < 202.5:
		compassDirection = "S"
	case directionDegrees < 247.5:
		compassDirection = "SW"
	case directionDegrees < 292.5:
		compassDirection = "W"
	default:
		compassDirection = "NW"
	}

	content.WriteString(fmt.Sprintf("Direction: %s (%.0f°)\n", compassDirection, directionDegrees))
	content.WriteString(fmt.Sprintf("Base strength: %.2f\n", baseStrength))
	content.WriteString(fmt.Sprintf("Seasonal modifier: %.2fx\n", seasonalMultiplier))
	content.WriteString(fmt.Sprintf("Effective strength: %.2f\n", effectiveStrength))
	// Weather conditions
	weatherDuration := windStats["weather_duration"].(int)
	weatherDescription := m.world.WindSystem.GetWeatherDescription()

	content.WriteString(fmt.Sprintf("Weather: %s (%d ticks remaining)\n", weatherDescription, weatherDuration))

	// Wind strength visualization
	strengthBars := int(effectiveStrength * 20)
	if strengthBars > 20 {
		strengthBars = 20
	}
	windBar := strings.Repeat("█", strengthBars) + strings.Repeat("░", 20-strengthBars)
	content.WriteString(fmt.Sprintf("Strength: [%s] %.1f%%\n", windBar, effectiveStrength*100))

	// Pollen Activity
	content.WriteString("\n=== POLLEN ACTIVITY ===\n")
	activePollenGrains := windStats["active_pollen_grains"].(int)
	totalPollenReleased := windStats["total_pollen_released"].(int)
	pollinationsThisTick := windStats["pollinations_this_tick"].(int)
	totalCrossPollinations := windStats["total_cross_pollinations"].(int)

	content.WriteString(fmt.Sprintf("Active pollen grains: %d\n", activePollenGrains))
	content.WriteString(fmt.Sprintf("Total pollen released: %d\n", totalPollenReleased))
	content.WriteString(fmt.Sprintf("Cross-pollinations this tick: %d\n", pollinationsThisTick))
	content.WriteString(fmt.Sprintf("Total cross-pollinations: %d\n", totalCrossPollinations))

	// Calculate pollen success rate
	successRate := 0.0
	if totalPollenReleased > 0 {
		successRate = float64(totalCrossPollinations) / float64(totalPollenReleased) * 100
	}
	content.WriteString(fmt.Sprintf("Pollination success rate: %.2f%%\n", successRate))

	// Insect Pollination Activity  
	content.WriteString("\n=== INSECT POLLINATION ===\n")
	pollinationStats := m.world.InsectPollinationSystem.GetPollinationStats()
	
	content.WriteString(fmt.Sprintf("Active flower patches: %d\n", pollinationStats["active_flower_patches"].(int)))
	content.WriteString(fmt.Sprintf("Active pollinators: %d\n", pollinationStats["active_pollinators"].(int)))
	content.WriteString(fmt.Sprintf("Total insect pollinations: %d\n", pollinationStats["total_pollinations"].(int)))
	content.WriteString(fmt.Sprintf("Cross-species pollinations: %d\n", pollinationStats["cross_species_pollinations"].(int)))
	content.WriteString(fmt.Sprintf("Cross-species rate: %.2f%%\n", pollinationStats["cross_species_rate"].(float64)*100))
	content.WriteString(fmt.Sprintf("Nectar produced: %.1f\n", pollinationStats["nectar_produced"].(float64)))
	content.WriteString(fmt.Sprintf("Nectar consumed: %.1f\n", pollinationStats["nectar_consumed"].(float64)))
	content.WriteString(fmt.Sprintf("Recent pollination events: %d\n", pollinationStats["recent_pollination_events"].(int)))
	content.WriteString(fmt.Sprintf("Seasonal modifier: %.2f\n", pollinationStats["seasonal_modifier"].(float64)))

	// Plant Reproduction Analysis
	content.WriteString("\n=== PLANT REPRODUCTION ===\n")

	// Count plants by type and reproduction status
	plantCounts := make(map[PlantType]int)
	reproducingPlants := make(map[PlantType]int)
	totalPlants := 0
	totalReproducing := 0

	for _, plant := range m.world.AllPlants {
		if plant.IsAlive {
			plantCounts[plant.Type]++
			totalPlants++
			if plant.CanReproduce() {
				reproducingPlants[plant.Type]++
				totalReproducing++
			}
		}
	}

	content.WriteString(fmt.Sprintf("Total plants: %d (%d capable of reproduction)\n", totalPlants, totalReproducing))

	// Show reproduction rates by plant type
	content.WriteString("\nReproduction by type:\n")
	plantConfigs := GetPlantConfigs()
	for plantType, config := range plantConfigs {
		if count, exists := plantCounts[plantType]; exists && count > 0 {
			reproducingCount := reproducingPlants[plantType]
			reproductionRate := float64(reproducingCount) / float64(count) * 100
			content.WriteString(fmt.Sprintf("  %s %s: %d total, %d reproducing (%.1f%%)\n",
				string(config.Symbol), config.Name, count, reproducingCount, reproductionRate))
		}
	}

	// Wind Map Visualization (simplified)
	content.WriteString("\n=== WIND MAP SAMPLE ===\n")
	content.WriteString("Wind vectors across world (sample 5x5 grid):\n")

	// Sample wind vectors from different parts of the world
	mapWidth := m.world.WindSystem.MapWidth
	mapHeight := m.world.WindSystem.MapHeight

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			// Sample from different parts of the wind map
			sampleX := (x * mapWidth) / 5
			sampleY := (y * mapHeight) / 5

			if sampleX < mapWidth && sampleY < mapHeight {
				windVector := m.world.WindSystem.WindMap[sampleY][sampleX]

				// Convert wind vector to arrow direction
				var arrow string
				if windVector.Strength < 0.1 {
					arrow = "·" // Calm
				} else {
					// Determine arrow direction from wind vector
					angle := math.Atan2(windVector.Y, windVector.X)
					angleDegrees := angle * 180.0 / math.Pi
					if angleDegrees < 0 {
						angleDegrees += 360
					}

					switch {
					case angleDegrees < 22.5 || angleDegrees >= 337.5:
						arrow = "→"
					case angleDegrees < 67.5:
						arrow = "↘"
					case angleDegrees < 112.5:
						arrow = "↓"
					case angleDegrees < 157.5:
						arrow = "↙"
					case angleDegrees < 202.5:
						arrow = "←"
					case angleDegrees < 247.5:
						arrow = "↖"
					case angleDegrees < 292.5:
						arrow = "↑"
					default:
						arrow = "↗"
					}
				}
				content.WriteString(arrow)
			} else {
				content.WriteString(" ")
			}
		}
		content.WriteString("\n")
	}

	// Seasonal Effects
	currentSeason := m.world.AdvancedTimeSystem.GetTimeState().Season
	content.WriteString("\n=== SEASONAL EFFECTS ===\n")
	content.WriteString(fmt.Sprintf("Current season: %s\n", seasonNames[currentSeason]))

	switch currentSeason {
	case Spring:
		content.WriteString("• Enhanced pollen dispersal (+20% wind strength)\n")
		content.WriteString("• Peak flowering season\n")
		content.WriteString("• Increased plant reproduction rates\n")
	case Summer:
		content.WriteString("• Calmer winds (-20% wind strength)\n")
		content.WriteString("• Reduced pollen dispersal\n")
		content.WriteString("• Focus on growth over reproduction\n")
	case Autumn:
		content.WriteString("• Strong winds (+40% wind strength)\n")
		content.WriteString("• Seed dispersal season\n")
		content.WriteString("• Final reproduction push\n")
	case Winter:
		content.WriteString("• Harsh winds (+60% wind strength)\n")
		content.WriteString("• Minimal plant reproduction\n")
		content.WriteString("• Survival mode for plants\n")
	}

	return content.String()
}

// speciesView renders plant species evolution and tracking information
func (m CLIModel) speciesView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Species Evolution & Tracking") + "\n\n")

	if m.world.SpeciationSystem == nil {
		content.WriteString("Speciation system not initialized\n")
		return content.String()
	}

	// System overview
	stats := m.world.SpeciationSystem.GetSpeciesStats()
	content.WriteString("=== SYSTEM OVERVIEW ===\n")
	content.WriteString(fmt.Sprintf("Active species: %v\n", stats["active_species"]))
	content.WriteString(fmt.Sprintf("Total species formed: %v\n", stats["total_species_formed"]))
	content.WriteString(fmt.Sprintf("Species extinct: %v\n", stats["total_species_extinct"]))
	content.WriteString(fmt.Sprintf("Max concurrent species: %v (tick %v)\n",
		stats["max_active_species"], stats["max_active_species_tick"]))
	content.WriteString(fmt.Sprintf("Genetic distance threshold: %.2f\n",
		stats["genetic_distance_threshold"]))

	if largestSpeciesName, ok := stats["largest_species_name"].(string); ok && largestSpeciesName != "" {
		content.WriteString(fmt.Sprintf("Largest species: %s (%v individuals)\n",
			largestSpeciesName, stats["largest_species_size"]))
	}

	// Active species list
	content.WriteString("\n=== ACTIVE SPECIES ===\n")
	speciesList := m.world.SpeciationSystem.GetActiveSpeciesList()

	if len(speciesList) == 0 {
		content.WriteString("No active species yet\n")
	} else {
		content.WriteString("Name               Type      Pop  Peak  Formation  Parent\n")
		content.WriteString("─────────────────────────────────────────────────────────\n")

		for _, species := range speciesList {
			name := species["name"].(string)
			if len(name) > 18 {
				name = name[:15] + "..."
			}

			originType := species["origin_type"].(PlantType)
			typeStr := GetPlantConfigs()[originType].Name
			if len(typeStr) > 8 {
				typeStr = typeStr[:8]
			}

			pop := species["current_population"].(int)
			peak := species["peak_population"].(int)
			formationTick := species["formation_tick"].(int)
			parentID := species["parent_species_id"].(int)

			parentStr := "root"
			if parentID > 0 {
				parentStr = fmt.Sprintf("S%d", parentID)
			}

			content.WriteString(fmt.Sprintf("%-18s %-8s %4d %4d %8d  %s\n",
				name, typeStr, pop, peak, formationTick, parentStr))
		}
	}

	// Recent events
	content.WriteString("\n=== RECENT EVOLUTION EVENTS ===\n")
	recentEvents := m.world.SpeciationSystem.GetRecentEvents(5)

	speciationEvents := recentEvents["speciation_events"].([]SpeciationEvent)
	extinctionEvents := recentEvents["extinction_events"].([]ExtinctionEvent)

	if len(speciationEvents) == 0 && len(extinctionEvents) == 0 {
		content.WriteString("No evolution events yet\n")
	} else {
		// Show recent speciation events
		if len(speciationEvents) > 0 {
			content.WriteString("🌱 Recent Speciation:\n")
			for i := len(speciationEvents) - 1; i >= 0 && i >= len(speciationEvents)-3; i-- {
				event := speciationEvents[i]
				content.WriteString(fmt.Sprintf("  Tick %d: S%d split from S%d (distance: %.3f, %d members)\n",
					event.Tick, event.NewSpeciesID, event.ParentSpeciesID,
					event.GeneticDistance, event.MemberCount))
			}
		}

		// Show recent extinction events
		if len(extinctionEvents) > 0 {
			content.WriteString("💀 Recent Extinctions:\n")
			for i := len(extinctionEvents) - 1; i >= 0 && i >= len(extinctionEvents)-3; i-- {
				event := extinctionEvents[i]
				content.WriteString(fmt.Sprintf("  Tick %d: %s (lifespan: %d ticks)\n",
					event.Tick, event.SpeciesName, event.Lifespan))
			}
		}
	}

	// Genetic diversity analysis
	content.WriteString("\n=== GENETIC DIVERSITY ===\n")

	// Count total plants in species vs unassigned
	totalPlantsInSpecies := 0
	for _, species := range speciesList {
		totalPlantsInSpecies += species["current_population"].(int)
	}

	totalPlants := 0
	for _, plant := range m.world.AllPlants {
		if plant.IsAlive {
			totalPlants++
		}
	}

	if totalPlants > 0 {
		speciesPercentage := float64(totalPlantsInSpecies) / float64(totalPlants) * 100
		content.WriteString(fmt.Sprintf("Plants in species: %d/%d (%.1f%%)\n",
			totalPlantsInSpecies, totalPlants, speciesPercentage))
	}

	// Species diversity index (simplified Shannon diversity)
	if len(speciesList) > 1 {
		diversity := 0.0
		for _, species := range speciesList {
			if totalPlantsInSpecies > 0 {
				proportion := float64(species["current_population"].(int)) / float64(totalPlantsInSpecies)
				if proportion > 0 {
					diversity -= proportion * math.Log(proportion)
				}
			}
		}
		content.WriteString(fmt.Sprintf("Species diversity index: %.3f\n", diversity))
	}

	// Individual species visualization section
	content.WriteString("\n=== INDIVIDUAL SPECIES DETAILS ===\n")
	content.WriteString("Use arrow keys to navigate species, Enter to select for detailed view\n\n")
	
	if len(speciesList) > 0 {
		// For now, show details for the first species - could be enhanced with selection
		selectedSpecies := speciesList[0]
		content.WriteString(m.renderSpeciesDetail(selectedSpecies))
	} else {
		content.WriteString("No species available for detailed view\n")
	}

	return content.String()
}

// renderSpeciesDetail creates a detailed visual representation of a species
func (m *CLIModel) renderSpeciesDetail(species map[string]interface{}) string {
	var detail strings.Builder
	
	name := species["name"].(string)
	originType := species["origin_type"].(PlantType)
	population := species["current_population"].(int)
	
	detail.WriteString(fmt.Sprintf("🌱 Species: %s\n", name))
	detail.WriteString(fmt.Sprintf("Origin Type: %s | Population: %d\n\n", 
		GetPlantConfigs()[originType].Name, population))
	
	// Find actual plants of this species to analyze
	speciesPlants := make([]*Plant, 0)
	speciesID := species["id"].(int)
	
	// Get the species from the speciation system
	if m.world.SpeciationSystem != nil {
		if speciesObj, exists := m.world.SpeciationSystem.ActiveSpecies[speciesID]; exists {
			speciesPlants = speciesObj.Members
		}
	}
	
	if len(speciesPlants) == 0 {
		detail.WriteString("No living plants found for this species\n")
		return detail.String()
	}
	
	// Visual representation based on plant traits
	detail.WriteString("Species Visual Representation:\n")
	detail.WriteString(m.renderSpeciesVisual(speciesPlants, originType))
	
	// Trait analysis
	detail.WriteString("\nGenetic Trait Analysis:\n")
	detail.WriteString(m.renderSpeciesTraits(speciesPlants))
	
	// Environmental adaptation
	detail.WriteString("\nEnvironmental Adaptation:\n")
	detail.WriteString(m.renderSpeciesHabitat(speciesPlants))
	
	return detail.String()
}

// renderSpeciesVisual creates a visual representation of what the species looks like
func (m *CLIModel) renderSpeciesVisual(plants []*Plant, originType PlantType) string {
	var visual strings.Builder
	
	// Analyze average traits to determine visual characteristics
	avgGrowth := 0.0
	avgSize := 0.0
	avgDefense := 0.0
	avgToxicity := 0.0
	
	for _, plant := range plants {
		if growthTrait, exists := plant.Traits["growth_efficiency"]; exists {
			avgGrowth += growthTrait.Value
		}
		avgSize += plant.Size
		if defenseTrait, exists := plant.Traits["defense"]; exists {
			avgDefense += defenseTrait.Value
		}
		if toxinTrait, exists := plant.Traits["toxin_production"]; exists {
			avgToxicity += toxinTrait.Value
		}
	}
	
	count := float64(len(plants))
	avgGrowth /= count
	avgSize /= count
	avgDefense /= count
	avgToxicity /= count
	
	// Base plant type symbol
	config := GetPlantConfigs()[originType]
	baseSymbol := config.Symbol
	
	// Visual representation with ASCII art
	visual.WriteString(fmt.Sprintf("Base Form: %c (%s)\n", baseSymbol, config.Name))
	
	// Size representation
	sizeDisplay := ""
	if avgSize > 20 {
		sizeDisplay = "████ Very Large"
	} else if avgSize > 15 {
		sizeDisplay = "███  Large"  
	} else if avgSize > 10 {
		sizeDisplay = "██   Medium"
	} else {
		sizeDisplay = "█    Small"
	}
	visual.WriteString(fmt.Sprintf("Size:     %s (%.1f)\n", sizeDisplay, avgSize))
	
	// Defense/armor representation
	defenseDisplay := ""
	if avgDefense > 0.7 {
		defenseDisplay = "▣▣▣ Heavily Armored"
	} else if avgDefense > 0.4 {
		defenseDisplay = "▣▣  Moderately Armored"
	} else if avgDefense > 0.1 {
		defenseDisplay = "▣   Lightly Armored"
	} else {
		defenseDisplay = "     No Armor"
	}
	visual.WriteString(fmt.Sprintf("Defense:  %s (%.1f)\n", defenseDisplay, avgDefense))
	
	// Toxicity representation
	toxinDisplay := ""
	if avgToxicity > 0.7 {
		toxinDisplay = "☠☠☠ Highly Toxic"
	} else if avgToxicity > 0.4 {
		toxinDisplay = "☠☠  Moderately Toxic"
	} else if avgToxicity > 0.1 {
		toxinDisplay = "☠   Mildly Toxic"
	} else {
		toxinDisplay = "     Non-toxic"
	}
	visual.WriteString(fmt.Sprintf("Toxicity: %s (%.1f)\n", toxinDisplay, avgToxicity))
	
	// Growth pattern
	growthDisplay := ""
	if avgGrowth > 0.7 {
		growthDisplay = "🌿🌿🌿 Rapid Growth"
	} else if avgGrowth > 0.4 {
		growthDisplay = "🌿🌿   Moderate Growth"
	} else {
		growthDisplay = "🌿     Slow Growth"
	}
	visual.WriteString(fmt.Sprintf("Growth:   %s (%.1f)\n", growthDisplay, avgGrowth))
	
	// Create a simple "profile" view with enhanced blocky design
	visual.WriteString("\nProfile View (Blocky Style):\n")
	
	// Top crown based on growth (more elaborate)
	if avgGrowth > 0.7 {
		visual.WriteString("       ╭🌿╮ ╭🌿╮\n")
		visual.WriteString("      ╱🌿🌿🌿🌿╲\n")
	} else if avgGrowth > 0.4 {
		visual.WriteString("        ╭🌿╮\n")
		visual.WriteString("       ╱🌿🌿╲\n")
	} else {
		visual.WriteString("        🌿\n")
		visual.WriteString("       ╱▲╲\n")
	}
	
	// Main body sections vary by size with more detail
	bodyHeight := int(avgSize/5) + 2
	for i := 0; i < bodyHeight; i++ {
		if avgDefense > 0.7 {
			// Heavy armor plating
			visual.WriteString("      ┃▣▣▣▣▣┃\n")
		} else if avgDefense > 0.4 {
			// Moderate armor
			visual.WriteString("      ┃▣ ● ▣┃\n")
		} else {
			// Normal body
			visual.WriteString("      ┃  ●  ┃\n")
		}
		
		// Add toxicity indicators inside body
		if avgToxicity > 0.5 && i == bodyHeight/2 {
			if avgDefense > 0.4 {
				visual.WriteString("      ┃▣ ☠ ▣┃\n")
			} else {
				visual.WriteString("      ┃  ☠  ┃\n")
			}
		}
	}
	
	// Base varies by root strength
	if avgGrowth > 0.5 {
		visual.WriteString("      └─────┘\n")
		visual.WriteString("       ╱╲╱╲╱╲  (strong roots)\n")
		visual.WriteString("      ╱  ╲╱  ╲\n")
	} else {
		visual.WriteString("      └─────┘\n")
		visual.WriteString("       ╱╲╱╲  (roots)\n")
	}
	
	return visual.String()
}

// renderSpeciesTraits shows genetic trait distribution for the species
func (m *CLIModel) renderSpeciesTraits(plants []*Plant) string {
	var traits strings.Builder
	
	traitSums := make(map[string]float64)
	traitCounts := make(map[string]int)
	
	// Collect all traits
	for _, plant := range plants {
		for traitName, trait := range plant.Traits {
			traitSums[traitName] += trait.Value
			traitCounts[traitName]++
		}
	}
	
	// Calculate averages and display as bars
	for traitName, sum := range traitSums {
		avg := sum / float64(traitCounts[traitName])
		bars := int(math.Abs(avg) * 10)
		if bars > 10 { bars = 10 }
		
		traits.WriteString(fmt.Sprintf("%-18s ", traitName))
		
		// Visual bar representation
		if avg >= 0 {
			traits.WriteString("[")
			for i := 0; i < 10; i++ {
				if i < bars {
					traits.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("█"))
				} else {
					traits.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("░"))
				}
			}
			traits.WriteString("]")
		} else {
			traits.WriteString("[")
			for i := 0; i < 10; i++ {
				if i < bars {
					traits.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("█"))
				} else {
					traits.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("░"))
				}
			}
			traits.WriteString("]")
		}
		traits.WriteString(fmt.Sprintf(" %.3f\n", avg))
	}
	
	return traits.String()
}

// renderSpeciesHabitat shows where the species is found and environmental preferences
func (m *CLIModel) renderSpeciesHabitat(plants []*Plant) string {
	var habitat strings.Builder
	
	// Analyze biome distribution
	biomeCount := make(map[BiomeType]int)
	totalPlants := len(plants)
	
	for _, plant := range plants {
		x, y := int(plant.Position.X), int(plant.Position.Y)
		if x >= 0 && x < m.world.Config.GridWidth && y >= 0 && y < m.world.Config.GridHeight {
			biome := m.world.Grid[y][x].Biome
			biomeCount[biome]++
		}
	}
	
	habitat.WriteString("Habitat Distribution:\n")
	for biomeType, count := range biomeCount {
		if count > 0 {
			percentage := float64(count) / float64(totalPlants) * 100
			biomeName := m.world.Biomes[biomeType].Name
			habitat.WriteString(fmt.Sprintf("  %s: %d plants (%.1f%%)\n", biomeName, count, percentage))
		}
	}
	
	// Find preferred biome
	maxCount := 0
	var preferredBiome BiomeType
	for biome, count := range biomeCount {
		if count > maxCount {
			maxCount = count
			preferredBiome = biome
		}
	}
	
	if maxCount > 0 {
		habitat.WriteString(fmt.Sprintf("\nPreferred Habitat: %s\n", m.world.Biomes[preferredBiome].Name))
	}
	
	return habitat.String()
}

// Helper map for season names (add this near other constants)
var seasonNames = map[Season]string{
	Spring: "Spring",
	Summer: "Summer",
	Autumn: "Autumn",
	Winter: "Winter",
}

// networkView renders plant network information and underground connections
func (m CLIModel) networkView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Plant Network System") + "\n\n")

	if m.world.PlantNetworkSystem == nil {
		content.WriteString("Plant network system not initialized\n")
		return content.String()
	}

	// Network overview
	stats := m.world.PlantNetworkSystem.GetNetworkStats()
	content.WriteString("=== NETWORK OVERVIEW ===\n")
	content.WriteString(fmt.Sprintf("Total connections: %v\n", stats["total_connections"]))
	content.WriteString(fmt.Sprintf("Active connections: %v\n", stats["active_connections"]))
	content.WriteString(fmt.Sprintf("Network clusters: %v\n", stats["cluster_count"]))
	content.WriteString(fmt.Sprintf("Chemical signals: %v\n", stats["active_signals"]))
	content.WriteString(fmt.Sprintf("Average connection strength: %.3f\n", stats["avg_connection_strength"]))
	content.WriteString(fmt.Sprintf("Network efficiency: %.1f%%\n", stats["network_efficiency"].(float64)*100))
	// Connection types breakdown
	content.WriteString("\n=== CONNECTION TYPES ===\n")
	if connectionTypes, ok := stats["connection_types"].(map[NetworkConnectionType]int); ok {
		typeNames := map[NetworkConnectionType]string{
			ConnectionMycorrhizal: "Mycorrhizal",
			ConnectionRoot:        "Root",
			ConnectionChemical:    "Chemical",
		}

		for connType, count := range connectionTypes {
			if name, exists := typeNames[connType]; exists {
				content.WriteString(fmt.Sprintf("  %s: %d\n", name, count))
			}
		}
	}

	// Chemical signals activity
	content.WriteString("\n=== CHEMICAL SIGNALS ===\n")
	if signals, ok := stats["signal_activity"].(map[ChemicalSignalType]int); ok {
		signalNames := map[ChemicalSignalType]string{
			SignalNutrientAvailable: "Nutrient sharing",
			SignalNutrientNeeded:    "Nutrient requests",
			SignalThreatDetected:    "Threat warnings",
			SignalOptimalGrowth:     "Growth signals",
			SignalReproductionReady: "Reproduction aid",
			SignalToxicConditions:   "Toxin warnings",
		}

		totalSignals := 0
		for _, count := range signals {
			totalSignals += count
		}

		if totalSignals > 0 {
			for signalType, count := range signals {
				if name, exists := signalNames[signalType]; exists && count > 0 {
					percentage := float64(count) / float64(totalSignals) * 100
					content.WriteString(fmt.Sprintf("  %s: %d (%.1f%%)\n", name, count, percentage))
				}
			}
		} else {
			content.WriteString("  No active chemical signals\n")
		}
	}

	// Network clusters
	content.WriteString("\n=== NETWORK CLUSTERS ===\n")
	if clusters, ok := stats["clusters"].([]map[string]interface{}); ok && len(clusters) > 0 {
		content.WriteString("Cluster ID   Size   Avg Health   Efficiency   Plant Types\n")
		content.WriteString("──────────────────────────────────────────────────────────\n")

		for i, cluster := range clusters {
			if i >= 10 { // Limit display to prevent overflow
				content.WriteString(fmt.Sprintf("... and %d more clusters\n", len(clusters)-10))
				break
			}

			id := cluster["id"].(int)
			size := cluster["size"].(int)
			avgHealth := cluster["avg_health"].(float64)
			efficiency := cluster["efficiency"].(float64)
			plantTypes := cluster["plant_types"].([]string)

			// Limit plant types display
			typesStr := strings.Join(plantTypes, ",")
			if len(typesStr) > 15 {
				typesStr = typesStr[:12] + "..."
			}

			content.WriteString(fmt.Sprintf("%-12d %-6d %-12.2f %-12.1f%% %s\n",
				id, size, avgHealth, efficiency*100, typesStr))
		}
	} else {
		content.WriteString("No network clusters formed yet\n")
	}

	// Resource sharing statistics
	content.WriteString("\n=== RESOURCE SHARING ===\n")
	if sharing, ok := stats["resource_sharing"].(map[string]interface{}); ok {
		content.WriteString(fmt.Sprintf("Total transfers this tick: %v\n", sharing["transfers_this_tick"]))
		content.WriteString(fmt.Sprintf("Total resources transferred: %.1f\n", sharing["total_resources_transferred"]))
		content.WriteString(fmt.Sprintf("Average transfer efficiency: %.1f%%\n", sharing["avg_transfer_efficiency"].(float64)*100))

		if beneficiaries, exists := sharing["recent_beneficiaries"].(int); exists {
			content.WriteString(fmt.Sprintf("Plants aided this cycle: %v\n", beneficiaries))
		}
	}

	// Recent network events
	content.WriteString("\n=== RECENT NETWORK ACTIVITY ===\n")
	if events, ok := stats["recent_events"].([]string); ok && len(events) > 0 {
		for i, event := range events {
			if i >= 5 { // Show last 5 events
				break
			}
			content.WriteString(fmt.Sprintf("• %s\n", event))
		}
	} else {
		content.WriteString("No recent network activity\n")
	}

	// Network health and maintenance
	content.WriteString("\n=== NETWORK HEALTH ===\n")
	if health, ok := stats["network_health"].(map[string]interface{}); ok {
		content.WriteString(fmt.Sprintf("Healthy connections: %v%%\n", int(health["healthy_percentage"].(float64)*100)))
		content.WriteString(fmt.Sprintf("Degrading connections: %v\n", health["degrading_connections"]))
		content.WriteString(fmt.Sprintf("Connections lost this tick: %v\n", health["connections_lost"]))
		content.WriteString(fmt.Sprintf("New connections formed: %v\n", health["new_connections"]))
	}

	return content.String()
}

// dnaView displays DNA and genetic information
func (m *CLIModel) dnaView() string {
	var content strings.Builder
	content.WriteString("=== DNA ANALYSIS ===\n\n")
	
	if m.world.DNASystem == nil {
		content.WriteString("DNA system not available\n")
		return content.String()
	}
	
	// Show DNA system statistics
	content.WriteString("DNA SYSTEM STATUS:\n")
	content.WriteString(fmt.Sprintf("Active trait-gene mappings: %d\n", len(m.world.DNASystem.TraitToGene)))
	content.WriteString(fmt.Sprintf("Gene length definitions: %d\n", len(m.world.DNASystem.GeneLength)))
	
	// Find sample entities with DNA
	sampleCount := 0
	for _, entity := range m.world.AllEntities {
		if !entity.IsAlive || sampleCount >= 5 {
			continue
		}
		
		// Look for cellular organism
		if organism := m.world.CellularSystem.OrganismMap[entity.ID]; organism != nil {
			if len(organism.Cells) > 0 && organism.Cells[0].DNA != nil {
				dna := organism.Cells[0].DNA
				
				content.WriteString(fmt.Sprintf("\n--- Entity %d DNA Analysis ---\n", entity.ID))
				content.WriteString(fmt.Sprintf("Species: %s\n", entity.Species))
				content.WriteString(fmt.Sprintf("Generation: %d\n", dna.Generation))
				content.WriteString(fmt.Sprintf("Chromosomes: %d\n", len(dna.Chromosomes)))
				content.WriteString(fmt.Sprintf("Total mutations: %d\n", dna.Mutations))
				
				// Show DNA sequence sample
				dnaString := m.world.DNASystem.GetDNAString(dna, 50)
				content.WriteString(fmt.Sprintf("DNA Sample: %s\n", dnaString))
				
				// Show trait expression
				content.WriteString("Trait Expression:\n")
				for traitName := range entity.Traits {
					originalValue := entity.GetTrait(traitName)
					dnaValue := m.world.DNASystem.ExpressTrait(dna, traitName)
					content.WriteString(fmt.Sprintf("  %s: %.3f (DNA: %.3f)\n", traitName, originalValue, dnaValue))
				}
				
				sampleCount++
			}
		}
	}
	
	if sampleCount == 0 {
		content.WriteString("\nNo DNA samples available for analysis\n")
	}
	
	return content.String()
}

// cellularView displays cellular-level information with visual representation
func (m *CLIModel) cellularView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Cellular Analysis & Visualization") + "\n\n")
	
	if m.world.CellularSystem == nil {
		content.WriteString("Cellular system not available\n")
		return content.String()
	}
	
	// System statistics
	stats := m.world.CellularSystem.GetCellularSystemStats()
	content.WriteString("CELLULAR SYSTEM STATUS:\n")
	content.WriteString(fmt.Sprintf("Total organisms: %v\n", stats["total_organisms"]))
	content.WriteString(fmt.Sprintf("Total cells: %v\n", stats["total_cells"]))
	content.WriteString(fmt.Sprintf("Next cell ID: %v\n", stats["next_cell_id"]))
	
	// Complexity distribution
	if complexityDist, ok := stats["complexity_distribution"].(map[int]int); ok {
		content.WriteString("\nComplexity Distribution:\n")
		for level := 1; level <= 5; level++ {
			count := complexityDist[level]
			content.WriteString(fmt.Sprintf("  Level %d: %d organisms\n", level, count))
		}
	}
	
	// Visual representation of selected organism
	content.WriteString("\n=== INDIVIDUAL ORGANISM VISUALIZATION ===\n")
	selectedOrganism := m.getSelectedOrganism()
	if selectedOrganism != nil {
		content.WriteString(m.renderOrganismVisual(selectedOrganism))
	} else {
		content.WriteString("No organism selected. Use arrows to navigate and Enter to select.\n")
	}
	
	// Sample organism details with enhanced visualization
	content.WriteString("\n=== ORGANISM SAMPLES ===\n")
	sampleCount := 0
	for entityID, organism := range m.world.CellularSystem.OrganismMap {
		if sampleCount >= 3 {
			break
		}
		
		content.WriteString(fmt.Sprintf("\n--- Organism %d ---\n", entityID))
		content.WriteString(fmt.Sprintf("Cells: %d\n", len(organism.Cells)))
		content.WriteString(fmt.Sprintf("Complexity Level: %d\n", organism.ComplexityLevel))
		content.WriteString(fmt.Sprintf("Total Energy: %.1f\n", organism.TotalEnergy))
		content.WriteString(fmt.Sprintf("Cell Divisions: %d\n", organism.CellDivisions))
		content.WriteString(fmt.Sprintf("Generation: %d\n", organism.Generation))
		
		// Visual cell layout representation
		content.WriteString("Cell Layout Visual:\n")
		content.WriteString(m.renderCellLayout(organism))
		
		// Organ systems
		content.WriteString("Organ Systems:\n")
		for systemName, cellIDs := range organism.OrganSystems {
			content.WriteString(fmt.Sprintf("  %s: %d cells\n", systemName, len(cellIDs)))
		}
		
		// Sample cell details with visual representation
		if len(organism.Cells) > 0 {
			cell := organism.Cells[0]
			content.WriteString(fmt.Sprintf("Sample Cell (ID %d):\n", cell.ID))
			content.WriteString(fmt.Sprintf("  Type: %s\n", m.world.CellularSystem.CellTypeNames[cell.Type]))
			content.WriteString(fmt.Sprintf("  Size: %.1f μm\n", cell.Size))
			content.WriteString(fmt.Sprintf("  Energy: %.1f\n", cell.Energy))
			content.WriteString(fmt.Sprintf("  Health: %.1f%%\n", cell.Health*100))
			content.WriteString(fmt.Sprintf("  Age: %d ticks\n", cell.Age))
			content.WriteString(fmt.Sprintf("  Activity: %.1f%%\n", cell.Activity*100))
			content.WriteString(fmt.Sprintf("  Organelles: %d types\n", len(cell.Organelles)))
			
			// Visual representation of the cell
			content.WriteString("  Cell Visual:\n")
			content.WriteString(m.renderCellVisual(cell))
		}
		
		sampleCount++
	}
	
	return content.String()
}

// getSelectedOrganism returns the currently selected organism (for now, just return the first one)
func (m *CLIModel) getSelectedOrganism() *CellularOrganism {
	if m.world.CellularSystem == nil || len(m.world.CellularSystem.OrganismMap) == 0 {
		return nil
	}
	
	// For now, return the first organism - later this could be enhanced with selection
	for _, organism := range m.world.CellularSystem.OrganismMap {
		return organism
	}
	return nil
}

// renderOrganismVisual creates a visual representation of an entire organism
func (m *CLIModel) renderOrganismVisual(organism *CellularOrganism) string {
	var visual strings.Builder
	
	visual.WriteString(fmt.Sprintf("🦠 Organism (Complexity Level %d)\n", organism.ComplexityLevel))
	visual.WriteString(fmt.Sprintf("Energy: %.1f | Cells: %d | Generation: %d\n\n", 
		organism.TotalEnergy, len(organism.Cells), organism.Generation))
	
	// Create a simple grid representation based on cell count and type
	cellsPerRow := int(math.Sqrt(float64(len(organism.Cells)))) + 1
	if cellsPerRow > 10 { cellsPerRow = 10 }
	
	visual.WriteString("Cellular Structure:\n")
	for i, cell := range organism.Cells {
		if i > 0 && i%cellsPerRow == 0 {
			visual.WriteString("\n")
		}
		
		// Get cell symbol based on type
		symbol := m.getCellSymbol(int(cell.Type))
		
		// Color based on health and energy
		if cell.Health > 0.8 && cell.Energy > 50 {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(symbol)) // Bright green
		} else if cell.Health > 0.5 {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(symbol)) // Yellow
		} else {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(symbol)) // Red
		}
		visual.WriteString(" ")
	}
	visual.WriteString("\n\n")
	
	// Organ system visualization
	visual.WriteString("Organ Systems:\n")
	for systemName, cellIDs := range organism.OrganSystems {
		visual.WriteString(fmt.Sprintf("  %s (%d cells): ", systemName, len(cellIDs)))
		for i, cellID := range cellIDs {
			if i >= 10 { // Limit display
				visual.WriteString("...")
				break
			}
			cell := m.findCellByID(organism, cellID)
			if cell != nil {
				symbol := m.getCellSymbol(int(cell.Type))
				visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(symbol))
			}
		}
		visual.WriteString("\n")
	}
	
	return visual.String()
}

// renderCellLayout creates a visual layout of cells in an organism
func (m *CLIModel) renderCellLayout(organism *CellularOrganism) string {
	var layout strings.Builder
	
	// Simple grid layout based on cell count
	cellsPerRow := 8
	for i, cell := range organism.Cells {
		if i > 0 && i%cellsPerRow == 0 {
			layout.WriteString("\n")
		}
		
		symbol := m.getCellSymbol(int(cell.Type))
		
		// Color coding: green=healthy, yellow=medium, red=unhealthy
		health := cell.Health
		if health > 0.7 {
			layout.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Render(symbol))
		} else if health > 0.4 {
			layout.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(symbol))
		} else {
			layout.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(symbol))
		}
		layout.WriteString(" ")
		
		// Limit display to avoid too large grids
		if i >= 63 { // 8x8 grid max
			layout.WriteString("...")
			break
		}
	}
	layout.WriteString("\n")
	
	return layout.String()
}

// renderCellVisual creates a detailed visual representation of a single cell
func (m *CLIModel) renderCellVisual(cell *Cell) string {
	var visual strings.Builder
	
	symbol := m.getCellSymbol(int(cell.Type))
	
	// Main cell representation with size indication
	sizeIndicator := ""
	if cell.Size > 20 {
		sizeIndicator = "●●●" // Large cell
	} else if cell.Size > 10 {
		sizeIndicator = "●●"  // Medium cell
	} else {
		sizeIndicator = "●"   // Small cell
	}
	
	visual.WriteString(fmt.Sprintf("    [%s] %s %s\n", symbol, sizeIndicator, m.world.CellularSystem.CellTypeNames[cell.Type]))
	
	// Health bar
	healthBars := int(cell.Health * 10)
	visual.WriteString("    Health: [")
	for i := 0; i < 10; i++ {
		if i < healthBars {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("█"))
		} else {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("░"))
		}
	}
	visual.WriteString("]\n")
	
	// Energy bar
	energyBars := int(cell.Energy / 10)
	if energyBars > 10 { energyBars = 10 }
	visual.WriteString("    Energy:  [")
	for i := 0; i < 10; i++ {
		if i < energyBars {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("█"))
		} else {
			visual.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("░"))
		}
	}
	visual.WriteString("]\n")
	
	// Organelles representation
	if len(cell.Organelles) > 0 {
		visual.WriteString("    Organelles: ")
		organelleSymbols := map[OrganelleType]string{
			0: "⬣", // Nucleus
			1: "⚡", // Mitochondria
			2: "🌱", // Chloroplast
			3: "⬢", // Ribosome
			4: "💧", // Vacuole
			5: "📦", // Golgi
			6: "🕸", // ER
			7: "🗑", // Lysosome
		}
		
		for organelleType, organelle := range cell.Organelles {
			if symbol, exists := organelleSymbols[organelleType]; exists {
				count := organelle.Count
				for i := 0; i < count && i < 3; i++ { // Limit display
					visual.WriteString(symbol)
				}
				if count > 3 {
					visual.WriteString(fmt.Sprintf("(%d)", count))
				}
			}
		}
		visual.WriteString("\n")
	}
	
	return visual.String()
}

// getCellSymbol returns a symbol representing the cell type
func (m *CLIModel) getCellSymbol(cellType int) string {
	symbols := map[int]string{
		0: "S", // Stem
		1: "N", // Nerve
		2: "M", // Muscle
		3: "D", // Digestive
		4: "R", // Reproductive
		5: "F", // Defensive
		6: "P", // Photosynthetic
		7: "T", // Storage
	}
	
	if symbol, exists := symbols[cellType]; exists {
		return symbol
	}
	return "?"
}

// findCellByID finds a cell in an organism by its ID
func (m *CLIModel) findCellByID(organism *CellularOrganism, cellID int) *Cell {
	for _, cell := range organism.Cells {
		if cell.ID == cellID {
			return cell
		}
	}
	return nil
}

// evolutionView displays macro-evolution information
func (m *CLIModel) evolutionView() string {
	var content strings.Builder
	content.WriteString("=== MACRO EVOLUTION ===\n\n")
	
	if m.world.MacroEvolutionSystem == nil {
		content.WriteString("Macro evolution system not available\n")
		return content.String()
	}
	
	// System statistics
	stats := m.world.MacroEvolutionSystem.GetEvolutionStats()
	content.WriteString("EVOLUTIONARY OVERVIEW:\n")
	content.WriteString(fmt.Sprintf("Total species tracked: %v\n", stats["total_species"]))
	content.WriteString(fmt.Sprintf("Living species: %v\n", stats["living_species"]))
	content.WriteString(fmt.Sprintf("Extinct species: %v\n", stats["extinct_species"]))
	content.WriteString(fmt.Sprintf("Total evolutionary events: %v\n", stats["total_events"]))
	content.WriteString(fmt.Sprintf("Recent events (last 100 ticks): %v\n", stats["recent_events"]))
	content.WriteString(fmt.Sprintf("Extinction events: %v\n", stats["extinction_events"]))
	
	// Phylogenetic tree info
	if treeDepth, ok := stats["tree_depth"]; ok {
		content.WriteString(fmt.Sprintf("Phylogenetic tree depth: %v\n", treeDepth))
	}
	if treeNodes, ok := stats["tree_nodes"]; ok {
		content.WriteString(fmt.Sprintf("Tree nodes: %v\n", treeNodes))
	}
	
	// Recent evolutionary events
	content.WriteString("\n=== RECENT EVOLUTIONARY EVENTS ===\n")
	recentEvents := m.world.MacroEvolutionSystem.GetRecentEvents(5)
	if len(recentEvents) == 0 {
		content.WriteString("No recent evolutionary events\n")
	} else {
		for _, event := range recentEvents {
			content.WriteString(fmt.Sprintf("Tick %d [%s]: %s\n", 
				event.Tick, strings.ToUpper(event.Type), event.Description))
			if event.Significance > 0.5 {
				content.WriteString("  *** HIGHLY SIGNIFICANT ***\n")
			}
		}
	}
	
	// Feedback Loop Evolution Data
	content.WriteString("\n=== FEEDBACK LOOP EVOLUTION ===\n")
	adaptationStats := m.calculateAdaptationStats()
	content.WriteString(fmt.Sprintf("Entities with dietary adaptations: %d\n", adaptationStats["dietary_memory_count"]))
	content.WriteString(fmt.Sprintf("Entities with environmental adaptations: %d\n", adaptationStats["env_memory_count"]))
	content.WriteString(fmt.Sprintf("Average dietary fitness: %.3f\n", adaptationStats["avg_dietary_fitness"]))
	content.WriteString(fmt.Sprintf("Average environmental fitness: %.3f\n", adaptationStats["avg_env_fitness"]))
	
	// Show evolutionary pressure indicators
	highPressureCount := 0
	for _, population := range m.world.Populations {
		for _, entity := range population.Entities {
			if !entity.IsAlive {
				continue
			}
			
			highPressure := false
			if entity.EnvironmentalMemory != nil && entity.EnvironmentalMemory.AdaptationFitness < 0.8 {
				highPressure = true
			}
			if entity.DietaryMemory != nil && entity.DietaryMemory.DietaryFitness < 0.6 {
				highPressure = true
			}
			
			if highPressure {
				highPressureCount++
			}
		}
	}
	content.WriteString(fmt.Sprintf("Entities under evolutionary pressure: %d\n", highPressureCount))
	
	// Species lineages
	content.WriteString("\n=== SPECIES LINEAGES ===\n")
	lineageCount := 0
	for speciesName, lineage := range m.world.MacroEvolutionSystem.SpeciesLineages {
		if lineageCount >= 8 { // Limit display
			remaining := len(m.world.MacroEvolutionSystem.SpeciesLineages) - lineageCount
			content.WriteString(fmt.Sprintf("... and %d more species\n", remaining))
			break
		}
		
		status := "LIVING"
		if lineage.ExtinctionTick != 0 {
			status = "EXTINCT"
		}
		
		content.WriteString(fmt.Sprintf("%s [%s]:\n", speciesName, status))
		content.WriteString(fmt.Sprintf("  Origin: Tick %d", lineage.OriginTick))
		if lineage.ParentSpecies != "" {
			content.WriteString(fmt.Sprintf(" (from %s)", lineage.ParentSpecies))
		}
		content.WriteString("\n")
		
		if lineage.ExtinctionTick != 0 {
			duration := lineage.ExtinctionTick - lineage.OriginTick
			content.WriteString(fmt.Sprintf("  Extinction: Tick %d (survived %d ticks)\n", 
				lineage.ExtinctionTick, duration))
		}
		
		content.WriteString(fmt.Sprintf("  Peak population: %d\n", lineage.PeakPopulation))
		content.WriteString(fmt.Sprintf("  Child species: %d\n", len(lineage.ChildSpecies)))
		content.WriteString(fmt.Sprintf("  Adaptations: %d\n", len(lineage.Adaptations)))
		
		// Show niches
		if len(lineage.Niches) > 0 {
			content.WriteString(fmt.Sprintf("  Niches: %s\n", strings.Join(lineage.Niches, ", ")))
		}
		
		lineageCount++
	}
	
	return content.String()
}

// topologyView displays world terrain and geological information with enhanced visualization
func (m *CLIModel) topologyView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("World Topology & Underground Visualization") + "\n\n")
	
	if m.world.TopologySystem == nil {
		content.WriteString("Topology system not available\n")
		return content.String()
	}
	
	// System statistics
	stats := m.world.TopologySystem.GetTopologyStats()
	content.WriteString("TOPOLOGY OVERVIEW:\n")
	content.WriteString(fmt.Sprintf("World size: %dx%d\n", m.world.TopologySystem.Width, m.world.TopologySystem.Height))
	content.WriteString(fmt.Sprintf("Sea level: %.2f\n", stats["sea_level"]))
	content.WriteString(fmt.Sprintf("Average elevation: %.3f\n", stats["avg_elevation"]))
	content.WriteString(fmt.Sprintf("Average slope: %.3f\n", stats["avg_slope"]))
	content.WriteString(fmt.Sprintf("Water coverage: %.1f%%\n", stats["water_coverage"].(float64)*100))
	content.WriteString(fmt.Sprintf("Erosion rate: %.6f\n", stats["erosion_rate"]))
	content.WriteString(fmt.Sprintf("Tectonic activity: %.2f\n", stats["tectonic_activity"]))
	
	// 3D-style visualization controls
	content.WriteString("\n=== VIEWING ANGLES ===\n")
	content.WriteString("🔍 [1] Surface View (Default) | [2] Cross-Section | [3] Underground | [4] Isometric\n")
	content.WriteString("Use number keys to switch viewing angles\n")
	
	// Enhanced topographic map with underground features
	content.WriteString("\n=== ENHANCED TOPOGRAPHIC MAP ===\n")
	content.WriteString(m.renderTopographicMap())
	
	// Underground features visualization
	content.WriteString("\n=== UNDERGROUND FEATURES ===\n")
	content.WriteString(m.renderUndergroundMap())
	
	// Cross-section view
	content.WriteString("\n=== CROSS-SECTION VIEW ===\n")
	content.WriteString(m.renderCrossSectionView())
	
	// Terrain features
	content.WriteString(fmt.Sprintf("\nTerrain features: %v\n", stats["terrain_features"]))
	content.WriteString(fmt.Sprintf("Water bodies: %v\n", stats["water_bodies"]))
	content.WriteString(fmt.Sprintf("Active geological events: %v\n", stats["geological_events"]))
	
	// Active geological events
	if len(m.world.TopologySystem.GeologicalEvents) > 0 {
		content.WriteString("\n=== ACTIVE GEOLOGICAL EVENTS ===\n")
		for _, event := range m.world.TopologySystem.GeologicalEvents {
			content.WriteString(fmt.Sprintf("%s (ID %d):\n", strings.ToTitle(event.Type), event.ID))
			content.WriteString(fmt.Sprintf("  Center: (%.1f, %.1f)\n", event.Center.X, event.Center.Y))
			content.WriteString(fmt.Sprintf("  Radius: %.1f\n", event.Radius))
			content.WriteString(fmt.Sprintf("  Intensity: %.2f\n", event.Intensity))
			content.WriteString(fmt.Sprintf("  Duration remaining: %d ticks\n", event.Duration))
		}
	}
	
	// Major terrain features with visual representation
	content.WriteString("\n=== MAJOR TERRAIN FEATURES ===\n")
	featureCount := 0
	for _, feature := range m.world.TopologySystem.TerrainFeatures {
		if featureCount >= 8 { // Limit display
			remaining := len(m.world.TopologySystem.TerrainFeatures) - featureCount
			content.WriteString(fmt.Sprintf("... and %d more features\n", remaining))
			break
		}
		
		typeName := m.world.TopologySystem.GetTerrainTypeName(feature.Type)
		featureSymbol := m.getTerrainFeatureSymbol(feature.Type)
		
		content.WriteString(fmt.Sprintf("%s %s (ID %d):\n", featureSymbol, typeName, feature.ID))
		content.WriteString(fmt.Sprintf("  Center: (%.1f, %.1f)\n", feature.Center.X, feature.Center.Y))
		content.WriteString(fmt.Sprintf("  Size: %.1f\n", feature.Radius))
		content.WriteString(fmt.Sprintf("  Height: %.3f\n", feature.Height))
		content.WriteString(fmt.Sprintf("  Age: %d ticks\n", feature.Age))
		content.WriteString(fmt.Sprintf("  Stability: %.2f\n", feature.Stability))
		content.WriteString(fmt.Sprintf("  Composition: %s\n", feature.Composition))
		
		// Visual mini-profile for each feature
		content.WriteString("  Profile: ")
		content.WriteString(m.renderFeatureProfile(feature))
		content.WriteString("\n")
		
		featureCount++
	}
	
	// Major water bodies with flow visualization
	if len(m.world.TopologySystem.WaterBodies) > 0 {
		content.WriteString("\n=== WATER BODIES ===\n")
		waterCount := 0
		for _, waterBody := range m.world.TopologySystem.WaterBodies {
			if waterCount >= 5 { // Limit display
				remaining := len(m.world.TopologySystem.WaterBodies) - waterCount
				content.WriteString(fmt.Sprintf("... and %d more water bodies\n", remaining))
				break
			}
			
			waterSymbol := m.getWaterBodySymbol(waterBody.Type)
			content.WriteString(fmt.Sprintf("%s %s (ID %d):\n", waterSymbol, strings.ToTitle(waterBody.Type), waterBody.ID))
			content.WriteString(fmt.Sprintf("  Depth: %.2f\n", waterBody.Depth))
			content.WriteString(fmt.Sprintf("  Flow: %.2f\n", waterBody.Flow))
			content.WriteString(fmt.Sprintf("  Salinity: %.1f%%\n", waterBody.Salinity*100))
			content.WriteString(fmt.Sprintf("  Points: %d\n", len(waterBody.Points)))
			
			// Flow direction visualization
			if waterBody.Flow > 0.1 {
				content.WriteString("  Flow: ")
				content.WriteString(m.renderWaterFlow(waterBody))
				content.WriteString("\n")
			}
			
			waterCount++
		}
	}
	
	return content.String()
}

// renderTopographicMap creates an enhanced topographic map
func (m *CLIModel) renderTopographicMap() string {
	var topo strings.Builder
	
	topo.WriteString("Surface Elevation Map (Minecraft/Rimworld style):\n")
	
	// Use a smaller sample of the world for display
	sampleWidth := min(m.world.TopologySystem.Width, 40)
	sampleHeight := min(m.world.TopologySystem.Height, 20)
	
	for y := 0; y < sampleHeight; y++ {
		for x := 0; x < sampleWidth; x++ {
			if x < m.world.TopologySystem.Width && y < m.world.TopologySystem.Height {
				cell := m.world.TopologySystem.TopologyGrid[x][y]
				symbol := m.getElevationSymbol(cell.Elevation)
				style := m.getElevationColor(cell.Elevation)
				topo.WriteString(style.Render(symbol))
			} else {
				topo.WriteString(" ")
			}
		}
		topo.WriteString("\n")
	}
	
	// Legend
	topo.WriteString("\nElevation Legend:\n")
	topo.WriteString("▲ High mountains   ▲ Medium mountains   △ Hills   . Plains   ~ Low areas   ≈ Water\n")
	
	return topo.String()
}

// renderUndergroundMap shows underground features like tunnels and caves
func (m *CLIModel) renderUndergroundMap() string {
	var underground strings.Builder
	
	underground.WriteString("Underground Structure Map:\n")
	
	// Check for environmental modifications (tunnels, burrows, etc.)
	if m.world.EnvironmentalModSystem != nil {
		sampleWidth := min(m.world.Config.GridWidth, 40)
		sampleHeight := min(m.world.Config.GridHeight, 20)
		
		// Create underground map
		undergroundGrid := make([][]string, sampleHeight)
		for y := range undergroundGrid {
			undergroundGrid[y] = make([]string, sampleWidth)
			for x := range undergroundGrid[y] {
				undergroundGrid[y][x] = "░" // Empty underground
			}
		}
		
		// Mark underground modifications
		for _, mod := range m.world.EnvironmentalModSystem.Modifications {
			x, y := int(mod.Position.X), int(mod.Position.Y)
			if x >= 0 && x < sampleWidth && y >= 0 && y < sampleHeight {
				switch mod.Type {
				case 0: // Tunnel
					undergroundGrid[y][x] = "═"
				case 1: // Burrow
					undergroundGrid[y][x] = "○"
				case 2: // Cache
					undergroundGrid[y][x] = "□"
				case 9: // Workshop (underground)
					undergroundGrid[y][x] = "⚒"
				case 10: // Farm (root system)
					undergroundGrid[y][x] = "┼"
				default:
					undergroundGrid[y][x] = "▓"
				}
			}
		}
		
		// Render underground map
		for y := 0; y < sampleHeight; y++ {
			for x := 0; x < sampleWidth; x++ {
				symbol := undergroundGrid[y][x]
				if symbol == "░" {
					underground.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(symbol))
				} else {
					underground.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(symbol))
				}
			}
			underground.WriteString("\n")
		}
		
		underground.WriteString("\nUnderground Legend:\n")
		underground.WriteString("═ Tunnels   ○ Burrows   □ Caches   ⚒ Workshops   ┼ Root Systems   ░ Empty\n")
	} else {
		underground.WriteString("No underground modification system available\n")
	}
	
	return underground.String()
}

// renderCrossSectionView shows a cross-section of the world
func (m *CLIModel) renderCrossSectionView() string {
	var section strings.Builder
	
	section.WriteString("World Cross-Section (Center slice):\n")
	
	// Take a vertical slice through the center of the world
	centerX := m.world.TopologySystem.Width / 2
	width := min(60, m.world.TopologySystem.Height) // Display width
	
	// Surface line
	section.WriteString("Surface: ")
	for y := 0; y < width; y++ {
		if y < m.world.TopologySystem.Height {
			cell := m.world.TopologySystem.TopologyGrid[centerX][y]
			symbol := m.getElevationSymbol(cell.Elevation)
			section.WriteString(symbol)
		}
	}
	section.WriteString("\n")
	
	// Underground layers (simulated)
	for layer := 1; layer <= 5; layer++ {
		section.WriteString(fmt.Sprintf("Layer %d: ", layer))
		for y := 0; y < width; y++ {
			// Simulate underground layers based on surface topology
			if y < m.world.TopologySystem.Height {
				cell := m.world.TopologySystem.TopologyGrid[centerX][y]
				symbol := m.getUndergroundLayerSymbol(cell, layer)
				section.WriteString(symbol)
			}
		}
		section.WriteString("\n")
	}
	
	section.WriteString("\nCross-Section Legend:\n")
	section.WriteString("▓ Rock   ░ Soil   ≈ Groundwater   ● Ore deposits   ○ Air pockets\n")
	
	return section.String()
}

// Helper functions for topology visualization

func (m *CLIModel) getElevationSymbol(elevation float64) string {
	if elevation > 0.8 {
		return "▲" // High mountains
	} else if elevation > 0.6 {
		return "▲" // Medium mountains
	} else if elevation > 0.3 {
		return "△" // Hills
	} else if elevation > 0.0 {
		return "." // Plains
	} else if elevation > -0.3 {
		return "~" // Low areas
	} else {
		return "≈" // Deep water
	}
}

func (m *CLIModel) getElevationColor(elevation float64) lipgloss.Style {
	if elevation > 0.8 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White (snow)
	} else if elevation > 0.6 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Gray (rock)
	} else if elevation > 0.3 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("130")) // Brown (hills)
	} else if elevation > 0.0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // Green (plains)
	} else if elevation > -0.3 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue (shallow water)
	} else {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("21")) // Dark blue (deep water)
	}
}

func (m *CLIModel) getTerrainFeatureSymbol(featureType TerrainType) string {
	switch featureType {
	case 0: // TerrainFlat
		return "▫"
	case 1: // TerrainHill
		return "△"
	case 2: // TerrainMountain
		return "▲"
	case 3: // TerrainValley
		return "∪"
	case 4: // TerrainRiver
		return "≈"
	case 5: // TerrainLake
		return "○"
	case 6: // TerrainCanyon
		return "◢"
	case 7: // TerrainCrater
		return "◯"
	case 8: // TerrainVolcano
		return "🌋"
	case 9: // TerrainGlacier
		return "❄"
	default:
		return "?"
	}
}

func (m *CLIModel) getWaterBodySymbol(bodyType string) string {
	switch bodyType {
	case "river":
		return "≈"
	case "lake":
		return "○"
	case "stream":
		return "~"
	case "ocean":
		return "◯"
	default:
		return "💧"
	}
}

func (m *CLIModel) renderFeatureProfile(feature *TerrainFeature) string {
	// Create a mini elevation profile
	profile := ""
	height := int(feature.Height * 10)
	for i := 0; i < 10; i++ {
		if i < height {
			profile += "█"
		} else {
			profile += "░"
		}
	}
	return profile
}

func (m *CLIModel) renderWaterFlow(waterBody *WaterBody) string {
	// Simulate flow direction based on water body properties
	if waterBody.Flow > 0.8 {
		return "►►► (Strong current)"
	} else if waterBody.Flow > 0.4 {
		return "►► (Moderate flow)"
	} else {
		return "► (Gentle flow)"
	}
}

func (m *CLIModel) getUndergroundLayerSymbol(cell TopologyCell, layer int) string {
	// Simulate underground composition based on surface features and depth
	if cell.WaterLevel > 0.5 {
		return "≈" // Groundwater
	} else if cell.Elevation > 0.7 && layer <= 2 {
		return "▓" // Rock in mountains
	} else if layer == 1 {
		return "░" // Topsoil
	} else if layer <= 3 {
		return "▓" // Bedrock
	} else {
		if (int(cell.Elevation*100)+layer*7)%10 == 0 {
			return "●" // Ore deposits
		} else if (int(cell.Elevation*100)+layer*3)%15 == 0 {
			return "○" // Air pockets
		} else {
			return "▓" // Rock
		}
	}
}

func (m *CLIModel) reproductionView() string {
	var content strings.Builder
	
	content.WriteString("=== REPRODUCTION & DECAY SYSTEM ===\n\n")
	
	// Reproduction system stats
	if m.world.ReproductionSystem != nil {
		rs := m.world.ReproductionSystem
		content.WriteString(fmt.Sprintf("Active eggs: %d\n", len(rs.Eggs)))
		content.WriteString(fmt.Sprintf("Decaying items: %d\n", len(rs.DecayingItems)))
		content.WriteString(fmt.Sprintf("Next egg ID: %d\n", rs.NextEggID))
		content.WriteString(fmt.Sprintf("Next item ID: %d\n", rs.NextItemID))
		
		// Egg details
		if len(rs.Eggs) > 0 {
			content.WriteString("\n=== ACTIVE EGGS ===\n")
			for i, egg := range rs.Eggs {
				if i >= 5 { // Limit display
					content.WriteString(fmt.Sprintf("... and %d more eggs\n", len(rs.Eggs)-i))
					break
				}
				
				content.WriteString(fmt.Sprintf("Egg %d:\n", egg.ID))
				content.WriteString(fmt.Sprintf("  Parents: %d + %d\n", egg.Parent1ID, egg.Parent2ID))
				content.WriteString(fmt.Sprintf("  Species: %s\n", egg.Species))
				content.WriteString(fmt.Sprintf("  Age: %d/%d ticks\n", m.world.Tick-egg.LayingTick, egg.HatchingPeriod))
				content.WriteString(fmt.Sprintf("  Energy: %.1f\n", egg.Energy))
				content.WriteString(fmt.Sprintf("  Position: (%.1f, %.1f)\n", egg.Position.X, egg.Position.Y))
				content.WriteString("\n")
			}
		}
		
		// Decaying items
		if len(rs.DecayingItems) > 0 {
			content.WriteString("\n=== DECAYING ITEMS ===\n")
			for i, item := range rs.DecayingItems {
				if i >= 5 { // Limit display
					content.WriteString(fmt.Sprintf("... and %d more items\n", len(rs.DecayingItems)-i))
					break
				}
				
				content.WriteString(fmt.Sprintf("%s %d:\n", item.ItemType, item.ID))
				content.WriteString(fmt.Sprintf("  Origin: %s\n", item.OriginSpecies))
				content.WriteString(fmt.Sprintf("  Age: %d/%d ticks\n", m.world.Tick-item.CreationTick, item.DecayPeriod))
				content.WriteString(fmt.Sprintf("  Nutrients: %.1f\n", item.NutrientValue))
				content.WriteString(fmt.Sprintf("  Size: %.1f\n", item.Size))
				content.WriteString(fmt.Sprintf("  Position: (%.1f, %.1f)\n", item.Position.X, item.Position.Y))
				content.WriteString("\n")
			}
		}
	}
	
	// Entity reproduction status
	content.WriteString("\n=== ENTITY REPRODUCTION STATUS ===\n")
	
	// Count entities by reproduction mode and status
	modeCount := make(map[string]int)
	strategyCount := make(map[string]int)
	pregnantCount := 0
	readyToMateCount := 0
	matingSeasonCount := 0
	migratingCount := 0
	
	for _, entity := range m.world.AllEntities {
		if !entity.IsAlive || entity.ReproductionStatus == nil {
			continue
		}
		
		rs := entity.ReproductionStatus
		modeCount[rs.Mode.String()]++
		strategyCount[rs.Strategy.String()]++
		
		if rs.IsPregnant {
			pregnantCount++
		}
		if rs.ReadyToMate {
			readyToMateCount++
		}
		if rs.MatingSeason {
			matingSeasonCount++
		}
		if rs.RequiresMigration {
			migratingCount++
		}
	}
	
	content.WriteString("Reproduction Statistics:\n")
	content.WriteString(fmt.Sprintf("  Pregnant entities: %d\n", pregnantCount))
	content.WriteString(fmt.Sprintf("  Ready to mate: %d\n", readyToMateCount))
	content.WriteString(fmt.Sprintf("  In mating season: %d\n", matingSeasonCount))
	content.WriteString(fmt.Sprintf("  Require migration: %d\n", migratingCount))
	
	content.WriteString("\nReproduction Modes:\n")
	for mode, count := range modeCount {
		content.WriteString(fmt.Sprintf("  %s: %d\n", mode, count))
	}
	
	content.WriteString("\nMating Strategies:\n")
	for strategy, count := range strategyCount {
		content.WriteString(fmt.Sprintf("  %s: %d\n", strategy, count))
	}
	
	// Show some example entities with detailed reproduction info
	content.WriteString("\n=== SAMPLE ENTITIES ===\n")
	entityCount := 0
	for _, entity := range m.world.AllEntities {
		if !entity.IsAlive || entity.ReproductionStatus == nil {
			continue
		}
		
		if entityCount >= 3 { // Show only first 3
			break
		}
		
		rs := entity.ReproductionStatus
		content.WriteString(fmt.Sprintf("Entity %d (%s):\n", entity.ID, entity.Species))
		content.WriteString(fmt.Sprintf("  Mode: %s, Strategy: %s\n", rs.Mode.String(), rs.Strategy.String()))
		content.WriteString(fmt.Sprintf("  Age: %d, Energy: %.1f\n", entity.Age, entity.Energy))
		
		if rs.IsPregnant {
			gestation := m.world.Tick - rs.GestationStartTick
			content.WriteString(fmt.Sprintf("  PREGNANT (%.1f%% complete)\n", float64(gestation)/float64(rs.GestationPeriod)*100))
		}
		
		if rs.MateID > 0 {
			content.WriteString(fmt.Sprintf("  Mated to: Entity %d\n", rs.MateID))
		}
		
		if rs.RequiresMigration {
			dx := entity.Position.X - rs.PreferredMatingLocation.X
			dy := entity.Position.Y - rs.PreferredMatingLocation.Y
			distance := math.Sqrt(dx*dx + dy*dy)
			content.WriteString(fmt.Sprintf("  Migration distance: %.1f units\n", distance))
		}
		
		content.WriteString("\n")
		entityCount++
	}
	
	return content.String()
}

// statisticalView shows comprehensive statistical analysis of the simulation
func (m *CLIModel) statisticalView() string {
	if m.world.StatisticalReporter == nil {
		return "Statistical analysis not available"
	}

	var content strings.Builder
	content.WriteString("📊 STATISTICAL ANALYSIS\n\n")

	// Summary statistics
	summary := m.world.StatisticalReporter.GetSummaryStatistics()
	content.WriteString("SUMMARY STATISTICS:\n")
	content.WriteString(fmt.Sprintf("  Total Events: %d\n", summary["total_events"]))
	content.WriteString(fmt.Sprintf("  Total Snapshots: %d\n", summary["total_snapshots"]))
	content.WriteString(fmt.Sprintf("  Total Anomalies: %d\n", summary["total_anomalies"]))
	content.WriteString(fmt.Sprintf("  Latest Tick: %d\n", summary["latest_tick"]))
	content.WriteString(fmt.Sprintf("  Total Entities: %d\n", summary["total_entities"]))
	content.WriteString(fmt.Sprintf("  Total Plants: %d\n", summary["total_plants"]))
	content.WriteString(fmt.Sprintf("  Total Energy: %.2f\n", summary["total_energy"]))
	content.WriteString(fmt.Sprintf("  Species Count: %d\n", summary["species_count"]))
	
	if baseline, ok := summary["energy_baseline"].(float64); ok && baseline > 0 {
		currentEnergy := summary["total_energy"].(float64)
		change := ((currentEnergy - baseline) / baseline) * 100
		content.WriteString(fmt.Sprintf("  Energy Change: %.2f%% from baseline\n", change))
	}
	content.WriteString("\n")

	// Trends
	if trend, ok := summary["energy_trend"].(string); ok {
		content.WriteString(fmt.Sprintf("Energy Trend: %s\n", trend))
	}
	if trend, ok := summary["population_trend"].(string); ok {
		content.WriteString(fmt.Sprintf("Population Trend: %s\n", trend))
	}
	content.WriteString("\n")

	// Recent anomalies summary
	recentAnomalies := m.world.StatisticalReporter.GetRecentAnomalies(100, m.world.Tick)
	content.WriteString(fmt.Sprintf("RECENT ANOMALIES (%d):\n", len(recentAnomalies)))
	
	anomalyCounts := make(map[AnomalyType]int)
	for _, anomaly := range recentAnomalies {
		anomalyCounts[anomaly.Type]++
	}
	
	for anomalyType, count := range anomalyCounts {
		content.WriteString(fmt.Sprintf("  %s: %d\n", anomalyType, count))
	}
	content.WriteString("\n")

	// Latest snapshot details
	if len(m.world.StatisticalReporter.Snapshots) > 0 {
		latest := m.world.StatisticalReporter.Snapshots[len(m.world.StatisticalReporter.Snapshots)-1]
		content.WriteString("LATEST SNAPSHOT:\n")
		content.WriteString(fmt.Sprintf("  Tick: %d\n", latest.Tick))
		content.WriteString(fmt.Sprintf("  Entities: %d, Plants: %d\n", latest.TotalEntities, latest.TotalPlants))
		content.WriteString(fmt.Sprintf("  Total Energy: %.2f\n", latest.TotalEnergy))
		
		// Physics metrics
		content.WriteString(fmt.Sprintf("  Total Momentum: %.4f\n", latest.PhysicsMetrics.TotalMomentum))
		content.WriteString(fmt.Sprintf("  Kinetic Energy: %.4f\n", latest.PhysicsMetrics.TotalKineticEnergy))
		content.WriteString(fmt.Sprintf("  Avg Velocity: %.4f\n", latest.PhysicsMetrics.AverageVelocity))
		content.WriteString(fmt.Sprintf("  Collisions: %d\n", latest.PhysicsMetrics.CollisionCount))
		
		// Communication metrics
		content.WriteString(fmt.Sprintf("  Active Signals: %d\n", latest.CommunicationMetrics.ActiveSignals))
		content.WriteString(fmt.Sprintf("  Signal Efficiency: %.4f\n", latest.CommunicationMetrics.SignalEfficiency))
	}
	content.WriteString("\n")

	// Recent events
	recentEvents := m.world.StatisticalReporter.Events
	if len(recentEvents) > 10 {
		recentEvents = recentEvents[len(recentEvents)-10:] // Last 10 events
	}
	
	content.WriteString("RECENT EVENTS:\n")
	for _, event := range recentEvents {
		content.WriteString(fmt.Sprintf("  T%d: %s (%s)\n", event.Tick, event.EventType, event.Category))
		if event.Change != 0 {
			content.WriteString(fmt.Sprintf("        Change: %.4f\n", event.Change))
		}
	}

	content.WriteString("\nControls: [v] Next View [E] Export Data [R] Reset Analysis")

	return content.String()
}

// anomaliesView shows detected anomalies and statistical issues
func (m *CLIModel) anomaliesView() string {
	if m.world.StatisticalReporter == nil {
		return "Statistical analysis not available"
	}

	var content strings.Builder
	content.WriteString("⚠️  ANOMALY DETECTION\n\n")

	recentAnomalies := m.world.StatisticalReporter.GetRecentAnomalies(50, m.world.Tick)
	
	if len(recentAnomalies) == 0 {
		content.WriteString("✅ No anomalies detected!\n")
		content.WriteString("The simulation appears to be running within expected parameters.\n\n")
	} else {
		content.WriteString(fmt.Sprintf("Found %d anomalies:\n\n", len(recentAnomalies)))
		
		// Group anomalies by type
		anomaliesByType := make(map[AnomalyType][]Anomaly)
		for _, anomaly := range recentAnomalies {
			anomaliesByType[anomaly.Type] = append(anomaliesByType[anomaly.Type], anomaly)
		}
		
		// Display each type
		for anomalyType, anomalies := range anomaliesByType {
			content.WriteString(fmt.Sprintf("🔍 %s (%d occurrences):\n", anomalyType, len(anomalies)))
			
			// Show most recent and most severe
			var mostRecent, mostSevere Anomaly
			for i, anomaly := range anomalies {
				if i == 0 {
					mostRecent = anomaly
					mostSevere = anomaly
				} else {
					if anomaly.Tick > mostRecent.Tick {
						mostRecent = anomaly
					}
					if anomaly.Severity > mostSevere.Severity {
						mostSevere = anomaly
					}
				}
			}
			
			content.WriteString(fmt.Sprintf("  Most Recent (T%d): %s\n", mostRecent.Tick, mostRecent.Description))
			content.WriteString(fmt.Sprintf("    Severity: %.2f, Confidence: %.2f\n", mostRecent.Severity, mostRecent.Confidence))
			
			if mostSevere.Tick != mostRecent.Tick {
				content.WriteString(fmt.Sprintf("  Most Severe (T%d): %s\n", mostSevere.Tick, mostSevere.Description))
				content.WriteString(fmt.Sprintf("    Severity: %.2f, Confidence: %.2f\n", mostSevere.Severity, mostSevere.Confidence))
			}
			content.WriteString("\n")
		}
		
		// Recommendations based on anomaly types
		content.WriteString("RECOMMENDATIONS:\n")
		
		if _, hasEnergyIssues := anomaliesByType[AnomalyEnergyConservation]; hasEnergyIssues {
			content.WriteString("• Energy Conservation: Check entity/plant death and birth rates\n")
			content.WriteString("• Verify energy gain/loss calculations are balanced\n")
		}
		
		if _, hasDistIssues := anomaliesByType[AnomalyUnrealisticDistribution]; hasDistIssues {
			content.WriteString("• Distribution Issues: Check mutation algorithms for proper randomization\n")
			content.WriteString("• Verify trait bounds and initialization\n")
		}
		
		if _, hasPhysicsIssues := anomaliesByType[AnomalyPhysicsViolation]; hasPhysicsIssues {
			content.WriteString("• Physics Violations: Check momentum and energy conservation in physics engine\n")
			content.WriteString("• Verify collision and force calculations\n")
		}
		
		if _, hasBioIssues := anomaliesByType[AnomalyBiologicalImplausibility]; hasBioIssues {
			content.WriteString("• Biological Issues: Check trait evolution and bounds\n")
			content.WriteString("• Verify species-specific behaviors are realistic\n")
		}
		
		if _, hasPopIssues := anomaliesByType[AnomalyPopulationAnomaly]; hasPopIssues {
			content.WriteString("• Population Issues: Check carrying capacity and reproduction rates\n")
			content.WriteString("• Verify death conditions and environmental pressures\n")
		}
	}

	content.WriteString("\nControls: [v] Next View [C] Clear Anomalies [A] Auto-Fix")

	return content.String()
}

// warfareView renders the colony warfare and diplomacy information
func (m CLIModel) warfareView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("⚔️ Colony Warfare & Diplomacy") + "\n\n")

	if m.world.ColonyWarfareSystem == nil {
		content.WriteString("Colony warfare system not initialized\n")
		return content.String()
	}

	// Get warfare statistics
	stats := m.world.ColonyWarfareSystem.GetWarfareStats()
	
	// General Statistics
	content.WriteString("=== SYSTEM STATUS ===\n")
	content.WriteString(fmt.Sprintf("Total colonies: %d\n", stats["total_colonies"].(int)))
	content.WriteString(fmt.Sprintf("Active conflicts: %d\n", stats["active_conflicts"].(int)))
	content.WriteString(fmt.Sprintf("Active alliances: %d\n", stats["total_alliances"].(int)))
	content.WriteString(fmt.Sprintf("Trade agreements: %d\n", stats["active_trade_agreements"].(int)))
	
	// Diplomatic Relations
	content.WriteString("\n=== DIPLOMATIC RELATIONS ===\n")
	totalRelations := stats["total_relations"].(int)
	if totalRelations > 0 {
		neutralPct := float64(stats["neutral_relations"].(int)) / float64(totalRelations) * 100
		alliedPct := float64(stats["allied_relations"].(int)) / float64(totalRelations) * 100
		enemyPct := float64(stats["enemy_relations"].(int)) / float64(totalRelations) * 100
		trucePct := float64(stats["truce_relations"].(int)) / float64(totalRelations) * 100
		
		content.WriteString(fmt.Sprintf("Neutral: %d (%.1f%%)\n", stats["neutral_relations"].(int), neutralPct))
		content.WriteString(fmt.Sprintf("Allied: %d (%.1f%%)\n", stats["allied_relations"].(int), alliedPct))
		content.WriteString(fmt.Sprintf("Enemy: %d (%.1f%%)\n", stats["enemy_relations"].(int), enemyPct))
		content.WriteString(fmt.Sprintf("Truce: %d (%.1f%%)\n", stats["truce_relations"].(int), trucePct))
	} else {
		content.WriteString("No diplomatic relations established\n")
	}
	
	// Active Conflicts
	content.WriteString("\n=== ACTIVE CONFLICTS ===\n")
	conflicts := m.world.ColonyWarfareSystem.ActiveConflicts
	if len(conflicts) == 0 {
		content.WriteString("No active conflicts - Peace prevails!\n")
	} else {
		for i, conflict := range conflicts {
			if i >= 5 { // Show only first 5 conflicts
				content.WriteString(fmt.Sprintf("... and %d more conflicts\n", len(conflicts)-5))
				break
			}
			
			conflictTypeStr := "Unknown"
			switch conflict.ConflictType {
			case BorderSkirmish:
				conflictTypeStr = "Border Skirmish"
			case ResourceWar:
				conflictTypeStr = "Resource War"
			case TotalWar:
				conflictTypeStr = "Total War"
			case Raid:
				conflictTypeStr = "Raid"
			}
			
			content.WriteString(fmt.Sprintf("Conflict #%d: %s\n", conflict.ID, conflictTypeStr))
			content.WriteString(fmt.Sprintf("  Attacker: Colony %d vs Defender: Colony %d\n", 
				conflict.Attacker, conflict.Defender))
			content.WriteString(fmt.Sprintf("  Duration: %d ticks, Intensity: %.2f\n", 
				conflict.TurnsActive, conflict.Intensity))
			content.WriteString(fmt.Sprintf("  Casualties: %d, War Goal: %s\n", 
				conflict.CasualtyCount, conflict.WarGoal))
			
			if len(conflict.TerritoryClaimed) > 0 {
				content.WriteString(fmt.Sprintf("  Territory claimed: %d areas\n", len(conflict.TerritoryClaimed)))
			}
			content.WriteString("\n")
		}
	}
	
	// Active Trade Agreements
	content.WriteString("=== ACTIVE TRADE AGREEMENTS ===\n")
	tradeAgreements := m.world.ColonyWarfareSystem.TradeAgreements
	activeTradeCount := 0
	for _, agreement := range tradeAgreements {
		if agreement.IsActive {
			activeTradeCount++
		}
	}
	
	if activeTradeCount == 0 {
		content.WriteString("No active trade agreements\n")
	} else {
		displayCount := 0
		for _, agreement := range tradeAgreements {
			if !agreement.IsActive {
				continue
			}
			if displayCount >= 5 { // Show only first 5 agreements
				content.WriteString(fmt.Sprintf("... and %d more trade agreements\n", activeTradeCount-5))
				break
			}
			
			content.WriteString(fmt.Sprintf("Trade Agreement #%d:\n", agreement.ID))
			content.WriteString(fmt.Sprintf("  Colonies: %d ↔ %d\n", 
				agreement.Colony1ID, agreement.Colony2ID))
			content.WriteString(fmt.Sprintf("  Volume: %.1f units/trade\n", agreement.TradeVolume))
			
			// Show what's being traded
			if len(agreement.ResourcesOffered) > 0 {
				content.WriteString("  Offering: ")
				first := true
				for resource, amount := range agreement.ResourcesOffered {
					if !first {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("%.1f %s", amount, resource))
					first = false
				}
				content.WriteString("\n")
			}
			
			if len(agreement.ResourcesWanted) > 0 {
				content.WriteString("  Wanting: ")
				first := true
				for resource, amount := range agreement.ResourcesWanted {
					if !first {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("%.1f %s", amount, resource))
					first = false
				}
				content.WriteString("\n")
			}
			
			// Enhanced: Show trade security information
			content.WriteString(fmt.Sprintf("  Security Level: %.1f%%\n", agreement.SecurityLevel*100))
			if agreement.EscortStrength > 0 {
				content.WriteString(fmt.Sprintf("  Escort Protection: %.1f%%\n", agreement.EscortStrength*100))
			}
			if len(agreement.RouteThreats) > 0 {
				content.WriteString(fmt.Sprintf("  Active Threats: %d (", len(agreement.RouteThreats)))
				for i, threat := range agreement.RouteThreats {
					if i > 0 {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("%s %.0f%%", threat.ThreatType, threat.Severity*100))
				}
				content.WriteString(")\n")
			}
			
			content.WriteString("\n")
			displayCount++
		}
	}
	
	// Active Alliances
	content.WriteString("=== ACTIVE ALLIANCES ===\n")
	alliances := m.world.ColonyWarfareSystem.Alliances
	activeAllianceCount := 0
	for _, alliance := range alliances {
		if alliance.IsActive {
			activeAllianceCount++
		}
	}
	
	if activeAllianceCount == 0 {
		content.WriteString("No active alliances\n")
	} else {
		displayCount := 0
		for _, alliance := range alliances {
			if !alliance.IsActive {
				continue
			}
			if displayCount >= 3 { // Show only first 3 alliances
				content.WriteString(fmt.Sprintf("... and %d more alliances\n", activeAllianceCount-3))
				break
			}
			
			content.WriteString(fmt.Sprintf("Alliance #%d (%s):\n", alliance.ID, alliance.AllianceType))
			content.WriteString(fmt.Sprintf("  Members: "))
			for i, memberID := range alliance.Members {
				if i > 0 {
					content.WriteString(", ")
				}
				content.WriteString(fmt.Sprintf("Colony %d", memberID))
			}
			content.WriteString("\n")
			
			if alliance.ResourceShare > 0 {
				content.WriteString(fmt.Sprintf("  Resource sharing: %.1f%%\n", alliance.ResourceShare*100))
			}
			
			if alliance.SharedDefense {
				content.WriteString("  Shared defense: Active\n")
			}
			
			// Enhanced: Show coordination features
			content.WriteString(fmt.Sprintf("  Coordination Level: %.1f%%\n", alliance.CoordinationLevel*100))
			if alliance.TradeProtection {
				content.WriteString("  Trade Route Protection: Active\n")
			}
			if alliance.IntelligenceSharing {
				content.WriteString("  Intelligence Sharing: Active\n")
			}
			if alliance.SharedTechnology > 0 {
				content.WriteString(fmt.Sprintf("  Technology Sharing: %.1f%%\n", alliance.SharedTechnology*100))
			}
			if len(alliance.JointOperations) > 0 {
				activeOps := 0
				for _, op := range alliance.JointOperations {
					if op.IsActive {
						activeOps++
					}
				}
				if activeOps > 0 {
					content.WriteString(fmt.Sprintf("  Active Joint Operations: %d\n", activeOps))
				}
			}
			
			// Show alliance age
			age := m.world.Tick - alliance.StartTick
			content.WriteString(fmt.Sprintf("  Age: %d ticks\n", age))
			
			content.WriteString("\n")
			displayCount++
		}
	}
	
	// Colony Information
	content.WriteString("=== COLONY OVERVIEW ===\n")
	colonies := m.world.CasteSystem.Colonies
	if len(colonies) == 0 {
		content.WriteString("No colonies established\n")
	} else {
		for i, colony := range colonies {
			if i >= 8 { // Show only first 8 colonies
				content.WriteString(fmt.Sprintf("... and %d more colonies\n", len(colonies)-8))
				break
			}
			
			diplomacy := m.world.ColonyWarfareSystem.ColonyDiplomacies[colony.ID]
			
			content.WriteString(fmt.Sprintf("Colony %d:\n", colony.ID))
			content.WriteString(fmt.Sprintf("  Size: %d members, Age: %d ticks\n", 
				colony.ColonySize, colony.ColonyAge))
			content.WriteString(fmt.Sprintf("  Territory: %d areas, Fitness: %.2f\n", 
				len(colony.Territory), colony.ColonyFitness))
			
			// Show resource stockpiles
			if len(colony.Resources) > 0 {
				content.WriteString("  Resources: ")
				first := true
				for resource, amount := range colony.Resources {
					if !first {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("%.1f %s", amount, resource))
					first = false
				}
				content.WriteString("\n")
			}
			
			if diplomacy != nil {
				content.WriteString(fmt.Sprintf("  Reputation: %.2f\n", diplomacy.Reputation))
				
				// Count relations
				allies := 0
				enemies := 0
				trading := 0
				for _, relation := range diplomacy.Relations {
					switch relation {
					case Allied:
						allies++
					case Enemy:
						enemies++
					case Trading:
						trading++
					}
				}
				
				if allies > 0 || enemies > 0 || trading > 0 {
					content.WriteString(fmt.Sprintf("  Allies: %d, Enemies: %d, Trading: %d\n", allies, enemies, trading))
				}
				
				// Show active conflicts for this colony
				activeConflictsForColony := 0
				for _, conflict := range conflicts {
					if conflict.Attacker == colony.ID || conflict.Defender == colony.ID {
						activeConflictsForColony++
					}
				}
				if activeConflictsForColony > 0 {
					content.WriteString(fmt.Sprintf("  Active conflicts: %d\n", activeConflictsForColony))
				}
				
				// Show trade agreements for this colony
				tradeCount := 0
				for _, agreement := range tradeAgreements {
					if agreement.IsActive && (agreement.Colony1ID == colony.ID || agreement.Colony2ID == colony.ID) {
						tradeCount++
					}
				}
				if tradeCount > 0 {
					content.WriteString(fmt.Sprintf("  Trade agreements: %d\n", tradeCount))
				}
			}
			
			content.WriteString("\n")
		}
	}
	
	// System Configuration
	content.WriteString("=== SYSTEM SETTINGS ===\n")
	content.WriteString(fmt.Sprintf("Border conflict chance: %.1f%%\n", stats["border_conflicts"].(float64)*100))
	content.WriteString(fmt.Sprintf("Resource competition: %.1f%%\n", stats["resource_competition"].(float64)*100))
	content.WriteString(fmt.Sprintf("Max simultaneous conflicts: %d\n", m.world.ColonyWarfareSystem.MaxActiveConflicts))
	
	content.WriteString("\nControls: [v] Next View")
	
	return content.String()
}

// exportStatisticalData exports statistical data to files
func (m *CLIModel) exportStatisticalData() {
	if m.world.StatisticalReporter == nil {
		return
	}

	// Export to CSV
	csvFilename := fmt.Sprintf("evosim_stats_%d.csv", m.world.Tick)
	if err := m.world.StatisticalReporter.ExportToCSV(csvFilename); err == nil {
		// In a real CLI app, we'd show a notification
		// For now, this happens silently
	}

	// Export to JSON
	jsonFilename := fmt.Sprintf("evosim_analysis_%d.json", m.world.Tick)
	if err := m.world.StatisticalReporter.ExportToJSON(jsonFilename); err == nil {
		// In a real CLI app, we'd show a notification
		// For now, this happens silently
	}
}

// toolsView renders the tool system information
func (m CLIModel) toolsView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("🔧 Tool System") + "\n\n")

	if m.world.ToolSystem == nil {
		content.WriteString("Tool system not initialized\n")
		return content.String()
	}

	// Get tool statistics from the system
	stats := m.world.ToolSystem.GetToolStats()
	
	// Basic tool statistics
	content.WriteString("=== TOOL STATISTICS ===\n")
	if totalTools, ok := stats["total_tools"].(int); ok {
		content.WriteString(fmt.Sprintf("Total Tools: %d\n", totalTools))
	}
	if ownedTools, ok := stats["owned_tools"].(int); ok {
		content.WriteString(fmt.Sprintf("Owned Tools: %d\n", ownedTools))
	}
	if droppedTools, ok := stats["dropped_tools"].(int); ok {
		content.WriteString(fmt.Sprintf("Dropped Tools: %d\n", droppedTools))
	}
	if avgDurability, ok := stats["avg_durability"].(float64); ok {
		content.WriteString(fmt.Sprintf("Average Durability: %.2f\n", avgDurability))
	}
	if avgEfficiency, ok := stats["avg_efficiency"].(float64); ok {
		content.WriteString(fmt.Sprintf("Average Efficiency: %.2f\n", avgEfficiency))
	}

	// Tool types breakdown
	content.WriteString("\n=== TOOL TYPES ===\n")
	if toolTypes, ok := stats["tool_types"].(map[string]int); ok && len(toolTypes) > 0 {
		for toolType, count := range toolTypes {
			content.WriteString(fmt.Sprintf("%s: %d\n", toolType, count))
		}
	} else {
		content.WriteString("No tools created yet\n")
	}

	// Tool usage analysis
	content.WriteString("\n=== TOOL USAGE ANALYSIS ===\n")
	if ownedTools, ok := stats["owned_tools"].(int); ok {
		if ownedTools == 0 {
			content.WriteString("Usage Level: No tool use\n")
		} else if ownedTools < 5 {
			content.WriteString("Usage Level: Basic tool use\n")
		} else {
			content.WriteString("Usage Level: Advanced tool use\n")
		}
	}

	// Tool distribution by entities
	content.WriteString("\n=== TOOL DISTRIBUTION ===\n")
	entityToolCount := make(map[int]int) // entity ID -> tool count
	toolsWithOwners := 0
	
	// Count tools per entity
	for _, tool := range m.world.ToolSystem.Tools {
		if tool.Owner != nil {
			entityToolCount[tool.Owner.ID]++
			toolsWithOwners++
		}
	}
	
	content.WriteString(fmt.Sprintf("Entities with tools: %d\n", len(entityToolCount)))
	content.WriteString(fmt.Sprintf("Tools with owners: %d\n", toolsWithOwners))
	
	if len(entityToolCount) > 0 {
		// Find min, max, average tools per entity
		minTools, maxTools, totalTools := entityToolCount[0], 0, 0
		for _, count := range entityToolCount {
			if count < minTools {
				minTools = count
			}
			if count > maxTools {
				maxTools = count
			}
			totalTools += count
		}
		avgTools := float64(totalTools) / float64(len(entityToolCount))
		
		content.WriteString(fmt.Sprintf("Tools per entity: min=%d, max=%d, avg=%.1f\n", minTools, maxTools, avgTools))
	}

	return content.String()
}

// environmentView renders environmental modification information
func (m CLIModel) environmentView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("🏗️ Environmental Modification") + "\n\n")

	if m.world.EnvironmentalModSystem == nil {
		content.WriteString("Environmental modification system not initialized\n")
		return content.String()
	}

	// Get modification statistics
	totalMods := len(m.world.EnvironmentalModSystem.Modifications)
	activeMods := 0
	inactiveMods := 0
	totalDurability := 0.0
	tunnelNetworks := 0
	modificationTypes := make(map[string]int)

	for _, mod := range m.world.EnvironmentalModSystem.Modifications {
		if mod.IsActive {
			activeMods++
		} else {
			inactiveMods++
		}
		totalDurability += mod.Durability

		// Count modification types
		modTypeName := m.getModificationTypeName(int(mod.Type))
		modificationTypes[modTypeName]++

		// Count tunnel networks (simplified)
		if mod.Type == EnvModTunnel { // Tunnel
			tunnelNetworks++
		}
	}

	avgDurability := 0.0
	if totalMods > 0 {
		avgDurability = totalDurability / float64(totalMods)
	}

	// Basic modification statistics
	content.WriteString("=== MODIFICATION STATISTICS ===\n")
	content.WriteString(fmt.Sprintf("Total Modifications: %d\n", totalMods))
	content.WriteString(fmt.Sprintf("Active Modifications: %d\n", activeMods))
	content.WriteString(fmt.Sprintf("Inactive Modifications: %d\n", inactiveMods))
	content.WriteString(fmt.Sprintf("Average Durability: %.2f\n", avgDurability))
	content.WriteString(fmt.Sprintf("Tunnel Networks: %d\n", tunnelNetworks))

	// Modification types breakdown
	content.WriteString("\n=== MODIFICATION TYPES ===\n")
	if len(modificationTypes) > 0 {
		for modType, count := range modificationTypes {
			content.WriteString(fmt.Sprintf("%s: %d\n", modType, count))
		}
	} else {
		content.WriteString("No environmental modifications yet\n")
	}

	// Activity level analysis
	content.WriteString("\n=== MODIFICATION ACTIVITY ===\n")
	if totalMods == 0 {
		content.WriteString("Activity Level: No modifications\n")
	} else if activeMods < 5 {
		content.WriteString("Activity Level: Basic environmental shaping\n")
	} else {
		content.WriteString("Activity Level: Advanced environmental engineering\n")
	}

	// Show some recent modifications
	content.WriteString("\n=== RECENT MODIFICATIONS ===\n")
	modCount := 0
	for _, mod := range m.world.EnvironmentalModSystem.Modifications {
		if modCount >= 5 {
			remaining := len(m.world.EnvironmentalModSystem.Modifications) - modCount
			content.WriteString(fmt.Sprintf("... and %d more modifications\n", remaining))
			break
		}

		modTypeName := m.getModificationTypeName(int(mod.Type))
		status := "ACTIVE"
		if !mod.IsActive {
			status = "INACTIVE"
		}

		content.WriteString(fmt.Sprintf("%s %s:\n", modTypeName, status))
		content.WriteString(fmt.Sprintf("  Position: (%.1f, %.1f)\n", mod.Position.X, mod.Position.Y))
		content.WriteString(fmt.Sprintf("  Durability: %.2f\n", mod.Durability))
		if mod.Creator != nil {
			content.WriteString(fmt.Sprintf("  Creator: Entity %d\n", mod.Creator.ID))
		}
		content.WriteString("\n")
		modCount++
	}

	return content.String()
}

// behaviorView renders emergent behavior information
func (m CLIModel) behaviorView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("🧠 Emergent Behavior") + "\n\n")

	if m.world.EmergentBehaviorSystem == nil {
		content.WriteString("Emergent behavior system not initialized\n")
		return content.String()
	}

	// Get behavior statistics
	stats := m.world.EmergentBehaviorSystem.GetBehaviorStats()

	// Basic behavior statistics
	content.WriteString("=== BEHAVIOR STATISTICS ===\n")
	if totalEntities, ok := stats["total_entities"].(int); ok {
		content.WriteString(fmt.Sprintf("Total Entities: %d\n", totalEntities))
	}
	if discoveredBehaviors, ok := stats["discovered_behaviors"].(int); ok {
		content.WriteString(fmt.Sprintf("Discovered Behaviors: %d\n", discoveredBehaviors))
	}

	// Behavior spread
	content.WriteString("\n=== BEHAVIOR SPREAD ===\n")
	if behaviorSpread, ok := stats["behavior_spread"].(map[string]int); ok && len(behaviorSpread) > 0 {
		for behaviorName, count := range behaviorSpread {
			content.WriteString(fmt.Sprintf("%s: %d entities\n", behaviorName, count))
		}
	} else {
		content.WriteString("No behaviors have spread yet\n")
	}

	// Average proficiency
	content.WriteString("\n=== AVERAGE PROFICIENCY ===\n")
	if avgProficiency, ok := stats["avg_proficiency"].(map[string]float64); ok && len(avgProficiency) > 0 {
		for behaviorName, proficiency := range avgProficiency {
			content.WriteString(fmt.Sprintf("%s: %.2f\n", behaviorName, proficiency))
		}
	} else {
		content.WriteString("No proficiency data available yet\n")
	}

	// Behavior development analysis
	content.WriteString("\n=== BEHAVIOR DEVELOPMENT ===\n")
	if discoveredBehaviors, ok := stats["discovered_behaviors"].(int); ok {
		if discoveredBehaviors == 0 {
			content.WriteString("Development Level: No emergent behaviors discovered yet\n")
		} else if discoveredBehaviors < 3 {
			content.WriteString("Development Level: Early behavior emergence\n")
		} else if discoveredBehaviors < 8 {
			content.WriteString("Development Level: Moderate behavior complexity\n")
		} else {
			content.WriteString("Development Level: Advanced behavioral evolution\n")
		}
	}

	// Show behavior trends
	content.WriteString("\n=== BEHAVIOR TRENDS ===\n")
	if behaviorSpread, ok := stats["behavior_spread"].(map[string]int); ok {
		totalBehaviorInstances := 0
		for _, count := range behaviorSpread {
			totalBehaviorInstances += count
		}
		
		if totalBehaviorInstances > 0 {
			content.WriteString("Most Common Behaviors:\n")
			// Sort behaviors by prevalence
			type behaviorCount struct {
				name  string
				count int
			}
			var behaviors []behaviorCount
			for name, count := range behaviorSpread {
				behaviors = append(behaviors, behaviorCount{name, count})
			}
			
			// Simple sorting by count (descending)
			for i := 0; i < len(behaviors)-1; i++ {
				for j := i + 1; j < len(behaviors); j++ {
					if behaviors[j].count > behaviors[i].count {
						behaviors[i], behaviors[j] = behaviors[j], behaviors[i]
					}
				}
			}
			
			for i, behavior := range behaviors {
				if i >= 5 { // Show top 5
					break
				}
				percentage := float64(behavior.count) * 100.0 / float64(totalBehaviorInstances)
				content.WriteString(fmt.Sprintf("  %s: %.1f%% of behaviors\n", behavior.name, percentage))
			}
		}
	}

	// Behavior innovation rate
	if totalEntities, ok := stats["total_entities"].(int); ok {
		if discoveredBehaviors, ok := stats["discovered_behaviors"].(int); ok {
			if totalEntities > 0 {
				innovationRate := float64(discoveredBehaviors) / float64(totalEntities) * 100
				content.WriteString(fmt.Sprintf("\nInnovation Rate: %.2f%% (behaviors per entity)\n", innovationRate))
			}
		}
	}

	return content.String()
}

// getModificationTypeName returns the name of a modification type
func (m CLIModel) getModificationTypeName(modType int) string {
	modNames := map[int]string{
		int(EnvModTunnel):      "Tunnel",
		int(EnvModBurrow):      "Burrow",
		int(EnvModCache):       "Cache",
		int(EnvModTrap):        "Trap",
		int(EnvModWaterhole):   "Waterhole",
		int(EnvModPath):        "Path",
		int(EnvModMarking):     "Marking",
		int(EnvModNest):        "Nest",
		int(EnvModBridge):      "Bridge",
		int(EnvModBarrier):     "Barrier",
		int(EnvModTerrace):     "Terrace",
		int(EnvModDam):         "Dam",
	}
	
	if name, exists := modNames[modType]; exists {
		return name
	}
	return fmt.Sprintf("Type%d", modType)
}

// RunCLI starts the CLI interface
func RunCLI(world *World) error {
	model := NewCLIModel(world)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
