package main

import (
	"math"
	"testing"
)

// TestBiomeDistributionRandomness tests that biomes are reasonably distributed across multiple runs
func TestBiomeDistributionRandomness(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  30,
		GridHeight: 30,
	}

	numRuns := 10
	biomeDistributions := make([]map[BiomeType]int, numRuns)
	
	// Generate multiple worlds and track biome distributions
	for run := 0; run < numRuns; run++ {
		world := NewWorld(config)
		biomeCounts := make(map[BiomeType]int)
		
		for y := 0; y < config.GridHeight; y++ {
			for x := 0; x < config.GridWidth; x++ {
				biome := world.Grid[y][x].Biome
				biomeCounts[biome]++
			}
		}
		
		biomeDistributions[run] = biomeCounts
		t.Logf("Run %d biome distribution: %v", run, biomeCounts)
	}
	
	// Check for distribution balance
	totalCells := config.GridWidth * config.GridHeight
	
	// Calculate average distribution across runs
	avgDistribution := make(map[BiomeType]float64)
	for biomeType := BiomePlains; biomeType <= BiomeCanyon; biomeType++ {
		sum := 0
		for run := 0; run < numRuns; run++ {
			sum += biomeDistributions[run][biomeType]
		}
		avgDistribution[biomeType] = float64(sum) / float64(numRuns * totalCells)
	}
	
	t.Logf("Average biome percentages across %d runs:", numRuns)
	for biomeType, percentage := range avgDistribution {
		biome := initializeBiomes()[biomeType]
		t.Logf("  %s: %.1f%%", biome.Name, percentage*100)
	}
	
	// Test 1: No single biome should dominate (> 70% of map)
	for biomeType, percentage := range avgDistribution {
		if percentage > 0.7 {
			biome := initializeBiomes()[biomeType]
			t.Errorf("Biome %s dominates the map with %.1f%% coverage (> 70%%)", biome.Name, percentage*100)
		}
	}
	
	// Test 2: Should have at least 3 different biome types represented
	biomesPresent := 0
	for _, percentage := range avgDistribution {
		if percentage > 0.01 { // At least 1% coverage
			biomesPresent++
		}
	}
	
	if biomesPresent < 3 {
		t.Errorf("Expected at least 3 biome types with >1%% coverage, got %d", biomesPresent)
	}
	
	// Test 3: Check for reasonable variation between runs (not too uniform)
	for biomeType := BiomePlains; biomeType <= BiomeCanyon; biomeType++ {
		if avgDistribution[biomeType] < 0.01 { // Skip rare biomes
			continue
		}
		
		// Calculate standard deviation for this biome type across runs
		variance := 0.0
		for run := 0; run < numRuns; run++ {
			percentage := float64(biomeDistributions[run][biomeType]) / float64(totalCells)
			diff := percentage - avgDistribution[biomeType]
			variance += diff * diff
		}
		stdDev := math.Sqrt(variance / float64(numRuns))
		
		// Standard deviation should be reasonable (not too low = too uniform, not too high = too random)
		if stdDev < 0.005 {
			biome := initializeBiomes()[biomeType]
			t.Logf("Warning: %s distribution is too uniform across runs (stddev=%.3f)", biome.Name, stdDev)
		}
		if stdDev > 0.3 {
			biome := initializeBiomes()[biomeType]
			t.Errorf("%s distribution is too variable across runs (stddev=%.3f)", biome.Name, stdDev)
		}
	}
}

// TestHighAltitudeDominance specifically tests if high altitude biomes are too prevalent
func TestHighAltitudeDominance(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  25,
		GridHeight: 25,
	}

	numRuns := 5
	highAltitudeBiomes := []BiomeType{BiomeMountain, BiomeHighAltitude, BiomeCanyon}
	
	for run := 0; run < numRuns; run++ {
		world := NewWorld(config)
		
		totalCells := config.GridWidth * config.GridHeight
		highAltitudeCells := 0
		elevationSum := 0.0
		elevationCount := 0
		
		for y := 0; y < config.GridHeight; y++ {
			for x := 0; x < config.GridWidth; x++ {
				biome := world.Grid[y][x].Biome
				
				// Count high altitude biomes
				for _, highAltBiome := range highAltitudeBiomes {
					if biome == highAltBiome {
						highAltitudeCells++
						break
					}
				}
				
				// Check topology elevation if available
				if world.TopologySystem != nil && x < len(world.TopologySystem.TopologyGrid) && y < len(world.TopologySystem.TopologyGrid[0]) {
					cell := world.TopologySystem.TopologyGrid[x][y]
					elevationSum += cell.Elevation
					elevationCount++
				}
			}
		}
		
		highAltitudePercentage := float64(highAltitudeCells) / float64(totalCells)
		avgElevation := elevationSum / float64(elevationCount)
		
		t.Logf("Run %d: High altitude biomes: %.1f%%, Average elevation: %.3f", 
			run, highAltitudePercentage*100, avgElevation)
		
		// High altitude biomes should not dominate the map
		if highAltitudePercentage > 0.5 {
			t.Errorf("Run %d: High altitude biomes cover %.1f%% of map (> 50%%)", run, highAltitudePercentage*100)
		}
		
		// Average elevation should be reasonable (not too high)
		if avgElevation > 0.6 {
			t.Errorf("Run %d: Average elevation %.3f is too high (> 0.6)", run, avgElevation)
		}
	}
}

// TestBiomeTransitions tests that biomes form reasonable patterns (not too scattered)
func TestBiomeTransitions(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	
	// Count biome transitions (adjacent cells with different biomes)
	transitions := 0
	totalAdjacencies := 0
	
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			currentBiome := world.Grid[y][x].Biome
			
			// Check right neighbor
			if x < config.GridWidth-1 {
				rightBiome := world.Grid[y][x+1].Biome
				if currentBiome != rightBiome {
					transitions++
				}
				totalAdjacencies++
			}
			
			// Check bottom neighbor
			if y < config.GridHeight-1 {
				bottomBiome := world.Grid[y+1][x].Biome
				if currentBiome != bottomBiome {
					transitions++
				}
				totalAdjacencies++
			}
		}
	}
	
	transitionRate := float64(transitions) / float64(totalAdjacencies)
	t.Logf("Biome transition rate: %.3f (transitions/adjacencies)", transitionRate)
	
	// Transition rate should be reasonable:
	// - Too low (< 0.1) means biomes are too clustered
	// - Too high (> 0.8) means biomes are too scattered
	if transitionRate < 0.1 {
		t.Errorf("Biome transition rate %.3f is too low - biomes may be too clustered", transitionRate)
	}
	if transitionRate > 0.8 {
		t.Errorf("Biome transition rate %.3f is too high - biomes may be too scattered", transitionRate)
	}
}

// TestPolarBiomePlacement tests that ice and tundra appear at edges as expected
func TestPolarBiomePlacement(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  20,
		GridHeight: 20,
	}

	world := NewWorld(config)
	
	polarBiomes := []BiomeType{BiomeIce, BiomeTundra}
	edgeDepth := 3 // How many cells from edge to consider "polar region"
	
	edgePolarCount := 0
	centerPolarCount := 0
	totalEdgeCells := 0
	totalCenterCells := 0
	
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			biome := world.Grid[y][x].Biome
			
			isPolar := false
			for _, polarBiome := range polarBiomes {
				if biome == polarBiome {
					isPolar = true
					break
				}
			}
			
			// Check if this cell is on the edge
			isEdge := x < edgeDepth || x >= config.GridWidth-edgeDepth || 
					  y < edgeDepth || y >= config.GridHeight-edgeDepth
			
			if isEdge {
				totalEdgeCells++
				if isPolar {
					edgePolarCount++
				}
			} else {
				totalCenterCells++
				if isPolar {
					centerPolarCount++
				}
			}
		}
	}
	
	edgePolarRate := float64(edgePolarCount) / float64(totalEdgeCells)
	centerPolarRate := float64(centerPolarCount) / float64(totalCenterCells)
	
	t.Logf("Polar biomes at edges: %.1f%% (%d/%d)", edgePolarRate*100, edgePolarCount, totalEdgeCells)
	t.Logf("Polar biomes in center: %.1f%% (%d/%d)", centerPolarRate*100, centerPolarCount, totalCenterCells)
	
	// Polar biomes should be more common at edges than center
	if edgePolarRate <= centerPolarRate {
		t.Errorf("Polar biomes should be more common at edges (%.1f%%) than center (%.1f%%)", 
			edgePolarRate*100, centerPolarRate*100)
	}
}

// TestBiomeConsistencyAcrossMapResets tests if biome generation is stable
func TestBiomeConsistencyAcrossMapResets(t *testing.T) {
	config := WorldConfig{
		Width:      100,
		Height:     100,
		GridWidth:  15,
		GridHeight: 15,
	}

	// Test if the issue mentions map resets - create world, run updates, check for consistency
	world := NewWorld(config)
	
	// Take initial snapshot
	initialBiomes := make([][]BiomeType, config.GridHeight)
	for y := 0; y < config.GridHeight; y++ {
		initialBiomes[y] = make([]BiomeType, config.GridWidth)
		for x := 0; x < config.GridWidth; x++ {
			initialBiomes[y][x] = world.Grid[y][x].Biome
		}
	}
	
	// Run some ticks to see if biomes change unexpectedly
	changesDetected := 0
	for tick := 1; tick <= 20; tick++ {
		// Update topology system (which might affect biomes)
		if world.TopologySystem != nil {
			world.TopologySystem.UpdateTopology(tick)
		}
		
		// Update biomes from topology
		world.updateBiomesFromTopology()
		
		// Check for changes
		for y := 0; y < config.GridHeight; y++ {
			for x := 0; x < config.GridWidth; x++ {
				if world.Grid[y][x].Biome != initialBiomes[y][x] {
					changesDetected++
				}
			}
		}
		
		if tick <= 10 && changesDetected > 0 {
			t.Logf("Tick %d: %d biome changes detected", tick, changesDetected)
		}
	}
	
	totalCells := config.GridWidth * config.GridHeight
	changeRate := float64(changesDetected) / float64(totalCells)
	
	t.Logf("Total biome changes in first 20 ticks: %d/%d (%.1f%%)", 
		changesDetected, totalCells, changeRate*100)
	
	// Rapid biome changes (> 20% of map changing in first 10 ticks) could indicate instability
	if changeRate > 0.2 {
		t.Errorf("Excessive biome changes detected: %.1f%% of map changed in 20 ticks", changeRate*100)
	}
}