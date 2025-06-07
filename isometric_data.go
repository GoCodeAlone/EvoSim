package main

import (
	"fmt"
	"math"
)

// IsometricViewData represents the data needed for client-side isometric rendering
type IsometricViewData struct {
	Tiles     []IsometricTile   `json:"tiles"`
	Entities  []IsometricEntity `json:"entities"`
	Plants    []IsometricPlant  `json:"plants"`
	Events    []IsometricEvent  `json:"events"`
	CameraX   float64           `json:"cameraX"`
	CameraY   float64           `json:"cameraY"`
	Zoom      float64           `json:"zoom"`
	WorldInfo WorldInfo         `json:"worldInfo"`
}

// IsometricTile represents a single tile in the isometric view
type IsometricTile struct {
	X         int     `json:"x"`
	Y         int     `json:"y"`
	BiomeType int     `json:"biomeType"`
	BiomeName string  `json:"biomeName"`
	Symbol    string  `json:"symbol"`
	Color     string  `json:"color"`
	Elevation float64 `json:"elevation"`
	Slope     float64 `json:"slope"`
	WaterLevel float64 `json:"waterLevel"`
	TerrainFeatures []int `json:"terrainFeatures,omitempty"`
	GeologicalEvents []IsometricGeologicalEvent `json:"geologicalEvents,omitempty"`
}

// IsometricGeologicalEvent represents geological events for enhanced visualization
type IsometricGeologicalEvent struct {
	ID        int     `json:"id"`
	Type      string  `json:"type"`
	Intensity float64 `json:"intensity"`
	Duration  int     `json:"duration"`
	StartTick int     `json:"startTick"`
	Color     string  `json:"color"`
}

// IsometricEntity represents an entity in the isometric view
type IsometricEntity struct {
	ID       int                    `json:"id"`
	X        float64                `json:"x"`
	Y        float64                `json:"y"`
	Species  string                 `json:"species"`
	Size     float64                `json:"size"`
	Energy   float64                `json:"energy"`
	Age      int                    `json:"age"`
	Color    string                 `json:"color"`
	Traits   map[string]float64     `json:"traits"`
	DNA      IsometricDNA           `json:"dna"`
}

// IsometricPlant represents a plant in the isometric view
type IsometricPlant struct {
	X                 float64 `json:"x"`
	Y                 float64 `json:"y"`
	Type              int     `json:"type"`
	TypeName          string  `json:"typeName"`
	Size              float64 `json:"size"`
	Energy            float64 `json:"energy"`
	Age               int     `json:"age"`
	IsAlive           bool    `json:"isAlive"`
	Color             string  `json:"color"`
	GrowthRate        float64 `json:"growthRate"`
	ReproductionRate  float64 `json:"reproductionRate"`
	MaxSize           float64 `json:"maxSize"`
	DiseaseResistance float64 `json:"diseaseResistance"`
}

// IsometricDNA represents DNA information for the isometric view
type IsometricDNA struct {
	GeneCount   int                    `json:"geneCount"`
	ActiveGenes int                    `json:"activeGenes"`
	Genes       []map[string]interface{} `json:"genes,omitempty"` // Simplified gene representation
}

// IsometricEvent represents a world event for visual effects
type IsometricEvent struct {
	ID          int     `json:"id"`
	Type        string  `json:"type"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Intensity   float64 `json:"intensity"`
	Duration    int     `json:"duration"`
	Age         int     `json:"age"`
	Color       string  `json:"color"`
	Description string  `json:"description"`
}

// WorldInfo provides general world information
type WorldInfo struct {
	Width     int `json:"width"`
	Height    int `json:"height"`
	Tick      int `json:"tick"`
	TotalEntities int `json:"totalEntities"`
	TotalPlants   int `json:"totalPlants"`
}

// IsometricViewManager manages the isometric view data generation
type IsometricViewManager struct {
	world *World
}

// NewIsometricViewManager creates a new isometric view manager
func NewIsometricViewManager(world *World) *IsometricViewManager {
	return &IsometricViewManager{
		world: world,
	}
}

// GenerateIsometricData generates isometric view data for the current world state
func (ivm *IsometricViewManager) GenerateIsometricData(viewportX, viewportY int, zoom float64, maxTiles int) *IsometricViewData {
	// Calculate visible area based on viewport and zoom
	tileRadius := int(math.Max(20, float64(maxTiles) / (zoom * 4)))
	
	startX := math.Max(0, float64(viewportX - tileRadius))
	endX := math.Min(float64(ivm.world.Config.GridWidth), float64(viewportX + tileRadius))
	startY := math.Max(0, float64(viewportY - tileRadius))
	endY := math.Min(float64(ivm.world.Config.GridHeight), float64(viewportY + tileRadius))
	
	data := &IsometricViewData{
		Tiles:     make([]IsometricTile, 0),
		Entities:  make([]IsometricEntity, 0),
		Plants:    make([]IsometricPlant, 0),
		Events:    make([]IsometricEvent, 0),
		CameraX:   float64(viewportX),
		CameraY:   float64(viewportY),
		Zoom:      zoom,
		WorldInfo: WorldInfo{
			Width:         ivm.world.Config.GridWidth,
			Height:        ivm.world.Config.GridHeight,
			Tick:          ivm.world.Tick,
			TotalEntities: len(ivm.world.AllEntities),
		},
	}
	
	// Generate tiles
	for y := int(startY); y < int(endY); y++ {
		for x := int(startX); x < int(endX); x++ {
			if y >= 0 && y < ivm.world.Config.GridHeight && x >= 0 && x < ivm.world.Config.GridWidth {
				cell := ivm.world.Grid[y][x]
				biome := ivm.world.Biomes[cell.Biome]
				
				// Get topology data
				elevation := 0.0
				slope := 0.0
				waterLevel := 0.0
				var terrainFeatures []int
				var geologicalEvents []IsometricGeologicalEvent
				
				if ivm.world.TopologySystem != nil {
					// Map grid coordinates to topology coordinates
					topoX := int((float64(x) / float64(ivm.world.Config.GridWidth)) * float64(ivm.world.TopologySystem.Width))
					topoY := int((float64(y) / float64(ivm.world.Config.GridHeight)) * float64(ivm.world.TopologySystem.Height))
					
					if topoX >= 0 && topoX < ivm.world.TopologySystem.Width && topoY >= 0 && topoY < ivm.world.TopologySystem.Height {
						topoCell := ivm.world.TopologySystem.TopologyGrid[topoX][topoY]
						elevation = topoCell.Elevation
						slope = topoCell.Slope
						waterLevel = topoCell.WaterLevel
						terrainFeatures = topoCell.Features
						
						// Add geological events affecting this cell
						for _, event := range ivm.world.TopologySystem.GeologicalEvents {
							// Check if event affects this cell
							distance := math.Sqrt((float64(x)-event.Center.X)*(float64(x)-event.Center.X) + 
							                     (float64(y)-event.Center.Y)*(float64(y)-event.Center.Y))
							if distance <= event.Radius {
								geoEvent := IsometricGeologicalEvent{
									ID:        event.ID,
									Type:      event.Type,
									Intensity: event.Intensity,
									Duration:  event.Duration,
									StartTick: event.StartTick,
									Color:     ivm.getGeologicalEventColor(event.Type),
								}
								geologicalEvents = append(geologicalEvents, geoEvent)
							}
						}
					}
				}
				
				tile := IsometricTile{
					X:                x,
					Y:                y,
					BiomeType:        int(cell.Biome),
					BiomeName:        biome.Name,
					Symbol:           string(biome.Symbol),
					Color:            ivm.getBiomeColorHex(cell.Biome),
					Elevation:        elevation,
					Slope:            slope,
					WaterLevel:       waterLevel,
					TerrainFeatures:  terrainFeatures,
					GeologicalEvents: geologicalEvents,
				}
				data.Tiles = append(data.Tiles, tile)
				
				// Add plants in this cell
				for _, plant := range cell.Plants {
					if plant.IsAlive {
						isometricPlant := IsometricPlant{
							X:                 plant.Position.X,
							Y:                 plant.Position.Y,
							Type:              int(plant.Type),
							TypeName:          ivm.getPlantTypeName(plant.Type),
							Size:              plant.Size,
							Energy:            plant.Energy,
							Age:               plant.Age,
							IsAlive:           plant.IsAlive,
							Color:             ivm.getPlantColorHex(plant.Type),
							GrowthRate:        plant.GrowthRate,
							ReproductionRate:  ivm.getPlantReproductionRate(plant.Type),
							MaxSize:           ivm.getPlantMaxSize(plant.Type),
							DiseaseResistance: ivm.getPlantDiseaseResistance(plant),
						}
						data.Plants = append(data.Plants, isometricPlant)
						data.WorldInfo.TotalPlants++
					}
				}
				
				// Add entities in this cell
				for _, entity := range cell.Entities {
					// Convert traits to map
					traits := map[string]float64{
						"speed":              ivm.getTraitValue(entity.Traits, "speed"),
						"aggression":         ivm.getTraitValue(entity.Traits, "aggression"),
						"intelligence":       ivm.getTraitValue(entity.Traits, "intelligence"),
						"cooperation":        ivm.getTraitValue(entity.Traits, "cooperation"),
						"defense":            ivm.getTraitValue(entity.Traits, "defense"),
						"size":               ivm.getTraitValue(entity.Traits, "size"),
						"endurance":          ivm.getTraitValue(entity.Traits, "endurance"),
						"strength":           ivm.getTraitValue(entity.Traits, "strength"),
						"aquatic_adaptation": ivm.getTraitValue(entity.Traits, "aquatic_adaptation"),
						"digging_ability":    ivm.getTraitValue(entity.Traits, "digging_ability"),
						"underground_nav":    ivm.getTraitValue(entity.Traits, "underground_nav"),
						"flying_ability":     ivm.getTraitValue(entity.Traits, "flying_ability"),
						"altitude_tolerance": ivm.getTraitValue(entity.Traits, "altitude_tolerance"),
					}
					
					// Convert DNA information (simplified since Entity doesn't have direct DNA field)
					dna := IsometricDNA{
						GeneCount:   len(entity.Traits),
						ActiveGenes: len(entity.Traits),
					}
					
					// Include first few traits as simplified "genes" for DNA visualization
					if len(entity.Traits) > 0 {
						maxGenes := 10
						if len(entity.Traits) < maxGenes {
							maxGenes = len(entity.Traits)
						}
						
						dna.Genes = make([]map[string]interface{}, 0, maxGenes)
						i := 0
						for traitName, trait := range entity.Traits {
							if i >= maxGenes {
								break
							}
							dna.Genes = append(dna.Genes, map[string]interface{}{
								"type":      traitName,
								"value":     trait.Value,
								"active":    true,
								"dominance": 1.0,
							})
							i++
						}
					}
					
					isometricEntity := IsometricEntity{
						ID:      entity.ID,
						X:       entity.Position.X,
						Y:       entity.Position.Y,
						Species: entity.Species,
						Size:    ivm.getTraitValue(entity.Traits, "size"),
						Energy:  entity.Energy,
						Age:     entity.Age,
						Color:   ivm.getEntityColorHex(entity.Species),
						Traits:  traits,
						DNA:     dna,
					}
					data.Entities = append(data.Entities, isometricEntity)
				}
			}
		}
	}
	
	// Generate recent events for visual effects
	ivm.addRecentEvents(data, viewportX, viewportY, tileRadius)
	
	return data
}

// getBiomeColorHex returns hex color for biomes
func (ivm *IsometricViewManager) getBiomeColorHex(biomeType BiomeType) string {
	colors := map[BiomeType]string{
		BiomePlains:       "#64C864", // Green
		BiomeForest:       "#329632", // Dark green
		BiomeWater:        "#3264C8", // Blue
		BiomeMountain:     "#969696", // Gray
		BiomeDesert:       "#C8B464", // Sandy
		BiomeTundra:       "#C8DCFF", // Light blue
		BiomeSwamp:        "#649664", // Murky green
		BiomeIce:          "#F0F0FF", // White
		BiomeRainforest:   "#147814", // Very dark green
		BiomeSoil:         "#8B4513", // Brown
		BiomeAir:          "#C8DCFF", // Transparent blue
		BiomeDeepWater:    "#143C96", // Dark blue
		BiomeHighAltitude: "#B4B4C8", // Light gray
		BiomeCanyon:       "#B47850", // Orange-brown
		BiomeRadiation:    "#FF6464", // Red
		BiomeHotSpring:    "#FF9664", // Orange
	}
	
	if color, exists := colors[biomeType]; exists {
		return color
	}
	return "#808080" // Default gray
}

// getPlantColorHex returns hex color for plant types
func (ivm *IsometricViewManager) getPlantColorHex(plantType PlantType) string {
	colors := map[PlantType]string{
		PlantGrass:    "#64FF64", // Bright green
		PlantBush:     "#32C832", // Medium green
		PlantTree:     "#8B4513", // Brown trunk
		PlantMushroom: "#C89664", // Brown
		PlantAlgae:    "#32FF96", // Cyan-green
		PlantCactus:   "#329632", // Dark green
	}
	
	if color, exists := colors[plantType]; exists {
		return color
	}
	return "#64C864" // Default green
}

// getEntityColorHex returns hex color for entity species
func (ivm *IsometricViewManager) getEntityColorHex(species string) string {
	// Use a hash of the species name to generate consistent colors
	hash := 0
	for _, char := range species {
		hash = int(char) + ((hash << 5) - hash)
	}
	
	// Convert hash to RGB
	r := (hash & 0xFF0000) >> 16
	g := (hash & 0x00FF00) >> 8
	b := hash & 0x0000FF
	
	// Ensure colors are bright enough
	if r < 100 { r += 100 }
	if g < 100 { g += 100 }
	if b < 100 { b += 100 }
	
	return fmt.Sprintf("#%02X%02X%02X", r&0xFF, g&0xFF, b&0xFF)
}

// getPlantTypeName returns human-readable plant type names
func (ivm *IsometricViewManager) getPlantTypeName(plantType PlantType) string {
	names := map[PlantType]string{
		PlantGrass:    "Grass",
		PlantBush:     "Bush",
		PlantTree:     "Tree",
		PlantMushroom: "Mushroom",
		PlantAlgae:    "Algae",
		PlantCactus:   "Cactus",
	}
	
	if name, exists := names[plantType]; exists {
		return name
	}
	return "Unknown Plant"
}

// getTraitValue safely gets a trait value from the entity's trait map
func (ivm *IsometricViewManager) getTraitValue(traits map[string]Trait, traitName string) float64 {
	if trait, exists := traits[traitName]; exists {
		return trait.Value
	}
	return 0.0
}

// getPlantReproductionRate gets reproduction rate from plant config
func (ivm *IsometricViewManager) getPlantReproductionRate(plantType PlantType) float64 {
	configs := GetPlantConfigs()
	if config, exists := configs[plantType]; exists {
		return config.ReproductionRate
	}
	return 0.01
}

// getPlantMaxSize gets max size from plant config
func (ivm *IsometricViewManager) getPlantMaxSize(plantType PlantType) float64 {
	configs := GetPlantConfigs()
	if config, exists := configs[plantType]; exists {
		return config.BaseSize * 3.0 // Assume max size is 3x base size
	}
	return 10.0
}

// getPlantDiseaseResistance calculates disease resistance from plant traits
func (ivm *IsometricViewManager) getPlantDiseaseResistance(plant *Plant) float64 {
	// Use a combination of plant traits to estimate disease resistance
	if resistanceTrait, exists := plant.Traits["disease_resistance"]; exists {
		return resistanceTrait.Value
	}
	// Fallback: use size and age as proxy for resistance
	return math.Min(1.0, (plant.Size + float64(plant.Age)*0.01) * 0.1)
}

// addRecentEvents adds recent world events for visual effects
func (ivm *IsometricViewManager) addRecentEvents(data *IsometricViewData, viewportX, viewportY, radius int) {
	maxEventAge := 50 // Show events for last 50 ticks
	
	// Process recent events from the world's event system
	for i, event := range ivm.world.Events {
		if event == nil {
			continue
		}
		
		eventAge := event.Duration // Use remaining duration as a proxy for age
		if eventAge <= 0 {
			continue
		}
		
		// Check if event is within visible area (roughly)
		eventX := event.Position.X
		eventY := event.Position.Y
		
		// Skip if outside visible area
		distance := math.Sqrt((eventX-float64(viewportX))*(eventX-float64(viewportX)) + 
		                     (eventY-float64(viewportY))*(eventY-float64(viewportY)))
		if distance > float64(radius) {
			continue
		}
		
		isometricEvent := IsometricEvent{
			ID:          i,
			Type:        event.EventType,
			X:           eventX,
			Y:           eventY,
			Intensity:   event.Intensity,
			Duration:    event.Duration,
			Age:         maxEventAge - event.Duration, // Calculate age from remaining duration
			Color:       ivm.getEventColor(event.EventType),
			Description: event.Description,
		}
		
		data.Events = append(data.Events, isometricEvent)
	}
}

// getGeologicalEventColor returns color for geological events
func (ivm *IsometricViewManager) getGeologicalEventColor(eventType string) string {
	colors := map[string]string{
		"earthquake":           "#8B4513", // Brown
		"volcanic_eruption":    "#FF4500", // Orange-red
		"landslide":            "#A0522D", // Sienna
		"flood":                "#1E90FF", // Dodger blue
		"continental_drift":    "#696969", // Dim gray
		"seafloor_spreading":   "#20B2AA", // Light sea green
		"mountain_uplift":      "#708090", // Slate gray
		"rift_valley":          "#8B0000", // Dark red
		"geyser_formation":     "#00FFFF", // Cyan
		"hot_spring_creation":  "#FFB347", // Peach
		"ice_sheet_advance":    "#F0F8FF", // Alice blue
		"glacial_retreat":      "#B0E0E6", // Powder blue
	}
	
	if color, exists := colors[eventType]; exists {
		return color
	}
	return "#808080" // Default gray
}

// getEventColor returns color for different event types
func (ivm *IsometricViewManager) getEventColor(eventType string) string {
	colors := map[string]string{
		"birth":           "#00FF00", // Green
		"death":           "#FF0000", // Red
		"reproduction":    "#FF69B4", // Pink
		"evolution":       "#FFD700", // Gold
		"mutation":        "#9932CC", // Purple
		"migration":       "#00BFFF", // Blue
		"combat":          "#FF4500", // Orange-red
		"cooperation":     "#32CD32", // Lime green
		"tool_creation":   "#8B4513", // Brown
		"structure_built": "#FFB347", // Orange
		"extinction":      "#8B0000", // Dark red
		"speciation":      "#FF1493", // Deep pink
		"environmental":   "#FFFF00", // Yellow
		"disaster":        "#DC143C", // Crimson
		"discovery":       "#00CED1", // Dark turquoise
	}
	
	if color, exists := colors[eventType]; exists {
		return color
	}
	return "#FFFFFF" // Default white
}