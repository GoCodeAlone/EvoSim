package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestAllViewsDataValidation tests that each view receives proper data from the view manager
func TestAllViewsDataValidation(t *testing.T) {
	// Create a test world
	world := createWebTestWorld(t)
	
	// Create view manager
	vm := NewViewManager(world)
	
	// Run a few simulation ticks to generate data
	for i := 0; i < 10; i++ {
		world.Update()
		time.Sleep(10 * time.Millisecond)
	}
	
	// Get view data
	viewData := vm.GetCurrentViewData()
	
	// Test each view has proper data structure
	testCases := []struct{
		name string
		validator func(*ViewData) []string
	}{
		{"GRID", validateGridView},
		{"STATS", validateStatsView},
		{"EVENTS", validateEventsView},
		{"POPULATIONS", validatePopulationsView},
		{"COMMUNICATION", validateCommunicationView},
		{"CIVILIZATION", validateCivilizationView},
		{"PHYSICS", validatePhysicsView},
		{"WIND", validateWindView},
		{"SPECIES", validateSpeciesView},
		{"NETWORK", validateNetworkView},
		{"DNA", validateDNAView},
		{"CELLULAR", validateCellularView},
		{"EVOLUTION", validateEvolutionView},
		{"TOPOLOGY", validateTopologyView},
		{"TOOLS", validateToolsView},
		{"ENVIRONMENT", validateEnvironmentView},
		{"BEHAVIOR", validateBehaviorView},
		{"REPRODUCTION", validateReproductionView},
		{"STATISTICAL", validateStatisticalView},
		{"ANOMALIES", validateAnomaliesView},
		{"ECOSYSTEM", validateEcosystemView},
		{"WARFARE", validateWarfareView},
		{"FUNGAL", validateFungalView},
		{"CULTURAL", validateCulturalView},
		{"SYMBIOTIC", validateSymbioticView},
		{"NEURAL", validateNeuralView},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := tc.validator(viewData)
			if len(errors) > 0 {
				t.Errorf("View %s has data validation errors:\n%v", tc.name, errors)
				// Print the actual data for debugging
				switch tc.name {
				case "SPECIES":
					data, _ := json.MarshalIndent(viewData.Species, "", "  ")
					t.Logf("SPECIES data: %s", string(data))
				}
			}
		})
	}
}

func createWebTestWorld(t *testing.T) *World {
	config := WorldConfig{
		Width:        50.0,
		Height:       50.0,
		GridWidth:    20,
		GridHeight:   20,
	}
	
	world := NewWorld(config)
	
	// Add test populations
	testPops := []PopulationConfig{
		{
			Name:    "TestHerbivores",
			Species: "herbivore",
			BaseTraits: map[string]float64{
				"size": 0.5,
				"speed": 0.3,
				"aggression": -0.8,
			},
			StartPos: Position{X: 10, Y: 10},
			Spread: 5.0,
		},
		{
			Name:    "TestPredators", 
			Species: "predator",
			BaseTraits: map[string]float64{
				"size": 0.8,
				"speed": 0.6,
				"aggression": 0.8,
			},
			StartPos: Position{X: 30, Y: 30},
			Spread: 5.0,
		},
	}
	
	for _, pop := range testPops {
		world.AddPopulation(pop)
	}
	
	return world
}

// Validation functions for each view
func validateGridView(data *ViewData) []string {
	var errors []string
	
	if data.Grid == nil {
		errors = append(errors, "Grid is nil")
		return errors
	}
	
	if len(data.Grid) == 0 {
		errors = append(errors, "Grid is empty")
		return errors
	}
	
	// Check grid structure
	for y, row := range data.Grid {
		if len(row) == 0 {
			errors = append(errors, fmt.Sprintf("Grid row %d is empty", y))
		}
		for x, cell := range row {
			if cell.Biome == "" {
				errors = append(errors, fmt.Sprintf("Cell [%d,%d] has empty biome", y, x))
			}
			if cell.BiomeSymbol == "" {
				errors = append(errors, fmt.Sprintf("Cell [%d,%d] has empty biome symbol", y, x))
			}
		}
	}
	
	return errors
}

func validateStatsView(data *ViewData) []string {
	var errors []string
	
	if data.Stats == nil {
		errors = append(errors, "Stats is nil")
		return errors
	}
	
	// Check required stats fields
	requiredFields := []string{"avg_fitness", "avg_energy", "avg_age"}
	for _, field := range requiredFields {
		if _, exists := data.Stats[field]; !exists {
			errors = append(errors, fmt.Sprintf("Missing required stats field: %s", field))
		}
	}
	
	return errors
}

func validateEventsView(data *ViewData) []string {
	var errors []string
	
	if data.Events == nil {
		errors = append(errors, "Events is nil")
		return errors
	}
	
	// Events can be empty, but structure should be valid
	for i, event := range data.Events {
		if event.Name == "" {
			errors = append(errors, fmt.Sprintf("Event %d has empty name", i))
		}
		if event.Type == "" {
			errors = append(errors, fmt.Sprintf("Event %d has empty type", i))
		}
	}
	
	return errors
}

func validatePopulationsView(data *ViewData) []string {
	var errors []string
	
	if data.Populations == nil {
		errors = append(errors, "Populations is nil")
		return errors
	}
	
	for i, pop := range data.Populations {
		if pop.Name == "" {
			errors = append(errors, fmt.Sprintf("Population %d has empty name", i))
		}
		if pop.Species == "" {
			errors = append(errors, fmt.Sprintf("Population %d has empty species", i))
		}
		if pop.TraitAverages == nil {
			errors = append(errors, fmt.Sprintf("Population %d has nil trait averages", i))
		}
	}
	
	return errors
}

func validateCommunicationView(data *ViewData) []string {
	var errors []string
	
	if data.Communication.SignalTypes == nil {
		errors = append(errors, "Communication SignalTypes is nil")
	}
	
	return errors
}

func validateCivilizationView(data *ViewData) []string {
	var errors []string
	
	// Basic structure check - civilization data might be empty but should be valid
	if data.Civilization.TribesCount < 0 {
		errors = append(errors, "Negative tribes count")
	}
	if data.Civilization.StructureCount < 0 {
		errors = append(errors, "Negative structure count")
	}
	
	return errors
}

func validatePhysicsView(data *ViewData) []string {
	var errors []string
	
	if data.Physics.CollisionsLastTick < 0 {
		errors = append(errors, "Negative collisions count")
	}
	if data.Physics.AverageVelocity < 0 {
		errors = append(errors, "Negative average velocity")
	}
	
	return errors
}

func validateWindView(data *ViewData) []string {
	var errors []string
	
	if data.Wind.WeatherPattern == "" {
		errors = append(errors, "Wind weather pattern is empty")
	}
	if data.Wind.PollenCount < 0 {
		errors = append(errors, "Negative pollen count")
	}
	
	return errors
}

func validateSpeciesView(data *ViewData) []string {
	var errors []string
	
	if data.Species.ActiveSpecies < 0 {
		errors = append(errors, "Negative active species count")
	}
	if data.Species.ExtinctSpecies < 0 {
		errors = append(errors, "Negative extinct species count")
	}
	if data.Species.TotalSpeciesEver < 0 {
		errors = append(errors, "Negative total species ever count")
	}
	if data.Species.SpeciesDetails == nil {
		errors = append(errors, "Species details is nil")
	}
	
	// Check species details structure
	for i, detail := range data.Species.SpeciesDetails {
		if detail.Name == "" {
			errors = append(errors, fmt.Sprintf("Species detail %d has empty name", i))
		}
		if detail.Population < 0 {
			errors = append(errors, fmt.Sprintf("Species detail %d has negative population", i))
		}
		if detail.PeakPopulation < 0 {
			errors = append(errors, fmt.Sprintf("Species detail %d has negative peak population", i))
		}
	}
	
	return errors
}

func validateNetworkView(data *ViewData) []string {
	var errors []string
	
	if data.Network.ConnectionCount < 0 {
		errors = append(errors, "Negative connection count")
	}
	if data.Network.SignalCount < 0 {
		errors = append(errors, "Negative signal count")
	}
	if data.Network.ClusterCount < 0 {
		errors = append(errors, "Negative cluster count")
	}
	
	return errors
}

func validateDNAView(data *ViewData) []string {
	var errors []string
	
	if data.DNA.OrganismCount < 0 {
		errors = append(errors, "Negative organism count")
	}
	if data.DNA.AverageMutations < 0 {
		errors = append(errors, "Negative average mutations")
	}
	if data.DNA.AverageComplexity < 0 {
		errors = append(errors, "Negative average complexity")
	}
	
	return errors
}

func validateCellularView(data *ViewData) []string {
	var errors []string
	
	if data.Cellular.TotalCells < 0 {
		errors = append(errors, "Negative total cells")
	}
	if data.Cellular.CellDivisions < 0 {
		errors = append(errors, "Negative cell divisions")
	}
	if data.Cellular.AverageComplexity < 0 {
		errors = append(errors, "Negative average complexity")
	}
	
	return errors
}

func validateEvolutionView(data *ViewData) []string {
	var errors []string
	
	if data.Evolution.SpeciationEvents < 0 {
		errors = append(errors, "Negative speciation events")
	}
	if data.Evolution.ExtinctionEvents < 0 {
		errors = append(errors, "Negative extinction events")
	}
	if data.Evolution.GeneticDiversity < 0 {
		errors = append(errors, "Negative genetic diversity")
	}
	if data.Evolution.ActivePlantCount < 0 {
		errors = append(errors, "Negative active plant count")
	}
	
	return errors
}

func validateTopologyView(data *ViewData) []string {
	var errors []string
	
	if data.Topology.FluidRegions < 0 {
		errors = append(errors, "Negative fluid regions")
	}
	if data.Topology.GeologicalAge < 0 {
		errors = append(errors, "Negative geological age")
	}
	if data.Topology.ElevationRange == "" {
		errors = append(errors, "Empty elevation range")
	}
	
	return errors
}

func validateToolsView(data *ViewData) []string {
	var errors []string
	
	if data.Tools.TotalTools < 0 {
		errors = append(errors, "Negative total tools")
	}
	if data.Tools.OwnedTools < 0 {
		errors = append(errors, "Negative owned tools")
	}
	if data.Tools.DroppedTools < 0 {
		errors = append(errors, "Negative dropped tools")
	}
	if data.Tools.ToolTypes == nil {
		errors = append(errors, "Tool types is nil")
	}
	
	return errors
}

func validateEnvironmentView(data *ViewData) []string {
	var errors []string
	
	if data.EnvironmentalMod.TotalModifications < 0 {
		errors = append(errors, "Negative total modifications")
	}
	if data.EnvironmentalMod.ActiveModifications < 0 {
		errors = append(errors, "Negative active modifications")
	}
	if data.EnvironmentalMod.InactiveModifications < 0 {
		errors = append(errors, "Negative inactive modifications")
	}
	if data.EnvironmentalMod.ModificationTypes == nil {
		errors = append(errors, "Modification types is nil")
	}
	
	return errors
}

func validateBehaviorView(data *ViewData) []string {
	var errors []string
	
	if data.EmergentBehavior.TotalEntities < 0 {
		errors = append(errors, "Negative total entities")
	}
	if data.EmergentBehavior.DiscoveredBehaviors < 0 {
		errors = append(errors, "Negative discovered behaviors")
	}
	if data.EmergentBehavior.BehaviorSpread == nil {
		errors = append(errors, "Behavior spread is nil")
	}
	if data.EmergentBehavior.AvgProficiency == nil {
		errors = append(errors, "Average proficiency is nil")
	}
	
	return errors
}

func validateReproductionView(data *ViewData) []string {
	var errors []string
	
	if data.Reproduction.ActiveEggs < 0 {
		errors = append(errors, "Negative active eggs")
	}
	if data.Reproduction.DecayingItems < 0 {
		errors = append(errors, "Negative decaying items")
	}
	if data.Reproduction.PregnantEntities < 0 {
		errors = append(errors, "Negative pregnant entities")
	}
	if data.Reproduction.ReadyToMate < 0 {
		errors = append(errors, "Negative ready to mate")
	}
	if data.Reproduction.ReproductionModes == nil {
		errors = append(errors, "Reproduction modes is nil")
	}
	if data.Reproduction.MatingStrategies == nil {
		errors = append(errors, "Mating strategies is nil")
	}
	
	return errors
}

func validateStatisticalView(data *ViewData) []string {
	var errors []string
	
	if data.Statistical.TotalEvents < 0 {
		errors = append(errors, "Negative total events")
	}
	if data.Statistical.TotalSnapshots < 0 {
		errors = append(errors, "Negative total snapshots")
	}
	if data.Statistical.TotalAnomalies < 0 {
		errors = append(errors, "Negative total anomalies")
	}
	if data.Statistical.RecentEvents == nil {
		errors = append(errors, "Recent events is nil")
	}
	
	return errors
}

func validateAnomaliesView(data *ViewData) []string {
	var errors []string
	
	if data.Anomalies.TotalAnomalies < 0 {
		errors = append(errors, "Negative total anomalies")
	}
	if data.Anomalies.RecentAnomalies == nil {
		errors = append(errors, "Recent anomalies is nil")
	}
	if data.Anomalies.AnomalyTypes == nil {
		errors = append(errors, "Anomaly types is nil")
	}
	if data.Anomalies.Recommendations == nil {
		errors = append(errors, "Recommendations is nil")
	}
	
	return errors
}

func validateEcosystemView(data *ViewData) []string {
	var errors []string
	
	if data.Ecosystem.BiodiversityIndex < 0 {
		errors = append(errors, "Negative biodiversity index")
	}
	if data.Ecosystem.TotalPopulation < 0 {
		errors = append(errors, "Negative total population")
	}
	if data.Ecosystem.SpeciesRichness < 0 {
		errors = append(errors, "Negative species richness")
	}
	if data.Ecosystem.ShannonDiversity < 0 {
		errors = append(errors, "Negative Shannon diversity")
	}
	
	return errors
}

func validateWarfareView(data *ViewData) []string {
	var errors []string
	
	if data.Warfare.TotalColonies < 0 {
		errors = append(errors, "Negative total colonies")
	}
	if data.Warfare.ActiveConflicts < 0 {
		errors = append(errors, "Negative active conflicts")
	}
	if data.Warfare.TotalAlliances < 0 {
		errors = append(errors, "Negative total alliances")
	}
	if data.Warfare.ActiveTradeAgreements < 0 {
		errors = append(errors, "Negative active trade agreements")
	}
	
	return errors
}

func validateFungalView(data *ViewData) []string {
	var errors []string
	
	if data.Fungal.TotalOrganisms < 0 {
		errors = append(errors, "Negative total organisms")
	}
	if data.Fungal.ActiveSpores < 0 {
		errors = append(errors, "Negative active spores")
	}
	if data.Fungal.NetworkConnections < 0 {
		errors = append(errors, "Negative network connections")
	}
	if data.Fungal.TotalBiomass < 0 {
		errors = append(errors, "Negative total biomass")
	}
	
	return errors
}

func validateCulturalView(data *ViewData) []string {
	var errors []string
	
	if data.Cultural.TotalKnowledgeTypes < 0 {
		errors = append(errors, "Negative total knowledge types")
	}
	if data.Cultural.TotalEntities < 0 {
		errors = append(errors, "Negative total entities")
	}
	if data.Cultural.ActiveInnovations < 0 {
		errors = append(errors, "Negative active innovations")
	}
	if data.Cultural.TotalTeachingEvents < 0 {
		errors = append(errors, "Negative total teaching events")
	}
	
	return errors
}

func validateSymbioticView(data *ViewData) []string {
	var errors []string
	
	if data.SymbioticRelationships.TotalRelationships < 0 {
		errors = append(errors, "Negative total relationships")
	}
	if data.SymbioticRelationships.ActiveRelationships < 0 {
		errors = append(errors, "Negative active relationships")
	}
	if data.SymbioticRelationships.AverageRelationshipAge < 0 {
		errors = append(errors, "Negative average relationship age")
	}
	if data.SymbioticRelationships.DiseaseTransmissionRate < 0 {
		errors = append(errors, "Negative disease transmission rate")
	}
	
	return errors
}

func validateNeuralView(data *ViewData) []string {
	var errors []string
	
	if data.Neural.TotalNetworks < 0 {
		errors = append(errors, "Negative total networks")
	}
	if data.Neural.TotalBehaviors < 0 {
		errors = append(errors, "Negative total behaviors")
	}
	if data.Neural.TotalLearningEvents < 0 {
		errors = append(errors, "Negative total learning events")
	}
	if data.Neural.EmergentBehaviors < 0 {
		errors = append(errors, "Negative emergent behaviors")
	}
	if data.Neural.AvgNetworkComplexity < 0 {
		errors = append(errors, "Negative average network complexity")
	}
	if data.Neural.SuccessRate < 0 || data.Neural.SuccessRate > 1 {
		errors = append(errors, "Invalid success rate (should be 0-1)")
	}
	if data.Neural.TotalExperience < 0 {
		errors = append(errors, "Negative total experience")
	}
	if data.Neural.AvgExperiencePerNetwork < 0 {
		errors = append(errors, "Negative average experience per network")
	}
	if data.Neural.BaseLearningRate < 0 {
		errors = append(errors, "Negative base learning rate")
	}
	if data.Neural.AdaptationRate < 0 {
		errors = append(errors, "Negative adaptation rate")
	}
	if data.Neural.ActiveNetworkCount < 0 {
		errors = append(errors, "Negative active network count")
	}
	if data.Neural.CollectiveBehaviorCount < 0 {
		errors = append(errors, "Negative collective behavior count")
	}
	if data.Neural.SuccessfulStrategies == nil {
		errors = append(errors, "Successful strategies is nil")
	}
	if data.Neural.EntityNetworks == nil {
		errors = append(errors, "Entity networks is nil")
	}
	
	return errors
}