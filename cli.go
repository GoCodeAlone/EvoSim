package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
}

// tickMsg represents an auto-advance tick
type tickMsg time.Time

// Key bindings
var keys = struct {
	up    key.Binding
	down  key.Binding
	enter key.Binding
	space key.Binding
	help  key.Binding
	quit  key.Binding
	view  key.Binding
	auto  key.Binding
}{
	up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
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

	return CLIModel{
		world:          world,
		viewModes:      []string{"grid", "stats", "events", "populations"},
		selectedView:   "grid",
		autoAdvance:    true,
		lastUpdateTime: time.Now(),
		speciesColors:  speciesColors,
		speciesSymbols: speciesSymbols,
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

	// Count active events
	activeEvents := len(m.world.Events)

	title := titleStyle.Render(fmt.Sprintf("🌍 Genetic Ecosystem - Tick %d", m.world.Tick))
	info := infoStyle.Render(fmt.Sprintf("%s | %s | Entities: %d | Pops: %d | Events: %d | View: %s",
		status, worldTime, entities, populations, activeEvents, strings.ToUpper(m.selectedView)))

	return lipgloss.JoinHorizontal(lipgloss.Left, title, " ", info)
}

// gridView renders the animated world grid
func (m CLIModel) gridView() string {
	if m.world.Config.GridWidth == 0 || m.world.Config.GridHeight == 0 {
		return "Grid not initialized"
	}

	var gridBuilder strings.Builder

	// Build grid representation
	for y := 0; y < m.world.Config.GridHeight; y++ {
		for x := 0; x < m.world.Config.GridWidth; x++ {
			cell := m.world.Grid[y][x]
			biome := m.world.Biomes[cell.Biome]

			// Start with biome symbol
			symbol := biome.Symbol
			style := biomeColors[cell.Biome]

			// If entities present, show the dominant species
			if len(cell.Entities) > 0 {
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
					if style, exists := speciesStyles[dominantSpecies]; exists {
						style = style
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

			gridBuilder.WriteString(style.Render(string(symbol)))
		}
		if y < m.world.Config.GridHeight-1 {
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
	content.WriteString(fmt.Sprintf("World Time: %s\n", m.world.Clock.Format("15:04 Day 2006-01-02")))
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

// eventsView renders active world events
func (m CLIModel) eventsView() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("World Events") + "\n\n")

	if len(m.world.Events) == 0 {
		content.WriteString("No active events")
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

	content.WriteString("\n\nPossible Events:")
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

// footerView renders the footer with controls
func (m CLIModel) footerView() string {
	controls := []string{
		"space: pause/resume",
		"v: cycle view",
		"enter: step",
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
  v          Cycle through views (grid/stats/events/populations)
  a          Toggle auto-advance
  ?          Toggle this help screen
  q          Quit

VIEWS:
  Grid       Real-time animated world map with entities and biomes
  Stats      Detailed world and population statistics
  Events     Active world events and their effects
  Populations Detailed view of each species population

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

The simulation runs in real-time, with entities moving, aging, interacting,
and evolving based on their traits and the biome they're in. Different
biomes provide different challenges and benefits to different species.

World events can occur randomly, affecting mutation rates, energy drain,
and even changing the landscape itself.

Press ? again to return to the simulation.
`
	return help
}

// RunCLI starts the CLI interface
func RunCLI(world *World) error {
	model := NewCLIModel(world)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
