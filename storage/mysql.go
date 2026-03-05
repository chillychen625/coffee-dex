package storage

import (
	"database/sql"
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

	if err := storage.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// DB returns the underlying database connection for use by other storage types
func (m *MySQLStorage) DB() *sql.DB {
	return m.db
}

// initTables creates the coffees and brews tables if they don't exist
func (m *MySQLStorage) initTables() error {
	// Create coffees table (bean info only)
	coffeesQuery := `
		CREATE TABLE IF NOT EXISTS coffees (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			origin VARCHAR(255),
			roaster VARCHAR(255),
			variety VARCHAR(255),
			roast_level VARCHAR(50),
			processing_method VARCHAR(100),
			roast_date DATE,
			is_finished BOOLEAN DEFAULT FALSE,
			created_at DATETIME,
			updated_at DATETIME
		)
	`

	if _, err := m.db.Exec(coffeesQuery); err != nil {
		return fmt.Errorf("failed to create coffees table: %w", err)
	}

	// Add columns if they don't exist (for existing databases)
	m.db.Exec("ALTER TABLE coffees ADD COLUMN roast_date DATE")
	m.db.Exec("ALTER TABLE coffees ADD COLUMN is_finished BOOLEAN DEFAULT FALSE")

	// Create brews table
	brewsQuery := `
		CREATE TABLE IF NOT EXISTS brews (
			id VARCHAR(36) PRIMARY KEY,
			coffee_id VARCHAR(36) NOT NULL,
			tasting_notes JSON,
			tasting_traits JSON,
			rating INT,
			recipe JSON,
			dripper VARCHAR(100),
			end_time_minutes INT,
			end_time_seconds INT,
			created_at DATETIME,
			FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE
		)
	`

	if _, err := m.db.Exec(brewsQuery); err != nil {
		return fmt.Errorf("failed to create brews table: %w", err)
	}

	// Create index for brews
	indexQuery := `
		CREATE INDEX IF NOT EXISTS idx_brews_coffee_id ON brews(coffee_id)
	`
	// Ignore error if index already exists (MySQL doesn't support IF NOT EXISTS for indexes in all versions)
	m.db.Exec(indexQuery)

	return nil
}

// Save stores a coffee entry in the database
func (m *MySQLStorage) Save(coffee models.Coffee) error {
	query := `
		INSERT INTO coffees (
			id, name, origin, roaster, variety, roast_level, processing_method,
			roast_date, is_finished, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert DateOnly to time.Time for SQL driver
	var roastDate interface{}
	if coffee.RoastDate != nil && !coffee.RoastDate.IsZero() {
		roastDate = coffee.RoastDate.Time()
	}

	_, err := m.db.Exec(
		query,
		coffee.ID, coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod, roastDate, coffee.IsFinished,
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
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at
		FROM coffees WHERE id = ?
	`

	row := m.db.QueryRow(query, id)

	var coffee models.Coffee
	var roastDate sql.NullTime
	var isFinished sql.NullBool

	err := row.Scan(
		&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
		&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
		&coffee.CreatedAt, &coffee.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return models.Coffee{}, fmt.Errorf("coffee not found")
	}
	if err != nil {
		return models.Coffee{}, fmt.Errorf("failed to get coffee: %w", err)
	}

	// Convert sql.NullTime to *DateOnly
	if roastDate.Valid {
		d := models.DateOnly(roastDate.Time)
		coffee.RoastDate = &d
	}

	// Convert sql.NullBool
	if isFinished.Valid {
		coffee.IsFinished = isFinished.Bool
	}

	return coffee, nil
}

// GetAll retrieves all coffees from the database
func (m *MySQLStorage) GetAll() ([]models.Coffee, error) {
	query := `
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at
		FROM coffees
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query coffees: %w", err)
	}
	defer rows.Close()

	var coffees []models.Coffee

	for rows.Next() {
		var coffee models.Coffee
		var roastDate sql.NullTime
		var isFinished sql.NullBool

		err := rows.Scan(
			&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
			&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
			&coffee.CreatedAt, &coffee.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee: %w", err)
		}

		// Convert sql.NullTime to *DateOnly
		if roastDate.Valid {
			d := models.DateOnly(roastDate.Time)
			coffee.RoastDate = &d
		}

		// Convert sql.NullBool
		if isFinished.Valid {
			coffee.IsFinished = isFinished.Bool
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
		SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		       roast_date, is_finished, created_at, updated_at
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
		var roastDate sql.NullTime
		var isFinished sql.NullBool

		err := rows.Scan(
			&coffee.ID, &coffee.Name, &coffee.Origin, &coffee.Roaster, &coffee.Variety,
			&coffee.RoastLevel, &coffee.ProcessingMethod, &roastDate, &isFinished,
			&coffee.CreatedAt, &coffee.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan coffee: %w", err)
		}

		// Convert sql.NullTime to *DateOnly
		if roastDate.Valid {
			d := models.DateOnly(roastDate.Time)
			coffee.RoastDate = &d
		}

		// Convert sql.NullBool
		if isFinished.Valid {
			coffee.IsFinished = isFinished.Bool
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
	query := `
		UPDATE coffees SET
			name=?, origin=?, roaster=?, variety=?, roast_level=?, processing_method=?,
			roast_date=?, is_finished=?, updated_at=?
		WHERE id=?
	`

	// Convert DateOnly to time.Time for SQL driver
	var roastDate interface{}
	if coffee.RoastDate != nil && !coffee.RoastDate.IsZero() {
		roastDate = coffee.RoastDate.Time()
	}

	result, err := m.db.Exec(
		query,
		coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod, roastDate, coffee.IsFinished,
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
