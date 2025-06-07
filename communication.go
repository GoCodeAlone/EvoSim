package main

import (
	"fmt"
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
	EventBus   *CentralEventBus // For event tracking
}

// NewCommunicationSystem creates a new communication system
func NewCommunicationSystem(eventBus *CentralEventBus) *CommunicationSystem {
	return &CommunicationSystem{
		Signals:    make([]Signal, 0),
		MaxSignals: 100, // Limit active signals
		EventBus:   eventBus,
	}
}

// SendSignal allows an entity to broadcast a signal
func (cs *CommunicationSystem) SendSignal(entity *Entity, signalType SignalType, data map[string]interface{}, tick int) {
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
		Timestamp: tick,
	}

	cs.addSignal(signal)

	// Emit event for signal creation
	if cs.EventBus != nil {
		signalTypeNames := []string{"danger", "food", "mating", "territory", "help", "migration"}
		signalTypeName := "unknown"
		if int(signalType) < len(signalTypeNames) {
			signalTypeName = signalTypeNames[signalType]
		}

		metadata := map[string]interface{}{
			"signal_type":     signalTypeName,
			"signal_strength": signal.Strength,
			"signal_range":    signal.Range,
			"intelligence":    intelligence,
			"cooperation":     cooperation,
		}
		if len(data) > 0 {
			metadata["signal_data"] = data
		}

		cs.EventBus.EmitEntityEvent(tick, "signal_sent", "communication", "communication_system",
			fmt.Sprintf("Entity %d sent %s signal with strength %.2f", entity.ID, signalTypeName, signal.Strength),
			entity, nil, signalTypeName, nil)
	}
}

// ReceiveSignals allows an entity to detect and respond to nearby signals
func (cs *CommunicationSystem) ReceiveSignals(entity *Entity, tick int) []Signal {
	var receivedSignals []Signal
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")

	// Only intelligent entities can receive signals
	if intelligence < 0.2 {
		return receivedSignals
	}

	for _, signal := range cs.Signals {
		dx := entity.Position.X - signal.Position.X
		dy := entity.Position.Y - signal.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Check if entity is in range and can understand the signal
		if distance <= signal.Range && signal.Strength > 0.1 {
			// Cooperative entities are better at understanding signals
			comprehension := intelligence*0.7 + cooperation*0.3
			if rand.Float64() < comprehension {
				receivedSignals = append(receivedSignals, signal)

				// Emit event for signal reception
				if cs.EventBus != nil {
					signalTypeNames := []string{"danger", "food", "mating", "territory", "help", "migration"}
					signalTypeName := "unknown"
					if int(signal.Type) < len(signalTypeNames) {
						signalTypeName = signalTypeNames[signal.Type]
					}

					cs.EventBus.EmitEntityEvent(tick, "signal_received", "communication", "communication_system",
						fmt.Sprintf("Entity %d received %s signal from distance %.2f", entity.ID, signalTypeName, distance),
						entity, nil, signalTypeName, nil)
				}
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
	Groups   []*Group
	NextID   int
	EventBus *CentralEventBus // For event tracking
}

// NewGroupBehaviorSystem creates a new group behavior system
func NewGroupBehaviorSystem(eventBus *CentralEventBus) *GroupBehaviorSystem {
	return &GroupBehaviorSystem{
		Groups:   make([]*Group, 0),
		NextID:   1,
		EventBus: eventBus,
	}
}

// FormGroup creates a new group with compatible entities
func (gbs *GroupBehaviorSystem) FormGroup(entities []*Entity, purpose string, tick int) *Group {
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
		if entity != nil && entity.IsAlive && entity.GetTrait("intelligence") > maxIntelligence {
			maxIntelligence = entity.GetTrait("intelligence")
			leader = entity
		}
	}

	// Ensure we found a valid leader
	if leader == nil {
		return nil // Cannot form group without a leader
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

	// Emit event for group formation
	if gbs.EventBus != nil {
		memberIDs := make([]int, len(entities))
		for i, entity := range entities {
			memberIDs[i] = entity.ID
		}

		metadata := map[string]interface{}{
			"group_id":        group.ID,
			"purpose":         purpose,
			"member_count":    len(entities),
			"member_ids":      memberIDs,
			"leader_id":       leader.ID,
			"avg_cooperation": avgCooperation,
			"cohesion":        group.Cohesion,
		}

		gbs.EventBus.EmitSystemEvent(tick, "group_formed", "communication", "group_behavior_system",
			fmt.Sprintf("Group %d formed with %d members for %s (leader: %d)", group.ID, len(entities), purpose, leader.ID),
			&leader.Position, metadata)
	}

	return group
}

// UpdateGroups maintains group integrity and behavior
func (gbs *GroupBehaviorSystem) UpdateGroups(tick int) {
	activeGroups := make([]*Group, 0)

	for _, group := range gbs.Groups {
		originalMemberCount := len(group.Members)

		// Remove dead members
		aliveMembers := make([]*Entity, 0)
		for _, member := range group.Members {
			if member != nil && member.IsAlive {
				aliveMembers = append(aliveMembers, member)
			}
		}

		// If leader is dead or nil, try to elect a new one from alive members
		leaderChanged := false
		oldLeaderID := 0
		if group.Leader != nil {
			oldLeaderID = group.Leader.ID
		}

		if group.Leader == nil || !group.Leader.IsAlive {
			if len(aliveMembers) > 0 {
				// Find new leader with highest intelligence
				var newLeader *Entity
				maxIntelligence := -1.0
				for _, member := range aliveMembers {
					if member.GetTrait("intelligence") > maxIntelligence {
						maxIntelligence = member.GetTrait("intelligence")
						newLeader = member
					}
				}
				if newLeader != nil && (group.Leader == nil || newLeader.ID != group.Leader.ID) {
					group.Leader = newLeader
					leaderChanged = true
				}
			}
		}

		// Emit event for leader change
		if leaderChanged && gbs.EventBus != nil && group.Leader != nil {
			metadata := map[string]interface{}{
				"group_id":      group.ID,
				"old_leader_id": oldLeaderID,
				"new_leader_id": group.Leader.ID,
				"purpose":       group.Purpose,
			}

			gbs.EventBus.EmitSystemEvent(tick, "group_leader_changed", "communication", "group_behavior_system",
				fmt.Sprintf("Group %d leader changed from %d to %d", group.ID, oldLeaderID, group.Leader.ID),
				&group.Leader.Position, metadata)
		}

		// Disband if too few members or leader is dead/nil
		if len(aliveMembers) < 2 || group.Leader == nil || !group.Leader.IsAlive {
			// Emit event for group disbanding
			if gbs.EventBus != nil {
				metadata := map[string]interface{}{
					"group_id":          group.ID,
					"purpose":           group.Purpose,
					"original_members":  originalMemberCount,
					"remaining_members": len(aliveMembers),
					"reason":            "insufficient_members_or_no_leader",
				}

				position := &Position{X: 0, Y: 0}
				if len(aliveMembers) > 0 {
					position = &aliveMembers[0].Position
				}

				gbs.EventBus.EmitSystemEvent(tick, "group_disbanded", "communication", "group_behavior_system",
					fmt.Sprintf("Group %d disbanded (%s, %d->%d members)", group.ID, group.Purpose, originalMemberCount, len(aliveMembers)),
					position, metadata)
			}
			continue
		}

		group.Members = aliveMembers

		// Group behavior based on purpose
		switch group.Purpose {
		case "hunting":
			gbs.coordinateHunting(group, tick)
		case "migration":
			gbs.coordinateMigration(group, tick)
		case "territory":
			gbs.defendTerritory(group, tick)
		}

		activeGroups = append(activeGroups, group)
	}

	gbs.Groups = activeGroups
}

// coordinateHunting makes group members work together to hunt
func (gbs *GroupBehaviorSystem) coordinateHunting(group *Group, tick int) {
	if group.Leader == nil || len(group.Members) == 0 {
		return
	}

	// Group hunting increases success rate
	for _, member := range group.Members {
		if member == nil || !member.IsAlive {
			continue
		}
		// Boost aggression and coordination
		originalAggression := member.GetTrait("aggression")
		member.SetTrait("aggression", originalAggression*1.2)
	}

	// Emit event for coordinated hunting
	if gbs.EventBus != nil {
		metadata := map[string]interface{}{
			"group_id":     group.ID,
			"member_count": len(group.Members),
			"leader_id":    group.Leader.ID,
			"cohesion":     group.Cohesion,
		}

		gbs.EventBus.EmitSystemEvent(tick, "group_hunting", "communication", "group_behavior_system",
			fmt.Sprintf("Group %d coordinating hunt with %d members", group.ID, len(group.Members)),
			&group.Leader.Position, metadata)
	}
}

// coordinateMigration makes group move together toward better biomes
func (gbs *GroupBehaviorSystem) coordinateMigration(group *Group, tick int) {
	if group.Leader == nil || !group.Leader.IsAlive {
		return
	}

	// Members follow leader toward better areas
	leaderPos := group.Leader.Position

	membersFollowing := 0
	for _, member := range group.Members {
		if member == nil || !member.IsAlive || member == group.Leader {
			continue
		}

		// Move toward leader position
		speed := member.GetTrait("speed") * 0.5 // Slower when in group
		member.MoveTo(leaderPos.X, leaderPos.Y, speed)
		membersFollowing++
	}

	// Emit event for coordinated migration
	if gbs.EventBus != nil {
		metadata := map[string]interface{}{
			"group_id":          group.ID,
			"leader_id":         group.Leader.ID,
			"members_following": membersFollowing,
			"destination_x":     leaderPos.X,
			"destination_y":     leaderPos.Y,
		}

		gbs.EventBus.EmitSystemEvent(tick, "group_migration", "communication", "group_behavior_system",
			fmt.Sprintf("Group %d migrating with %d members following leader %d", group.ID, membersFollowing, group.Leader.ID),
			&group.Leader.Position, metadata)
	}
}

// defendTerritory makes group protect an area
func (gbs *GroupBehaviorSystem) defendTerritory(group *Group, tick int) {
	if group.Leader == nil || len(group.Members) == 0 {
		return
	}

	// Increase defensive traits when near territory center
	defendersCount := 0
	for _, member := range group.Members {
		if member == nil || !member.IsAlive {
			continue
		}
		originalDefense := member.GetTrait("defense")
		member.SetTrait("defense", originalDefense*1.3)
		defendersCount++
	}

	// Emit event for territory defense
	if gbs.EventBus != nil {
		metadata := map[string]interface{}{
			"group_id":        group.ID,
			"leader_id":       group.Leader.ID,
			"defenders_count": defendersCount,
			"cohesion":        group.Cohesion,
		}

		gbs.EventBus.EmitSystemEvent(tick, "group_territory_defense", "communication", "group_behavior_system",
			fmt.Sprintf("Group %d defending territory with %d members", group.ID, defendersCount),
			&group.Leader.Position, metadata)
	}
}
