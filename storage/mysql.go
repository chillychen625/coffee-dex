package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// MySQLStorage implements CoffeeStorage using MySQL database
type MySQLStorage struct {
	db *sql.DB
}

// NewMySQLStorage creates a new MySQL storage and initializes the database
func NewMySQLStorage(host, user, password, dbname string) (*MySQLStorage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbname)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	storage := &MySQLStorage{db: db}
	
	if err := storage.initTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}
	
	return storage, nil
}

// initTable creates the coffees table if it doesn't exist
func (m *MySQLStorage) initTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS coffees (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			origin VARCHAR(255),
			roaster VARCHAR(255),
			roast_level VARCHAR(50),
			processing_method VARCHAR(100),
			tasting_notes JSON,
			tasting_traits JSON,
			rating INT,
			recipe JSON,
			dripper VARCHAR(100),
			end_time_minutes INT,
			end_time_seconds INT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`
	
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	
	return nil
}

// Save stores a coffee entry in the database
func (m *MySQLStorage) Save(coffee models.Coffee) error {
	tastingNotesJSON, err := json.Marshal(coffee.TastingNotes)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting notes: %w", err)
	}
	
	tastingTraitsJSON, err := json.Marshal(coffee.TastingTraits)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting traits: %w", err)
	}
	
	recipeJSON, err := json.Marshal(coffee.Recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe: %w", err)
	}
	
	query := `
		INSERT INTO coffees (
			id, name, origin, roaster, roast_level, processing_method,
			tasting_notes, tasting_traits, rating, recipe, dripper,
			end_time_minutes, end_time_seconds, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = m.db.Exec(
		query,
		coffee.ID, coffee.Name, coffee.Origin, coffee.Roaster,
		coffee.RoastLevel, coffee.ProcessingMethod,
		tastingNotesJSON, tastingTraitsJSON, coffee.Rating, recipeJSON, coffee.Dripper,
		coffee.EndTime.Minutes, coffee.EndTime.Seconds,
		coffee.CreatedAt, coffee.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save coffee: %w", err)
	}
	
	return nil
}

// GetByID retrieves a coffee by ID from the database
func (m *MySQLStorage) GetByID(id string) (models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, roast_level, processing_method,
		       tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, updated_at
		FROM coffees WHERE id = ?
	`
	
	row := m.db.QueryRow(query, id)
	
	var coffee models.Coffee
	var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
	
	err := row.Scan(
		&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster,
		&coffee.RoastLevel, &coffee.ProcessingMethod,
		&tastingNotesJSON, &tastingTraitsJSON, &coffee.Rating, &recipeJSON, &coffee.Dripper,
		&coffee.EndTime.Minutes, &coffee.EndTime.Seconds,
		&coffee.CreatedAt, &coffee.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return models.Coffee{}, fmt.Errorf("coffee not found")
	}
	if err != nil {
		return models.Coffee{}, fmt.Errorf("failed to get coffee: %w", err)
	}
	
	if err := json.Unmarshal(tastingNotesJSON, &coffee.TastingNotes); err != nil {
		return models.Coffee{}, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
	}
	
	if err := json.Unmarshal(tastingTraitsJSON, &coffee.TastingTraits); err != nil {
		return models.Coffee{}, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
	}
	
	if err := json.Unmarshal(recipeJSON, &coffee.Recipe); err != nil {
		return models.Coffee{}, fmt.Errorf("failed to unmarshal recipe: %w", err)
	}
	
	return coffee, nil
}

// GetAll retrieves all coffees from the database
func (m *MySQLStorage) GetAll() ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, roast_level, processing_method,
		       tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, updated_at
		FROM coffees
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffees: %w", err)
	}
	defer rows.Close()
	
	var coffees []models.Coffee
	
	for rows.Next() {
		var coffee models.Coffee
		var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
		
		err := rows.Scan(
			&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster,
			&coffee.RoastLevel, &coffee.ProcessingMethod,
			&tastingNotesJSON, &tastingTraitsJSON, &coffee.Rating, &recipeJSON, &coffee.Dripper,
			&coffee.EndTime.Minutes, &coffee.EndTime.Seconds,
			&coffee.CreatedAt, &coffee.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee: %w", err)
		}
		
		if err := json.Unmarshal(tastingNotesJSON, &coffee.TastingNotes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
		}
		
		if err := json.Unmarshal(tastingTraitsJSON, &coffee.TastingTraits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
		}
		
		if err := json.Unmarshal(recipeJSON, &coffee.Recipe); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
		}
		
		coffees = append(coffees, coffee)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return coffees, nil
}

// GetRecent retrieves the most recent coffees from the database
func (m *MySQLStorage) GetRecent(limit int) ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, roast_level, processing_method,
		       tasting_notes, tasting_traits, rating, recipe, dripper,
		       end_time_minutes, end_time_seconds, created_at, updated_at
		FROM coffees
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent coffees: %w", err)
	}
	defer rows.Close()
	
	var coffees []models.Coffee
	
	for rows.Next() {
		var coffee models.Coffee
		var tastingNotesJSON, tastingTraitsJSON, recipeJSON []byte
		
		err := rows.Scan(
			&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster,
			&coffee.RoastLevel, &coffee.ProcessingMethod,
			&tastingNotesJSON, &tastingTraitsJSON, &coffee.Rating, &recipeJSON, &coffee.Dripper,
			&coffee.EndTime.Minutes, &coffee.EndTime.Seconds,
			&coffee.CreatedAt, &coffee.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee: %w", err)
		}
		
		if err := json.Unmarshal(tastingNotesJSON, &coffee.TastingNotes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting notes: %w", err)
		}
		
		if err := json.Unmarshal(tastingTraitsJSON, &coffee.TastingTraits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tasting traits: %w", err)
		}
		
		if err := json.Unmarshal(recipeJSON, &coffee.Recipe); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
		}
		
		coffees = append(coffees, coffee)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return coffees, nil
}

// Update modifies an existing coffee entry
func (m *MySQLStorage) Update(id string, coffee models.Coffee) error {
	tastingNotesJSON, err := json.Marshal(coffee.TastingNotes)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting notes: %w", err)
	}
	
	tastingTraitsJSON, err := json.Marshal(coffee.TastingTraits)
	if err != nil {
		return fmt.Errorf("failed to marshal tasting traits: %w", err)
	}
	
	recipeJSON, err := json.Marshal(coffee.Recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe: %w", err)
	}
	
	query := `
		UPDATE coffees SET
			name=?, origin=?, roaster=?, roast_level=?, processing_method=?,
			tasting_notes=?, tasting_traits=?, rating=?, recipe=?, dripper=?,
			end_time_minutes=?, end_time_seconds=?, updated_at=?
		WHERE id=?
	`
	
	result, err := m.db.Exec(
		query,
		coffee.Name, coffee.Origin, coffee.Roaster,
		coffee.RoastLevel, coffee.ProcessingMethod,
		tastingNotesJSON, tastingTraitsJSON, coffee.Rating, recipeJSON, coffee.Dripper,
		coffee.EndTime.Minutes, coffee.EndTime.Seconds,
		coffee.UpdatedAt, id,
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

// Delete removes a coffee entry from the database
func (m *MySQLStorage) Delete(id string) error {
	query := "DELETE FROM coffees WHERE id = ?"
	
	result, err := m.db.Exec(query, id)
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

// Close closes the database connection
func (m *MySQLStorage) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}