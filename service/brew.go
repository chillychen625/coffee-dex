package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"time"

	"github.com/google/uuid"
)

// BrewService handles business logic for brew operations
type BrewService struct {
	storage       storage.BrewStorage
	coffeeStorage storage.CoffeeStorage
}

// NewBrewService creates a new brew service
func NewBrewService(brewStorage storage.BrewStorage, coffeeStorage storage.CoffeeStorage) *BrewService {
	return &BrewService{
		storage:       brewStorage,
		coffeeStorage: coffeeStorage,
	}
}

// CreateBrew creates a new brew entry for a coffee
func (s *BrewService) CreateBrew(brew models.Brew) (models.Brew, error) {
	// Verify coffee exists
	_, err := s.coffeeStorage.GetByID(brew.CoffeeID)
	if err != nil {
		return models.Brew{}, fmt.Errorf("coffee not found: %w", err)
	}

	brew.ID = uuid.New().String()
	brew.CreatedAt = time.Now()

	if err := brew.Validate(); err != nil {
		return models.Brew{}, err
	}

	if err := s.storage.Save(brew); err != nil {
		return models.Brew{}, err
	}

	return brew, nil
}

// GetBrew retrieves a brew by ID
func (s *BrewService) GetBrew(id string) (models.Brew, error) {
	return s.storage.GetByID(id)
}

// GetBrewsForCoffee retrieves all brews for a specific coffee
func (s *BrewService) GetBrewsForCoffee(coffeeID string) ([]models.Brew, error) {
	return s.storage.GetByCoffeeID(coffeeID)
}

// GetBrewCount returns the number of brews for a specific coffee
func (s *BrewService) GetBrewCount(coffeeID string) (int, error) {
	return s.storage.GetBrewCount(coffeeID)
}

// ListBrews retrieves all brews
func (s *BrewService) ListBrews() ([]models.Brew, error) {
	return s.storage.GetAll()
}

// GetRecentBrews retrieves the most recent brews
func (s *BrewService) GetRecentBrews(limit int) ([]models.Brew, error) {
	return s.storage.GetRecent(limit)
}

// DeleteBrew removes a brew entry
func (s *BrewService) DeleteBrew(id string) error {
	return s.storage.Delete(id)
}

// CanGeneratePokemon checks if a coffee has enough brews to generate a Pokemon
func (s *BrewService) CanGeneratePokemon(coffeeID string) (bool, error) {
	count, err := s.storage.GetBrewCount(coffeeID)
	if err != nil {
		return false, err
	}
	return count >= models.RequiredBrewsForPokemon, nil
}

// GetBrewProgress returns the brew progress for Pokemon generation
func (s *BrewService) GetBrewProgress(coffeeID string, hasPokemon bool) (models.BrewProgress, error) {
	count, err := s.storage.GetBrewCount(coffeeID)
	if err != nil {
		return models.BrewProgress{}, err
	}

	return models.BrewProgress{
		Count:              count,
		Required:           models.RequiredBrewsForPokemon,
		CanGeneratePokemon: count >= models.RequiredBrewsForPokemon && !hasPokemon,
		HasPokemon:         hasPokemon,
	}, nil
}

// GetAggregatedData computes aggregated data from all brews of a coffee
func (s *BrewService) GetAggregatedData(coffeeID string) (*models.AggregatedBrewData, error) {
	brews, err := s.storage.GetByCoffeeID(coffeeID)
	if err != nil {
		return nil, err
	}

	if len(brews) == 0 {
		return nil, fmt.Errorf("no brews found for coffee %s", coffeeID)
	}

	return &models.AggregatedBrewData{
		CoffeeID:        coffeeID,
		BrewCount:       len(brews),
		AverageRating:   models.AverageRating(brews),
		AverageTraits:   models.AverageTraits(brews),
		CombinedNotes:   models.CombineNotes(brews),
		MostUsedDripper: models.MostUsedDripper(brews),
	}, nil
}
