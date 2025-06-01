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

// Mutate applies random mutations to the entity's traits
func (e *Entity) Mutate(mutationRate float64, mutationStrength float64) {
	for name, trait := range e.Traits {
		if rand.Float64() < mutationRate {
			// Apply Gaussian noise for mutation
			mutation := rand.NormFloat64() * mutationStrength
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

// MoveTo moves the entity towards a target position
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

		// Moving costs energy
		energyCost := speed * 0.1
		e.Energy -= energyCost
	}
}

// MoveRandomly moves the entity in a random direction
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

// Eat consumes another entity for energy
func (e *Entity) Eat(other *Entity) bool {
	if !e.CanEat(other) {
		return false
	}

	// Gain energy based on the consumed entity's size and remaining energy
	energyGain := (other.GetTrait("size")+1.0)*10 + other.Energy*0.5
	e.Energy += energyGain

	// Eating costs some energy
	e.Energy -= 5

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

	// Natural energy decay
	e.Energy -= 1.0 + float64(e.Age)*0.01

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

// EatPlant consumes a plant for energy
func (e *Entity) EatPlant(plant *Plant) bool {
	if !e.CanEatPlant(plant) {
		return false
	}

	// Calculate how much to eat based on entity size and hunger
	eatAmount := 10 + e.GetTrait("size")*5
	if e.Energy < 30 {
		eatAmount *= 1.5 // Eat more when hungry
	}

	// Get nutrition and toxicity
	nutrition := plant.Consume(eatAmount)
	toxicity := plant.GetToxicity()

	// Apply nutrition
	e.Energy += nutrition

	// Apply toxicity damage
	if toxicity > 0 {
		resistance := e.GetTrait("toxin_resistance")
		damage := toxicity * (1.0 - resistance*0.5)
		e.Energy -= damage
	}

	// Eating costs some energy
	e.Energy -= 2

	return true
}

// CheckStarvation handles starvation effects and potential species evolution
func (e *Entity) CheckStarvation(world *World) {
	if !e.IsAlive || e.Energy > 15 {
		return
	}

	// Severe starvation can trigger evolutionary adaptation
	if e.Energy < 5 && e.Species == "predator" {
		// Check if there are any herbivores or omnivores nearby
		hasPreyNearby := false
		for _, other := range world.AllEntities {
			if other.IsAlive && other.Species != "predator" && e.DistanceTo(other) < 20 {
				hasPreyNearby = true
				break
			}
		}

		// If no prey nearby, consider evolutionary pressure
		if !hasPreyNearby && rand.Float64() < 0.01 { // 1% chance per tick
			e.evolveSpecies("omnivore", world)
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
	case "omnivore":
		// Develop omnivore traits
		e.SetTrait("diet_flexibility", e.GetTrait("diet_flexibility")+0.3)
		e.SetTrait("toxin_resistance", e.GetTrait("toxin_resistance")+0.2)
		e.SetTrait("aggression", e.GetTrait("aggression")-0.1) // Less aggressive
	case "herbivore":
		// Develop herbivore traits
		e.SetTrait("digestion_efficiency", e.GetTrait("digestion_efficiency")+0.4)
		e.SetTrait("toxin_resistance", e.GetTrait("toxin_resistance")+0.3)
		e.SetTrait("aggression", e.GetTrait("aggression")-0.3) // Much less aggressive
	case "predator":
		// Develop predator traits
		e.SetTrait("aggression", e.GetTrait("aggression")+0.3)
		e.SetTrait("strength", e.GetTrait("strength")+0.2)
		e.SetTrait("speed", e.GetTrait("speed")+0.1)
	}

	// Log the evolution
	details := fmt.Sprintf("Evolved due to environmental pressure (energy: %.1f)", e.Energy)
	world.EventLogger.LogSpeciesEvolution(world.Tick, newSpecies, oldSpecies, details)
}
