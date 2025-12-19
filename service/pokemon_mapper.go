package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"go-coffee-log/models"
)

// PokemonType represents a Pokemon type with its characteristics
type PokemonType struct {
	Name        string
	Color       string
	Description string
}

// TypeScore represents how well a coffee matches a Pokemon type
type TypeScore struct {
	Type  string
	Score float64
}

// PokemonMapper handles the sophisticated mapping of coffee to Pokemon types
type PokemonMapper struct {
	typeRules map[string]TypeMappingRule
}

// TypeMappingRule defines how a Pokemon type is determined
type TypeMappingRule struct {
	Type              string
	PrimaryTraits     []TraitWeight
	SecondaryTraits   []TraitWeight
	KeywordMatches    []string
	ProcessingBonus   map[string]float64
	RoastLevelBonus   map[string]float64
	MinimumThreshold  float64
}

// TraitWeight defines a trait and its weight in type determination
type TraitWeight struct {
	Trait  string
	Weight float64
	Min    int // Minimum value needed to count
	Max    int // Maximum value for optimal score
}

// NewPokemonMapper creates a new Pokemon mapper with all type rules
func NewPokemonMapper() *PokemonMapper {
	mapper := &PokemonMapper{
		typeRules: make(map[string]TypeMappingRule),
	}
	mapper.initializeTypeRules()
	return mapper
}

// initializeTypeRules sets up the sophisticated type mapping rules
func (pm *PokemonMapper) initializeTypeRules() {
	// Normal: Generic Coffee Taste - balanced, no strong characteristics
	pm.typeRules["normal"] = TypeMappingRule{
		Type: "normal",
		PrimaryTraits: []TraitWeight{
			{Trait: "cleanliness", Weight: 2.0, Min: 6, Max: 9},
			{Trait: "body", Weight: 1.5, Min: 4, Max: 7},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "sweetness", Weight: 1.0, Min: 4, Max: 6},
			{Trait: "bitterness", Weight: 1.0, Min: 3, Max: 6},
		},
		ProcessingBonus: map[string]float64{"washed": 1.3},
		RoastLevelBonus: map[string]float64{"medium": 1.4, "light medium": 1.2},
		MinimumThreshold: 0.4,
	}

	// Fire: Roasty or Savory OR Peppery
	pm.typeRules["fire"] = TypeMappingRule{
		Type: "fire",
		PrimaryTraits: []TraitWeight{
			{Trait: "roast_intensity", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "savory", Weight: 2.0, Min: 6, Max: 10},
			{Trait: "spice", Weight: 2.2, Min: 7, Max: 10}, // Peppery
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "bitterness", Weight: 1.2, Min: 6, Max: 9},
			{Trait: "body", Weight: 1.0, Min: 7, Max: 10},
		},
		KeywordMatches: []string{"pepper", "roast", "smoke", "char", "burnt", "toast", "caramel"},
		RoastLevelBonus: map[string]float64{"dark": 1.8, "medium dark": 1.5},
		MinimumThreshold: 0.6,
	}

	// Water: Seaweed/Fishy (rare in coffee)
	pm.typeRules["water"] = TypeMappingRule{
		Type: "water",
		PrimaryTraits: []TraitWeight{
			{Trait: "cleanliness", Weight: 2.0, Min: 8, Max: 10},
			{Trait: "body", Weight: 1.5, Min: 2, Max: 5}, // Light body
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "sweetness", Weight: 1.0, Min: 3, Max: 6},
		},
		KeywordMatches: []string{"water", "clean", "crisp", "mineral", "seaweed", "ocean"},
		ProcessingBonus: map[string]float64{"washed": 1.5},
		MinimumThreshold: 0.5,
	}

	// Grass: Grass/Vegetal/Floral
	pm.typeRules["grass"] = TypeMappingRule{
		Type: "grass",
		PrimaryTraits: []TraitWeight{
			{Trait: "florality", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "aromatic_intensity", Weight: 2.0, Min: 6, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "cleanliness", Weight: 1.3, Min: 6, Max: 9},
			{Trait: "sweetness", Weight: 1.0, Min: 5, Max: 8},
		},
		KeywordMatches: []string{"floral", "jasmine", "rose", "grass", "vegetal", "green", "herbal", "tea"},
		ProcessingBonus: map[string]float64{"washed": 1.3, "honey": 1.2},
		RoastLevelBonus: map[string]float64{"light": 1.5, "light medium": 1.3},
		MinimumThreshold: 0.55,
	}

	// Electric: Sharp Acidity
	pm.typeRules["electric"] = TypeMappingRule{
		Type: "electric",
		PrimaryTraits: []TraitWeight{
			{Trait: "citrus_fruits_intensity", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "aromatic_intensity", Weight: 2.0, Min: 7, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "cleanliness", Weight: 1.5, Min: 7, Max: 10},
			{Trait: "body", Weight: -1.0, Min: 2, Max: 5}, // Negative weight for light body
		},
		KeywordMatches: []string{"citrus", "lemon", "lime", "orange", "grapefruit", "bright", "zesty", "tangy", "acidic"},
		ProcessingBonus: map[string]float64{"washed": 1.4},
		RoastLevelBonus: map[string]float64{"light": 1.6, "light medium": 1.3},
		MinimumThreshold: 0.6,
	}

	// Ice: Minty
	pm.typeRules["ice"] = TypeMappingRule{
		Type: "ice",
		PrimaryTraits: []TraitWeight{
			{Trait: "cleanliness", Weight: 2.5, Min: 8, Max: 10},
			{Trait: "aromatic_intensity", Weight: 2.0, Min: 7, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "florality", Weight: 1.5, Min: 6, Max: 9},
		},
		KeywordMatches: []string{"mint", "menthol", "eucalyptus", "cooling", "fresh", "crisp"},
		ProcessingBonus: map[string]float64{"washed": 1.4},
		MinimumThreshold: 0.65,
	}

	// Poison: Spice OR Funky
	pm.typeRules["poison"] = TypeMappingRule{
		Type: "poison",
		PrimaryTraits: []TraitWeight{
			{Trait: "spice", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "savory", Weight: 2.0, Min: 7, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "aromatic_intensity", Weight: 1.5, Min: 7, Max: 10},
			{Trait: "bitterness", Weight: 1.0, Min: 5, Max: 8},
		},
		KeywordMatches: []string{"spice", "funky", "ferment", "wild", "unusual", "complex", "intense"},
		ProcessingBonus: map[string]float64{"natural": 1.5, "experimental": 1.8, "coferment": 1.7},
		MinimumThreshold: 0.6,
	}

	// Ground: Earthy/Grain
	pm.typeRules["ground"] = TypeMappingRule{
		Type: "ground",
		PrimaryTraits: []TraitWeight{
			{Trait: "body", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "savory", Weight: 2.0, Min: 6, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "roast_intensity", Weight: 1.5, Min: 5, Max: 8},
			{Trait: "bitterness", Weight: 1.0, Min: 4, Max: 7},
		},
		KeywordMatches: []string{"earth", "soil", "grain", "wheat", "cereal", "nutty", "almond", "hazelnut"},
		ProcessingBonus: map[string]float64{"natural": 1.3, "honey": 1.2},
		MinimumThreshold: 0.55,
	}

	// Rock: Stonefruits
	pm.typeRules["rock"] = TypeMappingRule{
		Type: "rock",
		PrimaryTraits: []TraitWeight{
			{Trait: "stonefruit_intensity", Weight: 3.0, Min: 7, Max: 10},
			{Trait: "sweetness", Weight: 2.0, Min: 6, Max: 9},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "body", Weight: 1.5, Min: 6, Max: 9},
			{Trait: "aromatic_intensity", Weight: 1.0, Min: 5, Max: 8},
		},
		KeywordMatches: []string{"peach", "apricot", "plum", "cherry", "nectarine", "stonefruit"},
		ProcessingBonus: map[string]float64{"natural": 1.4, "honey": 1.3},
		MinimumThreshold: 0.6,
	}

	// Dark: Roasty (alternative to Fire, less spicy)
	pm.typeRules["dark"] = TypeMappingRule{
		Type: "dark",
		PrimaryTraits: []TraitWeight{
			{Trait: "roast_intensity", Weight: 2.5, Min: 7, Max: 10},
			{Trait: "bitterness", Weight: 2.0, Min: 6, Max: 9},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "body", Weight: 1.5, Min: 7, Max: 10},
			{Trait: "sweetness", Weight: -1.0, Min: 2, Max: 5}, // Lower sweetness
		},
		KeywordMatches: []string{"dark", "chocolate", "cocoa", "roast", "bold", "intense"},
		RoastLevelBonus: map[string]float64{"dark": 2.0, "medium dark": 1.6},
		MinimumThreshold: 0.6,
	}

	// Fairy: Sugary Sweets
	pm.typeRules["fairy"] = TypeMappingRule{
		Type: "fairy",
		PrimaryTraits: []TraitWeight{
			{Trait: "sweetness", Weight: 3.0, Min: 8, Max: 10},
			{Trait: "aromatic_intensity", Weight: 2.0, Min: 7, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "florality", Weight: 1.5, Min: 6, Max: 9},
			{Trait: "berry_intensity", Weight: 1.5, Min: 6, Max: 9},
		},
		KeywordMatches: []string{"sweet", "candy", "sugar", "honey", "vanilla", "caramel", "syrup", "dessert"},
		ProcessingBonus: map[string]float64{"natural": 1.4, "honey": 1.5},
		MinimumThreshold: 0.65,
	}

	// Psychic: Highly Specific Notes (complex, unusual combinations)
	pm.typeRules["psychic"] = TypeMappingRule{
		Type: "psychic",
		PrimaryTraits: []TraitWeight{
			{Trait: "aromatic_intensity", Weight: 2.5, Min: 8, Max: 10},
			{Trait: "cleanliness", Weight: 2.0, Min: 7, Max: 10},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "florality", Weight: 1.5, Min: 6, Max: 9},
			{Trait: "berry_intensity", Weight: 1.0, Min: 6, Max: 9},
		},
		ProcessingBonus: map[string]float64{"experimental": 1.8, "coferment": 1.6},
		MinimumThreshold: 0.7, // High threshold for "delusional" specificity
	}

	// Bug: Spice notes (just for vibes) - same as Poison but lower threshold
	pm.typeRules["bug"] = TypeMappingRule{
		Type: "bug",
		PrimaryTraits: []TraitWeight{
			{Trait: "spice", Weight: 2.0, Min: 5, Max: 9},
			{Trait: "aromatic_intensity", Weight: 1.5, Min: 5, Max: 9},
		},
		SecondaryTraits: []TraitWeight{
			{Trait: "body", Weight: 1.0, Min: 4, Max: 7},
		},
		KeywordMatches: []string{"spice", "cinnamon", "cardamom", "clove", "insect", "bug"},
		ProcessingBonus: map[string]float64{"natural": 1.2, "experimental": 1.3},
		MinimumThreshold: 0.45,
	}
}

// CalculatePokemonTypes determines primary and secondary types for a coffee
func (pm *PokemonMapper) CalculatePokemonTypes(coffee models.Coffee) (string, string, map[string]float64) {
	scores := make(map[string]float64)

	// Calculate score for each type
	for typeName, rule := range pm.typeRules {
		score := pm.calculateTypeScore(coffee, rule)
		scores[typeName] = score
	}

	// Sort types by score
	var typeScores []TypeScore
	for typeName, score := range scores {
		typeScores = append(typeScores, TypeScore{Type: typeName, Score: score})
	}
	sort.Slice(typeScores, func(i, j int) bool {
		return typeScores[i].Score > typeScores[j].Score
	})

	// Get primary and secondary types
	primaryType := "normal"
	secondaryType := ""

	if len(typeScores) > 0 && typeScores[0].Score >= pm.typeRules[typeScores[0].Type].MinimumThreshold {
		primaryType = typeScores[0].Type
	}

	if len(typeScores) > 1 && typeScores[1].Score >= pm.typeRules[typeScores[1].Type].MinimumThreshold*0.8 {
		secondaryType = typeScores[1].Type
	}

	return primaryType, secondaryType, scores
}

// calculateTypeScore calculates how well a coffee matches a type rule
func (pm *PokemonMapper) calculateTypeScore(coffee models.Coffee, rule TypeMappingRule) float64 {
	score := 0.0
	maxPossibleScore := 0.0

	// Calculate primary trait scores
	for _, tw := range rule.PrimaryTraits {
		traitValue := pm.getTraitValue(coffee.TastingTraits, tw.Trait)
		maxPossibleScore += tw.Weight * 10.0

		if traitValue >= tw.Min {
			// Scale score based on how close to optimal range
			normalizedValue := float64(traitValue)
			if normalizedValue > float64(tw.Max) {
				normalizedValue = float64(tw.Max)
			}
			contribution := (normalizedValue / 10.0) * tw.Weight * 10.0
			score += contribution
		}
	}

	// Calculate secondary trait scores
	for _, tw := range rule.SecondaryTraits {
		traitValue := pm.getTraitValue(coffee.TastingTraits, tw.Trait)
		maxPossibleScore += tw.Weight * 10.0

		if traitValue >= tw.Min {
			normalizedValue := float64(traitValue)
			if normalizedValue > float64(tw.Max) {
				normalizedValue = float64(tw.Max)
			}
			contribution := (normalizedValue / 10.0) * tw.Weight * 10.0
			score += contribution
		}
	}

	// Keyword matching bonus
	if len(rule.KeywordMatches) > 0 {
		keywordScore := pm.calculateKeywordScore(coffee.TastingNotes, rule.KeywordMatches)
		score += keywordScore * 20.0 // Keyword matches are valuable
		maxPossibleScore += 20.0
	}

	// Processing method bonus
	if bonus, ok := rule.ProcessingBonus[coffee.ProcessingMethod]; ok {
		score *= bonus
	}

	// Roast level bonus
	if bonus, ok := rule.RoastLevelBonus[coffee.RoastLevel]; ok {
		score *= bonus
	}

	// Normalize score to 0-1 range
	if maxPossibleScore > 0 {
		return math.Min(score/maxPossibleScore, 1.0)
	}

	return 0.0
}

// getTraitValue extracts a trait value from TastingTraits
func (pm *PokemonMapper) getTraitValue(traits models.TastingTraits, traitName string) int {
	switch traitName {
	case "berry_intensity":
		return traits.BerryIntensity
	case "stonefruit_intensity":
		return traits.StonefruitIntensity
	case "roast_intensity":
		return traits.RoastIntensity
	case "citrus_fruits_intensity":
		return traits.CitrusFruitsIntensity
	case "bitterness":
		return traits.Bitterness
	case "florality":
		return traits.Florality
	case "spice":
		return traits.Spice
	case "sweetness":
		return traits.Sweetness
	case "aromatic_intensity":
		return traits.AromaticIntensity
	case "savory":
		return traits.Savory
	case "body":
		return traits.Body
	case "cleanliness":
		return traits.Cleanliness
	default:
		return 0
	}
}

// calculateKeywordScore checks tasting notes for keyword matches
func (pm *PokemonMapper) calculateKeywordScore(tastingNotes [5]string, keywords []string) float64 {
	matches := 0
	for _, note := range tastingNotes {
		if note == "" {
			continue
		}
		noteLower := strings.ToLower(note)
		for _, keyword := range keywords {
			if strings.Contains(noteLower, keyword) {
				matches++
				break // Count each note only once
			}
		}
	}
	return float64(matches) / 5.0 // Normalize to 0-1
}

// GetTypeDescription returns a description of why a type was chosen
func (pm *PokemonMapper) GetTypeDescription(typeName string, coffee models.Coffee) string {
	rule, ok := pm.typeRules[typeName]
	if !ok {
		return fmt.Sprintf("Unknown type: %s", typeName)
	}

	description := fmt.Sprintf("This coffee exhibits %s-type characteristics", typeName)
	
	// Add specific trait mentions
	highTraits := []string{}
	for _, tw := range rule.PrimaryTraits {
		value := pm.getTraitValue(coffee.TastingTraits, tw.Trait)
		if value >= tw.Min {
			highTraits = append(highTraits, tw.Trait)
		}
	}

	if len(highTraits) > 0 {
		description += " with strong " + strings.Join(highTraits, ", ")
	}

	return description
}