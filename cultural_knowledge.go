package main

import (
	"math"
	"math/rand"
)

// KnowledgeType represents different types of cultural knowledge
type KnowledgeType int

const (
	ToolCrafting       KnowledgeType = iota // Knowledge of tool creation and use
	FoodSources                             // Knowledge of food locations and quality
	DangerAwareness                         // Knowledge of threats and how to avoid them
	NavigationSkills                        // Knowledge of territory and landmarks
	SocialCooperation                       // Knowledge of cooperation strategies
	TechnologyUse                           // Knowledge of technology and innovation
	ResourceManagement                      // Knowledge of resource conservation
	TerritoryDefense                        // Knowledge of defensive strategies
)

// String returns string representation of KnowledgeType
func (kt KnowledgeType) String() string {
	switch kt {
	case ToolCrafting:
		return "Tool Crafting"
	case FoodSources:
		return "Food Sources"
	case DangerAwareness:
		return "Danger Awareness"
	case NavigationSkills:
		return "Navigation Skills"
	case SocialCooperation:
		return "Social Cooperation"
	case TechnologyUse:
		return "Technology Use"
	case ResourceManagement:
		return "Resource Management"
	case TerritoryDefense:
		return "Territory Defense"
	default:
		return "Unknown Knowledge"
	}
}

// CulturalKnowledge represents a piece of cultural knowledge
type CulturalKnowledge struct {
	ID              int           `json:"id"`
	Type            KnowledgeType `json:"type"`
	Effectiveness   float64       `json:"effectiveness"`     // 0.0-1.0, how effective this knowledge is
	Complexity      float64       `json:"complexity"`        // 0.0-1.0, how hard it is to learn/teach
	AgeInGeneration int           `json:"age_in_generation"` // How long this knowledge has existed
	TeacherCount    int           `json:"teacher_count"`     // How many entities can teach this
	LearnerCount    int           `json:"learner_count"`     // How many entities have learned this
	SuccessRate     float64       `json:"success_rate"`      // How often teaching this succeeds
	DecayRate       float64       `json:"decay_rate"`        // How fast this knowledge fades without use
	Innovation      bool          `json:"innovation"`        // Whether this is a recent innovation
	LastUsed        int           `json:"last_used"`         // Last tick when this knowledge was used
	Description     string        `json:"description"`       // Human-readable description
}

// CulturalMemory represents an entity's cultural knowledge storage
type CulturalMemory struct {
	EntityID          int                        `json:"entity_id"`
	KnownKnowledge    map[int]*CulturalKnowledge `json:"known_knowledge"`    // Knowledge ID -> Knowledge
	TeachingAbility   float64                    `json:"teaching_ability"`   // 0.0-1.0, how good at teaching
	LearningAbility   float64                    `json:"learning_ability"`   // 0.0-1.0, how good at learning
	InnovationChance  float64                    `json:"innovation_chance"`  // Chance to create new knowledge
	KnowledgeCapacity int                        `json:"knowledge_capacity"` // Max number of knowledge pieces
	RecentlyTaught    []int                      `json:"recently_taught"`    // Knowledge IDs taught recently
	RecentlyLearned   []int                      `json:"recently_learned"`   // Knowledge IDs learned recently
	MentorEntityID    int                        `json:"mentor_entity_id"`   // Current mentor (if any)
	StudentEntityIDs  []int                      `json:"student_entity_ids"` // Current students
}

// CulturalKnowledgeSystem manages cultural knowledge transfer and evolution
type CulturalKnowledgeSystem struct {
	AllKnowledge        map[int]*CulturalKnowledge `json:"all_knowledge"`   // All existing knowledge
	EntityMemories      map[int]*CulturalMemory    `json:"entity_memories"` // Entity ID -> Cultural Memory
	NextKnowledgeID     int                        `json:"next_knowledge_id"`
	GenerationCount     int                        `json:"generation_count"` // Track generations for knowledge aging
	TotalTeachingEvents int                        `json:"total_teaching_events"`
	TotalLearningEvents int                        `json:"total_learning_events"`
	TotalInnovations    int                        `json:"total_innovations"`
	KnowledgeLossEvents int                        `json:"knowledge_loss_events"`

	// System parameters
	BaseTeachingChance  float64 `json:"base_teaching_chance"`  // Base chance for teaching to occur
	BaseLearningSuccess float64 `json:"base_learning_success"` // Base success rate for learning
	InnovationRate      float64 `json:"innovation_rate"`       // Rate of new knowledge creation
	KnowledgeDecayRate  float64 `json:"knowledge_decay_rate"`  // Rate at which unused knowledge fades
}

// NewCulturalKnowledgeSystem creates a new cultural knowledge system
func NewCulturalKnowledgeSystem() *CulturalKnowledgeSystem {
	return &CulturalKnowledgeSystem{
		AllKnowledge:        make(map[int]*CulturalKnowledge),
		EntityMemories:      make(map[int]*CulturalMemory),
		NextKnowledgeID:     1,
		GenerationCount:     0,
		BaseTeachingChance:  0.05,   // 5% chance per tick when near teacher
		BaseLearningSuccess: 0.7,    // 70% success rate for learning
		InnovationRate:      0.001,  // 0.1% chance per tick for innovation
		KnowledgeDecayRate:  0.0001, // Very slow decay
	}
}

// RegisterEntity creates cultural memory for a new entity
func (cks *CulturalKnowledgeSystem) RegisterEntity(entity *Entity) {
	if cks.EntityMemories[entity.ID] != nil {
		return // Already registered
	}

	// Create cultural memory based on entity traits
	memory := &CulturalMemory{
		EntityID:          entity.ID,
		KnownKnowledge:    make(map[int]*CulturalKnowledge),
		TeachingAbility:   math.Min(1.0, entity.GetTrait("intelligence")*0.5+entity.GetTrait("cooperation")*0.3+entity.GetTrait("communication_skill")*0.2),
		LearningAbility:   math.Min(1.0, entity.GetTrait("intelligence")*0.6+entity.GetTrait("curiosity")*0.4),
		InnovationChance:  math.Min(0.05, entity.GetTrait("intelligence")*0.02+entity.GetTrait("curiosity")*0.02),
		KnowledgeCapacity: int(10 + entity.GetTrait("intelligence")*20), // 10-30 knowledge pieces
		RecentlyTaught:    make([]int, 0),
		RecentlyLearned:   make([]int, 0),
		MentorEntityID:    -1,
		StudentEntityIDs:  make([]int, 0),
	}

	cks.EntityMemories[entity.ID] = memory

	// Give new entities some basic knowledge based on their traits
	cks.initializeBasicKnowledge(entity, memory)
}

// initializeBasicKnowledge gives new entities some starting cultural knowledge
func (cks *CulturalKnowledgeSystem) initializeBasicKnowledge(entity *Entity, memory *CulturalMemory) {
	// Create basic knowledge types if they don't exist
	cks.ensureBasicKnowledgeExists()

	// Give entities knowledge based on their traits
	for _, knowledge := range cks.AllKnowledge {
		learnChance := 0.0

		switch knowledge.Type {
		case ToolCrafting:
			learnChance = entity.GetTrait("tool_use") * 0.3
		case FoodSources:
			learnChance = entity.GetTrait("foraging_efficiency") * 0.4
		case DangerAwareness:
			learnChance = entity.GetTrait("vigilance") * 0.5
		case NavigationSkills:
			learnChance = entity.GetTrait("territorial_range") * 0.01 // Higher range = better navigation
		case SocialCooperation:
			learnChance = entity.GetTrait("cooperation") * 0.6
		case TechnologyUse:
			learnChance = entity.GetTrait("intelligence") * 0.2
		case ResourceManagement:
			learnChance = entity.GetTrait("intelligence") * 0.3
		case TerritoryDefense:
			learnChance = entity.GetTrait("aggression")*0.2 + entity.GetTrait("cooperation")*0.2
		}

		if rand.Float64() < learnChance {
			cks.learnKnowledge(memory, knowledge)
		}
	}
}

// ensureBasicKnowledgeExists creates fundamental knowledge types if they don't exist
func (cks *CulturalKnowledgeSystem) ensureBasicKnowledgeExists() {
	basicKnowledge := []struct {
		Type          KnowledgeType
		Effectiveness float64
		Complexity    float64
		Description   string
	}{
		{ToolCrafting, 0.6, 0.7, "Basic tool creation and maintenance"},
		{FoodSources, 0.8, 0.3, "Knowledge of safe and nutritious food sources"},
		{DangerAwareness, 0.9, 0.4, "Recognition and avoidance of common threats"},
		{NavigationSkills, 0.7, 0.5, "Territory navigation and landmark recognition"},
		{SocialCooperation, 0.8, 0.6, "Cooperation strategies and social coordination"},
		{TechnologyUse, 0.5, 0.8, "Use and improvement of existing technologies"},
		{ResourceManagement, 0.7, 0.7, "Efficient resource allocation and conservation"},
		{TerritoryDefense, 0.6, 0.6, "Defensive strategies and territory protection"},
	}

	for _, kb := range basicKnowledge {
		exists := false
		for _, existing := range cks.AllKnowledge {
			if existing.Type == kb.Type {
				exists = true
				break
			}
		}

		if !exists {
			knowledge := &CulturalKnowledge{
				ID:              cks.NextKnowledgeID,
				Type:            kb.Type,
				Effectiveness:   kb.Effectiveness,
				Complexity:      kb.Complexity,
				AgeInGeneration: 0,
				TeacherCount:    0,
				LearnerCount:    0,
				SuccessRate:     cks.BaseLearningSuccess,
				DecayRate:       cks.KnowledgeDecayRate,
				Innovation:      false,
				LastUsed:        0,
				Description:     kb.Description,
			}

			cks.AllKnowledge[cks.NextKnowledgeID] = knowledge
			cks.NextKnowledgeID++
		}
	}
}

// Update processes cultural knowledge for one tick
func (cks *CulturalKnowledgeSystem) Update(entities []*Entity, tick int) {
	// Register new entities
	for _, entity := range entities {
		if entity.IsAlive {
			cks.RegisterEntity(entity)
		}
	}

	// Process knowledge teaching and learning
	cks.processTeachingAndLearning(entities, tick)

	// Process knowledge innovation
	cks.processInnovation(entities, tick)

	// Process knowledge decay
	cks.processKnowledgeDecay(tick)

	// Update statistics
	cks.updateStatistics()

	// Clean up memories of dead entities
	cks.cleanupDeadEntities(entities)
}

// processTeachingAndLearning handles knowledge transfer between entities
func (cks *CulturalKnowledgeSystem) processTeachingAndLearning(entities []*Entity, tick int) {
	nearbyPairs := cks.findNearbyEntityPairs(entities)

	for _, pair := range nearbyPairs {
		teacher := pair[0]
		student := pair[1]

		teacherMemory := cks.EntityMemories[teacher.ID]
		studentMemory := cks.EntityMemories[student.ID]

		if teacherMemory == nil || studentMemory == nil {
			continue
		}

		// Check if teaching should occur
		teachingChance := cks.BaseTeachingChance * teacherMemory.TeachingAbility

		// Cooperation increases teaching likelihood
		if teacher.GetTrait("cooperation") > 0.7 && student.GetTrait("cooperation") > 0.7 {
			teachingChance *= 2.0
		}

		if rand.Float64() < teachingChance {
			cks.attemptKnowledgeTransfer(teacherMemory, studentMemory, tick)
		}
	}
}

// findNearbyEntityPairs finds entities close enough for knowledge transfer
func (cks *CulturalKnowledgeSystem) findNearbyEntityPairs(entities []*Entity) [][2]*Entity {
	pairs := make([][2]*Entity, 0)

	for i, entity1 := range entities {
		if !entity1.IsAlive {
			continue
		}

		for j := i + 1; j < len(entities); j++ {
			entity2 := entities[j]
			if !entity2.IsAlive {
				continue
			}

			// Check if entities are close enough (within 3 units)
			distance := math.Sqrt(math.Pow(entity1.Position.X-entity2.Position.X, 2) +
				math.Pow(entity1.Position.Y-entity2.Position.Y, 2))

			if distance <= 3.0 {
				pairs = append(pairs, [2]*Entity{entity1, entity2})
			}
		}
	}

	return pairs
}

// attemptKnowledgeTransfer tries to transfer knowledge from teacher to student
func (cks *CulturalKnowledgeSystem) attemptKnowledgeTransfer(teacher, student *CulturalMemory, tick int) {
	// Find knowledge that teacher has but student doesn't
	teachableKnowledge := make([]*CulturalKnowledge, 0)

	for knowledgeID, knowledge := range teacher.KnownKnowledge {
		if student.KnownKnowledge[knowledgeID] == nil {
			// Student doesn't have this knowledge
			if len(student.KnownKnowledge) < student.KnowledgeCapacity {
				teachableKnowledge = append(teachableKnowledge, knowledge)
			}
		}
	}

	if len(teachableKnowledge) == 0 {
		return // Nothing to teach
	}

	// Select random knowledge to teach
	knowledge := teachableKnowledge[rand.Intn(len(teachableKnowledge))]

	// Calculate learning success chance
	successChance := cks.BaseLearningSuccess * student.LearningAbility

	// Complexity affects learning success
	complexityPenalty := knowledge.Complexity * 0.3
	successChance *= (1.0 - complexityPenalty)

	// Teacher's ability affects success
	successChance *= (0.5 + teacher.TeachingAbility*0.5)

	if rand.Float64() < successChance {
		// Learning successful
		cks.learnKnowledge(student, knowledge)
		teacher.RecentlyTaught = append(teacher.RecentlyTaught, knowledge.ID)
		student.RecentlyLearned = append(student.RecentlyLearned, knowledge.ID)

		knowledge.LastUsed = tick
		cks.TotalTeachingEvents++
		cks.TotalLearningEvents++

		// Update knowledge statistics
		knowledge.TeacherCount = cks.countKnowledgeTeachers(knowledge.ID)
		knowledge.LearnerCount = cks.countKnowledgeLearners(knowledge.ID)
	}
}

// learnKnowledge adds knowledge to an entity's memory
func (cks *CulturalKnowledgeSystem) learnKnowledge(memory *CulturalMemory, knowledge *CulturalKnowledge) {
	// Create a copy of the knowledge for this entity
	personalKnowledge := &CulturalKnowledge{
		ID:              knowledge.ID,
		Type:            knowledge.Type,
		Effectiveness:   knowledge.Effectiveness,
		Complexity:      knowledge.Complexity,
		AgeInGeneration: knowledge.AgeInGeneration,
		TeacherCount:    knowledge.TeacherCount,
		LearnerCount:    knowledge.LearnerCount,
		SuccessRate:     knowledge.SuccessRate,
		DecayRate:       knowledge.DecayRate,
		Innovation:      knowledge.Innovation,
		LastUsed:        knowledge.LastUsed,
		Description:     knowledge.Description,
	}

	memory.KnownKnowledge[knowledge.ID] = personalKnowledge
}

// processInnovation handles creation of new cultural knowledge
func (cks *CulturalKnowledgeSystem) processInnovation(entities []*Entity, tick int) {
	for _, entity := range entities {
		if !entity.IsAlive {
			continue
		}

		memory := cks.EntityMemories[entity.ID]
		if memory == nil {
			continue
		}

		// Innovation chance based on entity traits and existing knowledge
		innovationChance := memory.InnovationChance

		// Having more knowledge increases innovation chance
		knowledgeBonus := float64(len(memory.KnownKnowledge)) * 0.001
		innovationChance += knowledgeBonus

		if rand.Float64() < innovationChance {
			cks.createInnovation(entity, memory, tick)
		}
	}
}

// createInnovation creates new cultural knowledge
func (cks *CulturalKnowledgeSystem) createInnovation(entity *Entity, memory *CulturalMemory, tick int) {
	// Determine type of innovation based on entity traits
	innovationType := cks.selectInnovationType(entity)

	// Create new knowledge
	effectiveness := 0.3 + rand.Float64()*0.4 // 0.3-0.7 initial effectiveness
	complexity := 0.5 + rand.Float64()*0.3    // 0.5-0.8 complexity for innovations

	innovation := &CulturalKnowledge{
		ID:              cks.NextKnowledgeID,
		Type:            innovationType,
		Effectiveness:   effectiveness,
		Complexity:      complexity,
		AgeInGeneration: 0,
		TeacherCount:    1,                             // The innovator can teach it
		LearnerCount:    1,                             // The innovator knows it
		SuccessRate:     cks.BaseLearningSuccess * 0.8, // New knowledge is harder to teach
		DecayRate:       cks.KnowledgeDecayRate * 1.5,  // New knowledge fades faster initially
		Innovation:      true,
		LastUsed:        tick,
		Description:     "Innovation: " + innovationType.String() + " by entity " + string(rune(entity.ID)),
	}

	cks.AllKnowledge[cks.NextKnowledgeID] = innovation
	cks.NextKnowledgeID++
	cks.TotalInnovations++

	// Innovator learns their own innovation
	cks.learnKnowledge(memory, innovation)
}

// selectInnovationType chooses what type of knowledge to innovate based on entity traits
func (cks *CulturalKnowledgeSystem) selectInnovationType(entity *Entity) KnowledgeType {
	// Weighted selection based on entity traits
	weights := map[KnowledgeType]float64{
		ToolCrafting:       entity.GetTrait("tool_use"),
		FoodSources:        entity.GetTrait("foraging_efficiency"),
		DangerAwareness:    entity.GetTrait("vigilance"),
		NavigationSkills:   entity.GetTrait("territorial_range") * 0.01,
		SocialCooperation:  entity.GetTrait("cooperation"),
		TechnologyUse:      entity.GetTrait("intelligence"),
		ResourceManagement: entity.GetTrait("intelligence") * 0.8,
		TerritoryDefense:   entity.GetTrait("aggression")*0.5 + entity.GetTrait("cooperation")*0.5,
	}

	// Convert to cumulative weights
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		// Fallback to random selection
		return KnowledgeType(rand.Intn(8))
	}

	randValue := rand.Float64() * totalWeight
	cumulative := 0.0

	for knowledgeType, weight := range weights {
		cumulative += weight
		if randValue <= cumulative {
			return knowledgeType
		}
	}

	return ToolCrafting // Fallback
}

// processKnowledgeDecay handles the gradual loss of unused knowledge
func (cks *CulturalKnowledgeSystem) processKnowledgeDecay(tick int) {
	for _, memory := range cks.EntityMemories {
		for knowledgeID, knowledge := range memory.KnownKnowledge {
			timeSinceLastUse := tick - knowledge.LastUsed

			// Knowledge decays if not used recently
			if timeSinceLastUse > 1000 { // After 1000 ticks without use
				decayChance := knowledge.DecayRate * float64(timeSinceLastUse-1000) * 0.001

				if rand.Float64() < decayChance {
					delete(memory.KnownKnowledge, knowledgeID)
					cks.KnowledgeLossEvents++
				}
			}
		}
	}
}

// updateStatistics updates knowledge system statistics
func (cks *CulturalKnowledgeSystem) updateStatistics() {
	for _, knowledge := range cks.AllKnowledge {
		knowledge.TeacherCount = cks.countKnowledgeTeachers(knowledge.ID)
		knowledge.LearnerCount = cks.countKnowledgeLearners(knowledge.ID)

		// Age the knowledge
		knowledge.AgeInGeneration++

		// Innovations become regular knowledge after some time
		if knowledge.Innovation && knowledge.AgeInGeneration > 500 {
			knowledge.Innovation = false
			knowledge.DecayRate = cks.KnowledgeDecayRate // Normalize decay rate
		}
	}
}

// countKnowledgeTeachers counts how many entities can teach specific knowledge
func (cks *CulturalKnowledgeSystem) countKnowledgeTeachers(knowledgeID int) int {
	count := 0
	for _, memory := range cks.EntityMemories {
		if memory.KnownKnowledge[knowledgeID] != nil && memory.TeachingAbility > 0.3 {
			count++
		}
	}
	return count
}

// countKnowledgeLearners counts how many entities have specific knowledge
func (cks *CulturalKnowledgeSystem) countKnowledgeLearners(knowledgeID int) int {
	count := 0
	for _, memory := range cks.EntityMemories {
		if memory.KnownKnowledge[knowledgeID] != nil {
			count++
		}
	}
	return count
}

// cleanupDeadEntities removes cultural memories of dead entities
func (cks *CulturalKnowledgeSystem) cleanupDeadEntities(entities []*Entity) {
	aliveEntities := make(map[int]bool)
	for _, entity := range entities {
		if entity.IsAlive {
			aliveEntities[entity.ID] = true
		}
	}

	// Remove memories of dead entities
	for entityID := range cks.EntityMemories {
		if !aliveEntities[entityID] {
			delete(cks.EntityMemories, entityID)
		}
	}
}

// GetCulturalStats returns statistics about the cultural knowledge system
func (cks *CulturalKnowledgeSystem) GetCulturalStats() map[string]interface{} {
	totalKnowledge := len(cks.AllKnowledge)
	totalEntitiesWithKnowledge := len(cks.EntityMemories)

	// Count innovations
	innovationCount := 0
	for _, knowledge := range cks.AllKnowledge {
		if knowledge.Innovation {
			innovationCount++
		}
	}

	// Knowledge type distribution
	typeDistribution := make(map[string]int)
	for _, knowledge := range cks.AllKnowledge {
		typeDistribution[knowledge.Type.String()]++
	}

	// Average knowledge per entity
	totalKnowledgeInstances := 0
	for _, memory := range cks.EntityMemories {
		totalKnowledgeInstances += len(memory.KnownKnowledge)
	}

	avgKnowledgePerEntity := 0.0
	if totalEntitiesWithKnowledge > 0 {
		avgKnowledgePerEntity = float64(totalKnowledgeInstances) / float64(totalEntitiesWithKnowledge)
	}

	return map[string]interface{}{
		"total_knowledge_types":       totalKnowledge,
		"total_entities":              totalEntitiesWithKnowledge,
		"active_innovations":          innovationCount,
		"total_teaching_events":       cks.TotalTeachingEvents,
		"total_learning_events":       cks.TotalLearningEvents,
		"total_innovations_created":   cks.TotalInnovations,
		"knowledge_loss_events":       cks.KnowledgeLossEvents,
		"avg_knowledge_per_entity":    avgKnowledgePerEntity,
		"knowledge_type_distribution": typeDistribution,
	}
}
