# Copilot Instructions for Genetic Algorithm Simulation Project

## Project Overview
This is a sophisticated genetic algorithm simulation written in Go that models evolving entities in a 2D world environment. The project features advanced systems including communication, civilization building, physics, time cycles, and a rich terminal-based user interface.

## Language & Technology Stack
- **Language**: Go 1.24.3
- **Build System**: Standard Go modules with `GOWORK=off` (due to parent directory go.work conflicts)
- **UI Framework**: Charm libraries (Bubble Tea for TUI, Lipgloss for styling)
- **Testing**: Standard Go testing with comprehensive coverage
- **Architecture**: Multi-system simulation with real-time visualization

## Project Structure & Key Files

### Core Simulation Files
- `main.go` - Entry point, CLI argument parsing, simulation initialization
- `world.go` - Main simulation loop, world state management, entity coordination
- `entity.go` - Core entity definition with 15+ genetic traits (vision, speed, aggression, etc.)
- `population.go` - Population management, reproduction, mutation, selection algorithms
- `evaluation.go` - Dynamic fitness evaluation system with environmental adaptation

### Advanced Systems
- `communication.go` - Signal-based communication with 6 signal types (alert, food, mating, territory, distress, cooperation)
- `civilization.go` - Tribal systems with 8 structure types (nests, food storage, watchtowers, etc.)
- `physics.go` - Physics engine with collision detection, environmental forces, momentum
- `time_cycles.go` - Day/night cycles and seasonal changes affecting entity behavior
- `plant.go` - Plant ecosystem with 6 types and biome interactions
- `wind.go` - Wind system with pollen dispersal and cross-pollination mechanics
- `eventlog.go` - World event logging and history tracking

### User Interface
- `cli.go` - Enhanced CLI with 8 view modes and real-time updates (Grid, Stats, Events, Populations, Communication, Civilization, Physics, Wind)
- `viewport.go` - Multi-zoom navigation system (1x, 2x, 4x, 8x zoom levels)

### Project Planning & Tracking
- `FEATURES.md` - **Comprehensive feature roadmap and implementation status tracking** - ALWAYS CHECK THIS FILE FOR CURRENT PRIORITIES AND STATUS
- `README.md` - Project documentation and setup instructions

**IMPORTANT**: The `FEATURES.md` file contains the complete development roadmap including:
- ✅ Completed features with detailed descriptions
- 🚧 Features currently in progress
- 📋 Prioritized feature backlog (HIGH/MEDIUM/LOW priority)
- 🎯 Next steps and immediate tasks
- 📊 Testing and validation requirements
Always consult FEATURES.md before starting new work to understand current priorities and avoid duplicating effort.

### Testing & Utilities
- `*_test.go` - Comprehensive test suite covering all major systems
- `integration_test.go` - End-to-end simulation testing
- `scripts/` - Build and utility scripts

## Core Concepts & Systems

### Genetic Algorithm
- **Entity Traits**: 15+ heritable traits including vision, speed, aggression, cooperation, intelligence
- **Mutation System**: Gaussian mutation with configurable rates and bounds
- **Selection**: Fitness-based reproduction with environmental pressure adaptation
- **Population Dynamics**: Birth/death cycles, carrying capacity, resource competition

### Communication System
```go
type SignalType int
const (
    AlertSignal SignalType = iota    // Danger warnings
    FoodSignal                       // Food source locations
    MatingSignal                     // Reproductive availability
    TerritorySignal                  // Territory claims
    DistressSignal                   // Help requests
    CooperationSignal               // Collaboration invitations
)
```

### Civilization Features
- **Tribal Organization**: Entities form tribes with shared territories
- **Structure Building**: 8 structure types (Nest, FoodStorage, Watchtower, Barrier, Workshop, Temple, TradePost, Academy)
- **Resource Management**: Collective resource gathering and sharing
- **Technology**: Knowledge accumulation and technological advancement

### Wind & Pollen System (NEW)
```go
type WindSystem struct {
    BaseWindDirection   float64     // Global wind direction (radians)
    BaseWindStrength    float64     // Global wind strength
    TurbulenceLevel     float64     // Chaos in wind patterns
    WindMap            [][]WindVector // 2D grid of wind vectors
    AllPollenGrains    []*PollenGrain // Active pollen particles
    PollenClouds       []*PollenCloud // Pollen concentration areas
    SeasonalMultiplier  float64     // Wind strength varies by season
    WeatherPattern      int         // Current weather (0=calm, 1=windy, 2=storm)
}
```

### Physics Engine
- **Collision Detection**: AABB and circle-based collision systems
- **Environmental Forces**: Wind, terrain effects, gravitational influences
- **Movement Physics**: Momentum, acceleration, realistic movement patterns
- **Pollen Physics**: Wind-driven pollen grain movement and dispersal

### Time & Environment
- **Day/Night Cycles**: Affecting visibility, activity patterns, energy consumption
- **Seasonal Changes**: Impacting resource availability and entity behavior
- **Weather Systems**: Environmental conditions affecting simulation dynamics

## CLI Interface & Controls

### View Modes (8 total)
1. **Grid View** - 2D world visualization with entities and environment
2. **Stats View** - Population statistics, trait distributions, averages
3. **Events View** - Recent world events and significant occurrences
4. **Populations View** - Detailed population analysis and demographics
5. **Communication View** - Signal activity and communication patterns
6. **Civilization View** - Tribal information and structure status
7. **Physics View** - Physics state, forces, and collision information
8. **Wind View** - Wind patterns, pollen dispersal, and cross-pollination activity

### Navigation Controls
- Arrow keys: Move viewport
- +/- or mouse wheel: Zoom in/out (4 zoom levels)
- Tab: Cycle through view modes
- Space: Pause/resume simulation
- R: Reset simulation
- Q: Quit application

### Viewport System
- **Multi-zoom support**: 1x, 2x, 4x, 8x magnification levels
- **Smooth navigation**: Real-time viewport movement
- **Adaptive rendering**: Efficient display based on zoom level

## Development Patterns & Conventions

### Code Organization
- **Single responsibility**: Each file handles one major system
- **Interface-driven**: Use interfaces for testability and modularity
- **Error handling**: Comprehensive error checking with descriptive messages
- **Configuration**: Centralized constants and configurable parameters

### Testing Strategy
- **Unit tests**: Individual component testing for all major systems
- **Integration tests**: End-to-end simulation validation
- **Benchmark tests**: Performance testing for critical paths
- **Test coverage**: Aim for >80% coverage on core simulation logic

### Performance Considerations
- **Spatial partitioning**: Efficient collision detection and entity queries
- **Update optimization**: Selective updates based on entity state changes
- **Memory management**: Efficient entity pooling and cleanup
- **Rendering optimization**: Viewport-based culling and level-of-detail

## Common Development Tasks

### Adding New Entity Traits
1. Add trait field to `Entity` struct in `entity.go`
2. Update mutation logic in `Mutate()` method
3. Modify fitness evaluation in `evaluation.go`
4. Add trait to statistics display in `cli.go`
5. Update tests in `entity_test.go`

### Implementing New Communication Signals
1. Add signal type to `SignalType` enum in `communication.go`
2. Implement signal processing logic
3. Update signal visualization in CLI
4. Add signal-specific entity behaviors
5. Test signal propagation and response

### Creating New Civilization Structures
1. Add structure type to `StructureType` enum in `civilization.go`
2. Define structure properties and requirements
3. Implement construction and maintenance logic
4. Add structure effects on entities
5. Update civilization view display

### Extending Physics System
1. Modify physics constants and forces in `physics.go`
2. Update collision detection algorithms
3. Implement new environmental effects
4. Test physics interactions thoroughly
5. Verify performance impact

## Build & Run Instructions

### Standard Build
```bash
GOWORK=off go build -o mutate
./mutate
```

### Development Build
```bash
GOWORK=off go run . [flags]
```

### Common Flags
- `-pop <int>`: Set initial population size
- `-mut <float>`: Set mutation rate
- `-seed <int>`: Set random seed for reproducibility

### Testing
```bash
GOWORK=off go test ./...           # Run all tests
GOWORK=off go test -v ./...        # Verbose test output
GOWORK=off go test -bench=.        # Run benchmarks
```

## Key Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `golang.org/x/term` - Terminal utilities
- Standard library: `math/rand`, `encoding/json`, `time`, etc.

## Important Notes
- **GOWORK Environment**: Always use `GOWORK=off` due to parent directory conflicts
- **Random Seeding**: Use consistent seeds for reproducible simulations
- **Performance Monitoring**: Watch for memory leaks during long simulations
- **Cross-platform**: Code is designed to work on Windows, macOS, and Linux
- **Terminal Compatibility**: Requires terminal with ANSI color support

## Debug & Troubleshooting

### Common Issues
- **Build failures**: Check GOWORK=off environment setting
- **Performance degradation**: Monitor entity count and collision detection
- **UI rendering issues**: Verify terminal size and color support
- **Random behavior**: Ensure proper seed management for reproducibility

### Debug Features
- Event logging system for tracking world state changes
- Statistics view for monitoring simulation health
- Physics view for debugging movement and collision issues
- Population view for analyzing genetic drift and selection pressure

## Future Development Areas
- Multi-threading for large populations
- Network simulation capabilities
- Advanced AI behaviors and learning
- Enhanced visualization options
- Persistence and simulation replay
- Genetic programming extensions
