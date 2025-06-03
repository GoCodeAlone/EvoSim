package main

import (
	"fmt"
	"math"
	"math/rand"
)

// Nucleotide represents the basic building blocks of DNA
type Nucleotide rune

const (
	Adenine  Nucleotide = 'A'
	Thymine  Nucleotide = 'T'
	Guanine  Nucleotide = 'G'
	Cytosine Nucleotide = 'C'
)

// Gene represents a functional unit of heredity
type Gene struct {
	Name      string       `json:"name"`       // Gene name (e.g., "size", "strength")
	Sequence  []Nucleotide `json:"sequence"`   // DNA sequence
	Dominant  bool         `json:"dominant"`   // Whether this allele is dominant
	Expression float64     `json:"expression"` // How strongly the gene is expressed (0-1)
}

// Chromosome represents a collection of genes
type Chromosome struct {
	ID    int    `json:"id"`    // Chromosome identifier
	Genes []Gene `json:"genes"` // Genes on this chromosome
}

// DNAStrand represents a complete DNA strand with multiple chromosomes
type DNAStrand struct {
	EntityID    int          `json:"entity_id"`    // Entity this DNA belongs to
	Chromosomes []Chromosome `json:"chromosomes"`  // All chromosomes
	Mutations   int          `json:"mutations"`    // Total mutations accumulated
	Generation  int          `json:"generation"`   // Generation number
}

// DNASystem manages the DNA-based genetic system
type DNASystem struct {
	TraitToGene    map[string]string  // Maps trait names to gene names
	GeneLength     map[string]int     // Standard length for each gene type
	MutationRates  map[string]float64 // Mutation rates for different gene types
	DominanceRules map[string]bool    // Default dominance for genes
	eventBus       *CentralEventBus   `json:"-"` // Event tracking
}

// NewDNASystem creates a new DNA management system
func NewDNASystem(eventBus *CentralEventBus) *DNASystem {
	return &DNASystem{
		TraitToGene: map[string]string{
			// Basic traits
			"size":        "SIZE",
			"strength":    "STR",
			"speed":       "SPD",
			"aggression":  "AGG",
			"intelligence": "INT",
			"vision":      "VIS",
			"defense":     "DEF",
			"energy":      "ENE",
			"reproduction": "REP",
			"cooperation":  "COO",
			"camouflage":  "CAM",
			"toxicity":    "TOX",
			"longevity":   "LON",
			"adaptation":  "ADA",
			"metabolism":  "MET",
		},
		GeneLength: map[string]int{
			"SIZE": 12, "STR": 10, "SPD": 8, "AGG": 6,
			"INT": 14, "VIS": 10, "DEF": 8, "ENE": 12,
			"REP": 10, "COO": 8, "CAM": 6, "TOX": 8,
			"LON": 12, "ADA": 14, "MET": 10,
		},
		MutationRates: map[string]float64{
			"SIZE": 0.001, "STR": 0.0012, "SPD": 0.0015, "AGG": 0.002,
			"INT": 0.0008, "VIS": 0.0012, "DEF": 0.0012, "ENE": 0.001,
			"REP": 0.0008, "COO": 0.0015, "CAM": 0.002, "TOX": 0.0018,
			"LON": 0.0006, "ADA": 0.0005, "MET": 0.001,
		},
		DominanceRules: map[string]bool{
			"SIZE": true, "STR": true, "SPD": false, "AGG": true,
			"INT": false, "VIS": false, "DEF": true, "ENE": false,
			"REP": false, "COO": false, "CAM": false, "TOX": true,
			"LON": false, "ADA": false, "MET": false,
		},
		eventBus: eventBus,
	}
}

// GenerateRandomDNA creates a random DNA strand for a new entity
func (ds *DNASystem) GenerateRandomDNA(entityID int, generation int) *DNAStrand {
	dna := &DNAStrand{
		EntityID:    entityID,
		Chromosomes: make([]Chromosome, 0),
		Mutations:   0,
		Generation:  generation,
	}

	// Create chromosomes (typically 2 for diploid organisms)
	for chromID := 0; chromID < 2; chromID++ {
		chromosome := Chromosome{
			ID:    chromID,
			Genes: make([]Gene, 0),
		}

		// Add genes for each trait
		for trait, geneName := range ds.TraitToGene {
			gene := ds.generateRandomGene(geneName, trait)
			chromosome.Genes = append(chromosome.Genes, gene)
		}

		dna.Chromosomes = append(dna.Chromosomes, chromosome)
	}

	return dna
}

// generateRandomGene creates a random gene sequence
func (ds *DNASystem) generateRandomGene(geneName, traitName string) Gene {
	length := ds.GeneLength[geneName]
	if length == 0 {
		length = 10 // Default length
	}

	sequence := make([]Nucleotide, length)
	nucleotides := []Nucleotide{Adenine, Thymine, Guanine, Cytosine}

	for i := 0; i < length; i++ {
		sequence[i] = nucleotides[rand.Intn(4)]
	}

	dominant := ds.DominanceRules[geneName]
	if rand.Float64() < 0.3 { // 30% chance to flip dominance
		dominant = !dominant
	}

	return Gene{
		Name:       geneName,
		Sequence:   sequence,
		Dominant:   dominant,
		Expression: rand.Float64()*0.8 + 0.2, // 0.2 to 1.0
	}
}

// ExpressTrait converts DNA information into trait values
func (ds *DNASystem) ExpressTrait(dna *DNAStrand, traitName string) float64 {
	geneName := ds.TraitToGene[traitName]
	if geneName == "" {
		return 0.0
	}

	// Find genes for this trait on both chromosomes
	var genes []Gene
	for _, chromosome := range dna.Chromosomes {
		for _, gene := range chromosome.Genes {
			if gene.Name == geneName {
				genes = append(genes, gene)
			}
		}
	}

	if len(genes) == 0 {
		return 0.0
	}

	// Calculate trait value based on genetic expression
	var totalValue float64
	var totalWeight float64

	for _, gene := range genes {
		// Convert DNA sequence to numeric value
		sequenceValue := ds.sequenceToValue(gene.Sequence)
		
		// Apply dominance and expression
		weight := gene.Expression
		if gene.Dominant && len(genes) > 1 {
			weight *= 1.5 // Dominant genes have more influence
		}

		totalValue += sequenceValue * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Normalize to range -1 to 1 (compatible with existing trait system)
	normalizedValue := (totalValue / totalWeight) * 2 - 1
	return math.Max(-1.0, math.Min(1.0, normalizedValue))
}

// sequenceToValue converts a DNA sequence to a numeric value
func (ds *DNASystem) sequenceToValue(sequence []Nucleotide) float64 {
	if len(sequence) == 0 {
		return 0.5
	}

	// Simple approach: count nucleotide frequencies
	counts := map[Nucleotide]int{Adenine: 0, Thymine: 0, Guanine: 0, Cytosine: 0}
	for _, nucleotide := range sequence {
		counts[nucleotide]++
	}

	// Calculate weighted value based on nucleotide composition
	total := float64(len(sequence))
	value := (float64(counts[Adenine])*0.1 + float64(counts[Thymine])*0.3 + 
		     float64(counts[Guanine])*0.7 + float64(counts[Cytosine])*0.9) / total

	return value
}

// MutateDNA applies mutations to DNA based on environmental factors
func (ds *DNASystem) MutateDNA(dna *DNAStrand, mutationPressure float64) {
	initialMutations := dna.Mutations
	
	for chromIdx := range dna.Chromosomes {
		for geneIdx := range dna.Chromosomes[chromIdx].Genes {
			gene := &dna.Chromosomes[chromIdx].Genes[geneIdx]
			
			// Get base mutation rate for this gene
			baseMutationRate := ds.MutationRates[gene.Name]
			if baseMutationRate == 0 {
				baseMutationRate = 0.001
			}

			// Apply environmental pressure
			effectiveMutationRate := baseMutationRate * (1.0 + mutationPressure)

			// Track mutations for this gene
			geneMutations := 0
			
			// Mutate individual nucleotides
			oldSequence := make([]Nucleotide, len(gene.Sequence))
			copy(oldSequence, gene.Sequence)
			
			for seqIdx := range gene.Sequence {
				if rand.Float64() < effectiveMutationRate {
					// Point mutation - change nucleotide
					nucleotides := []Nucleotide{Adenine, Thymine, Guanine, Cytosine}
					oldNucleotide := gene.Sequence[seqIdx]
					gene.Sequence[seqIdx] = nucleotides[rand.Intn(4)]
					dna.Mutations++
					geneMutations++
					
					// Emit point mutation event
					if ds.eventBus != nil {
						metadata := map[string]interface{}{
							"entity_id":         dna.EntityID,
							"chromosome_id":     chromIdx,
							"gene_name":         gene.Name,
							"gene_index":        geneIdx,
							"sequence_position": seqIdx,
							"old_nucleotide":    string(oldNucleotide),
							"new_nucleotide":    string(gene.Sequence[seqIdx]),
							"mutation_rate":     effectiveMutationRate,
							"mutation_pressure": mutationPressure,
							"generation":        dna.Generation,
						}
						
						ds.eventBus.EmitSystemEvent(
							-1, // No specific tick for DNA operations
							"point_mutation",
							"dna",
							"dna_system",
							fmt.Sprintf("Point mutation in entity %d: %s gene position %d (%c -> %c)", 
								dna.EntityID, gene.Name, seqIdx, oldNucleotide, gene.Sequence[seqIdx]),
							nil,
							metadata,
						)
					}
				}
			}

			// Occasional dominance shifts
			if rand.Float64() < effectiveMutationRate*0.1 {
				oldDominant := gene.Dominant
				gene.Dominant = !gene.Dominant
				dna.Mutations++
				geneMutations++
				
				// Emit dominance shift event
				if ds.eventBus != nil {
					metadata := map[string]interface{}{
						"entity_id":     dna.EntityID,
						"chromosome_id": chromIdx,
						"gene_name":     gene.Name,
						"old_dominant":  oldDominant,
						"new_dominant":  gene.Dominant,
						"generation":    dna.Generation,
					}
					
					ds.eventBus.EmitSystemEvent(
						-1,
						"dominance_shift",
						"dna",
						"dna_system",
						fmt.Sprintf("Dominance shift in entity %d: %s gene (%t -> %t)", 
							dna.EntityID, gene.Name, oldDominant, gene.Dominant),
						nil,
						metadata,
					)
				}
			}

			// Expression level mutations
			if rand.Float64() < effectiveMutationRate*0.5 {
				oldExpression := gene.Expression
				change := (rand.Float64() - 0.5) * 0.2
				gene.Expression = math.Max(0.1, math.Min(1.0, gene.Expression+change))
				
				// Emit expression change event
				if ds.eventBus != nil && math.Abs(change) > 0.05 { // Only for significant changes
					metadata := map[string]interface{}{
						"entity_id":       dna.EntityID,
						"chromosome_id":   chromIdx,
						"gene_name":       gene.Name,
						"old_expression":  oldExpression,
						"new_expression":  gene.Expression,
						"expression_change": change,
						"generation":      dna.Generation,
					}
					
					ds.eventBus.EmitSystemEvent(
						-1,
						"expression_change",
						"dna",
						"dna_system",
						fmt.Sprintf("Expression change in entity %d: %s gene (%.3f -> %.3f)", 
							dna.EntityID, gene.Name, oldExpression, gene.Expression),
						nil,
						metadata,
					)
				}
			}
		}
	}
	
	// Emit overall mutation summary if any mutations occurred
	totalNewMutations := dna.Mutations - initialMutations
	if ds.eventBus != nil && totalNewMutations > 0 {
		metadata := map[string]interface{}{
			"entity_id":         dna.EntityID,
			"total_mutations":   dna.Mutations,
			"new_mutations":     totalNewMutations,
			"mutation_pressure": mutationPressure,
			"generation":        dna.Generation,
		}
		
		ds.eventBus.EmitSystemEvent(
			-1,
			"dna_mutation_summary",
			"dna",
			"dna_system",
			fmt.Sprintf("Entity %d accumulated %d new mutations (total: %d)", 
				dna.EntityID, totalNewMutations, dna.Mutations),
			nil,
			metadata,
		)
	}
}

// CrossoverDNA performs genetic crossover between two parent DNA strands
func (ds *DNASystem) CrossoverDNA(parent1, parent2 *DNAStrand, childID int) *DNAStrand {
	child := &DNAStrand{
		EntityID:    childID,
		Chromosomes: make([]Chromosome, len(parent1.Chromosomes)),
		Mutations:   0,
		Generation:  int(math.Max(float64(parent1.Generation), float64(parent2.Generation))) + 1,
	}

	// Perform crossover for each chromosome
	for chromIdx := range child.Chromosomes {
		child.Chromosomes[chromIdx] = Chromosome{
			ID:    chromIdx,
			Genes: make([]Gene, 0),
		}

		// Get genes from both parents
		parent1Genes := parent1.Chromosomes[chromIdx].Genes
		parent2Genes := parent2.Chromosomes[chromIdx].Genes

		// Combine genes from both parents
		maxGenes := int(math.Max(float64(len(parent1Genes)), float64(len(parent2Genes))))
		
		for geneIdx := 0; geneIdx < maxGenes; geneIdx++ {
			var selectedGene Gene
			
			if geneIdx < len(parent1Genes) && geneIdx < len(parent2Genes) {
				// Both parents have this gene - choose randomly or recombine
				if rand.Float64() < 0.5 {
					selectedGene = parent1Genes[geneIdx]
				} else {
					selectedGene = parent2Genes[geneIdx]
				}

				// Possibility of recombination within gene
				if rand.Float64() < 0.1 {
					selectedGene = ds.recombineGenes(parent1Genes[geneIdx], parent2Genes[geneIdx])
				}
			} else if geneIdx < len(parent1Genes) {
				selectedGene = parent1Genes[geneIdx]
			} else if geneIdx < len(parent2Genes) {
				selectedGene = parent2Genes[geneIdx]
			} else {
				continue
			}

			child.Chromosomes[chromIdx].Genes = append(child.Chromosomes[chromIdx].Genes, selectedGene)
		}
	}

	// Emit crossover event
	if ds.eventBus != nil {
		metadata := map[string]interface{}{
			"child_id":            childID,
			"parent1_id":          parent1.EntityID,
			"parent2_id":          parent2.EntityID,
			"child_generation":    child.Generation,
			"parent1_generation":  parent1.Generation,
			"parent2_generation":  parent2.Generation,
			"parent1_mutations":   parent1.Mutations,
			"parent2_mutations":   parent2.Mutations,
			"chromosome_count":    len(child.Chromosomes),
		}
		
		ds.eventBus.EmitSystemEvent(
			-1,
			"dna_crossover",
			"dna",
			"dna_system",
			fmt.Sprintf("DNA crossover: entity %d created from parents %d and %d (generation %d)", 
				childID, parent1.EntityID, parent2.EntityID, child.Generation),
			nil,
			metadata,
		)
	}

	return child
}

// recombineGenes performs intra-gene recombination
func (ds *DNASystem) recombineGenes(gene1, gene2 Gene) Gene {
	// Ensure sequences are the same length
	minLength := int(math.Min(float64(len(gene1.Sequence)), float64(len(gene2.Sequence))))
	if minLength == 0 {
		return gene1
	}

	// Random crossover point
	crossoverPoint := rand.Intn(minLength)
	
	newSequence := make([]Nucleotide, minLength)
	
	// Copy from gene1 up to crossover point
	copy(newSequence[:crossoverPoint], gene1.Sequence[:crossoverPoint])
	
	// Copy from gene2 after crossover point
	if crossoverPoint < minLength {
		copy(newSequence[crossoverPoint:], gene2.Sequence[crossoverPoint:minLength])
	}

	// Combine other properties
	return Gene{
		Name:       gene1.Name,
		Sequence:   newSequence,
		Dominant:   gene1.Dominant || gene2.Dominant, // At least one dominant
		Expression: (gene1.Expression + gene2.Expression) / 2.0,
	}
}

// AnalyzeDNA provides detailed analysis of a DNA strand
func (ds *DNASystem) AnalyzeDNA(dna *DNAStrand) map[string]interface{} {
	analysis := make(map[string]interface{})
	
	analysis["entity_id"] = dna.EntityID
	analysis["generation"] = dna.Generation
	analysis["total_mutations"] = dna.Mutations
	analysis["chromosome_count"] = len(dna.Chromosomes)
	
	// Gene analysis
	geneCount := 0
	dominantGenes := 0
	avgExpression := 0.0
	
	for _, chromosome := range dna.Chromosomes {
		geneCount += len(chromosome.Genes)
		for _, gene := range chromosome.Genes {
			if gene.Dominant {
				dominantGenes++
			}
			avgExpression += gene.Expression
		}
	}
	
	if geneCount > 0 {
		avgExpression /= float64(geneCount)
	}
	
	analysis["total_genes"] = geneCount
	analysis["dominant_genes"] = dominantGenes
	analysis["avg_expression"] = avgExpression
	
	// Nucleotide composition
	totalNucleotides := map[Nucleotide]int{Adenine: 0, Thymine: 0, Guanine: 0, Cytosine: 0}
	totalLength := 0
	
	for _, chromosome := range dna.Chromosomes {
		for _, gene := range chromosome.Genes {
			for _, nucleotide := range gene.Sequence {
				totalNucleotides[nucleotide]++
				totalLength++
			}
		}
	}
	
	if totalLength > 0 {
		analysis["nucleotide_composition"] = map[string]float64{
			"A": float64(totalNucleotides[Adenine]) / float64(totalLength),
			"T": float64(totalNucleotides[Thymine]) / float64(totalLength),
			"G": float64(totalNucleotides[Guanine]) / float64(totalLength),
			"C": float64(totalNucleotides[Cytosine]) / float64(totalLength),
		}
	}
	
	return analysis
}

// GetDNAString returns a human-readable DNA sequence string
func (ds *DNASystem) GetDNAString(dna *DNAStrand, maxLength int) string {
	var result string
	totalLength := 0
	
	for chromIdx, chromosome := range dna.Chromosomes {
		if chromIdx > 0 {
			result += " | "
		}
		result += fmt.Sprintf("Chr%d: ", chromIdx+1)
		
		for geneIdx, gene := range chromosome.Genes {
			if geneIdx > 0 {
				result += "-"
			}
			
			geneStr := ""
			for _, nucleotide := range gene.Sequence {
				geneStr += string(nucleotide)
				totalLength++
				if totalLength >= maxLength {
					return result + geneStr + "..."
				}
			}
			result += geneStr
		}
	}
	
	return result
}