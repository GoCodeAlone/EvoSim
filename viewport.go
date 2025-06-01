package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// ZoomLevel represents different levels of detail
type ZoomLevel int

const (
	ZoomWorld  ZoomLevel = iota // Full world view
	ZoomRegion                  // Regional view (4x4 cells)
	ZoomLocal                   // Local view (single cell)
	ZoomEntity                  // Entity detail view
)

// ViewportSystem manages zoom, pan, and multi-level detail viewing
type ViewportSystem struct {
	ZoomLevel      ZoomLevel
	CenterX        float64 // World coordinates of viewport center
	CenterY        float64
	ViewWidth      int     // Terminal width available for view
	ViewHeight     int     // Terminal height available for view
	WorldWidth     float64 // Total world width
	WorldHeight    float64 // Total world height
	SelectedEntity *Entity // Entity selected for detailed view
	ShowDetails    bool    // Whether to show detailed information
	DetailLevel    int     // 0=basic, 1=moderate, 2=full details
}

// NewViewportSystem creates a new viewport system
func NewViewportSystem(worldWidth, worldHeight float64) *ViewportSystem {
	return &ViewportSystem{
		ZoomLevel:   ZoomWorld,
		CenterX:     worldWidth / 2,
		CenterY:     worldHeight / 2,
		WorldWidth:  worldWidth,
		WorldHeight: worldHeight,
		DetailLevel: 1,
	}
}

// SetViewSize updates the viewport size
func (vs *ViewportSystem) SetViewSize(width, height int) {
	vs.ViewWidth = width
	vs.ViewHeight = height
}

// ZoomIn increases zoom level
func (vs *ViewportSystem) ZoomIn() {
	if vs.ZoomLevel < ZoomEntity {
		vs.ZoomLevel++
	}
}

// ZoomOut decreases zoom level
func (vs *ViewportSystem) ZoomOut() {
	if vs.ZoomLevel > ZoomWorld {
		vs.ZoomLevel--
		if vs.ZoomLevel != ZoomEntity {
			vs.SelectedEntity = nil
		}
	}
}

// Pan moves the viewport center
func (vs *ViewportSystem) Pan(deltaX, deltaY float64) {
	// Scale pan speed based on zoom level
	panSpeed := vs.getPanSpeed()

	vs.CenterX += deltaX * panSpeed
	vs.CenterY += deltaY * panSpeed

	// Clamp to world bounds
	vs.CenterX = math.Max(0, math.Min(vs.WorldWidth, vs.CenterX))
	vs.CenterY = math.Max(0, math.Min(vs.WorldHeight, vs.CenterY))
}

// getPanSpeed returns appropriate pan speed for current zoom level
func (vs *ViewportSystem) getPanSpeed() float64 {
	switch vs.ZoomLevel {
	case ZoomWorld:
		return 5.0
	case ZoomRegion:
		return 2.0
	case ZoomLocal:
		return 0.5
	case ZoomEntity:
		return 0.1
	default:
		return 1.0
	}
}

// GetVisibleBounds returns the world coordinates visible in current viewport
func (vs *ViewportSystem) GetVisibleBounds() (minX, minY, maxX, maxY float64) {
	scale := vs.getScale()

	halfWidth := float64(vs.ViewWidth) * scale / 2
	halfHeight := float64(vs.ViewHeight) * scale / 2

	minX = vs.CenterX - halfWidth
	maxX = vs.CenterX + halfWidth
	minY = vs.CenterY - halfHeight
	maxY = vs.CenterY + halfHeight

	return
}

// getScale returns world units per terminal character
func (vs *ViewportSystem) getScale() float64 {
	switch vs.ZoomLevel {
	case ZoomWorld:
		return math.Max(vs.WorldWidth/float64(vs.ViewWidth), vs.WorldHeight/float64(vs.ViewHeight))
	case ZoomRegion:
		return vs.getScale() / 4 // 4x zoom
	case ZoomLocal:
		return vs.getScale() / 16 // 16x zoom from world
	case ZoomEntity:
		return 0.1 // Very close view
	default:
		return 1.0
	}
}

// SelectEntityAt attempts to select an entity at screen coordinates
func (vs *ViewportSystem) SelectEntityAt(screenX, screenY int, entities []*Entity) *Entity {
	minX, minY, _, _ := vs.GetVisibleBounds()
	scale := vs.getScale()

	// Convert screen coordinates to world coordinates
	worldX := minX + float64(screenX)*scale
	worldY := minY + float64(screenY)*scale

	// Find closest entity
	var closest *Entity
	minDistance := math.Inf(1)

	for _, entity := range entities {
		if !entity.IsAlive {
			continue
		}

		distance := math.Sqrt(math.Pow(entity.Position.X-worldX, 2) + math.Pow(entity.Position.Y-worldY, 2))
		if distance < minDistance && distance < scale*2 { // Within 2 character radius
			minDistance = distance
			closest = entity
		}
	}

	vs.SelectedEntity = closest
	if closest != nil {
		vs.ZoomLevel = ZoomEntity
		vs.CenterX = closest.Position.X
		vs.CenterY = closest.Position.Y
	}

	return closest
}

// RenderViewport renders the world at current zoom level
func (vs *ViewportSystem) RenderViewport(world *World) string {
	switch vs.ZoomLevel {
	case ZoomWorld:
		return vs.renderWorldView(world)
	case ZoomRegion:
		return vs.renderRegionView(world)
	case ZoomLocal:
		return vs.renderLocalView(world)
	case ZoomEntity:
		return vs.renderEntityView(world)
	default:
		return "Invalid zoom level"
	}
}

// renderWorldView renders the full world overview
func (vs *ViewportSystem) renderWorldView(world *World) string {
	var builder strings.Builder

	// Use existing grid view but with navigation indicators
	for y := 0; y < world.Config.GridHeight; y++ {
		for x := 0; x < world.Config.GridWidth; x++ {
			cell := world.Grid[y][x]
			symbol := vs.getCellSymbol(cell, world, ZoomWorld)
			style := vs.getCellStyle(cell, world, ZoomWorld)

			// Highlight viewport center
			worldX := (float64(x) + 0.5) * (world.Config.Width / float64(world.Config.GridWidth))
			worldY := (float64(y) + 0.5) * (world.Config.Height / float64(world.Config.GridHeight))

			if math.Abs(worldX-vs.CenterX) < 2 && math.Abs(worldY-vs.CenterY) < 2 {
				style = style.Copy().Background(lipgloss.Color("240"))
			}

			builder.WriteString(style.Render(string(symbol)))
		}
		if y < world.Config.GridHeight-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// renderRegionView renders a 4x4 cell region with more detail
func (vs *ViewportSystem) renderRegionView(world *World) string {
	var builder strings.Builder

	minX, minY, maxX, maxY := vs.GetVisibleBounds()

	// Calculate grid cell range
	startGridX := int(minX / world.Config.Width * float64(world.Config.GridWidth))
	endGridX := int(maxX / world.Config.Width * float64(world.Config.GridWidth))
	startGridY := int(minY / world.Config.Height * float64(world.Config.GridHeight))
	endGridY := int(maxY / world.Config.Height * float64(world.Config.GridHeight))

	// Clamp to valid range
	startGridX = int(math.Max(0, float64(startGridX)))
	endGridX = int(math.Min(float64(world.Config.GridWidth-1), float64(endGridX)))
	startGridY = int(math.Max(0, float64(startGridY)))
	endGridY = int(math.Min(float64(world.Config.GridHeight-1), float64(endGridY)))

	for y := startGridY; y <= endGridY; y++ {
		for x := startGridX; x <= endGridX; x++ {
			cell := world.Grid[y][x]

			// Render each cell as a 2x2 block for more detail
			symbols := vs.getDetailedCellSymbols(cell, world)
			styles := vs.getDetailedCellStyles(cell, world)

			for i, symbol := range symbols {
				builder.WriteString(styles[i].Render(string(symbol)))
			}

			if x < endGridX {
				builder.WriteString(" ")
			}
		}
		if y < endGridY {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// renderLocalView renders a single cell with maximum detail
func (vs *ViewportSystem) renderLocalView(world *World) string {
	// Find the cell containing the viewport center
	gridX := int(vs.CenterX / world.Config.Width * float64(world.Config.GridWidth))
	gridY := int(vs.CenterY / world.Config.Height * float64(world.Config.GridHeight))

	// Clamp to valid range
	gridX = int(math.Max(0, math.Min(float64(world.Config.GridWidth-1), float64(gridX))))
	gridY = int(math.Max(0, math.Min(float64(world.Config.GridHeight-1), float64(gridY))))

	cell := world.Grid[gridY][gridX]

	var content strings.Builder
	content.WriteString(fmt.Sprintf("=== LOCAL VIEW: Cell (%d, %d) ===\n", gridX, gridY))
	// Biome information
	biome := world.Biomes[cell.Biome]
	content.WriteString(fmt.Sprintf("Biome: %s %s\n", string(biome.Symbol), biome.Name))
	content.WriteString(fmt.Sprintf("Energy Drain: %.1f, Mutation Rate: %.3f\n", biome.EnergyDrain, biome.MutationRate))

	// Entities in cell
	if len(cell.Entities) > 0 {
		content.WriteString(fmt.Sprintf("\nEntities (%d):\n", len(cell.Entities)))
		for i, entity := range cell.Entities {
			if i >= 5 {
				content.WriteString(fmt.Sprintf("... and %d more\n", len(cell.Entities)-5))
				break
			}
			content.WriteString(fmt.Sprintf("  %s #%d (E:%.1f, Age:%d)\n",
				entity.Species, entity.ID, entity.Energy, entity.Age))
		}
	}

	// Plants in cell
	if len(cell.Plants) > 0 {
		content.WriteString(fmt.Sprintf("\nPlants (%d):\n", len(cell.Plants)))
		plantCounts := make(map[PlantType]int)
		for _, plant := range cell.Plants {
			if plant.IsAlive {
				plantCounts[plant.Type]++
			}
		}

		for plantType, count := range plantCounts {
			config := GetPlantConfigs()[plantType]
			content.WriteString(fmt.Sprintf("  %s %s: %d\n", string(config.Symbol), config.Name, count))
		}
	}

	// Environmental effects
	if cell.Event != nil {
		content.WriteString(fmt.Sprintf("\nActive Event: %s\n", cell.Event.Name))
		content.WriteString(fmt.Sprintf("Effect: %s\n", cell.Event.Description))
	}

	return content.String()
}

// renderEntityView renders detailed information about selected entity
func (vs *ViewportSystem) renderEntityView(world *World) string {
	if vs.SelectedEntity == nil || !vs.SelectedEntity.IsAlive {
		return "No entity selected"
	}

	entity := vs.SelectedEntity
	var content strings.Builder

	content.WriteString(fmt.Sprintf("=== ENTITY DETAILS: %s #%d ===\n", entity.Species, entity.ID))
	content.WriteString(fmt.Sprintf("Position: (%.1f, %.1f)\n", entity.Position.X, entity.Position.Y))
	content.WriteString(fmt.Sprintf("Energy: %.1f, Age: %d, Generation: %d\n", entity.Energy, entity.Age, entity.Generation))
	content.WriteString(fmt.Sprintf("Fitness: %.3f\n", entity.Fitness))

	// Traits
	content.WriteString("\nTraits:\n")
	for name, trait := range entity.Traits {
		bar := vs.renderProgressBar(trait.Value, -2.0, 2.0, 20)
		content.WriteString(fmt.Sprintf("  %-12s: %s %.3f\n", name, bar, trait.Value))
	}

	// Nearby entities
	nearbyEntities := vs.findNearbyEntities(entity, world.AllEntities, 10.0)
	if len(nearbyEntities) > 0 {
		content.WriteString(fmt.Sprintf("\nNearby Entities (%d):\n", len(nearbyEntities)))
		for i, nearby := range nearbyEntities {
			if i >= 3 {
				content.WriteString(fmt.Sprintf("... and %d more\n", len(nearbyEntities)-3))
				break
			}
			distance := entity.DistanceTo(nearby)
			content.WriteString(fmt.Sprintf("  %s #%d (%.1f units away)\n",
				nearby.Species, nearby.ID, distance))
		}
	}

	// Location context
	gridX := int(entity.Position.X / world.Config.Width * float64(world.Config.GridWidth))
	gridY := int(entity.Position.Y / world.Config.Height * float64(world.Config.GridHeight))
	gridX = int(math.Max(0, math.Min(float64(world.Config.GridWidth-1), float64(gridX))))
	gridY = int(math.Max(0, math.Min(float64(world.Config.GridHeight-1), float64(gridY))))

	cell := world.Grid[gridY][gridX]
	biome := world.Biomes[cell.Biome]
	content.WriteString(fmt.Sprintf("\nCurrent Biome: %s (%s)\n", biome.Name, string(biome.Symbol)))

	// Relationship potential
	compatibleEntities := 0
	for _, other := range nearbyEntities {
		if entity.CanMerge(other) {
			compatibleEntities++
		}
	}
	if compatibleEntities > 0 {
		content.WriteString(fmt.Sprintf("Compatible entities nearby: %d\n", compatibleEntities))
	}

	return content.String()
}

// findNearbyEntities finds entities within a certain range
func (vs *ViewportSystem) findNearbyEntities(center *Entity, allEntities []*Entity, maxDistance float64) []*Entity {
	var nearby []*Entity

	for _, entity := range allEntities {
		if entity == center || !entity.IsAlive {
			continue
		}

		distance := center.DistanceTo(entity)
		if distance <= maxDistance {
			nearby = append(nearby, entity)
		}
	}

	return nearby
}

// renderProgressBar creates a visual progress bar
func (vs *ViewportSystem) renderProgressBar(value, min, max float64, width int) string {
	normalized := (value - min) / (max - min)
	normalized = math.Max(0, math.Min(1, normalized))

	filled := int(normalized * float64(width))

	var bar strings.Builder
	bar.WriteString("[")

	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	bar.WriteString("]")
	return bar.String()
}

// getCellSymbol returns the appropriate symbol for a cell at given zoom level
func (vs *ViewportSystem) getCellSymbol(cell GridCell, world *World, zoom ZoomLevel) rune {
	// Same logic as existing gridView but could be enhanced for different zoom levels
	biome := world.Biomes[cell.Biome]
	symbol := biome.Symbol

	if len(cell.Entities) > 0 {
		// Show dominant species
		speciesCount := make(map[string]int)
		for _, entity := range cell.Entities {
			speciesCount[entity.Species]++
		}

		maxCount := 0
		dominantSpecies := ""
		for species, count := range speciesCount {
			if count > maxCount {
				maxCount = count
				dominantSpecies = species
			}
		}

		switch dominantSpecies {
		case "herbivore":
			symbol = '●'
		case "predator":
			symbol = '▲'
		case "omnivore":
			symbol = '◆'
		}

		// Show multiple entities with numbers for world view
		if zoom == ZoomWorld && len(cell.Entities) > 1 {
			if len(cell.Entities) < 10 {
				symbol = rune('0' + len(cell.Entities))
			} else {
				symbol = '+'
			}
		}
	} else if len(cell.Plants) > 0 {
		// Show plants if no entities
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
		}
	}

	return symbol
}

// getCellStyle returns the appropriate style for a cell at given zoom level
func (vs *ViewportSystem) getCellStyle(cell GridCell, world *World, zoom ZoomLevel) lipgloss.Style {
	// Use existing biome colors but could be enhanced for zoom levels
	return biomeColors[cell.Biome]
}

// getDetailedCellSymbols returns multiple symbols for regional view
func (vs *ViewportSystem) getDetailedCellSymbols(cell GridCell, world *World) []rune {
	// For region view, show up to 4 symbols representing cell contents
	symbols := make([]rune, 0, 4)

	// Biome base
	biome := world.Biomes[cell.Biome]
	symbols = append(symbols, biome.Symbol)

	// Entities (up to 3 more symbols)
	entityCount := 0
	for _, entity := range cell.Entities {
		if entityCount >= 3 {
			break
		}

		switch entity.Species {
		case "herbivore":
			symbols = append(symbols, '●')
		case "predator":
			symbols = append(symbols, '▲')
		case "omnivore":
			symbols = append(symbols, '◆')
		}
		entityCount++
	}

	// Fill remaining with plants if available
	if len(symbols) < 4 && len(cell.Plants) > 0 {
		plantSymbol := '.'
		if len(cell.Plants) > 5 {
			plantSymbol = '■'
		}
		symbols = append(symbols, plantSymbol)
	}

	// Pad to 4 symbols
	for len(symbols) < 4 {
		symbols = append(symbols, ' ')
	}

	return symbols[:4]
}

// getDetailedCellStyles returns styles for regional view symbols
func (vs *ViewportSystem) getDetailedCellStyles(cell GridCell, world *World) []lipgloss.Style {
	styles := make([]lipgloss.Style, 4)
	baseStyle := biomeColors[cell.Biome]

	for i := 0; i < 4; i++ {
		styles[i] = baseStyle
	}

	return styles
}

// Update key bindings for viewport controls
var viewportKeys = struct {
	zoomIn   key.Binding
	zoomOut  key.Binding
	panUp    key.Binding
	panDown  key.Binding
	panLeft  key.Binding
	panRight key.Binding
	select_  key.Binding
	details  key.Binding
}{
	zoomIn: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+", "zoom in"),
	),
	zoomOut: key.NewBinding(
		key.WithKeys("-", "_"),
		key.WithHelp("-", "zoom out"),
	),
	panUp: key.NewBinding(
		key.WithKeys("w", "up"),
		key.WithHelp("w/↑", "pan up"),
	),
	panDown: key.NewBinding(
		key.WithKeys("s", "down"),
		key.WithHelp("s/↓", "pan down"),
	),
	panLeft: key.NewBinding(
		key.WithKeys("a", "left"),
		key.WithHelp("a/←", "pan left"),
	),
	panRight: key.NewBinding(
		key.WithKeys("d", "right"),
		key.WithHelp("d/→", "pan right"),
	),
	select_: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "select entity"),
	),
	details: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle details"),
	),
}
