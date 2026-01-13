package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
)

// MySQLBrewStorage implements BrewStorage using MySQL database
type MySQLBrewStorage struct {
	db *sql.DB
}

// NewMySQLBrewStorage creates a new MySQL brew storage
func NewMySQLBrewStorage(db *sql.DB) *MySQLBrewStorage {
	return &MySQLBrewStorage{db: db}
}

// Save stores a brew entry in the database
func (m *MySQLBrewStorage) Save(brew models.Brew) error {
	tastingNotesJSON, err := json.Marshal(brew.TastingNotes)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting notes: %w", err)
	}

	tastingTraitsJSON, err := json.Marshal(brew.TastingTraits)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting traits: %w", err)
	}

	recipeJSON, err := json.Marshal(brew.Recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe: %w", err)
	}

	query := `
		INSERT INTO brews (
			id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
			end_time_minutes, end_time_seconds, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = m.db.Exec(
		query,
		brew.ID, brew.CoffeeID,
		tastingNotesJSON, tastingTraitsJSON, brew.Rating, recipeJSON, brew.Dripper,
		brew.EndTime.Minutes, brew.EndTime.Seconds,
		brew.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save brew: %w", err)
	}

	return nil
}

// GetByID retrieves a brew by ID from the database
func (m *MySQLBrewStorage) GetByID(id string) (models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at
		FROM brews WHERE id = ?
	`

	row := m.db.QueryRow(query, id)
	return m.scanBrew(row)
}

// GetByCoffeeID retrieves all brews for a specific coffee
func (m *MySQLBrewStorage) GetByCoffeeID(coffeeID string) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at
		FROM brews
		WHERE coffee_id = ?
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query, coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return m.scanBrews(rows)
}

// GetBrewCount returns the number of brews for a specific coffee
func (m *MySQLBrewStorage) GetBrewCount(coffeeID string) (int, error) {
	query := `SELECT COUNT(*) FROM brews WHERE coffee_id = ?`

	var count int
	err := m.db.QueryRow(query, coffeeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count brews: %w", err)
	}

	return count, nil
}

// GetAll retrieves all brews from the database
func (m *MySQLBrewStorage) GetAll() ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at
		FROM brews
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return m.scanBrews(rows)
}

// GetRecent retrieves the most recent brews from the database
func (m *MySQLBrewStorage) GetRecent(limit int) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at
		FROM brews
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent brews: %w", err)
	}
	defer rows.Close()

	return m.scanBrews(rows)
}

// Delete removes a brew entry from the database
func (m *MySQLBrewStorage) Delete(id string) error {
	query := "DELETE FROM brews WHERE id = ?"

	result, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brew: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("brew not found")
	}

	return nil
}

// scanBrew scans a single row into a Brew struct
func (m *MySQLBrewStorage) scanBrew(row *sql.Row) (models.Brew, error) {
	var brew models.Brew
	var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte

	err := row.Scan(
		&brew.ID, &brew.CoffeeID,
		&tastingNotesJSON, &tastingTraitsJSON, &brew.Rating, &recipeJSON, &brew.Dripper,
		&brew.EndTime.Minutes, &brew.EndTime.Seconds,
		&brew.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return models.Brew{}, fmt.Errorf("brew not found")
	}
	if err != nil {
		return models.Brew{}, fmt.Errorf("failed to get brew: %w", err)
	}

	if err := json.Unmarshal(tastingNotesJSON, &brew.TastingNotes); err != nil {
		return models.Brew{}, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
	}

	if err := json.Unmarshal(tastingTraitsJSON, &brew.TastingTraits); err != nil {
		return models.Brew{}, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
	}

	if err := json.Unmarshal(recipeJSON, &brew.Recipe); err != nil {
		return models.Brew{}, fmt.Errorf("failed to unmarshal recipe: %w", err)
	}

	return brew, nil
}

// scanBrews scans multiple rows into a slice of Brew structs
func (m *MySQLBrewStorage) scanBrews(rows *sql.Rows) ([]models.Brew, error) {
	var brews []models.Brew

	for rows.Next() {
		var brew models.Brew
		var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte

		err := rows.Scan(
			&brew.ID, &brew.CoffeeID,
			&tastingNotesJSON, &tastingTraitsJSON, &brew.Rating, &recipeJSON, &brew.Dripper,
			&brew.EndTime.Minutes, &brew.EndTime.Seconds,
			&brew.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan brew: %w", err)
		}

		if err := json.Unmarshal(tastingNotesJSON, &brew.TastingNotes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
		}

		if err := json.Unmarshal(tastingTraitsJSON, &brew.TastingTraits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
		}

		if err := json.Unmarshal(recipeJSON, &brew.Recipe); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
		}

		brews = append(brews, brew)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return brews, nil
}
