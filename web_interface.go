package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
	
	"golang.org/x/net/websocket"
)

// WebInterface manages the web-based interface for the simulation
type WebInterface struct {
	world           *World
	viewManager     *ViewManager
	clients         map[*websocket.Conn]bool
	clientsMutex    sync.RWMutex
	broadcastChan   chan *ViewData
	stopChan        chan bool
	updateInterval  time.Duration
	playerManager   *PlayerManager
	clientPlayers   map[*websocket.Conn]string // maps websocket connections to player IDs
}

// NewWebInterface creates a new web interface
func NewWebInterface(world *World) *WebInterface {
	return &WebInterface{
		world:          world,
		viewManager:    NewViewManager(world),
		clients:        make(map[*websocket.Conn]bool),
		broadcastChan:  make(chan *ViewData, 100),
		stopChan:       make(chan bool),
		updateInterval: 100 * time.Millisecond, // 10 FPS
		playerManager:  NewPlayerManager(),
		clientPlayers:  make(map[*websocket.Conn]string),
	}
}

// RunWebInterface starts the web interface server
func RunWebInterface(world *World, port int) error {
	webInterface := NewWebInterface(world)
	
	// Start the simulation update loop
	go webInterface.simulationLoop()
	
	// Start the broadcast loop
	go webInterface.broadcastLoop()
	
	// Set up HTTP routes
	http.HandleFunc("/", webInterface.serveHome)
	http.HandleFunc("/api/status", webInterface.handleStatus)
	http.Handle("/ws", websocket.Handler(webInterface.handleWebSocket))
	
	// Serve static files (CSS, JS)
	http.HandleFunc("/static/", webInterface.serveStatic)
	
	address := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting web interface on http://localhost%s\n", address)
	fmt.Println("Press Ctrl+C to stop the server")
	
	return http.ListenAndServe(address, nil)
}

// serveHome serves the main HTML page
func (wi *WebInterface) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>EvoSim - Genetic Ecosystem Simulation</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            margin: 0;
            padding: 20px;
            background-color: #1a1a1a;
            color: #ffffff;
        }
        
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        
        .status-bar {
            background-color: #2a2a2a;
            padding: 10px;
            border-radius: 5px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .main-content {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 20px;
        }
        
        .simulation-view {
            background-color: #2a2a2a;
            border-radius: 5px;
            padding: 20px;
            position: relative;
        }
        
        .info-panel {
            background-color: #2a2a2a;
            border-radius: 5px;
            padding: 20px;
        }
        
        .grid-container {
            font-family: monospace;
            font-size: 12px;
            line-height: 14px;
            white-space: pre;
            background-color: #000000;
            padding: 10px;
            border-radius: 3px;
            overflow-x: auto;
        }
        
        .controls {
            margin-bottom: 20px;
        }
        
        .controls button {
            background-color: #4a4a4a;
            color: white;
            border: none;
            padding: 8px 16px;
            margin: 2px;
            border-radius: 3px;
            cursor: pointer;
        }
        
        .controls button:hover {
            background-color: #5a5a5a;
        }
        
        .controls button.active {
            background-color: #6a6a6a;
        }
        
        .stats-section {
            margin-bottom: 20px;
            padding: 10px;
            background-color: #3a3a3a;
            border-radius: 3px;
        }
        
        .stats-section h3 {
            margin: 0 0 10px 0;
            color: #cccccc;
        }
        
        .connection-status {
            padding: 5px 10px;
            border-radius: 3px;
            font-size: 14px;
        }
        
        .connected {
            background-color: #2d5a2d;
            color: #90ee90;
        }
        
        .disconnected {
            background-color: #5a2d2d;
            color: #ff6b6b;
        }
        
        .legend {
            font-size: 11px;
            line-height: 16px;
        }
        
        .view-tabs {
            display: flex;
            flex-wrap: wrap;
            margin-bottom: 10px;
        }
        
        .view-tab {
            background-color: #4a4a4a;
            color: white;
            border: none;
            padding: 5px 10px;
            margin: 2px;
            border-radius: 3px;
            cursor: pointer;
            font-size: 12px;
        }
        
        .view-tab:hover {
            background-color: #5a5a5a;
        }
        
        .view-tab.active {
            background-color: #6a9bd2;
        }
        
        .population-item {
            background-color: #3a3a3a;
            padding: 10px;
            margin: 5px 0;
            border-radius: 5px;
            border-left: 3px solid transparent;
            transition: all 0.3s ease;
        }
        
        .population-item.updating {
            border-left-color: #4CAF50;
            background-color: #2d4a2d;
        }
        
        .update-indicator {
            color: #4CAF50;
            animation: pulse 1s infinite;
        }
        
        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }
        
        .event-item {
            background-color: #3a3a3a;
            padding: 8px;
            margin: 5px 0;
            border-radius: 3px;
            border-left: 3px solid #ff6b6b;
        }
        
        .species-item {
            background-color: #3a3a3a;
            padding: 5px;
            margin: 3px 0;
            border-radius: 3px;
        }
        
        .grid-container {
            font-family: monospace;
            font-size: 12px;
            line-height: 14px;
            white-space: pre;
            background-color: #000000;
            padding: 10px;
            border-radius: 3px;
            overflow-x: auto;
            position: relative;
        }
        
        /* Rich graphics for grid cells */
        .grid-cell {
            display: inline-block;
            width: 12px;
            height: 14px;
            position: relative;
        }
        
        .biome-plains { background-color: #2d5a2d; }
        .biome-forest { background-color: #1a3d1a; }
        .biome-desert { background-color: #8b8000; }
        .biome-mountain { background-color: #696969; }
        .biome-water { background-color: #191970; }
        .biome-radiation { background-color: #8b0000; }
        
        .entity-herbivore { color: #90ee90; }
        .entity-predator { color: #ff6b6b; }
        .entity-omnivore { color: #87ceeb; }
        
        .plant-grass { color: #32cd32; }
        .plant-bush { color: #228b22; }
        .plant-tree { color: #006400; }
        .plant-mushroom { color: #9370db; }
        .plant-algae { color: #00ffff; }
        .plant-cactus { color: #808000; }
        
        .rich-grid {
            font-family: monospace;
            line-height: 1;
            background-color: #000000;
            padding: 10px;
            border-radius: 3px;
            overflow-x: auto;
        }
        
        .grid-row {
            white-space: nowrap;
        }
        
        .grid-cell {
            display: inline-block;
            width: 16px;
            height: 16px;
            position: relative;
            text-align: center;
            font-size: 12px;
            line-height: 16px;
            border-radius: 1px;
            margin: 0;
            vertical-align: top;
        }
        
        .has-event {
            animation: blink 1s infinite;
        }
        
        @keyframes blink {
            0%, 50% { opacity: 1; }
            51%, 100% { opacity: 0.5; }
        }
        
        .event-overlay {
            position: absolute;
            top: 0;
            right: 0;
            font-size: 8px;
            color: yellow;
        }
        
        /* Player Controls Styles */
        .player-controls {
            background-color: #2a4a2a;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 15px;
            border: 2px solid #4CAF50;
        }
        
        .join-form, .species-form, .control-form {
            background-color: #3a3a3a;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 15px;
            border: 1px solid #555;
        }
        
        .join-form input, .species-form input, .control-form input, .control-form select {
            width: 100%;
            padding: 8px;
            margin: 5px 0;
            border: 1px solid #555;
            border-radius: 3px;
            background-color: #2a2a2a;
            color: white;
        }
        
        .trait-adjustments {
            margin: 10px 0;
        }
        
        .trait-adjustments label {
            display: block;
            margin: 8px 0;
            font-size: 14px;
        }
        
        .trait-adjustments input[type="range"] {
            width: 60%;
            margin: 0 10px;
        }
        
        .control-commands {
            margin: 15px 0;
        }
        
        .move-controls, .action-controls {
            margin: 10px 0;
            padding: 10px;
            background-color: #2a2a2a;
            border-radius: 3px;
        }
        
        .form-buttons {
            margin-top: 15px;
        }
        
        .form-buttons button, .control-commands button {
            margin: 5px;
            padding: 8px 15px;
        }
        
        .error-message {
            color: #ff6b6b;
            margin-top: 10px;
            padding: 8px;
            background-color: #5a2d2d;
            border-radius: 3px;
        }
        
        #player-status {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            font-weight: bold;
        }
        
        .control-buttons {
            display: flex;
            gap: 10px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🌍 EvoSim - Genetic Ecosystem Simulation</h1>
    </div>
    
    <div class="status-bar">
        <div>
            <span id="tick">Tick: 0</span> |
            <span id="time">Time: Unknown</span> |
            <span id="entities">Entities: 0</span> |
            <span id="plants">Plants: 0</span> |
            <span id="populations">Populations: 0</span>
        </div>
        <div class="connection-status" id="connection-status">
            Disconnected
        </div>
    </div>
    
    <div class="main-content">
        <div class="simulation-view">
            <!-- Player Controls Section -->
            <div class="player-controls" id="player-controls" style="display: none;">
                <h3>🎮 Player Controls</h3>
                <div id="player-status">
                    <span id="player-name">Not logged in</span>
                    <span id="player-species-count">0 species</span>
                </div>
                <div class="control-buttons">
                    <button id="create-species-btn" onclick="showCreateSpeciesForm()">🧬 Create Species</button>
                    <button id="control-species-btn" onclick="showControlSpeciesForm()">🎯 Control Species</button>
                </div>
            </div>

            <!-- Join Game Form -->
            <div class="join-form" id="join-form">
                <h3>🎮 Join as Player</h3>
                <input type="text" id="player-name-input" placeholder="Enter your name (letters and numbers only)" maxlength="50">
                <button onclick="joinAsPlayer()">Join Game</button>
                <div id="join-error" class="error-message" style="display: none;"></div>
            </div>

            <!-- Create Species Form -->
            <div class="species-form" id="create-species-form" style="display: none;">
                <h3>🧬 Create New Species</h3>
                <input type="text" id="species-name-input" placeholder="Species name" maxlength="50">
                <div class="trait-adjustments">
                    <h4>Trait Adjustments (±0.3 range):</h4>
                    <label>Speed: <input type="range" id="speed-trait" min="-0.3" max="0.3" step="0.1" value="0"> <span id="speed-value">0.0</span></label>
                    <label>Aggression: <input type="range" id="aggression-trait" min="-0.3" max="0.3" step="0.1" value="0"> <span id="aggression-value">0.0</span></label>
                    <label>Cooperation: <input type="range" id="cooperation-trait" min="-0.3" max="0.3" step="0.1" value="0"> <span id="cooperation-value">0.0</span></label>
                    <label>Intelligence: <input type="range" id="intelligence-trait" min="-0.3" max="0.3" step="0.1" value="0"> <span id="intelligence-value">0.0</span></label>
                </div>
                <div class="form-buttons">
                    <button onclick="createSpecies()">Create Species</button>
                    <button onclick="hideCreateSpeciesForm()">Cancel</button>
                </div>
                <div id="create-species-error" class="error-message" style="display: none;"></div>
            </div>

            <!-- Control Species Form -->
            <div class="control-form" id="control-species-form" style="display: none;">
                <h3>🎯 Control Your Species</h3>
                <select id="species-select">
                    <option value="">Select a species to control</option>
                </select>
                <div class="control-commands">
                    <div class="move-controls">
                        <h4>Movement Control</h4>
                        <div>Click on the grid below to move your species to that location</div>
                        <div>Current target: <span id="move-target">None</span></div>
                        <button onclick="executeMove()">Move to Target</button>
                    </div>
                    <div class="action-controls">
                        <h4>Action Commands</h4>
                        <button onclick="executeGather()">🌱 Gather Resources</button>
                        <button onclick="executeReproduce()">👶 Encourage Reproduction</button>
                    </div>
                </div>
                <button onclick="hideControlSpeciesForm()">Close Controls</button>
                <div id="control-species-error" class="error-message" style="display: none;"></div>
            </div>
            
            <div class="controls">
                <button id="pause-btn" onclick="togglePause()">⏸ Pause</button>
                <button onclick="resetSimulation()">🔄 Reset</button>
                <button onclick="saveState()">💾 Save</button>
                <button onclick="loadState()">📁 Load</button>
                <input type="file" id="load-file" accept=".json" style="display: none;" onchange="handleFileLoad(event)">
            </div>
            
            <div class="view-tabs" id="view-tabs">
                <!-- View tabs will be populated by JavaScript -->
            </div>
            
            <div id="view-content">
                <div class="grid-container" id="grid-view">
                    Loading simulation...
                </div>
            </div>
        </div>
        
        <div class="info-panel">
            <div class="stats-section">
                <h3>📊 Statistics</h3>
                <div id="stats-content">
                    <div>Average Fitness: <span id="avg-fitness">0.00</span></div>
                    <div>Average Energy: <span id="avg-energy">0.00</span></div>
                    <div>Average Age: <span id="avg-age">0.00</span></div>
                </div>
            </div>
            
            <div class="stats-section">
                <h3>👥 Populations</h3>
                <div id="populations-content">
                    No populations
                </div>
            </div>
            
            <div class="stats-section">
                <h3>📡 Communication</h3>
                <div id="communication-content">
                    Active Signals: <span id="active-signals">0</span>
                </div>
            </div>
            
            <div class="stats-section">
                <h3>🌬️ Wind System</h3>
                <div id="wind-content">
                    <div>Direction: <span id="wind-direction">0°</span></div>
                    <div>Strength: <span id="wind-strength">0.0</span></div>
                    <div>Weather: <span id="weather-pattern">Calm</span></div>
                </div>
            </div>
            
            <div class="stats-section legend">
                <h3>🌱 Legend</h3>
                <div>
                    <strong>Biomes:</strong><br>
                    • Plains | ♠ Forest | ~ Desert | ^ Mountain<br>
                    ≈ Water | ☢ Radiation | ■ Soil | ○ Air<br>
                    ❄ Ice | 🌳 Rainforest | ≈ Deep Water | ▲ High Altitude<br>
                    ◉ Hot Spring | ○ Tundra | ≋ Swamp | ◢ Canyon<br><br>
                    
                    <strong>Entities (single):</strong><br>
                    🐰 Herbivore | 🐺 Predator | 🐻 Omnivore | 🦋 Generic<br>
                    Numbers = Multiple entities<br><br>
                    
                    <strong>Plants:</strong><br>
                    🌱 Grass | 🌿 Bush | 🌳 Tree<br>
                    🍄 Mushroom | 🌊 Algae | 🌵 Cactus
                </div>
            </div>
        </div>
    </div>
    
    <script>
        let ws = null;
        let isPaused = false;
        let currentView = 'GRID';
        let playerID = null;
        let playerSpecies = [];
        let selectedSpecies = null;
        let moveTarget = null;
        
        const viewModes = [
            'GRID', 'STATS', 'EVENTS', 'POPULATIONS', 'COMMUNICATION',
            'CIVILIZATION', 'PHYSICS', 'WIND', 'SPECIES', 'NETWORK',
            'DNA', 'CELLULAR', 'EVOLUTION', 'TOPOLOGY', 'TOOLS', 'ENVIRONMENT', 'BEHAVIOR',
            'STATISTICAL', 'ANOMALIES'
        ];
        
        // Initialize view tabs
        function initViewTabs() {
            const tabsContainer = document.getElementById('view-tabs');
            viewModes.forEach(mode => {
                const button = document.createElement('button');
                button.className = 'view-tab';
                button.textContent = mode;
                if (mode === currentView) {
                    button.classList.add('active');
                }
                button.onclick = () => switchView(mode);
                tabsContainer.appendChild(button);
            });
        }
        
        // Switch view mode
        function switchView(mode) {
            currentView = mode;
            document.querySelectorAll('.view-tab').forEach(tab => {
                tab.classList.toggle('active', tab.textContent === mode);
            });
            
            // Update content based on view
            updateViewContent();
        }
        
        // Connect to WebSocket
        function connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws';
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                console.log('Connected to simulation');
                document.getElementById('connection-status').textContent = 'Connected';
                document.getElementById('connection-status').className = 'connection-status connected';
            };
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                
                // Check if this is a player-specific message
                if (data.type && ['player_joined', 'species_created', 'command_executed', 'error'].includes(data.type)) {
                    handlePlayerMessage(data);
                    return;
                }
                
                // Otherwise treat it as simulation data
                console.log('WebSocket data received, tick:', data.tick, 'entities:', data.entity_count, 'grid length:', data.grid ? data.grid.length : 'null');
                updateDisplay(data);
            };
            
            ws.onclose = function() {
                console.log('Disconnected from simulation');
                document.getElementById('connection-status').textContent = 'Disconnected';
                document.getElementById('connection-status').className = 'connection-status disconnected';
                
                // Attempt to reconnect after 3 seconds
                setTimeout(connect, 3000);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        }
        
        // Update display with new simulation data
        function updateDisplay(data) {
            // Update status bar
            document.getElementById('tick').textContent = 'Tick: ' + data.tick;
            document.getElementById('time').textContent = 'Time: ' + data.time_string;
            document.getElementById('entities').textContent = 'Entities: ' + data.entity_count;
            document.getElementById('plants').textContent = 'Plants: ' + data.plant_count;
            document.getElementById('populations').textContent = 'Populations: ' + data.population_count;
            
            // Update stats
            if (data.stats.avg_fitness !== undefined) {
                document.getElementById('avg-fitness').textContent = data.stats.avg_fitness.toFixed(2);
            }
            if (data.stats.avg_energy !== undefined) {
                document.getElementById('avg-energy').textContent = data.stats.avg_energy.toFixed(2);
            }
            if (data.stats.avg_age !== undefined) {
                document.getElementById('avg-age').textContent = data.stats.avg_age.toFixed(1);
            }
            
            // Update populations
            let populationsHtml = '';
            data.populations.forEach(pop => {
                populationsHtml += '<div><strong>' + pop.name + '</strong>: ' + pop.count + 
                                 ' (Fitness: ' + pop.avg_fitness.toFixed(2) + ')</div>';
            });
            document.getElementById('populations-content').innerHTML = populationsHtml || 'No populations';
            
            // Update communication
            document.getElementById('active-signals').textContent = data.communication.active_signals;
            
            // Update wind
            document.getElementById('wind-direction').textContent = (data.wind.direction * 180 / Math.PI).toFixed(1) + '°';
            document.getElementById('wind-strength').textContent = data.wind.strength.toFixed(2);
            document.getElementById('weather-pattern').textContent = data.wind.weather_pattern;
            
            // Update main view content
            updateViewContent(data);
        }
        
        // Update view content based on current view mode
        function updateViewContent(data = null) {
            if (!data) return;
            
            const viewContent = document.getElementById('view-content');
            
            switch (currentView) {
                case 'GRID':
                    const gridHtml = renderGrid(data.grid);
                    console.log('Grid HTML length:', gridHtml.length, 'First 100 chars:', gridHtml.substring(0, 100));
                    // Update the existing grid container directly
                    const gridContainer = document.getElementById('grid-view');
                    if (gridContainer) {
                        gridContainer.innerHTML = gridHtml;
                        // Add click listener for movement control
                        gridContainer.onclick = handleGridClick;
                    } else {
                        viewContent.innerHTML = '<div class="grid-container" id="grid-view" onclick="handleGridClick(event)">' + gridHtml + '</div>';
                    }
                    break;
                    
                case 'STATS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderStats(data.stats) + '</div>';
                    break;
                    
                case 'EVENTS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderEvents(data.events) + '</div>';
                    break;
                    
                case 'POPULATIONS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderPopulations(data.populations, data.population_history) + '</div>';
                    break;
                    
                case 'COMMUNICATION':
                    viewContent.innerHTML = '<div class="stats-section">' + renderCommunication(data.communication, data.communication_history) + '</div>';
                    break;
                    
                case 'CIVILIZATION':
                    viewContent.innerHTML = '<div class="stats-section">' + renderCivilization(data.civilization) + '</div>';
                    break;
                    
                case 'PHYSICS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderPhysics(data.physics, data.physics_history) + '</div>';
                    break;
                    
                case 'WIND':
                    viewContent.innerHTML = '<div class="stats-section">' + renderWind(data.wind) + '</div>';
                    break;
                    
                case 'SPECIES':
                    viewContent.innerHTML = '<div class="stats-section">' + renderSpecies(data.species) + '</div>';
                    break;
                    
                case 'NETWORK':
                    viewContent.innerHTML = '<div class="stats-section">' + renderNetwork(data.network) + '</div>';
                    break;
                    
                case 'DNA':
                    viewContent.innerHTML = '<div class="stats-section">' + renderDNA(data.dna) + '</div>';
                    break;
                    
                case 'CELLULAR':
                    viewContent.innerHTML = '<div class="stats-section">' + renderCellular(data.cellular) + '</div>';
                    break;
                    
                case 'EVOLUTION':
                    viewContent.innerHTML = '<div class="stats-section">' + renderEvolution(data.evolution) + '</div>';
                    break;
                    
                case 'TOPOLOGY':
                    viewContent.innerHTML = '<div class="stats-section">' + renderTopology(data.topology) + '</div>';
                    break;
                    
                case 'TOOLS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderTools(data.tools) + '</div>';
                    break;
                    
                case 'ENVIRONMENT':
                    viewContent.innerHTML = '<div class="stats-section">' + renderEnvironment(data.environmental_mod) + '</div>';
                    break;
                    
                case 'BEHAVIOR':
                    viewContent.innerHTML = '<div class="stats-section">' + renderBehavior(data.emergent_behavior) + '</div>';
                    break;
                    
                case 'STATISTICAL':
                    viewContent.innerHTML = '<div class="stats-section">' + renderStatistical(data.statistical) + '</div>';
                    break;
                    
                case 'ANOMALIES':
                    viewContent.innerHTML = '<div class="stats-section">' + renderAnomalies(data.anomalies) + '</div>';
                    break;
                    
                default:
                    viewContent.innerHTML = '<div class="stats-section"><h3>' + currentView + '</h3><p>View not yet implemented</p></div>';
            }
        }
        
        // Render grid view with rich graphics
        function renderGrid(grid) {
            if (!grid || grid.length === 0) {
                return '<div>No grid data available</div>';
            }
            
            let result = '<div class="rich-grid">';
            for (let y = 0; y < grid.length; y++) {
                result += '<div class="grid-row">';
                for (let x = 0; x < grid[y].length; x++) {
                    const cell = grid[y][x];
                    let cellClass = 'grid-cell ';
                    let cellContent = '';
                    
                    // Determine biome background
                    cellClass += getBiomeClass(cell.biome);
                    
                    // Determine content (entities take priority over plants over biome)
                    if (cell.entity_count > 0) {
                        cellClass += ' ' + getEntityClass(cell.entity_symbol);
                        cellContent = getEntityDisplay(cell.entity_symbol, cell.entity_count);
                    } else if (cell.plant_count > 0) {
                        cellClass += ' ' + getPlantClass(cell.plant_symbol);
                        cellContent = getPlantDisplay(cell.plant_symbol, cell.plant_count);
                    } else {
                        cellContent = getBiomeDisplay(cell.biome_symbol);
                    }
                    
                    // Ensure we always have some content
                    if (!cellContent || cellContent.trim() === '') {
                        cellContent = '.'; // Default fallback
                    }
                    
                    // Add special indicators
                    if (cell.has_event) {
                        cellClass += ' has-event';
                        cellContent += '<span class="event-overlay">⚡</span>';
                    }
                    
                    result += '<span class="' + cellClass + '" title="' + getCellTooltip(cell) + '">' + cellContent + '</span>';
                }
                result += '</div>';
            }
            result += '</div>';
            return result;
        }
        
        function getBiomeClass(biome) {
            const biomeClasses = {
                'Plains': 'biome-plains',
                'Forest': 'biome-forest', 
                'Desert': 'biome-desert',
                'Mountain': 'biome-mountain',
                'Water': 'biome-water',
                'Radiation': 'biome-radiation'
            };
            return biomeClasses[biome] || 'biome-plains';
        }
        
        function getEntityClass(symbol) {
            if (symbol === 'H') return 'entity-herbivore';
            if (symbol === 'P') return 'entity-predator';
            if (symbol === 'O') return 'entity-omnivore';
            return 'entity-generic';
        }
        
        function getPlantClass(symbol) {
            const plantClasses = {
                '.': 'plant-grass',
                '♦': 'plant-bush',
                '♠': 'plant-tree',
                '♪': 'plant-mushroom',
                '≈': 'plant-algae',
                '†': 'plant-cactus'
            };
            return plantClasses[symbol] || 'plant-grass';
        }
        
        function getEntityDisplay(symbol, count) {
            if (count === 1) {
                // Use styled symbols for single entities
                const entitySymbols = {
                    'H': '🐰', // Herbivore
                    'P': '🐺', // Predator  
                    'O': '🐻', // Omnivore
                    'E': '🦋'  // Generic entity
                };
                return entitySymbols[symbol] || symbol;
            } else {
                // Use count for multiple entities
                return count < 10 ? count.toString() : '+';
            }
        }
        
        function getPlantDisplay(symbol, count) {
            const plantSymbols = {
                '.': '🌱', // Grass
                '♦': '🌿', // Bush
                '♠': '🌳', // Tree
                '♪': '🍄', // Mushroom
                '≈': '🌊', // Algae
                '†': '🌵'  // Cactus
            };
            return plantSymbols[symbol] || symbol;
        }
        
        function getBiomeDisplay(symbol) {
            // Show biome symbols for empty cells
            return symbol || '.';
        }
        
        function getCellTooltip(cell) {
            let tooltip = 'Biome: ' + cell.biome;
            if (cell.entity_count > 0) {
                tooltip += ', Entities: ' + cell.entity_count;
            }
            if (cell.plant_count > 0) {
                tooltip += ', Plants: ' + cell.plant_count;
            }
            if (cell.has_event) {
                tooltip += ', Event Active';
            }
            return tooltip;
        }
        
        // Render stats view with enhanced information
        function renderStats(stats) {
            let html = '<h3>📊 World Statistics</h3>';
            
            html += '<h4>General Stats:</h4>';
            for (const [key, value] of Object.entries(stats)) {
                const displayKey = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                const displayValue = typeof value === 'number' ? value.toFixed(2) : value;
                html += '<div>' + displayKey + ': ' + displayValue + '</div>';
            }
            
            html += '<h4>System Health:</h4>';
            if (stats.avg_fitness !== undefined) {
                if (stats.avg_fitness < 0.3) {
                    html += '<div style="color: orange;">⚠️ Low average fitness - population struggling</div>';
                } else if (stats.avg_fitness < 0.6) {
                    html += '<div style="color: yellow;">⚡ Moderate fitness - population stable</div>';
                } else {
                    html += '<div style="color: lightgreen;">✅ High fitness - population thriving</div>';
                }
            }
            
            if (stats.avg_energy !== undefined) {
                if (stats.avg_energy < 30) {
                    html += '<div style="color: orange;">⚠️ Low energy levels - resource scarcity</div>';
                } else if (stats.avg_energy < 60) {
                    html += '<div style="color: yellow;">⚡ Moderate energy - adequate resources</div>';
                } else {
                    html += '<div style="color: lightgreen;">✅ High energy - abundant resources</div>';
                }
            }
            
            // Enhanced ecosystem analysis
            html += '<h4>🌍 Ecosystem Analysis:</h4>';
            const fitnessEnergy = (stats.avg_fitness || 0) * (stats.avg_energy || 0) / 100;
            if (fitnessEnergy < 0.2) {
                html += '<div style="color: red;">🔥 Critical ecosystem stress</div>';
            } else if (fitnessEnergy < 0.5) {
                html += '<div style="color: orange;">⚠️ Ecosystem under pressure</div>';
            } else if (fitnessEnergy < 0.8) {
                html += '<div style="color: yellow;">⚡ Stable ecosystem</div>';
            } else {
                html += '<div style="color: lightgreen;">🌟 Thriving ecosystem</div>';
            }
            
            return html;
        }
        
        // Track previous population data for stable ordering
        let previousPopulations = [];
        let populationUpdateIndicators = {};
        
        // Render populations view with stable ordering and historical data
        function renderPopulations(populations, populationHistory = []) {
            let html = '<h3>👥 Population Details</h3>';
            
            // Sort populations by name for stable ordering
            const sortedPopulations = [...populations].sort((a, b) => a.name.localeCompare(b.name));
            
            // Track which populations have changed
            const currentTime = Date.now();
            sortedPopulations.forEach(pop => {
                const prevPop = previousPopulations.find(p => p.name === pop.name);
                if (prevPop && (
                    prevPop.count !== pop.count || 
                    Math.abs(prevPop.avg_fitness - pop.avg_fitness) > 0.01 ||
                    Math.abs(prevPop.avg_energy - pop.avg_energy) > 0.01
                )) {
                    populationUpdateIndicators[pop.name] = currentTime;
                }
            });
            
            sortedPopulations.forEach(pop => {
                const isUpdating = populationUpdateIndicators[pop.name] && 
                                 (currentTime - populationUpdateIndicators[pop.name]) < 2000; // Show indicator for 2 seconds
                
                html += '<div class="population-item' + (isUpdating ? ' updating' : '') + '">';
                html += '<h4>' + pop.name + (isUpdating ? ' <span class="update-indicator">●</span>' : '') + '</h4>';
                html += '<div>Count: ' + pop.count + '</div>';
                html += '<div>Average Fitness: ' + pop.avg_fitness.toFixed(2) + '</div>';
                html += '<div>Average Energy: ' + pop.avg_energy.toFixed(2) + '</div>';
                html += '<div>Average Age: ' + pop.avg_age.toFixed(1) + '</div>';
                
                if (pop.trait_averages && Object.keys(pop.trait_averages).length > 0) {
                    html += '<h5>Average Traits:</h5>';
                    Object.entries(pop.trait_averages).forEach(([trait, value]) => {
                        html += '<div style="font-size: 0.9em; margin-left: 10px;">' + 
                               trait + ': ' + value.toFixed(3) + '</div>';
                    });
                }
                html += '</div>';
            });
            
            // Add historical data if available
            if (populationHistory && populationHistory.length > 0) {
                html += '<h4>📈 Population History (Last ' + populationHistory.length + ' snapshots):</h4>';
                html += '<div style="max-height: 200px; overflow-y: auto;">';
                populationHistory.slice(-10).forEach(snapshot => {
                    html += '<div style="margin: 5px 0; padding: 5px; background-color: #444; border-radius: 3px;">';
                    html += '<strong>Tick ' + snapshot.tick + '</strong> (' + snapshot.timestamp + ')<br>';
                    snapshot.populations.forEach(pop => {
                        html += '<span style="font-size: 0.8em;">' + pop.name + ': ' + pop.count + 
                               ' (Fitness: ' + pop.avg_fitness.toFixed(2) + ')</span><br>';
                    });
                    html += '</div>';
                });
                html += '</div>';
            }
            
            // Update previous populations for next comparison
            previousPopulations = [...sortedPopulations];
            
            return html;
        }
        
        // Render communication view with historical data
        function renderCommunication(comm, commHistory = []) {
            let html = '<h3>📡 Communication System</h3>';
            html += '<h4>Active Signals:</h4>';
            html += '<div>Total Active: ' + comm.active_signals + '</div>';
            
            if (Object.keys(comm.signal_types).length > 0) {
                html += '<h4>Signal Types:</h4>';
                const signalIcons = {
                    'Danger': '🚨',
                    'Food': '🍎', 
                    'Mating': '💕',
                    'Territory': '🏴',
                    'Help': '🆘',
                    'Migration': '🧭'
                };
                
                for (const [type, count] of Object.entries(comm.signal_types)) {
                    const icon = signalIcons[type] || '📡';
                    html += '<div>' + icon + ' ' + type + ': ' + count + ' active</div>';
                }
            } else {
                html += '<div>No active signals</div>';
            }
            
            html += '<h4>Communication Stats:</h4>';
            if (comm.active_signals === 0) {
                html += '<div>Activity Level: Silent</div>';
            } else if (comm.active_signals < 5) {
                html += '<div>Activity Level: Low communication</div>';
            } else if (comm.active_signals < 15) {
                html += '<div>Activity Level: Moderate communication</div>';
            } else {
                html += '<div>Activity Level: High communication</div>';
            }
            
            // Add historical data if available
            if (commHistory && commHistory.length > 0) {
                html += '<h4>📈 Communication History (Last ' + commHistory.length + ' snapshots):</h4>';
                html += '<div style="max-height: 200px; overflow-y: auto;">';
                commHistory.slice(-10).forEach(snapshot => {
                    html += '<div style="margin: 5px 0; padding: 5px; background-color: #444; border-radius: 3px;">';
                    html += '<strong>Tick ' + snapshot.tick + '</strong> (' + snapshot.timestamp + ')<br>';
                    html += '<span style="font-size: 0.8em;">Active Signals: ' + snapshot.active_signals + '</span><br>';
                    if (snapshot.signal_types && Object.keys(snapshot.signal_types).length > 0) {
                        html += '<span style="font-size: 0.7em;">Types: ';
                        const types = [];
                        for (const [type, count] of Object.entries(snapshot.signal_types)) {
                            types.push(type + ': ' + count);
                        }
                        html += types.join(', ') + '</span>';
                    }
                    html += '</div>';
                });
                html += '</div>';
            }
            
            return html;
        }
        
        // Render events view
        function renderEvents(events) {
            let html = '<h3>🌪️ World Events & Event Log</h3>';
            
            // Separate active and historical events
            const activeEvents = events.filter(event => event.type === 'active');
            const historicalEvents = events.filter(event => event.type === 'historical');
            
            // Active Events Section
            html += '<h4>Active Events:</h4>';
            if (activeEvents.length === 0) {
                html += '<div>No active events</div>';
            } else {
                activeEvents.forEach(event => {
                    html += '<div class="event-item">';
                    html += '<strong>' + event.name + '</strong><br>';
                    html += event.description + '<br>';
                    html += '<small>Duration: ' + event.duration + ' ticks remaining</small>';
                    html += '</div>';
                });
            }
            
            // Historical Events Section
            html += '<h4>Recent History:</h4>';
            if (historicalEvents.length === 0) {
                html += '<div>No historical events recorded</div>';
            } else {
                historicalEvents.forEach(event => {
                    html += '<div class="event-item" style="border-left-color: #888;">';
                    html += '<strong>' + event.name + '</strong> ';
                    html += '<small style="color: #aaa;">(' + event.timestamp + ')</small><br>';
                    html += event.description + '<br>';
                    html += '<small>Tick: ' + event.tick + '</small>';
                    html += '</div>';
                });
            }
            
            return html;
        }
        
        // Render civilization view
        function renderCivilization(civilization) {
            let html = '<h3>🏛️ Civilization System</h3>';
            html += '<div>Active Tribes: ' + civilization.tribes_count + '</div>';
            html += '<div>Total Structures: ' + civilization.structure_count + '</div>';
            html += '<div>Total Resources: ' + civilization.total_resources + '</div>';
            
            if (civilization.tribes_count === 0) {
                html += '<br><div>No tribes formed yet</div>';
            } else {
                html += '<br><h4>Development Status:</h4>';
                if (civilization.structure_count === 0) {
                    html += '<div>Civilization Level: Primitive</div>';
                } else if (civilization.structure_count < 5) {
                    html += '<div>Civilization Level: Developing</div>';
                } else if (civilization.structure_count < 15) {
                    html += '<div>Civilization Level: Advanced</div>';
                } else {
                    html += '<div>Civilization Level: Highly Advanced</div>';
                }
            }
            
            return html;
        }
        
        // Render physics view with historical data
        function renderPhysics(physics, physicsHistory = []) {
            let html = '<h3>⚡ Physics System</h3>';
            html += '<h4>Movement Statistics:</h4>';
            html += '<div>Average Velocity: ' + physics.average_velocity.toFixed(2) + '</div>';
            html += '<div>Total Momentum: ' + physics.total_momentum.toFixed(2) + '</div>';
            
            html += '<h4>Collision Statistics:</h4>';
            html += '<div>Collisions This Tick: ' + physics.collisions_last_tick + '</div>';
            
            if (physics.average_velocity < 0.1) {
                html += '<br><div>Activity Level: Low (mostly stationary entities)</div>';
            } else if (physics.average_velocity < 0.5) {
                html += '<br><div>Activity Level: Medium (moderate movement)</div>';
            } else {
                html += '<br><div>Activity Level: High (active movement)</div>';
            }
            
            // Add historical data if available
            if (physicsHistory && physicsHistory.length > 0) {
                html += '<h4>📈 Physics History (Last ' + physicsHistory.length + ' snapshots):</h4>';
                html += '<div style="max-height: 200px; overflow-y: auto;">';
                physicsHistory.slice(-10).forEach(snapshot => {
                    html += '<div style="margin: 5px 0; padding: 5px; background-color: #444; border-radius: 3px;">';
                    html += '<strong>Tick ' + snapshot.tick + '</strong> (' + snapshot.timestamp + ')<br>';
                    html += '<span style="font-size: 0.8em;">Collisions: ' + snapshot.collisions + 
                           ', Velocity: ' + snapshot.average_velocity.toFixed(2) + 
                           ', Momentum: ' + snapshot.total_momentum.toFixed(2) + '</span>';
                    html += '</div>';
                });
                html += '</div>';
            }
            
            return html;
        }
        
        // Render species view with enhanced details
        function renderSpecies(species) {
            let html = '<h3>🐾 Species Tracking</h3>';
            html += '<div>Active Species: ' + species.active_species + '</div>';
            html += '<div>Extinct Species: ' + species.extinct_species + '</div>';
            
            // Calculate diversity metrics
            const totalSpecies = species.active_species + species.extinct_species;
            const survivalRate = totalSpecies > 0 ? (species.active_species / totalSpecies * 100).toFixed(1) : 100;
            
            html += '<h4>📈 Diversity Metrics:</h4>';
            html += '<div>Species Survival Rate: ' + survivalRate + '%</div>';
            
            if (survivalRate < 30) {
                html += '<div style="color: red;">🔥 High extinction rate - evolutionary crisis</div>';
            } else if (survivalRate < 60) {
                html += '<div style="color: orange;">⚠️ Moderate extinction pressure</div>';
            } else if (survivalRate < 85) {
                html += '<div style="color: yellow;">⚡ Natural selection in progress</div>';
            } else {
                html += '<div style="color: lightgreen;">🌟 Species diversity stable</div>';
            }
            
            if (species.species_details && species.species_details.length > 0) {
                html += '<h4>Species Details:</h4>';
                // Sort by population for better display
                const sortedSpecies = [...species.species_details].sort((a, b) => {
                    if (a.is_extinct !== b.is_extinct) {
                        return a.is_extinct ? 1 : -1; // Active species first
                    }
                    return b.population - a.population; // Higher population first
                });
                
                sortedSpecies.forEach(detail => {
                    html += '<div class="species-item">';
                    html += '<strong>' + detail.name + '</strong>';
                    if (detail.is_extinct) {
                        html += ' <span style="color: red;">💀 (Extinct)</span>';
                    } else {
                        html += ' - Population: ' + detail.population;
                        // Add population health indicator
                        if (detail.population < 5) {
                            html += ' <span style="color: orange;">⚠️ Endangered</span>';
                        } else if (detail.population < 15) {
                            html += ' <span style="color: yellow;">⚡ Vulnerable</span>';
                        } else {
                            html += ' <span style="color: lightgreen;">✅ Stable</span>';
                        }
                    }
                    html += '</div>';
                });
            } else {
                html += '<br><div>No species data available</div>';
            }
            
            return html;
        }
        
        // Render network view
        function renderNetwork(network) {
            let html = '<h3>🌐 Plant Network System</h3>';
            html += '<div>Network Connections: ' + network.connection_count + '</div>';
            html += '<div>Chemical Signals: ' + network.signal_count + '</div>';
            html += '<div>Network Clusters: ' + network.cluster_count + '</div>';
            
            if (network.connection_count === 0) {
                html += '<br><div>No plant networks formed yet</div>';
            } else {
                html += '<br><h4>Network Status:</h4>';
                if (network.cluster_count === 0) {
                    html += '<div>Network Density: Sparse connections</div>';
                } else if (network.cluster_count < 3) {
                    html += '<div>Network Density: Few clusters formed</div>';
                } else {
                    html += '<div>Network Density: Multiple active clusters</div>';
                }
            }
            
            return html;
        }
        
        // Render DNA view
        function renderDNA(dna) {
            let html = '<h3>🧬 DNA System</h3>';
            html += '<div>Organisms: ' + dna.organism_count + '</div>';
            html += '<div>Average Mutations: ' + dna.average_mutations.toFixed(2) + '</div>';
            html += '<div>Average Complexity: ' + dna.average_complexity.toFixed(2) + '</div>';
            
            if (dna.organism_count === 0) {
                html += '<br><div>No DNA-based organisms present</div>';
            } else {
                html += '<br><h4>Genetic Status:</h4>';
                if (dna.average_complexity < 1.0) {
                    html += '<div>Complexity Level: Simple organisms</div>';
                } else if (dna.average_complexity < 3.0) {
                    html += '<div>Complexity Level: Moderate complexity</div>';
                } else {
                    html += '<div>Complexity Level: Complex organisms</div>';
                }
            }
            
            return html;
        }
        
        // Render cellular view
        function renderCellular(cellular) {
            let html = '<h3>🔬 Cellular System</h3>';
            html += '<div>Total Cells: ' + cellular.total_cells + '</div>';
            html += '<div>Average Complexity: ' + cellular.average_complexity.toFixed(2) + '</div>';
            html += '<div>Cell Divisions: ' + cellular.cell_divisions + '</div>';
            
            if (cellular.total_cells === 0) {
                html += '<br><div>No cellular activity detected</div>';
            } else {
                html += '<br><h4>Cellular Activity:</h4>';
                if (cellular.cell_divisions === 0) {
                    html += '<div>Division Activity: No recent divisions</div>';
                } else if (cellular.cell_divisions < 10) {
                    html += '<div>Division Activity: Low division rate</div>';
                } else {
                    html += '<div>Division Activity: Active cell division</div>';
                }
            }
            
            return html;
        }
        
        // Render evolution view
        function renderEvolution(evolution) {
            let html = '<h3>🌿 Evolution Tracking</h3>';
            html += '<div>Speciation Events: ' + evolution.speciation_events + '</div>';
            html += '<div>Extinction Events: ' + evolution.extinction_events + '</div>';
            html += '<div>Genetic Diversity: ' + evolution.genetic_diversity.toFixed(2) + '</div>';
            
            html += '<br><h4>Evolutionary Status:</h4>';
            if (evolution.speciation_events === 0) {
                html += '<div>No speciation detected yet</div>';
            } else if (evolution.speciation_events < 3) {
                html += '<div>Early speciation phase</div>';
            } else {
                html += '<div>Active evolutionary divergence</div>';
            }
            
            if (evolution.extinction_events > 0) {
                html += '<div style="color: orange;">Warning: ' + evolution.extinction_events + ' extinction event(s) occurred</div>';
            }
            
            return html;
        }
        
        // Render topology view
        function renderTopology(topology) {
            let html = '<h3>🗻 World Topology</h3>';
            html += '<div>Elevation Range: ' + topology.elevation_range + '</div>';
            html += '<div>Fluid Regions: ' + topology.fluid_regions + '</div>';
            html += '<div>Geological Age: ' + topology.geological_age + '</div>';
            
            html += '<br><h4>Terrain Analysis:</h4>';
            if (topology.fluid_regions === 0) {
                html += '<div>Terrain Type: Dry landscape</div>';
            } else if (topology.fluid_regions < 3) {
                html += '<div>Terrain Type: Semi-arid with water sources</div>';
            } else {
                html += '<div>Terrain Type: Water-rich environment</div>';
            }
            
            return html;
        }

        // Render wind view with enhanced information
        function renderWind(wind) {
            let html = '<h3>🌬️ Wind System</h3>';
            html += '<div>Direction: ' + (wind.direction * 180 / Math.PI).toFixed(1) + '°</div>';
            html += '<div>Strength: ' + wind.strength.toFixed(2) + '</div>';
            html += '<div>Turbulence: ' + wind.turbulence_level.toFixed(2) + '</div>';
            html += '<div>Weather: ' + wind.weather_pattern + '</div>';
            html += '<div>Pollen Count: ' + wind.pollen_count + '</div>';
            
            // Add detailed analysis
            html += '<h4>🌪️ Wind Analysis:</h4>';
            const windDirection = (wind.direction * 180 / Math.PI + 360) % 360;
            let directionName = '';
            if (windDirection < 45 || windDirection >= 315) directionName = 'North';
            else if (windDirection < 135) directionName = 'East';
            else if (windDirection < 225) directionName = 'South';
            else directionName = 'West';
            
            html += '<div>Cardinal Direction: ' + directionName + '</div>';
            
            // Wind strength analysis
            if (wind.strength < 0.2) {
                html += '<div style="color: lightblue;">🌿 Gentle breeze - minimal pollen dispersal</div>';
            } else if (wind.strength < 0.5) {
                html += '<div style="color: yellow;">💨 Moderate wind - good for plant reproduction</div>';
            } else if (wind.strength < 0.8) {
                html += '<div style="color: orange;">🌪️ Strong wind - high pollen dispersal</div>';
            } else {
                html += '<div style="color: red;">⛈️ Storm conditions - disrupted ecosystem</div>';
            }
            
            // Pollen dispersal analysis
            html += '<h4>🌸 Pollen Dispersal:</h4>';
            if (wind.pollen_count === 0) {
                html += '<div>No active pollen dispersal</div>';
            } else if (wind.pollen_count < 10) {
                html += '<div style="color: lightgreen;">Low pollen activity</div>';
            } else if (wind.pollen_count < 30) {
                html += '<div style="color: yellow;">Moderate pollen dispersal</div>';
            } else {
                html += '<div style="color: orange;">High pollen activity - peak breeding season</div>';
            }
            
            return html;
        }
        
        // Control functions
        function togglePause() {
            isPaused = !isPaused;
            const btn = document.getElementById('pause-btn');
            btn.textContent = isPaused ? '▶ Resume' : '⏸ Pause';
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({action: 'toggle_pause'}));
            }
        }
        
        function resetSimulation() {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({action: 'reset'}));
            }
        }
        
        function saveState() {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({action: 'save_state'}));
            }
        }
        
        function loadState() {
            document.getElementById('load-file').click();
        }
        
        function handleFileLoad(event) {
            const file = event.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    try {
                        const stateData = JSON.parse(e.target.result);
                        if (ws && ws.readyState === WebSocket.OPEN) {
                            ws.send(JSON.stringify({
                                action: 'load_state',
                                data: stateData
                            }));
                        }
                    } catch (error) {
                        alert('Error loading file: Invalid JSON format');
                    }
                };
                reader.readAsText(file);
            }
        }
        
        // Initialize the interface
        window.onload = function() {
            initViewTabs();
            initTraitSliders();
            connect();
        };
        
        // Initialize trait sliders
        function initTraitSliders() {
            const traits = ['speed', 'aggression', 'cooperation', 'intelligence'];
            traits.forEach(trait => {
                const slider = document.getElementById(trait + '-trait');
                const valueSpan = document.getElementById(trait + '-value');
                if (slider && valueSpan) {
                    slider.oninput = function() {
                        valueSpan.textContent = parseFloat(this.value).toFixed(1);
                    };
                }
            });
        }
        
        // Player Management Functions
        function joinAsPlayer() {
            const nameInput = document.getElementById('player-name-input');
            const playerName = nameInput.value.trim();
            const errorDiv = document.getElementById('join-error');
            
            if (!playerName) {
                showError(errorDiv, 'Please enter your name');
                return;
            }
            
            if (!/^[a-zA-Z0-9\s]+$/.test(playerName)) {
                showError(errorDiv, 'Name can only contain letters, numbers, and spaces');
                return;
            }
            
            hideError(errorDiv);
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    action: 'join_as_player',
                    data: { name: playerName }
                }));
            }
        }
        
        function showCreateSpeciesForm() {
            document.getElementById('create-species-form').style.display = 'block';
            document.getElementById('control-species-form').style.display = 'none';
        }
        
        function hideCreateSpeciesForm() {
            document.getElementById('create-species-form').style.display = 'none';
            document.getElementById('species-name-input').value = '';
            resetTraitSliders();
        }
        
        function showControlSpeciesForm() {
            updateSpeciesSelect();
            document.getElementById('control-species-form').style.display = 'block';
            document.getElementById('create-species-form').style.display = 'none';
        }
        
        function hideControlSpeciesForm() {
            document.getElementById('control-species-form').style.display = 'none';
        }
        
        function resetTraitSliders() {
            const traits = ['speed', 'aggression', 'cooperation', 'intelligence'];
            traits.forEach(trait => {
                const slider = document.getElementById(trait + '-trait');
                const valueSpan = document.getElementById(trait + '-value');
                if (slider && valueSpan) {
                    slider.value = 0;
                    valueSpan.textContent = '0.0';
                }
            });
        }
        
        function createSpecies() {
            const nameInput = document.getElementById('species-name-input');
            const speciesName = nameInput.value.trim();
            const errorDiv = document.getElementById('create-species-error');
            
            if (!speciesName) {
                showError(errorDiv, 'Please enter a species name');
                return;
            }
            
            if (!/^[a-zA-Z0-9\s]+$/.test(speciesName)) {
                showError(errorDiv, 'Species name can only contain letters, numbers, and spaces');
                return;
            }
            
            hideError(errorDiv);
            
            // Collect trait adjustments
            const traits = {};
            ['speed', 'aggression', 'cooperation', 'intelligence'].forEach(trait => {
                const slider = document.getElementById(trait + '-trait');
                if (slider) {
                    traits[trait] = parseFloat(slider.value);
                }
            });
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    action: 'create_species',
                    data: { 
                        name: speciesName,
                        traits: traits
                    }
                }));
            }
        }
        
        function updateSpeciesSelect() {
            const select = document.getElementById('species-select');
            select.innerHTML = '<option value="">Select a species to control</option>';
            
            playerSpecies.forEach(species => {
                const option = document.createElement('option');
                option.value = species;
                option.textContent = species;
                select.appendChild(option);
            });
        }
        
        function executeMove() {
            const select = document.getElementById('species-select');
            const selectedSpecies = select.value;
            const errorDiv = document.getElementById('control-species-error');
            
            if (!selectedSpecies) {
                showError(errorDiv, 'Please select a species first');
                return;
            }
            
            if (!moveTarget) {
                showError(errorDiv, 'Please click on the grid to set a target location first');
                return;
            }
            
            hideError(errorDiv);
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    action: 'control_species',
                    data: {
                        species: selectedSpecies,
                        command: 'move',
                        x: moveTarget.x,
                        y: moveTarget.y
                    }
                }));
            }
        }
        
        function executeGather() {
            const select = document.getElementById('species-select');
            const selectedSpecies = select.value;
            const errorDiv = document.getElementById('control-species-error');
            
            if (!selectedSpecies) {
                showError(errorDiv, 'Please select a species first');
                return;
            }
            
            hideError(errorDiv);
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    action: 'control_species',
                    data: {
                        species: selectedSpecies,
                        command: 'gather'
                    }
                }));
            }
        }
        
        function executeReproduce() {
            const select = document.getElementById('species-select');
            const selectedSpecies = select.value;
            const errorDiv = document.getElementById('control-species-error');
            
            if (!selectedSpecies) {
                showError(errorDiv, 'Please select a species first');
                return;
            }
            
            hideError(errorDiv);
            
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    action: 'control_species',
                    data: {
                        species: selectedSpecies,
                        command: 'reproduce'
                    }
                }));
            }
        }
        
        function showError(errorDiv, message) {
            errorDiv.textContent = message;
            errorDiv.style.display = 'block';
        }
        
        function hideError(errorDiv) {
            errorDiv.style.display = 'none';
        }
        
        // Handle player-specific WebSocket messages
        function handlePlayerMessage(data) {
            switch (data.type) {
                case 'player_joined':
                    playerID = data.player_id;
                    document.getElementById('player-name').textContent = data.name;
                    document.getElementById('join-form').style.display = 'none';
                    document.getElementById('player-controls').style.display = 'block';
                    console.log('Player joined:', data.message);
                    break;
                    
                case 'species_created':
                    playerSpecies.push(data.species_name);
                    updatePlayerSpeciesCount();
                    hideCreateSpeciesForm();
                    console.log('Species created:', data.message);
                    break;
                    
                case 'command_executed':
                    console.log('Command executed:', data.message);
                    break;
                    
                case 'error':
                    console.error('Server error:', data.message);
                    alert('Error: ' + data.message);
                    break;
            }
        }
        
        function updatePlayerSpeciesCount() {
            document.getElementById('player-species-count').textContent = playerSpecies.length + ' species';
        }
        
        // Add grid click handling for movement
        function handleGridClick(event) {
            if (document.getElementById('control-species-form').style.display === 'block') {
                const gridContainer = document.getElementById('grid-view');
                const rect = gridContainer.getBoundingClientRect();
                const x = event.clientX - rect.left;
                const y = event.clientY - rect.top;
                
                // Convert pixel coordinates to world coordinates (simplified)
                const worldX = (x / rect.width) * 100; // Assuming world width is 100
                const worldY = (y / rect.height) * 100; // Assuming world height is 100
                
                moveTarget = { x: worldX, y: worldY };
                document.getElementById('move-target').textContent = '(' + worldX.toFixed(1) + ', ' + worldY.toFixed(1) + ')';
            }
        }
        
        // Render tools view
        function renderTools(tools) {
            let html = '<h3>🔧 Tool System</h3>';
            html += '<div>Total Tools: ' + tools.total_tools + '</div>';
            html += '<div>Owned Tools: ' + tools.owned_tools + '</div>';
            html += '<div>Dropped Tools: ' + tools.dropped_tools + '</div>';
            html += '<div>Average Durability: ' + tools.avg_durability.toFixed(2) + '</div>';
            html += '<div>Average Efficiency: ' + tools.avg_efficiency.toFixed(2) + '</div>';
            
            if (tools.tool_types && Object.keys(tools.tool_types).length > 0) {
                html += '<h4>Tool Types:</h4>';
                Object.entries(tools.tool_types).forEach(([type, count]) => {
                    html += '<div>' + type + ': ' + count + '</div>';
                });
            } else {
                html += '<br><div>No tools created yet</div>';
            }
            
            html += '<br><h4>Tool Usage:</h4>';
            if (tools.owned_tools === 0) {
                html += '<div>Usage Level: No tool use</div>';
            } else if (tools.owned_tools < 5) {
                html += '<div>Usage Level: Basic tool use</div>';
            } else {
                html += '<div>Usage Level: Advanced tool use</div>';
            }
            
            return html;
        }
        
        // Render environment view
        function renderEnvironment(envMod) {
            let html = '<h3>🏗️ Environmental Modification</h3>';
            html += '<div>Total Modifications: ' + envMod.total_modifications + '</div>';
            html += '<div>Active Modifications: ' + envMod.active_modifications + '</div>';
            html += '<div>Inactive Modifications: ' + envMod.inactive_modifications + '</div>';
            html += '<div>Average Durability: ' + envMod.avg_durability.toFixed(2) + '</div>';
            html += '<div>Tunnel Networks: ' + envMod.tunnel_networks + '</div>';
            
            if (envMod.modification_types && Object.keys(envMod.modification_types).length > 0) {
                html += '<h4>Modification Types:</h4>';
                Object.entries(envMod.modification_types).forEach(([type, count]) => {
                    html += '<div>' + type + ': ' + count + '</div>';
                });
            } else {
                html += '<br><div>No environmental modifications yet</div>';
            }
            
            html += '<br><h4>Modification Activity:</h4>';
            if (envMod.total_modifications === 0) {
                html += '<div>Activity Level: No modifications</div>';
            } else if (envMod.active_modifications < 3) {
                html += '<div>Activity Level: Low modification activity</div>';
            } else {
                html += '<div>Activity Level: High modification activity</div>';
            }
            
            return html;
        }
        
        // Render behavior view
        function renderBehavior(behavior) {
            let html = '<h3>🧠 Emergent Behavior</h3>';
            html += '<div>Total Entities: ' + behavior.total_entities + '</div>';
            html += '<div>Discovered Behaviors: ' + behavior.discovered_behaviors + '</div>';
            
            if (behavior.behavior_spread && Object.keys(behavior.behavior_spread).length > 0) {
                html += '<h4>Behavior Spread:</h4>';
                Object.entries(behavior.behavior_spread).forEach(([behavior_name, count]) => {
                    html += '<div>' + behavior_name + ': ' + count + ' entities</div>';
                });
            }
            
            if (behavior.avg_proficiency && Object.keys(behavior.avg_proficiency).length > 0) {
                html += '<h4>Average Proficiency:</h4>';
                Object.entries(behavior.avg_proficiency).forEach(([behavior_name, proficiency]) => {
                    html += '<div>' + behavior_name + ': ' + proficiency.toFixed(2) + '</div>';
                });
            }
            
            if (behavior.discovered_behaviors === 0) {
                html += '<br><div>No emergent behaviors discovered yet</div>';
            } else {
                html += '<br><h4>Behavior Development:</h4>';
                if (behavior.discovered_behaviors < 3) {
                    html += '<div>Development Level: Early behavior emergence</div>';
                } else if (behavior.discovered_behaviors < 8) {
                    html += '<div>Development Level: Moderate behavior complexity</div>';
                } else {
                    html += '<div>Development Level: Advanced behavioral evolution</div>';
                }
            }
            
            return html;
        }
        
        // Render statistical analysis view
        function renderStatistical(statistical) {
            if (!statistical) {
                return '<h3>📊 Statistical Analysis</h3><div>Statistical analysis not available</div>';
            }
            
            let html = '<h3>📊 Statistical Analysis</h3>';
            
            // Summary statistics
            html += '<h4>Summary Statistics:</h4>';
            html += '<div>Total Events: ' + statistical.total_events + '</div>';
            html += '<div>Total Snapshots: ' + statistical.total_snapshots + '</div>';
            html += '<div>Total Anomalies: ' + statistical.total_anomalies + '</div>';
            if (statistical.total_energy !== undefined) {
                html += '<div>Total Energy: ' + statistical.total_energy.toFixed(2) + '</div>';
            }
            if (statistical.energy_change !== undefined) {
                const changePercent = (statistical.energy_change * 100).toFixed(1);
                html += '<div>Energy Change: ' + changePercent + '% from baseline</div>';
            }
            
            // Trends
            html += '<h4>Trends:</h4>';
            html += '<div>Energy Trend: ' + (statistical.energy_trend || 'unknown') + '</div>';
            html += '<div>Population Trend: ' + (statistical.population_trend || 'unknown') + '</div>';
            
            // Recent events
            if (statistical.recent_events && statistical.recent_events.length > 0) {
                html += '<h4>Recent Events (' + statistical.recent_events.length + '):</h4>';
                html += '<div style="max-height: 200px; overflow-y: auto;">';
                statistical.recent_events.slice(0, 10).forEach(event => {
                    html += '<div style="margin: 5px 0; padding: 5px; background-color: #2a2a2a; border-radius: 3px;">';
                    html += '<strong>T' + event.tick + '</strong> ';
                    html += event.type + ' (' + event.target + ')';
                    if (event.change !== 0) {
                        html += ': ' + event.change.toFixed(4);
                    }
                    if (event.description) {
                        html += '<br><small>' + event.description + '</small>';
                    }
                    html += '</div>';
                });
                html += '</div>';
            }
            
            // Latest snapshot
            if (statistical.latest_snapshot) {
                const snapshot = statistical.latest_snapshot;
                html += '<h4>Latest Snapshot (T' + snapshot.tick + '):</h4>';
                html += '<div>Total Energy: ' + snapshot.total_energy.toFixed(2) + '</div>';
                html += '<div>Population Count: ' + snapshot.population_count + '</div>';
                
                if (snapshot.trait_averages && Object.keys(snapshot.trait_averages).length > 0) {
                    html += '<h5>Trait Averages:</h5>';
                    Object.entries(snapshot.trait_averages).forEach(([trait, avg]) => {
                        html += '<div>' + trait + ': ' + avg.toFixed(3) + '</div>';
                    });
                }
            }
            
            return html;
        }
        
        // Render anomalies detection view
        function renderAnomalies(anomalies) {
            if (!anomalies) {
                return '<h3>⚠️ Anomaly Detection</h3><div>Anomaly detection not available</div>';
            }
            
            let html = '<h3>⚠️ Anomaly Detection</h3>';
            
            if (anomalies.total_anomalies === 0) {
                html += '<div style="color: #4CAF50;">✅ No anomalies detected!</div>';
                html += '<div>The simulation appears to be running within expected parameters.</div>';
            } else {
                html += '<div>Found ' + anomalies.total_anomalies + ' anomalies:</div><br>';
                
                // Anomaly types summary
                if (anomalies.anomaly_types && Object.keys(anomalies.anomaly_types).length > 0) {
                    html += '<h4>Anomaly Types:</h4>';
                    Object.entries(anomalies.anomaly_types).forEach(([type, count]) => {
                        html += '<div>' + type.replace(/_/g, ' ') + ': ' + count + '</div>';
                    });
                    html += '<br>';
                }
                
                // Recent anomalies
                if (anomalies.recent_anomalies && anomalies.recent_anomalies.length > 0) {
                    html += '<h4>Recent Anomalies:</h4>';
                    html += '<div style="max-height: 300px; overflow-y: auto;">';
                    anomalies.recent_anomalies.slice(0, 10).forEach(anomaly => {
                        html += '<div style="margin: 10px 0; padding: 10px; background-color: #4a2a2a; border-radius: 5px; border-left: 3px solid #ff6b6b;">';
                        html += '<strong>🔍 ' + anomaly.type.replace(/_/g, ' ') + '</strong>';
                        html += ' (T' + anomaly.tick + ')';
                        html += '<br>' + anomaly.description;
                        html += '<br><small>Severity: ' + anomaly.severity.toFixed(2) + ', Confidence: ' + anomaly.confidence.toFixed(2) + '</small>';
                        html += '</div>';
                    });
                    html += '</div>';
                }
                
                // Recommendations
                if (anomalies.recommendations && anomalies.recommendations.length > 0) {
                    html += '<h4>Recommendations:</h4>';
                    anomalies.recommendations.forEach(rec => {
                        html += '<div style="margin: 5px 0; padding: 5px; background-color: #2a3a4a; border-radius: 3px; border-left: 3px solid #4CAF50;">• ' + rec + '</div>';
                    });
                }
            }
            
            return html;
        }
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleStatus provides a simple status endpoint
func (wi *WebInterface) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"tick":         wi.world.Tick,
		"entities":     len(wi.world.AllEntities),
		"plants":       len(wi.world.AllPlants),
		"populations":  len(wi.world.Populations),
		"status":       "running",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleWebSocket handles WebSocket connections
func (wi *WebInterface) handleWebSocket(ws *websocket.Conn) {
	defer ws.Close()
	
	// Add client to the list
	wi.clientsMutex.Lock()
	wi.clients[ws] = true
	wi.clientsMutex.Unlock()
	
	log.Printf("Client connected. Total clients: %d", len(wi.clients))
	
	// Send initial data
	viewData := wi.viewManager.GetCurrentViewData()
	wi.sendToClient(ws, viewData)
	
	// Listen for client messages
	for {
		var msg map[string]interface{}
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			break
		}
		
		// Handle client commands
		if action, ok := msg["action"].(string); ok {
			var data interface{}
			if d, exists := msg["data"]; exists {
				data = d
			}
			wi.handleClientAction(ws, action, data)
		}
	}
	
	// Clean up client connection
	wi.clientsMutex.Lock()
	delete(wi.clients, ws)
	if playerID, exists := wi.clientPlayers[ws]; exists {
		wi.playerManager.RemovePlayer(playerID)
		delete(wi.clientPlayers, ws)
	}
	wi.clientsMutex.Unlock()
	
	log.Printf("Client disconnected. Total clients: %d", len(wi.clients))
}

// handleClientAction processes actions from web clients
func (wi *WebInterface) handleClientAction(ws *websocket.Conn, action string, data interface{}) {
	switch action {
	case "join_as_player":
		wi.handlePlayerJoin(ws, data)
		
	case "create_species":
		wi.handleCreateSpecies(ws, data)
		
	case "control_species":
		wi.handleControlSpecies(ws, data)
		
	case "toggle_pause":
		wi.world.TogglePause()
		log.Printf("Client requested pause toggle - now paused: %v", wi.world.IsPaused())
		
	case "reset":
		log.Printf("Client requested reset")
		wi.world.Reset()
		// Reinitialize with default populations after reset
		wi.reinitializeWorld()
		
	case "save_state":
		log.Printf("Client requested state save")
		// Create state manager and save to default file
		stateManager := NewStateManager(wi.world)
		filename := fmt.Sprintf("web_save_%d.json", time.Now().Unix())
		err := stateManager.SaveToFile(filename)
		if err != nil {
			log.Printf("Error saving state: %v", err)
		} else {
			log.Printf("State saved to %s", filename)
		}
	
	case "load_state":
		log.Printf("Client requested state load")
		if stateData, ok := data.(map[string]interface{}); ok {
			// Create state manager and load from provided data
			stateManager := NewStateManager(wi.world)
			err := stateManager.LoadFromData(stateData)
			if err != nil {
				log.Printf("Error loading state: %v", err)
			} else {
				log.Printf("State loaded successfully")
			}
		} else {
			log.Printf("Invalid state data format")
		}
	}
}

// serveStatic serves static files (for future CSS/JS files)
func (wi *WebInterface) serveStatic(w http.ResponseWriter, r *http.Request) {
	// For now, just return 404 since we're embedding everything
	http.NotFound(w, r)
}

// simulationLoop runs the simulation and broadcasts updates
func (wi *WebInterface) simulationLoop() {
	ticker := time.NewTicker(wi.updateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Update the simulation
			wi.world.Update()
			
			// Get current view data
			viewData := wi.viewManager.GetCurrentViewData()
			
			// Send to broadcast channel (non-blocking)
			select {
			case wi.broadcastChan <- viewData:
			default:
				// Channel is full, skip this update
			}
			
		case <-wi.stopChan:
			return
		}
	}
}

// broadcastLoop handles broadcasting updates to all connected clients
func (wi *WebInterface) broadcastLoop() {
	for {
		select {
		case viewData := <-wi.broadcastChan:
			wi.broadcastToClients(viewData)
			
		case <-wi.stopChan:
			return
		}
	}
}

// broadcastToClients sends data to all connected WebSocket clients
func (wi *WebInterface) broadcastToClients(data *ViewData) {
	wi.clientsMutex.RLock()
	clients := make([]*websocket.Conn, 0, len(wi.clients))
	for client := range wi.clients {
		clients = append(clients, client)
	}
	wi.clientsMutex.RUnlock()
	
	// Send to each client
	for _, client := range clients {
		wi.sendToClient(client, data)
	}
}

// sendToClient sends data to a specific client
func (wi *WebInterface) sendToClient(ws *websocket.Conn, data *ViewData) {
	err := websocket.JSON.Send(ws, data)
	if err != nil {
		// Client disconnected, remove from list
		wi.clientsMutex.Lock()
		delete(wi.clients, ws)
		wi.clientsMutex.Unlock()
	}
}

// sendJSONToClient sends any JSON-serializable data to a specific client
func (wi *WebInterface) sendJSONToClient(ws *websocket.Conn, data interface{}) {
	err := websocket.JSON.Send(ws, data)
	if err != nil {
		// Client disconnected, remove from list
		wi.clientsMutex.Lock()
		delete(wi.clients, ws)
		wi.clientsMutex.Unlock()
	}
}

// Stop stops the web interface
func (wi *WebInterface) Stop() {
	close(wi.stopChan)
}

// reinitializeWorld reinitializes the world with default populations after reset
func (wi *WebInterface) reinitializeWorld() {
	// Add default populations back to the world
	populations := []PopulationConfig{
		{
			Name:    "Herbivores",
			Species: "herbivore",
			BaseTraits: map[string]float64{
				"size":               -0.5, // Smaller
				"speed":              0.3,  // Moderate speed
				"aggression":         -0.8, // Very peaceful
				"defense":            0.2,  // Some defense
				"cooperation":        0.6,  // Cooperative
				"intelligence":       0.1,  // Basic intelligence
				"endurance":          0.4,  // Good endurance
				"strength":           -0.2, // Weaker
				"aquatic_adaptation": -0.5, // Poor in water initially
				"digging_ability":    0.1,  // Basic digging
				"underground_nav":    -0.2, // Poor underground navigation initially
				"flying_ability":     -0.8, // Cannot fly initially
				"altitude_tolerance": -0.6, // Poor altitude tolerance initially
			},
			StartPos:         Position{X: 25, Y: 25},
			Spread:           15.0,
			Color:            "green",
			BaseMutationRate: 0.05, // Lower mutation rate - stable herbivores
		},
		{
			Name:    "Predators",
			Species: "predator",
			BaseTraits: map[string]float64{
				"size":               0.4,  // Larger
				"speed":              0.6,  // Fast
				"aggression":         0.8,  // Aggressive
				"defense":            0.4,  // Good defense
				"cooperation":        -0.3, // Mostly solitary
				"intelligence":       0.5,  // Higher intelligence
				"endurance":          0.2,  // Moderate endurance
				"strength":           0.7,  // Strong
				"aquatic_adaptation": -0.3, // Moderate water adaptation
				"digging_ability":    -0.1, // Some digging
				"underground_nav":    -0.4, // Poor underground navigation initially
				"flying_ability":     -0.6, // Cannot fly initially
				"altitude_tolerance": -0.4, // Poor altitude tolerance initially
			},
			StartPos:         Position{X: 75, Y: 75},
			Spread:           10.0,
			Color:            "red",
			BaseMutationRate: 0.08, // Moderate mutation rate - adaptive predators
		},
		{
			Name:    "Omnivores",
			Species: "omnivore",
			BaseTraits: map[string]float64{
				"size":               0.0,  // Medium size
				"speed":              0.4,  // Moderate speed
				"aggression":         0.1,  // Slightly aggressive
				"defense":            0.3,  // Moderate defense
				"cooperation":        0.2,  // Some cooperation
				"intelligence":       0.4,  // Good intelligence
				"endurance":          0.3,  // Good endurance
				"strength":           0.1,  // Moderate strength
				"aquatic_adaptation": 0.0,  // Neutral water adaptation
				"digging_ability":    0.2,  // Some digging
				"underground_nav":    0.0,  // Neutral underground navigation
				"flying_ability":     -0.5, // Cannot fly initially
				"altitude_tolerance": -0.2, // Poor altitude tolerance initially
			},
			StartPos:         Position{X: 50, Y: 20},
			Spread:           12.0,
			Color:            "blue",
			BaseMutationRate: 0.10, // Moderate mutation rate - adaptable
		},
	}

	// Add populations to the world
	for _, popConfig := range populations {
		wi.world.AddPopulation(popConfig)
	}
}

// handlePlayerJoin handles a player joining the game
func (wi *WebInterface) handlePlayerJoin(ws *websocket.Conn, data interface{}) {
	wi.clientsMutex.Lock()
	defer wi.clientsMutex.Unlock()
	
	// Parse player data
	playerData, ok := data.(map[string]interface{})
	if !ok {
		wi.sendErrorToClient(ws, "Invalid player data format")
		return
	}
	
	playerName, ok := playerData["name"].(string)
	if !ok {
		wi.sendErrorToClient(ws, "Player name is required")
		return
	}
	
	// Generate player ID (simple approach using connection address + timestamp)
	playerID := fmt.Sprintf("player_%d_%p", time.Now().UnixNano(), ws)
	
	// Add player
	player, err := wi.playerManager.AddPlayer(playerID, playerName)
	if err != nil {
		wi.sendErrorToClient(ws, fmt.Sprintf("Failed to add player: %v", err))
		return
	}
	
	// Map connection to player
	wi.clientPlayers[ws] = playerID
	
	log.Printf("Player '%s' joined with ID %s", player.Name, playerID)
	
	// Send success response
	response := map[string]interface{}{
		"type":      "player_joined",
		"player_id": playerID,
		"name":      player.Name,
		"message":   fmt.Sprintf("Welcome, %s! You can now create your own species.", player.Name),
	}
	wi.sendJSONToClient(ws, response)
}

// handleCreateSpecies handles a player creating a new species
func (wi *WebInterface) handleCreateSpecies(ws *websocket.Conn, data interface{}) {
	wi.clientsMutex.Lock()
	defer wi.clientsMutex.Unlock()
	
	// Get player ID for this connection
	playerID, exists := wi.clientPlayers[ws]
	if !exists {
		wi.sendErrorToClient(ws, "You must join as a player first")
		return
	}
	
	// Parse species data
	speciesData, ok := data.(map[string]interface{})
	if !ok {
		wi.sendErrorToClient(ws, "Invalid species data format")
		return
	}
	
	speciesName, ok := speciesData["name"].(string)
	if !ok {
		wi.sendErrorToClient(ws, "Species name is required")
		return
	}
	
	// Validate and clean species name
	cleanSpeciesName, err := ValidatePlayerName(speciesName)
	if err != nil {
		wi.sendErrorToClient(ws, fmt.Sprintf("Invalid species name: %v", err))
		return
	}
	
	// Check if species name already exists in the world
	if _, exists := wi.world.Populations[cleanSpeciesName]; exists {
		wi.sendErrorToClient(ws, "A species with this name already exists")
		return
	}
	
	// Create basic traits for the new species (limited control)
	baseTraits := map[string]float64{
		"size":               0.0,  // Neutral
		"speed":              0.1,  // Slightly above average
		"aggression":         -0.2, // Slightly peaceful
		"defense":            0.0,  // Neutral
		"cooperation":        0.1,  // Slightly cooperative
		"intelligence":       0.2,  // Slightly intelligent
		"endurance":          0.1,  // Slightly higher endurance
		"strength":           0.0,  // Neutral
		"aquatic_adaptation": -0.5, // Poor in water initially
		"digging_ability":    0.0,  // Neutral
		"underground_nav":    -0.3, // Poor underground navigation initially
		"flying_ability":     -0.8, // Cannot fly initially
		"altitude_tolerance": -0.6, // Poor altitude tolerance initially
	}
	
	// Allow players to make minor adjustments to some traits
	if adjustments, ok := speciesData["traits"].(map[string]interface{}); ok {
		// Only allow adjustment of certain traits within limits
		allowedTraits := []string{"speed", "aggression", "cooperation", "intelligence"}
		for _, traitName := range allowedTraits {
			if value, exists := adjustments[traitName]; exists {
				if floatValue, ok := value.(float64); ok {
					// Limit adjustments to ±0.3 range
					adjustment := math.Max(-0.3, math.Min(0.3, floatValue))
					baseTraits[traitName] += adjustment
					// Ensure final values stay within reasonable bounds
					baseTraits[traitName] = math.Max(-1.0, math.Min(1.0, baseTraits[traitName]))
				}
			}
		}
	}
	
	// Generate random starting position
	startX := rand.Float64() * wi.world.Config.Width
	startY := rand.Float64() * wi.world.Config.Height
	
	// Create population config
	popConfig := PopulationConfig{
		Name:             cleanSpeciesName,
		Species:          cleanSpeciesName,
		BaseTraits:       baseTraits,
		StartPos:         Position{X: startX, Y: startY},
		Spread:           15.0,
		Color:            "purple", // Player species get purple color
		BaseMutationRate: 0.08,     // Moderate mutation rate
	}
	
	// Add population to world
	wi.world.AddPopulation(popConfig)
	
	// Add species to player
	err = wi.playerManager.AddPlayerSpecies(playerID, cleanSpeciesName)
	if err != nil {
		wi.sendErrorToClient(ws, fmt.Sprintf("Failed to assign species to player: %v", err))
		return
	}
	
	// Update player activity
	wi.playerManager.UpdatePlayerActivity(playerID)
	
	log.Printf("Player %s created species '%s'", playerID, cleanSpeciesName)
	
	// Send success response
	response := map[string]interface{}{
		"type":         "species_created",
		"species_name": cleanSpeciesName,
		"message":      fmt.Sprintf("Successfully created species '%s'! You can now control its entities.", cleanSpeciesName),
		"traits":       baseTraits,
	}
	wi.sendJSONToClient(ws, response)
}

// handleControlSpecies handles player commands to control their species
func (wi *WebInterface) handleControlSpecies(ws *websocket.Conn, data interface{}) {
	wi.clientsMutex.Lock()
	defer wi.clientsMutex.Unlock()
	
	// Get player ID for this connection
	playerID, exists := wi.clientPlayers[ws]
	if !exists {
		wi.sendErrorToClient(ws, "You must join as a player first")
		return
	}
	
	// Parse control data
	controlData, ok := data.(map[string]interface{})
	if !ok {
		wi.sendErrorToClient(ws, "Invalid control data format")
		return
	}
	
	speciesName, ok := controlData["species"].(string)
	if !ok {
		wi.sendErrorToClient(ws, "Species name is required")
		return
	}
	
	// Check if player can control this species
	if !wi.playerManager.CanPlayerControlSpecies(playerID, speciesName) {
		wi.sendErrorToClient(ws, "You can only control your own species")
		return
	}
	
	// Get the population
	population, exists := wi.world.Populations[speciesName]
	if !exists {
		wi.sendErrorToClient(ws, "Species not found")
		return
	}
	
	// Parse control command
	command, ok := controlData["command"].(string)
	if !ok {
		wi.sendErrorToClient(ws, "Command is required")
		return
	}
	
	// Update player activity
	wi.playerManager.UpdatePlayerActivity(playerID)
	
	// Handle different control commands
	switch command {
	case "move":
		wi.handleMoveCommand(ws, playerID, population, controlData)
	case "gather":
		wi.handleGatherCommand(ws, playerID, population, controlData)
	case "reproduce":
		wi.handleReproduceCommand(ws, playerID, population, controlData)
	default:
		wi.sendErrorToClient(ws, fmt.Sprintf("Unknown command: %s", command))
	}
}

// handleMoveCommand handles movement commands for player species
func (wi *WebInterface) handleMoveCommand(ws *websocket.Conn, playerID string, population *Population, controlData map[string]interface{}) {
	// Parse movement parameters
	targetX, xOk := controlData["x"].(float64)
	targetY, yOk := controlData["y"].(float64)
	
	if !xOk || !yOk {
		wi.sendErrorToClient(ws, "Target coordinates (x, y) are required for movement")
		return
	}
	
	// Ensure coordinates are within world bounds
	targetX = math.Max(0, math.Min(wi.world.Config.Width, targetX))
	targetY = math.Max(0, math.Min(wi.world.Config.Height, targetY))
	
	// Apply movement tendency to all entities in the population
	moveCount := 0
	for _, entity := range population.Entities {
		if entity.IsAlive {
			// Calculate direction to target
			dx := targetX - entity.Position.X
			dy := targetY - entity.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)
			
			if distance > 1.0 { // Only move if not already close
				// Normalize direction and apply movement based on speed trait
				speed := entity.Traits["speed"].Value
				moveDistance := (0.5 + speed*0.5) * 2.0 // Base movement + speed modifier
				
				entity.Position.X += (dx / distance) * moveDistance
				entity.Position.Y += (dy / distance) * moveDistance
				
				// Ensure entity stays within bounds
				entity.Position.X = math.Max(0, math.Min(wi.world.Config.Width, entity.Position.X))
				entity.Position.Y = math.Max(0, math.Min(wi.world.Config.Height, entity.Position.Y))
				
				moveCount++
			}
		}
	}
	
	log.Printf("Player %s moved %d entities towards (%f, %f)", playerID, moveCount, targetX, targetY)
	
	// Send response
	response := map[string]interface{}{
		"type":        "command_executed",
		"command":     "move",
		"entities_affected": moveCount,
		"message":     fmt.Sprintf("Moved %d entities towards target location", moveCount),
	}
	wi.sendJSONToClient(ws, response)
}

// handleGatherCommand handles gathering commands for player species
func (wi *WebInterface) handleGatherCommand(ws *websocket.Conn, playerID string, population *Population, controlData map[string]interface{}) {
	gatherCount := 0
	
	// Apply gathering behavior to all entities in the population
	for _, entity := range population.Entities {
		if entity.IsAlive && entity.Energy < 80 { // Only gather if not at high energy
			// Find nearby plants to gather from
			for _, plant := range wi.world.AllPlants {
				if plant.IsAlive {
					// Calculate distance to plant
					dx := plant.Position.X - entity.Position.X
					dy := plant.Position.Y - entity.Position.Y
					distance := math.Sqrt(dx*dx + dy*dy)
					
					if distance <= 5.0 { // Within gathering range
						// Entity gains energy, plant loses energy
						energyGained := math.Min(10.0, plant.Energy)
						entity.Energy += energyGained
						plant.Energy -= energyGained
						
						if plant.Energy <= 0 {
							plant.IsAlive = false
						}
						
						gatherCount++
						break // Only gather from one plant per entity
					}
				}
			}
		}
	}
	
	log.Printf("Player %s performed gathering with %d entities", playerID, gatherCount)
	
	// Send response
	response := map[string]interface{}{
		"type":             "command_executed",
		"command":          "gather",
		"entities_affected": gatherCount,
		"message":          fmt.Sprintf("%d entities performed gathering actions", gatherCount),
	}
	wi.sendJSONToClient(ws, response)
}

// handleReproduceCommand handles reproduction commands for player species
func (wi *WebInterface) handleReproduceCommand(ws *websocket.Conn, playerID string, population *Population, controlData map[string]interface{}) {
	reproductionCount := 0
	
	// Find entities with sufficient energy for reproduction
	for _, entity := range population.Entities {
		if entity.IsAlive && entity.Energy > 70 { // High energy threshold for reproduction
			// Find nearby entities of the same species for mating
			for _, mate := range population.Entities {
				if mate.IsAlive && mate.ID != entity.ID && mate.Energy > 70 {
					// Calculate distance to potential mate
					dx := mate.Position.X - entity.Position.X
					dy := mate.Position.Y - entity.Position.Y
					distance := math.Sqrt(dx*dx + dy*dy)
					
					if distance <= 3.0 { // Within mating range
						// Create offspring through the existing reproduction system
						offspring := wi.world.CreateOffspring(entity, mate)
						if offspring != nil {
							// Add offspring to population
							population.Entities = append(population.Entities, offspring)
							wi.world.AllEntities = append(wi.world.AllEntities, offspring)
							
							// Reduce parent energy
							entity.Energy -= 30
							mate.Energy -= 30
							
							reproductionCount++
						}
						break // Only one reproduction per entity per command
					}
				}
			}
		}
	}
	
	log.Printf("Player %s triggered reproduction resulting in %d new entities", playerID, reproductionCount)
	
	// Send response
	response := map[string]interface{}{
		"type":      "command_executed",
		"command":   "reproduce",
		"offspring": reproductionCount,
		"message":   fmt.Sprintf("Reproduction successful! %d new entities born", reproductionCount),
	}
	wi.sendJSONToClient(ws, response)
}

// sendErrorToClient sends an error message to a specific client
func (wi *WebInterface) sendErrorToClient(ws *websocket.Conn, message string) {
	errorResponse := map[string]interface{}{
		"type":    "error",
		"message": message,
	}
	wi.sendJSONToClient(ws, errorResponse)
}