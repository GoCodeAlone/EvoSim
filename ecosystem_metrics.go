package main

import (
	"fmt"
	"math"
)

// EcosystemMetrics provides advanced metrics for monitoring ecosystem health and diversity
type EcosystemMetrics struct {
	// Diversity metrics
	ShannonDiversity float64 `json:"shannon_diversity"`
	SimpsonDiversity float64 `json:"simpson_diversity"`
	SpeciesRichness  int     `json:"species_richness"`
	SpeciesEvenness  float64 `json:"species_evenness"`

	// Population metrics
	TotalPopulation     int            `json:"total_population"`
	PopulationBySpecies map[string]int `json:"population_by_species"`
	ExtinctionRate      float64        `json:"extinction_rate"`
	SpeciationRate      float64        `json:"speciation_rate"`

	// Network connectivity
	NetworkConnectivity   float64 `json:"network_connectivity"`
	AveragePathLength     float64 `json:"average_path_length"`
	ClusteringCoefficient float64 `json:"clustering_coefficient"`

	// Pollination metrics
	PollinationSuccess      float64            `json:"pollination_success"`
	CrossSpeciesPollination float64            `json:"cross_species_pollination"`
	PollinatorEfficiency    map[string]float64 `json:"pollinator_efficiency"`

	// Dispersal metrics
	AverageDispersalDistance float64        `json:"average_dispersal_distance"`
	DispersalMethods         map[string]int `json:"dispersal_methods"`
	SeedBankSize             int            `json:"seed_bank_size"`
	GerminationRate          float64        `json:"germination_rate"`

	// Ecosystem health
	EcosystemStability  float64 `json:"ecosystem_stability"`
	BiodiversityIndex   float64 `json:"biodiversity_index"`
	EcosystemResilience float64 `json:"ecosystem_resilience"`
	CarryingCapacity    float64 `json:"carrying_capacity"`
}

// EcosystemMonitor tracks and calculates ecosystem-wide metrics
type EcosystemMonitor struct {
	HistoricalMetrics []EcosystemMetrics `json:"historical_metrics"`
	CurrentMetrics    EcosystemMetrics   `json:"current_metrics"`
	MaxHistorySize    int                `json:"max_history_size"`
}

// NewEcosystemMonitor creates a new ecosystem monitoring system
func NewEcosystemMonitor(maxHistory int) *EcosystemMonitor {
	return &EcosystemMonitor{
		HistoricalMetrics: make([]EcosystemMetrics, 0),
		MaxHistorySize:    maxHistory,
	}
}

// UpdateMetrics calculates all ecosystem metrics for the current state
func (em *EcosystemMonitor) UpdateMetrics(world *World) {
	metrics := EcosystemMetrics{}

	// Calculate diversity metrics
	em.calculateDiversityMetrics(world, &metrics)

	// Calculate population metrics
	em.calculatePopulationMetrics(world, &metrics)

	// Calculate network connectivity
	em.calculateNetworkMetrics(world, &metrics)

	// Calculate pollination metrics
	em.calculatePollinationMetrics(world, &metrics)

	// Calculate dispersal metrics
	em.calculateDispersalMetrics(world, &metrics)

	// Calculate ecosystem health
	em.calculateEcosystemHealth(world, &metrics)

	// Store the metrics
	em.CurrentMetrics = metrics
	em.addToHistory(metrics)
}

// calculateDiversityMetrics computes Shannon and Simpson diversity indices
func (em *EcosystemMonitor) calculateDiversityMetrics(world *World, metrics *EcosystemMetrics) {
	speciesCounts := make(map[string]int)
	totalEntities := 0

	// Count entities by species
	for _, entity := range world.AllEntities {
		speciesName := entity.Species
		if speciesName == "" {
			speciesName = "Unknown"
		}
		speciesCounts[speciesName]++
		totalEntities++
	}

	// Also count plants by species using SpeciationSystem
	plantSpeciesCounts := make(map[string]int)
	totalPlants := 0

	if world.SpeciationSystem != nil {
		for _, species := range world.SpeciationSystem.ActiveSpecies {
			count := len(species.Members)
			if count > 0 {
				plantSpeciesCounts[species.Name] = count
				totalPlants += count
			}
		}
	} else {
		// Fallback: count by plant type if no speciation system
		for _, plant := range world.AllPlants {
			if plant.IsAlive {
				plantType := fmt.Sprintf("Plant_%d", int(plant.Type))
				plantSpeciesCounts[plantType]++
				totalPlants++
			}
		}
	}

	// Combine entity and plant counts
	allSpeciesCounts := make(map[string]int)
	for species, count := range speciesCounts {
		allSpeciesCounts[species] = count
	}
	for species, count := range plantSpeciesCounts {
		allSpeciesCounts[species] = count
	}

	totalOrganisms := totalEntities + totalPlants
	metrics.TotalPopulation = totalOrganisms
	metrics.PopulationBySpecies = allSpeciesCounts
	metrics.SpeciesRichness = len(allSpeciesCounts)

	if totalOrganisms == 0 {
		metrics.ShannonDiversity = 0
		metrics.SimpsonDiversity = 0
		metrics.SpeciesEvenness = 0
		return
	}

	// Calculate Shannon Diversity Index: H = -Σ(pi * ln(pi))
	shannonSum := 0.0
	simpsonSum := 0.0

	for _, count := range allSpeciesCounts {
		if count > 0 {
			proportion := float64(count) / float64(totalOrganisms)
			shannonSum += proportion * math.Log(proportion)
			simpsonSum += proportion * proportion
		}
	}

	metrics.ShannonDiversity = -shannonSum

	// Calculate Simpson Diversity Index: D = 1 - Σ(pi²)
	metrics.SimpsonDiversity = 1.0 - simpsonSum

	// Calculate species evenness: E = H / ln(S)
	if metrics.SpeciesRichness > 1 {
		metrics.SpeciesEvenness = metrics.ShannonDiversity / math.Log(float64(metrics.SpeciesRichness))
	} else {
		metrics.SpeciesEvenness = 1.0
	}
}

// calculatePopulationMetrics computes population-related metrics
func (em *EcosystemMonitor) calculatePopulationMetrics(world *World, metrics *EcosystemMetrics) {
	// Calculate extinction and speciation rates
	if len(em.HistoricalMetrics) > 0 {
		previousSpecies := em.HistoricalMetrics[len(em.HistoricalMetrics)-1].SpeciesRichness
		currentSpecies := metrics.SpeciesRichness

		if previousSpecies > 0 {
			speciesChange := currentSpecies - previousSpecies
			if speciesChange < 0 {
				metrics.ExtinctionRate = float64(-speciesChange) / float64(previousSpecies)
			} else {
				metrics.SpeciationRate = float64(speciesChange) / float64(previousSpecies)
			}
		}
	}
}

// calculateNetworkMetrics computes network connectivity metrics
func (em *EcosystemMonitor) calculateNetworkMetrics(world *World, metrics *EcosystemMetrics) {
	if world.PlantNetworkSystem == nil {
		metrics.NetworkConnectivity = 0
		metrics.AveragePathLength = 0
		metrics.ClusteringCoefficient = 0
		return
	}

	totalConnections := 0
	totalPossibleConnections := 0
	totalPlants := len(world.AllPlants)

	if totalPlants < 2 {
		metrics.NetworkConnectivity = 0
		return
	}

	// Count actual connections
	for _, connection := range world.PlantNetworkSystem.Connections {
		if connection != nil && connection.Health > 0.2 {
			totalConnections++
		}
	}

	// Calculate maximum possible connections (complete graph)
	totalPossibleConnections = totalPlants * (totalPlants - 1) / 2

	if totalPossibleConnections > 0 {
		metrics.NetworkConnectivity = float64(totalConnections) / float64(totalPossibleConnections)
	}

	// Simple clustering coefficient calculation
	// For now, just use connection density as a proxy
	metrics.ClusteringCoefficient = metrics.NetworkConnectivity

	// Average path length estimation (simplified)
	if metrics.NetworkConnectivity > 0 {
		metrics.AveragePathLength = 1.0 / metrics.NetworkConnectivity
	} else {
		metrics.AveragePathLength = float64(totalPlants)
	}
}

// calculatePollinationMetrics computes pollination success metrics
func (em *EcosystemMonitor) calculatePollinationMetrics(world *World, metrics *EcosystemMetrics) {
	if world.InsectPollinationSystem == nil {
		metrics.PollinationSuccess = 0
		metrics.CrossSpeciesPollination = 0
		metrics.PollinatorEfficiency = make(map[string]float64)
		return
	}

	totalAttempts := 0
	successfulPollinations := 0
	crossSpeciesPollinations := 0
	pollinatorEfficiency := make(map[string]float64)
	pollinatorAttempts := make(map[string]int)
	pollinatorSuccesses := make(map[string]int)

	// Analyze recent pollination events from statistics
	for _, event := range world.StatisticalReporter.Events {
		if event.EventType == "pollination_attempt" {
			totalAttempts++
			if event.Metadata != nil {
				if success, ok := event.Metadata["success"].(bool); ok && success {
					successfulPollinations++
				}
				if crossSpecies, ok := event.Metadata["cross_species"].(bool); ok && crossSpecies {
					crossSpeciesPollinations++
				}
				if pollinatorType, ok := event.Metadata["pollinator_type"].(string); ok {
					pollinatorAttempts[pollinatorType]++
					if success, ok := event.Metadata["success"].(bool); ok && success {
						pollinatorSuccesses[pollinatorType]++
					}
				}
			}
		}
	}

	// Calculate success rates
	if totalAttempts > 0 {
		metrics.PollinationSuccess = float64(successfulPollinations) / float64(totalAttempts)
		metrics.CrossSpeciesPollination = float64(crossSpeciesPollinations) / float64(totalAttempts)
	}

	// Calculate pollinator efficiency
	for pollinatorType, attempts := range pollinatorAttempts {
		if attempts > 0 {
			successes := pollinatorSuccesses[pollinatorType]
			pollinatorEfficiency[pollinatorType] = float64(successes) / float64(attempts)
		}
	}
	metrics.PollinatorEfficiency = pollinatorEfficiency
}

// calculateDispersalMetrics computes seed dispersal metrics
func (em *EcosystemMonitor) calculateDispersalMetrics(world *World, metrics *EcosystemMetrics) {
	if world.SeedDispersalSystem == nil {
		metrics.AverageDispersalDistance = 0
		metrics.DispersalMethods = make(map[string]int)
		metrics.SeedBankSize = 0
		metrics.GerminationRate = 0
		return
	}

	totalDistance := 0.0
	dispersalCount := 0
	dispersalMethods := make(map[string]int)
	totalSeeds := 0
	germinatedSeeds := 0

	// Count seeds in seed bank
	for _, bank := range world.SeedDispersalSystem.SeedBanks {
		totalSeeds += len(bank.Seeds)
	}
	metrics.SeedBankSize = totalSeeds

	// Analyze dispersal events from statistics
	for _, event := range world.StatisticalReporter.Events {
		if event.EventType == "seed_dispersal" {
			dispersalCount++
			if event.Metadata != nil {
				if distance, ok := event.Metadata["distance"].(float64); ok {
					totalDistance += distance
				}
				if method, ok := event.Metadata["method"].(string); ok {
					dispersalMethods[method]++
				}
			}
		} else if event.EventType == "seed_germination" {
			germinatedSeeds++
		}
	}

	// Calculate average dispersal distance
	if dispersalCount > 0 {
		metrics.AverageDispersalDistance = totalDistance / float64(dispersalCount)
	}
	metrics.DispersalMethods = dispersalMethods

	// Calculate germination rate
	if totalSeeds > 0 {
		metrics.GerminationRate = float64(germinatedSeeds) / float64(totalSeeds)
	}
}

// calculateEcosystemHealth computes overall ecosystem health metrics
func (em *EcosystemMonitor) calculateEcosystemHealth(world *World, metrics *EcosystemMetrics) {
	// Biodiversity index combines Shannon diversity and species richness
	if metrics.SpeciesRichness > 0 {
		metrics.BiodiversityIndex = metrics.ShannonDiversity * math.Log(float64(metrics.SpeciesRichness))
	}

	// Ecosystem stability based on population variance
	if len(em.HistoricalMetrics) >= 3 {
		populations := make([]float64, 0)
		for _, hist := range em.HistoricalMetrics[len(em.HistoricalMetrics)-3:] {
			populations = append(populations, float64(hist.TotalPopulation))
		}
		mean, stdDev := em.calculateMeanAndStdDev(populations)
		if mean > 0 {
			metrics.EcosystemStability = 1.0 - (stdDev / mean) // Lower coefficient of variation = higher stability
		}
	} else {
		metrics.EcosystemStability = 1.0 // Assume stable if not enough history
	}

	// Carrying capacity estimation (simple approach)
	maxPopulation := metrics.TotalPopulation
	for _, hist := range em.HistoricalMetrics {
		if hist.TotalPopulation > maxPopulation {
			maxPopulation = hist.TotalPopulation
		}
	}
	metrics.CarryingCapacity = float64(maxPopulation) * 1.1 // Add 10% buffer

	// Ecosystem resilience based on diversity and stability
	metrics.EcosystemResilience = (metrics.ShannonDiversity + metrics.EcosystemStability) / 2.0
}

// calculateMeanAndStdDev calculates mean and standard deviation of a slice of float64
func (em *EcosystemMonitor) calculateMeanAndStdDev(values []float64) (float64, float64) {
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
	sumSquaredDiffs := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}
	variance := sumSquaredDiffs / float64(len(values))
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}

// addToHistory adds metrics to historical record
func (em *EcosystemMonitor) addToHistory(metrics EcosystemMetrics) {
	em.HistoricalMetrics = append(em.HistoricalMetrics, metrics)

	// Maintain maximum history size
	if len(em.HistoricalMetrics) > em.MaxHistorySize {
		em.HistoricalMetrics = em.HistoricalMetrics[1:]
	}
}

// GetTrends returns trend analysis for key metrics
func (em *EcosystemMonitor) GetTrends() map[string]string {
	if len(em.HistoricalMetrics) < 3 {
		return map[string]string{
			"diversity":  "insufficient_data",
			"population": "insufficient_data",
			"stability":  "insufficient_data",
		}
	}

	trends := make(map[string]string)
	recent := em.HistoricalMetrics[len(em.HistoricalMetrics)-3:]

	// Diversity trend
	if recent[2].ShannonDiversity > recent[1].ShannonDiversity && recent[1].ShannonDiversity > recent[0].ShannonDiversity {
		trends["diversity"] = "increasing"
	} else if recent[2].ShannonDiversity < recent[1].ShannonDiversity && recent[1].ShannonDiversity < recent[0].ShannonDiversity {
		trends["diversity"] = "decreasing"
	} else {
		trends["diversity"] = "stable"
	}

	// Population trend
	if recent[2].TotalPopulation > recent[1].TotalPopulation && recent[1].TotalPopulation > recent[0].TotalPopulation {
		trends["population"] = "growing"
	} else if recent[2].TotalPopulation < recent[1].TotalPopulation && recent[1].TotalPopulation < recent[0].TotalPopulation {
		trends["population"] = "declining"
	} else {
		trends["population"] = "stable"
	}

	// Stability trend
	if recent[2].EcosystemStability > recent[1].EcosystemStability && recent[1].EcosystemStability > recent[0].EcosystemStability {
		trends["stability"] = "improving"
	} else if recent[2].EcosystemStability < recent[1].EcosystemStability && recent[1].EcosystemStability < recent[0].EcosystemStability {
		trends["stability"] = "degrading"
	} else {
		trends["stability"] = "stable"
	}

	return trends
}

// GetHealthScore returns an overall ecosystem health score (0-100)
func (em *EcosystemMonitor) GetHealthScore() float64 {
	current := em.CurrentMetrics

	// Weighted combination of key metrics
	diversityScore := math.Min(current.ShannonDiversity*20, 30) // Max 30 points
	stabilityScore := current.EcosystemStability * 25           // Max 25 points
	connectivityScore := current.NetworkConnectivity * 20       // Max 20 points
	pollinationScore := current.PollinationSuccess * 15         // Max 15 points
	resilienceScore := current.EcosystemResilience * 10         // Max 10 points

	totalScore := diversityScore + stabilityScore + connectivityScore + pollinationScore + resilienceScore

	return math.Min(totalScore, 100) // Cap at 100
}
