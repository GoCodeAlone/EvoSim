package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// EvaluationRule represents a rule for evaluating entity traits
type EvaluationRule struct {
	Name       string  `json:"name"`
	Expression string  `json:"expression"` // Simple expression like "strength + agility * 0.5"
	Weight     float64 `json:"weight"`     // Weight of this rule in overall fitness
	Target     float64 `json:"target"`     // Target value (for optimization problems)
	Minimize   bool    `json:"minimize"`   // Whether to minimize or maximize this rule
}

// EvaluationEngine provides dynamic evaluation of entities
type EvaluationEngine struct {
	Rules []EvaluationRule
}

// NewEvaluationEngine creates a new evaluation engine
func NewEvaluationEngine() *EvaluationEngine {
	return &EvaluationEngine{
		Rules: make([]EvaluationRule, 0),
	}
}

// AddRule adds an evaluation rule
func (e *EvaluationEngine) AddRule(name, expression string, weight, target float64, minimize bool) {
	rule := EvaluationRule{
		Name:       name,
		Expression: expression,
		Weight:     weight,
		Target:     target,
		Minimize:   minimize,
	}
	e.Rules = append(e.Rules, rule)
}

// Evaluate calculates the fitness of an entity based on all rules
func (e *EvaluationEngine) Evaluate(entity *Entity) float64 {
	totalFitness := 0.0

	for _, rule := range e.Rules {
		value, err := e.evaluateExpression(rule.Expression, entity)
		if err != nil {
			continue // Skip invalid expressions
		}

		// Calculate fitness based on target and minimize/maximize preference
		var ruleFitness float64
		if rule.Minimize {
			// For minimization, fitness is higher when value is closer to target (lower)
			distance := math.Abs(value - rule.Target)
			ruleFitness = 1.0 / (1.0 + distance)
		} else {
			// For maximization, fitness increases with value
			if rule.Target > 0 {
				ruleFitness = value / rule.Target
			} else {
				ruleFitness = value
			}
		}

		totalFitness += ruleFitness * rule.Weight
	}

	return totalFitness
}

// evaluateExpression evaluates a simple mathematical expression using entity traits
func (e *EvaluationEngine) evaluateExpression(expression string, entity *Entity) (float64, error) {
	// Simple expression parser for basic operations
	// Supports: +, -, *, /, parentheses, trait names, and numbers

	// Replace trait names with their values
	expr := expression
	for traitName, trait := range entity.Traits {
		expr = strings.ReplaceAll(expr, traitName, fmt.Sprintf("%.6f", trait.Value))
	}

	// Evaluate the mathematical expression
	return e.evaluateMathExpression(expr)
}

// evaluateMathExpression evaluates a mathematical expression with numbers and operators
func (e *EvaluationEngine) evaluateMathExpression(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "") // Remove spaces

	// Simple recursive descent parser for basic math expressions
	result, _, err := e.parseExpression(expr, 0)
	return result, err
}

// parseExpression parses addition and subtraction (lowest precedence)
func (e *EvaluationEngine) parseExpression(expr string, pos int) (float64, int, error) {
	left, newPos, err := e.parseTerm(expr, pos)
	if err != nil {
		return 0, pos, err
	}

	for newPos < len(expr) {
		op := expr[newPos]
		if op != '+' && op != '-' {
			break
		}

		right, nextPos, err := e.parseTerm(expr, newPos+1)
		if err != nil {
			return 0, newPos, err
		}

		if op == '+' {
			left += right
		} else {
			left -= right
		}
		newPos = nextPos
	}

	return left, newPos, nil
}

// parseTerm parses multiplication and division (higher precedence)
func (e *EvaluationEngine) parseTerm(expr string, pos int) (float64, int, error) {
	left, newPos, err := e.parseFactor(expr, pos)
	if err != nil {
		return 0, pos, err
	}

	for newPos < len(expr) {
		op := expr[newPos]
		if op != '*' && op != '/' {
			break
		}

		right, nextPos, err := e.parseFactor(expr, newPos+1)
		if err != nil {
			return 0, newPos, err
		}

		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, newPos, errors.New("division by zero")
			}
			left /= right
		}
		newPos = nextPos
	}

	return left, newPos, nil
}

// parseFactor parses numbers and parentheses (highest precedence)
func (e *EvaluationEngine) parseFactor(expr string, pos int) (float64, int, error) {
	if pos >= len(expr) {
		return 0, pos, errors.New("unexpected end of expression")
	}

	if expr[pos] == '(' {
		// Parse parenthesized expression
		result, newPos, err := e.parseExpression(expr, pos+1)
		if err != nil {
			return 0, pos, err
		}
		if newPos >= len(expr) || expr[newPos] != ')' {
			return 0, newPos, errors.New("missing closing parenthesis")
		}
		return result, newPos + 1, nil
	}

	// Parse number
	start := pos
	if pos < len(expr) && (expr[pos] == '-' || expr[pos] == '+') {
		pos++ // Handle sign
	}

	for pos < len(expr) && (expr[pos] >= '0' && expr[pos] <= '9' || expr[pos] == '.') {
		pos++
	}

	if start == pos {
		return 0, pos, errors.New("expected number")
	}

	value, err := strconv.ParseFloat(expr[start:pos], 64)
	if err != nil {
		return 0, pos, err
	}

	return value, pos, nil
}

// CreateFitnessFunction creates a fitness function using the evaluation engine
func (e *EvaluationEngine) CreateFitnessFunction() func(*Entity) float64 {
	return func(entity *Entity) float64 {
		return e.Evaluate(entity)
	}
}
