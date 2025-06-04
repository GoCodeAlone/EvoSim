package main

import (
	"math"
	"math/rand"
)

// LifeStage represents the current developmental stage of an entity
type LifeStage int

const (
	StageEgg LifeStage = iota
	StageLarva
	StagePupa  // Metamorphosis stage
	StageAdult
	StageElder // Post-reproductive stage for some species
)

// String returns the string representation of LifeStage
func (ls LifeStage) String() string {
	switch ls {
	case StageEgg:
		return "egg"
	case StageLarva:
		return "larva"
	case StagePupa:
		return "pupa"
	case StageAdult:
		return "adult"
	case StageElder:
		return "elder"
	default:
		return "unknown"
	}
}

// MetamorphosisType represents different transformation strategies
type MetamorphosisType int

const (
	NoMetamorphosis     MetamorphosisType = iota // Direct development (most entities)
	SimpleMetamorphosis                          // Egg -> Larva -> Adult (gradual change)
	CompleteMetamorphosis                        // Egg -> Larva -> Pupa -> Adult (complete transformation)
	HolometabolousMetamorphosis                  // Complex transformation with multiple pupal substages
)

// String returns the string representation of MetamorphosisType
func (mt MetamorphosisType) String() string {
	switch mt {
	case NoMetamorphosis:
		return "none"
	case SimpleMetamorphosis:
		return "simple"
	case CompleteMetamorphosis:
		return "complete"
	case HolometabolousMetamorphosis:
		return "holometabolous"
	default:
		return "unknown"
	}
}

// MetamorphosisStatus tracks an entity's developmental progress
type MetamorphosisStatus struct {
	Type                  MetamorphosisType `json:"type"`
	CurrentStage          LifeStage         `json:"current_stage"`
	StageProgress         float64           `json:"stage_progress"`         // 0.0 to 1.0 progress through current stage
	StageStartTick        int               `json:"stage_start_tick"`       // When this stage began
	StageEnergy           float64           `json:"stage_energy"`           // Energy accumulated for metamorphosis
	RequiredEnergy        float64           `json:"required_energy"`        // Energy needed to advance to next stage
	EnvironmentalTriggers map[string]float64 `json:"environmental_triggers"` // Environmental factors affecting metamorphosis
	IsMetamorphosing      bool              `json:"is_metamorphosing"`      // Currently undergoing transformation
	PupalShelter          bool              `json:"pupal_shelter"`          // Has protective shelter during pupa stage
	CanMove               bool              `json:"can_move"`               // Whether entity can move in current stage
	VulnerabilityModifier float64           `json:"vulnerability_modifier"`  // Stage-specific vulnerability (1.0 = normal)
}

// MetamorphosisSystem manages life stage transitions and development
type MetamorphosisSystem struct {
	StageMinDurations   map[LifeStage]int            `json:"stage_min_durations"`   // Minimum ticks per stage
	StageEnergyRequired map[LifeStage]float64        `json:"stage_energy_required"` // Energy thresholds for advancement
	StageTraitModifiers map[LifeStage]map[string]float64 `json:"stage_trait_modifiers"` // How traits change per stage
	EnvironmentModifiers map[string]float64          `json:"environment_modifiers"` // Environmental effects on development
	SeasonalModifiers   map[string]float64          `json:"seasonal_modifiers"`    // Seasonal effects on metamorphosis
}

// NewMetamorphosisSystem creates a new metamorphosis management system
func NewMetamorphosisSystem() *MetamorphosisSystem {
	ms := &MetamorphosisSystem{
		StageMinDurations:   make(map[LifeStage]int),
		StageEnergyRequired: make(map[LifeStage]float64),
		StageTraitModifiers: make(map[LifeStage]map[string]float64),
		EnvironmentModifiers: make(map[string]float64),
		SeasonalModifiers:   make(map[string]float64),
	}
	
	ms.initializeStageRequirements()
	ms.initializeTraitModifiers()
	ms.initializeEnvironmentalFactors()
	
	return ms
}

// initializeStageRequirements sets up duration and energy requirements for each stage
func (ms *MetamorphosisSystem) initializeStageRequirements() {
	// Minimum stage durations (in ticks)
	ms.StageMinDurations[StageEgg] = 240      // ~12 hours
	ms.StageMinDurations[StageLarva] = 960    // ~2 days (growth phase)
	ms.StageMinDurations[StagePupa] = 480     // ~1 day (transformation)
	ms.StageMinDurations[StageAdult] = 2400   // ~5 days (reproductive phase)
	ms.StageMinDurations[StageElder] = 1200   // ~2.5 days (post-reproductive)
	
	// Energy requirements to advance to next stage
	ms.StageEnergyRequired[StageEgg] = 50.0    // Energy to hatch
	ms.StageEnergyRequired[StageLarva] = 150.0 // Energy to pupate
	ms.StageEnergyRequired[StagePupa] = 100.0  // Energy to emerge as adult
	ms.StageEnergyRequired[StageAdult] = 200.0 // Energy to become elder
	ms.StageEnergyRequired[StageElder] = 0.0   // Final stage
}

// initializeTraitModifiers sets up how traits change at each life stage
func (ms *MetamorphosisSystem) initializeTraitModifiers() {
	// Larva stage - focused on growth and feeding
	ms.StageTraitModifiers[StageLarva] = map[string]float64{
		"size":              0.3,  // Smaller during larval stage
		"speed":             0.6,  // Slower movement
		"vision":            0.7,  // Reduced vision
		"endurance":         1.2,  // Higher endurance for growth
		"aggression":        0.4,  // Less aggressive
		"cooperation":       0.3,  // Limited cooperation
		"intelligence":      0.5,  // Lower intelligence
		"reproduction_rate": 0.0,  // Cannot reproduce
		"energy_efficiency": 1.3,  // More efficient at converting food to energy
		"feeding_frequency": 1.5,  // Need to feed more often
	}
	
	// Pupa stage - immobile transformation
	ms.StageTraitModifiers[StagePupa] = map[string]float64{
		"size":              0.4,  // Compact form
		"speed":             0.0,  // Cannot move
		"vision":            0.1,  // Minimal vision
		"endurance":         0.8,  // Vulnerable to environmental stress
		"aggression":        0.0,  // Cannot fight
		"cooperation":       0.0,  // Cannot interact
		"intelligence":      0.2,  // Minimal awareness
		"reproduction_rate": 0.0,  // Cannot reproduce
		"energy_efficiency": 0.5,  // Energy used for transformation
		"feeding_frequency": 0.0,  // Does not feed
	}
	
	// Adult stage - fully developed capabilities
	ms.StageTraitModifiers[StageAdult] = map[string]float64{
		"size":              1.0,  // Full size
		"speed":             1.0,  // Full speed
		"vision":            1.0,  // Full vision
		"endurance":         1.0,  // Normal endurance
		"aggression":        1.0,  // Full aggression
		"cooperation":       1.0,  // Full cooperation
		"intelligence":      1.0,  // Full intelligence
		"reproduction_rate": 1.2,  // Enhanced reproduction
		"energy_efficiency": 1.0,  // Normal efficiency
		"feeding_frequency": 1.0,  // Normal feeding
	}
	
	// Elder stage - declining but experienced
	ms.StageTraitModifiers[StageElder] = map[string]float64{
		"size":              0.9,  // Slightly smaller
		"speed":             0.7,  // Slower
		"vision":            0.8,  // Declining vision
		"endurance":         0.6,  // Lower endurance
		"aggression":        0.8,  // Less aggressive
		"cooperation":       1.3,  // More cooperative (wisdom)
		"intelligence":      1.2,  // Higher intelligence (experience)
		"reproduction_rate": 0.2,  // Minimal reproduction
		"energy_efficiency": 0.8,  // Less efficient
		"feeding_frequency": 0.8,  // Less appetite
	}
}

// initializeEnvironmentalFactors sets up environmental influences on metamorphosis
func (ms *MetamorphosisSystem) initializeEnvironmentalFactors() {
	// Environmental modifiers (how environment affects development speed)
	ms.EnvironmentModifiers["temperature"] = 1.0 // Will be modified by actual temperature
	ms.EnvironmentModifiers["humidity"] = 1.0
	ms.EnvironmentModifiers["food_availability"] = 1.0
	ms.EnvironmentModifiers["safety"] = 1.0
	ms.EnvironmentModifiers["population_density"] = 1.0
	
	// Seasonal modifiers
	ms.SeasonalModifiers["spring"] = 1.2  // Faster development
	ms.SeasonalModifiers["summer"] = 1.0  // Normal development
	ms.SeasonalModifiers["autumn"] = 0.8  // Slower development
	ms.SeasonalModifiers["winter"] = 0.5  // Very slow development (diapause-like)
}

// NewMetamorphosisStatus creates initial metamorphosis status for an entity
func NewMetamorphosisStatus(entity *Entity, metamorphosisSystem *MetamorphosisSystem) *MetamorphosisStatus {
	// Determine metamorphosis type based on entity traits
	metamorphosisType := metamorphosisSystem.determineMetamorphosisType(entity)
	
	status := &MetamorphosisStatus{
		Type:                  metamorphosisType,
		CurrentStage:          StageEgg,
		StageProgress:         0.0,
		StageStartTick:        0,
		StageEnergy:           entity.Energy,
		RequiredEnergy:        0.0,
		EnvironmentalTriggers: make(map[string]float64),
		IsMetamorphosing:      false,
		PupalShelter:          false,
		CanMove:               false, // Eggs cannot move
		VulnerabilityModifier: 1.5,   // Eggs are more vulnerable
	}
	
	// If entity doesn't undergo metamorphosis, start as adult
	if metamorphosisType == NoMetamorphosis {
		status.CurrentStage = StageAdult
		status.CanMove = true
		status.VulnerabilityModifier = 1.0
	}
	
	return status
}

// determineMetamorphosisType determines what type of metamorphosis an entity undergoes
func (ms *MetamorphosisSystem) determineMetamorphosisType(entity *Entity) MetamorphosisType {
	// Check if entity has insect-like traits
	swarmCapability := entity.GetTrait("swarm_capability")
	pollinationEfficiency := entity.GetTrait("pollination_efficiency")
	size := entity.GetTrait("size")
	intelligence := entity.GetTrait("intelligence")
	
	// Small, swarming entities with pollination abilities undergo complete metamorphosis
	if size < -0.1 && swarmCapability > 0.3 && pollinationEfficiency > 0.0 {
		if intelligence > 0.6 {
			return HolometabolousMetamorphosis // Most complex
		}
		return CompleteMetamorphosis
	}
	
	// Medium-sized social entities may undergo simple metamorphosis
	if size < 0.2 && swarmCapability > 0.2 {
		return SimpleMetamorphosis
	}
	
	// Most entities develop directly
	return NoMetamorphosis
}

// Update processes metamorphosis for an entity during a world tick
func (ms *MetamorphosisSystem) Update(entity *Entity, currentTick int, environment map[string]float64) bool {
	if entity.MetamorphosisStatus == nil {
		// Initialize metamorphosis status if missing
		entity.MetamorphosisStatus = NewMetamorphosisStatus(entity, ms)
		ms.applyStageModifiers(entity)
		return false
	}
	
	status := entity.MetamorphosisStatus
	
	// Skip processing if no metamorphosis or already adult/elder
	if status.Type == NoMetamorphosis || status.CurrentStage == StageElder {
		return false
	}
	
	// Calculate development progress
	stageChanged := ms.processStageProgression(entity, currentTick, environment)
	
	// Apply current stage modifiers
	if stageChanged {
		ms.applyStageModifiers(entity)
	}
	
	return stageChanged
}

// processStageProgression handles advancement through life stages
func (ms *MetamorphosisSystem) processStageProgression(entity *Entity, currentTick int, environment map[string]float64) bool {
	status := entity.MetamorphosisStatus
	ticksInStage := currentTick - status.StageStartTick
	minDuration := ms.StageMinDurations[status.CurrentStage]
	
	// Check if minimum time requirement is met
	if ticksInStage < minDuration {
		status.StageProgress = float64(ticksInStage) / float64(minDuration)
		return false
	}
	
	// Check energy requirements
	requiredEnergy := ms.StageEnergyRequired[status.CurrentStage]
	if requiredEnergy > 0 && entity.Energy < requiredEnergy {
		// Not enough energy to advance
		status.StageProgress = math.Min(1.0, entity.Energy/requiredEnergy)
		return false
	}
	
	// Check environmental triggers
	if !ms.checkEnvironmentalTriggers(entity, environment) {
		return false
	}
	
	// Ready to advance to next stage
	return ms.advanceToNextStage(entity, currentTick)
}

// checkEnvironmentalTriggers verifies environmental conditions are suitable for metamorphosis
func (ms *MetamorphosisSystem) checkEnvironmentalTriggers(entity *Entity, environment map[string]float64) bool {
	status := entity.MetamorphosisStatus
	
	// Different stages have different environmental requirements
	switch status.CurrentStage {
	case StageEgg:
		// Eggs need stable temperature and humidity
		temp := environment["temperature"]
		humidity := environment["humidity"]
		return temp > 0.3 && temp < 0.9 && humidity > 0.4
		
	case StageLarva:
		// Larvae need food availability
		foodAvailability := environment["food_availability"]
		return foodAvailability > 0.3
		
	case StagePupa:
		// Pupae need safe environment and stable conditions
		safety := environment["safety"]
		temp := environment["temperature"]
		return safety > 0.4 && temp > 0.2 && temp < 0.8
		
	case StageAdult:
		// Adults can transition to elder based on age and stress
		return entity.Age > ms.StageMinDurations[StageAdult]*3
	}
	
	return true
}

// advanceToNextStage moves entity to the next developmental stage
func (ms *MetamorphosisSystem) advanceToNextStage(entity *Entity, currentTick int) bool {
	status := entity.MetamorphosisStatus
	currentStage := status.CurrentStage
	
	// Determine next stage based on metamorphosis type
	var nextStage LifeStage
	switch status.Type {
	case SimpleMetamorphosis:
		switch currentStage {
		case StageEgg:
			nextStage = StageLarva
		case StageLarva:
			nextStage = StageAdult
		case StageAdult:
			nextStage = StageElder
		default:
			return false
		}
		
	case CompleteMetamorphosis, HolometabolousMetamorphosis:
		switch currentStage {
		case StageEgg:
			nextStage = StageLarva
		case StageLarva:
			nextStage = StagePupa
		case StagePupa:
			nextStage = StageAdult
		case StageAdult:
			nextStage = StageElder
		default:
			return false
		}
		
	default:
		return false
	}
	
	// Consume energy for transformation
	energyCost := ms.StageEnergyRequired[currentStage]
	entity.Energy -= energyCost
	
	// Update status
	status.CurrentStage = nextStage
	status.StageStartTick = currentTick
	status.StageProgress = 0.0
	status.IsMetamorphosing = (nextStage == StagePupa)
	
	// Update movement and vulnerability based on new stage
	ms.updateStageCapabilities(entity, nextStage)
	
	return true
}

// updateStageCapabilities updates entity capabilities based on life stage
func (ms *MetamorphosisSystem) updateStageCapabilities(entity *Entity, stage LifeStage) {
	status := entity.MetamorphosisStatus
	
	switch stage {
	case StageEgg:
		status.CanMove = false
		status.VulnerabilityModifier = 1.5
		
	case StageLarva:
		status.CanMove = true
		status.VulnerabilityModifier = 1.2
		
	case StagePupa:
		status.CanMove = false
		status.VulnerabilityModifier = 2.0 // Very vulnerable during transformation
		status.PupalShelter = rand.Float64() < 0.7 // 70% chance of creating protective shelter
		
	case StageAdult:
		status.CanMove = true
		status.VulnerabilityModifier = 1.0
		status.IsMetamorphosing = false
		status.PupalShelter = false
		
	case StageElder:
		status.CanMove = true
		status.VulnerabilityModifier = 1.3 // More vulnerable due to age
	}
}

// applyStageModifiers applies trait modifications based on current life stage
func (ms *MetamorphosisSystem) applyStageModifiers(entity *Entity) {
	if entity.MetamorphosisStatus == nil {
		return
	}
	
	stage := entity.MetamorphosisStatus.CurrentStage
	modifiers := ms.StageTraitModifiers[stage]
	
	if modifiers == nil {
		return
	}
	
	// Store original trait values if not already stored
	if entity.OriginalTraits == nil {
		entity.OriginalTraits = make(map[string]float64)
		for traitName := range entity.Traits {
			entity.OriginalTraits[traitName] = entity.GetTrait(traitName)
		}
	}
	
	// Apply stage-specific modifiers
	for traitName, modifier := range modifiers {
		if originalValue, exists := entity.OriginalTraits[traitName]; exists {
			newValue := originalValue * modifier
			entity.SetTrait(traitName, newValue)
		}
	}
}

// GetStageDescription returns a human-readable description of the entity's developmental stage
func (ms *MetamorphosisSystem) GetStageDescription(entity *Entity) string {
	if entity.MetamorphosisStatus == nil {
		return "Unknown developmental stage"
	}
	
	status := entity.MetamorphosisStatus
	
	baseDescription := status.CurrentStage.String()
	if status.IsMetamorphosing {
		baseDescription += " (metamorphosing)"
	}
	
	if status.CurrentStage == StagePupa && status.PupalShelter {
		baseDescription += " (sheltered)"
	}
	
	return baseDescription
}

// GetMetamorphosisStats returns statistics about metamorphosis in the population
func (ms *MetamorphosisSystem) GetMetamorphosisStats(entities []*Entity) map[string]interface{} {
	stats := make(map[string]interface{})
	
	stageCounts := make(map[LifeStage]int)
	typeCounts := make(map[MetamorphosisType]int)
	metamorphosisCount := 0
	shelterCount := 0
	
	for _, entity := range entities {
		if entity == nil || !entity.IsAlive || entity.MetamorphosisStatus == nil {
			continue
		}
		
		status := entity.MetamorphosisStatus
		stageCounts[status.CurrentStage]++
		typeCounts[status.Type]++
		
		if status.IsMetamorphosing {
			metamorphosisCount++
		}
		
		if status.PupalShelter {
			shelterCount++
		}
	}
	
	stats["stage_counts"] = stageCounts
	stats["type_counts"] = typeCounts
	stats["currently_metamorphosing"] = metamorphosisCount
	stats["pupal_shelters"] = shelterCount
	
	return stats
}