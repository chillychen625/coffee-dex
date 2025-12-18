package service

import (
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"time"

	"github.com/google/uuid"
)

// CoffeeService handles business logic for coffee operations
// TODO: Add the following field:
//   - storage (storage.CoffeeStorage) - the storage implementation to use
type CoffeeService struct {
	storage storage.CoffeeStorage
}

// NewCoffeeService creates a new coffee service
func NewCoffeeService(storage storage.CoffeeStorage) *CoffeeService {
	return &CoffeeService{storage: storage}
}

// CreateCoffee creates a new coffee entry
// TODO: Implement this method
// Requirements:
//   - Generate a unique ID (you can use a simple counter or UUID)
//   - Set CreatedAt and UpdatedAt to current time
//   - Validate the coffee data
//   - Save to storage
// HINT: Use time.Now() for timestamps
func (s *CoffeeService) CreateCoffee(coffee models.Coffee) (models.Coffee, error) {
	coffee.ID = uuid.New().String()
	coffee.CreatedAt = time.Now()
	coffee.UpdatedAt = time.Now()
	
	if err := coffee.Validate(); err != nil {
		return models.Coffee{}, err
	}
	
	if err := s.storage.Save(coffee); err != nil {
		return models.Coffee{}, err
	}
	
	return coffee, nil
}

// GetCoffee retrieves a coffee by ID
// TODO: Implement this method
// HINT: Delegate to storage.GetByID
func (s *CoffeeService) GetCoffee(id string) (models.Coffee, error) {
	coffee, err := s.storage.GetByID(id)
	if err != nil {
		return models.Coffee{}, err
	}
	return coffee, nil
}

// ListCoffees retrieves all coffees
// TODO: Implement this method
// HINT: Delegate to storage.GetAll
func (s *CoffeeService) ListCoffees() ([]models.Coffee, error) {
	return s.storage.GetAll()
}

// UpdateCoffee modifies an existing coffee
// TODO: Implement this method
// Requirements:
//   - Update the UpdatedAt timestamp
//   - Validate the new data
//   - Save to storage
func (s *CoffeeService) UpdateCoffee(id string, coffee models.Coffee) (models.Coffee, error) {
	coffee.ID = id  // Set the ID from the URL
	coffee.UpdatedAt = time.Now()
	
	if err := coffee.Validate(); err != nil {
		return models.Coffee{}, err
	}
	
	if err := s.storage.Update(id, coffee); err != nil {
		return models.Coffee{}, err
	}
	
	return coffee, nil  // ← Return the updated coffee, not empty!
}

// DeleteCoffee removes a coffee entry
// TODO: Implement this method
// HINT: Delegate to storage.Delete
func (s *CoffeeService) DeleteCoffee(id string) error {
	if err := s.storage.Delete(id); err != nil {
		return err
	}
	return nil
}