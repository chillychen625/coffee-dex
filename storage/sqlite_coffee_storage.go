package storage

import (
	"database/sql"
	"fmt"
	"go-coffee-log/models"
	"time"
)

// Time helpers shared across SQLite storage implementations.

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseTime(s sql.NullString) time.Time {
	if !s.Valid || s.String == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s.String); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s.String); err == nil {
		return t
	}
	return time.Time{}
}

func parseDateOnly(s sql.NullString) *models.DateOnly {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s.String)
	if err != nil {
		return nil
	}
	d := models.DateOnly(t)
	return &d
}

// SQLiteCoffeeStorage implements CoffeeStorage using SQLite.
type SQLiteCoffeeStorage struct {
	db *sql.DB
}

// NewSQLiteCoffeeStorage creates a new SQLiteCoffeeStorage backed by db.
func NewSQLiteCoffeeStorage(db *sql.DB) *SQLiteCoffeeStorage {
	return &SQLiteCoffeeStorage{db: db}
}

// Save stores a coffee entry in the database.
func (s *SQLiteCoffeeStorage) Save(coffee models.Coffee) error {
	query := `
		INSERT INTO coffees (
			id, name, origin, roaster, variety, roast_level, processing_method,
			roast_date, is_finished, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var roastDate interface{}
	if coffee.RoastDate != nil && !coffee.RoastDate.IsZero() {
		roastDate = formatDateOnly(coffee.RoastDate.Time())
	}

	isFinished := 0
	if coffee.IsFinished {
		isFinished = 1
	}

	_, err := s.db.Exec(
		query,
		coffee.ID, coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod, roastDate, isFinished,
		formatTime(coffee.CreatedAt), formatTime(coffee.UpdatedAt),
	)

	if err != nil {
		return fmt.Errorf("failed to save coffee: %w", err)
	}

	return nil
}

// GetByID retrieves a coffee by ID from the database.
func (s *SQLiteCoffeeStorage) GetByID(id string) (models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	return s.scanCoffee(row)
}

// GetAll retrieves all coffees from the database ordered by created_at DESC.
func (s *SQLiteCoffeeStorage) GetAll() ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffees: %w", err)
	}
	defer rows.Close()

	return s.scanCoffees(rows)
}

// GetRecent retrieves the most recent coffees up to limit entries.
func (s *SQLiteCoffeeStorage) GetRecent(limit int) ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent coffees: %w", err)
	}
	defer rows.Close()

	return s.scanCoffees(rows)
}

// Update modifies an existing coffee entry.
func (s *SQLiteCoffeeStorage) Update(id string, coffee models.Coffee) error {
	var finishedAt interface{}
	if coffee.FinishedAt != nil {
		finishedAt = formatTime(*coffee.FinishedAt)
	}

	query := `
		UPDATE coffees SET
			name=?, origin=?, roaster=?, variety=?, roast_level=?, processing_method=?,
			roast_date=?, is_finished=?, updated_at=?, finished_at=?
		WHERE id=?
	`

	var roastDate interface{}
	if coffee.RoastDate != nil && !coffee.RoastDate.IsZero() {
		roastDate = formatDateOnly(coffee.RoastDate.Time())
	}

	isFinished := 0
	if coffee.IsFinished {
		isFinished = 1
	}

	result, err := s.db.Exec(
		query,
		coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod, roastDate, isFinished,
		formatTime(coffee.UpdatedAt), finishedAt, id,
	)

	if err != nil {
		return fmt.Errorf("failed to update coffee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("coffee not found")
	}

	return nil
}

// Delete removes a coffee entry from the database.
func (s *SQLiteCoffeeStorage) Delete(id string) error {
	query := "DELETE FROM coffees WHERE id = ?"

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete coffee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("coffee not found")
	}

	return nil
}

// scanCoffee scans a single *sql.Row into a Coffee struct.
func (s *SQLiteCoffeeStorage) scanCoffee(row *sql.Row) (models.Coffee, error) {
	var coffee models.Coffee
	var roastDate sql.NullString
	var isFinished sql.NullInt64
	var createdAt, updatedAt, finishedAt sql.NullString

	err := row.Scan(
		&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
		&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
		&createdAt, &updatedAt, &finishedAt,
	)

	if err == sql.ErrNoRows {
		return models.Coffee{}, fmt.Errorf("coffee not found")
	}
	if err != nil {
		return models.Coffee{}, fmt.Errorf("failed to get coffee: %w", err)
	}

	coffee.RoastDate = parseDateOnly(roastDate)
	coffee.IsFinished = isFinished.Valid && isFinished.Int64 != 0
	coffee.CreatedAt = parseTime(createdAt)
	coffee.UpdatedAt = parseTime(updatedAt)
	if t := parseTime(finishedAt); !t.IsZero() {
		coffee.FinishedAt = &t
	}

	return coffee, nil
}

// scanCoffees scans multiple rows into a Coffee slice.
func (s *SQLiteCoffeeStorage) scanCoffees(rows *sql.Rows) ([]models.Coffee, error) {
	var coffees []models.Coffee

	for rows.Next() {
		var coffee models.Coffee
		var roastDate sql.NullString
		var isFinished sql.NullInt64
		var createdAt, updatedAt, finishedAt sql.NullString

		err := rows.Scan(
			&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
			&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
			&createdAt, &updatedAt, &finishedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee: %w", err)
		}

		coffee.RoastDate = parseDateOnly(roastDate)
		coffee.IsFinished = isFinished.Valid && isFinished.Int64 != 0
		coffee.CreatedAt = parseTime(createdAt)
		coffee.UpdatedAt = parseTime(updatedAt)
		if t := parseTime(finishedAt); !t.IsZero() {
			coffee.FinishedAt = &t
		}

		coffees = append(coffees, coffee)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return coffees, nil
}
