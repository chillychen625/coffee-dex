package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PokemonService handles business logic for Pokemon operations
type PokemonService struct {
	storage      storage.PokemonStorage
	coffeeService *CoffeeService
	llmService   *LLMService
	mapper       *PokemonMapper
}

// NewPokemonService creates a new Pokemon service
func NewPokemonService(
	pokemonStorage storage.PokemonStorage,
	coffeeService *CoffeeService,
	llmService *LLMService,
) *PokemonService {
	return &PokemonService{
		storage:      pokemonStorage,
		coffeeService: coffeeService,
		llmService:   llmService,
		mapper:       NewPokemonMapper(),
	}
}

// MapCoffeeToPokemon maps a coffee to a Pokemon using enhanced type system + LLM
func (s *PokemonService) MapCoffeeToPokemon(coffee models.Coffee) (*models.CoffeePokemon, error) {
	// 1. Use enhanced mapper to determine Pokemon types
	primaryType, secondaryType, typeScores := s.mapper.CalculatePokemonTypes(coffee)
	log.Printf("Coffee types: primary=%s, secondary=%s, scores=%v", primaryType, secondaryType, typeScores)
	
	// 2. Get candidate Pokemon based on types
	candidates := s.getTypedCandidates(primaryType, secondaryType)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no Pokemon candidates found for types %s/%s", primaryType, secondaryType)
	}

	// 3. Use LLM to pick the best Pokemon from candidates with type context
	var selectedPokemon *models.Pokemon
	var confidence float64
	var description string
	var traitMapping []models.TraitMapping

	if s.llmService != nil {
		// Give LLM the type context to help it choose
		llmResponse, err := s.llmService.MapCoffeeToPokemon(coffee, candidates)
		if err != nil {
			log.Printf("LLM mapping failed, using best type match: %v", err)
			selectedPokemon, confidence, description, traitMapping = s.getBestTypeMatch(coffee, candidates, primaryType, typeScores[primaryType])
		} else {
			// Find the Pokemon by name from LLM response
			for _, candidate := range candidates {
				if strings.ToLower(candidate.Name) == strings.ToLower(llmResponse.SelectedPokemon) {
					selectedPokemon = &candidate
					break
				}
			}
			if selectedPokemon == nil {
				log.Printf("LLM selected unknown Pokemon: %s, using best type match", llmResponse.SelectedPokemon)
				selectedPokemon, confidence, description, traitMapping = s.getBestTypeMatch(coffee, candidates, primaryType, typeScores[primaryType])
			} else {
				confidence = llmResponse.Confidence
				description = llmResponse.Description
				traitMapping = llmResponse.TraitMapping
			}
		}
	} else {
		selectedPokemon, confidence, description, traitMapping = s.getBestTypeMatch(coffee, candidates, primaryType, typeScores[primaryType])
	}

	// 4. Ensure uniqueness
	finalPokemon, err := s.ensureUniquePokemon(coffee.ID, *selectedPokemon)
	if err != nil {
		return nil, fmt.Errorf("no unique Pokemon available: %w", err)
	}

	// 5. Create mapping with type info
	typeDescription := s.mapper.GetTypeDescription(primaryType, coffee)
	if secondaryType != "" {
		typeDescription += fmt.Sprintf(" and %s", s.mapper.GetTypeDescription(secondaryType, coffee))
	}
	
	mapping := &models.CoffeePokemon{
		ID:                uuid.New().String(),
		CoffeeID:          coffee.ID,
		PokemonID:         finalPokemon.ID,
		PokemonName:       finalPokemon.Name,
		Nickname:          "",
		Level:             s.calculateLevel(coffee.Rating),
		MappingConfidence: confidence,
		LLMDescription:    fmt.Sprintf("%s\n\nType Analysis: %s", description, typeDescription),
		TraitMapping:      traitMapping,
		CreatedAt:         time.Now(),
	}

	if err := s.storage.CreateCoffeePokemon(*mapping); err != nil {
		return nil, fmt.Errorf("failed to create Pokemon mapping: %w", err)
	}
	return mapping, nil
}

// getTypedCandidates gets Pokemon candidates based on calculated types
func (s *PokemonService) getTypedCandidates(primaryType, secondaryType string) []models.Pokemon {
	candidates := make([]models.Pokemon, 0)
	
	// Get Pokemon of primary type
	primary, err := s.storage.GetPokemonByType(primaryType)
	if err != nil {
		log.Printf("Failed to get Pokemon by type %s: %v", primaryType, err)
	} else {
		candidates = append(candidates, primary...)
	}
	
	// Get Pokemon of secondary type if exists
	if secondaryType != "" {
		secondary, err := s.storage.GetPokemonByType(secondaryType)
		if err != nil {
			log.Printf("Failed to get Pokemon by type %s: %v", secondaryType, err)
		} else {
			candidates = append(candidates, secondary...)
		}
	}
	
	// If no matches, get some normal types
	if len(candidates) == 0 {
		normal, err := s.storage.GetPokemonByType("Normal")
		if err == nil {
			candidates = append(candidates, normal...)
		}
	}
	
	// Limit to 10 candidates for LLM
	if len(candidates) > 10 {
		candidates = candidates[:10]
	}
	
	return candidates
}

// getBestTypeMatch selects best Pokemon from candidates based on type score
func (s *PokemonService) getBestTypeMatch(coffee models.Coffee, candidates []models.Pokemon, primaryType string, typeScore float64) (*models.Pokemon, float64, string, []models.TraitMapping) {
	if len(candidates) == 0 {
		// Fallback to a basic Pokemon
		return &models.Pokemon{
			ID:          1,
			Name:        "Bulbasaur",
			Type:        "Grass/Poison",
			Description: "A basic Pokemon for coffee mapping",
		}, 0.5, "Fallback mapping - no candidates available", []models.TraitMapping{}
	}
	
	// Select first candidate from type matches
	selected := candidates[0]
	confidence := typeScore * 0.9 // Type score as base confidence
	description := fmt.Sprintf("Type-based mapping: %s (%s-type) matches coffee's %s characteristics with %.0f%% confidence",
		selected.Name, selected.Type, primaryType, confidence*100)
	
	// Build trait mapping based on dominant traits
	traitMapping := s.buildTraitMapping(coffee.TastingTraits, selected)
	
	return &selected, confidence, description, traitMapping
}

// buildTraitMapping creates trait mappings based on coffee characteristics
func (s *PokemonService) buildTraitMapping(traits models.TastingTraits, pokemon models.Pokemon) []models.TraitMapping {
	mappings := []models.TraitMapping{}
	
	if traits.Sweetness >= 7 {
		mappings = append(mappings, models.TraitMapping{
			Trait:       "sweetness",
			PokemonStat: "HP",
			Reasoning:   "High sweetness provides sustained energy like HP",
		})
	}
	if traits.Bitterness >= 7 {
		mappings = append(mappings, models.TraitMapping{
			Trait:       "bitterness",
			PokemonStat: "Attack",
			Reasoning:   "Bold bitterness represents attacking flavors",
		})
	}
	if traits.Body >= 7 {
		mappings = append(mappings, models.TraitMapping{
			Trait:       "body",
			PokemonStat: "Defense",
			Reasoning:   "Full body provides defensive structure",
		})
	}
	if traits.CitrusFruitsIntensity >= 7 {
		mappings = append(mappings, models.TraitMapping{
			Trait:       "citrus",
			PokemonStat: "Speed",
			Reasoning:   "Bright citrus notes provide quick, energetic speed",
		})
	}
	if traits.AromaticIntensity >= 7 {
		mappings = append(mappings, models.TraitMapping{
			Trait:       "aroma",
			PokemonStat: "Special",
			Reasoning:   "Complex aroma represents special characteristics",
		})
	}
	
	return mappings
}


// ensureUniquePokemon ensures each Pokemon is unique
func (s *PokemonService) ensureUniquePokemon(coffeeID string, pokemon models.Pokemon) (*models.Pokemon, error) {
	used, err := s.storage.IsPokemonUsed(pokemon.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check Pokemon usage: %w", err)
	}

	if !used {
		return &pokemon, nil
	}

	// Find alternative Pokemon with similar characteristics
	alternatives, err := s.storage.GetPokemonByType(pokemon.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get alternative Pokemon: %w", err)
	}

	for _, alt := range alternatives {
		altUsed, err := s.storage.IsPokemonUsed(alt.ID)
		if err != nil {
			continue
		}
		if !altUsed {
			return &alt, nil
		}
	}

	// If no alternatives, return original (will fail on database constraint)
	return &pokemon, fmt.Errorf("Pokemon %s already used and no alternatives available", pokemon.Name)
}

// calculateLevel calculates Pokemon level based on coffee rating
func (s *PokemonService) calculateLevel(rating int) int {
	// Level 1-50 based on rating 0-10
	return rating * 5
}

// calculateTraitVariance calculates variance in coffee traits
func (s *PokemonService) calculateTraitVariance(traits models.TastingTraits) int {
	traitValues := []int{
		traits.BerryIntensity, traits.StonefruitIntensity, traits.RoastIntensity,
		traits.CitrusFruitsIntensity, traits.Bitterness, traits.Florality,
		traits.Spice, traits.Sweetness, traits.AromaticIntensity,
		traits.Savory, traits.Body, traits.Cleanliness,
	}

	// Calculate mean
	sum := 0
	for _, val := range traitValues {
		sum += val
	}
	mean := sum / len(traitValues)

	// Calculate variance
	variance := 0
	for _, val := range traitValues {
		diff := val - mean
		variance += diff * diff
	}

	return variance / len(traitValues)
}

// GetCoffeePokemon gets Pokemon mapping for a specific coffee
func (s *PokemonService) GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error) {
	return s.storage.GetCoffeePokemon(coffeeID)
}

// GetAllCoffeePokemon gets all coffee-Pokemon mappings
func (s *PokemonService) GetAllCoffeePokemon() ([]models.CoffeePokemon, error) {
	return s.storage.GetAllCoffeePokemon()
}

// UpdateNickname updates Pokemon nickname
func (s *PokemonService) UpdateNickname(coffeeID, nickname string) error {
	return s.storage.UpdateCoffeePokemonNickname(coffeeID, nickname)
}

// InitializePokemonData checks if Pokemon data exists in database
func (s *PokemonService) InitializePokemonData() error {
	// Check if Pokemon data already exists
	existing, err := s.storage.GetAllPokemon()
	if err == nil && len(existing) > 0 {
		log.Printf("Pokemon data already loaded: %d Pokemon in database", len(existing))
		return nil
	}

	// Pokemon data should be loaded via sql/pokemon_gen1_data.sql
	log.Println("Warning: No Pokemon data found. Please run sql/pokemon_gen1_data.sql to initialize the database")
	
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}