package main

import (
	"testing"
	"time"
)

func TestValidatePlayerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"ValidName123", "ValidName123", false},
		{"  Valid Name  ", "Valid Name", false},
		{"Multiple   Spaces", "Multiple Spaces", false},
		{"", "", true},                    // Empty name
		{"  ", "", true},                  // Only spaces
		{"A", "", true},                   // Too short
		{"ValidName!@#", "", true},        // Invalid characters
		{"Name_With_Underscore", "", true}, // Underscore not allowed
		{"Name-With-Dash", "", true},      // Dash not allowed
		{"ValidName", "ValidName", false},
		{string(make([]byte, 51)), "", true}, // Too long
	}

	for _, test := range tests {
		result, err := ValidatePlayerName(test.input)
		if test.hasError && err == nil {
			t.Errorf("Expected error for input '%s', but got none", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
		}
		if !test.hasError && result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func TestPlayerManager(t *testing.T) {
	pm := NewPlayerManager()

	// Test adding players
	player1, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player1: %v", err)
	}
	if player1.Name != "TestPlayer1" {
		t.Errorf("Expected player name 'TestPlayer1', got '%s'", player1.Name)
	}
	if !player1.IsActive {
		t.Error("New player should be active")
	}

	// Test adding duplicate player
	player1Again, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to handle duplicate player: %v", err)
	}
	if player1Again != player1 {
		t.Error("Should return same player instance for duplicate ID")
	}

	// Test adding player with invalid name
	_, err = pm.AddPlayer("player2", "Invalid@Name")
	if err == nil {
		t.Error("Should reject invalid player name")
	}

	// Test adding second valid player
	player2, err := pm.AddPlayer("player2", "  Valid Player 2  ")
	if err != nil {
		t.Fatalf("Failed to add player2: %v", err)
	}
	if player2.Name != "Valid Player 2" {
		t.Errorf("Expected cleaned name 'Valid Player 2', got '%s'", player2.Name)
	}

	// Test removing player
	pm.RemovePlayer("player1")
	if player1.IsActive {
		t.Error("Player should be inactive after removal")
	}
	if pm.ActivePlayers["player1"] {
		t.Error("Player should not be in active players list")
	}
}

func TestPlayerSpeciesManagement(t *testing.T) {
	pm := NewPlayerManager()

	// Add players
	_, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	_, err = pm.AddPlayer("player2", "TestPlayer2")
	if err != nil {
		t.Fatalf("Failed to add player2: %v", err)
	}

	// Test adding species to player
	err = pm.AddPlayerSpecies("player1", "Herbivores")
	if err != nil {
		t.Fatalf("Failed to add species to player: %v", err)
	}

	// Check species ownership
	ownerID, exists := pm.GetSpeciesOwner("Herbivores")
	if !exists {
		t.Error("Species should exist")
	}
	if ownerID != "player1" {
		t.Errorf("Expected owner 'player1', got '%s'", ownerID)
	}

	// Test control permissions
	if !pm.CanPlayerControlSpecies("player1", "Herbivores") {
		t.Error("Player1 should be able to control their species")
	}
	if pm.CanPlayerControlSpecies("player2", "Herbivores") {
		t.Error("Player2 should not be able to control player1's species")
	}

	// Test adding same species to different player (should fail)
	err = pm.AddPlayerSpecies("player2", "Herbivores")
	if err == nil {
		t.Error("Should not allow same species to be owned by multiple players")
	}

	// Test adding species to non-existent player
	err = pm.AddPlayerSpecies("nonexistent", "SomeSpecies")
	if err == nil {
		t.Error("Should fail when adding species to non-existent player")
	}

	// Test getting player species
	species := pm.GetPlayerSpecies("player1")
	if len(species) != 1 || species[0] != "Herbivores" {
		t.Errorf("Expected ['Herbivores'], got %v", species)
	}

	// Test getting species for non-existent player
	species = pm.GetPlayerSpecies("nonexistent")
	if len(species) != 0 {
		t.Errorf("Expected empty slice for non-existent player, got %v", species)
	}
}

func TestSpeciesExtinctionAndSubSpecies(t *testing.T) {
	pm := NewPlayerManager()

	// Add player and species
	_, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	err = pm.AddPlayerSpecies("player1", "ParentSpecies")
	if err != nil {
		t.Fatalf("Failed to add species: %v", err)
	}

	// Test marking species as extinct
	pm.MarkSpeciesExtinct("ParentSpecies")
	if playerSpecies, exists := pm.PlayerSpecies["ParentSpecies"]; exists {
		if !playerSpecies.IsExtinct {
			t.Error("Species should be marked as extinct")
		}
	} else {
		t.Error("Species should still exist when marked extinct")
	}

	// Test adding sub-species
	err = pm.AddSubSpecies("ParentSpecies", "SubSpecies1")
	if err != nil {
		t.Fatalf("Failed to add sub-species: %v", err)
	}

	// Check that sub-species is owned by same player
	ownerID, exists := pm.GetSpeciesOwner("SubSpecies1")
	if !exists {
		t.Error("Sub-species should exist")
	}
	if ownerID != "player1" {
		t.Errorf("Sub-species should be owned by same player, got '%s'", ownerID)
	}

	// Check that parent species has sub-species record
	if parentSpecies, exists := pm.PlayerSpecies["ParentSpecies"]; exists {
		if len(parentSpecies.SubSpecies) != 1 || parentSpecies.SubSpecies[0] != "SubSpecies1" {
			t.Errorf("Parent species should have sub-species record")
		}
	}

	// Check that player's species list includes sub-species
	species := pm.GetPlayerSpecies("player1")
	foundParent := false
	foundSub := false
	for _, s := range species {
		if s == "ParentSpecies" {
			foundParent = true
		}
		if s == "SubSpecies1" {
			foundSub = true
		}
	}
	if !foundParent || !foundSub {
		t.Errorf("Player should have both parent and sub-species, got %v", species)
	}

	// Test adding sub-species to non-existent parent
	err = pm.AddSubSpecies("NonExistentSpecies", "SubSpecies2")
	if err == nil {
		t.Error("Should fail when adding sub-species to non-existent parent")
	}
}

func TestPlayerStats(t *testing.T) {
	pm := NewPlayerManager()

	// Add player and species
	_, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Wait a tiny bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err = pm.AddPlayerSpecies("player1", "Species1")
	if err != nil {
		t.Fatalf("Failed to add species1: %v", err)
	}

	err = pm.AddPlayerSpecies("player1", "Species2")
	if err != nil {
		t.Fatalf("Failed to add species2: %v", err)
	}

	// Mark one species as extinct
	pm.MarkSpeciesExtinct("Species1")

	// Add sub-species
	err = pm.AddSubSpecies("Species2", "SubSpecies1")
	if err != nil {
		t.Fatalf("Failed to add sub-species: %v", err)
	}

	// Get stats
	stats := pm.GetPlayerStats("player1")

	if stats["player_name"] != "TestPlayer1" {
		t.Errorf("Wrong player name in stats: %v", stats["player_name"])
	}
	if stats["species_count"] != 3 { // Species1, Species2, SubSpecies1
		t.Errorf("Wrong species count: %v", stats["species_count"])
	}
	if stats["extinct_species"] != 1 {
		t.Errorf("Wrong extinct species count: %v", stats["extinct_species"])
	}
	if stats["active_species"] != 2 { // Species2, SubSpecies1
		t.Errorf("Wrong active species count: %v", stats["active_species"])
	}
	if stats["sub_species"] != 1 { // SubSpecies1 under Species2
		t.Errorf("Wrong sub-species count: %v", stats["sub_species"])
	}
	if stats["is_active"] != true {
		t.Error("Player should be active")
	}

	// Test stats for non-existent player
	stats = pm.GetPlayerStats("nonexistent")
	if len(stats) != 0 {
		t.Errorf("Should return empty stats for non-existent player, got %v", stats)
	}
}

func TestGetActivePlayers(t *testing.T) {
	pm := NewPlayerManager()

	// Add players
	_, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player1: %v", err)
	}

	_, err = pm.AddPlayer("player2", "TestPlayer2")
	if err != nil {
		t.Fatalf("Failed to add player2: %v", err)
	}

	// Should have 2 active players
	activePlayers := pm.GetActivePlayers()
	if len(activePlayers) != 2 {
		t.Errorf("Expected 2 active players, got %d", len(activePlayers))
	}

	// Remove one player
	pm.RemovePlayer("player1")

	// Should have 1 active player
	activePlayers = pm.GetActivePlayers()
	if len(activePlayers) != 1 {
		t.Errorf("Expected 1 active player after removal, got %d", len(activePlayers))
	}
	if activePlayers[0].ID != "player2" {
		t.Errorf("Expected remaining player to be player2, got %s", activePlayers[0].ID)
	}
}

func TestUpdatePlayerActivity(t *testing.T) {
	pm := NewPlayerManager()

	// Add player
	player, err := pm.AddPlayer("player1", "TestPlayer1")
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	originalActivity := player.LastActivity

	// Wait a bit and update activity
	time.Sleep(1 * time.Millisecond)
	pm.UpdatePlayerActivity("player1")

	if !player.LastActivity.After(originalActivity) {
		t.Error("LastActivity should be updated to a later time")
	}

	// Test updating non-existent player (should not panic)
	pm.UpdatePlayerActivity("nonexistent")
}