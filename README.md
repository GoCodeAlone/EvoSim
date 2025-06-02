# EvoSim: Advanced Evolutionary Ecosystem Simulation

A sophisticated genetic algorithm simulation featuring a complete evolutionary ecosystem with entities, plants, tools, environmental modification, emergent behaviors, and real-time visualization.

## 🌟 Features

### Core Simulation
- **Dynamic Genetic System**: 15+ heritable traits with DNA/RNA representation and cellular evolution
- **Multi-Species Ecosystem**: Herbivores, predators, omnivores with complex interactions
- **Plant Life System**: 6 plant types with underground networks and wind-based pollen dispersal
- **Environmental Systems**: Day/night cycles, seasons, weather patterns, and geological events
- **Species Formation**: Automatic speciation through genetic distance and reproductive barriers

### Advanced Systems
- **🔧 Tool Creation**: 10 tool types with crafting, durability, and modification systems
- **🏗️ Environmental Modification**: 12 modification types including tunnels, traps, shelters, and farms
- **🧠 Emergent Behaviors**: 8 discoverable behaviors with social learning and natural emergence
- **🌐 Real-time Web Interface**: WebSocket-based web visualization with all 14 view modes
- **💾 State Persistence**: Complete save/load functionality with JSON serialization
- **🔬 Underground Networks**: Plant communication and resource sharing through mycorrhizal networks

### Visualization & Interface
- **CLI Interface**: 14 interactive view modes (Grid, Stats, Events, Populations, Communication, etc.)
- **Web Interface**: Modern, responsive web UI with real-time updates
- **Multi-zoom Viewport**: Navigate and explore the simulation world at different scales
- **Comprehensive Statistics**: Track evolution, behaviors, tools, and environmental changes

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/GoCodeAlone/EvoSim.git
cd EvoSim

# Run with CLI interface
GOWORK=off go run .

# Run with web interface
GOWORK=off go run . --web --web-port 8080
```

### Basic Commands

```bash
# Start simulation with custom parameters
GOWORK=off go run . --pop-size 50 --width 200 --height 200

# Save simulation state
GOWORK=off go run . --save my_simulation.json

# Load saved state
GOWORK=off go run . --load my_simulation.json

# Run web interface
GOWORK=off go run . --web

# Show all options
GOWORK=off go run . --help
```

## 🎮 Controls

### CLI Interface
- **Space**: Pause/Resume simulation
- **V**: Cycle through view modes
- **Arrow Keys**: Navigate viewport
- **+/-**: Zoom in/out
- **?**: Toggle help screen
- **Q**: Quit

### Web Interface
- Access via browser at `http://localhost:8080`
- Real-time simulation updates
- Interactive view switching
- Responsive design for all devices

## 🔬 Scientific Features

### Genetic Evolution
- **DNA/RNA System**: Complete nucleotide sequences with realistic inheritance
- **Cellular Complexity**: 8 cell types and 8 organelle types
- **Macro Evolution**: Species trees and phylogenetic tracking
- **Environmental Pressure**: Natural selection based on environmental conditions

### Ecosystem Dynamics
- **Resource Competition**: Limited food sources and territory
- **Predator-Prey Relationships**: Dynamic population balancing
- **Communication Systems**: 6 signal types for entity coordination
- **Seasonal Variation**: Changing environmental conditions affect survival

### Emergent Intelligence
- **Tool Discovery**: Entities naturally discover tool-making based on intelligence
- **Social Learning**: Cooperative entities learn behaviors from others
- **Environmental Adaptation**: Entities modify their environment for survival
- **Cultural Evolution**: Behaviors spread through populations over time

## 📊 View Modes

1. **Grid**: Main simulation visualization with entities, plants, and environment
2. **Stats**: Population statistics and trait distributions
3. **Events**: World events and significant occurrences
4. **Populations**: Detailed population analysis and demographics
5. **Communication**: Signal activity and communication patterns
6. **Civilization**: Tribal structures and technology development
7. **Physics**: Physics simulation state and forces
8. **Wind**: Wind patterns and pollen dispersal
9. **Species**: Species tracking and genetic distance analysis
10. **Network**: Underground plant network visualization
11. **DNA**: Genetic sequences and inheritance patterns
12. **Cellular**: Cell types and organelle development
13. **Evolution**: Macro evolution tracking and phylogenetic trees
14. **Topology**: World terrain and geological features

## 🧪 Testing

```bash
# Run all tests
GOWORK=off go test ./...

# Run with verbose output
GOWORK=off go test -v ./...

# Test web interface functionality
./test_web_interface.sh

# Run benchmarks
GOWORK=off go test -bench=.
```

## 🏗️ Architecture

### Core Systems
- **World**: Main simulation manager and update loop
- **Entities**: Individual organisms with genetic traits and behaviors
- **Plants**: Plant ecosystem with reproduction and networking
- **Tools**: Tool creation, usage, and modification system
- **Environment**: Environmental modifications and persistent structures
- **Communication**: Signal-based entity communication
- **Civilization**: Tribal organization and structure building

### Advanced Features
- **DNA System**: Genetic representation with chromosomes and alleles
- **Cellular System**: Cell specialization and organelle development
- **Behavior System**: Emergent behavior discovery and social learning
- **Network System**: Underground plant communication networks
- **Wind System**: Atmospheric simulation with pollen dispersal
- **Physics System**: Collision detection and environmental forces

## 📈 Examples of Emergent Behavior

- **Tool Making**: Intelligent entities discover stone tool creation when needing better equipment
- **Tunnel Networks**: Entities in dangerous areas learn to dig protective underground passages
- **Resource Caching**: Entities hide food supplies for later retrieval during scarcity
- **Trap Setting**: Aggressive entities learn to set traps near food sources
- **Cooperative Building**: Groups work together to create complex structures
- **Social Learning**: Successful behaviors spread through cooperative populations

## 🌍 Environmental Features

- **Biomes**: Grassland, forest, desert, mountain, lake, and river environments
- **Weather**: Storms, volcanic eruptions, earthquakes affecting evolution
- **Seasonal Cycles**: Spring/summer/autumn/winter with varying conditions
- **Plant Networks**: Underground fungal networks connecting compatible plants
- **Wind Dispersal**: Realistic pollen movement and cross-pollination
- **Geological Events**: Terrain changes affecting population distribution

## 🔧 Configuration

### Command Line Options
- `--width`, `--height`: World dimensions
- `--pop-size`: Initial population size per species
- `--seed`: Random seed for reproducible results
- `--web`: Enable web interface mode
- `--web-port`: Web server port (default: 8080)
- `--save`: Save simulation state to file
- `--load`: Load simulation state from file

### Advanced Configuration
Most simulation parameters can be adjusted in the source code:
- Mutation rates and genetic diversity
- Environmental conditions and biome distributions
- Tool creation requirements and effectiveness
- Behavior discovery rates and learning parameters
- Communication signal strengths and ranges

## 📝 Project Status

This simulation represents a comprehensive evolutionary ecosystem with:
- ✅ Complete genetic algorithm implementation
- ✅ Multi-species ecosystem with realistic interactions
- ✅ Tool creation and environmental modification systems
- ✅ Emergent behavior discovery and social learning
- ✅ Real-time web interface with WebSocket updates
- ✅ Complete state persistence functionality
- ✅ Extensive test coverage and validation

See [FEATURES.md](FEATURES.md) for detailed implementation status and roadmap.

## 🤝 Contributing

This project demonstrates advanced evolutionary simulation concepts. Feel free to:
- Experiment with parameters and configurations
- Add new tool types or environmental modifications
- Implement additional emergent behaviors
- Enhance the web interface with new visualizations
- Improve genetic algorithms and selection mechanisms

## 📄 License

This project is available for educational and research purposes. See the repository for license details.