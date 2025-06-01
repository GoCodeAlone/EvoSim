package main

import (
	"math"
	"math/rand"
)

// SignalType represents different types of signals entities can send
type SignalType int

const (
	SignalDanger SignalType = iota
	SignalFood
	SignalMating
	SignalTerritory
	SignalHelp
	SignalMigration
)

// Signal represents a communication signal between entities
type Signal struct {
	Type      SignalType
	Strength  float64                // How strong/far the signal travels
	Position  Position               // Where the signal originates
	Data      map[string]interface{} // Additional signal data
	Decay     float64                // How fast the signal fades
	Range     float64                // Maximum effective range
	Timestamp int                    // When the signal was created
}

// CommunicationSystem manages entity signaling and collaboration
type CommunicationSystem struct {
	Signals    []Signal
	MaxSignals int
}

// NewCommunicationSystem creates a new communication system
func NewCommunicationSystem() *CommunicationSystem {
	return &CommunicationSystem{
		Signals:    make([]Signal, 0),
		MaxSignals: 100, // Limit active signals
	}
}

// SendSignal allows an entity to broadcast a signal
func (cs *CommunicationSystem) SendSignal(entity *Entity, signalType SignalType, data map[string]interface{}) {
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")

	// Only intelligent, cooperative entities can send complex signals
	if intelligence < 0.3 || cooperation < 0.2 {
		return
	}

	signal := Signal{
		Type:      signalType,
		Strength:  0.5 + intelligence*0.5,
		Position:  entity.Position,
		Data:      data,
		Decay:     0.05,                     // Signals fade over time
		Range:     10.0 + intelligence*20.0, // Smarter entities signal farther
		Timestamp: 0,                        // Set by world
	}

	cs.addSignal(signal)
}

// ReceiveSignals allows an entity to detect and respond to nearby signals
func (cs *CommunicationSystem) ReceiveSignals(entity *Entity) []Signal {
	var receivedSignals []Signal
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")

	// Only intelligent entities can receive signals
	if intelligence < 0.2 {
		return receivedSignals
	}

	for _, signal := range cs.Signals {
		distance := math.Sqrt(math.Pow(entity.Position.X-signal.Position.X, 2) +
			math.Pow(entity.Position.Y-signal.Position.Y, 2))

		// Check if entity is in range and can understand the signal
		if distance <= signal.Range && signal.Strength > 0.1 {
			// Cooperative entities are better at understanding signals
			comprehension := intelligence*0.7 + cooperation*0.3
			if rand.Float64() < comprehension {
				receivedSignals = append(receivedSignals, signal)
			}
		}
	}

	return receivedSignals
}

// Update fades signals over time and removes expired ones
func (cs *CommunicationSystem) Update() {
	activeSignals := make([]Signal, 0)

	for i := range cs.Signals {
		cs.Signals[i].Strength -= cs.Signals[i].Decay
		if cs.Signals[i].Strength > 0.05 {
			activeSignals = append(activeSignals, cs.Signals[i])
		}
	}

	cs.Signals = activeSignals
}

// addSignal adds a new signal to the system
func (cs *CommunicationSystem) addSignal(signal Signal) {
	if len(cs.Signals) >= cs.MaxSignals {
		// Remove oldest signal
		cs.Signals = cs.Signals[1:]
	}
	cs.Signals = append(cs.Signals, signal)
}

// Group represents a collaborative group of entities
type Group struct {
	ID       int
	Members  []*Entity
	Leader   *Entity
	Purpose  string  // "hunting", "migration", "territory", etc.
	Cohesion float64 // How well the group works together
}

// GroupBehaviorSystem manages entity groups and collaboration
type GroupBehaviorSystem struct {
	Groups []*Group
	NextID int
}

// NewGroupBehaviorSystem creates a new group behavior system
func NewGroupBehaviorSystem() *GroupBehaviorSystem {
	return &GroupBehaviorSystem{
		Groups: make([]*Group, 0),
		NextID: 1,
	}
}

// FormGroup creates a new group with compatible entities
func (gbs *GroupBehaviorSystem) FormGroup(entities []*Entity, purpose string) *Group {
	if len(entities) < 2 {
		return nil
	}

	// Check compatibility - entities should have similar traits and be cooperative
	avgCooperation := 0.0
	for _, entity := range entities {
		avgCooperation += entity.GetTrait("cooperation")
	}
	avgCooperation /= float64(len(entities))

	if avgCooperation < 0.4 {
		return nil // Not cooperative enough to form a group
	}

	// Find the most intelligent entity as leader
	var leader *Entity
	maxIntelligence := -1.0
	for _, entity := range entities {
		if entity.GetTrait("intelligence") > maxIntelligence {
			maxIntelligence = entity.GetTrait("intelligence")
			leader = entity
		}
	}

	group := &Group{
		ID:       gbs.NextID,
		Members:  entities,
		Leader:   leader,
		Purpose:  purpose,
		Cohesion: avgCooperation,
	}

	gbs.NextID++
	gbs.Groups = append(gbs.Groups, group)

	return group
}

// UpdateGroups maintains group integrity and behavior
func (gbs *GroupBehaviorSystem) UpdateGroups() {
	activeGroups := make([]*Group, 0)

	for _, group := range gbs.Groups {
		// Remove dead members
		aliveMembers := make([]*Entity, 0)
		for _, member := range group.Members {
			if member.IsAlive {
				aliveMembers = append(aliveMembers, member)
			}
		}

		// Disband if too few members or leader is dead
		if len(aliveMembers) < 2 || !group.Leader.IsAlive {
			continue
		}

		group.Members = aliveMembers

		// Group behavior based on purpose
		switch group.Purpose {
		case "hunting":
			gbs.coordinateHunting(group)
		case "migration":
			gbs.coordinateMigration(group)
		case "territory":
			gbs.defendTerritory(group)
		}

		activeGroups = append(activeGroups, group)
	}

	gbs.Groups = activeGroups
}

// coordinateHunting makes group members work together to hunt
func (gbs *GroupBehaviorSystem) coordinateHunting(group *Group) {
	// Group hunting increases success rate
	for _, member := range group.Members {
		// Boost aggression and coordination
		originalAggression := member.GetTrait("aggression")
		member.SetTrait("aggression", originalAggression*1.2)
	}
}

// coordinateMigration makes group move together toward better biomes
func (gbs *GroupBehaviorSystem) coordinateMigration(group *Group) {
	if group.Leader == nil {
		return
	}

	// Members follow leader toward better areas
	leaderPos := group.Leader.Position

	for _, member := range group.Members {
		if member == group.Leader {
			continue
		}

		// Move toward leader position
		speed := member.GetTrait("speed") * 0.5 // Slower when in group
		member.MoveTo(leaderPos.X, leaderPos.Y, speed)
	}
}

// defendTerritory makes group protect an area
func (gbs *GroupBehaviorSystem) defendTerritory(group *Group) {
	// Increase defensive traits when near territory center
	for _, member := range group.Members {
		originalDefense := member.GetTrait("defense")
		member.SetTrait("defense", originalDefense*1.3)
	}
}
