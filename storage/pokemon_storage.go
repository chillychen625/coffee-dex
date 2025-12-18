package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

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

// MySQLPokemonStorage implements PokemonStorage using MySQL
type MySQLPokemonStorage struct {
	db *sql.DB
}

// NewMySQLPokemonStorage creates a new Pokemon storage
func NewMySQLPokemonStorage(db *sql.DB) *MySQLPokemonStorage {
	return &MySQLPokemonStorage{db: db}
}

// initPokemonTable creates the Pokemon-related tables
func (m *MySQLPokemonStorage) initPokemonTable() error {
	// Pokemon reference table
	query := `
		CREATE TABLE IF NOT EXISTS pokemons (
			id INT PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			type VARCHAR(50) NOT NULL,
			sprite_path VARCHAR(255) NOT NULL,
			base_stats JSON NOT NULL,
			description TEXT
		)
	`
	
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create pokemons table: %w", err)
	}
	
	// Coffee-Pokemon mapping table
	query = `
		CREATE TABLE IF NOT EXISTS coffee_pokemon (
			id VARCHAR(36) PRIMARY KEY,
			coffee_id VARCHAR(36) NOT NULL,
			pokemon_id INT NOT NULL,
			nickname VARCHAR(100),
			level INT DEFAULT 1,
			mapping_confidence REAL,
			llm_description TEXT,
			trait_mapping JSON,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (coffee_id) REFERENCES coffees(id),
			FOREIGN KEY (pokemon_id) REFERENCES pokemons(id)
		)
	`
	
	_, err = m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create coffee_pokemon table: %w", err)
	}
	
	// Unique index to prevent duplicate Pokemon
	query = `CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_pokemon ON coffee_pokemon(pokemon_id)`
	_, err = m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create unique index: %w", err)
	}
	
	return nil
}

// GetAllPokemon retrieves all Pokemon
func (m *MySQLPokemonStorage) GetAllPokemon() ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons
		ORDER BY id
	`
	
	rows, err := m.db.Query(query)
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
	
	return pokemons, nil
}

// GetPokemonByID retrieves a Pokemon by ID
func (m *MySQLPokemonStorage) GetPokemonByID(id int) (*models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE id = ?
	`
	
	row := m.db.QueryRow(query, id)
	
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

// GetPokemonByType retrieves Pokemon by type
func (m *MySQLPokemonStorage) GetPokemonByType(pokemonType string) ([]models.Pokemon, error) {
	query := `
		SELECT id, name, type, sprite_path, base_stats, description
		FROM pokemons WHERE type LIKE ?
		ORDER BY id
	`
	
	rows, err := m.db.Query(query, "%"+pokemonType+"%")
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
	
	return pokemons, nil
}

// IsPokemonUsed checks if a Pokemon is already mapped to a coffee
func (m *MySQLPokemonStorage) IsPokemonUsed(pokemonID int) (bool, error) {
	query := "SELECT COUNT(*) FROM coffee_pokemon WHERE pokemon_id = ?"
	
	var count int
	err := m.db.QueryRow(query, pokemonID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check Pokemon usage: %w", err)
	}
	
	return count > 0, nil
}

// ReservePokemon reserves a Pokemon for a coffee (placeholder for future use)
func (m *MySQLPokemonStorage) ReservePokemon(pokemonID int, coffeeID string) error {
	// For now, just create the mapping to reserve the Pokemon
	mapping := models.CoffeePokemon{
		ID:          fmt.Sprintf("reserved_%d_%s", pokemonID, coffeeID),
		CoffeeID:    coffeeID,
		PokemonID:   pokemonID,
		PokemonName: "Reserved",
		Level:       1,
		CreatedAt:   time.Now(),
	}
	
	return m.CreateCoffeePokemon(mapping)
}

// CreateCoffeePokemon creates a new coffee-Pokemon mapping
func (m *MySQLPokemonStorage) CreateCoffeePokemon(mapping models.CoffeePokemon) error {
	traitMappingJSON, err := json.Marshal(mapping.TraitMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal trait mapping: %w", err)
	}
	
	query := `
		INSERT INTO coffee_pokemon (
			id, coffee_id, pokemon_id, nickname, level,
			mapping_confidence, llm_description, trait_mapping
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = m.db.Exec(
		query,
		mapping.ID, mapping.CoffeeID, mapping.PokemonID,
		mapping.Nickname, mapping.Level,
		mapping.MappingConfidence, mapping.LLMDescription,
		traitMappingJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create coffee Pokemon mapping: %w", err)
	}
	
	return nil
}

// GetCoffeePokemon retrieves Pokemon mapping for a coffee
func (m *MySQLPokemonStorage) GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.llm_description, cp.created_at,
		       p.name, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		WHERE cp.coffee_id = ?
	`
	
	row := m.db.QueryRow(query, coffeeID)
	
	var mapping models.CoffeePokemon
	var traitMappingJSON []byte
	
	err := row.Scan(
		&mapping.ID, &mapping.CoffeeID, &mapping.PokemonID,
		&mapping.Nickname, &mapping.Level,
		&mapping.MappingConfidence, &mapping.LLMDescription,
		&mapping.CreatedAt, &mapping.PokemonName,
		&traitMappingJSON,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Pokemon mapping not found for coffee")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coffee Pokemon: %w", err)
	}
	
	if err := json.Unmarshal(traitMappingJSON, &mapping.TraitMapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trait mapping: %w", err)
	}
	
	return &mapping, nil
}

// GetAllCoffeePokemon retrieves all coffee-Pokemon mappings
func (m *MySQLPokemonStorage) GetAllCoffeePokemon() ([]models.CoffeePokemon, error) {
	query := `
		SELECT cp.id, cp.coffee_id, cp.pokemon_id, cp.nickname, cp.level,
		       cp.mapping_confidence, cp.llm_description, cp.created_at,
		       p.name, cp.trait_mapping
		FROM coffee_pokemon cp
		JOIN pokemons p ON cp.pokemon_id = p.id
		ORDER BY cp.created_at DESC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffee Pokemon: %w", err)
	}
	defer rows.Close()
	
	var mappings []models.CoffeePokemon
	
	for rows.Next() {
		var mapping models.CoffeePokemon
		var traitMappingJSON []byte
		
		err := rows.Scan(
			&mapping.ID, &mapping.CoffeeID, &mapping.PokemonID,
			&mapping.Nickname, &mapping.Level,
			&mapping.MappingConfidence, &mapping.LLMDescription,
			&mapping.CreatedAt, &mapping.PokemonName,
			&traitMappingJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee Pokemon: %w", err)
		}
		
		if err := json.Unmarshal(traitMappingJSON, &mapping.TraitMapping); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trait mapping: %w", err)
		}
		
		mappings = append(mappings, mapping)
	}
	
	return mappings, nil
}

// UpdateCoffeePokemonNickname updates the nickname of a Pokemon
func (m *MySQLPokemonStorage) UpdateCoffeePokemonNickname(coffeeID, nickname string) error {
	query := "UPDATE coffee_pokemon SET nickname = ? WHERE coffee_id = ?"
	
	result, err := m.db.Exec(query, nickname, coffeeID)
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