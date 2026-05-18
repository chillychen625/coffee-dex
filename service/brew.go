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
	// Verify coffee exists and get roast date
	coffee, err := s.coffeeStorage.GetByID(brew.CoffeeID)
	if err != nil {
		return models.Brew{}, fmt.Errorf("coffee not found: %w", err)
	}

	brew.ID = uuid.New().String()
	brew.CreatedAt = time.Now()

	// Calculate days off roast at brew time
	brew.DaysOffRoast = coffee.DaysOffRoastAt(brew.CreatedAt)

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
	brew, err := s.storage.GetByID(id)
	if err != nil {
		return brew, err
	}

	// Populate days off roast if not already set
	if brew.DaysOffRoast == 0 {
		coffee, err := s.coffeeStorage.GetByID(brew.CoffeeID)
		if err == nil {
			brew.DaysOffRoast = coffee.DaysOffRoastAt(brew.CreatedAt)
		}
	}

	return brew, nil
}

// GetBrewsForCoffee retrieves all brews for a specific coffee
func (s *BrewService) GetBrewsForCoffee(coffeeID string) ([]models.Brew, error) {
	brews, err := s.storage.GetByCoffeeID(coffeeID)
	if err != nil {
		return nil, err
	}

	// Get coffee to calculate days off roast
	coffee, err := s.coffeeStorage.GetByID(coffeeID)
	if err != nil {
		return brews, nil // Return brews without days_off_roast if coffee lookup fails
	}

	// Populate days off roast for each brew
	for i := range brews {
		if brews[i].DaysOffRoast == 0 {
			brews[i].DaysOffRoast = coffee.DaysOffRoastAt(brews[i].CreatedAt)
		}
	}

	return brews, nil
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

// GetRecentBrewsWithCoffee retrieves the most recent brews with coffee name and origin
func (s *BrewService) GetRecentBrewsWithCoffee(limit int) ([]models.BrewWithCoffee, error) {
	return s.storage.GetRecentWithCoffee(limit)
}

// DeleteBrew removes a brew entry
func (s *BrewService) DeleteBrew(id string) error {
	return s.storage.Delete(id)
}

// CanGeneratePokemon checks if a coffee has enough brews to generate a Pokemon
func (s *BrewService) CanGeneratePokemon(coffeeID string, isFinished bool) (bool, error) {
	count, err := s.storage.GetBrewCount(coffeeID)
	if err != nil {
		return false, err
	}
	// Finished coffees can always generate Pokemon (even with 0 brews)
	if isFinished {
		return true, nil
	}
	return count >= models.RequiredBrewsForPokemon, nil
}

// GetBrewProgress returns the brew progress for Pokemon generation
func (s *BrewService) GetBrewProgress(coffeeID string, hasPokemon bool, isFinished bool) (models.BrewProgress, error) {
	count, err := s.storage.GetBrewCount(coffeeID)
	if err != nil {
		return models.BrewProgress{}, err
	}

	// Finished coffees can always generate Pokemon (even with 0 brews)
	var canGenerate bool
	if isFinished {
		canGenerate = !hasPokemon
	} else {
		canGenerate = count >= models.RequiredBrewsForPokemon && !hasPokemon
	}

	return models.BrewProgress{
		Count:              count,
		Required:           models.RequiredBrewsForPokemon,
		CanGeneratePokemon: canGenerate,
		HasPokemon:         hasPokemon,
		IsFinished:         isFinished,
	}, nil
}

// GetAggregatedData computes aggregated data from all brews of a coffee
func (s *BrewService) GetAggregatedData(coffeeID string) (*models.AggregatedBrewData, error) {
	brews, err := s.storage.GetByCoffeeID(coffeeID)
	if err != nil {
		return nil, err
	}

	if len(brews) == 0 {
		// Return empty/default aggregated data for coffees with no brews
		return &models.AggregatedBrewData{
			CoffeeID:  coffeeID,
			BrewCount: 0,
		}, nil
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
