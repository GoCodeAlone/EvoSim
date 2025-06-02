package main

import (
	"math"
	"testing"
)

func TestInsectSystem(t *testing.T) {
	// Create insect system
	is := NewInsectSystem()
	
	if len(is.PheromoneTrails) != 0 {
		t.Error("New insect system should have no pheromone trails")
	}
	if len(is.SwarmUnits) != 0 {
		t.Error("New insect system should have no swarm units")
	}
	if is.NextTrailID != 1 {
		t.Error("Next trail ID should start at 1")
	}
	if is.NextSwarmID != 1 {
		t.Error("Next swarm ID should start at 1")
	}
}

func TestAddInsectTraitsToEntity(t *testing.T) {
	// Test entity that should get insect traits (small, cooperative, intelligent)
	insectLike := NewEntity(1, []string{"size", "cooperation", "intelligence"}, "test", Position{})
	insectLike.SetTrait("size", -0.5)        // Small
	insectLike.SetTrait("cooperation", 0.6)  // Cooperative
	insectLike.SetTrait("intelligence", 0.5) // Intelligent

	AddInsectTraitsToEntity(insectLike)

	// Check that insect traits were added
	if insectLike.GetTrait("swarm_capability") == 0.0 {
		t.Error("Insect-like entity should have swarm capability")
	}
	if insectLike.GetTrait("pheromone_sensitivity") == 0.0 {
		t.Error("Insect-like entity should have pheromone sensitivity")
	}
	if insectLike.GetTrait("pheromone_production") == 0.0 {
		t.Error("Insect-like entity should have pheromone production")
	}
	if insectLike.GetTrait("colony_loyalty") == 0.0 {
		t.Error("Insect-like entity should have colony loyalty")
	}

	// Test entity that should NOT get insect traits (large, non-cooperative)
	nonInsectLike := NewEntity(2, []string{"size", "cooperation", "intelligence"}, "test", Position{})
	nonInsectLike.SetTrait("size", 0.5)        // Large
	nonInsectLike.SetTrait("cooperation", 0.2) // Non-cooperative
	nonInsectLike.SetTrait("intelligence", 0.2) // Low intelligence

	AddInsectTraitsToEntity(nonInsectLike)

	// Check that insect traits were NOT added
	if nonInsectLike.GetTrait("swarm_capability") != 0.0 {
		t.Error("Non-insect entity should not have swarm capability")
	}
}

func TestPheromoneTrailCreation(t *testing.T) {
	is := NewInsectSystem()
	
	// Create entity with pheromone production ability
	entity := NewEntity(1, []string{"pheromone_production"}, "test", Position{X: 0, Y: 0})
	entity.SetTrait("pheromone_production", 0.8)

	startPos := Position{X: 0, Y: 0}
	endPos := Position{X: 10, Y: 10}

	// Create pheromone trail
	trail := is.CreatePheromoneTrail(entity, TrailPheromone, startPos, endPos)

	if trail == nil {
		t.Fatal("Should be able to create pheromone trail with sufficient production")
	}

	if trail.Type != TrailPheromone {
		t.Error("Trail should have correct type")
	}
	if trail.ProducerID != entity.ID {
		t.Error("Trail should have correct producer ID")
	}
	if len(trail.Positions) < 3 {
		t.Error("Trail should have at least 3 positions")
	}
	if len(trail.Strength) != len(trail.Positions) {
		t.Error("Trail should have strength values for each position")
	}

	// Check that trail was added to system
	if len(is.PheromoneTrails) != 1 {
		t.Error("Trail should be added to insect system")
	}

	// Test entity with low pheromone production
	weakEntity := NewEntity(2, []string{"pheromone_production"}, "test", Position{})
	weakEntity.SetTrait("pheromone_production", 0.1) // Too low

	weakTrail := is.CreatePheromoneTrail(weakEntity, TrailPheromone, startPos, endPos)
	if weakTrail != nil {
		t.Error("Should not be able to create trail with low pheromone production")
	}
}

func TestPheromoneTrailFollowing(t *testing.T) {
	is := NewInsectSystem()
	
	// Create entity with pheromone sensitivity
	follower := NewEntity(1, []string{"pheromone_sensitivity"}, "test", Position{X: 5, Y: 5})
	follower.SetTrait("pheromone_sensitivity", 0.8)

	producer := NewEntity(2, []string{"pheromone_production"}, "test", Position{})
	producer.SetTrait("pheromone_production", 0.8)

	// Create trail near follower
	startPos := Position{X: 0, Y: 0}
	endPos := Position{X: 10, Y: 10}
	trail := is.CreatePheromoneTrail(producer, FoodPheromone, startPos, endPos)
	
	if trail == nil {
		t.Fatal("Should be able to create pheromone trail")
	}

	// Try to follow trail
	targetX, targetY, found := is.FollowPheromoneTrail(follower, FoodPheromone)

	if !found {
		t.Error("Should be able to follow nearby pheromone trail")
	}

	// Check that target is reasonable (should be along the trail)
	distance := math.Sqrt(math.Pow(targetX-follower.Position.X, 2) + math.Pow(targetY-follower.Position.Y, 2))
	if distance > 20.0 {
		t.Error("Target should be reasonably close to follower")
	}

	// Test entity with low sensitivity
	insensitive := NewEntity(3, []string{"pheromone_sensitivity"}, "test", Position{X: 5, Y: 5})
	insensitive.SetTrait("pheromone_sensitivity", 0.1) // Too low

	_, _, foundByInsensitive := is.FollowPheromoneTrail(insensitive, FoodPheromone)
	if foundByInsensitive {
		t.Error("Entity with low sensitivity should not be able to follow trail")
	}

	// Test wrong pheromone type
	_, _, foundWrongType := is.FollowPheromoneTrail(follower, AlarmPheromone)
	if foundWrongType {
		t.Error("Should not find trail of different type")
	}
}

func TestPheromoneTrailDecay(t *testing.T) {
	is := NewInsectSystem()
	
	// Create entity and trail
	entity := NewEntity(1, []string{"pheromone_production"}, "test", Position{})
	entity.SetTrait("pheromone_production", 0.8)

	startPos := Position{X: 0, Y: 0}
	endPos := Position{X: 10, Y: 10}
	trail := is.CreatePheromoneTrail(entity, TrailPheromone, startPos, endPos)

	if trail == nil {
		t.Fatal("Failed to create trail")
	}

	originalStrength := trail.Strength[0]

	// Simulate several updates to decay the trail
	for i := 0; i < 50; i++ {
		is.updatePheromoneTrails()
	}

	// Trail should still exist but be weaker
	if len(is.PheromoneTrails) != 1 {
		t.Error("Trail should still exist after moderate decay")
	}

	if trail.Strength[0] >= originalStrength {
		t.Error("Trail strength should have decreased")
	}

	// Simulate many more updates to completely decay the trail
	for i := 0; i < 1000; i++ {
		is.updatePheromoneTrails()
	}

	// Trail should be removed
	if len(is.PheromoneTrails) != 0 {
		t.Error("Trail should be removed after extensive decay")
	}
}

func TestSwarmUnitCreation(t *testing.T) {
	is := NewInsectSystem()
	
	// Create entities suitable for swarming
	entities := make([]*Entity, 6)
	traitNames := []string{"swarm_capability", "cooperation", "intelligence"}
	
	for i := 0; i < 6; i++ {
		entities[i] = NewEntity(i+1, traitNames, "test_species", Position{X: float64(i * 2), Y: 0})
		entities[i].SetTrait("swarm_capability", 0.5 + float64(i)*0.1)
		entities[i].SetTrait("cooperation", 0.6 + float64(i)*0.05)
		entities[i].SetTrait("intelligence", 0.4 + float64(i)*0.1)
	}

	// Create swarm
	swarm := is.CreateSwarmUnit(entities, "foraging")

	if swarm == nil {
		t.Fatal("Should be able to create swarm from suitable entities")
	}

	if len(swarm.Members) != 6 {
		t.Errorf("Expected 6 swarm members, got %d", len(swarm.Members))
	}

	if swarm.SwarmPurpose != "foraging" {
		t.Error("Swarm should have correct purpose")
	}

	if swarm.LeaderEntity == nil {
		t.Error("Swarm should have a leader")
	}

	// Check that members are marked as swarm members
	for _, member := range swarm.Members {
		if member.GetTrait("swarm_member") != 1.0 {
			t.Error("Swarm members should be marked as swarm members")
		}
		if member.GetTrait("swarm_id") != float64(swarm.ID) {
			t.Error("Swarm members should have correct swarm ID")
		}
	}

	// Check that swarm was added to system
	if len(is.SwarmUnits) != 1 {
		t.Error("Swarm should be added to insect system")
	}

	// Test insufficient entities for swarm
	tooFew := entities[:3] // Only 3 entities
	noSwarm := is.CreateSwarmUnit(tooFew, "foraging")
	if noSwarm != nil {
		t.Error("Should not be able to create swarm with too few entities")
	}

	// Test entities unsuitable for swarming
	unsuitableEntities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		unsuitableEntities[i] = NewEntity(i+10, traitNames, "test", Position{})
		unsuitableEntities[i].SetTrait("swarm_capability", 0.1) // Too low
		unsuitableEntities[i].SetTrait("cooperation", 0.2)      // Too low
	}

	unsuitableSwarm := is.CreateSwarmUnit(unsuitableEntities, "foraging")
	if unsuitableSwarm != nil {
		t.Error("Should not be able to create swarm from unsuitable entities")
	}
}

func TestSwarmMovement(t *testing.T) {
	is := NewInsectSystem()
	
	// Create swarm
	entities := make([]*Entity, 5)
	traitNames := []string{"swarm_capability", "cooperation", "speed"}
	
	for i := 0; i < 5; i++ {
		entities[i] = NewEntity(i+1, traitNames, "test", Position{X: float64(i * 2), Y: 0})
		entities[i].SetTrait("swarm_capability", 0.7)
		entities[i].SetTrait("cooperation", 0.8)
		entities[i].SetTrait("speed", 0.6)
	}

	swarm := is.CreateSwarmUnit(entities, "migration")
	if swarm == nil {
		t.Fatal("Failed to create swarm")
	}

	// Set target position
	swarm.TargetPosition = Position{X: 20, Y: 20}

	// Record original positions
	originalPositions := make([]Position, len(swarm.Members))
	for i, member := range swarm.Members {
		originalPositions[i] = member.Position
	}

	// Update swarm movement
	is.UpdateSwarmMovement(swarm)

	// Check that members moved
	moved := false
	for i, member := range swarm.Members {
		if member.Position.X != originalPositions[i].X || member.Position.Y != originalPositions[i].Y {
			moved = true
			break
		}
	}

	if !moved {
		t.Error("Swarm members should have moved toward target")
	}

	// Check that swarm maintains formation (members should be close to each other)
	maxDistance := 0.0
	for i := 0; i < len(swarm.Members); i++ {
		for j := i + 1; j < len(swarm.Members); j++ {
			distance := swarm.Members[i].DistanceTo(swarm.Members[j])
			if distance > maxDistance {
				maxDistance = distance
			}
		}
	}

	if maxDistance > swarm.SwarmRadius*2 {
		t.Error("Swarm members should maintain formation")
	}
}

func TestSwarmFormations(t *testing.T) {
	is := NewInsectSystem()
	
	// Create test swarm
	entities := make([]*Entity, 4)
	for i := 0; i < 4; i++ {
		entities[i] = NewEntity(i+1, []string{"swarm_capability", "cooperation"}, "test", Position{})
		entities[i].SetTrait("swarm_capability", 0.7)
		entities[i].SetTrait("cooperation", 0.8)
	}

	// Test different formation types
	purposes := []string{"foraging", "defense", "migration"}
	
	for _, purpose := range purposes {
		swarm := is.CreateSwarmUnit(entities, purpose)
		if swarm == nil {
			t.Fatalf("Failed to create %s swarm", purpose)
		}

		swarm.TargetPosition = Position{X: 10, Y: 10}
		formations := is.calculateSwarmFormation(swarm)

		if len(formations) != len(swarm.Members) {
			t.Errorf("Formation should have position for each member in %s swarm", purpose)
		}

		// All formations should be valid positions
		for i, pos := range formations {
			if math.IsNaN(pos.X) || math.IsNaN(pos.Y) {
				t.Errorf("Formation position %d should be valid in %s swarm", i, purpose)
			}
		}
	}
}

func TestSwarmUpdate(t *testing.T) {
	is := NewInsectSystem()
	
	// Create swarm with some entities
	entities := make([]*Entity, 5)
	for i := 0; i < 5; i++ {
		entities[i] = NewEntity(i+1, []string{"swarm_capability", "cooperation"}, "test", Position{})
		entities[i].SetTrait("swarm_capability", 0.7)
		entities[i].SetTrait("cooperation", 0.8)
	}

	swarm := is.CreateSwarmUnit(entities, "foraging")
	if swarm == nil {
		t.Fatal("Failed to create swarm")
	}

	// Kill some members
	entities[0].IsAlive = false
	entities[1].IsAlive = false

	// Update system
	is.updateSwarmUnits()

	// Check that dead members were removed
	if len(swarm.Members) != 3 {
		t.Errorf("Expected 3 alive members, got %d", len(swarm.Members))
	}

	// Check that swarm markings were removed from dead entities
	if entities[0].GetTrait("swarm_member") != 0.0 {
		t.Error("Dead entity should not have swarm member marking")
	}
	if entities[1].GetTrait("swarm_member") != 0.0 {
		t.Error("Dead entity should not have swarm member marking")
	}

	// Kill more members to disband swarm
	entities[2].IsAlive = false
	entities[3].IsAlive = false

	is.updateSwarmUnits()

	// Swarm should be disbanded (less than 3 members)
	if len(is.SwarmUnits) != 0 {
		t.Error("Swarm should be disbanded when too few members remain")
	}

	// Remaining entity should not have swarm markings
	if entities[4].GetTrait("swarm_member") != 0.0 {
		t.Error("Remaining entity should not have swarm member marking after disbanding")
	}
}

func TestIsEntityInsectLike(t *testing.T) {
	// Test insect-like entity
	insectLike := NewEntity(1, []string{"size", "cooperation", "swarm_capability"}, "test", Position{})
	insectLike.SetTrait("size", -0.3)          // Small
	insectLike.SetTrait("cooperation", 0.6)    // Cooperative
	insectLike.SetTrait("swarm_capability", 0.5) // Swarm capable

	if !IsEntityInsectLike(insectLike) {
		t.Error("Entity with small size, high cooperation, and swarm capability should be insect-like")
	}

	// Test non-insect-like entity (large)
	large := NewEntity(2, []string{"size", "cooperation", "swarm_capability"}, "test", Position{})
	large.SetTrait("size", 0.5)               // Large
	large.SetTrait("cooperation", 0.6)        // Cooperative
	large.SetTrait("swarm_capability", 0.5)   // Swarm capable

	if IsEntityInsectLike(large) {
		t.Error("Large entity should not be considered insect-like")
	}

	// Test non-insect-like entity (non-cooperative)
	nonCooperative := NewEntity(3, []string{"size", "cooperation", "swarm_capability"}, "test", Position{})
	nonCooperative.SetTrait("size", -0.3)     // Small
	nonCooperative.SetTrait("cooperation", 0.2) // Non-cooperative
	nonCooperative.SetTrait("swarm_capability", 0.5) // Swarm capable

	if IsEntityInsectLike(nonCooperative) {
		t.Error("Non-cooperative entity should not be considered insect-like")
	}

	// Test non-insect-like entity (low swarm capability)
	lowSwarm := NewEntity(4, []string{"size", "cooperation", "swarm_capability"}, "test", Position{})
	lowSwarm.SetTrait("size", -0.3)          // Small
	lowSwarm.SetTrait("cooperation", 0.6)    // Cooperative
	lowSwarm.SetTrait("swarm_capability", 0.2) // Low swarm capability

	if IsEntityInsectLike(lowSwarm) {
		t.Error("Entity with low swarm capability should not be considered insect-like")
	}
}

func TestPheromoneStrengthAtPosition(t *testing.T) {
	is := NewInsectSystem()
	
	// Create entity and trail
	entity := NewEntity(1, []string{"pheromone_production"}, "test", Position{})
	entity.SetTrait("pheromone_production", 0.8)

	startPos := Position{X: 0, Y: 0}
	endPos := Position{X: 10, Y: 10}
	trail := is.CreatePheromoneTrail(entity, FoodPheromone, startPos, endPos)

	if trail == nil {
		t.Fatal("Failed to create trail")
	}

	// Test position on trail
	onTrail := Position{X: 5, Y: 5} // Middle of trail
	strength := is.GetPheromoneStrengthAtPosition(onTrail, FoodPheromone)
	
	if strength <= 0 {
		t.Error("Should detect pheromone strength on trail")
	}

	// Test position far from trail
	farAway := Position{X: 50, Y: 50}
	strengthFar := is.GetPheromoneStrengthAtPosition(farAway, FoodPheromone)
	
	if strengthFar > 0 {
		t.Error("Should not detect pheromone strength far from trail")
	}

	// Test wrong pheromone type
	strengthWrong := is.GetPheromoneStrengthAtPosition(onTrail, AlarmPheromone)
	
	if strengthWrong > 0 {
		t.Error("Should not detect wrong pheromone type")
	}
}

func TestPheromoneTrailReinforcement(t *testing.T) {
	is := NewInsectSystem()
	
	// Create initial trail
	producer := NewEntity(1, []string{"pheromone_production"}, "test", Position{})
	producer.SetTrait("pheromone_production", 0.5)

	startPos := Position{X: 0, Y: 0}
	endPos := Position{X: 10, Y: 10}
	trail := is.CreatePheromoneTrail(producer, TrailPheromone, startPos, endPos)

	if trail == nil {
		t.Fatal("Failed to create trail")
	}

	originalStrength := trail.Strength[2] // Middle of trail

	// Create reinforcing entity near trail
	reinforcer := NewEntity(2, []string{"pheromone_production"}, "test", Position{X: 5, Y: 5})
	reinforcer.SetTrait("pheromone_production", 0.6)

	// Reinforce trail
	is.ReinforcePheromoneTrail(reinforcer, trail)

	// Trail strength should be increased
	if trail.Strength[2] <= originalStrength {
		t.Error("Trail strength should increase after reinforcement")
	}

	// Test entity with no pheromone production
	nonProducer := NewEntity(3, []string{"pheromone_production"}, "test", Position{X: 5, Y: 5})
	nonProducer.SetTrait("pheromone_production", 0.05) // Too low

	originalStrengthBeforeNonReinforcement := trail.Strength[2]
	is.ReinforcePheromoneTrail(nonProducer, trail)

	// Trail strength should not change
	if trail.Strength[2] != originalStrengthBeforeNonReinforcement {
		t.Error("Trail strength should not change when reinforced by non-producer")
	}
}