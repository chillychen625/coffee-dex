package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PokemonService handles business logic for Pokemon operations
type PokemonService struct {
	storage       storage.PokemonStorage
	coffeeService *CoffeeService
	brewService   *BrewService
	mapper        *PokemonMapper
}

type PokemonCandidates struct {
	CoffeeID      string
	CoffeeName    string
	PrimaryType   string
	SecondaryType string
	TypeScores    map[string]float64
	Curated       []models.Pokemon
	All           []models.Pokemon
}

// NewPokemonService creates a new Pokemon service
func NewPokemonService(
	pokemonStorage storage.PokemonStorage,
	coffeeService *CoffeeService,
	brewService *BrewService,
) *PokemonService {
	return &PokemonService{
		storage:       pokemonStorage,
		coffeeService: coffeeService,
		brewService:   brewService,
		mapper:        NewPokemonMapper(),
	}
}

// GetPokemonCandidates computes the coffee's trait-derived types and returns
// both the curated (type-matching, best first) and full lists of unassigned
// Pokemon. The picker decides which list to display; an empty curated list is
// legitimate (all type matches taken) and should fall back to All there.
func (s *PokemonService) GetPokemonCandidates(coffeeID string) (*PokemonCandidates, error) {
	existing, err := s.storage.GetCoffeePokemon(coffeeID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("Pokemon already generated for this coffee - regeneration not allowed")
	}

	coffee, err := s.coffeeService.GetCoffee(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coffee: %w", err)
	}

	canGenerate, err := s.brewService.CanGeneratePokemon(coffeeID, coffee.IsFinished)
	if err != nil {
		return nil, fmt.Errorf("failed to check brew count: %w", err)
	}
	if !canGenerate {
		count, _ := s.brewService.GetBrewCount(coffeeID)
		return nil, fmt.Errorf("need %d more brews to generate Pokemon (current: %d, required: %d)",
			models.RequiredBrewsForPokemon-count, count, models.RequiredBrewsForPokemon)
	}

	aggregated, err := s.brewService.GetAggregatedData(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated brew data: %w", err)
	}

	primaryType, secondaryType, typeScores := s.mapper.CalculatePokemonTypesFromTraits(
		aggregated.AverageTraits,
		coffee.ProcessingMethod,
		coffee.RoastLevel,
		aggregated.CombinedNotes,
	)

	availablePokemon, err := s.getAvailablePokemon()
	if err != nil {
		return nil, fmt.Errorf("failed to get available Pokemon: %w", err)
	}
	if len(availablePokemon) == 0 {
		return nil, fmt.Errorf("all 151 Gen 1 Pokemon are already assigned")
	}

	curated := make([]models.Pokemon, 0, len(availablePokemon))
	for _, pkmn := range availablePokemon {
		if typeMatches(pkmn.Type, primaryType) || typeMatches(pkmn.Type, secondaryType) {
			curated = append(curated, pkmn)
		}
	}

	// Best matches first; the ID tiebreaker keeps the order stable.
	sort.SliceStable(curated, func(i, j int) bool {
		si, sj := matchScore(curated[i], typeScores), matchScore(curated[j], typeScores)
		if si != sj {
			return si > sj
		}
		return curated[i].ID < curated[j].ID
	})

	return &PokemonCandidates{
		CoffeeID:      coffeeID,
		CoffeeName:    coffee.Name,
		PrimaryType:   primaryType,
		SecondaryType: secondaryType,
		TypeScores:    typeScores,
		Curated:       curated,
		All:           availablePokemon,
	}, nil
}

// Match returns the best trait-derived type score for p (0 if none of its
// types scored), so the picker UI can display fit without duplicating the
// scoring rules.
func (c *PokemonCandidates) Match(p models.Pokemon) float64 {
	return matchScore(p, c.TypeScores)
}

// CreateManualMapping persists a user-chosen Pokemon and hand-written
// description for a coffee. Any unassigned Pokemon is allowed; the mapping
// confidence records how well the choice matched the coffee's trait-derived
// types, so browse-all picks honestly report a lower match.
func (s *PokemonService) CreateManualMapping(coffeeID string, pokemonID int, description string) (*models.CoffeePokemon, error) {
	existing, _ := s.storage.GetCoffeePokemon(coffeeID)
	if existing != nil {
		return nil, fmt.Errorf("pokemon already mapped")
	}

	coffee, err := s.coffeeService.GetCoffee(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coffee: %w", err)
	}

	// Re-check eligibility: brews may have changed since the picker opened.
	canGenerate, err := s.brewService.CanGeneratePokemon(coffeeID, coffee.IsFinished)
	if err != nil {
		return nil, fmt.Errorf("failed to check brew count: %w", err)
	}
	if !canGenerate {
		return nil, fmt.Errorf("this coffee no longer qualifies for a Pokemon")
	}

	aggregated, err := s.brewService.GetAggregatedData(coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated brew data: %w", err)
	}

	// The chosen Pokemon must still be unassigned; otherwise the DB's
	// UNIQUE(pokemon_id) constraint would surface as a raw SQL error.
	available, err := s.getAvailablePokemon()
	if err != nil {
		return nil, fmt.Errorf("failed to get available Pokemon: %w", err)
	}
	var pkmn *models.Pokemon
	for i := range available {
		if available[i].ID == pokemonID {
			pkmn = &available[i]
			break
		}
	}
	if pkmn == nil {
		return nil, fmt.Errorf("that Pokemon is already assigned to another coffee")
	}

	_, _, typeScores := s.mapper.CalculatePokemonTypesFromTraits(
		aggregated.AverageTraits,
		coffee.ProcessingMethod,
		coffee.RoastLevel,
		aggregated.CombinedNotes,
	)

	level := calculateLevel(int(aggregated.AverageRating + 0.5))
	// Clamped: out-of-type picks floor at 35%, strong matches stay below 100%
	// (per-rule score normalization is only roughly comparable across types).
	confidence := min(max(matchScore(*pkmn, typeScores), 0.35), 0.95)

	mapping := &models.CoffeePokemon{
		ID:                uuid.New().String(),
		CoffeeID:          coffeeID,
		PokemonID:         pokemonID,
		PokemonName:       pkmn.Name,
		PokemonType:       pkmn.Type,
		Nickname:          "",
		Level:             level,
		MappingConfidence: confidence,
		Description:       description,
		TraitMapping:      []models.TraitMapping{},
		CreatedAt:         time.Now(),
	}

	if err := s.storage.CreateCoffeePokemon(*mapping); err != nil {
		return nil, fmt.Errorf("failed to create Pokemon mapping: %w", err)
	}

	log.Printf("Pokemon assigned: %s (#%d) for coffee %s (confidence: %.0f%%)", pkmn.Name, pkmn.ID, coffee.Name, confidence*100)
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

func typeMatches(pokemonType, target string) bool {
	if target == "" {
		return false
	}
	for _, t := range strings.Split(pokemonType, "/") {
		if strings.EqualFold(strings.TrimSpace(t), target) {
			return true
		}
	}
	return false
}

// matchScore returns the best trait-derived type score for a Pokemon
// across its types (0 if none of them scored).
func matchScore(p models.Pokemon, typeScores map[string]float64) float64 {
	best := 0.0
	for _, t := range strings.Split(p.Type, "/") {
		key := strings.ToLower(strings.TrimSpace(t))
		if v, ok := typeScores[key]; ok && v > best {
			best = v
		}
	}
	return best
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
