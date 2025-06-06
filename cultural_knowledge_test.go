package main

import (
	"testing"
)

// TestCulturalKnowledgeSystemBasics tests basic functionality of the cultural knowledge system
func TestCulturalKnowledgeSystemBasics(t *testing.T) {
	system := NewCulturalKnowledgeSystem()

	// Create test entities with all required traits for cultural knowledge
	traitNames := []string{
		"intelligence", "cooperation", "curiosity", "tool_use", "foraging_efficiency",
		"vigilance", "territorial_range", "aggression", "communication_skill",
	}

	entity1 := NewEntity(1, traitNames, "testspecies", Position{X: 0, Y: 0})
	entity1.SetTrait("intelligence", 0.8)
	entity1.SetTrait("cooperation", 0.7)
	entity1.SetTrait("curiosity", 0.6)
	entity1.SetTrait("tool_use", 0.5)
	entity1.SetTrait("foraging_efficiency", 0.6)
	entity1.SetTrait("vigilance", 0.7)
	entity1.SetTrait("territorial_range", 0.8)
	entity1.SetTrait("aggression", 0.3)
	entity1.SetTrait("communication_skill", 0.6)
	entity1.IsAlive = true

	entity2 := NewEntity(2, traitNames, "testspecies", Position{X: 1, Y: 1})
	entity2.SetTrait("intelligence", 0.6)
	entity2.SetTrait("cooperation", 0.8)
	entity2.SetTrait("curiosity", 0.5)
	entity2.SetTrait("tool_use", 0.4)
	entity2.SetTrait("foraging_efficiency", 0.7)
	entity2.SetTrait("vigilance", 0.6)
	entity2.SetTrait("territorial_range", 0.5)
	entity2.SetTrait("aggression", 0.4)
	entity2.SetTrait("communication_skill", 0.7)
	entity2.IsAlive = true

	entities := []*Entity{entity1, entity2}

	// Test entity registration
	system.Update(entities, 0)

	if len(system.EntityMemories) != 2 {
		t.Errorf("Expected 2 entity memories, got %d", len(system.EntityMemories))
	}

	// Check that entities have cultural memories
	memory1 := system.EntityMemories[entity1.ID]
	if memory1 == nil {
		t.Error("Entity 1 should have cultural memory")
		return
	}

	memory2 := system.EntityMemories[entity2.ID]
	if memory2 == nil {
		t.Error("Entity 2 should have cultural memory")
	}

	// Check that basic knowledge types exist
	if len(system.AllKnowledge) == 0 {
		t.Error("System should have basic knowledge types")
	}

	// Check that entities have some starting knowledge
	if len(memory1.KnownKnowledge) == 0 {
		t.Error("Entity 1 should have some starting knowledge")
	}

	// Test teaching ability calculation
	if memory1.TeachingAbility <= 0.0 {
		t.Errorf("Entity 1 should have positive teaching ability, got %f", memory1.TeachingAbility)
	}

	if memory1.LearningAbility <= 0.0 {
		t.Errorf("Entity 1 should have positive learning ability, got %f", memory1.LearningAbility)
	}
}

// TestCulturalKnowledgeTransfer tests knowledge transfer between entities
func TestCulturalKnowledgeTransfer(t *testing.T) {
	system := NewCulturalKnowledgeSystem()

	// Create teacher with high traits
	teacher := NewEntity(1, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 0, Y: 0})
	teacher.SetTrait("intelligence", 0.9)
	teacher.SetTrait("cooperation", 0.8)
	teacher.IsAlive = true

	// Create student with medium traits
	student := NewEntity(2, []string{"intelligence", "cooperation"}, "testspecies", Position{X: 1, Y: 1})
	student.SetTrait("intelligence", 0.6)
	student.SetTrait("cooperation", 0.7)
	student.IsAlive = true

	entities := []*Entity{teacher, student}

	// Initialize the system
	system.Update(entities, 0)

	teacherMemory := system.EntityMemories[teacher.ID]
	studentMemory := system.EntityMemories[student.ID]

	// Record initial knowledge counts
	initialTeacherKnowledge := len(teacherMemory.KnownKnowledge)
	initialStudentKnowledge := len(studentMemory.KnownKnowledge)

	t.Logf("Initial knowledge - Teacher: %d, Student: %d", initialTeacherKnowledge, initialStudentKnowledge)

	// Run multiple updates to allow teaching to occur
	for i := 1; i <= 100; i++ {
		system.Update(entities, i)
	}

	// Check if any teaching occurred
	teachingEvents := system.TotalTeachingEvents
	learningEvents := system.TotalLearningEvents

	if teachingEvents > 0 {
		t.Logf("Teaching events occurred: %d", teachingEvents)
	}

	if learningEvents > 0 {
		t.Logf("Learning events occurred: %d", learningEvents)
	}

	// The system should track teaching/learning events even if they're rare
	if teachingEvents > 0 && learningEvents != teachingEvents {
		t.Errorf("Teaching events (%d) should equal learning events (%d)", teachingEvents, learningEvents)
	}
}

// TestCulturalKnowledgeInnovation tests knowledge innovation
func TestCulturalKnowledgeInnovation(t *testing.T) {
	system := NewCulturalKnowledgeSystem()

	// Create highly intelligent entity
	innovator := NewEntity(1, []string{"intelligence", "curiosity"}, "testspecies", Position{X: 0, Y: 0})
	innovator.SetTrait("intelligence", 1.0)
	innovator.SetTrait("curiosity", 1.0)
	innovator.IsAlive = true

	// Increase innovation rate for testing
	system.InnovationRate = 0.1 // 10% chance per tick

	entities := []*Entity{innovator}

	// Run many updates to trigger innovation
	for i := 0; i < 100; i++ {
		system.Update(entities, i)
	}

	// Check if any innovations occurred
	innovations := system.TotalInnovations

	if innovations > 0 {
		t.Logf("Innovations created: %d", innovations)

		// Check if innovations exist in the knowledge base
		innovationCount := 0
		for _, knowledge := range system.AllKnowledge {
			if knowledge.Innovation {
				innovationCount++
			}
		}

		if innovationCount == 0 {
			t.Error("Innovation count should be greater than 0 when innovations were created")
		}
	}
}

// TestCulturalKnowledgeStats tests the statistics function
func TestCulturalKnowledgeStats(t *testing.T) {
	system := NewCulturalKnowledgeSystem()

	// Create test entity with all required traits
	traitNames := []string{
		"intelligence", "cooperation", "curiosity", "tool_use", "foraging_efficiency",
		"vigilance", "territorial_range", "aggression", "communication_skill",
	}

	entity := NewEntity(1, traitNames, "testspecies", Position{X: 0, Y: 0})
	entity.SetTrait("intelligence", 0.7)
	entity.SetTrait("cooperation", 0.6)
	entity.SetTrait("tool_use", 0.5)
	entity.SetTrait("foraging_efficiency", 0.6)
	entity.SetTrait("vigilance", 0.7)
	entity.SetTrait("territorial_range", 0.8)
	entity.SetTrait("aggression", 0.3)
	entity.SetTrait("communication_skill", 0.6)
	entity.SetTrait("curiosity", 0.5)
	entity.IsAlive = true

	entities := []*Entity{entity}

	// Update system multiple times to give entity chance to acquire knowledge
	for i := 0; i < 10; i++ {
		system.Update(entities, i)
	}

	// Get statistics
	stats := system.GetCulturalStats()

	// Check that statistics are populated
	if stats["total_knowledge_types"] == nil {
		t.Error("Stats should include total_knowledge_types")
	}

	if stats["total_entities"] == nil {
		t.Error("Stats should include total_entities")
	}

	if stats["avg_knowledge_per_entity"] == nil {
		t.Error("Stats should include avg_knowledge_per_entity")
	}

	if stats["knowledge_type_distribution"] == nil {
		t.Error("Stats should include knowledge_type_distribution")
	}

	// Verify reasonable values
	totalEntities := stats["total_entities"].(int)
	if totalEntities != 1 {
		t.Errorf("Expected 1 entity, got %d", totalEntities)
	}

	avgKnowledge := stats["avg_knowledge_per_entity"].(float64)
	if avgKnowledge <= 0.0 {
		t.Errorf("Average knowledge per entity should be positive, got %f", avgKnowledge)
	}
}
