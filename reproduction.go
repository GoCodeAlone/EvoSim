package main

import (
	"math/rand"
	"time"
)

// ReproductionMode represents different ways entities can reproduce
type ReproductionMode int

const (
	DirectCoupling ReproductionMode = iota // Immediate offspring creation
	EggLaying                              // Laying eggs that hatch later
	LiveBirth                              // Gestation period then live birth
	Budding                                // Asexual reproduction through budding
	Fission                                // Splitting into multiple entities
)

// String returns the string representation of ReproductionMode
func (rm ReproductionMode) String() string {
	switch rm {
	case DirectCoupling:
		return "direct_coupling"
	case EggLaying:
		return "egg_laying"
	case LiveBirth:
		return "live_birth"
	case Budding:
		return "budding"
	case Fission:
		return "fission"
	default:
		return "unknown"
	}
}

// MatingStrategy represents how entities approach mating
type MatingStrategy int

const (
	Monogamous MatingStrategy = iota // Single mate for life
	Polygamous                       // Multiple mates
	Sequential                       // One mate at a time, can change
	Promiscuous                      // No preference, any compatible mate
)

// String returns the string representation of MatingStrategy
func (ms MatingStrategy) String() string {
	switch ms {
	case Monogamous:
		return "monogamous"
	case Polygamous:
		return "polygamous"
	case Sequential:
		return "sequential"
	case Promiscuous:
		return "promiscuous"
	default:
		return "unknown"
	}
}

// ReproductionStatus tracks an entity's current reproductive state
type ReproductionStatus struct {
	Mode               ReproductionMode `json:"mode"`
	Strategy           MatingStrategy   `json:"strategy"`
	IsPregnant         bool             `json:"is_pregnant"`
	GestationStartTick int              `json:"gestation_start_tick"`
	GestationPeriod    int              `json:"gestation_period"`
	Mate               *Entity          `json:"-"` // Current or preferred mate (exclude from JSON to avoid cycles)
	MateID             int              `json:"mate_id"`
	OffspringCount     int              `json:"offspring_count"`
	LastMatingTick     int              `json:"last_mating_tick"`
	MatingLocation     Position         `json:"mating_location"`
	PreferredMatingLocation Position    `json:"preferred_mating_location"` // Location entity prefers to mate at
	ReadyToMate        bool             `json:"ready_to_mate"`
	MatingSeason       bool             `json:"mating_season"`
	MigrationDistance  float64          `json:"migration_distance"`        // How far entity will travel to mate
	RequiresMigration  bool             `json:"requires_migration"`        // Whether entity needs to migrate for mating
}

// Egg represents an egg that can hatch into an entity
type Egg struct {
	ID             int      `json:"id"`
	Position       Position `json:"position"`
	Parent1ID      int      `json:"parent1_id"`
	Parent2ID      int      `json:"parent2_id"`
	LayingTick     int      `json:"laying_tick"`
	HatchingPeriod int      `json:"hatching_period"`
	Energy         float64  `json:"energy"`
	IsViable       bool     `json:"is_viable"`
	Species        string   `json:"species"`
}

// DecayableItem represents an item that can decay over time
type DecayableItem struct {
	ID           int      `json:"id"`
	Position     Position `json:"position"`
	ItemType     string   `json:"item_type"` // "corpse", "fruit", "organic_matter"
	CreationTick int      `json:"creation_tick"`
	DecayPeriod  int      `json:"decay_period"`
	NutrientValue float64  `json:"nutrient_value"`
	IsDecayed    bool     `json:"is_decayed"`
	OriginSpecies string   `json:"origin_species"`
	Size         float64  `json:"size"`
}

// ReproductionSystem manages reproduction, gestation, and decay processes
type ReproductionSystem struct {
	Eggs          []*Egg          `json:"eggs"`
	DecayingItems []*DecayableItem `json:"decaying_items"`
	NextEggID     int             `json:"next_egg_id"`
	NextItemID    int             `json:"next_item_id"`
}

// NewReproductionSystem creates a new reproduction system
func NewReproductionSystem() *ReproductionSystem {
	return &ReproductionSystem{
		Eggs:          make([]*Egg, 0),
		DecayingItems: make([]*DecayableItem, 0),
		NextEggID:     1,
		NextItemID:    1,
	}
}

// NewReproductionStatus creates a new reproduction status for an entity
func NewReproductionStatus() *ReproductionStatus {
	// Randomly assign reproduction mode and strategy based on entity traits
	mode := ReproductionMode(rand.Intn(5))
	strategy := MatingStrategy(rand.Intn(4))
	
	return &ReproductionStatus{
		Mode:            mode,
		Strategy:        strategy,
		IsPregnant:      false,
		GestationPeriod: 50 + rand.Intn(100), // 50-150 ticks
		OffspringCount:  0,
		ReadyToMate:     true,
		MatingSeason:    true,
		MigrationDistance: 10.0 + rand.Float64()*20.0, // 10-30 units
		RequiresMigration: rand.Float64() < 0.3,        // 30% chance of requiring migration
		PreferredMatingLocation: Position{
			X: rand.Float64() * 100.0, // Random preferred location
			Y: rand.Float64() * 100.0,
		},
	}
}

// CanMate determines if an entity can mate with another
func (rs *ReproductionStatus) CanMate(other *ReproductionStatus, otherEntityID int, currentTick int) bool {
	if !rs.ReadyToMate || !rs.MatingSeason {
		return false
	}
	
	if rs.IsPregnant {
		return false
	}
	
	// Check strategy-specific constraints
	switch rs.Strategy {
	case Monogamous:
		// Can only mate if no current mate or mate is the same entity
		return rs.MateID == 0 || rs.MateID == otherEntityID
	case Sequential:
		// Need time between different mates
		if rs.LastMatingTick > 0 && currentTick-rs.LastMatingTick < 100 {
			return rs.MateID == otherEntityID // Can mate with same partner
		}
		return true
	default:
		return true
	}
}

// StartMating initiates the mating process between two entities
func (rs *ReproductionSystem) StartMating(entity1, entity2 *Entity, currentTick int) bool {
	if entity1.ReproductionStatus == nil || entity2.ReproductionStatus == nil {
		return false
	}
	
	if !entity1.ReproductionStatus.CanMate(entity2.ReproductionStatus, entity2.ID, currentTick) {
		return false
	}
	
	if !entity2.ReproductionStatus.CanMate(entity1.ReproductionStatus, entity1.ID, currentTick) {
		return false
	}
	
	// Record mating
	entity1.ReproductionStatus.LastMatingTick = currentTick
	entity2.ReproductionStatus.LastMatingTick = currentTick
	entity1.ReproductionStatus.MateID = entity2.ID
	entity2.ReproductionStatus.MateID = entity1.ID
	
	// Determine reproduction outcome based on mode
	switch entity1.ReproductionStatus.Mode {
	case DirectCoupling:
		// Immediate offspring - use existing crossover
		return true // Handled by calling code
	case EggLaying:
		return rs.LayEgg(entity1, entity2, currentTick)
	case LiveBirth:
		return rs.StartGestation(entity1, entity2, currentTick)
	case Budding:
		return rs.Bud(entity1, currentTick)
	case Fission:
		return rs.Split(entity1, currentTick)
	}
	
	return false
}

// LayEgg creates an egg from two parents
func (rs *ReproductionSystem) LayEgg(parent1, parent2 *Entity, currentTick int) bool {
	// Choose location between parents with some variation
	eggPos := Position{
		X: (parent1.Position.X + parent2.Position.X) / 2.0 + (rand.Float64()-0.5)*5.0,
		Y: (parent1.Position.Y + parent2.Position.Y) / 2.0 + (rand.Float64()-0.5)*5.0,
	}
	
	egg := &Egg{
		ID:             rs.NextEggID,
		Position:       eggPos,
		Parent1ID:      parent1.ID,
		Parent2ID:      parent2.ID,
		LayingTick:     currentTick,
		HatchingPeriod: 30 + rand.Intn(70), // 30-100 ticks to hatch
		Energy:         (parent1.Energy + parent2.Energy) * 0.2, // Inherit some energy
		IsViable:       true,
		Species:        parent1.Species,
	}
	
	rs.Eggs = append(rs.Eggs, egg)
	rs.NextEggID++
	
	// Parents lose some energy
	parent1.Energy -= 15.0
	parent2.Energy -= 10.0
	
	return true
}

// StartGestation begins the gestation period for live birth
func (rs *ReproductionSystem) StartGestation(parent1, parent2 *Entity, currentTick int) bool {
	// Usually the first parent carries the offspring
	parent1.ReproductionStatus.IsPregnant = true
	parent1.ReproductionStatus.GestationStartTick = currentTick
	
	// Store mating location for potential migration behavior
	parent1.ReproductionStatus.MatingLocation = parent1.Position
	
	// Energy cost for gestation start
	parent1.Energy -= 20.0
	parent2.Energy -= 10.0
	
	return true
}

// Bud creates offspring through asexual budding
func (rs *ReproductionSystem) Bud(parent *Entity, currentTick int) bool {
	if parent.Energy < 50.0 {
		return false // Not enough energy
	}
	
	// This will be handled by creating a clone with mutation
	parent.Energy -= 30.0
	return true
}

// Split divides an entity into multiple offspring
func (rs *ReproductionSystem) Split(parent *Entity, currentTick int) bool {
	if parent.Energy < 80.0 {
		return false // Not enough energy
	}
	
	// This will create multiple offspring
	parent.Energy -= 60.0
	return true
}

// Update processes reproduction system each tick
func (rs *ReproductionSystem) Update(currentTick int) ([]*Entity, []*DecayableItem) {
	newEntities := make([]*Entity, 0)
	fertilizers := make([]*DecayableItem, 0)
	
	// Process egg hatching
	var remainingEggs []*Egg
	for _, egg := range rs.Eggs {
		if egg.IsViable && currentTick-egg.LayingTick >= egg.HatchingPeriod {
			// Hatch the egg - create new entity
			hatchling := rs.HatchEgg(egg)
			if hatchling != nil {
				newEntities = append(newEntities, hatchling)
			}
		} else if egg.IsViable && currentTick-egg.LayingTick < egg.HatchingPeriod*2 {
			// Keep viable eggs that haven't exceeded maximum incubation time
			remainingEggs = append(remainingEggs, egg)
		}
		// Eggs that are too old or not viable are discarded
	}
	rs.Eggs = remainingEggs
	
	// Process decay
	var remainingItems []*DecayableItem
	for _, item := range rs.DecayingItems {
		if !item.IsDecayed && currentTick-item.CreationTick >= item.DecayPeriod {
			// Item has finished decaying - becomes fertilizer
			item.IsDecayed = true
			fertilizers = append(fertilizers, item)
		} else if !item.IsDecayed {
			// Keep items still decaying
			remainingItems = append(remainingItems, item)
		}
		// Fully decayed items are removed from tracking
	}
	rs.DecayingItems = remainingItems
	
	return newEntities, fertilizers
}

// HatchEgg creates a new entity from an egg
func (rs *ReproductionSystem) HatchEgg(egg *Egg) *Entity {
	// Create new entity at egg position
	hatchling := &Entity{
		ID:       egg.ID, // Reuse egg ID for simplicity
		Position: egg.Position,
		Energy:   egg.Energy,
		Age:      0,
		IsAlive:  true,
		Species:  egg.Species,
	}
	
	// Initialize traits (this will be enhanced when we integrate with existing parents)
	hatchling.Traits = make(map[string]Trait)
	// For now, create random traits - in real use, this would inherit from parents
	traitNames := []string{"strength", "speed", "intelligence", "vision", "aggression"}
	for _, name := range traitNames {
		hatchling.Traits[name] = Trait{
			Name:  name,
			Value: rand.Float64()*2 - 1,
		}
	}
	
	// Initialize reproduction status
	hatchling.ReproductionStatus = NewReproductionStatus()
	
	return hatchling
}

// AddDecayingItem adds an item that will decay over time
func (rs *ReproductionSystem) AddDecayingItem(itemType string, position Position, nutrientValue float64, originSpecies string, size float64, currentTick int) {
	item := &DecayableItem{
		ID:            rs.NextItemID,
		Position:      position,
		ItemType:      itemType,
		CreationTick:  currentTick,
		DecayPeriod:   100 + rand.Intn(200), // 100-300 ticks to decay
		NutrientValue: nutrientValue,
		IsDecayed:     false,
		OriginSpecies: originSpecies,
		Size:          size,
	}
	
	rs.DecayingItems = append(rs.DecayingItems, item)
	rs.NextItemID++
}

// UpdateMatingSeasons updates whether entities are in mating season
func (rs *ReproductionSystem) UpdateMatingSeasons(entities []*Entity, season string) {
	for _, entity := range entities {
		if entity.ReproductionStatus == nil {
			continue
		}
		
		// Determine if it's mating season based on season and entity traits
		switch season {
		case "Spring":
			entity.ReproductionStatus.MatingSeason = true
		case "Summer":
			// Some entities prefer summer mating
			entity.ReproductionStatus.MatingSeason = entity.GetTrait("summer_mating") > 0
		case "Autumn":
			// Migration and final mating before winter
			entity.ReproductionStatus.MatingSeason = entity.GetTrait("autumn_mating") > 0
		case "Winter":
			// Most entities don't mate in winter
			entity.ReproductionStatus.MatingSeason = false
		default:
			entity.ReproductionStatus.MatingSeason = true
		}
	}
}

// CheckGestation checks if any entities are ready to give birth
func (rs *ReproductionSystem) CheckGestation(entities []*Entity, currentTick int) []*Entity {
	newborns := make([]*Entity, 0)
	
	for _, entity := range entities {
		if entity.ReproductionStatus == nil || !entity.ReproductionStatus.IsPregnant {
			continue
		}
		
		gestationTime := currentTick - entity.ReproductionStatus.GestationStartTick
		if gestationTime >= entity.ReproductionStatus.GestationPeriod {
			// Give birth
			offspring := rs.GiveBirth(entity, currentTick)
			if offspring != nil {
				newborns = append(newborns, offspring...)
			}
			
			// Reset pregnancy status
			entity.ReproductionStatus.IsPregnant = false
			entity.ReproductionStatus.GestationStartTick = 0
			entity.ReproductionStatus.OffspringCount++
		}
	}
	
	return newborns
}

// GiveBirth creates offspring from a pregnant entity
func (rs *ReproductionSystem) GiveBirth(parent *Entity, currentTick int) []*Entity {
	offspring := make([]*Entity, 0)
	
	// Number of offspring depends on species and traits
	numOffspring := 1
	if parent.GetTrait("fertility") > 0.5 {
		numOffspring = 2
	}
	if parent.GetTrait("multiple_births") > 0.7 {
		numOffspring = 3
	}
	
	for i := 0; i < numOffspring; i++ {
		// Create offspring near parent
		childPos := Position{
			X: parent.Position.X + (rand.Float64()-0.5)*3.0,
			Y: parent.Position.Y + (rand.Float64()-0.5)*3.0,
		}
		
		child := &Entity{
			ID:         rs.NextEggID, // Reuse ID system
			Position:   childPos,
			Energy:     parent.Energy * 0.3, // Inherit some energy
			Age:        0,
			IsAlive:    true,
			Species:    parent.Species,
			Generation: parent.Generation + 1,
		}
		
		// Initialize traits (simplified for now)
		child.Traits = make(map[string]Trait)
		for name, trait := range parent.Traits {
			// Inherit trait with some mutation
			childValue := trait.Value + (rand.Float64()-0.5)*0.2
			child.Traits[name] = Trait{Name: name, Value: childValue}
		}
		
		// Initialize reproduction status
		child.ReproductionStatus = NewReproductionStatus()
		
		offspring = append(offspring, child)
		rs.NextEggID++
	}
	
	// Parent loses energy from birth
	parent.Energy -= float64(numOffspring) * 25.0
	
	return offspring
}

func init() {
	rand.Seed(time.Now().UnixNano())
}