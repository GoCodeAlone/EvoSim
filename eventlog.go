package main

import (
	"fmt"
	"time"
)

// LogEvent represents a significant event in the ecosystem
type LogEvent struct {
	Timestamp   time.Time
	Tick        int
	Type        string
	Description string
	Data        map[string]interface{}
}

// LogEventType constants for different event types
const (
	EventSpeciesExtinction = "species_extinction"
	EventSpeciesEvolution  = "species_evolution"
	EventWorldEvent        = "world_event"
	EventPopulationBoom    = "population_boom"
	EventPopulationCrash   = "population_crash"
	EventNewSpecies        = "new_species"
	EventMutation          = "major_mutation"
	EventPlantEvolution    = "plant_evolution"
	EventEcosystemShift    = "ecosystem_shift"
)

// EventLogger manages the ecosystem event log
type EventLogger struct {
	Events         []LogEvent
	MaxEvents      int
	lastPopCounts  map[string]int
	lastPlantCount int
}

// NewEventLogger creates a new event logger
func NewEventLogger(maxEvents int) *EventLogger {
	return &EventLogger{
		Events:        make([]LogEvent, 0),
		MaxEvents:     maxEvents,
		lastPopCounts: make(map[string]int),
	}
}

// LogSpeciesExtinction records when a species goes extinct
func (el *EventLogger) LogSpeciesExtinction(tick int, species string, lastCount int) {
	event := LogEvent{
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        EventSpeciesExtinction,
		Description: fmt.Sprintf("Species %s has gone EXTINCT (last count: %d)", species, lastCount),
		Data: map[string]interface{}{
			"species":    species,
			"last_count": lastCount,
		},
	}
	el.addEvent(event)
}

// LogPopulationChange records significant population changes
func (el *EventLogger) LogPopulationChange(tick int, species string, oldCount, newCount int) {
	threshold := 0.5 // 50% change threshold
	if oldCount == 0 && newCount > 0 {
		// Species emergence/revival
		event := LogEvent{
			Timestamp:   time.Now(),
			Tick:        tick,
			Type:        EventNewSpecies,
			Description: fmt.Sprintf("Species %s emerged with %d individuals", species, newCount),
			Data: map[string]interface{}{
				"species":   species,
				"new_count": newCount,
			},
		}
		el.addEvent(event)
		return
	}

	if oldCount > 0 {
		change := float64(newCount-oldCount) / float64(oldCount)
		if change > threshold {
			event := LogEvent{
				Timestamp:   time.Now(),
				Tick:        tick,
				Type:        EventPopulationBoom,
				Description: fmt.Sprintf("Population BOOM: %s increased from %d to %d (+%.1f%%)", species, oldCount, newCount, change*100),
				Data: map[string]interface{}{
					"species":    species,
					"old_count":  oldCount,
					"new_count":  newCount,
					"change_pct": change * 100,
				},
			}
			el.addEvent(event)
		} else if change < -threshold {
			event := LogEvent{
				Timestamp:   time.Now(),
				Tick:        tick,
				Type:        EventPopulationCrash,
				Description: fmt.Sprintf("Population CRASH: %s decreased from %d to %d (%.1f%%)", species, oldCount, newCount, change*100),
				Data: map[string]interface{}{
					"species":    species,
					"old_count":  oldCount,
					"new_count":  newCount,
					"change_pct": change * 100,
				},
			}
			el.addEvent(event)
		}
	}
}

// LogWorldEvent records world events
func (el *EventLogger) LogWorldEvent(tick int, eventName, description string) {
	event := LogEvent{
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        EventWorldEvent,
		Description: fmt.Sprintf("World Event: %s - %s", eventName, description),
		Data: map[string]interface{}{
			"event_name":        eventName,
			"event_description": description,
		},
	}
	el.addEvent(event)
}

// LogSpeciesEvolution records when a species undergoes major evolution
func (el *EventLogger) LogSpeciesEvolution(tick int, species string, fromSpecies string, details string) {
	event := LogEvent{
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        EventSpeciesEvolution,
		Description: fmt.Sprintf("EVOLUTION: %s evolved from %s - %s", species, fromSpecies, details),
		Data: map[string]interface{}{
			"species":      species,
			"from_species": fromSpecies,
			"details":      details,
		},
	}
	el.addEvent(event)
}

// LogEcosystemShift records major ecosystem changes
func (el *EventLogger) LogEcosystemShift(tick int, description string, data map[string]interface{}) {
	event := LogEvent{
		Timestamp:   time.Now(),
		Tick:        tick,
		Type:        EventEcosystemShift,
		Description: fmt.Sprintf("Ecosystem Shift: %s", description),
		Data:        data,
	}
	el.addEvent(event)
}

// UpdatePopulationCounts checks for population changes and logs significant ones
func (el *EventLogger) UpdatePopulationCounts(tick int, populations map[string]*Population) {
	currentCounts := make(map[string]int)

	// Count current populations
	for species, pop := range populations {
		aliveCount := 0
		for _, entity := range pop.Entities {
			if entity.IsAlive {
				aliveCount++
			}
		}
		currentCounts[species] = aliveCount
	}

	// Check for changes
	for species, currentCount := range currentCounts {
		if lastCount, exists := el.lastPopCounts[species]; exists {
			if currentCount == 0 && lastCount > 0 {
				// Species extinction
				el.LogSpeciesExtinction(tick, species, lastCount)
			} else if currentCount != lastCount {
				// Population change
				el.LogPopulationChange(tick, species, lastCount, currentCount)
			}
		}
	}

	// Check for extinct species that are no longer in populations
	for species, lastCount := range el.lastPopCounts {
		if _, exists := currentCounts[species]; !exists && lastCount > 0 {
			el.LogSpeciesExtinction(tick, species, lastCount)
		}
	}

	// Update last counts
	el.lastPopCounts = currentCounts
}

// addEvent adds an event to the log, maintaining max size
func (el *EventLogger) addEvent(event LogEvent) {
	el.Events = append(el.Events, event)

	// Remove old events if we exceed max
	if len(el.Events) > el.MaxEvents {
		el.Events = el.Events[len(el.Events)-el.MaxEvents:]
	}
}

// GetRecentEvents returns the most recent events
func (el *EventLogger) GetRecentEvents(count int) []LogEvent {
	if count >= len(el.Events) {
		return el.Events
	}
	return el.Events[len(el.Events)-count:]
}

// GetEventsByType returns events of a specific type
func (el *EventLogger) GetEventsByType(eventType string) []LogEvent {
	var filtered []LogEvent
	for _, event := range el.Events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventsSince returns events since a specific tick
func (el *EventLogger) GetEventsSince(tick int) []LogEvent {
	var filtered []LogEvent
	for _, event := range el.Events {
		if event.Tick >= tick {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
