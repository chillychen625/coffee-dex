package storage

import "go-coffee-log/models"

// CoffeeStorage defines the interface for coffee data persistence
type CoffeeStorage interface {
	Save(coffee models.Coffee) error
	GetByID(id string) (models.Coffee, error)
	GetAll() ([]models.Coffee, error)
	GetRecent(limit int) ([]models.Coffee, error)
	Update(id string, coffee models.Coffee) error
	Delete(id string) error
}

// BrewStorage defines the interface for brew data persistence
type BrewStorage interface {
	Save(brew models.Brew) error
	GetByID(id string) (models.Brew, error)
	GetByCoffeeID(coffeeID string) ([]models.Brew, error)
	GetBrewCount(coffeeID string) (int, error)
	GetAll() ([]models.Brew, error)
	GetRecent(limit int) ([]models.Brew, error)
	GetRecentWithCoffee(limit int) ([]models.BrewWithCoffee, error)
	Delete(id string) error
}
