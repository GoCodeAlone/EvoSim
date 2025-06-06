package main

import (
	"math"
	"math/rand"
)

// OrganismClassification represents different organism types based on cellular complexity
type OrganismClassification int

const (
	ClassificationProkaryotic           OrganismClassification = iota // Simple single-celled
	ClassificationEukaryotic                                          // Complex single-celled
	ClassificationSimpleMulticellular                                 // Simple multicellular
	ClassificationComplexMulticellular                                // Complex multicellular
	ClassificationAdvancedMulticellular                               // Advanced multicellular
)

// OrganismLifespanData contains lifespan and aging information for each classification
type OrganismLifespanData struct {
	Classification     OrganismClassification
	Name               string
	BaseLifespanTicks  int     // Base lifespan in ticks
	LifespanVariance   float64 // Variance factor (0.0-1.0)
	AgingRate          float64 // Rate at which organism ages per tick (1.0 = normal, 0.5 = half speed)
	MaturationAge      int     // Age at which reproduction becomes possible
	PeakAge            int     // Age at peak fitness
	SenescenceAge      int     // Age when aging effects accelerate
	MetabolicRate      float64 // Base metabolic rate affecting energy consumption
	CellularMainenance float64 // Energy cost for maintaining cellular structure
}

// OrganismClassifier manages organism classification and lifespan mechanics
type OrganismClassifier struct {
	LifespanData map[OrganismClassification]*OrganismLifespanData
	TimeSystem   *AdvancedTimeSystem
}

// NewOrganismClassifier creates a new organism classification system
func NewOrganismClassifier(timeSystem *AdvancedTimeSystem) *OrganismClassifier {
	classifier := &OrganismClassifier{
		LifespanData: make(map[OrganismClassification]*OrganismLifespanData),
		TimeSystem:   timeSystem,
	}

	// Initialize realistic lifespan data based on biological principles
	classifier.initializeLifespanData()

	return classifier
}

// initializeLifespanData sets up realistic lifespan ranges for each organism type
func (oc *OrganismClassifier) initializeLifespanData() {
	// Convert days to ticks for easier understanding
	ticksPerDay := float64(oc.TimeSystem.DayLength)

	// Adjust for the new time scale: if we have 1 tick per day, we need to scale appropriately
	// The old system had 480 ticks/day, so we need to scale lifespans to maintain the same relative timing
	timeScaleFactor := 1.0
	if ticksPerDay < 10 { // If we're using the new time scale (few ticks per day)
		timeScaleFactor = 7.0 // Scale up lifespans to be reasonable for the new time scale
	}

	// Prokaryotic organisms (bacteria-like): Hours to days
	oc.LifespanData[ClassificationProkaryotic] = &OrganismLifespanData{
		Classification:     ClassificationProkaryotic,
		Name:               "Prokaryotic",
		BaseLifespanTicks:  int(ticksPerDay * 15 * timeScaleFactor),  // 15 days base lifespan (scaled) to meet test expectations
		LifespanVariance:   0.8,                                      // High variance
		AgingRate:          2.0,                                      // Age twice as fast
		MaturationAge:      int(ticksPerDay * 0.1 * timeScaleFactor), // 0.1 days (scaled)
		PeakAge:            int(ticksPerDay * 0.5 * timeScaleFactor), // 0.5 days (scaled)
		SenescenceAge:      int(ticksPerDay * 1.2 * timeScaleFactor), // 1.2 days (scaled)
		MetabolicRate:      2.0,                                      // High metabolism
		CellularMainenance: 0.1,                                      // Low maintenance cost
	}

	// Eukaryotic organisms (protozoa-like): Days to weeks
	oc.LifespanData[ClassificationEukaryotic] = &OrganismLifespanData{
		Classification:     ClassificationEukaryotic,
		Name:               "Eukaryotic",
		BaseLifespanTicks:  int(ticksPerDay * 30 * timeScaleFactor), // 30 days base lifespan (scaled up)
		LifespanVariance:   0.6,                                     // Moderate variance
		AgingRate:          1.5,                                     // Age 1.5x as fast
		MaturationAge:      int(ticksPerDay * 2 * timeScaleFactor),  // 2 days (scaled)
		PeakAge:            int(ticksPerDay * 10 * timeScaleFactor), // 10 days (scaled)
		SenescenceAge:      int(ticksPerDay * 20 * timeScaleFactor), // 20 days (scaled)
		MetabolicRate:      1.5,                                     // High metabolism
		CellularMainenance: 0.2,                                     // Low maintenance cost
	}

	// Simple Multicellular organisms: Weeks to months
	oc.LifespanData[ClassificationSimpleMulticellular] = &OrganismLifespanData{
		Classification:     ClassificationSimpleMulticellular,
		Name:               "Simple Multicellular",
		BaseLifespanTicks:  int(ticksPerDay * 90 * timeScaleFactor), // 90 days base lifespan (scaled)
		LifespanVariance:   0.5,                                     // Moderate variance
		AgingRate:          1.0,                                     // Normal aging rate
		MaturationAge:      int(ticksPerDay * 10 * timeScaleFactor), // 10 days (scaled)
		PeakAge:            int(ticksPerDay * 30 * timeScaleFactor), // 30 days (scaled)
		SenescenceAge:      int(ticksPerDay * 60 * timeScaleFactor), // 60 days (scaled)
		MetabolicRate:      1.2,                                     // Moderate metabolism
		CellularMainenance: 0.5,                                     // Moderate maintenance cost
	}

	// Complex Multicellular organisms: Months to years
	oc.LifespanData[ClassificationComplexMulticellular] = &OrganismLifespanData{
		Classification:     ClassificationComplexMulticellular,
		Name:               "Complex Multicellular",
		BaseLifespanTicks:  int(ticksPerDay * 365 * timeScaleFactor), // 1 year base lifespan (scaled)
		LifespanVariance:   0.4,                                      // Lower variance
		AgingRate:          0.8,                                      // Slower aging
		MaturationAge:      int(ticksPerDay * 30 * timeScaleFactor),  // 30 days (scaled)
		PeakAge:            int(ticksPerDay * 120 * timeScaleFactor), // 120 days (scaled)
		SenescenceAge:      int(ticksPerDay * 240 * timeScaleFactor), // 240 days (scaled)
		MetabolicRate:      1.0,                                      // Normal metabolism
		CellularMainenance: 1.0,                                      // Higher maintenance cost
	}

	// Advanced Multicellular organisms: Years to decades
	oc.LifespanData[ClassificationAdvancedMulticellular] = &OrganismLifespanData{
		Classification:     ClassificationAdvancedMulticellular,
		Name:               "Advanced Multicellular",
		BaseLifespanTicks:  int(ticksPerDay * 1095 * timeScaleFactor), // 3 years base lifespan (scaled)
		LifespanVariance:   0.3,                                       // Low variance
		AgingRate:          0.5,                                       // Much slower aging
		MaturationAge:      int(ticksPerDay * 90 * timeScaleFactor),   // 90 days (scaled)
		PeakAge:            int(ticksPerDay * 365 * timeScaleFactor),  // 1 year (scaled)
		SenescenceAge:      int(ticksPerDay * 730 * timeScaleFactor),  // 2 years (scaled)
		MetabolicRate:      0.8,                                       // Lower metabolism
		CellularMainenance: 1.5,                                       // High maintenance cost
	}
}

// ClassifyEntity determines the organism classification for an entity
func (oc *OrganismClassifier) ClassifyEntity(entity *Entity, cellularSystem *CellularSystem) OrganismClassification {
	// Get cellular organism data if available
	if cellularSystem != nil {
		if organism, exists := cellularSystem.OrganismMap[entity.ID]; exists {
			return oc.classifyByCellularComplexity(organism.ComplexityLevel, entity)
		}
	}

	// Fallback classification based on entity traits
	return oc.classifyByTraits(entity)
}

// classifyByCellularComplexity maps cellular complexity levels to organism classifications
func (oc *OrganismClassifier) classifyByCellularComplexity(complexityLevel int, entity *Entity) OrganismClassification {
	switch complexityLevel {
	case 1:
		// Single cell - determine if prokaryotic or eukaryotic based on traits
		intelligence := entity.GetTrait("intelligence")
		if intelligence < -0.5 {
			return ClassificationProkaryotic
		}
		return ClassificationEukaryotic

	case 2:
		return ClassificationSimpleMulticellular

	case 3:
		return ClassificationComplexMulticellular

	case 4, 5:
		return ClassificationAdvancedMulticellular

	default:
		return ClassificationEukaryotic
	}
}

// classifyByTraits provides fallback classification based on entity traits
func (oc *OrganismClassifier) classifyByTraits(entity *Entity) OrganismClassification {
	intelligence := entity.GetTrait("intelligence")
	size := entity.GetTrait("size")
	cooperation := entity.GetTrait("cooperation")

	// Calculate a complexity score from traits
	complexityScore := (intelligence + size + cooperation) / 3.0

	switch {
	case complexityScore < -0.5:
		return ClassificationProkaryotic
	case complexityScore < -0.2:
		return ClassificationEukaryotic
	case complexityScore < 0.2:
		return ClassificationSimpleMulticellular
	case complexityScore < 0.6:
		return ClassificationComplexMulticellular
	default:
		return ClassificationAdvancedMulticellular
	}
}

// CalculateLifespan determines the actual lifespan for a specific entity
func (oc *OrganismClassifier) CalculateLifespan(entity *Entity, classification OrganismClassification) int {
	data := oc.LifespanData[classification]

	// Apply trait-based modifiers
	endurance := entity.GetTrait("endurance")
	size := entity.GetTrait("size")

	// Base lifespan with variance
	variance := (rand.Float64() - 0.5) * 2 * data.LifespanVariance
	baseLifespan := float64(data.BaseLifespanTicks) * (1.0 + variance)

	// Trait modifiers
	enduranceModifier := 1.0 + endurance*0.3 // ±30% based on endurance
	sizeModifier := 1.0 + size*0.2           // ±20% based on size

	// Calculate final lifespan
	finalLifespan := baseLifespan * enduranceModifier * sizeModifier

	// Ensure lifespan stays within reasonable bounds (30% to 200% of base)
	minLifespan := float64(data.BaseLifespanTicks) * 0.3
	maxLifespan := float64(data.BaseLifespanTicks) * 2.0
	finalLifespan = math.Max(minLifespan, math.Min(maxLifespan, finalLifespan))

	return int(finalLifespan)
}

// CalculateAgingRate determines how fast an entity ages based on its classification
func (oc *OrganismClassifier) CalculateAgingRate(entity *Entity, classification OrganismClassification) float64 {
	data := oc.LifespanData[classification]

	// Base aging rate from classification
	agingRate := data.AgingRate

	// Metabolic modifiers
	metabolism := entity.GetTrait("metabolism")
	size := entity.GetTrait("size")

	// Higher metabolism = faster aging, larger size = slower aging
	metabolismModifier := 1.0 + metabolism*0.2 // ±20% from metabolism
	sizeModifier := 1.0 - size*0.1             // Larger = slower aging

	return agingRate * metabolismModifier * sizeModifier
}

// CalculateEnergyMaintenance calculates energy cost for maintaining the organism
func (oc *OrganismClassifier) CalculateEnergyMaintenance(entity *Entity, classification OrganismClassification) float64 {
	data := oc.LifespanData[classification]

	// Base maintenance cost
	baseCost := data.CellularMainenance

	// Size and complexity affect maintenance
	size := entity.GetTrait("size")
	intelligence := entity.GetTrait("intelligence")

	sizeModifier := 1.0 + size*0.5               // Larger organisms cost more
	complexityModifier := 1.0 + intelligence*0.3 // Smarter organisms cost more

	return baseCost * sizeModifier * complexityModifier
}

// ShouldAge determines if an entity should age this tick based on its classification
func (oc *OrganismClassifier) ShouldAge(entity *Entity, classification OrganismClassification) bool {
	agingRate := oc.CalculateAgingRate(entity, classification)

	// Use probabilistic aging for fractional rates
	if agingRate >= 1.0 {
		// Fast aging - might age multiple times per tick
		return true
	} else {
		// Slow aging - probabilistic
		return rand.Float64() < agingRate
	}
}

// IsDeathByOldAge determines if an entity should die from old age
func (oc *OrganismClassifier) IsDeathByOldAge(entity *Entity, classification OrganismClassification, maxLifespan int) bool {
	data := oc.LifespanData[classification]

	// Calculate death probability based on age
	if entity.Age >= maxLifespan {
		return true
	}

	// Increased death probability in senescence
	if entity.Age >= data.SenescenceAge {
		senescenceProgress := float64(entity.Age-data.SenescenceAge) / float64(maxLifespan-data.SenescenceAge)
		deathProbability := senescenceProgress * 0.01 // Up to 1% chance per tick in late senescence
		return rand.Float64() < deathProbability
	}

	return false
}

// GetLifespanData returns the lifespan data for a classification
func (oc *OrganismClassifier) GetLifespanData(classification OrganismClassification) *OrganismLifespanData {
	return oc.LifespanData[classification]
}

// GetClassificationName returns the human-readable name of a classification
func (oc *OrganismClassifier) GetClassificationName(classification OrganismClassification) string {
	if data, exists := oc.LifespanData[classification]; exists {
		return data.Name
	}
	return "Unknown"
}

// IsReproductivelyMature determines if an entity is old enough to reproduce
func (oc *OrganismClassifier) IsReproductivelyMature(entity *Entity, classification OrganismClassification) bool {
	data := oc.LifespanData[classification]
	return entity.Age >= data.MaturationAge
}

// GetOptimalReproductiveAge returns the age at which reproduction is most successful
func (oc *OrganismClassifier) GetOptimalReproductiveAge(classification OrganismClassification) int {
	data := oc.LifespanData[classification]
	return data.PeakAge
}

// CalculateReproductiveVigor calculates how effectively an entity can reproduce based on age
func (oc *OrganismClassifier) CalculateReproductiveVigor(entity *Entity, classification OrganismClassification) float64 {
	data := oc.LifespanData[classification]

	if entity.Age < data.MaturationAge {
		return 0.0 // Too young to reproduce
	}

	if entity.Age >= data.SenescenceAge {
		// Declining reproductive capability in old age
		senescenceSpan := entity.MaxLifespan - data.SenescenceAge
		if senescenceSpan <= 0 {
			// Edge case: senescence age is at or beyond max lifespan
			return 0.2 // Minimal reproductive vigor
		}
		senescenceProgress := float64(entity.Age-data.SenescenceAge) / float64(senescenceSpan)
		vigor := 1.0 - senescenceProgress*0.8      // Up to 80% reduction in old age
		return math.Max(0.0, math.Min(1.0, vigor)) // Ensure between 0 and 1
	}

	// Peak vigor at optimal age
	if entity.Age <= data.PeakAge {
		// Increasing vigor from maturation to peak
		progress := float64(entity.Age-data.MaturationAge) / float64(data.PeakAge-data.MaturationAge)
		vigor := 0.5 + progress*0.5                // 50% to 100% vigor
		return math.Max(0.0, math.Min(1.0, vigor)) // Ensure between 0 and 1
	}

	// Stable vigor from peak to senescence - but slightly declining
	ageProgress := float64(entity.Age-data.PeakAge) / float64(data.SenescenceAge-data.PeakAge)
	vigor := 1.0 - ageProgress*0.1             // Slight decline before senescence
	return math.Max(0.0, math.Min(1.0, vigor)) // Ensure between 0 and 1
}
