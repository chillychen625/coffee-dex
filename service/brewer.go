package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"time"

	"github.com/google/uuid"
)

// BrewerService handles brewer business logic
type BrewerService struct {
	storage storage.BrewerStorage
}

// NewBrewerService creates a new brewer service
func NewBrewerService(storage storage.BrewerStorage) *BrewerService {
	return &BrewerService{
		storage: storage,
	}
}

// CreateBrewer creates a new brewer
func (s *BrewerService) CreateBrewer(name, pokeballType string) (models.Brewer, error) {
	brewer := models.Brewer{
		ID:           uuid.New().String(),
		Name:         name,
		PokeballType: pokeballType,
		CreatedAt:    time.Now(),
	}
	
	if err := brewer.Validate(); err != nil {
		return models.Brewer{}, err
	}
	
	if err := s.storage.SaveBrewer(brewer); err != nil {
		return models.Brewer{}, err
	}
	
	return brewer, nil
}

// GetBrewerByID retrieves a brewer by ID
func (s *BrewerService) GetBrewerByID(id string) (models.Brewer, error) {
	return s.storage.GetBrewerByID(id)
}

// GetAllBrewers retrieves all brewers
func (s *BrewerService) GetAllBrewers() ([]models.Brewer, error) {
	return s.storage.GetAllBrewers()
}

// DeleteBrewer removes a brewer and all its recipes
func (s *BrewerService) DeleteBrewer(id string) error {
	return s.storage.DeleteBrewer(id)
}

// AddStandaloneRecipe adds a standalone brewing recipe to a brewer
func (s *BrewerService) AddStandaloneRecipe(brewerID, name string, steps []string) error {
	brewer, err := s.storage.GetBrewerByID(brewerID)
	if err != nil {
		return err
	}
	
	// Check recipe limit
	if len(brewer.Recipes) >= 4 {
		return fmt.Errorf("brewer already has maximum of 4 recipes")
	}
	
	// Create new recipe
	recipe := models.Recipe{
		ID:    uuid.New().String(),
		Name:  name,
		Steps: steps,
	}
	
	// Add recipe to brewer
	brewer.Recipes = append(brewer.Recipes, recipe)
	
	return s.storage.UpdateBrewerRecipes(brewerID, brewer.Recipes)
}

// RemoveStandaloneRecipe removes a standalone recipe from a brewer
func (s *BrewerService) RemoveStandaloneRecipe(brewerID, recipeID string) error {
	brewer, err := s.storage.GetBrewerByID(brewerID)
	if err != nil {
		return err
	}
	
	// Find and remove recipe
	var updatedRecipes []models.Recipe
	found := false
	for _, recipe := range brewer.Recipes {
		if recipe.ID != recipeID {
			updatedRecipes = append(updatedRecipes, recipe)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("recipe not found")
	}
	
	return s.storage.UpdateBrewerRecipes(brewerID, updatedRecipes)
}

// GetAvailablePokeballTypes returns the list of valid pokeball types
func (s *BrewerService) GetAvailablePokeballTypes() []string {
	return []string{"poke-ball", "great-ball", "ultra-ball", "fast-ball"}
}

// ValidateBrewerLimit checks if we've reached the maximum of 4 brewers
func (s *BrewerService) ValidateBrewerLimit() error {
	brewers, err := s.storage.GetAllBrewers()
	if err != nil {
		return err
	}
	
	if len(brewers) >= 4 {
		return fmt.Errorf("maximum of 4 brewers allowed")
	}
	
	return nil
}