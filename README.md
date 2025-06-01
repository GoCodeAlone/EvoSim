# Dynamic Genetic Algorithm in Go

A sophisticated genetic algorithm implementation that supports dynamic trait-based entities with configurable evaluation rules.

## Features

- **Dynamic Entities**: Entities store data as traits that determine their abilities and functionality
- **Flexible Evaluation**: Dynamic evaluation system that doesn't require hardcoded functions
- **Genetic Operations**: Mutation, recombination (crossover), and evolution
- **Population Management**: Tournament selection, elitism, and generation tracking
- **Expression Parsing**: Mathematical expression evaluation for complex fitness functions
- **Comprehensive Testing**: Full test coverage with integration tests

## Architecture

### Core Components

1. **Entity** (`entity.go`): Represents individuals with dynamic traits
2. **Population** (`population.go`): Manages collections of entities and evolution
3. **EvaluationEngine** (`evaluation.go`): Dynamic fitness evaluation system
4. **Main** (`main.go`): Demonstration application

### Key Concepts

- **Traits**: Named attributes with floating-point values that define entity capabilities
- **Fitness Functions**: Dynamically created based on mathematical expressions using trait names
- **Evolution**: Population-based improvement through selection, crossover, and mutation

## Usage

### Running the Application

Due to a `go.work` file in the parent directory, use one of these methods:

```bash
# Option 1: Use the provided script
./run.sh

# Option 2: Use GOWORK=off directly
GOWORK=off go run .

# Option 3: Standard Go run (if no workspace conflicts)
go run .
```

### Running Tests

```bash
# Option 1: Use the provided script
./test.sh

# Option 2: Use GOWORK=off directly
GOWORK=off go test -v

# Option 3: Standard Go test (if no workspace conflicts)
go test -v
```

## Example Usage

```go
// Create entities with dynamic traits
traitNames := []string{"strength", "agility", "intelligence"}
population := NewPopulation(100, traitNames, 0.1, 0.2)

// Create dynamic evaluation rules
engine := NewEvaluationEngine()
engine.AddRule("combat", "strength + agility", 1.0, 2.0, false)
engine.AddRule("magic", "intelligence * 2", 0.8, 4.0, false)

// Create fitness function
fitnessFunc := engine.CreateFitnessFunction()

// Evolve population
for generation := 0; generation < 50; generation++ {
    population.EvaluateFitness(fitnessFunc)
    population.Evolve()
}

// Get results
best := population.GetBest()
fmt.Printf("Best entity: %s\n", best.String())
```

## Configuration

### Population Parameters
- **Size**: Number of entities in the population
- **Mutation Rate**: Probability of trait mutation (0.0-1.0)
- **Mutation Strength**: Standard deviation of mutation noise
- **Elite Size**: Number of top entities preserved each generation
- **Tournament Size**: Number of entities competing in selection

### Evaluation Rules
- **Expression**: Mathematical formula using trait names
- **Weight**: Importance of this rule in overall fitness
- **Target**: Optimal value for the expression
- **Minimize**: Whether to minimize or maximize the expression

## Testing

The project includes comprehensive tests:

- **Unit Tests**: Individual component testing
- **Integration Tests**: Full system testing
- **Convergence Tests**: Evolution effectiveness validation

Run all tests with: `./test.sh` or `GOWORK=off go test -v`

## Project Structure

```
mutate/
├── entity.go           # Entity definition and genetic operations
├── population.go       # Population management and evolution
├── evaluation.go       # Dynamic evaluation engine
├── main.go            # Demo application
├── entity_test.go     # Entity unit tests
├── population_test.go # Population unit tests
├── evaluation_test.go # Evaluation unit tests
├── integration_test.go # Integration tests
├── run.sh             # Run script (isolated from go.work)
├── test.sh            # Test script (isolated from go.work)
├── go.mod             # Go module definition
└── README.md          # This file
```

## Key Features Demonstrated

1. **Dynamic Trait System**: Entities can have any number of named traits
2. **Expression-Based Fitness**: Fitness functions defined as mathematical expressions
3. **Genetic Algorithm**: Complete implementation with selection, crossover, and mutation
4. **Flexible Architecture**: Easy to extend with new traits and evaluation rules
5. **Real-time Evolution**: Observable improvement over generations

The system successfully demonstrates evolution in action, typically showing significant fitness improvements over 50 generations while maintaining genetic diversity through mutation and crossover operations.
