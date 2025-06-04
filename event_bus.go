package main

import (
	"encoding/json"
	"sync"
	"time"
)

// CentralEvent represents a unified event in the simulation
type CentralEvent struct {
	ID           int                    `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Tick         int                    `json:"tick"`
	Type         string                 `json:"type"`
	Category     string                 `json:"category"`     // "entity", "plant", "environment", "system", "physics", "communication", etc.
	SubCategory  string                 `json:"sub_category"` // More specific classification
	Source       string                 `json:"source"`       // Which system generated the event
	Description  string                 `json:"description"`
	EntityID     int                    `json:"entity_id,omitempty"`
	PlantID      int                    `json:"plant_id,omitempty"`
	Position     *Position              `json:"position,omitempty"`
	OldValue     interface{}            `json:"old_value,omitempty"`
	NewValue     interface{}            `json:"new_value,omitempty"`
	Change       float64                `json:"change,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	ImpactedIDs  []int                  `json:"impacted_ids,omitempty"`
	Severity     string                 `json:"severity"` // "low", "medium", "high", "critical"
}

// EventBusListener represents a listener function for events
type EventBusListener func(event CentralEvent)

// CentralEventBus manages all events in the simulation in chronological order
type CentralEventBus struct {
	events       []CentralEvent
	listeners    []EventBusListener
	maxEvents    int
	nextID       int
	mutex        sync.RWMutex
	
	// Event filtering and querying
	eventsByType     map[string][]int // Maps event type to event indices
	eventsByCategory map[string][]int // Maps category to event indices
	eventsByTick     map[int][]int    // Maps tick to event indices
}

// NewCentralEventBus creates a new central event bus
func NewCentralEventBus(maxEvents int) *CentralEventBus {
	return &CentralEventBus{
		events:           make([]CentralEvent, 0),
		listeners:        make([]EventBusListener, 0),
		maxEvents:        maxEvents,
		nextID:           1,
		eventsByType:     make(map[string][]int),
		eventsByCategory: make(map[string][]int),
		eventsByTick:     make(map[int][]int),
	}
}

// AddListener adds a listener function that will be called for all events
func (eb *CentralEventBus) AddListener(listener EventBusListener) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	eb.listeners = append(eb.listeners, listener)
}

// EmitEvent adds a new event to the bus and notifies all listeners
func (eb *CentralEventBus) EmitEvent(tick int, eventType, category, subCategory, source, description string, metadata map[string]interface{}) {
	event := CentralEvent{
		ID:          eb.nextID,
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        eventType,
		Category:    category,
		SubCategory: subCategory,
		Source:      source,
		Description: description,
		Metadata:    metadata,
		Severity:    "medium", // Default severity
	}
	
	eb.addEvent(event)
}

// EmitEntityEvent emits an event related to a specific entity
func (eb *CentralEventBus) EmitEntityEvent(tick int, eventType, subCategory, source, description string, entity *Entity, oldValue, newValue interface{}, impactedEntities []*Entity) {
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

	metadata := map[string]interface{}{
		"species":    entity.Species,
		"energy":     entity.Energy,
		"age":        entity.Age,
		"is_alive":   entity.IsAlive,
	}

	event := CentralEvent{
		ID:          eb.nextID,
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        eventType,
		Category:    "entity",
		SubCategory: subCategory,
		Source:      source,
		Description: description,
		EntityID:    entity.ID,
		Position:    &entity.Position,
		OldValue:    oldValue,
		NewValue:    newValue,
		Change:      change,
		Metadata:    metadata,
		ImpactedIDs: impactedIDs,
		Severity:    eb.determineSeverity(eventType, "entity", change),
	}

	eb.addEvent(event)
}

// EmitPlantEvent emits an event related to a specific plant
func (eb *CentralEventBus) EmitPlantEvent(tick int, eventType, subCategory, source, description string, plant *Plant, oldValue, newValue interface{}) {
	var change float64
	if oldVal, ok := oldValue.(float64); ok {
		if newVal, ok := newValue.(float64); ok {
			change = newVal - oldVal
		}
	}

	metadata := map[string]interface{}{
		"type":      plant.Type,
		"energy":    plant.Energy,
		"age":       plant.Age,
		"is_alive":  plant.IsAlive,
		"size":      plant.Size,
	}

	event := CentralEvent{
		ID:          eb.nextID,
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        eventType,
		Category:    "plant",
		SubCategory: subCategory,
		Source:      source,
		Description: description,
		PlantID:     plant.ID,
		Position:    &plant.Position,
		OldValue:    oldValue,
		NewValue:    newValue,
		Change:      change,
		Metadata:    metadata,
		Severity:    eb.determineSeverity(eventType, "plant", change),
	}

	eb.addEvent(event)
}

// EmitSystemEvent emits a system-wide event
func (eb *CentralEventBus) EmitSystemEvent(tick int, eventType, subCategory, source, description string, position *Position, metadata map[string]interface{}) {
	event := CentralEvent{
		ID:          eb.nextID,
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        eventType,
		Category:    "system",
		SubCategory: subCategory,
		Source:      source,
		Description: description,
		Position:    position,
		Metadata:    metadata,
		Severity:    eb.determineSeverity(eventType, "system", 0),
	}

	eb.addEvent(event)
}

// addEvent adds an event to the bus and maintains indices
func (eb *CentralEventBus) addEvent(event CentralEvent) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// Add event to main list
	eb.events = append(eb.events, event)
	eb.nextID++

	// Update indices
	eventIndex := len(eb.events) - 1
	
	// Index by type
	eb.eventsByType[event.Type] = append(eb.eventsByType[event.Type], eventIndex)
	
	// Index by category
	eb.eventsByCategory[event.Category] = append(eb.eventsByCategory[event.Category], eventIndex)
	
	// Index by tick
	eb.eventsByTick[event.Tick] = append(eb.eventsByTick[event.Tick], eventIndex)

	// Maintain max events limit
	if len(eb.events) > eb.maxEvents {
		eb.removeOldestEvent()
	}

	// Notify listeners
	for _, listener := range eb.listeners {
		listener(event)
	}
}

// removeOldestEvent removes the oldest event and updates indices
func (eb *CentralEventBus) removeOldestEvent() {
	if len(eb.events) == 0 {
		return
	}

	// Remove from main list
	eb.events = eb.events[1:]

	// Update all indices by decrementing them
	for eventType, indices := range eb.eventsByType {
		newIndices := make([]int, 0)
		for _, index := range indices {
			if index > 0 {
				newIndices = append(newIndices, index-1)
			}
		}
		if len(newIndices) == 0 {
			delete(eb.eventsByType, eventType)
		} else {
			eb.eventsByType[eventType] = newIndices
		}
	}

	for category, indices := range eb.eventsByCategory {
		newIndices := make([]int, 0)
		for _, index := range indices {
			if index > 0 {
				newIndices = append(newIndices, index-1)
			}
		}
		if len(newIndices) == 0 {
			delete(eb.eventsByCategory, category)
		} else {
			eb.eventsByCategory[category] = newIndices
		}
	}

	for tick, indices := range eb.eventsByTick {
		newIndices := make([]int, 0)
		for _, index := range indices {
			if index > 0 {
				newIndices = append(newIndices, index-1)
			}
		}
		if len(newIndices) == 0 {
			delete(eb.eventsByTick, tick)
		} else {
			eb.eventsByTick[tick] = newIndices
		}
	}
}

// determineSeverity determines event severity based on type and context
func (eb *CentralEventBus) determineSeverity(eventType, category string, change float64) string {
	// Critical events
	if eventType == "extinction" || eventType == "system_failure" || eventType == "critical_error" {
		return "critical"
	}
	
	// High severity events
	if eventType == "death" || eventType == "birth" || eventType == "evolution" || eventType == "speciation" {
		return "high"
	}
	
	// Check for large changes
	if change != 0 && (change > 100 || change < -100) {
		return "high"
	}
	
	// Medium severity for most other events
	if eventType == "reproduction" || eventType == "communication" || eventType == "movement" {
		return "medium"
	}
	
	// Low severity for routine events
	return "low"
}

// GetAllEvents returns all events in chronological order
func (eb *CentralEventBus) GetAllEvents() []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	events := make([]CentralEvent, len(eb.events))
	copy(events, eb.events)
	return events
}

// GetEventsByType returns events of a specific type
func (eb *CentralEventBus) GetEventsByType(eventType string) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	indices, exists := eb.eventsByType[eventType]
	if !exists {
		return []CentralEvent{}
	}
	
	events := make([]CentralEvent, len(indices))
	for i, index := range indices {
		events[i] = eb.events[index]
	}
	return events
}

// GetEventsByCategory returns events of a specific category
func (eb *CentralEventBus) GetEventsByCategory(category string) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	indices, exists := eb.eventsByCategory[category]
	if !exists {
		return []CentralEvent{}
	}
	
	events := make([]CentralEvent, len(indices))
	for i, index := range indices {
		events[i] = eb.events[index]
	}
	return events
}

// GetEventsByTick returns events from a specific tick
func (eb *CentralEventBus) GetEventsByTick(tick int) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	indices, exists := eb.eventsByTick[tick]
	if !exists {
		return []CentralEvent{}
	}
	
	events := make([]CentralEvent, len(indices))
	for i, index := range indices {
		events[i] = eb.events[index]
	}
	return events
}

// GetEventsSince returns events since a specific tick
func (eb *CentralEventBus) GetEventsSince(tick int) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	var events []CentralEvent
	for _, event := range eb.events {
		if event.Tick >= tick {
			events = append(events, event)
		}
	}
	return events
}

// GetRecentEvents returns the most recent N events
func (eb *CentralEventBus) GetRecentEvents(count int) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	if count >= len(eb.events) {
		events := make([]CentralEvent, len(eb.events))
		copy(events, eb.events)
		return events
	}
	
	events := make([]CentralEvent, count)
	copy(events, eb.events[len(eb.events)-count:])
	return events
}

// GetEventsBySeverity returns events of a specific severity level
func (eb *CentralEventBus) GetEventsBySeverity(severity string) []CentralEvent {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	var events []CentralEvent
	for _, event := range eb.events {
		if event.Severity == severity {
			events = append(events, event)
		}
	}
	return events
}

// GetEventStats returns statistics about the events in the bus
func (eb *CentralEventBus) GetEventStats() map[string]interface{} {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_events": len(eb.events),
		"max_events":   eb.maxEvents,
		"next_id":      eb.nextID,
	}
	
	// Count by type
	typeStats := make(map[string]int)
	for eventType, indices := range eb.eventsByType {
		typeStats[eventType] = len(indices)
	}
	stats["events_by_type"] = typeStats
	
	// Count by category
	categoryStats := make(map[string]int)
	for category, indices := range eb.eventsByCategory {
		categoryStats[category] = len(indices)
	}
	stats["events_by_category"] = categoryStats
	
	// Count by severity
	severityStats := make(map[string]int)
	for _, event := range eb.events {
		severityStats[event.Severity]++
	}
	stats["events_by_severity"] = severityStats
	
	return stats
}

// ExportToJSON exports all events to JSON format
func (eb *CentralEventBus) ExportToJSON() ([]byte, error) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	
	exportData := map[string]interface{}{
		"events":      eb.events,
		"statistics":  eb.GetEventStats(),
		"export_time": time.Now(),
	}
	
	return json.MarshalIndent(exportData, "", "  ")
}

// ClearEvents removes all events from the bus
func (eb *CentralEventBus) ClearEvents() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	
	eb.events = make([]CentralEvent, 0)
	eb.eventsByType = make(map[string][]int)
	eb.eventsByCategory = make(map[string][]int)
	eb.eventsByTick = make(map[int][]int)
	eb.nextID = 1
}