package main

import (
	"math"
	"math/rand"
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
