package main

import (
	"testing"
)

func TestViewportFunctionality(t *testing.T) {
	// Create a simple world config
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		GridWidth:      40,
		GridHeight:     25,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	webInterface := NewWebInterface(world)
	
	// Test default viewport settings
	if webInterface.viewportX != 0 {
		t.Errorf("Expected default viewportX 0, got %d", webInterface.viewportX)
	}
	if webInterface.viewportY != 0 {
		t.Errorf("Expected default viewportY 0, got %d", webInterface.viewportY)
	}
	if webInterface.zoomLevel != 1.0 {
		t.Errorf("Expected default zoomLevel 1.0, got %f", webInterface.zoomLevel)
	}
}

func TestZoomFunctionality(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		GridWidth:      40,
		GridHeight:     25,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	webInterface := NewWebInterface(world)
	
	// Test zoom in
	webInterface.zoomIn()
	if webInterface.zoomLevel != 1.5 {
		t.Errorf("Expected zoom level 1.5 after zoom in, got %f", webInterface.zoomLevel)
	}
	
	// Test zoom out
	webInterface.zoomOut()
	if webInterface.zoomLevel != 1.0 {
		t.Errorf("Expected zoom level 1.0 after zoom out, got %f", webInterface.zoomLevel)
	}
	
	// Test zoom bounds - maximum
	webInterface.setZoomLevel(10.0)
	if webInterface.zoomLevel != 8.0 {
		t.Errorf("Expected maximum zoom level 8.0, got %f", webInterface.zoomLevel)
	}
	
	// Test zoom bounds - minimum
	webInterface.setZoomLevel(0.1)
	if webInterface.zoomLevel != 0.5 {
		t.Errorf("Expected minimum zoom level 0.5, got %f", webInterface.zoomLevel)
	}
}

func TestViewportClamping(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		GridWidth:      40,
		GridHeight:     25,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	webInterface := NewWebInterface(world)
	
	// Test viewport clamping with normal zoom
	webInterface.viewportX = -10
	webInterface.viewportY = -10
	webInterface.clampViewport()
	
	if webInterface.viewportX != 0 {
		t.Errorf("Expected clamped viewportX 0, got %d", webInterface.viewportX)
	}
	if webInterface.viewportY != 0 {
		t.Errorf("Expected clamped viewportY 0, got %d", webInterface.viewportY)
	}
	
	// Test viewport clamping with high values
	webInterface.viewportX = 1000
	webInterface.viewportY = 1000
	webInterface.clampViewport()
	
	// Should be clamped to valid maximum values
	if webInterface.viewportX < 0 || webInterface.viewportX > config.GridWidth {
		t.Errorf("ViewportX %d should be clamped within bounds", webInterface.viewportX)
	}
	if webInterface.viewportY < 0 || webInterface.viewportY > config.GridHeight {
		t.Errorf("ViewportY %d should be clamped within bounds", webInterface.viewportY)
	}
}

func TestViewportWithZoom(t *testing.T) {
	config := WorldConfig{
		Width:          100.0,
		Height:         100.0,
		GridWidth:      40,
		GridHeight:     25,
		NumPopulations: 1,
		PopulationSize: 5,
	}
	
	world := NewWorld(config)
	webInterface := NewWebInterface(world)
	
	// Test viewport with 2x zoom
	webInterface.setZoomLevel(2.0)
	
	// With 2x zoom, we should see half the world, so max viewport should be smaller
	webInterface.viewportX = 1000
	webInterface.viewportY = 1000
	webInterface.clampViewport()
	
	// At 2x zoom, visible area is half the size, so max viewport should be different
	maxViewportX := config.GridWidth - int(float64(config.GridWidth)/2.0)
	maxViewportY := config.GridHeight - int(float64(config.GridHeight)/2.0)
	
	if webInterface.viewportX > maxViewportX {
		t.Errorf("ViewportX %d should not exceed max %d at 2x zoom", webInterface.viewportX, maxViewportX)
	}
	if webInterface.viewportY > maxViewportY {
		t.Errorf("ViewportY %d should not exceed max %d at 2x zoom", webInterface.viewportY, maxViewportY)
	}
}