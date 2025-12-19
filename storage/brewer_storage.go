package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
)

// BrewerStorage defines the interface for brewer data persistence
type BrewerStorage interface {
	SaveBrewer(brewer models.Brewer) error
	GetBrewerByID(id string) (models.Brewer, error)
	GetAllBrewers() ([]models.Brewer, error)
	DeleteBrewer(id string) error
	UpdateBrewerRecipes(brewerID string, recipes []models.Recipe) error
	
	// Legacy coffee-based recipes (kept for backward compatibility)
	AddRecipeToBrewer(brewerID, coffeeID string) error
	GetBrewerRecipes(brewerID string) ([]models.Coffee, error)
	GetBrewerWithRecipes(brewerID string) (models.BrewerWithRecipes, error)
	GetAllBrewersWithRecipes() ([]models.BrewerWithRecipes, error)
	RemoveRecipeFromBrewer(brewerID, coffeeID string) error
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

// initTables creates the brewers and brewer_recipes tables if they don't exist
func (m *MySQLBrewerStorage) initTables() error {
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
		return fmt.Errorf("failed to create brewers table: %w", err)
	}
	
	recipeTableQuery := `
		CREATE TABLE IF NOT EXISTS brewer_recipes (
			id VARCHAR(36) PRIMARY KEY,
			brewer_id VARCHAR(36) NOT NULL,
			coffee_id VARCHAR(36) NOT NULL,
			created_at DATETIME,
			FOREIGN KEY (brewer_id) REFERENCES brewers(id) ON DELETE CASCADE,
			FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
			UNIQUE KEY unique_brewer_coffee (brewer_id, coffee_id)
		)
	`
	
	if _, err := m.db.Exec(recipeTableQuery); err != nil {
		return fmt.Errorf("failed to create brewer_recipes table: %w", err)
	}
	
	return nil
}

// SaveBrewer stores a brewer in the database
func (m *MySQLBrewerStorage) SaveBrewer(brewer models.Brewer) error {
	recipesJSON, err := json.Marshal(brewer.Recipes)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}
	
	query := `
		INSERT INTO brewers (id, name, pokeball_type, recipes, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	_, err = m.db.Exec(query, brewer.ID, brewer.Name, brewer.PokeballType, recipesJSON, brewer.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save brewer: %w", err)
	}
	
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
	query := `
		SELECT id, name, pokeball_type, recipes, created_at
		FROM brewers
		ORDER BY created_at ASC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brewers: %w", err)
	}
	defer rows.Close()
	
	var brewers []models.Brewer
	for rows.Next() {
		var brewer models.Brewer
		var recipesJSON []byte
		if err := rows.Scan(&brewer.ID, &brewer.Name, &brewer.PokeballType, &recipesJSON, &brewer.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan brewer: %w", err)
		}
		
		// Unmarshal recipes
		if len(recipesJSON) > 0 {
			if err := json.Unmarshal(recipesJSON, &brewer.Recipes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal recipes: %w", err)
			}
		}
		
		brewers = append(brewers, brewer)
	}
	
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

// AddRecipeToBrewer adds a coffee recipe to a brewer (max 4 per brewer)
func (m *MySQLBrewerStorage) AddRecipeToBrewer(brewerID, coffeeID string) error {
	// Check if brewer already has 4 recipes
	var count int
	countQuery := "SELECT COUNT(*) FROM brewer_recipes WHERE brewer_id = ?"
	if err := m.db.QueryRow(countQuery, brewerID).Scan(&count); err != nil {
		return fmt.Errorf("failed to count recipes: %w", err)
	}
	
	if count >= 4 {
		return fmt.Errorf("brewer already has maximum of 4 recipes")
	}
	
	// Generate ID for the brewer_recipe entry
	recipeID := fmt.Sprintf("%s-%s", brewerID, coffeeID)
	
	query := `
		INSERT INTO brewer_recipes (id, brewer_id, coffee_id, created_at)
		VALUES (?, ?, ?, NOW())
	`
	
	_, err := m.db.Exec(query, recipeID, brewerID, coffeeID)
	if err != nil {
		return fmt.Errorf("failed to add recipe to brewer: %w", err)
	}
	
	return nil
}

// GetBrewerRecipes retrieves all coffee recipes for a brewer (up to 4)
func (m *MySQLBrewerStorage) GetBrewerRecipes(brewerID string) ([]models.Coffee, error) {
	query := `
		SELECT c.id, c.name, c.origin, c.roaster, c.roast_level, c.processing_method,
		       c.tasting_notes, c.tasting_traits, c.rating, c.recipe, c.dripper,
		       c.end_time_minutes, c.end_time_seconds, c.created_at, c.updated_at
		FROM coffees c
		INNER JOIN brewer_recipes br ON c.id = br.coffee_id
		WHERE br.brewer_id = ?
		ORDER BY br.created_at DESC
		LIMIT 4
	`
	
	rows, err := m.db.Query(query, brewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query brewer recipes: %w", err)
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
		
		// Unmarshal JSON fields
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
	
	return coffees, nil
}

// GetBrewerWithRecipes retrieves a brewer with all its recipes
func (m *MySQLBrewerStorage) GetBrewerWithRecipes(brewerID string) (models.BrewerWithRecipes, error) {
	brewer, err := m.GetBrewerByID(brewerID)
	if err != nil {
		return models.BrewerWithRecipes{}, err
	}
	
	recipes, err := m.GetBrewerRecipes(brewerID)
	if err != nil {
		return models.BrewerWithRecipes{}, err
	}
	
	return models.BrewerWithRecipes{
		Brewer:  brewer,
		Recipes: recipes,
	}, nil
}

// GetAllBrewersWithRecipes retrieves all brewers with their recipes
func (m *MySQLBrewerStorage) GetAllBrewersWithRecipes() ([]models.BrewerWithRecipes, error) {
	brewers, err := m.GetAllBrewers()
	if err != nil {
		return nil, err
	}
	
	var result []models.BrewerWithRecipes
	for _, brewer := range brewers {
		recipes, err := m.GetBrewerRecipes(brewer.ID)
		if err != nil {
			return nil, err
		}
		
		result = append(result, models.BrewerWithRecipes{
			Brewer:  brewer,
			Recipes: recipes,
		})
	}
	
	return result, nil
}

// RemoveRecipeFromBrewer removes a recipe from a brewer
func (m *MySQLBrewerStorage) RemoveRecipeFromBrewer(brewerID, coffeeID string) error {
	query := "DELETE FROM brewer_recipes WHERE brewer_id = ? AND coffee_id = ?"
	
	result, err := m.db.Exec(query, brewerID, coffeeID)
	if err != nil {
		return fmt.Errorf("failed to remove recipe: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("recipe not found for this brewer")
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