package main

import (
	"math"
	"math/rand"
)

// TerrainType represents different terrain features
type TerrainType int

const (
	TerrainFlat TerrainType = iota
	TerrainHill
	TerrainMountain
	TerrainValley
	TerrainRiver
	TerrainLake
	TerrainCanyon
	TerrainCrater
	TerrainVolcano
	TerrainGlacier
)

// WaterBody represents bodies of water
type WaterBody struct {
	ID       int         `json:"id"`
	Type     string      `json:"type"`     // "river", "lake", "stream", "ocean"
	Points   []Position  `json:"points"`   // Path for rivers, outline for lakes
	Flow     float64     `json:"flow"`     // Water flow rate
	Depth    float64     `json:"depth"`    // Average depth
	Salinity float64     `json:"salinity"` // Salt content (0=fresh, 1=salt)
	IsActive bool        `json:"is_active"` // Whether water is flowing/present
}

// TerrainFeature represents a geographic feature
type TerrainFeature struct {
	ID          int           `json:"id"`
	Type        TerrainType   `json:"type"`
	Center      Position      `json:"center"`
	Radius      float64       `json:"radius"`
	Height      float64       `json:"height"`      // Elevation relative to sea level
	Slope       float64       `json:"slope"`       // Steepness (0-1)
	Age         int           `json:"age"`         // How long feature has existed
	Stability   float64       `json:"stability"`   // Resistance to change (0-1)
	Composition string        `json:"composition"` // "rock", "soil", "sand", "ice", etc.
	IsActive    bool          `json:"is_active"`   // For volcanoes, glaciers
}

// TopologyCell represents terrain information for a grid cell
type TopologyCell struct {
	Elevation   float64          `json:"elevation"`    // Height above sea level
	Slope       float64          `json:"slope"`        // Local slope (0-1)
	Aspect      float64          `json:"aspect"`       // Direction of slope (radians)
	Drainage    float64          `json:"drainage"`     // Water drainage rate (0-1)
	Erosion     float64          `json:"erosion"`      // Current erosion rate
	Sediment    float64          `json:"sediment"`     // Accumulated sediment
	WaterLevel  float64          `json:"water_level"`  // Current water level
	Features    []int            `json:"features"`     // IDs of terrain features in this cell
	WaterBodies []int            `json:"water_bodies"` // IDs of water bodies in this cell
	SoilDepth   float64          `json:"soil_depth"`   // Depth of soil layer
	Hardness    float64          `json:"hardness"`     // Rock hardness (affects erosion)
}

// GeologicalEvent represents dynamic geological processes
type GeologicalEvent struct {
	ID          int         `json:"id"`
	Type        string      `json:"type"`        // "earthquake", "volcanic_eruption", "landslide", etc.
	Center      Position    `json:"center"`
	Radius      float64     `json:"radius"`      // Area of effect
	Intensity   float64     `json:"intensity"`   // Strength of event (0-1)
	Duration    int         `json:"duration"`    // Ticks remaining
	StartTick   int         `json:"start_tick"`
	Effects     map[string]float64 `json:"effects"` // Effect type -> magnitude
}

// TopologySystem manages world terrain and geological processes
type TopologySystem struct {
	Width            int                       `json:"width"`
	Height           int                       `json:"height"`
	TopologyGrid     [][]TopologyCell          `json:"topology_grid"`
	TerrainFeatures  map[int]*TerrainFeature   `json:"terrain_features"`
	WaterBodies      map[int]*WaterBody        `json:"water_bodies"`
	GeologicalEvents []GeologicalEvent         `json:"geological_events"`
	SeaLevel         float64                   `json:"sea_level"`
	NextFeatureID    int                       `json:"next_feature_id"`
	NextWaterID      int                       `json:"next_water_id"`
	NextEventID      int                       `json:"next_event_id"`
	ErosionRate      float64                   `json:"erosion_rate"`
	TectonicActivity float64                   `json:"tectonic_activity"`
	ClimateHumidity  float64                   `json:"climate_humidity"`
	CurrentTick      int                       `json:"current_tick"`

	// Terrain generation parameters
	MountainThreshold  float64 `json:"mountain_threshold"`
	HillThreshold      float64 `json:"hill_threshold"`
	ValleyThreshold    float64 `json:"valley_threshold"`
	WaterThreshold     float64 `json:"water_threshold"`
}

// NewTopologySystem creates a new terrain management system
func NewTopologySystem(width, height int) *TopologySystem {
	ts := &TopologySystem{
		Width:            width,
		Height:           height,
		TopologyGrid:     make([][]TopologyCell, width),
		TerrainFeatures:  make(map[int]*TerrainFeature),
		WaterBodies:      make(map[int]*WaterBody),
		GeologicalEvents: make([]GeologicalEvent, 0),
		SeaLevel:         0.0,
		NextFeatureID:    1,
		NextWaterID:      1,
		NextEventID:      1,
		ErosionRate:      0.001,
		TectonicActivity: 0.1,
		ClimateHumidity:  0.5,
		MountainThreshold: 0.7,
		HillThreshold:     0.3,
		ValleyThreshold:   -0.3,
		WaterThreshold:    -0.2,
	}

	// Initialize grid
	for x := 0; x < width; x++ {
		ts.TopologyGrid[x] = make([]TopologyCell, height)
		for y := 0; y < height; y++ {
			ts.TopologyGrid[x][y] = TopologyCell{
				Elevation:   0.0,
				Slope:       0.0,
				Aspect:      0.0,
				Drainage:    0.5,
				Erosion:     0.0,
				Sediment:    0.0,
				WaterLevel:  0.0,
				Features:    make([]int, 0),
				WaterBodies: make([]int, 0),
				SoilDepth:   1.0,
				Hardness:    0.5,
			}
		}
	}

	return ts
}

// GenerateInitialTerrain creates the initial world topology
func (ts *TopologySystem) GenerateInitialTerrain() {
	// Generate base elevation using noise
	ts.generateBaseElevation()
	
	// Add major terrain features
	ts.addMountainRanges()
	ts.addValleys()
	ts.addWaterBodies()
	
	// Calculate derived properties
	ts.calculateSlopes()
	ts.calculateDrainage()
	ts.updateTerrainClassification()
}

// generateBaseElevation creates base elevation using Perlin-like noise
func (ts *TopologySystem) generateBaseElevation() {
	// Simple noise generation (in practice, you'd use proper Perlin noise)
	for x := 0; x < ts.Width; x++ {
		for y := 0; y < ts.Height; y++ {
			// Multi-octave noise
			elevation := 0.0
			amplitude := 0.4 // Reduced amplitude for more reasonable base elevation
			frequency := 0.01
			
			for i := 0; i < 4; i++ {
				elevation += amplitude * ts.noise(float64(x)*frequency, float64(y)*frequency)
				amplitude *= 0.5
				frequency *= 2.0
			}
			
			// Normalize elevation to reasonable range (-0.5 to 0.5)
			// Total amplitude sum is 0.4 + 0.2 + 0.1 + 0.05 = 0.75
			elevation = elevation / 0.75 * 0.5
			
			ts.TopologyGrid[x][y].Elevation = elevation
			
			// Set hardness based on elevation (higher = harder rock)
			ts.TopologyGrid[x][y].Hardness = 0.3 + math.Abs(elevation)*0.4
		}
	}
}

// Simple noise function (replace with proper Perlin noise in production)
func (ts *TopologySystem) noise(x, y float64) float64 {
	return math.Sin(x*2.3+y*1.7)*math.Cos(x*1.1-y*2.9)*0.5 + 
		   math.Sin(x*4.7)*math.Cos(y*3.1)*0.3 +
		   math.Sin(x*8.1+y*6.3)*0.2
}

// addMountainRanges creates mountain ranges
func (ts *TopologySystem) addMountainRanges() {
	numRanges := 2 + rand.Intn(4)
	
	for i := 0; i < numRanges; i++ {
		// Random mountain range
		centerX := rand.Float64() * float64(ts.Width)
		centerY := rand.Float64() * float64(ts.Height)
		length := 20.0 + rand.Float64()*50.0
		width := 5.0 + rand.Float64()*15.0
		height := 0.2 + rand.Float64()*0.4 // Reduced mountain height from 0.5-1.5 to 0.2-0.6
		angle := rand.Float64() * 2 * math.Pi
		
		// Create mountain range feature
		feature := &TerrainFeature{
			ID:          ts.NextFeatureID,
			Type:        TerrainMountain,
			Center:      Position{X: centerX, Y: centerY},
			Radius:      length,
			Height:      height,
			Slope:       0.6 + rand.Float64()*0.3,
			Age:         rand.Intn(10000),
			Stability:   0.8 + rand.Float64()*0.2,
			Composition: "rock",
			IsActive:    false,
		}
		
		ts.TerrainFeatures[ts.NextFeatureID] = feature
		ts.NextFeatureID++
		
		// Apply mountain elevation to grid
		for x := 0; x < ts.Width; x++ {
			for y := 0; y < ts.Height; y++ {
				distance := ts.distanceToLineSegment(float64(x), float64(y), centerX, centerY, angle, length)
				if distance < width {
					influence := math.Exp(-distance*distance/(width*width*0.5))
					elevation := height * influence
					ts.TopologyGrid[x][y].Elevation += elevation
					
					// Add feature reference
					ts.TopologyGrid[x][y].Features = append(ts.TopologyGrid[x][y].Features, feature.ID)
					
					// Increase hardness for mountain rock
					ts.TopologyGrid[x][y].Hardness += 0.2 * influence
				}
			}
		}
	}
}

// addValleys creates valleys and lowlands
func (ts *TopologySystem) addValleys() {
	numValleys := 1 + rand.Intn(3)
	
	for i := 0; i < numValleys; i++ {
		centerX := rand.Float64() * float64(ts.Width)
		centerY := rand.Float64() * float64(ts.Height)
		length := 15.0 + rand.Float64()*40.0
		width := 8.0 + rand.Float64()*20.0
		depth := 0.3 + rand.Float64()*0.5
		angle := rand.Float64() * 2 * math.Pi
		
		feature := &TerrainFeature{
			ID:          ts.NextFeatureID,
			Type:        TerrainValley,
			Center:      Position{X: centerX, Y: centerY},
			Radius:      length,
			Height:      -depth,
			Slope:       0.1 + rand.Float64()*0.2,
			Age:         rand.Intn(8000),
			Stability:   0.6 + rand.Float64()*0.2,
			Composition: "soil",
			IsActive:    false,
		}
		
		ts.TerrainFeatures[ts.NextFeatureID] = feature
		ts.NextFeatureID++
		
		// Apply valley depression to grid
		for x := 0; x < ts.Width; x++ {
			for y := 0; y < ts.Height; y++ {
				distance := ts.distanceToLineSegment(float64(x), float64(y), centerX, centerY, angle, length)
				if distance < width {
					influence := math.Exp(-distance*distance/(width*width*0.5))
					elevation := -depth * influence
					ts.TopologyGrid[x][y].Elevation += elevation
					
					ts.TopologyGrid[x][y].Features = append(ts.TopologyGrid[x][y].Features, feature.ID)
					
					// Softer soil in valleys
					ts.TopologyGrid[x][y].Hardness = math.Max(0.1, ts.TopologyGrid[x][y].Hardness-0.2*influence)
					ts.TopologyGrid[x][y].SoilDepth += 2.0 * influence
				}
			}
		}
	}
}

// addWaterBodies creates rivers and lakes
func (ts *TopologySystem) addWaterBodies() {
	// Add lakes
	numLakes := 1 + rand.Intn(3)
	for i := 0; i < numLakes; i++ {
		ts.createLake()
	}
	
	// Add rivers
	numRivers := 2 + rand.Intn(4)
	for i := 0; i < numRivers; i++ {
		ts.createRiver()
	}
}

// createLake creates a lake in a low-lying area
func (ts *TopologySystem) createLake() {
	// Find low-elevation area
	minElevation := math.Inf(1)
	lakeX, lakeY := 0, 0
	
	for attempt := 0; attempt < 100; attempt++ {
		x := rand.Intn(ts.Width)
		y := rand.Intn(ts.Height)
		elevation := ts.TopologyGrid[x][y].Elevation
		
		if elevation < minElevation {
			minElevation = elevation
			lakeX, lakeY = x, y
		}
	}
	
	radius := 3.0 + rand.Float64()*8.0
	depth := 0.2 + rand.Float64()*0.3
	
	waterBody := &WaterBody{
		ID:       ts.NextWaterID,
		Type:     "lake",
		Points:   []Position{{X: float64(lakeX), Y: float64(lakeY)}},
		Flow:     0.0, // Lakes don't flow
		Depth:    depth,
		Salinity: rand.Float64() * 0.1, // Mostly fresh water
		IsActive: true,
	}
	
	ts.WaterBodies[ts.NextWaterID] = waterBody
	ts.NextWaterID++
	
	// Apply lake to grid
	for x := 0; x < ts.Width; x++ {
		for y := 0; y < ts.Height; y++ {
			distance := math.Sqrt(float64((x-lakeX)*(x-lakeX) + (y-lakeY)*(y-lakeY)))
			if distance < radius {
				influence := math.Exp(-distance*distance/(radius*radius*0.5))
				
				// Depress elevation for lake bed
				ts.TopologyGrid[x][y].Elevation -= depth * influence
				ts.TopologyGrid[x][y].WaterLevel = depth * influence
				ts.TopologyGrid[x][y].WaterBodies = append(ts.TopologyGrid[x][y].WaterBodies, waterBody.ID)
				
				// Increase drainage near water
				ts.TopologyGrid[x][y].Drainage = math.Min(1.0, ts.TopologyGrid[x][y].Drainage+0.3*influence)
			}
		}
	}
}

// createRiver creates a river flowing from high to low elevation
func (ts *TopologySystem) createRiver() {
	// Find high starting point
	maxElevation := math.Inf(-1)
	startX, startY := 0, 0
	
	for attempt := 0; attempt < 100; attempt++ {
		x := rand.Intn(ts.Width)
		y := rand.Intn(ts.Height)
		elevation := ts.TopologyGrid[x][y].Elevation
		
		if elevation > maxElevation {
			maxElevation = elevation
			startX, startY = x, y
		}
	}
	
	// Trace river path downhill
	riverPoints := make([]Position, 0)
	currentX, currentY := float64(startX), float64(startY)
	
	for len(riverPoints) < 100 { // Maximum river length
		riverPoints = append(riverPoints, Position{X: currentX, Y: currentY})
		
		// Find steepest downhill direction
		bestX, bestY := currentX, currentY
		bestElevation := ts.getElevationAt(currentX, currentY)
		
		for dx := -1.0; dx <= 1.0; dx++ {
			for dy := -1.0; dy <= 1.0; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				
				newX := currentX + dx
				newY := currentY + dy
				
				if newX >= 0 && newX < float64(ts.Width) && newY >= 0 && newY < float64(ts.Height) {
					elevation := ts.getElevationAt(newX, newY)
					if elevation < bestElevation {
						bestElevation = elevation
						bestX, bestY = newX, newY
					}
				}
			}
		}
		
		// If no downhill path, stop
		if bestX == currentX && bestY == currentY {
			break
		}
		
		currentX, currentY = bestX, bestY
		
		// Stop if we reach sea level or existing water
		if bestElevation <= ts.SeaLevel || ts.TopologyGrid[int(currentX)][int(currentY)].WaterLevel > 0 {
			break
		}
	}
	
	if len(riverPoints) < 5 {
		return // River too short
	}
	
	flow := 0.5 + rand.Float64()*0.5
	waterBody := &WaterBody{
		ID:       ts.NextWaterID,
		Type:     "river",
		Points:   riverPoints,
		Flow:     flow,
		Depth:    0.1 + flow*0.2,
		Salinity: 0.0, // Fresh water
		IsActive: true,
	}
	
	ts.WaterBodies[ts.NextWaterID] = waterBody
	ts.NextWaterID++
	
	// Apply river to grid
	riverWidth := 1.0 + flow*2.0
	for _, point := range riverPoints {
		for x := 0; x < ts.Width; x++ {
			for y := 0; y < ts.Height; y++ {
				distance := math.Sqrt((float64(x)-point.X)*(float64(x)-point.X) + 
					                 (float64(y)-point.Y)*(float64(y)-point.Y))
				if distance < riverWidth {
					influence := math.Exp(-distance*distance/(riverWidth*riverWidth*0.5))
					
					ts.TopologyGrid[x][y].WaterLevel = math.Max(ts.TopologyGrid[x][y].WaterLevel, 
						                                      waterBody.Depth*influence)
					ts.TopologyGrid[x][y].WaterBodies = append(ts.TopologyGrid[x][y].WaterBodies, waterBody.ID)
					ts.TopologyGrid[x][y].Drainage = math.Min(1.0, ts.TopologyGrid[x][y].Drainage+0.5*influence)
				}
			}
		}
	}
}

// UpdateTopology processes dynamic geological changes
func (ts *TopologySystem) UpdateTopology(currentTick int) {
	ts.CurrentTick = currentTick
	
	// Process ongoing geological events
	ts.processGeologicalEvents()
	
	// Apply erosion
	if currentTick%10 == 0 { // Every 10 ticks
		ts.applyErosion()
	}
	
	// Trigger random geological events
	if currentTick%100 == 0 { // Check every 100 ticks
		ts.triggerRandomEvents()
	}
	
	// Update water flow
	if currentTick%5 == 0 { // Every 5 ticks
		ts.updateWaterFlow()
	}
	
	// Recalculate terrain properties periodically
	if currentTick%50 == 0 {
		ts.calculateSlopes()
		ts.updateTerrainClassification()
	}
}

// processGeologicalEvents handles active geological events
func (ts *TopologySystem) processGeologicalEvents() {
	for i := len(ts.GeologicalEvents) - 1; i >= 0; i-- {
		event := &ts.GeologicalEvents[i]
		event.Duration--
		
		// Apply event effects
		ts.applyGeologicalEventEffects(event)
		
		// Remove expired events
		if event.Duration <= 0 {
			ts.GeologicalEvents = append(ts.GeologicalEvents[:i], ts.GeologicalEvents[i+1:]...)
		}
	}
}

// applyGeologicalEventEffects applies the effects of a geological event
func (ts *TopologySystem) applyGeologicalEventEffects(event *GeologicalEvent) {
	centerX := int(event.Center.X)
	centerY := int(event.Center.Y)
	radius := int(event.Radius)
	
	for x := centerX - radius; x <= centerX + radius; x++ {
		for y := centerY - radius; y <= centerY + radius; y++ {
			if x < 0 || x >= ts.Width || y < 0 || y >= ts.Height {
				continue
			}
			
			distance := math.Sqrt(float64((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY)))
			if distance > event.Radius {
				continue
			}
			
			influence := (event.Radius - distance) / event.Radius * event.Intensity
			
			switch event.Type {
			case "earthquake":
				// Randomly alter elevation
				change := (rand.Float64() - 0.5) * influence * 0.2
				ts.TopologyGrid[x][y].Elevation += change
				ts.TopologyGrid[x][y].Erosion += influence * 0.1
				
			case "volcanic_eruption":
				// Add elevation and new rock
				ts.TopologyGrid[x][y].Elevation += influence * 0.5
				ts.TopologyGrid[x][y].Hardness = math.Min(1.0, ts.TopologyGrid[x][y].Hardness+influence*0.3)
				
			case "landslide":
				// Move sediment downhill
				ts.TopologyGrid[x][y].Sediment += influence * 0.3
				ts.TopologyGrid[x][y].Elevation -= influence * 0.1
				
			case "flood":
				// Increase water level and erosion
				ts.TopologyGrid[x][y].WaterLevel += influence * 0.2
				ts.TopologyGrid[x][y].Erosion += influence * 0.2
				
			// New plate tectonics events
			case "continental_drift":
				// Gradual elevation changes over large areas
				change := (rand.Float64() - 0.5) * influence * 0.05 // Very gradual
				ts.TopologyGrid[x][y].Elevation += change
				
			case "seafloor_spreading":
				// Create new oceanic crust - lower elevation, harder rock
				if ts.TopologyGrid[x][y].Elevation < 0.2 { // Only in low areas
					ts.TopologyGrid[x][y].Elevation -= influence * 0.3
					ts.TopologyGrid[x][y].Hardness = math.Min(1.0, ts.TopologyGrid[x][y].Hardness+influence*0.4)
				}
				
			case "mountain_uplift":
				// Create mountain ranges - increase elevation and hardness
				ts.TopologyGrid[x][y].Elevation += influence * 0.8
				ts.TopologyGrid[x][y].Hardness = math.Min(1.0, ts.TopologyGrid[x][y].Hardness+influence*0.5)
				ts.TopologyGrid[x][y].SoilDepth = math.Max(0.1, ts.TopologyGrid[x][y].SoilDepth-influence*0.3)
				
			case "rift_valley":
				// Create deep valleys - decrease elevation
				ts.TopologyGrid[x][y].Elevation -= influence * 0.6
				ts.TopologyGrid[x][y].Erosion += influence * 0.3
				ts.TopologyGrid[x][y].Sediment += influence * 0.2
				
			case "geyser_formation":
				// Create localized hot water features
				if distance < 2.0 { // Very localized effect
					ts.TopologyGrid[x][y].WaterLevel += influence * 0.4
					// Mark as potential hot spring location (would need biome integration)
				}
				
			case "hot_spring_creation":
				// Create hot water features
				if distance < 3.0 {
					ts.TopologyGrid[x][y].WaterLevel += influence * 0.3
					// Mark as hot spring location
				}
				
			case "ice_sheet_advance":
				// Create ice coverage - affects surface conditions
				if ts.TopologyGrid[x][y].Elevation > 0.3 { // Higher elevations more susceptible
					ts.TopologyGrid[x][y].WaterLevel += influence * 0.2 // Ice coverage
					ts.TopologyGrid[x][y].Erosion += influence * 0.1 // Glacial erosion
				}
				
			case "glacial_retreat":
				// Remove ice coverage, expose underlying terrain
				ts.TopologyGrid[x][y].WaterLevel = math.Max(0, ts.TopologyGrid[x][y].WaterLevel-influence*0.2)
				ts.TopologyGrid[x][y].Sediment += influence * 0.3 // Glacial deposits
			}
		}
	}
}

// triggerRandomEvents creates random geological events
func (ts *TopologySystem) triggerRandomEvents() {
	if rand.Float64() < ts.TectonicActivity * 0.01 { // 1% chance with tectonic activity
		eventTypes := []string{
			"earthquake", "volcanic_eruption", "landslide", "flood",
			// New plate tectonics events
			"continental_drift", "seafloor_spreading", "mountain_uplift", 
			"rift_valley", "geyser_formation", "hot_spring_creation",
			"ice_sheet_advance", "glacial_retreat",
		}
		eventType := eventTypes[rand.Intn(len(eventTypes))]
		
		centerX := rand.Float64() * float64(ts.Width)
		centerY := rand.Float64() * float64(ts.Height)
		
		var radius, intensity float64
		var duration int
		
		switch eventType {
		case "earthquake":
			radius = 5.0 + rand.Float64()*15.0
			intensity = 0.3 + rand.Float64()*0.7
			duration = 1 + rand.Intn(3)
			
		case "volcanic_eruption":
			radius = 3.0 + rand.Float64()*8.0
			intensity = 0.5 + rand.Float64()*0.5
			duration = 5 + rand.Intn(20)
			
		case "landslide":
			radius = 2.0 + rand.Float64()*5.0
			intensity = 0.4 + rand.Float64()*0.4
			duration = 1 + rand.Intn(2)
			
		case "flood":
			radius = 8.0 + rand.Float64()*20.0
			intensity = 0.2 + rand.Float64()*0.5
			duration = 10 + rand.Intn(30)
			
		// New plate tectonics events
		case "continental_drift":
			radius = 30.0 + rand.Float64()*50.0 // Large scale
			intensity = 0.1 + rand.Float64()*0.3
			duration = 100 + rand.Intn(300) // Very long duration
			
		case "seafloor_spreading":
			radius = 15.0 + rand.Float64()*25.0
			intensity = 0.2 + rand.Float64()*0.4
			duration = 50 + rand.Intn(150)
			
		case "mountain_uplift":
			radius = 10.0 + rand.Float64()*20.0
			intensity = 0.4 + rand.Float64()*0.6
			duration = 20 + rand.Intn(80)
			
		case "rift_valley":
			radius = 12.0 + rand.Float64()*18.0
			intensity = 0.3 + rand.Float64()*0.5
			duration = 30 + rand.Intn(100)
			
		case "geyser_formation":
			radius = 1.0 + rand.Float64()*3.0 // Small, localized
			intensity = 0.6 + rand.Float64()*0.4
			duration = 50 + rand.Intn(200)
			
		case "hot_spring_creation":
			radius = 2.0 + rand.Float64()*4.0
			intensity = 0.5 + rand.Float64()*0.3
			duration = 40 + rand.Intn(150)
			
		case "ice_sheet_advance":
			radius = 20.0 + rand.Float64()*40.0
			intensity = 0.3 + rand.Float64()*0.5
			duration = 80 + rand.Intn(200)
			
		case "glacial_retreat":
			radius = 15.0 + rand.Float64()*30.0
			intensity = 0.2 + rand.Float64()*0.4
			duration = 60 + rand.Intn(150)
		}
		
		event := GeologicalEvent{
			ID:        ts.NextEventID,
			Type:      eventType,
			Center:    Position{X: centerX, Y: centerY},
			Radius:    radius,
			Intensity: intensity,
			Duration:  duration,
			StartTick: ts.CurrentTick,
			Effects:   make(map[string]float64),
		}
		
		ts.GeologicalEvents = append(ts.GeologicalEvents, event)
		ts.NextEventID++
	}
}

// Helper functions

func (ts *TopologySystem) distanceToLineSegment(px, py, x1, y1, angle, length float64) float64 {
	x2 := x1 + length*math.Cos(angle)
	y2 := y1 + length*math.Sin(angle)
	
	// Distance from point to line segment
	A := px - x1
	B := py - y1
	C := x2 - x1
	D := y2 - y1
	
	dot := A*C + B*D
	lenSq := C*C + D*D
	param := -1.0
	
	if lenSq != 0 {
		param = dot / lenSq
	}
	
	var xx, yy float64
	if param < 0 {
		xx, yy = x1, y1
	} else if param > 1 {
		xx, yy = x2, y2
	} else {
		xx = x1 + param*C
		yy = y1 + param*D
	}
	
	dx := px - xx
	dy := py - yy
	return math.Sqrt(dx*dx + dy*dy)
}

func (ts *TopologySystem) getElevationAt(x, y float64) float64 {
	if x < 0 || x >= float64(ts.Width) || y < 0 || y >= float64(ts.Height) {
		return ts.SeaLevel
	}
	return ts.TopologyGrid[int(x)][int(y)].Elevation
}

func (ts *TopologySystem) calculateSlopes() {
	for x := 1; x < ts.Width-1; x++ {
		for y := 1; y < ts.Height-1; y++ {
			// Calculate slope using neighboring cells
			dzdx := (ts.TopologyGrid[x+1][y].Elevation - ts.TopologyGrid[x-1][y].Elevation) / 2.0
			dzdy := (ts.TopologyGrid[x][y+1].Elevation - ts.TopologyGrid[x][y-1].Elevation) / 2.0
			
			slope := math.Sqrt(dzdx*dzdx + dzdy*dzdy)
			ts.TopologyGrid[x][y].Slope = math.Min(1.0, slope)
			
			// Calculate aspect (direction of slope)
			ts.TopologyGrid[x][y].Aspect = math.Atan2(dzdy, dzdx)
		}
	}
}

func (ts *TopologySystem) calculateDrainage() {
	for x := 0; x < ts.Width; x++ {
		for y := 0; y < ts.Height; y++ {
			// Drainage based on slope and surrounding water
			drainage := ts.TopologyGrid[x][y].Slope * 0.5
			
			// Higher drainage near water bodies
			if ts.TopologyGrid[x][y].WaterLevel > 0 {
				drainage += 0.5
			}
			
			ts.TopologyGrid[x][y].Drainage = math.Min(1.0, drainage)
		}
	}
}

func (ts *TopologySystem) updateTerrainClassification() {
	// This would update the biome classification based on terrain
	// Integration with existing biome system would happen here
}

func (ts *TopologySystem) applyErosion() {
	for x := 0; x < ts.Width; x++ {
		for y := 0; y < ts.Height; y++ {
			cell := &ts.TopologyGrid[x][y]
			
			// Erosion rate based on slope, water, and hardness
			erosionRate := ts.ErosionRate * cell.Slope * (1.0 + cell.WaterLevel) * (1.0 - cell.Hardness)
			
			if erosionRate > 0 {
				elevation_loss := erosionRate * (1.0 + rand.Float64()*0.5)
				cell.Elevation -= elevation_loss
				cell.Sediment += elevation_loss * 0.7 // Some material becomes sediment
				cell.Erosion = erosionRate
			}
		}
	}
}

func (ts *TopologySystem) updateWaterFlow() {
	// Simplified water flow simulation
	for _, waterBody := range ts.WaterBodies {
		if waterBody.Type == "river" && waterBody.IsActive {
			// Rivers can change course over time
			if rand.Float64() < 0.01 { // 1% chance per update
				ts.adjustRiverCourse(waterBody)
			}
		}
	}
}

func (ts *TopologySystem) adjustRiverCourse(river *WaterBody) {
	// Slight random adjustment to river path
	if len(river.Points) > 2 {
		pointIndex := 1 + rand.Intn(len(river.Points)-2)
		point := &river.Points[pointIndex]
		
		// Small random movement
		point.X += (rand.Float64() - 0.5) * 2.0
		point.Y += (rand.Float64() - 0.5) * 2.0
		
		// Clamp to bounds
		point.X = math.Max(0, math.Min(float64(ts.Width-1), point.X))
		point.Y = math.Max(0, math.Min(float64(ts.Height-1), point.Y))
	}
}

// GetTopologyStats returns comprehensive topology statistics
func (ts *TopologySystem) GetTopologyStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["terrain_features"] = len(ts.TerrainFeatures)
	stats["water_bodies"] = len(ts.WaterBodies)
	stats["geological_events"] = len(ts.GeologicalEvents)
	stats["sea_level"] = ts.SeaLevel
	stats["erosion_rate"] = ts.ErosionRate
	stats["tectonic_activity"] = ts.TectonicActivity
	
	// Calculate terrain statistics
	totalElevation := 0.0
	totalSlope := 0.0
	waterCells := 0
	
	for x := 0; x < ts.Width; x++ {
		for y := 0; y < ts.Height; y++ {
			cell := ts.TopologyGrid[x][y]
			totalElevation += cell.Elevation
			totalSlope += cell.Slope
			if cell.WaterLevel > 0 {
				waterCells++
			}
		}
	}
	
	totalCells := ts.Width * ts.Height
	stats["avg_elevation"] = totalElevation / float64(totalCells)
	stats["avg_slope"] = totalSlope / float64(totalCells)
	stats["water_coverage"] = float64(waterCells) / float64(totalCells)
	
	return stats
}

// GetTerrainAt returns terrain information for a specific location
func (ts *TopologySystem) GetTerrainAt(x, y int) *TopologyCell {
	if x < 0 || x >= ts.Width || y < 0 || y >= ts.Height {
		return nil
	}
	return &ts.TopologyGrid[x][y]
}

// GetTerrainTypeName returns a human-readable terrain type name
func (ts *TopologySystem) GetTerrainTypeName(terrainType TerrainType) string {
	names := map[TerrainType]string{
		TerrainFlat:     "Flat",
		TerrainHill:     "Hill",
		TerrainMountain: "Mountain",
		TerrainValley:   "Valley",
		TerrainRiver:    "River",
		TerrainLake:     "Lake",
		TerrainCanyon:   "Canyon",
		TerrainCrater:   "Crater",
		TerrainVolcano:  "Volcano",
		TerrainGlacier:  "Glacier",
	}
	
	if name, exists := names[terrainType]; exists {
		return name
	}
	return "Unknown"
}