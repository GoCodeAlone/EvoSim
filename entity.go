package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Trait represents a dynamic trait with a name and value
type Trait struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Position represents coordinates in the world
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DietaryMemory tracks what an entity has been eating to create evolutionary dependencies
type DietaryMemory struct {
	PlantTypePreferences   map[int]float64 `json:"plant_type_preferences"`   // Plant type -> preference strength (0.0-2.0)
	PreySpeciesPreferences map[string]float64 `json:"prey_species_preferences"` // Species -> preference strength (0.0-2.0)
	ConsumptionHistory     []ConsumptionRecord `json:"consumption_history"`      // Recent consumption records
	DietaryFitness         float64 `json:"dietary_fitness"`                     // How well adapted to current diet
}

// ConsumptionRecord tracks a single feeding event
type ConsumptionRecord struct {
	Tick       int    `json:"tick"`        // When it happened
	FoodType   string `json:"food_type"`   // "plant" or "entity"
	FoodID     string `json:"food_id"`     // Specific identifier
	Nutrition  float64 `json:"nutrition"`  // Nutritional value gained
	Toxicity   float64 `json:"toxicity"`   // Any toxin exposure
}

// EnvironmentalMemory tracks environmental pressures for evolutionary adaptation  
type EnvironmentalMemory struct {
	BiomeExposure        map[BiomeType]float64 `json:"biome_exposure"`         // Biome type -> exposure time ratio
	TemperaturePressure  float64 `json:"temperature_pressure"`                 // Cumulative temperature stress
	RadiationPressure    float64 `json:"radiation_pressure"`                   // Cumulative radiation exposure  
	SeasonalPressure     map[string]float64 `json:"seasonal_pressure"`        // Season -> stress level
	EventPressure        map[string]float64 `json:"event_pressure"`           // Event type -> cumulative stress
	AdaptationFitness    float64 `json:"adaptation_fitness"`                   // How well adapted to environment
}

// Entity represents an individual in the population with dynamic traits
type Entity struct {
	ID         int              `json:"id"`
	Traits     map[string]Trait `json:"traits"`
	Fitness    float64          `json:"fitness"`
	Position   Position         `json:"position"`
	Energy     float64          `json:"energy"`
	Age        int              `json:"age"`
	IsAlive    bool             `json:"is_alive"`
	Species    string           `json:"species"`
	Generation int              `json:"generation"`
	TribeID    int              `json:"tribe_id"` // ID of the tribe this entity belongs to (0 = no tribe)
	
	// Molecular system components
	MolecularNeeds     *MolecularNeeds     `json:"molecular_needs"`
	MolecularMetabolism *MolecularMetabolism `json:"molecular_metabolism"`
	MolecularProfile   *MolecularProfile    `json:"molecular_profile"` // What this entity is made of (for consumption)
	
	// Feedback loop systems for evolutionary adaptation
	DietaryMemory       *DietaryMemory       `json:"dietary_memory"`        // Tracks feeding history for evolutionary dependency
	EnvironmentalMemory *EnvironmentalMemory `json:"environmental_memory"`  // Tracks environmental pressure for adaptation
	
	// Reproduction system
	ReproductionStatus *ReproductionStatus `json:"reproduction_status"` // Tracks reproduction state and behaviors
}

// NewEntity creates a new entity with random traits
func NewEntity(id int, traitNames []string, species string, position Position) *Entity {
	entity := &Entity{
		ID:         id,
		Traits:     make(map[string]Trait),
		Fitness:    0.0,
		Position:   position,
		Energy:     100.0, // Starting energy
		Age:        0,
		IsAlive:    true,
		Species:    species,
		Generation: 0,
	}

	// Initialize random traits
	for _, name := range traitNames {
		entity.Traits[name] = Trait{
			Name:  name,
			Value: rand.Float64()*2 - 1, // Random value between -1 and 1
		}
	}

	// Initialize molecular systems
	entity.MolecularNeeds = NewMolecularNeeds(entity)
	entity.MolecularMetabolism = NewMolecularMetabolism(entity)
	entity.MolecularProfile = CreateEntityMolecularProfile(entity)

	// Initialize feedback loop systems
	entity.DietaryMemory = NewDietaryMemory()
	entity.EnvironmentalMemory = NewEnvironmentalMemory()
	
	// Initialize reproduction system
	entity.ReproductionStatus = NewReproductionStatus()

	return entity
}

// GetTrait safely gets a trait value, returning 0 if not found
func (e *Entity) GetTrait(name string) float64 {
	if trait, exists := e.Traits[name]; exists {
		return trait.Value
	}
	return 0.0
}

// SetTrait sets or updates a trait
func (e *Entity) SetTrait(name string, value float64) {
	e.Traits[name] = Trait{Name: name, Value: value}
}

// Mutate applies random mutations to the entity's traits with feedback loop influence
func (e *Entity) Mutate(mutationRate float64, mutationStrength float64) {
	// Calculate mutation pressure from feedback loops
	
	// Environmental pressure increases mutation rate and strength
	if e.EnvironmentalMemory != nil {
		environmentalPressure := 0.0
		
		// High radiation exposure increases mutation
		environmentalPressure += e.EnvironmentalMemory.RadiationPressure * 0.1
		
		// Temperature stress increases mutation  
		environmentalPressure += e.EnvironmentalMemory.TemperaturePressure * 0.05
		
		// Poor environmental adaptation increases mutation
		if e.EnvironmentalMemory.AdaptationFitness < 0.8 {
			environmentalPressure += (0.8 - e.EnvironmentalMemory.AdaptationFitness) * 0.3
		}
		
		// Apply environmental pressure to mutation
		mutationRate += environmentalPressure
		mutationStrength += environmentalPressure * 0.5
	}
	
	// Dietary stress can also influence mutation (starvation drives evolution)
	if e.DietaryMemory != nil && e.DietaryMemory.DietaryFitness < 0.6 {
		dietaryStress := (0.6 - e.DietaryMemory.DietaryFitness) * 0.2
		mutationRate += dietaryStress  
		mutationStrength += dietaryStress * 0.3
	}
	
	// Cap mutation rate and strength to reasonable values
	mutationRate = math.Min(mutationRate, 0.5) // Max 50% mutation chance
	mutationStrength = math.Min(mutationStrength, 0.8) // Max strength
	
	for name, trait := range e.Traits {
		if rand.Float64() < mutationRate {
			// Apply Gaussian noise for mutation
			mutation := rand.NormFloat64() * mutationStrength
			
			// Bias mutations toward beneficial directions based on feedback
			mutation = e.biasedMutation(name, mutation)
			
			newValue := trait.Value + mutation

			// Clamp values to reasonable bounds
			newValue = math.Max(-2.0, math.Min(2.0, newValue))

			e.Traits[name] = Trait{
				Name:  name,
				Value: newValue,
			}
		}
	}
}

// Clone creates a deep copy of the entity
func (e *Entity) Clone() *Entity {
	clone := &Entity{
		ID:         e.ID,
		Traits:     make(map[string]Trait),
		Fitness:    e.Fitness,
		Position:   e.Position,
		Energy:     e.Energy,
		Age:        e.Age,
		IsAlive:    e.IsAlive,
		Species:    e.Species,
		Generation: e.Generation,
	}

	for name, trait := range e.Traits {
		clone.Traits[name] = Trait{
			Name:  trait.Name,
			Value: trait.Value,
		}
	}

	// Clone memory systems (shallow clone for now - they'll be updated by inheritance)
	if e.DietaryMemory != nil {
		clone.DietaryMemory = NewDietaryMemory()
	}
	if e.EnvironmentalMemory != nil {
		clone.EnvironmentalMemory = NewEnvironmentalMemory()
	}
	
	// Clone reproduction status
	if e.ReproductionStatus != nil {
		clone.ReproductionStatus = NewReproductionStatus()
		// Copy some heritable traits
		clone.ReproductionStatus.Mode = e.ReproductionStatus.Mode
		clone.ReproductionStatus.Strategy = e.ReproductionStatus.Strategy
	}

	return clone
}

// String returns a string representation of the entity
func (e *Entity) String() string {
	return fmt.Sprintf("Entity{ID: %d, Species: %s, Fitness: %.3f, Energy: %.1f, Pos: (%.1f,%.1f), Alive: %t}",
		e.ID, e.Species, e.Fitness, e.Energy, e.Position.X, e.Position.Y, e.IsAlive)
}

// DistanceTo calculates the distance to another entity
func (e *Entity) DistanceTo(other *Entity) float64 {
	dx := e.Position.X - other.Position.X
	dy := e.Position.Y - other.Position.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// MoveTo moves the entity towards a target position with environment-specific adaptations
func (e *Entity) MoveTo(targetX, targetY float64, speed float64) {
	if !e.IsAlive {
		return
	}

	dx := targetX - e.Position.X
	dy := targetY - e.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 0 {
		// Normalize direction and apply speed
		dx = (dx / distance) * speed
		dy = (dy / distance) * speed

		e.Position.X += dx
		e.Position.Y += dy

		// Moving costs energy (base cost)
		energyCost := speed * 0.1
		e.Energy -= energyCost
	}
}

// MoveToWithEnvironment moves the entity with environment-specific adaptations
func (e *Entity) MoveToWithEnvironment(targetX, targetY float64, speed float64, biome BiomeType) {
	if !e.IsAlive {
		return
	}

	// Calculate environment-specific movement efficiency
	effectiveSpeed := speed
	energyMultiplier := 1.0

	switch biome {
	case BiomeWater:
		// Aquatic movement - efficiency based on aquatic adaptation
		aquaticAdaptation := e.GetTrait("aquatic_adaptation")
		if aquaticAdaptation < 0 {
			// Poor adaptation - slower and more costly
			effectiveSpeed *= (1.0 + aquaticAdaptation*0.5) // Reduce speed
			energyMultiplier = 2.0 + math.Abs(aquaticAdaptation) // Higher energy cost
		} else {
			// Good adaptation - potentially faster swimming
			effectiveSpeed *= (1.0 + aquaticAdaptation*0.3)
			energyMultiplier = 0.8 // Lower energy cost
		}

	case BiomeSoil:
		// Underground movement - efficiency based on digging ability
		diggingAbility := e.GetTrait("digging_ability")
		undergroundNav := e.GetTrait("underground_nav")
		
		if diggingAbility < 0 || undergroundNav < 0 {
			// Poor soil adaptation - much slower and more costly
			effectiveSpeed *= 0.3 // Very slow underground
			energyMultiplier = 3.0 // Very high energy cost
		} else {
			// Good soil adaptation
			effectiveSpeed *= (0.7 + diggingAbility*0.3)
			energyMultiplier = 1.5 - undergroundNav*0.3
		}

	case BiomeAir:
		// Aerial movement - efficiency based on flying ability
		flyingAbility := e.GetTrait("flying_ability")
		altitudeTolerance := e.GetTrait("altitude_tolerance")
		
		if flyingAbility < -0.5 {
			// Cannot fly - fall or struggle
			effectiveSpeed *= 0.1 // Extremely slow
			energyMultiplier = 5.0 // Very high energy cost
			e.Energy -= 2.0 // Additional damage from struggling at altitude
		} else if flyingAbility > 0 {
			// Good flying - faster and more efficient
			effectiveSpeed *= (1.0 + flyingAbility*0.5)
			energyMultiplier = 0.6 + altitudeTolerance*0.2
		}

	default:
		// Land movement - default behavior
		effectiveSpeed = speed
		energyMultiplier = 1.0
	}

	// Apply movement
	dx := targetX - e.Position.X
	dy := targetY - e.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 0 {
		// Normalize direction and apply effective speed
		dx = (dx / distance) * effectiveSpeed
		dy = (dy / distance) * effectiveSpeed

		e.Position.X += dx
		e.Position.Y += dy

		// Apply environment-specific energy cost
		energyCost := effectiveSpeed * 0.1 * energyMultiplier
		e.Energy -= energyCost
	}
}

// MoveRandomly moves the entity in a random direction with environment considerations
func (e *Entity) MoveRandomly(maxDistance float64) {
	if !e.IsAlive {
		return
	}

	angle := rand.Float64() * 2 * math.Pi
	distance := rand.Float64() * maxDistance

	e.Position.X += math.Cos(angle) * distance
	e.Position.Y += math.Sin(angle) * distance

	// Random movement costs less energy
	e.Energy -= distance * 0.05
}

// MoveRandomlyWithEnvironment moves the entity randomly with environment-specific adaptations
func (e *Entity) MoveRandomlyWithEnvironment(maxDistance float64, biome BiomeType) {
	if !e.IsAlive {
		return
	}

	angle := rand.Float64() * 2 * math.Pi
	distance := rand.Float64() * maxDistance

	// Apply environment-specific movement constraints
	effectiveDistance := distance
	energyCostMultiplier := 1.0

	switch biome {
	case BiomeWater:
		aquaticAdaptation := e.GetTrait("aquatic_adaptation")
		if aquaticAdaptation < 0 {
			effectiveDistance *= (1.0 + aquaticAdaptation*0.5) // Struggle in water
			energyCostMultiplier = 2.0 + math.Abs(aquaticAdaptation)
		} else {
			effectiveDistance *= (1.0 + aquaticAdaptation*0.3)
			energyCostMultiplier = 0.8
		}

	case BiomeSoil:
		diggingAbility := e.GetTrait("digging_ability")
		undergroundNav := e.GetTrait("underground_nav")
		
		if diggingAbility < 0 || undergroundNav < 0 {
			effectiveDistance *= 0.4 // Slow underground
			energyCostMultiplier = 6.0 // Very high energy cost
		} else {
			effectiveDistance *= (0.7 + diggingAbility*0.3)
			energyCostMultiplier = 1.2 - undergroundNav*0.2
		}

	case BiomeAir:
		flyingAbility := e.GetTrait("flying_ability")
		altitudeTolerance := e.GetTrait("altitude_tolerance")
		
		if flyingAbility < -0.5 {
			effectiveDistance *= 0.1 // Cannot fly properly
			energyCostMultiplier = 5.0
			e.Energy -= 2.0 // Additional damage from struggling at altitude
		} else if flyingAbility > 0 {
			effectiveDistance *= (1.0 + flyingAbility*0.5)
			energyCostMultiplier = 0.6 + altitudeTolerance*0.2
		}
	}

	e.Position.X += math.Cos(angle) * effectiveDistance
	e.Position.Y += math.Sin(angle) * effectiveDistance

	// Apply environment-specific energy cost
	energyCost := effectiveDistance * 0.1 * energyCostMultiplier
	e.Energy -= energyCost
}

// CanKill determines if this entity can kill another based on traits
func (e *Entity) CanKill(other *Entity) bool {
	if !e.IsAlive || !other.IsAlive || e.Species == other.Species {
		return false
	}

	// Killing ability based on aggression, strength, and size difference
	myPower := e.GetTrait("aggression") + e.GetTrait("strength") + e.GetTrait("size")
	theirPower := other.GetTrait("defense") + other.GetTrait("strength") + other.GetTrait("size")

	// Add some randomness to combat
	myPower += (rand.Float64() - 0.5) * 0.5

	return myPower > theirPower && e.Energy > 20
}

// Kill attempts to kill another entity
func (e *Entity) Kill(other *Entity) bool {
	if !e.CanKill(other) {
		return false
	}

	// Killing costs energy but may provide rewards
	e.Energy -= 15
	other.IsAlive = false
	other.Energy = 0

	// Gain some energy from the kill
	energyGain := other.Energy * 0.3
	e.Energy += energyGain

	return true
}

// CanEat determines if this entity can eat another (usually dead entities)
func (e *Entity) CanEat(other *Entity) bool {
	if !e.IsAlive {
		return false
	}

	// Can eat dead entities or much smaller living ones
	if !other.IsAlive {
		return true
	}

	// Can eat living entities if much larger and they're herbivores
	mySize := e.GetTrait("size")
	theirSize := other.GetTrait("size")

	return mySize > theirSize+1.0 && other.GetTrait("aggression") < 0.0
}

// Eat consumes another entity for energy using molecular system
func (e *Entity) Eat(other *Entity, tick int) bool {
	if !e.CanEat(other) {
		return false
	}

	// Use molecular system for consumption if available
	if other.MolecularProfile != nil && e.MolecularNeeds != nil && e.MolecularMetabolism != nil {
		// Calculate molecular desirability 
		desirability := GetMolecularDesirability(other.MolecularProfile, e.MolecularNeeds)
		
		// Base consumption based on prey size
		consumptionRate := (other.GetTrait("size") + 1.0) * 0.1
		
		// Consume nutrients using molecular system
		energyGained, toxinDamage := other.MolecularProfile.ConsumeNutrients(
			e.MolecularNeeds,
			e.MolecularMetabolism,
			consumptionRate)
		
		// Apply molecular energy gain
		e.Energy += energyGained * 15.0 // Higher energy gain from meat
		e.Energy -= toxinDamage
		
		// Also get remaining energy from the consumed entity
		e.Energy += other.Energy * 0.3
		
		// Desirability affects hunting motivation for future reference
		if desirability > 0.7 {
			// High-quality prey - remember this for future behavior evolution
			e.SetTrait("prey_preference_" + other.Species, math.Min(1.0, e.GetTrait("prey_preference_" + other.Species) + 0.1))
		}
	} else {
		// Fallback to traditional system
		energyGain := (other.GetTrait("size")+1.0)*10 + other.Energy*0.5
		e.Energy += energyGain
	}

	// Eating costs some energy
	e.Energy -= 5

	// Record consumption in dietary memory for feedback loop
	e.recordEntityConsumption(other, tick)

	// Consumed entity dies if it wasn't already dead
	if other.IsAlive {
		other.IsAlive = false
	}
	other.Energy = 0

	return true
}

// CanMerge determines if this entity can merge with another
func (e *Entity) CanMerge(other *Entity) bool {
	if !e.IsAlive || !other.IsAlive {
		return false
	}

	// Can only merge with same species
	if e.Species != other.Species {
		return false
	}

	// Both entities must have sufficient energy and similar traits
	if e.Energy < 50 || other.Energy < 50 {
		return false
	}

	// Check trait compatibility (similar intelligence and cooperation)
	intelligenceDiff := math.Abs(e.GetTrait("intelligence") - other.GetTrait("intelligence"))
	cooperationSum := e.GetTrait("cooperation") + other.GetTrait("cooperation")

	return intelligenceDiff < 0.5 && cooperationSum > 0.5
}

// Merge combines this entity with another, creating a new entity
func (e *Entity) Merge(other *Entity, newID int) *Entity {
	if !e.CanMerge(other) {
		return nil
	}

	// Create new merged entity
	merged := &Entity{
		ID:         newID,
		Traits:     make(map[string]Trait),
		Fitness:    0.0,
		Position:   Position{X: (e.Position.X + other.Position.X) / 2, Y: (e.Position.Y + other.Position.Y) / 2},
		Energy:     (e.Energy + other.Energy) * 0.9, // 10% energy loss in merge
		Age:        0,                               // New entity starts young
		IsAlive:    true,
		Species:    e.Species,
		Generation: int(math.Max(float64(e.Generation), float64(other.Generation))) + 1,
	}

	// Merge traits by averaging
	allTraits := make(map[string]bool)
	for name := range e.Traits {
		allTraits[name] = true
	}
	for name := range other.Traits {
		allTraits[name] = true
	}

	for name := range allTraits {
		val1 := e.GetTrait(name)
		val2 := other.GetTrait(name)
		avgValue := (val1 + val2) / 2.0

		// Add small random variation
		avgValue += (rand.Float64() - 0.5) * 0.1
		avgValue = math.Max(-2.0, math.Min(2.0, avgValue))

		merged.SetTrait(name, avgValue)
	}

	// Original entities die in the merge
	e.IsAlive = false
	other.IsAlive = false

	return merged
}

// Crossover performs recombination between two entities
func Crossover(parent1, parent2 *Entity, childID int, species string) *Entity {
	// Calculate position between parents
	childPos := Position{
		X: (parent1.Position.X + parent2.Position.X) / 2.0,
		Y: (parent1.Position.Y + parent2.Position.Y) / 2.0,
	}

	child := &Entity{
		ID:         childID,
		Traits:     make(map[string]Trait),
		Fitness:    0.0,
		Position:   childPos,
		Energy:     (parent1.Energy + parent2.Energy) / 2.0,
		Age:        0,
		IsAlive:    true,
		Species:    species,
		Generation: int(math.Max(float64(parent1.Generation), float64(parent2.Generation))) + 1,
	}

	// Get all trait names from both parents
	traitNames := make(map[string]bool)
	for name := range parent1.Traits {
		traitNames[name] = true
	}
	for name := range parent2.Traits {
		traitNames[name] = true
	}

	// For each trait, randomly choose from one parent or take average
	for name := range traitNames {
		val1 := parent1.GetTrait(name)
		val2 := parent2.GetTrait(name)

		var childValue float64
		if rand.Float64() < 0.5 {
			// Take from parent1
			childValue = val1
		} else {
			// Take from parent2
			childValue = val2
		}

		// Sometimes blend the values (25% chance)
		if rand.Float64() < 0.25 {
			childValue = (val1 + val2) / 2.0
		}

		child.SetTrait(name, childValue)
	}

	// Initialize memory systems and inherit preferences from parents
	child.DietaryMemory = NewDietaryMemory()
	child.EnvironmentalMemory = NewEnvironmentalMemory()
	
	// Inherit dietary preferences (averaged from parents)
	if parent1.DietaryMemory != nil && parent2.DietaryMemory != nil {
		child.inheritDietaryPreferences(parent1, parent2)
	}
	
	// Inherit environmental adaptations (averaged from parents) 
	if parent1.EnvironmentalMemory != nil && parent2.EnvironmentalMemory != nil {
		child.inheritEnvironmentalAdaptations(parent1, parent2)
	}

	return child
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Update handles entity aging, energy decay, and natural death
func (e *Entity) Update() {
	if !e.IsAlive {
		return
	}

	e.Age++

	// Update molecular needs (deficiencies increase over time)
	if e.MolecularNeeds != nil {
		e.MolecularNeeds.UpdateDeficiencies(1.0) // 1 time step
	}

	// Natural energy decay (affected by nutritional status)
	baseDecay := 1.0 + float64(e.Age)*0.01
	
	// Molecular nutritional status affects energy decay
	if e.MolecularNeeds != nil {
		nutritionalStatus := e.MolecularNeeds.GetOverallNutritionalStatus()
		// Poor nutrition increases energy decay
		nutritionalMultiplier := 1.0 + (1.0-nutritionalStatus)*0.5
		baseDecay *= nutritionalMultiplier
	}
	
	e.Energy -= baseDecay

	// Die if energy is too low
	if e.Energy <= 0 {
		e.IsAlive = false
		e.Energy = 0
	}

	// Die of old age (based on endurance trait)
	maxAge := int(100 + e.GetTrait("endurance")*50)
	if e.Age > maxAge {
		e.IsAlive = false
	}
}

// SeekNutrition makes the entity move toward the most nutritionally desirable nearby food sources
func (e *Entity) SeekNutrition(plants []*Plant, entities []*Entity, maxDistance float64) (targetX, targetY float64, found bool) {
	if !e.IsAlive || e.MolecularNeeds == nil {
		return 0, 0, false
	}

	bestDesirability := 0.0
	bestX, bestY := 0.0, 0.0
	found = false

	// Check nearby plants
	for _, plant := range plants {
		if !plant.IsAlive || plant.MolecularProfile == nil {
			continue
		}

		// Calculate distance
		dx := plant.Position.X - e.Position.X
		dy := plant.Position.Y - e.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > maxDistance {
			continue
		}

		// Calculate molecular desirability
		desirability := GetMolecularDesirability(plant.MolecularProfile, e.MolecularNeeds)
		
		// Adjust for distance (closer is better)
		adjustedDesirability := desirability * (1.0 - distance/maxDistance*0.3)

		if adjustedDesirability > bestDesirability {
			bestDesirability = adjustedDesirability
			bestX, bestY = plant.Position.X, plant.Position.Y
			found = true
		}
	}

	// Check nearby dead entities (if carnivorous)
	if e.GetTrait("aggression") > 0.0 { // Aggressive entities are more carnivorous
		for _, other := range entities {
			if other == e || other.IsAlive || other.MolecularProfile == nil {
				continue
			}

			// Calculate distance
			dx := other.Position.X - e.Position.X
			dy := other.Position.Y - e.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance > maxDistance {
				continue
			}

			// Calculate molecular desirability for meat
			desirability := GetMolecularDesirability(other.MolecularProfile, e.MolecularNeeds)
			
			// Meat is generally more desirable for carnivores
			if e.GetTrait("aggression") > 0.3 {
				desirability *= 1.5
			}
			
			// Adjust for distance
			adjustedDesirability := desirability * (1.0 - distance/maxDistance*0.3)

			if adjustedDesirability > bestDesirability {
				bestDesirability = adjustedDesirability
				bestX, bestY = other.Position.X, other.Position.Y
				found = true
			}
		}
	}

	return bestX, bestY, found
}

// GetMolecularNutritionalStress returns a value indicating how much nutritional stress the entity is under
func (e *Entity) GetMolecularNutritionalStress() float64 {
	if e.MolecularNeeds == nil {
		return 0.0
	}
	
	return 1.0 - e.MolecularNeeds.GetOverallNutritionalStatus()
}

// CanEatPlant determines if this entity can eat a plant
func (e *Entity) CanEatPlant(plant *Plant) bool {
	if !e.IsAlive || !plant.IsAlive {
		return false
	}

	// Herbivores and omnivores can eat plants
	// Predators can only eat plants if starving
	switch e.Species {
	case "herbivore":
		return true
	case "omnivore":
		return true
	case "predator":
		// Predators can only eat plants when desperate (very low energy)
		return e.Energy < 20
	default:
		return false
	}
}

// EatPlant consumes a plant for energy using molecular system
func (e *Entity) EatPlant(plant *Plant, tick int) bool {
	if !e.CanEatPlant(plant) {
		return false
	}

	// Calculate how much to eat based on entity size and hunger
	baseEatAmount := 10 + e.GetTrait("size")*5
	if e.Energy < 30 {
		baseEatAmount *= 1.5 // Eat more when hungry
	}

	// Use molecular system to determine desirability and consumption
	if plant.MolecularProfile != nil && e.MolecularNeeds != nil && e.MolecularMetabolism != nil {
		// Calculate molecular desirability
		desirability := GetMolecularDesirability(plant.MolecularProfile, e.MolecularNeeds)
		
		// Adjust eating amount based on desirability
		eatAmount := baseEatAmount * (0.5 + desirability*0.5)
		
		// Consume nutrients using molecular system
		energyGained, toxinDamage := plant.MolecularProfile.ConsumeNutrients(
			e.MolecularNeeds, 
			e.MolecularMetabolism, 
			eatAmount/100.0) // Convert to consumption rate
		
		// Apply energy gain and toxin damage
		e.Energy += energyGained * 10.0 // Scale up energy gain
		e.Energy -= toxinDamage
		
		// Traditional plant consumption for backwards compatibility
		nutrition := plant.Consume(eatAmount)
		e.Energy += nutrition * 0.3 // Reduced to avoid double-counting
	} else {
		// Fallback to traditional system if molecular system not available
		nutrition := plant.Consume(baseEatAmount)
		toxicity := plant.GetToxicity()

		// Apply nutrition
		e.Energy += nutrition

		// Apply toxicity damage
		if toxicity > 0 {
			resistance := e.GetTrait("toxin_resistance")
			damage := toxicity * (1.0 - resistance*0.5)
			e.Energy -= damage
		}
	}

	// Eating costs some energy
	e.Energy -= 2

	// Record consumption in dietary memory for feedback loop
	e.recordPlantConsumption(plant, tick)

	return true
}

// CheckStarvation handles starvation effects and potential species evolution
func (e *Entity) CheckStarvation(world *World) {
	if !e.IsAlive || e.Energy > 15 {
		return
	}

	// Apply evolutionary pressure based on current conditions and species
	e.checkEvolutionaryPressure(world)
}

// checkEvolutionaryPressure applies environmental and survival pressure to drive evolution
func (e *Entity) checkEvolutionaryPressure(world *World) {
	if !e.IsAlive {
		return
	}

	// Get current biome
	biome := world.getBiomeAtPosition(e.Position.X, e.Position.Y)
	
	// Different evolutionary pressures based on current species and conditions
	switch e.Species {
	case "microbe":
		e.handleMicrobeEvolution(world, biome)
	case "simple":
		e.handleSimpleOrganismEvolution(world, biome)
	case "predator":
		e.handlePredatorEvolution(world, biome)
	case "herbivore":
		e.handleHerbivoreEvolution(world, biome)
	case "omnivore":
		e.handleOmnivoreEvolution(world, biome)
	}
}

// handleMicrobeEvolution manages evolution from microbes to simple organisms
func (e *Entity) handleMicrobeEvolution(world *World, biome BiomeType) {
	// Microbes can evolve when they have sufficient energy and age
	if e.Energy > 30 && e.Age > 20 {
		// Chance to evolve based on environmental pressure
		evolutionChance := 0.001 // Base chance per tick
		
		// Increase chance based on biome suitability
		switch biome {
		case BiomeWater:
			if e.GetTrait("aquatic_adaptation") > 0.3 {
				evolutionChance *= 3.0 // Good water adaptation
			}
		case BiomePlains, BiomeForest:
			evolutionChance *= 2.0 // Good for basic evolution
		}
		
		if rand.Float64() < evolutionChance {
			e.evolveSpecies("simple", world)
		}
	}
}

// handleSimpleOrganismEvolution manages evolution from simple organisms to complex species
func (e *Entity) handleSimpleOrganismEvolution(world *World, biome BiomeType) {
	// Simple organisms can evolve to herbivores, or specialized forms
	if e.Energy > 40 && e.Age > 30 {
		evolutionChance := 0.0005
		targetSpecies := "herbivore" // Default evolution path
		
		// Environment influences evolution direction
		switch biome {
		case BiomeWater:
			if e.GetTrait("aquatic_adaptation") > 0.0 {
				targetSpecies = "aquatic_herbivore"
				evolutionChance *= 2.0
			}
		case BiomeSoil:
			if e.GetTrait("digging_ability") > -0.5 {
				targetSpecies = "soil_dweller"
				evolutionChance *= 1.5
			}
		case BiomeAir:
			if e.GetTrait("flying_ability") > -0.8 {
				targetSpecies = "aerial_herbivore"
				evolutionChance *= 1.2
			}
		case BiomeForest:
			// High competition might drive predatory evolution
			nearbyEntities := world.getEntitiesNearPosition(e.Position, 10.0)
			if len(nearbyEntities) > 5 {
				targetSpecies = "predator"
				evolutionChance *= 0.8
			}
		}
		
		if rand.Float64() < evolutionChance {
			e.evolveSpecies(targetSpecies, world)
		}
	}
}

// handlePredatorEvolution manages predator species evolution
func (e *Entity) handlePredatorEvolution(world *World, biome BiomeType) {
	if e.Energy < 5 && e.Age > 50 {
		// Check if there are any herbivores or omnivores nearby
		hasPreyNearby := false
		for _, other := range world.AllEntities {
			if other.IsAlive && other.Species != "predator" && e.DistanceTo(other) < 20 {
				hasPreyNearby = true
				break
			}
		}

		// If no prey nearby, consider evolutionary adaptation
		if !hasPreyNearby && rand.Float64() < 0.001 {
			// Environment influences evolution direction
			switch biome {
			case BiomeWater:
				if e.GetTrait("aquatic_adaptation") > -0.5 {
					e.evolveSpecies("aquatic_predator", world)
				} else {
					e.evolveSpecies("omnivore", world)
				}
			case BiomeSoil:
				if e.GetTrait("digging_ability") > -0.3 {
					e.evolveSpecies("underground_predator", world)
				} else {
					e.evolveSpecies("omnivore", world)
				}
			default:
				e.evolveSpecies("omnivore", world)
			}
		}
	}
}

// handleHerbivoreEvolution manages herbivore species evolution  
func (e *Entity) handleHerbivoreEvolution(world *World, biome BiomeType) {
	// Herbivores might evolve under predation pressure
	if e.Energy < 10 && e.Age > 40 {
		nearbyPredators := 0
		for _, other := range world.AllEntities {
			if other.IsAlive && other.Species == "predator" && e.DistanceTo(other) < 15 {
				nearbyPredators++
			}
		}
		
		if nearbyPredators > 2 && rand.Float64() < 0.0005 {
			// High predation pressure - might evolve defenses or become omnivore
			if biome == BiomeWater && e.GetTrait("aquatic_adaptation") > 0 {
				e.evolveSpecies("aquatic_herbivore", world)
			} else if biome == BiomeSoil && e.GetTrait("digging_ability") > 0 {
				e.evolveSpecies("soil_dweller", world)
			} else {
				e.evolveSpecies("omnivore", world)
			}
		}
	}
}

// handleOmnivoreEvolution manages omnivore species evolution
func (e *Entity) handleOmnivoreEvolution(world *World, biome BiomeType) {
	// Omnivores are generally stable but can specialize
	if e.Energy > 60 && e.Age > 60 {
		evolutionChance := 0.0001
		
		switch biome {
		case BiomeWater:
			if e.GetTrait("aquatic_adaptation") > 0.5 {
				evolutionChance *= 2.0
				if rand.Float64() < evolutionChance {
					e.evolveSpecies("aquatic_omnivore", world)
				}
			}
		case BiomeSoil:
			if e.GetTrait("digging_ability") > 0.5 {
				evolutionChance *= 1.5
				if rand.Float64() < evolutionChance {
					e.evolveSpecies("underground_omnivore", world)
				}
			}
		case BiomeAir:
			if e.GetTrait("flying_ability") > 0.3 {
				evolutionChance *= 1.2
				if rand.Float64() < evolutionChance {
					e.evolveSpecies("aerial_omnivore", world)
				}
			}
		}
	}
}

// evolveSpecies changes an entity's species under evolutionary pressure
func (e *Entity) evolveSpecies(newSpecies string, world *World) {
	if e.Species == newSpecies {
		return
	}

	oldSpecies := e.Species
	e.Species = newSpecies

	// Adjust traits for new species
	switch newSpecies {
	case "simple":
		// Evolution from microbe to simple organism
		e.SetTrait("size", e.GetTrait("size")+0.3)
		e.SetTrait("intelligence", e.GetTrait("intelligence")+0.2)
		e.SetTrait("speed", e.GetTrait("speed")+0.2)
		e.SetTrait("strength", e.GetTrait("strength")+0.2)

	case "herbivore":
		// Evolution to herbivore
		e.SetTrait("size", e.GetTrait("size")+0.2)
		e.SetTrait("intelligence", e.GetTrait("intelligence")+0.3)
		e.SetTrait("cooperation", e.GetTrait("cooperation")+0.4)
		e.SetTrait("aggression", e.GetTrait("aggression")-0.2)

	case "aquatic_herbivore":
		// Specialized aquatic herbivore
		e.SetTrait("aquatic_adaptation", e.GetTrait("aquatic_adaptation")+0.5)
		e.SetTrait("size", e.GetTrait("size")+0.1)
		e.SetTrait("speed", e.GetTrait("speed")+0.3) // Fast swimming
		e.SetTrait("cooperation", e.GetTrait("cooperation")+0.3)

	case "soil_dweller":
		// Specialized soil-dwelling organism
		e.SetTrait("digging_ability", e.GetTrait("digging_ability")+0.6)
		e.SetTrait("underground_nav", e.GetTrait("underground_nav")+0.5)
		e.SetTrait("size", e.GetTrait("size")-0.2) // Smaller for tunneling
		e.SetTrait("endurance", e.GetTrait("endurance")+0.3)

	case "aerial_herbivore":
		// Specialized aerial herbivore  
		e.SetTrait("flying_ability", e.GetTrait("flying_ability")+0.7)
		e.SetTrait("altitude_tolerance", e.GetTrait("altitude_tolerance")+0.6)
		e.SetTrait("size", e.GetTrait("size")-0.3) // Lighter for flying
		e.SetTrait("speed", e.GetTrait("speed")+0.4)

	case "aquatic_predator":
		// Aquatic predator evolution
		e.SetTrait("aquatic_adaptation", e.GetTrait("aquatic_adaptation")+0.6)
		e.SetTrait("aggression", e.GetTrait("aggression")+0.3)
		e.SetTrait("strength", e.GetTrait("strength")+0.2)
		e.SetTrait("speed", e.GetTrait("speed")+0.4)

	case "underground_predator":
		// Underground predator evolution
		e.SetTrait("digging_ability", e.GetTrait("digging_ability")+0.7)
		e.SetTrait("underground_nav", e.GetTrait("underground_nav")+0.6)
		e.SetTrait("aggression", e.GetTrait("aggression")+0.2)
		e.SetTrait("strength", e.GetTrait("strength")+0.3)

	case "omnivore":
		// Standard omnivore evolution
		e.SetTrait("diet_flexibility", e.GetTrait("diet_flexibility")+0.3)
		e.SetTrait("toxin_resistance", e.GetTrait("toxin_resistance")+0.2)
		e.SetTrait("aggression", e.GetTrait("aggression")-0.1)
		e.SetTrait("intelligence", e.GetTrait("intelligence")+0.2)

	case "aquatic_omnivore":
		// Aquatic omnivore
		e.SetTrait("aquatic_adaptation", e.GetTrait("aquatic_adaptation")+0.5)
		e.SetTrait("diet_flexibility", e.GetTrait("diet_flexibility")+0.4)
		e.SetTrait("intelligence", e.GetTrait("intelligence")+0.3)

	case "underground_omnivore":
		// Underground omnivore
		e.SetTrait("digging_ability", e.GetTrait("digging_ability")+0.5)
		e.SetTrait("underground_nav", e.GetTrait("underground_nav")+0.4)
		e.SetTrait("diet_flexibility", e.GetTrait("diet_flexibility")+0.3)

	case "aerial_omnivore":
		// Aerial omnivore
		e.SetTrait("flying_ability", e.GetTrait("flying_ability")+0.6)
		e.SetTrait("altitude_tolerance", e.GetTrait("altitude_tolerance")+0.5)
		e.SetTrait("diet_flexibility", e.GetTrait("diet_flexibility")+0.3)
		e.SetTrait("size", e.GetTrait("size")-0.2)

	case "predator":
		// Standard predator evolution
		e.SetTrait("aggression", e.GetTrait("aggression")+0.3)
		e.SetTrait("strength", e.GetTrait("strength")+0.2)
		e.SetTrait("speed", e.GetTrait("speed")+0.1)
		e.SetTrait("intelligence", e.GetTrait("intelligence")+0.2)
	}

	// Ensure traits stay within bounds
	for name, trait := range e.Traits {
		value := math.Max(-2.0, math.Min(2.0, trait.Value))
		e.SetTrait(name, value)
	}

	// Log the evolution
	details := fmt.Sprintf("Evolved from %s due to environmental pressure (energy: %.1f)", oldSpecies, e.Energy)
	world.EventLogger.LogSpeciesEvolution(world.Tick, newSpecies, oldSpecies, details)
}

// NewDietaryMemory creates a new dietary memory system for tracking feeding patterns
func NewDietaryMemory() *DietaryMemory {
	return &DietaryMemory{
		PlantTypePreferences:   make(map[int]float64),
		PreySpeciesPreferences: make(map[string]float64),
		ConsumptionHistory:     make([]ConsumptionRecord, 0, 100), // Keep last 100 records
		DietaryFitness:         1.0, // Start neutral
	}
}

// NewEnvironmentalMemory creates a new environmental memory system for tracking pressures
func NewEnvironmentalMemory() *EnvironmentalMemory {
	return &EnvironmentalMemory{
		BiomeExposure:     make(map[BiomeType]float64),
		TemperaturePressure: 0.0,
		RadiationPressure: 0.0,
		SeasonalPressure:  make(map[string]float64),
		EventPressure:     make(map[string]float64),
		AdaptationFitness: 1.0, // Start neutral
	}
}

// inheritDietaryPreferences inherits dietary preferences from parents with some variation
func (e *Entity) inheritDietaryPreferences(parent1, parent2 *Entity) {
	// Inherit plant preferences
	allPlantTypes := make(map[int]bool)
	for plantType := range parent1.DietaryMemory.PlantTypePreferences {
		allPlantTypes[plantType] = true
	}
	for plantType := range parent2.DietaryMemory.PlantTypePreferences {
		allPlantTypes[plantType] = true
	}
	
	for plantType := range allPlantTypes {
		pref1 := parent1.DietaryMemory.PlantTypePreferences[plantType]
		pref2 := parent2.DietaryMemory.PlantTypePreferences[plantType]
		avgPref := (pref1 + pref2) / 2.0
		
		// Add some variation (mutation in dietary preference)
		variation := (rand.Float64() - 0.5) * 0.2
		inheritedPref := math.Max(0.0, math.Min(2.0, avgPref + variation))
		
		if inheritedPref > 0.1 { // Only inherit significant preferences
			e.DietaryMemory.PlantTypePreferences[plantType] = inheritedPref
		}
	}
	
	// Inherit prey species preferences  
	allPreySpecies := make(map[string]bool)
	for species := range parent1.DietaryMemory.PreySpeciesPreferences {
		allPreySpecies[species] = true
	}
	for species := range parent2.DietaryMemory.PreySpeciesPreferences {
		allPreySpecies[species] = true
	}
	
	for species := range allPreySpecies {
		pref1 := parent1.DietaryMemory.PreySpeciesPreferences[species]
		pref2 := parent2.DietaryMemory.PreySpeciesPreferences[species]
		avgPref := (pref1 + pref2) / 2.0
		
		// Add some variation
		variation := (rand.Float64() - 0.5) * 0.2
		inheritedPref := math.Max(0.0, math.Min(2.0, avgPref + variation))
		
		if inheritedPref > 0.1 {
			e.DietaryMemory.PreySpeciesPreferences[species] = inheritedPref
		}
	}
}

// inheritEnvironmentalAdaptations inherits environmental adaptations from parents
func (e *Entity) inheritEnvironmentalAdaptations(parent1, parent2 *Entity) {
	// Inherit biome exposure patterns (averaged)
	allBiomes := make(map[BiomeType]bool)
	for biome := range parent1.EnvironmentalMemory.BiomeExposure {
		allBiomes[biome] = true
	}
	for biome := range parent2.EnvironmentalMemory.BiomeExposure {
		allBiomes[biome] = true
	}
	
	for biome := range allBiomes {
		exp1 := parent1.EnvironmentalMemory.BiomeExposure[biome]
		exp2 := parent2.EnvironmentalMemory.BiomeExposure[biome]
		avgExp := (exp1 + exp2) / 2.0
		
		if avgExp > 0.05 { // Only inherit significant exposures
			e.EnvironmentalMemory.BiomeExposure[biome] = avgExp
		}
	}
	
	// Inherit environmental pressures (averaged with some variation)
	e.EnvironmentalMemory.TemperaturePressure = (parent1.EnvironmentalMemory.TemperaturePressure + parent2.EnvironmentalMemory.TemperaturePressure) / 2.0
	e.EnvironmentalMemory.RadiationPressure = (parent1.EnvironmentalMemory.RadiationPressure + parent2.EnvironmentalMemory.RadiationPressure) / 2.0
	
	// Inherit seasonal pressure patterns
	allSeasons := make(map[string]bool)
	for season := range parent1.EnvironmentalMemory.SeasonalPressure {
		allSeasons[season] = true
	}
	for season := range parent2.EnvironmentalMemory.SeasonalPressure {
		allSeasons[season] = true
	}
	
	for season := range allSeasons {
		press1 := parent1.EnvironmentalMemory.SeasonalPressure[season]
		press2 := parent2.EnvironmentalMemory.SeasonalPressure[season]
		avgPress := (press1 + press2) / 2.0
		
		if avgPress > 0.05 {
			e.EnvironmentalMemory.SeasonalPressure[season] = avgPress
		}
	}
}

// recordPlantConsumption records a plant consumption event and updates dietary preferences
func (e *Entity) recordPlantConsumption(plant *Plant, tick int) {
	if e.DietaryMemory == nil {
		return
	}
	
	// Calculate nutritional value from the consumption
	var nutrition, toxicity float64
	if plant.MolecularProfile != nil {
		// Use molecular profile for accurate nutrition calculation
		desirability := 0.0
		if e.MolecularNeeds != nil {
			desirability = GetMolecularDesirability(plant.MolecularProfile, e.MolecularNeeds)
		}
		nutrition = desirability
		toxicity = plant.MolecularProfile.Toxicity
	} else {
		// Fallback estimation
		nutrition = 0.5
		toxicity = plant.GetToxicity()
	}
	
	// Record consumption history
	record := ConsumptionRecord{
		Tick:     tick,
		FoodType: "plant",
		FoodID:   fmt.Sprintf("plant_%d", int(plant.Type)),
		Nutrition: nutrition,
		Toxicity: toxicity,
	}
	
	// Add to history (maintain rolling window)
	e.DietaryMemory.ConsumptionHistory = append(e.DietaryMemory.ConsumptionHistory, record)
	if len(e.DietaryMemory.ConsumptionHistory) > 100 {
		e.DietaryMemory.ConsumptionHistory = e.DietaryMemory.ConsumptionHistory[1:]
	}
	
	// Update plant type preference based on success
	plantType := int(plant.Type)
	currentPref := e.DietaryMemory.PlantTypePreferences[plantType]
	
	// Successful feeding (high nutrition, low toxicity) increases preference
	success := nutrition - toxicity*0.5
	preferenceChange := success * 0.1 // Gradual adaptation
	
	newPref := math.Max(0.0, math.Min(2.0, currentPref + preferenceChange))
	e.DietaryMemory.PlantTypePreferences[plantType] = newPref
	
	// Update overall dietary fitness based on recent consumption patterns
	e.updateDietaryFitness()
}

// recordEntityConsumption records an entity consumption event and updates prey preferences
func (e *Entity) recordEntityConsumption(prey *Entity, tick int) {
	if e.DietaryMemory == nil {
		return
	}
	
	// Calculate nutritional value
	var nutrition, toxicity float64
	if prey.MolecularProfile != nil && e.MolecularNeeds != nil {
		desirability := GetMolecularDesirability(prey.MolecularProfile, e.MolecularNeeds)
		nutrition = desirability
		toxicity = prey.MolecularProfile.Toxicity
	} else {
		// Fallback estimation based on prey size
		nutrition = (prey.GetTrait("size") + 1.0) * 0.2
		toxicity = 0.0
	}
	
	// Record consumption history
	record := ConsumptionRecord{
		Tick:     tick,
		FoodType: "entity",
		FoodID:   prey.Species,
		Nutrition: nutrition,
		Toxicity: toxicity,
	}
	
	// Add to history
	e.DietaryMemory.ConsumptionHistory = append(e.DietaryMemory.ConsumptionHistory, record)
	if len(e.DietaryMemory.ConsumptionHistory) > 100 {
		e.DietaryMemory.ConsumptionHistory = e.DietaryMemory.ConsumptionHistory[1:]
	}
	
	// Update prey species preference
	species := prey.Species
	currentPref := e.DietaryMemory.PreySpeciesPreferences[species]
	
	success := nutrition - toxicity*0.5
	preferenceChange := success * 0.1
	
	newPref := math.Max(0.0, math.Min(2.0, currentPref + preferenceChange))
	e.DietaryMemory.PreySpeciesPreferences[species] = newPref
	
	// Update overall dietary fitness
	e.updateDietaryFitness()
}

// updateDietaryFitness calculates how well-adapted the entity is to its current diet
func (e *Entity) updateDietaryFitness() {
	if e.DietaryMemory == nil || len(e.DietaryMemory.ConsumptionHistory) == 0 {
		e.DietaryMemory.DietaryFitness = 1.0
		return
	}
	
	// Calculate fitness based on recent consumption success
	recentHistory := e.DietaryMemory.ConsumptionHistory
	if len(recentHistory) > 20 {
		recentHistory = recentHistory[len(recentHistory)-20:] // Last 20 consumptions
	}
	
	totalSuccess := 0.0
	for _, record := range recentHistory {
		success := record.Nutrition - record.Toxicity*0.5
		totalSuccess += success
	}
	
	avgSuccess := totalSuccess / float64(len(recentHistory))
	// Normalize to 0.0-2.0 range (fitness can be above 1.0 for very well adapted)
	e.DietaryMemory.DietaryFitness = math.Max(0.0, math.Min(2.0, avgSuccess + 1.0))
}

// trackEnvironmentalExposure updates environmental memory based on current conditions
func (e *Entity) trackEnvironmentalExposure(biome BiomeType, season string, event *WorldEvent, tick int) {
	if e.EnvironmentalMemory == nil {
		return
	}
	
	// Track biome exposure
	e.EnvironmentalMemory.BiomeExposure[biome] += 0.01 // Small increment per tick
	
	// Track seasonal pressure
	seasonalStress := 0.0
	switch season {
	case "Winter":
		seasonalStress = 0.2 // Winter is generally stressful
	case "Summer":
		seasonalStress = 0.1 // Mild stress
	case "Spring", "Autumn":
		seasonalStress = 0.05 // Low stress
	}
	
	current := e.EnvironmentalMemory.SeasonalPressure[season]
	e.EnvironmentalMemory.SeasonalPressure[season] = current + seasonalStress*0.1
	
	// Track environmental events
	if event != nil {
		eventStress := event.GlobalDamage * 0.1
		current := e.EnvironmentalMemory.EventPressure[event.Name]
		e.EnvironmentalMemory.EventPressure[event.Name] = current + eventStress
		
		// Add to radiation pressure if it's a radiation event
		if event.GlobalMutation > 0 {
			e.EnvironmentalMemory.RadiationPressure += event.GlobalMutation * 0.1
		}
	}
	
	// Update adaptation fitness based on environmental match
	e.updateEnvironmentalFitness(biome)
}

// updateEnvironmentalFitness calculates how well-adapted the entity is to current environment
func (e *Entity) updateEnvironmentalFitness(currentBiome BiomeType) {
	if e.EnvironmentalMemory == nil {
		e.EnvironmentalMemory.AdaptationFitness = 1.0
		return
	}
	
	// Check how well entity's traits match the current environment
	fitness := 1.0
	
	switch currentBiome {
	case BiomeWater:
		aquaticAdaptation := e.GetTrait("aquatic_adaptation")
		fitness += aquaticAdaptation * 0.5 // Well-adapted entities get bonus
	case BiomeSoil:
		diggingAbility := e.GetTrait("digging_ability")
		undergroundNav := e.GetTrait("underground_nav")
		fitness += (diggingAbility + undergroundNav) * 0.25
	case BiomeAir:
		flyingAbility := e.GetTrait("flying_ability")
		altitudeTolerance := e.GetTrait("altitude_tolerance")
		fitness += (flyingAbility + altitudeTolerance) * 0.25
	case BiomeRadiation:
		endurance := e.GetTrait("endurance")
		fitness += endurance * 0.3 // Endurance helps with radiation
	case BiomeDesert:
		endurance := e.GetTrait("endurance")
		fitness += endurance * 0.4 // Endurance crucial in desert
	}
	
	// Factor in accumulated environmental pressure
	totalPressure := e.EnvironmentalMemory.TemperaturePressure + e.EnvironmentalMemory.RadiationPressure
	fitness -= totalPressure * 0.1 // High pressure reduces fitness
	
	e.EnvironmentalMemory.AdaptationFitness = math.Max(0.0, math.Min(2.0, fitness))
}

// biasedMutation applies directional bias to mutations based on feedback loops
func (e *Entity) biasedMutation(traitName string, mutation float64) float64 {
	// Start with the base mutation
	biasedMutation := mutation
	biasStrength := 0.8 // How strongly to bias the mutation
	
	// Environmental bias - encourage traits that help with current environment exposure
	if e.EnvironmentalMemory != nil {
		for biome, exposure := range e.EnvironmentalMemory.BiomeExposure {
			if exposure > 0.2 { // Significant exposure to this biome
				currentValue := e.GetTrait(traitName)
				switch biome {
				case BiomeWater:
					if traitName == "aquatic_adaptation" && currentValue < 1.0 {
						// Bias toward positive aquatic adaptation
						bias := math.Abs(mutation) * biasStrength * exposure
						if mutation < 0 && currentValue < 0.5 {
							// Flip negative mutations to positive when trait is poor
							biasedMutation = bias
						} else if mutation > 0 {
							// Amplify positive mutations
							biasedMutation += bias
						}
					}
				case BiomeSoil:
					if (traitName == "digging_ability" || traitName == "underground_nav") && currentValue < 1.0 {
						bias := math.Abs(mutation) * biasStrength * exposure
						if mutation < 0 && currentValue < 0.5 {
							biasedMutation = bias
						} else if mutation > 0 {
							biasedMutation += bias
						}
					}
				case BiomeAir:
					if (traitName == "flying_ability" || traitName == "altitude_tolerance") && currentValue < 1.0 {
						bias := math.Abs(mutation) * biasStrength * exposure
						if mutation < 0 && currentValue < 0.5 {
							biasedMutation = bias
						} else if mutation > 0 {
							biasedMutation += bias
						}
					}
				case BiomeRadiation:
					if traitName == "endurance" && currentValue < 1.0 {
						bias := math.Abs(mutation) * biasStrength * exposure
						if mutation < 0 && currentValue < 0.5 {
							biasedMutation = bias
						} else if mutation > 0 {
							biasedMutation += bias
						}
					}
				case BiomeDesert:
					if traitName == "endurance" && currentValue < 1.0 {
						bias := math.Abs(mutation) * biasStrength * exposure
						if mutation < 0 && currentValue < 0.5 {
							biasedMutation = bias
						} else if mutation > 0 {
							biasedMutation += bias
						}
					}
				}
			}
		}
	}
	
	// Dietary bias - encourage traits that help with diet specialization or diversification
	if e.DietaryMemory != nil {
		// If entity has strong preferences, bias toward supporting traits
		strongPlantPrefs := 0
		for _, pref := range e.DietaryMemory.PlantTypePreferences {
			if pref > 1.3 {
				strongPlantPrefs++
			}
		}
		
		strongPreyPrefs := 0
		for _, pref := range e.DietaryMemory.PreySpeciesPreferences {
			if pref > 1.3 {
				strongPreyPrefs++
			}
		}
		
		currentValue := e.GetTrait(traitName)
		
		// If specialized (strong preferences), bias toward supporting traits
		if strongPlantPrefs > 0 {
			if traitName == "toxin_resistance" && currentValue < 1.0 {
				bias := math.Abs(mutation) * biasStrength * 0.5
				if mutation < 0 && currentValue < 0.5 {
					biasedMutation = bias
				} else if mutation > 0 {
					biasedMutation += bias
				}
			}
		}
		
		if strongPreyPrefs > 0 {
			if (traitName == "aggression" || traitName == "strength" || traitName == "speed") && currentValue < 1.0 {
				bias := math.Abs(mutation) * biasStrength * 0.4
				if mutation < 0 && currentValue < 0.5 {
					biasedMutation = bias
				} else if mutation > 0 {
					biasedMutation += bias
				}
			}
		}
		
		// If dietary fitness is low, bias toward flexibility traits
		if e.DietaryMemory.DietaryFitness < 0.7 {
			if traitName == "diet_flexibility" && currentValue < 1.0 {
				bias := math.Abs(mutation) * biasStrength * (0.7 - e.DietaryMemory.DietaryFitness)
				if mutation < 0 && currentValue < 0.5 {
					biasedMutation = bias
				} else if mutation > 0 {
					biasedMutation += bias
				}
			}
		}
	}
	
	return biasedMutation
}

// Statistical tracking helper methods for entities

// SetEnergyWithTracking sets entity energy and logs the change for statistical analysis
func (e *Entity) SetEnergyWithTracking(newEnergy float64, world *World, reason string) {
	if world.StatisticalReporter != nil {
		oldEnergy := e.Energy
		world.StatisticalReporter.LogEntityEvent(world.Tick, "energy_change", e, oldEnergy, newEnergy, nil)
		
		// Add detailed metadata about the change
		metadata := map[string]interface{}{
			"reason":     reason,
			"magnitude": newEnergy - oldEnergy,
		}
		world.StatisticalReporter.LogSystemEvent(world.Tick, "entity_energy_change", reason, metadata)
	}
	e.Energy = newEnergy
}

// ModifyEnergyWithTracking modifies entity energy by a delta amount and logs the change
func (e *Entity) ModifyEnergyWithTracking(delta float64, world *World, reason string) {
	newEnergy := e.Energy + delta
	e.SetEnergyWithTracking(newEnergy, world, reason)
}

// SetTraitWithTracking sets a trait value and logs the change for statistical analysis
func (e *Entity) SetTraitWithTracking(name string, value float64, world *World, reason string) {
	if world.StatisticalReporter != nil {
		oldValue := e.GetTrait(name)
		e.SetTrait(name, value)
		world.StatisticalReporter.LogEntityEvent(world.Tick, "trait_change", e, oldValue, value, nil)
		
		// Add detailed metadata
		metadata := map[string]interface{}{
			"trait":      name,
			"reason":     reason,
			"magnitude":  value - oldValue,
		}
		world.StatisticalReporter.LogSystemEvent(world.Tick, "entity_trait_change", reason, metadata)
	} else {
		e.SetTrait(name, value)
	}
}

// LogEntityDeath logs when an entity dies with context about the cause
func (e *Entity) LogEntityDeath(world *World, cause string, contributingFactors map[string]interface{}) {
	if world.StatisticalReporter != nil {
		metadata := map[string]interface{}{
			"cause":     cause,
			"age":       e.Age,
			"energy":    e.Energy,
			"species":   e.Species,
			"factors":   contributingFactors,
		}
		world.StatisticalReporter.LogEntityEvent(world.Tick, "entity_death", e, true, false, nil)
		world.StatisticalReporter.LogSystemEvent(world.Tick, "entity_death", cause, metadata)
	}
	e.IsAlive = false
}

// LogEntityBirth logs when an entity is born with parental information
func (e *Entity) LogEntityBirth(world *World, parent1, parent2 *Entity) {
	if world.StatisticalReporter != nil {
		metadata := map[string]interface{}{
			"parent1_id":      parent1.ID,
			"parent1_species": parent1.Species,
			"parent1_fitness": parent1.Fitness,
		}
		
		if parent2 != nil {
			metadata["parent2_id"] = parent2.ID
			metadata["parent2_species"] = parent2.Species
			metadata["parent2_fitness"] = parent2.Fitness
		}
		
		var impactedEntities []*Entity
		if parent2 != nil {
			impactedEntities = []*Entity{parent1, parent2}
		} else {
			impactedEntities = []*Entity{parent1}
		}
		
		world.StatisticalReporter.LogEntityEvent(world.Tick, "entity_birth", e, false, true, impactedEntities)
		world.StatisticalReporter.LogSystemEvent(world.Tick, "entity_birth", "reproduction", metadata)
	}
}
