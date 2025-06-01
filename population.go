package main

import (
	"fmt"
	"math/rand"
	"sort"
)

// Population represents a collection of entities that can evolve
type Population struct {
	Entities         []*Entity
	Generation       int
	MutationRate     float64
	MutationStrength float64
	EliteSize        int
	TournamentSize   int
	TraitNames       []string
	Species          string
}

// NewPopulation creates a new population with the specified parameters
func NewPopulation(size int, traitNames []string, mutationRate, mutationStrength float64) *Population {
	pop := &Population{
		Entities:         make([]*Entity, size),
		Generation:       0,
		MutationRate:     mutationRate,
		MutationStrength: mutationStrength,
		EliteSize:        size / 10, // Top 10% are elite
		TournamentSize:   3,
		TraitNames:       traitNames,
		Species:          "default", // Default species name
	}

	// Initialize random entities
	for i := 0; i < size; i++ {
		pos := Position{X: rand.Float64() * 100, Y: rand.Float64() * 100}
		pop.Entities[i] = NewEntity(i, traitNames, pop.Species, pos)
	}

	return pop
}

// EvaluateFitness evaluates all entities using a dynamic fitness function
func (p *Population) EvaluateFitness(fitnessFunc func(*Entity) float64) {
	for _, entity := range p.Entities {
		entity.Fitness = fitnessFunc(entity)
	}
}

// SortByFitness sorts entities by fitness in descending order (highest first)
func (p *Population) SortByFitness() {
	sort.Slice(p.Entities, func(i, j int) bool {
		return p.Entities[i].Fitness > p.Entities[j].Fitness
	})
}

// GetBest returns the entity with the highest fitness
func (p *Population) GetBest() *Entity {
	if len(p.Entities) == 0 {
		return nil
	}

	best := p.Entities[0]
	for _, entity := range p.Entities[1:] {
		if entity.Fitness > best.Fitness {
			best = entity
		}
	}
	return best
}

// GetStats returns population statistics
func (p *Population) GetStats() (float64, float64, float64) {
	if len(p.Entities) == 0 {
		return 0, 0, 0
	}

	sum := 0.0
	min := p.Entities[0].Fitness
	max := p.Entities[0].Fitness

	for _, entity := range p.Entities {
		sum += entity.Fitness
		if entity.Fitness < min {
			min = entity.Fitness
		}
		if entity.Fitness > max {
			max = entity.Fitness
		}
	}

	return sum / float64(len(p.Entities)), min, max
}

// TournamentSelection selects an entity using tournament selection
func (p *Population) TournamentSelection() *Entity {
	best := p.Entities[rand.Intn(len(p.Entities))]

	for i := 1; i < p.TournamentSize; i++ {
		candidate := p.Entities[rand.Intn(len(p.Entities))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}

	return best
}

// Evolve performs one generation of evolution
func (p *Population) Evolve() {
	// Sort by fitness
	p.SortByFitness()

	// Create new generation
	newGeneration := make([]*Entity, len(p.Entities))
	nextID := 0

	// Keep elite individuals
	for i := 0; i < p.EliteSize && i < len(p.Entities); i++ {
		newGeneration[i] = p.Entities[i].Clone()
		newGeneration[i].ID = nextID
		nextID++
	}
	// Fill the rest through crossover and mutation
	for i := p.EliteSize; i < len(p.Entities); i++ {
		parent1 := p.TournamentSelection()
		parent2 := p.TournamentSelection()

		child := Crossover(parent1, parent2, nextID, p.Species)
		child.Mutate(p.MutationRate, p.MutationStrength)

		newGeneration[i] = child
		nextID++
	}

	p.Entities = newGeneration
	p.Generation++
}

// String returns a string representation of the population
func (p *Population) String() string {
	avg, min, max := p.GetStats()
	return fmt.Sprintf("Generation %d: Size=%d, Avg=%.3f, Min=%.3f, Max=%.3f",
		p.Generation, len(p.Entities), avg, min, max)
}
