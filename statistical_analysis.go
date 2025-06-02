package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

// StatisticalEvent represents a detailed event for statistical analysis
type StatisticalEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	Tick         int                    `json:"tick"`
	EventType    string                 `json:"event_type"`
	Category     string                 `json:"category"`      // "entity", "plant", "environment", "system"
	EntityID     int                    `json:"entity_id,omitempty"`
	PlantID      int                    `json:"plant_id,omitempty"`
	Position     *Position              `json:"position,omitempty"`
	OldValue     interface{}            `json:"old_value,omitempty"`
	NewValue     interface{}            `json:"new_value,omitempty"`
	Change       float64                `json:"change,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	ImpactedIDs  []int                  `json:"impacted_ids,omitempty"` // IDs of other entities/plants affected
}

// StatisticalSnapshot represents a complete system state at a given tick
type StatisticalSnapshot struct {
	Tick                int                        `json:"tick"`
	Timestamp           time.Time                  `json:"timestamp"`
	TotalEntities       int                        `json:"total_entities"`
	TotalPlants         int                        `json:"total_plants"`
	TotalEnergy         float64                    `json:"total_energy"`
	SpeciesCount        int                        `json:"species_count"`
	PopulationsBySpecies map[string]int            `json:"populations_by_species"`
	TraitDistributions  map[string][]float64       `json:"trait_distributions"`
	BiomeDistributions  map[string]int             `json:"biome_distributions"`
	ResourceDistribution map[string]float64        `json:"resource_distribution"`
	PhysicsMetrics      PhysicsSnapshot            `json:"physics_metrics"`
	CommunicationMetrics CommunicationSnapshot     `json:"communication_metrics"`
}

// PhysicsSnapshot captures physics system state
type PhysicsSnapshot struct {
	TotalMomentum       float64 `json:"total_momentum"`
	TotalKineticEnergy  float64 `json:"total_kinetic_energy"`
	CollisionCount      int     `json:"collision_count"`
	AverageVelocity     float64 `json:"average_velocity"`
}

// CommunicationSnapshot captures communication system state
type CommunicationSnapshot struct {
	ActiveSignals       int                `json:"active_signals"`
	SignalsByType       map[string]int     `json:"signals_by_type"`
	SignalEfficiency    float64            `json:"signal_efficiency"`
}

// AnomalyType represents different types of statistical anomalies
type AnomalyType string

const (
	AnomalyEnergyConservation    AnomalyType = "energy_conservation"
	AnomalyUnrealisticDistribution AnomalyType = "unrealistic_distribution"
	AnomalyMathematicalInconsistency AnomalyType = "mathematical_inconsistency"
	AnomalyBiologicalImplausibility AnomalyType = "biological_implausibility"
	AnomalyPopulationAnomaly       AnomalyType = "population_anomaly"
	AnomalyPhysicsViolation        AnomalyType = "physics_violation"
)

// Anomaly represents a detected statistical anomaly
type Anomaly struct {
	Type        AnomalyType            `json:"type"`
	Severity    float64                `json:"severity"`    // 0-1 scale
	Tick        int                    `json:"tick"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Confidence  float64                `json:"confidence"`  // 0-1 scale
}

// StatisticalReporter handles comprehensive data collection and analysis
type StatisticalReporter struct {
	Events                []StatisticalEvent    `json:"events"`
	Snapshots            []StatisticalSnapshot `json:"snapshots"`
	Anomalies            []Anomaly             `json:"anomalies"`
	MaxEvents            int                   `json:"max_events"`
	MaxSnapshots         int                   `json:"max_snapshots"`
	SnapshotInterval     int                   `json:"snapshot_interval"`    // Take snapshot every N ticks
	AnalysisInterval     int                   `json:"analysis_interval"`    // Run analysis every N ticks
	lastSnapshot         *StatisticalSnapshot
	totalEnergyBaseline  float64               // Expected total energy
	detectedAnomalies    map[AnomalyType]int   // Count of each anomaly type
}

// NewStatisticalReporter creates a new statistical reporter
func NewStatisticalReporter(maxEvents, maxSnapshots, snapshotInterval, analysisInterval int) *StatisticalReporter {
	return &StatisticalReporter{
		Events:           make([]StatisticalEvent, 0),
		Snapshots:        make([]StatisticalSnapshot, 0),
		Anomalies:        make([]Anomaly, 0),
		MaxEvents:        maxEvents,
		MaxSnapshots:     maxSnapshots,
		SnapshotInterval: snapshotInterval,
		AnalysisInterval: analysisInterval,
		detectedAnomalies: make(map[AnomalyType]int),
	}
}

// LogEntityEvent logs entity-related events with full context
func (sr *StatisticalReporter) LogEntityEvent(tick int, eventType string, entity *Entity, oldValue, newValue interface{}, impactedEntities []*Entity) {
	impactedIDs := make([]int, len(impactedEntities))
	for i, e := range impactedEntities {
		impactedIDs[i] = e.ID
	}

	var change float64
	if oldVal, ok := oldValue.(float64); ok {
		if newVal, ok := newValue.(float64); ok {
			change = newVal - oldVal
		}
	}

	event := StatisticalEvent{
		Timestamp:   time.Now(),
		Tick:        tick,
		EventType:   eventType,
		Category:    "entity",
		EntityID:    entity.ID,
		Position:    &entity.Position,
		OldValue:    oldValue,
		NewValue:    newValue,
		Change:      change,
		ImpactedIDs: impactedIDs,
		Metadata: map[string]interface{}{
			"species":       entity.Species,
			"energy":        entity.Energy,
			"age":          entity.Age,
			"is_alive":     entity.IsAlive,
		},
	}
	sr.addEvent(event)
}

// LogPlantEvent logs plant-related events
func (sr *StatisticalReporter) LogPlantEvent(tick int, eventType string, plant *Plant, oldValue, newValue interface{}) {
	var change float64
	if oldVal, ok := oldValue.(float64); ok {
		if newVal, ok := newValue.(float64); ok {
			change = newVal - oldVal
		}
	}

	event := StatisticalEvent{
		Timestamp: time.Now(),
		Tick:      tick,
		EventType: eventType,
		Category:  "plant",
		PlantID:   plant.ID,
		Position:  &plant.Position,
		OldValue:  oldValue,
		NewValue:  newValue,
		Change:    change,
		Metadata: map[string]interface{}{
			"type":         plant.Type,
			"energy":       plant.Energy,
			"age":         plant.Age,
			"is_alive":    plant.IsAlive,
			"size":        plant.Size,
		},
	}
	sr.addEvent(event)
}

// LogSystemEvent logs system-wide events
func (sr *StatisticalReporter) LogSystemEvent(tick int, eventType string, description string, data map[string]interface{}) {
	event := StatisticalEvent{
		Timestamp: time.Now(),
		Tick:      tick,
		EventType: eventType,
		Category:  "system",
		Metadata:  data,
	}
	sr.addEvent(event)
}

// TakeSnapshot captures complete system state for analysis
func (sr *StatisticalReporter) TakeSnapshot(world *World) {
	snapshot := StatisticalSnapshot{
		Tick:              world.Tick,
		Timestamp:         time.Now(),
		TotalEntities:     len(world.AllEntities),
		TotalPlants:       len(world.AllPlants),
		PopulationsBySpecies: make(map[string]int),
		TraitDistributions: make(map[string][]float64),
		BiomeDistributions: make(map[string]int),
		ResourceDistribution: make(map[string]float64),
	}

	// Calculate total energy in system
	totalEnergy := 0.0
	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			totalEnergy += entity.Energy
		}
	}
	for _, plant := range world.AllPlants {
		if plant.IsAlive {
			totalEnergy += plant.Energy
		}
	}
	snapshot.TotalEnergy = totalEnergy

	// Count populations by species
	speciesSet := make(map[string]bool)
	for _, pop := range world.Populations {
		count := 0
		for _, entity := range pop.Entities {
			if entity.IsAlive {
				count++
			}
		}
		snapshot.PopulationsBySpecies[pop.Species] = count
		speciesSet[pop.Species] = true
	}
	snapshot.SpeciesCount = len(speciesSet)

	// Collect trait distributions
	traits := []string{"vision", "speed", "size", "energy_efficiency", "aggression", "cooperation", 
					  "intelligence", "curiosity", "social", "territorial", "nocturnal", "endurance",
					  "strength", "defense", "stealth"}
	
	for _, trait := range traits {
		values := make([]float64, 0)
		for _, entity := range world.AllEntities {
			if entity.IsAlive {
				if val := entity.GetTrait(trait); val != 0 || entity.Traits[trait].Name != "" {
					values = append(values, val)
				}
			}
		}
		snapshot.TraitDistributions[trait] = values
	}

	// Count biome distributions
	for y := 0; y < len(world.Grid); y++ {
		for x := 0; x < len(world.Grid[y]); x++ {
			biome := world.Biomes[world.Grid[y][x].Biome]
			snapshot.BiomeDistributions[biome.Name]++
		}
	}

	// Physics metrics
	if world.PhysicsSystem != nil {
		snapshot.PhysicsMetrics = sr.calculatePhysicsMetrics(world)
	}

	// Communication metrics
	if world.CommunicationSystem != nil {
		snapshot.CommunicationMetrics = sr.calculateCommunicationMetrics(world)
	}

	sr.addSnapshot(snapshot)
	sr.lastSnapshot = &snapshot

	// Set energy baseline on first snapshot
	if sr.totalEnergyBaseline == 0 {
		sr.totalEnergyBaseline = totalEnergy
	}
}

// calculatePhysicsMetrics computes physics-related metrics
func (sr *StatisticalReporter) calculatePhysicsMetrics(world *World) PhysicsSnapshot {
	totalMomentum := 0.0
	totalKineticEnergy := 0.0
	totalVelocity := 0.0
	activeEntities := 0

	for _, entity := range world.AllEntities {
		if entity.IsAlive {
			if physics, exists := world.PhysicsComponents[entity.ID]; exists {
				velocity := math.Sqrt(physics.Velocity.X*physics.Velocity.X + physics.Velocity.Y*physics.Velocity.Y)
				mass := entity.GetTrait("size") // Use size as mass
				
				totalMomentum += mass * velocity
				totalKineticEnergy += 0.5 * mass * velocity * velocity
				totalVelocity += velocity
				activeEntities++
			}
		}
	}

	avgVelocity := 0.0
	if activeEntities > 0 {
		avgVelocity = totalVelocity / float64(activeEntities)
	}

	return PhysicsSnapshot{
		TotalMomentum:       totalMomentum,
		TotalKineticEnergy:  totalKineticEnergy,
		CollisionCount:      world.PhysicsSystem.CollisionsThisTick,
		AverageVelocity:     avgVelocity,
	}
}

// calculateCommunicationMetrics computes communication-related metrics
func (sr *StatisticalReporter) calculateCommunicationMetrics(world *World) CommunicationSnapshot {
	activeSignals := len(world.CommunicationSystem.Signals)
	signalsByType := make(map[string]int)

	for _, signal := range world.CommunicationSystem.Signals {
		signalType := fmt.Sprintf("%v", signal.Type)
		signalsByType[signalType]++
	}

	// Calculate signal efficiency (simplified metric)
	efficiency := 0.0
	if activeSignals > 0 && len(world.AllEntities) > 0 {
		efficiency = float64(activeSignals) / float64(len(world.AllEntities))
	}

	return CommunicationSnapshot{
		ActiveSignals:    activeSignals,
		SignalsByType:    signalsByType,
		SignalEfficiency: efficiency,
	}
}

// PerformAnalysis analyzes collected data for anomalies and issues
func (sr *StatisticalReporter) PerformAnalysis(world *World) []Anomaly {
	newAnomalies := make([]Anomaly, 0)

	// Only analyze if we have enough data
	if len(sr.Snapshots) < 2 {
		return newAnomalies
	}

	// Energy conservation analysis
	if anomaly := sr.analyzeEnergyConservation(world); anomaly != nil {
		newAnomalies = append(newAnomalies, *anomaly)
	}

	// Trait distribution analysis
	if anomalies := sr.analyzeTraitDistributions(); len(anomalies) > 0 {
		newAnomalies = append(newAnomalies, anomalies...)
	}

	// Population dynamics analysis
	if anomaly := sr.analyzePopulationDynamics(); anomaly != nil {
		newAnomalies = append(newAnomalies, *anomaly)
	}

	// Physics conservation analysis
	if anomaly := sr.analyzePhysicsConservation(); anomaly != nil {
		newAnomalies = append(newAnomalies, *anomaly)
	}

	// Add new anomalies to collection
	for _, anomaly := range newAnomalies {
		sr.addAnomaly(anomaly)
	}

	return newAnomalies
}

// analyzeEnergyConservation checks for energy conservation violations
func (sr *StatisticalReporter) analyzeEnergyConservation(world *World) *Anomaly {
	if len(sr.Snapshots) < 2 {
		return nil
	}

	current := sr.Snapshots[len(sr.Snapshots)-1]
	previous := sr.Snapshots[len(sr.Snapshots)-2]

	energyChange := current.TotalEnergy - previous.TotalEnergy
	allowedDeviation := sr.totalEnergyBaseline * 0.1 // 10% allowed deviation

	// Check for unrealistic energy changes
	if math.Abs(energyChange) > allowedDeviation {
		severity := math.Min(1.0, math.Abs(energyChange)/allowedDeviation)
		
		return &Anomaly{
			Type:        AnomalyEnergyConservation,
			Severity:    severity,
			Tick:        current.Tick,
			Description: fmt.Sprintf("Large energy change detected: %.2f (%.1f%% of baseline)", energyChange, (energyChange/sr.totalEnergyBaseline)*100),
			Data: map[string]interface{}{
				"energy_change":     energyChange,
				"previous_energy":   previous.TotalEnergy,
				"current_energy":    current.TotalEnergy,
				"baseline_energy":   sr.totalEnergyBaseline,
				"allowed_deviation": allowedDeviation,
			},
			Confidence: 0.8,
		}
	}

	return nil
}

// analyzeTraitDistributions checks for unrealistic trait distributions
func (sr *StatisticalReporter) analyzeTraitDistributions() []Anomaly {
	anomalies := make([]Anomaly, 0)

	if len(sr.Snapshots) == 0 {
		return anomalies
	}

	current := sr.Snapshots[len(sr.Snapshots)-1]

	for trait, values := range current.TraitDistributions {
		if len(values) < 10 { // Need enough data points
			continue
		}

		// Calculate basic statistics
		mean, stdDev := sr.calculateMeanAndStdDev(values)

		// Check for unrealistic distributions
		if stdDev < 0.01 { // Too uniform
			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyUnrealisticDistribution,
				Severity:    0.6,
				Tick:        current.Tick,
				Description: fmt.Sprintf("Trait '%s' has unrealistically uniform distribution (stddev=%.4f)", trait, stdDev),
				Data: map[string]interface{}{
					"trait":     trait,
					"mean":      mean,
					"std_dev":   stdDev,
					"count":     len(values),
				},
				Confidence: 0.7,
			})
		}

		// Check for values outside reasonable bounds
		outOfBounds := 0
		for _, value := range values {
			if value < 0 || value > 1 {
				outOfBounds++
			}
		}

		if outOfBounds > 0 {
			severity := float64(outOfBounds) / float64(len(values))
			anomalies = append(anomalies, Anomaly{
				Type:        AnomalyBiologicalImplausibility,
				Severity:    severity,
				Tick:        current.Tick,
				Description: fmt.Sprintf("Trait '%s' has %d values outside normal bounds [0,1]", trait, outOfBounds),
				Data: map[string]interface{}{
					"trait":          trait,
					"out_of_bounds":  outOfBounds,
					"total_values":   len(values),
					"violation_rate": severity,
				},
				Confidence: 0.9,
			})
		}
	}

	return anomalies
}

// analyzePopulationDynamics checks for unrealistic population changes
func (sr *StatisticalReporter) analyzePopulationDynamics() *Anomaly {
	if len(sr.Snapshots) < 3 {
		return nil
	}

	current := sr.Snapshots[len(sr.Snapshots)-1]
	previous := sr.Snapshots[len(sr.Snapshots)-2]

	totalChange := current.TotalEntities - previous.TotalEntities
	changeRate := float64(totalChange) / float64(previous.TotalEntities)

	// Check for extreme population changes (>50% in one snapshot interval)
	if math.Abs(changeRate) > 0.5 {
		severity := math.Min(1.0, math.Abs(changeRate))
		
		return &Anomaly{
			Type:        AnomalyPopulationAnomaly,
			Severity:    severity,
			Tick:        current.Tick,
			Description: fmt.Sprintf("Extreme population change: %d entities (%.1f%% change)", totalChange, changeRate*100),
			Data: map[string]interface{}{
				"population_change": totalChange,
				"change_rate":       changeRate,
				"previous_pop":      previous.TotalEntities,
				"current_pop":       current.TotalEntities,
			},
			Confidence: 0.8,
		}
	}

	return nil
}

// analyzePhysicsConservation checks for physics law violations
func (sr *StatisticalReporter) analyzePhysicsConservation() *Anomaly {
	if len(sr.Snapshots) < 2 {
		return nil
	}

	current := sr.Snapshots[len(sr.Snapshots)-1]
	previous := sr.Snapshots[len(sr.Snapshots)-2]

	// Check momentum conservation (should be relatively stable in closed system)
	momentumChange := math.Abs(current.PhysicsMetrics.TotalMomentum - previous.PhysicsMetrics.TotalMomentum)
	
	// Allow some deviation due to system interactions, but flag large changes
	if momentumChange > previous.PhysicsMetrics.TotalMomentum*0.5 {
		severity := math.Min(1.0, momentumChange/previous.PhysicsMetrics.TotalMomentum)
		
		return &Anomaly{
			Type:        AnomalyPhysicsViolation,
			Severity:    severity,
			Tick:        current.Tick,
			Description: fmt.Sprintf("Large momentum change detected: %.4f", momentumChange),
			Data: map[string]interface{}{
				"momentum_change":    momentumChange,
				"previous_momentum":  previous.PhysicsMetrics.TotalMomentum,
				"current_momentum":   current.PhysicsMetrics.TotalMomentum,
			},
			Confidence: 0.6,
		}
	}

	return nil
}

// calculateMeanAndStdDev calculates mean and standard deviation
func (sr *StatisticalReporter) calculateMeanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(values))
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}

// ExportToCSV exports collected data to CSV format
func (sr *StatisticalReporter) ExportToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := []string{"Tick", "Timestamp", "EventType", "Category", "EntityID", "PlantID", 
					   "OldValue", "NewValue", "Change", "Metadata"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write events
	for _, event := range sr.Events {
		record := []string{
			strconv.Itoa(event.Tick),
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			event.Category,
			strconv.Itoa(event.EntityID),
			strconv.Itoa(event.PlantID),
			fmt.Sprintf("%v", event.OldValue),
			fmt.Sprintf("%v", event.NewValue),
			strconv.FormatFloat(event.Change, 'f', 6, 64),
			fmt.Sprintf("%v", event.Metadata),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportToJSON exports all data to JSON format
func (sr *StatisticalReporter) ExportToJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sr)
}

// GetAnomaliesByType returns anomalies of a specific type
func (sr *StatisticalReporter) GetAnomaliesByType(anomalyType AnomalyType) []Anomaly {
	filtered := make([]Anomaly, 0)
	for _, anomaly := range sr.Anomalies {
		if anomaly.Type == anomalyType {
			filtered = append(filtered, anomaly)
		}
	}
	return filtered
}

// GetRecentAnomalies returns anomalies from recent ticks
func (sr *StatisticalReporter) GetRecentAnomalies(ticksBack int, currentTick int) []Anomaly {
	filtered := make([]Anomaly, 0)
	cutoffTick := currentTick - ticksBack
	
	for _, anomaly := range sr.Anomalies {
		if anomaly.Tick >= cutoffTick {
			filtered = append(filtered, anomaly)
		}
	}
	
	// Sort by tick (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Tick > filtered[j].Tick
	})
	
	return filtered
}

// GetSummaryStatistics returns summary statistics for the current state
func (sr *StatisticalReporter) GetSummaryStatistics() map[string]interface{} {
	if len(sr.Snapshots) == 0 {
		return make(map[string]interface{})
	}

	latest := sr.Snapshots[len(sr.Snapshots)-1]
	
	stats := map[string]interface{}{
		"total_events":        len(sr.Events),
		"total_snapshots":     len(sr.Snapshots),
		"total_anomalies":     len(sr.Anomalies),
		"latest_tick":         latest.Tick,
		"total_entities":      latest.TotalEntities,
		"total_plants":        latest.TotalPlants,
		"total_energy":        latest.TotalEnergy,
		"species_count":       latest.SpeciesCount,
		"energy_baseline":     sr.totalEnergyBaseline,
		"anomaly_breakdown":   sr.detectedAnomalies,
	}

	// Add trend analysis if we have enough data
	if len(sr.Snapshots) >= 5 {
		stats["energy_trend"] = sr.calculateEnergyTrend()
		stats["population_trend"] = sr.calculatePopulationTrend()
	}

	return stats
}

// calculateEnergyTrend calculates energy change trend
func (sr *StatisticalReporter) calculateEnergyTrend() string {
	if len(sr.Snapshots) < 5 {
		return "insufficient_data"
	}

	recent := sr.Snapshots[len(sr.Snapshots)-5:]
	
	// Simple linear trend calculation
	totalChange := recent[len(recent)-1].TotalEnergy - recent[0].TotalEnergy
	
	if totalChange > sr.totalEnergyBaseline*0.05 {
		return "increasing"
	} else if totalChange < -sr.totalEnergyBaseline*0.05 {
		return "decreasing"
	} else {
		return "stable"
	}
}

// calculatePopulationTrend calculates population change trend
func (sr *StatisticalReporter) calculatePopulationTrend() string {
	if len(sr.Snapshots) < 5 {
		return "insufficient_data"
	}

	recent := sr.Snapshots[len(sr.Snapshots)-5:]
	
	totalChange := recent[len(recent)-1].TotalEntities - recent[0].TotalEntities
	percentChange := float64(totalChange) / float64(recent[0].TotalEntities) * 100

	if percentChange > 10 {
		return "growing"
	} else if percentChange < -10 {
		return "declining"
	} else {
		return "stable"
	}
}

// Helper methods for managing collections
func (sr *StatisticalReporter) addEvent(event StatisticalEvent) {
	sr.Events = append(sr.Events, event)
	
	// Remove old events if we exceed max
	if len(sr.Events) > sr.MaxEvents {
		sr.Events = sr.Events[len(sr.Events)-sr.MaxEvents:]
	}
}

func (sr *StatisticalReporter) addSnapshot(snapshot StatisticalSnapshot) {
	sr.Snapshots = append(sr.Snapshots, snapshot)
	
	// Remove old snapshots if we exceed max
	if len(sr.Snapshots) > sr.MaxSnapshots {
		sr.Snapshots = sr.Snapshots[len(sr.Snapshots)-sr.MaxSnapshots:]
	}
}

func (sr *StatisticalReporter) addAnomaly(anomaly Anomaly) {
	sr.Anomalies = append(sr.Anomalies, anomaly)
	sr.detectedAnomalies[anomaly.Type]++
}