package main

import (
	"math"
	"math/rand"
	"strings"
)

// MolecularType represents different types of molecules
type MolecularType int

const (
	// Proteins - essential for growth and structure
	ProteinStructural MolecularType = iota // For building body structure
	ProteinEnzymatic                       // For metabolism and reactions
	ProteinTransport                       // For moving other molecules
	ProteinDefensive                       // For immune responses

	// Amino Acids - building blocks of proteins
	AminoEssential    // Essential amino acids that can't be synthesized
	AminoNonEssential // Non-essential amino acids
	AminoConditional  // Conditional amino acids (sometimes essential)

	// Lipids - for energy storage and membrane structure
	LipidSaturated   // Energy storage
	LipidUnsaturated // Membrane flexibility
	LipidPhospho     // Cell membrane components

	// Carbohydrates - for immediate energy
	CarboSimple  // Quick energy (sugars)
	CarboComplex // Sustained energy (starches)
	CarboFiber   // Digestive aid

	// Nucleic Acids - for genetic material and energy
	NucleicDNA // Genetic information
	NucleicRNA // Protein synthesis and regulation
	NucleicATP // Energy currency

	// Minerals and vitamins
	MineralTrace // Essential trace minerals
	MineralSalt  // Sodium and other salts (added for aquatic species)
	VitaminFat   // Fat-soluble vitamins
	VitaminWater // Water-soluble vitamins

	// Toxins and defensive compounds
	ToxinAlkaloid   // Plant defensive compounds
	ToxinGlycoside  // Plant toxins
	ToxinTannin     // Bitter defensive compounds
	ToxinHeavyMetal // Heavy metal toxins (added for soil species)
)

// MolecularComponent represents a specific molecule with concentration
type MolecularComponent struct {
	Type          MolecularType `json:"type"`
	Concentration float64       `json:"concentration"` // Amount present (0.0-1.0)
	Quality       float64       `json:"quality"`       // Quality/bioavailability (0.0-1.0)
	Freshness     float64       `json:"freshness"`     // How fresh/active (0.0-1.0, decays over time)
}

// MolecularProfile represents the complete molecular composition of an entity or food source
type MolecularProfile struct {
	Components   map[MolecularType]MolecularComponent `json:"components"`
	TotalBiomass float64                              `json:"total_biomass"` // Total molecular content
	Diversity    float64                              `json:"diversity"`     // Molecular diversity (0.0-1.0)
	Toxicity     float64                              `json:"toxicity"`      // Overall toxicity level
}

// MolecularNeeds represents what an entity requires for optimal function
type MolecularNeeds struct {
	Requirements map[MolecularType]float64 `json:"requirements"` // How much of each type needed (0.0-1.0)
	Priorities   map[MolecularType]float64 `json:"priorities"`   // Priority weights for each type
	Deficiencies map[MolecularType]float64 `json:"deficiencies"` // Current deficiency levels
	Tolerances   map[MolecularType]float64 `json:"tolerances"`   // Tolerance to toxins
}

// MolecularMetabolism represents how an entity processes molecules
type MolecularMetabolism struct {
	Efficiency         map[MolecularType]float64 `json:"efficiency"`          // How efficiently each type is processed
	StorageCapacity    map[MolecularType]float64 `json:"storage_capacity"`    // How much can be stored
	CurrentStorage     map[MolecularType]float64 `json:"current_storage"`     // Current storage levels
	ProcessingRate     float64                   `json:"processing_rate"`     // Overall processing speed
	DetoxificationRate float64                   `json:"detoxification_rate"` // Ability to neutralize toxins
}

// NewMolecularProfile creates a new molecular profile
func NewMolecularProfile() *MolecularProfile {
	return &MolecularProfile{
		Components:   make(map[MolecularType]MolecularComponent),
		TotalBiomass: 0.0,
		Diversity:    0.0,
		Toxicity:     0.0,
	}
}

// NewMolecularNeeds creates new molecular needs based on entity traits
func NewMolecularNeeds(entity *Entity) *MolecularNeeds {
	needs := &MolecularNeeds{
		Requirements: make(map[MolecularType]float64),
		Priorities:   make(map[MolecularType]float64),
		Deficiencies: make(map[MolecularType]float64),
		Tolerances:   make(map[MolecularType]float64),
	}

	// Base requirements for all entities
	baseRequirements := map[MolecularType]float64{
		ProteinStructural: 0.3,
		ProteinEnzymatic:  0.2,
		AminoEssential:    0.4,
		AminoNonEssential: 0.2,
		LipidSaturated:    0.2,
		LipidUnsaturated:  0.3,
		CarboSimple:       0.4,
		CarboComplex:      0.3,
		NucleicATP:        0.5,
		MineralTrace:      0.1,
		VitaminWater:      0.2,
	}

	// Adjust requirements based on entity traits and species
	strength := entity.GetTrait("strength")
	intelligence := entity.GetTrait("intelligence")
	speed := entity.GetTrait("speed")
	size := entity.GetTrait("size")

	// Apply species-specific nutritional dependencies
	needs.applySpeciesNutritionalRequirements(entity.Species)

	// Apply environmental adaptations to water needs
	needs.applyEnvironmentalWaterRequirements(entity)
	aggression := entity.GetTrait("aggression")

	// Stronger entities need more structural proteins
	baseRequirements[ProteinStructural] += strength * 0.2
	baseRequirements[AminoEssential] += strength * 0.1

	// More intelligent entities need more ATP and enzymatic proteins
	baseRequirements[NucleicATP] += intelligence * 0.3
	baseRequirements[ProteinEnzymatic] += intelligence * 0.2

	// Faster entities need more simple carbs and ATP
	baseRequirements[CarboSimple] += speed * 0.2
	baseRequirements[NucleicATP] += speed * 0.2

	// Larger entities need more of everything
	sizeMultiplier := 1.0 + size*0.3
	for molType := range baseRequirements {
		baseRequirements[molType] *= sizeMultiplier
	}

	// Aggressive entities have higher tolerance to toxins
	toxinTolerance := 0.1 + aggression*0.4
	needs.Tolerances[ToxinAlkaloid] = toxinTolerance
	needs.Tolerances[ToxinGlycoside] = toxinTolerance
	needs.Tolerances[ToxinTannin] = toxinTolerance

	// Set requirements and initial priorities
	for molType, requirement := range baseRequirements {
		needs.Requirements[molType] = math.Min(requirement, 1.0)
		needs.Priorities[molType] = requirement   // Higher requirement = higher priority
		needs.Deficiencies[molType] = requirement // Start with full deficiency
	}

	return needs
}

// NewMolecularMetabolism creates new molecular metabolism based on entity traits
func NewMolecularMetabolism(entity *Entity) *MolecularMetabolism {
	metabolism := &MolecularMetabolism{
		Efficiency:         make(map[MolecularType]float64),
		StorageCapacity:    make(map[MolecularType]float64),
		CurrentStorage:     make(map[MolecularType]float64),
		ProcessingRate:     0.5,
		DetoxificationRate: 0.3,
	}

	// Base processing efficiencies
	baseEfficiency := 0.6
	intelligence := entity.GetTrait("intelligence")
	size := entity.GetTrait("size")
	cooperation := entity.GetTrait("cooperation")

	// Intelligence improves processing efficiency
	metabolism.ProcessingRate = 0.5 + intelligence*0.3
	metabolism.DetoxificationRate = 0.3 + intelligence*0.2

	// Size affects storage capacity
	storageMultiplier := 1.0 + size*0.5

	// Set efficiencies and storage for all molecule types
	for molType := ProteinStructural; molType <= ToxinTannin; molType++ {
		efficiency := baseEfficiency + rand.Float64()*0.2 - 0.1 // Some variation

		// Cooperative entities are better at processing social molecules
		if molType == ProteinTransport || molType == VitaminWater {
			efficiency += cooperation * 0.2
		}

		metabolism.Efficiency[molType] = math.Min(efficiency, 1.0)
		metabolism.StorageCapacity[molType] = (0.3 + rand.Float64()*0.4) * storageMultiplier
		metabolism.CurrentStorage[molType] = 0.0
	}

	return metabolism
}

// AddComponent adds a molecular component to the profile
func (mp *MolecularProfile) AddComponent(molType MolecularType, concentration, quality, freshness float64) {
	mp.Components[molType] = MolecularComponent{
		Type:          molType,
		Concentration: math.Min(concentration, 1.0),
		Quality:       math.Min(quality, 1.0),
		Freshness:     math.Min(freshness, 1.0),
	}
	mp.updateDerivedMetrics()
}

// updateDerivedMetrics calculates derived metrics from components
func (mp *MolecularProfile) updateDerivedMetrics() {
	totalMolecules := 0.0
	totalBiomass := 0.0
	totalToxicity := 0.0
	diversityCount := 0

	for molType, component := range mp.Components {
		weightedConcentration := component.Concentration * component.Quality * component.Freshness
		totalMolecules += weightedConcentration
		totalBiomass += weightedConcentration

		// Count toxins
		if molType >= ToxinAlkaloid {
			totalToxicity += weightedConcentration
		}

		if component.Concentration > 0.01 {
			diversityCount++
		}
	}

	mp.TotalBiomass = totalBiomass
	mp.Diversity = float64(diversityCount) / float64(len(mp.Components))
	mp.Toxicity = totalToxicity
}

// GetNutritionalValue calculates the nutritional value for specific needs
func (mp *MolecularProfile) GetNutritionalValue(needs *MolecularNeeds) float64 {
	totalValue := 0.0
	totalWeight := 0.0

	for molType, requirement := range needs.Requirements {
		if component, exists := mp.Components[molType]; exists {
			priority := needs.Priorities[molType]
			componentValue := component.Concentration * component.Quality * component.Freshness

			// Value is based on how well it meets the requirement
			satisfactionRatio := math.Min(componentValue/requirement, 1.0)
			weightedValue := satisfactionRatio * priority

			totalValue += weightedValue
			totalWeight += priority
		}
	}

	if totalWeight > 0 {
		return totalValue / totalWeight
	}
	return 0.0
}

// ConsumeNutrients simulates consuming nutrients from this profile based on needs and metabolism
func (mp *MolecularProfile) ConsumeNutrients(needs *MolecularNeeds, metabolism *MolecularMetabolism, consumptionRate float64) (energyGained float64, toxinDamage float64) {
	energyGained = 0.0
	toxinDamage = 0.0

	for molType, component := range mp.Components {
		if component.Concentration <= 0 {
			continue
		}

		// Calculate how much can be consumed
		maxConsumption := component.Concentration * consumptionRate
		efficiency := metabolism.Efficiency[molType]
		actualConsumption := maxConsumption * efficiency

		// Check if it's a toxin
		if molType >= ToxinAlkaloid {
			tolerance := needs.Tolerances[molType]
			detoxRate := metabolism.DetoxificationRate

			toxinEffect := actualConsumption * (1.0 - tolerance) * (1.0 - detoxRate)
			toxinDamage += toxinEffect * 10.0 // Scale toxin damage
		} else {
			// It's a nutrient - calculate energy gain
			requirement := needs.Requirements[molType]
			deficiency := needs.Deficiencies[molType]

			// Energy gain is higher when there's a deficiency
			deficiencyBonus := 1.0 + deficiency*2.0
			energyContribution := actualConsumption * component.Quality * deficiencyBonus

			// Different molecule types provide different energy amounts
			energyMultiplier := getMolecularEnergyMultiplier(molType)
			energyGained += energyContribution * energyMultiplier

			// Reduce deficiency
			deficiencyReduction := actualConsumption / requirement
			needs.Deficiencies[molType] = math.Max(0.0, needs.Deficiencies[molType]-deficiencyReduction)
		}

		// Update storage if there's capacity
		storageSpace := metabolism.StorageCapacity[molType] - metabolism.CurrentStorage[molType]
		if storageSpace > 0 {
			stored := math.Min(actualConsumption*0.3, storageSpace) // Store 30% of consumed
			metabolism.CurrentStorage[molType] += stored
		}
	}

	return energyGained, toxinDamage
}

// getMolecularEnergyMultiplier returns the energy multiplier for different molecule types
func getMolecularEnergyMultiplier(molType MolecularType) float64 {
	switch molType {
	case CarboSimple:
		return 4.0 // Quick energy
	case CarboComplex:
		return 3.5 // Sustained energy
	case LipidSaturated, LipidUnsaturated:
		return 9.0 // High energy density
	case ProteinStructural, ProteinEnzymatic:
		return 4.0 // Moderate energy, mainly for building
	case AminoEssential, AminoNonEssential, AminoConditional:
		return 3.0 // Building blocks
	case NucleicATP:
		return 8.0 // Direct energy currency
	case NucleicDNA, NucleicRNA:
		return 2.0 // Information storage, some energy
	case VitaminFat, VitaminWater:
		return 1.0 // Essential but low energy
	case MineralTrace:
		return 0.5 // Essential but no direct energy
	default:
		return 1.0
	}
}

// UpdateDeficiencies updates molecular deficiencies based on time and consumption
func (needs *MolecularNeeds) UpdateDeficiencies(timeStep float64) {
	for molType, requirement := range needs.Requirements {
		// Deficiencies increase over time as molecules are used up
		metabolicRate := 0.05 * timeStep // Base metabolic consumption rate

		// Some molecules are consumed faster
		switch molType {
		case CarboSimple, NucleicATP:
			metabolicRate *= 2.0 // Fast consumption
		case CarboComplex, LipidSaturated:
			metabolicRate *= 1.5 // Moderate consumption
		case ProteinStructural, AminoEssential:
			metabolicRate *= 0.5 // Slow consumption
		}

		currentDeficiency := needs.Deficiencies[molType]
		newDeficiency := math.Min(requirement, currentDeficiency+metabolicRate)
		needs.Deficiencies[molType] = newDeficiency
	}
}

// GetOverallNutritionalStatus calculates overall nutritional health (0.0-1.0)
func (needs *MolecularNeeds) GetOverallNutritionalStatus() float64 {
	totalDeficiency := 0.0
	totalRequirement := 0.0

	for molType, requirement := range needs.Requirements {
		deficiency := needs.Deficiencies[molType]
		priority := needs.Priorities[molType]

		weightedDeficiency := deficiency * priority
		weightedRequirement := requirement * priority

		totalDeficiency += weightedDeficiency
		totalRequirement += weightedRequirement
	}

	if totalRequirement > 0 {
		satisfactionRatio := 1.0 - (totalDeficiency / totalRequirement)
		return math.Max(0.0, math.Min(1.0, satisfactionRatio))
	}
	return 1.0
}

// CreatePlantMolecularProfile creates a molecular profile for a plant based on its type and traits
func CreatePlantMolecularProfile(plant *Plant) *MolecularProfile {
	profile := NewMolecularProfile()

	// Base profile varies by plant type
	switch plant.Type {
	case PlantGrass:
		// High in simple carbs and fiber, moderate proteins
		profile.AddComponent(CarboSimple, 0.6, 0.8, 1.0)
		profile.AddComponent(CarboFiber, 0.8, 0.7, 1.0)
		profile.AddComponent(ProteinStructural, 0.3, 0.6, 1.0)
		profile.AddComponent(VitaminWater, 0.4, 0.8, 1.0)
		profile.AddComponent(MineralTrace, 0.3, 0.7, 1.0)

	case PlantBush:
		// Balanced nutrients, some defensive compounds
		profile.AddComponent(CarboComplex, 0.5, 0.7, 1.0)
		profile.AddComponent(ProteinStructural, 0.4, 0.7, 1.0)
		profile.AddComponent(LipidUnsaturated, 0.3, 0.8, 1.0)
		profile.AddComponent(VitaminFat, 0.3, 0.6, 1.0)
		profile.AddComponent(ToxinTannin, 0.2, 0.8, 1.0)

	case PlantTree:
		// High structural content, some toxins for defense
		profile.AddComponent(CarboComplex, 0.7, 0.8, 1.0)
		profile.AddComponent(CarboFiber, 0.9, 0.9, 1.0)
		profile.AddComponent(ProteinStructural, 0.6, 0.8, 1.0)
		profile.AddComponent(LipidSaturated, 0.4, 0.7, 1.0)
		profile.AddComponent(ToxinAlkaloid, 0.3, 0.7, 1.0)

	case PlantMushroom:
		// High protein, unique amino acids, some toxins
		profile.AddComponent(ProteinEnzymatic, 0.8, 0.9, 1.0)
		profile.AddComponent(AminoEssential, 0.7, 0.9, 1.0)
		profile.AddComponent(AminoConditional, 0.6, 0.8, 1.0)
		profile.AddComponent(VitaminFat, 0.5, 0.8, 1.0)
		profile.AddComponent(ToxinGlycoside, 0.4, 0.6, 1.0)

	case PlantAlgae:
		// High protein, omega fatty acids, vitamins
		profile.AddComponent(ProteinStructural, 0.7, 0.9, 1.0)
		profile.AddComponent(LipidUnsaturated, 0.8, 0.9, 1.0)
		profile.AddComponent(VitaminWater, 0.6, 0.9, 1.0)
		profile.AddComponent(VitaminFat, 0.4, 0.8, 1.0)
		profile.AddComponent(MineralTrace, 0.5, 0.9, 1.0)

	case PlantCactus:
		// Water storage, some nutrients, defensive compounds
		profile.AddComponent(CarboSimple, 0.4, 0.6, 1.0)
		profile.AddComponent(VitaminWater, 0.8, 0.7, 1.0)
		profile.AddComponent(MineralTrace, 0.4, 0.8, 1.0)
		profile.AddComponent(ToxinAlkaloid, 0.6, 0.8, 1.0)
		profile.AddComponent(ToxinTannin, 0.3, 0.7, 1.0)
	}

	// Modify based on plant traits
	if plantTraits, ok := plant.Traits["nutrition_value"]; ok {
		nutritionMultiplier := 1.0 + plantTraits.Value*0.5
		for molType, component := range profile.Components {
			if molType < ToxinAlkaloid { // Only boost non-toxins
				component.Quality *= nutritionMultiplier
				profile.Components[molType] = component
			}
		}
	}

	if plantTraits, ok := plant.Traits["toxicity"]; ok {
		toxicityMultiplier := 1.0 + plantTraits.Value*0.8
		for molType, component := range profile.Components {
			if molType >= ToxinAlkaloid { // Only boost toxins
				component.Concentration *= toxicityMultiplier
				profile.Components[molType] = component
			}
		}
	}

	// Age affects freshness
	ageFactor := math.Max(0.1, 1.0-float64(plant.Age)*0.01)
	for molType, component := range profile.Components {
		component.Freshness *= ageFactor
		profile.Components[molType] = component
	}

	profile.updateDerivedMetrics()
	return profile
}

// CreateEntityMolecularProfile creates a molecular profile for an entity based on its traits and species
func CreateEntityMolecularProfile(entity *Entity) *MolecularProfile {
	profile := NewMolecularProfile()

	// Base profile for entities (what they're made of when consumed)
	size := entity.GetTrait("size")
	strength := entity.GetTrait("strength")
	intelligence := entity.GetTrait("intelligence")

	// Larger, stronger entities have more structural proteins
	structuralProtein := 0.4 + size*0.2 + strength*0.2
	profile.AddComponent(ProteinStructural, structuralProtein, 0.8, 1.0)

	// More intelligent entities have more enzymatic proteins and nucleic acids
	enzymaticProtein := 0.3 + intelligence*0.3
	profile.AddComponent(ProteinEnzymatic, enzymaticProtein, 0.8, 1.0)
	profile.AddComponent(NucleicDNA, 0.2+intelligence*0.2, 0.9, 1.0)
	profile.AddComponent(NucleicRNA, 0.3+intelligence*0.1, 0.8, 1.0)

	// Essential amino acids from muscle tissue
	aminoContent := 0.5 + strength*0.2 + size*0.1
	profile.AddComponent(AminoEssential, aminoContent, 0.9, 1.0)
	profile.AddComponent(AminoNonEssential, aminoContent*0.8, 0.7, 1.0)

	// Lipids for energy storage
	lipidContent := 0.3 + size*0.3
	profile.AddComponent(LipidSaturated, lipidContent, 0.7, 1.0)
	profile.AddComponent(LipidUnsaturated, lipidContent*0.6, 0.8, 1.0)

	// Transport proteins for circulatory systems
	transportProtein := 0.2 + intelligence*0.1
	profile.AddComponent(ProteinTransport, transportProtein, 0.7, 1.0)

	// Minerals and vitamins from organ tissue
	profile.AddComponent(MineralTrace, 0.3, 0.8, 1.0)
	profile.AddComponent(VitaminFat, 0.2, 0.6, 1.0)
	profile.AddComponent(VitaminWater, 0.3, 0.7, 1.0)

	// Species-specific modifications
	switch entity.Species {
	case "herbivore", "aquatic_herbivore", "aerial_herbivore":
		// Herbivores have less protein, more carbohydrates
		for molType, component := range profile.Components {
			if molType >= ProteinStructural && molType <= ProteinDefensive {
				component.Concentration *= 0.8
				profile.Components[molType] = component
			}
		}
		profile.AddComponent(CarboComplex, 0.4, 0.6, 1.0)

	case "carnivore", "predator":
		// Carnivores have more protein, defensive compounds
		for molType, component := range profile.Components {
			if molType >= ProteinStructural && molType <= ProteinDefensive {
				component.Concentration *= 1.3
				component.Quality *= 1.2
				profile.Components[molType] = component
			}
		}
		profile.AddComponent(ProteinDefensive, 0.3, 0.8, 1.0)

	case "omnivore", "aerial_omnivore":
		// Omnivores are balanced
		profile.AddComponent(CarboSimple, 0.3, 0.7, 1.0)
		profile.AddComponent(ProteinDefensive, 0.2, 0.7, 1.0)
	}

	// Environmental adaptations affect molecular composition
	if aquaticAdaptation := entity.GetTrait("aquatic_adaptation"); aquaticAdaptation > 0 {
		// Aquatic creatures have more unsaturated fats and transport proteins
		if component, exists := profile.Components[LipidUnsaturated]; exists {
			component.Concentration *= 1.0 + aquaticAdaptation*0.5
			component.Quality *= 1.0 + aquaticAdaptation*0.3
			profile.Components[LipidUnsaturated] = component
		}
		if component, exists := profile.Components[ProteinTransport]; exists {
			component.Concentration *= 1.0 + aquaticAdaptation*0.4
			profile.Components[ProteinTransport] = component
		}
	}

	if flyingAbility := entity.GetTrait("flying_ability"); flyingAbility > 0 {
		// Flying creatures have lighter, more efficient proteins
		for molType, component := range profile.Components {
			if molType >= ProteinStructural && molType <= ProteinDefensive {
				component.Quality *= 1.0 + flyingAbility*0.3
				profile.Components[molType] = component
			}
		}
		profile.AddComponent(NucleicATP, 0.4+flyingAbility*0.3, 0.9, 1.0)
	}

	// Age affects freshness and some nutritional quality
	ageFactor := math.Max(0.3, 1.0-float64(entity.Age)*0.005)
	for molType, component := range profile.Components {
		component.Freshness *= ageFactor
		// Older entities may have more complex, tougher proteins
		if molType == ProteinStructural {
			component.Quality *= (0.8 + float64(entity.Age)*0.002)
		}
		profile.Components[molType] = component
	}

	profile.updateDerivedMetrics()
	return profile
}

// GetMolecularDesirability calculates how desirable a food source is based on molecular content
func GetMolecularDesirability(foodProfile *MolecularProfile, entityNeeds *MolecularNeeds) float64 {
	if foodProfile == nil || entityNeeds == nil {
		return 0.0
	}

	// Base nutritional value
	nutritionalValue := foodProfile.GetNutritionalValue(entityNeeds)

	// Penalty for toxicity
	toxinPenalty := 0.0
	for molType, component := range foodProfile.Components {
		if molType >= ToxinAlkaloid {
			tolerance := entityNeeds.Tolerances[molType]
			if tolerance == 0 {
				tolerance = 0.1 // Minimum tolerance
			}
			toxinEffect := component.Concentration * (1.0 - tolerance)
			toxinPenalty += toxinEffect
		}
	}

	// Bonus for molecular diversity
	diversityBonus := foodProfile.Diversity * 0.2

	// Overall desirability
	desirability := nutritionalValue + diversityBonus - toxinPenalty*2.0
	return math.Max(0.0, math.Min(1.0, desirability))
}

// applySpeciesNutritionalRequirements adjusts molecular needs based on species type
func (needs *MolecularNeeds) applySpeciesNutritionalRequirements(species string) {
	// Herbivore species requirements
	if strings.Contains(species, "herbivore") {
		// High carbohydrate and fiber needs, low protein needs
		needs.Requirements[CarboComplex] *= 1.5
		needs.Requirements[CarboFiber] *= 2.0
		needs.Requirements[VitaminWater] *= 1.3
		needs.Requirements[ProteinStructural] *= 0.7
		needs.Requirements[ProteinEnzymatic] *= 0.6
		// Improved toxin tolerance for plant consumption
		needs.Tolerances[ToxinTannin] = 0.8
		needs.Tolerances[ToxinAlkaloid] = 0.6

	} else if strings.Contains(species, "carnivore") {
		// High protein and lipid needs, low carbohydrate needs
		needs.Requirements[ProteinStructural] *= 1.8
		needs.Requirements[ProteinEnzymatic] *= 1.6
		needs.Requirements[AminoEssential] *= 1.5
		needs.Requirements[LipidSaturated] *= 1.4
		needs.Requirements[CarboComplex] *= 0.5
		needs.Requirements[CarboSimple] *= 0.6
		// Better processing of animal proteins
		needs.Priorities[ProteinStructural] = 1.0
		needs.Priorities[AminoEssential] = 0.9

	} else if strings.Contains(species, "omnivore") {
		// Balanced requirements, good adaptability
		needs.Requirements[ProteinStructural] *= 1.2
		needs.Requirements[CarboComplex] *= 1.2
		needs.Requirements[VitaminFat] *= 1.1
		needs.Requirements[VitaminWater] *= 1.1
		// Moderate toxin tolerance
		needs.Tolerances[ToxinTannin] = 0.5
		needs.Tolerances[ToxinAlkaloid] = 0.4
	}

	// Size-based species requirements
	if strings.Contains(species, "large") {
		// Larger species need more total nutrition
		for molType := range needs.Requirements {
			needs.Requirements[molType] *= 1.4
		}
	} else if strings.Contains(species, "small") {
		// Smaller species need less total but higher quality nutrition
		for molType := range needs.Requirements {
			needs.Requirements[molType] *= 0.7
		}
		// Higher metabolic needs
		needs.Requirements[NucleicATP] *= 1.3
		needs.Requirements[MineralTrace] *= 1.2
	}

	// Aquatic species requirements
	if strings.Contains(species, "aquatic") {
		needs.Requirements[MineralSalt] *= 2.0      // High salt needs
		needs.Requirements[LipidUnsaturated] *= 1.3 // Insulation
		needs.Tolerances[MineralSalt] = 1.0         // High salt tolerance
	}

	// Aerial species requirements
	if strings.Contains(species, "aerial") {
		needs.Requirements[NucleicATP] *= 1.5   // High energy for flight
		needs.Requirements[CarboSimple] *= 1.4  // Quick energy
		needs.Requirements[MineralTrace] *= 0.8 // Lighter bones
	}

	// Underground species requirements
	if strings.Contains(species, "soil") || strings.Contains(species, "underground") {
		needs.Requirements[VitaminWater] *= 0.8 // Less water loss
		needs.Requirements[MineralTrace] *= 1.3 // Mineral-rich environment
		needs.Tolerances[ToxinHeavyMetal] = 0.6 // Heavy metal tolerance
	}
}

// applyEnvironmentalWaterRequirements adjusts water needs based on environmental adaptations
func (needs *MolecularNeeds) applyEnvironmentalWaterRequirements(entity *Entity) {
	baseWaterNeed := needs.Requirements[VitaminWater]

	// Aquatic adaptation reduces water dependency
	aquaticAdaptation := entity.GetTrait("aquatic_adaptation")
	if aquaticAdaptation > 0.5 {
		needs.Requirements[VitaminWater] = baseWaterNeed * (0.6 + 0.4*aquaticAdaptation)
		// Aquatic entities need salt balance
		needs.Requirements[MineralSalt] *= (1.0 + aquaticAdaptation*0.5)
	}

	// Desert/endurance adaptation affects water efficiency
	endurance := entity.GetTrait("endurance")
	if endurance > 0.7 {
		// High endurance means efficient water use
		needs.Requirements[VitaminWater] = baseWaterNeed * (0.8 - endurance*0.2)
	}

	// Size affects water needs
	size := entity.GetTrait("size")
	waterSizeMultiplier := 1.0 + size*0.5 // Larger entities need more water
	needs.Requirements[VitaminWater] *= waterSizeMultiplier

	// Underground navigation reduces surface water needs
	undergroundNav := entity.GetTrait("underground_nav")
	if undergroundNav > 0.5 {
		needs.Requirements[VitaminWater] *= (0.9 - undergroundNav*0.2)
	}

	// Flying ability increases water needs (dehydration)
	flyingAbility := entity.GetTrait("flying_ability")
	if flyingAbility > 0.5 {
		needs.Requirements[VitaminWater] *= (1.0 + flyingAbility*0.3)
	}

	// Set water priority based on environmental adaptation
	avgAdaptation := (aquaticAdaptation + endurance + undergroundNav) / 3.0
	needs.Priorities[VitaminWater] = 0.7 + avgAdaptation*0.3
}
