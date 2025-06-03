package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Player represents a web client who can control species in the simulation
type Player struct {
	ID           string    `json:"id"`           // Unique player identifier
	Name         string    `json:"name"`         // Player display name
	ConnectedAt  time.Time `json:"connected_at"` // When player connected
	Species      []string  `json:"species"`      // Species names controlled by this player
	IsActive     bool      `json:"is_active"`    // Whether player is currently connected
	LastActivity time.Time `json:"last_activity"` // Last action timestamp
}

// PlayerSpecies tracks a species owned by a player
type PlayerSpecies struct {
	PlayerID    string    `json:"player_id"`     // ID of controlling player
	SpeciesName string    `json:"species_name"`  // Name of the species
	CreatedAt   time.Time `json:"created_at"`    // When species was created
	IsExtinct   bool      `json:"is_extinct"`    // Whether species has died out
	SubSpecies  []string  `json:"sub_species"`   // Any sub-species that split off
}

// PlayerManager manages all players and their species ownership
type PlayerManager struct {
	Players        map[string]*Player        `json:"players"`         // PlayerID -> Player
	PlayerSpecies  map[string]*PlayerSpecies `json:"player_species"`  // SpeciesName -> PlayerSpecies
	ActivePlayers  map[string]bool           `json:"active_players"`  // Currently connected players
}

// NewPlayerManager creates a new player manager
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players:       make(map[string]*Player),
		PlayerSpecies: make(map[string]*PlayerSpecies),
		ActivePlayers: make(map[string]bool),
	}
}

// ValidatePlayerName ensures the player name only contains alphanumeric characters
// and removes redundant spaces
func ValidatePlayerName(name string) (string, error) {
	// Remove leading/trailing spaces and collapse multiple spaces
	cleanName := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(name, " "))
	
	// Check if empty after cleaning
	if cleanName == "" {
		return "", fmt.Errorf("player name cannot be empty")
	}
	
	// Check for alphanumeric characters and spaces only
	if !regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString(cleanName) {
		return "", fmt.Errorf("player name can only contain letters, numbers, and spaces")
	}
	
	// Check reasonable length limits
	if len(cleanName) < 2 {
		return "", fmt.Errorf("player name must be at least 2 characters long")
	}
	
	if len(cleanName) > 50 {
		return "", fmt.Errorf("player name cannot exceed 50 characters")
	}
	
	return cleanName, nil
}

// AddPlayer adds a new player to the manager
func (pm *PlayerManager) AddPlayer(playerID, playerName string) (*Player, error) {
	// Validate and clean the player name
	cleanName, err := ValidatePlayerName(playerName)
	if err != nil {
		return nil, err
	}
	
	// Check if player already exists
	if _, exists := pm.Players[playerID]; exists {
		return pm.Players[playerID], nil
	}
	
	// Create new player
	player := &Player{
		ID:           playerID,
		Name:         cleanName,
		ConnectedAt:  time.Now(),
		Species:      make([]string, 0),
		IsActive:     true,
		LastActivity: time.Now(),
	}
	
	pm.Players[playerID] = player
	pm.ActivePlayers[playerID] = true
	
	return player, nil
}

// RemovePlayer removes a player and marks them as inactive
func (pm *PlayerManager) RemovePlayer(playerID string) {
	if player, exists := pm.Players[playerID]; exists {
		player.IsActive = false
		delete(pm.ActivePlayers, playerID)
	}
}

// AddPlayerSpecies assigns a species to a player
func (pm *PlayerManager) AddPlayerSpecies(playerID, speciesName string) error {
	// Check if player exists
	if _, exists := pm.Players[playerID]; !exists {
		return fmt.Errorf("player %s not found", playerID)
	}
	
	// Check if species is already owned
	if _, exists := pm.PlayerSpecies[speciesName]; exists {
		return fmt.Errorf("species %s is already owned by another player", speciesName)
	}
	
	// Create player species mapping
	playerSpecies := &PlayerSpecies{
		PlayerID:    playerID,
		SpeciesName: speciesName,
		CreatedAt:   time.Now(),
		IsExtinct:   false,
		SubSpecies:  make([]string, 0),
	}
	
	pm.PlayerSpecies[speciesName] = playerSpecies
	pm.Players[playerID].Species = append(pm.Players[playerID].Species, speciesName)
	pm.Players[playerID].LastActivity = time.Now()
	
	return nil
}

// GetPlayerSpecies returns the species owned by a player
func (pm *PlayerManager) GetPlayerSpecies(playerID string) []string {
	if player, exists := pm.Players[playerID]; exists {
		return player.Species
	}
	return make([]string, 0)
}

// GetSpeciesOwner returns the player ID who owns a species
func (pm *PlayerManager) GetSpeciesOwner(speciesName string) (string, bool) {
	if playerSpecies, exists := pm.PlayerSpecies[speciesName]; exists {
		return playerSpecies.PlayerID, true
	}
	return "", false
}

// CanPlayerControlSpecies checks if a player can control a specific species
func (pm *PlayerManager) CanPlayerControlSpecies(playerID, speciesName string) bool {
	ownerID, exists := pm.GetSpeciesOwner(speciesName)
	return exists && ownerID == playerID
}

// MarkSpeciesExtinct marks a species as extinct
func (pm *PlayerManager) MarkSpeciesExtinct(speciesName string) {
	if playerSpecies, exists := pm.PlayerSpecies[speciesName]; exists {
		playerSpecies.IsExtinct = true
	}
}

// AddSubSpecies adds a sub-species to the parent species record
func (pm *PlayerManager) AddSubSpecies(parentSpecies, subSpeciesName string) error {
	if playerSpecies, exists := pm.PlayerSpecies[parentSpecies]; exists {
		// Add the sub-species to the parent
		playerSpecies.SubSpecies = append(playerSpecies.SubSpecies, subSpeciesName)
		
		// Create a new species record for the sub-species with the same owner
		subPlayerSpecies := &PlayerSpecies{
			PlayerID:    playerSpecies.PlayerID,
			SpeciesName: subSpeciesName,
			CreatedAt:   time.Now(),
			IsExtinct:   false,
			SubSpecies:  make([]string, 0),
		}
		
		pm.PlayerSpecies[subSpeciesName] = subPlayerSpecies
		pm.Players[playerSpecies.PlayerID].Species = append(pm.Players[playerSpecies.PlayerID].Species, subSpeciesName)
		
		return nil
	}
	return fmt.Errorf("parent species %s not found", parentSpecies)
}

// GetActivePlayers returns a list of currently active players
func (pm *PlayerManager) GetActivePlayers() []*Player {
	players := make([]*Player, 0, len(pm.ActivePlayers))
	for playerID := range pm.ActivePlayers {
		if player, exists := pm.Players[playerID]; exists {
			players = append(players, player)
		}
	}
	return players
}

// UpdatePlayerActivity updates the last activity timestamp for a player
func (pm *PlayerManager) UpdatePlayerActivity(playerID string) {
	if player, exists := pm.Players[playerID]; exists {
		player.LastActivity = time.Now()
	}
}

// GetPlayerStats returns statistics about a player's species
func (pm *PlayerManager) GetPlayerStats(playerID string) map[string]interface{} {
	stats := make(map[string]interface{})
	
	if player, exists := pm.Players[playerID]; exists {
		stats["player_name"] = player.Name
		stats["connected_at"] = player.ConnectedAt
		stats["species_count"] = len(player.Species)
		stats["is_active"] = player.IsActive
		stats["last_activity"] = player.LastActivity
		
		// Count extinct vs active species
		extinctCount := 0
		activeCount := 0
		subSpeciesCount := 0
		
		for _, speciesName := range player.Species {
			if playerSpecies, exists := pm.PlayerSpecies[speciesName]; exists {
				if playerSpecies.IsExtinct {
					extinctCount++
				} else {
					activeCount++
				}
				subSpeciesCount += len(playerSpecies.SubSpecies)
			}
		}
		
		stats["extinct_species"] = extinctCount
		stats["active_species"] = activeCount
		stats["sub_species"] = subSpeciesCount
	}
	
	return stats
}