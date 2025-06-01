package main

import (
	"fmt"
	"math"
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
		viewModes:      []string{"grid", "stats", "events", "populations", "communication", "civilization", "physics", "wind", "species", "network"},
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
	infoText := fmt.Sprintf("%s | %s %s | Entities: %d | Species: %d | Events: %d | View: %s",
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
						if sym, exists := m.speciesSymbols[dominantSpecies]; exists {
							symbol = sym
						}
						if entityStyle, exists := speciesStyles[dominantSpecies]; exists {
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
	for biomeType, biome := range m.world.Biomes {
		style := biomeColors[biomeType]
		legend.WriteString(fmt.Sprintf("%s %s\n",
			style.Render(string(biome.Symbol)), biome.Name))
	}

	legend.WriteString("\n👥 Species:\n")
	for species, symbol := range m.speciesSymbols {
		style := speciesStyles[species]
		legend.WriteString(fmt.Sprintf("%s %s\n",
			style.Render(string(symbol)), strings.Title(species)))
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

	// Biome distribution
	content.WriteString("\n\nBiome Distribution:\n")
	biomeCount := make(map[BiomeType]int)
	for y := 0; y < m.world.Config.GridHeight; y++ {
		for x := 0; x < m.world.Config.GridWidth; x++ {
			biomeCount[m.world.Grid[y][x].Biome]++
		}
	}

	total := m.world.Config.GridWidth * m.world.Config.GridHeight
	for biomeType, count := range biomeCount {
		biome := m.world.Biomes[biomeType]
		percentage := float64(count) * 100.0 / float64(total)
		content.WriteString(fmt.Sprintf("  %s: %d cells (%.1f%%)\n",
			biome.Name, count, percentage))
	}

	return content.String()
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

	for species, pop := range m.world.Populations {
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

	return content.String()
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

// RunCLI starts the CLI interface
func RunCLI(world *World) error {
	model := NewCLIModel(world)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
