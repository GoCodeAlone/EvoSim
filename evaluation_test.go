package main

import (
	"testing"
)

// newTestEntity creates a simple entity for testing without molecular/feedback systems
func newTestEntity(id int, traitNames []string, species string, position Position) *Entity {
	entity := &Entity{
		ID:         id,
		Traits:     make(map[string]Trait),
		Fitness:    0.0,
		Position:   position,
		Energy:     100.0,
		Age:        0,
		IsAlive:    true,
		Species:    species,
		Generation: 0,
	}

	// Initialize traits only (no molecular or feedback systems)
	for _, name := range traitNames {
		entity.Traits[name] = Trait{
			Name:  name,
			Value: 0.0, // Initialize to 0 for predictable test results
		}
	}

	return entity
}

func TestNewEvaluationEngine(t *testing.T) {
	engine := NewEvaluationEngine()

	if engine == nil {
		t.Error("Expected non-nil evaluation engine")
		return
	}

	if len(engine.Rules) != 0 {
		t.Errorf("Expected 0 initial rules, got %d", len(engine.Rules))
	}
}

func TestEvaluationEngineAddRule(t *testing.T) {
	engine := NewEvaluationEngine()

	engine.AddRule("test_rule", "strength + agility", 1.0, 2.0, false)

	if len(engine.Rules) != 1 {
		t.Errorf("Expected 1 rule after adding, got %d", len(engine.Rules))
	}

	rule := engine.Rules[0]
	if rule.Name != "test_rule" {
		t.Errorf("Expected rule name 'test_rule', got %s", rule.Name)
	}
	if rule.Expression != "strength + agility" {
		t.Errorf("Expected expression 'strength + agility', got %s", rule.Expression)
	}
	if rule.Weight != 1.0 {
		t.Errorf("Expected weight 1.0, got %f", rule.Weight)
	}
	if rule.Target != 2.0 {
		t.Errorf("Expected target 2.0, got %f", rule.Target)
	}
	if rule.Minimize != false {
		t.Errorf("Expected minimize false, got %t", rule.Minimize)
	}
}

func TestEvaluationEngineEvaluate(t *testing.T) {
	engine := NewEvaluationEngine()
	engine.AddRule("simple", "strength", 1.0, 1.0, false)

	pos := Position{X: 0, Y: 0}
	entity := newTestEntity(1, []string{"strength"}, "test", pos)
	entity.SetTrait("strength", 0.5)

	fitness := engine.Evaluate(entity)

	// With strength = 0.5, target = 1.0, maximize = false
	// fitness should be 0.5 / 1.0 = 0.5
	expected := 0.5
	if fitness != expected {
		t.Errorf("Expected fitness %f, got %f", expected, fitness)
	}
}

func TestEvaluationEngineEvaluateMinimize(t *testing.T) {
	engine := NewEvaluationEngine()
	engine.AddRule("minimize_test", "strength", 1.0, 0.0, true)
	pos := Position{X: 0, Y: 0}
	entity := newTestEntity(1, []string{"strength"}, "test", pos)
	entity.SetTrait("strength", 0.5)

	fitness := engine.Evaluate(entity)

	// With strength = 0.5, target = 0.0, minimize = true
	// distance = |0.5 - 0.0| = 0.5
	// fitness = 1.0 / (1.0 + 0.5) = 2/3 ≈ 0.667
	expected := 1.0 / (1.0 + 0.5)
	if fitness != expected {
		t.Errorf("Expected fitness %f, got %f", expected, fitness)
	}
}

func TestEvaluationEngineEvaluateExpression(t *testing.T) {
	engine := NewEvaluationEngine()
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength", "agility"}, "test", pos)
	entity.SetTrait("strength", 2.0)
	entity.SetTrait("agility", 3.0)

	tests := []struct {
		expression string
		expected   float64
	}{
		{"strength", 2.0},
		{"agility", 3.0},
		{"strength + agility", 5.0},
		{"strength * agility", 6.0},
		{"strength - agility", -1.0},
		{"agility / strength", 1.5},
		{"(strength + agility) * 2", 10.0},
	}

	for _, test := range tests {
		result, err := engine.evaluateExpression(test.expression, entity)
		if err != nil {
			t.Errorf("Error evaluating expression '%s': %v", test.expression, err)
			continue
		}

		if result != test.expected {
			t.Errorf("Expression '%s': expected %f, got %f", test.expression, test.expected, result)
		}
	}
}

func TestEvaluationEngineMathExpression(t *testing.T) {
	engine := NewEvaluationEngine()

	tests := []struct {
		expression string
		expected   float64
	}{
		{"5", 5.0},
		{"2 + 3", 5.0},
		{"10 - 4", 6.0},
		{"3 * 4", 12.0},
		{"15 / 3", 5.0},
		{"2 + 3 * 4", 14.0}, // Should follow order of operations
		{"(2 + 3) * 4", 20.0},
		{"-5", -5.0},
		{"3.5 + 2.5", 6.0},
	}

	for _, test := range tests {
		result, err := engine.evaluateMathExpression(test.expression)
		if err != nil {
			t.Errorf("Error evaluating math expression '%s': %v", test.expression, err)
			continue
		}

		if result != test.expected {
			t.Errorf("Math expression '%s': expected %f, got %f", test.expression, test.expected, result)
		}
	}
}

func TestEvaluationEngineCreateFitnessFunction(t *testing.T) {
	engine := NewEvaluationEngine()
	engine.AddRule("test", "strength * 2", 1.0, 2.0, false)

	fitnessFunc := engine.CreateFitnessFunction()
	pos := Position{X: 0, Y: 0}
	entity := newTestEntity(1, []string{"strength"}, "test", pos)
	entity.SetTrait("strength", 1.0)

	fitness := fitnessFunc(entity)

	// strength * 2 = 1.0 * 2 = 2.0
	// With target = 2.0, fitness = 2.0 / 2.0 = 1.0
	expected := 1.0
	if fitness != expected {
		t.Errorf("Expected fitness function result %f, got %f", expected, fitness)
	}
}

func TestEvaluationEngineComplexExpression(t *testing.T) {
	engine := NewEvaluationEngine()
	pos := Position{X: 0, Y: 0}
	entity := NewEntity(1, []string{"strength", "agility", "intelligence"}, "test", pos)
	entity.SetTrait("strength", 1.0)
	entity.SetTrait("agility", 2.0)
	entity.SetTrait("intelligence", 3.0)

	// Test complex expression: (strength + agility) * intelligence / 2
	result, err := engine.evaluateExpression("(strength + agility) * intelligence / 2", entity)
	if err != nil {
		t.Errorf("Error evaluating complex expression: %v", err)
	}

	expected := (1.0 + 2.0) * 3.0 / 2.0 // = 3 * 3 / 2 = 4.5
	if result != expected {
		t.Errorf("Complex expression: expected %f, got %f", expected, result)
	}
}
