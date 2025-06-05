package main

import (
	"math"
	"math/rand"
)

// AirborneChemicalSignal represents airborne chemical signals (distinct from underground network signals)
type AirborneChemicalSignal struct {
	ID           int                    `json:"id"`
	Type         string                 `json:"type"` // "toxin", "attractant", "warning", "scent"
	Position     Position               `json:"position"`
	Strength     float64                `json:"strength"`
	MaxRange     float64                `json:"max_range"`
	DecayRate    float64                `json:"decay_rate"`
	Age          int                    `json:"age"`
	MaxAge       int                    `json:"max_age"`
	ProducerID   int                    `json:"producer_id"`
	ProducerType string                 `json:"producer_type"` // "plant" or "entity"
	Data         map[string]interface{} `json:"data"`
	IsAirborne   bool                   `json:"is_airborne"` // Always true for this type
}

// ChemicalEcologySystem manages airborne chemical interactions that complement the underground network system
type ChemicalEcologySystem struct {
	AirborneSignals         []*AirborneChemicalSignal  `json:"airborne_signals"`
	NextSignalID            int                        `json:"next_signal_id"`
	ToxinResistance         map[int]float64            `json:"toxin_resistance"` // Entity ID -> resistance level
	ChemicalMemory          map[int]map[string]float64 `json:"chemical_memory"`  // Entity memory of chemical signals
	PlantDefenseActivations int                        `json:"plant_defense_activations"`
	AttractantEffects       int                        `json:"attractant_effects"`
	ToxinEffects            int                        `json:"toxin_effects"`
	WarningEffects          int                        `json:"warning_effects"`
}

// NewChemicalEcologySystem creates a new chemical ecology system
func NewChemicalEcologySystem() *ChemicalEcologySystem {
	return &ChemicalEcologySystem{
		AirborneSignals:         make([]*AirborneChemicalSignal, 0),
		NextSignalID:            1,
		ToxinResistance:         make(map[int]float64),
		ChemicalMemory:          make(map[int]map[string]float64),
		PlantDefenseActivations: 0,
		AttractantEffects:       0,
		ToxinEffects:            0,
		WarningEffects:          0,
	}
}

// EmitPlantAirborneSignal allows plants to release airborne chemical signals
func (ces *ChemicalEcologySystem) EmitPlantAirborneSignal(plant *Plant, signalType string, world *World) {
	// Check if plant can produce chemical signals
	toxinProduction := plant.GetTrait("toxin_production")

	var strength float64
	var maxRange float64
	var decayRate float64
	var maxAge int

	switch signalType {
	case "toxin":
		if toxinProduction < 0.3 {
			return // Not toxic enough
		}
		strength = toxinProduction * plant.Size
		maxRange = 8.0 + plant.Size*4.0
		decayRate = 0.1
		maxAge = 20
		ces.PlantDefenseActivations++

	case "attractant":
		// Plants with high nutrition might attract beneficial entities
		nutritionDensity := plant.GetTrait("nutrition_density")
		if nutritionDensity < 0.2 {
			return
		}
		strength = nutritionDensity * plant.Energy / 100.0
		maxRange = 12.0 + nutritionDensity*8.0
		decayRate = 0.15
		maxAge = 15
		ces.AttractantEffects++

	case "warning":
		// Plants under stress warn nearby plants
		if plant.Energy > plant.Size*20 { // Not stressed
			return
		}
		strength = (1.0 - plant.Energy/(plant.Size*30)) * plant.Size
		maxRange = 15.0 + plant.Size*5.0
		decayRate = 0.2
		maxAge = 25
		ces.WarningEffects++

	default:
		return
	}

	// Create the airborne chemical signal
	signal := &AirborneChemicalSignal{
		ID:           ces.NextSignalID,
		Type:         signalType,
		Position:     plant.Position,
		Strength:     strength,
		MaxRange:     maxRange,
		DecayRate:    decayRate,
		Age:          0,
		MaxAge:       maxAge,
		ProducerID:   plant.ID,
		ProducerType: "plant",
		Data:         make(map[string]interface{}),
		IsAirborne:   true,
	}

	// Add signal-specific data
	switch signalType {
	case "toxin":
		signal.Data["toxin_type"] = plant.Type
		signal.Data["toxicity_level"] = plant.Toxicity
	case "attractant":
		signal.Data["nutrition_value"] = plant.NutritionVal
		signal.Data["plant_type"] = plant.Type
	case "warning":
		signal.Data["threat_level"] = strength
		signal.Data["plant_health"] = plant.Energy
	}

	ces.NextSignalID++
	ces.AirborneSignals = append(ces.AirborneSignals, signal)
}

// EmitEntityAirborneSignal allows entities to release airborne chemical signals
func (ces *ChemicalEcologySystem) EmitEntityAirborneSignal(entity *Entity, signalType string, world *World) {
	// Only certain signal types make sense for entities
	if signalType != "scent" {
		return
	}

	var strength float64
	var maxRange float64
	var decayRate float64
	var maxAge int

	switch signalType {
	case "scent":
		entitySize := entity.GetTrait("size")
		strength = entitySize + entity.GetTrait("strength")*0.5
		maxRange = 5.0 + entitySize*2.0
		decayRate = 0.05
		maxAge = 50
	}

	signal := &AirborneChemicalSignal{
		ID:           ces.NextSignalID,
		Type:         signalType,
		Position:     entity.Position,
		Strength:     strength,
		MaxRange:     maxRange,
		DecayRate:    decayRate,
		Age:          0,
		MaxAge:       maxAge,
		ProducerID:   entity.ID,
		ProducerType: "entity",
		Data:         make(map[string]interface{}),
		IsAirborne:   true,
	}

	signal.Data["entity_species"] = entity.Species
	signal.Data["entity_size"] = entity.GetTrait("size")

	ces.NextSignalID++
	ces.AirborneSignals = append(ces.AirborneSignals, signal)
}

// Update updates all airborne chemical signals and handles interactions
func (ces *ChemicalEcologySystem) Update(world *World) {
	// Update existing signals
	for i := len(ces.AirborneSignals) - 1; i >= 0; i-- {
		signal := ces.AirborneSignals[i]
		ces.updateSignal(signal, world)

		// Remove expired signals
		if signal.Age > signal.MaxAge || signal.Strength <= 0.01 {
			ces.AirborneSignals = append(ces.AirborneSignals[:i], ces.AirborneSignals[i+1:]...)
		}
	}

	// Process plant chemical responses
	ces.processPlantChemicalResponses(world)

	// Process entity chemical responses
	ces.processEntityChemicalResponses(world)
}

// updateSignal updates individual signal properties
func (ces *ChemicalEcologySystem) updateSignal(signal *AirborneChemicalSignal, world *World) {
	signal.Age++
	signal.Strength *= (1.0 - signal.DecayRate)

	// Airborne signals are affected by wind
	if world.WindSystem != nil && len(world.WindSystem.WindMap) > 0 {
		// Move signal with wind (simplified)
		mapWidth := len(world.WindSystem.WindMap)
		mapHeight := len(world.WindSystem.WindMap[0])

		gridX := int(signal.Position.X)
		if gridX < 0 {
			gridX = 0
		} else if gridX >= mapWidth {
			gridX = mapWidth - 1
		}

		gridY := int(signal.Position.Y)
		if gridY < 0 {
			gridY = 0
		} else if gridY >= mapHeight {
			gridY = mapHeight - 1
		}

		wind := world.WindSystem.WindMap[gridX][gridY]
		signal.Position.X += wind.X * 0.2
		signal.Position.Y += wind.Y * 0.2

		// Wind disperses airborne signals faster
		signal.Strength *= (1.0 - wind.Strength*0.1)
	}
}

// processPlantChemicalResponses handles how plants respond to airborne chemical signals
func (ces *ChemicalEcologySystem) processPlantChemicalResponses(world *World) {
	for _, plant := range world.AllPlants {
		if !plant.IsAlive {
			continue
		}

		// Check for nearby chemical signals
		for _, signal := range ces.AirborneSignals {
			distance := math.Sqrt(math.Pow(plant.Position.X-signal.Position.X, 2) +
				math.Pow(plant.Position.Y-signal.Position.Y, 2))

			if distance <= signal.MaxRange {
				ces.applyChemicalEffectToPlant(plant, signal, distance, world)
			}
		}
	}
}

// processEntityChemicalResponses handles how entities respond to airborne chemical signals
func (ces *ChemicalEcologySystem) processEntityChemicalResponses(world *World) {
	for _, entity := range world.AllEntities {
		if !entity.IsAlive {
			continue
		}

		// Initialize entity chemical memory if needed
		if ces.ChemicalMemory[entity.ID] == nil {
			ces.ChemicalMemory[entity.ID] = make(map[string]float64)
		}

		// Check for nearby chemical signals
		for _, signal := range ces.AirborneSignals {
			distance := math.Sqrt(math.Pow(entity.Position.X-signal.Position.X, 2) +
				math.Pow(entity.Position.Y-signal.Position.Y, 2))

			if distance <= signal.MaxRange {
				ces.applyChemicalEffectToEntity(entity, signal, distance, world)
			}
		}
	}
}

// applyChemicalEffectToPlant applies airborne chemical signal effects to plants
func (ces *ChemicalEcologySystem) applyChemicalEffectToPlant(plant *Plant, signal *AirborneChemicalSignal, distance float64, world *World) {
	effectStrength := signal.Strength * (1.0 - distance/signal.MaxRange)

	switch signal.Type {
	case "warning":
		// Plants respond to warning signals by increasing defenses
		if rand.Float64() < effectStrength*0.3 {
			ces.EmitPlantAirborneSignal(plant, "toxin", world)
		}
	}
}

// applyChemicalEffectToEntity applies airborne chemical signal effects to entities
func (ces *ChemicalEcologySystem) applyChemicalEffectToEntity(entity *Entity, signal *AirborneChemicalSignal, distance float64, world *World) {
	effectStrength := signal.Strength * (1.0 - distance/signal.MaxRange)

	// Update chemical memory
	ces.ChemicalMemory[entity.ID][signal.Type] = math.Max(
		ces.ChemicalMemory[entity.ID][signal.Type], effectStrength)

	switch signal.Type {
	case "toxin":
		// Check entity toxin resistance
		resistance := ces.ToxinResistance[entity.ID]
		if resistance == 0 {
			// Initialize resistance based on entity traits
			hardiness := entity.GetTrait("hardiness")
			resistance = 0.1 + hardiness*0.3
			ces.ToxinResistance[entity.ID] = resistance
		}

		// Apply toxin effects
		if effectStrength > resistance {
			damage := (effectStrength - resistance) * 5.0
			entity.Energy -= damage
			ces.ToxinEffects++

			// Entities can develop resistance over time
			ces.ToxinResistance[entity.ID] += 0.01
		}

	case "attractant":
		// Attractant signals draw entities toward plants
		if effectStrength > 0.3 {
			// Move entity toward signal source
			dx := signal.Position.X - entity.Position.X
			dy := signal.Position.Y - entity.Position.Y
			moveDistance := effectStrength * 0.5

			if distance > 0.1 {
				entity.Position.X += (dx / distance) * moveDistance
				entity.Position.Y += (dy / distance) * moveDistance
			}
			ces.AttractantEffects++
		}

	case "warning":
		// Warning signals might repel some entities
		defense := entity.GetTrait("defense")
		if defense < 0.5 && effectStrength > 0.4 {
			// Move away from warning source
			dx := entity.Position.X - signal.Position.X
			dy := entity.Position.Y - signal.Position.Y
			moveDistance := effectStrength * 0.3

			if distance > 0.1 {
				entity.Position.X += (dx / distance) * moveDistance
				entity.Position.Y += (dy / distance) * moveDistance
			}
			ces.WarningEffects++
		}

	case "scent":
		// Entity responses to animal scent depends on species and traits
		aggression := entity.GetTrait("aggression")
		cooperation := entity.GetTrait("cooperation")

		if signal.ProducerID != entity.ID {
			if aggression > 0.6 && cooperation < 0.4 {
				// Aggressive entities might be attracted to challenge
				dx := signal.Position.X - entity.Position.X
				dy := signal.Position.Y - entity.Position.Y
				moveDistance := effectStrength * 0.2

				if distance > 0.1 {
					entity.Position.X += (dx / distance) * moveDistance
					entity.Position.Y += (dy / distance) * moveDistance
				}
			} else if cooperation > 0.6 {
				// Cooperative entities might investigate
				dx := signal.Position.X - entity.Position.X
				dy := signal.Position.Y - entity.Position.Y
				moveDistance := effectStrength * 0.1

				if distance > 0.1 {
					entity.Position.X += (dx / distance) * moveDistance
					entity.Position.Y += (dy / distance) * moveDistance
				}
			}
		}
	}
}

// TriggerPlantDefenses triggers defensive chemical releases based on threats
func (ces *ChemicalEcologySystem) TriggerPlantDefenses(plant *Plant, threat *Entity, world *World) {
	// Plants might release toxins when being consumed
	distance := math.Sqrt(math.Pow(threat.Position.X-plant.Position.X, 2) +
		math.Pow(threat.Position.Y-plant.Position.Y, 2))

	if distance < 2.0 { // Close enough to be threatening
		ces.EmitPlantAirborneSignal(plant, "toxin", world)

		// Also emit warning signal to nearby plants
		if rand.Float64() < 0.6 {
			ces.EmitPlantAirborneSignal(plant, "warning", world)
		}
	}
}

// GetChemicalEnvironment returns chemical signal information at a position
func (ces *ChemicalEcologySystem) GetChemicalEnvironment(pos Position) map[string]float64 {
	environment := make(map[string]float64)

	for _, signal := range ces.AirborneSignals {
		distance := math.Sqrt(math.Pow(pos.X-signal.Position.X, 2) +
			math.Pow(pos.Y-signal.Position.Y, 2))

		if distance <= signal.MaxRange {
			effectStrength := signal.Strength * (1.0 - distance/signal.MaxRange)
			environment[signal.Type] = math.Max(environment[signal.Type], effectStrength)
		}
	}

	return environment
}

// GetStats returns chemical ecology statistics
func (ces *ChemicalEcologySystem) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count signals by type
	signalCounts := make(map[string]int)

	for _, signal := range ces.AirborneSignals {
		signalCounts[signal.Type]++
	}

	stats["total_airborne_signals"] = len(ces.AirborneSignals)
	stats["airborne_signal_counts"] = signalCounts
	stats["plant_defense_activations"] = ces.PlantDefenseActivations
	stats["attractant_effects"] = ces.AttractantEffects
	stats["toxin_effects"] = ces.ToxinEffects
	stats["warning_effects"] = ces.WarningEffects
	stats["entities_with_resistance"] = len(ces.ToxinResistance)

	return stats
}
