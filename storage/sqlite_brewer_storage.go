package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
)

// SQLiteBrewerStorage implements BrewerStorage using SQLite.
type SQLiteBrewerStorage struct {
	db *sql.DB
}

// NewSQLiteBrewerStorage creates a new SQLiteBrewerStorage backed by db.
func NewSQLiteBrewerStorage(db *sql.DB) *SQLiteBrewerStorage {
	return &SQLiteBrewerStorage{db: db}
}

// SaveBrewer stores a brewer in the database.
func (s *SQLiteBrewerStorage) SaveBrewer(brewer models.Brewer) error {
	recipesJSON, err := json.Marshal(brewer.Recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}

	query := `
		INSERT INTO brewers (id, name, pokeball_type, recipes, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, brewer.ID, brewer.Name, brewer.PokeballType, recipesJSON, formatTime(brewer.CreatedAt))
	if err != nil {
		return fmt.Errorf("failed to save brewer: %w", err)
	}

	return nil
}

// GetBrewerByID retrieves a brewer by ID.
func (s *SQLiteBrewerStorage) GetBrewerByID(id string) (models.Brewer, error) {
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers WHERE id = ?
	`

	var brewer models.Brewer
	var recipesJSON []byte
	var createdAtStr sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &createdAtStr,
	)

	if err == sql.ErrNoRows {
		return models.Brewer{}, fmt.Errorf("brewer not found")
	}
	if err != nil {
		return models.Brewer{}, fmt.Errorf("failed to get brewer: %w", err)
	}

	brewer.CreatedAt = parseTime(createdAtStr)

	if len(recipesJSON) > 0 {
		if err := json.Unmarshal(recipesJSON, &brewer.Recipes); err != nil {
			return models.Brewer{}, fmt.Errorf("failed to unmarshal recipes: %w", err)
		}
	}

	return brewer, nil
}

// GetAllBrewers retrieves all brewers ordered by created_at ASC.
func (s *SQLiteBrewerStorage) GetAllBrewers() ([]models.Brewer, error) {
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brewers: %w", err)
	}
	defer rows.Close()

	var brewers []models.Brewer

	for rows.Next() {
		var brewer models.Brewer
		var recipesJSON []byte
		var createdAtStr sql.NullString

		if err := rows.Scan(&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan brewer: %w", err)
		}

		brewer.CreatedAt = parseTime(createdAtStr)

		if len(recipesJSON) > 0 {
			if err := json.Unmarshal(recipesJSON, &brewer.Recipes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal recipes: %w", err)
			}
		}

		brewers = append(brewers, brewer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return brewers, nil
}

// DeleteBrewer removes a brewer from the database.
func (s *SQLiteBrewerStorage) DeleteBrewer(id string) error {
	query := "DELETE FROM brewers WHERE id = ?"

	result, err := s.db.Exec(query, id)
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

// UpdateBrewerRecipes updates the standalone recipes for a brewer (max 4).
func (s *SQLiteBrewerStorage) UpdateBrewerRecipes(brewerID string, recipes []models.Recipe) error {
	if len(recipes) > 4 {
		return fmt.Errorf("maximum of 4 recipes allowed per brewer")
	}

	recipesJSON, err := json.Marshal(recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}

	query := "UPDATE brewers SET recipes = ? WHERE id = ?"
	result, err := s.db.Exec(query, recipesJSON, brewerID)
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
