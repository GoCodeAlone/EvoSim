package main

import (
	"fmt"
	"math"
	"math/rand"
)

// ToolType represents different types of tools entities can create and use
type ToolType int

const (
	ToolStone       ToolType = iota // Basic stone tool
	ToolStick                       // Simple wooden tool
	ToolSpear                       // Hunting weapon
	ToolHammer                      // Construction tool
	ToolBlade                       // Cutting tool
	ToolDigger                      // Excavation tool
	ToolCrusher                     // Processing tool
	ToolContainer                   // Storage tool
	ToolFire                        // Fire-making tool
	ToolWeavingTool                 // Crafting tool
)

// getToolTypeName returns the string name for a tool type
func getToolTypeName(toolType ToolType) string {
	switch toolType {
	case ToolStone:
		return "stone"
	case ToolStick:
		return "stick"
	case ToolSpear:
		return "spear"
	case ToolHammer:
		return "hammer"
	case ToolBlade:
		return "blade"
	case ToolDigger:
		return "digger"
	case ToolCrusher:
		return "crusher"
	case ToolContainer:
		return "container"
	case ToolFire:
		return "fire"
	case ToolWeavingTool:
		return "weaving_tool"
	default:
		return "unknown"
	}
}

// getMaterialTypeName returns the string name for a material type
func getMaterialTypeName(materialType MaterialType) string {
	switch materialType {
	case MaterialStone:
		return "stone"
	case MaterialWood:
		return "wood"
	case MaterialBone:
		return "bone"
	case MaterialMetal:
		return "metal"
	case MaterialPlant:
		return "plant"
	case MaterialComposite:
		return "composite"
	default:
		return "unknown"
	}
}

// Tool represents a tool that can be created, used, and passed down
type Tool struct {
	ID            int                `json:"id"`
	Type          ToolType           `json:"type"`
	Creator       *Entity            `json:"-"` // Entity that created it
	Owner         *Entity            `json:"-"` // Current owner
	Position      Position           `json:"position"`
	Durability    float64            `json:"durability"` // 0.0 to 1.0
	MaxDurability float64            `json:"max_durability"`
	Efficiency    float64            `json:"efficiency"` // How well it performs its function
	CreatedTick   int                `json:"created_tick"`
	LastUsedTick  int                `json:"last_used_tick"`
	Material      MaterialType       `json:"material"`
	Modifications []ToolModification `json:"modifications"` // Improvements made to the tool
}

// MaterialType represents the material a tool is made from
type MaterialType int

const (
	MaterialStone MaterialType = iota
	MaterialWood
	MaterialBone
	MaterialMetal
	MaterialPlant
	MaterialComposite // Combination of materials
)

// ToolModification represents an improvement or modification made to a tool
type ToolModification struct {
	Type        ModificationType `json:"type"`
	Modifier    *Entity          `json:"-"`           // Entity that made the modification
	Improvement float64          `json:"improvement"` // Amount of improvement (0.0 to 1.0)
	AppliedTick int              `json:"applied_tick"`
}

// ModificationType represents different ways tools can be modified
type ModificationType int

const (
	ModificationSharpening ModificationType = iota
	ModificationReinforcement
	ModificationHandle
	ModificationWeight
	ModificationBalance
	ModificationBinding
)

// ToolSystem manages all tools in the simulation
type ToolSystem struct {
	Tools       map[int]*Tool           `json:"tools"` // All tools by ID
	NextToolID  int                     `json:"next_tool_id"`
	ToolRecipes map[ToolType]ToolRecipe `json:"tool_recipes"` // How to create tools
	eventBus    *CentralEventBus        `json:"-"`            // Event tracking
}

// ToolRecipe defines what's needed to create a tool
type ToolRecipe struct {
	Type              ToolType                 `json:"type"`
	RequiredMaterials map[MaterialType]float64 `json:"required_materials"`
	RequiredSkill     float64                  `json:"required_skill"` // Intelligence threshold
	RequiredEnergy    float64                  `json:"required_energy"`
	CreationTime      int                      `json:"creation_time"` // Ticks to create
	BaseDurability    float64                  `json:"base_durability"`
	BaseEfficiency    float64                  `json:"base_efficiency"`
}

// NewToolSystem creates a new tool management system
func NewToolSystem(eventBus *CentralEventBus) *ToolSystem {
	ts := &ToolSystem{
		Tools:       make(map[int]*Tool),
		NextToolID:  1,
		ToolRecipes: make(map[ToolType]ToolRecipe),
		eventBus:    eventBus,
	}

	// Initialize tool recipes
	ts.initializeRecipes()

	return ts
}

// initializeRecipes sets up the basic tool creation recipes
func (ts *ToolSystem) initializeRecipes() {
	ts.ToolRecipes[ToolStone] = ToolRecipe{
		Type:              ToolStone,
		RequiredMaterials: map[MaterialType]float64{MaterialStone: 1.0},
		RequiredSkill:     0.1,
		RequiredEnergy:    5.0,
		CreationTime:      2,
		BaseDurability:    0.6,
		BaseEfficiency:    0.4,
	}

	ts.ToolRecipes[ToolStick] = ToolRecipe{
		Type:              ToolStick,
		RequiredMaterials: map[MaterialType]float64{MaterialWood: 1.0},
		RequiredSkill:     0.05,
		RequiredEnergy:    3.0,
		CreationTime:      1,
		BaseDurability:    0.4,
		BaseEfficiency:    0.3,
	}

	ts.ToolRecipes[ToolSpear] = ToolRecipe{
		Type:              ToolSpear,
		RequiredMaterials: map[MaterialType]float64{MaterialWood: 1.0, MaterialStone: 0.5},
		RequiredSkill:     0.3,
		RequiredEnergy:    10.0,
		CreationTime:      5,
		BaseDurability:    0.8,
		BaseEfficiency:    0.7,
	}

	ts.ToolRecipes[ToolHammer] = ToolRecipe{
		Type:              ToolHammer,
		RequiredMaterials: map[MaterialType]float64{MaterialStone: 2.0, MaterialWood: 0.5},
		RequiredSkill:     0.4,
		RequiredEnergy:    15.0,
		CreationTime:      7,
		BaseDurability:    0.9,
		BaseEfficiency:    0.8,
	}

	ts.ToolRecipes[ToolDigger] = ToolRecipe{
		Type:              ToolDigger,
		RequiredMaterials: map[MaterialType]float64{MaterialStone: 1.5, MaterialWood: 1.0},
		RequiredSkill:     0.25,
		RequiredEnergy:    8.0,
		CreationTime:      4,
		BaseDurability:    0.7,
		BaseEfficiency:    0.6,
	}

	ts.ToolRecipes[ToolContainer] = ToolRecipe{
		Type:              ToolContainer,
		RequiredMaterials: map[MaterialType]float64{MaterialPlant: 2.0},
		RequiredSkill:     0.2,
		RequiredEnergy:    6.0,
		CreationTime:      3,
		BaseDurability:    0.5,
		BaseEfficiency:    0.5,
	}
}

// CreateTool attempts to create a tool for an entity
func (ts *ToolSystem) CreateTool(creator *Entity, toolType ToolType, position Position) *Tool {
	recipe, exists := ts.ToolRecipes[toolType]
	if !exists {
		return nil
	}

	// Check if entity has required skill (intelligence)
	intelligence := creator.GetTrait("intelligence")
	if intelligence < recipe.RequiredSkill {
		return nil
	}

	// Check if entity has enough energy
	if creator.Energy < recipe.RequiredEnergy {
		return nil
	}

	// Simulate material gathering for now (could be expanded to require actual materials)
	availableMaterials := ts.getAvailableMaterials(position)
	for materialType, required := range recipe.RequiredMaterials {
		if availableMaterials[materialType] < required {
			return nil // Not enough materials available
		}
	}

	// Create the tool
	tool := &Tool{
		ID:            ts.NextToolID,
		Type:          toolType,
		Creator:       creator,
		Owner:         creator,
		Position:      position,
		Durability:    recipe.BaseDurability,
		MaxDurability: recipe.BaseDurability,
		Efficiency:    recipe.BaseEfficiency,
		CreatedTick:   0, // Will be set by caller
		LastUsedTick:  0,
		Material:      ts.getPrimaryMaterial(recipe),
		Modifications: make([]ToolModification, 0),
	}

	// Improve tool quality based on creator's skill
	skillBonus := (intelligence - recipe.RequiredSkill) * 0.5
	tool.Efficiency = math.Min(1.0, tool.Efficiency+skillBonus)
	tool.Durability = math.Min(1.0, tool.Durability+skillBonus*0.3)
	tool.MaxDurability = tool.Durability

	// Consume energy
	creator.Energy -= recipe.RequiredEnergy

	// Add to system
	ts.Tools[tool.ID] = tool
	ts.NextToolID++

	// Emit tool creation event
	if ts.eventBus != nil {
		metadata := map[string]interface{}{
			"tool_id":         tool.ID,
			"tool_type":       getToolTypeName(toolType),
			"creator_id":      creator.ID,
			"creator_species": creator.Species,
			"intelligence":    intelligence,
			"skill_bonus":     skillBonus,
			"energy_cost":     recipe.RequiredEnergy,
			"durability":      tool.Durability,
			"efficiency":      tool.Efficiency,
			"material":        getMaterialTypeName(tool.Material),
			"creation_time":   recipe.CreationTime,
		}

		ts.eventBus.EmitSystemEvent(
			0, // Will be updated by caller with actual tick
			"tool_created",
			"tools",
			"tool_system",
			fmt.Sprintf("Entity %d (%s) created %s tool %d (efficiency: %.2f, durability: %.2f)",
				creator.ID, creator.Species, getToolTypeName(toolType), tool.ID, tool.Efficiency, tool.Durability),
			&position,
			metadata,
		)
	}

	return tool
}

// getAvailableMaterials simulates material availability at a position
func (ts *ToolSystem) getAvailableMaterials(position Position) map[MaterialType]float64 {
	// Simplified: assume all materials are available
	// In a full implementation, this would check nearby resources
	return map[MaterialType]float64{
		MaterialStone: 10.0,
		MaterialWood:  8.0,
		MaterialBone:  2.0,
		MaterialPlant: 15.0,
	}
}

// getPrimaryMaterial determines the primary material for a tool recipe
func (ts *ToolSystem) getPrimaryMaterial(recipe ToolRecipe) MaterialType {
	maxAmount := 0.0
	var primaryMaterial MaterialType

	for material, amount := range recipe.RequiredMaterials {
		if amount > maxAmount {
			maxAmount = amount
			primaryMaterial = material
		}
	}

	return primaryMaterial
}

// UseTool uses a tool for a specific purpose and reduces durability
func (ts *ToolSystem) UseTool(tool *Tool, user *Entity, intensity float64) float64 {
	if tool == nil || !tool.IsUsable() {
		return 0.0
	}

	oldDurability := tool.Durability
	oldOwner := tool.Owner

	// Calculate effectiveness based on tool efficiency and user skill
	userSkill := user.GetTrait("intelligence")
	effectiveness := tool.Efficiency * (0.5 + userSkill*0.5) * intensity

	// Reduce durability based on use intensity
	durabilityLoss := intensity * 0.05 * (1.0 + rand.Float64()*0.2)
	tool.Durability = math.Max(0.0, tool.Durability-durabilityLoss)

	// Update last used tick
	tool.LastUsedTick = 0 // Will be set by caller

	// Transfer ownership if different user
	if tool.Owner != user {
		tool.Owner = user
	}

	// Emit tool usage event
	if ts.eventBus != nil {
		metadata := map[string]interface{}{
			"tool_id":           tool.ID,
			"tool_type":         getToolTypeName(tool.Type),
			"user_id":           user.ID,
			"user_species":      user.Species,
			"user_skill":        userSkill,
			"intensity":         intensity,
			"effectiveness":     effectiveness,
			"durability_before": oldDurability,
			"durability_after":  tool.Durability,
			"durability_loss":   durabilityLoss,
			"ownership_changed": oldOwner != user,
			"tool_broken":       tool.Durability <= 0,
		}

		eventType := "tool_used"
		if tool.Durability <= 0 {
			eventType = "tool_broken"
		}

		ts.eventBus.EmitSystemEvent(
			0, // Will be updated by caller
			eventType,
			"tools",
			"tool_system",
			fmt.Sprintf("Entity %d used %s tool %d (effectiveness: %.2f, durability: %.2f -> %.2f)",
				user.ID, getToolTypeName(tool.Type), tool.ID, effectiveness, oldDurability, tool.Durability),
			&tool.Position,
			metadata,
		)
	}

	return effectiveness
}

// ModifyTool applies a modification to improve a tool
func (ts *ToolSystem) ModifyTool(tool *Tool, modifier *Entity, modificationType ModificationType) bool {
	if tool == nil || !tool.IsUsable() {
		return false
	}

	// Check if modifier has sufficient skill
	modifierSkill := modifier.GetTrait("intelligence")
	requiredSkill := 0.3 + float64(len(tool.Modifications))*0.1

	if modifierSkill < requiredSkill {
		return false
	}

	// Check energy requirements
	energyCost := 5.0 + float64(len(tool.Modifications))*2.0
	if modifier.Energy < energyCost {
		return false
	}

	// Apply modification
	improvement := modifierSkill * 0.2 * (1.0 + rand.Float64()*0.3)

	modification := ToolModification{
		Type:        modificationType,
		Modifier:    modifier,
		Improvement: improvement,
		AppliedTick: 0, // Will be set by caller
	}

	// Apply the improvement to the tool
	switch modificationType {
	case ModificationSharpening:
		tool.Efficiency = math.Min(1.0, tool.Efficiency+improvement*0.3)

	case ModificationReinforcement:
		tool.MaxDurability = math.Min(1.0, tool.MaxDurability+improvement*0.2)
		tool.Durability = math.Min(tool.MaxDurability, tool.Durability+improvement*0.1)

	case ModificationHandle:
		tool.Efficiency = math.Min(1.0, tool.Efficiency+improvement*0.15)

	case ModificationWeight:
		tool.Efficiency = math.Min(1.0, tool.Efficiency+improvement*0.1)

	case ModificationBalance:
		tool.Efficiency = math.Min(1.0, tool.Efficiency+improvement*0.2)

	case ModificationBinding:
		tool.MaxDurability = math.Min(1.0, tool.MaxDurability+improvement*0.15)
	}

	tool.Modifications = append(tool.Modifications, modification)
	modifier.Energy -= energyCost

	return true
}

// IsUsable checks if a tool can still be used
func (tool *Tool) IsUsable() bool {
	return tool.Durability > 0.1 // Tools become unusable when very damaged
}

// GetToolEffectiveness returns the current effectiveness of a tool
func (tool *Tool) GetToolEffectiveness() float64 {
	if !tool.IsUsable() {
		return 0.0
	}

	// Effectiveness decreases as durability decreases
	durabilityFactor := math.Sqrt(tool.Durability / tool.MaxDurability)
	return tool.Efficiency * durabilityFactor
}

// RepairTool attempts to repair a damaged tool
func (ts *ToolSystem) RepairTool(tool *Tool, repairer *Entity) bool {
	if tool == nil || tool.Durability >= tool.MaxDurability {
		return false
	}

	repairerSkill := repairer.GetTrait("intelligence")
	energyCost := 8.0

	if repairer.Energy < energyCost {
		return false
	}

	// Repair amount based on skill
	repairAmount := repairerSkill * 0.3 * (1.0 + rand.Float64()*0.2)
	tool.Durability = math.Min(tool.MaxDurability, tool.Durability+repairAmount)

	repairer.Energy -= energyCost

	return true
}

// DropTool removes a tool from an entity's possession
func (ts *ToolSystem) DropTool(tool *Tool, position Position) {
	if tool != nil {
		tool.Owner = nil
		tool.Position = position
	}
}

// PickupTool allows an entity to pick up a tool
func (ts *ToolSystem) PickupTool(tool *Tool, entity *Entity) bool {
	if tool == nil || tool.Owner != nil {
		return false
	}

	// Check if entity is close enough to the tool
	distance := math.Sqrt(math.Pow(entity.Position.X-tool.Position.X, 2) +
		math.Pow(entity.Position.Y-tool.Position.Y, 2))

	if distance > 2.0 { // Maximum pickup distance
		return false
	}

	tool.Owner = entity
	return true
}

// GetEntityTools returns all tools owned by an entity
func (ts *ToolSystem) GetEntityTools(entity *Entity) []*Tool {
	tools := make([]*Tool, 0)

	for _, tool := range ts.Tools {
		if tool.Owner == entity {
			tools = append(tools, tool)
		}
	}

	return tools
}

// GetNearbyTools returns tools near a position
func (ts *ToolSystem) GetNearbyTools(position Position, radius float64) []*Tool {
	tools := make([]*Tool, 0)

	for _, tool := range ts.Tools {
		if tool.Owner == nil { // Only unowned tools
			distance := math.Sqrt(math.Pow(position.X-tool.Position.X, 2) +
				math.Pow(position.Y-tool.Position.Y, 2))

			if distance <= radius {
				tools = append(tools, tool)
			}
		}
	}

	return tools
}

// UpdateTools maintains all tools in the system
func (ts *ToolSystem) UpdateTools(tick int) {
	for _, tool := range ts.Tools {
		// Natural decay for unused tools
		if tool.Owner == nil && tick-tool.LastUsedTick > 100 {
			tool.Durability *= 0.999 // Very slow decay
		}

		// Remove completely broken tools
		if tool.Durability <= 0.0 {
			delete(ts.Tools, tool.ID)
		}
	}
}

// GetToolStats returns statistics about the tool system
func (ts *ToolSystem) GetToolStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalTools := len(ts.Tools)
	ownedTools := 0
	avgDurability := 0.0
	avgEfficiency := 0.0

	toolTypeCounts := make(map[ToolType]int)

	for _, tool := range ts.Tools {
		if tool.Owner != nil {
			ownedTools++
		}

		avgDurability += tool.Durability
		avgEfficiency += tool.Efficiency
		toolTypeCounts[tool.Type]++
	}

	if totalTools > 0 {
		avgDurability /= float64(totalTools)
		avgEfficiency /= float64(totalTools)
	}

	stats["total_tools"] = totalTools
	stats["owned_tools"] = ownedTools
	stats["dropped_tools"] = totalTools - ownedTools
	stats["avg_durability"] = avgDurability
	stats["avg_efficiency"] = avgEfficiency
	stats["tool_types"] = toolTypeCounts

	return stats
}

// GetToolTypeName returns the name of a tool type
func GetToolTypeName(toolType ToolType) string {
	names := map[ToolType]string{
		ToolStone:       "Stone",
		ToolStick:       "Stick",
		ToolSpear:       "Spear",
		ToolHammer:      "Hammer",
		ToolBlade:       "Blade",
		ToolDigger:      "Digger",
		ToolCrusher:     "Crusher",
		ToolContainer:   "Container",
		ToolFire:        "Fire Tool",
		ToolWeavingTool: "Weaving Tool",
	}

	if name, exists := names[toolType]; exists {
		return name
	}
	return "Unknown Tool"
}

// GetMaterialTypeName returns the name of a material type
func GetMaterialTypeName(materialType MaterialType) string {
	names := map[MaterialType]string{
		MaterialStone:     "Stone",
		MaterialWood:      "Wood",
		MaterialBone:      "Bone",
		MaterialMetal:     "Metal",
		MaterialPlant:     "Plant",
		MaterialComposite: "Composite",
	}

	if name, exists := names[materialType]; exists {
		return name
	}
	return "Unknown Material"
}
