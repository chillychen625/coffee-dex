package storage

import (
	"context"
	"go-coffee-log/models"
	"time"
)

// CoffeeStorage defines the interface for coffee data persistence
type CoffeeStorage interface {
	Save(ctx context.Context, coffee models.Coffee) error
	GetByID(ctx context.Context, id string) (models.Coffee, error)
	GetAll(ctx context.Context) ([]models.Coffee, error)
	GetRecent(ctx context.Context, limit int) ([]models.Coffee, error)
	Update(ctx context.Context, id string, coffee models.Coffee) error
	Delete(ctx context.Context, id string) error
}

// BrewerStorage defines the interface for brewer data persistence
type BrewerStorage interface {
	SaveBrewer(ctx context.Context, brewer models.Brewer) error
	GetBrewerByID(ctx context.Context, id string) (models.Brewer, error)
	GetAllBrewers(ctx context.Context) ([]models.Brewer, error)
	DeleteBrewer(ctx context.Context, id string) error
	UpdateBrewerRecipes(ctx context.Context, brewerID string, recipes []models.Recipe) error
}

// PokemonStorage defines the interface for Pokemon data operations
type PokemonStorage interface {
	GetAllPokemon(ctx context.Context) ([]models.Pokemon, error)
	GetPokemonByID(ctx context.Context, id int) (*models.Pokemon, error)
	GetPokemonByType(ctx context.Context, pokemonType string) ([]models.Pokemon, error)
	IsPokemonUsed(ctx context.Context, pokemonID int) (bool, error)
	ReservePokemon(ctx context.Context, pokemonID int, coffeeID string) error
	CreateCoffeePokemon(ctx context.Context, mapping models.CoffeePokemon) error
	GetCoffeePokemon(ctx context.Context, coffeeID string) (*models.CoffeePokemon, error)
	GetAllCoffeePokemon(ctx context.Context) ([]models.CoffeePokemon, error)
	UpdateCoffeePokemonNickname(ctx context.Context, coffeeID, nickname string) error
}

// BrewStorage defines the interface for brew data persistence
type BrewStorage interface {
	Save(ctx context.Context, brew models.Brew) error
	GetByID(ctx context.Context, id string) (models.Brew, error)
	GetByCoffeeID(ctx context.Context, coffeeID string) ([]models.Brew, error)
	GetBrewCount(ctx context.Context, coffeeID string) (int, error)
	GetAll(ctx context.Context) ([]models.Brew, error)
	GetRecent(ctx context.Context, limit int) ([]models.Brew, error)
	GetRecentWithCoffee(ctx context.Context, limit int) ([]models.BrewWithCoffee, error)
	GetLastBrewDates(ctx context.Context) (map[string]time.Time, error)
	ToggleBrewLearning(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
