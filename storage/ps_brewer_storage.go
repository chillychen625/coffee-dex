package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-coffee-log/models"

	"github.com/jackc/pgx/v5"
)

// PgBrewerStorage implements BrewerStorage using Postgres.
type PgBrewerStorage struct {
	db DBTX
}

// NewPgBrewerStorage creates a new PgBrewerStorage backed by db.
func NewPgBrewerStorage(db DBTX) *PgBrewerStorage {
	return &PgBrewerStorage{db: db}
}

// SaveBrewer stores a brewer in the database.
func (s *PgBrewerStorage) SaveBrewer(ctx context.Context, brewer models.Brewer) error {
	recipesJSON, err := json.Marshal(brewer.Recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}

	query := `
		INSERT INTO brewers (id, name, pokeball_type, recipes, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = s.db.Exec(ctx, query, brewer.ID, brewer.Name, brewer.PokeballType, recipesJSON, formatTime(brewer.CreatedAt))
	if err != nil {
		return fmt.Errorf("failed to save brewer: %w", err)
	}

	return nil
}

// GetBrewerByID retrieves a brewer by ID.
func (s *PgBrewerStorage) GetBrewerByID(ctx context.Context, id string) (models.Brewer, error) {
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers WHERE id = $1
	`

	var brewer models.Brewer
	var recipesJSON []byte
	var createdAtStr sql.NullString

	err := s.db.QueryRow(ctx, query, id).Scan(
		&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &createdAtStr,
	)

	if errors.Is(err, pgx.ErrNoRows) {
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
func (s *PgBrewerStorage) GetAllBrewers(ctx context.Context) ([]models.Brewer, error) {
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(ctx, query)
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
func (s *PgBrewerStorage) DeleteBrewer(ctx context.Context, id string) error {
	query := "DELETE FROM brewers WHERE id = $1"

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brewer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("brewer not found")
	}

	return nil
}

// UpdateBrewerRecipes updates the standalone recipes for a brewer (max 4).
func (s *PgBrewerStorage) UpdateBrewerRecipes(ctx context.Context, brewerID string, recipes []models.Recipe) error {
	if len(recipes) > 4 {
		return fmt.Errorf("maximum of 4 recipes allowed per brewer")
	}

	recipesJSON, err := json.Marshal(recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}

	query := "UPDATE brewers SET recipes = $1 WHERE id = $2"
	result, err := s.db.Exec(ctx, query, recipesJSON, brewerID)
	if err != nil {
		return fmt.Errorf("failed to update brewer recipes: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("brewer not found")
	}

	return nil
}
