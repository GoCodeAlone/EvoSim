package main

import (
	"fmt"
	"math"
	"math/rand"
)

// EmergentBehaviorSystem manages emergent behaviors that can develop naturally
type EmergentBehaviorSystem struct {
	BehaviorPatterns map[int]*BehaviorPattern `json:"behavior_patterns"` // Entity ID -> Behavior Pattern
	LearnedBehaviors map[string]*LearnedBehavior `json:"learned_behaviors"` // Behavior name -> Learned Behavior
	NextBehaviorID   int                         `json:"next_behavior_id"`
}

// BehaviorPattern represents an entity's learned behavior pattern
type BehaviorPattern struct {
	EntityID          int                        `json:"entity_id"`
	KnownBehaviors    map[string]float64         `json:"known_behaviors"`     // Behavior name -> proficiency
	PreferredActions  []string                   `json:"preferred_actions"`
	LearningRate      float64                    `json:"learning_rate"`
	Curiosity         float64                    `json:"curiosity"`
	ExplorationTendency float64                  `json:"exploration_tendency"`
	SocialLearning    bool                       `json:"social_learning"`     // Can learn from others
	ToolPreferences   map[ToolType]float64       `json:"tool_preferences"`    // Tool type -> preference score
}

// LearnedBehavior represents a behavior that has been discovered and can spread
type LearnedBehavior struct {
	Name            string    `json:"name"`
	Discoverer      *Entity   `json:"-"`
	DiscoveredTick  int       `json:"discovered_tick"`
	Complexity      float64   `json:"complexity"`        // How hard it is to learn
	Effectiveness   float64   `json:"effectiveness"`     // How beneficial it is
	Prerequisites   []string  `json:"prerequisites"`     // Required skills/behaviors
	Spread          int       `json:"spread"`            // How many entities know it
	Description     string    `json:"description"`
}

// NewEmergentBehaviorSystem creates a new emergent behavior system
func NewEmergentBehaviorSystem() *EmergentBehaviorSystem {
	ebs := &EmergentBehaviorSystem{
		BehaviorPatterns: make(map[int]*BehaviorPattern),
		LearnedBehaviors: make(map[string]*LearnedBehavior),
		NextBehaviorID:   1,
	}
	
	// Initialize some basic discoverable behaviors
	ebs.initializeBasicBehaviors()
	
	return ebs
}

// initializeBasicBehaviors sets up behaviors that can be discovered
func (ebs *EmergentBehaviorSystem) initializeBasicBehaviors() {
	behaviors := []LearnedBehavior{
		{
			Name:          "tool_making",
			Complexity:    0.4,
			Effectiveness: 0.7,
			Prerequisites: []string{},
			Description:   "Creating basic tools from available materials",
		},
		{
			Name:          "tunnel_digging",
			Complexity:    0.6,
			Effectiveness: 0.8,
			Prerequisites: []string{"tool_making"},
			Description:   "Creating underground passages for protection and travel",
		},
		{
			Name:          "cache_hiding",
			Complexity:    0.3,
			Effectiveness: 0.6,
			Prerequisites: []string{},
			Description:   "Storing resources in hidden locations for later use",
		},
		{
			Name:          "trap_setting",
			Complexity:    0.7,
			Effectiveness: 0.9,
			Prerequisites: []string{"tool_making"},
			Description:   "Creating environmental traps to catch prey",
		},
		{
			Name:          "path_making",
			Complexity:    0.2,
			Effectiveness: 0.4,
			Prerequisites: []string{},
			Description:   "Creating efficient travel routes between locations",
		},
		{
			Name:          "cooperative_building",
			Complexity:    0.8,
			Effectiveness: 1.0,
			Prerequisites: []string{"tool_making"},
			Description:   "Working together to create complex structures",
		},
		{
			Name:          "resource_sharing",
			Complexity:    0.5,
			Effectiveness: 0.6,
			Prerequisites: []string{"cache_hiding"},
			Description:   "Sharing stored resources with tribe members",
		},
		{
			Name:          "tool_modification",
			Complexity:    0.6,
			Effectiveness: 0.8,
			Prerequisites: []string{"tool_making"},
			Description:   "Improving existing tools for better performance",
		},
	}
	
	for _, behavior := range behaviors {
		ebs.LearnedBehaviors[behavior.Name] = &behavior
	}
}

// InitializeEntityBehavior creates a behavior pattern for a new entity
func (ebs *EmergentBehaviorSystem) InitializeEntityBehavior(entity *Entity) {
	// Increased learning rates and made social learning more accessible
	pattern := &BehaviorPattern{
		EntityID:            entity.ID,
		KnownBehaviors:      make(map[string]float64),
		PreferredActions:    make([]string, 0),
		LearningRate:        0.2 + entity.GetTrait("intelligence") * 0.4, // Base rate + intelligence bonus
		Curiosity:           entity.GetTrait("curiosity"),
		ExplorationTendency: entity.GetTrait("curiosity") * 0.7, // Increased exploration
		SocialLearning:      entity.GetTrait("cooperation") > 0.2, // Lowered threshold
		ToolPreferences:     make(map[ToolType]float64),
	}
	
	// Initialize tool preferences based on traits with better scaling
	pattern.ToolPreferences[ToolStone] = 0.3 + entity.GetTrait("strength") * 0.5
	pattern.ToolPreferences[ToolSpear] = 0.2 + entity.GetTrait("aggression") * 0.6
	pattern.ToolPreferences[ToolDigger] = 0.2 + entity.GetTrait("intelligence") * 0.5
	pattern.ToolPreferences[ToolContainer] = 0.3 + entity.GetTrait("cooperation") * 0.4
	pattern.ToolPreferences[ToolHammer] = 0.2 + entity.GetTrait("strength") * 0.4
	pattern.ToolPreferences[ToolBlade] = 0.1 + entity.GetTrait("aggression") * 0.5
	
	ebs.BehaviorPatterns[entity.ID] = pattern
}

// UpdateEntityBehaviors processes behavioral learning and decision making
func (ebs *EmergentBehaviorSystem) UpdateEntityBehaviors(world *World) {
	for _, entity := range world.AllEntities {
		if !entity.IsAlive {
			continue
		}
		
		pattern, exists := ebs.BehaviorPatterns[entity.ID]
		if !exists {
			ebs.InitializeEntityBehavior(entity)
			pattern = ebs.BehaviorPatterns[entity.ID]
		}
		
		// Try to discover new behaviors - increased rate and made intelligence-dependent
		discoveryRate := pattern.Curiosity * entity.GetTrait("intelligence") * 0.02
		if rand.Float64() < discoveryRate {
			ebs.attemptBehaviorDiscovery(entity, pattern, world)
		}
		
		// Try to learn from nearby entities - made trait-dependent
		socialRate := pattern.Curiosity * entity.GetTrait("cooperation") * 0.08
		if pattern.SocialLearning && rand.Float64() < socialRate {
			ebs.attemptSocialLearning(entity, pattern, world)
		}
		
		// Execute learned behaviors
		ebs.executeEntityBehaviors(entity, pattern, world)
	}
}

// attemptBehaviorDiscovery tries to discover new behaviors based on circumstances
func (ebs *EmergentBehaviorSystem) attemptBehaviorDiscovery(entity *Entity, pattern *BehaviorPattern, world *World) {
	for behaviorName, behavior := range ebs.LearnedBehaviors {
		// Skip if already known
		if pattern.KnownBehaviors[behaviorName] > 0.0 {
			continue
		}
		
		// Check prerequisites
		hasPrerequisites := true
		for _, prereq := range behavior.Prerequisites {
			if pattern.KnownBehaviors[prereq] < 0.3 {
				hasPrerequisites = false
				break
			}
		}
		
		if !hasPrerequisites {
			continue
		}
		
		// Check if entity has sufficient intelligence to discover this behavior
		intelligence := entity.GetTrait("intelligence")
		curiosityBonus := pattern.Curiosity * 0.5
		discoveryChance := (intelligence - behavior.Complexity + curiosityBonus) * 0.15
		
		if discoveryChance > 0 && rand.Float64() < discoveryChance {
			// Successfully discovered behavior!
			pattern.KnownBehaviors[behaviorName] = 0.1 // Start with basic proficiency
			behavior.Spread++
			
			// Log the discovery
			world.EventLogger.LogWorldEvent(world.Tick, "Behavior Discovery", 
				fmt.Sprintf("Entity %d discovered %s", entity.ID, behaviorName))
		}
	}
}

// attemptSocialLearning tries to learn behaviors from nearby entities
func (ebs *EmergentBehaviorSystem) attemptSocialLearning(entity *Entity, pattern *BehaviorPattern, world *World) {
	// Find nearby entities
	for _, other := range world.AllEntities {
		if other.ID == entity.ID || !other.IsAlive {
			continue
		}
		
		// Check if nearby
		distance := math.Sqrt(math.Pow(entity.Position.X-other.Position.X, 2) + 
							 math.Pow(entity.Position.Y-other.Position.Y, 2))
		
		if distance > 5.0 {
			continue
		}
		
		// Get other entity's behavior pattern
		otherPattern, exists := ebs.BehaviorPatterns[other.ID]
		if !exists {
			continue
		}
		
		// Try to learn behaviors the other entity knows
		for behaviorName, otherProficiency := range otherPattern.KnownBehaviors {
			if otherProficiency <= 0.05 { // Lowered threshold for learning
				continue
			}
			
			myProficiency := pattern.KnownBehaviors[behaviorName]
			if myProficiency >= otherProficiency { // Can't learn from less skilled
				continue
			}
			
			// Learning chance based on cooperation and intelligence - increased rates
			baseChance := entity.GetTrait("cooperation") * entity.GetTrait("intelligence") * 0.08
			proximityBonus := math.Max(0, (5.0 - distance) / 5.0) * 0.03
			learningChance := baseChance + proximityBonus
			
			if rand.Float64() < learningChance {
				// Learn a bit of the behavior - increased learning rate
				improvement := pattern.LearningRate * 0.15
				pattern.KnownBehaviors[behaviorName] = math.Min(otherProficiency, myProficiency + improvement)
			}
		}
	}
}

// executeEntityBehaviors makes entities perform their learned behaviors
func (ebs *EmergentBehaviorSystem) executeEntityBehaviors(entity *Entity, pattern *BehaviorPattern, world *World) {
	// Prioritize behaviors based on current needs and proficiency
	bestBehavior := ""
	bestScore := 0.0
	
	for behaviorName, proficiency := range pattern.KnownBehaviors {
		if proficiency < 0.1 {
			continue
		}
		
		// Calculate behavior score based on current situation
		score := ebs.calculateBehaviorScore(entity, behaviorName, proficiency, world)
		
		if score > bestScore {
			bestScore = score
			bestBehavior = behaviorName
		}
	}
	
	// Lowered threshold to make behaviors more likely to be executed early
	if bestBehavior != "" && bestScore > 0.2 {
		ebs.executeBehavior(entity, bestBehavior, pattern, world)
	}
}

// calculateBehaviorScore determines how valuable a behavior is in the current situation
func (ebs *EmergentBehaviorSystem) calculateBehaviorScore(entity *Entity, behaviorName string, proficiency float64, world *World) float64 {
	score := proficiency * 0.5 // Base score from proficiency
	
	switch behaviorName {
	case "tool_making":
		// More valuable if entity has no tools or poor tools
		tools := world.ToolSystem.GetEntityTools(entity)
		if len(tools) == 0 {
			score += 0.8
		} else {
			avgEfficiency := 0.0
			for _, tool := range tools {
				avgEfficiency += tool.GetToolEffectiveness()
			}
			avgEfficiency /= float64(len(tools))
			score += (1.0 - avgEfficiency) * 0.6
		}
		
	case "tunnel_digging":
		// More valuable in dangerous areas or for entities with high aggression nearby
		nearbyDanger := ebs.assessNearbyDanger(entity, world)
		score += nearbyDanger * 0.7
		
	case "cache_hiding":
		// More valuable when entity has excess energy
		if entity.Energy > 80.0 {
			score += 0.6
		}
		
	case "trap_setting":
		// More valuable for aggressive entities or when food is scarce
		aggression := entity.GetTrait("aggression")
		score += aggression * 0.5
		
		// Check if food is scarce
		nearbyPlants := ebs.countNearbyResources(entity, world)
		if nearbyPlants < 3 {
			score += 0.4
		}
		
	case "path_making":
		// More valuable if entity travels frequently
		score += entity.GetTrait("speed") * 0.3
		
	case "cooperative_building":
		// More valuable if entity is in a tribe with others
		cooperation := entity.GetTrait("cooperation")
		nearbyAllies := ebs.countNearbyAllies(entity, world)
		score += cooperation * float64(nearbyAllies) * 0.2
		
	case "resource_sharing":
		// More valuable if entity has high cooperation and others nearby are low on energy
		cooperation := entity.GetTrait("cooperation")
		needyAllies := ebs.countNeedyAllies(entity, world)
		score += cooperation * float64(needyAllies) * 0.3
		
	case "tool_modification":
		// More valuable if entity has tools that could be improved
		tools := world.ToolSystem.GetEntityTools(entity)
		for _, tool := range tools {
			if tool.GetToolEffectiveness() < 0.8 {
				score += 0.4
				break
			}
		}
	}
	
	return math.Max(0.0, score)
}

// executeBehavior performs the chosen behavior
func (ebs *EmergentBehaviorSystem) executeBehavior(entity *Entity, behaviorName string, pattern *BehaviorPattern, world *World) {
	proficiency := pattern.KnownBehaviors[behaviorName]
	
	switch behaviorName {
	case "tool_making":
		// Try to create a tool based on preferences and situation
		bestToolType := ebs.chooseBestToolType(entity, pattern, world)
		if bestToolType != -1 {
			tool := world.ToolSystem.CreateTool(entity, ToolType(bestToolType), entity.Position)
			if tool != nil {
				// Improve proficiency through practice
				pattern.KnownBehaviors[behaviorName] = math.Min(1.0, proficiency + 0.05)
			}
		}
		
	case "tunnel_digging":
		// Create a tunnel for protection
		direction := rand.Float64() * 2 * math.Pi
		length := 3.0 + proficiency*5.0
		tunnel := world.EnvironmentalModSystem.CreateTunnel(entity, entity.Position, direction, length)
		if tunnel != nil {
			pattern.KnownBehaviors[behaviorName] = math.Min(1.0, proficiency + 0.03)
		}
		
	case "cache_hiding":
		// Create a cache to store resources
		cache := world.EnvironmentalModSystem.CreateCache(entity, entity.Position)
		if cache != nil {
			pattern.KnownBehaviors[behaviorName] = math.Min(1.0, proficiency + 0.04)
		}
		
	case "trap_setting":
		// Set a trap near a resource area
		trapPos := ebs.findGoodTrapLocation(entity, world)
		trap := world.EnvironmentalModSystem.CreateTrap(entity, trapPos, "basic")
		if trap != nil {
			pattern.KnownBehaviors[behaviorName] = math.Min(1.0, proficiency + 0.02)
		}
		
	case "tool_modification":
		// Try to improve an existing tool
		tools := world.ToolSystem.GetEntityTools(entity)
		if len(tools) > 0 {
			tool := tools[rand.Intn(len(tools))]
			modificationType := ModificationType(rand.Intn(6)) // Random modification type
			success := world.ToolSystem.ModifyTool(tool, entity, modificationType)
			if success {
				pattern.KnownBehaviors[behaviorName] = math.Min(1.0, proficiency + 0.03)
			}
		}
	}
}

// Helper functions for behavior scoring

func (ebs *EmergentBehaviorSystem) assessNearbyDanger(entity *Entity, world *World) float64 {
	danger := 0.0
	for _, other := range world.AllEntities {
		if other.ID == entity.ID || !other.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(entity.Position.X-other.Position.X, 2) + 
							 math.Pow(entity.Position.Y-other.Position.Y, 2))
		
		if distance < 10.0 {
			otherAggression := other.GetTrait("aggression")
			danger += otherAggression / (distance + 1.0)
		}
	}
	return math.Min(1.0, danger)
}

func (ebs *EmergentBehaviorSystem) countNearbyResources(entity *Entity, world *World) int {
	count := 0
	for _, plant := range world.AllPlants {
		if !plant.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(entity.Position.X-plant.Position.X, 2) + 
							 math.Pow(entity.Position.Y-plant.Position.Y, 2))
		
		if distance < 8.0 {
			count++
		}
	}
	return count
}

func (ebs *EmergentBehaviorSystem) countNearbyAllies(entity *Entity, world *World) int {
	count := 0
	myCooperation := entity.GetTrait("cooperation")
	
	for _, other := range world.AllEntities {
		if other.ID == entity.ID || !other.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(entity.Position.X-other.Position.X, 2) + 
							 math.Pow(entity.Position.Y-other.Position.Y, 2))
		
		if distance < 6.0 {
			otherCooperation := other.GetTrait("cooperation")
			if myCooperation > 0.5 && otherCooperation > 0.5 {
				count++
			}
		}
	}
	return count
}

func (ebs *EmergentBehaviorSystem) countNeedyAllies(entity *Entity, world *World) int {
	count := 0
	myCooperation := entity.GetTrait("cooperation")
	
	for _, other := range world.AllEntities {
		if other.ID == entity.ID || !other.IsAlive {
			continue
		}
		
		distance := math.Sqrt(math.Pow(entity.Position.X-other.Position.X, 2) + 
							 math.Pow(entity.Position.Y-other.Position.Y, 2))
		
		if distance < 6.0 && other.Energy < 40.0 {
			otherCooperation := other.GetTrait("cooperation")
			if myCooperation > 0.5 && otherCooperation > 0.5 {
				count++
			}
		}
	}
	return count
}

func (ebs *EmergentBehaviorSystem) chooseBestToolType(entity *Entity, pattern *BehaviorPattern, world *World) int {
	bestScore := 0.0
	bestType := -1
	
	for toolType, preference := range pattern.ToolPreferences {
		// Check if entity can create this tool type
		recipe, exists := world.ToolSystem.ToolRecipes[toolType]
		if !exists {
			continue
		}
		
		if entity.GetTrait("intelligence") < recipe.RequiredSkill || entity.Energy < recipe.RequiredEnergy {
			continue
		}
		
		score := preference
		
		// Bonus for tools entity doesn't have
		hasThisType := false
		for _, tool := range world.ToolSystem.GetEntityTools(entity) {
			if tool.Type == toolType {
				hasThisType = true
				break
			}
		}
		
		if !hasThisType {
			score += 0.5
		}
		
		if score > bestScore {
			bestScore = score
			bestType = int(toolType)
		}
	}
	
	return bestType
}

func (ebs *EmergentBehaviorSystem) findGoodTrapLocation(entity *Entity, world *World) Position {
	// Find a location near resources but not too close to entity
	bestPos := entity.Position
	bestScore := 0.0
	
	for i := 0; i < 10; i++ {
		angle := rand.Float64() * 2 * math.Pi
		distance := 3.0 + rand.Float64()*5.0
		
		testPos := Position{
			X: entity.Position.X + math.Cos(angle)*distance,
			Y: entity.Position.Y + math.Sin(angle)*distance,
		}
		
		// Count nearby resources
		resourceCount := 0
		for _, plant := range world.AllPlants {
			if !plant.IsAlive {
				continue
			}
			
			plantDistance := math.Sqrt(math.Pow(testPos.X-plant.Position.X, 2) + 
									  math.Pow(testPos.Y-plant.Position.Y, 2))
			
			if plantDistance < 4.0 {
				resourceCount++
			}
		}
		
		score := float64(resourceCount)
		if score > bestScore {
			bestScore = score
			bestPos = testPos
		}
	}
	
	return bestPos
}

// GetBehaviorStats returns statistics about emergent behaviors
func (ebs *EmergentBehaviorSystem) GetBehaviorStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalEntities := len(ebs.BehaviorPatterns)
	avgProficiency := make(map[string]float64)
	behaviorCounts := make(map[string]int)
	
	for _, pattern := range ebs.BehaviorPatterns {
		for behaviorName, proficiency := range pattern.KnownBehaviors {
			if proficiency > 0.0 {
				behaviorCounts[behaviorName]++
				avgProficiency[behaviorName] += proficiency
			}
		}
	}
	
	// Calculate averages
	for behaviorName, total := range avgProficiency {
		if behaviorCounts[behaviorName] > 0 {
			avgProficiency[behaviorName] = total / float64(behaviorCounts[behaviorName])
		}
	}
	
	stats["total_entities"] = totalEntities
	stats["behavior_spread"] = behaviorCounts
	stats["avg_proficiency"] = avgProficiency
	stats["discovered_behaviors"] = len(ebs.LearnedBehaviors)
	
	return stats
}