package main

import (
	"fmt"
	"math"
	"math/rand"
)

// Season constants
const (
	SeasonSpring = "spring"
	SeasonWinter = "winter"
)

// PlantType represents different types of plants
type PlantType int

const (
	PlantGrass PlantType = iota
	PlantBush
	PlantTree
	PlantMushroom
	PlantAlgae
	PlantCactus
	// Aquatic plant types for water dispersal system
	PlantLily // Water lilies - floating surface plants
	PlantReed // Marsh reeds - wetland plants
	PlantKelp // Kelp - deep water plants
)

// Plant represents a plant entity in the ecosystem
type Plant struct {
	ID           int              `json:"id"`
	Type         PlantType        `json:"type"`
	Position     Position         `json:"position"`
	Traits       map[string]Trait `json:"traits"`
	Energy       float64          `json:"energy"`
	Age          int              `json:"age"`
	IsAlive      bool             `json:"is_alive"`
	Generation   int              `json:"generation"`
	Size         float64          `json:"size"`
	NutritionVal float64          `json:"nutrition_value"`
	Toxicity     float64          `json:"toxicity"`
	GrowthRate   float64          `json:"growth_rate"`

	// Molecular system
	MolecularProfile *MolecularProfile `json:"molecular_profile"`

	// Enhanced nutrient system
	SoilNutrients   map[string]float64 `json:"soil_nutrients"`   // Available nutrients in soil
	WaterLevel      float64            `json:"water_level"`      // Available water
	SunlightLevel   float64            `json:"sunlight_level"`   // Available sunlight
	NutrientNeeds   map[string]float64 `json:"nutrient_needs"`   // Required nutrients for growth
	WaterDependency float64            `json:"water_dependency"` // How much water this plant needs
	SoilPH          float64            `json:"soil_ph"`          // Soil acidity level (6-8 optimal)
	RootDepth       float64            `json:"root_depth"`       // How deep roots go for nutrients
}

// PlantConfig defines the characteristics of different plant types
type PlantConfig struct {
	Name             string
	Symbol           rune
	BaseEnergy       float64
	BaseSize         float64
	BaseNutrition    float64
	BaseToxicity     float64
	BaseGrowthRate   float64
	PreferredBiomes  []BiomeType
	MutationRate     float64
	ReproductionRate float64
	MaxAge           int
}

// GetPlantConfigs returns configurations for different plant types
func GetPlantConfigs() map[PlantType]PlantConfig {
	return map[PlantType]PlantConfig{
		PlantGrass: {
			Name:             "Grass",
			Symbol:           '.',
			BaseEnergy:       20,
			BaseSize:         0.5,
			BaseNutrition:    15,
			BaseToxicity:     0,
			BaseGrowthRate:   0.8,
			PreferredBiomes:  []BiomeType{BiomePlains, BiomeForest},
			MutationRate:     0.1,
			ReproductionRate: 0.3,
			MaxAge:           50,
		},
		PlantBush: {
			Name:             "Bush",
			Symbol:           '♦',
			BaseEnergy:       40,
			BaseSize:         1.0,
			BaseNutrition:    25,
			BaseToxicity:     0.1,
			BaseGrowthRate:   0.5,
			PreferredBiomes:  []BiomeType{BiomePlains, BiomeForest, BiomeDesert},
			MutationRate:     0.08,
			ReproductionRate: 0.2,
			MaxAge:           100,
		},
		PlantTree: {
			Name:             "Tree",
			Symbol:           '♠',
			BaseEnergy:       100,
			BaseSize:         3.0,
			BaseNutrition:    40,
			BaseToxicity:     0,
			BaseGrowthRate:   0.2,
			PreferredBiomes:  []BiomeType{BiomeForest, BiomePlains},
			MutationRate:     0.05,
			ReproductionRate: 0.1,
			MaxAge:           300,
		},
		PlantMushroom: {
			Name:             "Mushroom",
			Symbol:           '♪',
			BaseEnergy:       30,
			BaseSize:         0.8,
			BaseNutrition:    20,
			BaseToxicity:     0.5,
			BaseGrowthRate:   0.6,
			PreferredBiomes:  []BiomeType{BiomeForest, BiomeRadiation},
			MutationRate:     0.15,
			ReproductionRate: 0.25,
			MaxAge:           30,
		},
		PlantAlgae: {
			Name:             "Algae",
			Symbol:           '~',
			BaseEnergy:       15,
			BaseSize:         0.3,
			BaseNutrition:    10,
			BaseToxicity:     0,
			BaseGrowthRate:   1.0,
			PreferredBiomes:  []BiomeType{BiomeWater},
			MutationRate:     0.12,
			ReproductionRate: 0.4,
			MaxAge:           20,
		},
		PlantCactus: {
			Name:             "Cactus",
			Symbol:           '†',
			BaseEnergy:       60,
			BaseSize:         1.5,
			BaseNutrition:    30,
			BaseToxicity:     0.2,
			BaseGrowthRate:   0.3,
			PreferredBiomes:  []BiomeType{BiomeDesert},
			MutationRate:     0.06,
			ReproductionRate: 0.15,
			MaxAge:           200,
		},
		PlantLily: {
			Name:             "Water Lily",
			Symbol:           '◐',
			BaseEnergy:       35,
			BaseSize:         0.8,
			BaseNutrition:    20,
			BaseToxicity:     0.05,
			BaseGrowthRate:   0.6,
			PreferredBiomes:  []BiomeType{BiomeWater, BiomeSwamp},
			MutationRate:     0.08,
			ReproductionRate: 0.25,
			MaxAge:           80,
		},
		PlantReed: {
			Name:             "Reed",
			Symbol:           '|',
			BaseEnergy:       25,
			BaseSize:         1.2,
			BaseNutrition:    18,
			BaseToxicity:     0.0,
			BaseGrowthRate:   0.7,
			PreferredBiomes:  []BiomeType{BiomeSwamp, BiomeWater},
			MutationRate:     0.09,
			ReproductionRate: 0.3,
			MaxAge:           60,
		},
		PlantKelp: {
			Name:             "Kelp",
			Symbol:           '⌐',
			BaseEnergy:       45,
			BaseSize:         2.0,
			BaseNutrition:    25,
			BaseToxicity:     0.1,
			BaseGrowthRate:   0.5,
			PreferredBiomes:  []BiomeType{BiomeDeepWater, BiomeWater},
			MutationRate:     0.07,
			ReproductionRate: 0.2,
			MaxAge:           120,
		},
	}
}

// NewPlant creates a new plant with random traits
func NewPlant(id int, plantType PlantType, position Position) *Plant {
	config := GetPlantConfigs()[plantType]

	plant := &Plant{
		ID:           id,
		Type:         plantType,
		Position:     position,
		Traits:       make(map[string]Trait),
		Energy:       config.BaseEnergy,
		Age:          0,
		IsAlive:      true,
		Generation:   0,
		Size:         config.BaseSize,
		NutritionVal: config.BaseNutrition,
		Toxicity:     config.BaseToxicity,
		GrowthRate:   config.BaseGrowthRate,
	}

	// Initialize traits with some randomness
	traitNames := []string{"growth_efficiency", "defense", "nutrition_density", "toxin_production", "hardiness", "reproduction_rate"}
	for _, name := range traitNames {
		plant.Traits[name] = Trait{
			Name:  name,
			Value: (rand.Float64()*2 - 1) * 0.5, // Smaller initial variation for plants
		}
	}

	// Apply trait influences
	plant.GrowthRate += plant.GetTrait("growth_efficiency") * 0.2
	plant.NutritionVal += plant.GetTrait("nutrition_density") * 10
	plant.Toxicity += math.Max(0, plant.GetTrait("toxin_production")*0.3)

	// Initialize molecular profile
	plant.MolecularProfile = CreatePlantMolecularProfile(plant)

	// Initialize enhanced nutrient system
	plant.SoilNutrients = make(map[string]float64)
	plant.WaterLevel = 0.5    // Start with moderate water
	plant.SunlightLevel = 1.0 // Full sunlight available initially
	plant.NutrientNeeds = initializePlantNutrientNeeds(plantType)
	plant.WaterDependency = getPlantWaterDependency(plantType)
	plant.SoilPH = 7.0 // Neutral pH preference initially
	plant.RootDepth = getPlantRootDepth(plantType)

	return plant
}

// GetTrait safely gets a trait value
func (p *Plant) GetTrait(name string) float64 {
	if trait, exists := p.Traits[name]; exists {
		return trait.Value
	}
	return 0.0
}

// SetTrait sets a trait value
func (p *Plant) SetTrait(name string, value float64) {
	p.Traits[name] = Trait{Name: name, Value: value}
}

// Update handles plant growth, aging, and natural death
func (p *Plant) Update(biome Biome) {
	if !p.IsAlive {
		return
	}

	p.Age++

	// Growth based on biome suitability and traits
	config := GetPlantConfigs()[p.Type]
	biomeSuitability := 1.0

	// Check if this biome is preferred
	preferred := false
	for _, preferredBiome := range config.PreferredBiomes {
		if biome.Type == preferredBiome {
			preferred = true
			break
		}
	}

	if !preferred {
		biomeSuitability = 0.3 // Much slower growth in non-preferred biomes
	}

	// Apply biome energy effects
	p.Energy -= biome.EnergyDrain * 0.5 // Plants are more resilient to energy drain

	// Growth
	hardiness := p.GetTrait("hardiness")
	growthRate := (p.GrowthRate + hardiness*0.1) * biomeSuitability

	if p.Energy > 10 {
		energyGrowth := growthRate * 2
		p.Energy += energyGrowth
		p.Size += growthRate * 0.1

		// Growing plants provide more nutrition
		p.NutritionVal = config.BaseNutrition + p.Size*5 + p.GetTrait("nutrition_density")*5
	}

	// Energy decay
	p.Energy -= 1.0 + p.Size*0.5

	// Death conditions
	if p.Energy <= 0 {
		p.IsAlive = false
	}

	// Death from old age
	maxAge := config.MaxAge + int(hardiness*20)
	if p.Age > maxAge {
		p.IsAlive = false
	}
}

// CanReproduce checks if the plant can reproduce
func (p *Plant) CanReproduce() bool {
	if !p.IsAlive || p.Age < 5 {
		return false
	}

	config := GetPlantConfigs()[p.Type]
	energyThreshold := config.BaseEnergy * 2

	reproductionBonus := p.GetTrait("reproduction_rate")
	reproductionChance := config.ReproductionRate + reproductionBonus*0.1

	return p.Energy > energyThreshold && rand.Float64() < reproductionChance
}

// Reproduce creates a new plant offspring
func (p *Plant) Reproduce(newID int) *Plant {
	if !p.CanReproduce() {
		return nil
	}

	// Create offspring near parent
	offset := 2.0 + rand.Float64()*3.0
	angle := rand.Float64() * 2 * math.Pi

	newPos := Position{
		X: p.Position.X + math.Cos(angle)*offset,
		Y: p.Position.Y + math.Sin(angle)*offset,
	}

	offspring := NewPlant(newID, p.Type, newPos)
	offspring.Generation = p.Generation + 1

	// Inherit traits with mutation
	config := GetPlantConfigs()[p.Type]
	for name, trait := range p.Traits {
		newValue := trait.Value

		// Mutation
		if rand.Float64() < config.MutationRate {
			mutation := rand.NormFloat64() * 0.1
			newValue += mutation
			newValue = math.Max(-1.0, math.Min(1.0, newValue))
		}

		offspring.SetTrait(name, newValue)
	}

	// Reproduction costs energy
	p.Energy -= config.BaseEnergy * 0.5

	return offspring
}

// GetNutritionValue returns the nutrition value for consumption
func (p *Plant) GetNutritionValue() float64 {
	if !p.IsAlive {
		return p.NutritionVal * 0.3 // Dead plants have less nutrition
	}
	return p.NutritionVal
}

// GetToxicity returns the toxicity level
func (p *Plant) GetToxicity() float64 {
	defenseBonus := p.GetTrait("defense") * 0.2
	return math.Max(0, p.Toxicity+defenseBonus)
}

// Consume handles being eaten by an entity
func (p *Plant) Consume(eatenAmount float64) float64 {
	if !p.IsAlive {
		return 0
	}

	// Calculate actual nutrition provided
	actualNutrition := math.Min(eatenAmount, p.NutritionVal)

	// Damage the plant
	damage := eatenAmount * (1.0 + p.Size*0.1)
	p.Energy -= damage
	p.Size = math.Max(0.1, p.Size-eatenAmount*0.1)

	// Update nutrition based on size
	config := GetPlantConfigs()[p.Type]
	p.NutritionVal = math.Max(5, config.BaseNutrition+p.Size*5)

	// Death from being eaten
	if p.Energy <= 0 || p.Size <= 0.1 {
		p.IsAlive = false
	}

	return actualNutrition
}

// Clone creates a copy of the plant
func (p *Plant) Clone() *Plant {
	clone := &Plant{
		ID:           p.ID,
		Type:         p.Type,
		Position:     p.Position,
		Traits:       make(map[string]Trait),
		Energy:       p.Energy,
		Age:          p.Age,
		IsAlive:      p.IsAlive,
		Generation:   p.Generation,
		Size:         p.Size,
		NutritionVal: p.NutritionVal,
		Toxicity:     p.Toxicity,
		GrowthRate:   p.GrowthRate,
	}

	for name, trait := range p.Traits {
		clone.Traits[name] = trait
	}

	return clone
}

// GetSymbol returns the display symbol for the plant
func (p *Plant) GetSymbol() rune {
	config := GetPlantConfigs()[p.Type]
	return config.Symbol
}

// Statistical tracking helper methods for plants

// SetEnergyWithTracking sets plant energy and logs the change for statistical analysis
func (p *Plant) SetEnergyWithTracking(newEnergy float64, world *World, reason string) {
	if world.StatisticalReporter != nil {
		oldEnergy := p.Energy
		world.StatisticalReporter.LogPlantEvent(world.Tick, "energy_change", p, oldEnergy, newEnergy)

		// Add detailed metadata about the change
		metadata := map[string]interface{}{
			"reason":     reason,
			"magnitude":  newEnergy - oldEnergy,
			"plant_type": p.Type,
		}
		world.StatisticalReporter.LogSystemEvent(world.Tick, "plant_energy_change", reason, metadata)
	}
	p.Energy = newEnergy
}

// ModifyEnergyWithTracking modifies plant energy by a delta amount and logs the change
func (p *Plant) ModifyEnergyWithTracking(delta float64, world *World, reason string) {
	newEnergy := p.Energy + delta
	p.SetEnergyWithTracking(newEnergy, world, reason)
}

// SetSizeWithTracking sets plant size and logs the change for statistical analysis
func (p *Plant) SetSizeWithTracking(newSize float64, world *World, reason string) {
	if world.StatisticalReporter != nil {
		oldSize := p.Size
		world.StatisticalReporter.LogPlantEvent(world.Tick, "size_change", p, oldSize, newSize)

		metadata := map[string]interface{}{
			"reason":     reason,
			"magnitude":  newSize - oldSize,
			"plant_type": p.Type,
		}
		world.StatisticalReporter.LogSystemEvent(world.Tick, "plant_size_change", reason, metadata)
	}
	p.Size = newSize
}

// LogPlantDeath logs when a plant dies with context about the cause
func (p *Plant) LogPlantDeath(world *World, cause string, contributingFactors map[string]interface{}) {
	metadata := map[string]interface{}{
		"cause":      cause,
		"age":        p.Age,
		"energy":     p.Energy,
		"size":       p.Size,
		"plant_type": p.Type,
		"factors":    contributingFactors,
	}

	// Log to central event bus
	if world.CentralEventBus != nil {
		plantTypeName := GetPlantConfigs()[p.Type].Name
		world.CentralEventBus.EmitPlantEvent(world.Tick, "death", "death", "plant_lifecycle",
			fmt.Sprintf("Plant %d (%s) died: %s", p.ID, plantTypeName, cause), p, true, false)
	}

	// Legacy statistical reporter logging
	if world.StatisticalReporter != nil {
		world.StatisticalReporter.LogPlantEvent(world.Tick, "plant_death", p, true, false)
		world.StatisticalReporter.LogSystemEvent(world.Tick, "plant_death", cause, metadata)
	}
	p.IsAlive = false
}

// LogPlantBirth logs when a plant is born/reproduced with parental information
func (p *Plant) LogPlantBirth(world *World, parent *Plant, reproductionType string) {
	metadata := map[string]interface{}{
		"parent_id":         parent.ID,
		"parent_type":       parent.Type,
		"reproduction_type": reproductionType,
	}

	// Log to central event bus
	if world.CentralEventBus != nil {
		plantTypeName := GetPlantConfigs()[p.Type].Name
		world.CentralEventBus.EmitPlantEvent(world.Tick, "birth", "birth", "plant_lifecycle",
			fmt.Sprintf("Plant %d (%s) born from parent %d via %s", p.ID, plantTypeName, parent.ID, reproductionType), p, false, true)
	}

	// Legacy statistical reporter logging
	if world.StatisticalReporter != nil {
		world.StatisticalReporter.LogPlantEvent(world.Tick, "plant_birth", p, false, true)
		world.StatisticalReporter.LogSystemEvent(world.Tick, "plant_birth", reproductionType, metadata)
	}
}

// Enhanced Plant Nutrient System Functions

// initializeSoilNutrients creates initial soil nutrient levels
func initializeSoilNutrients() map[string]float64 {
	nutrients := make(map[string]float64)

	// Primary nutrients
	nutrients["nitrogen"] = 0.3 + rand.Float64()*0.4   // 0.3-0.7
	nutrients["phosphorus"] = 0.2 + rand.Float64()*0.3 // 0.2-0.5
	nutrients["potassium"] = 0.2 + rand.Float64()*0.3  // 0.2-0.5

	// Secondary nutrients
	nutrients["calcium"] = 0.1 + rand.Float64()*0.2   // 0.1-0.3
	nutrients["magnesium"] = 0.1 + rand.Float64()*0.2 // 0.1-0.3
	nutrients["sulfur"] = 0.1 + rand.Float64()*0.1    // 0.1-0.2

	// Micronutrients
	nutrients["iron"] = 0.05 + rand.Float64()*0.05      // 0.05-0.1
	nutrients["manganese"] = 0.02 + rand.Float64()*0.03 // 0.02-0.05
	nutrients["zinc"] = 0.01 + rand.Float64()*0.02      // 0.01-0.03

	return nutrients
}

// initializeWaterLevel sets initial water based on biome
func initializeWaterLevel(biome BiomeType) float64 {
	switch biome {
	case BiomeWater, BiomeDeepWater, BiomeSwamp:
		return 0.9 + rand.Float64()*0.1 // 0.9-1.0
	case BiomeRainforest:
		return 0.7 + rand.Float64()*0.2 // 0.7-0.9
	case BiomeForest, BiomePlains:
		return 0.4 + rand.Float64()*0.3 // 0.4-0.7
	case BiomeDesert:
		return 0.1 + rand.Float64()*0.2 // 0.1-0.3
	case BiomeIce, BiomeTundra:
		return 0.2 + rand.Float64()*0.2 // 0.2-0.4 (frozen water)
	case BiomeMountain, BiomeHighAltitude:
		return 0.3 + rand.Float64()*0.2 // 0.3-0.5
	case BiomeHotSpring:
		return 0.6 + rand.Float64()*0.3 // 0.6-0.9 (hot water)
	default:
		return 0.5 // Moderate water level
	}
}

// initializePlantNutrientNeeds sets nutrient requirements based on plant type
func initializePlantNutrientNeeds(plantType PlantType) map[string]float64 {
	needs := make(map[string]float64)

	switch plantType {
	case PlantGrass:
		// Grass needs moderate nutrients, grows fast
		needs["nitrogen"] = 0.3
		needs["phosphorus"] = 0.2
		needs["potassium"] = 0.2
		needs["water"] = 0.4

	case PlantBush:
		// Bushes need balanced nutrients
		needs["nitrogen"] = 0.4
		needs["phosphorus"] = 0.3
		needs["potassium"] = 0.3
		needs["water"] = 0.5

	case PlantTree:
		// Trees need lots of nutrients and water
		needs["nitrogen"] = 0.6
		needs["phosphorus"] = 0.4
		needs["potassium"] = 0.5
		needs["calcium"] = 0.3
		needs["water"] = 0.7

	case PlantMushroom:
		// Mushrooms thrive on organic matter, less sunlight
		needs["nitrogen"] = 0.5
		needs["phosphorus"] = 0.3
		needs["organic_matter"] = 0.6
		needs["water"] = 0.6

	case PlantAlgae:
		// Algae needs lots of water and some nutrients
		needs["nitrogen"] = 0.4
		needs["phosphorus"] = 0.4
		needs["water"] = 0.9

	case PlantCactus:
		// Cacti need minimal water, store nutrients
		needs["nitrogen"] = 0.2
		needs["phosphorus"] = 0.2
		needs["potassium"] = 0.3
		needs["water"] = 0.2

	case PlantLily:
		// Water lilies thrive in water, need aquatic nutrients
		needs["nitrogen"] = 0.3
		needs["phosphorus"] = 0.3
		needs["potassium"] = 0.2
		needs["water"] = 0.95

	case PlantReed:
		// Reeds grow in wetlands, filter water
		needs["nitrogen"] = 0.4
		needs["phosphorus"] = 0.2
		needs["potassium"] = 0.2
		needs["water"] = 0.8

	case PlantKelp:
		// Kelp grows in deep water, needs minerals
		needs["nitrogen"] = 0.5
		needs["phosphorus"] = 0.4
		needs["magnesium"] = 0.3
		needs["water"] = 1.0
	}

	return needs
}

// getPlantWaterDependency returns how much water a plant type needs
func getPlantWaterDependency(plantType PlantType) float64 {
	switch plantType {
	case PlantGrass:
		return 0.4
	case PlantBush:
		return 0.5
	case PlantTree:
		return 0.7
	case PlantMushroom:
		return 0.6
	case PlantAlgae:
		return 0.9
	case PlantCactus:
		return 0.2
	case PlantLily:
		return 0.95 // Very high water dependency
	case PlantReed:
		return 0.8 // High water dependency
	case PlantKelp:
		return 1.0 // Maximum water dependency
	default:
		return 0.5
	}
}

// getPlantRootDepth returns how deep a plant's roots grow
func getPlantRootDepth(plantType PlantType) float64 {
	switch plantType {
	case PlantGrass:
		return 0.3 // Shallow roots
	case PlantBush:
		return 0.6 // Medium roots
	case PlantTree:
		return 1.0 // Deep roots
	case PlantMushroom:
		return 0.2 // Very shallow
	case PlantAlgae:
		return 0.1 // Surface level
	case PlantCactus:
		return 0.8 // Deep tap roots for water
	case PlantLily:
		return 0.2 // Shallow water roots
	case PlantReed:
		return 0.4 // Medium depth in wetland soil
	case PlantKelp:
		return 0.0 // No soil roots, water nutrients only
	default:
		return 0.5
	}
}

// updatePlantNutrients handles realistic plant nutrition from soil, water, and decay
func (p *Plant) updatePlantNutrients(gridCell *GridCell, season string) float64 {
	if !p.IsAlive {
		return 0.0
	}

	nutritionalHealth := 1.0

	// Check water availability vs needs
	waterAvailable := gridCell.WaterLevel
	waterNeeded := p.NutrientNeeds["water"]
	waterRatio := waterAvailable / math.Max(waterNeeded, 0.1)

	if waterRatio < 0.5 {
		// Water stress
		nutritionalHealth *= 0.5 + waterRatio
		p.Energy -= 2.0 // Water stress drains energy
	} else if waterRatio > 2.0 {
		// Too much water (flooding)
		nutritionalHealth *= 0.8
	}

	// Check soil nutrients vs needs
	for nutrient, needed := range p.NutrientNeeds {
		if nutrient == "water" {
			continue // Already handled above
		}

		available := gridCell.SoilNutrients[nutrient]
		if available < needed {
			// Nutrient deficiency
			deficiencyRatio := available / needed
			nutritionalHealth *= 0.7 + deficiencyRatio*0.3

			// Consume available nutrients
			gridCell.SoilNutrients[nutrient] = math.Max(0, available-needed*0.1)
		} else {
			// Adequate nutrients, consume some
			gridCell.SoilNutrients[nutrient] = math.Max(0, available-needed*0.05)
		}
	}

	// pH effects
	pHDifference := math.Abs(gridCell.SoilPH - 7.0) // Most plants prefer neutral pH
	if pHDifference > 1.0 {
		nutritionalHealth *= 1.0 - (pHDifference-1.0)*0.2
	}

	// Soil compaction effects
	compactionPenalty := gridCell.SoilCompaction * 0.3
	nutritionalHealth *= 1.0 - compactionPenalty

	// Organic matter bonus
	organicBonus := gridCell.OrganicMatter * 0.5
	nutritionalHealth += organicBonus

	// Seasonal effects
	seasonalMultiplier := 1.0
	switch season {
	case SeasonSpring:
		seasonalMultiplier = 1.2 // Growing season
	case "summer":
		seasonalMultiplier = 1.1 // Good growth but water stress possible
	case "autumn":
		seasonalMultiplier = 0.9 // Slowing down
	case SeasonWinter:
		seasonalMultiplier = 0.6 // Dormant season
	}

	nutritionalHealth *= seasonalMultiplier

	// Apply nutritional health to plant growth
	if nutritionalHealth > 1.0 {
		// Optimal conditions - bonus growth
		p.Energy += (nutritionalHealth - 1.0) * 5.0
		p.Size += (nutritionalHealth - 1.0) * 0.1
	} else if nutritionalHealth < 0.7 {
		// Poor conditions - stress
		p.Energy -= (0.7 - nutritionalHealth) * 3.0
	}

	// Update plant's recorded nutrient levels for decision making
	p.SoilNutrients = make(map[string]float64)
	for k, v := range gridCell.SoilNutrients {
		p.SoilNutrients[k] = v
	}
	p.WaterLevel = gridCell.WaterLevel
	p.SoilPH = gridCell.SoilPH

	return nutritionalHealth
}

// addDecayNutrientsToSoil adds nutrients from decaying organic matter to soil
func addDecayNutrientsToSoil(gridCell *GridCell, decayItem *DecayableItem) {
	if gridCell.SoilNutrients == nil {
		gridCell.SoilNutrients = initializeSoilNutrients()
	}

	// Different decay sources provide different nutrients
	nutrientContribution := decayItem.NutrientValue * 0.01 // Convert to soil nutrient levels

	switch decayItem.ItemType {
	case "corpse":
		// Animal corpses provide lots of nitrogen and phosphorus
		gridCell.SoilNutrients["nitrogen"] += nutrientContribution * 2.0
		gridCell.SoilNutrients["phosphorus"] += nutrientContribution * 1.5
		gridCell.SoilNutrients["potassium"] += nutrientContribution * 0.5
		gridCell.OrganicMatter += nutrientContribution * 0.5

	case "plant_matter":
		// Dead plants provide balanced nutrients
		gridCell.SoilNutrients["nitrogen"] += nutrientContribution * 1.0
		gridCell.SoilNutrients["phosphorus"] += nutrientContribution * 0.5
		gridCell.SoilNutrients["potassium"] += nutrientContribution * 1.0
		gridCell.SoilNutrients["calcium"] += nutrientContribution * 0.3
		gridCell.OrganicMatter += nutrientContribution * 0.8

	case "organic_matter":
		// General organic matter
		gridCell.SoilNutrients["nitrogen"] += nutrientContribution * 1.0
		gridCell.SoilNutrients["phosphorus"] += nutrientContribution * 0.8
		gridCell.SoilNutrients["potassium"] += nutrientContribution * 0.8
		gridCell.OrganicMatter += nutrientContribution * 0.6
	}

	// Decay also affects soil pH towards neutral
	if gridCell.SoilPH < 7.0 {
		gridCell.SoilPH += 0.02 // Slight alkalizing effect
	} else if gridCell.SoilPH > 7.0 {
		gridCell.SoilPH -= 0.02 // Slight acidifying effect
	}

	// Organic matter helps reduce soil compaction
	gridCell.SoilCompaction *= 0.98
}

// processRainfall adds water and nutrients from precipitation
func processRainfall(gridCell *GridCell, intensity float64) {
	// Rain adds water
	gridCell.WaterLevel = math.Min(1.0, gridCell.WaterLevel+intensity*0.3)

	// Rain brings some nutrients (nitrogen from atmosphere)
	if gridCell.SoilNutrients == nil {
		gridCell.SoilNutrients = initializeSoilNutrients()
	}

	// Nitrogen fixation from rain
	gridCell.SoilNutrients["nitrogen"] += intensity * 0.01

	// Heavy rain can wash away nutrients (leaching)
	if intensity > 0.7 {
		leachingRate := (intensity - 0.7) * 0.1
		for nutrient := range gridCell.SoilNutrients {
			gridCell.SoilNutrients[nutrient] *= (1.0 - leachingRate)
		}
	}

	// Rain slightly reduces soil compaction
	gridCell.SoilCompaction *= (1.0 - intensity*0.05)
}
