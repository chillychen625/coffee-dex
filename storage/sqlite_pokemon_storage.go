package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"time"
)

// SQLitePokemonStorage implements PokemonStorage using SQLite.
type SQLitePokemonStorage struct {
	db *sql.DB
}

// NewSQLitePokemonStorage creates a new SQLitePokemonStorage backed by db.
func NewSQLitePokemonStorage(db *sql.DB) *SQLitePokemonStorage {
	return &SQLitePokemonStorage{db: db}
}

// GetAllPokemon retrieves all Pokemon ordered by id.
func (s *SQLitePokemonStorage) GetAllPokemon() ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons
		ORDER BY id
	`

	rows, err := s.db.Query(query)
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
func (s *SQLitePokemonStorage) GetPokemonByID(id int) (*models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE id = ?
	`

	row := s.db.QueryRow(query, id)

	var pokemon models.Pokemon
	var statsJSON []byte

	err := row.Scan(
		&pokemon.ID, &pokemon.Name, &pokemon.Type,
		&pokemon.SpritePath, &statsJSON, &pokemon.Description,
	)

	if err == sql.ErrNoRows {
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
func (s *SQLitePokemonStorage) GetPokemonByType(pokemonType string) ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE type LIKE ?
		ORDER BY id
	`

	rows, err := s.db.Query(query, "%"+pokemonType+"%")
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
func (s *SQLitePokemonStorage) IsPokemonUsed(pokemonID int) (bool, error) {
	query := "SELECT COUNT(*) FROM coffee_pokemon WHERE pokemon_id = ?"

	var count int
	err := s.db.QueryRow(query, pokemonID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check Pokemon usage: %w", err)
	}

	return count > 0, nil
}

// ReservePokemon reserves a Pokemon for a coffee by creating a minimal mapping record.
func (s *SQLitePokemonStorage) ReservePokemon(pokemonID int, coffeeID string) error {
	mapping := models.CoffeePokemon{
		ID:          fmt.Sprintf("reserved_%d_%s", pokemonID, coffeeID),
		CoffeeID:    coffeeID,
		PokemonID:   pokemonID,
		PokemonName: "Reserved",
		Level:       1,
		CreatedAt:   time.Now(),
	}

	return s.CreateCoffeePokemon(mapping)
}

// CreateCoffeePokemon creates a new coffee-Pokemon mapping.
func (s *SQLitePokemonStorage) CreateCoffeePokemon(mapping models.CoffeePokemon) error {
	traitMappingJSON, err := json.Marshal(mapping.TraitMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal trait mapping: %w", err)
	}

	query := `
		INSERT INTO coffee_pokemon (
			id, coffee_id, pokemon_id, nickname, level,
			mapping_confidence, llm_description, trait_mapping, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(
		query,
		mapping.ID, mapping.CoffeeID, mapping.PokemonID,
		mapping.Nickname, mapping.Level,
		mapping.MappingConfidence, mapping.LLMDescription,
		traitMappingJSON, formatTime(mapping.CreatedAt),
	)

	if err != nil {
		return fmt.Errorf("failed to create coffee Pokemon mapping: %w", err)
	}

	return nil
}

// GetCoffeePokemon retrieves the Pokemon mapping for a coffee.
func (s *SQLitePokemonStorage) GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.llm_description, cp.created_at,
		       p.name, p.type, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		WHERE cp.coffee_id = ?
	`

	row := s.db.QueryRow(query, coffeeID)

	var mapping models.CoffeePokemon
	var traitMappingJSON []byte
	var createdAtStr sql.NullString

	err := row.Scan(
		&mapping.ID, &mapping.CoffeeID, &mapping.PokemonID,
		&mapping.Nickname, &mapping.Level,
		&mapping.MappingConfidence, &mapping.LLMDescription,
		&createdAtStr, &mapping.PokemonName, &mapping.PokemonType,
		&traitMappingJSON,
	)

	if err == sql.ErrNoRows {
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
func (s *SQLitePokemonStorage) GetAllCoffeePokemon() ([]models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.llm_description, cp.created_at,
		       p.name, p.type, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		ORDER BY cp.created_at DESC
	`

	rows, err := s.db.Query(query)
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
			&mapping.MappingConfidence, &mapping.LLMDescription,
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
func (s *SQLitePokemonStorage) UpdateCoffeePokemonNickname(coffeeID, nickname string) error {
	query := "UPDATE coffee_pokemon SET nickname = ? WHERE coffee_id = ?"

	result, err := s.db.Exec(query, nickname, coffeeID)
	if err != nil {
		return fmt.Errorf("failed to update nickname: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Pokemon mapping not found for coffee")
	}

	return nil
}
