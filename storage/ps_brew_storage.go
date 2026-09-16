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

type PgBrewStorage struct {
	db DBTX
}

func NewPgBrewStorage(db DBTX) *PgBrewStorage {
	return &PgBrewStorage{db: db}
}

func (s *PgBrewStorage) Save(ctx context.Context, brew models.Brew) error {
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

	isLearning := 0
	if brew.IsLearning {
		isLearning = 1
	}

	query := `
		INSERT INTO brews (
			id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
			end_time_minutes, end_time_seconds, created_at, is_learning
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = s.db.Exec(
		ctx,
		query,
		brew.ID, brew.CoffeeID,
		tastingNotesJSON, tastingTraitsJSON, brew.Rating, recipeJSON, brew.Dripper,
		brew.EndTime.Minutes, brew.EndTime.Seconds,
		formatTime(brew.CreatedAt),
		isLearning,
	)

	if err != nil {
		return fmt.Errorf("failed to save brew: %w", err)
	}

	return nil
}

func (s *PgBrewStorage) GetByID(ctx context.Context, id string) (models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews WHERE id = $1
	`
	row := s.db.QueryRow(ctx, query, id)
	return s.scanBrew(row)
}

func (s *PgBrewStorage) GetByCoffeeID(ctx context.Context, coffeeID string) ([]models.Brew, error){
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		WHERE coffee_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

func (s *PgBrewStorage) GetBrewCount(ctx context.Context, coffeeID string) (int, error){
	query := `SELECT COUNT(*) FROM brews WHERE coffee_id = $1`

	var count int
	err := s.db.QueryRow(ctx, query, coffeeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count brews: %w", err)
	}

	return count, nil
}

func (s *PgBrewStorage) GetAll(ctx context.Context) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

func (s *PgBrewStorage) GetRecent(ctx context.Context, limit int) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

func (s *PgBrewStorage) GetRecentWithCoffee(ctx context.Context, limit int) ([]models.BrewWithCoffee, error) {
	query := `
		SELECT b.id, b.coffee_id, b.tasting_notes, b.tasting_traits, b.rating, b.recipe, b.dripper,
		       b.end_time_minutes, b.end_time_seconds, b.created_at, COALESCE(b.is_learning, 0),
		       c.name AS coffee_name, c.origin AS coffee_origin, c.roast_date
		FROM brews b
		JOIN coffees c ON b.coffee_id = c.id
		ORDER BY b.created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent brews with coffee: %w", err)
	}
	defer rows.Close()

	var results []models.BrewWithCoffee

	for rows.Next() {
		var bwc models.BrewWithCoffee
		var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
		var createdAtStr sql.NullString
		var roastDate sql.NullString
		var isLearningInt int

		err := rows.Scan(
			&bwc.ID, &bwc.CoffeeID,
			&tastingNotesJSON, &tastingTraitsJSON, &bwc.Rating, &recipeJSON, &bwc.Dripper,
			&bwc.EndTime.Minutes, &bwc.EndTime.Seconds,
			&createdAtStr, &isLearningInt,
			&bwc.CoffeeName, &bwc.CoffeeOrigin, &roastDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brew with coffee: %w", err)
		}

		bwc.CreatedAt = parseTime(createdAtStr)
		bwc.IsLearning = isLearningInt != 0

		if err := json.Unmarshal(tastingNotesJSON, &bwc.TastingNotes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
		}

		if err := json.Unmarshal(tastingTraitsJSON, &bwc.TastingTraits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
		}

		if err := json.Unmarshal(recipeJSON, &bwc.Recipe); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
		}

		// Calculate days off roast from the coffee's roast date.
		rd := parseDateOnly(roastDate)
		if rd != nil {
			bwc.DaysOffRoast = int(bwc.CreatedAt.Sub(rd.Time()).Hours() / 24)
		} else {
			bwc.DaysOffRoast = -1
		}

		results = append(results, bwc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}
func (s *PgBrewStorage) GetLastBrewDates(ctx context.Context) (map[string]time.Time, error) {
	query := `SELECT coffee_id, MAX(created_at) FROM brews GROUP BY coffee_id`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last brew dates: %w", err)
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var coffeeID string
		var ts sql.NullString
		if err := rows.Scan(&coffeeID, &ts); err != nil {
			return nil, fmt.Errorf("failed to scan last brew date: %w", err)
		}
		result[coffeeID] = parseTime(ts)
	}
	return result, rows.Err()
}

func (s *PgBrewStorage) ToggleBrewLearning(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, "UPDATE brews SET is_learning = 1 - COALESCE(is_learning, 0) WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to toggle brew learning: %w", err)
	}
	return nil
}

func (s *PgBrewStorage) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM brews WHERE id = $1"

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brew: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("brew not found")
	}

	return nil
}

func (s *PgBrewStorage) scanBrew(row pgx.Row) (models.Brew, error) {
	var brew models.Brew
	var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
	var createdAtStr sql.NullString
	var isLearningInt int

	err := row.Scan(
		&brew.ID, &brew.CoffeeID,
		&tastingNotesJSON, &tastingTraitsJSON, &brew.Rating, &recipeJSON, &brew.Dripper,
		&brew.EndTime.Minutes, &brew.EndTime.Seconds,
		&createdAtStr, &isLearningInt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Brew{}, fmt.Errorf("brew not found")
	}
	if err != nil {
		return models.Brew{}, fmt.Errorf("failed to get brew: %w", err)
	}

	brew.CreatedAt = parseTime(createdAtStr)
	brew.IsLearning = isLearningInt != 0

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

func (s *PgBrewStorage) scanBrews(rows pgx.Rows) ([]models.Brew, error) {
	var brews []models.Brew

	for rows.Next() {
		var brew models.Brew
		var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
		var createdAtStr sql.NullString
		var isLearningInt int

		err := rows.Scan(
			&brew.ID, &brew.CoffeeID,
			&tastingNotesJSON, &tastingTraitsJSON, &brew.Rating, &recipeJSON, &brew.Dripper,
			&brew.EndTime.Minutes, &brew.EndTime.Seconds,
			&createdAtStr, &isLearningInt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan brew: %w", err)
		}

		brew.CreatedAt = parseTime(createdAtStr)
		brew.IsLearning = isLearningInt != 0

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