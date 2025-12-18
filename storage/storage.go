package storage

import "go-coffee-log/models"

// CoffeeStorage defines the interface for coffee data persistence
// This allows us to swap different storage implementations (memory, database, etc.)
// TODO: Define the following methods:
//   - Save(coffee models.Coffee) error
//   - GetByID(id string) (models.Coffee, error)
//   - GetAll() ([]models.Coffee, error)
//   - Update(id string, coffee models.Coffee) error
//   - Delete(id string) error
type CoffeeStorage interface {
	Save(coffee models.Coffee) error
	GetByID(id string) (models.Coffee, error)
	GetAll() ([]models.Coffee, error)
	Update(id string, coffee models.Coffee) error
	Delete(id string) error
}