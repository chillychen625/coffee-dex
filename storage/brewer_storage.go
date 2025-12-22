package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"log"
)

// BrewerStorage defines the interface for brewer data persistence
type BrewerStorage interface {
	SaveBrewer(brewer models.Brewer) error
	GetBrewerByID(id string) (models.Brewer, error)
	GetAllBrewers() ([]models.Brewer, error)
	DeleteBrewer(id string) error
	UpdateBrewerRecipes(brewerID string, recipes []models.Recipe) error
}

// MySQLBrewerStorage implements BrewerStorage using MySQL database
type MySQLBrewerStorage struct {
	db            *sql.DB
	coffeeStorage CoffeeStorage
}

// NewMySQLBrewerStorage creates a new MySQL brewer storage
func NewMySQLBrewerStorage(db *sql.DB, coffeeStorage CoffeeStorage) *MySQLBrewerStorage {
	storage := &MySQLBrewerStorage{
		db:            db,
		coffeeStorage: coffeeStorage,
	}
	
	if err := storage.initTables(); err != nil {
		panic(fmt.Sprintf("failed to initialize brewer tables: %v", err))
	}
	
	return storage
}

// initTables creates the brewers table if it doesn't exist
func (m *MySQLBrewerStorage) initTables() error {
	log.Printf("DEBUG: initTables - Creating brewers table if needed")
	brewerTableQuery := `
		CREATE TABLE IF NOT EXISTS brewers (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			pokeball_type VARCHAR(50) NOT NULL,
			recipes JSON,
			created_at DATETIME
		)
	`
	
	if _, err := m.db.Exec(brewerTableQuery); err != nil {
		log.Printf("ERROR: initTables - Failed to create brewers table: %v", err)
		return fmt.Errorf("failed to create brewers table: %w", err)
	}
	
	log.Printf("INFO: initTables - Brewers table created/verified successfully")
	return nil
}

// SaveBrewer stores a brewer in the database
func (m *MySQLBrewerStorage) SaveBrewer(brewer models.Brewer) error {
	log.Printf("DEBUG: SaveBrewer - Saving brewer: %s (ID: %s)", brewer.Name, brewer.ID)
	recipesJSON, err := json.Marshal(brewer.Recipes)
	if err != nil {
		log.Printf("ERROR: SaveBrewer - Marshal recipes failed: %v", err)
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}
	
	query := `
		INSERT INTO brewers (id, name, pokeball_type, recipes, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	_, err = m.db.Exec(query, brewer.ID, brewer.Name, brewer.PokeballType, recipesJSON, brewer.CreatedAt)
	if err != nil {
		log.Printf("ERROR: SaveBrewer - Insert failed: %v", err)
		return fmt.Errorf("failed to save brewer: %w", err)
	}
	
	log.Printf("INFO: SaveBrewer - Successfully saved brewer: %s", brewer.Name)
	return nil
}

// GetBrewerByID retrieves a brewer by ID
func (m *MySQLBrewerStorage) GetBrewerByID(id string) (models.Brewer, error) {
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers WHERE id = ?
	`
	
	var brewer models.Brewer
	var recipesJSON []byte
	err := m.db.QueryRow(query, id).Scan(
		&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &brewer.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return models.Brewer{}, fmt.Errorf("brewer not found")
	}
	if err != nil {
		return models.Brewer{}, fmt.Errorf("failed to get brewer: %w", err)
	}
	
	// Unmarshal recipes
	if len(recipesJSON) > 0 {
		if err := json.Unmarshal(recipesJSON, &brewer.Recipes); err != nil {
			return models.Brewer{}, fmt.Errorf("failed to unmarshal recipes: %w", err)
		}
	}
	
	return brewer, nil
}

// GetAllBrewers retrieves all brewers
func (m *MySQLBrewerStorage) GetAllBrewers() ([]models.Brewer, error) {
	log.Printf("DEBUG: GetAllBrewers - Starting query")
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers
		ORDER BY created_at ASC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		log.Printf("ERROR: GetAllBrewers - Query failed: %v", err)
		return nil, fmt.Errorf("failed to query brewers: %w", err)
	}
	defer rows.Close()
	
	var brewers []models.Brewer
	for rows.Next() {
		var brewer models.Brewer
		var recipesJSON []byte
		if err := rows.Scan(&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &brewer.CreatedAt); err != nil {
			log.Printf("ERROR: GetAllBrewers - Scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan brewer: %w", err)
		}
		
		// Unmarshal recipes
		if len(recipesJSON) > 0 {
			if err := json.Unmarshal(recipesJSON, &brewer.Recipes); err != nil {
				log.Printf("ERROR: GetAllBrewers - Unmarshal recipes failed: %v", err)
				return nil, fmt.Errorf("failed to unmarshal recipes: %w", err)
			}
		}
		
		brewers = append(brewers, brewer)
	}
	
	log.Printf("DEBUG: GetAllBrewers - Successfully retrieved %d brewers", len(brewers))
	return brewers, nil
}

// DeleteBrewer removes a brewer and all its recipes
func (m *MySQLBrewerStorage) DeleteBrewer(id string) error {
	query := "DELETE FROM brewers WHERE id = ?"
	
	result, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brewer: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("brewer not found")
	}
	
	return nil
}


// UpdateBrewerRecipes updates the standalone recipes for a brewer
func (m *MySQLBrewerStorage) UpdateBrewerRecipes(brewerID string, recipes []models.Recipe) error {
	// Validate recipe count (max 4)
	if len(recipes) > 4 {
		return fmt.Errorf("maximum of 4 recipes allowed per brewer")
	}
	
	recipesJSON, err := json.Marshal(recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}
	
	query := "UPDATE brewers SET recipes = ? WHERE id = ?"
	result, err := m.db.Exec(query, recipesJSON, brewerID)
	if err != nil {
		return fmt.Errorf("failed to update brewer recipes: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("brewer not found")
	}
	
	return nil
}