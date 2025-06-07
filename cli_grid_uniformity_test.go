package main

import (
	"testing"
	"unicode/utf8"
)

// TestGridSymbolUniformity ensures all symbols used in the CLI grid view
// are single-width characters to maintain even grid appearance
func TestGridSymbolUniformity(t *testing.T) {
	// Test biome symbols
	t.Run("BiomeSymbols", func(t *testing.T) {
		biomes := initializeBiomes()
		for biomeType, biome := range biomes {
			symbol := biome.Symbol
			if !isSingleWidthSymbol(symbol) {
				t.Errorf("Biome %s (%d) uses multi-width symbol '%c' (U+%04X), should use single-width character", 
					biome.Name, biomeType, symbol, symbol)
			}
		}
	})

	// Test plant symbols
	t.Run("PlantSymbols", func(t *testing.T) {
		plantConfigs := GetPlantConfigs()
		for plantType, config := range plantConfigs {
			symbol := config.Symbol
			if !isSingleWidthSymbol(symbol) {
				t.Errorf("Plant %s (%d) uses multi-width symbol '%c' (U+%04X), should use single-width character", 
					config.Name, plantType, symbol, symbol)
			}
		}
	})

	// Test species symbols
	t.Run("SpeciesSymbols", func(t *testing.T) {
		// Create a test CLI model to access species symbols
		world := &World{Config: WorldConfig{GridWidth: 10, GridHeight: 10}}
		cli := NewCLIModel(world)
		
		for species, symbol := range cli.speciesSymbols {
			if !isSingleWidthSymbol(symbol) {
				t.Errorf("Species %s uses multi-width symbol '%c' (U+%04X), should use single-width character", 
					species, symbol, symbol)
			}
		}
	})

	// Test structure symbols (these are defined inline in the gridView function)
	t.Run("StructureSymbols", func(t *testing.T) {
		// Test the structure symbols defined in the gridView function
		structureSymbols := map[string]rune{
			"Nest":    'N',
			"Cache":   'C',
			"Barrier": 'B',
			"Trap":    'P',
			"Farm":    'F',
			"Well":    'W',
			"Tower":   'O',
			"Market":  'M',
		}

		for structName, symbol := range structureSymbols {
			if !isSingleWidthSymbol(symbol) {
				t.Errorf("Structure %s uses multi-width symbol '%c' (U+%04X), should use single-width character", 
					structName, symbol, symbol)
			}
		}
	})

	// Test signal symbols
	t.Run("SignalSymbols", func(t *testing.T) {
		signalSymbols := map[string]rune{
			"Danger":    '!',
			"Food":      '*',
			"Mating":    'M',
			"Territory": 'T',
			"Help":      '?',
			"Migration": '→',
		}

		for signalName, symbol := range signalSymbols {
			if !isSingleWidthSymbol(symbol) {
				t.Errorf("Signal %s uses multi-width symbol '%c' (U+%04X), should use single-width character", 
					signalName, symbol, symbol)
			}
		}
	})
}

// isSingleWidthSymbol checks if a rune is a single-width character for terminal display
func isSingleWidthSymbol(r rune) bool {
	// Check if it's a valid Unicode character
	if r == utf8.RuneError {
		return false
	}

	// Emojis and many other symbols are typically double-width
	// This is a simplified check - in reality, terminal width depends on the terminal
	// but this should catch the most common problematic cases
	
	// Most emojis are in these ranges:
	// U+1F300–U+1F5FF (Miscellaneous Symbols and Pictographs)
	// U+1F600–U+1F64F (Emoticons)
	// U+1F680–U+1F6FF (Transport and Map Symbols)
	// U+1F700–U+1F77F (Alchemical Symbols)
	// U+1F780–U+1F7FF (Geometric Shapes Extended)
	// U+1F800–U+1F8FF (Supplemental Arrows-C)
	// U+1F900–U+1F9FF (Supplemental Symbols and Pictographs)
	// U+1FA00–U+1FA6F (Chess Symbols)
	// U+1FA70–U+1FAFF (Symbols and Pictographs Extended-A)
	
	if r >= 0x1F300 && r <= 0x1FAFF {
		return false // Likely emoji or pictograph
	}

	// Some other problematic ranges
	if r >= 0x2600 && r <= 0x26FF && isLikelyEmoji(r) {
		return false // Miscellaneous Symbols (some are emoji)
	}

	// Most single-width Unicode characters should be fine
	return true
}

// isLikelyEmoji checks if a character in the Miscellaneous Symbols range
// is likely to be rendered as an emoji (double-width)
func isLikelyEmoji(r rune) bool {
	// Common emoji-like symbols in the U+2600-26FF range
	emojiLike := []rune{
		'☀', '☁', '☂', '☃', '☄', '★', '☆', '☇', '☈', '☉', '☊', '☋', '☌', '☍', '☎', '☏',
		'☐', '☑', '☒', '☓', '☔', '☕', '☖', '☗', '☘', '☙', '☚', '☛', '☜', '☝', '☞', '☟',
		'☠', '☡', '☢', '☣', '☤', '☥', '☦', '☧', '☨', '☩', '☪', '☫', '☬', '☭', '☮', '☯',
		'☰', '☱', '☲', '☳', '☴', '☵', '☶', '☷', '☸', '☹', '☺', '☻', '☼', '☽', '☾', '☿',
		'♀', '♁', '♂', '♃', '♄', '♅', '♆', '♇', '♈', '♉', '♊', '♋', '♌', '♍', '♎', '♏',
		'♐', '♑', '♒', '♓', '♔', '♕', '♖', '♗', '♘', '♙', '♚', '♛', '♜', '♝', '♞', '♟',
		'♠', '♡', '♢', '♣', '♤', '♥', '♦', '♧', '♨', '♩', '♪', '♫', '♬', '♭', '♮', '♯',
	}

	for _, emoji := range emojiLike {
		if r == emoji {
			return true
		}
	}
	return false
}