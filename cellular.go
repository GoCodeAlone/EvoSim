package main

import (
	"fmt"
	"math"
	"math/rand"
)

// CellType represents different types of cells
type CellType int

const (
	CellTypeStem CellType = iota
	CellTypeNerve
	CellTypeMuscle
	CellTypeDigestive
	CellTypeReproductive
	CellTypeDefensive
	CellTypePhotosynthetic
	CellTypeStorage
)

// OrganelleType represents cellular organelles
type OrganelleType int

const (
	OrganelleNucleus OrganelleType = iota
	OrganelleMitochondria
	OrganelleChloroplast
	OrganelleRibosome
	OrganelleVacuole
	OrganelleGolgi
	OrganelleER // Endoplasmic Reticulum
	OrganelleLysosome
)

// Organelle represents cellular components
type Organelle struct {
	Type       OrganelleType `json:"type"`
	Count      int           `json:"count"`      // Number of this organelle
	Efficiency float64       `json:"efficiency"` // How well it functions (0-1)
	Energy     float64       `json:"energy"`     // Energy stored/produced
}

// Cell represents a single cell with organelles and functions
type Cell struct {
	ID          int                          `json:"id"`
	Type        CellType                     `json:"type"`
	Size        float64                      `json:"size"`        // Cell diameter
	Energy      float64                      `json:"energy"`      // Current energy
	Health      float64                      `json:"health"`      // Cell health (0-1)
	Age         int                          `json:"age"`         // Cell age in ticks
	DNA         *DNAStrand                   `json:"dna"`         // Genetic material
	Organelles  map[OrganelleType]*Organelle `json:"organelles"`  // Cellular components
	Position    Position                     `json:"position"`    // Position within organism
	Connections []int                        `json:"connections"` // Connected cell IDs
	Activity    float64                      `json:"activity"`    // Current activity level (0-1)
	Specialized bool                         `json:"specialized"` // Whether cell is specialized
}

// CellularOrganism represents a single-cell or multi-cell entity
type CellularOrganism struct {
	EntityID        int              `json:"entity_id"`
	Cells           []*Cell          `json:"cells"`
	ComplexityLevel int              `json:"complexity_level"` // 1=single-cell, 2+=multi-cell
	TotalEnergy     float64          `json:"total_energy"`
	CellDivisions   int              `json:"cell_divisions"` // Total divisions performed
	Generation      int              `json:"generation"`
	OrganSystems    map[string][]int `json:"organ_systems"` // System name -> cell IDs
}

// CellularSystem manages cellular-level evolution and processes
type CellularSystem struct {
	NextCellID           int                       `json:"next_cell_id"`
	OrganismMap          map[int]*CellularOrganism `json:"organism_map"` // EntityID -> Organism
	CellTypeNames        map[CellType]string       `json:"cell_type_names"`
	OrganelleNames       map[OrganelleType]string  `json:"organelle_names"`
	ComplexityThresholds map[int]int               `json:"complexity_thresholds"` // Level -> min cells
	DNASystem            *DNASystem                `json:"-"`
	eventBus             *CentralEventBus          `json:"-"` // Event tracking
}

// NewCellularSystem creates a new cellular management system
func NewCellularSystem(dnaSystem *DNASystem, eventBus *CentralEventBus) *CellularSystem {
	return &CellularSystem{
		NextCellID:  1,
		OrganismMap: make(map[int]*CellularOrganism),
		CellTypeNames: map[CellType]string{
			CellTypeStem:           "Stem",
			CellTypeNerve:          "Nerve",
			CellTypeMuscle:         "Muscle",
			CellTypeDigestive:      "Digestive",
			CellTypeReproductive:   "Reproductive",
			CellTypeDefensive:      "Defensive",
			CellTypePhotosynthetic: "Photosynthetic",
			CellTypeStorage:        "Storage",
		},
		OrganelleNames: map[OrganelleType]string{
			OrganelleNucleus:      "Nucleus",
			OrganelleMitochondria: "Mitochondria",
			OrganelleChloroplast:  "Chloroplast",
			OrganelleRibosome:     "Ribosome",
			OrganelleVacuole:      "Vacuole",
			OrganelleGolgi:        "Golgi Apparatus",
			OrganelleER:           "Endoplasmic Reticulum",
			OrganelleLysosome:     "Lysosome",
		},
		ComplexityThresholds: map[int]int{
			1: 1,   // Single cell
			2: 5,   // Simple multicellular
			3: 20,  // Complex multicellular
			4: 100, // Advanced multicellular
			5: 500, // Highly complex
		},
		DNASystem: dnaSystem,
		eventBus:  eventBus,
	}
}

// CreateSingleCellOrganism creates a new single-cell organism
func (cs *CellularSystem) CreateSingleCellOrganism(entityID int, dna *DNAStrand) *CellularOrganism {
	// Create the primary cell
	cell := cs.createCell(CellTypeStem, dna, Position{X: 0, Y: 0})

	organism := &CellularOrganism{
		EntityID:        entityID,
		Cells:           []*Cell{cell},
		ComplexityLevel: 1,
		TotalEnergy:     cell.Energy,
		CellDivisions:   0,
		Generation:      dna.Generation,
		OrganSystems:    make(map[string][]int),
	}

	// Initialize basic organ systems
	organism.OrganSystems["core"] = []int{cell.ID}

	cs.OrganismMap[entityID] = organism
	return organism
}

// createCell creates a new cell with organelles based on DNA
func (cs *CellularSystem) createCell(cellType CellType, dna *DNAStrand, position Position) *Cell {
	cellID := cs.NextCellID
	cs.NextCellID++

	cell := &Cell{
		ID:          cellID,
		Type:        cellType,
		Size:        cs.calculateCellSize(dna),
		Energy:      100.0,
		Health:      1.0,
		Age:         0,
		DNA:         dna,
		Organelles:  make(map[OrganelleType]*Organelle),
		Position:    position,
		Connections: make([]int, 0),
		Activity:    0.5,
		Specialized: cellType != CellTypeStem,
	}

	// Add organelles based on cell type and DNA
	cs.addOrganellesToCell(cell, dna)

	return cell
}

// addOrganellesToCell adds appropriate organelles to a cell
func (cs *CellularSystem) addOrganellesToCell(cell *Cell, dna *DNAStrand) {
	// All cells have nucleus and ribosomes
	cell.Organelles[OrganelleNucleus] = &Organelle{
		Type:       OrganelleNucleus,
		Count:      1,
		Efficiency: 0.8 + cs.DNASystem.ExpressTrait(dna, "intelligence")*0.2,
		Energy:     10.0,
	}

	ribosomeCount := int(5 + cs.DNASystem.ExpressTrait(dna, "metabolism")*5)
	cell.Organelles[OrganelleRibosome] = &Organelle{
		Type:       OrganelleRibosome,
		Count:      ribosomeCount,
		Efficiency: 0.7 + cs.DNASystem.ExpressTrait(dna, "energy")*0.3,
		Energy:     5.0,
	}

	// Mitochondria for energy production
	mitoCount := int(3 + cs.DNASystem.ExpressTrait(dna, "energy")*7)
	cell.Organelles[OrganelleMitochondria] = &Organelle{
		Type:       OrganelleMitochondria,
		Count:      mitoCount,
		Efficiency: 0.6 + cs.DNASystem.ExpressTrait(dna, "metabolism")*0.4,
		Energy:     20.0,
	}

	// Cell type specific organelles
	switch cell.Type {
	case CellTypePhotosynthetic:
		chloroCount := int(2 + cs.DNASystem.ExpressTrait(dna, "energy")*3)
		cell.Organelles[OrganelleChloroplast] = &Organelle{
			Type:       OrganelleChloroplast,
			Count:      chloroCount,
			Efficiency: 0.5 + cs.DNASystem.ExpressTrait(dna, "adaptation")*0.5,
			Energy:     15.0,
		}

	case CellTypeStorage:
		cell.Organelles[OrganelleVacuole] = &Organelle{
			Type:       OrganelleVacuole,
			Count:      1,
			Efficiency: 0.8 + cs.DNASystem.ExpressTrait(dna, "size")*0.2,
			Energy:     30.0,
		}

	case CellTypeDigestive:
		lysoCount := int(2 + cs.DNASystem.ExpressTrait(dna, "metabolism")*3)
		cell.Organelles[OrganelleLysosome] = &Organelle{
			Type:       OrganelleLysosome,
			Count:      lysoCount,
			Efficiency: 0.7 + cs.DNASystem.ExpressTrait(dna, "strength")*0.3,
			Energy:     8.0,
		}

	case CellTypeNerve:
		// Enhanced ER for neurotransmitter production
		cell.Organelles[OrganelleER] = &Organelle{
			Type:       OrganelleER,
			Count:      2,
			Efficiency: 0.6 + cs.DNASystem.ExpressTrait(dna, "intelligence")*0.4,
			Energy:     12.0,
		}

	case CellTypeMuscle:
		// Enhanced mitochondria for muscle cells
		if mito, exists := cell.Organelles[OrganelleMitochondria]; exists {
			mito.Count *= 2
			mito.Efficiency += 0.1
		}
	}

	// All cells get Golgi apparatus
	cell.Organelles[OrganelleGolgi] = &Organelle{
		Type:       OrganelleGolgi,
		Count:      1,
		Efficiency: 0.7,
		Energy:     7.0,
	}
}

// calculateCellSize determines cell size based on DNA
func (cs *CellularSystem) calculateCellSize(dna *DNAStrand) float64 {
	baseSizeFromDNA := cs.DNASystem.ExpressTrait(dna, "size")

	// Convert to actual size (micrometers)
	size := 5.0 + baseSizeFromDNA*15.0 // 5-20 micrometers
	return math.Max(1.0, size)
}

// UpdateCellularOrganisms updates all cellular organisms
func (cs *CellularSystem) UpdateCellularOrganisms() {
	for _, organism := range cs.OrganismMap {
		cs.updateOrganism(organism)
	}
}

// updateOrganism updates a single cellular organism
func (cs *CellularSystem) updateOrganism(organism *CellularOrganism) {
	totalEnergy := 0.0

	// Update each cell
	for _, cell := range organism.Cells {
		cs.updateCell(cell)
		totalEnergy += cell.Energy

		// Check for cell division
		if cs.shouldCellDivide(cell, organism) {
			cs.performCellDivision(cell, organism)
		}

		// Check for cell death
		if cell.Health <= 0 || cell.Energy <= 0 {
			cs.handleCellDeath(cell, organism)
		}
	}

	organism.TotalEnergy = totalEnergy

	// Update complexity level
	organism.ComplexityLevel = cs.calculateComplexityLevel(len(organism.Cells))

	// Update organ systems
	cs.updateOrganSystems(organism)
}

// updateCell updates a single cell's state
func (cs *CellularSystem) updateCell(cell *Cell) {
	cell.Age++

	// Calculate energy production and consumption
	energyProduced := 0.0
	energyConsumed := 0.0

	for _, organelle := range cell.Organelles {
		switch organelle.Type {
		case OrganelleMitochondria:
			energyProduced += float64(organelle.Count) * organelle.Efficiency * 2.0
		case OrganelleChloroplast:
			energyProduced += float64(organelle.Count) * organelle.Efficiency * 1.5
		case OrganelleRibosome:
			energyConsumed += float64(organelle.Count) * 0.1
		case OrganelleNucleus:
			energyConsumed += 0.5
		default:
			energyConsumed += float64(organelle.Count) * 0.05
		}
	}

	// Apply energy changes
	netEnergy := energyProduced - energyConsumed
	cell.Energy += netEnergy

	// Apply aging effects
	agingFactor := 1.0 - (float64(cell.Age) * 0.0001)
	cell.Health *= agingFactor

	// Update activity based on energy and health
	cell.Activity = math.Min(1.0, (cell.Energy/100.0)*cell.Health)

	// Clamp values
	cell.Energy = math.Max(0, cell.Energy)
	cell.Health = math.Max(0, math.Min(1.0, cell.Health))
}

// shouldCellDivide determines if a cell should divide
func (cs *CellularSystem) shouldCellDivide(cell *Cell, organism *CellularOrganism) bool {
	// Conditions for cell division
	energyThreshold := 150.0
	healthThreshold := 0.8
	ageMinimum := 50
	maxCells := 1000 // Prevent unlimited growth

	return cell.Energy > energyThreshold &&
		cell.Health > healthThreshold &&
		cell.Age > ageMinimum &&
		len(organism.Cells) < maxCells &&
		rand.Float64() < 0.01 // 1% chance per tick
}

// performCellDivision creates a new cell through division
func (cs *CellularSystem) performCellDivision(parentCell *Cell, organism *CellularOrganism) {
	// Create daughter cell
	daughterCell := cs.createCell(parentCell.Type, parentCell.DNA, parentCell.Position)

	originalParentEnergy := parentCell.Energy

	// Split energy between parent and daughter
	parentCell.Energy *= 0.6
	daughterCell.Energy = parentCell.Energy * 0.4

	// Slight positioning offset
	daughterCell.Position.X += (rand.Float64() - 0.5) * 2.0
	daughterCell.Position.Y += (rand.Float64() - 0.5) * 2.0

	// Connect cells
	parentCell.Connections = append(parentCell.Connections, daughterCell.ID)
	daughterCell.Connections = append(daughterCell.Connections, parentCell.ID)

	// Add to organism
	organism.Cells = append(organism.Cells, daughterCell)
	organism.CellDivisions++

	// Emit cell division event
	if cs.eventBus != nil {
		metadata := map[string]interface{}{
			"entity_id":           organism.EntityID,
			"parent_cell_id":      parentCell.ID,
			"daughter_cell_id":    daughterCell.ID,
			"cell_type":           cs.CellTypeNames[parentCell.Type],
			"original_energy":     originalParentEnergy,
			"parent_energy":       parentCell.Energy,
			"daughter_energy":     daughterCell.Energy,
			"total_divisions":     organism.CellDivisions,
			"organism_complexity": organism.ComplexityLevel,
			"total_cells":         len(organism.Cells),
			"generation":          organism.Generation,
		}

		cs.eventBus.EmitSystemEvent(
			-1,
			"cell_division",
			"cellular",
			"cellular_system",
			fmt.Sprintf("Cell division in entity %d: %s cell %d created daughter cell %d (%d total cells)",
				organism.EntityID, cs.CellTypeNames[parentCell.Type], parentCell.ID, daughterCell.ID, len(organism.Cells)),
			&parentCell.Position,
			metadata,
		)
	}

	// Potential for specialization in multicellular organisms
	if organism.ComplexityLevel >= 2 && rand.Float64() < 0.3 {
		cs.specializeDaughterCell(daughterCell, organism)
	}
}

// specializeDaughterCell may specialize a daughter cell
func (cs *CellularSystem) specializeDaughterCell(cell *Cell, organism *CellularOrganism) {
	// Choose specialization based on organism needs
	cellTypes := []CellType{
		CellTypeNerve, CellTypeMuscle, CellTypeDigestive,
		CellTypeDefensive, CellTypeStorage,
	}

	oldType := cell.Type
	newType := cellTypes[rand.Intn(len(cellTypes))]
	cell.Type = newType
	cell.Specialized = true

	// Update organelles for new specialization
	cs.addOrganellesToCell(cell, cell.DNA)

	// Emit cell specialization event
	if cs.eventBus != nil {
		metadata := map[string]interface{}{
			"entity_id":        organism.EntityID,
			"cell_id":          cell.ID,
			"old_type":         cs.CellTypeNames[oldType],
			"new_type":         cs.CellTypeNames[newType],
			"complexity_level": organism.ComplexityLevel,
			"total_cells":      len(organism.Cells),
			"generation":       organism.Generation,
		}

		cs.eventBus.EmitSystemEvent(
			-1,
			"cell_specialization",
			"cellular",
			"cellular_system",
			fmt.Sprintf("Cell specialization in entity %d: cell %d changed from %s to %s",
				organism.EntityID, cell.ID, cs.CellTypeNames[oldType], cs.CellTypeNames[newType]),
			&cell.Position,
			metadata,
		)
	}
}

// handleCellDeath removes dead cells from organism
func (cs *CellularSystem) handleCellDeath(deadCell *Cell, organism *CellularOrganism) {
	// Remove dead cell from organism
	for i, cell := range organism.Cells {
		if cell.ID == deadCell.ID {
			organism.Cells = append(organism.Cells[:i], organism.Cells[i+1:]...)
			break
		}
	}

	// Remove connections to dead cell
	for _, cell := range organism.Cells {
		for i, connID := range cell.Connections {
			if connID == deadCell.ID {
				cell.Connections = append(cell.Connections[:i], cell.Connections[i+1:]...)
				break
			}
		}
	}
}

// calculateComplexityLevel determines organism complexity based on cell count
func (cs *CellularSystem) calculateComplexityLevel(cellCount int) int {
	for level := 5; level >= 1; level-- {
		if cellCount >= cs.ComplexityThresholds[level] {
			return level
		}
	}
	return 1
}

// updateOrganSystems organizes cells into functional systems
func (cs *CellularSystem) updateOrganSystems(organism *CellularOrganism) {
	// Clear existing systems
	organism.OrganSystems = make(map[string][]int)

	// Group cells by type
	cellTypeGroups := make(map[CellType][]int)
	for _, cell := range organism.Cells {
		cellTypeGroups[cell.Type] = append(cellTypeGroups[cell.Type], cell.ID)
	}

	// Create organ systems
	for cellType, cellIDs := range cellTypeGroups {
		systemName := cs.CellTypeNames[cellType]
		organism.OrganSystems[systemName] = cellIDs
	}

	// Create composite systems for complex organisms
	if organism.ComplexityLevel >= 3 {
		cs.createCompositeOrganSystems(organism)
	}
}

// createCompositeOrganSystems creates higher-level organ systems
func (cs *CellularSystem) createCompositeOrganSystems(organism *CellularOrganism) {
	// Nervous system
	if nerveIDs, exists := organism.OrganSystems["Nerve"]; exists && len(nerveIDs) >= 3 {
		organism.OrganSystems["Nervous System"] = nerveIDs
	}

	// Digestive system
	if digestiveIDs, exists := organism.OrganSystems["Digestive"]; exists && len(digestiveIDs) >= 2 {
		if storageIDs, storageExists := organism.OrganSystems["Storage"]; storageExists {
			combined := append(digestiveIDs, storageIDs...)
			organism.OrganSystems["Digestive System"] = combined
		}
	}

	// Muscular system
	if muscleIDs, exists := organism.OrganSystems["Muscle"]; exists && len(muscleIDs) >= 4 {
		organism.OrganSystems["Muscular System"] = muscleIDs
	}
}

// GetOrganismStats returns detailed stats about an organism
func (cs *CellularSystem) GetOrganismStats(entityID int) map[string]interface{} {
	organism, exists := cs.OrganismMap[entityID]
	if !exists {
		return nil
	}

	stats := make(map[string]interface{})
	stats["entity_id"] = entityID
	stats["cell_count"] = len(organism.Cells)
	stats["complexity_level"] = organism.ComplexityLevel
	stats["total_energy"] = organism.TotalEnergy
	stats["cell_divisions"] = organism.CellDivisions
	stats["generation"] = organism.Generation

	// Cell type distribution
	cellTypeCounts := make(map[string]int)
	totalActivity := 0.0
	avgHealth := 0.0
	avgAge := 0.0

	for _, cell := range organism.Cells {
		typeName := cs.CellTypeNames[cell.Type]
		cellTypeCounts[typeName]++
		totalActivity += cell.Activity
		avgHealth += cell.Health
		avgAge += float64(cell.Age)
	}

	if len(organism.Cells) > 0 {
		stats["avg_activity"] = totalActivity / float64(len(organism.Cells))
		stats["avg_health"] = avgHealth / float64(len(organism.Cells))
		stats["avg_age"] = avgAge / float64(len(organism.Cells))
	}

	stats["cell_type_distribution"] = cellTypeCounts
	stats["organ_systems"] = len(organism.OrganSystems)

	return stats
}

// GetCellularSystemStats returns overall system statistics
func (cs *CellularSystem) GetCellularSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_organisms"] = len(cs.OrganismMap)

	totalCells := 0
	complexityLevels := make(map[int]int)

	for _, organism := range cs.OrganismMap {
		totalCells += len(organism.Cells)
		complexityLevels[organism.ComplexityLevel]++
	}

	stats["total_cells"] = totalCells
	stats["complexity_distribution"] = complexityLevels
	stats["next_cell_id"] = cs.NextCellID

	return stats
}
