package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-coffee-log/models"
	"time"

	"github.com/jackc/pgx/v5"
)

// PgPokemonStorage implements PokemonStorage using Postgres.
type PgPokemonStorage struct {
	db DBTX
}

// NewPgPokemonStorage creates a new PgPokemonStorage backed by db.
func NewPgPokemonStorage(db DBTX) *PgPokemonStorage {
	return &PgPokemonStorage{db: db}
}

// GetAllPokemon retrieves all Pokemon ordered by id.
func (s *PgPokemonStorage) GetAllPokemon(ctx context.Context) ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons
		ORDER BY id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query Pokemon: %w", err)
	}
	defer rows.Close()

	var pokemons []models.Pokemon

	for rows.Next() {
		var pokemon models.Pokemon
		var statsJSON []byte

		err := rows.Scan(
			&pokemon.ID, &pokemon.Name, &pokemon.Type,
			&pokemon.SpritePath, &statsJSON, &pokemon.Description,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Pokemon: %w", err)
		}

		if err := json.Unmarshal(statsJSON, &pokemon.BaseStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
		}

		pokemons = append(pokemons, pokemon)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return pokemons, nil
}

// GetPokemonByID retrieves a Pokemon by its numeric ID.
func (s *PgPokemonStorage) GetPokemonByID(ctx context.Context, id int) (*models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE id = $1
	`

	row := s.db.QueryRow(ctx, query, id)

	var pokemon models.Pokemon
	var statsJSON []byte

	err := row.Scan(
		&pokemon.ID, &pokemon.Name, &pokemon.Type,
		&pokemon.SpritePath, &statsJSON, &pokemon.Description,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("Pokemon not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get Pokemon: %w", err)
	}

	if err := json.Unmarshal(statsJSON, &pokemon.BaseStats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return &pokemon, nil
}

// GetPokemonByType retrieves Pokemon whose type contains pokemonType.
// ILIKE keeps SQLite's case-insensitive LIKE semantics (Postgres LIKE is
// case-sensitive).
func (s *PgPokemonStorage) GetPokemonByType(ctx context.Context, pokemonType string) ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE type ILIKE $1
		ORDER BY id
	`

	rows, err := s.db.Query(ctx, query, "%"+pokemonType+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query Pokemon by type: %w", err)
	}
	defer rows.Close()

	var pokemons []models.Pokemon

	for rows.Next() {
		var pokemon models.Pokemon
		var statsJSON []byte

		err := rows.Scan(
			&pokemon.ID, &pokemon.Name, &pokemon.Type,
			&pokemon.SpritePath, &statsJSON, &pokemon.Description,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Pokemon: %w", err)
		}

		if err := json.Unmarshal(statsJSON, &pokemon.BaseStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
		}

		pokemons = append(pokemons, pokemon)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return pokemons, nil
}

// IsPokemonUsed checks if a Pokemon is already mapped to a coffee.
func (s *PgPokemonStorage) IsPokemonUsed(ctx context.Context, pokemonID int) (bool, error) {
	query := "SELECT COUNT(*) FROM coffee_pokemon WHERE pokemon_id = $1"

	var count int
	err := s.db.QueryRow(ctx, query, pokemonID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check Pokemon usage: %w", err)
	}

	return count > 0, nil
}

// ReservePokemon reserves a Pokemon for a coffee by creating a minimal mapping record.
func (s *PgPokemonStorage) ReservePokemon(ctx context.Context, pokemonID int, coffeeID string) error {
	mapping := models.CoffeePokemon{
		ID:          fmt.Sprintf("reserved_%d_%s", pokemonID, coffeeID),
		CoffeeID:    coffeeID,
		PokemonID:   pokemonID,
		PokemonName: "Reserved",
		Level:       1,
		CreatedAt:   time.Now(),
	}

	return s.CreateCoffeePokemon(ctx, mapping)
}

// CreateCoffeePokemon creates a new coffee-Pokemon mapping.
func (s *PgPokemonStorage) CreateCoffeePokemon(ctx context.Context, mapping models.CoffeePokemon) error {
	traitMappingJSON, err := json.Marshal(mapping.TraitMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal trait mapping: %w", err)
	}

	query := `
		INSERT INTO coffee_pokemon (
			id, coffee_id, pokemon_id, nickname, level,
			mapping_confidence, description, trait_mapping, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = s.db.Exec(
		ctx,
		query,
		mapping.ID, mapping.CoffeeID, mapping.PokemonID,
		mapping.Nickname, mapping.Level,
		mapping.MappingConfidence, mapping.Description,
		traitMappingJSON, formatTime(mapping.CreatedAt),
	)

	if err != nil {
		return fmt.Errorf("failed to create coffee Pokemon mapping: %w", err)
	}

	return nil
}

// GetCoffeePokemon retrieves the Pokemon mapping for a coffee.
func (s *PgPokemonStorage) GetCoffeePokemon(ctx context.Context, coffeeID string) (*models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.description, cp.created_at,
		       p.name, p.type, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		WHERE cp.coffee_id = $1
	`

	row := s.db.QueryRow(ctx, query, coffeeID)

	var mapping models.CoffeePokemon
	var traitMappingJSON []byte
	var createdAtStr sql.NullString

	err := row.Scan(
		&mapping.ID, &mapping.CoffeeID, &mapping.PokemonID,
		&mapping.Nickname, &mapping.Level,
		&mapping.MappingConfidence, &mapping.Description,
		&createdAtStr, &mapping.PokemonName, &mapping.PokemonType,
		&traitMappingJSON,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("Pokemon mapping not found for coffee")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coffee Pokemon: %w", err)
	}

	mapping.CreatedAt = parseTime(createdAtStr)

	if err := json.Unmarshal(traitMappingJSON, &mapping.TraitMapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trait mapping: %w", err)
	}

	return &mapping, nil
}

// GetAllCoffeePokemon retrieves all coffee-Pokemon mappings ordered by created_at DESC.
func (s *PgPokemonStorage) GetAllCoffeePokemon(ctx context.Context) ([]models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.description, cp.created_at,
		       p.name, p.type, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		ORDER BY cp.created_at DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffee Pokemon: %w", err)
	}
	defer rows.Close()

	var mappings []models.CoffeePokemon

	for rows.Next() {
		var mapping models.CoffeePokemon
		var traitMappingJSON []byte
		var createdAtStr sql.NullString

		err := rows.Scan(
			&mapping.ID, &mapping.CoffeeID, &mapping.PokemonID,
			&mapping.Nickname, &mapping.Level,
			&mapping.MappingConfidence, &mapping.Description,
			&createdAtStr, &mapping.PokemonName, &mapping.PokemonType,
			&traitMappingJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee Pokemon: %w", err)
		}

		mapping.CreatedAt = parseTime(createdAtStr)

		if err := json.Unmarshal(traitMappingJSON, &mapping.TraitMapping); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trait mapping: %w", err)
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return mappings, nil
}

// UpdateCoffeePokemonNickname updates the nickname of a Pokemon mapping for a coffee.
func (s *PgPokemonStorage) UpdateCoffeePokemonNickname(ctx context.Context, coffeeID, nickname string) error {
	query := "UPDATE coffee_pokemon SET nickname = $1 WHERE coffee_id = $2"

	result, err := s.db.Exec(ctx, query, nickname, coffeeID)
	if err != nil {
		return fmt.Errorf("failed to update nickname: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Pokemon mapping not found for coffee")
	}

	return nil
}
