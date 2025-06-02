package main

import (
	"math"
	"math/rand"
)

// EnvironmentalModification represents a persistent change to the environment
type EnvironmentalModification struct {
	ID          int                       `json:"id"`
	Type        EnvironmentalModType      `json:"type"`
	Position    Position                  `json:"position"`
	Creator     *Entity                   `json:"-"`         // Entity that created it
	CreatedTick int                       `json:"created_tick"`
	LastUsedTick int                      `json:"last_used_tick"`
	Durability  float64                   `json:"durability"` // 0.0 to 1.0
	MaxDurability float64                 `json:"max_durability"`
	Depth       float64                   `json:"depth"`      // For tunnels/holes
	Width       float64                   `json:"width"`      // Size of modification
	IsActive    bool                      `json:"is_active"`
	Properties  map[string]float64        `json:"properties"` // Custom properties
	ConnectedTo []int                     `json:"connected_to"` // IDs of connected modifications
}

// EnvironmentalModType represents different types of environmental modifications
type EnvironmentalModType int

const (
	EnvModTunnel      EnvironmentalModType = iota // Underground passage
	EnvModBurrow                                  // Shelter hole
	EnvModCache                                   // Hidden storage
	EnvModTrap                                    // Environmental trap
	EnvModWaterhole                               // Dug water source
	EnvModPath                                    // Worn path
	EnvModMarking                                 // Scent/territorial marking
	EnvModNest                                    // Natural shelter
	EnvModBridge                                  // Crossing structure
	EnvModBarrier                                 // Environmental barrier
	EnvModTerrace                                 // Farming terrace
	EnvModDam                                     // Water control
)

// EnvironmentalModificationSystem manages environmental changes
type EnvironmentalModificationSystem struct {
	Modifications map[int]*EnvironmentalModification `json:"modifications"`
	NextModID     int                                 `json:"next_mod_id"`
	TunnelNetwork map[int][]int                       `json:"tunnel_network"` // Tunnel ID -> Connected tunnel IDs
}

// NewEnvironmentalModificationSystem creates a new environmental modification system
func NewEnvironmentalModificationSystem() *EnvironmentalModificationSystem {
	return &EnvironmentalModificationSystem{
		Modifications: make(map[int]*EnvironmentalModification),
		NextModID:     1,
		TunnelNetwork: make(map[int][]int),
	}
}

// CreateTunnel creates a tunnel system for entities to use
func (ems *EnvironmentalModificationSystem) CreateTunnel(creator *Entity, start Position, direction float64, length float64) *EnvironmentalModification {
	// Check if entity has sufficient skill for tunneling
	diggerSkill := creator.GetTrait("intelligence") + creator.GetTrait("strength")
	if diggerSkill < 0.4 {
		return nil
	}
	
	// Check energy requirements
	energyCost := length * 10.0
	if creator.Energy < energyCost {
		return nil
	}
	
	// Create the tunnel
	tunnel := &EnvironmentalModification{
		ID:            ems.NextModID,
		Type:          EnvModTunnel,
		Position:      start,
		Creator:       creator,
		CreatedTick:   0, // Will be set by caller
		LastUsedTick:  0,
		Durability:    0.8,
		MaxDurability: 0.8,
		Depth:         1.0 + diggerSkill*0.5,
		Width:         0.5 + creator.GetTrait("size")*0.3,
		IsActive:      true,
		Properties:    make(map[string]float64),
		ConnectedTo:   make([]int, 0),
	}
	
	// Set tunnel-specific properties
	tunnel.Properties["length"] = length
	tunnel.Properties["direction"] = direction
	tunnel.Properties["concealment"] = diggerSkill * 0.5
	tunnel.Properties["capacity"] = tunnel.Width * tunnel.Depth * 2.0
	
	// Consume energy
	creator.Energy -= energyCost
	
	// Add to system
	ems.Modifications[tunnel.ID] = tunnel
	ems.NextModID++
	
	return tunnel
}

// CreateBurrow creates a simple shelter burrow
func (ems *EnvironmentalModificationSystem) CreateBurrow(creator *Entity, position Position) *EnvironmentalModification {
	// Check skill requirements
	diggingSkill := creator.GetTrait("intelligence") + creator.GetTrait("strength")*0.5
	if diggingSkill < 0.2 {
		return nil
	}
	
	energyCost := 15.0
	if creator.Energy < energyCost {
		return nil
	}
	
	burrow := &EnvironmentalModification{
		ID:            ems.NextModID,
		Type:          EnvModBurrow,
		Position:      position,
		Creator:       creator,
		CreatedTick:   0,
		LastUsedTick:  0,
		Durability:    0.7,
		MaxDurability: 0.7,
		Depth:         0.8 + diggingSkill*0.3,
		Width:         1.0 + creator.GetTrait("size")*0.4,
		IsActive:      true,
		Properties:    make(map[string]float64),
		ConnectedTo:   make([]int, 0),
	}
	
	// Set burrow properties
	burrow.Properties["shelter_value"] = diggingSkill * 0.7
	burrow.Properties["concealment"] = diggingSkill * 0.6
	burrow.Properties["capacity"] = burrow.Width * burrow.Depth
	
	creator.Energy -= energyCost
	
	ems.Modifications[burrow.ID] = burrow
	ems.NextModID++
	
	return burrow
}

// CreateCache creates a hidden storage area
func (ems *EnvironmentalModificationSystem) CreateCache(creator *Entity, position Position) *EnvironmentalModification {
	intelligenceReq := 0.3
	if creator.GetTrait("intelligence") < intelligenceReq {
		return nil
	}
	
	energyCost := 12.0
	if creator.Energy < energyCost {
		return nil
	}
	
	cache := &EnvironmentalModification{
		ID:            ems.NextModID,
		Type:          EnvModCache,
		Position:      position,
		Creator:       creator,
		CreatedTick:   0,
		LastUsedTick:  0,
		Durability:    0.6,
		MaxDurability: 0.6,
		Depth:         0.5,
		Width:         0.8,
		IsActive:      true,
		Properties:    make(map[string]float64),
		ConnectedTo:   make([]int, 0),
	}
	
	// Set cache properties
	intelligence := creator.GetTrait("intelligence")
	cache.Properties["storage_capacity"] = 10.0 + intelligence*15.0
	cache.Properties["concealment"] = intelligence * 0.8
	cache.Properties["preservation"] = intelligence * 0.6
	cache.Properties["stored_resources"] = 0.0
	
	creator.Energy -= energyCost
	
	ems.Modifications[cache.ID] = cache
	ems.NextModID++
	
	return cache
}

// CreateTrap creates an environmental trap
func (ems *EnvironmentalModificationSystem) CreateTrap(creator *Entity, position Position, trapType string) *EnvironmentalModification {
	// Require intelligence and some aggression for trap making
	skillReq := creator.GetTrait("intelligence")*0.7 + creator.GetTrait("aggression")*0.3
	if skillReq < 0.4 {
		return nil
	}
	
	energyCost := 20.0
	if creator.Energy < energyCost {
		return nil
	}
	
	trap := &EnvironmentalModification{
		ID:            ems.NextModID,
		Type:          EnvModTrap,
		Position:      position,
		Creator:       creator,
		CreatedTick:   0,
		LastUsedTick:  0,
		Durability:    0.5,
		MaxDurability: 0.5,
		Depth:         0.3,
		Width:         1.2,
		IsActive:      true,
		Properties:    make(map[string]float64),
		ConnectedTo:   make([]int, 0),
	}
	
	// Set trap properties
	trap.Properties["effectiveness"] = skillReq * 0.8
	trap.Properties["trigger_sensitivity"] = creator.GetTrait("intelligence") * 0.7
	trap.Properties["concealment"] = skillReq * 0.9
	trap.Properties["damage_potential"] = creator.GetTrait("aggression") * 0.6
	
	creator.Energy -= energyCost
	
	ems.Modifications[trap.ID] = trap
	ems.NextModID++
	
	return trap
}

// CreatePath creates a worn path between locations
func (ems *EnvironmentalModificationSystem) CreatePath(creator *Entity, start Position, end Position) *EnvironmentalModification {
	// Paths form naturally through repeated use
	distance := math.Sqrt(math.Pow(end.X-start.X, 2) + math.Pow(end.Y-start.Y, 2))
	energyCost := distance * 2.0
	
	if creator.Energy < energyCost {
		return nil
	}
	
	path := &EnvironmentalModification{
		ID:            ems.NextModID,
		Type:          EnvModPath,
		Position:      start,
		Creator:       creator,
		CreatedTick:   0,
		LastUsedTick:  0,
		Durability:    0.3, // Paths start faint
		MaxDurability: 1.0,
		Depth:         0.1,
		Width:         0.5,
		IsActive:      true,
		Properties:    make(map[string]float64),
		ConnectedTo:   make([]int, 0),
	}
	
	// Set path properties
	path.Properties["length"] = distance
	path.Properties["end_x"] = end.X
	path.Properties["end_y"] = end.Y
	path.Properties["usage_count"] = 1.0
	path.Properties["speed_bonus"] = 0.1 // Small movement speed bonus
	
	creator.Energy -= energyCost
	
	ems.Modifications[path.ID] = path
	ems.NextModID++
	
	return path
}

// UseModification allows an entity to use an environmental modification
func (ems *EnvironmentalModificationSystem) UseModification(mod *EnvironmentalModification, user *Entity, tick int) float64 {
	if mod == nil || !mod.IsActive {
		return 0.0
	}
	
	// Calculate distance to modification
	distance := math.Sqrt(math.Pow(user.Position.X-mod.Position.X, 2) + 
						 math.Pow(user.Position.Y-mod.Position.Y, 2))
	
	// Must be close enough to use
	maxUseDistance := mod.Width + 1.0
	if distance > maxUseDistance {
		return 0.0
	}
	
	mod.LastUsedTick = tick
	benefit := 0.0
	
	switch mod.Type {
	case EnvModTunnel:
		benefit = ems.useTunnel(mod, user)
		
	case EnvModBurrow:
		benefit = ems.useBurrow(mod, user)
		
	case EnvModCache:
		benefit = ems.useCache(mod, user)
		
	case EnvModPath:
		benefit = ems.usePath(mod, user)
		
	case EnvModTrap:
		// Traps are usually triggered accidentally
		if user != mod.Creator && rand.Float64() < mod.Properties["trigger_sensitivity"] {
			benefit = ems.triggerTrap(mod, user)
		}
	}
	
	return benefit
}

// useTunnel provides benefits from using a tunnel
func (ems *EnvironmentalModificationSystem) useTunnel(tunnel *EnvironmentalModification, user *Entity) float64 {
	// Tunnels provide concealment and fast travel
	concealment := tunnel.Properties["concealment"]
	
	// Energy bonus from protection
	energyBonus := concealment * 2.0
	user.Energy = math.Min(100.0, user.Energy + energyBonus)
	
	// Movement speed bonus
	speedBonus := tunnel.Properties["length"] * 0.1
	
	return concealment + speedBonus
}

// useBurrow provides shelter benefits
func (ems *EnvironmentalModificationSystem) useBurrow(burrow *EnvironmentalModification, user *Entity) float64 {
	shelterValue := burrow.Properties["shelter_value"]
	
	// Energy restoration from shelter
	energyRestoration := shelterValue * 3.0
	user.Energy = math.Min(100.0, user.Energy + energyRestoration)
	
	// Protection from environmental hazards
	protection := burrow.Properties["concealment"]
	
	return shelterValue + protection
}

// useCache allows storing/retrieving resources
func (ems *EnvironmentalModificationSystem) useCache(cache *EnvironmentalModification, user *Entity) float64 {
	// Simple resource interaction - could be expanded
	storageCapacity := cache.Properties["storage_capacity"]
	currentStored := cache.Properties["stored_resources"]
	
	if user.Energy > 80.0 && currentStored < storageCapacity {
		// Store some energy as resources
		storeAmount := math.Min(5.0, storageCapacity - currentStored)
		cache.Properties["stored_resources"] = currentStored + storeAmount
		user.Energy -= storeAmount
		return storeAmount * 0.5
	} else if user.Energy < 40.0 && currentStored > 0.0 {
		// Retrieve resources as energy
		retrieveAmount := math.Min(10.0, currentStored)
		cache.Properties["stored_resources"] = currentStored - retrieveAmount
		user.Energy = math.Min(100.0, user.Energy + retrieveAmount)
		return retrieveAmount
	}
	
	return 0.0
}

// usePath provides movement benefits
func (ems *EnvironmentalModificationSystem) usePath(path *EnvironmentalModification, user *Entity) float64 {
	// Increment usage count to strengthen the path
	path.Properties["usage_count"] += 1.0
	
	// Strengthen the path with use
	strengthIncrease := 0.01
	path.Durability = math.Min(path.MaxDurability, path.Durability + strengthIncrease)
	
	// Return speed bonus
	return path.Properties["speed_bonus"] * path.Durability
}

// triggerTrap handles trap activation
func (ems *EnvironmentalModificationSystem) triggerTrap(trap *EnvironmentalModification, victim *Entity) float64 {
	if !trap.IsActive {
		return 0.0
	}
	
	damage := trap.Properties["damage_potential"] * 10.0
	
	// Apply damage to victim
	victim.Energy = math.Max(0.0, victim.Energy - damage)
	
	// Reduce trap durability after use
	trap.Durability -= 0.3
	if trap.Durability <= 0.0 {
		trap.IsActive = false
	}
	
	return -damage // Negative benefit for the victim
}

// ConnectTunnels connects two tunnels to form a network
func (ems *EnvironmentalModificationSystem) ConnectTunnels(tunnel1ID, tunnel2ID int) bool {
	tunnel1, exists1 := ems.Modifications[tunnel1ID]
	tunnel2, exists2 := ems.Modifications[tunnel2ID]
	
	if !exists1 || !exists2 || tunnel1.Type != EnvModTunnel || tunnel2.Type != EnvModTunnel {
		return false
	}
	
	// Check if tunnels are close enough to connect
	distance := math.Sqrt(math.Pow(tunnel1.Position.X-tunnel2.Position.X, 2) + 
						 math.Pow(tunnel1.Position.Y-tunnel2.Position.Y, 2))
	
	maxConnectionDistance := tunnel1.Properties["length"] + tunnel2.Properties["length"]
	if distance > maxConnectionDistance {
		return false
	}
	
	// Add connections
	tunnel1.ConnectedTo = append(tunnel1.ConnectedTo, tunnel2ID)
	tunnel2.ConnectedTo = append(tunnel2.ConnectedTo, tunnel1ID)
	
	// Update tunnel network
	if ems.TunnelNetwork[tunnel1ID] == nil {
		ems.TunnelNetwork[tunnel1ID] = make([]int, 0)
	}
	if ems.TunnelNetwork[tunnel2ID] == nil {
		ems.TunnelNetwork[tunnel2ID] = make([]int, 0)
	}
	
	ems.TunnelNetwork[tunnel1ID] = append(ems.TunnelNetwork[tunnel1ID], tunnel2ID)
	ems.TunnelNetwork[tunnel2ID] = append(ems.TunnelNetwork[tunnel2ID], tunnel1ID)
	
	return true
}

// GetNearbyModifications returns environmental modifications near a position
func (ems *EnvironmentalModificationSystem) GetNearbyModifications(position Position, radius float64) []*EnvironmentalModification {
	nearby := make([]*EnvironmentalModification, 0)
	
	for _, mod := range ems.Modifications {
		if !mod.IsActive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(position.X-mod.Position.X, 2) + 
							 math.Pow(position.Y-mod.Position.Y, 2))
		
		if distance <= radius {
			nearby = append(nearby, mod)
		}
	}
	
	return nearby
}

// UpdateModifications maintains environmental modifications
func (ems *EnvironmentalModificationSystem) UpdateModifications(tick int) {
	for id, mod := range ems.Modifications {
		if !mod.IsActive {
			continue
		}
		
		// Natural decay
		decayRate := 0.0001
		if tick - mod.LastUsedTick > 500 { // Faster decay if unused
			decayRate *= 2.0
		}
		
		mod.Durability -= decayRate
		
		// Remove completely decayed modifications
		if mod.Durability <= 0.0 {
			mod.IsActive = false
			
			// Remove from tunnel network if it's a tunnel
			if mod.Type == EnvModTunnel {
				delete(ems.TunnelNetwork, id)
				// Remove connections to this tunnel
				for _, connectedID := range mod.ConnectedTo {
					if connectedMod, exists := ems.Modifications[connectedID]; exists {
						// Remove this tunnel from connected tunnel's connections
						newConnections := make([]int, 0)
						for _, cid := range connectedMod.ConnectedTo {
							if cid != id {
								newConnections = append(newConnections, cid)
							}
						}
						connectedMod.ConnectedTo = newConnections
					}
				}
			}
		}
	}
}

// GetModificationStats returns statistics about environmental modifications
func (ems *EnvironmentalModificationSystem) GetModificationStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalMods := 0
	activeMods := 0
	typeCounts := make(map[EnvironmentalModType]int)
	avgDurability := 0.0
	
	for _, mod := range ems.Modifications {
		totalMods++
		if mod.IsActive {
			activeMods++
			avgDurability += mod.Durability
		}
		typeCounts[mod.Type]++
	}
	
	if activeMods > 0 {
		avgDurability /= float64(activeMods)
	}
	
	stats["total_modifications"] = totalMods
	stats["active_modifications"] = activeMods
	stats["inactive_modifications"] = totalMods - activeMods
	stats["avg_durability"] = avgDurability
	stats["tunnel_networks"] = len(ems.TunnelNetwork)
	stats["modification_types"] = typeCounts
	
	return stats
}

// GetEnvironmentalModTypeName returns the name of an environmental modification type
func GetEnvironmentalModTypeName(modType EnvironmentalModType) string {
	names := map[EnvironmentalModType]string{
		EnvModTunnel:    "Tunnel",
		EnvModBurrow:    "Burrow",
		EnvModCache:     "Cache",
		EnvModTrap:      "Trap",
		EnvModWaterhole: "Waterhole",
		EnvModPath:      "Path",
		EnvModMarking:   "Marking",
		EnvModNest:      "Nest",
		EnvModBridge:    "Bridge",
		EnvModBarrier:   "Barrier",
		EnvModTerrace:   "Terrace",
		EnvModDam:       "Dam",
	}
	
	if name, exists := names[modType]; exists {
		return name
	}
	return "Unknown Modification"
}