package main

import (
	"encoding/json"
	"fmt"
	"log"
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
            <div class="controls">
                <button id="pause-btn" onclick="togglePause()">⏸ Pause</button>
                <button onclick="resetSimulation()">🔄 Reset</button>
                <button onclick="saveState()">💾 Save</button>
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
                    . Plains | ♠ Forest | ~ Desert<br>
                    ^ Mountain | ≈ Water | ☢ Radiation<br><br>
                    
                    <strong>Entities:</strong><br>
                    H Herbivore | P Predator | O Omnivore<br>
                    Numbers = Multiple entities<br><br>
                    
                    <strong>Plants:</strong><br>
                    . Grass | ♦ Bush | ♠ Tree<br>
                    ♪ Mushroom | ≈ Algae | † Cactus
                </div>
            </div>
        </div>
    </div>
    
    <script>
        let ws = null;
        let isPaused = false;
        let currentView = 'GRID';
        
        const viewModes = [
            'GRID', 'STATS', 'EVENTS', 'POPULATIONS', 'COMMUNICATION',
            'CIVILIZATION', 'PHYSICS', 'WIND', 'SPECIES', 'NETWORK',
            'DNA', 'CELLULAR', 'EVOLUTION', 'TOPOLOGY'
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
                    viewContent.innerHTML = '<div class="grid-container">' + renderGrid(data.grid) + '</div>';
                    break;
                    
                case 'STATS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderStats(data.stats) + '</div>';
                    break;
                    
                case 'POPULATIONS':
                    viewContent.innerHTML = '<div class="stats-section">' + renderPopulations(data.populations) + '</div>';
                    break;
                    
                case 'COMMUNICATION':
                    viewContent.innerHTML = '<div class="stats-section">' + renderCommunication(data.communication) + '</div>';
                    break;
                    
                case 'WIND':
                    viewContent.innerHTML = '<div class="stats-section">' + renderWind(data.wind) + '</div>';
                    break;
                    
                default:
                    viewContent.innerHTML = '<div class="stats-section"><h3>' + currentView + '</h3><p>View not yet implemented</p></div>';
            }
        }
        
        // Render grid view
        function renderGrid(grid) {
            let result = '';
            for (let y = 0; y < grid.length; y++) {
                for (let x = 0; x < grid[y].length; x++) {
                    const cell = grid[y][x];
                    if (cell.entity_count > 0) {
                        result += cell.entity_symbol;
                    } else if (cell.plant_count > 0) {
                        result += cell.plant_symbol;
                    } else {
                        result += cell.biome_symbol;
                    }
                }
                if (y < grid.length - 1) {
                    result += '\n';
                }
            }
            return result;
        }
        
        // Render stats view
        function renderStats(stats) {
            let html = '<h3>📊 Detailed Statistics</h3>';
            for (const [key, value] of Object.entries(stats)) {
                html += '<div>' + key.replace('_', ' ') + ': ' + (typeof value === 'number' ? value.toFixed(2) : value) + '</div>';
            }
            return html;
        }
        
        // Render populations view
        function renderPopulations(populations) {
            let html = '<h3>👥 Population Details</h3>';
            populations.forEach(pop => {
                html += '<div class="stats-section">';
                html += '<h4>' + pop.name + '</h4>';
                html += '<div>Count: ' + pop.count + '</div>';
                html += '<div>Average Fitness: ' + pop.avg_fitness.toFixed(2) + '</div>';
                html += '<div>Average Energy: ' + pop.avg_energy.toFixed(2) + '</div>';
                html += '<div>Average Age: ' + pop.avg_age.toFixed(1) + '</div>';
                html += '</div>';
            });
            return html;
        }
        
        // Render communication view
        function renderCommunication(comm) {
            let html = '<h3>📡 Communication System</h3>';
            html += '<div>Active Signals: ' + comm.active_signals + '</div>';
            
            if (Object.keys(comm.signal_types).length > 0) {
                html += '<h4>Signal Types:</h4>';
                for (const [type, count] of Object.entries(comm.signal_types)) {
                    html += '<div>' + type + ': ' + count + '</div>';
                }
            }
            
            return html;
        }
        
        // Render wind view
        function renderWind(wind) {
            let html = '<h3>🌬️ Wind System</h3>';
            html += '<div>Direction: ' + (wind.direction * 180 / Math.PI).toFixed(1) + '°</div>';
            html += '<div>Strength: ' + wind.strength.toFixed(2) + '</div>';
            html += '<div>Turbulence: ' + wind.turbulence_level.toFixed(2) + '</div>';
            html += '<div>Weather: ' + wind.weather_pattern + '</div>';
            html += '<div>Pollen Count: ' + wind.pollen_count + '</div>';
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
        
        // Initialize the interface
        window.onload = function() {
            initViewTabs();
            connect();
        };
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
			wi.handleClientAction(action)
		}
	}
	
	// Remove client from list
	wi.clientsMutex.Lock()
	delete(wi.clients, ws)
	wi.clientsMutex.Unlock()
	
	log.Printf("Client disconnected. Total clients: %d", len(wi.clients))
}

// handleClientAction processes actions from web clients
func (wi *WebInterface) handleClientAction(action string) {
	switch action {
	case "toggle_pause":
		// For now, just log the action
		// In a full implementation, this would control the simulation
		log.Printf("Client requested pause toggle")
		
	case "reset":
		log.Printf("Client requested reset")
		
	case "save_state":
		log.Printf("Client requested state save")
		// Could trigger a state save here
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

// Stop stops the web interface
func (wi *WebInterface) Stop() {
	close(wi.stopChan)
}