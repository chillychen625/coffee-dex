package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"time"
)

// SQLiteBrewStorage implements BrewStorage using SQLite.
type SQLiteBrewStorage struct {
	db *sql.DB
}

// NewSQLiteBrewStorage creates a new SQLiteBrewStorage backed by db.
func NewSQLiteBrewStorage(db *sql.DB) *SQLiteBrewStorage {
	return &SQLiteBrewStorage{db: db}
}

// Save stores a brew entry in the database.
func (s *SQLiteBrewStorage) Save(brew models.Brew) error {
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(
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

// GetByID retrieves a brew by ID from the database.
func (s *SQLiteBrewStorage) GetByID(id string) (models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	return s.scanBrew(row)
}

// GetByCoffeeID retrieves all brews for a specific coffee ordered by created_at DESC.
func (s *SQLiteBrewStorage) GetByCoffeeID(coffeeID string) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		WHERE coffee_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, coffeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

// GetBrewCount returns the number of brews for a specific coffee.
func (s *SQLiteBrewStorage) GetBrewCount(coffeeID string) (int, error) {
	query := `SELECT COUNT(*) FROM brews WHERE coffee_id = ?`

	var count int
	err := s.db.QueryRow(query, coffeeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count brews: %w", err)
	}

	return count, nil
}

// GetAll retrieves all brews from the database ordered by created_at DESC.
func (s *SQLiteBrewStorage) GetAll() ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

// GetRecent retrieves the most recent brews up to limit entries.
func (s *SQLiteBrewStorage) GetRecent(limit int) ([]models.Brew, error) {
	query := `
		SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, COALESCE(is_learning, 0)
		FROM brews
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent brews: %w", err)
	}
	defer rows.Close()

	return s.scanBrews(rows)
}

// GetRecentWithCoffee retrieves the most recent brews joined with coffee info.
func (s *SQLiteBrewStorage) GetRecentWithCoffee(limit int) ([]models.BrewWithCoffee, error) {
	query := `
		SELECT b.id, b.coffee_id, b.tasting_notes, b.tasting_traits, b.rating, b.recipe, b.dripper,
		       b.end_time_minutes, b.end_time_seconds, b.created_at, COALESCE(b.is_learning, 0),
		       c.name AS coffee_name, c.origin AS coffee_origin, c.roast_date
		FROM brews b
		JOIN coffees c ON b.coffee_id = c.id
		ORDER BY b.created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
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

// GetLastBrewDates returns a map of coffeeID -> most recent brew time.
func (s *SQLiteBrewStorage) GetLastBrewDates() (map[string]time.Time, error) {
	query := `SELECT coffee_id, MAX(created_at) FROM brews GROUP BY coffee_id`
	rows, err := s.db.Query(query)
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

// ToggleBrewLearning flips the is_learning flag for a brew.
func (s *SQLiteBrewStorage) ToggleBrewLearning(id string) error {
	_, err := s.db.Exec("UPDATE brews SET is_learning = 1 - COALESCE(is_learning, 0) WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to toggle brew learning: %w", err)
	}
	return nil
}

// Delete removes a brew entry from the database.
func (s *SQLiteBrewStorage) Delete(id string) error {
	query := "DELETE FROM brews WHERE id = ?"

	result, err := s.db.Exec(query, id)
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

// scanBrew scans a single *sql.Row into a Brew struct.
func (s *SQLiteBrewStorage) scanBrew(row *sql.Row) (models.Brew, error) {
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

	if err == sql.ErrNoRows {
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

// scanBrews scans multiple rows into a Brew slice.
func (s *SQLiteBrewStorage) scanBrews(rows *sql.Rows) ([]models.Brew, error) {
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
