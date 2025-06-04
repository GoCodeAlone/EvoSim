# EvoSim Systems Documentation

## Overview

EvoSim is a sophisticated evolutionary ecosystem simulation that models complex biological, social, and environmental systems. This document provides comprehensive documentation of all systems, their interactions, and when/how they are triggered.

## Table of Contents

1. [Core Systems](#core-systems)
2. [Advanced AI Systems](#advanced-ai-systems) 
3. [Ecological Systems](#ecological-systems)
4. [Social and Behavioral Systems](#social-and-behavioral-systems)
5. [Environmental Systems](#environmental-systems)
6. [User Interface Systems](#user-interface-systems)
7. [System Interactions and Triggers](#system-interactions-and-triggers)

---

## Core Systems

### 1. World Management System (`world.go`)

**Purpose**: Main simulation controller that orchestrates all other systems.

**Key Components**:
- **Main Update Loop**: Runs every tick (100ms by default)
- **Entity Management**: Tracks all living entities and their lifecycle
- **Grid System**: Manages spatial relationships in 2D world
- **Tick Counter**: Global time tracking for all systems

**Trigger Conditions**:
- Starts immediately upon simulation launch
- Updates all systems in a specific order each tick
- Handles entity birth/death lifecycle events
- Manages world state persistence (save/load)

**Critical Functions**:
- `Update()`: Master update function that calls all subsystem updates
- `AddEntity()`: Creates new entities and assigns them spatial positions
- `RemoveEntity()`: Handles entity death and cleanup
- `GetNeighbors()`: Spatial queries for entity interactions

---

### 2. Entity System (`entity.go`)

**Purpose**: Individual organism behavior, genetics, and lifecycle management.

**Key Components**:
- **Genetic Traits**: 15+ heritable traits (speed, intelligence, aggression, etc.)
- **DNA System**: Complete nucleotide sequences with realistic inheritance  
- **Cellular Structure**: 8 specialized cell types and 8 organelle types
- **Lifecycle Management**: Birth, aging, reproduction, death

**Genetic Traits**:
1. **speed** - Movement rate and agility
2. **vision** - Environmental awareness radius
3. **intelligence** - Learning capacity and decision-making
4. **aggression** - Combat tendency and dominance
5. **cooperation** - Social behavior and group cohesion
6. **energy_efficiency** - Metabolic optimization
7. **size** - Physical size affecting many interactions
8. **defense** - Natural armor and protection
9. **fertility** - Reproductive success rate
10. **lifespan** - Maximum age before natural death
11. **adaptability** - Mutation resistance and flexibility
12. **endurance** - Sustained activity capacity
13. **stealth** - Predator avoidance and hunting
14. **territorial** - Space claiming and defense
15. **curiosity** - Exploration and discovery drive

**Additional Environmental Traits**:
- **aquatic_adaptation** - Swimming efficiency in water biomes
- **digging_ability** - Underground movement and tunnel creation
- **underground_nav** - Navigation in soil biomes
- **flying_ability** - Aerial movement capabilities
- **altitude_tolerance** - High-altitude environment adaptation

**Trigger Conditions**:
- Entity creation: When population reproduction occurs
- Trait updates: Every tick for living entities
- Aging: Gradual age increment each tick
- Death: When energy reaches 0 or age exceeds lifespan
- Reproduction: When conditions are met (season, energy, compatibility)

---

### 3. Population Management System (`population.go`)

**Purpose**: Manages species populations, reproduction, and evolutionary pressure.

**Key Components**:
- **Species Tracking**: Maintains distinct species based on genetic distance
- **Reproduction Logic**: Handles mating, genetic crossover, mutations
- **Selection Pressure**: Fitness-based survival and reproduction success
- **Population Limits**: Carrying capacity enforcement

**Reproduction Modes**:
1. **DirectCoupling** - Traditional mating between two entities
2. **EggLaying** - Oviparous reproduction with incubation periods
3. **LiveBirth** - Viviparous reproduction with gestation
4. **Budding** - Asexual reproduction creating genetic clones
5. **Fission** - Cellular division for simple organisms

**Mating Strategies**:
1. **Monogamous** - Single long-term partner
2. **Polygamous** - Multiple simultaneous partners  
3. **Sequential** - Series of short-term partnerships
4. **Promiscuous** - Opportunistic mating

**Trigger Conditions**:
- Reproduction checks: Every 50 ticks for fertile entities
- Species formation: When genetic distance exceeds threshold (0.8)
- Population limits: When entity count exceeds carrying capacity
- Mutation events: During reproduction with configurable rates

---

### 4. Evaluation System (`evaluation.go`)

**Purpose**: Calculates fitness scores that drive natural selection.

**Key Components**:
- **Survival Metrics**: Energy management, predator avoidance, longevity
- **Reproductive Success**: Mating frequency and offspring survival
- **Environmental Adaptation**: Biome-specific performance measures
- **Social Fitness**: Cooperation benefits and group survival

**Fitness Calculation Formula**:
```
Total Fitness = Base Survival × Reproductive Success × Environmental Adaptation × Social Bonus × Molecular Nutrition
```

**Trigger Conditions**:
- Fitness evaluation: Every 20 ticks for all living entities
- Selection pressure: During reproduction events
- Environmental stress: When environmental events occur
- Comparative ranking: For population management decisions

---

## Advanced AI Systems

### 5. Neural Networks System (`neural_networks.go`)

**Purpose**: Provides entities with learning capabilities and adaptive decision-making.

#### Neural Network Architecture

**Network Types**:
1. **FeedForward** - Basic forward propagation networks
2. **Recurrent** - Networks with memory capabilities
3. **Convolutional** - Pattern recognition networks
4. **Reinforcement** - Trial-and-error learning networks

**Network Structure**:
- **Input Layer**: 5 neurons processing environmental data
  - Vision/Environmental awareness (0-1)
  - Energy level (0-1) 
  - Threat detection (0-1)
  - Food availability (0-1)
  - Social interaction level (0-1)
- **Hidden Layer**: Variable size based on entity intelligence
- **Output Layer**: 3 neurons controlling behavior
  - moveX: Horizontal movement direction (-1 to 1)
  - moveY: Vertical movement direction (-1 to 1)
  - actionIntensity: Behavior intensity (0-1)

#### Learning Process

**When Neural Networks Are Created**:
- Triggered when an entity has intelligence trait > 0.3
- Created automatically during first `processNeuralDecisions()` call
- Network complexity scales with entity intelligence

**What Entities "Learn"**:
1. **Movement Optimization**: Efficient pathfinding and navigation
2. **Energy Management**: When to rest vs. when to be active
3. **Threat Response**: Predator avoidance patterns
4. **Food Seeking**: Successful foraging strategies
5. **Social Behavior**: Cooperation vs. competition decisions

**Why Entities Disappear from Neural View**:
1. **Entity Death**: Network is deleted when entity dies (`cleanupDeadEntities()`)
2. **Intelligence Drop**: If intelligence trait falls below 0.3 threshold
3. **Learning Completion**: Networks don't "finish" learning - they continuously adapt
4. **Experience Decay**: Old unused networks may become inactive

**What Happens to Learned Information**:
1. **Network Weights**: Synaptic connections strengthen with successful actions
2. **Experience Points**: Accumulated through reward/punishment feedback
3. **Behavior Patterns**: Successful strategies stored in memory
4. **Inheritance**: Neural capabilities can be passed to offspring through intelligence trait

**Learning Feedback Loop**:
```
Environmental Input → Neural Processing → Action → Outcome → Feedback → Weight Adjustment
```

**Feedback Calculation**:
- **Positive Reward**: Successful food finding, threat avoidance, energy gain
- **Negative Punishment**: Energy loss, danger exposure, failed actions
- **Reward Range**: -1.0 (failure) to +1.0 (success)

**Trigger Conditions**:
- Network creation: When intelligent entity (intelligence > 0.3) first needs decision
- Decision processing: Every tick for entities with neural networks
- Learning updates: After each action outcome is evaluated
- Experience decay: Every 100 ticks to prevent memory overflow
- Network cleanup: When entities die or lose intelligence

#### Performance Metrics

**Individual Network Tracking**:
- **Total Decisions**: Number of actions taken
- **Correct Decisions**: Successful actions that improved fitness
- **Success Rate**: Percentage of beneficial decisions
- **Experience**: Accumulated learning from all actions
- **Network Complexity**: Number of neurons and connections
- **Response Time**: Speed of decision making

**System-Wide Statistics**:
- **Total Networks**: All active neural networks
- **Active Networks**: Networks that made decisions recently
- **Learning Events**: Total number of learning updates
- **Emergent Behaviors**: Novel strategies discovered
- **Collective Intelligence**: Shared knowledge between entities

---

### 6. Emergent Behavior System (`emergent_behavior.go`)

**Purpose**: Allows entities to discover and learn new behaviors through environmental pressure.

**Discoverable Behaviors**:
1. **Tool Making** - Creating and using tools for enhanced survival
2. **Cache Building** - Storing food for future use
3. **Trap Setting** - Creating traps to catch prey
4. **Tunnel Building** - Digging underground passages
5. **Group Hunting** - Coordinated predator behavior
6. **Nest Construction** - Building protective shelters
7. **Resource Hoarding** - Systematic resource accumulation
8. **Social Learning** - Teaching behaviors to others

**Discovery Mechanism**:
- Intelligence-based discovery rates
- Environmental pressure triggers
- Curiosity trait influences exploration
- Social learning spreads behaviors through populations

**Trigger Conditions**:
- Discovery checks: Every 100 ticks for intelligent entities
- Learning spread: When cooperative entities interact
- Behavior selection: Based on current environmental needs
- Proficiency improvement: Through repeated practice

---

## Ecological Systems

### 7. Plant Ecosystem (`plant.go`, `plant_network.go`)

**Purpose**: Manages plant life, growth, reproduction, and inter-plant communication.

**Plant Types**:
1. **Grass** - Fast-growing, low-nutrition ground cover
2. **Bush** - Medium-growth shrubs with moderate nutrition
3. **Tree** - Slow-growing, high-nutrition giants
4. **Mushroom** - Decomposer organisms that recycle nutrients
5. **Algae** - Aquatic plants in water biomes
6. **Cactus** - Desert-adapted succulents

**Underground Network System**:
- **Mycorrhizal Networks**: Fungal connections between compatible plants
- **Resource Sharing**: Nutrients and energy distribution
- **Chemical Communication**: 6 signal types for plant coordination
- **Network Formation**: Proximity and compatibility-based connections

**Plant Reproduction**:
- **Wind Dispersal**: Pollen and seed movement via wind system
- **Animal Dispersal**: Seeds carried by entities
- **Cross-Pollination**: Genetic mixing between compatible species
- **Speciation**: Automatic species formation when genetic distance exceeds threshold

**Trigger Conditions**:
- Growth: Every tick based on environmental conditions
- Reproduction: Seasonal triggers (spring/summer peak)
- Network formation: When compatible plants are within connection range
- Death: From environmental stress, age, or consumption

---

### 8. Wind and Dispersal System (`wind.go`, `seed_dispersal.go`)

**Purpose**: Atmospheric simulation affecting plant reproduction and environmental dynamics.

**Wind Mechanics**:
- **Global Wind Patterns**: Base direction and strength
- **Turbulence**: Chaotic variations in wind flow
- **Seasonal Effects**: Wind strength varies by season
- **Weather Integration**: Storm systems affect wind patterns

**Dispersal Types**:
1. **Wind Dispersal** - Pollen and seeds carried by air currents
2. **Animal Dispersal** - Seeds picked up and dropped by entities
3. **Water Dispersal** - Seeds transported by water flow
4. **Explosive Dispersal** - Seeds forcefully ejected from plants

**Pollen System**:
- **Pollen Grains**: Individual particles with genetic information
- **Pollen Clouds**: Concentrated areas of high pollen density
- **Cross-Pollination**: Genetic mixing between compatible plant species
- **Wind-Driven Movement**: Realistic physics-based dispersal

**Trigger Conditions**:
- Wind updates: Every tick for atmospheric simulation
- Pollen release: During plant reproduction events
- Seed dispersal: When plants reach reproductive maturity
- Weather effects: Storm systems modify dispersal patterns

---

### 9. Molecular Evolution System (`molecular.go`)

**Purpose**: Manages nutritional needs, molecular composition, and metabolic processes.

**Molecule Types** (20+ types):
- **Proteins**: Essential for growth and repair
- **Amino Acids**: Building blocks of proteins
- **Lipids**: Energy storage and cell membranes
- **Carbohydrates**: Primary energy source
- **Nucleic Acids**: DNA/RNA components
- **Vitamins**: Metabolic cofactors
- **Minerals**: Structural and catalytic elements
- **Toxins**: Defensive compounds in plants

**Nutritional System**:
- **Species-Specific Needs**: Different requirements for herbivores, carnivores, omnivores
- **Molecular Metabolism**: Processing efficiency based on traits
- **Nutritional Status**: Health tracking based on molecular needs
- **Evolutionary Pressure**: Fitness affected by nutritional adequacy

**Trigger Conditions**:
- Molecular consumption: When entities eat plants or prey
- Metabolism: Continuous processing of stored molecules
- Nutritional assessment: Regular evaluation of molecular balance
- Evolutionary adaptation: Gradual optimization of dietary preferences

---

## Social and Behavioral Systems

### 10. Communication System (`communication.go`)

**Purpose**: Enables entities to share information through various signal types.

**Signal Types**:
1. **Alert** - Danger warnings and threat notifications
2. **Food** - Location of food sources
3. **Mating** - Reproductive availability and courtship
4. **Territory** - Territorial claims and boundaries
5. **Distress** - Calls for help and assistance
6. **Cooperation** - Invitations for group activities

**Signal Mechanics**:
- **Propagation**: Signals spread outward from source with decay
- **Reception**: Entities detect signals based on vision and intelligence
- **Response**: Behavioral changes triggered by received signals
- **Overlap**: Multiple signals can exist simultaneously

**Trigger Conditions**:
- Signal emission: Based on entity state (fear, hunger, mating drive)
- Signal decay: Gradual weakening over time and distance
- Signal response: When compatible entities receive signals
- Environmental interference: Weather and terrain affect propagation

---

### 11. Caste System (`caste_system.go`)

**Purpose**: Social specialization creating complex colony structures.

**Caste Roles**:
1. **Queen** - Reproductive specialist (3.0x reproduction rate)
2. **Worker** - General labor and resource gathering
3. **Soldier** - Defense and protection specialists
4. **Drone** - Male reproductive supporting queens
5. **Scout** - Exploration and information gathering
6. **Nurse** - Care for young and colony maintenance
7. **Builder** - Construction and infrastructure
8. **Specialist** - Unique roles based on specific needs

**Colony Organization**:
- **Automatic Role Assignment**: Based on entity traits and colony needs
- **Dynamic Reassignment**: Roles can change as needs evolve
- **Colony Fitness**: Performance bonuses for optimal role distribution
- **Cross-Species Compatibility**: Advanced entities can join other species' colonies

**Trigger Conditions**:
- Colony formation: When entities with sufficient cooperation traits gather
- Role assignment: Evaluated every 50 ticks for optimization
- Colony expansion: When resources and population support growth
- Role changes: Based on efficiency and changing colony needs

---

### 12. Hive Mind System (`hive_mind.go`)

**Purpose**: Collective intelligence enabling group decision-making and shared knowledge.

**Hive Mind Types**:
1. **SimpleCollective** - Basic shared decision making
2. **SwarmIntelligence** - Coordinated movement and behavior
3. **NeuralNetwork** - Distributed processing network
4. **QuantumMind** - Advanced collective consciousness

**Collective Capabilities**:
- **Shared Memory**: Knowledge of food sources, threats, and safe zones
- **Group Decisions**: Consensus-based action selection
- **Coordinated Movement**: Formation-based group navigation
- **Collective Intelligence**: Enhanced problem-solving through cooperation

**Memory Types**:
- **Food Source Memory**: Locations and quality of food
- **Threat Memory**: Dangerous areas and predator patterns
- **Safe Zone Memory**: Protected areas and shelter locations
- **Behavior Memory**: Successful strategies and techniques

**Trigger Conditions**:
- Hive formation: When compatible entities (intelligence + cooperation) gather
- Memory updates: When hive members discover new information
- Decision making: When group actions are needed
- Memory decay: Gradual forgetting of outdated information

---

### 13. Colony Warfare System (`colony_warfare.go`)

**Purpose**: Inter-colony conflicts and diplomatic relationships.

**Diplomatic Relations**:
1. **Neutral** - No specific relationship
2. **Friendly** - Positive interactions and cooperation
3. **Allied** - Formal alliance with trade and military cooperation
4. **Rival** - Competitive relationship with occasional conflicts
5. **Hostile** - Active antagonism and resource competition
6. **Enemy** - Formal warfare state

**Conflict Types**:
1. **Border Skirmish** - Minor territorial disputes
2. **Resource War** - Competition for limited resources
3. **Total War** - Full-scale military engagement
4. **Raid** - Quick attacks for resources or territory

**Military Mechanics**:
- **Battle Resolution**: Strength calculations based on participants
- **Casualties**: Entities can die in conflicts
- **Territory Changes**: Winners claim defeated colony territories
- **Relationship Effects**: War outcomes affect future diplomatic relations

**Trigger Conditions**:
- Border conflicts: When colonies expand into contested areas
- Resource competition: When critical resources become scarce
- Diplomatic events: Alliance formations and treaty negotiations
- War declarations: Formal conflict initiation based on relationship deterioration

---

## Environmental Systems

### 14. Biome System (`biome_*.go`)

**Purpose**: Environmental diversity creating different survival challenges.

**Biome Types** (16 total):
1. **Plains** - Grassland with moderate conditions
2. **Forest** - Dense vegetation with shelter
3. **Desert** - Hot, dry conditions with water scarcity
4. **Mountain** - High altitude with temperature/pressure challenges
5. **Water** - Aquatic environments requiring swimming
6. **Radiation** - Dangerous zones causing mutations
7. **Soil** - Underground environments for digging specialists
8. **Air** - Aerial zones for flying entities
9. **Ice** - Frozen regions with extreme cold
10. **Rainforest** - High humidity tropical environments
11. **Deep Water** - High-pressure aquatic zones
12. **High Altitude** - Mountain peaks with low oxygen
13. **Hot Spring** - Geothermal areas with unique chemistry
14. **Tundra** - Cold, low-vegetation regions
15. **Swamp** - Wetland areas with specialized challenges
16. **Canyon** - Deep valleys with unique microclimates

**Biome Effects**:
- **Temperature Stress**: Heat and cold tolerance requirements
- **Pressure Effects**: Altitude and depth adaptations needed
- **Oxygen Levels**: Breathing efficiency in different environments
- **Resource Availability**: Food and water distribution varies

**Biome Boundaries**:
- **Soft Boundaries**: Gradual transitions between biomes
- **Sharp Boundaries**: Abrupt environmental changes
- **Ecotone Zones**: Enhanced biodiversity transition areas
- **Barrier Boundaries**: Movement restrictions between zones

**Trigger Conditions**:
- Environmental stress: Continuous evaluation in unsuitable biomes
- Adaptation pressure: Evolutionary pressure toward biome-specific traits
- Migration effects: Movement bonuses/penalties at boundaries
- Resource distribution: Biome-specific food and water availability

---

### 15. Environmental Events System (`environmental_*.go`)

**Purpose**: Dynamic environmental changes creating evolutionary pressure.

**Event Types**:
1. **Geological Events**: Earthquakes, volcanic eruptions, plate tectonics
2. **Weather Events**: Storms, hurricanes, tornadoes, blizzards
3. **Climate Events**: Temperature shifts, precipitation changes
4. **Catastrophic Events**: Cosmic radiation, magnetic storms, ash clouds
5. **Biome Changes**: Terrain transformation from geological activity

**Environmental Pressures**:
1. **Climate Change** - Long-term temperature and precipitation shifts
2. **Pollution Events** - Environmental contamination
3. **Habitat Fragmentation** - Landscape division and isolation
4. **Resource Depletion** - Periodic scarcity cycles
5. **Invasive Species** - Competition from introduced organisms
6. **Extreme Weather** - Severe storms and temperature events

**Event Effects**:
- **Direct Mortality**: Entities can die from severe events
- **Habitat Modification**: Biome changes affect survival
- **Resource Distribution**: Food and water availability changes
- **Migration Pressure**: Entities forced to relocate

**Trigger Conditions**:
- Random events: Probabilistic occurrence based on environmental conditions
- Pressure activation: When thresholds for specific pressures are met
- Seasonal events: Weather patterns tied to time cycles
- Chain reactions: Events triggering other environmental changes

---

### 16. Time and Seasonal System (`time_cycles.go`)

**Purpose**: Day/night cycles and seasonal changes affecting behavior and resources.

**Time Cycles**:
- **Day/Night**: 24-hour cycles affecting activity patterns
- **Seasons**: Spring, Summer, Autumn, Winter with different characteristics
- **Years**: Long-term cycles for major environmental changes

**Seasonal Effects**:
- **Spring**: High reproduction rates, increased plant growth
- **Summer**: Peak activity, maximum resource availability
- **Autumn**: Preparation behaviors, resource storage
- **Winter**: Reduced activity, survival challenges

**Circadian Effects**:
- **Diurnal Entities**: Active during day, sleep at night
- **Nocturnal Entities**: Active at night, rest during day
- **Crepuscular Entities**: Most active at dawn and dusk

**Trigger Conditions**:
- Time progression: Continuous advancement of day/night cycles
- Seasonal transitions: Gradual changes in environmental conditions
- Behavioral shifts: Entity activity patterns change with time
- Reproductive cycles: Mating seasons triggered by time of year

---

## User Interface Systems

### 17. Command Line Interface (`cli.go`)

**Purpose**: Terminal-based real-time visualization and control of the simulation.

**View Modes** (26 total):
1. **Grid** - Main 2D world visualization
2. **Stats** - Population statistics and trait distributions
3. **Events** - Recent world events and occurrences
4. **Populations** - Detailed demographic analysis
5. **Communication** - Signal activity and patterns
6. **Civilization** - Colony structures and development
7. **Physics** - Physics simulation state
8. **Wind** - Atmospheric patterns and dispersal
9. **Species** - Species tracking and genetic analysis
10. **Network** - Underground plant network visualization
11. **DNA** - Genetic sequences and inheritance
12. **Cellular** - Cell types and organelle development
13. **Evolution** - Macro evolution and phylogenetic trees
14. **Topology** - World terrain and geological features
15. **Tools** - Tool creation and usage statistics
16. **Environment** - Environmental modifications and structures
17. **Behavior** - Emergent behavior discovery and spread
18. **Reproduction** - Mating patterns and reproductive success
19. **Statistical** - Data analysis and trend detection
20. **Ecosystem** - Overall ecosystem health metrics
21. **Anomalies** - Unusual events and statistical outliers
22. **Warfare** - Colony conflicts and diplomatic relations
23. **Fungal** - Decomposer organisms and nutrient cycling
24. **Cultural** - Knowledge systems and cultural evolution
25. **Symbiotic** - Parasitic and mutualistic relationships
26. **Neural** - Neural network learning and behavior

**Navigation Controls**:
- **Arrow Keys**: Move viewport around the world
- **+/-**: Zoom in/out (4 zoom levels: 1x, 2x, 4x, 8x)
- **V**: Cycle through view modes
- **Space**: Pause/resume simulation
- **R**: Reset simulation to initial state
- **Q**: Quit application

**Interactive Features**:
- **Entity Selection**: Click on entities for detailed information
- **Follow Mode**: Track specific entities as they move
- **Real-time Updates**: Display refreshes every 100ms
- **Statistics Overlays**: Numerical data displayed alongside visuals

---

### 18. Web Interface (`web_interface.go`)

**Purpose**: Browser-based interface for remote simulation access and enhanced visualization.

**Web Features**:
- **Real-time Updates**: WebSocket-based live data streaming
- **Responsive Design**: Mobile-friendly adaptive layout
- **All View Modes**: Complete feature parity with CLI interface
- **Player Controls**: Interactive game elements for user participation
- **Export Functions**: Data export in CSV/JSON formats

**Player System**:
- **Player Registration**: Join game with custom player name
- **Species Creation**: Create custom species with trait modifications
- **Species Control**: Direct movement and actions for owned species
- **Event Notifications**: Real-time alerts for species events

**Data Visualization**:
- **Grid View**: Rich graphical representation of world state
- **Statistics Charts**: Real-time graphs and metrics
- **Interactive Elements**: Click-to-explore detailed information
- **Legend System**: Comprehensive symbol and color explanations

**Trigger Conditions**:
- WebSocket connection: Automatic connection on page load
- Data updates: Broadcast every 100ms to all connected clients
- User interactions: Real-time command processing
- Reconnection: Automatic retry on connection loss

---

## System Interactions and Triggers

### Update Order (Each Tick)

The main simulation loop (`world.Update()`) processes systems in this order:

1. **Time System**: Advance day/night cycles and seasons
2. **Environmental Events**: Process geological and weather events
3. **Entity Updates**: Update all living entities (movement, aging, energy)
4. **Plant System**: Plant growth, reproduction, network communication
5. **Wind System**: Atmospheric simulation and dispersal
6. **Biorhythm System**: Entity activity patterns and needs
7. **Molecular System**: Nutritional processing and metabolism
8. **Insect Pollination**: Plant-insect interactions
9. **Communication System**: Signal propagation and response
10. **Civilization System**: Colony management and structures
11. **Emergent Behavior**: Behavior discovery and learning
12. **Reproduction System**: Mating and genetic inheritance
13. **Population Management**: Species tracking and selection pressure
14. **Caste System**: Social role assignment and optimization
15. **Hive Mind System**: Collective intelligence processing
16. **Neural AI System**: Learning network updates
17. **Neural Decision Processing**: AI-driven entity behavior
18. **Colony Warfare**: Diplomatic relations and conflicts
19. **Biome Boundaries**: Environmental transition effects
20. **Statistics Collection**: Data gathering and analysis
21. **Event Logging**: Record significant occurrences
22. **UI Updates**: Refresh display systems

### Cross-System Dependencies

**Critical Dependencies**:
- **Neural Networks** depend on **Entity Intelligence** (trait > 0.3)
- **Reproduction** depends on **Seasonal System** (spring/summer peak)
- **Plant Networks** depend on **Plant System** and **Wind System**
- **Colony Warfare** depends on **Caste System** and **Communication**
- **Emergent Behavior** depends on **Intelligence** and **Environmental Pressure**

**Feedback Loops**:
- **Neural Learning** → **Entity Behavior** → **Fitness** → **Reproduction** → **Genetic Traits** → **Neural Capacity**
- **Environmental Events** → **Selection Pressure** → **Genetic Adaptation** → **Environmental Resistance**
- **Communication** → **Group Behavior** → **Collective Success** → **Social Evolution**

### Performance Considerations

**Update Frequencies**:
- **Core Systems**: Every tick (100ms)
- **Statistics**: Every 20 ticks (2 seconds)
- **Cleanup Operations**: Every 100 ticks (10 seconds)
- **Experience Decay**: Every 100 ticks (10 seconds)
- **Environmental Pressures**: Every 200 ticks (20 seconds)

**Optimization Strategies**:
- **Spatial Partitioning**: Efficient neighbor queries
- **Conditional Updates**: Skip processing for inactive systems
- **Batch Operations**: Group similar operations together
- **Memory Management**: Regular cleanup of dead entities and old data

This documentation provides a complete overview of how EvoSim's complex systems work together to create a realistic evolutionary ecosystem simulation.