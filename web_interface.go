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
	webInterface := &WebInterface{
		world:          world,
		viewManager:    NewViewManager(world),
		clients:        make(map[*websocket.Conn]bool),
		broadcastChan:  make(chan *ViewData, 100),
		stopChan:       make(chan bool),
		updateInterval: 100 * time.Millisecond, // 10 FPS
		playerManager:  NewPlayerManager(),
		clientPlayers:  make(map[*websocket.Conn]string),
	}
	
	// Set up player events callback
	world.PlayerEventsCallback = webInterface.handlePlayerEvent
	
	return webInterface
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
	http.HandleFunc("/api/export/events", webInterface.handleExportEvents)
	http.HandleFunc("/api/export/analysis", webInterface.handleExportAnalysis)
	http.HandleFunc("/api/export/anomalies", webInterface.handleExportAnomalies)
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
        
        /* Warfare view styles */
        .colony-item {
            margin: 8px 0;
            padding: 8px;
            background-color: #4a4a4a;
            border-radius: 3px;
            border-left: 3px solid #6a9bd2;
        }
        
        .conflict-item {
            margin: 8px 0;
            padding: 8px;
            background-color: #4a4a4a;
            border-radius: 3px;
            border-left: 3px solid #F44336;
        }
        
        .event-item {
            margin: 4px 0;
            padding: 4px 8px;
            background-color: #4a4a4a;
            border-radius: 2px;
            font-size: 13px;
        }
        
        .colony-list, .events-list {
            max-height: 300px;
            overflow-y: auto;
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
            'REPRODUCTION', 'STATISTICAL', 'ANOMALIES', 'WARFARE', 'FUNGAL'
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
                if (data.type && ['player_joined', 'species_created', 'command_executed', 'species_extinct', 'subspecies_formed', 'new_species_detected', 'error'].includes(data.type)) {
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
            if (data.stats.avg_age_percent !== undefined) {
                // Use percentage of lifespan for better representation
                document.getElementById('avg-age').textContent = data.stats.avg_age_percent.toFixed(1) + '% of lifespan';
            } else if (data.stats.avg_age !== undefined) {
                // Fallback to raw age if percentage not available
                document.getElementById('avg-age').textContent = data.stats.avg_age.toFixed(1) + ' ticks';
            }
            
            // Update populations
            let populationsHtml = '';
            // Sort populations by name for consistent ordering
            const sortedPops = [...data.populations].sort((a, b) => a.name.localeCompare(b.name));
            sortedPops.forEach(pop => {
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
                    
                case 'REPRODUCTION':
                    viewContent.innerHTML = '<div class="stats-section">' + renderReproduction(data.reproduction) + '</div>';
                    break;
                    
                case 'STATISTICAL':
                    viewContent.innerHTML = '<div class="stats-section">' + renderStatistical(data.statistical) + '</div>';
                    break;
                    
                case 'ANOMALIES':
                    viewContent.innerHTML = '<div class="stats-section">' + renderAnomalies(data.anomalies) + '</div>';
                    break;
                    
                case 'WARFARE':
                    viewContent.innerHTML = '<div class="stats-section">' + renderWarfare(data.warfare) + '</div>';
                    break;
                    
                case 'FUNGAL':
                    viewContent.innerHTML = '<div class="stats-section">' + renderFungal(data.fungal) + '</div>';
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
        
        // Setup event delegation for species modal interactions
        let speciesModalEventsSetup = false;
        
        function setupSpeciesModalEvents() {
            // Only setup once to avoid multiple event listeners
            if (speciesModalEventsSetup) {
                return;
            }
            speciesModalEventsSetup = true;
            
            // Use event delegation on the document to handle clicks on species items
            document.addEventListener('click', function(event) {
                const speciesItem = event.target.closest('.clickable-species');
                if (speciesItem) {
                    const speciesName = speciesItem.getAttribute('data-species-name');
                    if (speciesName) {
                        showSpeciesDetail(speciesName);
                    }
                }
            });
        }
        
        // Create modal elements once when needed
        function ensureSpeciesModalExists() {
            const existingModal = document.getElementById('species-detail-modal');
            if (!existingModal) {
                // Create modal
                const modal = document.createElement('div');
                modal.id = 'species-detail-modal';
                modal.style.cssText = 'display: none; position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background-color: #1a1a1a; border: 2px solid #444; border-radius: 10px; padding: 20px; max-width: 80%; max-height: 80%; overflow-y: auto; z-index: 1000;';
                
                const closeButtonHtml = '<div style="text-align: right;"><button id="species-modal-close" style="background-color: #666; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer;">✕ Close</button></div>';
                const contentHtml = '<div id="species-detail-content"></div>';
                modal.innerHTML = closeButtonHtml + contentHtml;
                
                // Create overlay
                const overlay = document.createElement('div');
                overlay.id = 'species-modal-overlay';
                overlay.style.cssText = 'display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.7); z-index: 999;';
                
                // Add close event listeners
                const closeButton = modal.querySelector('#species-modal-close');
                closeButton.addEventListener('click', hideSpeciesDetail);
                overlay.addEventListener('click', hideSpeciesDetail);
                
                // Add to body
                document.body.appendChild(modal);
                document.body.appendChild(overlay);
            }
        }
        
        // Render species view with enhanced details and individual visualization
        function renderSpecies(species) {
            let html = '<h3>🐾 Species Tracking & Individual Visualization</h3>';
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
            
            // Individual Species Visualization Section
            html += '<h4>🔍 Individual Species Visualization:</h4>';
            html += '<div style="margin: 10px 0; padding: 10px; background-color: #2a2a2a; border-radius: 5px;">';
            html += 'Click on any species below to see detailed visual representation of what it looks like based on its genetic traits.';
            html += '</div>';
            
            if (species.species_details && species.species_details.length > 0) {
                html += '<h4>Species Gallery:</h4>';
                // Sort by population for better display
                const sortedSpecies = [...species.species_details].sort((a, b) => {
                    if (a.is_extinct !== b.is_extinct) {
                        return a.is_extinct ? 1 : -1; // Active species first
                    }
                    return b.population - a.population; // Higher population first
                });
                
                sortedSpecies.forEach(detail => {
                    html += '<div class="species-item clickable-species" data-species-name="' + detail.name.replace(/"/g, '&quot;') + '" style="cursor: pointer; padding: 8px; margin: 5px 0; background-color: #333; border-radius: 3px; border-left: 4px solid ' + (detail.is_extinct ? '#ff4444' : '#44ff44') + ';">';
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
                    html += '<div style="font-size: 0.8em; color: #ccc; margin-top: 3px;">Click to view detailed visualization →</div>';
                    html += '</div>';
                });
            } else {
                html += '<br><div>No species data available</div>';
            }
            
            return html;
        }
        
        // Show detailed species visualization
        function showSpeciesDetail(speciesName) {
            // Ensure modal elements exist
            ensureSpeciesModalExists();
            
            const modal = document.getElementById('species-detail-modal');
            const overlay = document.getElementById('species-modal-overlay');
            const content = document.getElementById('species-detail-content');
            
            if (!modal || !overlay || !content) {
                console.error('Species modal elements not found');
                return;
            }
            
            // Create detailed visualization for the species
            let detailHtml = '<h2>🦠 ' + speciesName + ' - Individual Visualization</h2>';
            
            // Simulated trait-based visualization
            detailHtml += renderSpeciesVisualization(speciesName);
            
            content.innerHTML = detailHtml;
            modal.style.display = 'block';
            overlay.style.display = 'block';
        }
        
        // Hide species detail modal
        function hideSpeciesDetail() {
            const modal = document.getElementById('species-detail-modal');
            const overlay = document.getElementById('species-modal-overlay');
            
            if (modal) {
                modal.style.display = 'none';
            }
            if (overlay) {
                overlay.style.display = 'none';
            }
        }
        
        // Render individual species visualization based on traits
        function renderSpeciesVisualization(speciesName) {
            let html = '';
            
            // Create a visual representation based on the species name patterns
            html += '<div style="display: flex; gap: 20px; flex-wrap: wrap;">';
            
            // Species Profile View
            html += '<div style="flex: 1; min-width: 300px;">';
            html += '<h3>🎨 Species Profile View</h3>';
            html += '<div style="text-align: center; background-color: #2a2a2a; padding: 20px; border-radius: 10px;">';
            
            // Generate visual representation based on species characteristics
            const visual = generateSpeciesVisual(speciesName);
            html += visual.profile;
            
            html += '</div>';
            html += '</div>';
            
            // Trait Analysis
            html += '<div style="flex: 1; min-width: 300px;">';
            html += '<h3>📊 Genetic Trait Analysis</h3>';
            html += '<div style="background-color: #2a2a2a; padding: 15px; border-radius: 10px;">';
            html += visual.traits;
            html += '</div>';
            html += '</div>';
            
            html += '</div>';
            
            // Cellular View
            html += '<div style="margin-top: 20px;">';
            html += '<h3>🔬 Cellular Structure View</h3>';
            html += '<div style="background-color: #2a2a2a; padding: 15px; border-radius: 10px;">';
            html += visual.cellular;
            html += '</div>';
            html += '</div>';
            
            // Habitat Information
            html += '<div style="margin-top: 20px;">';
            html += '<h3>🌍 Environmental Adaptation</h3>';
            html += '<div style="background-color: #2a2a2a; padding: 15px; border-radius: 10px;">';
            html += visual.habitat;
            html += '</div>';
            html += '</div>';
            
            return html;
        }
        
        // Generate visual representation for a species
        function generateSpeciesVisual(speciesName) {
            // Extract characteristics from species name patterns
            const isGrass = speciesName.toLowerCase().includes('grass');
            const isTree = speciesName.toLowerCase().includes('tree');
            const isBush = speciesName.toLowerCase().includes('bush');
            const isAlgae = speciesName.toLowerCase().includes('algae');
            const isCactus = speciesName.toLowerCase().includes('cactus');
            const isMushroom = speciesName.toLowerCase().includes('mushroom');
            
            // Generate pseudo-random traits based on name hash
            const hash = stringHash(speciesName);
            const traits = {
                size: 0.3 + (hash % 100) / 100 * 0.7,
                defense: 0.1 + (hash % 80) / 100 * 0.6,
                toxicity: 0.0 + (hash % 60) / 100 * 0.8,
                growth: 0.2 + (hash % 90) / 100 * 0.6,
                hardiness: 0.1 + (hash % 85) / 100 * 0.7
            };
            
            // Adjust traits based on type
            if (isTree) {
                traits.size += 0.3;
                traits.defense += 0.2;
            } else if (isGrass) {
                traits.growth += 0.3;
                traits.size -= 0.2;
            } else if (isCactus) {
                traits.defense += 0.4;
                traits.toxicity += 0.3;
            } else if (isMushroom) {
                traits.toxicity += 0.5;
                traits.defense -= 0.1;
            }
            
            const profile = generateProfileView(speciesName, traits);
            const traitBars = generateTraitBars(traits);
            const cellular = generateCellularView(speciesName, traits);
            const habitat = generateHabitatView(speciesName, traits);
            
            return {
                profile: profile,
                traits: traitBars,
                cellular: cellular,
                habitat: habitat
            };
        }
        
        // Generate profile view visualization with authentic Minecraft/Rimworld style
        function generateProfileView(name, traits) {
            let html = '';
            
            // True blocky, pixelated representation like Minecraft blocks
            const sizeBlocks = Math.max(3, Math.floor(traits.size * 6) + 2); // 3-8 blocks
            const defenseLevel = Math.floor(traits.defense * 4); // 0-3 defense levels
            const blockSize = 16; // Standard block size like Minecraft
            
            html += '<div style="margin: 20px 0; text-align: center;">';
            html += '<div style="display: inline-block; position: relative;">';
            
            // Create a grid-based organism using individual blocks
            const totalSize = sizeBlocks * blockSize;
            html += '<div style="display: inline-grid; grid-template-columns: repeat(' + sizeBlocks + ', ' + blockSize + 'px); ';
            html += 'grid-template-rows: repeat(' + sizeBlocks + ', ' + blockSize + 'px); gap: 0; ';
            html += 'image-rendering: pixelated; image-rendering: -moz-crisp-edges; image-rendering: crisp-edges;">';
            
            // Generate organism body as a collection of blocks
            const centerX = Math.floor(sizeBlocks / 2);
            const centerY = Math.floor(sizeBlocks / 2);
            
            for (let y = 0; y < sizeBlocks; y++) {
                for (let x = 0; x < sizeBlocks; x++) {
                    const distFromCenter = Math.abs(x - centerX) + Math.abs(y - centerY);
                    const isBodyBlock = distFromCenter <= Math.floor(sizeBlocks / 2);
                    const isCore = distFromCenter <= Math.floor(sizeBlocks / 4);
                    const isEdge = (x === 0 || x === sizeBlocks - 1 || y === 0 || y === sizeBlocks - 1) && isBodyBlock;
                    
                    let blockColor = '#222'; // Empty space
                    let topLight = 'rgba(255,255,255,0.1)';
                    let bottomShadow = 'rgba(0,0,0,0.3)';
                    
                    if (isBodyBlock) {
                        // Main body blocks
                        if (isCore) {
                            blockColor = traits.toxicity > 0.4 ? '#AA3333' : '#4488FF'; // Core is blue or toxic red
                            topLight = 'rgba(255,255,255,0.4)';
                        } else {
                            blockColor = traits.size > 0.6 ? '#3377CC' : '#2266BB'; // Body varies by size
                            topLight = 'rgba(255,255,255,0.3)';
                        }
                        
                        // Defense armor blocks
                        if (defenseLevel > 0 && isEdge) {
                            blockColor = '#BB7722'; // Orange armor blocks
                            topLight = 'rgba(255,255,255,0.4)';
                        }
                        
                        // Growth/fertility indicators
                        if (traits.growth > 0.6 && (x + y) % 2 === 0 && !isCore) {
                            blockColor = '#44AA44'; // Green growth blocks
                        }
                    }
                    
                    // Create individual block with Minecraft-style shading
                    html += '<div style="width: ' + blockSize + 'px; height: ' + blockSize + 'px; ';
                    html += 'background: ' + blockColor + '; ';
                    html += 'box-shadow: ';
                    html += 'inset 0 ' + Math.floor(blockSize/4) + 'px 0 ' + topLight + ', '; // Top highlight
                    html += 'inset 0 -' + Math.floor(blockSize/4) + 'px 0 ' + bottomShadow + ', '; // Bottom shadow
                    html += 'inset ' + Math.floor(blockSize/4) + 'px 0 0 rgba(255,255,255,0.2), '; // Left highlight
                    html += 'inset -' + Math.floor(blockSize/4) + 'px 0 0 rgba(0,0,0,0.2); '; // Right shadow
                    html += 'image-rendering: pixelated; ';
                    html += 'border: 0;"></div>';
                }
            }
            html += '</div>';
            
            // Add toxicity skull overlay if toxic
            if (traits.toxicity > 0.5) {
                html += '<div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); ';
                html += 'font-size: ' + Math.floor(totalSize / 4) + 'px; z-index: 10;">💀</div>';
            }
            
            // Add defensive spikes as separate blocks
            if (defenseLevel > 1) {
                const spikeOffset = totalSize / 2 + blockSize / 2;
                // Top spike
                html += '<div style="position: absolute; top: -' + blockSize + 'px; left: 50%; transform: translateX(-50%); ';
                html += 'width: ' + blockSize + 'px; height: ' + blockSize + 'px; ';
                html += 'background: #BB7722; ';
                html += 'box-shadow: inset 0 4px 0 rgba(255,255,255,0.4), inset 0 -4px 0 rgba(0,0,0,0.3); ';
                html += 'image-rendering: pixelated;"></div>';
                
                // Side spikes for high defense
                if (defenseLevel > 2) {
                    html += '<div style="position: absolute; top: 50%; left: -' + blockSize + 'px; transform: translateY(-50%); ';
                    html += 'width: ' + blockSize + 'px; height: ' + blockSize + 'px; ';
                    html += 'background: #BB7722; ';
                    html += 'box-shadow: inset 0 4px 0 rgba(255,255,255,0.4), inset 0 -4px 0 rgba(0,0,0,0.3); ';
                    html += 'image-rendering: pixelated;"></div>';
                    
                    html += '<div style="position: absolute; top: 50%; right: -' + blockSize + 'px; transform: translateY(-50%); ';
                    html += 'width: ' + blockSize + 'px; height: ' + blockSize + 'px; ';
                    html += 'background: #BB7722; ';
                    html += 'box-shadow: inset 0 4px 0 rgba(255,255,255,0.4), inset 0 -4px 0 rgba(0,0,0,0.3); ';
                    html += 'image-rendering: pixelated;"></div>';
                }
            }
            
            html += '</div>';
            html += '<div style="font-size: 0.8em; margin-top: 15px; color: #aaa; font-family: monospace; font-weight: bold;">' + name + '</div>';
            html += '</div>';
            
            return html;
        }
        
        // Helper function to get trait-based colors
        function getTraitColor(value, baseColor, intenseColor) {
            const intensity = Math.max(0, Math.min(1, value));
            return baseColor; // Simplified for now
        }
        
        // Generate blocky trait bars Minecraft-style
        function generateBlockyTraitBars(traits) {
            let html = '';
            
            Object.entries(traits).forEach(([trait, value]) => {
                const blockCount = Math.floor(value * 10);
                const percentage = (value * 100).toFixed(0);
                
                html += '<div style="margin: 8px 0; font-family: monospace;">';
                html += '<div style="display: flex; justify-content: space-between; margin-bottom: 3px;">';
                html += '<span style="text-transform: capitalize; font-weight: bold;">' + trait + '</span>';
                html += '<span style="color: #aaa;">' + percentage + '%</span>';
                html += '</div>';
                
                // Create blocky progress bar
                html += '<div style="display: flex; gap: 1px; height: 16px;">';
                for (let i = 0; i < 10; i++) {
                    const isActive = i < blockCount;
                    const blockColor = isActive ? getTraitBlockColor(trait) : '#333';
                    html += '<div style="width: 16px; height: 16px; background: ' + blockColor + '; ';
                    html += 'border: 1px solid #555; image-rendering: pixelated; ';
                    html += 'box-shadow: inset 1px 1px 0 rgba(255,255,255,0.2), inset -1px -1px 0 rgba(0,0,0,0.3);"></div>';
                }
                html += '</div>';
                html += '</div>';
            });
            
            return html;
        }
        
        // Get color for trait blocks
        function getTraitBlockColor(trait) {
            const colors = {
                'size': '#4a9eff',
                'defense': '#ff9944', 
                'toxicity': '#ff4444',
                'growth': '#44ff44',
                'hardiness': '#9944ff',
                'fertility': '#ff44ff'
            };
            return colors[trait] || '#44ff44';
        }
        
        // Generate trait bars with authentic Minecraft-style design
        function generateTraitBars(traits) {
            let html = '';
            
            Object.entries(traits).forEach(([trait, value]) => {
                const blockCount = Math.floor(value * 10);
                const percentage = (value * 100).toFixed(1);
                
                html += '<div style="margin: 10px 0; font-family: monospace;">';
                html += '<div style="display: flex; justify-content: space-between; margin-bottom: 5px;">';
                html += '<span style="text-transform: capitalize; font-weight: bold; color: #fff; text-shadow: 1px 1px 0 #000;">' + trait + '</span>';
                html += '<span style="color: #aaa; font-weight: bold; text-shadow: 1px 1px 0 #000;">' + percentage + '%</span>';
                html += '</div>';
                
                // Create authentic Minecraft-style progress bar with blocks
                html += '<div style="display: flex; gap: 1px; background: #2a2a2a; padding: 3px; border: 2px inset #444; border-radius: 2px;">';
                for (let i = 0; i < 10; i++) {
                    const isActive = i < blockCount;
                    const blockColor = isActive ? getMinecraftTraitColor(trait) : '#1a1a1a';
                    
                    // Create authentic Minecraft block with proper lighting
                    html += '<div style="width: 18px; height: 18px; ';
                    html += 'background: ' + blockColor + '; ';
                    if (isActive) {
                        html += 'box-shadow: ';
                        html += 'inset 0 4px 0 rgba(255,255,255,0.4), '; // Top highlight
                        html += 'inset 0 -4px 0 rgba(0,0,0,0.4), '; // Bottom shadow  
                        html += 'inset 4px 0 0 rgba(255,255,255,0.2), '; // Left highlight
                        html += 'inset -4px 0 0 rgba(0,0,0,0.2); '; // Right shadow
                    } else {
                        html += 'box-shadow: ';
                        html += 'inset 0 2px 0 rgba(255,255,255,0.1), '; // Subtle top
                        html += 'inset 0 -2px 0 rgba(0,0,0,0.3); '; // Subtle bottom
                    }
                    html += 'border: none; ';
                    html += 'image-rendering: pixelated; ';
                    html += 'image-rendering: -moz-crisp-edges; ';
                    html += 'image-rendering: crisp-edges;"></div>';
                }
                html += '</div>';
                html += '</div>';
            });
            
            return html;
        }
        
        // Get authentic Minecraft-style colors for trait blocks
        function getMinecraftTraitColor(trait) {
            const colors = {
                'size': '#3366CC',      // Blue like lapis lazuli blocks
                'defense': '#CC6600',   // Orange like copper blocks
                'toxicity': '#CC3333',  // Red like redstone blocks
                'growth': '#33AA33',    // Green like emerald blocks
                'hardiness': '#6633CC', // Purple like amethyst blocks
                'fertility': '#CC33AA'  // Pink like coral blocks
            };
            return colors[trait] || '#33AA33';
        }
        
        // Generate cellular view with blocky, Minecraft-style representation
        // Generate cellular view with authentic Minecraft-style blocks
        function generateCellularView(name, traits) {
            let html = '';
            
            html += '<div><strong>Cellular Structure (Minecraft-Style Blocks):</strong></div>';
            html += '<div style="margin: 15px 0; background: #1a1a1a; padding: 20px; border: 3px solid #333; image-rendering: pixelated;">';
            
            // Create authentic blocky cell structure using grid
            const cellGridSize = Math.max(6, Math.floor(traits.size * 8) + 4); // 6-12 blocks
            const blockSize = 12;
            
            html += '<div style="display: inline-grid; ';
            html += 'grid-template-columns: repeat(' + cellGridSize + ', ' + blockSize + 'px); ';
            html += 'grid-template-rows: repeat(' + cellGridSize + ', ' + blockSize + 'px); ';
            html += 'gap: 1px; ';
            html += 'background: #000; ';
            html += 'padding: 2px; ';
            html += 'border: 2px solid #444; ';
            html += 'image-rendering: pixelated; ';
            html += 'image-rendering: -moz-crisp-edges; ';
            html += 'image-rendering: crisp-edges;">';
            
            // Generate cell as blocks
            const centerX = Math.floor(cellGridSize / 2);
            const centerY = Math.floor(cellGridSize / 2);
            
            for (let y = 0; y < cellGridSize; y++) {
                for (let x = 0; x < cellGridSize; x++) {
                    const distFromCenter = Math.abs(x - centerX) + Math.abs(y - centerY);
                    const isMembraneBlock = distFromCenter === Math.floor(cellGridSize / 2);
                    const isInteriorBlock = distFromCenter < Math.floor(cellGridSize / 2);
                    const isNucleusBlock = distFromCenter <= 1;
                    
                    let blockColor = '#000'; // Background
                    let topHighlight = 'rgba(255,255,255,0.1)';
                    let bottomShadow = 'rgba(0,0,0,0.5)';
                    
                    if (isMembraneBlock) {
                        // Cell membrane - brown/orange blocks
                        blockColor = '#8B4513';
                        topHighlight = 'rgba(255,255,255,0.3)';
                        bottomShadow = 'rgba(0,0,0,0.4)';
                    } else if (isNucleusBlock) {
                        // Nucleus - blue blocks  
                        blockColor = '#3366CC';
                        topHighlight = 'rgba(255,255,255,0.4)';
                        bottomShadow = 'rgba(0,0,0,0.3)';
                    } else if (isInteriorBlock) {
                        // Cytoplasm with organelles
                        if (traits.toxicity > 0.4 && (x + y) % 4 === 0) {
                            // Toxin organelles - red blocks
                            blockColor = '#CC3333';
                            topHighlight = 'rgba(255,255,255,0.3)';
                        } else if (traits.defense > 0.4 && (x + y) % 5 === 0) {
                            // Defense organelles - orange blocks
                            blockColor = '#CC6600';
                            topHighlight = 'rgba(255,255,255,0.3)';
                        } else if (traits.growth > 0.5 && (x + y) % 3 === 0) {
                            // Growth organelles - green blocks
                            blockColor = '#33AA33';
                            topHighlight = 'rgba(255,255,255,0.3)';
                        } else if (traits.size > 0.5 && (x + y) % 6 === 0) {
                            // Mitochondria - yellow blocks
                            blockColor = '#CCAA33';
                            topHighlight = 'rgba(255,255,255,0.3)';
                        } else {
                            // Basic cytoplasm - gray blocks
                            blockColor = '#555555';
                            topHighlight = 'rgba(255,255,255,0.2)';
                        }
                        bottomShadow = 'rgba(0,0,0,0.3)';
                    }
                    
                    // Create individual block with Minecraft lighting
                    if (blockColor !== '#000') {
                        html += '<div style="width: ' + blockSize + 'px; height: ' + blockSize + 'px; ';
                        html += 'background: ' + blockColor + '; ';
                        html += 'box-shadow: ';
                        html += 'inset 0 ' + Math.floor(blockSize/3) + 'px 0 ' + topHighlight + ', ';
                        html += 'inset 0 -' + Math.floor(blockSize/3) + 'px 0 ' + bottomShadow + ', ';
                        html += 'inset ' + Math.floor(blockSize/3) + 'px 0 0 rgba(255,255,255,0.15), ';
                        html += 'inset -' + Math.floor(blockSize/3) + 'px 0 0 rgba(0,0,0,0.15); ';
                        html += 'image-rendering: pixelated; ';
                        html += 'image-rendering: -moz-crisp-edges; ';
                        html += 'image-rendering: crisp-edges;"></div>';
                    } else {
                        // Empty space block
                        html += '<div style="width: ' + blockSize + 'px; height: ' + blockSize + 'px; background: #000;"></div>';
                    }
                }
            }
            html += '</div>';
            
            // Legend for organelles
            html += '<div style="margin-top: 15px; font-size: 0.8em; color: #aaa; text-align: left;">';
            html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #8B4513; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Cell Membrane</div>';
            html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #3366CC; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Nucleus</div>';
            if (traits.toxicity > 0.4) {
                html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #CC3333; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Toxin Organelles</div>';
            }
            if (traits.defense > 0.4) {
                html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #CC6600; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Defense Structures</div>';
            }
            if (traits.growth > 0.5) {
                html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #33AA33; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Growth Centers</div>';
            }
            if (traits.size > 0.5) {
                html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #CCAA33; margin-right: 5px; vertical-align: middle; image-rendering: pixelated;"></span>Energy Organelles</div>';
            }
            html += '</div>';
            
            // Cell health/energy indicators with Minecraft-style blocks
            html += '<div style="display: flex; gap: 20px; margin-top: 15px;">';
            
            // Health indicator
            const health = (traits.defense + traits.hardiness) / 2;
            html += '<div style="flex: 1;">';
            html += '<div style="margin-bottom: 5px; font-weight: bold; color: #fff; text-shadow: 1px 1px 0 #000;">CELL HEALTH:</div>';
            html += generateMinecraftIndicator(health, '#33AA33', '#AA3333');
            html += '</div>';
            
            // Energy indicator  
            const energy = (traits.size + traits.growth) / 2;
            html += '<div style="flex: 1;">';
            html += '<div style="margin-bottom: 5px; font-weight: bold; color: #fff; text-shadow: 1px 1px 0 #000;">ENERGY LEVEL:</div>';
            html += generateMinecraftIndicator(energy, '#CCAA33', '#664400');
            html += '</div>';
            
            html += '</div>';
            html += '</div>';
            
            return html;
        }
        
        // Helper function to generate Minecraft-style indicators
        function generateMinecraftIndicator(value, fullColor, emptyColor) {
            let html = '<div style="display: flex; gap: 1px;">';
            const blocks = Math.floor(value * 8);
            
            for (let i = 0; i < 8; i++) {
                const isActive = i < blocks;
                const color = isActive ? fullColor : emptyColor;
                html += '<div style="width: 16px; height: 16px; background: ' + color + '; ';
                if (isActive) {
                    html += 'box-shadow: ';
                    html += 'inset 0 4px 0 rgba(255,255,255,0.3), ';
                    html += 'inset 0 -4px 0 rgba(0,0,0,0.3), ';
                    html += 'inset 4px 0 0 rgba(255,255,255,0.15), ';
                    html += 'inset -4px 0 0 rgba(0,0,0,0.15); ';
                } else {
                    html += 'box-shadow: ';
                    html += 'inset 0 2px 0 rgba(255,255,255,0.1), ';
                    html += 'inset 0 -2px 0 rgba(0,0,0,0.4); ';
                }
                html += 'border: none; ';
                html += 'image-rendering: pixelated; ';
                html += 'image-rendering: -moz-crisp-edges; ';
                html += 'image-rendering: crisp-edges;"></div>';
            }
            
            html += '</div>';
            return html;
        }
        

        
        // Generate habitat view
        function generateHabitatView(name, traits) {
            let html = '';
            
            html += '<div><strong>Preferred Habitat:</strong></div>';
            
            // Determine habitat based on traits
            let habitat = 'Plains';
            let description = 'Adaptable to various conditions';
            
            if (traits.defense > 0.7) {
                habitat = 'Mountain/Desert';
                description = 'Hardy species adapted to harsh conditions';
            } else if (traits.growth > 0.7) {
                habitat = 'Forest/Rainforest';
                description = 'Fast-growing species thriving in rich environments';
            } else if (traits.toxicity > 0.6) {
                habitat = 'Radiation/Wasteland';
                description = 'Toxic species adapted to contaminated areas';
            } else if (traits.size > 0.7) {
                habitat = 'Forest';
                description = 'Large species requiring space and resources';
            }
            
            html += '<div style="margin: 10px 0;">';
            html += '<strong>Primary Habitat:</strong> ' + habitat + '<br>';
            html += '<strong>Adaptation:</strong> ' + description + '<br>';
            html += '</div>';
            
            // Environmental tolerance
            html += '<div><strong>Environmental Tolerance:</strong></div>';
            html += '<div style="margin-left: 10px; font-size: 0.9em;">';
            html += 'Temperature: ' + (traits.hardiness > 0.5 ? 'Wide range' : 'Moderate range') + '<br>';
            html += 'Moisture: ' + (traits.growth > 0.5 ? 'High requirements' : 'Low requirements') + '<br>';
            html += 'Soil Quality: ' + (traits.size > 0.5 ? 'Rich soil preferred' : 'Any soil type') + '<br>';
            html += '</div>';
            
            return html;
        }
        
        // Get base symbol for species type
        function getBaseSymbol(name) {
            if (name.toLowerCase().includes('grass')) return '🌱';
            if (name.toLowerCase().includes('tree')) return '🌳';
            if (name.toLowerCase().includes('bush')) return '🌿';
            if (name.toLowerCase().includes('algae')) return '🌊';
            if (name.toLowerCase().includes('cactus')) return '🌵';
            if (name.toLowerCase().includes('mushroom')) return '🍄';
            return '🌱'; // Default
        }
        
        // Simple string hash function
        function stringHash(str) {
            let hash = 0;
            for (let i = 0; i < str.length; i++) {
                const char = str.charCodeAt(i);
                hash = ((hash << 5) - hash) + char;
                hash = hash & hash; // Convert to 32bit integer
            }
            return Math.abs(hash);
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
            let html = '<h3>🔬 Cellular System & Individual Organism View</h3>';
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
                
                // Add individual organism visualization
                html += '<br><h4>🦠 Individual Organism Visualization:</h4>';
                html += '<div style="background-color: #2a2a2a; padding: 15px; border-radius: 10px; margin: 10px 0;">';
                html += renderSimulatedOrganism(cellular);
                html += '</div>';
                
                // Complexity levels breakdown
                html += '<br><h4>📊 Complexity Distribution:</h4>';
                html += '<div style="margin-left: 15px;">';
                html += '<div>Level 1 (Single-cell): ' + Math.round(cellular.total_cells * 0.4) + ' organisms</div>';
                html += '<div>Level 2 (Simple multi-cell): ' + Math.round(cellular.total_cells * 0.3) + ' organisms</div>';
                html += '<div>Level 3 (Moderate complexity): ' + Math.round(cellular.total_cells * 0.2) + ' organisms</div>';
                html += '<div>Level 4 (Complex): ' + Math.round(cellular.total_cells * 0.08) + ' organisms</div>';
                html += '<div>Level 5 (Highly complex): ' + Math.round(cellular.total_cells * 0.02) + ' organisms</div>';
                html += '</div>';
            }
            
            return html;
        }
        
        // Render a simulated organism based on cellular data
        function renderSimulatedOrganism(cellular) {
            let html = '';
            
            // Create a visual representation of a typical organism
            html += '<div style="text-align: center;">';
            html += '<h5>🦠 Sample Multi-cellular Organism</h5>';
            html += '<div style="font-family: monospace; font-size: 1.2em; line-height: 1.2; margin: 15px 0;">';
            
            // Generate organism structure based on complexity
            const complexity = Math.min(5, Math.max(1, cellular.average_complexity));
            
            if (complexity >= 1) {
                html += '<div style="color: #44ff44;">🌟 Nucleus (Core)</div>';
            }
            if (complexity >= 2) {
                html += '<div style="color: #ffaa44;">⚡⚡ Mitochondria (Energy)</div>';
            }
            if (complexity >= 3) {
                html += '<div style="color: #44aaff;">🔵🔵🔵 Specialized Cells</div>';
            }
            if (complexity >= 4) {
                html += '<div style="color: #ff44aa;">📦📦 Organ Systems</div>';
            }
            if (complexity >= 5) {
                html += '<div style="color: #aa44ff;">🧠 Neural Network</div>';
            }
            
            html += '</div>';
            
            // Cell type breakdown
            html += '<div style="text-align: left; margin-top: 15px;">';
            html += '<strong>Cell Type Distribution:</strong><br>';
            html += '<div style="margin-left: 10px; font-size: 0.9em;">';
            html += '<span style="color: #44ff44;">S</span> Stem cells: ' + Math.round(cellular.total_cells * 0.15) + '<br>';
            html += '<span style="color: #ffaa44;">N</span> Nerve cells: ' + Math.round(cellular.total_cells * 0.10) + '<br>';
            html += '<span style="color: #ff4444;">M</span> Muscle cells: ' + Math.round(cellular.total_cells * 0.20) + '<br>';
            html += '<span style="color: #4444ff;">D</span> Digestive cells: ' + Math.round(cellular.total_cells * 0.25) + '<br>';
            html += '<span style="color: #ff44ff;">R</span> Reproductive cells: ' + Math.round(cellular.total_cells * 0.05) + '<br>';
            html += '<span style="color: #44ffff;">F</span> Defensive cells: ' + Math.round(cellular.total_cells * 0.15) + '<br>';
            html += '<span style="color: #aaff44;">P</span> Photosynthetic cells: ' + Math.round(cellular.total_cells * 0.08) + '<br>';
            html += '<span style="color: #ffaa44;">T</span> Storage cells: ' + Math.round(cellular.total_cells * 0.02) + '<br>';
            html += '</div>';
            html += '</div>';
            
            // Organelle visualization
            html += '<div style="margin-top: 15px;">';
            html += '<strong>Organelle Profile:</strong><br>';
            html += '<div style="display: flex; flex-wrap: wrap; gap: 5px; margin-top: 5px;">';
            html += '<span title="Nucleus">⬣</span>';
            html += '<span title="Mitochondria">⚡</span>';
            html += '<span title="Chloroplast">🌱</span>';
            html += '<span title="Ribosome">⬢</span>';
            html += '<span title="Vacuole">💧</span>';
            html += '<span title="Golgi">📦</span>';
            html += '<span title="ER">🕸</span>';
            html += '<span title="Lysosome">🗑</span>';
            html += '</div>';
            html += '</div>';
            
            html += '</div>';
            
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
            
            // Add seed dispersal information
            html += '<h4>🌱 Seed Dispersal System</h4>';
            html += '<div>Active Seeds: ' + (wind.seed_count || 0) + '</div>';
            html += '<div>Seed Banks: ' + (wind.seed_banks || 0) + '</div>';
            html += '<div>Germination Events: ' + (wind.germination_events || 0) + '</div>';
            html += '<div>Dormancy Activations: ' + (wind.dormancy_activations || 0) + '</div>';
            
            // Display dispersal statistics
            if (wind.dispersal_stats) {
                html += '<h4>🎯 Dispersal Methods</h4>';
                if (wind.dispersal_stats.dispersal_wind) html += '<div>Wind Dispersal: ' + wind.dispersal_stats.dispersal_wind + '</div>';
                if (wind.dispersal_stats.dispersal_animal) html += '<div>Animal Dispersal: ' + wind.dispersal_stats.dispersal_animal + '</div>';
                if (wind.dispersal_stats.dispersal_explosive) html += '<div>Explosive Dispersal: ' + wind.dispersal_stats.dispersal_explosive + '</div>';
                if (wind.dispersal_stats.dispersal_gravity) html += '<div>Gravity Dispersal: ' + wind.dispersal_stats.dispersal_gravity + '</div>';
                if (wind.dispersal_stats.dispersal_water) html += '<div>Water Dispersal: ' + wind.dispersal_stats.dispersal_water + '</div>';
                
                // Display seed type statistics
                if (wind.dispersal_stats.active_seeds_by_type) {
                    html += '<h4>🎯 Active Seed Types</h4>';
                    const seedTypes = wind.dispersal_stats.active_seeds_by_type;
                    for (const [type, count] of Object.entries(seedTypes)) {
                        html += '<div>' + type.replace(/_/g, ' ').toUpperCase() + ': ' + count + '</div>';
                    }
                }
            }
            
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
            
            // Initialize species modal functionality
            setupSpeciesModalEvents();
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
                    
                case 'species_extinct':
                    // Remove from player species list
                    const extinctIndex = playerSpecies.indexOf(data.species_name);
                    if (extinctIndex > -1) {
                        playerSpecies.splice(extinctIndex, 1);
                        updatePlayerSpeciesCount();
                    }
                    
                    // Show extinction notification
                    alert('⚰️ Species Extinction\n\n' + data.message);
                    console.log('Species extinct:', data.message);
                    break;
                    
                case 'subspecies_formed':
                    // Add new subspecies to player species list
                    playerSpecies.push(data.species_name);
                    updatePlayerSpeciesCount();
                    
                    // Show subspecies notification
                    alert('🧬 Species Split!\n\n' + data.message);
                    console.log('Subspecies formed:', data.message);
                    break;
                    
                case 'new_species_detected':
                    // Just log this for now - informational only
                    console.log('New species detected:', data.message);
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
                    html += '<div>' + behavior_name.replace(/_/g, ' ') + ': ' + count + ' entities</div>';
                });
            }
            
            if (behavior.avg_proficiency && Object.keys(behavior.avg_proficiency).length > 0) {
                html += '<h4>Average Proficiency:</h4>';
                Object.entries(behavior.avg_proficiency).forEach(([behavior_name, proficiency]) => {
                    html += '<div>' + behavior_name.replace(/_/g, ' ') + ': ' + proficiency.toFixed(2) + '</div>';
                });
            }
            
            if (behavior.discovered_behaviors === 0) {
                html += '<br><div>No emergent behaviors discovered yet</div>';
                html += '<div style="color: #888; font-style: italic;">Behaviors emerge naturally as entities explore and learn from their environment and each other.</div>';
            } else {
                html += '<br><h4>Behavior Development:</h4>';
                if (behavior.discovered_behaviors < 3) {
                    html += '<div>Development Level: Early behavior emergence</div>';
                } else if (behavior.discovered_behaviors < 6) {
                    html += '<div>Development Level: Moderate behavior complexity</div>';
                } else {
                    html += '<div>Development Level: Advanced behavioral evolution</div>';
                }
                
                // Show list of behaviors that have been discovered by entities
                if (behavior.behavior_spread && Object.keys(behavior.behavior_spread).length > 0) {
                    html += '<h4>Active Behaviors:</h4>';
                    const sortedBehaviors = Object.entries(behavior.behavior_spread)
                        .sort((a, b) => b[1] - a[1]); // Sort by count descending
                    sortedBehaviors.forEach(([behavior_name, count]) => {
                        if (count > 0) {
                            const proficiency = behavior.avg_proficiency[behavior_name] || 0;
                            html += '<div>• <strong>' + behavior_name.replace(/_/g, ' ') + '</strong>: ' + count + ' entities (avg proficiency: ' + proficiency.toFixed(2) + ')</div>';
                        }
                    });
                }
            }
            
            return html;
        }
        
        // Render reproduction view
        function renderReproduction(reproduction) {
            let html = '<h3>🥚 Reproduction & Life Cycle</h3>';
            html += '<div>Active Eggs: ' + reproduction.active_eggs + '</div>';
            html += '<div>Decaying Items: ' + reproduction.decaying_items + '</div>';
            html += '<div>Pregnant Entities: ' + reproduction.pregnant_entities + '</div>';
            html += '<div>Ready to Mate: ' + reproduction.ready_to_mate + '</div>';
            html += '<div>Mating Season Entities: ' + reproduction.mating_season_entities + '</div>';
            html += '<div>Migrating Entities: ' + reproduction.migrating_entities + '</div>';
            html += '<div>Cross-Species Mating: ' + reproduction.cross_species_mating + '</div>';
            html += '<div>Territories with Mating: ' + reproduction.territories_with_mating + '</div>';
            html += '<div>Seasonal Mating Rate: ' + reproduction.seasonal_mating_rate.toFixed(2) + 'x</div>';
            
            if (reproduction.reproduction_modes && Object.keys(reproduction.reproduction_modes).length > 0) {
                html += '<h4>Reproduction Modes:</h4>';
                Object.entries(reproduction.reproduction_modes).forEach(([mode, count]) => {
                    html += '<div>' + mode + ': ' + count + ' entities</div>';
                });
            }
            
            if (reproduction.mating_strategies && Object.keys(reproduction.mating_strategies).length > 0) {
                html += '<h4>Mating Strategies:</h4>';
                Object.entries(reproduction.mating_strategies).forEach(([strategy, count]) => {
                    html += '<div>' + strategy + ': ' + count + ' entities</div>';
                });
            }
            
            html += '<br><h4>Reproduction Activity:</h4>';
            if (reproduction.ready_to_mate === 0) {
                html += '<div>Activity Level: No active mating</div>';
            } else if (reproduction.ready_to_mate < 5) {
                html += '<div>Activity Level: Low reproductive activity</div>';
            } else {
                html += '<div>Activity Level: High reproductive activity</div>';
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
            
            let html = '<h3>⚠️ Anomaly Detection & Historical Analysis</h3>';
            
            if (anomalies.total_anomalies === 0) {
                html += '<div style="color: #4CAF50;">✅ No anomalies detected!</div>';
                html += '<div>The simulation appears to be running within expected parameters.</div>';
                html += '<div style="margin-top: 10px; color: #888; font-style: italic;">Anomaly detection monitors for unusual patterns in energy conservation, population dynamics, trait distributions, and system behaviors.</div>';
            } else {
                html += '<div>Found ' + anomalies.total_anomalies + ' anomalies in simulation history:</div><br>';
                
                // Anomaly types summary with enhanced information
                if (anomalies.anomaly_types && Object.keys(anomalies.anomaly_types).length > 0) {
                    html += '<h4>📊 Anomaly Categories:</h4>';
                    Object.entries(anomalies.anomaly_types).forEach(([type, count]) => {
                        let severity = '';
                        let color = '#ffcc00';
                        if (count > 10) {
                            severity = ' - High frequency';
                            color = '#ff6666';
                        } else if (count > 5) {
                            severity = ' - Moderate frequency';
                            color = '#ffaa00';
                        } else {
                            severity = ' - Low frequency';
                            color = '#88cc88';
                        }
                        html += '<div style="color: ' + color + ';">• ' + type.replace(/_/g, ' ').toUpperCase() + ': ' + count + ' occurrences' + severity + '</div>';
                    });
                    html += '<br>';
                }
                
                // Recent anomalies with enhanced display
                if (anomalies.recent_anomalies && anomalies.recent_anomalies.length > 0) {
                    html += '<h4>📋 Recent Anomaly History (Last ' + anomalies.recent_anomalies.length + ' events):</h4>';
                    html += '<div style="max-height: 300px; overflow-y: auto; border: 1px solid #444; padding: 10px; border-radius: 5px;">';
                    
                    // Sort anomalies by tick (newest first)
                    const sortedAnomalies = [...anomalies.recent_anomalies].sort((a, b) => b.tick - a.tick);
                    
                    sortedAnomalies.forEach(anomaly => {
                        let severityColor = '#ffcc00';
                        let severityIcon = '⚠️';
                        let confidenceIcon = '🔍';
                        
                        if (anomaly.severity >= 0.8) {
                            severityColor = '#ff4444';
                            severityIcon = '🚨';
                        } else if (anomaly.severity >= 0.6) {
                            severityColor = '#ff8800';
                            severityIcon = '⚠️';
                        } else if (anomaly.severity >= 0.4) {
                            severityColor = '#ffcc00';
                            severityIcon = '⚡';
                        } else {
                            severityColor = '#88cc88';
                            severityIcon = 'ℹ️';
                        }
                        
                        if (anomaly.confidence >= 0.8) {
                            confidenceIcon = '🎯';
                        } else if (anomaly.confidence >= 0.6) {
                            confidenceIcon = '🔍';
                        } else {
                            confidenceIcon = '❓';
                        }
                        
                        html += '<div style="margin: 8px 0; padding: 8px; background-color: rgba(68, 68, 68, 0.3); border-radius: 3px; border-left: 3px solid ' + severityColor + ';">';
                        html += '<div style="display: flex; justify-content: space-between; align-items: center;">';
                        html += '<strong>' + severityIcon + ' ' + anomaly.type.replace(/_/g, ' ').toUpperCase() + '</strong>';
                        html += '<span style="color: #aaa; font-size: 0.9em;">Tick ' + anomaly.tick + '</span>';
                        html += '</div>';
                        html += '<div style="margin: 5px 0;">' + anomaly.description + '</div>';
                        html += '<div style="display: flex; gap: 15px; font-size: 0.9em; color: #ccc;">';
                        html += '<span>Severity: ' + severityIcon + ' ' + (anomaly.severity * 100).toFixed(0) + '%</span>';
                        html += '<span>Confidence: ' + confidenceIcon + ' ' + (anomaly.confidence * 100).toFixed(0) + '%</span>';
                        html += '</div>';
                        html += '</div>';
                    });
                    html += '</div>';
                }
                
                // Recommendations with enhanced formatting
                if (anomalies.recommendations && anomalies.recommendations.length > 0) {
                    html += '<h4>💡 Diagnostic Recommendations:</h4>';
                    html += '<div style="background-color: rgba(76, 175, 80, 0.1); border-left: 3px solid #4CAF50; padding: 10px; border-radius: 3px;">';
                    anomalies.recommendations.forEach(rec => {
                        html += '<div style="margin: 5px 0;">💡 ' + rec + '</div>';
                    });
                    html += '</div>';
                } else if (anomalies.total_anomalies > 0) {
                    html += '<h4>💡 General Recommendations:</h4>';
                    html += '<div style="background-color: rgba(255, 193, 7, 0.1); border-left: 3px solid #ffc107; padding: 10px; border-radius: 3px;">';
                    html += '<div>• Monitor system parameters for patterns</div>';
                    html += '<div>• Check if anomalies correlate with specific events</div>';
                    html += '<div>• Consider adjusting simulation parameters if anomalies persist</div>';
                    html += '</div>';
                }
            }
            
            return html;
        }
        
        // Render warfare view
        function renderWarfare(warfare) {
            if (!warfare) {
                return '<h3>⚔️ Warfare & Diplomacy</h3><div>Warfare data not available</div>';
            }
            
            let html = '<h3>⚔️ Multi-Colony Warfare & Diplomacy</h3>';
            
            // Colony overview
            if (warfare.colonies && warfare.colonies.length > 0) {
                html += '<h4>🏰 Active Colonies (' + warfare.colonies.length + '):</h4>';
                html += '<div class="colony-list">';
                warfare.colonies.forEach(colony => {
                    const relationColor = getRelationColor(colony.dominant_relation);
                    html += '<div class="colony-item">';
                    html += '<strong>' + colony.name + '</strong> ';
                    html += '<span style="color: ' + relationColor + '">(' + colony.dominant_relation + ')</span><br>';
                    html += 'Size: ' + colony.size + ' | Territory: ' + colony.territory_size + ' cells<br>';
                    html += 'Military: ' + colony.military_strength.toFixed(1) + ' | Resources: ' + colony.total_resources + '<br>';
                    if (colony.recent_activity) {
                        html += '<small style="color: #888;">Recent: ' + colony.recent_activity + '</small>';
                    }
                    html += '</div>';
                });
                html += '</div>';
            }
            
            // Active conflicts
            if (warfare.active_conflicts && warfare.active_conflicts.length > 0) {
                html += '<h4>⚔️ Active Conflicts (' + warfare.active_conflicts.length + '):</h4>';
                warfare.active_conflicts.forEach(conflict => {
                    const intensityColor = getConflictIntensityColor(conflict.intensity);
                    html += '<div class="conflict-item">';
                    html += '<strong style="color: ' + intensityColor + '">' + conflict.type + '</strong><br>';
                    html += conflict.attacker + ' vs ' + conflict.defender + '<br>';
                    html += 'Duration: ' + conflict.duration + ' ticks | Intensity: ' + conflict.intensity.toFixed(2) + '<br>';
                    html += 'Casualties: ' + conflict.casualties + ' | Status: ' + conflict.status + '<br>';
                    if (conflict.cause) {
                        html += '<small>Cause: ' + conflict.cause + '</small>';
                    }
                    html += '</div>';
                });
            } else {
                html += '<h4>⚔️ Active Conflicts: None</h4>';
                html += '<div style="color: #4CAF50;">🕊️ All colonies are currently at peace</div>';
            }
            
            // Diplomatic relations summary
            if (warfare.diplomatic_summary) {
                html += '<h4>🤝 Diplomatic Relations Summary:</h4>';
                Object.entries(warfare.diplomatic_summary).forEach(([relation, count]) => {
                    const relationColor = getRelationColor(relation);
                    html += '<div style="color: ' + relationColor + ';">' + relation + ': ' + count + '</div>';
                });
            }
            
            // Trade activity
            if (warfare.trade_activity) {
                html += '<h4>💰 Trade Activity:</h4>';
                html += '<div>Active Agreements: ' + warfare.trade_activity.active_agreements + '</div>';
                html += '<div>Trade Volume: ' + warfare.trade_activity.total_volume + '</div>';
                html += '<div>Trade Efficiency: ' + (warfare.trade_activity.efficiency * 100).toFixed(1) + '%</div>';
            }
            
            // Recent warfare events
            if (warfare.recent_events && warfare.recent_events.length > 0) {
                html += '<h4>📰 Recent Warfare Events:</h4>';
                html += '<div class="events-list">';
                warfare.recent_events.slice(0, 10).forEach(event => {
                    html += '<div class="event-item">';
                    html += '<small>[Tick ' + event.tick + ']</small> ';
                    html += event.description;
                    html += '</div>';
                });
                html += '</div>';
            }
            
            // Statistics
            if (warfare.statistics) {
                html += '<h4>📊 Warfare Statistics:</h4>';
                html += '<div>Total Conflicts: ' + warfare.statistics.total_conflicts + '</div>';
                html += '<div>Total Casualties: ' + warfare.statistics.total_casualties + '</div>';
                html += '<div>Peace Treaties: ' + warfare.statistics.peace_treaties + '</div>';
                html += '<div>Alliance Formations: ' + warfare.statistics.alliance_formations + '</div>';
            }
            
            return html;
        }
        
        function getRelationColor(relation) {
            const colors = {
                'Allied': '#4CAF50',
                'Enemy': '#F44336', 
                'Neutral': '#FFC107',
                'Trading': '#2196F3',
                'Truce': '#FF9800',
                'Vassal': '#9C27B0'
            };
            return colors[relation] || '#888';
        }
        
        function getConflictIntensityColor(intensity) {
            if (intensity < 0.3) return '#FFC107'; // Low intensity - yellow
            if (intensity < 0.7) return '#FF9800'; // Medium intensity - orange  
            return '#F44336'; // High intensity - red
        }
        
        function renderFungal(fungal) {
            if (!fungal) {
                return '<h3>🍄 Fungal Networks</h3><div>Fungal system data not available</div>';
            }
            
            let html = '<h3>🍄 Fungal Networks & Decomposer System</h3>';
            
            // Decomposer overview
            if (fungal.total_decomposers !== undefined) {
                html += '<h4>🧪 Decomposer Status:</h4>';
                html += '<div class="stats-row">';
                html += '<div class="stat-item">Total Decomposers: <strong>' + (fungal.total_decomposers || 0) + '</strong></div>';
                html += '<div class="stat-item">Active Decomposers: <strong>' + (fungal.active_decomposers || 0) + '</strong></div>';
                html += '</div>';
            }
            
            // Nutrient cycling statistics
            if (fungal.nutrient_cycling) {
                html += '<h4>♻️ Nutrient Cycling:</h4>';
                html += '<div class="stats-row">';
                html += '<div class="stat-item">Decomposition Rate: <strong>' + (fungal.nutrient_cycling.decomposition_rate || 0).toFixed(2) + '/tick</strong></div>';
                html += '<div class="stat-item">Nutrients Released: <strong>' + (fungal.nutrient_cycling.nutrients_released || 0).toFixed(1) + '</strong></div>';
                html += '</div>';
                
                if (fungal.nutrient_cycling.nutrient_types) {
                    html += '<div class="nutrient-breakdown">';
                    Object.entries(fungal.nutrient_cycling.nutrient_types).forEach(([nutrient, amount]) => {
                        html += '<div class="nutrient-item">' + nutrient + ': ' + amount.toFixed(1) + '</div>';
                    });
                    html += '</div>';
                }
            }
            
            // Spore networks
            if (fungal.spore_networks) {
                html += '<h4>🌐 Spore Networks:</h4>';
                html += '<div class="stats-row">';
                html += '<div class="stat-item">Network Connections: <strong>' + (fungal.spore_networks.connections || 0) + '</strong></div>';
                html += '<div class="stat-item">Network Efficiency: <strong>' + ((fungal.spore_networks.efficiency || 0) * 100).toFixed(1) + '%</strong></div>';
                html += '</div>';
            }
            
            // Fungal reproduction
            if (fungal.reproduction) {
                html += '<h4>🌱 Fungal Reproduction:</h4>';
                html += '<div class="stats-row">';
                html += '<div class="stat-item">Spores Released: <strong>' + (fungal.reproduction.spores_released || 0) + '</strong></div>';
                html += '<div class="stat-item">Successful Germinations: <strong>' + (fungal.reproduction.germinations || 0) + '</strong></div>';
                html += '</div>';
            }
            
            // Recent events
            if (fungal.recent_events && fungal.recent_events.length > 0) {
                html += '<h4>📋 Recent Fungal Activity:</h4>';
                html += '<div class="event-list">';
                fungal.recent_events.slice(0, 10).forEach(event => {
                    html += '<div class="event-item">• ' + event + '</div>';
                });
                html += '</div>';
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

// handleExportEvents exports all events from the central event bus
func (wi *WebInterface) handleExportEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for filtering
	eventType := r.URL.Query().Get("type")
	category := r.URL.Query().Get("category")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	var events []CentralEvent
	
	if wi.world.CentralEventBus != nil {
		if eventType != "" {
			events = wi.world.CentralEventBus.GetEventsByType(eventType)
		} else if category != "" {
			events = wi.world.CentralEventBus.GetEventsByCategory(category)
		} else {
			events = wi.world.CentralEventBus.GetAllEvents()
		}
	}

	exportData := map[string]interface{}{
		"events":      events,
		"total_count": len(events),
		"export_time": time.Now(),
		"filters": map[string]string{
			"type":     eventType,
			"category": category,
			"format":   format,
		},
	}

	if format == "csv" {
		wi.exportEventsAsCSV(w, events)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=events_export.json")
		json.NewEncoder(w).Encode(exportData)
	}
}

// handleExportAnalysis exports statistical analysis data
func (wi *WebInterface) handleExportAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	var analysisData map[string]interface{}
	
	if wi.world.StatisticalReporter != nil {
		analysisData = map[string]interface{}{
			"summary_statistics": wi.world.StatisticalReporter.GetSummaryStatistics(),
			"recent_events":      wi.world.StatisticalReporter.Events,
			"snapshots":          wi.world.StatisticalReporter.Snapshots,
			"export_time":        time.Now(),
		}
	} else {
		analysisData = map[string]interface{}{
			"error":       "Statistical reporter not available",
			"export_time": time.Now(),
		}
	}

	if format == "csv" {
		wi.exportAnalysisAsCSV(w, analysisData)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=analysis_export.json")
		json.NewEncoder(w).Encode(analysisData)
	}
}

// handleExportAnomalies exports anomaly detection data
func (wi *WebInterface) handleExportAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	var anomaliesData map[string]interface{}
	
	if wi.world.StatisticalReporter != nil {
		anomaliesData = map[string]interface{}{
			"anomalies":      wi.world.StatisticalReporter.Anomalies,
			"total_count":    len(wi.world.StatisticalReporter.Anomalies),
			"anomaly_types":  wi.world.StatisticalReporter.detectedAnomalies,
			"export_time":    time.Now(),
		}
	} else {
		anomaliesData = map[string]interface{}{
			"error":       "Statistical reporter not available",
			"export_time": time.Now(),
		}
	}

	if format == "csv" {
		wi.exportAnomaliesAsCSV(w, anomaliesData)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=anomalies_export.json")
		json.NewEncoder(w).Encode(anomaliesData)
	}
}

// exportEventsAsCSV exports events in CSV format
func (wi *WebInterface) exportEventsAsCSV(w http.ResponseWriter, events []CentralEvent) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=events_export.csv")

	w.Write([]byte("ID,Timestamp,Tick,Type,Category,SubCategory,Source,Description,EntityID,PlantID,Position,Severity\n"))
	
	for _, event := range events {
		position := ""
		if event.Position != nil {
			position = fmt.Sprintf("%.2f;%.2f", event.Position.X, event.Position.Y)
		}
		
		line := fmt.Sprintf("%d,%s,%d,%s,%s,%s,%s,\"%s\",%d,%d,%s,%s\n",
			event.ID,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Tick,
			event.Type,
			event.Category,
			event.SubCategory,
			event.Source,
			event.Description,
			event.EntityID,
			event.PlantID,
			position,
			event.Severity,
		)
		w.Write([]byte(line))
	}
}

// exportAnalysisAsCSV exports analysis data in CSV format
func (wi *WebInterface) exportAnalysisAsCSV(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=analysis_export.csv")

	// Simple CSV with key-value pairs for analysis data
	w.Write([]byte("Key,Value\n"))
	
	if stats, ok := data["summary_statistics"].(map[string]interface{}); ok {
		for key, value := range stats {
			w.Write([]byte(fmt.Sprintf("%s,%v\n", key, value)))
		}
	}
}

// exportAnomaliesAsCSV exports anomalies in CSV format
func (wi *WebInterface) exportAnomaliesAsCSV(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=anomalies_export.csv")

	w.Write([]byte("Type,Severity,Tick,Description,Confidence\n"))
	
	if anomalies, ok := data["anomalies"].([]Anomaly); ok {
		for _, anomaly := range anomalies {
			line := fmt.Sprintf("%s,%.3f,%d,\"%s\",%.3f\n",
				anomaly.Type,
				anomaly.Severity,
				anomaly.Tick,
				anomaly.Description,
				anomaly.Confidence,
			)
			w.Write([]byte(line))
		}
	}
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

// handlePlayerEvent handles world events related to players (extinctions, species splits)
func (wi *WebInterface) handlePlayerEvent(eventType string, data map[string]interface{}) {
	speciesName, ok := data["species_name"].(string)
	if !ok {
		return
	}
	
	// Find the player who owns this species
	playerID, exists := wi.playerManager.GetSpeciesOwner(speciesName)
	if !exists {
		return // Species not owned by a player
	}
	
	// Find the WebSocket connection for this player
	var playerWS *websocket.Conn
	wi.clientsMutex.RLock()
	for ws, pID := range wi.clientPlayers {
		if pID == playerID {
			playerWS = ws
			break
		}
	}
	wi.clientsMutex.RUnlock()
	
	if playerWS == nil {
		return // Player not currently connected
	}
	
	// Handle different event types
	switch eventType {
	case "species_extinct":
		wi.playerManager.MarkSpeciesExtinct(speciesName)
		
		notification := map[string]interface{}{
			"type":         "species_extinct",
			"species_name": speciesName,
			"message":      fmt.Sprintf("Your species '%s' has gone extinct! You can create a new species.", speciesName),
			"last_count":   data["last_count"],
			"tick":         data["tick"],
		}
		wi.sendJSONToClient(playerWS, notification)
		
		log.Printf("Player %s notified of species extinction: %s", playerID, speciesName)
		
	case "subspecies_formed":
		parentSpecies := data["parent_species"].(string)
		
		// Check if the player owns the parent species
		if wi.playerManager.CanPlayerControlSpecies(playerID, parentSpecies) {
			// Add the subspecies to the player
			err := wi.playerManager.AddSubSpecies(parentSpecies, speciesName)
			if err != nil {
				log.Printf("Error adding subspecies %s to player %s: %v", speciesName, playerID, err)
				return
			}
			
			notification := map[string]interface{}{
				"type":           "subspecies_formed",
				"species_name":   speciesName,
				"parent_species": parentSpecies,
				"message":        fmt.Sprintf("Your species '%s' has split! You can now control the new subspecies '%s'.", parentSpecies, speciesName),
				"entity_count":   data["entity_count"],
				"tick":           data["tick"],
			}
			wi.sendJSONToClient(playerWS, notification)
			
			log.Printf("Player %s notified of subspecies formation: %s from %s", playerID, speciesName, parentSpecies)
		}
		
	case "new_species_detected":
		// This is a new species that appeared but doesn't seem related to player species
		// We can notify for informational purposes but don't give control
		notification := map[string]interface{}{
			"type":         "new_species_detected",
			"species_name": speciesName,
			"message":      fmt.Sprintf("A new species '%s' has appeared in the simulation.", speciesName),
			"entity_count": data["entity_count"],
			"tick":         data["tick"],
		}
		wi.sendJSONToClient(playerWS, notification)
	}
}