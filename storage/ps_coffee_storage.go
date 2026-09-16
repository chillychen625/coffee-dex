package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-coffee-log/models"

	"github.com/jackc/pgx/v5"
)

// PgCoffeeStorage implements CoffeeStorage using Postgres.
type PgCoffeeStorage struct {
	db DBTX
}

// NewPgCoffeeStorage creates a new PgCoffeeStorage backed by db.
func NewPgCoffeeStorage(db DBTX) *PgCoffeeStorage {
	return &PgCoffeeStorage{db: db}
}

// Save stores a coffee entry in the database.
func (s *PgCoffeeStorage) Save(ctx context.Context, coffee models.Coffee) error {
	query := `
		INSERT INTO coffees (
			id, name, origin, roaster, variety, roast_level, processing_method,
			roast_date, is_finished, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
		ctx,
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
func (s *PgCoffeeStorage) GetByID(ctx context.Context, id string) (models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees WHERE id = $1
	`

	row := s.db.QueryRow(ctx, query, id)
	return s.scanCoffee(row)
}

// GetAll retrieves all coffees from the database ordered by created_at DESC.
func (s *PgCoffeeStorage) GetAll(ctx context.Context) ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffees: %w", err)
	}
	defer rows.Close()

	return s.scanCoffees(rows)
}

// GetRecent retrieves the most recent coffees up to limit entries.
func (s *PgCoffeeStorage) GetRecent(ctx context.Context, limit int) ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at, finished_at
		FROM coffees
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent coffees: %w", err)
	}
	defer rows.Close()

	return s.scanCoffees(rows)
}

// Update modifies an existing coffee entry.
func (s *PgCoffeeStorage) Update(ctx context.Context, id string, coffee models.Coffee) error {
	var finishedAt interface{}
	if coffee.FinishedAt != nil {
		finishedAt = formatTime(*coffee.FinishedAt)
	}

	query := `
		UPDATE coffees SET
			name=$1, origin=$2, roaster=$3, variety=$4, roast_level=$5, processing_method=$6,
			roast_date=$7, is_finished=$8, updated_at=$9, finished_at=$10
		WHERE id=$11
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
		ctx,
		query,
		coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod, roastDate, isFinished,
		formatTime(coffee.UpdatedAt), finishedAt, id,
	)

	if err != nil {
		return fmt.Errorf("failed to update coffee: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("coffee not found")
	}

	return nil
}

// Delete removes a coffee entry from the database.
func (s *PgCoffeeStorage) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM coffees WHERE id = $1"

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete coffee: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("coffee not found")
	}

	return nil
}

// scanCoffee scans a single pgx.Row into a Coffee struct.
func (s *PgCoffeeStorage) scanCoffee(row pgx.Row) (models.Coffee, error) {
	var coffee models.Coffee
	var roastDate sql.NullString
	var isFinished sql.NullInt64
	var createdAt, updatedAt, finishedAt sql.NullString

	err := row.Scan(
		&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
		&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
		&createdAt, &updatedAt, &finishedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
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
func (s *PgCoffeeStorage) scanCoffees(rows pgx.Rows) ([]models.Coffee, error) {
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
