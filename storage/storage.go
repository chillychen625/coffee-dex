package storage

import (
	"go-coffee-log/models"
	"time"
)

// CoffeeStorage defines the interface for coffee data persistence
type CoffeeStorage interface {
	Save(coffee models.Coffee) error
	GetByID(id string) (models.Coffee, error)
	GetAll() ([]models.Coffee, error)
	GetRecent(limit int) ([]models.Coffee, error)
	Update(id string, coffee models.Coffee) error
	Delete(id string) error
}

// BrewerStorage defines the interface for brewer data persistence
type BrewerStorage interface {
	SaveBrewer(brewer models.Brewer) error
	GetBrewerByID(id string) (models.Brewer, error)
	GetAllBrewers() ([]models.Brewer, error)
	DeleteBrewer(id string) error
	UpdateBrewerRecipes(brewerID string, recipes []models.Recipe) error
}

// PokemonStorage defines the interface for Pokemon data operations
type PokemonStorage interface {
	GetAllPokemon() ([]models.Pokemon, error)
	GetPokemonByID(id int) (*models.Pokemon, error)
	GetPokemonByType(pokemonType string) ([]models.Pokemon, error)
	IsPokemonUsed(pokemonID int) (bool, error)
	ReservePokemon(pokemonID int, coffeeID string) error
	CreateCoffeePokemon(mapping models.CoffeePokemon) error
	GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error)
	GetAllCoffeePokemon() ([]models.CoffeePokemon, error)
	UpdateCoffeePokemonNickname(coffeeID, nickname string) error
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
	GetLastBrewDates() (map[string]time.Time, error)
	ToggleBrewLearning(id string) error
	Delete(id string) error
}
