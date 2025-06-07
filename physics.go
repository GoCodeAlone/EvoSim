package main

import (
	"fmt"
	"math"
)

// Vector2D represents a 2D vector for physics calculations
type Vector2D struct {
	X, Y float64
}

// Add adds two vectors
func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{X: v.X + other.X, Y: v.Y + other.Y}
}

// Multiply scales a vector by a scalar
func (v Vector2D) Multiply(scalar float64) Vector2D {
	return Vector2D{X: v.X * scalar, Y: v.Y * scalar}
}

// Magnitude returns the length of the vector
func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a unit vector in the same direction
func (v Vector2D) Normalize() Vector2D {
	mag := v.Magnitude()
	if mag == 0 {
		return Vector2D{0, 0}
	}
	return Vector2D{X: v.X / mag, Y: v.Y / mag}
}

// PhysicsComponent adds physics properties to entities
type PhysicsComponent struct {
	Velocity     Vector2D // Current velocity
	Acceleration Vector2D // Current acceleration
	Mass         float64  // Entity mass (affects inertia)
	Drag         float64  // Air/water resistance
	MaxVelocity  float64  // Terminal velocity
}

// NewPhysicsComponent creates physics component based on entity traits
func NewPhysicsComponent(entity *Entity) *PhysicsComponent {
	size := entity.GetTrait("size")

	mass := 1.0 + size*2.0 // Larger entities are heavier
	maxVel := 2.0 + entity.GetTrait("speed")*3.0

	return &PhysicsComponent{
		Velocity:     Vector2D{0, 0},
		Acceleration: Vector2D{0, 0},
		Mass:         mass,
		Drag:         0.1,
		MaxVelocity:  maxVel,
	}
}

// PhysicsSystem manages world physics
type PhysicsSystem struct {
	Gravity            Vector2D              // Global gravity force
	FluidDensity       float64               // Density of the medium (affects drag)
	ViscosityMap       map[BiomeType]float64 // Viscosity by biome type
	CollisionsThisTick int                   // Collisions detected this tick
	TotalCollisions    int                   // Total collisions over time
}

// NewPhysicsSystem creates a new physics system
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		Gravity:            Vector2D{X: 0, Y: 0.1}, // Slight downward gravity
		FluidDensity:       1.0,
		CollisionsThisTick: 0,
		TotalCollisions:    0,
		ViscosityMap: map[BiomeType]float64{
			BiomePlains:    0.1,  // Low resistance
			BiomeForest:    0.2,  // Trees create drag
			BiomeDesert:    0.15, // Sand creates some resistance
			BiomeMountain:  0.3,  // Rocky terrain is hard to move through
			BiomeWater:     0.8,  // Water has high viscosity
			BiomeRadiation: 0.25, // Thick atmosphere
		},
	}
}

// ApplyPhysics updates entity physics for one time step
func (ps *PhysicsSystem) ApplyPhysics(entity *Entity, physics *PhysicsComponent, biome BiomeType, deltaTime float64) {
	// Apply gravity
	gravityForce := ps.Gravity.Multiply(physics.Mass)
	physics.Acceleration = physics.Acceleration.Add(gravityForce)

	// Apply biome-specific viscosity/drag
	viscosity := ps.ViscosityMap[biome]
	dragForce := physics.Velocity.Multiply(-viscosity * physics.Drag)
	physics.Acceleration = physics.Acceleration.Add(dragForce.Multiply(1.0 / physics.Mass))

	// Update velocity with acceleration
	physics.Velocity = physics.Velocity.Add(physics.Acceleration.Multiply(deltaTime))

	// Apply velocity limits
	speed := physics.Velocity.Magnitude()
	if speed > physics.MaxVelocity {
		physics.Velocity = physics.Velocity.Normalize().Multiply(physics.MaxVelocity)
	}

	// Update position
	entity.Position.X += physics.Velocity.X * deltaTime
	entity.Position.Y += physics.Velocity.Y * deltaTime

	// Energy cost based on movement against resistance
	energyCost := speed * viscosity * 0.01
	entity.Energy -= energyCost

	// Reset acceleration for next frame
	physics.Acceleration = Vector2D{0, 0}
}

// ApplyForce adds a force to an entity's physics
func (ps *PhysicsSystem) ApplyForce(physics *PhysicsComponent, force Vector2D) {
	acceleration := force.Multiply(1.0 / physics.Mass)
	physics.Acceleration = physics.Acceleration.Add(acceleration)
}

// CalculateAttraction computes gravitational-like attraction between entities
func (ps *PhysicsSystem) CalculateAttraction(entity1, entity2 *Entity, physics1, physics2 *PhysicsComponent) Vector2D {
	dx := entity2.Position.X - entity1.Position.X
	dy := entity2.Position.Y - entity1.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < 0.1 {
		return Vector2D{0, 0} // Avoid division by zero
	}

	// Force magnitude based on mass and cooperation
	cooperation1 := entity1.GetTrait("cooperation")
	cooperation2 := entity2.GetTrait("cooperation")

	if entity1.Species == entity2.Species && cooperation1 > 0.3 && cooperation2 > 0.3 {
		// Attractive force between cooperative same-species entities
		forceMagnitude := (physics1.Mass * physics2.Mass) / (distance * distance) * 0.01
		direction := Vector2D{X: dx / distance, Y: dy / distance}
		return direction.Multiply(forceMagnitude)
	} else if entity1.Species != entity2.Species {
		// Repulsive force between different species (unless predator-prey)
		aggression1 := entity1.GetTrait("aggression")
		if aggression1 > 0.5 {
			// Predators are attracted to prey
			forceMagnitude := (physics1.Mass * physics2.Mass) / (distance * distance) * 0.005
			direction := Vector2D{X: dx / distance, Y: dy / distance}
			return direction.Multiply(forceMagnitude)
		} else {
			// Repulsion
			forceMagnitude := (physics1.Mass * physics2.Mass) / (distance * distance) * 0.002
			direction := Vector2D{X: -dx / distance, Y: -dy / distance}
			return direction.Multiply(forceMagnitude)
		}
	}

	return Vector2D{0, 0}
}

// FluidRegion represents areas with special fluid properties
type FluidRegion struct {
	Center    Position
	Radius    float64
	Density   float64  // How dense the fluid is
	Viscosity float64  // How resistant to movement
	Flow      Vector2D // Current/wind direction and strength
}

// ApplyFluidEffects applies fluid dynamics to entities in special regions
func (ps *PhysicsSystem) ApplyFluidEffects(entity *Entity, physics *PhysicsComponent, regions []FluidRegion) {
	for _, region := range regions {
		dx := entity.Position.X - region.Center.X
		dy := entity.Position.Y - region.Center.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= region.Radius {
			// Entity is in this fluid region

			// Apply flow/current
			flowForce := region.Flow.Multiply(region.Density)
			ps.ApplyForce(physics, flowForce)

			// Increase drag in dense fluids
			extraDrag := physics.Velocity.Multiply(-region.Viscosity * 0.5)
			ps.ApplyForce(physics, extraDrag)

			// Buoyancy effect - larger entities float better
			size := entity.GetTrait("size")
			buoyancy := Vector2D{X: 0, Y: -0.05 * size} // Upward force
			ps.ApplyForce(physics, buoyancy)
		}
	}
}

// CollisionSystem handles entity-entity and entity-environment collisions
type CollisionSystem struct {
	CollisionRadius float64
}

// NewCollisionSystem creates a new collision system
func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{
		CollisionRadius: 0.5,
	}
}

// CheckCollisions detects and resolves collisions between entities
func (cs *CollisionSystem) CheckCollisions(entities []*Entity, physicsComponents map[int]*PhysicsComponent, physicsSystem *PhysicsSystem, world *World) {
	for i, entity1 := range entities {
		if !entity1.IsAlive {
			continue
		}

		physics1 := physicsComponents[entity1.ID]
		if physics1 == nil {
			continue
		}

		for j, entity2 := range entities {
			if i >= j || !entity2.IsAlive {
				continue
			}

			physics2 := physicsComponents[entity2.ID]
			if physics2 == nil {
				continue
			}

			distance := entity1.DistanceTo(entity2)
			collisionDistance := cs.CollisionRadius*(1.0+entity1.GetTrait("size")) +
				cs.CollisionRadius*(1.0+entity2.GetTrait("size"))

			if distance < collisionDistance {
				// Emit collision event to central event bus
				if world != nil && world.CentralEventBus != nil {
					metadata := map[string]interface{}{
						"entity1_id":      entity1.ID,
						"entity2_id":      entity2.ID,
						"entity1_species": entity1.Species,
						"entity2_species": entity2.Species,
						"distance":        distance,
						"collision_force": collisionDistance - distance,
					}
					world.CentralEventBus.EmitSystemEvent(world.Tick, "collision", "physics", "collision_system",
						fmt.Sprintf("Collision between entity %d (%s) and entity %d (%s)", entity1.ID, entity1.Species, entity2.ID, entity2.Species),
						&entity1.Position, metadata)
				}

				// Collision detected - apply elastic collision physics
				cs.resolveCollision(entity1, entity2, physics1, physics2)
				// Track collision
				if physicsSystem != nil {
					physicsSystem.IncrementCollisionCount()
				}
			}
		}
	}
}

// resolveCollision handles the physics of entity collisions
func (cs *CollisionSystem) resolveCollision(entity1, entity2 *Entity, physics1, physics2 *PhysicsComponent) {
	// Calculate collision normal
	dx := entity2.Position.X - entity1.Position.X
	dy := entity2.Position.Y - entity1.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance == 0 {
		return // Entities are at same position
	}

	normal := Vector2D{X: dx / distance, Y: dy / distance}

	// Relative velocity
	relativeVel := Vector2D{
		X: physics2.Velocity.X - physics1.Velocity.X,
		Y: physics2.Velocity.Y - physics1.Velocity.Y,
	}

	// Velocity along collision normal
	velAlongNormal := relativeVel.X*normal.X + relativeVel.Y*normal.Y

	// Don't resolve if velocities are separating
	if velAlongNormal > 0 {
		return
	}

	// Restitution (bounciness) - depends on entity traits
	flexibility1 := entity1.GetTrait("endurance") * 0.5
	flexibility2 := entity2.GetTrait("endurance") * 0.5
	restitution := (flexibility1 + flexibility2) * 0.5

	// Calculate impulse scalar
	impulseScalar := -(1 + restitution) * velAlongNormal
	impulseScalar /= (1/physics1.Mass + 1/physics2.Mass)

	// Apply impulse
	impulse := normal.Multiply(impulseScalar)

	physics1.Velocity = physics1.Velocity.Add(impulse.Multiply(-1.0 / physics1.Mass))
	physics2.Velocity = physics2.Velocity.Add(impulse.Multiply(1.0 / physics2.Mass))

	// Energy loss from collision
	collisionEnergy := math.Abs(impulseScalar) * 0.01
	entity1.Energy -= collisionEnergy
	entity2.Energy -= collisionEnergy

	// Separate entities to prevent overlap
	separation := normal.Multiply(0.1)
	entity1.Position.X -= separation.X
	entity1.Position.Y -= separation.Y
	entity2.Position.X += separation.X
	entity2.Position.Y += separation.Y
}

// ResetCollisionCounters resets collision tracking for new tick
func (ps *PhysicsSystem) ResetCollisionCounters() {
	ps.CollisionsThisTick = 0
}

// IncrementCollisionCount tracks a collision
func (ps *PhysicsSystem) IncrementCollisionCount() {
	ps.CollisionsThisTick++
	ps.TotalCollisions++
}
