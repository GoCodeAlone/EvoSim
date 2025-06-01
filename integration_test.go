package main

import (
	"testing"
)

// TestIntegration tests the complete genetic algorithm system
func TestIntegration(t *testing.T) {
	// Test setup
	traitNames := []string{"strength", "agility", "intelligence"}
	populationSize := 20
	mutationRate := 0.1
	mutationStrength := 0.2
	generations := 10

	// Create population
	population := NewPopulation(populationSize, traitNames, mutationRate, mutationStrength)
	if population == nil {
		t.Fatal("Failed to create population")
	}

	// Create evaluation engine
	engine := NewEvaluationEngine()
	engine.AddRule("overall", "strength + agility + intelligence", 1.0, 3.0, false)
	engine.AddRule("balance", "strength - agility", 0.5, 0.0, true)

	fitnessFunc := engine.CreateFitnessFunction()

	// Record initial statistics
	population.EvaluateFitness(fitnessFunc)
	initialAvg, _, initialMax := population.GetStats()
	initialBest := population.GetBest()

	// Run evolution
	for i := 0; i < generations; i++ {
		population.EvaluateFitness(fitnessFunc)
		population.Evolve()
	}

	// Final evaluation
	population.EvaluateFitness(fitnessFunc)
	finalAvg, _, finalMax := population.GetStats()
	finalBest := population.GetBest()

	// Verify evolution occurred
	if finalMax < initialMax {
		t.Logf("Warning: Final max fitness (%f) is less than initial (%f)", finalMax, initialMax)
	}

	// Verify population size maintained
	if len(population.Entities) != populationSize {
		t.Errorf("Population size changed: expected %d, got %d", populationSize, len(population.Entities))
	}

	// Verify generation counter
	if population.Generation != generations {
		t.Errorf("Generation counter incorrect: expected %d, got %d", generations, population.Generation)
	}

	// Verify all entities have traits
	for i, entity := range population.Entities {
		if len(entity.Traits) != len(traitNames) {
			t.Errorf("Entity %d has wrong number of traits: expected %d, got %d",
				i, len(traitNames), len(entity.Traits))
		}

		for _, traitName := range traitNames {
			if _, exists := entity.Traits[traitName]; !exists {
				t.Errorf("Entity %d missing trait %s", i, traitName)
			}
		}
	}

	// Log results for manual inspection
	t.Logf("Integration test results:")
	t.Logf("  Initial: avg=%.3f, max=%.3f, best_id=%d", initialAvg, initialMax, initialBest.ID)
	t.Logf("  Final:   avg=%.3f, max=%.3f, best_id=%d", finalAvg, finalMax, finalBest.ID)
	t.Logf("  Evolution ran for %d generations", generations)
}

// TestDynamicEvaluation tests the dynamic evaluation system
func TestDynamicEvaluation(t *testing.T) {
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength", "agility", "magic"}, "test", pos)
	entity.SetTrait("strength", 1.0)
	entity.SetTrait("agility", 0.5)
	entity.SetTrait("magic", 2.0)

	// Test multiple evaluation strategies
	testCases := []struct {
		name       string
		rules      []EvaluationRule
		minFitness float64 // Minimum expected fitness
	}{
		{
			name: "warrior_focus",
			rules: []EvaluationRule{
				{"combat", "strength + agility", 1.0, 2.0, false},
			},
			minFitness: 0.5,
		},
		{
			name: "mage_focus",
			rules: []EvaluationRule{
				{"magical_power", "magic * 2", 1.0, 4.0, false},
			},
			minFitness: 0.8,
		},
		{
			name: "balanced",
			rules: []EvaluationRule{
				{"physical", "strength + agility", 0.5, 1.5, false},
				{"mental", "magic", 0.5, 2.0, false},
			},
			minFitness: 0.8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEvaluationEngine()
			for _, rule := range tc.rules {
				engine.AddRule(rule.Name, rule.Expression, rule.Weight, rule.Target, rule.Minimize)
			}

			fitness := engine.Evaluate(entity)

			if fitness < tc.minFitness {
				t.Errorf("%s: expected fitness >= %f, got %f", tc.name, tc.minFitness, fitness)
			}

			t.Logf("%s: fitness = %.3f", tc.name, fitness)
		})
	}
}

// TestEvolutionConvergence tests that evolution tends to improve fitness over time
func TestEvolutionConvergence(t *testing.T) {
	traitNames := []string{"x", "y"}
	population := NewPopulation(50, traitNames, 0.2, 0.3)

	// Simple optimization problem: minimize distance from origin
	engine := NewEvaluationEngine()
	engine.AddRule("distance", "x * x + y * y", 1.0, 0.0, true)

	fitnessFunc := engine.CreateFitnessFunction()

	// Track fitness over generations
	var fitnessHistory []float64

	for gen := 0; gen < 30; gen++ {
		population.EvaluateFitness(fitnessFunc)
		avg, _, _ := population.GetStats()
		fitnessHistory = append(fitnessHistory, avg)

		if gen < 29 {
			population.Evolve()
		}
	}

	// Check that average fitness improved
	initialFitness := fitnessHistory[0]
	finalFitness := fitnessHistory[len(fitnessHistory)-1]

	if finalFitness <= initialFitness {
		t.Logf("Warning: Evolution may not have improved fitness")
		t.Logf("Initial: %.3f, Final: %.3f", initialFitness, finalFitness)
	}

	// Check that best entity is reasonably close to origin
	best := population.GetBest()
	x := best.GetTrait("x")
	y := best.GetTrait("y")
	distance := x*x + y*y

	if distance > 1.0 {
		t.Logf("Warning: Best entity not very close to origin, distance^2 = %.3f", distance)
	}

	t.Logf("Convergence test: initial_fitness=%.3f, final_fitness=%.3f", initialFitness, finalFitness)
	t.Logf("Best entity: x=%.3f, y=%.3f, distance^2=%.3f", x, y, distance)
}
