package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PokemonService handles business logic for Pokemon operations
type PokemonService struct {
	storage       storage.PokemonStorage
	coffeeService *CoffeeService
	brewService   *BrewService
	claudeService *ClaudeService
	mapper        *PokemonMapper
}

// NewPokemonService creates a new Pokemon service
func NewPokemonService(
	pokemonStorage storage.PokemonStorage,
	coffeeService *CoffeeService,
	brewService *BrewService,
	claudeService *ClaudeService,
) *PokemonService {
	return &PokemonService{
		storage:       pokemonStorage,
		coffeeService: coffeeService,
		brewService:   brewService,
		claudeService: claudeService,
		mapper:        NewPokemonMapper(),
	}
}

// MapCoffeeToPokemon maps a coffee to a Pokemon using aggregated brew data
func (s *PokemonService) MapCoffeeToPokemon(coffeeID string) (*models.CoffeePokemon, error) {
	// 1. Check if Pokemon already exists
	existing, err := s.storage.GetCoffeePokemon(coffeeID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("Pokemon already generated for this coffee - regeneration not allowed")
	}

	// 2. Get coffee info
	coffee, err := s.coffeeService.GetCoffee(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coffee: %w", err)
	}

	// 3. Validate brew count
	canGenerate, err := s.brewService.CanGeneratePokemon(coffeeID, coffee.IsFinished)
	if err != nil {
		return nil, fmt.Errorf("failed to check brew count: %w", err)
	}
	if !canGenerate {
		count, _ := s.brewService.GetBrewCount(coffeeID)
		return nil, fmt.Errorf("need %d more brews to generate Pokemon (current: %d, required: %d)",
			models.RequiredBrewsForPokemon-count, count, models.RequiredBrewsForPokemon)
	}

	// 4. Aggregate brew data
	aggregated, err := s.brewService.GetAggregatedData(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated brew data: %w", err)
	}

	// 5. Get type suggestions from mapper (hints for Claude, not constraints)
	_, _, typeScores := s.mapper.CalculatePokemonTypesFromTraits(
		aggregated.AverageTraits,
		coffee.ProcessingMethod,
		coffee.RoastLevel,
		aggregated.CombinedNotes,
	)

	// 6. Get all unassigned Pokemon (uniqueness handled upstream)
	availablePokemon, err := s.getAvailablePokemon()
	if err != nil {
		return nil, fmt.Errorf("failed to get available Pokemon: %w", err)
	}
	if len(availablePokemon) == 0 {
		return nil, fmt.Errorf("all 151 Gen 1 Pokemon are already assigned")
	}

	log.Printf("Pokemon generation: coffee=%s, %d brews, %d available Pokemon", coffee.Name, aggregated.BrewCount, len(availablePokemon))

	// 7. Ask Claude to pick a Pokemon
	var selectedPokemon *models.Pokemon
	var confidence float64
	var description string

	if s.claudeService != nil {
		claudeResp, err := s.claudeService.SelectPokemon(coffee, aggregated.AverageTraits, aggregated.CombinedNotes, typeScores, availablePokemon)
		if err != nil {
			log.Printf("Claude selection failed, using type-based fallback: %v", err)
			selectedPokemon, confidence, description = s.fallbackSelect(aggregated.AverageTraits, typeScores, availablePokemon)
		} else {
			// Find the Pokemon by ID from Claude's response
			for i := range availablePokemon {
				if availablePokemon[i].ID == claudeResp.PokemonID {
					selectedPokemon = &availablePokemon[i]
					break
				}
			}
			if selectedPokemon == nil {
				log.Printf("Claude returned unknown Pokemon ID %d, using fallback", claudeResp.PokemonID)
				selectedPokemon, confidence, description = s.fallbackSelect(aggregated.AverageTraits, typeScores, availablePokemon)
			} else {
				confidence = claudeResp.Confidence
				description = claudeResp.Description
			}
		}
	} else {
		log.Printf("No Claude service available, using type-based fallback")
		selectedPokemon, confidence, description = s.fallbackSelect(aggregated.AverageTraits, typeScores, availablePokemon)
	}

	// 8. Build and save the mapping
	level := calculateLevel(int(aggregated.AverageRating + 0.5))

	mapping := &models.CoffeePokemon{
		ID:                uuid.New().String(),
		CoffeeID:          coffee.ID,
		PokemonID:         selectedPokemon.ID,
		PokemonName:       selectedPokemon.Name,
		PokemonType:       selectedPokemon.Type,
		Nickname:          "",
		Level:             level,
		MappingConfidence: confidence,
		LLMDescription:    description,
		TraitMapping:      []models.TraitMapping{},
		CreatedAt:         time.Now(),
	}

	if err := s.storage.CreateCoffeePokemon(*mapping); err != nil {
		return nil, fmt.Errorf("failed to create Pokemon mapping: %w", err)
	}

	log.Printf("Pokemon assigned: %s (#%d) for coffee %s (confidence: %.0f%%)", selectedPokemon.Name, selectedPokemon.ID, coffee.Name, confidence*100)
	return mapping, nil
}

// getAvailablePokemon returns all Pokemon that haven't been assigned yet
func (s *PokemonService) getAvailablePokemon() ([]models.Pokemon, error) {
	allPokemon, err := s.storage.GetAllPokemon()
	if err != nil {
		return nil, fmt.Errorf("failed to get all Pokemon: %w", err)
	}

	available := make([]models.Pokemon, 0, len(allPokemon))
	for _, p := range allPokemon {
		used, err := s.storage.IsPokemonUsed(p.ID)
		if err != nil {
			continue
		}
		if !used {
			available = append(available, p)
		}
	}

	return available, nil
}

// fallbackSelect picks a Pokemon based on type scores when Claude isn't available
func (s *PokemonService) fallbackSelect(traits models.TastingTraits, typeScores map[string]float64, available []models.Pokemon) (*models.Pokemon, float64, string) {
	// Find the best type match
	bestType := "Normal"
	bestScore := 0.0
	for t, score := range typeScores {
		if score > bestScore {
			bestType = t
			bestScore = score
		}
	}

	// Collect all available Pokemon of the best type, then pick one at random
	matches := make([]int, 0)
	for i := range available {
		if containsType(available[i].Type, bestType) {
			matches = append(matches, i)
		}
	}
	if len(matches) > 0 {
		pick := matches[rand.Intn(len(matches))]
		desc := fmt.Sprintf("A %s-type Pokemon matched to this coffee's flavor profile. %s's characteristics align with the %s notes in your brews.",
			available[pick].Type, available[pick].Name, bestType)
		return &available[pick], bestScore * 0.8, desc
	}

	// No type match — pick a random available Pokemon (avoids pokedex creep)
	if len(available) > 0 {
		pick := rand.Intn(len(available))
		return &available[pick], 0.5, fmt.Sprintf("%s was matched to this coffee's unique flavor profile.", available[pick].Name)
	}

	// Should never reach here since we check upstream
	return &models.Pokemon{ID: 1, Name: "Bulbasaur", Type: "Grass/Poison"}, 0.3, "A trusty companion for your coffee journey."
}

// containsType checks if a Pokemon's type string contains a specific type
func containsType(pokemonType, targetType string) bool {
	for _, t := range splitTypes(pokemonType) {
		if t == targetType {
			return true
		}
	}
	return false
}

// splitTypes splits a type string like "Fire/Water" into individual types
func splitTypes(typeStr string) []string {
	parts := make([]string, 0, 2)
	for _, p := range split(typeStr, "/") {
		trimmed := trim(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// split is a simple string split helper
func split(s, sep string) []string {
	result := make([]string, 0)
	for {
		i := indexOf(s, sep)
		if i < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:i])
		s = s[i+len(sep):]
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// calculateLevel calculates Pokemon level from coffee rating
func calculateLevel(rating int) int {
	if rating < 0 {
		rating = 0
	}
	if rating > 10 {
		rating = 10
	}
	return rating * 5
}

// GetCoffeePokemon gets Pokemon mapping for a specific coffee
func (s *PokemonService) GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error) {
	return s.storage.GetCoffeePokemon(coffeeID)
}

// HasPokemon checks if a coffee has a Pokemon mapping
func (s *PokemonService) HasPokemon(coffeeID string) bool {
	mapping, err := s.storage.GetCoffeePokemon(coffeeID)
	return err == nil && mapping != nil
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
	existing, err := s.storage.GetAllPokemon()
	if err == nil && len(existing) > 0 {
		log.Printf("Pokemon data already loaded: %d Pokemon in database", len(existing))
		return nil
	}

	log.Println("Warning: No Pokemon data found. Please run sql/pokemon_gen1_data.sql to initialize the database")
	return nil
}
